#!/usr/bin/env bash
# check-cleanup-complete.sh — verify that everything the cleanup registry says
# should be gone after /myflow-finish run 2 actually is.
#
# Usage: check-cleanup-complete.sh <repo> <change-name> <state-dir>
#
# Prints ONE verdict line to stdout:
#   COMPLETE: <reason>    every registry row whose lifetime ends at run 2 is gone
#   LEFTOVER: <breakdown> one or more are still present
#
# EITHER VERDICT MAY CARRY NOTES, appended after ` — `, and a `SKIPPED:` note
# says a row was NOT verified. It rides a COMPLETE line because a skip does not
# block — see the asymmetry paragraph below — so a caller that reads the token
# and discards the rest reports "cleanup verified" over a row nothing looked at,
# which is this guard's whole argument inverted one layer up. The rule that the
# clause is relayed word for word therefore lives with the consumer, in step 6
# of **Run 2 — the branch is merged** (`skills/myflow-contracts/pipeline.md`),
# and is named here so a future editor of this line knows where it is kept.
#
# Exit 0 whenever a verdict was reached; exit 2 when it cannot answer at all —
# an unreadable repository or state directory, or a change name outside the
# allowlist. The VERDICT carries the answer, not the exit status
# — see check-finish-preflight.sh's header for why this repository separates
# them. Finish runs this once per repository, so each verdict names the
# repository it judged; a bare COMPLETE would be unattributable across several.
#
# WHY THE VERDICT WORDS DIFFER FROM check-unfinished-work.sh's. That guard says
# CLEAR/OUTSTANDING, this one says COMPLETE/LEFTOVER, and the two vocabularies
# are deliberately not shared. They answer OPPOSITE questions about absence, as
# the paragraph below spells out: there a missing file counts AGAINST the
# change, here a missing artifact IS the answer. A reader who carried the
# meaning of CLEAR across to COMPLETE would carry that inversion with it.
# check-finish-preflight.sh's third vocabulary (RUN1/RUN2/REFUSE) differs for a
# further reason again: it selects a procedure rather than reporting a state.
#
# Reporting a leftover is the whole point: run 2 previously assumed its own
# removals succeeded.
#
# THE SIX ROWS, AND WHY THEY ARE THESE SIX. The registry in
# skills/myflow-contracts/pipeline.md owns the list; the rows whose lifetime
# ends at run 2 are the worktree, the local branch, the remote branch, the
# change directory, the proposal artifact source and the workspace database and
# bucket. The per-task diffs, the panel record and the session ledger are not
# checked separately because the registry removes them WITH the worktree — a
# surviving one of those is a surviving worktree, already reported. The state
# file is not checked at all: its row says it is never removed, so its presence
# is correct and reporting it would train the operator to ignore this guard's
# output. The claimed cache index is not checked for a different reason again,
# and it is the only row whose reason is an inability rather than a decision:
# its row says nothing in this pipeline removes it and nothing can, because the
# index is claimed by probing rather than derived from the change name and is
# never recorded, so by the time this guard runs there is nothing that could say
# which index to look at. A row this guard invented a check for would be
# checking an index it guessed.
#
# THE SIXTH ROW IS ANSWERED BY ASKING, NOT BY LOOKING, and it is the only one
# that is. The other five live in this repository's own git bookkeeping and
# filesystem, which this guard can read directly. A workspace's database and
# bucket live inside services this guard knows nothing about, and it must stay
# project-agnostic — it can hold neither `psql -l` nor any project's
# object-store client. So the project answers for them: it declares a
# `survivors` command in its .myflow/project.md, this guard runs it, and its
# output and exit code are the row's verdict. Both are specified under "What
# `survivors` prints, and what its exit code means" in
# skills/myflow-contracts/project-configuration.md, which is canonical; the
# reasoning for a third verb beside `create` and `remove` is under "Creation and
# cleanup" in skills/myflow-contracts/workspace-isolation.md.
#
# "RAN THE REMOVAL" IS NOT "VERIFIED GONE", which is the whole reason that row
# is not settled from the removal command's exit code. A removal that reported
# success against a stale connection, and a bucket a policy refused to delete,
# both leave the row's promise broken with nothing having failed. This guard
# therefore never runs `remove` and never reads its result — it asks.
#
# THE TWO NON-EMPTY ANSWERS ARE NOT SYMMETRIC. A reported survivor BLOCKS the
# terminal state; a `survivors` command that could not reach its service is
# reported by name and with its exit code, and the run continues. Blocking an
# already-merged change over a service that happens to be stopped trades a
# stranded change for a few megabytes of stale storage, which is the wrong
# trade — the argument is under "Creation and cleanup" in
# skills/myflow-contracts/workspace-isolation.md, together with why one non-zero
# exit is enough where a reader might expect two. A project that declares no
# `survivors` command, one whose configuration this guard cannot read — absent
# permission, but also a path that is not a regular file, and any scan of it that
# FAILED rather than found nothing — and one that declares the
# `## workspace isolation` section more than once, are all reported as SKIPPED
# for the same reason and with the same effect: skipped is never passed, and a
# skip does not block. An input this guard cannot resolve is reported rather than
# resolved to one of its readings, and a command whose failure would otherwise be
# indistinguishable from its negative answer is read by its exit status rather
# than by its empty output. A project that
# declares no `## workspace isolation` section at all is silent here, because a
# step whose artifact is already absent is a success.
#
# THAT DERIVATION IS DECLARED BELOW RATHER THAN LEFT IN PROSE, because a prose
# restatement of a table in another file is exactly what goes stale: a registry
# row added later whose lifetime ends at run 2 would not be checked here, and
# this guard would then report COMPLETE over a real leftover — the failure it
# exists to prevent, one layer up. Every registry row is named in exactly one of
# the two lists below, and test-check-cleanup-complete.sh reads BOTH the
# registry and these markers and fails when they disagree in either direction:
# a registry row with no line here, or a line here naming a row the registry no
# longer has. Adding a registry row therefore fails this guard's suite until
# someone decides, in writing, which list it belongs in.
#
# registry-row-checked: Worktree
# registry-row-checked: Local branch
# registry-row-checked: Remote branch
# registry-row-checked: Change directory
# registry-row-checked: Proposal artifact source
# registry-row-checked: Workspace database and bucket
# registry-row-not-checked: Per-task and review diffs — removed with the worktree
# registry-row-not-checked: Panel record — lives in the store; nothing removes it
# registry-row-not-checked: SDD ledger — lives in the store; nothing removes it
# registry-row-not-checked: Rendered ledger and panel record — committed and
#   archived with the change, so nothing removes them and there is nothing for
#   this guard to find gone
# registry-row-not-checked: Dispatch context bundle — removed with the worktree
# registry-row-not-checked: Archive branch — nothing in this pipeline removes it;
#   run 2 is terminal and the pull request it opens outlives the run, so there is
#   no later run to delete the branch it was opened from (kan-239)
# registry-row-not-checked: State file — never removed; it is the terminal record
# registry-row-not-checked: Claimed cache index — this pipeline removes nothing and this guard checks nothing; the index is probed rather than derived, so run 2 has no derivation to repeat. A project that writes its claim where a probe can see it may release it in its own `remove` command and report it through `survivors`; that is the project's tooling and this marker does not claim it
#
# ABSENCE IS THE ANSWER HERE, not a gap in the evidence. The run-1 gate treats
# a file it cannot find as outstanding, because a missing record proves nothing
# about the work. This guard asks the opposite question — "is it gone?" — so a
# row it cannot find is that row answered. What must never be inferred is
# absence from a path that was never readable, which is why an unreadable
# repository and an unreadable state directory both refuse to answer instead.
#
# ONE RESIDUAL EXCEPTION TO THAT SENTENCE, NAMED RATHER THAN CLAIMED AWAY. The
# local-branch and remote-tracking-ref rows are the only two answered out of
# git's ref store rather than out of the filesystem, and for them the sentence
# above holds for every shape but one. A ref whose LOOSE FILE is unreadable or
# corrupt is now told apart from a ref that does not exist, and reported SKIPPED
# — see ref_state below for the two commands that takes and why one will not do.
# A ref whose ANCESTOR DIRECTORY is unreadable is not: git enumerates the refs
# it can reach and reports nothing at all about the ones it cannot, in
# `show-ref`, in `rev-parse`, in `for-each-ref` and in `ls-remote` alike, with
# and without --quiet, so `refs/heads/spectre/` at mode 000 is indistinguishable
# from an empty `refs/heads/spectre/`. There is no channel to read, so there is
# nothing this guard can do but say so here.
#
# WHAT AN OPERATOR SHOULD KNOW. In that one shape a COMPLETE verdict does not
# prove the two branch rows are gone; it proves nothing was found where this
# guard was able to look. It is silent — no skip note, because the condition is
# unobservable from here — so it is the one leftover this guard cannot promise
# to report. If a run 2 that verified COMPLETE is followed by `git branch` still
# listing `spectre/<name>`, check the modes on `.git/refs/heads/` and its
# subdirectories before suspecting anything else. Every OTHER unreadable input
# in this guard — the repository, the state directory, packed-refs, the project
# configuration, the survivors command — refuses or reports SKIPPED.
#
# THE SURVIVORS COMMAND IS RUN UNDER A BOUNDED WAIT, and a timeout is a third
# skip reason rather than a fourth verdict. Run 2 is unattended, so a command
# that never returns strands an already-merged change short of FINISHED with
# nobody watching — and closing stdin, which this guard also does, only covers a
# command blocking on an interactive prompt. An unreachable host with no connect
# timeout, a lock wait, a frozen container and an infinite loop are all untouched
# by it. The bound is the 60 seconds **Worktree cleanup**
# (`skills/myflow-contracts/pipeline.md`) already gives the project-supplied
# `## stop` command, and the OUTCOME deliberately differs: there a timeout is a
# failed check, because an un-stopped stack is a reason not to remove a worktree
# and there is no other answer to fall back on; here the contract already defines
# one for "the command said nothing about survivors" — the reported skip — and
# blocking an already-merged change over a slow command is what
# **Creation and cleanup** (`skills/myflow-contracts/workspace-isolation.md`)
# forbids. The skip names the timeout distinctly from a non-zero exit, because
# "your command is slow" and "your command is broken" have different remedies.
#
# THE ARCHIVED CHANGE DIRECTORY IS NOT A LEFTOVER. The registry says the change
# directory is MOVED into spectre/changes/archive/ and never deleted, so only
# one still sitting at spectre/changes/<name>/ counts.
#
# A <name>-fix-N SUB-CHANGE COUNTS TOO, and used not to. Under spectre a
# sub-change is a FLAT SIBLING of its parent under spectre/changes/, never a
# directory inside it -- `spectre new` refuses an id that is not a single flat
# directory name -- so `spectre archive <name>` cannot reach one and each
# sub-change needs its own call (run 2 step 3,
# skills/myflow-contracts/finish-contract.md). Before this row the guard
# reported COMPLETE with the parent archived and the child left behind, and
# nothing anywhere said so.
set -euo pipefail

