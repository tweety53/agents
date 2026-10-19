# self-review-kan-998-bugfixes-superset-header-focus-mode

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

**Goal:** close kan-998's six self-review findings — the merge-order record visual verification
skips, the absence sweep that stops at the task's own paths, the bare script names no `PATH`
reaches, the live record accepted with no before figure, the stale sidecar line found only at
compose, and the identical baselines no stage names.

**Architecture:** two prose tasks in `skills/flow/`; one `setup.sh` change writing exec shims into
`~/.local/bin`, tested in `stats/internal/setuptest/`; three new `flow-guard` subcommands in
`stats/internal/guard/`, each with its `scripts/<name>.sh` shim, its Go table test and the skill
sentence that runs it.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-19-self-review-kan-998-bugfixes-superset-header-focus-mode-design.md`
in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/` — this change's `verbatim-moves.txt` is already committed with the plan. Never merge or push.
- Every task runs in this worktree, `/Users/tweety53/Projects/agents-worktrees/self-review-kan-998-bugfixes-superset-header-focus-mode`, branch `self-review-kan-998-bugfixes-superset-header-focus-mode`; every `**Files:**` path is relative to its root. Never touch a main checkout.
- A prose diff below is applied with `git apply` from the worktree's root, after saving the lines between its fence markers to a scratch file outside the worktree. Its sentences land character for character: `scripts/check-verbatim-moves.sh` matches them against `verbatim-moves.txt` exactly. A diff that does not apply is reported, never hand-merged by guesswork.
- Every new guard follows `.flow/project.md`'s rule: Go in `stats/internal/guard/<file>.go`, registered as `func init() { Registry["<name>"] = <fn> }`, its `scripts/<name>.sh` a `flow_guard_exec <name> 2 "<name>:" "$@"` shim in the shape of `scripts/check-frame-list.sh` (header comment = the contract: usage, every output line, every exit code, the kan-998 reason), mode 0755, and its test `stats/internal/guard/<name with - as _>_test.go` — the companion `scripts/run-guard-tests.sh` accepts. Tests call the guard through the package's `runGuard(name, args, Env{…})` helper, every input built under `t.TempDir()`.
- Every test-bearing task proves its tests bite: each new test fails with the mutation its RED step names, and passes with it restored.
- Never run `setup.sh global` with `HOME` at the real home; the setuptest harness sandboxes it.
- Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container.

## Review Focus

- Task 3: a user's own file in `~/.local/bin` is never overwritten or removed; a shim of a deleted script is removed on the next run; the harness never writes the real home's `~/.local/bin`.
- Task 4: a record whose only before line is a table header still passes (lenient by design); `Before figure: not measured` fails; an unreadable record exits 2, never 0.
- Task 5: a template literal `` `cardio-guide-${frame}.png` `` reaches `cardio-guide-G1.png`; a sidecar whose spec is absent names every line; `node_modules` and `.worktrees` are never walked.
- Task 6: a fresh capture and its own baseline in different directories are never grouped; an empty stdin exits 2.

- [x] 1. Route visual-verify step 3 to the recipe's merge-order record (F-1.1)
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -c -F 'merge-order record' skills/flow/visual-verify.md`: it prints `0`.
    <!-- measured: the grep above in this worktree @ 46db35e0 → 0 -->
  - [x] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:produced in a scratch clone at 46db35e0; check-references.sh, check-dispatch-paragraphs.sh, check-ask-silence.sh, check-installed-citations.sh and check-markdown-integrity.py exit 0 with it applied
--- a/skills/flow/visual-verify.md
+++ b/skills/flow/visual-verify.md
@@ -193,7 +193,11 @@
    parent runs that entry's recipe — the `kickoff-worktree.sh --from` call, then the regression
    checkout's `setup` (`skills/flow/cross-repo-worktrees.md`) — before the checks below, and adds
    the worktree to the set: kan-983's first verifier was dispatched without it and blocked on the
-   missing checkout, ~4.45M tokens spent for nothing. With the parent's own Bash call, before the
+   missing checkout, ~4.45M tokens spent for nothing. The recipe's merge-order record runs with
+   it, its planning commit included (**Isolate the workspace — additional worktrees**,
+   `skills/flow/cross-repo-worktrees.md`): a repository the canonical `link.md`'s `## Merge order`
+   omits is one run 1 never sequences and the self-review bundle never reads (kan-998). With the
+   parent's own Bash call, before the
    dispatch below, run the four checks — ports held only by their own app, the base-URL
    configuration resolving to the workspace's values, origin allowed, one Playwright checkout per
    workspace — against this worktree:
