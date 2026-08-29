# KAN-363 — a change's spectre tree in every repo it touches

The approved design for this change is its own artifact, and is canonical there rather than copied
here:

- `spectre/changes/kan-363-let-a-change-s-spectre-tree-live-in-every-repo/design.md` — the context,
  the decisions and what each one ruled out.
- `spectre/changes/kan-363-let-a-change-s-spectre-tree-live-in-every-repo/proposal.md` — why the
  change exists and what it changes.

The change spans two repositories: `spectre` (the `link.md` tree format, `spectre link`, and the
`validate`/`list` handling) and `agents` (the guards that follow a link, and the flow contracts that
create one and land the repositories in its recorded order).
