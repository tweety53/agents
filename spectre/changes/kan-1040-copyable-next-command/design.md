## Context

KAN-1040. Two operator asks of 2026-10-27, folded into one change at the operator's instruction:
a one-tap copyable next command, and no stop after filing a requested Jira issue. The design is
`.superpowers/sdd/2026-10-27-kan-1040-copyable-next-command-design.md` in this worktree.

No Go, shell or hook code renders or checks a handoff's `Next:` line, its last line or its fences
(explored at plan time over `stats/`, `scripts/`, `hooks/`); the change is prose in the contracts
and the handoff templates only.

## Decisions

### The next command alone in its own fence

**ID:** next-command-fenced
**Status:** active
**Chosen:** `/flow <name>` is printed alone inside a plain triple-backtick fence, no language tag,
the last thing in the handoff; `/clear` stays a bare line above it — the operator's answer: they
always type `/clear` themselves, and a fence holding only the command copies exactly what is pasted.
**Considered:** `/clear` and the command in one fence — one tap copies both lines, and `/clear`
cannot be pasted together with the command; a fence for each — the operator types `/clear`
themselves, so its fence is noise.

### /flow-status keeps its table

**ID:** status-unchanged
**Status:** active
**Chosen:** `/flow-status` keeps its inline next commands — operator's choice.
**Considered:** one fence per open change below the table — a fence cannot sit in a table cell, and
the operator declined the extra block list.

### Field lines stay rendered Markdown

**ID:** blocks-still-rendered
**Status:** active
**Chosen:** the rule that a block prints as rendered Markdown, never inside a code fence, stays for
every line but the next command, which alone moves into a fence.
**Considered:** fencing the whole block — its `**Field:**` markup would reach the operator as
literal asterisks, the defect the rule exists to prevent.

### A requested issue is filed, then planned on

**ID:** create-then-continue
**Status:** active
**Chosen:** a bullet in **Resolution** (`skills/flow-contracts/jira-integration.md`): an argument
asking for a new issue files one in the `## jira` project, labelled per **Labels on issues the
pipeline creates**, and the new key is then the user-supplied key; the run continues into the In
Progress transition and planning. A failed create is the existing `⚠ Jira: skipped` line and
resolves as no hit.
**Considered:** a rule in `skills/flow/brainstorm.md` section **A** — **Resolution** is canonical
for how `jiraIssue` is decided and section **A** defers to it, so a second home would restate it.

## Open questions

## Recommended approach

Edit the handoff rules in `skills/flow-contracts/pipeline.md` **Handoff output** and their
rationale, then every template that prints `Next:`, so each ends on a fence holding the command
alone (decision `next-command-fenced`; the rest of the block untouched per
`blocks-still-rendered`; `/flow-status` left as is per `status-unchanged`). Add one bullet to
**Resolution** for a requested issue (`create-then-continue`).

## Critical files

- `skills/flow-contracts/pipeline.md` — **Handoff output**: the canonical shape and rules.
- `skills/flow-contracts/pipeline-rationale.md` — why the next command is printed as it is.
- `skills/flow-contracts/handoff-blocks.md` — the per-state block templates.
- `skills/flow/brainstorm.md`, `skills/flow/integrate.md`, `skills/flow/verify-and-handoff.md`,
  `skills/flow/cleanup.md`, `skills/flow-plan/SKILL.md` — templates printing `Next:`.
- `skills/flow-contracts/finish-contract-run1.md` — names run 1's handoff's last line.
- `skills/flow-contracts/jira-integration.md` — **Resolution**.

## Existing code to reuse

- **Labels on issues the pipeline creates** (`skills/flow-contracts/jira-integration-finish.md`) —
  the label set a filed issue carries.
- **Never blocking** (`skills/flow-contracts/jira-integration.md`) — the failed-create line.

## Verification

- `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh` and every other
  `scripts/check-*.sh` guard in `.flow/project.md`'s `## lint` exit 0.
- `git grep -n -F -e 'The next command is the last line' -e 'stay bare lines' -- skills` prints
  nothing.
- Each edited template, read rendered: its outer template fence is four backticks, so the inner
  three-backtick fence around `/flow <name>` survives as output.

## Live check

None — prose contracts only; no service, store or runtime state changes.
