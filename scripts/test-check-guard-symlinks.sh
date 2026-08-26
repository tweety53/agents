#!/usr/bin/env bash
# Assertion harness for check-guard-symlinks.sh. Builds throwaway repository
# trees under a sandboxed TMPDIR — each with its own scripts/ and skills/
# directories — and asserts the guard's violation lines, its verdict and its
# exit status. Never modifies the real repository tree, with one deliberate
# read: the last case runs the guard against this repository's own real path,
# because tasks 1-3 already landed and the guard must exit 0 against them —
# a fixture copy would be a second place for that agreement to drift.
#
# READ THIS BEFORE ADDING OR "FIXING" A CASE. Assert against the four rules
# stated in the frozen openspec/changes/archive/
# 2026-08-18-kan-73-install-guard-scripts-alongside-skills/tasks.md's
# task 5 and design.md's "The guard-to-skill map" / "The
# $SCRIPT_DIR/.. hazard" — never against observed output.
#
# Bash 3.2 is the floor, as test-check-finish-preflight.sh's header records:
# indexed arrays only, no associative arrays.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
GUARD="$SCRIPT_DIR/check-guard-symlinks.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }
skip() { printf 'skip: %s (%s)\n' "$1" "$2"; }

# An indexed array, not a space-separated string: sandbox paths come from
# mktemp under TMPDIR, which may contain spaces, and word-splitting a string
# would then rm -rf the fragments.
SANDBOXES=()
cleanup() {
  [ "${#SANDBOXES[@]}" -eq 0 ] && return 0
  for s in "${SANDBOXES[@]}"; do rm -rf "$s"; done
}
trap cleanup EXIT

# run_guard <root> -> sets OUT (stdout only), ERR, RC. The two streams are
# captured separately, exactly as test-check-workspace-isolation.sh does,
# because a refusal must put its message on stderr and leave stdout empty —
# a merged capture cannot tell an empty stdout from one carrying the message.
run_guard() {
  local errfile
  errfile="$(mktemp "${TMPDIR:-/tmp}/check-guard-symlinks-stderr.XXXXXX")"
  SANDBOXES+=("$errfile")
  set +e
  OUT="$(CHECK_GUARD_SYMLINKS_ROOT="$1" "$GUARD" 2>"$errfile")"
  RC=$?
  set -e
  ERR="$(cat "$errfile")"
}

# new_repo -> sets REPO to an empty repository root carrying scripts/ and
# skills/, both required for the guard's own directory checks to pass before
# it looks at anything inside them.
new_repo() {
  REPO="$(mktemp -d "${TMPDIR:-/tmp}/check-guard-symlinks-repo.XXXXXX")"
  SANDBOXES+=("$REPO")
  mkdir -p "$REPO/scripts" "$REPO/skills"
}

# add_real_guard <name> <body> -> writes $REPO/scripts/<name>, executable.
add_real_guard() {
  printf '%s\n' "$2" > "$REPO/scripts/$1"
  chmod +x "$REPO/scripts/$1"
}

# link_guard <skill> <name> -> a correct relative symlink from
# skills/<skill>/scripts/<name> into ../../../scripts/<name>, exactly the
# shape task 2 committed.
link_guard() {
  mkdir -p "$REPO/skills/$1/scripts"
  ln -s "../../../scripts/$2" "$REPO/skills/$1/scripts/$2"
}

# write_skill_md <skill> <body> -> $REPO/skills/<skill>/SKILL.md.
write_skill_md() {
  mkdir -p "$REPO/skills/$1"
  printf '%s\n' "$2" > "$REPO/skills/$1/SKILL.md"
}

PLAIN_GUARD_BODY='#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
echo ok'

FIXED_DEPTH_GUARD_BODY='#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
echo "$REPO_ROOT"'

# F5 — the same fixed-depth defect, spelled two other ways. dirname() and a
# two-step cd chain answer the identical "one level above $SCRIPT_DIR"
# question the literal `$SCRIPT_DIR/..` form does, and rule 4 must catch all
# three.
FIXED_DEPTH_GUARD_BODY_DIRNAME='#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"
echo "$REPO_ROOT"'

FIXED_DEPTH_GUARD_BODY_CD_CHAIN='#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR" && cd ..
REPO_ROOT="$(pwd)"
echo "$REPO_ROOT"'

# assert_ok <label> — exit 0 and a GUARD-SYMLINKS-OK verdict on the FIRST
# line. First, not last: scripts/lib/coverage.sh's per-skill breakdown, when
# the corpus is non-empty, is an indented second line the verdict now always
# carries on a clean run — checking the first line keeps this assertion
# agnostic to whether that second line is present.
assert_ok() {
  if [ "$RC" -ne 0 ]; then
    fail "$1: expected exit 0, got rc=$RC out=$OUT err=$ERR"
    return 0
  fi
  case "$(printf '%s\n' "$OUT" | head -n 1)" in
    "GUARD-SYMLINKS-OK:"*) pass "$1" ;;
    *) fail "$1: expected a GUARD-SYMLINKS-OK verdict, got: $OUT" ;;
  esac
}

