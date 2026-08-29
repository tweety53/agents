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
# CONCURRENT (KAN-362 task 3, design.md's citations-parallel-not-injected):
# every case's fixture is independent by construction (its own mktemp -d
# tree), but check-installed-citations.py/.sh are NOT touched — the
# rejected alternative was a sandbox injection point inside the guard
# itself. Instead this harness runs in two sequential passes over the same
# case list: a BUILD pass (register_case / register_job / push_job) that
# builds every fixture and stages one guard (or, for the HOME-outside unit
# test, one python3) invocation per case as a scripts/lib/parallel.sh job —
# fixture-building itself is cheap (no guard invocation), so nothing is
# lost running it single-threaded — then one parallel_run call that
# actually runs every guard/python3 invocation (the CPU-bound part)
# concurrently, then a CHECK pass that replays each case's already-captured
# OUT/ERR/RC through exactly the same judgment logic assert_case used to
# apply inline, in the original case order, so the ok:/FAIL: line order,
# the final tail line and the non-zero-on-any-failure exit code all survive
# unchanged — a reader of this harness's output cannot tell it became
# concurrent except from the wall clock. See register_job's own comment for
# why OUT and ERR are still captured on genuinely separate streams despite
# parallel.sh's own capture being merged.
#
# TRAP CHAINING (the amended requirement task 1's per-task review found):
# scripts/lib/parallel.sh installs its own EXIT/INT/TERM traps at source
# time, and this harness installs its own `trap cleanup EXIT` (plus INT and
# TERM) guarding its SANDBOXES fixture trees. Sourcing the library into this
# process would silently clobber whichever trap was installed first —
# either every fixture tree leaks, or the library's own tmp dir leaks — so
# this harness re-installs a combined trap immediately after sourcing the
# library, per that library's own header recipe.
#
# Bash 3.2 is the floor, as scripts/test-check-finish-preflight.sh's header
# records: indexed arrays only, no associative arrays.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-installed-citations.sh"
LIB="$SCRIPT_DIR/lib/parallel.sh"
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
trap 'cleanup; trap - INT; kill -s INT "$$"' INT
trap 'cleanup; trap - TERM; kill -s TERM "$$"' TERM

# shellcheck source=lib/parallel.sh
source "$LIB"

# scripts/lib/parallel.sh installs its own EXIT/INT/TERM traps at source
# time (see its header), silently overwriting the ones just above — which
# would leak either this harness's own SANDBOXES fixture trees or the
# library's own tmp dir depending on which trap happened to survive. Chain
# them explicitly, per that header's own instructions.
trap 'cleanup; _parallel_cleanup' EXIT
trap 'cleanup; _parallel_cleanup; trap - INT; kill -s INT "$$"' INT
trap 'cleanup; _parallel_cleanup; trap - TERM; kill -s TERM "$$"' TERM

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

# ===========================================================================
# BUILD-phase primitives. CMDS/ERRFILES/CHECKERS are indexed in parallel —
# job i's command is CMDS[i], its captured stderr file (when the job is a
# guard-CLI invocation) is ERRFILES[i], and the CHECK pass invokes
# "${CHECKERS[i]}" i once parallel_run has returned. Dense from 0 —every
# call site below sets CHECKERS[$IDX] immediately after obtaining IDX, so
# the CHECK pass never indexes a hole.
# ===========================================================================
CMDS=()
ERRFILES=()
CHECKERS=()
CASE_LABEL=()
CASE_RELPATH=()
CASE_EXPECT=()

