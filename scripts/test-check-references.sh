#!/usr/bin/env bash
# Assertion harness for check-references.sh. Builds fixture trees under a
# sandboxed TMPDIR and asserts the guard's exit status and output. Never
# touches the real repository tree.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-references.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

# run_guard <fixture_dir> -> sets RC and OUT
#
# Points the guard at the fixture via CHECK_REFERENCES_ROOT rather than `cd`-ing
# into it. The guard resolves REPO_ROOT from its own BASH_SOURCE by default
# (so a real invocation from any cwd still scans this repo, never silently
# nothing) — CHECK_REFERENCES_ROOT is the one explicit, opt-in override that
# lets this harness exercise it against a sandboxed fixture tree without that
# real-repo default getting in the way.
run_guard() {
  set +e
  OUT="$(CHECK_REFERENCES_ROOT="$1" "$GUARD" 2>&1)"
  RC=$?
  set -e
}

new_fixture() {
  FIXTURE="$(mktemp -d "${TMPDIR:-/tmp}/refs-test.XXXXXX")"
  mkdir -p "$FIXTURE/rules" "$FIXTURE/skills/demo"
}

# 1. A live reference passes.
new_fixture
printf '## State file\n\nbody\n' > "$FIXTURE/rules/never-touch-production.mdc"
printf 'Resolve it per **State file** in `rules/never-touch-production.mdc`.\n' \
  > "$FIXTURE/skills/demo/SKILL.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "live reference passes" || fail "live reference: rc=$RC out=$OUT"

# 2. A moved section fails and is reported as file:line.
new_fixture
printf '## Something else\n\nbody\n' > "$FIXTURE/rules/never-touch-production.mdc"
printf 'Resolve it per **State file** in `rules/never-touch-production.mdc`.\n' \
  > "$FIXTURE/skills/demo/SKILL.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "moved section fails" || fail "moved section: expected non-zero"
case "$OUT" in
  *"skills/demo/SKILL.md:1"*) pass "reports file:line" ;;
  *) fail "reports file:line: out=$OUT" ;;
esac

# 3. All three phrasing variants are checked identically.
new_fixture
printf '## State file\n\nbody\n' > "$FIXTURE/rules/never-touch-production.mdc"
{
  printf 'see **State file** in `rules/never-touch-production.mdc`\n'
  printf 'per **State file** in `rules/never-touch-production.mdc`\n'
  printf 'defined once under **State file** in `rules/never-touch-production.mdc`\n'
} > "$FIXTURE/skills/demo/SKILL.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "phrasing variants pass" || fail "phrasing variants: out=$OUT"

new_fixture
printf '## Gone\n\nbody\n' > "$FIXTURE/rules/never-touch-production.mdc"
{
  printf 'see **State file** in `rules/never-touch-production.mdc`\n'
  printf 'per **State file** in `rules/never-touch-production.mdc`\n'
  printf 'defined once under **State file** in `rules/never-touch-production.mdc`\n'
} > "$FIXTURE/skills/demo/SKILL.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "phrasing variants all fail when absent" || fail "variants: expected non-zero"
for n in 1 2 3; do
  case "$OUT" in
    *"SKILL.md:$n"*) ;;
    *) fail "variant on line $n not reported: out=$OUT" ;;
  esac
done

# 4. Multi-section lines pass when at least one bold token resolves.
new_fixture
printf '## Stage transitions\n\nbody\n\n## State file\n\nbody\n' \
  > "$FIXTURE/rules/never-touch-production.mdc"
printf 'Follow `rules/never-touch-production.mdc` — sections **Stage transitions**, **State file**.\n' \
  > "$FIXTURE/skills/demo/SKILL.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "multi-section line passes" || fail "multi-section: out=$OUT"

