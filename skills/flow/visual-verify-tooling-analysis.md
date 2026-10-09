# Visual verification — the tooling analyst

Loaded by **A missed defect — the tooling analysis** (`skills/flow/visual-verify.md`) only on a fix
run with at least one miss.

`subagent_type: flow-high` (`agents/flow-high.md`, effort `high`), the Agent tool's `model`
parameter set to the literal `opus` (**Model and effort**, `skills/flow/brainstorm-planner.md`), mapped on harness
`zcode` per **Harness mapping** (`skills/flow-contracts/model-policy.md`). Its prompt carries,
verbatim:

> Before anything else, read `~/.claude/rules/agent-baseline.md` and follow it for this whole task.
> Include this instruction verbatim in any prompt you write for another agent.

**The prompt also carries the TOOLS paragraph**:

> **TOOLS:** Every tool you need that is not already listed in your tool set — `SendMessage`,
> `Monitor`, an MCP tool — is loaded in one `select:<name>,<name>` ToolSearch in your first turn,
> before anything else. Never ToolSearch for a tool already listed, and never a wildcard query: a
> schema loaded later changes your tool list and re-prices your whole context at full input rate.

**The prompt also carries the NO DELEGATION paragraph**:

> **NO DELEGATION:** Do this work yourself. Never call the `Agent` tool, and never spawn a
> subagent, background agent or helper of any kind — you are the leaf of this run, and any child
> you start is unrecorded and outside the parent's closed list (**Dispatch sites — the
> parent's closed list**, `skills/flow/implement.md`). Reading, searching, reproducing and
> fixing are your own Read, Bash and Edit calls.

**The prompt also carries the MODEL HANDSHAKE paragraph**:

> **MODEL HANDSHAKE:** the first line of your first reply is `Model: <the model named in your own
> system prompt>` and nothing else on that line. Answer it before any tool call.

It also states: the absolute worktree path; each missed defect, verbatim from the fix
instructions; the paths of the earlier round's `verify-report-*.md` under
`<abs-worktree>/.superpowers/sdd/` and
`<changeRoot>/visual-verification.md`, and the captures they cite; the absolute path of
`skills/flow/visual-verify-verifier.md`; and
this task, as written:

> For each missed defect, name the sweep in steps 8–10 of `skills/flow/visual-verify-verifier.md` that
> should have caught it and why it did not, then name the defect's class — the property, element
> kind or state no sweep read. Write the sweeps that close each class to
> `<changeRoot>/sweeps-<n>.md`: a new sweep, or a sharpened copy of an existing one under its own
> name, each in step 10's sweep shape — what it reads from the capture, and from the frame where
> one is composed, and what makes it fail. A sweep targets the class, never the instance: it must
> be able to find another defect of the class in a view the fix instructions do not name. You fix
> nothing, edit no source and run no capture. Your turn ends with a `## Report` block listing
> each sweep by name with the miss it closes, and your last act before it is writing the same
> block to `<abs-worktree>/.superpowers/sdd/tooling-analysis-<n>.md`.

`<n>` is this fix run's ordinal, the one `flow.document-fix` recorded.

**Recording.** The parent records the dispatch `-role planner`, `-task` omitted, `-model
opus -effort high`, `-key tooling-analysis-<n>`, suffixed `-<worktree basename>` under
the verifier's rule — the pair's semantics are section 4 of `skills/flow/implement.md`, cited here,
not restated. **Handshake** as **The verifier dispatch** (`skills/flow/visual-verify.md`) states it, compared against `opus`.

**The re-run.** The verifier's prompt then carries `sweeps-<n>.md`'s path, and the verifier runs
its sweeps in step 10 beside the sweeps listed there, on every view and frame the change touches
— never only the view the miss was reported in — each where its own wording applies, as
step 10's frame-gating note states for the others. A departure an added sweep finds is a defect
the verifier reports — attributed at the merge base, and blocking either way (**the sweep**,
`skills/flow-contracts/known-bugs.md`). The report carries one line per added sweep.

An analyst that ends without a `## Report`, or whose agent dies, is closed `-outcome aborted`,
and the verifier is dispatched without added sweeps. Either outcome is reported on the handoff's
`**Decisions:**` lines (`skills/flow/verify-and-handoff.md`). A
completed analysis names its sweeps file by absolute path, so the operator can decide whether its
sweeps join step 10 for every later change; the run itself never edits `visual-verify-verifier.md`.
