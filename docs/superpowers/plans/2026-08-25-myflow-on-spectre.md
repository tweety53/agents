# myflow on spectre Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move the myflow pipeline off OpenSpec onto `spectre`, leaving the existing `openspec/` tree frozen and unmigrated.

**Architecture:** One cutover on one branch. The guard scripts flip first — they scan a project's change folders, and the new `spectre/` tree is empty, so they pass vacuously while the prose still says OpenSpec. The contracts and skills flip next, and the documentation last. Three integration points are deleted rather than ported: delta specs, `openspec instructions`, and `openspec status --change`.

**Tech Stack:** Bash, Python 3 (standard library only), Markdown. The `spectre` binary (v0.1.0, `github.com/tweety53/spectre`) must be on `PATH`.

**Spec:** `docs/superpowers/specs/2026-08-25-myflow-on-spectre-design.md`

## Global Constraints

- The frozen tree is never touched. `git status --porcelain openspec/` must print nothing at the end of every task.
- Historical artifacts keep saying OpenSpec: `docs/self-review/`, `docs/superpowers/ledgers/`, `docs/superpowers/reviews/`, `docs/superpowers/reports/`, and every pre-existing file in `docs/superpowers/specs/`. Never rewrite them.
- No dual-path support anywhere. Nothing detects which artifact tool a project uses.
- No `spectre/config.md` in this repository. Defaults only.
- Python guards are standard library only. Bash guards stay POSIX-compatible with the existing style.
- Repository markdown follows the brevity rule already in force: prose states a thing once and does not restate what another file states canonically.
- A commit's scope names the module or area moved — `scripts`, `skills`, `contracts`, `commands`, `docs` — never a task id, never a change name, never a list.
- Every task ends with its own named guard tests passing. The FULL suite in `.myflow/project.md`'s `## test` section is only required green at Task 9.

---

## File Structure

| Area | Files | Responsibility |
|------|-------|----------------|
| New tree | `spectre/specs/`, `spectre/changes/` | this repository's own artifact tree, starting empty |
| New skill | `skills/spectre-research/SKILL.md`, `commands/spectre-research.md` | research mode, replacing `openspec-explore` |
| Guards (path) | 11 scripts + tests | change-folder paths: `openspec/changes/` → `spectre/changes/` |
| Guards (artifact) | 5 scripts + tests | plan provenance, build-green tags, commit fields, self-review reports |
| Guards (corpus) | 6 scripts + tests | citations, references, markdown integrity, owned corpus, symlinks |
| Contracts | `skills/myflow-contracts/*.md` (10 of 27 name OpenSpec) | the pipeline's normative text |
| Skills | `skills/myflow-{start,do,finish,fast,status}/SKILL.md` | the commands' behaviour and their CLI calls |
| Distribution | `commands/`, `commands-claude/`, `CLAUDE.md`, `AGENTS.md`, `README.md`, `skills/README.md`, `.myflow/project.md`, `setup.sh` | what installs and what the harnesses read |

**Not implemented:** `spectre refs` is not wired into the pipeline. It has no OpenSpec counterpart and no myflow caller; adding one is a separate change.

---

## Task 1: Create the spectre tree and record the freeze

**Files:**
- Create: `spectre/specs/.gitkeep`, `spectre/changes/.gitkeep`
- Modify: `.myflow/project.md`
- Test: the repository's own `spectre validate` run

**Interfaces:**
- Consumes: nothing.
- Produces: the tree every later task's guards and prose point at — `spectre/specs/` and `spectre/changes/`, with `spectre/changes/archive/` created on first archive.

- [ ] **Step 1: Confirm the binary is present and the tree is absent**

Run:

```bash
cd /Users/tweety53/Projects/agents
spectre --help | head -3
test ! -d spectre && echo "no tree yet, as expected"
```

Expected: the usage banner prints, and `no tree yet, as expected`. If `spectre` is not found, stop and report — every later task depends on it.

- [ ] **Step 2: Create the empty tree**

```bash
mkdir -p spectre/specs spectre/changes
touch spectre/specs/.gitkeep spectre/changes/.gitkeep
```

