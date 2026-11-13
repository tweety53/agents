---
name: flow-prompt-tuning
description: Apply the generic implementer and reviewer prompt updates the self-review's prompt-tuning angle saved, after a replay on each case's repro data proves them, and remove an applied update a replay no longer needs; record every case in docs/prompt-tuning/LEDGER.md, remove its folder and land the result. Standalone, not a pipeline stage. Use for /flow-prompt-tuning.
allowed-tools: Bash(git:*), Bash(flow:*), Bash(scripts/flow-guard check-prompt-tuning-cases:*), Bash(scripts/flow-guard project-get:*)
license: MIT
compatibility: Runs from the agents repository's main checkout, which stays read-only; needs the flow CLI and the Agent tool.
---

Walk the prompt-tuning cases the self-review's angle 8, `flow-prompt-tuning`, saved —
step 2a of **Workflow** (`skills/flow-self-review/SKILL.md`) is canonical for the case folder and its fields — under
`<agents repo>/docs/prompt-tuning/<case-id>/`, prove each proposed prompt update by a replay
on the case's repro data, apply the updates that pass, remove the applied updates a replay no
longer needs, record every case in
`<agents repo>/docs/prompt-tuning/LEDGER.md`, delete the case folders and land the result.

**This is a standalone command, not a pipeline stage.** It takes no argument, writes no state
file and marks no `flow stage` call. An argument is reported rather than ignored.

**Announce at start:** "Using flow-prompt-tuning."

## Workspace

The agents main checkout stays read-only. All work happens in the worktree
`<agents repo>-worktrees/prompt-tuning-<date>`, on a branch of the same name off
`origin/<default-branch>`, `<date>` being the UTC date `date -u +%Y-%m-%d` prints:

```bash
git -C <agents repo> fetch origin
git -C <agents repo> worktree add -b prompt-tuning-<date> <agents repo>-worktrees/prompt-tuning-<date> origin/<default-branch>
```

`<worktree>` below is that worktree.

## Every dispatch

Every dispatch below — replay and grader — runs on opus at low effort, passed explicitly as
`subagent_type: "flow-low"` and `model: "opus"`, and its prompt carries these two sentences
verbatim:

> Before anything else, read `~/.claude/rules/agent-baseline.md` and follow it for this whole task.
> Include this instruction verbatim in any prompt you write for another agent.

Every dispatch is one-shot: no `SendMessage` to it once it returns. A replay dispatch never writes
in a repository.

## Workflow

### 1. Walk

List `<worktree>/docs/prompt-tuning/*/`. None: go straight to step 5a, and with no candidate there
either print `none pending`, remove the worktree and stop.
Run `flow-guard check-prompt-tuning-cases <agents repo>-worktrees/prompt-tuning-<date>/docs/prompt-tuning`; a case it names in a red line is reported
by its folder and left untouched — never replayed, applied, recorded or deleted.

### 2. Group

Cases with the same `target` and the same defect class become one group with one merged generic
`after`. The merge states the principle once — one idea per update. A group's evidence is every
case in it.

### 3. Re-anchor

If a group's `before` no longer occurs in the target at `<worktree>`'s HEAD, re-locate it. If the
gap is gone, the group is rejected as `already covered`, with the line now covering it quoted.
Otherwise the update is rewritten against the current text and re-checked by the genericity
rules of `flow-guard check-prompt-tuning-cases`: the `after` names no source change, no Jira key, no
`kan-<digits>` key and no path or basename from the case's `bad.diff`/`good.diff` headers.

### 4. Replay

For every case of the group, one run, carrying the target's text with the update applied. No run
carries the target's current text. A run that comes back `avoided` proves the updated text
produced the right behaviour on this input, not that the current text would have failed there.

- **`reviewer` case** — a reviewer dispatch over `bad.diff` (a miss) or `good.diff` plus the
  finding's location (a false positive), with the target prompt as its instructions. A false
  positive's dispatch also returns one line on that location: the finding it raises there, or
  `no finding at <location>`.
- **verifier case** — a `reviewer` case whose `target` is `skills/flow/visual-verify-verifier.md`
  or `skills/flow/visual-verify.md` is replayed instead by a verifier dispatch judging the case's
  `capture.png` against its `mockup.png`, with the target prompt as its instructions. A false
  positive's dispatch is also given the element `finding.md` names and returns the same one line
  on that element.
- **`implementer` case** — an implementer dispatch given `task.md` and the pre-change code,
  writing in a scratch directory under the session's scratchpad, never in a repository. The
  pre-change code is the tree at `source.fix_sha`'s parent in the checkout of `source.repo`; a
  case whose repository the session cannot find on this machine is reported by its folder and
  left untouched, as a red case of step 1 is.

A replay dispatch returns what its judge reads: an implementer its diff, a reviewer or verifier
its findings with their locations.

**Verdict.** Each run ends in one of three verdicts, judged against the case's `finding.md`:

- `avoided` — the output reaches the finding's location and the defect is absent: on a miss the
  finding is raised; on a false positive it is not raised, or is attributed elsewhere.
- `present` — the defect is there. The case is rejected `no effect`, with no second run.
- `ambiguous` — the output never reaches the finding's location, so neither answer can be
  quoted. The case gets one second run, judged the same way: `present` on it rejects the case
  `no effect`, and `ambiguous` rejects it `not replayable`.

A group passes when every one of its cases is `avoided`.

**Who judges.**

