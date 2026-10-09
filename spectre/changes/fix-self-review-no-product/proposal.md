# fix-self-review-no-product

## self-review-scope

### Why

The KAN-924 self-review offered, and filed, two gymie product bugs: `skills/flow-self-review/SKILL.md` classified a product-code finding as `big` and offered it for filing when Important or worse. The operator: "why do I see gymie product-related issues during self-review? It must be only about flow itself and maybe gymie dev tooling at most. Product related issues MUST be fixed during the flow run".

### What changes

- The self-review pass covers the pipeline and the project's own dev tooling (build scripts, dev-stack and test-harness config, guards) only; a product-code finding is never offered, filed or recorded.
- A product defect a run finds is fixed inside that run, before integrate lands the change, by a fix run at `IN_PROGRESS`.

## known-bugs-retired

### Why

The operator: "SO I don't like the idea of KNOWN_BUGS.md in general riight now. Must be fixed when it was found." `skills/flow-contracts/known-bugs.md` had a run record a pre-existing test failure or visual departure in `<project>/KNOWN-BUGS.md` and feed `## known failures`, never repairing it, and let it pass verify without blocking.

### What changes

- `skills/flow-contracts/known-bugs.md` is rewritten in place to the fix-in-run rule: a product defect a run finds — pre-existing, at the merge base, or on a surface another change owns — blocks and is fixed in that run, never logged to a `KNOWN-BUGS.md` file, added to `## known failures`, or deferred. Kept at its path so no citation dangles.
- Every citer's "except a failing test the sweep classifies pre-existing" exemption is removed; a verification change's sweep turns every finding into an appended task.
- `skills/flow/implement.md` no longer mentions `<project>/KNOWN-BUGS.md`.
