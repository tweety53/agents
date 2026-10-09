# fix-self-review-no-product

## self-review-scope

### Why

The KAN-924 self-review offered, and filed, two gymie product bugs: `skills/flow-self-review/SKILL.md` classified a product-code finding as `big` and offered it for filing when Important or worse. The operator: "why do I see gymie product-related issues during self-review? It must be only about flow itself and maybe gymie dev tooling at most. Product related issues MUST be fixed during the flow run".

### What changes

- The self-review pass covers the pipeline and the project's own dev tooling (build scripts, dev-stack and test-harness config, guards) only; a product-code finding is never offered, filed or recorded.
- A product defect a run finds is fixed inside that run, before integrate lands the change, by a fix run at `IN_PROGRESS`.
