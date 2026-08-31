# kan-279-give-the-parent-a-mutation-helper

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.

> **Relocation:** no

Four tasks. Task 1 is the script itself (verified by syntax-check alone, since its own behavioral
test harness is task 2's own commit). Task 2 adds that harness. Task 3 installs the script the same
way every other guard/util script is installed. Task 4 updates the one skill file that documents the
duty this script mechanizes.

- [x] 1. Write `scripts/mutate-and-verify.sh`

  Create `scripts/mutate-and-verify.sh`, `set -euo pipefail`, bash 3.2-compatible (indexed arrays
  only, no associative arrays, no `wait -n` — this repository's floor, per `run-guard-tests.sh`'s own
  header).

  **Usage:** `mutate-and-verify.sh <patch-file> <harness>...` — at least one harness required.
  Missing either → usage line on stderr, exit 4.

  **Exit codes** (mirrors `run-reproducer.sh`'s own philosophy: the code reports this script's own
  mechanics, never a verdict on the mutation):
  - `0` — ran clean: the patch applied, every harness ran before and after, the mutation was
    reverted, and the touched files were verified clean afterward. The per-harness verdict (surviving
    mutant / caught / suspicious blast radius) is in the report body, not the exit code.
  - `2` — refused before mutating anything: the patch does not apply cleanly, a file it touches
    already has uncommitted changes, or a named harness does not exist or is not executable.
  - `3` — could not fully restore: after `git checkout --` the touched files are still not clean;
    the residual `git status --porcelain` output is printed and the files may still be mutated.
  - `4` — cannot answer: bad usage, not inside a git worktree, or a harness produced no readable
    `ok:`/`FAIL:` line at all on either run (so pass/fail counts cannot be computed).

  **Steps the script performs, in order:**

  1. Resolve the repo root via `git rev-parse --show-toplevel`; exit 4 if that fails (not a git
     worktree).
  2. Verify `<patch-file>` exists and is readable; exit 4 otherwise.
  3. `git apply --check "<patch-file>"` — on failure, refuse with exit 2 and the check's own stderr,
     without touching anything.
  4. Derive the touched file list from `git apply --numstat "<patch-file>"` (its third column).
  5. For each named harness: `-f` and `-x` or refuse with exit 2, naming the harness.
  6. `git status --porcelain -- <touched files>` — refuse with exit 2 if non-empty (a file the patch
     touches already has uncommitted changes; never conflate that with the mutation's own diff at
     restore time).
  7. Install an `EXIT` trap now, before anything mutates: on any exit, if the patch was applied
     (tracked in a flag variable), run `git checkout -- <touched files>` best-effort, then re-check
     `git status --porcelain -- <touched files>`; if still non-empty, print the residual status on
     stderr and force the script's own exit code to `3` (overriding whatever exit code the main flow
     was already carrying), so a mutated file is never left behind on any exit path, this one
     included.
  8. Define a helper that runs one harness, captures its stdout, and returns: total `^ok: ` count,
     total `^FAIL: ` count, and the set of `FAIL:` case names (one per line). If a harness's output
     contains neither an `ok:` nor a `FAIL:` line, that is exit 4 (cannot answer), named per harness.
  9. **Baseline pass** — run every harness once, before applying the patch, through the helper above.
     Report a harness whose baseline already carries failures explicitly ("baseline NOT clean —
     <names>") rather than silently mixing them into the mutated count.
  10. Apply the mutation: `git apply "<patch-file>"`; set the applied-flag the trap checks.
  11. **Mutated pass** — run every harness again through the same helper.
  12. Per harness, compute `new failures` = mutated FAIL-name set minus baseline FAIL-name set (a
      `comm`/`grep -Fxv` set difference — no associative arrays, matching the bash-3.2 floor).
  13. Resolve the blast-radius bound: `MUTATE_AND_VERIFY_MAX_NEW_FAILURES` if set and numeric,
      else default `5` (an env-var override, same idiom as `RUN_REPRODUCER_BOUND_SECONDS` and
      `RUN_GUARD_TESTS_ROOT`).
  14. Per-harness verdict: 0 new failures → `surviving mutant`; 1..bound → `caught` (name the new
      failures); more than bound → `suspicious blast radius` (name the count, the bound, and the new
      failures) — printed as a flag, never as a pass.
  15. Print the full report: patch path and the files it touches, baseline-not-clean notes (if any),
      per-harness baseline/mutated ok+fail counts, per-harness verdict with the new-failure names,
      then let the `EXIT` trap perform the restore and its own clean-check.

  Verify: `bash -n scripts/mutate-and-verify.sh` exits 0.

**Build:** green
**Files:** `scripts/mutate-and-verify.sh`
**Tests:** **none** added — this task's own behavioral coverage is task 2's commit; this task's
verification is the syntax check above.
**Regression:** reverting this task removes the script entirely; task 2's harness (which requires
it) and task 3's symlink (which points at it) would then both fail, so a partial revert is caught
by either.
**Baseline:** n/a — no prior version of this script exists.
**Commit:** `feat(scripts): add mutate-and-verify.sh`

