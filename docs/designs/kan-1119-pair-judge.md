# kan-1119-pair-judge
> Design record of change kan-1119-pair-judge (KAN-1119), written at landing. Each decision's own **Status:** line is the status.

## Context

Today a dispatch's `{model, effort}` pair comes from one of four sources, none of which looks at the
work:

- **The role's first base pair.** This covers the gated reviewer and re-reviewer bundles,
  `late-fix`, `design-reviewer`, the self-review and pipeline-fix loops, and every Explore, Plan
  and planner dispatch through `flow-guard plan-agent-pair`.
- **A name-hash roll.** `plan-class` prints `effort_roll = sha256("<name>effort") mod 100`. Below
  90, the implementer, the fixer and every group take `low` (`medium` on a hard seam). The pass-1
  panel dispatches take `low` below 60 and `medium` otherwise. Two more rolls pick `sonnet` or
  `opus` for the always-on extra dispatches.
- **A count-based scorer.** `plan-agent-pair … planner auto <round-file>` scores `**Scope:**`,
  architectural headings and path counts.
- **A fixed pair.** A verification-only group always runs on `opus`/`low`, and the rerun pair is
  always at `low`.

`sonnet/flow-low` is an admitted implementer pair, yet sonnet carries about 5% of spend. Count
signals misjudge (a mass rename touches many files yet is trivial), and a roll knows nothing at
all about the work.

This is one change because the judge, the role it runs on, the record of its pick, the removal of
the rolls it replaces, and the rule text at every dispatch site that reads a pair all move
together. If the rolls were removed without the judge, Decide would have no rule. If the judge
were added without the sites calling it, nothing would change.

Bounds that stay in force, because the judge only picks inside them:

- admission by `flow-guard check-dispatch-pair` / `check-loop-dispatch`;
- `session_cap` and session overrides;
- `flow record decision`'s refusal of an implementer, fixer or group pair at `high`;
- marked pairs (`No-progress escalation`, `Operator-approved high`) and opt-in pairs;
- the no-progress escalation, **Three failed rounds** and **Caps never stop**
  (`skills/flow-contracts/pipeline.md`), all unchanged. A pick that is too weak costs one failed
  round, and the existing escalation takes over from there.

This change reverses KAN-1110 fix round 2's "model choice stays the row's".

## Recommended approach

Add a Go guard, `flow-guard pair-judge <role> <digest-file> [-opus-only]`
(`stats/internal/guard/pairjudge.go`). It narrows a judged role's allowed pairs to a candidate list
and, when two or more candidates remain, asks a headless `claude -p` judge, running on the new
`pair-judge` role's first pair, to pick one. It prints one line,
`<model> <effort> <judge|fallback|single> <reason>`.

The guard has three outcomes:

- **One candidate:** printed as `single`, with no model call.
- **The judge answers:** its pick is printed as `judge`.
- **Any failure:** the first candidate is printed as `fallback`, with the cause in the reason, and
  the guard exits 0. Failures are a missing `claude`, a non-zero exit, a 60 s timeout, an
  unparseable answer, a pick outside the candidates, or no `pair-judge` pair to run on.

A judge failure therefore never blocks a run (decisions `fallback-first-candidate`,
`missing-judge-row-falls-back`).

**Candidates:**

- Start from the role's base pairs, with any open experiment's overlay applied.
- Drop `xhigh`/`max`.
- Drop `high` for `implementer` and every fixer: `panel-fix`, `self-review-fix`, `pipeline-fix-fix`.
- Under `-opus-only`, drop every non-`opus` pair, unless that leaves none (`opus-only-yields-when-empty`).
- Marked and opt-in pairs are never candidates (`judge-candidates-capped`).
- A judged role with no candidate left exits 2.

**The judge's input** is a fixed system prompt plus the digest, and nothing else. The system prompt
carries:

- the role's `Does` line, read from the roles table of `skills/flow-contracts/model-policy.md`
  (`does-line-read-from-the-roles-table`);
- the candidates;
- the signals to weigh: verification-only, docs, config or test-only work, and scope (counts are
  hints, never the verdict);
- the instruction to pick the cheapest pair near-certain to succeed on the first try;
- the answer format `PAIR: <pair>` / `REASON: <one line>`.

The `claude -p` call runs with no tools, no setting sources and no MCP. The digest is capped at
200 KB, with a truncation marker (`judge-timeout-60s-digest-200kb`). A command seam,
`FLOW_PAIR_JUDGE_CMD`, lets tests stub the model call (`judge-command-seam`).

**Which roles the guard judges.** It holds the judged set and exits 2 on any other role
(`guard-refuses-unjudged-role`):

- planners: `planner`, `self-review-plan`, `pipeline-fix-plan`;
- implementer: `implementer`;
- fixers: `panel-fix`, `self-review-fix`, `pipeline-fix-fix`;
- reviewers: `design-reviewer`, `gated-reviewer`, `gated-re-reviewer`, `panel-reviewer`,
  `panel-sanity-efficiency`, `panel-mutation-failure`, `panel-rerun`, `late-fix`,
  `self-review-pass`, `self-review-review-1`, `self-review-review-r`, `pipeline-fix-review-1`,
  `pipeline-fix-review-r`;
