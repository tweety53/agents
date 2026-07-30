# Widen the plan-provenance guard's scan scope beyond tasks.md

**Jira:** KAN-20 · **Change:** `kan-20-widen-plan-provenance-guard-scan-scope` · **Date:** 2026-07-30
**Effort:** high · **Measured against:** `c1f5aa6` (main, agents)

## The problem, as filed

`scripts/check-plan-provenance.py` scans `openspec/changes/*/tasks.md` and nothing else. KAN-20 asks
for that scope to widen, citing four defects that survived many review passes inside KAN-14's own
artifacts — files the guard does not read.

## What the measurements actually showed

Three findings from exploration changed the shape of this change. All were taken by importing the
guard module and calling its own `ProvenanceGuard.check_file`, so they are the guard's real
classifier rather than a proxy for it.

### 1. The ticket's cost estimate was an order of magnitude high

KAN-20 projected the number rule firing on 5–71 lines per `design.md`, using a bare-number proxy
(`grep -cE '[^a-zA-Z0-9_.-][0-9]{1,}'`). The real `CLAIM_RE` fires only on a digit group followed by
one of eight unit words (`tests|failures|errors|files|lines|seconds|minutes|ms`), so the true
figures are far smaller:

| Candidate scope | Real violations | Classification aborts (exit 4) |
|---|---|---|
| `design.md` × 4 | 3 | 0 |
| `proposal.md` × 4 | 3 | 0 |
| change `specs/**` × 21 | 1 | 0 |

<!-- measured: a throwaway script (machine-local scratchpad, not preserved) that imports
     scripts/check-plan-provenance.py via importlib and calls ProvenanceGuard.check_file on each
     openspec/changes/archive/*/{design,proposal}.md and specs/**/*.md @ c1f5aa6. Re-runnable by
     rewriting those ~20 lines against the same commit; the numbers come from the guard's own
     classifier, not a proxy. -->

Zero aborts means the container-prefix grammar handles all three file types cleanly; the mechanical
risk of widening is not where the ticket expected it.

### 2. Most of those hits are false positives, in two classes

- **Self-referential quotation.** `"194 tests"` — the historical invented baseline the whole
  contract exists to prevent — is flagged in KAN-14's `design.md`, `proposal.md` *and* `spec.md`,
  because the documentation must quote the bad number to explain the rule. No tag is honest for it:
  `measured:` would assert a run that never happened, `predicted:` is nonsense.
- **Issue-key adjacency.** `catches all six KAN-6 errors` is flagged because the classifier reads
  `6 errors` out of `KAN-6 errors`. The module docstring claims the rule deliberately excludes
  `KAN-14` and `IU-262` — true only because those keys are not followed by a unit word. The code
  does not have the property the docstring describes.

Genuine catches: `~85 lines instead of ~385` (kan-10 `design.md`) and `25 seconds later`
(kan-14 `proposal.md`).

### 3. Widening the scope would not have caught the defects the ticket cites

Of KAN-20's four motivating defects:

- *"NINE" vs ten* now lives only in `check-plan-provenance.py`'s docstring — outside
  `openspec/changes/` entirely, so no widening within a change directory reaches it.
- *the itemisation summing to 9 against a headline of 12* carries a `measured:` tag. The guard
  checks tag presence, never tag truth.
- *six `measured:` tags citing `@ d38372a` where the scripts do not exist* — five of the six sat in
  `tasks.md`, already fully in scope, and passed clean. The guard does not parse the `@ <ref>` half,
  and says so.
- *the "pass-7 ambient-tracker" misattribution* is prose, not a numeric claim.

The third is decisive: that defect had five instances inside the one file the guard already read.
**Scope was not the binding constraint — the tag-presence-versus-tag-truth boundary was.** This
change is therefore scoped honestly: it widens the net, catches the two real unattributed numbers
the widening does reach, and leaves the truth boundary alone.

## Decisions

### Widen the scope and fix the false-positive classes, rather than re-aiming at tag truth

