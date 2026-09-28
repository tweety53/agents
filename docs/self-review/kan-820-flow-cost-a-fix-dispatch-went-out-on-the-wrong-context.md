# Self-review context bundle for kan-820-flow-cost-a-fix-dispatch-went-out-on-the-wrong

found: 6 of 7 sources; skipped: 1 of 7 sources
skipped: change summary (absent)

## .superpowers/sdd/ledgers/kan-820-flow-cost-a-fix-dispatch-went-out-on-the-wrong.md

# SDD ledger — kan-820-flow-cost-a-fix-dispatch-went-out-on-the-wrong

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — implementer

- Task: 1
- Role: implementer
- Key: task-1-implementer
- Model: glm-5.3-flash effort=high
- Commit: 46ae943896002af39da03c791ee13df374584349
- Outcome: completed
- Started: 2026-09-27T22:34:17Z
- Tokens: cost unattributed — session never bound

## Dispatch 2 — implementer

- Task: 2
- Role: implementer
- Key: task-2-implementer
- Model: glm-5.3-flash effort=high
- Commit: 31403b12d6d3fcda2ce96980dd431ab16930fc0d
- Outcome: completed
- Started: 2026-09-27T22:41:46Z
- Tokens: not measured

## Dispatch 3 — implementer

- Task: 3
- Role: implementer
- Key: task-3-implementer
- Model: glm-5.3-flash effort=high
- Commit: 93cd4a493ef5ff7bf6863d09d0ffd79d92e34ce2
- Outcome: completed
- Started: 2026-09-27T22:47:46Z
- Tokens: not measured

## Dispatch 4 — implementer

- Task: 4
- Role: implementer
- Key: task-4-implementer
- Model: glm-5.3-flash effort=high
- Commit: 22326c6872c2dddeeda9fc7155a651450bf11172
- Outcome: completed
- Started: 2026-09-27T22:48:55Z
- Tokens: not measured

## Dispatch 5 — implementer

- Task: 5
- Role: implementer
- Key: task-5-implementer
- Model: glm-5.3-flash effort=high
- Commit: 14008f475dcd98c4567c4ab9426ff47fa22c81e0
- Outcome: completed
- Started: 2026-09-27T22:50:01Z
- Tokens: not measured

## Dispatch 6 — implementer

- Task: 6
- Role: implementer
- Key: task-6-implementer
- Model: glm-5.3-flash effort=high
- Commit: 00c2879243ef4733b191e2115e10b817f8c96e01
- Outcome: completed
- Started: 2026-09-27T22:52:39Z
- Tokens: not measured

## Dispatch 7 — implementer

- Task: 7
- Role: implementer
- Key: task-7-implementer
- Model: glm-5.3-flash effort=high
- Commit: 9acb87ebb45bb461f3d88e0f34b7b6c69a7014cb
- Outcome: completed
- Started: 2026-09-27T22:53:14Z
- Tokens: not measured

## Dispatch 8 — reviewer

- Task: 1
- Role: reviewer
- Slot: primary+principles
- Key: task-1+2-reviewer
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Diff base: 6cae37f0601965214d6551c3d57996293d58f6d1
- Outcome: fix
- Started: 2026-09-27T23:09:28Z
- Tokens: not measured

## Dispatch 9 — panel-fix

- Task: 1
- Role: panel-fix
- Key: task-1-implementer-fix-1
- Model: glm-5.3-flash effort=high
- Commit: 0ccf11dd8b4eec4914d3f2abc0989224a1cc195e
- Outcome: completed
- Started: 2026-09-27T23:29:23Z
- Tokens: not measured

## Dispatch 10 — reviewer

- Task: 1
- Role: reviewer
- Slot: primary+principles
- Key: task-1-reviewer-fix-1
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Diff base: 0ccf11dd8b4eec4914d3f2abc0989224a1cc195e
- Outcome: clean
- Started: 2026-09-27T23:30:46Z
- Tokens: not measured

## Dispatch 11 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: fix
- Started: 2026-09-27T23:35:03Z
- Tokens: not measured

## Dispatch 12 — panel-fix

- Task: no task
- Role: panel-fix
- Key: panel-fix-1
- Model: glm-5.3-flash effort=high
- Commit: 554c5a22281f1c24d9358a49ba81143605cf0276
- Outcome: completed
- Started: 2026-09-28T00:04:34Z
- Tokens: not measured

## Dispatch 13 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary
- Key: panel-1-primary
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Diff base: 2ef96a9eeab380fec5844254b65267da1801053b
- Outcome: fix
- Started: 2026-09-28T00:09:24Z
- Tokens: not measured

