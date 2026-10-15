## Context

- Two non-`big` findings of the kan-982 self-review pass, planned as one change per item 1 of
  **3. Fix every finding that is not `big`** (`skills/flow-self-review/SKILL.md`), each carried by
  exactly one task. Both were dropped by the first planner (archived
  `self-review-kan-982-prod-fixes-group-add-person-tap-set-value-bleed/design.md`, **Context**
  items 6 and 7); the operator chose to fix them now, and each decision below answers the
  objection recorded there.
- The agents project's `.flow/project.md` binds both: new guard logic is Go in
  `stats/internal/guard/`, registered in `guard.Registry`, with Go table tests, behind a thin
  `flow_guard_exec` shim.
- Finding 6 (Task 1). Run 1's reshape (`reshape-branch.sh`) squashes every commit of the branch
  into one implementation commit, so the gymie history the finding must be measured on survives
  only as unreachable objects. Reconstructed pre-squash snapshots (the newest unreachable tip per
  change whose chain carries a `Task-Id:` commit) hold 1635 commits, of which a guard listing every
  non-chore, trailerless commit would list 236.
  <!-- measured: python3 -I /tmp/spk2/snap.py <repo> over git fsck --unreachable --no-reflogs of gymie, gymie-frontend and gymie-playwright (machine-local spike script, not re-runnable from the tree) @ each repository's object store on 2026-10-15 -->
  Most of them are the pipeline's own sanctioned commits — `fix(<module>): review Minors`, panel
  rounds, full-suite fixes, re-recorded captures — which carry no mark a guard could read. Dispatch
  rows name the sha of only 5 of 31 such gymie-backend commits.
  <!-- measured: flow record dispatches -change <name> for each reconstructed gymie change, sha prefix match (machine-local spike script /tmp/spk2/rows.py, not re-runnable from the tree) @ the flowd store on 2026-10-15 -->
  On kan-982's own newest frontend snapshot, the guard this change adds lists 1 of 25 commits,
  `1a97b2624334 test(update): drop checks on tags the silent reload removed` — the F3 commit
  touching `QueueUpdateGateUiTest.kt` — and no other.
  <!-- measured: scripts/check-untasked-commits.sh <shared clone of gymie-frontend at 14250ef9> $(git merge-base HEAD origin/develop) with Task 1 applied in a scratch clone @ 8650ac1f (exit 1, one commit listed) -->
- Finding 7 (Task 2). `tcfCheckCommitScope` (`stats/internal/guard/taskcommitfields.go`) refuses
  only the three shapes `rules/commit-scope-is-the-module.mdc` names. Candidate refusals were
  measured over every scoped, non-merge commit of the three gymie repositories, the prohibited
  shapes excluded: a scope matching no token of the touched paths, 161 of 1494; that, or a scope
  an earlier commit used, or a word of the repository name, 31 of 1494; the design below, 10 of
  1494 — gymie 3 of 1137, gymie-frontend 5 of 228, gymie-playwright 2 of 129 — each a first use of
  an area scope (`build`, `ci`, `spectre`, `auth-ui`, `style`, `workout-editor`, `shell`, `specs`).
  <!-- measured: python3 -I /tmp/spk2/scope.py and /tmp/spk2/r4s.py over git log --all --no-merges of the three repositories, sibling trees at HEAD (machine-local spike scripts, not re-runnable from the tree) @ each repository's HEAD on 2026-10-15 -->
  The same design refuses F16's `fix(group)` at gymie-playwright `7479ba1`, both in the spike and
  in Task 2's Go function run against the real repositories.
  <!-- measured: tcfCheckScopeSibling over gymie-playwright 7479ba1 with gymie's .flow/project.md, in a throwaway Go test in a scratch clone with Task 2 applied @ 8650ac1f (refused; fidelity, visual and e2e pass) -->
  The population is every scoped commit, not only task commits, and the sibling trees are read at
  HEAD rather than at each commit's date, so the rate is an approximation of what the guard sees.