Git does not track empty directories, and the guards in Tasks 3–5 need both paths to exist in a fresh clone.

- [ ] **Step 3: Verify spectre reads it**

Run:

```bash
spectre validate ; echo "validate exit=$?"
spectre list ; echo "list exit=$?"
spectre list --specs ; echo "specs exit=$?"
```

Expected: `no findings` and exit 0 from validate; empty output and exit 0 from both list forms.

- [ ] **Step 4: Record the freeze in the project file**

In `.myflow/project.md`, add a short subsection under the repository's own description stating: the pipeline's artifact tree is `spectre/`; `openspec/` is frozen at the 2026-08-25 cutover and is never written to again; the two changes left open in it — `kan-295-cut-pipeline-load-cost-split-by-consumer` and `kan-327-review-per-round-delta-after-round-1` — are abandoned. Three sentences, no more; the design document holds the reasoning.

- [ ] **Step 5: Confirm the frozen tree is untouched, then commit**

```bash
git status --porcelain openspec/    # must print nothing
git add spectre .myflow/project.md
git commit -m "feat(spectre): start this repository's own spectre tree"
```

---

## Task 2: Replace explore mode with spectre-research

**Files:**
- Create: `skills/spectre-research/SKILL.md`, `commands/spectre-research.md`
- Delete: `skills/openspec-explore/SKILL.md`, `commands/opsx-explore.md`
- Modify: `setup.sh`, `scripts/check-references.sh`
- Test: `scripts/test-check-references.sh`, `scripts/test-setup.sh`

**Interfaces:**
- Consumes: the tree from Task 1.
- Produces: the skill name `spectre-research` and the command `/spectre-research`, which Task 8's documentation refers to.

- [ ] **Step 1: Read what is being replaced**

Read `skills/openspec-explore/SKILL.md` (290 lines) and `commands/opsx-explore.md` (174 lines). Keep the stance — a thinking partner that investigates, asks, and captures, and never writes application code. Drop everything built on the OpenSpec CLI: the `allowed-tools: Bash(openspec:*)` grant, the status probing, the change-exists branching, and the artifact-creation choreography.

- [ ] **Step 2: Write the new skill**

Create `skills/spectre-research/SKILL.md`. Frontmatter:

```yaml
---
name: spectre-research
description: Research mode - a thinking partner for exploring ideas, investigating problems, and clarifying requirements before or during a change. Touches no pipeline state.
---
```

The body covers, in this order and no more than roughly 120 lines total:

1. **The stance** — investigate and clarify; never write application code; never advance pipeline state; never commit.
2. **Reading the tree** — `spectre list`, `spectre list --specs`, and reading `spectre/specs/*.md` and `spectre/changes/<id>/` directly. No CLI grant is needed to read markdown.
3. **What to do** — read code and specs, ask one question at a time, name the options and the trade-offs, and say plainly when the answer is not yet knowable.
4. **Capturing the outcome** — a summary in the conversation, and, only if the user asks, notes written into an existing change's `design.md`. Creating a change is `/myflow-start`'s job, not this mode's.
5. **Guardrails** — no code, no state, no commits, no `spectre new`, no `spectre archive`.

- [ ] **Step 3: Write the command**

Create `commands/spectre-research.md` following the shape of the other files in `commands/` — read one of the existing `commands/myflow-*.md` files first and match its frontmatter and structure. It invokes the `spectre-research` skill and passes the user's argument through as the topic.

- [ ] **Step 4: Remove the old skill and command**

```bash
git rm -r skills/openspec-explore commands/opsx-explore.md
```

- [ ] **Step 5: Update the installer and the references guard**

In `setup.sh`, replace every `openspec-explore` and `opsx-explore` entry with the new names. In `scripts/check-references.sh`, the allow-list at roughly line 572 names `skills/openspec-explore/SKILL.md`; replace it with `skills/spectre-research/SKILL.md`.

- [ ] **Step 6: Verify**

Run:

