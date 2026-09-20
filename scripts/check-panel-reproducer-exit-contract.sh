#!/usr/bin/env bash
# check-panel-reproducer-exit-contract.sh <worktree> <change-name>
#
# THE MECHANICAL EXIT-CODE CONTRACT CHECK FOR PANEL REPRODUCERS (KAN-554).
# kan-468's review panel supplied a reproducer for its Important finding
# whose exit-code condition was inverted: it read "defect not demonstrated"
# on the actually-buggy code, and the inversion was caught only by a
# human-grade deferred self-review pass — after the panel record had been
# rendered and relied on. Until this guard existed, nothing between the
# finding being recorded and the fix being dispatched executed the
# reproducer and compared its answer to the claim the finding makes. The
# lexical sibling, check-panel-reproducers.sh, validates the recorded
# command's SHAPE and never runs it; the prose in skills/flow/review-panel.md
# reads exit codes only at fix-dispatch time, as a bounce loop handled by
# attention. This guard is the check that runs instead of attention:
# before the panel relies on a reproducer, the reproducer must demonstrate.
#
# THE CLAIM, AND THE CONTRACT. An OPEN finding claims its defect is present
# in the tree under review, and its reproducer's claim is the same claim:
# run against this worktree, the command must exit non-zero — what
# run-reproducer.sh answers as "defect demonstrated". A reproducer that
# exits 0 here contradicts the claim it is recorded under, whichever way
# its author meant the condition: that inverted reading is the defect class
# this guard exists to name, mechanically, before anything downstream is
# built on the reproducer. Findings whose status is not exactly `open` —
# fixed, deferred, withdrawn — claim nothing about the current tree, so
# their reproducers are skipped: a fixed finding's reproducer is SUPPOSED
# to read "not demonstrated" now, and demanding the opposite verdict of it
# would invert this guard into nonsense. The `none — <reason>` exemption
# and a bare `none` claim nothing runnable and are skipped too: the
# exemption's own rules (legal everywhere but Important) and the bare-none
# violation are check-panel-reproducers.sh's subjects, and that guard runs
# first in the pipeline that calls this one.
#
# THE RUNNER IS THE VERDICT, NEVER A RE-IMPLEMENTATION. Each runnable
# reproducer is executed by "$SCRIPT_DIR/run-reproducer.sh" — the same
# script the per-finding dispatch decisions run — because the exit-code
# vocabulary (demonstrated / not demonstrated / refused / unverifiable /
# cannot answer), the argv-exec barrier, the resolved containment and the
# process-group kill all live there, tested by their own harness. This
# guard adds exactly one thing: the comparison of the runner's verdict to
# the claim, as a gate with its own exit code. The runner is invoked BARE —
# worktree and command line, no --pre-fix-exit — because at dispatch time
# the expected pre-fix verdict is always "demonstrated": passing the flag
# would turn every first verdict into the ambiguity refusal and make the
# gate unanswerable.
#
# THE INSTRUMENT AUDIT (KAN-606). The exit-code contract above reads only
# the verdict, and a verdict is only as good as the instrument behind it:
# kan-552's deferred self-review caught a reproducer whose greps had
# captured mismatched text and cited the wrong test's assertion — the
# script flipped green exactly as recorded, but what it demonstrated was
# not the defect — and only the round-1 re-run reviewers had audited the
# instrument at all. So before this guard invokes the runner, every
# runnable reproducer's own text must declare what it demonstrates and
# where: one `# demonstrates: <path>:<line>:<content>` line per cited
# location, in the grep-output shape the author pastes straight off the
# defect-present tree, inside the same first-10-lines window the
# mutation-reproducer convention (KAN-568) reads its own declaration from.
# Each citation is resolved against the tree under review: the declared
# path must be relative and stay inside the worktree (the sibling lexical
# guard's own shape classes), the file must exist, the line must exist, and
# the declared content must appear on that line. Any citation that does not
# resolve is a violation and the reproducer is NOT run — the runner's
# verdict would answer a question this audit has already settled, and a
# "demonstrated" spent on an unresolvable instrument is exactly the green
# flip this guard exists to deny. A mutation-declared reproducer is exempt
# from the audit: the content it demonstrates is the mutated tree it builds
# at run time, not a location on this tree, and its instrument is audited
# by the KAN-568 sha-pin machinery — demanding resolution here would invert
# the audit into nonsense for exactly the convention KAN-568 added.
#
# Exit codes:
#   0  every open finding with a runnable reproducer demonstrated the
#      defect; findings claiming nothing about the current tree are skipped
#   1  violations found: at least one open finding's reproducer read "defect
#      not demonstrated" on this tree — the inverted class —, was refused
#      by the runner as unusable (a shape the lexical guard's earlier pass
#      did not see: the record changed after it ran, or the command resolves
#      outside the worktree through a symlink), or failed the instrument
#      audit — no `# demonstrates:` declaration within the first 10 lines,
#      a malformed one, a citation outside the worktree, a file, line or
#      content the tree does not carry, or a script that cannot be read to
#      audit at all; each named on stderr
#   2  cannot answer at all — usage, a worktree or change name that fails
#      containment, the store unreachable, jq failing, an open finding
#      carrying no reproducer field at all, or any reproducer the runner
#      could not verdict (timeout, surviving process, plumbing failure).
#      Cannot-answer outranks exit 1: a read that could not be completed is
#      never reported as a verdict, the same precedence its sibling guard
#      gives a null reproducer over a clean answer.
set -euo pipefail

