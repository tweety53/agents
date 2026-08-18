# Review panel — kan-200-self-review-filing-ask-per-angle

Roster: `light` (Primary · Principles · Code review (low)) · panel model `sonnet`
Diff size: 831 lines, under cap (`check-panel-diff-size.sh` exit 0, no operator prompt fired)
Merge base: d5cd0a5 · Head: a2fabeb
Optional slots: none — no triggers fired

## Round 1

### Slot: Principles (sonnet) — 1 Minor
- Minor: design.md said "sixteen" declared pre-rule reports; the guard correctly declares seventeen (16 + the KAN-197 report this change recovers). FIXED by the parent in all four design.md copies — planning-path file, not an implementer's to edit.

### Slot: Code review, low effort (sonnet) — 3 Major, 4 Minor
- Major 1: `filed:` disposition accepts any non-empty string (`filed: yes`, `filed: 201`, `filed: KAN-` all pass). Spec requires a filed finding to NAME an issue key. check-self-review-report.sh:275-280.
- Major 2: `find` failure inside a process substitution is unchecked; an unreadable target folds into "empty corpus" and exits 1 instead of 2, contradicting the guard's own header (lines 76-84). check-self-review-report.sh:216.
- Major 3: a `###` heading does not reset the awk scanner's `cur`, so its body is misattributed to the previous `##` section — a misplaced none-marker then SUPPRESSES that section's own empty violation. check-self-review-report.sh:190-201.
- Minor 4: a finding line before any `##` heading is silently dropped (not counted, not flagged). :201-205
- Minor 5: two `##` sections carrying the same angle label are silently merged; no duplicate-section violation exists. :248-256
- Minor 6: `is_declared_basename()` short-circuits content checking in explicit-directory mode, where no matching `coverage_declare` was made. :223 vs :162-164
- Minor 7: CRLF line endings are never stripped, producing misleading violations rather than a targeted message.
- Harness weakness tied to Major 1: case 5 only tests a fully empty key, never a non-empty-but-invalid one.

## Fix round 1 — all ten findings repaired

Folded into their original task commits via `git commit --fixup` + `git rebase -i --autosquash`.
Rebased: ee56284 (task 1, unchanged) -> fe432eb (task 2, F1-F9) -> 2e90913 (task 3, F10 contract
half) -> 38a1d28 (task 4, unchanged) -> cb650a5 (task 5, unchanged).

- F1: `ISSUE_KEY_RE='^[A-Z][A-Z0-9]*-[0-9]+$'`; harness case added for a non-empty-but-invalid key
  (nine cases -> ten; old 6-9 renumbered 7-10).
- F2: `find` writes to a temp file whose exit status is checked directly, replacing the process
  substitution that discarded it. Untraversable target now exits 2, not 1 — the guard's header
  claim is now true rather than aspirational.
- F3: second awk branch `/^#+[[:space:]]/` resets `cur` for any heading level.
- F4: `LAST_IDX` per report; a heading whose index is below the max seen is named out of order.
- F5: a second heading for a label already seen is a named duplicate, citing the first line.
- F6: awk emits ORPHAN records for a finding-shaped line while `cur==""`.
- F7: `HAS_NONE && FINDING_COUNT>0` is a named violation.
- F8: the declared-basename fast path is gated on a `BARE_INVOCATION` flag, so it applies only in
  the mode where the declarations were actually registered.
- F9: a trailing `\r` is stripped during scanning — chosen over failing, since a CRLF checkout is an
  environment artifact rather than a content defect.
- F10: `operator-prompts.md` gains a multi-select variant section (909 -> 1349 bytes, budget row
  raised 1136 -> 1687); `finish-contract.md` gains the citation it was missing.
  `skills/myflow-finish/SKILL.md` needed no edit — it already cited the contract correctly, so no
  fixup was made against cb650a5.

**Latent bug the fix round found on its own, which no slot had raised:** bash's `read` with
`IFS=$'\t'` collapses *consecutive* tabs, because tab is whitespace-class IFS — an empty label field
vanished and shifted every later column left by one. Found while implementing F6; fixed by giving
ORPHAN a non-empty placeholder label.

## Round 2 — FULL escalation

Escalated automatically, not by choice: **the fix altered a guard's behaviour**, which is its own
clause in the escalation ladder. Every required slot re-runs against the rewritten
`final-review.diff` (1112 lines, `check-panel-diff-size.sh` reports 976, under cap, exit 0). No
conditional slot's trigger fired, so none is re-run.