```bash
scripts/test-check-references.sh
scripts/test-setup.sh
grep -rn 'openspec-explore\|opsx-explore' --include='*.sh' --include='*.md' skills commands commands-claude setup.sh scripts || echo "no stale references"
git status --porcelain openspec/
```

Expected: both test scripts pass; `no stale references`; nothing from the `openspec/` status.

- [ ] **Step 7: Commit**

```bash
git add -A skills commands setup.sh scripts/check-references.sh
git commit -m "feat(skills): replace explore mode with spectre-research"
```

---

## Task 3: Flip the path guards

**Files:**
- Modify: `scripts/prepare-workspace.sh`, `scripts/commit-split.sh`, `scripts/check-stage-mark-calls.sh`, `scripts/check-cleanup-complete.sh`, `scripts/check-unfinished-work.sh`, `scripts/check-worktree-processes.sh`, `scripts/gather-dispatch-context.sh`, `scripts/gather-self-review-context.sh`, `scripts/plan-dispatch-bundles.sh`, `scripts/plan-dispatch-bundles.py`, `scripts/resolve-base-branch.sh`
- Test: `scripts/test-prepare-workspace.sh`, `scripts/test-commit-split.sh`, `scripts/test-check-cleanup-complete.sh`, `scripts/test-check-unfinished-work.sh`, `scripts/test-check-worktree-processes.sh`, `scripts/test-gather-dispatch-context.sh`, `scripts/test-gather-self-review-context.sh`, `scripts/test-plan-dispatch-bundles.sh`, `scripts/test-resolve-base-branch.sh`, `scripts/test-prepare-archive-branch.sh`, `scripts/test-check-stage-mark-calls.sh`

**Interfaces:**
- Consumes: the tree from Task 1.
- Produces: the canonical strings every later task uses — `spectre/changes/<name>/` for a change folder, `spectre/changes/archive/` for archived ones, `spectre/specs/<capability>.md` for a capability spec. Note the capability path is a FLAT FILE, not `specs/<capability>/spec.md` as OpenSpec had it.

These guards take a change's path and inspect it. They do not read this repository's own tree, so flipping them while the prose still says OpenSpec is safe.

- [ ] **Step 1: See the true extent before changing anything**

```bash
cd /Users/tweety53/Projects/agents
for f in prepare-workspace.sh commit-split.sh check-stage-mark-calls.sh check-cleanup-complete.sh \
         check-unfinished-work.sh check-worktree-processes.sh gather-dispatch-context.sh \
         gather-self-review-context.sh plan-dispatch-bundles.sh plan-dispatch-bundles.py resolve-base-branch.sh; do
  printf '%-34s %s\n' "$f" "$(grep -c openspec scripts/$f)"
done
```

Record these counts in your report; they are what your final grep must drive to zero.

- [ ] **Step 2: Confirm the tests fail for the right reason first**

Pick `scripts/test-check-unfinished-work.sh` and change ONLY its fixture paths from `openspec/changes` to `spectre/changes`, then run it. It must FAIL, because the guard still looks under `openspec/`. That failure is what proves the test actually exercises the path you are about to change.

Run: `scripts/test-check-unfinished-work.sh`
Expected: FAIL naming the missing change folder.

- [ ] **Step 3: Flip guard and test together, one pair at a time**

For each script and its test, replace `openspec/changes` with `spectre/changes` and `openspec/specs/<x>/spec.md` shapes with `spectre/specs/<x>.md`. Work one pair at a time and run that pair's test before moving to the next — a single sweeping `sed` across all 22 files at once will silently rewrite prose in comments that should have been reworded instead.

Read each comment you touch. Where a comment explains WHY a path is excluded — `scripts/check-references.sh` has one about `openspec/changes/<name>/tasks.md` being legitimate — the reasoning must be reworded for spectre, not just path-substituted.

- [ ] **Step 4: Verify the group**

