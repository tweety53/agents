#!/usr/bin/env bash
# check-model-keys.sh — validate `.flow/project.md`'s `## planning model` and
# `## self review model` keys against the store's own `ValidModels` set.
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
# The valid-model set is never hardcoded: it is extracted from
# `<agents repo>/stats/internal/store/settings.go`'s own `ValidModels` map
# literal, so a model added there is valid here without a second edit.
#
# A key that is absent is not a violation — absence is a supported, valid
# state per project-configuration.md's own "Optional" rows.
#
# Exit 0 when every present key is valid, 1 when any is invalid, 2 when it
# cannot answer at all (settings.go missing/unreadable, or a project root
# that is not a readable directory).
set -uo pipefail

die() {
  echo "check-model-keys: $*" >&2
  exit 2
}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/project-section.sh"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SETTINGS_GO="$REPO_ROOT/stats/internal/store/settings.go"

[[ -r "$SETTINGS_GO" ]] || die "cannot read $SETTINGS_GO"

# Extract the ValidModels map literal's quoted keys — the lines between
# `var ValidModels = map[string]bool{` and its closing `}`, each of shape
# `"name": true,`. Read from the actual source rather than a guessed list,
# per this guard's own header above.
VALID_MODELS=()
while IFS= read -r m; do
  VALID_MODELS+=("$m")
done < <(awk '
  /^var ValidModels = map\[string\]bool\{/ { grabbing = 1; next }
  grabbing && /^\}/ { exit }
  grabbing {
    if (match($0, /"[^"]+"/)) {
      s = substr($0, RSTART + 1, RLENGTH - 2)
      print s
    }
  }
' "$SETTINGS_GO")

[[ "${#VALID_MODELS[@]}" -gt 0 ]] || die "no ValidModels members found in $SETTINGS_GO"

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

  for key in "planning model" "self review model"; do
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
