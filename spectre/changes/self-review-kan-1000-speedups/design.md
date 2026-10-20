## Context

Six non-`big` findings of the KAN-1000 self-review pass, planned as one change per
`skills/flow-self-review/SKILL.md` step 3. Blast radius: the flow skill docs, the verify-report
guard (Go plus shim) and the agent-baseline hook with its harness — 11 files.
<!-- measured: git diff --stat 09bec571..scratch in a scratch clone with all six task diffs applied → 11 files changed -->

Evidence, from the KAN-1000 worktree's `.superpowers/sdd/verify-report-visual-verify*.md`:
`verify-report-visual-verify-fix-1-gymie.md` (a scoped round) ran 88 mapped specs in 14.1 minutes
and reported `10 not done | 11 not done | motion not done`; `-fix-2-full` reported `10 not done`;
<!-- measured: grep -n '^- verify:\|^- steps:' verify-report-visual-verify*.md in /Users/tweety53/Projects/gymie-worktrees/kan-1000-fe-group-sessions-through-the-same-doors/.superpowers/sdd @ 2026-10-20 → fix-1: 88 mapped specs (14.1m), 10 not done; fix-2-full: 10 not done; fix-2/3/4 keyed -full -->
Rounds 2, 3 and 4 were all `-full` runs.
<!-- measured: ls verify-report-visual-verify*.md in the same directory @ 2026-10-20 → -fix-2-full, -fix-3-full, -fix-4-full -->
The fork prompt is in this session's transcript
(`You are the fork. Run the operator's requested change to the flow skill … as a self-review fix loop`).

A stopped earlier attempt left uncommitted edits in seven of these files. They were read as a
draft; the task diffs keep their intent and supersede them (decision `draft-superseded`).

All six task diffs were applied in order to a scratch clone at `09bec571` and checked there:
`go test ./internal/guard/` and `scripts/test-enforce-agent-baseline-hook.sh` pass, each new case
fails with its production change reverted, and `check-references.sh`,
`check-installed-citations.sh`, `check-markdown-integrity.py`, `check-dispatch-paragraphs.sh`,
`check-python-suppressions.sh`, `gofmt -l`, `go vet` and — with this change's
`verbatim-moves.txt` — `check-verbatim-moves.sh` are clean.

## Decisions

### Step 10 may block only on step 8 or on compose's own exit

**ID:** step10-block-admissible
**Status:** active
**Chosen:** `check-verify-report` refuses `10 blocked — …` unless step 8 is also `blocked — …` or the cause cites `compose-mockup-frames exit <n>` (non-zero) — step 10 measures the captures step 8 wrote, so any other block is the verifier's own choice (KAN-1000).
**Considered:** the draft's step-8-only rule — a broken sidecar map is a block step 10 itself prescribes (compose exit 1), and refusing it would send the `-rest` remainder into the same exit forever; refusing every `10 blocked` — a blocked step 8 leaves nothing to compose, so step 10 could never be honestly reported.

### `10 not done` is already refused

**ID:** not-done-already-enforced
**Status:** active
**Chosen:** no guard change for `not done` — the default branch already refuses it, pinned by the KAN-870 case `steps 9-11 not done`; the KAN-1000 reports carrying `10 not done` were planned from anyway, so that half of the finding is fixed by the no-plan-from-incomplete sentence (decision `no-plan-from-incomplete`).
**Considered:** a second test for `10 not done` alone — duplicates a passing case.

### An incomplete report is never planned from

**ID:** no-plan-from-incomplete
**Status:** active
**Chosen:** one clause in the Exit 1 paragraph of **The verifier dispatch** (`skills/flow/visual-verify.md`), which already owns the remainder re-dispatch: no fix task is planned from an incomplete report, its departures included, until its remainder completes it.
**Considered:** a sentence in `skills/flow/verify-fix-loop.md` step 1 — the loop only runs on a final report, and the incompleteness rule lives with the guard's verdict.

### Under a round base `verify` never runs the whole suite

**ID:** scoped-verify-never-whole
**Status:** active
**Chosen:** stated once in verifier step 7, which owns how `specs` feeds `verify`: under a round base a non-zero `specs` exit or a `specs` with no `<merge-base>` runs no `verify` spec, and the whole suite runs once, in the loop's closing unscoped run; `skills/flow/verify-fix-loop.md` adds that the round a `-full` run's defect opens re-runs scoped, its round base that run's HEAD.
**Considered:** the draft's sentence in verify-fix-loop step 4 — restates step 7 in a file the verifier never reads; leaving the `-full` paragraph as it was — rounds 2–4 of KAN-1000 each ran unscoped.

### gymie's every-spec fallback is project configuration

**ID:** gymie-specs-fallback-out-of-scope
**Status:** active
**Chosen:** no change to gymie's `scripts/specs-for-diff.mjs`, whose documented fallback selects every spec for a touched UI file declaring no spec-reached tag — the likely source of the scoped round's 88 mapped specs; it is that project's declared `specs` behaviour, outside this repository.
**Considered:** a flow rule capping how many specs a round may map — a scoped list that is long because the fix touched a shared file is still the right list.

