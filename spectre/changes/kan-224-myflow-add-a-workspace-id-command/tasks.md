# kan-224-myflow-add-a-workspace-id-command

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.

> **Relocation:** no

- [x] 1. Add `deriveWorkspaceID` and its tests
  - [x] **Step 1: write `deriveWorkspaceID(name string) string` in `stats/cmd/flow/workspaceid.go`**

    Port the four-step derivation from `skills/flow-contracts/workspace-isolation.md`'s **The
    workspace id** exactly — digest, normalised name, segment-bounded 12-char prefix, `prefix-digest`
    join — operating on `name`'s UTF-8 bytes (`[]byte(name)`), matching the byte-wise semantics
    `scripts/prepare-workspace.sh:143-152`'s `LC_ALL=C` calls already implement in Bash: `A`-`Z`
    lowered by 32, `a`-`z`/`0`-`9` kept, every other byte becomes `-`; the prefix is the longest
    leading run of whole `-`-separated segments totalling at most 12 characters, or the first
    segment's first 12 characters if that segment alone exceeds 12; trailing `-` trimmed from the
    prefix only.

    verified:the digest step is `crypto/sha256` over `[]byte(name)`, hex-encoded, first 4 chars —
    the same `shasum -a 256 | cut -c1-4` `prepare-workspace.sh:149` already runs
  - [x] **Step 2: add `stats/cmd/flow/workspaceid_test.go` covering the contract's own worked examples**

    Table-test the four examples `workspace-isolation.md`'s verified bash block names verbatim:
    `kan-15-parallel-myflow-do-task-lanes` → `kan-15-55a6`, `Demo_X` → `demo-x-5197`,
    `KAN-99-Fix.Thing` → `kan-99-fix-feef`, `İstanbul-test` → `--stanbul-6b8a` (multi-byte UTF-8
    character, each byte becomes its own `-`). Add one case for a name whose first segment alone
    exceeds 12 characters, to cover the second prefix-truncation branch.

**Files:** `stats/cmd/flow/workspaceid.go`, `stats/cmd/flow/workspaceid_test.go`
**Tests:** `TestDeriveWorkspaceID`
**Regression:** reverting this task's commit removes `deriveWorkspaceID`, and `workspaceid_test.go` fails to compile since task 2's `runWorkspaceID` also references it
**Baseline:** before=0 after=6
**Build:** green
**Commit:** `feat(flow-cli): derive the workspace id from a change name`

- [x] 2. Wire `flow workspace-id <name>` into the CLI
  - [x] **Step 1: add `runWorkspaceID` in `stats/cmd/flow/workspaceid.go`**

    Mirror `stats/cmd/flow/tasks.go`'s `runTasksTick` shape (a `flag.NewFlagSet`, though this
    command takes no flags — only `flag.ErrHelp` handling and the `NArg() != 1` check apply):
    require exactly one positional argument, the change name; print `deriveWorkspaceID(name)` to
    stdout with a trailing newline; exit 0. No filesystem or network access — this command reads
    only its argument.
  - [x] **Step 2: register `workspace-id` in `main.go`**

    Add `workspace-id <name>  print a change's workspace id, derived from its name` to the `usage`
    string (alongside the existing `tasks tick` line) and a `case "workspace-id":` arm in `run`'s
    switch dispatching to `runWorkspaceID(args[1:], stdout, stderr)`.
  - [x] **Step 3: add cmd-level test coverage in `stats/cmd/flow/workspaceid_test.go`**

    Add cmd-level cases mirroring `tasks_test.go`'s `TestTasksTickWritesFile`/
    `TestTasksTickTaskNotFoundExitsNonZero` shape: `run` with `["workspace-id",
    "kan-15-parallel-myflow-do-task-lanes"]` prints `kan-15-55a6\n` and exits 0; `run` with
    `["workspace-id"]` (no argument) exits 2 and prints a usage line to stderr; `run` with
    `["workspace-id", "a", "b"]` (too many arguments) exits 2.

**Files:** `stats/cmd/flow/workspaceid.go`, `stats/cmd/flow/main.go`, `stats/cmd/flow/workspaceid_test.go`
**Tests:** `TestWorkspaceIDCommandPrintsID`, `TestWorkspaceIDCommandNoArgExitsNonZero`, `TestWorkspaceIDCommandTooManyArgsExitsNonZero`
**Regression:** reverting this task's commit removes the `workspace-id` case from `main.go`'s switch, so `flow workspace-id <name>` falls into the `unknown command` branch (exit 2) and the three tests above fail
**Baseline:** before=6 after=9
**Build:** green
**Commit:** `feat(flow-cli): add the workspace-id command`

- [x] 3. Point finish-contract run 2 at the new command
  - [x] **Step 1: update `skills/flow-contracts/finish-contract-run2.md` step 5**

    In the paragraph beginning "The removal runs the project's `remove` command...", replace "**Run
    2 is not handed that id and does not need to be**: it is derived from the change name and from
    nothing else, deterministically and without ever being recorded, per **The workspace id**
    (`skills/flow-contracts/workspace-isolation.md`) — so run 2 re-derives it and arrives at the id
    `/myflow-do` used, in a session that shared nothing with it." with text instructing run 2 to
    call `flow workspace-id <name>` for the id rather than re-deriving it from the contract by hand,
    keeping the "not handed the id, derives it independently, and always arrives at the id
    `/myflow-do` used" property — only the mechanism (a CLI call instead of a hand re-derivation)
    changes. `workspace-isolation.md` itself is not edited by this task — it stays the canonical
    spec, cited rather than restated, exactly as the ticket asks.
  - [x] **Step 2: grep the repo for other prose instructing a hand re-derivation of the workspace id**

    `grep -rn "re-derives it\|hand.derive\|hand.*workspace id" skills/ scripts/` and confirm no
    other contract or skill file (besides the one edited in Step 1) tells an agent to re-derive the
    id from `workspace-isolation.md`'s bash block by hand. `scripts/prepare-workspace.sh`'s own
    inline derivation is out of scope (design.md's `no-touch-prepare-workspace`) and is expected to
    still appear in this grep — do not edit it.

**Files:** `skills/flow-contracts/finish-contract-run2.md`
**Tests:** none
**Regression:** none — a documentation-only task
**Baseline:** before=9 after=9
**Build:** green
**Commit:** `docs(flow-contracts): point run 2 at the workspace-id command`
