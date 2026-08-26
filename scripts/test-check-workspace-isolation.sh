#!/usr/bin/env bash
# Assertion harness for check-workspace-isolation.sh. Builds throwaway project
# roots under a sandboxed TMPDIR, writes a `.flow/project.md` into each, and
# asserts the guard's violation lines, its verdict and its exit status. Never
# reads or writes the real repository tree, and never reads another real
# project on this machine — section 14's `.myflow/` -> `.flow/` hard-cutover
# cases (design.md's dotmyflow-hard-cutover) are fixtures this suite owns
# rather than a read of gymie's or spectre-e2e's real, still-`.myflow/`-only
# declarations, which this change does not rename.
#
# READ THIS BEFORE ADDING OR "FIXING" A CASE. Assert against the stated contract
# in skills/myflow-contracts/project-configuration.md — the `## workspace
# isolation` row of the key table, the four `In a workspace` cell forms, "What a
# `url` row may reference, and what it may not", and the two bullets under "An
# isolation row resolves under the same rules this file applies to everything
# else it consumes". Never assert against observed output.
# test-check-plan-provenance.sh's header records that suite encoding the guard's
# own defects as its specification more than once, which then made each defect
# look verified.
#
# WHY EVERY CASE ASSERTS ON THE REASON TEXT. The contract's remedy for a bad row
# is that the row is "reported by name and dropped", so a guard that answered
# only "this file is invalid" would satisfy every exit-status assertion while
# leaving the operator nothing to act on. The needles below are short and
# behavioural — evidence that the named row, cell and rule reached the report,
# not a transcript of its prose.
#
# WHY THIS DUPLICATES test-check-cleanup-complete.sh's HELPERS instead of
# sharing them. The two guards read the same section of the same file and are
# deliberately kept independent: that guard asks whether the removal happened,
# this one asks whether the declaration is well formed, and a shared harness
# library would couple their suites so that a change to one guard's contract
# could only be made by editing a file the other one also runs. The duplication
# is the cheaper of the two, and it is recorded here rather than left
# unexplained.
#
# Bash 3.2 is the floor, as test-check-finish-preflight.sh's header records:
# indexed arrays only, no associative arrays.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
GUARD="$SCRIPT_DIR/check-workspace-isolation.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

# skip <label> <reason> — a case this environment cannot decide, reported as
# neither a pass nor a failure. Printing "ok" for an assertion that never ran is
# the vacuous pass every guard in this repository is written against.
skip() { printf 'skip: %s (%s)\n' "$1" "$2"; }

# An indexed array, not a space-separated string: sandbox paths come from mktemp
# under TMPDIR, which may contain spaces, and word-splitting a string would then
# rm -rf the fragments.
SANDBOXES=()
cleanup() {
  [ "${#SANDBOXES[@]}" -eq 0 ] && return 0
  for s in "${SANDBOXES[@]}"; do rm -rf "$s"; done
}
trap cleanup EXIT

WORK="$(mktemp -d "${TMPDIR:-/tmp}/workspace-isolation-test.XXXXXX")"
SANDBOXES+=("$WORK")
ERRFILE="$WORK/stderr"

# run_guard <arg ...> -> sets OUT (stdout only), ERR, RC.
# The two streams are captured SEPARATELY rather than merged with 2>&1, because
# the contract distinguishes them: a refusal puts its message on stderr and must
# leave stdout empty, and a merged capture cannot tell an empty stdout from a
# stdout carrying the message.
run_guard() {
  set +e
  OUT="$("$GUARD" "$@" 2>"$ERRFILE")"
  RC=$?
  set -e
  ERR="$(cat "$ERRFILE")"
}

# new_project -> sets REPO to an empty project root. No .flow directory yet:
# the absent-file case is the first thing this guard has to get right.
new_project() {
  REPO="$(mktemp -d "${TMPDIR:-/tmp}/workspace-isolation-repo.XXXXXX")"
  SANDBOXES+=("$REPO")
}

# write_config <body> -> the whole of $REPO/.flow/project.md.
write_config() {
  mkdir -p "$REPO/.flow"
  printf '%s\n' "$1" > "$REPO/.flow/project.md"
}

# The rows a correct declaration is built from, kept in one place so a case that
# breaks one row is visibly a case about THAT row. Every cell form the contract
# names appears here, so this set doubles as the positive case: two removable
# resources, the probed index, a port, and the three `url` shapes — one built
# from a `<value:…>` port reference, one from a `<value:…>` bucket reference,
# and one carrying no token at all.
VALID_RES='| `database` | `DB_URL` | `jdbc:postgresql://localhost:5432/appdb` | `jdbc:postgresql://localhost:5432/appdb_<id_underscored>` |
| `bucket` | `MEDIA_BUCKET` | `appdb-media` | `appdb-media-<id>` |
| `cache index` | `CACHE_INDEX` | `0` | `probed` |
| `port` | `API_PORT` | `8080` | `+<offset>` |
| `url` | `MEDIA_BASE_URL` | `http://localhost:9000/appdb-media` | `http://localhost:9000/<value:MEDIA_BUCKET>` |
| `url` | `WEB_URL` | `http://localhost:8080` | `http://localhost:<value:API_PORT>` |
| `url` | `STATIC_URL` | `http://localhost:9000` | `http://localhost:9000` |'

VALID_CMD='| `create` | `./scripts/workspace create` |
| `remove` | `./scripts/workspace remove <id>` |
| `survivors` | `./scripts/workspace survivors <id>` |'

# write_iso <resource-rows> <command-rows> -> a project.md whose
# `## workspace isolation` section carries the two tables with those rows, and
# prose around them, because prose beside the tables is permitted and is never
# read. A section written without any is not the shape a project actually
# writes, and a parser that only ever met bare tables would not be tested
# against the one it will meet.
write_iso() {
  write_config "# fixture project configuration

## apps

Nothing here is read by this guard.

## workspace isolation

Prose above the tables, which the resolver ignores.

| Resource | Variable | Default | In a workspace |
|----------|----------|---------|----------------|
$1

Prose between the tables, likewise ignored.

| Command | Runs |
|---------|------|
$2

Prose below the tables, saying what this project has deliberately not isolated.

## lint

Also not read by this guard."
}

# assert_ok <label> — exit 0 and an ISOLATION-OK verdict on the last line.
assert_ok() {
  if [ "$RC" -ne 0 ]; then
    fail "$1: expected exit 0, got rc=$RC out=$OUT err=$ERR"
    return 0
  fi
  case "$(printf '%s\n' "$OUT" | tail -n 1)" in
    "ISOLATION-OK:"*) pass "$1" ;;
    *) fail "$1: expected an ISOLATION-OK verdict, got: $OUT" ;;
  esac
}

# assert_silent <label> — exit 0, and the verdict line is the ONLY line. A
# project that declares no section is the overwhelmingly common case, and a
# guard that printed a note per project would make every lint run noisier for
# every repository that has nothing to declare.
assert_silent() {
  local lines
  assert_ok "$1"
  lines="$(printf '%s\n' "$OUT" | wc -l | tr -d ' ')"
  [ "$lines" = "1" ] && pass "$1: says exactly one line" \
    || fail "$1: expected exactly one stdout line, got $lines: $OUT"
}

# assert_invalid <label> — exit 1 and an ISOLATION-INVALID verdict. Exit 1 is
# "violations found", kept distinct from exit 2, "cannot answer at all".
assert_invalid() {
  if [ "$RC" -ne 1 ]; then
    fail "$1: expected exit 1, got rc=$RC out=$OUT err=$ERR"
    return 0
  fi
  case "$(printf '%s\n' "$OUT" | tail -n 1)" in
    "ISOLATION-INVALID:"*) pass "$1" ;;
    *) fail "$1: expected an ISOLATION-INVALID verdict, got: $OUT" ;;
  esac
}

