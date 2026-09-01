# kan-288-skip-dispatch-context-bundle-rebuild-unchanged

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.

> **Relocation:** no

Two tasks. Task 1 is the script's own contract change plus its test harness — the two must land
together because a caller that starts passing `<output-path>` against the old stdout-only script
would just have that argument silently ignored as a stray positional. Task 2 is every prose caller
following that new contract, split out because it touches no code the first task didn't already
make correct — only the invocation shape and the reporting each dispatching stage already does.

- [x] 1. Give `gather-dispatch-context.sh` an `<output-path>` argument and a skip-when-unchanged write

  `scripts/gather-dispatch-context.sh` currently takes four arguments and prints the whole bundle
  to stdout, leaving the caller to redirect it. Add a required 5th argument, `<output-path>`, and
  have the script write there itself:

  1. Extend the argument check: `OUTPUT_PATH="${5:-}"` alongside the existing four, added to the
     same `if [ -z ... ]` guard so a missing 5th argument is exit 2 with the usage line, updated to
     read `usage: gather-dispatch-context.sh <worktree> <change-root> <name> <principles-path>
     <output-path>`.
  2. Refactor the final section — from the `echo "# Dispatch context bundle for $NAME"` line to the
     trailing `exit 0` — so everything currently printed **after** the `generated:`/`head:` header
     lines (the found/skipped/refused census and every `## <label>` section) is produced by a
     function, `render_body`, and captured once: `BODY="$(render_body)"`. The header lines
     themselves (`# Dispatch context bundle for $NAME`, `generated: …`, `head: …`) stay outside
     `render_body` — they are metadata about the call, not part of what a dispatch reads, and must
     never enter the hash (`design.md`'s `header-excluded-from-hash`).
  3. Add a `sha256_hex` helper, modelled on `scripts/check-cleanup-complete.sh`'s function of the
     same name (`shasum -a 256`, falling back to `sha256sum` then `openssl dgst -sha256`, selecting
     the 64-hex-character field by shape) — copied inline rather than extracted to
     `scripts/lib/`, since this script's callers reach it through the same
     `skills/*/scripts/` symlink farm `check-cleanup-complete.sh` itself is already reached
     through, and extracting a shared helper for exactly two present-day callers is exactly the
     abstraction `rules/build-the-simplest-thing.mdc` asks to defer until there is a reason beyond
     "it would be reusable."
  4. Compute `NEW_HASH="$(sha256_hex "$BODY" 2>/dev/null || true)"` and `HASH_PATH="$OUTPUT_PATH.hash"`.
     Decide `REASON` before writing anything:
     - `NEW_HASH` empty (no hash tool on this machine) → `REASON="no hash tool available"`.
     - `HASH_PATH` or `OUTPUT_PATH` missing → `REASON="no cached bundle"`.
     - `$(cat "$HASH_PATH")` differs from `$NEW_HASH` → `REASON="inputs changed"`.
     - otherwise `REASON` stays empty.
  5. `REASON` empty → print `gather-dispatch-context: bundle unchanged — reusing $OUTPUT_PATH` to
     stderr and `exit 0` **without touching `$OUTPUT_PATH` or `$HASH_PATH`**.
  6. `REASON` non-empty → `mkdir -p "$(dirname "$OUTPUT_PATH")"` (defensive; every existing caller
     already does this before invoking the script, but the script now owns the write), then write
     header + `$BODY` to `$OUTPUT_PATH`; if `NEW_HASH` is non-empty write it to `$HASH_PATH`,
     otherwise `rm -f "$HASH_PATH"` (a stale hash must not survive a run where hashing itself is
     unavailable, or the next run with a hash tool present would compare against it and wrongly
     skip). Print `gather-dispatch-context: bundle rebuilt — $REASON` to stderr, then `exit 0`.
  7. Update the file's own header comment: the "Prints one bundle to stdout" paragraph now
     describes writing to `<output-path>` (falling back to reusing the existing file, unchanged,
     when its hash matches), and the usage line gains `<output-path>`.

  Everything else in the script — the three-step path validation, `add_fixed_source`'s
  found/skipped/refused disposition, the project-commands extraction, the change-name allowlist —
  is unchanged.

  **Add `scripts/test-gather-dispatch-context.sh`**, following `scripts/test-generate-relocation-
  comparison.sh`'s fixture pattern: a fresh directory tree under `mktemp -d
  "${TMPDIR:-/tmp}/gather-dispatch-context-test.XXXXXX"` per case (worktree root, a
  `spectre/changes/<name>/` change-root inside it with `proposal.md`/`design.md`/`tasks.md`, a
  principles file, and `.flow/project.md`), cleaned up in a `trap … EXIT`. Cases:

  - **missing `<output-path>`** — call with only the original four arguments; assert exit 2 and the
    usage line names `<output-path>`.
  - **first call, no existing bundle** — assert exit 0, `$OUTPUT_PATH` and `$OUTPUT_PATH.hash`
    both created, stderr contains `bundle rebuilt — no cached bundle`.
  - **second call, nothing changed** — capture `$OUTPUT_PATH`'s mtime and content, call again,
    assert exit 0, stderr contains `bundle unchanged — reusing`, and both the mtime and the content
    are byte-for-byte unchanged.
  - **edit `tasks.md`, call again** — assert exit 0, stderr contains `bundle rebuilt — inputs
    changed`, and `$OUTPUT_PATH`'s content now contains the edited text.
  - **delete only `$OUTPUT_PATH.hash`, call again with no other change** — assert exit 0, stderr
    contains `bundle rebuilt — no cached bundle` (a missing cache is treated as absent, not as "no
    change to compare against").
  - **no hash tool on `PATH`** — run with `PATH` overridden to a directory containing none of
    `shasum`/`sha256sum`/`openssl`; assert exit 0, stderr contains `no hash tool available`, and
    `$OUTPUT_PATH` is rewritten every call (no stale `.hash` file survives to cause a wrong skip on
    a later call once a hash tool is available again).

  Verify: `scripts/test-gather-dispatch-context.sh` prints `all cases passed`, and
  `bash -n scripts/gather-dispatch-context.sh` exits 0. Then run the project's `## lint` list and
  confirm it is clean.

**Build:** green
**Files:** `scripts/gather-dispatch-context.sh`, `scripts/test-gather-dispatch-context.sh`
**Tests:** five new cases in `scripts/test-gather-dispatch-context.sh` — first call, unchanged
second call, rebuild-on-edit, rebuild-on-missing-hash-file, no-hash-tool fallback — plus the
existing case 9 retargeted to cover the missing-`<output-path>`-argument shape instead of
duplicating it. **Correction:** this plan assumed `scripts/test-gather-dispatch-context.sh` did
not exist. It already did, with 27 cases covering the script's path-validation, symlink-handling
and project-commands-extraction behavior, predating this change. The implementer found this during
task 1 and merged the new cases into the existing suite rather than overwriting it; the numbers
below are corrected to match.
**Regression:** reverting this task returns the script to always printing to stdout with no
`<output-path>` argument at all, so task 2's callers (passing that argument) would fail with the
old four-argument usage error — the five new cases, plus the retargeted case 9, catch a partial
revert of just the hashing logic while the argument stays.
**Baseline:** before=48 after=56 assertions in `scripts/test-gather-dispatch-context.sh` (32 cases;
was 27 cases / 48 assertions before this task).
<!-- measured: scripts/test-gather-dispatch-context.sh | grep -c '^ok:' @ branch spectre/kan-288-skip-dispatch-context-bundle-rebuild-unchanged, commit f3261b3 -->
**Commit:** `feat(scripts): skip gather-dispatch-context's rebuild when inputs are unchanged`

- [x] 2. Point every caller at the new `<output-path>` argument and report the reuse/rebuild line

  Three call sites currently redirect the script's stdout to
  `<worktree>/.superpowers/sdd/dispatch-context.md`. Change each to pass that same path as the 5th
  argument instead, and surface the script's stderr line (`bundle unchanged — reusing …` or
  `bundle rebuilt — …`) as part of that stage's own reporting — this is the ticket's "must be
  observable" requirement, satisfied by relaying the script's own message rather than inventing a
  second one.

  **`skills/flow/implement.md`, section 4 ("Execute (SDD + TDD)")** — the
  `gather-dispatch-context.sh <worktree> <changeRoot> <name> <principles-path> >
  <worktree>/.superpowers/sdd/dispatch-context.md` block becomes
  `gather-dispatch-context.sh <worktree> <changeRoot> <name> <principles-path>
  <worktree>/.superpowers/sdd/dispatch-context.md`, keeping the `mkdir -p` line before it. Add one
  sentence after the existing "Confirm the bundle was actually written" sentence: report the
  script's stderr line for this stage.

  **`skills/flow/review-panel.md`, the "Rebuild the dispatch context bundle at the start of this
  stage too" block** — same shape: drop the `>` redirection, add `<output-path>` as a 5th argument.

  **`skills/flow/review-panel.md`, "Rebuild the dispatch context bundle before dispatching the fix
  subagent"** — same shape again, at the panel-fix-round call site.

  Leave every other line in all three files untouched — in particular, `review-panel.md`'s own
  "never reused from `skills/flow/implement.md`'s run" sentence stays exactly as written: the
  script is still invoked at all three sites on every run, so nothing about *when* it is called
  changes, only whether that call's own effect on disk is a write or a no-op.

  Verify: `grep -rn 'gather-dispatch-context.sh' skills/flow/*.md` shows all three call sites
  passing five arguments with no `>` redirection remaining, and `scripts/check-references.sh`
  (or the project's configured equivalent, if declared) is clean.

**Build:** green
**Files:** `skills/flow/implement.md`, `skills/flow/review-panel.md`
**Tests:** none — these are prose call sites; task 1's harness is the regression surface for the
script's own contract.
**Regression:** reverting this task returns the three call sites to `>` redirection against a
script that (after task 1) requires a 5th argument, so every dispatching stage would fail its usage
check (exit 2) on the very next `/flow` run — loud, not silent.
**Baseline:** before=0 after=0 — no test file this task adds or changes.
**Commit:** `docs(flow): pass gather-dispatch-context its output path instead of redirecting stdout`
