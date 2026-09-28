# Self-review context bundle for kan-797-flow-fix-default-landing-route-never-resolves

found: 6 of 7 sources; skipped: 1 of 7 sources
skipped: change summary (absent)

## .superpowers/sdd/ledgers/kan-797-flow-fix-default-landing-route-never-resolves.md

# SDD ledger — kan-797-flow-fix-default-landing-route-never-resolves

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — implementer

- Task: 1
- Role: implementer
- Key: task-1-implementer
- Model: glm-5.3-flash effort=high
- Commit: bf90dfbd
- Outcome: completed
- Started: 2026-09-27T22:51:25Z
- Tokens: cost unattributed — session never bound

## Dispatch 2 — implementer

- Task: 2
- Role: implementer
- Key: task-2-implementer
- Model: glm-5.3-flash effort=high
- Commit: 6ede4c1ac8313ddb9483ac842a38f1fafb39938f
- Outcome: completed
- Started: 2026-09-27T22:56:04Z
- Tokens: not measured

## Dispatch 3 — reviewer

- Task: no task
- Role: reviewer
- Key: task-1+2-reviewer
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: clean
- Started: 2026-09-27T23:20:05Z
- Tokens: not measured

## Dispatch 4 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: clean
- Started: 2026-09-27T23:41:35Z
- Tokens: not measured

## Dispatch 5 — verifier

- Task: no task
- Role: verifier
- Key: verify
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-28T00:05:52Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-797-flow-fix-default-landing-route-never-resolves-panel.md

# Review panel — kan-797-flow-fix-default-landing-route-never-resolves

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary+principles | minor | skills/flow-contracts/project-configuration.md:91 | the match paragraph says "exactly one literal" where ## self review model's vocabulary is the ValidModels member set, not literals |   |
| F2 | primary | minor | skills/flow/archive.md:191 | tr -d strips all backticks in the head line where the contract says surrounding backticks removed |   |
| F3 | primary | minor | stats/internal/guard/check_model_keys_test.go:332 | the divergence comment says "Two bodies deliberately diverge" while three sweep bodies left the sweep and four subtests follow |   |
| F4 | primary | minor | KNOWN-BUGS.md:96 | the task-2 entry cites check_model_keys_test.go:330 but the comment sits at :332 |   |
| F5 | primary | minor | spectre/changes/kan-797-flow-fix-default-landing-route-never-resolves/tasks.md:89 | Step 5 and the Measured bullet name scripts/check-contract-budget.sh, deleted on main before the merge base, so the step cannot run as written |   |
| F6 | primary | minor | stats/internal/guard/modelkeys.go:110-116 | the port skips a head line that strips empty while the contract words and the archive snippet take the first raw non-blank line |   |
| F7 | principles | minor | skills/flow/archive.md:191 | the head rule lives in three copies (contract, port, snippet) that have already drifted on backtick scope and stripped-empty-head selection |   |

findings-total: 7
finding-status: F1 deferred wording names the wrong vocabulary class for one key
finding-status: F2 deferred no realistic body carries an interior backtick in the value line
finding-status: F3 deferred comment count stale against the moved set
finding-status: F4 deferred one-line citation correction in the same file
finding-status: F5 deferred the named guard was ported off main after the plan was written
finding-status: F6 deferred benign divergence on a no-realistic-input edge, pinned by tests
finding-status: F7 deferred three-copy rule is the recorded design; drift noted for the next change

reproducers-total: 7
finding-reproducer: F1 none — wording only
finding-reproducer: F2 none — no realistic body carries an interior backtick in the value line
finding-reproducer: F3 none — comment wording only
finding-reproducer: F4 none — stale citation
finding-reproducer: F5 none — plan text
finding-reproducer: F6 none — benign divergence, live-verified both sides
finding-reproducer: F7 none — design observation

## Pass log

### Round 0

- roster: compact — 68
- diff size 385 under cap
- docs-only: exit 1 — scripts/check-model-keys.sh
- no addition this round — the resolved list ran alone
## spectre/changes/archive/kan-797-flow-fix-default-landing-route-never-resolves/tasks.md

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

