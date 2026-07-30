# Widen the plan-provenance guard's scan scope beyond tasks.md

## Why

`scripts/check-plan-provenance.py` scans `openspec/changes/*/tasks.md` and nothing else. A change's
`design.md` and `proposal.md` state numeric claims and carry code blocks like any plan, and nothing
checks them.

KAN-20 was filed on the premise that this narrow scope let several defects survive review inside
KAN-14's own artifacts. Measurement against `c1f5aa6` only partly supports that premise, and saying
so plainly is what keeps this change's scope honest:

- The defect with the most instances — `measured:` tags citing a ref where the cited scripts do not
  exist — had most of its instances **inside `tasks.md`**, already fully in scope, and the guard
  passed clean on them. The guard checks that provenance is *stated*, never that it is true, so
  scope was never the binding constraint there.
- Another cited defect carries a `measured:` tag whose arithmetic is wrong — again a truth problem,
  not a scope problem.
- A third now lives only in a Python module docstring, outside `openspec/changes/` entirely.

What widening genuinely buys is smaller and real: two unattributed numeric claims in archived
`design.md`/`proposal.md` files that the guard cannot currently see, and the closing of a gap where
a plan's most argued-over prose is the least checked.

Widening also exposes two false-positive classes that must be fixed in the same change, because
shipping them would invite exactly the suppression-marker response this guard's design exists to
avoid.

## What Changes

- The guard scans `design.md` and `proposal.md` alongside `tasks.md`, applying both the fenced-block
  rule and the numeric-claim rule to all three.
- A change's `specs/` directory stays out of scope. Spec text legislates rather than describes, and
  a spec must be able to quote the thing it governs.
- The numeric classifier stops reading `KAN-6 errors` as a claim of six errors — an issue key
  followed by a unit word is not a numeric claim.
- A number enclosed in a matched pair of double quotes — straight or curly — is not a claim. This is
  what lets the provenance documentation quote the historical invented baseline it exists to
  explain, without demanding a tag that could not truthfully be written. Double quotes are the
  **whole** delimiter set: inline code spans were a second class during implementation and were
  removed, because every measured false positive is quote-delimited and the code-span class failed
  open five times. Three vetoes keep the exemption fail-closed — a backslash-escaped delimiter, a `<`
  anywhere on the line, and a delimiter class that does not fully pair each withdraw it, and the
  guard now says which of the three fired and how to fix it. The `<` rule is deliberately the
  coarsest one available: the veto it replaced located raw-HTML constructs with a regex, that regex
  truncated a tag at the first `>` where §6.6 permits one inside a single-quoted attribute value, and
  the exemption failed open for a sixth time. Approximating a spec grammar caused all six, so the
  grammar is gone rather than improved; the measured cost across this repository is two exempted
  claims, neither of them in the guard's scan scope.
- The containment check that refuses a symlinked or escaping `tasks.md` applies to every scanned
  file, not just `tasks.md`.
- The contract, the spec, and the guard's own module docstring are corrected to describe the
  behaviour the code actually has.

Explicitly **not** in this change: verifying that a `measured:` tag's `@ <ref>` resolves, or that a
cited command exists there. That is the change that would have caught the defect above, and it
amends a deliberate, documented boundary — it deserves its own ticket rather than a silent
enlargement of this one.

## Capabilities

- `myflow-plan-provenance` — modified

## Impact

- `scripts/check-plan-provenance.py` — scan set, containment call sites, scan-integrity invariant,
  numeric classifier, module docstring
- `scripts/test-check-plan-provenance.sh` — new cases for the widened scope, the exemption, and the
  generalised containment
- `skills/myflow-contracts/plan-provenance.md` — the scope section, plus a new subsection for the
  exemption
- `openspec/specs/myflow-plan-provenance/spec.md` — via this change's delta spec

Once this ships, the guard scans this change's own `design.md` and `proposal.md`, and `/myflow-do`
runs the lint step at the end of its run. These artifacts are therefore written tag-clean under the
new rules, which doubles as the end-to-end proof that the widening works.
