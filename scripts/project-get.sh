#!/usr/bin/env bash
# project-get.sh — print the body of one `## <key>` heading from a project's
# `.flow/project.md`, over the shared scripts/lib/project-section.sh
# extractor, so every /flow phase file resolves a project-configured key
# through one script instead of loading a contract file to read one value.
#
# Usage: project-get.sh <project-root> <key>
#
# <project-root> is the directory holding `.flow/project.md`, never the file
# itself. <key> is the heading text after `## `, quoted when it carries
# spaces (`project-get.sh <root> "default landing route"`).
#
# THREE EXIT CODES:
#   0  the key is declared exactly once; its body is on stdout.
#   1  `<project-root>/.flow/project.md` is absent, or declares no
#      `## <key>` heading — one line on stderr says which. Every
#      "optional key absent -> skip" case in the phase files is this exit.
#   2  usage (not exactly two arguments), `<project-root>` is not a
#      directory, or `## <key>` is declared more than once — the same
#      ambiguity refusal check-visual-trigger.sh makes: a second
#      declaration is ambiguous, so neither is read.
#
# Derives no repository root by walking a fixed number of directory levels
# above its own location (check-guard-symlinks.sh rule 4) — the project
# root is always an argument, never inferred from where this script lives.
#
# RESOLUTION ORDER (KAN-520): when <project-root> sits inside a git work tree
# and HEAD carries .flow/project.md, the key is counted and read from HEAD's
# copy (`git show HEAD:<path>`), never from the working tree — a staged-but-
# uncommitted edit there must not silently override the landed value every
# /flow* stage reads (the kan-512 run-2 incident). The working tree is read
# only when HEAD lacks the file or the root is not a git repository — today's
# behavior, unchanged there. When the working tree's file exists and diverges
# from HEAD on the key being resolved, one stderr warning names the file, the
# key, and that HEAD won.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/project-section.sh"

[ "$#" -eq 2 ] || { echo "usage: project-get.sh <project-root> <key>" >&2; exit 2; }
ROOT="$1"; KEY="$2"
[ -d "$ROOT" ] || { echo "project-get: $ROOT is not a directory" >&2; exit 2; }
CFG="$ROOT/.flow/project.md"

# A caller's ambient GIT_DIR/GIT_WORK_TREE/GIT_INDEX_FILE — a hook, a
# rebase --exec — would hijack which repository the git calls below read;
# pin them to $ROOT alone.
unset GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE

SRC="$CFG"
if REL="$(git -C "$ROOT" rev-parse --show-prefix 2>/dev/null)"; then
  if git -C "$ROOT" cat-file -e "HEAD:${REL}.flow/project.md" 2>/dev/null; then
    SRC="$(mktemp "${TMPDIR:-/tmp}/project-get.XXXXXX")"
    trap 'rm -f "$SRC"' EXIT
    git -C "$ROOT" show "HEAD:${REL}.flow/project.md" > "$SRC"
    if [ -f "$CFG" ] && \
       [ "$(project_section "$SRC" "$KEY")" != "$(project_section "$CFG" "$KEY")" ]; then
      echo "project-get: WARNING: $CFG diverges from HEAD on '## $KEY' — resolving from HEAD; commit or restore the working tree to silence this" >&2
    fi
  fi
fi

[ -f "$SRC" ] || { echo "project-get: $CFG does not exist" >&2; exit 1; }
COUNT="$(strip_bom_cat "$SRC" | awk -v key="## $KEY" '$0 == key { n++ } END { print n + 0 }')"
case "$COUNT" in
  0) echo "project-get: $CFG declares no '## $KEY' section" >&2; exit 1 ;;
  1) project_section "$SRC" "$KEY" ;;
  *) echo "project-get: $CFG declares $COUNT '## $KEY' sections — a second declaration is ambiguous, so neither was read" >&2; exit 2 ;;
esac
