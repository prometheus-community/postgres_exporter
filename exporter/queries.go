// Copyright 2021 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package exporter

import (
	"errors"
	"fmt"

	"github.com/blang/semver"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"gopkg.in/yaml.v2"
)

// UserQuery represents a user defined query
type UserQuery struct {
	Query        string    `yaml:"query"`
	Metrics      []Mapping `yaml:"metrics"`
	Master       bool      `yaml:"master"`        // Querying only for master database
	CacheSeconds uint64    `yaml:"cache_seconds"` // Number of seconds to cache the namespace result metrics for.
	RunOnServer  string    `yaml:"runonserver"`   // Querying to run on which server version
}

// UserQueries represents a set of UserQuery objects
type UserQueries map[string]UserQuery

// OverrideQuery 's are run in-place of simple namespace look ups, and provide
// advanced functionality. But they have a tendency to postgres version specific.
// There aren't too many versions, so we simply store customized versions using
// the semver matching we do for columns.
type OverrideQuery struct {
	versionRange semver.Range
	query        string
}

// Overriding queries for namespaces above.
// TODO: validate this is a closed set in tests, and there are no overlaps
var queryOverrides = map[string][]OverrideQuery{
	"pg_locks": {
		{
			semver.MustParseRange(">0.0.0"),
			`SELECT pg_database.datname,tmp.mode,COALESCE(count,0) as count
			FROM
				(
				  VALUES ('accesssharelock'),
				         ('rowsharelock'),
				         ('rowexclusivelock'),
				         ('shareupdateexclusivelock'),
				         ('sharelock'),
				         ('sharerowexclusivelock'),
				         ('exclusivelock'),
				         ('accessexclusivelock'),
					 ('sireadlock')
				) AS tmp(mode) CROSS JOIN pg_database
			LEFT JOIN
			  (SELECT database, lower(mode) AS mode,count(*) AS count
			  FROM pg_locks WHERE database IS NOT NULL
			  GROUP BY database, lower(mode)
			) AS tmp2
			ON tmp.mode=tmp2.mode and pg_database.oid = tmp2.database ORDER BY 1`,
		},
	},

	"pg_stat_replication": {
		{
			semver.MustParseRange(">=10.0.0"),
			`
			SELECT *,
				(case pg_is_in_recovery() when 't' then null else pg_current_wal_lsn() end) AS pg_current_wal_lsn,
				(case pg_is_in_recovery() when 't' then null else pg_wal_lsn_diff(pg_current_wal_lsn(), pg_lsn('0/0'))::float end) AS pg_current_wal_lsn_bytes,
				(case pg_is_in_recovery() when 't' then null else pg_wal_lsn_diff(pg_current_wal_lsn(), replay_lsn)::float end) AS pg_wal_lsn_diff
			FROM pg_stat_replication
			`,
		},
		{
			semver.MustParseRange(">=9.2.0 <10.0.0"),
			`
			SELECT *,
				(case pg_is_in_recovery() when 't' then null else pg_current_xlog_location() end) AS pg_current_xlog_location,
				(case pg_is_in_recovery() when 't' then null else pg_xlog_location_diff(pg_current_xlog_location(), replay_location)::float end) AS pg_xlog_location_diff
			FROM pg_stat_replication
			`,
		},
		{
			semver.MustParseRange("<9.2.0"),
			`
			SELECT *,
				(case pg_is_in_recovery() when 't' then null else pg_current_xlog_location() end) AS pg_current_xlog_location
			FROM pg_stat_replication
			`,
		},
	},

	"pg_replication_slots": {
		{
			semver.MustParseRange(">=9.4.0 <10.0.0"),
			`
			SELECT slot_name, database, active, pg_xlog_location_diff(pg_current_xlog_location(), restart_lsn)
			FROM pg_replication_slots
			`,
		},
		{
			semver.MustParseRange(">=10.0.0"),
			`
			SELECT slot_name, database, active, pg_wal_lsn_diff(pg_current_wal_lsn(), restart_lsn)
			FROM pg_replication_slots
			`,
		},
	},

	"pg_stat_archiver": {
		{
			semver.MustParseRange(">=0.0.0"),
			`
			SELECT *,
				extract(epoch from now() - last_archived_time) AS last_archive_age
			FROM pg_stat_archiver
			`,
		},
	},

	"pg_stat_activity": {
		// This query only works
		{
			semver.MustParseRange(">=9.2.0"),
			`
			SELECT
				pg_database.datname,
				tmp.state,
				COALESCE(count,0) as count,
				COALESCE(max_tx_duration,0) as max_tx_duration
			FROM
				(
				  VALUES ('active'),
				  		 ('idle'),
				  		 ('idle in transaction'),
				  		 ('idle in transaction (aborted)'),
				  		 ('fastpath function call'),
				  		 ('disabled')
				) AS tmp(state) CROSS JOIN pg_database
			LEFT JOIN
			(
				SELECT
					datname,
					state,
					count(*) AS count,
					MAX(EXTRACT(EPOCH FROM now() - xact_start))::float AS max_tx_duration
				FROM pg_stat_activity GROUP BY datname,state) AS tmp2
				ON tmp.state = tmp2.state AND pg_database.datname = tmp2.datname
			`,
		},
		{
			semver.MustParseRange("<9.2.0"),
			`
			SELECT
				datname,
				'unknown' AS state,
				COALESCE(count(*),0) AS count,
				COALESCE(MAX(EXTRACT(EPOCH FROM now() - xact_start))::float,0) AS max_tx_duration
			FROM pg_stat_activity GROUP BY datname
			`,
		},
	},
}

