# kan-395-flow-fingerprint-the-served-bundle-before-a

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.
> **Relocation:** no

Four commits in dependency order: the guard admits the word (task 1) before the contract
documents it (task 2), the stage uses it (task 3) and this repository declares it (task 4) —
task 4's row fails `scripts/check-visual-verification.sh .` until task 1 lands. The approved
design is
`docs/superpowers/specs/2026-09-04-kan-395-flow-fingerprint-the-served-bundle-before-a-design.md`;
`design.md` beside this file carries the decisions.

**Baseline, measured before any edit:**

- `scripts/test-check-visual-verification.sh` prints 65 `ok:` lines and `all cases passed`.
  <!-- measured: scripts/test-check-visual-verification.sh | grep -c '^ok' @ c93da10 -->
- `scripts/check-contract-budget.sh` reports 71 owned files within budget. The three prose files
  this plan edits measure 26559, 45467 and 23952 bytes against rows of 33199, 48175 and 26450
  (`skills/flow/verify-and-handoff.md`, `skills/flow-contracts/project-configuration.md`,
  `.flow/project.md`); with every diff below applied they measure 27554, 45976 and 24597 — all
  within budget, so task 4's `check-contract-budget.sh` row edit is not needed.
  <!-- measured: scripts/check-contract-budget.sh; wc -c on the three files before and after applying the four diffs; grep in scripts/check-contract-budget.sh @ c93da10 -->
- `scripts/check-normative-inventory.sh` prints the same inventory before and after the four
  diffs — no `SHALL`/`MUST` sentence is added or lost.
  <!-- measured: scripts/check-normative-inventory.sh > before; apply; scripts/check-normative-inventory.sh > after; diff before after @ c93da10 -->
- Case 28 below fails against the unchanged guard — 4 assertions, each with the guard's own
  `is not one of \`setup\`, \`verify\` or \`capture\`` message in the output — and passes once
  task 1's guard change is applied.
  <!-- measured: scripts/test-check-visual-verification.sh with the harness diff applied and the guard unchanged, then with both applied @ c93da10 -->

Every diff below was authored for this plan, applied together to the working tree, checked with
`scripts/check-vocabulary.sh`, `check-references.sh`, `check-contract-budget.sh`,
`check-markdown-integrity.py`, `check-stage-mark-calls.sh`, `check-dispatch-paragraphs.sh`,
`check-installed-citations.sh`, `check-workspace-isolation.sh`, `check-guard-symlinks.sh`,
`check-visual-verification.sh .` and the harness, then reverted; `git apply --check` accepts each
against `c93da10`.

---

- [x] 1. Admit `fingerprint` into `check-visual-verification.sh`'s closed command vocabulary

One awk condition and one message gain the fourth word; two header comments stop claiming three.
Case 28 asserts the row is accepted; every earlier case declares no `fingerprint` row and passes or
fails on other grounds alone, which is the absence-is-silent half.

**Files:** `scripts/check-visual-verification.sh`, `scripts/test-check-visual-verification.sh`
**Tests:** `scripts/test-check-visual-verification.sh` — Case 28
**Regression:** reverting this commit restores the three-word condition — case 28's `fingerprint`
row is reported as `not one of \`setup\`, \`verify\` or \`capture\`` and its four assertions fail.
**Baseline:** before=65 after=75 `ok:` lines in `scripts/test-check-visual-verification.sh`
<!-- measured: scripts/test-check-visual-verification.sh | grep -c '^ok', before and with this diff applied plus the panel round 0 mutation fixup (case 29) -->
**Commit:** `feat(scripts): admit fingerprint into check-visual-verification.sh's command vocabulary`
**Build:** green

  - [x] **Step 1: Apply the diff** — from the repository root, `git apply` the block below verbatim.