# assert_silent <label> — exit 0, the verdict on line 1, and at most a
# coverage breakdown on line 2 — never a "path:line: message" violation
# mixed in. Every fixture reaching this assertion has a non-empty corpus (at
# least one skill directory), so coverage_report always renders a line; "one
# line" from before this task is now "one or two."
assert_silent() {
  local lines
  assert_ok "$1"
  lines="$(printf '%s\n' "$OUT" | wc -l | tr -d ' ')"
  case "$lines" in
    1|2) pass "$1: verdict plus at most a coverage breakdown" ;;
    *) fail "$1: expected 1 or 2 stdout lines, got $lines: $OUT" ;;
  esac
}

# assert_invalid <label> — exit 1 and a GUARD-SYMLINKS-INVALID verdict.
assert_invalid() {
  if [ "$RC" -ne 1 ]; then
    fail "$1: expected exit 1, got rc=$RC out=$OUT err=$ERR"
    return 0
  fi
  case "$(printf '%s\n' "$OUT" | tail -n 1)" in
    "GUARD-SYMLINKS-INVALID:"*) pass "$1" ;;
    *) fail "$1: expected a GUARD-SYMLINKS-INVALID verdict, got: $OUT" ;;
  esac
}

# assert_reports <needle> <label> — the report names this path or rule.
assert_reports() {
  case "$OUT" in
    *"$1"*) pass "$2" ;;
    *) fail "$2: the report does not name '$1': $OUT" ;;
  esac
}

# assert_refuses <label> — the refusal shape: exit 2, nothing on stdout, and
# the guard's own name on stderr. The needle carries the colon deliberately —
# without it the shell's own "No such file or directory" would satisfy the
# case even if the guard itself never ran.
assert_refuses() {
  [ "$RC" -eq 2 ] && pass "$1: exits 2" \
    || fail "$1: expected exit 2, got rc=$RC out=$OUT"
  [ -z "$OUT" ] && pass "$1: writes nothing to stdout" \
    || fail "$1: emitted a verdict line: $OUT"
  case "$ERR" in
    *"check-guard-symlinks: "*) pass "$1: names the failure on stderr" ;;
    *) fail "$1: no named message on stderr: $ERR" ;;
  esac
}

# ---------------------------------------------------------------------------
# 1. A clean tree — the passing fixture. One real guard with no sibling
#    dependency and no fixed-depth root, invoked with a placeholder argument
#    inside a bash fence, correctly symlinked into the one skill that cites
#    it.
# ---------------------------------------------------------------------------
new_repo
add_real_guard "check-foo.sh" "$PLAIN_GUARD_BODY"
link_guard "myflow-do" "check-foo.sh"
write_skill_md "myflow-do" '# myflow-do fixture

Run the guard:

```bash
check-foo.sh <worktree>
```
'
run_guard "$REPO"
assert_silent "a clean tree passes with exactly one verdict line"

# ---------------------------------------------------------------------------
# 2. Rule 1 — every entry under skills/*/scripts/ is a symlink, and it
#    resolves.
# ---------------------------------------------------------------------------

# 2a. An entry that is a regular file, not a symlink.
new_repo
add_real_guard "check-foo.sh" "$PLAIN_GUARD_BODY"
mkdir -p "$REPO/skills/myflow-do/scripts"
printf 'not a symlink\n' > "$REPO/skills/myflow-do/scripts/check-bogus.sh"
write_skill_md "myflow-do" "# fixture, no citations"
run_guard "$REPO"
assert_invalid "a non-symlink entry under skills/*/scripts/ is a rule 1 violation"
assert_reports "check-bogus.sh" "rule 1: names the offending entry"
assert_reports "rule 1" "rule 1: names the rule"

# 2b. A dangling symlink.
new_repo
mkdir -p "$REPO/skills/myflow-do/scripts"
ln -s "../../../scripts/does-not-exist.sh" "$REPO/skills/myflow-do/scripts/check-missing.sh"
write_skill_md "myflow-do" "# fixture, no citations"
run_guard "$REPO"
assert_invalid "a dangling symlink under skills/*/scripts/ is a rule 1 violation"
assert_reports "check-missing.sh" "rule 1: names the dangling entry"
assert_reports "does not resolve" "rule 1: says it does not resolve"

# 2c. An absolute-target symlink that still resolves.
new_repo
add_real_guard "check-foo.sh" "$PLAIN_GUARD_BODY"
mkdir -p "$REPO/skills/myflow-do/scripts"
ln -s "$REPO/scripts/check-foo.sh" "$REPO/skills/myflow-do/scripts/check-foo.sh"
write_skill_md "myflow-do" "# fixture, no citations"
run_guard "$REPO"
assert_invalid "an absolute symlink target under skills/*/scripts/ is a rule 1 violation"
assert_reports "is absolute" "rule 1: says the target is absolute"

