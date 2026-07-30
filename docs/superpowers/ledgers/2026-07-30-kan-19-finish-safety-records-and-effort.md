# SDD ledger — plan: openspec/changes/kan-19-finish-safety-records-and-effort/tasks.md

Merge base: a85a729ff120ba820688178e64ea3ac5eb58e607
Mode: myflow-do — NO COMMITS; per-task diffs are `git diff TASK_BASE`.

Task 1: implemented (uncommitted, model: opus) — check-finish-preflight.sh + harness, 15 assertions green.
Task 1: review 1 (model: sonnet) — spec ✅, 1 Important (signal (b) ordering vs base-ref resolution), 2 Minor.
Task 1: fix round 1/5 dispatched (model: opus) — ordering fix + new harness case + 2 minors.
Task 3: implemented dispatch (model: opus) — preserve-session-records.sh + harness. Ledger path .superpowers/sdd/tasks/progress.md confirmed by controller.
Task 1: fix round 1/5 (3 addressed, 0 open; re-review model: sonnet) — 16 assertions green, 3 lint guards green.
Task 1: complete (uncommitted, review clean, model: opus)
Task 3: implemented (uncommitted, model: opus) — 29 assertions green; fixed a cross-change glob collision the plan draft had.
Task 2: dispatched (model: opus) — contract + finish-skill wiring.
Task 3: review (model: sonnet) — spec OK, quality approved, 0 Critical/Important.
Task 3: minor (deferred): report said 29 assertions, live run shows 28 ok lines (cosmetic).
Task 3: minor (deferred): unwritable-destination case is a no-op if the suite runs as root (disclosed).
Task 3: minor (deferred): change names with glob metacharacters are not sanitized before `find -name` (not reachable under openspec naming).
Task 3: complete (uncommitted, review clean, model: opus)
Task 2: implemented (uncommitted, model: opus) — pipeline.md Finish contract + myflow-finish SKILL.md.
Task 2: review (model: sonnet) — spec FAIL, 1 Important (PR-CLI clause described as script behaviour).
Task 2: fix round 1/5 dispatched (model: opus).
Tasks 6+7: dispatched together (model: opus) — effort field; ## Effort section must precede Task 6's reference.
Task 2: fix round 1/5 (1 addressed, 0 open; re-review model: sonnet) — 3 lint guards green.
Task 2: complete (uncommitted, review clean, model: opus)
Tasks 6+7: implemented (uncommitted, model: opus). Flagged: `effort` still missing from carry-forward lists in myflow-do/SKILL.md, myflow-finish/SKILL.md, myflow-status/SKILL.md jq — routed to the Task 8 sweep.
Tasks 4+5: dispatched together (model: opus) — same two files, sequential inserts.
Tasks 6+7: review (model: sonnet) — spec OK, quality approved, 0 Critical/Important. Reviewer independently reproduced the check-references.sh adjacency/line-based weakness in a sandbox.
Tasks 6+7: minor (deferred): uneven line wrapping in state-file.md / state-self-heal.md prose (cosmetic).
Tasks 6+7: complete (uncommitted, review clean, model: opus)
SWEEP (Task 8) must also fix, confirmed by the reviewer against specs/myflow-effort/spec.md:
  - skills/myflow-do/SKILL.md:237 — add `effort` to the carry-forward list
  - skills/myflow-finish/SKILL.md:162 — add `effort` to the carry-forward list
  - skills/myflow-status/SKILL.md:39 — add `.effort` to the jq selector
Tasks 4+5: implemented (uncommitted, model: opus).
Tasks 4+5: review (model: sonnet) — spec OK, 1 Important (preserve call after `git add -A` in myflow-do), 1 Minor (enumeration drift).
Tasks 4+5: fix round 1/5 dispatched (model: opus).
Pre-existing, out of scope for this change: skills/myflow-finish/SKILL.md guardrail cites "step 2.1", which is not a heading in that file.

## RESUMED 2026-07-30
Tasks 4+5: fix round 1/5 applied (model: opus) — chose explicit re-staging over hoisting, because
  /myflow-do's staging is unconditional while its commit is not. Lint green.
