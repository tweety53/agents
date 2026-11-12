# kan-1106-cheap-prompt-tuning
> Design record of change kan-1106-cheap-prompt-tuning (KAN-1106), written at landing. Each decision's own **Status:** line is the status.

## Context

KAN-1106. `/flow-prompt-tuning` (`skills/flow-prompt-tuning/SKILL.md`) walks the cases
self-review's angle 8 saved under `docs/prompt-tuning/<case-id>/`, groups them, re-anchors each
update, replays it, applies what passes, ledgers every case and lands. On the base:

- Step 4 prescribes "3 runs per arm" on two arms with a grader dispatch per run, and
  `## Every dispatch` puts every dispatch on opus: 6 runs and 6 graders a case.
- `docs/prompt-tuning/LEDGER.md` holds its header only. 16 case folders are pending: 13
  implementer cases and 3 reviewer cases, two of the three being verifier cases that carry
  `capture.png` and `mockup.png`.
- A case records no reproducer command, pattern or test. Its mechanical material is `bad.diff`,
  `good.diff`, `finding.md`, `task.md`, `source.repo` and `source.fix_sha`; `finding.md` carries
  `**Severity:**`, `**Slot:**`, `**Location:**` and the finding's text.
- `flow-guard check-prompt-tuning-cases` pins the ledger line shape and accepts any free text
  after `rejected:`. No guard or test pins the replay's prose.
- `hooks/enforce-agent-baseline.py` does not treat `/flow-prompt-tuning` as a flow session: only
  the two baseline sentences are enforced on its dispatches. Opus at low effort is
  `subagent_type: "flow-low"`, `model: "opus"`. The skill records no dispatch row, and
  `stats/internal/guard/stagemarkcalls.go` lists it as an expected zero.
- The stats store's `finding_patterns` is empty: the per-class counts KAN-1106's analysis gives
  are a one-off keyword classification of 564 Critical or Important findings since 2026-10-01,
  not a figure a run can query.
- `skills/flow/review-panel-fix-round.md` carries **Mutation is the one role never re-dispatched
  after a fix**, the rule the rejection of `kan-1032-prompt-tuning-angle-1` rests on.

The pile, by the analysis's grouping:

| Group | Cases | Disposition |
|---|---|---|
| Test gap | `kan-1046-review-dynamics-dashboard-1` (implementer), `kan-1017-templates-step-3-admin-templates-queue-create-2` (test-sanity reviewer) | stay pending |
| Test gap, superseded | `kan-1032-prompt-tuning-angle-1` | rejected in this change |
| Visual-verify loop | `kan-1024-templates-warm-up-cool-down-any-workout-1` (verifier false positive), `kan-1061-flow-add-flow-interactive-fast-conversational-1`, `kan-1047-review-dynamics-heatmap-by-project-1`, `kan-1059-repeat-row-keep-awake-calendar-swipe-1` (implementer), `kan-1067-flow-allowed-pairs-experiments-stored-in-flowd-1` (verifier miss) | stay pending |
| Rare class | `kan-1013-templates-step-3-be-review-states-admin-1`, `kan-1017-templates-step-3-admin-templates-queue-create-1`, `kan-1058-port-remaining-sh-to-go-1`, `kan-1059-repeat-row-keep-awake-calendar-swipe-2`, `kan-1061-flow-add-flow-interactive-fast-conversational-2`, `kan-1062-flow-implement-one-shot-subagents-1`, `kan-1075-impl-alignment-judge-1`, `kan-1102-exercise-ids-lookup-and-server-search-1` | rejected in this change |

The change is one change because the cheaper replay and the triage answer the same question — what
the next `/flow-prompt-tuning` run costs — and the triage's ledger lines are written in the shape
the skill's step 6 states. Nothing a user sees changes, so no prototype was built.

## Recommended approach

Rewrite the replay of `skills/flow-prompt-tuning/SKILL.md` in place and triage the pile by hand in
the same change; no Go, no guard and no case field changes.

**`## Every dispatch`.** Every dispatch — replay and grader — is `subagent_type: "flow-low"`,
`model: "opus"`. The two baseline sentences, one-shot and never-writes-in-a-repository stay as
they are.

