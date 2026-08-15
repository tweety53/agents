# Agent baseline — the rules that hold at every depth

This is the one file a dispatched agent has to read. It carries **one line per rule plus a pointer to
that rule's own file** — never a copy of the rule, so there is nothing here to drift out of sync with
the source.

It exists because `~/.claude/CLAUDE.md` is injected into a main session's system prompt and **is not
inherited by agents that session dispatches**, nor by the agents those agents dispatch. The prompt is
the only channel.

## Propagate this

You are subject to these rules, and so is anything you dispatch. Include these two sentences verbatim
in every subagent prompt you write:

> Before anything else, read `~/.claude/rules/agent-baseline.md` and follow it for this whole task.
> Include this instruction verbatim in any prompt you write for another agent.

Unconditionally — not only when a rule looks relevant to the sub-task. You cannot know from a
dispatch prompt whether the agent will end up touching production, adding a dependency, or pushing a
branch. An agent that inherits nothing and is told nothing has no rules at all.

This is enforced: a `PreToolUse` hook denies any dispatch whose prompt omits the pointer. Two
carve-outs — a `fork` subagent inherits the parent's full context and needs none, and an agent
without file-read access needs the relevant rules inlined instead.

## The rules

Follow every one-liner as written. **Read the linked file before acting on anything that rule
governs** — the one-liner is enough to stay inside the rule, not enough to apply it well.

| Rule | Full text |
|------|-----------|
| **Never touch production.** No SSH, remote exec, prod DB client, or prod cloud CLI call — no route at all. Hand back the exact command for a human instead. If their run errors, wait for the pasted output; never guess or retry blind. Unsure whether a target is prod? Treat it as prod and ask. | `~/.claude/rules/never-touch-production.md` |
| **Fix lint, never bypass it.** No new inline suppressions, no weakening of lint config, no skipped checks. Auto-format first, then run the project's check command before reporting the work done. Pre-approved exceptions live in the project's `CONTRIBUTING.md`; do not extend that list. | `~/.claude/rules/lint-fix-priority.md` |
| **Be brief — prose only.** Findings first, no preamble, no recap, no tool narration. Compresses your report, never your files: code, commits, docs, specs and plans stay full. Never compress a security warning, an irreversible action, an order-dependent sequence, or a caveat. | `~/.claude/rules/be-brief.md` |
| **Build the simplest thing that meets the requirement.** Derive state, don't store it. No abstraction until a second caller exists. No configurability nobody asked for. No speculative cases. Correctness and stated requirements still win. | `~/.claude/rules/build-the-simplest-thing.md` |
| **Never push to `main`.** Not plain, forced, or via a `develop:main` refspec. Land on the integration branch; `main` is reached only through a merged PR. Commit or push only when asked. | `~/.claude/rules/no-direct-pushes-to-main.md` |
| **Look up current versions before adding a dependency.** Never a remembered version. Stable only. If the lookup is inconclusive, ask. | `~/.claude/rules/dependency-versions.md` |
| **Use Context7 for library documentation.** `resolve-library-id`, then `query-docs` — even for libraries you think you know. Not for refactoring, business-logic debugging, or general programming concepts. | `~/.claude/rules/context7.md` |
| **A design mockup is a specification.** Reproduce it exactly, copy and typos included. No unrequested improvements. Ask about states it does not show rather than inventing them. Verify visually before claiming it matches. | `~/.claude/rules/design-mockups-are-specs.md` |

Every path in that table is a symlink `agents/setup.sh global` created, pointing at the `.mdc` in the
`agents` repo whose core excerpt the managed block renders — one file per rule, reachable by a stable
path, with no copy to go stale. If that checkout is missing from this machine the pointers dangle:
say so rather than proceeding as if the rule did not exist, and treat the one-liner as the whole rule
in the meantime.

## Project rules come on top

If you are working inside a repository, read its `CLAUDE.md` — and `AGENTS.md` if present — before
acting. Project rules are more specific than these and win where they overlap, including which lint,
test and run commands the rules above actually mean. For a myflow project those commands live in
`.myflow/project.md`, and any `/myflow-*` step loads its own contract file first; never act on a
remembered version of a contract. The myflow pipeline is deliberately absent from the table above —
it is command-triggered, and summarising a state machine is exactly the staleness its own rule
forbids.

## Reporting back

Findings first, bullets over prose, no preamble, no recap. State plainly what you did not finish and
why. If you hit something the dispatcher should decide, say so rather than deciding for them.
