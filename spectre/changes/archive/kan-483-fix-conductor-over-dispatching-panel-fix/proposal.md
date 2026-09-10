# kan-483-fix-conductor-over-dispatching-panel-fix

## Why

On the KAN-449 `/flow` run the conductor spawned Agent-tool subagents for work the contract
scopes to the conductor itself: four parallel fix subagents at the review panel's fix step, and
two more at the fix-verification step — one to "verify fixes (reading)" and one to "mutation
re-verify" — where `skills/flow/review-panel.md` has the conductor run `run-reproducer.sh` and
read the fix subagent's report directly. KAN-482 fixed the first instance with a fully stated
one-panel-fix-dispatch contract and a store-side guard. The second instance has no fix, and the
class of drift behind both — a conductor reading "the parent does X" as a description rather than
a prohibition on delegating X — is addressed nowhere: every other conductor-scoped
script-run-and-read step in the pipeline is stated in the same phrasing that already failed once.

A store-side guard cannot cover this class. The KAN-449 ledger holds exactly one `panel-fix-0`
row while six unwarranted subagents ran: a rogue dispatch is never recorded, so
`flow record dispatches` has nothing to count. The fix has to reach the conductor before it
launches — in the text it reads and in the prompt it is dispatched with.

The linked issue also narrows what a verifier subagent may run: `## lint` and `## test` run
inline in the conductor, always. Today `flow.verify`'s verifier runs exactly those two lists plus
`check-spec-reach.sh`, so with lint and test carved out it would be a one-script dispatch — the
very shape this change forbids — and the operator chose to drop it: `flow.verify` runs inline in
the conductor, and the verifier dispatch survives for `flow.visual-verify` alone.

## What changes

- `skills/flow/implement.md`'s **Dispatch the conductor** gains a closed table of the conductor's
  permitted Agent-tool dispatch sites — implementer per group, panel bundle per round, one
  panel-fix per round, visual-verify verifier per worktree — with a pre-dispatch self-check: before
  any Agent-tool call the conductor names the table row it is; no row, no dispatch. Every other
  step in the conductor's three files — every `check-*.sh`, `run-reproducer.sh`,
  `gather-dispatch-context.sh`, `prepare-workspace.sh`, `## lint`/`## test`, every `flow record` and
  `flow stage` call, throwaway-worktree add/remove, every report read and diff walk — is stated as
  the conductor's own work, never delegated. The conductor's dispatch prompt cites that subsection;
  KAN-482's fix-dispatch sentence in the prompt folds into the citation. **Inline — the parent
  implements** takes the same table minus the implementer and panel-fix rows.
- `skills/flow/review-panel.md`, `skills/flow/implement.md` section 4 and
  `skills/flow/verify-and-handoff.md` state, at each conductor-scoped step, that the conductor runs
  it itself and never dispatches a subagent for it — as directly as the one-fix-subagent rule
  already is. Additions only: no existing normative sentence is cut or reworded.
- `flow.verify` runs inline: the conductor runs `## lint`, `## test` and `check-spec-reach.sh`
  itself, in order, not stopping at the first failure; a non-zero exit gets one inline re-run of
  that command, then `## Question` with the output verbatim. **The verifier dispatch** is rescoped
  to `flow.visual-verify` only; its TOOLS and MODEL HANDSHAKE paragraphs stay where
  `check-dispatch-paragraphs.sh` expects them. The inline run is recorded per worktree as
  `-role verifier -key verify -agent-id inline`, the shape inline implementer and panel-fix rows
  already use. `skills/flow/SKILL.md`'s `VERIFY_MODEL` paragraph governs one dispatch.
- `scripts/check-contract-budget.sh`'s row for `skills/flow/implement.md` is raised for the new
  subsection — the documented response to a deliberate section addition.

No script, store or spec changes. Observable difference: a `/flow` run's review-panel fix round
dispatches exactly one fix subagent, its reproducer and mutation-proof verification after a fix
round runs as direct conductor work with no subagent spawned for it, and its lint and test
commands run in the conductor rather than in a verifier.
