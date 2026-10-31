## Context

The approved design is round 2 of the planner session
(`.superpowers/sdd/planner/round-2.md`, answers in `answers-1.md` and `answers-2.md`); this file
records it. It carries the KAN-1072 umbrella follow-up
(`docs/audits/2026-10-31-spectre-archive-audit.md` § "The umbrella follow-up", items 1–4; item 5
landed as agents `50e0a1f7`) and goes one step past its item list: the change directory is never
committed at all, instead of committed and dropped.

It is one change because every piece removes the same commit family. Gitignoring
`spectre/changes/` without the guard rewrites leaves `check-plan-unchanged.sh` always passing
(`git ls-files --exclude-standard` lists nothing ignored) and `check-late-fix-trigger.sh` always
exiting 4 (`ls-files --error-unmatch` fails on an ignored plan); rewriting the guards without the
gitignore leaves planning commits in place; and the contract text names both.

It spans six repositories through `spectre/peers`: agents (canonical), spectre, gymie,
gymie-frontend, gymie-playwright and flow-eval. Every repository's worktree is
`/Users/tweety53/Projects/<repo>-worktrees/kan-1073-commit-only-what-is-needed-permanently` on
branch `spectre/kan-1073-commit-only-what-is-needed-permanently`; `link.md` records the merge order.

This run itself is integrated by the main checkout's installed tooling, which still commits the
plan, merges the merge-order record and drops the directory at run 1. Its planning commits and its
drop are the last of their kind; the reinstall after landing (`./setup.sh global`,
`make -C stats restart`) moves every later run onto the new behaviour.

## Decisions

### The change directory is gitignored, never committed

**ID:** change-dir-gitignored
**Status:** active
**Chosen:** add `spectre/changes/` to `.gitignore` next to `.superpowers/` in all six repositories;
the directory keeps its path and lives only on disk in the change's worktree — no `git add`,
commit, drop, archive or restore of it anywhere. Visual evidence (`visual-verification/`,
`.frame.png`, `matrix/*.json`, `evidence/`) lives under it, which settles umbrella item 1 with no
rule of its own.
**Considered:** keep committing and drop at landing (today's shape, the umbrella's item list) —
history still carries every plan and drop commit, and the drop leaves residue (`50e0a1f7`); move
the directory into `.superpowers/` — every path in every contract and guard changes for no gain
over one ignore line.

### flowd keeps two files per change, saved once at cleanup

**ID:** change-files-two-rows-at-cleanup
**Status:** active
**Chosen:** a `change_files (change_id, path, body, saved_at)` table, unique on
`(change_id, path)`. `flow change save <name>` upserts exactly `tasks.md` (from
`spectre/changes/<name>/`) and `self-review-context.md` (from `.superpowers/sdd/<name>/`) of the
canonical worktree and refuses when either is missing; `flow change file <name> <path>` prints one.
It runs once, in run 2's cleanup, before the worktrees are removed; `remove-change-worktrees`
refuses while the change has no `change_files` rows. A worktree lost mid-run loses its plan — the
accepted cost (operator answer round 1 Q1).
**Considered:** save at the plan gate and at cleanup — a second save call for a lost-worktree case
the operator accepted; save on every `flow state` write — couples state writes to the change
directory; save the whole directory — no reader remains for `design.md`, `proposal.md`, narratives,
link files or images once the rejection search is gone.

### The rejection search over archived designs is deleted

**ID:** drop-rejection-search
**Status:** active
**Chosen:** the kan-970 bullet of `skills/spectre-planning/SKILL.md` keeps its `git log --since`
check and its cheap reproducer; the archived-design search and its "remedy an archived decision
rejected" clause go, with every passage citing them. No `flow change search` (operator answer
round 1 Q2).
**Considered:** full-text search over saved `design.md` bodies — keeps the safeguard but makes
flowd store designs no other reader needs; one flowd row per `## Decisions` entry recorded at the
plan gate — a `## Decisions` parser and a record step for the same safeguard.

### No self-review report; the bundle is written to disk and saved to flowd

