# Review panel — kan-318-myflow-a-red-green-task-pair-pays-for-context

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | bugbot | important | scripts/plan-dispatch-bundles.py:182 | no test in test-plan-dispatch-bundles.sh exercises an unchecked task carrying a malformed/ungated Squash-with value; dropping the squash.partners is not None guard at line 182 is invisible to all 16 existing cases and causes a TypeError on that exact input |

findings-total: 1
finding-status: F1 fixed

reproducers-total: 1
finding-reproducer: F1 none — see report at .superpowers/sdd/panel-report-0-bugbot.md: mutate line 182 to drop the 'squash.partners is not None' guard, then run printf tasks.md with an unchecked task carrying '**Squash-with:** not-a-valid-id' through scripts/plan-dispatch-bundles.py — raises TypeError: 'NoneType' object is not iterable at compute_bundles line 291, uncaught by any of the 16 existing test cases
