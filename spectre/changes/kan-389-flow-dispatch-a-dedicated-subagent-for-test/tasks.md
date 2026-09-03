# kan-389-flow-dispatch-a-dedicated-subagent-for-test

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.
> **Relocation:** no

## Global constraints

- `design.md` in this change root is canonical for the verifier's dispatch shape, report shapes,
  handshake and recording; the prose each task adds cites it by section and restates nothing that
  `skills/flow/implement.md` section 4 already states for a dispatch record.
- `~/.claude/rules/be-brief.md` — cut, never paraphrase. Every existing normative sentence in
  `skills/flow/verify-and-handoff.md` that is not replaced by the dispatch stays word-for-word.
- Every `scripts/check-*.sh` guard named in `.flow/project.md`'s `## lint` exits clean after every
  task, `scripts/check-contract-budget.sh` and `scripts/check-references.sh` included. Byte budgets
  at `d4df668`: `skills/flow/SKILL.md` 14062 of 16278, `skills/flow/verify-and-handoff.md` 21913
  of 24862, `skills/flow/implement.md` 25724 of 32500.
  <!-- measured: wc -c skills/flow/SKILL.md skills/flow/verify-and-handoff.md skills/flow/implement.md; sed -n 187,198p scripts/check-contract-budget.sh @ d4df668 -->
  A task whose growth is deliberate and cannot be trimmed under its row raises that row in
  `budgets()` of `scripts/check-contract-budget.sh` — that file is that task's allowed collateral.
- `cd stats && go vet ./... && gofmt -l .` exits clean after task 1.

---

- [x] 1. Add the `verifier` dispatch role to `flow record dispatch`

**Build:** green

**Files:**
- Modify: `stats/cmd/flow/record.go`
- Modify: `stats/cmd/flow/record_test.go`
- Modify: `skills/flow/implement.md`

