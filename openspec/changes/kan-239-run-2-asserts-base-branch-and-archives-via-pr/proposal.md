## Why

`/myflow-finish` run 2 archives "on the base branch in the main checkout", but nothing asserts that
the main checkout is on the base branch, and nothing says the archive push must not go to `main`
directly. Five failures across KAN-39, KAN-197, KAN-200, KAN-201 and KAN-236 all follow from those
two omissions — including two recorded direct pushes to `main` (KAN-236's `332e593` and KAN-102's
`7796750`, the latter today) and two self-review reports silently stranded on an already-merged
archive branch.

Every one of them is a case of the contract being followed literally and still producing the wrong
result, which is why the fix has to be structural rather than a clearer sentence.

## What Changes

- **Run 2 positions the main checkout before it archives.** A new guard,
  `prepare-archive-branch.sh <main-checkout> <base> <archive-branch>`, puts the checkout on
  `chore/archive-<name>` cut from a `BASE` fast-forwarded to `origin/<base>`. Clean and off-base
  switches; dirty and off-base refuses, naming the branch found and the branch required.
- **BREAKING (contract):** the step-3 clause *"when finish is invoked with a non-base branch checked
  out, it commits there, merges into the base branch, and pushes that"* is **deleted**. It is the
  clause that would have merged KAN-200's unmerged work into `main`.
- **The archive no longer reaches the base branch by a direct push.** Run 2 commits the archive on
  `chore/archive-<name>` and, as its final step, pushes that branch and opens a pull request —
  matching this repository's own precedent (KAN-201, KAN-209) and the global
  no-direct-pushes-to-`main` rule, which the contract previously contradicted.
- **The self-review report rides in the same pull request.** It is committed onto the archive branch
  before the push, so there is no window in which the archive PR is merged while the report is still
  unwritten — the timing race that stranded KAN-197's and KAN-200's reports is removed rather than
  guarded against.
- **Run 2 restores the main checkout to `BASE`** when it finishes.
- **`gather-self-review-context.sh` accepts an optional fourth positional `<repo-root>`**, validated
  the same way the archived path already is. Omitted, the derived `--git-common-dir` anchor is
  unchanged, so every existing call site keeps working.
- **One new stage key**, `finish.push-archive`; `finish.sync-archive`'s prose name is reworded to
  name the positioning it now includes.
- The **Temporary artifacts registry** gains a row for the archive branch, stating honestly that
  nothing in this pipeline removes it.

Run 2 stays a **single terminal run**. This change adds no third run, no new pipeline state, and no
new state-file field.

## Capabilities

### New Capabilities

*(none — the new guard's behaviour is stated as requirements on the existing finish capability,
where every other run-2 step already lives.)*

### Modified Capabilities

- `myflow-finish-cleanup`: run 2 gains a positioning step before the archive move; the archive is
  committed on an archive branch rather than the base branch, with the non-base-branch merge clause
  removed; and a new final step pushes that branch and opens its pull request.
- `myflow-self-review`: the report is committed on the archive branch rather than "the base branch
  in the main checkout", and is pushed by run 2's push step rather than by self-review itself;
  `gather-self-review-context.sh` accepts an optional explicit repository root.

## Impact

- **New:** `scripts/prepare-archive-branch.sh`, `scripts/test-prepare-archive-branch.sh`, and guard
  symlinks under `skills/myflow-finish/scripts/` and `skills/myflow-fast/scripts/`.
- **Modified:** `scripts/gather-self-review-context.sh` and `scripts/test-gather-self-review-context.sh`;
  `skills/myflow-contracts/finish-contract.md`; `skills/myflow-finish/SKILL.md`;
  `skills/myflow-contracts/pipeline.md` (Git boundaries, Temporary artifacts registry); `README.md`
  (Level 1 stage table); `stats/internal/stages/names.go`; `.myflow/project.md` (`## test`).
- **Unaffected:** run 1, `/myflow-do`, `/myflow-start`, `/myflow-status`, the state file's shape, and
  `check-finish-preflight.sh`'s RUN1/RUN2/REFUSE verdict.
