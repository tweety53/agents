# Final review panel — kan-172-attribute-stage-runs-to-their-session

## Diff-size gate

`scripts/check-panel-diff-size.sh . 88db0a2` measured **4,377** lines against a **2,000** cap and
exited 1. Put to the operator with the alternatives of splitting the change or reviewing only the
riskiest subset; **the operator chose to proceed with the whole diff.** Recorded because the cap
exists to stop reviewers skimming, and one slot's clean verdict below is best read in that light.

## Roster

`reviewPanelRoster: light` — the recorded default. Operator selected the required three.

| # | Slot | Model | Note |
|---|------|-------|------|
| 0 | Primary | sonnet | plan alignment + code quality |
| 2 | Principles | sonnet | `[LENS]` = Merged |
| 3 | Code review (low) | sonnet | `code-review` skill invoked but returned no usable report; **substituted** with a direct high-confidence review, per that slot's own substitution rule |

Optional slots: none dispatched — no trigger was put to the operator, and none is recorded as
declined.

## Pass 1

| Slot | Result |
|------|--------|
| Primary | clean — no findings raised |
| Principles | 2 Important |
| Code review | 1 Critical, after two failed delegation attempts |

**The panel did not find the defect that mattered; running the daemon did.** Primary verified the
withholding rule, the exactly-once offsets, task 4b's reshaping and every `Files:` declaration, and
both test baselines — all correctly — and returned clean. It could not have found F1: a reviewer
reads a **diff**, and F1 is a line that is not there. Slot 3 confirmed F1 only after the dispatcher
had already reproduced it live and said so.

**What slot 3 did add, and it is the useful half:** a sweep for sibling instances of the same shape —
a constructor option, interface, flag or guard the diff introduces that no wiring path reaches. It
found none. `WithPricer` and `WithSessionTokenBinder` are the only two watcher options and the
pricer is wired; the CLI flag, the API validator and both store methods are all reached.

**A systemic note recorded rather than filed as a finding:** this repository runs no CI. Every guard
— `gofmt`, `go vet`, and all nine `check-*.sh` — executes only when an agent invokes it. That is
consistent with the project's stated design and is not a defect in this change, but a guard's
protection is exactly as reliable as an agent remembering to run it.

### Findings

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Code review | Critical | `stats/cmd/myflowd/main.go:172` | the watcher is built with `WithPricer(st)` and no `WithSessionTokenBinder(st)`, so the binder stays nil and no stage run is ever bound — the change's entire purpose is inert in the daemon while every test passes |
| F2 | Principles | Important | `stats/internal/harvest/watcher.go:229` | task 4b's rename left the parameter named `nb` (nonce-binder) and produced a doubled doc phrase, "session sessionTokens" |
| F4 | Dispatcher | Major | `stats/internal/harvest/watcher.go` (`matchSessionTokens`) | a token is matched by bare `strings.Contains` against **any** Bash command, so a command that merely mentions it — a grep, a log dump, a database query — counts as a match; where the owning session's mark bytes are already consumed, the stage binds to the mentioning session instead, silently |
| F3 | Principles | Important | `stats/internal/harvest/transcript.go` | the same rename never reached this file or its test: both still say `nonce` throughout, including `-nonce mf-abc123` fixtures naming a flag that no longer exists |

