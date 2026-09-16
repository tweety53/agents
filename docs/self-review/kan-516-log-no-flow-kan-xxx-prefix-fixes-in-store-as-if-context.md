# Self-review context bundle for kan-516-log-no-flow-kan-xxx-prefix-fixes-in-store-as-if

found: 4 of 4 sources; skipped: 0 of 4 sources; refused: 0 of 4 sources

## .superpowers/sdd/ledgers/kan-516-log-no-flow-kan-xxx-prefix-fixes-in-store-as-if.md

# SDD ledger — kan-516-log-no-flow-kan-xxx-prefix-fixes-in-store-as-if

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — reviewer

- Task: 1
- Role: reviewer
- Key: task-1-reviewer
- Model: sonnet effort=default
- Commit: no commit
- Outcome: clean
- Started: 2026-09-16T14:39:17Z
- Tokens: input 20, output 3964, cache read 293081, cache creation 54338

## Dispatch 2 — reviewer

- Task: 2
- Role: reviewer
- Key: task-2-reviewer
- Model: sonnet effort=default
- Commit: no commit
- Outcome: clean
- Started: 2026-09-16T14:46:00Z
- Tokens: input 20, output 2327, cache read 297113, cache creation 53168

## Dispatch 3 — reviewer

- Task: 3
- Role: reviewer
- Key: task-3-reviewer
- Model: sonnet effort=default
- Commit: no commit
- Outcome: clean
- Started: 2026-09-16T14:53:10Z
- Tokens: input 30, output 3700, cache read 487277, cache creation 85480

## Dispatch 4 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: sonnet effort=medium
- Commit: no commit
- Outcome: completed
- Started: 2026-09-16T14:57:01Z
- Tokens: input 82, output 10977, cache read 3789475, cache creation 208430

## Dispatch 5 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-1-primary+principles
- Model: sonnet effort=medium
- Commit: no commit
- Diff base: 1375641
- Outcome: completed
- Started: 2026-09-16T15:08:36Z
- Tokens: input 54, output 6093, cache read 1188673, cache creation 85964

## Dispatch 6 — verifier

- Task: no task
- Role: verifier
- Key: verify
- Model: sonnet effort=medium
- Commit: no commit
- Outcome: completed
- Started: 2026-09-16T15:17:43Z
- Tokens: not measured


## .superpowers/sdd/reviews/kan-516-log-no-flow-kan-xxx-prefix-fixes-in-store-as-if-panel.md

# Review panel — kan-516-log-no-flow-kan-xxx-prefix-fixes-in-store-as-if

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Minor | skills/flow/SKILL.md:215 | misattributes token generation to implement.md section 3 (which only uses the token) instead of the router's own generate-once rule at skills/flow/SKILL.md:247-250 |   |
| F2 | primary | Minor | hooks/flow-active-change.py:26 | uses str \| None (PEP 604) vs. the repo's established typing.Optional[str] convention; would break at import on Python < 3.10 with no minimum-version contract declared |   |
| F3 | principles | Critical | hooks/flow-active-change.py:86 | new # noqa: BLE001 inline suppression, forbidden outright by the Lint Fix Priority hard invariant; no CONTRIBUTING.md pre-approves it |   |
| F4 | principles | Minor | setup.sh:399 | the per-hook registration-check-and-warn shape is now duplicated a second time inside install_hooks/install_hooks_zcode; defensible WET, worth a shared helper if a third hook is added |   |
| F5 | primary | Minor | skills/flow/SKILL.md:216 | F1's own fix says 'per Generate this run's session token once above' but that section sits below (line 247), not above |   |

findings-total: 5
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 deferred structural DRY observation, not a defect; a shared helper is only worth building if a third hook is added
finding-status: F5 fixed

reproducers-total: 5
finding-reproducer: F1 none — prose-citation nit, not independently testable
finding-reproducer: F2 none — cosmetic/convention nit
finding-reproducer: F3 .superpowers/sdd/reproducers/0-principles-1.sh
finding-reproducer: F4 none — structural observation, not a defect
finding-reproducer: F5 .superpowers/sdd/reproducers/1-primary-1.sh