# assert_reports <needle> <label> — the report names this row, cell or rule.
assert_reports() {
  case "$OUT" in
    *"$1"*) pass "$2" ;;
    *) fail "$2: the report does not name '$1': $OUT" ;;
  esac
}

# assert_not_reported <needle> <label> — the report does NOT name it. A guard
# that reported a row which is in fact well formed sends the operator to fix
# something that is not broken, and every exit-status assertion still passes
# while it does.
assert_not_reported() {
  case "$OUT" in
    *"$1"*) fail "$2: the report names a row that is well formed ('$1'): $OUT" ;;
    *) pass "$2" ;;
  esac
}

# assert_refuses <label> — the refusal shape: exit 2, nothing on stdout, and the
# guard's own name on stderr. The needle carries the colon deliberately. Without
# it the shell's own "…/check-workspace-isolation.sh: No such file or directory"
# satisfies the case, so it passes while the guard does not exist — a vacuous
# assertion that proves nothing about the guard's own reporting.
assert_refuses() {
  [ "$RC" -eq 2 ] && pass "$1: exits 2" \
    || fail "$1: expected exit 2, got rc=$RC out=$OUT"
  [ -z "$OUT" ] && pass "$1: writes nothing to stdout" \
    || fail "$1: emitted a verdict line: $OUT"
  case "$ERR" in
    *"check-workspace-isolation: "*) pass "$1: names the failure on stderr" ;;
    *) fail "$1: no named message on stderr: $ERR" ;;
  esac
}

# ---------------------------------------------------------------------------
# 1. A project that declares nothing.
# ---------------------------------------------------------------------------

# 1a. No .flow/project.md at all. The file is optional and its absence is a
#     supported, ordinary case — never an error.
new_project
run_guard "$REPO"
assert_silent "a project with no .flow/project.md passes silently"

# 1b. A project.md with no `## workspace isolation` section. This is the
#     overwhelmingly common case across projects myflow is installed into.
new_project
write_config "# fixture

## test

\`\`\`bash
scripts/test-setup.sh
\`\`\`
"
run_guard "$REPO"
assert_silent "a project declaring no isolation section passes silently"

# 1c. The agents repository itself, read from its real path rather than from a
#     fixture. It declares its own `## workspace isolation` section, and this
#     guard is on its own lint list — a guard that failed on the repository it
#     ships in would be reverted rather than fixed. assert_silent checks a
#     single-line verdict, not empty output, so a well-formed section still
#     passes: it reports ISOLATION-OK on one line.
run_guard "$REPO_ROOT"
assert_silent "the agents repository's own section validates cleanly"

# 1d. Invoked with no argument at all, which is how the `## lint` list runs it.
#     It resolves its own repository root, so a lint step is one word long.
run_guard
assert_silent "invoked bare, the guard checks its own repository"

# ---------------------------------------------------------------------------
# 2. Inputs it cannot answer about. Never fail open: a file it cannot read is
#    not a project that declares nothing.
# ---------------------------------------------------------------------------

# 2a. A path that is not a directory. Reading "declares nothing" out of a
#     mistyped path would report every project valid.
run_guard "$WORK/no-such-project-root"
assert_refuses "a path that is not a directory"

# 2b. A project.md that exists but cannot be read. Absence read out of a path
#     that was never readable is the false pass this guard exists to prevent.
#
#     THE GATE IS `-r` ON THE FILE ITSELF, not `id -u`. Non-root does not imply
#     a mode-000 file is unreadable: a privileged capability, a filesystem
#     mounted with permissions turned off, or an ACL that outranks the mode bits
#     all leave a non-root user reading it, and this case then failed with
#     `rc=0` and `ISOLATION-OK` — measured once in six runs on this machine
#     before the gate changed. `-r` asks the runtime condition the guard itself
#     branches on instead of guessing it from the user id, which is what
#     test-check-cleanup-complete.sh's case 26 already does for the same
#     scenario.
new_project
write_config "# fixture"
chmod 000 "$REPO/.flow/project.md"
if [ -r "$REPO/.flow/project.md" ]; then
  skip "an unreadable project.md refuses" "this user can read a mode-000 file"
else
  run_guard "$REPO"
  assert_refuses "an unreadable project.md"
fi
chmod 644 "$REPO/.flow/project.md"

# 2c. A `.flow/project.md` that is a DIRECTORY. `-r` is true of a directory,
#     and `grep -qiE` then fails with "Is a directory" while its exit status is
#     indistinguishable from "no match" — so this shipped as
#     `ISOLATION-OK: … declares no `## workspace isolation` section` and exit 0,
#     the guard's own NEVER FAIL OPEN invariant broken in the one script written
#     to hold it. Anyone able to land a pull request can create the symlink.
new_project
mkdir -p "$REPO/.flow/project.md"
run_guard "$REPO"
assert_refuses "a project.md that is a directory"

# 2d. A `.flow/project.md` that is a symlink to nowhere. `-e` follows the link,
#     so the absent test answers true and the guard would report the ordinary,
#     silent "no .flow/project.md" verdict — an absence manufactured out of a
#     path someone deliberately pointed at nothing.
new_project
mkdir -p "$REPO/.flow"
ln -s "$WORK/no-such-configuration-target" "$REPO/.flow/project.md"
run_guard "$REPO"
assert_refuses "a project.md that is a dangling symlink"

# 2e. A `.flow/project.md` that is a symlink to a REAL file is still read. `-f`
#     follows symlinks, so the refusals above exclude a file type rather than an
#     indirection — a project keeping its configuration behind a link is a
#     supported, ordinary case.
new_project
mkdir -p "$REPO/.flow"
printf '# fixture\n' > "$WORK/linked-project.md"
ln -s "$WORK/linked-project.md" "$REPO/.flow/project.md"
run_guard "$REPO"
assert_silent "a project.md that is a symlink to a real file is read"

# ---------------------------------------------------------------------------
# 3. The declaration that must pass.
# ---------------------------------------------------------------------------

# 3a. Every cell form the contract names, in one section.
new_project
write_iso "$VALID_RES" "$VALID_CMD"
run_guard "$REPO"
assert_ok "a declaration using every cell form passes"

# ---------------------------------------------------------------------------
# 4. The closed `Resource` vocabulary. Cleanup selects the rows it must remove
#    by this word, so a spelling it does not know is a resource nobody removes.
# ---------------------------------------------------------------------------

# 4a. A word outside the five. `queue` is the shape a project reaches for when
#     it isolates something the contract does not name, and it is exactly the
#     one that must be reported rather than guessed about.
new_project
write_iso '| `queue` | `QUEUE_NAME` | `appq` | `appq-<id>` |
| `port` | `API_PORT` | `8080` | `+<offset>` |' "$VALID_CMD"
run_guard "$REPO"
assert_invalid "an unknown Resource word is a violation"
assert_reports "QUEUE_NAME" "the unknown-Resource report names the row"
assert_reports "queue" "the unknown-Resource report names the word it did not know"
assert_not_reported "API_PORT" "the well-formed row beside it is not reported"