````diff verified:authored for this plan, applied to the working tree at c93da10 with the other three diffs, every `## lint` guard named in this plan's baseline green, then reverted; git apply --check passes at c93da10
--- a/scripts/check-visual-verification.sh
+++ b/scripts/check-visual-verification.sh
@@ -75,7 +75,8 @@
 # THE INPUT IS ATTACKER-INFLUENCED, the same fact check-workspace-isolation.sh
 # is written against: `.flow/project.md` is tracked and editable in any pull
 # request. NOTHING read here is executed — this guard never runs `setup`,
-# `verify` or `capture`, and never interpolates a cell into a shell. All of
+# `verify`, `capture` or `fingerprint`, and never interpolates a cell into a
+# shell. All of
 # the table parsing happens inside one awk program whose only input is the
 # file's text, and every violation line the awk program prints, plus every
 # cell this guard interpolates into a message afterward, passes through
@@ -347,12 +348,15 @@
   }
 
   # One row of the commands table. The `Command` vocabulary is closed —
-  # `setup`, `verify` and `capture` — for the identical reason.
+  # `setup`, `verify`, `capture` and `fingerprint` — for the identical reason.
+  # `fingerprint` is optional and nothing below requires it (KAN-395): a
+  # project with no served bundle declares none, and its absence is silent
+  # here exactly as `setup`'s is.
   function check_command_row(lineno, cells,   key, disp) {
     disp = trimcell(cells[1])
     key = foldcell(cells[1])
-    if (key != "setup" && key != "verify" && key != "capture") {
-      violation(lineno, "Command `" disp "` is not one of `setup`, `verify` or `capture` — the vocabulary is closed, so the row is dropped")
+    if (key != "setup" && key != "verify" && key != "capture" && key != "fingerprint") {
+      violation(lineno, "Command `" disp "` is not one of `setup`, `verify`, `capture` or `fingerprint` — the vocabulary is closed, so the row is dropped")
       return
     }
     if (key in cmd_seen) {
--- a/scripts/test-check-visual-verification.sh
+++ b/scripts/test-check-visual-verification.sh
@@ -742,7 +742,33 @@
 assert_out_contains "case 27" "ui paths"
 assert_out_contains "case 27" "no usable glob"
 assert_out_not_contains "case 27" "VISUAL-OK"
+
+# ===========================================================================
+# Case 28 (KAN-395): a `fingerprint` row is a member of the closed `Command`
+# vocabulary -- accepted, never reported as an unknown command -- and it is
+# OPTIONAL: every case above declares none and passes or fails on other
+# grounds alone, which is the absence-is-silent half of this case, asserted
+# by cases 1 and 24 rather than restated here.
+# ===========================================================================
+new_root
+write_cfg "## visual verification
+
+| Setting | Value |
+|---------|-------|
+| \`ui paths\` | \`stats/web/src/**\` |
+| \`screenshots\` | \`stats/web/tests/visual\` |
 
+| Command | Runs |
+|---------|------|
+| \`verify\` | \`npm run test:visual\` |
+| \`capture\` | \`npx playwright test <spec>\` |
+| \`fingerprint\` | \`curl -sf http://127.0.0.1:4174/ \| cmp -s - internal/web/dist/index.html\` |"
+run_guard
+assert_rc "case 28" 0
+assert_out_contains "case 28" "VISUAL-OK"
+assert_out_not_contains "case 28" "vocabulary is closed"
+assert_out_not_contains "case 28" "fingerprint"
+
 if [ "$FAILURES" -ne 0 ]; then
   printf '%s case(s) failed\n' "$FAILURES" >&2
   exit 1
````

  - [x] **Step 2: Run the harness** — `scripts/test-check-visual-verification.sh` prints 75 `ok:`
    lines and `all cases passed`.
    <!-- predicted: scripts/test-check-visual-verification.sh | grep -c '^ok' after step 1 -->
  - [x] **Step 3: Commit** with the `**Commit:**` subject above.

- [x] 2. Document the `fingerprint` row in `project-configuration.md`

The `## visual verification` contract gains the fourth command row, its summary row names the
command, and "three rows below" becomes four.

**Files:** `skills/flow-contracts/project-configuration.md`
**Tests:** **none**
**Regression:** none — prose; reverting leaves the guard accepting a row the contract does not
name, which `scripts/check-vocabulary.sh` and `check-references.sh` cannot see.
**Baseline:** before=65 after=65 `ok:` lines in `scripts/test-check-visual-verification.sh`
<!-- measured: unchanged by this task; the harness never reads the contract @ c93da10 -->
**Commit:** `docs(flow-contracts): document the optional fingerprint command row`
**Build:** green

  - [x] **Step 1: Apply the diff** — `git apply` the block below verbatim.

````diff verified:authored for this plan, applied to the working tree at c93da10 with the other three diffs, every `## lint` guard named in this plan's baseline green, then reverted; git apply --check passes at c93da10
diff --git a/skills/flow-contracts/project-configuration.md b/skills/flow-contracts/project-configuration.md
index 563d663..1af846d 100644
--- a/skills/flow-contracts/project-configuration.md
+++ b/skills/flow-contracts/project-configuration.md
@@ -33,7 +33,7 @@ bodies are constrained to what that procedure accepts.
 | `## planning model` | Optional. One member of `<agents repo>/stats/internal/store/settings.go`'s `ValidModels` map and nothing else, never free-form prose, matching `## default landing route`'s own single-line-literal shape. When present and valid it wins over the settings store's `planningModel` in `PLANNING_MODEL` resolution (**Model resolution**, `skills/flow/SKILL.md`); absent falls through to the store. A body matching no `ValidModels` member is reported by name (quoting what was found) and dropped, resolving as if the key were absent. |
 | `## self review model` | Optional. One member of `<agents repo>/stats/internal/store/settings.go`'s `ValidModels` map and nothing else, never free-form prose. When present and valid it wins over the settings store's `selfReviewModel` in `SELF_REVIEW_MODEL` resolution (**Model resolution**, `skills/flow/SKILL.md`); absent falls through to the store. A body matching no `ValidModels` member is reported by name (quoting what was found) and dropped, resolving as if the key were absent. |
 | `## workspace isolation` | Optional. The resources an apply worktree runs against its own copy of: for each one, the environment variable that carries it and the value it falls back to, plus the commands that create them, remove them and report which of them survived. Most of those values are derived from the workspace id, and one is not — the cache index is claimed at run time and is not derived from it. Its rows are resolved and validated rather than read, so the two tables specified below are the whole of what this body means to the resolver — prose beside them is for the reader, and is where a project records what it has deliberately **not** isolated. Absent means the project is not isolated, which is a supported state and not a misconfiguration — again, see below. Which resources there are, and how each derived value is derived, is stated once under **What the id derives** (`skills/flow-contracts/workspace-isolation.md`). |
-| `## visual verification` | Optional. What the `flow.visual-verify` stage validates before it runs: which UI paths in a change's diff trigger it, where captured screenshots land, the `setup`/`verify`/`capture` commands, and an optional `regression checkout` naming the repository the spec and its PNGs commit to, with `regression repo` recording the identity — never an authorisation — that checkout's real `origin` must equal. Its rows are resolved and validated rather than read, so the two tables specified below are the whole of what this body means to the resolver — prose beside them is for the reader. Absent means the stage is not configured for this project, a supported state and not a misconfiguration. Mechanically enforced by `<agents repo>/scripts/check-visual-verification.sh`, canonical for what it checks. |
+| `## visual verification` | Optional. What the `flow.visual-verify` stage validates before it runs: which UI paths in a change's diff trigger it, where captured screenshots land, the `setup`/`verify`/`capture`/`fingerprint` commands, and an optional `regression checkout` naming the repository the spec and its PNGs commit to, with `regression repo` recording the identity — never an authorisation — that checkout's real `origin` must equal. Its rows are resolved and validated rather than read, so the two tables specified below are the whole of what this body means to the resolver — prose beside them is for the reader. Absent means the stage is not configured for this project, a supported state and not a misconfiguration. Mechanically enforced by `<agents repo>/scripts/check-visual-verification.sh`, canonical for what it checks. |
 | `## review panel citation check` | Optional. A single fenced command block and nothing else in the section — unlike `## workspace isolation` and `## visual verification` above, this key gets no dedicated parsing guard: the whole of what it means to the resolver is "a fenced block exists, read literally". **Absent means the whole pre-panel step is skipped**, exactly like every other optional key. The command runs from the apply worktree only when `check-panel-citation-trigger.sh` exits 0 (**Guard resolution**, `skills/flow-contracts/pipeline.md`, is how `skills/flow/review-panel.md` names it, which is also canonical for the full wiring procedure); its combined stdout+stderr is captured verbatim and its exit code is never gating. |
 
 **How a `## standards` entry resolves to a file.** Every entry in the `## standards` section is
@@ -476,13 +476,14 @@ rows below is reported and its row dropped, never silently ignored.
 The commands table's header folds the same way, to `command|runs` — the same heading
 `## workspace isolation` above already establishes for a table of project-declared commands, reused
 here rather than invented a second time. **Its `Command` vocabulary is closed** too: a name outside
-the three rows below is reported and its row dropped.
+the four rows below is reported and its row dropped.
 
 | Command | Required | Meaning |
 |---------|----------|---------|
 | `setup` | no | Run once before `verify` when the toolchain is missing. |
 | `verify` | yes | Runs the checked-in baseline suite. |
 | `capture` | yes | Runs the per-change spec and **creates** this change's baseline — it is expected to write PNGs that do not yet exist, and that is success, not failure. `verify` above is the regression gate, guarding an already-committed baseline; `capture` is not, and the project's own command must be the variant that succeeds on a first-run write rather than the one built to fail a comparison against nothing. |
+| `fingerprint` | no | Exits 0 when the app the stage's probe answered from is serving the worktree's own build, non-zero otherwise. What a bundle's identity is differs per stack, so the project owns the comparison; `flow.visual-verify` owns what a non-zero exit does (one restart from `## run`, then block). Absent means the stage reports `fingerprint: not declared` and proves nothing about the served bundle — a supported state for a project with no served bundle, not a misconfiguration. |
 
 **No push is ever automatic.** `flow.visual-verify` commits the per-change spec and its PNGs to the
 `regression checkout` when one is declared and stops there; the handoff prints the push command for
````

  - [x] **Step 2: Lint** — `scripts/check-vocabulary.sh && scripts/check-references.sh &&
    scripts/check-contract-budget.sh && scripts/check-markdown-integrity.py` exit 0.
  - [x] **Step 3: Commit** with the `**Commit:**` subject above.

- [x] 3. Add the fingerprint step, report row and blocking rule to `verify-and-handoff.md`

New step 5 between the probe and `verify`; steps 5–10 become 6–11 and every reference to a
renumbered step — the conductor/verifier split, step 7's recursive search, step 10's commit,
step 11's stop, and the two mentions under **Stage** and **The `Visual:` line** — moves with it.
The report template gains `- fingerprint:`; **Blocking** gains the second mismatch.

**Files:** `skills/flow/verify-and-handoff.md`
**Tests:** **none**
**Regression:** none — prose; reverting restores the probe-and-reuse gap KAN-395 names.
**Baseline:** before=65 after=65 `ok:` lines in `scripts/test-check-visual-verification.sh`
<!-- measured: unchanged by this task; the harness never reads the skill @ c93da10 -->
**Commit:** `feat(flow): fingerprint the served bundle before a screenshot counts as evidence`
**Build:** green

  - [x] **Step 1: Apply the diff** — `git apply` the block below verbatim.

````diff verified:authored for this plan, applied to the working tree at c93da10 with the other three diffs, every `## lint` guard named in this plan's baseline green, then reverted; git apply --check passes at c93da10
diff --git a/skills/flow/verify-and-handoff.md b/skills/flow/verify-and-handoff.md
index 902b15a..8a723fb 100644
--- a/skills/flow/verify-and-handoff.md
+++ b/skills/flow/verify-and-handoff.md
@@ -150,12 +150,12 @@ Reads the `## visual verification` section, canonical in
 this pipeline restates it. Resolve once per worktree in this run's resolved set, the same set
 **Verify** above resolved:
 
-Steps 1, 2 and 9 are the conductor's. Steps 3–8 and 10 are run by one verifier per worktree
+Steps 1, 2 and 10 are the conductor's. Steps 3–9 and 11 are run by one verifier per worktree
 surviving steps 1–2, dispatched per **The verifier dispatch** above with `-key visual-verify`; the
 conductor applies **Blocking** to its report. Its prompt states: the absolute worktree path; the
-`KEY=value` lines **Verify** exported for it; this section's resolved `setup`, `verify`, `capture`
-commands and `screenshots` root; the worktree-resolved URL of each app `ui paths` matched; the
-project's `## run` commands; the views touched; `<changeRoot>`; and to run steps 3–8 and 10 below
+`KEY=value` lines **Verify** exported for it; this section's resolved `setup`, `verify`, `capture`,
+`fingerprint` commands and `screenshots` root; the worktree-resolved URL of each app `ui paths` matched; the
+project's `## run` commands; the views touched; `<changeRoot>`; and to run steps 3–9 and 11 below
 as written, committing and pushing nothing.
 
 1. **Resolve the section** — read that worktree's own `<project>/.flow/project.md` directly, by
@@ -177,9 +177,18 @@ as written, committing and pushing nothing.
 4. **Probe before starting anything.** Probe the URL of each app `ui paths` matched, resolved for
    this worktree per **What the id derives** (`skills/flow-contracts/workspace-isolation.md`) —
    never the project's declared default. If nothing answers, start the stack from `## run` and
-   record that this stage started it — needed at step 10.
-5. **Run `verify`.** A non-zero exit blocks.
-6. **Capture** — author a spec covering the views this change touched, then run `capture` with
+   record that this stage started it — needed at step 11.
+5. **Fingerprint the served bundle, if `fingerprint` is declared.** A screenshot is evidence only
+   of what the app was serving when it was taken, and a stack step 4 found already running may be
+   serving a build older than the worktree — KAN-29's last fix round captured, and nearly accepted,
+   the bug the fix had removed. Run `fingerprint`. Exit 0 → the served bundle is the worktree's
+   build; continue. Non-zero → stop the stack, start it from `## run`, record that this stage
+   started it (step 11 stops it), and run `fingerprint` once more. A second non-zero exit blocks,
+   carrying the command's output. No row declared → report `fingerprint: not declared` and
+   continue; the report makes the gap visible in every handoff, but this stage cannot prove what
+   it was never told how to check.
+6. **Run `verify`.** A non-zero exit blocks.
+7. **Capture** — author a spec covering the views this change touched, then run `capture` with
    `<spec>` substituted for the spec's path. `screenshots`'s root-not-leaf shape is canonical in
    `skills/flow-contracts/project-configuration.md`; nothing here restates it. **`capture` creates
    this change's baseline**: writing a PNG that does not yet exist is its success path, not a
@@ -188,7 +197,7 @@ as written, committing and pushing nothing.
    `check-spec-reach.sh <worktree>` — the spec `capture` just wrote must be reached by a
    `package.json` script of the `regression checkout`; exit 1 (an orphan, named) or 2 (cannot
    answer) blocks.
-7. **Read every captured PNG — resolve their paths with the guard, not by eye.** Run
+8. **Read every captured PNG — resolve their paths with the guard, not by eye.** Run
 
    ```bash
    resolve-visual-screenshots.sh <worktree> <spec's basename>
@@ -200,9 +209,9 @@ as written, committing and pushing nothing.
    that never rendered; exit 2 (cannot answer) blocks the same way. **Read every printed path** — no
    script can do that — and state, per view, what was seen. An unreadable PNG is reported and blocks
    too.
-8. **Write `<changeRoot>/visual-verification.md`** — one entry per view: its absolute screenshot
-   path, resolved by the same recursive search step 7 used, and what was seen.
-9. **Commit the spec and its PNGs, and stop there.** A declared `regression checkout` receives
+9. **Write `<changeRoot>/visual-verification.md`** — one entry per view: its absolute screenshot
+   path, resolved by the same recursive search step 8 used, and what was seen.
+10. **Commit the spec and its PNGs, and stop there.** A declared `regression checkout` receives
    them; with none declared, commit to the change's own branch instead. **Never push** — see
    `no-automatic-push` (design.md): a file inside a repository cannot authorise a push to another
    repository, so no guard here grants one. `regression repo` still records which repository the
@@ -213,13 +222,14 @@ as written, committing and pushing nothing.
    ```bash
    git -C <regression checkout> push
    ```
-10. **Stop the stack only if step 4 started it.** A stack the operator already had running is left
+11. **Stop the stack only if step 4 or step 5 started it.** A stack the operator already had running is left
     alone.
 
 ```text verified:design.md section 3 of this change
 ## Report
 - setup: <not declared | exit <n>>
 - stack: <already running | started and stopped | could not be started — <output>>
+- fingerprint: <not declared | exit 0 | mismatch → restarted → exit <n>>
 - verify: exit <n>
   <output, verbatim or last 40 lines>
 - capture: exit <n>
@@ -235,7 +245,8 @@ is carried in the report; the `Visual:` handoff line is built from its view entr
 
 **Blocking.** This stage blocks the `IN_PROGRESS` handoff on: a failed `setup`, a failed `verify`,
 a genuine `capture` failure — **never a first-run snapshot write, which is `capture`'s own success
-path per step 6 above** — a stack that could not be started, a `check-spec-reach.sh` exit 1 or 2,
+path per step 7 above** — a stack that could not be started, **a `fingerprint` that still exits
+non-zero after step 5's restart**, a `check-spec-reach.sh` exit 1 or 2,
 an unreadable PNG, and **a defect the
 verifier reports in a captured screenshot — even when every assertion passed.** That last one is the whole
 point of this stage: three defects have shipped invisible to a diff, a five-pass review panel and
@@ -325,7 +336,7 @@ Resolve the run instructions for the handoff's `Run it:` section. It writes no f
   `flow` database inside it are never stopped, restarted or dropped by any run —
   `<project>/CLAUDE.md` states that prohibition and this rule does not weaken it. Where a project's
   own `## run` names that service, the prohibition wins over this reload rule, never the reverse.
-  This is separate from **Visual verification**'s own start/stop rule above (step 10): that stage
+  This is separate from **Visual verification**'s own start/stop rule above (step 11): that stage
   stops only the stack it started for its own probe, and that rule is not restated here. This rule
   reloads whatever the run instructions name, on every fix run, regardless of whether that stage ran
   or started anything.
@@ -341,7 +352,7 @@ Resolve the run instructions for the handoff's `Run it:` section. It writes no f
   <project>/CLAUDE.md's "Never stop the dev workspace's stats service or its storage".
   ```
 
-  `flow.visual-verify`'s own `make ui-test-up`/`make ui-test-down` pair (step 4 and step 10 above)
+  `flow.visual-verify`'s own `make ui-test-up`/`make ui-test-down` pair (steps 4–5 and step 11 above)
   is a different mechanism entirely — it starts and stops the disposable UI-test stack on
   `127.0.0.1:4174` for that stage's own probe, and `4174` is not an application `## apps` names at
   all, so it is never this rule's reload target.
@@ -441,7 +452,7 @@ line is printed the same way — always, `unknown` included.**
 
 **The `Visual:` line reports `flow.visual-verify`'s own outcome.** Every screenshot path in it is
 absolute, per **Handoff output** (`skills/flow-contracts/pipeline.md`)'s every-path-is-absolute
-rule — the operator must be able to open the PNG. **Its push clause appears only when step 9
+rule — the operator must be able to open the PNG. **Its push clause appears only when step 10
 committed to a `regression checkout`** — the stage never pushes itself, per `no-automatic-push`, so
 this is the command the operator runs by hand to land that commit.
 
````

  - [x] **Step 2: Check the renumbering** — `grep -n 'step [0-9]' skills/flow/verify-and-handoff.md`
    names no step by a number the list no longer carries: the conductor's are 1, 2 and 10, the
    verifier's 3–9 and 11, and no reference to step 5 means anything but the fingerprint.
  - [x] **Step 3: Lint** — `scripts/check-vocabulary.sh && scripts/check-references.sh &&
    scripts/check-contract-budget.sh && scripts/check-stage-mark-calls.sh &&
    scripts/check-dispatch-paragraphs.sh && scripts/check-markdown-integrity.py` exit 0.
  - [x] **Step 4: Commit** with the `**Commit:**` subject above.

- [x] 4. Declare this repository's `fingerprint` row in `.flow/project.md`

The commands table gains the row and a paragraph says why `index.html` is the fingerprint and why
the rebuild comes first. The `\|` inside the cell is Markdown's escaped pipe, which the guard's
`split_cells` already honours — case 28 writes the same shape.

**Files:** `.flow/project.md`
**Tests:** **none**
**Regression:** none — configuration; reverting makes the stage report `fingerprint: not declared`
for this repository again.
**Baseline:** before=65 after=65 `ok:` lines in `scripts/test-check-visual-verification.sh`
<!-- measured: unchanged by this task; the harness builds its own fixtures @ c93da10 -->
**Commit:** `feat(project): declare the SPA bundle fingerprint for visual verification`
**Build:** green

  - [x] **Step 1: Apply the diff** — `git apply` the block below verbatim.

````diff verified:authored for this plan, applied to the working tree at c93da10 with the other three diffs, every `## lint` guard named in this plan's baseline green, then reverted; git apply --check passes at c93da10
diff --git a/.flow/project.md b/.flow/project.md
index 8ea523c..e3a7ec3 100644
--- a/.flow/project.md
+++ b/.flow/project.md
@@ -365,6 +365,13 @@ still exits non-zero on the identical case, so it is not the mechanism this proj
 `verify` above carries no such flag — it must keep failing on real drift against the committed
 baseline, which is the whole point of a regression gate.
 
+**`fingerprint` compares `index.html`, rebuilt first.** Vite content-hashes its asset filenames and
+rewrites `index.html` to reference them, so the served `index.html` names the exact bundle the
+daemon embeds, and `web.Handler` serves that embedded file byte-for-byte at `/`. The `npm run
+build` in front makes the comparison mean "the worktree's source", not "whatever `dist/` last
+held"; `make ui-test-up` rebuilds too, so a stack this stage started matches first time and only a
+reused stack pays the restart.
+
 | Setting | Value |
 |---------|-------|
 | `ui paths` | `stats/web/src/**` |
@@ -375,3 +382,4 @@ baseline, which is the whole point of a regression gate.
 | `setup` | `cd stats/web && npm install && npx playwright install chromium` |
 | `verify` | `cd stats/web && npm run test:visual` |
 | `capture` | `cd stats/web && npx playwright test <spec> --update-snapshots` |
+| `fingerprint` | `cd stats/web && npm run build && curl -sf http://127.0.0.1:4174/ \| cmp -s - ../internal/web/dist/index.html` |
````

  - [x] **Step 2: Validate the section** — `scripts/check-visual-verification.sh .` prints
    `VISUAL-OK` and no `vocabulary is closed` line.
  - [x] **Step 3: Run the row against a live stack** — with `flow-postgres` up on 5433: `cd stats
    && make ui-test-up`, then from the repository root run the `fingerprint` cell verbatim; it exits
    0. Then the mismatch path: append one character to any file under `stats/web/src/`, run the
    cell again, expect a non-zero exit, and revert the edit. Finish with `cd stats && make
    ui-test-down`.
  - [x] **Step 4: Lint** — `scripts/check-workspace-isolation.sh && scripts/check-contract-budget.sh
    && scripts/check-markdown-integrity.py` exit 0.
  - [x] **Step 5: Commit** with the `**Commit:**` subject above.
