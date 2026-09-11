#!/usr/bin/env python3
"""plan-dispatch-groups.py — the mechanical default for grouping a plan's
dispatch bundles (`plan-dispatch-bundles.py`) into implementer groups,
biased toward fewer, larger groups when bundles aren't independently
parallel (KAN-513).

Rule (canonical definition: this repository's
spectre/changes/kan-513-flow-bias-free-implementer-grouping-toward/design.md,
sections 1-2 — do not restate the reasoning or the worked examples here,
the same Single Source of Truth discipline plan-dispatch-bundles.py's own
docstring already follows). Over bundles `B_k` with members `M_k` (task
ids) and resolved after-set `A_k` (plan-dispatch-bundles.py's own
`after_sets[k-1]`, which already applies the serial default and may name a
bundle's own member when one member's declared `**After:**` names a
fellow bundle-mate):

  deps(k) = A_k \\ M_k — what bundle k actually waits on.

Two passes, both in plan order:

  1. Chain merge. Walk bundles by id. Bundle k joins the first existing
     group g such that deps(k) intersects members(g) and deps(k) is a
     subset of members(g) union ready(g); otherwise it opens a new group.
     Joining never changes ready(g).
  2. Fold to at most two per ready-set. Among the chain-merged groups,
     those sharing an identical ready set are folded down to exactly two,
     alternating in plan order (1st and 3rd together, 2nd and 4th
     together, ...).

Scope is a single file per invocation, matching plan-dispatch-bundles.py's
own scope. `plan-dispatch-groups.sh` is the thin wrapper, one-argument
form only.

Exit codes mirror plan-dispatch-bundles.py exactly, since this script
computes on top of its output:
  0  groups computed — printed one per line on stdout, ordered by each
     group's lowest bundle id, as `group <g>: <bundle ids>` with the ids
     in plan (bundle-id) order. A file with zero unchecked tasks (zero
     bundles) prints nothing and still exits 0.
  1  plan-dispatch-bundles.py's own violations, passed through unchanged.
  2  invocation error — wrong argument count, the file cannot be read, or
     it cannot be decoded as text.

Standard library only (see check-plan-provenance.py's module docstring
for why this repository restricts itself to that).
"""

from __future__ import annotations

import os
import sys
from typing import Dict, FrozenSet, List, Set


def _load_check_file():
    """`plan-dispatch-bundles.py`'s module name contains hyphens, which are
    not a legal Python identifier — import it by file path instead of by
    module name, the same way a sibling script under this directory would
    have to."""
    import importlib.util

    path = os.path.join(os.path.dirname(os.path.realpath(__file__)), "plan-dispatch-bundles.py")
    spec = importlib.util.spec_from_file_location("plan_dispatch_bundles", path)
    module = importlib.util.module_from_spec(spec)
    assert spec.loader is not None
    # dataclasses.dataclass looks the defining module up in sys.modules
    # while processing the class body, so the module must be registered
    # there before exec_module runs it.
    sys.modules[spec.name] = module
    spec.loader.exec_module(module)
    return module.check_file


check_file = _load_check_file()


class _Group:
    __slots__ = ("bundle_ids", "members", "ready")

    def __init__(self, bundle_id: int, members: Set[str], deps: Set[str]) -> None:
        self.bundle_ids: List[int] = [bundle_id]
        self.members: Set[str] = set(members)
        self.ready: Set[str] = set(deps)


def compute_groups(
    bundles: List[List], after_sets: List[List[str]]
) -> List[List[int]]:
    """`bundles[i]` is bundle `i+1`'s member Task list (plan-dispatch-
    bundles.py's own `Task` objects); `after_sets[i]` is that bundle's
    resolved after-set. Returns groups as lists of 1-based bundle ids,
    each list in plan (bundle-id) order, the groups themselves ordered by
    each one's lowest bundle id."""
    groups: List[_Group] = []

    for index, members_tasks in enumerate(bundles):
        bundle_id = index + 1
        members = {t.id for t in members_tasks}
        deps = set(after_sets[index]) - members

        joined = None
        if deps:
            for g in groups:
                if deps & g.members and deps <= (g.members | g.ready):
                    joined = g
                    break
        if joined is not None:
            joined.bundle_ids.append(bundle_id)
            joined.members |= members
        else:
            groups.append(_Group(bundle_id, members, deps))

    # Fold pass: bucket the chain-merged groups by their ready set,
    # preserving first-appearance order within each bucket (which is
    # already plan order, since groups are created/extended in
    # increasing bundle-id order above).
    buckets: Dict[FrozenSet[str], List[_Group]] = {}
    order: List[FrozenSet[str]] = []
    for g in groups:
        key = frozenset(g.ready)
        if key not in buckets:
            buckets[key] = []
            order.append(key)
        buckets[key].append(g)

    folded: List[List[int]] = []
    for key in order:
        bucket = buckets[key]
        if len(bucket) <= 2:
            for g in bucket:
                folded.append(sorted(g.bundle_ids))
            continue
        # Alternate: 1st and 3rd together, 2nd and 4th together, ...
        even = [g for i, g in enumerate(bucket) if i % 2 == 0]
        odd = [g for i, g in enumerate(bucket) if i % 2 == 1]
        for pair in (even, odd):
            if not pair:
                continue
            merged_ids: List[int] = []
            for g in pair:
                merged_ids.extend(g.bundle_ids)
            folded.append(sorted(merged_ids))

    folded.sort(key=lambda ids: min(ids))
    return folded


def main(argv: List[str]) -> int:
    if len(argv) != 2:
        print(f"usage: {argv[0]} <path-to-tasks.md>", file=sys.stderr)
        return 2

    path = argv[1]
    try:
        bundles, after_sets, violations = check_file(path)
    except (OSError, UnicodeDecodeError) as exc:
        print(f"{path}: cannot read file: {exc}", file=sys.stderr)
        return 2

    if violations:
        for violation in violations:
            print(violation, file=sys.stderr)
        return 1

    groups = compute_groups(bundles, after_sets)
    for group_number, bundle_ids in enumerate(groups, start=1):
        ids = " ".join(str(i) for i in bundle_ids)
        print(f"group {group_number}: {ids}")
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv))
