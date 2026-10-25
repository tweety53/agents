## Context

- The self-review pass (`skills/flow-self-review/SKILL.md`, step 2) is one opus dispatch over the
  run's bundle. The bundle carries the panel record, so every finding with its round, slot,
  severity and status, plus the `git log --stat` of the change's commits. The pass already reads
  beyond the bundle for the commits and files the bundle names (angle 7 does), and the session
  already extracts files out of git history into a scratch directory before step 4.
- The angle table in step 2 is canonical, and `scripts/check-self-review-report.sh` (Go,
  `stats/internal/guard`) reads that table to decide which sections a report must carry. A new row
  is therefore enforced in every new report without a guard change; reports written before the row
  existed are already tolerated (the guard's test fixtures carry five- and six-angle reports).
- The prompts being tuned all live in `<agents repo>`: the implementer instructions
  (`skills/flow/implement.md`, `skills/flow/sdd-dispatch.md`,
  `skills/flow/engineering-principles.md`) and the reviewer instructions
  (`skills/flow/*-reviewer-prompt.md`, `skills/flow/reviewer-calibration.md`). Cases from a run in
  any project are therefore saved in `<agents repo>`.
- Ideas taken from `/claude-api prompt-audit` and its eval-hillclimb guide: describe the behaviour,
  never paste the incident's content into the prompt; scale a change to how many cases show the
  behaviour, never to one vivid failure; treat a prompt change as a hypothesis proven by a
  before/after behavioural probe; one idea per change; the keep list (context and reasons are never
  cruft, and re-baselining can add text as well as remove it).

## 1. The `flow-prompt-tuning` angle

Row 8 of the angle table: `| 8 | Instruction gaps that cost review-panel fix rounds, and the generic
implementer or reviewer prompt update that would have avoided them | flow-prompt-tuning |`.

**What it looks at.** The review-panel findings of the run, each judged alone on its own severity —
no severity needs another alongside it. A Critical finding qualifies; an Important finding
qualifies. A Minor finding qualifies only when the pass judges it genuinely useful, meaning it is
in a class the prompts should handle and would recur. A finding is a case only when its root cause
is a generic gap in instructions:

- **`implementer`**: a rule in the implementer instructions would have prevented the bad code.
- **`reviewer`**: either a **miss**, a defect raised only in a later round or by a later slot that a
  round-1 reviewer instruction would have caught, or a **false positive**, a finding later disproved
  or withdrawn that cost a fix round for nothing.

**The bar for picking a case**, all four stated in the case's `why_picked`:

1. The defect class in one sentence, with no names from the task.
2. The current prompt text quoted, showing that no rule covers the class. The case's target file
   and its `before` text are that quote.
3. Why the update generalises: other code it would apply to, or other runs that show the class.
   Being one vivid incident is never enough.
4. Why the update is not patch accretion: it states a principle, not a special case.

**Zero cases is a normal result.** The section then carries the none-marker like any empty angle.
The pass never lowers the bar to produce a case. A case that fails any of the four is not picked,
and the pass does not mention it.

**Disposition.** Every case is recorded `noted`, with the report line naming its folder
(`- **[flow-prompt-tuning]** <defect class> — <type> — docs/prompt-tuning/<case-id>/ — noted`). A
case is never fixed in step 3, never offered for filing in step 4 and never asked about: the angle
needs no operator answer, in a run or standalone.

## 2. The case folder

`<agents repo>/docs/prompt-tuning/<case-id>/`, `<case-id>` = `<change-name>-<n>`, `<n>` counting
from 1 within the change:

| File | Content |
|---|---|
| `case.json` | `id`, `type` (`implementer`/`reviewer`), `reviewer_kind` (`miss`/`false-positive`, reviewer only), `source` {`project`, `change`, `jira`, `repo`, `fix_sha`, `finding_ref`, `round`, `slot`, `severity`}, `target` (agents-repo-relative prompt path), `target_sha` (the agents commit the `before` text was read at), `defect_class`, `why_picked`, `before`, `after` |
| `finding.md` | the reviewer's own sentence, severity, location |
| `bad.diff` | the code as the implementer first wrote it: the hunk(s) the finding names, from the commit before the fix |
| `good.diff` | the fix-round commit's diff for those hunks (`git show <fix_sha>`) |
| `task.md` | the task text as `tasks.md` recorded it (implementer cases; `none` for reviewer cases) |
| `index.html` | rendered by `flow prompt-tuning render <dir>` from the files above, never written by hand |

`index.html` shows, in order: the type, the defect class, the finding with the bad code, the fix with
the good code, why it was picked, and the prompt update with the before and after texts side by side
and the target path. It is one self-contained page: inline CSS, no script, no external asset. Code
and prompt text are HTML-escaped through `html/template`.

**Who writes it.** The pass returns every case's fields (it is read-only). This session, after the
pass and before step 3, extracts `bad.diff`, `good.diff` and `task.md` with `git -C <repo> show`
exactly as angle 7's extraction does, writes `case.json` and `finding.md`, runs
`flow prompt-tuning render`, and runs `scripts/check-prompt-tuning-cases.sh` on the new folders.
A red guard line is fixed by rewriting the offending field. If no generic rewrite passes the guard,
the case is dropped and the report says so in its line, disposition `noted`.

## 3. The genericity guard

`scripts/check-prompt-tuning-cases.sh [dir]` is a shim over `flow-guard check-prompt-tuning-cases`
(`stats/internal/guard/prompttuningcases.go`). It scans every folder under `docs/prompt-tuning/`,
or `dir`, and `LEDGER.md`:

- every file of the table above present; `case.json` parses, every field non-empty; `type` in the
  enum; `reviewer_kind` present exactly when `type` is `reviewer`;
- `target` exists under `skills/` and `before` occurs verbatim in `git show <target_sha>:<target>`;
- **genericity**: `after`, and the part of `after` that differs from `before`, contain none of: the
  source change name, its Jira key, any `kan-<digits>` (case-insensitive), any path or basename
  named in `bad.diff`/`good.diff` headers;
- `index.html` is byte-identical to what `flow prompt-tuning render` produces for the folder;
- every `LEDGER.md` entry has the shape of §5.

Exit codes: 0 clean, 1 one `file: reason` line per violation, 2 cannot answer. It is added to
`.flow/project.md`'s `## lint`.

## 4. Landing the cases

The case commit, `docs(prompt-tuning): <name> cases`, rides the same branch as the pass's fixes:

- in a run whose canonical repository is `<agents repo>`, the change's own branch (`spectre/<name>`
  on `/flow`, `<name>` on `/flow-fast`), landed by the run's own route;
