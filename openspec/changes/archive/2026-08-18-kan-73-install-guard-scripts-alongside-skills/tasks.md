> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

**Goal:** A guard a `/myflow-*` command invokes is reachable from that command's installed skill, in
every harness, and a guard that is missing is reported once and loudly instead of discovered
silently at each call site.

## Global Constraints

- **Tasks 1-6 run in this order, and the order is load-bearing.** Task 1 fixes a repo-root
  assumption that task 2 would otherwise make wrong. Task 2 creates the symlink farms. Task 3
  states the resolution rule and rewrites every invoking call site. Task 4 adds the presence check.
  Task 5 adds the guard that keeps tasks 1–3 true — landing it earlier would land it red. Task 6
  adds the install-level regression test.
- **Task 7 is independent of all of them and was added mid-run**, at the operator's instruction, to
  carry prose that is unrelated to KAN-73's subject. It may run at any point and shares no file with
  tasks 1, 2, 4, 5 or 6; it shares `scripts/check-contract-budget.sh` and
  `skills/myflow-fast/SKILL.md` with task 3, so the two never run concurrently.
- **No guard's logic changes.** This change moves guards and fixes how they are addressed. The only
  edits inside a guard's body are task 1's two root derivations. No check gains, loses or alters a
  rule.
- **Symlinks are committed as symlinks.** `git` stores a symlink as its target string; do not
  `git add` a dereferenced copy. Verify with `git ls-files -s`, where a symlink has mode `120000`.
- **Every path in a skill's `scripts/` directory is relative.** An absolute target would bake this
  machine's checkout path into the repository and break every other clone.
- **No task edits `openspec/` or `docs/superpowers/`.**

## Baseline

All measured 2026-08-17 against `7922242`.

- `scripts/test-setup.sh` reports `✓ PASS — 208 assertions, 0 failures`.
  <!-- measured: scripts/test-setup.sh @ 7922242 -->
- `scripts/check-references.sh` and `scripts/check-vocabulary.sh` both exit clean.
  <!-- measured: scripts/check-references.sh; scripts/check-vocabulary.sh @ 7922242 -->
- `scripts/check-contract-budget.sh` exits clean at `HEAD`, where `skills/myflow-fast/SKILL.md` is
  17915 bytes against its 18225 budget.
  <!-- measured: git show HEAD:skills/myflow-fast/SKILL.md | wc -c, and scripts/check-contract-budget.sh @ 7922242 -->
- The working tree at the time of planning carried 21 uncommitted lines across
  `skills/myflow-contracts/finish-contract.md` and `skills/myflow-fast/SKILL.md`, unrelated to this
  change, which push that file to 19128 bytes and make the budget guard fail. The worktree this
  plan is implemented in branches from `HEAD` and is unaffected.
  <!-- measured: git diff --stat and scripts/check-contract-budget.sh against the working tree on 2026-08-17 -->

