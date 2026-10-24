## Context

- Six non-`big` findings of the kan-1022 self-review pass, planned per item 1 of **3. Fix every
  finding that is not `big`** (`skills/flow-self-review/SKILL.md`) on
  `self-review-kan-1022-templates-step-2-be-own-templates-crud-from`, based at `86c20900`
  (`origin/main`). Each finding is carried by exactly one task; the brainstorm design is
  `.superpowers/sdd/2026-10-25-self-review-kan-1022-templates-step-2-be-own-templates-crud-from-design.md`
  in this worktree.
  - A1-1 (Task 1): `verifyreport.go` sets `naAllowed` for steps 4, 10 and `motion` only, so steps 8
    and 9 can never be n/a. The open change `spectre/changes/verifier-completeness` holds only its
    `proposal.md` and `verbatim-moves.txt` and moves nothing on these steps.
  - A1-2 (Task 2): step 3 of `skills/flow/visual-verify.md` creates a worktree for the regression
    checkout alone; `kickoff-worktree.sh --from` already cuts from `origin/<base>` when the change's
    worktree records a flowBase.
  - A1-3 (Task 3): the rerun precondition in `skills/flow/visual-verify.md` is prose only. The
    kan-1022 store rows show it broken: `visual-verify` (blocked) at `dd71e0bb`, then
    `visual-verify-2` (completed, report `rerun of … failed specs at unchanged HEAD`) at `4497b60c`.
  - A4-1 (Task 4): signal nine (`stats/internal/guard/unfinishedwork.go`) clears a failure only by
    a later capture commit, a narrative line or a false-positive verdict.
  - A1-4 (Task 5) and A1-5 (Task 6): `skills/flow/review-panel.md`'s authoring pattern and mutation
    brief; no other file carries either section.
- History: eleven commits since 2026-10-24 touch these files, all already in `86c20900`; each
  task's new test fails at `86c20900` without its fix, so none of the six is already fixed.
- Spiked at plan time, every edit reverted: the six diffs apply in task order at `86c20900`
  (`git apply`); with all applied `go test ./internal/guard/` and `scripts/run-guard-tests.sh`
  pass, `check-installed-citations.sh`, `check-references.sh` and `check-markdown-integrity.sh`
  exit 0, and `check-verbatim-moves.sh` fails only on the five sentences this change's
  `verbatim-moves.txt` lists. `render-slot-prompt` renders the mutation brief, so the A1-5
  sentence names no `<…>` placeholder.

## Decisions

### Steps 8–10 may be n/a beside a wire contract

**ID:** wire-contract-na
**Status:** active
**Chosen:** the guard admits `n/a — <reason>` on steps 8, 9 and 10 when the report carries a
`- wire contract:` line whose value is not `n/a` — the line the verifier writes exactly when the
dispatch named a wire contract, i.e. named no view.
**Considered:** a fourth guard argument from the parent (`<views named>`) — a second optional
positional after `<scoped specs>` is ambiguous and every caller changes; reading the `- capture:`
line's `not run` — free text with no fixed shape.

### Repair the kan-1022 report in the same task

**ID:** repair-kan-1022-report
**Status:** active
**Chosen:** Task 1's last step rewrites the `- steps:` line of gymie's
`verify-report-visual-verify-2.md` (gitignored, in the kan-1022 backend worktree) to the n/a
statuses and re-runs the new guard on it; nothing is committed in gymie.
**Considered:** leaving the false `done` record — the defect wrote it, and a fix repairs what its
defect wrote (`skills/flow-self-review/SKILL.md`, step 3).

### A flowBase change gets a worktree in every `## apps` repository

**ID:** flowbase-apps-worktrees
**Status:** active
**Chosen:** one sentence in step 3 extending the regression-checkout recipe to every `## apps`
repository the set lacks, when the change's worktree records a flowBase.
**Considered:** a `check-visual-preflight` FAIL on a served main checkout lacking
`origin/<flowBase>` — it stops the stage where creating the worktree cures it; kept out as the
finding marked it optional.

### Refuse a rerun across HEADs at the finish gate

**ID:** rerun-same-head-gate
**Status:** active
**Chosen:** `check-visual-verify-dispatched` reports MISSING when the newest qualifying verifier
row's report (`.superpowers/sdd/verify-report-<key>.md`) carries a `- verify:` item naming `rerun
of` and the `visual-verify…` verifier row recorded just before it, of any outcome, carries a
different `-diff-base`; a report no longer on disk reads as no rerun. The rule text also voids
earlier passes on an environment cure that changes what the stack serves.
**Considered:** a new dispatch-time guard comparing the previous row's `-diff-base` before the
`rerun specs:` line is added — a new shim, Go guard and companion test for a check the existing
gate, which already reads every verifier row and refuses a scoped closing run, absorbs in one
block; prose alone — the precondition already was prose and was broken.

### Signal nine reads the next report's passing rerun

**ID:** signal-nine-rerun
**Status:** active
**Chosen:** a failure of report R is addressed when the report next after R by mtime opens its
`- verify:` item `exit 0 — rerun of` — a rerun takes exactly the previous report's failed specs.
**Considered:** any later report reading `verify: exit 0` — a scoped round-base run reads the same
and may not include the spec; any later `rerun of` report — a rerun after another failing report
re-ran that report's failures, not R's.

### First review effort and fixer grouping

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — Tasks 1, 3 and 4 change Go guard logic two gates read (the verifier
report gate and the run-1 finish gate), so the first review reads with medium effort.
Fixer groups, in series: group 1 is Tasks 1, 2 and 3 (visual verification: `verifyreport.go`, the
verifier contract, `visual-verify.md` twice, `visualverifydispatched.go`); group 2 is Tasks 4, 5
and 6 (`unfinishedwork.go`, `review-panel.md` twice).
**Considered:** `flow-low` — the diffs are spiked and small, but two finish-gate verdicts change;
one fixer for all six — one context carries both guard families for no gain; six fixers — each
pays start-up for a diff of a few lines.

## Open questions

## Live check

None — the change edits guard logic and prose, adds no store write beyond the verdict rows the
guards already record, and each task's Go test drives its guard against fixtures.
