## Context

- Nine non-`big` findings from the kan-989 self-review pass, planned as one change per item 1 of
  **3. Fix every finding that is not `big`** (`skills/flow-self-review/SKILL.md`). Each is carried
  by exactly one task; none is dropped and none was disproved. Fixes ride the change's own branch
  `spectre/kan-989-stats-ui-dashboard-review` from `2c0befa3`. Checked against this worktree at
  `2c0befa3`, the pre-reshape history in its reflog, and read-only `flow record` reads of the
  store:
  - F1 (Task 1). **The exit 2 is traced**, offline — a detached worktree at `d46966cd` (the fix
    round's ticked tip, 07:21:02Z), a `flow` shim on `PATH` serving the store's reads as they
    stood at 07:21:05Z (dispatch rows seq ≤ 20, pass rows rounds 0–1), telemetry off, nothing
    written: `check-unfinished-work.sh <wt> kan-989-stats-ui-dashboard-review <wt>` exits 2 with
    `ticked appended task(s) (Task 11, 12) and no <worktree>=<merge-base> range was passed —
    cannot judge their commits` (`unfinishedwork.go`'s signal-four commit half, the line the pass
    named). Passing `<wt>=04a3a161` prints `CLEAR:` at exit 0: dispatch rows 18–20 cover tasks 11
    and 12 and both carry `Task-Id:` commits. So the handoff omitted the ranges
    **Every run writing `IN_PROGRESS`** (`skills/flow/verify-and-handoff.md`) already prescribes,
    and even with them the gate is blind to the two gaps the run left: the last `verify` row
    (seq 16) started 01:20 and the last `visual-verify` (seq 17) 01:44, while `0a042b64` (09:53,
    `stats/web/`) and `b433ccb7` (10:15) landed after both. Blast radius 5: the guard, its test,
    its shim header, `verify-and-handoff.md`, `finish-contract-run1.md`.
  - F2 (Task 2). `check-plan-shape.sh` exits 1 on `tasks.md` at `a455eb3b`, `d25bd8fb` and
    `d46966cd` — the pre-reshape twins of `ee543138`, `48882a7b`, `04c60d88` — naming tasks 11 and
    12's whole-suite verify steps and a shared `seed.go` with no `**After:**`; it exits 0 at the
    plan-gate commit `f37972d7`. **Writing plans** runs the guard; `document-fix.md`'s append does
    not. Blast radius 1.
  - F3 (Task 3). Dispatch 20 is role `reviewer`, key `panel-fix1-primary`;
    `records.PanelDispatchRound` (`stats/internal/records/render.go`) needs `panel-<digits>-…`, so
    signal five never saw the round. Blast radius 2.
  - F4 (Task 4). `stats/internal/guard/helpers_test.go` and `stats/cmd/flow/journal_test.go`'s
    `TestMain` set no `FLOW_GUARD_TELEMETRY`, so every guard a test execs posts to
    `recordsAddr`'s default `http://127.0.0.1:4173`; `check_task_commit_fields_test.go`,
    `mutate_and_verify_test.go` and `runreproducer_test.go` each build a child env from scratch,
    which a `TestMain` setting cannot reach. Blast radius 5. Stored rows are not deleted
    (retain-stored-data).
  - F5 (Task 5). `stats/web/playwright.config.ts` `expect.toHaveScreenshot.maxDiffPixelRatio:
    0.02` — up to 26,214 pixels of the pinned 1280×1024 frame — set for an antialiasing flake;
    Playwright's own default is zero differing pixels. 18 baselines under
    `stats/web/tests/visual/*-snapshots/`. Blast radius up to 19.
  - F6 (Task 6). Dispatch rows 18 (`task-11-12`) and 19 (`task-12`) are role `implementer`;
    `uwTaskKey` (`unfinishedwork.go`) reads `task-11-12` as task 11 alone. The documented keys are
    `task-<n>-implementer` (`skills/flow/implement.md`) and `task-<n+n>-implementer-fix-<k>`
    (`skills/flow/gated-review-fix.md`); the self-review and pipeline-fix loops' fixers are role
    `implementer` too (`stats/internal/guard/loop.go`), under `self-review-<name>-fix…` and
    `pipeline-fix-<k>-…` keys. Blast radius 2, shared with F3.
  - F7 (Task 7). Row 21 `verify` was closed `blocked` only after row 25 `verify-fix-1` opened;
    **The loop**'s step 1 says nothing about the failing row. Blast radius 1.
  - F8 (Task 8). `visual-verify-verifier.md`'s report template has only `exit <n>` for a command's
    result, so a denied command got an invented one. Blast radius 1.
  - S1 (Task 9). **Inline verify** runs every `## lint` and `## test` command in one serial list;
    `flow suite list` records stats-go at 1m41s–2m31s, guard-tests at 54s–1m19s, stats-spa at
    3–8s, and this repository's `## lint` holds 28 commands. Blast radius 1 — no `.flow/project.md`
    edit (decision `verify-two-groups`).
- **Prose wording is fixed here.** Tasks 1, 2, 7, 8 and 9 each apply the diff their step gives:
  `check-verbatim-moves.sh` admits a new run-loaded sentence only by its exact text in this
  change's `verbatim-moves.txt`. Spiked at plan time in a throwaway scratch clone at `2c0befa3`,
  never in this worktree: with all five diffs applied, `check-verbatim-moves.sh d0f4c430`,
  `check-references.sh`, `check-dispatch-paragraphs.sh`, `check-ask-silence.sh`,
  `check-installed-citations.sh` and `check-markdown-integrity.py` exit 0.

## Decisions

### The guard derives a missing range

**ID:** uw-derive-missing-range
**Status:** active
**Chosen:** when no `<worktree>=<merge-base>` argument is passed, `check-unfinished-work` derives
one for its own worktree — `resolveBaseBranch` in-process, then `git merge-base HEAD
origin/<base>` — the value every caller is told to pass; a derivation that fails leaves the
ranges empty, so signal four's commit half keeps its exit 2 and the two new signals below are not
run.
**Considered:** keeping the exit 2 and rewording the handoff — the handoff already states the
ranges, and the run still dropped them; a derivation failure that is itself exit 2 — every
existing no-range fixture, which has no `origin`, would turn exit 2; deriving every peer
worktree's range too — the guard is called once per worktree and knows only its own.

### The guard folds in the visual verdict

**ID:** uw-fold-visual-verdict
**Status:** active
**Chosen:** with a range for its own worktree, the guard runs `checkVisualVerifyDispatched`
in-process on `<worktree> <name> <merge-base> <canonical>`; its exit 1 adds its verdict line to
the breakdown, its exit 2 is the guard's exit 2. Integrate's separate visual call and its hard
`STOP: visual-verify` are unchanged.
**Considered:** a new prose step at the handoff — the run already skipped two prescribed guard
calls; replacing integrate's separate call — would let `--accept-outstanding` pass a missing
visual verification, which **Run 1** forbids.

### A code commit needs a later completed verify row

**ID:** uw-verify-row-after-code
**Status:** active
**Chosen:** in the worktree holding the plan, an unarchived change whose newest commit touching a
path outside `<spec-root>/changes/` — over every range, by author time — is later than the
`startedAt` of its newest `completed` `verifier` row keyed `verify` or `verify-…` is unfinished;
author time, since a rebase rewrites committer time and keeps it. Archived changes skip the
signal: their later commits are the self-review loop's, gated by its own reviews.
**Considered:** committer time — a rebase before integrate's gate would date every commit after
its verify row; any verify row regardless of outcome — a `blocked` row is a failed run, not a
verified tree; a per-worktree key match — the `-<repository name>` suffix exists only on
multi-worktree runs, and one row per change is what kan-989 lacked.

### Fix-run appends run the plan guards before committing

**ID:** fix-append-runs-plan-guards
**Status:** active
**Chosen:** one paragraph in `document-fix.md` running `check-plan-shape.sh` and
`check-plan-provenance.sh` over the appended plan, exits handled as **D. Basic Workflow #3**
handles them.
**Considered:** adding the guards to `check-planning-commit-location` — a commit-time guard
for one command's own step, larger than the defect.

### Dispatch begin refuses off-convention keys

**ID:** dispatch-key-shapes
**Status:** active
**Chosen:** `flow record dispatch begin` exits 2 before any store contact for a `reviewer` row
whose key starts `panel-` and `records.PanelDispatchRound` rejects, and for an `implementer` row
whose key matches neither `^task-[0-9]+(\+[0-9]+)*-implementer(-fix-[0-9]+)?$` nor starts
`self-review-` or `pipeline-fix-`; each message names the shape it wants.
**Considered:** refusing at `dispatch end` — the row is already written; normalising a bad key —
invents a key the caller never sent; a guard over stored rows — finds the defect after the run
that made it.

### Test packages turn guard telemetry off

**ID:** test-telemetry-off
**Status:** active
**Chosen:** `os.Setenv("FLOW_GUARD_TELEMETRY", "off")` first in both `TestMain`s, the variable
added to the three hand-built child envs, and one test asserting a child started the package's
way sees `off`. Stored rows stay.
**Considered:** pointing `FLOW_RECORDS_ADDR` at a dead port — still costs each guard the
telemetry timeout; deleting the 6,065 rows — retain-stored-data.

### Playwright renders deterministically at zero tolerance

**ID:** playwright-deterministic-zero-tolerance
**Status:** active
**Chosen:** Chromium launch flags in `use.launchOptions.args` that remove the run-to-run
rasterisation variance — `--font-render-hinting=none`, `--disable-font-subpixel-positioning`,
`--disable-lcd-text`, `--disable-gpu`, `--force-color-profile=srgb` — the `expect` tolerance
block deleted so Playwright's zero-pixel default applies, every baseline re-captured against the
seeded UI-test stack on 127.0.0.1:4174, proved by three clean runs with the flags and failing runs
without them. Per **A determinism fix whose flake does not reproduce is not committed**
(`skills/flow/verify-fix-loop.md`), flags the reproduction shows no need for are dropped, the
tolerance removal stands.
**Considered:** a smaller ratio — still a widened tolerance (`rules/fix-determinism-at-the-source.mdc`);
a bundled web font — a new asset when the baselines are already per-platform (`-darwin.png`).

### Inline verify runs two concurrent groups

**ID:** verify-two-groups
**Status:** active
**Chosen:** the lint list (with `check-spec-reach.sh`) and the test list run at once, each in
order, each detached and polled as the verifier's step 7 runs a long `verify`; no
`.flow/project.md` grouping.
**Considered:** a `.flow/project.md` grouping the pass proposed — configuration only this
repository would set, while lint is read-only in every project; one group per `## test` command
— two test commands of one project may share a database or ports; background-and-poll alone —
fixes the timeout and keeps the ten minutes.

### The failing verify row closes before the loop's append

**ID:** loop-closes-failing-row
**Status:** active
**Chosen:** one sentence at the end of **The loop**'s step 1.
**Considered:** making `dispatch begin` refuse while a same-stage row is open — a store change for
a one-step ordering.

### A denied command is blocked, not an exit code

**ID:** denied-command-blocked
**Status:** active
**Chosen:** one paragraph ahead of the verifier's steps: `blocked — environment: <denial text>` in
place of `exit <n>`.
**Considered:** editing every `exit <n>` line of the report template — nine lines for one rule.

### First review effort and fixer grouping

**ID:** first-review-effort
**Status:** active
**Chosen:** first review on `flow-medium` (effort `medium`) — Task 1 adds two signals and a
derivation to a gate every integrate runs. Three fixer dispatches, in series: g1 (key
`self-review-kan-989-stats-ui-dashboard-review-fix-g1`) Tasks 4, 3, 6 — first, so every later
test run is already telemetry-free; g2 (`…-fix-g2`) Tasks 1, 9, 2, 7, 8 — the guard and every
prose diff; g3 (`…-fix-g3`) Task 5 — the Playwright work against the UI-test stack.
**Considered:** `flow-low` — Task 1's signals are new gate logic a medium read should check; one
fixer — Task 5's stack runs and baseline re-captures would crowd the guard work's context; one per
task — nine dispatches for five two-line diffs.

## Open questions

## Live check

None for this change's run: Task 1's signals are pinned by fixture cases against real git
repositories and a faked store, Tasks 3, 4 and 6 by unit tests, Task 5 by its own proof runs
against the UI-test stack, and the rest is prose.