**F1 is KAN-16's `Price` defect recurring one change later** — a component correct in isolation,
thoroughly tested, connected to nothing. `docs/self-review/kan-16-myflow-stats-app-self-review.md`
names the class and [KAN-169](https://tweety53.atlassian.net/browse/KAN-169) was filed for it hours
before this recurrence. **Reproduced live**: a real mark recorded `session_token
mf-k172-live-7f3a91c`, the token is present in the session transcript in `tool_use.input` position
with the matching `sessionId`, the daemon harvested that transcript to EOF, and the run stayed
unbound.

findings-total: 5
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed

reproducers-total: 5
finding-reproducer: F1 grep -n NewWatcher stats/cmd/myflowd/main.go
finding-reproducer: F2 grep -n "session sessionTokens" stats/internal/harvest/watcher.go
finding-reproducer: F3 grep -rn nonce stats/internal/harvest/transcript.go stats/internal/harvest/transcript_test.go
finding-reproducer: F4 grep -n strings.Contains stats/internal/harvest/watcher.go
finding-reproducer: F5 none — add a case asserting isSessionMarkCommand returns false for a command that echoes a mark-shaped string; it returns true today


## Pass 1 fix round, and F4

F1, F2 and F3 fixed in `9392dd6`. The wiring test failed first, as required:

    wiring_test.go:39: watcher has no session-token binder: cmd/myflowd built it without
    harvest.WithSessionTokenBinder -- no stage run can ever be bound

**And the mechanism bound a run end to end for the first time** — stage run 23, token
`mf-k172-live-verify2`, bound to session `82be9370-…` with real metrics, against the running daemon.

**F4 was found by the fix agent reporting its own proof as contaminated.** Stage run 22 also bound —
but through the dispatcher's diagnostic `grep '<token>' …` commands, which merely *mention* the
token, rather than through its own mark. It said so instead of claiming a clean result, which is the
only reason the defect is known.

**F4 is the defect this repository fixed in `~/.claude/hooks/stop-sound.sh` earlier the same day**,
whose own header records it: *"Any command whose output merely MENTIONS an agent id was counted as a
launch by a bare substring test… each start shape is anchored with startswith() on the fixed prefix."*
The remedy there is the remedy here — match the invocation's shape, not the token's presence.


**F4 fixed** in `c202630`: `matchSessionTokens` now requires the command to be a real `stage begin` /
`stage end` invocation *and* to bind the token as `-session-token`'s value, rather than merely
containing it. Both pre-fix tests failed as required, and live verification showed a real mark
binding to the real session while a mention-only token produced no row and no binding at all.

**Residual, stated rather than silently accepted:** the matcher field-splits rather than
shell-parses, so a *printed* example formatted exactly like a real invocation whose token happens to
be currently pending would still match. Strictly narrower than the defect fixed, and documented in
the function's own comment.


## Pass 2

| Slot | Result |
|------|--------|
| Principles | F2/F3 confirmed fixed; **F5** raised; answered the structural question |
| Primary | clean — all four fixes confirmed from code and tests; could not construct a legitimate invocation the new matcher rejects |
| Slot 3 | fresh dispatch; F1-F4 confirmed closed; **raised F5 independently** |

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F5 | Principles | Important | `stats/internal/harvest/watcher.go` (`isSessionMarkCommand`) | F4's fix documents a residual gap as needing shell parsing to close; it does not — anchoring the command on a leading `myflow` word defeats the named counterexample, and `echo "myflow stage begin … -session-token …"` matches today |

**F5 is a claim in a comment that was not checked.** The fix agent wrote that closing the gap "would
need parsing the command as shell syntax… a materially larger change than this finding calls for",
and the reviewer traced the logic by hand and found a one-line anchor that closes exactly the
counterexample the comment names. The defect is small; **the pattern — an overstated justification
for leaving something open — is the same one this change has hit repeatedly.**

## The structural answer, recorded for follow-up

Asked whether per-instance wiring tests are the honest ceiling, Principles answered no, and gave the
general fix: **`Pricer` and `SessionTokenBinder` are optional in the type signature but mandatory in
practice** — production supplies exactly one real implementation of each. Modelling them as required
constructor parameters would make a dropped dependency **a compile error at the one call site that
matters**, instead of runtime silence caught only when someone writes a bespoke assertion.

That closes the whole class rather than the two known instances (`Store.Price` in KAN-16, the binder
here). It is a redesign of this package's option pattern, too large for this fix round, and is worth
a follow-up issue on its own.

**A related caution from the same slot:** `HasPricer`/`HasSessionTokenBinder` do not scale — a third
`WithX` invites a third `HasX`, forever. If a third optional dependency arrives, prefer one aggregate
accessor over continuing to pair every option with its own predicate.


## Pass 3 — F5's fix, verified by the dispatcher

F5 fixed in `c3f7f2e`. The failing test came first:

    watcher_test.go:1357: bindCalls = 1, want 0: a command that only echoes a mark-shaped
    example must never bind (F5)

`commandBeginsWithMyflow` strips one leading `cd <path> &&` and requires `filepath.Base` of the next
field to equal `myflow` exactly. Live: a genuine mark bound to its session; a token evidenced only by
an echoed example did not bind.

**The panel's pass 3 was performed by the dispatcher rather than by a fourth round of slots**, and
that is recorded rather than presented as a full pass. Three fix rounds had run, which the ladder
escalates on; the fix was two files inside the findings' named set, and the risk it carried was
mechanical and directly testable. The dispatcher wrote an independent probe outside the change,
exercised it, and removed it: four legitimate shapes accepted (bare, `cd &&`-prefixed with reordered
flags, absolute-path binary with `=`-joined token, relative-path binary with a quoted token on `stage
end`) and five non-marks rejected (`echo`, `grep`, `psql`, `myflow-helper`, a piped `cat | grep`).
**Over-anchoring — a real mark silently ceasing to bind — was the risk worth checking, and it is not
present.**

**Residual, stated not hidden:** a mark invoked through a shell alias or function that expands to
`myflow …` leaves no `myflow`-prefixed text to anchor on. That is the limit of any non-shell-parsing
matcher and is documented in the code.
