# kan-472-flow-dynamic-review-panel-roster-repo-scoped — design

## Context

`/flow` makes three per-run choices statically: execution shape (always the conductor tree in
`skills/flow/implement.md`), implementer/fixer model (`DEFAULT_MODEL`, **Model resolution** in
`skills/flow/SKILL.md`) and review panel (the settings-store `reviewers` list on `DEFAULT_MODEL`,
`skills/flow/review-panel.md`). This change makes each of the three dynamic behind its own
repo-scoped toggle, decided once per run by the planner, recorded in the store, enforced at
dispatch, and reported in two new stats views. KAN-475 is joined: its inline-vs-SDD decision is the
first link of the chain KAN-472's two decisions follow. KAN-478 is joined too (duplicate, closed
with KAN-472): its bundling of reviewer roles and of implementer bundles into fewer subagents is the
**Bundled dispatch** and **Implementer merging** sections below, decided by the same planner step.

Existing seams this design reuses rather than adds to: `.flow/project.md` single-literal keys
(**Project configuration**, `skills/flow-contracts/project-configuration.md`); the `Model:`
handshake `skills/flow/brainstorm.md` and `implement.md` already run for the planner and conductor;
`dispatches.slot` / `findings.slot` (`stats/internal/store/migrations/0010_run_records.sql`);
`GET /api/v1/stats/{view}` and the SPA's `VIEW_NAMES`; `/flow`'s existing re-entry by checkbox
state, `worktrees` map and store findings.

No `spectre/specs/` capability exists for the pipeline; none is edited.

## The decision chain

Decided together, once, by the planner at the end of section **D** (`skills/flow/brainstorm-planner.md`),
in this order:

1. **execution mode** — `inline` or `sdd` (only when `## execution mode` is `dynamic`; else `sdd`)
2. **implementer/fixer model + effort** (only when `## implementer model` is `dynamic` **and** step 1
   came out `sdd`; recorded `skipped — inline` when step 1 is inline, `default` when the toggle is)
3. **review panel** — roster, per-slot model + effort, compact, experimental, rerun policy, and the
   **grouping** of the roster into at most two dispatches (only when `## review panel` is
   `dynamic`; else today's roster and rule, recorded as `default`, grouped by the static table under
   **Bundled dispatch**)
4. **implementer groups** — which of `plan-dispatch-bundles.sh`'s bundles merge into one implementer
   dispatch (on every run whose step 1 came out `sdd`, whatever the toggles; `null` when inline) —
   **Implementer merging** below

Every step's result is written into the `## Decision` block whether it was decided or defaulted, so
the operator always sees all four rows.

### Inputs and the class

Read from `<changeRoot>/tasks.md` and the resolved worktree set — nothing the planner judges:

| Input | Source |
|---|---|
| `tasks` | count of column-0 `- [ ] <n>.` lines |
| `files` | size of the union of every task's `**Files:**` paths |
| `repos` | size of the resolved worktree set |
| `migration` | any `**Files:**` path under `stats/internal/store/migrations/` or ending `.sql` |
| `spec` | any `**Files:**` path under `spectre/specs/` |
| `red` | any task tagged `Build: red` |
| `unverified` | any `unverified:` provenance tag in the plan |

| Class | Rule |
|---|---|
| **small** | `tasks ≤ 5` and `files ≤ 12` and `repos = 1` and not `migration` and not `spec` |
| **big** | `tasks ≥ 15` or `files ≥ 40` or (`repos > 1` and `tasks ≥ 8`) or (`migration` and `tasks ≥ 8`) |
| **regular** | everything else |

`red` and `unverified` are recorded and move no class. **The planner may raise the class one step**
(small→regular, regular→big), never lower it, with a one-line reason recorded as `override`; the
mechanical class is recorded beside it as `class_mechanical`.

### The rolls

`compact_roll = sha256("<name>") mod 100`, `experimental_roll = sha256("<name>exp") mod 100`,
`bundle_roll = sha256("<name>bundle") mod 100`, all over the change name's UTF-8 bytes, computed
with `printf '%s' … | shasum -a 256` and the first 8 hex digits read as an integer. Compact when
`compact_roll < 70` (small) or `< 30` (regular, big). Experimental when `experimental_roll < 30`
(every class). Grouping is **static** when `bundle_roll < 30` and **free** otherwise (**Bundled
dispatch › Grouping**). Reproducible per change; one experimental slot at most.

### The tree

**`code-review-low` is retired from this tree entirely, on every class and every roll —
`simple-reviewer` replaces it everywhere.** `code-review-low` remains a legal id only for the
`## review panel` `default` path (the harness-wide settings-store roster, an unrelated,
pre-existing surface this change does not touch) — never a role this dynamic tree assigns.

| Class | Execution | Implementer / fixer (sdd only) | Full roster (slot: model/effort) | Compact roster | Rerun policy |
|---|---|---|---|---|---|
| **small** | inline | — | primary: sonnet/medium; simple-reviewer: haiku/medium; principles: haiku/medium | primary: sonnet/medium; simple-reviewer: haiku/medium; principles: haiku/medium | delta |
| **regular** | inline | — | primary: sonnet/high; simple-reviewer: haiku/medium; principles: sonnet/medium; mutation: sonnet/medium | primary: sonnet/high; simple-reviewer: haiku/medium; principles: sonnet/medium | delta |
| **big** | sdd | opus/high | primary: opus/high; simple-reviewer: sonnet/high; principles: opus/medium; mutation: sonnet/high; bugbot; security | primary: opus/high; simple-reviewer: sonnet/high; principles: opus/medium | full |

