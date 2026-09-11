## Context

The fix-dispatch contract (`skills/flow/review-panel.md`, the fix step) gives **one** fix subagent
the combined list of every surviving open finding, once per round — `check-panel-fix-single-dispatch.sh`
holds every run to exactly that shape (key `panel-fix-<round>`, one original + at most one
`-retry`). KAN-459's round 1 ran 24 findings through that one dispatch across two repos and read
~110M cache tokens; low-count rounds 3–5 held ~30M. The dispatch site is one of
`skills/flow/implement.md`'s four-row closed list, and KAN-482 (one dispatch per reviewer/finding)
is the abuse the single-dispatch rule and the guard exist to prevent — any relaxation must keep
that abuse caught.

## Decisions

### Cap findings per panel-fix dispatch at 10, chunked sequentially within a round

**ID:** chunk-panel-fix-dispatch
**Status:** active
**Chosen:** cap at most 10 findings per panel-fix dispatch — matches the measured ~30M-per-round
cache reads at low finding counts, bounds every round (single- or multi-repo), and needs no new
key vocabulary beyond a chunk suffix.
**Considered:** splitting fix dispatches by repo — helps only multi-repo changes (a single-repo
24-finding round still blows up), and worktree/repo names would leak into dispatch keys; keeping
one dispatch per round with no bound — the measured failure this change exists to fix.

### First chunk keeps the bare key; later chunks number from 2

**ID:** chunk-key-shape
**Status:** active
**Chosen:** `panel-fix-<round>` stays the key when one dispatch suffices (the common case —
today's canonical shape, every doc and the guard's existing cases unchanged); a chunked round adds
`panel-fix-<round>-2`, `-3`, … contiguous, each with the same one-retry rule
(`panel-fix-<round>[|-<n>]-retry`).
**Considered:** numbering every chunk from 1 — churns the single-dispatch shape that rounds 3–5
(the majority) already use and rewrites every existing reference for no mechanical gain.

### The guard bounds chunks from the findings store, not from prose

**ID:** guard-chunk-bound
**Status:** active
**Chosen:** `check-panel-fix-single-dispatch.sh` accepts the chunked shape, requires numbered
chunks contiguous from 2 with the bare key present, and refuses a round whose chunk count exceeds
`ceil(findings raised in rounds before it / 10)` read via `flow record findings` — the same store
verb `check-panel-findings-closed.sh` uses, unreadable means exit 2, never a clean verdict. Without
the bound, well-formed per-finding keys would legalize the exact KAN-482 abuse the guard exists to
catch. Per-chunk ≤10 itself stays a dispatch-time obligation of the parent (the guard sees only
dispatch rows, not chunk contents — no schema change).
**Considered:** key-shape and contiguity alone — a 24-dispatch round of `panel-fix-1-2..25` would
pass shape and contiguity checks while recreating the 110M-token shape one dispatch at a time;
recording chunk contents in the dispatch row — a store-schema change no consumer needs.

### flow-fast is untouched

**ID:** flow-fast-inline-fixes
**Status:** active
**Chosen:** `skills/flow-fast/review.md` fixes Critical/Major findings inline in the parent (no
panel-fix subagent, no dispatch), and `check-panel-fix-single-dispatch.sh` is not in flow-fast's
guard set — the chunked contract is `/flow`-only.
**Considered:** teaching flow-fast's inline fix loop the cap — nothing dispatches there, so there
is no dispatch size to cap.

## Open questions

### Does per-chunk ≤10 need mechanical enforcement?

**ID:** per-chunk-cap-mechanical
**Status:** open
**Why it is open:** the guard reads dispatch rows and finding rows, never chunk contents; proving
each dispatch held ≤10 findings needs either a schema field or report-file parsing, both wider than
this change.
**What it affects:** if a future run shows chunks persistently overrunning the cap, the dispatch
record or the panel record gains a chunk-size field; today the report file each chunk writes names
the findings it addressed, which is the audit trail.