# 4b. A near miss, not a random word. `databases` differs from `database` by one
#     character, which is how the mistake is actually made — and a guard that
#     matched on a prefix or a substring would pass it.
new_project
write_iso '| `databases` | `DB_URL` | `appdb` | `appdb_<id_underscored>` |' "$VALID_CMD"
run_guard "$REPO"
assert_invalid "a near-miss Resource word is a violation"
assert_reports "databases" "the near-miss report names the word"

# 4c. `cache index` is two words, and its internal space is part of the word.
#     A guard that folded whitespace away would accept `cacheindex`, and one
#     that split the cell on whitespace would reject the real spelling.
new_project
write_iso '| `cacheindex` | `CACHE_INDEX` | `0` | `probed` |' "$VALID_CMD"
run_guard "$REPO"
assert_invalid 'cacheindex is not the two-word cache index'
assert_reports "cacheindex" "the report names the run-together spelling"

# 4d. Case and surrounding whitespace are not the mistake. A project that writes
#     `Database` has spelled the word; rejecting it would send an author hunting
#     for a typo that is not there.
new_project
write_iso '| `Database` | `DB_URL` | `appdb` | `appdb_<id_underscored>` |
| ` CACHE INDEX ` | `CACHE_INDEX` | `0` | `probed` |' "$VALID_CMD"
run_guard "$REPO"
assert_ok "the Resource word is read case-insensitively and untrimmed"

# ---------------------------------------------------------------------------
# 5. The `Variable` cell. A run EXPORTS these, so a cell that is not a legal
#    environment-variable name is a row nothing can carry.
# ---------------------------------------------------------------------------

# 5a. Empty.
new_project
write_iso '| `port` | `` | `8080` | `+<offset>` |' "$VALID_CMD"
run_guard "$REPO"
assert_invalid "an empty Variable is a violation"
assert_reports "Variable" "the empty-Variable report names the cell"

# 5b. A name a shell cannot export. The hyphen is the one an author writes
#     without noticing, having copied it from a flag or a hostname.
new_project
write_iso '| `port` | `API-PORT` | `8080` | `+<offset>` |' "$VALID_CMD"
run_guard "$REPO"
assert_invalid "a Variable that is not a legal environment-variable name"
assert_reports "API-PORT" "the report names the illegal variable"

# 5c. A leading digit, which the contract's own shape excludes.
new_project
write_iso '| `port` | `8080_PORT` | `8080` | `+<offset>` |' "$VALID_CMD"
run_guard "$REPO"
assert_invalid "a Variable starting with a digit"
assert_reports "8080_PORT" "the report names the variable starting with a digit"

# 5d. A lowercase name with a leading underscore is legal and must pass. The
#     contract's shape is `[A-Za-z_][A-Za-z0-9_]*`, not SCREAMING_CASE, and a
#     guard that demanded convention would reject a project's real spelling.
new_project
write_iso '| `port` | `_api_port` | `8080` | `+<offset>` |' "$VALID_CMD"
run_guard "$REPO"
assert_ok "a lowercase Variable with a leading underscore is legal"

# ---------------------------------------------------------------------------
# 6. The bare-integer `Default`. A `port` cell is read as this row's `Default`
#    plus an offset, and a `cache index` default is the index the empty-id case
#    selects — there is nothing to add an offset to in `localhost:8080`.
# ---------------------------------------------------------------------------

# 6a. The shape an author actually writes: the port with its host in front.
new_project
write_iso '| `port` | `API_PORT` | `localhost:8080` | `+<offset>` |' "$VALID_CMD"
run_guard "$REPO"
assert_invalid "a port Default carrying a host is a violation"
assert_reports "API_PORT" "the port-Default report names the row"
assert_reports "localhost:8080" "the port-Default report names the value it rejected"

# 6b. A port default that is empty.
new_project
write_iso '| `port` | `API_PORT` | `` | `+<offset>` |' "$VALID_CMD"
run_guard "$REPO"
assert_invalid "an empty port Default is a violation"

# 6c. A cache index default that is not a bare integer, on the same rule and for
#     the same reason: an index is a number.
new_project
write_iso '| `cache index` | `CACHE_INDEX` | `db0` | `probed` |' "$VALID_CMD"
run_guard "$REPO"
assert_invalid "a cache index Default that is not a bare integer"
assert_reports "CACHE_INDEX" "the cache-index-Default report names the row"
assert_reports "db0" "the cache-index-Default report names the value it rejected"

# 6d. `8080 ` with a trailing space is the same number, and the surrounding
#     whitespace is not the mistake. A guard that rejected it would send an
#     author hunting for a typo that is not there.
new_project
write_iso '| `port` | `API_PORT` |   `8080`   | `+<offset>` |' "$VALID_CMD"
run_guard "$REPO"
assert_ok "a padded bare integer is still a bare integer"

# 6e. The rule binds `port` and `cache index` and NOTHING ELSE. A `database`
#     default is a connection string and a `url` default is a URL; a guard that
#     demanded an integer of every `Default` would reject every real project.
new_project
write_iso '| `database` | `DB_URL` | `jdbc:postgresql://localhost:5432/appdb` | `appdb_<id_underscored>` |
| `url` | `WEB_URL` | `http://localhost:3000` | `http://localhost:3000` |' "$VALID_CMD"
run_guard "$REPO"
assert_ok "a non-integer Default on a database or url row is legal"

# ---------------------------------------------------------------------------
# 7. The four `In a workspace` cell forms, one violated at a time. The form is
#    decided by the ROW's Resource, so each case below is a cell that would be
#    legal under some other form.
# ---------------------------------------------------------------------------

# 7a. Form 1, `database`: an empty cell. Nothing is derived from it, so the
#     workspace would run against the shared database with nothing having
#     failed.
new_project
write_iso '| `database` | `DB_URL` | `appdb` | `` |' "$VALID_CMD"
run_guard "$REPO"
assert_invalid "an empty database cell is a violation"
assert_reports "DB_URL" "the empty-database-cell report names the row"

# 7b. Form 1, `database`: a token the contract does not name. Exactly two are
#     substituted in a `database` cell and no others, so `<workspace>` reaches
#     the value as a literal nobody intended.
new_project
write_iso '| `database` | `DB_URL` | `appdb` | `appdb_<workspace>` |' "$VALID_CMD"
run_guard "$REPO"
assert_invalid "an unknown token in a database cell is a violation"
assert_reports "<workspace>" "the unknown-token report names the token"

# 7c. Form 1, `bucket`: `<value:…>` is legal in a `url` cell and NOWHERE ELSE. A
#     bucket cell is what the removal commands target, and a removal target
#     assembled out of other rows would depend on those rows having resolved
#     first.
new_project
write_iso '| `port` | `API_PORT` | `8080` | `+<offset>` |
| `bucket` | `MEDIA_BUCKET` | `appdb-media` | `appdb-media-<value:API_PORT>` |' "$VALID_CMD"
run_guard "$REPO"
assert_invalid "a <value:…> reference in a bucket cell is a violation"
assert_reports "MEDIA_BUCKET" "the misplaced-reference report names the row"

# 7d. Form 2, `port`: the cell that carries no token at all. This is the shape
#     the contract calls out by name — `9090` names no token, so a check that
#     only looked at tokens would pass it while the value was added to its own
#     `Default`.
new_project
write_iso '| `port` | `API_PORT` | `8080` | `9090` |' "$VALID_CMD"
run_guard "$REPO"
assert_invalid "a port cell carrying a literal number is a violation"
assert_reports "API_PORT" "the port-cell report names the row"
assert_reports "9090" "the port-cell report names the cell it rejected"