- otherwise, and standalone, `self-review-<name>`. When step 3 created no worktree (no non-`big`
  finding), this session creates it with `pipeline-fix.sh start <name>`, commits the cases and
  lands it by the agents repo's `## default landing route`, as step 3 item 5 lands a fix branch.

`check-loop-landing.sh` refuses a `self-review-*` branch that has no plan, fix and clean review in
flowd. It gains one admission: a branch whose every commit since `origin/<agents-base>` touches only
paths under `docs/prompt-tuning/` lands with `LOOP-LANDING-OK`, without flowd rows. Such a branch
carries no code, so there is nothing for the loop's review to have judged.

## 5. `/flow-prompt-tuning`

A standalone command (`skills/flow-prompt-tuning/SKILL.md`,
`commands-claude/flow-prompt-tuning.md`). It takes no argument and is not a pipeline stage: no
state file and no stage marks. It runs from the agents main checkout, which stays read-only. All
work happens in a worktree `<agents repo>-worktrees/prompt-tuning-<UTC date>`, on a branch of the
same name off `origin/<default-branch>`.

1. **Walk.** List `docs/prompt-tuning/*/`. None: print `none pending` and stop. Run the guard; a
   red case is reported and left untouched.
2. **Group.** Cases with the same `target` and the same defect class become one group with one merged
   generic `after`. The merge states the principle once (one idea per update). A group's evidence is
   every case in it.
3. **Re-anchor.** If a group's `before` no longer occurs in the current target, re-locate it. If
   the gap is gone, the group is rejected as `already covered`, with the line now covering it
   quoted. Otherwise the update is rewritten against the current text and re-checked by the
   genericity rules.
