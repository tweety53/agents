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
# run against this worktree, the command must produce the runner's "defect
# demonstrated" verdict under whichever exit-code convention it declares —
# a generic one demonstrates with a non-zero exit, a declared
# mutation-reproducer with exit 0 (the build succeeds with the mutation
# landed, KAN-568). A reproducer whose verdict here is "defect not
# demonstrated" contradicts the claim it is recorded under, whichever way
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
# worktree and command line, no --pre-fix-verdict — because at dispatch time
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
# Each citation is resolved against the finding's own tree — the tree the
# per-finding resolution below names, never unconditionally the guard's
# worktree argument: the declared
# path must be relative and stay inside that worktree — lexically (the
# sibling lexical guard's own shape classes) and then physically, realpath-
# resolved and required to remain under that worktree exactly as
# run-reproducer.sh resolves the reproducer's own path token —, the file
# must exist, the line must exist, and
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
#      containment, the store unreachable, the change's state record absent
#      or unreadable (a cross-repo change's guards answer only through its
#      canonical worktree, so a peer tree's store read is a cannot-answer
#      and never a clean verdict — KAN-658), jq failing, an open finding
#      carrying no reproducer field at all, a reproducer path token
#      resolving in several of the change's recorded worktrees, or any
#      reproducer the runner could not verdict (timeout, surviving process,
#      plumbing failure).
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

# THE STATE RECORD IS THE CHANGE'S WORKTREES BOOTSTRAP (KAN-658). The store
# addresses records by project and change name together, and the project key
# resolves from the worktree argument — so on a cross-repo change only the
# canonical worktree resolves the record at all, and a peer worktree's
# findings read below would answer `[]` at exit 0: a clean verdict on a
# change this invocation never actually saw. The record is read FIRST for
# exactly that reason, and this block is DUPLICATED, on purpose, in
# check-panel-reproducers.sh, whose store read has the same blind spot; the
# two harnesses assert the same refused shapes, which is what keeps the
# copies from drifting. `flow state get` reached-and-absent exits 1 — a fact
# about this project/change pair, and on a cross-repo change the signature
# of a guard invoked on a peer tree — reported at this guard's own exit 2,
# never a clean answer. Store-unreachable exits 0 with the fallback record
# only when one exists, so an empty or non-JSON stdout at exit 0 is the same
# cannot-answer class the findings read below already gives an unreachable
# store. The `worktrees` map this record carries keys every worktree of the
# change by absolute path; it is what the per-finding resolution below
# resolves a reproducer's tree from.
# The CLI's own stderr rides along in the refusal, so a missing `flow`
# binary or a dead daemon is reported with its evidence instead of being
# flattened into the cross-repo prose (panel finding F3, kan-658 round 0) —
# the same shape its sibling guard carries.
STATE_ERR_FILE=""
FLOW_STATE_ERR=""
if STATE_ERR_FILE="$(mktemp "${TMPDIR:-/tmp}/check-panel-reproducer-exit-contract-state.XXXXXX" 2>/dev/null)"; then
  STATE_OUT="$(flow state get -C "$WORKTREE" "$NAME" 2>"$STATE_ERR_FILE")" && STATE_RC=0 || STATE_RC=$?
  FLOW_STATE_ERR="$(tr '\n' ' ' < "$STATE_ERR_FILE" 2>/dev/null || true)"
  rm -f "$STATE_ERR_FILE"
else
  STATE_OUT="$(flow state get -C "$WORKTREE" "$NAME" 2>/dev/null)" && STATE_RC=0 || STATE_RC=$?
fi
if [ "$STATE_RC" -ne 0 ]; then
  case "$STATE_RC" in
    126|127)
      # The CLI never ran — reporting the store fact would be a guess
      # wearing evidence's clothes (panel finding F3, kan-658 round 0).
      echo "check-panel-reproducer-exit-contract: cannot run the flow CLI (exit $STATE_RC${FLOW_STATE_ERR:+ — flow said: $FLOW_STATE_ERR}) — cannot determine anything" >&2
      ;;
    *)
      echo "check-panel-reproducer-exit-contract: the store has no record of change '$NAME' under this worktree's project (flow exit $STATE_RC${FLOW_STATE_ERR:+ — flow said: $FLOW_STATE_ERR}) — a cross-repo change's guards answer only through its canonical worktree" >&2
      ;;
  esac
  exit 2
fi
if ! printf '%s' "$STATE_OUT" | jq -e 'type == "object"' >/dev/null 2>&1; then
  echo "check-panel-reproducer-exit-contract: cannot read the state record for '$NAME' — cannot determine anything${FLOW_STATE_ERR:+ — flow said: $FLOW_STATE_ERR}" >&2
  exit 2
fi
if ! WORKTREES_LIST="$(printf '%s' "$STATE_OUT" | jq -r '(.worktrees // {}) | keys[]' 2>/dev/null)"; then
  echo "check-panel-reproducer-exit-contract: jq failed — cannot determine anything" >&2
  exit 2
