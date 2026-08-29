#!/usr/bin/env bash
# check-unfinished-work.sh — report whether a change carries unfinished work,
# for /myflow-finish's run-1 gate.
#
# Usage: check-unfinished-work.sh <worktree> <change-name> [canonical-worktree]
#
# Prints ONE verdict line to stdout:
#   CLEAR: <reason>          nothing outstanding
#   OUTSTANDING: <breakdown> one or more signals fired
#
# Exit 0 whenever a verdict was reached; exit 2 when it cannot answer at all —
# an unreadable worktree, a change name outside the allowlist, or a file that
# exists but cannot be read. The VERDICT carries the answer, not the exit
# status — see check-finish-preflight.sh's header for why this repository
# separates them. Finish runs this once per worktree, so each verdict names the
# worktree it judged; a bare CLEAR would be unattributable across several.
#
# WHY THE VERDICT WORDS DIFFER FROM check-cleanup-complete.sh's. That guard says
# COMPLETE/LEFTOVER, this one says CLEAR/OUTSTANDING, and the two vocabularies
# are deliberately not shared. They answer OPPOSITE questions about absence:
# here a file that is missing counts AGAINST the change, because a missing
# record proves nothing about the work; there a missing artifact IS the answer,
# because the question is "is it gone?". A reader who carried the meaning of
# CLEAR across to COMPLETE would carry that inversion with it. Distinct words
# are what stop the two verdicts reading as one. check-finish-preflight.sh's
# third vocabulary (RUN1/RUN2/REFUSE) differs for a further reason again: it
# selects a procedure rather than reporting a state.
#
# A MISSING FILE COUNTS AS OUTSTANDING, NOT CLEAR. Treating an absent plan or
# an absent finding record as clearance is the failure this guard exists to
# prevent: a branch merged over unfinished work because nothing was written
# down.
#
# SIGNAL TWO READS THE STORE, NOT A RENDERED FILE. A change's findings are
# rows in the store, keyed by change name, and `flow record findings -change
# <name> -C <worktree>` answers with the decoded JSON array — the same verb
# check-panel-reproducers.sh reads. There is no longer a rendered panel record
# for this guard to resolve a path to, and no missing-file case for it either:
# a change the store has never heard of, or one that genuinely raised no
# findings, answers `[]` at exit 0, which is CLEAR on this signal.
#
# Both signals are counted independently and reported together on the one
# line, so the operator sees the whole picture in one prompt rather than being
# sent back around the loop one signal at a time.
#
# WHY THE CHECKLIST PATTERN IS ANCHORED. `- [ ]` is matched only at the start of
# a line (leading whitespace allowed, for a nested item). Unanchored, it also
# matches a plan that merely QUOTES a checklist inside a fenced example, which
# this repository's own plans do — measured: 44 unanchored hits against 42
# anchored ones in this repository's own kan-17-finish-gate-jira-and-commit-hygiene
# change's tasks.md (now archived), the two extra being `printf` lines in fenced
# shell. A guard that
# fires on every plan that documents a checkbox is a guard the operator learns
# to click past. Fence tracking was rejected as the fuller fix: it is real
# complexity, and what it would still catch — a fenced example whose line BEGINS
# with `- [ ]` — errs toward OUTSTANDING, which prompts the operator rather than
# clearing the gate silently.
#
# A SATELLITE'S MISSING LOCAL PLAN IS NOT THE MISSING-RECORD CASE ABOVE
# REJECTS. "A missing file counts as OUTSTANDING, not CLEAR" is about a
# change whose plan was simply never written — the absence itself is the
# only evidence there is, and treating it as clearance is the exact failure
# this guard exists to prevent. A satellite (KAN-363) carries no local
# `tasks.md` by design: `link.md`'s `## Part of` names a peer tree and a
# canonical change id, and the plan the operator actually needs to check is
# there, not here. Reading that absence the same way as a genuinely missing
# plan would report OUTSTANDING with nothing an operator can act on — the
# plan is not missing, it is elsewhere, and this guard's job is to reach it,
# per `scripts/lib/change-plan.sh` (task 7) and design.md's
# `guards-take-the-canonical-worktree-path`. So a satellite takes one of two
# outcomes and never the third: its canonical plan resolves and is counted
# exactly like a local one would be, or it does not resolve at all and the
# guard refuses outright (exit 2, "cannot determine anything") rather than
# reporting OUTSTANDING over a plan it never actually read. A change
# directory that carries a `link.md` with no `## Part of` — the canonical
# side's own copy, which names its `## Parts` instead — is not a satellite by
# this definition and falls through to the ordinary missing-plan case: this
# guard is never asked to judge a canonical change directory as if it were
# one of its own parts.
#
set -euo pipefail