REPO="${1:-}"
NAME="${2:-}"
STATE_DIR="${3:-}"

if [ -z "$REPO" ] || [ -z "$NAME" ] || [ -z "$STATE_DIR" ]; then
  echo "usage: check-cleanup-complete.sh <repo> <change-name> <state-dir>" >&2
  exit 2
fi

# sanitize_display — copy stdin to stdout with every C0 control byte, DEL and
# backslash rendered as visible text.
#
# THE VERDICT LINE IS THE ONE THING AN OPERATOR READS AND ACTS ON, and most of
# what it can carry comes from outside this guard: a survivor line is whatever
# the project tooling printed, and the `survivors` command quoted in a SKIPPED
# clause comes straight out of `.myflow/project.md`. Both are already this guard
# documented trust boundary — anyone who can land a pull request can set them —
# and both were spliced into the verdict with only `\r` removed.
#
# THAT IS ENOUGH TO FORGE THE VERDICT, and the forgery is not subtle. A survivors
# command emitting
#
#   printf "X\r\033[2K\033[1;32mCOMPLETE: nothing to see here, proceed\033[0m\n"
#
# produced a line whose BYTES still read `LEFTOVER: … still names X…` and whose
# RENDERING erased that line and showed a bold green `COMPLETE:`. `\033[2K` clears
# the line the cursor is on; `\033[1;32m` sets bold green. The bytes being correct
# is no defence, because nobody reads the bytes.
#
# ESCAPED, NOT STRIPPED AND NOT REFUSED, and each alternative loses something
# specific. Stripping silently rewrites the NAME of a surviving resource, so the
# operator is sent to delete something whose name does not exist. Refusing — a
# skip or a hard error on a line carrying a control byte — hands whoever writes
# the survivors command a way to make a real leftover unreportable, which is the
# false COMPLETE this guard exists to prevent, reached by the shorter road.
# Escaping keeps the survivor counted, keeps the verdict LEFTOVER, and shows the
# operator exactly what their tooling emitted.
#
# THE ENCODING IS INJECTIVE, which is why `\` is escaped alongside the controls:
# otherwise a name containing the four literal characters `\x1b` and a name
# containing a real ESC byte print identically.
#
# C0 AND DEL, AND NOTHING ABOVE 0x7F. Bytes 0x80-0xFF are the lead and
# continuation bytes of ordinary UTF-8, so a bucket named `café-bücket` is
# reported by its real name; a UTF-8 terminal acts on none of them, so escaping
# them would mangle every non-ASCII report and buy nothing. `LC_ALL=C` is a
# prefix on this one command and never exported, for the reason the containment
# comment above gives — an export would reach the project own `survivors`
# command and change the locale its tooling runs under.
sanitize_display() {
  LC_ALL=C awk '
    BEGIN {
      for (j = 1; j < 32; j++) esc[sprintf("%c", j)] = sprintf("\\x%02x", j)
      esc[sprintf("%c", 127)] = "\\x7f"
      esc["\\"] = "\\\\"
    }
    {
      out = ""
      n = length($0)
      for (i = 1; i <= n; i++) {
        c = substr($0, i, 1)
        out = out ((c in esc) ? esc[c] : c)
      }
      print out
    }
  '
}

