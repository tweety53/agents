# self-review-backlog-flow-8

## Context

- Five Backlog Flow findings — the next five live ones not already fixed, oldest first: KAN-821,
  KAN-831, KAN-837, KAN-883, KAN-885 — planned as one change per item 1 of **3. Fix every finding
  that is not `big`** (`skills/flow-self-review/SKILL.md`). Brainstorm design:
  `.superpowers/sdd/2026-10-19-self-review-backlog-flow-8-design.md` in this worktree. Every
  question was auto-resolved to its recommended option
  (`skills/flow-contracts/operator-prompts-auto-resolution.md`).
- **Not in this change.** KAN-822, KAN-824, KAN-827, KAN-832, KAN-833, KAN-834, KAN-835 and
  KAN-836 were already fixed and moved to Done before this plan. KAN-764 is a design ask, a frame
  only the operator can draw, and stays in Backlog Flow. Two candidates were found already fixed
  on `origin/main` while verifying:
  - **KAN-878** — `check-verbatim-moves` FAIL lines print the whole sentence: both FAIL
    `Fprintf`s (`stats/internal/guard/verbatimmoves.go:139,151`) print `vmAckLine(s)`, the full
    sentence escaped for the acknowledgement file (`7d32f7a6`); only the passing `ok` lines go
    through `vmShort`'s 100-rune cut. The Go guard has printed FAIL sentences whole since its
    port (`5bc5c130`, `verbatimmoves.go:131,143` there).
    <!-- measured: scripts/check-verbatim-moves.sh 9df14302 in a scratch clone with one long run-loaded sentence removed and one added @ 9df14302 → a 189-character and a 176-character FAIL line, each sentence whole -->
  - **KAN-879** — the premise "a micro panel dispatches nothing" no longer holds: since
    `b74e4d5c` every pass 1 dispatches **The always-on extra dispatch**, "on a decided and a
    `default` panel alike" (`skills/flow/review-panel.md:241-242`), recorded as one `dispatches`
    row, and `/flow-fast`'s `default` panel runs it too (`skills/flow-fast/SKILL.md:112`). A
    completed panel with no dispatch row is therefore a real records loss on every class, so the
    note stays.
    <!-- measured: read-only query of the dev store's decisions and dispatches tables, decisions with panel "default" recorded after 2026-10-15 — a machine-local observation, not re-runnable at a ref → kan-991 7 dispatch rows (1 test-sanity), kan-960 4 (1 test-sanity) -->
- **KAN-821 — Task 1.** No skill scopes a frame a sibling ticket finishes: **D. Basic Workflow #3
  — Writing plans** asks a capture and sidecar line per frame, and the verifier's attribution
  paragraph (`skills/flow/visual-verify-verifier.md`) knows three outcomes — pre-existing, this
  change's, mis-measured. Source: gymie `docs/self-review/kan-743-…` (decision
  `p3-p4-strip-parts-only`; gymie-playwright `bd59623`'s `fidelity/pinned-set-strip.spec.ts.mockups`
  comment). Blast radius: two files.
- **KAN-831 — Task 2.** `runReproducer` takes its bound from `Env.ReproducerBound`, then
  `RUN_REPRODUCER_BOUND_SECONDS`, then `rrDefaultBnd` (`stats/internal/guard/runreproducer.go:262`);
  the reproducer's own file is read only for `# mutation-reproducer`. **review-panel-fix-round.md**
  asks every caller to prefix the knob. Blast radius: five files — the guard, its test, the
  script header and the two panel files; no file cites a changed name or section.
- **KAN-837 — Task 3.** **The loop**'s shared-source paragraph (`skills/flow/verify-fix-loop.md`)
  recaptures every affected spec with `--update-snapshots=all`, which writes every capture the
  spec takes — an unrelated failure's included. Source: gymie
  `docs/self-review/kan-748-menu-glyph-more-visible-round-dots-24dp-full-ink-self-review.md`.
  Blast radius: one file.
- **KAN-883 — Task 4.** `checkWorktreeLocation` (`stats/internal/guard/worktreelocation.go`)
  counts every linked worktree outside `<project>-worktrees/` as STRAY. The strays the source runs
  met were Claude Code subagent worktrees under `<project>/.claude/worktrees/` (kan-839's
  narrative: `.claude/worktrees/agent-ae1ef826791eba5ee`), locked by their live session. Blast
  radius: three files.
- **KAN-885 — Task 5.** `planClass` passes the caller's `<merge-base>` straight to
  `git diff <merge-base>` and `panelTouchedPaths`' `<merge-base>..HEAD`
  (`stats/internal/guard/planclass.go:255-264`); both compare trees, so a base ref that moved
  after the worktree was cut counts the base's new commits. Source:
  `docs/self-review/kan-781-record-failed-approaches-in-the-run-narrative-self-review.md`. Blast
  radius: three files.
- **Task 6 — found while verifying.** `TestNoPipeIntoEarlyExitGrep` fails on `origin/main`:
  `scripts/test-check-task-records.sh:392,411` pipe `printf` into `grep -q` (added by
  `646ac917`). Fixed in this change per the sweep. Blast radius: one file.
  <!-- measured: cd stats && go test ./internal/guard -run TestNoPipeIntoEarlyExitGrep in a scratch clone @ 9df14302 → FAIL, test-check-task-records.sh:392 and :411 -->

## Decisions

### A sibling's parts are declared expected in the sidecar comment

