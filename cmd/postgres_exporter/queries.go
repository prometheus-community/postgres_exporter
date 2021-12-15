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

package postgres_exporter

import "fmt"

// ColumnUsage should be one of several enum values which describe how a
// queried row is to be converted to a Prometheus metric.
type ColumnUsage int

// ColumnMapping is the user-friendly representation of a prometheus descriptor map
type ColumnMapping struct {
	Usage       ColumnUsage
	Description string
	Mapping     map[string]float64
}

// IntermediateMetricMap holds the partially loaded metric map parsing.
type IntermediateMetricMap struct {
	ColumnMappings map[string]ColumnMapping
}

// MappingOptions is a copy of ColumnMapping used only for parsing
type MappingOptions struct {
	Usage       string
	Description string
	Mapping     map[string]float64
}

// Mapping represents a set of MappingOptions
type Mapping map[string]MappingOptions

func MetricMaps() map[string]IntermediateMetricMap {
	return map[string]IntermediateMetricMap{
		"pg_stat_bgwriter": {
			map[string]ColumnMapping{
				"checkpoints_timed":     {COUNTER, "Number of scheduled checkpoints that have been performed", nil},
				"checkpoints_req":       {COUNTER, "Number of requested checkpoints that have been performed", nil},
				"checkpoint_write_time": {COUNTER, "Total amount of time that has been spent in the portion of checkpoint processing where files are written to disk, in milliseconds", nil},
				"checkpoint_sync_time":  {COUNTER, "Total amount of time that has been spent in the portion of checkpoint processing where files are synchronized to disk, in milliseconds", nil},
				"buffers_checkpoint":    {COUNTER, "Number of buffers written during checkpoints", nil},
				"buffers_clean":         {COUNTER, "Number of buffers written by the background writer", nil},
				"maxwritten_clean":      {COUNTER, "Number of times the background writer stopped a cleaning scan because it had written too many buffers", nil},
				"buffers_backend":       {COUNTER, "Number of buffers written directly by a backend", nil},
				"buffers_backend_fsync": {COUNTER, "Number of times a backend had to execute its own fsync call (normally the background writer handles those even when the backend does its own write)", nil},
				"buffers_alloc":         {COUNTER, "Number of buffers allocated", nil},
				"stats_reset":           {COUNTER, "Time at which these statistics were last reset", nil},
			},
		},
		"pg_stat_database": {
			map[string]ColumnMapping{
				"datid":          {LABEL, "OID of a database", nil},
				"datname":        {LABEL, "Name of this database", nil},
				"numbackends":    {GAUGE, "Number of backends currently connected to this database. This is the only column in this view that returns a value reflecting current state; all other columns return the accumulated values since the last reset.", nil},
				"xact_commit":    {COUNTER, "Number of transactions in this database that have been committed", nil},
				"xact_rollback":  {COUNTER, "Number of transactions in this database that have been rolled back", nil},
				"blks_read":      {COUNTER, "Number of disk blocks read in this database", nil},
				"blks_hit":       {COUNTER, "Number of times disk blocks were found already in the buffer cache, so that a read was not necessary (this only includes hits in the PostgreSQL buffer cache, not the operating system's file system cache)", nil},
				"tup_returned":   {COUNTER, "Number of rows returned by queries in this database", nil},
				"tup_fetched":    {COUNTER, "Number of rows fetched by queries in this database", nil},
				"tup_inserted":   {COUNTER, "Number of rows inserted by queries in this database", nil},
				"tup_updated":    {COUNTER, "Number of rows updated by queries in this database", nil},
				"tup_deleted":    {COUNTER, "Number of rows deleted by queries in this database", nil},
				"conflicts":      {COUNTER, "Number of queries canceled due to conflicts with recovery in this database. (Conflicts occur only on standby servers; see pg_stat_database_conflicts for details.)", nil},
				"temp_files":     {COUNTER, "Number of temporary files created by queries in this database. All temporary files are counted, regardless of why the temporary file was created (e.g., sorting or hashing), and regardless of the log_temp_files setting.", nil},
				"temp_bytes":     {COUNTER, "Total amount of data written to temporary files by queries in this database. All temporary files are counted, regardless of why the temporary file was created, and regardless of the log_temp_files setting.", nil},
				"deadlocks":      {COUNTER, "Number of deadlocks detected in this database", nil},
				"blk_read_time":  {COUNTER, "Time spent reading data file blocks by backends in this database, in milliseconds", nil},
				"blk_write_time": {COUNTER, "Time spent writing data file blocks by backends in this database, in milliseconds", nil},
				"stats_reset":    {COUNTER, "Time at which these statistics were last reset", nil},
			},
		},
		"pg_stat_database_conflicts": {
			map[string]ColumnMapping{
				"datid":            {LABEL, "OID of a database", nil},
				"datname":          {LABEL, "Name of this database", nil},
				"confl_tablespace": {COUNTER, "Number of queries in this database that have been canceled due to dropped tablespaces", nil},
				"confl_lock":       {COUNTER, "Number of queries in this database that have been canceled due to lock timeouts", nil},
				"confl_snapshot":   {COUNTER, "Number of queries in this database that have been canceled due to old snapshots", nil},
				"confl_bufferpin":  {COUNTER, "Number of queries in this database that have been canceled due to pinned buffers", nil},
				"confl_deadlock":   {COUNTER, "Number of queries in this database that have been canceled due to deadlocks", nil},
			},
		},
		"pg_locks": {
			map[string]ColumnMapping{
				"datname": {LABEL, "Name of this database", nil},
				"mode":    {LABEL, "Type of Lock", nil},
				"count":   {GAUGE, "Number of locks", nil},
			},
		},
		"pg_stat_replication": {
			map[string]ColumnMapping{
				"pid":              {DISCARD, "Process ID of a WAL sender process", nil},
				"usesysid":         {DISCARD, "OID of the user logged into this WAL sender process", nil},
				"usename":          {DISCARD, "Name of the user logged into this WAL sender process", nil},
				"application_name": {LABEL, "Name of the application that is connected to this WAL sender", nil},
				"client_addr":      {LABEL, "IP address of the client connected to this WAL sender. If this field is null, it indicates that the client is connected via a Unix socket on the server machine.", nil},
				"client_hostname":  {DISCARD, "Host name of the connected client, as reported by a reverse DNS lookup of client_addr. This field will only be non-null for IP connections, and only when log_hostname is enabled.", nil},
				"client_port":      {DISCARD, "TCP port number that the client is using for communication with this WAL sender, or -1 if a Unix socket is used", nil},
				"backend_start": {DISCARD, "with time zone	Time when this process was started, i.e., when the client connected to this WAL sender", nil},
				"backend_xmin":             {DISCARD, "The current backend's xmin horizon.", nil},
				"state":                    {LABEL, "Current WAL sender state", nil},
				"sent_lsn":                 {DISCARD, "Last transaction log position sent on this connection", nil},
				"write_lsn":                {DISCARD, "Last transaction log position written to disk by this standby server", nil},
				"flush_lsn":                {DISCARD, "Last transaction log position flushed to disk by this standby server", nil},
				"replay_lsn":               {DISCARD, "Last transaction log position replayed into the database on this standby server", nil},
				"sync_priority":            {DISCARD, "Priority of this standby server for being chosen as the synchronous standby", nil},
				"sync_state":               {DISCARD, "Synchronous state of this standby server", nil},
				"slot_name":                {LABEL, "A unique, cluster-wide identifier for the replication slot", nil},
				"plugin":                   {DISCARD, "The base name of the shared object containing the output plugin this logical slot is using, or null for physical slots", nil},
				"slot_type":                {DISCARD, "The slot type - physical or logical", nil},
				"datoid":                   {DISCARD, "The OID of the database this slot is associated with, or null. Only logical slots have an associated database", nil},
				"database":                 {DISCARD, "The name of the database this slot is associated with, or null. Only logical slots have an associated database", nil},
				"active":                   {DISCARD, "True if this slot is currently actively being used", nil},
				"active_pid":               {DISCARD, "Process ID of a WAL sender process", nil},
				"xmin":                     {DISCARD, "The oldest transaction that this slot needs the database to retain. VACUUM cannot remove tuples deleted by any later transaction", nil},
				"catalog_xmin":             {DISCARD, "The oldest transaction affecting the system catalogs that this slot needs the database to retain. VACUUM cannot remove catalog tuples deleted by any later transaction", nil},
				"restart_lsn":              {DISCARD, "The address (LSN) of oldest WAL which still might be required by the consumer of this slot and thus won't be automatically removed during checkpoints", nil},
				"pg_current_xlog_location": {DISCARD, "pg_current_xlog_location", nil},
				"pg_current_wal_lsn":       {DISCARD, "pg_current_xlog_location", nil},
				"pg_current_wal_lsn_bytes": {GAUGE, "WAL position in bytes", nil},
				"pg_wal_lsn_diff":          {GAUGE, "Lag in bytes between master and slave", nil},
				"confirmed_flush_lsn":      {DISCARD, "LSN position a consumer of a slot has confirmed flushing the data received", nil},
				"write_lag":                {DISCARD, "Time elapsed between flushing recent WAL locally and receiving notification that this standby server has written it (but not yet flushed it or applied it). This can be used to gauge the delay that synchronous_commit level remote_write incurred while committing if this server was configured as a synchronous standby.", nil},
				"flush_lag":                {DISCARD, "Time elapsed between flushing recent WAL locally and receiving notification that this standby server has written and flushed it (but not yet applied it). This can be used to gauge the delay that synchronous_commit level remote_flush incurred while committing if this server was configured as a synchronous standby.", nil},
				"replay_lag":               {DISCARD, "Time elapsed between flushing recent WAL locally and receiving notification that this standby server has written, flushed and applied it. This can be used to gauge the delay that synchronous_commit level remote_apply incurred while committing if this server was configured as a synchronous standby.", nil},
			},
		},
		"pg_replication_slots": {
			map[string]ColumnMapping{
				"slot_name":       {LABEL, "Name of the replication slot", nil},
				"database":        {LABEL, "Name of the database", nil},
				"active":          {GAUGE, "Flag indicating if the slot is active", nil},
				"pg_wal_lsn_diff": {GAUGE, "Replication lag in bytes", nil},
			},
		},
		"pg_stat_archiver": {
			map[string]ColumnMapping{
				"archived_count":     {COUNTER, "Number of WAL files that have been successfully archived", nil},
				"last_archived_wal":  {DISCARD, "Name of the last WAL file successfully archived", nil},
				"last_archived_time": {DISCARD, "Time of the last successful archive operation", nil},
				"failed_count":       {COUNTER, "Number of failed attempts for archiving WAL files", nil},
				"last_failed_wal":    {DISCARD, "Name of the WAL file of the last failed archival operation", nil},
				"last_failed_time":   {DISCARD, "Time of the last failed archival operation", nil},
				"stats_reset":        {DISCARD, "Time at which these statistics were last reset", nil},
				"last_archive_age":   {GAUGE, "Time in seconds since last WAL segment was successfully archived", nil},
			},
		},
		"pg_stat_activity": {
			map[string]ColumnMapping{
				"datname":         {LABEL, "Name of this database", nil},
				"state":           {LABEL, "connection state", nil},
				"count":           {GAUGE, "number of connections in this state", nil},
				"max_tx_duration": {GAUGE, "max duration in seconds any active transaction has been running", nil},
			},
		},
	}
}

