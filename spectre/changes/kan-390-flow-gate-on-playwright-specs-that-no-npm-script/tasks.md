# kan-390-flow-gate-on-playwright-specs-that-no-npm-script

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.
> **Relocation:** no

Three commits in dependency order: the guard and its harness (task 1), shipping it into the
symlink farm and the lists that name every shipped guard (task 2), and the two call sites in the
stage prose (task 3). The approved design is
`docs/superpowers/specs/2026-09-03-kan-390-flow-gate-on-playwright-specs-that-no-npm-script-design.md`;
`design.md` beside this file carries the decisions.

**Baseline, measured before any edit:**

- 42 guard harnesses under `scripts/test-*.sh`; `scripts/check-contract-budget.sh` reports 71 owned
  files within budget; `check-vocabulary.sh` and `check-references.sh` exit clean.
  <!-- measured: ls scripts/test-*.sh | wc -l; scripts/check-contract-budget.sh @ d31ebd9 -->
- `skills/flow/verify-and-handoff.md` is 26051 bytes against a 26100-byte budget row — task 3's
  prose outgrows it, so task 3 raises that row.
  <!-- measured: wc -c skills/flow/verify-and-handoff.md; grep verify-and-handoff scripts/check-contract-budget.sh @ d31ebd9 -->
- The drafted guard, run against `/Users/tweety53/Projects/gymie` (whose `regression checkout` is
  `/Users/tweety53/Projects/gymie-playwright`, real Playwright 1.62.1), enumerates 29 specs and
  reports one orphan, `pad-frames.spec.ts`, in 5.9s wall — a live finding, not a fixture. Fixing
  that checkout is outside this change; once task 3 lands, gymie's `flow.verify` blocks until
  `pad-frames.spec.ts` is reached by a script or removed.
  <!-- measured: time check-spec-reach.sh /Users/tweety53/Projects/gymie, the task 1 script run from a scratch copy with lib/ symlinked @ d31ebd9 -->

**Every task that grows an owned `.md` file runs `scripts/check-contract-budget.sh` and reconciles
its `budgets()` row in the same commit** — raise a row only when the guard actually fails on that
file, and only to the size its own rule gives (the new size plus 25%).

---

- [x] 1. Add `check-spec-reach.sh` and its harness

Create `scripts/check-spec-reach.sh` — the guard exactly as below, header included — and
`scripts/test-check-spec-reach.sh`. Both were authored and run for this plan: the harness passes
all 12 cases (26 assertions) against the guard, and the guard reproduces the live gymie finding in
the baseline above.

**Files:** `scripts/check-spec-reach.sh`, `scripts/test-check-spec-reach.sh`
**Tests:** `scripts/test-check-spec-reach.sh` — cases 1 through 17
**Regression:** reverting this commit removes the guard; tasks 2 and 3 then cite a basename that
resolves to nothing, and `check-guard-symlinks.sh` rule 1 reports the dangling symlink.
**Baseline:** before=0 after=17 cases (36 `ok:` assertions) in `scripts/test-check-spec-reach.sh`;
guard harnesses 42 → 43
<!-- measured: scripts/test-check-spec-reach.sh | grep -c '^ok:' on the scratch copy; ls scripts/test-*.sh | wc -l @ d31ebd9 -->
**Commit:** `feat(scripts): add check-spec-reach.sh, a gate on Playwright specs no npm script reaches`
**Build:** green

  - [x] **Step 1: Write the harness** as `scripts/test-check-spec-reach.sh`, `chmod +x`:

```bash verified:authored and run for this plan against the task 1 guard — 12 cases, 26 assertions pass
#!/usr/bin/env bash
# Assertion harness for check-spec-reach.sh.
#
# Builds throwaway project roots and regression checkouts under TMPDIR and
# runs the REAL scripts/check-spec-reach.sh against them — never a copy of its
# logic. Playwright is not installed in a fixture: a fake `npm` placed first
# on PATH answers `npm run -s <script> -- --list --reporter=json` with the
# canned JSON the fixture wrote as `<checkout>/.list-<script>.json` (a
# `.fail-<script>` file makes it exit 7 instead), so script selection, the
# union across scripts, orphan reporting and every exit code are exercised
# without a browser. `node` is real: the guard needs it to read package.json
# and the reporter JSON, and it is on every machine Playwright runs on.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-spec-reach.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

DIRS=()
cleanup() {
  [ "${#DIRS[@]}" -eq 0 ] && return 0
  local d
  for d in "${DIRS[@]}"; do rm -rf "$d"; done
}
trap cleanup EXIT

FAKE_BIN="$(mktemp -d "${TMPDIR:-/tmp}/check-spec-reach-bin.XXXXXX")"
DIRS+=("$FAKE_BIN")
cat > "$FAKE_BIN/npm" <<'NPM'
#!/usr/bin/env bash
# fake npm: `npm run -s <script> -- --list --reporter=json`, answered from cwd.
script="$3"
if [ -e ".fail-$script" ]; then echo "playwright: boom for $script" >&2; exit 7; fi
cat ".list-$script.json"
NPM
chmod +x "$FAKE_BIN/npm"
export PATH="$FAKE_BIN:$PATH"

# new_root -> ROOT (project root with .flow/) and CHECKOUT (regression checkout).
new_root() {
  ROOT="$(mktemp -d "${TMPDIR:-/tmp}/check-spec-reach-root.XXXXXX")"
  CHECKOUT="$(mktemp -d "${TMPDIR:-/tmp}/check-spec-reach-co.XXXXXX")"
  DIRS+=("$ROOT" "$CHECKOUT")
  mkdir -p "$ROOT/.flow"
  CFG="$ROOT/.flow/project.md"
}

# cfg_with_checkout [bom] — a minimal section naming $CHECKOUT.
cfg_with_checkout() {
  local prefix=""
  [ "${1:-}" = "bom" ] && prefix=$'\xef\xbb\xbf'
  printf '%s## visual verification\n\n| Setting | Value |\n|---|---|\n| `screenshots` | `.` |\n| `regression checkout` | `%s` |\n' "$prefix" "$CHECKOUT" > "$CFG"
}

# package_json <script>=<command> ...
package_json() {
  local entries="" kv
  for kv in "$@"; do
    entries="$entries\"${kv%%=*}\": \"${kv#*=}\", "
  done
  printf '{ "scripts": { %s } }\n' "${entries%, }" > "$CHECKOUT/package.json"
}

# listing <script> <file>... — canned reporter JSON for one script.
listing() {
  local script="$1"; shift
  local suites="" f
  for f in "$@"; do suites="$suites{\"file\": \"$f\", \"suites\": []}, "; done
  printf '{ "config": { "rootDir": "%s" }, "suites": [ %s ] }\n' "$CHECKOUT" "${suites%, }" > "$CHECKOUT/.list-$script.json"
}

spec() { mkdir -p "$(dirname "$CHECKOUT/$1")"; : > "$CHECKOUT/$1"; }

run_guard() {
  set +e
  OUT="$("$GUARD" "$ROOT" 2>&1)"
  RC=$?
  set -e
}

assert_rc() {
  local case_name="$1" want="$2"
  [ "$RC" -eq "$want" ] && pass "$case_name: exit $want" \
    || fail "$case_name: expected exit $want, got rc=$RC out=$OUT"
}

assert_out_contains() {
  local case_name="$1" needle="$2"
  case "$OUT" in
    *"$needle"*) pass "$case_name: output names '$needle'" ;;
    *) fail "$case_name: expected output to contain '$needle', got: $OUT" ;;
  esac
}

assert_out_not_contains() {
  local case_name="$1" needle="$2"
  case "$OUT" in
    *"$needle"*) fail "$case_name: expected output NOT to contain '$needle', got: $OUT" ;;
    *) pass "$case_name: output correctly omits '$needle'" ;;
  esac
}

# Case 1: no .flow/project.md — not configured, exit 0.
new_root; rm -f "$CFG"
run_guard; assert_rc "case 1" 0; assert_out_contains "case 1" "Spec reach: not configured"

# Case 2: no `## visual verification` section — exit 0.
new_root; printf '# Project\n\n## run\n\necho hi\n' > "$CFG"
run_guard; assert_rc "case 2" 0; assert_out_contains "case 2" "Spec reach: not configured"