- **`primary` + `simple-reviewer` + `principles`, bundled as one dispatch, is the universal floor —
  every roster, compact or full, every class.** Nothing in this tree ever dispatches without this
  three-role bundle present; it alone already uses the bundle's full ≤3-role capacity. Small's full
  and compact rosters are identical (the floor is the whole roster, all three roles at small's own
  economy model/effort — `principles` was never part of small's roster before this floor existed,
  and now takes `haiku/medium` there, the same tier `simple-reviewer` already runs small class at);
  regular's and big's full rosters layer their remaining roles (`mutation`, and for big `bugbot`/
  `security`) into the second dispatch (**Bundled dispatch › Grouping**, below, is canonical — the
  floor's own three roles leave no spare capacity, so the second dispatch is exactly whatever
  remains, up to its own ≤3-role cap).
- `bugbot` and `security` are prompt-driven roles like every other slot (**Bugbot and Security
  are prompts**, below): on big they take the class's `simple-reviewer` model/effort (sonnet/high),
  recorded like any other slot's.
- **At least one code-quality reviewer is always present, every class, every roll** — the floor
  bundle's own `simple-reviewer`. `primary` checks plan alignment only — proposal, design and each
  task's declared `**Files:**`/`**Tests:**`/`**Commit:**` fields — never code quality, which is
  `simple-reviewer`'s and Bugbot's job (**The roster**, `skills/flow/review-panel.md`;
  **`simple-reviewer` is a new persistent slot**, below).
- **delta** is today's rule under **Panel re-runs** (`skills/flow/review-panel.md`), unchanged.
  **full** keeps that rule for every fix round and adds one final pass after the last fix round
  closes clean: every slot in the roster re-reads the whole `final-review.diff` (bugbot/mutation in
  pass-1 shape). A finding from that pass opens an ordinary fix round; the final pass then repeats.
- The experimental slot, when rolled, is appended to whichever roster was chosen, on
  sonnet/medium (small, regular) or sonnet/high (big) — unless the roster already fills two
  dispatches of three (big, full): then it is skipped and recorded `experimental: skipped — bundle
  cap`. A persistent role is never displaced for a trial one.
- Effort values are the three the harness exposes: `low`, `medium`, `high`.
- The docs-only reduction (`skills/flow/review-panel.md`) still applies to a dynamic roster and
  still only removes.

### The `## Decision` block

The last section of the planner's `## Plan` return, and the exact text the parent prints:

```markdown verified:authored in-tree for this change; the shape every task prints and reads
## Decision

| Setting | Toggle | Result |
|---|---|---|
| execution mode | dynamic | inline |
| implementer model | dynamic | skipped — inline |
| review panel | dynamic | compact — primary sonnet/high, simple-reviewer haiku/medium, principles sonnet/medium; experimental: skipped — bundle cap; rerun delta; dispatches: primary+simple-reviewer+principles |
| implementer groups | — | skipped — inline |

class: regular (mechanical: small; override: two tasks touch the harvest attribution path)
inputs: tasks=4 files=9 repos=1 migration=no spec=no red=no unverified=yes
rolls: compact 41 (<70 → compact) · experimental 12 (<30 → exp-failure-modes) · bundle 63 (≥30 → free)
```

The review-panel cell ends with `dispatches: <group> · <group>`, each group its roles `+`-joined
in roster order, and on a free grouping the `grouping_reason` on the next line as `grouping: free
— <reason>`. The implementer-groups row shows `skipped — inline`, or the groups as
`[1,2] · [3] · [4,5]` (bundle ids from `plan-dispatch-bundles.sh`) followed by `— <reason>`.

The parent prints it verbatim, then records it — no prompt, no wait — and marks the step:

```bash unverified:flow.decide and flow record decision are added by tasks 3 and 8; the mark shape is every other stage's
flow stage begin -command '/flow' -stage flow.decide -harness <harness> -session-token mf-<literal-token> <name>
flow record decision -change <name> -session-token mf-<literal-token> -file <abs-worktree>/.superpowers/sdd/decision.json
flow stage end -command '/flow' -stage flow.decide -outcome completed <name>
```

`<abs-worktree>/.superpowers/sdd/decision.json` is written by the planner beside the block and is
the machine form of it — the same run-artifact directory every report file already lands in, never
a planning path. The classifier and rolls are one shipped script, `plan-class.sh <tasks.md>
<repos>`, printing the `inputs:`, `class:` and `rolls:` lines the block carries, so the planner's
arithmetic is reproducible by anyone from the plan file. A record write never blocks.

### Resume and fix runs

A resumed run (`STARTED` past writing-plans, or `IN_PROGRESS`) reads the newest `decisions` row
for the change and follows it — never re-rolls. A fix run's planner (`flow.document-fix`) re-runs
the classifier over the appended plan and records a second row; the rolls, being name-derived, are
identical, so only the class can change.

## Toggles

Three optional `.flow/project.md` keys, each holding exactly one of the literals `default` or
`dynamic`, matched byte-for-byte after trimming, report-by-name-and-drop on anything else, absent
meaning `default` — the same shape `## default landing route` already has:

| Key | `default` | `dynamic` |
|---|---|---|
| `## execution mode` | sdd via the conductor, as today | class decides |
| `## implementer model` | `DEFAULT_MODEL`, effort `default` | class decides model + effort |
| `## review panel` | settings-store roster on `DEFAULT_MODEL`, delta rerun | class + rolls decide |

