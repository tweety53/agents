# Review panel — kan-362-myflow-guard-test-suite-takes-119s-sequentially

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Principles | Minor | scripts/lib/parallel.sh:209 | parallel_replay_failures is exported, documented and unit-tested as part of the library public interface, but neither real caller uses it: run-guard-tests.sh reimplements its own list-order/failure-only replay loop, and test-check-installed-citations.sh replays through per-case check functions instead. The honest answer is it serves zero callers today. |
| F2 | Primary | Major | .flow/project.md | Raised by the dispatcher on the evidence gathered after task 4 committed, not by the Primary slot itself. The runtime note records the intermittent failure as "an apparent flake under nested-parallel oversubscription ... not a reproducible defect". That is now false: it was reproduced (3 of 8 runs), root-caused as the temp-dir-observing tests scanning the global TMPDIR and picking up a sibling process directory, and fixed in 2e3ed4f. Shipping a note that tells a future reader the failure was unreproducible would send them down the wrong path. |
| F3 | Code review (low) | Minor | scripts/lib/parallel.sh:157 | parallel_run does not neutralize set -e/pipefail around its xargs -P pipeline. Under a caller s set -euo pipefail, calling parallel_run without first doing set +e aborts the calling shell silently, before the return code is even inspected, the moment any job in the batch fails. All three current call sites bracket the call correctly, but the library header says nothing about this requirement. |
| F4 | Bugbot | Minor | scripts/test-check-installed-citations.sh:949 | Surviving mutant. The CHECK-phase loop carries an explicit ordering contract in its own comment - replays every case in the original build order so ok:/FAIL: lines print in exactly the order a reader of the old sequential harness would have seen. Reversing the loop leaves the harness exiting 0 with all cases passing and zero FAIL lines; no test in this file or any other inspects the line ordering. |

findings-total: 4
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed

reproducers-total: 4
finding-reproducer: F1 grep -rln parallel_replay_failures scripts/ .flow/
finding-reproducer: F2 git -C <worktree> log --oneline 2e3ed4f -1; grep -n "looks like a flake" <worktree>/.flow/project.md
finding-reproducer: F3 printf "#!/bin/bash\nset -euo pipefail\nsource \$(pwd)/scripts/lib/parallel.sh\nparallel_run \"exit 0\" \"exit 3\"\necho overall_rc=\$?\n" > /tmp/repro.sh; /bin/bash /tmp/repro.sh; echo "script exit=$?"
finding-reproducer: F4 cp scripts/test-check-installed-citations.sh /tmp/tci.sh; reverse the CHECK-phase while loop to iterate from ${#CMDS[@]}-1 down to 0; bash /tmp/tci.sh  # still exits 0, all cases pass