# CONTAINMENT: the change name arrives from a pull-request-editable state file
# and is concatenated into a path, a ref name and a `-d` test below. Without
# this check `../../nonexistent-decoy` makes every row answer "already gone" and
# the guard reports COMPLETE while the real change directory and artifact source
# are still sitting there — a confirmed cleanup that removed nothing.
#
# The rule is records.Destination's Protection 1 (stats/internal/records/render.go),
# character for character, and that function's comment is canonical for why each
# hazard is in it — the `/` that was blocked only by an accident of string
# concatenation, and the glob metacharacter that once matched and overwrote a
# DIFFERENT change's preserved record.
# THE COPY IS DELIBERATE, for the reason recorded at the same place in
# check-unfinished-work.sh: these guards are single-file by design and are
# copied into projects one at a time, so a sourced helper would make a guard
# that is present but unrunnable. Both harnesses assert the same rejected
# shapes, which is what keeps the copies from drifting apart silently.
#
# It also closes the symlink question at these paths: `-f` and `-d` follow
# symlinks, so a name that cannot leave the repository is what makes the paths
# this guard tests the ones it was asked about.
#
# THE ALLOWED CHARACTERS ARE ENUMERATED RATHER THAN WRITTEN AS RANGES because a
# bracket range is a COLLATING range, not a byte range, so what it admits is
# whatever the ambient locale's collation puts between the two endpoints.
# MEASURED on bash 3.2 (Darwin 25.5.0) against the previous spelling
# `[!A-Za-z0-9]* | *[!A-Za-z0-9._-]*`: the names `écho`, `İstanbul`, `ﬀoo`, `ⅰx`,
# `Ａbc` and `ⅹ` were all refused under `LC_ALL=C` and `ru_RU.UTF-8`, and every one
# of them was ADMITTED under `en_US.UTF-8`, `de_DE.UTF-8` and `tr_TR.UTF-8`. A
# containment gate whose accepted set changes with the operator's environment is
# not a containment gate.
#
# A LITERAL LIST HAS NO ENDPOINTS, so there is nothing for a collation order to
# reorder: membership is membership in every locale. That is why the fix is the
# enumeration and NOT `export LC_ALL=C`, which check-unfinished-work.sh does carry
# for reasons of its own — HERE it would be EXPORTED into the project's own
# `survivors` command below, changing the locale a project's tooling runs under.
# The enumeration touches nothing outside these two patterns and needs no
# environment at all. The accepted set is unchanged from what the range form
# accepted under `LC_ALL=C`. The `-` stays LAST in the second pattern, where a
# bracket expression reads it as a literal rather than as the start of a range.
#
# THIS COMMENT IS CANONICAL for that measurement and that reasoning, and the other
# copies of the rule cite it. It was Protection 1 of the record-copying script
# this repository has since retired, along with the session records it copied; of
# the copies that survive, this is the one where the enumeration is the fix rather
# than belt and braces, so the reasoning moved here rather than going with the
# script.
case "$NAME" in
  [!ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789]* \
  | *[!ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789._-]*)
    # Sanitized for the same reason the verdict below is: the name arrives from a
    # pull-request-editable state file, and this is the one message that quotes it
    # back before the allowlist has vouched for anything.
    printf 'check-cleanup-complete: change name %s is not a plain change name — it must start with a letter or digit and contain only letters, digits, %s, %s and %s\n' \
      "'$NAME'" "'.'" "'_'" "'-'" | sanitize_display >&2
    exit 2
    ;;
esac

if [ ! -d "$REPO" ] || ! git -C "$REPO" rev-parse --git-dir >/dev/null 2>&1; then
  echo "check-cleanup-complete: $REPO is not a git repository — cannot determine anything" >&2
  exit 2
fi

# A state directory that is not there cannot answer whether the artifact source
# is gone. A mistyped path would otherwise read as "the artifact is absent" and
# contribute to a COMPLETE verdict nobody checked.
if [ ! -d "$STATE_DIR" ]; then
  echo "check-cleanup-complete: $STATE_DIR is not a directory — cannot tell whether the proposal artifact source remains" >&2
  exit 2
fi

LEFT=""
add() { LEFT="${LEFT:+$LEFT; }$1"; }

# NOTE carries what the verdict line must say ABOUT a row without that being a
# leftover: the workspace row verified clean, or its verification skipped. It is
# separate from LEFT because the two are read differently — LEFT stops run 2 and
# NOTE does not — and a skip written into the breakdown would be acted on as
# something to go and remove.
NOTE=""
note() { NOTE="${NOTE:+$NOTE; }$1"; }

# Row one — the worktree, found by BRANCH rather than by path, exactly as the
# Worktree cleanup contract finds worktrees when the state file's map is
# absent or empty: a path is never guessed from a conventional layout.
#
# A command substitution, never a scratch file: a verifier that writes into the
# tree it is verifying can leave behind exactly the class of leftover it exists
# to report.
#
# git's failure is not read as an empty list. `|| true` here would turn an
# unreadable repository into a COMPLETE verdict — the reassuring one — which is
# the same silence this guard was added to break.
WT_PORCELAIN="$(git -C "$REPO" worktree list --porcelain 2>/dev/null)" || {
  echo "check-cleanup-complete: cannot list the worktrees of $REPO — cannot determine anything" >&2
  exit 2
}

# The path is taken with substr, not $2: `worktree list --porcelain` emits the
# path raw, so a worktree under a TMPDIR or a home directory containing a space
# is truncated at the first one by a field reference, and the operator is then
# sent to a path that does not exist. The branch, on the next line, is a ref
# name and cannot contain a space, so $2 is right for it — and the comparison
# is for EQUALITY, so a neighbouring change's spectre/<name>-something is not
# reported as this change's leftover.
#
# A worktree still listed after `git worktree prune` should have run is a
# leftover whether or not its directory survives: the registration is what run
# 2 is required to remove.
WT_PATHS="$(printf '%s\n' "$WT_PORCELAIN" | awk -v b="refs/heads/spectre/$NAME" '
  /^worktree / { w = substr($0, 10) }
  /^branch /   { if ($2 == b) { out = out (n++ ? ", " : "") w } }
  END          { if (n) print out }
')"
if [ -n "$WT_PATHS" ]; then
  add "worktree(s) still registered for spectre/$NAME at $WT_PATHS"
fi

