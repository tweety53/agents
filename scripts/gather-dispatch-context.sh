#!/usr/bin/env bash
# gather-dispatch-context.sh — deterministically collect one change's
# planning context into a single bundle, so every dispatch in a /flow implement
# stage (each panel slot, each implementer, each fix subagent) is given
# demonstrably identical inputs instead of separately locating and reading
# proposal.md, design.md, tasks.md and the engineering principles on its
# own.
#
# Usage: gather-dispatch-context.sh <worktree> <change-root> <name> <principles-path> <output-path> [<task-ids> [<canonical-worktree> [<shape>]]]
#
# <task-ids> is optional: a comma-separated list of integer task ids. When
# given, the "## tasks.md" section carries only the plan's header (every
# line before the first task line) plus the named tasks' own blocks, in
# document order, instead of the whole file — a named id the plan does not
# carry is a malformed invocation (exit 2), reported like any other plan
# defect. Task-line, body-boundary and fence grammar are
# plan-dispatch-bundles.py's (see its docstring) — the ids that script
# printed are the ids this finds.
#
# <worktree> and <change-root> are absolute paths; <change-root> is expected
# to sit under <worktree> (spectre/changes/<name>/, but this script does not
# assume that literal shape — it only requires containment). <principles-path>
# is the absolute path of engineering-principles.md, resolved by the CALLER
# (skills/flow/review-panel.md's own [PRINCIPLES_PATH] rule) and never derived
# here — it may legitimately sit OUTSIDE <worktree> entirely, e.g. under a
# global install (~/.claude/skills/flow/engineering-principles.md), so
# it is validated but never checked for containment under <worktree>.
#
# Writes the bundle to <output-path>, falling back to reusing the existing
# file unchanged when its content hash matches the freshly computed one (see
# SKIP-WHEN-UNCHANGED below): a header naming the change, the generating
# instant and the worktree's HEAD sha; a found/skipped/refused census line;
# one "refused: <src> (resolves outside the change directory)" line per source
# whose leaf resolves outside <change-root> (see LEAF VALIDATION below), one
# "skipped: <src> (absent)" line per missing source; then each found source
# under its own "## " heading, in this fixed order: proposal.md, design.md,
# tasks.md, then the principles file.
#
# SKIP-WHEN-UNCHANGED. Everything the bundle prints after the header's
# `generated:`/`head:` lines (the census plus every "## <label>" section) is
# hashed with sha256Hex (stats/internal/guard/sha256.go) and compared against a sidecar <output-path>.hash
# from the previous call. A match skips the write entirely — the existing
# <output-path> and its .hash are left untouched — because a dispatch reads
# that body, and identical bytes need not be rewritten just to refresh a
# timestamp. Either outcome is reported to stderr as
# "gather-dispatch-context: bundle unchanged — reusing <output-path>" or
# "gather-dispatch-context: bundle rebuilt — <reason>", and this path always
# exits 0.
#
# THERE IS NO SPEC SOURCE, AND ADDING ONE BACK WOULD BE A MISTAKE. This
# script used to also carry every <change-root>/specs/*/spec.md — the
# OpenSpec delta specs. Under spectre a change has no specs/ subdirectory at
# all: spec edits go straight into <project>/spectre/specs/<capability>.md on
# the change's own branch, so there was nothing left for that glob to find
# and it reported "skipped: specs/*/spec.md (absent)" on every run of every
# change. The capability specs themselves are NOT bundled in its place —
# which ones a change touches is a judgement from its proposal, and this
# script is deterministic collection, not judgement; skills/flow/implement.md
# names them for the reader instead.
#
# `## standards` entries are NEVER read or carried here — the spec's own
# Requirement forbids it: they resolve through the entry-form table and
# containment rule in skills/flow-contracts/project-configuration.md,
# belong to the principles slot alone, and re-implementing that containment
# logic here would duplicate the contract it depends on. This script globs
# nothing anywhere under <change-root>: it reads three leaves by exact name,
# so any other file placed there (a standards file, a stray note) never
# reaches the bundle.
#
# Exit 2 on a malformed invocation: a missing argument, a change name outside
# the allowlist below, or any of the three paths failing validation (see
# below). Exit 0 in every other case, including every missing source — a
# change may legitimately carry no design.md, and this script performs no
# judgement about that.
#
# VALIDATION. Modelled on scripts/the retired self-review gather's (kan-526) own
# mechanism — read that script's header before touching this one. Each of
# <worktree>, <change-root> and <principles-path> is:
#   1. lexically normalized into an absolute path by pure string
#      manipulation (a component-stack walk: push each `/`-separated
#      segment, pop on `..`, never past the root, skip `.` and empty
#      segments) — touching the filesystem nowhere, so it cannot itself be
#      fooled by a symlink;
#   2. semantically resolved by following every symlink at every component
#      (the leaf, via a bounded readlink loop, and every ancestor directory,
#      via `cd -P`), the resolver always given the LEXICAL form from step 1
#      rather than the raw argument, so a bare "." or a trailing slash —
#      neither of which step 1 ever produces — can never itself masquerade
#      as a divergence;
#   3. compared for EXACT equality against the lexical form.
# For <worktree> and <change-root>, any divergence found in step 3 — an
# ancestor symlink at any depth, the leaf itself being a symlink, a `..`
# component, in any combination — means the invocation is refused, exactly
# as the retired self-review gather's (kan-526) header documents: one general
# mechanism rather than several bounded, individually-bypassable checks.
# They must additionally resolve to an existing directory (not merely pass
# the lexical/real comparison), and <change-root> must resolve INSIDE
# <worktree> — a change-root outside the repository is a malformed
# invocation, not a legitimate absence.
# <principles-path> is normalized and resolved the same way but is NEVER
# refused for a divergence found in step 3, and carries no existence or
# containment requirement: it may be absent (reported as a skipped source,
# exit 0) or sit outside <worktree> (the global-install case above) or be
# reached through a symlink at any depth — a global install reaches every
# skill, and therefore every principles-path, through one by construction.
# It is a trusted argument the CALLER resolves (skills/flow/review-panel.md's
# own [PRINCIPLES_PATH] rule), never attacker-influenced content the way
# <change-root>'s leaves are, so it is simply resolved and read.
#
# THE CHANGE-NAME ALLOWLIST is copied verbatim from
# the retired self-review gather (kan-526): one leading alphanumeric, then letters,
# digits, '.', '_' and '-'.
#
# LEAF VALIDATION. The three CONTENT sources (proposal.md, design.md and
# tasks.md) are never passed as arguments, so the
# exit-2 VALIDATION above does not apply to them — a change legitimately
# carries no design.md, and a symlinked leaf must not abort a whole
# dispatching stage the way a malformed invocation does. Each is instead
# resolved and boundary-checked against <change-root> the same way
# the retired self-review gather's (kan-526) check_boundary() validates its own leaves
# (LEDGER_FILE, PANEL_FILE, TASKS_FILE) before reading them: a leaf that
# resolves outside <change-root> is refused per-source — omitted from the
# bundle, reported as "refused", but the run still exits 0 — never treated
# as a missing source, so an attack is never indistinguishable from a
# legitimate absence. See addFixedSource in
# stats/internal/guard/gatherdispatch.go.
#
# THE PLAN LIVES IN ONE TREE (KAN-393). proposal.md, design.md and tasks.md
# are read from the CONTENT DIRECTORY, resolved once before any leaf is
# read: the <change-root> itself whenever it carries a tasks.md — every
# single-repo change, whose bundle is therefore byte-for-byte what this
# script always wrote — and otherwise, when <change-root> carries a
# satellite's link.md (## Part of) and no tasks.md, the canonical change
# directory scripts/lib/change-plan.sh's rules (ported to
# stats/internal/guard/changeplan.go) resolve through the optional
# seventh argument <canonical-worktree>, or — only when that argument is
# empty — through <worktree>/<spec-root>/peers, the same resolution order
# check-unfinished-work.sh and check-task-commit-fields.sh have used since
# KAN-363. The caller (skills/flow/implement.md's per-bundle gather,
# skills/flow/review-panel.md's rebuild) passes the canonical worktree on
# every call: the member of the run's resolved worktree set whose own
# <project>/<spec-root>/changes/<name>/tasks.md exists, inert on a
# single-repo change. Three consequences, each deliberate:
#
#   1. The boundary the three leaves are checked against is the CONTENT
#      DIRECTORY, not the argument <change-root>: the canonical directory is
#      the same change's tracked content, reached through a link whose peer
#      name and change id change-plan.sh allowlist-checks character for
#      character. The invocation-level containment rules above are
#      unchanged — <change-root> must still sit inside <worktree>, exit 2.
#   2. A satellite whose canonical plan cannot be reached is NOT a refusal:
#      the three plan leaves are reported as skipped with the distinct
#      label "(satellite plan unresolved — link.md at <path>)" and the run
#      still exits 0. This is the deliberate inverse of
#      check-unfinished-work.sh's exit-2 refusal for the same condition — a
#      guard verdict must never silently clear, but this bundle has no
#      verdict to give: it never gates a run (skills/flow/implement.md), so
#      refusing here would abort a dispatching stage a bundle has no
#      authority to stop. A change with no tasks.md and no link.md at all
#      is not a satellite, and its absent leaves are still reported exactly
#      as before.
#   3. When the plan resolved remotely, the three plan sections' labels
#      carry a "(canonical <peer>:<change-id>)" suffix — composed with the
#      existing "(scoped to task(s) ...)" suffix where both apply — so a
#      dispatch reading the bundle can see the plan came from another
#      tree. Everything else stays worktree-local: "project commands" from
#      <worktree>/.flow/project.md, the incidents section from
#      `flow record incidents -C <worktree>`, and the header's head: sha
#      from <worktree>'s own HEAD — a satellite implementer's bundle
#      carries its own repository's commands, its own project's incident
#      log, and the canonical plan.
#
# THE HAZARDS SECTION (KAN-452). After "## incidents" the bundle renders
# "## hazards", sourced from `flow hazards -C <worktree> -shape <shape>`:
# this project's proactive warnings -- incidents' proactive sibling, one
# "- **name (applies):** body" bullet per row, part of the hashed BODY so a
# newly recorded hazard forces a rebuild. <shape> is the change's shape the
# CALLER computes (its resolved worktree set spans more than one repository
# -> cross-repo, otherwise single-repo); the empty/omitted eighth argument
# becomes the literal "all", so a caller that does not know the shape still
# fails open to the always-on rows rather than over-injecting. Anything
# outside the closed vocabulary is a malformed invocation (exit 2), the
# same contract the task-ids argument carries. The section never gates a
# run: flow absent or failing skips as "hazards (flow unavailable)", an
# empty result as "hazards (none)".
#
# <canonical-worktree>, when given, is validated nowhere beyond what
# change-plan.sh's own [ -f ] probes imply — the same trust level as
# check-unfinished-work.sh's third argument: a caller-supplied worktree
# path, not attacker-influenced change content, and every name concatenated
# under it (the canonical change id) is allowlist-checked inside
# change-plan.sh itself.

# Header corrected for the Go port (KAN-760): the hash helper, the leaf
# check and the plan resolver are named as they now are. The body's
# reasoning for every branch is in stats/internal/guard/gatherdispatch.go.
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
. "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "gather-dispatch-context: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec gather-dispatch-context 2 "gather-dispatch-context:" "$@"