# 2d. F9 — an absolute target pointing OUTSIDE the repository entirely must
#     stop at the rule 1 violation, not fall through into REAL_TARGETS_FILE
#     and have rule 4 grep a file outside the repository this guard was
#     asked to scan. The off-repo target here is unreadable, so a fall-
#     through would surface as a SECOND violation ("cannot read this guard's
#     real source") on top of the rule 1 one — exactly one violation line is
#     the assertion that pins the fix.
# Uses "myflow-start" — one of the guard's own declared expected-zero
# skills — rather than "myflow-do": this fixture's own assertion below
# counts violation lines exactly, and an undeclared zero-coverage skill
# would add a second one that has nothing to do with what F9 tests.
new_repo
mkdir -p "$REPO/skills/myflow-start/scripts"
OUTSIDE_TARGET="$(mktemp "${TMPDIR:-/tmp}/check-guard-symlinks-outside.XXXXXX")"
SANDBOXES+=("$OUTSIDE_TARGET")
printf 'unrelated content, unreadable\n' > "$OUTSIDE_TARGET"
chmod 000 "$OUTSIDE_TARGET"
ln -s "$OUTSIDE_TARGET" "$REPO/skills/myflow-start/scripts/check-outside.sh"
write_skill_md "myflow-start" "# fixture, no citations"
run_guard "$REPO"
assert_invalid "an absolute off-repo symlink target is a rule 1 violation"
assert_reports "is absolute" "rule 1 (F9): says the target is absolute"
if [ "$(id -u)" = "0" ]; then
  skip "F9: exactly one violation line, not a second from rule 4 reading outside the repo" "running as root; mode 000 is still readable"
else
  N_VIOL_LINES="$(printf '%s\n' "$OUT" | grep -vc '^GUARD-SYMLINKS-')"
  [ "$N_VIOL_LINES" = "1" ] && pass "F9: exactly one violation line, not a second from rule 4 reading outside the repo" \
    || fail "F9: expected exactly 1 violation line, got $N_VIOL_LINES: $OUT"
fi
chmod 644 "$OUTSIDE_TARGET"

# ---------------------------------------------------------------------------
# 3. Rule 2 — every guard invoked in a skill's own text has a symlink in that
#    skill's own scripts/ directory.
# ---------------------------------------------------------------------------

# 3a. Invoked in a bash fence, never symlinked in at all.
new_repo
add_real_guard "check-baz.sh" "$PLAIN_GUARD_BODY"
mkdir -p "$REPO/skills/myflow-do/scripts"
write_skill_md "myflow-do" '# myflow-do fixture

```bash
check-baz.sh <worktree>
```
'
run_guard "$REPO"
assert_invalid "an invoked guard with no symlink is a rule 2 violation"
assert_reports "check-baz.sh" "rule 2: names the missing guard"
assert_reports "rule 2" "rule 2: names the rule"

# 3b. A sibling dependency, resolved from the required guard's own source,
#     missing from the skill that needs the guard beside it.
new_repo
add_real_guard "check-with-lib.sh" '#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/helper.sh"'
mkdir -p "$REPO/scripts/lib"
printf 'true\n' > "$REPO/scripts/lib/helper.sh"
link_guard "myflow-do" "check-with-lib.sh"
write_skill_md "myflow-do" '# myflow-do fixture

```bash
check-with-lib.sh <worktree>
```
'
run_guard "$REPO"
assert_invalid "a required sibling with no symlink is a rule 2 violation"
assert_reports "lib" "rule 2: names the missing sibling"
assert_reports "sibling dependency" "rule 2: says it is a sibling dependency"

# 3c. F6 — a fence scanner reading only a line's leading token misses two
#     real invocation shapes: `./guard.sh <args>` (the leading token stops at
#     the `.`) and `bash guard.sh <args>` / `sh guard.sh` / `zsh guard.sh`
#     (the guard is the SECOND token, the interpreter the first). Both are
#     never-symlinked here, so both must be caught as rule 2 violations.
new_repo
add_real_guard "check-dotslash.sh" "$PLAIN_GUARD_BODY"
add_real_guard "check-viabash.sh" "$PLAIN_GUARD_BODY"
mkdir -p "$REPO/skills/myflow-do/scripts"
write_skill_md "myflow-do" '# myflow-do fixture

```bash
./check-dotslash.sh <worktree>
```

```bash
bash check-viabash.sh <worktree>
```
'
run_guard "$REPO"
assert_invalid "a ./guard.sh or bash guard.sh invocation with no symlink is a rule 2 violation"
assert_reports "check-dotslash.sh" "rule 2 (F6): names the ./guard.sh form"
assert_reports "check-viabash.sh" "rule 2 (F6): names the bash guard.sh form"

# 3d. F7 — positional backtick pairing. A stray, unpaired backtick earlier in
#     the SAME line must not cause the real citation later on the same line
#     to be dropped. `check-strayed.sh` is never symlinked, so if the
#     citation below it is read at all, it is a rule 2 violation; before the
#     fix, the stray backtick consumed the pairing and this citation was
#     silently never seen.
new_repo
add_real_guard "check-strayed.sh" "$PLAIN_GUARD_BODY"
mkdir -p "$REPO/skills/myflow-do/scripts"
write_skill_md "myflow-do" '# myflow-do fixture

