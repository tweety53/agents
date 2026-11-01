# kan-372-flow-the-router-eager-loads-11-contracts-before

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.
> **Relocation:** yes — every task moves existing prose within `skills/flow/` and
> `skills/flow-contracts/`; no normative sentence, exit-code contract, scenario, or rejected-
> alternative rationale is reworded.

## Global constraints

- `~/.claude/rules/be-brief.md` — cut, never paraphrase. Delete a restating passage; never reword
  one to say the same thing shorter.
- `scripts/check-references.sh`, `scripts/check-contract-budget.sh`, and every other
  `scripts/check-*.sh` guard must exit clean after every task.
- No `/flow*` run ever loads a `*-rationale.md` appendix — that arrangement is untouched by this
  plan.

- [x] 1. Relocate the 10 non-`pipeline.md` eager loads out of `skills/flow/SKILL.md`

**Build:** green

**Files:**
- Modify: `skills/flow/SKILL.md`
- Modify: `skills/flow/brainstorm.md`
- Modify: `skills/flow/implement.md`
- Modify: `skills/flow/integrate.md`
- Modify: `skills/flow/archive.md`
- Modify: `skills/flow/review-panel.md`
- Modify: `skills/flow/verify-and-handoff.md`

**Tests:** **none** — prose relocation, no code
**Regression:** n/a — no test declared
**Baseline:** n/a — no test declared
**Commit:** `docs(flow): make the router's contract loads lazy`

  - [ ] **Step 1: Delete the eager-load paragraph in `skills/flow/SKILL.md`**

  Delete the paragraph beginning "**Load `skills/flow-contracts/worktree-resolution.md`,
  `skills/flow-contracts/session-records.md`, ...**" (currently lines 33-37, ending "... rather
  than restating their content."). Keep the preceding "**Load `skills/flow-contracts/pipeline.md`
  first**" paragraph unchanged — the router still needs it eagerly, to resolve state before any
  phase file is dispatched into.

  - [ ] **Step 2: Relocate the `model-policy.md` paragraph**

  Cut the paragraph beginning "**`skills/flow-contracts/model-policy.md` is only partly current
  for `/flow`.**" from the SKILL.md preamble (immediately after the deleted paragraph in Step 1).
  Paste it, unchanged, immediately above the `## Model resolution` heading later in the same file.

  - [ ] **Step 3: Add the missing explicit Load lines, one per phase file, at the point each
  contract is actually used**

  For each row below, insert a sentence of the shape `**Load `skills/flow-contracts/<file>.md`**
  — <existing citation prose, reused>.` immediately before the first paragraph in that phase file
  that currently cites the contract without an explicit "Load" imperative. Do not alter the cited
  prose itself — only add the imperative sentence introducing it.

  | Contract | Phase file | Insert before |
  |---|---|---|
  | `jira-integration.md` | `brainstorm.md` | "**Resolve the linked Jira issue first**" (section A) |
  | `jira-integration.md` | `implement.md` | "If the fix adds scope the linked Jira issue does not describe" (step 3) |
  | `jira-integration.md` | `integrate.md` | "**Sub-step: transition the issue to In Review**" (step 4) |
  | `jira-integration.md` | `archive.md` | "**Transition the issue to Done**" (step 9 preamble) |
  | `plan-provenance.md` | `brainstorm.md` | "While enriching `tasks.md`, tag every fenced block" (section D) |
  | `build-green.md` | `brainstorm.md` | "Also tag every task with `**Build:**`" (section D) |
  | `project-configuration.md` | `implement.md` | "Resolve `[STANDARDS_PATHS]`" equivalent / lint-test resolution point (step 1's proposal/tasks read, or wherever it first cites the file) |
  | `project-configuration.md` | `review-panel.md` | "read `<project>/.flow/project.md`'s `## review panel citation check` key" (top of file) |
  | `project-configuration.md` | `verify-and-handoff.md` | "Run the `## lint` and `## test` commands" (Verify section) |
  | `project-configuration.md` | `archive.md` | "runs the project's `remove` command" (step 5) |
  | `workspace-isolation.md` | `implement.md` | "compute this worktree's workspace id" (step 2) |
  | `workspace-isolation.md` | `verify-and-handoff.md` | "A declared `cache index` row" (Verify section) |

  `worktree-resolution.md`, `session-records.md`, `git-boundaries.md`, and `artifacts-registry.md`
  already carry their own explicit Load line at point of use in `verify-and-handoff.md`,
  `integrate.md`, and `archive.md` — no change needed there. Add one to `implement.md`, before
  "**This run's resolved worktree set**" (step 2), for `worktree-resolution.md`, since that file
  currently only cites "Resolving a change's worktrees" without an explicit Load imperative.

  - [ ] **Step 4: Verify no content was reworded, only relocated**

  Run: `git diff --stat skills/flow/SKILL.md skills/flow/brainstorm.md skills/flow/implement.md skills/flow/integrate.md skills/flow/archive.md skills/flow/review-panel.md skills/flow/verify-and-handoff.md`
  Expected: every changed file shows added lines that are new imperative sentences (or, for
  SKILL.md, one paragraph moved) — confirm by reading the diff that no existing sentence's wording
  changed.

  - [ ] **Step 5: Run the reference and budget guards**

  Run: `scripts/check-references.sh && scripts/check-contract-budget.sh`
  Expected: both exit 0. `skills/flow/SKILL.md` shrinks well under its 16278-byte budget row; no
  budget row needs raising for this task.

  - [ ] **Step 6: Commit**

```bash unverified:confirm exact paths against the worktree at execution time
git add skills/flow/SKILL.md skills/flow/brainstorm.md skills/flow/implement.md \
  skills/flow/integrate.md skills/flow/archive.md skills/flow/review-panel.md \
  skills/flow/verify-and-handoff.md
git commit -m "docs(flow): make the router's contract loads lazy"
```

---

- [x] 2. Split `finish-contract.md` into `finish-contract-run1.md` and `finish-contract-run2.md`

**Build:** green

**Files:**
- Create: `skills/flow-contracts/finish-contract-run1.md`
- Create: `skills/flow-contracts/finish-contract-run2.md`
- Delete: `skills/flow-contracts/finish-contract.md`
- Modify: `skills/flow/integrate.md`
- Modify: `skills/flow/archive.md`
- Modify: `skills/flow-contracts/worktree-resolution.md`
- Modify: `skills/flow-contracts/jira-followups.md`
- Modify: `skills/flow-contracts/jira-integration.md`
- Modify: `skills/flow-contracts/state-file.md`
- Modify: `skills/flow-contracts/artifacts-registry.md`
- Modify: `skills/flow-contracts/project-configuration.md`
- Modify: `skills/flow-contracts/pipeline.md`
- Modify: `skills/flow-contracts/SKILL.md`
- Modify: `scripts/check-contract-budget.sh`
- Modify: `skills/flow-contracts/artifacts-registry-rationale.md`
- Modify: `skills/flow-contracts/handoff-blocks-rationale.md`
- Modify: `skills/flow-contracts/model-policy-rationale.md`
- Modify: `skills/flow-contracts/model-policy.md`
- Modify: `skills/flow-contracts/pipeline-rationale.md`
- Modify: `skills/flow-contracts/project-configuration-rationale.md`
- Modify: `skills/flow-status/SKILL.md`
- Modify: `AGENTS.md`
- Modify: `CLAUDE.md`
- Modify: `README.md`
- Modify: `rules/flow-manual-review.mdc`
- Modify: `scripts/check-cleanup-complete.sh`
- Modify: `scripts/test-check-cleanup-complete.sh`
- Modify: `scripts/test-check-contract-budget.sh`

**Ruling (added during implementation):** the task's original Files list only named the 11 citers
`grep -n "finish-contract"` surfaced during planning; a full-repo `check-references.sh`-equivalent
scan found 7 more real citers (rationale-appendix cross-references, plus `flow-status/SKILL.md`,
which loads this contract too). Added here since they are real, required repoints — leaving them
undeclared would only hide a plan gap from the commit-fields guard, not remove the need to fix them.

**Second ruling (added at `flow.verify`):** running this change's full lint/test suite (not just
per-task and panel guards) surfaced further dangling `finish-contract.md` citations the original
survey's `skills/` scope missed — `AGENTS.md`, `CLAUDE.md`, `README.md` (repo root), and
`rules/flow-manual-review.mdc` (an always-on rule file) — plus a stale fixture filename in
`scripts/test-check-contract-budget.sh` (its "large sibling" fixture assumed `finish-contract.md`
still had a declared budget row) and two stale comment citations in
`scripts/check-cleanup-complete.sh`/`scripts/test-check-cleanup-complete.sh`. Fixed and folded into
this task's commit via `git commit --fixup` + `git rebase --autosquash`, same mechanism as the
first ruling.

