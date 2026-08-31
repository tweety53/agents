## Context

The mutation-proof duty already has a structured population to mechanize against: every
`scripts/test-*.sh` harness emits `ok: <case>` / `FAIL: <case>` lines (see `test-check-base-moved.sh`),
and `run-guard-tests.sh` already treats each `test-*.sh` file as "a harness." This change scopes to
that population — Bash guard + its test harness(es) — rather than generalizing to Go/Playwright/other
runners, which have no such structured convention and no demonstrated need yet.

## Decisions

### Mutation input is a unified diff, applied via git apply

**ID:** mutation-input-is-git-diff
**Status:** active
**Chosen:** a unified-diff patch file, applied with `git apply` and reverted with `git checkout --` —
one command, one revert path.
**Considered:** an old-text/new-text string-replacement pair, matching how mutations were actually
hand-done in KAN-77. Rejected: needs its own apply/revert logic and its own uniqueness checks (old
text must match exactly once), where `git apply`/`git checkout` are already proven, git-native, and
match this repo's existing git-apply usage (the Bugbot worktree transplant in `review-panel.md`).

### Scope is Bash guard+harness pairs only

**ID:** scope-bash-guard-harness-only
**Status:** active
**Chosen:** the script targets a `check-*.sh` guard and one or more `test-*.sh` harnesses, reusing the
`ok:`/`FAIL:` line convention every harness already emits.
**Considered:** generalizing to any language/test runner (Go, Playwright) from the start. Rejected:
needs a pluggable failure-parser per framework, more code and more edge cases up front, for a need
not yet demonstrated outside Bash guards — this exact ticket's own KAN-77 example is Bash guards. A
later change can widen it if the parent's mutation duty ever needs Go/Playwright coverage.

### Blast-radius flag uses a fixed default threshold, env-overridable

**ID:** blast-radius-fixed-threshold
**Status:** active
**Chosen:** flag "suspicious blast radius" when a mutation's new-failure count (mutated minus
baseline) exceeds a default of 5, overridable via `MUTATE_AND_VERIFY_MAX_NEW_FAILURES` — the same
override idiom `RUN_REPRODUCER_BOUND_SECONDS` and `RUN_GUARD_TESTS_ROOT` already use in this
repository.
**Considered:** requiring the caller to declare expected failing case(s) up front, flagging anything
outside that set. Rejected: more precise, but assumes the parent always knows the exact case name in
advance — often the thing being discovered — and adds a required argument to every invocation.

### No file argument separate from the patch

**ID:** no-separate-file-arg
**Status:** active
**Chosen:** touched files are derived from the diff itself (`git apply --numstat`), never passed as a
separate argument.
**Considered:** the ticket's own "possible shape" names "a file, a patch, and one or more harnesses"
as three inputs. Rejected as redundant: a unified diff already names the file(s) it touches in its
headers, and a separate argument would only ever need to agree with the diff or be a source of drift
between the two.

### Exit codes report the script's own mechanics, never a verdict on the mutation

**ID:** exit-codes-mechanical-only
**Status:** active
**Chosen:** `0` ran clean (the verdict — surviving mutant / caught / suspicious blast radius — is in
the report body, not the exit code), `2` refused before mutating anything (patch does not apply, a
touched file already has uncommitted changes, a named harness is missing or not executable), `3`
could not fully restore (touched files not clean after `git checkout --`, named explicitly), `4`
cannot answer (bad usage, not a git worktree, a harness produced no readable `ok:`/`FAIL:` line on
either run).
**Considered:** none — this mirrors `run-reproducer.sh`'s existing exit-code philosophy directly, so
no alternative was weighed.

## Open questions

None.
