# Review panel — kan-271-move-panel-findings-out-of-markdown-and-into

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Code review (low) | Critical | stats/cmd/myflow/record.go:47 | validateFindingStatus uses strings.CutPrefix(status, "withdrawn") with no separator check, so a concatenated value like "withdrawnfoo" (no space) is accepted as a legal status and can be written to the store. |
| F2 | Code review (low) | Important | stats/cmd/myflow/record.go:66 | validateFindingReproducer treats a whitespace-only string (e.g. " ") as neither empty nor a bare 'none', so it passes as a legal reproducer; check-panel-reproducers.sh's jq empty-check and command-safety loop both also let it through, reporting REPRODUCERS-OK for an effectively blank reproducer. |
| F3 | Principles | Important | scripts/check-workspace-isolation.sh:109 | Four files (check-workspace-isolation.sh, lib/resolve-file.sh, check-self-review-report.sh, lib/coverage.sh) cite scripts/lib/panel-record.sh's header as the canonical source of the -a/rc>1/-- grep disciplines, but this diff deletes that file — the citations now point at nothing, an SSOT violation. |
| F4 | Principles | Important | scripts/check-panel-reproducers.sh:183 | jq calls in check-panel-reproducers.sh and check-unfinished-work.sh are unguarded plain command substitutions under set -euo pipefail; a missing jq binary aborts with exit 127 and jq's own error text instead of the guard's documented exit-2 'cannot determine anything' contract. |
| F5 | Primary | Critical | stats/cmd/myflow/record.go:1067 | runRecordFindings only normalizes Findings to []records.Finding{} on the ErrNotFound branch; a change that exists (has dispatch rows) but raised zero findings leaves Findings nil, so json.Marshal prints literal null instead of [], which both rewritten guards' jq pipelines (starting with .[]) crash on under set -euo pipefail. |

findings-total: 5
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed

reproducers-total: 5
finding-reproducer: F1 .superpowers/sdd/repro/F1.sh
finding-reproducer: F2 .superpowers/sdd/repro/F2.sh
finding-reproducer: F3 .superpowers/sdd/repro/F3.sh
finding-reproducer: F4 .superpowers/sdd/repro/F4.sh
finding-reproducer: F5 .superpowers/sdd/repro/F5.sh
