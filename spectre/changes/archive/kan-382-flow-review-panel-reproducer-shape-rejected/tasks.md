# kan-382-flow-review-panel-reproducer-shape-rejected

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.

> **Relocation:** no

- [x] 1. Refuse a malformed panel reproducer at `flow record finding` write time

**Build:** green — `cmd/flow` compiles and its suite passes on the untouched branch
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`
**Tests:** `TestValidateFindingReproducerRejectsMetacharacters`, `TestValidateFindingReproducerRejectsLeadingDashPathToken`, `TestValidateFindingReproducerRejectsURL`, `TestValidateFindingReproducerRejectsAbsolutePathToken`, `TestValidateFindingReproducerRejectsDotDotSegment`, `TestValidateFindingReproducerRejectsHyphenInsteadOfEmDash`, `TestValidateFindingReproducerRejectsSpaceIndentedExemptionReason`, `TestValidateFindingReproducerAcceptsPlainPathWithArguments`, `TestReproducerMetacharSetSyncsWithBashSource`
**Regression:** reverted, each new refusal test fails and `flow record finding` lands the malformed reproducer in the store again — the read-time guard then rejects it after the store write, the defect KAN-382 records
**Baseline:** before=4 after=13
<!-- measured: go test ./cmd/flow -run 'Reproducer' -count=1 -v @ branch spectre/kan-382-flow-review-panel-reproducer-shape-rejected (before task 1: 4 pass) -->
<!-- measured: the same selector after task 1: 13 pass -->

**Commit:** fix(stats): validate panel reproducer shape at record time

  - [x] **Step 1: mirror the guard's lexical shape rules in `validateFindingReproducer`** — in
    `stats/cmd/flow/record.go`, keep the empty and `none — <reason>` checks, then for any other
    value refuse, each with its own message naming the broken shape: a shell metacharacter
    anywhere in the text (the `REPRODUCER_METACHARS` set, carried as a Go const documented as
    mirrored from `scripts/reproducer-metachars.sh`); a leading `-` on the first
    whitespace-separated token; `://` anywhere; an absolute token; a `..` path segment on any
    token. Rewrite the doc comment, which today says shape "stays a guard-side check".

  - [x] **Step 2: unit-test every rule and pin the metachar set to its bash source** — refusals:
    metacharacters including both quote characters and the backtick, a leading-dash path token, a
    `://` URL, an absolute token, a `..` segment, and `none - reason` (hyphen, not em dash);
    acceptances: a bare relative script path with plain arguments and `none — <reason>`. The sync
    test parses `REPRODUCER_METACHARS` out of `scripts/reproducer-metachars.sh` (resolved relative
    to the package directory) and asserts equality with the Go const.

  - [x] **Step 3: verify the task**

```bash verified:selector measured during planning in this worktree
cd stats && go test ./cmd/flow -run 'Reproducer' -count=1
```

```bash verified:commands declared in .flow/project.md ## lint
cd stats && gofmt -l . && go vet ./...
```

- [x] 2. State the accepted reproducer shape verbatim in the slot brief

**Build:** green — documentation only, nothing compiles differently
**Files:** `skills/flow/review-panel.md`
**Tests:** **none**
**Regression:** reverted, every slot's dispatch prompt again omits the accepted shape, so slots keep recording shell command lines and mangled exemption forms that write-time validation then bounces
**Baseline:** before=0 after=0
<!-- predicted: task 2 adds no tests, so the counts hold trivially -->

**Commit:** docs(flow): state the accepted reproducer shape in the slot brief

  - [x] **Step 1: extend the "Every slot must supply, per finding, a reproducer" paragraph** —
    after the existing script-not-abandoned sentence, state the accepted form verbatim: the bare
    relative path of the slot's reproducer script (`.superpowers/sdd/reproducers/<round>-<id>-<n>.sh`),
    optionally followed by plain arguments — no shell metacharacter, no leading `-` on the path
    token, no `://`, no absolute path, no `..` segment — or exactly `none — <reason>` (em dash);
    and that `flow record finding` refuses anything else before the store is contacted, bouncing
    the finding back to the raising slot.

  - [x] **Step 2: verify the task**

```bash verified:commands declared in .flow/project.md ## lint, all four scan skills/ prose
scripts/check-vocabulary.sh
scripts/check-references.sh
scripts/check-dispatch-paragraphs.sh
scripts/check-contract-budget.sh
```
