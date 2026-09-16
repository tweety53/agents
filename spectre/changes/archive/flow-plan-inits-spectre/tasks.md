> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

# flow-plan-inits-spectre

Implementation plan. Design and decisions: `design.md` beside this file — every task below
implements a section of it and restates nothing.

Four tasks, one repository (`agents`). No spec, no Go change, no runtime to verify against: the
change edits skill prose, two Bash scripts and their harnesses, and deletes files; nothing here
runs against `flowd` or its store.

**Every task's verify step is the lint guards its own files need plus its own harness**, never
`scripts/run-guard-tests.sh` — that run is the last bundle's FULL SUITE paragraph and
`flow.verify`'s.

---

- [x] 1. `/flow-plan` ends at `STARTED`; `/flow` drops the seed path and adopts an existing branch

**Build:** green
**Files:** `skills/flow-plan/SKILL.md`, `commands/flow-plan.md`, `commands-claude/flow-plan.md`, `skills/flow/SKILL.md`, `skills/flow/brainstorm.md`, `skills/flow/brainstorm-planner.md`, `skills/flow/SKILL-rationale.md`, `skills/flow/implement.md`, `skills/flow/verify-and-handoff.md`, `skills/flow-contracts/pipeline.md`, `skills/flow-contracts/git-boundaries.md`, `skills/flow-contracts/finish-contract-run1.md`, `skills/flow-fast/SKILL.md`, `skills/flow-plan/scripts/check-worktree-location.sh`, `skills/flow-plan/scripts/project-get.sh`, `skills/flow-plan/scripts/plan-dispatch-bundles.sh`, `skills/flow-plan/scripts/plan-dispatch-bundles.py`, `skills/flow-plan/scripts/plan-dispatch-groups.sh`, `skills/flow-plan/scripts/plan-dispatch-groups.py`
**Tests:** none — prose; the guards in the verify step are the check
**Regression:** reverting this commit restores the note path in every file at once; the guards
in the verify step would then pass on the old tree as they do today, so the proof is the design's
section 1 and 2 read against the files.
**After:** none
**Commit:** `feat(flow-plan): end the session at STARTED with the change's own artifacts`

  - [x] **Step 1: `skills/flow-plan/SKILL.md`.** Per design section 1. Rewrite the frontmatter
    description and the stance bullets (research reads the main checkout read-only; the session
    creates the change at capture; commits only the planning artifacts, once, on the change
    branch). Delete **The research worktree**. In **Reading the Tree** drop the
    `docs/research/` check. Replace **Staging a Note — the strict research-artifact path**, **The
    plan and the decision** and **Landing the note** with one section, **Capturing a new
    change**, that cites in order **A. Resolve the change and write `STARTED`** and the kickoff
    steps (`skills/flow/brainstorm.md`), **C. Create the change and its artifacts**, **D. Basic
    Workflow #3 — Writing plans**, **Decide** and **Plan review gate**
    (`skills/flow/brainstorm-planner.md`); states how the note's structure maps onto `design.md`
    and `proposal.md`; gives the `flow record decision -change <name> -session-token
    fp-<literal-token> -file <abs-worktree>/.superpowers/sdd/decision.json` call; and on **Yes**
    commits `spectre/changes/<name>/` with `chore(spectre): plan` and pushes `spectre/<name>`. Keep
    the bare-topic Jira Task creation as the way the key is known before naming. Keep **The
    Fixed Section Structure** and **The Step-by-Step Breakdown** as the shape of `design.md`'s
    body, dropping the `Source:` line and the template's title line. Close `plan.session` with
    `-outcome started` on this path; `captured` and `abandoned` stay. Rewrite the guardrails:
    drop "Don't run `spectre new`" and the three-paths commit rule; add "Don't create the change
    before the capture offer is accepted" (decision `init-after-capture-offer`). Delete every `docs/research`, `<stem>`, `plan-<stem>`,
    `_plan-<stem>` and seed-lookup mention. Name, in an invoking sentence, every guard the
    cited sections run — `check-worktree-location.sh`, `project-get.sh`, `check-plan-shape.sh`,
    `plan-class.sh`, `plan-dispatch-bundles.sh`, `plan-dispatch-groups.sh` — and add the
    missing symlinks under `skills/flow-plan/scripts/` with their `.py` siblings, since **Guard
    resolution** resolves a guard against the running command's own skill directory and
    `check-guard-symlinks.sh` requires the symlink for every invoked basename.
  - [x] **Step 2: Both command files.** `commands/flow-plan.md` and `commands-claude/flow-plan.md`:
    the summary paragraph says the session may end at a `STARTED` change committed on
    `spectre/<name>`; **When done** says "Ready to implement? Run `/flow <key>`". Keep the two
    files' frontmatter difference as it is.
  - [x] **Step 3: `skills/flow/SKILL.md`.** Delete the research-seed print block and its
    paragraph, and the **On a seeded creating run** startup block; keep the `planning:` /
    `toggles:` / `models:` lines joining the post-writing-plans `## Decision` print. In the
    guardrails drop `or <project>/docs/research/`.
  - [x] **Step 4: `skills/flow/brainstorm.md`.** Section A, "With a linked issue": before
    deriving a slug, resolve the candidate set per **Change name resolution**
    (`skills/flow-contracts/pipeline.md`); exactly one candidate whose name starts with
    `<lowercased-key>-` is this change, resumed at its recorded state; more than one is an
    **AskUserQuestion**; none derives the slug. Kickoff step 3: after the fetch, when
    `git -C <project> rev-parse -q --verify origin/spectre/<name>` succeeds, run
    `git worktree add <project>/.worktrees/<name> spectre/<name>` (a local branch tracking the
    remote), else the existing `-b` form. Delete the **On the fully-seeded bypass** block and the
    "On a no-seed run" qualifier on the gate sentence, and "A fully-seeded run skips it".
  - [x] **Step 5: `skills/flow/brainstorm-planner.md`.** Delete **Seed from a staged research
    note, if one exists** and **One shared mechanism, not two copies** whole. In C delete the
    **Delete the adopted staging note** paragraph. In D delete the **A seeded plan replaces the
    invocation below** paragraph and in Decide the **A seeded decision replaces the roll**
    paragraph. In **Plan review gate** the sentence naming where it runs becomes: at the end of
    every `/flow-plan` capture (**Capturing a new change**, `skills/flow-plan/SKILL.md`) and in
    `/flow` after the `flow.decide` record sequence; drop the fully-seeded clause.
  - [x] **Step 6: Contracts and the other phase files.** `pipeline.md`: delete the second bounded
    exception paragraph; the planning-path bullet names `<project>/spectre/changes/` alone; in
    **Command surface** `/flow-plan` is no longer "read-only" — it creates a `STARTED` change
    and nothing past it. `git-boundaries.md`: the `/flow-plan` row becomes "change captured —
    commits once, the planning artifacts on `spectre/<name>` in the change worktree, and pushes
    it"; the "No command touches the main checkout" paragraph says `/flow-plan` reads it and
    creates `<project>/.worktrees/<name>` at capture through `flow.kickoff`; delete the
    research-branch exception sentence; the pathspec block and its symlink sentence name
    `spectre/changes/` alone. `finish-contract-run1.md`: drop "together with the deletion of any
    adopted `<project>/docs/research/` note". `implement.md` and `verify-and-handoff.md`: drop
    the `docs/research/` half of each sentence. `skills/flow-fast/SKILL.md`: drop the three
    `<project>/docs/research/` mentions and "no research seed is ever read, so". `SKILL-rationale.md`:
    delete the **brainstorm-planner.md — Seed from a staged research note** entry; in the
    branch-backup entry drop the `/flow-plan` research-branch sentence; in the integrate
    preamble entry leave the history as written.
  - [x] **Step 7: Verify.**

  ```bash verified:every command is in .flow/project.md's ## lint list at branch spectre/flow-plan-inits-spectre
  scripts/check-vocabulary.sh
  scripts/check-references.sh
  scripts/check-contract-budget.sh
  scripts/check-markdown-integrity.py
  scripts/check-stage-mark-calls.sh
  scripts/check-dispatch-paragraphs.sh
  scripts/check-installed-citations.sh
  scripts/check-normative-inventory.sh
  scripts/check-model-keys.sh
  scripts/check-model-resolution-shell.sh
  ```