**ID:** self-review-bundle-to-flowd
**Status:** active
**Chosen:** no `docs/self-review/<name>-self-review.md` is written or committed; findings already
live in `self_review_findings`. The in-run bundle is built from the on-disk files and written to
`.superpowers/sdd/<name>/self-review-context.md`; its finish range is `merge-base..HEAD`, since no
planning commit remains to skip. A deferred `/flow-self-review` reads the saved bundle through
`flow change file <name> self-review-context.md`, never individual files (operator answer round 1
Q3). `land-self-review-report` and `check-self-review-report` are deleted.
**Considered:** report and bundle under `.superpowers/` only — a deferred self-review after cleanup
becomes impossible; drop the deferred `/flow-self-review` — removes a working command to save one
row.

### The self-review land check reads the stage row

**ID:** self-review-land-check-stage-row
**Status:** active
**Chosen:** integrate's land step checks that flowd holds a `flow.self-review` stage run with
outcome `completed` for the change (umbrella item 2), through a new
`GET /api/v1/stages/{project}/{change}/{stage}/completed` route over the existing
`store.StageCompleted` and a client method `StageCompleted`; where it checked
`HEAD:docs/self-review/<name>-self-review.md` before, it asks flowd now.
**Considered:** keep a report file under `.superpowers/` as the land marker — a second record of
what the stage row already records.

### Delete the tracked report and change files from every tip

**ID:** delete-tracked-leftovers-from-tip
**Status:** active
**Chosen:** each repository's tip loses every file still tracked under `spectre/changes/`
(decision leftovers-untracked-by-planning-commit) and `docs/self-review/` (agents 246 reports, gymie 112, gymie-frontend 5,
gymie-playwright 1, flow-eval 5, spectre 2), plus the placeholder `spectre/specs/.gitkeep` where it
is the only file (agents, gymie-frontend, flow-eval). History keeps every file (operator answer
round 1 Q4).
**Considered:** leave them — no reader exists, and leaving them keeps the tree carrying what the
change says is never committed.

### The tracked change-directory leftovers leave by a planning commit

