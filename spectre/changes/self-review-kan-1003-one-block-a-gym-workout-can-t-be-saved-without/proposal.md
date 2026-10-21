# self-review-kan-1003-one-block-a-gym-workout-can-t-be-saved-without

## Why

kan-1003's self-review pass found seven pipeline defects, none `big`:

- **A (flow-fix).** `check-done-when-paths` judges a branch name as a missing path once no local or
  remote-tracking ref carries it. A clone of gymie `develop` without the feature branches exits 1
  with 8 `DONE-WHEN-PATH` lines (`feature/cardio` twice, `feature/workout-rows` six times), so
  every gymie archive stops once those branches are deleted; kan-1003 stopped on
  `feature/one-block`.
- **B (flow-fix).** Planning recorded decision `lands-on-develop` (gymie plan-gate commit
  6f3cff23, 2026-10-21) against the handoff ticket's Done-when ("Branch `feature/one-block` in both
  repos, never develop") and left the ticket unamended, so integrate stopped and asked again. The
  ticket's line now reads "Landed on `develop` (operator decision at integrate, 2026-10-22)." —
  false: the plan made that decision on 2026-10-21.
- **C (flow-fix).** The planning session committed gymie's design handoff 0497d943 and pushed it
  in the same command, without the `Flow-Key: flow.brainstorm` trailer; the push check ran before
  the commit existed. Integrate's gate listed it `UNTASKED`, the operator was told it was their
  own commit, and `flow record verdict false-positive` flagged verdict 646 — a `CLEAR` row of
  `check-unfinished-work`, since `check-untasked-commits.sh` records no verdict.
- **D (flow-fix).** `hooks/protect-main-checkout.py` reads a quoted `>` as a redirect: from
  gymie's main checkout it denies `python3 grab.py "<worktree file>" '>Save<' 1 1200` and a
  `python3 - <<'EOF'` body containing `s.find('>Save<')`.
- **F (flow-cost).** The visual verifier inferred `flex: 1 1 auto` from pixels while the handoff's
  HTML declared `grid-template-columns:1fr 1.3fr 1fr; height:32px`; four fix rounds were spent on
  it.
- **G (flow-automation).** `full-app-suite.zip` went stale twice: its rebuild exists only as inline
  bash in step 8 of `skills/flow/visual-verify-verifier.md`, and an in-run recapture never runs it.
- **H (flow-stats-app).** Inline dispatch rows were recorded after the fact: back-fills with a
  guessed `-started-at $(date -u -v-45M …)`, and rows for tasks 10–22 opened with `begin` after
  their task commits landed, closing at 3–53 s with no commit. Only the 3-second floor checks.

## What changes

- `check-done-when-paths` reads a dotless name whose first segment is no tracked directory as a
  branch; gymie's `develop` clone passes.
- `skills/flow/brainstorm-planner.md` carries the Done-when amendment rule; gymie's KAN-1003 ticket
  cites the plan decision `lands-on-develop` of 2026-10-21.
- The hook denies a command that commits and pushes a `spectre/` branch when that commit's stated
  message carries no trailer; the gate never flags an `UNTASKED` listing; `flowd unflag-verdict`
  clears verdict 646.
- The hook judges only standalone redirect operators — `>|` and `&>` included — so a quoted `>`
  is data.
- The verifier's departure lines cite the handoff's HTML/CSS declaration for an expected value.
- `scripts/rebuild-full-app-suite-zip.sh` rebuilds the zip for the verifier, every in-run
  recapture and the base sync's recapture.
- The store refuses an inline task row closed with no commit, or with a commit made before its
  begin; an inline back-fill carrying `-commit` derives its instants; `flowd retime-dispatch`
  and `flowd restamp-dispatch` repair kan-1003's 17 mis-timed rows.
