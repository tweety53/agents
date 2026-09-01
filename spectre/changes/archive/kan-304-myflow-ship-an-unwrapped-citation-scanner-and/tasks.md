# kan-304-myflow-ship-an-unwrapped-citation-scanner-and

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.

Three layers on one branch: the trigger guard (task 1), the contract and the wiring that reads it
(tasks 2–4), and this repository's own declaration (task 5). `design.md` is canonical for every
decision below.

**Baseline, measured before any edit:**

- The Go suite passes in every one of its seventeen packages.
  <!-- measured: cd stats && go test ./... -count=1 @ 3a44bfd (before this change) -->
- `scripts/` carries 32 guards and 44 test harnesses.
  <!-- measured: ls scripts/ | grep -c '^check-'; ls scripts/ | grep -c '^test-' @ 3a44bfd (before this change) -->
- `scripts/check-references.sh` runs clean over this repository's own tree, in about 34s.
  <!-- measured: time scripts/check-references.sh @ 3a44bfd (before this change) -->

**Every task that grows an owned `.md` file runs `scripts/check-contract-budget.sh` and reconciles
its `budgets()` row in the same commit.** That guard is in the project's `## lint` list, so a commit
leaving it red leaves lint red. Raise a row only when the guard actually fails on that file, and
only to the size its own rule gives.

---

## Part 1 — the trigger guard

- [x] 1. Guard and harness for `check-panel-citation-trigger.sh`

  Write `scripts/test-check-panel-citation-trigger.sh` first, then
  `scripts/check-panel-citation-trigger.sh` against it — RED before GREEN.

  The guard takes two arguments, `<worktree> <merge-base>`, and decides whether the change's own
  paths — the union of what's committed since `<merge-base>`, staged, and unstaged (design.md:
  touched-paths-include-index-and-worktree, the same rule `check-base-moved.sh` already applies) —
  include at least one path ending `.md` or `.mdc`. It prints nothing; the exit code is the whole
  answer.

  ```bash verified:authored in-tree for this change
  #!/usr/bin/env bash
  # check-panel-citation-trigger.sh — decide whether the review panel's
  # citation pre-check should run, in a guard instead of prose.
  #
  # Usage: check-panel-citation-trigger.sh <worktree> <merge-base>
  #
  # Prints nothing; the exit code is the whole answer.
  #   0  the change's own paths include at least one path ending .md or .mdc
  #   1  none do
  #   2  usage error: a missing argument, <worktree> not a directory or not a
  #      git worktree, <merge-base> not resolving, or a git invocation failed
  set -euo pipefail

  WORKTREE="${1:-}"
  MERGE_BASE="${2:-}"

  if [ -z "$WORKTREE" ] || [ -z "$MERGE_BASE" ]; then
    echo "check-panel-citation-trigger: usage: check-panel-citation-trigger.sh <worktree> <merge-base>" >&2
    exit 2
  fi

  if [ ! -d "$WORKTREE" ]; then
    echo "check-panel-citation-trigger: $WORKTREE is not a directory — cannot determine anything" >&2
    exit 2
  fi

  if ! git -C "$WORKTREE" rev-parse --git-dir >/dev/null 2>&1; then
    echo "check-panel-citation-trigger: $WORKTREE is not a git worktree — cannot determine anything" >&2
    exit 2
  fi

  if ! git -C "$WORKTREE" rev-parse --verify --end-of-options "${MERGE_BASE}^{commit}" >/dev/null 2>&1; then
    echo "check-panel-citation-trigger: merge base '$MERGE_BASE' does not resolve in $WORKTREE" >&2
    exit 2
  fi

  COMMITTED="$(git -C "$WORKTREE" diff --name-only --end-of-options "${MERGE_BASE}..HEAD" 2>/dev/null)" || {
    echo "check-panel-citation-trigger: cannot list this change's committed paths in $WORKTREE" >&2
    exit 2
  }
  STAGED="$(git -C "$WORKTREE" diff --name-only --cached 2>/dev/null)" || {
    echo "check-panel-citation-trigger: cannot list this change's staged paths in $WORKTREE" >&2
    exit 2
  }
  UNSTAGED="$(git -C "$WORKTREE" diff --name-only 2>/dev/null)" || {
    echo "check-panel-citation-trigger: cannot list this change's unstaged paths in $WORKTREE" >&2
    exit 2
  }

  MATCH="$(printf '%s\n%s\n%s\n' "$COMMITTED" "$STAGED" "$UNSTAGED" | awk 'NF && ($0 ~ /\.mdc?$/) { print; exit }')"

  if [ -n "$MATCH" ]; then
    exit 0
  fi
  exit 1
  ```

  The harness builds throwaway git repositories under a sandboxed `TMPDIR`, removed by an `EXIT`
  trap — the same shape `test-check-base-moved.sh` uses, sourcing
  `scripts/lib/test-git-shim.sh` if it fits this guard's setup. Cases, each **proved by mutation**
  (change the `.mdc?$` pattern to `.never-matches$`, confirm case 2 below flips from pass to fail,
  then revert):

  1. A committed `.md` change since the merge base → exit 0.
  2. A committed `.mdc` change since the merge base → exit 0.
  3. Only non-Markdown committed changes → exit 1.
  4. A staged-only `.md` change, nothing committed → exit 0.
  5. An unstaged-only `.md` change, nothing committed or staged → exit 0.
  6. No changes at all (worktree at the merge base, clean) → exit 1.
  7. Missing arguments → exit 2.
  8. `<worktree>` not a directory → exit 2.
  9. `<worktree>` not a git repository → exit 2.
  10. `<merge-base>` not resolving in the worktree → exit 2.

