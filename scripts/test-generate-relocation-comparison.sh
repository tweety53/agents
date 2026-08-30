#!/usr/bin/env bash
# Assertion harness for generate-relocation-comparison.sh. Builds a
# throwaway git repo fixture under a sandboxed TMPDIR for every case, with
# a real merge-base commit and a real worktree-state commit — never
# touches this repository's own tree or its own spectre/changes/.
#
# Modeled on test-plan-dispatch-bundles.sh's fixture-driven pattern: a
# fixture directory per case via new_fixture, the generator invoked
# through a thin run_generator helper that captures RC and reads the real
# output file, and every case ending with an explicit pass/fail assertion
# against that file's actual content — never a hand-simulated expectation.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GENERATOR="$SCRIPT_DIR/generate-relocation-comparison.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'PASS: %s\n' "$1"; }

FIXTURES=()
cleanup() {
  [ "${#FIXTURES[@]}" -eq 0 ] && return 0
  for fixture in "${FIXTURES[@]}"; do
    rm -rf "$fixture"
  done
}
trap cleanup EXIT

# new_fixture — a fresh git repo under TMPDIR, with user identity set so
# commits succeed unattended.
new_fixture() {
  FIXTURE="$(mktemp -d "${TMPDIR:-/tmp}/generate-relocation-comparison-test.XXXXXX")"
  FIXTURES+=("$FIXTURE")
  git -C "$FIXTURE" init -q
  git -C "$FIXTURE" config user.email test@example.com
  git -C "$FIXTURE" config user.name test
}

# run_generator <worktree> <changeRoot> <mergeBase> -> sets RC, ERR, and
# OUTPUT_PATH (the file the generator is expected to write, whether or not
# it actually did).
run_generator() {
  set +e
  ERR="$("$GENERATOR" "$1" "$2" "$3" 2>&1 1>/dev/null)"
  RC=$?
  set -e
  OUTPUT_PATH="$1/.superpowers/sdd/relocation-comparison.md"
}

write_tasks_md() {
  # write_tasks_md <path> <relocation-value> <files...>
  local path="$1" relocation="$2"
  shift 2
  {
    printf '# fixture plan\n\n'
    printf '> **Relocation:** %s\n\n' "$relocation"
    printf -- '- [ ] 1. Only task\n\n'
    printf '**Files:**\n'
    for f in "$@"; do
      printf -- '- Modify: `%s`\n' "$f"
    done
    printf '\n'
  } > "$path"
}

# ===========================================================================
# Case (a): a paragraph moved verbatim between two files -> classified
# "moved".
# ===========================================================================
new_fixture
{
  printf '## Section One\n\n'
  printf 'This is the paragraph that will move verbatim to another file\nacross this change.\n\n'
} > "$FIXTURE/source.md"
printf '## Section Two\n\nUnrelated content that stays put.\n' > "$FIXTURE/dest.md"
git -C "$FIXTURE" add source.md dest.md
git -C "$FIXTURE" commit -qm base
BASE_A="$(git -C "$FIXTURE" rev-parse HEAD)"

printf '## Section One\n\nUnrelated content that stays put.\n' > "$FIXTURE/source.md"
{
  printf '## Section Two\n\n'
  printf 'This is the paragraph that will move verbatim to another file\nacross this change.\n\n'
} > "$FIXTURE/dest.md"
git -C "$FIXTURE" add source.md dest.md
git -C "$FIXTURE" commit -qm move

write_tasks_md "$FIXTURE/tasks.md" "yes" "source.md" "dest.md"
run_generator "$FIXTURE" "$FIXTURE" "$BASE_A"
if [ "$RC" -ne 0 ]; then
  fail "case a: generator exited $RC: $ERR"
elif [ ! -f "$OUTPUT_PATH" ]; then
  fail "case a: no output file written"
elif ! grep -q 'moved' "$OUTPUT_PATH"; then
  fail "case a: no 'moved' row in output:"$'\n'"$(cat "$OUTPUT_PATH")"
elif ! grep -q 'This is the paragraph' "$OUTPUT_PATH"; then
  fail "case a: moved passage text not found in output:"$'\n'"$(cat "$OUTPUT_PATH")"
else
  pass "case a: verbatim move classified moved"
fi

# ===========================================================================
# Case (b): a paragraph moved with a repointed citation (backticked path +
# bold section token changed) -> classified "repointed".
# ===========================================================================
new_fixture
BEFORE_CITATION='See the rule in `old.md`, **Old Section**, for the full rule.'
AFTER_CITATION='See the rule in `new.md`, **New Section**, for the full rule.'
printf '## Alpha\n\n%s\n\n' "$BEFORE_CITATION" > "$FIXTURE/alpha.md"
printf '## Beta\n\nUnrelated content.\n' > "$FIXTURE/beta.md"
git -C "$FIXTURE" add alpha.md beta.md
git -C "$FIXTURE" commit -qm base
BASE_B="$(git -C "$FIXTURE" rev-parse HEAD)"

printf '## Alpha\n\nUnrelated content.\n' > "$FIXTURE/alpha.md"
printf '## Beta\n\n%s\n\n' "$AFTER_CITATION" > "$FIXTURE/beta.md"
git -C "$FIXTURE" add alpha.md beta.md
git -C "$FIXTURE" commit -qm repoint

write_tasks_md "$FIXTURE/tasks.md" "yes" "alpha.md" "beta.md"
run_generator "$FIXTURE" "$FIXTURE" "$BASE_B"
if [ "$RC" -ne 0 ]; then
  fail "case b: generator exited $RC: $ERR"
elif [ ! -f "$OUTPUT_PATH" ]; then
  fail "case b: no output file written"
elif ! grep -q 'repointed' "$OUTPUT_PATH"; then
  fail "case b: no 'repointed' row in output:"$'\n'"$(cat "$OUTPUT_PATH")"
else
  pass "case b: repointed citation classified repointed"
fi

# ===========================================================================
# Case (c): a wholly new paragraph -> classified "added".
# ===========================================================================
new_fixture
printf '## Gamma\n\nOriginal content only.\n' > "$FIXTURE/gamma.md"
git -C "$FIXTURE" add gamma.md
git -C "$FIXTURE" commit -qm base
BASE_C="$(git -C "$FIXTURE" rev-parse HEAD)"

{
  printf '## Gamma\n\nOriginal content only.\n\n'
  printf 'A brand new paragraph that never existed before this change landed.\n\n'
} > "$FIXTURE/gamma.md"
git -C "$FIXTURE" add gamma.md
git -C "$FIXTURE" commit -qm add

write_tasks_md "$FIXTURE/tasks.md" "yes" "gamma.md"
run_generator "$FIXTURE" "$FIXTURE" "$BASE_C"
if [ "$RC" -ne 0 ]; then
  fail "case c: generator exited $RC: $ERR"
elif [ ! -f "$OUTPUT_PATH" ]; then
  fail "case c: no output file written"
elif ! grep -q 'added' "$OUTPUT_PATH"; then
  fail "case c: no 'added' row in output:"$'\n'"$(cat "$OUTPUT_PATH")"
elif ! grep -q 'A brand new paragraph' "$OUTPUT_PATH"; then
  fail "case c: added passage text not found in output:"$'\n'"$(cat "$OUTPUT_PATH")"
else
  pass "case c: new paragraph classified added"
fi

