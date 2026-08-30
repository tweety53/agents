# Review panel — kan-291-myflow-add-myflow-tasks-tick-so-ticking-a

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Principles | Critical | stats/cmd/flow/tick.go:59 | tickTask's body-boundary scan has no fence tracking, unlike scripts/lib/plan_grammar.py's iter_tasks (FENCE_RE/in_fence) it claims to mirror exactly -- a fenced example line inside a task's own body that merely looks like a task/heading line silently truncates that task's real body early, leaving later real step checkboxes unticked |
| F2 | Principles | Important | stats/cmd/flow/tick.go:40 | headingPattern (^(## \|### )) requires a literal trailing space, diverging from the canonical BODY_BOUNDARY_RE (^#{2,3}(?:\s\|$)) it claims to match, so a bare heading with no title (## alone) is not recognised as a body boundary; Bugbot's mutation pass separately found the end-scan loop's heading-check half of the OR condition is uncaught by any existing test (no fixture places a heading with no following task line) |
| F3 | Bugbot | Important | stats/cmd/flow/tasks.go:67 | changeNamePattern's containment restriction (must start with a letter or digit, blocking a leading '.'/'_'/'-') is entirely untested in stats/cmd/flow/tasks_test.go -- broadening it is uncaught by the Go suite, even though two panel reviewers independently confirmed containment holds via live CLI runs against ../escape and /etc/passwd |
| F4 | Bugbot | Minor | stats/cmd/flow/tick.go:35 | taskLinePattern requires a space after the id's dot (12. Title, not 12.Title); no test asserts this requirement, so dropping it is uncaught |
| F5 | Principles | Minor | stats/cmd/flow/tasks.go:52 | specRootLeaf ports spec-root.sh's spectre-vs-openspec probe but drops the shell version's stderr note flagging a half-migrated project (both trees present) -- functionally harmless, purely a lost diagnostic |
| F6 | Bugbot | Minor | stats/cmd/flow/tick.go:127 | the step-checkbox scan loop's own fence tracking has zero coverage of its own -- TestFenceTruncatesBody only exercises the body-boundary scan's fence handling, not this third loop's; forcing inFence=false in this loop alone survives the full suite |
| F7 | Bugbot | Minor | stats/cmd/flow/tick.go:47 | fencePattern's {0,3} indent limit (CommonMark's fence-indent boundary) has no test; widening it to {0,4} survives the full suite |
| F8 | Bugbot | Minor | stats/cmd/flow/tick.go:42 | headingPattern's #{2,3} level restriction has no test against a level-1 heading; widening to #{1,3} survives the full suite |
| F9 | Bugbot | Minor | stats/cmd/flow/tick.go:54 | stepCheckboxPattern's exact two-space indent requirement is untested; widening it to any indent width survives the full suite |
| F10 | Bugbot | Minor | stats/cmd/flow/tick.go:54 | stepCheckboxPattern's [ ] mark (unticked only) is untested; widening it to match any single character (including an already-ticked [x]) survives the full suite |
| F11 | Bugbot | Minor | stats/cmd/flow/tasks.go:60 | specRootLeaf's info.IsDir() check on both branches is untested; a plain file at the probed path (instead of a directory) is not distinguished by any existing test |
| F12 | Bugbot | Minor | stats/cmd/flow/tasks.go:94 | flow tasks tick -h/--help's exit-0 special case (errors.Is(err, flag.ErrHelp)) is untested |
| F13 | Bugbot | Minor | stats/cmd/flow/tasks.go:101 | extra trailing positional arguments beyond <change> <task-id> are untested; NArg()!=2 could regress to a looser check unnoticed |
| F14 | Bugbot | Minor | stats/cmd/flow/tasks.go:74 | changeNamePattern's trailing $ anchor is unverified by any test -- every existing rejection case fails on its first character, none prove the match must cover the whole string (e.g. a name valid as a prefix but carrying trailing '/../' traversal characters) |

findings-total: 14
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 fixed
finding-status: F9 fixed
finding-status: F10 fixed
finding-status: F11 fixed
finding-status: F12 fixed
finding-status: F13 fixed
finding-status: F14 fixed

reproducers-total: 14
finding-reproducer: F1 cd stats && go test ./cmd/flow/... -run TestFenceTruncatesBody -v
finding-reproducer: F2 none — regex-behavior claim, exercised by a new unit test in the fix rather than a standalone shell command
finding-reproducer: F3 cd stats && sed -i.bak 's/\^\[A-Za-z0-9\]/^[A-Za-z0-9._-]/' cmd/flow/tasks.go && go test ./cmd/flow/... -v; mv cmd/flow/tasks.go.bak cmd/flow/tasks.go
finding-reproducer: F4 cd stats && sed -i.bak 's/\\. `/\\.`/' cmd/flow/tick.go && go test ./cmd/flow/... -run TestTickTask -v; mv cmd/flow/tick.go.bak cmd/flow/tick.go
finding-reproducer: F5 none — an observability gap, not a behavior difference (both versions return the same spec-root value when both trees exist)
finding-reproducer: F6 none — a coverage gap, exercised by a new unit test in the fix rather than a standalone shell command
finding-reproducer: F7 none — a coverage gap, exercised by a new unit test in the fix rather than a standalone shell command
finding-reproducer: F8 none — a coverage gap, exercised by a new unit test in the fix rather than a standalone shell command
finding-reproducer: F9 none — a coverage gap, exercised by a new unit test in the fix rather than a standalone shell command
finding-reproducer: F10 none — a coverage gap, exercised by a new unit test in the fix rather than a standalone shell command
finding-reproducer: F11 none — a coverage gap, exercised by a new unit test in the fix rather than a standalone shell command
finding-reproducer: F12 none — a coverage gap, exercised by a new unit test in the fix rather than a standalone shell command
finding-reproducer: F13 none — a coverage gap, exercised by a new unit test in the fix rather than a standalone shell command
finding-reproducer: F14 none — a coverage gap, exercised by a new unit test in the fix rather than a standalone shell command