# Case 3: section without a `regression checkout` row — exit 0.
new_root; printf '## visual verification\n\n| Setting | Value |\n|---|---|\n| `screenshots` | `.` |\n' > "$CFG"
run_guard; assert_rc "case 3" 0; assert_out_contains "case 3" "Spec reach: not configured"

# Case 4: section declared twice — ambiguous, exit 2.
new_root; cfg_with_checkout; printf '\n## visual verification\n' >> "$CFG"
run_guard; assert_rc "case 4" 2; assert_out_contains "case 4" "ambiguous"

# Case 5: checkout is not a directory — exit 2.
new_root; cfg_with_checkout; rm -rf "$CHECKOUT"
run_guard; assert_rc "case 5" 2; assert_out_contains "case 5" "not an existing directory"

# Case 6: checkout holds no package.json — exit 2.
new_root; cfg_with_checkout
run_guard; assert_rc "case 6" 2; assert_out_contains "case 6" "no package.json"

# Case 7: no script contains `playwright test` — exit 2.
new_root; cfg_with_checkout; package_json "test=vitest run"
run_guard; assert_rc "case 7" 2; assert_out_contains "case 7" "playwright test"

# Case 8: two scripts, union covers every spec — exit 0; a non-playwright
# script is ignored, a spec under node_modules is not enumerated.
new_root; cfg_with_checkout
package_json "test:e2e=playwright test tests/" "shots=playwright test shots.spec.ts --headed" "lint=eslint ."
spec tests/a.spec.ts; spec shots.spec.ts; spec node_modules/dep/x.spec.ts
listing "test:e2e" tests/a.spec.ts; listing shots shots.spec.ts
run_guard; assert_rc "case 8" 0; assert_out_contains "case 8" "Spec reach: 2 spec(s), all reached"

# Case 9: one orphan — exit 1, named relative to the checkout, reached ones silent.
new_root; cfg_with_checkout
package_json "test:e2e=playwright test tests/"
spec tests/a.spec.ts; spec kan-347-route-e.spec.ts
listing "test:e2e" tests/a.spec.ts
run_guard; assert_rc "case 9" 1
assert_out_contains "case 9" "kan-347-route-e.spec.ts: reached by no package.json script"
assert_out_not_contains "case 9" "tests/a.spec.ts:"
assert_out_contains "case 9" "1 of 2 spec(s)"

# Case 10: a `--list` run fails — exit 2, its output relayed.
new_root; cfg_with_checkout
package_json "test:e2e=playwright test tests/"
spec tests/a.spec.ts; : > "$CHECKOUT/.fail-test:e2e"
run_guard; assert_rc "case 10" 2; assert_out_contains "case 10" "boom for test:e2e"

# Case 11: a `--list` run prints no JSON — exit 2.
new_root; cfg_with_checkout
package_json "test:e2e=playwright test tests/"
spec tests/a.spec.ts; echo "not json" > "$CHECKOUT/.list-test:e2e.json"
run_guard; assert_rc "case 11" 2; assert_out_contains "case 11" "no parsable JSON"