````

  - [x] **Step 3: Verify.** `grep -c -F 'merge-order record' skills/flow/visual-verify.md` prints `1`;
    `scripts/check-references.sh`, `python3 scripts/check-markdown-integrity.py` and
    `scripts/check-verbatim-moves.sh` exit 0.
  - [x] **Step 4: Commit** the file with the subject below.

**Build:** green
**Files:** `skills/flow/visual-verify.md`
**Tests:** **none** — prose; `check-verbatim-moves.sh` covers its acknowledged sentence.
**Regression:** none — no test is declared.
**Commit:** docs(flow): run the merge-order record with visual-verify's regression worktree
**After:** none

**Decision:** merge-order-in-visual-verify

- [x] 2. Sweep a removed label in the regression checkout too (F-1.2)
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -c -F 'swept in that checkout too' skills/flow/brainstorm-planner.md`: it prints `0`.
    <!-- measured: the grep above in this worktree @ 46db35e0 → 0 -->
  - [x] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:produced in a scratch clone at 46db35e0; check-references.sh and check-markdown-integrity.py exit 0 with it applied
diff --git a/skills/flow/brainstorm-planner.md b/skills/flow/brainstorm-planner.md
index f9518e11..199ea4ff 100644
--- a/skills/flow/brainstorm-planner.md
+++ b/skills/flow/brainstorm-planner.md
@@ -349,6 +349,13 @@ paragraph under **The build-green tag** (`skills/flow-contracts/build-green.md`)
 > an identifier — its last verify step is one `git grep -n -F` over every string it removes, each
 > passed as its own `-e '<string>'`, scoped to the paths the task touches: the sweep prints
 > nothing when the task is done, and any hit means it is not.
+>
+> A removed UI tag, test id or label that a declared `regression checkout` drives
+> (**Project configuration — visual verification**,
+> `skills/flow-contracts/project-configuration-visual.md`) is swept in that checkout too, with
+> the same `-e` list, and the plan carries the task that updates every hit there: kan-998's
+> sweep, scoped to its own paths, missed the labels gymie-playwright drove, and the regression
+> suite failed 83 tests on them.
 
 > **List the rationale appendix a changed rule's reasoning lives in.** A step that changes a rule
 > whose reasoning a `*-rationale.md` appendix records lists that appendix in its task's
````

  - [x] **Step 3: Verify.** `grep -c -F 'swept in that checkout too' skills/flow/brainstorm-planner.md`
    prints `1`; `scripts/check-references.sh`, `python3 scripts/check-markdown-integrity.py` and
    `scripts/check-verbatim-moves.sh` exit 0.
  - [x] **Step 4: Commit** the file with the subject below.

**Build:** green
**Files:** `skills/flow/brainstorm-planner.md`
**Tests:** **none** — prose; `check-verbatim-moves.sh` covers its acknowledged sentence.
**Regression:** none — no test is declared.
**Commit:** docs(flow): extend a removal task's absence sweep to the regression checkout
**After:** none

**Decision:** sweep-regression-checkout

- [x] 3. Install exec shims for the bare-name scripts into ~/.local/bin (F-1.3)
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -c -F '.local/bin' setup.sh`: it prints `0` — `setup.sh global` puts no script on `PATH`.
    <!-- measured: the grep above in this worktree @ 46db35e0 → 0 -->
  - [x] **Step 2: Write the failing test** `TestGlobalInstallsScriptShims` in
    `stats/internal/setuptest/harnesses_test.go`. It builds its own fixture
    (`makeFixtureRepo(filepath.Join(g.dir, "fixture-repo-shims"), false)`, as
    `TestReinstallPrunesDeletedSourceLinks` does) and seeds two scripts into the fixture's scripts directory, `doomed.sh` (0755,
    `#!/bin/sh\necho doomed\n`) and `test-demo.sh` (0755). Then, against `home := g.newHome()`:
    - `g.runSetup(fx, home, "global", "")` → rc 0; `<home>/.local/bin/project-get.sh` and
      `<home>/.local/bin/doomed.sh` exist, are executable (`g.assertExecutable`), are not symlinks,
      and contain the marker line `# flow script shim — written by setup.sh global; do not edit`;
      `<home>/.local/bin/test-demo.sh` is absent (`g.assertAbsent`).
    - Running `<home>/.local/bin/project-get.sh <proj> 'default landing route'`, `<proj>` a
      directory seeded with `.flow/project.md` reading `# p\n\n## default landing route\n\nmerge and push\n`,
      prints `merge and push` and exits 0 — the shim reaches the script's own `lib/`.
    - Delete the fixture's `doomed.sh` script and write `<home>/.local/bin/mine.sh` (no
      marker); re-run → rc 0, `doomed.sh` absent from the bin dir, `mine.sh` unchanged.
    - In a fresh home, pre-write `<home>/.local/bin/doomed.sh` without the marker, re-seed
      the fixture's `doomed.sh` script, run → rc 1 (SKIPPED), the pre-written file byte-identical.
  - [x] **Step 3: Run it to verify it fails.** `cd stats && go test ./internal/setuptest/ -run 'TestGlobalInstallsScriptShims' -count=1` → FAIL: `.local/bin/project-get.sh` absent.
  - [x] **Step 4: Implement `install_script_shims <bin dir>` in `setup.sh`** and call it from
    `install_global` as `install_script_shims "$home_dir/.local/bin"`, after `install_hooks_zcode`;
    add a `Scripts  → $home_dir/.local/bin/` line to the banner. The function: `mkdir -p` the dir;
    remove every regular file in it whose content carries the marker line; then for each
    `"$SCRIPT_DIR"/scripts/*.sh` whose basename does not start with `test-`: if a file of that
    name exists (no marker — the marked ones are gone), `warn` naming it and increment `SKIPPED`;
    otherwise write
    `#!/bin/sh` / the marker line / `exec '<abs script path>' "$@"` and `chmod 0755` it. A header
    comment states why a shim and not a symlink (the `BASH_SOURCE` sibling load) and why
    `~/.local/bin` (on `PATH`, holds `flow`). A `SCRIPT_DIR` containing `'` is refused with `die`
    before anything is written (`# ponytail:` comment — single-quoted exec target).
  - [x] **Step 5: Extend `realHomeFingerprint`** (`stats/internal/setuptest/fingerprint_test.go`)
    to list `<home>/.local/bin` the way it lists `skills`, so a shim leaking into the real home
    fails the harness's closing fingerprint check.
  - [x] **Step 6: RED, then GREEN.** With the call in `install_global` commented out, the test fails on the absent shim; restored, `cd stats && go test ./internal/setuptest/ -run 'TestGlobalInstallsScriptShims' -count=1` passes.
  - [x] **Step 7: Verify.** `bash -n setup.sh`; `cd stats && gofmt -l ./internal/setuptest` prints nothing; `cd stats && go vet ./internal/setuptest`; `scripts/check-installed-rules.sh` exits 0.
  - [x] **Step 8: Commit** the three files with the subject below.

