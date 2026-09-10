#!/usr/bin/env bash
# Assertion harness for check-dispatch-paragraphs.sh.
#
# Builds throwaway fixture trees under TMPDIR and points the guard at each
# with CHECK_DISPATCH_PARAGRAPHS_ROOT, invoking the REAL
# scripts/check-dispatch-paragraphs.sh as a subprocess against REAL fixture
# files on disk — never a copy of its logic, and never a hand-written
# expected-output string asserted without having run the guard to produce
# it. Never edits this repository's own skills/flow/*.md.
#
# Cases 1-7 cover REPRODUCE, DON'T READ, matching the plan's original
# minimum: both sites correct; the label absent from review-panel.md; the
# label present but missing a shared phrase; implement.md carrying only one
# block; implement.md carrying two blocks of the same variant (the guard
# must match variants, not just count blocks); a scoped file missing; a
# scoped file that is a symlink. Case 1's review-panel.md fixture also
# carries a correct VERBATIM REPORT — THE FACT block, since that site now
# requires both paragraphs to exit 0.
#
# Cases 8-11 cover KAN-217's VERBATIM REPORT — THE FACT paragraph, added
# when the guard was generalized from one hard-coded paragraph to a table
# of them: present and clean (case 1, above); the block missing entirely;
# and one case per required phrase, each dropped in turn.
#
# Case 12 covers CHECK_DISPATCH_PARAGRAPHS_ROOT set but empty (distinct
# from unset). Case 13 covers a required site path that is a directory,
# not a regular file. Case 14 covers a required site file that exists but
# is not readable. Case 15 is the block-text-bleed regression: two
# required blocks glued with no blank line between them, where the first
# is deficient and must not pass by absorbing the second's phrases.
#
# Cases 16-19 cover KAN-263's FOREGROUND BUILDS paragraph, required twice
# in each of implement.md and review-panel.md: case 1's fixtures (and
# every other case's fixtures asserted clean) now carry two correct
# FOREGROUND BUILDS blocks per file alongside their existing blocks; case
# 16 is the label absent entirely from implement.md; cases 17-19 are one
# case per required phrase, each dropped in turn.
#
# Cases 20-21 close a gap Bugbot found in review: cases 16-19 only cover
# "label entirely absent" (block_count 0) and "a block present but missing
# one required phrase" (a phrase-level failure, block_count still meets
# min_blocks since the deficient block still counts as one) — never the
# min_blocks=2 threshold itself, i.e. exactly ONE fully-correct FOREGROUND
# BUILDS block present and the SECOND required occurrence entirely
# missing. That is precisely the failure mode a dispatcher forgetting the
# second required occurrence would produce. Mutating either foreground
# SITE_MIN_BLOCKS entry from 2 to 1 left every existing case green; case 20
# (implement.md, one correct block, second omitted) and case 21
# (review-panel.md, same shape) each pin the min-blocks violation's own
# message, so that mutation now fails them.
#
# Cases 22-26 cover KAN-441's TARGETED TESTS paragraph, required once in
# each of implement.md and review-panel.md (the panel-fix dispatch): case
# 1's fixtures (and every other case's fixtures asserted clean) now carry
# one correct TARGETED_BLOCK per file alongside their existing blocks; case
# 22 is the label absent entirely from implement.md; cases 23-25 are one
# case per required phrase, each dropped in turn; case 26 is the label
# absent entirely from review-panel.md.
#
# Cases 27-30 cover KAN-464's MUTATION PROOF paragraph, required once in
# review-panel.md (the panel-fix subagent dispatch) and nowhere else: case
# 1's review-panel.md fixture (the only fixture this guard asserts clean)
# now also carries one correct MUTATION_BLOCK; case 27 is the label absent
# entirely from review-panel.md; cases 28-30 are one case per required
# phrase, each dropped in turn.
#
# Cases 31-34 cover KAN-473's TOOLS paragraph, required twice in each of
# implement.md and review-panel.md and once in each of brainstorm.md and
# verify-and-handoff.md: new_root now seeds every sandbox with a correct
# TOOLS_BLOCK in brainstorm.md and verify-and-handoff.md by default (no
# other case exercises either file), and case 1's review-panel.md and
# implement.md fixtures (and every other case's, which still carry only
# the paragraphs those cases' own SITE tables register) each gain two
# TOOLS_BLOCK blocks. Case 31 is the label absent entirely from
# brainstorm.md; cases 32-34 are one case per required phrase, each
# dropped from one of implement.md's two TOOLS blocks in turn while the
# other stays correct.
#
# Cases 42-44 cover the three assert phrases KAN-257 added to the MUTATION
# PROOF entry, one case per phrase, each dropped in turn from the
# paragraph's assert sentences: every MUTATION_BLOCK-shaped fixture now
# carries the extended paragraph, and each new case asserts the guard
# names its dropped phrase.
#
# Cases 35-41 cover KAN-472 task 10's MODEL HANDSHAKE paragraph, required
# twice in each of implement.md and review-panel.md and once in each of
# brainstorm.md and verify-and-handoff.md — the same six sites TOOLS
# occupies: new_root now also seeds every sandbox with a correct
# HANDSHAKE_BLOCK in brainstorm.md and verify-and-handoff.md by default, and
# CLEAN_REVIEW_PANEL/CLEAN_IMPLEMENT (and case 1's own fixtures) each gain
# two HANDSHAKE_BLOCK blocks. Case 35 is the label absent entirely from
# implement.md; case 36 the label absent from review-panel.md; case 37 the
# label absent from brainstorm.md (overriding new_root's default); case 38
# the label absent from verify-and-handoff.md (same override); cases 39-41
# are one case per required phrase, each dropped from one of implement.md's
# two HANDSHAKE blocks in turn while the other stays correct.
#
# Case 45 covers KAN-472 task 20's INDEPENDENT PASSES paragraph, required
# once at review-panel.md's bundle prompt and nowhere else: CLEAN_REVIEW_PANEL
# (and case 1's own fixture) now also carry one correct INDEPENDENT_BLOCK.
# Case 45 is the label absent entirely from review-panel.md.
#
# Cases 46-51 cover KAN-484's NO DELEGATION paragraph, required once each in
# implement.md (the implementer dispatch) and verify-and-handoff.md (the
# verifier dispatch), twice in review-panel.md (the panel slot and
# panel-fix subagent dispatches): new_root now also seeds every sandbox's
# verify-and-handoff.md with a correct DELEGATION_BLOCK by default, and
# CLEAN_IMPLEMENT gains one copy, CLEAN_REVIEW_PANEL two, alongside case 1's
# own fixtures. Case 46 is the label absent entirely from implement.md;
# cases 47-49 are one case per required phrase, each dropped from one of
# review-panel.md's two blocks while the other stays correct; case 50 is
# the label absent from verify-and-handoff.md (overriding new_root's
# default); case 51 is review-panel.md with exactly one correct block and
# the second required occurrence entirely missing, pinning the min-blocks
# threshold from the start rather than after a review finds the gap, as
# cases 20-21 had to for FOREGROUND BUILDS.
#
# Per KAN-197, every check this file targets was mutation-tested by hand
# during authoring: the check was disabled or removed from a throwaway copy
# of the guard, the same fixture re-run, and the case's failure signal
# confirmed absent — proving the fixture's failure depended on that check,
# not on some other check incidentally catching the same fixture. That is
# an authoring-time review practice, not a mechanism this file runs itself
# (contrast test-check-plan-shape.sh, which automates the same idea
# in-suite via a throwaway mutated copy per case). Case 6's inline comment
# below records one such mutation catching a real surviving-mutant bug, and
# cases 2 and 4 assert on the min-blocks violation's own message — pinned
# after hand-mutating SITE_MIN_BLOCKS showed the exit-code-only assertion
# passing regardless of that check.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-dispatch-paragraphs.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

# Every case leaves one sandbox directory behind, removed on exit including
# on a failed assertion. An indexed array, not a space-separated string:
# mktemp paths under TMPDIR may contain spaces.
DIRS=()
cleanup() {
  [ "${#DIRS[@]}" -eq 0 ] && return 0
  local d
  for d in "${DIRS[@]}"; do
    rm -rf "$d"
  done
}
trap cleanup EXIT