**Step 4, retitled `### 4. Replay`.** One run per case, carrying the target's text with the update
applied; no run carries the target's current text (decision `replay-arm-b-only`). The three
dispatch forms stay — reviewer over `bad.diff` (a miss) or `good.diff` plus the finding's location
(a false positive), verifier over `capture.png` and `mockup.png`, implementer over `task.md` and
the pre-change code in a scratch directory. A replay dispatch returns what its judge reads: an
implementer its diff, a reviewer or verifier its findings with locations.

- **Pre-change code** is stated in one new sentence (decision `pre-change-code-sentence`).
- **Verdict per run**, against `finding.md`:
  - `avoided` — the output reaches the finding's location and the defect is absent (a miss: the
    finding is raised; a false positive: it is not raised, or is attributed elsewhere);
  - `present` — the defect is there: the case is rejected `no effect`, with no second run;
  - `ambiguous` — the output never reaches the finding's location, so neither answer can be
    quoted: one second run, and on it `present` rejects the case `no effect` and `ambiguous`
    rejects it `not replayable` (decision `ambiguous-twice-not-replayable`).
- The `not reproducible` rejection and the 2-of-3 pass rule go. A group passes when every case in
  it is `avoided`.
- The skill states once what such a run proves: the updated text produced the right behaviour on
  this input, not that the old text would have failed there.

**Who judges.**

- **Implementer case — one grader dispatch per replay run, blind** (decisions
  `blind-grader-on-implementer-cases`, `grader-sees-no-update`, `grader-verdict-is-final`).
  Inputs, and no others: the replay's returned diff, and the case's `finding.md`, `bad.diff` and
  `good.diff`. Judging rule: `avoided` when the diff touches the code `finding.md`'s
  `**Location:**` names and does not carry the defect `finding.md` describes, `bad.diff` being the
  defect's shape and `good.diff` one correct shape, not the only one; `present` when it carries
  the defect; `ambiguous` when it never touches that code. It returns exactly one of the three
  verdicts and one quoted line of the diff as evidence — for `ambiguous`, the statement that no
  hunk touches the location. An `ambiguous` first run therefore costs a second replay and a second
  grader.
- **Reviewer and verifier case — the main session, by the literal rule** (decision
  `raised-means-same-file`). No dispatch. The session reads the replay's returned findings and
  records the verdict per case in the report, quoting both lines, the output's and `finding.md`'s.
  An output the session cannot quote a deciding line from is `ambiguous`.

**Guardrail and apply text.** "Never apply an update whose group did not pass the replay" stands
unchanged. Step 5's rule-file header wording "each proven by an A/B replay" becomes "each proven
by a replay".

**Prose that names the replay "A/B".** `CLAUDE.md`, `AGENTS.md`, `templates/CLAUDE.md`,
`templates/AGENTS.md`, `skills/README.md`, `README.md`, `commands-claude/flow-prompt-tuning.md`
and the `SKILL.md` frontmatter and intro each drop "A/B"; `commands-claude/flow-prompt-tuning.md`
also drops "on opus" for the replay's new pair.

**Triage** (decisions `triage-in-this-change`, `rare-class-reasons`, `kan-1032-superseded`). This
change appends nine `rejected:` lines to `docs/prompt-tuning/LEDGER.md` and `git rm -r`s those
nine folders, in one `docs(prompt-tuning): ledger <date>` commit. No `flow prompt-tuning record`
call: that records applied cases only. The ledger's header sentence drops "`/flow-prompt-tuning`
replayed", so it reads "Each entry records one prompt-tuning case — applied with the commit that
applied it, or rejected with the reason — once its case folder is removed
(`skills/flow-prompt-tuning/SKILL.md`)."

**Budget of the next run.** Seven pending cases, no two sharing a target and class: four
implementer, one reviewer, two verifier.

| | Replay runs | Grader dispatches | Total |
|---|---|---|---|
| Every first run decides | 7 | 4 | 11 |
| Every first run `ambiguous` | 14 | 8 | 22 |
| Today's skill on 16 cases (6 runs + 6 graders each) | 96 | 96 | 192 |

Per case: an implementer case costs 2 dispatches, 4 at worst; a reviewer or verifier case 1, 2 at
worst.

Rejected alternatives, one line each:

- a control run with the target's current text — doubles the replay cost (`replay-arm-b-only`);
- the main session judging every case — it merged the update it would judge
  (`blind-grader-on-implementer-cases`);
- a mechanical per-case check — needs a new case field and a pattern no behavioural defect has
  (`blind-grader-on-implementer-cases`);
- a grader that sees the update's text — it would know the hoped-for answer
  (`grader-sees-no-update`);
