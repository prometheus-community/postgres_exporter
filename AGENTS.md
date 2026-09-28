# Contributor guidance for AI agents

Keep changes focused and validate claims with repository evidence. Read
`README.md` and `CONTRIBUTING.md` for user-facing behavior and contributor
setup;

## Evidence and review

- Verify Postgres behavior and version availability with official Postgres
  documentation. Verify specificities of different Postgres flavors with 
  their specific documentation websites.
- Verify metric semantics against official [Prometheus documentation](https://prometheus.io/docs/practices/naming/) and, when
  necessary, upstream `client_golang`, Go, or database-driver source.
- Support findings with concrete code, unit tests, integration tests, documentation, or reproduced
  behavior. Separate correctness problems from optional improvements.
- Use `.github/workflows/integration.yml` as the source of
  truth for the versions currently exercised by CI.

## Architecture and compatibility

- Metric names, types, labels, help text, and meanings are public compatibility
  surfaces. Avoid changing them without explicit justification and migration
  consideration.
- Collector flags and configuration keys are also compatibility surfaces.
- Check documentation, recording rules, alerts, and dashboards when changing
  metrics or flags.
- Review every new label for boundedness and cardinality risk.
- `collector.Runtime` combines the legacy `exporter.Exporter` path with the
  native `collector.PostgresCollector` path. Trace both before concluding that
  a configuration, timeout, datasource, or collector change is complete.
- Check both single-target `/metrics` and multi-target `/probe` behavior when
  connection or configuration code changes.
- We try to follow https://github.com/prometheus/prometheus-opentelemetry-collector/blob/main/docs/embeddable-exporters.md, to enable
  the codebase present at https://github.com/prometheus/prometheus-opentelemetry-collector/tree/main/receivers/postgres. 
  It is not a strict rule and we can make small breaking changes to the downstream project
  if there's a way for them to work around the breakage.

## PostgreSQL collectors

- Follow the nearest collector and its tests as the implementation pattern.
- Account deliberately for PostgreSQL-version differences, nullable columns,
  unsupported views, extension requirements, and required database grants.
- Use context-aware database operations and preserve collection-timeout
  cancellation through connection setup, queries, iteration, and retry delays.
- Check query, scan, and iteration errors, including `rows.Err()`. Do not emit
  or cache partial metric sets after an iteration error.
- Tests should assert metric names, types, labels, and values—not only that a
  scrape succeeds.
- Use a real PostgreSQL integration test when behavior depends on server query
  semantics, permissions, extensions, or version-specific schemas; `sqlmock`
  alone is insufficient for those cases.
- The `exporter` package is considered legacy and we're slowly deprecating and
  deleting that codebase. Avoid adding extra functionality unless necessary to 
  advance the `collectors` package.

## Secrets

- Never log credentials, unredacted DSNs, authentication tokens, or secret-file
  contents.

## Validation and generated files

- Use `.github/workflows/` and `.promu.yml` as the source of truth for CI
  toolchains rather than hard-coding versions here.
- `Makefile.common` is shared Prometheus infrastructure; generic changes belong
  in its upstream repository rather than being maintained locally.
- For mixin changes, edit sources under `postgres_mixin/` and use its Makefile
  to lint and generate outputs. Inspect the resulting diff for unexplained
  generated changes.