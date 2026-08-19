> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

**Goal:** Gather each change's planning context once per dispatching stage so no dispatch re-derives
it, carry the panel's located findings across the fix boundary, and measure subagent cost per
dispatch instead of per stage.

**Architecture:** Two independent halves. Half A is one Bash gather script writing
`.superpowers/sdd/dispatch-context.md`, plus the skill text that builds it and names it on every
dispatch prompt. Half B extends the existing harvest pipeline with a third breakout of the same token
delta — `dispatches.<agentId>`, built exactly as the existing `models.<model>` breakout is — and
surfaces it as an expandable row on a route that already lists stage runs.

**Tech Stack:** Bash 3.2 (guard + harness), Go 1.x (`stats/internal/harvest`), React + TypeScript +
Vitest (`stats/web`), Postgres `jsonb_deep_add` for additive metric merging.

**Spec:** `openspec/changes/kan-201-reduce-context-rediscovery-across-review-panel/design.md`, with
delta specs under that change's `specs/`.

## Global Constraints

- **Additive only, on the Go side.** `tokens.main`, `tokens.sidechain` and `models.<model>` keep
  their exact current meaning and continue to be written as before. A stage run harvested before this
  change must not become retroactively wrong.
- **Absence is never a value.** A record with no `agentId` creates no dispatch entry; a missing meta
  sidecar omits descriptors rather than guessing them. No `unknown` or `""` key is ever fabricated.
- **The bundle never gates a run.** A missing script, a missing source, or a refused path degrades to
  the pre-change dispatch shape. Nothing about the review panel is reduced or skipped on its account.
- **No new inline suppressions and no weakened lint config**, per this repository's `CLAUDE.md`.
  `cd stats && gofmt -w .` is the only auto-fix command available; the guard scripts have none.
- **No task edits `openspec/` or `docs/superpowers/`.** `/myflow-finish` commits those.
- **Bash 3.2 only** — no associative arrays in any new shell code, matching every existing guard.

## Baseline

All measured 2026-08-18 against `097046f`.

- `stats/internal/harvest` runs **76** test functions, all passing.
  <!-- measured: cd stats && go test ./internal/harvest/ -count=1 -v | grep -c '^=== RUN' @ 097046f -->
- The SPA suite runs **157** tests across 9 files, all passing.
  <!-- measured: cd stats/web && npm test @ 097046f -->
- `skills/myflow-do/SKILL.md` is **57995** bytes against a declared budget of **71317**;
  `skills/myflow-contracts/pipeline.md` is **46105** against **55728**. Both additions fit existing
  headroom, so no `scripts/check-contract-budget.sh` budget row needs raising.
  <!-- measured: wc -c over each path, and the budgets() table in scripts/check-contract-budget.sh @ 097046f -->
- `skills/myflow-do/scripts/` carries **14** guard entries and `skills/myflow-fast/scripts/` carries
  **17**; each is a relative symlink into `scripts/`.
  <!-- measured: ls skills/myflow-do/scripts skills/myflow-fast/scripts @ 097046f -->

## Facts this plan rests on

Established by reading a real transcript tree during planning, not assumed.

A subagent's transcript is its own file — `~/.claude/projects/<project>/<session>/subagents/agent-<id>.jsonl`
— alongside a sibling `agent-<id>.meta.json`. `discoverTranscripts` already walks both, and the
`.jsonl`'s lines already carry the **parent's** `sessionId`, which is why their tokens reach the right
stage run today and land in one merged `sidechain` bucket.

```json verified:read from ~/.claude/projects/-Users-tweety53-Projects/68a79072-a074-4035-a066-33d1494a1c82/subagents/agent-ac60357c3b8d0e177.meta.json during planning
{"agentType":"general-purpose","description":"Implement Task 3 checkpoint mode","toolUseId":"toolu_01DuDwNWDG136mB5nqhkgQQU","spawnDepth":1,"model":"haiku"}
```

The `.jsonl` line's own keys include `agentId`, `isSidechain`, `sessionId`, `parentUuid`, `uuid`,
`promptId`, `type`, `message`, `timestamp`, `cwd`, `version` and `gitBranch`. `agentId` is the field
this change keys on.

`0005_jsonb_deep_add.sql` sums numeric leaves at any depth and resolves two strings at one key as
last-write-wins. That is what makes `dispatches.<agentId>.tokens.*` additive across batches and
`dispatches.<agentId>.description` stable, with no merge code and no migration.

