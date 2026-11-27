# The gated per-task reviewer

Loaded during `flow.implement` by the load directive of **4. Execute (SDD + TDD)**
(`skills/flow/implement.md`), only when the decision's `groups` holds more than one group or any
task is marked `**Size:** big`. Every section name below without a path is a section of
`skills/flow/implement.md`.

**The gated per-task reviewer.** One combined review per task — spec compliance and
code quality together — but **one dispatch per batch of returned tasks, never one per task**,
the discipline **Bundled dispatch** (`skills/flow/review-panel.md`) applies to panel rounds.
**The bundle is a batch of weight 5**: returned tasks not yet reviewed accumulate, in plan order, and
at the boundary where their weight reaches 5 — a `**Size:** big` task weighing 3, every other task 1 —
or where no implementer of the repository is left in flight or queued, every one of them, of one repository, goes out in one reviewer Agent call beside the implementers the boundary launches, on the
`gated-reviewer` role's pair (**Model and effort**, `skills/flow-contracts/model-policy.md`), its prompt opening with the
line `Role: gated-reviewer`. A fix re-review bundle runs on its own row instead, its prompt opening
with `Role: gated-re-reviewer` (`skills/flow/gated-review-fix.md`). Each task inside the bundle keeps its own pass: its own
commit-range diff `git diff <task-sha>^..<task-sha>` — a real commit diff, never a snapshot of
the working tree, which the next implementer is editing — its own verdict and its own report
file. Record the bundle as one `dispatches` row (`-role gated-reviewer`, `-key task-<n+n+n>-reviewer`
with the task ids `+`-joined in plan order, the same convention as `panel-<round>-<slot|slot+slot>`;
`-task <n>` only on a one-task bundle, omitted otherwise) and close it with **`-outcome clean`
when every pass is clean, `-outcome fix` when any pass is `fix`**; each pass's own verdict is
its report file's `## Verdict`, so per-task review yield stays measurable.
The reviewer writes every finding its passes raise to the bundle's findings file,
`<abs-worktree>/.superpowers/sdd/findings-gated-<n+n+n>.jsonl`, one line per finding and no ref
(`skills/flow/gated-reviewer-brief.md`), and the parent records them all in one call,
`flow record finding -change <name> -slot gated-reviewer -round 0 -dispatch-seq <the bundle's seq>
-status open -from
<that file>`, which gives each line the change's next free ref, so the run's Findings count holds
what the gated reviewer caught; the panel's rendered record skips `gated-reviewer` rows.
**A returning bundle is one Bash call**, its verdicts read from its final message — one line per
pass — never its report bodies: the findings call above — riding the
boundary call (**A group boundary is one `ENDED`, one Bash call and one launch message**) when the next group's implementer returns in the same wake. The hook ends its
row `clean` or `fix` from those lines.
**A pass whose findings are all Minor is `clean`** — no re-review; its Minors stay open findings for the stage-end fixer
(`skills/flow/gated-review-fix.md`). A Minor no change to the tree can resolve is named,
one clause, in the parent's output, and recorded `withdrawn <that clause>`. A `fix` pass
sends its Minors to the same fixer. **A mixed-verdict bundle is handled per task**: every `fix` task waits for the stage-end fix path
on its own sha, independently of its bundle-mates.

**Load `skills/flow/gated-review-fix.md` only when**, once the last group's guard has passed and
every gated bundle has returned, a gated pass closed `fix` or left a Minor open — it carries the stage-end fix path.

**Each pass is a primary pass, scoped to its task.** The bundle's stub names
`skills/flow/primary-reviewer-prompt.md` once, every placeholder but the two per-task ones
filled per the placeholder table of `skills/flow/review-panel.md`.
Each **PASS task-`<n>`** section gives that pass's values for the two, which are its Scope
fields:

- `[DIFF_PATH]` — `<abs-worktree>/.superpowers/sdd/task-<n>.diff`, the parent's
  `git diff <task-sha>^..<task-sha>` written there before the Agent call; on a fix re-review,
  `task-<n>-fix-<k>.diff`, of the range `skills/flow/gated-review-fix.md` names;
- `[ARTIFACT_PATHS]` — the absolute paths of the change's `proposal.md`, `design.md` and
  `tasks.md`, as the table states them.

**Every gated reviewer bundle's prompt is a stub too; its brief is
`skills/flow/gated-reviewer-brief.md`**, read in full first. The prompt is these lines, in this order,
then one `PASS task-<n>` block per task of the group in plan order, each carrying its three Scope
values and its own REPORT FILE line:

```text
Role: gated-reviewer
Before anything else, read `~/.claude/rules/agent-baseline.md` and follow it for this whole task.
Include this instruction verbatim in any prompt you write for another agent.
Report tersely: findings first, bullets over prose, no preamble, no recap. Full prose in whatever you write to disk.
Key: task-<n+n+n>-reviewer
Read <skill-dir>/gated-reviewer-brief.md in full first — it is your brief. Each pass below reviews per <skill-dir>/primary-reviewer-prompt.md, [CALIBRATION_PATH] being <skill-dir>/reviewer-calibration.md and its two Scope placeholders the values its PASS block names.
FINDINGS FILE: <abs-worktree>/.superpowers/sdd/findings-gated-<n+n+n>.jsonl
PASS task-<n>:
[DIFF_PATH] <abs-worktree>/.superpowers/sdd/task-<n>.diff
[ARTIFACT_PATHS] <changeRoot>/proposal.md <changeRoot>/design.md <changeRoot>/tasks.md
REPORT FILE: <abs-worktree>/.superpowers/sdd/reviewer-report-task-<n>.md
```

A fix re-review's `Role:` line is `gated-re-reviewer`, its `Key:` `task-<n+n+n>-reviewer-fix-<k>` and
its PASS values those `skills/flow/gated-review-fix.md` names.
