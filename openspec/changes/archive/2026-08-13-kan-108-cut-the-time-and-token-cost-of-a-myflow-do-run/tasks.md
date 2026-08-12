> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

**Goal:** Stop `/myflow-do` spending a full panel re-run on every fix round in a contracts
repository, and stop it dispatching fix instructions that cannot be shown to fix anything.

**Architecture:** One new guard script plus edits to `skills/myflow-do/SKILL.md`'s section 5. The
trigger reword is a one-clause edit backed by a new requirement. The reproducer rule is three parts: a
brief on every slot dispatch, a second marker block in the panel record, and a pre-dispatch run of
each supplied command whose non-zero exit is what clears the finding for the fixer.
`check-panel-reproducers.sh` is pure Bash over anchored greps, in the shape of
`check-unfinished-work.sh`, which reads the same record and likewise never parses the findings table.

**Tech Stack:** Bash, Markdown skills and contracts, the `openspec` CLI. No runnable application in
this repository; verification is the guards declared under `## lint` and the harnesses under `## test`
in `.myflow/project.md`, plus reading the diff.

## Global Constraints

- **No new suppression markers, no guard weakening.** A lint hit is fixed by editing the offending
  line, never by narrowing a guard's scope or deleting a row from its table.
- **The new script does not join `## lint`.** It needs a worktree carrying a change in flight, exactly
  why `check-panel-diff-size.sh`, `check-finish-preflight.sh` and `check-unfinished-work.sh` are
  already excluded. It joins `## test` in Task 5.
- **`check-unfinished-work.sh` is not edited by any task.** The reproducer lines live in their own
  block precisely so that guard keeps passing unchanged; a task that finds itself editing it has
  drifted, and interleaving the two blocks is what that guard's unbroken-span rule forbids.
- **No change to the roster presets, the optional-slot trigger table's rows, the findings table, the
  `finding-status:` marker rules, the panel diff cap, or the zero-open-findings bar.** Every rule here
  changes which fix instructions get dispatched and how wide a re-run is, never what clears the gate.
- **`/myflow-fast` is not edited.** It inherits everything through the `/myflow-do` sections it cites.
  A task editing `skills/myflow-fast/SKILL.md` has drifted.
- **The em dash in `none — <reason>` is literal.** The exemption form is matched as written, so a test
  fixture writing a hyphen instead is testing a different string.

## Baseline

<!-- verified: wc -c on each path and `sed -n '/^## test/,/^## lint/p' .myflow/project.md | grep -c '^scripts/'`, working tree at the start of this change (commit 44e3faf) -->

| File | Bytes now | Budget in `budgets()` |
|------|-----------|------------------------|
| `skills/myflow-do/SKILL.md` | 46899 | 58623 |
| `skills/myflow-do/principles-reviewer-prompt.md` | 11044 | not covered by the guard |
| `skills/myflow-do/adversarial-reviewer-prompt.md` | 2599 | not covered by the guard |

`skills/myflow-do/SKILL.md` has 11724 bytes of headroom under its current row, and this plan adds
roughly 2k of prose to section 5, so **no task re-anchors the budget table** — a re-anchor with that
much headroom left would lower the row for no reason. `.myflow/project.md` declares 17 commands under
`## test` today; Task 5 makes it 18.

`scripts/check-panel-reproducers.sh` and `scripts/test-check-panel-reproducers.sh` are new and take no
budget row: the guard covers `skills/myflow-contracts/*.md`, `skills/*/SKILL.md` and
`skills/*/SKILL-rationale.md`, and nothing under `scripts/`.

---

### 1 `scripts/check-panel-reproducers.sh` — every finding declares a reproducer

**Build:** green

**Files:**
- Create: `scripts/check-panel-reproducers.sh`
- Create: `scripts/test-check-panel-reproducers.sh`

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: `check-panel-reproducers.sh <worktree>`, called by `myflow-do` §5 in Task 3.

- [x] **Step 1: Write the harness first, with its ten cases**

`scripts/test-check-panel-reproducers.sh` builds a temporary worktree-shaped directory carrying
`.superpowers/sdd/final-review-panel.md`, runs the guard, and asserts the exit code and the reported
reason. Follow `scripts/test-check-panel-diff-size.sh`'s structure — a `case_N` function per case, a
counter, and a non-zero exit when any case fails.

