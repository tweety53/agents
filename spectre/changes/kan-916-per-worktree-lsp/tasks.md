# kan-916-per-worktree-lsp

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Tasks appended:** 1

Tasks 1–2 port the location guard and move kickoff to the sibling layout, task 3 moves the hook and
the board mod, tasks 4–6 build the `worktree-lsp` wrapper and its plugin, tasks 7–8 update the
skills, task 9 adds the migration guard; `flow.verify`'s live check runs the migration. `design.md` is canonical for every decision; each
task cites its entry by ID.

**Global constraints**

- `<main>` is a main checkout's physical path; the sibling root is `$(dirname <main>)/$(basename <main>)-worktrees`.
- Go: `cd stats && gofmt -l . && go vet ./...` clean after every Go task; tests run with `-race -count=1`.
- No task starts, stops or touches `flowd` on `127.0.0.1:4173` or `flow-postgres` (`CLAUDE.md`).
- No task edits `~/.claude/settings.json`; no task runs `setup.sh global` against the real `$HOME`.

- [x] 1. Port `check-worktree-location` to `flow-guard` at byte parity
  - [x] **Step 1: Read `scripts/check-worktree-location.sh` and `scripts/test-check-worktree-location.sh` at the merge base — the script is the specification: arguments, every stdout line byte for byte, the stdout/stderr split, exit codes 0/1/2.**
  - [x] **Step 2: Write `stats/internal/guard/check_worktree_location_test.go` with `TestCheckWorktreeLocation`, one subtest per case of the bash harness (named after the harness's `pass` labels), building real repositories with `git worktree add` in `t.TempDir()`. Run it red: `cd stats && go test ./internal/guard/ -run TestCheckWorktreeLocation -count=1` fails (guard unregistered).**
  - [x] **Step 3: Port the script to `stats/internal/guard/worktreelocation.go`, registered as `Registry["check-worktree-location"]` from its `init()`; the script body's comments move beside the code they explain. Run the test green.**
  - [x] **Step 4: Replace the script body with the `flow_guard_exec check-worktree-location 2 "check-worktree-location:" "$@"` shim (template: `scripts/check-main-checkout-drift.sh`'s tail), header comment kept; `git rm scripts/test-check-worktree-location.sh`.**
  - [x] **Step 5: Verify: `cd stats && go test ./internal/guard/ -run TestCheckWorktreeLocation -race -count=1 && gofmt -l . && go vet ./internal/guard/ && cd .. && scripts/check-worktree-location.sh "$(git worktree list --porcelain | awk '/^worktree /{print substr($0,10); exit}')"; scripts/check-guard-symlinks.sh`.**
**Build:** green
**Files:** `stats/internal/guard/worktreelocation.go`, `stats/internal/guard/check_worktree_location_test.go`, `scripts/check-worktree-location.sh`, `scripts/test-check-worktree-location.sh`, `stats/internal/guard/finishpreflight.go`
**Tests:** `TestCheckWorktreeLocation`
**Regression:** reverting the commit restores the bash body and deletes `TestCheckWorktreeLocation`; with only the shim reverted the test still runs against the Go port, and with the port deleted `TestCheckWorktreeLocation` fails (unregistered guard).
**Baseline:** before=156 after=157
<!-- measured: grep -h '^func Test' stats/internal/guard/*_test.go | wc -l @ branch spectre/kan-916-per-worktree-lsp -->
**Commit:** refactor(guard): port check-worktree-location to flow-guard
**After:** none

Correction (2026-10-07): `finishpreflight.go`'s comment said the location guard stays bash; the port made it false, so the comment was reworded in this task's commit. The old-vs-new byte diff over every harness case plus symlink, missing path, regular file, non-repo, empty, zero and two arguments came back empty (2377 bytes) once the port refused an empty argument as bash's `cd ""` does.

**Decision:** port-location-guard-first

- [x] 2. Location guard and kickoff on the sibling layout
  - [x] **Step 1: Add `TestCheckWorktreeLocationSiblingLayout` to `stats/internal/guard/check_worktree_location_test.go`: a worktree at `<parent>/<repo>-worktrees/x` → `LOCATION-OK`; one at `<repo>/.worktrees/x` → `STRAY`; one at `<parent>/<repo>-worktrees-x/y` (prefix lookalike) → `STRAY`; the main checkout itself never flagged. Add `TestKickoffWorktreeSiblingLayout` to `stats/internal/guard/kickoff_worktree_test.go`: kickoff creates `<parent>/<repo>-worktrees/<name>`, writes nothing to `.git/info/exclude`, and the `worktree:` line names that path. Run both red.**
  - [x] **Step 2: `worktreelocation.go` — the accepted root becomes `filepath.Dir(root)+"/"+filepath.Base(root)+"-worktrees"`, compared physically (`EvalSymlinks`), "at or under" meaning equal or prefix-plus-separator; rewrite the header comments and `STRAY` wording that name `.worktrees`. Update `scripts/check-worktree-location.sh`'s header comment to the new rule.**
  - [x] **Step 3: `kickoffworktree.go` — drop step 2 (the `.worktrees` info/exclude append) and create the worktree at the sibling root (`os.MkdirAll` the root first); renumber the step comments; update `scripts/kickoff-worktree.sh`'s header to match.**
  - [x] **Step 4: Move every Go test fixture that builds a `.worktrees/` worktree a guard now execs the location check against to the sibling layout: `stats/internal/guard/check_finish_preflight_test.go`, `stats/internal/guard/check_planning_commit_location_test.go`, `stats/internal/guard/reshape_branch_test.go`, `stats/internal/guard/resolve_visual_screenshots_test.go`, `stats/internal/guard/kickoff_worktree_test.go`. Verify: `cd stats && go test ./internal/guard/ -race -count=1 && gofmt -l . && go vet ./internal/guard/`.**
**Build:** green
**Files:** `stats/internal/guard/worktreelocation.go`, `stats/internal/guard/check_worktree_location_test.go`, `stats/internal/guard/kickoffworktree.go`, `stats/internal/guard/kickoff_worktree_test.go`, `stats/internal/guard/check_finish_preflight_test.go`, `scripts/check-worktree-location.sh`, `scripts/kickoff-worktree.sh`
**Tests:** `TestCheckWorktreeLocationSiblingLayout`, `TestKickoffWorktreeSiblingLayout`, `TestKickoffWorktreeSymlinkedProject`
**Regression:** reverting the commit puts kickoff back under `.worktrees/` and the guard back on its old root: `TestCheckWorktreeLocationSiblingLayout` fails (sibling reported `STRAY`, `.worktrees` accepted) and `TestKickoffWorktreeSiblingLayout` fails (worktree path and info/exclude append).
**Baseline:** before=157 after=160
<!-- measured: grep -h '^func Test' stats/internal/guard/*_test.go | wc -l @ branch spectre/kan-916-per-worktree-lsp (157 once task 1 lands) -->
**Commit:** feat(guard): enforce the sibling <repo>-worktrees layout
**After:** Task 1

Correction (2026-10-07): the plan moved five fixtures; three (`check_planning_commit_location_test.go`, `reshape_branch_test.go`, `resolve_visual_screenshots_test.go`) exec no location check and pass unchanged, and moving the last would empty its `.worktrees` prune case, so they stay. Kickoff derives the sibling root from the physical project path (`EvalSymlinks`) — a symlinked project otherwise gets a worktree the location guard reports `STRAY` — pinned by `TestKickoffWorktreeSymlinkedProject`; no `os.MkdirAll`, since `git worktree add` creates leading directories. Kickoff's steps renumber 1–5 → 1–4.

**Decision:** sibling-worktrees-layout

- [x] 3. Main-checkout hook, board mod and `.gitignore` on the sibling layout
  - [x] **Step 1: `scripts/test-protect-main-checkout.sh` — add a case labelled `sibling layout`: an Edit to `<parent>/<repo>-worktrees/<change>/x` (existing worktree, and a not-yet-existing one) is allowed, and the hook's suggestion text names `<repo>-worktrees`. `mods/subagent-board/hooks/register.test.tsx` — add a `sibling layout` case: the pending rows are read from `../<repo>-worktrees/<change>/spectre/changes/<change>/tasks.md` relative to the main checkout. Run both red (`scripts/test-protect-main-checkout.sh`; `cd mods/subagent-board && npx vitest run hooks/register.test.tsx` — use the command the mod's `package.json` declares).**
  - [x] **Step 2: `hooks/protect-main-checkout.py` — the not-yet-existing-worktree rule recognises the nearest existing ancestor named `<basename top>-worktrees` beside the main checkout (and keeps `.worktrees` only as the pre-migration name it no longer suggests); the suggested `git worktree add` names `$(dirname top)/$(basename top)-worktrees/<change>`. Update the module docstring.**
  - [x] **Step 3: `mods/subagent-board/hooks/register.tsx` — read the plan from `../${basename(cwd)}-worktrees/${change}/…`; update the comment.**
  - [x] **Step 4: `.gitignore` — drop the `.worktrees/` entry and its comment. `scripts/installer-sandbox-diff.sh` — the usage example names the sibling path. Verify: `scripts/test-protect-main-checkout.sh && (cd mods/subagent-board && npx vitest run hooks/register.test.tsx)`.**
**Build:** green
**Files:** `hooks/protect-main-checkout.py`, `scripts/test-protect-main-checkout.sh`, `mods/subagent-board/hooks/register.tsx`, `mods/subagent-board/hooks/register.test.tsx`, `.gitignore`, `scripts/installer-sandbox-diff.sh`, `README.md`
**Tests:** `sibling layout`
**Regression:** reverting the commit makes the hook block an Edit into a not-yet-existing sibling worktree and suggest `.worktrees/`, and the board reads no pending rows for a sibling worktree — both `sibling layout` cases fail.
**Baseline:** before=0 after=9
<!-- measured: grep -ho 'sibling layout' scripts/test-protect-main-checkout.sh mods/subagent-board/hooks/register.test.tsx | wc -l @ branch spectre/kan-916-per-worktree-lsp -->
**Commit:** feat(hooks): follow the sibling worktree layout
**After:** Task 7

Correction (2026-10-07): the plan declared `npx vitest run hooks/register.test.tsx` and a relative `../${basename(cwd)}-worktrees/…` board read; the mod has no `package.json`, so its tests run as `claude plugin test mods/subagent-board`, and the board reads the absolute `${await $.session.root()}-worktrees/${change}/…` because `$.fs.read` resolves relative paths against the mod directory, and `session.root()` is the launch directory where `session.cwd()` follows a `cd`. The hook's `<X>-worktrees` rule requires `X` to be a main checkout's toplevel, so a `lib-worktrees/` inside a protected checkout stays protected (case 47). `README.md`'s `mods/` row named the old path and moved with it.

**Decision:** sibling-worktrees-layout

- [x] 4. `lspmux`: framing, routing, lazy children, static initialize
  - [x] **Step 1: Tests first in `stats/internal/lspmux/mux_test.go`, against a fake LSP child — the test binary re-exec'd with an env switch (`TestMain`), which records every message it receives to a file and answers requests with its own root in the result: `TestStaticInitialize` (initialize answered by the mux with the fixed capability set, no child started), `TestRouteByWorktreeRoot` (didOpen/hover for files in two git worktrees reach two children; a file in no repository reaches the session-root child), `TestLazySpawnRewritesRoot` (the child's `initialize` carries `rootUri`/`workspaceFolders` of its worktree and runs with that cwd), `TestServerRequestsAnsweredLocally` (`workspace/configuration` → one null per item; `window/workDoneProgress/create`, `client/registerCapability`, `client/unregisterCapability`, `window/showMessageRequest` → null; `workspace/applyEdit` → `{"applied":false}`; nothing reaches the client; `textDocument/publishDiagnostics` is forwarded).**
  - [x] **Step 2: Implement `stats/internal/lspmux`: `Content-Length` framing reader/writer; `Mux` with `Run(ctx, in io.Reader, out io.Writer)`; root resolution `git -C <dir> rev-parse --show-toplevel` cached per directory; children `exec.Cmd` with `Dir = root`, `initialize` params copied from the client's with root fields rewritten, then `initialized`; client→child request ids remapped per child; `shutdown`/`exit` fan out to every child. Uses the standard library only.**
  - [x] **Step 3: Verify: `cd stats && go test ./internal/lspmux/ -run 'TestStaticInitialize|TestRouteByWorktreeRoot|TestLazySpawnRewritesRoot|TestServerRequestsAnsweredLocally' -race -count=1 && gofmt -l . && go vet ./internal/lspmux/`.**
**Build:** green
**Files:** `stats/internal/lspmux/mux.go`, `stats/internal/lspmux/framing.go`, `stats/internal/lspmux/mux_test.go`
**Tests:** `TestStaticInitialize`, `TestRouteByWorktreeRoot`, `TestLazySpawnRewritesRoot`, `TestServerRequestsAnsweredLocally`, `TestRespawnReopensDocuments`
**Regression:** reverting the commit deletes the package; every declared test is gone with it and nothing else compiles against it yet.
**Baseline:** before=0 after=6
<!-- measured: find stats/internal/lspmux -name '*_test.go' -exec grep -h '^func Test' {} + 2>/dev/null | wc -l @ branch spectre/kan-916-per-worktree-lsp -->
**Commit:** feat(lspmux): route LSP traffic to a server per git worktree
**After:** none

Correction (2026-10-07): the baseline counts `TestMain` (after=5). Review found every stdin write held `child.mu`, which the reader answering a server request also needs: a server writing while not reading stalled the whole mux. Each child now has its own writer goroutine draining an outbox, pinned by `TestBlockingServerNeverStallsTheMux`.

**Decision:** wrapper-per-session

**Decision:** static-initialize

- [x] 5. `lspmux`: readiness, index log, symbol fan-out, call hierarchy, stop
  - [x] **Step 1: Tests first in `stats/internal/lspmux/ready_test.go`: `TestReadinessWaitsForProgressEnd` (a request is answered only after the fake child's `$/progress` begin…end and the settle window; a child that reports no progress is ready after the settle window), `TestIndexLogLine` (one `<RFC3339> <server> <root> ready <ms>` line appended to the log path, which tests set through an option), `TestWorkspaceSymbolFanOut` (two running children, results concatenated, one child's error does not drop the other's results), `TestCallHierarchyRoutesByItem` (`callHierarchy/incomingCalls`/`outgoingCalls` go to the child owning `item.uri`), `TestRemovedRootStopsChild` (removing a worktree directory stops its child within the watcher interval; a later request for a still-existing root still works), `TestStopUnderKillsChildren` (`StopUnder(dir)` kills a real child process whose cwd is under `dir`, leaves one outside it, and `scripts/check-worktree-processes.sh <dir>` then prints `CLEAR:`).**
  - [x] **Step 2: Implement in `stats/internal/lspmux/ready.go` and `stats/internal/lspmux/stop.go`: per-child progress-token set with a settle timer; queued requests released on ready; log path default `${XDG_CACHE_HOME:-$HOME/.cache}/worktree-lsp/index.log`; fan-out for `workspace/symbol`; item-URI routing for call hierarchy; a ticker that stats each root; `StopUnder(dir string) error` — find processes by `lsof -a -d cwd` whose cwd is at or under `dir` and whose parent's command is `worktree-lsp`, then SIGTERM, then SIGKILL after a grace period. unverified:confirm kotlin-lsp reports indexing through `$/progress` begin/end on a cold gymie worktree — if it does not, readiness must key on what it does report (log it in the task's report).**
  - [x] **Step 3: `stats/internal/guard/removechangeworktrees.go` calls `lspmux.StopUnder(wt)` immediately before check 6; add `TestRemoveChangeWorktreesStopsLSPChildren` to `stats/internal/guard/remove_change_worktrees_test.go`. `scripts/lib/flow-guard.sh` — `flow_guard_key` also hashes the non-test `internal/lspmux/*.go`, header comment updated.**
  - [x] **Step 4: Verify: `cd stats && go test ./internal/lspmux/ ./internal/guard/ -race -count=1 && gofmt -l . && go vet ./... && cd .. && scripts/check-guard-symlinks.sh`.**
**Build:** green
**Files:** `stats/internal/lspmux/ready.go`, `stats/internal/lspmux/stop.go`, `stats/internal/lspmux/mux.go`, `stats/internal/lspmux/ready_test.go`, `stats/internal/lspmux/mux_test.go`, `stats/internal/guard/removechangeworktrees.go`, `stats/internal/guard/remove_change_worktrees_test.go`, `scripts/lib/flow-guard.sh`, `scripts/remove-change-worktrees.sh`, `scripts/test-lib-flow-guard.sh`
**Tests:** `TestReadinessWaitsForProgressEnd`, `TestIndexLogLine`, `TestWorkspaceSymbolFanOut`, `TestCallHierarchyRoutesByItem`, `TestRemovedRootStopsChild`, `TestStopUnderKillsChildren`, `TestRemoveChangeWorktreesStopsLSPChildren`
**Regression:** reverting the commit answers requests before indexing ends, sends workspace symbol to one child, never stops a child, and leaves cleanup's check 6 `HELD` by a wrapper child — every declared test fails.
**Baseline:** before=165 after=172
<!-- measured: find stats/internal/lspmux stats/internal/guard -name '*_test.go' -exec grep -h '^func Test' {} + 2>/dev/null | wc -l @ branch spectre/kan-916-per-worktree-lsp (156 at the merge base; 163 once tasks 1, 2 and 4 land) -->
**Commit:** feat(lspmux): wait for indexing, fan out symbols, stop servers at cleanup
**After:** Task 4

Correction (2026-10-07): the mux sets `window.workDoneProgress` in every server's `initialize`, since gopls reports progress only to a client that declares it. A worktree whose `## stop` failed keeps its servers, since it is not removed (`TestRemoveChangeWorktreesStopsLSPChildren`'s second subtest); a held request for a server that dies before it is ready gets an error answer (`TestQueuedRequestFailsWhenServerDiesBeforeReady`). measured: kotlin-lsp on a cold gymie clone reports `$/progress` "Importing" from 8.55s and "Indexing" from 73s, both ended at 248.4s, then short "Indexing" bursts for the rest of a 25-minute probe; with the 2s settle a root is ready at ~250.6s and stays ready. `mux_test.go`'s helper moved to a temp index log and a 1ms settle, `scripts/remove-change-worktrees.sh`'s header names the stop, and `scripts/test-lib-flow-guard.sh` case 5b pins the cache key over `internal/lspmux/*.go`.

**Decision:** readiness-progress-settle

**Decision:** workspace-symbol-fanout

**Decision:** stop-at-cleanup-and-removal

- [x] 6. `worktree-lsp` command and its skills-dir plugin
  - [x] **Step 1: `stats/cmd/worktree-lsp/main_test.go` — `TestUsage`: no `--` or no server argument exits 2 with a usage line on stderr. Run red.**
  - [x] **Step 2: `stats/cmd/worktree-lsp/main.go` — `worktree-lsp -- <server> [args…]` runs `lspmux.Mux` on stdin/stdout with the server command; stderr for its own log lines.**
  - [x] **Step 3: `mods/worktree-lsp/.claude-plugin/plugin.json` — `name` `worktree-lsp`, `lspServers.gopls` (`command` `worktree-lsp`, `args` `["--","gopls"]`, `extensionToLanguage` `{".go":"go"}`) and `lspServers.kotlin-lsp` (`args` `["--","kotlin-lsp","--stdio"]`, `.kt`/`.kts` → `kotlin`), each with `requestTimeout` raised to 30 minutes and `startupTimeout` covering a first `go run` build (values in milliseconds).
    <!-- predicted: a cold gymie Kotlin index finishes well inside 30 minutes — the live check's index.log line confirms it --> `mods/worktree-lsp/bin/worktree-lsp` — executable shim: resolve its own physical path (`cd -P`), `cd` to the checkout's `stats/`, `exec go run ./cmd/worktree-lsp "$@"` with `GOOS`/`GOARCH`/`GOFLAGS` unset.**
  - [x] **Step 4: `setup.sh` — after `install_mods`, when `~/.claude/settings.json` enables `gopls-lsp@claude-plugins-official` or `kotlin-lsp@claude-plugins-official`, print a warning naming `claude plugin disable gopls-lsp@claude-plugins-official` / `claude plugin disable kotlin-lsp@claude-plugins-official`; never edit the file (`install_hooks`' rule). Verify: `cd stats && go test ./cmd/worktree-lsp/ -race -count=1 && gofmt -l . && go vet ./cmd/worktree-lsp/ && cd .. && SANDBOX="$(mktemp -d)" && HOME="$SANDBOX" ./setup.sh global >/dev/null && test -L "$SANDBOX/.claude/skills/worktree-lsp" && scripts/installer-sandbox-diff.sh`, then a live start: `printf 'Content-Length: …' | mods/worktree-lsp/bin/worktree-lsp -- gopls` answers an `initialize`.**
**Build:** green
**Files:** `stats/cmd/worktree-lsp/main.go`, `stats/cmd/worktree-lsp/main_test.go`, `mods/worktree-lsp/.claude-plugin/plugin.json`, `mods/worktree-lsp/bin/worktree-lsp`, `setup.sh`, `README.md`, `stats/internal/setuptest/containment_test.go`
**Tests:** `TestUsage`
**Regression:** reverting the commit removes the command and the plugin: Claude Code falls back to the official plugins and `TestUsage` is gone.
**Baseline:** before=0 after=1
<!-- measured: find stats/cmd/worktree-lsp -name '*_test.go' -exec grep -h '^func Test' {} + 2>/dev/null | wc -l @ branch spectre/kan-916-per-worktree-lsp -->
**Commit:** feat(worktree-lsp): ship the wrapper as a skills-dir LSP plugin
**After:** Task 3, 5, 7

Correction (2026-10-07): the shim builds and `exec`s the wrapper rather than `go run` (`plugin-as-mod-build-exec`), so SIGTERM reaches it. A plugin's `bin/` reaches only the Bash tool's PATH, never an LSP server's (Claude Code 2.1.292 bundle), so each `lspServers` `command` is `${CLAUDE_PLUGIN_ROOT}/bin/worktree-lsp`; `startupTimeout` 120000, `requestTimeout` 1800000. `scripts/installer-sandbox-diff.sh` takes `<old-tree> <new-tree>`; the sandbox diff against the task-5 tree differs by the one `worktree-lsp` link. The setup warning is pinned in `stats/internal/setuptest/containment_test.go`, and `README.md`'s `mods/` row names the plugin.

**Decision:** plugin-as-mod-go-run

**Decision:** plugin-as-mod-build-exec

**Decision:** official-plugins-warned

- [x] 7. Skills and README: the sibling layout
  - [x] **Step 1: Replace every `<project>/.worktrees/<…>` and `.worktrees/` path in the files below with `<project>-worktrees/<…>` (the sibling of the main checkout) and drop every statement about ignoring `.worktrees` through `info/exclude` (kickoff step 2). `skills/flow-fast/SKILL.md`'s `git worktree add <project>/.worktrees/<name>` becomes the sibling path. Leave archived change directories, `docs/` reports and applied migrations untouched — they are records.**
  - [x] **Step 2: Verify: `grep -rn '\.worktrees' README.md skills/` prints nothing; `scripts/check-references.sh && scripts/check-markdown-integrity.py && scripts/check-installed-citations.sh && scripts/check-normative-inventory.sh && scripts/check-verbatim-moves.sh`.**
**Build:** green
**Files:** `README.md`, `skills/flow-contracts/finish-contract-rationale.md`, `skills/flow-contracts/finish-contract-run1.md`, `skills/flow-contracts/git-boundaries.md`, `skills/flow-contracts/pipeline.md`, `skills/flow-contracts/state-file-internals.md`, `skills/flow-contracts/state-file.md`, `skills/flow-fast/SKILL.md`, `skills/flow-plan/SKILL.md`, `skills/flow-self-review/SKILL.md`, `skills/flow/brainstorm-planner.md`, `skills/flow/brainstorm.md`, `skills/flow/cross-repo-worktrees.md`, `skills/flow/implement.md`, `skills/flow/resume.md`, `skills/flow/SKILL-rationale.md`, `stats/internal/guard/installedcitations.go`, `stats/internal/guard/check_installed_citations_test.go`
**Allowed-collateral:** `AGENTS.md`, `CLAUDE.md`, `rules/flow-manual-review.mdc`
**Tests:** `TestCheckInstalledCitations`
**Regression:** reverting the commit drops `<project>-worktrees` from the citation guard's placeholder roots: the new `TestCheckInstalledCitations` subtests fail and the skills' sibling citations are reported as naming no root.
**Baseline:** before=0 after=0
**Commit:** docs(flow): name the sibling worktree layout
**After:** Task 2

Correction (2026-10-07): `check-installed-citations.sh` rejected every `<project>-worktrees/…` citation (19, "names no root"), so `<project>-worktrees` and `<agents repo>-worktrees` joined `cicPlaceholderRoots` with three subtests, a lookalike `<project>-worktrees-old/…` still reported. `skills/flow-status/SKILL.md` is untouched: its one match is the state record's jq field `.worktrees`, so Step 2's grep always prints that line. The rewording is acknowledged in `verbatim-moves.txt`.

**Decision:** sibling-worktrees-layout

- [x] 8. Skills: kickoff waits for the LSP index; cleanup re-runs the migration
  - [x] **Step 1: `skills/flow/brainstorm.md` section A, after `kickoff-worktree.sh` and still inside `flow.kickoff`: for each wrapped language (`*.go`, `*.kt`) with a tracked file in the new worktree, call the LSP tool `documentSymbol` on the first such file `git -C <abs-worktree> ls-files` prints; the call returns once the server is ready; copy the matching `index.log` ready line into the run's narrative; a language with no tracked file, or no LSP tool in the session, is skipped with one line. `skills/flow-fast/SKILL.md` — the same step after its worktree add. `skills/flow/cross-repo-worktrees.md` — the step runs once per repository worktree.**
  - [x] **Step 2: `skills/flow-contracts/finish-contract-run2.md` — check 6 states that `remove-change-worktrees` stops the worktree's `worktree-lsp` children first; run 2 runs `migrate-worktrees.sh <main-checkout>` once per repository before its worktree removal, so a worktree an older kickoff created under `.worktrees/` is moved before the strict guard reads it.**
  - [x] **Step 3: Verify: `scripts/check-references.sh && scripts/check-markdown-integrity.py && scripts/check-stage-mark-calls.sh && scripts/check-guard-symlinks.sh && scripts/check-normative-inventory.sh`.**
**Build:** green
**Files:** `skills/flow/brainstorm.md`, `skills/flow-fast/SKILL.md`, `skills/flow/cross-repo-worktrees.md`, `skills/flow-contracts/finish-contract-run2.md`, `skills/flow/cleanup.md`, `skills/flow/scripts/migrate-worktrees.sh`, `stats/internal/guard/check_guard_symlinks_test.go`
**Allowed-collateral:** `AGENTS.md`, `CLAUDE.md`, `rules/flow-manual-review.mdc`
**Tests:** none — repairs existing tests: `TestShimSiblingsDeclared`
**Regression:** none — the task declares no tests
**Baseline:** before=0 after=0
**Commit:** docs(flow): wait for the worktree's LSP index at kickoff
**After:** Task 6, 7, 9

Correction (2026-10-07): `check-guard-symlinks.sh` rule 2 never scans `skills/flow-contracts/`, so `finish-contract-run2.md` naming the guard required no symlink; `skills/flow/cleanup.md` names `migrate-worktrees.sh` beside `remove-change-worktrees.sh`, which makes the rule require `skills/flow/scripts/migrate-worktrees.sh`, and `TestShimSiblingsDeclared` gains its `lib`/`check-worktree-processes.sh` row.

Correction (2026-10-07, panel round 1): the migration step moved out of run 2 — the finish preflight refuses a `.worktrees/` worktree as a stray before run 2 is ever entered — to **Migrate retired-layout worktrees before the preflight** in `skills/flow-contracts/finish-contract-run1.md`, cited from `skills/flow/integrate.md` just before `check-finish-preflight.sh`, as one `migrate-worktrees.sh` call naming every distinct main checkout (`design.md`: migrate-before-preflight). `skills/flow/integrate.md` now names the guard, keeping `skills/flow/scripts/migrate-worktrees.sh` required by `check-guard-symlinks.sh` rule 2. `/flow-fast` runs neither the location guard nor the preflight and is unchanged.

**Decision:** kickoff-warmup-via-lsp-tool

**Decision:** stop-at-cleanup-and-removal

**Decision:** migrate-all-now

- [x] 9. `migrate-worktrees`: move every `.worktrees/` worktree to the sibling layout
  - [x] **Step 1: `stats/internal/guard/migrate_worktrees_test.go` — `TestMigrateWorktrees`: in a temp repository with two `.worktrees/` worktrees and a fake `flow` on `PATH` (records its arguments, serves a state record whose `worktrees` map names one of them): the free one is moved to `<parent>/<repo>-worktrees/<name>` (`git worktree list` shows the new path), the record is re-`set` with the key rewritten and its merge-base value kept; a worktree a child process holds (cwd inside it) is skipped with a `HELD:` line and left in place; exit 0; `MIGRATED: <old> -> <new>` per move. Run red.**
  - [x] **Step 2: `stats/internal/guard/migrateworktrees.go`, registered as `migrate-worktrees`: `migrate-worktrees <main-checkout>`; for each registered worktree under `<main>/.worktrees/`: `check-worktree-processes.sh` verdict `HELD:` → print and skip; else `git worktree move <old> <new>`; then for each record `flow state list -C <main>` returns whose `worktrees` map has `<old>`, `flow state get` → rewrite the key → `flow state set -C <main> <name>`; finally remove `<main>/.worktrees` when empty. Exit 0 when every worktree was moved or held, 1 when a move or a record rewrite failed, 2 when it cannot answer. `scripts/migrate-worktrees.sh` is the `flow_guard_exec` shim. unverified:confirm `flow state get` output is accepted back by `flow state set` unchanged but for `worktrees` (drop `synthetic` records rather than writing them).**
  - [x] **Step 3: Symlink `scripts/migrate-worktrees.sh` into every skill `scripts/` directory whose files name it, as `scripts/check-guard-symlinks.sh` requires. The real migration of the five repositories is not run here — `flow.verify`'s live check runs it (`design.md` `## Live check`).**
  - [x] **Step 4: Verify: `cd stats && go test ./internal/guard/ -run TestMigrateWorktrees -race -count=1 && gofmt -l . && go vet ./internal/guard/ && cd .. && scripts/check-guard-symlinks.sh`.**
**Build:** green
**Files:** `stats/internal/guard/migrateworktrees.go`, `stats/internal/guard/migrate_worktrees_test.go`, `scripts/migrate-worktrees.sh`
**Allowed-collateral:** `skills/*/scripts/migrate-worktrees.sh`
**Tests:** `TestMigrateWorktrees`
**Regression:** reverting the commit removes the guard: `TestMigrateWorktrees` is gone and a `.worktrees/` worktree stays where the strict guard reports it `STRAY`.
**Baseline:** before=161 after=162
<!-- measured: grep -h '^func Test' stats/internal/guard/*_test.go | wc -l @ branch spectre/kan-916-per-worktree-lsp (156 at the merge base; 160 once tasks 1, 2 and 5 land) -->
**Commit:** feat(guard): migrate .worktrees worktrees to the sibling layout
**After:** Task 2, 5

Correction (2026-10-07): every record is read (`flow state list` plus one `get` per change) before anything moves, so an incomplete list or a failed get exits 2 with nothing moved; an occupied destination is refused (`FAILED: <old> — <new> already exists`, exit 1), since `git worktree move` into an existing directory nests the worktree and exits 0; a process check that cannot answer leaves the worktree in place with a `FAILED:` line, exit 1; a nested path keeps its relative path under `<repo>-worktrees/`. measured: `flow state get` prints `toDTO`'s shape, accepted back by `state set` once `synthetic` records are dropped. No skill names the guard yet, so Step 3 needed no symlink; task 8 wires it into cleanup run 2.

Correction (2026-10-07, panel round 1): `migrate-worktrees <main-checkout> [<main-checkout>…]` takes every repository at once — moves each one's `.worktrees/` worktrees, builds one rename map from every repository's `git worktree list`, and rewrites the records of every listed project — so a cross-repo change's record, which names another repository's worktree, is rewritten too; `TestMigrateWorktrees` gains a two-repository subtest. The sibling root comes from one `siblingRoot` helper shared with the location guard and kickoff.

**Decision:** migrate-all-now

- [x] 10. In-run fix 1 — flow.verify: `migrate-worktrees` loses its own scripts directory when it moves the worktree it runs from
  - [x] **Step 1: `stats/internal/guard/migrate_worktrees_test.go` — a `TestMigrateWorktrees` subtest: the guard runs from a `scripts/` directory inside one of the `.worktrees/` worktrees it migrates (the shim's `SCRIPT_DIR`, so `check-worktree-processes.sh` resolves there), sorted before a second repository's worktree; both are `MIGRATED`, no `FAILED: … check-worktree-processes.sh could not answer`, exit 0. Run red.**
  - [x] **Step 2: `stats/internal/guard/migrateworktrees.go`: after each successful `git worktree move <old> <new>`, a scripts directory under `<old>/` is re-pointed to the same relative path under `<new>/` (compared physically), so every later process check resolves `check-worktree-processes.sh` at its new path.**
  - [x] **Step 3: Verify: `cd stats && go test ./internal/guard/ -run TestMigrateWorktrees -race -count=1 && gofmt -l . && go vet ./internal/guard/`.**
**Build:** green
**Files:** `stats/internal/guard/migrateworktrees.go`, `stats/internal/guard/migrate_worktrees_test.go`
**Tests:** `TestMigrateWorktrees`
**Regression:** reverting the commit brings back the live failure: every worktree migrated after the one holding the guard reports `FAILED: … check-worktree-processes.sh could not answer` and stays in `.worktrees/`.
**Baseline:** before=162 after=162
<!-- measured: grep -h '^func Test' stats/internal/guard/*_test.go | wc -l @ branch spectre/kan-916-per-worktree-lsp -->
**Commit:** fix(guard): re-point migrate-worktrees at its moved scripts directory
**After:** Task 9

Found by flow.verify's live check (2026-10-07). measured: `migrate-worktrees.sh` run from this change's worktree over agents and the four gymie repositories moved all ten agents worktrees, this one included, then printed `FAILED: <wt> — check-worktree-processes.sh could not answer` for all fourteen gymie worktrees, exit 1; `check-worktree-processes.sh` at the moved path answers `CLEAR` for the same worktree.

**Decision:** migrate-before-preflight
