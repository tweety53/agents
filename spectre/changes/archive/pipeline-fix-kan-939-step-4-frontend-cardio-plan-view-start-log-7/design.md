## Context

`bash_hits` in `hooks/protect-main-checkout.py` tokenizes each logical line, expands only the
`NAME=value` assignments the command text itself makes, and, on `git … push`, hands the push's
words and its resolved repository (`target`) to `unarchived_landing` and `loop_landing`. `resolve`
returns `None` for any path holding `$` or a backtick, so `git -C $W push …` and `cd $W && git
push …` reach both checks with `target` None, and both return None at once. With a literal
repository but an unexpanded source (`"${B}:main"`, `$(…)`, backticks), `unarchived_landing`'s
`ls-tree` of the literal source fails and returns None. Either way a landing push passes
unchecked; agents commit 0b055e56 landed a `pipeline-fix-*` branch with no flowd records that way.

Reproduced at plan time on the untouched hook (worktree HEAD 881de558): the six deny cases the
plan adds all come back `allow`, and pass once the fix is applied.

The change is one hook function and its harness cases — one task, one commit.

## Decisions

### The check lives in the hook's push scan

**ID:** check-in-push-scan
**Status:** active
**Chosen:** a new `unreadable_landing(args, target)` in `hooks/protect-main-checkout.py`, called
from `bash_hits`' push branch before the two existing landing checks, which are skipped when it
denies — the only place that sees the typed command text.
**Considered:** `scripts/check-loop-landing.sh` — it is handed a resolved branch and never sees the
command text, so it cannot know a name was unexpanded; a git `pre-push` hook — not installed in
every clone, and a session can bypass it with `--no-verify`.

### Deny, never resolve the expansion

**ID:** deny-unreadable-push
**Status:** active
**Chosen:** deny with a reason asking for literal names (`git -C /abs/path push origin
<branch>:<default>`), and naming `scripts/pipeline-fix.sh land <worktree>` for a fix loop.
**Considered:** evaluating the expansion (running `$(git branch --show-current)` or reading the
environment) — the hook process holds no session environment, and running command text from a
hook executes untrusted code; letting it through as now — that is the defect.

### An unknown repository triggers the check as an expansion does

**ID:** unknown-repository-counts
**Status:** active
**Chosen:** the check runs when `target` is None (a `-C $W`, a `cd $W`) or any push word holds `$`
or a backtick; a `HEAD` destination with no repository is undeterminable.
**Considered:** inspecting only the `-C` token — leaves `cd $W && git push origin HEAD:main`
landing unchecked, the same defect under another spelling.

### A literal non-default destination passes

**ID:** literal-non-default-passes
**Status:** active
**Chosen:** deny only when some destination is the default branch (`origin/HEAD`'s target, else
main/master/develop), an expansion, or undeterminable; an expansion in an option word (a split
backtick refspec, `--repo=$X`) also counts as undeterminable.
**Considered:** deny every push carrying an expansion — breaks the branch-backup pushes that
`skills/flow/implement.md`, `skills/flow-fast/SKILL.md` and
`skills/flow-contracts/git-boundaries.md` have agents type after every commit, often with a
variable worktree path, and none of which lands on the default branch.

### Existing callers keep their typed shapes; the two saved memories are updated

**ID:** callers-and-memories
**Status:** active
**Chosen:** no skill text changes — every typed push in `skills/` names `<placeholders>` an agent
substitutes, and the deny reason carries the literal form; scripts (`land-self-review-report.sh`,
`pipeline-fix.sh land`, `refresh-main-checkout.sh`) are untouched since the hook sees only typed
Bash. The fixer updates the two gymie memories that state the hook's behaviour on `$VAR`:
`protect-main-checkout-hook-quirks.md` ("`$VAR` targets … pass") gains that a default-branch push
named through `$VAR` is denied, and `archive-push-base-checked-out-in-main.md` ("update-ref …
then push") says to push with a literal `-C` path and branch name.
**Considered:** a sentence in each skill forbidding variables in pushes — restates the hook in
prose that drifts from it.

### First review on medium

**ID:** first-review-effort
**Status:** active
**Chosen:** the first review, `pipeline-fix-7-review-1`, runs on `flow-medium` (effort `medium`)
on `opus`: the check sits on every typed push and turns on shell-tokenization edge cases.
**Considered:** `low` — too little for a gate on the landing path whose misses are silent.

### One fixer group

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group — Task 1 in one `pipeline-fix-7-fix` dispatch.
**Considered:** none — the plan has one task.

## Open questions

## Live check

No runtime to verify against: the hook is a one-shot process per tool call, and the harness feeds
the real hook file the same payloads the harness runs.