# ===========================================================================
# Case (d): a paragraph deleted outright (not relocated) -> classified
# "removed".
# ===========================================================================
new_fixture
{
  printf '## Delta\n\nContent that stays.\n\n'
  printf 'This paragraph is simply deleted, not moved anywhere else.\n\n'
} > "$FIXTURE/delta.md"
git -C "$FIXTURE" add delta.md
git -C "$FIXTURE" commit -qm base
BASE_D="$(git -C "$FIXTURE" rev-parse HEAD)"

printf '## Delta\n\nContent that stays.\n' > "$FIXTURE/delta.md"
git -C "$FIXTURE" add delta.md
git -C "$FIXTURE" commit -qm remove

write_tasks_md "$FIXTURE/tasks.md" "yes" "delta.md"
run_generator "$FIXTURE" "$FIXTURE" "$BASE_D"
if [ "$RC" -ne 0 ]; then
  fail "case d: generator exited $RC: $ERR"
elif [ ! -f "$OUTPUT_PATH" ]; then
  fail "case d: no output file written"
elif ! grep -q 'removed' "$OUTPUT_PATH"; then
  fail "case d: no 'removed' row in output:"$'\n'"$(cat "$OUTPUT_PATH")"
elif ! grep -q 'This paragraph is simply deleted' "$OUTPUT_PATH"; then
  fail "case d: removed passage text not found in output:"$'\n'"$(cat "$OUTPUT_PATH")"
else
  pass "case d: deleted paragraph classified removed"
fi

# ===========================================================================
# Case (e): a plan declaring **Relocation:** no -> exits 0, writes no
# output file.
# ===========================================================================
new_fixture
printf '## Epsilon\n\nSome content.\n' > "$FIXTURE/epsilon.md"
git -C "$FIXTURE" add epsilon.md
git -C "$FIXTURE" commit -qm base
BASE_E="$(git -C "$FIXTURE" rev-parse HEAD)"
printf '## Epsilon\n\nSome different content.\n' > "$FIXTURE/epsilon.md"
git -C "$FIXTURE" add epsilon.md
git -C "$FIXTURE" commit -qm change

write_tasks_md "$FIXTURE/tasks.md" "no" "epsilon.md"
run_generator "$FIXTURE" "$FIXTURE" "$BASE_E"
if [ "$RC" -ne 0 ]; then
  fail "case e: generator exited $RC: $ERR"
elif [ -f "$OUTPUT_PATH" ]; then
  fail "case e: output file written despite **Relocation:** no"
else
  pass "case e: **Relocation:** no is a no-op, no output written"
fi

# ===========================================================================
# Case (f): an unresolvable merge-base ref -> exits 2.
# ===========================================================================
new_fixture
printf '## Zeta\n\nSome content.\n' > "$FIXTURE/zeta.md"
git -C "$FIXTURE" add zeta.md
git -C "$FIXTURE" commit -qm base

write_tasks_md "$FIXTURE/tasks.md" "yes" "zeta.md"
run_generator "$FIXTURE" "$FIXTURE" "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
if [ "$RC" -ne 2 ]; then
  fail "case f: expected exit 2 for unresolvable merge-base, got $RC: $ERR"
elif ! echo "$ERR" | grep -q "does not resolve to a commit"; then
  # Isolates the merge-base-resolves guard specifically (F10b): exit 2
  # alone is not proof this guard fired — a downstream git-show failure
  # against the same bogus ref would also exit 2 for an unrelated reason.
  # Asserting the guard's own diagnostic text is what makes this case
  # actually exercise `_merge_base_resolves`, not just its exit code.
  fail "case f: exit 2 but not via the merge-base-resolves guard's own diagnostic: $ERR"
else
  pass "case f: unresolvable merge-base exits 2 via the merge-base-resolves guard"
fi

# ===========================================================================
# Case (g): a scoped file present at merge-base but missing from the
# worktree on disk -> exits 2.
# ===========================================================================
new_fixture
printf '## Eta\n\nSome content.\n' > "$FIXTURE/eta.md"
git -C "$FIXTURE" add eta.md
git -C "$FIXTURE" commit -qm base
BASE_G="$(git -C "$FIXTURE" rev-parse HEAD)"
rm "$FIXTURE/eta.md"

write_tasks_md "$FIXTURE/tasks.md" "yes" "eta.md"
run_generator "$FIXTURE" "$FIXTURE" "$BASE_G"
if [ "$RC" -eq 2 ]; then
  pass "case g: scoped file missing from worktree exits 2"
else
  fail "case g: expected exit 2 for missing worktree file, got $RC: $ERR"
fi

# ===========================================================================
# Case (h): an unreadable/missing tasks.md -> exits 2.
# ===========================================================================
new_fixture
printf '## Theta\n\nSome content.\n' > "$FIXTURE/theta.md"
git -C "$FIXTURE" add theta.md
git -C "$FIXTURE" commit -qm base
BASE_H="$(git -C "$FIXTURE" rev-parse HEAD)"

run_generator "$FIXTURE" "$FIXTURE" "$BASE_H"
if [ "$RC" -eq 2 ]; then
  pass "case h: missing tasks.md exits 2"
else
  fail "case h: expected exit 2 for missing tasks.md, got $RC: $ERR"
fi

# ===========================================================================
# Case (i): a plan with no **Relocation:** header line at all (absent, not
# `no`, not malformed) -> no-op, exit 0, writes nothing.
# ===========================================================================
new_fixture
printf '## Iota\n\nSome content.\n' > "$FIXTURE/iota.md"
git -C "$FIXTURE" add iota.md
git -C "$FIXTURE" commit -qm base
BASE_I="$(git -C "$FIXTURE" rev-parse HEAD)"
printf '## Iota\n\nSome different content.\n' > "$FIXTURE/iota.md"
git -C "$FIXTURE" add iota.md
git -C "$FIXTURE" commit -qm change

{
  printf '# fixture plan\n\n'
  printf -- '- [ ] 1. Only task\n\n'
  printf '**Files:**\n'
  printf -- '- Modify: `iota.md`\n\n'
} > "$FIXTURE/tasks.md"
run_generator "$FIXTURE" "$FIXTURE" "$BASE_I"
if [ "$RC" -ne 0 ]; then
  fail "case i: generator exited $RC: $ERR"
elif [ -f "$OUTPUT_PATH" ]; then
  fail "case i: output file written despite absent **Relocation:** header"
else
  pass "case i: absent **Relocation:** header is a no-op, no output written"
fi

# ===========================================================================
# Case (j): an in-scope passage byte-identical at the same (file, heading)
# location in both merge-base and worktree -> not a row at all.
# ===========================================================================
new_fixture
{
  printf '## Kappa\n\n'
  printf 'This passage never changes at all across the whole revision.\n\n'
  printf 'This passage does change between the two revisions.\n\n'
} > "$FIXTURE/kappa.md"
git -C "$FIXTURE" add kappa.md
git -C "$FIXTURE" commit -qm base
BASE_J="$(git -C "$FIXTURE" rev-parse HEAD)"

{
  printf '## Kappa\n\n'
  printf 'This passage never changes at all across the whole revision.\n\n'
  printf 'This passage has now changed between the two revisions.\n\n'
} > "$FIXTURE/kappa.md"
git -C "$FIXTURE" add kappa.md
git -C "$FIXTURE" commit -qm change