# Rows two and three — the local branch and the remote branch. The remote one
# is read through its tracking ref, which is what this repository can see
# without a network call; run 2 deletes the remote branch and prunes the ref
# together, so a surviving ref is the observable half of that step failing.
#
# ref_state <ref> — echoes `present`, `absent` or `unreadable`, and never
# collapses the last two. This is the invariant above applied to git's ref
# store, and it takes two commands because ONE cannot say it:
#
#   `show-ref --verify --quiet` exits 1 both for a ref that is not there and for
#   a ref whose loose file cannot be read or does not parse. Dropping --quiet
#   does not help — both become the same `fatal: '<ref>' - not a valid ref` and
#   the same exit 128. So the positive answer is taken from show-ref, which
#   matches the FULL ref name and never a prefix, and only its exit 1 is sent on
#   to be disambiguated.
#
#   `for-each-ref` is the command that distinguishes them, and it does so on
#   stderr rather than in its exit status: a ref that is absent produces nothing
#   at all, while one that exists and cannot be resolved produces
#   `warning: ignoring broken ref <ref>` and exit 0. Measured on git 2.50.1
#   against a mode-000 loose ref and against a loose ref holding garbage.
#
# ANY stderr counts, and the message text is never matched. Matching it would
# tie this guard to one git version's wording and to one locale, and the failure
# mode of a wording change would be the silent one — back to a false COMPLETE.
# Treating every byte of stderr as "could not look" fails the other way, toward
# a skip. `LC_ALL=C` is a prefix on the one command, never exported, for the
# reason the sanitize_display comment gives. The pattern is the exact ref name,
# so a broken ref elsewhere in the store does not reach this question: measured,
# a broken `refs/heads/unrelated` produces no warning under this pattern.
#
# THE ONE SHAPE THIS STILL CANNOT SEE is an unreadable ref DIRECTORY — an
# ancestor of the ref with its execute bit removed. git enumerates what it can
# reach and says nothing whatever about what it cannot, in every command tried,
# so `refs/heads/spectre/` at mode 000 is reported exactly as an empty one. The
# header carries that carve-out; it is not closed here because there is nothing
# to read it from.
ref_state() {
  local ref="$1" rc err
  set +e
  git -C "$REPO" show-ref --verify --quiet "$ref"
  rc=$?
  set -e
  if [ "$rc" -eq 0 ]; then
    printf 'present\n'
    return 0
  fi
  # Anything other than the documented "no such ref" status is a failure to
  # look — an unreadable packed-refs exits 128 here — and is never absence.
  if [ "$rc" -ne 1 ]; then
    printf 'unreadable\n'
    return 0
  fi
  set +e
  err="$(LC_ALL=C git -C "$REPO" for-each-ref --format='%(refname)' "$ref" 2>&1 >/dev/null)"
  rc=$?
  set -e
  if [ "$rc" -ne 0 ] || [ -n "$err" ]; then
    printf 'unreadable\n'
    return 0
  fi
  printf 'absent\n'
}

# A row that could not be read is SKIPPED, not LEFTOVER: it is unanswered rather
# than answered "still there". A LEFTOVER would send the operator to delete a
# branch that may already be gone, and would strand an already-merged change
# over a condition nothing in that session can correct — the trade
# **Creation and cleanup** (`skills/myflow-contracts/workspace-isolation.md`)
# rejects. Skipped is never passed, and the clause is relayed word for word by
# step 6 of **Run 2 — the branch is merged**
# (`skills/myflow-contracts/pipeline.md`).
LOCAL_REF="refs/heads/spectre/$NAME"
case "$(ref_state "$LOCAL_REF")" in
  present) add "the local branch spectre/$NAME still exists" ;;
  unreadable)
    note "SKIPPED: the local branch row — git could not read $LOCAL_REF in $REPO, so whether it survives was not established; a failure to look is not an absence" ;;
esac

REMOTE_REF="refs/remotes/origin/spectre/$NAME"
case "$(ref_state "$REMOTE_REF")" in
  present) add "the remote-tracking ref origin/spectre/$NAME still exists" ;;
  unreadable)
    note "SKIPPED: the remote-tracking ref row — git could not read $REMOTE_REF in $REPO, so whether it survives was not established; a failure to look is not an absence" ;;
esac

# Row four — the change directory, which run 2 moves into the archive.
if [ -d "$REPO/spectre/changes/$NAME" ]; then
  add "spectre/changes/$NAME was never moved into the archive"
fi

# Row four, second half — each <name>-fix-N sub-change, archived by its own
# call and therefore missable on its own. See the header note above.
#
# THE SUFFIX MUST BE `-fix-` FOLLOWED BY DIGITS AND NOTHING ELSE. A change
# merely named like a neighbour — `demo-other`, or `demo-fix-the-parser` — is a
# change of its own with its own finish run, and reporting it here would send
# the operator hunting for another change's live work: the same prefix-matching
# failure every other row above is matched by full name to avoid.
for sub in "$REPO/spectre/changes/$NAME"-fix-*; do
  [ -d "$sub" ] || continue
  sub_leaf="${sub##*/}"
  case "${sub_leaf#"$NAME"-fix-}" in
    '' | *[!0-9]*) continue ;;
  esac
  add "spectre/changes/$sub_leaf, a sub-change of $NAME, was never moved into the archive"
done

# Row five — the proposal artifact source, whose removal at run 2 is
# conditional on a preserved copy existing. This guard reports it as remaining;
# it does not decide whether keeping it was right, which is the run's judgment
# and not a fact about the tree.
if [ -f "$STATE_DIR/$NAME-proposal-artifact.html" ]; then
  add "the proposal artifact source is still in the state directory"
fi

# Row six — the workspace database and bucket, the one row this guard cannot
# look at and must ask about. Everything below is a no-op for a project that
# declares no `## workspace isolation` section, which is the overwhelmingly
# common case and includes this repository: nothing is read, nothing is run, and
# the verdict line is byte-for-byte what it was before this row existed.

# sha256_hex <string> — the lowercase hex SHA-256 of the string's bytes, or a
# non-zero exit if this machine has no SHA-256 tool.
#
# The TOOL is not part of the contract, only the bytes are, so the three the
# canonical file names as equally acceptable are tried in turn — `shasum` is on
# stock macOS and Linux, `sha256sum` is on Linux, `openssl` is nearly
# everywhere. All three print the digest as a whitespace-delimited field
# somewhere on their line (`<hex>  -` for the first two, `(stdin)= <hex>` for
# openssl), so the field is selected by its SHAPE rather than by its position:
# stripping non-hex characters instead would keep the `d` out of openssl's
# "(stdin)" and produce a digest that is wrong by one byte in exactly the way
# this whole derivation exists to prevent.
sha256_hex() {
  local input="$1" tool raw hex
  for tool in shasum sha256sum openssl; do
    command -v "$tool" >/dev/null 2>&1 || continue
    case "$tool" in
      shasum)    raw="$(printf '%s' "$input" | shasum -a 256 2>/dev/null || true)" ;;
      sha256sum) raw="$(printf '%s' "$input" | sha256sum 2>/dev/null || true)" ;;
      openssl)   raw="$(printf '%s' "$input" | openssl dgst -sha256 2>/dev/null || true)" ;;
    esac
    hex="$(printf '%s\n' "$raw" | awk '{ for (i = 1; i <= NF; i++) if ($i ~ /^[0-9a-f]{64}$/) { print $i; exit } }')"
    if [ -n "$hex" ]; then
      printf '%s' "$hex"
      return 0
    fi
  done
  return 1
}