```bash
for t in test-prepare-workspace test-commit-split test-check-cleanup-complete test-check-unfinished-work \
         test-check-worktree-processes test-gather-dispatch-context test-gather-self-review-context \
         test-plan-dispatch-bundles test-resolve-base-branch test-prepare-archive-branch test-check-stage-mark-calls; do
  scripts/$t.sh >/dev/null 2>&1 && echo "PASS $t" || echo "FAIL $t"
done
grep -l openspec scripts/prepare-workspace.sh scripts/commit-split.sh scripts/check-stage-mark-calls.sh \
  scripts/check-cleanup-complete.sh scripts/check-unfinished-work.sh scripts/check-worktree-processes.sh \
  scripts/gather-dispatch-context.sh scripts/gather-self-review-context.sh scripts/plan-dispatch-bundles.sh \
  scripts/plan-dispatch-bundles.py scripts/resolve-base-branch.sh || echo "group clean"
git status --porcelain openspec/
```

Expected: every line `PASS`, then `group clean`, then nothing from the `openspec/` status.

- [ ] **Step 5: Commit**

```bash
git add scripts
git commit -m "refactor(scripts): point the path guards at the spectre tree"
```

---

## Task 4: Flip the artifact-content guards

**Files:**
- Modify: `scripts/check-plan-provenance.sh`, `scripts/check-plan-provenance.py`, `scripts/check-task-build-green.sh`, `scripts/check-task-commit-fields.sh`, `scripts/check-task-commit-fields.py`, `scripts/check-self-review-report.sh`
- Test: `scripts/test-check-plan-provenance.sh`, `scripts/test-check-task-build-green.sh`, `scripts/test-check-task-commit-fields.sh`, `scripts/test-check-self-review-report.sh`

**Interfaces:**
- Consumes: the canonical paths established in Task 3.
- Produces: guards that read `spectre/changes/<name>/tasks.md` and `proposal.md` for provenance tags, build-green tags and commit fields.

These read a change's artifact CONTENT rather than just its path. The artifact shapes themselves are unchanged — `tasks.md` still carries `- [ ] 1.` lines with `**Build:**` and `**Squash-with:**` fields — so only the containing paths move.

- [ ] **Step 1: Confirm the artifact shapes are genuinely unchanged**

```bash
grep -n 'Build:\|Squash-with:' skills/myflow-contracts/plan-provenance.md | head
spectre validate --root . ; echo "exit=$?"
```

spectre's own task-line rule is `- [ ] <n>. <text>`, which is what these guards already expect. If any guard depends on OpenSpec's nested `1.1` numbering, note it in your report — spectre flattens numbering, and that guard needs its expectation changed rather than its path.

- [ ] **Step 2: Flip each guard with its test**

Same discipline as Task 3: one pair at a time, test after each. Both `.sh` wrappers and their `.py` implementations must move together — `check-plan-provenance.sh` execs `check-plan-provenance.py`, and a path changed in only one of them fails confusingly.

- [ ] **Step 3: Verify**

```bash
for t in test-check-plan-provenance test-check-task-build-green test-check-task-commit-fields test-check-self-review-report; do
  scripts/$t.sh >/dev/null 2>&1 && echo "PASS $t" || echo "FAIL $t"
done
grep -l openspec scripts/check-plan-provenance.* scripts/check-task-build-green.sh scripts/check-task-commit-fields.* scripts/check-self-review-report.sh || echo "group clean"
git status --porcelain openspec/
```

Expected: four `PASS` lines, `group clean`, nothing from `openspec/`.

- [ ] **Step 4: Commit**

```bash
git add scripts
git commit -m "refactor(scripts): read change artifacts from the spectre tree"
```

---

## Task 5: Flip the corpus guards

**Files:**
- Modify: `scripts/check-references.sh`, `scripts/check-installed-citations.py`, `scripts/check-markdown-integrity.py`, `scripts/check-vocabulary.sh`, `scripts/lib/owned-corpus.sh`, `scripts/check-guard-symlinks.sh`
- Leave alone deliberately: `scripts/check-contract-budget.sh`
- Test: `scripts/test-check-references.sh`, `scripts/test-check-installed-citations.sh`, `scripts/test-check-markdown-integrity.sh`, `scripts/test-check-vocabulary.sh`, `scripts/test-check-guard-symlinks.sh`, `scripts/test-check-normative-inventory.sh`, `scripts/test-lib-coverage.sh`, `scripts/test-check-contract-budget.sh`

