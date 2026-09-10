# kan-490-flow-fast-a-reduced-ceremony-flow-variant-drop

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

Ten tasks, dependency order. `design.md` is canonical for every decision, table and rule a task
implements — a task names the section it implements and what the edit must say, never a second
copy of it. Tasks 2–6 create `skills/flow-fast/`; tasks 7–9 are edits to existing `/flow` files;
task 10 is the plan-class threshold raise, independent of the fast skill itself
(`design.md`'s `flow-fast-plan-class-thresholds`).

**Baseline, measured before any edit, on the clean branch:**

- `scripts/check-vocabulary.sh`, `scripts/check-markdown-integrity.py`,
  `scripts/check-contract-budget.sh` (74 owned Markdown files within budget),
  `scripts/check-installed-citations.sh` (79 files scanned), `scripts/check-installed-rules.sh`,
  `scripts/check-normative-inventory.sh` (8 normative sentences from 74 files),
  `scripts/check-guard-symlinks.sh` (122 guards across 4 skills), `scripts/check-stage-mark-calls.sh`
  (48 call sites), `scripts/check-dispatch-paragraphs.sh` (20 sites), `scripts/check-model-keys.sh`
  and `scripts/check-model-resolution-shell.sh` (9 cases) all exit 0.
  <!-- measured: each script run from the worktree root @ branch spectre/kan-490-flow-fast-a-reduced-ceremony-flow-variant-drop -->
- `scripts/check-references.sh` exits 1 on the clean branch already:
  `skills/flow/verify-and-handoff.md:291: no bold token resolves to a heading in
  skills/flow-contracts/project-configuration.md`. Pre-existing, unrelated to every file this plan
  touches — no task here fixes it, and no task's verify step re-asserts it clean.
  <!-- measured: scripts/check-references.sh @ branch spectre/kan-490-flow-fast-a-reduced-ceremony-flow-variant-drop -->
- Byte sizes against `scripts/check-contract-budget.sh`'s `budgets()` rows, all with headroom for
  every edit below (edits shrink two of these): `skills/flow/SKILL.md` 18,687 of 20,520;
  `skills/flow-contracts/pipeline.md` 27,048 of 36,155; `skills/flow-contracts/pipeline-rationale.md`
  14,925 of 20,935; `skills/flow-contracts/finish-contract-run2.md` 33,347 of 36,019;
  `skills/flow/brainstorm-planner.md` 29,217 of 31,664; `CLAUDE.md` 6,154 of 15,195.
  <!-- measured: wc -c on each file and the budgets() rows at scripts/check-contract-budget.sh @ branch spectre/kan-490-flow-fast-a-reduced-ceremony-flow-variant-drop -->
- `scripts/plan-class.sh` 4,427 bytes; `scripts/test-plan-class.sh` 7,460 bytes; `scripts/run-guard-tests.sh`
  discovers 59 `scripts/test-*.sh` harnesses.
  <!-- measured: wc -c on each file; ls scripts/test-*.sh | wc -l @ branch spectre/kan-490-flow-fast-a-reduced-ceremony-flow-variant-drop -->

**Every task that adds a new owned `.md` file (tasks 2–6) adds that file's row to
`budgets()` in `scripts/check-contract-budget.sh`, in the same commit** — measure the file's
actual byte count with `wc -c` after writing it, and record its budget row as that count plus 25%
headroom, sorted into the table alongside its sibling `skills/flow-fast/…` / `commands/…` /
`commands-claude/…` rows (the table is otherwise unsorted across sections, per the existing rows —
match the nearest existing block's ordering). This is not a placeholder: the number is measured at
commit time, never invented ahead of it.

**Every prose task's verify step names the guards that scan owned Markdown**
(`check-vocabulary.sh`, `check-references.sh` — informational only, per the pre-existing failure
above — `check-markdown-integrity.py`, `check-contract-budget.sh`) **plus whichever of
`check-installed-citations.sh`, `check-guard-symlinks.sh`, `check-stage-mark-calls.sh`,
`check-dispatch-paragraphs.sh`, `check-model-keys.sh`, `check-model-resolution-shell.sh` and
`check-normative-inventory.sh` its own edit could plausibly trip** — never the project's whole
`## lint` list. No task here adds a test file: every edit is Markdown or a Bash guard script, and
`scripts/test-plan-class.sh` (task 10) is the one test file this plan touches, updated alongside
the script it tests.

---

- [x] 1. Delete the adopted staging note

`docs/superpowers/research/flow-fast.md` seeded this change's brainstorm (`design.md`'s Context).
Its content is now folded into `proposal.md`, `design.md` and this file — delete it.

  - [x] **Step 1: Delete** `docs/superpowers/research/flow-fast.md`.
  - [x] **Step 2: Verify** — `scripts/check-vocabulary.sh` and `scripts/check-references.sh` from
    the worktree root; both exit the same as the pre-edit baseline (reference guard still shows
    only the one pre-existing, unrelated hit).

**Files:** `docs/superpowers/research/flow-fast.md`
**Tests:** none — deletion of an adopted staging note; the verify step's guard scripts are the check
**Regression:** reverting this commit leaves a stale staging note in the research tree after its
content has already landed in this change's own artifacts, which is what the "delete once adopted"
rule (`skills/flow/brainstorm-planner.md` section B) exists to prevent.
**Baseline:** before=0 after=0 — no test file changes
**Commit:** `docs(flow-research): delete the adopted flow-fast staging note`
**Build:** green

- [x] 2. `skills/flow-fast/SKILL.md` — the router

Per `design.md`'s `flow-fast-shape`, `flow-fast-state-interop` and `flow-fast-model-resolution`.
New file, modeled on `skills/flow/SKILL.md`'s own shape but slimmed to what a fixed-decision
command needs — do not copy `skills/flow/SKILL.md`'s Model resolution block, Stage keys table, or
Guardrails wholesale; restate only what differs.

Content: an announce line ("Using flow-fast for change `<name>`.") printing **no** `/rename` /
`/color` lines (`design.md`'s `flow-fast-drop-rename-color` — `/flow-fast` never had them, nothing
to remove); "**Load `skills/flow-contracts/pipeline.md` first**" for the three states and the
transition table (cited, never restated); task-list registration per **Progress visibility**
(`skills/flow-contracts/pipeline.md`), at the same per-stage granularity `skills/flow/SKILL.md`
uses since `/flow-fast` runs everything inline (no conductor, so no coarser conductor-stage
granularity applies — every stage's own steps register); a stage-key table naming exactly
`flow.kickoff`, `flow.brainstorm`, `flow.create-artifacts`, `flow.writing-plans`, `flow.decide`
(brainstorm.md), `flow.load-context`, `flow.isolate-workspace`, `flow.document-fix`, `flow.sdd-tdd`
(implement.md), `flow.review-panel` (review.md), `flow.verify`, `flow.stage-diff`,
`flow.run-instructions`, `flow.write-in-progress` (review.md — no `flow.visual-verify`, `/flow-fast`
runs no visual verification per the note), `flow.preflight`, `flow.unfinished-work-gate`,
`flow.landing-question`, `flow.preserve-sessions`, `flow.commit-two`, `flow.landing-routes`
(finish.md, run 1), `flow.verify-merge`, `flow.sync-archive`, `flow.commit-archive`, `flow.cleanup`,
`flow.write-finished`, `flow.push-archive` (finish.md, run 2 — **no** `flow.verify-cleanup` or
`flow.self-review`, both skipped per `design.md`'s `flow-fast-finish`) — every mark carries
`-command '/flow-fast'`; a "Reading the state" section identical in structure to `skills/flow/SKILL.md`'s
own (same `flow state get` dispatch table, same exit-1/STARTED/IN_PROGRESS-with-argument/
IN_PROGRESS-bare/FINISHED rows, cited against this file's own brainstorm.md / implement.md /
review.md / finish.md instead of `/flow`'s); a "Model resolution" section resolving only
`DEFAULT_MODEL` via `flow settings get`, the same unreachable-store fallback to the literal
`sonnet` `skills/flow/SKILL.md`'s block uses, and stating explicitly that no `PLANNING_MODEL`,
`SELF_REVIEW_MODEL`, `VERIFY_MODEL` or project toggle (`## execution mode`, `## implementer
model`, `## review panel`) is read; a "Guard presence check" line citing
`skills/flow-contracts/pipeline.md`'s own section, restricted to the seven guards named in
`design.md`'s `flow-fast-guards`; a session-token-once-per-run rule identical to `skills/flow/SKILL.md`'s.

  - [x] **Step 1: Write** `skills/flow-fast/SKILL.md` per the content above.
  - [x] **Step 2: Add its budget row** to `scripts/check-contract-budget.sh`'s `budgets()`.
  - [x] **Step 3: Verify** — `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
    `scripts/check-markdown-integrity.py`, `scripts/check-contract-budget.sh`,
    `scripts/check-stage-mark-calls.sh`, `scripts/check-guard-symlinks.sh`,
    `scripts/check-model-keys.sh`, `scripts/check-model-resolution-shell.sh` all exit 0.

**Files:** `skills/flow-fast/SKILL.md`, `scripts/check-contract-budget.sh`,
`scripts/check-stage-mark-calls.sh`, `skills/flow-fast/scripts/check-base-moved.sh`,
`skills/flow-fast/scripts/check-cleanup-complete.sh`, `skills/flow-fast/scripts/check-finish-preflight.sh`,
`skills/flow-fast/scripts/check-panel-findings-closed.sh`, `skills/flow-fast/scripts/check-unfinished-work.sh`,
`skills/flow-fast/scripts/check-workspace-isolation.sh`, `skills/flow-fast/scripts/check-worktree-processes.sh`,
`skills/flow-fast/scripts/lib`, `skills/flow-fast/scripts/prepare-workspace.sh`
**Tests:** none — new Markdown skill file; the verify step's guard scripts are the check
**Regression:** reverting this commit removes `/flow-fast`'s router entirely — no other task
depends on it existing to compile or run, but `commands/flow-fast.md` (task 3) cites it by name.
**Baseline:** before=0 after=0 — no test file changes
**Commit:** `feat(flow-fast): add the /flow-fast router skill`
**Build:** green

- [x] 3. `commands/flow-fast.md` and `commands-claude/flow-fast.md`

Per `design.md`'s `flow-fast-shape`. New files, one per command tree, each modeled on that tree's
own `flow.md` (read it first) but describing `/flow-fast`'s fixed behavior instead of `/flow`'s
dynamic one: no planning-effort/model/review-panel-roster question ever (nothing to ask — every
choice is fixed by design, not resolved from a settings store); brainstorm is inline with no
planner subagent and auto-picks every recommended option, skipping the design-approval gate;
implementation is inline with no conductor; the review panel is always `primary` +
`simple-reviewer`; ending at `IN_PROGRESS` exactly as `/flow`'s own command file states. Use the
**flow-fast** skill — installed globally, resolved by name. Same **Input**/**When done** shape as
`flow.md`'s own command file, restated for `/flow-fast`'s own commands.

  - [x] **Step 1: Write** `commands/flow-fast.md` (Claude Code's own frontmatter shape,
    `commands/flow.md`'s `---name:/id:/category:/description:---` block, `id: flow-fast`).
  - [x] **Step 2: Write** `commands-claude/flow-fast.md` (`commands-claude/flow.md`'s
    `---model:/description:---` block).
  - [x] **Step 3: Add both files' budget rows** to `scripts/check-contract-budget.sh`.
  - [x] **Step 4: Verify** — `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
    `scripts/check-markdown-integrity.py`, `scripts/check-contract-budget.sh`,
    `scripts/check-installed-citations.sh` all exit 0.

**Files:** `commands/flow-fast.md`, `commands-claude/flow-fast.md`, `scripts/check-contract-budget.sh`
**Tests:** none — new command-dispatch stubs; the verify step's guard scripts are the check
**Regression:** reverting this commit leaves `/flow-fast` with no invocable command in either
tree — the skill (task 2) would exist but nothing routes a slash command to it.
**Baseline:** before=0 after=0 — no test file changes
**Commit:** `feat(flow-fast): add the /flow-fast command stubs`
**Build:** green

- [x] 4. `skills/flow-fast/brainstorm.md`

Per `design.md`'s `flow-fast-brainstorm`. New file. Resolve the linked Jira issue and transition it
to In Progress exactly as `skills/flow/brainstorm.md` section **A** does, citing
`skills/flow-contracts/jira-integration.md` rather than restating it; write `STARTED` immediately
after the name resolves, same JSON shape `skills/flow/brainstorm.md` shows. Then: no planner
dispatch — the parent session itself runs `superpowers:brainstorming`'s checklist inline, auto-
picking the recommended option at every round; ask the operator only when no option is recommended
or the round cannot proceed without an answer. Create the change worktree (same
`check-worktree-location.sh` / `git check-ignore` / `git worktree add` / `## worktree setup`
sequence `skills/flow/brainstorm.md`'s "The three returns" runs, cited rather than copied), run
`spectre new`, write `proposal.md`/`design.md`/`tasks.md` in `skills/flow/brainstorm-planner.md`
section C's exact shape (`## Decisions`/`## Open questions` included), delete an adopted staging
note per that same section. Run `skills/flow/brainstorm-planner.md` section D's writing-plans
enrichment (exact paths, verification commands, `**Files:**`/`**Tests:**`/`**Regression:**`/
`**Baseline:**`/`**Commit:**`/`**Build:**` fields, the plan-provenance tagging rule) — but skip its
**Decide** roll entirely: instead, write `<abs-worktree>/.superpowers/sdd/decision.json` with the
fixed decision `design.md`'s `flow-fast-state-interop` names (`execution: inline`, roster `primary`
+ `simple-reviewer` on `DEFAULT_MODEL` + haiku, no groups, no rolls — a literal object, not a roll
result) and record it with `flow record decision`, so a `/flow` run resuming this change reads a
valid record. Mark `flow.kickoff`, `flow.brainstorm`, `flow.create-artifacts`, `flow.writing-plans`,
`flow.decide` — never `flow.design-approval`.

  - [x] **Step 1: Write** `skills/flow-fast/brainstorm.md` per the content above.
  - [x] **Step 2: Add its budget row** to `scripts/check-contract-budget.sh`.
  - [x] **Step 3: Verify** — `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
    `scripts/check-markdown-integrity.py`, `scripts/check-contract-budget.sh`,
    `scripts/check-stage-mark-calls.sh`, `scripts/check-guard-symlinks.sh` all exit 0.

**Files:** `skills/flow-fast/brainstorm.md`, `scripts/check-contract-budget.sh`,
`skills/flow-fast/scripts/project-get.sh`
**Tests:** none — new Markdown skill file; the verify step's guard scripts are the check
**Regression:** reverting this commit removes `/flow-fast`'s only path to a `STARTED`/`IN_PROGRESS`
transition — a creating `/flow-fast` run would dispatch into a missing file.
**Baseline:** before=0 after=0 — no test file changes
**Commit:** `feat(flow-fast): add the inline auto-pick brainstorm phase`
**Build:** green

- [x] 5. `skills/flow-fast/implement.md`

Per `design.md`'s `flow-fast-implement`. New file. `prepare-workspace.sh <worktree>` (which itself
runs `check-workspace-isolation.sh` first — cite `skills/flow/verify-and-handoff.md`'s own
paragraph on this rather than restating it) once per resolved worktree; then, inline in the parent
(no conductor, no implementer subagent dispatch), task by task from `tasks.md`: write the failing
test first, make it pass, run that task's own touched-package tests via the build tool's own
selector (`-run '<name>'` for `go test`, `-t '<name>'` for vitest, etc. — never a bare module or
repository suite) plus lint on the task's own touched files only, then one commit per task carrying
the exact commit-field shape `skills/flow/brainstorm-planner.md` section D requires
(`**Files:**`/`**Tests:**`/`**Regression:**`/`**Baseline:**`/`**Commit:**`/`**Build:**` —
`check-task-commit-fields.sh` itself is never run, per `design.md`'s `flow-fast-guards`, but the
format is kept so a later `/flow` resume or manual audit still reads it). State explicitly: a full
`## test` / `## lint` run happens only when the operator's own instruction text asks for it, never
automatically and never before handoff by default. No SDD ceremony — no spec-delta guard, no
plan-provenance guard invocation — though a task naming a `spectre/specs/` path in its own
`**Files:**` still gets that spec edit written and committed in that task's own commit. Mark
`flow.load-context`, `flow.isolate-workspace`, `flow.sdd-tdd` per task; `flow.document-fix` only on
a fix run at `IN_PROGRESS` (an appended-plan re-entry, same shape `skills/flow/implement.md`'s own
`flow.document-fix` uses, cited rather than copied).

  - [x] **Step 1: Write** `skills/flow-fast/implement.md` per the content above.
  - [x] **Step 2: Add its budget row** to `scripts/check-contract-budget.sh`.
  - [x] **Step 3: Verify** — `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
    `scripts/check-markdown-integrity.py`, `scripts/check-contract-budget.sh`,
    `scripts/check-stage-mark-calls.sh`, `scripts/check-guard-symlinks.sh` all exit 0.

**Files:** `skills/flow-fast/implement.md`, `scripts/check-contract-budget.sh`
**Tests:** none — new Markdown skill file; the verify step's guard scripts are the check
**Regression:** reverting this commit removes `/flow-fast`'s only implementation phase.
**Baseline:** before=0 after=0 — no test file changes
**Commit:** `feat(flow-fast): add the inline TDD implementation phase`
**Build:** green

- [x] 6. `skills/flow-fast/review.md`

Per `design.md`'s `flow-fast-review`. New file. Dispatch exactly two subagents, bundled as one
`-slot primary+simple-reviewer` dispatch (`flow record dispatch begin/end` per
`skills/flow/review-panel.md`'s own dispatch-record shape, cited rather than copied): `primary` on
`DEFAULT_MODEL`, briefed against `final-review.diff`, `proposal.md`, `design.md` and each task's
commit fields in `tasks.md` — plan alignment only, never code quality (`skills/flow/review-panel.md`'s
own **The roster** row for `primary`, cited); `simple-reviewer` on `haiku`, briefed by
`skills/flow/simple-reviewer-prompt.md` for high-confidence defects only. Never consult the
settings-store reviewer list or the docs-only reduction — the roster is fixed. Findings recorded
with `flow record finding`; Critical and Major fixed inline by the parent (no conductor, no
panel-fix subagent — the parent itself edits, per `design.md`'s `flow-fast-implement`'s "inline"
choice extended to fixes); every Minor recorded and deferred, never fixed, no per-finding judgment.
After a fix round, both slots re-run on that round's delta diff (same per-slot-per-worktree
last-reviewed-sha mechanism `skills/flow/review-panel.md`'s **Panel re-runs** describes, cited) until
`check-panel-findings-closed.sh <worktree> <change>` exits 0. Then: stage the diff, print run
instructions and the `IN_PROGRESS` block per `skills/flow-contracts/handoff-blocks.md`'s own shape
(cited, never copied), write `IN_PROGRESS`. No visual verification — no `flow.visual-verify` mark,
no visual-trigger or visual-verification guard invocation. Mark `flow.review-panel`, `flow.verify`,
`flow.stage-diff`, `flow.run-instructions`, `flow.write-in-progress`.

  - [x] **Step 1: Write** `skills/flow-fast/review.md` per the content above.
  - [x] **Step 2: Add its budget row** to `scripts/check-contract-budget.sh`.
  - [x] **Step 3: Verify** — `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
    `scripts/check-markdown-integrity.py`, `scripts/check-contract-budget.sh`,
    `scripts/check-stage-mark-calls.sh`, `scripts/check-guard-symlinks.sh` all exit 0.

**Files:** `skills/flow-fast/review.md`, `scripts/check-contract-budget.sh`,
`scripts/check-stage-mark-calls.sh`
**Tests:** none — new Markdown skill file; the verify step's guard scripts are the check
**Regression:** reverting this commit removes `/flow-fast`'s only review/handoff phase — no path
from implementation to a staged, reviewed `IN_PROGRESS` diff.
**Baseline:** before=0 after=0 — no test file changes
**Commit:** `feat(flow-fast): add the fixed two-slot review and handoff phase`
**Build:** green

- [x] 7. `skills/flow-fast/finish.md`

Per `design.md`'s `flow-fast-finish`. New file. Run 1: follow
`skills/flow-contracts/finish-contract-run1.md` exactly as `skills/flow/integrate.md` does — the
unfinished-work gate (`check-unfinished-work.sh`), the landing question (or a project's `##
default landing route`), session preservation, the two commits, the landing routes — citing that
contract and `skills/flow/integrate.md`'s own structure rather than restating either. Run 2: follow
`skills/flow-contracts/finish-contract-run2.md` exactly as `skills/flow/archive.md` does, through
step 8 (write `FINISHED`), then **skip step 9 (self-review) and step 7 (verify-cleanup) outright**
per that contract's own `/flow-fast` exemption (task 8 adds it) — cleanup itself (step 5) still
runs — then step 10 (push and land) and step 11 (remove the landing worktree). Every mark in both
runs carries `-command '/flow-fast'` and marks exactly: `flow.preflight`,
`flow.unfinished-work-gate`, `flow.landing-question`, `flow.preserve-sessions`, `flow.commit-two`,
`flow.landing-routes` (run 1); `flow.verify-merge`, `flow.sync-archive`, `flow.commit-archive`,
`flow.cleanup`, `flow.write-finished`, `flow.push-archive` (run 2) — never `flow.verify-cleanup` or
`flow.self-review`.

  - [x] **Step 1: Write** `skills/flow-fast/finish.md` per the content above.
  - [x] **Step 2: Add its budget row** to `scripts/check-contract-budget.sh`.
  - [x] **Step 3: Verify** — `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
    `scripts/check-markdown-integrity.py`, `scripts/check-contract-budget.sh`,
    `scripts/check-stage-mark-calls.sh`, `scripts/check-guard-symlinks.sh` all exit 0.

**Files:** `skills/flow-fast/finish.md`, `scripts/check-contract-budget.sh`
**Tests:** none — new Markdown skill file; the verify step's guard scripts are the check
**Regression:** reverting this commit removes `/flow-fast`'s only path from a staged
`IN_PROGRESS` diff to `FINISHED`.
**Baseline:** before=0 after=0 — no test file changes
**Commit:** `feat(flow-fast): add the finish phase, skipping self-review and verify-cleanup`
**Build:** green

- [x] 8. `finish-contract-run2.md` gains the `/flow-fast` exemption

Per `design.md`'s `flow-fast-finish`. Edit `skills/flow-contracts/finish-contract-run2.md`: at the
end of step 7's paragraph (the one beginning "**Verify the cleanup.**", currently ending
"'not verified' is never reported as verified." at the "When the script is absent" sub-paragraph),
add one sentence: "**A `/flow-fast` run skips this step entirely** — cleanup itself (step 5) still
runs; only its separate verification pass does not, per that command's own reduced guard set." At
the end of step 9's opening sentence (the one beginning "**Run self-review** — after `FINISHED` is
written"), add: "**A `/flow-fast` run skips this step entirely** — no self-review subagent is
dispatched, per that command's own design." Touch no other line in either step; do not renumber.

  - [x] **Step 1: Add the step-7 sentence** at the end of the paragraph identified above.
  - [x] **Step 2: Add the step-9 sentence** at the end of the opening sentence identified above.
  - [x] **Step 3: Verify** — `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
    `scripts/check-markdown-integrity.py`, `scripts/check-contract-budget.sh` all exit 0.

**Files:** `skills/flow-contracts/finish-contract-run2.md`
**Tests:** none — prose edit; the verify step's guard scripts are the check
**Regression:** reverting this commit leaves `finish-contract-run2.md` silent about `/flow-fast`,
so a careful reader of `skills/flow-fast/finish.md` (task 7) would see a step it names skipping
without the contract itself ever saying that exemption exists.
**Baseline:** before=0 after=0 — no test file changes
**Commit:** `docs(flow-contracts): exempt /flow-fast from self-review and verify-cleanup`
**Build:** green

- [x] 9. Drop `/rename` and `/color`; add the `flow-fast` skill-index row

Per `design.md`'s `flow-fast-drop-rename-color` and `flow-fast-shape`. Three edits:

`skills/flow/SKILL.md` — delete the fenced block
```text
/rename <change-name>
/color cyan
```
and the sentence introducing it ("Immediately after that line, print these two commands for the
operator to paste, per **Handoff output** … that section fixes the colour and records why they are
printed rather than invoked:"), leaving the announce-line instruction in place.

`skills/flow-contracts/pipeline.md` — delete the same fenced block and its introducing sentence
under **Handoff output**, and delete the entire **The tab commands, printed at the start of a run**
subsection (from its `### The tab commands, printed at the start of a run` heading through the
paragraph ending "…so the naming happens at the start of the run: a command that silently skips it
because its harness offers no tool has dropped the requirement, not adapted it.", immediately
before **Artifact brevity**).

`skills/flow-contracts/pipeline-rationale.md` — delete the entire `### The tab commands, printed at
the start of a run` section (lines 120–144 on the clean branch, from that heading through "…so the
next reader neither repeats the investigation nor treats the printing as an oversight to correct:"
and its two closing bullets), immediately before `## Artifact brevity`.

`CLAUDE.md` — add one row to the **Skill index** table, alongside the existing `skills/flow/`,
`skills/flow-status/`, `skills/flow-research/`, `skills/flow-settings/`, `skills/flow-contracts/`
rows: `| \`skills/flow-fast/\` | \`/flow-fast\` | Reduced-ceremony /flow variant: inline
brainstorm with auto-pick and no design gate, inline TDD implementation with targeted-only
tests/lint, a fixed primary+simple-reviewer panel, seven guards, and the same finish contracts
minus self-review and verify-cleanup. Same state record and \`flow.*\` stage keys as \`/flow\` |`.

  - [x] **Step 1: Edit** `skills/flow/SKILL.md` per the deletion above.
  - [x] **Step 2: Edit** `skills/flow-contracts/pipeline.md` per both deletions above.
  - [x] **Step 3: Edit** `skills/flow-contracts/pipeline-rationale.md` per the deletion above.
  - [x] **Step 4: Edit** `CLAUDE.md` — add the skill-index row.
  - [x] **Step 5: Verify** — `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
    `scripts/check-markdown-integrity.py`, `scripts/check-contract-budget.sh`,
    `scripts/check-installed-citations.sh` all exit 0.

**Files:** `skills/flow/SKILL.md`, `skills/flow-contracts/pipeline.md`,
`skills/flow-contracts/pipeline-rationale.md`, `CLAUDE.md`
**Tests:** none — prose edit; the verify step's guard scripts are the check
**Regression:** reverting this commit brings back the deprecated `/rename`/`/color` printed lines
on every `/flow` run, and drops `/flow-fast` from `CLAUDE.md`'s skill index (it would still work,
just be undiscoverable from the project's own instruction file).
**Baseline:** before=0 after=0 — no test file changes
**Commit:** `docs(flow): drop the /rename and /color handoff lines; index flow-fast`
**Build:** green

- [x] 10. Raise the dynamic plan-class thresholds

Per `design.md`'s `flow-fast-plan-class-thresholds`. TDD: update the test fixtures to the new
boundary numbers first (red against the unmodified script), then move the script's thresholds
(green).

`scripts/test-plan-class.sh` — change the four boundary cases to the new numbers: `small-5-12`
(5 tasks/12 files) becomes a fixture with **10 tasks / 25 files**, still asserting `class: small`
and `inputs: tasks=10 files=25 repos=1 …`; `regular-6` (6 tasks, asserting `regular`) becomes
**11 tasks** (one past the new small ceiling); `big-15` (15 tasks, asserting `big`) becomes
**30 tasks**; `migration-8-big` / `migration-7-regular` (asserting `big` / `regular` either side of
the migration boundary) become **15 tasks / 14 tasks**. Every other case (override-never-appears,
roll-stability, rolls-line-shape, bundle-static/free, missing-file, bad-repos) is untouched — none
depends on the class thresholds.

`scripts/plan-class.sh` — change the `small` condition from `tasks<=5 and files<=12` to `tasks<=10
and files<=25`, and the `big` condition from `tasks>=15 or files>=40 or (repos>1 and tasks>=8) or
(migration and tasks>=8)` to `tasks>=30 or files>=80 or (repos>1 and tasks>=15) or (migration and
tasks>=15)`; update the header comment's own restatement of both rules to match, and its `(kan-472-…)`
attribution comment to also name this change.

`skills/flow/brainstorm-planner.md` — in section D's **Decide**, change "Compact when
`compact_roll < 70` (small) or `< 30` (regular, big)" to "Compact when `compact_roll < 90` (small)
or `< 60` (regular, big)"; touch nothing else in that sentence (the experimental- and bundle-roll
cutoffs are unchanged, per `design.md`'s table).

  - [x] **Step 1: Edit** `scripts/test-plan-class.sh` per the four boundary changes above; run it
    against the **unmodified** `scripts/plan-class.sh` and confirm it now fails on exactly those
    four cases (red).
  - [x] **Step 2: Edit** `scripts/plan-class.sh` per the threshold and comment changes above.
  - [x] **Step 3: Edit** `skills/flow/brainstorm-planner.md` per the compact-roll sentence change.
  - [x] **Step 4: Verify** — `scripts/test-plan-class.sh` exits 0, all cases `pass`; then
    `scripts/check-vocabulary.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.py`,
    `scripts/check-contract-budget.sh`, `scripts/check-guard-symlinks.sh` all exit 0.

**Files:** `scripts/plan-class.sh`, `scripts/test-plan-class.sh`, `skills/flow/brainstorm-planner.md`
**Tests:** `test-plan-class.sh`'s four boundary cases, re-targeted to the new numbers (no new test
function — the existing cases are the coverage, now asserting the new thresholds)
**Regression:** reverting this commit drops `small`/`big` back to the old, tighter thresholds and
`compact_roll` back to `<70`/`<30` — every `/flow` and `/flow-fast` run with `## execution mode` or
`## review panel` set to `dynamic` would again classify more changes `regular`/`big` and roll
compact less often, the exact regression `test-plan-class.sh`'s four re-targeted cases would catch
by reasserting the old boundary and failing against the new script.
**Baseline:** before=16 after=16 — 16 cases in `test-plan-class.sh`, none added or removed, four
re-targeted
<!-- measured: grep -c '^run_guard' scripts/test-plan-class.sh @ branch spectre/kan-490-flow-fast-a-reduced-ceremony-flow-variant-drop -->
**Commit:** `feat(flow): raise the dynamic plan-class thresholds`
**Build:** green