- [x] 2. Repository docs describe `/flow-plan` as ending at `STARTED`

**Build:** green
**Files:** `README.md`, `skills/README.md`, `CLAUDE.md`, `AGENTS.md`
**Tests:** none — prose
**Regression:** reverting this commit leaves README's brainstorm expansion describing a
`docs/research/` check no stage performs.
**After:** none
**Commit:** `docs(readme): describe /flow-plan as ending at STARTED`

  - [x] **Step 1: `README.md`.** Line 47's tree comment and line 656's table row: "thinking-partner
    mode; a captured session ends at a `STARTED` change for `/flow` to resume". Line 108: the
    plan session is "recorded against the Jira key until the session's own `STARTED` write
    creates the change". In **Brainstorm — `/flow`** delete the sentence from "Before the
    checklist opens" to "`flow-plan-staging`)."
  - [x] **Step 2: `skills/README.md`, `CLAUDE.md`, `AGENTS.md`.** The `/flow-plan` table row in
    each: "Thinking-partner mode — explore ideas, investigate, no implementation; a captured
    session creates the change at `STARTED` for `/flow` to resume". `CLAUDE.md` and `AGENTS.md`
    carry the same row text.
  - [x] **Step 3: Verify.**

  ```bash verified:every command is in .flow/project.md's ## lint list at branch spectre/flow-plan-inits-spectre
  scripts/check-vocabulary.sh
  scripts/check-references.sh
  scripts/check-contract-budget.sh
  scripts/check-markdown-integrity.py
  scripts/check-installed-rules.sh
  ```

- [x] 3. `commit-split.sh` and `recover-guard-incident.sh` treat `spectre/changes/` as the only planning path