write_tasks_md "$FIXTURE/tasks.md" "yes" "kappa.md"
run_generator "$FIXTURE" "$FIXTURE" "$BASE_J"
if [ "$RC" -ne 0 ]; then
  fail "case j: generator exited $RC: $ERR"
elif [ ! -f "$OUTPUT_PATH" ]; then
  fail "case j: no output file written"
elif grep -q 'This passage never changes' "$OUTPUT_PATH"; then
  fail "case j: unchanged passage appeared as a row:"$'\n'"$(cat "$OUTPUT_PATH")"
else
  pass "case j: byte-identical same-location passage produces no row"
fi

# ===========================================================================
# Case (k): a worktree file that is not valid UTF-8 -> exits 2 (F1
# regression — round 1's fix was only verified by manual ad hoc
# reproduction, never by an automated case).
# ===========================================================================
new_fixture
printf '## Kappa2\n\nSome content.\n' > "$FIXTURE/mu.md"
git -C "$FIXTURE" add mu.md
git -C "$FIXTURE" commit -qm base
BASE_K="$(git -C "$FIXTURE" rev-parse HEAD)"
# Overwrite the worktree copy with invalid UTF-8 bytes; never committed —
# the generator reads the worktree copy straight off disk.
printf '## Kappa2\n\nSome content \xff\xfe binary garbage.\n' > "$FIXTURE/mu.md"

write_tasks_md "$FIXTURE/tasks.md" "yes" "mu.md"
run_generator "$FIXTURE" "$FIXTURE" "$BASE_K"
if [ "$RC" -eq 2 ]; then
  pass "case k: non-UTF-8 worktree file exits 2 (F1 regression)"
else
  fail "case k: expected exit 2 for non-UTF-8 worktree file, got $RC: $ERR"
fi

# ===========================================================================
# Case (l): `git show` against the merge-base fails for a reason OTHER than
# "path absent at that ref" (a corrupt loose object) -> exits 2, not
# misread as an ordinary "added" file (F2 regression, same rationale as
# case k).
# ===========================================================================
new_fixture
printf '## Nu\n\nSome content.\n' > "$FIXTURE/nu.md"
git -C "$FIXTURE" add nu.md
git -C "$FIXTURE" commit -qm base
BASE_L="$(git -C "$FIXTURE" rev-parse HEAD)"
BLOB_SHA_L="$(git -C "$FIXTURE" rev-parse "$BASE_L:nu.md")"
OBJFILE_L="$FIXTURE/.git/objects/${BLOB_SHA_L:0:2}/${BLOB_SHA_L:2}"
chmod u+w "$OBJFILE_L"
printf 'corrupt' > "$OBJFILE_L"

write_tasks_md "$FIXTURE/tasks.md" "yes" "nu.md"
run_generator "$FIXTURE" "$FIXTURE" "$BASE_L"
if [ "$RC" -ne 2 ]; then
  fail "case l: expected exit 2 for corrupt merge-base blob, got $RC: $ERR"
elif echo "$ERR" | grep -q "does not exist in"; then
  fail "case l: corrupt blob misread as an ordinary missing path: $ERR"
else
  pass "case l: corrupt merge-base blob exits 2, not misread as missing path (F2 regression)"
fi

# ===========================================================================
# Case (m): a bold span far from any backtick-quoted path changes on its
# own -> classified removed+added, never repointed (F3 regression: the
# adjacency-narrowed CITATION_RE must not sweep in a bold span that is not
# genuinely beside a citation).
# ===========================================================================
new_fixture
BEFORE_FAR='See `note.md` for details. Also **Alpha** far from the citation and not adjacent to it at all.'
AFTER_FAR='See `note.md` for details. Also **Beta** far from the citation and not adjacent to it at all.'
printf '## Xi\n\n%s\n\n' "$BEFORE_FAR" > "$FIXTURE/xi.md"
git -C "$FIXTURE" add xi.md
git -C "$FIXTURE" commit -qm base
BASE_M="$(git -C "$FIXTURE" rev-parse HEAD)"
printf '## Xi\n\n%s\n\n' "$AFTER_FAR" > "$FIXTURE/xi.md"
git -C "$FIXTURE" add xi.md
git -C "$FIXTURE" commit -qm change

write_tasks_md "$FIXTURE/tasks.md" "yes" "xi.md"
run_generator "$FIXTURE" "$FIXTURE" "$BASE_M"
if [ "$RC" -ne 0 ]; then
  fail "case m: generator exited $RC: $ERR"
elif [ ! -f "$OUTPUT_PATH" ]; then
  fail "case m: no output file written"
elif grep -q 'repointed' "$OUTPUT_PATH"; then
  fail "case m: far bold-span edit misclassified repointed (F3 regression):"$'\n'"$(cat "$OUTPUT_PATH")"
elif ! grep -q 'removed' "$OUTPUT_PATH" || ! grep -q 'added' "$OUTPUT_PATH"; then
  fail "case m: expected removed+added rows for an unrelated far bold-span edit:"$'\n'"$(cat "$OUTPUT_PATH")"
else
  pass "case m: far bold-span edit not misclassified repointed (F3 regression guard)"
fi

# ===========================================================================
# Case (n): a passage containing a literal `|` moves verbatim -> the output
# table cell escapes it as `\|` (F5 regression: an unescaped pipe would be
# misread as a column separator).
# ===========================================================================
new_fixture
PIPE_TEXT='Value | here is piped content that moves across files verbatim.'
printf '## Omicron\n\n%s\n\n' "$PIPE_TEXT" > "$FIXTURE/pi_src.md"
printf '## Rho\n\nUnrelated content that stays put.\n' > "$FIXTURE/pi_dst.md"
git -C "$FIXTURE" add pi_src.md pi_dst.md
git -C "$FIXTURE" commit -qm base
BASE_N="$(git -C "$FIXTURE" rev-parse HEAD)"

printf '## Omicron\n\nUnrelated content that stays put.\n' > "$FIXTURE/pi_src.md"
printf '## Rho\n\n%s\n\n' "$PIPE_TEXT" > "$FIXTURE/pi_dst.md"
git -C "$FIXTURE" add pi_src.md pi_dst.md
git -C "$FIXTURE" commit -qm move

write_tasks_md "$FIXTURE/tasks.md" "yes" "pi_src.md" "pi_dst.md"
run_generator "$FIXTURE" "$FIXTURE" "$BASE_N"
if [ "$RC" -ne 0 ]; then
  fail "case n: generator exited $RC: $ERR"
elif [ ! -f "$OUTPUT_PATH" ]; then
  fail "case n: no output file written"
elif ! grep -qF 'Value \| here' "$OUTPUT_PATH"; then
  fail "case n: literal pipe not escaped in output (F5 regression):"$'\n'"$(cat "$OUTPUT_PATH")"
else
  pass "case n: literal pipe escaped in moved passage cell (F5 regression guard)"
fi

# ===========================================================================
# Case (o): a paragraph moved with a repointed citation where the backtick
# path and the bold section token are separated by nothing but whitespace —
# the corpus's own no-separator citation style (e.g. `` `foo.md` **Bar**
# `` ), not the comma-and-space style case (b) already covers -> classified
# "repointed", not removed+added (F7 regression: the pre-fix adjacency
# pattern only matched an optional comma-and-space, so this exact corpus
# shape was misclassified).
# ===========================================================================
new_fixture
BEFORE_CITATION_O='See `old.md` **Old Section** for the full rule.'
AFTER_CITATION_O='See `new.md` **New Section** for the full rule.'
printf '## AlphaO\n\n%s\n\n' "$BEFORE_CITATION_O" > "$FIXTURE/alpha_o.md"
printf '## BetaO\n\nUnrelated content.\n' > "$FIXTURE/beta_o.md"
git -C "$FIXTURE" add alpha_o.md beta_o.md
git -C "$FIXTURE" commit -qm base
BASE_O="$(git -C "$FIXTURE" rev-parse HEAD)"

