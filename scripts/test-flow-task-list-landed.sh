#!/usr/bin/env bash
# test-flow-task-list-landed.sh — runs the flow-task-list mod's landedIn
# (mods/flow-task-list/hooks/landed.ts) against a real git repository. The
# mod's own suite (`claude plugin test`) cannot spawn a process, so it only
# replays git's answers; this harness proves the argv it builds resolves the
# base and parses the Task-Id trailers the way real git answers them.
#
# The fixture: origin's main gains two other changes' Task-Id commits (5, 6)
# after the clone, so the local main is behind origin; origin's feat carries
# Task-Id 8. A change branch cut from origin/main with no recorded base must
# land only its own tasks (base refs/remotes/origin/HEAD); a change branch cut
# from origin/feat with branch.<it>.flowBase=feat must land only its own task
# (base refs/remotes/origin/feat); a directory that is no repository lands
# nothing.
#
# Exit codes: 0 every case passed, 1 a case failed.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LANDED="$SCRIPT_DIR/../mods/flow-task-list/hooks/landed.ts"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

g() { git -c user.name=t -c user.email=t@t -c commit.gpgsign=false -c init.defaultBranch=main "$@"; }
commit() { g -C "$1" commit -q --allow-empty -m "$2"; }

g init -q "$TMP/origin"
commit "$TMP/origin" init
g clone -q "$TMP/origin" "$TMP/work"
commit "$TMP/origin" $'other change\n\nTask-Id: 5'
g -C "$TMP/origin" checkout -q -b feat
commit "$TMP/origin" $'feat work\n\nTask-Id: 8'
g -C "$TMP/origin" checkout -q main
commit "$TMP/origin" $'more other\n\nTask-Id: 6'
g -C "$TMP/work" fetch -q origin

# The main checkout stays on its local main, behind origin; each change is a worktree beside it.
g -C "$TMP/work" worktree add -q -b change "$TMP/change" origin/main
commit "$TMP/change" $'task one\n\nTask-Id: 1'
commit "$TMP/change" $'tasks two and three\n\nTask-Id: 2+3'
g -C "$TMP/work" worktree add -q -b feat-change "$TMP/feat" origin/feat
g -C "$TMP/feat" config branch.feat-change.flowBase feat
commit "$TMP/feat" $'task four\n\nTask-Id: 4'

landed_in() {
  node --no-warnings --input-type=module -e '
    import { spawnSync } from "node:child_process"
    const { landedIn } = await import(process.argv[1])
    const run = async argv => {
      const r = spawnSync(argv[0], argv.slice(1), { encoding: "utf8" })
      return { exitCode: r.status ?? 1, stdout: r.stdout }
    }
    console.log((await landedIn(run, process.argv[2])).sort((a, b) => a - b).join(" "))
  ' "$LANDED" "$1"
}

fail=0
check() {
  local got
  got="$(landed_in "$2")"
  if [ "$got" = "$3" ]; then
    echo "ok   $1"
  else
    echo "FAIL $1: want '$3', got '$got'"
    fail=1
  fi
}

check "no recorded base: since refs/remotes/origin/HEAD" "$TMP/change" "1 2 3"
check "recorded flowBase: since refs/remotes/origin/<it>" "$TMP/feat" "4"
check "no repository: nothing landed" "$TMP" ""
exit "$fail"
