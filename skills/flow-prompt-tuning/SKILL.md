---
name: flow-prompt-tuning
description: Apply the generic implementer and reviewer prompt updates the self-review's prompt-tuning angle saved, after an A/B replay on each case's repro data proves them; record every case in docs/prompt-tuning/LEDGER.md, remove its folder and land the result. Standalone, not a pipeline stage. Use for /flow-prompt-tuning.
allowed-tools: Bash(git:*), Bash(flow:*), Bash(scripts/check-prompt-tuning-cases.sh:*), Bash(scripts/project-get.sh:*)
license: MIT
compatibility: Runs from the agents repository's main checkout, which stays read-only; needs the flow CLI and the Agent tool.
---

Walk the prompt-tuning cases the self-review's angle 8, `flow-prompt-tuning`, saved —
step 2a of **Workflow** (`skills/flow-self-review/SKILL.md`) is canonical for the case folder and its fields — under
`<agents repo>/docs/prompt-tuning/<case-id>/`, prove each proposed prompt update by an A/B replay
on the case's repro data, apply the updates that pass, record every case in
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

Every dispatch below runs on **opus**, passed explicitly as `model: "opus"`, and its prompt
carries these two sentences verbatim:

> Before anything else, read `~/.claude/rules/agent-baseline.md` and follow it for this whole task.
> Include this instruction verbatim in any prompt you write for another agent.

Every dispatch is one-shot: no `SendMessage` to it once it returns. A replay dispatch never writes
in a repository.

## Workflow

### 1. Walk

List `<worktree>/docs/prompt-tuning/*/`. None: print `none pending`, remove the worktree and stop.
Run `check-prompt-tuning-cases.sh <agents repo>-worktrees/prompt-tuning-<date>/docs/prompt-tuning`; a case it names in a red line is reported
by its folder and left untouched — never replayed, applied, recorded or deleted.

### 2. Group

Cases with the same `target` and the same defect class become one group with one merged generic
`after`. The merge states the principle once — one idea per update. A group's evidence is every
case in it.

### 3. Re-anchor

If a group's `before` no longer occurs in the target at `<worktree>`'s HEAD, re-locate it. If the
gap is gone, the group is rejected as `already covered`, with the line now covering it quoted.
Otherwise the update is rewritten against the current text and re-checked by the genericity
rules of `check-prompt-tuning-cases.sh`: the `after` names no source change, no Jira key, no
`kan-<digits>` key and no path or basename from the case's `bad.diff`/`good.diff` headers.

### 4. Replay (A/B)

For every case of the group, 3 runs per arm. Arm A carries the target's current text, arm B the
text with the update applied.

- **`reviewer` case** — a reviewer dispatch over `bad.diff` (a miss) or `good.diff` plus the
  finding's location (a false positive), with the target prompt as its instructions.
- **`implementer` case** — an implementer dispatch given `task.md` and the pre-change code,
  writing in a scratch directory under the session's scratchpad, never in a repository.
- **Grader** — a grader dispatch judges each run against the case's `finding.md`: was the defect
  present (or raised), yes or no, with one line of evidence.

**Pass rule:** arm A shows the problem in at least 2 of 3 runs, and arm B avoids it in at least 2
of 3. A shows it in fewer than 2: the case is rejected `not reproducible`. B fails: rejected `no
effect`. A group passes when every one of its cases passes.

### 5. Apply

Apply each passing group's update to its target in `<worktree>`, one commit per group. A
non-`generic` group whose `rules/` target does not exist creates it as a new file of its own kind,
a prompt-tuning rule file: a header stating the file holds rules `/flow-prompt-tuning` applied for
scope `<scope>`, each proven by an A/B replay, and that a project receives it only by naming it in
its `<project>/.flow/project.md` `## standards`; then the update as its first rule. A prompt-tuning update
never goes into a hand-written standard such as `rules/kotlin-backend-development-standard.mdc`,
nor reuses its header. Then run `<agents repo>/.flow/project.md`'s `## lint` lines for the touched files; a red
line is fixed before the next step.

### 6. Ledger and cleanup

Append one entry per case to `<agents repo>/docs/prompt-tuning/LEDGER.md` and `git rm -r` its folder. Rejected
cases are recorded too, so the same case is never re-mined:

```markdown
- <date> · `<case-id>` · <type> · `<target>` · <defect class> — <one-line summary of the update> — applied <sha>
- <date> · `<case-id>` · <type> · `<target>` · <defect class> — <one-line summary of the update> — rejected: <reason>
```

For an applied case, `<sha>` is the commit that applied it. The ledger entries and the folder
removals go in one commit, `docs(prompt-tuning): ledger <date>`, after the update commits.
`check-prompt-tuning-cases.sh` must exit 0 on the result before it is committed.

### 7. Land

Land the branch by `<agents repo>`'s `## default landing route`
(`project-get.sh <agents repo> 'default landing route'`), without asking.
Then remove the worktree.

## Report

One line per group: target, defect class, applied or rejected with its reason. A group that
created a `rules/` file says the file reaches a project only once that project names it in
`## standards`.

## Guardrails

- No review panel runs over the prompt edits: the replay and the `## lint` guards are the gate.
- Never edit the main checkout.
- Never apply an update whose group did not pass the replay.
