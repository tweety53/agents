## Context

- Five Jira tickets planned as self-review findings per item 1 of **3. Fix every finding that is not
  `big`** (`skills/flow-self-review/SKILL.md`), on `self-review-kan-urgent-5` based at `ebc10e49`.
  Each fixable finding is carried by exactly one task; the brainstorm design is
  `.superpowers/sdd/2026-10-18-self-review-kan-urgent-5-design.md` in this worktree.
  - KAN-963 (Task 1), blast radius 2: reproduced at `ebc10e49` — with `<sib>/lsub` a symlink to
    `<main>/sub`, the hook allows `cat > <sib>/lsub/../f`, which the shell writes as `<main>/f`. The
    ticket's literal form, `link-to-main/sub/../f` with `sub` a real directory, is already denied:
    the lexical collapse stays inside the link, which still resolves into the checkout.
    `protected()` is the hook's one path judge (`hooks/protect-main-checkout.py`, its only call site
    the per-path loop), so the fix sits there.
  - KAN-972 (Task 2), blast radius 3: `tcfCheckTaskCommit` (`stats/internal/guard/taskcommitfields.go`)
    ran `tcfCheckCommitSubject` on every listed commit; the shim header
    (`scripts/check-task-commit-fields.sh`, THE COMMIT MAP) states the rule.
  - KAN-976 (Task 3), blast radius 3: the guard builds one `tcfRepoDiff` per map pair, so a commit
    judged alone that finishes a task fails on the declared files and tests it never touched. The
    files the ticket counts as citing the guard cite its name and arguments, neither of which the
    fix changes, so none is in the radius.
  - KAN-984: disproved, already fixed. `bd73322c` (2026-10-15) matches a satellite worktree's diff
    against the canonical `ui paths` with its repository's name prefixed; gymie's
    `docs/self-review/kan-982-prod-fixes-group-add-person-tap-set-value-bleed-self-review.md`
    records it as "the case KAN-984 filed"; at `ebc10e49`
    `go test ./internal/guard/ -run VisualVerifyDispatched` passes, including "canonical ui paths
    name the satellite: no verifier is MISSING"; every caller — `integrate-change.sh`'s
    `visual_ok`, `check-unfinished-work`'s fold, `verify-and-handoff.md`'s per-worktree run — passes
    the canonical worktree. No task.
  - KAN-987 (Task 4), blast radius 6: `cd stats && go list -deps ./cmd/flow ./cmd/flow-guard | grep
    internal/web` prints nothing (exit 1) at `ebc10e49`, nor does it for `./internal/guard/`. The
    `## test` and `## lint` lines that compile `internal/web` (`go test ./...`, `go vet ./...`) or
    need `stats/web/node_modules` (`npm test`, `npx tsc -b`) are the only consumers of the build
    the setup made; `CLAUDE.md` and `AGENTS.md` repeat the two lint lines.
- Spiked at plan time, every edit reverted: the four diffs apply in task order at `ebc10e49`
  (`git apply`); with all applied the whole `## lint` list exits 0, `scripts/run-guard-tests.sh`
  passes (42 harnesses), and `go test ./internal/guard/` passes. Each task's new case fails with
  its fix reverted and its test kept.

## Decisions

### Judge a write path after its symlinks resolve

**ID:** protect-realpath
**Status:** active
**Chosen:** `protected()` opens with `path = os.path.realpath(path)`, so `existing_dir` and the
worktree-root comparison see the path the write lands on.
**Considered:** `realpath` inside `existing_dir` alone — the worktree-root branch compares
`abspath(path)` against the resolved directory and would misjudge; resolving in `resolve()` — it
also feeds non-write callers, and a path the shell would expand is let through there before any
resolution.

### A cross-repo task's commits may each name their own scope

**ID:** subject-scope-per-repository
**Status:** active
**Chosen:** `tcfCheckSubjects` — one listed commit carries the declared subject exactly, and every
listed repository carries it on one of its commits with the `(<scope>)` ignored (`tcfUnscoped`); a
repository with none is reported on its first commit in `tcfCheckCommitSubject`'s words, and one
commit keeps today's verdict.
**Considered:** a per-repository `Commit:` syntax in `tasks.md` — a grammar change across the field
parser, `check-plan-shape` and `check-task-records`; dropping the subject check for every commit
past the first — a second repository's commit could then carry any subject.

### One verdict over every commit carrying the task's trailer

**ID:** task-id-commit-union
**Status:** active
**Chosen:** with no explicit parent, each listed pair is joined by every other commit from the
worktree's merge base (`vmDefaultBase`: the recorded `flowBase`, else origin/HEAD, else origin/main,
else main) to the listed commit whose `Task-Id:` trailer names the task; the existing merge —
union of files, concatenated diffs, summed counts, tree check, subject rule — judges them as one.
A worktree with no resolvable base judges the listed commit alone.
**Considered:** `close-task` gathering the commits — its merge base is the canonical repository's
alone, and the fix round's re-run and a hand run would still judge one commit; reading trailers
over the whole history — the base branch carries earlier changes' trailers (measured: 14 on
gymie-frontend's origin/HEAD, 4 on this repository's origin/main over 2000 commits).

### Build the SPA lazily, never at worktree setup

**ID:** web-dist-lazy
**Status:** active
**Chosen:** delete `## worktree setup`; `stats/Makefile` gains `web-dist`, a file target on
`internal/web/dist/index.html` running `$(MAKE) web-build`; the four `## test`/`## lint` lines that
need the build open with `make -s web-dist`, as do their two copies in `CLAUDE.md` and `AGENTS.md`.
**Considered:** keeping the setup behind a `[ -d internal/web/dist ] ||` guard — a fresh worktree
never has the directory, so it would still build on every creating run; skipping the setup by the
change's paths — the setup runs before any plan exists.

### First review effort and fixer grouping

**ID:** first-review-effort
**Status:** active
**Chosen:** first review on `flow-medium` (effort `medium`); one fixer dispatch (key
`self-review-kan-urgent-5-fix`) carrying Tasks 1–4 in order — each task is a verified diff applied
with `git apply`.
**Considered:** `flow-low` for the first review — Task 3 widens what every task close judges, and
its merge-base bound decides which history the guard reads, judgment a low-effort read may skim;
one fixer per task — four dispatches for four `git apply`s.

## Open questions

## Live check

No runtime to verify against: the hook, the guard and the Makefile target run per invocation and
hold no state; their harness cases exercise each against real git repositories.