export LC_ALL=C

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"

# The runner is this guard's sibling in scripts/: resolved from SCRIPT_DIR,
# never from PATH, so the guard under test in its own harness sandbox picks
# up the sandbox's stub rather than the real checkout's copy. Readability
# is checked before the first call, and the truthful answer to a missing
# runner is "cannot answer at all", never a verdict.
if [ ! -r "$SCRIPT_DIR/run-reproducer.sh" ]; then
  echo "check-panel-reproducer-exit-contract: cannot read $SCRIPT_DIR/run-reproducer.sh — cannot run any reproducer" >&2
  exit 2
fi

WORKTREE="${1:-}"
NAME="${2:-}"
[[ -n "$WORKTREE" && -d "$WORKTREE" ]] || { echo "check-panel-reproducer-exit-contract: not a directory: ${WORKTREE:-<missing>}" >&2; exit 2; }
[[ -n "$NAME" ]] || { echo "usage: check-panel-reproducer-exit-contract.sh <worktree> <change-name>" >&2; exit 2; }

# CONTAINMENT, the same `case` block check-panel-reproducers.sh and
# check-unfinished-work.sh carry — duplicated on purpose per those guards'
# own convention: the change name reaches this guard from state a pull
# request can edit and is passed to `flow record findings -change`, so the
# same shapes are refused here, and this suite asserts the same rejected
# list so the three copies cannot drift apart silently. `export LC_ALL=C`
# above makes the enumeration byte-wise; `cd ... && pwd -P` canonicalises
# the worktree before it is ever handed to the runner, for the same
# dash-prefixed-relative-path and symlinked-TMPDIR reasons both siblings'
# headers record.
case "$NAME" in
  [!ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789]* \
  | *[!ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789._-]*)
    echo "check-panel-reproducer-exit-contract: change name '$NAME' is not a plain change name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'" >&2
    exit 2
    ;;
esac
WORKTREE="$(cd -- "$WORKTREE" && pwd -P)" || { echo "check-panel-reproducer-exit-contract: worktree vanished before it could be resolved: ${WORKTREE}" >&2; exit 2; }

# THE STORE IS QUERIED ONCE, and a non-zero exit from `flow record
# findings` is this guard's own exit 2 — the same reading its sibling guard
# gives the same call, for the same reason: a read that could not be
# performed is never a clean answer and never a violation.
if ! FINDINGS_JSON="$(flow record findings -change "$NAME" -C "$WORKTREE" 2>&1)"; then
  echo "check-panel-reproducer-exit-contract: cannot read findings for '$NAME' from the store — cannot determine anything: $FINDINGS_JSON" >&2
  exit 2
fi