# 5. An allow marker suppresses a line. Written to a declared expected-zero
# member: with the only line's bold token suppressed, this file's own
# checked-reference count is genuinely zero.
new_fixture
mkdir -p "$FIXTURE/skills/openspec-explore"
printf '## Gone\n\nbody\n' > "$FIXTURE/rules/never-touch-production.mdc"
printf '**Do not** copy from `rules/never-touch-production.mdc` <!-- refs-guard:allow -->\n' \
  > "$FIXTURE/skills/openspec-explore/SKILL.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "allow marker suppresses" || fail "allow marker: out=$OUT"

# 6. Fenced code blocks are skipped. Written to a declared expected-zero
# member: with the only candidate reference fenced out, this file's own
# checked-reference count is genuinely zero.
new_fixture
mkdir -p "$FIXTURE/skills/openspec-explore"
printf '## Gone\n\nbody\n' > "$FIXTURE/rules/never-touch-production.mdc"
{
  printf '```\n'
  printf 'see **State file** in `rules/never-touch-production.mdc`\n'
  printf '```\n'
} > "$FIXTURE/skills/openspec-explore/SKILL.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "fenced block skipped" || fail "fenced block: out=$OUT"

# 7. A path that does not resolve to a file is skipped, not failed. Written
# to a declared expected-zero member (skills/openspec-explore/SKILL.md):
# with the reference unresolvable, this file's own checked-reference count is
# genuinely zero, and it is the only file in this fixture's corpus.
new_fixture
mkdir -p "$FIXTURE/skills/openspec-explore"
printf 'see **Whatever** in `openspec/changes/<name>/tasks.md`\n' \
  > "$FIXTURE/skills/openspec-explore/SKILL.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "unresolvable path skipped" || fail "unresolvable path: out=$OUT"

# 8. Paths relative to the referring file resolve. The passing half alone would
# be vacuous — an unresolvable path is skipped, which also exits 0 — so the
# failing half below is what proves the relative candidate is really opened.
new_fixture
mkdir -p "$FIXTURE/skills/openspec-explore"
printf '## Panel re-runs\n\nbody\n' > "$FIXTURE/skills/openspec-explore/SKILL.md"
printf 'Follow **Panel re-runs** in `../openspec-explore/SKILL.md`.\n' \
  > "$FIXTURE/skills/demo/SKILL.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "relative path resolves" || fail "relative path: out=$OUT"

new_fixture
mkdir -p "$FIXTURE/skills/openspec-explore"
printf '## Something else\n\nbody\n' > "$FIXTURE/skills/openspec-explore/SKILL.md"
printf 'Follow **Panel re-runs** in `../openspec-explore/SKILL.md`.\n' \
  > "$FIXTURE/skills/demo/SKILL.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "a stale reference through a relative path still fails" \
  || fail "relative path stale ref: expected non-zero, out=$OUT"

# 9. Headings match case-insensitively and ignore backticks in the heading.
new_fixture
printf '## The `widget` setting\n\nbody\n' > "$FIXTURE/rules/never-touch-production.mdc"
printf 'see **The widget setting** in `rules/never-touch-production.mdc`\n' \
  > "$FIXTURE/skills/demo/SKILL.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "heading normalization" || fail "heading normalization: out=$OUT"

# 10. With no CHECK_REFERENCES_ROOT override, the guard ignores cwd entirely
# and always scans the real repo (BASH_SOURCE-derived REPO_ROOT) — this is
# what makes it argument-free and self-scoped from any directory. Proven by
# seeding a stale reference inside an unrelated fixture used as cwd: if the
# guard ever reported it, REPO_ROOT would have to be cwd-derived again, which
# is exactly the false-clean-from-elsewhere regression this assertion exists
# to catch. This assumes the real repo itself is clean at harness-run time,
# which check-references.sh's own baseline (verified separately) guarantees.
new_fixture
printf '## Gone\n\nbody\n' > "$FIXTURE/rules/never-touch-production.mdc"
printf 'Resolve it per **State file** in `rules/never-touch-production.mdc`.\n' \
  > "$FIXTURE/skills/demo/SKILL.md"
set +e
OUT="$(cd "$FIXTURE" && env -u CHECK_REFERENCES_ROOT "$GUARD" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 0 ] && pass "no-override run ignores cwd, scans the real repo instead" \
  || fail "no-override run: expected rc=0 (real repo clean), got rc=$RC out=$OUT"