**ID:** `scope-not-truth`
**Status:** active
**Chosen:** Stay close to the ticket — widen to `design.md`/`proposal.md` and fix the two
false-positive classes first. Honest about catching two real defects.
**Considered:** *Block rule only* — zero cost but catches neither real number defect, near a no-op
on this repo's evidence. *Re-aim at verifying the `@ <ref>` half of `measured:` tags* — this is what
would have caught the `d38372a` defect, but it amends "the guard never checks truth", a deliberate
and documented boundary. *Both* — largest scope, and wants decomposing into two changes rather than
one.

### `specs/` stays out of scope

**ID:** `specs-excluded`
**Status:** active
**Chosen:** Exclude a change's `specs/` directory entirely. Spec text legislates rather than
describes, and the only measured hit across 21 files is the quotation false positive.
**Considered:** *Include it* — raises why `openspec/specs/`, the synced copy of the same text, stays
unscanned. *Include for the block rule only* — invents a per-file rule split to save nothing
measurable.

### The quoted-number false positive is fixed syntactically, not with a fifth tag

**ID:** `quotation-syntactic`
**Status:** active
**Chosen:** A number enclosed in a delimiter pair on its own line is a quotation, not a claim.
Mechanical, no vocabulary change, and an automatic exemption never tempts anyone toward a
suppression marker.
**Considered:** *A fifth tag `quoted:<source>`* — auditable and nothing escapes silently, but it
amends "four tags exist, and no others" in both the contract and the spec to cover three instances
of a single self-referential quotation. *A hybrid of both* — strictly more surface to specify and
test for a case that has not arisen.

## The design

### Scan scope

Three files per non-archived change directory:

| File | Block rule | Number rule |
|---|---|---|
| `tasks.md` | yes (unchanged) | yes (unchanged) |
| `design.md` | new | new |
| `proposal.md` | new | new |
| `specs/**` | no | no |
| anything under `archive/` | no | no |

Both rules widen together. The asymmetry the contract establishes is between *how a block is
tagged* and *whether a number may appear at all*; it was never an asymmetry about which files each
rule reads, and splitting scope per rule would invent a second, harder-to-explain distinction.

The archive exclusion needs no work: it is applied at the change-directory level
(`if e == "archive": continue`), not per file, so every file type added to a change's scan set
inherits it.

Three mechanical consequences in `main()`'s scan loop:

1. **Containment generalises.** The containment gate was per-file and named for `tasks.md`; it
   becomes `verify_scanned_file_containment`, taking the filename as a parameter. It must run
   against each of the three candidates, or a PR-authored `design.md` symlink walks past
   the check that exists to stop exactly that. Exit 3's behaviour is unchanged: refuse that one
   candidate, keep scanning, still outrank every other exit code.
2. **The scan-integrity failure is restated.** "Change directories exist but none produced a
   `tasks.md`" becomes "none produced any of the three", so a change mid-planning does not trip a
   failure that means nothing.
3. **The `scanned` counter and success line** stop being about `tasks.md` specifically.

Exit codes, their ranking, and the exit-4 stop-scanning-this-file behaviour are untouched.

### Classifier changes

**Issue-key adjacency** — a bug fix. `CLAIM_RE` gains one fixed-width negative lookbehind
`(?<![A-Z]-)`, so `KAN-6 errors` no longer parses as `6 errors`. The docstring sentence describing
the `KAN-14`/`IU-262` exclusion is corrected to describe what the code now actually does.

**Quotation exemption** — new behaviour. A `CLAIM_RE` match falling inside a delimiter pair on its
own line is a quotation:

- **Delimiters:** paired straight double quotes (`U+0022`), paired curly double quotes
  (`U+201C`/`U+201D`), and CommonMark inline code spans (matched backtick runs). Curly quotes have
  zero occurrences in the corpus today, but a curly-quoted number is the identical false positive
  and covering it costs one character class rather than a later fix.
  **Single quotes and apostrophes are deliberately not delimiters.** Prose apostrophes
  (`the guard's scope`, `KAN-14's own artifacts`) are unpaired by nature, so admitting `'` would
  make pairing meaningless across most lines in this repository's documentation.
- **Line-scoped only.** A pair spanning a line break creates no region, keeping the exemption a
  line-local predicate rather than a second piece of parser state that must agree with the fence
  tracker.
- **Fail closed.** An unpaired delimiter yields no region, so the number stays a violation. The
  exemption can only remove a hit, and only where the number is demonstrably enclosed.
