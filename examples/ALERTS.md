# Alerts example

## Max Connections

* Warning when Postgresql reach 70% of max connections 
* Critical when Postgresql reach 90% of max connections
* Also trigger when `pg_runtime_variable_max_connections` is empty (it happens when prometheus cannot scrap exporter).

```
ALERT PostgresqlTooManyConnections
  IF sum(pg_stat_activity_count) >  sum(pg_runtime_variable_max_connections * 0.7 )
  OR absent(pg_runtime_variable_max_connections)
  FOR 5m
  LABELS {
    service = "postgresql",
    severity = "warning",
  }
  ANNOTATIONS {
    summary = "Postgresql max client connections",
    description = "Postgresql reached 70% of the max client connections.",
  }
  
ALERT PostgresqlTooManyConnections
  IF sum(pg_stat_activity_count) >  sum(pg_runtime_variable_max_connections * 0.9 )
  OR absent(pg_runtime_variable_max_connections)
  FOR 5m
  LABELS {
    service = "postgresql",
    severity = "critical",
  }
  ANNOTATIONS {
    summary = "Postgresql max client connections",
    description = "Postgresql reached 90% of the max client connections.",
  }
```
