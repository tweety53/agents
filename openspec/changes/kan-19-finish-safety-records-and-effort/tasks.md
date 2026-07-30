# KAN-19 — finish safety, session records, and planning effort

> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

**Goal:** Stop `/myflow-finish` from archiving and force-deleting a branch that was never committed,
preserve the session records that today die with the worktree, account for the proposal artifact
source, and let the operator size `/myflow-start`'s reasoning once per change.

**Architecture:** Two new repository-local shell scripts carry the logic that must be testable — the
run-1/run-2 verdict and the record copy — each with an assertion harness beside it, matching this
repository's existing `check-*` / `test-*` pairs. Everything else is contract and skill text:
`pipeline.md` stays canonical and points at the scripts rather than restating them.

**Tech Stack:** Bash 3.2 (macOS system bash), `git` plumbing, Markdown contracts, OpenSpec.

## Global Constraints

- **No suppression markers, ever.** This repository's Lint Fix Priority rule forbids adding a
  suppression directive or weakening a guard's configuration to make a check pass. Fix the line.
- **No auto-fix command exists** in this repository — `.myflow/project.md` `## lint` states this
  explicitly. The auto-fix-first step is inapplicable here, not skipped.
- **Lint is three commands:** `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
  `scripts/check-plan-provenance.sh`. All three must exit 0 before any task is claimed done.
- **`pipeline.md` is canonical.** A procedure that gates a destructive operation lives there once.
  Skills point at it; they never restate it. `skills/myflow-finish/SKILL.md` carries this as a
  standing guardrail.
- **Bash 3.2 only.** No `declare -A`, no `${var^^}`, no `mapfile`. macOS ships bash 3.2 and this
  repository has no dependency management.
- **`check-references.sh` validates backticked `.md`/`.mdc` paths and the sections they name.** Any
  section named in prose must exist with that exact heading, or the guard fails.
- **New scripts are repository-local.** `setup.sh` installs skills, commands and rules — not
  `scripts/`. The contract text that points at a script is installed; the script is not.

---

## 1. The finish preflight check

### Task 1: `check-finish-preflight.sh` and its harness

**Files:**
- Create: `scripts/check-finish-preflight.sh`
- Create: `scripts/test-check-finish-preflight.sh`

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: an executable at `scripts/check-finish-preflight.sh` taking three positional arguments —
  `<worktree-path> <base-ref> <recorded-merge-base|->` — printing one line to stdout of the form
  `RUN1: <reason>`, `RUN2: <reason>` or `REFUSE: <reason>`, exiting `0` on any verdict and `2` when
  it cannot determine anything. Task 2 wires this into the contract and the skill.

- [x] **Step 1: Write the harness with every case, before the script exists**

Create `scripts/test-check-finish-preflight.sh`. The `new_repo` helper builds a throwaway repository
with a base branch and a change branch; each case then shapes it and asserts one verdict.

```bash unverified:confirm `git init -b` is accepted by the git on this machine; fall back to `git init` + `git symbolic-ref HEAD refs/heads/main` if not
#!/usr/bin/env bash
# Assertion harness for check-finish-preflight.sh. Builds throwaway git
# repositories under a sandboxed TMPDIR and asserts the guard's verdict and
# exit status. Never touches the real repository tree.
#
# READ THIS BEFORE ADDING OR "FIXING" A CASE. Assert against the stated
# contract in skills/myflow-contracts/pipeline.md, never against observed
# output. test-check-plan-provenance.sh's header records that suite encoding
# the guard's own defects as its specification more than once, which then made
# each defect look verified.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-finish-preflight.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

# run_guard <worktree> <base-ref> <recorded-merge-base> -> sets RC and OUT
run_guard() {
  set +e
  OUT="$("$GUARD" "$1" "$2" "$3" 2>&1)"
  RC=$?
  set -e
}

# new_repo -> sets REPO, BASE_REF, MERGE_BASE
# A repository on `main` with one commit, plus a branch `openspec/demo`
# checked out at that same commit. MERGE_BASE is that commit.
new_repo() {
  REPO="$(mktemp -d "${TMPDIR:-/tmp}/finish-preflight-test.XXXXXX")"
  git -C "$REPO" init -q -b main
  git -C "$REPO" config user.email test@example.invalid
  git -C "$REPO" config user.name "Test"
  echo base > "$REPO/file.txt"
  git -C "$REPO" add file.txt
  git -C "$REPO" commit -qm "base"
  MERGE_BASE="$(git -C "$REPO" rev-parse HEAD)"
  BASE_REF=main
  git -C "$REPO" checkout -q -b openspec/demo
}
```

- [x] **Step 2: Add the assertion cases to the same file**

Append to `scripts/test-check-finish-preflight.sh`. Case 1 is the KAN-14 shape and is the reason this
script exists.

> **Fix-wave note (amends this step).** The shipped harness carries more cases than the block below,
> and `scripts/test-check-finish-preflight.sh` on disk is authoritative over this illustrative
> listing. The block was left as first written rather than re-pasted, because a second copy of a file
> that keeps changing is what went stale here in the first place. What the two fix waves added, and
> why each case exists rather than being a variation on an existing one:
>
> - **1b, 1c** — the zero-commit shape with an unresolvable base ref, and with a completely clean
>   tree. Case 1 alone passes a guard with signal (b) deleted outright, because its staged entry makes
>   signal (d) refuse anyway; only a clean zero-commit tree isolates signal (b), which is the one
>   thing between that shape and run 2's `git worktree remove --force`.
> - **7b** — a base ref that does not resolve is `REFUSE`, not `RUN1`. The ancestor test fails
>   identically for "not merged" and "no such ref".
> - **8b** — `git status` failing on the merged-and-clean shape is exit 2, not a verdict. The
>   original piped `status` straight into `wc`, so a failed git left the pipeline's status to `wc`
>   and the count read `0` — falling through to `RUN2`, the destructive verdict.
> - **Case 5's reason text** (second wave) — asserting the `REFUSE` prefix alone could not tell
>   "the `-` sentinel is handled" from "`-^{commit}` failed to resolve by accident", since both
>   produce the same prefix. Deleting the sentinel branch left every case green; it now fails two.
> - The exit-status assertions beside the verdict assertions, and the sandbox cleanup trap using an
>   **indexed array** so a `TMPDIR` containing a space cannot word-split into `rm -rf` fragments.

```bash unverified:confirm `git status --porcelain` is non-empty for a staged-only change on this git version
# 1. Zero-commit branch with staged work: RUN1, never RUN2.
#    HEAD is still the merge base, so the ancestor test alone says "merged".
new_repo
echo staged > "$REPO/new.txt"
git -C "$REPO" add new.txt
run_guard "$REPO" "$BASE_REF" "$MERGE_BASE"
case "$OUT" in
  RUN1*) pass "zero-commit branch with staged work -> RUN1" ;;
  *) fail "zero-commit branch: expected RUN1, got rc=$RC out=$OUT" ;;
esac

# 2. Genuinely merged, clean tree: RUN2.
new_repo
echo work > "$REPO/new.txt"
git -C "$REPO" add new.txt
git -C "$REPO" commit -qm "work"
git -C "$REPO" checkout -q main
git -C "$REPO" merge -q --no-ff -m "merge" openspec/demo
git -C "$REPO" checkout -q openspec/demo
run_guard "$REPO" "$BASE_REF" "$MERGE_BASE"
case "$OUT" in
  RUN2*) pass "merged with clean tree -> RUN2" ;;
  *) fail "merged clean: expected RUN2, got rc=$RC out=$OUT" ;;
