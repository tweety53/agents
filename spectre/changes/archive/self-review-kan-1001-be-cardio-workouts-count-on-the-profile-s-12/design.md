## Context

Found by the self-review pass of `/flow` change `kan-1001-be-cardio-workouts-count-on-the-profile-s-12`
(gymie). The agents fixes ride branch `self-review-kan-1001-be-cardio-workouts-count-on-the-profile-s-12`,
based at `origin/main` `1c4c3e38`; the gymie fixes ride that change's own branch
`spectre/kan-1001-be-cardio-workouts-count-on-the-profile-s-12` in
`/Users/tweety53/Projects/gymie-worktrees/kan-1001-be-cardio-workouts-count-on-the-profile-s-12`
(archived at `94c2e80f`), landing with its merge and push. Every patch under `patches/` beside this
file was applied, tested and reverted at plan time — the gymie ones in order, 4 then 5 then 6, each
on the tree the one before leaves.

**Finding P1 — the pass cannot be closed (flow-fix, traced to a verdict: real).** `checkLoopDispatch`
(`stats/internal/guard/loop.go`) returns `LOOP-DISPATCH-OK` for the `pass` family before its `flow
record dispatch begin` call, deliberately — `TestCheckLoopDispatch`'s case "the pass is checked,
never recorded", from `96742f7d` — since the pass runs before any loop change exists to file a row
under. `pfEnd` (`stats/internal/guard/pipelinefix.go`) nonetheless calls `flow record dispatch end`
for it, which answers 404 wherever it runs. No session-token or project-key mismatch is involved: the
live store holds no `self-review-*` row at all for `kan-1001-be-cardio-workouts-count-on-the-profile-s-12`
<!-- measured: flow record dispatches -change kan-1001-be-cardio-workouts-count-on-the-profile-s-12 -C <gymie wt> | filter key~self-review @ 2026-10-21 → 4 rows, none self-review -->.
With `patches/task-1.diff` applied, `TestPipelineFixEndPassRecordsNothing` passes; with
`pipelinefix.go` reverted it fails (`rc 0:` and no `PIPELINE-FIX-END-NA` line).

**Finding F2 — data heredocs denied over their prose (flow-fix, blast radius 2 files + 1 memory).**
<!-- measured: Task 2's **Files:** names 2, plus the saved memory outside the repository @ 1c4c3e38 -->
`strip_data_heredocs` keeps a body scanned when `target in "\n".join(lines[end + 1 :])` — a substring
anywhere after the body — and `may_run_code` returns true for any path- or expansion-named command
word. Reproduced against the main checkout at `1c4c3e38`: `C=<wt>/spectre; cat > $C/t.md <<'EOF'` /
`> **Execution:** a > b; rm x` / `EOF` / `check-plan-shape.sh $C/t.md` is denied, and so is the same
body followed by `~/.claude/skills/flow/scripts/check-plan-shape.sh /x/other.md`; followed by `echo
done` it is allowed. With `patches/task-2.diff` applied all 109 harness cases pass; with the hook
reverted cases 118 and 119 fail and 120, 121 still deny.

**Finding S1 — the bundle drops live-verification.md (flow-stats-app, blast radius 5).** `archivedFiles`
lists four files; kan-1001's archive carries `live-verification.md` beside them. Adding it shifts every
"found: N of 8 sources" expectation to "of 9" in `bundle_test.go`. Two run-loaded sentences list the
bundle's archived files — `skills/flow-contracts/finish-contract-run1.md` and `skills/flow-fast/SKILL.md`
— and are reworded; `verbatim-moves.txt` lists the four FAIL sentences `check-verbatim-moves.sh` printed
with `patches/task-3.diff` applied. `go test ./internal/selfreview/ ./internal/api/ ./cmd/flow/` passes;
the new test fails with `git.go` reverted.

