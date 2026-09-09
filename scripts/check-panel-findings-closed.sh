#!/usr/bin/env bash
# check-panel-findings-closed.sh <worktree> <change-name>
#
# THE CHANGE NAME IS A SECOND ARGUMENT for the same reason it is one on
# check-panel-reproducers.sh: a change's findings are rows in the store,
# keyed by change name, not a file at a fixed path. `flow record findings
# -change <name> -C <worktree>` answers with the decoded JSON array of that
# change's findings, and this guard reads only that array.
#
# This is the gate design.md's `gate-is-a-guard` decision chose: nothing
# checked, before this guard existed, that a review panel actually closed
# every finding it verified — `check-unfinished-work.sh` only ever caught
# the omission at /flow's integrate gate, after the work was done. This
# guard runs at the panel's own close, immediately before
# `flow stage end -command '/flow' -stage flow.review-panel`.
#
# THE OPEN-FINDING PREDICATE IS DUPLICATED, on purpose, from
# check-unfinished-work.sh's own copy — design.md's `duplicate-the-predicate`
# decision, following that guard's own precedent for why a one-line
# predicate gains nothing from being centralized and loses the property that
# both harnesses assert the same shape.
#
# THE GUARD NEVER CONSULTS THE JOURNAL — design.md's `no-journal-excuse`
# decision. `flow record status` never blocks, so a store outage journals a
# close instead of landing it, and a finding whose close only reached the
# journal still reads `open` here and reports exit 1. That is the honest
# verdict; the existing handback is where the operator resolves it.
#
# Exit codes:
#   0  no finding's status is open
#   1  one or more are; each still-open ref is named on stderr
#   2  cannot answer at all — no worktree, no change name, a change name
#      outside the allowlist, a worktree that is not a directory, the store
#      unreachable, or jq failing
set -euo pipefail

export LC_ALL=C

WORKTREE="${1:-}"
NAME="${2:-}"
[[ -n "$WORKTREE" && -d "$WORKTREE" ]] || { echo "check-panel-findings-closed: not a directory: ${WORKTREE:-<missing>}" >&2; exit 2; }
[[ -n "$NAME" ]] || { echo "usage: check-panel-findings-closed.sh <worktree> <change-name>" >&2; exit 2; }

# CONTAINMENT, identical to check-panel-reproducers.sh's own copy and its own
# comment canonical for why the six-line `case` block stays duplicated
# rather than centralized: the change name arrives from a pull-request-
# editable state file and is passed to `flow record findings -change`, so
# `../../../planted` and a glob metacharacter are hazards here exactly as
# they are there.
case "$NAME" in
  [!ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789]* \
  | *[!ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789._-]*)
    echo "check-panel-findings-closed: change name '$NAME' is not a plain change name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'" >&2
    exit 2
    ;;
esac

# Canonicalise before the worktree is ever passed to `flow` as `-C`, and
# refuse rather than proceed if it vanished between the `-d` check above and
# here — see check-panel-reproducers.sh's own comment on both hazards,
# canonical for this exact sequence.
WORKTREE="$(cd -- "$WORKTREE" && pwd -P)" || { echo "check-panel-findings-closed: worktree vanished before it could be resolved: ${WORKTREE}" >&2; exit 2; }

# STDOUT AND STDERR ARE CAPTURED SEPARATELY — never `2>&1`. See
# check-unfinished-work.sh's own comment on this exact call, canonical for
# why: `flow` prints diagnostics such as `flow: using FLOW_ADDR=...` to
# stderr, and folding them into stdout puts a non-JSON line at the head of
# what jq parses, breaking a run that actually succeeded.
FINDINGS_ERR="$(mktemp)"
if ! FINDINGS_JSON="$(flow record findings -change "$NAME" -C "$WORKTREE" 2>"$FINDINGS_ERR")"; then
  echo "check-panel-findings-closed: cannot read findings for '$NAME' from the store — cannot determine anything: $(cat "$FINDINGS_ERR")" >&2
  rm -f "$FINDINGS_ERR"
  exit 2
fi
rm -f "$FINDINGS_ERR"

# An open finding is any finding whose status is neither `fixed` nor a
# `withdrawn <reason>` value — `startswith("withdrawn")` covers the whole
# family, reason text included, without comparing the reason itself.
if ! OPEN_REFS="$(printf '%s' "$FINDINGS_JSON" | jq -r '[.[] | select((.status != "fixed") and (.status | startswith("withdrawn") | not)) | .ref] | join(" ")')"; then
  echo "check-panel-findings-closed: jq failed — cannot determine anything" >&2
  exit 2
fi

if [ -n "$OPEN_REFS" ]; then
  echo "check-panel-findings-closed: finding(s) still open: $OPEN_REFS" >&2
  exit 1
fi

echo "FINDINGS-CLOSED"
