> **Execution:** `/myflow-do` runs Basic Workflow #2–#6 via the apply skill (#7 runs later, in
> `/myflow-review`). Mark each checkbox when its task passes spec + quality review (SDD #6).
>
> **Goal:** Replace myflow's twelve stages and eighteen commands with three states and three
> commands, fold review+test into one gate, make finish a two-run integrate-then-archive
> command, and fix handoff output.
>
> **Architecture:** `skills/myflow-contracts/pipeline.md` is canonical and is rewritten first;
> every other layer is then swept to agree with it. `scripts/check-vocabulary.sh` is taught the
> retired literals up front so the sweep has a mechanical progress signal.
>
> **Tech stack:** Markdown contracts, Bash guard scripts, `setup.sh`. No application code, so
> "the test" is a guard script assertion that fails against the current tree.

## Global constraints

- Absolute paths in every generated output and every handoff. Never `../sibling`, never a
  main-checkout path while a worktree holds the work.
- Never add a lint/guard suppression to make a check pass; fix the offending line. The
  `vocab-guard:allow` marker is only for lines that must quote retired vocabulary to do their
  job (the guard's own literal list, this plan, migration notes).
- No backward compatibility: no aliases, no deprecation shims, no state-file migration.
- Every command file in **both** `commands/` and `commands-claude/` must state exactly the
  states its row in `myflow-command-surface` lists, and agree with the skill it delegates to.
- The state file is never staged, committed, or archived.
- Verification is `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
  `scripts/test-check-references.sh`, `scripts/test-setup.sh` — run from a worktree that
  carries `scripts/`; all are cwd-independent and take no arguments.

---

## 1. Teach the vocabulary guard the retired literals

**Files:** Modify `scripts/check-vocabulary.sh:204` (`check_retired_stage_vocabulary`)

**Produces:** `scripts/check-vocabulary.sh` failing with a hit list that later tasks drive to
zero. The hit count is this change's progress metric.

- [x] 1.1 Extend the stage pattern with every retired state value and command name.

  In `check_retired_stage_vocabulary`, replace the `pattern` line with:

  ```bash
  local pattern='awaiting-review|awaiting-test|(^|[^-])\bcode-review\b' # vocab-guard:allow
  pattern+='|awaiting-proposal-review|proposal-done|awaiting-do-review'  # vocab-guard:allow
  pattern+='|do-review-started|do-done|awaiting-fix-review'              # vocab-guard:allow
  pattern+='|fix-review-started|awaiting-manual-test|manual-test-done'   # vocab-guard:allow
  pattern+='|awaiting-pr-review|review-done'                            # vocab-guard:allow
  pattern+='|myflow-full|myflow-fast-path|myflow-manual-test'           # vocab-guard:allow
  pattern+='|myflow-start-fix|myflow-start-done|myflow-do-fix'          # vocab-guard:allow
  pattern+='|myflow-do-done|myflow-do-manual-review|myflow-review-done' # vocab-guard:allow
  pattern+='|myflow-state-advance|automerge|skip-manual-test'           # vocab-guard:allow
  pattern+='|skip-review|skip-propose|propose-only|full-panel'          # vocab-guard:allow
  pattern+='|commit-during-apply'                                       # vocab-guard:allow
  ```

  `checkpoint` is deliberately **not** in this list: it is an ordinary English word that appears
  legitimately in prose about review checkpoints, so a literal match would produce false hits
  that could only be silenced with markers — and a marker on a line that is not drift teaches the
  guard to lie. It is swept by hand in task 15.6 instead.

  Update the guard's failure message to point at the state transition table in
  `skills/myflow-contracts/pipeline.md` rather than the rule file's `## Stage transitions`.

- [x] 1.2 Run the guard and confirm it now FAILS.

  ```bash
  scripts/check-vocabulary.sh; echo "exit=$?"
  ```

  Expected: exit 1, a long `⚠ Retired myflow stage vocabulary found:` list. Record the baseline
  count — later tasks must drive it down:

  ```bash
  scripts/check-vocabulary.sh 2>&1 | grep -c ':[0-9]*:' 
  ```

- [x] 1.3 Confirm the guard still passes on a tree with none of the literals, so the pattern
  has no accidental always-match:

  ```bash
  TMPD="$(mktemp -d)"; printf 'clean file\n' >"$TMPD/a.md"
  scripts/check-vocabulary.sh "$TMPD"; echo "exit=$?"   # expect 0
  rm -rf "$TMPD"
  ```

- [x] 1.4 Stage the change (`git add scripts/check-vocabulary.sh`). Do not commit — `/myflow-do`
  stages only.

---

## 2. Rewrite `pipeline.md` as the canonical five-state contract

**Files:** Rewrite `skills/myflow-contracts/pipeline.md`

**Consumes:** Nothing. **Produces:** The section names every other layer references —
`## States`, `## Command surface`, `## State transitions`, `## Git boundaries`,
`## Mismatch handoff`, `## Handoff output`, `## Finish contract`, `## Model policy`,
`## Change name resolution`, `## IntelliJ commands`.

- [x] 2.1 Write `## States` — the five values, what each means, and the human gate each implies.
  Include the diagram:

  ```text
  /myflow-start  → STARTED      human: read the proposal artifact
  /myflow-do     → IN_PROGRESS  human: review the staged diff
  /myflow-test   → TEST         human: run the apps
  /myflow-review → REVIEW       human: review + merge the PR
  /myflow-finish → FINISHED     terminal
  ```

- [x] 2.2 Write `## Command surface` and `## State transitions` — the accepted-states table from
  the `myflow-command-surface` delta spec, verbatim in meaning:

  | Command | Accepts | Ends at |
  |---------|---------|---------|
  | `/myflow-start` | *(none)* or `STARTED` | `STARTED` |
  | `/myflow-do` | `STARTED`, `IN_PROGRESS`, `TEST`, `REVIEW` | `IN_PROGRESS` from `STARTED`; else unchanged |
  | `/myflow-test` | `IN_PROGRESS` | `TEST` |
  | `/myflow-review` | `IN_PROGRESS`, `TEST` | `REVIEW` |
  | `/myflow-finish` | `REVIEW` | `FINISHED` |
  | `/myflow-fast` | *(none)* | `REVIEW` |
  | `/myflow-status`, `/myflow-info` | any | unchanged |

  State explicitly: a fix never moves the state; `/myflow-do` advances only `STARTED` →
  `IN_PROGRESS`; there is no `originStage`.

- [x] 2.3 Write `## Git boundaries` — per state, what git actions are permitted. Include the one
  carve-out: `/myflow-do` commits and pushes only at `REVIEW`, because a staged-only fix would
  be invisible on an open PR.

- [x] 2.4 Write `## Mismatch handoff` — stop, report actual state / expected states / suggested
  command, AskUserQuestion with "run the suggested command instead" as default and recommended.

- [x] 2.5 Write `## Handoff output` — the fixed shape, absolute paths, no pasted bodies, the
  bare next-command last line, and the `openspec/` exclusion from the review diff.

- [x] 2.6 Write `## Finish contract` — the five ordered steps and the four gating preflight checks from
  the `myflow-finish-cleanup` delta spec.

- [x] 2.7 Carry forward `## Model policy`, `## Change name resolution` and `## IntelliJ commands`
  from the existing file, updating them to the new command names and five states. The IntelliJ
  table becomes: `STARTED` → main checkout; `IN_PROGRESS`/`TEST`/`REVIEW` → apply worktree root.

- [x] 2.8 Delete from the file every section made obsolete: `## Pipeline stages`, `## Stage
  transitions`, `## Fix re-entry`, `## Stage boundaries`, `## Manual review (Gate B)`,
  `## Manual test (Gate C)`, `## Fix rounds`, `## Review`, `## Auto-merge (opt-in)`,
  `## PR review (Gate D)`, `## Finish`, `## Full cycle gates`, `## Opt-out (explicit only)`.

- [x] 2.9 Verify the file names no retired literal:

  ```bash
  scripts/check-vocabulary.sh skills/myflow-contracts/pipeline.md; echo "exit=$?"  # expect 0
  ```

- [x] 2.10 Stage.

---

## 3. Rewrite the state-file and self-heal contracts

**Files:** Rewrite `skills/myflow-contracts/state-file.md`, `skills/myflow-contracts/state-self-heal.md`

**Consumes:** `## States` from task 2.

- [x] 3.1 Rewrite `state-file.md` around the new object. Keep the `--git-common-dir` →
  `<project-key>` resolution and the "never staged, committed, or archived" rule unchanged.

  ```json
  {
    "state": "TEST",
    "branch": "openspec/<name>",
    "worktrees": { "/abs/path/to/worktree": "<merge-base sha>" },
    "artifactUrl": null,
    "jiraIssue": null,
    "prUrl": null,
    "tested": null,
    "fast": false,
    "updatedAt": "<ISO-8601 UTC>",
    "updatedBy": "/myflow-test"
  }
  ```

- [x] 3.2 In `state-file.md`, state that `worktrees` keys are the authoritative list of affected
  worktrees for a multi-repo change (replacing the old `MERGE_BASE`-keyed map), that `tested` is
  `null` / `"skipped"` / `true` with `/myflow-review` its only writer, and that every write
  renders the whole object carrying forward unowned fields.

- [x] 3.3 In `state-file.md`, write the monotonicity rule: never write an earlier state, except
  `REVIEW` → `TEST` on conclusively-established PR non-existence. Remove every mention of
  `gates`, `originStage` and `REVIEWED_TREE`.

- [x] 3.4 Rewrite `state-self-heal.md`'s contradiction table to the five rows in the
  `myflow-state-machine` delta spec. Keep "read artifacts from the apply worktree when one
  exists", "a check that cannot be performed is not a contradiction", and "never infer
  `tested: true`".

- [x] 3.5 Remove the legacy-stage migration mapping — with no backward compatibility, a file
  carrying an old stage value is unparseable and self-heal rewrites it from artifacts.

- [x] 3.6 Verify and stage:

  ```bash
  scripts/check-vocabulary.sh skills/myflow-contracts; echo "exit=$?"  # expect 0
  ```

---

## 4. Add the `## stop` key to the project-configuration contract

**Files:** Modify `skills/myflow-contracts/project-configuration.md`; modify `.myflow/project.md`

- [x] 4.1 Add `## stop` to the key table in `project-configuration.md`: the command that stops
  the project's local stack, run by `/myflow-finish` before its stack-stopped check. State that
  an absent key means the check is **skipped, not failed**.

- [x] 4.2 Add a `## stop` section to this repo's own `.myflow/project.md` recording that there is
  no stack to stop, so the key's absence is a documented fact rather than an oversight.

- [x] 4.3 Update the `## test` and `## lint` sections of `.myflow/project.md` to drop
  `scripts/test-state-advance.sh` (deleted in task 12).

- [x] 4.4 Stage.

---

## 5. Rewrite the always-on rule stub

**Files:** Modify `rules/myflow-manual-review.mdc`

- [x] 5.1 Rewrite the frontmatter `description` to name the five states and the six commands. It
  must contain no retired stage value — the description is inlined into every global managed
  block, so a stale one is the most-read stale text in the system.

- [x] 5.2 Replace the title and twelve-stage diagram with the five-state diagram from task 2.1.

- [x] 5.3 Keep `## Load the pipeline before acting` and the four-contract table unchanged in
  shape; confirm each path still resolves.

- [x] 5.4 Verify the file is still small and clean:

  ```bash
  wc -c rules/myflow-manual-review.mdc          # expect well under 8192
  scripts/check-vocabulary.sh rules; echo "exit=$?"   # expect 0
  ```

- [x] 5.5 Stage.

---


## 6. Create `skills/myflow-start/`

**Files:** Create `skills/myflow-start/SKILL.md`; delete-later sources
`skills/openspec-propose-superpowers/`, `skills/openspec-propose-fix-superpowers/`,
`skills/openspec-propose/`

- [x] 6.1 Create `skills/myflow-start/SKILL.md` from `openspec-propose-superpowers`, frontmatter
  `name: myflow-start`. Accepts *(none)* or `STARTED`; ends at `STARTED`.

- [x] 6.2 Inline the artifact-creation steps currently delegated to `openspec-propose` (the
  `openspec new change` / `openspec status` / `openspec instructions` loop). The delegation must
  not survive — `openspec-propose` is deleted in task 11.

- [x] 6.3 Fold in `openspec-propose-fix-superpowers`: when invoked at `STARTED`, revise the
  existing artifacts and republish the proposal artifact to the same deterministic source path
  (`/Users/tweety53/Agents/myflow/state/<project-key>/<name>-proposal-artifact.html`), keeping the
  URL stable. Preserve the decision-supersede-by-ID rule.

- [x] 6.4 Handoff in the fixed shape: what happened, the artifact URL, the absolute main-checkout
  `open -na "IntelliJ IDEA"` command, then the bare last line `/myflow-do <name>`.

- [x] 6.5 Verify `scripts/check-vocabulary.sh skills/myflow-start` is clean, then stage.

---

## 7. Create `skills/myflow-do/` — implementation

**Files:** Create `skills/myflow-do/SKILL.md` plus the reviewer-prompt assets it inherits;
delete-later sources `skills/openspec-apply-superpowers/`,
`skills/openspec-apply-fix-superpowers/`, `skills/openspec-apply-change/`

- [x] 7.1 Create `skills/myflow-do/SKILL.md` from `openspec-apply-superpowers`, carrying its
  reviewer prompt files and `engineering-principles.md` into the new directory unchanged.

- [x] 7.2 Fold in `openspec-apply-fix-superpowers`: one skill, branching on incoming state. From
  `STARTED` it creates the worktree and runs the full plan; from `IN_PROGRESS` it resumes the
  existing worktree and applies a fix, documenting it in `proposal.md`/`tasks.md` or a nested
  `<name>-fix-N` sub-change first.

- [x] 7.3 Write the state rule explicitly: advance to `IN_PROGRESS` only from `STARTED`; otherwise
  write the state back unchanged. Remove every reference to `originStage` and fix re-entry.

- [x] 7.4 Write the git rule: `git add` always; commit **and push** only when the state file
  already records a `prUrl` (a PR exists, so a staged-only fix would be invisible on it).

- [x] 7.5 Keep the review panel's roster and trigger table unchanged, but **set every slot to
  Sonnet**: pass Sonnet explicitly to the primary, principles, adversarial and extra-lens
  reviewers; pass no model override to the `subagent_type` slots (Bugbot, Security Review).
  **Delete the provider-family economy mapping table** and every mention of parent-model
  inheritance and economy tiers. Remove the `full-panel` flag; breadth rests on the escalation
  triggers alone.

- [x] 7.6 Run this repo's guards as the verification step, since nothing runs them later:

  ```bash
  scripts/check-vocabulary.sh && scripts/check-references.sh \
    && scripts/test-check-references.sh && scripts/test-setup.sh
  ```

  A non-zero exit blocks the handoff.

---

## 8. `skills/myflow-do/` — the test guide and the handoff

**Files:** `skills/myflow-do/SKILL.md`; delete-later source
`skills/openspec-manual-test-superpowers/`

- [x] 8.1 Inline the guide-writing steps from `openspec-manual-test-superpowers` so `/myflow-do`
  writes `docs/manual-test/<name>.md` in the same run as the code. **Delete the skip prompt** and
  every mention of `gates.tested` and `SKIPPED`-marked guides — there is nothing to ask.

- [x] 8.2 Keep the rule that run commands use **absolute apply-worktree roots** for every app in
  scope, resolved from `git worktree list` or the state file's `worktrees` keys — never a relative
  sibling path, never a main checkout while a worktree holds the work.

- [x] 8.3 On a fix re-run, refresh the guide alongside the code, preserving already-ticked boxes,
  so diff and guide never drift apart.

- [x] 8.4 Write the handoff: worktree absolute path, the guide's absolute path (never its body),
  confirmation the changes are staged, and the diff commands **excluding planning artifacts**:

  ```bash
  git -C <abs-worktree> status
  git -C <abs-worktree> diff --cached --stat -- . ':(exclude)openspec/'
  git -C <abs-worktree> diff --cached        -- . ':(exclude)openspec/'
  ```

  Then the bare last line `/myflow-finish <name>`.

- [x] 8.5 Verify `scripts/check-vocabulary.sh skills/myflow-do` is clean, then stage.

---

## 9. Create `skills/myflow-finish/` — run 1, integration

**Files:** Create `skills/myflow-finish/SKILL.md`; delete-later sources
`skills/openspec-review-superpowers/`, `skills/openspec-archive-superpowers/`

- [x] 9.1 Create `skills/myflow-finish/SKILL.md`. Accepts `IN_PROGRESS`. Branch on whether the
  branch has reached the base branch — **that alone** decides run 1 vs run 2. Record no
  "integration started" field.

- [x] 9.2 Resolve the base branch from the repository rather than assuming:

  ```bash
  BASE="$(git -C "$WT" symbolic-ref --quiet --short refs/remotes/origin/HEAD 2>/dev/null | sed 's#^origin/##')"
  BASE="${BASE:-$(git -C "$WT" rev-parse --abbrev-ref HEAD@{upstream} 2>/dev/null | sed 's#^origin/##')}"
  ```

  Stop and ask if neither resolves. Never hardcode `main` or `develop`.

- [x] 9.3 Ask **before any git action** how to land the branch:

  > **How should this branch land?**
  > - **Open a pull request** *(default, recommended)*
  > - **Merge and push**
  > - **Handle it manually**

  Then run to completion without asking again. Never remember the answer between runs.

- [x] 9.4 Implement the three routes. All commit the staged work first (including
  `docs/manual-test/<name>.md` and the `openspec/` planning artifacts). PR route: push, open a PR
  via `gh` when usable, else print the forge's create-PR URL and ask whether it was opened; record
  `prUrl`. Merge route: push, merge into the base branch, push that. Manual route: push the branch
  only and say what is left to do.

- [x] 9.5 **Run no tests, linters, or coverage check.** Delete every such step inherited from
  `openspec-review-superpowers` — the coverage-vs-delta-spec comparison included.

- [x] 9.6 End run 1 at `IN_PROGRESS`, with the bare last line `/myflow-finish <name>` — the same
  command, because that is what the operator runs once the branch is merged.

---

## 10. `skills/myflow-finish/` — run 2, archive and cleanup

**Files:** `skills/myflow-finish/SKILL.md`; delete-later sources
`skills/openspec-archive-change/`, `skills/openspec-sync-specs/`

- [x] 10.1 Verify the merge: `gh` when usable, else `git merge-base --is-ancestor`, which must
  stay reachable on its own as the non-GitHub evidence path. Not merged → do not archive.

- [x] 10.2 Inline the delta-sync steps from `openspec-sync-specs` and the archive-move steps from
  `openspec-archive-change`. **Do this before task 11 deletes them**, and verify:

  ```bash
  grep -rn "openspec-sync-specs\|openspec-archive-change" skills/myflow-finish/   # expect no output
  ```

  Archive any nested `<name>-fix-N` sub-changes in the same operation.

- [x] 10.3 Commit and push the archive on the base branch in the main checkout:

  ```bash
  git -C "$MAIN_CHECKOUT" add openspec/
  git -C "$MAIN_CHECKOUT" commit -m "chore(openspec): archive <name>"
  git -C "$MAIN_CHECKOUT" push
  ```

  Document that no merge of the change branch happens — step 10.1 already proved it merged — and
  that a non-base branch checked out means commit there, merge into base, push that.

- [x] 10.4 Worktree cleanup. Resolve the set from `worktrees` keys, else scan:

  ```bash
  git -C "$REPO" worktree list --porcelain \
    | awk '/^worktree /{w=$2} /^branch /{if ($2=="refs/heads/openspec/<name>") print w}'
  ```

  Never guess a path. Run all four gating checks before removing anything, then disclose the
  ignored files `--force` will destroy and confirm:

  ```bash
  git -C "$WT" status --porcelain --untracked-files=no        # 1. tracked changes — empty
  git -C "$WT" ls-files --others --exclude-standard           # 2. untracked+unignored — empty
  # 3. commits that exist only here — satisfied by the branch already being merged
  # 4. the project's `## stop` command, when declared
  git -C "$WT" ls-files --others --ignored --exclude-standard # disclosure — show and confirm
  ```

  Then, only if all passed:

  ```bash
  git -C "$REPO" worktree remove --force "$WT"
  git -C "$REPO" branch -d "openspec/<name>"
  git -C "$REPO" worktree prune
  ```

- [x] 10.5 Document the cleanup rules: `--force` is for ignored build output only, which is why
  check 1 runs first and separately; `git branch -d` never `-D`; an already-removed worktree is
  success; **any failed check leaves every worktree alone** and reports why.

- [x] 10.6 Write `FINISHED`, set `worktrees` to `{}`, carry `jiraIssue`, `artifactUrl` and `prUrl`
  forward. Handoff names no next command.

- [x] 10.7 Verify `scripts/check-vocabulary.sh skills/myflow-finish` is clean, then stage.

---

## 11. Rewrite `myflow-info` and `myflow-status`, then delete the retired

**Files:** Rewrite `skills/myflow-info/SKILL.md`, `skills/myflow-status/SKILL.md`; delete twelve
skill directories, one script, and the retired command files in both trees

- [x] 11.1 Rewrite `myflow-info` to read `skills/myflow-contracts/pipeline.md` and explain three
  states and three commands.

- [x] 11.2 Rewrite `myflow-status` to report `state`, `prUrl`, worktree paths (absolute), last
  update, and the next command per state. Remove every gate column and the `tested` column.

- [x] 11.3 Confirm nothing references a skill about to be deleted:

  ```bash
  for s in openspec-propose openspec-propose-superpowers openspec-propose-fix-superpowers \
           openspec-apply-change openspec-apply-superpowers openspec-apply-fix-superpowers \
           openspec-manual-test-superpowers openspec-review-superpowers \
           openspec-archive-change openspec-archive-superpowers openspec-sync-specs \
           openspec-update-change openspec-full-cycle-superpowers \
           openspec-fast-path-superpowers myflow-state-advance; do
    printf '%-40s %s\n' "$s" "$(grep -rl "$s" skills/myflow-* commands commands-claude 2>/dev/null | tr '\n' ' ')"
  done
  ```

  Every line must be blank after the name. A hit is an un-inlined delegation — fix before deleting.

- [x] 11.4 `git rm -r` the fifteen retired skill directories listed above. Keep
  `openspec-explore`, `myflow-contracts`, `myflow-info`, `myflow-status`, and the three new ones —
  seven in total.

- [x] 11.5 `git rm scripts/test-state-advance.sh`. The script it asserted against lives at
  `skills/myflow-state-advance/state-advance.sh` and goes with that directory in 11.4 — it is
  **not** under `scripts/`.

- [x] 11.6 `git rm` the thirteen retired command files from **both** `commands/` and
  `commands-claude/`, and the five retired `commands/opsx-*.md` (keeping `opsx-explore.md`).

- [x] 11.7 Confirm the tree:

  ```bash
  ls skills/ | tr '\n' ' '
  # expect: README.md myflow-contracts myflow-do myflow-finish myflow-info
  #         myflow-start myflow-status openspec-explore
  ```

- [x] 11.8 Stage.

---

## 12. Rewrite the command files in both trees

**Files:** Create five files each in `commands/` and `commands-claude/`

- [x] 12.1 Write `commands-claude/myflow-{start,do,finish,info,status}.md`. Each names the skill it
  loads and the states it accepts, and keeps `model:` frontmatter — `opus` for `myflow-start`,
  `sonnet` for the rest.

- [x] 12.2 Write the five matching `commands/myflow-*.md` for Cursor, carrying the manual
  model-switch note for `/myflow-start`.

- [x] 12.3 Cross-check each command's accepted states against the table in task 2.2. A command
  that contradicts its skill is a defect, not a shorthand.

- [x] 12.4 Verify:

  ```bash
  ls commands-claude/ | wc -l   # expect 5
  ls commands/ | wc -l          # expect 6  (5 myflow + opsx-explore)
  ```

- [x] 12.5 Stage.

---

## 13. Sweep the docs

**Files:** `README.md`, `CLAUDE.md`, `AGENTS.md`, `skills/README.md`

- [x] 13.1 In `README.md`, replace the twelve-stage block with a mermaid transition graph:

  ````markdown
  ```mermaid
  stateDiagram-v2
      [*] --> STARTED: /myflow-start
      STARTED --> STARTED: /myflow-start (revise)
      STARTED --> IN_PROGRESS: /myflow-do
      IN_PROGRESS --> IN_PROGRESS: /myflow-do (fix)
      IN_PROGRESS --> IN_PROGRESS: /myflow-finish (run 1 — integrate)
      IN_PROGRESS --> FINISHED: /myflow-finish (run 2 — after merge)
      FINISHED --> [*]
  ```
  ````

- [x] 13.2 Rewrite `README.md`'s command reference table to the five commands and its skills tree
  to the seven directories. **Delete the "Flags" line entirely** rather than shortening it — there
  are none, and a section listing zero flags invites one to be added back.

- [x] 13.3 Rewrite the `/myflow` sections of `CLAUDE.md` and `AGENTS.md` — skill index, command
  table, states. Delete their **Flags** paragraphs outright, remove the `myflow-state-advance`
  row, and remove the pure-state-write paragraph.

- [x] 13.4 Rewrite `skills/README.md`'s command map.

- [x] 13.5 Stage.

---

## 14. Drive the guards to clean and verify the installer

- [x] 14.1 Add `myflow-test` and `myflow-review` to the retired-literal list in
  `scripts/check-vocabulary.sh` (they were live commands when task 1 ran, so they were not in it).

- [x] 14.2 Drive the vocabulary guard to zero:

  ```bash
  scripts/check-vocabulary.sh; echo "exit=$?"   # expect 0
  ```

  Every remaining hit is either real drift (fix the line) or a line that must quote retired
  vocabulary to do its job (add a trailing `vocab-guard:allow` marker to **that line only** —
  never a whole file, never to silence real drift).

- [x] 14.3 Run the reference guard, which catches sections that moved rather than literals
  retired. This also clears the **pre-existing** failure at `rules/myflow-manual-review.mdc:23`
  introduced by `2e3f973`:

  ```bash
  scripts/check-references.sh; echo "exit=$?"   # expect 0
  ```

- [x] 14.4 Run both harnesses:

  ```bash
  scripts/test-check-references.sh; echo "exit=$?"   # expect 0
  scripts/test-setup.sh;            echo "exit=$?"   # expect 0
  ```

  A `test-setup.sh` failure means the installer broke on the changed rule/skill/command set — fix
  `setup.sh` or its input, never the assertion.

- [x] 14.5 Exercise the installer against a sandboxed HOME, never the real one:

  ```bash
  SANDBOX="$(mktemp -d)"
  HOME="$SANDBOX" ./setup.sh global
  ls "$SANDBOX/.claude/commands/" | tr '\n' ' '     # expect the 5 myflow commands
  ls "$SANDBOX/.claude/skills/"   | tr '\n' ' '     # expect the 7 skill dirs
  wc -c "$SANDBOX/.claude/CLAUDE.md"                # expect well under 12288
  scripts/check-vocabulary.sh "$SANDBOX/.claude/CLAUDE.md"; echo "exit=$?"   # expect 0
  rm -rf "$SANDBOX"
  ```

- [x] 14.6 Manually sweep for completeness. The vocabulary guard matches literals only and proves
  nothing about paraphrases; read the diff for stale *descriptions* of the flow — "twelve stages",
  "five states", "Gate B/C/D", "pure state write", "manual test stage", `checkpoint` — and fix
  them.

- [x] 14.7 `git add -A` in the worktree, confirm `git status` shows everything staged and
  uncommitted, and hand off.
