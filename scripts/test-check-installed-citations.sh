#!/usr/bin/env bash
# test-check-installed-citations.sh — assertion harness for
# check-installed-citations.sh/.py. Builds throwaway fixture repositories
# under a sandboxed TMPDIR — each carrying its own runnable setup.sh, so a
# case can change what "installed" means without touching this repository —
# and asserts the guard's classifier, refusal and report-shape behaviour
# against them. Never modifies the real repository tree.
#
# Modelled on scripts/test-check-guard-symlinks.sh: separate stdout/stderr
# capture (run_guard), a fixture builder (new_fixture_repo), one ok:/FAIL:
# line per assertion, and a final tail line + non-zero exit on any failure.
#
# Bash 3.2 is the floor, as scripts/test-check-finish-preflight.sh's header
# records: indexed arrays only, no associative arrays.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-installed-citations.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

# An indexed array, not a space-separated string — a sandbox path from
# mktemp under TMPDIR may contain spaces.
SANDBOXES=()
cleanup() {
  [ "${#SANDBOXES[@]}" -eq 0 ] && return 0
  local s
  for s in "${SANDBOXES[@]}"; do rm -rf "$s"; done
}
trap cleanup EXIT

# run_guard <root> -> sets OUT (stdout only), ERR, RC. The two streams are
# captured separately so a refusal's "message on stderr, nothing on stdout"
# contract can actually be checked — a merged capture cannot tell an empty
# stdout from one carrying the message.
run_guard() {
  local errfile
  errfile="$(mktemp "${TMPDIR:-/tmp}/check-installed-citations-stderr.XXXXXX")"
  SANDBOXES+=("$errfile")
  set +e
  OUT="$(CHECK_INSTALLED_CITATIONS_ROOT="$1" "$GUARD" 2>"$errfile")"
  RC=$?
  set -e
  ERR="$(cat "$errfile")"
}