# 7e. Form 2, `port`: arithmetic the cell cannot carry. The token is there, so a
#     token-only check passes it; the form says the literal stands alone.
new_project
write_iso '| `port` | `API_PORT` | `8080` | `8080+<offset>` |' "$VALID_CMD"
run_guard "$REPO"
assert_invalid "a port cell with anything before the token is a violation"

# 7f. Form 3, `cache index`: a literal index rather than the literal `probed`.
#     The value is claimed at run time, so a cell naming one is a cell claiming
#     something the run does not honour.
new_project
write_iso '| `cache index` | `CACHE_INDEX` | `0` | `3` |' "$VALID_CMD"
run_guard "$REPO"
assert_invalid "a cache index cell naming an index is a violation"
assert_reports "CACHE_INDEX" "the cache-index-cell report names the row"

# 7g. Form 3, `cache index`: `probed` with words around it. The literal is
#     alone, likewise.
new_project
write_iso '| `cache index` | `CACHE_INDEX` | `0` | `probed at run time` |' "$VALID_CMD"
run_guard "$REPO"
assert_invalid "a cache index cell with prose around the literal is a violation"

# 7h. Form 4, `url`: an unknown token. All three of the contract tokens are
#     legal here and there is no fourth.
new_project
write_iso '| `url` | `WEB_URL` | `http://localhost:3000` | `http://localhost:<port>` |' "$VALID_CMD"
run_guard "$REPO"
assert_invalid "an unknown token in a url cell is a violation"
assert_reports "<port>" "the url-token report names the token"

# 7i. Form 4, `url`: an empty cell.
new_project
write_iso '| `url` | `WEB_URL` | `http://localhost:3000` | `` |' "$VALID_CMD"
run_guard "$REPO"
assert_invalid "an empty url cell is a violation"

# 7j. The forms that must NOT be reported. A `url` row carrying no token at all
#     is a legitimate declaration rather than a mistake, and `<id_underscored>`
#     is legal in a `url` cell rather than only in a `database` one — both are
#     stated in the contract, and a guard that rejected either would make a real
#     project unable to declare what it actually has.
new_project
write_iso '| `url` | `STATIC_URL` | `http://localhost:9000` | `http://localhost:9000` |
| `url` | `DB_CONSOLE_URL` | `http://localhost:8081/?db=appdb` | `http://localhost:8081/?db=appdb_<id_underscored>` |' "$VALID_CMD"
run_guard "$REPO"
assert_ok "a url row with no token, and one built from <id_underscored>, both pass"

# 7k. A STRAY SPACE INSIDE THE BRACKETS. `<value: API_PORT>` matches the token
#     shape nowhere, so the cell carries no token at all, every token rule passes
#     over it, and the value reaches a run with the brackets still in it. A space
#     after the colon is the likeliest typo at this keyboard and it was the one
#     the guard could not see: both rows below shipped as ISOLATION-OK, exit 0.
new_project
write_iso '| `port` | `API_PORT` | `8080` | `+<offset>` |
| `url` | `WEB_URL` | `http://localhost:8080` | `http://localhost:<value: API_PORT>` |' "$VALID_CMD"
run_guard "$REPO"
assert_invalid "a space inside a <value:…> reference is a violation"
assert_reports "<value: API_PORT>" "the spaced-token report quotes the run exactly as written"
assert_reports "WEB_URL" "the spaced-token report names the row that carries it"

# 7l. The same defect on a substitution token rather than a reference.
new_project
write_iso '| `database` | `DB_URL` | `appdb` | `appdb_< id_underscored>` |' "$VALID_CMD"
run_guard "$REPO"
assert_invalid "a space inside <id_underscored> is a violation"
assert_reports "< id_underscored>" "the spaced-token report quotes the run as written"

# 7m. A bracketed run that is NOT a near miss of any token this contract names is
#     still ordinary text, and must stay that way. Widening the token shape to
#     admit whitespace would have caught 7k at the cost of reporting every
#     redirection and every angle-bracketed prose fragment, so the rule is
#     "squeezes to a token this contract defines" rather than "looks bracketed" —
#     and this case is what holds that line.
new_project
write_iso '| `url` | `WEB_URL` | `http://localhost:3000/?q=a<b c>d` | `http://localhost:3000/?q=a<b c>d` |' "$VALID_CMD"
run_guard "$REPO"
assert_ok "a bracketed run that names no contract token is not reported"

# ---------------------------------------------------------------------------
# 8. What a `url` row may reference, and what it may not. One level of
#    reference resolves in a single pass, so a reference naming no row, or a row
#    of either excluded kind, is reported by name and dropped.
# ---------------------------------------------------------------------------

# 8a. A reference naming no row in the table. The typo an author makes is a
#     variable that exists in their own configuration but has no row here.
new_project
write_iso '| `port` | `API_PORT` | `8080` | `+<offset>` |
| `url` | `WEB_URL` | `http://localhost:8080` | `http://localhost:<value:GATEWAY_PORT>` |' "$VALID_CMD"
run_guard "$REPO"
assert_invalid "a <value:…> naming no row is a violation"
assert_reports "GATEWAY_PORT" "the dangling-reference report names the variable it could not find"
assert_reports "WEB_URL" "the dangling-reference report names the row that carries it"

# 8b. A reference naming another `url` row. Multi-level references reinstate a
#     cycle to detect, a resolution order to declare rows in, and a
#     partially-resolved value to reason about.
new_project
write_iso '| `url` | `BASE_URL` | `http://localhost:9000` | `http://localhost:9000` |
| `url` | `MEDIA_URL` | `http://localhost:9000/m` | `<value:BASE_URL>/m` |' "$VALID_CMD"
run_guard "$REPO"
assert_invalid "a <value:…> naming another url row is a violation"
assert_reports "BASE_URL" "the url-reference report names the row it pointed at"
assert_reports "MEDIA_URL" "the url-reference report names the row that carries it"

# 8c. A reference naming a `cache index` row. That value is claimed by probing
#     rather than derived, so a `url` row built on it would be resolvable in one
#     place in a run and not in another.
new_project
write_iso '| `cache index` | `CACHE_INDEX` | `0` | `probed` |
| `url` | `CACHE_URL` | `redis://localhost:6379/0` | `redis://localhost:6379/<value:CACHE_INDEX>` |' "$VALID_CMD"
run_guard "$REPO"
assert_invalid "a <value:…> naming a cache index row is a violation"
assert_reports "CACHE_INDEX" "the cache-index-reference report names the row it pointed at"
assert_reports "CACHE_URL" "the cache-index-reference report names the row that carries it"

# 8d. A reference that resolves BACKWARDS, to a row written after it. One level
#     resolves in a single pass precisely so there is no resolution order to
#     declare rows in — a guard that resolved as it read would reject this and
#     make the table order load-bearing.
new_project
write_iso '| `url` | `WEB_URL` | `http://localhost:8080` | `http://localhost:<value:API_PORT>` |
| `port` | `API_PORT` | `8080` | `+<offset>` |' "$VALID_CMD"
run_guard "$REPO"
assert_ok "a <value:…> resolving to a later row passes"

# 8e. All three referenceable kinds resolve: `database`, `bucket` and `port`.
new_project
write_iso '| `database` | `DB_NAME` | `appdb` | `appdb_<id_underscored>` |
| `bucket` | `MEDIA_BUCKET` | `appdb-media` | `appdb-media-<id>` |
| `port` | `API_PORT` | `8080` | `+<offset>` |
| `url` | `DB_URL` | `x/appdb` | `x/<value:DB_NAME>` |
| `url` | `MEDIA_URL` | `y/appdb-media` | `y/<value:MEDIA_BUCKET>` |
| `url` | `WEB_URL` | `http://localhost:8080` | `http://localhost:<value:API_PORT>` |' "$VALID_CMD"
run_guard "$REPO"
assert_ok "references to database, bucket and port rows all resolve"