**The gap, as it stands:** no `scripts/` directory exists under any installed `myflow-*` skill.
`~/.claude/skills/myflow-finish/` holds `SKILL.md` and `SKILL-rationale.md` and nothing else.
<!-- measured: ls -R ~/.claude/skills/myflow-finish/ and ls ~/.claude/skills/*/scripts on 2026-08-17 -->

**The mechanism this plan rests on, verified before it was planned:** a relative symlink inside a
symlinked skill directory resolves through the installed path, and the guards'
`SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"` idiom then finds a `lib/`
symlinked beside it.
<!-- measured: scratch tree reproducing the installed layout, invoked via the symlinked skill dir; printed SCRIPT_DIR=<scratch>/repo/skills/myflow-finish/scripts, then LIB-SOURCED, exit 0 -->

---

### 1 Two guards stop deriving a repository root from a fixed depth

**Build:** green

**Files:**
- Create: `scripts/lib/resolve-file.sh`
- Modify: `scripts/plan-dispatch-bundles.sh`
- Modify: `scripts/check-workspace-isolation.sh`
- Modify: `scripts/test-plan-dispatch-bundles.sh`
- Modify: `scripts/test-check-workspace-isolation.sh`

`scripts/lib/resolve-file.sh` was **not** in this task's original plan, which said to port the
`resolve_file` helper verbatim rather than extract it. The review panel raised that as a Critical
(F1) and it was right: the instruction produced five copies of one helper, and two of them had
already drifted apart on whether to pass `--` to `readlink`, `dirname` and `cd`. The extracted
library carries the hardened form, and ships to the skills through the `lib` symlink that already
existed for `scripts/lib/panel-record.sh`. The two pre-existing copies in
`scripts/preserve-session-records.sh` and `scripts/gather-self-review-context.sh` are left alone as
out of this change's scope, and are recorded as a follow-up.

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: a root derivation that is the same whichever path the guard was invoked through — the
  precondition task 2 needs before a guard becomes reachable from a second directory.

Both guards compute a repository root as exactly one level above their own directory:

```bash verified:read from the two files at 7922242
# scripts/plan-dispatch-bundles.sh:59
REPO_ROOT="${PLAN_DISPATCH_BUNDLES_ROOT:-$(cd "$SCRIPT_DIR/.." && pwd)}"
# scripts/check-workspace-isolation.sh:104
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
```

That holds only while the guard lives at `<repo>/scripts/`. Once task 2 makes it reachable at
`skills/myflow-do/scripts/`, `$SCRIPT_DIR/..` is `<repo>/skills/myflow-do` — a directory that
exists, so the guard proceeds against it and returns a confident wrong answer rather than an error.

The blast radius is bounded today: both take an explicit target, and the `..` form is only the
no-argument default. Fixing it is still task 1, because a wrong default that silently names a skill
directory is worse than no default.

- [x] **Step 1: The failing tests, written first**

Add one case to each harness. Each creates a scratch tree with the guard reachable at two depths —
its real home and a `skills/<name>/scripts/` symlink — invokes it through the symlink with no root
argument, and asserts the root it derives is the repository root, not the skill directory.

`scripts/test-check-workspace-isolation.sh` already builds scratch project trees; follow its
existing fixture shape. `scripts/test-plan-dispatch-bundles.sh` already exercises the
`PLAN_DISPATCH_BUNDLES_ROOT` override; the new case is the *absence* of that override.

Run them and watch both fail before touching either guard:

```bash unverified:neither case exists yet — confirm the -run/name filter matches once step 1 has written them
scripts/test-plan-dispatch-bundles.sh 2>&1 | tail -20
scripts/test-check-workspace-isolation.sh 2>&1 | tail -20
```

- [x] **Step 2: Derive the root from the resolved physical location**

In both guards, replace the `$SCRIPT_DIR/..` derivation with one that resolves the script's own
symlinks first, so the answer is the guard's real home regardless of the path it was invoked
through. Keep `PLAN_DISPATCH_BUNDLES_ROOT` as the explicit override it already is, and keep
`check-workspace-isolation.sh`'s explicit project-root argument taking precedence exactly as today.

macOS `/bin/sh` has no `readlink -f`; resolve with a portable loop, or reuse the one
`scripts/preserve-session-records.sh` already carries at lines 138-145 and
`scripts/gather-self-review-context.sh` at 341-348 rather than writing a third.

- [x] **Step 3: Say why, in each guard's own comment**

Each derivation gets a comment stating that the guard is reachable from more than one directory,
that a fixed depth therefore has more than one answer, and that the wrong one is a directory which
exists — which is why this resolves physically instead of counting levels.

- [x] **Step 4: Both harnesses green**

```bash verified:both harnesses exist and pass at 7922242; the assertion counts rise by the cases step 1 adds
scripts/test-plan-dispatch-bundles.sh && scripts/test-check-workspace-isolation.sh
```

**Tests:** the new root-derivation case in `scripts/test-plan-dispatch-bundles.sh`, and the new
root-derivation case in `scripts/test-check-workspace-isolation.sh`.

**Regression:** Reverting this task makes both guards derive a skill directory as the repository
root whenever they are invoked through a skill's `scripts/` directory with no explicit root —
`plan-dispatch-bundles.sh` then scans `skills/myflow-do/openspec/changes`, which does not exist, and
reports nothing to dispatch rather than failing.

**Baseline:** before=0 after=2 root-derivation cases across the two harnesses.

**Commit:** `fix(kan-73-install-guard-scripts-alongside-skills): derive a guard's repo root from its resolved location`

---

### 2 The per-skill `scripts/` symlink farms

**Build:** green

**Files:**
- Create: `skills/myflow-do/scripts/check-panel-diff-size.sh`
- Create: `skills/myflow-do/scripts/check-panel-reproducers.sh`
- Create: `skills/myflow-do/scripts/check-task-commit-fields.py`
- Create: `skills/myflow-do/scripts/check-task-commit-fields.sh`
- Create: `skills/myflow-do/scripts/check-unfinished-work.sh`
- Create: `skills/myflow-do/scripts/check-workspace-isolation.sh`
- Create: `skills/myflow-do/scripts/commit-split.sh`
- Create: `skills/myflow-do/scripts/lib`
- Create: `skills/myflow-do/scripts/plan-dispatch-bundles.py`
- Create: `skills/myflow-do/scripts/plan-dispatch-bundles.sh`
- Create: `skills/myflow-do/scripts/prepare-workspace.sh`
- Create: `skills/myflow-do/scripts/preserve-session-records.sh`
- Create: `skills/myflow-do/scripts/reproducer-metachars.sh`
- Create: `skills/myflow-do/scripts/run-reproducer.sh`
- Create: `skills/myflow-fast/scripts/check-cleanup-complete.sh`
- Create: `skills/myflow-fast/scripts/check-finish-preflight.sh`
- Create: `skills/myflow-fast/scripts/check-panel-diff-size.sh`
- Create: `skills/myflow-fast/scripts/check-panel-reproducers.sh`
- Create: `skills/myflow-fast/scripts/check-task-commit-fields.py`
- Create: `skills/myflow-fast/scripts/check-task-commit-fields.sh`
- Create: `skills/myflow-fast/scripts/check-unfinished-work.sh`
- Create: `skills/myflow-fast/scripts/check-workspace-isolation.sh`
- Create: `skills/myflow-fast/scripts/commit-split.sh`
- Create: `skills/myflow-fast/scripts/gather-self-review-context.sh`
- Create: `skills/myflow-fast/scripts/lib`
- Create: `skills/myflow-fast/scripts/plan-dispatch-bundles.py`
- Create: `skills/myflow-fast/scripts/plan-dispatch-bundles.sh`
- Create: `skills/myflow-fast/scripts/prepare-workspace.sh`
- Create: `skills/myflow-fast/scripts/preserve-session-records.sh`
- Create: `skills/myflow-fast/scripts/reproducer-metachars.sh`
- Create: `skills/myflow-fast/scripts/run-reproducer.sh`
- Create: `skills/myflow-finish/scripts/check-cleanup-complete.sh`
- Create: `skills/myflow-finish/scripts/check-finish-preflight.sh`
- Create: `skills/myflow-finish/scripts/check-unfinished-work.sh`
- Create: `skills/myflow-finish/scripts/commit-split.sh`
- Create: `skills/myflow-finish/scripts/gather-self-review-context.sh`
- Create: `skills/myflow-finish/scripts/lib`
- Create: `skills/myflow-finish/scripts/preserve-session-records.sh`

Every entry above is a relative symlink into repo-root `scripts/`. `myflow-status` was dropped from
this list during implementation: it invokes no guard, so it ships none — see `design.md`'s corrected
`scripts-dirs-on-command-skills-only` decision. The list is written out in full,
rather than as the four directories it forms, because `scripts/check-task-commit-fields.sh` checks
each declared path against the real commit — a directory does not satisfy it.

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: task 1's fixed root derivation.
- Produces: the directories tasks 3, 4, 5 and 6 all address.

Every entry is a relative symlink counted from the symlink's own directory, so `../../../<name>`
from `skills/<skill>/scripts/` lands on the repository root:

```bash verified:the resolution mechanism was tested against a scratch tree reproducing the installed layout before this plan was written
ln -s ../../../scripts/check-unfinished-work.sh skills/myflow-finish/scripts/check-unfinished-work.sh
ln -s ../../../scripts/lib                      skills/myflow-finish/scripts/lib
```

`lib` is symlinked as a **directory**, not file by file — `check-unfinished-work.sh` and
`check-panel-reproducers.sh` both source `$SCRIPT_DIR/lib/panel-record.sh`, and a directory symlink
keeps a future library file reachable without a fifth edit here.

`myflow-start` gets **no** `scripts/` directory: its plan-provenance and build-green step runs *the
project's configured* guards, resolved through that project's `.myflow/project.md`, so it invokes
none of its own. `myflow-contracts` gets none either — it is never the running command.

- [x] **Step 1: Create the four directories and their symlinks**

Follow the map in `Files:` above exactly. Sibling dependencies are part of the map and not
optional — a guard shipped without the sibling it resolves from `$SCRIPT_DIR` fails at the moment it
is needed:

| Guard | Needs beside it |
|---|---|
| `check-unfinished-work.sh` | `lib/` |
| `check-panel-reproducers.sh` | `lib/`, `reproducer-metachars.sh` |
| `run-reproducer.sh` | `reproducer-metachars.sh` |
| `check-task-commit-fields.sh` | `check-task-commit-fields.py` |
| `plan-dispatch-bundles.sh` | `plan-dispatch-bundles.py` |
| `prepare-workspace.sh` | `check-workspace-isolation.sh` |

<!-- measured: grep -oE '\$\{?SCRIPT_DIR\}?/[A-Za-z0-9./_-]+' over each pipeline guard @ 7922242 -->

- [x] **Step 2: Verify every link resolves and is stored as a link**

```bash unverified:the directories do not exist yet — run once step 1 has created them
for d in skills/myflow-do skills/myflow-finish skills/myflow-status skills/myflow-fast; do
  for e in "$d"/scripts/*; do
    [ -e "$e" ] || { echo "BROKEN: $e"; exit 1; }
    [ -L "$e" ] || { echo "NOT A SYMLINK: $e"; exit 1; }
  done
done
git add skills/*/scripts && git ls-files -s skills/*/scripts | grep -v '^120000' && echo "DEREFERENCED COPY COMMITTED" && exit 1
echo "every entry resolves and is stored as a symlink"
```

- [x] **Step 3: Verify a guard runs through an installed skill path**

Install into a sandbox `HOME` and invoke the guard that has the hardest dependency — it sources
`lib/` — through the installed path:

```bash unverified:depends on step 1's links; the sandbox install idiom is .myflow/project.md's own `## run` recipe
SANDBOX="$(mktemp -d)"
HOME="$SANDBOX" ./setup.sh global >/dev/null
"$SANDBOX/.claude/skills/myflow-finish/scripts/check-unfinished-work.sh" 2>&1 | head -3
```

Absent arguments the guard emits no verdict line and exits non-zero — that is its documented
behaviour and is the expected result here. What this step proves is the opposite of a `lib` failure:
it must **not** print `cannot find` or name a missing library.

**Tests:** none added by this task. Its verification is step 2's resolve-and-symlink assertions and
step 3's installed-path invocation, run inline. Both become permanent checks later — in the guard
task 5 creates, and in the install regression task 6 adds — and are declared there, by the tasks
whose commits actually carry those files.

**Regression:** Reverting this task returns every installed skill to having no `scripts/` directory,
so every guard call site resolves to nothing and every check silently falls back to a hand-run —
the defect KAN-73 was filed for.

**Baseline:** before=0 after=4 skill directories carrying a `scripts/` directory.

**Commit:** `feat(kan-73-install-guard-scripts-alongside-skills): ship each command skill's guards beside it`

---

### 3 One resolution rule, and every invoking call site rewritten to it

**Build:** green

**Files:**
- Modify: `skills/myflow-contracts/pipeline.md`
- Modify: `skills/myflow-contracts/finish-contract.md`
- Modify: `skills/myflow-do/SKILL.md`
- Modify: `skills/myflow-finish/SKILL.md`

`skills/myflow-status/SKILL.md`, `skills/myflow-fast/SKILL.md` and
`scripts/check-contract-budget.sh` were expected here when this plan was written and turned out not
to need touching: `myflow-fast` carries no guard citation of its own and delegates entirely by
citation, `myflow-status`'s single citation was already non-invoking prose, and `pipeline.md`'s
growth stayed inside its existing budget row. They are struck from the list rather than left in it,
because `scripts/check-task-commit-fields.sh` checks that the commit touched nothing it did not
declare — not that it touched everything it did — so an over-broad declaration is a claim nothing
would have caught.

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: task 2's directories, which are what the rule resolves against.
- Produces: the basename citation form task 5's guard checks for.

- [x] **Step 1: State the rule once, in `pipeline.md`**

Add one short section. It states that a named guard resolves to `<the running command's own skill
directory>/scripts/<name>`; that a skill or contract names an invoked guard by **basename** and
never by a repository-relative path; that resolution against the *running command's* directory is
what lets a contract loaded by several commands name a guard at all; and that
`skills/myflow-contracts/` is never a running command and carries no `scripts/` directory.

It also states the boundary that keeps this from over-reaching: **prose describing this
repository's own lint and test guards is not an invocation** and keeps its repository-relative path,
because there it names a file in this repository rather than a guard a command runs.

- [x] **Step 2: Rewrite the invoking call sites**

Only invoking sites. These are the ones a command runs:

| File | Sites |
|---|---|
| `skills/myflow-do/SKILL.md` | `plan-dispatch-bundles.sh`, `check-task-commit-fields.sh`, `check-panel-diff-size.sh`, `check-panel-reproducers.sh`, `run-reproducer.sh`, `prepare-workspace.sh`, `preserve-session-records.sh`, `commit-split.sh` |
| `skills/myflow-contracts/finish-contract.md` | `check-finish-preflight.sh`, `check-unfinished-work.sh`, `preserve-session-records.sh`, `check-cleanup-complete.sh`, `gather-self-review-context.sh` |
| `skills/myflow-finish/SKILL.md`, `skills/myflow-status/SKILL.md`, `skills/myflow-fast/SKILL.md` | whichever of the above each names |

`skills/myflow-do/SKILL.md:746`'s `<agents repo>/scripts/prepare-workspace.sh` is the same defect
wearing a different disguise — an explicit escape hatch written because the bare path did not
resolve. It becomes the basename form like every other site.

Leave every **describing** citation alone: `check-references.sh`, `check-stage-mark-calls.sh`,
`check-vocabulary.sh`, `check-contract-budget.sh`, `check-plan-provenance.sh` and
`check-task-build-green.sh` appear in skill text only as prose about this repository's guards.
<!-- measured: grep -rn -B1 -A1 over each of the six in skills/, classified as prose vs invocation @ 7922242 -->

- [x] **Step 3: Raise `pipeline.md`'s budget row, deliberately**

`pipeline.md` grows by step 1's section, which is what `check-contract-budget.sh`'s ratchet exists
to notice. Raise that one row in its `budgets()` table to the file's new size, and no other row.

```bash unverified:the rule section does not exist yet — run after step 1 to read the new size
wc -c skills/myflow-contracts/pipeline.md
scripts/check-contract-budget.sh
```

- [x] **Step 4: References and vocabulary stay green**

```bash verified:both guards exist and exit 0 at 7922242
scripts/check-references.sh && scripts/check-vocabulary.sh && scripts/check-contract-budget.sh
```

**Tests:** none added by this task — it rewrites prose and adds one contract section, and the
project's own reference and budget guards, run in step 4, are what check it. The permanent check
that no invoking call site regresses to a repository-relative path is added by task 5 and declared
there, by the task whose commit carries that guard.

**Regression:** Reverting this task returns every call site to a repository-relative path, which
resolves only when the project being worked on is this repository — so task 2's directories exist
and nothing addresses them.

**Baseline:** before=0 after=1 statements of the resolution rule in `pipeline.md`.

**Commit:** `refactor(kan-73-install-guard-scripts-alongside-skills): resolve a named guard against the running command's skill`

---

### 4 Each command reports a missing guard once, at the start of its run

**Build:** green

**Files:**
- Modify: `skills/myflow-do/SKILL.md`
- Modify: `skills/myflow-finish/SKILL.md`
- Modify: `skills/myflow-status/SKILL.md`
- Modify: `skills/myflow-fast/SKILL.md`
- Modify: `skills/myflow-contracts/pipeline.md`
- Modify: `skills/myflow-contracts/handoff-blocks.md`
- Modify: `scripts/check-contract-budget.sh`

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: task 3's resolution rule, which is what tells the check where to look.
- Produces: the one-block report shape the handoff refers to.

The shape is stated once, in `pipeline.md`, beside the resolution rule, and each command's first
stage cites it rather than restating it:

```text unverified:the block does not exist yet — this is the shape the operator chose during brainstorming
⚠ GUARDS MISSING — 3 of 6 not found at
  <skill-dir>/scripts/
    check-finish-preflight.sh
    check-unfinished-work.sh
    preserve-session-records.sh
These checks will be performed BY HAND.
Re-run ./setup.sh global to install them.
```

- [x] **Step 1: State the check once, in `pipeline.md`**

One paragraph: each command checks, once at the start of its run, that every guard it can invoke is
present in its own `scripts/` directory; a complete set prints nothing; any absence prints exactly
one block naming every missing guard, the directory searched, and the install command, then the run
continues.

State plainly what it is not: **a report, never a gate.** Each contract's existing hand-run fallback
still governs the call site, and the handoff says those checks were run manually. A guard is never
skipped for want of the script — this changes when absence is noticed, not what happens next.

- [x] **Step 2: Add the check to each command skill's first stage**

`/myflow-do` at **1. Load context and validate the plan**; `/myflow-finish` at its preflight;
`/myflow-status` at the start of its report; `/myflow-fast` at its state gate. Each names the guards
*it* can invoke — task 2's map — and cites `pipeline.md` for the block shape.

- [x] **Step 3: The handoff says a check was run by hand**

Where a command's handoff already carries a Panel or Staged line, add the manual-check note to the
same block, so the operator reads it where they read everything else about the run. Do not invent a
second handoff shape.

- [x] **Step 4: Lint stays green**

```bash verified:all three guards exist and exit 0 at 7922242, with pipeline.md's budget row raised by task 3
scripts/check-references.sh && scripts/check-vocabulary.sh && scripts/check-contract-budget.sh
```

`pipeline.md` grows again here; if its raised row no longer covers it, raise that one row again
rather than trimming the section.

**Tests:** none added by this task — it adds one contract section and a paragraph to each command
skill, and the project's own reference guard, run in step 4, is what checks the new
cross-references. The permanent check that each command's declared guard list matches what it can
invoke is added by task 5 and declared there.

**Regression:** Reverting this task returns absence to being discovered silently at each call site —
the half of KAN-73 that the symlinks alone do not fix, since an install can be stale or partial.

**Baseline:** before=0 after=4 command skills carrying a start-of-run presence check.

**Commit:** `feat(kan-73-install-guard-scripts-alongside-skills): report a missing guard once at the start of a run`

---

### 5 `check-guard-symlinks.sh` — the guard that keeps tasks 1–3 true

**Build:** green

**Files:**
- Create: `scripts/check-guard-symlinks.sh`
- Create: `scripts/test-check-guard-symlinks.sh`
- Modify: `.myflow/project.md`

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: tasks 1, 2 and 3 — it lands last precisely so it lands green.
- Produces: a lint step that fails on the reintroduction of either defect.

Four rules, each reported by name with the offending path and line:

1. Every entry under `skills/*/scripts/` is a symlink, and it resolves.
2. Every guard **invoked** in a skill's text has a symlink in that skill's own `scripts/`
   directory — sibling dependencies included. Task 2's map, checked rather than trusted.
3. No skill text carries a repository-relative `scripts/…` path **in an invoking position**. Prose
   about this repository's own guards is exempt, per task 3's boundary.
4. No shipped guard derives a root as a fixed number of levels above `$SCRIPT_DIR`.

- [x] **Step 1: The harness first, with a fixture per rule**

`scripts/test-check-guard-symlinks.sh`, following the fixture-tree shape
`scripts/test-check-workspace-isolation.sh` already uses. One passing fixture, then one violating
fixture per rule above, each asserting the guard exits 1 **and** names the offending path — a guard
that fails without saying what failed is a guard nobody can act on.

Add the two cases every guard in this repository carries: an unreadable tree exits 2 with no
verdict line, and a clean tree exits 0 with one.

- [x] **Step 2: The guard**

Distinguishing rule 3's invoking position from prose is the whole difficulty. Classify by the
line's own shape — a path inside a fenced command block, or on a line whose imperative names running
it, is an invocation; a path inside a sentence about what a guard checks is prose. Where the
classifier cannot tell, **report it** rather than pass it: an unclassifiable line is a line a reader
cannot classify either.

Carry the three disciplines `scripts/lib/panel-record.sh`'s header states for every guard in this
repository — `-a` on every `grep`, the `rc > 1` split between "no match" and a real error, and `--`
before every path — and state in the header that they were adopted from there rather than invented.

- [x] **Step 3: Wire it into the project's own lint and test lists**

Add `scripts/check-guard-symlinks.sh` to `.myflow/project.md`'s `## lint` list and
`scripts/test-check-guard-symlinks.sh` to its `## test` list. Add its budget row to
`scripts/check-contract-budget.sh` only if the guard's own path falls under that guard's scope;
`scripts/` does not, so expect no row.