---

### 1 `scripts/gather-dispatch-context.sh` — the gather script and its harness

**Build:** green

**Files:**
- Create: `scripts/gather-dispatch-context.sh`
- Create: `scripts/test-gather-dispatch-context.sh`
- Create: `scripts/lib/within-root.sh`
- Create: `scripts/lib/lexical-normalize.sh`
- Modify: `scripts/lib/resolve-file.sh`
- Modify: `scripts/test-check-guard-symlinks.sh`
- Modify: `scripts/gather-self-review-context.sh`
- Modify: `.myflow/project.md`

**Allowed-collateral:** *(none)*

**Tests:** `scripts/test-gather-dispatch-context.sh`

**Regression:** Reverting this commit removes the only deterministic way to build a dispatch context
bundle, returning every panel slot, implementer and fix subagent to locating five sources
independently — the exact waste KAN-201 records.

**Baseline:** before=0 after=21 — twenty-one cases (54 assertions), from a corpus of zero before the
harness exists. Fourteen at plan time; the panel's leaf-symlink finding added three, its
principles-path and root-path findings added three more, and its refusal-diagnostic finding one.
<!-- measured: grep -c '^# CASE' scripts/test-gather-dispatch-context.sh and the harness's own ok: line count @ branch openspec/kan-201-reduce-context-rediscovery-across-review-panel -->

**Commit:** `feat(kan-201-reduce-context-rediscovery-across-review-panel): gather dispatch context once per stage`

**Interfaces:**
- Consumes: nothing.
- Produces: `scripts/gather-dispatch-context.sh`, invoked as
  `gather-dispatch-context.sh <worktree> <change-root> <name> <principles-path>`, printing the bundle
  to stdout. Exit 2 on a malformed invocation; exit 0 otherwise, always. Task 2 invokes it.

The script is modelled on `scripts/gather-self-review-context.sh` and reuses its established
mechanisms rather than inventing new ones: the change-name allowlist, `within_root()`, and above all
the lexical-normalize / semantic-resolve / compare-for-exact-equality validation that file's header
documents as the single general mechanism replacing four bounded, individually-bypassable checks.
Read that header before writing this script.

- [x] **Step 1: Write the failing harness with its first case**

Create `scripts/test-gather-dispatch-context.sh` in the shape of
`scripts/test-gather-self-review-context.sh` — a sandboxed git repository per case under `TMPDIR`,
an indexed `TREES=()` array (bash 3.2 has no associative arrays) removed by an `EXIT` trap, and
`fail`/`pass` helpers counting into `FAILURES`.

```bash unverified:the exact fixture helper is written during implementation; the shape is copied from test-gather-self-review-context.sh
new_repo() {
  REPO="$(mktemp -d)"
  TREES+=("$REPO")
  git -C "$REPO" init -q
  mkdir -p "$REPO/openspec/changes/demo/specs/cap"
  printf 'why\n'   > "$REPO/openspec/changes/demo/proposal.md"
  printf 'how\n'   > "$REPO/openspec/changes/demo/design.md"
  printf 'tasks\n' > "$REPO/openspec/changes/demo/tasks.md"
  printf 'delta\n' > "$REPO/openspec/changes/demo/specs/cap/spec.md"
  printf 'principles\n' > "$REPO/PRINCIPLES.md"
}
```

The first case asserts the whole contract in one shot: a fully-populated change yields a bundle
carrying all five sources.

- [x] **Step 2: Run it and watch it fail**

Run: `scripts/test-gather-dispatch-context.sh`
Expected: FAIL — `gather-dispatch-context.sh` does not exist yet.

- [x] **Step 3: Write the script**

Structure, in order: `set -euo pipefail`; four positional arguments with a usage message on stderr
and exit 2 when any is empty; the change-name `case` allowlist copied from
`gather-self-review-context.sh` (one leading alphanumeric, then letters, digits, `.`, `_`, `-`);
`within_root()`; `validate_path()` applying the lexical/real comparison to each of `<worktree>`,
`<change-root>` and `<principles-path>`; then the bundle.

Emit the header first, then a found/skipped census line, then one `skipped: <src> (absent)` line per
missing source, then each found source under its own `## ` heading — the same document order
`gather-self-review-context.sh` uses, so a reader of one recognises the other.