# 8f. Two rows holding the same `Variable`. The contract resolves a reference to
#     "the row whose `Variable` column holds `VARIABLE`" — singular — so two
#     rows holding it mean the reference names no unique row, and a guard that
#     took the first would resolve to whichever the author happened to write
#     first.
new_project
write_iso '| `port` | `API_PORT` | `8080` | `+<offset>` |
| `port` | `API_PORT` | `8081` | `+<offset>` |' "$VALID_CMD"
run_guard "$REPO"
assert_invalid "two rows holding the same Variable is a violation"
assert_reports "API_PORT" "the duplicate-Variable report names the variable"

# ---------------------------------------------------------------------------
# 9. How many times each `Resource` word may appear. `port` and `url` may
#    repeat — once per port the project publishes, once per composed value it
#    declares — while `database`, `bucket` and `cache index` each appear at most
#    once.
# ---------------------------------------------------------------------------

# 9a. Two `database` rows, under two different variables, so the duplicate-
#     Variable rule cannot be what catches this.
new_project
write_iso '| `database` | `DB_URL` | `appdb` | `appdb_<id_underscored>` |
| `database` | `REPORTING_DB_URL` | `appreport` | `appreport_<id_underscored>` |' "$VALID_CMD"
run_guard "$REPO"
assert_invalid "a second database row is a violation"
assert_reports "REPORTING_DB_URL" "the second-database report names the row"

# 9b. Two `bucket` rows.
new_project
write_iso '| `bucket` | `MEDIA_BUCKET` | `appdb-media` | `appdb-media-<id>` |
| `bucket` | `BACKUP_BUCKET` | `appdb-backup` | `appdb-backup-<id>` |' "$VALID_CMD"
run_guard "$REPO"
assert_invalid "a second bucket row is a violation"
assert_reports "BACKUP_BUCKET" "the second-bucket report names the row"

# 9c. Two `cache index` rows. Only one index is claimed by probing.
new_project
write_iso '| `cache index` | `CACHE_INDEX` | `0` | `probed` |
| `cache index` | `SESSION_INDEX` | `1` | `probed` |' "$VALID_CMD"
run_guard "$REPO"
assert_invalid "a second cache index row is a violation"
assert_reports "SESSION_INDEX" "the second-cache-index report names the row"

# 9d. Several `port` and several `url` rows, which is the ordinary shape of a
#     real declaration and must not be reported.
new_project
write_iso '| `port` | `GATEWAY_PORT` | `8080` | `+<offset>` |
| `port` | `SERVER_PORT` | `8081` | `+<offset>` |
| `port` | `ADMIN_PANEL_PORT` | `3001` | `+<offset>` |
| `url` | `BACKEND_URL` | `http://localhost:8081` | `http://localhost:<value:SERVER_PORT>` |
| `url` | `FRONTEND_URL` | `http://localhost:3000` | `http://localhost:3000` |' "$VALID_CMD"
run_guard "$REPO"
assert_ok "several port and url rows are legal"

# ---------------------------------------------------------------------------
# 10. The command table: three verbs, two columns.
# ---------------------------------------------------------------------------

# 10a. A fourth verb. `/myflow-finish` calls `remove` and `survivors` by name
#      and nothing calls anything else, so a row named otherwise is a command
#      nobody runs — which reads to its author as a command that ran.
new_project
write_iso "$VALID_RES" '| `create` | `./scripts/workspace create` |
| `verify` | `./scripts/workspace verify` |'
run_guard "$REPO"
assert_invalid "a command verb outside the three is a violation"
assert_reports "verify" "the unknown-verb report names the verb"

# 10b. The same verb twice, which two different commands can hide behind.
new_project
write_iso "$VALID_RES" '| `remove` | `./scripts/workspace remove` |
| `remove` | `./scripts/other remove` |'
run_guard "$REPO"
assert_invalid "the same command verb twice is a violation"
assert_reports "remove" "the duplicate-verb report names the verb"

# 10c. An empty `Runs` cell. A verb declared with no command is a step the
#      pipeline would run as the empty string.
new_project
write_iso "$VALID_RES" '| `create` | `` |'
run_guard "$REPO"
assert_invalid "an empty Runs cell is a violation"
assert_reports "create" "the empty-Runs report names the verb"

# 10d. A token the pipeline does not substitute, which would otherwise reach a
#      shell as a literal nobody intended.
new_project
write_iso "$VALID_RES" '| `remove` | `docker exec <container> dropdb <id_underscored>` |'
run_guard "$REPO"
assert_invalid "an unsubstituted token in a command is a violation"
assert_reports "<container>" "the command-token report names the token"

# 10e. `<value:…>` is NOT substituted in a command: a command that needs a
#      derived value reads the exported variable, which is already in its
#      environment.
new_project
write_iso "$VALID_RES" '| `remove` | `./scripts/workspace remove <value:API_PORT>` |'
run_guard "$REPO"
assert_invalid "a <value:…> reference in a command is a violation"
assert_reports "<value:API_PORT>" "the command-reference report names the token"

# 10f. A project that declares `create` and `remove` and no `survivors` is a
#      supported state — the verification is reported as skipped rather than
#      passed, and that is /myflow-finish run 2 to decide, not this guard.
new_project
write_iso "$VALID_RES" '| `create` | `./scripts/workspace create` |
| `remove` | `./scripts/workspace remove` |'
run_guard "$REPO"
assert_ok "a command table with no survivors row is legal"

# 10g. An ordinary shell redirection is not a token. `<` and `>` around a
#      filename is how a real command is written, and a guard whose token shape
#      admitted whitespace would report it.
new_project
write_iso "$VALID_RES" '| `survivors` | `./scripts/workspace survivors < /dev/null > /tmp/out.txt` |'
run_guard "$REPO"
assert_ok "a shell redirection in a command is not read as a token"

# 10h. A `|` inside a command cell, written `\|`, is part of the command and not
#      a column boundary. The contract names filtering the project own tooling
#      as the reason a command contains one, so this is the shape a real
#      project writes rather than an exotic one.
new_project
write_iso "$VALID_RES" '| `survivors` | `./gradlew -q survivors \| awk "/appdb/"` |'
run_guard "$REPO"
assert_ok "an escaped pipe inside a command cell does not split the row"

# 10i. A stray space inside a command's token, which the pipeline then does not
#      substitute — the same blind spot as 7k, on the cell that is EXECUTED.
new_project
write_iso "$VALID_RES" '| `remove` | `./scripts/workspace remove < id>` |'
run_guard "$REPO"
assert_invalid "a space inside a command token is a violation"
assert_reports "< id>" "the command spaced-token report quotes the run as written"

# 10j. The redirection case again, now that a whitespace-carrying bracketed run
#      can be reported at all. `cmd < in.txt > out.txt` is a bracketed run holding
#      only token characters and a space, so a rule keyed on "bracketed and
#      spaced" would report it — and a guard that cried wolf on real commands
#      would be worked around rather than fixed. 10g covers `/dev/null`, whose
#      slashes fall outside the token characters on their own; this one does not,
#      so it is the case that actually pins the rule.
new_project
write_iso "$VALID_RES" '| `survivors` | `./scripts/workspace survivors < in.txt > out.txt` |'
run_guard "$REPO"
assert_ok "a redirection whose filename is token-shaped is still not a token"

