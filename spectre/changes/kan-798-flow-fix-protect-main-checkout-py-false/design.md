# kan-798-flow-fix-protect-main-checkout-py-false — design

## Context

The hook is a PreToolUse guard: fail-open on any internal fault, deny-never-rewrite, and a
deliberate token-scan stance — "widen the lists when a new shape shows up, rather than trying to
model the shell". All three fixes stay inside that stance; the fail-open contract and the deny
reason are untouched. Reachability at base `4a278320`: mode 2 and mode 3 reproduce; mode 1's
denial is already gone, and the issue's prescribed expansion adds precision while closing a live
false negative.

## Decisions

### Expand variables collected from the command text itself

**ID:** expand-same-string-assignments
**Status:** active
**Chosen:** a pre-pass collects every pure `NAME=value` token into an env map, `$NAME`/`${NAME}`
expand longest-name-first in tokens before resolution, and tokens still carrying `$` or a backtick
stay let-through — closes the `C=<main>; echo x > $C/f.txt` false negative without guessing.
**Considered:** keeping pure let-through (the base's mode-1 state) — rejected: the redirect-into-
main-through-a-variable hole stays open; reading the session's environment — rejected: the hook
process sees the harness's env, never the shell's, so it is unknowable from one command string and
the approximation would be unbounded.

### Treat a future path under `.worktrees` as outside the protected tree

**ID:** future-paths-narrow-worktrees-rule
**Status:** active
**Chosen:** when `existing_dir()`'s nearest existing ancestor is a directory named `.worktrees`,
the path is future worktree content and lets through; a path whose nearest existing ancestor is
`<main>` itself still denies.
**Considered:** the issue's looser wording, "any not-yet-existing path is outside" — rejected: it
opens a write-new-files-in-main hole (`<main>/new.txt` would sail through), defeating the guard's
purpose. Loose files written directly into `<main>/.worktrees/` are allowed — a gitignored area
the worktree-location guard already owns.

### Pre-split the command into logical lines; scan each line separately

**ID:** newline-pre-split-lines
**Status:** active
**Chosen:** split on newlines outside quotes, join backslash-newline continuations, tokenize each
logical line with the existing tokenizer, and thread `cd` state (`cur`) across lines as bash does.
**Considered:** dropping `\n` from shlex whitespace so newline becomes a standalone token —
rejected on a live spike: `\n` glues into words (`draft.md\necho`) instead of tokenizing, so scan
terminators would still miss; stopping scans at a newline token — same mechanism, same failure.

## Open questions