- [x] 1. State the head-of-body rule in the prose that carries the match

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

  - [x] **Step 1: Reword the match paragraph and the four table rows in `skills/flow-contracts/project-configuration.md`** — capture the normative inventory first (`scripts/check-normative-inventory.sh > /tmp/norm-before.txt`).
  - [x] **Step 2: Reword the landing-route resolution in `skills/flow/integrate.md` and the `## self review` match sentence in `skills/flow/archive.md`** to the same head-first wording.
  - [x] **Step 3: Extend `skills/flow/archive.md`'s `## self review model` shell snippet** with the head-extraction pipeline from the verified block above.
  - [x] **Step 4: Diff the normative inventory** (`scripts/check-normative-inventory.sh > /tmp/norm-after.txt; diff /tmp/norm-before.txt /tmp/norm-after.txt`) — every hunk is this task's deliberate reword or a restored sentence.
  - [x] **Step 5: Run the task's lint set** — `scripts/check-contract-budget.sh`, `scripts/check-references.sh`, `scripts/check-vocabulary.sh`, `scripts/check-markdown-integrity.py`, `scripts/check-dispatch-paragraphs.sh` — and fix any hit by editing the offending line.
  - [x] **Step 6: Verify the rule reads correctly**: resolve a prose-bearing `## default landing route` body through the new sentence (the verified snippet's shape) and confirm the value is `merge and push`.
  - [x] **Step 7: Commit** — `git add skills/flow-contracts/project-configuration.md skills/flow/integrate.md skills/flow/archive.md && git commit -m "fix(flow-contracts): single-line-literal keys resolve the body's head"`

- [x] 2. check-model-keys reads the literal key's head

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

  - [x] **Step 1: Write the failing cases** in `check_model_keys_test.go`'s table — prose below a valid head passes; prose below an invalid head fails naming the head.
  - [x] **Step 2: Run them and confirm they fail** — `cd stats && go test ./internal/guard -run TestCheckModelKeys -count=1`.
  - [x] **Step 3: Implement the head extraction in `modelkeys.go`** per the block above, dropping the multi-line violation branch and rewriting its comment.
  - [x] **Step 4: Run the full `TestCheckModelKeys`** — move each parity body the head rule changed into an explicit KAN-797 divergence subtest pinned to the new expected output; leave agreeing bodies in the sweep.
  - [x] **Step 5: Reword `scripts/check-model-keys.sh`'s shim header** to the head rule.
  - [x] **Step 6: Auto-fix then check** — `cd stats && gofmt -w . && go vet ./internal/guard/ && gofmt -l internal/guard/ && go test ./internal/guard -run TestCheckModelKeys -count=1` — all clean.
  - [x] **Step 7: Commit** — `git add stats/internal/guard/modelkeys.go stats/internal/guard/check_model_keys_test.go scripts/check-model-keys.sh && git commit -m "fix(guard): model-keys reads the literal key's head"`
## spectre/changes/archive/kan-797-flow-fix-default-landing-route-never-resolves/design.md

# kan-797-flow-fix-default-landing-route-never-resolves — design

## Context

`.flow/project.md`'s single-line-literal keys are read by every `/flow*` phase through
`project-get.sh` (the shared `project_section` extractor) and matched against their vocabularies
per `project-configuration.md`'s match paragraph. The extractor must stay generic — `## lint`,
`## worktree setup` and friends are genuinely multi-line — so the fix lives in the **match rule**,
not the extraction. The one Go consumer in the family is `check-model-keys`
(`stats/internal/guard/modelkeys.go`), which validates `## self review model` against
`ValidModels` via its `mkSectionBody` twin. Full problem statement and the reproduced evidence:
`proposal.md` `## Why`. The approved design this file adapts:
`.superpowers/sdd/2026-09-28-kan-797-flow-fix-default-landing-route-never-resolves-design.md`
(gitignored, never committed).

## Resolution rule

For `## default landing route`, `## self review model`, `## self review` and `## handoff`:

- The value is the body's **first non-blank line**, whitespace-trimmed, surrounding backticks
  removed, matched byte-for-byte against that key's vocabulary. No case-folding, no synonym list.
- Lines below the head are documentation for the reader, never read — the `## workspace
  isolation` "prose beside it is for the reader" pattern.
- A head matching no literal stays a malformed row: reported by name (quoting what was found),
  dropped, resolving as if the key were absent.
- The four table rows drop their "holds that value and nothing else, never free-form prose"
  wording for the head shape.
- `## jira` is untouched — its body legitimately holds multiple keys and is JQL-constrained.

## Guard follow

`modelkeys.go` reads the head — first non-blank line, trim, backtick-strip — before the
`ValidModels` compare for `## self review model`, so a documented body the contract calls legal
passes. Its Go tests gain a prose-bearing-valid case and a malformed-head case. No new
family-wide validation: a malformed head is already reported at resolution time.

## Unchanged, deliberately

- `project_section` / `projectSection` / `project-get.sh` — generic extraction is load-bearing.
- This repo's own `.flow/project.md` — bare already; resolves under both rules.
- No head-extraction helper, no family-wide head-validation guard — wider than the finding.

## Decisions

### Resolve the value as the body's head, not guard authoring

**ID:** head-of-body-resolution
**Status:** active
**Chosen:** head-of-body rule — fixes the recorded failure mode (two runs tripped) while keeping
a documented default expressible.
**Considered:** authoring guard alone — keeps authoring strict but leaves the recorded failure
one prose line away from recurring; head rule + head-validating guard — declined as wider than
the finding (operator chose follow-only).

### One shared rule across the four literal keys

**ID:** four-literal-keys-family
**Status:** active
**Chosen:** all four byte-for-byte keys take the head rule, stated once in the shared match
paragraph.
**Considered:** landing route only — leaves the siblings carrying the same failure mode and
desyncs `check-model-keys` from the contract.

### The guard follows the rule without new validation

**ID:** guard-follows-head-rule
**Status:** active
**Chosen:** `modelkeys.go` reads the head for `## self review model` only.
**Considered:** validating the whole family's heads at lint time — declined: wider than the
finding; resolution-time reporting already covers a malformed head.

## Open questions
## spectre/changes/archive/kan-797-flow-fix-default-landing-route-never-resolves/narrative.md

# kan-797-flow-fix-default-landing-route-never-resolves — session narrative

## 2026-09-28 — creating run

- **Base moved mid-run, twice.** At load-context origin/main was 35 commits past the recorded
  base with none of this plan's six paths among them (no collision; recorded). At panel entry the
  same check found the movement overlapping `KNOWN-BUGS.md` and
  `stats/internal/guard/check_model_keys_test.go`; the operator chose **Rebase onto main now**.
  The rebase applied four commits cleanly and conflicted on `KNOWN-BUGS.md` (main appended its own
  deferred-finding entries at the same tail); the panel stage closed `stopped` per the conflict
  rule and the run handed the resolution back. The operator then directed **"resolve and finish
  the task"**; the conflict was resolved keeping both sides' entries, the rebase completed, and
  the rewritten branch went out with `--force-with-lease`. Consequence: the task commits this
  run's early records cite (`bf90dfbd`, `6ede4c1a`) are pre-rebase shas; the landed ones are
  `3c7c941f` and `ee94734c`.
- **The first full Go-suite run failed with the failing package lost** to a `tail` truncation;
  four subsequent fully-logged runs (`-race -count=1`, 20–21 packages) were green. Probable
  match: main's `KNOWN-BUGS.md` already records `TestConcurrentAppendVersusRetirePreservesEveryEntry`
  (`internal/reconcile`) failing once under load — a pre-existing flake this change did not
  introduce and could not reproduce.
- **`scripts/check-contract-budget.sh` disappeared from main mid-run** (guard ported in the 35
  commits): task 1's plan-time lint run used it successfully pre-rebase, and the rebased tree no
  longer carries it, so plan Step 5 names a command that cannot run — deferred as F5 rather than
  fixed, the plan text being a planning-path record.
- **`smcHas` is `smcInOrder`** — parts must appear in order without overlapping; a first
  assertion whose two parts shared the words "is not" never matched and cost one test iteration
  to diagnose.
- **Panel and gated-reviewer slots dispatched as `general-purpose`** — this harness registers no
  `flow-high` agent type, so the NO DELEGATION guarantee is prompt-carried here, not
  tool-allowlist-carried as the flow-<effort> family provides on Claude Code.
- **Jira:** KAN-797 transitioned To Do → In Progress at kickoff. No description sync — the
  operator added no scope beyond the issue.

## 2026-09-28 — integrate run

- Preflight RUN1, foreign-staged and drift both clean, unfinished-work gate CLEAR,
  visual-verify-dispatched OK (no UI paths touched).
- **The base moved twice during landing.** The sync rebase conflicted on `KNOWN-BUGS.md` a second
  time (main's newest 11 commits added another entry at the same tail). The in-place resolution
  initially took `--theirs` wholesale, which dropped main's kan-798 entry; repaired immediately
  after with `88c26cac` restoring the union — both sides' entries present, verified by grep. The
  hand-merged hunk forced the full `## lint` + `## test` lists again per the finish contract.
- Landing route: `merge and push` — this project's configured default, not asked.
## git log --stat

commit 519d0a3d49a7893da3872f0dfaefbed39cc12c16
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 28 03:20:00 2026 +0300

    chore(spectre): plan

 .../narrative.md                                              | 11 +++++++++++
 1 file changed, 11 insertions(+)

commit 82792badba1227053727fb12a837677e7230db81
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 28 03:26:51 2026 +0300

    chore(spectre): archive kan-797-flow-fix-default-landing-route-never-resolves

 .../design.md                                      |  0
 .../ledger.md                                      | 60 ++++++++++++++++++++++
 .../narrative.md                                   |  0
 .../panel.md                                       | 40 +++++++++++++++
 .../proposal.md                                    |  0
 .../tasks.md                                       |  0
 6 files changed, 100 insertions(+)

## Session narrative

Run 2 archived a change whose whole life was one working day: planned against base 4a278320, the
base moved twice before landing (35 commits at panel entry, 11 more at integrate), forcing one
operator-chosen rebase, one stage-stopped conflict handback, one operator-directed resolution,
and one in-place conflict resolution whose first `--theirs` cut dropped main's newest KNOWN-BUGS
entry — repaired by a follow-up commit rather than a redo. The work itself needed no fix round:
both tasks passed their guards first commit, the gated review and the panel were clean
Minor-only, and all ten deferrals are wording or staleness nits. The lesson the run keeps
teaching: KNOWN-BUGS.md's tail is contested ground on this repository, and every rebase onto a
moved main will meet it.
