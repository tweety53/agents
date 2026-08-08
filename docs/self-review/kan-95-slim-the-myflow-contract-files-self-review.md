# Self-review — kan-95-slim-the-myflow-contract-files

**Rating: 3/5** (operator)

**Outcome:** a `/myflow-do` run loaded 210,481 bytes of contract and skill prose before dispatching
its first implementer; it now loads 166,534 — a 20.88% cut. The proposal projected ~128,500, so the
projection was missed by 26%.

**Cost:** ~40 subagent dispatches — 10 implementers, ~12 task reviews and re-reviews, 21 panel
slot-runs across three full passes and one targeted pass, and 5 fix waves. 37 findings, zero left
open.

---

## 1. Problems, and the pipeline change that would avoid each

### Fixes introduced roughly 9 of the 37 findings

Every fix dispatch named a finding's location, and every fixer fixed that location. Nothing required
them to ask what *else* described the same thing.

| Fix | What it left behind |
|-----|---------------------|
| F5's fix restored a rule to `pipeline.md` | left the verbatim original in `README.md`, uncited — one copy became two (F22) |
| F7's fix reworded the two sentences the finding quoted | the four sites that actually read the map were untouched, so the vacuous pass survived (F13) |
| F13's fix corrected the execution layer | the description layer in `README.md`, `CLAUDE.md`, `AGENTS.md` and both command files still said the old thing (A2, F23) |
| Round 4's fix for a silent skip | turned it into a guaranteed hard block on `/myflow-do`'s first run (F32, **Critical**) |

Round 4 carried an ad-hoc sweep instruction and immediately found two further copies nobody had
flagged. It worked the first time it was tried. **Filed as KAN-103.**

### The reference guard cannot see the two classes that produced both Criticals

`scripts/check-references.sh` resolves citations against the agents repo — which is not the
resolution a run performs. Two distinct failures follow:

- **A citation from an installed file to a path `setup.sh` never installs.** `CLAUDE.md` and
  `AGENTS.md` are *copied into every target project* and `CLAUDE.md` is auto-loaded in every session
  there; they carried eight citations resolving into the project's own tracked, PR-editable
  `README.md` (F12). `finish-contract.md` and `myflow-finish/SKILL.md` cited `openspec/` as
  *canonical for the procedure* of the step that runs immediately after run 2's destructive cleanup
  (S1). Both Critical, both introduced by this change. **Filed as KAN-102.**
- **A citation that resolves but reaches wrong or absent content.** The appendix told a future editor
  the pipeline had a "five-command surface" when it had four (F10) — after Task 1 had fixed that
  exact phrase elsewhere and declared it done. Partly covered by the open KAN-84.

### The controller's own delta-spec edits are reviewed by nothing

Implementers are barred from `openspec/`, so the controller hand-edits the deltas — and nothing
reviews them. Four defects came from that: a spec amendment whose test was **vacuous** (F17, true of
every conceivable split), a capability with no delta at all whose live `SHALL` the change
contradicted (F20), a delta carrying vocabulary the change had eliminated everywhere else (F36), and
a design record contradicting what shipped, twice (F18, F37).

A fifth surfaced during run 2's own spec sync: `myflow-planning-effort` — a **ninth** capability the
deltas missed — justified two live rules by reference to state self-heal, which this change deleted.
Corrected in the sync step and disclosed in that commit. **Filed as KAN-104.**

### Two known defects fired live

- **KAN-77.** `preserve-session-records.sh` looks for `.superpowers/sdd/tasks/progress.md`; this run
  wrote the flat `.superpowers/sdd/progress.md` that `myflow-do`'s own skill documents. The script
  reported `skipped … (absent)` and run 2 would have destroyed the SDD ledger with the worktree.
  Caught only by reading that line. The ledger and all eleven per-move ledgers were preserved by hand.
- **KAN-56.** `/myflow-start` stages planning artifacts in the main checkout; run 1's merge then
  failed against thirteen of them. All were verified byte-identical to the branch's committed copies
  before being discarded.

### Five parts in one change

Raised at planning with a decomposition option offered; the operator chose one change. Both
Criticals were consequences of part 2 (removing `/myflow-info`) interacting with parts nobody was
looking at when it landed.

## 2. Cost, and what would reduce it without losing quality

The dominant reducible cost is the escalation rule. `three or more fix rounds have already run` is
**monotonic** — permanently true once reached — so passes 3 and 4 both escalated on a counter rather
than on anything about the fixes in front of them.

| Pass | Roster | Findings |
|------|--------|----------|
| 1 | 7 slots (conditional selection) | 11 |
| 2 | 7 slots (escalated — deltas altered) | 11 |
| 3 | 7 slots (escalated — "3+ rounds") | 8 |
| 4 | 4 slots (targeted, operator decision) | 5 |

Pass 4 found more per slot than pass 3 and caught the Critical round 4 had introduced. Roughly one
full 7-slot pass was spent on the counter alone. **Filed as KAN-105.**

## 3. What went well, and how to reproduce it

- **The per-move ledger requirement did its job**, and Task 9's implementer invented something
  stronger than specified: reverse every core edit and confirm it reproduces the pre-task image
  byte-for-byte. Worth promoting into the contract.
- **Byte-verification by script rather than by eye** caught a rewrap in Task 7 that the implementer's
  own review had missed — words identical, line breaks not.
- **Instructing reviewers to verify negative claims** — "the fixer reports no third instance exists;
  check that yourself" — repeatedly found real things. A clean negative from the party that already
  missed something is worth re-deriving.
- **Roster breadth paid off exactly once, on the most important finding.** Security and Principles
  reached F12 by different routes, and that convergence is what raised it to Critical.
- **Adversarial's posture instruction** — assume it is broken; guards green is the condition under
  which the remaining defects are the invisible ones — produced the best findings of passes 1 and 2.

## 4. What could be automated

- **KAN-102's guard** is the highest-value item here and is mechanically checkable: for every file
  `setup.sh` installs or copies, every named-section citation must resolve to a path `setup.sh` also
  ships. It would have caught F12, F14 and S1 without a human reading anything.
- **`preserve-session-records.sh` should glob** `.superpowers/sdd/**/progress.md`, or `myflow-do`
  should state the workspace path it actually uses. The two disagree today, and the disagreement
  silently loses the ledger (KAN-77).
- **The sweep in KAN-103 is prose, not automation** — but it belongs in the fix-dispatch template
  the way the four required implementer blocks already do.

---

## Findings filed

| Issue | Priority | What |
|-------|----------|------|
| **KAN-102** | High | Guard: a citation in an installed file must resolve to an installed path |
| **KAN-103** | High | Every panel fix dispatch must require a corpus sweep |
| **KAN-104** | High | The controller's own delta-spec edits are reviewed by nothing |
| **KAN-105** | Medium | Escalation trigger "three or more fix rounds have run" is permanently true |

**KAN-101** (Highest) was filed earlier in the run, during the panel: bare `scripts/*.sh` citations in
installed files resolve into the target project, so a pull request adding
`scripts/check-unfinished-work.sh` gets it **executed** by `/myflow-finish` run 1. Pre-existing, not
introduced by this change, and strictly worse than anything this change did introduce.

**Not re-filed:** KAN-77 and KAN-56 already exist and both fired live in this run — the evidence above
belongs on those issues rather than on new ones.