**Build:** green
**Files:** `setup.sh`, `stats/internal/setuptest/harnesses_test.go`, `stats/internal/setuptest/fingerprint_test.go`
**Tests:** `TestGlobalInstallsScriptShims`
**Regression:** reverting the commit leaves `~/.local/bin` without shims — `TestGlobalInstallsScriptShims` fails on the absent `project-get.sh` shim.
**Commit:** fix(setup): install exec shims for the flow scripts into ~/.local/bin
**After:** none

**Decision:** script-shims-local-bin

- [x] 4. Refuse a live-verification record with no measured before figure (F-1.4)
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run
    `ls scripts/check-live-verification.sh`: it fails, no such file — nothing reads the record.
    <!-- measured: the ls above in this worktree @ 46db35e0 → No such file or directory -->
  - [x] **Step 2: Write the failing test** `TestCheckLiveVerification` in
    `stats/internal/guard/check_live_verification_test.go`, a table over a record written to
    `t.TempDir()`:
    - `before table header passes`: `| Check | Before | After |\n|---|---|---|\n| a | 1 | 2 |\n` → 0, stdout `LIVE-VERIFICATION-OK: <record>`.
    - `before line passes`: `- Before: 40 requests\n- After: 28\n` → 0.
    - `no before figure fails`: `| Check | After |\n| a | 2 |\n` → 1, stdout `<record>: no before figure — no line names one`.
    - `before not measured fails` (kan-998's line): `| Check | After |\n\nBefore figure: not measured on a merge-base stack; the pre-change shape is what V12 created.\n` → 1, stdout `<record>:3: before figure not measured — Before figure: not measured on a merge-base stack; the pre-change shape is what V12 created.`
    - `unmeasured and not taken fail`: `before: unmeasured\nBEFORE not taken\n` → 1, two lines, `:1:` and `:2:`.
    - `word inside another word is no before`: `beforehand 3\n` → 1, `no before figure`.
    - exit 2 with nothing on stdout for: no argument, two arguments, an absent path, a directory.
  - [x] **Step 3: Run it to verify it fails.** `cd stats && make -s web-dist && go test ./internal/guard -run 'TestCheckLiveVerification'` → FAIL: unknown guard `check-live-verification`.
  - [x] **Step 4: Implement `checkLiveVerification(args []string, env Env, stdout, stderr io.Writer) int`** in `stats/internal/guard/liveverification.go`, registered as `check-live-verification`: one argument, the record; a line names a before figure when it matches `(?i)\bbefore\b`; it is unmeasured when it also matches `(?i)\b(not measured|unmeasured|not taken)\b`. Output exactly as the test pins; usage and read errors go to stderr prefixed `check-live-verification: `. Write `scripts/check-live-verification.sh` per Global constraints.
  - [x] **Step 5: Apply the skill sentence** (`git apply <the saved diff>`):

````diff verified:produced in a scratch clone at 46db35e0; check-references.sh and check-markdown-integrity.py exit 0 with it applied
diff --git a/skills/flow/verify-and-handoff.md b/skills/flow/verify-and-handoff.md
index 24c00751..238a05e1 100644
--- a/skills/flow/verify-and-handoff.md
+++ b/skills/flow/verify-and-handoff.md
@@ -147,7 +147,11 @@ planning path the write-in-progress planning commit carries. **A before figure i
 detached worktree at the merge base** — `git -C <worktree> worktree add --detach
 <project>-worktrees/<name>-before <merge-base>`, the base's stack run from there, then stopped, and the worktree
 removed with `git worktree remove` once the figure is written — never by checking base sources out
-over the change worktree. The `## Report` gains one line, `live check — <matched | failure look>`. A figure that
+over the change worktree. Then run `check-live-verification.sh
+<changeRoot>/live-verification.md`; its header is canonical for the record it accepts. Exit 1 → the
+record carries no before figure, or states one was not measured: measure it from the merge base as
+above and rewrite the record, never proceed past it (kan-998). Exit 2 → it cannot answer; correct
+the input and run it again. The `## Report` gains one line, `live check — <matched | failure look>`. A figure that
 shows the failure look is the branch's own defect and takes **The loop**
 (`skills/flow/verify-fix-loop.md`); a stack that will not start ends the turn as a failed
 command's environment cause does. **A step of the check the operator waives runs on its cheapest safe proxy instead** — a scratch
````

  - [x] **Step 6: RED, then GREEN.** Drop `|unmeasured|not taken` and `not measured` from the marker pattern → the three marker cases fail; restore → `go test ./internal/guard -run 'TestCheckLiveVerification'` passes. Then `scripts/check-live-verification.sh` over kan-998's record text (the `not measured` case's body in a scratch file) exits 1.
  - [x] **Step 7: Verify.** `cd stats && gofmt -l ./internal/guard` prints nothing; `cd stats && make -s web-dist && go vet ./internal/guard`; `scripts/check-guard-symlinks.sh`, `scripts/check-references.sh`, `python3 scripts/check-markdown-integrity.py` and `scripts/check-verbatim-moves.sh` exit 0.
  - [x] **Step 8: Commit** the four files with the subject below.

