# kan-1040-copyable-next-command

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode

**Goal:** print every handoff's next command alone in a copyable fence, and keep a creating run
going after it files a requested Jira issue.

**Architecture:** prose edits to the handoff contract, its templates, and Jira **Resolution**.

**Spec:** `design.md` beside this file.

## Global constraints

- Work in `/Users/tweety53/Projects/agents-worktrees/kan-1040-copyable-next-command`, branch
  `spectre/kan-1040-copyable-next-command`; paths below are relative to its root. Never touch the
  main checkout, never create a worktree, never write `spectre/changes/`.
- A template block whose fence now holds an inner triple-backtick fence opens and closes with four
  backticks (a four-backtick `text` fence), so the inner fence is part of the template.

- [ ] 1. Fence the handoff's next command
  - [ ] **Step 1: Rules** (`skills/flow-contracts/pipeline.md`, **Handoff output**). In the shape
    template, replace the bare `/flow <name>` under `Next:` with the command alone in its own
    triple-backtick fence and make the template's own fence four backticks. In the first bullet replace
    "The next command and `/clear` stay bare lines, per the next rule." with "`/clear` stays a bare
    line, and the next command is fenced, per the next rule." Replace the bullet "**The next command
    is the last line** — bare, copy-pasteable, with no prose after it." with "**The next command is
    the last thing printed, alone in its own fenced code block** — a plain triple-backtick fence, no
    language tag, holding the command and nothing else, with no prose after it, so a client with a
    code-block copy button copies it in one tap (kan-1040)." keeping its "See **Handoff output** …
    for why." sentence. In the `/clear` bullet, "on the line above the next command" becomes "on the
    line above the next command's fence" and "the next command stays the last line" becomes "the
    next command's fence stays last".
  - [ ] **Step 2: Rationale** (`skills/flow-contracts/pipeline-rationale.md`, **The block each state
    renders**). The bullet's head becomes "**The next command is the last thing printed, alone in its
    own fence**"; append: "The fence is what the Claude mobile and desktop apps put a copy button on —
    a bare line has to be selected by hand (kan-1040); it holds the command alone so the copy is
    exactly what is pasted, and `/clear` stays outside it because the operator types it."
  - [ ] **Step 3: Templates.** In every block below, the `/flow <name>` line under `Next:` moves into
    its own triple-backtick fence (a `/clear` line above it stays bare, outside the fence) and the
    template's outer fence becomes four backticks: `skills/flow-contracts/handoff-blocks.md` (three
    blocks, at the `Next:` lines), `skills/flow/brainstorm.md` (**The `STARTED` handoff block**),
    `skills/flow/integrate.md`, `skills/flow/verify-and-handoff.md`, `skills/flow/cleanup.md`,
    `skills/flow-plan/SKILL.md`. In `skills/flow-plan/SKILL.md` "The last line is the next command,
    bare:" becomes "The handoff ends on the next command, alone in its own fence (**Handoff output**,
    `skills/flow-contracts/pipeline.md`):". In `skills/flow-contracts/finish-contract-run1.md`
    "its handoff's last line is `/flow <name>` again" becomes "its handoff ends on `/flow <name>`
    again".
  - [ ] **Step 4: Verify.** `scripts/check-references.sh && scripts/check-markdown-integrity.sh`
    exit 0, then the absence sweep
    `git grep -n -F -e 'The next command is the last line' -e 'stay bare lines' -e 'The last line is the next command' -- skills`
    prints nothing.

**Build:** green
**Files:** `skills/flow-contracts/pipeline.md`, `skills/flow-contracts/pipeline-rationale.md`, `skills/flow-contracts/handoff-blocks.md`, `skills/flow/brainstorm.md`, `skills/flow/integrate.md`, `skills/flow/verify-and-handoff.md`, `skills/flow/cleanup.md`, `skills/flow-plan/SKILL.md`, `skills/flow-contracts/finish-contract-run1.md`
**Tests:** **none** — prose contracts and templates; the guards and the absence sweep cover them.
**Regression:** none — no test is declared.
**Commit:** feat(flow-contracts): fence the handoff's next command for one-tap copy
**After:** none

**Decision:** next-command-fenced
**Decision:** blocks-still-rendered

- [ ] 2. Continue planning after filing a requested issue
  - [ ] **Step 1: Edit** `skills/flow-contracts/jira-integration.md` **Resolution**: after the
    bullet "**Only a key passed literally in the command arguments is user-supplied.** …" add the
    bullet: "**An argument asking for a new issue** — \"create a jira task\" and the like — files one
    first (`createJiraIssue`, on the `twg` route `twg jira workitem create`) in the `## jira`
    project, typed Task, its summary and description drawn from the request, labelled per **Labels
    on issues the pipeline creates** (`skills/flow-contracts/jira-integration-finish.md`); the new
    key is then the user-supplied key. Filing it is part of resolving the issue — never the whole
    request and never a stop: the run continues into the In Progress transition and planning
    (kan-1040: a run filed the issue and ended its turn). A failed create is one `⚠ Jira: skipped —
    <reason>` line and resolves as no hit."
  - [ ] **Step 2: Verify.** `scripts/check-references.sh && scripts/check-markdown-integrity.sh`
    exit 0.

**Build:** green
**Files:** `skills/flow-contracts/jira-integration.md`
**Tests:** **none** — a prose contract; the guards cover it.
**Regression:** none — no test is declared.
**Commit:** fix(flow-contracts): keep planning after filing a requested Jira issue
**After:** none

**Decision:** create-then-continue
