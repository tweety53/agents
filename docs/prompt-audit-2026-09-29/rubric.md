# Slimming-analysis rubric (read-only task)

> The brief each of the seven audit passes in this directory received, verbatim, at `700e184`. "Your output path" was that pass's `audit-*.md` file. Plan: [`README.md`](README.md).

Repository: /home/user/agents — the source of the `/flow` Claude Code skills (Markdown prompts).
Goal: find how to shrink the instruction text a `/flow` session loads WITHOUT changing behaviour.
You are doing analysis only. NEVER edit, create, move or delete any file inside /home/user/agents.
Write your report only to the output path your task names.

## How sessions load text (the operator runs /clear between phases)

Each phase is its own fresh session; each session pays for every byte it loads on every later turn.
- Every session: global always-on rules block, project CLAUDE.md, commands-claude/flow.md,
  skills/flow/SKILL.md (router), skills/flow-contracts/pipeline.md.
- PLANNING session (creating run, all inline in the parent): skills/flow/brainstorm.md,
  skills/flow/brainstorm-planner.md, flow-contracts/jira-integration.md, plan-provenance.md,
  build-green.md; cited at point of use: worktree-resolution.md, git-boundaries.md,
  handoff-blocks.md, operator-prompts.md.
- IMPLEMENTATION session (parent orchestrates; execution is `inline` for micro/small/regular class,
  `sdd` subagent implementers only for `big`): skills/flow/implement.md, review-panel.md,
  verify-and-handoff.md, flow-contracts/artifacts-registry.md, worktree-resolution.md,
  session-records.md, git-boundaries.md; cited: operator-prompts.md, model-policy.md,
  known-bugs.md, the *-reviewer-prompt.md templates the parent fills; conditional:
  review-panel-optional-slots.md (roster has bugbot/mutation/exp-), workspace-isolation.md,
  visual-verify.md (UI paths touched).
- FINISH session (bare /flow at IN_PROGRESS; this operator's default route is merge-and-push, which
  chains run 1 into run 2 in one session): skills/flow/integrate.md, finish-contract-run1.md,
  archive.md, finish-contract-run2.md, jira-integration.md, git-boundaries.md, session-records.md,
  artifacts-registry.md, worktree-resolution.md; conditional: jira-followups.md.
- `*-rationale.md` files are NEVER loaded by a run (repo convention: reasoning for editors).

## The repo's own rules for a behaviour-neutral trim (rules/be-brief.mdc, precedent KAN-378)

- Cut, never paraphrase: delete a restating passage or move text VERBATIM; never reword.
- Never cut: a normative statement (imperatives, "never"/"always"/"only"/"must", conditions,
  exceptions, definitions, scope qualifiers), an exit-code contract, an ordering constraint, a
  scenario, a worked example, operator-prompt wording, stage keys, guard names, field names.
- A recorded reason a rejected alternative was rejected may be MOVED verbatim to the sibling
  `-rationale.md`, never deleted.

## Levers — classify every candidate passage as exactly one

- RATIONALE — reasoning for an editor, not an instruction for the run: rejected alternatives,
  design history, KAN-/decision-id provenance, incident narratives, "measured" provenance, why a
  rule exists. The instruction it explains stays; only the explanation moves (verbatim) to the
  sibling -rationale.md. Do NOT count a clause that changes what the model does (a condition,
  exception, scope qualifier, definition, tie-breaker) — that is instruction, even if it says
  "because". Mark confidence high only when removing it could not change any action.
- DUPLICATE — restates something stated canonically elsewhere IN THE SAME SESSION'S LOAD SET (or
  earlier in the same file). Name the canonical file:line. A cross-reference that merely cites a
  heading is not a duplicate. Note if the copies already differ (drift = a latent bug, report it).
- STALE — text about something that no longer exists or is no longer true. Verify with ls/grep
  (e.g. a named script, file, heading, harness, flag that is gone). Give the evidence.
- LAZY-SPLIT — a contiguous block needed only under a condition decidable at load time (execution
  mode, class, fix run vs first run, a fix round with findings, a landing route, harness zcode, a
  roster slot, first run only...). It would move verbatim into a sibling file loaded by a
  "Load X only when Y" directive (precedent: review-panel-optional-slots.md, visual-verify.md).
  State the condition, and how often you expect it to be false in a typical run.
- MECHANICS — a procedural recipe (shell pipeline, multi-step derivation, a table of mechanical
  rules) that a script or the `flow`/`flow-guard` CLI could execute, leaving a one-line call plus
  its exit contract. Needs code + parity tests, so it is a bigger change; flag it separately.
- MISPLACED — text loaded by a consumer that never acts on it (e.g. dispatcher-only notes inside a
  prompt a subagent receives, subagent-only text the parent carries, /flow-status-only text in a
  /flow file).

## Report format (Markdown, written to your output path)

1. `## Totals` — table: file | bytes | RATIONALE | DUPLICATE | STALE | LAZY-SPLIT | MECHANICS |
   MISPLACED (bytes each, conservative; count a byte once).
2. `## Candidates` — one table row per passage:
   `| id | file:startline-endline | bytes | lever | condition or canonical location or evidence |
   first ~10 words verbatim | confidence high/med/low | behaviour-risk note |`
   Measure bytes exactly: `sed -n 'A,Bp' FILE | wc -c`. Prefer whole paragraphs/sections.
3. `## Top 5` — the five highest-value moves for these files, one line each, bytes and session.
4. `## Drift / bugs found` — any restated copies that already disagree, any broken instruction.
5. `## Not slimmable` — big blocks you judged must stay, one line why (so nobody re-litigates).

Be conservative: a false "safe to cut" is worse than a missed saving. Read every assigned file in
full (use Read with offset/limit for long files; do not skim).

Report tersely back in chat: totals and top 5 only, no preamble. Full detail goes in the file.