```bash unverified:authored in-tree for this change; the helper shape follows scripts/test-check-panel-diff-size.sh but this file has not been run yet
#!/usr/bin/env bash
# Assertion harness for check-panel-reproducers.sh.
set -uo pipefail

GUARD="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/check-panel-reproducers.sh"
FAILED=0

make_worktree() {
  # $1 = panel record body; prints the worktree path
  local wt
  wt="$(mktemp -d)"
  mkdir -p "$wt/.superpowers/sdd"
  printf '%s\n' "$1" > "$wt/.superpowers/sdd/final-review-panel.md"
  printf '%s' "$wt"
}

expect_exit() {
  # $1 = label, $2 = expected exit, $3... = command
  local label="$1" want="$2"; shift 2
  local out got
  out="$("$@" 2>&1)"; got=$?
  if [[ "$got" != "$want" ]]; then
    printf 'FAIL %s: expected exit %s, got %s\n%s\n' "$label" "$want" "$got" "$out"
    FAILED=1
  fi
}
```

- [x] **Step 2: Write the ten cases**

Each case's panel record is written literally, including the em dash in the exemption form.

```bash unverified:authored in-tree for this change; the record bodies mirror the format the delta spec states but have not been run through the guard yet
case_1_all_present() {
  local wt; wt="$(make_worktree 'findings-total: 2
finding-status: F1 fixed
finding-status: F2 fixed

reproducers-total: 2
finding-reproducer: F1 scripts/test-check-panel-reproducers.sh
finding-reproducer: F2 none — prose-only, no runnable check')"
  expect_exit 'case 1 all present' 0 "$GUARD" "$wt"
}

case_2_missing_line() {
  local wt; wt="$(make_worktree 'findings-total: 2
finding-status: F1 fixed
finding-status: F2 open

reproducers-total: 1
finding-reproducer: F1 scripts/test-check-panel-reproducers.sh')"
  expect_exit 'case 2 missing reproducer for F2' 1 "$GUARD" "$wt"
}
```

The remaining eight follow the same shape:

1. every finding has a well-formed reproducer — exit 0 *(case 1 above)*
2. a finding named in the status block has no reproducer line — exit 1, naming `F2` *(case 2 above)*
3. `reproducers-total` disagrees with the number of anchored reproducer lines — exit 1
4. two `finding-reproducer:` lines reuse `F1` — exit 1, naming the reused identifier
5. a `finding-reproducer:` line is indented, or carries an identifier with nothing after it — exit 1
6. `none — <reason>` is accepted as well-formed — exit 0
7. a record with `findings-total: 0`, no status markers, `reproducers-total: 0` and no reproducer
   lines — exit 0
8. no panel record at the expected path — exit 2
9. the argument is not a directory — exit 2
10. a reproducer line names `F9`, which the status block does not name — exit 1

- [x] **Step 3: Run the harness and watch it fail**

Run: `scripts/test-check-panel-reproducers.sh`
Expected: FAIL — the guard does not exist yet, so every case reports a non-matching exit code.

- [x] **Step 4: Write `scripts/check-panel-reproducers.sh`**

Pure Bash over anchored greps, argument-validated, with its exit-code contract in its own header. It
reads the marker blocks only, never the findings table, for the reason
`check-unfinished-work.sh`'s own header already records about that record.

```bash unverified:authored in-tree for this change; the grep -cE and comm pipelines are standard but this script has not been run yet
#!/usr/bin/env bash
# check-panel-reproducers.sh <worktree>
#
# Every finding in a panel record must declare how it was reproduced, so that
# `/myflow-do` can run that command and require it to FAIL before dispatching a
# fix instruction built on it. A finding with no runnable check declares the
# exemption form `none — <reason>` instead.
#
# Reads ONLY the two marker blocks. It never parses the findings table — not its
# header, not its column order, not where it starts or stops — for the same
# reason check-unfinished-work.sh does not.
#
# Exit codes:
#   0  every F<n> in the finding-status block has exactly one well-formed
#      finding-reproducer line, and reproducers-total equals their number
#   1  violations found; each is reported on stderr
#   2  cannot answer at all — no worktree, no record, unreadable record
set -uo pipefail

WORKTREE="${1:-}"
[[ -n "$WORKTREE" && -d "$WORKTREE" ]] || { echo "check-panel-reproducers: not a directory: ${WORKTREE:-<missing>}" >&2; exit 2; }
PANEL="$WORKTREE/.superpowers/sdd/final-review-panel.md"
[[ -r "$PANEL" ]] || { echo "check-panel-reproducers: no readable panel record at $PANEL" >&2; exit 2; }

VIOLATIONS=()
add() { VIOLATIONS+=("$1"); }

ids_of() { grep -oE "$1" "$PANEL" | grep -oE 'F[0-9]+' | sort -u; }

STATUS_IDS="$(ids_of '^finding-status: F[0-9]+ [^[:space:]]')"
REPRO_IDS="$(ids_of '^finding-reproducer: F[0-9]+ [^[:space:]]')"
```

