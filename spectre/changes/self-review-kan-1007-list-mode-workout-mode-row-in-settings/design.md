## Context

The eight findings of `proposal.md` are planned as one self-review change on
`self-review-kan-1007-list-mode-workout-mode-row-in-settings`, in its worktree
`/Users/tweety53/Projects/agents-worktrees/self-review-kan-1007-list-mode-workout-mode-row-in-settings`,
fix-base `origin/main` (`e111a883`); the branch already carries 54603278 and 7d8f01ae. No archived
change records a rejection of any remedy below.

Evidence at the branch tip:

- F1: `scripts/integrate-change.sh`'s `cmd_prepare` runs
  `git push -q origin "refs/remotes/origin/${BASES[fj]}:refs/heads/$fb"` and prints
  `BASE-CREATED:`; `skills/flow/SKILL.md` says kickoff creates the same missing base. **What still
  stops**' "Outward-facing or irreversible actions" bullet lists push unconditionally and its last
  bullet lists every integrate prompt.
- F2: `tokenize("git -C /x push -q 2>&1 | tail -1")` yields `… 'push', '-q', '2', '>&', '1', '|' …`;
  the push collector stops only at `&& || ; |`, so `push_refspecs` reads remote `2` and refspecs
  `>&`, `1`, finds no `spectre/` destination, and `untasked_push` returns before running the checker.
  The commit collector keeps `2 >& 1` as pathspec words, so a planning commit pathspec-limited to
  `spectre/changes/` reads as a trailer-less product commit. A scratch copy of
  `scripts/test-protect-main-checkout.sh` with cases 160–162 below failed all three on the unfixed
  hook (160 and 161 allowed, 162 denied) and passed the whole harness once one shared collector,
  stopping where the checkout/restore one stops, served all three verbs.
  <!-- measured: bash <scratch copy>/scripts/test-protect-main-checkout.sh @ 7d8f01ae → "3 failure(s)", then "all protect-main-checkout cases pass" with the git_args spike -->
- F3: the count `git rev-list --count origin/<base>..origin/<the branch it tracks>` appears only in
  `skills/flow-contracts/operator-prompts-auto-resolution.md`; no script or verify step makes it.
- F4: `runRecordDispatchBegin` (`stats/cmd/flow/record.go`) checks a panel key's round floor and
  slot, never the plan. In flowd, dispatch seq 36 is row id **7043**: key
  `task-8-implementer-retry`, `agentId` inline, commit `bf6a0acaeae61a94a0545e02fd8a52660e3df2e8`,
  stamped `2026-10-22T15:21:37+03:00`–`15:23:29+03:00`. Its back-fill took both instants from the
  commit (author `12:18:29Z`, committer `12:23:29Z`); the parent's transcript
  `587a769e-8249-48a8-9769-a13c7ccce16e.jsonl` shows the work starting after the fix-2 verifier
  report, `2026-10-22T11:59:47Z`. `flowd restamp-dispatch <id> <parent transcript>` re-attributes
  over the row's stored window (`harvest.InlineDispatchPatch`), so the window is set first with
  `flowd retime-dispatch`, which also rewrites `commit_sha` (an empty commit would null it).
  <!-- measured: curl -s http://127.0.0.1:4173/api/v1/records/gymie-7c1f238a/kan-1007-list-mode-workout-mode-row-in-settings @ 2026-10-22 → id 7043, seq 36, startedAt 15:21:37+03:00 -->
- F5, F6, F8: the paragraphs the findings name read as the findings quote them.
- F7: `checkSidecarReach` takes exactly one argument and reads only `fields[0]` of a sidecar line.

Repair: F4's row 7043 is repaired in Task 4 by the parent. 889037cb (F2) is already on origin and
`git push` cannot unpush it; its missing trailer is a history fact, not a record a command can
repair. F3's folded squash landed with KAN-1010's archive; rewriting the base's published history
would be irreversible and outward-facing, so it stays as landed. No other defect wrote a wrong record.

## Live check

Task 4's repair writes the flow store. After the parent runs it,
`curl -s http://127.0.0.1:4173/api/v1/records/gymie-7c1f238a/kan-1007-list-mode-workout-mode-row-in-settings`
shows row 7043 with `startedAt` `2026-10-22T14:59:47+03:00`, `endedAt` `2026-10-22T15:23:29+03:00`,
`commitSha` `bf6a0aca…`, and `metrics.tokens.main.output` different from 1047. Not working looks
like `startedAt` still `15:21:37+03:00`, or `output` still 1047 (the restamp never ran or found no
share). Every other change edits a hook, a guard, a CLI refusal or skill prose — no runtime.

