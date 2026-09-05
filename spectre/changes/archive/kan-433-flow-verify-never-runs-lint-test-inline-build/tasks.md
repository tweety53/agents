# kan-433-flow-verify-never-runs-lint-test-inline-build

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

**Spec:** `design.md` beside this file. Four tasks, every one `Build: green`, no shared files
between them. Tasks 1–3 are documentation; the mechanical checks are the shipped guards named in
each task's last step. Task 4 is Go and carries the one test this plan adds.

- [x] 1. Add the `## worktree setup` key to the project-configuration contract and declare it for this repository

**Build:** green
**Files:** `skills/flow-contracts/project-configuration.md`, `.flow/project.md`
**Tests:** none — documentation; the guards in step 3 are the check
**Regression:** none — no test is added; `scripts/check-references.sh`, `scripts/check-workspace-isolation.sh` and `scripts/check-contract-budget.sh` pass before and after
**Baseline:** before=0 after=0
<!-- measured: no test suite covers skills/ or .flow/ prose; the count is the number of tests this task touches @ branch main -->
<!-- predicted: unchanged after this task, on branch spectre/kan-433-flow-verify-never-runs-lint-test-inline-build -->
**Commit:** `docs(flow-contracts): add the worktree setup key`

  - [x] **Step 1: The contract row.** In `skills/flow-contracts/project-configuration.md`'s key
    table (the one opening with the `## apps` row), insert this row directly after the `## test`
    row (`| \`## test\` | The command(s) that run the project's tests. |`):

```markdown verified:column layout copied from the key table at skills/flow-contracts/project-configuration.md:24-37 on main
| `## worktree setup` | Optional. A fenced command block run once per worktree, from the worktree root, immediately after **2. Isolate the workspace** (`skills/flow/implement.md`) creates it and before anything else touches the tree — the place for a build that a gitignored, embedded artifact needs before the project's first `go test`/`go build` can succeed. Absent means nothing runs. Same shape as `## lint` and `## test`: one command per line inside the fence, read literally and run in order; resolved through `project-get.sh <worktree> "worktree setup"`, whose exit 1 is the absent case. |
```

  - [x] **Step 2: This repository's section.** In `.flow/project.md`, insert a new level-2
    section directly before `## lint` (after the `check-installed-citations.sh` runtime paragraph
    that closes `## test`):

````markdown verified:authored in-tree for this change; the gitignore line and embed pattern are .gitignore:23 and stats/internal/web/embed.go:29 on main
## worktree setup

```bash
cd stats && make web-build
```

`stats/internal/web/dist/` is gitignored and `stats/internal/web/embed.go`'s `//go:embed all:dist`
refuses to compile without it, so a fresh worktree's first `go test ./...` or `go build ./...`
fails until the SPA is built once. kan-389's verifier hit exactly that, and its conductor then ran
the whole `## test` list itself to get past it — a fresh-worktree fact, not a branch defect, and
the reason this key exists. `make web-build` is `npm ci && npm run build` in `stats/web`
(`stats/Makefile`), the same prerequisite `make test` and `make build` already carry.
````

  - [x] **Step 3: Verify.** From the repository root run `scripts/check-references.sh`,
    `scripts/check-workspace-isolation.sh`, `scripts/check-markdown-integrity.py`,
    `scripts/check-contract-budget.sh` and `scripts/project-get.sh . "worktree setup"` — the
    first four exit 0; the last prints the fenced block's body and exits 0. Both files are under
    their budgets (`project-configuration.md` 45976 of 48175, `.flow/project.md` 24597 of
    26450); if `check-contract-budget.sh` nevertheless trips, raise that row's budget by the
    overage and say so in the commit body. Commit.
<!-- measured: wc -c on both files and grep of the budgets() table in scripts/check-contract-budget.sh @ branch main -->

- [x] 2. State worktree creation inline in `implement.md` §2 and drop `superpowers:using-git-worktrees`