## Pass log

### Round 0

- panel diff size 801, under cap
- docs-only: no (first non-doc path docs/research/kan-516/decision.json) — resolved roster runs: primary+principles
- roster: compact — sonnet/medium, free grouping (bundle roll 39)
fix-mutation: hooks/flow-active-change.py — none — F3's fix: removed the noqa comment and switched str | None -> Optional[str] for import symmetry — annotations and a comment string only, no executable behaviour changed; reproducer 0-principles-1.sh confirmed non-zero pre-fix, zero post-fix
fix-mutation: skills/flow/SKILL.md — none — F1's fix: corrected a prose citation in a Markdown skill file — no executable behaviour
fix-mutations-total: 2

### Round 1

- round 1: primary and principles both raised findings last round -> both re-run on delta; delta cap 470 under cap; docs-only: no (unchanged) — roster stays primary+principles
fix-mutation: skills/flow/SKILL.md — none — F5's fix: corrected a citation direction word (above->below) — no executable behaviour; reproducer 1-primary-1.sh confirmed failing pre-fix, passing post-fix, and the reverted-word mutation was re-verified to trip it
fix-mutations-total: 1

## /Users/tweety53/Projects/agents/.worktrees/_landing-kan-516-log-no-flow-kan-xxx-prefix-fixes-in-store-as-if/spectre/changes/archive/kan-516-log-no-flow-kan-xxx-prefix-fixes-in-store-as-if/tasks.md

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

# kan-516-log-no-flow-kan-xxx-prefix-fixes-in-store-as-if

KAN-516 · log unprefixed fixes in the store as `/flow <name>` fix runs · implementation plan.

**Goal:** a problem report typed as a plain message, in a session whose last `/flow <name>` run
left the change at `IN_PROGRESS`, runs as the fix run `/flow <name> <message>` would — planning
pass, implementation, panel, verify, handoff — under a fresh session token, so the store records
it as one more fix run. A `UserPromptSubmit` hook names the change on every plain prompt from the
transcript's last `/flow` mark; the `/flow` skill carries the rule. `docs/research/kan-516.md`
(sections 2–5) is canonical for every decision a task implements until `/flow` folds it into this
change's `design.md`.

Three tasks, one repository. No spec, no migration.

**Baseline, measured before any edit, on `main` at `ad45fa5`:** `test-setup.sh` carries 146
assertions.
<!-- measured: grep -c '^assert_' scripts/test-setup.sh @ main ad45fa5 -->

**Every task's verify step is the lint guards its own files need plus its own harness**, never
`scripts/run-guard-tests.sh` — that run is the last bundle's FULL SUITE paragraph and
`flow.verify`'s.

---

- [x] 1. The prompt hook and its harness

