## Context

The five findings of `proposal.md` are planned as one self-review change on kan-923's own branch
`spectre/kan-923-flow-improvement-port-the-python-backed-guards`, in its worktree
`/Users/tweety53/Projects/agents-worktrees/kan-923-flow-improvement-port-the-python-backed-guards`,
fix-base `5611091f`. No path a finding names changed between the branch and `origin/main`
(`444634d0`), and no archived change records a rejection of any remedy below.

Evidence at `5611091f`:

- F-A: the three shims carry the `resolve-file.sh` + `SELF_REAL` block; `flow_guard_root` is
  `scripts/lib/flow-guard.sh:80`. A sandbox copy of `check-task-build-green.sh` with the block
  replaced by `FLOW_GUARD_REPO_ROOT="$(flow_guard_root)"`, its `scripts/lib/flow-guard.sh` copied
  and `stats/` linked to the checkout's, run through a `skills/flow/scripts/` symlink from `/`,
  exits 1 naming the sandbox's dirty plan. The Go tests' sandboxes link `scripts/lib` to the real
  checkout's, which `flow_guard_root`'s `cd -P` follows to the real root — so they change with the
  shims (decision `shim-root-flow-guard-root`).
- F-B: `planunchanged.go`'s `planStatus` runs `git status --porcelain --untracked-files=all` over
  the plan tree; `flow tasks tick` (`stats/cmd/flow/tick.go`) flips `- [ ] ` to `- [x] ` on the
  task line and `  - [ ] ` to `  - [x] ` on its step lines, which turns a clean status into
  ` M …/tasks.md`.
- F-C: `tcfBodyFields` assigns `fields[current] = parts` on every field line. F1 of
  `check-plan-shape` flags any repeated `Files`, `Tests`, `Commit`, `Baseline` or
  `Allowed-collateral` line — the only fields `tcfBodyFields`' callers read — and it finds none over
  the 205 tracked `tasks.md`, so first-wins parses every tracked plan as last-wins did. <!-- measured: for f in $(git ls-files '*tasks.md'); do scripts/check-plan-shape.sh "$f"; done 2>&1 | grep -c 'declares a second' @ 5611091f → 0 -->
- F-D: `findByRole("region")` resolves while `Panel` shows `Loading…`; the self-review pass
  reproduced the failure with a 500 ms delayed mock and the fix with `findAllByRole`.
  <!-- measured: the self-review pass's delayed-mock spike @ 5611091f, machine-local, not re-run here -->
- F-E: `writepaneldiff.go` runs `git diff <range>` with no `-D`; no consumer of a panel diff
  (`renderslotprompt.go`, `dispatchparagraphs.go`, `check-panel-diff-size.sh`) reads a deleted
  file's body.

Repair: none of the five defects persisted a wrong record, row or file — F-B's false fires were
cleared in-run, F-C's misread was a guard's in-memory parse, F-D and F-E cost time and tokens.

## Live check

None — the change edits guard shims, two guards' Go source, view tests and reviewer prose; no
daemon, store or running service is touched.

## Decisions

### Shims derive their root with flow_guard_root

**ID:** shim-root-flow-guard-root
**Status:** active
**Chosen:** each of the three shims replaces its `resolve-file.sh` block with
`FLOW_GUARD_REPO_ROOT="$(flow_guard_root)"`, as `plan-class.sh` does; the Go tests' symlinked
sandboxes copy `lib/flow-guard.sh` into the sandbox and link `stats/` to the checkout's, and the
two "no lib/resolve-file.sh" pins are removed.
**Considered:** a new shared helper — `flow_guard_root` already is one; keeping the sandboxes'
`lib` symlink — `cd -P` would resolve them to the real checkout and scan its plans.

### Plan-unchanged snapshots content and tolerates tick flips

**ID:** plan-unchanged-content-ticks
**Status:** active
**Chosen:** the snapshot is, for each path `git ls-files -z --cached --others --exclude-standard`
lists under the plan tree (sorted), a `== <path>` line then the file's content, or
`== <path> (missing)` for a listed path absent on disk. Verify passes when the dumps are equal, or
have the same line count and every differing line pair is the snapshot line with its first
`- [ ] ` replaced by `- [x] `, where the snapshot line, leading spaces trimmed, opens with
`- [ ] `; it prints `PLAN-UNCHANGED-OK: … matches its pre-dispatch snapshot (checkbox ticks only)`
then. Anything else is exit 1 with the diff, as today.
**Considered:** keep the status snapshot and ignore ` M tasks.md` — would also pass a reviewer's
prose edit to `tasks.md`; forbid ticking during a flight — serialises the parent's tick behind
every reviewer, the slowdown the bundled dispatch exists to avoid; tolerate unticks too — no
pipeline step unticks while a reviewer flies.

### A repeated task field keeps its first occurrence

**ID:** task-field-first-wins
**Status:** active
**Chosen:** `tcfBodyFields` records a field only when it is not yet recorded; F1's message reads
"the parser keeps the first occurrence and ignores this one".
**Considered:** last-wins plus a guard refusing prose that opens with a field — F1 already flags
the duplicate, and a later prose line is never the declaration.

### Await the loaded panel content

**ID:** await-loaded-content
**Status:** active
**Chosen:** at each of the five sites the first read of loaded content becomes its `findBy*` /
`findAllBy*` form and is awaited; later sync reads in the same test stay sync. Proven by a
temporary 500 ms delay in each file's `fetchStatsView` mock factory, which fails every old site and
passes every new one, then reverted.
<!-- predicted: Task 4 Step 1's delayed vitest run fails the five old sites -->
**Considered:** `waitFor` around the block — a broader retry than the one query that races; a
longer `findBy` timeout — widens a tolerance (`rules/fix-determinism-at-the-source.mdc`).

### Panel diffs omit deleted bodies

**ID:** panel-diff-irreversible-delete
**Status:** active
**Chosen:** `write-panel-diff` passes `--irreversible-delete` to the diff call of every kind (not
the `--name-status` call); the parity test's hand recipe takes the same flag. The shim header and
the ENTRY CONTEXT paragraph (`skills/flow/review-panel.md`) say a deleted file shows as its header
alone and its old body is `git show <sha>:<path>`, the sha the section's diff range starts from —
for `slot-delta` with a held sha, the held sha; otherwise the sha its section's `# worktree:`
header names.
**Considered:** `final` only — every kind reads the same way and pays the same bytes;
`--diff-filter=d` — drops the deletion from the diff entirely, so a reviewer cannot see it.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** flow-medium — two guard verdicts move (plan-unchanged's pass condition, the field
parser every plan and commit guard reads), besides shims, tests and prose.
**Considered:** flow-low — fits prose-only or mechanical changes, not a changed guard verdict.

### Fixer groups

**ID:** fixer-groups
**Status:** active
**Chosen:** one fixer dispatch, g1, Tasks 1, 2, 3, 4, 5 in order — five small file-disjoint tasks
in one repository.
**Considered:** one dispatch per task — five cold starts for five small tasks; two groups (Go
guards, then web and prose) — fixer groups run in series, so a second cold start buys nothing.

## Open questions