- [x] **Step 5: Add the count, duplicate and malformed-line checks**

```bash unverified:authored in-tree for this change; mirrors check-unfinished-work.sh's counting approach, not yet run
R_ANCHORED="$(grep -cE '^finding-reproducer: F[0-9]+ [^[:space:]]' "$PANEL")"
R_NAMED="$(grep -cE 'finding-reproducer:' "$PANEL")"
(( R_NAMED == R_ANCHORED )) || add "$(( R_NAMED - R_ANCHORED )) line(s) naming finding-reproducer: that are not marker lines — a marker must begin its line, as 'finding-reproducer: F<n> <command | none — reason>'"

T_WELLFORMED="$(grep -cE '^reproducers-total: (0|[1-9][0-9]*)[[:space:]]*$' "$PANEL")"
(( T_WELLFORMED == 1 )) || add "the panel record must carry exactly one 'reproducers-total: <n>' line"
if (( T_WELLFORMED == 1 )); then
  DECLARED="$(grep -m1 -oE '^reproducers-total: [0-9]+' "$PANEL" | grep -oE '[0-9]+')"
  (( DECLARED == R_ANCHORED )) || add "the panel record declares reproducers-total: $DECLARED but carries $R_ANCHORED finding-reproducer: marker line(s)"
fi

DUPES="$(grep -oE '^finding-reproducer: F[0-9]+' "$PANEL" | grep -oE 'F[0-9]+' | sort | uniq -d | tr '\n' ' ')"
[[ -z "${DUPES// /}" ]] || add "the reproducer block reuses identifier(s) ${DUPES% } — each F<n> must have exactly one reproducer line"

MISSING="$(comm -23 <(printf '%s\n' "$STATUS_IDS") <(printf '%s\n' "$REPRO_IDS") | tr '\n' ' ')"
EXTRA="$(comm -13 <(printf '%s\n' "$STATUS_IDS") <(printf '%s\n' "$REPRO_IDS") | tr '\n' ' ')"
[[ -z "${MISSING// /}" ]] || add "finding(s) ${MISSING% } carry no reproducer — each needs a 'finding-reproducer: F<n> …' line, or the exemption form 'none — <reason>'"
[[ -z "${EXTRA// /}" ]] || add "reproducer line(s) name ${EXTRA% }, which the finding-status block does not name"

if (( ${#VIOLATIONS[@]} > 0 )); then
  printf 'check-panel-reproducers: %s\n' "${VIOLATIONS[@]}" >&2
  exit 1
fi
echo "REPRODUCERS-OK ($R_ANCHORED finding(s) declared)"
```

- [x] **Step 6: Run the harness and watch all ten cases pass**

Run: `scripts/test-check-panel-reproducers.sh`
Expected: PASS — ten cases, no failures.

- [x] **Step 7: Make both files executable and commit**

```bash verified:the chmod and git invocations are this repository's existing pattern for a new guard plus harness
chmod +x scripts/check-panel-reproducers.sh scripts/test-check-panel-reproducers.sh
git add scripts/check-panel-reproducers.sh scripts/test-check-panel-reproducers.sh
```