Read through `project-get.sh <worktree> "<key>"`, in **Model resolution** (`skills/flow/SKILL.md`)
beside the existing `## planning model` read, into `EXECUTION_MODE_TOGGLE`,
`IMPLEMENTER_MODEL_TOGGLE`, `REVIEW_PANEL_TOGGLE`. A plain-language session instruction overrides a
*result* (today's "use opus for the panel" rule), never a toggle; the override is recorded in the
decision row's `overrides` field and never written back.

## Enforcement

- **Agent definitions.** `<agents repo>/agents/flow-<model>-<effort>.md` for `model ∈ {sonnet, opus,
  haiku}` × `effort ∈ {low, medium, high}` — nine files. Frontmatter: `name`, `description`,
  `model: <model>`, `effort: <effort>`, no `tools:` restriction. Body: one line, "General-purpose
  flow role; the dispatch prompt carries every instruction." `setup.sh global` symlinks them into
  `~/.claude/agents/`; `scripts/check-guard-symlinks.sh` gains the same presence rule it applies to
  `skills/flow/scripts/`. A dynamic dispatch passes both `subagent_type: flow-<model>-<effort>` and
  `model: <model>`.
- **Universal handshake.** Every dispatched role — implementer, panel slot, panel-fix, verifier,
  conductor, planner — opens its first reply with `Model: <model>`; every dispatch prompt says so.
  Mismatch: close the row `-outcome fallback`, re-dispatch once under `<key>-retry` on the same
  type; a second mismatch closes that row `-outcome fallback` and ends the turn with `## Question`
  naming the requested model and both answers, options **Continue on `<answered>`** / **Stop the
  run**. The planner's own `opus` fallback (`skills/flow/brainstorm.md`) is unchanged; the
  conductor's continue-on-mismatch rule (`skills/flow/implement.md`) is replaced by this one.
- **Effort is recorded, not handshaken** — a model cannot report its own effort. `-effort` on
  `flow record dispatch begin` carries what the dispatcher set: `low` / `medium` / `high`, or
  `default` where none was set. `dispatches.effort TEXT NOT NULL DEFAULT 'default'`.
- **Inline.** The parent records its own model and effort once per run in the decision row's
  `parent` field, from the harness's own report (the model named in its system prompt; effort as
  the harness states it, else `unknown`).

## Store

### `decisions`

```sql unverified:confirm against 0010_run_records.sql's conventions when task 1 writes 0019_decisions.sql
CREATE TABLE decisions (
  id            BIGSERIAL PRIMARY KEY,
  change_id     BIGINT NOT NULL REFERENCES changes(id),
  session_token TEXT NOT NULL,
  recorded_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  decision      JSONB NOT NULL,
  CONSTRAINT decisions_session_key UNIQUE (change_id, session_token)
);
CREATE INDEX decisions_change_id ON decisions (change_id);
CREATE INDEX decisions_decision_gin ON decisions USING GIN (decision);
```

`(change_id, session_token)` unique makes `flow record decision` idempotent per run, the same
replay property `dispatches.key` has. The JSON:

```json verified:authored in-tree for this change; the JSON tasks 3, 7 and 17 read and write
{
  "toggles": {"executionMode": "dynamic", "implementerModel": "dynamic", "reviewPanel": "dynamic"},
  "class": "regular", "classMechanical": "small", "override": "<reason or null>",
  "inputs": {"tasks": 4, "files": 9, "repos": 1, "migration": false, "spec": false, "red": false, "unverified": true},
  "rolls": {"compact": 41, "experimental": 12, "bundle": 63},
  "execution": "inline",
  "implementer": {"model": "opus", "effort": "high"} ,
  "groups": null,
  "panel": {
    "compact": true, "rerun": "delta",
    "roster": [
      {"slot": "primary", "model": "sonnet", "effort": "medium", "experimental": false},
      {"slot": "exp-failure-modes", "model": "sonnet", "effort": "medium", "experimental": true,
       "prompt": "skills/flow/experimental/failure-modes.md",
       "description": "What the diff does under error, timeout and partial write"}
    ],
    "grouping": "free",
    "dispatches": [["primary"], ["exp-failure-modes"]],
    "grouping_reason": "two reading roles on a four-task diff — no context to save by bundling"
  },
  "parent": {"model": "sonnet", "effort": "unknown"},
  "overrides": []
}
```

`implementer` is the string `"skipped — inline"` or `"default"` where not decided; `panel` is the
string `"default"` where the toggle is off. `groups` is `null` when execution is inline, otherwise
an array of arrays of bundle ids plus a sibling `"groups_reason": "<one line>"` — present on every
sdd run, `## execution mode` toggle or not. `panel.grouping` is `"static"` or `"free"`;
`panel.dispatches` is one to two arrays of one to three slot ids each, every id in `roster` in
exactly one of them, in roster order within an array; `panel.grouping_reason` is `null` on a static
grouping. An experimental slot skipped for the cap appears in no array and is recorded as the
string `"experimental": "skipped — bundle cap"` beside `roster`. Wire: `POST /api/v1/records/{project}/{change}/decisions`,
`GET …/decisions` (newest first), a `records.Decision` type, `flow record decision` and
`flow record decisions -change <name>` (JSON out, what a resumed run reads).

### `deferred <reason>`

`validateFindingStatus` (`stats/cmd/flow/record.go`) accepts `deferred <non-empty reason>` beside
`open`, `fixed`, `withdrawn <reason>`; the store's `SetFindingStatus` refuses it with a named error
when the finding's severity is not `Minor`. `check-unfinished-work.sh` and
`check-panel-findings-closed.sh` count `deferred` as closed. The renderer emits
`finding-status: F<n> deferred <reason>`.

## Dynamic panel dispatch

`skills/flow/review-panel.md` resolves its roster from the decision row when `REVIEW_PANEL_TOGGLE`
is `dynamic`: **The roster** table's `Model` column becomes "the decision's model/effort for this
slot", `-model`/`-effort` on each dispatch row carry them, and the experimental slot is dispatched
as a general-purpose reviewer (by `flow-<model>-<effort>` type) whose prompt is the file the decision
names, read by absolute path, with the same REPORT FILE / REPRODUCER / CONTEXT BUNDLE / WORKTREES
paragraphs every slot carries. Its `-slot` and every `flow record finding -slot` are the `exp-` id.
The rendered panel record's Slot column shows the id, so the prefix survives into the archive.

`skills/flow/experimental/<name>.md`: line 1 `description: <one sentence>`, then the prompt. The
planner validates a rolled prompt exists; an absent directory or an empty one records
`experimental: none available` and dispatches nothing extra. Shipped first: `failure-modes.md` —
reviews each changed behaviour under error return, timeout, partial write and concurrent re-entry,
an angle no persistent slot owns. Prompts stay in the agents repo; nothing copies them into a target
project.

