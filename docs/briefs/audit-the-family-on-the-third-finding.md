# Audit the family on the third finding

When the same finding shape is raised a third time against one change — the same mechanism,
a different instance, across fix rounds or review passes — stop fixing instances and audit
the whole family once: enumerate every member of the class, check each against the shape,
fix the ones the audit confirms, and prove the closure, instead of waiting for the next
instance to surface it.

An audit converts an open-ended fix stream into a bounded one. The class is enumerated once;
each member is checked; the closure is proved, not discovered piecemeal by whichever instance
a reviewer happens to find next.

## Why

kan-575's review panel raised the same defect shape four times across fix rounds 2–8: a
relative navigator mutation (`back()`, `finish()`, `replaceTop()`) reached by a late or
duplicate callback (F10, F11+F15, F19, F25/F27, F30). Each round guarded one instance, and
the class stayed open until round 9 audited every `ShellNavigator` mutator in one pass,
guarded the class closed, and confirmed the remaining mutators could not empty the graph,
pop a room, or remove Workouts. The audit, not the fourth individual fix, is what closed the
class.

## How to run one

- Collect the findings that share a shape and name the mechanism once, in the shape's own
  words.
- Enumerate the family mechanically where a listing exists — every implementor of the
  interface, every caller of the helper, every route of the navigator — never from memory.
- Check each member against the shape and fix the ones that fail, in an order where no fix
  invalidates another's guard.
- Prove the closure: a guard per member, a mutation per guard, or the equivalent witness
  that a new instance of the shape cannot appear unobserved.

Recorded from kan-575 (`kan-575-step-2-frontend-one-navigation-graph-rooms`), whose deferred
self-review is where the pattern was named; filed as KAN-721.
