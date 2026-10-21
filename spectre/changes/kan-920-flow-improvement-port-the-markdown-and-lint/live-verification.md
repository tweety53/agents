# kan-920 live verification

Each shim run once on this worktree with no arguments, against the base's bash run from a detached
worktree at the merge base `75261910`, the bash pointed at this worktree (`*_ROOT` override) where
it takes one. Compared: exit code, stdout and stderr, byte for byte.

| Guard | before (bash @ 75261910) | after (shim @ HEAD) | match |
|---|---|---|---|
| `check-normative-inventory` | exit 0, 8 sentences from 119 files | exit 0, 8 sentences from 119 files | stdout and stderr byte-identical |
| `check-mutation-reproducer-pin` | exit 0, `MUTATION-REPRODUCER-PIN-OK` | exit 0, `MUTATION-REPRODUCER-PIN-OK` | byte-identical |
| `check-python-suppressions` | exit 0, `suppressions-ok` | exit 0, `suppressions-ok` | byte-identical |

Failure look (design.md `## Live check`): a shim printing nothing and exiting 0 where the bash
printed an inventory or a verdict line, or any byte difference — not seen.

<!-- measured: git worktree add --detach <wt>-before 75261910; <before>/scripts/<name>.sh vs <wt>/scripts/<name>.sh, cmp of stdout and stderr @ HEAD of spectre/kan-920-flow-improvement-port-the-markdown-and-lint -->