# ---------------------------------------------------------------------------
# 11. Declare the section at most once. Two sections are an ambiguous
#     declaration, not a merge: they can name two different commands against two
#     different services.
# ---------------------------------------------------------------------------

# 11a. The heading twice, which a bad merge or a copied block produces. The two
#      sections carry DIFFERENT verbs, so nothing else in this fixture is
#      malformed: with the same verb in both, the duplicate-verb rule would
#      report the violation and this case would pass without the heading rule
#      existing at all.
new_project
write_config "# fixture

## workspace isolation

| Command | Runs |
|---------|------|
| \`create\` | \`./one create\` |

## workspace isolation

| Command | Runs |
|---------|------|
| \`survivors\` | \`./two survivors\` |
"
run_guard "$REPO"
assert_invalid "a duplicated heading is a violation"
assert_reports "2" "the duplicate-heading report names the count"
assert_reports "workspace isolation" "the duplicate-heading report names the heading"

# 11b. The count is what the operator has to fix, so three is reported as three
#      rather than as "more than one".
new_project
write_config "# fixture

## workspace isolation

## workspace isolation

## Workspace Isolation
"
run_guard "$REPO"
assert_invalid "three headings are a violation"
assert_reports "3" "the duplicate-heading report counts headings case-insensitively"

# 11c. A `### workspace isolation` is a different heading and is not the
#      section. Counting it would report a duplicate over a subsection an author
#      wrote inside their own prose.
new_project
write_config "# fixture

## workspace isolation

| Command | Runs |
|---------|------|
| \`survivors\` | \`./one survivors\` |

### workspace isolation notes

Prose that is not a second declaration.
"
run_guard "$REPO"
assert_ok "a deeper heading is not a second declaration"

# ---------------------------------------------------------------------------
# 12. The shape of the two tables. A shape this guard does not recognise must
#     never pass as valid: the resource table has four columns IN THIS ORDER,
#     and a table whose header says otherwise is a table whose cells mean
#     something else.
# ---------------------------------------------------------------------------

# 12a. A missing column. The rows below it then line up one column to the left,
#      so every `Default` is read as an `In a workspace` cell.
new_project
write_config "# fixture

## workspace isolation

| Resource | Variable | In a workspace |
|----------|----------|----------------|
| \`port\` | \`API_PORT\` | \`+<offset>\` |
"
run_guard "$REPO"
assert_invalid "a resource table missing a column is a violation"
# The report has to carry the order the columns were meant to be in. Naming only
# the header it found tells the author their table is wrong without telling them
# what right looks like, and the two differ by one column out of four.
assert_reports "Resource | Variable | Default | In a workspace" "the header report names the order it expected"

# 12b. A reordered header. Every cell is present and every one means something
#      else — the failure a guard that only counted columns would pass.
new_project
write_config "# fixture

## workspace isolation

| Resource | Variable | In a workspace | Default |
|----------|----------|----------------|---------|
| \`port\` | \`API_PORT\` | \`+<offset>\` | \`8080\` |
"
run_guard "$REPO"
assert_invalid "a reordered resource header is a violation"

# 12c. A header this guard does not recognise at all. Skipping it silently is
#      the fail-open this guard exists to prevent: the section would then
#      declare nothing and pass.
new_project
write_config "# fixture

## workspace isolation

| Thing | Value |
|-------|-------|
| \`database\` | \`appdb\` |
"
run_guard "$REPO"
assert_invalid "an unrecognised table header is a violation"

# 12d. A data row with the wrong number of cells. The row cannot be read against
#      a header it does not line up with, and reading its short tail as an empty
#      `In a workspace` cell would report the wrong rule broken.
new_project
write_iso '| `port` | `API_PORT` | `8080` |' "$VALID_CMD"
run_guard "$REPO"
assert_invalid "a resource row with too few cells is a violation"
# The needle is the COUNT, not the word "cell". A short row also leaves an empty
# `In a workspace` cell, and that rule reports the word — so a case keyed on it
# would pass while nothing counted the columns and the operator was sent to fix
# a cell that is missing rather than empty.
assert_reports "expected 4" "the short-row report names the cell count it expected"

# 12e. A section carrying no table at all. A heading with nothing under it is
#      how a mistyped header actually presents, and passing it would report a
#      project isolated that declared nothing.
new_project
write_config "# fixture

## workspace isolation

Prose describing what this project isolates, and no tables at all.
"
run_guard "$REPO"
assert_invalid "a section with neither table is a violation"

# 12f. A resource table with no command table beside it. This is a partial
#      declaration rather than a malformed one — the contract already supports
#      `create` and `remove` with no `survivors`, and treats a command that is
#      not declared as one that is not run.
new_project
write_config "# fixture

## workspace isolation

| Resource | Variable | Default | In a workspace |
|----------|----------|---------|----------------|
| \`port\` | \`API_PORT\` | \`8080\` | \`+<offset>\` |
"
run_guard "$REPO"
assert_ok "a resource table with no command table is a partial declaration, not a violation"

# 12g. The two tables written back to back, with no blank line and no prose
#      between them. Markdown separates tables with a blank line, so without one
#      they are a single contiguous run of rows to the scanner — the parser read
#      the resource header and evaluated every command row against four columns,
#      emitting four "cell count mismatch" violations and validating neither
#      table. It failed closed, which was right; it pointed at four lines that
#      were each individually correct, which was not. The assertion is that the
#      report names the real cause, and that the four misleading lines are gone.
new_project
write_config "# fixture

## workspace isolation

| Resource | Variable | Default | In a workspace |
|----------|----------|---------|----------------|
| \`port\` | \`API_PORT\` | \`8080\` | \`+<offset>\` |
| Command | Runs |
|---------|------|
| \`create\` | \`./scripts/workspace create\` |
| \`remove\` | \`./scripts/workspace remove <id>\` |
| \`survivors\` | \`./scripts/workspace survivors <id>\` |
"
run_guard "$REPO"
assert_invalid "two tables with no blank line between them is a violation"
assert_reports "a second table header" "the merged-tables report names the real cause"
assert_reports "blank line" "the merged-tables report says what to do about it"
assert_not_reported "cell(s) where its header has" "the merged-tables report drops the misleading cell-count lines"

# ---------------------------------------------------------------------------
# 13. The parser hazards. The input is tracked in the repository and editable in
#     any pull request, so none of these may change what a row is read as.
# ---------------------------------------------------------------------------

# 13a. A `|` inside a resource cell, written `\|`. A field split would cut the
#      row there and leave a shorter row that still looks well formed.
new_project
write_iso '| `url` | `WEB_URL` | `http://localhost:3000/?a=1\|2` | `http://localhost:3000/?a=1\|2` |' "$VALID_CMD"
run_guard "$REPO"
assert_ok "an escaped pipe inside a resource cell does not split the row"

# 13b. Rows with no trailing `|`, which Markdown permits and an author writes.
new_project
write_config "# fixture

## workspace isolation

