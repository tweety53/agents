# Design — kan-854-flow-fix-audit-behaviour-decisions

One decision per KAN-854 item, numbered as the issue numbers them. Each names what was done and
the alternative rejected. Line numbers are gone on purpose; re-locate by quoted text.

## Decisions

### D1 — the panel-fix retry key follows the guard

review-panel B1. The text now says `panel-fix-<round>[-<chunk>]-retry`, the only retry shape
`check-panel-fix-single-dispatch`'s `pfdKeyRE` accepts. Rejected: widening the guard to accept
`<round>-fix-retry`. Every other panel-fix key uses the `panel-fix-` prefix, and the text was the
outlier.

### D2 — a fix run re-decides inside implement.md section 3

planning #6, implement #15.

- The rule now sits at the end of **3. Documenting a fix, before implementing it**, where a fix run
  reads it. It runs Decide from `plan-class.sh` on and records a second `flow.decide` row.
- Neither **Plan review gate** nor the closing `flow.writing-plans` mark runs on a fix run.
- brainstorm.md keeps a citation only.
- The router's stage-key table names `flow.decide (fix runs)` under implement.md.
- Rejected: dropping the fix-run re-Decide. The class would then go stale whenever appended tasks
  move it, which is the defect itself.

### D3 — the inline parent reads hazards itself

implement #11. Hazards are store rows, not a `project.md` key, so the parent runs
`flow hazards -C <canonical-worktree> -shape <shape>` once per run, before the first edit. That is
the same source and filter the gather uses, and a failed call is reported and never blocks.
Rejected: gathering a bundle inline and reading it back. That contradicts "never read the bundle
back" and puts the whole bundle into the parent's context.

### D4 — the inline binding names its five exclusions

implement #12. Work paragraphs still bind the parent "in the same words". NO DELEGATION, MODEL
HANDSHAKE, TOOLS, CONTEXT BUNDLE and REPORT FILE govern only a child's own channel, and are named as
excluded, along with what replaces the bundle's inputs. Rejected: listing every binding paragraph
positively. That list would drift each time a paragraph is added.

### D5 — the gated reviewer on micro/small/regular is one `flow-low` bundle

implement #13, #14.

- On `micro`, `small` and `regular`, the run's one bundle goes out as `subagent_type: flow-low`,
  with `DEFAULT_MODEL` (or the session override) as `model`, recorded `-effort low`. This is the
  `default`-panel reviewer's pair.
- `micro` joins the one-bundle rule, and `check-task-reviewer-single-dispatch` now enforces it
  for `micro`.
- Rejected: reusing `panel.dispatches[0]`'s pair (a micro panel has none).
- Rejected: micro skipping the gated reviewer (its gate can still fire).

### D6 — implement.md's handshake is canonical

implement #6. The zcode exception in **The handshake** was added deliberately after model-policy.md's
"compares against `glm-5.3-flash`" sentence, which now cites it instead. Rejected: moving the
exception into model-policy.md, a restructure beyond reconciling two copies.

### D7 — the SDD skill loads only under `execution: sdd`

implement #16. The skill-map row, and skills/README's copy of it, are now conditional. Rejected:
deleting the row, because section 4 still cites that skill's parallel-dispatch guidance.

### D8 — every worktree section 2 creates runs `## worktree setup`

implement #19. Peer and `## apps` worktrees run it before their link, commit or first task, handled
as brainstorm.md step 4 handles the kickoff worktree. A resumed worktree runs nothing. Rejected:
rewording project-configuration.md to list each creation site. The row is true once section 2
runs setup.

### D9 — rule 2 reads both sibling spellings

every-session D11.

- `check-guard-symlinks` gains a second matcher for `$(dirname -- "${BASH_SOURCE[0]}")/<name>`
  and `$(dirname "$0")/<name>`. A name starting with `.` is never a sibling, so `/..` is excluded.
- SKILL.md's two-shim sentence becomes a rule covering every `flow-guard` shim, and
  KNOWN-BUGS F8 is closed.
- Rejected: hand-listing the shims (it goes stale).
- Rejected: rewriting 21 shims to use `$SCRIPT_DIR`.
- The real tree still passes the guard, because every skill carrying a shim already carries `lib/`.