**Files:** `scripts/check-panel-citation-trigger.sh`, `scripts/test-check-panel-citation-trigger.sh`
**Tests:** `scripts/test-check-panel-citation-trigger.sh`
**Regression:** reverting this commit removes the only mechanical answer to "should the panel's
citation pre-check run", leaving `flow.review-panel` with no testable trigger to call.
**Baseline:** before=44 after=45 harnesses in `scripts/`
<!-- predicted: ls scripts/ | grep -c '^test-' after task 1 -->
**Commit:** `feat(scripts): guard the review panel citation-check trigger`
**Build:** green

## Part 2 — the contract

- [x] 2. The `## review panel citation check` key in `project-configuration.md`

  Add the key to `skills/flow-contracts/project-configuration.md`'s key table, worded like the
  `## stop` row: **optional**, a single fenced command block and nothing else in the section, run
  by `skills/flow/review-panel.md`'s own procedure rather than a dedicated parsing guard — the
  command's *shape* needs no validation beyond "a fenced block exists", so this key gets no
  `check-*.sh` of its own, unlike `## workspace isolation` and `## visual verification`'s
  multi-row tables.

  State plainly: absent key means the whole pre-panel step is skipped, exactly like every other
  optional key; the command runs from the apply worktree only when
  `check-panel-citation-trigger.sh` exits 0 (**Guard resolution**,
  `skills/flow-contracts/pipeline.md`, is how `review-panel.md` names it); its combined
  stdout+stderr is captured verbatim and never gates dispatch — the command's exit code is read by
  nobody.

  Do not restate the wiring procedure — task 4 writes it and it is canonical in
  `skills/flow/review-panel.md`.

  Run `scripts/check-contract-budget.sh` and reconcile this file's row in this commit.

**Files:** `skills/flow-contracts/project-configuration.md`
**Tests:** **none** added — this task adds prose to a contract file; the guards that verify prose in
this repository are `scripts/check-references.sh`, `scripts/check-vocabulary.sh`,
`scripts/check-contract-budget.sh` and `scripts/check-markdown-integrity.py`
**Regression:** reverting this commit leaves `review-panel.md` reading a key no contract defines,
so a project author has nothing to write the section against.
**Baseline:** before=0 after=0 tests added
<!-- predicted: no test file is added by this task -->
**Commit:** `docs(flow-contracts): define the review panel citation check key`
**Build:** red
**Squash-with:** Task 4

## Part 3 — the wiring

- [x] 3. Install the guard everywhere a guard is installed

  Add `check-panel-citation-trigger.sh` to `skills/flow/scripts/` (the symlink farm
  `setup.sh` installs as one unit — no per-guard edit needed there, same as `kan-171`'s task 3
  measured) and to the union `skills/flow/SKILL.md`'s **Check guard presence** paragraph
  enumerates, inserted alphabetically between `check-panel-diff-size.sh`'s predecessor and itself.

  Run `scripts/check-guard-symlinks.sh` and `scripts/test-setup.sh`.