- [x] **Step 4: Green against the real tree**

```bash unverified:the guard does not exist yet — it must pass against this repository once tasks 1-3 have landed
scripts/test-check-guard-symlinks.sh && scripts/check-guard-symlinks.sh
```

A failure here against the real tree is a defect in tasks 1–3, not a reason to narrow the guard.

**Tests:** `scripts/test-check-guard-symlinks.sh` — the passing fixture, one violating fixture per
rule, the unreadable-tree case and the clean-tree case.

**Regression:** Reverting this task leaves nothing checking that a new guard is symlinked into the
skill that invokes it, that a call site has not reverted to a repository-relative path, or that a
guard has not reintroduced a fixed-depth root — all three of which are silent failures.

**Baseline:** before=0 after=1 guard checking the symlink farms, and before=0 after=1 harness
covering it.

**Commit:** `feat(kan-73-install-guard-scripts-alongside-skills): add check-guard-symlinks.sh`

---

### 6 `test-setup.sh` proves the guards survive an install

**Build:** green

**Files:**
- Modify: `scripts/test-setup.sh`

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: every earlier task.
- Produces: the assertion that would have caught the original bug.

Task 5's guard checks the *repository*. This task checks the *install*, and only this one would
have caught KAN-73: the symlinks could be perfect in the repository and still not reach
`~/.claude/skills/` if `install_skills()` ever stopped carrying them.