- planning agents: `explorer`, `plan-designer`.

`operator-asked`, `visual-verifier`, `tooling-analyst`, `fix-run-judge` and `panel-merger` keep
today's rule (`judged-set`, `judged-set-by-function`). With today's rows, only `planner`,
`self-review-plan`, `pipeline-fix-plan`, `implementer`, `panel-reviewer`,
`panel-sanity-efficiency`, `panel-mutation-failure`, `panel-rerun`, `explorer` and `plan-designer`
have two or more candidates. Every other judged role prints `single`, so its pick stays unchanged
until the operator widens its row (`single-candidate-skips-judge`).

**Who calls it.** Each caller writes the digest from what it already holds (`digest-is-the-prompt`):

- **Decide** (`skills/flow/brainstorm-planner.md`) makes one call per recorded pair. The digest is
  the plan's tasks (a group's own tasks, for a group), plus `plan-class`'s `inputs:` line, plus a
  panel dispatch's slots. The calls are:
  - the implementer as `implementer`;
  - the fixer as `panel-fix`;
  - each group as `implementer`, with `-opus-only` on a hard seam (`hard-seam-opus-only-flag`);
  - each `panel.dispatches` entry as `panel-reviewer`, with `-opus-only` on a hard seam;
  - `panel.rerun_dispatch` as `panel-rerun`;
  - `extra_dispatch` / `extra_dispatch_mutation` as `panel-sanity-efficiency` /
    `panel-mutation-failure`, on every class, micro included.

  The pass-1 panel, extra and rerun pairs are judged once there, from the plan
  (`panel-judged-at-decide`). A micro panel still records no pair and runs on the literal `opus`
  (`micro-panel-stays-opus`). A fix run's carried decision row judges each new group the same way
  (`carried-row-groups-judged`).
- **The planner dispatch** (`skills/flow/brainstorm.md`) calls `pair-judge planner`. In round 1 the
  digest is the ticket or seeded note; in later rounds, the latest round file.
- **The self-review and pipeline-fix plan dispatches** call the guard with their own role, the
  findings as digest. Their planner judges the first review's pair over its `tasks.md` and records
  it as the loop's `first-review-effort` decision (`first-review-pair-judged-by-planner`).
- **Explore and Plan dispatches** call `pair-judge explorer|plan-designer` on the dispatch prompt.
- **Every other judged dispatch** calls `pair-judge <role> <prompt-file>`. This covers gated
  bundles, `late-fix`, `design-reviewer`, and the loops' pass, fix and review dispatches
  (`loop-reviews-through-guard`).
- `/flow-fast` and `/flow-interactive` call it the same way.

**Recording.**

- In `decision.json`, each judged pair's `reason` reads `judge: …`, `fallback: …` or `single: …`
  (`decision-reason-carries-source`).
- On the dispatch row, `flow record dispatch begin -pair-pick '<source>: <reason>'` stores the
  same text in a new nullable `dispatches.pair_pick` column (`pair-pick-column`). A loop dispatch,
  whose row `flow-guard check-loop-dispatch` opens, carries a `Pair pick: <source>: <reason>`
  prompt line the guard forwards (`loop-pair-pick-prompt-line`). A dispatch on a recorded decision
  pair passes none: the decision row's `reason` is its record (`decision-pairs-carry-no-pair-pick`).
- A `session_cap` or session override applies on top of the pick. A pick above the cap becomes the
  first candidate within it, with the reason `override: …` (`cap-and-override-win`).
- No new guard checks that a dispatch ran on its judged pair (`no-dispatch-equals-pick-guard`).

**What goes.** The judge replaces every roll and fixed pick that chose among a judged role's pairs
(`judge-replaces-all-rolls`, `delete-dead-rolls`, `delete-planner-auto-scorer`):

- the implementer/fixer/group `effort_roll` rule, its planner-chosen band above 90, and the fixed
  verification-only group pair;
- the panel's 60/40 effort roll;
- both extra-dispatch model rolls (the `extra:` lines now print their bundle alone);
- the rerun pair's planner-chosen model and its "effort `low`, fixed" sentence. `panel-rerun`'s
  pairs are all `low`, so the row is the bound (`row-is-the-bound`);
- `plan-class`'s `effort` roll, and `decision.json`'s `rolls.effort`. `flow decision render` stops
  requiring it, and an older file that still carries it still renders;
- `plan-agent-pair`'s `auto <round-file>` form, `papAutoEffort`, and the no-effort form whose
  callers move to `pair-judge`. The explicit-effort form stays for the guard redispatches
  (`plan-agent-pair-keeps-effort-form`).

`high` still reaches a pass-1 panel dispatch only through the plan gate's **Raise `<dispatch>` to
`high`** question, as a marked pair that is never judged. The `extra-dispatch-sonnet-carve-out`
still holds through the rows: `panel-reviewer`'s base pairs are all `opus`.

**The `pair-judge` role.** Migration `0065` (`migration-number-0065`) does two things:

- seeds `('pair-judge','base','sonnet/flow-low','',0)`, with no opt-in;
- adds `dispatches.pair_pick`.

