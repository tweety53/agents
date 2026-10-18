# pipeline-fix-kan-941-step-6-playwright-fidelity-specs-for-the-cardio-22

## Why

`flow self-review bundle -change kan-941-step-6-playwright-fidelity-specs-for-the-cardio` (gymie)
fails with `client: store unavailable: response body overflows the 1048576-byte cap`: the served
bundle is 2,828,523 bytes. The bundle's matrix elision (`elideMatrixRows`,
`stats/internal/selfreview/git.go`) only recognises a `####` heading as a matrix heading, and that
change's `visual-verification.md` writes its 10,047 element × property matrix rows under
`### <surface>:<id> matrix (<frame>)` headings, so every row is served. No self-review of that
change, or of any other whose verifier wrote its matrices at another heading level, can run.

## What changes

- `elideMatrixRows` treats any heading whose text holds `matrix`, in any case, at any level, as a
  matrix heading, and drops the table rows under it as before.
- `stats/internal/selfreview/bundle_test.go` gains
  `TestBundleElidesMatrixRowsUnderAnyMatrixHeading`, covering a `###` and a `##` matrix heading.
- The run-1 contract sentence describing the elision (**Run the self-review pass**,
  `skills/flow-contracts/finish-contract-run1.md`) names the new rule.
- Once landed and the live `flow` binary rebuilt, the kan-941 step-6 bundle is about 121 KB and is
  served.
