# kan-798-flow-fix-protect-main-checkout-py-false

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

Three fixes to `hooks/protect-main-checkout.py`, one per false-positive mode, in dependency order:
per-line scanning restructures how the command reaches the tokenizer (task 1), variable expansion
is a pre-pass over the tokens task 1 produces (task 2), and the future-path rule changes how a
resolved path is judged (task 3). Each task extends `scripts/test-protect-main-checkout.sh` with
the cases pinning its own behavior; `design.md` is canonical for every decision.

**No live-verification task: the change touches a PreToolUse hook, not a running service or
persistent state — the harness run below against a scratch origin/clone is the whole runtime.**

---

- [x] 1. Scan bash commands per logical line (mode 3)

`bash_hits()` currently tokenizes the whole command at once; newlines are shlex whitespace, so the
`sed`/`tee`/`cp`/`mv`/`rm` argument scans and the redirect targets run across line boundaries and
swallow later lines' arguments. Pre-split the command text into logical lines and scan each line
separately, threading `cd` state (`cur`) across lines — bash treats a newline as a command
separator, not a session reset.

**Files:** `hooks/protect-main-checkout.py`, `scripts/test-protect-main-checkout.sh`
**Tests:** `26 multi-line cp cannot swallow a later line's target`, `27 multi-line second line judged on its own`, `32 cd state threads across lines without a separator`
**Regression:** revert drops the per-line split and both cases fail: 26 re-denies a legitimate
multi-line `cp` (the mode-3 false positive KAN-798 observed), 27 misses a real `git reset` smuggled
onto a later line
**Baseline:** before=24 after=26
<!-- measured: grep -c 'expect "' scripts/test-protect-main-checkout.sh @ 4a278320 (merge base, the count BEFORE this change) -->
**Commit:** fix(hooks): scan bash commands per logical line
**After:** none
**Build:** green

**Decision:** newline-pre-split-lines

  - [x] **Step 1: Split.** Add a helper that splits the raw command into logical lines: walk the
    text tracking single/double quote state (backslash escapes honored inside double quotes),
    join a backslash-newline outside quotes into the current line, and cut on a newline outside
    quotes. A quoted newline stays inside its line.
  - [x] **Step 2: Scan per line.** In `bash_hits()`, tokenize each logical line with the existing
    `tokenize()` and run the same scan loop over it, carrying `cur` (the `cd` target) from one
    line into the next; accumulate hits across lines.
  - [x] **Step 3: Cases.** Add harness cases 26 and 27: 26 runs a multi-line command —
    `cp notes.md draft.md`, `echo done`, `git add <MAIN>/README.md` on separate lines, cwd `$ROOT`
    — and expects allow; 27 runs `echo start` then `cd <MAIN> && git reset HEAD~1` on separate
    lines and expects deny.
  - [x] **Step 4: Verify.** `scripts/test-protect-main-checkout.sh` passes (all cases, old and
    new) and `scripts/check-python-suppressions.sh` is clean.

- [x] 2. Expand command-text assignments before path resolution (mode 1)

A redirect into the main checkout through a variable set in the same command
(`C=<MAIN>; echo x > $C/f.txt`) sails past: `resolve()` lets every `$` path through. Collect the
command's own assignments and expand them before resolution, keeping the let-through for anything
still unknown.

**Files:** `hooks/protect-main-checkout.py`, `scripts/test-protect-main-checkout.sh`
**Tests:** `28 assignment-set redirect into main resolves and denies`, `29 nested variable stays let-through`, `33 longest-name expansion wins the collision`, `34 braced expansion resolves the variable`, `36 unset near-name variable is not expanded greedily`, `37 use before set is not expanded`
**Regression:** revert drops expansion and both cases fail: 28 stops denying a redirect that lands
in the main checkout through `C`, 29 still passes but only by accident of the same hole
**Baseline:** before=26 after=28
<!-- measured: grep -c 'expect "' scripts/test-protect-main-checkout.sh @ branch spectre/kan-798-flow-fix-protect-main-checkout-py-false -->
**Commit:** fix(hooks): expand command-text assignments before path resolution
**After:** Task 1
**Build:** green

**Decision:** expand-same-string-assignments

  - [x] **Step 1: Collect.** Before scanning, map every token of the form `NAME=value` (whole
    token, `[A-Za-z_][A-Za-z0-9_]*` name) into an env dict; values are taken literally, no nested
    expansion. Document the approximation in the module docstring beside the existing token-scan
    note.
  - [x] **Step 2: Expand.** Longest-name-first `$NAME` and `${NAME}` replacement over every token
    before it is used as a path; a token still containing `$` or a backtick resolves exactly as
    today (let-through).
  - [x] **Step 3: Cases.** Add harness cases 28 and 29: 28 runs `C=<MAIN>; echo x > $C/f.txt`
    (cwd `$ROOT`) and expects deny; 29 runs `B=<MAIN>; A=$B; echo x > $A/f.txt` and expects allow.
    Case 24 (worktree var redirect) must still allow, unchanged.
  - [x] **Step 4: Verify.** `scripts/test-protect-main-checkout.sh` passes and
    `scripts/check-python-suppressions.sh` is clean.

- [x] 3. Treat future `.worktrees` paths as outside the main checkout (mode 2)

`cp /tmp/a.md <MAIN>/.worktrees/new-landing/spectre/foo.md` is denied: the destination does not
exist, `existing_dir()` walks up to `<MAIN>/.worktrees`, and git resolves that directory under the
main checkout on `main`. When the nearest existing ancestor is a directory named `.worktrees`, the
path is future worktree content — outside the protected tree.

**Files:** `hooks/protect-main-checkout.py`, `scripts/test-protect-main-checkout.sh`
**Tests:** `30 cp into a not-yet-existing worktree path`, `31 loose file directly in .worktrees`, `35 rm of the .worktrees root itself stays denied`
**Regression:** revert drops the rule and both cases fail: 30 re-denies the landing-worktree write
(the mode-2 false positive KAN-798 observed), 31 denies a harmless write into the gitignored
worktree root
**Baseline:** before=28 after=30
<!-- measured: grep -c 'expect "' scripts/test-protect-main-checkout.sh @ branch spectre/kan-798-flow-fix-protect-main-checkout-py-false -->
**Commit:** fix(hooks): treat future .worktrees paths as outside the main checkout
**After:** Task 1, 2
**Build:** green

**Decision:** future-paths-narrow-worktrees-rule

  - [x] **Step 1: Rule.** In `protected()` (or its caller), when the nearest existing ancestor
    `existing_dir()` returned has basename `.worktrees`, return no protection — the path is future
    worktree content. Every other not-yet-existing path keeps today's walk-up: a new file whose
    nearest existing ancestor is `<MAIN>` itself still denies.
  - [x] **Step 2: Cases.** Add harness cases 30 and 31: 30 runs the `cp` above (cwd `$ROOT`) and
    expects allow; 31 is a `Write` of `<MAIN>/.worktrees/loose.txt` expecting allow. Case 1 (Write
    `<MAIN>/new.txt` → deny) must still deny, unchanged.
  - [x] **Step 3: Verify.** `scripts/test-protect-main-checkout.sh` passes and
    `scripts/check-python-suppressions.sh` is clean.