# Case 12: a BOM before the heading still resolves the checkout.
new_root; cfg_with_checkout bom
package_json "test:e2e=playwright test tests/"
spec tests/a.spec.ts; listing "test:e2e" tests/a.spec.ts
run_guard; assert_rc "case 12" 0; assert_out_contains "case 12" "all reached"

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'all cases passed\n'
```

  - [x] **Step 2: Run it before the guard exists** — `scripts/test-check-spec-reach.sh`; expected:
    every case fails with `No such file or directory` for the guard, exit 1.
  - [x] **Step 3: Write the guard** as `scripts/check-spec-reach.sh`, `chmod +x`:

```bash verified:authored and run for this plan — 12 harness cases pass; live run against /Users/tweety53/Projects/gymie exits 1 naming pad-frames.spec.ts; run against this repository prints `Spec reach: not configured` and exits 0
#!/usr/bin/env bash
# check-spec-reach.sh — fail when a Playwright spec in a project's
# `regression checkout` is reached by no package.json script.
#
# Usage: check-spec-reach.sh <project-root>
#
# Why this exists: seven route A–F capture suites in gymie-playwright never
# ran once — `test:e2e` reached `tests/`, `test:e2e:full` reached
# `tests-full/`, and the root `kan-*.spec.ts` files flow.visual-verify's
# `capture` writes matched neither. One passed green against a 404 page for
# its whole life (KAN-29 self-review, KAN-390).
#
# HOW REACHABILITY IS DECIDED: by Playwright itself, never by reading the
# scripts' arguments as globs. Every package.json script whose command
# contains `playwright test` is run as `npm run -s <script> -- --list
# --reporter=json` from the checkout, and the union of `config.rootDir` +
# `suites[].file` is what those scripts reach. A static reading is wrong on
# the first real case: `playwright test kan-*.spec.ts` works only because npm
# runs the script through `sh`, which expands the glob — Playwright treats a
# positional argument as a regular expression against the absolute path, and
# the same argument quoted lists 0 files.
#
# WHAT IS ENUMERATED: every `*.spec.ts` under the checkout, `node_modules`
# and `.git` pruned. Not Playwright's wider default testMatch (`.test.ts`,
# `.spec.js`, `.spec.mjs`) — a vitest `*.test.ts` file beside a Playwright
# suite would otherwise become an orphan of a runner it never belonged to.
#
# STATED LIMIT: a script reaching Playwright indirectly (`npm run other`) is
# not followed; the files only it reaches count as orphans. Name
# `playwright test` in the script that reaches them.
#
# THREE EXIT CODES:
#   0  every enumerated spec is reached — or the project is NOT CONFIGURED
#      (no .flow/project.md, no `## visual verification` section, or no
#      `regression checkout` row), which prints `Spec reach: not configured`.
#      Both call sites (flow.verify, flow.visual-verify) run this guard as one
#      more command in a list where any non-zero exit blocks, so "nothing to
#      check" must be 0 — unlike check-visual-trigger.sh's exit 2, whose
#      caller distinguishes its exits by hand.
#   1  at least one orphan — one `<path>: reached by no package.json script`
#      line each.
#   2  cannot answer: not exactly one argument, the root is not a directory,
#      the section is declared twice (ambiguous — neither is read), the
#      checkout is not a directory, it holds no package.json, no script
#      contains `playwright test`, `node` or `npm` is absent, or a `--list`
#      run exits non-zero or prints unparsable JSON (its output is relayed).
#
# THE INPUT IS ATTACKER-INFLUENCED, exactly as check-visual-verification.sh's
# header states: .flow/project.md and the checkout's package.json are tracked
# and editable in any pull request. This guard runs `npm run` on scripts the
# project already runs through `verify` and `capture` — the same trust level,
# no widening — and every interpolated value reaching an error message passes
# through sanitize_display first.
#
# THE HEADING SCAN AND TABLE PARSER ARE check-visual-verification.sh's OWN
# CONVENTIONS, reused rather than reinvented, exactly as check-visual-trigger.sh
# and resolve-visual-screenshots.sh reuse them; split_cells/trimcell/foldcell
# come from scripts/lib/visual-table-cells.awk. No apostrophe may appear in
# the awk program below, comments included.
#
# Derives no repository root from its own location (check-guard-symlinks.sh
# rule 4) — the project root is always the argument.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/sanitize-display.sh"
source "$SCRIPT_DIR/lib/strip-bom.sh"

[ "$#" -eq 1 ] || { echo "usage: check-spec-reach.sh <project-root>" >&2; exit 2; }
ROOT="$1"
[ -d "$ROOT" ] || { echo "check-spec-reach: $ROOT is not a directory" >&2; exit 2; }
ROOT_ABS="$(cd "$ROOT" && pwd)"
CFG="$ROOT_ABS/.flow/project.md"

if [ ! -f "$CFG" ] || [ ! -r "$CFG" ]; then
  echo "Spec reach: not configured"
  exit 0
fi

VV_HEADING='^##[[:space:]]+visual verification[[:space:]]*$'
set +e
VV_COUNT="$(grep -ciE "$VV_HEADING" <(strip_bom_cat "$CFG"))"
VV_COUNT_RC=$?
set -e
if [ "$VV_COUNT_RC" -ge 2 ]; then
  echo "check-spec-reach: grep exited $VV_COUNT_RC while looking for the '## visual verification' heading in $CFG — a failure to look, not an absence" >&2
  exit 2