- [x] 2. Write `scripts/test-mutate-and-verify.sh`

  Follow `scripts/test-check-base-moved.sh`'s fixture pattern: a real, throwaway git repository per
  case under `mktemp -d`, removed by an `EXIT` trap, `fail()`/`pass()` helpers printing `FAIL: <case>`
  / `ok: <case>` (matching `test-check-base-moved.sh`'s own `fail()`/`pass()`), and a final
  `FAILURES` count gating the harness's own exit code.

  Each case's fixture is a small git repo containing one fixture "guard" script and one fixture
  `test-*.sh` harness for it (its own `ok:`/`FAIL:` lines, 2–3 cases, cheap and fast — these are
  fixtures for task 1's script, not real guards). Cases:

  - **caught** — a patch that changes the fixture guard's behavior in a way exactly one fixture
    harness case catches. Assert exit 0, the report names verdict `caught` with exactly that one new
    failure, and after the run `git status --porcelain` for the touched file is empty.
  - **surviving mutant** — a patch that changes the fixture guard in a way no fixture harness case
    catches. Assert exit 0, verdict `surviving mutant`, no new failures named, tree clean afterward.
  - **suspicious blast radius** — a fixture harness with more than 5 cases, and a patch that breaks
    the fixture guard broadly (e.g. an early unconditional `exit 1`) so every case fails. Assert
    exit 0, verdict `suspicious blast radius`, the reported new-failure count exceeds the default
    bound of 5, tree clean afterward.
  - **patch does not apply** — a patch built against text the fixture file does not contain. Assert
    exit 2, the fixture file is byte-for-byte unchanged, and no harness ran at all (stdout/stderr
    carry no baseline or mutated report).
  - **dirty touched file refused** — before invoking the script, make an uncommitted edit to the file
    the patch also touches. Assert exit 2, the uncommitted edit is still present and unchanged (never
    reverted), and the patch was never applied.
  - **missing harness refused** — pass a harness path that does not exist. Assert exit 2, naming the
    harness; the patch was never applied.
  - **non-executable harness refused** — pass a harness path that exists but is not `+x`. Assert
    exit 2, naming the harness.
  - **MUTATE_AND_VERIFY_MAX_NEW_FAILURES override** — the same broad-breakage patch as the blast
    radius case, run with the env var raised above the actual new-failure count. Assert exit 0,
    verdict `caught` rather than `suspicious blast radius` (proving the bound is read, not
    hardcoded).

  Verify: `scripts/test-mutate-and-verify.sh` prints `all cases passed`, and `bash -n
  scripts/test-mutate-and-verify.sh` exits 0. Then `scripts/run-guard-tests.sh` — the new harness is
  picked up by its glob automatically, no wiring needed.

**Build:** green
**Files:** `scripts/test-mutate-and-verify.sh`
**Tests:** eight new cases in `scripts/test-mutate-and-verify.sh` — caught, surviving mutant,
suspicious blast radius, patch-does-not-apply refusal, dirty-touched-file refusal, missing-harness
refusal, non-executable-harness refusal, and the blast-radius env-var override.
**Regression:** reverting this task leaves `mutate-and-verify.sh` with no behavioral coverage at
all — a later defect in it (failing to restore, a broken set-difference, a hardcoded bound) would
ship silently until the parent hit it by hand.
**Baseline:** before=46 test-*.sh harnesses, after=47.
<!-- measured: ls scripts/test-*.sh | wc -l @ branch main, before task 1 lands -->
**Commit:** `test(scripts): add mutate-and-verify.sh's own harness`

- [x] 3. Install the script alongside the skill

  1. `ln -s ../../../scripts/mutate-and-verify.sh skills/flow/scripts/mutate-and-verify.sh` — same
     relative-symlink shape every other guard/util script under `skills/flow/scripts/` already uses
     (confirmed against `check-base-moved.sh` and `run-reproducer.sh`'s own symlinks).
  2. In `skills/flow/SKILL.md`'s **Reading the state** section, add `mutate-and-verify.sh` to the
     guard-presence-check literal list — alphabetically, between `commit-split.sh` and
     `plan-dispatch-bundles.sh`.

  Verify: `scripts/check-guard-symlinks.sh` exits 0, and `ls -la skills/flow/scripts/mutate-and-verify.sh`
  shows the symlink resolving to `../../../scripts/mutate-and-verify.sh`.

**Build:** green
**Files:** `skills/flow/scripts/mutate-and-verify.sh` (new symlink), `skills/flow/SKILL.md`
**Tests:** **none** added — this task installs an existing script rather than adding new behavior;
`scripts/check-guard-symlinks.sh` verifies the installation but adds no case of its own for this one
script by name (it scans by glob, per its own rule 2).
**Regression:** reverting this task removes the symlink and the guard-list entry; a `/flow` run would
then report `mutate-and-verify.sh` missing from `skills/flow/scripts/` in its guard-presence check
(**Guard presence check**, `skills/flow-contracts/pipeline.md`) rather than silently having no
script at all.
**Baseline:** n/a — no prior symlink or list entry exists.
**Commit:** `chore(flow): install mutate-and-verify.sh alongside the skill`

- [x] 4. Point the fix round's mutation-proof duty at the script

  In `skills/flow/review-panel.md`'s **The fix round mutation-proves what it changed** section, add
  a paragraph naming `scripts/mutate-and-verify.sh` as the mechanism for the mechanical steps —
  backup, apply, run, report, restore — while keeping unchanged, explicitly, that **which** mechanism
  to mutate and whether a survivor is real or equivalent remain the parent's own judgment calls (the
  existing prose already says this; do not weaken it). Note that the `fix-mutation:` log-line content
  (`<path> — <what was mutated> — <the test that failed>`) can be drawn directly from the script's own
  per-harness report line rather than typed by hand.

  Do not change the requirement itself, the `fix-mutation:`/`fix-mutations-total:` line format, or
  any marker-block rule in this section — this task only names the tool available for the mechanics
  already required.

  Verify: `scripts/check-references.sh` and `scripts/check-contract-budget.sh` both exit 0 against
  the edited file.

**Build:** green
**Files:** `skills/flow/review-panel.md`
**Tests:** **none** added — this task adds prose to a skill file, verified by
`scripts/check-references.sh` and `scripts/check-contract-budget.sh`, neither of which this task's
own commit is expected to trip (a citation of an existing script's basename, within the file's
existing budget).
**Regression:** reverting this task leaves the mutation-proof section undocumented as using the new
script — no functional regression, since the script itself (tasks 1–3) is unaffected, only the
parent's own instructions for finding it.
**Baseline:** n/a — a prose-only addition.
**Commit:** `docs(review-panel): point the fix round at mutate-and-verify.sh`