esac

# 3. Merged by ancestry but the worktree is dirty: REFUSE.
new_repo
echo work > "$REPO/new.txt"
git -C "$REPO" add new.txt
git -C "$REPO" commit -qm "work"
git -C "$REPO" checkout -q main
git -C "$REPO" merge -q --no-ff -m "merge" openspec/demo
git -C "$REPO" checkout -q openspec/demo
echo dirty > "$REPO/file.txt"
run_guard "$REPO" "$BASE_REF" "$MERGE_BASE"
case "$OUT" in
  REFUSE*) pass "merged but dirty -> REFUSE" ;;
  *) fail "merged dirty: expected REFUSE, got rc=$RC out=$OUT" ;;
esac

# 4. Unmerged and dirty: RUN1, NOT a refusal. This is the ordinary
#    IN_PROGRESS state and must not prompt the operator on every finish.
new_repo
echo work > "$REPO/new.txt"
git -C "$REPO" add new.txt
git -C "$REPO" commit -qm "work"
echo dirty > "$REPO/file.txt"
run_guard "$REPO" "$BASE_REF" "$MERGE_BASE"
case "$OUT" in
  RUN1*) pass "unmerged and dirty -> RUN1" ;;
  *) fail "unmerged dirty: expected RUN1, got rc=$RC out=$OUT" ;;
esac

# 5. No recorded merge base: REFUSE, never an inferred verdict.
new_repo
echo work > "$REPO/new.txt"
git -C "$REPO" add new.txt
git -C "$REPO" commit -qm "work"
git -C "$REPO" checkout -q main
git -C "$REPO" merge -q --no-ff -m "merge" openspec/demo
git -C "$REPO" checkout -q openspec/demo
run_guard "$REPO" "$BASE_REF" -
case "$OUT" in
  REFUSE*) pass "absent merge base -> REFUSE" ;;
  *) fail "absent merge base: expected REFUSE, got rc=$RC out=$OUT" ;;
esac

# 6. Abbreviated recorded sha, genuinely merged: RUN2. An abbreviation must
#    not read as a difference from a full HEAD.
new_repo
SHORT="$(git -C "$REPO" rev-parse --short "$MERGE_BASE")"
echo work > "$REPO/new.txt"
git -C "$REPO" add new.txt
git -C "$REPO" commit -qm "work"
git -C "$REPO" checkout -q main
git -C "$REPO" merge -q --no-ff -m "merge" openspec/demo
git -C "$REPO" checkout -q openspec/demo
run_guard "$REPO" "$BASE_REF" "$SHORT"
case "$OUT" in
  RUN2*) pass "abbreviated merge base -> RUN2" ;;
  *) fail "abbreviated merge base: expected RUN2, got rc=$RC out=$OUT" ;;
esac

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'check-finish-preflight: all cases pass\n'
```

- [x] **Step 3: Make the harness executable and run it to watch it fail**

```bash verified:the same two-command shape opens scripts/test-check-plan-provenance.sh's own run
chmod +x scripts/test-check-finish-preflight.sh
./scripts/test-check-finish-preflight.sh
```

Expected: every case fails, because `scripts/check-finish-preflight.sh` does not exist yet. If a
case *passes* here, the harness is not invoking the guard — fix the harness before continuing.

- [x] **Step 4: Write the guard**

Create `scripts/check-finish-preflight.sh`. The check order is the fix: signal `b` before signal `c`,
signal `d` after.

> **Fix-wave note (amends this step).** `scripts/check-finish-preflight.sh` on disk is authoritative
> over the block below, which is the pre-fix-wave draft and is kept only to show the signal order this
> task set out to establish. Three things in the shipped guard are not in it, each found by the review
> panel and each proved by a harness case first:
>
> - **`--end-of-options` on every `rev-parse` and on `merge-base`.** Both the base ref and the
>   recorded merge base arrive from a state file. A value beginning with `-` was parsed as a git
>   option instead of read as a ref.
> - **The base ref is resolved before the ancestor test**, so an unresolvable base ref is its own
>   `REFUSE` rather than an accidental `RUN1` — the ancestor test cannot distinguish the two.
> - **The dirty count no longer pipes `git status` straight into `wc`.** The status output is captured
>   on its own line and a git failure is exit 2, named. In a pipeline, `wc` and `tr` succeeding make
>   the pipeline's status theirs, so a failed git produced a silent `0` that fell through to `RUN2`.
>   The shipped script's own comment states this; the block below is the shape it warns about.

```bash unverified:confirm `git rev-parse --verify <ref>^{commit}` errors rather than printing on a bad ref, and that `merge-base --is-ancestor` exit 1 means "not an ancestor" on this git version
#!/usr/bin/env bash
# check-finish-preflight.sh — decide whether /myflow-finish should integrate
# (run 1) or archive (run 2), or refuse because it cannot tell.
#
# Usage: check-finish-preflight.sh <worktree> <base-ref> <recorded-merge-base|->
#
# Prints ONE verdict line to stdout:
#   RUN1: <reason>    integrate — the branch has not reached the base branch
#   RUN2: <reason>    archive   — the branch is merged and nothing is outstanding
#   REFUSE: <reason>  stop and ask the operator
#
# Exit 0 whenever a verdict was reached; exit 2 when the tree cannot be read.
# The VERDICT carries the answer, not the exit status: an exit-code-only
# protocol makes "cannot determine" indistinguishable from "violation", which
# is the exact confusion this script exists to remove. Exit 2 keeps the
# meaning this repository's other guards give it.
#
# WHY SIGNAL ORDER IS THE WHOLE FIX. A branch with no commits of its own is an
# ancestor of EVERY branch, so `merge-base --is-ancestor` answers "merged" on a
# branch whose work is staged and never committed — the normal IN_PROGRESS
# state, since /myflow-do may not commit before a PR exists. Run 2 would then
# archive the change and `git worktree remove --force` the worktree holding all
# of it. Comparing HEAD with the merge base RECORDED IN THE STATE FILE catches
# it; counting commits ahead of the base branch does NOT, because a genuinely
# merged branch is also zero ahead once its commit joins the base branch.
#
# Base-branch resolution deliberately stays in pipeline.md's Finish contract
# and is passed in. That resolution already carries a hard-won guard against
# HEAD@{upstream}; a second copy here could drift from it.
set -euo pipefail

WORKTREE="${1:-}"
BASE_REF="${2:-}"
RECORDED="${3:-}"

if [ -z "$WORKTREE" ] || [ -z "$BASE_REF" ] || [ -z "$RECORDED" ]; then
  echo "usage: check-finish-preflight.sh <worktree> <base-ref> <recorded-merge-base|->" >&2
  exit 2
fi

if [ ! -d "$WORKTREE" ]; then
  echo "check-finish-preflight: $WORKTREE is not a directory — cannot determine anything" >&2
  exit 2
fi

if ! git -C "$WORKTREE" rev-parse --git-dir >/dev/null 2>&1; then
  echo "check-finish-preflight: $WORKTREE is not a git worktree — cannot determine anything" >&2
  exit 2
fi