VIOLATIONS=()
CANNOT_ANSWER=""
RUNNABLE=0
add() { VIOLATIONS+=("$1"); }
cannot() {
  [ -n "$CANNOT_ANSWER" ] && return 0
  CANNOT_ANSWER="$1"
}

# A finding carrying NO reproducer field at all is cannot-answer (exit 2),
# not a skip: the sibling write-time validation makes the shape unreachable
# in practice, and assuming the impossibility holds would quietly drop an
# open finding's claim from this gate. Checked once, up front, scoped to
# OPEN findings — findings at any other status are skipped wholesale below,
# so whatever reproducer field they carry or lack claims nothing here.
if ! EMPTY_REFS="$(printf '%s' "$FINDINGS_JSON" | jq -r '[.[] | select(.status == "open" and (.reproducer == null or .reproducer == "")) | .ref] | join(" ")')"; then
  echo "check-panel-reproducer-exit-contract: jq failed — cannot determine anything" >&2
  exit 2
fi
if [ -n "$EMPTY_REFS" ]; then
  cannot "finding(s) $EMPTY_REFS carry no reproducer field at all"
fi

if ! REFS_TEXT="$(printf '%s' "$FINDINGS_JSON" | jq -r '.[].ref')"; then
  echo "check-panel-reproducer-exit-contract: jq failed — cannot determine anything" >&2
  exit 2
fi
REFS=()
if [ -n "$REFS_TEXT" ]; then
  mapfile -t REFS <<< "$REFS_TEXT"
fi