- the main session overruling or completing the grader's verdict (`grader-verdict-is-final`);
- "raised" only at the same file and line (`raised-means-same-file`);
- a payback step in the skill, or an operator-named reject list (`triage-in-this-change`).

## Critical files

- `skills/flow-prompt-tuning/SKILL.md` — the skill; its frontmatter, intro, `## Every dispatch`,
  step 4 and step 5's rule-file header wording are edited.
- `docs/prompt-tuning/LEDGER.md` — header sentence edited, nine `rejected:` lines appended.
- `docs/prompt-tuning/<case-id>/` — the nine folders of decisions `rare-class-reasons` and
  `kan-1032-superseded`, removed; each one's `case.json` supplies its ledger line's type, target,
  defect class and update summary.
- `CLAUDE.md`, `AGENTS.md`, `templates/CLAUDE.md`, `templates/AGENTS.md`, `skills/README.md`,
  `README.md` — the skill-index row for `/flow-prompt-tuning`, each dropping "A/B".
- `commands-claude/flow-prompt-tuning.md` — the command file; drops "A/B" twice and "on opus".
- `stats/internal/guard/prompttuningcases.go` — `check-prompt-tuning-cases`; read, not edited: it
  pins the ledger line shape the nine lines must take.
- `skills/flow/review-panel-fix-round.md` — read, not edited: the rule `kan-1032-superseded`
  cites.

## Existing code to reuse

- `flow-guard check-prompt-tuning-cases` (`stats/internal/guard/prompttuningcases.go`) — validates
  the ledger lines and the seven remaining folders.
- The ledger line shape of step 6 of `skills/flow-prompt-tuning/SKILL.md`, unchanged.
- The `flow-low` agent definition with `model: "opus"` — the replay's and the grader's pair, with
  no new `Role:` row.
- Step 1's red-case handling ("reported by its folder and left untouched") — the outcome a case
  whose repository is not on the machine takes.

## Verification

- `scripts/flow-guard run-verify-list <worktree> lint` exits clean: the change edits Markdown under
  `skills/`, so the whole lint list runs.
- `scripts/flow-guard check-prompt-tuning-cases <worktree>/docs/prompt-tuning` prints no line for
  `LEDGER.md` and none for any of the seven remaining folders.
- `ls docs/prompt-tuning` shows `LEDGER.md` and exactly seven folders — the test-gap and
  visual-verify rows of `## Context`'s table that stay pending — and `LEDGER.md` holds exactly
  nine entries, each ending in a `rejected:` reason of decisions `rare-class-reasons` or
  `kan-1032-superseded`.
- `git grep -n -F -e 'A/B' -- CLAUDE.md AGENTS.md templates/CLAUDE.md templates/AGENTS.md
  skills/README.md README.md commands-claude/flow-prompt-tuning.md
  skills/flow-prompt-tuning/SKILL.md` prints nothing, and so does the same sweep for
  `-e 'not reproducible' -e '3 runs per arm' -e '2 of 3'` over
  `skills/flow-prompt-tuning/SKILL.md`.
- A read of `skills/flow-prompt-tuning/SKILL.md` against `## Recommended approach`: every rule
  there is present once, and the grader's input list names nothing beyond the four inputs.

## Live check

None — the change edits Markdown a later `/flow-prompt-tuning` session reads and removes case
folders; it touches no service, daemon, store or scheduler, and the replay that would exercise the
new text is deliberately outside this change (decision `operator-runs-the-replay`).

## Non-goals

- No new case field, no reproducer in `case.json`, no change to `check-prompt-tuning-cases`.
- No `Role:` row for the replay or the grader in `skills/flow-contracts/model-policy.md`.
- No rationale appendix for the skill.
- No payback step, no operator-named reject list, no filter in angle 8 of
  `skills/flow-self-review/SKILL.md` (decision `mining-unchanged`).
- No replay of the seven pending cases (decision `operator-runs-the-replay`).

## Risks

- A run with the update applied passes an update on an input where the old text would not have
  shown the defect either. Limited by the sentence the skill carries on what such a run proves;
  accepted with decision `replay-arm-b-only`.
- `kan-1061-flow-add-flow-interactive-fast-conversational-1`'s task seeds data and takes captures
  against a running UI; a low-effort run in a scratch directory may never reach the leftover
  hand-back the update governs. The case then rejects itself `not replayable`; accepted with
  decision `applied-only-through-replay`.
