# Review panel — kan-797-flow-fix-default-landing-route-never-resolves

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary+principles | minor | skills/flow-contracts/project-configuration.md:91 | the match paragraph says "exactly one literal" where ## self review model's vocabulary is the ValidModels member set, not literals |   |
| F2 | primary | minor | skills/flow/archive.md:191 | tr -d strips all backticks in the head line where the contract says surrounding backticks removed |   |
| F3 | primary | minor | stats/internal/guard/check_model_keys_test.go:332 | the divergence comment says "Two bodies deliberately diverge" while three sweep bodies left the sweep and four subtests follow |   |
| F4 | primary | minor | KNOWN-BUGS.md:96 | the task-2 entry cites check_model_keys_test.go:330 but the comment sits at :332 |   |
| F5 | primary | minor | spectre/changes/kan-797-flow-fix-default-landing-route-never-resolves/tasks.md:89 | Step 5 and the Measured bullet name scripts/check-contract-budget.sh, deleted on main before the merge base, so the step cannot run as written |   |
| F6 | primary | minor | stats/internal/guard/modelkeys.go:110-116 | the port skips a head line that strips empty while the contract words and the archive snippet take the first raw non-blank line |   |
| F7 | principles | minor | skills/flow/archive.md:191 | the head rule lives in three copies (contract, port, snippet) that have already drifted on backtick scope and stripped-empty-head selection |   |

findings-total: 7
finding-status: F1 deferred wording names the wrong vocabulary class for one key
finding-status: F2 deferred no realistic body carries an interior backtick in the value line
finding-status: F3 deferred comment count stale against the moved set
finding-status: F4 deferred one-line citation correction in the same file
finding-status: F5 deferred the named guard was ported off main after the plan was written
finding-status: F6 deferred benign divergence on a no-realistic-input edge, pinned by tests
finding-status: F7 deferred three-copy rule is the recorded design; drift noted for the next change

reproducers-total: 7
finding-reproducer: F1 none — wording only
finding-reproducer: F2 none — no realistic body carries an interior backtick in the value line
finding-reproducer: F3 none — comment wording only
finding-reproducer: F4 none — stale citation
finding-reproducer: F5 none — plan text
finding-reproducer: F6 none — benign divergence, live-verified both sides
finding-reproducer: F7 none — design observation

## Pass log

### Round 0

- roster: compact — 68
- diff size 385 under cap
- docs-only: exit 1 — scripts/check-model-keys.sh
- no addition this round — the resolved list ran alone