### Bugbot and Security are prompts

`bugbot` and `security` are prompt-driven general-purpose roles on **both** toggle values, exactly
like `primary`, `principles`, `code-review-low` and `mutation`: `skills/flow/bugbot-reviewer-prompt.md`
(new — the defect hunt plus the mutation-testing brief, run in the slot's own throwaway worktree
copy) and `skills/flow/security-reviewer-prompt.md` (rewritten from its Cursor-era "substitute"
framing into the same shape: the angle, what to read, the finding format with severity and
reproducer, the standards-as-data clause). They are dispatched by `flow-<model>-<effort>` on
`dynamic` and as general-purpose on `DEFAULT_MODEL` on `default`, carry the MODEL HANDSHAKE, and
record `-model`/`-effort` like any other slot. The `subagent_type: bugbot` / `security-review`
dispatch, the `unknown (agent-defined)` model value, the **An unspawnable id is substituted, not
skipped** section and every recording exception narrowed by it are removed from
`skills/flow/review-panel.md`. Slot ids, `ValidReviewers`, the throwaway-worktree treatment for
`bugbot` and `mutation`, and `security`'s read-only sharing of `<worktree>` are unchanged.

## Bundled dispatch

**At most two review dispatches per panel round, each carrying one to three roles**, on both
`## review panel` values and in both execution modes. A dispatch carrying one role covers that role
alone; a roster the two dispatches cannot hold shrinks to what they hold — the operator's rule:
separate means two reviewers at most, bundled means more.

### One row per bundle

A bundle is **one** `dispatches` row: the harvester attributes a subagent transcript's tokens to a
row by agent id, and two rows sharing one id make every record ambiguous and drop the batch
(`stats/internal/harvest/watcher.go`, `bestDispatchWindow`'s identity pass). The row's `-slot` is
the bundle's roles `+`-joined in roster order (`primary+principles+security`), its key
`panel-<round>-<that slot>`, its `-model`/`-effort` the **highest** model and effort among its
roles (a bundle is one agent of one type; the per-role intended values stay in the decision's
`roster`). Every finding still records its own single role in `-slot`, with the bundle's
`-dispatch-seq`. A one-role dispatch is unchanged from today.

### The bundle prompt

The shared paragraphs once — CONTEXT BUNDLE, WORKTREES, TOOLS, FOREGROUND BUILDS, REPRODUCE DON'T
READ, CITATION CHECK, MODEL HANDSHAKE, the reproducer rule — then one **PASS `<id>`** section per
role in roster order, each carrying exactly the brief that role's solo dispatch carries today
(Primary's plan-alignment brief; `principles-reviewer-prompt.md` by absolute path with
`[PRINCIPLES_PATH]`/`[STANDARDS_PATHS]` resolved; `bugbot-reviewer-prompt.md` with its copies; an
`exp-` prompt by path; …) and its own REPORT FILE line naming `panel-report-<round>-<id>.md` — one
report per role, unchanged. The return message carries one findings summary per role under a
heading naming the role; the parent records each finding under that role. Mutating roles
(`mutation`, `bugbot`) are always the last passes of a bundle and still work in their throwaway
copies; the reading passes before them read the shared `<worktree>`.

One new verbatim paragraph, registered in `scripts/check-dispatch-paragraphs.sh` with
`review-panel.md` as its site:

> **INDEPENDENT PASSES:** each pass starts from `final-review.diff` and the code, never from an
> earlier pass's report or conclusions. Do not cite, defer to, or skip a defect because an earlier
> pass raised it — if it sits in this pass's angle, raise it again under this pass. Write each
> pass's report file before beginning the next pass.

The parent does **not** de-duplicate across roles: the same defect raised by two passes is two
`F<n>` rows, and the `reviewers` view counting it twice is the anchoring measurement this design
exists to take.

### Grouping

Which roles share a dispatch is decided per run by `bundle_roll` (**The rolls**):

- **`bundle_roll < 30` — static.** The class's row below, unchanged, no override: this 30% is the
  baseline every free grouping is compared against, and an overridden baseline is no baseline.
- **`bundle_roll ≥ 30` — free.** The planner groups the roster itself within ≤2 × ≤3, any shape
  including two unbundled roles, and records a one-line `grouping_reason` — the reasoning is data
  for the `decisions` view, not a justification of a deviation.

| Class | Static grouping (full roster) |
|---|---|
| **small** | `primary+simple-reviewer+principles` |
| **regular** | `primary+simple-reviewer+principles` · `mutation` |
| **big** | `primary+simple-reviewer+principles` · `mutation+bugbot+security` |

**The floor bundle (`primary+simple-reviewer+principles`) is always the first dispatch, on every
roster — compact or full, every class** — and it already carries three roles, the bundle cap's
full capacity, so it never has spare room for anything else. Small's roster (compact or full —
they are the same three roles) is the floor alone, one dispatch, nothing more to group. Regular's
and big's full rosters layer their remaining roles entirely into the second dispatch: regular's
one remaining role (`mutation`) fills it alone; big's three remaining roles (`mutation`, `bugbot`,
`security` — all defect-hunting/mutating, the same "mutating roles together" half of the
pre-bundling design's split) fill it exactly. Since the floor already holds every reading/judgment
role there is (`primary`, `simple-reviewer`, `principles`), the second dispatch is always purely
defect-hunting/mutating — there is no overflow case left to resolve (the floor never needs to
absorb a role from the second dispatch, or vice versa; **`floor-bundle-absorbs-reading-role-on-overflow`
is retired, its own decision record below**).

