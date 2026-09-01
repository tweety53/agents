# kan-374-flow-run-planning-stages-on-fable-opus-fallback

Jira: KAN-374.

## Why

- `commands-claude/flow.md` pins `/flow` to `model: sonnet`, and every planning stage —
  brainstorming, design approval, artifact creation, writing-plans — runs in that session. The
  model-policy contract still says planning "should run on Opus (or the harness's strongest
  available model)"; nothing enforces it since `/myflow-start` collapsed into `/flow`.
- A command's `model:` frontmatter covers only the turn it starts, and no mechanism lets the agent
  switch the session model mid-run, so the pin was never doing what the policy note claims even
  before the collapse.
- The planning model is a harness-wide preference that some projects will want to override, and
  no setting holds it today.
- `flow.resolve-change` marks nothing: its `stage begin` needs a resolved name, so it fires back to
  back with `flow.kickoff` and records a duration of roughly zero on every creating run.

## What changes

- **Planning runs in a planner subagent** dispatched on `PLANNING_MODEL`; the parent session stays
  on sonnet and relays every operator question and answer. The subagent's context survives across
  rounds, returning at the three stage boundaries so each mark still brackets real work.
- **`PLANNING_MODEL` resolves** from `<project>/.flow/project.md`'s new `## planning model` key,
  then the settings store's new `planningModel` field, then the literal `fable`. A session
  instruction overrides it for one run and is never written back.
- **`/flow-settings` asks a fourth question**, "Planning model"; `flow settings set` gains
  `-planning-model`; `ValidModels` gains `fable`.
- **Fallback is verified, never assumed.** The planner names its own model in its first reply; a
  mismatch re-dispatches once on `opus`, a second mismatch proceeds and reports. The dispatch is
  recorded with the actual model under a new `planner` role.
- **`flow.document-fix`** on a fix run dispatches the same planner to rewrite the proposal and plan.
- **`/flow-research`** dispatches a research subagent on the same `PLANNING_MODEL`, parent relaying.
- **`flow.resolve-change` is removed** from the skill, the README stage table, the CLI's stage
  table and the one test that used it as an example key. Historical rows stay in the store.
- **Model-policy and the command files** are corrected: planning happens in a subagent on
  `PLANNING_MODEL`, and the Claude Code frontmatter note no longer claims to enforce a session model.