- [x] **Step 1: Assert each guard is executable through the installed skill path**

Extend the existing sandbox-`HOME` install `scripts/test-setup.sh` already performs. For each of the
four skills carrying a `scripts/` directory, assert every guard in it is present and executable at
`$SANDBOX/.claude/skills/<skill>/scripts/<name>`.

Assert it for `.cursor` and `.codex` too — `install_skills` is called for all three, and a
regression that reached only one of them is exactly the kind this test exists to catch.

- [x] **Step 2: Assert a guard actually runs, and finds its `lib/`**

Presence is not reachability. Invoke `check-unfinished-work.sh` through the installed path and
assert its output does **not** name a missing library. It is the guard with the hardest dependency,
so it is the one worth running.

- [x] **Step 3: The whole harness green**

```bash verified:the harness reports "✓ PASS — 208 assertions, 0 failures" at 7922242; the count rises by the assertions steps 1 and 2 add
scripts/test-setup.sh
```

**Tests:** the installed-path presence assertions across `.claude`, `.cursor` and `.codex`, and the
installed-path `check-unfinished-work.sh` invocation.

**Regression:** Reverting this task removes the only check that guards reach an installed harness at
all. Every repository-level check in task 5 would still pass while `install_skills()` shipped
nothing — which is precisely the state KAN-73 describes.