## Dispatch 14 — reviewer

- Task: no task
- Role: reviewer
- Slot: principles
- Key: panel-1-principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Diff base: 2ef96a9eeab380fec5844254b65267da1801053b
- Outcome: clean
- Started: 2026-09-28T00:09:24Z
- Tokens: not measured

## Dispatch 15 — verifier

- Task: no task
- Role: verifier
- Key: verify
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-28T00:16:32Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-820-flow-cost-a-fix-dispatch-went-out-on-the-wrong-panel.md

# Review panel — kan-820-flow-cost-a-fix-dispatch-went-out-on-the-wrong

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary+principles | Important | skills/flow/SKILL.md:56 | the resolver's tr -d backtick + xargs normalization mangles mismatched bodies into valid models instead of reporting-and-dropping, and is a second drifted implementation of the body rule the Go guard implements |   |
| F2 | primary+principles | Important | skills/flow/SKILL.md:82 | the kept outage paragraph instructs falling back to the literal opus, contradicting the implemented project-wins-on-store-outage property, and the bare settings-get read aborts under set -e before the project key is read |   |
| F3 | primary | Minor | spectre/changes/kan-820-flow-cost-a-fix-dispatch-went-out-on-the-wrong/tasks.md:132 | task 3's checked verify step names scripts/check-contract-budget.sh, which no longer exists on main or at the merge base |   |
| F4 | primary | Minor | spectre/changes/kan-820-flow-cost-a-fix-dispatch-went-out-on-the-wrong/tasks.md:185 | task 5's commit fence still stages verify-and-handoff.md, contradicting the task's own recorded correction and its actual commit |   |
| F5 | primary+principles | Minor | skills/flow/SKILL.md:52 | the outage-safe read's 2>/dev/null swallows the CLI's stderr, so a store outage resolves silently and the rewritten paragraph dropped the report-the-stderr duty |   |

findings-total: 5
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 deferred the fix diff introduced it; one-token fix deferred with the sibling archive.md stderr-visible pattern named

reproducers-total: 5
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-primary-3.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/0-primary-4.sh
finding-reproducer: F5 .superpowers/sdd/reproducers/1-primary-1.sh

## Pass log

### Round 0

- roster: compact — 22
- diff size: 523 changed lines, cap 600 — under cap, proceed
- docs-only reduction: not applied — exit 1, first non-doc path scripts/check-model-keys.sh
- no addition this round — the resolved list ran alone
## spectre/changes/archive/kan-820-flow-cost-a-fix-dispatch-went-out-on-the-wrong/tasks.md

# kan-820-flow-cost-a-fix-dispatch-went-out-on-the-wrong

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.

> **Relocation:** no