**Build:** green
**Files:** `stats/internal/guard/liveverification.go`, `stats/internal/guard/check_live_verification_test.go`, `scripts/check-live-verification.sh`, `skills/flow/scripts/check-live-verification.sh`, `skills/flow/verify-and-handoff.md`
**Tests:** `TestCheckLiveVerification`
**Regression:** reverting the commit removes the guard — `TestCheckLiveVerification` is gone with it, and a record stating its before figure was not measured passes **Live check** again.
**Commit:** feat(guard): refuse a live-verification record with no measured before figure
**After:** none

**Decision:** live-record-before-rule

- [x] 5. Pre-flight every sidecar line against its spec (F-4.1)
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run
    `ls scripts/check-sidecar-reach.sh`: it fails, no such file — a stale sidecar line is found only by `compose-mockup-frames.sh` after the capture.
    <!-- measured: the ls above in this worktree @ 46db35e0 → No such file or directory -->
  - [x] **Step 2: Write the failing test** `TestCheckSidecarReach` in
    `stats/internal/guard/check_sidecar_reach_test.go`, a table over a capture root under
    `t.TempDir()` holding `fidelity/a.spec.ts` and `fidelity/a.spec.ts.mockups`:
    - `literal reaches`: spec `await capture(page, 'a-1.png');`, sidecar `# c\n\na-1.png A-1\n` → 0, stdout `SIDECAR-REACH-OK: 1 lines in 1 sidecars`.
      <!-- predicted: go test ./internal/guard -run 'TestCheckSidecarReach' after Task 5 -->
    - `double-quoted literal reaches`: spec `shot("a-2.png")`, sidecar `a-2.png A-2` → 0.
    - `template literal reaches`: spec ``toHaveScreenshot(`cardio-guide-${frame}.png`)``, sidecar `cardio-guide-G1.png G1` → 0.
    - `stale line fails`: spec `'a-1.png'`, sidecar `a-1.png A-1\npin-p2-pad-up.png pin-P2\n` → 1, stdout `fidelity/a.spec.ts.mockups:2: pin-p2-pad-up.png — taken by no screenshot in fidelity/a.spec.ts`.
    - `absent spec fails`: sidecar `fidelity/gone.spec.ts.mockups` with `g-1.png G1`, no spec → 1, stdout `fidelity/gone.spec.ts.mockups:1: g-1.png — fidelity/gone.spec.ts does not exist`.
    - `pruned dirs are never walked`: a stale sidecar under `node_modules/x/`, `.git/` and `.worktrees/c/` → 0.
    - `a substring is no literal`: spec `'a-10.png'`, sidecar `a-1.png A-1` → 1.
    - exit 2 with nothing on stdout for: no argument, two arguments, a root that is not a directory.
  - [x] **Step 3: Run it to verify it fails.** `cd stats && make -s web-dist && go test ./internal/guard -run 'TestCheckSidecarReach'` → FAIL: unknown guard `check-sidecar-reach`.
  - [x] **Step 4: Implement `checkSidecarReach(args []string, env Env, stdout, stderr io.Writer) int`** in `stats/internal/guard/sidecarreach.go`, registered as `check-sidecar-reach`: `filepath.WalkDir` the root, skipping directories named `node_modules`, `.git`, `.worktrees`; a sidecar is a file ending `.mockups`, its spec that path without the suffix; a line's name is its first whitespace field, blank and `#` lines skipped. The spec's literals are the matches of `'([^'\n]*)'|"([^"\n]*)"|` + "`([^`]*)`"; a name is reached when it equals a literal, or fully matches a backtick literal containing `${`, built as `^` + the `regexp.QuoteMeta` of each piece between `${…}` holes joined by `.+` + `$`. Paths in output are relative to the root, sorted by sidecar path then line. Write `scripts/check-sidecar-reach.sh` per Global constraints, its header stating the stated limit: a name built outside any literal (concatenation, a computed variable) reads as unreached.
  - [x] **Step 5: Apply the skill paragraph** (`git apply <the saved diff>`):

