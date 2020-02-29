// package builtin contains the default metrics packaged with the exporter.
package builtin

import (
	. "github.com/wrouesnel/postgres_exporter/pkg/queries/metricmaps"
)

var builtin = QueryMap{
	Global:   &QueryConfig{
		MetricMap:      MetricMaps{
			"pg_stat_bgwriter": {
				map[string]ColumnMapping{
					"checkpoints_timed":     {COUNTER, "Number of scheduled checkpoints that have been performed", nil, nil},
					"checkpoints_req":       {COUNTER, "Number of requested checkpoints that have been performed", nil, nil},
					"checkpoint_write_time": {COUNTER, "Total amount of time that has been spent in the portion of checkpoint processing where files are written to disk, in milliseconds", nil, nil},
					"checkpoint_sync_time":  {COUNTER, "Total amount of time that has been spent in the portion of checkpoint processing where files are synchronized to disk, in milliseconds", nil, nil},
					"buffers_checkpoint":    {COUNTER, "Number of buffers written during checkpoints", nil, nil},
					"buffers_clean":         {COUNTER, "Number of buffers written by the background writer", nil, nil},
					"maxwritten_clean":      {COUNTER, "Number of times the background writer stopped a cleaning scan because it had written too many buffers", nil, nil},
					"buffers_backend":       {COUNTER, "Number of buffers written directly by a backend", nil, nil},
					"buffers_backend_fsync": {COUNTER, "Number of times a backend had to execute its own fsync call (normally the background writer handles those even when the backend does its own write)", nil, nil},
					"buffers_alloc":         {COUNTER, "Number of buffers allocated", nil, nil},
					"stats_reset":           {COUNTER, "Time at which these statistics were last reset", nil, nil},
				},
				true,
				0,
			},
			"pg_stat_database": {
				map[string]ColumnMapping{
					"datid":          {LABEL, "OID of a database", nil, nil},
					"datname":        {LABEL, "Name of this database", nil, nil},
					"numbackends":    {GAUGE, "Number of backends currently connected to this database. This is the only column in this view that returns a value reflecting current state; all other columns return the accumulated values since the last reset.", nil, nil},
					"xact_commit":    {COUNTER, "Number of transactions in this database that have been committed", nil, nil},
					"xact_rollback":  {COUNTER, "Number of transactions in this database that have been rolled back", nil, nil},
					"blks_read":      {COUNTER, "Number of disk blocks read in this database", nil, nil},
					"blks_hit":       {COUNTER, "Number of times disk blocks were found already in the buffer cache, so that a read was not necessary (this only includes hits in the PostgreSQL buffer cache, not the operating system's file system cache)", nil, nil},
					"tup_returned":   {COUNTER, "Number of rows returned by queries in this database", nil, nil},
					"tup_fetched":    {COUNTER, "Number of rows fetched by queries in this database", nil, nil},
					"tup_inserted":   {COUNTER, "Number of rows inserted by queries in this database", nil, nil},
					"tup_updated":    {COUNTER, "Number of rows updated by queries in this database", nil, nil},
					"tup_deleted":    {COUNTER, "Number of rows deleted by queries in this database", nil, nil},
					"conflicts":      {COUNTER, "Number of queries canceled due to conflicts with recovery in this database. (Conflicts occur only on standby servers; see pg_stat_database_conflicts for details.)", nil, nil},
					"temp_files":     {COUNTER, "Number of temporary files created by queries in this database. All temporary files are counted, regardless of why the temporary file was created (e.g., sorting or hashing), and regardless of the log_temp_files setting.", nil, nil},
					"temp_bytes":     {COUNTER, "Total amount of data written to temporary files by queries in this database. All temporary files are counted, regardless of why the temporary file was created, and regardless of the log_temp_files setting.", nil, nil},
					"deadlocks":      {COUNTER, "Number of deadlocks detected in this database", nil, nil},
					"blk_read_time":  {COUNTER, "Time spent reading data file blocks by backends in this database, in milliseconds", nil, nil},
					"blk_write_time": {COUNTER, "Time spent writing data file blocks by backends in this database, in milliseconds", nil, nil},
					"stats_reset":    {COUNTER, "Time at which these statistics were last reset", nil, nil},
				},
				true,
				0,
			},
			"pg_stat_database_conflicts": {
				map[string]ColumnMapping{
					"datid":            {LABEL, "OID of a database", nil, nil},
					"datname":          {LABEL, "Name of this database", nil, nil},
					"confl_tablespace": {COUNTER, "Number of queries in this database that have been canceled due to dropped tablespaces", nil, nil},
					"confl_lock":       {COUNTER, "Number of queries in this database that have been canceled due to lock timeouts", nil, nil},
					"confl_snapshot":   {COUNTER, "Number of queries in this database that have been canceled due to old snapshots", nil, nil},
					"confl_bufferpin":  {COUNTER, "Number of queries in this database that have been canceled due to pinned buffers", nil, nil},
					"confl_deadlock":   {COUNTER, "Number of queries in this database that have been canceled due to deadlocks", nil, nil},
				},
				true,
				0,
			},
			"pg_locks": {
				map[string]ColumnMapping{
					"datname": {LABEL, "Name of this database", nil, nil},
					"mode":    {LABEL, "Type of Lock", nil, nil},
					"count":   {GAUGE, "Number of locks", nil, nil},
				},
				true,
				0,
			},
			"pg_stat_replication": {
				map[string]ColumnMapping{
					"procpid":          {DISCARD, "Process ID of a WAL sender process", nil, MustParseSemverRange("<9.2.0")},
					"pid":              {DISCARD, "Process ID of a WAL sender process", nil, MustParseSemverRange(">=9.2.0")},
					"usesysid":         {DISCARD, "OID of the user logged into this WAL sender process", nil, nil},
					"usename":          {DISCARD, "Name of the user logged into this WAL sender process", nil, nil},
					"application_name": {LABEL, "Name of the application that is connected to this WAL sender", nil, nil},
					"client_addr":      {LABEL, "IP address of the client connected to this WAL sender. If this field is null, it indicates that the client is connected via a Unix socket on the server machine.", nil, nil},
					"client_hostname":  {DISCARD, "Host name of the connected client, as reported by a reverse DNS lookup of client_addr. This field will only be non-null for IP connections, and only when log_hostname is enabled.", nil, nil},
					"client_port":      {DISCARD, "TCP port number that the client is using for communication with this WAL sender, or -1 if a Unix socket is used", nil, nil},
					"backend_start": {DISCARD, "with time zone	Time when this process was started, i.e., when the client connected to this WAL sender", nil, nil},
					"backend_xmin":             {DISCARD, "The current backend's xmin horizon.", nil, nil},
					"state":                    {LABEL, "Current WAL sender state", nil, nil},
					"sent_location":            {DISCARD, "Last transaction log position sent on this connection", nil, MustParseSemverRange("<10.0.0")},
					"write_location":           {DISCARD, "Last transaction log position written to disk by this standby server", nil, MustParseSemverRange("<10.0.0")},
					"flush_location":           {DISCARD, "Last transaction log position flushed to disk by this standby server", nil, MustParseSemverRange("<10.0.0")},
					"replay_location":          {DISCARD, "Last transaction log position replayed into the database on this standby server", nil, MustParseSemverRange("<10.0.0")},
					"sent_lsn":                 {DISCARD, "Last transaction log position sent on this connection", nil, MustParseSemverRange(">=10.0.0")},
					"write_lsn":                {DISCARD, "Last transaction log position written to disk by this standby server", nil, MustParseSemverRange(">=10.0.0")},
					"flush_lsn":                {DISCARD, "Last transaction log position flushed to disk by this standby server", nil, MustParseSemverRange(">=10.0.0")},
					"replay_lsn":               {DISCARD, "Last transaction log position replayed into the database on this standby server", nil, MustParseSemverRange(">=10.0.0")},
					"sync_priority":            {DISCARD, "Priority of this standby server for being chosen as the synchronous standby", nil, nil},
					"sync_state":               {DISCARD, "Synchronous state of this standby server", nil, nil},
					"slot_name":                {LABEL, "A unique, cluster-wide identifier for the replication slot", nil, MustParseSemverRange(">=9.2.0")},
					"plugin":                   {DISCARD, "The base name of the shared object containing the output plugin this logical slot is using, or null for physical slots", nil, nil},
					"slot_type":                {DISCARD, "The slot type - physical or logical", nil, nil},
					"datoid":                   {DISCARD, "The OID of the database this slot is associated with, or null. Only logical slots have an associated database", nil, nil},
					"database":                 {DISCARD, "The name of the database this slot is associated with, or null. Only logical slots have an associated database", nil, nil},
					"active":                   {DISCARD, "True if this slot is currently actively being used", nil, nil},
					"active_pid":               {DISCARD, "Process ID of a WAL sender process", nil, nil},
					"xmin":                     {DISCARD, "The oldest transaction that this slot needs the database to retain. VACUUM cannot remove tuples deleted by any later transaction", nil, nil},
					"catalog_xmin":             {DISCARD, "The oldest transaction affecting the system catalogs that this slot needs the database to retain. VACUUM cannot remove catalog tuples deleted by any later transaction", nil, nil},
					"restart_lsn":              {DISCARD, "The address (LSN) of oldest WAL which still might be required by the consumer of this slot and thus won't be automatically removed during checkpoints", nil, nil},
					"pg_current_xlog_location": {DISCARD, "pg_current_xlog_location", nil, nil},
					"pg_current_wal_lsn":       {DISCARD, "pg_current_xlog_location", nil, MustParseSemverRange(">=10.0.0")},
					"pg_current_wal_lsn_bytes": {GAUGE, "WAL position in bytes", nil, MustParseSemverRange(">=10.0.0")},
					"pg_xlog_location_diff":    {GAUGE, "Lag in bytes between master and slave", nil, MustParseSemverRange(">=9.2.0 <10.0.0")},
					"pg_wal_lsn_diff":          {GAUGE, "Lag in bytes between master and slave", nil, MustParseSemverRange(">=10.0.0")},
					"confirmed_flush_lsn":      {DISCARD, "LSN position a consumer of a slot has confirmed flushing the data received", nil, nil},
					"write_lag":                {DISCARD, "Time elapsed between flushing recent WAL locally and receiving notification that this standby server has written it (but not yet flushed it or applied it). This can be used to gauge the delay that synchronous_commit level remote_write incurred while committing if this server was configured as a synchronous standby.", nil, MustParseSemverRange(">=10.0.0")},
					"flush_lag":                {DISCARD, "Time elapsed between flushing recent WAL locally and receiving notification that this standby server has written and flushed it (but not yet applied it). This can be used to gauge the delay that synchronous_commit level remote_flush incurred while committing if this server was configured as a synchronous standby.", nil, MustParseSemverRange(">=10.0.0")},
					"replay_lag":               {DISCARD, "Time elapsed between flushing recent WAL locally and receiving notification that this standby server has written, flushed and applied it. This can be used to gauge the delay that synchronous_commit level remote_apply incurred while committing if this server was configured as a synchronous standby.", nil, MustParseSemverRange(">=10.0.0")},
				},
				true,
				0,
			},
			"pg_stat_archiver": {
				map[string]ColumnMapping{
					"archived_count":     {COUNTER, "Number of WAL files that have been successfully archived", nil, nil},
					"last_archived_wal":  {DISCARD, "Name of the last WAL file successfully archived", nil, nil},
					"last_archived_time": {DISCARD, "Time of the last successful archive operation", nil, nil},
					"failed_count":       {COUNTER, "Number of failed attempts for archiving WAL files", nil, nil},
					"last_failed_wal":    {DISCARD, "Name of the WAL file of the last failed archival operation", nil, nil},
					"last_failed_time":   {DISCARD, "Time of the last failed archival operation", nil, nil},
					"stats_reset":        {DISCARD, "Time at which these statistics were last reset", nil, nil},
					"last_archive_age":   {GAUGE, "Time in seconds since last WAL segment was successfully archived", nil, nil},
				},
				true,
				0,
			},
			"pg_stat_activity": {
				map[string]ColumnMapping{
					"datname":         {LABEL, "Name of this database", nil, nil},
					"state":           {LABEL, "connection state", nil, MustParseSemverRange(">=9.2.0")},
					"count":           {GAUGE, "number of connections in this state", nil, nil},
					"max_tx_duration": {GAUGE, "max duration in seconds any active transaction has been running", nil, nil},
				},
				true,
				0,
			},
		},
		QueryOverrides: QueryOverrides{
			"pg_locks": {
				{
					MustParseSemverRange(">0.0.0"),
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
				         ('accessexclusivelock')
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
					MustParseSemverRange(">=10.0.0"),
					`
			SELECT *,
				(case pg_is_in_recovery() when 't' then null else pg_current_wal_lsn() end) AS pg_current_wal_lsn,
				(case pg_is_in_recovery() when 't' then null else pg_wal_lsn_diff(pg_current_wal_lsn(), pg_lsn('0/0'))::float end) AS pg_current_wal_lsn_bytes,
				(case pg_is_in_recovery() when 't' then null else pg_wal_lsn_diff(pg_current_wal_lsn(), replay_lsn)::float end) AS pg_wal_lsn_diff
			FROM pg_stat_replication
			`,
				},
				{
					MustParseSemverRange(">=9.2.0 <10.0.0"),
					`
			SELECT *,
				(case pg_is_in_recovery() when 't' then null else pg_current_xlog_location() end) AS pg_current_xlog_location,
				(case pg_is_in_recovery() when 't' then null else pg_xlog_location_diff(pg_current_xlog_location(), replay_location)::float end) AS pg_xlog_location_diff
			FROM pg_stat_replication
			`,
				},
				{
					MustParseSemverRange("<9.2.0"),
					`
			SELECT *,
				(case pg_is_in_recovery() when 't' then null else pg_current_xlog_location() end) AS pg_current_xlog_location
			FROM pg_stat_replication
			`,
				},
			},

			"pg_stat_archiver": {
				{
					MustParseSemverRange(">=0.0.0"),
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
					MustParseSemverRange(">=9.2.0"),
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
					MustParseSemverRange("<9.2.0"),
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
		},
	},
	ByServer: nil,
}