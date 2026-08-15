# Design — kan-174

## Decision: recognise a mark by its invocation, not by its position

**Chosen.** Drop `commandBeginsWithMyflow`. A command is a mark when it contains `stage begin` or
`stage end` **and** binds the correlator as `-session-token`'s value — KAN-172's state after F4, before
F5 over-anchored it.

**Alternatives rejected:**

- *Anchor `myflow` at a command position* (start of text, or after `;`, `&&`, `||`, `|`, newline).
  Handles every real shape and keeps most of F5's protection. Rejected as more machinery for a
  narrower benefit than it appears: the echoed text it still admits merely has to contain a newline
  before `myflow`, which is common in exactly the multi-line blocks this defect is about.
- *Match the CLI's own printed output instead of the command text.* Immune to wrapping, since the CLI
  controls it. Rejected: it re-opens the console-noise question settled against in KAN-172, and moves
  the correlator from something the caller writes to something the callee prints, which is a larger
  change than a matcher fix.

## Decision: accept the echoed-example false positive, and say so

**The asymmetry is the whole argument.** A false negative means no mark is ever recognised: silent,
total, and exactly the state this change repairs. A false positive needs a mark-shaped string
carrying a **currently pending** correlator, and where two sessions match it, the ambiguity rule
already refuses to bind rather than choosing.

Two reviewers rated the echoed-example gap Important in KAN-172, and closing it produced this defect.
That is evidence about the trade, not against it: the closure was worth attempting and its cost was
higher than its benefit. **The residual is documented at the matcher rather than left implicit** —
KAN-172's own comment claiming shell parsing was needed to close it was already disproved once, so
the comment here states what is admitted and why it is preferred to the alternative.

## Decision: read state, then mark

**Chosen.** `/myflow-fast`'s state gate reads the change's state before emitting its first mark, and a
record whose only author is a synthetic mark does not satisfy a state gate.

Observed on this change's own creating run: marking `do.state-gate` auto-created
`{"state":"STARTED","updatedBy":"myflow stage begin (synthetic)"}`, and `/myflow-fast` accepts only
*no state* or `IN_PROGRESS`. **The gate manufactured the state that would make it refuse.**

**Both halves, not either.** Reordering fixes this command; the synthetic-record rule fixes the class,
because any command that marks before reading hits the same wall. Ordering is the cheap fix and the
rule is the durable one.

## What this change does not touch

The correlator mechanism, the binding, one-way binding, the bounded give-up, the ambiguity refusal,
withholding-while-unbound, the stage-key vocabulary, and the third absence state. This is a matcher
predicate and an ordering rule.

## On the test corpus

The tests are hand-written and must cover the shapes actually emitted: a bare invocation, one behind
`cd … &&`, one after variable assignments, one on a later line of a multi-line block, one with the
token quoted, one with `-session-token=`. **The defect this change repairs was found in production
after a probe of four invented shapes reported no over-anchoring** — the corpus is only worth what its
resemblance to reality is worth.
