#!/usr/bin/env bash
# Assertion harness for run-verify-list.sh.
#
# Builds throwaway project roots under TMPDIR, each with a `.flow/project.md`
# whose `## lint` / `## test` fences hold small shell commands, and runs the
# REAL scripts/run-verify-list.sh against them — never a copy of its logic.
# The fixture roots are no git repositories, so project-get.sh reads the file
# on disk and check-spec-reach.sh answers `Spec reach: not configured`.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN="$SCRIPT_DIR/run-verify-list.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

ROOT="$(cd "$(mktemp -d "${TMPDIR:-/tmp}/run-verify-list-root.XXXXXX")" && pwd)"
trap 'rm -rf "$ROOT"' EXIT
mkdir -p "$ROOT/.flow"
cat > "$ROOT/.flow/project.md" <<'CFG'
## lint

Prose outside the fence is never run.

```bash
# a comment line is never run
printf 'one\n'
printf 'to-stderr\n' >&2; exit 3
cat
seq 1 25
pwd
```

## test

```bash
printf 'ok\n'
printf nonl
echo next
```
CFG

# run_bin <args...> -> OUT, RC; stdin carries a line a command must never read.
run_bin() {
  set +e
  OUT="$(printf 'LEAKED-STDIN\n' | "$BIN" "$@" 2>&1)"
  RC=$?
  set -e
}

expect_rc() { if [ "$RC" -eq "$2" ]; then pass "$1: exit $2"; else fail "$1: exit $RC, want $2"; fi; }
expect_has() { case "$OUT" in *"$2"*) pass "$1" ;; *) fail "$1: missing [$2] in: $OUT" ;; esac; }
expect_lacks() { case "$OUT" in *"$2"*) fail "$1: unexpected [$2] in: $OUT" ;; *) pass "$1" ;; esac; }

run_bin "$ROOT" lint
expect_rc "a failing command fails the list" 1
expect_has "first command reported" $'- `printf \'one\\n\'` — exit 0\n  one'
expect_has "stderr captured and exit kept" $'exit 3\n  to-stderr'
expect_has "the list runs on past a failure" '- `pwd` — exit 0'
expect_lacks "stdin is detached" 'LEAKED-STDIN'
expect_lacks "comments and prose never run" 'comment line'
expect_has "long output truncated to its last 20" $'(truncated — last 20 of 25 lines)\n  6\n'
expect_lacks "truncated head dropped" $'\n  5\n'
expect_has "commands run from the worktree" "  $ROOT"
expect_has "lint appends check-spec-reach.sh" $'- `check-spec-reach.sh '"$ROOT"$'` — exit 0\n  Spec reach: not configured'

run_bin "$ROOT" test
expect_rc "a passing list" 0
expect_has "test list reported" $'- `printf \'ok\\n\'` — exit 0\n  ok'
expect_has "output without a final newline ends its line" $'  nonl\n- `echo next` — exit 0'
expect_lacks "test never appends check-spec-reach.sh" 'check-spec-reach.sh'

printf '## lint\n\n```bash\ntrue\n```\n' > "$ROOT/.flow/project.md"
run_bin "$ROOT" test
expect_rc "an undeclared list" 3

printf '## test\n\n```bash\nprintf ran-first\n./gradlew verifyChange :<module>:test --tests %s\n```\n' "'<selectors>'" > "$ROOT/.flow/project.md"
run_bin "$ROOT" test
expect_rc "a placeholder line refuses the list" 2
expect_has "the refusal names the line" "test line is a template, not a command — move it out of the fence: ./gradlew verifyChange :<module>:test --tests '<selectors>'"
expect_lacks "nothing runs before the refusal" 'ran-first'

# THE GREEN SHA: a git worktree whose `test` list counts its own runs.
G="$ROOT/git"
mkdir -p "$G/.flow" "$G/web" "$G/go"
cat > "$G/.flow/project.md" <<'CFG'
## test

```bash
# inputs: web/
echo ran >> "$(git rev-parse --git-dir)/web-runs"; echo web
# inputs: go/
echo go
echo all
```
CFG
echo a > "$G/web/a"
echo b > "$G/go/b"
git -C "$G" init -q
git -C "$G" -c user.name=t -c user.email=t@t add -A
git -C "$G" -c user.name=t -c user.email=t@t commit -q -m one
SHA1="$(git -C "$G" rev-parse HEAD)"
run_bin "$G" test
expect_rc "a first green run" 0
expect_has "a first run runs every command" '- `echo all` — exit 0'
GREEN="$(git -C "$G" rev-parse --path-format=absolute --git-path flow-test-green)"
[ "$(cat "$GREEN" 2>/dev/null)" = "$SHA1" ] && pass "a green run records HEAD in the git dir" || fail "green sha: $(cat "$GREEN" 2>/dev/null)"
[ -z "$(git -C "$G" status --porcelain)" ] && pass "the record leaves the tree clean" || fail "the record dirtied the tree"
run_bin "$G" test
expect_rc "a re-run at the green sha" 0
expect_has "every command skipped at the green sha" "- \`echo all\` — skipped, no input changed since $SHA1 passed the list"
[ "$(wc -l < "$(git -C "$G" rev-parse --absolute-git-dir)/web-runs")" -eq 1 ] && pass "nothing re-ran at the green sha" || fail "a command re-ran at the green sha"
echo c > "$G/go/b"
git -C "$G" -c user.name=t -c user.email=t@t commit -q -am two
run_bin "$G" test
expect_has "a command whose inputs the delta skips is skipped" "- \`echo ran >> \"\$(git rev-parse --git-dir)/web-runs\"; echo web\` — skipped, no input changed since $SHA1"
expect_has "a command whose inputs the delta touches runs" $'- `echo go` — exit 0'
expect_has "a command declaring no inputs runs" $'- `echo all` — exit 0'
echo dirty > "$G/web/a"
run_bin "$G" test
expect_has "a dirty tree runs the whole list" $'echo web` — exit 0'
git -C "$G" checkout -q -- web/a
run_bin "$G" test
sed -i.bak 's/^echo go$/echo go2/' "$G/.flow/project.md" && rm "$G/.flow/project.md.bak"
git -C "$G" -c user.name=t -c user.email=t@t commit -q -am three
run_bin "$G" test
expect_has "an edited command runs though its inputs are untouched" $'- `echo go2` — exit 0'

run_bin "$ROOT" build
expect_rc "an unknown list" 2
run_bin "$ROOT/missing" lint
expect_rc "a missing worktree" 2

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'all cases passed\n'
