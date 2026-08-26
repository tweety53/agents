#!/usr/bin/env bash
# test-check-normative-inventory.sh — harness for check-normative-inventory.sh.
#
# Every case runs against a sandboxed fixture tree under mktemp, reached through
# CHECK_NORMATIVE_INVENTORY_ROOT, so the harness never depends on the real
# repository's current prose. A harness that asserted against the real corpus
# would go red every time a requirement was legitimately edited — and the whole
# point of this guard is that the corpus IS edited, in bulk, while the guard
# stays trustworthy.
#
# The six cases the guard's requirement names, in the order the requirement
# names them: a deleted SHALL sentence changes the inventory; a reflowed
# sentence does not; a reworded sentence does; the output is order-independent;
# an excluded tree contributes nothing; an unreadable scope root exits 2. The
# cases after those exercise the contract's remaining edges — the exit-2 shapes,
# the block-boundary rule, and the deliberate absence of de-duplication.
#
# EVERY CASE HERE WAS PROVED AGAINST A DELIBERATELY BROKEN GUARD before it was
# trusted: a copy of the guard with one mutation applied, run through this
# harness, which had to go red. A case that stays green against the mutation
# that breaks the behaviour it names is testing nothing, and "exit 0" satisfies
# an assertion-free case. The mutations used are recorded in this change's task
# record, not re-run here — a mutation harness pinned to sed patterns over the
# guard's source goes stale the first time the guard is reworded, and a stale
# mutation that silently stops applying is the same vacuous pass in a costlier
# form.
set -euo pipefail

GUARD="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/check-normative-inventory.sh"
[ -x "$GUARD" ] || { printf 'guard not executable: %s\n' "$GUARD" >&2; exit 2; }

failures=0

ok() { printf 'ok   %s\n' "$1"; }
bad() {
  printf 'FAIL %s\n' "$1"
  failures=$((failures + 1))
}

FIX="$(mktemp -d)"
trap 'chmod -R u+rwX "$FIX" 2>/dev/null || true; rm -rf "$FIX"' EXIT

# mkroot <dir> — a fixture root carrying every scope root the guard requires.
# A root missing one of them is an exit-2 case, exercised deliberately below;
# every other case starts from a complete root so it tests what it says it does.
mkroot() {
  mkdir -p "$1/skills" "$1/rules" "$1/spectre/specs" "$1/commands" \
    "$1/commands-claude" "$1/.flow"
}

# inv <root> — the inventory for <root>. A non-zero exit here is a harness
# abort, not a recorded failure: every caller of this helper is a case whose
# premise is that the guard CAN answer, and continuing past a broken premise
# would report unrelated nonsense.
inv() {
  local root="$1" out rc=0
  out="$(CHECK_NORMATIVE_INVENTORY_ROOT="$root" "$GUARD" 2>/dev/null)" || rc=$?
  if [ "$rc" -ne 0 ]; then
    printf 'harness: guard exited %d on fixture %s, wanted 0\n' "$rc" "$root" >&2
    exit 2
  fi
  printf '%s\n' "$out"
}

# expect_exit <label> <want-exit> <root>
expect_exit() {
  local label="$1" want="$2" root="$3" got=0
  CHECK_NORMATIVE_INVENTORY_ROOT="$root" "$GUARD" >/dev/null 2>&1 || got=$?
  if [ "$got" -eq "$want" ]; then
    ok "$label"
  else
    bad "$label: exit $got, wanted $want"
  fi
}

# --- Case 1: a deleted SHALL sentence changes the inventory ----------------
#
# The failure this whole guard exists to catch. Two trees identical but for one
# deleted normative sentence inside an otherwise untouched paragraph.

mkroot "$FIX/deleted-before"
cat > "$FIX/deleted-before/skills/a.md" <<'EOF'
# A skill

The guard SHALL print every normative sentence. It runs over the owned corpus
only. Reading it is cheap.
EOF

mkroot "$FIX/deleted-after"
cat > "$FIX/deleted-after/skills/a.md" <<'EOF'
# A skill

It runs over the owned corpus only. Reading it is cheap.
EOF

before="$(inv "$FIX/deleted-before")"
after="$(inv "$FIX/deleted-after")"
if [ "$before" = 'The guard SHALL print every normative sentence.' ] \
  && [ -z "${after//[[:space:]]/}" ]; then
  ok 'a deleted SHALL sentence changes the inventory'