# (a) No recorded merge base: an honest unknown, never an inferred verdict.
if [ "$RECORDED" = "-" ]; then
  echo "REFUSE: no merge base recorded for $WORKTREE — cannot tell an unmerged branch from a merged one"
  exit 0
fi

HEAD_SHA="$(git -C "$WORKTREE" rev-parse --verify HEAD^{commit} 2>/dev/null)" || {
  echo "check-finish-preflight: cannot resolve HEAD in $WORKTREE" >&2
  exit 2
}

RECORDED_SHA="$(git -C "$WORKTREE" rev-parse --verify "${RECORDED}^{commit}" 2>/dev/null)" || {
  echo "REFUSE: recorded merge base '$RECORDED' does not resolve in $WORKTREE"
  exit 0
}

# (b) HEAD is still the merge base: the branch has no commits of its own.
#     This MUST be tested before the ancestor test.
if [ "$HEAD_SHA" = "$RECORDED_SHA" ]; then
  echo "RUN1: HEAD is still the recorded merge base — the branch has no commits of its own"
  exit 0
fi

# (c) The ancestor test.
if ! git -C "$WORKTREE" merge-base --is-ancestor "$HEAD_SHA" "$BASE_REF" 2>/dev/null; then
  echo "RUN1: HEAD is not an ancestor of $BASE_REF — not merged"
  exit 0
fi

# (d) Merged by ancestry, so nothing should be outstanding. Tracked changes
#     and untracked-unignored files both count; ignored files do not, because
#     they are disclosed separately at removal time rather than gating.
DIRTY="$(git -C "$WORKTREE" status --porcelain --untracked-files=normal 2>/dev/null | wc -l | tr -d ' ')"
if [ "$DIRTY" != "0" ]; then
  echo "REFUSE: $BASE_REF contains HEAD, but $WORKTREE has $DIRTY uncommitted entr(y/ies) — a merged change should have nothing left to commit"
  exit 0
fi

echo "RUN2: HEAD is an ancestor of $BASE_REF, differs from the recorded merge base, and the worktree is clean"
exit 0
```

- [x] **Step 5: Make it executable and run the harness to green**

```bash verified:mirrors the invocation .myflow/project.md `## test` already uses for the other harnesses
chmod +x scripts/check-finish-preflight.sh
./scripts/test-check-finish-preflight.sh
```

Expected: `check-finish-preflight: all cases pass`.

- [x] **Step 6: Run the three lint guards**

```bash verified:these are the three commands `.myflow/project.md` `## lint` declares
./scripts/check-vocabulary.sh
./scripts/check-references.sh
./scripts/check-plan-provenance.sh
```

Expected: all three exit 0.

---

## 2. Wire the preflight into the contract and the finish skill

### Task 2: `pipeline.md` and `myflow-finish/SKILL.md` consult the script

**Files:**
- Modify: `skills/myflow-contracts/pipeline.md` — the `## Finish contract` section, under
  `### Run 2 — the branch is merged`
- Modify: `skills/myflow-finish/SKILL.md` — the `## Deciding which run this is` section and run 2
  step 1

**Interfaces:**
- Consumes: `scripts/check-finish-preflight.sh` from Task 1, including its three-argument signature
  and its three verdict words.
- Produces: contract text later tasks reference by section name. Do not rename
  `## Finish contract`, `### Run 1 — the branch is not merged` or `### Run 2 — the branch is merged`
  — `check-references.sh` resolves named sections, and `skills/myflow-finish/SKILL.md` and
  `skills/myflow-status/SKILL.md` point at them.

- [x] **Step 1: Replace the run-decision text in `pipeline.md`**

In `## Finish contract`, the paragraph that currently says the branch's merge status is the only
source of truth stays. Add immediately after it:

```markdown unverified:confirm the surrounding heading levels after insertion — the block must sit under `## Finish contract`, above `### Run 1 — the branch is not merged`
**The merge status is decided by three signals, in this order, and by a script — not by prose.**
`scripts/check-finish-preflight.sh <worktree> <base-ref> <recorded-merge-base|->` prints one verdict:

| Verdict | Meaning |
|---------|---------|
| `RUN1` | integrate — the branch has not reached the base branch |
| `RUN2` | archive — merged, and nothing is outstanding |
| `REFUSE` | stop and ask the operator before anything is archived |

1. **`HEAD` against the merge base recorded in the state file's `worktrees` map.** Equal means the
   branch has no commits of its own → `RUN1`, whatever the ancestor test says.
2. **The ancestor test** — `git merge-base --is-ancestor`, or a PR CLI when one is usable.
3. **The worktree's cleanliness.** Merged by ancestry with uncommitted entries → `REFUSE`.

**Signal 1 precedes signal 2, and that ordering is the point.** A branch with no commits of its own
is an ancestor of every branch, so the ancestor test alone reports *merged* on a branch whose work is
staged and never committed — after which run 2 archives the change and `--force`-removes the worktree
holding all of it.

**Never substitute a commit count.** `git rev-list --count <base>..HEAD` is zero both for a branch
with no commits and for a branch whose commits have joined the base branch, so it cannot separate the
dangerous state from the correct terminal one, and using it would refuse every legitimate archive.

On a `REFUSE`, stop before touching anything, report `HEAD`, the base branch and the uncommitted
count, and ask the operator explicitly. On a multi-repo change, run the script once per `worktrees`
key and proceed to run 2 only when **every** worktree returns `RUN2`.

**When the script is absent** — a harness whose repository does not carry it — perform the same three
signals by hand in the same order and say in the handoff that the check was run manually. The check is
never skipped for want of the script.
```

- [x] **Step 2: Point the finish skill at it, without restating it**

In `skills/myflow-finish/SKILL.md`, `## Deciding which run this is` currently lists the two outcomes
as a plain bullet pair. Replace that pair with:

```markdown unverified:confirm the bullet pair being replaced still reads "not merged -> run 1" / "merged -> run 2" at implementation time
Run `scripts/check-finish-preflight.sh` once per worktree recorded in the state file's `worktrees`
map, per **Finish contract**, and act on the verdict:

- **`RUN1`** → run 1 (integrate)
- **`RUN2`** from every worktree → run 2 (archive and clean up)
- **`REFUSE`** → stop, report what the script reported, and ask the operator before anything else

A PR a human merged on the forge, a colleague's merge, and run 1's own merge are indistinguishable
here, and that is correct — all three mean the same thing.
```

- [x] **Step 3: Add the guardrail**

Append to `skills/myflow-finish/SKILL.md`'s `## Guardrails` list, keeping the existing `- **Never**`
phrasing:

```markdown verified:matches the existing guardrail style in skills/myflow-finish/SKILL.md's Guardrails list
- **Never** decide run 1 versus run 2 from the ancestor test alone, and never from a commit count —
  both answer wrongly on a branch that was never committed.
```

- [x] **Step 4: Run the lint guards**

