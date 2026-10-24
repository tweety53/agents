## Context

Self-review fix change for `kan-1029-bugfixes-group-copy-eval-note-block-times`, planned on
`self-review-kan-1029-bugfixes-group-copy-eval-note-block-times` at agents `a988a619`
(`origin/main`). Nine non-`big` findings, F1–F9 as `proposal.md` names them, each carried by one
task in `tasks.md`; F2 and F7 land on gymie's own change branch
`spectre/kan-1029-bugfixes-group-copy-eval-note-block-times` (frontend at `a66d4aeb`, backend at
`02afc59b`), F8 is a store write. Each was traced before its fix was designed; `origin/main` holds
no fix for any of them, and the one archived statement on F3 — kan-1028's "`mutation`'s pass is
never rendered … stays parent-typed" — missed that a parent handed a rendered bundle path types
nothing more. Every question was auto-resolved to its recommended option (**Auto-resolution**,
`skills/flow-contracts/operator-prompts-auto-resolution.md`). The brainstorm design, with the
per-finding trace, is
`.superpowers/sdd/2026-10-24-self-review-kan-1029-bugfixes-group-copy-eval-note-block-times-design.md`
in this worktree. Every fix was built and tested in this worktree and saved as
`patches/task-<n>.diff`; every reworded or added run-loaded sentence is listed in
`verbatim-moves.txt`.

## Decisions

### flow.verify's stage end runs the live check in-process

