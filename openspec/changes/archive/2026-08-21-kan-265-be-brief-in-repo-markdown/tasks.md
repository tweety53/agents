> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

## 1. The standard

### 1.1 Reword `be-brief.mdc` and add the repository-prose clause

**Build:** green
**Files:** `rules/be-brief.mdc`
**Tests:** none — prose-only change to a rule file, no executable surface.
**Regression:** not applicable — no harness for rule prose. `scripts/check-vocabulary.sh` and
`scripts/check-markdown-integrity.py` are the regression signals; both must stay green.
**Baseline:** not applicable — no test harness for a rule file.
**Commit:** `feat(be-brief): state brevity for repository prose and generated artifacts`

Steps:

- [x] Reword the existing sentence `**Prose only** — code, commits, docs and specs stay full.` so it
      reads as a **completeness** requirement — every requirement, scenario, exit code and worked
      example present — rather than as an exemption from brevity.
- [x] Add a clause naming exactly two subjects: **this repository** (by name — see decision
      `clause-two-subjects` in `design.md` for why it is not generalised) and **the artifacts a
      `/myflow-*` run generates in any project** (`proposal.md`, `design.md`, `tasks.md`, delta
      specs, panel records, self-review reports).
- [x] State explicitly that another project's pre-existing documentation is not subject to the
      clause.
- [x] State brevity as **non-repetition**: a file states a thing once and does not restate what
      another file states canonically.
- [x] State that **deletion** is permitted and **paraphrase** is not, with the reason: a normative
      sentence re-said differently is a changed requirement no guard detects.
- [x] List what is never cut: any normative sentence, exit-code contract, ordering constraint,
      scenario, worked example, or recorded reason a rejected alternative was rejected.
- [x] Keep the clause inside the file's `<!-- core -->` markers if and only if it belongs in every
      session's prompt; check where the existing markers sit before deciding.
- [x] Verify: `scripts/check-vocabulary.sh && python3 scripts/check-markdown-integrity.py`

## 2. The proof

### 2.1 Add `check-normative-inventory.sh` and its harness

**Build:** green
**Files:** `scripts/check-normative-inventory.sh`, `scripts/test-check-normative-inventory.sh`,
`scripts/lib/owned-corpus.sh`, `.myflow/project.md`
**Tests:** `scripts/test-check-normative-inventory.sh` — cases for: a deleted SHALL sentence
changing the inventory, a reflowed sentence not changing it, a reworded sentence changing it,
order-independence across two runs, an excluded tree contributing nothing, and exit 2 on an
unreadable scope root.
**Regression:** reverting this task removes the only check that would notice a deleted requirement;
every trim task below loses its verification command.
**Baseline:** `before=0 after=6` cases in the new harness.
<!-- predicted: scripts/test-check-normative-inventory.sh after task 2.1 -->
**Commit:** `feat(check-normative-inventory): inventory the corpus's normative sentences`

Steps:

- [x] Write the guard per **A guard inventories the corpus's normative sentences**
      (`openspec/changes/kan-265-be-brief-in-repo-markdown/specs/agents-repo-verification/spec.md`).
- [x] Resolve the corpus **the same way `scripts/check-contract-budget.sh` will after task 9.1** —
      read that script rather than re-deriving the rule, so the two cannot disagree about ownership.
      Until 9.1 lands, factor the resolution so both can call it.
- [x] Keywords are `SHALL`, `SHALL NOT`, `MUST`, `MUST NOT` as whole words only. Not `SHOULD`, not
      `MAY`.
- [x] Normalise internal whitespace to single spaces and strip the ends. Normalise nothing else.
- [x] Print one sentence per line, sorted, so two runs are `diff`-comparable.
- [x] Exit codes: `0` printed, `2` cannot answer. No violation code — it reports a set.
- [x] Honor `CHECK_NORMATIVE_INVENTORY_ROOT` as an opt-in override for the harness alone.
- [x] Add the guard to `## lint` and the harness to `## test` in `.myflow/project.md`.
- [x] Prove each harness case fails against a deliberately broken guard before trusting it.
- [x] Verify: `scripts/test-check-normative-inventory.sh && scripts/check-normative-inventory.sh | wc -l`

### 2.2 Record the pre-trim inventory as this change's baseline

**Build:** green
**Files:** none — the parent performs this step; no implementer commit is made.
**Allowed-collateral:** `openspec/changes/kan-265-be-brief-in-repo-markdown/normative-baseline.txt`
**Tests:** none — the artifact is data, not behaviour.
**Regression:** not applicable — removing the baseline leaves tasks 4.1–7.1 with nothing to diff
against.
**Baseline:** not applicable — no test harness for a captured artifact.
**Commit:** none — `/myflow-finish` commits it with the rest of `openspec/`.