A stray ` mark appears here, then Run `check-strayed.sh` for the real thing.
'
run_guard "$REPO"
assert_invalid "a real citation after a stray backtick on the same line is still a rule 2 violation"
assert_reports "check-strayed.sh" "rule 2 (F7): the citation after the stray backtick was not dropped"

# 3e. F10 (dedup regression) — two required guards on the SAME skill sharing
#     ONE sibling dependency. This does not trigger the rc>=2 refusal (not
#     externally reproducible without white-box access to this guard's own
#     scratch directory — see the report), but it does exercise the
#     "already seen" grep on its happy path twice over, which the discipline
#     fix must not break: the sibling is required exactly once, and BOTH
#     citing guards are still individually enforced.
new_repo
add_real_guard "check-one.sh" '#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/shared.sh"'
add_real_guard "check-two.sh" '#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/shared.sh"'
mkdir -p "$REPO/scripts/lib"
printf 'true\n' > "$REPO/scripts/lib/shared.sh"
link_guard "myflow-do" "check-one.sh"
link_guard "myflow-do" "check-two.sh"
write_skill_md "myflow-do" '# myflow-do fixture

```bash
check-one.sh <worktree>
check-two.sh <worktree>
```
'
run_guard "$REPO"
assert_invalid "one sibling shared by two required guards is still a rule 2 violation"
N_SHARED_LIB="$(printf '%s\n' "$OUT" | grep -c 'sibling dependency' || true)"
[ "$N_SHARED_LIB" = "1" ] && pass "F10: the shared sibling is reported exactly once, not once per citing guard" \
  || fail "F10: expected exactly 1 sibling-dependency violation, got $N_SHARED_LIB: $OUT"

# ---------------------------------------------------------------------------
# 4. Rule 3 — no skill text carries a repository-relative scripts/<name> path
#    in an invoking position. Prose is exempt.
# ---------------------------------------------------------------------------

# 4a. The repository-relative form inside a bash fence — an invocation.
new_repo
add_real_guard "check-qux.sh" "$PLAIN_GUARD_BODY"
link_guard "myflow-do" "check-qux.sh"
write_skill_md "myflow-do" '# myflow-do fixture

```bash
scripts/check-qux.sh <worktree>
```
'
run_guard "$REPO"
assert_invalid "a repository-relative path inside a bash fence is a rule 3 violation"
assert_reports "rule 3" "rule 3: names the rule"

# 4b. The SAME guard, named with its repository-relative path in a plain
#     descriptive sentence — never an invocation. Rule 3's own boundary: this
#     must NOT be reported, or the classifier could not tell prose from a
#     call site at all.
new_repo
add_real_guard "check-qux.sh" "$PLAIN_GUARD_BODY"
link_guard "myflow-do" "check-qux.sh"
write_skill_md "myflow-do" '# myflow-do fixture

Run the guard:

```bash
check-qux.sh <worktree>
```

`scripts/check-qux.sh` reads only the marker block. It never parses the table.
'
run_guard "$REPO"
assert_silent "a repository-relative path in a descriptive sentence is prose, not a rule 3 violation"

# 4c. F3 — the six project-configured guards named in design.md are exempt
#     from rule 3 wherever they appear, in a fence or in imperative prose,
#     since they resolve through the project's own .flow/project.md rather
#     than being invoked by any command. A future "Run
#     `scripts/check-vocabulary.sh` before committing" must not fail CI
#     wrongly.
# Uses "myflow-start" (a declared expected-zero skill): neither citation
# below is in a form rule 2's classifier reads as a required guard, so this
# skill's required set is genuinely empty — declaring it here keeps this
# fixture about rule 3's exemption, not coverage.
new_repo
write_skill_md "myflow-start" '# myflow-start fixture

Run `scripts/check-vocabulary.sh` before committing.

```bash
scripts/check-references.sh
```
'
run_guard "$REPO"
assert_silent "a project-configured guard keeps its repository-relative path without tripping rule 3 (F3)"

# 4d. F8 — the fence-language test is anchored to a full word. A
#     ```shellsession fence must not be scanned as though it were bash/sh/zsh
#     — before the fix, "shellsession" matched the unanchored `^sh` prefix by
#     accident, and a repository-relative citation of a NON-exempt guard
#     inside it was wrongly flagged.
# Uses "myflow-start" (declared expected-zero): a citation inside a
# ```shellsession fence is never scanned by rule 2, so this skill's required
# set is genuinely empty here too — declaring it keeps F8 about the fence
# language anchor, not coverage.
new_repo
write_skill_md "myflow-start" '# myflow-start fixture

```shellsession
scripts/check-qux.sh <worktree>
```
'
run_guard "$REPO"
assert_silent "a \`\`\`shellsession fence is not scanned as bash/sh/zsh (F8)"

