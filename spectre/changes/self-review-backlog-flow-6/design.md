# self-review-backlog-flow-6

## Context

- Five Backlog Flow findings — the oldest five, KAN-702, KAN-714, KAN-715, KAN-718, KAN-719 —
  planned as one change per item 1 of **3. Fix every finding that is not `big`**
  (`skills/flow-self-review/SKILL.md`). Brainstorm design:
  `.superpowers/sdd/2026-10-18-self-review-backlog-flow-6-design.md` in this worktree. Every
  question was auto-resolved to its recommended option (`skills/flow-contracts/operator-prompts-auto-resolution.md`).
- **KAN-702 — Task 1.** **Visual verification** step 2 (`skills/flow/verify-and-handoff.md`)
  pipes `git -C <worktree> diff --name-only <merge-base>..HEAD` into `check-visual-trigger.sh
  --repo` for every satellite worktree, and nothing in the step says which `<merge-base>`.
  **Inline verify** in the same file already names "its merge base from this run's working notes";
  the visual step does not. KAN-659's half — the guard matching another worktree's diff — landed in
  `56260f9f`. Blast radius: one file.
- **KAN-714 — already fixed, no task.** Since `e75f19d0` (2026-09-24) every `spectre link` is
  followed by a link commit of "each `<project>/spectre/changes/<name>/link.md` the link wrote —
  the canonical worktree's and the satellite's" (`skills/flow-contracts/git-boundaries.md`,
  **Planning commits**); a satellite's `link.md` is tracked like every other.
- **KAN-715 — already fixed, no task.** Since `5b64e1a9` (2026-09-30) the panel's base sync is
  `sync-panel-base.sh`, which rebases a `MOVED`-with-no-overlap worktree itself, with no prompt
  (`skills/flow/review-panel.md`) — the wrapper entry point the ticket asked for; the classifier
  sees the script, never a bare `git rebase`.
- **KAN-718 — already fixed, no task.** `aca9b26a` (2026-09-24) and `c05173e9` (2026-09-28)
  narrowed `hooks/protect-main-checkout.py` to writes into a main checkout on its default branch,
  skipping unexpanded paths; read-only commands naming the main checkout or a `$VAR` path pass.
  The zsh word-splitting trap and the absolute/`-C` path rule are **The parent's shell
  discipline** (`skills/flow/implement.md`).
  <!-- measured: echo '{"tool_name":"Bash","tool_input":{"command":"<c>"},"cwd":"/Users/tweety53/Projects/agents"}' | python3 hooks/protect-main-checkout.py for c in 'ls $W/scripts', 'flow record pass -C /Users/tweety53/Projects/agents x', 'cat /Users/tweety53/Projects/agents/README.md | head', 'git -C $W status' @ cfb9b564 → exit 0, no deny, each -->
- **KAN-719 — already fixed, no task.** gymie-frontend `55634cf3` (2026-10-18) declares a
  `## worktree setup` line copying the main checkout's `local.properties` into a fresh worktree;
  `kickoff-worktree.sh` runs it at step 3.
  <!-- measured: git -C gymie-frontend show origin/develop:.flow/project.md | grep -A4 '## worktree setup' @ gymie-frontend 55634cf3 → the cp line -->

- **More tickets on the operator's ask.** The next eleven Backlog Flow tickets after KAN-719,
  oldest first, checked against the tree; the first four still live joined this change:
- **KAN-733 — Task 5.** gymie `origin/develop` @ `8d2bb89d`: `AccessServiceImpl.setLevel` is not
  `@Transactional` and gates on `FriendshipReadRepository.findPair`, a plain `SELECT`;
  `FriendshipRepositoryAdapter.unfriend` is `@Transactional` but reads the pair with a plain
  `SELECT` before deleting the grants, then the row. Under read committed an unfriend that deletes
  the grants before a concurrent `setLevel` commits its write leaves that write standing. The only
  `FriendshipReadRepository` implementation is the adapter. Blast radius: three main files and two
  tests in gymie's `profile` module.
- **KAN-740 — Task 2.** `git grep -n -i waive -- skills` finds nothing about a live step;
  **Live check** (`skills/flow/verify-and-handoff.md`) says what to run, not what a waived step
  becomes. Blast radius: one file.