else
  bad "a deleted SHALL sentence: before=[$before] after=[$after]"
fi

# --- Case 2: reflowing a paragraph is not a change -------------------------
#
# The property that makes the guard usable: the same wording rewrapped across
# different line boundaries, with different indentation and a tab thrown in,
# must produce byte-identical output. Without it every trim task would drown in
# diffs that carry no lost requirement.

mkroot "$FIX/flow-one-line"
cat > "$FIX/flow-one-line/rules/a.mdc" <<'EOF'
It SHALL normalise each sentence's internal whitespace to single spaces and strip surrounding whitespace, so a reflowed paragraph does not read as a changed requirement.
EOF

mkroot "$FIX/flow-wrapped"
printf 'It SHALL normalise each sentence'"'"'s internal   whitespace to single\n' \
  > "$FIX/flow-wrapped/rules/a.mdc"
printf '\tspaces and strip surrounding whitespace, so a reflowed\n' \
  >> "$FIX/flow-wrapped/rules/a.mdc"
printf '  paragraph does not read as a changed requirement.\n' \
  >> "$FIX/flow-wrapped/rules/a.mdc"

if [ "$(inv "$FIX/flow-one-line")" = "$(inv "$FIX/flow-wrapped")" ]; then
  ok 'a reflowed sentence does not change the inventory'
else
  bad "a reflowed sentence changed the inventory: [$(inv "$FIX/flow-one-line")] vs [$(inv "$FIX/flow-wrapped")]"
fi

# --- Case 3: rewording a requirement is a change ---------------------------
#
# The other half of case 2. Whitespace is the ONLY thing normalised, so a
# same-meaning reword is a visible difference — which is what makes
# "cut, never paraphrase" enforceable rather than merely asked for.

mkroot "$FIX/reword"
cat > "$FIX/reword/rules/a.mdc" <<'EOF'
It SHALL emit each sentence's internal whitespace as single spaces.
EOF

if [ "$(inv "$FIX/flow-one-line")" != "$(inv "$FIX/reword")" ]; then
  ok 'a reworded sentence changes the inventory'
else
  bad 'a reworded sentence did not change the inventory'
fi

# --- Case 4: the output is order-independent --------------------------------
#
# The same two sentences, distributed across differently-named files in
# different scope roots so the traversal order genuinely differs between the two
# trees. Running the same tree twice would pass without any sort at all, since
# find's order over one tree is stable — that fixture would test nothing.

mkroot "$FIX/order-a"
printf 'The inventory SHALL be sorted.\n' > "$FIX/order-a/skills/zzz.md"
printf 'The corpus MUST be resolved once.\n' > "$FIX/order-a/rules/aaa.mdc"

mkroot "$FIX/order-b"
printf 'The corpus MUST be resolved once.\n' > "$FIX/order-b/skills/aaa.md"
printf 'The inventory SHALL be sorted.\n' > "$FIX/order-b/rules/zzz.mdc"

order_a="$(inv "$FIX/order-a")"
if [ "$order_a" = "$(inv "$FIX/order-b")" ]; then
  ok 'the inventory is independent of file order'
else
  bad 'the inventory depends on file order'
fi
if [ "$order_a" = "$(printf 'The corpus MUST be resolved once.\nThe inventory SHALL be sorted.')" ]; then
  ok 'the inventory is sorted, not merely stable'
else
  bad "the inventory is not sorted: [$order_a]"
fi

# --- Case 5: an excluded tree contributes nothing ---------------------------
#
# Each excluded path carries a normative sentence that must not appear. The
# exclusions are structural — a path component named node_modules or
# .superpowers anywhere, and the spectre/changes/archive and docs/superpowers
# prefixes — never a list of filenames, so these fixtures use ordinary names
# that would otherwise be scanned.

mkroot "$FIX/excluded-base"
printf 'The owned corpus SHALL be scanned.\n' > "$FIX/excluded-base/skills/a.md"

mkroot "$FIX/excluded-plus"
printf 'The owned corpus SHALL be scanned.\n' > "$FIX/excluded-plus/skills/a.md"
mkdir -p "$FIX/excluded-plus/skills/vendor/node_modules/pkg" \
  "$FIX/excluded-plus/.superpowers/sdd" \
  "$FIX/excluded-plus/docs/superpowers/skills" \
  "$FIX/excluded-plus/spectre/changes/archive/kan-1/specs"