# 11. A literal "**" inside inline code must not desync bold-delimiter
# counting/pairing. The reference resolves (a live "State file" heading), so
# this must PASS despite the odd-looking raw "**" count on the line.
new_fixture
printf '## State file\n\nbody\n' > "$FIXTURE/rules/never-touch-production.mdc"
printf 'See **State file** and code `x**y` in `rules/never-touch-production.mdc`.\n' \
  > "$FIXTURE/skills/demo/SKILL.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "in-code ** does not desync a live reference" \
  || fail "in-code **: out=$OUT"

# 12. The same in-code "**" shape must NOT blind the guard to a genuinely
# stale reference on the same line — masking must not become a new way to
# suppress a real hit.
new_fixture
printf '## Something else\n\nbody\n' > "$FIXTURE/rules/never-touch-production.mdc"
printf 'See **Gone** and code `x**y` in `rules/never-touch-production.mdc`.\n' \
  > "$FIXTURE/skills/demo/SKILL.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "in-code ** does not mask a genuinely stale reference" \
  || fail "in-code ** masking stale ref: expected non-zero, out=$OUT"

# 13. Round 1's soft-wrapped-bold-span fix must not regress now that bold
# extraction runs on the code-masked line.
new_fixture
printf '## Jira integration\n\nbody\n' > "$FIXTURE/rules/never-touch-production.mdc"
printf 'sync** in **Jira integration** (`rules/never-touch-production.mdc`) — canonical.\n' \
  > "$FIXTURE/skills/demo/SKILL.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "soft-wrapped bold span still tokenizes correctly" \
  || fail "soft-wrapped bold span regressed: out=$OUT"

# 14. A single `#` title is a heading like any other — headings_of must not skip
# level-1. The referenced file carries a SECOND, non-matching `##` heading on
# purpose: without it, a regression that skipped level-1 would leave the file
# with no headings at all, the guard would skip the reference, and this
# assertion would pass for the wrong reason. With it, skipping level-1 leaves a
# non-empty heading set that does not contain "state file", so the reference
# fails and this assertion catches the regression. (#15 covers the other half —
# that a genuinely stale reference to an `#`-titled file still fails.)
new_fixture
printf '# State file\n\n## Body notes\n\nbody\n' > "$FIXTURE/rules/never-touch-production.mdc"
printf 'Resolve it per **State file** in `rules/never-touch-production.mdc`.\n' \
  > "$FIXTURE/skills/demo/SKILL.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "a live reference to an #-titled file passes" \
  || fail "#-titled file: rc=$RC out=$OUT"

# 15. The same #-titled file must still catch a genuinely stale reference —
# proving the check runs at all, not merely that it always passes.
new_fixture
printf '# Something else\n\nbody\n' > "$FIXTURE/rules/never-touch-production.mdc"
printf 'Resolve it per **State file** in `rules/never-touch-production.mdc`.\n' \
  > "$FIXTURE/skills/demo/SKILL.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "a stale reference to an #-titled file still fails" \
  || fail "#-titled file stale ref: expected non-zero, out=$OUT"

# 16. A `#` COMMENT inside a fenced code block in the REFERENCED file must not
# be mistaken for a heading — the dangerous direction: if it were, a stale
# reference would silently pass instead of failing. This is the fence-aware
# regression guard for round-1's H1 widening (`^#\{1,4\}`), which — before
# headings_of was made fence-aware — could match a shell "# comment" inside a
# ```bash example exactly as if it were a real section title.
new_fixture
{
  printf '# Something else\n\n'
  printf '```bash\n# State file\necho hi\n```\n'
} > "$FIXTURE/rules/never-touch-production.mdc"
printf 'Resolve it per **State file** in `rules/never-touch-production.mdc`.\n' \
  > "$FIXTURE/skills/demo/SKILL.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "a fenced # comment in the referenced file does not satisfy a reference" \
  || fail "fenced # comment falsely satisfied a reference: rc=$RC out=$OUT"

