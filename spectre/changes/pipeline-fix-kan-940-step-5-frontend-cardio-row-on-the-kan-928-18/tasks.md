# pipeline-fix-kan-940-step-5-frontend-cardio-row-on-the-kan-928-18

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** a worktree a change adds after kickoff is cut from and lands on the change's recorded
base whenever its repository's origin has it, and integrate refuses a set that does not.

**Global constraints:** guard logic stays Go in `stats/internal/guard/` behind its shell shim; the
shims and harnesses run on Bash 3.2; lint is the `<agents repo>/.flow/project.md` `## lint` lines
each task's files need.

**Review focus:**
- A gymie change with no `--base`: gymie resolves `develop`, gymie-playwright `main` — `prepare`
  must not stop (Task 3's case covers it through the fixture's no-flowBase default paths).
- `--from` pointing at a worktree with no flowBase: the worktree is cut from that repository's
  default branch, exactly as without `--from`, and no flowBase is written (Task 1).
- An explicit `<base>` with `--from`: the explicit name wins over the recorded one (Task 1).

- [ ] 1. kickoff-worktree.sh --from inherits the change's base and state record

**Build:** green
**After:** none
**Files:** `stats/internal/guard/kickoffworktree.go`, `stats/internal/guard/kickoff_worktree_test.go`, `scripts/kickoff-worktree.sh`
**Allowed-collateral:** `spectre/changes/pipeline-fix-kan-940-step-5-frontend-cardio-row-on-the-kan-928-18/**`
**Tests:** `--from: cut from the recorded base, recorded under the change's project`, `--from: base missing on that origin: exit 1, both remedies named, nothing added`
**Regression:** `--from: cut from the recorded base…` — reverted, `--from` is a usage error (exit 2); with the parse kept but the base not read, the worktree is cut from origin/main and records no flowBase; with `-C <project>` kept, `flow state add-worktree` finds no STARTED record and exits 1. `--from: base missing…` — reverted, the message names no remedy.
**Baseline:** before=9 after=11
<!-- measured: grep -c 't.Run("' stats/internal/guard/kickoff_worktree_test.go @ 2d9dc724 -->
**Commit:** fix(scripts): kickoff-worktree --from cuts an added worktree from the change's base

  - [ ] **Step 1:** In `kickoff_worktree_test.go`'s `TestKickoffWorktree`, add subtest `--from: cut from the recorded base, recorded under the change's project`: two fixtures `canon := kwNew(t, "", true)` and `peer := kwNew(t, "", true)`; push an empty commit to `refs/heads/feature/x` on each origin (the `named base` subtest's clone-commit-push pattern); `canon.started(t)`; `canon.run(t, canon.project, kwName, "feature/x")` exits 0; then `canon.run(t, "--from", canon.project+"-worktrees/"+kwName, peer.project, kwName)` — run under `canon`'s env, so the state dir is the canonical one — asserts rc 0, the peer worktree `peer.project+"-worktrees/"+kwName` at peer's `origin/feature/x` tip, `config --get branch.spectre/<kwName>.flowBase` = `feature/x` there, and `canon.flow(t, "", "state", "get", "-C", canon.project, kwName)` containing the peer worktree's path.
  - [ ] **Step 2:** Add subtest `--from: base missing on that origin: exit 1, both remedies named, nothing added`: same, but `feature/x` pushed to canon's origin only; the `--from` run exits 1, stderr contains `origin/feature/x does not exist`, `push feature/x` and `default branch`, and the peer worktree directory does not exist.
  - [ ] **Step 3:** Run `cd stats && go test ./internal/guard -run 'TestKickoffWorktree'` — the two new subtests fail (exit 2, usage).
  - [ ] **Step 4:** In `kickoffworktree.go`, accept a leading `--from <dir>` pair before `<project> <name> [<base>]` (an empty or missing `<dir>` is usage, exit 2); usage becomes `usage: kickoff-worktree.sh [--from <worktree>] <project> <name> [<base>]` — update the `usage: exit 2` subtest's expected string. With `--from` and no `<base>`, the base is `git -C <dir> config --get branch.spectre/<name>.flowBase` (exit 1 = none, base stays empty), then validated as a named base is. `flow state add-worktree` runs with `-C <dir>` when `--from` is given, `-C <project>` otherwise. The missing-base message becomes `origin/<base> does not exist in <project> — stop and ask the operator: push <base> to that origin, or re-run naming that repository's default branch as <base>`.
  - [ ] **Step 5:** Update `scripts/kickoff-worktree.sh`'s header: the usage line, a paragraph on `--from` (base default, state record, the explicit `<base>` winning), and the exit-1 list naming the missing base.
  - [ ] **Step 6:** `cd stats && gofmt -w . && gofmt -l . && go vet ./... && go test ./internal/guard -run 'TestKickoffWorktree'` — all pass; then `scripts/check-guard-symlinks.sh` and `scripts/check-references.sh`.