**Build:** green
**Files:** `skills/flow/implement.md`, `skills/README.md`, `skills/flow-contracts/finish-contract-run1.md`
**Tests:** none — documentation; the guards in step 5 are the check
**Regression:** none — no test is added; `scripts/check-references.sh`, `scripts/check-installed-citations.sh` and `scripts/check-dispatch-paragraphs.sh` pass before and after
**Baseline:** before=0 after=0
<!-- measured: no test suite covers skills/ prose; the count is the number of tests this task touches @ branch main -->
<!-- predicted: unchanged after this task, on branch spectre/kan-433-flow-verify-never-runs-lint-test-inline-build -->
**Commit:** `docs(flow): state worktree creation inline and drop the worktree skill`

  - [x] **Step 1: The step table.** In `skills/flow/implement.md`'s table at the top, replace
    row **2** (`| **2** | **superpowers:using-git-worktrees** | Before the first code change, on a first run |`)
    with:

```markdown verified:column layout copied from skills/flow/implement.md:7-15 on main
| **2** | worktree creation and `## worktree setup`, stated in **2. Isolate the workspace** below | Before the first code change, on a first run |
```

  - [x] **Step 2: The section.** In **2. Isolate the workspace (first run only)**, replace the
    paragraph opening "Invoke **superpowers:using-git-worktrees**. Branch `spectre/<name>`." —
    through its last sentence "…the state file's `worktrees` map is written only at the end of
    `skills/flow/verify-and-handoff.md`." — with:

```markdown verified:authored in-tree for this change; the worktree location is finish-contract-run1.md:330-331 on main and project-get.sh's exit codes are its own header
Create the worktree yourself — no skill is invoked here, and no baseline test suite runs: `flow.verify`
runs the project's `## test` list at the end of this same run, and the base-branch guards cover the
base. Never implement on the default branch without explicit consent.

1. `git check-ignore -q .worktrees` from the project root. Where it exits non-zero, append
   `.worktrees/` to `<project>/.gitignore` and commit that on the default branch first — an
   unignored worktree directory commits the whole tree into the repository.
2. `git worktree add <project>/.worktrees/<name> -b spectre/<name>` from the default branch's HEAD.
3. `project-get.sh <worktree> "worktree setup"`. Exit 0: run every printed line from the worktree
   root, in order, in the foreground. Exit 1: the project declares no `## worktree setup`; say so
   and continue. Exit 2: stop the run, relaying the script's own line. **A command's non-zero exit
   ends your turn with `## Question`** naming the command and its output — a worktree that cannot
   be set up fails `flow.verify` later anyway, and the operator should see it here. The key is
   canonical in **Project configuration** (`skills/flow-contracts/project-configuration.md`).

Record each worktree's merge base and absolute path in this run's own working notes as soon as the
worktree exists — the state file's `worktrees` map is written only at the end of
`skills/flow/verify-and-handoff.md`.
```

    Everything after that paragraph in the section — the change-directory copy-and-remove, the
    fix-run resume rule, `worktree-resolution.md`, `spectre link`, `flow workspace-id` and the stage
    `end` mark — is unchanged. Afterwards `grep -n "using-git-worktrees" skills/flow/implement.md`
    prints nothing.
  - [x] **Step 3: `skills/README.md`.** In the Superpowers Basic Workflow map, replace the row
    `| **2** | using-git-worktrees | \`/flow\` (implementation) |` with:

```markdown verified:column layout copied from skills/README.md:29-37 on main
| **2** | worktree creation, stated in `skills/flow/implement.md` (no skill) | `/flow` (implementation) |
```

  - [x] **Step 4: `finish-contract-run1.md`.** At `skills/flow-contracts/finish-contract-run1.md:331`,
    replace "(git-ignored, per `superpowers:using-git-worktrees`)" with
    "(git-ignored, per **2. Isolate the workspace** in `skills/flow/implement.md`)". Afterwards
    `grep -rn "using-git-worktrees" skills scripts rules commands commands-claude CLAUDE.md`
    prints nothing.
  - [x] **Step 5: Verify.** From the repository root run `scripts/check-references.sh`,
    `scripts/check-installed-citations.sh`, `scripts/check-dispatch-paragraphs.sh`,
    `scripts/check-vocabulary.sh`, `scripts/check-markdown-integrity.py` and
    `scripts/check-contract-budget.sh` — each exits 0. Commit.

