# self-review-kan-1021-flow-improvement-move-the-bash-hook-and-setup

## Why

The self-review pass of `kan-1021-flow-improvement-move-the-bash-hook-and-setup` found three
fixable pipeline defects:

- `project-get.sh <worktree> "review panel citation check"` prints the key's fenced block with its
  fence lines, and `skills/flow/review-panel.md` says only "run its command": an `eval` of the
  printed body runs a bare `bash` from the opening fence line, which hangs on open stdin.
- The `flow-low`, `flow-medium` and `flow-high` agents (`agents/flow-*.md`) carry no Skill tool,
  yet the implementer dispatch's two REQUIRED SUB-SKILL paragraphs (`skills/flow/implement.md`) and
  the panel-fix subagent's sentence (`skills/flow/review-panel-fix-round.md`) tell them to invoke
  `superpowers:test-driven-development` and `superpowers:systematic-debugging`. This run's
  panel-fix agent reported it had no Skill tool.
- `skills/flow/review-panel-fix-round.md` sends every unresolving demonstrates citation back to its
  reviewer once; this run's parent corrected F1's citation line (379 to 380) in place instead,
  against the contract, because a bounce costs a slot round-trip for a one-line offset.

## What changes

- `skills/flow/review-panel.md` says the citation check's command is the lines between the fence
  markers, never an `eval` of the printed body.
- The two implementer-dispatch REQUIRED SUB-SKILL paragraphs and the panel-fix sentence name the
  installed `~/.claude/plugins/cache/*/superpowers/*/skills/<skill>/SKILL.md` file to read and
  follow, and `check-dispatch-paragraphs` requires both paragraphs to carry that path.
- A demonstrates citation whose cited path holds the claimed content verbatim on another line is
  corrected in place by the parent, re-audited and recorded in the round's `flow record pass`,
  never bounced.
