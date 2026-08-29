# kan-298-preflight-accepts-bare-local-branch-as-base-ref

**Jira:** KAN-298

## Why

`check-finish-preflight.sh` and `check-base-moved.sh` both print

```text verified:quoted from scripts/check-finish-preflight.sh and scripts/check-base-moved.sh at branch main
usage: <guard>.sh <worktree> <base-ref> <recorded-merge-base|->
```

on a missing argument. Nothing in that message says what `<base-ref>` should be, or that the guard
substitutes `refs/remotes/origin/<base-ref>` when it resolves. A caller reading only the usage
message — the failure KAN-298 reports from KAN-295, where `resolve-base-branch.sh` printed `main`
and the natural next call passed `main` straight through — still learns nothing about the rule.

KAN-298's preferred fix, *have the guard compose `origin/` itself*, already landed under KAN-88 as
`scripts/lib/resolve-remote-base.sh`'s `resolve_remote_base`, sourced by both guards. What is
outstanding is KAN-298's own last sentence: **update the usage line either way.**

## What changes

- `scripts/check-finish-preflight.sh` — the runtime usage message gains a `<base-ref>` line stating
  that bare and remote-tracking names are both accepted, and that the guard prefers
  `refs/remotes/origin/<base-ref>` when it resolves.
- `scripts/check-base-moved.sh` — the same line, guard name substituted. It shares the lib and
  misleads identically.
- `scripts/test-check-finish-preflight.sh`, `scripts/test-check-base-moved.sh` — one source-grep
  case each, asserting the rule's load-bearing phrase survives; plus two runtime assertions each,
  added at the review panel's insistence, that the no-arguments path leaves stdout empty and writes
  exactly the usage message to stderr.
- `scripts/lib/base-ref-usage.sh` — new. Holds the expected usage text once, as an independent
  golden value both harnesses source. It is never derived from either guard: doing so would make
  both assertions tautological.

No behaviour changes. Every verdict, exit code and signal order is untouched — the guards' own
production code carries only the replaced usage message.