The role is listed in `skills/flow-contracts/model-policy.md` under a new `### Pair choice` area,
with the row "Picks each dispatch's model and effort from its role's allowed pairs."
(`pair-judge-own-area`). It joins the lists and fixtures `fix-run-judge` joined — the store's seed
test, the guard's policy fixture, the roles table and its test, the Experiments view — but not
`launchRoleID` or `roleDispatchSQL`, which map dispatch rows to roles: the judge is a `claude -p`
call and never a dispatch (`judge-role-has-no-dispatch-rows`).

**Where the rule is canonical.** The judge's rule (the call, its flags, the recording and the
cap-and-override step) is written once, in **Model and effort**
(`skills/flow/brainstorm-planner.md`), which already declares itself canonical for how a
dispatch's model is chosen. Every other site names the `pair-judge` call for its role and cites
that section (`judge-rule-canonical-in-model-and-effort`).

**zcode and muse** call the same guard. Their **Harness mapping** replaces the pick at dispatch
(`no-harness-special-case`). The visual verifier's Go pick in `visualverifypass_prepare.go` keeps
its first pair (`visual-verifier-not-a-caller`).

Rejected alternatives, each recorded in its decision:

- a skill-prose judge, or a Go rubric with no model (`judge-runs-as-go-guard`);
- judging every base pair (`judge-candidates-capped`);
- hard seams left to the judge's prompt (`hard-seam-opus-only-flag`);
- keeping the auto scorer (`delete-planner-auto-scorer`);
- a separate digest builder (`digest-is-the-prompt`);
- the `notes` column (`pair-pick-column`);
- a new decision field (`decision-reason-carries-source`);
- a harness switch (`no-harness-special-case`);
- enforcement that the dispatch matches the pick (`no-dispatch-equals-pick-guard`);
- unbounded inputs (`judge-timeout-60s-digest-200kb`);
- rolls kept beside the judge (`judge-replaces-all-rolls`);
- per-dispatch-time panel judging (`panel-judged-at-decide`);
- a guard-side effort filter (`row-is-the-bound`);
- a reviewer-specific path (`loop-reviews-through-guard`);
- judging any role given (`guard-refuses-unjudged-role`);
- a shared Go caller for the visual verifier (`visual-verifier-not-a-caller`);
- judging a micro panel (`micro-panel-stays-opus`);
- a `-cap` flag (`cap-and-override-win`).

## Critical files

Go: guard, store, records, CLI.

- `stats/internal/guard/pairjudge.go` — new: the `pair-judge` guard, its contract in the leading
  comment, registered in `guard.Registry`.
- `stats/internal/guard/pairjudge_test.go` — new: candidates, `single`, `judge` through the stubbed
  `FLOW_PAIR_JUDGE_CMD`, every fallback cause, `-opus-only`, the refused role, the digest cap.
- `stats/internal/guard/modelpolicy.go` — `policyTable`, the pair reader with the experiment
  overlay, which the guard reads its candidates from.
- `stats/internal/guard/planagentpair.go`, `stats/internal/guard/plan_agent_pair_test.go` — delete
  the `auto` form, `papAutoEffort` and the no-effort form; keep the explicit-effort form.
- `stats/internal/guard/planclass.go`, `stats/internal/guard/plan_class_test.go` — drop the
  `rolls:` line's `effort`; the two `extra:` lines print their bundle alone.
- `stats/cmd/flow/decision.go`, `stats/cmd/flow/decision_test.go` — `rolls.effort` no longer
  required; an older file carrying it still renders.
- `stats/internal/store/migrations/0065_pair_judge.sql` — new: the `pair-judge` base row and the
  nullable `dispatches.pair_pick` column.
- `stats/internal/records/types.go` — `records.Dispatch.PairPick`.
- `stats/internal/store/records.go`, `stats/internal/store/records_test.go` — the dispatch insert
  and read paths persist and return `pair_pick`.
- `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go` — `dispatch begin -pair-pick`, an
  optional flag, begin-only in `dispatch backfill`.
- `stats/internal/guard/loop.go`, `stats/internal/guard/check_loop_dispatch_test.go` — a loop
  dispatch's `Pair pick:` line forwarded as `-pair-pick`.
- `stats/internal/guard/guard.go` — `Env.PairJudgeTimeout`, the test-only deadline override.
- `stats/internal/records/types.go` — also `GuardHelpers` gains `pair-judge`, which refuses nothing.
- `stats/internal/guard/dispatchparagraphs.go`, `stats/internal/guard/check_dispatch_paragraphs_test.go`
  — PLANNING DELEGATION's required phrase becomes `flow-guard pair-judge`.
- `stats/internal/store/rolepairs_test.go`, `stats/internal/modelpolicy/modelpolicy_test.go`,
  `stats/internal/guard/testdata/policy.json` — the role lists and fixtures `fix-run-judge` joined.
- `stats/internal/repotest/planner_pair_test.go` — its planner-pair pin follows the judge;
  `stats/internal/repotest/pair_judge_sites_test.go` — new: every judged site's call.

Web, seed and hooks.

- `stats/web/src/views/Experiments.tsx` — `SHORT_DOES` entry for `pair-judge`.
- `hooks/enforce-agent-baseline.py`, `stats/internal/repotest/enforce_agent_baseline_hook_test.go`
  — read only: the judge is not an Agent dispatch, so no key pattern is added.
