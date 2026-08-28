# Forward the reviewer's verbatim report to the fix agent

**Jira:** KAN-217 — self-review finding 2.2 from KAN-189
**Change:** `kan-217-myflow-forward-the-reviewer-s-verbatim-report`

This is the design as approved in brainstorming. The change's own
`spectre/changes/kan-217-myflow-forward-the-reviewer-s-verbatim-report/design.md` is canonical for
the decisions register, the alternatives ruled out, and the guard's resulting shape; this file does
not restate them.

## Problem

A panel slot reports to the dispatcher; the dispatcher distills the report into a one-line
`flow record finding -note` and discards the rest. Nothing under `skills/` persists a reviewer's own
words — verified by `grep -rn verbatim skills/`, which finds no report-forwarding mechanism at any
site.

At fix time, **Carry each surviving finding to the fix subagent as a structured block**
(`skills/flow/review-panel.md`) hands the fix subagent exactly that distillation — `F<n>`, slot,
severity, `file:line`, theme, reproducer text, bounces — and forbids inlining source beside it. The
fix agent therefore receives the dispatcher's verdict without the reviewer's evidence, and treats
the dispatcher's assertions as established fact. On KAN-189 the summary itself introduced findings
F20 and F23, which the implementer then built and documented as specified.

## The five parts

1. **Capture.** Each dispatched slot's returned report is written byte for byte to
   `<abs-worktree>/.superpowers/sdd/panel-report-<round>-<id>.md` at the point that slot's
   `flow record dispatch end` is recorded. Every dispatched slot writes one, including a clean slot
   and a substituted slot. A report that cannot be captured writes the file carrying
   `no verbatim report captured — <reason>`.

2. **Forward.** The fix-subagent dispatch prompt gains a **VERBATIM REPORT — THE FACT** blockquote,
   and the per-finding structured block gains its slot's report path as a field. The paragraph
   states that the file is the reviewer's own report unedited, that the block is direction and never
   a source of fact, and that the report wins where the two disagree. Only slots behind this round's
   surviving findings have their reports forwarded.

3. **`-note`.** **Recording findings** requires `-note` to carry the reviewer's own sentence naming
   the defect, not a dispatcher restatement — closing the paraphrase hole where it is created,
   rather than only downstream.

4. **Guard.** `scripts/check-reproduce-not-read.sh` becomes
   `scripts/check-dispatch-paragraphs.sh`, its single-label constants collapsed into a table of
   required dispatch paragraphs, so the new paragraph is guarded by the machinery that already
   guards REPRODUCE, DON'T READ. Block extraction, the `file:line` reporting shape and the
   `0`/`1`/`2` exit contract are unchanged.

5. **Registry.** `skills/flow-contracts/artifacts-registry.md` gains the row the report files
   require: created by `/flow`'s review panel, living in `<abs-worktree>/.superpowers/sdd/`, removed
   with the worktree at finish run 2.

## Scope boundaries

- **`stats/` is untouched** — no CLI flag, no schema, no migration, no render change. The report
  files are `.superpowers/`-local, matched by that directory's existing `.gitignore` entry, so they
  are never committed and never archived.
- **The per-task review in `skills/flow/implement.md` is out of scope.** It hands back to the same
  implementer, which already holds the context this change exists to stop re-deriving.
- **The bounce path is out of scope.** A bounced finding returns to the slot that authored the
  report.

## Testing

`scripts/test-check-dispatch-paragraphs.sh`, renamed from the existing harness and extended: cases
for the new paragraph's presence, for each of its required phrases missing, and for the site
missing the block entirely, plus a mutation test per required phrase, per KAN-197. Parts 1–3 and 5
are prose in skills and contracts, covered by the repository's existing `check-references.sh`,
`check-vocabulary.sh`, `check-markdown-integrity.py`, `check-contract-budget.sh`,
`check-installed-citations.sh` and `check-normative-inventory.sh` lint steps.
