## Context

The six findings of `proposal.md` ride kan-918's own branch,
`spectre/kan-918-flow-improvement-port-the-git-snapshot-and`, in its worktree, as one self-review
change. Every fix was drafted, applied and checked in that worktree at `eff5ca1c` before planning,
then reverted; `tasks.md` carries each as the diff to apply.

**F4's trace.** The slot's transcript is
`~/.claude/projects/-Users-tweety53-Projects-agents/5c844d2a-d709-4945-b838-b398826a1566/subagents/agent-a49d6b737c5810fb7.jsonl`
(meta: `flow-low`, `sonnet`, "panel-1-primary+principles re-review of fixes"). Its Agent prompt
carried the baseline pointer, MODEL HANDSHAKE, CONTEXT BUNDLE, the reproducer rule and "Read
`…/slot-prompt-1-primary+principles.md` in full first — it is your brief"; the REPORT FILE
paragraphs live only in that brief. The slot never read it: one Bash call (`git show bd7ff841`
and three greps), then `SubagentHandback` with two verdict lines, from 17:45:39Z to 17:45:53Z.
Its whole context was the 19,515-token prompt prefix — the ledger's cache-read figure — and the
10m4s on the ledger is the parent closing the dispatch, not the slot working. The parent then
wrote both report files as `no verbatim report captured` placeholders. The defect is the
pipeline's: the one obligation the round's guards check (`check-panel-findings-closed` refuses a
re-run with no report file, KAN-915) reaches the slot only through a file it may skip.

**F3's measurement.** F15 was run over all 195 archived plans in the main checkout, each at the
parent of the commit that first added its `tasks.md`, by a throwaway script calling the patched
guard's `_check_f15`. A first draft matching any declaration starting with the name, in any
directory, flagged four plans; three of its hits were new tests whose names prefix a test in
another package (kan-760's `TestSpecRoot` and `TestMain`, kan-916's `TestUsage`). The rule
recorded under **f15-match-rule** flags two plans, both artifacts of the measurement's base:
kan-357's `tasks.md` was committed after `51d73706` created `agentfile_test.go`, and kan-873's
after `49f7b3dd` added `TestCheckReferences`. At a true plan-time base neither name existed, so
the rule raises no false positive over the archive, and it flags kan-918's task 3.

**F6's measurement.** Serially the harness took 6:30 and 5:39 wall on two runs at `eff5ca1c`;
with the fix, 52 s, 195 checks passing either way.

## Live check

None: no daemon, store or schema changes — `flow record finding -from` sends each line through the unchanged single-finding call, which `TestRecordFindingFromFile` exercises against a fake daemon.

## Decisions

### PLAN FIELDS binds the fold route only

**ID:** plan-fields-fold-only
**Status:** active
**Chosen:** the PLAN FIELDS paragraph and the fix round's walk name the `fold-fixup.sh` route; an on-top commit carrying `Flow-Key:` is outside every task and edits no task's fields, the reason recorded under the PLAN FIELDS entry of `SKILL-rationale.md`.
**Considered:** teaching `check-task-records` rule 2 to accept Files widened by a later `Flow-Key:` commit — the record would then claim files the task's commit never touched, the drift rule 2 exists to catch; widening the fields and also folding the fix — the on-top route exists because the branch is already pushed.

### The live check waits for both groups

**ID:** live-check-after-groups
**Status:** active
**Chosen:** "once both groups of the lint and test list have exited — never while either still runs", with kan-918's watched-tree failure as the stated reason, in the sentence that orders the step.
**Considered:** writing `live-verification.md` outside the watched tree — the record is a planning path the write-in-progress commit carries, and the check's other writes (a started stack) would still race the polled suite.

### F15 extends check-plan-shape

**ID:** f15-in-plan-shape
**Status:** active
**Chosen:** a new finding F15 in `scripts/check-plan-shape.py`, the existing Python guard, reading the base tree with `git grep` at the merge base of HEAD with the branch's recorded base, else `refs/remotes/origin/HEAD`; a plan outside git, or with no resolvable base, is not checked for F15.
**Considered:** a new guard — a second plan-time pass over the `tasks.md` the shape guard already parses; reading the working tree or HEAD — at lint time the plan's own later commits add tests that would read as pre-existing.

### F15 matches by directory and file subject

**ID:** f15-match-rule
**Status:** active
**Chosen:** a name is flagged when a test-shaped file outside the task's `**Files:**` and `**Allowed-collateral:**`, in a directory one of those files lies in, declares (`func`/`fun`/`def`) that exact name, or a name starting with it while the file's subject (`fold_fixup_test.go` → `fold_fixup`) folds to the name without its `Test` prefix. A new `TestFoo` beside an existing `TestFooBar` passes, in the task's own file and in another file of the package alike.
**Considered:** any declaration starting with the name, anywhere — three false positives over the archive (**Context**); exact names only — misses kan-918's `TestFoldFixup`, which was only a prefix; exact names in any directory — Go allows one name per package, and `TestMain` is in many.

### The REPORT FILE path rides the Agent call

**ID:** report-path-in-call
**Status:** active
**Chosen:** the Agent call's prompt carries each role's REPORT FILE path beside `read <rendered path> in full first — it is your brief`, so a slot that skips its brief still owes its file.
**Considered:** the whole rendered brief inline in the Agent call — the render-to-file split exists to keep the call small; a parent re-dispatch when the file is missing — the placeholder and `check-panel-findings-closed` already handle a missing file, the defect is the slot never being told.

### Findings recorded from a JSON-lines file

**ID:** record-finding-from-jsonl
**Status:** active
**Chosen:** `flow record finding -change <name> -from <file>`, each non-blank line a JSON object whose keys are the single-finding flags without their dash, recorded in file order through the existing single-finding path; the first failing line stops the run naming its line. `review-panel.md` names `<abs-worktree>/.superpowers/sdd/findings-<round>.jsonl` as the file.
**Considered:** TSV — a note carrying a tab breaks a row, and optional columns need a placeholder; a batch store endpoint — validation, the store call and the journal fallback already work per finding, and a batch path would be a second copy of them.

### One background subshell per harness case

**ID:** integrate-harness-parallel
**Status:** active
**Chosen:** every `# --- <case> ---` section of `test-integrate-change.sh` runs in `( … ) &` with its own `$WORK`, counters and log, the EXIT trap re-armed inside; the two helpers defined mid-file and used by later cases (`peer_fix_case`, `ui_a`/`flag`) move above the first case; the parent waits, replays each log and sums each case's last-line tally, a log with no tally counting as a failure.
**Considered:** a concurrency cap — the 41 cases finish in 52 s uncapped, and a cap is configuration nobody asked for; one function per case — the same isolation with a larger diff.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — F15 is new guard logic run on every plan, and the harness restructure changes how its failures are counted; both need a full first look.
**Considered:** `flow-low` — four of the six are prose or a one-path CLI flag, each diff already verified.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, one fixer — key `self-review-kan-918-flow-improvement-port-the-git-snapshot-and-fix`, Tasks 1 to 6 in order: each applies a diff verified at `eff5ca1c`, and Tasks 1, 4 and 5 edit `skills/flow/review-panel.md` in that order.
**Considered:** one fixer per task — six dispatches for six `git apply` runs; prose and code as two groups — no shared state between them to justify the second dispatch.

## Open questions