```bash verified:these are the three commands `.myflow/project.md` `## lint` declares
./scripts/check-vocabulary.sh
./scripts/check-references.sh
./scripts/check-plan-provenance.sh
```

Expected: all three exit 0. `check-references.sh` is the one that matters here — it resolves every
backticked path and named section the new text introduces.

---

## 3. Session-record preservation

### Task 3: `preserve-session-records.sh` and its harness

**Files:**
- Create: `scripts/preserve-session-records.sh`
- Create: `scripts/test-preserve-session-records.sh`

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: an executable taking `<worktree> <change-name> <state-dir>`, copying three sources into
  `docs/superpowers/{ledgers,reviews,artifacts}/` under the worktree, printing one line per source
  (`preserved:` or `skipped:`), and exiting `0` unless it could not write. Task 4 wires it in;
  Task 5 depends on the artifact copy having happened.

- [x] **Step 1: Confirm the ledger's path before writing anything**

The panel record's path is verified: `skills/myflow-do/SKILL.md` writes
`.superpowers/sdd/final-review-panel.md`. The ledger's path is **not** — it is taken from
`openspec/specs/myflow-model-policy/spec.md`'s prose and from the KAN-19 issue text, and no worktree
existed while this plan was written.

```bash unverified:this is the confirmation step itself — run it in the apply worktree and use what it finds
ls -la .superpowers/sdd/tasks/progress.md .superpowers/sdd/final-review-panel.md 2>&1
grep -rn "progress.md" skills/ | head
```

If the ledger sits elsewhere, use the real path and say so in the task report. Do **not** keep the
planned path because the plan named it.

- [x] **Step 2: Write the harness first**

Create `scripts/test-preserve-session-records.sh`.

> **Fix-wave note (amends this step).** `scripts/test-preserve-session-records.sh` on disk is
> authoritative over the block below. The shipped harness adds, each case written to fail first:
>
> - **1b, 1c, 2b** — the copied content actually arrives; the destination is date-stamped as the spec
>   states; and the existing-file search does not adopt another change's record (`*-demo.md` matches
>   `2020-01-01-other-demo.md`).
> - **4, 4b, 4c** — a write that was attempted and could not be made exits non-zero while the other
>   records are still preserved; a destination symlinked out of the worktree is refused; a symlink
>   pointing inside it is followed.
> - **4c-ii** (second wave) — the reported destination is the **resolved** directory. The boundary
>   check validated a resolved path while `find` and `cp` still went through the unresolved argument,
>   so `cp` re-followed the symlink components the check had just cleared. A background process
>   swapping the destination won that race in most trials: the script printed `preserved:` while the
>   content landed outside the worktree.
> - **4c-iii, 4c-iv** (second wave) — a **source** symlinked out of its root is refused, not read, and
>   one pointing inside its root is followed. Only the destination had been checked, which left an
>   arbitrary-file-read: with `progress.md` replaced by a symlink to a file holding a planted
>   credential, the script reported `preserved:` and the credential was in the repository, ready to be
>   committed and pushed by the run-1 path that invokes it.
> - **4d, 4d-ii, 4d-iii** (allowlist widened in the second wave) — a change name outside the allowlist
>   is rejected before any directory is touched, a glob name cannot adopt another change's record, and
>   the name shapes `/myflow-start` really produces are accepted.
> - The cleanup trap uses an **indexed array**, matching `scripts/test-check-finish-preflight.sh`.
>   As a space-separated string it word-split under a `TMPDIR` containing a space and leaked every
>   sandbox it created — 34 of them in the reproduction, against none for the sibling harness.
>   <!-- measured: TMPDIR="<scratch>/space dir/" ./scripts/test-preserve-session-records.sh, pre-fix copy @ branch openspec/kan-19-finish-safety-records-and-effort -->


```bash unverified:confirm the `set -- ` reset pattern behaves under `set -u` on bash 3.2
#!/usr/bin/env bash
# Assertion harness for preserve-session-records.sh. Builds sandboxed source
# and destination trees; never touches the real repository tree.
#
# Assert against the contract in skills/myflow-contracts/pipeline.md, never
# against observed output — see test-check-plan-provenance.sh's header.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCRIPT="$SCRIPT_DIR/preserve-session-records.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

# new_tree -> sets WT and STATE_DIR, with all three sources present
new_tree() {
  WT="$(mktemp -d "${TMPDIR:-/tmp}/preserve-test-wt.XXXXXX")"
  STATE_DIR="$(mktemp -d "${TMPDIR:-/tmp}/preserve-test-state.XXXXXX")"
  mkdir -p "$WT/.superpowers/sdd/tasks"
  printf 'ledger body\n' > "$WT/.superpowers/sdd/tasks/progress.md"
  printf 'panel body\n' > "$WT/.superpowers/sdd/final-review-panel.md"
  printf '<p>artifact</p>\n' > "$STATE_DIR/demo-proposal-artifact.html"
}

run_it() {
  set +e
  OUT="$("$SCRIPT" "$WT" demo "$STATE_DIR" 2>&1)"
  RC=$?
  set -e
}

# 1. First copy places all three records under docs/superpowers/.
new_tree
run_it
[ "$RC" -eq 0 ] || fail "first copy: rc=$RC out=$OUT"
[ -n "$(find "$WT/docs/superpowers/ledgers" -name '*-demo.md' 2>/dev/null)" ] \
  && pass "ledger preserved" || fail "ledger missing: $OUT"
[ -n "$(find "$WT/docs/superpowers/reviews" -name '*-demo-panel.md' 2>/dev/null)" ] \
  && pass "panel preserved" || fail "panel missing: $OUT"
[ -n "$(find "$WT/docs/superpowers/artifacts" -name '*-demo.html' 2>/dev/null)" ] \
  && pass "artifact preserved" || fail "artifact missing: $OUT"

# 2. A re-copy overwrites in place and creates no second dated file.
new_tree
run_it
mkdir -p "$WT/docs/superpowers/ledgers"
LEDGER="$(find "$WT/docs/superpowers/ledgers" -name '*-demo.md' | head -1)"
mv "$LEDGER" "${LEDGER%/*}/1999-01-01-demo.md"
printf 'ledger body v2\n' > "$WT/.superpowers/sdd/tasks/progress.md"
run_it
COUNT="$(find "$WT/docs/superpowers/ledgers" -name '*-demo.md' | wc -l | tr -d ' ')"
[ "$COUNT" = "1" ] && pass "re-copy does not duplicate" || fail "re-copy made $COUNT ledger files"
grep -q 'v2' "$WT/docs/superpowers/ledgers/1999-01-01-demo.md" \
  && pass "re-copy overwrites the existing dated path" \
  || fail "re-copy did not overwrite the existing dated path"

# 3. Each source missing independently is skipped, never fatal.
for missing in ledger panel artifact; do
  new_tree
  case "$missing" in
    ledger) rm "$WT/.superpowers/sdd/tasks/progress.md" ;;
    panel) rm "$WT/.superpowers/sdd/final-review-panel.md" ;;
    artifact) rm "$STATE_DIR/demo-proposal-artifact.html" ;;
  esac
  run_it
  [ "$RC" -eq 0 ] && pass "missing $missing is not fatal" \
    || fail "missing $missing: rc=$RC out=$OUT"
  case "$OUT" in
    *skipped:*) pass "missing $missing is reported" ;;
    *) fail "missing $missing was not reported: $OUT" ;;
  esac