**Build:** green
**Files:** `hooks/flow-active-change.py`, `scripts/test-flow-active-change-hook.sh`
**Tests:** `test-flow-active-change-hook.sh` — `Case 1: one /flow mark — the context line names
its change`, `Case 2: two /flow marks for different changes — the last one wins`, `Case 3: only
/flow-plan and /flow-fast marks — no output, exit 0`, `Case 4: a prompt starting with / — no
output, exit 0`, `Case 5: malformed stdin — no output, exit 0`, `Case 6: no transcript_path — the
rollout under FLOW_ZCODE_ROLLOUTS_DIR named by session_id is read`
**Regression:** reverting this commit removes the hook, so no plain prompt ever carries the
change name and the rule in task 3 rests on the session's own memory alone.
**Baseline:** `test-flow-active-change-hook.sh` before=0 after=6
<!-- predicted: grep -c '^# Case [0-9]' scripts/test-flow-active-change-hook.sh after task 1 -->
**Commit:** `feat(hooks): inject the session's active flow change on every plain prompt`

  - [x] **Step 1: The hook.** `hooks/flow-active-change.py`, Python 3 standard library only, the
    module docstring stating the contract as `hooks/enforce-agent-baseline.py`'s does: reads the
    stdin JSON; returns 0 with no output when the payload is not an object, when `prompt` starts
    with `/`, or on any exception. Transcript: `transcript_path` when present, else
    `<root>/model-io-sess_<session_id>.jsonl` with `<root>` = `FLOW_ZCODE_ROLLOUTS_DIR` or
    `~/.zcode/cli/rollout` and `<session_id>` = the payload's `session_id` or the
    `CLAUDE_SESSION_ID` environment variable. Read the file as bytes decoded with
    `errors="replace"`, take the last match of

  ```python unverified:confirm the pattern matches a real Claude Code transcript line carrying a fix run's flow.document-fix mark — the mark's flag order is the one skills/flow/implement.md section 3 prints
  re.compile(r"flow stage begin -command '/flow' [^;|&\n]*?-session-token mf-[A-Za-z0-9_-]+ ([A-Za-z0-9._-]+)")
  ```

    and print `{"hookSpecificOutput": {"hookEventName": "UserPromptSubmit", "additionalContext":
    LINE}}` via `json.dumps`, `LINE` being `flow: this session's last /flow run marked change
    <name>. A plain message reporting a problem or asking for a change to it is a fix run of
    <name> — **A plain message at IN_PROGRESS** (skills/flow/SKILL.md).` No match, no file: exit
    0, no output.
  - [x] **Step 2: Harness.** `scripts/test-flow-active-change-hook.sh`, `set -euo pipefail`, a
    `mktemp -d` sandbox removed on exit, one `# Case N: …` comment per case exactly as
    `**Tests:**` names them, a `run_hook <payload-json>` helper that pipes the payload to
    `python3 "$REPO/hooks/flow-active-change.py"` capturing stdout and rc. Fixture transcripts
    are JSONL lines whose `"command"` string carries the mark, written with `printf` so the
    single quotes survive. Case 1 asserts rc 0 and stdout containing `marked change
    fixture-one`; Case 2 writes marks for `fixture-one` then `fixture-two` and asserts
    `fixture-two`; Case 3 writes `-command '/flow-plan'` and `-command '/flow-fast'` marks only and
    asserts empty stdout; Case 4 sends `"prompt": "/flow x"` over Case 1's transcript and asserts
    empty; Case 5 pipes `not json` and asserts rc 0 and empty; Case 6 omits `transcript_path`,
    sets `FLOW_ZCODE_ROLLOUTS_DIR` to the sandbox holding `model-io-sess_abc.jsonl` with a mark,
    sends `"session_id": "abc"` and asserts `marked change`. Print `PASS`/`FAIL` per case and exit
    non-zero on any failure, the shape `scripts/test-check-self-review-report.sh` uses.
  - [x] **Step 3: Verify.**

  ```bash verified:the harness is this task's own file; check-vocabulary.sh is a .flow/project.md ## lint row and scans scripts/ and hooks/
  scripts/test-flow-active-change-hook.sh
  scripts/check-vocabulary.sh
  ```

- [x] 2. Registration reporting, README, setup tests and the live check