Every task is file-disjoint except the declared `After:` edge (task 7 needs task 1's guard).
Merge base: `4a278320` (origin/main). No `spectre/specs/` capability exists for project
configuration or model policy — `check-spec-reach.sh .` passes with nothing to reach, so no
spec-edit task exists.

- [x] 1. Go model-key guard validates the `## model` key
**Build:** green
**Files:** `stats/internal/guard/modelkeys.go`, `stats/internal/guard/check_model_keys_test.go`
**Tests:** `TestCheckModelKeysDispatchKey`
**Commit:** feat(guard): model-keys validates the ## model dispatch key
**Regression:** `TestCheckModelKeysDispatchKey` — reverting it leaves `.flow/project.md`'s `## model` body unvalidated, so a typo'd policy silences into the run-time drop instead of failing lint
**Baseline:** before=1 after=2
<!-- measured: grep -c '^func Test' stats/internal/guard/check_model_keys_test.go @ merge-base 4a278320 (the count BEFORE this change) -->
**After:** none

**Decision:** project-key-home

  - [x] **Step 1: RED — write the failing test.** In `stats/internal/guard/check_model_keys_test.go`, add `TestCheckModelKeysDispatchKey` beside the existing test, table-driven over three fixtures of a project root's `.flow/project.md`: body `opus` (present, valid → exit 0), body `gpt-9` (present, invalid → exit 1 naming the key and body), section absent (→ exit 0). Reuse the existing test's fixture and invocation helpers — read them first; the valid set comes from the same source the existing test uses.
  - [x] **Step 2: Run it to verify it fails.** Run: `cd stats && go test ./internal/guard/ -run TestCheckModelKeysDispatchKey -count=1 | tail -5`. Expected: FAIL — `## model` is not scanned (modelkeys.go:104 scans only `self review model`).
  - [x] **Step 3: GREEN — extend the guard.** Generalize the scanned key set in `stats/internal/guard/modelkeys.go` from the single `const key = "self review model"` to both `self review model` and `model`, keeping one shared code path (loop over a key slice; no per-key duplication). Both keys stay optional and independent: absence of either is exit 0.
  - [x] **Step 4: Run the package and the shim.** Run: `cd stats && go test ./internal/guard/ -count=1 | tail -3` (expected PASS, whole package) and `scripts/check-model-keys.sh` (expected exit 0 — this repo declares only `## self review model` today; `## model` absent is valid).
  - [x] **Step 5: Commit.**

```bash verified:pathspec, excludes and reset order mirror the FLOW — COMMIT-PER-TASK sequence, skills/flow/implement.md
cd stats && gofmt -w internal/guard && git reset -q -- spectre/changes/ openspec/changes/ docs/superpowers/ && git add -- internal/guard/modelkeys.go internal/guard/check_model_keys_test.go ':(exclude)spectre/changes/' ':(exclude)openspec/changes/' ':(exclude)docs/superpowers/' && git commit -m "feat(guard): model-keys validates the ## model dispatch key" -m "Task-Id: 1" -- internal/guard/modelkeys.go internal/guard/check_model_keys_test.go
```

- [x] 2. Model resolution resolves the project key; its guard asserts it
**Build:** green
**Files:** `skills/flow/SKILL.md`, `scripts/check-model-resolution-shell.sh`
**Tests:** `project-model-wins` `project-model-invalid-drops` `store-down-project-wins` `store-null-falls-back`
**Commit:** feat(scripts): model resolution reads the project key behind the store default
**Regression:** the four new cases — reverting them loses the proof that the extracted SKILL.md block actually implements project-beats-store, drops an invalid body to the store, survives a store outage on the project key, and falls back to the literal when the store answers null
**Baseline:** before=8 after=12
<!-- measured: grep -c '^run_case' scripts/check-model-resolution-shell.sh @ merge-base 4a278320 (the count BEFORE this change) -->
**After:** none

**Decision:** resolution-precedence

  - [x] **Step 1: RED — extend the guard's cases first.** In `scripts/check-model-resolution-shell.sh`, parameterize the default assertion: the `run_case` verdict line becomes `[[ "$got_default" != "${EXPECTED_DEFAULT:-sonnet}" ]]`. Add a fail mode to the `flow` stub — when the case's JSON argument is empty, `settings get` exits 3 (store unreachable). Then add four cases after the existing seven:

```bash verified:stub and case shapes mirror the guard's own run_case at merge-base 4a278320, read in this session
# Case 8: project ## model valid — beats the store's sonnet.
EXPECTED_DEFAULT=opus
run_case '{"defaultModel":"sonnet","reviewers":[],"selfReviewModel":""}' \
  "fable" "project-model-wins" '## model
opus'

# Case 9: project ## model invalid — dropped, the store value stands.
run_case '{"defaultModel":"sonnet","reviewers":[],"selfReviewModel":""}' \
  "fable" "project-model-invalid-drops" '## model
gpt-9'

# Case 10: store unreachable — the project key still resolves.
run_case '' \
  "fable" "store-down-project-wins" '## model
opus'

# Case 11: store answers null — the literal fallback fires.
run_case '{"defaultModel":null,"reviewers":[],"selfReviewModel":""}' \
  "fable" "store-null-falls-back"
EXPECTED_DEFAULT=sonnet
```

  - [x] **Step 2: Run the guard to verify the new cases fail.** Run: `scripts/check-model-resolution-shell.sh | tail -5`. Expected: exit 1 — `project-model-wins`, `store-down-project-wins` and `store-null-falls-back` resolve `sonnet`/wrong today because the block has no project-key arm.

    Correction (2026-09-28): the plan's case block left `EXPECTED_DEFAULT=opus` set across case 9
    (`project-model-invalid-drops`), whose expected default is the store's `sonnet` — the drop
    must leave the store value standing. As implemented, `EXPECTED_DEFAULT` is set per case
    (opus for `project-model-wins`, sonnet for the drop case, opus for the two fallback cases)
    so each case asserts its own resolution; the plan's single leading assignment would have
    failed the guard's green run.
  - [x] **Step 3: GREEN — replace SKILL.md's Model resolution block** (the first ```bash fence under `## Model resolution`) with:

```bash unverified:scripts/check-model-resolution-shell.sh extracts and runs this block verbatim once this task's cases land
SETTINGS_JSON="$(flow settings get)"
MODEL_ROOT="${MAIN_CHECKOUT:-$(cd "$(dirname "$(git rev-parse --git-common-dir)")" && pwd -P)}"
PROJECT_MODEL="$(project-get.sh "$MODEL_ROOT" 'model' 2>&1)"; rc=$?
case "$rc" in
  0) PROJECT_MODEL="$(printf '%s' "$PROJECT_MODEL" | tr -d '`' | xargs)" ;;
  1) PROJECT_MODEL="" ;;
  *) echo "⛔ flow: project-get.sh exited $rc: $PROJECT_MODEL — stop the run" >&2; exit 2 ;;