**Baseline:** before=208 after≥228 assertions in `scripts/test-setup.sh`.
<!-- measured: scripts/test-setup.sh @ 7922242 printed "✓ PASS — 208 assertions, 0 failures"; the lower bound is 4 skills × 3 harnesses plus the invocation, and the exact figure is whatever step 1's loop adds -->

**Commit:** `test(kan-73-install-guard-scripts-alongside-skills): assert guards reach every installed harness`

---

### 7 Carry the `/myflow-fast` check-4 override onto this branch

**Build:** green

**Files:**
- Modify: `skills/myflow-fast/SKILL.md`
- Modify: `skills/myflow-contracts/finish-contract.md`
- Modify: `scripts/check-contract-budget.sh`

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: nothing from earlier tasks — the prose is already written and is carried verbatim.
- Produces: a main checkout whose `scripts/check-contract-budget.sh` exits clean again.

**Added at the operator's instruction, mid-run, and unrelated to KAN-73's subject.** Two files in
the main checkout carried 21 uncommitted lines when this change was planned: the `/myflow-fast`
check-4 worktree-cleanup override in `skills/myflow-fast/SKILL.md`, and the paragraph in
`skills/myflow-contracts/finish-contract.md` naming that override from the contract side so the two
files cannot silently disagree about a step that destroys files.