**This task is performed by the parent, not dispatched.** The baseline lives in the change
directory, and COMMIT-PER-TASK forbids an implementer from committing anything under `openspec/`.
The two cannot both hold, so the parent captures the file and `finish.commit-two` commits it. The
original plan declared an implementer commit here; that was a plan defect, corrected in place.

Steps:

- [x] Run `scripts/check-normative-inventory.sh > openspec/changes/kan-265-be-brief-in-repo-markdown/normative-baseline.txt`
      **before any file in tasks 4.1–7.1 is edited**.
- [x] Confirm the capture equals a fresh run of the guard against the same tree.
- [x] Record the line count: **1078 normative sentences from 92 files**.
      <!-- measured: scripts/check-normative-inventory.sh @ branch openspec/kan-265-be-brief-in-repo-markdown -->
- [x] Confirm the keyword reconciliation: 1344 occurrences in the corpus, 1344 in the inventory.
      <!-- measured: grep -rhoE keyword count vs the same count over the inventory @ branch openspec/kan-265-be-brief-in-repo-markdown -->

## 3. Trim — the rule that governs tasks 4 through 8

Every task in sections 4–8 follows the same procedure and the same acceptance test. It is stated
once here rather than repeated in each task, which is the standard this change is about.

- Cut only. Delete a passage; never reword one.
- A passage may be deleted as restatement only if the file that still carries the statement can be
  named. Record that file in the commit body.
- After editing, `scripts/check-normative-inventory.sh` must be byte-identical to
  `normative-baseline.txt`. A difference is resolved by restoring the sentence, never by updating
  the baseline.
- The full `## lint` set and the full `## test` set must be green before the task's commit.
- Do not touch `scripts/check-contract-budget.sh`'s table; task 9.1 rewrites it once, last.

## 4. Trim — the largest runtime-loaded text

### 4.1 `skills/myflow-do/` and `skills/myflow-contracts/`

**Build:** green
**Files:** the files this trim actually touches, all under the two globs below.
**Allowed-collateral:** `skills/myflow-do/*.md`, `skills/myflow-do/**/*.md`,
`skills/myflow-contracts/*.md`
**Tests:** none — prose-only. Verified by the inventory diff and the existing lint and test sets.
**Regression:** not applicable — `scripts/check-references.sh`,
`scripts/check-installed-citations.sh`, `scripts/check-markdown-integrity.py` and
`scripts/check-normative-inventory.sh` are the regression signals.
**Baseline:** not applicable — no test harness for prose content.
**Commit:** `refactor(skills): cut restatement from the myflow-do and contracts text`

Steps:

- [x] Apply section 3's procedure to `skills/myflow-do/SKILL.md` (71 KB), its rationale sibling and
      its reviewer-prompt files.
      <!-- measured: wc -c skills/myflow-do/SKILL.md @ branch main -->
- [x] Apply it to every `*.md` under `skills/myflow-contracts/`, `pipeline.md` (47 KB) and
      `finish-contract.md` (41 KB) included.
      <!-- measured: wc -c skills/myflow-contracts/pipeline.md skills/myflow-contracts/finish-contract.md @ branch main -->
- [x] Leave the core/rationale partitions exactly where they are — repartitioning is out of scope
      per `design.md`'s `cuts-not-paraphrase` and the proposal's non-goals.
- [x] Verify: `diff <(scripts/check-normative-inventory.sh) openspec/changes/kan-265-be-brief-in-repo-markdown/normative-baseline.txt`

### 4.2 The remaining `skills/`

**Build:** green
**Files:** the files this trim actually touches, all under the globs below.
**Allowed-collateral:** `skills/myflow-start/*.md`, `skills/myflow-finish/*.md`,
`skills/myflow-fast/*.md`, `skills/myflow-status/*.md`, `skills/openspec-explore/*.md`,
`skills/*/**/*.md`
**Tests:** none — prose-only, as 4.1.
**Regression:** not applicable — same guards as 4.1.
**Baseline:** not applicable — no test harness for prose content.
**Commit:** `refactor(skills): cut restatement from the remaining skill text`

Steps:

- [x] Apply section 3's procedure to each remaining skill directory.
- [x] Watch for passages that restate `skills/myflow-contracts/` content — those cite the contract
      already, and the citation is what survives.
- [x] Verify: `diff <(scripts/check-normative-inventory.sh) openspec/changes/kan-265-be-brief-in-repo-markdown/normative-baseline.txt`

### 4.3 `skills/README.md` — the file no task's scope reached

**Build:** green
**Files:** `skills/README.md`
**Tests:** none — prose-only, as 4.1.
**Regression:** not applicable — same guards as 4.1.
**Baseline:** not applicable — no test harness for prose content.
**Commit:** `refactor(skills): cut restatement from the skills README`