# new_root -> sets ROOT to a fresh sandbox directory carrying
# skills/flow/, matching the required-site table's scan-root-relative
# paths. brainstorm.md and verify-and-handoff.md are required TOOLS and
# MODEL HANDSHAKE sites (min 1 block each) that no case below otherwise
# exercises, so every root is seeded with a correct TOOLS_BLOCK and
# HANDSHAKE_BLOCK in each by default — a case testing something else never
# has to think about these two files, and case 31 (TOOLS) / case 37-38
# (MODEL HANDSHAKE) below are the ones that override a default.
new_root() {
  ROOT="$(mktemp -d "${TMPDIR:-/tmp}/check-dispatch-paragraphs-test.XXXXXX")"
  DIRS+=("$ROOT")
  mkdir -p "$ROOT/skills/flow"
  printf '%s\n\n%s\n' "$TOOLS_BLOCK" "$HANDSHAKE_BLOCK" > "$ROOT/skills/flow/brainstorm.md"
  printf '%s\n\n%s\n\n%s\n' "$TOOLS_BLOCK" "$HANDSHAKE_BLOCK" "$DELEGATION_BLOCK" > "$ROOT/skills/flow/verify-and-handoff.md"
}

# run_guard -> sets RC and OUT, running the real guard against $ROOT.
run_guard() {
  set +e
  OUT="$(CHECK_DISPATCH_PARAGRAPHS_ROOT="$ROOT" "$GUARD" 2>&1)"
  RC=$?
  set -e
}

# The two REPRODUCE, DON'T READ variants, reproduced verbatim from
# design.md's Part 1 blockquotes.
REVIEWER_BLOCK='> **REPRODUCE, DON'"'"'T READ:** Where a behaviour crosses a boundary — the store, the filesystem, a
> guard, a real transcript, a real process — at least one check you make MUST exercise the real
> thing. A claim you did not run is worth less than one you did: a doc comment, a type signature
> and a passing test can each read plausibly and be false. Run it before you accept it, and run it
> before you reject it.'

IMPLEMENTER_BLOCK='> **REPRODUCE, DON'"'"'T READ:** Where a behaviour crosses a boundary — the store, the filesystem, a
> guard, a real transcript, a real process — at least one test you write MUST exercise the real
> thing. A test backed by a fake or a hand-built value passes while the real integration is broken:
> the shape you construct by hand is not the shape the real producer emits. Build the value the way
> production builds it, or assert against the real boundary.'

# A reviewer block with its "exercise the real thing" phrase gutted — case 3.
REVIEWER_BLOCK_NO_SHARED_PHRASE='> **REPRODUCE, DON'"'"'T READ:** Where a behaviour crosses a boundary — the store, the filesystem, a
> guard, a real transcript, a real process — at least one check you make MUST do the real
> thing. A claim you did not run is worth less than one you did: a doc comment, a type signature
> and a passing test can each read plausibly and be false. Run it before you accept it, and run it
> before you reject it.'

# The VERBATIM REPORT — THE FACT paragraph, reproduced verbatim from
# skills/flow/review-panel.md.
VERBATIM_BLOCK='> **VERBATIM REPORT — THE FACT:** each finding below names the file its slot'"'"'s report was written
> to. That file is the reviewer'"'"'s own report, unedited — read it before you act on the finding.
> The structured block is the dispatcher'"'"'s summary of it: direction on what to work on, and
> never a source of fact. Where the two disagree the report wins. Where the block asserts
> something the report does not, treat it as unchecked and establish it yourself before building
> on it.'

# Variants of VERBATIM_BLOCK, each with exactly one required phrase dropped
# while staying a plausible paragraph — cases 10-12.
VERBATIM_BLOCK_NO_REVIEWERS_OWN_REPORT='> **VERBATIM REPORT — THE FACT:** each finding below names the file its slot'"'"'s report was written
> to. That file is the report, unedited — read it before you act on the finding.
> The structured block is the dispatcher'"'"'s summary of it: direction on what to work on, and
> never a source of fact. Where the two disagree the report wins. Where the block asserts
> something the report does not, treat it as unchecked and establish it yourself before building
> on it.'

VERBATIM_BLOCK_NO_NEVER_A_SOURCE_OF_FACT='> **VERBATIM REPORT — THE FACT:** each finding below names the file its slot'"'"'s report was written
> to. That file is the reviewer'"'"'s own report, unedited — read it before you act on the finding.
> The structured block is the dispatcher'"'"'s summary of it: direction on what to work on, and
> is not itself a fact. Where the two disagree the report wins. Where the block asserts
> something the report does not, treat it as unchecked and establish it yourself before building
> on it.'

VERBATIM_BLOCK_NO_REPORT_WINS='> **VERBATIM REPORT — THE FACT:** each finding below names the file its slot'"'"'s report was written
> to. That file is the reviewer'"'"'s own report, unedited — read it before you act on the finding.
> The structured block is the dispatcher'"'"'s summary of it: direction on what to work on, and
> never a source of fact. Where the two disagree the reviewer'"'"'s report is followed. Where the block asserts
> something the report does not, treat it as unchecked and establish it yourself before building
> on it.'

# The FOREGROUND BUILDS paragraph, reproduced verbatim from
# skills/flow/implement.md and skills/flow/review-panel.md.
FOREGROUND_BLOCK='> **FOREGROUND BUILDS:** Never end your turn with a build, test run, or other long-running
> command still executing in the background. Run it in the foreground, or poll it to
> completion, before you stop.'

# Variants of FOREGROUND_BLOCK, each with exactly one required phrase
# dropped while staying a plausible paragraph — cases 17-19.
FOREGROUND_BLOCK_NO_STILL_EXECUTING='> **FOREGROUND BUILDS:** Never end your turn with a build, test run, or other long-running
> command left running in the background. Run it in the foreground, or poll it to
> completion, before you stop.'

FOREGROUND_BLOCK_NO_RUN_FOREGROUND='> **FOREGROUND BUILDS:** Never end your turn with a build, test run, or other long-running
> command still executing in the background. Bring it to the foreground, or poll it to
> completion, before you stop.'

FOREGROUND_BLOCK_NO_POLL_COMPLETION='> **FOREGROUND BUILDS:** Never end your turn with a build, test run, or other long-running
> command still executing in the background. Run it in the foreground, or wait for it to
> finish, before you stop.'

# The TARGETED TESTS paragraph, reproduced verbatim from
# skills/flow/implement.md and skills/flow/review-panel.md.
TARGETED_BLOCK='> **TARGETED TESTS:** Run only the tests this task'"'"'s `**Tests:**` field names, through the build
> tool'"'"'s own selector — `--tests '"'"'<class>'"'"'` for Gradle, `-run '"'"'<name>'"'"'` for `go test`, `-t
> '"'"'<name>'"'"'` for vitest — once for RED, once for GREEN, and again only after a source edit. Never
> run the module or repository suite mid-task: the full `## test` list runs once per worktree at
> the last bundle, and again in `flow.verify`. Pipe a test run'"'"'s output through `tail` so a green
> run costs lines of context, not a build log.'

# Variants of TARGETED_BLOCK, each with exactly one required phrase dropped
# while staying a plausible paragraph — cases 23-25.
TARGETED_BLOCK_NO_SELECTOR='> **TARGETED TESTS:** Run only the tests this task'"'"'s `**Tests:**` field names, through whichever
> selector applies — `--tests '"'"'<class>'"'"'` for Gradle, `-run '"'"'<name>'"'"'` for `go test`, `-t
> '"'"'<name>'"'"'` for vitest — once for RED, once for GREEN, and again only after a source edit. Never
> run the module or repository suite mid-task: the full `## test` list runs once per worktree at
> the last bundle, and again in `flow.verify`. Pipe a test run'"'"'s output through `tail` so a green
> run costs lines of context, not a build log.'

TARGETED_BLOCK_NO_RED_GREEN='> **TARGETED TESTS:** Run only the tests this task'"'"'s `**Tests:**` field names, through the build
> tool'"'"'s own selector — `--tests '"'"'<class>'"'"'` for Gradle, `-run '"'"'<name>'"'"'` for `go test`, `-t
> '"'"'<name>'"'"'` for vitest — once before the fix and once after, and again only after a source edit.
> Never run the module or repository suite mid-task: the full `## test` list runs once per worktree
> at the last bundle, and again in `flow.verify`. Pipe a test run'"'"'s output through `tail` so a
> green run costs lines of context, not a build log.'

TARGETED_BLOCK_NO_SUITE='> **TARGETED TESTS:** Run only the tests this task'"'"'s `**Tests:**` field names, through the build
> tool'"'"'s own selector — `--tests '"'"'<class>'"'"'` for Gradle, `-run '"'"'<name>'"'"'` for `go test`, `-t
> '"'"'<name>'"'"'` for vitest — once for RED, once for GREEN, and again only after a source edit. Never
> run the whole test suite mid-task: the full `## test` list runs once per worktree at
> the last bundle, and again in `flow.verify`. Pipe a test run'"'"'s output through `tail` so a green
> run costs lines of context, not a build log.'