# 17. The same fenced comment must not stop a REAL #-title in the same file
# from resolving — fence-awareness must exclude the fenced line without
# excluding the genuine heading that sits alongside it. As in #14, the extra
# `##` heading is what stops this from passing vacuously: it keeps the heading
# set non-empty, so an over-eager exclusion of the real `#` title surfaces as a
# failure here instead of as a silently skipped reference.
new_fixture
{
  printf '# State file\n\n## Body notes\n\n'
  printf '```bash\n# State file\necho hi\n```\n'
} > "$FIXTURE/rules/never-touch-production.mdc"
printf 'Resolve it per **State file** in `rules/never-touch-production.mdc`.\n' \
  > "$FIXTURE/skills/demo/SKILL.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "a real #-title still resolves alongside a fenced comment of the same text" \
  || fail "real #-title blocked by fenced comment: rc=$RC out=$OUT"

# 18. Adjacency premise: a bold token that is emphasis, not a section name, and
# is not written next to the path, must not be demanded as a heading. This is
# the shape that produced 28 false failures on this repo's own tree and forced
# the suppression markers that then switched off real checks. Written to a
# declared expected-zero member: with neither line's bold token associated to
# a path, this file's own checked-reference count is genuinely zero.
new_fixture
mkdir -p "$FIXTURE/skills/openspec-explore"
printf '## State file\n\nbody\n' > "$FIXTURE/rules/never-touch-production.mdc"
{
  printf '**Never** commit during apply. The contract lives in `rules/never-touch-production.mdc`.\n'
  printf 'Run it **after** step 2 — see the note in `rules/never-touch-production.mdc`.\n'
} > "$FIXTURE/skills/openspec-explore/SKILL.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "unassociated emphasis does not demand a heading" \
  || fail "unassociated emphasis: rc=$RC out=$OUT"

# 19. The adjacency premise must not become a way to smuggle a stale reference
# past the guard: a bold token written in a real reference shape is still
# checked, on the same line as unrelated emphasis.
new_fixture
printf '## Something else\n\nbody\n' > "$FIXTURE/rules/never-touch-production.mdc"
printf '**Never** skip this — see **State file** in `rules/never-touch-production.mdc`.\n' \
  > "$FIXTURE/skills/demo/SKILL.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "an associated token is still checked beside emphasis" \
  || fail "associated token beside emphasis: expected non-zero, out=$OUT"

# 20. A bold token belongs to the NEAREST path it is associated with. Here
# "Panel re-runs" is a live heading of the second file and must not also be
# demanded of the first, which it merely sits after.
new_fixture
printf '## Apps\n\nbody\n' > "$FIXTURE/rules/never-touch-production.mdc"
printf '## Panel re-runs\n\nbody\n' > "$FIXTURE/rules/context7.mdc"
printf 'the apps in `rules/never-touch-production.mdc` (see **Panel re-runs** in `rules/context7.mdc`)\n' \
  > "$FIXTURE/skills/demo/SKILL.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "a token is assigned to its nearest path only" \
  || fail "nearest-path assignment: rc=$RC out=$OUT"

# 21. Containment: repository Markdown is attacker-influenceable (any pull
# request can add a line), so a backticked path that normalizes outside the root
# is refused rather than opened — and the run FAILS, because this guard's
# contract is that a clean exit means every reference was checked.
#
# Containment is decided from the path SHAPE, before any existence test, so the
# two cases below must behave identically: if the verdict depended on whether
# the out-of-tree target happened to exist, the same repository would warn on a
# laptop and pass in CI, and the attacker-probe case would be the quiet one.
new_fixture
OUTSIDE_DIR="$(mktemp -d "${TMPDIR:-/tmp}/refs-outside.XXXXXX")"
printf '## Something else\n\nbody\n' > "$OUTSIDE_DIR/target.md"
printf 'see **State file** in `%s/target.md`\n' "$OUTSIDE_DIR" \
  > "$FIXTURE/skills/demo/SKILL.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "an absolute path outside the root fails the run" \
  || fail "absolute outside path (target exists): expected non-zero, out=$OUT"