Task 4.2 found this file matches no glob in any trim task — not 4.1's, not its own, and not
tasks 5.1 through 7.1's. It is 5832 bytes inside the owned corpus and the inventory already scans
it, so it would have reached task 8.1's ratchet without any task having read it. A file nobody looked
at is a gap in the trim, not a file with nothing to cut.

Steps:

- [x] Apply section 3's procedure to `skills/README.md`.
- [x] If nothing in it is cuttable, say so explicitly and commit nothing — an empty commit is not
      required, and "read, nothing to cut" is a valid outcome that must be *reported* rather than
      left silent.
- [x] Verify: `diff <(scripts/check-normative-inventory.sh) openspec/changes/kan-265-be-brief-in-repo-markdown/normative-baseline.txt`

## 5. Trim — rules, commands and project configuration

### 5.1 `rules/`, `commands/`, `commands-claude/`, `.myflow/`

**Build:** green
**Files:** the files this trim actually touches, all under the globs below.
**Allowed-collateral:** `rules/*.mdc`, `rules/agent-baseline.md`, `commands/*.md`,
`commands-claude/*.md`, `.myflow/project.md`
**Tests:** none — prose-only, as 4.1.
**Regression:** not applicable — same guards as 4.1, plus `scripts/check-workspace-isolation.sh`
and `scripts/check-uitest-overrides.sh`, which read `.myflow/project.md`'s declared sections.
**Baseline:** not applicable — no test harness for prose content.
**Commit:** `refactor(rules): cut restatement from the rule, command and project text`

Steps:

- [x] Apply section 3's procedure. `rules/be-brief.mdc` was already edited in task 1.1 — trim it
      here like any other file, without undoing that clause.
- [x] `rules/agent-baseline.md` is one line per rule by design; cutting a row would remove a rule
      from every dispatched agent. Trim its prose, never its table.
- [x] `.myflow/project.md`'s `## lint`, `## test`, `## run` and workspace-isolation sections are
      read by guards. Trim commentary; leave every declared entry.
- [x] Verify: `diff <(scripts/check-normative-inventory.sh) openspec/changes/kan-265-be-brief-in-repo-markdown/normative-baseline.txt`

## 6. Trim — the specs

### 6.1 `openspec/specs/`

**Build:** green
**Files:** the files this trim actually touches, all under the glob below.
**Allowed-collateral:** `openspec/specs/*.md`, `openspec/specs/**/*.md`
**Tests:** none — prose-only, as 4.1.
**Regression:** not applicable — same guards as 4.1. `openspec validate --strict` on every open
change is an additional signal.
**Baseline:** not applicable — no test harness for prose content.
**Commit:** `refactor(openspec): cut restatement from the specs`

Steps:

- [x] Apply section 3's procedure across all 31 spec files, 462 KB, the largest unratcheted block.
      <!-- measured: wc -c openspec/specs/**/*.md @ branch main -->
- [x] A `#### Scenario:` block is never cut — it is a scenario, which section 3 protects.
- [x] A requirement's `### Requirement:` heading is never reworded: it is the match key an archive
      uses against a delta's MODIFIED header, and a reworded heading aborts a later archive.
- [x] Trim `## Purpose` prose and inter-requirement commentary, not requirement bodies.
- [x] Verify: `diff <(scripts/check-normative-inventory.sh) openspec/changes/kan-265-be-brief-in-repo-markdown/normative-baseline.txt`

## 7. Trim — the root files

### 7.1 `README.md` and the other root Markdown

**Build:** green
**Files:** the files this trim actually touches, all under the glob below.
**Allowed-collateral:** `*.md`
**Tests:** none — prose-only, as 4.1.
**Regression:** not applicable — same guards as 4.1.
**Baseline:** not applicable — no test harness for prose content.
**Commit:** `docs: cut restatement from the root documentation`

Steps:

- [x] Apply section 3's procedure. `README.md` is 40 KB.
      <!-- measured: wc -c README.md @ branch main -->
- [x] This repository has no `CONTRIBUTING.md` — the file the lint rule cites by that name belongs
      to a consuming project. Confirmed by `git ls-files`; nothing to preserve here.
- [x] Verify: `diff <(scripts/check-normative-inventory.sh) openspec/changes/kan-265-be-brief-in-repo-markdown/normative-baseline.txt`

### 7.2 The per-citation tail, swept corpus-wide

**Build:** green
**Files:** the files this sweep actually touches.
**Allowed-collateral:** `skills/*.md`, `skills/*/*.md`, `skills/*/**/*.md`, `rules/*.mdc`,
`rules/*.md`, `openspec/specs/**/*.md`, `commands/*.md`, `commands-claude/*.md`, `*.md`
**Tests:** none — prose-only, as 4.1.
**Regression:** not applicable — same guards as 4.1.
**Baseline:** not applicable — no test harness for prose content.
**Commit:** `refactor: cut the per-citation do-not-restate tails`