# NO ANSWER THIS GUARD GIVES MAY DEPEND ON THE CALLER'S LOCALE. Be precise about
# what this line does and does not buy, because a comment claiming more than its
# code delivers is itself one of the defects this file was rewritten to remove.
#
# WHAT THIS LINE DOES BUY, measured rather than assumed: `case`'s `A-Za-z0-9`
# are COLLATING ranges. On bash 3.2 the change name `démo` is rejected by the
# allowlist below under `LC_ALL=C` and ACCEPTED under `en_US.UTF-8` — so without
# this line the guard's accepted input set changes with the operator's
# environment, and this copy of the change-name allowlist diverges from
# check-cleanup-complete.sh's under some locales and not others. It also pins
# `[[:space:]]` and `sort`'s ordering to bytes, which costs nothing and removes
# the question. test-check-unfinished-work.sh case 8e-i asserts the allowlist
# behaves identically under both locales.
export LC_ALL=C

WORKTREE="${1:-}"
NAME="${2:-}"
CANONICAL_WORKTREE="${3:-}"

if [ -z "$WORKTREE" ] || [ -z "$NAME" ]; then
  echo "usage: check-unfinished-work.sh <worktree> <change-name> [canonical-worktree]" >&2
  exit 2
fi

# CONTAINMENT: the change name arrives from a pull-request-editable state file
# and is concatenated into every path below, and passed to `flow record
# findings -change`. Without this check `../../../planted/clear` makes the
# gate read a plan outside the worktree entirely and report CLEAR for a
# change that has no plan of its own, and a glob metacharacter reaches the
# `case` patterns and the `find` filter further down.
#
# The rule is records.Destination's Protection 1
# (stats/internal/records/render.go), character for character, and that
# function's comment is canonical for why each hazard is in it — including the
# incident where `*` in a name "matched and overwrote a DIFFERENT change's
# preserved record". THE `case` BLOCK ITSELF STAYS DUPLICATED, on
# purpose, in this file and in check-panel-reproducers.sh: both harnesses
# assert the same rejected shapes, which is what keeps the two copies from
# drifting apart silently, and a six-line containment check gains nothing
# from being centralized.
#
# It also closes the symlink question at these paths: `-f` and `-d` follow
# symlinks, so a name that cannot leave the change's own directory is what makes
# the paths this guard reads the ones it was asked about.
#
# THE ALLOWED CHARACTERS ARE ENUMERATED RATHER THAN WRITTEN AS RANGES, and here
# that is belt AND braces rather than the fix: `export LC_ALL=C` above already
# makes this copy byte-wise, for the reasons stated there. The enumeration is what
# the other copy needs — check-cleanup-complete.sh runs a project-supplied command
# and so cannot export a locale at all — and this copy takes it so the rule stays
# identical character for character across both, which is the property both
# harnesses assert and the only thing keeping them from drifting.
# check-cleanup-complete.sh's own Protection 1 comment is canonical for the
# measurement and the reasoning; it took that role from the record-copying script
# this repository has since retired. The accepted set is unchanged.
case "$NAME" in
  [!ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789]* \
  | *[!ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789._-]*)
    echo "check-unfinished-work: change name '$NAME' is not a plain change name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'" >&2
    exit 2
    ;;
esac

if [ ! -d "$WORKTREE" ]; then
  echo "check-unfinished-work: $WORKTREE is not a directory — cannot determine anything" >&2
  exit 2
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/spec-root.sh"
source "$SCRIPT_DIR/lib/change-plan.sh"

CHANGES="$WORKTREE/$(spec_root_leaf "$WORKTREE")/changes"
LOCAL_PLAN="$CHANGES/$NAME/tasks.md"

REASONS=""
add() { REASONS="${REASONS:+$REASONS; }$1"; }

# unreadable <path> — a file that exists but cannot be read is neither signal:
# it is an honest unknown, and this guard refuses rather than guessing. Reading
# it as zero unticked boxes would fail toward the reassuring answer, which is
# the one failure direction this whole guard exists to remove.
unreadable() {
  echo "check-unfinished-work: cannot read $1 — cannot determine anything" >&2
  exit 2
}

# count_matching <file> <ere> — the number of lines of <file> matching <ere>.
# Signal one's only helper: signal two no longer greps a rendered file at all,
# so this is defined here rather than sourced from a shared library. `-a` so a
# stray NUL byte — a bad merge, a truncated write — cannot put grep into binary
# mode and read a corrupted file as a clean one; the `rc > 1` split so a real
# read error (2+) is propagated to the caller as a refusal rather than folded
# into a reassuring empty result the way a blanket `|| true` would.
count_matching() {
  local file="$1" ere="$2" out rc=0
  out="$(grep -acE -- "$ere" "$file")" || rc=$?
  if [ "$rc" -gt 1 ]; then
    return "$rc"
  fi
  printf '%s\n' "${out:-0}"
}