**ID:** sibling-parts-sidecar-comment
**Status:** active
**Chosen:** a **D.** paragraph beside the capture-per-frame one: the frame's built parts are a recorded decision, and its sidecar line stays, under a `#` comment naming the expected-to-differ parts, the sibling ticket and the decision ID; the verifier's attribution paragraph reads a declared part as `expected — <sibling ticket>, <decision ID>`, never blocking, and judges every other part as usual — kan-743's practice, made the rule.
**Considered:** dropping the frame's sidecar line, as kan-743 did — the verifier's own rule blocks a frame on the change's list with no sidecar line, and the built parts would go uncompared; a sidecar syntax the compose script parses — the comment is read by the verifier, not by a script, and `compose-mockup-frames.sh` already ignores `#` lines.

### A reproducer declares its own bound

**ID:** reproducer-declared-bound
**Status:** active
**Chosen:** `# reproducer-bound: <seconds>`, a positive whole number on an exact line in the first 10 lines — the window `# mutation-reproducer` uses — read by `rrDeclaredBound` and passed to `rrDeadline` as the default; the injected test bound and `RUN_REPRODUCER_BOUND_SECONDS` still win; a malformed declaration leaves 20 seconds. The panel files ask authors for the line and drop the per-call knob.
<!-- measured: grep -n -e 'rrDefaultBnd = "20"' -e 'headLines(path, 10)' stats/internal/guard/runreproducer.go @ 9df14302 → line 41 rrDefaultBnd = "20", line 503 headLines(path, 10) -->
**Considered:** the header line overriding the knob — the knob is the operator's and the harness's instrument, and a declared bound only needs to beat the default; removing the knob — the guard tests drive their timeout cases through it (`check_panel_reproducer_exit_contract_test.go`, `runreproducer_test.go`), and an operator may still need to raise one run's bound; naming `rrDeclaredBound`'s parameter `path` like `declaresMutation`'s — `check-mutation-reproducer-pin.sh` reads every `range headLines(path, <N>)` as the mutation window and refuses two, so the parameter is `file`.

### Re-baseline per capture by blob classification

**ID:** per-capture-blob-rebaseline
**Status:** active
**Chosen:** when the suite also fails for a reason the fix did not cause, run it without `--update-snapshots`, classify each failing capture's diff-image blobs against the fix's expected footprint, copy only a fully fitting capture's `-actual.png`, read each before staging, repeat since a test stops at its first failing capture, and route an outside-footprint blob to **The sweep** — kan-748's procedure.
**Considered:** a Go command that reads each diff PNG and classifies its blobs against a declared footprint (KAN-837's `flow-automation` half) — the footprint is a per-fix judgement with no declared format today, and kan-748 ran the procedure by hand in three passes; keeping `--update-snapshots=all` and reading every rewritten PNG — reading catches a wrong screen, not an unrelated defect baked in under the fix's own pixels.

### A harness agent worktree is FOREIGN, not STRAY

**ID:** harness-worktree-foreign
**Status:** active
**Chosen:** a linked worktree under `<project>/.claude/worktrees/` prints `FOREIGN: <path> (<branch>)` and is never counted; every other path outside `<project>-worktrees/` stays STRAY. `check-finish-preflight` relays only `STRAY: ` lines, so it needs no change.
**Considered:** scoping to worktrees whose branch carries `branch.<name>.flowBase`, as KAN-883 proposed — kickoff writes flowBase only when an explicit base is given (`scripts/kickoff-worktree.sh:7-17`, `skills/flow-fast/SKILL.md:140` "--base only"), and none of `<agents repo>`'s twelve linked worktrees carries one, so the guard would check almost nothing; exempting every locked worktree — a flow worktree misplaced by a locked session would pass; printing exempted ones as `STRAY:` — `check-finish-preflight` relays every `STRAY: ` line into its refusal.
<!-- measured: git -C /Users/tweety53/Projects/agents config --get-regexp 'branch\..*flowbase' and git worktree list --porcelain — a machine-local observation, not re-runnable at a ref → no flowBase key, twelve linked worktrees -->

### plan-class measures from the fork point

**ID:** plan-class-fork-point
**Status:** active
**Chosen:** after validating `<merge-base>`, `planClass` replaces it with `git merge-base HEAD <merge-base>` for both the touched surface and the numstat; a recorded worktree sha is its own fork point, a moved ref yields the fork point, and no shared history is exit 2.
**Considered:** reading the worktree's `flowBase`, as KAN-885 proposed — absent on most worktrees (see `harness-worktree-foreign`) and a branch name moves as `origin/main` does; moving the fork-point step into the shared `panelValidateWorktree` — its four other callers pass a recorded sha or round base by contract, and the change would widen to their tests for no reported defect.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** flow-medium — Tasks 2, 4 and 5 change guard behaviour every review panel, kickoff and Decide run through.
**Considered:** flow-low — Tasks 1, 3 and 6 are prose or a two-line test fix, but the three Go tasks change runtime behaviour a low-effort pass could under-read.

### One fixer group

**ID:** fixer-groups
**Status:** active
**Chosen:** one group, Tasks 1–6 in order, one fixer dispatch (`self-review-backlog-flow-8-fix`): every task is a `git apply` of a diff verified in a scratch clone, and the files are disjoint.
**Considered:** one group per task — six dispatches for work one fixer finishes in series; a Go group and a prose group — saves no wall time, the prose tasks are minutes.

## Open questions

## Live check

No running service or persistent state is touched: the three Go changes are `flow-guard`
subcommands built per call from the checkout by `scripts/lib/flow-guard.sh`, and the rest is prose.
