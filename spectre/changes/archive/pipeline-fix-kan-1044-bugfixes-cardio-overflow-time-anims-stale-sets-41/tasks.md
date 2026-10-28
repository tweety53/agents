# pipeline-fix-kan-1044-bugfixes-cardio-overflow-time-anims-stale-sets-41

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

**Goal:** no flow step runs a load reproduction on the operator's machine.

**Architecture:** rewrite three paragraphs in place; tighten the `proofruns` dispatch-paragraph guard.

**Spec:** `design.md` of this change.

## Global constraints

- Pipeline-fix fixer: one commit per task with its `**Commit:**` subject, no trailer; never write `spectre/changes/`.
- Do NOT run any load reproduction, multi-worker suite repeat or concurrent app/browser boot while doing this task.
- Gradle/Go test runs use Bash `timeout: 600000`.
- No new inline suppressions; no weakening of lint config.

- [x] 1. Forbid load reproductions in flake handling
  - [x] **Step 1 — guard first.** In `stats/internal/guard/dispatchparagraphs.go`, `proofruns` entry: replace the
    required phrases with `"never a load reproduction"`, `"single worker"`, `"at most 5 runs"`,
    `"shows the check failing with the mechanism removed"`. In `check_dispatch_paragraphs_test.go` add a case
    (next free number after the highest existing) whose `implement.md` carries the current PROOF RUNS block
    verbatim ("10–15 runs by default …") and asserts exit 1 naming `never a load reproduction`. Update any
    fixture block that carries the old PROOF RUNS text so passing cases still pass. Run
    `cd stats && go test ./internal/guard/ -run DispatchParagraphs` — the new case passes, and the real-tree
    check fails until Step 2.
  - [x] **Step 2 — `skills/flow/implement.md` PROOF RUNS block**, replace with:
    > **PROOF RUNS:** A flake fix is proven by forcing its cause in one isolated run — a delay or fault
    > injected at the seam, single worker, the test alone — never a load reproduction: no multi-worker
    > suite repeats, no `--repeat-each` or loop of a suite, no concurrent browser or app-boot probes, no
    > "reproduce under full-suite load". Where a repeat is still needed it runs the test alone on a
    > single worker, at most 5 runs. A flake seen only under parallel load is diagnosed from the code and
    > the failing run's saved artifacts; where its cause cannot be forced, the fix lands on that
    > code-reading evidence and the commit body says so. Your report states the forced condition and the
    > run count. This never replaces the proof `~/.claude/rules/fix-determinism-at-the-source.md`
    > requires: a determinism fix still shows the check failing with the mechanism removed.
    Also line ~728: keep as is (it cites the paragraph).
  - [x] **Step 3 — `rules/fix-determinism-at-the-source.mdc` § "Read the flake before touching anything".**
    First paragraph: replace "Run the test until you have seen it fail, and name what actually varied" with
    "Read the failing run's saved artifacts and the code, and name what actually varied"; replace "A flake
    you have not reproduced is a guess" with "A cause you have not forced is a guess". Second paragraph,
    replace whole with: "Make it fail by forcing that cause in one isolated, single-worker run — a delay or
    fault injected at the suspected seam, a pinned clock moved, a throttled connection — never by a load
    reproduction: no parallel load, no CPU cap, no multi-worker or looped suite repeats, no concurrent app
    or browser boots; they exhaust the operator's machine (kan-1044). The same forced condition is the one
    the fix and the mechanism-removed run below are proven under. Where the cause cannot be forced, fix it
    on the code-reading evidence and say so." Leave `<!-- core -->` and "never widen" text unchanged.
  - [x] **Step 4 — `skills/flow/verify-fix-loop.md` item 2.** Delete the paragraph "**A task fixing a test
    that passes alone and fails only in the full suite closes after two more full runs.** … never left to the
    next round." In the "**An observed failure is never closed as "not reproduced".**" paragraph replace
    "reproduces it under the conditions of the run that failed — the same concurrent load and the full-suite
    ordering, never the test alone or classes run one after another — and keeps" with "traces it from the
    code and the failing run's saved artifacts, never by a load reproduction (PROOF RUNS,
    `skills/flow/implement.md`), and keeps"; replace "One still unreproduced after runs sized as that
    paragraph sizes them is fixed anyway" with "One whose cause cannot be forced is fixed anyway"; replace
    "its body states the measured rate — runs and failures — and each forced reproduction" with "its body
    states each forced reproduction, or that the fix rests on code-reading evidence".
  - [x] **Step 5.** `grep -rnE "twice more|full-suite load|concurrent load|10.15 runs" rules skills` returns
    only `visual-verify-verifier.md:236` (unrelated measurement). Run `cd stats && go test ./internal/guard/ -run DispatchParagraphs`
    and the repo's `## lint` lines; all green.
  - [x] **Step 6 — memory.** Append to "How to apply" in
    `/Users/tweety53/.claude/projects/-Users-tweety53-Projects-gymie/memory/no-load-testing-on-laptop.md`:
    "Canonical: `rules/fix-determinism-at-the-source.mdc` § Read the flake before touching anything, and
    PROOF RUNS in `skills/flow/implement.md`." (outside the repo; not committed)

**Files:** `stats/internal/guard/dispatchparagraphs.go`, `stats/internal/guard/check_dispatch_paragraphs_test.go`, `skills/flow/implement.md`, `rules/fix-determinism-at-the-source.mdc`, `skills/flow/verify-fix-loop.md`
**Tests:** `go test ./internal/guard/ -run DispatchParagraphs` (new case: old PROOF RUNS text exits 1 naming `never a load reproduction`)
**Build:** green

**Commit:** `fix(fix-determinism): forbid load reproductions in flake proof`
