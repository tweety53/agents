## Task 7 report: Update the documentation digests and command files

**Status:** DONE

### What changed

All six files named in the brief, and only those:

- `CLAUDE.md` — skill-index row (`skills/myflow-do/`), the digest paragraph, and the
  `/myflow-do` table row's planning-path list. Also fixed the gate row's "run the apps against
  the guide" wording, which the brief didn't enumerate by line but which names the retired
  artifact.
- `AGENTS.md` — same three sites, in that file's own (already slightly different) voice, plus
  the same gate-row fix.
- `rules/myflow-manual-review.mdc` — the enumerated body line (19), plus the frontmatter
  `description:` line, which also named the manual test guide and isn't a file the brief's line
  list called out but falls under the same "no active file names the guide" requirement.
- `README.md` — all three sites: the pipeline-digest paragraph (~line 58), the Level-1 stage
  table's `/myflow-do` row (~line 97), and the commands-reference table row (~line 511/512).
  Also fixed both stage-table and reference-table gate cells that said "run the apps against the
  guide."
- `commands/myflow-do.md` and `commands-claude/myflow-do.md` — frontmatter `description:` and
  body line, exactly per Step 2's given text, plus the trailing "When done" line's "against the
  guide" wording.

Every site now states: `/myflow-do` produces the staged diff **and** the run instructions, so
reviewing and running are one sitting/gate. The gate wording "review the staged diff and run the
apps" is preserved verbatim everywhere it already appeared; only the guide-specific tail
("...against the guide") was dropped, since pipeline.md's own canonical gate wording carries no
such tail either.

The `CLAUDE.md`/`AGENTS.md`/`README.md`/rule-file copies of the pipeline digest were deliberately
**not** made byte-identical to each other or to `pipeline.md` — each kept its own existing voice
(e.g. "Codex loads this file" vs "Claude Code loads this file"), per the brief's explicit
instruction.

The planning-path list in `CLAUDE.md` and `AGENTS.md` (`/myflow-do` table row) was trimmed from
three paths (`openspec/`, `docs/manual-test/`, `docs/superpowers/`) to the two that
`skills/myflow-contracts/pipeline.md` now names (`openspec/`, `docs/superpowers/`), matching
Task 3's change to the actual guard.

### Verification

Documentation prose has no test harness asserting its content — noted rather than invented.

Step 3's installer round-trip, run exactly as given in the brief:

```
SANDBOX="$(mktemp -d)"
HOME="$SANDBOX" ./setup.sh global
scripts/test-setup.sh
```

- `./setup.sh global` → exit 0, `✅ Global setup complete.`
- `scripts/test-setup.sh` → `✓ PASS — 192 assertions, 0 failures`
- `grep -rl "manual test guide\|docs/manual-test" "$SANDBOX"` → no matches (grep exit 1)

Additionally ran every guard in `.myflow/project.md`'s `## lint` section against the working
tree, all exit 0 / clean:

- `scripts/check-vocabulary.sh` → `✓ Stage-vocabulary guard: clean`, `✓ Panel-vocabulary guard: clean`
- `scripts/check-references.sh` → `check-references: all referenced sections resolve`
- `scripts/check-plan-provenance.sh` → `3 file(s) scanned, all provenance stated`
- `scripts/check-task-build-green.sh` → exit 0
- `scripts/check-workspace-isolation.sh` → `ISOLATION-OK` (repo declares no section)
- `scripts/check-contract-budget.sh` → `BUDGET-OK: 24 contract file(s) within budget` (none of my
  six files are contract files, so the ratchet owes nothing — confirmed by the count and by the
  fact none of my paths appear in its budgeted-file list)

`git diff --stat` confirms only the six named files changed:

```
 AGENTS.md                      | 8 ++++----
 CLAUDE.md                      | 8 ++++----
 README.md                      | 8 ++++----
 commands-claude/myflow-do.md   | 6 +++---
 commands/myflow-do.md          | 6 +++---
 rules/myflow-manual-review.mdc | 4 ++--
 6 files changed, 20 insertions(+), 20 deletions(-)
```

Ran `grep -n "guide"` and `grep -n "docs/manual-test"` across the six files after editing — zero
matches in both, confirming the "no active file outside `openspec/changes/archive/` and
`docs/superpowers/` names the guide" interface constraint is satisfied for these six files.

### Plan provenance check

The brief's fenced block for Step 3 is tagged `verified:authored in-tree for this change` (not
`verified:copied from .myflow/project.md` as the task assignment message stated) — the two
commands (`HOME="$SANDBOX" ./setup.sh global` and `scripts/test-setup.sh`) match exactly what
exists at the repo root and ran clean with no adaptation needed, so the block fits regardless of
which provenance tag is correct on it. Flagging the tag-name mismatch between the assignment
message and the brief file for visibility; it did not block or change the work.

### Concerns

None. Went slightly beyond the brief's literal enumerated line list in three places — the
`rules/myflow-manual-review.mdc` frontmatter description, and the "run the apps against the
guide" phrase in `CLAUDE.md`, `AGENTS.md`, `README.md` (two sites), and both command files' "When
done" line — because leaving them would have left an active, non-archived file naming the retired
guide, contradicting the task's own "Produces" interface note. All of these edits landed inside
the same six files already in scope; no new files were touched.
