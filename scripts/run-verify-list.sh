#!/usr/bin/env bash
# run-verify-list.sh — run one of a project's verify lists the way **Inline
# verify** (skills/flow/verify-and-handoff.md) prescribes, and print its
# `## Report` lines, so no run hand-builds the runner (the
# self-review-backlog-flow-6 self-review pass found the parent writing one
# per run, its stdin detach patched in on a second call).
#
# Usage: run-verify-list.sh <worktree> lint|test
#
# THE LIST is every non-blank, non-`#` line inside a ```bash fence of
# `project-get.sh <worktree> <list>`'s output, in the order printed; `lint`
# appends `check-spec-reach.sh <worktree>`. A `# inputs: <pathspec>...`
# line of the `test` list declares what the one command below it reads; every
# other `#` line is a comment. A line holding a `<word>`
# placeholder — `./gradlew verifyChange :<module>:test --tests '<selectors>'`
# — is a template, not a command: the shell reads `<module>` as a redirect,
# so the list is refused before anything runs (kan-983 self-review). Each command runs from
# <worktree> in a fresh `bash -c`, stdin detached (</dev/null) so none waits
# on a terminal, and the whole list always runs: a failure never stops it.
# The caller exports the KEY=value lines prepare-workspace.sh printed for
# <worktree> first; every command inherits them. This script retries
# nothing — the one inline re-run of a failing command stays the caller's.
#
# THE GREEN SHA (KAN-887's self-review: verify re-ran the whole list minutes
# after a green full-suite run, the delta a test rename). A `test` run that
# exits 0 on a tree clean before and after it — `git status --porcelain
# --untracked-files=all` empty — writes HEAD to the worktree's own git dir
# (`git rev-parse --git-path flow-test-green`), never into the tree. A later
# `test` run on a clean tree skips every command when HEAD is that sha, and
# otherwise skips each command whose `# inputs:` pathspecs, plus
# `.flow/project.md` itself (an edited command must run), `git diff --quiet
# <sha> HEAD` finds untouched. A command declaring no inputs reads everything
# and always runs; a dirty tree, no git, no recorded sha or a failing diff
# runs the whole list. A skip passes and prints instead of output:
#   - `<command>` — skipped, no input changed since <sha> passed the list
# ponytail: the sha ignores the exported KEY=value lines; a run under other
# isolation values reuses the green, so key the file by them if they ever vary.
#
# OUTPUT, one block per command in the order run, stdout and stderr together:
#   - `<command>` — exit <n>
#     <each output line, indented two spaces; past 20 lines,
#     `(truncated — last 20 of <N> lines)` and then the last 20>
#
# Exit codes:
#   0  every command exited 0
#   1  at least one command exited non-zero
#   2  usage (not exactly `<worktree> lint|test`, or <worktree> is not a
#      directory), a list line holding a `<word>` placeholder — its stderr
#      line names it — or project-get.sh exit 2 — its stderr line says why
#   3  the list is not declared (project-get.sh exit 1): the caller
#      auto-detects, as **Inline verify** says
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"

if [ "$#" -ne 2 ] || [ ! -d "$1" ] || { [ "$2" != lint ] && [ "$2" != test ]; }; then
  echo "usage: run-verify-list.sh <worktree> lint|test" >&2
  exit 2
fi
WT="$1"
LIST="$2"

BODY="$("$SCRIPT_DIR/project-get.sh" "$WT" "$LIST")"
case $? in
  0) ;;
  1) exit 3 ;;
  *) exit 2 ;;
esac

# INPUTS[i] is the pathspec list the `# inputs:` line above CMDS[i] named.
CMDS=()
INPUTS=()
PENDING=""
while IFS= read -r line; do
  if [[ "$line" =~ ^[[:space:]]*#\ inputs:(.*)$ ]]; then
    PENDING="${BASH_REMATCH[1]}"
    continue
  fi
  CMDS+=("$line")
  INPUTS+=("$PENDING")
  PENDING=""
done < <(printf '%s\n' "$BODY" | awk '
  /^```bash[[:space:]]*$/ { f = 1; next }
  /^```/                  { f = 0; next }
  f && NF && ($1 !~ /^#/ || /^[[:space:]]*# inputs:/)
')

for i in "${!CMDS[@]}"; do
  if grep -Eq '<[A-Za-z][A-Za-z0-9_-]*>' <<<"${CMDS[$i]}"; then
    echo "run-verify-list.sh: $LIST line is a template, not a command — move it out of the fence: ${CMDS[$i]}" >&2
    exit 2
  fi
done

HEAD_SHA=""
SINCE=""
GREEN_FILE=""
if [ "$LIST" = test ] && HEAD_SHA="$(git -C "$WT" rev-parse -q --verify HEAD 2>/dev/null)" &&
  STATUS="$(git -C "$WT" status --porcelain --untracked-files=all 2>/dev/null)" && [ -z "$STATUS" ] &&
  GREEN_FILE="$(git -C "$WT" rev-parse --path-format=absolute --git-path flow-test-green 2>/dev/null)"; then
  SINCE="$(cat "$GREEN_FILE" 2>/dev/null || true)"
else
  HEAD_SHA=""
fi

# unchanged <inputs> -> 0 when a recorded green sha covers the command.
unchanged() {
  local specs
  [ -n "$SINCE" ] || return 1
  [ "$SINCE" = "$HEAD_SHA" ] && return 0
  [ -n "${1// /}" ] || return 1
  read -ra specs <<<"$1"
  git -C "$WT" diff --quiet "$SINCE" "$HEAD_SHA" -- "${specs[@]}" .flow/project.md 2>/dev/null
}

OUT="$(mktemp "${TMPDIR:-/tmp}/run-verify-list.XXXXXX")"
trap 'rm -f "$OUT"' EXIT

FAILED=0
# run <command-as-shown> <command-to-run>
run() {
  local rc n
  (cd "$WT" && bash -c "$2") </dev/null >"$OUT" 2>&1
  rc=$?
  [ "$rc" -eq 0 ] || FAILED=1
  printf -- '- `%s` — exit %d\n' "$1" "$rc"
  n="$(awk 'END{print NR}' "$OUT")"
  if [ "$n" -gt 20 ]; then
    printf '  (truncated — last 20 of %d lines)\n' "$n"
    tail -n 20 "$OUT" | awk '{print "  " $0}'
  else
    awk '{print "  " $0}' "$OUT"
  fi
}

for i in "${!CMDS[@]}"; do
  if unchanged "${INPUTS[$i]}"; then
    printf -- '- `%s` — skipped, no input changed since %s passed the list\n' "${CMDS[$i]}" "$SINCE"
    continue
  fi
  run "${CMDS[$i]}" "${CMDS[$i]}"
done
if [ -n "$HEAD_SHA" ] && [ "$FAILED" -eq 0 ] &&
  STATUS="$(git -C "$WT" status --porcelain --untracked-files=all 2>/dev/null)" && [ -z "$STATUS" ]; then
  printf '%s\n' "$HEAD_SHA" >"$GREEN_FILE"
fi
if [ "$LIST" = lint ]; then
  run "check-spec-reach.sh $WT" "\"$SCRIPT_DIR/check-spec-reach.sh\" \"$WT\""
fi

exit "$FAILED"