Citations across the corpus carry a trailing instruction to a future editor — "that section is the
one to read; do not restate it here", "cited rather than restated", "not re-argued here". Task 4.2
identified these as the largest remaining restatement shape and deliberately left them, to keep its
register consistent with task 4.1 rather than decide a corpus-wide question inside one task's scope.

**They became cuttable during this change.** Until task 1.1, no file stated that rule canonically, so
each tail was the only place its own reader would meet it. `rules/be-brief.mdc` now states it for the
whole repository and is always-on, present in every session — so the tails restate a rule that has a
holder, which is exactly what section 3 permits deleting.

This runs **after** tasks 4.1–7.1 and **before** task 8.1, so the ratchet is written once over the
final text.

Steps:

- [x] Cut the trailing instruction; **keep the citation itself**. The citation names which file is
      canonical, which is information; the instruction not to restate it is the rule `be-brief.mdc`
      now holds.
- [x] Leave any tail that carries something beyond the rule — a reason, a scope limit, a named
      exception. Those are content, not the reminder.
- [x] Sweep every area, so the corpus ends in one register rather than two.
- [x] Verify: `diff <(scripts/check-normative-inventory.sh) openspec/changes/kan-265-be-brief-in-repo-markdown/normative-baseline.txt`

## 8. The ratchet

### 8.1 Widen `check-contract-budget.sh` to the whole owned corpus

**Build:** green
**Files:** `scripts/check-contract-budget.sh`, `scripts/test-check-contract-budget.sh`,
`.myflow/project.md`
**Tests:** `scripts/test-check-contract-budget.sh` — existing cases retained, plus new cases for: a
spec file covered, a file under an excluded tree not covered, a directory total rejected as a
substitute for per-file rows, and an undeclared root `.md` failing.
**Regression:** reverting drops every newly covered file back out of the ratchet and lets the trim
regrow unnoticed.
**Baseline:** `before=24 after=28` cases in the harness.
<!-- measured: scripts/test-check-contract-budget.sh | grep -cE "^(ok|PASS|✓)" @ branch main -->
**Commit:** `feat(check-contract-budget): ratchet every owned Markdown file`

Steps:

- [x] Widen the scan root past `skills/` to `rules/`, `openspec/specs/`, `commands/`,
      `commands-claude/`, `.myflow/` and the repository root.
- [x] Exclude `node_modules/`, `.superpowers/`, `openspec/changes/archive/` and `docs/superpowers/`
      **structurally** — a path prefix test, not a list of files.
- [x] Keep the table hand-maintained and inside the script. Add **no** regenerate mode; see
      decision `ratchet-widened`.
- [x] Keep per-file rows. A directory total is not an acceptable substitute.
- [x] **This task runs last among the file-altering tasks**, per **A byte budget ratchets every
      owned Markdown file**'s once-as-the-last-action rule: write each row as the file's real size
      after tasks 4–7, plus 25%.
- [x] Prove each new harness case fails against the pre-widening guard before trusting it.
- [x] Verify: `scripts/test-check-contract-budget.sh && scripts/check-contract-budget.sh`

## 9. End-to-end verification

### 9.1 Prove the inventory is unchanged against the base branch, and run everything

**Build:** green
**Files:** none — verification only; no file is modified by this task.
**Allowed-collateral:** `openspec/changes/kan-265-be-brief-in-repo-markdown/verification.md`
**Tests:** none — this task runs the existing sets rather than adding one.
**Regression:** not applicable — a verification task has nothing to revert.
**Baseline:** `before=30 after=32` harness scripts under `## test`.
<!-- measured: ls scripts/test-*.sh | wc -l @ branch main -->
**Commit:** `chore(openspec): record end-to-end verification for the trim`

Steps:

- [x] `git worktree add --detach <tmp> <base-branch>`; run
      `CHECK_NORMATIVE_INVENTORY_ROOT=<tmp> scripts/check-normative-inventory.sh` there and diff it
      against a run on this branch. This is the whole-change check; tasks 4–7 each checked only
      their own group against the recorded baseline.
- [x] Remove the temporary worktree.
- [x] Run the complete `## lint` set and the complete `## test` set. Every entry green.
- [x] Record the before and after byte totals for the corpus, per area, as `measured:` claims.
- [x] Write `verification.md` with the inventory diff result, both totals, and the lint and test
      output summary.
- [x] Confirm no new lint suppression was added anywhere in the change: `git diff <base>..HEAD |
      grep -nE "noqa|shellcheck disable|eslint-disable|@Suppress"` returns nothing.
