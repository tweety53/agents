#!/usr/bin/env bash
# Assertion harness for stats/Makefile: the `build` target's recipe, and the
# UI-test stack's agreement with the pidfile myflowd writes for itself.
#
# THREE KINDS OF ASSERTION, and the split is deliberate:
#
# - Cases 1-3 run `make -n build` in stats/ and assert on the PRINTED RECIPE.
#   `-n` is the whole point: the defect they pin is a MISSING `-o` in the
#   recipe text, and `-n` shows the recipe without running it -- no real
#   compile, no `npm ci`, no Vite run, so this harness is cheap enough to sit
#   in the project's `## test` list beside the other scripts/test-*.sh suites.
#
# - Cases 4-5 assert on stats/Makefile's SOURCE TEXT, not on a `make -n`
#   expansion. They are about ui-test-up and ui-test-down, and `make -n` for
#   either of those targets is not free of consequences to reach: ui-test-up's
#   prerequisites include web-build. Reading the file is the cheap way to pin
#   a fact that lives in the file anyway -- which variable is assigned what,
#   and which redirection is absent.
#
# - Case 6 EXECUTES a fragment lifted out of that source text. Cases 4-5 can
#   only ever say what the Makefile is spelled like; case 6 asks whether the
#   shell those recipes hand to `kill` actually produces a pid, by running
#   each pid-reading command substitution against a real two-line pidfile in
#   the format internal/pidfile writes. A text assertion cannot catch a
#   `cat` that is spelled exactly as intended and still yields two lines --
#   which is the defect it exists for.
#
# WHY THIS FILE EXISTS. `make build` ran `go build ./...` alone, which
# compiles every package and then discards the resulting binaries. The
# documented launch sequence therefore needed a second, hand-written
# `go build -o bin/myflowd ./cmd/myflowd` after it, and an operator who
# skipped that half ran whatever stale `bin/myflowd` was already on disk.
# Cases 1-3 were written and run against the OLD recipe first, and cases 1
# and 2 failed there; cases 4 and 5 were written and run against the OLD
# Makefile first, and both failed there; case 6 was written and run against
# a Makefile whose three pid reads were still `$$(cat $(UITEST_PIDFILE))`,
# and failed there naming all three -- which is what makes this suite
# evidence rather than a description of what the file is hoped to say.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
STATS_DIR="$(cd "$SCRIPT_DIR/.." && pwd)/stats"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

# The recipe text, captured once. A `make -n` that itself fails is a hard
# stop rather than three confusing assertion failures: every case below
# reads this variable, so an empty capture would report the Makefile as
# missing all three lines when the real fault is that Make never parsed it.
set +e
RECIPE="$(cd "$STATS_DIR" && make -n build 2>&1)"
RC=$?
set -e
if [ "$RC" -ne 0 ]; then
  printf 'FAIL: make -n build exited %s in %s:\n%s\n' "$RC" "$STATS_DIR" "$RECIPE" >&2
  exit 1
fi

# assert_recipe_matches <case-name> <extended-regex>
assert_recipe_matches() {
  local name="$1" pattern="$2"
  if printf '%s\n' "$RECIPE" | grep -Eq "$pattern"; then
    pass "$name"
  else
    fail "$name: no recipe line matched /$pattern/; recipe was:
$RECIPE"
  fi
}

# The Makefile's own source text, for cases 4-6. Captured with the same
# hard-stop discipline as the recipe above: an unreadable Makefile is one
# error, not two assertion failures blaming it for lines it does not have.
if [ ! -f "$STATS_DIR/Makefile" ]; then
  printf 'FAIL: %s/Makefile not found\n' "$STATS_DIR" >&2
  exit 1
fi
MAKEFILE_TEXT="$(cat "$STATS_DIR/Makefile")"

# assert_makefile_matches_all <case-name> <extended-regex>...
# One case, several patterns: the case passes only when EVERY pattern matches
# some line, and prints exactly one `ok:`/`FAIL:` either way. Used where a
# single property is spelled across more than one line of the Makefile, and
# half of it holding is not the property holding.
assert_makefile_matches_all() {
  local name="$1"
  shift
  local pattern missing=""
  for pattern in "$@"; do
    printf '%s\n' "$MAKEFILE_TEXT" | grep -Eq "$pattern" || missing="$missing
  /$pattern/"
  done
  if [ -z "$missing" ]; then
    pass "$name"
  else
    fail "$name: no line of stats/Makefile matched:$missing"
  fi
}

