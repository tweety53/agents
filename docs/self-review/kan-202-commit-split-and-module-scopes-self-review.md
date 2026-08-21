# Self-review — kan-202-commit-split-and-module-scopes

**Rating: 4/5** (operator)

**Outcome:** a commit's scope now names the module, never the change name or a task id — stated in
`/myflow-start`'s `**Commit:**` field spec, enforced by `check-task-commit-fields.py`, carried as an
always-on rule, and applied to finish run 1's two messages at all four sites. Folded in: KAN-202's
contract edit naming `commit-split.sh`, a fifth rule in `check-guard-symlinks.sh`, and a rewrite of
`gather-self-review-context.sh`'s commit resolution, which changing the subjects broke.

**Cost:** 12 tasks, 8 task commits reshaped into 2 at finish, 5 review-panel passes, 7 fix rounds,
23 panel findings with zero left open, 8 parent-run mutations, ~2.3M subagent tokens.

---

## Problems encountered, and the pipeline change that would avoid each — `myflow-fix`

The ledger was written to `.superpowers/sdd/ledger.md`; `preserve-session-records.sh` reads only
`.superpowers/sdd/tasks/progress.md`, reported `skipped: … (absent)` and exited 0. Caught solely
because the skip line was read — otherwise `git worktree remove --force` would have destroyed 334
lines of run record.

Four plan defects surfaced during execution, every one in a field this run's controller wrote and
every one caught by a guard rather than by review: `**Tests:**` parses backticked tokens as declared
test names; `**Files:**` omitted both an always-on rule's required guard registration and the
contract-budget script; two `**Baseline:**` figures were invalidated by a later fix round.

A contract budget was raised to exactly the file's size — zero headroom against the documented
plus-25% formula — and `check-contract-budget.sh` reported `BUDGET-OK` either way.

- **[myflow-fix]** The SDD ledger was written to a path nothing preserves and was one `--force` from destruction; the tenth recorded occurrence of a defect whose merged ticket already absorbs nine — filed: KAN-77
- **[myflow-fix]** `/myflow-fast`'s handoff block predates commit-per-task, telling the operator `staged and uncommitted` and `git diff --cached` when the branch is fully committed and the index empty — filed: KAN-255
- **[myflow-fix]** Run 2's archive and self-review commits still carry change-name scopes, contradicting the rule this change landed minutes earlier; `ARCHIVE_SHA`'s resolution depends on the old shape, so it is not a two-line edit — filed: KAN-256
- **[myflow-fix]** Four plan-field defects, all self-inflicted and all caught mechanically; the field family worked, but its sharp edges surface only when the guard runs, long after planning — declined
- **[myflow-fix]** A contract budget set to zero headroom passed its own guard, the second case this run where a guard passed on a wrong value — declined

## Token and time cost, and what would reduce it without quality loss — `myflow-cost`

Once three fix rounds have run, the escalation ladder's round-count clause fires unconditionally and
permanently, so every later fix demands a Full three-slot pass — including the last one, which was
comment-only and mechanically verified to change zero non-comment lines. Passes 4 and 5 each spent
three slots on a 1400-line diff to confirm prose.

The change grew from 8 tasks to 12 because changing commit subjects broke a downstream consumer no
plan modelled: `gather-self-review-context.sh` located a change's commits by grepping the change
name as the scope — the exact handle this change removes. Planning did not surface that dependency;
the review panel did.

- **[myflow-cost]** The round-count escalation clause fires unconditionally and forever, buying a full panel pass for a comment-only fix; KAN-76 already proposes deciding escalation on blast radius, and this run is direct evidence for it — declined
- **[myflow-cost]** Scope grew 50% because a downstream consumer of the commit-subject shape was unmodelled at planning time; no concrete pipeline change is implied beyond noticing it — declined

## What went well, and how to reproduce it — `myflow-improvement`

The cheapest slot in the roster was by far the most productive: code review at `low` effort found the
only Major in pass 1 — which three per-task reviewers and the full-effort principles slot all missed
— then seven of nine findings in pass 2, five of them Major.

Parent-run mutation testing caught two defective fixes that green suites did not. The first
path-based resolution passed its own tests while resolving the wrong commit, because its fixtures
modelled a commit shape no real run produces.

Resuming the same reviewer for a targeted re-run preserved its context and produced a correction to
the controller's own evidence — it pointed out that `check_boundary` is never reached in the sections
the controller claimed it proved, and verified them directly instead.

- **[myflow-improvement]** The `light` roster's cheapest slot out-found every other slot; keep `light` as the default rather than reaching for `full`, and treat this as a datapoint for KAN-218 rather than a conclusion from one run — declined
- **[myflow-improvement]** Parent mutation testing caught two fixes whose own tests were green; treat a fix subagent's green result as a claim to verify, never as evidence — declined
- **[myflow-improvement]** Resuming a reviewer rather than dispatching a fresh one preserved context and produced a correction; this is already the contract's Targeted behaviour and worked exactly as intended — declined

## What could be automated or moved to a script — `myflow-automation`

Three mutations this run silently failed to apply — a `sed` delimiter collision, a pattern matching a
comment instead of the code, and a regex against code that had moved into an extracted function — and
each reported identically to a surviving mutant. Two turned out to be genuinely caught once re-run
through an edit that asserts the target text was found; had they not been, the change would have
carried three phantom findings, each buying a test for a mechanism already covered.

- **[myflow-automation]** A mutation whose edit never applies is indistinguishable from a survivor; a helper that refuses when the target text is not found, then restores and verifies byte-identity, removes the whole class — filed: KAN-257
- **[myflow-automation]** The ledger-path problem is also an automation candidate — have the preservation script report what it found unpreserved rather than what it looked for — declined

## What could move to the Go app or its persistent storage — `myflow-stats-app`

That the `light` roster's low-effort slot out-found every other slot on this change is recorded in
prose and nowhere queryable. Panel pass count and fix round count per change are likewise unrecorded:
the stage marks capture each stage's duration, but nothing counts the repetitions of
`do.review-panel` as a property of the change.

- **[myflow-stats-app]** Per-slot productivity exists only as prose; KAN-218 covers moving panel findings into the store, which is the numerator this needs — declined
- **[myflow-stats-app]** Panel pass count and fix round count per change are unrecorded and would fold into whatever KAN-218 lands — declined