**ID:** leftovers-untracked-by-planning-commit
**Status:** active
**Chosen:** the files still tracked under `spectre/changes/` (agents 33, spectre 4, gymie-frontend
11, gymie-playwright 9, flow-eval 1) are untracked by the parent after the last task closes, as one
trailerless `chore(spectre): untrack the leftover change files` planning commit per repository
behind the installed `check-planning-commit-location.sh`, this change's own directory excluded; the
task commits carry the ignore line, the `docs/self-review/` deletion and the placeholder only.
**Considered:** deleting them in the per-repository task commit — `check-task-commit-planning-paths.sh`
refuses any task commit touching `spectre/changes/` (found while planning: `check-plan-shape.sh`
F23 refused the plan's `spectre/changes/**` collateral).

### No backward compatibility

**ID:** no-backward-compat
**Status:** active
**Chosen:** `stats/internal/landed`, the `chore(spectre): archive` fallback, the "bundle an older run
committed under `docs/self-review/`" path, the merge-order `landed` fallback and every
live-else-archive read are deleted, not kept.
**Considered:** keep the history fallbacks for changes landed before this one — every such change
has landed and been cleaned up; no in-flight run depends on them after the reinstall.

### spectre loses archive, refs and citations with their dead code

**ID:** spectre-archive-refs-removed
**Status:** active
**Chosen:** remove `spectre archive`, `spectre refs`, cross-tree requirement citations and every
archive-aware path (`tree.ArchiveDir`, the archived read, `Change.Archived`, the reserved `archive`
id, list's archived peer fallback, link's archive search and "archive skew", `testtree.Change`'s
`archived` parameter), with their docs (umbrella item 4). The list/validate half of
`TestFencedTaskExampleDoesNotFoolListValidateOrArchive` survives as
`TestFencedTaskExampleDoesNotFoolListOrValidate`.
**Considered:** keep the commands for repositories outside flow — nothing archives once the
directory is never committed, and an unused code path rots.

### Guards read the plan from disk; snapshot guards walk the directory

**ID:** guards-read-disk
**Status:** active
**Chosen:** every in-run guard that read the plan from a commit or from history reads the
worktree's file. `check-plan-unchanged.sh` walks `spectre/changes/<name>/` with `filepath.WalkDir`
instead of `git ls-files --exclude-standard`. `check-late-fix-trigger.sh`'s new-task rule compares
the canonical `tasks.md` against `.superpowers/sdd/<name>/tasks-at-close.md`, the copy the review
panel's close step writes where it used to commit the plan; a missing copy with a `since` other
than `-` is exit 4, as an untracked plan is today.
**Considered:** keep `git` reads with `--ignored` — reads the same file through a slower and less
obvious route; drop the late-fix new-task rule — removes a working scope guard.

### Merge order: spectre first

**ID:** merge-order-spectre-first
**Status:** active
**Chosen:** `link.md`'s `## Merge order` is spectre, agents, gymie, gymie-frontend,
gymie-playwright, flow-eval: spectre ships the CLI the agents tooling calls; the four project
repositories carry only the ignore line, the tip cleanup and project text.
**Considered:** agents first, as kickoff wrote it — agents would land text describing a spectre
without `archive` before spectre drops it.

### gymie keeps its specs

**ID:** gymie-specs-stay
**Status:** active
**Chosen:** gymie's `spectre/specs/` (33 real files) stays tracked; only placeholder-only trees lose
their `.gitkeep`.
**Considered:** delete every `spectre/specs/` — gymie's are live specs planners read.

### flow-eval's miner reads plan snapshots from the change's worktree

**ID:** plan-snapshots-for-the-miner
**Status:** active
**Chosen:** where a planning commit used to be made — the plan gate's **Yes**, `/flow-plan`'s
capture, and a fix run's tasks append (`flow.document-fix`) — the run copies `tasks.md`,
`proposal.md` and `design.md` of `spectre/changes/<name>/` into the next numbered directory
`.superpowers/sdd/<name>/plan-snapshots/<k>/` of the canonical worktree, never committed and never
rewritten (**Plan snapshots**, `skills/flow/brainstorm.md`). flow-eval's `case add impl` reads, for
task n, the earliest snapshot whose `tasks.md` holds task n — the plan as it stood before task n
was implemented — and refuses with `no plan snapshot … holds task n` when none does; `case add
impl|review` reads the ledger and panel record from the rendered
`.superpowers/sdd/ledgers/<change>.md` and `.superpowers/sdd/reviews/<change>-panel.md`, and refuses
with `plan files gone — mine before cleanup` once `spectre/changes/<change>/` is gone. `_archive`,
`planAt`, the plan-repo `prereshape` call and every `git show`/`git log` on `spectre/changes` are
deleted, with no fallback; the repo's `prereshape` and `taskCommits` stay — they find the task
commits, not the plan. Mining runs in the self-review's step 5 inside integrate run 1, on every
route, before cleanup removes the worktree; the pinned `refs/flow/pre-reshape/<change>` on origin
keeps a case's commits reachable without waiting for the merge.

**Audit — every input a case ships carries only pre-task state** (operator follow-up: a case must
never carry its solution in any file it ships). A trial receives only `run._inputs` per role;
`case.json`, `heldout.patch`, `reference.stat` and `labels.json` are the hidden grading answer.

| Input | Reaches the agent | Source | Verdict |
|-------|-------------------|--------|---------|
| impl `task.md` | yes | task n's section of the earliest snapshot holding task n | pre-task — was the live plan, whose review-fix Corrections name the fix: the leak this decision closes |
| impl `context.md` (`proposal.md`, `design.md`) | yes | the same snapshot | pre-task — was the live plan, whose design decisions later fixes rewrite: closed with `task.md` |
| impl tree | yes | `git archive` of the reference commit's parent, re-initialised as one commit `base` | pre-task: an ancestor of the task; no history, message or trailer of any later commit is reachable; the gitignored plan and `.superpowers/` are not in a tree |
| review `diff.patch` | yes | `git diff base head`, `spectre/changes` excluded, head the last implementer commit before the first panel-fix | pre-review: no panel-fix change, no commit message |
| review tree | yes | `git archive` of head, re-initialised | pre-review, as above |
| visual `mockup.png`, `capture.png` | yes | the mockup and the capture the operator names | the evaluated pair itself; the verdict and defects stay in `labels.json` |
| narrative, `visual-verification.md` notes | no | not read for an impl or review case | — |
| ledger | no | picks the reference commit; its rows reach no shipped file | grading only |
| panel record | no | `labels.json` | grading only (post-task by design) |
| reference commit, its subject and trailers | no | `heldout.patch`, `reference.stat`, `case.json`'s `notes` | grading only |

`TestMineShipsOnlyPreTaskState` plants `_leak` in every post-task source — the live plan, a later
snapshot, the ledger, the panel record, the panel-fix commit's code and message — and asserts no
input a trial receives carries it.

**Considered:** read the live plan as it stands when mined (this decision's previous form, operator
answer at the plan gate "can it just mine it while files exist?") — rejected by the operator: a
review fix writes Corrections into `tasks.md` and decisions into `design.md` that name the fix, so
the case carries its own answer, exactly what flow-eval's `_leak` test guards; keep the history
reads — they see no change landed after this one, since nothing under `spectre/changes/` is
committed; reuse `.superpowers/sdd/<name>/tasks-at-close.md` (decision guards-read-disk) — written
once at the review panel's close, after the review fixes, so it is the leaking plan, and it holds no
`proposal.md` or `design.md`; read `tasks.md` from `change_files` — saved only at cleanup, after
every fix, and without `proposal.md` and `design.md`; one snapshot per task at its dispatch — a step
in every implementer dispatch for what the plan gate's and the fix run's snapshot already give, since
no task is added anywhere else.

## Recommended approach

Gitignore `spectre/changes/` and make every consumer read the worktree's file on disk; keep in
flowd only the two files a reader needs after the worktree is gone, saved once at cleanup; delete
everything that existed to commit, drop, archive, restore or find the plan in history, the
self-review report and its guards, and the spectre commands built on archiving. Rejected:
commit-and-drop (change-dir-gitignored), saving more than two files or saving earlier
(change-files-two-rows-at-cleanup), keeping a rejection search (drop-rejection-search), keeping a
report file (self-review-bundle-to-flowd, self-review-land-check-stage-row), keeping history
fallbacks (no-backward-compat).

The spectre CLI lands first, then agents' stats tooling (flowd table and CLI before the guards and
cleanup that call it), then the contract text, then each repository's ignore line and tip cleanup.
The agents ignore line lands last among agents' tasks, so the installed tooling's planning commits
of this run's own fix rounds still stage the plan without `-f` until then.