**Build:** green
**Files:** `setup.sh`, `scripts/test-setup.sh`, `README.md`
**Tests:** `test-setup.sh` — the assertions `the flow-active-change hook is installed`, `an
unregistered flow-active-change hook is reported` and `the zcode flow-active-change hook is
linked`
**Regression:** reverting this commit makes `setup.sh` report a fresh install as fully
registered while the prompt hook is linked and never fires.
**Baseline:** `test-setup.sh` before=146 after=149
<!-- predicted: grep -c '^assert_' scripts/test-setup.sh after task 2 -->
**Commit:** `feat(setup): report and document the flow-active-change hook's registration`
**After:** Task 1

  - [x] **Step 1: `install_hooks`.** After the existing `enforce-agent-baseline` grep, a second
    `grep -q 'flow-active-change' "$settings"` under the same `require_grep_ok` discipline; return
    early only when both are registered. When the prompt hook is missing, print `⚠ The
    flow-active-change hook is installed but NOT registered, so a plain problem report is never
    named as a fix of the open change. Add this to "hooks" in $settings:` and the snippet

  ```json verified:the shape is the PreToolUse snippet install_hooks already prints, with the event and command swapped
      "UserPromptSubmit": [
        {
          "hooks": [
            { "type": "command", "command": "python3 \"$HOME/.claude/hooks/flow-active-change.py\"" }
          ]
        }
      ]
  ```

  - [x] **Step 2: `install_hooks_zcode`.** The same second grep against `$config` and the
    ZCode-shaped snippet: a `UserPromptSubmit` entry beside `PreToolUse` under `events`, command
    `python3 "$HOME/.zcode/hooks/flow-active-change.py"`.
  - [x] **Step 3: README.** In the tree listing, under `hooks/`, add
    `flow-active-change.py ← UserPromptSubmit hook: names the session's last /flow change on
    every plain prompt, so a problem report runs as its fix run`; in the ZCode **Step 2**
    snippet, add the `UserPromptSubmit` event beside `PreToolUse`.
  - [x] **Step 4: Setup tests.** In `make_fixture_repo`, beside the baseline fixture hook, write
    `$d/hooks/flow-active-change.py` with the same two-line stub. Beside the three existing hook
    assertions add the three `**Tests:**` names: `assert_exists` on
    `$home/.claude/hooks/flow-active-change.py`, `assert_contains "$RUN_LOG"
    'flow-active-change.py'` in the global group, `assert_symlink` on
    `$home/.zcode/hooks/flow-active-change.py` in the ZCode group.
  - [x] **Step 5: Live check.** Pick a real transcript:
    `grep -l "flow stage begin -command '/flow' " ~/.claude/projects/-Users-tweety53-Projects-agents/*.jsonl | head -1`
    — the grep confirms the file holds a `/flow` mark, so an empty result below is a failure and
    not an unmarked transcript. Pipe `{"prompt":"the button is misaligned","transcript_path":"<that
    file>"}` to `python3 hooks/flow-active-change.py`; success is one JSON line whose
    `additionalContext` names the change that transcript's last mark carried.
  - [x] **Step 6: Verify.**

  ```bash verified:test-setup.sh is this task's own harness; the three guards are .flow/project.md ## lint rows covering setup.sh's fixture and README.md
  scripts/test-setup.sh
  scripts/check-vocabulary.sh
  scripts/check-references.sh
  scripts/check-contract-budget.sh
  ```

- [x] 3. The router rule