**Build:** green
**Files:** `scripts/commit-split.sh`, `scripts/test-commit-split.sh`, `scripts/recover-guard-incident.sh`, `scripts/test-recover-guard-incident.sh`, `scripts/gather-self-review-context.sh`, `scripts/check-stage-mark-calls.sh`, `scripts/test-check-stage-mark-calls.sh`
**Tests:** `test-commit-split.sh` — `Case 7: a docs/research/ file is implementation — it lands in the implementation commit and the planning commit is skipped`; `test-recover-guard-incident.sh` — `Case 13: default-path-excludes-docs-research`; `test-check-stage-mark-calls.sh` — `Case 32: an unlisted -stage key is caught`, `Case 33: a listed -stage key passes`, `Case 34: a quoted listed -stage key passes`, `Case 35: a near-miss substring of a listed -stage key is caught`
**Regression:** reverting this commit puts a `docs/research/` file back in the planning commit,
which Case 7 asserts against; case 10 of `test-recover-guard-incident.sh` then sees
`docs/research` in the default plan.
**Baseline:** `test-commit-split.sh` before=6 after=7
<!-- measured: grep -cE '^# +[0-9]+\.' scripts/test-commit-split.sh @ branch spectre/flow-plan-inits-spectre -->
**After:** none
**Commit:** `fix(scripts): drop docs/research from the planning paths`

  - [x] **Step 1: `scripts/commit-split.sh`.** The `reset -q -- "$plan_dir" docs/research/` line
    and the `':(exclude)docs/research/'` pathspec lose `docs/research/`; the header's symlink
    paragraph names `spectre/changes/` alone.
  - [x] **Step 2: `scripts/test-commit-split.sh`.** `new_repo` no longer seeds `docs/research/`.
    Case 2 writes `spectre/changes/only.md`. Case 5's comment names `spectre/changes/` alone. Add
    Case 7: `new_repo`, write `docs/research/note.md`, run the script; assert the implementation
    commit exists, the planning commit is absent, and `git log -1 --name-only` of the
    implementation commit lists `docs/research/note.md`.
  - [x] **Step 3: `scripts/recover-guard-incident.sh`.** The `set -- docs/research spectre/changes`
    default and both usage lines name `spectre/changes` alone.
  - [x] **Step 4: `scripts/test-recover-guard-incident.sh`.** `P1` becomes
    `spectre/changes/kan-423/proposal.md` and `stash_planning_files` creates only
    `spectre/changes/kan-423`; case 10's second assertion checks `stash@{0}^3:spectre/changes`
    is absent from the custom-path plan. Every case that asserted the default plan named
    `docs/research` asserts `spectre/changes` instead.
  - [x] **Step 5a (panel F4, F5): `scripts/test-recover-guard-incident.sh`** gains Case 13, a no-argument
    run with a `docs/research/` file in the stash whose plan must not name it;
    `scripts/check-stage-mark-calls.sh` gains the rule that a `stage begin`'s `-stage` key is a row
    of README.md's Level 1 table, read from the repository's README on every run with the key
    shape `stats/internal/stages/names_test.go`'s `stageKeyRE` accepts (panel F7), with Cases 32,
    33 and 34 (a quoted key, panel F8) in its harness and the two fixtures that used the retired `flow.state-gate` key
    repointed at `flow.kickoff`.
  - [x] **Step 5: `scripts/gather-self-review-context.sh`.** Comments only: `docs/research/` is
    named as a retired planning path beside `docs/superpowers/`, per design decision
    `gather-keeps-retired-research-path`; the exclusion expression is unchanged.
  - [x] **Step 6: Verify.**

  ```bash verified:every harness runs standalone from the repository root at branch spectre/flow-plan-inits-spectre
  scripts/test-commit-split.sh
  scripts/test-recover-guard-incident.sh
  scripts/test-check-stage-mark-calls.sh
  scripts/check-stage-mark-calls.sh
  scripts/check-vocabulary.sh
  ```

- [x] 4. Delete `docs/research/`

**Build:** green
**Files:** `docs/research/kan-492.md`, `docs/research/kan-492/tasks.md`, `docs/research/kan-492/decision.json`, `docs/research/kan-512.md`, `docs/research/kan-512/tasks.md`, `docs/research/kan-512/decision.json`, `docs/research/flow-speedup.md`, `docs/research/flow-gymie-implementation-speedup.md`
**Tests:** none — a deletion
**Regression:** reverting this commit restores five notes no command reads.
**After:** none
**Commit:** `chore(docs): delete the docs/research staging notes`

  - [x] **Step 1: Delete.** `git rm -r docs/research`; the directory no longer exists. The
    `kan-516` note and its directory were already removed upstream (main's `chore(spectre): archive
    kan-516-…` moved them into the archive) before this branch rebased onto it, so this commit
    deletes the eight remaining files.
  - [x] **Step 2: Verify.**

  ```bash verified:the guard is in .flow/project.md's ## lint list at branch spectre/flow-plan-inits-spectre
  scripts/check-references.sh
  ```