# ---------------------------------------------------------------------------
# 5. Rule 4 — no shipped guard derives a repository root as a fixed number of
#    levels above $SCRIPT_DIR.
# ---------------------------------------------------------------------------
new_repo
add_real_guard "check-fixed-depth.sh" "$FIXED_DEPTH_GUARD_BODY"
link_guard "myflow-do" "check-fixed-depth.sh"
write_skill_md "myflow-do" "# fixture, no citations"
run_guard "$REPO"
assert_invalid "a shipped guard deriving \$SCRIPT_DIR/.. is a rule 4 violation"
assert_reports "check-fixed-depth.sh" "rule 4: names the offending guard"
assert_reports "rule 4" "rule 4: names the rule"

# A project-configured guard that is NEVER shipped (no symlink anywhere) may
# keep the $SCRIPT_DIR/.. form without tripping rule 4 — the guard scopes
# rule 4 to what rule 1 found actually symlinked in, never every file under
# scripts/. Uses "myflow-start" (declared expected-zero): this fixture cites
# nothing, so its required set is genuinely empty, and declaring it keeps
# this test about rule 4's scoping, not coverage.
new_repo
add_real_guard "check-project-only.sh" "$FIXED_DEPTH_GUARD_BODY"
write_skill_md "myflow-start" "# fixture, no citations"
run_guard "$REPO"
assert_silent "an unshipped guard keeping \$SCRIPT_DIR/.. is not a rule 4 violation"

# 5c. F5 — the dirname() spelling of the identical defect.
new_repo
add_real_guard "check-fixed-dirname.sh" "$FIXED_DEPTH_GUARD_BODY_DIRNAME"
link_guard "myflow-do" "check-fixed-dirname.sh"
write_skill_md "myflow-do" "# fixture, no citations"
run_guard "$REPO"
assert_invalid "a shipped guard deriving dirname(\$SCRIPT_DIR) is a rule 4 violation (F5)"
assert_reports "check-fixed-dirname.sh" "rule 4 (F5): names the offending guard (dirname form)"
assert_reports "rule 4" "rule 4 (F5): names the rule (dirname form)"

# 5d. F5 — the two-step `cd $SCRIPT_DIR && cd ..` spelling of the identical
#     defect.
new_repo
add_real_guard "check-fixed-cdchain.sh" "$FIXED_DEPTH_GUARD_BODY_CD_CHAIN"
link_guard "myflow-do" "check-fixed-cdchain.sh"
write_skill_md "myflow-do" "# fixture, no citations"
run_guard "$REPO"
assert_invalid "a shipped guard deriving cd \$SCRIPT_DIR && cd .. is a rule 4 violation (F5)"
assert_reports "check-fixed-cdchain.sh" "rule 4 (F5): names the offending guard (cd-chain form)"
assert_reports "rule 4" "rule 4 (F5): names the rule (cd-chain form)"

# ---------------------------------------------------------------------------
# F4 — rule 2 for a DELEGATING skill. myflow-fast invokes no guard of its
# own; instead its own "**Check guard presence.**" paragraph names other
# commands' presence checks by slash-command, and its required set must be
# the union of theirs — never silently empty. Reproduces the parent's own
# mutation check: deleting the delegate's guard from the delegating skill's
# scripts/ must be caught.
# ---------------------------------------------------------------------------

# F4a. myflow-do invokes check-deleg.sh directly; the delegating skill's own
#      presence paragraph names /myflow-do by slash-command and carries no
#      symlink for it — a rule 2 violation, reported against myflow-deleg.
new_repo
add_real_guard "check-deleg.sh" "$PLAIN_GUARD_BODY"
link_guard "myflow-do" "check-deleg.sh"
write_skill_md "myflow-do" '# myflow-do fixture

**Check guard presence.** Confirm every guard this command invokes:

```bash
check-deleg.sh <worktree>
```
'
write_skill_md "myflow-deleg" '# myflow-deleg fixture, a myflow-fast stand-in

**Check guard presence.** Confirm every guard named in `/myflow-do` own
presence checks is present in `skills/myflow-deleg/scripts/`.
'
run_guard "$REPO"
assert_invalid "a delegating skill whose delegate requires a guard it does not carry is a rule 2 violation (F4)"
assert_reports "check-deleg.sh" "rule 2 (F4): names the guard required by delegation"
assert_reports "myflow-deleg" "rule 2 (F4): names the delegating skill"

# F4b. Same shape, but the delegating skill DOES carry the symlink — silent.
new_repo
add_real_guard "check-deleg.sh" "$PLAIN_GUARD_BODY"
link_guard "myflow-do" "check-deleg.sh"
link_guard "myflow-deleg" "check-deleg.sh"
write_skill_md "myflow-do" '# myflow-do fixture

**Check guard presence.** Confirm every guard this command invokes:

```bash
check-deleg.sh <worktree>
```
'
write_skill_md "myflow-deleg" '# myflow-deleg fixture, a myflow-fast stand-in

**Check guard presence.** Confirm every guard named in `/myflow-do` own
presence checks is present in `skills/myflow-deleg/scripts/`.
'
run_guard "$REPO"
assert_silent "a delegating skill carrying the delegated guard is not a rule 2 violation (F4)"

