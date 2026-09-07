# kan-448 — design

## Context

Two KAN-423 self-review findings, filed as KAN-448 and companion KAN-439. KAN-442 already removed
the incident's own mechanism (`check-task-commit-fields.py` is read-only in git now), so part 1
delivers the ticket's post-guard sanity check generically and wires it into the two guards that
still mutate a tree — `scripts/mutate-and-verify.sh` and `scripts/prepare-archive-branch.sh`.
Part 2 automates the cross-repo `spectre link` so peer resolution runs where it works and a
failure gates the stage. One change: both are flow-automation in this repository's own scripts and
skills, share the incident origin, and `spectre/specs/` holds no capability specs to amend — the
contracts live in the scripts' own headers and `skills/flow/implement.md`.

Full design record: `docs/superpowers/specs/2026-09-07-kan-448-design.md` (this file is canonical
for the decisions; that document for the design presentation and its evidence).

## Part 1 — post-mutation self-check

**`scripts/lib/post-mutation-check.sh`** — sourced library, bash 3.2 floor (indexed arrays only,
no associative arrays):

- `snapshot_tree_state <worktree>` — prints `git -C <worktree> status --porcelain=v2`, a sentinel
  line, then `git -C <worktree> stash list`. The caller captures it before its first mutation.
- `check_tree_restored <worktree> <snapshot>` — recomputes both and diffs against the snapshot.
  Each stash entry the snapshot does not hold prints as `new stash entry: <name>`; each status
  line the snapshot does not hold prints as an unexpected status line. Nonzero on any output,
  zero when the tree matches the snapshot. Only new entries are drift; a deleted stash entry is
  not residue the ticket names.
- Both read-only.

**`scripts/mutate-and-verify.sh`:** snapshot after all refusal paths, immediately before step 10's
`git apply`. The EXIT trap keeps its order — restore, then the touched-file residual test (exit 3,
precedence) — and runs `check_tree_restored` after it: drift forces exit 2 with the names printed.
The header's exit-code table documents both meanings of 2 ("refused before mutating anything" and
"post-restore drift detected, names in the output"); exit 3's contract is unchanged. A SIGTERM
mid-harness still fires the trap, so residue is caught and named on the killing exit itself.

**`scripts/prepare-archive-branch.sh`:** snapshot immediately after the dirty-tree refusals; check
after `git checkout -q $BASE` and the `git merge --ff-only` succeed; drift → stash names printed,
exit 2. Header gains the exit-2 drift meaning.

**Tests:** new `scripts/test-lib-post-mutation-check.sh` (clean roundtrip passes; planted stash
detected and named; unexpected status line detected; a snapshot-recorded stash still present is
not drift). `scripts/test-mutate-and-verify.sh`: planted stash → exit 2 naming it; touched-file
residual plus a planted stash still exits 3. `scripts/test-prepare-archive-branch.sh`: planted
stash after the refusal point → exit 2 naming it.

## Part 2 — cross-repo spectre link

`skills/flow/implement.md` §2: after each additional worktree's `git worktree add`, run as one
step `cd <that repo's primary checkout>` then `spectre link --root <worktree>/spectre
<canonical-peer>:<name>`. cwd at the primary checkout is what makes the peers file's relative
entries resolve (`ResolvePeer`, `spectre/internal/tree/peer.go:60`, stats the declared path
against the process working directory); `--root <worktree>/spectre` is what writes the
satellite-side `link.md` inside the worktree, where `check-unfinished-work.sh:252` reads it at
integrate. A refusal hard-fails `flow.isolate-workspace`; the "the link is not a gate" sentence is
deleted. A change with one worktree runs nothing (unchanged); the canonical worktree's creation
carries no link (it is the link's anchor, not a consumer).

Consequences recorded: the peer-side write (`## Parts` / `## Merge order`) lands in the canonical
repo's primary tree at its declared peers path and is that project's to commit by its own
convention; the worktree's `link.md` is committed per `skills/flow-contracts/git-boundaries.md`'s
existing carve-out, unchanged. KAN-439 closes when this change lands;
`check-unfinished-work.sh` is untouched.

## Decisions

### Part 2 resolves KAN-439

**ID:** part-2-closes-kan-439
**Status:** active
**Chosen:** the link automation is KAN-439's own first-named fix, so KAN-439 is closed when this
change lands; no extra guard work in scope and `check-unfinished-work.sh` is not touched.
**Considered:** the guard-side `plan-repo:` fallback (KAN-439's second alternative) — declined by
the operator as larger scope touching the guard; keeping the issues fully separate — rejected
because it risks the same failure being fixed twice or not at all.

### Scope after KAN-442

**ID:** post-kan-442-scope
**Status:** active
**Chosen:** a generic self-check library wired into the two surviving mutating guards,
`mutate-and-verify.sh` and `prepare-archive-branch.sh` — the ticket's premise (revert/reset-cycle
guards) is partly obsolete post-KAN-442, and the generic mechanism is what the next mutating guard
inherits. **Considered:** wiring `mutate-and-verify.sh` only — rejected, leaves
`prepare-archive-branch.sh`'s checkout/ff unguarded and hardcodes the check; declaring part 1
already resolved by KAN-442 — rejected, `mutate-and-verify.sh`'s trap is narrower than the ask
(touched files only, no stash delta), so the blind-retry failure stays reachable.

### Exit codes in mutate-and-verify.sh

**ID:** exit-2-and-3-coexistence
**Status:** active
**Chosen:** the new post-restore drift check exits 2 per the ticket, reusing the
refused-before-mutating code with both meanings documented in the header; the touched-file
residual keeps its documented exit 3 and takes precedence when both drift.
**Considered:** a fresh exit code for drift — rejected, contradicts the ticket's stated contract;
retiring exit 3 — rejected, it is the more specific failure and its documented contract and tests
already exist.

### The link runs from the primary checkout

**ID:** link-runs-from-primary-cwd
**Status:** active
**Chosen:** peer resolution executes with cwd at the satellite repo's primary checkout; a refusal
hard-fails the stage instead of logging a non-gating skip. **Considered:** running the link in the
worktree (today's text) — rejected, the peers entries are cwd-relative, resolve into
`.worktrees/` from a worktree, and the link is always refused (KAN-439's logged cause).

### Placement is after git worktree add, deviating from the ticket's "before"

**ID:** link-placement-after-add
**Status:** active
**Chosen:** after-add with `--root <worktree>/spectre`, per the shipped-code evidence: peers
paths resolve against cwd, `spectre link` writes the satellite side under the `--root` tree, and
`check-unfinished-work.sh:252` reads `link.md` in the worktree — so the ticket's literal
before-add placement writes `link.md` into the primary tree only and reproduces the false
OUTSTANDING the ticket exists to remove. **Considered:** before-add plus a copy step into the
worktree — rejected, a hand-rolled second step where one `--root` invocation does both halves.
Supersedes nothing; recorded before any artifact existed.

### prepare-archive-branch.sh counts as a guard

**ID:** archive-branch-script-is-a-guard
**Status:** active
**Chosen:** the ticket's "every guard that mutates the tree" covers it — its checkout and
fast-forward in the landing worktree can be killed mid-flight like any guard's mutation.
**Considered:** limiting part 1 to scripts named `check-*`/`mutate-*` — rejected, the failure mode
(residue after a killed mutation) rather than the filename is what the ticket targets.

## Open questions

None recorded — the convergence check closed with no unresolved item.