fi

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
  # The path token is derived the way run-reproducer.sh's own tokenizer
  # derives it — `IFS=$' \t' read -ra`, first element — never
  # `${reproducer%% *}`: that idiom splits on a literal space only, so a
  # tab-separated command line legal everywhere else in the pipeline would
  # reach this audit as one token and be bounced as unreadable (panel
  # finding F1, kan-606 round 0).
  IFS=$' \t' read -ra AUDIT_TOKENS <<< "$reproducer" || true
  repro_path_token="${AUDIT_TOKENS[0]:-}"
  # THE FINDING'S TREE, RESOLVED PER FINDING (KAN-658). The canonical tree
  # first — every single-repo change resolves here, exactly as this guard
  # always has — then the change's recorded worktrees: on a cross-repo
  # change a peer finding's reproducer script lives in ITS repository's
  # worktree, the only tree its relative path token and its demonstrates
  # citations resolve in, and running both against the canonical tree was
  # the could-not-be-read false failure that made every affected reproducer
  # a hand-run substitution. Exactly one recorded worktree carrying the path
  # resolves the tree; several cannot be told apart, which is no verdict for
  # this finding rather than a guessed one; none is the existing unreadable
  # class below, unchanged. A recorded path that is not a directory on disk
  # is skipped like an absent one — the map records git worktrees, and a
  # vanished entry resolves nothing.
  TREE="$WORKTREE"
  if [ ! -e "$WORKTREE/$repro_path_token" ]; then
    TREE_MATCHES=()
    while IFS= read -r recorded_wt; do
      [ -n "$recorded_wt" ] || continue
      [ -d "$recorded_wt" ] || continue
      # `cd ... && pwd -P` gives the recorded path the same physical shape
      # $WORKTREE itself carries — the containment `case` below compares
      # realpath answers, and a map entry reached through a symlinked
      # prefix would otherwise lose every comparison on its shape alone.
      recorded_wt="$(cd -- "$recorded_wt" && pwd -P)" || continue
      # Two map entries can name one physical tree through a symlinked
      # prefix; counting the alias twice would refuse an unambiguous
      # reproducer as ambiguous (panel finding F4, kan-658 round 0).
      case " ${TREE_MATCHES[*]:-} " in
        *" $recorded_wt "*) continue ;;
      esac
      [ -e "$recorded_wt/$repro_path_token" ] || continue
      TREE_MATCHES+=("$recorded_wt")
    done <<< "$WORKTREES_LIST"
    if [ "${#TREE_MATCHES[@]}" -eq 1 ]; then
      TREE="${TREE_MATCHES[0]}"
    elif [ "${#TREE_MATCHES[@]}" -gt 1 ]; then
      cannot "$ref's reproducer path token '$repro_path_token' resolves in ${#TREE_MATCHES[@]} of the change's recorded worktrees (${TREE_MATCHES[*]}) — the finding's tree is ambiguous, so no verdict is possible"
      continue
    fi
  fi
  repro_path="$TREE/$repro_path_token"
  if [ ! -f "$repro_path" ] || [ ! -r "$repro_path" ]; then
    add "$ref's reproducer script '$repro_path_token' could not be read — its demonstrates declaration cannot be audited, so its claim cannot be checked"
    continue
  fi
  if ! head -n 10 -- "$repro_path" | grep -qx '# mutation-reproducer'; then
    DECLS="$(head -n 10 -- "$repro_path" | grep '^# demonstrates: ' || true)"
    if [ -z "$DECLS" ]; then
      add "$ref's reproducer carries no '# demonstrates: <path>:<line>:<content>' declaration within its first 10 lines — what the instrument reads and expects is unaudited"
      continue
    fi
    audit_violations="${#VIOLATIONS[@]}"
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
      target="$TREE/$dpath"
      if [ ! -f "$target" ]; then
        add "$ref's reproducer demonstrates declaration cites '$dpath' — the tree under review carries no such file"
        continue
      fi
      # Resolved containment, the runner's own pattern: `realpath` follows
      # `..`, `.` and symlinks in one step, so a citation whose declared
      # path carries no lexical `..` segment but escapes through a symlink
      # inside the worktree is caught here rather than read outside the
      # tree (panel finding F4, kan-606 round 0). $TREE is the guard's own
      # `pwd -P`-canonical worktree or a recorded map path resolved against
      # it, the same physical shape realpath answers.
      resolved="$(realpath -- "$target" 2>/dev/null)" || resolved=""
      case "$resolved" in
        "$TREE"/*) : ;;
        *)
          add "$ref's reproducer demonstrates declaration cites '$dpath' — it resolves to '${resolved:-an unresolvable path}', outside the worktree under review — a symlink escape"
          continue
          ;;
      esac
      has_line="$(awk -v n="$dline" 'NR==n{f=1} END{print f+0}' "$resolved")"
      if [ "$has_line" != "1" ]; then
        add "$ref's reproducer demonstrates declaration cites '$dpath':$dline — past the end of the file"
        continue
      fi
      line_text="$(sed -n "${dline}p" "$resolved")"
      if [[ "$line_text" != *"$dcontent"* ]]; then
        add "$ref's reproducer demonstrates declaration cites content absent from '$dpath':$dline — the instrument's citation does not resolve on this tree"
        continue
      fi
    done <<< "$DECLS"
    # A `continue` inside the declaration loop above continues that loop,
    # never the finding loop — so the skip the audit promises is enforced
    # here: any violation this finding's declarations added means the
    # runner is never invoked for it.
    if [ "${#VIOLATIONS[@]}" -gt "$audit_violations" ]; then
      continue
    fi
  fi

  printf 'check-panel-reproducer-exit-contract: %s — running %s\n' "$ref" "$reproducer" >&2
  set +e
  runner_out="$("$SCRIPT_DIR/run-reproducer.sh" "$TREE" "$reproducer" 2>&1)"
  runner_rc=$?
  set -e
  case "$runner_rc" in
    0)
      RUNNABLE=$((RUNNABLE + 1))
      printf 'check-panel-reproducer-exit-contract: %s — defect demonstrated, claim holds\n' "$ref" >&2
      ;;
    1)
      add "$ref's reproducer read 'defect not demonstrated' on the tree under review — an open finding's reproducer must read demonstrated here under whichever exit-code convention it declares: a generic one demonstrates with a non-zero exit, a declared mutation-reproducer with exit 0 (the build succeeds with the mutation landed) — this inverted reading is the exit-code class this guard exists for (KAN-554)"
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
