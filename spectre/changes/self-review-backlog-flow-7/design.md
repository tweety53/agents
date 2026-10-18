# self-review-backlog-flow-7

## Context

- Five Backlog Flow findings — the next five live ones, oldest first: KAN-771, KAN-790, KAN-801,
  KAN-804, KAN-815 — planned as one change per item 1 of **3. Fix every finding that is not
  `big`** (`skills/flow-self-review/SKILL.md`). Brainstorm design:
  `.superpowers/sdd/2026-10-19-self-review-backlog-flow-7-design.md` in this worktree. Every
  question was auto-resolved to its recommended option
  (`skills/flow-contracts/operator-prompts-auto-resolution.md`).
- **Not in this change.** The five oldest Backlog Flow tickets — KAN-702, KAN-733, KAN-740,
  KAN-802, KAN-803 — were fixed by self-review-backlog-flow-6 and moved to Done. KAN-764 is a
  design ask, a frame only the operator can draw, and stays in Backlog Flow. KAN-783 (`a1f0a853`,
  `-tokens`), KAN-793 (`a2b0901f`, `flow record pass`), KAN-810 (`3dfef567`,
  `check-dev-stack-fresh.sh`) and KAN-814 (`489ce395`, verdict rows) were already fixed and moved
  to Done.
- **KAN-771 — Task 1.** `selfreview.Bundle` (`stats/internal/selfreview/git.go`) prints
  `skipped: <label> (absent)` for every source it could not find, and nothing in the bundle names
  the daemon that served it; `stats/cmd/flowd/main.go` reads no build identity. The live binary
  already carries the stamp `go build` writes. Blast radius: six files under `stats/`.
  <!-- measured: go version -m /Users/tweety53/Projects/agents/stats/bin/flowd | grep vcs @ main 02a3d012 → vcs=git, vcs.revision=02a3d012e12a9b66dc9296d1fbafb249d98ac98e, vcs.time=2026-10-18T20:46:11Z, vcs.modified=false -->
- **KAN-790 — Task 2.** The PIXEL PROBE paragraph rides only the review panel's fix dispatch
  (`skills/flow/review-panel.md`); **The loop** (`skills/flow/verify-fix-loop.md`) step 2 gives
  the in-run fix implementer MUTATION PROOF alone, and no visual-verify file asks for a probe per
  fixed departure. Source: gymie `docs/self-review/kan-577-…` (`theRightsControlMatchesNavP2`).
  Blast radius: one file.
- **KAN-801 — Task 3.** No skill classifies departures whose frames disagree; the verifier's
  shared-component diff (`skills/flow/visual-verify-verifier.md`, state 1) finds the drift, and
  nothing decides it. Source: gymie `docs/self-review/kan-741-…`. Blast radius: one file.
