## Context

See `proposal.md` for motivation. The declaration adapter currently decodes a sandbox without workspace configuration, lifecycle planning always derives the declaration directory as a scalar workspace path, and the Docker adapter renders that path into a private `.sbxenv.yaml`. Docker Sandboxes 0.39.x represents direct mode as a scalar path and clone mode as a workspace object containing `path` and `clone: true`.

Validation reports currently contain errors only. The `validate` command renders the full report, while `up` consumes the same report but prints only its successful-validation status before lifecycle work. The architecture requires declaration decoding to remain in the outbound declaration adapter, reusable defaulting policy in the configuration domain, lifecycle coordination in application, backend-neutral creation inputs in the sandbox port, Docker document rendering in the outbound sbx adapter, and output formatting in the inbound CLI adapter.

## Goals / Non-Goals

**Goals:**

- Preserve whether workspace mode was omitted long enough to emit a migration warning while also deriving an effective mode for lifecycle behavior.
- Keep the public mode vocabulary independent of Docker's experimental boolean representation.
- Carry warnings through the shared validation result so `validate` and `up` apply one policy consistently.
- Preserve byte-for-byte-equivalent direct-mode environment semantics for legacy declarations.

**Non-Goals:**

- Selecting a workspace path other than the declaration directory.
- Proactively determining whether Docker can clone the repository.
- Reconciling mode for an existing sandbox during ordinary `up`.
- Changing the implicit mode to `clone` in this change.
- Adding CPU, memory, additional-workspace, or other Docker environment options.

## Decisions

### Use an extensible public workspace object with a mode enum

The public declaration will accept `sandbox.workspace.mode` with `direct` and `clone`. `workspace` will reject an empty object with `minProperties: 1` and unknown properties with `additionalProperties: false`. The workspace model will preserve presence separately from its effective value, using an optional workspace object or equivalent presence-aware representation, so omission can produce a warning.

The domain will normalize omission to effective `direct` for this release. It will not expose Docker's `clone` boolean outside the outbound adapter. This leaves room for a later `workspace.path` or additional mutually exclusive modes without making Docker's experimental document shape the sbxflow contract.

Alternatives considered:

- `workspace.clone: true` mirrors Docker but makes a boolean an awkward basis for future modes.
- Requiring `mode` unconditionally whenever the object exists prevents a future path-only declaration from inheriting the effective mode. With only `mode` recognized today, `minProperties: 1` still makes a present workspace require it in practice.

### Derive the repository path independently from workspace mode

Lifecycle planning will continue to derive the absolute workspace source from the directory containing `sbxflow.yaml`. The sandbox port environment will carry that path and the effective mode as separate backend-neutral inputs. A future path-selection change can replace path derivation without changing the mode contract.

The Docker adapter will render canonical documents by behavior:

```yaml
# direct, whether omitted or explicit
workspace: /absolute/repository/path

# clone
workspace:
  path: /absolute/repository/path
  clone: true
```

Rendering explicit direct mode as an object with `clone: false` was considered, but it creates two representations for identical behavior and changes the legacy document unnecessarily.

### Keep clone viability Docker-authoritative

Structural validation will reject unsupported modes, but sbxflow will not inspect Git metadata or invoke Git to predict clone viability. Docker's creation failure and diagnostics will remain visible through the existing provisioning error path.

Adding proactive Git checks was considered, but worktrees, bare repositories, filesystem permissions, and Docker-specific eligibility could cause sbxflow to disagree with the actual authority. Targeted checks can be added later if field experience shows Docker's diagnostics are insufficient.

### Add first-class non-fatal warnings to configuration validation

Resolution and validation results will gain a warning collection distinct from errors. The domain will add the omitted-mode warning after a declaration has been structurally decoded and before external local-kit validation. The application validator will preserve warnings even when a later validation phase adds errors. `Valid()` will continue to depend only on errors.

Warnings will be plain advisory findings with stable ordering. The initial producer is only workspace-mode omission; unrelated advisory policy will not be added in this change.

The CLI will use shared warning rendering for both commands:

- `validate` will render warnings on standard error and retain a zero exit status for a warning-only report.
- `up` will render warnings on standard error alongside its successful-validation status before sandbox inspection or mutation, then continue when no errors exist.
- When errors exist, diagnostics may include both warnings and errors, but lifecycle work remains gated by `Valid()`.

Keeping the warning in documentation only was rejected because users may upgrade without revisiting documentation. Warning only from `validate` was rejected because `up` is the routine workflow and must carry the migration signal.

### Keep teardown identity independent of workspace mode

Destroy and shared removal do not need the declared mode to target Docker resources, so the minimal lifecycle-target parser will not gain workspace fields. Destructive documentation and prompts will use generic language covering persisted state and work stored only inside the sandbox. This remains accurate for both modes without coupling teardown to presentation-only mode detection.

Mode-specific prompts were considered but would expand the minimal teardown model without changing removal behavior.

### Treat mode as creation-time configuration

The current lifecycle branching remains intact. A missing sandbox or successful recreation uses the effective mode from the validated plan. Ordinary `up` enters an exact-name existing sandbox without inspecting or reconciling its mode. Initialization commands execute in Docker's effective workspace, so direct-mode changes reach host files and clone-mode changes remain private until transferred through Git.

Rollback continues to remove a newly created sandbox after initialization failure. Documentation and generic warnings will distinguish unreverted direct host changes from private-clone changes discarded by successful cleanup.

## Risks / Trade-offs

- **[Users ignore the omission warning]** → Make explicit mode the normal form in README and examples, and state both the current behavior and planned pre-1.0 change in the warning.
- **[Docker changes its experimental workspace document]** → Confine `clone: true` translation to the outbound sbx adapter and retain sbxflow's mode vocabulary across adapter updates.
- **[Warning output disrupts scripts that assume clean standard error]** → Keep warnings non-fatal, deterministic, and limited to declarations that rely on an implicit behavior scheduled to change.
- **[Clone creation fails for a repository Docker cannot clone]** → Preserve Docker's diagnostic output and document that Docker is authoritative for eligibility.
- **[Private work is lost during destroy or recreation]** → Expand generic destructive messaging and lifecycle documentation to include work stored only inside the sandbox.
- **[A later default switch silently changes old declarations]** → Encourage and warn for explicit modes now; implement the actual switch as a separate announced breaking change rather than in this change.

## Migration Plan

1. Release the optional mode with omission normalized to `direct` and an advisory warning.
2. Update first-party examples and primary documentation to select a mode explicitly, recommending `clone` for new declarations and `direct` when immediate host visibility is intended.
3. Allow existing declarations to continue operating while their maintainers resolve the warning by declaring their intended mode.
4. Handle the future implicit switch to `clone` in a separate pre-1.0 breaking change with release-note and configuration-contract review.

Rollback of this change removes the new field and warning support; declarations that have adopted `sandbox.workspace` would then fail against the older strict schema, so release notes must continue to advise users about version compatibility when downgrading.
