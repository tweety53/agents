# Review slot brief

Every review panel dispatch (`skills/flow/review-panel.md`) reads this file in full first: its
assembled brief names it, never these paragraphs. The WORKTREES list, the SECTION repository and
each pass's REPORT FILE path are filled in by that brief.

> **FOREGROUND BUILDS:** Never end your turn with a build, test run, or other long-running
> command still executing in the background. Run it in the foreground, or poll it to
> completion, before you stop.

> **TOOLS:** Every tool you need that is not already listed in your tool set — `SendMessage`,
> `Monitor`, an MCP tool — is loaded in one `select:<name>,<name>` ToolSearch in your first turn,
> before anything else. Never ToolSearch for a tool already listed, and never a wildcard query: a
> schema loaded later changes your tool list and re-prices your whole context at full input rate.

> **NO DELEGATION:** Do this work yourself. Never call the `Agent` tool, and never spawn a
> subagent, background agent or helper of any kind — you are the leaf of this run, and any child
> you start is unrecorded and outside the parent's closed list (**Dispatch sites — the
> parent's closed list**, `skills/flow/implement.md`). Reading, searching, reproducing and
> fixing are your own Read, Bash and Edit calls.

> **REPORT FILE:** write your full report to
> `<abs-worktree>/.superpowers/sdd/panel-report-<round>-<id>.md` before you end your turn — every
> finding with its `file:line`, severity, the sentence naming the defect, how to see it (a command where one exists), and the
> id of the task whose `Task-Id:` commit its `file:line` falls in — none where it spans tasks or
> falls in none. A finding's `file:line` is a line of that file in the worktree under review,
> never a line number of `final-review.diff` or of any other diff. Your return message is the findings summary, each finding naming its task id
> too; the file is the record. Open the file with one `Critical: <n>`, one `Important: <n>` and one
> `Minor: <n>` line counting the findings it raises at each severity, then one `Unread: none` line, or
> `Unread: <repository>[, <repository>…]` naming each `# worktree:` section you did not read in
> full. Write this file's own text, never one derived from another role's report or from a shared
> text by substitution: a substitution rewrites every match in the findings as well.

> **WORKTREES:** this change spans `<abs-worktree-1>`, `<abs-worktree-2>`, …; `final-review.diff`
> is sectioned by worktree, each section headed `# worktree: <path> — merge base <sha>`. When more
> than one is listed, prefix every finding's `file:line` with that worktree's repository name —
> its main checkout's basename (`gymie-frontend:src/Foo.tsx:42`).

> **SECTION:** this dispatch reads only the `# worktree:` section of `<repo>` in its diff file.
> Follow a seam into another section only as far as judging this section needs.

> **INDEPENDENT PASSES:** each pass starts from `final-review.diff` and the code, never from an
> earlier pass's report or conclusions. Do not cite, defer to, or skip a defect because an earlier
> pass raised it — if it sits in this pass's angle, raise it again under this pass. Write each
> pass's report file before beginning the next pass.