done

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'preserve-session-records: all cases pass\n'
```

- [x] **Step 3: Run the harness to watch it fail**

```bash verified:same invocation shape as the other harnesses in scripts/
chmod +x scripts/test-preserve-session-records.sh
./scripts/test-preserve-session-records.sh
```

Expected: failures on every case — the script does not exist.

- [x] **Step 4: Write the script**

> **Fix-wave note (amends this step).** `scripts/preserve-session-records.sh` on disk is authoritative
> over the block below, which carries none of the path protections the two fix waves added. The
> shipped script differs in four ways, each with a harness case behind it:
>
> - **The change name is checked against an allowlist** — one leading alphanumeric, then letters,
>   digits, `.`, `_`, `-` — before any directory is created. Two hazards, one check: a name containing
>   `/` was blocked only by the accident that `$TODAY-` shares the path component with the start of the
>   name, and a name carrying a glob metacharacter defeated the digit-anchored `find`, so the change
>   name `*` matched and overwrote an unrelated change's preserved ledger.
> - **Every destination directory must resolve inside the worktree.** The three directories under
>   `docs/superpowers/` are ordinary tracked repo paths, editable in any pull request; `mkdir -p` is a
>   no-op on an existing symlink and `cp` follows it.
> - **Every source must resolve inside the root it comes from** (second wave) — the worktree for the
>   two `.superpowers/` sources, the state directory for the artifact. `.superpowers/` being gitignored
>   is no protection: `.gitignore` gates only untracked paths, so a symlink forced in with
>   `git add -f` is on disk after a checkout. A refused source is reported on stderr and non-zero,
>   deliberately not the silent `skipped:` path, which means only "this change has no such record".
> - **The destination path is built from the resolved directory, and `cp` reads the resolved source**
>   (second wave), so the check and the write cannot act on different paths. The residual window —
>   the resolved directory itself being replaced — is stated in the script's header rather than
>   claimed to be closed.
>
> The `preserve` helper therefore takes a source root as well, and the three call sites pass it.
> `find`, `cp` and the reported `preserved:` line all use the resolved destination.

```bash unverified:confirm `date -u +%Y-%m-%d` and BSD `find`'s `-name` behave as used here on macOS
#!/usr/bin/env bash
# preserve-session-records.sh — copy a change's session records out of the
# gitignored worktree and into the repository, so they survive worktree
# removal.
#
# Usage: preserve-session-records.sh <worktree> <change-name> <state-dir>
#
# Prints one line per source: "preserved: <dest>" or "skipped: <src> (absent)".
# A MISSING SOURCE IS NEVER FATAL. A change may legitimately have no panel
# record, and a preservation step able to block an integration would be a worse
# failure than the gap it closes. Exit non-zero only when a copy was attempted
# and could not be written.
#
# The destination date is fixed at the FIRST copy: an existing file for this
# change is reused, so a fix round overwrites in place instead of leaving one
# dated duplicate per round.
set -euo pipefail

WORKTREE="${1:-}"
NAME="${2:-}"
STATE_DIR="${3:-}"

if [ -z "$WORKTREE" ] || [ -z "$NAME" ] || [ -z "$STATE_DIR" ]; then
  echo "usage: preserve-session-records.sh <worktree> <change-name> <state-dir>" >&2
  exit 2
fi

TODAY="$(date -u +%Y-%m-%d)"

# preserve <source> <dest-dir> <suffix>
# <suffix> is what follows the change name, e.g. ".md" or "-panel.md".
preserve() {
  src="$1"; dest_dir="$2"; suffix="$3"
  if [ ! -f "$src" ]; then
    echo "skipped: $src (absent)"
    return 0
  fi
  mkdir -p "$dest_dir"
  existing="$(find "$dest_dir" -maxdepth 1 -name "*-${NAME}${suffix}" 2>/dev/null | head -1)"
  if [ -n "$existing" ]; then
    dest="$existing"
  else
    dest="$dest_dir/$TODAY-$NAME$suffix"
  fi
  cp "$src" "$dest" || {
    echo "preserve-session-records: could not write $dest" >&2
    return 1
  }
  echo "preserved: $dest"
}

RC=0
preserve "$WORKTREE/.superpowers/sdd/tasks/progress.md" \
  "$WORKTREE/docs/superpowers/ledgers" ".md" || RC=1
preserve "$WORKTREE/.superpowers/sdd/final-review-panel.md" \
  "$WORKTREE/docs/superpowers/reviews" "-panel.md" || RC=1
preserve "$STATE_DIR/$NAME-proposal-artifact.html" \
  "$WORKTREE/docs/superpowers/artifacts" ".html" || RC=1

exit "$RC"
```

- [x] **Step 5: Run the harness to green**

```bash verified:same invocation shape as the other harnesses in scripts/
chmod +x scripts/preserve-session-records.sh
./scripts/test-preserve-session-records.sh
```

Expected: `preserve-session-records: all cases pass`.

- [x] **Step 6: Run the lint guards**

```bash verified:these are the three commands `.myflow/project.md` `## lint` declares
./scripts/check-vocabulary.sh
./scripts/check-references.sh
./scripts/check-plan-provenance.sh
```

---

## 4. Wire preservation into the commands

### Task 4: finish run 1 and `/myflow-do`'s commit path preserve the records

**Files:**
- Modify: `skills/myflow-finish/SKILL.md` — `## 1.2 Commit the staged work`
- Modify: `skills/myflow-do/SKILL.md` — the paragraph beginning "**The one commit exception.**"
- Modify: `skills/myflow-contracts/pipeline.md` — `### Run 1 — the branch is not merged`

**Interfaces:**
- Consumes: `scripts/preserve-session-records.sh` from Task 3, three positional arguments.
- Produces: the guarantee Task 5 relies on — by the time run 2 removes the state-directory artifact
  source, a copy exists under `docs/superpowers/artifacts/`.

- [x] **Step 1: Add the copy to finish run 1, before staging**

In `skills/myflow-finish/SKILL.md` `## 1.2 Commit the staged work`, insert before the sentence about
running `git add -A`:

```markdown unverified:confirm the paragraph ordering after insertion — the copy must be described before `git add -A`, or the records miss the commit
**Preserve the session records first.** Run
`scripts/preserve-session-records.sh <worktree> <name> <state-dir>` before staging, so the SDD ledger,
the review panel record and the proposal artifact source land in the same commit as the work they
describe. A source that does not exist is reported and skipped — never a failure, and never a reason
to stop the integration.
```

- [x] **Step 2: Add the refresh to `/myflow-do`'s commit path**

In `skills/myflow-do/SKILL.md`, append to the paragraph beginning "**The one commit exception.**":

```markdown unverified:confirm the paragraph still opens with "**The one commit exception.**" at implementation time
Run `scripts/preserve-session-records.sh` before that commit as well, so a fix round raised after a
PR is open refreshes the preserved records rather than leaving them a round stale. It overwrites in
place; it never creates a second dated copy.
```

- [x] **Step 3: Record the duty in the contract**

In `skills/myflow-contracts/pipeline.md`, under `### Run 1 — the branch is not merged`, extend the
sentence listing what all three routes commit:

```markdown unverified:confirm the existing sentence still enumerates implementation, the manual test guide, and the openspec artifacts
All three routes first commit the staged work — implementation, `docs/manual-test/<name>.md`, the
`openspec/` planning artifacts, and the session records preserved under `docs/superpowers/` (the SDD
ledger, the review panel record, and the proposal artifact source). The records are copied out of the
gitignored worktree before staging; a missing source is reported and skipped, never fatal.
```

- [x] **Step 4: Run the lint guards**

```bash verified:these are the three commands `.myflow/project.md` `## lint` declares
./scripts/check-vocabulary.sh
./scripts/check-references.sh
./scripts/check-plan-provenance.sh
```

---

### Task 5: run 2 removes the artifact source

**Files:**
- Modify: `skills/myflow-contracts/pipeline.md` — `### Run 2 — the branch is merged`, in the numbered
  sequence, and `#### Worktree cleanup`