# The MUTATION PROOF paragraph, reproduced verbatim from
# skills/flow/review-panel.md.
MUTATION_BLOCK='> **MUTATION PROOF:** every executable behaviour your fix changed is mutation-proved before you
> end your turn — not only the test cases this round adds. Mutate the mechanism: revert it in a
> scratch tree, or flip the single value it turns on — but before the tests run, confirm the edit
> landed: the target changed where you intended, not nowhere and not somewhere else. An edit that
> never applied is a refusal, not a surviving mutant — redo it with a working mechanism; it never
> buys a test. Then confirm an existing test fails, and restore.
> `mutate-and-verify.sh` mechanizes backup, apply, run, report and restore
> for a mutation expressed as a patch file against one or more test harnesses; which mechanism to
> mutate is your judgment, not the script'"'"'s. Each mutation alters one mechanism — where a single
> revert would also change state a second check reads, split it into surgical mutations, one per
> mechanism. A mutation no test catches is a surviving mutant: add the test that catches it before
> your turn ends. Record one `fix-mutation:` line per behaviour in your report, plus a
> `fix-mutations-total:` count, in the shape the review-panel contract'"'"'s fenced block gives. Where
> you cannot judge whether a survivor is real or an equivalent mutant, say so in the report rather
> than deciding it yourself.'

# Variants of MUTATION_BLOCK, each with exactly one required phrase dropped
# while staying a plausible paragraph — cases 28-30 (the original three
# phrases) and cases 35-37 (the assert phrases KAN-257 added). Dropping a
# phrase means the whole block stops carrying it; where a phrase's wording
# occurs inside another phrase's sentence — "a surviving mutant" inside
# "a refusal, not a surviving mutant" — both occurrences go.
MUTATION_BLOCK_NO_MUTATION_PROVED='> **MUTATION PROOF:** every executable behaviour your fix changed is mutation-proved before your
> turn ends — not only the test cases this round adds. Mutate the mechanism: revert it in a
> scratch tree, or flip the single value it turns on — but before the tests run, confirm the edit
> landed: the target changed where you intended, not nowhere and not somewhere else. An edit that
> never applied is a refusal, not a surviving mutant — redo it with a working mechanism; it never
> buys a test. Then confirm an existing test fails, and restore.
> `mutate-and-verify.sh` mechanizes backup, apply, run, report and restore
> for a mutation expressed as a patch file against one or more test harnesses; which mechanism to
> mutate is your judgment, not the script'"'"'s. Each mutation alters one mechanism — where a single
> revert would also change state a second check reads, split it into surgical mutations, one per
> mechanism. A mutation no test catches is a surviving mutant: add the test that catches it before
> your turn ends. Record one `fix-mutation:` line per behaviour in your report, plus a
> `fix-mutations-total:` count, in the shape the review-panel contract'"'"'s fenced block gives. Where
> you cannot judge whether a survivor is real or an equivalent mutant, say so in the report rather
> than deciding it yourself.'

MUTATION_BLOCK_NO_CONFIRM_AND_RESTORE='> **MUTATION PROOF:** every executable behaviour your fix changed is mutation-proved before you
> end your turn — not only the test cases this round adds. Mutate the mechanism: revert it in a
> scratch tree, or flip the single value it turns on — but before the tests run, confirm the edit
> landed: the target changed where you intended, not nowhere and not somewhere else. An edit that
> never applied is a refusal, not a surviving mutant — redo it with a working mechanism; it never
> buys a test. Then check that an existing test fails, and restore.
> `mutate-and-verify.sh` mechanizes backup, apply, run, report and restore
> for a mutation expressed as a patch file against one or more test harnesses; which mechanism to
> mutate is your judgment, not the script'"'"'s. Each mutation alters one mechanism — where a single
> revert would also change state a second check reads, split it into surgical mutations, one per
> mechanism. A mutation no test catches is a surviving mutant: add the test that catches it before
> your turn ends. Record one `fix-mutation:` line per behaviour in your report, plus a
> `fix-mutations-total:` count, in the shape the review-panel contract'"'"'s fenced block gives. Where
> you cannot judge whether a survivor is real or an equivalent mutant, say so in the report rather
> than deciding it yourself.'

MUTATION_BLOCK_NO_SURVIVING_MUTANT='> **MUTATION PROOF:** every executable behaviour your fix changed is mutation-proved before you
> end your turn — not only the test cases this round adds. Mutate the mechanism: revert it in a
> scratch tree, or flip the single value it turns on — but before the tests run, confirm the edit
> landed: the target changed where you intended, not nowhere and not somewhere else. An edit that
> never applied is a refusal to redo with a working mechanism; it never buys a test. Then confirm
> an existing test fails, and restore.
> `mutate-and-verify.sh` mechanizes backup, apply, run, report and restore
> for a mutation expressed as a patch file against one or more test harnesses; which mechanism to
> mutate is your judgment, not the script'"'"'s. Each mutation alters one mechanism — where a single
> revert would also change state a second check reads, split it into surgical mutations, one per
> mechanism. A mutation no test catches is an uncaught mutation: add the test that catches it before
> your turn ends. Record one `fix-mutation:` line per behaviour in your report, plus a
> `fix-mutations-total:` count, in the shape the review-panel contract'"'"'s fenced block gives. Where
> you cannot judge whether a survivor is real or an equivalent mutant, say so in the report rather
> than deciding it yourself.'

MUTATION_BLOCK_NO_EDIT_LANDED='> **MUTATION PROOF:** every executable behaviour your fix changed is mutation-proved before you
> end your turn — not only the test cases this round adds. Mutate the mechanism: revert it in a
> scratch tree, or flip the single value it turns on — but before the tests run, the edit is
> asserted: the target changed where you intended, not nowhere and not somewhere else. An edit that
> never applied is a refusal, not a surviving mutant — redo it with a working mechanism; it never
> buys a test. Then confirm an existing test fails, and restore.
> `mutate-and-verify.sh` mechanizes backup, apply, run, report and restore
> for a mutation expressed as a patch file against one or more test harnesses; which mechanism to
> mutate is your judgment, not the script'"'"'s. Each mutation alters one mechanism — where a single
> revert would also change state a second check reads, split it into surgical mutations, one per
> mechanism. A mutation no test catches is a surviving mutant: add the test that catches it before
> your turn ends. Record one `fix-mutation:` line per behaviour in your report, plus a
> `fix-mutations-total:` count, in the shape the review-panel contract'"'"'s fenced block gives. Where
> you cannot judge whether a survivor is real or an equivalent mutant, say so in the report rather
> than deciding it yourself.'

MUTATION_BLOCK_NO_REFUSAL='> **MUTATION PROOF:** every executable behaviour your fix changed is mutation-proved before you
> end your turn — not only the test cases this round adds. Mutate the mechanism: revert it in a
> scratch tree, or flip the single value it turns on — but before the tests run, confirm the edit
> landed: the target changed where you intended, not nowhere and not somewhere else. An edit that
> never applied is an unapplied edit — redo it with a working mechanism; it never buys a test.
> Then confirm an existing test fails, and restore.
> `mutate-and-verify.sh` mechanizes backup, apply, run, report and restore
> for a mutation expressed as a patch file against one or more test harnesses; which mechanism to
> mutate is your judgment, not the script'"'"'s. Each mutation alters one mechanism — where a single
> revert would also change state a second check reads, split it into surgical mutations, one per
> mechanism. A mutation no test catches is a surviving mutant: add the test that catches it before
> your turn ends. Record one `fix-mutation:` line per behaviour in your report, plus a
> `fix-mutations-total:` count, in the shape the review-panel contract'"'"'s fenced block gives. Where
> you cannot judge whether a survivor is real or an equivalent mutant, say so in the report rather
> than deciding it yourself.'

MUTATION_BLOCK_NO_NEVER_BUYS='> **MUTATION PROOF:** every executable behaviour your fix changed is mutation-proved before you
> end your turn — not only the test cases this round adds. Mutate the mechanism: revert it in a
> scratch tree, or flip the single value it turns on — but before the tests run, confirm the edit
> landed: the target changed where you intended, not nowhere and not somewhere else. An edit that
> never applied is a refusal, not a surviving mutant — redo it with a working mechanism before
> moving on. Then confirm an existing test fails, and restore.
> `mutate-and-verify.sh` mechanizes backup, apply, run, report and restore
> for a mutation expressed as a patch file against one or more test harnesses; which mechanism to
> mutate is your judgment, not the script'"'"'s. Each mutation alters one mechanism — where a single
> revert would also change state a second check reads, split it into surgical mutations, one per
> mechanism. A mutation no test catches is a surviving mutant: add the test that catches it before
> your turn ends. Record one `fix-mutation:` line per behaviour in your report, plus a
> `fix-mutations-total:` count, in the shape the review-panel contract'"'"'s fenced block gives. Where
> you cannot judge whether a survivor is real or an equivalent mutant, say so in the report rather
> than deciding it yourself.'

