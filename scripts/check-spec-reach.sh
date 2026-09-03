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