- `kan-1024-templates-warm-up-cool-down-any-workout-1`'s update asks the verifier to measure a
  component at the merge base, which a replay holding two PNGs cannot do; its best outcome is
  "attribution needs the merge-base measure". Accepted with the same decision.
- `kan-1017-templates-step-3-admin-templates-queue-create-2`'s replay hands the reviewer a 41-line
  `bad.diff`, far narrower than the diff the original reviewer read. Accepted with the same
  decision.
- A rejected rare class recurs. Limited by the mining side being unchanged: self-review saves the
  case again, and it costs the next run one dispatch.

## Decisions

### What a replay runs to prove an update

**ID:** replay-arm-b-only
**Status:** active
**Chosen:** one run per case, on every case, carrying the target's text with the update applied, on opus at low effort; a second run only after an `ambiguous` first one — one dispatch per case, which is what the ticket plans.
**Considered:** a run with the target's current text beside it on implementer cases only — leaves the same blind spot on reviewer cases, `kan-1017-templates-step-3-admin-templates-queue-create-2`'s narrowed diff most of all, and doubles the cost of a pile heavy in implementer cases. One run per arm on every case — 2 to 4 dispatches a case, past the ticket's budget once more than about five cases are pending, and one run per arm still mislabels a defect that appears only some of the time. Today's three runs per arm — about 12 dispatches a case.

### Who judges a replay's output against `finding.md`

**ID:** blind-grader-on-implementer-cases
**Status:** active
**Chosen:** one opus-low grader dispatch per implementer replay run; the main session judges reviewer and verifier cases by the literal rule of `raised-means-same-file` — the session that merged the update does not read its own experiment where the reading is a judgment of code.
**Considered:** the main session judging every case with one quoted line — no grader dispatch, but the session that merged the `after` text reads its own experiment generously. A mechanical check, each case recording a pattern or command — needs a new `case.json` field, the `Case` struct and renderer, the guard and its test, step 2a of `skills/flow-self-review/SKILL.md` and a hand-written pattern for every pending case, and fails on a defect no pattern expresses and on a correct output written differently from `good.diff`.

### What the grader is given

**ID:** grader-sees-no-update
**Status:** active
**Chosen:** the replay's diff, `finding.md`, `bad.diff` and `good.diff`, and nothing else — its prompt carries nothing that says an update was applied, what the update says, or which verdict is hoped for, so it judges the code, not whether the code obeys a new sentence.
**Considered:** a grader that also gets the update's `after` text — a judge that knows which answer the experiment hopes for.

### What the grader returns and who may change it

**ID:** grader-verdict-is-final
**Status:** active
**Chosen:** exactly one of `avoided`, `present` or `ambiguous` with one quoted line of the diff as evidence; the main session records the verdict and the line as returned and never overrules it.
**Considered:** the main session may overrule the grader — puts the session that merged the update back into the judgment the grader was added to take from it. The grader returns two verdicts and the main session decides `ambiguous` itself (18 dispatches at worst) — the same objection.

### When a reviewer or verifier replay has raised the finding

**ID:** raised-means-same-file
**Status:** active
**Chosen:** the finding is raised when the output names the defect `finding.md` describes at the file its `**Location:**` names — for a verifier case, at the element it names — whatever the line; the report quotes both lines, the output's and `finding.md`'s.
**Considered:** raised only at the same file and line — a correct review anchored two lines away costs a second run and may reject a working update `not replayable`.

### A case still `ambiguous` after its second run

**ID:** ambiguous-twice-not-replayable
**Status:** active
**Chosen:** rejected `not replayable`, ledgered, its folder removed — the pile stays finite, and the class is mined again if it recurs.
**Considered:** rejected `no effect` — one reason for every case that did not pass, which hides that the replay never reached the finding. Left pending and reported — the folder stays, and every later run spends up to two dispatches on it again.

### How an implementer replay finds its pre-change code

**ID:** pre-change-code-sentence
**Status:** active
**Chosen:** one sentence in step 4 — the pre-change code is the tree at `source.fix_sha`'s parent in the checkout of `source.repo`; a case whose repository the session cannot find on this machine is reported by its folder and left untouched, as a red case of step 1 is. `source.repo` is a bare repository name on 14 of 16 cases, and the skill said only "the pre-change code".
**Considered:** leaving the text as it is — each run resolves the gap as it can. Handing the replay `task.md` and only the files `bad.diff`'s headers name at that parent commit — an implementer without the rest of the repository cannot write the change the task asks for.

