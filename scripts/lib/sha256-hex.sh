# scripts/lib/sha256-hex.sh — sha256_hex, defined once.
#
# Extracted from scripts/check-cleanup-complete.sh's own copy once
# scripts/gather-dispatch-context.sh grew a second, byte-for-byte-identical
# one (kan-288's own review panel, F1) — gather-dispatch-context.sh's
# original comment argued against extracting it on the grounds that
# rules/build-the-simplest-thing.mdc defers abstraction "until a second
# caller exists"; a second caller is exactly what had just been added, which
# is the case FOR extraction under that same rule, not against it. Both
# callers ship through the skills/flow/scripts/ symlink farm, alongside its
# own `lib` symlink into scripts/lib/ — the "SAFELY REACH IT" criterion
# scripts/lib/resolve-file.sh's own header states for when a guard may
# source a sibling instead of carrying its own copy.
#
# Not meant to be executed directly — a caller sources it and calls
# sha256_hex; it sets no `set -euo pipefail` of its own and relies on the
# sourcing script's.

# sha256_hex <string> — the lowercase hex SHA-256 of the string's bytes, or a
# non-zero exit if this machine has no SHA-256 tool.
#
# The TOOL is not part of the contract, only the bytes are, so the three
# canonical names are tried in turn — `shasum` is on stock macOS and Linux,
# `sha256sum` is on Linux, `openssl` is nearly everywhere. All three print
# the digest as a whitespace-delimited field somewhere on their line
# (`<hex>  -` for the first two, `(stdin)= <hex>` for openssl), so the field
# is selected by its SHAPE rather than by its position: stripping non-hex
# characters instead would keep the `d` out of openssl's "(stdin)" and
# produce a digest that is wrong by one byte in exactly the way this whole
# derivation exists to prevent.
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