They have no change of their own to land through, and they leave `scripts/check-contract-budget.sh`
failing in the main checkout, so they are carried here. **The prose is carried verbatim** — this
task authors none of it and rewrites none of it.

The captured patch is at
`/private/tmp/claude-501/-Users-tweety53-Projects-agents/3fdab6f1-4045-4ff3-871d-e7c38ac96897/scratchpad/kan-73-carried-fast-check4-override.patch`,
with a second copy at `~/kan-73-carried-fast-check4-override.patch.bak`.

- [x] **Step 1: Apply the captured patch to the worktree**

Apply it verbatim. Do not reword, reflow, or "improve" a line of it — it is the operator's own text,
and this task's whole job is to move it, not to edit it.

```bash verified:the patch was captured with `git diff` from the main checkout on 2026-08-17 and `git apply --check --reverse` confirmed it matches that working tree exactly
git -C <worktree> apply /private/tmp/claude-501/-Users-tweety53-Projects-agents/3fdab6f1-4045-4ff3-871d-e7c38ac96897/scratchpad/kan-73-carried-fast-check4-override.patch
git -C <worktree> diff --stat
```

Expect exactly `skills/myflow-contracts/finish-contract.md | 8 ++++++++` and
`skills/myflow-fast/SKILL.md | 13 +++++++++++++`.
<!-- measured: git diff --stat against the main checkout's working tree on 2026-08-17 -->