## Critical files

agents (canonical):

- `stats/internal/store/migrations/0059_change_files.sql`, `stats/internal/store/changefiles.go` — the new table and its store functions.
- `stats/internal/api/changefiles.go`, `stats/internal/api/server.go` — `PUT`/`GET` routes for a change file.
- `stats/internal/api/stages_completed.go` — the stage-completed read route.
- `stats/internal/client/changefiles.go`, `stats/internal/client/stages_completed.go` — client methods.
- `stats/cmd/flow/change.go`, `stats/cmd/flow/main.go` — `flow change save|file`.
- `stats/internal/api/stats.go` — `readPlanFile`: disk, else `change_files`.
- `stats/internal/landed/` — deleted.
- `stats/internal/selfreview/git.go`, `stats/internal/api/selfreview.go`, `stats/cmd/flow/selfreview.go` — the bundle from disk, written to `.superpowers/sdd/<name>/self-review-context.md`.
- `stats/internal/guard/integratechange.go`, `integratechange_commit.go`, `integratechange_land.go`, `integratechange_prepare.go` — canonical lookup, merge order, archive stages, report land check.
- `stats/internal/guard/visualverifydispatched.go`, `verbatimmoves.go`, `unfinishedwork.go`, `loop.go`, `planunchanged.go`, `latefixtrigger.go` — plan reads.
- `stats/internal/guard/commitarchive.go`, `archivescope.go`, `planningcommitlocation.go`, `asideplanningartifacts.go`, `landselfreviewreport.go`, `selfreviewreport.go` and their scripts — deleted.
- `stats/internal/guard/cleanupcomplete.go`, `pipelinefix.go`, `taskrecords.go`, `recoverguardincident.go`, `removechangeworktrees.go` — drop rows, `<k>` uniqueness, stash recovery, the save gate.
- `stats/cmd/flow/state.go`, `stats/cmd/flow/record.go` — listing and checkout check.
- `stats/internal/lessons/lessons.go` — lesson sources.
- `stats/internal/stages/names.go`, `stats/web/src/lib/phase.ts`, `stats/web/src/lib/stageLabel.ts`, `stats/web/src/lib/mainNow.test.ts`, `mods/flow-task-list/hooks/register.tsx` — the two archive stage keys.
- `hooks/protect-main-checkout.py` — the drop-enforcing refusal.
- `skills/**` contract text listed per task; `README.md`, `.flow/project.md`, `.gitignore`.