- Modify: `skills/myflow-finish/SKILL.md` — run 2's numbered outline

**Interfaces:**
- Consumes: the preservation guarantee from Task 4 — the removal is conditional on a preserved copy
  existing.
- Produces: nothing later tasks depend on.

- [x] **Step 1: Add the removal step to the contract**

In `skills/myflow-contracts/pipeline.md` `### Run 2 — the branch is merged`, add after the
worktree-cleanup step and before the `FINISHED` write:

```markdown unverified:confirm the surrounding numbered list renumbers correctly after insertion
5. **Remove the proposal artifact source.** `/myflow-start` wrote
   `<state-dir>/<name>-proposal-artifact.html` so a revision round could republish to the same URL.
   Delete it here, disclosed the same way the worktree removal is — **but only if a preserved copy
   exists** under `docs/superpowers/artifacts/`. The terminal state file keeps `artifactUrl`
   indefinitely; removing the only source that could republish it would leave a URL advertised and
   unrepublishable. No preserved copy → leave the file and say so.
```

- [x] **Step 2: Mirror it in the skill's outline**

In `skills/myflow-finish/SKILL.md`, run 2's numbered outline gains the matching entry, pointing at
the contract rather than restating the rationale:

```markdown unverified:confirm the numbered outline's existing items keep their order after insertion
5. **Remove the proposal artifact source** from the state directory, per **Finish contract** — only
   when the preserved copy under `docs/superpowers/artifacts/` exists.
```

- [x] **Step 3: Add the guardrail**

```markdown verified:matches the existing guardrail style in skills/myflow-finish/SKILL.md's Guardrails list
- **Never** delete the proposal artifact source without a preserved copy in the repository.
```

- [x] **Step 4: Run the lint guards**

```bash verified:these are the three commands `.myflow/project.md` `## lint` declares
./scripts/check-vocabulary.sh
./scripts/check-references.sh
./scripts/check-plan-provenance.sh
```

---

## 5. The effort field

### Task 6: `effort` in the state-file and self-heal contracts

**Files:**
- Modify: `skills/myflow-contracts/state-file.md` — the JSON shape block and the field list beneath it
- Modify: `skills/myflow-contracts/state-self-heal.md` — the closed-schema paragraph

**Interfaces:**
- Consumes: nothing.
- Produces: the field name `effort` and its three values `low` / `medium` / `high`, plus the rule that
  an absent key is legal. Task 7 writes the field; every command carries it forward.

- [x] **Step 1: Add the field to the documented JSON shape**

In `skills/myflow-contracts/state-file.md`, the example object gains one key. Keep the existing keys
byte-identical:

```json unverified:confirm the surrounding fence and the key order in the file at implementation time
{
  "state": "IN_PROGRESS",
  "branch": "openspec/<name>",
  "worktrees": {
    "/absolute/path/to/worktree": "<merge-base sha>"
  },
  "artifactUrl": null,
  "jiraIssue": null,
  "effort": null,
  "prUrl": null,
  "updatedAt": "2026-07-28T10:00:00Z",
  "updatedBy": "/myflow-do"
}
```

- [x] **Step 2: Document the field, and the absence carve-out**

Add to the bulleted field list in the same file, after the `jiraIssue` bullet:

```markdown unverified:confirm the bullet list's existing formatting (leading `- ` and backticked field name) at implementation time
- `effort` — the reasoning effort chosen for this change's planning: `"low"`, `"medium"`, `"high"`, or
  `null` when none was chosen. Written only by `/myflow-start`, on the run that **creates** the
  change; every other command **carries it forward verbatim**. It governs `/myflow-start`'s own
  reasoning depth and nothing else — no command derives behaviour from it, and the review panel's
  breadth is never scaled from it. See **Effort** below.

**A state file that omits `effort` entirely is valid**, and is read as `null`. This is a deliberate
exception to the closed-schema rule in **State self-heal**
(`skills/myflow-contracts/state-self-heal.md`), which otherwise makes a file unparseable both for
missing a documented field and for carrying an undocumented one. Without the exception every file
written before this field existed would be routed through self-heal, which announces unrecovered
fields and rewrites from artifact inference — a loud correction for a value nobody had the chance to
set. `effort` is the first field added since the schema closed, so the carve-out is stated rather
than inferred: `artifactUrl`, `jiraIssue` and `prUrl` are all *present and nullable*, which is a
different thing from *absent*.
```

- [x] **Step 3: Name the exception where the closed-schema rule is stated**

In `skills/myflow-contracts/state-self-heal.md`, append to the paragraph that closes the schema in
both directions:

```markdown unverified:confirm the closed-schema paragraph still ends with the sentence about recognising a legacy field
**One documented exception exists: `effort`.** A file that omits it is valid and reads as `null`, per
**State file** (`skills/myflow-contracts/state-file.md`). An omitted `effort` is therefore not a
failed recovery and is never named among the unrecovered fields; every other absent documented field
still makes the file unparseable.
```

- [x] **Step 4: Run the lint guards**

```bash verified:these are the three commands `.myflow/project.md` `## lint` declares
./scripts/check-vocabulary.sh
./scripts/check-references.sh
./scripts/check-plan-provenance.sh
```

Expected: all three exit 0. `check-references.sh` resolves the `**Effort**` and `**State file**`
section references introduced here — if it fails, the named section does not exist yet, and Task 7
creates it. Reorder these two tasks rather than weakening the reference.

---

### Task 7: `/myflow-start` asks for the effort level

**Files:**
- Modify: `skills/myflow-start/SKILL.md` — a new section before `## B. Basic Workflow #1 —
  Brainstorming`, plus `## F. Write state and hand off` and `## Guardrails`
- Modify: `skills/myflow-contracts/state-file.md` — add the `## Effort` section Task 6 referenced

**Interfaces:**
- Consumes: the `effort` field and its values from Task 6.
- Produces: nothing later tasks depend on.

- [x] **Step 1: Add the `## Effort` section to the state-file contract**

Append to `skills/myflow-contracts/state-file.md`:

```markdown unverified:confirm no `## Effort` heading already exists in the file
## Effort

Three levels, offered by `/myflow-start` on the run that creates a change, with `medium` the default:

| Level | What it changes |
|-------|-----------------|
| `low` | Questions batched rather than asked one at a time; the design presented once; `tasks.md` grouped more coarsely |
| `medium` | The checklist followed with related questions grouped |
| `high` | Each checklist item worked separately, alternatives enumerated per open question, each design section approved on its own |

**No level may switch a gate off.** Brainstorming runs, the design approval gate holds,
writing-plans runs, and `tasks.md` is never left a thin scaffold — at every level. A lower level
means fewer rounds and coarser grouping, never a gate that does not run. An effort level able to
skip a gate would be a way to skip review rather than a way to size the thinking inside it.
```

- [x] **Step 2: Add the ask to `/myflow-start`**

Insert into `skills/myflow-start/SKILL.md`, between `## A. Resolve the change` and
`## B. Basic Workflow #1 — Brainstorming`:

