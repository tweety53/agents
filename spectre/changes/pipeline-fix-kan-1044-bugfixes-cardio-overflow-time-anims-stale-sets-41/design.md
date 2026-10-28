## Context

Defect: operator instruction 2026-10-28 after kan-1044 Task 11 OOM'd the laptop. Text that prescribes it:

- `rules/fix-determinism-at-the-source.mdc` § "Read the flake before touching anything", second paragraph
  (parallel load / CPU cap as the way to make a flake fail; rate under that condition sizes proof runs).
- `skills/flow/implement.md` `> **PROOF RUNS:**` block (~line 674): "10–15 runs by default".
- `skills/flow/verify-fix-loop.md` item 2 (~lines 104–121): "closes after two more full runs … twice
  more", and "reproduces it under … the same concurrent load and the full-suite ordering, never the
  test alone".
- `skills/flow-self-review/SKILL.md` line 53–54 only defers to verify-fix-loop; no own wording to change.
- `stats/internal/guard/dispatchparagraphs.go` `proofruns` entry pins PROOF RUNS phrases
  ("size the run count to the flake rate you measured", …) — must change with the paragraph.

Blast radius: 20 files (`grep -rlE "PROOF RUNS|fix-determinism-at-the-source|twice more|full-suite ordering"`
outside archive = 19, plus the gymie memory file). Not `big`. Files actually edited: 5 in agents + 1 memory.
Citing files (pipeline.md, agent-baseline.md, visual-verify*.md, primary-reviewer-prompt.md,
unfinished-work-gate.md, verify-and-handoff.md, known-bugs.md, references.go, installedcitations.go,
docs/*) cite the section/rule name only, which does not change — no edit.

## Decisions

- **Edit in place, no new rule file.** Simplest: three paragraphs rewritten; rule name and section
  headings unchanged, so no citation moves.
- **Guard = existing `proofruns` phrase list.** Replace its phrases with ones from the new paragraph
  ("never a load reproduction", "single worker", "at most 5 runs", "shows the check failing with the
  mechanism removed"); add a test case whose implement.md fixture carries the old "10–15 runs by
  default" paragraph and expects exit 1. Fails without the fix.
- **Core excerpt of the rule (`<!-- core -->`) untouched** — it carries no load wording; the rendered
  global CLAUDE.md block stays as is.
- **Memory** `/Users/tweety53/.claude/projects/-Users-tweety53-Projects-gymie/memory/no-load-testing-on-laptop.md`:
  add to "How to apply" one line "Canonical: `rules/fix-determinism-at-the-source.mdc` § Read the flake
  before touching anything; PROOF RUNS in `skills/flow/implement.md`." No other wording change.
- No repair of past records: the defect left no store row or file wrong (kan-1044's plan task is that
  run's own concern, already stopped by the operator).

## Open questions

None.
