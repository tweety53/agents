---
name: myflow-state-advance
description: Advance a myflow change to an explicit stage. Pure state write — validates the incoming stage, writes the new one, prints the next step. Used by all /myflow-*-done and /myflow-*-manual-review commands.
allowed-tools: Bash(openspec:*), Bash(git:*), Bash(jq:*)
license: MIT
compatibility: Requires openspec CLI and jq.
metadata:
  author: gymie
  version: "1.0"
---

Advance one OpenSpec change to an explicit pipeline stage. **This skill only writes state.** It
never reads artifacts, never runs git write operations, never verifies anything. That separation
is deliberate — see **Pipeline stages** in `rules/myflow-manual-review.mdc`.

**Announce at start:** "Using myflow-state-advance: `<name>` → `<TARGET_STAGE>`."

Follow **rules/myflow-manual-review.mdc** — sections **Stage transitions**, **State file**,
**State self-heal**, **IntelliJ commands**.

## Parameters (supplied by the invoking command)

- `TARGET_STAGE` — the stage to write. Either a **static** literal or the **dynamic** form
  `originStage` — see **Target forms** below.
- `ACCEPTED_STAGES` — the stage(s) this transition may start from.

## Target forms

- **Static target** (six of the seven callers) — the invoking command supplies a literal stage
  string, e.g. `TARGET_STAGE: do-done`. Write it as given.
- **Dynamic target** (`/myflow-do-fix-done` only) — the invoking command supplies
  `TARGET_STAGE: originStage`, meaning: read the state file's `originStage` field and target
  *that* stage, per **Fix re-entry**. In this form:
  - If `originStage` is `null` or missing, **stop** and report it — do not guess a stage.
  - **Validate the value against the six legal origins** — `awaiting-do-review`,
    `do-review-started`, `do-done`, `awaiting-manual-test`, `manual-test-done`,
    `awaiting-pr-review`. Anything else (including `awaiting-fix-review`, `fix-review-started`,
    `proposal-done`, `finished`, or a typo/legacy string) is **corruption, not a stage**: **stop
    and report** it. Never target it, never fall back to a default, never self-heal it into a
    plausible value. Report the actual value, the six legal ones, and tell the user to re-run
    `/myflow-do-fix <name>` from a legal origin (which rewrites `originStage`) or to correct the
    state file by hand.
  - If `originStage` is `do-review-started`, target `awaiting-do-review` instead — per **Fix
    re-entry**, the diff changed, so the prior review is stale and review restarts. The other five
    origins target themselves.
  - After writing the new stage, also clear `originStage` to `null` in the same write.

## Workflow

### 1. Resolve the change

```bash
openspec list --json
```

With a name, use it. Without one, filter to non-archived changes whose current stage is in
`ACCEPTED_STAGES`; exactly one match → use it and announce; multiple → **AskUserQuestion**;
zero → stop and say which stage was expected and what to run instead.

### 2. Read and validate state

Resolve the state file per **State file**:

```bash
MAIN_CHECKOUT="$(cd "$(dirname "$(git rev-parse --git-common-dir)")" && pwd)"
PROJECT_KEY="$(basename "$MAIN_CHECKOUT")-$(printf '%s' "$MAIN_CHECKOUT" | shasum | cut -c1-8)"
STATE_FILE="/Users/tweety53/Agents/myflow/state/$PROJECT_KEY/<name>.json"
jq -r '.stage' "$STATE_FILE"
```

If the current stage is not in `ACCEPTED_STAGES`, **stop** and emit the mismatch handoff from
**Stage transitions**, then AskUserQuestion for an explicit override (default: **No — run the
suggested command instead**). Missing or contradicted file → self-heal per **State self-heal**,
announce the correction, then re-apply this check to the healed stage.

If `TARGET_STAGE` is the dynamic form (`originStage`), resolve the actual target now per
**Target forms** before proceeding — read `originStage` from the same state file, validate it
against the six legal origins, apply the `do-review-started` special case, and stop if it is
`null`/missing or holds any value outside that set.

### 3. Write the new stage

Write the full object, preserving every field you did not change:

- `stage` → `TARGET_STAGE` (the literal value, or the resolved dynamic target)
- `updatedAt` → ISO-8601 UTC now
- `updatedBy` → the invoking command, e.g. `"/myflow-do-done"`
- when `TARGET_STAGE` was the dynamic form, also `originStage` → `null`
- **everything else carried forward verbatim** — all four `gates` values, `worktree`, `branch`,
  `originStage` (unless just cleared above), `artifactUrl`, `jiraIssue`, `fastPath`,
  `REVIEWED_TREE`, `MERGE_BASE`, and any other field present in the file you did not explicitly
  change

**Gate values are monotonic** — never lower one, never infer `gates.tested: true`, never overwrite
`"skipped"`. This skill has no reason to change a gate at all; if you find yourself writing one,
stop.

### 4. Hand off

Report the transition and what comes next. When the new stage waits on a human, include the
IntelliJ command from **IntelliJ commands** with an absolute path:

```
## Stage advanced

**Change:** <name>
**Stage:** <old> → <TARGET_STAGE>
**Next:** <the command or human action that follows>

**Open in IntelliJ:**
open -na "IntelliJ IDEA" --args "<absolute path>"
```

Omit the IntelliJ block when the next step is an agent command rather than a human action.

## Guardrails

- **Never** verify, test, commit, push, merge, or archive — this skill writes one field (plus
  `originStage` when clearing it under the dynamic form).
- **Never** touch Jira. `*-done` / `*-manual-review` commands are pure state writes; giving one an
  external side effect would break that invariant — see **Jira integration** in
  `rules/myflow-manual-review.mdc`. Carry `jiraIssue` forward untouched.
- **Never** advance from a stage outside `ACCEPTED_STAGES` without an explicit user override.
- **Never** modify gates; carry them forward untouched.
- **Never** guess a dynamic target — if `originStage` is `null`/missing, stop and report it.
- **Never** target an out-of-range `originStage`. Only the six origins listed in **Target forms**
  are legal; any other value is corruption — stop and report, never repair it into a stage.
- Never guess a change name when multiple match.