for ref in "${REFS[@]}"; do
  if ! status="$(printf '%s' "$FINDINGS_JSON" | jq -r --arg ref "$ref" '.[] | select(.ref == $ref) | .status')"; then
    echo "check-panel-reproducer-exit-contract: jq failed — cannot determine anything" >&2
    exit 2
  fi
  if ! reproducer="$(printf '%s' "$FINDINGS_JSON" | jq -r --arg ref "$ref" '.[] | select(.ref == $ref) | .reproducer')"; then
    echo "check-panel-reproducer-exit-contract: jq failed — cannot determine anything" >&2
    exit 2
  fi

  # ONLY AN OPEN FINDING CLAIMS THE CURRENT TREE. Every other status is
  # skipped before the reproducer is even classified: running a fixed
  # finding's reproducer and demanding "demonstrated" would require the
  # defect this same pipeline fixed to still be present.
  if [ "$status" != "open" ]; then
    continue
  fi

  # The two `none` forms are skipped, not run: the bare form is the lexical
  # guard's violation and that guard runs first in this pipeline, and the
  # exemption form claims nothing runnable by construction. One word-boundary
  # guard covers both, since the exemption form also begins `none` + a space;
  # unlike check-panel-reproducers.sh's two-branch shape, no branch here
  # diverges. A command that merely STARTS with the letters `none`
  # (`nonexistent-script`) matches neither pattern and is a runnable command
  # line.
  case "$reproducer" in
    none | none[[:space:]]*) continue ;;
  esac

  # THE INSTRUMENT AUDIT (KAN-606), before the runner is ever invoked: a
  # verdict spent on an instrument whose citation does not resolve is the
  # green flip this guard exists to deny, so an audit failure is recorded
  # and the run is skipped for that finding. The first-10-lines window and
  # the exact-line mutation declaration are the KAN-568 convention's own.
  repro_path="$WORKTREE/${reproducer%% *}"
  if [ ! -f "$repro_path" ] || [ ! -r "$repro_path" ]; then
    add "$ref's reproducer script '${reproducer%% *}' could not be read — its demonstrates declaration cannot be audited, so its claim cannot be checked"
    continue
  fi
  if ! head -n 10 -- "$repro_path" | grep -qx '# mutation-reproducer'; then
    DECLS="$(head -n 10 -- "$repro_path" | grep '^# demonstrates: ' || true)"
    if [ -z "$DECLS" ]; then
      add "$ref's reproducer carries no '# demonstrates: <path>:<line>:<content>' declaration within its first 10 lines — what the instrument reads and expects is unaudited"
      continue
    fi
    while IFS= read -r decl; do
      [ -n "$decl" ] || continue
      rest="${decl#\# demonstrates: }"
      dpath="${rest%%:*}"
      rest2="${rest#*:}"
      dline="${rest2%%:*}"
      dcontent="${rest2#*:}"
      if [ -z "$dpath" ] || [ "$rest2" = "$rest" ] || case "$rest2" in *:*) false ;; *) true ;; esac || \
         case "$dline" in ''|*[!0-9]*) true ;; *) false ;; esac || [ -z "$dcontent" ]; then
        add "$ref's reproducer carries a malformed demonstrates declaration ('$decl') — the form is '# demonstrates: <path>:<line>:<content>'"
        continue
      fi
      case "$dpath" in
        /*|..|../*|*/..|*/../*)
          add "$ref's reproducer demonstrates declaration cites '$dpath' — a citation names a path relative to and inside the worktree under review"
          continue
          ;;
      esac
      target="$WORKTREE/$dpath"
      if [ ! -f "$target" ]; then
        add "$ref's reproducer demonstrates declaration cites '$dpath' — the tree under review carries no such file"
        continue
      fi
      has_line="$(awk -v n="$dline" 'NR==n{f=1} END{print f+0}' "$target")"
      if [ "$has_line" != "1" ]; then
        add "$ref's reproducer demonstrates declaration cites '$dpath':$dline — past the end of the file"
        continue
      fi
      line_text="$(sed -n "${dline}p" "$target")"
      if [[ "$line_text" != *"$dcontent"* ]]; then
        add "$ref's reproducer demonstrates declaration cites content absent from '$dpath':$dline — the instrument's citation does not resolve on this tree"
        continue
      fi
    done <<< "$DECLS"
  fi

  printf 'check-panel-reproducer-exit-contract: %s — running %s\n' "$ref" "$reproducer" >&2
  set +e
  runner_out="$("$SCRIPT_DIR/run-reproducer.sh" "$WORKTREE" "$reproducer" 2>&1)"
  runner_rc=$?
  set -e
  case "$runner_rc" in
    0)
      RUNNABLE=$((RUNNABLE + 1))
      printf 'check-panel-reproducer-exit-contract: %s — defect demonstrated, claim holds\n' "$ref" >&2
      ;;
    1)
      add "$ref's reproducer read 'defect not demonstrated' on the tree under review — an open finding's reproducer must exit non-zero here, and this inverted reading is the exit-code class this guard exists for (KAN-554)"
      printf 'check-panel-reproducer-exit-contract: %s — runner said: %s\n' "$ref" "$runner_out" >&2
      ;;
    2)
      add "$ref's reproducer was refused by run-reproducer.sh as unusable — its claim cannot hold on any tree"
      printf 'check-panel-reproducer-exit-contract: %s — runner said: %s\n' "$ref" "$runner_out" >&2
      ;;
    3)
      cannot "$ref's reproducer could not be verdicted — the runner reported a timeout or a surviving process, which is no verdict at all"
      printf 'check-panel-reproducer-exit-contract: %s — runner said: %s\n' "$ref" "$runner_out" >&2
      ;;
    *)
      cannot "$ref's reproducer could not be verdicted — the runner cannot answer"
      printf 'check-panel-reproducer-exit-contract: %s — runner said: %s\n' "$ref" "$runner_out" >&2
      ;;
  esac
done

# CANNOT-ANSWER OUTRANKS VIOLATIONS. A run whose reads could not be
# completed has produced no verdict for at least one finding, and a partial
# verdict printed at exit 1 would read as a completed check.
if [ -n "$CANNOT_ANSWER" ]; then
  echo "check-panel-reproducer-exit-contract: cannot determine anything — $CANNOT_ANSWER" >&2
  exit 2
fi

if [ "${#VIOLATIONS[@]}" -gt 0 ]; then
  printf 'check-panel-reproducer-exit-contract: %s\n' "${VIOLATIONS[@]}" >&2
  exit 1
fi
# The count is what the loop above actually ran and credited — never a
# second, jq-side re-derivation of the runnable classification, which would
# be a second encoding of the same rule and the copy that drifts.
echo "REPRODUCER-EXIT-CONTRACT-OK ($RUNNABLE runnable open finding(s) demonstrated)"