**Build:** green
**Files:** `skills/flow/SKILL.md`, `skills/flow-contracts/pipeline.md`, `commands/flow.md`, `commands-claude/flow.md`, `scripts/check-contract-budget.sh`
**Tests:** **none** — a skill, a contract and two stubs carry prose; the guards in the verify
step are their checks.
**Regression:** reverting this commit leaves the hook naming a change on every prompt with no
rule that turns the message into a fix run, so the store stays as empty as today.
**Baseline:** `test-setup.sh` before=149 after=149 (unchanged — this task adds no test)
<!-- predicted: no harness change in task 3 -->
**Commit:** `feat(flow): run a plain problem report at IN_PROGRESS as a fix run of the session's change`
**After:** none

  - [x] **Step 1: Transition table.** In `skills/flow-contracts/pipeline.md`, after the
    `IN_PROGRESS, with an argument` row:
    ``| *(none — a plain message)* | `IN_PROGRESS`, in the session whose last `/flow` run marked the change | fix run; state unchanged — **A plain message at `IN_PROGRESS`** (`skills/flow/SKILL.md`) |``.
    In **Every invocation is re-entrant**, after the "with an argument" bullet: `**At
    IN_PROGRESS, with no invocation at all** — a plain message reporting a problem or asking for
    a change, typed in the session that ran the last /flow <name> — is that same fix run, the
    message its instructions.` The bare-invocation paragraph that starts integrate is untouched.
  - [x] **Step 2: The subsection.** In `skills/flow/SKILL.md`, under **Reading the state**,
    after the state bullets and before **Check guard presence**, add `### A plain message at
    IN_PROGRESS` carrying, in this order: (a) the trigger — a message with no `/flow` invocation
    that reports a problem or asks for a change to the change's code or artifacts, in a session
    whose most recent `/flow` run marked `<name>`; `<name>` comes from the `flow: this session's
    last /flow run marked change <name>` context line `hooks/flow-active-change.py` injects, or
    from this session's own context when no hook is registered; (b) the gate — `flow state get
    <name>` must answer `IN_PROGRESS`; any other answer, or an unreachable store, makes the
    message an ordinary turn, never a wrong-state handoff, since nothing was invoked; (c) the
    ambiguity prompt — a message that reads as either a question or a change request asks once,
    shape per **The shape** (`skills/flow-contracts/operator-prompts.md`): `**Run this as a fix
    of <name>?**` — **Yes** *(recommended)* / **No — ordinary turn**; silence takes Yes and the
    handoff carries the ⚠ line; (d) the run — announce `Using flow for change <name> — fix run
    from a plain message`, generate a fresh `mf-` token exactly as the paragraph below generates
    one, and continue at **3. Documenting a fix, before implementing it**
    (`skills/flow/implement.md`) with the verbatim message as the fix instructions — nothing
    else about the fix run changes; (e) what is not a trigger — a question, a discussion, an
    unrelated task, a session that never ran `/flow`, and a Jira key named in prose without the
    slash. Add `— or a plain message, per **A plain message at IN_PROGRESS** below` to the
    `IN_PROGRESS, an argument present` bullet.
  - [x] **Step 3: Stubs.** In both `commands/flow.md` and `commands-claude/flow.md`, after the
    sentence ending "the argument is fix instructions.": `A plain problem report typed with no
    /flow at all, in the session that ran the last /flow <name>, is the same fix run (**A plain
    message at IN_PROGRESS**, skills/flow/SKILL.md).`
  - [x] **Step 4: Budget.** `skills/flow/SKILL.md` sits at its `budgets()` row exactly (17812);
    set the row to the file's new size plus 25%. Check `commands/flow.md`,
    `commands-claude/flow.md` and `skills/flow-contracts/pipeline.md` against their rows after the
    edits and raise any the growth crosses.
  - [x] **Step 5: Verify.**

  ```bash verified:each guard is a .flow/project.md ## lint row and runs without arguments over skills/, commands/ and commands-claude/
  scripts/check-vocabulary.sh
  scripts/check-references.sh
  scripts/check-contract-budget.sh
  scripts/check-stage-mark-calls.sh
  scripts/check-installed-citations.sh
  ```

## git log --stat

commit 687cd8bff92322241ba2d91c4475f18bede766cb
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Wed Sep 16 18:27:04 2026 +0300

    feat(flow): treat a plain problem report at IN_PROGRESS as a fix run of the session's change

 README.md                               |  12 +++-
 commands-claude/flow.md                 |   4 +-
 commands/flow.md                        |   4 +-
 hooks/flow-active-change.py             |  88 +++++++++++++++++++++++++
 scripts/check-contract-budget.sh        |   2 +-
 scripts/test-flow-active-change-hook.sh | 110 ++++++++++++++++++++++++++++++++
 scripts/test-setup.sh                   |   9 +++
 setup.sh                                |  87 +++++++++++++++++++------
 skills/flow-contracts/pipeline.md       |   4 ++
 skills/flow/SKILL.md                    |  27 +++++++-
 10 files changed, 321 insertions(+), 26 deletions(-)
commit c446935468d2f68ddd8f7af76e57c72b9b3ae56e
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Wed Sep 16 18:27:04 2026 +0300

    chore(spectre): plan

 docs/research/kan-516.md                           | 213 ---------------------
 docs/research/kan-516/decision.json                |  46 -----
 .../design.md                                      |  77 ++++++++
 .../narrative.md                                   |  52 +++++
 .../proposal.md                                    |  20 ++
 .../tasks.md                                       |  38 ++--
 6 files changed, 168 insertions(+), 278 deletions(-)
commit 8388557b7d4f905b845230185b40adb900d89a5c
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Wed Sep 16 18:31:11 2026 +0300

    chore(spectre): archive kan-516-log-no-flow-kan-xxx-prefix-fixes-in-store-as-if

 .../kan-516-log-no-flow-kan-xxx-prefix-fixes-in-store-as-if/design.md     | 0
 .../kan-516-log-no-flow-kan-xxx-prefix-fixes-in-store-as-if/narrative.md  | 0
 .../kan-516-log-no-flow-kan-xxx-prefix-fixes-in-store-as-if/proposal.md   | 0
 .../kan-516-log-no-flow-kan-xxx-prefix-fixes-in-store-as-if/tasks.md      | 0
 4 files changed, 0 insertions(+), 0 deletions(-)


## design.md

## Context

The fix path already exists — `skills/flow/implement.md` section 3 runs whenever the router reads
`IN_PROGRESS` with an argument. The only gap is a router entry: the router is entered only by a
`/flow` invocation, and a plain message never reaches it. This is one change: a hook that supplies
the change name deterministically on every plain prompt, plus the router rule that classifies a
message and hands it to the existing fix path with a fresh session token.

## Decisions

### Full fix run, not a thinner inline path

**ID:** unprefixed-fix-full-ceremony
**Status:** active
**Chosen:** An unprefixed problem report runs exactly as `/flow <name> <message>` would — the
document-fix planning pass, implementation, review panel, verification and handoff — so the store
sees an ordinary fix run.
**Considered:** An inline fix wrapped in marks only (tokens and a fix iteration would land, but the
planning pass, panel and verify records would not — a second, thinner kind of fix run the store
could not tell apart from the real one) — rejected. Asking each time whether to run it as a fix —
rejected: a click on every message defeats the convenience the ticket is about.

### Trigger is change requests only

**ID:** unprefixed-fix-trigger-change-requests
**Status:** active
**Chosen:** A message that reports a problem or asks for a change to the change's code or artifacts
is a fix run; a question, a discussion, or an unrelated task stays an ordinary turn.
**Considered:** Every non-slash message — rejected: would turn "why did you do that?" into a
planning pass.

### Ambiguous messages ask once

**ID:** unprefixed-fix-ambiguity-ask-once
**Status:** active
**Chosen:** A message that reads either way is one `AskUserQuestion` — "Run this as a fix of
`<name>`?", Yes (recommended) / No — ordinary turn; silence takes Yes and the fix run's handoff
carries a `⚠` line.
**Considered:** Treating doubt as an ordinary turn — rejected: the silent data loss the ticket
reports. Treating doubt as a fix run unconditionally — rejected: a planning pass over a plain
question.

### Cold sessions and bare Jira mentions stay out of scope

**ID:** unprefixed-fix-cold-sessions-out-of-scope
**Status:** active
**Chosen:** Only a session that already ran `/flow <name>` is in scope. A session with no prior run
in the change's worktree, and a Jira key named in prose without the slash, both stay ordinary turns.
**Considered:** Extending the rule to either case — rejected: each is a discovery mechanism of its
own, and neither is the case the ticket reports.

### Discovery via a prompt hook plus skill text

**ID:** unprefixed-fix-discovery-hook-plus-skill
**Status:** active
**Chosen:** The rule itself lives in the `/flow` skill; a `UserPromptSubmit` hook
(`hooks/flow-active-change.py`) supplies the change name deterministically on every plain prompt by
reading the transcript's last `/flow` stage-begin mark, so the rule survives context compaction and
long sessions.
**Considered:** Skill text alone — rejected: works on every harness but is lost once the skill's
text falls out of context. A hook alone — rejected: the hook can name the change but cannot classify
a message, and the rule must be readable by a session with no hook registered.

### The fix run's state gate is the existing one

**ID:** unprefixed-fix-state-gate-existing
**Status:** active
**Chosen:** The fix run starts only when `flow state get <name>` answers `IN_PROGRESS`. Any other
state — `STARTED`, `FINISHED`, no state, or a store that cannot be reached — makes the message an
ordinary turn, never a wrong-state handoff, since the operator invoked nothing.
**Considered:** Treating a mismatch as a wrong-state handoff like an actual `/flow` invocation —
rejected: there is no command to report a wrong state for when nothing was invoked.

## Open questions

None — this round was fully seeded by `docs/research/kan-516.md`, whose own convergence checks
(`/flow-plan`'s per-topic closing check) already resolved every question this design raises.

## Session narrative

# kan-516-log-no-flow-kan-xxx-prefix-fixes-in-store-as-if — session narrative

## 2026-09-16 — creating run

Fully-seeded from `docs/research/kan-516.md` — the checklist and convergence confirm were both
skipped per that exception, straight into artifact creation and the copied plan/decision. Inline
execution (class `small`), compact panel (`primary+principles`, one bundled dispatch).

Three tasks landed clean, each gated for review (all three commits crossed the 40-line/task
threshold) and closed by a dispatched reviewer with no findings.

The panel's own round-0 dispatch caught a real defect this session introduced: a new `# noqa:
BLE001` inline suppression in `hooks/flow-active-change.py`, forbidden outright by this project's
Lint Fix Priority rule. Fixed by dropping the comment (the file's own `## lint` list runs no
Python linter over `hooks/*.py`, so nothing was actively suppressed — the rule still forbids the
marker on principle). Also fixed two Minor findings inline: a wrong `str | None` vs the repo's
`Optional[str]` convention, and a stale prose citation ("above" instead of "below") in
`skills/flow/SKILL.md`. Deferred one Minor (a DRY observation about `setup.sh`'s two near-identical
hook-registration blocks — not worth a shared helper for two occurrences).

