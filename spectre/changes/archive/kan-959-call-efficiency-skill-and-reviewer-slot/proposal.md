# kan-959-call-efficiency-skill-and-reviewer-slot

## Why

The KAN-924 call-efficiency audit (`gymie/docs/audits/kan-924-call-efficiency.md`, Findings and
Prevention) found three finding classes that recur: the same read sent by two independent callers
(F1, F2, F3, F7), the same row or query repeated within one request (F11, F12, F14), and one query
per item of a list (F9, F13). None of them shows in a diff's tests, and no review-panel slot asks
about them, so they ship. The audit's measuring method lives only in that one report, so the next
audit starts from nothing.

The panel's test-sanity reviewer has been an experimental slot, rolled onto a panel only 30% of the
time and skipped for the bundle cap on most compact rosters; the operator wants it on every run.

## What changes

- A new standalone skill, `/call-efficiency-audit`, carries the KAN-924 measuring method for any project's
  dev stack: a browser request log over cold opens, per-endpoint SQL statement counts, Redis
  `MONITOR`, the per-layer checklists, the findings table and the before/after comparison.
- A new review-panel slot, `call-efficiency`, reads the diff and asks, of every new or changed read
  or effect, who else reads or causes the same thing, whether the same row or query repeats within
  one request, and whether a list maps to one query per item.
- `test-sanity` stops being experimental: its prompt becomes a regular slot template.
- Every `/flow` and `/flow-fast` panel — every class, `micro` included, and under every automatic
  reduction — dispatches one extra pass-1 bundle, `test-sanity+call-efficiency`, beside the
  primary+principles floor. Its model is a reproducible per-change 50/50 roll between `sonnet` and
  `opus`, always at effort `low`; `plan-class.sh` prints it and the decision records it as
  `extra_dispatch`.