4. **Replay (A/B).** For every case of the group, on opus, 3 runs per arm. Arm A carries the
   target's current text, arm B the text with the update applied.
   - `reviewer` case: a reviewer dispatch over `bad.diff` (a miss) or `good.diff` plus the finding's
     location (a false positive), with the target prompt as its instructions.
   - `implementer` case: an implementer dispatch given `task.md` and the pre-change code, writing in
     a scratch directory under the session's scratchpad, never in a repository.
   - A grader dispatch (opus) judges each run against `finding.md`: was the defect present (or
     raised), yes or no, with one line of evidence.
   - **Pass rule:** arm A shows the problem in at least 2 of 3 runs, and arm B avoids it in at least
     2 of 3. A shows it in fewer than 2: rejected `not reproducible`. B fails: rejected `no effect`.
     A group passes when every one of its cases passes.
5. **Apply** each passing group's update to its target in the worktree; run `.flow/project.md`'s
   `## lint` lines for the touched files.
6. **Ledger and cleanup.** Append one entry per case to `docs/prompt-tuning/LEDGER.md` and
   `git rm -r` its folder. Rejected cases are recorded too, so the same case is never re-mined:

   ```markdown unverified:the ledger entry shape this change defines; the guard of Task 2 enforces it
   - <UTC date> · `<case-id>` · <type> · `<target>` · <defect class> — <one-line summary of the update> — applied <sha> | rejected: <reason>
   ```

   For an applied case, `<sha>` is the commit that applied it. The ledger entry and the folder
   removal go in one commit, `docs(prompt-tuning): ledger <date>`, after the update commits.
7. **Land** by the agents repo's `## default landing route`, without asking. Then remove the
   worktree.

The final report is one line per group: target, defect class, applied or rejected with its
reason.

## Step-by-step breakdown

### `flow prompt-tuning render`

**What:** a `flow` CLI subcommand plus `stats/internal/prompttuning` (the `Case` type, `Load`,
`Render`) that renders a case folder's `index.html` through `html/template`.
**Why:** a hand-written HTML page drifts and escapes nothing; one renderer makes the page
deterministic, so the guard can compare it byte for byte.
**Uses:** `stats/cmd/flow/main.go`, `html/template`, `encoding/json`.

### `check-prompt-tuning-cases`

**What:** a flow-guard check with a `scripts/` shim. It validates case completeness, the verbatim
`before`, genericity, the rendered page and the ledger shape.
**Why:** "generic, not tied to the task" is the requirement the angle is most likely to slip on. A
mechanical check catches what a prose rule would not.
**Uses:** `stats/internal/guard`, `stats/internal/prompttuning`, `scripts/lib/flow-guard.sh`,
`.flow/project.md` `## lint`.

### Loop-landing admission

**What:** `check-loop-landing` admits a `self-review-*` branch whose commits touch only
`docs/prompt-tuning/`.
**Why:** a pass with cases but no non-`big` findings must still land its cases without a fake
plan, fix and review.
**Uses:** `stats/internal/guard/looplanding.go`.

### Angle 8 in `/flow-self-review`

**What:** the table row, its remit and picking bar, the case-writing step, the `noted`
disposition, the landing rule, and "seven" becoming "eight" wherever the angle count is stated.
**Why:** the pass is the only place that reads a whole run's findings with their fixes.
**Uses:** `skills/flow-self-review/SKILL.md`, `skills/flow-self-review/SKILL-rationale.md`,
`skills/flow-contracts/finish-contract-run1.md`, `skills/flow-fast/SKILL.md`,
`commands-claude/flow-self-review.md`, `scripts/pipeline-fix.sh`.

### `/flow-prompt-tuning`

**What:** the new skill and command: walk, group, re-anchor, A/B replay, apply, ledger, cleanup,
land.
**Why:** the operator wants proposed updates validated against their repro data before they change
a prompt, and a durable trace of each once its folder is gone.
**Uses:** `skills/flow-prompt-tuning/SKILL.md`, `commands-claude/flow-prompt-tuning.md`,
`docs/prompt-tuning/LEDGER.md`, `check-prompt-tuning-cases.sh`, `project-get.sh`, the Agent tool
(opus).

### Index rows

**What:** the skill index rows in `CLAUDE.md`, `AGENTS.md`, `templates/CLAUDE.md`,
`templates/AGENTS.md`, `skills/README.md` and `README.md`.
**Why:** every command is listed in each harness's instruction file and in the READMEs.
**Uses:** `scripts/check-installed-rules.sh`, `scripts/check-references.sh`.

## Decisions

### cases-in-agents-repo

