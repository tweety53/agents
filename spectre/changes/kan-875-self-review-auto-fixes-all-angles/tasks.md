# kan-875-self-review-auto-fixes-all-angles

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Tasks appended:** 1

**Goal:** `/flow-self-review` fixes and lands every non-`big` finding across all six angles on one
branch behind one `opus` review loop, records every finding's outcome in the flow store, and files
only the `big` ones.

**Architecture:** Tasks 1–3 add the `self_review_findings` row bottom-up — store, then API and
client, then the `flow self-review finding`/`findings` CLI. Task 4 teaches the report guard the
`fixed: <sha>` disposition. Task 5 rewrites the skill to use both. Task 6 exercises the real CLI
against a real daemon.

**Spec:** `design.md` beside this file is canonical for every scope decision. Tasks cite its
`## Decisions` by ID and never restate them.

## Global Constraints

- A commit's scope names the module it moved (`rules/commit-scope-is-the-module.mdc`).
- Go tasks follow the shape of the existing incident record end to end — `records.Incident`
  (`stats/internal/records/types.go`), `Store.RecordIncident`/`ListIncidents`
  (`stats/internal/store/records.go`), `recordHandler.recordIncident`/`listIncidents`
  (`stats/internal/api/records.go`), `Client.RecordIncident`/`ListIncidents`
  (`stats/internal/client/client.go`) — and the existing `flow self-review bundle` path
  (`stats/internal/api/selfreview.go`, `stats/internal/client/selfreview.go`,
  `stats/cmd/flow/selfreview.go`) for where the new code lives.
- Every Go task's verify step runs `cd stats && gofmt -w . && gofmt -l . && go vet ./...` plus its
  own targeted tests; store tests need `flow-postgres` on 5433 (`stats/docker-compose.yml`), which
  is already up — never stop it (`CLAUDE.md`, "Never stop the dev workspace's stats service").
- Every run-loaded sentence Task 5 adds, deletes or rewords is listed, exactly as
  `scripts/check-verbatim-moves.sh` prints it after `::`, in
  `spectre/changes/kan-875-self-review-auto-fixes-all-angles/verbatim-moves.txt`, under a first line
  `# Operator, 2026-10-04: KAN-875 — self-review fixes every non-big finding.` The file is edited in
  the worktree and **left uncommitted**: it is a planning path, carried by the next
  `chore(spectre): plan` commit (**Planning commits**, `skills/flow-contracts/git-boundaries.md`).

## Review Focus

- **Closed vocabulary.** The store refuses a disposition outside `fixed`/`filed`/`declined`, a
  `fixed` without a sha-shaped ref and a `filed` without a key-shaped ref.
- **Never blocks.** `flow self-review finding` exits 0 with one warning line when the store is
  unreachable (`store-write-never-blocks`).
- **Sha after landing.** The skill reads each fixed finding's sha off `<default-branch>` after the
  rebase, never from the pre-rebase branch.
- **Main checkout.** The fix worktree is never the main checkout.

---

- [x] 1. Store: the `self_review_findings` table and its write/list

