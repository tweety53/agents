# Self-review — kan-108-cut-the-time-and-token-cost-of-a-myflow-do-run

**Rating: 3/5** — delivered, expensively.

**Route:** `/myflow-fast` created the change and implemented it; `/myflow-finish` integrated by
merge-and-push and archived. Recorded settings: planning effort `default`, all three model roles
`sonnet`, review panel roster `light`.

**Shape of the run:** 7 tasks, 6 commits, 7 review panel passes, 5 fix rounds plus one
operator-directed trim, 56 findings. Ended with 4 findings open by the operator's decision and 3
panel slots holding stale results, both recorded on the panel record's own marker lines.

---

## 1. What went wrong, and the pipeline change that would avoid it

### The implementer ran on Sonnet, and the fix rounds paid for it

`/myflow-fast` records `models.implementation: sonnet` rather than asking, which silently overrides
the standing policy defaulting implementers to Opus. Much of what the panel then found was ordinary
Bash: `[[ -r ]]` true for a directory, no `--` before a grep filename, `pgrep -P` walking one level,
a path token split on space where the tokenizer split on space and tab, quotes missing from a banned
set, `set -e` dropped from a harness. Five fix rounds of it.

Not filed at the operator's decision. Recorded here because the causal chain is legible: the fast
path's economy choice on the implementer was paid back, several times over, in panel passes that
each cost more than the model difference would have.

### The trim shredded prose, and the verification looked at the wrong file

`skills/myflow-do/SKILL.md` reached 58617 bytes against a 58623 budget. The operator chose trimming
over re-anchoring, and the trim moved reasoning into `SKILL-rationale.md`, tearing sentences apart: a
paragraph promising a reason that never arrived, a sentence ending mid-clause at "the trade", a
blockquote whose first line lost its marker, and an operational rule duplicated with a pointer from
the rationale file into itself.

Every guard passed the whole time. `check-references.sh` checks that headings resolve; they did.
`check-contract-budget.sh` checks size; the file had shrunk. Nothing checks that a sentence is whole.

The verification actively concealed it: the trimming agent supplied a rule-presence list, that list
covered `SKILL.md` — the source — and nobody read `SKILL-rationale.md`, the destination. Found two
passes later by a reviewer who read it. **Filed as KAN-157.**

### A `.pyc` was committed and had to be un-committed

The first implementation commit included
`scripts/__pycache__/check-plan-provenance.cpython-314.pyc`, because the Python guard wrappers
regenerate it and nothing ignored it. Caught before any push; both commits were rebuilt from the
merge base with `__pycache__/` added to `.gitignore`, so no trace remains in branch history. Fixed
in-change, nothing to file.

### The SDD ledger did not exist until finish

Model policy requires each dispatch to record its model in the ledger. This run recorded models in
the dispatch prompts and in each subagent's report, and `/myflow-finish` run 1 found
`.superpowers/sdd/tasks/progress.md` absent at the preservation step. It was reconstructed from
session history — accurate, and after the fact, which the committed ledger says about itself.
**Filed as KAN-155.**

---

## 2. Token and time cost, and what would reduce it without losing quality

### The saving KAN-108 was filed for is real, and is now bounded by a different trigger

The reworded clause fired **zero times** across seven passes, which is the ticket's item 1 working.
But two fix rounds measured 301 and 313 changed lines and escalated to Full on the **~150-line**
condition instead, at seven slots each. From round 3 onward the "three or more fix rounds" condition
fired too, so those rounds were escalated by two independent conditions. **Filed as KAN-154.**

### Four of seven passes ran at full breadth, and the fourth was stopped mid-flight

Pass 1 put the optional-slot prompt, it went unanswered, and the contract's "including all of them is
recommended" widened a `light` roster from three slots to seven — after which every Full escalation
re-ran the same seven. The operator stopped three slots during pass 4 because the cost had stopped
buying anything. **Filed as KAN-156.**

Worth stating against that: the conditional slots found both Criticals the required three missed —
the unconstrained execution of reviewer-supplied commands, and the one-level descendant walk that
left a live process behind under a clean verdict. Excluding them by default is not free.

### The one number this run cannot report is its own cost

