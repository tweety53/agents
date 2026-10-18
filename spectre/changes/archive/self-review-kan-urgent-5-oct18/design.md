## Context

- Five Jira tickets planned as self-review findings per item 1 of **3. Fix every finding that is not
  `big`** (`skills/flow-self-review/SKILL.md`), on `self-review-kan-urgent-5-oct18` based at agents
  `b61a1e6a`. Each fixable finding is carried by exactly one task; the brainstorm design is
  `.superpowers/sdd/2026-10-18-self-review-kan-urgent-5-oct18-design.md` in this worktree. Two
  tasks reach outside this repository, each committed on branch `self-review-kan-urgent-5-oct18`
  in a worktree of its own repository off that repository's default branch (**Pipeline defects
  found mid-run**, `skills/flow-contracts/pipeline.md`): spectre at
  `/Users/tweety53/Projects/spectre-worktrees/self-review-kan-urgent-5-oct18` off `origin/main`
  (`c3b62de`), gymie at `/Users/tweety53/Projects/gymie-worktrees/self-review-kan-urgent-5-oct18`
  off `origin/develop` (`6a8eddc7`). Both worktrees exist, clean.
  - KAN-975 (Task 1), blast radius 4: reproduced at spectre `c3b62de` — a counterpart change only
    under `<peer>-worktrees/x/spectre` makes `spectre validate` report `one-sided link: x names
    can:x, but can:x does not link back to it`, and `spectre link` refuse `canonical change
    "kan-9-cross" does not exist in peer "can"`. Both route through `tree.CounterpartWorktree`
    (`internal/tree/peer.go`), its two callers `check.resolveCounterpart` and `cmd.Link`, which
    builds `<repo>/.worktrees/<id>/spectre` only. The radius is `peer.go`, `cmd/link.go` (a
    comment naming the one layout) and the two test files; no file in this repository cites the
    function, and none documents the hand-made `.worktrees/<name>` symlink workaround.
  - KAN-965, link half (Task 1): the peer resolves to its main checkout by design (`repoParent`,
    `internal/tree/tree.go`) and the counterpart is then sought through `CounterpartWorktree` —
    the one-sided link from a sibling-layout worktree is KAN-975's defect and Task 1's fix. No
    blast radius of its own.
  - KAN-965, merge-order half: disproved, already fixed. spectre `849e8e1` (2026-10-12) accepts a
    declared peer in a link's `## Merge order`, `c3b62de` documents a canonical link holding
    `## Merge order` alone; at `c3b62de` `go test ./internal/check/ -run
    'TestLinkFindingsQuiet/merge_order_names_a_declared_peer_that_is_not_a_part'` passes on a
    merge-order-only `link.md`, and the installed `~/go/bin/spectre` is built from `c3b62de`
    (`go version -m`). No task.
  - KAN-967 (Task 2), blast radius 10: traced in the dev store (`flow-postgres`, read-only
    queries). The ticket's KAN-924 run predates five attribution fixes (`4dcc2bc5` … `96f8e077`,
    2026-10-10 to 2026-10-13); the rows recorded since 2026-10-17 still leave two causes, every
    other row measured:
    - Fix-loop dispatches (`self-review-*`, `pipeline-fix-*` keys) are all unmeasured with no
      agent id. Their begin is recorded by `check-loop-dispatch.sh` inside the PreToolUse hook
      (`hooks/enforce-loop-dispatch.py`), so no `flow record dispatch begin` command reaches the
      transcript and `stampDispatchAgents` (`stats/internal/harvest/watcher.go`) has no begin to
      pair the launch with; their session token, the loop's name, matches no stage run, so the
      window fallback finds nothing either. The launch itself is in the transcript, an
      `async_launched` tool result whose `prompt` carries the `Key:` line (confirmed on
      `self-review-kan-urgent-5`'s five launches).
    - Inline rows of a run that marked no stage under its token (`kan-941`'s `mf-k941r3`, its
      task-8 to task-11 implementers) are unmeasured: `DispatchWindowsForSession`
      (`stats/internal/store/records.go`) reaches a dispatch only through a stage run's
      `session_token`.
    The radius is the ten files Task 2's diff touches; the names it adds are new and no file
    cites them.
  - KAN-973 (Task 3), blast radius 2: `rules/fix-determinism-at-the-source.mdc` already requires
    reproducing before touching, naming what varied, and the mechanism-removed failing run;
    `skills/flow/implement.md`'s PROOF RUNS paragraph sizes runs to the measured rate and
    requires the report to state it; MUTATION PROOF and the fix round's mutation-proof already
    make mutation proofs the default. Missing: how to make a rare flake fail (throttling) and
    measure its rate under that condition, and any reviewer prompt asking a flake fix for its
    measured proof. The rule's edited section is outside its `<!-- core -->` excerpt, so no
    rendered copy changes; no section or name is renamed, so no citer is in the radius.
  - KAN-962 (Task 4), blast radius 5: reproduced at gymie `6a8eddc7` — `./gradlew
    verifyFeatureCatalog` in the gymie worktree fails `Orphan @Feature annotations not in
    catalog`; the frontend's `shared/src/commonTest` and `desktopTest` carry codes of domains the
    catalog never indexed (`workout-ui`, `ui-design-system`, … — OpenSpec capability names), and
    25 `auth-ui` codes the catalog lacks. With the scan narrowed and those back-filled, two stale
    entries the orphan failure hid surface as `Implemented catalog entries without @Feature`:
    `auth-ui.navigation.open-register` (its annotation left with gymie-frontend `144c15c7`,
    invite-only registration) and `baseline.schema.auth-tables-only` (its annotation left with
    gymie `24457ac6`, the workout module). The radius is the gradle script, `catalog.yaml` and
    the three READMEs listing codes.
- Spiked at plan time, every edit reverted: each task's diff applies in its worktree (`git
  apply`); with it applied the task's targeted tests pass, each new test fails with its fix
  reverted, and `go vet` and `gofmt -l` are clean. Task 4's diff plus a stub back-fill made
  `./gradlew verifyFeatureCatalog` print `Feature catalog OK`.

## Decisions

### Resolve the sibling worktree layout beside the in-repo one

**ID:** counterpart-sibling-layout
**Status:** active
**Chosen:** `CounterpartWorktree` tries `<repo>/.worktrees/<id>/spectre`, then
`<repo>-worktrees/<id>/spectre`, returning the first that exists and opens.
**Considered:** the sibling layout alone — a repository still on the retired in-repo layout would
lose the fallback it has today; resolving the counterpart through `git worktree list` — a git call
per lookup, and plain-filesystem trees (the package's tests) have no git.

### Stamp a fix-loop launch by its prompt's Key line

**ID:** loop-launch-stamp-by-key
**Status:** active
**Chosen:** `ParseAgentLaunches` reads the launch result's `prompt` and keeps the
`Key: pipeline-fix-…` or `Key: self-review-…` value as `AgentLaunch.LoopKey`; the watcher stamps
such a launch with the new `StampLoopDispatchAgent` — the latest row under that key, empty
`agent_id` only — then credits the agent file as a paired stamp does, and keeps loop launches out
of the begin pairing.
**Considered:** recording the agent id at `pipeline-fix.sh end` — the parent would type it, and an
id stamped after the agent ran loses what was harvested before; deriving the loop's session token
from the key to reuse `StampDispatchAgent` — a `pipeline-fix-<k>` key does not carry the change
name its token holds; a PostToolUse hook stamping the id — a second hook and a new CLI verb for
what the transcript already records.

### Record the begin's session on the dispatch row

**ID:** dispatch-session-id
**Status:** active
**Chosen:** `flow record dispatch begin` sends `CLAUDE_CODE_SESSION_ID` (else `MUSE_SESSION_ID`,
`sessionIDFromEnv`) as `sessionId`; migration `0040` adds `dispatches.session_id`;
`DispatchWindowsForSession` matches it beside the stage-run token join.
**Considered:** binding a dispatch token from the transcript as stage-run tokens are bound — a new
binding table and watcher pass for a fact the CLI already holds; marking a stage in every fix run
— a contract change across `/flow` that would leave every older run unmeasured.

### Throttle a rare flake until it fails; the primary reviewer asks for the proof

**ID:** flake-proof-throttle-and-review
**Status:** active
**Chosen:** one paragraph in the rule's **Read the flake before touching anything** (throttle,
then measure the rate under that condition, which sizes the proof runs); one paragraph in the
primary reviewer prompt's **What to Check** asking a flake fix for its measured rate, run count
and mechanism-removed run, citing the rule.
**Considered:** restating the method in `implement.md`'s PROOF RUNS — it already cites the rule,
and a second statement is the repetition this repository refuses; the test-sanity slot — rolled
per change, so a flake fix could pass review unasked.

### Narrow the frontend scan to the catalog's domains and back-fill auth-ui

**ID:** feature-catalog-frontend-scope
**Status:** active
**Chosen:** frontend `@Feature` codes are checked only where their domain is one the catalog
holds (derived from its entries); the catalog gains the frontend's uncatalogued `auth-ui` codes
and drops the two entries whose annotations are gone, the READMEs listing codes following.
**Considered:** back-filling every frontend code — over a thousand entries in domains the catalog
never documented, and every frontend change would then need a gymie commit; dropping the
frontend scan — the catalog's `auth-ui` entries would lose their coverage check.

### Land spectre by pull request

**ID:** spectre-landing-route
**Status:** active
**Chosen:** spectre declares no `## default landing route` (it has no `.flow/project.md`), so its
branch lands by the flow rule's default for that case, a pull request; gymie lands by its declared
merge and push, this repository by its own.
**Considered:** merge and push for spectre — its recent history is direct pushes, but no
declaration says so.

### Files fields stay relative to the task's own worktree

**ID:** task-files-worktree-relative
**Status:** active
**Chosen:** each task names its repository's worktree in its steps and in Global constraints, and
its `**Files:**` paths are relative to that worktree's root.
**Considered:** absolute or repo-prefixed paths in `**Files:**` — the field is matched against
each commit's repository-relative changed paths, so `check-task-commit-fields.sh` would refuse
every task.

### First review effort and fixer grouping

**ID:** first-review-effort
**Status:** active
**Chosen:** first review on `flow-medium` (effort `medium`); one fixer dispatch (key
`self-review-kan-urgent-5-oct18-fix`) carrying Tasks 1–4 in order — three are verified diffs
applied with `git apply`, and Task 4 adds a back-fill the fixer writes from the spec.
**Considered:** `flow-low` for the first review — Task 2 changes which rows the harvester credits
and adds a migration to the store every run writes through; one fixer per task — four dispatches
where three are `git apply`s.

## Open questions

## Live check

Against the dev workspace's `flowd`, after the landing restarts it (migration `0040` applied):
run a fix loop's dispatch and an inline row under a token no stage run carries, then read
`flow record dispatches -change <loop change>` — before: every loop row has no `agentId` and empty
`metrics`; after: each loop row carries its launch's `agentId` and a `tokens` bag once the agent
file is harvested, and the inline row carries `tokens.main`. The change not working looks like
today: `flow record cost-status` still counting every loop row `not measured`. spectre and gymie
hold no runtime state: `spectre validate` on a sibling-layout cross-repo change and
`./gradlew verifyFeatureCatalog` in a gymie checkout are their checks, run by the tasks. flow
reaches Task 1's fix only through the installed `~/go/bin/spectre`, rebuilt with
`go install ./cmd/spectre` from spectre's `main` once its pull request merges; until then
`go version -m ~/go/bin/spectre` still names `c3b62de`.