**Finding F1 — seeding before the .env link (gymie dev tooling, blast radius 4).** `linkEnvFile`
(`gradle/git-hooks.gradle.kts`) is a dependency of `devStart` alone; `./gradlew -q -m dbSeed` and
`./gradlew -q -m devWorkspaceCreate` at `94c2e80f` print no `:linkEnvFile` line, `devStart`'s does.
`seedDatabase` reads `DEV_INVITE_EMAIL` through `dotEnvValues()`, an empty map when `.env` is absent.
`scripts/test-env-link-wiring.sh` asserts the dry-run graph of all three tasks; it fails for `dbSeed` and
`devWorkspaceCreate` with the wiring reverted. `CiWorkflowTest` requires every `scripts/test-*.sh` in
`ci.yml`'s shell-guard step, so the step names it; `CiWorkflowTest` passes (6 tests).
<!-- measured: ./gradlew :app:test --tests 'com.gymie.CiWorkflowTest' with patches/task-4.diff applied @ 94c2e80f → TEST-com.gymie.CiWorkflowTest.xml tests=6 failures=0 -->

**Finding A1 — the hand-copied bootRun arguments (gymie dev tooling, blast radius 4).** The recipe
duplicated `seedDatabase`'s argument assembly. The builder moves to `buildSrc` as `bootRunArgs` (pure,
unit-tested like `CacheIndex`), `workspaceBootRunArgs` in `dev-lifecycle.gradle.kts` fills it from a
workspace, and both the seed run and the new `devBootRunArgs` task call it. Measured in the kan-1001
worktree: `./gradlew -q devBootRunArgs -PbootPort=18080` prints one `--args=` line naming
`gymie_kan_1001_be_5a0d`, bucket `gymie-media-kan-1001-be-5a0d` and cache index 1, matching
`devWorkspaceInventory`; without `-PbootPort` it fails "-PbootPort=<free port> is required".
`./gradlew verifyChange` is green.

**Finding F3 — workspaces outliving the before worktree (gymie dev tooling + repair, blast radius 1).**
`./gradlew devWorkspaceInventory` lists seven workspaces; one, `kan-1001-be-5a0d`, is the live change's.
`./gradlew devWorkspace -PworkspaceChange=<name>` for every live gymie and gymie-frontend worktree name
derives `kan-1001-be-5a0d`, `ci-only-on-2d89`, `docs-ce28`, `kan-627-flow-9a73`, `review-a3ed` and
`self-review-b008` — none of `kan-1001-be-9051`, `kan-1000-fe-a183`, `kan-692-flow-23fd`,
`kan-825-flow-9836`, `fix-backend-f56d`, `self-review-4fe5`.

## Decisions

### Closing the pass

**ID:** pass-end-is-a-no-op
**Status:** active
**Chosen:** `pfEnd` answers a `pass`-family key with `PIPELINE-FIX-END-NA <key>: …` and exit 0, before any `flow` call, whatever the outcome argument.
**Considered:** recording the pass in `check-loop-dispatch` — no loop change exists when the pass runs (standalone, `pipeline-fix.sh start` comes after it), so the begin would fail and deny the pass; rewording step 3 of the self-review skill to exempt the pass — a parent that still runs the end call hits the same 404, and a prose rule no tool enforces is what this loop's mechanics replaced.

### When a data heredoc's body stays scanned

**ID:** heredoc-target-run-not-named
**Status:** active
**Chosen:** drop the substring re-mention test; `may_run_code(text, target)` counts a path- or expansion-named command word only when `names_target` holds — the same spelling as the target, the same file name (`"$D"/x.sh` for `$SIB/x.sh`), or a file name containing `$`. Runner words and substitutions anywhere still keep the body scanned, which covers `bash $T`, `source $T` and `sh "$D"/x.sh`.
**Considered:** the finding's two rules kept separate (a re-mention scan plus a path rule) — the runner-word rule already covers the runner-argument case, so one predicate is the smaller change; matching the target's full spelling only — `/abs/x.sh` run after `cat > $D/x.sh` would read as data.
**Known ceiling:** a body copied under another file name and run (`cp $T /tmp/y && /tmp/y`) is no longer scanned; the hook guards against accidental main-checkout writes, not an adversary, and the copy's own command still is.

### Where the bundle lists live-verification.md

