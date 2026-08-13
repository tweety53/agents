> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

**Goal:** Close the six items KAN-108 left open, and close the *class* behind each rather than the
instance — one definition of the panel-record marker helpers, one guard that reads this repository's
own Markdown for structural damage, an escalation trigger that discriminates, and a reproducer whose
survivors are found by process group rather than by parentage.

**Architecture:** One new sourced Bash library, one new Python guard with its own harness, edits to
two existing guards, one edit to `scripts/run-reproducer.sh`, two prose edits to
`skills/myflow-do/SKILL.md`, and two declarations in `.myflow/project.md`. Nothing here couples: the
tasks share no state beyond Task 4 needing Task 3's script to exist.

**Tech Stack:** Bash, Python 3 (standard library only, `/usr/bin/python3`), Markdown skills and
contracts, the `openspec` CLI. No runnable application in this repository; verification is the guards
declared under `## lint` and the harnesses under `## test` in `.myflow/project.md`, plus reading the
diff.

## Global Constraints

- **No new suppression markers, no guard weakening.** A lint hit is fixed by editing the offending
  line, never by narrowing a guard's scope or deleting a row from its table.
- **The extraction in Task 1 is behaviour-preserving.** Both guards' exit codes and reported output
  stay identical, and both existing harnesses pass **unchanged** — no harness case is edited,
  relaxed, or deleted to accommodate the library. A task that finds itself editing an existing
  assertion has drifted.
- **The two `ids_of` shapes both survive.** `check-unfinished-work.sh` needs duplicates retained;
  `check-panel-reproducers.sh` needs them removed. A library that exports one shape and adapts the
  other caller's expectations at the call site is the wrong answer — the shape is the parameter.
- **`scripts/check-markdown-integrity.py` uses the standard library only.** No pip, no third-party
  import, no network, matching `scripts/check-plan-provenance.py`'s own constraint.
- **The new guard scans `skills/**/*.md` and `rules/*.mdc` and nothing else.** Not `docs/`, not
  `openspec/`, not the repository root's Markdown. Widening the scope is a separate change.
- **No change to the roster presets, the findings table, the `finding-status:` marker rules, the
  panel diff cap, or the zero-open-findings bar.** Nothing here changes what clears the gate.
- **`skills/myflow-fast/SKILL.md` is not edited.** It inherits everything through the `/myflow-do`
  sections it cites. A task editing it has drifted.
- **A backticked token in a `**Tests:**` field is a declaration, not prose.**
  `check-task-commit-fields.py` reads `Case <n>` labels first and, failing those, every backticked
  token, and requires each to appear in the task's own commit diff. A field that backticks a harness
  path, a script another task added, or a configuration heading declares a test this task must
  produce and did not. Tasks 1 and 4 each tripped this once; both fields were rewritten to name those
  things in plain prose. Where a task genuinely adds no test, say so without backticks.

## Baseline

<!-- verified: wc -c on each path; `sed -n '/^## test/,/^## lint/p' .myflow/project.md | grep -c '^scripts/'` and the same for the lint block; `grep -cE '^expect_exit|^expect_exit_and_names|^assert_' <harness>`; working tree at commit a6d2855 -->

| Measure | Now | After this change |
|---------|-----|-------------------|
| `skills/myflow-do/SKILL.md` | 52389 bytes, budget row 58623 (6234 headroom) | still under the row |
| `scripts/check-unfinished-work.sh` | 24459 bytes | smaller — helpers move out |
| `scripts/check-panel-reproducers.sh` | 20838 bytes | smaller — helpers move out |
| `scripts/run-reproducer.sh` | 28420 bytes | larger by the group wrapper |
| `.myflow/project.md` `## test` commands | 19 | 20 |
| `.myflow/project.md` `## lint` commands | 6 | 7 |
| `scripts/test-check-unfinished-work.sh` assertions | 86 | see the harness's own output |
| `scripts/test-run-reproducer.sh` cases | 15 | see the harness's own output |

`skills/myflow-do/SKILL.md` has 6234 bytes of headroom under its budget row, and this plan adds
roughly 1k of prose to it, so **no task re-anchors the budget table**. `scripts/lib/panel-record.sh`,
`scripts/check-markdown-integrity.py` and `scripts/test-check-markdown-integrity.sh` are new and take
no budget row: `check-contract-budget.sh` covers `skills/myflow-contracts/*.md`, `skills/*/SKILL.md`
and `skills/*/SKILL-rationale.md`, and nothing under `scripts/`.

