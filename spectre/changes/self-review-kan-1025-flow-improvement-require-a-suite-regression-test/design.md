## Context

Self-review fixes for `kan-1025-flow-improvement-require-a-suite-regression-test`, on the change's
own branch `spectre/kan-1025-flow-improvement-require-a-suite-regression-test` in
`/Users/tweety53/Projects/agents-worktrees/kan-1025-flow-improvement-require-a-suite-regression-test`
(the run is inside `<agents repo>`, so the fixes ride the change, `fixes-ride-the-change`).

**F-A.** `### The mutation-testing brief` (`skills/flow/review-panel.md`) ends on "Its test command
runs every test of the package, file or module the mutated code belongs to, never a filter naming
the tests that exist today", per kan-1032. The slot ceiling is 15 minutes; `go test ./internal/guard` <!-- measured: kan-1025 self-review pass evidence, quoted from the plan dispatch -->
is 332 tests, 165 s idle and about 455 s contended, so one package run eats up to half the slot <!-- measured: kan-1025 self-review pass evidence, quoted from the plan dispatch -->
and a mutation pass needs several.
The brief is rendered verbatim into the slot prompt by
`render-slot-prompt.sh` (`stats/internal/guard/renderslotprompt.go`, `rspSection`), so the edit
reaches the slot with no code change. `check-mutation-reproducer-pin.sh` pins the backticked
`# mutation-reproducer` literal and every "first <N> lines" statement in the file; the edit adds
neither. The inverted exit convention is stated plainly in
`skills/flow/test-sanity-reviewer-prompt.md` (lines 117-118) and only by implication in the brief.

**F-B.** `### No forking, and a wall-clock ceiling on every slot` (`skills/flow/review-panel.md`)
says the dispatcher "tracks each in-flight slot's elapsed time itself" with no mechanism.
**Turn discipline** (`skills/flow/implement.md`) already defines the wait — one foreground bounded
loop (48 × 5 s) printing `ready` or `still-running` — and says a ceiling "is tracked across the
calls", again with no mechanism. The ledger shows the breach caught at 21:56:56 → 22:52:00.

**F-C.** `strip_data_heredocs` (`hooks/protect-main-checkout.py`) removes a heredoc body from the
scan in two cases only: a `cat`/`tee` writer to a named file whose written file the rest of the
command does not run (`may_run_code`), and a `PROGRAM_RUNNERS` opener (`python`, `python3`,
`node`), whose body keeps only its git invocations (`program_git_lines`). Every other opener keeps
its body scanned as shell. Reproduced at plan time: the hook, fed
`flow record summary -change x -file - <<'EOF2'` / `step 1 -> step 2; rm nothing > here` / `EOF2`
with `cwd` the main checkout, returns `permissionDecision: deny`.

