#!/usr/bin/env bash
# check-guard-symlinks.sh — the guard that keeps KAN-73's tasks 1-3 true.
#
# Argument-free and self-scoped, exactly like check-references.sh: the scan set
# is this repository's own skills/ and scripts/ directories, resolved from this
# script's own location, so no call site can narrow it.
# CHECK_GUARD_SYMLINKS_ROOT is an explicit, opt-in override honored only when
# set (mirrors CHECK_REFERENCES_ROOT), so the guard's Go tests
# (stats/internal/guard/check_guard_symlinks_test.go) can point it at a
# sandboxed fixture tree without touching this repository — never set it for
# a normal invocation.
#
# Six rules, each reported by name with the offending path and line:
#
#   1. Every entry under skills/*/scripts/ is a symlink, it resolves, and its
#      target is relative — an absolute target would bake this machine's
#      checkout path into the repository. `__pycache__` is skipped: Python
#      writes it beside a .py it imports through the symlinked path, and it
#      is gitignored.
#   2. Every guard INVOKED in a skill's own text — any token of a
#      ```bash/sh/zsh fenced command line (a leading word, a pipeline
#      segment, an `&&` continuation or a command substitution), or a
#      backtick-quoted basename a nearby "Run"/"Invoke"/"Execute"/"invocation"
#      names, or one followed by a `<placeholder>` usage argument, directly
#      or after subcommand words (`name sub <arg>`) — has a symlink in that
#      skill's own scripts/ directory. A guard's sibling dependency — read
#      from the guard's OWN source rather than a hardcoded table, by grepping
#      it for `$SCRIPT_DIR/<name>` or `$(dirname -- "${BASH_SOURCE[0]}")/<name>`
#      — is required exactly where the guard it belongs beside is required.
#      An invoked basename matching NO guard this
#      repository ships is itself a rule 2 violation: any `.sh`-shaped name
#      in an invoking position — the shape every guard here carries — is a
#      typo'd guard name until proven otherwise, and before KAN-530's F6 such
#      a name fell out of the required set silently, shrinking that skill's
#      count while the run still printed GUARD-SYMLINKS-OK. Names without the
#      `.sh` shape stay prose: invoking text is full of `git`, `flow` and
#      `echo` that name no guard and never will.
#   3. No skill text carries a repository-relative `scripts/<name>` path in an
#      invoking position — a non-comment line inside a bash/sh/zsh fence, or a
#      "Run"/"Invoke"/"Execute" immediately before the backtick. Prose about
#      this repository's own guards is exempt, per task 3's boundary: it never
#      matches either shape, so the classifier never has to name the six
#      project-configured guards to leave them alone.
#   4. No guard actually shipped (present as a symlink under some skill's
#      scripts/) derives a repository root as `$SCRIPT_DIR/..` — a fixed
#      number of levels above itself. "Shipped" is read from rule 1's own scan
#      rather than a hardcoded guard list, so a newly shipped guard is covered
#      the moment it is symlinked in.
#   5. No skill directory carries a symlink directly at its own top level —
#      `skills/<skill>/scripts/` is the only place a skill's symlinks may
#      live, which rule 1 already validates; this rule closes the placement
#      rule 1 does not reach. `scripts/` itself is a directory, never a
#      symlink, so it is not a hit here, and a symlink INSIDE scripts/ is
#      rule 1's business, not this one's — this rule never re-reports what
#      rule 1 already governs. It exists for a file a skill's text names but
#      does not carry — most concretely `engineering-principles.md` and the
#      reviewer-prompt files, which `skills/flow/review-panel.md` resolves
#      BESIDE ITSELF, in `skills/flow/` — which is resolved by reading
#      where the text says it lives, never by symlinking a copy into the
#      running command's own skill directory: that symlink makes a wrong
#      reading of the resolution rule work, which is what keeps the wrong
#      reading alive. Prose already said so once and a session created three
#      such symlinks anyway, roughly five hours later, in
#      `skills/flow-fast/`.
#   6. Every *.sh symlink a command skill carries that rule 1 passes must be
#      in that skill's required set — cited in the skill's own text,
#      delegated to it, or a sibling dependency. Rule 2 only ever flags a
#      citation with no symlink and never the reverse; this is that reverse,
#      and it is what turns a skill's prose guard list into a checked one:
#      flow-plan's invoking paragraph named four guards nothing cited, their
#      symlinks were carried on memory alone, and dropping the basename from
#      the paragraph — or pruning the "unused" symlink later — moved only an
#      informational coverage count (KAN-532, F10). Under this rule it is a
#      violation, so the paragraph's basename list cannot drift from the tree
#      in either direction without one. A pair whose requirement genuinely
#      lives outside scannable text — a project-configured guard resolved
#      through .flow/project.md, or a step canonical in a flow-contract file
#      rule 2 deliberately does not scan — is exempt via DECLARED_RULE6, the
#      written declaration in this guard's own source, never inferred.
#
# Prints one violation line per finding (`path:line: message`), then the
# verdict:
#   GUARD-SYMLINKS-OK:      <root> — N guard(s) across M skill(s) validated
#     <skill> <count> · <skill> <count> (declared) · ...
#   GUARD-SYMLINKS-INVALID: <root> — N violation(s)
# The second OK line is coverage.go's report fragment (the Go twin of
# scripts/lib/coverage.sh's coverage_report):
# how many rule-2-required guards this run actually found for EACH command
# skill, so a rule computing an empty required set for one skill (KAN-73's
# own defect, for skills/flow-fast/, before delegation resolution existed)
# is visible on a passing run rather than reading as nothing to check. A
# skill's zero is either named "(declared)" — this guard's own written
# coverage_declare call for it — or, if undeclared, a fifth violation class
# folded into the ordinary `path:line: message` violations above and the
# INVALID count, never a separate exit status or report shape.
#
# Exit 0 clean, 1 with any violation, 2 when it cannot answer at all — a root
# that is not a readable directory, a skills/ or scripts/ directory that
# cannot be scanned, or a file this guard needed to read but could not. A
# refusal writes its reason to stderr and NOTHING to stdout — the caller must
# never read an absent report as a clean one.
#
# THREE DISCIPLINES CARRIED HERE, in the Go port's shape: a file is read
# whole as bytes, so a stray NUL cannot put a matcher into binary "no match"
# mode and read a corrupted guard as a clean one; a failed read is a failure
# to look, which this guard refuses or reports on rather than reads as empty;
# and no path is ever parsed as an option, because none reaches an argv.
#
# WHY RULE 2 IS SCOPED TO A SKILL'S OWN DIRECTORY, NOT EVERY CONTRACT ITS TEXT
# CITES. skills/flow-contracts/*.md is shared prose, cited by path from
# every command skill for reasons that have nothing to do with which guards
# that skill runs — flow-status cites the finish contract to explain how it
# COMBINES a merge-status answer, not to invoke every guard that contract's
# own prose happens to name. Expanding rule 2's scan into every contract a
# skill's text happens to name was tried and produced false positives on this
# repository's own real, correct tree (skills that invoke no guard were
# reported as needing guards found only in a cited contract, never one the
# skill itself carried or ran). Scoping to the skill's own files trades
# a small amount of theoretical coverage — a guard invoked ONLY inside a shared
# contract's usage-synopsis prose, naming no skill by scope — for zero false
# failures on a tree already known correct. A guard already shipped that this
# narrower scan does not detect as required is not a violation either way:
# rule 2 only ever flags a citation with NO corresponding symlink, never a
# symlink with no detected citation.
#
# Ported to Go (KAN-841): the body's reasoning for every step is in
# stats/internal/guard/guardsymlinks.go. The binary does not live in this
# checkout, so this shim exports FLOW_GUARD_REPO_ROOT — this script's parent
# directory, derived as the bash guard derived its ROOT — as the root scanned
# when CHECK_GUARD_SYMLINKS_ROOT is unset.
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
. "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-guard-symlinks: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
FLOW_GUARD_REPO_ROOT="$(cd "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
export FLOW_GUARD_REPO_ROOT
flow_guard_exec check-guard-symlinks 2 "check-guard-symlinks:" "$@"
