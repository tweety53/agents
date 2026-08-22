# Self-review — kan-77-sdd-ledger-canonical-path

**Rating: 3 / 5 — got there, expensively.**

11 fix rounds, 8 review rounds, 23 panel findings. The change's own subject — the ledger's canonical
path, the rescue, the MISSING outcome, artifact brevity — was clean at review pass 1 and never moved
after it. Everything since was the operator-approved guard-gap scope addition.

## Problems encountered, and what pipeline change would avoid them — `myflow-fix`

- **[myflow-fix]** Planning artifacts staged in the main checkout blocked an unrelated change's merge and archive; resolving it required an operator decision about another change's uncommitted work — filed: KAN-56
- **[myflow-fix]** `git commit --fixup` commits the whole index and the pipeline never warns fix subagents; it swept a wrong task's file and then a planning path into task commits, both needing history surgery — filed: KAN-273
- **[myflow-fix]** A `Tests:` field's backticked tokens are read as declared test names, so prose in that field fails the guard — filed: KAN-274
- **[myflow-fix]** `Baseline:` is skipped-not-verified here, so it went stale three times across eleven fix rounds while claiming 113 against a real 170 — filed: KAN-275

## Token/time cost, and what would reduce it without quality loss — `myflow-cost`

- **[myflow-cost]** Six instances of one root cause were found one per round until a sweep was dispatched; findings per round did not decline, and the sweep then found three more in a single pass — filed: KAN-276
- **[myflow-cost]** The panel re-read the whole branch diff eight times — 3,793 lines against an 886-line actual delta — with the scoped diff built by hand from round 7 — filed: KAN-277

## What went well, and how to reproduce it — `myflow-improvement`

- **[myflow-improvement]** Asking a fix subagent "is anything left that two files must agree about with nothing enforcing it?" found a defect no review round had, and triggered the sweep that found three more — declined
- **[myflow-improvement]** Cross-harness pinning replaced a comment that had been wrong three times; mutating the shared module now fails both guards' suites, which the comment never did — declined
- **[myflow-improvement]** Telling reviewers that a clean report is a meaningful signal preceded Code-review's one genuinely clean round, after it had found a real defect every previous round — declined
- **[myflow-improvement]** `openspec archive` refused and caught four scenarios about to be silently deleted, three of them security containment scenarios — declined

## What could be automated or moved to a script — `myflow-automation`

- **[myflow-automation]** Nothing checks a MODIFIED block against the current spec until archive time; `openspec validate --strict` passed all nineteen rounds and the abort came at the final irreversible step — filed: KAN-278
- **[myflow-automation]** Mutation-proving is entirely hand-run — roughly forty mutations, one of which was blunt enough to prove nothing — filed: KAN-279
- **[myflow-automation]** Commit messages quoting assertion counts go stale every fix round; this one went stale twice, once still asserting a parity claim the code no longer had — filed: KAN-280

## What could move to the Go app or its persistent storage — `myflow-stats-app`

- **[myflow-stats-app]** This run is the store-native argument lived: 23 findings, ~40 mutations and 19 rounds recorded in one markdown file, at a path this very change had to fix — filed: KAN-258
- **[myflow-stats-app]** Nothing reconciles the stage-mark journal after a store outage, so this run's stage timings are permanently partial — declined

## Orchestrator errors, all caught and corrected

Recorded because a run that hides its own mistakes teaches nothing. None is a finding against the
pipeline; each is mine.

1. `git add -N` on the rescued ledger staged an **empty blob** — had it reached finish run 1, that
   ledger would have been committed empty, destroying the record this change exists to preserve.
   Caught by a fix subagent hitting it as a stash failure.
2. Relocating the F12 fix into task 2's commit **reverted it at HEAD** on the first attempt. Caught
   by reading the file rather than trusting the rebase.
3. A fix-round dispatch **bundled two tasks' files into one commit**, producing error 2's cleanup.
4. A **blunt mutation** broke a function outright — 61 and 23 failures — proving nothing about the
   mechanism it meant to test. Discarded and redone surgically.
5. A dispatch **misstated which harness had a cleanup trap**. The subagent verified rather than
   followed, and said so.

## Deferred and recorded

- **KAN-272** — three unclosed-fence cases that name the wrong defect; linked to KAN-114 and KAN-77.
- The commit-fields harness leaks fixture directories; `FENCE_RE`'s toggle is naive about delimiter
  kind. Both pre-existing, both stated in the panel record rather than left implicit.

## Operator override recorded

`no-direct-pushes-to-main` was **explicitly overridden** for this change: it landed by merge-and-push
directly to `main`, which carries no branch protection, so that rule was the only enforcement. The
override is recorded here and in the implementation commit's body.
