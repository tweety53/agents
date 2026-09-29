# Design — kan-855-slim-every-session-load

Row ids are the audit's (`audit-every-session.md`, `## Candidates`). The audit ran at `700e184`;
main has since moved (KAN-853, KAN-854, and the planner-chooses-models change), so every row was
re-located by quoted text on `8cc85ee5`.

## Decisions

### D1 — `flow-manual-review.mdc`'s core (A1) stays always-on

The issue's first open decision. Kept: the core is what a session sees before `/flow` loads anything,
and moving it is a behaviour change for every project's non-flow sessions, not a trim. Rejected:
cutting it to a one-line pointer (saves 2.3 KB in every session, but it would be a design change
riding on a byte-neutral trim).

### D2 — the "See … (`-rationale.md`) for why" pointers stay

The second open decision. Kept: they are the repo's breadcrumb convention, and deleting them is a
convention change the audit itself lists under **Not slimmable**. Rejected: deleting them
(≈4.6 KB across `pipeline.md`, `model-policy.md` and `project-configuration.md`).

### D3 — split files are siblings, not the authoring file

G1–G3 went to a new `project-configuration-isolation.md`, not into
`project-configuration-authoring.md`. That file opens "nothing here is consulted by any run", and a
run still reads G1–G3 when `prepare-workspace.sh` cannot be located, so appending there would
make its first sentence false. G6 → `project-configuration-visual.md`, G8 →
`project-configuration-standards.md`. The core carries a `**Load …** only when …` directive for
each:
- standards: when resolving a `## standards` entry;
- isolation: when `prepare-workspace.sh` or `check-cleanup-complete.sh` cannot be located, since
  run 2 step 7's by-hand check applies the `survivors` contract, and step 7 carries its own
  directive;
- visual: when `flow.visual-verify` begins, because steps 1–2 already read the section's shape.

G9, **Where the agents repository is**, stays in the core. It is the single definition of
`<agents repo>`, and the planning (Decide's experimental listing) and finish (integrate's
scoped re-verification) sessions resolve `<agents repo>` paths outside standards resolution. The citers that mean the moved text are repointed: `workspace-isolation.md` (4),
`artifacts-registry.md` (2), `visual-verify.md` (3), `brainstorm-planner.md`, `verify-and-handoff.md`
and `review-panel.md`. The `**Project configuration**` citers that mean the key table or the
optional-file rule still cite the core.

### D4 — G4–G5 inline in run 2 step 5

The `create`/`remove`/`survivors` command table, the working-directory bullets and the token rule
are now indented under step 5, right after the paragraph that runs `remove`. The survivors output
contract (G2) went with G1/G3, because the run reads `check-cleanup-complete.sh`'s verdict, not the
contract.

### D5 — directives in `pipeline.md` stand in for moved sections cited by many files

`## Hand-verifying a guard verdict` (P18) and `## Auto-resolution` (O3, in `operator-prompts.md`)
keep their headings, each now holding one `**Load …** only when …` directive. Their 5 and 14
citers resolve unchanged and land on the directive. That is the one decidable condition that covers
every site: a gate verdict fired, or a prompt arises in implementation, a fix run, or planning
under `## decisions: recommended`.

### D6 — O1 moved to `/flow-self-review`, not run 2 step 9

KAN-854 moved the self-review filing ask out of run 2. The multi-select variant's one live call
site is **3. Explain, then ask** (`skills/flow-self-review/SKILL.md`), so the section sits there.

### D7 — K3 Stage keys → `skills/flow/stage-keys.md`

This is not README.md, because `names_test.go` parses README's Level 1 table and README is outside
the run-loaded corpus. It is not SKILL-rationale.md either, because the table is a lookup, not
reasoning. Citers repointed: SKILL.md, pipeline.md, README.md. `commands-claude/flow.md`'s
citation went with C3.

### D8 — rationale clauses split out of a sentence

P5, P9, P10–P12, P14, P15 and F4 each lose a "because"/"—" clause to the sibling
`-rationale.md`; the remaining sentence is listed as new text. The three REVIEW lines the guard
prints (P10, P14, F7) are reasoning, not rules: the token/`CLAUDE_CODE_SESSION_ID` fallback
mechanics, why `/clear` is safe, and why absent legacy keys are tolerated. The rule each explains
stays in the run file.

