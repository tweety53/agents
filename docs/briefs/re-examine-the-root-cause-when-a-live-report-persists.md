# Re-examine the root cause when a live report persists

When a live report — the operator's screenshot, a fresh measurement, a
reproduction on the real thing — still shows the defect after a fix has been
written, re-examine the root cause from the new evidence before extending the
first patch. The persistence is data about the cause, not about the patch's
width: the second look asks what the report actually shows now, and that
question has room to delete the first fix's approach entirely, not only to
add to it.

## Why

kan-527's post-handoff F22 fixed an overflowing in-plot caption by measuring
the widest gridline label's rendered width and flooring the caption past it.
The operator's live screenshot then showed the fix incomplete: a third text —
the "NOT COUNTED" band label — shared the same row and had never been
accounted for. The superseding fix did not add a third width measurement; it
re-derived the root cause from the new evidence and found the caption was a
straight duplicate of the header already above the chart — the pattern Weight
had already dropped — so the caption was deleted outright, its day count
folded into the headline, and the first fix's plumbing removed as moot. An
existing header test already covered the folded behaviour. The second look
deleted code where the first had threaded a measurement through it, because
it asked a different question: not "what did the fix miss", but "why does
this caption exist at all".

## How to run one

- Treat a persisting live report as new evidence about the cause, never as
  "the patch needs one more case" — the first patch's framing is itself a
  suspect.
- Re-derive the root cause from what the report now shows, against the
  untouched tree, before touching anything.
- Expect deletion: a root cause found under the first patch's premise often
  dissolves the premise, and the first fix's plumbing goes with it.
- Before writing the new fix, check what already covers the behaviour — an
  existing test can settle it without a new one.

Recorded from KAN-694, filed from the deferred self-review of kan-527
(`kan-527-group-workouts-body-fixes-2`), whose post-handoff F22 is where the
pattern was named.