# The TOOLS paragraph, reproduced verbatim from design.md, required at six
# dispatch sites (KAN-473): implement.md and review-panel.md twice each,
# brainstorm.md and verify-and-handoff.md once each.
TOOLS_BLOCK='> **TOOLS:** Every tool you need that is not already listed in your tool set — `SendMessage`,
> `Monitor`, an MCP tool — is loaded in one `select:<name>,<name>` ToolSearch in your first turn,
> before anything else. Never ToolSearch for a tool already listed, and never a wildcard query: a
> schema loaded later changes your tool list and re-prices your whole context at full input rate.'

# Variants of TOOLS_BLOCK, each with exactly one required phrase dropped
# while staying a plausible paragraph — cases 32-34.
TOOLS_BLOCK_NO_FIRST_TURN='> **TOOLS:** Every tool you need that is not already listed in your tool set — `SendMessage`,
> `Monitor`, an MCP tool — is loaded in one `select:<name>,<name>` ToolSearch before your first
> tool call, before anything else. Never ToolSearch for a tool already listed, and never a
> wildcard query: a schema loaded later changes your tool list and re-prices your whole context at
> full input rate.'

TOOLS_BLOCK_NO_WILDCARD='> **TOOLS:** Every tool you need that is not already listed in your tool set — `SendMessage`,
> `Monitor`, an MCP tool — is loaded in one `select:<name>,<name>` ToolSearch in your first turn,
> before anything else. Never ToolSearch for a tool already listed, and never a broad query: a
> schema loaded later changes your tool list and re-prices your whole context at full input rate.'

TOOLS_BLOCK_NO_REPRICES='> **TOOLS:** Every tool you need that is not already listed in your tool set — `SendMessage`,
> `Monitor`, an MCP tool — is loaded in one `select:<name>,<name>` ToolSearch in your first turn,
> before anything else. Never ToolSearch for a tool already listed, and never a wildcard query: a
> schema loaded later changes your tool list and costs your whole context again at full input
> rate.'

