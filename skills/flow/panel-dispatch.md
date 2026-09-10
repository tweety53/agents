# Panel dispatch — shared mechanics

The parts of a review-panel dispatch that `/flow` (`skills/flow/review-panel.md`) and `/flow-fast`
(`skills/flow-fast/review.md`) both need verbatim: writing `final-review.diff` and three paragraphs
every dispatch prompt carries.

## Writing `final-review.diff`

Write `<abs-worktree>/.superpowers/sdd/final-review.diff` (the canonical worktree's) once per round
from **every** worktree in the change's resolved set (**Resolving a change's worktrees**,
`skills/flow-contracts/worktree-resolution.md`), in resolved order — each worktree's section opened
by a header naming it and its own working-notes merge base, then that worktree's `git diff
<merge-base>` (staged and unstaged):

```sh
: > <abs-worktree>/.superpowers/sdd/final-review.diff
# for each <worktree> in the resolved set, in order:
printf '# worktree: %s — merge base %s\n' "<worktree>" "<merge-base>" \
  >> <abs-worktree>/.superpowers/sdd/final-review.diff
git -C <worktree> diff <merge-base> >> <abs-worktree>/.superpowers/sdd/final-review.diff
```

A single-worktree change writes the same shape with one header.

## Dispatch paragraphs

Every dispatch prompt carries these three paragraphs verbatim.

> **INDEPENDENT PASSES:** each pass starts from `final-review.diff` and the code, never from an
> earlier pass's report or conclusions. Do not cite, defer to, or skip a defect because an earlier
> pass raised it — if it sits in this pass's angle, raise it again under this pass. Write each
> pass's report file before beginning the next pass.

> **TOOLS:** Every tool you need that is not already listed in your tool set — `SendMessage`,
> `Monitor`, an MCP tool — is loaded in one `select:<name>,<name>` ToolSearch in your first turn,
> before anything else. Never ToolSearch for a tool already listed, and never a wildcard query: a
> schema loaded later changes your tool list and re-prices your whole context at full input rate.

> **MODEL HANDSHAKE:** the first line of your first reply is `Model: <the model named in your own
> system prompt>` and nothing else on that line. Answer it before any tool call.
