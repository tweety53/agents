#!/usr/bin/env bash
# check-model-keys.sh — validate `.flow/project.md`'s `## self review model`
# key against the store's own `ValidModels` set.
#
# Usage: check-model-keys.sh [<project root> ...]
#
# With no arguments it checks the repository this script ships in (mirrors
# check-workspace-isolation.sh's own convention). Each argument is a PROJECT
# ROOT — the directory holding `.flow/project.md` — never the file itself.
#
# **Project configuration** (`skills/flow-contracts/project-configuration.md`)
# is canonical: both keys are optional, and each body — leading/trailing
# whitespace trimmed, nothing else normalized — must match exactly one
# `ValidModels` member and nothing else. A resolver reading a mismatched body
# reports it by name and falls back as if the key were absent; THIS guard is
# stricter on purpose — it exists so that fallback never has to happen in the
# first place, so a mismatch here is a hard failure, not a silent drop.
#
# The valid-model set is asked, not regexed: `flow settings models` prints
# the harness's compiled-in `ValidModels` set (the same store.ValidModels
# the daemon serves), one name per line, and the checkout's own
# settings.go is cross-checked against it. When the two agree, the CLI set
# governs. When they drift — an installed binary predating the checkout —
# the source set governs the verdict, so a stale install cannot fail a
# value the checkout declares, and the drift is announced with each
# side's unique members. The map-literal parse is also the offline
# fallback when the CLI cannot answer at all.
#
# A key that is absent is not a violation — absence is a supported, valid
# state per project-configuration.md's own "Optional" rows.
#
# Exit 0 when every present key is valid, 1 when any is invalid, 2 when it
# cannot answer at all (no usable `flow settings models` answer AND a
# settings.go that is missing/unreadable or yields no members, or a project
# root that is not a readable directory).
set -uo pipefail

die() {
  echo "check-model-keys: $*" >&2
  exit 2
}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/project-section.sh"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SETTINGS_GO="$REPO_ROOT/stats/internal/store/settings.go"

[[ -r "$SETTINGS_GO" ]] && HAS_SOURCE=1 || HAS_SOURCE=0

# The map-literal extraction shared by the offline fallback and the
# cross-check: the quoted keys between `var ValidModels =
# map[string]bool{` and its closing `}`, each of shape `"name": true,`.
source_set() {
  awk '
    /^var ValidModels = map\[string\]bool\{/ { grabbing = 1; next }
    grabbing && /^\}/ { exit }
    grabbing {
      if (match($0, /"[^"]+"/)) {
        s = substr($0, RSTART + 1, RLENGTH - 2)
        print s
      }
    }
  ' "$SETTINGS_GO"
}
SOURCE_SET=""
[[ "$HAS_SOURCE" -eq 1 ]] && SOURCE_SET="$(source_set)"

# Ask the installed `flow` CLI first — its `settings models` subcommand
# prints the compiled-in set with no daemon needed. Its stderr is relayed
# to this guard's stderr rather than discarded, so a warn-but-answer ask
# is still visible, and the fallback announcement below can say why the
# ask failed, not only that it did.
VALID_MODELS=()
ASK_NOTE=" — no flow on PATH"
if command -v flow >/dev/null 2>&1; then
  ASK_ERR="$(mktemp "${TMPDIR:-/tmp}/check-model-keys-askerr.XXXXXX")"
  while IFS= read -r m; do
    [[ -n "$m" ]] && VALID_MODELS+=("$m")
  done < <(flow settings models 2>"$ASK_ERR")
  if [[ -s "$ASK_ERR" ]]; then
    ASK_NOTE=" — $(tr '\n' ' ' < "$ASK_ERR" | sed 's/[[:space:]]*$//')"
    # Relay only when the ask answered: on a failed ask the fallback
    # announcement below already embeds the same diagnostic, and one
    # emission per stderr stream is the difference between a warning and
    # noise.
    if [[ "${#VALID_MODELS[@]}" -gt 0 ]]; then
      echo "check-model-keys: flow settings models reported:${ASK_NOTE}" >&2
    fi
  elif [[ "${#VALID_MODELS[@]}" -eq 0 ]]; then
    ASK_NOTE=" — empty answer"
  else
    ASK_NOTE=""
  fi
  rm -f "$ASK_ERR"
fi

if [[ "${#VALID_MODELS[@]}" -eq 0 ]]; then
  echo "check-model-keys: no answer from flow settings models${ASK_NOTE} — falling back to parsing $SETTINGS_GO" >&2
  [[ "$HAS_SOURCE" -eq 1 ]] || die "cannot read $SETTINGS_GO"
  [[ -n "$SOURCE_SET" ]] || die "no ValidModels members found in $SETTINGS_GO"
  while IFS= read -r m; do
    VALID_MODELS+=("$m")
  done <<< "$SOURCE_SET"
