# self-review-kan-998-bugfixes-superset-header-focus-mode

## Why

kan-998's self-review pass found six pipeline defects, none `big`, planned here as one self-review
fix change (`design.md`, **Context**):

- **F-1.1** — `flow.visual-verify` step 3 creates the `regression checkout`'s worktree and adds it
  to the set, but never writes it into the canonical `link.md`'s `## Merge order`, which every
  repository of the set must be named in (`skills/flow/cross-repo-worktrees.md`); the self-review
  bundle's `peerLogs` then omits that repository's log.
- **F-1.2** — a removal task's absence sweep covers only the task's own paths, so a removed UI
  label the `regression checkout` drives stays referenced there; kan-998's regression suite
  failed on it.
- **F-1.3** — the skills invoke scripts by bare name (`integrate-change.sh`, `project-get.sh`, …),
  and none of them is on `PATH`; `~/.local/bin`, which holds `flow`, is.
- **F-1.4** — a `live-verification.md` stating "Before figure: not measured on a merge-base stack"
  passed, although **Live check** (`skills/flow/verify-and-handoff.md`) requires the before figure
  from a merge-base worktree; nothing reads the record.
- **F-4.1** — a `<spec>.mockups` sidecar line naming a screenshot no test takes is found only when
  compose fails, after the full capture run.
- **F-4.2** — two byte-identical baselines in one spec's snapshot directory (kan-998's pad-up and
  pad-down) were caught only by a reviewer.

## What changes

- `skills/flow/visual-verify.md` step 3 runs the recipe's merge-order record with the worktree it
  adds.
- The absence-sweep paragraph of `skills/flow/brainstorm-planner.md` extends the sweep, and the
  plan, to a declared `regression checkout`.
- `setup.sh global` writes an exec shim for every `scripts/*.sh` entry script into
  `~/.local/bin`, so a bare script name resolves.
- New guard `check-live-verification.sh`, run by **Live check** after the record is written.
- New guard `check-sidecar-reach.sh`, run in `flow.visual-verify` step 3's pre-flight.
- New guard `check-identical-baselines.sh`, run by the verifier's step 9, each identical group
  explained in its report.
