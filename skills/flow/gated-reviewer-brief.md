# Gated reviewer brief

Every gated reviewer bundle of **4. Execute (SDD + TDD)** (`skills/flow/implement.md`) — an original
bundle, a fix re-review or a handshake retry — reads this file in full first: its prompt is a stub
naming this file and one `PASS task-<n>` block per task it reviews.

> **TOOLS:** Every tool you need that is not already listed in your tool set — `SendMessage`,
> `Monitor`, an MCP tool — is loaded in one `select:<name>,<name>` ToolSearch in your first turn,
> before anything else. Never ToolSearch for a tool already listed, and never a wildcard query: a
> schema loaded later changes your tool list and re-prices your whole context at full input rate.

> **FOREGROUND BUILDS:** Never end your turn with a build, test run, or other long-running
> command still executing in the background. Run it in the foreground, or poll it to
> completion, before you stop.

> **REPRODUCE, DON'T READ:** Where a behaviour crosses a boundary — the store, the filesystem, a
> guard, a real transcript, a real process — at least one check you make MUST exercise the real
> thing. A claim you did not run is worth less than one you did: a doc comment, a type signature
> and a passing test can each read plausibly and be false. Run it before you accept it, and run it
> before you reject it.

> **NO DELEGATION:** Do this work yourself. Never call the `Agent` tool, and never spawn a
> subagent, background agent or helper of any kind — you are the leaf of this run, and any child
> you start is unrecorded and outside the parent's closed list (**Dispatch sites — the parent's
> closed list**, `skills/flow/implement.md`). Reading, searching, reproducing and
> fixing are your own Read, Bash and Edit calls.

> **READ-ONLY REVIEW:** You review a tree other agents are working in — never mutate it. No
> command that writes: no `git checkout`, `git restore`, `git reset`, `git stash`, `git clean`,
> no commit, no index change, and no file edit outside your own report file — the panel's
> mutating slots are the one declared exception, and they work in throwaway copies, not this
> tree. Inspect with Read, Grep, and the read-only git forms — `git show`, `git diff`,
> `git log`, `git status`. A mutation you cause is indistinguishable from a defect the next
> implementer inherits, and a restore you perform is a claim nobody can check: gymie KAN-635's
> reviewer ran `git checkout <sha> -- .` mid-review, destroyed uncommitted planning artifacts,
> and reported the tree restored.

> **INDEPENDENT PASSES:** each pass reviews its own task's diff range and the code, never an
> earlier pass's report or conclusions; a defect that sits in two passes' ranges is raised under
> each. Write each pass's report file before beginning the next pass.

and, for each `PASS task-<n>` block of your prompt:

> **REPORT FILE:** write this pass's report to
> `<abs-worktree>/.superpowers/sdd/reviewer-report-task-<n>.md` before beginning the next pass,
> the bundle's last one as your **last** act — a `## Verdict` section carrying exactly `clean` or
> `fix`, and each finding with its file, line and severity — Critical, Important or Minor as
> `skills/flow/reviewer-calibration.md` defines them. The verdict is `fix`
> only when a Critical or Important finding stands; Minor-only findings are `clean`. The dispatcher waits on every pass's
> file (`test -s` on each), and a fix round's re-review writes
> `reviewer-report-task-<n>-fix-<k>.md`.