### Round 2 results

**Round 2 — 4 Major, 3 Minor (two pairs converged on one root cause each).**

- Primary: Major (harness has no regression coverage for 8 of the 9 fixes) + Minor (a body line
  beginning with a literal tab is swallowed by the record protocol and silently accepted).
- Principles: Major (a malformed line inside a section is silently dropped) + Minor
  (`BARE_INVOCATION` is redundant with `$#`, which no code path mutates) + Minor
  (`operator-prompts.md`'s new multi-select section asserts "None is the default" as universal,
  which `skills/myflow-do/SKILL.md`'s existing optional-slot multi-select contradicts — its safe
  default is "all of them").
- Code review (low): Major (harness coverage, **proven by mutation** — each of the eight fixes
  reverted individually in a scratch tree, harness reported "all cases passed" all eight times) +
  Major (the ORPHAN shape pattern allows `[[:space:]]*` where `FINDING_LINE_RE` requires exactly
  one, so the same malformed finding-shaped line is caught before a heading and invisible after
  one) + Major (a trailing tab on a `filed: <KEY>` line is stripped by `IFS=$'\t' read` before
  validation, so a malformed disposition passes).

**Principles' Major and code review's second Major are the same defect**, and code review's version
carries the correct root cause and scope: the gap is finding-**shaped** lines, not prose. A report is
prose plus finding lines, so flagging every unrecognized body line would reject blank lines and
ordinary paragraphs. Fix round 2 is scoped to the shape predicate, not to all prose — recorded here
because implementing Principles' finding as literally worded would have broken every real report.

**Primary's Minor and code review's third Major are also one defect** — the `IFS=$'\t' read` record
protocol loses tab characters at both field boundaries. F6 patched one manifestation with a
placeholder label; the protocol itself is what needs repairing.

## Fix round 2

All five repaired, folded into the original task commits. Rebased: ee56284 (task 1) -> c91ec9e
(task 2, G1-G4) -> 478c61f (task 3, G5) -> 92a222c (task 4) -> 17fed87 (task 5).

- G1: harness cases 11-18, one per uncovered behaviour, **each mutation-proved** — the fix reverted
  in a scratch tree, the harness confirmed to fail, the fix restored. Eight for eight.
- G2: `FINDING_SHAPE_RE` defined once as the loose predicate and used by both the orphan check and a
  new in-section branch, so the two patterns can no longer disagree. Passing it into awk needed
  `${FINDING_SHAPE_RE//\\/\\\\}` — awk's `-v` silently strips backslashes from escapes it does not
  recognise, the same class of bug the script's own header already documents for bash's `[[ =~ ]]`.
- G3: the record protocol moved from `IFS=$'\t'` to ASCII Unit Separator (`\037`), which is not
  IFS-whitespace, so `read` neither collapses runs of it nor trims it at a record edge. Round 1's
  placeholder-label workaround was removed — the protocol now carries an empty field correctly.
- G4: `BARE_INVOCATION` removed; both call sites read `[[ $# -eq 0 ]]` directly.
- G5: the multi-select section now defines the shape and requires the call site to name its own
  default, instead of asserting one polarity as universal; it notes the two existing call sites
  choose oppositely. 1349 -> 1932 bytes, budget row 1687 -> 2415.

## Round 3 — FULL escalation

Same clause as round 2: the fix altered a guard's behaviour. All three required slots re-run against
the rewritten `final-review.diff` (1449 lines; `check-panel-diff-size.sh` reports 1313, under cap,
exit 0). Commit-fields guard clean on all four rebased task commits.

### Round 3 results

- Primary: **clean, no findings.** Independently mutation-re-proved 3 of G1's 8 cases (F4, F6, F9 —
  each revert failed exactly the expected case), and verified `FINDING_SHAPE_RE` matches identically
  in bash and awk across six boundary strings, so the two call sites genuinely share one predicate.