# The MODEL HANDSHAKE paragraph, reproduced verbatim from design.md, required
# at the same six dispatch sites as TOOLS (KAN-472 task 10): implement.md and
# review-panel.md twice each, brainstorm.md and verify-and-handoff.md once
# each.
HANDSHAKE_BLOCK='> **MODEL HANDSHAKE:** the first line of your first reply is `Model: <the model named in your own
> system prompt>` and nothing else on that line. Answer it before any tool call.'

# Variants of HANDSHAKE_BLOCK, each with exactly one required phrase dropped
# while staying a plausible paragraph — cases 39-41.
HANDSHAKE_BLOCK_NO_FIRST_LINE='> **MODEL HANDSHAKE:** the very first line you send is `Model: <the model named in your own
> system prompt>` and nothing else on that line. Answer it before any tool call.'

HANDSHAKE_BLOCK_NO_NOTHING_ELSE='> **MODEL HANDSHAKE:** the first line of your first reply is `Model: <the model named in your own
> system prompt>` and only that on the line. Answer it before any tool call.'

HANDSHAKE_BLOCK_NO_BEFORE_TOOL_CALL='> **MODEL HANDSHAKE:** the first line of your first reply is `Model: <the model named in your own
> system prompt>` and nothing else on that line. Answer it before making any tool call.'

# The INDEPENDENT PASSES paragraph, reproduced verbatim from design.md,
# required once at review-panel.md's bundle prompt (KAN-472 task 20).
INDEPENDENT_BLOCK='> **INDEPENDENT PASSES:** each pass starts from `final-review.diff` and the code, never from an
> earlier pass'"'"'s report or conclusions. Do not cite, defer to, or skip a defect because an earlier
> pass raised it — if it sits in this pass'"'"'s angle, raise it again under this pass. Write each
> pass'"'"'s report file before beginning the next pass.'

# The NO DELEGATION paragraph, reproduced verbatim from design.md, required
# once each in implement.md and verify-and-handoff.md, twice in
# review-panel.md (KAN-484).
DELEGATION_BLOCK='> **NO DELEGATION:** Do this work yourself. Never call the `Agent` tool, and never spawn a
> subagent, background agent or helper of any kind — you are the leaf of this run, and any child
> you start is unrecorded and outside the conductor'"'"'s closed list (**Dispatch sites — the
> conductor'"'"'s closed list**, `skills/flow/implement.md`). Reading, searching, reproducing and
> fixing are your own Read, Bash and Edit calls.'

# Variants of DELEGATION_BLOCK, each with exactly one required phrase
# dropped while staying a plausible paragraph — cases 47-49.
DELEGATION_BLOCK_NO_NEVER_CALL_AGENT='> **NO DELEGATION:** Do this work yourself. Do not use the `Agent` tool, and never spawn a
> subagent, background agent or helper of any kind — you are the leaf of this run, and any child
> you start is unrecorded and outside the conductor'"'"'s closed list (**Dispatch sites — the
> conductor'"'"'s closed list**, `skills/flow/implement.md`). Reading, searching, reproducing and
> fixing are your own Read, Bash and Edit calls.'

DELEGATION_BLOCK_NO_NEVER_SPAWN='> **NO DELEGATION:** Do this work yourself. Never call the `Agent` tool, and start no subagent,
> background agent or helper of any kind — you are the leaf of this run, and any child
> you start is unrecorded and outside the conductor'"'"'s closed list (**Dispatch sites — the
> conductor'"'"'s closed list**, `skills/flow/implement.md`). Reading, searching, reproducing and
> fixing are your own Read, Bash and Edit calls.'

DELEGATION_BLOCK_NO_LEAF='> **NO DELEGATION:** Do this work yourself. Never call the `Agent` tool, and never spawn a
> subagent, background agent or helper of any kind — you are the last agent in this chain, and any child
> you start is unrecorded and outside the conductor'"'"'s closed list (**Dispatch sites — the
> conductor'"'"'s closed list**, `skills/flow/implement.md`). Reading, searching, reproducing and
> fixing are your own Read, Bash and Edit calls.'

write_site() {
  local relpath="$1" content="$2"
  printf '%s\n' "$content" > "$ROOT/$relpath"
}

# ===========================================================================
# Case 1: both required sites correct — exit 0. review-panel.md now
# carries both the REPRODUCE reviewer block and the VERBATIM REPORT block,
# plus two FOREGROUND BUILDS blocks (panel slot dispatch, panel-fix
# dispatch), one TARGETED TESTS block (panel-fix dispatch), one MUTATION
# PROOF block (panel-fix dispatch), two TOOLS blocks (panel slot,
# panel-fix dispatch) and one INDEPENDENT PASSES block (the bundle prompt,
# KAN-472 task 20); implement.md carries two FOREGROUND BUILDS blocks
# too (implementer dispatch, the conductor's own §4 instruction), one
# TARGETED TESTS block (implementer dispatch) and two TOOLS blocks
# (conductor dispatch, implementer dispatch), plus two MODEL HANDSHAKE
# blocks each (KAN-472 task 10) and one NO DELEGATION block (implementer
# dispatch); review-panel.md carries two NO DELEGATION blocks (panel slot,
# panel-fix dispatch), same as its two TOOLS blocks (KAN-484). new_root
# already seeded brainstorm.md and verify-and-handoff.md with their own
# required TOOLS and MODEL HANDSHAKE blocks — verify-and-handoff.md is also
# seeded with its own required NO DELEGATION block.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK

$VERBATIM_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK

$MUTATION_BLOCK

$TOOLS_BLOCK

$TOOLS_BLOCK

$HANDSHAKE_BLOCK

$HANDSHAKE_BLOCK

$INDEPENDENT_BLOCK

$DELEGATION_BLOCK

$DELEGATION_BLOCK"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK

$TOOLS_BLOCK

$TOOLS_BLOCK

$HANDSHAKE_BLOCK

$HANDSHAKE_BLOCK

$DELEGATION_BLOCK"
run_guard
[ "$RC" -eq 0 ] && pass "case 1: both sites correct exits 0" \
  || fail "case 1: expected exit 0, got rc=$RC out=$OUT"

# ===========================================================================
# Case 2: the REPRODUCE label is absent from review-panel.md — exit 1,
# names the file.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "No REPRODUCE, DON'T READ paragraph here at all.

$VERBATIM_BLOCK"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK"
run_guard
[ "$RC" -eq 1 ] && pass "case 2: exits 1" || fail "case 2: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"review-panel.md"*) pass "case 2: names review-panel.md" ;;
  *) fail "case 2: expected review-panel.md named in output, got: $OUT" ;;
esac
# Pin the min-blocks violation's OWN message, not merely the overall exit
# code: with the label wholly absent, block_count is 0 for BOTH the
# min-blocks check and the reviewer-variant check, so an assertion on exit
# code alone cannot tell them apart. Hand-mutating SITE_MIN_BLOCKS's
# review-panel.md entry from 1 to 0 confirmed this exact message
# disappears while the case's RC==1 assertion above still passes (the
# variant violation alone keeps it red) — this line is what actually pins
# the min-blocks check for that site.
case "$OUT" in
  *"requires at least 1 block(s)"*) pass "case 2: names the min-blocks violation" ;;
  *) fail "case 2: expected the min-blocks violation message, got: $OUT" ;;
esac

# ===========================================================================
# Case 3: the REPRODUCE label is present but a shared phrase is missing —
# exit 1, names the phrase.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK_NO_SHARED_PHRASE

$VERBATIM_BLOCK"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK"
run_guard
[ "$RC" -eq 1 ] && pass "case 3: exits 1" || fail "case 3: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"exercise the real thing"*) pass "case 3: names the missing phrase" ;;
  *) fail "case 3: expected 'exercise the real thing' named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 4: implement.md carries only one block — exit 1.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK

$VERBATIM_BLOCK"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK"
run_guard
[ "$RC" -eq 1 ] && pass "case 4: exits 1" || fail "case 4: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"implement.md"*) pass "case 4: names implement.md" ;;
  *) fail "case 4: expected implement.md named in output, got: $OUT" ;;
esac
# Pin implement.md's own min-blocks violation message (its site requires 2
# blocks). Hand-mutating SITE_MIN_BLOCKS's implement.md entry from 2 to 1
# confirmed this exact message disappears while the case's RC==1 assertion
# above still passes (the missing-implementer-variant violation alone
# keeps it red) — this line is what actually pins the min-blocks check for
# that site.
case "$OUT" in
  *"requires at least 2 block(s)"*) pass "case 4: names the min-blocks violation" ;;
  *) fail "case 4: expected the min-blocks violation message, got: $OUT" ;;
esac

# ===========================================================================
# Case 5: implement.md carries two blocks of the SAME variant — exit 1. The
# reviewer/implementer pair is the requirement, not the count alone.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK

$VERBATIM_BLOCK"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$REVIEWER_BLOCK"
run_guard
[ "$RC" -eq 1 ] && pass "case 5: two same-variant blocks still exits 1" \
  || fail "case 5: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"implement.md"*"implementer"*) pass "case 5: names implement.md's missing implementer variant" ;;
  *) fail "case 5: expected implement.md's missing implementer variant named, got: $OUT" ;;
esac

# ===========================================================================
# Case 6: a scoped file is missing entirely — exit 2.
# ===========================================================================
new_root
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK"
# review-panel.md is never written.
run_guard
[ "$RC" -eq 2 ] && pass "case 6: a missing scoped file exits 2" \
  || fail "case 6: expected exit 2, got rc=$RC out=$OUT"
case "$OUT" in
  *"review-panel.md"*) pass "case 6: names the missing file" ;;
  *) fail "case 6: expected review-panel.md named in output, got: $OUT" ;;
esac
# Pin the message to its OWN code path, not merely to the exit code. Exit 2 is
# reached here through four redundant layers — the existence check, the
# regular-file check, the readability check, and finally grep's own rc>=2 — so
# asserting rc==2 alone leaves the existence branch a SURVIVING MUTANT:
# deleting that branch outright still left this harness printing
# `all cases passed`, because a later layer caught the same fixture and
# reported a different reason. Found by mutation-testing during task 1's
# review. Asserting the branch's own wording is what makes deleting
# it fail here.
case "$OUT" in
  *"does not exist"*) pass "case 6: reports it through the existence branch" ;;
  *) fail "case 6: expected the 'does not exist' message, got: $OUT" ;;
esac

# ===========================================================================
# Case 7: a scoped file is a symlink — exit 2.
# ===========================================================================
new_root
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK"
OUTSIDE="$(mktemp "${TMPDIR:-/tmp}/check-dispatch-paragraphs-test-outside.XXXXXX")"
DIRS+=("$OUTSIDE")
printf '%s\n' "$REVIEWER_BLOCK" > "$OUTSIDE"
ln -s "$OUTSIDE" "$ROOT/skills/flow/review-panel.md"
run_guard
[ "$RC" -eq 2 ] && pass "case 7: a symlinked scoped file exits 2" \
  || fail "case 7: expected exit 2, got rc=$RC out=$OUT"
case "$OUT" in
  *"review-panel.md"*) pass "case 7: names the symlinked file" ;;
  *) fail "case 7: expected review-panel.md named in output, got: $OUT" ;;
esac

# Case 1 above already demonstrates the VERBATIM REPORT — THE FACT
# paragraph present and clean, alongside a correct REPRODUCE block, exiting
# 0 — no separate case is needed for that.

# ===========================================================================
# Case 8: the VERBATIM REPORT — THE FACT block is missing entirely from
# review-panel.md — exit 1, names the file and the required block count.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK"
run_guard
[ "$RC" -eq 1 ] && pass "case 8: exits 1" || fail "case 8: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"review-panel.md"*"VERBATIM REPORT"*) pass "case 8: names review-panel.md and the missing VERBATIM REPORT block" ;;
  *) fail "case 8: expected review-panel.md and VERBATIM REPORT named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 9: the VERBATIM REPORT block is present but missing
# "the reviewer's own report" — exit 1, names the phrase.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK

$VERBATIM_BLOCK_NO_REVIEWERS_OWN_REPORT"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK"
run_guard
[ "$RC" -eq 1 ] && pass "case 9: exits 1" || fail "case 9: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"the reviewer's own report"*) pass "case 9: names the missing phrase" ;;
  *) fail "case 9: expected \"the reviewer's own report\" named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 10: the VERBATIM REPORT block is present but missing
# "never a source of fact" — exit 1, names the phrase.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK

$VERBATIM_BLOCK_NO_NEVER_A_SOURCE_OF_FACT"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK"
run_guard
[ "$RC" -eq 1 ] && pass "case 10: exits 1" || fail "case 10: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"never a source of fact"*) pass "case 10: names the missing phrase" ;;
  *) fail "case 10: expected \"never a source of fact\" named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 11: the VERBATIM REPORT block is present but missing
# "the report wins" — exit 1, names the phrase.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK

$VERBATIM_BLOCK_NO_REPORT_WINS"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK"
run_guard
[ "$RC" -eq 1 ] && pass "case 11: exits 1" || fail "case 11: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"the report wins"*) pass "case 11: names the missing phrase" ;;
  *) fail "case 11: expected \"the report wins\" named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 12: CHECK_DISPATCH_PARAGRAPHS_ROOT is set but empty — exit 2, names
# the refusal (distinct from leaving the variable unset, which defaults to
# scanning this repository itself).
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK

$VERBATIM_BLOCK"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK"
set +e
OUT="$(CHECK_DISPATCH_PARAGRAPHS_ROOT="" "$GUARD" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 2 ] && pass "case 12: a set-but-empty root exits 2" \
  || fail "case 12: expected exit 2, got rc=$RC out=$OUT"
case "$OUT" in
  *"set but empty"*) pass "case 12: names the set-but-empty refusal" ;;
  *) fail "case 12: expected the set-but-empty refusal named, got: $OUT" ;;
esac

# ===========================================================================
# Case 13: a required site path exists but is a directory, not a regular
# file — exit 2, names the refusal.
# ===========================================================================
new_root
mkdir -p "$ROOT/skills/flow/review-panel.md"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK"
run_guard
[ "$RC" -eq 2 ] && pass "case 13: a directory at a site path exits 2" \
  || fail "case 13: expected exit 2, got rc=$RC out=$OUT"
case "$OUT" in
  *"not a regular file"*) pass "case 13: names the not-a-regular-file refusal" ;;
  *) fail "case 13: expected the not-a-regular-file refusal named, got: $OUT" ;;
esac

# ===========================================================================
# Case 14: a required site path exists as a regular file but is not
# readable — exit 2, names the refusal. Skipped when running as root,
# since root ignores a file's own permission bits.
# ===========================================================================
if [ "$(id -u)" -ne 0 ]; then
  new_root
  write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK

$VERBATIM_BLOCK"
  write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK"
  chmod 000 "$ROOT/skills/flow/implement.md"
  run_guard
  chmod 644 "$ROOT/skills/flow/implement.md"
  [ "$RC" -eq 2 ] && pass "case 14: an unreadable site file exits 2" \
    || fail "case 14: expected exit 2, got rc=$RC out=$OUT"
  case "$OUT" in
    *"is not readable"*) pass "case 14: names the not-readable refusal" ;;
    *) fail "case 14: expected the not-readable refusal named, got: $OUT" ;;
  esac
else
  pass "case 14: skipped (running as root, permission bits are not enforced)"
fi

# ===========================================================================
# Case 15: two required blocks glued together with no blank line between
# them — the first (deficient) block must NOT absorb the second block's
# text and pass by borrowing its phrases. Regression for the block-text
# bleed across adjacent blockquote paragraphs found in KAN-217's review:
# extract_block_text used to keep walking through any '>' continuation
# line regardless of whether it started a second required block, so a
# missing-phrase reviewer block glued directly to a correct VERBATIM block
# passed clean by absorbing the VERBATIM block's own phrases.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "> **REPRODUCE, DON'T READ:** Where a behaviour crosses a boundary — the store, the filesystem, a
> guard, a real transcript, a real process — please just read the code and trust it, worth less than one you did.
> **VERBATIM REPORT — THE FACT:** the reviewer's own report, never a source of fact, the report wins, exercise the real thing"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK"
run_guard
[ "$RC" -eq 1 ] && pass "case 15: a glued deficient block still exits 1" \
  || fail "case 15: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"exercise the real thing"*) pass "case 15: names the phrase the glued block did not actually carry" ;;
  *) fail "case 15: expected 'exercise the real thing' named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 16: the FOREGROUND BUILDS label is absent entirely from
# implement.md — exit 1, names the file and the missing block.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK

$VERBATIM_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK"
run_guard
[ "$RC" -eq 1 ] && pass "case 16: exits 1" || fail "case 16: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"implement.md"*"FOREGROUND BUILDS"*) pass "case 16: names implement.md and the missing FOREGROUND BUILDS block" ;;
  *) fail "case 16: expected implement.md and FOREGROUND BUILDS named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 17: a FOREGROUND BUILDS block is present but missing
# "still executing in the background" — exit 1, names the phrase.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK

$VERBATIM_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK

$FOREGROUND_BLOCK_NO_STILL_EXECUTING

$FOREGROUND_BLOCK"
run_guard
[ "$RC" -eq 1 ] && pass "case 17: exits 1" || fail "case 17: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"still executing in the background"*) pass "case 17: names the missing phrase" ;;
  *) fail "case 17: expected \"still executing in the background\" named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 18: a FOREGROUND BUILDS block is present but missing
# "Run it in the foreground" — exit 1, names the phrase.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK

$VERBATIM_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK

$FOREGROUND_BLOCK_NO_RUN_FOREGROUND

$FOREGROUND_BLOCK"
run_guard
[ "$RC" -eq 1 ] && pass "case 18: exits 1" || fail "case 18: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"Run it in the foreground"*) pass "case 18: names the missing phrase" ;;
  *) fail "case 18: expected \"Run it in the foreground\" named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 19: a FOREGROUND BUILDS block is present but missing
# "poll it to completion" — exit 1, names the phrase.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK

$VERBATIM_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK

$FOREGROUND_BLOCK_NO_POLL_COMPLETION

$FOREGROUND_BLOCK"
run_guard
[ "$RC" -eq 1 ] && pass "case 19: exits 1" || fail "case 19: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"poll it to completion"*) pass "case 19: names the missing phrase" ;;
  *) fail "case 19: expected \"poll it to completion\" named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 20: implement.md carries exactly ONE correct FOREGROUND BUILDS
# block; the second required occurrence is entirely missing (not
# deficient — simply absent) — exit 1, names implement.md and the
# min-blocks violation at its own threshold (2).
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK

$VERBATIM_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK

$FOREGROUND_BLOCK"
run_guard
[ "$RC" -eq 1 ] && pass "case 20: exits 1" || fail "case 20: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"implement.md"*) pass "case 20: names implement.md" ;;
  *) fail "case 20: expected implement.md named in output, got: $OUT" ;;
esac
case "$OUT" in
  *"requires at least 2 block(s) carrying the label \"**FOREGROUND BUILDS:**\", found 1"*) \
    pass "case 20: names the min-blocks violation at its own threshold" ;;
  *) fail "case 20: expected the FOREGROUND BUILDS min-blocks violation message, got: $OUT" ;;
esac

# ===========================================================================
# Case 21: review-panel.md carries exactly ONE correct FOREGROUND BUILDS
# block; the second required occurrence is entirely missing — exit 1,
# names review-panel.md and the min-blocks violation at its own threshold
# (2). Same shape as case 20, other site.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK

$VERBATIM_BLOCK

$FOREGROUND_BLOCK"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK"
run_guard
[ "$RC" -eq 1 ] && pass "case 21: exits 1" || fail "case 21: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"review-panel.md"*) pass "case 21: names review-panel.md" ;;
  *) fail "case 21: expected review-panel.md named in output, got: $OUT" ;;
esac
case "$OUT" in
  *"requires at least 2 block(s) carrying the label \"**FOREGROUND BUILDS:**\", found 1"*) \
    pass "case 21: names the min-blocks violation at its own threshold" ;;
  *) fail "case 21: expected the FOREGROUND BUILDS min-blocks violation message, got: $OUT" ;;
esac

# ===========================================================================
# Case 22: the TARGETED TESTS label is absent entirely from implement.md —
# exit 1, names the file and the missing block.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK

$VERBATIM_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK"
run_guard
[ "$RC" -eq 1 ] && pass "case 22: exits 1" || fail "case 22: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"implement.md"*"TARGETED TESTS"*) pass "case 22: names implement.md and the missing TARGETED TESTS block" ;;
  *) fail "case 22: expected implement.md and TARGETED TESTS named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 23: a TARGETED TESTS block is present but missing "the build tool's
# own selector" — exit 1, names the phrase.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK

$VERBATIM_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK_NO_SELECTOR"
run_guard
[ "$RC" -eq 1 ] && pass "case 23: exits 1" || fail "case 23: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"implement.md"*"the build tool"*"own selector"*) pass "case 23: names implement.md and the missing phrase" ;;
  *) fail "case 23: expected implement.md and the build tool's own selector named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 24: a TARGETED TESTS block is present but missing "once for RED,
# once for GREEN" — exit 1, names the phrase.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK

$VERBATIM_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK_NO_RED_GREEN"
run_guard
[ "$RC" -eq 1 ] && pass "case 24: exits 1" || fail "case 24: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"implement.md"*"once for RED, once for GREEN"*) pass "case 24: names implement.md and the missing phrase" ;;
  *) fail "case 24: expected implement.md and 'once for RED, once for GREEN' named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 25: a TARGETED TESTS block is present but missing "module or
# repository suite mid-task" — exit 1, names the phrase.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK

$VERBATIM_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK_NO_SUITE"
run_guard
[ "$RC" -eq 1 ] && pass "case 25: exits 1" || fail "case 25: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"implement.md"*"module or repository suite mid-task"*) pass "case 25: names implement.md and the missing phrase" ;;
  *) fail "case 25: expected implement.md and 'module or repository suite mid-task' named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 26: the TARGETED TESTS label is absent entirely from
# review-panel.md — exit 1, names the file and the missing block.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK

$VERBATIM_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK"
run_guard
[ "$RC" -eq 1 ] && pass "case 26: exits 1" || fail "case 26: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"review-panel.md"*"TARGETED TESTS"*) pass "case 26: names review-panel.md and the missing TARGETED TESTS block" ;;
  *) fail "case 26: expected review-panel.md and TARGETED TESTS named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 27: the MUTATION PROOF label is absent entirely from
# review-panel.md — exit 1, names the file and the missing block.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK

$VERBATIM_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK"
run_guard
[ "$RC" -eq 1 ] && pass "case 27: exits 1" || fail "case 27: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"review-panel.md"*"MUTATION PROOF"*) pass "case 27: names review-panel.md and the missing MUTATION PROOF block" ;;
  *) fail "case 27: expected review-panel.md and MUTATION PROOF named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 28: a MUTATION PROOF block is present but missing "mutation-proved
# before you end your turn" — exit 1, names the phrase.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK

$VERBATIM_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK

$MUTATION_BLOCK_NO_MUTATION_PROVED"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK"
run_guard
[ "$RC" -eq 1 ] && pass "case 28: exits 1" || fail "case 28: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"mutation-proved before you end your turn"*) pass "case 28: names the missing phrase" ;;
  *) fail "case 28: expected 'mutation-proved before you end your turn' named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 29: a MUTATION PROOF block is present but missing "confirm an
# existing test fails, and restore" — exit 1, names the phrase.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK

$VERBATIM_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK

$MUTATION_BLOCK_NO_CONFIRM_AND_RESTORE"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK"
run_guard
[ "$RC" -eq 1 ] && pass "case 29: exits 1" || fail "case 29: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"confirm an existing test fails, and restore"*) pass "case 29: names the missing phrase" ;;
  *) fail "case 29: expected 'confirm an existing test fails, and restore' named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 30: a MUTATION PROOF block is present but missing "a surviving
# mutant" — exit 1, names the phrase.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK

$VERBATIM_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK

$MUTATION_BLOCK_NO_SURVIVING_MUTANT"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK"
run_guard
[ "$RC" -eq 1 ] && pass "case 30: exits 1" || fail "case 30: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"a surviving mutant"*) pass "case 30: names the missing phrase" ;;
  *) fail "case 30: expected 'a surviving mutant' named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 42: a MUTATION PROOF block is present but missing "confirm the edit
# landed" — exit 1, names the phrase.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK

$VERBATIM_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK

$MUTATION_BLOCK_NO_EDIT_LANDED"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK"
run_guard
[ "$RC" -eq 1 ] && pass "case 42: exits 1" || fail "case 42: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"confirm the edit landed"*) pass "case 42: names the missing phrase" ;;
  *) fail "case 42: expected 'confirm the edit landed' named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 43: a MUTATION PROOF block is present but missing "a refusal, not a
# surviving mutant" — exit 1, names the phrase.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK

$VERBATIM_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK

$MUTATION_BLOCK_NO_REFUSAL"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK"
run_guard
[ "$RC" -eq 1 ] && pass "case 43: exits 1" || fail "case 43: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"a refusal, not a surviving mutant"*) pass "case 43: names the missing phrase" ;;
  *) fail "case 43: expected 'a refusal, not a surviving mutant' named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 44: a MUTATION PROOF block is present but missing "never buys a
# test" — exit 1, names the phrase.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK

$VERBATIM_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK

$MUTATION_BLOCK_NO_NEVER_BUYS"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK"
run_guard
[ "$RC" -eq 1 ] && pass "case 44: exits 1" || fail "case 44: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"never buys a test"*) pass "case 44: names the missing phrase" ;;
  *) fail "case 44: expected 'never buys a test' named in output, got: $OUT" ;;
esac

# review-panel.md and implement.md content matching case 1's fully-correct
# fixtures, reused as the baseline for cases 31-34 below so only the one
# TOOLS deficiency under test stands out.
CLEAN_REVIEW_PANEL="$REVIEWER_BLOCK

$VERBATIM_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK

$MUTATION_BLOCK

$TOOLS_BLOCK

$TOOLS_BLOCK

$HANDSHAKE_BLOCK

$HANDSHAKE_BLOCK

$INDEPENDENT_BLOCK

$DELEGATION_BLOCK

$DELEGATION_BLOCK"

CLEAN_IMPLEMENT="$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK

$TOOLS_BLOCK

$TOOLS_BLOCK

$HANDSHAKE_BLOCK

$HANDSHAKE_BLOCK

$DELEGATION_BLOCK"

# ===========================================================================
# Case 31: the TOOLS label is absent entirely from brainstorm.md (its
# default seeded by new_root is overridden with prose and no block) —
# exit 1, names brainstorm.md and the missing TOOLS block.
# ===========================================================================
new_root
write_site "skills/flow/brainstorm.md" "No TOOLS paragraph here at all, just prose."
write_site "skills/flow/review-panel.md" "$CLEAN_REVIEW_PANEL"
write_site "skills/flow/implement.md" "$CLEAN_IMPLEMENT"
run_guard
[ "$RC" -eq 1 ] && pass "case 31: exits 1" || fail "case 31: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"brainstorm.md"*"TOOLS"*) pass "case 31: names brainstorm.md and TOOLS" ;;
  *) fail "case 31: expected brainstorm.md and TOOLS named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 32: implement.md carries one correct TOOLS_BLOCK plus one variant
# missing "in your first turn" — exit 1, names the dropped phrase.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$CLEAN_REVIEW_PANEL"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK

$TOOLS_BLOCK

$TOOLS_BLOCK_NO_FIRST_TURN"
run_guard
[ "$RC" -eq 1 ] && pass "case 32: exits 1" || fail "case 32: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"in your first turn"*) pass "case 32: names the dropped phrase" ;;
  *) fail "case 32: expected 'in your first turn' named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 33: implement.md carries one correct TOOLS_BLOCK plus one variant
# missing "never a wildcard query" — exit 1, names the dropped phrase.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$CLEAN_REVIEW_PANEL"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK

$TOOLS_BLOCK

$TOOLS_BLOCK_NO_WILDCARD"
run_guard
[ "$RC" -eq 1 ] && pass "case 33: exits 1" || fail "case 33: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"never a wildcard query"*) pass "case 33: names the dropped phrase" ;;
  *) fail "case 33: expected 'never a wildcard query' named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 34: implement.md carries one correct TOOLS_BLOCK plus one variant
# missing "re-prices your whole context" — exit 1, names the dropped
# phrase.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$CLEAN_REVIEW_PANEL"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK

$TOOLS_BLOCK

$TOOLS_BLOCK_NO_REPRICES"
run_guard
[ "$RC" -eq 1 ] && pass "case 34: exits 1" || fail "case 34: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"re-prices your whole context"*) pass "case 34: names the dropped phrase" ;;
  *) fail "case 34: expected 're-prices your whole context' named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 35: the MODEL HANDSHAKE label is absent entirely from implement.md —
# exit 1, names the file and the missing block.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$CLEAN_REVIEW_PANEL"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK

$TOOLS_BLOCK

$TOOLS_BLOCK"
run_guard
[ "$RC" -eq 1 ] && pass "case 35: exits 1" || fail "case 35: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"implement.md"*"MODEL HANDSHAKE"*) pass "case 35: names implement.md and the missing MODEL HANDSHAKE block" ;;
  *) fail "case 35: expected implement.md and MODEL HANDSHAKE named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 36: the MODEL HANDSHAKE label is absent entirely from
# review-panel.md — exit 1, names the file and the missing block.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK

$VERBATIM_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK

$MUTATION_BLOCK

$TOOLS_BLOCK

$TOOLS_BLOCK"
write_site "skills/flow/implement.md" "$CLEAN_IMPLEMENT"
run_guard
[ "$RC" -eq 1 ] && pass "case 36: exits 1" || fail "case 36: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"review-panel.md"*"MODEL HANDSHAKE"*) pass "case 36: names review-panel.md and the missing MODEL HANDSHAKE block" ;;
  *) fail "case 36: expected review-panel.md and MODEL HANDSHAKE named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 37: the MODEL HANDSHAKE label is absent entirely from brainstorm.md
# (its default seeded by new_root is overridden with prose and no block) —
# exit 1, names brainstorm.md and the missing MODEL HANDSHAKE block.
# ===========================================================================
new_root
write_site "skills/flow/brainstorm.md" "No MODEL HANDSHAKE paragraph here at all, just prose."
write_site "skills/flow/review-panel.md" "$CLEAN_REVIEW_PANEL"
write_site "skills/flow/implement.md" "$CLEAN_IMPLEMENT"
run_guard
[ "$RC" -eq 1 ] && pass "case 37: exits 1" || fail "case 37: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"brainstorm.md"*"MODEL HANDSHAKE"*) pass "case 37: names brainstorm.md and MODEL HANDSHAKE" ;;
  *) fail "case 37: expected brainstorm.md and MODEL HANDSHAKE named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 38: the MODEL HANDSHAKE label is absent entirely from
# verify-and-handoff.md (its default seeded by new_root is overridden) —
# exit 1, names verify-and-handoff.md and the missing block.
# ===========================================================================
new_root
write_site "skills/flow/verify-and-handoff.md" "No MODEL HANDSHAKE paragraph here at all, just prose."
write_site "skills/flow/review-panel.md" "$CLEAN_REVIEW_PANEL"
write_site "skills/flow/implement.md" "$CLEAN_IMPLEMENT"
run_guard
[ "$RC" -eq 1 ] && pass "case 38: exits 1" || fail "case 38: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"verify-and-handoff.md"*"MODEL HANDSHAKE"*) pass "case 38: names verify-and-handoff.md and MODEL HANDSHAKE" ;;
  *) fail "case 38: expected verify-and-handoff.md and MODEL HANDSHAKE named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 39: implement.md carries one correct HANDSHAKE_BLOCK plus one variant
# missing "the first line of your first reply" — exit 1, names the dropped
# phrase.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$CLEAN_REVIEW_PANEL"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK

$TOOLS_BLOCK

$TOOLS_BLOCK

$HANDSHAKE_BLOCK

$HANDSHAKE_BLOCK_NO_FIRST_LINE"
run_guard
[ "$RC" -eq 1 ] && pass "case 39: exits 1" || fail "case 39: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"the first line of your first reply"*) pass "case 39: names the dropped phrase" ;;
  *) fail "case 39: expected 'the first line of your first reply' named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 40: implement.md carries one correct HANDSHAKE_BLOCK plus one variant
# missing "and nothing else on that line" — exit 1, names the dropped
# phrase.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$CLEAN_REVIEW_PANEL"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK

$TOOLS_BLOCK

$TOOLS_BLOCK

$HANDSHAKE_BLOCK

$HANDSHAKE_BLOCK_NO_NOTHING_ELSE"
run_guard
[ "$RC" -eq 1 ] && pass "case 40: exits 1" || fail "case 40: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"and nothing else on that line"*) pass "case 40: names the dropped phrase" ;;
  *) fail "case 40: expected 'and nothing else on that line' named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 41: implement.md carries one correct HANDSHAKE_BLOCK plus one variant
# missing "before any tool call" — exit 1, names the dropped phrase.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$CLEAN_REVIEW_PANEL"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK

$TOOLS_BLOCK

$TOOLS_BLOCK

$HANDSHAKE_BLOCK

$HANDSHAKE_BLOCK_NO_BEFORE_TOOL_CALL"
run_guard
[ "$RC" -eq 1 ] && pass "case 41: exits 1" || fail "case 41: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"before any tool call"*) pass "case 41: names the dropped phrase" ;;
  *) fail "case 41: expected 'before any tool call' named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 45: the INDEPENDENT PASSES label is absent entirely from
# review-panel.md (KAN-472 task 20) — exit 1, names the file and the
# missing block.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK

$VERBATIM_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK

$MUTATION_BLOCK

$TOOLS_BLOCK

$TOOLS_BLOCK

$HANDSHAKE_BLOCK

$HANDSHAKE_BLOCK"
write_site "skills/flow/implement.md" "$CLEAN_IMPLEMENT"
run_guard
[ "$RC" -eq 1 ] && pass "case 45: exits 1" || fail "case 45: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"review-panel.md"*"INDEPENDENT PASSES"*) pass "case 45: names review-panel.md and the missing INDEPENDENT PASSES block" ;;
  *) fail "case 45: expected review-panel.md and INDEPENDENT PASSES named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 46: the NO DELEGATION label is absent entirely from implement.md
# (KAN-484) — exit 1, names the file and the missing block.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$CLEAN_REVIEW_PANEL"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK

$TOOLS_BLOCK

$TOOLS_BLOCK

$HANDSHAKE_BLOCK

$HANDSHAKE_BLOCK"
run_guard
[ "$RC" -eq 1 ] && pass "case 46: exits 1" || fail "case 46: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"implement.md"*"NO DELEGATION"*) pass "case 46: names implement.md and the missing NO DELEGATION block" ;;
  *) fail "case 46: expected implement.md and NO DELEGATION named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 47: review-panel.md carries one correct NO DELEGATION block plus one
# variant missing "Never call the `Agent` tool" (KAN-484) — exit 1, names
# the dropped phrase.
# ===========================================================================
new_root
write_site "skills/flow/implement.md" "$CLEAN_IMPLEMENT"
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK

$VERBATIM_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK

$MUTATION_BLOCK

$TOOLS_BLOCK

$TOOLS_BLOCK

$HANDSHAKE_BLOCK

$HANDSHAKE_BLOCK

$INDEPENDENT_BLOCK

$DELEGATION_BLOCK

$DELEGATION_BLOCK_NO_NEVER_CALL_AGENT"
run_guard
[ "$RC" -eq 1 ] && pass "case 47: exits 1" || fail "case 47: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"review-panel.md"*'missing the required phrase: "Never call the `Agent` tool"'*) \
    pass "case 47: names the dropped phrase" ;;
  *) fail "case 47: expected the dropped phrase named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 48: review-panel.md carries one correct NO DELEGATION block plus one
# variant missing "never spawn a subagent" (KAN-484) — exit 1, names the
# dropped phrase.
# ===========================================================================
new_root
write_site "skills/flow/implement.md" "$CLEAN_IMPLEMENT"
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK

$VERBATIM_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK

$MUTATION_BLOCK

$TOOLS_BLOCK

$TOOLS_BLOCK

$HANDSHAKE_BLOCK

$HANDSHAKE_BLOCK

$INDEPENDENT_BLOCK

$DELEGATION_BLOCK

$DELEGATION_BLOCK_NO_NEVER_SPAWN"
run_guard
[ "$RC" -eq 1 ] && pass "case 48: exits 1" || fail "case 48: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"review-panel.md"*'missing the required phrase: "never spawn a subagent"'*) \
    pass "case 48: names the dropped phrase" ;;
  *) fail "case 48: expected the dropped phrase named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 49: review-panel.md carries one correct NO DELEGATION block plus one
# variant missing "the leaf of this run" (KAN-484) — exit 1, names the
# dropped phrase.
# ===========================================================================
new_root
write_site "skills/flow/implement.md" "$CLEAN_IMPLEMENT"
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK

$VERBATIM_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK

$MUTATION_BLOCK

$TOOLS_BLOCK

$TOOLS_BLOCK

$HANDSHAKE_BLOCK

$HANDSHAKE_BLOCK

$INDEPENDENT_BLOCK

$DELEGATION_BLOCK

$DELEGATION_BLOCK_NO_LEAF"
run_guard
[ "$RC" -eq 1 ] && pass "case 49: exits 1" || fail "case 49: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"review-panel.md"*'missing the required phrase: "the leaf of this run"'*) \
    pass "case 49: names the dropped phrase" ;;
  *) fail "case 49: expected the dropped phrase named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 50: the NO DELEGATION label is absent entirely from
# verify-and-handoff.md (its default seeded by new_root is overridden with
# only TOOLS and MODEL HANDSHAKE, KAN-484) — exit 1, names the file and the
# missing block. The other two sites stay clean.
# ===========================================================================
new_root
write_site "skills/flow/verify-and-handoff.md" "$TOOLS_BLOCK

$HANDSHAKE_BLOCK"
write_site "skills/flow/review-panel.md" "$CLEAN_REVIEW_PANEL"
write_site "skills/flow/implement.md" "$CLEAN_IMPLEMENT"
run_guard
[ "$RC" -eq 1 ] && pass "case 50: exits 1" || fail "case 50: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"verify-and-handoff.md"*"NO DELEGATION"*) pass "case 50: names verify-and-handoff.md and the missing NO DELEGATION block" ;;
  *) fail "case 50: expected verify-and-handoff.md and NO DELEGATION named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 51: review-panel.md carries exactly ONE correct NO DELEGATION block;
# the second required occurrence is entirely missing (KAN-484) — exit 1,
# names review-panel.md and the min-blocks violation at its own threshold
# (2).
# ===========================================================================
new_root
write_site "skills/flow/implement.md" "$CLEAN_IMPLEMENT"
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK

$VERBATIM_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK

$MUTATION_BLOCK

$TOOLS_BLOCK

$TOOLS_BLOCK

$HANDSHAKE_BLOCK

$HANDSHAKE_BLOCK

$INDEPENDENT_BLOCK

$DELEGATION_BLOCK"
run_guard
[ "$RC" -eq 1 ] && pass "case 51: exits 1" || fail "case 51: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"review-panel.md"*'requires at least 2 block(s) carrying the label "**NO DELEGATION:**", found 1'*) \
    pass "case 51: names the min-blocks violation at its own threshold" ;;
  *) fail "case 51: expected the NO DELEGATION min-blocks violation message, got: $OUT" ;;
esac

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'all cases passed\n'
