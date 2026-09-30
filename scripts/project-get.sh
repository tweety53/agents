#!/usr/bin/env bash
# project-get.sh — print the body of one `## <key>` heading from a project's
# `.flow/project.md`, so every /flow phase file resolves a project-configured
# key through one script instead of loading a contract file to read one value.
#
# Usage: project-get.sh <project-root> <key>
#        project-get.sh <project-root> <key> --enum <literal>...
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
# --enum RESOLVES A SINGLE-LINE-LITERAL KEY (skills/flow-contracts/
# project-configuration.md). The body's head — its first non-blank line,
# whitespace-trimmed, surrounding backticks removed, trimmed again — is
# matched byte-for-byte against each <literal>:
#   0  the head matches; the literal alone is on stdout.
#   1  unchanged: the file or the key is absent.
#   2  unchanged; also `--enum` given no literal.
#   3  the key is declared but its head matches no literal — one stderr
#      line quotes the head; the caller reports it and resolves as absent.
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
#
# The logic lives in stats/internal/guard/projectget.go, a byte-for-byte port
# of this script's bash (ae805186); flow-guard is built from this checkout,
# never taken from PATH — scripts/lib/flow-guard.sh derives it, and exits 2
# with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
. "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "project-get: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec project-get 2 "project-get:" "$@"