- [x] 3. Bind the conductor in `verify-and-handoff.md` and widen the stale-result rule in `review-panel.md`

**Build:** green
**Files:** `skills/flow/verify-and-handoff.md`, `skills/flow/review-panel.md`
**Tests:** none — documentation; the guards in step 5 are the check
**Regression:** none — no test is added; `scripts/check-references.sh`, `scripts/check-stage-mark-calls.sh` and `scripts/check-normative-inventory.sh` pass before and after
**Baseline:** before=0 after=0
<!-- measured: no test suite covers skills/ prose; the count is the number of tests this task touches @ branch main -->
<!-- predicted: unchanged after this task, on branch spectre/kan-433-flow-verify-never-runs-lint-test-inline-build -->
**Commit:** `docs(flow): keep the conductor out of lint, test and source after the panel`

  - [x] **Step 0: Capture the normative inventory.** From the repository root run
    `scripts/check-normative-inventory.sh > /tmp/kan-433-inventory-before.txt` before any edit
    below; step 5 diffs against it.
  - [x] **Step 1: The conductor's boundary.** In `skills/flow/verify-and-handoff.md` **Verify**,
    directly after the paragraph opening "**This step does not call the project's `create`
    command.**", insert:

```markdown verified:authored in-tree for this change; the cited section names exist at skills/flow/review-panel.md:528 and skills/flow/implement.md:149 on main
**After the panel closes, the conductor edits no source and runs none of the `## lint` or `## test`
commands itself.** Its work in this stage is `prepare-workspace.sh`, the verifier dispatch(es) below
and the ledger render — nothing else. Any source change from here on makes every slot's result stale
(**Panel re-runs**, `skills/flow/review-panel.md`), and the only path that changes source is a fix
run the operator starts. A failing check is never "just re-run to see" by the conductor: the
verifier is re-dispatched with it, per **The verifier dispatch** below.
```

  - [x] **Step 2: The re-dispatch.** In **The verifier dispatch**, directly after the
    **Recording.** paragraph (the one ending "…the pair's semantics are section 4 of
    `skills/flow/implement.md`, cited here, not restated."), insert:

```markdown verified:authored in-tree for this change; -key verify is the existing key at skills/flow/verify-and-handoff.md:87-90 on main
**A `## Report` carrying any non-zero exit is re-dispatched once.** The second dispatch is recorded
under `-key verify-2` here and `visual-verify-2` in **Visual verification** below — the same
`-<worktree basename>` suffix rule — with a prompt identical to the first plus the first `## Report`
verbatim under a `## Previous attempt` heading, so the verifier can tell an environmental failure
(a build the worktree lacked, a flaky harness) from a defect in the branch. **The second report is
final.** Another non-zero exit ends your turn with `## Question` naming the failing command and its
output, verbatim; the operator resolves it through a fix run. Never run the failing command yourself
to check it, and never dispatch a third verifier. The ledger render and this stage's `end` mark
follow whichever report was last.
```

  - [x] **Step 3: Guardrails.** In `skills/flow/verify-and-handoff.md` **Guardrails**, append
    after the last bullet ("**Never** mark a task's checkbox before that task's review passes."):

```markdown verified:authored in-tree for this change
- **Never** edit source, and **never** run a `## lint` or `## test` command yourself, after the
  panel closes — the verifier runs them, and a fix run changes source.