Tasks 4+5: fix re-review dispatched (model: sonnet).
SWEEP (Task 8) also to carry: a clause in pipeline.md documenting why preserve precedes staging in
  finish (commit unconditional) and follows it in do (commit is the exception), so a later
  "harmonization" cannot reintroduce the bug.
Tasks 4+5: fix round 1/5 (2 addressed, 0 open; re-review model: sonnet) — reviewer independently
  endorsed the do/finish asymmetry on the merits. Lint green.
Tasks 4+5: complete (uncommitted, review clean, model: opus)
Task 8: dispatched (model: opus) — sweep + effort carry-forward + the asymmetry clause.
Task 8: implemented (uncommitted, model: opus) — sweep + 4 beyond-brief items.
Task 8: review (model: sonnet) — spec OK, quality approved, 0 findings; reviewer re-ran all 8 declared
  commands, openspec validate --strict, and diffed the working tree against task-8.diff.
Task 8: complete (uncommitted, review clean, model: opus)
All 8 tasks complete. 40/40 checkboxes.

## Panel pass 1
Slot 2 (principles, merged, sonnet): clean — 0/0/0.
Slot 0 (primary, sonnet): 1 IMPORTANT — pipeline.md's `## Model policy` still says the ledger is
  session-scoped and does not survive archival, and tells the reader that making it durable belongs
  to a future change. This change makes it durable. The delta spec
  specs/myflow-model-policy/spec.md REMOVES that requirement and ADDS the durable replacement, and
  proposal.md promises the rewrite — but tasks.md has no task for it. CONFIRMED by the controller at
  skills/myflow-contracts/pipeline.md `## Model policy` (the "This record is session-scoped."
  paragraph). This is a plan gap, not a plan contradiction: the work is in scope and unwritten.
Slot 5 (principles lens B, sonnet): 1 IMPORTANT — Single Source of Truth: the effort level
  descriptions and the "default medium" fact are hand-duplicated across state-file.md (canonical),
  myflow-start/SKILL.md's AskUserQuestion block, CLAUDE.md, AGENTS.md and myflow-status/SKILL.md.
  NOTE: the AskUserQuestion wording is PLAN-MANDATED verbatim by tasks.md Task 7 Step 2 — that half
  is the operator's call, not mine to override. The CLAUDE.md / AGENTS.md / myflow-status
  hardcodings are not plan-verbatim and can be pointed at the canonical default.
  1 MINOR (deferred): check-finish-preflight.sh's header restates pipeline.md's signal-order
  rationale; defensible as local context.
Slot 6 (principles lens C, sonnet): 1 IMPORTANT — check-finish-preflight.sh's `git status` dirty-count
  assignment is the one git call not guarded against failure; under `set -euo pipefail` a failure
  inside `VAR="$(a | b)"` aborts with git's own exit code, printing NO verdict and breaking the
  script's documented "verdict, or exit 2" contract. One-line fix, same idiom as the ancestor test.
  1 MINOR: myflow-finish/SKILL.md's verdict table does not name the exit-2/no-verdict branch.