- **No fence interaction.** Inside an open fence the scan loop tests only for the closing line and
  never consults `CLAIM_RE`, so in-fence numbers are already never claim-checked.

Measured effect of both fixes on the widened scope: 6 hits fall to 2, and both survivors are the
genuine unattributed numbers.

<!-- measured: a throwaway script (machine-local scratchpad, not preserved) applying the
     (?<![A-Z]-) lookbehind and the quotation predicate to the guard's own CLAIM_RE over
     openspec/changes/archive/*/{design,proposal}.md @ c1f5aa6. Task 1 of the plan re-establishes
     this as a real test case, at which point the harness is the reproducible record. -->

**Accepted cost:** a genuine claim written inside quotes — `Baseline: "197 tests"` — escapes the
guard. Weighted low because the failure this contract exists to stop was an *asserted* baseline, not
a quoted one.

**Correction, recorded during implementation.** The prototype behind the figure above implemented
the predicate by counting delimiter *parity* across the line. That is unsound: a stray unpaired
delimiter plus an unrelated later pair manufactures a bogus region and exempts a bare asserted
number, which fails **open** against this design's own fail-closed requirement. The measurement over
the archived corpus was still real — that corpus contains no such line — but a friendly-corpus run
was never proof of correctness, and the block carrying it was tagged `verified:` as though it were.
The implementation replaces parity with nearest-enclosing-pair matching. The entry is left here
rather than edited away because this design is *about* the difference between a check performed and
a check claimed.

### Contract, spec, and tests

Documentation amendments:

1. `skills/myflow-contracts/plan-provenance.md` — "The guard's scope, and why it is narrow" is
   rewritten for the three-file set. The over-firing rationale stays but is re-pointed: it now
   justifies excluding `specs/` and the quotation exemption, rather than excluding `design.md` and
   `proposal.md`. A new subsection documents the exemption and its fail-closed rule.
2. `openspec/specs/myflow-plan-provenance/spec.md` — the requirement *"Provenance is enforced
   mechanically, and only where it applies"* has its scope `SHALL` amended to the three-file set,
   and the scenario *"Documentation outside a plan is not scanned"* reworded (skills, contracts and
   READMEs stay out; the justification changes). Three scenarios are added: a quoted number is not
   flagged, an untagged number in `design.md` is flagged, `specs/` is not scanned.
3. `scripts/check-plan-provenance.py`'s module docstring — the scope paragraph, and the numeric-rule
   paragraph that claims a property the code does not have.

Tests, written first, into `scripts/test-check-plan-provenance.sh` (currently passing):

- untagged block and untagged number in each of `design.md` and `proposal.md` → exit 1
- a quoted number in each of the three delimiter styles → exit 0
- an unpaired delimiter → exit 1 (fail closed)
- `KAN-6 errors` → exit 0, against a bare `6 errors` → exit 1
- an untagged number under `specs/` → exit 0 (not scanned)
- a `design.md` symlink → exit 3 (containment generalised)
- a change directory with `design.md` but no `tasks.md` → scanned, not a scan-integrity failure
- an archived change's `design.md` → not scanned

Verification uses the repo's declared commands: lint is `check-vocabulary.sh`,
`check-references.sh`, `check-plan-provenance.sh`; tests are the five `test-*.sh`. There is no
auto-fix command in this repository, so the Lint Fix Priority rule's auto-fix step is inapplicable
rather than skipped.

**Bootstrapping consequence.** Once implemented, the widened guard scans this change's own
`design.md` and `proposal.md`, and `/myflow-do` runs lint at the end of its run. Those artifacts
must therefore be written tag-clean under the new rules from the start. This is also the most direct
end-to-end proof that the widening works.

## Out of scope

- **Tag truth.** Verifying that `@ <ref>` resolves, or that a cited command exists there, is
  deliberately not in this change — see `scope-not-truth`. It is the change that would have caught
  the `d38372a` defect and deserves its own ticket.
- **Commit messages.** Outside any guard's reach, as KAN-20 itself records.
- **`openspec/specs/`.** The published, synced copies belong to no change in flight; excluding
  `specs/` from the change scope keeps this consistent by construction.