**Files:** `skills/flow/SKILL.md`, `skills/flow/scripts/check-panel-citation-trigger.sh`
**Tests:** **none** added — this task installs an existing guard rather than adding a new one;
`scripts/check-guard-symlinks.sh` and `scripts/test-setup.sh` verify the installation but neither
is modified by this task's own commit
**Regression:** reverting this commit leaves the guard present in `scripts/` but absent from
`skills/flow/scripts/`, where **Guard resolution** (`skills/flow-contracts/pipeline.md`) resolves
it — so `review-panel.md` would fall through to the hand-run fallback on every run.
**Baseline:** before=0 after=0 tests added
<!-- predicted: test-setup.sh gains assertions, not a new file -->
**Commit:** `chore(setup): install the panel citation-check trigger guard`
**Build:** red
**Squash-with:** Task 4

- [x] 4. Wire the citation pre-check into `flow.review-panel`

  Add the step to `skills/flow/review-panel.md`, inside the existing `flow.review-panel` stage,
  immediately before "Rebuild the dispatch context bundle" — no new stage key, per design.md's
  `no-new-stage-key`:

  1. Run `check-panel-citation-trigger.sh <worktree> <merge-base>`.
  2. On exit 0: read `.flow/project.md`'s `## review panel citation check` key (**Guard
     resolution**, `skills/flow-contracts/pipeline.md`). If declared, run its command from the
     apply worktree and capture combined stdout+stderr verbatim to
     `<worktree>/.superpowers/sdd/citation-check.md`.
  3. On exit 1, or the key absent, skip silently — no file written, no paragraph added.
  4. Every panel slot's dispatch prompt gains a new **CITATION CHECK** paragraph, present only
     when step 2 wrote a file, naming its absolute path alongside `final-review.diff`:

     > **CITATION CHECK:** `<abs-path>/.superpowers/sdd/citation-check.md` carries this project's
     > pre-panel citation scan, captured before your dispatch. It is informational — its own exit
     > code was not gating — but a stale citation it reports is worth raising as your own finding
     > if it sits in this diff's blast radius.

  State the never-blocks rule in full: the configured command's exit code is read by nobody — a
  non-zero exit (findings present) still just writes the file and adds the paragraph. The real
  gate stays `flow.verify`'s existing `## lint` run, unchanged by this task.

  Run `scripts/check-contract-budget.sh` and reconcile this file's row in this commit.

**Files:** `skills/flow/review-panel.md`, `scripts/check-contract-budget.sh`
**Tests:** **none** added — this task adds prose to a skill file, verified by
`scripts/check-references.sh` and `scripts/check-contract-budget.sh`, neither of which this task's
own commit modifies
**Regression:** reverting this commit removes the only place the pipeline runs the citation check
before the panel, leaving tasks 1–3's guard and contract declared and installed but read by
nothing.
**Baseline:** before=0 after=0 tests added
<!-- predicted: no test file is added by this task -->
**Commit:** `feat(flow): run the citation check before the review panel`
**Build:** green

## Part 4 — this repository's own declaration

- [x] 5. Declare `## review panel citation check` for `agents`

  Add the section to `.flow/project.md`:

  ```markdown verified:written against this project's own tree
  ## review panel citation check

  \`\`\`bash
  scripts/check-references.sh
  \`\`\`
  ```

  This reuses the existing scan — no new algorithm — so every review panel dispatched on a change
  that touches this repository's own Markdown now sees whether it left a stale citation behind,
  before any reviewer starts reading the diff.

  Run `scripts/check-contract-budget.sh` and reconcile this file's row in this commit.

**Files:** `.flow/project.md`, `scripts/check-contract-budget.sh`
**Tests:** **none** added — declaring the key exercises tasks 1 and 4's own code paths on the next
`/flow` run in this repository; there is no standalone guard for this key's shape (task 2)
**Regression:** reverting this commit leaves this repository — the one KAN-304 was filed
about — the one project the new pre-check never runs on.
**Baseline:** before=0 after=0 tests added
<!-- predicted: no test file is added by this task -->
**Commit:** `chore(flow-config): declare the review panel citation check for agents`
**Build:** green