**Tests:** `TestRecordAcceptsVerifierRole`
**Regression:** `TestRecordAcceptsVerifierRole` — reverting the commit makes `record dispatch
begin -role verifier` exit 2 before contacting the store, so the test's `contacted` stays false
and its exit-code assertion fails.
**Baseline:** before=135 after=136
<!-- measured: cd stats && go test ./cmd/flow/ -v 2>&1 | grep -c '^--- PASS' @ d4df668 (before); predicted: the same command after this task's commit (after) -->
**Commit:** `feat(stats): accept verifier as a flow record dispatch role`

  - [x] **Step 1: RED — write `TestRecordAcceptsVerifierRole`**

  In `stats/cmd/flow/record_test.go`, directly after `TestRecordAcceptsConductorRole`, add a test in
  `TestRecordAcceptsPlannerRole`'s exact shape (`gitRepo`, `isolatedStateRoot`, a `genuineDaemon`
  `httptest` server that flips `contacted`), with `-role verifier -model sonnet -key k1
  -session-token mf-record-verifier-role`, the server body
  `{"id":1,"seq":1,"role":"verifier","model":"sonnet","startedAt":"2026-01-02T03:04:05Z"}`, and a
  doc comment naming the role as the one `/flow` records for the subagent
  `skills/flow/verify-and-handoff.md` dispatches for `flow.verify` and `flow.visual-verify`.
  Also add `"verifier"` to the accepted-role slice the unrecognised-role test iterates (the
  `for _, role := range []string{"implementer", "reviewer", "panel-fix", "red-partner", "planner",
  "conductor"}` loop at `record_test.go:441`).

  Run: `cd stats && go test ./cmd/flow/ -run 'TestRecordAcceptsVerifierRole|TestRecordRefusesUnknownRole' -v`
  Expected: `TestRecordAcceptsVerifierRole` fails with `exit code = 2, want 0`; the unrecognised-role
  test fails on `stderr does not name the accepted role "verifier"`. (Confirm the unrecognised-role
  test's real name from the file before running — it is the test containing the loop at line 441.)

  - [x] **Step 2: GREEN — add the role**

  In `stats/cmd/flow/record.go`, change the `recordRoles` literal at line 44 to

```go verified:read from stats/cmd/flow/record.go:44 @ d4df668, one element appended
var recordRoles = []string{"implementer", "reviewer", "panel-fix", "red-partner", "planner", "conductor", "verifier"}
```

  and extend its doc comment: after the `conductor` clause, add "and verifier being the subagent
  `/flow` dispatches from skills/flow/verify-and-handoff.md to run the project's lint and test
  commands and its visual verification steps on a fixed model (design.md's
  verifier-role-recorded)". Update the comment's quoted list `implementer · reviewer · panel-fix ·
  red-partner · planner · conductor` to end `· verifier`.

  Run: `cd stats && gofmt -w . && go vet ./... && gofmt -l . && go test ./cmd/flow/`
  Expected: `gofmt -l .` prints nothing; `ok  	github.com/tweety53/agents/stats/cmd/flow`.

  - [x] **Step 3: Name the role in `implement.md`'s record sentence**

  In `skills/flow/implement.md` at line 292, change "`-role` is one of `implementer`, `reviewer`,
  `panel-fix` or `red-partner`;" to "`-role` is one of `implementer`, `reviewer`, `panel-fix`,
  `red-partner` or `verifier` (**Verify**, `skills/flow/verify-and-handoff.md`);". Nothing else
  on that line changes.

  Run: `scripts/check-references.sh && scripts/check-contract-budget.sh && scripts/check-vocabulary.sh`
  Expected: all exit 0.

  - [x] **Step 4: Commit**

```bash verified:paths are this task's own Files list
git add stats/cmd/flow/record.go stats/cmd/flow/record_test.go skills/flow/implement.md
git commit -m "feat(stats): accept verifier as a flow record dispatch role" -m "Task-Id: 1"
```

---

- [x] 2. Document `VERIFY_MODEL` in `skills/flow/SKILL.md`'s Model resolution

**Build:** green

**Files:**
- Modify: `skills/flow/SKILL.md`

**Tests:** **none** — prose only
**Regression:** n/a — no test declared
**Baseline:** n/a — no test declared
**Commit:** `docs(flow): pin the verifier dispatch to sonnet in model resolution`

  - [x] **Step 1: Add the literal to the resolution block**

  In `skills/flow/SKILL.md`'s `## Model resolution` fenced block, immediately after line 78
  (`[ -z "$PLANNING_MODEL" ] && PLANNING_MODEL=fable`), add one line:

```bash verified:literal per design.md section 4
VERIFY_MODEL=sonnet
```

  - [x] **Step 2: Add the `VERIFY_MODEL` paragraph**

  Immediately before the paragraph beginning "**`DEFAULT_MODEL` is the model for all three roles
  this run dispatches on**" (line 120), insert one paragraph in the shape of the
  `SELF_REVIEW_MODEL` and `PLANNING_MODEL` paragraphs above it, stating exactly: `VERIFY_MODEL`
  governs the two verifier dispatches — `flow.verify`'s and `flow.visual-verify`'s (**Verify** and
  **Visual verification**, `skills/flow/verify-and-handoff.md`); it is the fixed literal `sonnet`,
  read from neither the settings store nor `<project>/.flow/project.md`; a plain-language session
  instruction does not override it; and it never falls back, because it is never resolved — the
  point is a predictable model for mechanical test and verification runs regardless of what
  `DEFAULT_MODEL` resolved to. Leave the `DEFAULT_MODEL` paragraph's "all three roles" sentence
  unchanged: the verifier is not one of them.

  Run: `scripts/check-references.sh && scripts/check-contract-budget.sh && scripts/check-vocabulary.sh && wc -c skills/flow/SKILL.md`
  Expected: all exit 0; `SKILL.md` stays at or under 16278 bytes. (The paragraph must fit in the
  2216 bytes of headroom measured above — write it tight; raise the row only if it cannot.)

  - [x] **Step 3: Commit**

```bash verified:paths are this task's own Files list
git add skills/flow/SKILL.md
git commit -m "docs(flow): pin the verifier dispatch to sonnet in model resolution" -m "Task-Id: 2"
```

---

- [x] 3. Dispatch the verifier from `flow.verify`

**Build:** green

**Files:**
- Modify: `skills/flow/verify-and-handoff.md`

**Allowed-collateral:** `scripts/check-contract-budget.sh`

**Tests:** **none** — prose only
**Regression:** n/a — no test declared
**Baseline:** n/a — no test declared
**Commit:** `docs(flow): run flow.verify's lint and test commands through a verifier subagent`

  - [x] **Step 1: Add a `### The verifier dispatch` subsection under `## Verify`**

  In `skills/flow/verify-and-handoff.md`, insert a `### The verifier dispatch` subsection
  immediately after the "**This step does not call the project's `create` command.**" paragraph
  (line 62–65) and before "Run the commands `project-get.sh <worktree> lint` …" (line 67). It
  states the rules of `design.md` section 1 once, in full — `design.md` is not a stable path the
  skill can cite, so the subsection is the canonical statement and task 4 cites it:

  - one verifier per worktree per stage; `subagent_type: general-purpose`; the Agent tool's
    `model` parameter set to `VERIFY_MODEL` (**Model resolution**, `skills/flow/SKILL.md`) — the
    literal `sonnet`, never `DEFAULT_MODEL` and never a session override;
  - the prompt carries the two agent-baseline sentences verbatim (quote them, exactly as
    **Dispatch the conductor** in `skills/flow/implement.md` does);
  - the relay contract: no operator channel, fixes nothing, edits no source, every command in the
    foreground, the turn ends with a single `## Report` block, the first line of the first reply
    `Model: <the model named in its own system prompt>`;
  - the record pair, `-role verifier`, `-task` omitted, `-model sonnet`, `-key verify` here and
    `visual-verify` in **Visual verification**, suffixed `-<worktree basename>` when the resolved
    set holds more than one worktree — semantics per section 4 of `skills/flow/implement.md`,
    cited not restated;
  - the handshake: compare `Model:` to `sonnet`; a mismatch records `-outcome fallback`, opens
    `<key>-<that model, lowercased>` with the model that answered, and continues on the running
    agent — no re-dispatch (give both `flow record dispatch` commands in one fenced block tagged
    `verified:shape of the conductor handshake in skills/flow/implement.md`);
  - a mark or a record never blocks; a verifier that ends without `## Report`, or dies, is closed
    `-outcome aborted` and blocks the handoff as a failed command would, naming the death.

  - [x] **Step 2: Replace the inline lint/test run with the dispatch**

  Replace the paragraph at line 67–69 ("Run the commands `project-get.sh <worktree> lint` and
  `project-get.sh <worktree> test` print (auto-detect on exit 1) and show the output. **Nothing
  runs them later** — `/flow`'s integrate phase has no verification gate — so a non-zero exit
  blocks this handoff.") with prose stating: resolve those same commands the same way; dispatch one
  verifier per worktree whose prompt states the absolute worktree path, the `KEY=value` lines
  `prepare-workspace.sh` printed for that worktree (to export before every command), the lint
  commands then the test commands in the order printed, and the report shape below; the verifier
  runs every command in order and does not stop at the first failure. Keep the sentence "**Nothing
  runs them later** — `/flow`'s integrate phase has no verification gate — so a non-zero exit
  blocks this handoff." verbatim, now reading against the report. Then the report shape, fenced and
  tagged `verified:design.md section 2 of this change`:

```text verified:design.md section 2 of this change
## Report
- `<command>` — exit <n>
  <the command's output, verbatim, or its last 40 lines when longer, stated as truncated>
```

  and one sentence: the conductor shows the report as this stage's output.

  - [x] **Step 3: Update the file preamble and guards**

  In the file's opening paragraph (lines 3–9) add one sentence after "…not part of either decision
  above.": "Both `flow.verify` and `flow.visual-verify` run their commands through a `verifier`
  subagent (**The verifier dispatch**, below); the conductor keeps every mark and every block
  decision." Nothing else in the preamble changes.

  Run: `scripts/check-references.sh && scripts/check-contract-budget.sh && scripts/check-vocabulary.sh && wc -c skills/flow/verify-and-handoff.md`
  Expected: all exit 0. If `check-contract-budget.sh` reports `verify-and-handoff.md` over 24862,
  first cut wording in the new subsection; if it still cannot fit, raise that file's row in
  `budgets()` of `scripts/check-contract-budget.sh` to the new `wc -c` value rounded up to the next
  hundred, and say so in the commit body.

  - [x] **Step 4: Commit**

```bash verified:paths are this task's own Files list plus its allowed collateral
git add skills/flow/verify-and-handoff.md scripts/check-contract-budget.sh
git commit -m "docs(flow): run flow.verify's lint and test commands through a verifier subagent" -m "Task-Id: 3"
```

---

- [x] 4. Dispatch the verifier from `flow.visual-verify`

**Build:** green

**Files:**
- Modify: `skills/flow/verify-and-handoff.md`

**Allowed-collateral:** `scripts/check-contract-budget.sh`

**Tests:** **none** — prose only
**Regression:** n/a — no test declared
**Baseline:** n/a — no test declared
**Commit:** `docs(flow): run flow.visual-verify's steps 3-8 and 10 through a verifier subagent`

  - [x] **Step 1: Split the ten steps between conductor and verifier**

  In `## Visual verification`, after the sentence ending "…the same set **Verify** above resolved:"
  (line 102) and before the numbered list, add one paragraph: steps 1, 2 and 9 are the conductor's;
  steps 3–8 and 10 are run by one verifier per worktree that survives steps 1–2, dispatched per
  **The verifier dispatch** above with `-key visual-verify`, and the conductor applies
  **Blocking** to its report. The prompt states: the absolute worktree path; the `KEY=value` lines
  **Verify** exported for it; the section's resolved `setup`, `verify` and `capture` commands and
  `screenshots` root; the worktree-resolved URL of each app `ui paths` matched; the project's
  `## run` commands; the views this change touched; `<changeRoot>`; and the instruction to run
  steps 3–8 and 10 below as written, committing nothing and pushing nothing.

  The ten steps themselves stay word-for-word — the paragraph above assigns them; do not edit the
  step text.

  - [x] **Step 2: Add the report shape**

  Immediately before the "**Blocking.**" paragraph (line 159), add the report shape, fenced and
  tagged `verified:design.md section 3 of this change`:

```text verified:design.md section 3 of this change
## Report
- setup: <not declared | exit <n>>
- stack: <already running | started and stopped | could not be started — <output>>
- verify: exit <n>
  <output, verbatim or last 40 lines>
- capture: exit <n>
  <output, verbatim or last 40 lines>
- spec: <absolute spec path>
- <view>: <absolute PNG path> — <what was seen, including any defect>
- visual-verification.md: written | not written — <reason>
```

  followed by one sentence: every non-zero exit, unreadable PNG, `resolve-visual-screenshots.sh`
  exit 1 or 2, and defect the verifier names is carried in the report, and the `Visual:` handoff
  line is built from its view entries. In the **Blocking** paragraph, change "a defect the agent
  sees in a captured screenshot" to "a defect the verifier reports in a captured screenshot" —
  that single substitution and nothing else in that paragraph.

  - [x] **Step 3: Repoint the `Visual:` handoff sentence**

  In the paragraph "**The `Visual:` line reports `flow.visual-verify`'s own outcome.**" (line 364),
  no wording changes — confirm it still reads correctly against the report-sourced entries and
  leave it.

  Run: `scripts/check-references.sh && scripts/check-contract-budget.sh && scripts/check-vocabulary.sh && wc -c skills/flow/verify-and-handoff.md`
  Expected: all exit 0, with the same budget rule as task 3 step 3.

  - [x] **Step 4: Whole-file read-through**

  Run: `git diff d4df668..HEAD -- skills/flow/verify-and-handoff.md`
  Expected: every removed line is either the replaced lint/test paragraph (task 3 step 2) or the
  one-word **Blocking** substitution; no other pre-existing sentence is reworded.

  - [x] **Step 5: Commit**

```bash verified:paths are this task's own Files list plus its allowed collateral
git add skills/flow/verify-and-handoff.md scripts/check-contract-budget.sh
git commit -m "docs(flow): run flow.visual-verify's steps 3-8 and 10 through a verifier subagent" -m "Task-Id: 4"
```