A compact roster's three roles are the floor bundle itself, deterministic, no roll: with only one
sensible grouping for three roles that must all bundle together, there is nothing to sample or
vary, and bundling still halves the dispatch count for free, the same benefit the full-roster case
takes. **A rolled experimental slot never has room in the floor bundle any more** (three of three
role slots already used, on every class) — it joins the second dispatch when one exists and has
room (regular: alongside `mutation`, two of three slots then used; big: only when big's second
dispatch is not already the full `mutation+bugbot+security`, which it always is, so big never has
room) or is skipped entirely (`experimental: skipped — bundle cap`, **The tree**) when there is no
second dispatch (small) or it has no room (big, always; regular, once already carrying two roles).

**On `## review panel` `default`** the settings-store list is grouped by the static table's logic,
deterministically, no roll and no planner: reading roles (`primary`, `principles`, `security`,
`code-review-low`) fill the first dispatch in that order up to three, the rest and the mutating
roles (`bugbot`, `mutation`) the second, up to three; a list the two cannot hold is truncated in
store order and the truncation recorded in `final-review-panel.md`. Everything else on `default` —
`DEFAULT_MODEL`, effort `default`, delta rerun, the docs-only reduction — is as today.

The panel record (`final-review-panel.md`) and the `IN_PROGRESS` handoff's `Panel:` line name the
dispatches as `+`-joined groups. The docs-only reduction still narrows to `primary` alone, one
dispatch. **Panel re-runs** are per role: a role re-runs when its own rule says so, and the roles
re-running in a round are re-grouped by the same grouping (a group whose other members are clean
dispatches with its re-running members only).

## Implementer merging

`plan-dispatch-bundles.sh`'s bundle stays the unit of file-overlap correctness. The planner, holding
`tasks.md`, runs it in **Decide** and records `groups`: which bundles one implementer works, as
arrays of bundle ids, on every run whose `execution` is `sdd` — `## execution mode` toggle or not.
A group's after-set is the union of its members'; a group is ready when every id in that union has
landed; one implementer works its bundles in plan order, one commit per task, one
`gather-dispatch-context.sh` call over the union of its task ids into
`dispatch-context-group-<g>.md`, one throwaway worktree when it shares a wave. Everything section
**4** of `skills/flow/implement.md` says about a bundle — the pick order, handback, FULL SUITE on
the plan-last unit, the guard overlap — holds for a group, with the group as the unit.

**At most two implementer dispatches in flight per wave**, on both `## execution mode` values;
ready groups beyond two queue in plan order and launch as earlier ones are picked. Grouping is the
planner's free judgment every run — no roll, no static table, no per-group ceiling — with a one-line
`groups_reason`; a group of one bundle is the common case and needs no reason beyond the line.
Inline execution has no implementer dispatch and records `groups: null`.

## Inline execution

Applies when the decision's `execution` is `inline`. `skills/flow/implement.md` gains an **Inline**
path beside **Dispatch the conductor**: the parent itself runs sections 1, 2 and 4, then
`review-panel.md` and `verify-and-handoff.md`, with these substitutions:

- No conductor, no implementer, no panel-fix dispatch. The parent does each bundle's TDD work in
  the canonical worktree, commits per task with the same `Task-Id:` trailer and `**Commit:**`
  subject, runs `check-task-commit-fields.sh` and ticks exactly as section 4 states. Waves are not
  parallel inline: bundles run in plan order.
- Every dispatch-prompt paragraph that instructs an implementer or fixer (COMMIT-PER-TASK, TDD,
  TARGETED TESTS, MUTATION PROOF, PLAN FIELDS, …) binds the parent in the same words.
- Panel slots and the verifier dispatch exactly as in sdd mode. Panel fixes are applied by the
  parent; the parent still runs every reproducer and the fix-diff walk before recording `fixed`.
- Records: one `dispatches` row per bundle with `-role implementer -model <parent model>
  -effort <parent effort> -agent-id inline`, and per fix round `-role panel-fix … -agent-id
  inline`, so cost attribution and the stats views see inline work under the same roles.

### The context ceiling