printf 'A vendored package SHALL not be inventoried.\n' \
  > "$FIX/excluded-plus/skills/vendor/node_modules/pkg/a.md"
printf 'A session record SHALL not be inventoried.\n' \
  > "$FIX/excluded-plus/.superpowers/sdd/a.md"
printf 'A vendored skill SHALL not be inventoried.\n' \
  > "$FIX/excluded-plus/docs/superpowers/skills/a.md"
printf 'An archived spec SHALL not be inventoried.\n' \
  > "$FIX/excluded-plus/spectre/changes/archive/kan-1/specs/a.md"

if [ "$(inv "$FIX/excluded-base")" = "$(inv "$FIX/excluded-plus")" ]; then
  ok 'an excluded tree contributes nothing'
else
  bad "an excluded tree contributed: [$(inv "$FIX/excluded-plus")]"
fi

# --- Case 6: an unreadable scope root cannot be answered --------------------
#
# Exit 2, never a quietly smaller inventory. A guard that skipped an unreadable
# scope root would report a clean diff against a baseline it never re-derived,
# which is the one failure a trim cannot survive.

#
# The fixture carries a readable file in a SECOND scope root on purpose. With
# only the unreadable root's own file in it, a guard that skipped that root
# would find no Markdown at all and exit 2 through the empty-corpus refusal
# below — passing this case for the wrong reason, and leaving the skip
# undetected.
mkroot "$FIX/unreadable"
printf 'The corpus SHALL be scanned.\n' > "$FIX/unreadable/skills/a.md"
printf 'The scope roots MUST all be readable.\n' > "$FIX/unreadable/rules/b.mdc"
chmod 000 "$FIX/unreadable/skills"
if [ "$(id -u)" = "0" ]; then
  printf 'skip an unreadable scope root cannot be answered (running as root reads it anyway)\n'
else
  expect_exit 'an unreadable scope root cannot be answered' 2 "$FIX/unreadable"
fi
chmod 755 "$FIX/unreadable/skills"

# --- The remaining exit-2 shapes -------------------------------------------

mkdir -p "$FIX/missing-root/skills" "$FIX/missing-root/rules"
printf 'The corpus SHALL be scanned.\n' > "$FIX/missing-root/skills/a.md"
expect_exit 'a missing scope root cannot be answered' 2 "$FIX/missing-root"

# A scope root replaced by a SYMLINK to a directory. The readability tests the
# corpus library applies to a scope root all follow symlinks, so a link passes
# them; `find` then runs in physical mode and does not descend into a symlinked
# starting point, so the entire scope root contributes zero files. That is the
# quietly smaller corpus the missing-root refusal above exists to prevent,
# reached by a different door — and the fixture proves it is refused rather than
# skipped by carrying a normative sentence in a SECOND scope root, so a guard
# that skipped the link would find Markdown, print a smaller inventory and exit
# 0 instead of falling through to the empty-corpus refusal for the wrong reason.
mkroot "$FIX/symlinked-root"
printf 'The corpus SHALL be scanned.\n' > "$FIX/symlinked-root/rules/b.mdc"
mkdir -p "$FIX/symlinked-root/real-skills"
printf 'A symlinked scope root SHALL be refused.\n' \
  > "$FIX/symlinked-root/real-skills/a.md"
rm -rf "$FIX/symlinked-root/skills"
ln -s "$FIX/symlinked-root/real-skills" "$FIX/symlinked-root/skills"
expect_exit 'a symlinked scope root cannot be answered' 2 "$FIX/symlinked-root"

# A symlinked DIRECTORY nested INSIDE a scope root, whose target holds Markdown.
# The scope-root refusal above does not reach it: it tests `$root/$dir` alone,
# and the `find` that follows runs in physical mode, so it walks past a
# symlinked subdirectory without descending into it. Everything the link points
# at is then absent from the inventory with no error and no warning — the same
# quietly smaller corpus as a symlinked scope root, one level deeper. The
# fixture puts a normative sentence on BOTH sides of the link, so a guard that
# skipped the link would print the visible sentence and exit 0 rather than
# falling through the empty-corpus refusal for the wrong reason.
mkroot "$FIX/nested-link-md"
printf 'The visible corpus SHALL be scanned.\n' > "$FIX/nested-link-md/skills/top.md"
mkdir -p "$FIX/nested-link-md/outside/hidden"
printf 'The hidden corpus SHALL be scanned too.\n' \
  > "$FIX/nested-link-md/outside/hidden/hidden.md"