**Interfaces:**
- Consumes: the canonical paths from Task 3.
- Produces: a corpus definition that includes `spectre/` and excludes the frozen `openspec/`, which Tasks 6–8 rely on when they rewrite prose.

- [ ] **Step 1: Understand what the corpus means now**

Read `scripts/lib/owned-corpus.sh` and `scripts/check-markdown-integrity.py`'s header comment, which says integrity covers the repository's own markdown and "not `openspec/`". That exclusion is now permanent and correct — the frozen tree must never be linted again. Make sure the exclusion survives your edit rather than being path-substituted into an exclusion of `spectre/`, which would be exactly backwards.

- [ ] **Step 2: Drop the frozen tree's rows from `check-contract-budget.sh`**

Its baseline table lists files under `openspec/specs/` with byte sizes, and five of those rows already FAIL on `main` — `myflow-contract-economy`, `myflow-run-record` and `myflow-state-store` over budget, `myflow-daemon-single-instance` and `myflow-stats-build` undeclared. Confirm that for yourself first:

```bash
scripts/check-contract-budget.sh; echo "exit=$?"
```

A frozen tree has no budget to enforce: nothing in it can grow. Remove every `openspec/specs/` row from `budgets()`, and make the guard skip that path rather than requiring rows for it. Say in a comment that the tree is frozen, so a future reader does not re-add the rows. Rows for files this repository still edits stay exactly as they are.

Verify the guard now passes on the real repository, not only its test harness:

```bash
scripts/check-contract-budget.sh && echo "budget guard green on the repo"
scripts/test-check-contract-budget.sh && echo "and its tests pass"
```

- [ ] **Step 3: Flip the remaining five, with their tests**

One pair at a time. In `check-vocabulary.sh`, lines 62–66 carry a comment explaining that scanning `openspec/specs` was tried and reverted. That is a record of a decision about the OLD tree; keep its meaning, and note that the frozen tree is not scanned because it is frozen, not because scanning was reverted.

- [ ] **Step 4: Verify**

```bash
for t in test-check-references test-check-installed-citations test-check-markdown-integrity \
         test-check-vocabulary test-check-guard-symlinks test-check-normative-inventory \
         test-lib-coverage test-check-contract-budget; do
  scripts/$t.sh >/dev/null 2>&1 && echo "PASS $t" || echo "FAIL $t"
done
git status --porcelain openspec/
```

Expected: eight `PASS` lines and nothing from `openspec/`.

- [ ] **Step 5: Commit**

```bash
git add scripts
git commit -m "refactor(scripts): define the owned corpus around the spectre tree"
```

---

## Task 6: Rewrite the contracts

**Files:**
- Modify: `skills/myflow-contracts/pipeline.md`, `finish-contract.md`, `artifacts-registry.md`, `state-file.md`, `git-boundaries.md`, `plan-provenance.md`, `operator-prompts.md`, `model-policy.md`, `model-policy-rationale.md`, `pipeline-rationale.md`
- Test: `scripts/test-check-references.sh`, `scripts/test-check-markdown-integrity.sh`, `scripts/test-check-contract-budget.sh`, `scripts/test-check-normative-inventory.sh`, `scripts/test-check-vocabulary.sh`

**Interfaces:**
- Consumes: the paths from Tasks 3–5.
- Produces: the normative text Task 7's skills cite. In particular it fixes what `/myflow-finish` run 2 does, which the finish skill then follows.

This task carries the design's substance. Two files change meaning; the other eight change vocabulary.

- [ ] **Step 1: Delete delta specs from the finish contract**

In `skills/myflow-contracts/finish-contract.md`, step 3 of run 2 currently reads:

> 3. **Sync delta specs** into `<project>/openspec/specs/`, then move the change into
>    `<project>/openspec/changes/archive/<date>-<name>/`. Any nested `<name>-fix-N` sub-changes are archived in
>    the same operation — never left behind, never archived alone.

