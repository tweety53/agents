# Review panel — kan-289-reproduce-rather-than-read-in-slot-template

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Code review (low) | Critical | stats/internal/store/testsupport_test.go:23 | ten test files and the Makefile pinned the pre-rename DSN credentials, so every database-backed package breaks after the operator cutover |
| F2 | Principles | Minor | scripts/check-guard-symlinks.sh:892 | a live present-tense comment still described flow-settings as calling the myflow CLI |
| F3 | Primary | Major | stats/README.md:1 | task 19 promised and declared stats/README.md runbook content that its commit never added |
| F4 | Primary | Major | spectre/changes/kan-289-reproduce-rather-than-read-in-slot-template/design.md:1 | the decision to pin the test DSN credentials was never recorded as a decision nor in the runbook |
| F5 | Primary | Minor | spectre/changes/kan-289-reproduce-rather-than-read-in-slot-template/tasks.md:943 | task 18's declared Commit field did not match the commit subject actually used |
| F6 | Code review (low) | Critical | stats/internal/store/testsupport_test.go:45 | the hand-rolled DSN derivation silently returned a corrupted connection string for a query value containing a slash, a socket-host DSN, a libpq keyword DSN and an empty DSN |
| F7 | Principles | Major | stats/internal/store/testsupport_test.go:43 | the same derivation was copied into seven files, so one bug needed seven fixes |
| F8 | Primary | Major | spectre/changes/kan-289-reproduce-rather-than-read-in-slot-template/design.md:422 | the design decision claimed both READMEs documented the DSN override; the root README had no mention of it at all |
| F9 | Primary | Major | spectre/changes/kan-289-reproduce-rather-than-read-in-slot-template/tasks.md:1 | task 10's Files field still carried a glob, so the guard rejected all six files its commit touched |
| F10 | Primary | Minor | stats/cmd/uitest-seed/guard_test.go:14 | a doc comment named flow_uitest_real while the literal beside it still said myflow_uitest_real |
| F11 | Code review (low) | Critical | stats/internal/dsn/dsn.go:78 | the new package routed on "://" appearing anywhere, so a keyword DSN carrying a scheme inside a value was parsed as a URL and reduced to "/scratch" with a nil error |
| F12 | Code review (low) | Minor | stats/internal/dsn/dsn.go:109 | the no-scheme guard was unreachable given the prefix routing, and its comment claimed it prevented silent corruption it could never see |
| F13 | Principles | Major | stats/internal/dsn/dsn.go:95 | the keyword-form refusal scanned the whole string, rejecting ordinary overrides carrying a quote in another keyword and contradicting the function's own documented contract |
| F14 | Principles | Minor | stats/internal/dsn/dsn.go:75 | one incident was narrated three times in one file; the inline retelling is now a pointer to the package doc |

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
finding-reproducer: F1 cd stats && grep -rln "myflow:myflow@localhost" --include=*.go .
finding-reproducer: F2 sed -n 892p scripts/check-guard-symlinks.sh
finding-reproducer: F3 git diff 07acd2e..HEAD --stat -- stats/README.md
finding-reproducer: F4 grep -n ADMIN_DSN spectre/changes/kan-289-reproduce-rather-than-read-in-slot-template/design.md
finding-reproducer: F5 git log -1 --format=%s $(git log --format='%H %s' 07acd2e..HEAD | grep 'contract-budget table audit' | cut -d' ' -f1)
finding-reproducer: F6 cd stats && go test ./internal/dsn/ -run TestForDatabaseHandlesTheShapesStringSurgeryCorrupted -count=1 -v
finding-reproducer: F7 grep -rlc "dsnutil.ForDatabase" --include=*_test.go stats/
finding-reproducer: F8 grep -c FLOW_STATS_ADMIN_DSN README.md
finding-reproducer: F9 scripts/check-task-commit-fields.sh <worktree> 10 <task-10-sha>
finding-reproducer: F10 sed -n 14p;34p stats/cmd/uitest-seed/guard_test.go
finding-reproducer: F11 cd stats && go test ./internal/dsn/ -run TestForDatabaseRoutesOnSchemeNotOnASubstring -count=1 -v
finding-reproducer: F12 delete the u.Scheme=="" block and run: cd stats && go test ./internal/dsn/ -count=1  (all tests still pass)
finding-reproducer: F13 cd stats && go test ./internal/dsn/ -run TestForDatabaseRoutesOnSchemeNotOnASubstring -count=1 -v
finding-reproducer: F14 none — a documentation-duplication judgement, verified by reading the three retellings in one 139-line file