```markdown unverified:confirm section letters B through F do not need renumbering after inserting an unlettered section
## Ask the effort — creating runs only

**Ask once, on the run that creates the change**, and never again for it. "Creates" means the state
file does not exist — not a guess about the operator or the conversation.

Use **AskUserQuestion**, the same mechanism `/myflow-finish` uses for its integration choice. Effort
is **never** an argument: the only argument this command accepts is the optional change name, and
anything else is still reported rather than interpreted.

> **How much effort should planning this change take?**
> - **Medium** *(default, recommended)* — the checklist followed with related questions grouped
> - **High** — each checklist item worked separately, every design section approved on its own
> - **Low** — questions batched, the design presented once

Record the answer for the state write in section F. The levels and what they may change are defined
under **Effort** in **State file** (`skills/myflow-contracts/state-file.md`) — that file is
canonical; do not restate the table here.

**Revising an existing proposal** (the change is already at `STARTED`): do not ask. Read `effort`
from the state file, state which level is being reused, and proceed at it. A file that records no
level is planned at `medium`, and that is said in the handoff too.
```

- [x] **Step 3: Write the field in section F**

In `## F. Write state and hand off`, the JSON block gains the key:

```json unverified:confirm the block's existing keys and their order at implementation time
{
  "state": "STARTED",
  "branch": null,
  "worktrees": {},
  "artifactUrl": "<published URL>",
  "jiraIssue": "<resolved key, or null>",
  "effort": "<low|medium|high, or null>",
  "prUrl": null,
  "updatedAt": "<ISO-8601 UTC now>",
  "updatedBy": "/myflow-start"
}
```

- [x] **Step 4: Add the effort line to the handoff block**

In the same section's fenced handoff template, add a line beneath the `**Jira:**` line:

```markdown unverified:confirm the handoff template's existing lines and their order at implementation time
**Effort:** <level> | <level> (reused from the creating run) | not recorded — planned at medium
```

- [x] **Step 5: Add the guardrails**

```markdown verified:matches the existing guardrail style in skills/myflow-start/SKILL.md's Guardrails list
- **Never** ask for an effort level on a revision round — read the recorded one and say so.
- **Never** let an effort level skip brainstorming, the design approval gate, writing-plans, or leave
  `tasks.md` a scaffold. It sizes the thinking inside the gates, never the gates.
```

- [x] **Step 6: Run the lint guards**

```bash verified:these are the three commands `.myflow/project.md` `## lint` declares
./scripts/check-vocabulary.sh
./scripts/check-references.sh
./scripts/check-plan-provenance.sh
```

---

## 6. Declared configuration and the full sweep

### Task 8: `.myflow/project.md`, the docs, and every guard green

**Files:**
- Modify: `.myflow/project.md` — `## test`, and the stale `## lint` note
- Modify: `CLAUDE.md` and `AGENTS.md` — the `/myflow-start` and `/myflow-finish` rows in the command
  tables, if they describe behaviour this change altered

**Interfaces:**
- Consumes: the four scripts from Tasks 1 and 3, by exact filename.
- Produces: nothing.

- [x] **Step 1: Add the two harnesses to `## test`**

`.myflow/project.md`'s `## test` section holds one fenced `bash` block. Its contents become exactly
these five lines, in this order — the two new harnesses appended to the three already there:

```bash unverified:confirm the three existing lines are unchanged before appending the two new ones
scripts/test-setup.sh
scripts/test-check-references.sh
scripts/test-check-plan-provenance.sh
scripts/test-check-finish-preflight.sh
scripts/test-preserve-session-records.sh
```

Do **not** add `check-finish-preflight.sh` or `preserve-session-records.sh` to `## lint`. Both need a
worktree, a branch and a resolved base ref; a lint step that cannot run against a bare tree would fail
on every unrelated invocation. State that reason in `## lint` so the omission reads as a decision.

- [x] **Step 2: Correct the stale provenance-guard note**

`.myflow/project.md` `## lint` states that `check-plan-provenance.sh` is currently expected to exit 1,
with its hits confined to `openspec/changes/kan-8-myflow-updates/tasks.md`. That change has since been
archived, and the guard excludes `openspec/changes/archive/`, so the note describes a state that no
longer exists. Confirm before rewriting:

```bash unverified:this is the confirmation step itself — use what it reports, not what this plan predicts
./scripts/check-plan-provenance.sh; echo "exit=$?"
```

Rewrite the note to match what it reports. If it exits 0, say the guard is expected to pass and keep
the exit-code contract summary, which is still accurate.

- [x] **Step 3: Update the command tables in `CLAUDE.md` and `AGENTS.md`**

Both files carry a `/myflow-finish` row describing run 2 as "removes the worktrees and branches after
four gating safety checks plus a disclosure". Extend the `/myflow-start` row to mention the effort ask
and the `/myflow-finish` row to mention the preflight verdict and the artifact-source removal. Keep
the two files in step with each other — they are near-duplicates by design, and
`check-references.sh` reads both.

- [x] **Step 4: Run every declared command, lint and test**

```bash verified:these are exactly the commands `.myflow/project.md` declares under `## lint` and `## test`
./scripts/check-vocabulary.sh
./scripts/check-references.sh
./scripts/check-plan-provenance.sh
./scripts/test-setup.sh
./scripts/test-check-references.sh
./scripts/test-check-plan-provenance.sh
./scripts/test-check-finish-preflight.sh
./scripts/test-preserve-session-records.sh
```

Expected: every command exits 0.

- [x] **Step 5: Exercise the installer in a sandbox**

```bash verified:copied from `.myflow/project.md` `## run`, which declares this as the way to exercise the installer
SANDBOX="$(mktemp -d)"
HOME="$SANDBOX" ./setup.sh global
```

Expected: exits 0. The two new scripts are repository-local and are **not** expected to appear in the
sandbox — only skills, commands and rules are installed. The contract text that points at them is.

- [x] **Step 6: Validate the OpenSpec change**

```bash verified:this exact invocation reported "Change 'kan-19-finish-safety-records-and-effort' is valid" while this plan was written
openspec validate kan-19-finish-safety-records-and-effort --strict
```

Expected: `Change 'kan-19-finish-safety-records-and-effort' is valid`.

---

## 7. The model record's own contract text

### Task 9: `pipeline.md`'s `## Model policy` section stops disclaiming its durability

Added during the first fix wave. The primary review slot found that this change ships a spec delta
its own contract text contradicts: `specs/myflow-model-policy/spec.md` REMOVES *Every dispatch
records the model it used, for the session that made it* and ADDS a durable replacement, and
`proposal.md` promises the capability "stops disclaiming its own durability" — while
`skills/myflow-contracts/pipeline.md`'s `## Model policy` section still asserted the record was
session-scoped, told the reader not to plan an audit around it, and named "making it durable" as a
change belonging to somebody else. Tasks 3 and 4 make it durable, so every sentence of that
paragraph was false on merge, and its last sentence pointed a future reader at the work this branch
had already done. The task was missing because `proposal.md`'s Impact list named
`pipeline.md` only for the Finish contract.

**Files:**
- Modify: `skills/myflow-contracts/pipeline.md` — the `## Model policy` section's scope paragraph
- Modify: `openspec/changes/kan-19-finish-safety-records-and-effort/proposal.md` — the Impact list

**Interfaces:**
- Consumes: the preservation guarantee from Tasks 3 and 4 — the prose is only true because run 1
  copies the ledger into `docs/superpowers/ledgers/`.
- Produces: nothing later tasks depend on.