func Queries() map[string]string {
	return map[string]string{"pg_locks": `SELECT pg_database.datname,tmp.mode,COALESCE(count,0) as count
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
		"pg_stat_replication": `
			SELECT *,
				(case pg_is_in_recovery() when 't' then null else pg_current_wal_lsn() end) AS pg_current_wal_lsn,
				(case pg_is_in_recovery() when 't' then null else pg_wal_lsn_diff(pg_current_wal_lsn(), pg_lsn('0/0'))::float end) AS pg_current_wal_lsn_bytes,
				(case pg_is_in_recovery() when 't' then null else pg_wal_lsn_diff(pg_current_wal_lsn(), replay_lsn)::float end) AS pg_wal_lsn_diff
			FROM pg_stat_replication`,

		"pg_replication_slots": `
			SELECT slot_name, database, active, pg_wal_lsn_diff(pg_current_wal_lsn(), restart_lsn) 
			FROM pg_replication_slots`,

		"pg_stat_archiver": `
			SELECT *,
				extract(epoch from now() - last_archived_time) AS last_archive_age
			FROM pg_stat_archiver`,

		"pg_stat_activity": `
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
				ON tmp.state = tmp2.state AND pg_database.datname = tmp2.datname`,

		"pg_stat_database_conflicts": `
			SELECT  datid,
					datname,
					confl_tablespace,
					confl_lock,
					confl_snapshot,
					confl_bufferpin,
					confl_deadlock
			FROM pg_stat_database_conflicts`,

		"pg_stat_database": `
			SELECT 
				datid,
				datname, 
				numbackends,
				xact_commit, 
				xact_rollback,
				blks_read,
				blks_hit,
				tup_returned,
				tup_fetched,
				tup_inserted,
				tup_updated,
				tup_deleted,
				conflicts,
				temp_files,
				temp_bytes,
				deadlocks,
				blk_read_time,
				blk_write_time,
				stats_reset
			FROM pg_stat_database `,

		"pg_stat_bgwriter": `
			SELECT 
				checkpoints_timed,
				checkpoints_req,
				checkpoint_write_time,
				checkpoint_sync_time,
				buffers_checkpoint,
				buffers_clean,
				maxwritten_clean,
				buffers_backend,
				buffers_backend_fsync,
				buffers_alloc,
				stats_reset
			FROM pg_stat_bgwriter`,
	}
}

func DumpMaps() {
	for name, cmap := range MetricMaps() {
		query, ok := Queries()[name]
		if !ok {
			fmt.Println(name)
		} else {
			fmt.Println(name, query)
		}

		for column, details := range cmap.ColumnMappings {
			fmt.Printf("  %-40s %v\n", column, details)
		}
		fmt.Println()
	}
}
