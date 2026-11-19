// Copyright The Prometheus Authors
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

package collector

import (
	"context"
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector(statDatabaseConflictsSubsystem, serverScope, NewPGStatDatabaseConflictsCollector)
}

type PGStatDatabaseConflictsCollector struct{}

func NewPGStatDatabaseConflictsCollector(collectorConfig) (Collector, error) {
	return &PGStatDatabaseConflictsCollector{}, nil
}

func statDatabaseConflictsDesc(name, help string) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName(namespace, statDatabaseConflictsSubsystem, name),
		help,
		[]string{"datid", "datname"},
		prometheus.Labels{},
	)
}

var (
	statDatabaseConflictsTablespaceDesc = statDatabaseConflictsDesc("confl_tablespace", "Number of queries in this database that have been canceled due to dropped tablespaces")
	statDatabaseConflictsLockDesc       = statDatabaseConflictsDesc("confl_lock", "Number of queries in this database that have been canceled due to lock timeouts")
	statDatabaseConflictsSnapshotDesc   = statDatabaseConflictsDesc("confl_snapshot", "Number of queries in this database that have been canceled due to old snapshots")
	statDatabaseConflictsBufferpinDesc  = statDatabaseConflictsDesc("confl_bufferpin", "Number of queries in this database that have been canceled due to pinned buffers")
	statDatabaseConflictsDeadlockDesc   = statDatabaseConflictsDesc("confl_deadlock", "Number of queries in this database that have been canceled due to deadlocks")

	statDatabaseConflictsQuery = `SELECT
		datid,
		datname,
		confl_tablespace,
		confl_lock,
		confl_snapshot,
		confl_bufferpin,
		confl_deadlock
	FROM pg_stat_database_conflicts`
)

func (PGStatDatabaseConflictsCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	rows, err := instance.getDB().QueryContext(ctx, statDatabaseConflictsQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	type row struct {
		datid, datname string
		values         [5]sql.NullInt64
	}
	var all []row
	for rows.Next() {
		var datid, datname sql.NullString
		var r row
		if err := rows.Scan(&datid, &datname, &r.values[0], &r.values[1], &r.values[2], &r.values[3], &r.values[4]); err != nil {
			return err
		}
		if !datid.Valid || !datname.Valid {
			continue
		}
		r.datid, r.datname = datid.String, datname.String
		all = append(all, r)
	}
	// Emit nothing if iteration failed part-way, so no partial metric set escapes.
	if err := rows.Err(); err != nil {
		return err
	}

	descs := [5]*prometheus.Desc{
		statDatabaseConflictsTablespaceDesc,
		statDatabaseConflictsLockDesc,
		statDatabaseConflictsSnapshotDesc,
		statDatabaseConflictsBufferpinDesc,
		statDatabaseConflictsDeadlockDesc,
	}
	for _, r := range all {
		for i, d := range descs {
			if !r.values[i].Valid {
				continue
			}
			ch <- prometheus.MustNewConstMetric(d, prometheus.CounterValue,
				int64CounterValue(r.values[i], instance.wrapLargeCounters), r.datid, r.datname)
		}
	}
	return nil
}