case "$OUT" in
  *"resolves outside the repository root"*) pass "the refusal is reported" ;;
  *) fail "the refusal is not reported: out=$OUT" ;;
esac

new_fixture
NOWHERE="${TMPDIR:-/tmp}/refs-nonexistent-$$"
printf 'see **State file** in `%s/target.md`\n' "$NOWHERE" \
  > "$FIXTURE/skills/demo/SKILL.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "an escaping path whose target does not exist fails identically" \
  || fail "absolute outside path (target absent): expected non-zero, out=$OUT"
case "$OUT" in
  *"resolves outside the repository root"*) pass "the refusal is reported for a nonexistent target" ;;
  *) fail "the refusal is not reported for a nonexistent target: out=$OUT" ;;
esac

new_fixture
printf 'see **State file** in `../../refs-escape-nowhere/target.md`\n' \
  > "$FIXTURE/skills/demo/SKILL.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "a traversing relative path fails the run" \
  || fail "traversing relative path: expected non-zero, out=$OUT"

# 22. A root that cannot be scanned must never report a clean run. An empty
# override and a nonexistent one are both configuration errors, not clean trees.
set +e
OUT="$(CHECK_REFERENCES_ROOT="" "$GUARD" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 2 ] && pass "an empty CHECK_REFERENCES_ROOT exits 2" \
  || fail "empty root: rc=$RC out=$OUT"

set +e
OUT="$(CHECK_REFERENCES_ROOT="${TMPDIR:-/tmp}/refs-does-not-exist-$$" "$GUARD" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 2 ] && pass "a nonexistent CHECK_REFERENCES_ROOT exits 2" \
  || fail "nonexistent root: rc=$RC out=$OUT"

new_fixture
rm -rf "$FIXTURE/rules" "$FIXTURE/skills"
run_guard "$FIXTURE"
[ "$RC" -eq 2 ] && pass "a root with no Markdown exits 2 instead of reporting clean" \
  || fail "empty tree: rc=$RC out=$OUT"

# ===========================================================================
# 23-25. KAN-197 — per-file coverage. A guard exits 0 identically whether it
# checked every reference in a file or checked nothing there at all; these
# three cases are what makes the difference visible on a passing run.
# ===========================================================================

# 23. A scanned file whose references all resolve reports its count, and the
# verdict's own breakdown line names it.
new_fixture
printf '## State file\n\nbody\n' > "$FIXTURE/rules/never-touch-production.mdc"
printf 'Resolve it per **State file** in `rules/never-touch-production.mdc`.\n' \
  > "$FIXTURE/skills/demo/SKILL.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "coverage: a file with a live reference reports rc=0" \
  || fail "coverage positive: rc=$RC out=$OUT"
case "$OUT" in
  *"skills/demo/SKILL.md 1"*) pass "coverage: the breakdown names the file with its checked count" ;;
  *) fail "coverage positive: expected the breakdown to show skills/demo/SKILL.md 1, out=$OUT" ;;
esac

# 24. A scanned file contributing zero checked references, undeclared, fails
# by name — this is the KAN-73-shaped case: a rule computing nothing to check
# for one corpus member, on an otherwise clean tree.
new_fixture
printf 'Just prose. No cross-reference syntax anywhere in this file.\n' \
  > "$FIXTURE/skills/demo/SKILL.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "coverage: an undeclared zero-coverage file fails" \
  || fail "coverage undeclared zero: expected non-zero, out=$OUT"
case "$OUT" in
  *"skills/demo/SKILL.md"*"0 checked"*"not declared expected-zero"*) \
    pass "coverage: names the file, the zero count, and that it is undeclared" ;;
  *) fail "coverage undeclared zero: expected the file named as an undeclared zero, out=$OUT" ;;
esac