**Tests:** **none** — prose relocation, no code
**Regression:** n/a — no test declared
**Baseline:** n/a — no test declared
**Commit:** `docs(flow-contracts): split finish-contract.md by run`

  - [ ] **Step 1: Create `finish-contract-run1.md`**

  Copy `skills/flow-contracts/finish-contract.md`'s content from its top (`# Finish contract...`)
  through the end of "### Resolving a change's worktrees" (verified: these sections are cited
  only by `integrate.md` and by `worktree-resolution.md`'s own reference to "Resolving a change's
  worktrees" — confirmed via `grep -rn "Resolving a change" skills/flow/*.md
  skills/flow-contracts/*.md` returning only `integrate.md` as a phase-file citer) into the new
  file, retitled `# Finish contract — run 1 (the branch is not merged)`. Update the file's own
  opening sentence ("This file is canonical for...") to say it is canonical for run 1's
  preflight-signal decision, run 1's procedure, and worktree resolution — dropping the reference to
  run 2 content that no longer lives here.

  - [ ] **Step 2: Create `finish-contract-run2.md`**

  Copy the remaining content — "### Run 2 — the branch is merged" through "### Worktree cleanup"
  and the trailing shell-comment block (verified: cited only by `archive.md`, via
  `grep -n "Worktree cleanup" skills/flow/*.md` returning only `archive.md`) into the new file,
  retitled `# Finish contract — run 2 (the branch is merged)`. Give it the same canonical-scope
  opening sentence, naming run 2's procedure and worktree cleanup.

  - [ ] **Step 3: Delete the old `finish-contract.md`**

  `git rm skills/flow-contracts/finish-contract.md`

  - [ ] **Step 4: Repoint every citer**

  For each file below, replace the cited section's path from `skills/flow-contracts/finish-contract.md`
  to whichever new file now carries that section — determined by which half (Step 1 or Step 2)
  the cited heading landed in:

  | File | Cited section(s) | New target |
  |---|---|---|
  | `skills/flow/integrate.md` | "Finish contract" overview, "Resolving a change's worktrees" | `finish-contract-run1.md` |
  | `skills/flow/archive.md` | "Run 2 — the branch is merged", "Worktree cleanup" | `finish-contract-run2.md` |
  | `skills/flow-contracts/worktree-resolution.md` | "Resolving a change's worktrees" | `finish-contract-run1.md` |
  | `skills/flow-contracts/jira-followups.md` | "Run 1 — the branch is not merged" (both citations) | `finish-contract-run1.md` |
  | `skills/flow-contracts/jira-integration.md` | "Run 2 — the branch is merged", step 9 | `finish-contract-run2.md` |
  | `skills/flow-contracts/state-file.md` | generic "Finish contract" re: `link.md`'s Merge order | read the citing sentence's context; point at `finish-contract-run1.md` if it concerns the preflight/Run-1 flow, `finish-contract-run2.md` if it concerns archive/cleanup |
  | `skills/flow-contracts/artifacts-registry.md` | generic (two citations: "procedure for the rows removed", "canonical for that copy, for the change-name") | both concern archive-time removal → `finish-contract-run2.md` |
  | `skills/flow-contracts/project-configuration.md` | "Run 2 — the branch is merged" (workspace remove step), "Worktree cleanup" | `finish-contract-run2.md` |
  | `skills/flow-contracts/pipeline.md` | generic "Finish contract" (line 4 intro, line 411 "Finish contract" section) | replace with two citations, one to each new file, or retitle the umbrella reference to name both — whichever reads cleanest; do not drop either half |
  | `skills/flow-contracts/SKILL.md` | its own index row for `finish-contract.md` | replace with two rows, one per new file, each naming which command loads it (`integrate.md` → run1; `archive.md` → run2) |

  - [ ] **Step 5: Update `scripts/check-contract-budget.sh`'s table**

  Remove the row `skills/flow-contracts/finish-contract.md 54286`. Add two rows, each sized at the
  new file's actual byte count plus 25%, per that script's own ratchet rule (see its header
  comment): `skills/flow-contracts/finish-contract-run1.md <bytes1>` and
  `skills/flow-contracts/finish-contract-run2.md <bytes2>`, where `<bytes1>`/`<bytes2>` come from
  `wc -c` on the two new files after Steps 1-2.

  - [ ] **Step 6: Verify**

  Run: `scripts/check-references.sh && scripts/check-contract-budget.sh`
  Expected: both exit 0 — no dangling reference to the deleted `finish-contract.md`, and both new
  files sit under their newly added budget rows.

  - [ ] **Step 7: Commit**

```bash unverified:confirm exact paths against the worktree at execution time
git add skills/flow-contracts/finish-contract-run1.md skills/flow-contracts/finish-contract-run2.md \
  skills/flow/integrate.md skills/flow/archive.md skills/flow-contracts/worktree-resolution.md \
  skills/flow-contracts/jira-followups.md skills/flow-contracts/jira-integration.md \
  skills/flow-contracts/state-file.md skills/flow-contracts/artifacts-registry.md \
  skills/flow-contracts/project-configuration.md skills/flow-contracts/pipeline.md \
  skills/flow-contracts/SKILL.md scripts/check-contract-budget.sh
git rm skills/flow-contracts/finish-contract.md
git commit -m "docs(flow-contracts): split finish-contract.md by run"
```

---

- [x] 3. Split `project-configuration.md` into resolution rules and authoring guidance

**Build:** green

**Files:**
- Modify: `skills/flow-contracts/project-configuration.md`
- Create: `skills/flow-contracts/project-configuration-authoring.md`
- Modify: `skills/flow-contracts/SKILL.md`
- Modify: `scripts/check-contract-budget.sh`
- Modify: `scripts/check-references.sh`
- Modify: any file citing a section that moved (identified in Step 2 below — expected candidates:
  `skills/flow-contracts/project-configuration-rationale.md`, `skills/flow/review-panel.md`,
  `skills/flow/verify-and-handoff.md`, `skills/flow/archive.md`, `skills/flow/implement.md`)

**Ruling (added during implementation):** the moved content (the standards `.mdc`-routing
advisory, the workspace-isolation worked example, and the "which rules a script checks vs. left to
the agent" framing paragraph) was cited by no live citer outside `project-configuration.md` itself
— a full-repo `grep -rn "project-configuration.md" skills/ scripts/` confirmed every real citer
points at resolution content that stayed put, so none of the expected candidate files needed
editing. `scripts/check-references.sh` did, though: the new
`project-configuration-authoring.md` cites `project-configuration-rationale.md` by bold-adjacent
section name but has zero incoming citations, which its coverage check flags unless the file is
declared expected-zero — added to `EXPECTED_ZERO_RATIONALE_DOCS`, the same category the sibling
prose-only rationale docs already sit in.

**Tests:** **none** — prose relocation, no code
**Regression:** n/a — no test declared
**Baseline:** n/a — no test declared
**Commit:** `docs(flow-contracts): split project-configuration.md by audience`

  - [ ] **Step 1: Read the whole file and mark each passage's audience**

  Read `skills/flow-contracts/project-configuration.md` top to bottom. For each paragraph, mark it
  **resolution** (a run consults it while resolving a section: the `## <key>` table, the standards
  entry-form table, the containment-normalization rules, the workspace-isolation/visual-
  verification/citation-check key tables and their validation rules) or **authoring** (it only
  helps a human write or edit their own `.flow/project.md`: rationale for why a key is shaped the
  way it is, prose that no run branches on). Per design.md's `project-configuration-split-
  resolution-vs-authoring` decision, expect both to appear inside the same `##` section — do not
  assume a heading-level split.

  - [ ] **Step 2: Cut every authoring-marked passage into the new file**

  Create `skills/flow-contracts/project-configuration-authoring.md`, titled `# Project
  configuration — authoring guidance`, opening with one sentence stating it is for whoever edits a
  project's own `.flow/project.md`, cross-referencing `project-configuration.md` as canonical for
  what a run resolves. Move each authoring-marked passage there verbatim, in its original relative
  order. Leave every resolution-marked passage in `project-configuration.md`, in place.

  - [ ] **Step 3: Repoint any citer of a moved passage**

  Run `grep -rn "project-configuration.md" skills/flow/*.md skills/flow-contracts/*.md` and, for
  each hit, check whether the specific content it cites moved in Step 2; if so, repoint it to
  `project-configuration-authoring.md`.

  - [ ] **Step 4: Add the new file's row to `skills/flow-contracts/SKILL.md`'s index**

  Add a row for `project-configuration-authoring.md`, and note in `project-configuration.md`'s
  existing row that it now covers resolution rules only.

  - [ ] **Step 5: Update `scripts/check-contract-budget.sh`'s table**

  Lower `skills/flow-contracts/project-configuration.md`'s row to its new (smaller) actual size
  plus 25% (optional — the ratchet never requires lowering, only forbids exceeding; leave the
  existing 48175 row if the file still sits under it). Add a new row for
  `skills/flow-contracts/project-configuration-authoring.md` at its actual size plus 25%.

  - [ ] **Step 6: Verify**

  Run: `scripts/check-references.sh && scripts/check-contract-budget.sh`
  Expected: both exit 0.

  - [ ] **Step 7: Commit**

```bash unverified:confirm exact paths against the worktree at execution time
git add skills/flow-contracts/project-configuration.md \
  skills/flow-contracts/project-configuration-authoring.md skills/flow-contracts/SKILL.md \
  scripts/check-contract-budget.sh
git commit -m "docs(flow-contracts): split project-configuration.md by audience"
```

---

- [x] 4. Scope the dispatch bundle's lint/test/run commands instead of a full `.flow/project.md` re-read

**Build:** green

**Files:**
- Modify: `scripts/gather-dispatch-context.sh` (canonical source; `skills/flow/scripts/gather-dispatch-context.sh` is a symlink to it — do not edit the symlink target's contents directly)
- Modify: `scripts/test-gather-dispatch-context.sh`
- Modify: `skills/flow/implement.md`
- Modify: `scripts/check-contract-budget.sh`

**Ruling (added during implementation):** `skills/flow/review-panel.md` was declared but not
actually touched — its CONTEXT BUNDLE paragraph is a pointer ("the same one `skills/flow/implement.md`'s
implementer dispatch carries"), not a literal duplicate, so editing `implement.md`'s copy alone
propagates to panel dispatches; leaving the row in Files causes no guard issue (declared-but-untouched
is not a violation). `scripts/check-contract-budget.sh` was undeclared but required: the added
sentence pushed `implement.md` to 16701 bytes, over its 16559-byte budget row — raised to 20877
(actual × 1.25) per the guard's own ratchet rule for deliberate growth.

**Tests:** `scripts/test-gather-dispatch-context.sh` (existing harness, extended with cases for the
new `## lint`/`## test`/`## run` extraction: present in `.flow/project.md`, absent, and
`.flow/project.md` itself absent from the project)
**Regression:** reverting this task's script changes makes the new extraction test cases fail —
the bundle would carry no `## project commands` section at all
**Baseline:** before=44 after=47
<!-- measured: scripts/test-gather-dispatch-context.sh @ branch spectre/kan-372-flow-the-router-eager-loads-11-contracts-before -->
**Commit:** `feat(flow/scripts): extract lint/test/run commands into the dispatch bundle`

  - [ ] **Step 1: Run the existing test harness to establish the baseline count**

  Run: `scripts/test-gather-dispatch-context.sh`
  Expected: PASS, 44 cases green (already confirmed during planning).

  - [ ] **Step 2: Write the failing test cases**

  Add three cases to `scripts/test-gather-dispatch-context.sh`, following that file's
  existing fixture-and-assert pattern (a scratch dir with a `.flow/project.md` fixture, then a
  call to `gather-dispatch-context.sh` and an assertion on stdout):

  1. `.flow/project.md` declares `## lint` and `## test` sections → the bundle's output contains a
     `## project commands` heading whose body reproduces those two sections' fenced command blocks
     verbatim.
  2. `.flow/project.md` declares no `## lint`/`## test`/`## run` sections → the bundle reports
     `skipped: project commands (absent)`, the same shape the script already uses for
     `proposal.md`/`design.md` absence.
  3. The project has no `.flow/project.md` at all → same skipped line, not a script error.

  - [ ] **Step 3: Run the new cases and confirm they fail**

  Run: `scripts/test-gather-dispatch-context.sh`
  Expected: FAIL on the 3 new cases — `gather-dispatch-context.sh` does not yet emit a
  `## project commands` section.

  - [ ] **Step 4: Implement the extraction**

  In `scripts/gather-dispatch-context.sh`, add a fourth source alongside the existing
  three (`proposal.md`, `design.md`, `tasks.md`): resolve `<worktree>/.flow/project.md` (or the
  project's own root — reuse whatever path resolution the script already applies for its
  containment checks), and when present, extract the `## lint`, `## test`, and `## run` sections
  (bounded by the next `##` heading or EOF) into a `## project commands` block in the bundle,
  appended after the principles file section. When `.flow/project.md` is absent, or none of the
  three keys are present, emit the existing `skipped: <src> (absent)` line instead — no new bundle
  section.

  - [ ] **Step 5: Run the tests and confirm they pass**

  Run: `scripts/test-gather-dispatch-context.sh`
  Expected: PASS, 47 cases green (44 existing + 3 new).

  - [ ] **Step 6: Update the CONTEXT BUNDLE dispatch paragraph**

  In `skills/flow/implement.md` and `skills/flow/review-panel.md`, find the "> **CONTEXT BUNDLE:**"
  blockquote paragraph (verbatim-identical in both files) and append one sentence to it: "It also
  carries this project's `## lint`/`## test`/`## run` commands, already resolved — you do not need
  to open `.flow/project.md` yourself for them." Leave the rest of the paragraph's existing wording
  unchanged. `scripts/check-dispatch-paragraphs.sh` does not track this paragraph (confirmed: its
  `SITE_ENTRY` table covers only `reproduce`, `verbatim`, and `foreground`), so this edit needs no
  guard-table update.

  - [ ] **Step 7: Verify the guards**

  Run: `scripts/check-references.sh && scripts/check-dispatch-paragraphs.sh`
  Expected: both exit 0 — the CONTEXT BUNDLE paragraph stays identical between the two files
  (`check-dispatch-paragraphs.sh` does not check it, but keep them byte-identical anyway, since
  every other dispatch paragraph in this pair is).

  - [ ] **Step 8: Commit**

```bash unverified:confirm exact paths against the worktree at execution time
git add scripts/gather-dispatch-context.sh scripts/test-gather-dispatch-context.sh \
  skills/flow/implement.md skills/flow/review-panel.md
git commit -m "feat(flow/scripts): extract lint/test/run commands into the dispatch bundle"
```

---

- [x] 5. Shrink `jira-followups.md` — cut restating passages

**Build:** green

**Files:**
- Modify: `skills/flow-contracts/jira-followups.md`

**Tests:** **none** — prose cut, no code
**Regression:** n/a — no test declared
**Baseline:** n/a — no test declared
**Commit:** `docs(flow-contracts): cut restating passages from jira-followups.md`

  - [ ] **Step 1: Read the whole file against the rest of the flow-contracts tree**

  Read `skills/flow-contracts/jira-followups.md` in full. For each passage, check whether the same
  fact is already stated canonically elsewhere in `skills/flow-contracts/` or `skills/flow/`
  (most likely candidates: `jira-integration.md` for transition/label mechanics,
  `finish-contract-run1.md` for the unfinished-work-gate courses this file's "File or join" option
  feeds into). List every restating passage found, with its own location and the canonical
  location it restates.

  - [ ] **Step 2: Cut each restating passage, replacing it with a citation**

  For each passage from Step 1, delete it and, where the surrounding prose needs the fact to read
  coherently, replace it with a short citation to the canonical location (`per **<Section>**
  (`<path>`)`) rather than a reworded restatement. Never touch a normative sentence, exit-code
  contract, scenario, or rejected-alternative rationale that is NOT a restatement of something
  stated elsewhere — those stay untouched.

  - [ ] **Step 3: Verify**

  Run: `scripts/check-references.sh`
  Expected: exit 0 — every new citation resolves to a real heading.
  Run: `wc -c skills/flow-contracts/jira-followups.md`
  Expected: smaller than 36284 (this file's size at `18a89b5`).

  - [ ] **Step 4: Commit**

```bash unverified:confirm exact paths against the worktree at execution time
git add skills/flow-contracts/jira-followups.md
git commit -m "docs(flow-contracts): cut restating passages from jira-followups.md"
```

---

- [x] 6. Shrink `review-panel.md` — cut restating passages

**Build:** green

**Files:**
- Modify: `skills/flow/review-panel.md`

**Tests:** **none** — prose cut, no code
**Regression:** n/a — no test declared
**Baseline:** n/a — no test declared
**Commit:** `docs(flow): cut restating passages from review-panel.md`

  - [ ] **Step 1: Read the whole file against `implement.md` and `pipeline.md`**

  Read `skills/flow/review-panel.md` in full, looking especially for dispatch-recording mechanics
  and operator-prompt shapes already stated once in `implement.md` (the dispatch-record pair
  format) or `skills/flow-contracts/operator-prompts.md` (prompt shape). List every restating
  passage found.

  - [ ] **Step 2: Cut each restating passage, replacing it with a citation**

  Same rule as Task 5, Step 2: delete, cite, never reword.

  - [ ] **Step 3: Verify**

  Run: `scripts/check-references.sh && scripts/check-dispatch-paragraphs.sh`
  Expected: both exit 0 — cutting must not touch any of the three guarded dispatch paragraphs
  (`reproduce`, `verbatim`, `foreground`) this file carries.
  Run: `wc -c skills/flow/review-panel.md`
  Expected: smaller than 33195 (this file's size at `18a89b5`), and under its 35985-byte budget
  row regardless.

  - [ ] **Step 4: Commit**

```bash unverified:confirm exact paths against the worktree at execution time
git add skills/flow/review-panel.md
git commit -m "docs(flow): cut restating passages from review-panel.md"
```

---

- [x] 7. Shrink `pipeline.md` — cut restating passages

**Build:** green

**Files:**
- Modify: `skills/flow-contracts/pipeline.md`

**Tests:** **none** — prose cut, no code
**Regression:** n/a — no test declared
**Baseline:** n/a — no test declared
**Commit:** `docs(flow-contracts): cut restating passages from pipeline.md`

  - [ ] **Step 1: Read the whole file against `skills/flow/SKILL.md` and `handoff-blocks.md`**

  Read `skills/flow-contracts/pipeline.md` in full (already touched once, in Task 2 Step 4, for
  its finish-contract citations — re-read the whole file this time). List every passage restating
  content this task's own earlier Task 2 changes, or `handoff-blocks.md`, already state
  canonically.

  - [ ] **Step 2: Cut each restating passage, replacing it with a citation**

  Same rule as Task 5, Step 2.

  - [ ] **Step 3: Verify**

  Run: `scripts/check-references.sh`
  Expected: exit 0.
  Run: `wc -c skills/flow-contracts/pipeline.md`
  Expected: smaller than 28766 (this file's size at `18a89b5`), and under its 36155-byte budget
  row regardless.

  - [ ] **Step 4: Commit**

```bash unverified:confirm exact paths against the worktree at execution time
git add skills/flow-contracts/pipeline.md
git commit -m "docs(flow-contracts): cut restating passages from pipeline.md"
```

---

- [x] 8. Shrink `state-file.md` — cut restating passages

**Build:** green

**Files:**
- Modify: `skills/flow-contracts/state-file.md`

**Tests:** **none** — prose cut, no code
**Regression:** n/a — no test declared
**Baseline:** n/a — no test declared
**Commit:** `docs(flow-contracts): cut restating passages from state-file.md`

  - [ ] **Step 1: Read the whole file against `jira-integration.md`, `finish-contract-run1.md`, and
  `finish-contract-run2.md`**

  Read `skills/flow-contracts/state-file.md` in full. List every passage restating content those
  three files (this task's own earlier Task 2 outputs included) already state canonically.

  - [ ] **Step 2: Cut each restating passage, replacing it with a citation**

  Same rule as Task 5, Step 2.

  - [ ] **Step 3: Verify**

  Run: `scripts/check-references.sh`
  Expected: exit 0.
  Run: `wc -c skills/flow-contracts/state-file.md`
  Expected: smaller than 27931 (this file's size at `18a89b5`).

  - [ ] **Step 4: Commit**

```bash unverified:confirm exact paths against the worktree at execution time
git add skills/flow-contracts/state-file.md
git commit -m "docs(flow-contracts): cut restating passages from state-file.md"
```

---

- [x] 9. Re-run the measurement and confirm the guard suite

**Build:** green

**Files:**
- Modify: `spectre/changes/kan-372-flow-the-router-eager-loads-11-contracts-before/proposal.md`
  (append the before/after measurement — never committed by this or any task commit; folded into
  the integrate phase's own planning-artifacts commit instead, per **Git boundaries**)
- Modify: `skills/flow-contracts/SKILL.md`
- Modify: `skills/flow/implement.md`

**Ruling (added during implementation):** the guard suite run in Step 3 found
`check-installed-citations.sh` failing on two pre-existing regressions from Tasks 3 and 4 — both
cited `.flow/project.md` without the `<project>/` root prefix the guard requires
(`skills/flow-contracts/SKILL.md`'s new `project-configuration-authoring.md` index row, and
`skills/flow/implement.md`'s CONTEXT BUNDLE paragraph addition). Fixed both (added the missing
prefix, no other change) rather than leaving the change unable to pass its own guard suite; this is
the actual commit Task 9 makes — the `proposal.md` measurement append stays uncommitted, per
**Git boundaries**, since planning artifacts are never part of a task commit.

**Tests:** **none** — measurement and guard run, no code
**Regression:** n/a — no test declared
**Baseline:** n/a — no test declared
**Commit:** `docs(flow-contracts): fix two .flow/project.md citations missing their root prefix`

  - [ ] **Step 1: Re-run the router eager-load measurement**

  Reproduce the ticket's own method: sum the byte size of `skills/flow/SKILL.md` plus every
  contract it still eagerly names (after Task 1, only `pipeline.md`), and separately the byte size
  of a full creating-run load set and a full integrate-run load set (SKILL.md + brainstorm.md +
  implement.md + review-panel.md + verify-and-handoff.md + engineering-principles.md +
  principles-reviewer-prompt.md for creating; SKILL.md + integrate.md + archive.md +
  finish-contract-run1.md + finish-contract-run2.md + operator-prompts.md for integrate), each
  divided by 4 for an approximate token count, matching the ticket's own methodology.

  - [ ] **Step 2: Record before/after totals**

  Append a `## Measurement` section to this change's `proposal.md`, with a table matching the
  ticket description's shape (Load set / bytes / ~tokens), one row per set, before (from the
  ticket's own `18a89b5` figures) and after (from Step 1).

  - [ ] **Step 3: Run the full guard suite**

  Run: `for g in scripts/check-*.sh; do echo "=== $g ==="; "$g" || echo "FAILED: $g"; done`
  Expected: every guard reports success; no `FAILED:` line.

  - [ ] **Step 4: Commit**

```bash unverified:confirm exact paths against the worktree at execution time
git add spectre/changes/kan-372-flow-the-router-eager-loads-11-contracts-before/proposal.md
git commit -m "docs(spectre): record kan-372 before/after load measurement"
```