- Principles: **1 Major.** G3 — the Unit Separator protocol, round 2's largest structural change —
  has no regression coverage. Reverting the protocol to `IFS=$'\t'` leaves all 18 cases green: the
  harness contains no literal tab byte and no `\037` byte anywhere. Verified independently by the
  parent: `grep -cP '\t'` and `grep -c $'\037'` over the harness both return 0. Fix round 2
  mutation-proved the eight cases it added and did not apply that standard to its own protocol swap.
  Plus an architectural **observation, explicitly not filed as a defect**: G2 and G3 are two taxes on
  the same awk->bash serialize/reparse boundary, and collapsing that boundary would delete the bug
  class rather than patch each manifestation. Recorded for self-review rather than actioned here.
- Code review (low): **1 Minor.** The protocol's own header claims the `rest` field survives
  "byte for byte, whatever it contains". False for a literal `\037` in report content: `read` still
  splits on it, so the byte is silently swallowed and a line differing from the none-marker only by a
  trailing `\037` is accepted as compliant. Practical risk is low — nobody types a Unit Separator
  into prose — but it is the same failure class G3 exists to close, and the guarantee is stated with
  no caveat. Independently re-verified rounds 1-2 reproducers (all still caught) and mutation-proved
  F2, F3, F9; honestly reported checking 3 of 8 rather than the 4 asked. Combined with Primary's F4,
  F6, F9, five of the eight cases have now been independently re-proved by a slot that did not write
  them.

## Fix round 3

Both repaired. Rebased: ee56284 (task 1) -> 2d1d141 (task 2, H1+H2) -> 4f329b0 (task 3) ->
abe0ee1 (task 4) -> 62af4c2 (task 5). Commit-fields guard clean on all four task commits.

- H1: harness cases 19 and 20, carrying **real** control bytes (ANSI-C quoting, matching case 18's
  own `\r` technique) and asserting the byte's presence before invoking the guard, so a case cannot
  silently degrade into a no-op.
- H2: chose to make the claim true rather than caveat it, per the repository's stated posture. A new
  `CTRLBYTE` record type checks every line for a raw 0x1F **before** any other classification and
  carries no content field — a raw 0x1F inside one would itself corrupt the US-delimited framing.
  Case 21 proves it. The header's "byte for byte, whatever it contains" claim now states the
  exclusion instead of being false.

**The fix round corrected the parent's own dispatch.** The combined revert the parent specified for
the mutation test would have produced a **false pass**: mutating `US` changes the variable the
protocol and the new control-byte check share, so case 20 passes by cross-contamination rather than
by the protocol working. It split into two surgical mutations instead — protocol-only, and
CTRLBYTE-only — and each new case failed exactly as required, with the other cases staying green.
A mutation test that is not itself isolated proves nothing, and this one nearly was not.

## Round 4 — FULL escalation

Third trigger, and now two apply at once: the fix altered a guard's behaviour, and three fix rounds
have run. All three required slots re-run against the rewritten `final-review.diff` (1567 lines;
`check-panel-diff-size.sh` reports 1431, under cap, exit 0).

### Round 4 results
- Principles: **clean — ready to hand off.** Ran both isolated mutations: each new case fails under
  its own mutation and stays green under the sibling. Also ran the parent's originally-specified
  *combined* revert and confirmed case 20 passes in full — RC and message substring both — purely
  through cross-contamination, empirically proving the split was necessary and the parent's dispatch
  was unsound as written. Confirmed H1/H2 stayed additive: still exactly one `awk` invocation, no
  restructuring, diff growth accounted for by three harness cases plus an 11-line check.
  Architectural observation restated and explicitly not a blocker; carried to self-review.
- Primary: **clean — recommends integration.** CTRLBYTE ordering verified unfoolable across six
  placements of a raw 0x1F (heading, none-marker, empty line, only byte, finding line, multiple).
  43/43 assertions, all guards green, plan alignment intact across 10 files.

  **It also corrected the parent's verification method.** The parent had grepped the harness source
  for raw `\t` / `\037` bytes and read 0 as corroboration that coverage was missing. That grep proves
  nothing in either direction: the cases use bash ANSI-C quoting (`$'\t'`), evaluated at runtime, so
  the raw byte never exists in the `.sh` source and the grep reads 0 whether the harness works or
  not. The round-3 finding was sound — it was established by mutation, not by that grep. Primary
  verified the bytes correctly instead: extract each case's fixture-building lines, run them, `xxd`
  the fixture. A genuine 0x09 and 0x1F are present, and each case asserts its own byte before
  invoking the guard. Recorded because the flawed check is the kind a later round would copy.