printf '## AlphaO\n\nUnrelated content.\n' > "$FIXTURE/alpha_o.md"
printf '## BetaO\n\n%s\n\n' "$AFTER_CITATION_O" > "$FIXTURE/beta_o.md"
git -C "$FIXTURE" add alpha_o.md beta_o.md
git -C "$FIXTURE" commit -qm repoint

write_tasks_md "$FIXTURE/tasks.md" "yes" "alpha_o.md" "beta_o.md"
run_generator "$FIXTURE" "$FIXTURE" "$BASE_O"
if [ "$RC" -ne 0 ]; then
  fail "case o: generator exited $RC: $ERR"
elif [ ! -f "$OUTPUT_PATH" ]; then
  fail "case o: no output file written"
elif ! grep -q 'repointed' "$OUTPUT_PATH"; then
  fail "case o: no-separator citation repoint misclassified (F7 regression):"$'\n'"$(cat "$OUTPUT_PATH")"
else
  pass "case o: no-separator citation repoint classified repointed (F7 regression guard)"
fi

# ===========================================================================
# Case (p): a moved bullet-list passage (`- some bullet text`) -> classified
# "moved" (F15).
# ===========================================================================
new_fixture
{
  printf '## SectionBulletA\n\n'
  printf -- '- This bullet line will move verbatim to another file.\n\n'
} > "$FIXTURE/bullet_src.md"
printf '## SectionBulletB\n\nUnrelated content that stays put.\n' > "$FIXTURE/bullet_dst.md"
git -C "$FIXTURE" add bullet_src.md bullet_dst.md
git -C "$FIXTURE" commit -qm base
BASE_P="$(git -C "$FIXTURE" rev-parse HEAD)"

printf '## SectionBulletA\n\nUnrelated content that stays put.\n' > "$FIXTURE/bullet_src.md"
{
  printf '## SectionBulletB\n\n'
  printf -- '- This bullet line will move verbatim to another file.\n\n'
} > "$FIXTURE/bullet_dst.md"
git -C "$FIXTURE" add bullet_src.md bullet_dst.md
git -C "$FIXTURE" commit -qm move

write_tasks_md "$FIXTURE/tasks.md" "yes" "bullet_src.md" "bullet_dst.md"
run_generator "$FIXTURE" "$FIXTURE" "$BASE_P"
if [ "$RC" -ne 0 ]; then
  fail "case p: generator exited $RC: $ERR"
elif [ ! -f "$OUTPUT_PATH" ]; then
  fail "case p: no output file written"
elif ! grep -q 'moved' "$OUTPUT_PATH"; then
  fail "case p: no 'moved' row in output:"$'\n'"$(cat "$OUTPUT_PATH")"
elif ! grep -q 'This bullet line will move' "$OUTPUT_PATH"; then
  fail "case p: moved bullet passage text not found in output:"$'\n'"$(cat "$OUTPUT_PATH")"
else
  pass "case p: moved bullet-list passage classified moved"
fi

# ===========================================================================
# Case (q): a moved Markdown table row (`| ... |`) -> classified "moved"
# (F15).
# ===========================================================================
new_fixture
{
  printf '## SectionTableA\n\n'
  printf '| MovedTableCell | second column value here |\n\n'
} > "$FIXTURE/table_src.md"
printf '## SectionTableB\n\nUnrelated content that stays put.\n' > "$FIXTURE/table_dst.md"
git -C "$FIXTURE" add table_src.md table_dst.md
git -C "$FIXTURE" commit -qm base
BASE_Q="$(git -C "$FIXTURE" rev-parse HEAD)"

printf '## SectionTableA\n\nUnrelated content that stays put.\n' > "$FIXTURE/table_src.md"
{
  printf '## SectionTableB\n\n'
  printf '| MovedTableCell | second column value here |\n\n'
} > "$FIXTURE/table_dst.md"
git -C "$FIXTURE" add table_src.md table_dst.md
git -C "$FIXTURE" commit -qm move

write_tasks_md "$FIXTURE/tasks.md" "yes" "table_src.md" "table_dst.md"
run_generator "$FIXTURE" "$FIXTURE" "$BASE_Q"
if [ "$RC" -ne 0 ]; then
  fail "case q: generator exited $RC: $ERR"
elif [ ! -f "$OUTPUT_PATH" ]; then
  fail "case q: no output file written"
elif ! grep -q 'moved' "$OUTPUT_PATH"; then
  fail "case q: no 'moved' row in output:"$'\n'"$(cat "$OUTPUT_PATH")"
elif ! grep -q 'MovedTableCell' "$OUTPUT_PATH"; then
  fail "case q: moved table-row passage text not found in output:"$'\n'"$(cat "$OUTPUT_PATH")"
else
  pass "case q: moved table-row passage classified moved"
fi

# ===========================================================================
# Case (r): `_format_location`'s `§ <heading>` suffix appears in the output
# for a passage under a heading (F15).
# ===========================================================================
new_fixture
{
  printf '## DistinctiveHeadingUpsilon\n\n'
  printf 'This paragraph moves under a heading so its location gets a section suffix.\n\n'
} > "$FIXTURE/heading_src.md"
printf '## HeadingDest\n\nUnrelated content that stays put.\n' > "$FIXTURE/heading_dst.md"
git -C "$FIXTURE" add heading_src.md heading_dst.md
git -C "$FIXTURE" commit -qm base
BASE_R="$(git -C "$FIXTURE" rev-parse HEAD)"

printf '## DistinctiveHeadingUpsilon\n\nUnrelated content that stays put.\n' > "$FIXTURE/heading_src.md"
{
  printf '## HeadingDest\n\n'
  printf 'This paragraph moves under a heading so its location gets a section suffix.\n\n'
} > "$FIXTURE/heading_dst.md"
git -C "$FIXTURE" add heading_src.md heading_dst.md
git -C "$FIXTURE" commit -qm move

write_tasks_md "$FIXTURE/tasks.md" "yes" "heading_src.md" "heading_dst.md"
run_generator "$FIXTURE" "$FIXTURE" "$BASE_R"
if [ "$RC" -ne 0 ]; then
  fail "case r: generator exited $RC: $ERR"
elif [ ! -f "$OUTPUT_PATH" ]; then
  fail "case r: no output file written"
elif ! grep -qF '§ HeadingDest' "$OUTPUT_PATH"; then
  fail "case r: no '§ <heading>' location suffix in output:"$'\n'"$(cat "$OUTPUT_PATH")"
else
  pass "case r: _format_location's § heading suffix appears in output"
fi

# ===========================================================================
# Case (s): `scoped_files` de-duplicates a file named in two different
# tasks' **Files:** fields (same file, two tasks) rather than double-
# counting its passages (F15).
# ===========================================================================
new_fixture
{
  printf '## SectionDup\n\n'
  printf 'This paragraph exists once and must not be reported twice.\n\n'
} > "$FIXTURE/dup_src.md"
printf '## SectionDupDest\n\nUnrelated content that stays put.\n' > "$FIXTURE/dup_dst.md"
git -C "$FIXTURE" add dup_src.md dup_dst.md
git -C "$FIXTURE" commit -qm base
BASE_S="$(git -C "$FIXTURE" rev-parse HEAD)"