Rewrite it so that the sync half is gone entirely and the archive half remains. Under spectre a change edits `spectre/specs/<capability>.md` directly on its branch, so by the time run 2 executes, the specs are already correct on the base branch — there is nothing to merge back. The replacement states: archive the change into `<project>/spectre/changes/archive/<date>-<name>/`, with nested `<name>-fix-N` sub-changes archived in the same operation, never left behind and never archived alone. Say in one clause WHY there is no sync step, so a future reader does not re-add it.

Scan the whole file for other delta-spec assumptions — preflight assertions, refusal conditions, the cleanup checklist — and remove each one on the same reasoning.

- [ ] **Step 2: Point the state machine at the new tree**

In `skills/myflow-contracts/pipeline.md`, three places name OpenSpec:

- roughly line 265: `/myflow-do` never stages `<project>/openspec/` — becomes `<project>/spectre/`.
- roughly line 466: the union of names includes what `openspec list --json` reports — becomes `spectre list --json`, whose shape is `{"changes":[{"id","done","total"}]}`, so the field read is `id`.
- roughly lines 473–474: names dropped once `<project>/openspec/changes/<name>/` reaches `.../archive/` — the same sentence over `spectre/changes/`.

- [ ] **Step 3: Rewrite the artifact registry**

`artifacts-registry.md` describes what a change folder holds. Under spectre that is `proposal.md`, `tasks.md`, and optionally `design.md` — and NOT a `specs/` subdirectory of delta files. Remove the delta entry; add, in one sentence, that a change's spec edits live in `spectre/specs/` on the branch.

- [ ] **Step 4: Vocabulary in the remaining seven**

`state-file.md`, `git-boundaries.md`, `plan-provenance.md`, `operator-prompts.md`, `model-policy.md`, `model-policy-rationale.md` and `pipeline-rationale.md` name OpenSpec in prose and paths. Replace both. Where a rationale file records WHY a past decision was made about OpenSpec, keep the record truthful: it explains history, so it may name OpenSpec in the past tense rather than being rewritten to claim the decision was about spectre.

- [ ] **Step 5: Verify**

```bash
grep -rn 'openspec\|OpenSpec' skills/myflow-contracts/ | grep -v 'past tense record' || echo "contracts clean"
grep -rn 'delta spec' skills/myflow-contracts/ || echo "no delta specs remain"
for t in test-check-references test-check-markdown-integrity test-check-contract-budget \
         test-check-normative-inventory test-check-vocabulary; do
  scripts/$t.sh >/dev/null 2>&1 && echo "PASS $t" || echo "FAIL $t"
done
git status --porcelain openspec/
```

Expected: any surviving OpenSpec mention is a deliberate past-tense record you can name; no delta-spec mentions; five `PASS` lines; nothing from `openspec/`.

- [ ] **Step 6: Commit**

```bash
git add skills/myflow-contracts
git commit -m "feat(contracts): drop delta specs and move the pipeline to spectre"
```

---

## Task 7: Rewrite the five myflow skills

**Files:**
- Modify: `skills/myflow-start/SKILL.md`, `skills/myflow-do/SKILL.md`, `skills/myflow-finish/SKILL.md`, `skills/myflow-fast/SKILL.md`, `skills/myflow-status/SKILL.md`
- Test: `scripts/test-check-references.sh`, `scripts/test-check-installed-citations.sh`, `scripts/test-check-markdown-integrity.sh`, `scripts/test-check-stage-mark-calls.sh`, `scripts/test-check-contract-budget.sh`

**Interfaces:**
- Consumes: the contracts from Task 6.
- Produces: the command behaviour Task 8's documentation describes.

- [ ] **Step 1: Replace `/myflow-start`'s creation block**

`skills/myflow-start/SKILL.md` around line 294 currently runs:

```bash
openspec new change "<name>"
openspec status --change "<name>" --json
openspec instructions <artifact-id> --change "<name>" --json
```

Replace with a single call:

```bash
spectre new "<name>"
```