- Every task's diff was applied in plan order to a scratch clone at `8650ac1f`: `gofmt -l` and
  `go vet ./internal/guard` are clean, `go test ./internal/guard/` and
  `scripts/run-guard-tests.sh` pass, each new Go test fails with its production change reverted,
  and `check-verbatim-moves.sh` (with this change's `verbatim-moves.txt`),
  `check-references.sh`, `check-markdown-integrity.py`, `check-dispatch-paragraphs.sh`,
  `check-installed-citations.sh`, `check-guard-symlinks.sh`, `check-stage-mark-calls.sh` and an
  unchanged `check-normative-inventory.sh` output hold.

## Decisions

### Sanctioned untasked commits carry a Flow-Key trailer

**ID:** flow-key-trailer
**Status:** active
**Chosen:** every route that commits outside a task writes `Flow-Key: <key>` — the dispatch key
its round is recorded under (`full-suite-fix-<n>`, `panel-fix-<round>`,
`task-<n+n>-implementer-fix-<k>`), or the stage key where the stage records no such row
(`flow.visual-verify`, `flow.brainstorm`). The guard checks the trailer is present and non-empty;
it does not look the key up in the store. The rule is canonical in **Commits outside every task**
(`skills/flow-contracts/git-boundaries.md`); the panel-fix route's commit line carries it, which
the full-suite fix and the gated reviewer's on-top fix inherit by committing "on the route the
branch's push state dictates", and the visual-verify capture and brainstorm handoff-asset commits
carry their stage keys.
**Considered:** exempting a commit a dispatch row names by sha — rows name 5 of 31 flagged
gymie-backend commits, and a round-boundary rebase or a fold rewrites the sha a row recorded;
checking the key against the store — a store outage would then decide the gate, and the panel
already reads the records; a fixed route vocabulary in the trailer — the dispatch key already
names the route and is what a reader traces.

### Planning-only and chore commits are exempt by derivation

**ID:** untasked-derived-exemptions
**Status:** active
**Chosen:** a `chore` subject (the plan, link, merge-order, narrative and archive commits of
**Planning commits**) and a commit touching only `<spec-root>/changes/` (a `docs(spectre)` plan
amendment) are never listed; merges are never read.
**Considered:** listing plan amendments too — they carry nothing the reshape loses, since run 1
commits the whole planning delta on its own; exempting `fixup!` subjects — a fixup still standing
at run 1 is a fold that never ran, worth listing.

### The untasked check runs at run 1's unfinished-work gate

**ID:** untasked-at-integrate-gate
**Status:** active
**Chosen:** `scripts/integrate-change.sh prepare` runs `check-untasked-commits.sh` per worktree
after `check-visual-verify-dispatched.sh`, `<base>` the fork point `git merge-base HEAD
origin/<base>`, or the archived re-run's base (`arch_base`) where the change is archived; exit 1
sets the same `ask` the other two gate guards set, and exit 2 is `STOP: no-verdict`. The finish
contract states the call and the verdicts.
**Considered:** the handoff (`flow.verify`) — a fix run after it adds commits it never sees, and
the gate is the last point before the reshape erases them; both places — two call sites for one
listing the gate already offers three courses for.

### integrate-change.sh gains a call, not guard logic

**ID:** integrate-call-site-in-bash
**Status:** active
**Chosen:** the call and its verdict `case` land in `scripts/integrate-change.sh`, the bash
orchestrator that already sequences the gate's Go guards (`check-unfinished-work`,
`check-visual-verify-dispatched`); every decision about a commit lives in the Go guard.
**Considered:** porting `integrate-change.sh` to Go first — it is an orchestrator, not a guard in
`guard.Registry`, and the project's port-before-extend rule binds guards.

### A scope naming only another repository's module is refused

**ID:** scope-sibling-module
**Status:** active
**Chosen:** `check-task-commit-fields.sh` refuses a declared scope that is a path token of another
`## apps` repository's tree and none of: a token of a path the commit touched; a word of the
commit's repository name (its main checkout's basename, split at `-`, `_`, `.`); a scope an
earlier commit of that repository used; a whole directory name of that repository at the commit,
or every `-` part of it one. Tokens are path segments, their extension-free stems, their
`-`/`_`/`.` parts and their camelCase words, lowercased. The `## apps` table is read from the
canonical worktree's `.flow/project.md` (the worktree's own when none is given); a repository git
cannot read refuses nothing, and a project with one repository is never refused. Measured: 10 of
1494 scoped gymie commits refused, F16 refused (**Context** above).
**Considered:** a scope matching no touched path (161 of 1494) — the prior objection itself; that
plus earlier scopes and the repository name (31 of 1494); file-name tokens of the commit's own
tree as passing — gymie-playwright's archived `kan-30-group-workouts-…` change directory then
passes `group` and F16 slips through; a project-declared scope vocabulary — the rule states no
vocabulary is maintained, and a declared list goes stale on every new area.

### The commit-scope rule names the fourth shape

**ID:** scope-rule-fourth-shape
**Status:** active
**Chosen:** `rules/commit-scope-is-the-module.mdc` lists "A module of another repository" under
**The four prohibited shapes**, with F16 as its example and `test(visual)` as a passing one; its
core excerpt is unchanged, so no managed block re-renders.
**Considered:** leaving the rule at three shapes — the guard would refuse what the rule says is
allowed.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `medium` — Task 1 adds a gate verdict integrate stops on and a trailer contract four
commit routes now owe, and Task 2 adds a refusal to the task-close guard every run passes; a
low-effort read could pass over a route the trailer misses or a scope the refusal misfires on.
**Considered:** `low` — both tasks ship with spiked diffs and passing tests, but both change what
stops a run.

### Fixer grouping

**ID:** fixer-groups
**Status:** active
**Chosen:** two groups, each one fixer, in plan order — g1: Task 1 (the untasked guard, its shim
and symlink, integrate, its test, and the five prose files); g2: Task 2 (the scope check and the
rule).
**Considered:** one group — no file is shared, and Task 1 alone is the larger diff; Task 2 first —
nothing orders them, plan order is the default.

## Open questions

## Live check

None — both changes are guards over git history and plan text with no running service or store
state: Task 1's guard reads only git, and its integrate wiring is exercised end to end by
`scripts/test-integrate-change.sh`'s `untasked` case; Task 2's check reads git and
`.flow/project.md`, and its Go test builds two real repositories.