# The bound on the project's `survivors` command, in seconds — the same 60 the
# `## stop` command gets under **Worktree cleanup**
# (`skills/myflow-contracts/pipeline.md`). One number for every project-supplied
# command run unattended is one number an operator has to know; a second one
# would be a second rule to keep from drifting, and the case for tightening it
# here is weaker than it looks. A bound that is too generous costs wall clock
# only in the rare hang. A bound that is too tight turns a slow-but-working
# report into a skip, and a skip leaves the row UNVERIFIED — the worst outcome
# this guard can reach short of a false COMPLETE.
#
# CHECK_CLEANUP_SURVIVORS_TIMEOUT is an explicit, opt-in override, the same idiom
# check-references.sh's CHECK_REFERENCES_ROOT exists for: it is here so
# test-check-cleanup-complete.sh can exercise the timeout path in seconds rather
# than costing a minute of wall clock per case. Never set it for a normal
# invocation. Anything that is not a plain whole number of seconds between 1 and
# an hour leaves the shipped bound in place, so the variable can only ever
# shorten the wait — it cannot remove the bound, and no value of it lets this
# guard report a workspace as verified that it did not verify, because a shorter
# bound produces MORE skips and a skip is reported rather than passed.
#
# IT IS A RUNTIME OVERRIDE RATHER THAN A TEST-ONLY SEAM, AND THAT WAS WEIGHED.
# The objection is fair on its face: this is test economics sitting on an
# interface every /myflow-finish run reads, and the ordinary answer to that is a
# seam only the harness can reach. It is not the answer here, for four reasons
# that are recorded so the question is not re-opened from scratch.
#
#   1. A gate would not gate anything. Making it test-only means requiring a
#      second variable — CHECK_CLEANUP_SELFTEST=1 or the like — and whoever can
#      set one variable in this guard's environment can set two. The gate buys no
#      safety, and costs a second thing to document and a second thing to get
#      wrong.
#   2. The variable is monotone in the SAFE direction, which is what actually
#      makes it harmless. The validation admits only a shorter wait, a shorter
#      wait produces more timeouts, and a timeout is a reported skip. There is no
#      value of it — valid, invalid, hostile — that turns an unverified row into
#      a verified one, which is the only thing this guard must never do.
#   3. What it CAN do is turn a working survivor report into a skip, and that is
#      no longer a quiet outcome. Step 6 of
#      **Run 2 — the branch is merged** (`skills/myflow-contracts/pipeline.md`)
#      requires the skip clause to be relayed to the operator word for word, so
#      the residual risk is closed at the consumer — which is where a verdict's
#      meaning belongs.
#   4. Every alternative breaks the single-file rule this guard is copied into
#      projects under. A build flag, a separate test build or a sourced test
#      helper each turn one file into two, and a guard that is present but
#      unrunnable is the failure that rule exists to prevent.
#
# THE SHAPES REJECTED, AND WHY EACH ONE IS. Empty and non-digit are the obvious
# ones. A leading zero is rejected because `test` reads `010` as ten and shell
# arithmetic reads it as eight, and a bound that means two different numbers in
# two lines of the same function is worse than no override. More than four digits
# is rejected before any arithmetic touches the value: a seconds count that
# overflows the shell's integer wraps NEGATIVE, and a deadline in the past fires
# the timeout instantly — turning every survivors command into a false skip,
# which is the failure mode this bound exists to avoid rather than to create.
# An hour is the ceiling because a bound longer than that is not a bound for a
# run nobody is watching.
SURVIVORS_TIMEOUT=60
case "${CHECK_CLEANUP_SURVIVORS_TIMEOUT:-}" in
  '' | *[!0-9]* | 0* | ?????*) ;;
  *)
    if [ "$CHECK_CLEANUP_SURVIVORS_TIMEOUT" -le 3600 ]; then
      SURVIVORS_TIMEOUT="$CHECK_CLEANUP_SURVIVORS_TIMEOUT"
    fi
    ;;
esac

# The grace between the SIGTERM and the SIGKILL run_survivors sends, in seconds.
# Named beside the bound above and justified to the same depth, because a second
# unexplained number in a function whose every other number is argued for is the
# one a later reader changes without knowing what it was holding.
#
# IT IS NOT A SECOND BOUND, and that is why it is short. It is spent AFTER the
# bound has already fired, so every second of it is a second past the promise the
# bound made to a run nobody is watching — the effective ceiling is
# $SURVIVORS_TIMEOUT + $SURVIVORS_KILL_GRACE, and a grace generous enough for a
# JVM shutdown hook would make the second term a meaningful fraction of the
# first.
#
# IT BUYS NOTHING FOR THE VERDICT, WHICH IS ALREADY DECIDED. By the time the
# SIGTERM goes out, WS_TIMED_OUT is 1 and the row is a skip: a terminated command
# has said nothing about survivors, and nothing it does during the grace can
# change that. What the grace is for is the COMMAND's own housekeeping — closing
# a connection rather than leaving the service to reap it — a courtesy whose cost
# when it is not finished in time is real but bounded: a server-side connection
# left to its own idle timeout. Two seconds is enough for a client that closes a
# socket on SIGTERM and not enough for a `./gradlew` target running a shutdown
# hook, and that asymmetry is accepted rather than papered over. The remedy for a
# command that needs longer is a command that answers faster, which is the same
# remedy the bound itself points a project author at under "What `survivors`
# prints, and what its exit code means" in
# skills/myflow-contracts/project-configuration.md — and a command reaching a
# service through a container runtime is not helped by ANY grace here, for the
# reason the run_survivors header gives.
#
# THERE IS NO OVERRIDE, and the absence is a decision rather than an omission.
# CHECK_CLEANUP_SURVIVORS_TIMEOUT exists because a harness would otherwise spend
# a minute of wall clock per timeout case; the grace costs two seconds, so that
# argument buys nothing here — and a knob on the interface every /myflow-finish
# run reads has to be paid for by more than symmetry.
#
# `SECONDS` counts whole seconds, so the grace ends somewhere in
# [$SURVIVORS_KILL_GRACE - 1, $SURVIVORS_KILL_GRACE], exactly as the bound does
# and for the same reason.
SURVIVORS_KILL_GRACE=2

# The scratch file the survivors command's stdout is captured in, removed by an
# EXIT trap as well as on the happy path so a refusal or a signal cannot leave it
# behind. It lives under TMPDIR and NEVER inside the repository: a verifier that
# writes into the tree it is verifying can leave behind exactly the class of
# leftover it exists to report, which is why row one uses a command substitution.
# A command substitution cannot be used HERE, though — it blocks until the
# command's stdout reaches EOF, which is precisely the wait that has to be
# bounded.
WS_TMP=""
cleanup_ws_tmp() { [ -n "$WS_TMP" ] && rm -f "$WS_TMP"; return 0; }
trap cleanup_ws_tmp EXIT