- [x] **Step 2: Raise the one budget row the added prose overruns**

`scripts/check-contract-budget.sh:90` is the row to raise, and the convention `.myflow/project.md`
states for every row is the file's size plus 25%.

**Measure the size after every edit in this bundle has landed; do not use a precomputed figure.**
Tasks 3 and 4 both add text to `skills/myflow-fast/SKILL.md` as well, so a number computed before
they ran is stale by construction. Run this task last in its bundle, then:

```bash unverified:the final size is not knowable until tasks 3, 4 and this task's step 1 have all edited the file
python3 -c "import os;s=os.path.getsize('skills/myflow-fast/SKILL.md');print(s, int(s*1.25))"
```

For reference only, the operator's prose alone takes the file from 17915 bytes at `HEAD` to 19128,
already over its declared 18225 — which is why the row needs raising at all.
<!-- measured: wc -c on the file at HEAD and with the captured patch applied, on 2026-08-17 -->

Raise **that row only**. Do not touch another row, and do not narrow the guard's scope.

- [x] **Step 3: The budget guard is green again**

```bash verified:the guard exists and exits 1 against the patched content at 18225; it must exit 0 once the row is raised
scripts/check-contract-budget.sh
scripts/check-references.sh && scripts/check-vocabulary.sh
```

**Tests:** none added — this task carries prose and one budget integer, and
`scripts/check-contract-budget.sh` is itself the check that it is correct. A test asserting a
specific budget number would restate the number rather than verify it.

**Regression:** Reverting this task returns `scripts/check-contract-budget.sh` to failing whenever
the operator's uncommitted prose is present, and removes the contract-side paragraph that stops
`finish-contract.md` and `myflow-fast/SKILL.md` disagreeing about whether check 4 asks before
`--force`-removing a worktree.

**Baseline:** before=1 after=0 files over budget, with the operator's prose applied.
<!-- measured: scripts/check-contract-budget.sh against the main checkout's working tree on 2026-08-17 reported exactly one file over budget, skills/myflow-fast/SKILL.md -->

**Commit:** `docs(kan-73-install-guard-scripts-alongside-skills): carry the /myflow-fast check-4 override`