---

### 1 `scripts/lib/panel-record.sh` — one definition of the marker helpers

**Build:** green

**Files:**
- Create: `scripts/lib/panel-record.sh`
- Modify: `scripts/check-unfinished-work.sh`
- Modify: `scripts/check-panel-reproducers.sh`

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: `count_matching`, `grep_lines_of`, `ids_of` for Task 2 and for both guards.

- [x] **Step 1: Write the library**

Three exported functions, each carrying the disciplines that are today stated twice in two headers:
`-a` on every `grep` so a stray NUL byte cannot silence a match; the `rc > 1` split that separates
grep's "no match" (an answer) from a real error (a refusal); and `--` before every path.

```bash unverified:authored in-tree for this change; the bodies are lifted from the two existing guards but this file has not been run yet
# ids_of <file> <ere> <shape>
#   shape "digits"     — bare digits, sorted, DUPLICATES RETAINED. check-unfinished-work.sh's
#                        repeated_ids reads this list and finds repeats with `uniq -d`; removing
#                        duplicates here would make that check pass vacuously.
#   shape "ids-unique" — F<n>, sorted and de-duplicated, for set comparison with `=`.
ids_of() {
  local file="$1" ere="$2" shape="$3" raw
  raw="$(grep_lines_of "$file" match "$ere")" || return "$?"
  [ -n "$raw" ] || return 0
  case "$shape" in
    digits)     printf '%s' "$raw" | tr -cd '0-9\n' | sort ;;
    ids-unique) printf '%s\n' "$raw" | grep -aoE 'F[0-9]+' | sort -u ;;
    *) echo "ids_of: unknown shape '$shape'" >&2; return 2 ;;
  esac
}
```

- [x] **Step 2: Source it from both guards and delete the local copies**

Each guard resolves the library from its own location, so an invocation from any working directory
finds it, and a library that cannot be sourced is exit 2 — cannot answer — never a guard that
proceeds with fewer checks.

```bash unverified:authored in-tree for this change; not yet run
LIB="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib/panel-record.sh"
# shellcheck source=lib/panel-record.sh
. "$LIB" || { echo "check-unfinished-work: cannot source $LIB — cannot determine anything" >&2; exit 2; }
```

`check-panel-reproducers.sh`'s call sites gain the `$PANEL` argument its local copies read from a
global; its `lines_of` / `full_lines_of` wrappers stay, now forwarding to the library's
`grep_lines_of` with the file argument. `check-unfinished-work.sh`'s `ids_of` call sites gain
`digits`; `check-panel-reproducers.sh`'s gain `ids-unique`.

**Tests:** No new test — this is a behaviour-preserving refactor, and a field naming one would
declare a test this task must not add. Verification is that both existing harnesses, the
unfinished-work one at 86 assertions and the panel-reproducers one at 37 cases plus its
metacharacter loop, pass with **no edit to either**, plus a run of both guards from a different
working directory, proving the BASH_SOURCE resolution.
**Regression:** Reverting this task restores two copies of the same helpers, which is the state in
which `check-unfinished-work.sh:347` drifted from its sibling's bound in the first place — the drift
the ticket measured. The behaviour-preserving claim is pinned by the unchanged harnesses: if either
guard's answer moved, a case fails.
**Baseline:** before=86 after=86 assertions in `scripts/test-check-unfinished-work.sh`; before=37
after=37 cases in `scripts/test-check-panel-reproducers.sh`. <!-- verified: grep -cE '^expect_exit|^expect_exit_and_names|^assert_' on each harness at commit a6d2855 -->
**Commit:** `refactor(kan-153-kan-108-follow-up): extract the panel-record marker helpers into one sourced library`

---

### 2 Bound the `findings-total` digit run

**Build:** green

**Files:**
- Modify: `scripts/check-unfinished-work.sh`
- Modify: `scripts/test-check-unfinished-work.sh`
- Modify: `scripts/lib/panel-record.sh`
- Modify: `scripts/check-panel-reproducers.sh`
- Modify: `scripts/test-check-panel-reproducers.sh`

**Interfaces:**
- Consumes: Task 1's library (the guard already sources it).
- Produces: `PANEL_RECORD_TOTAL_DIGITS`, the bound's single definition.