### How a case that cannot pay back leaves the pile

**ID:** triage-in-this-change
**Status:** active
**Chosen:** a one-off triage in this change — nine `rejected:` ledger lines and nine folders removed; `SKILL.md` gains no payback step.
**Considered:** a payback step in `SKILL.md`, the main session rejecting a class it judges too rare — no stored recurrence figure exists (`finding_patterns` is empty), so the rule is a judgment call no run can check and two runs can decide one case differently. An operator-named list as the command's argument — an operator cannot tell from a case id what the case is without a fresh analysis beside it.

### The reason each rare-class ledger line carries

**ID:** rare-class-reasons
**Status:** active
**Chosen:** the count where KAN-1106's analysis has one — `rejected: rare class, <n> of 564 Critical or Important findings in six weeks (KAN-1106)`, `<n>` being 6 for `kan-1013-templates-step-3-be-review-states-admin-1`, 4 for `kan-1058-port-remaining-sh-to-go-1`, 6 for `kan-1059-repeat-row-keep-awake-calendar-swipe-2`, 5 for `kan-1061-flow-add-flow-interactive-fast-conversational-2`, 1 for `kan-1062-flow-implement-one-shot-subagents-1` and 11 for `kan-1102-exercise-ids-lookup-and-server-search-1`; `rejected: source finding is a Minor, class not countable (KAN-1106)` for `kan-1075-impl-alignment-judge-1`, whose count is keyword noise; `rejected: rare class, cannot pay back (KAN-1106)` for `kan-1017-templates-step-3-admin-templates-queue-create-1`, the one case the analysis gives no figure for — a reader of the ledger sees the measure the rejection rests on.
**Considered:** `rejected: rare class, cannot pay back (KAN-1106)` on all eight — drops the figure the rejection rests on where one exists.

### The reason `kan-1032-prompt-tuning-angle-1`'s ledger line carries

**ID:** kan-1032-superseded
**Status:** active
**Chosen:** `rejected: superseded — Mutation is never re-dispatched after a fix, so the round this targeted no longer exists; the update adds per-clause flips to every task`.
**Considered:** `rejected: already covered` with the line of `skills/flow/review-panel-fix-round.md` quoted — step 3's existing reason, although the gap the case names, an unpinned clause, is not itself closed by that line.

### The three visual-verify cases the ticket assigns to neither side

**ID:** three-visual-cases-stay-pending
**Status:** active
**Chosen:** `kan-1047-review-dynamics-heatmap-by-project-1`, `kan-1059-repeat-row-keep-awake-calendar-swipe-1` and `kan-1067-flow-allowed-pairs-experiments-stored-in-flowd-1` stay pending and the next run replays them — three more dispatches, and a case the replay cannot reach rejects itself.
**Considered:** rejecting all three here — decides without a replay what a replay can decide for one dispatch each. Rejecting the two implementer cases and keeping the verifier miss — the two need a running UI to replay, but `ambiguous-twice-not-replayable` already removes them if so.

### How the cases the ticket names for applying are applied

**ID:** applied-only-through-replay
**Status:** active
**Chosen:** through the replay — each is applied only if its run comes back `avoided`, and a rejection is accepted if it comes; the guardrail "Never apply an update whose group did not pass the replay" stands.
**Considered:** a replay "at the decision point" that hands the implementer the measured departure and asks only for the hand-back — a new replay form in `SKILL.md` whose prompt partly gives the answer away. Applying the named cases without a replay on the analysis's argument — the guardrail gains an exception.

### Who replays the seven pending cases

**ID:** operator-runs-the-replay
**Status:** active
**Chosen:** the operator, with `/flow-prompt-tuning`, after this change lands — this change edits the skill, the prose naming it and the triage, and the replay runs under the landed skill.
**Considered:** a task of this plan running the replay, applying what passes and writing those ledger lines — the replay would run under a skill text not yet landed, and its applied updates would widen this change to every target file a case names.

### Whether self-review keeps saving rare-class cases

**ID:** mining-unchanged
**Status:** active
**Chosen:** angle 8 of `skills/flow-self-review/SKILL.md` is not touched — a rare-class case costs the next run one dispatch, and a filter is added when a second pile shows the need.
**Considered:** a payback condition in angle 8 — widens this change to the self-review skill and carries the weakness of the payback step `triage-in-this-change` rejects: no stored figure to judge from.

## Open questions