Checked before every bundle and before panel pass 1, from the harness's remaining-budget figure
(Claude Code's `<total_tokens>` reminder). Stop when remaining < **250,000** before a bundle or
< **400,000** before the panel. Where the harness exposes no figure, stop after the **6th** bundle
of one session. The stop closes the open stage `-outcome stopped`, writes no state, and prints:

```markdown verified:authored in-tree for this change; the handoff block task 14 adds
## Context ceiling — clear and resume

Inline run stopped before `<next step>` with `<remaining>` tokens left. Paste:

/clear
/flow <name>
```

The next run resumes by the existing re-entry rules and the recorded decision.

## Severity-gated minors

In `skills/flow/review-panel.md`'s fix round, both modes: every Critical and Important goes to the
fix (subagent or parent). For each Minor the dispatcher decides before the fix goes out — **fix**
when the change is confined to the lines the finding names and needs no new test, or when judgment
says the defect is worth a fix; otherwise `flow record status -ref F<n> -status 'deferred
<reason>'`. The contract states the rough 10% target as guidance and forbids computing it: no
counter, no draw, 0% is acceptable. The `IN_PROGRESS` handoff gains a `### Deferred minors` list
(`F<n> <location> — <note> — <reason>`), empty stated as `none`. The `skills/flow/SKILL.md`
guardrail "never hand off with an open finding" is unchanged — a deferred finding is not open.

## Stats views

Both under `GET /api/v1/stats/{view}`, period- and project-filtered like the existing views,
registered in `internal/api/stats.go`'s view switch and the SPA's `VIEW_NAMES`, rendered through
`ViewFrame` + `DataTable`.

- **`reviewers`** — one row per role: `dispatches.slot` is split on `+`
  (`unnest(string_to_array(slot, '+'))`) so a bundled dispatch counts once for each role it carried,
  and findings join their dispatch by `dispatch_id` (the `-dispatch-seq` every finding records), no
  longer by `(change_id, slot)` — a role with zero findings still counts its dispatch. Columns
  within the period: dispatches, changes, critical, major, minor, findings per dispatch, deferred
  share, withdrawn share. `experimental` is `slot LIKE 'exp-%'`; `description` is the newest
  `decisions.decision->'panel'->'roster'` entry naming that slot. The SPA badges experimental
  rows and shows `description` behind an info icon (`title` attribute plus a visible-on-focus
  tooltip).
- **`decisions`** — one row per `decisions` row: change, recorded at, class (+ `↑` when
  overridden), execution, implementer model/effort, roster size, compact, experimental slot,
  rerun, grouping (`static` / `free` / `default`), dispatches (the `+`-joined groups ` · `-joined,
  one string so equal groupings group together), implementer groups (the same shape over bundle
  ids, or empty); joined to the run by `session_token`: wall-clock (`stage_runs` first begin to last end),
  input/output/cache tokens and cost summed from `dispatches.metrics`, findings by severity,
  fix rounds (max `findings.round`), fallback and timed-out dispatch counts. A summary table above
  it groups by class × execution with means of the same columns.

## Decisions

### Who decides, and when

**ID:** planner-decides-at-end-of-d
**Status:** active
**Chosen:** the planner, as the last step of writing-plans, in its `## Plan` return — it already runs on `PLANNING_MODEL` and holds the whole plan; no extra dispatch, no gate.
**Considered:** a separate decision subagent between `## Plan` and the conductor — one more dispatch per run for a seam nobody asked for; visibility "before the design-approval gate" as KAN-472 wrote it — impossible, the plan the decision is sized from does not exist at that gate.

### What inline replaces

**ID:** inline-parent-implements-panel-dispatches
**Status:** active
**Chosen:** the parent implements and fixes; panel slots and the verifier stay subagents — a session reviewing its own diff is not a review.
**Considered:** everything inline, panel included — fewest subagents, no review independence; dropping only the conductor — smallest change, smallest saving, and not what KAN-475 describes.

### Class from mechanical inputs with a bounded override

**ID:** mechanical-class-one-step-override
**Status:** active
**Chosen:** counts from `tasks.md` and the worktree set decide the class; the planner may raise it one step with a recorded reason.
**Considered:** mechanical only — misses the two-task change to a concurrency path; planner judgment outright — not reproducible, which KAN-475 requires.

### Thresholds

**ID:** thresholds-5-12-15-40
**Status:** active
**Chosen:** the table under **Inputs and the class** — tuned later from stored decisions.
**Considered:** the first proposal (≤3/≤6, ≥8/≥20, a lone migration or second repo forcing big) — rejected by the operator as too low; two classes only — loses roster sizing for big.

### The v1 tree

**ID:** tree-v1-with-effort
**Status:** active
**Chosen:** the table under **The tree**, per-slot model and effort included, stored per run.
**Considered:** model only, effort deferred — would leave KAN-475's effort question unrecorded from day one and cost a second schema change.

### Enforcement

**ID:** agent-definitions-universal-handshake
**Status:** active
**Chosen:** nine `flow-<model>-<effort>` agent definitions + `model` parameter, `Model:` handshake on every role, fallback + one retry, second mismatch stops with a question.
**Considered:** keep continue-on-mismatch — the fallback row would be the only record, and KAN-475 asks for a guarantee; `model` parameter alone with effort as prose — effort unenforced.

### Decision storage

**ID:** decisions-jsonb-row
**Status:** active
**Chosen:** one `decisions` JSONB row per run, unique on `(change_id, session_token)`, GIN-indexed.
**Considered:** one column per cell — a migration per tuning round; stashing it in `dispatches.notes` — invisible to any aggregate.

### Experimental slot marking

**ID:** exp-slot-prefix
**Status:** active
**Chosen:** the slot id itself carries `exp-`; no schema change, distinguishable in every existing record.
**Considered:** a boolean on `findings` and `dispatches` — two columns and every writer touched, for a fact the id already states.

### Three independent toggles

**ID:** three-toggles
**Status:** active
**Chosen:** `## execution mode`, `## implementer model`, `## review panel`, each `default`|`dynamic`.
**Considered:** folding execution into implementer model — a repo could not keep sdd fixed while varying its model; one toggle — contradicts KAN-472's "each independently".

### Context ceiling

**ID:** ceiling-250k-400k-6-bundles
**Status:** active
**Chosen:** stop below 250k before a bundle, 400k before the panel, 6 bundles where no figure exists; a printed clear-and-resume handoff; the stored decision is reused on resume.
**Considered:** rely on auto-compaction — contradicts KAN-475's explicit stop-and-clear.

### Deferred minors

**ID:** deferred-status-minor-only
**Status:** active
**Chosen:** a third terminal status `deferred <reason>`, refused on any severity but Minor, closed for both guards, listed in the handoff; no Jira write.
**Considered:** reuse `withdrawn` — loses the operator-vs-policy distinction in every stat; a Jira follow-up per run — one more outward write per change.

### Both stats views now

**ID:** both-views-now
**Status:** active
**Chosen:** `reviewers` and `decisions` in this change — the decisions view is what makes "tune later" possible.
**Considered:** reviewers only — KAN-475's analysis would wait on a follow-up; one combined view — denser page, no simpler code.

### One change, seven groups

**ID:** one-change-seven-groups
**Status:** active
**Chosen:** one branch, one panel, one PR; the split is the plan's structure, and `plan-dispatch-bundles.sh` waves the independent groups.
**Considered:** seven sub-changes with Jira subtasks — seven brainstorms and panels, later ones blocked on the first merge; two changes — stats deferred.

### KAN-475 in Jira

**ID:** kan-475-linked-closed-with-472
**Status:** active
**Chosen:** link KAN-475 to KAN-472 now, close both at archive; the parent session performs the writes.
**Considered:** leave it open for a manual scope check; do nothing.

### Two dispatches per round, everywhere

**ID:** two-dispatch-cap-everywhere
**Status:** active
**Chosen:** ≤2 review dispatches × ≤3 roles per round on both `## review panel` values; `default` grouped by the static table's logic with no roll; ≤2 implementer dispatches in flight per wave on both `## execution mode` values.
**Considered:** cap on `dynamic` only — the operator stated the cap as a requirement, not a mode, and `default` runs would then produce no comparable rows; cap the panel but not implementers — two vocabularies for one saving.

### Bugbot and Security as prompt roles

**ID:** bugbot-security-prompt-roles
**Status:** active
**Chosen:** retire the `subagent_type` dispatch and the substitution mechanism on both toggle values; each is a general-purpose role with its own prompt file, recorded like every other slot.
**Considered:** always substitute inside a bundle on `dynamic` — keeps a "substitute" that is no longer standing in for anything; keep the real agents for `default` only — two mechanisms behind one slot id, and a `reviewers` row blending two different reviewers; drop them from the dynamic big roster — loses the security angle; a real Bugbot alone in one dispatch — the big roster shrinks to four.

### One `dispatches` row per bundle, compound slot

**ID:** compound-slot-one-row-per-bundle
**Status:** active
**Chosen:** one row, `-slot` the roles `+`-joined in roster order, findings keep their single role; the `reviewers` view unnests the compound slot and joins findings by dispatch id.
**Considered:** one row per bundled role sharing an `-agent-id` — the harvester drops a batch whose agent id matches two windows, so every bundle's cost would be lost; a `dispatches.bundle` column — a migration and every writer touched for a fact the id states.

### Independent passes, no de-duplication

**ID:** independent-passes-no-dedup
**Status:** active
**Chosen:** one verbatim INDEPENDENT PASSES paragraph; each pass re-raises a defect under its own angle; the parent records every row.
**Considered:** parent de-duplication by roster order — fewer rows, and per-role catch counts that no longer say what each pass caught; no paragraph — anchoring unmeasured and unmitigated.

### Grouping rolled: 30% static, 70% free

**ID:** grouping-roll-70-free-30-static
**Status:** active
**Chosen:** `bundle_roll < 30` takes the class's static grouping with no override; otherwise the planner groups freely within the cap and records why.
**Considered:** one static default with reasoned deviation — samples one point, converges immediately, the opposite of what the stored decisions are for; a named catalogue of three or four variants per class rolled uniformly — comparable but bounded to hand-picked shapes; free every run — no stable baseline to compare against.

### Experimental slot skipped over the cap

**ID:** exp-skipped-over-cap
**Status:** active
**Chosen:** a big full roster plus an experimental slot is seven roles; the experimental slot is skipped and recorded `skipped — bundle cap`.
**Considered:** displace `mutation` for that run — confounds both the grouping and the roster data; a per-bundle max of four for that case — a second cap for one case.

### Implementer groups free, no roll

**ID:** implementer-groups-free-no-roll
**Status:** active
**Chosen:** the planner records merge groups over `plan-dispatch-bundles.sh`'s bundles on every sdd run, free judgment with a one-line reason, ≤2 in flight per wave, no per-group ceiling.
**Considered:** the panel's 70/30 roll and static table — implementer merging is about focus, not a measured comparison; a cap with no merging — queues the excess and saves nothing; a per-group bundle ceiling — a number nobody asked for.

### KAN-478 joined as the eighth group

**ID:** kan-478-joined-eighth-group
**Status:** active
**Chosen:** KAN-478 (bundled reviewers, bundled implementers) is a duplicate of this change's scope, linked and closed with KAN-472; its work is Group 8 of the plan, appended after the seven landed groups because a ticked task's commit exists and cannot be re-opened.
**Considered:** a separate follow-up change — the panel's dispatch shape would ship twice; editing the landed tasks' text — a plan that no longer describes its own commits.

### Every roster carries a code-quality reviewer, floored on `primary` + `simple-reviewer`

**ID:** compact-roster-always-code-quality
**Status:** active
**Chosen:** `primary` + `simple-reviewer`, bundled as one dispatch, is the universal floor for
every roster this tree assigns — compact or full, every class. `primary` checks plan alignment
only (proposal, design, each task's declared fields), never code quality, which is
`simple-reviewer`'s and Bugbot's job. Small's original row shipped with `primary` alone in its
compact roster — no code-quality reviewer at all, a real coverage gap; the fix went through three
widening drafts before landing here: first scoped to small's compact roster only reusing
`code-review-low`, then a distinct `simple-reviewer` id scoped to small's compact roster only,
then widened to every class's compact roster, and finally — this record — to every roster the
tree assigns, compact and full alike, retiring `code-review-low` from the tree entirely (its own
record, below).
**Considered:** leaving small's compact roster at `primary` alone — leaves 70% of small-class
compact runs with zero code-quality coverage; stopping at "every class's compact roster" without
extending the floor to full rosters too — leaves regular's and big's full rosters (30% of those
classes' runs) still pairing `primary` with a role that used to be `code-review-low` but was not
guaranteed to sit in the same dispatch as `primary`, which is what "floor" now means; each
narrower scope was this change's own earlier draft, corrected by the operator in turn.

### `simple-reviewer` is a new persistent slot, `code-review-low` retired from this tree

**ID:** simple-reviewer-new-slot
**Status:** active
**Chosen:** the floor bundle's code-quality reviewer is a brand-new persistent slot,
`simple-reviewer` — its own distinct prompt file (`skills/flow/simple-reviewer-prompt.md`, the same
depth/scope as `code-review-low`'s brief: high-confidence defects only, general-purpose + prompt,
no `subagent_type`), its own id in `stats/internal/store/settings.go`'s `ValidReviewers`, and its
own row in `skills/flow/review-panel.md`'s **The roster** table — never a rename or a reuse of
`code-review-low`. **`code-review-low` is retired from this tree entirely** — it no longer appears
in any class's roster, compact or full — because there is no coherent reason to run two different
"cheap code-quality check" personas across one roster's dispatches, and the floor bundle already
covers that role's purpose everywhere `code-review-low` used to appear. `code-review-low` remains a
legal id only on the unrelated `## review panel` `default` path (the pre-existing settings-store
roster), which this change does not touch. The `reviewers` stats view needs no change: task 22's
`Reviewers` query counts by `findings.slot`/`dispatches.slot` generically (confirmed, not assumed,
by reading `stats/internal/store/aggregate.go`'s `Reviewers` — no slot id is hardcoded anywhere in
that query), so a new slot id is visible in the view the moment a run dispatches it.
**Considered:** keeping `code-review-low` alongside `simple-reviewer` in full rosters (regular's
and big's full rosters would then carry both) — two personas doing the same job in one roster, with
no distinguishing angle between them, is redundant coverage the `reviewers` view could not usefully
tell apart by what each one actually caught; reusing `code-review-low` for the floor bundle itself
rather than a new persona — rejected in the decision above's own history, for the same
distinguishability reason.

### The floor bundle absorbs a reading role when the remainder overflows

**ID:** floor-bundle-absorbs-reading-role-on-overflow
**Status:** active
**Chosen:** on a full roster whose non-floor roles exceed the second dispatch's ≤3 cap (big: four
remaining roles — `principles`, `mutation`, `bugbot`, `security`), the floor bundle's own spare
capacity (it holds 2 of its 3 possible roles) absorbs one more role before the second dispatch is
built — specifically the reading/judgment role among the remainder (`principles`), so the second
dispatch holds only the defect-hunting/mutating roles (`mutation`, `bugbot`, `security`, exactly
three). This reconciles the mandatory floor bundle with the pre-existing "reading roles together,
mutating roles together" categorization: big's static grouping becomes
`primary+simple-reviewer+principles` · `mutation+bugbot+security`, still two dispatches of at most
three, nothing left over. Regular's two remaining roles (`principles`, `mutation`) already fit the
second dispatch without needing the floor's spare slot, so this rule is inert for regular and
small.
**Considered:** capping the floor bundle at exactly two roles always, moving `principles` into a
would-be third dispatch — violates the hard ≤2-dispatches-per-round ceiling this whole feature
exists to enforce; dropping a role from big's full roster to make it fit two dispatches of three
without borrowing the floor's spare slot — loses coverage for a preference of which dispatch a role
sits in, not a real constraint.

### Compact rosters always bundle, deterministically

**ID:** compact-roster-always-bundles
**Status:** active
**Chosen:** a compact roster's roles (the floor bundle, every class — three roles since
`principles-joins-the-floor` below) are the floor bundle itself — one dispatch, deterministic, no
roll, the same mechanics as any other bundle (one `dispatches` row, compound `-slot`, the
bundle-prompt PASS-section shape).
**Considered:** dispatching a compact roster's roles separately, as the original (pre-fix) design
stated — no reason not to bundle a small, fixed roster into one dispatch: nothing to sample or
vary, since there is only one sensible grouping, so bundling is a free halving (or better) of
dispatch count with no coverage or independence cost (each PASS still starts from
`final-review.diff` fresh, per **INDEPENDENT PASSES**).

### `principles` joins the floor bundle

**ID:** principles-joins-the-floor
**Status:** active
**Chosen:** the mandatory floor bundle is `primary` + `simple-reviewer` + `principles` — three
roles, one dispatch, on every roster this tree assigns, compact or full, every class. It already
uses the bundle cap's full three-role capacity, so it never has spare room for anything else.
Small's roster gains `principles` for the first time (it was never part of small's roster at all
before this record) at `haiku/medium`, the same economy tier `simple-reviewer` already runs small
class at. This resolves cleanly against every class: small's roster is the floor alone (no second
dispatch); regular's one remaining role (`mutation`) fills the second dispatch alone; big's three
remaining roles (`mutation`, `bugbot`, `security` — all defect-hunting/mutating) fill the second
dispatch exactly, three of three. Because the floor now holds every reading/judgment role there
is, the second dispatch is always purely defect-hunting/mutating, and the **overflow case
`floor-bundle-absorbs-reading-role-on-overflow` existed to handle no longer arises — that record
is retired** (marked below), its own reasoning superseded by this simpler, uniform floor.
**Consequence, stated plainly:** a rolled experimental slot can no longer join the floor (three of
three roles already used, every class) or a compact roster (which has no second dispatch); it
joins the second dispatch only on regular (alongside `mutation`, room for one more) or is skipped
everywhere else (`experimental: skipped — bundle cap`) — small always, big always (its second
dispatch is always the full `mutation+bugbot+security`), and regular once that dispatch already
holds two roles. Small class therefore never runs the experimental slot under this tree, a real
narrowing from the pre-floor design, accepted as the cost of a uniform three-role floor.
**Considered:** keeping the floor at two roles (`primary` + `simple-reviewer`) and layering
`principles` on top like every other non-floor role — the shape this change replaces; rejected
because it left `principles` review no more guaranteed than any other layered role, and the
operator's own stated reason for this change was to make it as mandatory as `simple-reviewer`
already is.

### `floor-bundle-absorbs-reading-role-on-overflow` — retired

**ID:** floor-bundle-absorbs-reading-role-on-overflow
**Status:** retired
**Chosen (historical, no longer in force):** on a full roster whose non-floor roles exceeded the
second dispatch's ≤3 cap (big: four remaining roles under the two-role floor), the floor bundle's
own spare capacity absorbed one more role — the reading/judgment role among the remainder — before
the second dispatch was built.
**Superseded by:** **`principles-joins-the-floor`**, above. Once the floor holds three roles
(`primary` + `simple-reviewer` + `principles`) instead of two, big's remaining roles drop from four
to three (`mutation`, `bugbot`, `security`) and fit the second dispatch exactly — there is no
overflow left for this record's mechanism to resolve. Kept here, marked retired, rather than
deleted, so the reasoning that produced it (and the record of it being superseded, not simply
wrong) stays visible.

## Open questions

*(none)*
