**Severity:** Important
**Location:** stats/cmd/flow/record.go:2066 (numberRefLessFindings)

> partial write: WriteFile truncates then writes in place, so a kill mid-write leaves a truncated findings file with no copy; the findings are the only source and are lost