esac
DEFAULT_MODEL="$(printf '%s' "$SETTINGS_JSON" | jq -r '.defaultModel')"
MODEL_SOURCE=store
if [ -n "$PROJECT_MODEL" ]; then
  if flow settings models | grep -qx -- "$PROJECT_MODEL"; then
    DEFAULT_MODEL="$PROJECT_MODEL"
    MODEL_SOURCE=project
  else
    echo "⚠ flow: .flow/project.md '## model' body '$PROJECT_MODEL' is not a valid model — dropped" >&2
  fi
fi
[ -n "$DEFAULT_MODEL" ] && [ "$DEFAULT_MODEL" != "null" ] || { DEFAULT_MODEL=opus; MODEL_SOURCE=fallback; }
REVIEWERS="$(printf '%s' "$SETTINGS_JSON" | jq -r '.reviewers[]')"
VERIFY_MODEL=opus
```

    Keep the surrounding prose's fallback paragraph and add one sentence: the resolved source (`project`, `store` or `fallback`) is reported beside the model, per **Model policy**. Do not mention `SELF_REVIEW_MODEL`, `PLANNING_MODEL` or `_TOGGLE` in the block — the guard's drift checks refuse those.
  - [x] **Step 4: Run the guard and its sandbox harness.** Run: `scripts/check-model-resolution-shell.sh` (expected exit 0, all 11 cases) and `scripts/test-check-model-resolution-shell.sh | tail -3` (expected pass). Then `scripts/check-markdown-integrity.py` and `scripts/check-references.sh`.
  - [x] **Step 5: Commit.**

```bash verified:pathspec, excludes and reset order mirror the FLOW — COMMIT-PER-TASK sequence, skills/flow/implement.md
git reset -q -- spectre/changes/ openspec/changes/ docs/superpowers/ && git add -- skills/flow/SKILL.md scripts/check-model-resolution-shell.sh ':(exclude)spectre/changes/' ':(exclude)openspec/changes/' ':(exclude)docs/superpowers/' && git commit -m "feat(scripts): model resolution reads the project key behind the store default" -m "Task-Id: 2" -- skills/flow/SKILL.md scripts/check-model-resolution-shell.sh
```

- [x] 3. Contracts: the `## model` key row and the resolution order
**Build:** green
**Files:** `skills/flow-contracts/project-configuration.md`, `skills/flow-contracts/model-policy.md`, `scripts/check-model-keys.sh`
**Tests:** none
**Commit:** docs(contracts): the ## model key and the DEFAULT_MODEL resolution order
**Regression:** none — prose and a shim header; the behavior is tasks 1–2's
**Baseline:** before=0 after=0
**After:** Task 1 2

**Decision:** project-key-home
**Decision:** resolution-precedence
**Decision:** governed-roles
**Decision:** enforcement-strength
**Decision:** mapping-interplay

  - [x] **Step 1: project-configuration.md.** Add a `## model` row to the key table beside `## self review model`'s: optional, literal single-line body, one member of the store's `ValidModels`, matched byte-for-byte with leading/trailing whitespace trimmed and nothing else normalized, reported by name and dropped otherwise, absent → store default. Name its consumer: `DEFAULT_MODEL`'s resolution (**Model resolution**, `skills/flow/SKILL.md`), the governed roles per **Model policy**.
  - [x] **Step 2: model-policy.md.** Under **Model policy**, restate the resolution order — project `## model` → store `defaultModel` → literal `opus`, resolved per run — the governed role set (implementer, fixer, panel dispatches, rerun pair; `VERIFY_MODEL` fixed, `SELF_REVIEW_MODEL` separate), the surfacing duty (resolution names model + source; Decide preamble, run summary and `decision.json` `resolved` carry it — nothing blocks), and the mapping sentence: the key governs the pre-mapping value; on harness `zcode` the mapping still replaces the model at dispatch and the ledger records the model actually run. Never restate what project-configuration.md's row canonically says — cite it.
  - [x] **Step 3: check-model-keys.sh header.** Its header already reads "both keys are optional" — make it name them: `## self review model` and `## model`, citing project-configuration.md's rows.
  - [x] **Step 4: Run the prose guards.** Run: `scripts/check-vocabulary.sh` and `scripts/check-references.sh` and `scripts/check-model-keys.sh` and `scripts/check-markdown-integrity.py` — all expected exit 0.

    Correction (2026-09-28): the step originally named the contract-budget ratchet too; that
    guard was removed from origin/main by another change while this one ran, and the step names
    only the guards this tree carries. The ratchet did run and pass at the task's own close,
    before the rebase that brought the removal in.
  - [x] **Step 5: Commit.**