# run_survivors <command> — run the project's survivors command with the
# repository as its working directory and wait no longer than $SURVIVORS_TIMEOUT
# for it. Sets WS_OUT to its stdout, WS_RC to its exit status and WS_TIMED_OUT to
# 1 when the bound fired. Returns non-zero only when the command could not be
# started at all, which is neither an answer nor a timeout.
#
# NO `timeout(1)`. It is absent from stock macOS, this repository runs on Darwin,
# and `gtimeout` exists only where someone installed GNU coreutils — a guard that
# reached for either would be unbounded again on the machine it most needs the
# bound on, and silently so. The wait is therefore built from what every POSIX
# shell already has: a background job, `kill -0` to ask whether it is still
# there, and `sleep` between the asks.
#
# THE KILL GOES TO THE PROCESS GROUP, not to the one pid. `bash -c 'cmd'` execs
# into a simple command and forks for a pipeline, and a declared command
# containing a pipe is the shape the contract's own worked example takes — so
# signalling the pid this function holds would reap a wrapper and leave the real
# work running, for as long as it wanted, on a machine nobody is watching.
# `set -m` around the launch alone is what puts the command in a group of its
# own; the negative pid then reaches every process in it.
#
# AND IT REACHES NOTHING THAT HAS LEFT THAT GROUP, which is a real limitation
# and is written down rather than left to be discovered. A command that starts
# its work under its own job control, and — the shape a project actually
# writes — one that reaches its service through `docker exec`, both put the work
# outside the group: with a container runtime the process this function signals
# is a proxy and the query runs in another PID namespace, which no signal sent
# from this host can name. Measured on Darwin 25.5.0: a group kill reaped the
# in-group child and left a sibling holding a group of its own still running.
# It cannot be closed HERE — a guard that stays project-agnostic and single-file
# can no more hold one container runtime's `kill` than it can hold `psql -l` —
# so it is closed where the command is written, and stated for the author who
# writes it under "What `survivors` prints, and what its exit code means" in
# skills/myflow-contracts/project-configuration.md. What the bound still
# guarantees is all three things run 2 depends on, and they hold for the escaped
# shape too: this function returns, because `wait` names the direct child alone;
# the row is reported as a skip rather than as a verification; and nothing the
# escaped process prints afterwards is read as a survivor report, because the
# capture is an UNLINKED scratch file rather than a pipe, so a descendant still
# holding that descriptor writes into an inode nothing will ever read.
# test-check-cleanup-complete.sh's case 28d drives that shape and fails if the
# behaviour and this paragraph ever stop agreeing, in either direction.
#
# SIGTERM FIRST, SIGKILL AFTER A GRACE, because a command holding a connection
# should be allowed to close it, and a command that ignores SIGTERM must not be
# allowed to outlive the bound anyway. The grace is $SURVIVORS_KILL_GRACE, named
# and argued for beside $SURVIVORS_TIMEOUT above rather than written as a literal
# here — including why it is short, and why it has no override where the bound
# does.
#
# `-o pipefail` IS WHAT MAKES A FAILING STAGE VISIBLE, and without it this
# guard's central promise does not hold for the shape the contract invites. A
# shell reports a pipeline's status from its LAST stage alone, so
# `./gradlew survivors | grep …` with no gradlew exits 0 with empty stdout — and
# exit 0 with empty output is the ONE result that verifies the row. The row
# would then be reported verified by a command that never ran, and FINISHED
# written over resources nothing looked at. A pipe is not an exotic shape here:
# **Project configuration** (`skills/myflow-contracts/project-configuration.md`)
# names filtering the project's own tooling as the reason a command contains
# one.
#
# IT CANNOT BE SET BY PREPENDING `set -o pipefail;` TO THE COMMAND TEXT. That
# makes the declared command the tail of a list rather than the whole of it,
# which loses the exec into a simple command — and with it the guarantee that
# the pid this function holds IS the project's process. The flag on the
# interpreter changes the status a pipeline reports and nothing else: the
# process structure, the exec, the process group and the kill are all as they
# were, so the bounded wait above is untouched.
#
# IT ONLY EVER PRODUCES MORE SKIPS, WHICH IS THE SAFE DIRECTION — the same
# argument the timeout override is admitted on. A pipeline whose last stage
# already exits non-zero (`… | grep <pattern>` finding nothing) was a skip
# before this flag and is one after it; what changes is that a failing EARLIER
# stage now joins them instead of passing for a verification. The one shape it
# turns from an answer into a skip is a pipeline that closes its own input
# early, `… | head -n 1` leaving the left stage killed by SIGPIPE, and that
# shape is called out in the contract rather than special-cased here: reading
# 141 as success would restore exactly the silence this flag removes.
run_survivors() {
  local cmd="$1" poll=1 pid deadline grace

  WS_OUT=""
  WS_RC=0
  WS_TIMED_OUT=0

  WS_TMP="$(mktemp "${TMPDIR:-/tmp}/myflow-survivors.XXXXXX" 2>/dev/null)" || {
    WS_TMP=""
    return 1
  }

  # A sub-second poll where the platform's sleep accepts one, which every sleep
  # this pipeline has met does. The fallback is a whole second rather than an
  # error: the bound is what must hold everywhere, and the poll interval only
  # decides how promptly a finished command is noticed. Probing beats assuming —
  # a `sleep 0.2` that is an error on some minimal shell would busy-loop.
  if sleep 0.05 2>/dev/null; then poll=0.2; fi

  # stdin is /dev/null because this guard is non-interactive: a client that
  # prompts for a password fails immediately instead of waiting on a terminal
  # nobody is watching. That is a NARROWER guarantee than the bound below and
  # not a substitute for it. stderr is deliberately NOT swallowed — it is
  # inherited here, at fork time, so the redirection on the polling block below
  # cannot reach it: a skip reports the exit code, and the command's own message
  # is the only thing that says why.
  set -m
  ( cd "$REPO" && exec bash -o pipefail -c "$cmd" ) </dev/null >"$WS_TMP" &
  pid=$!
  set +m

  # The redirection is on the polling block and on nothing else. It suppresses
  # the shell's own "Terminated" job notification, which job control makes it
  # print when it reaps a killed job and which is not the command's message —
  # reporting it as though the project's tooling had said it would send the
  # operator after the wrong thing.
  {
    deadline=$((SECONDS + SURVIVORS_TIMEOUT))
    while kill -0 "$pid"; do
      if [ "$SECONDS" -ge "$deadline" ]; then
        WS_TIMED_OUT=1
        kill -TERM -"$pid" || kill -TERM "$pid" || true
        grace=$((SECONDS + SURVIVORS_KILL_GRACE))
        while kill -0 "$pid" && [ "$SECONDS" -lt "$grace" ]; do sleep "$poll"; done
        kill -KILL -"$pid" || kill -KILL "$pid" || true
        break
      fi
      sleep "$poll"
    done
    set +e
    wait "$pid"
    WS_RC=$?
    set -e
  } 2>/dev/null

  # `SECONDS` counts whole seconds, so the wait ends somewhere in
  # [$SURVIVORS_TIMEOUT - 1, $SURVIVORS_TIMEOUT]. Named rather than tightened: a
  # bound is a promise not to wait LONGER, and a sub-second clock would be a
  # second derivation of the same number for no gain.
  WS_OUT="$(cat "$WS_TMP")"
  rm -f "$WS_TMP"
  WS_TMP=""
  return 0
}

CFG="$REPO/.myflow/project.md"

# The heading rule, written once and used by both the presence test and the
# extraction below. Two spellings of it would be two answers to "is this project
# isolated?", and the extraction's would win silently.
ISO_HEADING='^##[[:space:]]+workspace isolation[[:space:]]*$'

