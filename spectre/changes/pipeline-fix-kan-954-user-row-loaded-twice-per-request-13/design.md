## Context

The guard applies one bound, `mvTimeout(env)` (`MUTATE_AND_VERIFY_TIMEOUT_SECONDS`, else 300s),
to every run of both passes (`stats/internal/guard/mutateandverify.go`, `mvRunBounded` via
`mvTarget.run`). A baseline run that hits it is cannot-answer (exit 4); a mutated run that hits
it is counted as killed. The two failing-under-load cases need the mutated run to time out and
the baseline not to, and set the bound to 3s — so a baseline (a bash script running `grep`)
slower than 3s under load fails them with exit 4.

Reproduced at 4963d4a6:

- Under load: the `-race` test binary, `-test.parallel 32`, beside 40 `yes >/dev/null`
  processes on a 10-core host, failed both cases with "timed out after 3s on the baseline
  pass" on 1 of 4 runs. <!-- measured: go test -c -race; 4 runs of guard.test -test.run TestMutateAndVerify under 40 yes processes @ 4963d4a6 (machine-local load, not re-runnable as-is) -->
- Deterministically: adding `sleep 4` after the mutated branch of the `project command timeout
  counts as killed` fixture's script fails the case every time with that message. <!-- measured: go test ./internal/guard/ -run 'TestMutateAndVerify/project_command_timeout_counts_as_killed' -count=1 @ 4963d4a6, FAIL -->

The same pattern sits in `mutated pass timeout counts as killed`. `baseline pass timeout cannot
answer` uses the same 3s knob but wants the baseline to time out (a 30s sleep against 3s), so
load only makes it more certain; it is not flaky.

Blast radius: the three files the fix touches (`guard.go`, `mutateandverify.go`,
`mutate_and_verify_test.go`) plus every file citing `mvRunBounded`, `mvRunHarness`,
`mvRunCommand`, `mvTarget` or `MUTATE_AND_VERIFY_TIMEOUT_SECONDS` — 6 files, under 60: not
`big`. <!-- measured: sort -u of the three files plus git grep -lF -- <name> HEAD per name @ 4963d4a6 -->

Seam proven on the fix applied at 4963d4a6: with the injection, 8 of 8 load runs passed and the
induced 4s baseline delay passes; with `expire: env.MutateExpire` replaced by `expire: nil`, both
cases fail (~31s each: the mutated `sleep 30` runs out, nothing is counted as killed). <!-- measured: same load loop and go test -run commands @ 4963d4a6 with this change's diff applied -->

## Decisions

### Inject the expiry, as the survivors bound already does

**ID:** inject-mutate-expire
**Status:** active
**Chosen:** `Env.MutateExpire <-chan time.Time`; when non-nil, `mvRunBounded` waits on it in
place of its timer. The two cases pass `ccExpireWhen(<dir>/mutated)` (the existing helper of
`check_cleanup_complete_test.go`), the marker touched only on the mutated branch of their
harness, so the bound fires once the mutated run is running and never during the baseline.
**Considered:** raising the seconds — a widened tolerance; the next slower host flakes again;
a package-level timer variable — every case runs `t.Parallel()`, so cases would race on it;
per-pass env knobs (`…_BASELINE_TIMEOUT_SECONDS`) — production configurability nobody asked
for, and a wall-clock bound still; a fifo the harness blocks on — the kill still needs a
trigger, which is the expiry anyway.

### One bound for both passes stays

**ID:** one-bound-both-passes
**Status:** active
**Chosen:** production behaviour is unchanged: one `MUTATE_AND_VERIFY_TIMEOUT_SECONDS` bound
for every run of both passes, as `scripts/mutate-and-verify.sh` and `skills/flow/review-panel.md`
state it.
**Considered:** a separate or unbounded baseline — an unbounded baseline hangs a proof on a
hanging harness, and a baseline slower than the bound would make the mutated pass's timeout a
false kill; the shared bound plus cannot-answer is the correct contract, not the defect.

### The two cases drop the seconds knob

**ID:** drop-knob-in-expiry-cases
**Status:** active
**Chosen:** the two cases set no `MUTATE_AND_VERIFY_TIMEOUT_SECONDS`, so the default 300s
applies and the expected line reads `timed out after 5m0s — counted as killed`. With the
injection removed, the mutated `sleep 30` finishes inside 300s and both cases fail every time.
**Considered:** keeping `"3"` — the message would still read 3s, but removing the injection
would fall back to a live 3s timer that passes on an idle host, so the seam could not be shown
failing on demand.

### The baseline-timeout case keeps the real timer

**ID:** baseline-case-real-timer
**Status:** active
**Chosen:** `baseline pass timeout cannot answer` is unchanged — the one case exercising the
production timer path, as `check_cleanup_complete_test.go` keeps one case on its production
timer.
**Considered:** converting it to an injected expiry too — it is not flaky (load only lengthens
its 30s sleep against a 3s bound) and would leave the real timer untested.

### First review on low

**ID:** first-review-effort
**Status:** active
**Chosen:** the first review runs at effort `low` on `opus`: a test-only seam in one guard,
three files, mirroring an existing field, with its mechanism shown failing on demand.
**Considered:** `medium` — no normative prose or contract change to read carefully.

### One fixer group

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, Task 1 alone.
**Considered:** splitting production and test hunks — the signature change does not compile
without its callers, so the hunks are one commit.

## Open questions

## Live check

None: the change is a test-only injection seam in a one-shot guard process; production leaves
`Env.MutateExpire` nil and holds no runtime state to exercise.