**ID:** cases-in-agents-repo
**Status:** active
**Chosen:** cases are saved under `<agents repo>/docs/prompt-tuning/<case-id>/`, whatever project
the run was in. This was the operator's answer.
**Considered:** the run's own project — rejected because the tuned prompts live in the agents repo,
and `/flow-prompt-tuning` would then have to run in every project.

### replay-ab-validation

**ID:** replay-ab-validation
**Status:** active
**Chosen:** `/flow-prompt-tuning` A/B-replays each update on the case's repro data and applies it
only on the pass rule of §5 (operator's answer).
**Considered:** operator review only — no evidence that the update changes behaviour; replay, then
operator approval — rejected because the operator chose the automatic route.

### apply-and-land

**ID:** apply-and-land
**Status:** active
**Chosen:** a passing update is applied and landed in the same run by the agents repo's default
landing route, without asking (operator's answer).
**Considered:** summarize only, with the operator applying the edit — rejected because it adds an
operator step per case.

### ledger-file

**ID:** ledger-file
**Status:** active
**Chosen:** one append-only `docs/prompt-tuning/LEDGER.md`, one entry per case; the case folder is
deleted entirely (operator's answer).
**Considered:** a stub `SUMMARY.md` left in each folder — rejected because the folders would pile
up forever.

### reviewer-both-kinds

**ID:** reviewer-both-kinds
**Status:** active
**Chosen:** a `reviewer` case is a miss or a false positive (operator's answer).
**Considered:** misses only — rejected because a false positive costs a fix round just the same.

### minor-when-useful

**ID:** minor-when-useful
**Status:** active
**Chosen:** each finding qualifies on its own severity — a Critical, an Important, or a Minor when the pass judges it
genuinely useful under the same four-part bar (operator's answer).
**Considered:** Critical and Important only — rejected because the operator asked for useful Minors.

### never-ask-in-run

**ID:** never-ask-in-run
**Status:** active
**Chosen:** the angle asks nothing, in a run or standalone. Every case is `noted`, never fixed or
offered (operator's instruction).
**Considered:** offering cases in step 4's filing prompt — rejected by the operator.

### picking-bar

**ID:** picking-bar
**Status:** active
**Chosen:** the four-part `why_picked` of §1, with zero cases a normal result (my call, from
prompt-audit's overgeneralization and patch-accretion rows).
**Considered:** any finding a prompt line could address — rejected because it would turn every round
into an incident-specific prompt patch.

### genericity-guard-in-go

**ID:** genericity-guard-in-go
**Status:** active
**Chosen:** a flow-guard check with a shim, scanning case completeness, the verbatim `before`,
genericity, the rendered page and the ledger (my call; new guard logic is Go per the repository's
convention).
**Considered:** a prose rule in the skill alone — rejected because a rule nothing checks is the
unenforced-instruction pattern.

### rendered-html

**ID:** rendered-html
**Status:** active
**Chosen:** `index.html` is rendered by `flow prompt-tuning render` from `case.json` and the repro
files (my call).
**Considered:** the session writing HTML by hand — rejected because it is non-deterministic, cannot
be guarded byte for byte and risks unescaped code.

### cases-ride-fix-branch

**ID:** cases-ride-fix-branch
**Status:** active
**Chosen:** the case commit rides the pass's fix branch (`spectre/<name>`, `<name>` or
`self-review-<name>`). With no fixes, `self-review-<name>` is created for the cases alone and
admitted by the new loop-landing rule (my call, following one-task-one-worktree).
**Considered:** a separate `prompt-tuning-<name>` branch — rejected because it adds a second
landing per run.

### three-runs-per-arm

**ID:** three-runs-per-arm
**Status:** active
**Chosen:** 3 runs per arm, pass on 2 of 3 each way, opus for every replay and grader dispatch (my
call). LLM output is nondeterministic, so an approximate match is the requirement here, not a
widened tolerance.
**Considered:** 1 run per arm — rejected because one sample cannot tell an effect from noise; 5
runs — rejected as cost without a stated need.

### no-review-panel-for-tuning

**ID:** no-review-panel-for-tuning
**Status:** active
**Chosen:** `/flow-prompt-tuning`'s gate is the replay plus the `## lint` guards. No review panel
runs over the prompt edits (my call).
**Considered:** a reviewer dispatch over the edit — rejected because the replay already measures
the behaviour the edit is for.

## Open questions
