# Review panel — kan-363-let-a-change-s-spectre-tree-live-in-every-repo

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Primary | Critical | skills/flow-contracts/finish-contract.md:67 | Task 11 never updated the guard-invocation call sites to resolve and pass the canonical-worktree argument tasks 7-9 added, so in a real run against a satellite worktree both guards fall through to peer resolution, fail by the design's own admission, and exit 2 — a human-in-the-loop failure for every satellite worktree. |
| F2 | Code review (low) | Major | internal/cmd/link.go:100 | The already-exists refusal only stats link.md, so spectre link silently writes into a local change directory that already holds an unrelated proposal.md and tasks.md under the same id, and validate then reports no findings because that directory no longer classifies as a satellite. |
| F3 | Code review (low) | Major | internal/cmd/link.go:166 | isValidID does not reject a backtick, and renderLink embeds the change id verbatim inside a backtick code span, so a valid id can write a malformed span into the peer's link.md that spectre validate immediately rejects while spectre link reported success. |
| F4 | Bugbot | Major | internal/parse/link.go:318 | Surviving mutant: changing the equal-bounds guard from lo >= hi to lo > hi passes the whole suite, so a Tasks here range of 5-5 is accepted and expanded to a single task, contradicting the function's own doc comment that lo must be strictly less than hi. |
| F5 | Bugbot | Major | internal/parse/link.go:288 | Surviving mutant: changing the overlap guard from lo <= max to lo < max passes the whole suite, so a token restating the exact running max is accepted and that task number is silently duplicated into TasksHere. |
| F6 | Bugbot | Major | internal/check/link.go:150 | Surviving mutant: replacing namesBack's tree.NamesFor resolution with a bare ChangeID comparison passes the whole suite, so the asymmetric-peer-name case the function's own comment calls out reports no findings where the correct code reports a one-sided link. |
| F7 | Bugbot | Major | internal/cmd/link.go:104 | Surviving mutant and real data loss: removing the local link.md already-exists refusal passes the whole suite, because the existing test's canonical side already names the satellite back so refusal 5 fires through a different path with a coincidentally identical error string; with the canonical unaware, the mutant silently overwrites a pre-existing local link.md. |
| F8 | Bugbot | Major | internal/cmd/link.go:109 | Surviving mutant: widening --force to also skip the local-exists refusal passes TestLinkForce, because that test's canonical side already lists the satellite so refusal 5 masks the widened scope, violating the documented rule that --force overrides the dirty-tree refusal alone. |
| F9 | Bugbot | Major | scripts/lib/change-plan.sh:474 | Surviving mutant: case 6c's fixture never plants a file at the escaped location, so removing the canonical change-id containment check fails for the same file-not-found reason and the test passes either way; with a real planted target the removal yields a genuine path-traversal read outside the change directory. |
| F10 | Bugbot | Minor | scripts/lib/change-plan.sh:470 | Surviving mutant: removing the peer-name containment check does not fail case 6b, because that fixture's peers lookup fails for an unrelated reason; not currently exploitable since the peer value is only a lookup key, but the test does not prove what it claims. |
| F11 | Bugbot | Major | scripts/lib/change-plan.sh:392 | Surviving mutant: removing the next-heading break that ends the Part of section scan passes all three harnesses, yet an empty Part of section followed by Parts makes the reader return a code span belonging to that later section, silently misattributing the link to the wrong peer and change id. |
| F12 | Bugbot | Minor | scripts/check-task-commit-fields.sh:149 | Surviving mutant: removing the link.md requirement from satellite detection passes the suite, but makes any change directory without a tasks.md count as a satellite candidate, so an ordinary pre-plan change beside a real satellite trips the exactly-one gate and the guard falsely refuses. |

findings-total: 12
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

reproducers-total: 12
finding-reproducer: F1 grep -c canonical-worktree skills/flow/integrate.md
finding-reproducer: F2 none — needs a two-repository filesystem fixture; reproduced by hand during review
finding-reproducer: F3 none — the id must contain a backtick, a shell metacharacter; reproduced by hand during review
finding-reproducer: F4 none — proven by a scratch test removed after verification; the fix must add it permanently
finding-reproducer: F5 none — proven by a scratch test removed after verification; the fix must add it permanently
finding-reproducer: F6 none — proven by a scratch test removed after verification; the fix must add it permanently
finding-reproducer: F7 none — proven by a scratch test removed after verification; the fix must add it permanently
finding-reproducer: F8 none — proven by a scratch test removed after verification; the fix must add it permanently
finding-reproducer: F9 none — surviving mutant, demonstrable only by replacing the changeid _change_plan_name_ok check in _change_plan_resolve_dir with true
finding-reproducer: F10 none — surviving mutant, demonstrable only by replacing the peer-name _change_plan_name_ok check in _change_plan_resolve_dir with true
finding-reproducer: F11 none — surviving mutant, demonstrable only by removing the next-heading break in _change_plan_link_part_of
finding-reproducer: F12 none — surviving mutant, demonstrable only by deleting the link.md test from the SATELLITES detection loop