`spectre new` refuses when the tree does not exist and refuses an id that is not a single flat directory name, so its failure modes are the ones worth stating. There is no `status` call — the change root is `<project>/spectre/changes/<name>/` by construction — and no `instructions` call, so the artifact list that `applyRequires` used to supply is stated here instead:

- **proposal.md** — what and why, carrying `## Why` and `## What changes`
- **tasks.md** — the plan, `- [ ] <n>. <text>` lines
- **design.md** — how, from the approved design, including `## Decisions` and `## Open questions`

The old "specs/ — delta specs, one file per capability named in the proposal" entry is DELETED. Spec edits go directly into `<project>/spectre/specs/<capability>.md` on the change branch; say that in one sentence where the deleted entry was.

- [ ] **Step 2: Replace `/myflow-do`'s context block**

`skills/myflow-do/SKILL.md` around line 74 currently runs `openspec status --change` and `openspec instructions apply --change`, then says "Read every path in `contextFiles`. Resolve paths from the CLI JSON, never from an assumed layout."

Under spectre the layout IS the contract, so that instruction inverts. Replace the two calls with `spectre validate "<name>"`, and state the context files as a fixed list: the change's `proposal.md`, `tasks.md`, `design.md` when present, and the capability files under `spectre/specs/` that the change's proposal names. The `state: "blocked"` and `state: "all_done"` branches are replaced by reading the task checkboxes — `spectre list --json` reports `done` and `total` per change, so `done == 0` and `done == total` are the two signals.

Around line 196, `<changeRoot>` is described as coming from `openspec status` output. It becomes `<project>/spectre/changes/<name>/`.

- [ ] **Step 3: Rewrite `/myflow-finish`'s archive half**

`skills/myflow-finish/SKILL.md` must match Task 6's finish contract: run 2 archives with `spectre archive "<name>"` and commits on `chore/archive-<name>`; there is no delta-spec sync. Note that `spectre archive` refuses a change with unchecked tasks, with no `tasks.md`, or with no tasks at all — all three overridable with `--force` — and refuses a destination collision, which `--force` does NOT override. The pipeline should never need `--force`; if it does, that is a signal the change is not finished.

- [ ] **Step 4: Update `/myflow-status` and `/myflow-fast`**

`myflow-status` reports non-archived changes: `spectre list --json` returns exactly those, so the enumeration becomes that call, reading `id`, `done` and `total`. `myflow-fast` chains the other three and needs only its OpenSpec vocabulary replaced.

- [ ] **Step 5: Verify**

```bash
grep -rn 'openspec\|OpenSpec' skills/myflow-*/SKILL.md || echo "skills clean"
grep -rn 'delta spec\|applyRequires\|contextFiles' skills/myflow-*/SKILL.md || echo "no CLI-supplied context remains"
for t in test-check-references test-check-installed-citations test-check-markdown-integrity \
         test-check-stage-mark-calls test-check-contract-budget; do
  scripts/$t.sh >/dev/null 2>&1 && echo "PASS $t" || echo "FAIL $t"
done
git status --porcelain openspec/
```

Expected: `skills clean`, `no CLI-supplied context remains`, five `PASS` lines, nothing from `openspec/`.

- [ ] **Step 6: Commit**

```bash
git add skills
git commit -m "feat(skills): run the pipeline on spectre"
```

---

## Task 8: Documentation and distribution

**Files:**
- Modify: `commands/myflow-{start,do,finish,fast,status}.md`, `commands-claude/myflow-{start,do,finish,fast,status}.md`, `CLAUDE.md`, `AGENTS.md`, `README.md`, `skills/README.md`, `.myflow/project.md`, `setup.sh`
- Test: the full suite named in `.myflow/project.md`'s `## test` section

**Interfaces:**
- Consumes: everything from Tasks 1–7.
- Produces: the installed surface every harness reads.

- [ ] **Step 1: The ten slash commands**

Each of the ten `myflow-*.md` files across `commands/` and `commands-claude/` names OpenSpec. Replace the vocabulary and any path. These files are thin — read one before editing to match its shape.