- Code review (low): **clean.** Probed CTRLBYTE across six adversarial placements. Investigated one
  genuine blind spot and declined to file it, correctly: a 0x1F inside a `##` line skips the
  heading-reset, so downstream content is misattributed to the previous section — but CTRLBYTE
  guarantees `VIOLATIONS>0` whenever any 0x1F exists and the corrupted section still surfaces as
  missing, so no violation is ever lost, only a diagnostic misattributed. Within the stated
  allowance. Independently reproduced fix round 3's cross-contamination claim and confirmed
  CTRLBYTE-only reversion yields a genuine false clean pass for case 21.

## Panel outcome — CLEAN after 4 rounds

Zero open findings at any severity. 17 findings repaired across 3 fix rounds; findings per round
10 -> 7 -> 2 -> 0. 21 harness cases, 43 assertions, twelve of them added by fix rounds and each
mutation-proved.

Per-slot totals across the four rounds — recorded because KAN-198 wants exactly this data, and
because a finding count alone would miss what happened here:

| Slot | R1 | R2 | R3 | R4 | Total |
|------|----|----|----|----|-------|
| Primary | 2 Major, 3 Minor | 1 Major, 1 Minor | clean | clean | 3 Major, 4 Minor |
| Principles | 1 Minor | 1 Major, 2 Minor | 1 Major | clean | 2 Major, 3 Minor |
| Code review (low) | 3 Major, 4 Minor | 3 Major | 1 Minor | clean | 6 Major, 5 Minor |

The low-effort slot led on volume for the third change running. But the single most consequential
finding of this change came from **Principles in round 3** — that fix round 2 had mutation-proved
its eight new cases and never its own protocol swap. Neither heavier-reading slot asked whether the
fix round had held itself to its own standard, and no finding-count metric would show that the
question, not the count, was the valuable part.

## Findings table

Identifiers are assigned here, once, across all four rounds. Round 1's ten dispatched repairs are
F1-F10; F11 is the Principles Minor the parent fixed directly (a planning-path file, not an
implementer's to edit). Round 2's G1-G5 are F12-F16; round 3's H1-H2 are F17-F18.

| Finding | Round | Slot | Severity | What |
|---|---|---|---|---|
| F1 | 1 | Code review | Major | `filed:` accepted any non-empty string as an issue key |
| F2 | 1 | Code review / Primary | Major | `find` failure folded into "empty corpus"; exit 1 where 2 was owed, contradicting the guard's own header |
| F3 | 1 | Code review | Major | `###` did not reset the scanner; a misplaced none-marker suppressed a real violation |
| F4 | 1 | Primary | Major | section order not enforced, though the spec requires it |
| F5 | 1 | Primary / Code review | Major | two sections for one angle silently merged |
| F6 | 1 | Code review | Minor | a finding line before any heading silently dropped |
| F7 | 1 | Primary | Minor | a section carrying both a none-marker and finding lines passed |
| F8 | 1 | Code review | Minor | `is_declared_basename` short-circuited content checks in explicit-directory mode |
| F9 | 1 | Code review | Minor | CRLF line endings produced misleading violations |
| F10 | 1 | Primary | Minor | multi-select cited `operator-prompts.md` for a shape it did not define |
| F11 | 1 | Principles | Minor | design.md said sixteen declared pre-rule reports; seventeen are declared |
| F12 | 2 | Primary / Code review | Major | harness could not detect the reversion of 8 of the 9 round-1 fixes |
| F13 | 2 | Principles / Code review | Major | a malformed finding-shaped line inside a section silently dropped |
| F14 | 2 | Primary / Code review | Major | the `IFS=$'\t'` record protocol lost tabs at both field boundaries |
| F15 | 2 | Principles | Minor | `BARE_INVOCATION` redundant with `$#` |
| F16 | 2 | Principles | Minor | the multi-select section asserted a default polarity that is not universal |
| F17 | 3 | Principles | Major | the protocol swap itself had no regression coverage |
| F18 | 3 | Code review | Minor | the protocol's byte-for-byte guarantee was false for a literal `\037` |

finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 fixed
finding-status: F9 fixed
finding-status: F10 fixed
finding-status: F11 fixed
finding-status: F12 fixed
finding-status: F13 fixed
finding-status: F14 fixed
finding-status: F15 fixed
finding-status: F16 fixed
finding-status: F17 fixed
finding-status: F18 fixed
findings-total: 18
