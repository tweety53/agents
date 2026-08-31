# kan-277-a-panel-re-run-should-read-a-since-last-review

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.

> **Relocation:** no

One task. It is prose-only — `check-panel-diff-size.sh` needs no interface change (per
`design.md`'s `no-interface-change-to-the-guard`), so the whole fix is a new subsection in
`skills/flow/review-panel.md`'s `## Panel re-runs`.

- [x] 1. Add the re-run diff-size cap check to the panel re-runs contract

  In `skills/flow/review-panel.md`, insert a new subsection immediately after the paragraph
  that ends "…and handoff still requires **zero open findings at any severity** from every
  agent that has run." (the paragraph right after the "Escalate automatically" paragraph, and
  right before "Union all **open** findings, dedupe by…"):

  ```markdown verified:authored in-tree for this change
  ### The diff-size cap check on a re-run

  **The cap check re-runs too, gated on the per-slot max this round, never the aggregate branch
  diff.** Pass 1's own check (**The roster**, above) is unaffected: every slot reads the full diff
  there, so the full-branch count already is the per-slot max.

  **On a Full re-run:**

  1. If Primary, or any other slot dispatched this round with no held last-reviewed sha, is in
     this round's dispatch, run `check-panel-diff-size.sh <worktree> <merge-base> <cap>` — the
     same full-branch count pass 1 measures.
  2. For every returning slot dispatched this round against a scoped delta, run
     `check-panel-diff-size.sh <worktree> <that slot's last-reviewed sha> <cap>`, once per
     **distinct** sha among this round's dispatched slots (two slots sharing a sha need one
     call, not two).
  3. **The gating count is the largest of every count from 1 and 2** — the largest single read
     any one agent this round actually faces. An exit-1 result from whichever call produced that
     count is what puts the over-cap choice to the operator (**The roster**, above), naming the
     gating count and, when it is not the full-branch count, the full-branch count too, for
     context.

  Record both the gating count and (when computed and different from it) the full-branch count
  in `<abs-worktree>/.superpowers/sdd/final-review-panel.md` for this round, alongside the
  existing mode/agents/why fields **Panel re-runs** already records.

  **On a Targeted re-run**, run `check-panel-diff-size.sh <worktree> <FIX_BASE> <cap>` once —
  `fix-round-N.diff`'s own range is the only diff any re-run slot reads this round, so its count
  is directly the gating count; there is no per-slot max to take. **Targeted carries no cap
  check today**; this is the same over-cap prompt as Full and pass 1, gated on this one number,
  closing that gap.
  ```

  Verify: `grep -n "diff-size cap check on a re-run" skills/flow/review-panel.md` finds the new
  subsection; `scripts/check-references.sh` and `scripts/check-plan-shape.sh` both exit 0;
  `scripts/check-plan-provenance.sh spectre/changes/kan-277-a-panel-re-run-should-read-a-since-last-review/tasks.md`
  exits 0. Then run the project's `## lint` list and confirm it is clean.

**Build:** green
**Files:** `skills/flow/review-panel.md`
**Tests:** none — prose-only change to a review-panel contract describing an interactive/dispatch
procedure; no automated harness parses this section's text, matching kan-373's task 3 precedent
for prose-only contract edits (`spectre/changes/archive/kan-373-archive-step-follows-landing-route-default/tasks.md`).
**Regression:** reverting this task returns the cap check to gating every re-run on the aggregate
full-branch diff, reproducing the repeated over-cap prompting KAN-77 hit — a silent behavior loss
with no test to catch it, since none exists for this prose.
**Baseline:** n/a — no test declared.
**Commit:** `docs(flow): gate the review panel's re-run diff-size cap on the per-slot max`