# count_unticked <file> — unticked checklist items in a markdown file. The
# pattern lives here alone: the primary plan and every fix sub-change's plan
# are the same knowledge about what an unticked box looks like, and separate
# copies would drift.
count_unticked() {
  count_matching "$1" '^[[:space:]]*- \[ \]'
}

# Signal one — the plan, including any fix sub-change.
#
# THE PRIMARY PLAN IS TRACKED SEPARATELY FROM THE REST, because their absences
# mean opposite things. `spectre/changes/<name>/tasks.md` is the change's plan
# and every flow change has one: missing, it is outstanding like any other
# missing record, and reporting "every plan item is checked" over a change with
# no plan at all is the same silent clearance the header rejects. A fix
# sub-change's plan is genuinely optional — most changes have none — so its
# absence is nothing.
#
# THE THREE GLOB SHAPES THIS USED TO READ ARE GONE. They fixed the layouts a fix
# run may write (`<name>-fix-N` beside the change, or nested inside it) at one
# level of nesting each, and nothing in the repository validates that a fix run
# lands in one of them — so a fix-of-a-fix, or a third layout nobody has written
# yet, silently contributed no signal. Every `tasks.md` at any depth under the
# change's own directory, and under any sibling directory whose name starts
# `<name>-fix-`, is read instead: the discovery no longer depends on guessing
# the layout. What remains undiscoverable is a sub-change whose name has no
# relation to its parent's, which the contract does not permit.
# THE PRIMARY PLAN IS RESOLVED THROUGH change-plan.sh (task 7), NOT COMPOSED
# DIRECTLY, so a satellite's own `link.md` is followed to its canonical
# `tasks.md` instead of reporting the local path's absence as a verdict. See
# this file's header for why a satellite's missing LOCAL plan is not the
# missing-record case the rest of this comment block rejects.
#
# change_plan_dir (not change_plan_path) is called here, because this guard
# needs both the plan AND the directory it lives in: THE FIX-SUB-CHANGE
# SWEEP BELOW MUST WALK THE SAME TREE THE PLAN RESOLVED FROM, not the
# worktree's own changes/ directory unconditionally. A satellite's canonical
# plan resolves under the CANONICAL worktree (or a peer tree), and a
# `<canonical-id>-fix-N` sibling lives beside the canonical change THERE —
# sweeping the satellite's own, plan-less changes/ instead would never see
# it, reporting CLEAR over an unchecked fix-round item. That is exactly the
# silent clearance this guard's own header exists to prevent, so
# SWEEP_CHANGES_DIR and SWEEP_ID below always name the tree and the id the
# plan actually came from — the satellite's own worktree only when the plan
# is, in fact, local.
#
# change_plan_dir returns 1 in two shapes this guard must tell apart: a
# plain change with no `tasks.md` and no `link.md` at all (or a `link.md`
# that names no `## Part of`, which is not a satellite by this guard's
# definition either) — the ordinary missing-plan case below, unchanged from
# before this task — versus a genuine satellite (a `link.md` that DOES carry
# `## Part of`) whose canonical plan could not be reached through either the
# supplied canonical worktree or `peers`. Only the second shape refuses
# outright; the first still falls through to "no plan at" so a change with no
# link at all behaves exactly as it did before this task, sweeping its own
# changes/ exactly as it always has.
PLAN_OPEN=0
if PLAN_DIR="$(change_plan_dir "$WORKTREE" "$NAME" "$CANONICAL_WORKTREE" 2>/dev/null)"; then
  PRIMARY_PLAN="$PLAN_DIR/tasks.md"
  N="$(count_unticked "$PRIMARY_PLAN")" || unreadable "$PRIMARY_PLAN"
  PLAN_OPEN=$((PLAN_OPEN + ${N:-0}))
  SWEEP_CHANGES_DIR="$(dirname "$PLAN_DIR")"
  SWEEP_ID="$(basename "$PLAN_DIR")"
else
  LINK="$CHANGES/$NAME/link.md"
  if [ -f "$LINK" ] && grep -qxF '## Part of' "$LINK" 2>/dev/null; then
    if [ -n "$CANONICAL_WORKTREE" ]; then
      echo "check-unfinished-work: '$NAME' is a satellite change (link.md at $LINK); its canonical plan was not found under the supplied canonical worktree $CANONICAL_WORKTREE, and a canonical worktree that was supplied is never retried against peers — cannot determine anything" >&2
    else
      echo "check-unfinished-work: '$NAME' is a satellite change (link.md at $LINK); no canonical worktree was supplied and its canonical plan could not be resolved through $WORKTREE/$(spec_root_leaf "$WORKTREE")/peers — cannot determine anything" >&2
    fi
    exit 2
  fi
  PRIMARY_PLAN="$LOCAL_PLAN"
  add "no plan at $PRIMARY_PLAN"
  SWEEP_CHANGES_DIR="$CHANGES"
  SWEEP_ID="$NAME"