// Convert the query override file to the version-specific query override file
// for the exporter.
func makeQueryOverrideMap(pgVersion semver.Version, queryOverrides map[string][]OverrideQuery, logger log.Logger) map[string]string {
	resultMap := make(map[string]string)
	for name, overrideDef := range queryOverrides {
		// Find a matching semver. We make it an error to have overlapping
		// ranges at test-time, so only 1 should ever match.
		matched := false
		for _, queryDef := range overrideDef {
			if queryDef.versionRange(pgVersion) {
				resultMap[name] = queryDef.query
				matched = true
				break
			}
		}
		if !matched {
			level.Warn(logger).Log("msg", "No query matched override, disabling metric space", "name", name)
			resultMap[name] = ""
		}
	}

	return resultMap
}

func parseUserQueries(content []byte, logger log.Logger) (map[string]intermediateMetricMap, map[string]string, error) {
	var userQueries UserQueries

	err := yaml.Unmarshal(content, &userQueries)
	if err != nil {
		return nil, nil, err
	}

	// Stores the loaded map representation
	metricMaps := make(map[string]intermediateMetricMap)
	newQueryOverrides := make(map[string]string)

	for metric, specs := range userQueries {
		level.Debug(logger).Log("msg", "New user metric namespace from YAML metric", "metric", metric, "cache_seconds", specs.CacheSeconds)
		newQueryOverrides[metric] = specs.Query
		metricMap, ok := metricMaps[metric]
		if !ok {
			// Namespace for metric not found - add it.
			newMetricMap := make(map[string]ColumnMapping)
			metricMap = intermediateMetricMap{
				columnMappings: newMetricMap,
				master:         specs.Master,
				cacheSeconds:   specs.CacheSeconds,
			}
			metricMaps[metric] = metricMap
		}
		for _, metric := range specs.Metrics {
			for name, mappingOption := range metric {
				var columnMapping ColumnMapping
				tmpUsage, _ := stringToColumnUsage(mappingOption.Usage)
				columnMapping.usage = tmpUsage
				columnMapping.description = mappingOption.Description

				// TODO: we should support cu
				columnMapping.mapping = nil
				// Should we support this for users?
				columnMapping.supportedVersions = nil

				metricMap.columnMappings[name] = columnMapping
			}
		}
	}
	return metricMaps, newQueryOverrides, nil
}

// Add queries to the builtinMetricMaps and queryOverrides maps. Added queries do not
// respect version requirements, because it is assumed that the user knows
// what they are doing with their version of postgres.
//
// This function modifies metricMap and queryOverrideMap to contain the new
// queries.
// TODO: test code for all cu.
// TODO: the YAML this supports is "non-standard" - we should move away from it.
func addQueries(content []byte, pgVersion semver.Version, server *Server) error {
	metricMaps, newQueryOverrides, err := parseUserQueries(content, server.Logger)
	if err != nil {
		return err
	}
	// Convert the loaded metric map into exporter representation
	partialExporterMap := makeDescMap(pgVersion, server.labels, metricMaps, "pg", server.Logger)

	// Merge the two maps (which are now quite flatteend)
	for k, v := range partialExporterMap {
		_, found := server.metricMap[k]
		if found {
			level.Debug(server.Logger).Log("msg", "Overriding metric from user YAML file", "metric", k)
		} else {
			level.Debug(server.Logger).Log("msg", "Adding new metric from user YAML file", "metric", k)
		}
		server.metricMap[k] = v
	}

	// Merge the query override map
	for k, v := range newQueryOverrides {
		_, found := server.queryOverrides[k]
		if found {
			level.Debug(server.Logger).Log("msg", "Overriding query override from user YAML file", "query_override", k)
		} else {
			level.Debug(server.Logger).Log("msg", "Adding new query override from user YAML file", "query_override", k)
		}
		server.queryOverrides[k] = v
	}
	return nil
}

func queryDatabases(server *Server) ([]string, error) {
	rows, err := server.db.Query("SELECT datname FROM pg_database WHERE datallowconn = true AND datistemplate = false AND datname != current_database()")
	if err != nil {
		return nil, fmt.Errorf("Error retrieving databases: %v", err)
	}
	defer rows.Close() // nolint: errcheck

	var databaseName string
	result := make([]string, 0)
	for rows.Next() {
		err = rows.Scan(&databaseName)
		if err != nil {
			return nil, errors.New(fmt.Sprintln("Error retrieving rows:", err))
		}
		result = append(result, databaseName)
	}

	return result, nil
}