```

  - [x] **Step 4: The stale-result rule.** In `skills/flow/review-panel.md` **Panel re-runs**,
    the sentence at line 596–598 reading "Handoff still requires **zero open findings at any
    severity** from every agent that has run, and no stale result — where **a slot's clean result
    is stale when the rule above required that slot to re-run and it has not**." becomes:

```markdown verified:the sentence being replaced is skills/flow/review-panel.md:596-598 on main
Handoff still requires **zero open findings at any severity** from every agent that has run, and
no stale result — where **a slot's clean result is stale when the rule above required that slot to
re-run and it has not, or when any commit or working-tree change to source landed after that slot's
last read, from any stage — `flow.verify` included**. An unrecorded edit after the panel closes is
stale by definition, not only one a fix round produced.
```

    The sentence that follows ("A non-Minor fix Bugbot or Mutation did not raise leaves that
    slot's result current: …") is unchanged.
  - [x] **Step 5: Verify.** From the repository root run `scripts/check-references.sh`,
    `scripts/check-stage-mark-calls.sh`, `scripts/check-dispatch-paragraphs.sh`,
    `scripts/check-vocabulary.sh`, `scripts/check-markdown-integrity.py`,
    `scripts/check-contract-budget.sh` — each exits 0 — and
    `scripts/check-normative-inventory.sh | diff /tmp/kan-433-inventory-before.txt -` — empty
    output, since this task adds no `MUST`/`SHALL` sentence. `verify-and-handoff.md` is 28214 of
    its 33199 budget and `review-panel.md` 46910 of 58732; if the budget guard nevertheless trips,
    raise that row by the overage and say so in the commit body. Commit.
<!-- measured: wc -c on both files and grep of the budgets() table in scripts/check-contract-budget.sh @ branch main -->

- [x] 4. List `verifier` in `flow record dispatch`'s usage text and pin it to `recordRoles`

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`
**Tests:** `TestRecordUsageNamesEveryRole`
**Regression:** `TestRecordUsageNamesEveryRole` — reverting this task's commit restores the usage line that omits `verifier`, and the test fails naming it
**Baseline:** before=136 after=137
<!-- measured: cd stats && go test ./cmd/flow -count=1 -v 2>&1 | grep -c '^--- PASS' @ branch main -->
<!-- predicted: the same command after this task, on branch spectre/kan-433-flow-verify-never-runs-lint-test-inline-build -->
**Commit:** `fix(stats): list verifier in record dispatch usage`

  - [x] **Step 1: Write the failing test.** Append to `stats/cmd/flow/record_test.go`
    (package `main`, `strings` and `testing` already imported):

```go verified:recordRoles and recordUsage are package-level in stats/cmd/flow/record.go:47 and :88 on main
// TestRecordUsageNamesEveryRole pins the usage text's "-role is one of:"
// line to recordRoles, so a role added to the slice cannot stay missing
// from the help a caller reads -- which is how "verifier" was accepted by
// the CLI while the usage still listed six roles.
func TestRecordUsageNamesEveryRole(t *testing.T) {
	for _, role := range recordRoles {
		if !strings.Contains(recordUsage, " "+role+",") && !strings.Contains(recordUsage, " "+role+".") {
			t.Errorf("recordUsage does not list role %q on its -role line", role)
		}
	}
}
```

  - [x] **Step 2: Run it red.** `cd stats && go test ./cmd/flow -run TestRecordUsageNamesEveryRole -count=1`
    — FAIL with `recordUsage does not list role "verifier" on its -role line`.
  - [x] **Step 3: Fix the usage line.** In `stats/cmd/flow/record.go`, the line
    `-role is one of: implementer, reviewer, panel-fix, red-partner, planner, conductor.` becomes:

```text verified:the line being replaced is stats/cmd/flow/record.go:158 on main
-role is one of: implementer, reviewer, panel-fix, red-partner, planner, conductor, verifier.
```

  - [x] **Step 4: Run it green, then the package.** `cd stats && go test ./cmd/flow -run TestRecordUsageNamesEveryRole -count=1`
    — PASS. Then `cd stats && go test ./cmd/flow -count=1 -v 2>&1 | grep -c '^--- PASS'` prints
    137.
  - [x] **Step 5: Lint and commit.** `cd stats && gofmt -w . && go vet ./... && gofmt -l .` —
    `gofmt -l` prints nothing, `go vet` exits 0. Commit.