# assert_makefile_lacks <case-name> <extended-regex>
# The mirror of the above: the case passes when NOTHING in the file matches.
# Reports the offending lines, since "some line matched" is useless without
# knowing which.
assert_makefile_lacks() {
  local name="$1" pattern="$2" hits
  hits="$(printf '%s\n' "$MAKEFILE_TEXT" | grep -nE "$pattern" || true)"
  if [ -z "$hits" ]; then
    pass "$name"
  else
    fail "$name: stats/Makefile still matches /$pattern/ at:
$hits"
  fi
}

# ===========================================================================
# 1-2. The target emits both binaries. Each `-o` path is anchored on a
#      following space or end-of-line so `bin/myflow` cannot be satisfied by
#      the `bin/myflowd` line, and the whitespace between the `-o` path and
#      the package is `[[:space:]]+` so the Makefile is free to align the
#      two lines with extra spaces.
# ===========================================================================
assert_recipe_matches "build target emits bin/myflowd" \
  '(^|[[:space:]])go build[[:space:]]+-o[[:space:]]+bin/myflowd[[:space:]]+\./cmd/myflowd([[:space:]]|$)'
assert_recipe_matches "build target emits bin/myflow" \
  '(^|[[:space:]])go build[[:space:]]+-o[[:space:]]+bin/myflow[[:space:]]+\./cmd/myflow([[:space:]]|$)'

# ===========================================================================
# 3. The compile-all pass survives. Replacing `go build ./...` with the two
#    `-o` builds alone would trade the target's coverage for the two
#    artifacts -- cmd/uitest-seed, and every library with no binary of its
#    own, would silently stop being compiled by this target.
# ===========================================================================
assert_recipe_matches "build target still compiles every package" \
  '(^|[[:space:]])go build[[:space:]]+\./\.\.\.([[:space:]]|$)'

# ===========================================================================
# 4. ui-test-up does not write the pidfile itself. myflowd writes
#    $TMPDIR/myflowd-<port>.pid on every start (internal/pidfile), so a
#    `& echo $! > $(UITEST_PIDFILE)` here would be a SECOND writer of one
#    file for one process -- and the two can disagree. `echo $!` records the
#    pid of the shell's background job, which for this recipe is the daemon
#    itself, but a future edit that wraps the launch (a `sh -c`, a `nohup ...
#    &`, an `env` prefix that forks) makes $! the wrapper and the daemon's
#    own line the truth. The tear-down `kill` then signals the wrong process.
#
#    The pattern is deliberately broader than the `echo $!` that was there:
#    ANY redirection into $(UITEST_PIDFILE) is a second writer, whatever
#    command produces it. `rm -f $(UITEST_PIDFILE)` is untouched by this --
#    it has no `>` -- and is meant to stay: it clears a file left behind by
#    a daemon this target killed, and myflowd overwriting a stale file does
#    not make removing one wrong.
# ===========================================================================
assert_makefile_lacks "ui-test-up does not write its own pidfile" \
  '>[[:space:]]*\$\(UITEST_PIDFILE\)'

# ===========================================================================
# 5. UITEST_PIDFILE names the file myflowd actually writes, and stays
#    `override`. Both halves of one property, which is why they are one case:
#
#    - `override UITEST_PIDFILE := /tmp/myflowd-$(UITEST_PORT).pid` --
#      port-derived, so the live stack (4173) and the UI-test stack (4174)
#      never contend for one file, and `override` so no command-line or
#      environment assignment can aim ui-test-down's `kill` and `rm -f` at a
#      caller-named path. scripts/check-uitest-overrides.sh enforces
#      `override` over the whole UITEST_* class; this case asserts it from
#      the consuming side, where the path being wrong is what actually
#      breaks.
#
#    - `TMPDIR=/tmp` on the daemon's launch. pidfile.Path joins
#      os.TempDir(), NOT a hardcoded /tmp, and on macOS os.TempDir() is the
#      per-user $TMPDIR (/var/folders/.../T/) -- so without this pin the
#      daemon would write /var/folders/.../T/myflowd-4174.pid while this
#      Makefile looked in /tmp, and every `[ -f $(UITEST_PIDFILE) ]` guard
#      below would silently take the "no previous daemon" branch. Pinning
#      the daemon's TMPDIR is what keeps the literal path above honest, and
#      it keeps the path itself non-caller-nameable -- deriving the Make
#      variable from $(TMPDIR) instead would hand a caller the `kill` target
#      that `override` exists to deny them.
# ===========================================================================
assert_makefile_matches_all "UITEST_PIDFILE is override and derived from UITEST_PORT" \
  '^override[[:space:]]+UITEST_PIDFILE[[:space:]]*:=[[:space:]]*/tmp/myflowd-\$\(UITEST_PORT\)\.pid[[:space:]]*$' \
  '(^|[[:space:]])TMPDIR=/tmp([[:space:]]|$)'