Every subagent reported its token usage to the parent. Nothing accumulated it. KAN-108 exists
because someone measured KAN-107 at ~3.44M subagent tokens and attributed 53% to the panel; KAN-108's
own total is recoverable only from a session transcript. A change filed to cut token cost cannot say
whether it cut any. Filed with the ledger work as **KAN-155**, and it is the most useful thing on
this list.

### What actually reduced cost, measurably

Replacing described procedure with an invocation. Task 7 moved the reproducer-execution rules into
`scripts/run-reproducer.sh` and took `skills/myflow-do/SKILL.md` from 54849 to **52389 bytes** — 2460
bytes out of a file every `/myflow-do` run loads, while making the rules enforceable. Prose that
describes machinery is paid for on every run and verified on none.

---

## 3. What went well, and how to reproduce it

### The reproducer rule caught its own reviewers

At pass 1, three of Adversarial's ten reproducers demonstrated the *opposite* of their claims — the
guard already rejected scattered markers, reversed identifiers and a leading-zero total. Because the
parent ran each reproducer before dispatching a fix, those three were re-scoped from Critical and
Major behaviour defects to one Minor coverage finding instead of sending a fixer after defects that
did not exist. That is the change validating itself on its own review, and it is reproducible by
keeping the rule the change added: run it, read the output, then decide.

### Mutation testing found what nobody else did

Adversarial found three surviving mutants in `check-panel-reproducers.sh` at pass 1 and the
one-level-`pgrep` Critical at pass 7, each missed by every other slot. When the operator dispatched a
single slot at the end, choosing that one on the strength of its record was right. Reproduce by
sending Adversarial at real code with a real harness, and not at prose.

### Building the runner ended a loop that reviewing could not

Eight findings across three passes were defects in prose describing a runner nothing implemented —
F17, F18, F19, F30, F34, F40, F44, F46. Each pass found another because there was nothing to test.
The clearest case: a fix cited `ps -o sid=`, which does not exist on macOS, and it survived three
passes of review because reading cannot catch that. Once the runner existed, its own implementer
disproved the replacement guess too (`ps -o sess=` always prints `0`) and found a real bug the plan
would have shipped (bare `cd && pwd` does not resolve symlinks, and macOS `$TMPDIR` is a symlink).
**When a review keeps finding gaps in the same unimplemented rules, build the thing.**

### Deviations were recorded rather than smoothed over

The panel record carries the stale slots, the four open findings with reasons, the operator's
withdrawal of F12 with its reason, the substitution of the Security slot, the over-cap diff decision,
and the correction of a fix agent's near-false lint claim. `FINISHED` was written only against a
`COMPLETE:` verdict whose clauses were relayed verbatim.

---

## 4. What could be automated or moved to a script

- **Ledger appends and token accumulation** — `scripts/append-ledger-line.sh`, called at each
  dispatch and each report. Makes the record a side effect rather than something to remember, and
  survives a session that ends mid-run. **KAN-155.**
- **Torn-prose heuristics** — flag a paragraph whose last line lacks terminal punctuation, and a
  `> ` line directly after a non-quoted line. Bounded, mechanical, and both patterns occurred here.
  **KAN-157.**
- **Reproducer resolvability at read time** — re-check that a recorded `finding-reproducer:` command
  still resolves when the record is read, not only when supplied. Not filed: speculative, and
  `run-reproducer.sh` already refuses a non-existent path at the moment it runs.
- **A `timeout` shim** — this run's own verification loop failed all 19 harnesses at once because
  `timeout` is absent on Darwin, which is the very fact the change documents. A tiny wrapper would
  stop that recurring.

---

## Findings filed and declined

| Finding | Outcome |
|---------|---------|
| The ~150-line escalation trigger is now the dominant path to a Full re-run | **filed — KAN-154** |
| Append the ledger per dispatch; accumulate per-dispatch token usage | **filed — KAN-155** |
| An unanswered optional-slot prompt should not widen a `light` panel | **filed — KAN-156** |
| A prose-moving edit must be verified by reading its destination | **filed — KAN-157** |
| `/myflow-fast` silently records Sonnet for the implementer role | declined |
| Re-check a reproducer's resolvability when the record is read | declined |

Outstanding work from the change itself — F25, F43, F46 and F54's residual gap — was filed
separately at the unfinished-work gate as **KAN-153**.