**ID:** verify-end-runs-live-check
**Status:** active
**Chosen:** `gateLiveVerification` in `stats/cmd/flow/stage.go`: a `completed` end of `/flow`'s `flow.verify` runs `guard.Registry["check-live-verification"]` on the change's directory under `-C` (`planPath`'s directory) and refuses, exit 1 and no store call, on the guard's exit 1 or 2. `skills/flow/verify-and-handoff.md` drops the chained shell call for one `flow stage end -C <canonical-worktree> …`.
**Considered:** refusing on exit 1 only — exit 2 is what a stage end run from any directory not holding the change gets, so the gate would be skipped by running it from the main checkout; a PreToolUse hook demanding the chain — a second mechanism matching command text, where the CLI owns the stage end; gating `/flow-fast`'s `flow.verify` too — it has no change directory to judge.

### The gymie-frontend suite drops the iOS simulator test run

**ID:** frontend-suite-drops-ios-link
**Status:** active
**Chosen:** `## test` runs `./gradlew test allTests :shared:compileTestKotlinIosSimulatorArm64 -x :shared:iosSimulatorArm64Test`: excluding the run task drops its `linkDebugTestIosSimulatorArm64` dependency, and naming the compile keeps it (`--dry-run` measured). The memory file `gymie-targets-pwa-and-web-only.md` is unchanged — it records the operator's instruction, not the defect.
**Considered:** `-x :shared:linkDebugTestIosSimulatorArm64` alone — Gradle then still schedules `iosSimulatorArm64Test`, which has nothing to run against; dropping `allTests` — it is also what runs `desktopTest` and `wasmJsTest`.

### The mutation pass is rendered

**ID:** render-mutation-pass
**Status:** active
**Chosen:** `render-slot-prompt` renders `## PASS mutation` in the order given: the body of review-panel.md's `### The mutation-testing brief`, one line naming the throwaway copies `<worktree>-mutation-<round>` per worktree given, and the REPORT FILE block; review-panel.md carrying no brief is exit 2. The MUTATION ENTRY CONTEXT paragraph and its inlined tests stay typed beside the rendered path.
**Considered:** rendering MUTATION ENTRY CONTEXT too — its tests come from the plan's `**Tests:**` fields and the touched test files, which the renderer does not read; a guard that refuses a bundle naming mutation unless the parent typed a brief — checks the prompt after the fact, where rendering removes the step.

### The hook accepts only the closed list's key shapes

**ID:** hook-closed-list-keys
**Status:** active
**Chosen:** `ROW_DESCRIPTION` in `hooks/enforce-agent-baseline.py` accepts `Task(s) <ids>/<n>` or a first word matching a Key shape of **Dispatch sites — the parent's closed list** (`skills/flow/implement.md`): `task-<ids>-implementer|reviewer[-…]`, `panel-<round>-<slots>[-cite]`, `panel-fix-<round>[-<chunk>][-retry]`, `visual-verify[-…]`, `tooling-analysis-<n>`, `pipeline-fix-…`. kan-1029's `verify-fix-1 …` and `verify live check …` are refused.
**Considered:** the optional harvester flag for a keyed launch with no row — the four other launches (`Tasks 9+10/11 (review)`, `Task 11/11 (review)`, `panel-fix-1-2`, and `visual-verify-gymie-frontend-rest` with its free-form repository suffix) match a prescribed shape, and telling them apart needs the run's class and repository names, a flowd change beyond the finding's fix; reading the closed list's table at hook time — a hook that parses Markdown on every dispatch, for a table that changes with the shapes it would read.

### The harvester reads strips of a loop variable

**ID:** harvest-loop-var-strips
**Status:** active
**Chosen:** `substituteLoopVar` (`stats/internal/harvest/watcher.go`) applies `${v#pat}`, `${v##pat}`, `${v%pat}` and `${v%%pat}` with `*` and `?` globs, so kan-1029's `-key ${k#*:}` reads `task-9-implementer`; `skills/flow/implement.md` names the loop-variable shapes the harvester reads.
**Considered:** a PreToolUse refusal of a non-literal `-key` — the kan-925 loop shape the harvester already reads would be refused too, and no Bash hook reads flow commands today; a warning in the harvester log — nobody reads it, and the ledger's "unattributed" count already says the same.

### Rows 7367 and 7368 are restamped

**ID:** k1029-dispatch-rows-restamp
**Status:** active
**Chosen:** `flowd restamp-dispatch 7367 …/agent-a6a4cc4cd5fa75a38.jsonl` and `7368 …/agent-ac28906c875a56d4f.jsonl` — the two launches described `Tasks 9+10/11 (Save focus, repeat both)` and `Task 11/11 (web-edge tests)`, both first written at 2026-10-23T20:20:58Z, five seconds before the begins; no row holds either id.
**Considered:** leaving them — a forward-only fix leaves the record wrong (**3. Fix every finding that is not `big`**, `skills/flow-self-review/SKILL.md`).

### A stage mark's token must end in a UTC stamp

**ID:** stage-token-utc-stamp-enforced
**Status:** active
**Chosen:** `validateSessionTokenStamp` refuses, exit 2 before any store call, a `stage begin` or `stage wrap` token not ending `-\d{8}T\d{6}Z` (`stage mark` routes through begin); **Stage marks** (`skills/flow-contracts/pipeline.md`) states the refusal. A run whose token predates kan-1028's rule reads the clock at its next mark and marks under the stamped token from there on: the harvester binds each token to the transcript it appears in, so both tokens bind to the same session. Record commands keep the substitution check alone.
**Considered:** exempting a legacy token the store already holds — needs a token field on the stage-runs listing and a query per mark, for a transition that ends with the in-flight runs; enforcing in the server's `validateSessionTokenShape` too — a journalled mark replayed after the change would be refused; enforcing on record commands — a mid-run record refused for a token the run's stage marks already carry would lose a dispatch row.

### Rows 17078 and 17079 are retokened

**ID:** k1029-stage-rows-retoken
**Status:** active
**Chosen:** a new `flowd retoken-stage-run <id> <token>` (`store.RetokenStageRun`, one `UPDATE`, `ErrStageRunNotFound` on an unknown id) sets both rows to `mf-k1029-20261024T002513Z` — the token the same session (`472ec71d-…`) used fourteen seconds later for the identical `flow.preflight` and `flow.unfinished-work-gate` pair. No command could change a stage run's token.
**Considered:** a raw `UPDATE` through `docker exec` — an unrecorded hand edit of the store; deleting the two rows — the session's first preflight did run.

### devPhone reverses the resolved ports to themselves

**ID:** devphone-resolved-ports
**Status:** active
**Chosen:** `devPhone` reverses the workspace's resolved `frontend` and `gateway` ports (`resolveWorkspaceBlock(workspace(), emptySet())`, the read `devStatus` makes) and 9000, each to the same number, and prints `http://localhost:<frontend port>`; the main checkout stays 3000/8080/9000.
**Considered:** the findings file's remap, phone 3000→frontend and 8080→gateway — the page calls the gateway on the port `gatewayPort.js` injects and, under `-PisolateOAuthPorts`, the gateway's CORS allows only the resolved frontend origin, so a page opened at `localhost:3000` cannot sign in at all; with this choice a moved stack signs in by password, as on the desktop, since Google's OAuth client knows only 3000 and 8080.

### gymie carries a hazard for `./gradlew --stop`

**ID:** gymie-gradle-stop-hazard
**Status:** active
**Chosen:** `flow hazard add -name gradle-stop-kills-dev-stack -applies all` for gymie, text naming that `--stop` stops every daemon of the wrapper's Gradle version, the frontend's `wasmJsBrowserDevelopmentRun` dev server among them, and kan-1029's verifier killing it.
**Considered:** a sentence in gymie's `CLAUDE.md` — a dispatched agent reads its bundle's hazards, where it may never read the project file.

### The zip rebuild refuses a root no PNG lies under

**ID:** zip-root-holds-the-pngs
**Status:** active
**Chosen:** `rebuild-full-app-suite-zip` exits 2, before the old zip is touched, when a PNG `resolve-visual-screenshots.sh` printed lies outside `<screenshots root>`. The parent deletes the stray `/Users/tweety53/Projects/gymie/full-app-suite.zip` itself.
**Considered:** refusing a root holding no `full-app-suite.spec.ts` — the spec lives in the regression checkout while the root may be a subdirectory of it (the test fixture's `shots/`), so presence of the spec is not the root's property; the PNGs the zip is built from are.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — Task 1 and Task 2 change what `flow stage` refuses on every run, Task 2 adds a store write command, and Tasks 2 and 5 write to the live store.
**Considered:** `flow-low` — most diffs are small, but a refusal shape wrong by one character stops every `/flow` run at its first mark.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, one fixer — key `self-review-kan-1029-bugfixes-group-copy-eval-note-block-times-fix`, Tasks 1 to 9 in order: Tasks 1–2 share `stats/cmd/flow/stage.go` and its test, Tasks 4–5 share `skills/flow/implement.md`, and the patches were built in that order.
**Considered:** one fixer per repository — three dispatches for two one-file gymie diffs and a command.

## Open questions

## Live check

Each step runs before landing, from this worktree with every task committed (`cd stats`,
`make -s web-dist` first):

1. **Stage-end gate.** In a scratch git repo holding `spectre/changes/x/design.md` with a `## Live
   check` section and no record, `go run ./cmd/flow stage end -addr http://127.0.0.1:9 -C <scratch>
   -command /flow -stage flow.verify -outcome completed x` — before: exit 0 (the
   old binary); after: exit 1 naming `check-live-verification`. Not working looks like exit 0.
2. **Token stamp.** `go run ./cmd/flow stage begin -addr http://127.0.0.1:9 -C <scratch> -command
   /flow -stage flow.preflight -harness claude-code -session-token mf-x-PLACEHOLDER x` — after:
   exit 2 naming `UTC`; with `mf-x-20261024T000000Z`: exit 0 (journalled, store unreachable).
3. **Store repairs.** `docker exec flow-postgres psql -U flow -d flow -c "select id, session_token
   from stage_runs where id in (17078,17079)"` — before: `mf-k1029-PLACEHOLDER` twice; after:
   `mf-k1029-20261024T002513Z` twice. `… "select id, agent_id from dispatches where id in
   (7367,7368)"` — before: both empty; after: `a6a4cc4cd5fa75a38`, `ac28906c875a56d4f`.
4. **Hazard.** `flow hazards -C /Users/tweety53/Projects/gymie` — before: `[]`; after: one row
   named `gradle-stop-kills-dev-stack`.