- **KAN-804 — Task 4.** gymie's kan-741 records the admin-frontend worktree as a bare worktree
  "not in the state record, because it carries no commits" (gymie `origin/develop`,
  `spectre/changes/archive/kan-741-full-visual-verification-of-current/narrative.md`). Since
  `6aa7aa1e` (2026-10-17) every worktree a later stage needs runs `kickoff-worktree.sh --from`,
  which persists it through `flow state add-worktree` with `git rev-parse HEAD` as its merge base
  (the script's header, step 2) — a commit-less worktree included, and nothing prunes a map entry
  for want of commits. The recipe's triggers still name a fix round's task and the visual
  stage's commit destination only, not a stack that must serve from a checkout of its own.
  Blast radius: one file.
- **KAN-815 — Task 5.** The ticket says run 2 archives; the archive moved to run 1, which
  `integrate-change.sh commit` runs in the canonical worktree only (`spectre archive`, then
  `commit-archive.sh "$CANON"`). A satellite's `spectre/changes/<name>/link.md` stays in place and
  `check-cleanup-complete.sh`'s change-directory row reports it `LEFTOVER`; gymie-frontend's main
  checkout holds ten such link-only directories today. `spectre archive` on a link-only satellite
  directory exits 0 with no `--force`. The resulting commit is a `chore` subject, which
  `check-untasked-commits.sh` exempts. Blast radius: three files.
  <!-- measured: ls /Users/tweety53/Projects/gymie-frontend/spectre/changes | grep -vc '^archive$' @ gymie-frontend 55634cf3 → 10, each directory holding link.md alone -->
  <!-- measured: spectre archive kan-748-menu-glyph-more-visible-round-dots-24dp-full-ink in a git copy of gymie-frontend's spectre/ tree → "archived …", exit 0, link.md renamed under archive/ -->

## Decisions

### The daemon's VCS stamp is the bundle's first note

**ID:** bundle-build-note
**Status:** active
**Chosen:** `flowd`'s `main` reads `debug.ReadBuildInfo()` (`vcs.revision` cut to 12, `vcs.time`, `vcs.modified`) through `formatBuild`, passes it as `api.WithBuild`, and `selfreview.Bundle` takes it as a `build` parameter rendered as the first `note:` line; an empty build renders none — the reader sees which build answered beside every skipped line.
**Considered:** `-ldflags -X` stamping — `go build` already writes the VCS stamp, a second mechanism adds nothing; reading the build info inside `Bundle` — a test binary carries no stamp, so the bundle would vary by how its binary was built; serving the build on a separate `/version` route — the bundle reader would need a second call to learn what the bundle could say.

### The in-run fix lands a probe per fixed departure

**ID:** departure-probe
**Status:** active
**Chosen:** **The loop**'s step 2 gives the implementer the PIXEL PROBE paragraph beside MUTATION PROOF, and one sentence requires every mockup departure the round fixes to land a permanent probe — the real composable, never a hand-copied constant — in the task's one commit.
**Considered:** a guard matching probe tests to fixed departures — which test pins which measured value is a judgement, not a pattern; a `dpSites` row for `verify-fix-loop.md` — the file cites the paragraph, it does not carry it.

### A shared component's own design source wins

**ID:** mockup-conflicts-own-source-wins
**Status:** active
**Chosen:** a paragraph at the top of **The loop**: agreeing frames → fix; frames disagreeing on a shared component → the component's own design source wins, the other frame's departure closes as a mockup conflict on the handoff's `**Decisions:**` lines, never fixed and never asked.
**Considered:** the rule in the verifier prompt — the verifier reports departures, the parent decides what a round fixes; asking per conflict — the operator's kan-741 answer was the rule itself.

### A commit-less worktree rides the existing map

**ID:** commit-less-worktree-in-map
**Status:** active
**Chosen:** name "a stack that must serve from a checkout of its own" among the needs that run `kickoff-worktree.sh --from`, and state a worktree that will carry no commit is recorded with its HEAD as its merge base and cleaned like any other.
**Considered:** a distinct marker or list in the state record, through `flow state set`, the store validators and run 2 — the map's merge-base value already carries a commit-less worktree, run 2's per-repository loop and `remove-change-worktrees.sh` already reach it, and a second shape would teach every reader twice; no task — the recipe's trigger list leaves the serving case to inference, which is how kan-741 improvised a bare `git worktree add`.

### Satellites archive in run 1, after the canonical worktree

**ID:** satellite-archive-run1
**Status:** active
**Chosen:** `integrate-change.sh commit`, after the canonical archive commit and before the pushes, runs `spectre archive <name>` and `commit-archive.sh` in every other worktree whose `<spec>/changes/<name>/` still exists — on an archived re-run too; run 1's **Archive on the change branch** says so and narrows the two sentences that assumed a satellite holds no archive commit.
**Considered:** archiving in run 2 — run 2 runs after the merge, so its archive would be a second landing; gating on the canonical archived decision — a stopped run that archived the canonical side would leave the satellite behind forever; a new Go guard — this is an orchestration step in an existing script, not a check, and the guards it calls are already Go.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** flow-medium — the change moves Go the daemon serves and the landing script every multi-repository run 1 goes through.
**Considered:** flow-low — Tasks 2–4 are prose, but Tasks 1 and 5 change runtime behaviour a low-effort pass could under-read.

### One fixer group

**ID:** fixer-groups
**Status:** active
**Chosen:** one group, Tasks 1–5 in order, one fixer dispatch (`self-review-backlog-flow-7-fix`): Tasks 2–4 are `git apply` of verified diffs, and Tasks 2 and 3 share `skills/flow/verify-fix-loop.md`.
**Considered:** one group per task — five dispatches for work one fixer finishes in series; a Go group and a prose group — saves no wall time, the prose tasks are minutes.

## Open questions

## Live check

`flowd` serves the self-review bundle, so once Task 1 is in a rebuilt daemon,
`flow self-review bundle -change self-review-backlog-flow-6` against it prints
`note: served by flowd build <rev12> (<time>)` as its first `note:` line, `<rev12>` the daemon
checkout's `git rev-parse --short=12 HEAD`. Not working looks like a bundle with no
`served by flowd` line. Restarting the protected `flowd` is the operator's act (`.flow/project.md`,
`## stop`); the check runs only on a daemon the operator restarted, and otherwise reports the line
unverified.