**Tests:** Case 1: a complete record exits 0. Case 2: a finding with no reproducer line exits 1 and
names it. Case 3: a `reproducers-total` mismatch exits 1. Case 4: a reused `F<n>` exits 1. Case 5: an
indented or identifier-only line exits 1 as a non-marker. Case 6: `none — <reason>` is well-formed and
exits 0. Case 7: a record with zero findings and `reproducers-total: 0` exits 0. Case 8: a missing
record exits 2. Case 9: a non-directory argument exits 2. Case 10: a reproducer naming a finding the
status block does not name exits 1.
**Regression:** Case 2 (missing line): dropping the set comparison would let a record omit the field
for exactly the findings a slot supplied nothing for, which is the omission the record exists to make
visible. Case 5 (non-marker lines): counting unanchored mentions would let a `finding-reproducer:` line
written inside a fenced example stand in for a real one, the substitution
`check-unfinished-work.sh`'s own anchoring already closes for `finding-status:`. Case 7 (zero
findings): rejecting a clean panel would make the guard fail every run that raised nothing. Cases 8
and 9 (exit 2): collapsing either into exit 0 would report a record nobody could read as one where
every finding declared a reproducer.
**Baseline:** before=0 after=10 cases in `scripts/test-check-panel-reproducers.sh` (new harness).
**Commit:** `feat(kan-108-cut-the-time-and-token-cost-of-a-myflow-do-run): add check-panel-reproducers.sh`

---

### 2 The auto-escalate trigger stops firing on every contract edit

**Build:** green

**Files:**
- Modify: `skills/myflow-do/SKILL.md` — the auto-escalate sentence under **Panel re-runs**
- **Allowed-collateral:** none

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: the reworded trigger set, which Task 6's sweep reads against the delta spec.

- [x] **Step 1: Reword the clause**

In the paragraph beginning `**Escalate automatically**`, replace `a public contract` with `a guard's
behaviour`. Nothing else in the sentence changes: the file-outside-the-findings clause, the ~150-line
clause, the new-Critical clause and the three-rounds clause all stand as written.

- [x] **Step 2: State why the clause was narrowed, in one sentence**

Immediately after the trigger sentence, record that a condition firing on every fix round of every
change selects nothing, and that the repair is to narrow the condition rather than to remove the
escalation — the requirement's own wording, so the file and the spec say the same thing.

Do **not** add a rationale paragraph here. The reasoning belongs in
`skills/myflow-do/SKILL-rationale.md` if it is wanted at all, and this task does not add one:
`SKILL.md` is what every `/myflow-do` run loads.

**Tests:** No automated test — skill prose. Whether a fix altered a guard's behaviour is not
mechanically decidable from a diff, which is why the design rejected guarding it; verification is the
two lint guards that cover this file plus reading the diff against the delta spec's **The
auto-escalate trigger set is stated and discriminates** scenarios.
**Regression:** Reverting this task restores a clause that fires on every fix round in this
repository, forcing Full escalation each time — the ~430k-versus-~190k gap the ticket measured on the
KAN-107 run.
**Baseline:** before=0 after=0 automated cases.
**Commit:** `fix(kan-108-cut-the-time-and-token-cost-of-a-myflow-do-run): narrow the vacuous escalation trigger`

---

### 3 A fix instruction is dispatched only after its reproducer fails

**Build:** green

**Files:**
- Modify: `skills/myflow-do/SKILL.md` — section 5's slot dispatch briefs, the panel record format, and
  the **Panel re-runs** fix-dispatch paragraph
- **Allowed-collateral:** none

**Interfaces:**
- Consumes: `scripts/check-panel-reproducers.sh <worktree>` from Task 1.
- Produces: the reproducer brief the prompt files echo in Task 4.

- [x] **Step 1: Add the reproducer requirement to the slot dispatch briefs**

Every slot the panel spawns must supply, per finding, a runnable command demonstrating the defect or
the literal form `none — <reason>`. State it once, where the panel's dispatch rules already sit, and
say explicitly that slots dispatched by `subagent_type` receive it as prompt text — the way Bugbot
already receives the mutation-testing brief — so no agent definition is edited.

- [x] **Step 2: Document the reproducer marker block**

Beside the existing marker block, document the second one, and state that it is separate because
`check-unfinished-work.sh` requires the `finding-status:` lines to occupy one unbroken span:

```text unverified:the format this change introduces; no record carries it until this task lands
reproducers-total: 2
finding-reproducer: F1 scripts/test-check-panel-reproducers.sh
finding-reproducer: F2 none — prose-only, no runnable check
```

Carry the existing warning across: do not quote the marker format inside a record, since a
validly-formatted marker written as an example reads the same as a real one.