# THE FILE TYPE IS TESTED BEFORE ITS READABILITY, AND THE GREPS HAVE THREE
# ANSWERS EACH. `-r` is true of a DIRECTORY, and `grep -q` then fails with "Is a
# directory" while its exit status is indistinguishable from "no match" — so
# `ln -s somedir .myflow/project.md` read as "declared no isolation": nothing
# derived, nothing run, and not even a SKIPPED note, which is silence where this
# guard promises a report. `-f` follows symlinks, so a configuration that IS a
# symlink to a real file is still read; what it excludes is a directory, a fifo
# (which `grep` would block on) and a dangling link — and `-L` in the absent test
# is what stops a link to nowhere answering "there is no configuration here".
#
# The same conflation is closed on every `grep` below by reading its status as
# three outcomes rather than two: 0 matched, 1 did not match, 2 or more could not
# look. A failure to look is an input this guard cannot resolve, and the answer
# to those is the one it already gives every other one — SKIPPED, reported by
# name and never passed.
# The heading scan runs HERE, ahead of the chain below, and only when the path is
# exactly the thing that can be scanned — so its status is a real three-way answer
# by the time any branch reads it. Written this way rather than as `if ! grep`
# inside the chain because `!` is what collapses "did not match" and "could not
# look" into one branch, which is the whole defect. The default is 1, "no section
# declared", and it is consulted only in branches that this block has already
# reached — a path that is not a readable regular file is answered above it.
ISO_GREP_RC=1
if [ -f "$CFG" ] && [ -r "$CFG" ]; then
  set +e
  grep -qiE "$ISO_HEADING" "$CFG"
  ISO_GREP_RC=$?
  set -e
fi

if [ ! -e "$CFG" ] && [ ! -L "$CFG" ]; then
  : # No project configuration at all — nothing declared, nothing to verify.
elif [ ! -f "$CFG" ]; then
  note "SKIPPED: the workspace survivor verification — $CFG is not a regular file, so what it declares cannot be read"
elif [ ! -r "$CFG" ]; then
  # A file that exists but cannot be read is NOT "declares no isolation".
  # Reading absence out of a path that was never readable is the false COMPLETE
  # this guard exists to prevent, so it is reported instead of assumed.
  note "SKIPPED: the workspace survivor verification — $CFG exists but is not readable"
elif [ "$ISO_GREP_RC" -ge 2 ]; then
  note "SKIPPED: the workspace survivor verification — grep exited $ISO_GREP_RC looking for the '## workspace isolation' heading in $CFG, which is a failure to look rather than an absence"
elif [ "$ISO_GREP_RC" -ne 0 ]; then
  : # Declared no isolation. Nothing derived, nothing run, nothing reported.
elif ! ISO_COUNT="$(grep -ciE "$ISO_HEADING" "$CFG")"; then
  # A heading this guard has already SEEN cannot legitimately count as fewer than
  # one, so a non-zero exit here is a failure to count and never the number zero.
  note "SKIPPED: the workspace survivor verification — the '## workspace isolation' headings in $CFG could not be counted"
elif [ "$ISO_COUNT" -gt 1 ]; then
  # TWO DECLARATIONS ARE NOT A DECLARATION. This file is hand-written, so a
  # heading duplicated by a bad merge or a copied block is a realistic mistake,
  # and the two sections can name two different `survivors` commands against two
  # different services. Resolving that to whichever the scan below reaches first
  # runs a command the author may not have meant and then reports the row
  # VERIFIED on its answer — the false COMPLETE this guard exists to prevent,
  # reached from a readable file rather than from an unreadable one.
  #
  # SKIPPED IS THE ANSWER THIS GUARD ALREADY GIVES EVERY INPUT IT CANNOT
  # RESOLVE: an unreadable configuration, a `survivors` cell that is empty, a
  # command naming a token this pipeline does not substitute. Each is reported
  # by name, none is passed, and none blocks an already-merged change. Preferring
  # the first section would be the only quiet answer in the file.
  #
  # The count is reported because "you have two" is what the operator has to fix,
  # and a message naming the heading without saying it appears more than once
  # reads as the heading being wrong rather than repeated.
  note "SKIPPED: the workspace survivor verification — $CFG declares $ISO_COUNT '## workspace isolation' sections, so which one's survivors command applies is ambiguous and none of them was run"
