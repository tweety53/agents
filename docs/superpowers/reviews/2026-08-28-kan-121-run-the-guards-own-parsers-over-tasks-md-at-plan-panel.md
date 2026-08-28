# Review panel — kan-121-run-the-guards-own-parsers-over-tasks-md-at-plan

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Code review (low) | blocking | scripts/check-plan-shape.py:310 | tests_opens_none reads only the **Tests:** field's own physical line, while the real parse_task_fields joins continuation lines before applying the same NONE_OPEN_RE. A none written on a continuation line is therefore missed: F6 fires falsely, and F4 is bypassed in exactly the contradiction it exists to catch. This is the two-definitions-of-one-grammar failure the change itself exists to prevent. |
| F2 | Primary | minor | spectre/changes/kan-121-run-the-guards-own-parsers-over-tasks-md-at-plan/tasks.md:106 | Task 2's **Files:** field declares scripts/test-check-plan-shape.sh, but commit c7eaa2d touches only three files and not that one. check_files flags undeclared collateral but never a declared file the diff does not touch, so neither the shipped guard nor the new one can see this class of defect. |
| F3 | Bugbot | severe | scripts/test-check-plan-shape.sh | Surviving mutant. No test invokes the guard through its skills/flow/scripts/ symlink, which is the production path — brainstorm.md step D invokes it bare, and a named guard resolves to <skill-dir>/scripts/<name>. Reverting REPO_ROOT to the naive $SCRIPT_DIR/.. form leaves the whole suite passing while the real symlinked invocation degrades to exit 0 with zero output: a silent false-clean on a guard documented as never silently skipping. |
| F4 | Bugbot | moderate | scripts/test-check-plan-shape.sh | Surviving mutant. No case asserts an exact file:line prefix; every assertion matches only the message substring. An off-by-one or off-by-two in F1's or F3a's file_line = body_start + offset + 1 passes the entire suite. |
| F5 | Bugbot | minor | scripts/test-check-plan-shape.sh | Surviving mutant. design.md states F3b suppresses F1/F2/F4/F6 for that task, so the cause is named rather than the consequence. Making F3b append instead of early-return passes the full suite: case 6 asserts only that the 'never closed' substring is present, never that the other messages are absent. |
| F6 | Bugbot | moderate | scripts/test-check-plan-shape.sh | Surviving mutant introduced by fix round 1. No case exercises a **Tests:** field whose none opens on a continuation line rather than the field's own line, so reverting tests_opens_none to the physical-line form passes the entire suite — nothing pins the behaviour the tests_value plumbing was added to fix. The guard still blocks either way, but reports F6 where F4 is correct: the wrong diagnosis. |
| F7 | Bugbot | minor | scripts/test-check-task-commit-fields.sh | Surviving mutant. NONE_OPEN_RE's word boundary is unexercised: dropping the \b from ^[\s*_]*none\b leaves both suites passing, while a Tests: field opening 'nonetheless this needs `test_alpha` verified' is then read as a none opt-out and declares zero tests. No fixture opens a Tests: value with a none-prefixed word that is not the literal token. NONE_OPEN_RE was introduced by this change, so the gap is this change's own. |

findings-total: 7
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed

reproducers-total: 7
finding-reproducer: F1 printf '%s\n' '- [ ] 1. Do a thing' '' '**Files:** `foo.py`' '**Tests:**' 'none of the tests changed' '**Commit:** `abc123`' > /tmp/t1.md && python3 scripts/check-plan-shape.py /tmp/t1.md   # reports F6; the real parser gives tests==[] via NONE_OPEN_RE
finding-reproducer: F2 git -C <worktree> show c7eaa2d --stat --name-only   # lists 3 files, not 4
finding-reproducer: F3 revert REPO_ROOT to "$(cd "$SCRIPT_DIR/.." && pwd)"; ./scripts/test-check-plan-shape.sh passes; ./skills/flow/scripts/check-plan-shape.sh prints nothing and exits 0
finding-reproducer: F4 sed -i '' 's/file_line = body_start + offset + 1$/file_line = body_start + offset/' scripts/check-plan-shape.py && ./scripts/test-check-plan-shape.sh   # still passes
finding-reproducer: F5 replace F3b's early return with an append; ./scripts/test-check-plan-shape.sh passes while F2 and F6 messages now appear alongside F3b's
finding-reproducer: F6 revert tests_opens_none to NONE_OPEN_RE.match(tests_field.value); ./scripts/test-check-plan-shape.sh still passes; a task with **Files:** `test-a.sh` and none on a Tests: continuation line then reports F6 instead of F4
finding-reproducer: F7 drop the \b from NONE_OPEN_RE; printf -- '- [ ] 1. T\n**Files:** `scripts/test-check-foo.sh`\n**Tests:** nonetheless this needs `test_alpha` verified\n' > tasks.md; python3 scripts/check-plan-shape.py tasks.md   # rc=1 under the mutation, rc=0 with \b
