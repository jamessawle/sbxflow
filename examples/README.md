# Examples

These examples show how repositories can declare Docker Sandbox kit sources and
their ordered use.

- [`personal-site/sbxflow.yaml`](personal-site/sbxflow.yaml) translates the
  existing `personal-site/.docker-sbx/sandbox.sh` configuration.
- [`source-types/sbxflow.yaml`](source-types/sbxflow.yaml) demonstrates Git,
  OCI, and local sources. A local source can select either a directory or a ZIP
  kit.

Each configuration separates two concerns:

- `sources` names and pins reusable kit locations.
- `use` selects kit artifacts in the order passed to Docker Sandboxes.

Kit-source trust is derived from the selected sources and is not declared in the
configuration.