printf '## SectionDup\n\nUnrelated content that stays put.\n' > "$FIXTURE/dup_src.md"
{
  printf '## SectionDupDest\n\n'
  printf 'This paragraph exists once and must not be reported twice.\n\n'
} > "$FIXTURE/dup_dst.md"
git -C "$FIXTURE" add dup_src.md dup_dst.md
git -C "$FIXTURE" commit -qm move

{
  printf '# fixture plan\n\n'
  printf '> **Relocation:** yes\n\n'
  printf -- '- [ ] 1. First task\n\n'
  printf '**Files:**\n'
  printf -- '- Modify: `dup_src.md`\n'
  printf -- '- Modify: `dup_dst.md`\n\n'
  printf -- '- [ ] 2. Second task, naming dup_src.md again\n\n'
  printf '**Files:**\n'
  printf -- '- Modify: `dup_src.md`\n\n'
} > "$FIXTURE/tasks.md"
run_generator "$FIXTURE" "$FIXTURE" "$BASE_S"
MOVED_ROW_COUNT="$(grep -c 'This paragraph exists once' "$OUTPUT_PATH" 2>/dev/null || true)"
if [ "$RC" -ne 0 ]; then
  fail "case s: generator exited $RC: $ERR"
elif [ ! -f "$OUTPUT_PATH" ]; then
  fail "case s: no output file written"
elif [ "$MOVED_ROW_COUNT" -ne 1 ]; then
  fail "case s: expected exactly 1 row for a file scoped by two tasks, got $MOVED_ROW_COUNT:"$'\n'"$(cat "$OUTPUT_PATH")"
else
  pass "case s: scoped_files de-duplicates a file named by two tasks"
fi

# ===========================================================================
# Case (t): a citation repoint where the separator style itself changes
# between before and after (comma-and-space -> no separator) -> still
# classified "repointed" (F15 / F11 harder case).
# ===========================================================================
new_fixture
BEFORE_CITATION_T='See the rule in `old.md`, **Old Section**, for the full rule.'
AFTER_CITATION_T='See the rule in `new.md` **New Section**, for the full rule.'
printf '## AlphaT\n\n%s\n\n' "$BEFORE_CITATION_T" > "$FIXTURE/alpha_t.md"
printf '## BetaT\n\nUnrelated content.\n' > "$FIXTURE/beta_t.md"
git -C "$FIXTURE" add alpha_t.md beta_t.md
git -C "$FIXTURE" commit -qm base
BASE_T="$(git -C "$FIXTURE" rev-parse HEAD)"

printf '## AlphaT\n\nUnrelated content.\n' > "$FIXTURE/alpha_t.md"
printf '## BetaT\n\n%s\n\n' "$AFTER_CITATION_T" > "$FIXTURE/beta_t.md"
git -C "$FIXTURE" add alpha_t.md beta_t.md
git -C "$FIXTURE" commit -qm repoint

write_tasks_md "$FIXTURE/tasks.md" "yes" "alpha_t.md" "beta_t.md"
run_generator "$FIXTURE" "$FIXTURE" "$BASE_T"
if [ "$RC" -ne 0 ]; then
  fail "case t: generator exited $RC: $ERR"
elif [ ! -f "$OUTPUT_PATH" ]; then
  fail "case t: no output file written"
elif ! grep -q 'repointed' "$OUTPUT_PATH"; then
  fail "case t: separator-style change misclassified (F11 harder case):"$'\n'"$(cat "$OUTPUT_PATH")"
else
  pass "case t: citation repoint with a changed separator style still classified repointed"
fi

# ===========================================================================
# Case (u): a scoped file that exists on disk but not at the merge-base ref
# (the ordinary shape of a file a plan creates fresh) -> produces an
# "added" row, not exit 2 (F14 regression).
# ===========================================================================
new_fixture
printf '## Phi\n\nSome unrelated content at the merge base.\n' > "$FIXTURE/phi.md"
git -C "$FIXTURE" add phi.md
git -C "$FIXTURE" commit -qm base
BASE_U="$(git -C "$FIXTURE" rev-parse HEAD)"
# freshly created file: never existed at BASE_U, present only in the
# worktree on disk (uncommitted) -> `git show <BASE_U>:new_file.md` would
# fail with git's "exists on disk, but not in" phrasing, not the "does not
# exist in" phrasing F14's old PATH_ABSENT_AT_REF_RE only matched.
printf '## Chi\n\nA brand new file this plan creates fresh, never at the merge base.\n' > "$FIXTURE/new_file.md"

write_tasks_md "$FIXTURE/tasks.md" "yes" "phi.md" "new_file.md"
run_generator "$FIXTURE" "$FIXTURE" "$BASE_U"
if [ "$RC" -ne 0 ]; then
  fail "case u: generator exited $RC: $ERR (F14 regression)"
elif [ ! -f "$OUTPUT_PATH" ]; then
  fail "case u: no output file written"
elif ! grep -q 'added' "$OUTPUT_PATH"; then
  fail "case u: no 'added' row for a file fresh on disk but absent at merge-base (F14 regression):"$'\n'"$(cat "$OUTPUT_PATH")"
elif ! grep -q 'A brand new file this plan creates fresh' "$OUTPUT_PATH"; then
  fail "case u: added passage text not found in output (F14 regression):"$'\n'"$(cat "$OUTPUT_PATH")"
else
  pass "case u: file on disk but absent at merge-base classified added, not exit 2 (F14 regression guard)"
fi

# ===========================================================================
# Case (v): the `` `red` ``/`` `**Squash-with:**` `` cross-span splicing
# shape -> a content edit between two unrelated backtick spans must
# classify as removed+added, not repointed (F13 regression). The edit sits
# in the plain text between the two spans — the exact text a single
# combined regex's separator-matching silently absorbed and hid, letting
# an unrelated content change pass as a permitted citation repoint.
# ===========================================================================
new_fixture
BEFORE_SPLICE='The color `red` describes it, similar to `**Squash-with:**` field, unrelated to any move.'
AFTER_SPLICE='The color `red` totally unrelated content edit here to `**Squash-with:**` field, unrelated to any move.'
printf '## Sigma\n\n%s\n\n' "$BEFORE_SPLICE" > "$FIXTURE/sigma.md"
git -C "$FIXTURE" add sigma.md
git -C "$FIXTURE" commit -qm base
BASE_V="$(git -C "$FIXTURE" rev-parse HEAD)"
printf '## Sigma\n\n%s\n\n' "$AFTER_SPLICE" > "$FIXTURE/sigma.md"
git -C "$FIXTURE" add sigma.md
git -C "$FIXTURE" commit -qm change

write_tasks_md "$FIXTURE/tasks.md" "yes" "sigma.md"
run_generator "$FIXTURE" "$FIXTURE" "$BASE_V"
if [ "$RC" -ne 0 ]; then
  fail "case v: generator exited $RC: $ERR (F13 regression)"
elif [ ! -f "$OUTPUT_PATH" ]; then
  fail "case v: no output file written"