- [x] **Step 3: Add the pre-dispatch run to the fix-dispatch paragraph**

Where section 5 unions the open findings and hands them to one fix subagent, insert the gate ahead of
that dispatch: run each supplied reproducer in the worktree, require a non-zero exit, and dispatch
only the findings whose reproducer failed plus those recorded `none — <reason>`.

```bash unverified:the invocation this task introduces; the script itself is written in Task 1
scripts/check-panel-reproducers.sh <worktree>
```

Exit 0 proceeds. Exit 1 is a record defect — the missing field is added before any fix is dispatched.
Exit 2 stops the run: a record the guard could not read is not a record in which every finding
declared a reproducer.

- [x] **Step 4: Add the bounce rule and its accounting**

A reproducer that exits 0 bounces the finding **once**, back to the slot that raised it, carrying the
passing output. A second passing reproducer stops the run at the operator handback already documented
in this section — take another round, withdraw with a reason, or stop — and never dispatches the
finding silently. State that a bounce is guard-class, citing the same accounting section 4 states for
`check-task-commit-fields.sh`: it consumes no fix-round slot and does not advance the escalation
ladder's round count, and it neither closes nor softens the finding.

- [x] **Step 5: State the not-supplied form**

Where a slot supplies nothing at all, the parent records `none — not supplied by <slot>` and dispatches
the finding unverified. Say that this is deliberate rather than a loophole: the pipeline does not
control a third-party agent's definition, and a record that says so is better than a run wedged on
output it cannot change.

**Tests:** No automated test — skill prose, as in Task 2. The mechanical half is Task 1's guard, which
this task calls. Verification is the two lint guards plus reading the diff against the delta spec's **A
fix instruction is verified by a failing reproducer before dispatch** and **A bounced finding consumes
no fix round** scenarios.
**Regression:** Reverting this task leaves `check-panel-reproducers.sh` written but never called, and
restores the dispatch path that spent three fix rounds each on two findings whose instructions could
not have worked — the second saving the ticket ranks.
**Baseline:** before=0 after=0 automated cases.
**Commit:** `feat(kan-108-cut-the-time-and-token-cost-of-a-myflow-do-run): verify a fix instruction's reproducer before dispatch`

---

### 4 The two reviewer prompt files ask for the reproducer

**Build:** green

**Files:**
- Modify: `skills/myflow-do/principles-reviewer-prompt.md`
- Modify: `skills/myflow-do/adversarial-reviewer-prompt.md`
- **Allowed-collateral:** none

**Interfaces:**
- Consumes: the brief's wording from Task 3.
- Produces: nothing later tasks read.

- [x] **Step 1: Add the requirement to each prompt's findings section**

In each file, where the reviewer is told how to report a finding, add that every finding carries either
a runnable command demonstrating the defect or `none — <reason>`, and that a command which passes
against the diff is not a reproducer. Match each file's existing voice; neither file gains a section.

The other three slots need no file edit: the primary slot is dispatched through
`superpowers:requesting-code-review`, the low-effort code review through the harness's own
`code-review` skill, and Bugbot and Security by `subagent_type` — all four carry the brief on the
dispatch prompt Task 3 wrote, which is where a slot whose definition this repository does not own has
to be told.

**Tests:** No automated test — reviewer prompt prose. Verification is the two lint guards plus reading
the diff.
**Regression:** Reverting this task leaves the two prompt files silent on the field while
`check-panel-reproducers.sh` requires it, so the principles and adversarial slots reliably produce
findings recorded `none — not supplied by <slot>` and dispatched unverified — the exemption becoming
the normal path for the two slots this repository does control.
**Baseline:** before=0 after=0 automated cases.
**Commit:** `feat(kan-108-cut-the-time-and-token-cost-of-a-myflow-do-run): ask reviewers for a reproducer per finding`

---

### 5 `.myflow/project.md` declares the new harness

**Build:** green

**Files:**
- Modify: `.myflow/project.md` — the `## test` list
- **Allowed-collateral:** none

**Interfaces:**
- Consumes: `scripts/test-check-panel-reproducers.sh` from Task 1.
- Produces: the declaration Task 6's sweep runs.

- [x] **Step 1: Add the harness to `## test`**

Append `scripts/test-check-panel-reproducers.sh` to the fenced list under `## test`, after
`scripts/test-plan-dispatch-bundles.sh`.