# F4c. A skill that merely CITES another command in ordinary prose — not
#      inside a "**Check guard presence.**" paragraph — is never read as
#      delegating. This is the false-positive this file's own header warns
#      about (rule 2 scoped to a skill's own directory, not every citation):
#      myflow-status used to cite /myflow-do constantly without invoking
#      anything, before KAN-236 gave it its own guard, so this fixture now
#      stands in with "myflow-start" instead — one of the guard's own
#      declared expected-zero skills, so its genuinely-empty required set
#      here reads as declared rather than tripping the coverage violation
#      added below (F12) — this fixture is about rule 2's delegation
#      boundary, not coverage.
new_repo
add_real_guard "check-deleg.sh" "$PLAIN_GUARD_BODY"
link_guard "myflow-do" "check-deleg.sh"
write_skill_md "myflow-do" '# myflow-do fixture

**Check guard presence.** Confirm every guard this command invokes:

```bash
check-deleg.sh <worktree>
```
'
write_skill_md "myflow-start" '# fixture standing in for myflow-start

This command explains what `/myflow-do` does elsewhere. It invokes nothing
of its own.
'
run_guard "$REPO"
assert_silent "an ordinary cross-command citation outside the presence paragraph is not delegation (F4)"

# ---------------------------------------------------------------------------
# F12 — the KAN-197 regression case. Reproduces KAN-73's own shape,
# generalized past myflow-fast's now-fixed delegation format: a skill that
# CARRIES A REAL SYMLINK — evidence it needs a guard beside it — but whose
# own text names no guard in a form rule 2's classifier can see (a plain
# prose mention of another command, not the recognized "**Check guard
# presence.**" delegation paragraph). Rule 2 itself has nothing to require,
# so nothing to violate — before this task, that read as silent, exactly the
# defect that let myflow-fast's own missing symlink go undetected through
# three reviewers. After this task, the skill's own empty required set is
# itself the finding: named, and non-zero exit, on an otherwise clean tree.
# ---------------------------------------------------------------------------
new_repo
add_real_guard "check-shadow.sh" "$PLAIN_GUARD_BODY"
link_guard "myflow-do" "check-shadow.sh"
link_guard "myflow-shadow" "check-shadow.sh"
write_skill_md "myflow-do" '# myflow-do fixture

**Check guard presence.** Confirm every guard this command invokes:

```bash
check-shadow.sh <worktree>
```
'
write_skill_md "myflow-shadow" '# myflow-shadow fixture — KANs own shape, generalized

See `/myflow-do` for details on this behavior. This skill delegates, but not
in a form the classifier above recognizes.
'
run_guard "$REPO"
assert_invalid "a skill carrying a real symlink but citing nothing rule 2 can see is a coverage violation, not silence (F12)"
assert_reports "myflow-shadow" "F12: names the under-covered skill"
assert_reports "coverage" "F12: reports it as a coverage finding"
assert_reports "0 checked" "F12: reports the zero count"
assert_reports "not declared expected-zero" "F12: says it was never declared"

# ---------------------------------------------------------------------------
# F13 — the paired case: a member legitimately at zero, and DECLARED in the
# guard's own source, reports the zero without failing. Reuses
# "myflow-start" — one of the three names this guard's own coverage_declare
# calls list — deliberately, the same reuse F9/F3/F8/rule-4's "unshipped
# guard" case and F4c above already rely on.
# ---------------------------------------------------------------------------
new_repo
write_skill_md "myflow-start" "# myflow-start fixture, no citations — declared expected-zero in the guard's own source"
run_guard "$REPO"
assert_silent "a declared expected-zero member reports its zero without failing (F13)"
assert_reports "declared" "F13: the coverage breakdown marks the zero as declared"

# ---------------------------------------------------------------------------
# 6. Inputs this guard cannot answer about. Never fail open.
# ---------------------------------------------------------------------------

# 6a. An unreadable skills/ directory — no verdict line, ever.
if [ "$(id -u)" = "0" ]; then
  skip "an unreadable skills/ directory refuses" "running as root; mode 000 is still readable"
else
  new_repo
  chmod 000 "$REPO/skills"
  run_guard "$REPO"
  assert_refuses "an unreadable skills/ directory"
  chmod 755 "$REPO/skills"
fi

# 6b. A root that is not a directory at all.
NOT_A_DIR="$(mktemp "${TMPDIR:-/tmp}/check-guard-symlinks-notadir.XXXXXX")"
SANDBOXES+=("$NOT_A_DIR")
run_guard "$NOT_A_DIR"
assert_refuses "a root that is not a directory"

