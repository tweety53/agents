# Review panel — kan-173-make-the-daemon-s-dependencies-required-so-a

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Principles | Minor | stats/cmd/flowd/wiring_test.go:201 | design.md's wiring-test-asserts-real-store decision records deleting the test and a broader reflection test as the alternatives considered, but not the narrower one actually foregone — a single exported Deps() accessor letting the test do a plain type assertion and drop reflect entirely. |

findings-total: 1
finding-status: F1 fixed

reproducers-total: 1
finding-reproducer: F1 cd stats && go test ./cmd/flowd/ -run TestDaemonWiresTheRealStore -v