fi
if [ "$VV_COUNT" -gt 1 ]; then
  echo "check-spec-reach: $CFG declares $VV_COUNT '## visual verification' sections — a second declaration is ambiguous, so neither was read" >&2
  exit 2
fi
if [ "$VV_COUNT" -eq 0 ]; then
  echo "Spec reach: not configured"
  exit 0
fi

CHECKOUT="$(awk -v heading_re="$VV_HEADING" \
  -f "$SCRIPT_DIR/lib/visual-table-cells.awk" \
  -f <(cat <<'AWK_PROG'
  BEGIN { in_sec = 0; SEC_LEVEL = 2 }
  /^#+[[:space:]]/ {
    match($0, /^#+/)
    hlevel = RLENGTH
    is_own = (tolower($0) ~ heading_re)
    if (in_sec && !is_own && hlevel > SEC_LEVEL) { next }
    in_sec = is_own
    next
  }
  !in_sec { next }
  {
    n = split_cells($0, cells)
    if (n != 2) next
    if (foldcell(cells[1]) == "regression checkout" && !found) {
      found = 1
      print trimcell(cells[2])
    }
  }
AWK_PROG
  ) <(strip_bom_cat "$CFG"))"

if [ -z "$CHECKOUT" ]; then
  echo "Spec reach: not configured"
  exit 0
fi
if [ ! -d "$CHECKOUT" ]; then
  printf 'check-spec-reach: `regression checkout` names `%s`, which is not an existing directory\n' "$CHECKOUT" | sanitize_display >&2
  exit 2
fi
CHECKOUT="$(cd "$CHECKOUT" && pwd)"
if [ ! -f "$CHECKOUT/package.json" ]; then
  printf 'check-spec-reach: `%s` holds no package.json — nothing declares what reaches its specs\n' "$CHECKOUT" | sanitize_display >&2
  exit 2
fi
for tool in node npm; do
  command -v "$tool" >/dev/null 2>&1 || { echo "check-spec-reach: $tool is not on PATH — cannot ask Playwright what the scripts reach" >&2; exit 2; }
done

# Every script whose command contains `playwright test`, one name per line.
set +e
SCRIPTS="$(node -e '
  const s = JSON.parse(require("fs").readFileSync(process.argv[1], "utf8")).scripts || {};
  for (const [k, v] of Object.entries(s)) if (String(v).includes("playwright test")) console.log(k);
' "$CHECKOUT/package.json" 2>&1)"
NODE_RC=$?
set -e
if [ "$NODE_RC" -ne 0 ]; then
  printf 'check-spec-reach: could not read scripts from `%s/package.json`: %s\n' "$CHECKOUT" "$SCRIPTS" | sanitize_display >&2
  exit 2
fi
if [ -z "$SCRIPTS" ]; then
  printf 'check-spec-reach: no script in `%s/package.json` contains `playwright test` — nothing can reach its specs\n' "$CHECKOUT" | sanitize_display >&2
  exit 2
fi

REACHED="$(mktemp "${TMPDIR:-/tmp}/check-spec-reach.XXXXXX")"
trap 'rm -f "$REACHED"' EXIT

while IFS= read -r script; do
  [ -n "$script" ] || continue
  set +e
  LISTING="$(cd "$CHECKOUT" && env -u PLAYWRIGHT_JSON_OUTPUT_NAME -u PLAYWRIGHT_JSON_OUTPUT_DIR -u PLAYWRIGHT_JSON_OUTPUT_FILE \
    npm run -s "$script" -- --list --reporter=json 2>&1)"
  LIST_RC=$?
  set -e
  if [ "$LIST_RC" -ne 0 ]; then
    printf 'check-spec-reach: `npm run -s %s -- --list --reporter=json` exited %s in `%s`:\n%s\n' "$script" "$LIST_RC" "$CHECKOUT" "$LISTING" | sanitize_display >&2
    exit 2
  fi
  set +e
  FILES="$(printf '%s' "$LISTING" | node -e '
    let s = ""; process.stdin.on("data", d => s += d).on("end", () => {
      const j = JSON.parse(s.slice(s.indexOf("{")));
      const root = j.config.rootDir;
      const walk = (x) => { if (x.file) console.log(require("path").resolve(root, x.file)); (x.suites || []).forEach(walk); };
      (j.suites || []).forEach(walk);
    });
  ' 2>&1)"
  PARSE_RC=$?
  set -e
  if [ "$PARSE_RC" -ne 0 ]; then
    printf 'check-spec-reach: `npm run -s %s -- --list --reporter=json` printed no parsable JSON in `%s`: %s\n' "$script" "$CHECKOUT" "$FILES" | sanitize_display >&2
    exit 2
  fi
  printf '%s\n' "$FILES" >> "$REACHED"