- **`implementer` case — a grader dispatch, one per replay run.** Its prompt carries the
  replay's returned diff and the case's `finding.md`, `bad.diff` and `good.diff`, and nothing
  else: nothing saying an update was applied, what the update says, or which verdict is hoped
  for. The grader answers `avoided` when the diff touches the code `finding.md`'s
  `**Location:**` names and does not carry the defect `finding.md` describes — `bad.diff` is the
  defect's shape, and `good.diff` one correct shape, not the only one; `present` when the diff
  carries the defect; `ambiguous` when no hunk touches that code. It returns exactly one of the
  three verdicts and one quoted line of the diff as evidence — for `ambiguous`, the statement
  that no hunk touches the location. The main session records the verdict and the line as
  returned and never overrules it.
- **`reviewer` or verifier case — the main session, with no dispatch.** The finding is raised
  when the output names the defect `finding.md` describes at the file its `**Location:**` names
  — for a verifier case, at the element it names — whatever the line. The session records the
  verdict per case in the report, quoting both lines, the output's and `finding.md`'s. An output
  it cannot quote a deciding line from is `ambiguous`.
  A false positive's one line on its location is its deciding line: one reading `no finding` is
  `avoided`.

### 5. Apply

Apply each passing group's update to its target in `<worktree>`, one commit per group. A
non-`generic` group whose `rules/` target does not exist creates it as a new file of its own kind,
a prompt-tuning rule file: a header stating the file holds rules `/flow-prompt-tuning` applied for
scope `<scope>`, each proven by a replay, and that a project receives it only by naming it in
its `<project>/.flow/project.md` `## standards`; then the update as its first rule. A prompt-tuning update
never goes into a hand-written standard such as `rules/kotlin-backend-development-standard.mdc`,
nor reuses its header. Then run `<agents repo>/.flow/project.md`'s `## lint` lines for the touched files; a red
line is fixed before the next step.

### 5a. Prune

**A line this command applied stays only while a replay still needs it** — without this step a
target only grows, since every run adds proven lines and none is ever proven again.

The candidates are `LEDGER.md`'s `applied <sha>` entries with no later `removed` entry for the same
case id, whose added lines — the `+` lines `git -C <worktree> show <sha> -- <target>` prints — still
occur verbatim in the target at `<worktree>`'s HEAD. Take at most five per run, the case whose
newest ledger entry is oldest first. A case this run applied has no ledger entry yet and is not a
candidate.

For each candidate, read the case's files from the commit before the one that removed its folder:
`git -C <worktree> log --diff-filter=D -1 --format=%H -- <agents repo>-worktrees/prompt-tuning-<date>/docs/prompt-tuning/<case-id>/`
names the removing commit `<rm>`, and `git -C <worktree> show <rm>^:docs/prompt-tuning/<case-id>/<file>`
prints each file, written under the session's scratchpad. Replay the case twice as step 4 does,
both runs carrying the target's text with those added lines removed; step 4's verdicts and judges
apply unchanged. Two runs, because one `avoided` can be chance and a line removed wrongly brings
its defect class back.

- **Both runs `avoided`** — the target no longer needs the lines: delete them from the target in
  `<worktree>`, one commit per case, then run the `## lint` lines as step 5 does.
- **Anything else** — the lines stay.

Each checked case gets one ledger entry in step 6, `kept` or `removed <sha>`.

### 6. Ledger and cleanup

Append one entry per case to `<agents repo>/docs/prompt-tuning/LEDGER.md` and `git rm -r` its folder. Rejected
cases are recorded too, so the same case is never re-mined:

```markdown
- <date> · `<case-id>` · <type> · `<target>` · <defect class> — <one-line summary of the update> — applied <sha>
- <date> · `<case-id>` · <type> · `<target>` · <defect class> — <one-line summary of the update> — rejected: <reason>
- <date> · `<case-id>` · <type> · `<target>` · <defect class> — <one-line summary of the update> — kept
- <date> · `<case-id>` · <type> · `<target>` · <defect class> — <one-line summary of the update> — removed <sha>
```

The last two are step 5a's: the case's `applied` entry copied under this run's date, ending `kept`,
or `removed <sha>` with the commit that deleted its lines.

Before a folder's `git rm`, every applied case is recorded in the stats store, which feeds the
Review dynamics view's prompt-change markers:

```bash
flow prompt-tuning record <case-dir> -sha <the applying commit's sha>
```

A non-zero exit is reported with its stderr and the step continues: the store is never a gate.

For an applied case, `<sha>` is the commit that applied it. The ledger entries and the folder
removals go in one commit, `docs(prompt-tuning): ledger <date>`, after the update commits.
Before it is committed, `flow-guard check-prompt-tuning-cases` over the same root prints no line for `LEDGER.md` and none for a folder this run replayed; a line for a red case step 1 left untouched is expected and stays.

### 7. Land

Land the branch by `<agents repo>`'s `## default landing route`
(`flow-guard project-get <agents repo> 'default landing route'`), without asking.
Then remove the worktree.

## Report

One line per group: target, defect class, applied or rejected with its reason. One line per
case step 5a checked: target, defect class, kept or removed. A group that
created a `rules/` file says the file reaches a project only once that project names it in
`## standards`.

## Guardrails

- No review panel runs over the prompt edits: the replay and the `## lint` guards are the gate.
- Never edit the main checkout.
- Never apply an update whose group did not pass the replay.
- Never remove an applied line on fewer than two `avoided` runs without it.