- **KAN-802 — Task 3.** `skills/flow/visual-verify-verifier.md`'s clock rule leaves the clock real
  for every view the server dates from its own clock, while gymie-playwright's `api.ts` already
  sends `X-Gymie-Today` (kan-741). Blast radius: one file.
- **KAN-803 — Task 4.** **The sweep** (`skills/flow-contracts/known-bugs.md`) stages "only the PNGs
  the failing run named" and the fix loop commits "every PNG whose bytes changed"
  (`skills/flow/verify-fix-loop.md`); neither reads them first. Blast radius: two files.
- **Outdated, no task.** KAN-720: gymie `origin/develop`'s `AGENTS.md` and `CLAUDE.md` carry the
  same no-blink section. KAN-722: gymie-frontend's Save is `enabled = … && state.loadError == null`.
  KAN-732: gymie-frontend's `devManifest.js` rewrites the history-fallback shell, injecting the
  gateway port on deep links. KAN-736: `full_access_required` appears in no gymie source, only in
  archived records. KAN-764: a design ask — a frame only the operator can draw. KAN-771 and later
  were not reached.
  <!-- measured: git -C gymie diff of the two files' no-blink sections @ origin/develop → identical; git -C gymie-frontend grep -n loadError EditProfileScreen.kt @ origin/develop → line 777; git -C gymie grep full_access_required origin/develop -- src → nothing -->

## Decisions

### The satellite's merge base comes from the working notes

**ID:** visual-trigger-working-notes-base
**Status:** active
**Chosen:** one sentence under **Visual verification** step 2 (`skills/flow/verify-and-handoff.md`): `<merge-base>` in both trigger calls is that worktree's own merge base from this run's working notes, never `origin/<base>` or a hand-picked ref.
**Considered:** the state file's `worktrees` map — `skills/flow/review-panel.md` already rules it out as a source, a creating run may not have written it yet; a guard refusing a match set carrying another change's `kan-*` paths — a project-specific heuristic for a defect the right base removes by construction; resolving the base inside `check-visual-trigger.sh` — the guard reads stdin paths and stays agnostic of where they came from.

### Four tickets are already fixed

**ID:** four-already-fixed
**Status:** active
**Chosen:** no task for KAN-714, KAN-715, KAN-718 or KAN-719 — each fix is in the tree, cited under **Context** with its commit.
**Considered:** re-stating each fix in a new sentence — restates what the tree already says canonically.

### setLevel holds the friendship row lock

**ID:** setlevel-friendship-row-lock
**Status:** active
**Chosen:** a `lockAcceptedPair` read (`SELECT … FOR UPDATE`) on `FriendshipReadRepository`, called from a `@Transactional` `setLevel`, and `FOR UPDATE` on `unfriend`'s first read — both sides lock the friendship row first, so either order serialises.
**Considered:** `@Transactional` alone — a plain `SELECT` takes no lock under read committed, the race stands; `SERIALIZABLE` isolation — retries on serialization failure the callers do not handle; locking the `access_grant` row — no row exists to lock before a first grant.

### KAN-733 lands in gymie off develop

**ID:** kan-733-gymie-worktree
**Status:** active
**Chosen:** a worktree `/Users/tweety53/Projects/gymie-worktrees/self-review-backlog-flow-6`, branch `self-review-backlog-flow-6` off `origin/develop`, created by Task 5's first step; landed by gymie's `## default landing route`, `merge and push` onto `develop`.
**Considered:** leaving it for a gymie change — the pass fixes every non-`big` finding.

### A waived live step runs on a proxy

**ID:** waived-live-step-proxy
**Status:** active
**Chosen:** one sentence in **Live check**: a waived step runs on its cheapest safe proxy — a scratch database, a throwaway container — and the report names it.
**Considered:** forbidding waivers — the operator owns the call; a reviewer prompt rule — the live check runs in `flow.verify`, not the panel.

### The server clock is pinned where it can be

**ID:** server-clock-pin
**Status:** active
**Chosen:** a sentence after the verifier's server-clock exception: where the checkout can send the server its today, both ends pin to the one instant.
**Considered:** naming `X-Gymie-Today` as a contract — a gymie mechanism; the sentence cites it as the example only.

### Rewritten baselines are read before staging

**ID:** rewritten-baseline-read
**Status:** active
**Chosen:** stated once under **The sweep**, cited from the fix loop's recapture.
**Considered:** a pixel-diff guard — whether a rewrite shows the intended screen is a judgement the reader makes, not a threshold.

## Open questions