### D10 — three rationale-only rules return to run files

every-session D24.

- pipeline.md now states "do not branch on `flow stage`'s exit code", with the journal-and-exit-0
  behaviour.
- SKILL.md now states that `<literal-token>` is written out, and that each invocation generates its
  own token.
- The rationale copies stay. Rejected: a cut-and-move, because rationale files keep the "why" by
  convention.

These stay rationale-only, deliberately:

- **Editor-facing "do not add" rules.** No run edits the contracts. The finish sync step, the
  artifact byte budget, handoff-field lockstep, the third Jira question, and suppression markers
  (already `## lint` policy).
- **Rules code already enforces.** `check-stage-mark-calls` covers a guessed `<change>` and a
  hardcoded `-harness`.
- **Stale rules.** The `docs/superpowers/` staging note, and the legacy `planningEffort` gate note.

### D11 — the fix subagent's dispatch states the `fix-mutation:` shape

review-panel B7. The MUTATION PROOF blockquote now carries the three line shapes and the
guard-script `<pre>→<post>` form, and `check-dispatch-paragraphs` pins the shape as a shared phrase.
Rejected: passing the contract's fenced block to the subagent as another file (more context for
three lines).

### D12 — the parent fills placeholders from a table in review-panel.md

subagent-prompts D2.

- The table sits under **The roster**, next to the rule that the parent never reads the templates.
  It defines `[DIFF_PATH]`, `[ARTIFACT_PATHS]`, `[CONTEXT_BUNDLE_PATHS]`, `[PRINCIPLES_PATH]`,
  `[STANDARDS_PATHS]` and `[GLOBAL_CONSTRAINTS]`.
- `[REPO_COPIES]` and `[PLAN_OR_REQUIREMENTS]` no longer exist; they left with the retired slots.
- `[GLOBAL_CONSTRAINTS]` is a fixed pointer to the bundle's `design.md` section.
- Rejected: the parent quoting constraints itself (a judgment call that costs context).

### D13 — slots read the plan and principles from the bundle

subagent-prompts D1. The primary and principles templates, and the CONTEXT BUNDLE paragraph, now
say to read `proposal.md`, `design.md`, `tasks.md` and the principles from the bundle and never
re-open them. They open a file directly only when the census lists it as skipped or refused, or
the bundle is absent. Rejected: dropping `[ARTIFACT_PATHS]`/`[PRINCIPLES_PATH]`, which the
CONTEXT BUNDLE FAILURE continue path still needs.

### D14 — README's claim is corrected; the baseline table stays

subagent-prompts D6.

- README now says CLAUDE.md delivery to subagents depends on the harness and the agent type. This
  session saw one Claude Code subagent type that received no CLAUDE.md.
- `agent-baseline.md`'s rule table is kept, because it is the one channel that reaches every
  subagent.
- **Pending operator verification:** what `flow-*` agents receive on the operator's machine. If they
  receive CLAUDE.md, a follow-up removes the table for Claude Code and reconciles
  `dispatch-carries-the-baseline.mdc`'s premise with it.
- Rejected: removing the table now, unverified.

### D15 — `/flow-fast` no longer re-reads CLAUDE.md/AGENTS.md

subagent-prompts D8. The harness already has the instruction file in context, and `/flow`'s own
load-context reads neither. Rejected: a conditional "read when not in context", which a session
cannot judge reliably.

### D16 — `SELF_REVIEW_MODEL` is dropped from `/flow`

finish D12.

- Removed: archive.md's resolution block (1.6 KB, two subprocesses, a possible exit-2 stop), the
  run-2 sentence and SKILL.md's paragraph.
- project-configuration.md, model-policy.md, flow-settings, flow-self-review and the store comments
  now call the field and key inert.
- `check-model-resolution-shell.sh` now tests `DEFAULT_MODEL`/`MODEL_SOURCE` with explicit
  expectations, which closes the two KNOWN-BUGS entries about inherited expectations. It refuses a
  re-added resolution. KNOWN-BUGS F7 (archive.md snippet drift) is closed.
- Rejected: recording the value, which pays the full cost for a record nothing reads.
- Deferred to a follow-up: retiring the field, flag, API and key end to end, because that crosses
  the store, API and CLI.

### D17 — check 4 has one trigger, in run 2