**ID:** live-verification-after-visual
**Status:** active
**Chosen:** last in `archivedFiles`, after `visual-verification.md`, absent reported `skipped` like every source.
**Considered:** before `narrative.md` — reorders every existing bundle for no reader's benefit.

### Wiring linkEnvFile

**ID:** link-env-before-seed
**Status:** active
**Chosen:** `dbSeed` and `devWorkspaceCreate` `dependsOn("linkEnvFile")`, declared in `gradle/git-hooks.gradle.kts` beside `devStart`'s, applied after `dev-lifecycle.gradle.kts` registers both; pinned by a dry-run (`-m`) harness `scripts/test-env-link-wiring.sh`.
**Considered:** a buildSrc test — buildSrc cannot see the script plugins' task graph, and a nested `./gradlew` from it deadlocks (`buildSrc/build.gradle.kts` records both); asserting the `.kts` text — proves nothing about the graph; calling the link from `seedDatabase` — a side effect inside a helper three tasks share, where a declared dependency shows in `-m`.

### One bootRun argument builder

**ID:** boot-run-args-builder
**Status:** active
**Chosen:** `bootRunArgs(port, jdbcUrl, bucket, storageEndpoint, cacheIndex, extra)` in `buildSrc/src/main/kotlin/com/gymie/gradle/BootRunArgs.kt`, filled by `Project.workspaceBootRunArgs` in `dev-lifecycle.gradle.kts`; the seed run passes `--app.seed.enabled=true`, the invite and the fixed day as `extra`; `devBootRunArgs` requires `-PbootPort` and prints `--args=<value>` through `println`, read with `-q`.
**Considered:** `-PbootPort` defaulting to `0` — Spring then picks a random port the recipe cannot know; keeping the builder in the `.kts` — no unit test can reach it.

### The teardown and the repair

**ID:** before-workspace-teardown
**Status:** active
**Chosen:** the recipe's last sentence runs `./gradlew devWorkspaceRemove -PworkspaceId=<its id>` before `git worktree remove ../<name>-before` when a db or dev task ran there. The six orphans are removed by the fixer only after it re-derives, at run time, every live gymie and gymie-frontend worktree's id with `./gradlew devWorkspace -PworkspaceChange=<worktree name>` and finds none of them; an id a live worktree derives is left and reported.
**Considered:** sweeping every inventory line not derived by a live worktree automatically — a new destructive command for a one-off cleanup; leaving the orphans — a fix repairs what its defect wrote.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — Task 2 narrows when a protective hook scans a heredoc body, and a missed shape there is a main-checkout write let through; that judgement wants depth.
**Considered:** `flow-low` — enough for the four mechanical patches, not for the hook's widened allow.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, one fixer — key `self-review-kan-1001-be-cardio-workouts-count-on-the-profile-s-12-fix`, Tasks 1–6 in order; Tasks 1–3 in the agents worktree, Tasks 4–6 in the gymie worktree, Task 6 carrying the repair.
**Considered:** a second fixer for the gymie tasks — one more dispatch for three pre-applied patches and a scripted repair.

## Open questions

## Live check

- **flowd (S1)**, once landed and flowd restarted: from the gymie worktree, `flow self-review bundle
  -change kan-1001-be-cardio-workouts-count-on-the-profile-s-12` prints a
  `## spectre/changes/archive/kan-1001-be-cardio-workouts-count-on-the-profile-s-12/live-verification.md`
  section and its summary line reads `of 9 sources`. Not working looks like `of 8 sources` or a
  `skipped: …/live-verification.md (absent)` line for this change, whose archive carries the file.
- **Workspaces (F3)**: before, `./gradlew devWorkspaceInventory` lists seven workspaces; after Task 6's
  repair it lists `kan-1001-be-5a0d` and nothing else that existed before. Not working looks like any
  of the six orphan ids still listed.
- **Hook (F2)**: from the agents main checkout, the reproduction command above is allowed and
  `cat > <wt>/x.sh <<'EOF'` / `rm <main>/f.txt` / `EOF` / `/elsewhere/x.sh` still denied.