- [ ] 2. Every added worktree runs the kickoff recipe with --from

**Build:** green
**After:** Task 1
**Files:** `skills/flow/cross-repo-worktrees.md`, `skills/flow/implement.md`, `skills/flow/SKILL.md`
**Allowed-collateral:** `spectre/changes/pipeline-fix-kan-940-step-5-frontend-cardio-row-on-the-kan-928-18/**`
**Tests:** **none** — prose contract; guarded by check-references and check-markdown-integrity.
**Regression:** reverted, a stage that needs a worktree after `flow.isolate-workspace` has no recipe and types its own `git worktree add` off the default branch — kan-940's path.
**Commit:** fix(flow): every worktree a change adds runs kickoff-worktree --from

**Decision:** kickoff-from

**Decision:** missing-base-stops

**Decision:** cardio-memory

  - [ ] **Step 1:** In `skills/flow/cross-repo-worktrees.md`, replace the `## apps` recipe — from "Every other entry gets the kickoff recipe in its own repository" through "`git -C <that worktree> push -u origin spectre/<name>` per **Branch backup** (`skills/flow-contracts/git-boundaries.md`)." — with one call, `<agents repo>/scripts/kickoff-worktree.sh --from <this change's worktree> <that repo> <name>`, its header canonical for the steps (location guard, resume or cut from `origin/<base>` when this change's worktree records a flowBase, else `origin/<default-branch>`, the flowBase recorded in that repository, the merge base persisted into this change's state record, `## worktree setup`, push). Keep "No `spectre link` runs for these worktrees …" and the hard-failure sentence, adding: exit 1 naming a missing `origin/<base>` stops the run and asks the operator the two remedies the message names. The paragraph opening "**Every worktree this stage creates**" keeps governing linked peers; for `## apps` worktrees the script now runs the setup.
  - [ ] **Step 2:** In `skills/flow/implement.md` **2. Isolate the workspace**, change "— never improvised mid-run by a later stage in need of a commit destination." so a later stage that finds it needs a worktree in a repository the set does not hold — a fix round's task, `flow.visual-verify`'s commit destination — runs that entry's recipe of `skills/flow/cross-repo-worktrees.md` (the `kickoff-worktree.sh --from` call, the regression checkout's toolchain, the merge-order extension), never a bare `git worktree add`.
  - [ ] **Step 3:** In `skills/flow/SKILL.md`'s `--base` paragraph, after `kickoff-worktree.sh <project> <name> <base>`, name `kickoff-worktree.sh --from` for every worktree added after kickoff.
  - [ ] **Step 4:** Outside the repository — never in **Files:**, never committed: in `/Users/tweety53/.claude/projects/-Users-tweety53-Projects-gymie/memory/cardio-feature-branch.md`, state that `feature/cardio` exists in gymie, gymie-frontend and gymie-playwright (gymie-playwright's `origin/feature/cardio` since 2026-10-14), so every cardio worktree, playwright's included, is cut from and lands on it — `kickoff-worktree.sh --from` inherits it; update the description frontmatter and the `MEMORY.md` index line "(gymie + frontend)" to match.
  - [ ] **Step 5:** `scripts/check-references.sh`, `scripts/check-markdown-integrity.py`, `scripts/check-installed-citations.sh`, `scripts/check-verbatim-moves.sh`, `scripts/check-normative-inventory.sh` — all pass.

- [ ] 3. integrate prepare stops on a worktree off the change's recorded base

**Build:** green
**After:** none
**Files:** `scripts/integrate-change.sh`, `scripts/test-integrate-change.sh`, `skills/flow/integrate.md`
**Allowed-collateral:** `spectre/changes/pipeline-fix-kan-940-step-5-frontend-cardio-row-on-the-kan-928-18/**`
**Tests:** `base_mismatch`
**Regression:** `base_mismatch` — reverted, `prepare` with `a` recording flowBase `feature/x` and `b` resolving `main` while `b`'s origin has `feature/x` runs past the gate, and `b` is landed on `main`.
**Baseline:** before=32 after=33
<!-- measured: grep -c '^# --- ' scripts/test-integrate-change.sh @ 2d9dc724 -->
**Commit:** fix(scripts): integrate prepare stops a worktree off the change's recorded base

**Decision:** base-mismatch-guard

  - [ ] **Step 1:** Add case `# --- base_mismatch ---` to `scripts/test-integrate-change.sh` (and its line in the header's Cases list): `fixture`; push `main` to `refs/heads/feature/x` on both origins (`g -C "$WORK/$r" push -q origin main:refs/heads/feature/x`); `g -C "$WORK/a-worktrees/demo" config branch.spectre/demo.flowBase feature/x`; `sut prepare "$WORK/a" demo`; checks: exit 1; last line matches `^STOP: base-mismatch — ` and names `$WORK/b-worktrees/demo` and `feature/x`; `! grep -q flow.preflight "$WORK/stages.log"`. Then `g -C "$WORK/b-worktrees/demo" config branch.spectre/demo.flowBase feature/x`, `sut prepare` again: its output carries no `STOP: base-mismatch`. Then a fresh `fixture` with `feature/x` pushed to `a`'s origin only and `a`'s flowBase set: `prepare` output carries no `STOP: base-mismatch`.
  - [ ] **Step 2:** `bash scripts/test-integrate-change.sh 2>&1 | grep -E '^(FAIL|ok - base_mismatch|[0-9]+ passed)'` with Bash `timeout: 600000` — the base_mismatch exit/stop checks fail.
  - [ ] **Step 3:** In `cmd_prepare`, after the `if [ "$run1" -eq 0 ]; then … fi` block and before `mark_begin flow.preflight`: for each `i` whose `git -C "${WT[i]}" config --get "branch.spectre/$NAME.flowBase"` prints `<b>`, for each `j` with `BASES[j]` ≠ `<b>` and `git -C "${WT[j]}" rev-parse -q --verify "refs/remotes/origin/<b>"` succeeding: `stop base-mismatch "${WT[j]} resolves base ${BASES[j]}, but ${WT[i]} records $b and origin/$b exists in ${WT[j]}'s repository — ask the operator; never land it on ${BASES[j]}"`. New locals only (`rb` is taken). Add the stop to the header's `prepare` paragraph.
  - [ ] **Step 4:** In `skills/flow/integrate.md`'s stop list, after `STOP: preflight` or `STOP: base`, add `STOP: base-mismatch` → report the line and ask the operator; the answer that moves the worktree is `git -C <wt> config branch.spectre/<name>.flowBase <b>`, `git -C <wt> rebase --onto origin/<b> <its recorded merge base>`, `flow state add-worktree <name> <wt> $(git -C <wt> merge-base HEAD origin/<b>)`, then re-run `prepare`.
  - [ ] **Step 5:** `bash scripts/test-integrate-change.sh` with Bash `timeout: 600000` — `0 failed`; `scripts/check-references.sh`, `scripts/check-markdown-integrity.py`, `scripts/check-guard-symlinks.sh`.