- `hooks/enforce-loop-dispatch.py` — read only: a judged loop pick is still "off its row".

Skills: the pair rule at each site.

- `skills/flow-contracts/model-policy.md` — the `### Pair choice` area and `pair-judge` row; the
  guard list under **Experiments**; the decision-pair sentences that name rolls or fixed rows.
- `skills/flow/brainstorm-planner.md` — **Explore and design agents** (pair from `pair-judge`),
  **Decide** steps 2–5, **Model and effort** (the canonical judge rule), and the **Plan review
  gate** raise question's "over the roll's effort".
- `skills/flow/brainstorm.md` — the planner and planning-agent pair lines and the auto form; the
  guard redispatches keep `plan-agent-pair … planner low|medium`.
- `skills/flow/document-fix.md` — the carried-row groups and the "rolls come out identical"
  sentence.
- `skills/flow/implement.md` — the gated reviewer pair and the implementer stub "on its row".
- `skills/flow/gated-review-fix.md` — the gated re-reviewer pair.
- `skills/flow/review-panel-late-fix.md` — the `late-fix` reviewer pair.
- `skills/flow/SKILL-rationale.md` — the `effort_roll` and 60/40 rationale paragraphs.
- `skills/flow-plan/SKILL.md` — planning-agent pair; the `plan-class` mention.
- `skills/flow-self-review/SKILL.md` — the self-review pass, plan, fix and review pairs, and the
  planner's planning-agent pair.
