# kan-797-flow-fix-default-landing-route-never-resolves

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

Two independent tasks: the rule stated in prose (contract plus the two phase files that restate
it), and the one Go guard whose verdict must agree. Neither depends on the other's files.
`design.md` is canonical for every decision.

**Measured at plan time:**

- A prose-bearing body resolves absent today: `project-get.sh` over a temp project whose body is
  `merge and push` plus an explanatory line returns the whole prose-bearing body, which fails the
  byte-for-byte match against `merge and push`.
  <!-- measured: mktemp project + scripts/project-get.sh + literal compare @ branch spectre/kan-797-flow-fix-default-landing-route-never-resolves -->
- The documented-default shape is the literal alone on the body's first non-blank line, prose on
  following lines: head extraction then yields `fable` / `merge and push` from such bodies. A
  prose tail on the SAME line stays part of the head and is a malformed row — reported by name.
  <!-- measured: mktemp project + project-get.sh + sed '/^[[:space:]]*$/d;q' + tr -d '`' | xargs @ branch spectre/kan-797-flow-fix-default-landing-route-never-resolves -->
- `check-contract-budget.sh` exits 0 at base; the `project-configuration.md` row is 61302 bytes
  against 51700 on disk — 9602 bytes of headroom, so the reword needs no budget raise unless the
  guard actually fails.
  <!-- measured: scripts/check-contract-budget.sh; grep budgets row; wc -c @ branch spectre/kan-797-flow-fix-default-landing-route-never-resolves -->
- `check_model_keys_test.go` holds exactly one `func Test` declaration
  (`TestCheckModelKeys`); new cases are table cases and `t.Run` subtests inside it, and its
  parity half compares the Go port against the bash at `d71a2327` — the sanctioned pattern for a
  deliberate divergence is explicit subtests naming the change, as KAN-823 did at
  `prepare_archive_branch_test.go:568`.
  <!-- measured: grep -c '^func Test' stats/internal/guard/check_model_keys_test.go; sed -n '568p' prepare_archive_branch_test.go @ branch spectre/kan-797-flow-fix-default-landing-route-never-resolves -->

**No live-verification task:** the change touches no running service and no persistent state —
`flowd` and its store are untouched, and the resolution surface (config-file reading) is
exercised by task 2's guard tests and task 1's own verified snippets.

---

- [ ] 1. State the head-of-body rule in the prose that carries the match

`skills/flow-contracts/project-configuration.md` is canonical: reword its match paragraph
(`## default landing route`'s byte-for-byte rule and the three siblings it binds) so that, for
`## default landing route`, `## self review model`, `## self review` and `## handoff`, the value
is the body's **first non-blank line** — whitespace-trimmed, surrounding backticks removed —
matched byte-for-byte against that key's vocabulary, with no case-folding and no synonym list;
lines below the head are documentation for the reader, never read; a head matching no literal
stays a malformed row, reported by name (quoting what was found) and dropped, resolving as if
the key were absent. Reword the same four table rows out of "holds that value and nothing else,
never free-form prose" into the head shape. `## jira` is untouched.

Reword the two phase files that restate the whole-body rule so they cannot direct a session back
into the recorded failure: `skills/flow/integrate.md`'s landing-route resolution and
`skills/flow/archive.md`'s `## self review` match sentence each take the head-first wording,
citing **Project configuration** as canonical. Extend `skills/flow/archive.md`'s
`## self review model` shell snippet with the head extraction ahead of its backtick/xargs
normalization, exactly as verified below — with a prose-bearing body the current snippet's
`tr -d '`' | xargs` flattens the whole body into one non-member string and silently falls back
to the store default, where the contract now says the head `fable` wins.

```bash verified:run at plan time against a temp project (see Measured above)
BODY="$(project-get.sh "$MAIN_CHECKOUT" 'self review model' 2>&1)"; rc=$?
case "$rc" in
  0) PROJECT_SRM="$(printf '%s' "$BODY" | sed '/^[[:space:]]*$/d;q' | tr -d '`' | xargs)" ;;
  1) PROJECT_SRM="" ;;
  *) echo "⛔ flow: project-get.sh exited $rc: $BODY — stop the run" >&2; exit 2 ;;