done <<< "$SCRIPTS"

ORPHANS=0
TOTAL=0
while IFS= read -r -d '' spec; do
  TOTAL=$((TOTAL + 1))
  if ! grep -qxF -- "$spec" "$REACHED"; then
    printf '%s: reached by no package.json script\n' "${spec#"$CHECKOUT"/}" | sanitize_display
    ORPHANS=$((ORPHANS + 1))
  fi
done < <(find "$CHECKOUT" \( -name node_modules -o -name .git \) -prune -o -name '*.spec.ts' -type f -print0 | sort -z)

if [ "$ORPHANS" -ne 0 ]; then
  echo "Spec reach: $ORPHANS of $TOTAL spec(s) reached by no package.json script" >&2
  exit 1
fi
echo "Spec reach: $TOTAL spec(s), all reached"
```

  - [x] **Step 4: Run the harness** — `scripts/test-check-spec-reach.sh`; expected: `all cases
    passed`, exit 0, 26 `ok:` lines.
  - [x] **Step 5: Smoke the live case** — `scripts/check-spec-reach.sh /Users/tweety53/Projects/gymie`;
    expected: exit 1, `pad-frames.spec.ts: reached by no package.json script`, then
    `Spec reach: 1 of 29 spec(s) reached by no package.json script` on stderr. Then
    `scripts/check-spec-reach.sh .`; expected: `Spec reach: not configured`, exit 0.
  - [x] **Step 6: Run the whole suite** — `scripts/run-guard-tests.sh`; expected: 43 harnesses,
    all pass (the runner discovers the new file by glob).
    <!-- predicted: scripts/run-guard-tests.sh after task 1 -->
  - [x] **Step 7: Commit** with the subject above.

- [x] 2. Ship the guard

Symlink the guard into `skills/flow/scripts/`, add it to the guard-presence union in
`skills/flow/SKILL.md`, exercise it from this repository's own `## lint`, and note in
`skills/flow-contracts/project-configuration.md` that `regression checkout` is also the root the
guard scans.

**Files:** `skills/flow/scripts/check-spec-reach.sh`, `skills/flow/SKILL.md`, `.flow/project.md`,
`skills/flow-contracts/project-configuration.md`
**Tests:** **none** — the symlink and list edits are covered by the existing
`scripts/check-guard-symlinks.sh` and `scripts/check-references.sh` runs in step 5.
**Regression:** reverting this commit leaves `check-spec-reach.sh` unshipped — task 3's stage text
cites a basename `check-guard-symlinks.sh` rule 2 then reports as having no symlink.
**Baseline:** before=42 after=43 harnesses discovered by `scripts/run-guard-tests.sh`, unchanged by
this task — task 1 already added the one this task ships
<!-- measured: ls scripts/test-*.sh | wc -l @ d31ebd9, plus task 1 -->
**Commit:** `feat(flow): ship check-spec-reach.sh`
**Build:** green

  - [x] **Step 1: Symlink** — `ln -s ../../../scripts/check-spec-reach.sh skills/flow/scripts/check-spec-reach.sh`;
    relative target, exactly as `skills/flow/scripts/check-visual-trigger.sh` is written.
  - [x] **Step 2: Guard-presence list** — in `skills/flow/SKILL.md`'s **Check guard presence**
    paragraph, insert `` `check-spec-reach.sh`, `` into the alphabetical run between
    `` `check-plan-shape.sh`, `` and `` `check-task-commit-fields.sh`, ``.
  - [x] **Step 3: `## lint`** — in `.flow/project.md`, after the line
    `printf 'stats/web/src/App.tsx\n' | scripts/check-visual-trigger.sh .`, add
    `scripts/check-spec-reach.sh .` — this repository declares no `regression checkout`, so the
    line exercises the not-configured path and exits 0.
  - [x] **Step 4: Contract note** — in `skills/flow-contracts/project-configuration.md`'s
    `## visual verification` settings table, extend the `regression checkout` row's Meaning cell
    with one sentence: `Also the root `<agents repo>/scripts/check-spec-reach.sh` enumerates
    `*.spec.ts` under and diffs against what the checkout's `package.json` scripts list — its
    header is canonical for that gate and its exit codes.`
  - [x] **Step 5: Verify** — run `scripts/check-guard-symlinks.sh`, `scripts/check-references.sh`,
    `scripts/check-vocabulary.sh`, `scripts/check-contract-budget.sh` and
    `scripts/check-spec-reach.sh .`; expected: every one exits 0, `check-guard-symlinks.sh`'s
    verdict counts one more guard for `flow`, and no budget row fails (`SKILL.md` and
    `project-configuration.md` both have over 1.5 KB of headroom).
    <!-- predicted: scripts/check-contract-budget.sh after task 2; headroom from wc -c against the budgets() rows @ d31ebd9 -->
  - [x] **Step 6: Commit** with the subject above.