```bash verified:pathspec, excludes and reset order mirror the FLOW — COMMIT-PER-TASK sequence, skills/flow/implement.md
git reset -q -- spectre/changes/ openspec/changes/ docs/superpowers/ && git add -- skills/flow-contracts/project-configuration.md skills/flow-contracts/model-policy.md scripts/check-model-keys.sh ':(exclude)spectre/changes/' ':(exclude)openspec/changes/' ':(exclude)docs/superpowers/' && git commit -m "docs(contracts): the ## model key and the DEFAULT_MODEL resolution order" -m "Task-Id: 3" -- skills/flow-contracts/project-configuration.md skills/flow-contracts/model-policy.md scripts/check-model-keys.sh
```

- [x] 4. Decide records the resolution: preamble line and decision.json
**Build:** green
**Files:** `skills/flow/brainstorm-planner.md`
**Tests:** none
**Commit:** docs(flow): Decide names the resolved model source and records it
**Regression:** none — the Decide section's own contract text; the record it mandates is written by runs, not tests
**Baseline:** before=0 after=0
**After:** Task 3

**Decision:** enforcement-strength

  - [x] **Step 1: preamble line.** In the Decide section's two prepended lines, the `models:` line carries the source: `models:    default <DEFAULT_MODEL> (<MODEL_SOURCE>) · reviewers <REVIEWERS>`.
  - [x] **Step 2: decision.json field.** In the JSON shape paragraph, add `resolved` (an object `{model, source}`, `source` one of `project`/`store`/`fallback`, written every run that resolves, beside `rolls`). State that a session-instruction override leaves `resolved` as resolved and lands in `overrides` as today.
  - [x] **Step 3: `## Decision` block.** The block's `models:` preamble row renders the source beside the model, and the decision side's implementer/panel rule cells stay as-is (the pair's model column already prints `DEFAULT_MODEL`).
  - [x] **Step 4: Run the prose guards.** Run: `scripts/check-vocabulary.sh` and `scripts/check-references.sh` and `scripts/check-stage-mark-calls.sh` and `scripts/check-dispatch-paragraphs.sh` and `scripts/check-markdown-integrity.py` — all expected exit 0.
  - [x] **Step 5: Commit.**

```bash verified:pathspec, excludes and reset order mirror the FLOW — COMMIT-PER-TASK sequence, skills/flow/implement.md
git reset -q -- spectre/changes/ openspec/changes/ docs/superpowers/ && git add -- skills/flow/brainstorm-planner.md ':(exclude)spectre/changes/' ':(exclude)openspec/changes/' ':(exclude)docs/superpowers/' && git commit -m "docs(flow): Decide names the resolved model source and records it" -m "Task-Id: 4" -- skills/flow/brainstorm-planner.md
```

- [x] 5. Run summary names the resolved model and its source
**Build:** green
**Files:** `skills/flow-contracts/pipeline.md`
**Tests:** none
**Commit:** docs(flow-contracts): run summary names the resolved model and source
**Regression:** none — handoff prose contract
**Baseline:** before=0 after=0
**After:** Task 3

**Decision:** enforcement-strength

Correction (2026-09-28): the plan declared `skills/flow/verify-and-handoff.md` as this task's
file, but that file carries no summary section of its own — it produces the handoff's
`Running:`/`Records:`/`Costs:`/`Deferred:` parts and cites the rest — and the run-summary content
contract is canonical in `skills/flow-contracts/pipeline.md`'s **Summary and live-stack line,
before every handoff**. The bullet landed there instead, and the commit scope names the module
the task actually touched. The `check-task-commit-fields.sh` refusal on the record as it stood
named exactly this path and subject, and the deviation was judged legitimate on that refusal
before this transcription.

  - [x] **Step 1: summary line.** In the summary's requirements, add one bullet: the live-stack/summary block names the run's resolved `DEFAULT_MODEL` and its source (`project`, `store` or `fallback`), so the operator reads the run's model policy off the handoff without opening the ledger.
  - [x] **Step 2: Run the prose guards.** Run: `scripts/check-vocabulary.sh` and `scripts/check-references.sh` and `scripts/check-stage-mark-calls.sh` and `scripts/check-dispatch-paragraphs.sh` and `scripts/check-markdown-integrity.py` — all expected exit 0.
  - [x] **Step 3: Commit.**

```bash verified:pathspec, excludes and reset order mirror the FLOW — COMMIT-PER-TASK sequence, skills/flow/implement.md
git reset -q -- spectre/changes/ openspec/changes/ docs/superpowers/ && git add -- skills/flow-contracts/pipeline.md ':(exclude)spectre/changes/' ':(exclude)openspec/changes/' ':(exclude)docs/superpowers/' && git commit -m "docs(flow-contracts): run summary names the resolved model and source" -m "Task-Id: 5" -- skills/flow-contracts/pipeline.md
```

- [x] 6. flow-plan resolves no model directly
**Build:** green
**Files:** `skills/flow-plan/SKILL.md`
**Tests:** none
**Commit:** docs(flow-plan): DEFAULT_MODEL resolves via Model resolution
**Regression:** none — citation repair; the direct `flow settings get` statement bypassed the project key
**Baseline:** before=0 after=0
**After:** Task 3

**Decision:** resolution-precedence

  - [x] **Step 1: repoint the citation.** Line 186 area states `DEFAULT_MODEL` comes "from `flow settings get`" — restate it as resolving per **Model resolution** (`skills/flow/SKILL.md`), which reads the project key first. Quote no resolution order here — one source of truth.
  - [x] **Step 2: sweep for siblings.** Run: `grep -n 'settings get' skills/flow-plan/SKILL.md skills/flow-fast/SKILL.md skills/flow-settings/SKILL.md`. flow-fast already cites **Model resolution** (leave); flow-settings writes the store (leave — it is the store's own command); any other direct `flow settings get` → `DEFAULT_MODEL` statement in flow-plan gets the same repoint.
  - [x] **Step 3: Run the prose guards.** Run: `scripts/check-references.sh` and `scripts/check-markdown-integrity.py` — expected exit 0.
  - [x] **Step 4: Commit.**

```bash verified:pathspec, excludes and reset order mirror the FLOW — COMMIT-PER-TASK sequence, skills/flow/implement.md
git reset -q -- spectre/changes/ openspec/changes/ docs/superpowers/ && git add -- skills/flow-plan/SKILL.md ':(exclude)spectre/changes/' ':(exclude)openspec/changes/' ':(exclude)docs/superpowers/' && git commit -m "docs(flow-plan): DEFAULT_MODEL resolves via Model resolution" -m "Task-Id: 6" -- skills/flow-plan/SKILL.md
```

- [x] 7. This repository declares `## model opus`
**Build:** green
**Files:** `.flow/project.md`
**Tests:** none
**Commit:** chore(project-config): declare ## model opus
**Regression:** none — a declaration; the store default here is already opus, so no run's model changes
**Baseline:** before=0 after=0
**After:** Task 1 3

**Decision:** project-key-home

  - [x] **Step 1: declare the key.** Add to `.flow/project.md`, beside `## self review model`:

```markdown verified:mirrors the store's DefaultModel const, stats/internal/store/settings.go:55
## model

opus
```

  - [x] **Step 2: Run the project-config guards.** Run: `scripts/check-model-keys.sh` (now validates both keys — expected exit 0), `scripts/check-workspace-isolation.sh` and `scripts/check-visual-verification.sh .` (both parse project.md — expected untripped by the new key).
  - [x] **Step 3: Commit.**

```bash verified:pathspec, excludes and reset order mirror the FLOW — COMMIT-PER-TASK sequence, skills/flow/implement.md
git reset -q -- spectre/changes/ openspec/changes/ docs/superpowers/ && git add -- .flow/project.md ':(exclude)spectre/changes/' ':(exclude)openspec/changes/' ':(exclude)docs/superpowers/' && git commit -m "chore(project-config): declare ## model opus" -m "Task-Id: 7" -- .flow/project.md
```
## spectre/changes/archive/kan-820-flow-cost-a-fix-dispatch-went-out-on-the-wrong/design.md

# Design

## Context

KAN-820 (flow-cost): the operator's model policy lives in memory and narratives, not in
anything the dispatcher reads; each mismatch costs a wasted dispatch plus a prose correction.
One change because the policy's home, its resolution order and its surfacing are one mechanism
split across five contract/skill files plus this repo's own project config — editing any
subset would leave the dispatcher reading two disagreeing sources.

Constraints: the settings store stays harness-wide and schema-frozen; `VERIFY_MODEL` stays a
fixed literal; `SELF_REVIEW_MODEL` keeps its own key and precedence; the zcode harness mapping
replaces every dispatch's model at dispatch and is not changed; a run-scoped operator
instruction still overrides in either direction and stays recorded; projects declaring nothing
must resolve exactly as today.

## Approach

One new literal-body key, `## model`, in `<project>/.flow/project.md`, and a new first step in
`DEFAULT_MODEL`'s resolution. Everything downstream of the resolution — effort selection, the
review panel's shape, the override rule, the zcode mapping — is unchanged. The resolution is
surfaced (preamble line, run summary) and recorded (`decision.json` `resolved` object) so the
pre-dispatch value and its provenance are machine-readable, per the finding's ask.

The key is validated exactly like `## self review model`, the precedent: one member of the
settings store's `ValidModels`, body matched byte-for-byte with leading/trailing whitespace
trimmed and nothing else normalized; a body matching no member is reported by name and dropped,
resolving as if the key were absent. `flow-settings` keeps governing the store only — the
project key is edited in the project's own committed config, which is what makes it
machine-readable to the dispatcher and reviewable in pull requests.

Because the key lives in the repository, it stays readable when the settings store is
unreachable: a project's declared policy survives a store outage, and the fallback chain
(project → store → literal `opus`) degrades in the same order of specificity.

## Decisions

### The per-project model policy lives in a project.md key

**ID:** project-key-home
**Status:** active
**Chosen:** `## model` in `<project>/.flow/project.md`, one `ValidModels` member, validated like `## self review model` — in-repo, dispatcher-readable, no store migration
**Considered:** settings-store per-project rows — a store schema change plus an authoring surface for exactly what a committed project file already carries; per-change state field `models.default` — policy becomes per-change, every change re-decides it, and `/flow` deliberately never asks

### The project key is the first source; store default and the literal follow

**ID:** resolution-precedence
**Status:** active
**Chosen:** project `## model` → store `defaultModel` → literal `opus`, resolved once per run — the most specific policy wins, and the key still resolves during a store outage
**Considered:** store-first — a harness-wide default would shadow every project's policy, making the key dead on any machine with a store row

### The key governs every role that reads DEFAULT_MODEL today

**ID:** governed-roles
**Status:** active
**Chosen:** implementer, fixer, every panel dispatch, the rerun pair — the resolution's source changes, the consumer set does not; `VERIFY_MODEL` stays the fixed literal, `SELF_REVIEW_MODEL` keeps its own key and precedence, effort stays the Decide step's
**Considered:** implementer+fixer only — matches the incident narrowly but splits one resolution into two sources; per-role key shape — a parsed multi-row section and more resolution rules for the same fix

### A policy violation is surfaced and recorded, never gated

**ID:** enforcement-strength
**Status:** active
**Chosen:** resolution names model + source once; the Decide preamble `models:` line and the run summary carry both; `decision.json` gains `resolved: {"model": …, "source": …}` — visible before dispatches go out, nothing blocks
**Considered:** hard pre-dispatch gate — on zcode it could only ever key on the pre-mapping value, blocking runs over a value the harness replaces anyway; lint-time guard — sees only this repository, not the consumer projects the policy is for

### The key governs the pre-mapping value on harnesses that map

**ID:** mapping-interplay
**Status:** active
**Chosen:** the zcode mapping still replaces every dispatch's model at dispatch; resolution and the decision/summary records name the resolved policy model, and the dispatch ledger keeps recording the mapped model actually run — no contract change to the mapping
**Considered:** key inert on mapping harnesses — honest about power but the policy never bites on this machine; policy declares the mapped model — makes the key per-harness in practice, defeating "records the operator's policy"

## Open questions

## Consequences

- Projects declaring nothing resolve exactly as today — store default, then the literal; no
  behavior change anywhere the key is absent.
- The mismatch cost the finding names (wasted dispatch, operator correction) becomes visible at
  resolution time: the run states its model and where it came from before anything dispatches,
  and an operator instruction that overrides the key is recorded beside a named source instead
  of riding in prose alone.
- This repository's own `.flow/project.md` gains `## model opus`, which equals the store
  default here — the path is exercised by every subsequent run of this repo without changing
  which model resolves.
## spectre/changes/archive/kan-820-flow-cost-a-fix-dispatch-went-out-on-the-wrong/narrative.md

# kan-820-flow-cost-a-fix-dispatch-went-out-on-the-wrong — session narrative

## 2026-09-28 — creating run

- The bare `/flow` that started this change first resolved, unanswered, toward kan-842 — whose
  worktree turned out to be mid-flight in a concurrent session (commits landing during this
  session's own turn). This run stood down there and took kan-820 only on the operator's word.
- The state file read `STARTED` for the whole implementation: `flow.write-in-progress` is the
  only step entitled to advance it, so a crash mid-run leaves `STARTED` behind a fully built
  branch — worth remembering when a resumed run reads an implausibly early state.
- One store outage journaled a stage mark mid-run (`⚠ flow: store unreachable — wrote local
  journal`); it replayed without intervention.
- The base moved mid-run (5 commits on origin/main, overlapping `.flow/project.md` and
  `skills/flow/brainstorm-planner.md`). The operator chose rebase. The rebase was textually
  clean but semantically not: main's new guard-test edits and this branch's appended test
  function merged to drop the `sync` import — a broken build no exit code caught. The gated
  per-task reviewer caught it (its dispatched sha was also stale post-rebase, which is how it
  noticed); fixed as one commit on top and re-reviewed clean. Lesson priced in: a clean rebase
  proves nothing about Go imports.
- The same base movement removed `check-contract-budget.sh` from the tree; task 3's plan step
  and task 5's commit fence both needed record corrections, the latter riding the fields guard's
  own refusal.
- Task 5's file premise was wrong — the run-summary contract lives in
  `skills/flow-contracts/pipeline.md`, not `verify-and-handoff.md` — corrected through the same
  refusal-then-transcribe route.
- Of five AskUserQuestion prompts this run asked, two went unanswered (candidate resolution,
  first plan gate), two were answered (design gate, base-moved), and the plan gate answered on
  the operator's re-invocation. The panel then raised 3 Important + 2 Minor; both Importants
  were real defects in the resolution block (body normalization, outage prose), fixed in one
  round, re-run clean, one new Minor deferred.
## 2026-09-28 — integrate run

- Preflight RUN1; unfinished-work gate CLEAR; the sync rebase onto origin/main met 55 new
  upstream commits and one real conflict — KNOWN-BUGS.md, where another change's deferred entries
  and this change's landed at the same anchor. Resolved in place keeping both sides; the
  resolution-needing rebase triggered the full lint and test lists, all green. No recorded
  baseline to recapture.
- The landing route is this project's configured default (merge and push), taken without asking.

- The second sync rebase (origin/main moved twice during the route) landed after the reshape, so
  the post-rebase lint hit `check-task-records` structurally: the plan's per-task Commit subjects
  are unreachable from HEAD because reshape-branch had already collapsed the seven task commits
  into the one implementation commit, as run 1 requires. Hand-verified against the primary
  records — seven of seven tasks ticked, every task's paths present in 8ee62ad3, no open
  findings — and overridden as structural; the store's verdict record only covers
  check-unfinished-work, so this line is the record.
## git log --stat

commit 8ee62ad32e63d25280b3d42128a4a6826b7f0522
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 28 03:38:29 2026 +0300

    feat(flow): the ## model key and the project-first model resolution

 .flow/project.md                               |  4 ++
 KNOWN-BUGS.md                                  |  4 ++
 scripts/check-model-keys.sh                    | 10 ++---
 scripts/check-model-resolution-shell.sh        | 35 +++++++++++++--
 skills/flow-contracts/model-policy.md          | 16 +++++++
 skills/flow-contracts/pipeline.md              |  3 ++
 skills/flow-contracts/project-configuration.md |  3 +-
 skills/flow-plan/SKILL.md                      |  2 +-
 skills/flow/SKILL.md                           | 34 +++++++++++++--
 skills/flow/brainstorm-planner.md              |  8 +++-
 stats/internal/guard/check_model_keys_test.go  | 60 ++++++++++++++++++++++++++
 stats/internal/guard/modelkeys.go              | 33 +++++++-------
 12 files changed, 181 insertions(+), 31 deletions(-)

commit b9c0857f45400bbf7c4b0e3c9df76afc29c2dab8
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 28 03:50:22 2026 +0300

    chore(spectre): plan

 .../narrative.md                                                  | 8 ++++++++
 1 file changed, 8 insertions(+)

commit 115313e08a5a798b8e89119fae7e04bc14a30c30
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 28 03:57:40 2026 +0300

    chore(spectre): archive kan-820-flow-cost-a-fix-dispatch-went-out-on-the-wrong

 .../design.md                                      |   0
 .../ledger.md                                      | 178 +++++++++++++++++++++
 .../narrative.md                                   |   0
 .../panel.md                                       |  34 ++++
 .../proposal.md                                    |   0
 .../tasks.md                                       |   0
 6 files changed, 212 insertions(+)