Slot 3 (security, sonnet): 1 IMPORTANT — preserve-session-records.sh's `mkdir -p` is a no-op when a
  destination directory is already a SYMLINK, so `cp` follows it and writes the preserved record
  outside the repository. Empirically confirmed against a throwaway tree. docs/superpowers/* is
  ordinary PR-editable repo content, and the script runs automatically from /myflow-do's push path
  and /myflow-finish run 1 — a destination-controlled arbitrary-file-write primitive.
  1 MINOR: $NAME is interpolated unsanitized into the destination; traversal is currently blocked
  only incidentally by the `$TODAY-` prefix sharing the path component. Worth an explicit `/` guard.
  check-finish-preflight.sh: NO findings — `--end-of-options` verified on every caller-supplied ref;
  could not force a RUN2 on an unmerged or dirty tree.
Slot 1 (defect hunt, sonnet): 1 IMPORTANT — same unguarded `git status` pipeline as slot 6, found
  independently and reproduced with a git shim (exit 129, no output). Notes that a silent 0 would be
  worse: it would fall through to RUN2. No other Critical/Important; verified --end-of-options works
  and the digit-anchored glob prevents the cross-change adoption.
Slot 4 (adversarial, sonnet): 2 IMPORTANT — (a) CONTROLLER ERROR: final-review.diff was built with
  the human-gate's ':(exclude)openspec/' pathspec, so no slot saw tasks.md or the five spec deltas;
  slot 4 compensated by diffing the merge base directly. Diff regenerated without the exclusion
  (24 files, 2835 insertions). (b) harness blind spot: no case covers zero-commit AND fully clean —
  slot 4 deleted signal (b) from a scratch copy and the real harness still passed the mutant, which
  printed RUN2.
OPERATOR RULING on slot 5's Important: keep the plan-mandated AskUserQuestion wording; de-duplicate
  only the derived "default medium" claims in CLAUDE.md, AGENTS.md and myflow-status/SKILL.md.
Panel fix wave 1: dispatched ONE fixer (model: opus) with all six findings, deduped.
Panel fix wave 1: all 6 findings fixed (model: opus). TDD evidence for F1-F4 (each case red first;
  F4 mutation-verified). 3 lint guards + 5 harnesses + openspec validate --strict all exit 0.
  Parked with ruling: skills/myflow-start/SKILL.md still names `medium` literally in its revision
  sentence and handoff template — both are plan-mandated verbatim by Task 7 Steps 2 and 4, and the
  operator's ruling covered only the derived claims elsewhere. Real but deliberate.
Panel pass 2: FULL re-run (escalated, not asked) — the fix altered proposal.md, tasks.md and a
  canonical contract section, and pass 1's diff omitted openspec/ entirely, so every pass-1 clean
  result is stale for that content. Diff rewritten unfiltered.
Pass 2 slot 0 (primary, sonnet): CLEAN — 0 Critical, 0 Important. First slot to review openspec/:
  44/44 checkboxes verified against real diff content, all five spec deltas checked scenario by
  scenario, all six fix-wave findings confirmed closed, `unknown (agent-defined)` rule intact.
  Triaged all five deferred minors as non-blocking. Verdict: ready for the human gate.
Pass 2 slot 2 (principles merged, sonnet): CLEAN — 0/0/0. Model-policy rewrite matches the delta spec
  without duplicating it; `unknown (agent-defined)` rule intact and strengthened. No suppressions, no
  lint-config weakening, no bash-3.2 violations.
Pass 2 slot 1 (defect hunt, sonnet): 2 NEW IMPORTANT.
  (a) test-preserve-session-records.sh: TREES is a space-separated string iterated unquoted, so a
      TMPDIR containing a space leaks sandboxes — reproduced, 34 directories left behind, while the
      sibling harness (indexed array, with a header comment naming exactly this hazard) left 0. The
      fix wave applied that lesson to one harness and not the other.
  (b) preserve-session-records.sh: check-then-use gap in the NEW symlink fix — `dest_real` is
      resolved and validated, but `find`/`cp` still act through the original unresolved `$dest_dir`.
      Replacing that path with a symlink between check and copy re-opens the pass-1 escape. Build
      `dest` from `$dest_real` instead.
  1 MINOR: $NAME is only restricted from `/`, not from glob metacharacters, which can confuse the
      idempotency `find`. Cannot escape the worktree.
  Confirmed correct: the DIRTY empty-vs-one-line handling, the `git status` failure path and its
  shim-based case, signal-order case 1c, and that the `/` rejection precedes both use sites.
Pass 2 slot 5 (principles lens B, sonnet): 1 NEW IMPORTANT — specs/myflow-effort/spec.md states
  "`medium` the default" and re-describes each level, duplicating state-file.md's `## Effort` table,
  which the change declares canonical. Confirms the pass-1 de-duplication was real deferral, not
  indirection, and correctly did not re-raise the parked myflow-start items.
  CONTROLLER NOTE for the fix wave: the reviewer's cited precedent is spec->spec deferral, whereas
  this is spec->skill. In OpenSpec, openspec/specs/ are the NORMATIVE requirement source and skills
  implement them, so deleting the requirement from the spec would invert that direction. Route it as
  "name which file is canonical for which purpose", not "strip the spec".
Pass 2 slot 6 (principles lens C, sonnet): pass-1 Important CLOSED (STATUS_OUT captured with an
  explicit `|| exit 2` before wc/tr; cannot yield an empty/non-numeric DIRTY that falls through to
  RUN2) and the pass-1 Minor closed too.
  1 NEW IMPORTANT — both call sites document only preserve-session-records.sh's non-fatal
  missing-source case, and say nothing about what an agent does when the script exits non-zero for a
  refused or failed WRITE. By contrast check-finish-preflight.sh got a four-outcome table in the same
  wave. One line at each of the two call sites.
  1 MINOR — specs/myflow-finish-cleanup/spec.md does not describe the destination-containment refusal
  or the `/`-in-name rejection, though both are implemented and tested.
Pass 2 slot 4 (adversarial, sonnet): pass-1 findings BOTH confirmed closed — re-ran the mutation
  itself (deleted signal (b) in a scratch copy: harness now fails 3 cases incl. the zero-commit clean
  shape) and additionally mutated (a), (c), (d) and swapped (b)/(c); all caught except (a), below.
  1 NEW IMPORTANT — tasks.md's embedded "delivered" code for Tasks 1 and 3 is stale against what
  shipped: it shows the guard WITHOUT --end-of-options and with the raw unguarded dirty-count pipe,
  shows only 6 harness cases where 10+ shipped, and shows preserve-session-records.sh with neither
  the symlink refusal nor the `/` rejection. Only one fix-wave repair (Task 9) got a documented
  trail; four others have none. check-plan-provenance.sh is structurally blind to this — it checks
  that fenced blocks carry a tag, not that the code matches the shipped file.
  1 MINOR — signal (a)'s `-` sentinel branch is not distinctly exercised: deleting it leaves every
  case passing, because `-^{commit}` fails to resolve and produces the same REFUSE prefix. Asserting
  the reason text would close it.
Pass 2 slot 3 (security, sonnet): 1 CRITICAL + 2 IMPORTANT.
  CRITICAL — preserve-session-records.sh validates only the DESTINATION. Sources get no check at all:
    `[ -f "$src" ]` and `cp` both follow symlinks. All three sources sit under paths an attacker can
    plant (a tracked symlink survives .gitignore, which only gates untracked paths). Confirmed by
    experiment: a symlinked progress.md published a file containing a planted credential into
    docs/superpowers/ledgers/. Since the script runs automatically on /myflow-finish run 1 before
    `git add -A`, and the result is committed and pushed, this is arbitrary-file-read-and-publish.
  IMPORTANT — same TOCTOU slot 1 found, confirmed by winning the race 4 of 5 trials with a background
    process swapping the destination for a symlink: the script printed `preserved:` while the content
    landed outside the worktree.
  IMPORTANT — change-name glob metacharacters defeat the digit-anchored `find`: invoking with name `*`
    matched and OVERWROTE another change's preserved record. The anchor's own comment assumes NAME has
    no metacharacters; nothing enforces that.
  Confirmed clean: check-finish-preflight.sh's dirty-count fix (exit 2, no verdict line, cannot be
  read as a forged RUN2), --end-of-options coverage, .myflow/project.md containment, no secrets.
Panel fix wave 2: dispatching ONE fixer (model: opus) with 1 Critical + 6 Important + 2 folded Minor.
Panel fix wave 2: all 9 items fixed (model: opus). Reproduction-then-fix evidence for the Critical
  and both Importants; 42 new assertions, each red first; the signal-(a) mutation now fails 2 cases
  where it previously passed everything; both harnesses leave 0 sandboxes under a TMPDIR with a space
  (was 34). Controller re-verified independently: 3 lint guards, 5 harnesses and openspec validate
  --strict all exit 0 on this tree.
  Residual limitation stated in the script header rather than papered over: the destination race is
  narrowed, not closed — no path COMPONENT is re-followable, but `cp` cannot rule out the resolved
  directory itself being swapped.
Panel pass 3: FULL re-run of all 7 slots (escalated automatically — a new Critical surfaced in pass 2,
  the fix altered a delta spec, and it touched files outside the findings' named set).
Pass 3 slot 2 (principles merged, sonnet): CLEAN — 0/0/0. Confirms the three-outcome model is defined
  once in pipeline.md with both skills pointing at it; resolve_file is the simplest thing that works
  under bash 3.2 + BSD readlink and is covered by four dedicated cases; the effort split is genuinely
  canonical-per-purpose; no new suppressions and no lint-config weakening across either fix wave.
Pass 3 slot 1 (defect hunt, sonnet): CLEAN — 0 Critical, 0 Important. Hand-exercised every named
  vector against throwaway trees: leaf-symlink source refused (not skipped) with the other two still
  attempted; intermediate-component symlink refused; prefix-sibling destination (wt vs wt-evil)
  refused; symlink loop reported as skipped without hanging; relative, multi-hop and dangling targets
  correct; roots that are themselves symlinks resolved; allowlist rejects * ../../etc a/b .hidden
  -lead "a b" a;b `$(id)` [0-9] and accepts every real /myflow-start name shape; one refusal does not
  suppress the others. Both harnesses leak nothing under a spaced TMPDIR.
Pass 3 slot 0 (primary, sonnet): CLEAN — 0/0/0, ready to merge. Traced every pass-2 item (Critical +
  6 Important + 2 folded Minor) to a concrete verifiable fix rather than trusting the notes; re-ran
  both relevant harnesses (19/19 and 105/105) and the three lint guards. Judged the tasks.md
  marked-stale illustrative blocks an honest record given nothing can mechanically enforce code
  equivalence. Confirmed three deferred minors are now CLOSED by wave 2 (glob metacharacters, the
  missing spec scenarios, the signal-(a) reason text); three remain open and non-blocking.
Pass 3 slot 5 (principles lens B, sonnet): CLEAN — 0/0/0, and it WITHDREW its own pass-2 Important on
  the merits: the effort split is single-source-of-truth per purpose because the two files serve
  different readerships (design-time record vs the runtime table a command loads) and precedence is
  stated as a testable rule ("a table that contradicts the requirement is this file's defect"). Also
  found skip-vs-refuse clean in code (one branch point, before any boundary check) and resolve_file
  free of accidental state, bounded by a 40-hop cycle guard.
Pass 3 slot 6 (principles lens C, sonnet): CLEAN — 0/0/0, its pass-2 Important closed. Verified the
  three-outcome table matches the script exactly and that both call sites state the actionable rule
  inline rather than only citing it; every refusal names the path and what it resolved to and is never
  routed through `skipped:`; one refusal never blocks the other two sources; exit is 0 only when every
  source was preserved or skipped; and the spec does NOT claim the race is eliminated, only
  refusal-on-resolve-outside-root.
Pass 3 slot 4 (adversarial, sonnet): CLEAN — 0 Critical, 0 Important, ready for the human gate.
  Mutation set, all CAUGHT: delete signal (a) [its own pass-2 minor, now 2 failures], (b) [3], (c)
  [1], (d) [4], swap (b)/(c) [1]; remove the source-boundary check [12 failures incl. the
  planted-secret assertion], remove the destination-boundary check [10], revert the TOCTOU fix
  [caught deterministically by the case asserting the reported line names the RESOLVED directory —
  so the fix is verified structurally, not by a flaky race trial]. No surviving mutant, no test
  theatre. Judged the marked-stale plan blocks honest: the disclaimer is encountered BEFORE the code.
  2 MINOR (deferred): the guard is enforced by contract text only, like every other guard here — not
  a regression this change introduces; and specs/myflow-finish-cleanup/spec.md does not carry the
  script's own "narrowed, not closed" TOCTOU caveat, so a spec-only reader could think containment is
  total (the caveat does exist in the script header and in tasks.md Task 10).
Pass 3 slot 3 (security, sonnet): CLEAN — 0 Critical, 0 Important. Re-ran the original Critical attack
  (tracked symlink source -> planted secret): now refused, exit 1, nothing written. Also refused:
  intermediate-component symlink, relative target, prefix-sibling destination, and a NAME beginning
  with `-`. Symlink loops and broken links degrade safely to `skipped:` without hanging. Ran ~15
  racing trials against the residual TOCTOU window with two racer strategies and won 0 — several
  landed inside the window but never resolved through a planted symlink. Confirmed the header's
  "narrowed, not closed" wording is honest rather than overclaimed. Allowlist rejects every
  find-significant metacharacter. Secret scan clean; no world-writable files; 19/19 preflight cases.

## PANEL CLEAN — pass 3, all 7 slots, 0 Critical / 0 Important
