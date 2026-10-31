## Context

KAN-1072 began as an audit of what `spectre/changes/archive/` holds, who reads it, and what it
costs. The approved design is `.superpowers/sdd/planner/round-2.md` in the kickoff worktree as
amended by `answers-2.md`, under the operator's direction in `answers-1.md` (cut everything unused
or rarely used; `tasks.md` is the most useful artifact during a run and is not needed after
integration). Two amendments after approval (`answers-3.md`) widened the change: it also deletes
the existing archives at the tip in every repository spectre has run in, and it carries the
umbrella follow-up's item 1 itself — the ephemeral change directory, replacing every archive step
in the flow with a landing deletion commit. The change is therefore no longer audit-only: it ships
the audit document, the tooling change, and the per-repository tip cleanup, and files one umbrella
follow-up for the rest.

Round 1's inventory and round 2's verdict table are the audit document's content; they are not
restated here. The repositories holding a `spectre/changes/archive/` directory today:

| Repository | Archived changes | Tracked files under the archive | Default branch |
|---|---|---|---|
| agents | 364 | 2063 | main |
| gymie | 127 | 9348 | develop |
| gymie-frontend | 39 | 49 | develop |
| gymie-playwright | 18 | 22 | main |
| flow-eval | 5 | 35 | main |
| spectre | 4 | 15 | main |
<!-- measured: for d in ~/Projects/*/; do ls $d/spectre/changes/archive | wc -l; git -C $d ls-files spectre/changes/archive | wc -l; git -C $d symbolic-ref --short refs/remotes/origin/HEAD; done @ each repository's main checkout, 2026-10-31 (machine-local working trees, not re-runnable at a ref) -->

gymie-admin-frontend and intellij-review-queue carry no `spectre/changes/archive/`.

## Decisions

### Where the audit document lives

**ID:** audit-doc-location
**Status:** active
**Chosen:** `docs/audits/2026-10-31-spectre-archive-audit.md` in the agents repository — reviewable, the commands sit next to the numbers, linked from the Jira comment.
**Considered:** `docs/spectre-archive-audit-2026-10-31/README.md` after the `docs/prompt-audit-2026-09-29/` folder precedent — a folder for one file; a Jira comment only — loses diff review, and the pipeline needs a committed deliverable.

### Plan text is committed during the run and deleted at land

**ID:** plan-text-ephemeral
**Status:** active
**Chosen:** `proposal.md`, `design.md`, `tasks.md` and every other file of the change directory are committed on the change branch while the run needs them (the plan-gate commit the provenance and plan-unchanged guards read) and deleted by one landing commit; their text stays in history.
**Considered:** never committing them — needs the provenance, plan-unchanged and task-record guards rewritten to read flowd; copying the text into flowd at landing — a flowd text store no consumer needs beyond what history gives.

### Lesson resolve reads briefs and learning notes

**ID:** lesson-resolve-briefs-and-notes
**Status:** active
**Chosen:** `flow lesson resolve` drops its scan of archived `narrative.md` files and reads `docs/briefs/` plus flowd learning notes — both exist, and briefs are the lessons home by contract. Carried by the umbrella follow-up, not by this change.
**Considered:** storing each narrative in flowd — keeps today's coverage at the cost of a new table; keeping `narrative.md` in git as the one exception — contradicts the ephemeral rule.

### Recorded rejections are searched in flowd

**ID:** rejection-search-in-flowd
**Status:** active
**Chosen:** each `## Decisions` entry is recorded in flowd at the plan gate, and the rejection check (`skills/spectre-planning/SKILL.md`) searches flowd instead of the archive. Carried by the umbrella follow-up (its lessons-and-rejection-search item).
**Considered:** searching git history over `spectre/changes/` — the operator chose flowd after the kan-970 safeguard was explained; cutting the rule — loses the kan-970 safeguard against a remedy rejected twice.

### The self-review report leaves git

**ID:** self-review-report-out-of-git
**Status:** active
**Chosen:** stop committing `docs/self-review/<name>.md`; the `land` guard checks flowd's self-review pass row instead. Carried by the umbrella follow-up.
**Considered:** storing the report prose in flowd — no reader of past reports; keeping it committed, folded into the landing commit — keeps a file nobody reads.

### The spectre cut takes only what is unused

**ID:** spectre-cut-unused-only
**Status:** active
**Chosen:** cut `spectre archive`, `spectre refs`, cross-tree requirement citations and the placeholder `specs/` trees; keep the three plan files, gymie's specs, `new`, `validate`, `list`, `link` and `init`. The spectre-repository edits are the umbrella follow-up's; this change only stops flow from calling `spectre archive`.
**Considered:** also folding `proposal.md` into `design.md` — 16 skill files and the reviewer prompts cite it, for a file that is ephemeral anyway; also dropping gymie's `specs/` — live (last edit 2026-10-24) and read by planners.

### No history rewrite

**ID:** no-history-rewrite
**Status:** active
**Chosen:** nothing rewrites history; the audit records the numbers and the procedure only.
**Considered:** an operator-run rewrite for gymie alone, or for gymie and agents — irreversible, breaks every cited sha, flowd's `commit_sha`/`round_base` and every checkout; the tip deletion fixes working-tree size, not clone size.

### One umbrella follow-up for what this change does not do

**ID:** one-umbrella-follow-up
**Status:** active
**Chosen:** one Jira task covering the remaining items — visual evidence out of git; the self-review report out of git (decision self-review-report-out-of-git); lessons and rejection search off the archive (decisions lesson-resolve-briefs-and-notes, rejection-search-in-flowd); the spectre CLI cuts (decision spectre-cut-unused-only). Round 2's item 1 (ephemeral change directory) and item 6 (tip cleanup) are done here, item 7 (history rewrite) is not filed. Labels `flow` + one `flow-<topic>` + one `flow-<angle>`; transitioned to TO DO URGENT (transition id 2) right after creation; filed only after the operator has seen the findings and the task text.
**Considered:** one task per item — the operator chose one umbrella; only the byte-heavy items — leaves measured costs unfiled.

### The existing archives are deleted in this change

**ID:** tip-cleanup-in-this-change
**Status:** active
**Chosen:** one tip-only commit per repository deletes `spectre/changes/archive/` in agents, gymie, gymie-frontend, gymie-playwright, flow-eval and spectre (the six repositories the Context table lists). History keeps the files.
**Considered:** a later follow-up once every reader moved — the operator asked for it in this change, now.

### Archive readers come up empty until the umbrella follow-up

**ID:** archive-readers-empty-until-follow-up
**Status:** active
**Chosen:** once the tip cleanup lands, `flow lesson resolve`'s narrative scan and the spectre-planning rejection search over `spectre/changes/archive/` find nothing at the tip. That is accepted until the umbrella follow-up repoints them (decisions lesson-resolve-briefs-and-notes, rejection-search-in-flowd). The board's plan view and the self-review bundle keep working, because they read landed changes from history (decision landed-change-history-reader).
**Considered:** keeping the archive until the follow-up lands — the operator chose the cleanup now.

### A landing deletion commit replaces the archive step

**ID:** landing-drop-replaces-archive
**Status:** active
**Chosen:** run 1's archive step (`spectre archive` plus the `chore(spectre): archive <name>` commit made by `commit-archive`) becomes one commit that runs `git rm -r` on `spectre/changes/<name>/` and each `<name>-fix-<N>/` sub-change, in the canonical worktree and in every satellite worktree holding `spectre/changes/<name>/`. The self-review sub-change (`self-review-<name>`) and the pipeline-fix loop (`pipeline-fix-<name>-<k>`) drop their change directories the same way. The rendered ledger and panel records are no longer copied into the change directory; self-review patches and `verbatim-moves.txt` leave with it. The Done-when path check that ran before the archive commit runs before the drop commit.
**Considered:** keeping the archive move and deleting at a later cleanup — two commits for one outcome and an archive that lives until cleanup.

### The drop commit's subject

**ID:** drop-commit-subject
**Status:** active
**Chosen:** `chore(spectre): drop <name>` — a `chore` subject, so the untasked-commit and main-checkout landing checks treat it as bookkeeping exactly as they treated the archive subject.
**Considered:** `chore(spectre): remove <name>` — no difference in behaviour; "drop" is shorter and not already used by any other bookkeeping subject.

### Landed changes are read from history

**ID:** landed-change-history-reader
**Status:** active
**Chosen:** one helper, `stats/internal/landed/landed.go`, returns a landed change's file from git history: the parent of the newest `chore(spectre): drop <name>` commit at `spectre/changes/<name>/<file>`, else the newest `chore(spectre): archive <name>` commit at `spectre/changes/archive/<name>/<file>` (changes landed before this change). The self-review bundle, the merge-order read, the board's plan of a finished change and the record checkout check use it; every other reader reads the live path only. Its own package because the guard, api, selfreview and cmd/flow packages all call it.
**Considered:** reading only the live path everywhere — the board and the self-review bundle would lose every landed change; a flowd copy of the plan text — rejected by decision plan-text-ephemeral.

### A fix run after the drop restores the change directory

**ID:** fix-run-restores-change-dir
**Status:** active
**Chosen:** a fix run that starts after run 1's drop commit restores `spectre/changes/<name>/` from the drop commit's parent (`git checkout <drop-sha>^ -- spectre/changes/<name>`) as its fix-run planning commit, so every in-run reader keeps the live path; run 1's next pass drops it again with a fresh drop commit. Run 1's "already archived" re-run detection becomes "already dropped": the drop subject is in the branch log and the live directory is absent. The re-run base the panel and visual-verify gates read is the newest drop commit.
**Considered:** fix runs writing into a dropped directory — there is none; keeping a live-else-history fallback in every reader — the operator asked for the fallbacks to collapse.

### The pipeline-fix key scan counts landed changes from history

**ID:** pipeline-fix-key-from-history
**Status:** active
**Chosen:** `pfHighestK` takes the highest `<k>` over the live `spectre/changes/pipeline-fix-<name>-<k>` directories on the base plus every `chore(spectre): drop pipeline-fix-<name>-<k>` and `chore(spectre): archive pipeline-fix-<name>-<k>` subject in the base's log, so a key stays unique once landed directories leave the tree.
**Considered:** branch names — deleted branches vanish; a flowd counter — a new store for what the log already holds.

### Guard, script and stage names stay

**ID:** keep-guard-and-stage-names
**Status:** active
**Chosen:** the `commit-archive` and `check-archive-scope` guards, their `scripts/*.sh` shims, their output tokens and the `flow.sync-archive` / `flow.commit-archive` stage keys keep their names; only their behaviour and prose change.
**Considered:** renaming to `commit-drop` and friends — touches the stage registry, flowd's historical stage rows, every shim-sibling declaration and every caller for no behaviour; a later change may rename.

### The tip cleanup is made outside every task, after the review panel

**ID:** tip-cleanup-outside-tasks
**Status:** active
**Chosen:** the six tip-cleanup commits are made by the parent, not by implementer tasks: one `chore(spectre): remove the change archive` commit per repository, carrying `-m "Flow-Key: flow.sdd-tdd"`, made after the review panel's last round closes and before `flow.verify`. In agents a task commit could not carry the deletion: `check-task-commit-planning-paths.sh` runs from the main checkout's build and refuses a `Task-Id:` commit touching `spectre/changes/`. Made after the panel, the deletions (several thousand files) never enter the panel diff or its size cap, and the tooling the panel reviewed precedes them on every branch.
**Considered:** implementer tasks per repository — refused in agents, and the deletions would swamp the panel diff; a guard exemption for `changes/archive/` deletions — the run's guards execute the main checkout's build, so the exemption would not exist until after landing.
<!-- measured: grep -n 'planning := ' stats/internal/guard/taskcommitplanningpaths.go and grep -n 'check-task-commit-planning-paths.sh <canonical-worktree>' skills/flow/implement.md @ branch spectre/kan-1072-audit-whether-archived-spectre-change-artifacts -->

### flow-eval and gymie-frontend become declared peers

**ID:** peers-declare-flow-eval-gymie-frontend
**Status:** active
**Chosen:** agents' `spectre/peers` gains `flow-eval ../flow-eval` and `gymie-frontend ../gymie-frontend` in this change's first tooling task, so `kickoff-worktree.sh --from` can create those two repositories' worktrees and name them in the merge-order record; it exits 2 for a repository the file does not declare.
**Considered:** leaving those two repositories' archives — contradicts decision tip-cleanup-in-this-change.

### verbatim-moves.txt is appended by the parent

**ID:** verbatim-moves-appended-by-parent
**Status:** active
**Chosen:** the prose tasks reword run-loaded sentences on purpose. Each prose task's verify step runs `check-verbatim-moves.sh`; the implementer reports every FAIL line it intends, and the parent appends those lines to `<changeRoot>/verbatim-moves.txt`, carried by the `chore(spectre): plan <name> — before review` planning commit. An implementer never writes `spectre/changes/`, and the sentences cannot be known exactly at plan time.
**Considered:** listing every sentence in the plan-gate commit — the rewording is not written yet.

### This change's own landing still archives

**ID:** own-landing-archives
**Status:** active
**Chosen:** this change's own run 1 integrates with the main checkout's installed tooling, which still archives; after it lands, `spectre/changes/archive/kan-1072-audit-whether-archived-spectre-change-artifacts/` remains in agents and an `archive/<name>/link.md` remains in each satellite. The audit document records these and the umbrella follow-up's text names their deletion.
**Considered:** a second cleanup inside this change after landing — run 2 has no task slot, and the follow-up is already filed.

### The audit document carries no Jira key

**ID:** doc-carries-no-jira-key
**Status:** active
**Chosen:** the audit document states the umbrella follow-up's content; the KAN-1072 comment links the document and the follow-up's key, so the document needs no edit after filing.
**Considered:** writing the key into the document — a commit after the operator's review, for a link the comment already holds.

## Open questions

## Recommended approach

The change lands in four parts, in this order.

1. **The audit document** — written from fresh measurements, each cell with its command, holding
   round 2's verdict table as amended, the reader replacements, the spectre surface table and the
   umbrella follow-up's text (decisions audit-doc-location, one-umbrella-follow-up,
   doc-carries-no-jira-key).
2. **The tooling change** — TDD tasks in the agents worktree: a history reader for landed changes
   (decision landed-change-history-reader); the landing drop commit replacing the archive commit in
   integrate, the self-review loop and the pipeline-fix loop (decisions landing-drop-replaces-archive,
   drop-commit-subject, fix-run-restores-change-dir); every live-else-archive fallback collapsed to
   the live path; the unlanded-archive scans cut; the pipeline-fix key scan reading history
   (decision pipeline-fix-key-from-history); cleanup's check and the main-checkout hook reworded;
   then the shim comments and every prose site. Names stay (decision keep-guard-and-stage-names).
3. **The tip cleanup** — after the panel, the parent creates the five peer worktrees (agents already
   has one) and commits one archive deletion per repository (decisions tip-cleanup-in-this-change,
   tip-cleanup-outside-tasks, peers-declare-flow-eval-gymie-frontend).
4. **The follow-up** — at integrate run 1, the operator sees the findings and the umbrella task's
   text; then it is filed, labelled and transitioned, and KAN-1072 gets the comment.

Rejected alternatives and their reasons are the `**Considered:**` lines of the decisions above.

## Critical files

- `docs/audits/2026-10-31-spectre-archive-audit.md` — the audit document (new).
- `spectre/peers` — declares the two new peers.
- `stats/internal/landed/landed.go` — the history reader for landed changes (new).
- `stats/internal/guard/changeplan.go` — `ChangeDir`, the live-else-archive resolver.
- `stats/internal/guard/commitarchive.go`, `stats/internal/guard/archivescope.go` — the archive commit and its scope check; become the drop commit and its scope check.
- `stats/internal/guard/integratechange.go`, `stats/internal/guard/integratechange_commit.go`, `stats/internal/guard/integratechange_prepare.go` — run 1's sequence: canonical discovery, merge order, the archive step, archive-undo and the archived re-run.
- `stats/internal/guard/donewhenpaths.go`, `stats/internal/guard/untaskedcommits.go` — comments naming the archive commit.
- `stats/internal/selfreview/git.go` — the self-review bundle's archive reads and the reserved commit subjects.
- `stats/internal/guard/unlandedarchives.go` and its callers `planparse.go`, `taskrecords.go`, `planprovenance.go`, `verbatimmoves.go` — the unlanded-archive scans.
- `stats/internal/guard/planunchanged.go`, `unfinishedwork.go`, `visualverifydispatched.go`, `specroot.go`, `ownedcorpus.go` — archive fallbacks and archive-entry skips.
- `stats/internal/guard/pipelinefix.go` — the key scan and the land refusal.
- `stats/internal/guard/cleanupcomplete.go` — cleanup's change-directory row.
- `hooks/protect-main-checkout.py` — the unarchived-landing refusal and its message.
- `stats/cmd/flow/tasks.go`, `stats/cmd/flow/record.go`, `stats/cmd/flow/state.go`, `stats/cmd/flow/stage.go`, `stats/internal/api/stats.go` — CLI and board readers of archived changes.
- `scripts/*.sh` shims whose headers describe the archive.
- `skills/flow-contracts/finish-contract-run1.md`, `finish-contract-run2.md`, `finish-hand-fallbacks.md`, `pipeline.md`, `artifacts-registry.md`, `git-boundaries.md`, `plan-provenance-guard.md`, `state-file.md`, `known-bugs.md`, their `-rationale.md` siblings — the contracts naming the archive.
- `skills/flow/integrate.md`, `cleanup.md`, `implement.md`, `document-fix.md`, `brainstorm-planner.md`, `SKILL.md`, `stage-keys.md`, `verify-fix-loop.md`, `review-panel-late-fix.md`, `review-panel-experimental-slot.md` — flow prose naming the archive.
- `skills/flow-self-review/SKILL.md`, `skills/flow-status/SKILL.md`, `skills/flow-plan/SKILL.md` — standalone skills naming the archive.
- `README.md`, `rules/flow-manual-review.mdc`, `.flow/project.md`, `AGENTS.md` — repository docs naming the archive.

## Existing code to reuse

- `guard.ChangeDir` (`stats/internal/guard/changeplan.go`) — every worktree reader already routes through it; collapsing it collapses them.
- `commitArchive` (`stats/internal/guard/commitarchive.go`) — keeps its branch check, Done-when check, staging and output tokens; only the staged content and subject change.
- `checkArchiveScope` (`stats/internal/guard/archivescope.go`) — the staged-diff scope check, narrowed to deletions under the change's directories.
- `archBase` / `ArchivedRerunBase` (`stats/internal/guard/integratechange.go`) — the re-run base derivation, re-keyed on the drop subject.
- `cleanupcomplete.go`'s row four — already tests the live directory's presence; only its wording changes.
- `scripts/kickoff-worktree.sh --from` — creates each peer worktree, records its merge base and extends the merge-order record.
- `scripts/live-store-copy.sh` — the live check's store stand-in.
- `scripts/break-and-prove.sh` — each Go task's mutation proof.

## Verification

- Every Go task's targeted tests and the full `cd stats && go test ./...` run after the last group.
- `scripts/run-verify-list.sh <worktree> lint` on every prose task.
- `git grep -n -F -e 'spectre archive' -e 'changes/archive/' -- stats/ scripts/ hooks/ skills/ rules/ README.md .flow/project.md` prints only the history reader's legacy branch, its tests, the historical `-rationale.md` and `docs/` mentions the tasks keep, and the lessons and rejection-search sites the umbrella follow-up owns.
- After the tip cleanup, `git -C <worktree> ls-tree -d HEAD spectre/changes/archive` prints nothing in each of the six worktrees.

## Live check

Against stand-ins, before landing:

1. **The drop commit.** In a scratch clone of the kickoff worktree's branch
   (`git clone --branch spectre/kan-1072-audit-whether-archived-spectre-change-artifacts <worktree> <scratch>`),
   create `spectre/changes/live-check-1/` with a `tasks.md`, commit it, and run the worktree-built
   `scripts/commit-archive.sh <scratch> live-check-1` on a `spectre/live-check-1` branch. Record:
   the commit subject (`git log -1 --format=%s`), `git ls-tree -d HEAD spectre/changes/live-check-1`
   and `git ls-tree -d HEAD spectre/changes/archive/live-check-1`. Working: subject
   `chore(spectre): drop live-check-1`, both `ls-tree` lines empty. Not working: an
   `archive/live-check-1` tree, or the live directory still present.
2. **The board reads a landed plan from history.** `scripts/live-store-copy.sh up kan1072-live`
   builds flowd from the worktree against a copy of the dev store; request
   `GET http://127.0.0.1:4175/api/v1/stats/state-board?change=<a FINISHED agents change landed before this change>`
   (the change-scoped board, whose finished row reads its plan through `planTasks`,
   `stats/internal/api/stats.go`), once against the dev flowd on 4173 (the installed build, which
   reads the archive directory) and once against the copy on 4175 (the worktree build, which reads
   no archive directory and finds the plan through `stats/internal/landed`), and record both task
   counts. Working: the two counts are equal and non-zero. Not working: zero tasks from 4175.
   `scripts/live-store-copy.sh down kan1072-live` afterwards.
3. **Cleanup's check.** `scripts/check-cleanup-complete.sh` (worktree build) against the scratch
   clone of step 1 before and after the drop commit. Working: LEFTOVER naming
   `spectre/changes/live-check-1` before, no such row after.
