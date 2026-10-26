## Context

Found by the self-review pass of `/flow-fast` change
`kan-1035-flow-self-review-analyse-visual-verify-prompts`; the fixes ride that change's own branch,
based at its HEAD `f9428445`. Evidence per finding is in `.superpowers/sdd/self-review-pass.md` in
this worktree; the brainstorm design is
`.superpowers/sdd/2026-10-26-self-review-kan-1035-flow-self-review-analyse-visual-verify-prompts-design.md`.

- **A (flow-fix, blast radius 1 file plus this change's `verbatim-moves.txt`).**
  `grep -i 'critical|important|minor|severity'` over `skills/flow/visual-verify.md` and
  `skills/flow/visual-verify-verifier.md` returns nothing; a departure blocks (**Blocking**,
  `skills/flow/visual-verify.md`). `stats/internal/guard/prompttuningcases.go` `ptCheckCase` refuses
  an empty `source.slot` or `source.severity`.
- **B (flow-fix, blast radius 2 files).** `skills/flow-prompt-tuning/SKILL.md` step 4 replays every
  <!-- measured: the blast-radius counts are the files each task's **Files:** names @ f9428445 -->
  `reviewer` case over `bad.diff`/`good.diff`; a verifier-target case's arm A then cannot show the
  miss and is rejected `not reproducible`.
- **C (flow-automation, blast radius 1 file).** `scripts/project-get.sh . lint` at `f9428445` prints
  the opening `bash` fence line as its first line and the closing fence line at line 31, then prose.
  <!-- measured: scripts/project-get.sh . lint | head -1 @ f9428445 → the opening bash fence line -->
  `scripts/run-verify-list.sh` (logic in `stats/internal/guard/runverifylist.go`) already runs
  every non-blank, non-`#` line inside the list's `bash` fence, stdin detached, and is
  what `skills/flow/verify-and-handoff.md` **Inline verify** already uses; `/flow-fast` section 5
  is the one caller that still runs the list by hand.

All three edits were applied in this worktree at `f9428445` and `scripts/check-verbatim-moves.sh`
printed exactly the five sentences this change's `verbatim-moves.txt` lists; the edits were then
reverted so the fixer commits them.

## Decisions

### Visual-verify severity and slot

**ID:** visual-verify-is-important
**Status:** active
**Chosen:** a visual-verify departure is judged as an Important — it blocks the handoff — and its case records `source.slot` `visual-verify` and `source.severity` `Important`; one sentence in angle 8.
**Considered:** teaching the verifier to grade severity — a prompt change to the verifier for a field only the self-review reads; relaxing the guard's non-empty check for visual cases — loses the field every other case keeps.

### Verifier case evidence and replay

**ID:** verifier-case-evidence
**Status:** active
**Chosen:** a verifier-target `reviewer` case also carries `capture.png` and `mockup.png`, copied into the case folder, and step 4 of `/flow-prompt-tuning` replays it with a verifier dispatch judging that capture against that mockup; prose only, no new `RequiredFiles` entry and no guard change.
**Considered:** citing the capture and mockup by path, as angle 7's visual template does — a capture under a run's state directory is removed with its worktree, and the case outlives it; a new case `type` `verifier` — the guard, `RequiredFiles`, the renderer and their tests all change for what the `target` field already tells apart; a guard check that a verifier-target case carries both images — `check-prompt-tuning-cases.sh` has no test file to extend, so the check would bring a new test harness for a case shape no case yet has.

### Lint list runner for /flow-fast

**ID:** flow-fast-uses-run-verify-list
**Status:** active
**Chosen:** `/flow-fast` section 5 runs `<agents repo>/scripts/run-verify-list.sh <worktree> lint`, the runner `/flow`'s **Inline verify** already uses, which parses only the fenced commands.
**Considered:** a `--commands` mode on `project-get.sh` with its own test, as the pass suggested — a second fence parser beside `runverifylist.go`'s, and the caller still runs the list by hand; editing `skills/flow/verify-and-handoff.md`'s "Resolve the commands" sentence — its run already goes through `run-verify-list.sh`, so it carries no defect.

### First review's effort

**ID:** first-review-effort
**Status:** active
**Chosen:** flow-low — three prose edits whose exact text is in the plan and was checked by `check-verbatim-moves.sh`; nothing compiles or runs differently.
**Considered:** flow-medium — the default when a fix touches code other code depends on; none here does.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, Tasks 1–3 in order, one fixer — Tasks 1 and 2 both edit `skills/flow-self-review/SKILL.md`, and the whole change is three applied diffs.
**Considered:** one fixer per task — three dispatches for three `git apply` calls, Tasks 1 and 2 serialised anyway on their shared file.

## Recommended approach

Three prose edits, each the diff its task gives: one sentence in angle 8 (decision
`visual-verify-is-important`), one sentence in step 2a plus one replay bullet in
`/flow-prompt-tuning` step 4 (decision `verifier-case-evidence`; a path citation and a new case
type rejected there), and `/flow-fast` section 5 pointing at `run-verify-list.sh` (decision
`flow-fast-uses-run-verify-list`; a `project-get.sh --commands` mode rejected there).

## Critical files

- `skills/flow-self-review/SKILL.md` — angle 8's severity gate (Task 1) and step 2a's case files (Task 2).
- `skills/flow-prompt-tuning/SKILL.md` — step 4's replay bullets (Task 2).
- `skills/flow-fast/SKILL.md` — section 5's lint run (Task 3).
- `stats/internal/guard/prompttuningcases.go` — the non-empty `source.slot`/`source.severity` check Task 1 satisfies; unchanged.
- `scripts/run-verify-list.sh`, `stats/internal/guard/runverifylist.go` — the runner Task 3 points at; unchanged.

## Existing code to reuse

- `scripts/run-verify-list.sh` (`stats/internal/guard/runverifylist.go`, tested by `stats/internal/guard/run_verify_list_test.go`) — the fenced-command runner.
- `prompttuning.RequiredFiles` (`stats/internal/prompttuning/prompttuning.go`) — unchanged; the two images are extra files beside it.

## Verification

`scripts/check-verbatim-moves.sh` exits 0 with this change's `verbatim-moves.txt`;
`scripts/check-references.sh` exits 0; `git grep -n -F -e 'Run every command in \`## lint\`' -- skills/flow-fast/SKILL.md`
prints nothing.

## Live check

None — the change edits skill prose only; no service, store or runtime state changes.

## Open questions

None — every choice was auto-resolved and is recorded above.