- [ ] **Step 2: The always-on rule text**

`CLAUDE.md` and `AGENTS.md` carry the myflow trigger and the contract table. Update the artifact-tool vocabulary. The contract table's rows and their load-before conditions do not change — only the tool named in the prose does.

- [ ] **Step 3: `README.md` and `skills/README.md`**

`README.md`'s tree diagram and its "How the pipeline works" section describe OpenSpec, including `openspec/` in the directory listing. The listing gains `spectre/` and keeps `openspec/` marked as frozen. `skills/README.md`'s command map names `openspec-explore`; it becomes `spectre-research`.

- [ ] **Step 4: `.myflow/project.md`**

Beyond Task 1's freeze note, this file's guard descriptions and its citation at roughly line 25 name OpenSpec paths. The citation at line 25 points into `openspec/changes/kan-14-plan-provenance/design.md` — a frozen, still-valid historical document. Keep the citation pointing there and say it is historical; do not repoint it at a spectre path that does not exist.

- [ ] **Step 5: `setup.sh`**

Beyond Task 2's skill rename, the installer's file lists and any OpenSpec-specific install steps need updating. Run the sandboxed installer test after.

- [ ] **Step 6: Verify the whole suite**

Run every script in `.myflow/project.md`'s `## test` section, in order. All must pass. Then:

```bash
grep -rn 'openspec\|OpenSpec' --include='*.md' --include='*.sh' --include='*.py' --include='*.mdc' \
  skills commands commands-claude scripts rules setup.sh CLAUDE.md AGENTS.md README.md .myflow \
  | grep -v 'frozen\|historical\|past' || echo "live surface clean"
git status --porcelain openspec/
```

Expected: every test passes; any remaining mention is explicitly marked frozen or historical; nothing from `openspec/`.

- [ ] **Step 7: Commit**

```bash
git add -A
git commit -m "docs: describe the pipeline in terms of spectre"
```

---

## Task 9: Prove the pipeline works end to end

**Files:**
- Create: nothing in this repository. A scratch project outside it.
- Test: the pipeline itself.

**Interfaces:**
- Consumes: everything.
- Produces: the evidence that this cutover works rather than merely lints.

A green guard suite proves the repository is self-consistent. It does not prove a `/myflow-*` run succeeds. This task is the difference.

- [ ] **Step 1: Build a scratch project**

Outside both repositories, create a git repository with a spectre tree, one capability spec and one trivial source file:

```bash
cd "$(mktemp -d)" && git init -q
mkdir -p spectre/specs spectre/changes src
printf '# greeting\n\n## Purpose\nHow the tool greets.\n\n## Requirements\n- R1: The tool SHALL print a greeting on start.\n' > spectre/specs/greeting.md
printf 'print("hello")\n' > src/main.py
git add -A && git commit -qm "init"
spectre validate && echo "scratch tree valid"
```

- [ ] **Step 2: Run the pipeline**

Run `/myflow-start`, then `/myflow-do`, then `/myflow-finish` twice, against a small change in that scratch project — for example, adding a second requirement and the line of code that satisfies it. Follow each command's own handoff.

- [ ] **Step 3: Record what actually happened**

For each command, record in your report: what it did, whether its guards ran, and anything it got wrong. A command that needed manual correction is a finding, not a success — name it precisely enough that a fix round can act on it.

- [ ] **Step 4: Confirm the end state**

```bash
spectre list                       # the change is gone from the open list
ls spectre/changes/archive/        # it is here instead
spectre validate; echo "exit=$?"   # clean, exit 0
git log --oneline | head
```

Expected: the change archived, the tree valid, and the commit history showing the task commits and the archive commit.

- [ ] **Step 5: Confirm this repository is unchanged by the exercise**

```bash
cd /Users/tweety53/Projects/agents
git status --porcelain            # nothing
git status --porcelain openspec/  # nothing
```

- [ ] **Step 6: Report**

No commit. This task's output is the report: the pipeline ran end to end on spectre, or it did not and here is exactly where it broke.