**Repairs of what the defects wrote.** None: F-A and F-B wrote no record (the timed-out slot's
breach is already in kan-1025's panel record), and F-C's denials wrote nothing.

## Decisions

### Mutation reproducer on a large package

**ID:** fa-family-prefix-filter
**Status:** active
**Chosen:** keep the whole-package rule; add that when one whole-package run, measured once in the
slot before the first mutation, takes more than a third of the slot's ceiling, the command may
filter by the name prefix the mutated code's test family shares (`-run '^TestCheckPanel'`), never
a full test name, so a fix's new test under that prefix still kills the mutant. The reproducer's
`# reproducer-bound:` is that measured run's seconds, doubled. State plainly, beside it, that the
reproducer follows the mutation-reproducer convention: exit 0 when every test passes with the
mutation in place (mutant survived, defect present), non-zero when a test kills it.
**Considered:** raising the slot ceiling — widens a tolerance and still leaves the panel slot
sized by the slowest package; a guard checking `-run` patterns in mutation reproducers — the brief
is the only author's instruction and a prose fix closes the gap at the source; requiring a fix's
test to reuse an existing name — constrains naming for no gain over a prefix.

### A named wake-up for the slot ceiling

**ID:** fb-bounded-wait-elapsed-check
**Status:** active
**Chosen:** in review-panel.md's ceiling paragraph, name the mechanism: the dispatcher notes each
slot's dispatch instant (`date +%s`) when it dispatches, waits on the slot only through
**Turn discipline**'s bounded loop (`skills/flow/implement.md`), and on every `still-running`
compares the elapsed time against the slot's ceiling before re-issuing; a slot past its ceiling
gets no further wait and takes the breach path at once. `implement.md` is left as is: its sentence
already points at this paragraph, which becomes the one place the mechanism is stated.
**Considered:** a background timer or `Monitor` until-loop — a second wait mechanism beside the
bounded loop, and **A wait never outlives its child** forbids unbounded loops; a guard over the
ledger — detects a breach after the fact, which the run already did.

### Heredoc bodies are data unless something runs them

**ID:** fc-heredoc-body-is-data
**Status:** active
**Chosen:** in `strip_data_heredocs`, after the `PROGRAM_RUNNERS` branch and before the `cat`/`tee`
branch, a heredoc whose opener is neither `cat` nor `tee` and whose delimiter closes has its body
dropped from the scan when (a) no word of the opener's segment — the text before the `<<` from the
last segment split, plus the text after the delimiter, split on whitespace, `=`, `$`, `(`, `)`,
backtick and quotes so `x=$(bash` yields `bash` — has a file name in `RUNNER_WORDS`, and (b)
`may_run_code(rest, None)` is false, `rest` the command outside the body; `may_run_code` gains
`target=None`, which skips its `names_target` arm. An unquoted-delimiter body keeps its `$(…)` and
backtick spans scanned exactly as the `cat`/`tee` branch does — that block becomes one helper,
`subst_spans(body)`, returning the spans or `None` when a span cannot be closed (then the whole
body stays scanned), called from both branches. A pipe after the opener keeps the body scanned, as
today.
**Considered:** a closed shell-interpreter list (`sh`, `bash`, `zsh`, `eval`, `source`) as the
only openers that keep a body scanned — misses `sudo bash`, `env sh`, `xargs`, `ssh h`, which
`RUNNER_WORDS` already lists and fails closed on; scanning only the body's git invocations
(`program_git_lines`) for every opener — a commit message or report naming `git commit` would be
denied, the same misfire one layer down; adding `flow` to an allow-list — fixes one opener of many.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — Task 3 changes the main-checkout containment hook, where a misfire
either blocks every session or lets a write through.
**Considered:** `flow-low` — Tasks 1 and 2 are prose, but the hook's fail-closed boundary needs a
careful read.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, `fix` — Tasks 1, 2, 3 in order, one fixer on `flow-low`/opus in the change's
worktree; Tasks 1 and 2 both edit `skills/flow/review-panel.md`.
**Considered:** one fixer per task — three agents in one worktree collide, and Tasks 1 and 2 share
a file.

## Open questions

none

## Recommended approach

Two prose edits to `skills/flow/review-panel.md` (F-A, F-B) and one hook change with its test
(F-C). Rejected alternatives: see `fa-family-prefix-filter`, `fb-bounded-wait-elapsed-check`,
`fc-heredoc-body-is-data`.

## Critical files

- `skills/flow/review-panel.md` — `### The mutation-testing brief`; `### No forking, and a wall-clock ceiling on every slot`.
- `hooks/protect-main-checkout.py` — `strip_data_heredocs`, `may_run_code`.
- `stats/internal/repotest/protect_main_checkout_test.go` — hook tests.

## Existing code to reuse

- `RUNNER_WORDS`, `may_run_code`, `SUBST_SPAN_RE` and the `cat`/`tee` branch's span handling in `hooks/protect-main-checkout.py`.
- `pmFixture` (`bash`, `run`, `git`, `write`) in `stats/internal/repotest/protect_main_checkout_test.go`.
- **Turn discipline**'s bounded wait loop in `skills/flow/implement.md`.

## Verification

- Tasks 1–2: `scripts/check-mutation-reproducer-pin.sh`, `scripts/check-dispatch-paragraphs.sh`, `scripts/check-markdown-integrity.sh`, `scripts/check-references.sh` exit 0; `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestRenderSlotPrompt' -count=1` passes.
- Task 3: `cd stats && make -s web-dist && go test ./internal/repotest/ -run 'ProtectMainCheckout|TestCommitThenPush' -count=1` passes; `python3 -m py_compile hooks/protect-main-checkout.py`.

## Live check

None — the change edits two prose sections and a PreToolUse hook that holds no runtime state; Task 3's test drives the hook itself.