# ===========================================================================
# 6. Every place the UI-test stack reads a pid out of $(UITEST_PIDFILE)
#    yields a pid `kill` accepts.
#
#    The file is myflowd's own, and it has TWO lines -- the decimal pid, then
#    the path of the executable that wrote it (stats/internal/pidfile: the
#    second line is what lets the identity check recognise the UI-test
#    stack's differently-named copy of the daemon). A `$(cat FILE)` therefore
#    captures both, joined by the newline command substitution does not strip
#    because it is not trailing, and `kill -0 "$pid"` then fails with
#    "illegal pid" -- ALWAYS, and never because the daemon is dead.
#
#    That failure is silent and it inverts both recipes. ui-test-up's
#    stop-previous-daemon guard takes its "no previous daemon" branch, `rm -f`s
#    the live daemon's pidfile and launches a second daemon that cannot bind.
#    ui-test-down skips its SIGTERM and the bounded wait behind it, and goes
#    straight to `workspace.sh remove` -- which, unlike ui-test-up's, carries
#    no --force and so races a daemon still holding its connection open.
#
#    So this case runs the real thing rather than describing it: it lifts every
#    `$$(... $(UITEST_PIDFILE))` command substitution out of the Makefile,
#    rewrites it into the shell text make would hand /bin/sh, and evaluates it
#    against a fixture pidfile naming a process that is genuinely alive. A read
#    that cannot be signalled is the defect, whatever it is spelled with.
# ===========================================================================
PIDFILE_FIXTURE="$(mktemp "${TMPDIR:-/tmp}/test-make-build-pidfile.XXXXXX")"
sleep 120 &
FIXTURE_PID=$!
trap 'kill "$FIXTURE_PID" 2>/dev/null || true; rm -f "$PIDFILE_FIXTURE"' EXIT
# The executable line is the UI-test daemon's, so the fixture is the exact
# file `make ui-test-down` reads on the stack this case is about.
printf '%s\n/tmp/myflow-uitest-myflowd\n' "$FIXTURE_PID" > "$PIDFILE_FIXTURE"

PID_READS="$(printf '%s\n' "$MAKEFILE_TEXT" | grep -oE '\$\$\([^()]+\$\(UITEST_PIDFILE\)\)' || true)"
READ_COUNT=0
UNSIGNALLABLE=""
while IFS= read -r pid_read; do
  [ -n "$pid_read" ] || continue
  READ_COUNT=$((READ_COUNT + 1))
  # `$$(` is how a recipe spells a shell command substitution to make, and
  # $(UITEST_PIDFILE) is expanded by make before /bin/sh ever sees the line.
  # Doing both by hand is what lets this case run the fragment without
  # running the target it came from -- ui-test-up drops a database.
  shell_read="$(printf '%s\n' "$pid_read" \
    | sed -e 's/^\$\$(/$(/' -e 's|\$(UITEST_PIDFILE)|'"$PIDFILE_FIXTURE"'|g')"
  eval "read_pid=\"$shell_read\""
  if ! kill -0 "$read_pid" 2>/dev/null; then
    UNSIGNALLABLE="$UNSIGNALLABLE
  $pid_read read $(printf '%q' "$read_pid"), which kill rejects"
  fi
done <<PID_READS_EOF
$PID_READS
PID_READS_EOF

# Zero reads is a failure, not a pass: the recipes signal a pid they get from
# somewhere, and a case that asserts over an empty set asserts nothing.
if [ "$READ_COUNT" -eq 0 ]; then
  fail "ui-test pid reads take only the pidfile's pid line: found no \$(UITEST_PIDFILE) read in stats/Makefile to check"
elif [ -n "$UNSIGNALLABLE" ]; then
  fail "ui-test pid reads take only the pidfile's pid line: the daemon's two-line pidfile is read whole by:$UNSIGNALLABLE"
else
  pass "ui-test pid reads take only the pidfile's pid line"
fi

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'make-build: all cases pass\n'
