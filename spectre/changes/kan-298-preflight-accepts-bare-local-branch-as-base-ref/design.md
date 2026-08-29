# kan-298-preflight-accepts-bare-local-branch-as-base-ref

## Context

KAN-298 reports that `check-finish-preflight.sh` accepts a bare local branch name as `<base-ref>`
without saying so, and that a stale local branch of that name would then silently feed the
RUN1/RUN2/REFUSE decision. The ticket offered two fixes and called the first safer: have the guard
compose `origin/` itself. That half landed under KAN-88 as `scripts/lib/resolve-remote-base.sh`'s
`resolve_remote_base`, sourced by `check-finish-preflight.sh:99` and `check-base-moved.sh:78`. This
change delivers the ticket's remaining sentence — update the usage line — and changes no behaviour.

## Where the text goes

Each guard states the base-ref rule in **two** places today, and only one of them is silent:

| Site | Currently | Action |
|------|-----------|--------|
| `check-finish-preflight.sh`'s header, "What THIS script owns (KAN-88 …)" paragraph | states the substitution in full | unchanged |
| `check-base-moved.sh`'s header, "Base-ref resolution is shared with check-finish-preflight.sh via scripts/lib/resolve-remote-base.sh" | states it by reference | unchanged |
| both guards' runtime usage message, printed to stderr on a missing argument | silent | **gains the rule** |

The runtime message is the site KAN-298 quotes, and the one a caller who got the invocation wrong
actually reads. Putting the text there and nowhere else leaves exactly one copy of the new phrase
per guard, so the harness grep has a single unambiguous target and there is no second copy to drift.

## The message

```text unverified:this is the text this change introduces; confirm against the guard's stderr after task 1
usage: check-finish-preflight.sh <worktree> <base-ref> <recorded-merge-base|->
  <base-ref>  the base branch name, bare (main) or remote-tracking
              (origin/main). The guard prefers refs/remotes/origin/<base-ref>
              when it resolves, so a bare name is never tested against a
              stale local branch.
```

Identical in `check-base-moved.sh` with its own guard name. The single-line `echo … >&2` becomes a
quoted heredoc, so the wrapped text is written literally rather than assembled from several echoes.

## Testing

A behaviour test on the emitted message would work, but the harnesses' established precedent for
"text whose deletion no verdict can reveal" is a source grep — `test-check-finish-preflight.sh`'s
case 14 already greps `lib/resolve-remote-base.sh` for `--end-of-options` for exactly this reason.
One case per harness, on the phrase `prefers refs/remotes/origin/<base-ref>`.

## Decisions

### Where the qualification is written

**ID:** usage-message-is-the-only-site
**Status:** active
**Chosen:** the runtime usage message alone — the site KAN-298 quotes, and the only one of the three
that is silent today. One copy per guard, so the harness grep is exact and nothing can drift.
**Considered:** *also restate it in each header's `# Usage:` synopsis line* — rejected: it creates a
second copy of a normative sentence that no guard compares against the first, and both headers
already carry the rule in prose. *Replace the header prose with the usage text* — rejected: the
header paragraphs carry the reasoning (why a stale local branch is dangerous), which a parameter
line has no room for.

### Both guards, not only the one KAN-298 names

**ID:** both-guards-share-the-defect
**Status:** active
**Chosen:** update `check-finish-preflight.sh` and `check-base-moved.sh` together. They source the
same `resolve_remote_base`, print the same three-argument usage shape, and are called back to back
by the same finish-contract step, so a caller misled by one is misled by the other.
**Considered:** *`check-finish-preflight.sh` only, exactly as filed* — rejected: it would leave the
identical silent usage line one file away, reachable from the same contract paragraph.

### How the deletion is caught

**ID:** source-grep-over-behaviour-test
**Status:** active
**Chosen:** `grep -qF` on the guard's own source, one case per harness, following
`test-check-finish-preflight.sh`'s case 14.
**Considered:** *invoke the guard with no arguments and assert stderr* — rejected: it tests the same
string through a longer path and would still pass if the phrase were reworded into something that no
longer states the rule, which is the failure mode worth catching. *No test* — rejected: comment and
usage text is exactly what a later edit drops without any verdict changing.

### The finish contract is not edited

**ID:** contract-guidance-stays
**Status:** active
**Chosen:** leave `skills/flow-contracts/finish-contract.md`'s "**`<base-ref>` is composed as
`origin/$BASE`**" instruction exactly as written.
**Considered:** *soften it now that the guard substitutes* — rejected: the guard's substitution is a
second line of defence, not a replacement. A caller that composes the remote-tracking ref itself is
still correct, and is the only form that works in a repository where the substitution finds nothing.

## Open questions

None.
