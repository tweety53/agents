## Context

In-run pipeline fix (**Pipeline defects found mid-run**, `skills/flow-contracts/pipeline.md`),
loop 28, found by run 2 cleanup of gymie change kan-1026, which stopped on `STOP: disclose` for
`UNCLASSIFIED: /Users/tweety53/Projects/gymie-worktrees/kan-1026-stored-data-first-on-every-open — .env`.
The entry was a symlink to `/Users/tweety53/Projects/gymie/.env`, created by gymie's `linkEnvFile`
(`/Users/tweety53/Projects/gymie/.flow/project.md:104`).

Check 4's preservation lives in `stats/internal/guard/removechangeworktrees.go`: the root-entry
case calls `rcwCounterpart`, which `os.Lstat`s both paths and requires two regular files that are
not the same file. A symlink's `Lstat` mode is not regular, so the linked `.env` falls to the
default, unpreserved case and the disclosure stop fires. The rule's text is the preservation
sentence under **Worktree cleanup** in `skills/flow-contracts/finish-contract-run2.md`, mirrored in
the header of `scripts/remove-change-worktrees.sh`.

**Reproduced.** Task 1's test case, added to the untouched `removechangeworktrees.go`, fails with
exit 3 and `UNCLASSIFIED: <wt> — .env` followed by the `DISCLOSE: stopped before removal` line —
kan-1026's stop exactly.
<!-- measured: cd stats && go test ./internal/guard/ -run '^TestRemoveChangeWorktreesDisclose$' -count=1 with removechangeworktrees.go at its untouched state and the rest of Task 1's diff applied @ 8e97fb1e → FAIL, exit 3, the .env line unpreserved; with the whole diff → ok -->

**History.** kan-924 added the identical-bytes preservation and kan-999 the differing-copy case
(their comments in `remove_change_worktrees_test.go`); no archived change rejected preserving a
link. The only other in-tree citation of the sentence is the in-flight `spectre/changes/faster-integrate/`
proposal and `verbatim-moves.txt`, which quote the kan-924 wording as that change's own history and
are not edited.
<!-- measured: grep -rln over the worktree for "preserved: identical" and "distinct regular file", spectre/changes/archive excluded @ 8e97fb1e → the 4 files Task 1 touches plus faster-integrate's proposal.md and verbatim-moves.txt -->

Blast radius: 4 files plus this change's `verbatim-moves.txt` — not `big`.
<!-- measured: git diff --stat of Task 1's diff @ 8e97fb1e → 4 files changed, 43 insertions(+), 4 deletions(-) -->

## Decisions

### What counts as a harmless link

**ID:** link-to-main-regular-file
**Status:** active
**Chosen:** an entry at the worktree root that is a symlink whose target is the very regular file at the same path in the main checkout (`os.SameFile(os.Stat(a), os.Lstat(b))`, `b` regular) — removing the worktree drops the link alone and the file survives.
**Considered:** every symlink at the root — a link to a file inside the worktree, or anywhere outside the main checkout, is not a main-checkout copy and may be the only reference a cleanup should ask about; a link to a main-checkout symlink in turn — a chain no tooling makes, left unpreserved.

### How the link is classified

**ID:** link-is-preserved
**Status:** active
**Chosen:** preserved, a third label `— preserved: linked to <main-checkout>` beside `identical in` and `differs from`, through a new `rcwLinkToCounterpart` case after the existing `rcwCounterpart` case — it never stops and is still relayed per check 4.
**Considered:** counting it as regeneratable — hides the entry from the relay check 4 requires for every unclassified entry; widening `rcwCounterpart` to follow links — its distinct-file clause is what keeps the reverse case (a main-checkout symlink into the worktree) unpreserved, and following links there would blur it.

### The rule's text

**ID:** link-sentence
**Status:** active
**Chosen:** extend the one preservation sentence in `skills/flow-contracts/finish-contract-run2.md` with the symlink case and its label, the script header to match, and list the old and new sentence in this change's `verbatim-moves.txt` for `check-verbatim-moves.sh`.
**Considered:** a second sentence — two sentences stating one rule's two halves; leaving the contract unchanged — the guard would print a label the contract never names.

### Nothing to repair

**ID:** no-repair-needed
**Status:** active
**Chosen:** no repair step — the defect wrote nothing wrong: the stopped cleanup removed nothing (exit 3 is "nothing removed, nothing run"), so the kan-1026 worktree and its `.env` link are intact, and the parent's resumed cleanup re-runs `remove-change-worktrees.sh` once this lands.
**Considered:** removing kan-1026's link by hand now — the resumed cleanup does it, and a hand removal outside the guard is the bypass the stop exists to prevent.

### First review's effort

**ID:** first-review-effort
**Status:** active
**Chosen:** flow-low — one predicate and one switch case, reproduced and tested; the only risk (preserving a link that should stop) is pinned by `SameFile` against a regular main-checkout file and by the existing reverse-symlink case, which still passes.
**Considered:** flow-medium — the diff carries no seam a medium review would weigh better.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group — Task 1 alone, one fixer dispatch.
**Considered:** more groups — the plan has one task.

## Open questions

## Live check

None — `remove-change-worktrees.sh` is a one-shot guard with no running service or persistent state; Task 1's test drives it end to end against a scratch repository and worktree, and the parent's resumed kan-1026 cleanup exercises it on the real worktree.
