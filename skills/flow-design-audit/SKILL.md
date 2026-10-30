---
name: flow-design-audit
description: Audit a Gymie flow doc or review page (frames, catalog, rules, data files) for inconsistency, ambiguity and missing critical dimensions. Fix the first two, present the third for approval.
---

# /flow-design-audit <doc>

Run on any Gymie design artifact: a `Gymie <Feature> Flow.dc.html`, a review page (e.g. `Exercise Catalog Review.dc.html`), or the data file behind it (`_gen/*.js`, `catalog/*.js`, `specs/*.md`).

## 0. Load the truth, not the render
- Read the data source the page is built from (catalog JS, `_gen` data, spec md) and evaluate it in a script — audit the data, then the copy.
- Read the decisions that bind it: `<project>/CLAUDE.md` (Flows so far), the rules spec (`specs/*.md`, R-numbers), `<project>/catalog/data_model.md`, shipped code / gymie-playwright fidelity snapshots for shipped screens.

## 1. Inconsistency — fix, then list as "was → now"
Things that say two different things.
- Copy vs rules: legends, captions, notes vs the R-rules and CLAUDE.md decisions.
- Copy vs data: a note or legend claims something the data never does.
- Data vs data: name vs equipment / key vs name / group vs members (a generic named for one implement holding another).
- Stale leftovers: buckets, enums or maps a later turn replaced; dead code maps.
- Sibling doors (parameter parity): a field present on one door missing on its siblings.
- Doc vs doc: the same list or default stated differently in two docs — pick the source of truth and make the other derive from it.
- Frames: fills, type, bars, menus vs shipped snapshots; no dead ends; annotations marked `data-note`.

## 2. Ambiguity — fix, then list as "was → now"
Things a builder could read two ways.
- Keys that contradict names (rename only while the key is still new; prod keys never change).
- A prod key repurposed for a different meaning.
- Counts that don't add up without an explanation.
- Columns / tags / statuses with no definition on the page.
- "Where does the member set this?" — every setting names its screen.
- Promises with nothing behind them (e.g. "search finds RDL" with no alias data).

## 3. Missing critical dimension — DO NOT build; list for approval
Ask: what axis does every sibling feature have that this one lacks? (Templates shipped without GOAL — that is the shape of miss to look for.)
Checklist of axes for Gymie:
- Goal (Hypertrophy · Strength · Endurance) and what changes per goal.
- Experience / difficulty (Profile → Experience).
- Equipment profile coverage (bodyweight only · dumbbells + bench · full gym) — can every slot be filled?
- Per side / unilateral; load entry type (total, per hand, added, assisted, stack).
- History / progression identity (what "Last time" reads after a swap).
- Muscles (target + secondary) wherever a rule weights by muscle.
- Time (setup, rests) wherever the UI promises a duration.
- Injury / joint limits; offline; read-only (switched-in) state; admin editability; search aliases.
- Defaults that conflict between the doc and the data.
Each finding: id (M1…), what, why it matters (which rule/screen breaks without it), in plain words with a concrete example. Present them in the doc ("MISSING · FOR APPROVAL") and as an ask_user chips question so the user ticks what to build. Mark each DONE / RESOLVED / OPEN as the user decides.

## 4. Record
- Add an `#audit` section to the doc: INCONSISTENCIES · FIXED, AMBIGUITIES · FIXED, MISSING · FOR APPROVAL.
- One line in the doc intro: "Audited <date> against …".
- Regenerate every derived file (CSV, spec md, dependent flow docs) from the fixed data; diff old vs new and confirm only intended rows/frames changed.
- Update CLAUDE.md "Flows so far" with the decisions once approved.