- [x] **Step 2: Leave `## lint` alone, and say why in the paragraph that already says it**

The paragraph naming `check-panel-diff-size.sh` and `plan-dispatch-bundles.sh` as `/myflow-do` helpers
excluded from lint gains `check-panel-reproducers.sh` for the same stated reason — it needs a change in
flight and a worktree passed in.

**Tests:** No automated test — this task edits a configuration file's declared command list.
Verification is the two lint guards plus Task 6's sweep, which runs every command the edited list
declares.
**Regression:** Reverting this task leaves the new harness undeclared, so nothing runs it and a
regression in the guard lands silently — the gap the four `/myflow-finish` helpers were given `## test`
entries to close.
**Baseline:** before=17 after=18 commands declared under `## test` in `.myflow/project.md`.

<!-- verified: sed -n '/^## test/,/^## lint/p' .myflow/project.md | grep -c '^scripts/' at the start of this change -->

**Commit:** `chore(kan-108-cut-the-time-and-token-cost-of-a-myflow-do-run): declare the new harness under ## test`

---

### 6 Full verification sweep

**Build:** green

**Files:**
- Modify: none — this task runs commands and fixes what they report
- **Allowed-collateral:** `skills/myflow-do/SKILL.md`, `skills/myflow-do/principles-reviewer-prompt.md`,
  `skills/myflow-do/adversarial-reviewer-prompt.md`, `scripts/check-panel-reproducers.sh`,
  `scripts/test-check-panel-reproducers.sh`, `.myflow/project.md`

**Interfaces:**
- Consumes: every earlier task.
- Produces: nothing.

- [x] **Step 1: Run every lint guard**

```bash verified:the six commands are the ## lint list in .myflow/project.md, read at the start of this change
scripts/check-vocabulary.sh
scripts/check-references.sh
scripts/check-plan-provenance.sh
scripts/check-task-build-green.sh
scripts/check-workspace-isolation.sh
scripts/check-contract-budget.sh
```

Expected: all six exit 0. `check-plan-provenance.sh` reports `all provenance stated`;
`check-workspace-isolation.sh` reports `ISOLATION-OK`. A hit is fixed by editing the offending line —
never by narrowing a guard or adding a suppression marker.

Note the pre-existing `openspec validate --specs --strict` failure on `dependency-versions` for a
missing `## Purpose`, tracked as KAN-135 and untouched by this change; a validation run's other
output is still read.

- [x] **Step 2: Run every test harness the edited `## test` list declares**

Expected: all 18 exit 0, `scripts/test-check-panel-reproducers.sh` among them with its ten cases.

- [x] **Step 3: Validate the change**

```bash verified:run during planning; `openspec validate --help` shows the change name as a positional argument and rejects `--change`
openspec validate --strict --type change kan-108-cut-the-time-and-token-cost-of-a-myflow-do-run
```

Expected: PASS.

**Tests:** No new test — this task runs the ones the earlier tasks declared.
**Regression:** Reverting this task removes the sweep, so a guard broken by an earlier task's edit
lands unnoticed; every fix this task makes belongs to the task that introduced the hit.
**Baseline:** before=18 after=18 commands declared under `## test` (Task 5 made it 18; this task adds
none).
**Commit:** `chore(kan-108-cut-the-time-and-token-cost-of-a-myflow-do-run): verification sweep`

---

### 7 `scripts/run-reproducer.sh` — the runner the rules describe

**Build:** green

**Added during implementation, at the operator's direction, after the review panel's fifth pass.**
Tasks 1–6 left every rule about *running* a reproducer as prose in `skills/myflow-do/SKILL.md` with no
code behind it. Eight findings across three panel passes — F17, F18, F19, F30, F34, F40, F44 and F46 —
were all defects in that prose, and each pass found another, because a rule with no implementation can
only be re-read rather than tested. F44 is the clearest case: the fix for F34 cited `ps -o sid=`, which
does not exist on Darwin (`ps: sid: keyword not found`), and nothing could have caught that except
running it. This task builds the runner so those rules are enforced instead of described.

**Files:**
- Create: `scripts/run-reproducer.sh`
- Create: `scripts/test-run-reproducer.sh`
- Modify: `skills/myflow-do/SKILL.md` — the reproducer-run paragraphs now cite the script rather than
  describing the procedure