ln -s "$FIX/nested-link-md/outside/hidden" "$FIX/nested-link-md/skills/nested-link"
expect_exit 'a nested symlinked directory holding Markdown cannot be answered' 2 \
  "$FIX/nested-link-md"

# The same shape carrying NO Markdown — this repository's own
# skills/flow*/scripts/lib links, which point at a directory of `.sh` files.
# Nothing is lost by not descending into one, so refusing it would fail the real
# repository on every run for no lost sentence. It contributes nothing and the
# run stays green.
mkroot "$FIX/nested-link-nomd"
printf 'The visible corpus SHALL be scanned.\n' > "$FIX/nested-link-nomd/skills/top.md"
mkdir -p "$FIX/nested-link-nomd/outside/lib"
printf 'echo hi\n' > "$FIX/nested-link-nomd/outside/lib/helper.sh"
ln -s "$FIX/nested-link-nomd/outside/lib" "$FIX/nested-link-nomd/skills/nested-link"
got_nested="$(inv "$FIX/nested-link-nomd")"
if [ "$got_nested" = 'The visible corpus SHALL be scanned.' ]; then
  ok 'a nested symlinked directory holding no Markdown is skipped, not refused'
else
  bad "a nested symlinked directory holding no Markdown: got [$got_nested]"
fi

expect_exit 'a nonexistent repository root cannot be answered' 2 "$FIX/absent"

# A complete set of scope roots holding no Markdown at all. Reporting an empty
# inventory here would be indistinguishable from a corpus whose every normative
# sentence had just been deleted — the exact comparison this guard feeds.
mkroot "$FIX/no-markdown"
expect_exit 'a corpus with no Markdown at all cannot be answered' 2 "$FIX/no-markdown"

# An override that is set but empty must be refused rather than falling back to
# the script's own location: a fallback there makes the harness scan the real
# repository while believing it scanned a fixture.
rc=0
CHECK_NORMATIVE_INVENTORY_ROOT='' "$GUARD" >/dev/null 2>&1 || rc=$?
if [ "$rc" -eq 2 ]; then
  ok 'an empty override is refused'
else
  bad "an empty override exits $rc, wanted 2"
fi

# --- The block-boundary rule ------------------------------------------------
#
# One fixture per Markdown shape the rule names, because each is a separate
# branch of the splitter and none of the cases above reaches any of them: a
# bullet, a table cell, a heading, a blockquote, and a fenced line. The
# assertion is the exact set of lines, so a shape that silently stops being
# scanned — or one that swallows its neighbour — fails here.

mkroot "$FIX/shapes"
cat > "$FIX/shapes/commands/a.md" <<'EOF'
## A heading that SHALL be its own block

- A bullet that MUST stand alone.
- Another bullet, with no obligation.

| Rule | Detail |
|------|--------|
| A cell that MUST stand alone. | An adjacent cell with no obligation. |

> A quoted line that SHALL be unwrapped
> across two source lines.

```bash
# A fenced line that MUST be scanned.
```
EOF

want_shapes="$(printf '%s\n' \
  '## A heading that SHALL be its own block' \
  '# A fenced line that MUST be scanned.' \
  'A bullet that MUST stand alone.' \
  'A cell that MUST stand alone.' \
  'A quoted line that SHALL be unwrapped across two source lines.' \
  | LC_ALL=C sort)"
got_shapes="$(inv "$FIX/shapes")"
if [ "$got_shapes" = "$want_shapes" ]; then
  ok 'each block shape yields its own sentence'
else
  bad "block shapes: got [$got_shapes] wanted [$want_shapes]"
fi

# --- Duplicates are kept ----------------------------------------------------
#
# This change deletes restatement, so the same normative sentence really does
# appear in two files today. De-duplicating would let one of the two copies be
# deleted with the inventory unchanged — a silently lost requirement, in exactly
# the files this change edits most.

mkroot "$FIX/dupes"
printf 'A restated sentence SHALL survive in both files.\n' > "$FIX/dupes/skills/a.md"
printf 'A restated sentence SHALL survive in both files.\n' > "$FIX/dupes/rules/a.mdc"
# `inv` is called into a BARE ASSIGNMENT, the way every other call site here
# does it, because that is the only shape its documented hard-fail survives:
# piping its output, or nesting it in an `if` condition, discards the exit 2 it
# uses to abort the harness on a broken premise, and would degrade that abort
# into a generic FAIL naming the wrong thing.
dupes="$(inv "$FIX/dupes")"
if [ "$(printf '%s\n' "$dupes" | wc -l | tr -d ' ')" = "2" ]; then
  ok 'a sentence restated in two files is printed twice'