| Resource | Variable | Default | In a workspace
|----------|----------|---------|----------------
| \`port\` | \`API_PORT\` | \`8080\` | \`+<offset>\`
"
run_guard "$REPO"
assert_ok "rows with no trailing pipe are read correctly"

# 13c. A file with CRLF line endings. Without the strip, `+<offset>\r` is not
#      `+<offset>` and every port row in the file is reported.
new_project
mkdir -p "$REPO/.flow"
printf '# fixture\r\n\r\n## workspace isolation\r\n\r\n| Resource | Variable | Default | In a workspace |\r\n|---|---|---|---|\r\n| `port` | `API_PORT` | `8080` | `+<offset>` |\r\n| `cache index` | `CACHE_INDEX` | `0` | `probed` |\r\n' > "$REPO/.flow/project.md"
run_guard "$REPO"
assert_ok "a CRLF file is read the same as an LF one"

# 13d. Indented rows, which a nested list or a quoted block produces.
new_project
write_config "# fixture

## workspace isolation

  | Resource | Variable | Default | In a workspace |
  |----------|----------|---------|----------------|
  | \`port\` | \`API_PORT\` | \`8080\` | \`+<offset>\` |
"
run_guard "$REPO"
assert_ok "indented rows are read correctly"

# 13e. Nothing in the file is EXECUTED. A cell that would be a command
#      substitution, a redirection or a rm if it ever reached a shell must leave
#      no trace — this guard validates text and runs nothing.
new_project
CANARY="$WORK/canary-$$"
write_iso "$VALID_RES" "| \`create\` | \`touch $CANARY\` |
| \`remove\` | \`\$(touch $CANARY.sub)\` |"
run_guard "$REPO"
[ ! -e "$CANARY" ] && pass "a declared command is never run" \
  || fail "the guard ran a declared command: $CANARY exists"
[ ! -e "$CANARY.sub" ] && pass "a command substitution in a cell is never evaluated" \
  || fail "the guard evaluated a command substitution: $CANARY.sub exists"

# 13f. A table OUTSIDE the section is not read. Every other key in the file is
#      free-form, and a `## apps` table has four columns of its own.
new_project
write_config "# fixture

## apps