# write_fixture_setup_sh — a minimal, real, runnable setup.sh for the
# fixture: `global` symlinks every skills/<name>/ directory and every
# always-on rules/*.mdc into $HOME/.claude/{skills,rules}; `all <proj>`
# copies CLAUDE.md into <proj> when present. This is deliberately not the
# real installer's shape — it exists so a case can change what gets
# installed (case 17 below) without touching scripts/setup.sh.
write_fixture_setup_sh() {
  cat > "$FIXTURE/setup.sh" <<'SETUPEOF'
#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MODE="${1:-}"
PROJ="${2:-$SCRIPT_DIR}"
: "${HOME:?HOME must be set}"
case "$MODE" in
  global)
    mkdir -p "$HOME/.claude/skills" "$HOME/.claude/rules"
    for d in "$SCRIPT_DIR"/skills/*/; do
      [ -d "$d" ] || continue
      ln -sfn "${d%/}" "$HOME/.claude/skills/$(basename "$d")"
    done
    for f in "$SCRIPT_DIR"/rules/*.mdc; do
      [ -f "$f" ] || continue
      if grep -q 'alwaysApply: true' "$f"; then
        ln -sfn "$f" "$HOME/.claude/rules/$(basename "$f")"
      fi
    done
    if [ -d "$SCRIPT_DIR/docs" ]; then
      ln -sfn "$SCRIPT_DIR/docs" "$HOME/.claude/docs"
    fi
    ;;
  all)
    mkdir -p "$PROJ"
    if [ -f "$SCRIPT_DIR/CLAUDE.md" ] && [ ! -f "$PROJ/CLAUDE.md" ]; then
      cp "$SCRIPT_DIR/CLAUDE.md" "$PROJ/CLAUDE.md"
    fi
    ;;
  *)
    echo "fixture setup.sh: unknown mode '$MODE'" >&2
    exit 1
    ;;
esac
SETUPEOF
  chmod +x "$FIXTURE/setup.sh"
}

# SAFE_CITATION — a citation that always classifies AND always judges
# clean (an installed-root path). Task 2 step 5's coverage discipline
# (scripts/lib/coverage.sh's coverage_declare/coverage_verdict, wired
# through this guard's wrapper — see the per-task review's finding 3)
# means a corpus member with ZERO checked citations is now itself a
# violation unless declared, and the wrapper's declared-zero list is
# scoped to this repository's own real paths, never a fixture's. Every
# fixture file below carries one of these by default so a case testing
# ONE specific citation shape does not also trip an incidental,
# unrelated "0 checked, not declared" finding on some OTHER fixture file
# nobody is testing.
SAFE_CITATION='`skills/other/SKILL.md`'

# new_fixture_repo -> sets FIXTURE to a fresh miniature agents repository:
# a real setup.sh (above), one always-installed skill (skills/myflow-do),
# one always-on rule, one opt-in rule (never installed, never scanned —
# case 16), a CLAUDE.md (a copy target) and a README.md (a real
# repository-root file the bare-filename rule can resolve against). Every
# installed file carries SAFE_CITATION by default — see its own comment.
new_fixture_repo() {
  FIXTURE="$(mktemp -d "${TMPDIR:-/tmp}/check-installed-citations-fixture.XXXXXX")"
  SANDBOXES+=("$FIXTURE")
  mkdir -p "$FIXTURE/skills/myflow-do" "$FIXTURE/skills/other" "$FIXTURE/rules"
  write_fixture_setup_sh
  printf '# myflow-do fixture\n%s\n' "$SAFE_CITATION" > "$FIXTURE/skills/myflow-do/SKILL.md"
  printf '# other fixture\n`skills/myflow-do/SKILL.md`\n' > "$FIXTURE/skills/other/SKILL.md"
  cat > "$FIXTURE/rules/always-on-rule.mdc" <<EOF
---
alwaysApply: true
---
# always-on rule fixture

$SAFE_CITATION
EOF
  cat > "$FIXTURE/rules/opt-in-rule.mdc" <<'EOF'
---
alwaysApply: false
---
# opt-in rule fixture

placeholder
EOF
  printf '# CLAUDE.md fixture\n%s\n' "$SAFE_CITATION" > "$FIXTURE/CLAUDE.md"
  printf '# README fixture\n' > "$FIXTURE/README.md"
}

# write_file <relpath> <content> — (over)writes $FIXTURE/<relpath>, creating
# parent directories as needed.
write_file() {
  mkdir -p "$(dirname "$FIXTURE/$1")"
  printf '%s' "$2" > "$FIXTURE/$1"
}

# assert_case <label> <relpath> <content> <reported|not-reported>
#
# Writes <content> to <relpath> inside a fresh fixture, plus SAFE_CITATION
# appended (see its own comment — keeps this one file's own coverage
# non-zero regardless of what <content> tests, since some cases test a
# token that is excluded outright and never reaches "checked" at all), and
# runs the guard, checking whether a violation naming <relpath> appears in
# its output. The fixture is rebuilt every time (new_fixture_repo) so
# cases never share state.
assert_case() {
  local label="$1" relpath="$2" content="$3" expect="$4" body
  new_fixture_repo
  body="$(printf '%s\n%s\n' "$content" "$SAFE_CITATION")"
  write_file "$relpath" "$body"
  run_guard "$FIXTURE"
  case "$expect" in
    reported)
      if [ "$RC" -eq 1 ] && printf '%s\n' "$OUT" | grep -aq -- "^$relpath:"; then
        pass "$label"
      else
        fail "$label: expected a violation naming $relpath, got rc=$RC out=$OUT err=$ERR"
      fi
      ;;
    not-reported)
      if [ "$RC" -eq 0 ] && ! printf '%s\n' "$OUT" | grep -aq -- "^$relpath:"; then
        pass "$label"
      else
        fail "$label: expected no violation for $relpath, got rc=$RC out=$OUT err=$ERR"
      fi
      ;;
    *)
      fail "$label: bad expectation '$expect'"
      ;;
  esac
}

# ===========================================================================
# SECTION: The citation classifier (17 cases, task 1 step 2's table)
# ===========================================================================

assert_case "bare-root-file" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture
`README.md`
' "reported"

assert_case "bare-generic-filename" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture
`tasks.md`
' "not-reported"

assert_case "installed-root" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture
`skills/other/SKILL.md`
' "not-reported"

assert_case "agents-repo-prefix" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture
`<agents repo>/README.md`
' "not-reported"

# agents-repo-prefix-bogus-path — a companion to the case above, per the
# per-task review finding: "not reported" alone cannot tell "recognised
# and correctly passed" apart from "silently dropped before judgment ever
# ran" — both look identical from the outside. `<agents repo>/…` is
# itself two whitespace-separated words (the placeholder's own embedded
# space), so a tokenizer that collapses a multi-word backtick span to its
# first word — the fix for the multi-word-shell-example false positive a
# few cases up — would otherwise discard everything from `<agents` on,
# never reaching classify_token/judges_ok at all. The path here does not
# exist (`does-not-exist/nested/path.md`): existence is never the guard's
# business for a slash-containing token (see specs/myflow-citation-roots/
# spec.md), so a bogus target still proves nothing except whether the
# CITATION ITSELF was seen. It is proven seen by its member's coverage
# count going to 1, not merely by the absence of a violation line.
new_fixture_repo
write_file "skills/myflow-do/SKILL.md" '# myflow-do fixture
`<agents repo>/does-not-exist/nested/path.md`
'
run_guard "$FIXTURE"
if [ "$RC" -eq 0 ] \
  && ! printf '%s\n' "$OUT" | grep -aq -- '^skills/myflow-do/SKILL.md:' \
  && printf '%s\n' "$OUT" | grep -aq -- 'skills/myflow-do/SKILL.md 1'; then
  pass "agents-repo-prefix-bogus-path: recognised (coverage count 1), not silently dropped"
else
  fail "agents-repo-prefix-bogus-path: rc=$RC out='$OUT'"
fi

assert_case "project-prefix" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture
`<project>/.myflow/project.md`
' "not-reported"

assert_case "unrooted-directory" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture
`.myflow/project.md`
' "reported"

assert_case "unrooted-script" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture
`scripts/check-vocabulary.sh`
' "reported"

assert_case "fenced-command" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture

```bash
git -C <abs-worktree> reset -- openspec/
```
' "not-reported"

assert_case "fenced-comment" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture

```bash
# see openspec/
```
' "reported"

# fenced-comment-agents-repo-prefix — per the per-task review finding: a
# shell-fence comment line is tokenized by a plain whitespace split
# (raw.split()), same as any other shell-word line, never by backtick
# extraction — there ARE no backticks inside a fence. Without the same
# `<agents repo>/…` merge the backtick path needed, this line's un-quoted
# `<agents repo>/README.md` splits into `<agents` and `repo>/README.md`,
# and the SECOND fragment alone gets judged and reported as a bogus,
# standalone unrooted citation `` `repo>/README.md` `` — a false positive
# with no corpus occurrence today, but one waves 4-6 are about to create
# the conditions for.
assert_case "fenced-comment-agents-repo-prefix" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture

```bash
# see <agents repo>/README.md for details
```
' "not-reported"

assert_case "absolute-path" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture
`/etc/hosts`
' "not-reported"

assert_case "home-path" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture
`~/.claude/skills/`
' "not-reported"

assert_case "url" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture
`https://example.test/a/b`
' "not-reported"

assert_case "shell-variable" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture
`$SCRIPT_DIR/lib/x.sh`
' "not-reported"

assert_case "git-ref" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture
`origin/main`
' "not-reported"

assert_case "regex-fragment" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture
`[A-Za-z0-9._-]+/x`
' "not-reported"

# opt-in-rule-out-of-scope — the citation sits in rules/opt-in-rule.mdc,
# which the fixture's own setup.sh only ever installs when alwaysApply is
# true (it is false here) — so this file is never part of the installed
# corpus and the guard must never even open it.
assert_case "opt-in-rule-out-of-scope" "rules/opt-in-rule.mdc" \
  '---
alwaysApply: false
---
# opt-in rule fixture

`README.md`
' "not-reported"

# newly-installed-directory — a directory (docs/) the fixture's setup.sh
# was NOT installing until this case adds it (write_fixture_setup_sh above
# already contains the `docs` install line unconditionally, gated on the
# directory's own presence — so simply creating $FIXTURE/docs is "editing
# what setup.sh installs" without editing setup.sh's own text again, and
# certainly without editing the guard). This is what proves the guard
# derives roots by RUNNING the installer rather than re-implementing it.
new_fixture_repo
mkdir -p "$FIXTURE/docs"
write_file "docs/somefile.md" '# docs fixture

`README.md`
'
run_guard "$FIXTURE"
if [ "$RC" -eq 1 ] && printf '%s\n' "$OUT" | grep -aq -- '^docs/somefile.md:'; then
  pass "newly-installed-directory"
else
  fail "newly-installed-directory: expected a violation naming docs/somefile.md, got rc=$RC out=$OUT err=$ERR"
fi

# ===========================================================================
# SECTION: Task 9's three classifier exclusions — the real corpus proved
# necessary after wave B refused to force a root onto four tokens the
# classifier called citations that were not. Four cases: one per
# exclusion, plus placeholder-rooted-is-recognised, which draws the same
# recognised-versus-never-seen distinction agents-repo-prefix-bogus-path
# already draws (that exact collapse is what hid this change's critical
# defect behind a green test the first time).
# ===========================================================================

assert_case "parent-relative-path" "skills/myflow-do/SKILL.md"   '# myflow-do fixture
`../<other-app>`
' "not-reported"

assert_case "placeholder-rooted" "skills/myflow-do/SKILL.md"   '# myflow-do fixture
`<state-dir>/<name>-proposal-artifact.html`
' "not-reported"

assert_case "quoted-program-output" "skills/myflow-do/SKILL.md"   "# myflow-do fixture
\`'openspec/<name>':\`
" "not-reported"

# placeholder-rooted-is-recognised — a companion to placeholder-rooted,
# same shape as agents-repo-prefix-bogus-path: "not reported" alone
# cannot tell "recognised and correctly passed" apart from "silently
# dropped before judgment ever ran" — both look identical from outside.
# The path here does not exist (`<state-dir>/does-not-exist.html`):
# existence is never checked for a slash-containing token, so a bogus
# target proves nothing except whether the CITATION ITSELF was seen —
# proven by its member's coverage count going non-zero, not merely by
# the absence of a violation line.
new_fixture_repo
write_file "skills/myflow-do/SKILL.md" '# myflow-do fixture
`<state-dir>/does-not-exist.html`
'
run_guard "$FIXTURE"
if [ "$RC" -eq 0 ]   && ! printf '%s\n' "$OUT" | grep -aq -- '^skills/myflow-do/SKILL.md:'   && printf '%s\n' "$OUT" | grep -aq -- 'skills/myflow-do/SKILL.md 1'; then
  pass "placeholder-rooted-is-recognised: recognised (coverage count 1), not silently dropped"
else
  fail "placeholder-rooted-is-recognised: rc=$RC out='$OUT'"
fi

# unrecognised-placeholder-is-reported — per the per-task review's
# critical finding: the first cut of the placeholder rule accepted ANY
# bracket-shaped first segment, which fails open — `<foo>/…` passed while
# naming an unrooted path, and so would a typo of a real placeholder
# (`<changeroot>/`, `<change-root>/`). This is the case whose ABSENCE let
# that through: every earlier placeholder case only ever exercised a
# MEMBER of the closed set, never a non-member, so a bracket-shape-only
# implementation passed all of them anyway.
assert_case "unrecognised-placeholder-is-reported" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture
`<foo>/spectre/specs/x.md`
' "reported"

# ===========================================================================
# SECTION: Task 10's two classifier exclusions — wave C's review found
# sites that satisfied a root judgment while remaining wrong: a git branch
# name read as an instruction to create a branch of that literal name, and
# a fabricated findings-table Location example taught a citation shape no
# real finding uses. Reaching zero by re-rooting either would have been
# syntactically clean and semantically false, so both are reverted to bare
# behind a new exclusion instead. Three cases: one per exclusion, plus
# abs-worktree-rooted-is-recognised, which draws the same recognised-
# versus-never-seen distinction agents-repo-prefix-bogus-path and
# placeholder-rooted-is-recognised already draw.
# ===========================================================================

assert_case "git-branch-name-is-not-a-citation" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture
Branch `spectre/<name>`.
' "not-reported"

assert_case "file-line-reference-is-not-a-citation" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture
| F1 | Bugbot | Minor | `src/Foo.kt:42` | replaced the silent catch |
' "not-reported"

# abs-worktree-rooted-is-recognised — the companion case task 10 itself
# calls for: `<abs-worktree>/…` must be RECOGNISED as a placeholder root,
# not merely absent from the violation list — a citation excluded before
# judgment ever ran would look identical from the outside. The path here
# does not exist (`<abs-worktree>/.superpowers/sdd/does-not-exist.diff`):
# existence is never the guard's business for a slash-containing token, so
# a bogus target proves nothing except whether the CITATION ITSELF was
# seen — proven by its member's coverage count going non-zero, not merely
# by the absence of a violation line.
new_fixture_repo
write_file "skills/myflow-do/SKILL.md" '# myflow-do fixture
`<abs-worktree>/.superpowers/sdd/does-not-exist.diff`
'
run_guard "$FIXTURE"
if [ "$RC" -eq 0 ] \
  && ! printf '%s\n' "$OUT" | grep -aq -- '^skills/myflow-do/SKILL.md:' \
  && printf '%s\n' "$OUT" | grep -aq -- 'skills/myflow-do/SKILL.md 1'; then
  pass "abs-worktree-rooted-is-recognised: recognised (coverage count 1), not silently dropped"
else
  fail "abs-worktree-rooted-is-recognised: rc=$RC out='$OUT'"
fi

# ===========================================================================
# SECTION: Second per-task review — three bugs the panel found in what
# task 10 shipped: the `origin` exclusion excluded ANY token whose first
# segment was `origin`, not only a ref shape; extract_backtick_tokens kept
# only a span's leading word, so a REAL second citation in the same span
# went not merely unreported but never seen; FILE_LINE_RE was confirmed,
# not merely left alone — narrowing it was tried and found to force
# re-rooting SKILL.md:485's own fabricated findings-table example.
# ===========================================================================

assert_case "origin-with-extension-is-a-citation" "skills/myflow-do/SKILL.md"   '# myflow-do fixture
`origin/README.md`
' "reported"

assert_case "origin-ref-with-nested-path-is-not-a-citation" "skills/myflow-do/SKILL.md"   '# myflow-do fixture
`origin/spectre/<name>`
' "not-reported"

# origin-directory-is-a-citation — panel round 3's finding: the ORIGINAL
# bound ("remainder carries no file extension") also silently excluded a
# DIRECTORY, since a directory has no extension either but is never a
# git ref. `origin/rules/` (trailing slash) must be classified and
# reported, not vanish.
assert_case "origin-directory-is-a-citation" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture
`origin/rules/`
' "reported"

assert_case "second-word-in-backtick-span-is-seen" "skills/myflow-do/SKILL.md"   '# myflow-do fixture
`see .myflow/project.md`
' "reported"

assert_case "shell-example-second-word-not-path-shaped" "skills/myflow-do/SKILL.md"   '# myflow-do fixture
`skills/other/SKILL.md verbose`
' "not-reported"

# file-line-reference-is-not-a-citation (task 10, above) already pins
# FILE_LINE_RE's kept, undivided behaviour — re-confirmed rather than
# re-tested here; see this guard's own module docstring for the narrowing
# attempt and why it was rejected.

# ===========================================================================
# SECTION: Third per-task review — a sixth placeholder root (<skill-dir>,
# collapsing six real corpus sites' three different wordings of the same
# concept) and Group B split into three separately-bounded exclusions
# rather than one that would have swallowed all three.
# ===========================================================================

assert_case "skill-dir-rooted" "skills/myflow-do/SKILL.md"   '# myflow-do fixture
`<skill-dir>/scripts/check-vocabulary.sh`
' "not-reported"

# skill-dir-rooted-is-recognised — the same recognised-versus-never-seen
# distinction every root in this guard is held to. The path here does not
# exist (`<skill-dir>/scripts/does-not-exist.sh`): existence is never the
# guard's business for a slash-containing token, so a bogus target proves
# nothing except whether the CITATION ITSELF was seen — proven by its
# member's coverage count going non-zero, not merely by the absence of a
# violation line.
new_fixture_repo
write_file "skills/myflow-do/SKILL.md" '# myflow-do fixture
`<skill-dir>/scripts/does-not-exist.sh`
'
run_guard "$FIXTURE"
if [ "$RC" -eq 0 ] \
  && ! printf '%s\n' "$OUT" | grep -aq -- '^skills/myflow-do/SKILL.md:' \
  && printf '%s\n' "$OUT" | grep -aq -- 'skills/myflow-do/SKILL.md 1'; then
  pass "skill-dir-rooted-is-recognised: recognised (coverage count 1), not silently dropped"
else
  fail "skill-dir-rooted-is-recognised: rc=$RC out='$OUT'"
fi

assert_case "spectre-branch-with-change-name-is-not-a-citation" "skills/myflow-do/SKILL.md"   '# myflow-do fixture
`spectre/<change-name>`
' "not-reported"

# spectre-shape-with-real-path-is-still-reported and
# spectre-shape-with-trailing-path-is-still-reported — the "second half"
# GIT_BRANCH_SPECTRE_RE's own comment calls out explicitly: the shape is
# anchored at BOTH ends so it cannot widen past exactly one bracket
# segment. Without these two cases, a shape rule that accidentally matched
# ANY `spectre/`-prefixed token would still pass every other case in this
# file (none of them exercises a real `spectre/...` path), which is
# exactly how task 9's own placeholder generalisation failed open the
# first time.
assert_case "spectre-shape-with-real-path-is-still-reported" "skills/myflow-do/SKILL.md"   '# myflow-do fixture
`spectre/specs/x.md`
' "reported"

assert_case "spectre-shape-with-trailing-path-is-still-reported" "skills/myflow-do/SKILL.md"   '# myflow-do fixture
`spectre/changes/<name>/`
' "reported"

# ===========================================================================
# SECTION: Task 9 (kan-239) — a second pipeline-created branch shape,
# `chore/archive-<name>`, needs the same GIT_BRANCH_SPECTRE_RE treatment:
# a branch name is not a path citation. Two cases, same bound as the
# spectre pair above: the bare shape is not reported, and a bracket
# segment followed by more path — `chore/archive-<name>/spec.md` — is the
# fail-open bound and must stay reported.
# ===========================================================================

assert_case "chore-archive-branch-name-is-not-a-citation" "skills/myflow-do/SKILL.md"   '# myflow-do fixture
Branch `chore/archive-<name>`.
' "not-reported"

assert_case "chore-archive-shape-with-trailing-path-is-still-reported" "skills/myflow-do/SKILL.md"   '# myflow-do fixture
`chore/archive-<name>/spec.md`
' "reported"

# This case pins the BRACKET-CONTENT bound, which the trailing-path case
# above does not: widening the class to `[^>]+` still rejects
# `chore/archive-<name>/spec.md` (that string carries only one `>`, so the
# wider class has nothing to smuggle past), and every other assertion in
# this file still passes. Only a bracket segment carrying a second `/`
# separates `[^/<>]+` from `[^>]+`, so only this case fails when someone
# later widens the regex "harmlessly" -- the exact fail-open class this
# file's history exists to prevent.
assert_case "chore-archive-bracket-cannot-smuggle-a-slash" "skills/myflow-do/SKILL.md"   '# myflow-do fixture
`chore/archive-<a/b>`
' "reported"

assert_case "html-comment-is-not-a-citation" "skills/myflow-do/SKILL.md"   '# myflow-do fixture
`<!-- measured: ./gradlew test @ c515c42 -->`
' "not-reported"

# html-comment-with-leading-whitespace-is-not-a-citation — panel round 5
# (the panel's code-quality reviewer): a space right inside the opening backtick
# (`` ` <!-- ... -->` ``) must not defeat the span-initial check. Dormant
# in the real corpus today (every live site is span-initial) but
# untested before this case, and the reviewer showed the fragile
# `startswith` alone lets `./gradlew` fall through to word-splitting and
# report.
assert_case "html-comment-with-leading-whitespace-is-not-a-citation" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture
` <!-- measured: ./gradlew test @ c515c42 -->`
' "not-reported"

assert_case "colon-segment-is-not-a-citation" "skills/myflow-do/SKILL.md"   '# myflow-do fixture
`measured:/predicted:`
' "not-reported"


# final-colon-segment-is-a-citation — panel round 3's finding: the
# ORIGINAL rule tested EVERY segment including the last, which silently
# dropped a real citation carrying a trailing colon inside its own span
# (e.g. "see `.myflow/project.md:` for the list"). Only a colon on a
# NON-FINAL segment signals "not a path" now.
assert_case "final-colon-segment-is-a-citation" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture
`.myflow/project.md:`
' "reported"

# redirection-is-not-merged-into-a-citation — panel round 2's finding: the
# unbounded merge joined a shell redirection's `<`, its target, and `>`
# into one garbled false citation, over ordinary prose documenting a
# command inline. The intervening word (`skills/other/SKILL.md`) is
# deliberately a REAL, passing citation on its own — so the OLD (buggy)
# merged form `< skills/other/SKILL.md >` reports a violation (first
# segment `< skills` names no root), while the FIXED, rejected-merge form
# reports nothing at all: `<` and `>` are each harmless standalone words,
# and `skills/other/SKILL.md` passes cleanly by itself. That gap — clean
# under the fix, a violation without it — is what the mutation below
# proves.
assert_case "redirection-is-not-merged-into-a-citation" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture
Run `some-cmd < skills/other/SKILL.md > output.log` to reproduce.
' "not-reported"

# ===========================================================================
# SECTION: Refusal cases (4 cases, task 1 step 3's table). Every refusal
# case asserts stdout is empty as well as the exit code.
# ===========================================================================

# CHECK_INSTALLED_CITATIONS_ROOT set but empty.
EMPTY_ROOT_ERRFILE="$(mktemp "${TMPDIR:-/tmp}/check-installed-citations-empty-root.XXXXXX")"
SANDBOXES+=("$EMPTY_ROOT_ERRFILE")
set +e
OUT="$(CHECK_INSTALLED_CITATIONS_ROOT= "$GUARD" 2>"$EMPTY_ROOT_ERRFILE")"
RC=$?
set -e
ERR="$(cat "$EMPTY_ROOT_ERRFILE")"
if [ "$RC" -eq 2 ] && [ -z "$OUT" ] && printf '%s' "$ERR" | grep -aq 'CHECK_INSTALLED_CITATIONS_ROOT'; then
  pass "CHECK_INSTALLED_CITATIONS_ROOT set but empty: exit 2, stderr names it, stdout empty"
else
  fail "CHECK_INSTALLED_CITATIONS_ROOT set but empty: rc=$RC out='$OUT' err='$ERR'"
fi

# Root is not a directory.
run_guard "${TMPDIR:-/tmp}/check-installed-citations-does-not-exist-$$"
if [ "$RC" -eq 2 ] && [ -z "$OUT" ]; then
  pass "root is not a directory: exit 2, stdout empty"
else
  fail "root is not a directory: rc=$RC out='$OUT' err='$ERR'"
fi

# The fixture's setup.sh exits non-zero.
FIXTURE="$(mktemp -d "${TMPDIR:-/tmp}/check-installed-citations-fixture.XXXXXX")"
SANDBOXES+=("$FIXTURE")
cat > "$FIXTURE/setup.sh" <<'EOF'
#!/usr/bin/env bash
echo "fixture setup.sh: deliberately failing" >&2
exit 1
EOF
chmod +x "$FIXTURE/setup.sh"
mkdir -p "$FIXTURE/skills" "$FIXTURE/rules"
run_guard "$FIXTURE"
if [ "$RC" -eq 2 ] && [ -z "$OUT" ] && printf '%s' "$ERR" | grep -aq 'setup.sh'; then
  pass "the fixture's setup.sh exits non-zero: exit 2, stderr names the failed install, stdout empty"
else
  fail "the fixture's setup.sh exits non-zero: rc=$RC out='$OUT' err='$ERR'"
fi

# The guard is asked to install with HOME outside its own sandbox — a
# unit-level test, calling run_setup() directly with a mismatched HOME.
# There is no CLI shape that can reach this: the guard always builds HOME
# from the sandbox it just created itself, so the only way to pin the
# refusal is to call the function that performs it directly and prove
# subprocess.run is never reached. Modelled on
# scripts/test-check-plan-provenance.sh's own direct-import idiom
# (importlib.util.spec_from_file_location, registered in sys.modules
# before exec_module).
#
# A here-document is deliberately NOT used inside the command substitution
# below: bash 3.2 (the floor this harness runs under) mis-parses a
# here-document nested inside `$(...)`.
new_fixture_repo
OUTSIDE="$(mktemp -d "${TMPDIR:-/tmp}/check-installed-citations-outside.XXXXXX")"
SANDBOXES+=("$OUTSIDE")
UNIT_RESULT="$(python3 -c '
import importlib.util
import os
import sys

spec = importlib.util.spec_from_file_location(
    "cic_under_test", os.path.join(sys.argv[1], "check-installed-citations.py")
)
mod = importlib.util.module_from_spec(spec)
sys.modules[spec.name] = mod
spec.loader.exec_module(mod)

called = []
def fake_run(*a, **k):
    called.append((a, k))
    raise AssertionError("subprocess.run must not be called on a sandbox refusal")
mod.subprocess.run = fake_run

repo_root = sys.argv[2]
outside_home = sys.argv[3]
try:
    mod.run_setup(repo_root, repo_root, outside_home, None, "global")
    print("NO_REFUSAL")
except mod.SandboxRefusal:
    print("REFUSED:%d" % len(called))
' "$SCRIPT_DIR" "$FIXTURE" "$OUTSIDE")"
if [ "$UNIT_RESULT" = "REFUSED:0" ]; then
  pass "the guard is asked to install with HOME outside its own sandbox: refused, no setup.sh invocation"
else
  fail "the guard is asked to install with HOME outside its own sandbox: got '$UNIT_RESULT'"
fi

# ===========================================================================
# SECTION: Report shape (task 1 step 4) — these assert on runs the cases
# above already made, per that step's own instruction, plus one dedicated
# two-violation fixture to pin the violation count in the verdict line.
# ===========================================================================

# A clean fixture: rerun the "installed-root" case's fixture (no violation)
# and check the verdict shape rather than just its absence.
new_fixture_repo
write_file "skills/myflow-do/SKILL.md" '# myflow-do fixture
`skills/other/SKILL.md`
'
run_guard "$FIXTURE"
if [ "$RC" -eq 0 ] && printf '%s\n' "$OUT" | head -n 1 | grep -aq '^INSTALLED-CITATIONS-OK:'; then
  pass "a clean fixture: exit 0, a verdict line naming the root and the file count"
else
  fail "a clean fixture: rc=$RC out='$OUT'"
fi
if printf '%s\n' "$OUT" | tail -n +2 | grep -aq 'skills/myflow-do/SKILL.md'; then
  pass "a clean fixture: the coverage fragment names the scanned member"
else
  fail "a clean fixture: no coverage fragment in: $OUT"
fi

# A fixture with two violations.
new_fixture_repo
write_file "skills/myflow-do/SKILL.md" '# myflow-do fixture
`.myflow/project.md`
`scripts/check-vocabulary.sh`
'
run_guard "$FIXTURE"
VIOLATION_LINES="$(printf '%s\n' "$OUT" | grep -ac -- '^skills/myflow-do/SKILL.md:' || true)"
if [ "$RC" -eq 1 ] && [ "$VIOLATION_LINES" -eq 2 ] && printf '%s\n' "$OUT" | grep -aq -- '2 violation(s)'; then
  pass "a fixture with two violations: exit 1, two path:line lines, a verdict naming 2"
else
  fail "a fixture with two violations: rc=$RC lines=$VIOLATION_LINES out='$OUT'"
fi

# ===========================================================================
# SECTION: Undeclared zero coverage (per-task review finding 3) —
# scripts/lib/coverage.sh's coverage_declare/coverage_verdict mechanism,
# wired through the wrapper, must make a member's zero checked citations
# a violation UNLESS the wrapper declared that member expected-zero. This
# fixture is otherwise entirely clean (every citation, real or the
# baseline SAFE_CITATION each fixture file already carries — see that
# comment — resolves against a recognised root); the only violation it
# can produce is an undeclared zero. skills/myflow-do/SKILL.md here
# carries none of that baseline: it is a fresh, real-content file with no
# citation at all, and no fixture path is ever on the wrapper's declared
# list (that list is scoped to this repository's own real paths).
# ===========================================================================
new_fixture_repo
write_file "skills/myflow-do/SKILL.md" '# myflow-do fixture, genuinely no citations at all

Just prose.
'
run_guard "$FIXTURE"
if [ "$RC" -eq 1 ] \
  && printf '%s\n' "$OUT" | grep -aq -- '^skills/myflow-do/SKILL.md:0: 0 checked, and not declared expected-zero'; then
  pass "undeclared zero coverage is a violation, not a silent pass"
else
  fail "undeclared zero coverage: rc=$RC out='$OUT'"
fi

# ===========================================================================
if [ "$FAILURES" -eq 0 ]; then
  printf 'check-installed-citations: all cases pass\n'
  exit 0
else
  printf 'check-installed-citations: %d case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
