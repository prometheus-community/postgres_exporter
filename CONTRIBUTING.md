# Contributing

See [Building and running](README.md#building-and-running) and
[Running the tests](README.md#running-the-tests) for development setup.
Use the Go toolchain specified in `.github/workflows/ci.yml` and `.promu.yml`.

## Changelog entries

For user-visible changes, create a release-note fragment:

```sh
make chlog-new NAME=short-description
```

Edit `.chloggen/short-description.yaml` and include it in your PR. Use a unique
descriptive filename; it does not need to match the PR number. Fill in `issues`
with the PR or tracking issue number (you can add the PR number after opening it).
Do not edit `CHANGELOG.md` directly for ordinary PRs.

| `change_type` | Generated prefix | Use for |
| --- | --- | --- |
| `breaking` | `[CHANGE]` | Changes requiring action or affecting compatibility |
| `deprecation` | `[CHANGE]` | Announced removals and their replacements |
| `new_component` | `[FEATURE]` | New collectors |
| `enhancement` | `[FEATURE]` | Features and improvements |
| `bug_fix` | `[BUGFIX]` | Corrections to existing behavior |

`component` is required and must be one of the following configured values:

| `component` | Use for |
| --- | --- |
| `collector` | Native collectors, collection runtime, discovery, and shared collection behavior |
| `config` | Configuration, datasource settings, authentication, and reload behavior |
| `web` | HTTP endpoints, TLS, and scrape request handling |
| `mixin` | Dashboards, recording rules, and alerts |
| `build` | Tooling, dependencies, CI, containers, and packaging |
| `all` | Changes spanning the project |

Put specific collector names in `note`, for example
`"stat_database: report errors encountered while reading query results"`.
Documentation-only changes can use the changelog exemptions below.
Author attribution is not required.

Use `subtext: |-` for Markdown upgrade instructions, warnings, tables, or examples.
The default template puts breaking-change and deprecation subtext in an
**Upgrade notes** section above the entry list. Other subtext stays beneath its
entry. A breaking change without subtext still appears in the entry list, so
include migration instructions explicitly when users need them.

```sh
make chlog-validate
make chlog-preview
```

The reusable OpenTelemetry PR workflow checks fragment changes, validates and
previews entries, checks preview links, and normally rejects direct changelog
edits. Its default exemptions are:

- `[chore]` anywhere in the PR title;
- the `Skip Changelog` or `dependencies` label;
- Dependabot and Renovate actors.

Use exemptions for internal maintenance or release preparation, and include
entries for dependency changes that affect users. The reusable workflow checks
for a changed `.chloggen/*.yaml` file; it does not strictly require a newly added
entry or exclude config/template changes. Exempt PRs skip its validation too.
Maintainers must configure its check as required in repository rules for merge
enforcement.

## Preparing a release

Validate, then generate the release section before tagging:

```sh
make chlog-update VERSION="0.x.0 / YYYY-MM-DD"
```

Replace the version and date with the intended release. This inserts a section
below `<!-- next version -->` in `CHANGELOG.md` and removes consumed fragments.
Review and commit both the changelog and fragment deletions. Keep the insertion
marker for future releases. The generator is pinned in the local Makefile and
does not change exporter runtime dependencies.

## Changelog example

See the [rendered example](.chloggen/examples/preview.md) and
[preview instructions](.chloggen/examples/README.md). The illustrative entries
are outside the pending-entry directory glob and are never included in normal
releases.