````diff verified:produced in a scratch clone at 46db35e0 after Task 1's diff; check-references.sh, check-dispatch-paragraphs.sh and check-markdown-integrity.py exit 0 with it applied
--- a/skills/flow/visual-verify.md
+++ b/skills/flow/visual-verify.md
@@ -220,6 +220,13 @@
    starts again only once it has. Exit 2 → it cannot answer; treat it as a failing check, carrying
    its stderr.
 
+   When `mockups` is declared, the same Bash call also runs `check-sidecar-reach.sh <capture
+   root>`, `<capture root>` the `regression checkout`'s worktree when one is declared, else this
+   worktree; its header is canonical for the rule. Exit 1 → each line names a sidecar line whose
+   screenshot no spec beside it takes, a failing check: kan-998's stale lines surfaced only at
+   compose, after a 13-minute capture. Exit 2 → it cannot answer; treat it as a failing check,
+   carrying its stderr.
+
    **Any failing check ends the stage here**: no verifier is dispatched, the stage's `end` mark
    carries `-outcome stopped`, and the report names every failing check with the evidence above.
    The run proceeds to the `IN_PROGRESS` handoff, which names the environment cause — the operator
````

  - [x] **Step 6: RED, then GREEN.** Replace the template branch with `false` → `template literal reaches` fails; restore → `go test ./internal/guard -run 'TestCheckSidecarReach'` passes. Then `scripts/check-sidecar-reach.sh ~/Projects/gymie-playwright` exits 0 (read-only; the spike measured 231 lines, 0 misses).
    <!-- measured: the python3 spike of this rule over ~/Projects/gymie-playwright @ e864161 → 231 lines, 0 misses -->
  - [x] **Step 7: Verify.** `cd stats && gofmt -l ./internal/guard` prints nothing; `cd stats && make -s web-dist && go vet ./internal/guard`; `scripts/check-guard-symlinks.sh`, `scripts/check-references.sh`, `python3 scripts/check-markdown-integrity.py` and `scripts/check-verbatim-moves.sh` exit 0.
  - [x] **Step 8: Commit** the four files with the subject below.