# 25. A member declared expected-zero in the guard's own source reports its
# zero without failing. Reuses "rules/never-touch-production.mdc", one of
# check-references.sh's own declared members (it genuinely carries no
# cross-reference syntax in the real repository — measured, not guessed),
# deliberately, the same reuse check-guard-symlinks's F13 case relies on.
new_fixture
printf 'Never access a production system directly. No file cross-references here.\n' \
  > "$FIXTURE/rules/never-touch-production.mdc"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "coverage: a declared expected-zero member passes" \
  || fail "coverage declared zero: rc=$RC out=$OUT"
case "$OUT" in
  *"rules/never-touch-production.mdc 0 (declared"*) pass "coverage: the breakdown marks the declared zero" ;;
  *) fail "coverage declared zero: expected a declared-zero breakdown entry, out=$OUT" ;;
esac

# ===========================================================================
# 26-27. kan-102-citations-resolve-to-installed-paths task 3 — a citation
# naming its root explicitly (`<agents repo>/…` or `<project>/…`) must
# resolve exactly as the same citation resolved before it carried a root.
# ===========================================================================

# 26. A citation prefixed with the literal `<agents repo>/` is stripped
# before resolution and still checked — proven by making its heading go
# stale. Without stripping, the prefixed path never resolves to a file and
# the reference is silently skipped instead of failing: a guard that stops
# checking without failing, which is the one failure a guard must never have.
new_fixture
printf '## Something else\n\nbody\n' > "$FIXTURE/agents-repo-fixture.md"
printf 'Resolve it per **Some section** in `<agents repo>/agents-repo-fixture.md`.\n' \
  > "$FIXTURE/skills/demo/SKILL.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "a prefixed citation is still checked" \
  || fail "agents-repo prefix: expected non-zero (stale heading), rc=$RC out=$OUT"
case "$OUT" in
  *"skills/demo/SKILL.md:1"*) pass "the prefixed reference is reported by file:line" ;;
  *) fail "agents-repo prefix: expected file:line report, out=$OUT" ;;
esac

# 27. A citation beginning `<project>/` names the target project, not this
# repository, and always resolves to nothing here. It must fall through the
# ordinary does-not-resolve path — neither checked nor read as an escape
# outside the repository root, which this guard treats as a hard failure
# (case 21 above).
#
# The fixture plants a COLLIDING file at the exact path stripping
# `<project>/` would produce — .myflow/project.md at the fixture root —
# carrying a heading that does NOT match the cited section. Without that
# collider, "correctly left alone" and "incorrectly stripped" are
# indistinguishable to this case: both leave the citation unresolved, RC=0,
# no escape message either way. With it, a regression that strips
# `<project>/` the way `<agents repo>/` is stripped makes the citation
# resolve to this real file, and the mismatched heading turns up as a
# reported stale reference (RC=1) instead of a silent pass. Written to a
# declared expected-zero member (skills/openspec-explore/SKILL.md), as case
# 7 above does for the same reason: with the reference correctly left
# unresolved, this file's own checked-reference count is genuinely zero.
#
# A `<project>/` prefix is a lexical marker, not `..` — no realistic
# mutation of this guard's `<project>/` handling ever introduces a genuine
# escape, so the absence of a containment-refusal message below is a plain
# sanity check on the current behavior, not a mutation-proven guarantee the
# way the heading check above is.
new_fixture
mkdir -p "$FIXTURE/skills/openspec-explore" "$FIXTURE/.myflow"
printf '## Something else\n\nbody\n' > "$FIXTURE/.myflow/project.md"
printf 'see **Whatever** in `<project>/.myflow/project.md`\n' \
  > "$FIXTURE/skills/openspec-explore/SKILL.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "a project-prefixed citation is neither resolved nor refused" \
  || fail "project prefix: rc=$RC out=$OUT"
case "$OUT" in
  *"resolves outside the repository root"*) fail "project prefix: misread as a containment escape: out=$OUT" ;;
  *) pass "project prefix: not read as a containment escape" ;;
esac

if [ "$FAILURES" -ne 0 ]; then
  printf '\n%d assertion(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf '\nAll check-references assertions passed\n'
