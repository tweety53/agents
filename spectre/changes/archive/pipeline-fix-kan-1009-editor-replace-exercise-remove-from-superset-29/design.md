## Context

Found by `/flow` of `kan-1009-editor-replace-exercise-remove-from-superset` at its self-review
bundle step. `curl` of the bundle route against the live flowd (pid 72512) was cut empty after
~40–55s (curl exit 52); `/api/v1/settings` answered. The machine's 15-minute load average stood at
~20 then; once it fell to ~5 the same route answered 200 in ~1s for both changes.
<!-- measured: uptime @ 19:34 → load averages 5.37 19.96 19.67; curl bundle kan-1009 → 200 1.13s -->

Cause: `Bundle` makes 26–27 git calls per change in series (counted with a timing runner against
`/Users/tweety53/Projects/gymie`), each under `DefaultGitBound` alone. Nothing bounds their sum, so
calls slowed by load add up past `writeTimeout` (30s, `internal/api/server.go`), whose expiry cuts
the connection rather than returning the degraded bundle `DefaultGitBound`'s comment promised.
`TestGitBoundInsideWriteBudget` checked one call's bound against `writeTimeout`, never the sum.
<!-- measured: Bundle with a timing ExecRunner, kan-1009 → 27 git calls 0.39s; kan-1026 → 26 calls 0.26s -->

Blast radius: 5 files cite `DefaultGitBound`, `ExecRunner` or `BundleBudget` outside archives; the
<!-- measured: git grep -l -e DefaultGitBound -e ExecRunner -e BundleBudget -- . ':!spectre/changes/archive' | wc -l @ 9e91b85c with this fix applied → 5 -->
fix touches 4 of them (`server.go` constructs `ExecRunner{}` unchanged). No saved memory encodes
this defect.

## Decisions

### Where the overall budget lives

**ID:** deadline-on-exec-runner
**Status:** active

A `Deadline` field on `ExecRunner`, set per request by the handler, caps each call's own bound at
the time left and refuses calls past it. Rejected: threading a `context.Context` through the
`Runner` interface — it touches every fake in `bundle_test.go` for the same effect; a wrapper
runner that only refuses calls past the deadline — the call in flight at the deadline would still
run its full 10s bound.

### Budget size

**ID:** budget-15s
**Status:** active

15s: the worst case is the budget plus one killed call's `waitAfterKill` (5s), 20s, inside the 30s
`writeTimeout`; a healthy bundle takes under 1s.
