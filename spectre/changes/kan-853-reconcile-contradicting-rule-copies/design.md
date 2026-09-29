# Design — kan-853-reconcile-contradicting-rule-copies

Every in-scope audit row is corrected to its canonical copy. This file records the rows where a
choice was made rather than read off a canonical file, and the rows left unchanged.

## Decisions

### D1 — the panel's fold route has one copy, in `review-panel.md`

Audit rows: implement #10 and review-panel B5.

- **Panel re-runs** (`skills/flow/review-panel.md`) is now the only copy.
- implement.md already cited it twice as "the route the branch's push state dictates".
- The rules only review-panel.md carried are tied to `FIX_BASE` and the reproducers:
  - the empty-fold drop;
  - "a clean autosquash is not evidence";
  - a guard-autosquash refusal stops the round.

  Moving them into implement.md would have meant rewording them.
- Two pieces moved in from implement.md: the `aside-planning-artifacts.sh <aside|restore>` wrap and
  the resolve-by-hand, keep-both-sides conflict rule.
- implement.md's restated route is now a citation.
- Not done: B5's related gap, the aside step for review-panel.md's base-movement rebases. That is
  new behaviour, so it belongs in KAN-854.

### D2 — the `+`-joined `-slot` sentence is deleted, not re-qualified

Audit row: review-panel B3.

- The qualifier that `ebaea04fa` dropped confined the sentence to the overlap between Primary and
  "Code review (low)".
- That same commit retired that slot, so a restored qualifier would govern nothing.
- The per-role rule and "No de-duplication across roles" stand.

### D3 — `begin` order follows the harvester

Audit row: review-panel B4.

- The harvester pairs a launch with the `begin` nearest it in time.
- A single dispatch records `begin` immediately before its launch.
- When launches go out together in one message, every `begin` is recorded in the next Bash call.
- Both files now state this.

### D4 — `-agent-id inline` is an explicit literal

Audit row: review-panel B9. "Never typed" is scoped to a slot's row; `inline` is the one typed
value, which implement.md already allowed.

### D5 — the principles template's standards fallback is removed

Audit row: subagent-prompts D3.

- review-panel.md's parent-side rule wins: with no `## standards` declared, the parent passes an
  empty value.
- The template's auto-detect of `CLAUDE.md` / `AGENTS.md` / `CONTRIBUTING.md` contradicted that
  rule. It also contradicted agent-baseline.md, which says the other harness's rendering is never
  read as well.

### D6 — generic project templates ship from `templates/`

Audit rows: every-session D21, D22.

- `setup.sh` copied this repository's own `CLAUDE.md` / `AGENTS.md` into any project that lacked
  one, carrying this repository's lint commands and dev-workspace rule. It now copies
  `templates/CLAUDE.md` / `templates/AGENTS.md`, which are generic.
- The template comment moved with them.
- The `CONTRIBUTING.md` sentence is gone, because that file never existed.

### D7 — pipeline.md's diagram gains the withdrawal line

Audit row: every-session D6.

- pipeline.md's own state table already has `(reachability end) → FINISHED (withdrawn)`.
- The diagram now shows it too, so the always-on rule's diagram is the correct copy, not the
  drifted one.

### D8 — `## decisions: recommended` covers the planning asks only

Audit row: every-session D20. The row is corrected to **Auto-resolution**
(`skills/flow-contracts/operator-prompts.md`). The pivot, outward-facing or irreversible actions,
and every integrate or archive prompt stay asked.

### D9 — `models` / `planningEffort` are legacy fields

Audit rows: every-session D18, D19 and verify D7.

- No run asks a model question or writes these fields.
- Each is carried forward as recorded and governs nothing.
- state-file.md, model-policy.md and verify-and-handoff.md now agree.

### D10 — the verify report quotes the last 20 lines

Audit row: verify D17. verify-and-handoff.md's `## Report` now says 20 lines, matching
implement.md's `| tail -20` rule for the parent's own `flow.verify` runs. visual-verify.md's
40-line allowance governs the verifier subagent's report, not the parent's, so it stays.

### D11 — the finish copies follow run 1 and run 2

Audit rows: finish D1–D7, D9, D13, D15–D21 and D24.

- Integrate has two test exceptions.
- The planning commit's *subject* is the fixed literal `chore(spectre): plan`, and the message
  also lists what the operator integrated over.
- Archive-branch push and delete behaviour follows run 2 step 10's two routes.
- "Merges nothing into the base branch" is scoped to before step 10.
- The stale `/flow-fast` override paragraph is replaced by what `/flow-fast` actually does.
- Related KNOWN-BUGS F13, the stale second filing site in `jira-followups.md`, is resolved and its
  entry removed.

### D12 — the §2 heading rename is coordinated

Audit rows: implement #1 and verify D16. `## 2. Isolate the workspace (first run only)` is now
`## 2. Isolate the workspace`, renamed together in implement.md, `stats/internal/stages/names.go`
and README Level 1. Every other citer already used the short form.

## Rows left as they are

- **every-session D10, second half.** The cited section already exists since KAN-852. Only the
  `flow.state-gate` reference was removed.
- **implement #7.** implement.md never stated a basename-suffix convention. The false citation
  was in verify-and-handoff.md, which now cites **The verifier dispatch**
  (`skills/flow/visual-verify.md`), where the suffix rule is stated.
- **Rows or parts of rows about the retired bugbot and security templates.** They are obsolete.

## Verification

- `## lint` passes, except `check-verbatim-moves.sh`, whose FAIL lines are exactly this change's
  intended rewordings. They are listed in `verbatim-moves.txt`.
- `check-normative-inventory.sh` output is unchanged against the pre-change capture.