finish D8. run 2's check 4 now reports both buckets and proceeds, asking only about an entry that
is irreplaceable and unpreserved. archive.md's override becomes a citation, and keeps the check-6
clarification. Rejected: keeping run 2's broader ask canonical. It would prompt on every `.env`,
and it reverses `/flow`'s current, argued behaviour.

### D18 — step 6 deletes a surviving legacy proposal artifact

finish D11. Step 6 removes `<state-dir>/<name>-proposal-artifact.html` when present, and says so
when absent. The guard's LEFTOVER report is correct and unchanged. Rejected: having the guard
ignore the file, which weakens a registry row and leaves the file behind forever.

### D19 — the archive-scope fallback uses the project root

finish D10. Both archive.md and run 2 now name `<project>/spectre/changes/`, the tree the guard
call checks. Rejected: none; `<agents repo>` was plainly wrong.

### D20 — explain-before-asking is stated at the gate that offers filing

finish D14, D16.

- The four sentences moved verbatim from jira-followups.md to integrate.md's unfinished-work gate,
  and jira-followups.md keeps a pointer.
- D16's stale second loading site was already removed by KAN-853.
- Rejected: two copies of the rule, the drift pattern KAN-853 removed.

### D21 — `<repos>` counts repository roots in `/flow` too

planning #7.

- Decide now passes the number of distinct repo roots the plan's `**Files:**` fall under, the
  meaning `/flow-plan` already uses. The `repos>1` rule can now fire in `/flow`.
- brainstorm.md drops the now-unused worktree-count handoff.
- `plan-class.sh` is unchanged apart from its header comment.
- Rejected: `plan-class.sh` deriving the count itself, which needs project-config parsing in a
  guard.

### D22 — `flow jira transition` stays unwired and documented

planning #15. jira-integration.md now says the command is not the pipeline's transition path. Its
exit 1 on an unrecognised status is that section's default **No**, and a caller that wires it in
must ask. Rejected: making it the transition path (every unconfigured flowd would silently skip).
Rejected: a distinct exit code, which has no caller to use it.

### D23 — wave-group copies have a registry row and a cleanup

verify-and-contracts D10.

- The registry gains a row, implement.md names the removal command, and run 2's **Worktree
  cleanup** removes a surviving copy. Checks 5 and 6 stay gates on a copy. Checks 1–4 become a
  disclosure (status and log) plus one ask, default remove: a copy starts from the worktree's
  uncommitted state, and run 1's reshape folds every pick, so a clean-tree or `git cherry` gate
  would fail on every copy and strand the change (found in review).
- `check-cleanup-complete` reports a **detached** worktree whose last path element is
  `<name>-wave-group-<digits>`. Requiring "detached" means a change named that way, which has a
  branch, is never matched.
- Rejected: a `registry-row-not-checked` marker. A dead run is exactly when the copy survives.
- Noted, not done: the mutation slot's copy has the same dead-run gap.

### D24 — the port contract matches the script: no probe, a collision is relayed

verify-and-contracts D2.

- workspace-isolation.md no longer promises an `lsof` check or rediscovery; it describes the
  deterministic block and the relayed refusal that verify-and-handoff.md already implements.
- The rationale's race paragraph is removed.
- The kan-302 self-review's nonexistent "KAN-314" now points at KAN-854.
- Rejected: implementing an `lsof` loop in `prepare-workspace.sh`. It adds a dependency, breaks
  deterministic ports, and only narrows a race the daemon already covers.

### D25 — `cost-status` is one call with `-C`

verify-and-contracts D15. One call against the canonical worktree the ledger render targets,
because `Costs:` is one line. Rejected: one `-C` call per worktree, which gives N lines for a
single field.

## Verification

- `## lint` is green: every guard, `gofmt -l`, and `go vet`. `check-verbatim-moves.sh` is clean with
  `verbatim-moves.txt`.
- `check-normative-inventory.sh` output is unchanged against the pre-change capture.
- The new Go tests pass. Each new test was mutation-proved: it fails with the code change
  reverted.
- The Go failures that remain (bash-parity pins, missing `web/dist`, container path
  normalisation) fail the same way on the base commit.
- `stats/web` `tsc -b` is clean. No SPA file changed.
