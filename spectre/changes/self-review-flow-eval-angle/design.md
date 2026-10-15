## Context

- One operator request, planned as one change per item 1 of **3. Fix every finding that is not
  `big`** (`skills/flow-self-review/SKILL.md`); its counted blast radius is 11 files, so it is not
  `big`.
  <!-- measured: cat the three task diffs | grep '^diff --git' | awk '{print $3}' | sort -u | grep -c . @ the planner's scratch clone at 5a1601fa (11) --> Every question was auto-resolved to its recommended option and is recorded below.
- The angle table is served, not copied: `check-self-review-report.sh` parses step 2's numbered
  table at run time (kan-585), and its Go tests read the real `SKILL.md`. A seventh row therefore
  turns every six-section fixture and every committed six-angle report red in the same commit that
  adds it, which is why Task 2 lands the row, the guard's second frozen list and the fixtures
  together.
- `docs/self-review/` holds 182 reports; 117 are on the frozen five-angle list and the other 65
  carry six angles.
  <!-- measured: ls docs/self-review | grep '\.md$' | grep -v -- '-context.md$' | grep -vxFf docs/self-review/five-angle-reports.txt | grep -c . @ 5a1601fa -->
  Only this repository runs the guard: gymie, gymie-frontend and flow-eval keep self-review reports
  but declare no `check-self-review-report.sh` in their `## lint`.
  <!-- measured: grep -c check-self-review-report <repo>/.flow/project.md for gymie, gymie-frontend, flow-eval (each 0) @ each main checkout on 2026-10-15 -->
- The self-review bundle (`flow self-review bundle`) serves the archived `tasks.md`, `design.md` and
  `narrative.md` but not `visual-verification.md`, the record that names a visual pair's capture,
  its mockup frame and the verdict — e.g. gymie's archived kan-982 carries it beside its
  `visual-verification/evidence/` PNGs, and the live bundle for that change reads `found: 6 of 7
  sources` with no such section.
  <!-- measured: cd gymie && flow self-review bundle -change kan-982-prod-fixes-group-add-person-tap-set-value-bleed | sed -n 3p @ the flowd store on 2026-10-15 -->
- flow-eval's miners read the change's archive from the `--plan-repo` checkout's filesystem
  (`ledger.md`, `panel.md`) and its commits from `--repo`, and stamp `case.json`'s `source.repo`
  with the `--repo` directory's basename, which `flow-eval run --repos <dir>` resolves as
  `<dir>/<source.repo>` (`<agents repo>/../flow-eval/internal/mine/mine.go`, `Impl` and `Review`).
  A case's `snapshot` is a sha of the source repository, so it must survive the change's landing.
- `~/.claude/hooks/protect-main-checkout.py` denies `git commit` in a main checkout on its default
  branch; flow-eval's main checkout sits on `main`, so mining commits in a worktree.
- `<state-dir>` is a root `check-installed-citations.sh` already recognises, and `flow state dir`
  prints it per repository (`/Users/tweety53/Agents/flow/state/gymie-7c1f238a` for gymie).
  <!-- measured: cd gymie && flow state dir @ 2026-10-15 -->
- Two citations name "step 5 of `/flow-self-review`" for the landing push's sha re-read
  (`skills/flow-contracts/finish-contract-run1.md`, `skills/flow-fast/SKILL.md`); that rule is item 5
  of step 3 — step 5 is filing. Fixed in Task 3, beside the lines it already touches.
