# kan-382-flow-review-panel-reproducer-shape-rejected

## Why

Panel slots record reproducers that `check-panel-reproducers.sh` rejects — shell command lines
carrying quotes and metacharacters, or exemption lines that are not the literal `none — <reason>`
form. The rejection fires at read time, after the findings are already in the store, so every
finding in the round loses its automated per-finding reproducer run and needs manual verification
instead. KAN-382 observed all 6 findings of one round rejected this way.

## What changes

- `flow record finding` validates the full accepted reproducer shape before the store is contacted
  and refuses a bad one with a per-rule message, so a malformed reproducer bounces to the raising
  slot instead of landing silently.
- The slot brief in `skills/flow/review-panel.md` states the accepted shape verbatim — bare
  relative path plus plain arguments, the five prohibitions, the exact `none — <reason>`
  exemption — so slots record conforming reproducers in the first place.
