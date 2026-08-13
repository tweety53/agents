# KAN-108 follow-up — close the classes KAN-108 left open

## Why

KAN-108 fixed two structural cost problems in `/myflow-do` and left six items behind, recorded on
KAN-153 at the operator's direction: four findings its own review panel raised and the operator
chose not to fix in that round, and two observations from the same run that never became findings.

They are independent of one another, and they share one theme. In every case a guard existed but
could not see the defect class — because the logic it needed had been copied into a sibling script
and drifted, because the requirement it enforced had been restated lossily in the skill that reads
it, or because no guard reads that kind of content at all. Fixing the six instances without closing
the classes would leave the same run repeatable.

Two of the items also carry a factual correction to the ticket, established by reading the repository
rather than the finding text:

- **F43's recorded location is clear, but its defect class is not.** The panel recorded it at
  `skills/myflow-do/SKILL.md:537`; a later fix round of the same KAN-108 run reworded that passage,
  so that line is genuinely gone. A live instance of the same class survives at
  `openspec/specs/myflow-review-panel-economics/spec.md:504` — a single-backtick span holding a
  backslash-escaped backtick, so the span closes early and half the banned-character list renders as
  literal text. Both that line and this change's delta spec, which restated it verbatim, are
  repaired. An earlier draft of this proposal claimed no instance survived; that claim rested on a
  sweep whose output was truncated, and it is withdrawn.
- **F25's two helper sets are not identical.** `check-unfinished-work.sh`'s `ids_of` returns bare
  digits sorted with duplicates retained — `repeated_ids` depends on that — while
  `check-panel-reproducers.sh`'s returns `F<n>` with `sort -u`, and the two `count_matching` bodies
  differ in signature. The extraction has to preserve both behaviours rather than pick one.

## What Changes

- **One definition of the panel-record marker helpers.** New `scripts/lib/panel-record.sh` exports
  `count_matching <file> <ere>`, `grep_lines_of <file> <mode> <ere>` and
  `ids_of <file> <ere> <shape>`, where `shape` is `digits` (sorted, duplicates retained) or
  `ids-unique` (`F<n>`, `sort -u`). `scripts/check-unfinished-work.sh` and
  `scripts/check-panel-reproducers.sh` source it and delete their local copies. Neither guard's
  observable behaviour changes. A library that cannot be sourced is a refusal (exit 2), never a
  guard that silently checks less.
- **The unbounded digit run is bounded.** `check-unfinished-work.sh`'s `findings-total` pattern
  becomes `(0|[1-9][0-9]{0,14})`, matching the bound its sibling already carries, so the two guards
  reading the same record format agree on which totals are well-formed. The ticket's claim that the
  26-digit **crash** is reachable here was checked against the code and is withdrawn: this guard
  compares the declared total to the marker count as strings, never arithmetic.
- **A new guard reads this repository's own Markdown for structural damage.**
  `scripts/check-markdown-integrity.py` — Python 3, standard library only, a real block-structure
  parser in the shape of `check-plan-provenance.py` — over `skills/**/*.md` and `rules/*.mdc`. Four
  signals: a **broken code span** (a backslash-escaped backtick inside a single-backtick span, or an
  odd backtick count outside a fence), an **orphaned blockquote body** (a `> ` line whose preceding
  non-blank line is un-marked prose that does not end a sentence), an **unterminated paragraph** (a
  paragraph outside fences, tables, headings and list items whose last line ends with no terminal
  punctuation), and a **dangling promise** (a paragraph that is the last block before a heading or
  end of file, whose final sentence ends in a colon, with nothing following). Exit 0 clean, 1
  violations found with lines, 2 cannot answer. It joins `## lint`; its harness joins `## test`.
- **The escalation ladder's size clause stops firing on its own.** `the fix diff exceeds
  approximately 150 changed lines` becomes an amplifier rather than an independent trigger: it
  escalates only alongside a risk signal — a new file, a delta spec, a migration, a guard's
  behaviour, or a file outside the set named in the findings. The other four triggers are unchanged.
  This applies KAN-108's own "every trigger SHALL discriminate" requirement to the clause that
  became dominant once the vacuous one was narrowed.
- **The materiality clause regains the qualifier it dropped.** `skills/myflow-do/SKILL.md` states
  that a fix fails materiality when its diff touches only the reproducer's own target **and no path
  the finding named** — the wording the capability spec already carries — so the ordinary
  guard-script case, where the reproducer's target *is* the named path, is no longer disqualified by
  a literal reading. A scenario pins that case.
- **A reproducer runs in its own process group.** `scripts/run-reproducer.sh` execs through
  `python3 -c 'import os,sys; os.setsid(); os.execvp(sys.argv[1], sys.argv[1:])'`, passing the argv
  vector through untouched with no shell involved. Survivor detection gains a `pgrep -g <pgid>` pass
  beside the existing breadth-first descendant walk, and cleanup kills the group before falling back
  to that walk. A fast double fork, which re-parents to launchd before the first poll and is
  invisible to any parentage-based technique, is now detected and named. A missing `python3` is a
  refusal, never a silent fall back to the ungrouped exec.
- **F43's surviving instance is repaired, and the class is closed going forward** by the new guard's
  first signal. The repaired line sits in `openspec/`, outside the guard's scope, so the guard would
  not have found it — recorded as this change's one open question rather than resolved by widening
  the scope late.

## Capabilities

**Modified Capabilities**

- `agents-repo-verification` — the repository's own guard set gains the Markdown-integrity guard and
  the shared marker-helper library, and `.myflow/project.md`'s declared `## lint` and `## test`
  lists grow accordingly.
- `myflow-review-panel-economics` — the auto-escalate trigger set's size clause becomes conditional;
  the materiality condition gains the case where the reproducer's target is the finding's named
  path; the reproducer's survivor detection gains process-group isolation.

**New Capabilities**

None. Every change lands inside a capability that already exists.

## Impact

- New: `scripts/lib/panel-record.sh`, `scripts/check-markdown-integrity.py`,
  `scripts/test-check-markdown-integrity.sh`.
- Modified: `openspec/specs/myflow-review-panel-economics/spec.md` (the one-line code-span repair
  above), `scripts/check-unfinished-work.sh`, `scripts/check-panel-reproducers.sh`,
  `scripts/run-reproducer.sh`, `scripts/test-run-reproducer.sh`,
  `scripts/test-check-unfinished-work.sh`, `skills/myflow-do/SKILL.md`, `.myflow/project.md`.
- No breaking changes. Both refactored guards keep their exit-code contracts and their reported
  output; both harnesses pass unchanged.