else
  bad "a restated sentence was de-duplicated: [$dupes]"
fi

# --- Only the four keywords count ------------------------------------------
#
# SHOULD and MAY state no obligation whose loss changes behaviour; a keyword
# glued to another word is not the keyword.

mkroot "$FIX/keywords"
cat > "$FIX/keywords/spectre/specs/a.md" <<'EOF'
The guard SHOULD be fast. It MAY print a count.

A caller MUST_NOT_TOUCH the table.

The exit code MUST be 2.
EOF
if [ "$(inv "$FIX/keywords")" = 'The exit code MUST be 2.' ]; then
  ok 'only SHALL and MUST as whole words are keywords'
else
  bad "keyword selection: [$(inv "$FIX/keywords")]"
fi

# --- A substantive change after an abbreviation is a change -----------------
#
# The defect this case pins: a splitter that treats an abbreviation's period as
# a sentence terminator inventories the stub before it and nothing else, so the
# obligation's entire substance can be replaced with the inventory unmoved. The
# stub is stable, which is what makes it dangerous — a keyword-occurrence
# reconciliation stays green through the swap, because the keyword never moves.

mkroot "$FIX/abbrev-before"
printf 'The tool SHALL validate e.g. input for XSS and SQL injection before use.\n' \
  > "$FIX/abbrev-before/skills/a.md"
mkroot "$FIX/abbrev-after"
printf 'The tool SHALL validate e.g. nothing at all here whatsoever.\n' \
  > "$FIX/abbrev-after/skills/a.md"

abbrev_before="$(inv "$FIX/abbrev-before")"
if [ "$abbrev_before" = 'The tool SHALL validate e.g. input for XSS and SQL injection before use.' ]; then
  ok 'a sentence is not truncated at an abbreviation'
else
  bad "a sentence was truncated at an abbreviation: [$abbrev_before]"
fi
if [ "$abbrev_before" != "$(inv "$FIX/abbrev-after")" ]; then
  ok 'a substantive change after an abbreviation changes the inventory'
else
  bad 'a substantive change after an abbreviation did not change the inventory'
fi

# Every abbreviation the guard exempts, one sentence each, each with its
# substance AFTER the abbreviation so a truncating splitter loses it. An
# abbreviation dropped from the guard's list fails here rather than silently
# reopening the defect above for that spelling.

mkroot "$FIX/abbrev-list"
cat > "$FIX/abbrev-list/rules/a.mdc" <<'EOF'
- A caller MUST pass a root, e.g. the repository it was asked to scan.
- A caller MUST pass one root, i.e. exactly one directory.
- The guard SHALL refuse a symlink, cf. the budget guard's own refusal.
- Bytes SHALL be counted, vs. lines, which a long-line file cheats.
- The corpus SHALL cover skills, rules, specs, etc. as one owned set.
- The finding SHALL cite Knuth et al. as its source.
- The report SHALL name Dr. Reviewer as its author.
- The row SHALL carry No. 7 as its identifier.
EOF

want_abbrev="$(printf '%s\n' \
  'A caller MUST pass a root, e.g. the repository it was asked to scan.' \
  'A caller MUST pass one root, i.e. exactly one directory.' \
  "The guard SHALL refuse a symlink, cf. the budget guard's own refusal." \
  'Bytes SHALL be counted, vs. lines, which a long-line file cheats.' \
  'The corpus SHALL cover skills, rules, specs, etc. as one owned set.' \
  'The finding SHALL cite Knuth et al. as its source.' \
  'The report SHALL name Dr. Reviewer as its author.' \
  'The row SHALL carry No. 7 as its identifier.' \
  | LC_ALL=C sort)"
got_abbrev="$(inv "$FIX/abbrev-list")"
if [ "$got_abbrev" = "$want_abbrev" ]; then
  ok 'every exempt abbreviation keeps its sentence whole'
else
  bad "abbreviation list: got [$got_abbrev] wanted [$want_abbrev]"
fi

if [ "$failures" -ne 0 ]; then
  printf '%d failure(s)\n' "$failures"
  exit 1
fi
printf 'all checks passed\n'
