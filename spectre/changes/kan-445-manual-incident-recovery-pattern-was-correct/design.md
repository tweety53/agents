# kan-445-manual-incident-recovery-pattern-was-correct — design

## Context

The recovery sequence this change scripts is fixed by the incident it reproduces (KAN-423, scripted
per KAN-445): diagnose from `git reflog -g HEAD`'s alternating reset pattern, `git revert --abort`,
then restore untracked planning files from the stash's third parent with `git show
"stash@{0}^3:<file>" > <file>` redirects. The ordering is not stylistic — a restore that stages
(`git checkout stash@{...} -- <path>`) is discarded by any later `revert --abort`, which is exactly
how this run's first restore attempt was lost.

Constraints: the repository is Bash + Python with every guard script bash 3.2–compatible and
`set -euo pipefail`-style strictness; test harnesses follow the `scripts/test-*.sh` naming that
`scripts/run-guard-tests.sh` discovers by glob; sandbox repos are `mktemp -d` with a cleanup trap.
No existing script, doc, or skill mentions the pattern (checked across `scripts/`, `docs/`,
`skills/`, `.flow/`), so this change adds new files only. `spectre/specs/` is empty — no capability
spec is reached or edited.

## Change

### CLI

```bash verified:authored verbatim in this file's CLI section — the design is the source
scripts/recover-guard-incident.sh [--apply] [repo-dir] [path...]
```

- `repo-dir` defaults to the caller's cwd and must resolve through `git rev-parse --show-toplevel`.
- `path...` defaults to `docs/superpowers` and `spectre/changes`; any paths given replace the
  defaults and are interpreted relative to the repo root, never the caller's cwd.
- `--help` prints usage. Exit codes: `0` success, `1` precondition failure (named cause printed),
  `2` usage error.

### Preconditions (checked in order; a non-repository exits 2 as a usage error; the remaining precondition failures exit 1 with their cause)

1. `repo-dir` is a git repository.
2. A revert is in progress: `git rev-parse -q --verify REVERT_HEAD` succeeds in that repo.
3. `git rev-parse -q --verify "stash@{0}^3"` succeeds. No stash at all and a stash created without
   `-u` (no third parent) are distinguished in the message — the remedies differ (wait/inspect
   versus re-stash with `-u` before any abort).
4. Every planned restore target is missing or untracked. A target tracked in HEAD or the index
   refuses the whole run — the tool never clobbers tracked state — reported on the first offending file. An
   untracked file already at a target is overwritten by the stash copy, named as an overwrite in
   the plan output.

### Diagnosis, plan, execution

Regardless of mode, print the last 15 `git reflog -g HEAD` entries so the alternating reset
pattern is visible at the point of decision.

Enumerate the restore set per path with `git ls-tree -r --name-only "stash@{0}^3" -- <path>`,
filtered through the missing/untracked check.

Dry-run (no flag): print each command that would run — `git -C <repo> revert --abort`, then per
file `mkdir -p <dir>` and `git -C <repo> show "stash@{0}^3:<file>" > <file>` — and exit 0 having
changed nothing.

`--apply`: run `git revert --abort` first, then every restore, each step printed as it runs, first
failing step exiting nonzero with the command shown. All aborts precede all restores; the
redirects leave restored files unstaged, so nothing after the abort point can discard them.

### Testing

`scripts/test-recover-guard-incident.sh` builds a `mktemp -d` sandbox repo per case (cleanup trap,
bash 3.2–compatible):

1. no revert in progress → exit 1, cause named, nothing changed;
2. revert in progress, stash taken without `-u` (no `^3` parent) → exit 1 with the
   no-third-parent cause;
3. dry-run happy path → exit 0, plan lists the abort and per-file restore commands, working tree
   and `REVERT_HEAD` untouched;
4. `--apply` happy path → `REVERT_HEAD` gone, planning files back with stash content, `git status`
   shows them untracked with nothing staged — the property the incident turned on;
5. tracked file at a target path → exit 1 refusal, revert still in progress, tracked content
   untouched.

## Decisions

### Generic tool over this-repo-only

**ID:** generic-recovery-tool
**Status:** active
**Chosen:** `scripts/recover-guard-incident.sh` takes any repo dir and any planning paths, with
this repo's trees as defaults — the kan-451 incident table is per-project, so incidents in other
projects are in scope for the same recovery.
**Considered:** this-repo-only hard-coded paths — simpler, but unusable for the same incident in
any other project; runbook-only doc — a doc reminds, the ticket asks for deliberate reproduction,
which a script enforces.

### Dry-run default, `--apply` to execute

**ID:** dry-run-default
**Status:** active
**Chosen:** without flags the tool verifies and prints only — matching how the incident was
actually recovered (diagnosis before action) and giving the harness a no-side-effect assertion
surface.
**Considered:** always-execute — one less flag, but a misdirected invocation does irreversible
work with no pause; interactive confirm — unusable from a harness without a `--yes` escape, which
is always-execute with extra steps.

### Restore via `git show` redirects, one abort before all restores

**ID:** redirect-restore-after-single-abort
**Status:** active
**Chosen:** the incident's own working order, made normative — abort first, then restore unstaged,
so no later abort can discard the restore (the failure that cost the first attempt).
**Considered:** `git checkout stash@{0}^3 -- <path>` — stages the restore, which a subsequent
`revert --abort` discards; ruled out by the incident itself.

### Refuse tracked restore targets

**ID:** refuse-tracked-targets
**Status:** active
**Chosen:** a target tracked in HEAD or the index refuses the whole run — the incident scenario
has the planning files absent, so a tracked file at a target means the wrong repo or the wrong
assumption, and clobbering tracked state silently is the one irreversible mistake the tool could
make on its own.
**Considered:** overwrite-with-warning for tracked files too — fewer cases, but it lets the tool
destroy committed content without a human ever seeing a prompt.

## Open questions

(none — everything the brainstorm surfaced was settled in-round.)