# ---------------------------------------------------------------------------
# 6c. scripts/lib/resolve-file.sh's own contract (F9, pass 2 of KAN-201's
#     own review panel) -- a deliberate exception to this file's own header
#     ("assert against the four [guard] rules ... never against observed
#     output"), because this is the one function this guard calls whose
#     misbehaviour cannot be shown by any combination of those four rules:
#     it needs resolve_file called directly, exactly as test-lib-coverage.sh
#     already does for scripts/lib/coverage.sh, the other library this guard
#     sources. This guard reads real filesystem entries via $entry (line
#     250, a path from a directory scan) -- the wider exposure the
#     dispatcher's own verification named -- so it is the natural home for
#     resolve_file's own regression cases, not a guard-behaviour case.
# Sourced directly into this shell, not a subshell: fail() increments the
# top-level FAILURES this file's own exit status reads at the bottom, which
# a subshell's own copy would silently lose.
# shellcheck source=lib/resolve-file.sh
source "$SCRIPT_DIR/lib/resolve-file.sh"

RESOLVE_FILE_DOT_GOT="$(resolve_file ".")"
RESOLVE_FILE_DOT_WANT="$(pwd -P)"
if [ "$RESOLVE_FILE_DOT_GOT" = "$RESOLVE_FILE_DOT_WANT" ]; then
  pass "resolve_file '.' resolves to the cwd itself, no spurious trailing '.' (F9)"
else
  fail "resolve_file '.' = '$RESOLVE_FILE_DOT_GOT', want '$RESOLVE_FILE_DOT_WANT' (F9)"
fi

TRAILING_SLASH_DIR="$(mktemp -d "${TMPDIR:-/tmp}/check-guard-symlinks-trailing.XXXXXX")"
SANDBOXES+=("$TRAILING_SLASH_DIR")
RESOLVE_FILE_NO_SLASH="$(resolve_file "$TRAILING_SLASH_DIR")"
RESOLVE_FILE_WITH_SLASH="$(resolve_file "$TRAILING_SLASH_DIR/")"
if [ "$RESOLVE_FILE_WITH_SLASH" = "$RESOLVE_FILE_NO_SLASH" ] && [ -n "$RESOLVE_FILE_WITH_SLASH" ]; then
  pass "resolve_file with a trailing slash matches the same path without one, leaf not dropped (F9)"
else
  fail "resolve_file with a trailing slash = '$RESOLVE_FILE_WITH_SLASH', want '$RESOLVE_FILE_NO_SLASH' (F9)"
fi

# 6d. Root itself, and a run of slashes that collapses to it (F16, pass 3
#     of KAN-201's own review panel): resolve_file used to answer "//" for
#     both, against this file's own header comment claiming root "is left
#     alone". Neither case touches $TMPDIR, so it is exercised regardless
#     of where the ambient temp directory happens to live.
RESOLVE_FILE_ROOT="$(resolve_file "/")"
if [ "$RESOLVE_FILE_ROOT" = "/" ]; then
  pass "resolve_file '/' resolves to '/', not '//' (F16)"
else
  fail "resolve_file '/' = '$RESOLVE_FILE_ROOT', want '/' (F16)"
fi

RESOLVE_FILE_MULTI_ROOT="$(resolve_file "///")"
if [ "$RESOLVE_FILE_MULTI_ROOT" = "/" ]; then
  pass "resolve_file '///' collapses to '/', not '//' (F16)"
else
  fail "resolve_file '///' = '$RESOLVE_FILE_MULTI_ROOT', want '/' (F16)"
fi

# 6e. A root-parented symlink (F16): a symlink whose own parent directory
#     is "/" -- $TMPDIR deliberately not used here, per this file's own
#     6c note above, since on this machine (and most macOS ones) it lives
#     under /var/folders/..., never under /tmp, and so never exercises
#     this shape. macOS's own /tmp -> private/tmp is exactly this case,
#     used directly rather than built, since a non-root process cannot
#     create a symlink inside "/" itself to test the general case. Skipped
#     on a platform where /tmp is a real directory (e.g. most Linux), where
#     this shape does not arise.
if [ -L /tmp ]; then
  RESOLVE_FILE_TMP="$(resolve_file /tmp)"
  RESOLVE_FILE_TMP_WANT="$(cd -P -- /tmp 2>/dev/null && pwd -P)"
  case "$RESOLVE_FILE_TMP" in
    //*)
      fail "resolve_file /tmp = '$RESOLVE_FILE_TMP', a root-parented symlink doubled its leading slash (F16)"
      ;;
    *)
      if [ "$RESOLVE_FILE_TMP" = "$RESOLVE_FILE_TMP_WANT" ]; then
        pass "resolve_file /tmp (a root-parented symlink) resolves without a doubled leading slash (F16)"
      else
        fail "resolve_file /tmp = '$RESOLVE_FILE_TMP', want '$RESOLVE_FILE_TMP_WANT' (F16)"
      fi
      ;;
  esac
else
  skip "a root-parented symlink (/tmp)" "/tmp is not a symlink on this platform"
fi

# ---------------------------------------------------------------------------
# 7. The agents repository itself, read from its real path. Tasks 1-3 have
#    already landed, so this guard must exit 0 against them — a failure here
#    is a defect in tasks 1-3 or in this guard, never a reason to narrow it.
# ---------------------------------------------------------------------------
run_guard "$REPO_ROOT"
assert_ok "the agents repository's own tree validates cleanly"