| App | Repo root | Kind | URL |
|-----|-----------|------|-----|
| demo | \`/tmp/demo\` | Kotlin | http://localhost:8080 |

## workspace isolation

| Resource | Variable | Default | In a workspace |
|----------|----------|---------|----------------|
| \`port\` | \`API_PORT\` | \`8080\` | \`+<offset>\` |

## lint

| Command | Runs |
|---------|------|
| \`nonsense\` | \`\` |
"
run_guard "$REPO"
assert_ok "tables outside the section are not read"

# 13g. A cell carrying an ANSI escape sequence. Every violation line quotes a cell
#      straight out of the file, so a cell holding `\033[2K` erases the line it is
#      printed on and one holding `\033[1;32m` turns the next one bold green — a
#      report whose bytes say ISOLATION-INVALID and whose rendering shows a
#      convincing ISOLATION-OK. The file is tracked and editable in any pull
#      request, which is this guard's documented trust boundary.
#
#      The assertion is on the BYTES: no raw ESC anywhere in the output, the
#      escape rendered visibly instead, and the verdict still INVALID. Escaped
#      rather than stripped, because the operator is being sent to find this cell
#      and a rewritten name is a name that is not in their file.
new_project
mkdir -p "$REPO/.flow"
printf '# fixture\n\n## workspace isolation\n\n| Resource | Variable | Default | In a workspace |\n|---|---|---|---|\n| `queue\033[2K\033[1;32mISOLATION-OK: all good\033[0m` | `Q` | `x` | `x-<id>` |\n' \
  > "$REPO/.flow/project.md"
run_guard "$REPO"
assert_invalid "a cell carrying an escape sequence is still reported as invalid"
case "$OUT" in
  *$'\033'*) fail "an ESC byte from the configuration reached stdout: $OUT" ;;
  *) pass "no raw ESC byte from the configuration reaches stdout" ;;
esac
assert_reports '\x1b[2K' "the escape sequence is rendered visibly rather than stripped"
assert_reports "ISOLATION-OK: all good" "the cell's own text is still quoted back to the operator"

# 13g2. THE ENCODING IS INJECTIVE — the property the escaper's header argues for
#      and the one 13g cannot see. 13g proves a real ESC byte is rendered rather
#      than passed through; it says nothing about whether that rendering is
#      REVERSIBLE. If `\` were left unescaped, a cell holding the four literal
#      characters `\x1b` and a cell holding a real ESC byte would produce a
#      byte-identical report, and the operator sent to find the offending cell
#      could not tell which of the two is in their file — the same forged-report
#      class 13g closes, reached by collision instead of by pass-through.
#
#      DRIVEN AS A COLLISION, not as a rendering. The two configurations differ
#      in that one byte alone and are written into the SAME project root, so the
#      two reports are comparable in full and the assertion is exactly the
#      contract's word: distinct inputs, distinct reports. The rendering
#      assertion beside it names WHICH way they differ, so a failure says
#      whether the escape is missing or the report changed shape.
new_project
mkdir -p "$REPO/.flow"
ISO_ESC_HEAD='# fixture

## workspace isolation

| Resource | Variable | Default | In a workspace |
|---|---|---|---|
'
printf '%s| `queue\033[2K` | `Q` | `x` | `x-<id>` |\n' "$ISO_ESC_HEAD" \
  > "$REPO/.flow/project.md"
run_guard "$REPO"
assert_invalid "a cell holding a real ESC byte is reported as invalid"
ISO_ESC_REPORT="$OUT"
assert_not_reported '\\x1b' "a real ESC byte renders with a single backslash"

printf '%s| `queue\\x1b[2K` | `Q` | `x` | `x-<id>` |\n' "$ISO_ESC_HEAD" \
  > "$REPO/.flow/project.md"
run_guard "$REPO"
assert_invalid "a cell holding the four literal characters \\x1b is reported as invalid"
ISO_LITERAL_REPORT="$OUT"
assert_reports '\\x1b[2K' "a literal backslash in the cell is doubled in the report"

if [ "$ISO_ESC_REPORT" = "$ISO_LITERAL_REPORT" ]; then
  fail "the display encoding is not injective: a real ESC byte and the literal characters \\x1b produce the same report ($ISO_ESC_REPORT)"
else
  pass "a real ESC byte and the literal characters \\x1b produce different reports"
fi

# 13h. Non-ASCII is NOT escaped. Bytes 0x80-0xFF are the lead and continuation
#      bytes of ordinary UTF-8 and a terminal acts on none of them, so a project
#      whose bucket carries an accent must be reported by its real name — an
#      escape rule that mangled it would make the report unusable for exactly the
#      operators who need it.
new_project
write_iso '| `bücket` | `MEDIA_BUCKET` | `café-media` | `café-media-<id>` |' "$VALID_CMD"
run_guard "$REPO"
assert_invalid "an accented Resource word is reported"
assert_reports "bücket" "a non-ASCII cell is reported by its real name, not escaped"

# ---------------------------------------------------------------------------
# 14. The `.myflow/` -> `.flow/` hard cutover (design.md's
#     dotmyflow-hard-cutover). Built as real fixture project roots on disk,
#     run through the real guard as a real subprocess — never a mocked stat
#     — per REPRODUCE, DON'T READ. `~/Projects/gymie` and
#     `~/Projects/spectre-e2e` are in exactly the 14a shape on this machine
#     today and are deliberately not read here: they are not renamed by this
#     change, and a fixture this suite owns is what stays stable regardless
#     of their state.
# ---------------------------------------------------------------------------

# 14a. A project carrying only the retired `.myflow/` and no `.flow/` is NOT
#      the ordinary "declares nothing" case: the guard must refuse rather
#      than silently report ISOLATION-OK for a project it never actually
#      read, naming the project root and the exact rename to perform.
new_project
mkdir -p "$REPO/.myflow"
run_guard "$REPO"
assert_refuses "a project carrying only the retired .myflow/"
case "$ERR" in
  *"$REPO"*) pass "the refusal names the project root" ;;
  *) fail "the refusal does not name the project root: $ERR" ;;
esac
case "$ERR" in
  *"git -C $REPO mv .myflow .flow"*) pass "the refusal names the exact rename to perform" ;;
  *) fail "the refusal does not name the exact rename to perform: $ERR" ;;
esac

# 14b. A project carrying BOTH directories mid-cutover reads `.flow/`
#      without ever consulting `.myflow/` — the `.myflow/project.md` below
#      is deliberately malformed (a bare word, no `## workspace isolation`
#      heading token even close to well-formed) so that a guard which fell
#      back to it, read both, or preferred it would fail this case instead
#      of passing it.
new_project
mkdir -p "$REPO/.myflow"
printf 'if this file is read at all, the guard consulted the retired directory\n' > "$REPO/.myflow/project.md"
write_iso "$VALID_RES" "$VALID_CMD"
run_guard "$REPO"
assert_ok "a project carrying both directories reads .flow/ only"
assert_reports "resource row(s) and" "the guard validated .flow/project.md's own rows, not the retired directory's malformed content"

# ---------------------------------------------------------------------------
# 15. The validator itself failing to run. An empty report is what a crashed
#     parser and a clean file look like from the outside, and telling them apart
#     is the last fail-open in this guard.
# ---------------------------------------------------------------------------

# 15a. `awk` unavailable, or failing for any reason. The whole of the validation
#      happens inside one awk program, so a run in which it never produced its
#      report has checked NOTHING — and reporting that as ISOLATION-OK would
#      pass every malformed project on a machine where the parser is broken.
#      The shim exits non-zero and prints nothing, which is exactly the shape a
#      missing interpreter produces.
new_project
write_iso "$VALID_RES" "$VALID_CMD"
SHIM="$WORK/shim"
mkdir -p "$SHIM"
printf '#!/bin/sh\nexit 3\n' > "$SHIM/awk"
chmod +x "$SHIM/awk"
set +e
OUT="$(PATH="$SHIM:$PATH" "$GUARD" "$REPO" 2>"$ERRFILE")"
RC=$?
set -e
ERR="$(cat "$ERRFILE")"
assert_refuses "a validator that could not run"

# ---------------------------------------------------------------------------
# 16. No-argument default derives the repository root from this script's OWN
#     resolved location, not from a fixed "one level up above $SCRIPT_DIR" —
#     which only holds while it lives at <repo>/scripts/. Built here: a
#     scratch tree where the guard is reachable at two depths, its real home
#     (root/scripts/) and a skills/myflow-do/scripts/ symlink, mirroring how
#     setup.sh's install carries it. Invoked through the symlink with no
#     argument, it must find THAT tree's own .flow/project.md — never the
#     skill directory's, which has none and would silently read as "declares
#     nothing" instead (see design.md, "The $SCRIPT_DIR/.. hazard").
# ---------------------------------------------------------------------------
ROOT_TEST="$(mktemp -d "${TMPDIR:-/tmp}/workspace-isolation-root-test.XXXXXX")"
SANDBOXES+=("$ROOT_TEST")
mkdir -p "$ROOT_TEST/scripts/lib" "$ROOT_TEST/skills/myflow-do/scripts"
cp "$GUARD" "$ROOT_TEST/scripts/check-workspace-isolation.sh"
chmod +x "$ROOT_TEST/scripts/check-workspace-isolation.sh"
cp "$REPO_ROOT/scripts/lib/resolve-file.sh" "$ROOT_TEST/scripts/lib/resolve-file.sh"
ln -s ../../../scripts/check-workspace-isolation.sh \
  "$ROOT_TEST/skills/myflow-do/scripts/check-workspace-isolation.sh"
ln -s ../../../scripts/lib "$ROOT_TEST/skills/myflow-do/scripts/lib"
REPO="$ROOT_TEST"
write_iso "$VALID_RES" "$VALID_CMD"
set +e
OUT="$("$ROOT_TEST/skills/myflow-do/scripts/check-workspace-isolation.sh" 2>"$ERRFILE")"
RC=$?
set -e
ERR="$(cat "$ERRFILE")"
assert_ok "case 16: no-arg default resolves through a skill-dir symlink to the real repo root"
assert_reports "resource row(s) and" "case 16: the real .flow/project.md was read and validated"
assert_not_reported "no .flow/project.md" "case 16: did not fall back to the skill directory, which has none"

# ---------------------------------------------------------------------------
# 17. F1/F11 regression: an explicit project-root argument — the shape every
#     real caller uses (`/myflow-do` against each apply worktree; this
#     repository's own `## lint` entry against $REPO_ROOT) — must never
#     depend on resolving this script's OWN location. Before F11's fix, that
#     resolution ran UNCONDITIONALLY at the top of the file, so a failure
#     there aborted every call, even one that never reads REPO_ROOT at all.
#     Reproduced by rigging PATH so `readlink` always fails, then invoking
#     the guard THROUGH a symlink (so resolve_file's [ -L ] test is true and
#     it actually calls the rigged readlink) WITH an explicit project root.
#     Watched fail first: before the fix this case reported
#     "cannot resolve this script's own location" and exit 2 even though a
#     perfectly good project root was given on the command line.
# ---------------------------------------------------------------------------
LAZY_TEST="$(mktemp -d "${TMPDIR:-/tmp}/workspace-isolation-lazy-test.XXXXXX")"
SANDBOXES+=("$LAZY_TEST")
mkdir -p "$LAZY_TEST/scripts/lib" "$LAZY_TEST/skills/myflow-do/scripts"
cp "$GUARD" "$LAZY_TEST/scripts/check-workspace-isolation.sh"
chmod +x "$LAZY_TEST/scripts/check-workspace-isolation.sh"
cp "$REPO_ROOT/scripts/lib/resolve-file.sh" "$LAZY_TEST/scripts/lib/resolve-file.sh"
ln -s ../../../scripts/check-workspace-isolation.sh \
  "$LAZY_TEST/skills/myflow-do/scripts/check-workspace-isolation.sh"
ln -s ../../../scripts/lib "$LAZY_TEST/skills/myflow-do/scripts/lib"

LAZY_PROJECT="$(mktemp -d "${TMPDIR:-/tmp}/workspace-isolation-lazy-project.XXXXXX")"
SANDBOXES+=("$LAZY_PROJECT")
REPO="$LAZY_PROJECT"
write_iso "$VALID_RES" "$VALID_CMD"

BADPATH_DIR="$(mktemp -d "${TMPDIR:-/tmp}/workspace-isolation-badpath.XXXXXX")"
SANDBOXES+=("$BADPATH_DIR")
printf '#!/bin/sh\nexit 1\n' > "$BADPATH_DIR/readlink"
chmod +x "$BADPATH_DIR/readlink"

set +e
OUT="$(PATH="$BADPATH_DIR:$PATH" "$LAZY_TEST/skills/myflow-do/scripts/check-workspace-isolation.sh" "$LAZY_PROJECT" 2>"$ERRFILE")"
RC=$?
set -e
ERR="$(cat "$ERRFILE")"
assert_ok "case 17 (F11): an explicit project root argument never triggers self-location resolution, so a rigged readlink cannot abort the run"
assert_reports "resource row(s) and" "case 17 (F11): the given project root was validated, not skipped"
case "$ERR" in
  *"cannot resolve this script's own location"*)
    fail "case 17 (F11): self-resolution was attempted even though an explicit root was given: $ERR" ;;
  *) pass "case 17 (F11): self-resolution was never attempted" ;;
esac

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'check-workspace-isolation: all cases pass\n'
