## Context

- In-run pipeline fix 19 for gymie `kan-940-step-5-frontend-cardio-row-on-the-kan-928`, per
  **Pipeline defects found mid-run** (`skills/flow-contracts/pipeline.md`), carrying the operator's
  2026-10-17 ruling "update flow not to ask such questions again". Checked against this worktree at
  `bc79a4d6`.
- Defect 1 — stale snapshot baselines on the base. After a sync's targeted recapture
  (`skills/flow/sync-onto-base.md`, landed by fix 17), tests in specs the change never touched
  failed on the base's own baselines, and the run asked whether to re-baseline. No contract says
  what to do: **the sweep** (`skills/flow-contracts/known-bugs.md`) blocks a pre-existing failure and
  sends it to **The loop**, and `skills/flow/visual-verify-verifier.md` step 7 blocks every non-zero
  `verify` exit, so the only answer left was a question.
- Defect 2 — run 2's cleanup stops on trivial residue. The classifier is check 4 of
  `stats/internal/guard/removechangeworktrees.go` (shim `scripts/remove-change-worktrees.sh`);
  `scripts/integrate-change.sh` `cmd_cleanup` maps its exit 3 to `STOP: disclose`. Check 4 counts
  only `rcwRegeneratable`'s path list; `.DS_Store` and gymie-playwright's root `NN-*.png` (written by
  `screenshots.spec.ts` beside itself, ignored by `/[0-9][0-9]-*.png`) fall to "unpreserved
  unclassified", which stops. A file byte-identical to the main checkout's — the `.env` included —
  is already printed `— preserved: identical in <main>` and never stops (the existing case
  `an entry identical in the main checkout proceeds`), so it needs no change: the stop came from the
  `.DS_Store` files and the PNGs.
- Blast radius: 47 files in this repository outside `spectre/changes/archive/` cite
  <!-- measured: grep -rl --exclude-dir=.git -e remove-change-worktrees -e sync-onto-base -e known-bugs.md -e visual-verify-verifier . | grep -v spectre/changes/archive | sort -u | wc -l in this worktree at bc79a4d6 -->
  `remove-change-worktrees`, `sync-onto-base`, `known-bugs.md` or `visual-verify-verifier`; the fix
  touches 7 of them plus this change's `verbatim-moves.txt` — under `big`'s 60.
- Concurrent fix 18 edits `scripts/kickoff-worktree.sh`, `scripts/integrate-change.sh` (`prepare`),
  `skills/flow/integrate.md` and `skills/flow-contracts/cross-repo-worktrees.md`. This change touches
  none of them: no overlap.
- Saved memory: no file in the gymie project's auto-memory directory encodes either ask (grep
  `baseline`, `snapshot`, `DS_Store`, `disclos`, `UNCLASSIFIED`, `preserved`: hits unrelated, and
  `gymie-screenshot-spec.md` only records where sweep output goes), so none is updated.
- Spiked at plan time, every edit reverted: both diffs in `tasks.md` apply at `bc79a4d6`;
  `go test ./internal/guard -run 'TestRemoveChangeWorktrees|TestRcwRegeneratable'` passes with Task
  1's diff and fails with `removechangeworktrees.go` reverted (`TestRcwRegeneratable` and
  `a .DS_Store and a Playwright checkout's root capture proceed`); `gofmt -l`, `go vet
  ./internal/guard`, `check-references.sh`, `check-markdown-integrity.py`,
  `check-dispatch-paragraphs.sh`, `check-installed-citations.sh`, `check-ask-silence.sh` and
  `check-hand-notes-in-step.sh` exit 0, and `check-verbatim-moves.sh` reports 0 violations with this
  change's `verbatim-moves.txt`.
- Playwright selects one test by location: `npx playwright test kan-339-route-a.spec.ts:149 --list`
  <!-- measured: npx playwright test kan-339-route-a.spec.ts:149 --list in gymie-playwright at eb0dbb6 -->
  prints `Total: 1 test in 1 file`.

## Decisions

### A stale snapshot baseline is re-baselined, canonical in the sweep

**ID:** stale-baseline-rebaselined
**Status:** active
**Chosen:** a screenshot mismatch in a spec whose file and `*-snapshots/` directory the change's
`git diff <merge-base>..HEAD` leaves untouched is re-baselined by whoever ran it, as its own
`test(fidelity): re-baseline <spec basenames>` commit, then re-run; any other failure kind is
reported and blocks. Stated once in **the sweep** (`skills/flow-contracts/known-bugs.md`), cited by
`skills/flow/sync-onto-base.md` and `skills/flow/visual-verify-verifier.md` step 7 —
`skills/flow/visual-verify.md` reads only the verifier's report, so it needs no edit.
**Considered:** attributing by re-running the failing test against a build of the merge base — a
second stack build per failure for an answer the untouched diff already gives; attributing by the
`specs` selection (a failure in a spec `specs` does not select) — a touched file declaring no
spec-reached tag selects every spec, which routes exactly the operator's case back to a block and an
ask; stating the rule in each of the three files — three copies of one normative paragraph.

### The re-baseline is targeted by test location

**ID:** rebaseline-by-test-location
**Status:** active
**Chosen:** `capture` with `<spec>` substituted by each failing test's `<file>:<line>`, staging only
the PNGs the failing run named and restoring every other file the capture rewrote.
**Considered:** the failing specs whole with `--update-snapshots` — it rewrites passing snapshots
too, committing noise the operator's "only failing snapshots" excludes.

### A Playwright checkout's root PNG is a screenshot spec's output

**ID:** playwright-root-png-regeneratable
**Status:** active
**Chosen:** check 4 counts an ignored `.png` with no `/` in its path as regeneratable when the
worktree root holds a `playwright.config.*`; a `.png` in any subdirectory stays unclassified.
**Considered:** reading the project's declared `regression checkout` — the guard is given only
`<repo>`, so `scripts/integrate-change.sh` `cmd_cleanup` would need a new argument, in the file
concurrent fix 18 edits, to reach the same answer; every ignored image in a Playwright checkout —
it would remove gymie-playwright's kept `screenshots/run-*` reference folders, which nothing
re-creates.

### `.DS_Store` is regeneratable by name

**ID:** ds-store-regeneratable
**Status:** active
**Chosen:** `rcwRegeneratable` matches a path whose last component is `.DS_Store`, anywhere — the one
name the bucket's "no allowlist of names" sentence now admits.
**Considered:** leaving it to the byte-identical rule — Finder writes a different `.DS_Store` per
directory view, so it is rarely identical to the main checkout's.

### Entries identical in the main checkout need no change

**ID:** identical-entries-unchanged
**Status:** active
**Chosen:** no change: such an entry is already printed preserved and never stops, pinned by the
existing case `an entry identical in the main checkout proceeds`.
**Considered:** renaming the preserved `UNCLASSIFIED:` line — an output-contract change in the
guard and `integrate-change.sh`'s relay that closes no stop.

### First review effort and fixer grouping

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-low` — two file-disjoint tasks in one group, one fixer dispatch (key
`pipeline-fix-19-fix`) applying diffs verified at plan time; Task 1's behaviour is pinned by tests
that fail without it, Task 2 is prose already passing every lint line.
**Considered:** `flow-medium` — the Go change is two path predicates, spiked and tested.

## Open questions