elif [[ -n "$SOURCE_SET" ]]; then
  # The CLI answered and the source is readable: compare the two. Drift
  # means the installed binary predates the checkout — the
  # check-installed-rules.sh class of failure — and is named with each
  # side's unique members. On drift the source set governs the verdict
  # below, so a stale install cannot fail a value the checkout declares;
  # the announcement is what tells the operator to rebuild flow.
  DRIFT="$(comm -3 \
    <(printf '%s\n' "${VALID_MODELS[@]}" | sort -u) \
    <(printf '%s\n' "$SOURCE_SET" | sort -u))"
  if [[ -n "$DRIFT" ]]; then
    echo "check-model-keys: installed flow's ValidModels set disagrees with $SETTINGS_GO — rebuild or reinstall flow; left column only-in-CLI, right column only-in-source:" >&2
    printf '%s\n' "$DRIFT" >&2
    echo "check-model-keys: the source set governs this run's verdict while the install is stale" >&2
    VALID_MODELS=()
    while IFS= read -r m; do
      VALID_MODELS+=("$m")
    done <<< "$SOURCE_SET"
  fi
elif [[ "$HAS_SOURCE" -eq 1 ]]; then
  echo "check-model-keys: note — $SETTINGS_GO yielded no ValidModels members; the source cross-check is off (did the map literal's shape change?)" >&2
fi

is_valid_model() {
  local target="$1" m
  for m in "${VALID_MODELS[@]}"; do
    [[ "$m" == "$target" ]] && return 0
  done
  return 1
}

# extract_section_body <file> <heading text> — prints the trimmed body of the
# first `## <heading text>` section (everything up to the next `^## ` heading
# or EOF), or nothing if the heading is absent. Leading/trailing blank lines
# are stripped, matching project-configuration.md's own "trimmed" rule for
# these two keys' single-line-literal bodies. Surrounding backticks are also
# stripped, matching every actual `.flow/project.md` in this repository
# (`## jira`, `## default landing route`, `## self review model` all write
# their single-line-literal value as a markdown code span) and
# jira-followups.md's own "strip surrounding backticks" rule for the same
# kind of body.
extract_section_body() {
  local file="$1" heading="$2"
  project_section "$file" "$heading" | awk '
    { lines[NR] = $0 }
    END {
      start = 1
      end = NR
      while (start <= end && lines[start] ~ /^[[:space:]]*$/) start++
      while (end >= start && lines[end] ~ /^[[:space:]]*$/) end--
      for (i = start; i <= end; i++) {
        line = lines[i]
        sub(/^[[:space:]]+/, "", line)
        sub(/[[:space:]]+$/, "", line)
        if (line ~ /^`.*`$/ && length(line) >= 2) {
          line = substr(line, 2, length(line) - 2)
        }
        print line
      }
    }
  '
}

VIOLATIONS=0
CHECKED=0

check_project() {
  local root="$1" pf key body
  [[ -d "$root" ]] || die "not a directory: $root"
  [[ -r "$root" ]] || die "cannot read directory: $root"
  pf="$root/.flow/project.md"
  [[ -e "$pf" ]] || { printf 'MODEL-KEYS-OK: %s — no .flow/project.md (nothing to check)\n' "$root"; CHECKED=$((CHECKED + 1)); return; }
  [[ -f "$pf" ]] || die "not a regular file: $pf"
  [[ -r "$pf" ]] || die "cannot read: $pf"

  for key in "self review model"; do
    body="$(extract_section_body "$pf" "$key")"
    [[ -n "$body" ]] || continue
    # A multi-line body already fails: the contract is a single-line literal.
    if [[ "$body" == *$'\n'* ]] || ! is_valid_model "$body"; then
      printf '%s: `## %s` value %q is not a ValidModels member\n' "$pf" "$key" "$body"
      VIOLATIONS=$((VIOLATIONS + 1))
    fi
  done
  CHECKED=$((CHECKED + 1))
}

if [[ $# -eq 0 ]]; then
  check_project "$REPO_ROOT"
else
  for root in "$@"; do
    check_project "$root"
  done
fi

if [[ "$VIOLATIONS" -gt 0 ]]; then
  echo "check-model-keys: $VIOLATIONS violation(s) across $CHECKED project(s) checked" >&2
  exit 1
fi

printf 'MODEL-KEYS-OK: %s project(s) checked\n' "$CHECKED"
exit 0