## Decisions

### Base creation is carved out of What still stops

**ID:** base-create-carve-out
**Status:** active
**Chosen:** the "Outward-facing or irreversible actions" bullet gains one sentence exempting the
creation of a `--base` a repository's origin lacks, from that repository's default branch, by
`kickoff-worktree.sh` and integrate's `prepare` — it adds a branch whose tip is the default
branch's own, overwrites nothing, and is what the operator's `--base` asked for.
**Considered:** reverting 54603278 to a stop — re-adds the question 54603278 removed on the
operator's instruction; a carve-out in `integrate.md` only — the contract file would still say
every push is confirmed.

### One argument walk for every git verb the hook reads

**ID:** git-args-stop-at-redirects
**Status:** active
**Chosen:** a module-level `git_args(tokens, start)` returns the words after the verb up to the
first separator, `)` or redirect token (a word starting `<`, `>` or `&>`), dropping a bare digit
just before the redirect (the fd of `2>&1`); the checkout/restore, commit and push collectors all
call it. The redirect test is `t.startswith(("<", ">", "&>"))`, never `t[0]`, so an empty word
(`-m ''`) cannot raise.
**Considered:** skipping each redirect and its target and continuing — no git command in the
pipeline puts arguments after a redirect, and the checkout walk already stops; patching only the
push loop — the commit loop has the same misread (case 162).

### A Go guard compares each recorded base with the branch it tracks