- Every task's diff was applied in plan order to a scratch clone at `5a1601fa` and reproduces the
  working tree it was cut from: `gofmt -l` and `go vet` clean; `go test ./internal/guard/
  ./internal/selfreview/` and `scripts/run-guard-tests.sh` (41 harnesses) pass; each new test
  fails with its production change reverted; `check-self-review-report.sh` passes over the real
  `docs/self-review/`; `check-verbatim-moves.sh` (with this change's `verbatim-moves.txt`),
  `check-references.sh`, `check-markdown-integrity.py`, `check-installed-citations.sh`,
  `check-dispatch-paragraphs.sh`, `check-ask-silence.sh`, `check-guard-symlinks.sh` and
  `check-stage-mark-calls.sh` exit 0, and `check-normative-inventory.sh`'s output is unchanged.

## Decisions

### A candidate is recorded noted

**ID:** candidate-disposition-noted
**Status:** active
**Chosen:** an angle-7 candidate names no defect, so it takes the existing `noted` disposition in
the store and in the report, mined or not; its report line names its kind and whether the operator
chose to mine it, and its template's fields sit beneath it as indented sub-bullets, which the guard
reads as prose. The store's `angle` column is free text, so `flow-eval` needs no schema change.
**Considered:** a new `mined: <sha>` disposition — it touches the guard, the store's CHECK
constraint (a migration), its validation, the CLI help and the SPA's type and view for a fact the
mining branch's commits already record; `declined` for an unchosen candidate — `declined` means a
real defect deliberately neither fixed nor filed (migration 0037); a one-line report entry with no
template — the template is the operator's mandatory record and would survive only in a scrolled-off
message body.

### The mining ask joins step 4's call

**ID:** mining-ask-in-step-4-call
**Status:** active
**Chosen:** mining candidates are offered items in step 4's one `AskUserQuestion` call, in
questions of their own — a question holds findings to file or candidates to mine, never both — each
with **None — mine nothing** as its default; the nine-item cap and the rating-less overflow call
count every offered item. Silence, a stalled prompt and auto-resolution mine nothing, and
`operator-prompts-auto-resolution.md` lists the mining among the outward-facing acts.
**Considered:** mixing filing and mining options in one question — one None option cannot be the
default for both acts; a separate call for mining — the operator asked for it to join step 4's call
or its overflow.

### Mining waits for the change's commits to be final

**ID:** mine-after-landing
**Status:** active
**Chosen:** a chosen candidate is mined once the change's commits are on its repository's default
branch: standalone at once (the default branch is already checked out), in a run after the landing
push of item 5 of step 3 on merge and push. On a pull-request or manual route nothing is mined and
the final report names each chosen candidate's `case add` command for after the merge. In a run,
`--repo`/`--plan-repo` are the change's worktrees (the main checkout may be stale after landing),
and `case.json`'s `source.repo` is set to the main checkout's directory name.
**Considered:** mining at step 5 whatever the route — a re-sync after a rejected push or a
squash-merged pull request rewrites the shas a case's `snapshot` names, leaving a case pointing at a
commit the source repository never keeps; deferring to run 2's cleanup — a pull-request route never
reaches run 2 in the same invocation, and run 2 removes the worktrees the miner reads.

### Mining commits on a flow-eval worktree branch, never landed

**ID:** mine-on-flow-eval-branch
**Status:** active
**Chosen:** one worktree and branch per pass, `flow-eval-cases-<name>` at
`<agents repo>/../flow-eval-worktrees/`, based at `origin/main`; `go run ./cmd/flow-eval case add …`
per candidate, the draft edited per `docs/cases.md`'s "What to check and edit", `case validate`,
one commit per case as `feat(cases): add <role> case <id>`; the branch is never merged or pushed,
so the operator's review of it is the label approval flow-eval's README names.
**Considered:** committing on flow-eval's main checkout — `protect-main-checkout.py` denies it, and
a direct commit to `main` skips the label review the README makes the commit review; pushing the
branch — an outward-facing act the operator did not ask for; the prebuilt `flow-eval` binary in the
repository root — it can lag the checkout, where `go run` builds what the worktree holds.

### flow-eval is found beside the agents repository

**ID:** flow-eval-beside-agents
**Status:** active
**Chosen:** `<agents repo>/../flow-eval`, the layout flow-eval's own README setup requires (the
agents repository at `../agents`); absent, the pass prints `⚠ flow-eval: skipped — <reason>` and
mines nothing, as Jira's absence skips filing.
**Considered:** a `.flow/project.md` key naming the checkout — configuration nobody asked for, for a
path flow-eval already fixes.

