# Tasks — panel-roster-follows-the-settings-list

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.

Four prose contracts, in dependency order: the resolver that reads the list (task 1), the panel
that dispatches it (task 2), the command whose own description of the roster becomes false
(task 3), and the rule for an id the running harness cannot spawn (task 4). `design.md` is canonical for every decision.

**Baseline, measured before any edit:**

- The settings store holds `defaultModel: sonnet` and four reviewer slots.
  <!-- measured: flow settings get @ 2026-08-27, after /flow-settings added bugbot -->
- Nothing reads `.reviewers`: the only settings reads in `skills/` are `SKILL.md`'s two lines, and
  neither takes that field.
  <!-- measured: grep -rn 'settings get' skills/ @ 1717d19 -->
- The CLI refuses an empty `-reviewers`.
  <!-- measured: read stats/cmd/flow/settings.go:123 @ 1717d19 -->

**Every task that grows an owned `.md` file runs `scripts/check-contract-budget.sh` and reconciles
its `budgets()` row in the same commit.** That guard is a ratchet: **raise a row only when the guard
actually fails on that file, and only to the size its own rule gives.** Raising a row it never
complained about loosens the ratchet silently, and nothing catches it.

---

- [x] 1. Resolve the roster beside the model

Extend `skills/flow/SKILL.md`'s **Model resolution** section to resolve the roster from the same
`flow settings get` call, so one read produces both values.

State all three resolution outcomes `design.md`'s **Resolution** table fixes: the list as-is when
the store answers with a non-empty one; `primary` alone when it answers with an empty one; and
`primary`, `principles`, `code-review-low` when the store cannot be reached — that last **named as
a fallback**, exactly as the existing text already names the literal `sonnet` fallback.

Say plainly that an empty list cannot be written through `/flow-settings`, citing
`stats/cmd/flow/settings.go`'s refusal, so the empty row is not read as an offer to disable review.

Do **not** restate what the panel does with the roster — task 2 owns that.

**Files:** `skills/flow/SKILL.md`, `scripts/check-vocabulary.sh`,
`scripts/test-check-vocabulary.sh`
**Tests:** `scripts/test-check-vocabulary.sh` — cases 6b through 6h

  **The guard and its harness are in this task's file set because resolving the roster forced a
  guard change, not as scope creep.** The live reviewer id `code-review-low` lexically contains the
  retired stage token `code-review`, so writing the resolution table tripped
  `check-vocabulary.sh`. A `vocab-guard:allow` marker was tried first and reverted — that guard's
  own header says a marker on a line telling the truth "teaches the guard to lie" — and the
  exemption was made structural instead, then narrowed again when a panel slot proved the first
  narrowing exempted every `code-review-*` compound.
  <!-- measured: scripts/test-check-vocabulary.sh, and grep -E probes against each retired token @ branch spectre/panel-roster-follows-the-settings-list -->
**Regression:** reverting this commit leaves `.reviewers` unread, so `/flow-settings` again writes a
field with no effect on any run.
**Baseline:** before=1 after=8 cases in `scripts/test-check-vocabulary.sh`'s code-review group;
41 assertions pass in the suite overall
<!-- measured: grep -oE 'case 6[b-h]' scripts/test-check-vocabulary.sh | sort -u, and scripts/test-check-vocabulary.sh | grep -c '^ok:' @ branch spectre/panel-roster-follows-the-settings-list -->
**Commit:** `feat(flow): resolve the panel roster from the settings store`
**Build:** green

- [x] 2. Dispatch the resolved roster

Rewrite `skills/flow/review-panel.md`'s roster so it dispatches the list task 1 resolves, replacing
the fixed **The three required slots** / **The two on-demand slots** split.

**Keep, unchanged:**
- each slot's own spawn shape — Bugbot and Security by `subagent_type`, no model override,
  recording `unknown (agent-defined)`;
- the per-run operator instruction that adds a slot for that run, from the round it was made and
  never retroactively to a closed pass, never written back to the store;
- the requirement that every slot supplies a reproducer per finding, and that the panel never hands
  off with an open finding of any severity.

**Cut, do not paraphrase**, the prose that only existed to justify a fixed roster — including the
slot-numbering note explaining a gap in the sequence, which describes a roster that no longer
exists. Per `CLAUDE.md`, never cut a normative sentence, an exit-code contract, an ordering
constraint, a scenario, a worked example, or a recorded reason a rejected alternative was rejected.

Record the new decision and mark the old one superseded per `design.md`'s
`supersedes-review-panel-fixed-3` — **without** editing the archived change that holds the original
entry.