# push_job <cmd> — low-level: appends <cmd> verbatim as the next
# parallel_run job. Sets IDX to its index.
push_job() {
  CMDS+=("$1")
  IDX=$((${#CMDS[@]} - 1))
}

# register_job <root> — stages a job that invokes the guard with
# CHECK_INSTALLED_CITATIONS_ROOT set to <root> (which may be empty, or a
# path that does not exist — the refusal cases need exactly that). stdout
# is left to parallel.sh's own per-job capture (becomes OUT); stderr is
# redirected inside the command string to a dedicated mktemp file this
# function creates, kept OUT of that merged capture on purpose — a
# refusal's "message on stderr, nothing on stdout" contract cannot be
# checked from a stream that merges the two. Sets IDX and ERRFILES[$IDX].
register_job() {
  local root="$1" errfile qguard qroot qerr
  errfile="$(mktemp "${TMPDIR:-/tmp}/check-installed-citations-stderr.XXXXXX")"
  SANDBOXES+=("$errfile")
  printf -v qguard '%q' "$GUARD"
  printf -v qroot '%q' "$root"
  printf -v qerr '%q' "$errfile"
  push_job "CHECK_INSTALLED_CITATIONS_ROOT=$qroot $qguard 2>$qerr"
  ERRFILES[$IDX]="$errfile"
}

# load_result <idx> — CHECK-phase companion to register_job: sets OUT, ERR,
# RC from job <idx>'s already-captured results.
load_result() {
  OUT="$(cat "${PARALLEL_OUTFILES[$1]}")"
  ERR="$(cat "${ERRFILES[$1]}")"
  RC="${PARALLEL_RC[$1]}"
}

# register_case <label> <relpath> <content> <expect> — BUILD-phase
# equivalent of the old assert_case: builds a fresh fixture, writes
# <content> plus SAFE_CITATION (see its own comment) to <relpath>, and
# stages a register_job for it. expect is "reported" or "not-reported".
register_case() {
  local label="$1" relpath="$2" content="$3" expect="$4" body
  new_fixture_repo
  body="$(printf '%s\n%s\n' "$content" "$SAFE_CITATION")"
  write_file "$relpath" "$body"
  register_job "$FIXTURE"
  CASE_LABEL[$IDX]="$label"
  CASE_RELPATH[$IDX]="$relpath"
  CASE_EXPECT[$IDX]="$expect"
  CHECKERS[$IDX]="check_generic"
}

# check_generic <idx> — CHECK-phase companion to register_case: the same
# reported/not-reported judgment assert_case used to make inline, now
# applied to already-captured results.
check_generic() {
  local idx="$1" label relpath expect
  label="${CASE_LABEL[$idx]}"
  relpath="${CASE_RELPATH[$idx]}"
  expect="${CASE_EXPECT[$idx]}"
  load_result "$idx"
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
# BUILD PHASE — every case below registers its fixture and its job; nothing
# here invokes the guard. Section comments are preserved from the original,
# single-pass version of this harness.
# ===========================================================================

# ---------------------------------------------------------------------------
# SECTION: The citation classifier (task 1 step 2's table)
# ---------------------------------------------------------------------------

register_case "bare-root-file" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture
`README.md`
' "reported"

register_case "bare-generic-filename" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture
`tasks.md`
' "not-reported"

register_case "installed-root" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture
`skills/other/SKILL.md`
' "not-reported"

register_case "agents-repo-prefix" "skills/myflow-do/SKILL.md" \
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
check_agents_repo_prefix_bogus_path() {
  local idx="$1"
  load_result "$idx"
  if [ "$RC" -eq 0 ] \
    && ! printf '%s\n' "$OUT" | grep -aq -- '^skills/myflow-do/SKILL.md:' \
    && printf '%s\n' "$OUT" | grep -aq -- 'skills/myflow-do/SKILL.md 1'; then
    pass "agents-repo-prefix-bogus-path: recognised (coverage count 1), not silently dropped"
  else
    fail "agents-repo-prefix-bogus-path: rc=$RC out='$OUT'"
  fi
}
new_fixture_repo
write_file "skills/myflow-do/SKILL.md" '# myflow-do fixture
`<agents repo>/does-not-exist/nested/path.md`
'
register_job "$FIXTURE"
CHECKERS[$IDX]="check_agents_repo_prefix_bogus_path"

register_case "project-prefix" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture
`<project>/.flow/project.md`
' "not-reported"

register_case "unrooted-directory" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture
`.flow/project.md`
' "reported"

register_case "unrooted-script" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture
`scripts/check-vocabulary.sh`
' "reported"

register_case "fenced-command" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture

```bash
git -C <abs-worktree> reset -- openspec/
```
' "not-reported"

register_case "fenced-comment" "skills/myflow-do/SKILL.md" \
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
register_case "fenced-comment-agents-repo-prefix" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture

```bash
# see <agents repo>/README.md for details
```
' "not-reported"

register_case "absolute-path" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture
`/etc/hosts`
' "not-reported"

register_case "home-path" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture
`~/.claude/skills/`
' "not-reported"

register_case "url" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture
`https://example.test/a/b`
' "not-reported"

register_case "shell-variable" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture
`$SCRIPT_DIR/lib/x.sh`
' "not-reported"

register_case "git-ref" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture
`origin/main`
' "not-reported"

register_case "regex-fragment" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture
`[A-Za-z0-9._-]+/x`
' "not-reported"

# opt-in-rule-out-of-scope — the citation sits in rules/opt-in-rule.mdc,
# which the fixture's own setup.sh only ever installs when alwaysApply is
# true (it is false here) — so this file is never part of the installed
# corpus and the guard must never even open it.
register_case "opt-in-rule-out-of-scope" "rules/opt-in-rule.mdc" \
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
check_newly_installed_directory() {
  local idx="$1"
  load_result "$idx"
  if [ "$RC" -eq 1 ] && printf '%s\n' "$OUT" | grep -aq -- '^docs/somefile.md:'; then
    pass "newly-installed-directory"
  else
    fail "newly-installed-directory: expected a violation naming docs/somefile.md, got rc=$RC out=$OUT err=$ERR"
  fi
}
new_fixture_repo
mkdir -p "$FIXTURE/docs"
write_file "docs/somefile.md" '# docs fixture

`README.md`
'
register_job "$FIXTURE"
CHECKERS[$IDX]="check_newly_installed_directory"

# ---------------------------------------------------------------------------
# SECTION: Task 9's three classifier exclusions — the real corpus proved
# necessary after wave B refused to force a root onto four tokens the
# classifier called citations that were not. Four cases: one per
# exclusion, plus placeholder-rooted-is-recognised, which draws the same
# recognised-versus-never-seen distinction agents-repo-prefix-bogus-path
# already draws (that exact collapse is what hid this change's critical
# defect behind a green test the first time).
# ---------------------------------------------------------------------------

register_case "parent-relative-path" "skills/myflow-do/SKILL.md" '# myflow-do fixture
`../<other-app>`
' "not-reported"

register_case "placeholder-rooted" "skills/myflow-do/SKILL.md" '# myflow-do fixture
`<state-dir>/<name>-proposal-artifact.html`
' "not-reported"

register_case "quoted-program-output" "skills/myflow-do/SKILL.md" "# myflow-do fixture
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
check_placeholder_rooted_is_recognised() {
  local idx="$1"
  load_result "$idx"
  if [ "$RC" -eq 0 ] \
    && ! printf '%s\n' "$OUT" | grep -aq -- '^skills/myflow-do/SKILL.md:' \
    && printf '%s\n' "$OUT" | grep -aq -- 'skills/myflow-do/SKILL.md 1'; then
    pass "placeholder-rooted-is-recognised: recognised (coverage count 1), not silently dropped"
  else
    fail "placeholder-rooted-is-recognised: rc=$RC out='$OUT'"
  fi
}
new_fixture_repo
write_file "skills/myflow-do/SKILL.md" '# myflow-do fixture
`<state-dir>/does-not-exist.html`
'
register_job "$FIXTURE"
CHECKERS[$IDX]="check_placeholder_rooted_is_recognised"

# unrecognised-placeholder-is-reported — per the per-task review's
# critical finding: the first cut of the placeholder rule accepted ANY
# bracket-shaped first segment, which fails open — `<foo>/…` passed while
# naming an unrooted path, and so would a typo of a real placeholder
# (`<changeroot>/`, `<change-root>/`). This is the case whose ABSENCE let
# that through: every earlier placeholder case only ever exercised a
# MEMBER of the closed set, never a non-member, so a bracket-shape-only
# implementation passed all of them anyway.
register_case "unrecognised-placeholder-is-reported" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture
`<foo>/spectre/specs/x.md`
' "reported"

# ---------------------------------------------------------------------------
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
# ---------------------------------------------------------------------------

register_case "git-branch-name-is-not-a-citation" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture
Branch `spectre/<name>`.
' "not-reported"

register_case "file-line-reference-is-not-a-citation" "skills/myflow-do/SKILL.md" \
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
check_abs_worktree_rooted_is_recognised() {
  local idx="$1"
  load_result "$idx"
  if [ "$RC" -eq 0 ] \
    && ! printf '%s\n' "$OUT" | grep -aq -- '^skills/myflow-do/SKILL.md:' \
    && printf '%s\n' "$OUT" | grep -aq -- 'skills/myflow-do/SKILL.md 1'; then
    pass "abs-worktree-rooted-is-recognised: recognised (coverage count 1), not silently dropped"
  else
    fail "abs-worktree-rooted-is-recognised: rc=$RC out='$OUT'"
  fi
}
new_fixture_repo
write_file "skills/myflow-do/SKILL.md" '# myflow-do fixture
`<abs-worktree>/.superpowers/sdd/does-not-exist.diff`
'
register_job "$FIXTURE"
CHECKERS[$IDX]="check_abs_worktree_rooted_is_recognised"

# ---------------------------------------------------------------------------
# SECTION: Second per-task review — three bugs the panel found in what
# task 10 shipped: the `origin` exclusion excluded ANY token whose first
# segment was `origin`, not only a ref shape; extract_backtick_tokens kept
# only a span's leading word, so a REAL second citation in the same span
# went not merely unreported but never seen; FILE_LINE_RE was confirmed,
# not merely left alone — narrowing it was tried and found to force
# re-rooting SKILL.md:485's own fabricated findings-table example.
# ---------------------------------------------------------------------------

register_case "origin-with-extension-is-a-citation" "skills/myflow-do/SKILL.md" '# myflow-do fixture
`origin/README.md`
' "reported"

register_case "origin-ref-with-nested-path-is-not-a-citation" "skills/myflow-do/SKILL.md" '# myflow-do fixture
`origin/spectre/<name>`
' "not-reported"

# origin-directory-is-a-citation — panel round 3's finding: the ORIGINAL
# bound ("remainder carries no file extension") also silently excluded a
# DIRECTORY, since a directory has no extension either but is never a
# git ref. `origin/rules/` (trailing slash) must be classified and
# reported, not vanish.
register_case "origin-directory-is-a-citation" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture
`origin/rules/`
' "reported"

register_case "second-word-in-backtick-span-is-seen" "skills/myflow-do/SKILL.md" '# myflow-do fixture
`see .flow/project.md`
' "reported"

register_case "shell-example-second-word-not-path-shaped" "skills/myflow-do/SKILL.md" '# myflow-do fixture
`skills/other/SKILL.md verbose`
' "not-reported"

# file-line-reference-is-not-a-citation (task 10, above) already pins
# FILE_LINE_RE's kept, undivided behaviour — re-confirmed rather than
# re-tested here; see this guard's own module docstring for the narrowing
# attempt and why it was rejected.

# ---------------------------------------------------------------------------
# SECTION: Third per-task review — a sixth placeholder root (<skill-dir>,
# collapsing six real corpus sites' three different wordings of the same
# concept) and Group B split into three separately-bounded exclusions
# rather than one that would have swallowed all three.
# ---------------------------------------------------------------------------

register_case "skill-dir-rooted" "skills/myflow-do/SKILL.md" '# myflow-do fixture
`<skill-dir>/scripts/check-vocabulary.sh`
' "not-reported"

# skill-dir-rooted-is-recognised — the same recognised-versus-never-seen
# distinction every root in this guard is held to. The path here does not
# exist (`<skill-dir>/scripts/does-not-exist.sh`): existence is never the
# guard's business for a slash-containing token, so a bogus target proves
# nothing except whether the CITATION ITSELF was seen — proven by its
# member's coverage count going non-zero, not merely by the absence of a
# violation line.
check_skill_dir_rooted_is_recognised() {
  local idx="$1"
  load_result "$idx"
  if [ "$RC" -eq 0 ] \
    && ! printf '%s\n' "$OUT" | grep -aq -- '^skills/myflow-do/SKILL.md:' \
    && printf '%s\n' "$OUT" | grep -aq -- 'skills/myflow-do/SKILL.md 1'; then
    pass "skill-dir-rooted-is-recognised: recognised (coverage count 1), not silently dropped"
  else
    fail "skill-dir-rooted-is-recognised: rc=$RC out='$OUT'"
  fi
}
new_fixture_repo
write_file "skills/myflow-do/SKILL.md" '# myflow-do fixture
`<skill-dir>/scripts/does-not-exist.sh`
'
register_job "$FIXTURE"
CHECKERS[$IDX]="check_skill_dir_rooted_is_recognised"

register_case "spectre-branch-with-change-name-is-not-a-citation" "skills/myflow-do/SKILL.md" '# myflow-do fixture
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
register_case "spectre-shape-with-real-path-is-still-reported" "skills/myflow-do/SKILL.md" '# myflow-do fixture
`spectre/specs/x.md`
' "reported"

register_case "spectre-shape-with-trailing-path-is-still-reported" "skills/myflow-do/SKILL.md" '# myflow-do fixture
`spectre/changes/<name>/`
' "reported"

# ---------------------------------------------------------------------------
# SECTION: Task 9 (kan-239) — a second pipeline-created branch shape,
# `chore/archive-<name>`, needs the same GIT_BRANCH_SPECTRE_RE treatment:
# a branch name is not a path citation. Two cases, same bound as the
# spectre pair above: the bare shape is not reported, and a bracket
# segment followed by more path — `chore/archive-<name>/spec.md` — is the
# fail-open bound and must stay reported.
# ---------------------------------------------------------------------------

register_case "chore-archive-branch-name-is-not-a-citation" "skills/myflow-do/SKILL.md" '# myflow-do fixture
Branch `chore/archive-<name>`.
' "not-reported"

register_case "chore-archive-shape-with-trailing-path-is-still-reported" "skills/myflow-do/SKILL.md" '# myflow-do fixture
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
register_case "chore-archive-bracket-cannot-smuggle-a-slash" "skills/myflow-do/SKILL.md" '# myflow-do fixture
`chore/archive-<a/b>`
' "reported"

register_case "html-comment-is-not-a-citation" "skills/myflow-do/SKILL.md" '# myflow-do fixture
`<!-- measured: ./gradlew test @ c515c42 -->`
' "not-reported"

# html-comment-with-leading-whitespace-is-not-a-citation — panel round 5
# (the panel's code-quality reviewer): a space right inside the opening backtick
# (`` ` <!-- ... -->` ``) must not defeat the span-initial check. Dormant
# in the real corpus today (every live site is span-initial) but
# untested before this case, and the reviewer showed the fragile
# `startswith` alone lets `./gradlew` fall through to word-splitting and
# report.
register_case "html-comment-with-leading-whitespace-is-not-a-citation" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture
` <!-- measured: ./gradlew test @ c515c42 -->`
' "not-reported"

register_case "colon-segment-is-not-a-citation" "skills/myflow-do/SKILL.md" '# myflow-do fixture
`measured:/predicted:`
' "not-reported"

# final-colon-segment-is-a-citation — panel round 3's finding: the
# ORIGINAL rule tested EVERY segment including the last, which silently
# dropped a real citation carrying a trailing colon inside its own span
# (e.g. "see `.flow/project.md:` for the list"). Only a colon on a
# NON-FINAL segment signals "not a path" now.
register_case "final-colon-segment-is-a-citation" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture
`.flow/project.md:`
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
register_case "redirection-is-not-merged-into-a-citation" "skills/myflow-do/SKILL.md" \
  '# myflow-do fixture
Run `some-cmd < skills/other/SKILL.md > output.log` to reproduce.
' "not-reported"

# ---------------------------------------------------------------------------
# SECTION: Refusal cases (task 1 step 3's table). Every refusal case
# asserts stdout is empty as well as the exit code.
# ---------------------------------------------------------------------------

# CHECK_INSTALLED_CITATIONS_ROOT set but empty.
check_empty_root() {
  local idx="$1"
  load_result "$idx"
  if [ "$RC" -eq 2 ] && [ -z "$OUT" ] && printf '%s' "$ERR" | grep -aq 'CHECK_INSTALLED_CITATIONS_ROOT'; then
    pass "CHECK_INSTALLED_CITATIONS_ROOT set but empty: exit 2, stderr names it, stdout empty"
  else
    fail "CHECK_INSTALLED_CITATIONS_ROOT set but empty: rc=$RC out='$OUT' err='$ERR'"
  fi
}
register_job ""
CHECKERS[$IDX]="check_empty_root"

# Root is not a directory.
check_root_not_a_directory() {
  local idx="$1"
  load_result "$idx"
  if [ "$RC" -eq 2 ] && [ -z "$OUT" ]; then
    pass "root is not a directory: exit 2, stdout empty"
  else
    fail "root is not a directory: rc=$RC out='$OUT' err='$ERR'"
  fi
}
register_job "${TMPDIR:-/tmp}/check-installed-citations-does-not-exist-$$"
CHECKERS[$IDX]="check_root_not_a_directory"

# The fixture's setup.sh exits non-zero.
check_setup_sh_fails() {
  local idx="$1"
  load_result "$idx"
  if [ "$RC" -eq 2 ] && [ -z "$OUT" ] && printf '%s' "$ERR" | grep -aq 'setup.sh'; then
    pass "the fixture's setup.sh exits non-zero: exit 2, stderr names the failed install, stdout empty"
  else
    fail "the fixture's setup.sh exits non-zero: rc=$RC out='$OUT' err='$ERR'"
  fi
}
FIXTURE="$(mktemp -d "${TMPDIR:-/tmp}/check-installed-citations-fixture.XXXXXX")"
SANDBOXES+=("$FIXTURE")
cat > "$FIXTURE/setup.sh" <<'EOF'
#!/usr/bin/env bash
echo "fixture setup.sh: deliberately failing" >&2
exit 1
EOF
chmod +x "$FIXTURE/setup.sh"
mkdir -p "$FIXTURE/skills" "$FIXTURE/rules"
register_job "$FIXTURE"
CHECKERS[$IDX]="check_setup_sh_fails"

# The guard is asked to install with HOME outside its own sandbox — a
# unit-level test, calling run_setup() directly with a mismatched HOME.
# There is no CLI shape that can reach this: the guard always builds HOME
# from the sandbox it just created itself, so the only way to pin the
# refusal is to call the function that performs it directly and prove
# subprocess.run is never reached. Modelled on
# scripts/test-check-plan-provenance.sh's own direct-import idiom
# (importlib.util.spec_from_file_location, registered in sys.modules
# before exec_module). The python source is staged to its own mktemp file
# (rather than embedded in the job's command string) so it never has to
# survive a second round of shell quoting on the way into a fresh `bash -c`
# child process.
check_home_outside_sandbox() {
  local idx="$1" result
  result="$(cat "${PARALLEL_OUTFILES[$idx]}")"
  if [ "$result" = "REFUSED:0" ]; then
    pass "the guard is asked to install with HOME outside its own sandbox: refused, no setup.sh invocation"
  else
    fail "the guard is asked to install with HOME outside its own sandbox: got '$result'"
  fi
}
new_fixture_repo
OUTSIDE="$(mktemp -d "${TMPDIR:-/tmp}/check-installed-citations-outside.XXXXXX")"
SANDBOXES+=("$OUTSIDE")
PYFILE="$(mktemp "${TMPDIR:-/tmp}/check-installed-citations-home-outside.XXXXXX")"
SANDBOXES+=("$PYFILE")
cat > "$PYFILE" <<'PYEOF'
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
PYEOF
HOME_OUTSIDE_ERR="$(mktemp "${TMPDIR:-/tmp}/check-installed-citations-home-outside-stderr.XXXXXX")"
SANDBOXES+=("$HOME_OUTSIDE_ERR")
printf -v qpy '%q' "$PYFILE"
printf -v qscriptdir '%q' "$SCRIPT_DIR"
printf -v qfixture '%q' "$FIXTURE"
printf -v qoutside '%q' "$OUTSIDE"
printf -v qerr '%q' "$HOME_OUTSIDE_ERR"
push_job "python3 $qpy $qscriptdir $qfixture $qoutside 2>$qerr"
CHECKERS[$IDX]="check_home_outside_sandbox"

# ---------------------------------------------------------------------------
# SECTION: Report shape (task 1 step 4) — these assert on runs the cases
# above already made, per that step's own instruction, plus one dedicated
# two-violation fixture to pin the violation count in the verdict line.
# ---------------------------------------------------------------------------

# A clean fixture: rerun the "installed-root" case's fixture (no violation)
# and check the verdict shape rather than just its absence.
check_clean_fixture_report_shape() {
  local idx="$1"
  load_result "$idx"
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
}
new_fixture_repo
write_file "skills/myflow-do/SKILL.md" '# myflow-do fixture
`skills/other/SKILL.md`
'
register_job "$FIXTURE"
CHECKERS[$IDX]="check_clean_fixture_report_shape"

# A fixture with two violations.
check_two_violations() {
  local idx="$1" violation_lines
  load_result "$idx"
  violation_lines="$(printf '%s\n' "$OUT" | grep -ac -- '^skills/myflow-do/SKILL.md:' || true)"
  if [ "$RC" -eq 1 ] && [ "$violation_lines" -eq 2 ] && printf '%s\n' "$OUT" | grep -aq -- '2 violation(s)'; then
    pass "a fixture with two violations: exit 1, two path:line lines, a verdict naming 2"
  else
    fail "a fixture with two violations: rc=$RC lines=$violation_lines out='$OUT'"
  fi
}
new_fixture_repo
write_file "skills/myflow-do/SKILL.md" '# myflow-do fixture
`.flow/project.md`
`scripts/check-vocabulary.sh`
'
register_job "$FIXTURE"
CHECKERS[$IDX]="check_two_violations"

# ---------------------------------------------------------------------------
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
# ---------------------------------------------------------------------------
check_undeclared_zero_coverage() {
  local idx="$1"
  load_result "$idx"
  if [ "$RC" -eq 1 ] \
    && printf '%s\n' "$OUT" | grep -aq -- '^skills/myflow-do/SKILL.md:0: 0 checked, and not declared expected-zero'; then
    pass "undeclared zero coverage is a violation, not a silent pass"
  else
    fail "undeclared zero coverage: rc=$RC out='$OUT'"
  fi
}
new_fixture_repo
write_file "skills/myflow-do/SKILL.md" '# myflow-do fixture, genuinely no citations at all

Just prose.
'
register_job "$FIXTURE"
CHECKERS[$IDX]="check_undeclared_zero_coverage"

# ===========================================================================
# RUN PHASE — every guard/python3 invocation staged above runs concurrently
# through scripts/lib/parallel.sh.
# ===========================================================================
set +e
parallel_run "${CMDS[@]}"
PARALLEL_RUN_RC=$?
set -e
if [ "$PARALLEL_RUN_RC" -eq 2 ]; then
  printf 'check-installed-citations: could not run the cases concurrently (see above)\n' >&2
  exit 2
fi

# ===========================================================================
# CHECK PHASE — replays every case's already-captured result through its
# own checker, in the original build order, so ok:/FAIL: lines print in
# exactly the order a reader of the old, sequential harness would have
# seen them.
# ===========================================================================
# CHECK_ORDER (F4) records the index this loop invokes its checker at, in
# the order it actually happens — the assertion right after the loop pins
# the comment's own contract (ascending build order) so a mutation that
# reverses this loop (e.g. counting from ${#CMDS[@]}-1 down to 0, which
# still runs every case and still exits 0 with all cases passing — nothing
# else here would notice) is caught HERE rather than passing silently.
CHECK_ORDER=()
i=0
n=${#CMDS[@]}
while [ "$i" -lt "$n" ]; do
  CHECK_ORDER+=("$i")
  "${CHECKERS[$i]}" "$i"
  i=$((i + 1))
done
EXPECTED_ORDER="$(i=0; while [ "$i" -lt "$n" ]; do printf '%s ' "$i"; i=$((i + 1)); done)"
ACTUAL_ORDER="$(printf '%s ' "${CHECK_ORDER[@]}")"
if [ "$ACTUAL_ORDER" = "$EXPECTED_ORDER" ]; then
  pass "CHECK PHASE replays every case in ascending build order"
else
  fail "CHECK PHASE order broken: expected [$EXPECTED_ORDER] got [$ACTUAL_ORDER]"
fi

# ---------------------------------------------------------------------------
# SECTION: EXIT-trap chaining (KAN-362 task 3's amended requirement).
# Proven against a small, dedicated child script reproducing — verbatim —
# this file's own recipe (trap cleanup EXIT/INT/TERM installed, THEN
# scripts/lib/parallel.sh sourced, THEN the combined trap re-installed),
# rather than by interrupting this harness's own 52-case BUILD/RUN:
# bash defers running a trapped signal's handler until the foreground
# command it is currently blocked on returns (documented bash behaviour),
# so an interrupt sent to a live `parallel_run` of 52 guard invocations
# would not actually take effect until that run's own natural completion —
# proving nothing extra while costing this harness's full wall clock a
# second time. The child sources the REAL scripts/lib/parallel.sh (not a
# stand-in) and calls its real parallel_run with one short `sleep` job, so
# the deferred-trap cost here is bounded by that job's own duration.
#
# Two sub-cases: a normal (uninterrupted) exit, and a real SIGINT sent to a
# real separate `bash` child process. `set -m` (job control) is required
# for the interrupt sub-case: a background job started with plain `cmd &`
# in a non-interactive shell has SIGINT/SIGQUIT set to ignored by bash
# itself, so `kill -INT` on it is silently swallowed and this assertion
# would pass even when the chaining above is broken, since the child would
# never actually see the signal — the exact false negative the task's own
# instructions warn about.
# ---------------------------------------------------------------------------

# write_trap_chain_child <path> — writes a runnable reproduction of this
# file's own EXIT/INT/TERM chaining recipe to <path>. Takes one argument
# at run time: how long its one parallel_run job should sleep.
write_trap_chain_child() {
  cat > "$1" <<CHILDEOF
#!/usr/bin/env bash
set -euo pipefail
SANDBOXES=()
cleanup() {
  [ "\${#SANDBOXES[@]}" -eq 0 ] && return 0
  local s
  for s in "\${SANDBOXES[@]}"; do rm -rf "\$s"; done
}
trap cleanup EXIT
trap 'cleanup; trap - INT; kill -s INT "\$\$"' INT
trap 'cleanup; trap - TERM; kill -s TERM "\$\$"' TERM

FIXTURE="\$(mktemp -d "\${TMPDIR:-/tmp}/trap-chain-fixture.XXXXXX")"
SANDBOXES+=("\$FIXTURE")

# shellcheck source=lib/parallel.sh
source "$LIB"

trap 'cleanup; _parallel_cleanup' EXIT
trap 'cleanup; _parallel_cleanup; trap - INT; kill -s INT "\$\$"' INT
trap 'cleanup; _parallel_cleanup; trap - TERM; kill -s TERM "\$\$"' TERM

parallel_run "sleep \${1:-0}"
CHILDEOF
  chmod +x "$1"
}

# leftover_trap_chain_dirs <base-dir> — the set of this section's own
# fixture/library tmp dirs currently on disk under <base-dir>, one per
# line, sorted. A before/after diff of this (never a bare emptiness check)
# is what lets this run correctly even though THIS process's own 52 cases
# already left their own — not yet cleaned, still pending this process's
# own EXIT trap — fixture trees sitting in the same directory.
#
# <base-dir> is each sub-case's own PRIVATE TMPDIR (its own mktemp -d),
# not the global $TMPDIR: scripts/run-guard-tests.sh runs 40+ harnesses
# concurrently via scripts/lib/parallel.sh, and this file, itself and
# scripts/test-lib-parallel.sh all create trap-chain-fixture.*/
# parallel-lib.* directories in that same shared $TMPDIR at the same time.
# Scanning the shared namespace can catch a SIBLING's live (not-yet-
# cleaned-up) directory in the "after" snapshot and misreport it as this
# sub-case's own leak — the exact failure reproduced ("a clean exit
# leaked: .../parallel-lib.5obuu0" while nothing here actually leaked). A
# private TMPDIR, passed to the child via env var, makes the namespace
# exclusive to that one child, so nothing else can ever appear in it.
leftover_trap_chain_dirs() {
  find "$1" -maxdepth 1 \
    \( -name 'trap-chain-fixture.*' -o -name 'parallel-lib.*' \) \
    2>/dev/null | sort
}

CHILD="$(mktemp "${TMPDIR:-/tmp}/check-installed-citations-trap-chain-child.XXXXXX")"
SANDBOXES+=("$CHILD")
write_trap_chain_child "$CHILD"

# Sub-case 1: a normal, uninterrupted exit — both traps must still have
# fired, since sourcing the library clobbers the EXIT trap too, not only
# INT/TERM.
PRIVATE_TMP_SUB1="$(mktemp -d "${TMPDIR:-/tmp}/trap-chain-sub1.XXXXXX")"
SANDBOXES+=("$PRIVATE_TMP_SUB1")
before="$(leftover_trap_chain_dirs "$PRIVATE_TMP_SUB1")"
TMPDIR="$PRIVATE_TMP_SUB1" bash "$CHILD" 0 >/dev/null 2>&1
after="$(leftover_trap_chain_dirs "$PRIVATE_TMP_SUB1")"
new_leftover="$(comm -13 <(printf '%s\n' "$before") <(printf '%s\n' "$after"))"
if [ -z "$new_leftover" ]; then
  pass "a clean exit leaves no fixture tree and no parallel.sh tmp dir behind"
else
  fail "a clean exit leaked: $new_leftover"
fi
rm -rf "$PRIVATE_TMP_SUB1"

# Sub-case 2: a real SIGINT to a real separate bash child process.
PRIVATE_TMP_SUB2="$(mktemp -d "${TMPDIR:-/tmp}/trap-chain-sub2.XXXXXX")"
SANDBOXES+=("$PRIVATE_TMP_SUB2")
before="$(leftover_trap_chain_dirs "$PRIVATE_TMP_SUB2")"
(
  set -m
  TMPDIR="$PRIVATE_TMP_SUB2" bash "$CHILD" 1 >/dev/null 2>&1 &
  child_pid=$!
  sleep 0.3
  kill -INT "$child_pid" 2>/dev/null || true
  wait "$child_pid" 2>/dev/null || true
) 2>/dev/null
after="$(leftover_trap_chain_dirs "$PRIVATE_TMP_SUB2")"
new_leftover="$(comm -13 <(printf '%s\n' "$before") <(printf '%s\n' "$after"))"
if [ -z "$new_leftover" ]; then
  pass "a real SIGINT during a live run leaves no fixture tree and no parallel.sh tmp dir behind"
else
  fail "a real SIGINT during a live run leaked: $new_leftover"
fi
rm -rf "$PRIVATE_TMP_SUB2"

# ===========================================================================
if [ "$FAILURES" -eq 0 ]; then
  printf 'check-installed-citations: all cases pass\n'
  exit 0
else
  printf 'check-installed-citations: %d case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
