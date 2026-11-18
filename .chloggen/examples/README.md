# Changelog example

These examples are temporary for maintainer discussion and must be removed
before merging the proposal, along with references to them.

[View the rendered example](preview.md). Breaking-change and deprecation details
appear in **Upgrade notes** above the entry list. Other details stay beneath
their entries. Both `new_component` and `enhancement` render as `[FEATURE]`;
author attribution is omitted.

These are illustrative historical changes, not pending release notes. Normal
generation reads only `.chloggen/*.yaml`, so these nested entries are excluded.
Validation uses chloggen directly with the same component allowlist as the main
configuration.

Preview the example:

```sh
make chlog-preview CHLOGGEN_CONFIG=.chloggen/examples/config.yaml
```

The checked-in preview was generated with:

```sh
go run go.opentelemetry.io/build-tools/chloggen@v0.30.0 \
  --config .chloggen/examples/config.yaml \
  update --dry --version "0.x.0 / YYYY-MM-DD"
```

The tool's `Generated changelog updates for user:` preamble and whitespace-only
indentation on blank lines are omitted from the preview file. The example config
targets the preview file, which has no release insertion marker; use it with
`chlog-preview`.
