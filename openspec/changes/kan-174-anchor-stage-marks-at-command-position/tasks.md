> **Execution:** `/myflow-do` implements this plan.

**Goal:** Make stage marks bind again. KAN-172's matcher rejects every shape the skills actually
emit, so no stage run has ever bound in normal use.

## Global Constraints

- **A false negative is worse than a false positive here.** A mark that is not recognised binds
  nothing, silently. Any tightening must be justified against that.
- **The correlator mechanism is not being redesigned** — only what counts as a mark, and when the
  state gate reads.
- **No task edits `openspec/` or `docs/superpowers/`.**

## Baseline

**Measured 2026-08-15 against `c905998`:** 334 top-level Go tests
(`cd stats && go test ./... -count=1 -v | grep -c '^--- PASS'`), 112 SPA tests. All 14 stage runs
from KAN-172's own finish sequence are unbound in the live store.

<!-- measured: both suites run, and the store queried, on 2026-08-15 -->

---

### 1 Recognise a mark by its invocation, not its position

**Build:** green

**Files:**
- Modify: `stats/internal/harvest/watcher.go`
- Modify: `stats/internal/harvest/watcher_test.go`

**Interfaces:**
- Consumes: KAN-172's `isSessionMarkCommand` and its `-session-token` value check.
- Produces: a matcher that recognises the shapes actually emitted.

- [x] **Step 1: Remove the position anchor**

Delete `commandBeginsWithMyflow` and its call. A command is a mark when it contains `stage begin` or
`stage end` **and** binds the correlator as `-session-token`'s value.

- [x] **Step 2: A corpus that resembles reality**

Replace the shape tests with cases covering, at minimum: a bare invocation; one behind `cd … &&`;
one after variable assignments on the same line; one on a later line of a multi-line block; one with
the token quoted; one with `-session-token=`. **Write the multi-line and variable-assignment cases
first and watch them fail** — they are the defect.

Keep every existing negative case: a `grep`, a `psql`, a piped `cat`, a bare mention. Those still
must not match.

- [x] **Step 3: State the residual where the code is**

The matcher's doc comment states plainly that a command reproducing a mark's text without performing
it is admitted, and why that is preferred: a false negative is silent and total, a false positive is
narrow and already refused when ambiguous. **Do not repeat KAN-172's error of justifying a gap with a
claim nobody checked.**

**Tests:** `watcher_test.go`. Verification is `cd stats && gofmt -l . && go vet ./... && go test ./...
-race -count=1`, **plus a live check**: emit a mark from inside a multi-statement shell block and
confirm its stage run binds.

**Regression:** Reverting this returns the matcher to rejecting every real mark.

**Baseline:** before=334 after=334+ Go top-level tests.
<!-- predicted: the new shape cases, none written yet -->

**Commit:** `fix(1): recognise a stage mark wherever it sits in the command`

---

### 1b A token written as a shell variable binds nothing, and nothing catches it

**Build:** green

**Found while verifying task 1**, by the implementer hitting it and by the dispatcher then finding it
in its own marks.

**Files:**
- Modify: `stats/cmd/myflow/stage.go`
- Modify: `stats/cmd/myflow/stage_test.go`

**Interfaces:**
- Consumes: the existing shell-substitution rejection, which already covers `$(…)`, backticks and
  `$VAR`.
- Produces: the same rejection reaching the shape that actually occurs.

**The defect.** A mark written `-session-token $T` is recorded in the transcript as the literal text
`$T`, because the transcript stores the command **as typed**, before the shell expands it. The mark
then binds nothing — silently, and while looking entirely correct at the call site.

**This is not hypothetical and not rare.** Of the 17 mark invocations this session emitted by hand,
**13 used `-session-token $T`** and only 3 carried a literal. Every one of those 13 was unbindable
regardless of the matcher.

<!-- measured: counted from the session transcript on 2026-08-15 -- 13 occurrences of `-session-token $T`, 2 of `-session-token $TOKEN`, against 3 literal occurrences of the run's actual token -->

**Why the existing defences miss it.** `stats/cmd/myflow/stage.go` already rejects a token containing
`$(`, a backtick or `$VAR` — but the CLI receives the value **after** the shell has expanded it, so
it sees a perfectly ordinary literal and accepts it. `scripts/check-stage-mark-calls.sh` catches the
shape, but only in skill **source**; an agent composing a mark by hand never passes through it.
**The rule is enforced everywhere except where it is actually broken.**

- [x] **Step 1: Reject at the only place that can see it**

**Outcome: not possible, and that is the finding.** The shell expands `-session-token $T` before
`os.Args` exists; nothing in argv, the environment or the process tree carries the pre-expansion
text. Any CLI-side rejection would have to reject ordinary literals too. **The guard already exists
elsewhere**: `resolveSessionTokens`' bounded give-up logs `session token unresolved after the bounded
window, giving up`, which is the observable signal for a token that never reaches a transcript,
whatever the cause. Shipped as documentation plus a test pinning the acceptance as deliberate.

The CLI cannot see the pre-expansion text — but it can see its own `argv`, and a value that arrived
via expansion is indistinguishable from a literal there. So the check must be on something else the
CLI *can* observe. Establish what that is before designing: read how the invocation reaches
`stage.go`, and determine whether the pre-expansion form is recoverable at all.

**If it is not recoverable, say so and stop** rather than shipping a check that cannot work. A
documented limitation with a guard elsewhere beats a defence that looks present and is not — that is
this change's own recurring theme.

- [x] **Step 2: Wherever the check lands, prove it against the real shape**

The test writes a mark the way an agent actually writes one, not the way a test author would.

**Tests:** `stage_test.go`. Verification is the Go suite plus a live check that a variable-written
token is refused or reported.

**Regression:** Reverting this returns to a mark that looks correct, exits 0, and binds nothing.

**Baseline:** before=335 after=335+ Go top-level tests.
<!-- predicted: the new cases, none written yet -->

**Commit:** `fix(1b): refuse a session token that cannot have reached the transcript literally`

---

### 2 The state gate reads before it marks

**Build:** green

**Files:**
- Modify: `skills/myflow-fast/SKILL.md`
- Modify: `stats/cmd/myflow/state.go`
- Modify: `stats/cmd/myflow/state_test.go`

**Interfaces:**
- Consumes: the synthetic-create path `stage begin` already has.
- Produces: a gate that cannot be failed by its own mark.

- [x] **Step 1: Reorder the gate**

`/myflow-fast`'s state gate reads the change's state **before** emitting its first mark. The mark
still fires; it fires second.

- [x] **Step 2: A synthetic record is not a state**

A record whose only author is a mark's own side effect does not satisfy a state gate expecting a
state a pipeline command wrote. Reordering fixes this command; this rule fixes the class, because
any command that marks before reading hits the same wall.

**Tests:** `state_test.go` covers a synthetic-only record not satisfying the gate. The skill change
is prose; its check is that a creating run proceeds rather than emitting a wrong-state handoff.

**Regression:** Reverting this makes every `/myflow-fast` creating run refuse on a state its own mark
created.

**Baseline:** before=334+ after=334++ Go top-level tests.
<!-- predicted: the new synthetic-record cases, none written yet -->

**Commit:** `fix(2): read the change's state before the state gate marks`