**Files:** `skills/flow/review-panel.md`, `README.md`, `skills/flow/SKILL.md`,
`skills/flow/verify-and-handoff.md`, `AGENTS.md`, `CLAUDE.md`, `commands/flow.md`,
`commands-claude/flow.md`, `skills/README.md`, `stats/internal/store/settings.go`
**Tests:** none added — as task 1

  **The file set grew to nine because the panel found the fact stated in seven more places.**
  Retiring a fixed roster is not done when the mechanism changes — every live file asserting
  "fixed at 3 required slots" as current behaviour had to move with it, including
  `skills/flow/SKILL.md`'s own guardrails (which contradicted its own resolution section), the
  operator-facing handoff template in `skills/flow/verify-and-handoff.md`, and `CLAUDE.md` /
  `AGENTS.md`, which are injected into every agent session. `stats/internal/store/settings.go`'s
  doc comments were the last of them, missed by the first sweep because `check-vocabulary.sh`'s
  default targets never scan `stats/` — **comments only; `ValidReviewers` and `DefaultReviewers`
  themselves are untouched**, since that map is what `review-panel.md` cites as canonical.
  <!-- measured: grep -rn 'fixed at 3|fixed 3-slot' across the live corpus, before and after @ branch spectre/panel-roster-follows-the-settings-list -->

  **`README.md` is in this task's file set because the rename breaks its citations, not as scope
  creep.** Three sentences there name the headings this task replaces, and one of them records that
  diff-size and touched-area triggers were replaced by an explicit-only mechanism — a fact that
  survives this change and is re-pointed rather than cut.
  <!-- measured: scripts/check-references.sh against the task-2 working tree @ branch spectre/panel-roster-follows-the-settings-list -->
**Regression:** reverting this commit restores a roster written in prose, so a `/flow-settings`
change again has no effect on which slots run.
**Baseline:** before=0 after=0 tests added
<!-- predicted: no test file is added by this task -->
**Commit:** `feat(flow): dispatch the resolved reviewer roster`
**Build:** green

- [x] 3. Correct the command's own description

`skills/flow-settings/SKILL.md` tells the operator that three slots are dispatched by every run and
the remaining two are on-demand-only. After tasks 1 and 2 that is false: every slot in the list is
dispatched, and on-demand means "added for one run without changing the store."

Correct it, and say what writing the list now does — because until this change it did nothing, and
an operator who learned that will not expect otherwise.

Leave the command's guardrails, its `ValidReviewers`-is-canonical rule, and its no-op-write rule
untouched.

**Files:** `skills/flow-settings/SKILL.md`
**Tests:** none added — as task 1
**Regression:** reverting this commit leaves the command describing a roster policy the pipeline no
longer implements.
**Baseline:** before=0 after=0 tests added
<!-- predicted: no test file is added by this task -->
**Commit:** `docs(flow-settings): describe what the reviewers list now does`
**Build:** green

- [x] 4. What the panel does with an id it cannot spawn

Making the list the roster makes a new state reachable: a resolved id whose agent type does not
exist in the harness actually running. This is not hypothetical — Claude Code in this session
offers no `bugbot` agent type, while the store's list carries `bugbot`.

<!-- measured: the session's own available agent types are claude, claude-code-guide, Explore, general-purpose, Plan, statusline-setup @ 2026-08-27 -->

State the rule in `skills/flow/review-panel.md`, beside the roster table:

- **A resolved id whose own agent type is unavailable is dispatched as a general-purpose subagent
  carrying that slot's brief**, rather than skipped. The panel is not silently reduced by the
  harness it happens to run in.
- **That substitute MUST perform mutation testing** — it does not merely read for defects, it
  changes the code to prove each finding is real and reverts afterwards. This is the operator's own
  requirement and is what keeps a substituted slot worth dispatching.
- **The substitution is recorded, never hidden.** The dispatch row names the slot it stood in for
  and that it ran as a general-purpose agent; `-model` records the model actually given, never
  `unknown (agent-defined)`, because a general-purpose dispatch does carry an override. The panel
  record and the handoff say which slots were substituted.

**Never let a substitution read as the real slot.** A dispatch recorded as Bugbot that was not
Bugbot corrupts the one record that says what reviewed this branch.

Do not restate resolution (task 1) or the roster mapping (task 2) — cite them.

**Files:** `skills/flow/review-panel.md`
**Tests:** none added — this task edits a contract file; the prose guards named in task 1 are what
verify it
**Regression:** reverting this commit leaves a resolved id that cannot be spawned undefined —
a run would either skip the slot silently or improvise, and neither is written down.
**Baseline:** before=0 after=0 tests added
<!-- predicted: no test file is added by this task -->
**Commit:** `feat(flow): define what the panel does with an unspawnable roster id`
**Build:** green