- [x] **Step 1: Replace the session-scoped paragraph with what the ADDED requirement states**

The paragraph beginning "**This record is session-scoped.**" was rewritten from the delta spec's
own **Scope: the record outlives the change** paragraph rather than from memory: the ledger is
authored under `.superpowers/` in a worktree run 2 removes, run 1 preserves it into
`docs/superpowers/ledgers/` first, and an after-the-fact audit therefore reads the preserved ledger.
The preservation duty is stated once, under `### Run 1 — the branch is not merged`, and this section
depends on it rather than restating it — `pipeline.md` is canonical, so the procedure is not
duplicated into the section that consumes it.

- [x] **Step 2: Leave the `unknown (agent-defined)` rule exactly as it stands**

The delta spec is explicit that durability must not create pressure to fill in a value nothing
measured. The two paragraphs stating that rule are unchanged; one sentence was added *after* them
saying durability is a **stronger** reason to leave an unobserved entry unobserved, since a
persisting record makes an invented model slug permanent. No step anywhere in this change writes or
rewrites ledger content on the way into the repository.

- [x] **Step 3: Name the section in the Impact list**

`proposal.md`'s Impact list named `skills/myflow-contracts/pipeline.md` for the Finish contract
only, which is why the section was missed. It now names the `## Model policy` scope paragraph and
why it had to move.

- [x] **Step 4: Run the lint guards and validate the change**

```bash verified:these are the three commands `.myflow/project.md` `## lint` declares, plus the validation Task 8 step 6 already runs
./scripts/check-vocabulary.sh
./scripts/check-references.sh
./scripts/check-plan-provenance.sh
openspec validate kan-19-finish-safety-records-and-effort --strict
```

Expected: all three guards exit 0 and the change validates. `check-vocabulary.sh` is the one that
matters for this task — the rewritten paragraph must not reintroduce retired vocabulary.

---

## 8. The record the fix waves themselves leave

### Task 10: the two fix waves are documented, and the contracts they touched name their own authority

Added during the second fix wave. Task 9 set the precedent that a repair found by the panel gets a
task of its own rather than arriving as an undocumented edit; this task closes the rest of that gap.
Four repairs from the first wave had shipped with no documented trail at all, and
`scripts/check-plan-provenance.sh` is structurally blind to the drift — it checks that a fenced block
carries a provenance tag, not that the code inside it still matches the file that shipped.

**Files:**
- Modify: `openspec/changes/kan-19-finish-safety-records-and-effort/tasks.md` — fix-wave notes on
  Steps 2 and 4 of Tasks 1 and 3
- Modify: `skills/myflow-contracts/pipeline.md` — `### Run 1 — the branch is not merged`, the
  preservation call's outcomes
- Modify: `skills/myflow-finish/SKILL.md` — `## 1.2 Commit the staged work`
- Modify: `skills/myflow-do/SKILL.md` — the paragraph beginning "**The one commit exception.**"
- Modify: `skills/myflow-contracts/state-file.md` — the `## Effort` section's opening
- Modify: `openspec/changes/kan-19-finish-safety-records-and-effort/specs/myflow-effort/spec.md`
- Modify: `openspec/changes/kan-19-finish-safety-records-and-effort/specs/myflow-finish-cleanup/spec.md`

**Interfaces:**
- Consumes: the shipped `scripts/preserve-session-records.sh` and its harness from Task 3, whose
  failure shapes this task documents.
- Produces: nothing later tasks depend on.

- [x] **Step 1: Amend Steps 2 and 4 of Tasks 1 and 3 with fix-wave notes**

Each note says what was found, why it was changed, and that **the file on disk is authoritative over
the plan's illustrative block**. The blocks are deliberately *not* re-pasted: a second copy of a file
that is still changing is exactly what went stale, and re-pasting would only reset the clock on the
same defect. Every fenced block keeps the provenance tag it had.

- [x] **Step 2: Give `preserve-session-records.sh`'s three outcomes an actionable table at the
  canonical site**

The script has three outcomes — an absent source (`skipped:`, exit 0), a successful copy
(`preserved:`, exit 0), and a copy that was attempted and refused or failed (stderr, non-zero) — and
both call sites documented only the first. Neither said what an agent does when the script exits
non-zero for another reason, while `check-finish-preflight.sh`'s four outcomes got exactly such a
table in the first wave. The two scripts' failure semantics are now documented with equal rigour.

The table lives in `pipeline.md`, which is canonical, and both skills point at it in one line each.
The rule it states is that a failed or refused write is **reported to the operator and never silent**,
and is **not** a reason to abandon an integration whose work is already committed — two directions
that both hold, which is why the rationale is stated once rather than paraphrased twice.

- [x] **Step 3: Name which file is canonical for the effort default, and for what**

The default was stated as fact in two places with neither naming the other. It is **not** resolved by
stripping the requirement from the spec: in OpenSpec, `openspec/specs/` is the normative requirement
source and skills implement it, so deleting the requirement would invert that direction. Instead the
spec states the normative requirement and says the operational table lives in the **Effort** section
of `skills/myflow-contracts/state-file.md`; that section says the spec is normative and that a change
goes there first. Both stay accurate, and a later editor can tell which to change first.

- [x] **Step 4: Add the refusal scenarios to the finish-cleanup delta spec**

Destination containment, source containment and change-name rejection were implemented and tested but
appeared nowhere in the spec, which under-described an observable failure mode. Three scenarios and
the two requirement paragraphs behind them were added under **Session records are preserved in the
repository**.

- [x] **Step 5: Run the lint guards, the harnesses, and validate the change**

```bash verified:these are the three lint commands and the five harnesses `.myflow/project.md` declares, plus the validation Task 8 step 6 runs
./scripts/check-vocabulary.sh
./scripts/check-references.sh
./scripts/check-plan-provenance.sh
./scripts/test-setup.sh
./scripts/test-check-references.sh
./scripts/test-check-plan-provenance.sh
./scripts/test-check-finish-preflight.sh
./scripts/test-preserve-session-records.sh
openspec validate kan-19-finish-safety-records-and-effort --strict
```

Expected: every command exits 0 and the change validates. `check-references.sh` is the one that
matters most here — this task adds cross-file section references in both directions.

---

## Verification summary

| Requirement | Task |
|-------------|------|
| Run 2 verifies the merge before changing anything (three signals, ordered) | 1, 2 |
| No recorded merge base produces a refusal | 1, 2 |
| A merged branch with a dirty worktree refuses | 1, 2 |
| Session records are preserved in the repository | 3, 4 |
| A missing source is skipped, not fatal | 3, 10 |
| A refused source or destination fails that copy, loudly, and is reported | 3, 10 |
| A change name that is not one plain component is rejected | 3 |
| A fix round refreshes rather than duplicates | 3, 4 |
| An unobservable model stays unobserved | 4 (no step writes ledger content), 9 |
| The model record's contract text matches its new durability | 9 |
| Run 2 removes the proposal artifact source | 5 |
| `/myflow-start` asks for an effort level once per change | 7 |
| Effort scales reasoning, never the gates | 7 |
| The chosen effort is recorded, and its absence is legal | 6, 7 |
| Effort joins the carried-forward and announced field lists | 6 |
| Each file states what it is canonical for, and the fix waves leave a trail | 9, 10 |
| This repository declares its own myflow project configuration | 8 |
