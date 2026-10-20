## Context

Found by the self-review pass of `/flow` change `kan-999-group-session-workout-names` (gymie);
the agents fixes ride branch `self-review-kan-999-group-session-workout-names`, based at
`origin/main` `c53570af`, and the gymie fix rides that change's own branch
`spectre/kan-999-group-session-workout-names` in
`/Users/tweety53/Projects/gymie-worktrees/kan-999-group-session-workout-names`, landing with its
merge and push. Every patch under `patches/` beside this file was applied, tested and reverted at
plan time.

**Finding A1-1 — the both-trees note on a frozen openspec tree (flow-fix, blast radius 3).**
<!-- measured: Task 1's **Files:** names 4, of which 3 are the fix and its two test files @ c53570af -->
`ls -A ~/Projects/gymie/openspec/changes` prints `archive` alone, so every `spec_root_leaf` call
on a gymie worktree printed the note. With `patches/task-1.diff` applied, `spec_root_leaf` on the
kan-999 gymie worktree prints `spectre` and nothing on stderr; `TestSpecRoot` and
`scripts/test-lib-change-plan.sh` case 17 fail with the fix reverted and pass with it. The third
probe, `specRootLeaf` in `stats/cmd/flow/tasks.go`, carries no note at all, so it is untouched. No
bash harness pinned the note's stderr before this change.

**Finding A1-3 — a grown type past its size rule (flow-fix, blast radius 1).** kan-999's
`narrative.md` records detekt `TooManyFunctions` on three types the plan grew under a "no new port"
constraint. `grep -n TooManyFunctions skills/flow/brainstorm-planner.md` finds nothing at
`c53570af`. No guard pins section D's text; `check-verbatim-moves.sh` reports the one new sentence,
listed in `verbatim-moves.txt`.

**Finding A6 — ledgers never show a dispatch's end (flow-speed, blast radius 2).**
`RenderLedger` prints `- Started:` and no line for `EndedAt`, which `RunRecord` already scans
(`stats/internal/store/records.go`). With `patches/task-3.diff` applied, `go test
./internal/records/ ./cmd/flow/ ./internal/harvest/` passes, and the new test fails with
`render.go` reverted.

**Finding A1-2 + A4-1 — live-check fixtures and the before stack (gymie dev tooling, blast radius
1).** kan-999's `narrative.md` and `live-verification.md`: hand-seeded rows met
`group_session_participant_active_user` and `user_slug_lower_unique`, one `psql -c` rolled back
silently, the cleanup removed the ten seeded sessions, and the before figure came from the merge
base's `:app:bootRun` in a detached worktree against the same database. The recipe's arguments are
the ones `seedDatabase` (`gradle/dev-lifecycle.gradle.kts`) passes its own `:app:bootRun`; the
placeholders' names are the variables `./gradlew devWorkspace` prints in that worktree
(`DB_URL=jdbc:postgresql://localhost:5432/gymie_kan_999_5287`, `REDIS_DATABASE`, `STORAGE_*`). The
backend authenticates a bearer token itself (`JwtAuthFilter`), so the before backend is called on
its own port, as kan-954's live check did.

## Decisions

### The note's condition

**ID:** note-only-for-open-openspec-change
**Status:** active
**Chosen:** both trees present still returns `spectre`; the note prints only when `openspec/changes/` holds an entry other than `archive/` — an empty `openspec/changes/` prints nothing either, since it hides no change.
**Considered:** dropping the note entirely — a real half-migration with an open `openspec/` change would again be silently invisible to every guard; probing for a `tasks.md` under `openspec/changes/*/` — a change mid-scaffold or a stray file is still worth naming, and the entry test is the one the finding states.

### Bash and Go change together

**ID:** spec-root-both-sides
**Status:** active
**Chosen:** narrow the existing branch in `scripts/lib/spec-root.sh` and in `specroot.go` identically, each with its own test (`test-lib-change-plan.sh` case 17, `TestSpecRoot` cases).
**Considered:** porting the bash callers onto the Go probe first, per the guards-in-Go rule — five bash callers (`check-plan-shape.sh`, `check-task-build-green.sh`, `plan-dispatch-bundles.sh`, `integrate-change.sh`, `lib/change-plan.sh`) for a one-condition narrowing of an existing message; the rule binds new guard logic, and this adds no verdict.

### The size-rule sentence's place

**ID:** size-rule-in-section-d
**Status:** active
**Chosen:** one blockquote sentence in section D of `skills/flow/brainstorm-planner.md`, right after the verify-step rule, naming detekt's `TooManyFunctions` as the example and kan-999 as the origin.
**Considered:** a gymie-only line in its `.flow/project.md` — the slip is the planner's in any project with a size rule; a plan guard counting members — a guard would need each project's lint thresholds and the type's member count, for a one-off slip.

### The Ended line's shape

**ID:** ledger-ended-line
**Status:** active
**Chosen:** `- Ended: <RFC3339 UTC> (<duration rounded to the second>)` directly after `- Started:`, only when `EndedAt` is non-nil and `StartedAt` is set; an open dispatch renders no Ended line.
**Considered:** `- Ended: not recorded` on an open dispatch — noise on every running row, the shape the Diff base line already rejects; the duration alone — a reader correlating with logs needs the timestamp.

### Where the gymie paragraphs sit

**ID:** gymie-live-check-paragraphs
**Status:** active
**Chosen:** two prose paragraphs in gymie's `## run`, right after the `count-request-queries.sh` paragraph, the before-stack command written inline rather than in a fenced block.
**Considered:** a fenced `bash` block — `## run`'s fences are what a stage starting the stack reads, and a `git worktree add` there could be run as a start step; a script — one run used the recipe, and its values are per worktree.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-low` — four small, pre-applied patches, each with its test or prose guard; the only logic is one directory-entry check and one optional render line.
**Considered:** `flow-medium` — buys depth no task needs; nothing here touches concurrency, the store's queries or a gate's verdict.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, one fixer — key `self-review-kan-999-group-session-workout-names-fix`, Tasks 1, 2, 3 and 4 in order; Task 4 in the gymie worktree.
**Considered:** a second fixer for the gymie task — one more dispatch for a one-paragraph patch applied with `git apply`.

## Open questions

## Live check

Finding A6 changes what `flow record render` writes. After this lands and `flowd` is rebuilt by the
operator's usual route, render any change's ledger (`flow record render -kind ledger -change <change> -repo <its worktree>`):
every closed dispatch shows `- Ended:` with a duration matching its end minus its start. Not working
looks like a ledger with `- Started:` lines and no `- Ended:` line at all. Findings A1-1, A1-3 and
A1-2 + A4-1 have no runtime: a guard's stderr pinned by its tests, and prose.