**Files:** `stats/internal/store/migrations/0035_self_review_findings.sql`, `stats/internal/records/types.go`, `stats/internal/store/selfreviewfindings.go`, `stats/internal/store/selfreviewfindings_test.go`
**Tests:** `TestSelfReviewFindingsRoundTrip`, `TestSelfReviewFindingRefusesBadDisposition`
**Regression:** `TestSelfReviewFindingsRoundTrip` fails without the table or the list ordering;
`TestSelfReviewFindingRefusesBadDisposition` fails if the store accepts an unknown disposition, a
`fixed` row without a sha or a `filed` row without an issue key.
**Baseline:** before=274 after=276
<!-- measured: cat stats/internal/store/*_test.go | grep -c '^func Test' @ 69463842 -->
**Commit:** `feat(store): record self-review findings`
**After:** none
**Build:** green

**Decision:** self-review-findings-table

  - [x] **Step 1: Write the failing tests** in `stats/internal/store/selfreviewfindings_test.go`
    on `newTestStore(t)` (`stats/internal/store/testsupport_test.go`): round-trip — record three
    rows for one change (`fixed` with ref `abc1234` and blast radius 3, `filed` with `KAN-9`,
    `declined` with no ref and no blast radius), list them back in insertion order with every
    field intact and the nil blast radius still nil; refusal — `bogus` disposition, `fixed` with
    ref `KAN-9`, `filed` with ref `abc1234`, `declined` with a ref, each an error wrapping a new
    `ErrSelfReviewFindingInvalid` sentinel. Run: `cd stats && go test ./internal/store -run
    'TestSelfReviewFinding' -count=1` — fails to compile.
  - [x] **Step 2: Migration** `0035_self_review_findings.sql`: table `self_review_findings` —
    `id BIGSERIAL PRIMARY KEY`, `project_key TEXT NOT NULL REFERENCES projects(project_key)`,
    `change TEXT NOT NULL`, `angle TEXT NOT NULL`, `note TEXT NOT NULL`,
    `disposition TEXT NOT NULL CHECK (disposition IN ('fixed','filed','declined'))`, `ref TEXT`,
    `blast_radius INT`, `recorded_at TIMESTAMPTZ NOT NULL DEFAULT now()`; index on
    `(project_key, change)`. Header comment states why it is not the panel `findings` table
    (`design.md`, `self-review-findings-table`). `change` is text, not an FK: the change is
    `FINISHED` and possibly archived when the row is written.
  - [x] **Step 3: Type** `records.SelfReviewFinding` in `stats/internal/records/types.go`, beside
    `Incident`: `ID int64`, `Change string`, `Angle string`, `Note string`, `Disposition string`,
    `Ref string` (`json:"ref,omitempty"`), `BlastRadius *int` (`json:"blastRadius,omitempty"`),
    `RecordedAt time.Time`, camelCase JSON tags.
  - [x] **Step 4: Store** `stats/internal/store/selfreviewfindings.go`:
    `RecordSelfReviewFinding(ctx, projectKey string, in records.SelfReviewFinding)` validating
    before the insert — disposition in the closed set; `fixed` ref matches `^[0-9a-f]{7,40}$`;
    `filed` ref matches `^[A-Z][A-Z0-9]*-[0-9]+$`; `declined` ref empty; `angle` and `note`
    non-empty — and `ListSelfReviewFindings(ctx, projectKey, change string)` ordered by `id`.
  - [x] **Step 5: Verify** — the Step 1 command passes; gofmt/vet per Global Constraints.
  - [x] **Step 6: Commit.**

- [x] 2. API and client: `/api/v1/self-review/{project}/{change}/findings`

**Files:** `stats/internal/api/selfreview.go`, `stats/internal/api/selfreview_test.go`, `stats/internal/api/server.go`, `stats/internal/client/selfreview.go`, `stats/internal/client/selfreview_test.go`, `stats/internal/api/records.go`, `stats/internal/api/changes_test.go`, `stats/internal/client/client_test.go`, `stats/internal/reconcile/record_test.go`, `stats/internal/web/embed_test.go`
**Tests:** `TestSelfReviewFindingsEndpoint`, `TestClientSelfReviewFindings`
**Regression:** `TestSelfReviewFindingsEndpoint` fails without the POST/GET routes or if a store
validation error is not a 400; `TestClientSelfReviewFindings` fails if the client's URL, method or
JSON shape drifts from the handler's.
**Baseline:** before=235 after=237
<!-- measured: cat stats/internal/api/*_test.go stats/internal/client/*_test.go | grep -c '^func Test' @ 69463842 -->
**Commit:** `feat(api): serve self-review findings`
**After:** Task 1
**Build:** green

**Decision:** self-review-findings-table

  - [x] **Step 1: Write the failing tests**, following the existing tests in each file:
    endpoint — POST a valid row → 201 and the stored row back; POST `bogus` disposition → 400;
    GET → the posted rows in order; client — `RecordSelfReviewFinding` and
    `ListSelfReviewFindings` against an `httptest` server asserting method, path and body. Run:
    `cd stats && go test ./internal/api ./internal/client -run 'SelfReviewFindings' -count=1` —
    fails to compile.
  - [x] **Step 2: Handler and routes** — `selfreviewHandler.recordFinding` and `.listFindings` in
    `stats/internal/api/selfreview.go`, the store interface it holds gaining the two Task 1
    methods; `POST` and `GET /api/v1/self-review/{project}/{change}/findings` registered beside
    the bundle route in `server.go`. `ErrSelfReviewFindingInvalid` maps to 400.
  - [x] **Step 3: Client** — `RecordSelfReviewFinding(ctx, project, change, in)` and
    `ListSelfReviewFindings(ctx, project, change)` in `stats/internal/client/selfreview.go`.
  - [x] **Step 4: Verify** — the Step 1 command passes; gofmt/vet.
  - [x] **Step 5: Commit.**

Correction (2026-10-04): the plan declared the store interface change in `selfreview.go` alone;
`server.go` hands the handler `rs`, an `api.RecordStore`, so the two methods also joined
`RecordStore` (`stats/internal/api/records.go`), and every test stub implementing it gained them —
`changes_test.go`'s `fakeStore` field, `client_test.go`'s `stubStageStore`,
`reconcile/record_test.go`'s `nopRecordStore`/`fakeRecordStore`, `web/embed_test.go`'s `fakeStore`.

- [x] 3. CLI: `flow self-review finding` and `flow self-review findings`

**Files:** `stats/cmd/flow/selfreview.go`, `stats/cmd/flow/selfreview_test.go`, `.flow/project.md`, `scripts/test-flow-addr-declaration.sh`
**Tests:** `TestSelfReviewFindingCLI`, `TestSelfReviewFindingsCLI`
**Regression:** `TestSelfReviewFindingCLI` fails if a missing required flag is not exit 2, or if
an unreachable store is not exit 0 with exactly one warning line on stderr;
`TestSelfReviewFindingsCLI` fails if the read does not print the rows as a JSON array or exits 0
on an unreachable store.
**Baseline:** before=262 after=264
<!-- measured: cat stats/cmd/flow/*_test.go | grep -c '^func Test' @ 69463842 -->
**Commit:** `feat(cli): record and list self-review findings`
**After:** Task 2
**Build:** green

**Decision:** self-review-findings-table

**Decision:** store-write-never-blocks

  - [x] **Step 1: Write the failing tests** following `selfreview_test.go`'s bundle tests:
    `finding` with every flag against a fake server → exit 0 and the posted JSON; missing
    `-change`, `-angle`, `-disposition` or `-note` → exit 2 naming the flag; `-blast-radius`
    negative → exit 2; unreachable `-addr` → exit 0, one stderr line starting `flow: warning:`
    (match the record family's existing warning prefix); `findings` → prints the JSON array;
    unreachable → non-zero, nothing on stdout. Run: `cd stats && go test ./cmd/flow -run
    'TestSelfReviewFinding' -count=1` — fails.
  - [x] **Step 2: Subcommands** — `finding` (`-change`, `-angle`, `-disposition`, `-note`
    required; `-ref`, `-blast-radius` optional, the latter unset → nil) and `findings`
    (`-change`), both on `registerRecordConnFlags` and `fallback.ProjectKey(f.dir)` as `bundle`
    does. A store validation refusal (400) is exit 2 with the store's message — a caller
    mistake, not a store miss. Extend `selfReviewUsage` with both, stating the write's
    never-block contract and the read's non-zero miss in the bundle paragraph's own style.
  - [x] **Step 3: Verify** — the Step 1 command passes; `go build ./cmd/flow`; gofmt/vet.
  - [x] **Step 4: Commit.**

Correction (2026-10-04): the never-block line is `⚠ flow: self-review finding not recorded — …`,
in the record family's `⚠ flow:` style (`journalRecordWrite`), not the `flow: warning:` the plan
guessed; it covers a reached store that failed the write (a 5xx) as well as an unreachable one. The two
new verbs resolve `FLOW_RECORDS_ADDR`, so `.flow/project.md`'s record-family sentence now names
them, and `scripts/test-flow-addr-declaration.sh`'s drops-verb mutation drops all three
`flow self-review` verbs — dropping `bundle` alone left `selfreview.go` declared and the case passed.
`bundle` now shares `parseSelfReviewFlags` with the two new verbs.

- [x] 4. Report guard: accept `fixed: <sha>`

**Files:** `stats/internal/guard/selfreviewreport.go`, `stats/internal/guard/check_self_review_report_test.go`, `scripts/check-self-review-report.sh`
**Tests:** Case 34, Case 35, Case 36 — in `TestCheckSelfReviewReport`
**Regression:** case 34 fails if `— fixed: abc1234` is a violation; case 35 fails if
`— fixed:` with no sha passes; case 36 fails if `— fixed: KAN-9` (not a sha) passes.
**Baseline:** before=149 after=149
<!-- measured: cat stats/internal/guard/*_test.go | grep -c '^func Test' @ 69463842 -->
**Commit:** `feat(guard): accept the fixed disposition in self-review reports`
**After:** none
**Build:** green

**Decision:** fixed-disposition

  - [x] **Step 1: Write the failing cases** 34–36 in `TestCheckSelfReviewReport`, built like
    cases 5 and 6 (`csrSec` on the cost section). Run: `cd stats && go test ./internal/guard -run
    TestCheckSelfReviewReport -count=1` — case 34 fails.
  - [x] **Step 2: Parse** — beside `srrFiled`, a `srrFixed` (`^fixed:[[:space:]]*(.*)$`) and a
    `srrSha` (`^[0-9a-f]{7,40}$`); a `fixed:` disposition with an empty sha is "marked fixed with
    no sha", a non-sha is "marked fixed with a malformed sha `<x>`"; the "neither" and
    "malformed line" messages name all three dispositions. Update the doc comments and the
    shim's header (`THE REPORT SHAPE`) to show the third line form.
  - [x] **Step 3: Verify** — the Step 1 command passes; `scripts/check-self-review-report.sh`
    exits 0 on the real tree; gofmt/vet.
  - [x] **Step 4: Commit.**

- [x] 5. Skill: fix every non-big finding, record each one

**Files:** `skills/flow-self-review/SKILL.md`, `skills/flow-contracts/finish-contract-run2.md`, `scripts/land-self-review-report.sh`, `skills/README.md`, `skills/flow-self-review/SKILL-rationale.md`, `skills/flow-self-review/scripts/project-get.sh`
**Tests:** none — skill prose; Tasks 1–4 carry the tests for what it calls
**Regression:** none — prose.
**Baseline:** before=0 after=0
<!-- predicted: no test files in this task -->
**Commit:** `feat(flow-self-review): fix and land every non-big finding, record each in the store`
**After:** Task 3, 4
**Build:** green

**Decision:** fix-all-non-big

**Decision:** inline-fix-one-branch-one-reviewer

**Decision:** rejected-fix-to-prompt

**Decision:** store-write-never-blocks

**Decision:** fixed-disposition

  - [x] **Step 1: Frontmatter** of `skills/flow-self-review/SKILL.md`: `description:` becomes
    `Run a change's self-review pass, inline on this session's model, from the context bundle
    \`/flow\` or \`/flow-fast\` saved; fix and land every finding that is not big, file the big
    ones, record each in the flow store, rate, write the report, delete the bundle. Standalone,
    not a pipeline stage. Use for /flow-self-review.`; `allowed-tools:` gains `Bash(flow:*)`.
  - [x] **Step 2: Intro.** Replace `canonical for the six angles, what may be filed, the
    filing-and-rating prompt and the report.` with `canonical for the six angles, what is fixed
    and what may be filed, the filing-and-rating prompt, the store record and the report.`, and
    `**The pass runs inline, in this session, on whatever model it is already on** — no subagent,
    no dispatch.` with `**The pass and its fixes run inline, in this session, on whatever model it
    is already on** — the one dispatch is the fix branch's reviewer (step 3).`
  - [x] **Step 3: Step 2's filing paragraph.** Replace `**A finding is filed only from the six
    angles, and only by the operator's choice.** A finding about the pipeline itself is offered
    under its angle. A finding about the project's own product code is offered only when it is
    Important or worse` with `**A finding is fixed or filed only from the six angles.** A finding
    about the pipeline itself is fixed in step 3 unless it is \`big\`, and offered under its angle
    when it is. A finding about the project's own product code is always \`big\`, and is offered
    only when it is Important or worse`; and `files nothing and records every finding
    \`declined\`.` with `files nothing and records every offered finding \`declined\`.`
  - [x] **Step 4: Insert step 3**, the block under **Task 5 — the new step 3** below, before
    `### 3. Explain, then ask`, and renumber the following headings: `### 4. Explain, then ask`,
    `### 5. File chosen findings`, `### 7. Write the report, delete the bundle, land it`,
    `### 8. Report`; insert the block under **Task 5 — the new step 6** below as `### 6.` before
    the report step.
  - [x] **Step 5: Explain, then ask.** After its first sentence, add `Every fixed finding is one
    line in the same body, naming its landed sha; it is not offered.` Replace `up to three
    multi-select questions of three findings each` with `up to three multi-select questions of
    three offered findings each`, and append to that paragraph `With no finding to offer, the call
    carries the rating alone.`
  - [x] **Step 6: Report step.** Replace `and its disposition (\`filed: <KEY>\` or
    \`declined\`)` with `and its disposition (\`fixed: <sha>\`, \`filed: <KEY>\` or
    \`declined\`)`. In `### 8. Report`, replace `the rating, and the Jira keys filed (or \`none\`).`
    with `the rating, the shas landed and the Jira keys filed (each \`none\` when empty).`
  - [x] **Step 7: Citations.** `skills/flow-contracts/finish-contract-run2.md` step 9: `canonical
    for the six angles, what may be filed, the` → `canonical for the six angles, what is fixed and
    what may be filed, the`. `scripts/land-self-review-report.sh` header: `SKILL.md step 5` →
    `SKILL.md step 7`. `skills/README.md`'s `/flow-self-review` row: append `Fixes and lands every
    non-big finding; files the big ones.` to its description. Then
    `grep -rn 'flow-self-review' skills rules scripts stats --include='*.md' --include='*.sh'
    --include='*.go'` and correct any other step-number citation the renumbering moved.
  - [x] **Step 8: Verify** — `grep -c '^### 3. Fix every finding that is not \`big\`$'
    skills/flow-self-review/SKILL.md` prints `1`; `grep -c '^### [0-9]\.' skills/flow-self-review/SKILL.md`
    prints `8`. Run `scripts/check-references.sh`, `scripts/check-markdown-integrity.py`,
    `scripts/check-installed-citations.sh`, `scripts/check-verbatim-moves.sh`,
    `scripts/check-dispatch-paragraphs.sh`, `scripts/check-guard-symlinks.sh`,
    `scripts/check-normative-inventory.sh`, `scripts/check-self-review-report.sh`; record every
    sentence `check-verbatim-moves.sh` flags in `verbatim-moves.txt`.
  - [x] **Step 9: Commit.**

Correction (2026-10-04): the renumbering moved a heading `skills/flow-self-review/SKILL-rationale.md`
cites (`## SKILL.md — 5. Write the report…` → `7.`), and the new step 3's `project-get.sh` call
needs the skill's own `scripts/project-get.sh` symlink, which `check-guard-symlinks.sh` rule 2
requires; both join this task's files.

### Task 5 — the new step 3

````markdown verified:authored for this change
### 3. Fix every finding that is not `big`

**Classify every finding before any work** by the blast-radius rule of **Pipeline defects found
mid-run** (`skills/flow-contracts/pipeline.md`): `big` when its fix touches 60 or more files,
reaches outside `<agents repo>`, or needs a design choice only the operator can make. A finding
about the project's own product code reaches outside `<agents repo>`, so it is always `big`. An
`In-run pipeline fix:` line naming a sha is already fixed and is not classified again.

**Every finding that is not `big` is fixed and landed without asking**, on one branch for the
whole pass. A pass with none creates no worktree and dispatches nothing.

1. **Fix, inline.** This session fixes them itself — it already holds the context a fresh fixer
   would re-load. It works in its own worktree,
   `git -C <agents repo> worktree add -b self-review-<name> <agents repo>/.worktrees/self-review-<name> origin/<default-branch>`
   — never the main checkout — and makes one commit per finding with a module scope, adding one
   test or guard that fails without the fix wherever the fix changes behaviour, and running the
   `<agents repo>/.flow/project.md` `## lint` lines its files need.
2. **Review** — one fresh dispatch on `opus`, `subagent_type: flow-high`, key
   `self-review-<name>-review-<r>`, over `git diff origin/<default-branch>...self-review-<name>`,
   its prompt naming each finding beside its commit. It is one-shot: no `SendMessage` to it once
   it returns.
3. **Fix the review's findings inline**, then step 2 again, `<r>` plus one, until a review comes
   back clean, under **Fewest operator actions** (`skills/flow-contracts/pipeline.md`). A commit
   the review judges not to fix its finding, or to make things worse, is dropped from the branch,
   and its finding is offered in step 4 as a `big` one is; so is a fix found `big` once under way.
4. **Land once** by `<agents repo>`'s `## default landing route`
   (`project-get.sh <agents repo> 'default landing route'`), without asking — merge and push as
   step 4 of **Pipeline defects found mid-run** states it, with `self-review-<name>` as the
   branch. Each fixed finding's sha is read off `<default-branch>` after the landing, never from
   the branch before its rebase.
````

### Task 5 — the new step 6

````markdown verified:authored for this change
### 6. Record every finding in the flow store

One call per finding, every disposition alike, before the report is written:

```bash
flow self-review finding -change <name> -angle <label> -disposition fixed|filed|declined [-ref <sha|KEY>] [-blast-radius <N>] -note '<the finding, one line>'
```

`-ref` is the landed sha for `fixed`, the issue key for `filed`, and omitted for `declined`;
`-blast-radius` is step 3's count, omitted for a product-code finding. A store failure is one
warning line and the pass continues — the report below is the durable record.
````

- [x] 6. Live verification: a real row through a real daemon

**Files:** none
**Tests:** none — verification task; the before/after reads it records are the check
**Regression:** none — no commit
**Baseline:** before=0 after=0
<!-- predicted: no test is added by this task -->
**After:** Task 1, 2, 3, 4, 5
**Build:** green

**Decision:** self-review-findings-table

**Decision:** store-write-never-blocks

  - [x] **Step 1: Bring up this worktree's own daemon** per `.flow/project.md`'s
    `## workspace isolation` (`scripts/workspace.sh`): its own `flow_<id>` database and port,
    built from this branch with `cd stats && make build` — never `flowd` on `127.0.0.1:4173`.
    Export `FLOW_ADDR` and `FLOW_RECORDS_ADDR` to that port.
  - [x] **Step 2: Before.** `flow self-review findings -change kan-875-self-review-auto-fixes-all-angles`
    → record the output (`[]` expected).
  - [x] **Step 3: Write.** One `fixed` row (ref = this branch's head short sha, blast radius 6),
    one `filed` (`KAN-875`), one `declined`; then `-disposition bogus` → exit 2 and nothing
    written.
  - [x] **Step 4: After.** The same read → exactly the three rows, in order, fields intact.
  - [x] **Step 5: Never blocks.** `flow self-review finding … -addr http://127.0.0.1:1` → exit 0,
    one warning line.
  - [x] **Step 6: Record** the before/after outputs and exit codes in `design.md` under a
    `## Live verification` heading, each tagged `measured:` with the command and the branch.
    **Failure looks like:** the after read still `[]` or missing a row, `bogus` accepted, or the
    unreachable write exiting non-zero.
  - [x] **Step 7: Tear down** the worktree daemon it started; leave `flow-postgres` up.

Correction (2026-10-04): built with `go build -o … ./cmd/flowd ./cmd/flow` rather than `make build`, whose
`web-build` prerequisite builds an SPA this change does not touch. The fresh database held no
`projects` row, so it was seeded with `flow state set` from this change's record before the writes.

This task commits nothing; its figures are committed with the change's artifacts.

---

- [x] 7. Skill: every fix-loop step is a one-shot subagent (fix round 1)

**Files:** `skills/flow-self-review/SKILL.md`
**Tests:** none — skill prose; no guard reads step 3's wording
**Regression:** none — prose.
**Baseline:** before=0 after=0
<!-- predicted: no test files in this task -->
**Commit:** `fix(flow-self-review): run every fix-loop step as a one-shot subagent`
**After:** Task 5
**Build:** green

**Decision:** one-shot-subagent-fix-loop

**Decision:** rejected-fix-to-prompt

Operator instruction at the human gate: the initial fix, every review-finding fix, every review and
every re-review run as one-shot subagents — the session fixes nothing inline.

  - [x] **Step 1: Replace the opening paragraph's inline-fix sentence** (`SKILL.md` lines 13–17,
    from `**The pass and its fixes run inline` to `resolves.`) with the block below.
  - [x] **Step 2: Replace step 3's body** from `**Every finding that is not \`big\` is fixed and
    landed without asking**` through item 3 with the block below; item 4 (**Land once**) and the
    `<agents-base>` paragraph stay as they are.
  - [x] **Step 3: Verify** `grep -n 'inline' skills/flow-self-review/SKILL.md` names no fix-loop
    step, and `scripts/check-references.sh` plus every `scripts/check-*.sh` guard the
    `## lint` section names exit clean.
    **Failure looks like:** a remaining "fixes inline", "This session fixes them itself", or the
    `opus`-session gate.

### Task 7 — the new opening sentence

````markdown verified:authored for this change
**The pass runs inline, in this session, on whatever model it is already on; its fixes never do**
— step 3 runs every fix, review and re-review as a one-shot `opus` dispatch. The model is
picked by picking the model this session runs on (`/model`) before invoking this command, not by
anything this skill itself resolves.
````

### Task 7 — the new step 3 body

````markdown verified:authored for this change
**Every finding that is not `big` is fixed and landed without asking**, on one branch for the
whole pass. A pass with none creates no worktree and dispatches nothing. **This session changes
and reviews nothing itself** (`design.md`, `one-shot-subagent-fix-loop`): each numbered step
below that does is a fresh dispatch on `opus`, `subagent_type: flow-high`, its prompt carrying the
paragraphs **Every dispatch in the loop is one-shot** (**Pipeline defects found mid-run**,
`skills/flow-contracts/pipeline.md`) names — a review's READ-ONLY REVIEW too — with **The
handshake** there applying to each reply. Every dispatch is one-shot: no `SendMessage` to it once
it returns.

`<agents-base>` is `<agents repo>`'s own default branch —
`git -C <agents repo> symbolic-ref --short refs/remotes/origin/HEAD` with its `origin/` prefix
dropped — never the project's `<default-branch>`, which step 1 binds and which `<agents repo>` may
not carry.

1. **Fix** — this session creates the worktree,
   `git -C <agents repo> worktree add -b self-review-<name> <agents repo>/.worktrees/self-review-<name> origin/<agents-base>`
   — never the main checkout — then dispatches key `self-review-<name>-fix`, its prompt carrying
   every non-`big` finding with its evidence and counted blast radius. The fixer works in that
   worktree and makes one commit per finding with a module scope, adding one test or guard that
   fails without the fix wherever the fix changes behaviour, and running the
   `<agents repo>/.flow/project.md` `## lint` lines its files need. It reports each finding's
   commit, or a finding it found `big` once under way, left uncommitted. It never merges or
   pushes.
2. **Review** — key `self-review-<name>-review-<r>`, over
   `git diff origin/<agents-base>...self-review-<name>`, its prompt naming each finding beside its
   commit.
3. **Fix the review's findings** — key `self-review-<name>-fix-<r>`, its prompt carrying the
   review's report verbatim. Each fix is folded into the commit of the finding it fixes
   (`git commit --fixup <that commit>`, then `git rebase --autosquash origin/<agents-base>`), so
   every finding stays one commit; a commit the review judges not to fix its finding, or to make
   things worse, is dropped from the branch. Then step 2 again, `<r>` plus one, until a review
   comes back clean, under **Fewest operator actions** (`skills/flow-contracts/pipeline.md`). A
   dropped commit's finding is offered in step 4 as a `big` one is; so is a fix found `big` once
   under way.
````

Correction (2026-10-04): the opening block's last sentence keeps the original "The model is picked …"
wording verbatim rather than "The pass's model is picked …" — the preceding sentence already scopes it
to the pass, and the unchanged sentence needs no verbatim-moves entry.