elif grep -q 'repointed' "$OUTPUT_PATH"; then
  fail "case v: cross-span splice misclassified repointed (F13 regression):"$'\n'"$(cat "$OUTPUT_PATH")"
elif ! grep -q 'removed' "$OUTPUT_PATH" || ! grep -q 'added' "$OUTPUT_PATH"; then
  fail "case v: expected removed+added rows for the cross-span splice content edit:"$'\n'"$(cat "$OUTPUT_PATH")"
else
  pass "case v: cross-span splice content edit not misclassified repointed (F13 regression guard)"
fi

# ===========================================================================
# Case (w): a repointed citation that ALSO drops a stale trailing
# `above`/`below` word -> classified "repointed", not removed+added (F16
# regression: the before text's own trailing punctuation was stripped
# together with the dropped "above", while the after text (which never had
# the word) kept its own trailing period, so the two signatures diverged by
# that one character alone).
# ===========================================================================
new_fixture
BEFORE_ABOVE_W='`old-file.md` **Old Heading** is the place to look above.'
AFTER_ABOVE_W='`new-file.md` **New Heading** is the place to look.'
printf '## AlphaW\n\n%s\n\n' "$BEFORE_ABOVE_W" > "$FIXTURE/alpha_w.md"
printf '## BetaW\n\nUnrelated content.\n' > "$FIXTURE/beta_w.md"
git -C "$FIXTURE" add alpha_w.md beta_w.md
git -C "$FIXTURE" commit -qm base
BASE_W="$(git -C "$FIXTURE" rev-parse HEAD)"

printf '## AlphaW\n\nUnrelated content.\n' > "$FIXTURE/alpha_w.md"
printf '## BetaW\n\n%s\n\n' "$AFTER_ABOVE_W" > "$FIXTURE/beta_w.md"
git -C "$FIXTURE" add alpha_w.md beta_w.md
git -C "$FIXTURE" commit -qm repoint

write_tasks_md "$FIXTURE/tasks.md" "yes" "alpha_w.md" "beta_w.md"
run_generator "$FIXTURE" "$FIXTURE" "$BASE_W"
if [ "$RC" -ne 0 ]; then
  fail "case w: generator exited $RC: $ERR (F16 regression)"
elif [ ! -f "$OUTPUT_PATH" ]; then
  fail "case w: no output file written"
elif ! grep -q 'repointed' "$OUTPUT_PATH"; then
  fail "case w: repointed citation + dropped stale above misclassified (F16 regression):"$'\n'"$(cat "$OUTPUT_PATH")"
else
  pass "case w: repointed citation with dropped trailing above classified repointed (F16 regression guard)"
fi

# ===========================================================================
# Case (x): a scoped path that is a DIRECTORY at the merge-base ref and a
# plain file in the worktree (a realistic directory-to-file relocation
# shape) -> exits 2, not garbage rows from a tree listing read as text
# (F18 regression).
# ===========================================================================
new_fixture
mkdir -p "$FIXTURE/psi"
printf 'nested file content, irrelevant to the scoped path itself\n' > "$FIXTURE/psi/inner.md"
git -C "$FIXTURE" add psi/inner.md
git -C "$FIXTURE" commit -qm base
BASE_X="$(git -C "$FIXTURE" rev-parse HEAD)"

git -C "$FIXTURE" rm -rq psi
printf 'This is now a plain file where a directory used to be.\n' > "$FIXTURE/psi"
git -C "$FIXTURE" add psi

write_tasks_md "$FIXTURE/tasks.md" "yes" "psi"
run_generator "$FIXTURE" "$FIXTURE" "$BASE_X"
if [ "$RC" -eq 2 ]; then
  pass "case x: directory-to-file scoped path exits 2, not garbage rows (F18 regression guard)"
else
  fail "case x: expected exit 2 for a scoped path that was a directory at merge-base, got $RC: $ERR"
fi

# ===========================================================================
# Case (y): a moved passage under a `###` heading (not just `##`) (F19.1).
# ===========================================================================
new_fixture
{
  printf '### SectionH3A\n\n'
  printf 'This paragraph moves verbatim under a level-three heading only.\n\n'
} > "$FIXTURE/h3_src.md"
printf '### SectionH3B\n\nUnrelated content that stays put.\n' > "$FIXTURE/h3_dst.md"
git -C "$FIXTURE" add h3_src.md h3_dst.md
git -C "$FIXTURE" commit -qm base
BASE_Y="$(git -C "$FIXTURE" rev-parse HEAD)"

printf '### SectionH3A\n\nUnrelated content that stays put.\n' > "$FIXTURE/h3_src.md"
{
  printf '### SectionH3B\n\n'
  printf 'This paragraph moves verbatim under a level-three heading only.\n\n'
} > "$FIXTURE/h3_dst.md"
git -C "$FIXTURE" add h3_src.md h3_dst.md
git -C "$FIXTURE" commit -qm move

write_tasks_md "$FIXTURE/tasks.md" "yes" "h3_src.md" "h3_dst.md"
run_generator "$FIXTURE" "$FIXTURE" "$BASE_Y"
if [ "$RC" -ne 0 ]; then
  fail "case y: generator exited $RC: $ERR"
elif [ ! -f "$OUTPUT_PATH" ]; then
  fail "case y: no output file written"
elif ! grep -q 'moved' "$OUTPUT_PATH"; then
  fail "case y: no 'moved' row for a passage under a ### heading (F19.1):"$'\n'"$(cat "$OUTPUT_PATH")"
else
  pass "case y: moved passage under a ### heading classified moved (F19.1)"
fi

# ===========================================================================
# Case (z): a moved bullet using `* ` instead of `- ` (F19.2).
# ===========================================================================
new_fixture
{
  printf '## SectionStarA\n\n'
  printf -- '* This star bullet line will move verbatim to another file.\n\n'
} > "$FIXTURE/star_src.md"
printf '## SectionStarB\n\nUnrelated content that stays put.\n' > "$FIXTURE/star_dst.md"
git -C "$FIXTURE" add star_src.md star_dst.md
git -C "$FIXTURE" commit -qm base
BASE_Z="$(git -C "$FIXTURE" rev-parse HEAD)"

printf '## SectionStarA\n\nUnrelated content that stays put.\n' > "$FIXTURE/star_src.md"
{
  printf '## SectionStarB\n\n'
  printf -- '* This star bullet line will move verbatim to another file.\n\n'
} > "$FIXTURE/star_dst.md"
git -C "$FIXTURE" add star_src.md star_dst.md
git -C "$FIXTURE" commit -qm move

write_tasks_md "$FIXTURE/tasks.md" "yes" "star_src.md" "star_dst.md"
run_generator "$FIXTURE" "$FIXTURE" "$BASE_Z"
if [ "$RC" -ne 0 ]; then
  fail "case z: generator exited $RC: $ERR"
elif [ ! -f "$OUTPUT_PATH" ]; then
  fail "case z: no output file written"
elif ! grep -q 'moved' "$OUTPUT_PATH"; then
  fail "case z: no 'moved' row for a * bullet (F19.2):"$'\n'"$(cat "$OUTPUT_PATH")"
elif ! grep -q 'This star bullet line will move' "$OUTPUT_PATH"; then
  fail "case z: moved * bullet passage text not found in output (F19.2):"$'\n'"$(cat "$OUTPUT_PATH")"
else
  pass "case z: moved * bullet passage classified moved (F19.2)"
fi