fi

if [ -d "$SWEEP_CHANGES_DIR" ]; then
  while IFS= read -r -d '' plan; do
    [ "$plan" != "$PRIMARY_PLAN" ] || continue
    case "$plan" in
      "$SWEEP_CHANGES_DIR/$SWEEP_ID/"* | "$SWEEP_CHANGES_DIR/$SWEEP_ID"-fix-*) : ;;
      *) continue ;;
    esac
    N="$(count_unticked "$plan")" || unreadable "$plan"
    PLAN_OPEN=$((PLAN_OPEN + ${N:-0}))
  done < <(find "$SWEEP_CHANGES_DIR" -type f -name tasks.md -print0 2>/dev/null || true)
fi

if [ "$PLAN_OPEN" -gt 0 ]; then
  add "$PLAN_OPEN unchecked plan item(s)"
fi

# Signal two — findings whose recorded status is not closed.
#
# THIS SIGNAL DOES NOT PARSE A FINDINGS TABLE OR A MARKER BLOCK, and the
# deletion is the point. Three review passes once hid an open Critical from a
# hand-rolled GFM table parser six distinct ways — a capitalised status, an
# unescaped `|` shifting columns, a row missing its leading pipe, a table
# inside a blockquote, a status compared through the locale's collating
# sequence, among others — because the guard was recovering one fact, "is any
# finding still open", by re-implementing a document grammar defined in
# another file's prose. None of that class of defect can occur once there is
# no document to parse: a finding is a JSON object in the store, decoded by
# `jq`, with no cells to split and no table boundary to track. Task 2's
# write-time validation also guarantees every stored status is exactly
# `open`, `fixed`, or `withdrawn <reason>` — so a malformed or unrecognised
# status, and a withdrawal with no reason, cannot reach this guard at all;
# there is no branch here to detect either shape any more.
#
# THE STORE IS QUERIED ONCE, and a non-zero exit from `flow record findings`
# is this guard's own exit 2 — "cannot determine anything" — never exit 1's
# "outstanding work found" and never exit 0's "clean". A change the store has
# genuinely never heard of, or one that raised no findings, prints `[]` at
# exit 0, which is the zero-findings case below, not a refusal.
# STDOUT AND STDERR ARE CAPTURED SEPARATELY, and that separation is the whole
# point. `flow` writes diagnostics to stderr -- most reliably the
# `flow: using FLOW_ADDR=...` line it prints whenever the address is
# overridden -- while the JSON this guard parses goes to stdout. Folding the
# two together with `2>&1` puts that diagnostic line at the head of the
# payload, and `jq` then fails with a parse error on a run that actually
# succeeded. The failure mode is invisible in ordinary use and certain under
# the one workflow this repository documents for pointing a session at a
# second daemon (`export FLOW_ADDR=http://127.0.0.1:4174`, per the ui-test
# stack), which is exactly the shape that makes it worth fixing rather than
# tolerating: it breaks only for the operator who followed the instructions.
FINDINGS_ERR="$(mktemp)"
if ! FINDINGS_JSON="$(flow record findings -change "$NAME" -C "$WORKTREE" 2>"$FINDINGS_ERR")"; then
  echo "check-unfinished-work: cannot read findings for '$NAME' from the store — cannot determine anything: $(cat "$FINDINGS_ERR")" >&2
  rm -f "$FINDINGS_ERR"
  exit 2
fi
rm -f "$FINDINGS_ERR"

# An open finding is any finding whose status is neither `fixed` nor a
# `withdrawn <reason>` value — `startswith("withdrawn")` covers the whole
# family, reason text included, without comparing the reason itself.
if ! OPEN_COUNT="$(printf '%s' "$FINDINGS_JSON" | jq '[.[] | select((.status != "fixed") and (.status | startswith("withdrawn") | not))] | length')"; then
  echo "check-unfinished-work: jq failed — cannot determine anything" >&2
  exit 2
fi
if [ "$OPEN_COUNT" -gt 0 ]; then
  add "$OPEN_COUNT open finding(s) in the review panel record"
fi

if [ -z "$REASONS" ]; then
  echo "CLEAR: $WORKTREE — every plan item is checked and no finding is open"
else
  echo "OUTSTANDING: $WORKTREE — $REASONS"
fi
exit 0
