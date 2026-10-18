# self-review-backlog-flow-1

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

**Goal:** close the three live Backlog Flow findings — KAN-539's commit message describing an
abandoned attempt, KAN-545's coordinate-driven fidelity specs, KAN-549's unreconciled frame list.

**Architecture:** two prose rules, each where its readers already look (`git-boundaries.md` cited
from the implementer prompt; writing-plans section D cited from the verifier's capture step), and
one new `flow-guard check-frame-list` subcommand behind a `flow_guard_exec` shim, run by
`flow.visual-verify`'s parent.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-18-self-review-backlog-flow-1-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/`. Never merge or push.
- Work in this worktree `/Users/tweety53/Projects/agents-worktrees/self-review-backlog-flow-1`, branch `self-review-backlog-flow-1`; never touch a main checkout.
- Each task's diff is the unified diff its Step 2 gives, applied with `git apply` from the worktree's root — applied and checked in order in a scratch clone at `7aea62c2` (`design.md`, **Context**). Save it by copying the lines between its fence markers into a scratch file outside the worktree. A diff that does not apply is reported, never hand-merged by guesswork.
- `spectre/changes/self-review-backlog-flow-1/verbatim-moves.txt`, committed with this plan, acknowledges every sentence the three diffs add, so `scripts/check-verbatim-moves.sh` passes once each task lands.
- Run `gofmt -w` on every Go file you change before committing; `gofmt -l` and `go vet` stay clean. Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`.

## Review Focus

- `check-frame-list` must not declare a frame `design.md` names only inside a longer id (`Q100` never declares `Q10`; `UNCITED-old` never declares `UNCITED`) — case "every declared frame captured and reported".
- A declared frame the verifier names absent with its reason on the `frames:` line is not a gap — case "a declared frame named absent with its reason".
- A `frames:` total above the declared count is not a gap: the verifier's list legitimately adds diff-changed and `moved:` frames.
- The structural-locator rule must not forbid the verifier's own `page.mouse.move(0, 0)` pointer park, which clicks and captures nothing.

- [ ] 1. A commit's message describes the tree it lands, never an abandoned attempt (KAN-539)
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -n "describes the tree it lands" skills/flow-contracts/git-boundaries.md skills/flow/implement.md; echo $?`:
    it prints only `1` — no rule says a commit's message describes its net tree.
    <!-- measured: the grep above in this worktree @ 7aea62c2 → exit 1, no line -->
  - [ ] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to a scratch clone at 7aea62c2; check-references.sh, check-markdown-integrity.py and check-verbatim-moves.sh 7aea62c2 (with this change's verbatim-moves.txt) exit 0
diff --git a/skills/flow-contracts/git-boundaries.md b/skills/flow-contracts/git-boundaries.md
index 12736743..3ef218f5 100644
--- a/skills/flow-contracts/git-boundaries.md
+++ b/skills/flow-contracts/git-boundaries.md
@@ -145,3 +145,11 @@ the artifact instead of re-deriving the rejected position from indirect evidence
 case that paid for it: a run whose design space held a reverted prototype re-derived the
 prototype's intended polarity from contract wording because the revert's body recorded only
 `This reverts commit <sha>` — that the code left, never why.
+
+**A commit's message describes the tree it lands, never an attempt that left it.** When work
+written inside a task or a fix round is undone before it lands — reverted in the working tree, or
+removed by a fixup folded into the commit — the commit's subject and body describe the net tree,
+and the body names the abandoned attempt only as abandoned, with why. To keep the attempt in the
+history instead, land it and back it out as two commits, the back-out under the paragraph above.
+KAN-539 is the case that paid for it: a guessed icon and its revert were folded into one commit
+whose message still described the icon its tree did not contain.
diff --git a/skills/flow/implement.md b/skills/flow/implement.md
index d38201dc..6e3c62fa 100644
--- a/skills/flow/implement.md
+++ b/skills/flow/implement.md
@@ -490,7 +490,9 @@ Every implementer dispatch **must** carry:
 > task is backed out, not erased**: when work an earlier commit landed is withdrawn, the
 > backing-out commit's body carries the rejection reason — what was wrong, and what would have to
 > change for it to come back (**Revert and back-out commits**,
-> `skills/flow-contracts/git-boundaries.md`).
+> `skills/flow-contracts/git-boundaries.md`). **Your commit's message describes the tree it lands**: an attempt you
+> wrote and undid before committing is never what its subject or body describes (**Revert and
+> back-out commits**, `skills/flow-contracts/git-boundaries.md`).
 
 **A `Build: red` task is dispatched with its `Squash-with:` partner, in one bundle.** The
 implementer runs the red task first — writes its tests, runs them, and reports the failing
````

  - [ ] **Step 3: Verify.** `grep -c "describes the tree it lands" skills/flow-contracts/git-boundaries.md skills/flow/implement.md`
    prints `1` for each file; `scripts/check-references.sh`, `python3 scripts/check-markdown-integrity.py`
    and `scripts/check-verbatim-moves.sh` exit 0.
    <!-- measured: grep -c after git apply in a scratch clone @ 7aea62c2 → 1 and 1 -->
  - [ ] **Step 4: Commit** the two files with the subject below.

**Build:** green
**Files:** `skills/flow-contracts/git-boundaries.md`, `skills/flow/implement.md`
**Tests:** **none** — a prose rule; `check-references.sh` and `check-verbatim-moves.sh` cover its citation and its acknowledged sentences.
**Regression:** none — no test is declared.
**Commit:** docs(flow): a commit's message describes the tree it lands, never an abandoned attempt
**After:** none

**Decision:** net-tree-message

- [ ] 2. Reach every control in end-to-end and fidelity specs by a structural locator (KAN-545)
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -n -i "structural locator" skills/flow/brainstorm-planner.md skills/flow/visual-verify-verifier.md; echo $?`:
    it prints only `1` — no rule tells a plan or a verifier how a spec reaches a control.
    <!-- measured: the grep above in this worktree @ 7aea62c2 → exit 1, no line -->
  - [ ] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to a scratch clone at 7aea62c2 after Task 1's diff; check-references.sh exits 0 (the verifier's citation resolves to section D's heading), check-markdown-integrity.py and check-verbatim-moves.sh 7aea62c2 exit 0
diff --git a/skills/flow/SKILL-rationale.md b/skills/flow/SKILL-rationale.md
index d6a2a269..3a7f04f4 100644
--- a/skills/flow/SKILL-rationale.md
+++ b/skills/flow/SKILL-rationale.md
@@ -170,6 +170,7 @@ text after it is the passage, verbatim.
 - ***Write a live-verification task when the change touches a running service or persistent state.** … the justification is required, the task is not.* — This is deliberately not an integration-test stage on every change: a step that usually resolves to "nothing to do" trains everyone to skip it.
 - *…never a live-verification task …* — (KAN-876) KAN-754's live task ran after the spec and capture tasks in one serial group and repeated, against a stack of its own, what `flow.verify` and visual verify already do against the running app.
 - *Write end-to-end spec and fidelity-capture tasks to run side by side* — (KAN-876) KAN-754 chained its fidelity task after its spec task with no file dependency, and the chain merge then ran them in series.
+- *Write every click and capture in an end-to-end or fidelity spec on a structural locator* — (KAN-545) one change's fidelity specs carried about fifteen hardcoded click coordinates, and a layout shift broke them twice in one change. A re-anchor probe after every layout-shifting task was rejected: it only detects the drift a locator removes.
 
 ### verify-and-handoff.md — Inline verify
 
diff --git a/skills/flow/brainstorm-planner.md b/skills/flow/brainstorm-planner.md
index 87ea9def..bf076960 100644
--- a/skills/flow/brainstorm-planner.md
+++ b/skills/flow/brainstorm-planner.md
@@ -278,6 +278,11 @@ paragraph under **The build-green tag** (`skills/flow-contracts/build-green.md`)
 > names the stack's start row and any build the tasks run against; no task's steps start, stop
 > or rebuild the stack.
 
+> **Write every click and capture in an end-to-end or fidelity spec on a structural locator.**
+> Each such task's steps reach every control the spec clicks, types into or captures by its role
+> and accessible name, its label, its text or its test id — never by a page coordinate, which
+> names a position in one layout and lands on another control once any task shifts that layout.
+
 > **Write a spec so it never waits out a production timing.** A scenario that depends on a
 > production timeout, retry backoff or polling interval plans a test-only override that
 > shortens it — the production value unchanged — in the task that adds the spec, and a suite
diff --git a/skills/flow/visual-verify-verifier.md b/skills/flow/visual-verify-verifier.md
index 873213d8..a269e7e7 100644
--- a/skills/flow/visual-verify-verifier.md
+++ b/skills/flow/visual-verify-verifier.md
@@ -36,7 +36,7 @@ never an exit code the command did not return.
    extended, when one covers the view; a new spec is authored only for a view none covers. Run
    `capture` with `<spec>` substituted for the spec's path. **Every capture asserts the view is on
    screen before it is taken** — gymie-playwright's `expectRoom` — so a missed navigation fails the
-   capture instead of becoming its baseline. **A capture reached through a click also asserts the state it clicked into, and parks the pointer before it is taken** — a positive control on a drawn, state-dependent property the wrong state cannot satisfy (the selected tab's fill, the opened sheet's tag), and the pointer moved clear of every control (`page.mouse.move(0, 0)`) after the last click — so a missed click fails the capture instead of becoming its baseline, and no hover ink reaches one. **A capture whose subject sits under a scrim, or any dimming or alpha layer, also pins that subject structurally** — the presence of each tag the change draws beneath the layer and the absence of each it removes, covered tags included — because the dimming can shrink a real difference below the screenshot comparison's per-pixel threshold; a dimmed capture whose baseline predates the change is recaptured, never trusted because its diff passes. `screenshots`'s root-not-leaf shape is canonical in
+   capture instead of becoming its baseline. **A capture reached through a click also asserts the state it clicked into, and parks the pointer before it is taken** — a positive control on a drawn, state-dependent property the wrong state cannot satisfy (the selected tab's fill, the opened sheet's tag), and the pointer moved clear of every control (`page.mouse.move(0, 0)`) after the last click — so a missed click fails the capture instead of becoming its baseline, and no hover ink reaches one. **A capture reaches every control it clicks or types into by a structural locator, never a page coordinate** (**D. Basic Workflow #3 — Writing plans**, `skills/flow/brainstorm-planner.md`). **A capture whose subject sits under a scrim, or any dimming or alpha layer, also pins that subject structurally** — the presence of each tag the change draws beneath the layer and the absence of each it removes, covered tags included — because the dimming can shrink a real difference below the screenshot comparison's per-pixel threshold; a dimmed capture whose baseline predates the change is recaptured, never trusted because its diff passes. `screenshots`'s root-not-leaf shape is canonical in
    `skills/flow-contracts/project-configuration-visual.md`; nothing here restates it. **Every screenshot
    this spec takes is the full page or viewport, never a clipped region.** A clip is the right tool
    for an implementer's own targeted assertion (a fixed piece of text, an icon), but this stage's
````

  - [ ] **Step 3: Verify.** `grep -c -i "structural locator" skills/flow/brainstorm-planner.md skills/flow/visual-verify-verifier.md skills/flow/SKILL-rationale.md`
    prints `1` for each file; `scripts/check-references.sh`, `python3 scripts/check-markdown-integrity.py`
    and `scripts/check-verbatim-moves.sh` exit 0.
    <!-- measured: grep -c after git apply in a scratch clone @ 7aea62c2 → 1, 1 and 1 -->
  - [ ] **Step 4: Commit** the three files with the subject below.

**Build:** green
**Files:** `skills/flow/brainstorm-planner.md`, `skills/flow/visual-verify-verifier.md`, `skills/flow/SKILL-rationale.md`
**Tests:** **none** — a prose rule; `check-references.sh` and `check-verbatim-moves.sh` cover its citation and its acknowledged sentences.
**Regression:** none — no test is declared.
**Commit:** docs(flow): reach every control in end-to-end and fidelity specs by a structural locator
**After:** none

**Decision:** structural-locators

- [ ] 3. Reconcile design.md's declared frames against the sidecars and the verifier report (KAN-549)
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `ls scripts/check-frame-list.sh`
    fails — nothing compares the frames `design.md` declares with what the verifier counted. Save
    Step 2's diff, apply its test hunk alone
    (`git apply --include='stats/internal/guard/check_frame_list_test.go' <the saved diff>`) and run
    `cd stats && make -s web-dist && go test ./internal/guard -run 'TestCheckFrameList' -count=1`:
    it fails with `rc=-1 out="check-frame-list is not registered"`.
    <!-- measured: both commands in a scratch clone @ 7aea62c2 with Tasks 1 and 2 applied → ls exit 1; --- FAIL: TestCheckFrameList, "check-frame-list is not registered" -->
  - [ ] **Step 2: Apply the fix.** Apply the rest
    (`git apply --exclude='stats/internal/guard/check_frame_list_test.go' <the saved diff>`); it
    creates `scripts/check-frame-list.sh` executable and `skills/flow/scripts/check-frame-list.sh`
    as a symlink to `../../../scripts/check-frame-list.sh`:

````diff verified:applied with git apply to a scratch clone at 7aea62c2 after Tasks 1 and 2; TestCheckFrameList passes all six table cases and the usage cases, and fails "verifier counted fewer than declared" with the frames-count check disabled; gofmt -l and go vet ./... clean; go test ./internal/guard/ and scripts/run-guard-tests.sh pass; check-references.sh, check-guard-symlinks.sh and check-verbatim-moves.sh 7aea62c2 exit 0
diff --git a/.flow/project.md b/.flow/project.md
index b42c07cb..47897120 100644
--- a/.flow/project.md
+++ b/.flow/project.md
@@ -14,7 +14,7 @@ halves live in the one repo and are covered below.
 | flow sources | `/Users/tweety53/Projects/agents` | Bash + Python + Go + Markdown | — | The skills/commands/rules half. Verification is the guard scripts below plus a sandboxed `setup.sh` run. |
 | flow stats daemon | `/Users/tweety53/Projects/agents/stats` | Go + React/Vite | `http://127.0.0.1:4173` | `flowd`, loopback-only. Backed by a dedicated `flow-postgres` container on host port 5433, independent of any other Postgres stack on this machine. Also the one application `## apps` names that a fix run's reload rule (`skills/flow/verify-and-handoff.md`) never reloads — the same protection, not a separate one. |
 
-**This repository is Bash + Python + Go, not Bash-only.** Several `scripts/*.py` guards (`check-task-build-green.py`, `check-plan-shape.py`, `check-task-commit-fields.py`, `check-markdown-integrity.py`, among others) run on Python 3, standard library only — `/usr/bin/python3`, no third-party imports, no pip, no network. The guards ported into `flow-guard` — among them `check-plan-provenance`, `check-installed-citations`, `check-references`, `check-unfinished-work`, `check-panel-reproducers`, `check-panel-findings-closed`, `check-stage-mark-calls`, `check-guard-symlinks`, `check-dispatch-paragraphs`, `mutate-and-verify`, `check-base-moved`, `check-panel-fix-single-dispatch`, `prove-reproducer`, `check-finish-preflight`, `check-workspace-isolation`, `check-task-reviewer-single-dispatch`, `recover-guard-incident`, `plan-class`, `check-installed-rules`, `resolve-base-branch`, `check-visual-verify-dispatched`, `check-task-commit-planning-paths`, `check-panel-citation-trigger`, `check-visual-trigger`, `check-visual-verification`, `check-verify-report`, `resolve-visual-screenshots`, `compose-mockup-frames`, `measure-visual-properties`, `commit-split`, `reshape-branch`, `check-foreign-staged`, `check-planning-commit-location`, `check-archive-scope`, `check-main-checkout-drift`, `refresh-main-checkout`, `land-self-review-report`, `check-self-review-report`, `break-and-prove`, `prepare-workspace` and `check-panel-docs-only` — are Go (`stats/internal/guard/`), and their `scripts/<name>.sh` is a `flow_guard_exec` shim over the binary built from this checkout.
+**This repository is Bash + Python + Go, not Bash-only.** Several `scripts/*.py` guards (`check-task-build-green.py`, `check-plan-shape.py`, `check-task-commit-fields.py`, `check-markdown-integrity.py`, among others) run on Python 3, standard library only — `/usr/bin/python3`, no third-party imports, no pip, no network. The guards ported into `flow-guard` — among them `check-plan-provenance`, `check-installed-citations`, `check-references`, `check-unfinished-work`, `check-panel-reproducers`, `check-panel-findings-closed`, `check-stage-mark-calls`, `check-guard-symlinks`, `check-dispatch-paragraphs`, `mutate-and-verify`, `check-base-moved`, `check-panel-fix-single-dispatch`, `prove-reproducer`, `check-finish-preflight`, `check-workspace-isolation`, `check-task-reviewer-single-dispatch`, `recover-guard-incident`, `plan-class`, `check-installed-rules`, `resolve-base-branch`, `check-visual-verify-dispatched`, `check-task-commit-planning-paths`, `check-panel-citation-trigger`, `check-visual-trigger`, `check-visual-verification`, `check-verify-report`, `check-frame-list`, `resolve-visual-screenshots`, `compose-mockup-frames`, `measure-visual-properties`, `commit-split`, `reshape-branch`, `check-foreign-staged`, `check-planning-commit-location`, `check-archive-scope`, `check-main-checkout-drift`, `refresh-main-checkout`, `land-self-review-report`, `check-self-review-report`, `break-and-prove`, `prepare-workspace` and `check-panel-docs-only` — are Go (`stats/internal/guard/`), and their `scripts/<name>.sh` is a `flow_guard_exec` shim over the binary built from this checkout.
 
 **Every new guard, and any new logic added to an existing guard, is written in Go in `flow-guard`** — `stats/internal/guard/`, registered in `guard.Registry`, with Go table tests — and its `scripts/<name>.sh` is only a thin `flow_guard_exec` shim onto it. A guard still in bash is ported to Go, byte-for-byte parity, before it is extended; it is never grown in bash. A new `check-*.sh` lands with its `test-check-*.sh` companion in the same change — `run-guard-tests.sh` refuses the suite before anything runs while the companion is missing, a Go test in `stats/internal/guard/<name>_test.go` standing in for it (kan-903).
 
diff --git a/scripts/check-frame-list.sh b/scripts/check-frame-list.sh
new file mode 100755
index 00000000..e4f2d273
--- /dev/null
+++ b/scripts/check-frame-list.sh
@@ -0,0 +1,52 @@
+#!/usr/bin/env bash
+# check-frame-list.sh — reconcile the mockup frames design.md declares against
+# the capture specs' `.mockups` sidecars and a flow.visual-verify verifier's
+# report, so a declared frame no capture reached is named rather than lost
+# (KAN-549: design.md declared 22 frames, two rounds reported 15/17 composed,
+# and nothing compared the two).
+#
+# Usage: check-frame-list.sh <design.md> <mockups dir> <report file>
+#
+# <design.md> is the change's <changeRoot>/design.md; <mockups dir> the
+# resolved `mockups` root (skills/flow-contracts/project-configuration-visual.md);
+# <report file> the verifier's
+# <abs-worktree>/.superpowers/sdd/verify-report-<key>.md
+# (skills/flow/visual-verify.md).
+#
+# THE DECLARED LIST is derived, never written down twice: every
+# `<frame id>.png` directly inside <mockups dir> whose <frame id> design.md
+# names as a whole token — bounded on each side by the start or end of a line
+# or by a character that is not a letter, digit, `_` or `-`.
+#
+# THE SIDECARS are `<spec>.mockups` beside every spec the report's
+# `- spec: <path>` lines name; a missing sidecar contributes no frame. A
+# sidecar line is `<screenshot name> <frame id>`, `#` comments and blank lines
+# ignored.
+#
+# A GAP is, for a declared frame:
+#   no sidecar line: <frame id>   no read sidecar maps a screenshot to it
+#   no report line: <frame id>    no `- <frame id>:` or `- <frame id> <…>:`
+#                                 line, and the `- frames:` line does not
+#                                 name it
+# and, for the report:
+#   no frames line                the report carries no `- frames: <n>/<m>`
+#                                 line while design.md declares a frame
+#   frames <n>/<m> below <k> declared   <m> is smaller than the declared count
+#
+# Prints ONE verdict line to stdout:
+#   FRAME-LIST-COMPLETE: <report file> — <k> declared
+#   FRAME-LIST-GAP: <report file> — <each gap, `; `-separated>
+#
+# Exit 0 no gap, 1 at least one gap, 2 cannot answer (a usage error, an
+# unreadable design.md or report, a mockups dir that is not a directory).
+#
+# Go, in flow-guard (stats/internal/guard/framelist.go).
+# flow-guard is built from this checkout, never taken from PATH:
+# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
+# cannot-answer code) with the cause when it cannot.
+set -euo pipefail
+[ -r "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" ] && . "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
+  echo "check-frame-list: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
+  exit 2
+}
+flow_guard_exec check-frame-list 2 "check-frame-list:" "$@"
diff --git a/skills/flow/SKILL-rationale.md b/skills/flow/SKILL-rationale.md
index 3a7f04f4..00d6aaa6 100644
--- a/skills/flow/SKILL-rationale.md
+++ b/skills/flow/SKILL-rationale.md
@@ -236,6 +236,7 @@ text after it is the passage, verbatim.
 - *…a list row carrying a leading radio, or a quiet row whose only indicator is a trailing filled circle on the selected one; a segmented control of N cells each fitting its own label — a differing kind is a departure before any number is taken* — (gymie KAN-437 final verification, Q1: the WEIGHT field rendered in the weight tab's headline-figure style at about three times the frame's field height; every ACTIVITY row carried a leading outlined radio where the frame draws no indicator on unselected rows and a trailing filled circle on the selected one; the GOAL segments' two-line labels ran into the neighbouring cell — none subtle at 1x, none seen, the frame never on the list below)
 - *…fore starting the sweep, and account for each one by name at handoff — done, or why not.** "Did a screen-by-screen sweep" is unfalsifiable without a fixed list: a partial pass reads exactly like a complete one if nothing forces naming what was skipped* — (gymie KAN-30 fix round 7: two prior rounds each reported partial coverage as if it were the whole set, and the gap was caught only when the operator asked "so is X fully compared and fixed?" of a specific frame)
 - *…gainst that list and names every declared frame absent from the per-frame lines, with its reason; the parent reconciles it against `design.md`'s own frame list before applying **Blocking**, and a declared frame with no line blocks as a departure would* — (gymie KAN-437 final verification: `design.md` declared 22 frames, two `flow.visual-verify` rounds each ran three of five fidelity specs and reported 15/17 composed — Q1–Q3 and M1–M2 never composed, never named as skipped, unnoticed by the parent, and every Q1 defect above surfaced only when the operator opened the screen)
+- *The frame-list half of that reconciliation is a guard* — (KAN-549) the parent's by-hand reconciliation above is what missed 22 declared against 15/17 composed; `check-frame-list.sh` derives the declared list from the mockups directory's frames `design.md` names, so no second copy of the list exists to drift.
 - *…s below and capture each one; a state nobody drove into is a state nobody verified.** Each is a state a resting capture and a full-page composite cannot show, so each is reached deliberately rather than by whatever the walkthrough happens to pass through* — (gymie KAN-30 fix rounds 7 and 8: a blind click landed on a day's "Repeat that workout" suggestion instead of "Create a group session" because an extra card existed on that day only, and the sweep almost recorded the wrong screen as verified; one dialog's "selected" fill differed between two sub-states, two sibling cards diverged in shadow, and "Add user" carried a border on one code path and none on another because the empty and populated roster states routed to two different shared button components; a time wheel scrolled correctly one way only, a state-derivation bug no resting capture shows; an empty picker shipped in its pre-restyle appearance beneath a restyled populated sibling; a picker offered an option submit always rejected, surfaced only as a post-hoc toast)
 - *…e shows and state, per run, that it is a display value at the precision the frame shows — a raw float, a `NaN`, `null`, `undefined`, an empty string where the frame draws a value, a placeholder is a defect whatever the frame's own numbers are* — (gymie KAN-437: `82.333333333333 kg` on an axis the frame drew as `82.3 kg`, read as a data difference)
 - *…ading, chart, chips, list, link — and compare their order: a control present in both but in a different position relative to its siblings is a departure the per-control crop never sees, since each crop matches its own control wherever it sits* — (gymie KAN-437: preset chips rendered above the chart the frame drew them below)
diff --git a/skills/flow/scripts/check-frame-list.sh b/skills/flow/scripts/check-frame-list.sh
new file mode 120000
index 00000000..68af2702
--- /dev/null
+++ b/skills/flow/scripts/check-frame-list.sh
@@ -0,0 +1 @@
+../../../scripts/check-frame-list.sh
\ No newline at end of file
diff --git a/skills/flow/visual-verify.md b/skills/flow/visual-verify.md
index 44c8ad45..b06d702e 100644
--- a/skills/flow/visual-verify.md
+++ b/skills/flow/visual-verify.md
@@ -263,6 +263,18 @@ not apply>`. Its remainder is dispatched exactly as a `check-verify-report.sh` e
 **Every report is checked for completeness** above, the reconciliation's list standing in for the
 guard's verdict line.
 
+**The frame-list half of that reconciliation is a guard**, run whenever `mockups` is declared,
+before the rest is reconciled by hand:
+
+```bash
+check-frame-list.sh <changeRoot>/design.md <mockups root> <abs-worktree>/.superpowers/sdd/verify-report-<key>.md
+```
+
+Its header is canonical for the declared list it derives and the gaps it names. Exit 1: a `no
+sidecar line` gap is a frame with no sidecar line, which blocks under **Blocking** below, and
+every other gap joins the reconciliation's list above. Exit 2 is a usage error or an unreadable
+input, corrected and the guard run again.
+
 **Blocking.** This stage blocks the `IN_PROGRESS` handoff on: **a workspace pre-flight that failed
 (step 3), which ends the stage before any dispatch**, a failed `setup`, a failed `verify`,
 a genuine `capture` failure — the full app suite's included — **a frame this change added or
diff --git a/stats/internal/guard/check_frame_list_test.go b/stats/internal/guard/check_frame_list_test.go
new file mode 100644
index 00000000..422d36b9
--- /dev/null
+++ b/stats/internal/guard/check_frame_list_test.go
@@ -0,0 +1,90 @@
+package guard
+
+import (
+	"os"
+	"path/filepath"
+	"strings"
+	"testing"
+)
+
+// Every case of the check-frame-list contract (scripts/check-frame-list.sh):
+// the declared list is the mockups dir's frames design.md names as whole
+// tokens, and each declared frame needs a sidecar line and a report line.
+func TestCheckFrameList(t *testing.T) {
+	const design = "Frames `Q1` and Q2 (see mockups/M1.png); Q100 and UNCITED-old are other ids.\n"
+	cases := []struct {
+		name    string
+		sidecar string // "" writes no sidecar
+		report  string // %SPEC% is replaced by the spec path
+		rc      int
+		want    string
+	}{
+		{"every declared frame captured and reported",
+			"q1 Q1\nq2 Q2\n# comment\nm1 M1\n",
+			"- spec: %SPEC%\n- Q1: /c/q1.png diff=0.1 — match\n- Q2 bands: 3/3 paired\n- M1: /c/m1.png diff=0.2 — match\n- frames: 3/3 — none absent\n",
+			0, "FRAME-LIST-COMPLETE"},
+		{"a declared frame named absent with its reason",
+			"q1 Q1\nq2 Q2\nm1 M1\n",
+			"- spec: %SPEC%\n- Q1: /c/q1.png diff=0.1 — match\n- M1: /c/m1.png diff=0.2 — match\n- frames: 2/3 — Q2: reached only by inference from Q1\n",
+			0, "3 declared"},
+		// KAN-549: design.md declared more frames than the verifier counted.
+		{"verifier counted fewer than declared",
+			"q1 Q1\nq2 Q2\nm1 M1\n",
+			"- spec: %SPEC%\n- Q1: /c/q1.png diff=0.1 — match\n- frames: 1/1\n",
+			1, "frames 1/1 below 3 declared; no report line: M1; no report line: Q2"},
+		{"no sidecar line for a declared frame",
+			"q1 Q1\nq2 Q2\n",
+			"- spec: %SPEC%\n- Q1: x\n- Q2: x\n- M1: x\n- frames: 3/3\n",
+			1, "no sidecar line: M1"},
+		{"no sidecar file at all",
+			"",
+			"- spec: %SPEC%\n- Q1: x\n- Q2: x\n- M1: x\n- frames: 3/3\n",
+			1, "no sidecar line: M1; no sidecar line: Q1; no sidecar line: Q2"},
+		{"no frames line",
+			"q1 Q1\nq2 Q2\nm1 M1\n",
+			"- spec: %SPEC%\n- Q1: x\n- Q2: x\n- M1: x\n",
+			1, "no frames line"},
+	}
+	for _, tc := range cases {
+		t.Run(tc.name, func(t *testing.T) {
+			dir := t.TempDir()
+			mockups := filepath.Join(dir, "mockups")
+			if err := os.Mkdir(mockups, 0o755); err != nil {
+				t.Fatal(err)
+			}
+			for _, f := range []string{"Q1.png", "Q2.png", "Q10.png", "M1.png", "UNCITED.png", "notes.txt"} {
+				if err := os.WriteFile(filepath.Join(mockups, f), nil, 0o644); err != nil {
+					t.Fatal(err)
+				}
+			}
+			spec := filepath.Join(dir, "kan-1.spec.ts")
+			if tc.sidecar != "" {
+				if err := os.WriteFile(spec+".mockups", []byte(tc.sidecar), 0o644); err != nil {
+					t.Fatal(err)
+				}
+			}
+			designPath := filepath.Join(dir, "design.md")
+			report := filepath.Join(dir, "verify-report-visual-verify.md")
+			if err := os.WriteFile(designPath, []byte(design), 0o644); err != nil {
+				t.Fatal(err)
+			}
+			if err := os.WriteFile(report, []byte(strings.ReplaceAll(tc.report, "%SPEC%", spec)), 0o644); err != nil {
+				t.Fatal(err)
+			}
+			r := runGuard("check-frame-list", []string{designPath, mockups, report}, Env{Getenv: os.Getenv})
+			if r.rc != tc.rc || !strings.Contains(r.out, tc.want) {
+				t.Fatalf("rc=%d out=%q; want rc=%d containing %q", r.rc, r.out, tc.rc, tc.want)
+			}
+		})
+	}
+	dir := t.TempDir()
+	empty := filepath.Join(dir, "design.md")
+	if err := os.WriteFile(empty, nil, 0o644); err != nil {
+		t.Fatal(err)
+	}
+	for _, args := range [][]string{nil, {empty, dir}, {filepath.Join(dir, "absent.md"), dir, empty}, {empty, empty, empty}, {empty, dir, filepath.Join(dir, "absent.md")}} {
+		if r := runGuard("check-frame-list", args, Env{Getenv: os.Getenv}); r.rc != 2 {
+			t.Fatalf("args %q: rc=%d out=%q; want exit 2", args, r.rc, r.out)
+		}
+	}
+}
diff --git a/stats/internal/guard/framelist.go b/stats/internal/guard/framelist.go
new file mode 100644
index 00000000..f9a26a22
--- /dev/null
+++ b/stats/internal/guard/framelist.go
@@ -0,0 +1,112 @@
+package guard
+
+import (
+	"fmt"
+	"io"
+	"os"
+	"regexp"
+	"sort"
+	"strconv"
+	"strings"
+)
+
+// checkFrameList is scripts/check-frame-list.sh: that script's header comment
+// is the contract. It derives the frames design.md declares from the mockups
+// directory and reconciles them against the capture specs' sidecars and the
+// verifier report's per-frame and `- frames:` lines.
+func init() { Registry["check-frame-list"] = checkFrameList }
+
+var (
+	flSpecLine   = regexp.MustCompile(`^\s*- spec:\s*(\S+)\s*$`)
+	flFramesLine = regexp.MustCompile(`^\s*- frames:\s*(\d+)/(\d+)(.*)$`)
+)
+
+// flToken matches id as a whole token: not preceded or followed by a letter,
+// digit, `_` or `-`.
+func flToken(id string) *regexp.Regexp {
+	return regexp.MustCompile(`(^|[^A-Za-z0-9_-])` + regexp.QuoteMeta(id) + `($|[^A-Za-z0-9_-])`)
+}
+
+func checkFrameList(args []string, env Env, stdout, stderr io.Writer) int {
+	const prog = "check-frame-list"
+	if len(args) != 3 {
+		fmt.Fprintf(stderr, "%s: usage: check-frame-list.sh <design.md> <mockups dir> <report file>\n", prog)
+		return 2
+	}
+	design, err := os.ReadFile(args[0])
+	if err != nil {
+		fmt.Fprintf(stderr, "%s: cannot read %s: %v\n", prog, args[0], err)
+		return 2
+	}
+	entries, err := os.ReadDir(args[1])
+	if err != nil {
+		fmt.Fprintf(stderr, "%s: cannot read mockups dir %s: %v\n", prog, args[1], err)
+		return 2
+	}
+	report, err := os.ReadFile(args[2])
+	if err != nil {
+		fmt.Fprintf(stderr, "%s: cannot read %s: %v\n", prog, args[2], err)
+		return 2
+	}
+
+	var declared []string
+	for _, e := range entries {
+		id, ok := strings.CutSuffix(e.Name(), ".png")
+		if ok && !e.IsDir() && id != "" && flToken(id).Match(design) {
+			declared = append(declared, id)
+		}
+	}
+	sort.Strings(declared)
+
+	lines := strings.Split(string(report), "\n")
+	sidecar := map[string]bool{}
+	framesLine := []string(nil)
+	for _, line := range lines {
+		if m := flSpecLine.FindStringSubmatch(line); m != nil {
+			data, err := os.ReadFile(m[1] + ".mockups")
+			if err != nil {
+				continue
+			}
+			for _, l := range strings.Split(string(data), "\n") {
+				f := strings.Fields(l)
+				if len(f) >= 2 && !strings.HasPrefix(f[0], "#") {
+					sidecar[f[1]] = true
+				}
+			}
+		}
+		if m := flFramesLine.FindStringSubmatch(line); m != nil {
+			framesLine = m
+		}
+	}
+
+	var gaps []string
+	if len(declared) > 0 {
+		if framesLine == nil {
+			gaps = append(gaps, "no frames line")
+		} else if m, _ := strconv.Atoi(framesLine[2]); m < len(declared) {
+			gaps = append(gaps, fmt.Sprintf("frames %s/%s below %d declared", framesLine[1], framesLine[2], len(declared)))
+		}
+	}
+	for _, id := range declared {
+		if !sidecar[id] {
+			gaps = append(gaps, "no sidecar line: "+id)
+		}
+		perFrame := regexp.MustCompile(`^\s*- ` + regexp.QuoteMeta(id) + `( [^:]*)?:`)
+		named := framesLine != nil && flToken(id).MatchString(framesLine[3])
+		for _, line := range lines {
+			if perFrame.MatchString(line) {
+				named = true
+				break
+			}
+		}
+		if !named {
+			gaps = append(gaps, "no report line: "+id)
+		}
+	}
+	if len(gaps) > 0 {
+		fmt.Fprintf(stdout, "FRAME-LIST-GAP: %s — %s\n", args[2], strings.Join(gaps, "; "))
+		return 1
+	}
+	fmt.Fprintf(stdout, "FRAME-LIST-COMPLETE: %s — %d declared\n", args[2], len(declared))
+	return 0
+}
````

  - [ ] **Step 3: Verify.** `cd stats && make -s web-dist && go test ./internal/guard -run 'TestCheckFrameList' -count=1`
    passes; `cd stats && go vet ./internal/guard && gofmt -l internal/guard` prints nothing;
    `scripts/check-guard-symlinks.sh`, `scripts/check-references.sh`, `python3 scripts/check-markdown-integrity.py`
    and `scripts/check-verbatim-moves.sh` exit 0.
  - [ ] **Step 4: Commit** the seven files with the subject below.

**Build:** green
**Files:** `stats/internal/guard/framelist.go`, `stats/internal/guard/check_frame_list_test.go`, `scripts/check-frame-list.sh`, `skills/flow/scripts/check-frame-list.sh`, `skills/flow/visual-verify.md`, `skills/flow/SKILL-rationale.md`, `.flow/project.md`
**Tests:** `TestCheckFrameList`
**Regression:** `TestCheckFrameList` — reverting the commit leaves `check-frame-list` unregistered: every case fails with `check-frame-list is not registered`, and `flow.visual-verify` again accepts a report counting fewer frames than `design.md` declares.
**Commit:** feat(guard): reconcile design.md's declared frames against the sidecars and the verifier report
**After:** Task 2

**Decision:** frame-list-guard
**Decision:** frame-list-derived