**Build:** green
**Files:** `stats/internal/guard/sidecarreach.go`, `stats/internal/guard/check_sidecar_reach_test.go`, `scripts/check-sidecar-reach.sh`, `skills/flow/scripts/check-sidecar-reach.sh`, `skills/flow/visual-verify.md`
**Tests:** `TestCheckSidecarReach`
**Regression:** reverting the commit removes the guard and its pre-flight call — a sidecar line whose test is gone is found again only when compose fails after the capture.
**Commit:** feat(guard): pre-flight every mockups sidecar line against its spec
**After:** Task 1

**Decision:** sidecar-reach-literals

- [x] 6. Name byte-identical baselines for the verifier to explain (F-4.2)
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run
    `ls scripts/check-identical-baselines.sh`: it fails, no such file — no stage compares baselines.
    <!-- measured: the ls above in this worktree @ 46db35e0 → No such file or directory -->
  - [x] **Step 2: Write the failing test** `TestCheckIdenticalBaselines` in
    `stats/internal/guard/check_identical_baselines_test.go`, a table whose stdin is absolute PNG
    paths, one per line, of files written under `t.TempDir()`:
    - `distinct files pass`: `s-snapshots/up.png` ≠ `s-snapshots/down.png` → 0, stdout `IDENTICAL-NONE: 2 PNGs`.
    - `identical pair fails`: `s-snapshots/pin-p2-pad-up.png` == `s-snapshots/pin-p3-pad-down.png` → 1, stdout `identical: <dir>/s-snapshots: pin-p2-pad-up.png pin-p3-pad-down.png`.
    - `group of three on one line`: three equal files in one dir → 1, one line naming the three, sorted.
    - `different directories never group`: equal `results/up.png` and `s-snapshots/up.png` → 0.
    - `blank stdin lines are skipped`: a blank line between two distinct paths → 0.
    - exit 2 with nothing on stdout for: an argument given, empty stdin, a path that does not exist.
  - [x] **Step 3: Run it to verify it fails.** `cd stats && make -s web-dist && go test ./internal/guard -run 'TestCheckIdenticalBaselines'` → FAIL: unknown guard `check-identical-baselines`.
  - [x] **Step 4: Implement `checkIdenticalBaselines(args []string, env Env, stdout, stderr io.Writer) int`** in `stats/internal/guard/identicalbaselines.go`, registered as `check-identical-baselines`: no arguments; read `env.Stdin` (nil reads as empty); group the paths by `filepath.Dir`, then by SHA-256 of the file's bytes (`crypto/sha256`); print groups of two or more sorted by directory then first name. Write `scripts/check-identical-baselines.sh` per Global constraints, its header naming its input as `resolve-visual-screenshots.sh`'s output.
  - [x] **Step 5: Apply the verifier sentences** (`git apply <the saved diff>`):