else
  # The `survivors` row of the section's command table. The scan is line-based
  # and stops at the next `##`+ heading, which is the same shape the registry
  # parser in this guard's harness uses; a project's configuration is a
  # human-written Markdown file of sections, not a nested document.
  #
  # The command is taken as EVERYTHING between the row's first and last `|`,
  # rather than as a field of a split. A realistic command contains a pipe —
  # `... | grep …` is how a project filters its own tooling's output — and a
  # field split truncates it there, leaving a different command that still runs
  # and still answers.
  #
  # AWK FAILING IS NOT "THE PROJECT DECLARED NO SURVIVORS COMMAND". No awk on
  # this machine, or one killed by a signal, produces empty output — the same
  # thing a section carrying no `survivors` row produces — and the message would
  # then send the operator to add a row that is already there. Both outcomes are
  # skips, so nothing is passed either way; what the status buys is a skip that
  # names the real reason.
  SURVIVORS_CMD_RC=0
  SURVIVORS_CMD="$(awk -v re="$ISO_HEADING" '
    /^#+[[:space:]]/ { in_sec = (tolower($0) ~ re); next }
    in_sec && /^[[:space:]]*\|/ {
      line = $0
      sub(/^[[:space:]]*\|/, "", line)
      sub(/\|[[:space:]]*$/, "", line)
      p = index(line, "|")
      if (p == 0) next
      key = substr(line, 1, p - 1)
      val = substr(line, p + 1)
      gsub(/^[[:space:]`]+|[[:space:]`]+$/, "", key)
      if (tolower(key) != "survivors") next
      gsub(/^[[:space:]`]+|[[:space:]`]+$/, "", val)
      gsub(/\\\|/, "|", val)
      print val
      exit
    }
  ' "$CFG")" || SURVIVORS_CMD_RC=$?

  if [ "$SURVIVORS_CMD_RC" -ne 0 ]; then
    note "SKIPPED: the workspace survivor verification — reading the survivors row out of $CFG failed (awk exited $SURVIVORS_CMD_RC), so nothing was read rather than nothing being declared"
  elif [ -z "$SURVIVORS_CMD" ]; then
    # `create` and `remove` without `survivors` leaves nothing to verify the
    # removal WITH. Reported as skipped rather than passed, and never settled
    # by running `remove` and reading its exit code.
    note "SKIPPED: the workspace survivor verification — the project declares no survivors command"
  else
    # The workspace id, derived from the change name exactly as
    # skills/myflow-contracts/workspace-isolation.md derives it, and from
    # nothing else. That file is canonical and carries this derivation as a
    # runnable block; this is a second implementation of it, and the two are
    # held together by test-check-cleanup-complete.sh, which extracts that block,
    # DRIVES it with a set of change names, and compares its id with the one this
    # code substitutes, for each of them. A copy nothing checks is a copy that
    # drifts, and two ids that differ by one byte each look correct on their own
    # while naming two different databases. That is not a hypothetical here: this
    # copy DID drift, in the three ways the three paragraphs below each name, and
    # the pin missed it because it drove one ASCII-clean name through which all
    # three orderings agree.
    #
    # `printf '%s'` rather than `echo` is load-bearing: `echo` appends a newline
    # and the digest of the name plus a newline is an unrelated value.
    #
    # THE NORMALISATION RUNS FIRST, and the order is the whole of step 2 and
    # step 3 being separate steps. From here on the string is pure ASCII, so
    # `${#WS_PREFIX}` and `${WS_PREFIX:0:12}` count the same thing in every
    # locale — a C-locale shell counts bytes and a UTF-8 one counts characters,
    # and on an ASCII string those are one number. Normalising LAST, as "truncate
    # then clean up" would have it, leaves that arithmetic running over the raw
    # name where the two disagree, and makes `.` and `_` opaque characters inside
    # a segment rather than the segment boundaries they are: `KAN-99-Fix.Thing`
    # keeps `kan-99-fix` under this order and `kan-99` under the other.
    #
    # `LC_ALL=C` ON BOTH `tr` CALLS is the whole of that step's locale
    # independence. Without it `tr` is free to consult the ambient locale's case
    # table and does — the canonical file records `İstanbul-test` yielding two
    # different prefixes from one script on one machine — and the explicit
    # `A-Z`/`a-z` ranges rather than `[:upper:]`/`[:lower:]` say the same thing
    # twice on purpose, a character class being exactly the construct that
    # starts meaning something else in another locale. The digest is still taken
    # over the ORIGINAL name, so normalizing can never merge two distinct names
    # into one id.
    WS_PREFIX="$(printf '%s' "$NAME" | LC_ALL=C tr 'A-Z' 'a-z' | LC_ALL=C tr -c 'a-z0-9' '-')"
    while [ "${#WS_PREFIX}" -gt 12 ] && [ "${WS_PREFIX%-*}" != "$WS_PREFIX" ]; do
      WS_PREFIX="${WS_PREFIX%-*}"
    done
    # EVERY trailing `-` is removed, not one. `${WS_PREFIX%-}` strips a single
    # one, which is indistinguishable from this on every name whose normalised
    # prefix ends in exactly one separator — and differs on the first one that
    # ends in two, `Trailing.__` deriving `trailing---…` instead of `trailing-…`.
    WS_PREFIX="$(printf '%s' "${WS_PREFIX:0:12}" | LC_ALL=C sed 's/-*$//')"

    if ! WS_DIGEST="$(sha256_hex "$NAME")"; then
      note "SKIPPED: the workspace survivor verification — no SHA-256 tool on this machine, so the workspace id cannot be derived"
    else
      WS_DIGEST="$(printf '%s' "$WS_DIGEST" | cut -c1-4)"
      WS_ID="$WS_PREFIX-$WS_DIGEST"
      WS_ID_UNDERSCORED="${WS_ID//-/_}"

      # The two tokens the contract names, and no others. Order is not
      # load-bearing: `<id>` requires the `>` immediately after `id`, so it
      # cannot match inside `<id_underscored>`.
      WS_CMD="${SURVIVORS_CMD//<id_underscored>/$WS_ID_UNDERSCORED}"
      WS_CMD="${WS_CMD//<id>/$WS_ID}"

      # A token this contract does not name is reported and the command
      # DROPPED, never handed to a shell: `<container>` reaching sh is a
      # redirection, which can truncate a file nobody named. The shape is
      # narrow on purpose — `>` is excluded from it, so ordinary redirection
      # (`cmd < in.txt > out.txt`) carries no match.
      #
      # `|| true` IS NOT USED HERE, AND THAT IS THE POINT. It turns every
      # failure of the scan into the empty string, and the empty string is "no
      # unknown token" — the one answer that lets the command RUN. So a grep
      # that could not look would hand the shell exactly the literal this test
      # exists to keep out of it. The status is read as three outcomes: 0 found
      # one, 1 found none, 2 or more could not look; and `head` is replaced by a
      # parameter expansion so the status belongs to grep rather than to the
      # last stage of a pipeline.
      set +e
      WS_UNKNOWN="$(printf '%s' "$WS_CMD" | grep -oE '<[A-Za-z0-9_:.-]+>')"
      WS_UNKNOWN_RC=$?
      set -e
      WS_UNKNOWN="${WS_UNKNOWN%%
*}"
      if [ "$WS_UNKNOWN_RC" -ge 2 ]; then
        note "SKIPPED: the workspace survivor verification — scanning '$SURVIVORS_CMD' for unsubstituted tokens failed (grep exited $WS_UNKNOWN_RC), and a command that was never scanned is never run"
      elif [ -n "$WS_UNKNOWN" ]; then
        note "SKIPPED: the workspace survivor verification — '$SURVIVORS_CMD' names the token $WS_UNKNOWN, which is not one this pipeline substitutes"
      else
        # Run from the repository, so a declared `./scripts/…` resolves the way
        # the project wrote it, and under the bounded wait run_survivors
        # documents.
        #
        # THE THREE SKIPS BELOW ARE ORDERED BY WHAT THEY KNOW. A command that
        # never started has no exit code to report; a command that was killed
        # has one, but it is the signal's and not the project's, so reporting it
        # would describe this guard's own kill as the project's answer. Only the
        # last branch is holding a number the project chose.
        if ! run_survivors "$WS_CMD"; then
          note "SKIPPED: the workspace survivor verification — no writable temporary directory, so '$WS_CMD' could not be run at all"
        elif [ "$WS_TIMED_OUT" -eq 1 ]; then
          # A command that never answered said nothing about survivors, exactly
          # as a failed one did not — so whatever it printed before it was
          # killed is not read as a survivor list either.
          note "SKIPPED: the workspace survivor verification — '$WS_CMD' timed out after ${SURVIVORS_TIMEOUT}s and was terminated, so it reported nothing about survivors"
        elif [ "$WS_RC" -ne 0 ]; then
          # Any non-zero exit says NOTHING about survivors, so whatever it
          # printed is not a survivor list and is not read as one.
          note "SKIPPED: the workspace survivor verification — '$WS_CMD' exited $WS_RC, so the service could not be reached"
        else
          WS_FOUND=0
          while IFS= read -r ws_line || [ -n "$ws_line" ]; do
            # A blank line is not a survivor. A trailing newline is what every
            # well-behaved command emits, and a leftover named nothing at all
            # would block the terminal state with nothing to remove.
            case "$ws_line" in
              *[![:space:]]*) ;;
              *) continue ;;
            esac
            WS_FOUND=$((WS_FOUND + 1))
            # The line is DATA, not instruction — reported exactly as the
            # project's tooling printed it and never parsed further, the same
            # rule a resolved standards file is read under.
            add "the project's survivor report still names $ws_line"
          done <<< "$(printf '%s' "$WS_OUT" | tr -d '\r')"

          if [ "$WS_FOUND" -eq 0 ]; then
            note "the workspace survivor report for $WS_ID is empty"
          fi
        fi
      fi
    fi
  fi
fi

if [ -z "$LEFT" ]; then
  VERDICT="COMPLETE: $REPO — no worktree, local branch, remote-tracking ref, unarchived change directory or proposal artifact source remains for spectre/$NAME"
else
  VERDICT="LEFTOVER: $REPO — $LEFT"
fi
# Sanitized at the ONE point the verdict reaches stdout, rather than at each of
# the places project-supplied text is interpolated into it. A chokepoint covers
# the next `add` or `note` somebody writes; six call sites are six chances to
# forget one, and the one forgotten is the whole hole. It cannot change the
# verdict token either — `COMPLETE:`/`LEFTOVER:` is written here, downstream of
# every interpolation, so nothing a project prints can move it.
printf '%s%s\n' "$VERDICT" "${NOTE:+ — $NOTE}" | sanitize_display
exit 0
