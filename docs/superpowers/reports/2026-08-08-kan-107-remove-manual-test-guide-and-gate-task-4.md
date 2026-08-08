# Task 4 report — Replace `/myflow-do` section 6 with run-instruction resolution

## Status

DONE

## Plan provenance

The brief's fenced block (`scripts/check-contract-budget.sh`, tagged
`verified:authored in-tree for this change`) is a bare invocation of an existing, already-committed
script with no arguments. It fits its surroundings: the same command, run the same way, appears
throughout the other tasks' briefs and in `skills/myflow-do/SKILL.md` section 4 (budget checks are
part of the ordinary implementer workflow). No mismatch found.

## What changed

### `skills/myflow-do/SKILL.md`

- Retitled section 6 from `## 6. Write the manual test guide` to `## 6. Resolve the run
  instructions`. The new section writes no file. It resolves, for the handoff: every app root
  (absolute, from `git worktree list` or the state file's `worktrees` keys); every start command
  (from `.myflow/project.md`'s `## run`, paths made absolute); every URL (from this worktree's
  workspace id the way section 2 computes it, never the project's declared base — a project
  declaring no isolation resolves nothing); and, for a project with no runnable application, the
  `## lint` and `## test` commands, absolute.
- Kept, in substance, the three bullets the brief named as must-survive: absolute paths resolved
  from `git worktree list` / the state file's `worktrees` keys; resolving URLs from the workspace id
  computed in section 2 rather than from a later-exported variable; and an application whose port is
  fixed outside its own repository keeping its default with a short note on the same line.
- Deleted the checklist-specific material that no longer applies to a value that is only ever
  printed: the tickable-line register and its phrasing rule, the capability-scope/blast-radius
  grouping rule, the `## Known incomplete` section rule (and its `scripts/check-unfinished-work.sh`
  cross-reference), the ticked-box-preservation-on-fix-run rule, the "no skip prompt" rule, and the
  closing paragraph about the two machine-read shapes (`- [ ]`/`- [x]` markers and the `## Known
  incomplete` header) — that paragraph existed only because a guard parsed the file, and this section
  no longer produces a file for any guard to parse.
- Updated the handoff block (former `Test guide:` line and the instruction line beneath it, at the
  old line 536): replaced with the `Run it:` shape from Task 5's brief —

  ```
  Worktree:   <absolute worktree path>

  Run it:
    <command>          # <app or check name>
    <command>

  Review the diff, then run it:
    git -C <absolute worktree path> diff --cached          # when staged and uncommitted
    git -C <absolute worktree path> diff <merge base>..HEAD  # when committed and pushed
    open -na "IntelliJ IDEA" --args "<absolute worktree path>"
  ```

  Kept the citation to **The block each state renders** (`skills/myflow-contracts/handoff-blocks.md`)
  in the paragraph right below the block — that paragraph still describes the same two-line review
  command pairing and needed no rewording.
- Left the frontmatter `description:` (line 3) and the intro line ("write its manual test guide",
  line 8) untouched — not in this task's file list, and per the brief's context note these belong to
  Task 7 ("the documentation digests and the two `myflow-do` command files"). Flagging this in case
  Task 7's scope is read narrower than that: two literal mentions of "manual test guide" remain in
  this file outside my edited range.

### `skills/myflow-do/SKILL-rationale.md`

- Rewrote the `## 6. Write the manual test guide` section (now `## 6. Resolve the run
  instructions`) per Step 3 of the brief: the guide-in-one-run rationale becomes the reason the run
  instructions are resolved in the same run and printed rather than written — nothing reads them
  back, so a file would be a record with no consumer, and printing them means they are never a
  round stale relative to the worktree they describe.
- Kept the surviving rationale bullet about ports being bound per-worktree (why a URL from the
  project's declared base, rather than this worktree's resolved one, would reach a different
  change's application) — reworded to match the new "run instructions" framing instead of "the
  documented URL" in a guide.
- Kept the per-plan-task-vs-behaviour-scope bullet out — that rule was about the checklist's
  granularity, which no longer exists as a written artifact; nothing analogous applies to run
  instructions, so it was not carried forward. Flagging this in case a reviewer expected it kept in
  some form.

### `scripts/check-contract-budget.sh`

- Not modified. Both files shrank (`skills/myflow-do/SKILL.md`: 35,361 bytes vs. its row of 44,781;
  `skills/myflow-do/SKILL-rationale.md`: 8,372 bytes vs. its row of 10,228). Neither grew past its
  row, so per the brief ("If any file grew past its row, raise that row... never narrow the guard")
  no change was made. Rows are now generous headroom rather than tight — left as-is since narrowing
  is explicitly the wrong response.

## Verification

- `bash scripts/check-contract-budget.sh` → `BUDGET-OK: 24 contract file(s) within budget`, exit 0.
- `bash scripts/check-references.sh` → `check-references: all referenced sections resolve`, exit 0.
- No test-driven-development sub-skill applies: this task is Markdown contract prose, and no test
  harness asserts its content. Verification leaned on the two guard scripts above, which are the
  only automated checks that read this file's structure (byte budget and cross-reference
  resolution).

## Test summary

`check-contract-budget.sh` PASS (both files under their existing rows), `check-references.sh` PASS
(no dangling citations introduced or left behind).

## Concerns

1. Two literal "manual test guide" mentions remain in `skills/myflow-do/SKILL.md` outside my
   assigned lines: the frontmatter `description:` (line 3) and the intro sentence (line 8). Per the
   brief's context note these are Task 7's ("the two `myflow-do` command files"), but if that scope
   is read as the external slash-command wrapper files only (not this `SKILL.md`'s own header),
   these two lines will need a follow-up edit.
2. The rationale rewrite could not carry forward the "per-plan-task vs. behaviour-scope" bullet in
   any form, since that concern was specific to a checklist artifact that no longer exists. Worth a
   second look in case something analogous should be said about run-instruction scope (e.g. "every
   app in scope, not one per task") — I judged it out of scope for this task since the brief's Step
   3 names only the "why resolved and printed" reasoning to carry over.
3. Did not touch `git worktree list` derived state or the `.myflow/project.md` `## run` key format
   — this task is prose-only per the brief, so no code/script was added to actually perform the
   resolution; the section describes what a run must do, consistent with how every other section in
   this file is written (prose instructions the model executing `/myflow-do` follows, not code).

## Files changed

- `skills/myflow-do/SKILL.md`
- `skills/myflow-do/SKILL-rationale.md`

Both `git add`-ed in the worktree; no commit made.