- [x] 3. Gate `flow.verify` and `flow.visual-verify`

Two edits in `skills/flow/verify-and-handoff.md`, and the budget row they outgrow.

**Files:** `skills/flow/verify-and-handoff.md`, `scripts/check-contract-budget.sh`
**Tests:** **none** — prose; `scripts/check-guard-symlinks.sh` rule 2 checks the cited basename
resolves and `scripts/check-contract-budget.sh` checks the row.
**Regression:** reverting this commit leaves the guard shipped but never run — the KAN-29 failure
recurs on the next `capture`.
**Baseline:** before=43 after=43 harnesses; `check-contract-budget.sh` before=71 after=71 owned
files within budget
<!-- predicted: scripts/run-guard-tests.sh and scripts/check-contract-budget.sh after task 3 -->
**Commit:** `feat(flow): gate flow.verify and flow.visual-verify on spec reach`
**Build:** green

  - [x] **Step 1: `flow.verify`** — in the **Verify** paragraph that begins `Resolve the commands
    `project-get.sh <worktree> lint``, replace `the lint commands, then the test commands, in the
    order printed;` with `the lint commands, then the test commands, in the order printed, then
    `check-spec-reach.sh <worktree>` — one more command in the same list, whose exit 0 line
    `Spec reach: not configured` is the ordinary case for a project with no `regression checkout`
    (its header is canonical for its exit codes);`. The existing sentence `a non-zero exit blocks
    this handoff` already covers exit 1 and 2.
  - [x] **Step 2: `flow.visual-verify` step 6** — append to step 6, after `(see **Blocking**
    below).`: `Then run `check-spec-reach.sh <worktree>` — the spec `capture` just wrote must be
    reached by a `package.json` script of the `regression checkout`; exit 1 (an orphan, named) or 2
    (cannot answer) blocks.`
  - [x] **Step 3: Report and Blocking** — in the visual `## Report` block, add
    `- spec reach: exit <n>` directly after the `- capture: exit <n>` entry and its output line;
    in **Blocking**, after `a stack that could not be started,` insert `a `check-spec-reach.sh`
    exit 1 or 2,`.
  - [x] **Step 4: Budget** — run `scripts/check-contract-budget.sh`; expected: it fails on
    `skills/flow/verify-and-handoff.md`. Raise that file's row in `budgets()` to its new
    `wc -c` size plus 25%, rounded to an integer, and re-run; expected: `BUDGET-OK`.
    <!-- predicted: scripts/check-contract-budget.sh after step 3 — the file has 49 bytes of headroom -->
  - [x] **Step 5: Verify** — `scripts/check-guard-symlinks.sh`, `scripts/check-references.sh`,
    `scripts/check-vocabulary.sh`, `scripts/check-normative-inventory.sh`; expected: all exit 0.
  - [x] **Step 6: Commit** with the subject above.