### D9 — rows that no longer apply on main

The audit rows that main had already resolved were skipped:
- K2, K4, P1, P4, M2, M3: fixed or removed by KAN-853/854 or the model change.
- K6: the `DEFAULT_MODEL` four-roles paragraph no longer exists. Its replacement, "Every other
  dispatch runs on its decision pair", is not stated elsewhere.
- F6: `models.default`'s stale "live consumer" line was already rewritten.

K11 is not in the issue's list and stays in SKILL.md.

### D10 — `check-verbatim-moves` gains a heading escape

`verbatim-moves.txt` treats `#` lines as comments, so no change could acknowledge a heading it adds
or cuts. A leading `\` is now dropped and the rest read literally. Two new cases cover it: an escaped
heading acknowledges; an unescaped one is still a comment. The first fails against the old code.

### D11 — model-policy.md keeps its override paragraph

M4 and M6 were cut as duplicates of SKILL.md, but they were `model-policy.md`'s only override
text, while its header and SKILL.md's "per **Model policy**" citation promise it. Both are kept;
only M5 is cut.

### D12 — `templates/CLAUDE.md` trimmed with `CLAUDE.md`

The template carries the same L2–L4 rationale and duplicates and seeds every new project's
always-loaded file, so it gets the same cut.

## Measurement

`scripts/load-sets.sh`, always-on block rendered sandboxed, base `8cc85ee5` → this change. Load sets
updated for the new files (auto-resolution under planning and implementation citations; standards
under implementation; visual, isolation and verdict files conditional).

| row | before | after |
|---|---:|---:|
| always-on (global block + CLAUDE.md) | 19,072 | 18,620 |
| router (command + SKILL.md + pipeline.md) | 46,133 | 32,350 |
| `project-configuration.md` | 51,536 | 13,096 |
| `state-file.md` | 21,311 | 14,398 |
| planning, incl. cited | 190,961 | 177,188 |
| implementation, incl. cited | 359,415 | 349,169 |
| finish, incl. cited | 220,402 | 203,696 |

Per session, including both contracts as `pipeline.md` directs:
- planning: 263,808 → 204,682 (−59.1 KB);
- implementation: 432,262 → 376,663 (−55.6 KB);
- finish: 293,249 → 231,190 (−62.1 KB).

Without the contract loads the savings are 13.8, 10.2 and 16.7 KB, below the issue's 19–24 KB
estimate because main had already cut part of the router before this change. The implementation
figure also absorbs the moves into `implement.md`, the standards file and the auto-resolution file,
which that session still reads.

## Review

An independent review pass found no Critical findings and two Important ones. Both were fixed
before landing, and are recorded above in D3:
- G9 left the core;
- the isolation directive missed run 2 step 7's by-hand path.

Of its Minor findings, these were fixed:
- the model-policy override (D11);
- run 2's dangling `state-file.md` citation, and its command-table wording;
- the two key rows' "specified below";
- `verify-and-handoff.md`'s by-hand citation;
- the visual directive's timing;
- a stray indent in `pipeline-rationale.md`.

These stay as they are:
- `state-file.md`'s "(`mainCheckoutPath` aside)": the field is defined in `state-file-internals.md`,
  and the exception still reads correctly;
- the multi-select section's "This contract fixes the shape": a verbatim move;
- `SKILL-rationale.md`'s "SKILL.md — Stage keys" heading: a rationale file no run loads.

## Verification

- `check-verbatim-moves`: 0 violations, 3 to review (D8).
- `check-references` and `check-installed-citations`: green. `commands-claude/flow.md` is declared
  expected-zero in both, as a stub that now cites no path.
- The rest of `## lint` is green, and the normative inventory is byte-identical.
- `go test ./...`: every package passes except `internal/guard`. Its 141 failing subtests fail
  identically on untouched `8cc85ee5` in this container, because they probe permissions and the
  tests run as root.
- `npx tsc -b` could not run: the container has no `node_modules`, and no SPA file changed.