# ===========================================================================
# Case (aa): a gap of 4+ characters between a backtick span and a bold span
# is NOT treated as a citation -> a content edit there classifies
# removed+added, never repointed (F19.3: CITATION_SEPARATOR_RE's bound is
# 0-3 chars).
# ===========================================================================
new_fixture
BEFORE_GAP='See `note.md` here and there **Alpha** unrelated text after the gap.'
AFTER_GAP='See `note.md` here and there **Beta** unrelated text after the gap.'
printf '## Omega\n\n%s\n\n' "$BEFORE_GAP" > "$FIXTURE/omega.md"
git -C "$FIXTURE" add omega.md
git -C "$FIXTURE" commit -qm base
BASE_AA="$(git -C "$FIXTURE" rev-parse HEAD)"
printf '## Omega\n\n%s\n\n' "$AFTER_GAP" > "$FIXTURE/omega.md"
git -C "$FIXTURE" add omega.md
git -C "$FIXTURE" commit -qm change

write_tasks_md "$FIXTURE/tasks.md" "yes" "omega.md"
run_generator "$FIXTURE" "$FIXTURE" "$BASE_AA"
if [ "$RC" -ne 0 ]; then
  fail "case aa: generator exited $RC: $ERR"
elif [ ! -f "$OUTPUT_PATH" ]; then
  fail "case aa: no output file written"
elif grep -q 'repointed' "$OUTPUT_PATH"; then
  fail "case aa: a 4+ char gap wrongly treated as a citation separator (F19.3):"$'\n'"$(cat "$OUTPUT_PATH")"
elif ! grep -q 'removed' "$OUTPUT_PATH" || ! grep -q 'added' "$OUTPUT_PATH"; then
  fail "case aa: expected removed+added rows for a bold-span edit beyond the citation gap:"$'\n'"$(cat "$OUTPUT_PATH")"
else
  pass "case aa: 4+ char gap not treated as citation adjacency (F19.3 guard)"
fi

# ===========================================================================
# Case (bb): a word merely containing "above"/"below" as a substring (e.g.
# ends in "...notabove") is NOT wrongly truncated by
# TRAILING_POSITION_WORD_RE's `\b` boundary -> the passage is classified
# removed+added, not repointed, since only the substring word differs
# (F19.4).
# ===========================================================================
new_fixture
BEFORE_SUBSTR='This sentence ends with the word notabove.'
AFTER_SUBSTR='This sentence ends with the word notbelow.'
printf '## Substr\n\n%s\n\n' "$BEFORE_SUBSTR" > "$FIXTURE/substr.md"
git -C "$FIXTURE" add substr.md
git -C "$FIXTURE" commit -qm base
BASE_BB="$(git -C "$FIXTURE" rev-parse HEAD)"
printf '## Substr\n\n%s\n\n' "$AFTER_SUBSTR" > "$FIXTURE/substr.md"
git -C "$FIXTURE" add substr.md
git -C "$FIXTURE" commit -qm change

write_tasks_md "$FIXTURE/tasks.md" "yes" "substr.md"
run_generator "$FIXTURE" "$FIXTURE" "$BASE_BB"
if [ "$RC" -ne 0 ]; then
  fail "case bb: generator exited $RC: $ERR"
elif [ ! -f "$OUTPUT_PATH" ]; then
  fail "case bb: no output file written"
elif grep -q 'repointed' "$OUTPUT_PATH"; then
  fail "case bb: substring-only 'above'/'below' word wrongly truncated (F19.4):"$'\n'"$(cat "$OUTPUT_PATH")"
elif ! grep -q 'removed' "$OUTPUT_PATH" || ! grep -q 'added' "$OUTPUT_PATH"; then
  fail "case bb: expected removed+added rows for a substring-only above/below edit:"$'\n'"$(cat "$OUTPUT_PATH")"
else
  pass "case bb: word merely containing above/below as a substring not truncated (F19.4 guard)"
fi

# ===========================================================================
# Case (cc): a plan declaring `**Relocation:** yes` where every task's
# `**Files:**` field is empty (zero scoped files total) -> no-op, exit 0,
# writes nothing (F19.5).
# ===========================================================================
new_fixture
printf '## Cc\n\nSome content.\n' > "$FIXTURE/cc.md"
git -C "$FIXTURE" add cc.md
git -C "$FIXTURE" commit -qm base
BASE_CC="$(git -C "$FIXTURE" rev-parse HEAD)"
printf '## Cc\n\nSome different content.\n' > "$FIXTURE/cc.md"
git -C "$FIXTURE" add cc.md
git -C "$FIXTURE" commit -qm change

{
  printf '# fixture plan\n\n'
  printf '> **Relocation:** yes\n\n'
  printf -- '- [ ] 1. Only task, no files scoped\n\n'
} > "$FIXTURE/tasks.md"
run_generator "$FIXTURE" "$FIXTURE" "$BASE_CC"
if [ "$RC" -ne 0 ]; then
  fail "case cc: generator exited $RC: $ERR"
elif [ -f "$OUTPUT_PATH" ]; then
  fail "case cc: output file written despite an empty file scope"
else
  pass "case cc: Relocation yes with zero scoped files is a no-op, no output written (F19.5)"
fi

# ===========================================================================
# Case (dd): a repointed citation where the BEFORE text's trailing `above`
# is followed by a RUN of punctuation (`above...`) and the AFTER text (which
# never had the word) ends in a different punctuation run (`!?`) on the same
# base sentence -> still classified "repointed", not removed+added (F20
# regression: the old single-character punctuation classes only ever
# stripped one mark from a narrow `.,;:` set, so a run, or `!`/`?`, left
# trailing punctuation behind after the word/run strip and diverged the two
# signatures even after a genuine repoint).
# ===========================================================================
new_fixture
BEFORE_RUN_DD='`old-file.md` **Old Heading** is the place to look above...'
AFTER_RUN_DD='`new-file.md` **New Heading** is the place to look!?'
printf '## AlphaDD\n\n%s\n\n' "$BEFORE_RUN_DD" > "$FIXTURE/alpha_dd.md"
printf '## BetaDD\n\nUnrelated content.\n' > "$FIXTURE/beta_dd.md"
git -C "$FIXTURE" add alpha_dd.md beta_dd.md
git -C "$FIXTURE" commit -qm base
BASE_DD="$(git -C "$FIXTURE" rev-parse HEAD)"

printf '## AlphaDD\n\nUnrelated content.\n' > "$FIXTURE/alpha_dd.md"
printf '## BetaDD\n\n%s\n\n' "$AFTER_RUN_DD" > "$FIXTURE/beta_dd.md"
git -C "$FIXTURE" add alpha_dd.md beta_dd.md
git -C "$FIXTURE" commit -qm repoint

write_tasks_md "$FIXTURE/tasks.md" "yes" "alpha_dd.md" "beta_dd.md"
run_generator "$FIXTURE" "$FIXTURE" "$BASE_DD"
if [ "$RC" -ne 0 ]; then
  fail "case dd: generator exited $RC: $ERR (F20 regression)"
elif [ ! -f "$OUTPUT_PATH" ]; then
  fail "case dd: no output file written"
elif ! grep -q 'repointed' "$OUTPUT_PATH"; then
  fail "case dd: punctuation-run repoint misclassified (F20 regression):"$'\n'"$(cat "$OUTPUT_PATH")"
else
  pass "case dd: trailing punctuation run (ellipsis / !?) not left behind after above-strip (F20 regression guard)"
