# Remove the manual test guide and its gate — implementation plan

> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development
> (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use
> checkbox (`- [ ]`) syntax for tracking.

**Goal:** Delete the generated manual test guide and everything that reads it, and print the run
instructions in the `/myflow-do` handoff instead.

**Architecture:** The guide was one artifact with four consumers — the `/myflow-do` writer, the
`IN_PROGRESS` handoff template, the unfinished-work guard, and the staging exclusion list. Each
consumer is removed in its own task, guard behaviour first (where tests exist to drive it), then the
contracts and skills that describe it, then the documentation digests. The guide's only surviving
content — the preamble stating how to run what is in scope — becomes a section of the `IN_PROGRESS`
handoff template.

**Tech Stack:** Bash (POSIX-ish, `bash` shebangs), Python 3 standard library (one guard only),
Markdown contracts, the `openspec` CLI.

## Global Constraints

- No suppression markers and no guard weakening. Every lint hit is fixed by editing the offending
  line. This repository declares no auto-fix command.
- `docs/manual-test/`, `openspec/changes/archive/` history and `check-vocabulary.sh`'s retired
  *command name* allowlist entries (`awaiting-manual-test`, `myflow-manual-test`,
  `skip-manual-test`, `manual-test-done`) are three different things. Only the first is deleted;
  the archive and the allowlist are untouched.
- Every path printed by any pipeline command stays absolute.
- `check-contract-budget.sh` is a ratchet keyed on the path relative to the repository root. A file
  that grows past its row must have its row raised deliberately, in the same task that grew it.
- The pipeline's three-line state digest is copied into several files on purpose. The **facts** are
  fixed; the **wording** is deliberately not byte-identical, so do not harmonise the copies.

---

## Task 1: Drop the guide signals from the unfinished-work guard

**Build:** green

**Files:**
- Modify: `scripts/check-unfinished-work.sh` — remove the `GUIDE` assignment and the guide signal
  block
- Modify: `scripts/test-check-unfinished-work.sh` — remove every guide fixture and assertion

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: a guard whose only signals are `openspec/changes/<name>/tasks.md` (plus nested
  `<name>-fix-N` sub-changes) and `.superpowers/sdd/final-review-panel.md`. Exit codes are
  unchanged: `0` `CLEAR`, `1` `OUTSTANDING` with a per-signal breakdown, `2` cannot determine, with
  no verdict line on `2`.

- [x] **Step 1: Make the harness state the new contract**

In `scripts/test-check-unfinished-work.sh`, delete every case that plants, mutates, chmods or
removes `$WT/docs/manual-test/demo.md`, together with the `mkdir -p "$WT/docs/manual-test"`
components of its fixture setup and the `assert_reason "no manual test guide"` case. Add one case
proving the guard no longer cares:

```bash unverified:confirm the harness's helper names — run_guard and assert_clear — match the surrounding cases in this file
# A worktree with no docs/manual-test/ at all is CLEAR when the plan and panel are clean.
rm -rf "$WT/docs/manual-test"
run_guard "$WT" demo
assert_clear "an absent docs/manual-test/ is not a signal"
```

- [x] **Step 2: Run the harness and watch it fail**

```bash verified:authored in-tree for this change
scripts/test-check-unfinished-work.sh
```

Expected: FAIL. The new case reports `OUTSTANDING` with reason `no manual test guide at …`, because
the guard still requires the file.

- [x] **Step 3: Remove the signals from the guard**

In `scripts/check-unfinished-work.sh`, delete the `GUIDE=` assignment near the `PANEL=`/`CHANGES=`
block, and delete the whole `# Signals one and four` section — the `if [ ! -f "$GUIDE" ]` branch,
the `BOXES` count, and the `SECTION_STATE` awk program with its three `KNOWN_*` branches.

Keep `count_unticked()`: the plan signal calls it. Keep `unreadable()`: the plan and panel signals
call it. Update the header comment that enumerates the signals to say two, and renumber the
surviving sections' comments so no comment refers to a signal that no longer exists.

- [x] **Step 4: Run the harness and watch it pass**

```bash verified:authored in-tree for this change
scripts/test-check-unfinished-work.sh
```

Expected: PASS, every case.

- [x] **Step 5: Prove the surviving signals still fire**

Break each one deliberately and confirm the guard goes red, then restore:

```bash unverified:confirm the fixture variable is $WT and the change name is demo in this harness
printf -- '- [ ] a task\n' >> "$WT/openspec/changes/demo/tasks.md"   # expect OUTSTANDING
printf -- '- finding: open\n' >> "$WT/.superpowers/sdd/final-review-panel.md"  # expect OUTSTANDING
```

---

## Task 2: Match both plan-commit wordings in the self-review context script

**Build:** green

**Files:**
- Modify: `scripts/gather-self-review-context.sh:447` — the `PLAN_SHA` `--grep`
- Modify: `scripts/gather-self-review-context.sh:456` — the `IMPL_SHA` `grep -Ev` exclusion
- Modify: `scripts/gather-self-review-context.sh:427` — the comment naming the commit
- Modify: `scripts/test-gather-self-review-context.sh` — add a new-wording fixture

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: `PLAN_SHA` resolves the plan commit under either subject. `IMPL_SHA` excludes the plan
  commit under either subject, so it never mistakes a plan commit for the implementation commit.

- [x] **Step 1: Add a failing fixture for the new subject**

`scripts/test-gather-self-review-context.sh` already builds fixtures using the old subject at lines
96 and 625. Leave both — they are the back-compat proof. Add a third fixture repository whose plan
commit uses the new subject, and assert the script resolves that commit as `PLAN_SHA` and not as
`IMPL_SHA`:

```bash unverified:confirm the fixture builder's helper name and its assertion helpers in this harness
git commit -q --allow-empty -m "chore(demo): plan and session records" \
  && git commit -q --allow-empty -m "chore(demo): archive demo"
```

- [x] **Step 2: Run the harness and watch it fail**

```bash verified:authored in-tree for this change
scripts/test-gather-self-review-context.sh
```

Expected: FAIL. The new-subject plan commit is not matched by `PLAN_SHA`, and `IMPL_SHA` picks it up
instead of the implementation commit.

- [x] **Step 3: Widen both greps to an alternation**

```bash verified:authored in-tree for this change
  PLAN_SHA="$(git -C "$REPO_ROOT" log -E \
    --grep="^chore\(${NAME_RE}\): plan(, test guide and| and) session records" \
    --max-count=1 --format='%H' 2>/dev/null || true)"
```

```bash verified:authored in-tree for this change
    | grep -Ev "plan(, test guide and| and) session records|^[0-9a-f]+ chore\(${NAME_RE}\): .*archive" \
```

Both sites gain a comment stating that the alternation exists so changes committed before the
subject was renamed keep resolving, and that removing the old branch would make `IMPL_SHA` resolve
those changes' plan commits as their implementation commits — a wrong answer, not a missing one.

- [x] **Step 4: Run the harness and watch it pass**

```bash verified:authored in-tree for this change
scripts/test-gather-self-review-context.sh
```

Expected: PASS, including the two old-subject fixtures.

---

## Task 3: Drop `docs/manual-test/` from the staging exclusions

**Build:** green

**Files:**
- Modify: `scripts/test-uncommitted-review-package.sh:329-410` — fixtures and assertions
- Modify: `skills/myflow-contracts/pipeline.md:145-146`, `:245` — the git chain and the prose
- Modify: `skills/myflow-do/SKILL.md:131`, `:481-482` — the note and the staging commands
- Modify: `skills/myflow-finish/SKILL.md:140`, `:158-159` — the prose and the git chain
- Modify: `skills/myflow-contracts/finish-contract.md:126` — the sentence naming the three paths

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: the planning paths are exactly `openspec/` and `docs/superpowers/`, everywhere.

- [x] **Step 1: Reduce the harness to two paths**

In `scripts/test-uncommitted-review-package.sh`, delete the `docs/manual-test` components of the
`mkdir -p` lines, the `printf 'manual test notes\n' > "$repo/docs/manual-test/notes.md"` writes, and
the two `docs/manual-test`-specific assertions. Reduce the combined check to the two surviving
paths:

```bash verified:authored in-tree for this change
if printf '%s\n' "$staged" | grep -qE '^(openspec/|docs/superpowers/)'; then
```

- [x] **Step 2: Run the harness and watch it pass**

```bash verified:authored in-tree for this change
scripts/test-uncommitted-review-package.sh
```

Expected: PASS. This step is `green` before the contract edits because the harness exercises the
helper script, not the contract prose.

- [x] **Step 3: Update the git chain in all three files**

The chain appears in `pipeline.md`, `skills/myflow-do/SKILL.md` and `skills/myflow-finish/SKILL.md`.
Each becomes:

```bash verified:authored in-tree for this change
git -C <abs-worktree> reset -q -- openspec/ docs/superpowers/ \
  && git -C <abs-worktree> add -A -- . ':(exclude)openspec/' ':(exclude)docs/superpowers/' \
  && { git -C <abs-worktree> diff --cached --quiet \
       || git -C <abs-worktree> commit -m "<type>(<name>): <what the implementation does>"; } \
  && git -C <abs-worktree> add -A \
  && { git -C <abs-worktree> diff --cached --quiet \
       || git -C <abs-worktree> commit -m "chore(<name>): plan and session records"; }
```

The second commit's subject changes here, in the same edit — Task 2 already taught the guard to
match it. `skills/myflow-do/SKILL.md`'s two bare staging lines take the same two-path treatment.

- [x] **Step 4: Update the prose that counts the paths**

Every sentence saying "the three paths", "those three planning paths" or listing
`docs/manual-test/` alongside the other two now says two and lists two. The symlinked-planning-path
rule in `pipeline.md` keeps its example, with `docs/manual-test/` swapped for `docs/superpowers/`.

- [x] **Step 5: Verify no stale reference survives**

```bash verified:authored in-tree for this change
grep -rn "docs/manual-test" skills/ scripts/ commands/ commands-claude/ rules/
```

Expected: no output.

---

## Task 4: Replace `/myflow-do` section 6 with run-instruction resolution

**Build:** green

**Files:**
- Modify: `skills/myflow-do/SKILL.md:325` onward — section 6, retitled
- Modify: `skills/myflow-do/SKILL.md:536` — the `Test guide:` line of its handoff block
- Modify: `skills/myflow-do/SKILL-rationale.md` — the matching rationale section
- Modify: `scripts/check-contract-budget.sh` — only if a row is exceeded

**Interfaces:**
- Consumes: the two-path staging rule from Task 3.
- Produces: section 6 titled **Resolve the run instructions**, writing no file, feeding the handoff
  block's `Run it:` section defined in Task 5.

- [x] **Step 1: Retitle and rewrite section 6**

Retitle to `## 6. Resolve the run instructions`. It writes no file. It resolves, for the handoff:

- every app root, absolute, from `git worktree list` or the state file's `worktrees` keys;
- every start command from `.myflow/project.md`'s `## run`, with paths absolute;
- every URL from this worktree's workspace id the way section 2 computes it, never the project's
  declared base — a project declaring no isolation resolves nothing and keeps its declared URLs;
- for a project with no runnable application, the `## lint` and `## test` commands, absolute.

Keep, verbatim in substance, the existing bullets on absolute paths, on resolving URLs from the
workspace id rather than from a later-exported variable, and on a project whose port is fixed
outside its own repository keeping its default with a note on the same line. Delete the checklist
rules, the `## Known incomplete` rule, the capability grouping and the tickable-line register.

- [x] **Step 2: Update the skill's handoff block**

Replace the `Test guide:` line and the instruction line beneath it with the `Run it:` shape defined
in Task 5, keeping the citation to **The block each state renders**.

- [x] **Step 3: Update the rationale appendix**

`skills/myflow-do/SKILL-rationale.md`'s section explaining why the guide is written in the same run
becomes the reason the run instructions are resolved in the same run and printed rather than
written: nothing reads them back, so a file would be a record with no consumer.

- [x] **Step 4: Check the budget ratchet**

```bash verified:authored in-tree for this change
scripts/check-contract-budget.sh
```

Expected: PASS. `skills/myflow-do/SKILL.md` shrinks. If any file grew past its row, raise that row
in `scripts/check-contract-budget.sh` in this task, never narrow the guard.

---

## Task 5: Put the run instructions into the handoff template

**Build:** green

**Files:**
- Modify: `skills/myflow-contracts/handoff-blocks.md:108-110` — the `IN_PROGRESS` template
- Modify: `skills/myflow-contracts/handoff-blocks-rationale.md` — if it names the guide
- Modify: `skills/myflow-status/SKILL.md:176` — the detail-view bullet

**Interfaces:**
- Consumes: section 6's resolution rules from Task 4.
- Produces: the one definition both `/myflow-do` and `/myflow-status` render.

- [x] **Step 1: Replace the template's guide lines**

```text verified:authored in-tree for this change
Worktree:   <absolute worktree path>

Run it:
  <command>          # <app or check name>
  <command>

Review the diff, then run it:
  <the review command that matches the git state on the Staged line — see below>
  open -na "IntelliJ IDEA" --args "<absolute worktree path>"
```

The `Run it:` section is **on-disk**, not `(run-only)`: it is resolved from the worktree and the
project's own configuration, so `/myflow-status` regenerates it exactly as `/myflow-do` prints it.

- [x] **Step 2: Drop the guide bullet from the status detail view**

Delete `- The manual test guide's **absolute** path + checked/total box count` from
`skills/myflow-status/SKILL.md` section 4. Add nothing: the regenerated block printed a few lines
later already carries the run instructions.

- [x] **Step 3: Verify both renderers agree**

```bash verified:authored in-tree for this change
scripts/check-references.sh
scripts/check-contract-budget.sh
```

Expected: PASS. Then read `skills/myflow-do/SKILL.md`'s block and the template side by side and
confirm the same folded lines, the same fields and the same order, per **The handoff block is
defined once and rendered by two commands**.

---

## Task 6: Delete `docs/manual-test/` and sweep the remaining contract prose

**Build:** green

**Files:**
- Delete: `docs/manual-test/` — the directory and all fourteen guides
- Modify: `skills/myflow-contracts/pipeline.md:31`, `:38`, `:240`, `:316`
- Modify: `skills/myflow-contracts/finish-contract.md` — the preserved-records sentence
- Modify: `skills/myflow-contracts/plan-provenance.md:280`, `:286` — the quoting-exception table
- Modify: `skills/myflow-contracts/workspace-isolation.md:171`
- Modify: `skills/myflow-contracts/project-configuration.md:197`
- Modify: `scripts/check-references.sh:414` — the comment
- Modify: `scripts/check-plan-provenance.py:588` — the comment
- Modify: `scripts/test-check-references.sh:111` — the fixture path
- Modify: `skills/README.md:60`

**Interfaces:**
- Consumes: Tasks 3-5, which removed every live consumer of the directory.
- Produces: a tree where `manual test guide` appears in no contract, skill or script.

- [x] **Step 1: Delete the directory**

```bash verified:authored in-tree for this change
git rm -r docs/manual-test/
```

- [x] **Step 2: Sweep the contract prose**

`pipeline.md`: the `IN_PROGRESS` row and the gate paragraph say the staged diff awaits review and
the run; the `Link, never paste` bullet names plans and diffs; the IntelliJ table's `IN_PROGRESS`
row names the worktree root alone. `finish-contract.md`'s two-commit sentence names `openspec/` and
the preserved `docs/superpowers/` records. `plan-provenance.md` loses the guide's row from the
quoting-exception table. `workspace-isolation.md` says the bound ports are printed in the
`/myflow-do` handoff. `project-configuration.md`'s aside about which step resolves every row names
the run instructions instead of the guide.

- [x] **Step 3: Fix the two stale comments and the fixture path**

`check-references.sh` and `check-plan-provenance.py` each carry a comment using
`docs/manual-test/<name>.md` as an example of a legitimate shape. Swap both for a shape that still
exists, e.g. `openspec/changes/<name>/tasks.md`. `test-check-references.sh:111` uses the same path
in a fixture — swap it and keep the test.

- [x] **Step 4: Verify the sweep**

```bash verified:authored in-tree for this change
grep -rln "manual test guide\|manual-test" skills/ scripts/ docs/ commands/ commands-claude/ rules/ \
  | grep -v 'docs/superpowers/'
```

Expected: no output. Preserved records under `docs/superpowers/` are historical and are not
rewritten.

```bash verified:authored in-tree for this change
scripts/check-references.sh
scripts/test-check-references.sh
```

Expected: PASS.

---

## Task 7: Update the documentation digests and command files

**Build:** green

**Files:**
- Modify: `CLAUDE.md:59`, `:93`, `:126`
- Modify: `AGENTS.md:105`, `:139`, `:172`
- Modify: `rules/myflow-manual-review.mdc:19`
- Modify: `README.md:58`, `:97`, `:511`
- Modify: `commands/myflow-do.md:5`, `:14`
- Modify: `commands-claude/myflow-do.md:3`, `:10`

**Interfaces:**
- Consumes: Tasks 4-6, which settled the wording the digests summarise.
- Produces: no file outside `openspec/changes/archive/` and `docs/superpowers/` names the guide.

- [x] **Step 1: Rewrite the six digest sites**

Each says `/myflow-do` produces the staged diff and the run instructions, so reviewing and running
are one sitting. Keep the gate wording "review the staged diff and run the apps" — the gate did not
change. Do **not** make the copies byte-identical to each other or to `pipeline.md`: the facts are
fixed, the wording deliberately is not.

- [x] **Step 2: Update both `myflow-do` command files**

Frontmatter `description:` becomes `Do — implement the plan with TDD and the review panel`. The body
line becomes: produces **both** the staged diff **and** the run instructions, so reviewing the code
and running the apps are one human gate.

- [x] **Step 3: Verify the installer still round-trips**

```bash verified:authored in-tree for this change
SANDBOX="$(mktemp -d)"
HOME="$SANDBOX" ./setup.sh global
scripts/test-setup.sh
```

Expected: PASS, and no file under `$SANDBOX` mentions the guide.

---

## Task 8: Full verification pass

**Build:** green

**Files:**
- Modify: whichever file a guard reports — no new file is created by this task

**Interfaces:**
- Consumes: Tasks 1-7.
- Produces: a tree where every declared lint and test command exits 0.

- [x] **Step 1: Run every test harness**

```bash verified:copied from .myflow/project.md's ## test section
scripts/test-setup.sh
scripts/test-check-references.sh
scripts/test-check-plan-provenance.sh
scripts/test-check-finish-preflight.sh
scripts/test-preserve-session-records.sh
scripts/test-check-unfinished-work.sh
scripts/test-check-cleanup-complete.sh
scripts/test-gather-self-review-context.sh
scripts/test-uncommitted-review-package.sh
scripts/test-check-task-build-green.sh
scripts/test-check-workspace-isolation.sh
scripts/test-check-contract-budget.sh
```

Expected: PASS, all twelve.
<!-- measured: ls scripts/test-*.sh @ branch openspec/kan-107-remove-manual-test-guide-and-gate -->

- [x] **Step 2: Run every lint guard**

```bash verified:copied from .myflow/project.md's ## lint section
scripts/check-vocabulary.sh
scripts/check-references.sh
scripts/check-plan-provenance.sh
scripts/check-task-build-green.sh
scripts/check-workspace-isolation.sh
scripts/check-contract-budget.sh
```

Expected: PASS, all six.
<!-- measured: .myflow/project.md ## lint section @ branch openspec/kan-107-remove-manual-test-guide-and-gate -->

Fix every hit by editing the offending line. Never add a suppression marker, never narrow a guard,
never lower a budget's scope. Raising a budget row is the correct response to a file that genuinely
grew.

- [x] **Step 3: Validate the OpenSpec change**

```bash verified:authored in-tree for this change
openspec validate kan-107-remove-manual-test-guide-and-gate --strict
```

Expected: `Change 'kan-107-remove-manual-test-guide-and-gate' is valid`.