spectre: `cmd/spectre/main.go`, `internal/cmd/{archive,refs,new,list,init}.go`,
`internal/check/{refs,citations,check,link}.go`, `internal/parse/{ref,spec}.go`,
`internal/model/model.go`, `internal/config/config.go`, `internal/tree/{tree,read}.go`,
`internal/testtree/testtree.go`, `README.md`, `docs/*.md`, `.claude/skills/spectre/SKILL.md`,
`.claude/skills/spectre-planning/{SKILL,UPSTREAM}.md`, `.gitignore`.

gymie, gymie-frontend, gymie-playwright, flow-eval: `.gitignore`, `docs/self-review/`,
`spectre/changes/`, `spectre/specs/.gitkeep` where placeholder-only, gymie `.flow/project.md`.

flow-eval: `internal/mine/mine.go`, `internal/mine/mine_test.go`, `cmd/flow-eval/main.go`,
`README.md`, `docs/cases.md` — the miner reads the worktree's plan snapshots.

## Existing code to reuse

- `store.StageCompleted` (`stats/internal/store/stageruns.go`) — the land check's query, exposed by a route.
- `store/requests.go`'s change-id resolver and `0056_requests.sql`'s `change_id … REFERENCES changes(id)` / `UNIQUE` shape — the `change_files` table and its resolver.
- `internal/api/specs.go`, `internal/client/specs.go`, `cmd/flow/spec.go` — the write/read endpoint, client and CLI pattern; `callRecord` for the CLI's flowd call.
- `api/selfreview.go`'s text/markdown response — the shape of the file GET.
- `selfreview.Bundle` (`stats/internal/selfreview/git.go`) — kept; only its sources change.
- `ChangeDir` in `stats/internal/api/stats.go` — the on-disk plan read already there.
- `scripts/live-store-copy.sh` — the live check's stand-in store.

## Verification

- Each task's own lint and targeted tests; the parent's full-suite run after the last group.
- `git -C <each worktree> ls-files spectre/changes docs/self-review` prints nothing after the tip
  cleanups, and `git check-ignore spectre/changes/x` succeeds in every repository.
- `git grep -n -F -e 'chore(spectre)' -e 'commit-archive' -e 'land-self-review-report' -e 'check-self-review-report' -e 'check-planning-commit-location' -- skills rules scripts stats hooks README.md` in agents prints nothing outside `docs/`.
- `spectre archive` and `spectre refs` print the unknown-command usage in the spectre worktree's build.
- Integrating this change itself: the main checkout's installed tooling still commits and drops (Context);
  the first `/flow` run after the reinstall produces a landed branch with no `chore(spectre)` and no
  `docs(self-review)` commit.

## Live check

The change adds a flowd migration, two routes and a cleanup gate, so it is checked against a copy of
the real store, never the dev stack:

1. `scripts/live-store-copy.sh up kan-1073-live` from the agents worktree, then export the printed
   `FLOW_ADDR` / `FLOW_RECORDS_ADDR`.
2. Before: `psql` on the copy, `SELECT count(*) FROM change_files` — the table exists (migration ran
   on the copy) and holds 0 rows.
3. Write `.superpowers/sdd/kan-1073-live/self-review-context.md` and
   `spectre/changes/kan-1073-live/tasks.md` (one task line) in the agents worktree; run
   `flow change save kan-1073-live`; after: the count is 2, and `flow change file kan-1073-live tasks.md`
   prints the written body byte for byte.
4. Remove `self-review-context.md` and run `flow change save kan-1073-live` again: it exits non-zero
   naming the missing file, and the count stays 2.
5. `curl -s "$FLOW_ADDR/api/v1/stages/<project>/kan-1073-live/flow.self-review/completed"` reports not
   completed; after `flow stage begin` / `end -outcome completed` of `flow.self-review` for the
   scratch change it reports completed.
6. `scripts/live-store-copy.sh down kan-1073-live`, and remove `spectre/changes/kan-1073-live/`.

Not working looks like: a save exiting 0 with the count still 0, a file read printing nothing, a
save with a missing file succeeding, or the completed route answering the same before and after the
stage end.