### Extracted files go under the state directory

**ID:** durable-extract-state-dir
**Status:** active
**Chosen:** a file that exists only in git history, or only in a worktree cleanup removes, is
extracted with `git -C <repo> show <sha>:<path>` into `<state-dir>/flow-eval/<name>/` —
`<state-dir>` as `flow state dir` prints it — before step 4's body is written, and cited there.
**Considered:** committing the PNGs beside the report in `docs/self-review/` — binary files in a
text record for a reference only the operator's review reads; a temporary directory — not durable,
which the operator required.

### Earlier reports are frozen lists

**ID:** six-angle-frozen-list
**Status:** active
**Chosen:** a second frozen list, `docs/self-review/six-angle-reports.txt`, holds every report
written before angle 7 to the first six angles; the guard reads the lists in table order
(`srrFrozenLists`), a report on both taking the first. The list is the reports present at Task 2's
parent not on the five-angle list; a rebase onto a moved base adds the six-angle reports landed
since, in the same fold.
**Considered:** demanding angle 7 only of a report carrying one — the operator wants every angle
explicit, and an omitted angle would pass silently again; adding an empty angle-7 section to the 65
reports — it rewrites records of passes that never ran the angle, a 65-file blast radius.

### The bundle serves visual-verification.md

**ID:** bundle-serves-visual-verification
**Status:** active
**Chosen:** `archivedFiles` in `stats/internal/selfreview/git.go` gains `visual-verification.md`,
served or reported skipped like the other archived files; angle 7 reads beyond the bundle only for
the commits and files the bundle names, to quote code and resolve paths.
**Considered:** letting the pass read the archive directory itself — it breaks "the pass covers what
the bundle holds", which keeps a deferred pass reproducible from its bundle.

### Product code is in angle 7's scope

**ID:** flow-eval-angle-reads-product-code
**Status:** active
**Chosen:** angle 7 reads its candidates from the run's own work in any repository the run touched;
the product-code exclusion governs defects, and a candidate is not one.
**Considered:** keeping the exclusion — every gymie candidate is product code, so the angle would be
empty on every product run.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `medium` — Task 2 changes a guard that runs on every self-review report, and Task 3
writes the contract for an outward-facing act (commits in another repository) behind an operator
ask; a low-effort read could miss a path where a candidate is mined without approval.
**Considered:** `low` — all three diffs are spiked and their tests pass, but Task 3 is prose a test
cannot check.

### Fixer grouping

**ID:** fixer-groups
**Status:** active
**Chosen:** one group, one fixer, tasks in plan order — g1: Task 1, Task 2, Task 3.
**Considered:** one group per task — Tasks 1 and 3 share `finish-contract-run1.md` and
`skills/flow-fast/SKILL.md`, and Tasks 2 and 3 share `skills/flow-self-review/SKILL.md`, so the
fixers would run in series anyway, each paying its own setup.

## Open questions

## Live check

Task 1 changes what the running `flowd` serves. Once the change has landed and the daemon runs a
build carrying it, run `flow self-review bundle -change
kan-982-prod-fixes-group-add-person-tap-set-value-bleed` from the gymie main checkout. Before:
`found: 6 of 7 sources` and no `visual-verification.md` section. After: `found: 7 of 8 sources`
and a `## spectre/changes/archive/kan-982-prod-fixes-group-add-person-tap-set-value-bleed/visual-verification.md`
section opening `# Visual verification — kan-982-prod-fixes-group-add-person-tap-set-value-bleed`.
Still `of 7 sources` means the daemon is serving an older build — not a pass. The angle, the ask and
the mining are prose; their first live exercise is the next self-review pass.