````diff verified:produced in a scratch clone at 46db35e0; check-references.sh, check-dispatch-paragraphs.sh and check-markdown-integrity.py exit 0 with it applied
diff --git a/skills/flow/visual-verify-verifier.md b/skills/flow/visual-verify-verifier.md
index 3354558d..77bdd546 100644
--- a/skills/flow/visual-verify-verifier.md
+++ b/skills/flow/visual-verify-verifier.md
@@ -127,6 +127,12 @@ never an exit code the command did not return.
    that never rendered; exit 2 (cannot answer) blocks the same way. **Read every printed path** — no
    script can do that — and state, per view, what was seen. An unreadable PNG is reported and blocks
    too.
+
+   Then pipe the same resolver output into `check-identical-baselines.sh`; its header is canonical
+   for what it compares. Exit 1 → each `identical:` line names PNGs of one directory that are
+   byte-identical, and the report carries one `- identical:` line per group stating why those views
+   draw the same pixels, or the defect that made them, which blocks: kan-998's pad-up and pad-down
+   baselines were one image, and only a reviewer saw it. Exit 2 → it cannot answer, and blocks.
 10. **Compose captured frames against their mockups, if `mockups` is declared.** With a
    `<spec>.mockups` sidecar beside the capture spec, run
 
@@ -570,6 +576,7 @@ or a measurement the pixels could not give is step 10 undone.
 - full app suite: <created | updated> — <absolute spec path>, <n> screens, capture exit <n>
 - full app suite zip: <absolute zip path>, <n> PNGs | not written — <reason>
 - spec reach: exit <n>
+- identical: <the PNG names of one byte-identical group> — <why the views draw the same pixels, or the defect> | none
 - spec: <absolute spec path>
 - <view>: <absolute PNG path> — <what was seen, including any defect>
 - mockups: <not declared | no map for <spec> | exit <n>>
````

  - [x] **Step 6: RED, then GREEN.** Group by hash alone, dropping the directory key → `different directories never group` fails; restore → `go test ./internal/guard -run 'TestCheckIdenticalBaselines'` passes.
  - [x] **Step 7: Verify.** `cd stats && gofmt -l ./internal/guard` prints nothing; `cd stats && make -s web-dist && go vet ./internal/guard`; `scripts/check-guard-symlinks.sh`, `scripts/check-references.sh`, `python3 scripts/check-markdown-integrity.py`, `scripts/check-dispatch-paragraphs.sh` and `scripts/check-verbatim-moves.sh` exit 0.
  - [x] **Step 8: Commit** the four files with the subject below.

**Build:** green
**Files:** `stats/internal/guard/identicalbaselines.go`, `stats/internal/guard/check_identical_baselines_test.go`, `scripts/check-identical-baselines.sh`, `skills/flow/scripts/check-identical-baselines.sh`, `spectre/changes/self-review-kan-998-bugfixes-superset-header-focus-mode/verbatim-moves.txt`, `skills/flow/visual-verify-verifier.md`
**Tests:** `TestCheckIdenticalBaselines`
**Regression:** reverting the commit removes the guard and the verifier's `- identical:` line — two byte-identical baselines in one snapshot directory pass visual verification unnamed again.
**Commit:** feat(guard): name byte-identical baselines for the verifier to explain
**After:** none

**Decision:** identical-baselines-explained