# ---------------------------------------------------------------------------
# ---------------------------------------------------------------------------
# 8. Rule 5 — no skill directory carries a symlink directly at its own top
#    level (KAN-202 task 9). skills/myflow-do/SKILL.md's [PRINCIPLES_PATH]
#    paragraph already says in prose that symlinking a copy in is the wrong
#    fix for a file a skill's text names but does not carry; a session did
#    exactly that anyway roughly five hours after that sentence was
#    committed, in skills/myflow-fast/. Rule 1 validates entries UNDER
#    skills/*/scripts/ and reaches nothing at a skill directory's own top
#    level, so this rule closes that gap.
#
#    Both fixture skills below are named "myflow-start" and
#    "myflow-research" — the guard's own two declared-expected-zero
#    members — so neither fixture's genuinely-empty required set trips the
#    unrelated KAN-197 coverage check; these cases are about rule 5 alone.
# ---------------------------------------------------------------------------

# 8a. The violation: a symlink directly under a skill directory, pointing at
#     a file in a sibling skill, with no other defect anywhere in the tree.
new_repo
mkdir -p "$REPO/skills/myflow-research"
printf 'principles\n' > "$REPO/skills/myflow-research/engineering-principles.md"
write_skill_md "myflow-research" "# fixture, no citations"
write_skill_md "myflow-start" "# fixture, no citations"
ln -s "../myflow-research/engineering-principles.md" "$REPO/skills/myflow-start/engineering-principles.md"
run_guard "$REPO"
assert_invalid "a symlink directly under a skill directory's top level is a rule 5 violation"
if printf '%s\n' "$OUT" | grep -q "engineering-principles.md" \
  && printf '%s\n' "$OUT" | grep -q "../myflow-research/engineering-principles.md"; then
  pass "rule 5: the report names both the symlink's path and its target"
else
  fail "rule 5: the report does not name both the symlink's path and its target: $OUT"
fi

# 8b. Acceptance: a symlink under skills/<skill>/scripts/ is rule 1's
#     territory, not rule 5's — it must not be re-reported here.
new_repo
add_real_guard "check-foo.sh" "$PLAIN_GUARD_BODY"
link_guard "myflow-start" "check-foo.sh"
write_skill_md "myflow-start" "# fixture, no citations"
run_guard "$REPO"
assert_silent "a symlink under skills/<skill>/scripts/ is not a rule 5 violation"

# 8c. Acceptance: a skill directory holding only regular files and
#     directories carries no rule 5 finding.
new_repo
mkdir -p "$REPO/skills/myflow-start"
write_skill_md "myflow-start" "# fixture, no citations"
run_guard "$REPO"
assert_silent "a skill directory with only regular files and directories is not a rule 5 violation"

# 8d (pass 2, finding D). Rule 5 must see skills/flow-contracts/ too — it
# is shared prose, not a command skill, so COMMAND_SKILLS_FILE (rules 1-4's
# and coverage's own scan set) excludes it, but rule 5's requirement is
# about EVERY skill directory's own top level. Before the fix, rule 5
# reused COMMAND_SKILLS_FILE and this symlink went entirely unreported.
new_repo
mkdir -p "$REPO/skills/flow-contracts"
printf 'contract\n' > "$REPO/skills/flow-contracts/pipeline.md"
printf 'other\n' > "$REPO/other-file.txt"
ln -s "../../other-file.txt" "$REPO/skills/flow-contracts/evil-link"
run_guard "$REPO"
assert_invalid "a symlink directly under skills/flow-contracts/ is a rule 5 violation"
assert_reports "evil-link" "rule 5: names the symlink under skills/flow-contracts/"

# 8e (pass 2, finding D). Rules 1-4 must keep ignoring flow-contracts:
# widening rule 5's scan must not widen theirs. A non-symlink entry under
# skills/flow-contracts/scripts/ — a rule 1 violation for any COMMAND
# skill — must stay unreported here, since flow-contracts is still
# excluded from COMMAND_SKILLS_FILE. A real, cleanly-linked command skill
# is added alongside it purely so the corpus is non-empty — an
# all-flow-contracts fixture trips the unrelated KAN-197 empty-corpus
# refusal, which is not what this case is about.
new_repo
add_real_guard "check-foo.sh" "$PLAIN_GUARD_BODY"
link_guard "myflow-do" "check-foo.sh"
write_skill_md "myflow-do" '# myflow-do fixture

Run the guard:

```bash
check-foo.sh <worktree>
```
'
mkdir -p "$REPO/skills/flow-contracts/scripts"
: > "$REPO/skills/flow-contracts/scripts/not-a-symlink.sh"
write_skill_md "flow-contracts" "# fixture, no citations"
run_guard "$REPO"
assert_silent "a non-symlink entry under skills/flow-contracts/scripts/ is not a rule 1 violation — flow-contracts stays out of rules 1-4's scan"

if [ "$FAILURES" -eq 0 ]; then
  printf '\n✓ PASS\n'
  exit 0
fi
printf '\n✗ FAIL — %s failure(s)\n' "$FAILURES" >&2
exit 1
