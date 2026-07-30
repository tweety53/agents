# SDD ledger — plan: openspec/changes/kan-20-widen-plan-provenance-guard-scan-scope/tasks.md
change: kan-20-widen-plan-provenance-guard-scan-scope
worktree: /Users/tweety53/Projects/agents-worktrees/openspec-kan-20-widen-plan-provenance-guard-scan-scope
merge-base: c1f5aa68ff64e1841d24842ba70b6565767639d6
mode: myflow /myflow-do — NO COMMITS; per-task diffs scoped by tree snapshot (git write-tree)
model policy: implementers Opus (myflow override of SDD's cheapest-tier guidance); panel slots Sonnet
base tree: e764c0dec3c144ebd2b8565eb7567d2861ee95ba
panel prep:
  [PRINCIPLES_PATH] /Users/tweety53/.claude/skills/myflow-do/engineering-principles.md (exists)
  [STANDARDS_PATHS] CLAUDE.md, AGENTS.md — both form 2 (bare, non-.mdc), resolved against the
    apply worktree root; normalized parent == project root, so both pass containment. None dropped.

Task 1: implemented (uncommitted; diff .superpowers/sdd/task-1.diff; model: opus) — review dispatched (model: sonnet)
  concern carried forward: scripts/check-vocabulary.sh blacklists retired panel vocabulary
  (e.g. "all six"). Fixture/prose in later tasks must avoid those terms; fix the prose, never the guard.
Task 1: complete (uncommitted, review clean — spec OK, quality approved, 0 Critical/Important; model: opus, reviewer: sonnet)
  note (not a defect): fixture prose changed from the brief to avoid check-vocabulary.sh's retired-term list.
  both reviewer "cannot verify" items resolved by controller: vocab trap carried into later dispatches;
  finditer/.start()/.end() dependency is exercised by Task 2's own tests.
task-2 base tree: 146e7a1b3888cebe3a016b13c1042779033fe246
Task 2: review CRITICAL — _is_quoted uses line-global parity, not nearest-enclosing-pair; fails OPEN.
  Verified by controller: 'a "quote" and a stray " mark, then 99 tests ran, "done"' exempts a bare claim.
  Root cause: the plan's own Step 3 block, tagged verified: over a friendly corpus only (never adversarial).
  Plan-mandated finding -> escalated to operator per SDD. Ruling: SPEC GOVERNS (fail closed); fix the predicate.
  fix round 1 base tree: 42f235e2f01bbef8435b84a26b9219dacd893981
Task 2: fix round 1/5 (1 addressed, 0 open; Critical fail-open resolved by nearest-enclosing-pair
  matching + class-wide veto; a second latent fail-open found and closed: quote inside a code span)
Task 2: IMPORTANT accepted with ruling — class-wide veto over-suppresses a genuinely closed pair when
  an unrelated stray delimiter of the same class sits elsewhere on the line. Ruling: ACCEPT.
  Spec mandates fail closed (operator ruled spec governs); the veto never admits a bare claim; the
  cost is a LOUD false positive fixable by rewording vs a SILENT false negative. Measured 0
  occurrences across 47 candidate files. Task 4 must document this property in plan-provenance.md.
Task 2: complete (uncommitted, re-review clean; model: opus, reviewers: sonnet)
task-3 base tree: 132c703fd32326101b37e1d0017b90a1b320e4fb
Task 3: complete (uncommitted, review clean — spec OK, quality approved, 0 Critical/Important;
  model: opus, reviewer: opus [scaled up: riskiest diff, containment + exit-code ranking])
  reviewer wrote 11 adversarial fixtures; containment per-filename, exit-3 rank, and
  no-abandon-on-refusal all confirmed, incl. the new intra-directory path.
  observed: this change's own design.md/proposal.md are now scanned (3 file(s)); they contain
  KAN-6 errors, covered by BOTH Task 1's lookbehind and Task 2's code-span exemption — the plan's
  fixes-before-widening ordering is now observed, not predicted.
Task 3: minor (deferred to Task 4): (1) containment docstring overstates - no basename assertion;
  (2) ProvenanceGuard docstring still says "per-directory loop"; (3) stale harness comment re
  path(s) vs tasks.md file(s); (4) ambiguous "no tasks.md, design.md, proposal.md found" wording;
  (5) ragged comment re-wrap; (6) TEST GAP: no case pins that a containment refusal on one candidate
  still lets a sibling candidate in the SAME dir be scanned; (7) chmod-000 dir now yields 3 messages
  where it yielded 1 - undocumented operator-visible change.
task-4 base tree: 621560b7e2117a2d72537ae903cb0e81fb427de4
Task 4: complete (uncommitted, review clean — spec OK, quality approved; model: opus, reviewer: sonnet)

PANEL (2 full passes, 4 fix waves, 3 scoped re-reviews) — final: 0 open Critical/Important.
  Pass 1: 2 Critical, 2 Important, 3 Minor -> fix wave 1 -> re-review found fail-open #4 (escapes)
  Fix round 2 (operator-approved): escape veto
  Pass 2: 1 Critical (#5 raw-HTML), 3 Important, 4 Minor -> fix wave 3 = DESIGN CHANGE
    (operator-approved): dropped code spans; every measured false positive was quote-delimited
  Re-review of wave 3 found fail-open #6 (tag grammar leaked on > inside a single-quoted attribute)
  Fix wave 4: deleted the HTML grammar entirely for a coarse "line contains <" veto.
  Root cause of all six: approximating a CommonMark grammar with regexes. No grammar now remains.
  Final fuzz: 0 fail-opens / 346,632 lines, non-vacuous (443 fail-opens on the pre-fix guard),
  token set enumerated in the report (includes the attribute forms that hid #6).