```bash unverified:exact header wording is settled during implementation; the two fields it must carry are fixed by the spec
echo "# Dispatch context bundle for $NAME"
echo "generated: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
echo "head: $(git -C "$WORKTREE" rev-parse --short HEAD 2>/dev/null || echo unknown)"
```

The five sources, in order: `<change-root>/proposal.md`, `<change-root>/design.md`,
`<change-root>/tasks.md`, every `<change-root>/specs/*/spec.md`, and `<principles-path>`'s content.

Do **not** resolve `## standards` entries and do **not** derive the principles path — the spec
forbids both, and the second is the reason the path is a parameter.

- [x] **Step 4: Run the first case and watch it pass**

Run: `scripts/test-gather-dispatch-context.sh`
Expected: the first case passes.

- [x] **Step 5: Add the remaining thirteen cases**

Each asserts against the delta spec's scenarios, never against whatever the script happens to print —
the mistake `test-check-plan-provenance.sh`'s header records:

1. all five sources present → all five appear (step 1's case)
2. `design.md` absent → `skipped: design.md (absent)` and exit 0
3. every source absent → five skipped lines, exit 0
4. two delta specs → both appear
5. no `specs/` directory → reported skipped, exit 0
6. header carries a generated instant
7. header carries the `HEAD` sha
8. a `## standards` file exists in the fixture → its content does **not** appear in the bundle
9. missing fourth argument → exit 2
10. change name containing `/` → exit 2
11. change name containing a glob metacharacter → exit 2
12. `<change-root>` reached through a symlinked ancestor → exit 2
13. `<principles-path>` that is itself a symlink → exit 2
14. `<change-root>` outside the repository → exit 2

Cases 12–14 are the ones that justify reusing `gather-self-review-context.sh`'s validation wholesale;
write them before writing the validation, and confirm each fails first.

- [x] **Step 6: Run the full harness**

Run: `scripts/test-gather-dispatch-context.sh`
Expected: 14 `ok:` lines, no `FAIL:`, exit 0.

- [x] **Step 7: Register the harness**

Add `scripts/test-gather-dispatch-context.sh` to `.myflow/project.md`'s `## test` list, immediately
after `scripts/test-gather-self-review-context.sh`. Add nothing to `## lint`: this script answers a
question about one change in flight and cannot run against a bare tree, exactly like
`check-finish-preflight.sh` and the three other helpers that section already excludes for that reason.

- [x] **Step 8: Verify the repository's own guards still pass**

Run: `scripts/check-vocabulary.sh && scripts/check-references.sh && scripts/check-guard-symlinks.sh`
Expected: all exit 0. `check-guard-symlinks.sh` passes here because no skill text invokes the new
guard yet — task 2 adds both the invocation and the symlinks that satisfy its rule 2.

---

### 2 Skill text — build the bundle, name it on every dispatch, carry the fix dossier

**Build:** green

**Files:**
- Modify: `skills/myflow-do/SKILL.md`
- Modify: `skills/myflow-contracts/pipeline.md`
- Modify: `scripts/check-cleanup-complete.sh`
- Create: `skills/myflow-do/scripts/gather-dispatch-context.sh` *(relative symlink to `../../../scripts/gather-dispatch-context.sh`)*
- Create: `skills/myflow-fast/scripts/gather-dispatch-context.sh` *(relative symlink to `../../../scripts/gather-dispatch-context.sh`)*

**Allowed-collateral:** *(none)*

**Tests:** *(none — this task changes prompt and contract text, so it adds no test; its
verification is the repository's own lint guards, listed in the steps below)*

**Regression:** Reverting this commit leaves the gather script installed but never invoked, so every
dispatch continues to locate its own context and the fix round continues to re-derive locations the
panel already recorded.

**Baseline:** before=0 after=0

**Commit:** `feat(kan-201-reduce-context-rediscovery-across-review-panel): dispatch every subagent with the gathered bundle`

**Interfaces:**
- Consumes: task 1's `gather-dispatch-context.sh` and its four-argument invocation.
- Produces: the bundle path `<worktree>/.superpowers/sdd/dispatch-context.md` and its registry row,
  which nothing later in this plan consumes.

- [x] **Step 1: Create the two symlinks**

```bash unverified:run during implementation
ln -s ../../../scripts/gather-dispatch-context.sh skills/myflow-do/scripts/gather-dispatch-context.sh
ln -s ../../../scripts/gather-dispatch-context.sh skills/myflow-fast/scripts/gather-dispatch-context.sh
```

Both targets are relative, per `check-guard-symlinks.sh` rule 1 — an absolute target would bake this
machine's checkout path into the repository. `skills/myflow-finish/scripts/` gets **no** symlink: no
`/myflow-finish` step invokes this guard.

- [x] **Step 2: Add the build point to section 4**

In `skills/myflow-do/SKILL.md`, immediately after the `do.sdd-tdd` stage-begin mark and before the
`plan-dispatch-bundles.sh` invocation, add the gather invocation. Name the guard by **basename**, per
**Guard resolution** (`skills/myflow-contracts/pipeline.md`) — a repository-relative `scripts/…` path
in an invoking position is a `check-guard-symlinks.sh` rule 3 violation.

```bash unverified:the surrounding wording is written during implementation; the basename form and the four arguments are fixed
gather-dispatch-context.sh <worktree> <changeRoot> <name> <principles-path> \
  > <worktree>/.superpowers/sdd/dispatch-context.md
```

State in prose that a non-zero exit is reported and dispatching continues without a bundle, and that
`<principles-path>` is the same absolute `[PRINCIPLES_PATH]` section 5 already requires be resolved
before dispatching any principles slot.

- [x] **Step 3: Add the build point to section 5, and before each fix round**

Add the same invocation immediately after the `do.review-panel` stage-begin mark, and a third under
**Panel re-runs**, before the fix subagent is dispatched. State explicitly that the bundle is
rebuilt at these points rather than gathered once per run, and why: a fix documented under section 3
edits `proposal.md` and `tasks.md`, so a run-scoped bundle would describe a plan that no longer
exists.

- [x] **Step 4: Add the CONTEXT BUNDLE paragraph to all three dispatch prompt families**

Add this block to the implementer dispatch requirements in section 4 (beside the existing
`REQUIRED SUB-SKILL` and `REQUIRED READING` blocks), to every panel slot's dispatch in section 5, and
to the fix subagent's dispatch under **Panel re-runs**:

```markdown unverified:wording settled during implementation; the two clauses it must carry are fixed by the delta spec
> **CONTEXT BUNDLE:** `<abs-worktree>/.superpowers/sdd/dispatch-context.md` carries this change's
> proposal, design, plan, delta specs and engineering principles, gathered for you — you need not go
> looking for them. You may open any file it names. You **must** still read the actual diff and the
> actual code you are reviewing or changing: the bundle is shared *input*, never a substitute for the
> source, and never a shared conclusion.
```

The final clause is not decoration — it is KAN-201's own constraint that sharing inputs is permitted
while sharing conclusions is not, placed where a dispatch will read it.

Slots dispatched by `subagent_type` (Bugbot, Security) receive it as prompt text, exactly the way
they already receive the mutation-testing brief and the reproducer requirement. No agent definition
is edited.

- [x] **Step 5: Extend the fix dossier**

Under **Panel re-runs**, where the surviving findings are given to one fix subagent as the combined
list, state that each finding is carried as a structured block: its `F<n>`, the slot that raised it,
its severity, its `file:line` **taken verbatim from the findings table's Location column**, its theme,
its `finding-reproducer:` text, and any bounce recorded against its defect identity. State that these
locations are established and are not to be re-derived, and state that **no source excerpt is
inlined** — the fix round edits the code it is given, so an excerpt would be invalidated by the
fixer's own work.

Change nothing about the reproducer verification that precedes dispatch, the re-run-and-materiality
condition that closes a finding, the bounce accounting, or the operator handback.

- [x] **Step 6: Add the registry row**

In `skills/myflow-contracts/pipeline.md`'s **Temporary artifacts registry**, add a row directly below
the panel record and SDD ledger rows:

```markdown unverified:column wording matched to the surrounding table during implementation
| Dispatch context bundle | `/myflow-do` | `.superpowers/sdd/` in the worktree | with the worktree, at run 2 |
```

The registry states that an artifact no row accounts for is a defect in the registry, so this row is
required rather than optional.

- [x] **Step 7: Run the guards**

Run: `scripts/check-vocabulary.sh && scripts/check-references.sh && scripts/check-guard-symlinks.sh && scripts/check-contract-budget.sh && scripts/check-markdown-integrity.py`
Expected: all exit 0. `check-guard-symlinks.sh` now requires the two symlinks step 1 created; the
budget check passes on the headroom the Baseline section measures.

---

### 3 Harvest reads `agentId` and the meta sidecar

**Build:** green

**Files:**
- Modify: `stats/internal/harvest/transcript.go`
- Modify: `stats/internal/harvest/transcript_test.go`

**Allowed-collateral:** *(none)*

**Tests:** `TestParseAssistantRecordsCarriesAgentID`, `TestParseAssistantRecordsAgentIDAbsent`,
`TestReadDispatchMeta`, `TestReadDispatchMetaAbsent`, `TestReadDispatchMetaMalformed`

**Regression:** Reverting this commit removes the only source of a dispatch identifier and its
descriptors, so task 4's breakout has nothing to key on and subagent cost collapses back into one
merged `sidechain` bucket per stage.

**Baseline:** before=76 after=81

**Commit:** `feat(kan-201-reduce-context-rediscovery-across-review-panel): read agentId and dispatch descriptors`

**Interfaces:**
- Consumes: nothing.
- Produces: `harvest.Record.AgentID string`; and
  `func ReadDispatchMeta(transcriptPath string) (DispatchMeta, bool)` returning
  `DispatchMeta{AgentType, Description, Model string; SpawnDepth int}` with `false` when the path is
  not a subagent transcript or its sidecar is absent or unreadable. Task 4 consumes both.

- [x] **Step 1: Write the failing tests for `AgentID`**

In `transcript_test.go`, add a test asserting that an assistant line carrying `"agentId":"abc123"`
parses to a `Record` whose `AgentID` is `"abc123"`, and a second asserting that a line with no
`agentId` parses to a `Record` whose `AgentID` is `""` — never a placeholder.

- [x] **Step 2: Run them and watch them fail**

Run: `cd stats && go test ./internal/harvest/ -run 'AgentID' -count=1 -v`
Expected: FAIL — `Record` has no field `AgentID`.

- [x] **Step 3: Add the field and decode it**

Add `AgentID string` to `Record` and `AgentID string \`json:"agentId"\`` to `rawLine`, and carry it
through in the loop that builds each `Record` — the same one-line addition `IsSidechain` already has.
`rawLine`'s doc comment already commits to naming only the fields this package needs, so nothing else
changes.

- [x] **Step 4: Run them and watch them pass**

Run: `cd stats && go test ./internal/harvest/ -run 'AgentID' -count=1 -v`
Expected: PASS.

- [x] **Step 5: Write the failing tests for the sidecar**

Add three tests over a `t.TempDir()` fixture: a `subagents/agent-x.jsonl` with a well-formed
`agent-x.meta.json` beside it returns the four descriptors and `true`; the same with no sidecar
returns `false`; a sidecar containing malformed JSON returns `false`. A path that is not under a
`subagents/` directory returns `false` without touching the filesystem.

- [x] **Step 6: Run them and watch them fail**

Run: `cd stats && go test ./internal/harvest/ -run 'DispatchMeta' -count=1 -v`
Expected: FAIL — `ReadDispatchMeta` is undefined.

- [x] **Step 7: Implement `ReadDispatchMeta`**

```go unverified:field names verified against a real meta.json during planning; the Go shape is written during implementation
type DispatchMeta struct {
	AgentType   string `json:"agentType"`
	Description string `json:"description"`
	Model       string `json:"model"`
	SpawnDepth  int    `json:"spawnDepth"`
}
```

Derive the sidecar path by replacing the transcript's `.jsonl` suffix with `.meta.json`, and return
`false` for any path whose parent directory is not named `subagents`. **Every failure returns
`false`, never an error** — a missing or unreadable sidecar must not stop a dispatch's tokens being
attributed, which is the spec's own rule.

- [x] **Step 8: Run them and watch them pass**

Run: `cd stats && go test ./internal/harvest/ -count=1`
Expected: PASS, 81 test functions.

- [x] **Step 9: Format and vet**

Run: `cd stats && gofmt -w . && go vet ./... && gofmt -l .`
Expected: `go vet` silent, `gofmt -l` prints nothing.

---

### 4 The `dispatches.<agentId>` breakout

**Build:** green

**Files:**
- Modify: `stats/internal/harvest/attribute.go`
- Modify: `stats/internal/harvest/attribute_test.go`
- Modify: `stats/internal/harvest/watcher.go`
- Modify: `stats/internal/harvest/watcher_test.go`
- Modify: `stats/internal/store/pricing.go`
- Modify: `stats/internal/store/stageruns_test.go`
- Modify: `stats/internal/store/harvest_test.go`

**Allowed-collateral:** `stats/internal/harvest/wireshape_test.go`

**Tests:** `TestAttributeSplitsByAgentID`, `TestAttributeNoAgentIDCreatesNoDispatch`,
`TestEncodePatchesCarriesDispatchDescriptors`, `TestEncodePatchesOmitsAbsentDescriptors`,
`TestExistingTokenKeysUnchanged`, `TestPriceDispatchGetsCostThroughSamePricingPath`,
`TestPriceDispatchWithNoRecordedModelGetsNoCost`,
`TestPriceDispatchModelWithNoPricingRowGetsNoCost`,
`TestPriceDispatchWithUnknownCacheSplitGetsNoCost`,
`TestPriceDispatchLookupErrorDoesNotDiscardModelsResult`

**Regression:** Reverting this commit returns every dispatch's tokens to one merged `sidechain`
bucket per stage run, making a review slot indistinguishable from a fix round — the measurement gap
KAN-201 exists to close.

**Baseline:** before=81 after=91 in `internal/harvest`, plus six new `TestPriceDispatch*`
tests in `internal/store`. The store tests arrived with the review fix that moved per-dispatch
pricing into the Go pricing path, which is why this task's file list reaches `internal/store` at all.
Eighty-six at plan time; four review rounds added a genuine-zero spawn-depth case, a re-send stability
case and two late-sidecar backfill cases to `watcher_test.go`, one of this task's own declared files.
<!-- measured: cd stats && go test ./internal/harvest/ -count=1 -v | grep -c '^=== RUN' @ branch openspec/kan-201-reduce-context-rediscovery-across-review-panel -->

**Commit:** `feat(kan-201-reduce-context-rediscovery-across-review-panel): attribute subagent tokens per dispatch`

**Interfaces:**
- Consumes: task 3's `Record.AgentID` and `ReadDispatchMeta`.
- Produces: `Delta.Dispatches map[string]TokenDelta`; `DispatchBucket{Tokens TokenDelta;
  AgentType, Description, Model string; SpawnDepth int}`; and
  `MetricsPatch.Dispatches map[string]DispatchBucket \`json:"dispatches,omitempty"\``, reaching the
  metrics bag at `dispatches.<agentId>`. Task 5 reads that key.

This task follows `Delta.Models` exactly. Read `Delta`'s and `MetricsPatch`'s own doc comments first:
they state why the per-model split exists and why a record with no model creates no entry, and the
per-dispatch rule is that rule applied to a second key.

- [x] **Step 1: Write the failing attribution tests**

Assert that two records with different `AgentID`s against one stage run produce two entries in
`Delta.Dispatches`, each carrying only its own record's usage; and that a record with an empty
`AgentID` contributes to `Total` and to `Models` while creating **no** entry in `Dispatches`.

Add a third test pinning the constraint that matters most: for a batch of mixed records,
`Total.Main`, `Total.Sidechain` and every `Models` entry are byte-identical to what they were before
this field existed.

- [x] **Step 2: Run them and watch them fail**

Run: `cd stats && go test ./internal/harvest/ -run 'Attribute(SplitsByAgentID|NoAgentID)|ExistingTokenKeys' -count=1 -v`
Expected: FAIL — `Delta` has no field `Dispatches`.

- [x] **Step 3: Add `Dispatches` to `Delta` and fill it**

Add `Dispatches map[string]TokenDelta` to `Delta` and, in `Attribute`'s existing loop, add one line
beside the `Models` fill, guarded on a non-empty `AgentID` — the same guard the model fill already
applies to a non-empty model.

- [x] **Step 4: Run them and watch them pass**

Run: `cd stats && go test ./internal/harvest/ -run 'Attribute(SplitsByAgentID|NoAgentID)|ExistingTokenKeys' -count=1 -v`
Expected: PASS.

- [x] **Step 5: Write the failing encode tests**

Assert that `encodePatches`, given a delta with dispatch entries and a `DispatchMeta` for the file,
produces a `MetricsPatch` whose `Dispatches` carries the tokens **and** the four descriptors; and
that with no meta available it carries the tokens with the descriptor fields empty, never invented.

- [x] **Step 6: Run them and watch them fail**

Run: `cd stats && go test ./internal/harvest/ -run 'EncodePatches' -count=1 -v`
Expected: FAIL — `encodePatches` takes one argument.

- [x] **Step 7: Thread the meta through**

Add `Dispatches map[string]DispatchBucket` to `MetricsPatch`, change `encodePatches` to
`encodePatches(deltas map[int64]Delta, meta DispatchMeta, hasMeta bool)`, and in `RunOnce`'s per-file
loop call `ReadDispatchMeta(path)` once and pass the result to it. The loop already has `path` in
hand, which is why the sidecar read belongs there and not inside `Attribute`, which sees only records.

Descriptors are strings and an int at fixed keys, so `jsonb_deep_add` resolves them last-write-wins
while the nested token leaves sum — which is exactly the behaviour both need.

- [x] **Step 8: Run the whole package**

Run: `cd stats && go test ./internal/harvest/ -count=1`
Expected: PASS, 86 test functions.

- [x] **Step 9: Run the whole Go suite with the race detector**

Run: `cd stats && go test ./... -race -count=1`
Expected: PASS. This is the step that catches a store or API test asserting on the metrics bag's
exact shape.

- [x] **Step 10: Format and vet**

Run: `cd stats && gofmt -w . && go vet ./... && gofmt -l .`
Expected: `go vet` silent, `gofmt -l` prints nothing.

---

### 5 Per-dispatch rows under a stage run

**Build:** green

**Files:**
- Modify: `stats/web/src/views/RunDetail.tsx`
- Modify: `stats/web/src/views/RunDetail.test.tsx`
- Modify: `stats/web/src/api.ts`
- Modify: `stats/web/src/components/StageRunTable.tsx`
- Modify: `stats/web/src/metrics.ts`

**Allowed-collateral:** `stats/web/src/format.ts`

**Tests:** `expands a stage run into its dispatches`, `orders dispatch rows by cost descending`,
`renders an unmeasured dispatch as unavailable`, `adds no view`

**Regression:** Reverting this commit leaves per-dispatch cost recorded in the metrics bag and
readable nowhere in the interface, which is the state the "build a measurement harness" decision
exists to move past.

**Baseline:** before=157 after=163 — 161 at plan time, then +1 for reading the real per-dispatch
cost and +2 for distinguishable toggles; the panel's F3 fix then nested the rows inside
`StageRunTable`'s existing per-row detail, which removed the toggle-uniqueness test along with the
workaround it existed to cover. That nesting is why this task's file list reaches
`StageRunTable.tsx` and `metrics.ts`.

**Commit:** `feat(kan-201-reduce-context-rediscovery-across-review-panel): show per-dispatch cost under a stage run`

**Interfaces:**
- Consumes: task 4's `dispatches.<agentId>` key on a stage run's metrics bag, already reaching the
  SPA through the stage run's existing `metrics` field.
- Produces: nothing later in this plan consumes.

- [x] **Step 1: Write the failing tests**

In `RunDetail.test.tsx`, add four tests over a fixture stage run whose metrics carry three dispatch
entries — one of them with tokens but no descriptors, and one with descriptors but no tokens:

1. the stage run's row expands to three dispatch rows;
2. they are ordered by cost descending;
3. the entry with no tokens renders through the existing `Unavailable` component rather than as `0`;
4. the set of views the interface serves is unchanged.

- [x] **Step 2: Run them and watch them fail**

Run: `cd stats/web && npx vitest run src/views/RunDetail.test.tsx`
Expected: FAIL — no expandable row exists.

- [x] **Step 3: Type the dispatch shape**

In `api.ts`, add the row type for one dispatch — agent id, agent type, description, model, tokens,
cost — with every field except the agent id optional, since the spec requires an absent descriptor
and an absent measurement both be representable.

- [x] **Step 4: Render the expandable row**

Follow `RunDetail.tsx`'s own local pattern. Its header comment records why it builds its own panel
chrome rather than importing `components/Panel.tsx` — the two carry structurally different unions —
and the same reasoning applies to this sub-table: reproduce the markup locally rather than bending a
shared component to a second, differently-shaped caller.

Sort by cost descending, treating an absent cost as sorting last rather than as zero.

- [x] **Step 5: Run them and watch them pass**

Run: `cd stats/web && npx vitest run src/views/RunDetail.test.tsx`
Expected: PASS.

- [x] **Step 6: Run the whole SPA suite and the type check**

Run: `cd stats/web && npm test && npx tsc -b`
Expected: 163 tests pass; `tsc -b` exits clean.
<!-- predicted: cd stats/web && npm test after task 5's four cases land -->
