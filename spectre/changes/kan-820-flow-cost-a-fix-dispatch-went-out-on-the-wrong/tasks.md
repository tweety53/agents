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
  - [x] **Step 4: Run the prose guards.** Run: `scripts/check-vocabulary.sh` and `scripts/check-references.sh` and `scripts/check-contract-budget.sh` and `scripts/check-model-keys.sh` and `scripts/check-markdown-integrity.py` — all expected exit 0.
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
git reset -q -- spectre/changes/ openspec/changes/ docs/superpowers/ && git add -- skills/flow/verify-and-handoff.md ':(exclude)spectre/changes/' ':(exclude)openspec/changes/' ':(exclude)docs/superpowers/' && git commit -m "docs(flow): handoff summary names the resolved model and source" -m "Task-Id: 5" -- skills/flow/verify-and-handoff.md
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