- `README.md` — the **Classify** step's `plan-class` line, plus the `pair-judge` command.
- Unchanged: `skills/flow/review-panel.md` (its rerun and extra pairs are still the decision's),
  `skills/flow-contracts/pipeline.md` (no sentence there picks among a role's base pairs),
  `skills/flow-fast/SKILL.md` (it runs **Decide** and the dispatch sites as written; the compact,
  experimental and bundle rolls its change basename keys remain) and
  `skills/flow-interactive/SKILL.md` (it dispatches `operator-asked` alone).

## Existing code to reuse

- `stats/internal/guard/modelpolicy.go` `policyTable` — the role's rows, with the experiment
  overlay and the base/marked/opt-in split. The candidates are its base row.
- `stats/internal/guard/modelpolicy.go` `pcFlow` and the `flow policy get` cache — the guard's only
  route to pairs, so flowd down falls back to the cache as every guard does.
- `stats/internal/guard/guard.go` `Registry` — the guard's registration.
- `hooks/enforce-agent-baseline.py`'s `DISPATCH_PAIR_CHECK` env override — the seam pattern
  `FLOW_PAIR_JUDGE_CMD` follows: an env var naming the command to run in place of the real one.
- Migration `stats/internal/store/migrations/0063_fix_run_judge_pair.sql` and every file it
  touched (listed above) — the template for adding the `pair-judge` role.
- The `dispatches.round_base` migration (`0046`) and its path through `records.Dispatch`,
  `stats/internal/store/records.go` and `flow record dispatch begin -round-base` — the template for
  a nullable `dispatches` column, `TestDispatchRoundBaseRoundTrips` the test to copy.
- `stats/internal/modelpolicy`'s `Rows` — the roles table parser the guard reads a role's `Does`
  line through.
- `stats/internal/guard/plan_agent_pair_test.go` / `helpers_test.go`'s `fakeFlowEnv` /
  `runGuard` — the test harness the new guard's tests use.

## Verification

- **Go:** `cd stats && make -s web-dist && go vet ./... && gofmt -l .` is clean. `go test` passes
  on the guard package (`-run 'PairJudge|PlanClass|PlanAgentPair|CheckLoopDispatch|CheckDispatchParagraphs'`),
  `stats/cmd/flow` (`-run 'Decision|RecordDispatch'`), the store (`-run 'RolePairs|Migrat|PairPick'`)
  and `stats/internal/repotest` (`-run 'PairJudgeSites|PlannerPair'`).
- **SPA:** `cd stats/web && npx --no-install tsc -b` is clean.
- **Prose guards:** `scripts/flow-guard run-verify-list <worktree> lint` is green.
- **The guard by hand, on a fixture policy:**
  - `pair-judge implementer <digest>` with `FLOW_PAIR_JUDGE_CMD` stubbed to answer
    `PAIR: sonnet/flow-low` prints `sonnet low judge …`.
  - The same call stubbed to exit 1 prints the first candidate as `fallback`.
  - `-opus-only` never prints a `sonnet` pair.
  - `pair-judge visual-verifier <digest>` exits 2.
  - `pair-judge panel-fix <digest>`, on a test-local policy whose `panel-fix` base is only
    `opus/flow-low`, prints `single` without invoking the stub.
- **Decide on a real plan:** `plan-class` over this change's own `tasks.md` prints a `rolls:` line
  with no `effort` and `extra:` lines with no roll or pair. `flow decision render` renders a
  `decision.json` without `rolls.effort`, and one with it.
- **End to end:** the `## Live check` below.

## Live check

The change touches flowd's store (migration `0065`, a new dispatch column) and a guard that calls a
real model. The check runs against a copy of the dev store, never the store itself, using
`scripts/flow-guard live-store-copy up kan1119-live` run from this worktree, which builds `flowd`
from this branch so the copy carries `0065`. It is torn down with
`scripts/flow-guard live-store-copy down kan1119-live`.

**Before** (dev store on 4173, read only):

1. Record `flow policy get`'s answer: no `pair-judge` role.
2. Record what today's rule picks for the two digests below: `implementer`'s first base pair
   (`opus` at the roll's effort).

**After** (the copy, with `FLOW_ADDR` exported from the `up` output):

1. `flow policy get` lists `pair-judge` with base `sonnet/flow-low`.
2. With the worktree's `scripts/flow-guard`, run `pair-judge implementer <digest>` on two digests
   under `flow state dir`:
   - **(a)** a three-task plan that renames one identifier across forty Markdown files;
   - **(b)** a two-task plan that adds a mutex-guarded cache to a concurrent store write path.

   Record each printed line. Both must show source `judge`, a pair from `implementer`'s
   candidates, and no `high`.
3. `pair-judge panel-fix <digest-a>` prints source `single`.
4. `pair-judge visual-verifier <digest-a>` exits 2.
5. `flow record dispatch begin -change kan-1119-pair-judge -role explorer -key explore-99 -model
   sonnet -effort low -session-token mf-kan1119-live -pair-pick 'judge: live check'`, then read the
   row back with `flow record dispatches -change kan-1119-pair-judge` on the copy. Its `pairPick`
   reads `judge: live check`.

**What not working looks like:**

- a `fallback` source on (a) or (b) — the judge was never reached; its reason names the cause;
- a pick outside the candidates, or a `high` effort;
- `single` on `implementer`;
- a dispatch row whose `pair_pick` is empty;
- `pair-judge` missing from the copy's `flow policy get`.

A difference between (a)'s and (b)'s picks is recorded but not required, since the judge may
rightly pick the same pair for both. The check proves that the judge runs and that its pick is
recorded, not which pair it picks.

## Non-goals

- No new enforcement that a dispatch ran on its judged pair (`no-dispatch-equals-pick-guard`).
- `operator-asked`, `visual-verifier`, `tooling-analyst`, `fix-run-judge`, `panel-merger` and any
  later role keep today's rule until a change adds them.
- No change to any role's pairs. Widening a single-pair row so the judge has a choice is the
  operator's edit in the Experiments view.
- Marked and opt-in pairs are untouched. `high` on a panel dispatch stays the plan gate's raise.
- No harness switch in the guard.
- The no-progress escalation and the round caps are unchanged.
- The `compact`, `experimental` and `bundle` rolls are unchanged.

## Risks

- **Nested `claude -p`.** The judge spawns `claude -p` from inside a Claude Code session (a Bash
  call of the parent or a planner). If the nested CLI refuses to run there, or inherits settings
  despite the flags, every pick falls back to the first candidate, which is today's behaviour. The
  run is never blocked, but the change is inert. It shows as `fallback:` reasons in `pair_pick`
  and `decision.json`, and the live check fails on it. Task 3's first step runs the nested call live
  from a Claude Code Bash call before the guard is written, and records the working flag set.
- **The window before reinstall.** Until `./setup.sh global` and `make -C stats restart` apply
  `0065` to the dev store, `flow policy get` has no `pair-judge` row. The guard then falls back
  (`missing-judge-row-falls-back`) rather than failing every dispatch site.
- **Latency and cost.** Decide makes up to about seven judge calls, and each other judged dispatch
  with two or more candidates makes one. Each call is bounded at 60 s and runs on
  `sonnet/flow-low`. Roles with a single candidate make no call.
- **No determinism.** Re-deciding a plan may record different pairs than the first row did; the
  name-hash rolls gave identical ones. A re-decide reads the newest row, so the run follows
  whichever pick was recorded.
- **A too-weak pick** costs one failed round, after which the existing no-progress escalation takes
  over. The two hard filters (no `high` for implementer/fixer, `-opus-only` on a hard seam) bound
  the worst case.
- **The digest leaves the session.** The digest goes to the same model provider every dispatch
  already uses, capped at 200 KB, so this adds no new exposure.

## Decisions

### How the pair judge runs

**ID:** judge-runs-as-go-guard
**Status:** active
**Chosen:** A Go guard, `flow-guard pair-judge`, that runs headless `claude -p` on the
`pair-judge` row's pair. Operator, round 1: "A. Go guard + headless claude -p (Recommended)".
**Considered:** A judge written as skill prose that each dispatcher reasons through itself — this
spends the dispatcher's own (often opus) context on every pick, and nothing records it the same
way twice. A deterministic Go rubric with no model — it rests on count signals again, which is
the misjudgement the ticket is about.

### Which roles the judge covers

**ID:** judged-set
**Status:** active
**Chosen:** Planners, implementers, reviewers, fixers, the self-review roles, `explorer` and
`plan-designer`. Every other role keeps today's rule. Operator, round 1: "only planners,
implementers, reviewers, fixers, self-review", clarified as the set the judge covers ("reviewers'
pairs included"), "and explorer, plan-designer included".
**Considered:** Excluding reviewers so that review keeps a fixed pair — the operator named
reviewers inside the set. Judging every role — `operator-asked` and `visual-verifier` were named
outside it.

### Reading the judged set by function

**ID:** judged-set-by-function
**Status:** active
**Chosen:** `pipeline-fix-plan`, `-fix`, `-review-1` and `-review-r` count as planner, fixer and
reviewers. `fix-run-judge`, `panel-merger` and `tooling-analyst` are not judged.
**Considered:** Judging only the literal role names the operator's answer used — this would leave
the pipeline-fix loop, which does the same jobs, on a different rule.

### A single candidate skips the judge

**ID:** single-candidate-skips-judge
**Status:** active
**Chosen:** A role with one candidate pair is printed with source `single`, with no model call.
**Considered:** Calling the judge on every dispatch — this spends a model call to choose from a
list of one.

### Which pairs are candidates

**ID:** judge-candidates-capped
**Status:** active
**Chosen:** The candidates are the role's base pairs, with the experiment overlay applied. `xhigh`
and `max` are always dropped, and `high` is dropped for `implementer` and every fixer (`panel-fix`, `self-review-fix`,
`pipeline-fix-fix`). Marked and
opt-in pairs are never candidates.
**Considered:** Letting the judge pick any base pair — this would let it pick `high` where
`flow record decision` refuses it, and `xhigh`/`max`, which stay grant-only.

### Hard seams are enforced by a flag

**ID:** hard-seam-opus-only-flag
**Status:** active
**Chosen:** The planner marks a hard seam and passes `-opus-only`. The guard then drops every
non-`opus` candidate.
**Considered:** Leaving "a hard seam stays on opus" to the judge's prompt — a rubric line can be
ignored, and the filter cannot.

### Delete the planner's count-based scorer

**ID:** delete-planner-auto-scorer
**Status:** active
**Chosen:** `plan-agent-pair`'s `auto <round-file>` form and `papAutoEffort` are deleted.
**Considered:** Keeping the count-based scorer as a hint source for the judge — its counts are the
misjudging signal, and the round file it reads is already the planner's digest.

### `plan-agent-pair` keeps only its explicit-effort form

**ID:** plan-agent-pair-keeps-effort-form
**Status:** active
**Chosen:** The no-effort form's callers (Explore, Plan and planner dispatches) move to
`pair-judge`, and that form is deleted. The explicit-effort form stays for the guard redispatches
`skills/flow/brainstorm.md` makes at `planner low` / `planner medium`.
**Considered:** Keeping the no-effort form beside the judge — it would leave a second,
work-blind route to the same roles' pairs.

### The digest is what the caller already holds

**ID:** digest-is-the-prompt
**Status:** active
**Chosen:** Outside Decide, the digest is the dispatch prompt itself (plus the ticket or latest
round file for the planner). Decide's digest is the tasks plus `plan-class`'s `inputs:` line, and a
panel dispatch's slots.
**Considered:** A separate digest builder — it would be a second summary of what the prompt
already says, with nothing to keep the two in step.

### Where the pick is recorded on a dispatch

**ID:** pair-pick-column
**Status:** active
**Chosen:** A new nullable `dispatches.pair_pick` column, set by
`flow record dispatch begin -pair-pick '<source>: <reason>'`. The flag is optional, so unjudged
sites record nothing.
**Considered:** Reusing the `notes` column — notes are free text, and the stats could not count
`judge`/`fallback`/`single` picks from them reliably.

### Where the pick is recorded in a decision

**ID:** decision-reason-carries-source
**Status:** active
**Chosen:** Each judged pair's existing `reason` reads `judge: …`, `fallback: …` or `single: …`.
**Considered:** A new decision field — `flow decision render` and `flow record decision` would
both have to learn it, for text the `reason` already carries.

### No harness special case

**ID:** no-harness-special-case
**Status:** active
**Chosen:** zcode and muse call the same guard, and their **Harness mapping** replaces the pick at
dispatch, as it replaces every recorded pair today.
**Considered:** A harness switch in the guard — it would duplicate the mapping that already lives
in `skills/flow-contracts/model-policy.md`.

### No check that a dispatch used its pick

**ID:** no-dispatch-equals-pick-guard
**Status:** active
**Chosen:** No new guard compares the dispatched pair with the judged one. Admission,
`session_cap` and the caps stay the checks.
**Considered:** A new enforcement path — the ticket did not ask for it, and a cap or override
legitimately departs from the pick.

### Bounds on the judge call

**ID:** judge-timeout-60s-digest-200kb
**Status:** active
**Chosen:** A 60 s timeout. The digest is truncated at 200 KB with a marker, keeping the judge
input well under 100k tokens.
**Considered:** Unbounded inputs — a large plan or prompt would make the judge cost more than the
dispatch it sizes, and a hung CLI would stall the run.

### Failure falls back to the first candidate

**ID:** fallback-first-candidate
**Status:** active
**Chosen:** A missing `claude`, a non-zero exit, a timeout, an unparseable answer or a pick outside
the candidates prints the first candidate with source `fallback` and a reason naming the cause,
and exits 0. Exit 2 is kept for a usage error, an unreadable policy, or a role outside the judged
set.
**Considered:** Failing the dispatch site on a judge failure — the judge is an optimisation, and
the first candidate is today's pick, so failing would block a run that would otherwise proceed.

### A missing `pair-judge` row falls back

**ID:** missing-judge-row-falls-back
**Status:** active
**Chosen:** When the policy carries no `pair-judge` pair, for example before `0065` reaches the
dev store, the guard prints the first candidate as `fallback`, with the reason "no pair-judge
pair".
**Considered:** Exit 2 as an unreadable policy — the policy is readable, and refusing would break
every judged dispatch site in the window between landing and restart. This follows
`fallback-first-candidate`.

### The test seam for the model call

**ID:** judge-command-seam
**Status:** active
**Chosen:** `FLOW_PAIR_JUDGE_CMD` names the command run in place of `claude`, following the
`DISPATCH_PAIR_CHECK` pattern.
**Considered:** Tests that call the real CLI — they would be slow, non-deterministic and need
credentials.

### The judge replaces every roll

**ID:** judge-replaces-all-rolls
**Status:** active
**Chosen:** The judge replaces the implementer/fixer/group `effort_roll` rule (with its band above
90 and the fixed verification-only group pair), the panel's 60/40 effort roll, both
extra-dispatch model rolls, and the rerun pair's planner-chosen model. This reverses KAN-1110 fix
round 2's "model choice stays the row's".
**Considered:** Keeping any roll beside the judge — two rules for one pick, one of which knows
nothing about the work.

### Delete the rolls nothing reads

**ID:** delete-dead-rolls
**Status:** active
**Chosen:** `plan-class` drops the `effort` roll and the two `extra:` rolls; its `extra:` lines
print their bundle alone. `decision.json` drops `rolls.effort`, and `flow decision render` stops
requiring it, while still rendering an older file that carries it.
**Considered:** Printing rolls nothing reads — a reader would take them for live inputs.

### Panel pairs are judged at Decide

**ID:** panel-judged-at-decide
**Status:** active
**Chosen:** The pass-1 panel, extra and rerun pairs are judged once at Decide, from the plan, and
recorded in `decision.json`.
**Considered:** Judging each one at dispatch time from the branch diff — this would add a judge
call per panel dispatch and per rerun, and move the record of the pick out of the decision row.

### The row is the effort bound for rerun and extras

**ID:** row-is-the-bound
**Status:** active
**Chosen:** The rerun's and the extra dispatches' "effort `low`" holds through their rows, which
are all `low` today.
**Considered:** A per-role effort filter in the guard — it would duplicate a bound the operator
already sets in the row.

### Loop reviews go through the guard

**ID:** loop-reviews-through-guard
**Status:** active
**Chosen:** The self-review and pipeline-fix review dispatches and the gated bundles call the
guard like every other judged role. Their single-pair rows keep today's pick (`single`).
**Considered:** A reviewer-specific path — a second route that would stop following the rows when
the operator widens one.

### The guard refuses roles it does not judge

**ID:** guard-refuses-unjudged-role
**Status:** active
**Chosen:** `pair-judge` holds the judged set and exits 2 on any other role.
**Considered:** Silently judging any role it is given — an unjudged role such as `operator-asked`
would drift onto the judge without the decision that adds it.

### The visual verifier is not a caller

**ID:** visual-verifier-not-a-caller
**Status:** active
**Chosen:** `stats/internal/guard/visualverifypass_prepare.go` keeps its first-pair pick, and no
Go code calls the judge.
**Considered:** A shared Go judge function that the visual verifier calls — `visual-verifier` is
outside the judged set, so the shared function had no second caller.

### A micro panel stays on `opus`

**ID:** micro-panel-stays-opus
**Status:** active
**Chosen:** A micro decision's panel records no pair and still runs on the literal `opus`. Only
its extra dispatches are judged.
**Considered:** Judging a pair micro never records — this would add a decision field to micro
for one dispatch.

### A cap or override applies on top

**ID:** cap-and-override-win
**Status:** superseded by cap-takes-highest-admitted
**Chosen:** A `session_cap` or session override applies on top of the judge's pick, as it does on
today's pick. A pick above the cap is replaced by the first candidate within it, with the reason
`override: …`.
**Considered:** A `-cap` flag on the guard — the cap is applied by the caller today, and the
guard would need a second input for it.
**Superseded because:** the round-0 review panel found its "first candidate within it" rule contradicting the canonical **`session_cap`** rule of `skills/flow-contracts/pipeline.md`, and the panel's fix changed `skills/flow/brainstorm-planner.md` to the canonical rule without superseding this entry (kan-1119 self-review, F5).

### A carried decision row judges its new groups

**ID:** carried-row-groups-judged
**Status:** active
**Chosen:** On a fix run whose class is unchanged, `skills/flow/document-fix.md`'s carried row
judges each recomputed group the same way Decide step 5 does: `pair-judge implementer` on the
group's tasks, with `-opus-only` on a hard seam. This replaces "the row's implementer pair (`opus`
at `medium` for a hard seam)". Auto-taken.
**Considered:** Keeping the row's implementer pair and the fixed `opus`/`medium` hard-seam pair —
this would keep a fixed pick that `judge-replaces-all-rolls` removes everywhere else.

### Where the judge rule is canonical

**ID:** judge-rule-canonical-in-model-and-effort
**Status:** active
**Chosen:** The call, its flags, the recording and the cap-and-override step are written once, in
**Model and effort** (`skills/flow/brainstorm-planner.md`). Every other site names its
`pair-judge` call and cites that section. Auto-taken.
**Considered:** `skills/flow-contracts/model-policy.md` as the home — that file is canonical for
which pairs a role admits, not for how one is chosen, and **Model and effort** already declares
itself canonical for the choice.

### The judge's role sits in its own area

**ID:** pair-judge-own-area
**Status:** active
**Chosen:** `pair-judge` is listed under a new `### Pair choice` area of **Allowed pairs**, with
the Does line "Picks each dispatch's model and effort from its role's allowed pairs."
**Considered:** Listing it under `### Planning` — it runs at every stage, not only planning, and
the Experiments view groups roles by area.

### Migration number

**ID:** migration-number-0065
**Status:** active
**Chosen:** The role and column migration is `0065`, because `0064` is taken
(`0064_guard_run_reasons.sql`).
**Considered:** Round 1's `0064` — it collides with the existing migration.

### Re-decide is not cached

**ID:** redecide-not-cached
**Status:** active
**Chosen:** A re-decide calls the judge again; nondeterminism across re-decides is accepted, and
the run follows the newest recorded row.
**Considered:** Caching the judge's answer keyed by the digest's hash in the decision row — it
restores identical re-decides, but adds stored state and a cache key for a case where the newest
row already decides what the run follows.

### A loop dispatch's pick reaches its row through a prompt line

**ID:** loop-pair-pick-prompt-line
**Status:** active
**Chosen:** A loop dispatch's prompt carries `Pair pick: <source>: <reason>`, and
`flow-guard check-loop-dispatch`, which opens every loop row, forwards it as `-pair-pick`.
Auto-taken at writing-plans.
**Considered:** No pick on loop rows — the hook opens them, so the session has no begin call to
pass the flag to, and the loops' judged planner picks would go unrecorded. A second recording call
after the hook's begin — two writes for one row.

### The loop planner judges the first review's pair

**ID:** first-review-pair-judged-by-planner
**Status:** active
**Chosen:** The self-review or pipeline-fix planner runs `flow-guard pair-judge
<loop>-review-1 <changeRoot>/tasks.md` and records the printed pair as its `first-review-effort`
decision; `check-loop-dispatch` keeps enforcing that decision unchanged. Auto-taken at
writing-plans.
**Considered:** The session judging review-1 at dispatch — a pick off the recorded effort is
refused by `check-loop-dispatch` whenever the row admits the recorded one, so the two rules would
fight. Deleting `first-review-effort` — a Go change to the loop guard the ticket did not ask for.

### A decision pair's dispatch row carries no pick

**ID:** decision-pairs-carry-no-pair-pick
**Status:** active
**Chosen:** Only a pair judged at dispatch is recorded with `-pair-pick`; a dispatch on a recorded
decision pair passes none, the decision row's `reason` being its record. Auto-taken at
writing-plans.
**Considered:** Copying the decision's `reason` onto every implementer and panel row — every
begin line of the implement and panel contracts changes, to store text the decision row already
holds.

### `-opus-only` yields when it would leave nothing

**ID:** opus-only-yields-when-empty
**Status:** active
**Chosen:** `-opus-only` is skipped when the role has no `opus` candidate — an open experiment's
override on another model then stands, as it stood over a hard seam before this change. Any other
filter leaving no candidate exits 2. Auto-taken at writing-plans.
**Considered:** Exit 2 on an empty `-opus-only` set — it would block Decide for every change with a
hard seam while such an experiment is open.

### The judge has no dispatch rows

**ID:** judge-role-has-no-dispatch-rows
**Status:** active
**Chosen:** `pair-judge` is not added to `launchRoleID` or `roleDispatchSQL`, which derive a role
from a dispatch row's role and key: the judge is a `claude -p` call, never a dispatch. Auto-taken
at writing-plans.
**Considered:** Mapping it as `fix-run-judge` is mapped — there is no key shape to map, so the
entries would select nothing.

### The judge reads a role's Does line from the roles table

**ID:** does-line-read-from-the-roles-table
**Status:** active
**Chosen:** The guard reads the role's `Does` line from `skills/flow-contracts/model-policy.md`
through `stats/internal/modelpolicy`'s `Rows`, under `FLOW_GUARD_REPO_ROOT`; an unreadable table
or row leaves the line out of the system prompt. Auto-taken at writing-plans, replacing plan
review 1's role-to-line map.
**Considered:** A role-to-line map in `pairjudge.go` — a second copy of the table's Does column,
which an edit to one would leave stale.

### A cap takes the highest pair the row admits under it

**ID:** cap-takes-highest-admitted
**Status:** active
**Chosen:** A `session_cap` or session override applies on top of the judge's pick, as it does on
today's pick. A pick above the cap is replaced by the highest pair the role's row admits at or
under it (**`session_cap`**, `skills/flow-contracts/pipeline.md`), with the reason `override: …`.
**Considered:** The first candidate within the cap — decision `cap-and-override-win`'s rule, which
contradicted the canonical **`session_cap`** rule.

## Open questions
