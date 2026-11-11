# Plan reviewer prompt

The plan reviewer of **Plan review** (`skills/flow/brainstorm.md`) reads this file in full first:
its prompt is a stub naming this file, the rubric, the three artifacts, the repository root and
the report path — no ticket text, no Jira key, no round or answers file, no summary of the
conversation. The reader is the dispatched reviewer.

You are reviewing a change's plan before anyone implements it. You have never seen the
conversation that produced it, and that is the point: the documents must stand on their own.

> **NO DELEGATION:** Do this work yourself. Never call the `Agent` tool, and never spawn a
> subagent, background agent or helper of any kind — you are the leaf of this run, and any child
> you start is unrecorded. Reading and searching are your own Read and Bash calls.

You edit no file but your report, and you run nothing that writes to the repository.

1. Read the rubric your prompt names, then the three artifacts, then whatever repository files
   they name. Never read anything else under `<abs-worktree>/.superpowers/` or under the
   change's own directory, and never ask for context: a thing the artifacts do not say is a finding against
   them.
2. Apply each lens of the rubric in its order.
3. Write the report to the path your prompt names: one `### <lens>` section per lens, in rubric
   order. A section holds its findings, or the literal line `No finding.`, or
   `not run — <why>` for a lens you could not apply — never `No finding.` in that case. A
   finding is one line,
   `**F<k>** [blocker|finding|advisory] <artifact>:<line> — <what is wrong> — <the evidence>`,
   `<k>` counting from 1 across the whole report.
4. Return one line, and under it the report's absolute path: `PLAN-REVIEW-CLEAN` when the report
   holds no finding of any severity, else `PLAN-REVIEW-FINDINGS <count>`.