**Why this task owns five files rather than two.** The bound is not one guard's business: the review
panel found the literal `(0|[1-9][0-9]{0,14})` copy-pasted at four call sites across both guards and
absent from the library they both source — Task 1 having centralized the marker *functions* while
leaving the *pattern* that actually drifted duplicated. Centralizing it defines the constant in the
library and rewrites all four call sites, so the task that owns the bound owns those files too. The
fix was first aimed at Task 1's commit and could not fold there: the call sites are lines this task
created, and a fixup cannot land in a commit older than the lines it edits.

- [x] **Step 1: Add the failing case first**

A record declaring `findings-total:` followed by a 26-digit run. The guard today matches it as
well-formed, where its sibling would reject it — the divergence, not a crash.

**Correction, established by running the unpatched guard rather than by reading the ticket.** The
Jira issue describes "the 26-digit crash KAN-108 fixed in its own guard" as still reachable here. It
is not: this guard compares the declared total to the marker count with a string `!=`, never
arithmetic, and a 26-digit total produces an ordinary `OUTSTANDING:` verdict at exit 0. The bound
still lands, because one record format read by two guards that disagree on which totals are
well-formed is the drift Task 1 exists to close — but the crash narrative is withdrawn.

- [x] **Step 2: Bound the pattern**

`^findings-total: (0|[1-9][0-9]*)[[:space:]]*$` becomes
`^findings-total: (0|[1-9][0-9]{0,14})[[:space:]]*$`, at both sites that carry it — the
`count_matching` call and the `grep -am1` that extracts the line — so the extraction cannot succeed
where the count refused. An over-long total now falls through to the existing "not a plain count"
report, which is the correct disposition: it is malformed, not absent.

**Tests:** One case in the unfinished-work harness: a record whose `findings-total:` carries 26
digits is reported as not a plain count. Plus one case in each guard's harness for the spec scenario
**A missing library is a refusal** — the guard copied to a sandbox without the library beside it must
exit 2 and name what it could not read. The assertion is made against this guard's own
contract — an `OUTSTANDING:` verdict at **exit 0**, exit 2 being reserved for a refusal — and not
against `check-panel-reproducers.sh`'s exit-1 contract, which is a different guard's. The 15-digit
boundary is covered by the existing well-formed cases.
**Regression:** Reverting this task restores a divergence in which two guards reading the same record
format disagree on which `findings-total` values are well-formed, so a record its sibling rejects is
accepted here.
**Baseline:** the harnesses' own final lines are the authority, not a figure restated here — see the
correction below for why this field stopped carrying one.

**Correction, twice over.** This field first declared `after=87`, counting one assertion for the new
case. That number was wrong, and a declared count is read as a constraint: the first implementation
dropped `assert_verdict` to reach it, leaving a case that would still pass if the guard regressed
from an `OUTSTANDING:` verdict at exit 0 to a refusal at exit 2. The per-task review caught it, and
the field was corrected to `88`. The panel round then added the missing-library cases and moved it
again. **A baseline figure is a measurement; this one was a guess twice, so it is no longer restated
here** — every count in this plan that a later round moved is a count that should have cited its
command instead of its value.
**Commit:** `fix(kan-153-kan-108-follow-up): bound the findings-total digit run in check-unfinished-work.sh`

---

### 3 `scripts/check-markdown-integrity.py` — read this repository's own Markdown

**Build:** green

**Files:**
- Create: `scripts/check-markdown-integrity.py`
- Create: `scripts/test-check-markdown-integrity.sh`

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: `check-markdown-integrity.py [<project-root>]`, declared in Task 4.

- [x] **Step 1: Write the harness first**

Follow `scripts/test-check-panel-reproducers.sh`'s structure — a temporary directory shaped like a
project root carrying `skills/` and `rules/`, a helper that writes one Markdown file and runs the
guard, and assertions on exit code and reported text. Twelve cases: one firing and one near-miss per
signal, plus the three environment cases (unreadable file, unknown argument, missing scope root).

- [x] **Step 2: Write the parser**

A line walker producing typed blocks — `fence`, `table`, `heading`, `list_item`, `blockquote`,
`paragraph` — in the shape `check-plan-provenance.py`'s classifier already uses. Fence state is
tracked by opening delimiter length and character, so a `~~~` block containing ``` is not misread.

- [x] **Step 3: Implement the four signals**

Each returns `(line, signal, message)` and the guard prints `file:line: signal: message`.

```python unverified:authored in-tree for this change; not yet run
# Signal 4 — dangling promise. Deliberately structural, never phrase-based: this
# repository ends hundreds of sections with a correct "See **X** (`f.md`) for why …"
# citation, and a phrase rule would fire on all of them with no legal way to quiet it.
def dangling_promise(blocks, i):
    block = blocks[i]
    if block.kind != "paragraph":
        return None
    if not block.text.rstrip().endswith(":"):
        return None
    nxt = blocks[i + 1] if i + 1 < len(blocks) else None
    if nxt is None or nxt.kind == "heading":
        return (block.last_line, "dangling promise",
                "a colon promises content, and the block ends with nothing following it")
    return None