fi

# ===========================================================================
# Case (ee): the two earlier punctuation-stripping stress shapes must still
# hold with the widened F20 regexes: (1) trailing punctuation present only
# on the after-side (no above/below word on either side) -> still classified
# "repointed" via the citation edit alone; (2) no above/below word at all,
# ordinary matching trailing punctuation on both sides -> also "repointed".
# ===========================================================================
new_fixture
BEFORE_EE='See the rule in `old.md`, **Old Section**, for the full rule'
AFTER_EE='See the rule in `new.md`, **New Section**, for the full rule.'
printf '## AlphaEE\n\n%s\n\n' "$BEFORE_EE" > "$FIXTURE/alpha_ee.md"
printf '## BetaEE\n\nUnrelated content.\n' > "$FIXTURE/beta_ee.md"
git -C "$FIXTURE" add alpha_ee.md beta_ee.md
git -C "$FIXTURE" commit -qm base
BASE_EE="$(git -C "$FIXTURE" rev-parse HEAD)"

printf '## AlphaEE\n\nUnrelated content.\n' > "$FIXTURE/alpha_ee.md"
printf '## BetaEE\n\n%s\n\n' "$AFTER_EE" > "$FIXTURE/beta_ee.md"
git -C "$FIXTURE" add alpha_ee.md beta_ee.md
git -C "$FIXTURE" commit -qm repoint

write_tasks_md "$FIXTURE/tasks.md" "yes" "alpha_ee.md" "beta_ee.md"
run_generator "$FIXTURE" "$FIXTURE" "$BASE_EE"
if [ "$RC" -ne 0 ]; then
  fail "case ee: generator exited $RC: $ERR"
elif [ ! -f "$OUTPUT_PATH" ]; then
  fail "case ee: no output file written"
elif ! grep -q 'repointed' "$OUTPUT_PATH"; then
  fail "case ee: punctuation-only-on-after-side / no-above-word stress case regressed by F20:"$'\n'"$(cat "$OUTPUT_PATH")"
else
  pass "case ee: punctuation-only-on-after-side and no-above-word stress cases still classified repointed"
fi

# ===========================================================================
# Case (ff): a scoped path that is a SYMLINK at the merge-base ref (pointing
# at a real blob) -> exits 2, not a spurious removed/added row (F21
# regression: `git cat-file -t` reports "blob" for a symlink object same as
# for ordinary content, but `git show <ref>:<path>` on a symlink returns the
# literal target path string, not resolved content).
# ===========================================================================
new_fixture
printf 'the real target file content, unrelated to the symlink path itself\n' > "$FIXTURE/real_target.md"
ln -s real_target.md "$FIXTURE/link_ff.md"
git -C "$FIXTURE" add real_target.md link_ff.md
git -C "$FIXTURE" commit -qm base
BASE_FF="$(git -C "$FIXTURE" rev-parse HEAD)"

# worktree side: same symlink, untouched -> a completely unchanged file if
# this classifier could compare it at all.

write_tasks_md "$FIXTURE/tasks.md" "yes" "link_ff.md"
run_generator "$FIXTURE" "$FIXTURE" "$BASE_FF"
if [ "$RC" -eq 2 ]; then
  pass "case ff: symlink scoped path at merge-base ref exits 2, not a spurious row (F21 regression guard)"
else
  fail "case ff: expected exit 2 for a symlink at the merge-base ref, got $RC: $ERR"
fi

# ===========================================================================
# Case (gg): a scoped path that is a SYMLINK in the WORKTREE, pointing at a
# real file whose content differs from the merge-base blob at that same
# path -> exits 2, not a spurious removed/added row (F21 regression: the
# worktree side is read via Python's `open()`, which follows the symlink and
# reads the real target's content, unlike the git-side literal target
# string).
# ===========================================================================
new_fixture
printf 'original content at the scoped path, as an ordinary regular file\n' > "$FIXTURE/gg.md"
git -C "$FIXTURE" add gg.md
git -C "$FIXTURE" commit -qm base
BASE_GG="$(git -C "$FIXTURE" rev-parse HEAD)"

# worktree side: gg.md is now a symlink to a different file with different
# content than the merge-base blob.
printf 'different real content the symlink target actually holds\n' > "$FIXTURE/gg_target.md"
rm "$FIXTURE/gg.md"
ln -s gg_target.md "$FIXTURE/gg.md"

write_tasks_md "$FIXTURE/tasks.md" "yes" "gg.md"
run_generator "$FIXTURE" "$FIXTURE" "$BASE_GG"
if [ "$RC" -eq 2 ]; then
  pass "case gg: symlink scoped path in the worktree exits 2, not a spurious row (F21 regression guard)"
else
  fail "case gg: expected exit 2 for a symlink in the worktree, got $RC: $ERR"
fi

# ===========================================================================
# Case (hh): a repointed citation where the BEFORE text's trailing `above`
# word is followed by WHITESPACE and then a punctuation run
# ("...look above  .", word, then whitespace, then punctuation) and the
# AFTER text (which never had the word) ends in the same punctuation with no
# extra whitespace ("...look .") -> still classified "repointed", not
# removed+added (F22 regression: TRAILING_POSITION_WORD_RE required the
# punctuation run to sit IMMEDIATELY after the word, so intervening
# whitespace left the word's trailing space behind after stripping,
# diverging the two signatures even though the only real edit was the
# permitted one).
# ===========================================================================
new_fixture
BEFORE_RUN_HH='`old-file.md` **Old Heading** is the place to look above  .'
AFTER_RUN_HH='`new-file.md` **New Heading** is the place to look .'
printf '## AlphaHH\n\n%s\n\n' "$BEFORE_RUN_HH" > "$FIXTURE/alpha_hh.md"
printf '## BetaHH\n\nUnrelated content.\n' > "$FIXTURE/beta_hh.md"
git -C "$FIXTURE" add alpha_hh.md beta_hh.md
git -C "$FIXTURE" commit -qm base
BASE_HH="$(git -C "$FIXTURE" rev-parse HEAD)"

printf '## AlphaHH\n\nUnrelated content.\n' > "$FIXTURE/alpha_hh.md"
printf '## BetaHH\n\n%s\n\n' "$AFTER_RUN_HH" > "$FIXTURE/beta_hh.md"
git -C "$FIXTURE" add alpha_hh.md beta_hh.md
git -C "$FIXTURE" commit -qm repoint

write_tasks_md "$FIXTURE/tasks.md" "yes" "alpha_hh.md" "beta_hh.md"
run_generator "$FIXTURE" "$FIXTURE" "$BASE_HH"
if [ "$RC" -ne 0 ]; then
  fail "case hh: generator exited $RC: $ERR (F22 regression)"
elif [ ! -f "$OUTPUT_PATH" ]; then
  fail "case hh: no output file written"
elif ! grep -q 'repointed' "$OUTPUT_PATH"; then
  fail "case hh: whitespace-before-punctuation repoint misclassified (F22 regression):"$'\n'"$(cat "$OUTPUT_PATH")"
else
  pass "case hh: trailing word + whitespace + punctuation not left diverging after above-strip (F22 regression guard)"
fi

# ===========================================================================
# Summary
# ===========================================================================
if [ "$FAILURES" -eq 0 ]; then
  echo "all cases passed"
  exit 0
else
  echo "$FAILURES case(s) failed"
  exit 1
fi