Two of the three reviewer-written reproducer scripts (`0-principles-1.sh` for the noqa finding,
`1-primary-1.sh` for the citation-direction finding) had inverted or dead exit-code logic — one
checked immutable line order instead of the actual cited word, the other never followed
`run-reproducer.sh`'s own "non-zero = defect present" contract. Both were rewritten and proved to
bite in both directions (defect present → fails; fixed → passes) before being trusted for
verification. Worth flagging to future reviewer prompts: a reproducer needs the same TDD-style
proof its own finding does.

The round-1 delta re-run (re-checking the fixes) caught one more real defect — the fix for the
citation finding fixed the substance but got its own directional word wrong ("above" when the
section is textually below) — fixed the same way. Round 1 raised only that one Minor, so no
further re-run was needed per the panel contract.

No UI paths touched (`stats/web/src/**`), so `flow.visual-verify` skipped cleanly. The only
runnable app this project declares is the protected flow dev stack (`127.0.0.1:4173`), so nothing
was started for the handoff.

## 2026-09-16 — integrate run

Preflight verdict `RUN1` (branch not yet merged). The unfinished-work gate and the
visual-verify-dispatched guard both came back clear with no operator prompt needed.
`check-base-moved.sh` reported `MOVED` — one commit on `origin/main` since the recorded merge
base — but that commit was the very `docs(research): kan-516 research note, plan and decision`
commit this branch already carries as its own first commit (landed to `main` by an earlier,
unrelated `/flow-plan` session while this change was in flight). `git rebase origin/main` was a
clean no-op for that reason; re-running `check-base-moved.sh` against the rebased merge base
confirmed `CLEAR`. The rebase needed the deliberately-unstaged `docs/research/kan-516*` deletions
out of the way first — stashed by unique tag, reapplied by SHA, and dropped, per this project's
stash-safety protocol. The three overlap paths it reported are the research-note data files
themselves, not scripts, so scoped re-verification found no discoverable guard test to run for any
of them. The project's `## default landing route` resolved to `merge and push`, so the landing
question was skipped and stated as coming from that default.

## 2026-09-16 — archive run (run 2)

Reached as the merge-and-push route's own same-invocation continuation, never a separate bare
`/flow` call. The merge was verified by ancestry (`git merge-base --is-ancestor`), no PR CLI
involved. Cleanup found nothing to disclose beyond the ordinary build-output noise under
`--ignored` (node_modules, dist, __pycache__); the 19 `.superpowers/sdd/` files force-remove also
destroyed were already durable — rendered to the store and to the pushed ledger before the
worktree came down. The workspace's derived database was never created for this change (no dev
daemon dispatch touched it), so `workspace.sh remove` reported nothing to drop. `## self review`
resolves to `defer` project-wide (the report series ended at kan-380 per that key's own note), so
this run writes the context bundle and runs no reasoning pass, leaving the actual self-review to a
later `/flow-self-review` invocation.