- Modify: `.myflow/project.md` — the `## test` list gains the new harness
- **Allowed-collateral:** none

**Interfaces:**
- Consumes: the reproducer line format `scripts/check-panel-reproducers.sh` already validates.
- Produces: `run-reproducer.sh <worktree> <reproducer-command-line>`, called by `myflow-do` §5 before
  it dispatches a fix subagent.

- [x] **Step 1: Write the harness first**

`scripts/test-run-reproducer.sh`, in the shape of `scripts/test-check-panel-reproducers.sh` — `set
-euo pipefail`, a local `set +e` around each call expected to exit non-zero, and a loud failure when a
fixture cannot be built.

- [x] **Step 2: Write the cases, then watch them fail**

Run: `scripts/test-run-reproducer.sh`
Expected: FAIL — the script does not exist, so every case reports a non-matching exit code.

Cases, each asserting the exit code **and** the message:

1. a contained relative path that exits non-zero — the defect is demonstrated, exit 0 from the runner
2. a contained relative path that exits 0 — the defect is not demonstrated, distinct exit code
3. an absolute path token — refused, never executed
4. a token containing a `..` segment — refused, never executed
5. an argument containing a `..` segment — refused, argument checked as well as path
6. any of the banned characters, including both quotes — refused
7. a path that does not exist inside the worktree — refused, never executed
8. a path escaping the worktree through a symlink — refused; this is the case the record-format guard
   deliberately cannot decide, and the runner is where it is decided
9. a command that runs longer than the bound — killed, reported unverifiable, never read as either a
   pass or a fail
10. a command that detaches itself and survives the process-group kill — reported as a surviving
    process with its pid, and the worktree re-checked

- [x] **Step 3: Write `scripts/run-reproducer.sh`**

Pure Bash, `set -euo pipefail`, its exit-code contract in its own header. It resolves the worktree,
splits the command line into a path token and arguments, applies the shape and containment checks,
then executes with the worktree as the working directory using an **argument vector, never a string
through a shell**. It bounds the run using the technique `scripts/check-cleanup-complete.sh` already
uses — `timeout(1)` is absent on Darwin — and it discovers the session-kill mechanism by **running**
the candidate command rather than by citing a flag: `ps -o sid=` is invalid on this platform, which is
what F44 recorded.

- [x] **Step 4: Run the harness until every case passes**

Run: `scripts/test-run-reproducer.sh`
Expected: PASS, every case.

- [x] **Step 5: Point the skill at the script**

Replace the described procedure in `skills/myflow-do/SKILL.md` with the invocation and its exit-code
reading. **The prose gets shorter, not longer** — the constraints move into the script, and the file is
at 54843 bytes against a 58623 budget, with the operator having already chosen trimming over raising
that budget once.

- [x] **Step 6: Declare the harness and commit**

Append `scripts/test-run-reproducer.sh` to `.myflow/project.md`'s `## test` list, and add the script to
the paragraph naming the `/myflow-do` helpers excluded from `## lint` — it needs a worktree and a
command line, exactly like its siblings.

**Tests:** Case 1: a failing contained reproducer exits 0 from the runner. Case 2: a passing one is
distinguished by its own exit code. Case 3: an absolute path is refused. Case 4: a `..` path token is
refused. Case 5: a `..` argument is refused. Case 6: each banned character, both quotes included, is
refused. Case 7: a non-existent path is refused. Case 8: a symlink escape is refused. Case 9: an
over-bound command is killed and reported unverifiable. Case 10: a detached survivor is named with its
pid.
**Regression:** Case 8 (symlink escape): dropping the resolved-path check would leave the only
containment enforcement lexical, which is what `check-panel-reproducers.sh` deliberately does not
decide — the escape would then reach execution. Case 9 (bound): reading a killed command's non-zero
exit as a demonstrated defect dispatches a fix on nothing, the defect F18 recorded. Case 6 (quotes):
omitting the quote characters is F45, where an unbalanced quote reached a shell.
**Baseline:** before=0 after=10 cases in `scripts/test-run-reproducer.sh` (new harness); the `## test`
list goes from 18 commands to 19.
**Commit:** `feat(kan-108-cut-the-time-and-token-cost-of-a-myflow-do-run): add run-reproducer.sh`
