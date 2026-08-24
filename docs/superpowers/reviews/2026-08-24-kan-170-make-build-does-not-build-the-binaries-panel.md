# Review panel — kan-170-make-build-does-not-build-the-binaries

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Principles | Minor | stats/internal/pidfile/pidfile.go:Release | Release no-ops silently when the file no longer parses or names another daemon, with no log line, unlike Check which logs every stale-file reason it finds; an operator gets no shutdown diagnostic for a takeover. The slot cited a diff-relative line; the real location is Release in a 238-line file. |
| F2 | Code review (low) | Critical | stats/Makefile:332 | pidfile.Write emits two lines (pid, executable path) but ui-test-up and ui-test-down still cat the whole file into pid, so kill -0 and kill get a multi-line argument and always fail; a live previous daemon is never detected or stopped, and ui-test-down proceeds to an unforced dropdb racing it. Also at stats/Makefile:333 and stats/Makefile:398. Fix: head -n1 at all three sites, plus a harness case that catches it — the slot's own shell-line reproducer was refused by check-panel-reproducers for carrying metacharacters and an absolute path, and the operator chose to close that gap with the harness case rather than leave the finding unverifiable. |
| F3 | Primary | Minor | stats/internal/pidfile/pidfile.go:Check | Check propagates any non-ErrNotExist read error (e.g. permission denied) as a hard error, aborting startup, but the spec enumerates four stale conditions and requires every other case to be logged, overwritten and started past; an unreadable file is a non-refusal condition that currently blocks startup. Demonstrated by chmod 0o000 on a valid pidfile: Check returns permission denied instead of nil. |
| F4 | Primary | Major | stats/internal/pidfile/pidfile.go:1 | Commits 0b6982f and 50ef6e6 carry bodies describing their pre-fix content: 50ef6e6 says the harness gains two cases and never mentions the cat-to-head-n1 fix or case 6, which is most of its diff; 0b6982f lists four stale conditions omitting unreadable, and never mentions Release taking a slog.Logger. The round-0 fixups were squashed in without updating the messages, so git log misdescribes both. |

findings-total: 4
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed

reproducers-total: 4
finding-reproducer: F1 none — observability gap, not a functional defect; demonstrable only by code inspection at the cited line
finding-reproducer: F2 scripts/test-make-build.sh
finding-reproducer: F3 none — the demonstration is a throwaway Go test inside package pidfile, not a runnable command
finding-reproducer: F4 none — the defect is in two commit messages, not in a file a command can exercise
