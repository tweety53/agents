# kan-441-frontend-implementer-loops-dominate-cache-read

## Why

Self-review finding from KAN-423 (flow-cost angle). Cache-read cost on that change was dominated
by the implementer loops of its frontend tasks — task 18: 145 M tokens, task 13: 83 M, task 15:
49 M, tasks 12/14: 15–29 M — against 2–6 M on its backend tasks. Every turn of an implementer
re-reads its whole context; a frontend task's loop is long because each iteration waits on a
5–10 minute Gradle run, and there are many iterations.

The transcripts behind the figures (gymie-frontend implementer sidechains) show the shape: 10–22
Gradle invocations per implementer, roughly half of them already `--tests`-targeted and the rest
whole-module `:shared:desktopTest` runs, 100–230 assistant turns per task, no `run_in_background`
and no timeouts. The cost is the number of turns, and the test runs are the slow, log-heavy part
of the loop: a whole-module run mid-task costs the same turn a targeted one does, takes minutes
longer, and pours a Gradle log into the context that every later turn re-reads.

The issue's third lever — running Gradle through `run_in_background` plus a monitor — is not
taken. KAN-263 added `FOREGROUND BUILDS` because four implementers did exactly that and produced
nothing: a dispatched subagent that ends its turn on a background command is not resumed when it
exits. And it saves no turn — a foreground call under the Bash tool's cap is one turn; a background
launch and its completion are two.

## What changes

- The implementer dispatch preamble in `skills/flow/implement.md` gains a guarded
  `**TARGETED TESTS:**` paragraph: run only the tests the task's `**Tests:**` field names, through
  the build tool's own selector, once for RED, once for GREEN, and again only after a source edit;
  never the module or repository suite mid-task; pipe the output through `tail` so a green run
  costs lines, not a log.
- The last dispatch bundle's implementer prompt alone carries a `**FULL SUITE:**` paragraph: run
  the resolved `## test` list once, in the foreground, after GREEN and before the commit. A failure
  inside the task's own `**Files:**` is fixed there; any other failure is reported verbatim and left
  unfixed. A report naming such a failure ends the conductor's turn with `## Question` before
  `final-review.diff` is written — the panel never runs on a red branch. `flow.verify` is unchanged
  and remains the gate.
- The panel-fix subagent dispatch in `skills/flow/review-panel.md` carries `TARGETED TESTS` too.
- `scripts/check-dispatch-paragraphs.sh` gains a `targeted` row requiring the paragraph at both
  files, and `scripts/test-check-dispatch-paragraphs.sh` the matching fixture cases.
- `FOREGROUND BUILDS`, its guard row, the verifier and `.flow/project.md`'s keys are untouched.
