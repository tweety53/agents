## Context

- In-run pipeline fix 17 for gymie `kan-940-step-5-frontend-cardio-row-on-the-kan-928`, per
  **Pipeline defects found mid-run** (`skills/flow-contracts/pipeline.md`). Checked against this
  worktree at `2d9dc724`.
- The defect: `skills/flow/sync-onto-base.md` l.40 lists "a binary file" among the conflicts with
  no honest resolution. The guard (`stats/internal/guard/syncontobase.go`, shim
  `scripts/sync-onto-base.sh`) does not classify unmerged paths at all — it prints one
  `CONFLICT:` line and exits 1; `scripts/integrate-change.sh` reads only that exit (l.503, l.747)
  and sends the parent to the Conflict bullet.
- Blast radius: 20 files in this repository outside `spectre/changes/archive/` cite
  <!-- measured: grep -rl sync-onto-base in this worktree at 2d9dc724, spectre/changes/archive excluded -->
  `sync-onto-base` (`grep -rl`); the fix touches 4 of them plus this change's
  `verbatim-moves.txt` — far under `big`'s 60. `/flow-fast` (`skills/flow-fast/SKILL.md` l.395)
  and the panel fix round cite the Conflict bullet by name and inherit the change unedited.
  `sync-panel-base.sh` prints the same `CONFLICT:` shape but never resolves a conflict
  (`skills/flow/review-panel.md` l.74), so it is out of scope.
- Saved memory: no file in the gymie project's auto-memory directory encodes a binary conflict as a
  stop (grep `binary`, `snapshot`: two unrelated hits), so none is updated.
- Spiked at plan time, every edit reverted: the diff in `tasks.md` applies at `2d9dc724`;
  `go test ./internal/guard -run TestSyncOntoBase` passes with it and fails with
  `syncontobase.go` reverted (the new case wants the two `RECAPTURE:` lines);
  `go test ./internal/guard`, `gofmt -l .`, `go vet ./internal/guard`, `check-references.sh`,
  `check-markdown-integrity.py`, `check-dispatch-paragraphs.sh`, `check-installed-citations.sh`,
  `check-ask-silence.sh`, `check-normative-inventory.sh` and `check-hand-notes-in-step.sh` exit 0,
  and `check-verbatim-moves.sh` reports 0 violations with this change's `verbatim-moves.txt`.

## Decisions

### The guard names the recapture paths

**ID:** guard-names-recapture-paths
**Status:** active
**Chosen:** `syncOntoBase` prints `RECAPTURE: <path>` after its `CONFLICT:` line, in that line's
order, for each unmerged path that is a PNG under any directory component ending `-snapshots`, or
whose basename is `full-app-suite.zip`; the exit stays 1, and a test case pins it.
**Considered:** the Conflict bullet alone classifying the paths — no test could fail without the
fix; a new exit code for a recapture-only conflict — both exit-1 callers in `integrate-change.sh`
would need a branch for what is still an in-place resolution; detecting binary content — the path
is what the operator's rule names, and a snapshot is binary by construction.

### The both-sides build merges the app's own base

**ID:** recapture-build-own-base
**Status:** active
**Chosen:** a throwaway branch in the change's app worktree merging that repository's own base tip,
served per steps 5–6 of `skills/flow/visual-verify.md`, never committed to the change's branch or
pushed, deleted after.
**Considered:** `origin/$BASE` — `$BASE` is the conflicting checkout's base (gymie playwright
`main`), not the frontend's (`develop`); locating the exact upstream frontend commit behind each
baseline — the base tip carries it, and no mechanism maps a PNG to a commit.

### First review effort and fixer grouping

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-low` — one task in one group, one fixer dispatch (key `pipeline-fix-17-fix`)
applying a diff verified at plan time, its behaviour pinned by a test that fails without it.
**Considered:** `flow-medium` — the Go change is a pure path predicate, and the spiked prose already
passes every lint line.

## Open questions