esac
```

Prose discipline for this task: capture `scripts/check-normative-inventory.sh`'s output before
the first edit and diff it after the last; every difference is either this task's deliberate
reword or a sentence restored. Never weaken a guard to pass.

**Build:** green
**Files:** `skills/flow-contracts/project-configuration.md`, `skills/flow/integrate.md`,
`skills/flow/archive.md`
**Tests:** none
**Regression:** reverting this commit restores the whole-body match in the contract and in the
two phase files — a documented default resolves absent again and the landing question is asked
on every run, the recorded kan-741/kan-577 failure mode.
**Baseline:** before=0 after=0
**Commit:** fix(flow-contracts): single-line-literal keys resolve the body's head

**Decision:** head-of-body-resolution
**Decision:** four-literal-keys-family

  - [ ] **Step 1: Reword the match paragraph and the four table rows in `skills/flow-contracts/project-configuration.md`** — capture the normative inventory first (`scripts/check-normative-inventory.sh > /tmp/norm-before.txt`).
  - [ ] **Step 2: Reword the landing-route resolution in `skills/flow/integrate.md` and the `## self review` match sentence in `skills/flow/archive.md`** to the same head-first wording.
  - [ ] **Step 3: Extend `skills/flow/archive.md`'s `## self review model` shell snippet** with the head-extraction pipeline from the verified block above.
  - [ ] **Step 4: Diff the normative inventory** (`scripts/check-normative-inventory.sh > /tmp/norm-after.txt; diff /tmp/norm-before.txt /tmp/norm-after.txt`) — every hunk is this task's deliberate reword or a restored sentence.
  - [ ] **Step 5: Run the task's lint set** — `scripts/check-contract-budget.sh`, `scripts/check-references.sh`, `scripts/check-vocabulary.sh`, `scripts/check-markdown-integrity.py`, `scripts/check-dispatch-paragraphs.sh` — and fix any hit by editing the offending line.
  - [ ] **Step 6: Verify the rule reads correctly**: resolve a prose-bearing `## default landing route` body through the new sentence (the verified snippet's shape) and confirm the value is `merge and push`.
  - [ ] **Step 7: Commit** — `git add skills/flow-contracts/project-configuration.md skills/flow/integrate.md skills/flow/archive.md && git commit -m "fix(flow-contracts): single-line-literal keys resolve the body's head"`

- [ ] 2. check-model-keys reads the literal key's head

`stats/internal/guard/modelkeys.go`'s `## self review model` check stops failing a multi-line
body whose head is a member: take the head — the first line of `mkSectionBody`'s output that is
non-blank after its per-line trim and backtick strip — and apply the member compare to the head
alone, dropping the `strings.Contains(body, "\n")` violation branch. Rewrite the now-false
comment ("A multi-line body already fails: the contract is a single-line literal.") to state the
head rule. Update `scripts/check-model-keys.sh`'s shim header, whose contract sentence still
describes whole-body matching.

```go unverified:confirm against mkSectionBody's per-line processing order when implementing
head := ""
for _, line := range strings.Split(body, "\n") {
    if line != "" {
        head = line
        break
    }
}
if head != "" && !mkMember(valid, head) {
    fmt.Fprintf(stdout, "%s: `## %s` value %s is not a ValidModels member\n", pf, key, smcQuote(head, utf8))
    violations++
}
```

Test changes in `stats/internal/guard/check_model_keys_test.go`, one case at a time, TDD:

- a body carrying prose below a valid head passes (`" + fable + "` shaped like the `key` helper,
  followed by an explanation line) — the shape verified at plan time;
- a body carrying prose below an invalid head fails, naming the head (`bogus-model`), not the
  whole body;
- the parity bodies whose value the head rule changes — at minimum `` "`a`\n\n`b`" `` (the
  violation now names `a`) and `` "``\n`alpha`" `` (now passes) — leave the shared `d71a2327`
  sweep and become explicit divergence subtests pinned to the new expected outputs, named as
  diverging for KAN-797, per the KAN-823 precedent at
  `stats/internal/guard/prepare_archive_branch_test.go:568`. Confirm the exact set by running
  the sweep; any body whose output still matches the bash stays in it.

**Build:** green
**Files:** `stats/internal/guard/modelkeys.go`, `stats/internal/guard/check_model_keys_test.go`,
`scripts/check-model-keys.sh`
**Tests:** `TestCheckModelKeys`
**Regression:** reverting this commit makes the guard fail a prose-bearing `## self review
model` body the old way — a documented body the contract calls legal is a hard lint failure, and
the guard disagrees with the contract task 1 lands.
**Baseline:** before=1 after=1
<!-- measured: grep -c '^func Test' stats/internal/guard/check_model_keys_test.go @ branch spectre/kan-797-flow-fix-default-landing-route-never-resolves -->
**Commit:** fix(guard): model-keys reads the literal key's head

**Decision:** guard-follows-head-rule

  - [ ] **Step 1: Write the failing cases** in `check_model_keys_test.go`'s table — prose below a valid head passes; prose below an invalid head fails naming the head.
  - [ ] **Step 2: Run them and confirm they fail** — `cd stats && go test ./internal/guard -run TestCheckModelKeys -count=1`.
  - [ ] **Step 3: Implement the head extraction in `modelkeys.go`** per the block above, dropping the multi-line violation branch and rewriting its comment.
  - [ ] **Step 4: Run the full `TestCheckModelKeys`** — move each parity body the head rule changed into an explicit KAN-797 divergence subtest pinned to the new expected output; leave agreeing bodies in the sweep.
  - [ ] **Step 5: Reword `scripts/check-model-keys.sh`'s shim header** to the head rule.
  - [ ] **Step 6: Auto-fix then check** — `cd stats && gofmt -w . && go vet ./internal/guard/ && gofmt -l internal/guard/ && go test ./internal/guard -run TestCheckModelKeys -count=1` — all clean.
  - [ ] **Step 7: Commit** — `git add stats/internal/guard/modelkeys.go stats/internal/guard/check_model_keys_test.go scripts/check-model-keys.sh && git commit -m "fix(guard): model-keys reads the literal key's head"`
