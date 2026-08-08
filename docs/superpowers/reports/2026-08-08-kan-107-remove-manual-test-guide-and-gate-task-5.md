## Task 5 report: Put the run instructions into the handoff template

**Status:** DONE

### What changed

**`skills/myflow-contracts/handoff-blocks.md`**

- Replaced the `IN_PROGRESS`-after-`/myflow-do` template's `Worktree:` / `Test guide:` / "Review the
  diff, then run the apps against the guide:" lines with the brief's exact block: `Worktree:`
  followed by a blank line, then `Run it:` with two placeholder command lines, then "Review the
  diff, then run it:" followed by the review-command placeholder and the `open -na "IntelliJ IDEA"`
  line. Pasted the brief's fenced block verbatim (tag `verified:authored in-tree for this change`);
  it fit its surroundings exactly, no adjustment needed.
- Added a new note directly after the existing "`Panel` is `(run-only)`" note: "`Run it:` is
  on-disk, not `(run-only)`" — explaining the commands are resolved from the worktree and the
  project's own configuration, not remembered from the run, so `/myflow-status <name>` regenerates
  the same lines rather than omitting the section. This follows the file's existing convention of a
  one-line callout after each template stating a field's on-disk/run-only status.

**`skills/myflow-contracts/handoff-blocks-rationale.md`**

- Section "Why `IN_PROGRESS` needs two renderings" named "a worktree path, a test-guide path and a
  staged-diff command" as fields wrong for run 1's rendering. Changed "a test-guide path" to "run
  instructions" so the rationale no longer names a guide that no longer exists.

**`skills/myflow-status/SKILL.md`**

- Deleted the bullet `- The manual test guide's **absolute** path + checked/total box count` from
  section 4 (Detail view). Added nothing — the regenerated handoff block, printed a few lines later
  in the same section, now carries the `Run it:` instructions per the updated template.
- Did not touch lines 112/158 (table rows still saying "review the diff + run the guide") — out of
  this task's named scope (`SKILL.md:176` only); flagging for whichever task owns the command-file
  digest sweep (brief says Task 7 owns digests/command files).

### Step 3 — side-by-side comparison

Read `skills/myflow-do/SKILL.md:499-522` (Task 4's landed block) against the new
`handoff-blocks.md:99-121` template. Folded lines, fields and order match exactly:

- `Worktree:   <absolute worktree path>` — same line, same label, same spacing.
- Blank line, then `Run it:` with two indented command lines — same shape (template uses
  placeholders `<command> # <app or check name>` / `<command>`; the skill's copy is the same
  placeholder form, since the skill only ever prints resolved commands at run time, not alternatives
  to enumerate).
- Blank line, then `Review the diff, then run it:` — identical label text in both files.
- The skill's copy enumerates the two `git diff` alternatives (staged/committed) as its own command
  emits only those two, plus a trailing comment naming which case each is for; the template's single
  `<the review command that matches the git state on the Staged line — see below>` placeholder
  describes the value space across both renderers (including the third, `/myflow-status`-only case)
  — this is the documented enumeration-is-a-refinement pattern, not a mismatch.
- Both end with the identical `open -na "IntelliJ IDEA" --args "<absolute worktree path>"` line.

No mismatch found; nothing needed reconciling.

### Verification

- `scripts/check-references.sh` → `check-references: all referenced sections resolve`
- `scripts/check-contract-budget.sh` → `BUDGET-OK: 24 contract file(s) within budget` — no row
  needed raising, so `scripts/check-contract-budget.sh` itself was not touched.
- No test exists or was written: this task is contract prose in Markdown: no harness asserts
  template content beyond the reference-resolution and byte-budget checks above, which is what the
  brief's Required Sub-Skill note anticipates. Said so plainly rather than inventing a test.

### Plan provenance

The brief's fenced block (tag `verified:authored in-tree for this change`) was checked against its
surroundings before pasting — labels, indentation and placeholder style all matched the existing
template conventions in `handoff-blocks.md`. Nothing needed adjusting.

### Files changed

- `skills/myflow-contracts/handoff-blocks.md`
- `skills/myflow-contracts/handoff-blocks-rationale.md`
- `skills/myflow-status/SKILL.md`

### Concerns

None. `scripts/check-unfinished-work.sh`, `scripts/test-check-unfinished-work.sh`,
`scripts/gather-self-review-context.sh`, `scripts/test-gather-self-review-context.sh`,
`scripts/test-uncommitted-review-package.sh`, `skills/myflow-contracts/pipeline.md`,
`skills/myflow-contracts/pipeline-rationale.md`, `skills/myflow-contracts/finish-contract.md`,
`skills/myflow-do/SKILL.md`, `skills/myflow-do/SKILL-rationale.md`, `skills/myflow-do/scripts/*`
and `skills/myflow-finish/SKILL.md` were not touched, per the brief's note that Tasks 1–4 already
landed uncommitted work there. `docs/manual-test/` deletion and the remaining contract sweep are
left for Task 6; documentation digests and command files are left for Task 7.