**ID:** check-base-behind-guard
**Status:** active
**Chosen:** `check-base-behind.sh <worktree>...`, Go in `stats/internal/guard/basebehind.go`
(the project's rule for new guards). Per worktree: the recorded base is
`branch.<current branch>.flowBase` (`recordedBase`); the tracked branch is `origin/HEAD`'s branch;
the guard runs `git fetch -q origin`, its failure ignored. Lines, exit 0 unless a `BASE-BEHIND`:
`BASE-CURRENT: <wt> — no recorded base`; `BASE-CURRENT: <wt> — <base> is the tracked branch`;
`BASE-CURRENT: <wt> — origin/<base> holds every commit of origin/<tracked>`; or
`BASE-BEHIND: <wt> — origin/<base> lacks <n> commits of origin/<tracked>; sync the base, never merge <tracked> into the change branch (kan-1007):`
followed by two two-space-indented command lines —
`T=$(git -C <wt> merge-tree --write-tree origin/<base> origin/<tracked>) && C=$(git -C <wt> commit-tree "$T" -p origin/<base> -p origin/<tracked> -m 'Merge <tracked> into <base>') && git -C <wt> push origin "$C:refs/heads/<base>"`
and `git -C <wt> fetch -q origin && sync-onto-base.sh <wt> origin/<base> <merge-base>`, `<merge-base>`
the literal `git merge-base HEAD origin/<base>` sha — exit 1. Exit 2: no argument, a path that is not
a git worktree, no `origin/HEAD`, or a failing `rev-list`/`merge-base`, cause on stderr. **Verify**
calls it once, on the run's first `flow.verify`, over the resolved worktree set; exit 1 is relayed
and its commands run — the push confirmed per **What still stops** — before Verify re-runs from its
`begin` mark; exit 2 is named in the handoff and Verify continues.
**Considered:** a bash script — the project's `.flow/project.md` requires every new guard in Go;
printing `git merge` in a temporary worktree — `merge-tree --write-tree` + `commit-tree` sync the
base with no checkout and stop on a conflict (merge-tree exits 1); running the push itself — a push
of a shared base is outward-facing and stays confirmed.

### A panel round refuses to open over an unticked task

**ID:** panel-refuses-unticked-tasks
**Status:** active
**Chosen:** `runRecordDispatchBegin` refuses (exit 2, nothing sent) a reviewer key
`records.PanelDispatchRound` accepts, `-started-at` unset, while the change's `tasks.md`
(`planPath(f.dir, f.change)`) holds any unticked column-0 task line outside a fence; an unreadable
plan skips the check. The message names the unticked task ids and `close-task.sh`.
**Considered:** unticked *appended* tasks only, measured against the first `flow tasks count`
observation — needs a store read, and every plan-time task is ticked before round 1 by
construction (`skills/flow/implement.md`), so any unticked task at a panel begin is the same
ordering defect; refusing in the panel's prose only — kan-1007's parent had the prose and still
opened round 4.

### Dispatch 7043 is retimed, then restamped

**ID:** retime-then-restamp-7043
**Status:** active
**Chosen:** the parent runs `flowd retime-dispatch 7043 2026-10-22T11:59:47Z 2026-10-22T12:23:29Z bf6a0acaeae61a94a0545e02fd8a52660e3df2e8`,
then `flowd restamp-dispatch 7043 <parent transcript>`. The window overlaps inline rows 7032/7033
(`verify-fix-3-*`, `12:21:37Z`–`12:23:02Z`); the attributor apportions overlapping inline windows.
**Considered:** `restamp-dispatch 36` as the finding wrote it — 36 is the row's seq, and the command
takes the store id; restamp alone — re-attributes over the wrong 2-minute window; ending the window
at the author date `12:18:29Z` — the commit was amended at `12:23:29Z`, inside task 8's work.

### A 1dp leftover is a leftover

**ID:** leftover-1dp-not-clean
**Status:** active
**Chosen:** one sentence appended to **The implementer measures before it hands back**: a leftover
of any size, a 1dp offset included, is closed or named with its cause, never reported clean.
**Considered:** a numeric threshold — every threshold is a tolerance the determinism rule forbids
widening, and the frame is the spec.

### Plan-time capture tasks measure before they commit

**ID:** plan-time-capture-measured
**Status:** active
**Chosen:** a section D blockquote, after **Write a capture for every frame on the change's frame
list**, requiring a capture task that baselines against a mockup to carry a step running
`compose-mockup-frames.sh` and `measure-visual-properties.sh` and fixing until clean or each
leftover is named with its cause, citing the fix-loop rule rather than restating its commands.
**Considered:** a guard over `tasks.md` requiring the step — no plan parser reads step text today,
and the rule is a planning instruction the planner reads in the same file.

### A sidecar line's frame must exist

**ID:** sidecar-frame-exists
**Status:** active
**Chosen:** `check-sidecar-reach.sh <capture root> [<mockups root>]`; with the second argument, a
line whose second field `<frame id>` has no `<mockups root>/<frame id>.png` prints
`<sidecar>:<line>: <name> — frame <frame id> has no <mockups root>/<frame id>.png` (exit 1);
a second argument that is not a directory is exit 2. `visual-verify.md` passes `<mockups root>`
whenever `mockups` is declared — the only call site, which already runs only then.
**Considered:** a separate guard — one more call and one more harness for one more check over the
same lines; making the root mandatory — the guard's own tests and any one-root caller would break
for nothing.

### A mockups-only round composes instead of verifying

**ID:** mockups-only-round-composes
**Status:** active
**Chosen:** in step 4, a round whose `git diff <round base>..HEAD` matches no `ui paths` entry in
any worktree and touches only paths under the mockups root dispatches no `visual-verify` verifier:
the parent runs `resolve-visual-screenshots.sh … | compose-mockup-frames.sh …` (the verifier's step
command) for each sidecar naming a changed `<frame id>.png`, then `measure-visual-properties.sh
--props bands,seams` on each pair; a leftover opens the next round; clean composites close the
round and need no `-full` run, the last verifier's report standing for every path outside the
mockups root. The `verify` re-run rule is unchanged.
**Considered:** also skipping `verify` for such a round — the finding measured the visual suite's
18m28s, and whether a project's lint or tests read its mockups root is not knowable here; a
scoped verifier — `specs` prints nothing for a mockups-only diff, so it ran the whole suite.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** flow-medium — four verdicts move (the main-checkout hook's push and commit reading, a
new guard, a CLI refusal, the sidecar guard's second check), besides contract prose.
**Considered:** flow-low — fits prose-only or mechanical changes, not changed guard verdicts.

### Fixer groups

**ID:** fixer-groups
**Status:** active
**Chosen:** one fixer dispatch, key `self-review-kan-1007-list-mode-workout-mode-row-in-settings-fix`,
Tasks 1–8 in order — eight small tasks in one repository, three of them one-paragraph prose; Tasks 1 and 3 share `operator-prompts-auto-resolution.md` and Tasks 4, 5 and 8 share `verify-fix-loop.md`, ordered by `**After:**`.
**Considered:** one dispatch per task — eight cold starts; two groups (code, then prose) — fixer
groups run in series, so a second cold start buys nothing.

## Open questions
