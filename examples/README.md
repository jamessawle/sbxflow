# Examples

These examples show how repositories can declare Docker Sandbox kit sources and
their ordered use.

Every declaration is checked against sbxflow's published
[Draft 2020-12 JSON Schema](../schema/sbxflow.schema.json). From an example
directory, run `sbxflow validate` to discover and validate its declaration.

- [`node-project/sbxflow.yaml`](node-project/sbxflow.yaml) demonstrates
  sandbox-scoped network resources and ordered initialization for a Node.js
  repository.
- [`source-types/sbxflow.yaml`](source-types/sbxflow.yaml) demonstrates Git,
  OCI, and local sources. A local source can select either a directory or a ZIP
  kit. Its local paths are illustrative; provide matching artifacts before
  running local Docker kit validation.

Each configuration separates two concerns:

- `sources` names and pins reusable kit locations.
- `use` selects kit artifacts in the order passed to Docker Sandboxes.

The examples also select workspace behavior explicitly. `node-project` uses
the recommended `clone` mode, where changes remain private until transferred
through Git. `source-types` uses `direct` to demonstrate immediate host
visibility. Omission remains accepted as a compatibility behavior equivalent
to direct mode, but emits a warning because the implicit mode is planned to
change before 1.0.

The Node project example also declares creation-only `hooks.initialize`
commands. Its bounded shell loop waits for the `language-node-npm` kit's
sandbox-local `node_modules` mount before `npm ci` writes dependencies. Hooks
run non-interactively in order and forward output; use an explicit shell for
shell syntax. Changed hooks require `up --recreate`. A failed or cancelled hook
removes the new sandbox and declared scoped resources. It cannot undo changes
to a direct host-mounted workspace; changes confined to a successfully removed
private clone are discarded, so bootstrap commands should be safe to retry.

Kit-source trust is derived from the selected sources and is not declared in the
configuration. Remote Git and OCI selections are normalized without network
access. Only selections whose declared source is `local` invoke
`sbx kit validate`; file packaging such as ZIP does not determine provenance.

Validation reports the effective `kit.allowedSources` list in first-selection
order, beginning with `docker.io/`, and reports `kit.allowLocalKits: true` only
when a local source is selected. These values describe the settings sbxflow
will apply to its own future lifecycle subprocesses; validation does not modify
Docker's global settings or sandbox state.