```

- [x] **Step 4: Exit codes and the root argument**

0 clean, 1 violations found (each reported with its line), 2 cannot answer — an unreadable file, an
argument the guard does not accept, or a scope root that does not exist. An optional project root
defaults to the repository the script lives in, so it runs against a bare tree.

**Tests:** Case 1: a code span with a backslash-escaped backtick exits 1. Case 2: a code span with
double-backtick delimiters around a literal backtick exits 0. Case 3: an odd backtick count outside a
fence exits 1. Case 4: a `> ` line after un-marked prose that does not end a sentence exits 1. Case
5: a `> ` line after a completed sentence exits 0. Case 6: a paragraph ending without terminal
punctuation exits 1. Case 7: the same text inside a fence, a table cell and a list item exits 0.
Case 8: a paragraph ending in a colon with a following list exits 0. Case 9: a paragraph ending in a
colon as the last block before a heading exits 1. Case 10: a section ending with a `See **X**
(`f.md`) for why …` citation exits 0. Case 11: an unreadable file exits 2. Case 12: a scope root
that does not exist exits 2.
**Regression:** Case 1 (broken span): removing the signal restores the class F43 recorded, where half
a character list renders as plain text and no guard sees it. Cases 4 and 6 (torn prose): removing
either restores the state in which a prose-moving edit tore three passages apart and
`check-references.sh` passed throughout, because it verifies that headings resolve and not that
sentences are whole. Cases 5, 7, 8 and 10 (near-misses): each pins a correct shape that a looser rule
would flag, and deleting any one of them would let a signal be widened without a failing test —
Case 10 in particular is the one that keeps signal 4 structural rather than phrase-based. Cases 11
and 12 (exit 2): collapsing either into exit 0 would report files nobody could read as clean.
**Baseline:** the harness's own final line is the authority. Twelve cases were planned; calibration
and three review rounds added more, and every figure written here was overtaken by the next round.

**Why four more than this field first declared.** The first raw run against `skills/` and `rules/`
reported 104 hits, nearly all false positives, and each one was a defect in a rule rather than in the
text: YAML frontmatter read as a paragraph, a bold-wrapped sentence ending read as unterminated, a
lazily-continued list item read as an orphan paragraph, a stray `|` in prose read as a table row, a
code span opened on one line and closed on the next read as two broken spans, a four-space indented
code example read as prose, and a paragraph deliberately introducing a fence or list read as
unterminated. Each rule was tightened and the shape it had misread was pinned as a near-miss case.
That is the calibration this task exists to do, and the four extra cases are its record.

**And five more after that, from review.** This field twice declared a count that a later round
moved, so the arithmetic is written out rather than restated: twelve required, plus four from
calibration, plus three from the first review round (a torn paragraph before an unrelated fence, the
introducing shape that must still pass, a heading before a blockquote), plus two from the second (a
sentence resuming after an embedded block, and the same shape where it does not resume) — and the
predecessor-kind fix in the panel round adds its own. **A count in a plan is a measurement, and this
one was repeatedly a guess**; the harness's own final line is the authority.
**Commit:** `feat(kan-153-kan-108-follow-up): add check-markdown-integrity.py`

---

### 4 Declare the guard, and repair what its first run finds

**Build:** green

**Files:**
- Modify: `.myflow/project.md`
- Modify: whichever `skills/**/*.md` or `rules/*.mdc` files the guard's first run reports

**Allowed-collateral:** `skills/**/*.md`, `rules/*.mdc`

**Interfaces:**
- Consumes: Task 3's script.
- Produces: a green `## lint` run for later tasks.

- [x] **Step 1: Declare it**

`scripts/check-markdown-integrity.py` joins `## lint` — it scans the repository tree and needs no
change in flight, which is the stated test for that list. `scripts/test-check-markdown-integrity.sh`
joins `## test`. Neither list is cited by count anywhere, which is deliberate and stays that way.

- [x] **Step 2: Run it and repair every hit**

Every violation is repaired by fixing the text — closing the span, restoring the blockquote marker,
completing the sentence, supplying what the colon promised. Never by narrowing the guard's scope,
adding a suppression marker, or deleting a signal. `skills/myflow-do/SKILL-rationale.md` is the file
the ticket names as damaged; the run decides whether that damage survived KAN-108's own later passes.

A hit that is genuinely a false positive is a defect in the signal's *rule*, fixed by tightening the
rule and adding the near-miss to Task 3's harness — not by exempting the file.

**Tests:** No new automated case — this task edits a configuration file's declared command list and
whatever prose the guard reports. Verification is the Markdown-integrity guard exiting 0 against this
repository, and every other lint guard still exiting 0 afterwards.
**Regression:** Reverting this task leaves the guard written but never run, so no `/myflow-*`
verification invokes it and the classes it closes stay open in practice.
**Baseline:** before=19 after=20 commands under `## test`; before=6 after=7 under `## lint`.
**Commit:** `chore(kan-153-kan-108-follow-up): declare check-markdown-integrity.py and repair what it finds`

---

### 4b Repair F43's surviving instance in the canonical spec

**Build:** green

**Files:**
- Modify: `openspec/specs/myflow-review-panel-economics/spec.md`

**Interfaces:**
- Consumes: nothing. Produces: nothing later tasks read.

- [x] **Step 1: Repair the broken code span**

`openspec/specs/myflow-review-panel-economics/spec.md:504` carries F43's defect class: a
single-backtick span holding a backslash-escaped backtick, so the span closes early and half the
banned-character list renders as literal text. Replace the delimiters with double backticks, where a
literal backtick needs no escape. This change's own delta spec restated that line verbatim and
carries the same repair.

**Why this task exists at all, and why it is numbered 4b.** The repair was made before any task
declared it — an edit to a canonical spec file owned by no task, appearing in no commit, which the
review panel caught precisely because nothing tracked it. It is recorded here rather than left
implicit. The file is a planning path, so no task commit stages it and
`scripts/check-task-commit-fields.sh` never sees it; `/myflow-finish` run 1 commits it with the
other planning artifacts.

**Tests:** No automated test. The repository's Markdown-integrity guard scans skills and rules only,
so it does not read this path — which is the gap this change records as its open question, and the
reason the instance survived to be found by hand. Verification is reading the rendered line.
**Regression:** Reverting this task restores a line whose banned-character list renders half as
literal text, in the canonical statement of the reproducer shape rule — the exact defect F43 named,
in the file a future reader consults for that rule.
**Baseline:** before=0 after=0 automated cases.
**Commit:** none of its own — a planning path, committed by `/myflow-finish` run 1 with the other
planning artifacts.

---

### 5 `skills/myflow-do/SKILL.md` — the materiality qualifier and the size clause

**Build:** green

**Files:**
- Modify: `skills/myflow-do/SKILL.md`

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: nothing later tasks read.

- [x] **Step 1: Restore the materiality qualifier**

The passage at section 5 that reads "A fix whose diff, for a given finding, touches only the
reproducer's own target, or touches a named path with no non-comment, non-whitespace change, is not
a fix" carries only the first half of the spec's condition. Restate it with both clauses — *touches
only the reproducer's own target **and no path the finding named*** — and add the sentence naming
the ordinary case explicitly, so a literal reading no longer disqualifies a guard-script fix whose
reproducer targets the very file the finding named.

- [x] **Step 2: Make the size clause conditional**

The auto-escalate list's `the fix diff exceeds ~150 changed lines` becomes a conjunction: it
escalates only alongside a risk signal — a new file, a delta spec, a migration, a guard's behaviour,
or a file outside the set named in the findings. State the reasoning in one sentence and leave the
longer form to the delta spec, which carries it.

**Tests:** No automated test — skill prose, as in KAN-108's own trigger reword. Whether a diff
carries a risk signal is not mechanically decidable from the diff alone, which is why the design
does not guard it. Verification is every lint guard that covers this file, the contract-budget one
included, plus reading the diff against the delta spec's
**A large mechanical fix carrying no risk signal stays Targeted**, **A large fix carrying a risk
signal escalates** and **A fix to a path that is both the finding's target and the reproducer's is
material** scenarios.
**Regression:** Reverting Step 1 restores a clause that reads as disqualifying the commonest correct
fix shape in this repository, where nearly every finding names a script and nearly every reproducer
targets that script. Reverting Step 2 restores the trigger that sent two KAN-108 fix rounds to seven
slots each on diff length alone.
**Baseline:** before=0 after=0 automated cases; `skills/myflow-do/SKILL.md` stays under its 58623-byte
budget row. <!-- verified: budget row read from scripts/check-contract-budget.sh:88 at commit a6d2855 -->
**Commit:** `fix(kan-153-kan-108-follow-up): restore the materiality qualifier and make the size trigger conditional`

---

### 6 `scripts/run-reproducer.sh` — a process group of its own

**Build:** green

**Files:**
- Modify: `scripts/run-reproducer.sh`
- Modify: `scripts/test-run-reproducer.sh`

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: nothing later tasks read.

- [x] **Step 1: Add the failing case first**

A fixture that double-forks a grandchild whose intermediate parent exits immediately, so the
grandchild re-parents before the first poll. Today the run returns "defect not demonstrated" and
leaves the process alive; the case asserts exit 3 with the survivor's pid named.

- [x] **Step 2: Wrap the exec in a new process group**

Darwin ships no `setsid` binary, so the group is established by a `python3` shim that calls
`os.setsid()` and then `os.execvp`, with the argv vector passed through untouched. No shell sees the
reproducer line, so the direct-exec guarantee and every containment check upstream are unaffected.

```bash verified:run against python3 3.14.6 on this machine during implementation; the two corrections below were each established by direct experiment
exec python3 -c 'import os,sys; os.setsid(); os.execvp(sys.argv[1], sys.argv[1:])' "$RESOLVED_PATH" "${ARGS[@]}"
```

**Two corrections to the sketch above, both found by running it.** First, `python3 -c` does **not**
strip a `--` separator: with it, `sys.argv` becomes `['-c', '--', …]` and `os.execvp` tries to exec a
file literally named `--`. It is dropped, which is safe because the resolved path is always absolute
and so can never be read as a flag. Second, the exec point sat inside a `set -m` backgrounded
subshell, and `set -m` makes that subshell a process group leader — `os.setsid()` raises
`PermissionError: [Errno 1] Operation not permitted` for a caller that is already a leader. `set -m`
is removed; the background job then inherits its parent's group, is not a leader, and `os.setsid()`
succeeds. The real variables are `RESOLVED_PATH` and `ARGS[@]`, not the `ARGV` the sketch assumed.

`python3` missing at run time is a refusal — exit 2, recorded unverifiable — never a fall back to the
ungrouped exec, because an ungrouped exec is precisely the condition in which the survivor goes
unseen.

- [x] **Step 3: Detect and kill by group**

Survivor detection gains a `pgrep -g <pgid>` pass beside the existing breadth-first descendant walk.
Cleanup signals the group (`kill -- -<pgid>`) with the same SIGTERM-then-grace-then-SIGKILL sequence,
then falls back to the deepest-first walk as a backstop. Both mechanisms stay: the walk still names
pids the group pass would report only as a count.

- [x] **Step 4: Replace the documented limitation**

The header's note that a fast double fork is undetectable is now false and is replaced by what the
group covers and what it does not: a reproducer that calls `setsid` itself leaves the group
deliberately and is the remaining gap.

**Tests:** Case 16: a fast double fork whose intermediate exits before the first poll exits 3 and
names the surviving pid. Case 17: an ordinary reproducer that exits cleanly still exits 0, and no
process is left behind — the control proving the group wrapper did not change the normal path. The
existing 15 cases, including the metacharacter loop and its control, pass unchanged.
**Regression:** Reverting this task restores the parentage-only detection, under which a fast double
fork returns "defect not demonstrated" while leaving a live process behind — the finding verified
after KAN-108's own partial fix. Case 17 pins the other direction: a group wrapper that broke the
ordinary exec would make every reproducer unverifiable.
**Baseline:** before=15 after=18 cases in `scripts/test-run-reproducer.sh` — the two above plus one
the per-task review required.

**Why one more than this field first declared.** The review found the survivor snapshot correctly
ordered only on the early-exit path: on the timeout path two signals fired before it, so a process
killed there could be reaped before the snapshot ran and its pid was never named to the operator —
the same reap-race the round before had closed elsewhere. Case 18 pins the timeout path specifically.
Closing it also required narrowing the timeout escalation's SIGKILL from the whole process group to
the command's own pid, so that a same-group survivor is killed by the retry sweep that reports it
rather than pre-empted by a blind group kill that does not. <!-- verified: the harness's own final line reads "all 15 cases plus the metacharacter loop and its control pass" at commit a6d2855 -->
**Commit:** `fix(kan-153-kan-108-follow-up): run a reproducer in its own process group`