### The implementer's measurement is prose in the loop's implement step

**ID:** implementer-measures-prose
**Status:** active
**Chosen:** one bold-led paragraph in step 2 of **The loop** (`skills/flow/verify-fix-loop.md`), conditional on the round's defects including a mockup departure, naming what the dispatch carries (verifier file path, capture spec and sidecar, mockups root, `mockup frame`, `capture`) and what the implementer runs and reports.
**Considered:** a new verbatim dispatch paragraph registered in `check-dispatch-paragraphs.sh` — one dispatch site, no second caller; `skills/flow/sdd-dispatch.md` — reaches every implementer, not only loop tasks.

### The base check extends the existing base-lag bullet

**ID:** base-check-in-auto-resolution
**Status:** active
**Chosen:** extend the `Failures that are already on the base` bullet of `skills/flow-contracts/operator-prompts-auto-resolution.md`: the parent runs `git fetch origin` and `git rev-list --count origin/<base>..origin/<the branch it tracks>` at the start of whichever verify stage runs first, and a count above 0 syncs the base; pushing it stays confirmed.
**Considered:** a new step in `skills/flow/verify-and-handoff.md` — would restate the bullet's sync rule in a second place.

### The 1dp allowance sits in Blocking

**ID:** allowance-in-blocking
**Status:** active
**Chosen:** a paragraph directly after the Blocking list of **Steps 3–13** (`skills/flow/visual-verify.md`) stating the allowance and its conditions; verifier step 10 reports every departure within 1dp with its value and leaves the call to the parent. The band and seam pairings' 2px thresholds are unchanged.
**Considered:** the draft's separate paragraph above Blocking — the operator named Blocking; folding the allowance into the band and seam thresholds — those are measurement-noise limits the verifier applies without a fix round, the opposite of an allowance granted only after one.

### The fork rule lives beside the closed dispatch list and is enforced in the baseline hook

**ID:** fork-denied-in-flow-session
**Status:** active
**Chosen:** a bold-led paragraph after **The self-check** under **Dispatch sites — the parent's closed list** (`skills/flow/implement.md`), cited from the self-review's "plans, changes and reviews nothing itself"; `hooks/enforce-agent-baseline.py`, which already sees every Agent call and holds the fork exemption, denies a fork when the session transcript records a `/flow`, `/flow-fast`, `/flow-plan` or `/flow-self-review` command or the prompt names `skills/flow`.
**Considered:** `hooks/enforce-loop-dispatch.py` — it acts only on loop-keyed prompts, and the KAN-1000 fork carried no `Key:` line; a flowd store lookup as `hooks/flow-active-change.py` does — a network call that misses marks younger than the harvest cycle; the prompt check alone — a terse fork ("run the loop above") evades it; matching every `/flow…` command — `/flow-status` and `/flow-settings` run no phase.

### verbatim-moves.txt is committed with the plan

**ID:** verbatim-moves-with-plan
**Status:** active
**Chosen:** the plan's diffs are exact, so every new or reworded run-loaded sentence is known now and listed in this change's `verbatim-moves.txt`, committed with the plan; the fixer never writes `spectre/changes/`.
**Considered:** the fixer appending each sentence in its task commit — writes `spectre/changes/`, which step 3 item 1 forbids the fixer.

### The draft is superseded by the task diffs

**ID:** draft-superseded
**Status:** active
**Chosen:** before Task 1 the fixer saves the stopped attempt's uncommitted edits as a patch outside the worktree and restores those seven files to HEAD; the task diffs carry every part of the draft this plan keeps.
**Considered:** committing the draft as found — its step-8-only guard rule loops on compose's own exit, its allowance sat outside Blocking, and its step-4 sentence restated verifier step 7; applying the diffs over the draft — they do not apply to it.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — Task 6 changes a PreToolUse hook every session's Agent calls pass through, so a wrong match denies dispatches everywhere, and Task 1 changes a guard verdict the verifier and the parent both act on.
**Considered:** `flow-low` — every diff is small, pre-verified and test-pinned, but a false deny in the hook reaches every session.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, one fixer — key `self-review-kan-1000-speedups-fix`, Tasks 1–6 in order: the diffs share files (`skills/flow/visual-verify.md` in Tasks 1 and 5, `skills/flow/verify-fix-loop.md` in Tasks 2 and 3, `skills/flow/visual-verify-verifier.md` in Tasks 2 and 5) and apply only in that order.
**Considered:** one fixer per task — six dispatches for pre-verified diffs that must run in series anyway.

## Open questions

## Live check

Task 6 changes the installed `~/.claude/hooks/enforce-agent-baseline.py` once the landing reaches
the checkout it links to. After landing: in a Claude Code session that has run `/flow`, an Agent
call with `subagent_type: "fork"` is denied with the `KAN-1000` reason; in a fresh session that ran
no `/flow*` command, a fork whose prompt names no `skills/flow` path goes through. Not working looks
like the fork in the `/flow` session being admitted, or a non-fork flow dispatch carrying the
baseline pointer being denied.
