# Alerts example

## Max Connections

* Warning when reach 70% of max connections 
* Critical when reach 90% of max connections
* Also trigger when `pg_runtime_variable_max_connections` is empty (it happens when prometheus cannot scrap exporter).

```
ALERT PostgresqlTooManyConnections
  IF sum(pg_runtime_variable_max_connections * 0.7) - sum(pg_stat_activity_count) == 0
  OR absent(pg_runtime_variable_max_connections)
  FOR 5m
  LABELS {
    service = "postgresql",
    severity = "warning",
  }
  ANNOTATIONS {
    summary = "Postgresql has too many connections",
    description = "Postgresql reach 70% of the max connections",
  }
  
ALERT PostgresqlTooManyConnections
  IF sum(pg_runtime_variable_max_connections * 0.9) - sum(pg_stat_activity_count) == 0
  OR absent(pg_runtime_variable_max_connections)
  FOR 5m
  LABELS {
    service = "postgresql",
    severity = "critical",
  }
  ANNOTATIONS {
    summary = "Postgresql has too many connections",
    description = "Postgresql reach 90% of the max connections",
  }
```
