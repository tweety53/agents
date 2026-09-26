package guard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

// Every case of the retired scripts/test-check-task-commit-fields.sh, one
// subtest per ok: label (the bash harness's own text), grouped under its
// case number. Each case builds its own git repository from a copy of one
// base repository, so cases run in parallel and nothing is written outside
// t.TempDir(). The plans live in tcfPlans so TestTaskFieldParseMatchesPython
// can run the still-live Python parser over the very same fixtures.

// bt turns the ¤ placeholder into a backtick: a Go raw string cannot hold one.
func bt(s string) string { return strings.ReplaceAll(s, "¤", "`") }

var tcfPlans = map[string]string{
	"1": bt(`- [ ] 1. Clean task

**Files:** ¤alpha.txt¤, ¤beta.txt¤
**Tests:** ¤test_alpha¤
**Commit:** add alpha and beta
**Build:** green
`),
	"2": bt(`- [ ] 2. Undeclared file

**Files:** ¤alpha.txt¤
**Tests:** ¤test_alpha¤
**Commit:** add alpha only
**Build:** green
`),
	"3": bt(`- [ ] 3. Test present

**Files:** ¤alpha.txt¤
**Tests:** ¤test_alpha¤
**Commit:** add alpha
**Build:** green
`),
	"4": bt(`- [ ] 4. Test missing

**Files:** ¤alpha.txt¤
**Tests:** ¤test_alpha¤
**Commit:** add alpha
**Build:** green
`),
	"5": bt(`- [ ] 5. Subject matches

**Files:** ¤alpha.txt¤
**Tests:** ¤test_alpha¤
**Commit:** add alpha for real
**Build:** green
`),
	"6": bt(`- [ ] 6. Subject mismatch

**Files:** ¤alpha.txt¤
**Tests:** ¤test_alpha¤
**Commit:** add alpha for real
**Build:** green
`),
	"7": bt(`- [ ] 7. Collateral covered

**Files:** ¤alpha.txt¤
**Allowed-collateral:** ¤docs/*.md¤
**Tests:** ¤test_alpha¤
**Commit:** add alpha and sweep docs
**Build:** green
`),
	"8": bt(`- [ ] 8. Prose cases present

**Files:** ¤guard_test.sh¤
**Tests:** Case 1: files subset of declared passes; Case 2: undeclared file
fails
**Commit:** add guard test cases
**Build:** green
`),
	"9": bt(`- [ ] 9. Prose case missing

**Files:** ¤guard_test.sh¤
**Tests:** Case 1: files subset of declared passes; Case 2: undeclared file
fails
**Commit:** add guard test case one only
**Build:** green
`),
	"10": bt(`- [ ] 10. No checkable tests declared

**Files:** ¤guard_test.sh¤
**Tests:** the 7 cases listed in task 3.2, run for the first time against
the wrapper this task adds
**Commit:** add the wrapper
**Build:** green
`),
	"15": bt(`- [ ] 15. Fence guard

**Files:** ¤alpha.txt¤

Example of the field grammar, not a real field:

¤¤¤
**Files:** ¤evil.txt¤
¤¤¤

**Tests:** ¤test_alpha¤
**Commit:** add alpha
**Build:** green
`),
	"17": bt(`- [ ] 17. Change-name scope

**Files:** ¤alpha.txt¤
**Commit:** feat(kan-900-some-change): add alpha
**Build:** green
`),
	"18": bt(`- [ ] 18. Bare-key scope

**Files:** ¤alpha.txt¤
**Commit:** feat(kan-900): add alpha
**Build:** green
`),
	"19": bt(`- [ ] 19. Numeric task id scope

**Files:** ¤alpha.txt¤
**Commit:** feat(3): add alpha
**Build:** green
`),
	"20": bt(`- [ ] 20. Dotted task id scope

**Files:** ¤alpha.txt¤
**Commit:** feat(3.2): add alpha
**Build:** green
`),
	"21": bt(`- [ ] 21. Module scope

**Files:** ¤alpha.txt¤
**Commit:** feat(scripts): add alpha
**Build:** green
`),
	"22": bt(`- [ ] 22. No scope

**Files:** ¤alpha.txt¤
**Commit:** feat: add alpha
**Build:** green
`),
	"23": bt(`- [ ] 23. Scope containing the key

**Files:** ¤alpha.txt¤
**Commit:** feat(kan-900-helpers): add alpha
**Build:** green
`),
	"24": bt(`- [ ] 24. Scope containing the change name

**Files:** ¤alpha.txt¤
**Commit:** feat(kan-900-some-change-helpers): add alpha
**Build:** green
`),
	"25": bt(`- [ ] 25. Breaking-change form names the change

**Files:** ¤alpha.txt¤
**Commit:** feat(kan-900-some-change)!: add alpha
**Build:** green
`),
	"26": bt(`- [ ] 26. Breaking-change form, real module

**Files:** ¤alpha.txt¤
**Commit:** feat(scripts)!: add alpha
**Build:** green
`),
	"27": bt(`- [ ] 27. Uppercase change-name scope

**Files:** ¤alpha.txt¤
**Commit:** feat(KAN-900-SOME-CHANGE): add alpha
**Build:** green
`),
	"28": bt(`- [ ] 28. Uppercase Jira-key scope

**Files:** ¤alpha.txt¤
**Commit:** feat(KAN-900): add alpha
**Build:** green
`),
	"29": bt(`- [ ] 29. Ambiguous key-shaped change name

**Files:** ¤alpha.txt¤
**Commit:** feat(release-2026): add alpha
**Build:** green
`),
	"30": bt(`- [ ] 1. Red half

**Files:** ¤alpha.txt¤
**Tests:** ¤test_alpha¤
**Commit:** test: add alpha
**Build:** red

**Squash-with:** Task 2

- [ ] 2. Green half

**Files:** ¤beta.txt¤
**Commit:** feat: add alpha and beta
**Build:** green
`),
	"32": bt(`- [ ] 1. Red half pointing at nothing

**Files:** ¤alpha.txt¤
**Commit:** test: add alpha
**Build:** red

**Squash-with:** Task 9

- [ ] 2. Green half

**Files:** ¤beta.txt¤
**Commit:** feat: add alpha and beta
**Build:** green
`),
	"33": bt(`- [ ] 1. Red half

**Files:** ¤alpha.txt¤
**Commit:** test: add alpha
**Build:** red

**Squash-with:** Task 2

- [ ] 2. Also red

**Files:** ¤beta.txt¤
**Commit:** feat: add alpha and beta
**Build:** red

**Squash-with:** Task 1
`),
	"34": bt(`- [ ] 1. Ordinary task

**Files:** ¤alpha.txt¤
**Commit:** feat: add alpha
**Build:** green

- [ ] 2. Another ordinary task

**Files:** ¤beta.txt¤
**Commit:** feat: add beta
**Build:** green
`),
	"36": bt(`- [ ] 1. Red third

**Files:** ¤alpha.txt¤
**Commit:** test: add alpha
**Build:** red

**Squash-with:** Task 2, 3

- [ ] 2. Green third

**Files:** ¤beta.txt¤
**Commit:** feat: add alpha beta and gamma
**Build:** green

- [ ] 3. Sibling green third

**Files:** ¤gamma.txt¤
**Commit:** feat: add alpha beta and gamma
**Build:** green
`),
	"37": bt(`- [ ] 1. Red third

**Files:** ¤alpha.txt¤
**Commit:** test: add alpha
**Build:** red

**Squash-with:** Task 2, 3

- [ ] 2. Green third

**Files:** ¤beta.txt¤
**Commit:** feat: add beta
**Build:** green

- [ ] 3. Sibling green third

**Files:** ¤gamma.txt¤
**Commit:** feat: add alpha beta and gamma
**Build:** green
`),
	"38": bt(`- [ ] 1. Red third

**Files:** ¤alpha.txt¤
**Commit:** test: add alpha
**Build:** red

**Squash-with:** Task 2, 9

- [ ] 2. Green third

**Files:** ¤beta.txt¤
**Commit:** feat: add alpha and beta
**Build:** green
`),
	"39": bt(`- [ ] 1. Red third

**Files:** ¤alpha.txt¤
**Commit:** test: add alpha
**Build:** red

**Squash-with:** Task 2, 3

- [ ] 2. Green third

**Files:** ¤beta.txt¤
**Commit:** feat: add alpha and beta
**Build:** green

- [ ] 3. Also red

**Files:** ¤gamma.txt¤
**Commit:** test: add gamma
**Build:** red

**Squash-with:** Task 1
`),
	"40": bt(`- [ ] 1. Red third

**Files:** ¤alpha.txt¤
**Commit:** test: add alpha
**Build:** red

**Squash-with:** Task 2, 3

- [ ] 2. Green third

**Files:** ¤beta.txt¤
**Commit:** feat: add alpha beta and gamma
**Build:** green

- [ ] 3. Sibling green third declaring no subject

**Files:** ¤gamma.txt¤
**Build:** green
`),
	"42": bt(`- [ ] 1. Red half

**Files:** ¤alpha.txt¤
**Commit:** test: add alpha
**Build:** red

**Squash-with:** Task 2

- [ ] 2. Green half declaring no subject

**Files:** ¤beta.txt¤
**Build:** green
`),
	"43": bt(`- [ ] 1. Ordinary task declaring no subject

**Files:** ¤alpha.txt¤
**Build:** green
`),
	"44": bt(`- [ ] 1. Red one

**Files:** ¤one.txt¤
**Commit:** test: add one
**Build:** red

**Squash-with:** Task 3, 4

- [ ] 2. Red two

**Files:** ¤two.txt¤
**Commit:** test: add two
**Build:** red

**Squash-with:** Task 3, 5

- [ ] 3. Shared green partner

**Files:** ¤shared.txt¤
**Build:** green

- [ ] 4. Green partner of red one

**Files:** ¤four.txt¤
**Commit:** feat: subject a
**Build:** green

- [ ] 5. Green partner of red two

**Files:** ¤five.txt¤
**Commit:** feat: subject b
**Build:** green
`),
	"45": bt(`- [ ] 1. Red one

**Files:** ¤one.txt¤
**Commit:** test: add one
**Build:** red

**Squash-with:** Task 3, 4

- [ ] 2. Red two

**Files:** ¤two.txt¤
**Commit:** test: add two
**Build:** red

**Squash-with:** Task 3, 5

- [ ] 3. Shared green partner

**Files:** ¤shared.txt¤
**Build:** green

- [ ] 4. Green partner of red one

**Files:** ¤four.txt¤
**Commit:** feat: the one folded subject
**Build:** green

- [ ] 5. Green partner of red two

**Files:** ¤five.txt¤
**Commit:** feat: the one folded subject
**Build:** green
`),
	"46a": bt(`- [ ] 1. Red half naming a missing partner

**Files:** ¤alpha.txt¤
**Commit:** test: add alpha
**Build:** red

**Squash-with:** Task 9
`),
	"46b": bt(`- [ ] 1. Red half naming a red partner

**Files:** ¤alpha.txt¤
**Commit:** test: add alpha
**Build:** red

**Squash-with:** Task 2

- [ ] 2. Partner that is itself red and declares no subject

**Files:** ¤beta.txt¤
**Build:** red
`),
	"47": bt(`- [ ] 1. Red task naming a missing partner and two disagreeing ones

**Files:** ¤alpha.txt¤
**Commit:** test: add alpha
**Build:** red

**Squash-with:** Task 2, 3, 9

- [ ] 2. Green partner declaring one subject

**Files:** ¤beta.txt¤
**Commit:** feat: subject a
**Build:** green

- [ ] 3. Green partner declaring a different subject

**Files:** ¤gamma.txt¤
**Commit:** feat: subject b
**Build:** green
`),
	"48a": bt(`- [ ] 1. Red one, whose partners declare no subject

**Files:** ¤one.txt¤
**Commit:** test: add one
**Build:** red

**Squash-with:** Task 3, 4

- [ ] 2. Red two, whose partner declares the fold subject

**Files:** ¤two.txt¤
**Commit:** test: add two
**Build:** red

**Squash-with:** Task 3, 5

- [ ] 3. Shared green partner declaring no subject

**Files:** ¤shared.txt¤
**Build:** green

- [ ] 4. Green partner of red one declaring no subject

**Files:** ¤four.txt¤
**Build:** green

- [ ] 5. Green partner of red two declaring the fold subject

**Files:** ¤five.txt¤
**Commit:** feat: the one folded subject
**Build:** green
`),
	"48b": bt(`- [ ] 1. Red one, whose partners declare no subject

**Files:** ¤one.txt¤
**Commit:** test: add one
**Build:** red

**Squash-with:** Task 3, 4

- [ ] 2. Red two, whose partners declare no subject either

**Files:** ¤two.txt¤
**Commit:** test: add two
**Build:** red

**Squash-with:** Task 3, 5

- [ ] 3. Shared green partner declaring no subject

**Files:** ¤shared.txt¤
**Build:** green

- [ ] 4. Green partner of red one declaring no subject

**Files:** ¤four.txt¤
**Build:** green

- [ ] 5. Green partner of red two declaring no subject

**Files:** ¤five.txt¤
**Build:** green
`),
	"49": bt(`- [ ] 1. Red task whose Squash-with carries free text

**Files:** ¤a.txt¤
**Commit:** test: add a
**Build:** red

**Squash-with:** Task 3 (see step 2)

- [ ] 2. An unrelated green task nobody folds with

**Files:** ¤unrelated.txt¤
**Build:** green

- [ ] 3. The real partner

**Files:** ¤b.txt¤
**Commit:** feat: add a and b
**Build:** green
`),
	"50": bt(`- [ ] 1. Red task naming a partner and mentioning a step number

**Files:** ¤alpha.txt¤
**Commit:** test: add alpha
**Build:** red

**Squash-with:** Task 2 (see step 3)

- [ ] 2. Green partner

**Files:** ¤beta.txt¤
**Commit:** feat: add alpha and beta
**Build:** green
`),
	"51": bt(`- [ ] 1. Red task with two partners

**Files:** ¤alpha.txt¤
**Commit:** test: add alpha
**Build:** red

**Squash-with:** Task 2 3

- [ ] 2. Green partner

**Files:** ¤beta.txt¤
**Commit:** feat: add alpha beta and gamma
**Build:** green

- [ ] 3. Sibling green partner

**Files:** ¤gamma.txt¤
**Commit:** feat: add alpha beta and gamma
**Build:** green
`),
	"52": bt(`- [ ] 1. Red task whose Squash-with is followed by prose

**Files:** ¤alpha.txt¤
**Commit:** test: add alpha
**Build:** red

**Squash-with:** Task 2
The fold is described in the paragraph above.

- [ ] 2. Green partner

**Files:** ¤beta.txt¤
**Commit:** feat: add alpha and beta
**Build:** green
`),
	"53": bt(`- [ ] 1. Red task whose partner ids wrap onto a second line

**Files:** ¤alpha.txt¤
**Commit:** test: add alpha
**Build:** red

**Squash-with:** Task 2,
3

- [ ] 2. Green partner

**Files:** ¤beta.txt¤
**Commit:** feat: add alpha beta and gamma
**Build:** green

- [ ] 3. Sibling green partner named only on the wrapped line

**Files:** ¤gamma.txt¤
**Commit:** feat: add alpha beta and gamma
**Build:** green
`),
	"54": bt(`- [ ] 1. A red task naming a partner that is itself red

**Files:** ¤alpha.txt¤
**Commit:** test: add alpha
**Build:** red

**Squash-with:** Task 9

- [ ] 9. A red task whose own partner is green

**Files:** ¤beta.txt¤
**Commit:** test: add beta
**Build:** red

**Squash-with:** Task 10

- [ ] 10. The green partner of task 9 fold

**Files:** ¤gamma.txt¤
**Commit:** feat: add beta and gamma
**Build:** green
`),
	"55": bt(`- [ ] 1. Red task whose Squash-with names its partner in free text

**Files:** ¤alpha.txt¤
**Commit:** test: add alpha
**Build:** red

**Squash-with:** Task 2 (see step 3)

- [ ] 2. The green partner that free text names

**Files:** ¤beta.txt¤
**Commit:** feat: add alpha and beta
**Build:** green
`),
	"57": bt(`- [ ] 1. Red task whose first Squash-with line does not gate

**Files:** ¤alpha.txt¤
**Commit:** test: add alpha
**Build:** red

**Squash-with:** Task 2 (see note below)
**Squash-with:** Task 2

- [ ] 2. The green partner named by the gating line

**Files:** ¤beta.txt¤
**Commit:** feat: add alpha and beta
**Build:** green
`),
	"58": bt(`- [ ] 1. Red task carrying two Build lines

**Files:** ¤alpha.txt¤
**Commit:** test: add alpha
**Build:** red
**Build:** green

**Squash-with:** Task 9
`),
	"59": bt(`- [ ] 1. Red task whose tag line is followed by prose

**Files:** ¤alpha.txt¤
**Commit:** test: add alpha
**Build:** red
this sentence explains the tag and is not part of it

**Squash-with:** Task 9
`),
	"60": bt(`- [ ] 1. Real task with a worked example in its body

**Files:** ¤alpha.txt¤
**Commit:** test: add alpha
**Build:** green

Example of a fold, shown but never declared:

¤¤¤
- [ ] 9. Example red task
**Build:** red
**Squash-with:** Task 8 (see the note)
¤¤¤
`),
	"61": bt(`- [ ] 1. First task line for this id

**Files:** ¤alpha.txt¤
**Commit:** test: add alpha
**Build:** green

- [ ] 1. Second task line reusing the id

**Files:** ¤zulu.txt¤
**Commit:** test: add zulu
**Build:** red

- [ ] 3. Red task folding into Task 1

**Files:** ¤gamma.txt¤
**Commit:** test: add gamma
**Build:** red

**Squash-with:** Task 1
`),
	"62": bt(`- [ ] 1. Real task with a tilde-fenced example field

**Files:** ¤real.txt¤
**Commit:** test: add real
**Build:** green

Example of how the field is written:

~~~markdown
**Files:** ¤example-only.txt¤
~~~
`),
	"63": bt(`- [ ] 1. Real task with an indented fenced example field

**Files:** ¤real.txt¤
**Commit:** test: add real
**Build:** green

Example of how the field is written:

   ¤¤¤markdown
**Files:** ¤example-only.txt¤
   ¤¤¤
`),
	"64a": bt(`- [ ] 1. Task whose body opens a fence it never closes

~~~

**Build:** green
**Files:** ¤alpha.txt¤
**Commit:** test: add alpha
`),
	"64b": bt(`- [ ] 1. Task whose body closes the fence it opens

**Build:** green
**Files:** ¤alpha.txt¤
**Commit:** test: add alpha

~~~
example
~~~
`),
	"65": bt(`- [ ] 1. Parent task, already done

**Files:** ¤parent-only.txt¤
**Tests:** ¤test_parent¤
**Commit:** test: parent
**Build:** green
`),
	"65-fix-1": bt(`- [ ] 1. Fix-round task

**Files:** ¤alpha.txt¤, ¤beta.txt¤
**Tests:** ¤test_alpha¤
**Commit:** fix: add alpha
**Build:** green
`),
	"66": bt(`- [ ] 1. Parent task

**Files:** ¤parent-only.txt¤
**Tests:** ¤test_parent¤
**Commit:** test: parent
**Build:** green
`),
	"66-fix-1": bt(`- [ ] 1. First fix round

**Files:** ¤fix-one-only.txt¤
**Tests:** ¤test_one¤
**Commit:** fix: one
**Build:** green
`),
	"66-fix-2": bt(`- [ ] 1. Second fix round

**Files:** ¤alpha.txt¤, ¤beta.txt¤
**Tests:** ¤test_alpha¤
**Commit:** fix: add alpha
**Build:** green
`),
	"67": bt(`- [ ] 1. Task

**Files:** ¤alpha.txt¤
**Tests:** ¤test_alpha¤
**Commit:** test: alpha
**Build:** green
`),
	"67-other": bt(`- [ ] 1. Another change entirely

**Files:** ¤beta.txt¤
**Tests:** ¤test_beta¤
**Commit:** test: beta
**Build:** green
`),
	"68": bt(`- [ ] 68. None opening with backticks

**Files:** ¤alpha.txt¤
**Tests:** none added — verified via the ¤scripts/check-references.sh¤ guard
**Commit:** test: alpha
**Build:** green
`),
	"69": bt(`- [ ] 69. None opening with a Case N label

**Files:** ¤alpha.txt¤
**Tests:** none — see Case 1 for context
**Commit:** test: alpha
**Build:** green
`),
	"70": bt(`- [ ] 70. None mid-sentence is unaffected

**Files:** ¤alpha.txt¤
**Tests:** added ¤test_alpha¤, none of the other paths change
**Commit:** test: alpha
**Build:** green
`),
	"71": bt(`- [ ] 71. A none-prefixed word is not the token

**Files:** ¤alpha.txt¤
**Tests:** nonetheless this needs ¤test_alpha¤ verified
**Commit:** test: alpha
**Build:** green
`),
	"72": bt(`- [ ] 72. Satellite resolves through the canonical worktree

**Files:** ¤alpha.txt¤
**Tests:** ¤test_alpha¤
**Commit:** add alpha for real
**Build:** green
`),
	"73": bt(`- [ ] 73. Root task beside a satellite

**Files:** ¤alpha.txt¤
**Tests:** ¤test_alpha¤
**Commit:** add alpha for real
**Build:** green
`),
	"75": bt(`- [ ] 75. Would-be resolved task

**Files:** ¤alpha.txt¤
**Commit:** add alpha
**Build:** green
`),
	"76": bt(`- [ ] 76. Satellite beside a pre-plan change resolves

**Files:** ¤alpha.txt¤
**Tests:** ¤test_alpha¤
**Commit:** add alpha for real
**Build:** green
`),
	"77": bt(`- [ ] 1. Named change task

**Files:** ¤alpha.txt¤
**Tests:** ¤test_alpha¤
**Commit:** test: alpha
**Build:** green
`),
	"77-other": bt(`- [ ] 1. Unrelated change task

**Files:** ¤beta.txt¤
**Commit:** test: beta
**Build:** green
`),
	"78": bt(`- [ ] 1. Parent task, already done

**Files:** ¤parent-only.txt¤
**Commit:** test: parent
**Build:** green
`),
	"78-fix-1": bt(`- [ ] 1. Fix-round task

**Files:** ¤alpha.txt¤
**Tests:** ¤test_alpha¤
**Commit:** fix: add alpha
**Build:** green
`),
	"78-third": bt(`- [ ] 1. Unrelated

**Files:** ¤gamma.txt¤
**Commit:** test: gamma
**Build:** green
`),
	"80": bt(`- [ ] 1. Satellite task named directly

**Files:** ¤alpha.txt¤
**Tests:** ¤test_alpha¤
**Commit:** add alpha for real
**Build:** green
`),
	"82": bt(`- [ ] 1. Clean task

**Files:** ¤alpha.txt¤
**Commit:** test: alpha
**Build:** green
`),
	"87": bt(`- [ ] 1. Red task whose tag carries trailing prose

**Files:** ¤alpha.txt¤
**Commit:** test: add alpha
**Build:** red — lands with its partner

**Squash-with:** Task 9
`),
	"88": bt(`- [ ] 1. Red half

**Files:** ¤alpha.txt¤
**Tests:** ¤test_alpha¤
**Commit:** test: add alpha
**Build:** red

**Squash-with:** Task 2

- [ ] 2. Half with a malformed tag

**Files:** ¤beta.txt¤
**Commit:** feat: add alpha and beta
**Build:** yellow
`),
	"89": bt(`- [ ] 1. Read-only guard

**Files:** ¤alpha.txt¤
**Tests:** ¤test_alpha¤
**Regression:** test_alpha fails if alpha.txt is reverted
**Baseline:** before=0 after=1
**Commit:** feat: add alpha
**Build:** green
`),
	"90a": bt(`- [ ] 1. Task one

**Files:**
- Modify: ¤alpha.txt¤
**After:** Task 2

**Tests:** ¤test_alpha¤
**Commit:** add alpha
**Build:** green
`),
	"90b": bt(`- [ ] 1. Task one

**Files:** alpha.txt
**After:** Task 2
**Tests:** ¤test_alpha¤
**Commit:** add alpha
**Build:** green
`),
	"91": bt(`- [ ] 1. Plan lives in the canonical repo only

**Files:** ¤alpha.txt¤
**Tests:** ¤test_alpha¤
**Commit:** add alpha for real
**Build:** green
`),
	"92": bt(`- [ ] 1. Any task

**Files:** ¤alpha.txt¤
**Tests:** ¤test_alpha¤
**Commit:** add alpha
**Build:** green
`),
	"94": bt(`- [ ] 1. Undeclared file

**Files:** ¤alpha.txt¤
**Tests:** ¤test_alpha¤
**Commit:** add alpha only
**Build:** green
`),
	"95": bt(`- [ ] 1. Documented call shape

**Files:** ¤alpha.txt¤
**Tests:** ¤test_alpha¤
**Commit:** add alpha
**Build:** green
`),
	"96": bt(`- [ ] 1. Baseline delta matches

**Files:** ¤alpha.txt¤
**Tests:** ¤test_alpha¤
**Baseline:** before=1 after=2
**Commit:** grow alpha tests
**Build:** green
`),
	"97": bt(`- [ ] 1. Baseline delta stale

**Files:** ¤alpha.txt¤
**Tests:** ¤test_alpha¤
**Baseline:** before=1 after=3
**Commit:** grow alpha tests
**Build:** green
`),
	"98": bt(`- [ ] 1. Baseline unit is not tests

**Files:** ¤alpha.txt¤
**Tests:** ¤test_alpha¤
**Baseline:** before=37 after=38
**Commit:** bump harness count
**Build:** green
`),
	"99": bt(`- [ ] 1. Removed test still declared

**Files:** ¤alpha.txt¤
**Tests:** ¤test_gone¤
**Commit:** remove gone test
**Build:** green
`),
	"100": bt(`- [ ] 1. Bare camelCase name stale

**Files:** ¤alpha.txt¤
**Tests:** covers calculateDailyCaloriesReturnsZero for empty input
**Commit:** add alpha
**Build:** green
`),
	"101": bt(`- [ ] 1. Names present in tree

**Files:** ¤alpha.txt¤
**Tests:** ¤test_alpha¤ and ¤test_beta¤
**Commit:** add alpha
**Build:** green
`),
	"102": bt(`- [ ] 1. None opens

**Files:** ¤alpha.txt¤
**Tests:** none — GitLabOnlyToken stays unexported, no test added
**Commit:** add alpha
**Build:** green
`),
	"103": bt(`- [ ] 1. Path token in Tests

**Files:** ¤helper.md¤
**Tests:** ¤helper.md¤
**Commit:** touch helper
**Build:** green
`),
	"104": bt(`- [ ] 1. Stale name in another plan

**Files:** ¤alpha.txt¤
**Tests:** ¤test_stale¤
**Commit:** remove stale test
**Build:** green
`),
	"104-archive": bt("- [ ] 1. Old plan task\n\n**Files:** ¤gone.txt¤\n**Tests:** ¤test_stale¤\n**Commit:** old\n"),
	"105": bt(`- [ ] 1. Lifecycle annotations are not tests

**Files:** ¤alpha.txt¤
**Tests:** ¤test_alpha¤
**Baseline:** before=1 after=2
**Commit:** grow alpha tests
**Build:** green
`),
	"107-b": bt(`- [ ] 1. Ambiguous B

**Files:** ¤beta.txt¤
**Tests:** ¤test_beta¤
**Commit:** add alpha for real
**Build:** green
`),
	"108": bt(`- [ ] 1. Prose field naming a script

**Files:** ¤alpha.txt¤
**Tests:** covered by the existing guards — ¤check-references.sh¤ and ¤check-contract-budget.sh¤
**Commit:** add alpha
**Build:** green
`),
	"109": bt(`- [ ] 1. Genuine missing test

**Files:** ¤alpha.txt¤
**Tests:** ¤test_alpha¤
**Commit:** add alpha
**Build:** green
`),
	"110": bt(`- [ ] 1. Labelled field missing its label

**Files:** ¤alpha.txt¤
**Tests:** Case 99 covers the parsing (see ¤helper_alpha¤)
**Commit:** add alpha
**Build:** green
`),
	"111": bt(`- [ ] 1. Bare camelCase tree miss

**Files:** ¤alpha.txt¤
**Tests:** ¤test_alpha¤ and helperAlpha
**Commit:** add alpha
**Build:** green
`),
	"112": tcfCounterPlan("Measured baseline matches", "before=5 after=6", "<!-- measured: cat counter.txt @ branch main -->\n"),
	"113": tcfCounterPlan("Measured baseline after is stale", "before=5 after=9", "<!-- measured: cat counter.txt @ branch main -->\n"),
	"114": tcfCounterPlan("Measured baseline before is stale", "before=4 after=6", "<!-- measured: cat counter.txt @ branch main -->\n"),
	"115": tcfCounterPlan("Unmeasured baseline skips", "before=37 after=38", ""),
	"116": tcfCounterPlan("Failing measurement skips", "before=5 after=6", "<!-- measured: false @ branch main -->\n"),
	"117": tcfCounterPlan("Non-integer measurement skips", "before=5 after=6", "<!-- measured: printf 'five' @ branch main -->\n"),
	"118": tcfCounterPlan("Ambiguous measurement skips", "before=5 after=6", "<!-- measured: cat counter.txt @ branch main -->\n<!-- measured: cat counter.txt | wc -c @ branch main -->\n"),
	"119": tcfCounterPlan("One command across two refs verifies", "before=5 after=6", "<!-- measured: cat counter.txt @ branch main -->\n<!-- measured: cat counter.txt @ branch other -->\n"),
	"120": tcfCounterPlan("Measured run leaves the worktree alone", "before=5 after=6", "<!-- measured: cat counter.txt @ branch main -->\n"),
	"121": tcfCounterPlan("Pipeline measurement", "before=1 after=1", "<!-- measured: cat counter.txt | wc -l @ branch main -->\n"),
	"122": tcfCounterPlan("At-bearing measurement", "before=1 after=99", "<!-- measured: grep -o '@' counter.txt | wc -l @ branch main -->\n"),
	"123": bt(`- [ ] 1. Measured command supersedes the static count

**Files:** ¤alpha.txt¤, ¤counter.txt¤
**Tests:** none — a counter file, not a test
**Baseline:** before=5 after=6
<!-- measured: cat counter.txt @ branch main -->
**Commit:** bump counter
**Build:** green
`),
	"125": bt(`- [ ] 1. Add alpha

**Files:** ¤alpha.txt¤
**Tests:** none — a plain file
**Commit:** add alpha

- [ ] 2. Add beta

**Files:** ¤beta.txt¤
**Tests:** none — a plain file
**Commit:** add beta
`),
	"130": bt(`- [ ] 130. Declared file untouched

**Files:** ¤alpha.txt¤, ¤ghost.txt¤
**Tests:** ¤test_alpha¤
**Commit:** add alpha only
**Build:** green
`),
	"131": bt(`- [ ] 131. Baseline drift both ways

**Files:** ¤j4-bar-one-face-darwin.png¤
**Tests:** ¤test_bar¤
**Commit:** swap the bar baseline
**Build:** green
`),
	"132": bt(`- [ ] 1. Red half

**Files:** ¤alpha.txt¤
**Tests:** ¤test_alpha¤
**Commit:** test: add alpha
**Build:** red

**Squash-with:** Task 2

- [ ] 2. Green half

**Files:** ¤beta.txt¤, ¤ghost.txt¤
**Commit:** feat: add alpha and beta
**Build:** green
`),
	"133": bt(`- [ ] 133. Deletion is a touch

**Files:** ¤alpha.txt¤, ¤beta.txt¤
**Tests:** ¤test_alpha¤
**Commit:** drop beta keep alpha
**Build:** green
`),
	"134": bt(`- [ ] 134. Decision glued to the commit subject

**Files:** ¤alpha.txt¤
**Tests:** none — a plain file
**Commit:** fix: widen alpha
**Decision:** widen-not-rename
**Build:** green
`),
	"135": bt(`Paths are relative to it; ¤gs¤ abbreviates
¤src/main/kotlin/com/gymie/gs¤, ¤gsTest¤
¤src/test/kotlin/com/gymie/gs¤.

- [ ] 135. Shorthand task

**Files:** ¤gs/core/Thing.kt¤, ¤gsTest/core/ThingTest.kt¤
**Tests:** none — a plain file
**Commit:** feat: add thing
**Build:** green
`),
	"136": bt(`- [ ] 136. No legend behind the abbreviation

**Files:** ¤gs/core/Thing.kt¤
**Tests:** none — a plain file
**Commit:** feat: add thing
**Build:** green
`),
	"137": bt(`- [ ] 137. Wrapped sentence

**Files:** ¤alpha.txt¤
**Tests:** ¤the guard folds wrapped lines before matching¤
**Commit:** add wrapped sentence
**Build:** green
`),
	"138": bt(`- [ ] 138. Absent sentence

**Files:** ¤alpha.txt¤
**Tests:** ¤a sentence nowhere in this commit¤
**Commit:** add alpha only
**Build:** green
`),
	"139": bt(`- [ ] 139. Export-ignored sentence

**Files:** ¤packed.txt¤, ¤.gitattributes¤
**Tests:** ¤the tree search sees every searchable blob¤
**Commit:** add export-ignored sentence
**Build:** green
`),
	"140": bt(`- [ ] 140. Evidence-free verified tag

Run the check:

¤¤¤bash verified:
echo hi
¤¤¤

**Files:** ¤alpha.txt¤
**Tests:** ¤test_alpha¤
**Commit:** add alpha
**Build:** green
`),
	"141": bt(`- [ ] 141. Evidence-free measured comment

Baseline: 197 tests
<!-- measured: -->

**Files:** ¤alpha.txt¤
**Tests:** ¤test_alpha¤
**Commit:** add alpha
**Build:** green
`),
	"142": bt(`- [ ] 142. Evidence carried

¤¤¤bash verified:ran it locally before writing
echo hi
¤¤¤

Baseline: 197 tests
<!-- measured: wc -l on the machine-local log; no ref, the file is machine-local -->

**Files:** ¤alpha.txt¤
**Tests:** ¤test_alpha¤
**Commit:** add alpha
**Build:** green
`),
	"143": bt(`- [ ] 143. Comment inside fence content

¤¤¤bash verified:ran it locally
echo "<!-- measured: -->"
¤¤¤

**Files:** ¤alpha.txt¤
**Tests:** ¤test_alpha¤
**Commit:** add alpha
**Build:** green
`),
	"144": bt(`- [ ] 144. Language-less fence with a bare tag

¤¤¤verified:
echo hi
¤¤¤

**Files:** ¤alpha.txt¤
**Tests:** ¤test_alpha¤
**Commit:** add alpha
**Build:** green
`),
	"145": bt(`- [ ] 145. Empty unverified fence

¤¤¤unverified:
echo hi
¤¤¤

**Files:** ¤alpha.txt¤
**Tests:** ¤test_alpha¤
**Commit:** add alpha
**Build:** green
`),
	"146": bt(`- [ ] 146. Empty predicted comment

After the change: 186 tests
<!-- predicted: -->

**Files:** ¤alpha.txt¤
**Tests:** ¤test_alpha¤
**Commit:** add alpha
**Build:** green
`),
}

// tcfCounterPlan is the one-task counter.txt plan cases 112-122 share,
// differing only in title, declared counts and recorded measured: lines.
func tcfCounterPlan(title, counts, measured string) string {
	return bt("- [ ] 1. " + title + "\n\n**Files:** ¤counter.txt¤\n**Tests:** none — a counter file, not a test\n**Baseline:** " +
		counts + "\n" + measured + "**Commit:** bump counter\n**Build:** green\n")
}

const tcfNoSubject = "no partner named by Squash-with: declares a Commit: subject"

// tcfFixture is built once per top-level test: a base repository with one
// root commit that every case copies, and the inert `flow` stub the bash
// harness put first on PATH so the change-plan store step never reaches a
// real flow CLI.
type tcfFixture struct{ base, inert string }

func tcfNewFixture(t *testing.T) *tcfFixture {
	t.Helper()
	dir := t.TempDir()
	fx := &tcfFixture{base: dir + "/base", inert: dir + "/flow-inert"}
	writeExec(t, fx.inert+"/flow", "#!/usr/bin/env bash\nexit 1\n")
	mkdir(t, fx.base)
	tcfGit(t, fx.base, "init", "-q")
	tcfGit(t, fx.base, "config", "user.email", "test@example.com")
	tcfGit(t, fx.base, "config", "user.name", "Test")
	writeFile(t, fx.base+"/root.txt", "root\n")
	tcfGit(t, fx.base, "add", "root.txt")
	tcfGit(t, fx.base, "commit", "-q", "-m", "root")
	return fx
}

// tcfGit runs git -C dir under fixtureGitEnv, returning
// stdout without its trailing newline.
func tcfGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), fixtureGitEnv...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, stderr.String())
	}
	return strings.TrimRight(string(out), "\n")
}

type tcfRes struct {
	rc  int
	out string // stdout and stderr interleaved, as the harness's 2>&1 read it
}

// run invokes the registered guard in-process with the inert flow stub first
// on PATH.
func (fx *tcfFixture) run(args ...string) tcfRes {
	return tcfRunWith(pathEnv(fx.inert), args...)
}

func tcfRunWith(getenv func(string) string, args ...string) tcfRes {
	var out bytes.Buffer
	rc := Registry["check-task-commit-fields"](args, Env{Getenv: getenv, Dir: os.TempDir()}, &out, &out)
	return tcfRes{rc, out.String()}
}

type tcfRepo struct {
	t           *testing.T
	fx          *tcfFixture
	dir, change string
}

// repo is the harness's new_repo: a fresh copy of the base repository with
// an empty spectre/changes/<change> directory.
func (fx *tcfFixture) repo(t *testing.T, change string) *tcfRepo {
	t.Helper()
	dir := t.TempDir() + "/repo"
	gdcCopyTree(t, fx.base, dir)
	mkdir(t, dir+"/spectre/changes/"+change)
	return &tcfRepo{t, fx, dir, change}
}

func (r *tcfRepo) git(args ...string) string { return tcfGit(r.t, r.dir, args...) }

func (r *tcfRepo) writePlan(body string) {
	writeFile(r.t, r.dir+"/spectre/changes/"+r.change+"/tasks.md", body)
}

// plan is write_tasks_md plus the harness's own "plan" commit.
func (r *tcfRepo) plan(body string) {
	r.writePlan(body)
	r.commitPaths("plan", "spectre/changes/"+r.change+"/tasks.md")
}

func (r *tcfRepo) commitPaths(subject string, paths ...string) string {
	r.git(append([]string{"add", "--"}, paths...)...)
	r.git("commit", "-q", "-m", subject)
	return r.head()
}

// head is `git rev-parse HEAD` read from the repository's files rather than
// a third git spawn per commit: HEAD names a branch whose loose ref the
// commit just wrote.
func (r *tcfRepo) head() string {
	r.t.Helper()
	b, err := os.ReadFile(r.dir + "/.git/HEAD")
	if err == nil {
		b, err = os.ReadFile(r.dir + "/.git/" + strings.TrimPrefix(strings.TrimSpace(string(b)), "ref: "))
	}
	if err != nil {
		r.t.Fatal(err)
	}
	return strings.TrimSpace(string(b))
}

// commit writes each path/content pair, stages exactly those paths, commits
// them under subject and returns the new HEAD.
func (r *tcfRepo) commit(subject string, pairs ...string) string {
	var paths []string
	for i := 0; i < len(pairs); i += 2 {
		writeFile(r.t, r.dir+"/"+pairs[i], pairs[i+1])
		paths = append(paths, pairs[i])
	}
	return r.commitPaths(subject, paths...)
}

// run is the harness's run_guard "$REPO" <args...>.
func (r *tcfRepo) run(args ...string) tcfRes {
	return tcfRunWith(r.getenv(), append([]string{r.dir}, args...)...)
}

// getenv is the fixture's guard environment with TMPDIR at the case's own
// t.TempDir(), so the guard's scratch worktrees stay inside it.
func (r *tcfRepo) getenv() func(string) string {
	getenv, tmp := pathEnv(r.fx.inert), filepath.Dir(r.dir)
	return func(k string) string {
		if k == "TMPDIR" {
			return tmp
		}
		return getenv(k)
	}
}

// ok is one ok: line of the bash harness: a leaf subtest named after it.
func ok(t *testing.T, label string, pass bool, res tcfRes) {
	t.Helper()
	t.Run(label, func(t *testing.T) {
		if !pass {
			t.Errorf("rc=%d out=%s", res.rc, res.out)
		}
	})
}

// has is the harness's `case "$OUT" in *a*b*)`: every part, in order.
func has(s string, parts ...string) bool {
	for _, p := range parts {
		i := strings.Index(s, p)
		if i < 0 {
			return false
		}
		s = s[i+len(p):]
	}
	return true
}

// tcfCounter is cases 112-121's repository: plan key's counter.txt plan,
// then a seed and a bump commit; it returns the bump.
func tcfCounter(t *testing.T, fx *tcfFixture, key string) (*tcfRepo, string) {
	t.Helper()
	r := fx.repo(t, "change-a")
	r.plan(tcfPlans[key])
	r.commit("seed counter", "counter.txt", "5\n")
	return r, r.commit("bump counter", "counter.txt", "6\n")
}

// tcfXRepos is cases 91 and 106's pair: repo B carries the commit and no
// change directory at all; directory A carries the plan only.
func tcfXRepos(t *testing.T) (b, a, sha, parent string) {
	b, a = t.TempDir()+"/xrepo-b", t.TempDir()+"/xcanon-a"
	mkdir(t, b)
	tcfGit(t, b, "init", "-q")
	tcfGit(t, b, "config", "user.email", "test@example.com")
	tcfGit(t, b, "config", "user.name", "Test")
	writeFile(t, b+"/root.txt", "root\n")
	tcfGit(t, b, "add", "root.txt")
	tcfGit(t, b, "commit", "-q", "-m", "root")
	writeFile(t, b+"/alpha.txt", "def test_alpha(): pass\n")
	tcfGit(t, b, "add", "alpha.txt")
	tcfGit(t, b, "commit", "-q", "-m", "add alpha for real")
	writeFile(t, a+"/spectre/changes/x-repo-change/tasks.md", tcfPlans["91"])
	return b, a, tcfGit(t, b, "rev-parse", "HEAD"), tcfGit(t, b, "rev-parse", "HEAD~1")
}

func tcfScriptsDir(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs("../../../scripts")
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestCheckTaskCommitFields(t *testing.T) {
	t.Parallel()
	fx := tcfNewFixture(t)
	cases := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{"case 1", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["1"])
			sha := r.commit("add alpha and beta", "alpha.txt", "a\n", "beta.txt", "# test_alpha covers alpha\n")
			res := r.run("1", sha)
			ok(t, "case 1: files subset of declared passes", res.rc == 0, res)
		}},
		{"case 2", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["2"])
			sha := r.commit("add alpha only", "alpha.txt", "# test_alpha covers alpha\n", "gamma.txt", "gamma\n")
			res := r.run("2", sha)
			ok(t, "case 2: undeclared file fails", res.rc == 1, res)
			ok(t, "case 2: names the undeclared file", has(res.out, "gamma.txt", "not declared"), res)
		}},
		{"case 3", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["3"])
			sha := r.commit("add alpha", "alpha.txt", "def test_alpha(): pass\n")
			res := r.run("3", sha)
			ok(t, "case 3: declared test name found in diff passes", res.rc == 0, res)
		}},
		{"case 4", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["4"])
			sha := r.commit("add alpha", "alpha.txt", "no tests here\n")
			res := r.run("4", sha)
			ok(t, "case 4: missing declared test fails", res.rc == 1, res)
			ok(t, "case 4: names the missing test", has(res.out, "test_alpha", "not found in the diff"), res)
		}},
		{"case 5", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["5"])
			sha := r.commit("add alpha for real", "alpha.txt", "def test_alpha(): pass\n")
			res := r.run("5", sha)
			ok(t, "case 5: commit subject matches declared Commit: passes", res.rc == 0, res)
		}},
		{"case 6", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["6"])
			sha := r.commit("add alpha, not quite right", "alpha.txt", "def test_alpha(): pass\n")
			res := r.run("6", sha)
			ok(t, "case 6: commit subject mismatch fails", res.rc == 1, res)
			ok(t, "case 6: reports the subject mismatch", has(res.out, "subject", "does not match"), res)
		}},
		{"case 7", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["7"])
			sha := r.commit("add alpha and sweep docs", "alpha.txt", "def test_alpha(): pass\n", "docs/notes.md", "swept\n")
			res := r.run("7", sha)
			ok(t, "case 7: extra path covered by Allowed-collateral: glob passes", res.rc == 0, res)
		}},
		{"case 8", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["8"])
			sha := r.commit("add guard test cases", "guard_test.sh", "# Case 1: files subset of declared passes\n# Case 2: undeclared file fails\n")
			res := r.run("8", sha)
			ok(t, "case 8: prose Case N: labels found in diff passes", res.rc == 0, res)
		}},
		{"case 9", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["9"])
			sha := r.commit("add guard test case one only", "guard_test.sh", "# Case 1: files subset of declared passes\n")
			res := r.run("9", sha)
			ok(t, "case 9: missing prose case fails", res.rc == 1, res)
			ok(t, "case 9: names Case 2, not the whole sentence", has(res.out, "Case 2", "not found in the diff"), res)
		}},
		{"case 10", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["10"])
			sha := r.commit("add the wrapper", "guard_test.sh", "wrapper body, no case markers at all\n")
			res := r.run("10", sha)
			ok(t, "case 10: prose with no Case N: or backticks declares nothing, never false-fails", res.rc == 0, res)
		}},
		{"case 15", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["15"])
			sha := r.commit("add alpha", "alpha.txt", "def test_alpha(): pass\n")
			res := r.run("15", sha)
			ok(t, "case 15: field-looking line inside a fenced block is not parsed as real field data", res.rc == 0, res)
		}},
		{"case 17", func(t *testing.T) {
			r := fx.repo(t, "kan-900-some-change")
			r.plan(tcfPlans["17"])
			sha := r.commit("feat(kan-900-some-change): add alpha", "alpha.txt", "a\n")
			res := r.run("17", sha)
			ok(t, "case 17: change-name scope fails", res.rc == 1, res)
			ok(t, "case 17: message names the task and the offending scope", has(res.out, "17", "kan-900-some-change"), res)
		}},
		{"case 18", func(t *testing.T) {
			r := fx.repo(t, "kan-900-some-change")
			r.plan(tcfPlans["18"])
			sha := r.commit("feat(kan-900): add alpha", "alpha.txt", "a\n")
			res := r.run("18", sha)
			ok(t, "case 18: bare-key scope fails", res.rc == 1, res)
			ok(t, "case 18: message names the task and the offending scope", has(res.out, "18", "kan-900"), res)
		}},
		{"case 19", func(t *testing.T) {
			r := fx.repo(t, "kan-900-some-change")
			r.plan(tcfPlans["19"])
			sha := r.commit("feat(3): add alpha", "alpha.txt", "a\n")
			res := r.run("19", sha)
			ok(t, "case 19: numeric task id scope fails", res.rc == 1, res)
			ok(t, "case 19: message names the task and reports the task-id shape", has(res.out, "19", "task id"), res)
		}},
		{"case 20", func(t *testing.T) {
			r := fx.repo(t, "kan-900-some-change")
			r.plan(tcfPlans["20"])
			sha := r.commit("feat(3.2): add alpha", "alpha.txt", "a\n")
			res := r.run("20", sha)
			ok(t, "case 20: dotted task id scope fails", res.rc == 1, res)
			ok(t, "case 20: message names the task and reports the task-id shape", has(res.out, "20", "task id"), res)
		}},
		{"case 21", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["21"])
			sha := r.commit("feat(scripts): add alpha", "alpha.txt", "a\n")
			res := r.run("21", sha)
			ok(t, "case 21: module scope passes", res.rc == 0, res)
			ok(t, "case 21: clean exit, no scope violation printed", res.out == "", res)
		}},
		{"case 22", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["22"])
			sha := r.commit("feat: add alpha", "alpha.txt", "a\n")
			res := r.run("22", sha)
			ok(t, "case 22: absent scope passes", res.rc == 0, res)
			ok(t, "case 22: clean exit, no scope violation printed", res.out == "", res)
		}},
		{"case 23", func(t *testing.T) {
			r := fx.repo(t, "kan-900-some-change")
			r.plan(tcfPlans["23"])
			sha := r.commit("feat(kan-900-helpers): add alpha", "alpha.txt", "a\n")
			res := r.run("23", sha)
			ok(t, "case 23: scope merely containing the key passes (equality, not substring)", res.rc == 0, res)
			ok(t, "case 23: clean exit, no scope violation printed", res.out == "", res)
		}},
		{"case 24", func(t *testing.T) {
			r := fx.repo(t, "kan-900-some-change")
			r.plan(tcfPlans["24"])
			sha := r.commit("feat(kan-900-some-change-helpers): add alpha", "alpha.txt", "a\n")
			res := r.run("24", sha)
			ok(t, "case 24: scope merely containing the change name passes (equality, not substring)", res.rc == 0, res)
			ok(t, "case 24: clean exit, no scope violation printed", res.out == "", res)
		}},
		{"case 25", func(t *testing.T) {
			r := fx.repo(t, "kan-900-some-change")
			r.plan(tcfPlans["25"])
			sha := r.commit("feat(kan-900-some-change)!: add alpha", "alpha.txt", "a\n")
			res := r.run("25", sha)
			ok(t, "case 25: breaking-change '!' form still catches a change-name scope", res.rc == 1, res)
			ok(t, "case 25: message reports the change-name shape", has(res.out, "25", "names the change"), res)
		}},
		{"case 26", func(t *testing.T) {
			r := fx.repo(t, "kan-900-some-change")
			r.plan(tcfPlans["26"])
			sha := r.commit("feat(scripts)!: add alpha", "alpha.txt", "a\n")
			res := r.run("26", sha)
			ok(t, "case 26: breaking-change '!' form with a real module scope passes", res.rc == 0, res)
			ok(t, "case 26: clean exit, no scope violation printed", res.out == "", res)
		}},
		{"case 27", func(t *testing.T) {
			r := fx.repo(t, "kan-900-some-change")
			r.plan(tcfPlans["27"])
			sha := r.commit("feat(KAN-900-SOME-CHANGE): add alpha", "alpha.txt", "a\n")
			res := r.run("27", sha)
			ok(t, "case 27: uppercase change-name scope fails (case-insensitive compare)", res.rc == 1, res)
			ok(t, "case 27: message reports the change-name shape", has(res.out, "27", "names the change"), res)
		}},
		{"case 28", func(t *testing.T) {
			r := fx.repo(t, "kan-900-some-change")
			r.plan(tcfPlans["28"])
			sha := r.commit("feat(KAN-900): add alpha", "alpha.txt", "a\n")
			res := r.run("28", sha)
			ok(t, "case 28: uppercase Jira-key scope fails (case-insensitive compare)", res.rc == 1, res)
			ok(t, "case 28: message reports the Jira-key shape", has(res.out, "28", "Jira key"), res)
		}},
		{"case 29", func(t *testing.T) {
			r := fx.repo(t, "release-2026-kan-450-cleanup")
			r.plan(tcfPlans["29"])
			sha := r.commit("feat(release-2026): add alpha", "alpha.txt", "a\n")
			res := r.run("29", sha)
			ok(t, "case 29: ambiguous key-shaped change name yields no leading key, scope passes", res.rc == 0, res)
			ok(t, "case 29: clean exit, no scope violation printed", res.out == "", res)
		}},
		{"case 30-31", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["30"])
			sha := r.commit("feat: add alpha and beta", "alpha.txt", "def test_alpha(): pass\n", "beta.txt", "b\n")
			res := r.run("1", sha)
			ok(t, "case 30: folded red task passes on its partner's subject", res.rc == 0, res)
			ok(t, "case 30: the red task's own declared subject is not required", !has(res.out, "does not match declared Commit"), res)
			ok(t, "case 30: the file set is the union of both tasks", !has(res.out, "not declared in Files:"), res)
			res = r.run("2", sha)
			ok(t, "case 31: the partner reaches the same verdict against the folded commit", res.rc == 0, res)
			ok(t, "case 31: the union holds when invoked for the partner too", !has(res.out, "not declared in Files:"), res)
		}},
		{"case 32", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["32"])
			sha := r.commit("feat: add alpha and beta", "alpha.txt", "a\n", "beta.txt", "b\n")
			res := r.run("1", sha)
			ok(t, "case 32: a Squash-with naming a missing partner fails", res.rc == 1, res)
			ok(t, "case 32: names the missing partner", has(res.out, "Task 9"), res)
		}},
		{"case 33", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["33"])
			sha := r.commit("feat: add alpha and beta", "alpha.txt", "a\n", "beta.txt", "b\n")
			res := r.run("1", sha)
			ok(t, "case 33: a partner that is itself red fails", res.rc == 1, res)
			ok(t, "case 33: message reports the red partner", has(res.out, "itself red"), res)
		}},
		{"case 34", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["34"])
			sha := r.commit("feat: add alpha", "alpha.txt", "a\n", "beta.txt", "b\n")
			res := r.run("1", sha)
			ok(t, "case 34: without Squash-with, another task's file is still undeclared", res.rc == 1, res)
			ok(t, "case 34: names the other task's file as undeclared", has(res.out, "beta.txt", "not declared"), res)
		}},
		{"case 35", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["34"])
			sha := r.commit("feat: add beta", "alpha.txt", "a\n")
			res := r.run("1", sha)
			ok(t, "case 35: without Squash-with, another task's subject is still a mismatch", res.rc == 1, res)
			ok(t, "case 35: reports the subject mismatch", has(res.out, "does not match declared Commit"), res)
		}},
		{"case 36", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["36"])
			sha := r.commit("feat: add alpha beta and gamma", "alpha.txt", "a\n", "beta.txt", "b\n", "gamma.txt", "g\n")
			res := r.run("1", sha)
			ok(t, "case 36: the red task unions both partners' files", res.rc == 0, res)
			res = r.run("2", sha)
			ok(t, "case 36: a green partner unions the red task's and its sibling's files", res.rc == 0, res)
			ok(t, "case 36: the sibling partner's file is inside the union", !has(res.out, "gamma.txt", "not declared"), res)
			res = r.run("3", sha)
			ok(t, "case 36: the sibling partner reaches the same verdict", res.rc == 0, res)
			ok(t, "case 36: the first partner's file is inside the union too", !has(res.out, "beta.txt", "not declared"), res)
		}},
		{"case 37", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["37"])
			sha := r.commit("feat: add alpha beta and gamma", "alpha.txt", "a\n", "beta.txt", "b\n", "gamma.txt", "g\n")
			res := r.run("1", sha)
			ok(t, "case 37: partners declaring different subjects fail", res.rc == 1, res)
			ok(t, "case 37: reports the disagreement between the partners", has(res.out, "different Commit:"), res)
			ok(t, "case 37: the commit is not blamed for a plan defect", !has(res.out, "does not match declared Commit"), res)
		}},
		{"case 38", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["38"])
			sha := r.commit("feat: add alpha and beta", "alpha.txt", "a\n", "beta.txt", "b\n")
			res := r.run("2", sha)
			ok(t, "case 38: a missing partner fails when checked via a green partner", res.rc == 1, res)
			ok(t, "case 38: names the missing partner from the green side", has(res.out, "Task 9"), res)
		}},
		{"case 39", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["39"])
			sha := r.commit("feat: add alpha and beta", "alpha.txt", "a\n", "beta.txt", "b\n", "gamma.txt", "g\n")
			res := r.run("2", sha)
			ok(t, "case 39: a red partner fails when checked via a green partner", res.rc == 1, res)
			ok(t, "case 39: reports the red partner from the green side", has(res.out, "itself red"), res)
		}},
		{"case 40", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["40"])
			sha := r.commit("feat: add alpha beta and gamma", "alpha.txt", "a\n", "beta.txt", "b\n", "gamma.txt", "g\n")
			res := r.run("1", sha)
			ok(t, "case 40: an undeclared Commit: on one partner is not a disagreement", res.rc == 0, res)
			ok(t, "case 40: no false disagreement is reported", !has(res.out, "different Commit:"), res)
			res = r.run("2", sha)
			ok(t, "case 40: the declaring partner reaches the same verdict", res.rc == 0, res)
			res = r.run("3", sha)
			ok(t, "case 40: the non-declaring partner reaches the same verdict", res.rc == 0, res)
		}},
		{"case 41", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["40"])
			sha := r.commit("chore: something else entirely", "alpha.txt", "a\n", "beta.txt", "b\n", "gamma.txt", "g\n")
			res := r.run("1", sha)
			ok(t, "case 41: the red id fails on the fold's one declared subject", res.rc == 1, res)
			ok(t, "case 41: the red id reports a subject mismatch", has(res.out, "does not match declared Commit"), res)
			res = r.run("3", sha)
			ok(t, "case 41: the non-declaring partner's id fails too", res.rc == 1, res)
			ok(t, "case 41: the non-declaring partner checks against the fold's subject, not nothing", has(res.out, "does not match declared Commit"), res)
		}},
		{"case 42", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["42"])
			sha := r.commit("chore: an entirely unrelated subject", "alpha.txt", "a\n", "beta.txt", "b\n")
			res := r.run("1", sha)
			ok(t, "case 42: a fold declaring no subject at all fails", res.rc == 1, res)
			ok(t, "case 42: reports that nothing declares the folded commit's subject", has(res.out, tcfNoSubject), res)
			ok(t, "case 42: names the red task, where the defective field is", has(res.out, "task 1:"), res)
			res = r.run("2", sha)
			ok(t, "case 42: the green partner's id gives the same verdict", res.rc == 1, res)
			ok(t, "case 42: the green side reports the same defect", has(res.out, tcfNoSubject), res)
		}},
		{"case 43", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["43"])
			sha := r.commit("chore: any subject at all", "alpha.txt", "a\n")
			res := r.run("1", sha)
			ok(t, "case 43: Commit: stays optional for an ordinary task", res.rc == 0, res)
		}},
		{"case 44", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["44"])
			sha := r.commit("feat: subject a", "one.txt", "1\n", "two.txt", "2\n", "shared.txt", "s\n", "four.txt", "4\n", "five.txt", "5\n")
			res := r.run("3", sha)
			ok(t, "case 44: the shared partner no longer passes a commit both red ids reject", res.rc == 1, res)
			ok(t, "case 44: the shared partner reports the folds' disagreement", has(res.out, "different Commit:"), res)
			ok(t, "case 44: the message names the shared task", has(res.out, "Task 3"), res)
			res = r.run("1", sha)
			ok(t, "case 44: the first red id reaches the same verdict", res.rc == 1, res)
			ok(t, "case 44: the first red id reports the same disagreement", has(res.out, "different Commit:"), res)
			res = r.run("2", sha)
			ok(t, "case 44: the second red id reaches the same verdict", res.rc == 1, res)
			ok(t, "case 44: the second red id reports the same disagreement", has(res.out, "different Commit:"), res)
		}},
		{"case 45", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["45"])
			sha := r.commit("feat: the one folded subject", "one.txt", "1\n", "two.txt", "2\n", "shared.txt", "s\n", "four.txt", "4\n", "five.txt", "5\n")
			res := r.run("1", sha)
			ok(t, "case 45: the first red id unions the sibling fold's files", res.rc == 0, res)
			ok(t, "case 45: the sibling fold's file is inside the combined union", !has(res.out, "five.txt", "not declared"), res)
			res = r.run("2", sha)
			ok(t, "case 45: the second red id reaches the same verdict", res.rc == 0, res)
			res = r.run("3", sha)
			ok(t, "case 45: the shared partner reaches the same verdict", res.rc == 0, res)
			res = r.run("4", sha)
			ok(t, "case 45: a partner of one fold sees the other fold's files too", res.rc == 0, res)
			res = r.run("5", sha)
			ok(t, "case 45: and so does a partner of the other", res.rc == 0, res)
		}},
		{"case 46", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["46a"])
			sha := r.commit("chore: an entirely unrelated subject", "alpha.txt", "a\n")
			res := r.run("1", sha)
			ok(t, "case 46: a missing partner still fails", res.rc == 1, res)
			ok(t, "case 46: reports the missing partner", has(res.out, "which does not exist in this plan"), res)
			ok(t, "case 46: the no-subject message does not fire behind a missing partner", !has(res.out, tcfNoSubject), res)

			r = fx.repo(t, "change-a")
			r.plan(tcfPlans["46b"])
			sha = r.commit("chore: an entirely unrelated subject", "alpha.txt", "a\n", "beta.txt", "b\n")
			res = r.run("1", sha)
			ok(t, "case 46: a partner that is itself red still fails", res.rc == 1, res)
			ok(t, "case 46: reports the red partner", has(res.out, "which is itself red"), res)
			ok(t, "case 46: the no-subject message does not fire behind a red partner", !has(res.out, tcfNoSubject), res)
		}},
		{"case 47", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["47"])
			sha := r.commit("chore: an entirely unrelated subject", "alpha.txt", "a\n", "beta.txt", "b\n", "gamma.txt", "g\n")
			res := r.run("1", sha)
			ok(t, "case 47: a missing partner still fails alongside disagreeing ones", res.rc == 1, res)
			ok(t, "case 47: reports the missing partner", has(res.out, "which does not exist in this plan"), res)
			ok(t, "case 47: the disagreement message does not fire behind a missing partner", !has(res.out, "different Commit:"), res)
			ok(t, "case 47: the no-subject message does not fire behind a missing partner either", !has(res.out, tcfNoSubject), res)
		}},
		{"case 48", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["48a"])
			sha := r.commit("feat: the one folded subject", "one.txt", "1\n", "two.txt", "2\n", "shared.txt", "s\n", "four.txt", "4\n", "five.txt", "5\n")
			res := r.run("1", sha)
			ok(t, "case 48: a red whose own partners declare no subject takes the fold's", res.rc == 0, res)
			ok(t, "case 48: the no-subject rule is asked of the combined fold", !has(res.out, tcfNoSubject), res)
			res = r.run("2", sha)
			ok(t, "case 48: the declaring red id reaches the same verdict", res.rc == 0, res)
			res = r.run("3", sha)
			ok(t, "case 48: the shared partner reaches the same verdict", res.rc == 0, res)
			res = r.run("4", sha)
			ok(t, "case 48: the non-declaring partner reaches the same verdict", res.rc == 0, res)
			res = r.run("5", sha)
			ok(t, "case 48: the declaring partner reaches the same verdict", res.rc == 0, res)
		}},
		{"case 48 (continued)", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["48b"])
			sha := r.commit("chore: an entirely unrelated subject", "one.txt", "1\n", "two.txt", "2\n", "shared.txt", "s\n", "four.txt", "4\n", "five.txt", "5\n")
			res := r.run("1", sha)
			ok(t, "case 48: a combined fold declaring no subject anywhere still fails", res.rc == 1, res)
			ok(t, "case 48: reports that nothing in the fold declares the commit's subject", has(res.out, tcfNoSubject), res)
			ok(t, "case 48: the message is anchored on the shared task", has(res.out, "task 3:"), res)
			res = r.run("2", sha)
			ok(t, "case 48: the second red id reaches the same verdict", res.rc == 1, res)
			res = r.run("3", sha)
			ok(t, "case 48: the shared partner reaches the same verdict", res.rc == 1, res)
			res = r.run("5", sha)
			ok(t, "case 48: a non-declaring partner reaches the same verdict", res.rc == 1, res)
		}},
		{"case 49", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["49"])
			sha := r.commit("feat: add a and b", "a.txt", "a\n", "unrelated.txt", "u\n", "b.txt", "b\n")
			res := r.run("1", sha)
			ok(t, "case 49: a malformed Squash-with field fails instead of passing an undeclared file", res.rc == 1, res)
			ok(t, "case 49: the message names the malformed field", has(res.out, "Squash-with:", "is not `Task"), res)
			res = r.run("2", sha)
			ok(t, "case 49: the unrelated task is not dragged into the fold by its own id", res.rc == 1, res)
		}},
		{"case 50", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["50"])
			sha := r.commit("feat: add alpha and beta", "alpha.txt", "a\n", "beta.txt", "b\n")
			res := r.run("1", sha)
			ok(t, "case 50: free text in Squash-with is a malformed field, not a partner list", res.rc == 1, res)
			ok(t, "case 50: no partner id is invented from the field's free text", !has(res.out, "which does not exist in this plan"), res)
			ok(t, "case 50: the message names the malformed field instead", has(res.out, "Squash-with:", "is not `Task"), res)
		}},
		{"case 51", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["51"])
			sha := r.commit("feat: add alpha beta and gamma", "alpha.txt", "a\n", "beta.txt", "b\n", "gamma.txt", "g\n")
			res := r.run("1", sha)
			ok(t, "case 51: whitespace-separated partners still resolve", res.rc == 0, res)
			res = r.run("2", sha)
			ok(t, "case 51: the first partner reaches the same verdict", res.rc == 0, res)
			res = r.run("3", sha)
			ok(t, "case 51: the sibling partner reaches the same verdict", res.rc == 0, res)
		}},
		{"case 52", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["52"])
			sha := r.commit("feat: add alpha and beta", "alpha.txt", "a\n", "beta.txt", "b\n")
			res := r.run("1", sha)
			ok(t, "case 52: an unblanked prose line is not slurped into the Squash-with value", res.rc == 0, res)
			res = r.run("2", sha)
			ok(t, "case 52: the partner reaches the same verdict", res.rc == 0, res)
		}},
		{"case 53", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["53"])
			sha := r.commit("feat: add alpha beta and gamma", "alpha.txt", "a\n", "beta.txt", "b\n", "gamma.txt", "g\n")
			res := r.run("1", sha)
			ok(t, "case 53: a wrapped partner id is not read as part of the field", res.rc == 1, res)
			ok(t, "case 53: the fold's file union is not widened to the wrapped id's task", has(res.out, "gamma.txt is not declared"), res)
		}},
		{"case 54", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["54"])
			sha := r.commit("feat: add beta and gamma", "beta.txt", "b\n", "gamma.txt", "g\n")
			res := r.run("10", sha)
			ok(t, "case 54: an unrelated valid fold is not contaminated by another red's invalid edge", res.rc == 0, res)
			res = r.run("9", sha)
			ok(t, "case 54: the red half of that valid fold reaches the same verdict", res.rc == 0, res)
			res = r.run("1", sha)
			ok(t, "case 54: the invalid edge is still a violation from the task carrying it", res.rc == 1, res)
			ok(t, "case 54: the violation names the red partner it could not resolve", has(res.out, "Task 9, which is itself red"), res)
		}},
		{"case 55", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["55"])
			sha := r.commit("feat: add alpha and beta", "alpha.txt", "a\n", "beta.txt", "b\n")
			res := r.run("2", sha)
			ok(t, "case 55: a malformed Squash-with elsewhere in the plan still fails the partner it names", res.rc == 1, res)
			ok(t, "case 55: the malformed field is discoverable from the partner id", has(res.out, "task 1: Squash-with:", "is not `Task"), res)
			ok(t, "case 55: the misleading undeclared-file message is not reported instead", !has(res.out, "alpha.txt is not declared"), res)
		}},
		// Case 56 pinned the bash wrapper's own missing-sibling exit. The Go
		// port compiles its grammar in, so the one missing piece a shipped
		// guard can still meet is the toolchain the shim builds flow-guard
		// with (Decision: guard-binary-built-from-checkout): exit 2 under the
		// header's COULD NOT JUDGE opening, run through real bash, with an
		// empty build cache so nothing is already built.
		{"case 56", func(t *testing.T) {
			cmd := exec.Command("/bin/bash", tcfScriptsDir(t)+"/check-task-commit-fields.sh", "wt", "1", "sha")
			cmd.Env = []string{"PATH=/usr/bin:/bin", "FLOW_GUARD_CACHE_DIR=" + t.TempDir()}
			out, err := cmd.CombinedOutput()
			if cmd.ProcessState == nil {
				t.Fatal(err)
			}
			res := tcfRes{cmd.ProcessState.ExitCode(), string(out)}
			ok(t, "case 56: a shim that cannot build flow-guard exits 2", res.rc == 2, res)
			ok(t, "case 56: the refusal carries the COULD NOT JUDGE opening and names the missing go",
				has(res.out, tcfNotAVerdict+" cannot build flow-guard from", "no go on PATH"), res)
		}},
		{"case 57", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["57"])
			sha := r.commit("feat: add alpha and beta", "alpha.txt", "a\n", "beta.txt", "b\n")
			res := r.run("1", sha)
			ok(t, "case 57: the gating line is the field, so the red half of the fold passes", res.rc == 0, res)
			res = r.run("2", sha)
			ok(t, "case 57: the green half reaches the same verdict", res.rc == 0, res)
			ok(t, "case 57: the non-gating line is not reported as the field's value", !has(res.out, "is not `Task"), res)
		}},
		{"case 58", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["58"])
			sha := r.commit("test: add alpha", "alpha.txt", "a\n")
			res := r.run("1", sha)
			ok(t, "case 58: the first Build line is the tag, so the task is red", res.rc == 1, res)
			ok(t, "case 58: the red task's missing partner is reported", has(res.out, "task 1: Squash-with: names Task 9, which does not exist in this plan"), res)
		}},
		{"case 59", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["59"])
			sha := r.commit("test: add alpha", "alpha.txt", "a\n")
			res := r.run("1", sha)
			ok(t, "case 59: a prose line under the tag does not unset it", res.rc == 1, res)
			ok(t, "case 59: the red task's missing partner is reported", has(res.out, "task 1: Squash-with: names Task 9, which does not exist in this plan"), res)
		}},
		{"case 60", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["60"])
			sha := r.commit("test: add alpha", "alpha.txt", "a\n")
			res := r.run("1", sha)
			ok(t, "case 60: a fenced example task line opens no task", res.rc == 0, res)
			ok(t, "case 60: the fenced example's Squash-with is not reported", !has(res.out, "is not `Task"), res)
		}},
		{"case 61", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["61"])
			sha := r.commit("test: add alpha", "alpha.txt", "a\n", "gamma.txt", "g\n")
			res := r.run("1", sha)
			ok(t, "case 61: a duplicated id resolves to its first task line", res.rc == 0, res)
			ok(t, "case 61: the later task line's fields are not what the commit is checked against", !has(res.out, "zulu.txt"), res)
		}},
		{"case 62", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["62"])
			sha := r.commit("test: add real", "real.txt", "r\n")
			res := r.run("1", sha)
			ok(t, "case 62: the real Files: declaration survives a tilde-fenced example", res.rc == 0, res)

			r = fx.repo(t, "change-a")
			r.plan(tcfPlans["62"])
			sha = r.commit("test: add real", "example-only.txt", "e\n")
			res = r.run("1", sha)
			ok(t, "case 62: a tilde-fenced example file is not declared", res.rc == 1, res)
			ok(t, "case 62: the violation names the fenced example's file", has(res.out, "example-only.txt"), res)
		}},
		{"case 63", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["63"])
			sha := r.commit("test: add real", "real.txt", "r\n")
			res := r.run("1", sha)
			ok(t, "case 63: the real Files: declaration survives an indented fenced example", res.rc == 0, res)
		}},
		{"case 64", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["64a"])
			sha := r.commit("test: add alpha", "alpha.txt", "a\n")
			res := r.run("1", sha)
			ok(t, "case 64: an unclosed fence in a task body fails", res.rc == 1, res)
			ok(t, "case 64: names the task and the line the fence opened on", has(res.out, "code fence opened at tasks.md line 3 is never closed"), res)
			ok(t, "case 64: the undeclared-file noise is replaced, not accompanied", !has(res.out, "is not declared in Files:"), res)

			r = fx.repo(t, "change-a")
			r.plan(tcfPlans["64b"])
			sha = r.commit("test: add alpha", "alpha.txt", "a\n")
			res = r.run("1", sha)
			ok(t, "case 64: a closed fence is not reported as unclosed", res.rc == 0, res)
		}},
		{"case 65", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.writePlan(tcfPlans["65"])
			writeFile(t, r.dir+"/spectre/changes/change-a-fix-1/tasks.md", tcfPlans["65-fix-1"])
			r.commitPaths("plan", "spectre/changes")
			sha := r.commit("fix: add alpha", "alpha.txt", "a\n", "beta.txt", "# test_alpha covers alpha\n")
			res := r.run("1", sha)
			ok(t, "case 65: a fix sub-change beside its parent is not an ambiguity refusal", res.rc != 2, res)
			ok(t, "case 65: no ambiguity message", !has(res.out, "more than one tasks.md"), res)
			ok(t, "case 65: the sub-change's own plan was the one read", res.rc == 0, res)
		}},
		{"case 66", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.writePlan(tcfPlans["66"])
			writeFile(t, r.dir+"/spectre/changes/change-a-fix-1/tasks.md", tcfPlans["66-fix-1"])
			writeFile(t, r.dir+"/spectre/changes/change-a-fix-2/tasks.md", tcfPlans["66-fix-2"])
			r.commitPaths("plan", "spectre/changes")
			sha := r.commit("fix: add alpha", "alpha.txt", "a\n", "beta.txt", "# test_alpha covers alpha\n")
			res := r.run("1", sha)
			ok(t, "case 66: the highest-numbered fix sibling is the plan read", res.rc == 0, res)
		}},
		{"case 67", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.writePlan(tcfPlans["67"])
			writeFile(t, r.dir+"/spectre/changes/change-a-fix-the-parser/tasks.md", tcfPlans["67-other"])
			r.commitPaths("plan", "spectre/changes")
			sha := r.commit("test: alpha", "alpha.txt", "a\n")
			res := r.run("1", sha)
			ok(t, "case 67: two unrelated changes still refuse with exit 2", res.rc == 2, res)
			ok(t, "case 67: the ambiguity message is still the one printed", has(res.out, "more than one tasks.md"), res)
			// Not a harness label: the bash glob sorts whole paths, so the
			// fix-the-parser plan is listed first ('-' sorts before '/').
			if !has(res.out, "change-a-fix-the-parser/tasks.md "+r.dir+"/spectre/changes/change-a/tasks.md\n") {
				t.Errorf("ambiguity list not in bash glob order: %s", res.out)
			}
		}},
		{"case 68", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["68"])
			sha := r.commit("test: alpha", "alpha.txt", "a\n")
			res := r.run("68", sha)
			ok(t, "case 68: none-opening field with backticks declares nothing", res.rc == 0, res)
		}},
		{"case 69", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["69"])
			sha := r.commit("test: alpha", "alpha.txt", "a\n")
			res := r.run("69", sha)
			ok(t, "case 69: none-opening field with a Case N label declares nothing", res.rc == 0, res)
		}},
		{"case 70", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["70"])
			sha := r.commit("test: alpha", "alpha.txt", "a\n")
			res := r.run("70", sha)
			ok(t, "case 70: mid-sentence none does not suppress the backtick test", res.rc == 1, res)
			ok(t, "case 70: names test_alpha as missing", has(res.out, "test_alpha", "not found in the diff"), res)
		}},
		{"case 71", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["71"])
			sha := r.commit("test: alpha", "alpha.txt", "a\n")
			res := r.run("71", sha)
			ok(t, "case 71: a none-prefixed word does not open a none field", res.rc == 1, res)
			ok(t, "case 71: names test_alpha as missing", has(res.out, "test_alpha", "not found in the diff"), res)
		}},
		{"case 72", func(t *testing.T) {
			r := fx.repo(t, "sat-demo")
			writeFile(t, r.dir+"/spectre/changes/sat-demo/link.md", "## Part of\n\n`peerx:canon-demo`\n")
			r.commitPaths("link", "spectre/changes/sat-demo/link.md")
			sha := r.commit("add alpha for real", "alpha.txt", "def test_alpha(): pass\n")
			canon := t.TempDir()
			writeFile(t, canon+"/spectre/changes/canon-demo/tasks.md", tcfPlans["72"])
			res := r.run("72", sha, "", canon)
			ok(t, "case 72: a satellite worktree resolves its task against the canonical worktree's plan", res.rc == 0, res)
		}},
		{"case 73", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.writePlan(tcfPlans["73"])
			writeFile(t, r.dir+"/spectre/changes/other-sat/link.md", "## Part of\n\n`peery:some-other-change`\n")
			r.commitPaths("plan", "spectre/changes")
			sha := r.commit("add alpha for real", "alpha.txt", "def test_alpha(): pass\n")
			res := r.run("73", sha)
			ok(t, "case 73: a satellite directory beside a root change is not ambiguity", res.rc == 0, res)
			ok(t, "case 73: no ambiguity message", !has(res.out, "more than one tasks.md"), res)
		}},
		{"case 74", func(t *testing.T) {
			r := fx.repo(t, "sat-demo-unresolvable")
			writeFile(t, r.dir+"/spectre/changes/sat-demo-unresolvable/link.md", "## Part of\n\n`peerz:canon-demo`\n")
			r.commitPaths("link", "spectre/changes/sat-demo-unresolvable/link.md")
			sha := r.commit("add alpha", "alpha.txt", "a\n")
			res := r.run("74", sha)
			ok(t, "case 74: an unresolvable satellite exits 2", res.rc == 2, res)
			ok(t, "case 74: reports no tasks.md could be found or resolved", has(res.out, "no tasks.md found"), res)
		}},
		{"case 75", func(t *testing.T) {
			r := fx.repo(t, "sat-demo-ambiguous")
			writeFile(t, r.dir+"/spectre/changes/sat-demo-ambiguous/link.md", "## Part of\n\n`peerx:canon-demo`\n")
			writeFile(t, r.dir+"/spectre/changes/sat-demo-ambiguous-second/link.md", "## Part of\n\n`peery:canon-demo`\n")
			r.commitPaths("two links", "spectre/changes")
			sha := r.commit("add alpha", "alpha.txt", "a\n")
			canon := t.TempDir()
			writeFile(t, canon+"/spectre/changes/canon-demo/tasks.md", tcfPlans["75"])
			res := r.run("75", sha, "", canon)
			ok(t, "case 75: two link-only satellite directories refuse rather than guessing", res.rc == 2, res)
			ok(t, "case 75: reports no tasks.md could be found or resolved", has(res.out, "no tasks.md found"), res)
		}},
		{"case 76", func(t *testing.T) {
			r := fx.repo(t, "sat-demo-76")
			mkdir(t, r.dir+"/spectre/changes/pre-plan-change")
			writeFile(t, r.dir+"/spectre/changes/sat-demo-76/link.md", "## Part of\n\n`peerx:canon-demo-76`\n")
			r.commitPaths("link", "spectre/changes/sat-demo-76/link.md")
			sha := r.commit("add alpha for real", "alpha.txt", "def test_alpha(): pass\n")
			canon := t.TempDir()
			writeFile(t, canon+"/spectre/changes/canon-demo-76/tasks.md", tcfPlans["76"])
			res := r.run("76", sha, "", canon)
			ok(t, "case 76: a satellite beside an ordinary pre-plan change directory still resolves", res.rc == 0, res)
			ok(t, "case 76: no false-ambiguity refusal", !has(res.out, "no tasks.md found"), res)
		}},
		{"case 77", func(t *testing.T) {
			r := fx.repo(t, "kan-367-demo")
			r.writePlan(tcfPlans["77"])
			writeFile(t, r.dir+"/spectre/changes/some-other-live-change/tasks.md", tcfPlans["77-other"])
			r.commitPaths("plan", "spectre/changes")
			sha := r.commit("test: alpha", "alpha.txt", "def test_alpha(): pass\n")
			res := r.run("1", sha, "", "", "kan-367-demo")
			ok(t, "case 77: a named change resolves directly despite an unrelated live root", res.rc == 0, res)
			ok(t, "case 77: no ambiguity message", !has(res.out, "more than one tasks.md"), res)
		}},
		{"case 78", func(t *testing.T) {
			r := fx.repo(t, "kan-367-fix-demo")
			r.writePlan(tcfPlans["78"])
			writeFile(t, r.dir+"/spectre/changes/kan-367-fix-demo-fix-1/tasks.md", tcfPlans["78-fix-1"])
			writeFile(t, r.dir+"/spectre/changes/unrelated-third-change/tasks.md", tcfPlans["78-third"])
			r.commitPaths("plan", "spectre/changes")
			sha := r.commit("fix: add alpha", "alpha.txt", "def test_alpha(): pass\n")
			res := r.run("1", sha, "", "", "kan-367-fix-demo")
			ok(t, "case 78: the named change's own fix sibling wins, scoped to its family", res.rc == 0, res)
		}},
		{"case 79", func(t *testing.T) {
			r := fx.repo(t, "kan-367-demo-79")
			sha := r.commit("test: alpha", "alpha.txt", "a\n")
			res := r.run("1", sha, "", "", "no-such-change")
			ok(t, "case 79: an unknown change name exits 2", res.rc == 2, res)
			ok(t, "case 79: names the unresolved change", has(res.out, "no tasks.md found for change 'no-such-change'"), res)
		}},
		{"case 80", func(t *testing.T) {
			r := fx.repo(t, "kan-367-sat")
			writeFile(t, r.dir+"/spectre/changes/kan-367-sat/link.md", "## Part of\n\n`peerx:canon-demo-367`\n")
			r.commitPaths("link", "spectre/changes/kan-367-sat/link.md")
			sha := r.commit("add alpha for real", "alpha.txt", "def test_alpha(): pass\n")
			canon := t.TempDir()
			writeFile(t, canon+"/spectre/changes/canon-demo-367/tasks.md", tcfPlans["80"])
			res := r.run("1", sha, "", canon, "kan-367-sat")
			ok(t, "case 80: a satellite named directly resolves through its link", res.rc == 0, res)
		}},
		{"case 81", func(t *testing.T) {
			r := fx.repo(t, "kan-367-fix-unplanned")
			r.writePlan(tcfPlans["78"])
			writeFile(t, r.dir+"/spectre/changes/kan-367-fix-unplanned-fix-1/tasks.md", tcfPlans["78-fix-1"])
			mkdir(t, r.dir+"/spectre/changes/kan-367-fix-unplanned-fix-2")
			r.commitPaths("plan", "spectre/changes")
			sha := r.commit("fix: add alpha", "alpha.txt", "def test_alpha(): pass\n")
			res := r.run("1", sha, "", "", "kan-367-fix-unplanned")
			ok(t, "case 81: an unplanned higher-numbered fix sibling is skipped for the planned one", res.rc == 0, res)
		}},
		{"case 82", func(t *testing.T) {
			r := fx.repo(t, "kan-367-invalid-name")
			r.writePlan(tcfPlans["82"])
			sha := r.commit("test: alpha", "alpha.txt", "def test_alpha(): pass\n")
			res := r.run("1", sha, "", "", "../x")
			ok(t, "case 82: a path-traversal-shaped change name exits 2", res.rc == 2, res)
			ok(t, "case 82: names it an invalid change name", has(res.out, "invalid change name"), res)
			res = r.run("1", sha, "", "", "a/b")
			ok(t, "case 82: a slash-containing change name exits 2", res.rc == 2, res)
			ok(t, "case 82: slash-containing name is named invalid too", has(res.out, "invalid change name"), res)
		}},
		{"case 87", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["87"])
			sha := r.commit("test: add alpha", "alpha.txt", "a\n")
			res := r.run("1", sha)
			ok(t, "case 87: red followed by prose is still red", res.rc == 1, res)
			ok(t, "case 87: the red task's missing partner is reported", has(res.out, "task 1: Squash-with: names Task 9, which does not exist in this plan"), res)
		}},
		{"case 88", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["88"])
			sha := r.commit("feat: add alpha and beta", "alpha.txt", "def test_alpha(): pass\n", "beta.txt", "b\n")
			res := r.run("1", sha)
			ok(t, "case 88: a malformed-tag partner is treated as not-red", res.rc == 0, res)
			ok(t, "case 88: no itself-red violation for the malformed-tag partner", !has(res.out, "itself red"), res)
			ok(t, "case 88: the file set is the union of both tasks", !has(res.out, "not declared in Files:"), res)
		}},
		{"case 89", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["89"])
			sha := r.commit("feat: add alpha", "alpha.txt", "def test_alpha(): pass\n")
			writeFile(t, r.dir+"/root.txt", "root\nstaged\n")
			r.git("add", "root.txt")
			writeFile(t, r.dir+"/alpha.txt", "def test_alpha(): pass\nunstaged\n")
			writeFile(t, r.dir+"/scratch.txt", "untracked\n")
			before := r.git("status", "--porcelain")
			res := r.run("1", sha)
			ok(t, "case 89: guard exits 0 on a dirty worktree with no ## test command", res.rc == 0, res)
			ok(t, "case 89: HEAD unchanged", r.git("rev-parse", "HEAD") == sha, res)
			ok(t, "case 89: index and working tree untouched", r.git("status", "--porcelain") == before, res)
			ok(t, "case 89: no stash entry and no skipped-not-verified notice",
				r.git("stash", "list") == "" && !has(res.out, "skipped, not verified"), res)
		}},
		{"case 90", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["90a"])
			sha := r.commit("add alpha", "alpha.txt", "def test_alpha(): pass\n")
			res := r.run("1", sha)
			ok(t, "case 90: the After line after a bullet run leaves the Files set intact", res.rc == 0, res)

			r = fx.repo(t, "change-a")
			r.plan(tcfPlans["90b"])
			sha = r.commit("add alpha", "alpha.txt", "def test_alpha(): pass\n")
			res = r.run("1", sha)
			ok(t, "case 90: the After line after a prose Files field leaves the declared path intact", res.rc == 0, res)
		}},
		{"case 91", func(t *testing.T) {
			b, a, sha, parent := tcfXRepos(t)
			res := fx.run(b, "1", sha, parent, a, "x-repo-change")
			ok(t, "case 91: named change with no local dir resolves through the canonical worktree", res.rc == 0, res)
			res = fx.run(b, "1", sha, parent, "", "x-repo-change")
			ok(t, "case 91b: named change with no local dir and no canonical worktree still refuses", res.rc == 2, res)
			ok(t, "case 91b: it names the missing plan", has(res.out, "no tasks.md found for change 'x-repo-change'"), res)
			res = fx.run(b, "1", sha, parent)
			ok(t, "case 91c: no change-name argument and no local dir still refuses through the glob path", res.rc == 2, res)
			ok(t, "case 91c: it names the glob-path refusal", has(res.out, "no tasks.md found under"), res)
		}},
		{"case 92-93", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["92"])
			r.commit("add alpha", "alpha.txt", "def test_alpha(): pass\n")
			res := r.run("1", "1e803d5564273")
			ok(t, "case 92: could not judge — mistyped commit sha exits 2 under the standard line", res.rc == 2, res)
			ok(t, "case 92: the standard opening carries git's own detail",
				has(res.out, "check-task-commit-fields: COULD NOT JUDGE — not a commit verdict", "ambiguous argument"), res)
			res = r.run()
			ok(t, "case 93: could not judge — one-argument call exits 2 under the standard line", res.rc == 2, res)
			ok(t, "case 93: the standard opening carries the usage detail",
				has(res.out, "check-task-commit-fields: COULD NOT JUDGE — not a commit verdict", "usage:"), res)
		}},
		{"case 94", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["94"])
			sha := r.commit("add alpha only", "alpha.txt", "# test_alpha covers alpha\n", "gamma.txt", "gamma\n")
			res := r.run("1", sha)
			ok(t, "case 94: real violations exit 1 with no could-not-judge line", res.rc == 1, res)
			ok(t, "case 94: the verdict carries no could-not-judge opening", !has(res.out, "COULD NOT JUDGE"), res)
		}},
		{"case 95", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["95"])
			sha := r.commit("add alpha", "alpha.txt", "def test_alpha(): pass\n")
			res := r.run("1", sha, "", r.dir, "change-a")
			ok(t, "case 95: the documented empty-placeholder shape resolves and derives the parent", res.rc == 0, res)
			ok(t, "case 95: clean exit, no refusal printed", res.out == "", res)
		}},
		{"case 96", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["96"])
			r.commit("seed alpha", "alpha.txt", "@Test fun a() {}\n")
			sha := r.commit("grow alpha tests", "alpha.txt", "@Test fun a() {}\n@Test fun b() {}\n# test_alpha\n")
			res := r.run("1", sha)
			ok(t, "case 96: baseline delta matching the measured @Test counts passes", res.rc == 0, res)
		}},
		{"case 97", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["97"])
			r.commit("seed alpha", "alpha.txt", "@Test fun a() {}\n")
			sha := r.commit("grow alpha tests", "alpha.txt", "@Test fun a() {}\n@Test fun b() {}\n# test_alpha\n")
			res := r.run("1", sha)
			ok(t, "case 97: stale baseline delta fails", res.rc == 1, res)
			ok(t, "case 97: names declared and measured counts", has(res.out, "before=1 after=3", "before=1 after=2"), res)
		}},
		{"case 98", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["98"])
			r.commit("seed alpha", "alpha.txt", "harness\n# test_alpha\n")
			sha := r.commit("bump harness count", "alpha.txt", "harness\ntwo\n# test_alpha\n")
			res := r.run("1", sha)
			ok(t, "case 98: baseline check skips when no @Test at either revision", res.rc == 0, res)
		}},
		{"case 99", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["99"])
			r.commit("seed alpha", "alpha.txt", "def test_gone(): pass\ndef test_kept(): pass\n")
			sha := r.commit("remove gone test", "alpha.txt", "def test_kept(): pass\n")
			res := r.run("1", sha)
			ok(t, "case 99: a test the commit removed fails the tree check", res.rc == 1, res)
			ok(t, "case 99: names the removed test and the tree", has(res.out, "test_gone", "not found in the tree"), res)
		}},
		{"case 100", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["100"])
			sha := r.commit("add alpha", "alpha.txt", "plain content\n")
			res := r.run("1", sha)
			ok(t, "case 100: stale bare camelCase name fails the tree check", res.rc == 1, res)
			ok(t, "case 100: names the stale camelCase name", has(res.out, "calculateDailyCaloriesReturnsZero", "not found in the tree"), res)
		}},
		{"case 101", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["101"])
			sha := r.commit("add alpha", "alpha.txt", "def test_alpha(): pass\n# test_beta\n")
			res := r.run("1", sha)
			ok(t, "case 101: names present in tree content pass", res.rc == 0, res)
		}},
		{"case 102", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["102"])
			sha := r.commit("add alpha", "alpha.txt", "plain content\n")
			res := r.run("1", sha)
			ok(t, "case 102: none-opening Tests stays vacuous for the tree check", res.rc == 0, res)
		}},
		{"case 103", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["103"])
			r.commit("seed helper", "helper.md", "original helper body\n")
			sha := r.commit("touch helper", "helper.md", "revised helper body\n")
			res := r.run("1", sha)
			ok(t, "case 103: a Tests token existing as a committed path passes", res.rc == 0, res)
		}},
		{"case 104", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.writePlan(tcfPlans["104"])
			writeFile(t, r.dir+"/spectre/changes/archive/old-change/tasks.md", tcfPlans["104-archive"])
			r.commitPaths("plan", "spectre/changes")
			r.commit("seed alpha", "alpha.txt", "x\n# test_stale\n")
			sha := r.commit("remove stale test", "alpha.txt", "x\n")
			res := r.run("1", sha)
			ok(t, "case 104: an archived plan cannot vouch for a stale name", res.rc == 1, res)
			ok(t, "case 104: names the stale name despite the archive copy", has(res.out, "test_stale", "not found in the tree"), res)
		}},
		{"case 105", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["105"])
			r.commit("seed alpha", "alpha.txt", "@TestFactory\n@Test fun a() {}\n")
			sha := r.commit("grow alpha tests", "alpha.txt", "@TestFactory\n@TestInstance(LIFECYCLE.PER_METHOD)\n@Test fun a() {}\n@Test fun b() {}\n# test_alpha\n")
			res := r.run("1", sha)
			ok(t, "case 105: lifecycle annotations do not count as @Test", res.rc == 0, res)
		}},
		{"case 106-107", func(t *testing.T) {
			b, a, sha, parent := tcfXRepos(t)
			stub := t.TempDir()
			calls, fixture := stub+"/calls.log", stub+"/fixture.json"
			writeExec(t, stub+"/bin/flow", "#!/usr/bin/env bash\nprintf '%s\\n' \"$*\" >> \""+calls+"\"\n"+
				"if [ \"$1 $2\" = \"state find\" ] && [ -f \""+fixture+"\" ]; then\n  cat \""+fixture+"\"\n  exit 0\nfi\nexit 1\n")
			env := pathEnv(stub + "/bin")

			writeFile(t, fixture, `{"source":"store","complete":true,"records":[{"projectKey":"proj-a","name":"x-repo-change","state":"IN_PROGRESS","worktrees":{"`+a+`":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},"updatedAt":"2026-09-10T10:00:00Z","updatedBy":"/flow-fast"}]}`+"\n")
			res := tcfRunWith(env, b, "1", sha, parent, "", "x-repo-change")
			ok(t, "case 106: store-resolves-canonical-plan-no-arg resolves the record's plan", res.rc == 0, res)
			logged, _ := os.ReadFile(calls)
			ok(t, "case 106: the store was asked for the change's own name", strings.Contains(string(logged), "state find x-repo-change"), tcfRes{res.rc, string(logged)})

			treeA, treeB := t.TempDir(), t.TempDir()
			writeFile(t, treeA+"/spectre/changes/x-repo-change/tasks.md", tcfPlans["91"])
			writeFile(t, treeB+"/spectre/changes/x-repo-change/tasks.md", tcfPlans["107-b"])
			writeFile(t, fixture, `{"source":"store","complete":true,"records":[`+"\n"+
				`{"projectKey":"proj-a","name":"x-repo-change","worktrees":{"`+treeA+`":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}},`+"\n"+
				`{"projectKey":"proj-b","name":"x-repo-change","worktrees":{"`+treeB+`":"cccccccccccccccccccccccccccccccccccccccc"}}]}`+"\n")
			res = tcfRunWith(env, b, "1", sha, parent, "", "x-repo-change")
			ok(t, "case 107: store-ambiguity-refuses exits 2 outright", res.rc == 2, res)
			ok(t, "case 107: the refusal names every project and path", has(res.out, "ambiguous state-record resolution", "proj-a", "proj-b"), res)
		}},
		{"case 108", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["108"])
			sha := r.commit("add alpha", "alpha.txt", "no tests here\n")
			res := r.run("1", sha)
			ok(t, "case 108: prose field naming a script fails", res.rc == 1, res)
			ok(t, "case 108: reports the parse rule with the missing name", has(res.out, "check-references.sh", "not found in the diff", "parsed, not read"), res)
		}},
		{"case 109", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["109"])
			sha := r.commit("add alpha", "alpha.txt", "no tests here\n")
			res := r.run("1", sha)
			ok(t, "case 109: genuine missing test fails", res.rc == 1, res)
			ok(t, "case 109: names the test first, parse rule second", has(res.out, "test_alpha", "not found in the diff", "parsed, not read"), res)
		}},
		{"case 110", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["110"])
			sha := r.commit("add alpha", "alpha.txt", "no tests here\n")
			res := r.run("1", sha)
			ok(t, "case 110: labelled field missing its label fails", res.rc == 1, res)
			ok(t, "case 110: states the label-alone rule, not the backtick claim", has(res.out, "Case 99", "not found in the diff", "checked by that label alone"), res)
		}},
		{"case 111", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["111"])
			sha := r.commit("add alpha", "alpha.txt", "no tests here\n")
			res := r.run("1", sha)
			ok(t, "case 111: bare camelCase tree miss fails", res.rc == 1, res)
			ok(t, "case 111: tree message carries the hint", has(res.out, "helperAlpha", "not found in the tree", "parsed, not read"), res)
			ok(t, "case 111: tree message states the camelCase rule", has(res.out, "helperAlpha", "not found in the tree", "camelCase"), res)
		}},
		// Cases 112-121 are one entry each, so their ten measured runs
		// proceed in parallel rather than one after another.
		{"case 112", func(t *testing.T) {
			r, sha := tcfCounter(t, fx, "112")
			res := r.run("1", sha)
			ok(t, "case 112: recorded command matching the declaration passes", res.rc == 0, res)
		}},
		{"case 113", func(t *testing.T) {
			r, sha := tcfCounter(t, fx, "113")
			res := r.run("1", sha)
			ok(t, "case 113: stale after-count fails", res.rc == 1, res)
			ok(t, "case 113: names declared and measured counts", has(res.out, "declares before=5 after=9", "measures before=5 after=6"), res)
		}},
		{"case 114", func(t *testing.T) {
			r, sha := tcfCounter(t, fx, "114")
			res := r.run("1", sha)
			ok(t, "case 114: stale before-count fails", res.rc == 1, res)
		}},
		{"case 115", func(t *testing.T) {
			r, sha := tcfCounter(t, fx, "115")
			res := r.run("1", sha)
			ok(t, "case 115: no recorded command skips the dynamic check", res.rc == 0, res)
		}},
		{"case 116", func(t *testing.T) {
			r, sha := tcfCounter(t, fx, "116")
			res := r.run("1", sha)
			ok(t, "case 116: a failing recorded command skips", res.rc == 0, res)
		}},
		{"case 117", func(t *testing.T) {
			r, sha := tcfCounter(t, fx, "117")
			res := r.run("1", sha)
			ok(t, "case 117: a non-integer measurement skips", res.rc == 0, res)
		}},
		{"case 118", func(t *testing.T) {
			r, sha := tcfCounter(t, fx, "118")
			res := r.run("1", sha)
			ok(t, "case 118: two distinct recorded commands skip", res.rc == 0, res)
		}},
		{"case 119", func(t *testing.T) {
			r, sha := tcfCounter(t, fx, "119")
			res := r.run("1", sha)
			ok(t, "case 119: one command across two refs verifies", res.rc == 0, res)
		}},
		{"case 120", func(t *testing.T) {
			r, sha := tcfCounter(t, fx, "120")
			trees := r.git("worktree", "list", "--porcelain")
			res := r.run("1", sha)
			ok(t, "case 120: measured run passes", res.rc == 0, res)
			ok(t, "case 120: the checked worktree's HEAD never moved", r.git("rev-parse", "HEAD") == sha, res)
			ok(t, "case 120: no throwaway worktree left registered", r.git("worktree", "list", "--porcelain") == trees, res)
		}},
		{"case 121", func(t *testing.T) {
			r, sha := tcfCounter(t, fx, "121")
			res := r.run("1", sha)
			ok(t, "case 121: a pipeline measurement verifies", res.rc == 0, res)
		}},
		{"case 122", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["122"])
			r.commit("seed counter", "counter.txt", "a@\n")
			sha := r.commit("bump counter", "counter.txt", "a@\nb@\n")
			res := r.run("1", sha)
			ok(t, "case 122: an at-bearing command is measured, not truncated", res.rc == 1, res)
			ok(t, "case 122: names the real measured counts", has(res.out, "measures before=1 after=2"), res)
		}},
		{"case 123", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["123"])
			r.commit("seed", "alpha.txt", "@Test fun a() {}\n@Test fun b() {}\n@Test fun c() {}\n", "counter.txt", "5\n")
			sha := r.commit("bump counter", "alpha.txt", "@Test fun a() {}\n@Test fun b() {}\n@Test fun c() {}\n# prose only\n", "counter.txt", "6\n")
			res := r.run("1", sha)
			ok(t, "case 123: a measured-and-matching Baseline passes despite a static @Test delta mismatch", res.rc == 0, res)
		}},
		// The harness pinned MEASURED_TIMEOUT_SECONDS to 1 in-process and
		// called _run_measured_at directly; the Go seam is the same call with
		// its ceiling as an argument.
		// The throwaway worktree goes under the guard's Env TMPDIR, never the
		// process's (panel round 2, F24): the recorded command prints 1 only
		// when it runs inside <TMPDIR>/ctcf-measured-*.
		{"case 124a", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			sha := r.commit("second root", "root2.txt", "root2\n")
			tmp, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			env := Env{Getenv: func(k string) string {
				if k == "TMPDIR" {
					return tmp
				}
				return os.Getenv(k)
			}}
			cmd := `case "$(pwd -P)" in '` + tmp + `'/ctcf-measured-*) echo 1 ;; *) echo 0 ;; esac`
			n, measured, err := tcfRunMeasuredAt(env, r.dir, sha, cmd, 10*time.Second)
			res := tcfRes{0, fmt.Sprintf("measured=%v n=%v err=%v", measured, n, err)}
			ok(t, "case 124a: the measured worktree is made under the guard's TMPDIR",
				err == nil && measured && n.Int64() == 1, res)
		}},
		{"case 124", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			sha := r.commit("second root", "root2.txt", "root2\n")
			start := time.Now()
			_, measured, err := tcfRunMeasuredAt(Env{Getenv: r.getenv()}, r.dir, sha, "sleep 30", time.Second)
			res := tcfRes{0, "measured=" + map[bool]string{true: "yes", false: "no"}[measured]}
			ok(t, "case 124: a recorded command past the ceiling skips and the guard returns",
				err == nil && !measured && time.Since(start) < 15*time.Second, res)
		}},
		{"case 125-129", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["125"])
			shaAlpha := r.commit("add alpha", "alpha.txt", "alpha\n")
			shaBeta := r.commit("add beta", "beta.txt", "beta\n")
			res := r.run("1 2", shaBeta)
			ok(t, "case 125: a joined multi-task id refuses with rc=2", res.rc == 2, res)
			ok(t, "case 125: names the caller mistake and the per-task rerun", has(res.out, "COULD NOT JUDGE", "one call per task"), res)
			ok(t, "case 125: never reads as the tasks missing from the plan", !has(res.out, "not found"), res)
			res = r.run("", shaBeta)
			ok(t, "case 126: an empty task id refuses with rc=2", res.rc == 2, res)
			ok(t, "case 126: COULD NOT JUDGE naming the caller mistake", has(res.out, "COULD NOT JUDGE", "one call per task"), res)
			ok(t, "case 126: never reads as the tasks missing from the plan", !has(res.out, "not found"), res)
			loop := true
			for _, spec := range [][2]string{{"1", shaAlpha}, {"2", shaBeta}} {
				res = r.run(spec[0], spec[1])
				if res.rc != 0 || has(res.out, "not found") {
					loop = false
					break
				}
			}
			ok(t, "case 127: one call per task judges each task on its own commit", loop, res)
			res = r.run("1.2", shaBeta)
			ok(t, "case 128: a dotted task id refuses with rc=2", res.rc == 2, res)
			ok(t, "case 128: names the dotted id and its flat-id remediation", has(res.out, "COULD NOT JUDGE", "use the task's own flat integer id"), res)
			ok(t, "case 128: never reads as the tasks missing from the plan", !has(res.out, "not found"), res)
			res = r.run("01", shaBeta)
			ok(t, "case 129: a leading-zero task id refuses with rc=2", res.rc == 2, res)
			ok(t, "case 129: names the caller mistake and the per-task rerun", has(res.out, "COULD NOT JUDGE", "one call per task"), res)
			ok(t, "case 129: never reads as the tasks missing from the plan", !has(res.out, "not found"), res)
		}},
		{"case 130", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["130"])
			sha := r.commit("add alpha only", "alpha.txt", "# test_alpha covers alpha\n")
			res := r.run("130", sha)
			ok(t, "case 130: a declared file the commit never touches fails", res.rc == 1, res)
			ok(t, "case 130: names the declared file the commit skips", has(res.out, "ghost.txt", "does not touch"), res)
		}},
		{"case 131", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["131"])
			sha := r.commit("swap the bar baseline", "j4-bar-overflow-chip-darwin.png", "test_bar\n", "j4-bar-after-roster-select-darwin.png", "y\n")
			res := r.run("131", sha)
			ok(t, "case 131: the kan-30 shape fails", res.rc == 1, res)
			ok(t, "case 131: names the declared baseline the commit skips", has(res.out, "one-face", "does not touch"), res)
			ok(t, "case 131: still names the undeclared baseline the commit touches", has(res.out, "overflow-chip", "not declared"), res)
		}},
		{"case 132", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["132"])
			sha := r.commit("feat: add alpha and beta", "alpha.txt", "def test_alpha(): pass\n", "beta.txt", "b\n")
			res := r.run("1", sha)
			ok(t, "case 132: the red id reports the fold's untouched declaration", res.rc == 1, res)
			ok(t, "case 132: names the untouched file from the red id", has(res.out, "ghost.txt", "does not touch"), res)
			res = r.run("2", sha)
			ok(t, "case 132: the partner id reaches the same verdict", res.rc == 1, res)
			ok(t, "case 132: names the untouched file from the partner id too", has(res.out, "ghost.txt", "does not touch"), res)
		}},
		{"case 133", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.commit("seed both files", "alpha.txt", "a\n", "beta.txt", "# test_alpha lives on\n")
			r.plan(tcfPlans["133"])
			r.git("rm", "-q", "beta.txt")
			sha := r.commit("drop beta keep alpha", "alpha.txt", "# test_alpha covers alpha\n")
			res := r.run("133", sha)
			ok(t, "case 133: a deleted declared file counts as touched", res.rc == 0, res)
		}},
		{"case 134", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.commit("seed alpha", "alpha.txt", "a\n")
			r.plan(tcfPlans["134"])
			sha := r.commit("fix: widen alpha", "alpha.txt", "a wider\n")
			res := r.run("134", sha)
			ok(t, "case 134: Decision after Commit does not join the subject", res.rc == 0, res)
			ok(t, "case 134: no subject mismatch is reported", !has(res.out, "does not match declared Commit"), res)
		}},
		{"case 135", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["135"])
			sha := r.commit("feat: add thing", "src/main/kotlin/com/gymie/gs/core/Thing.kt", "class Thing\n", "src/test/kotlin/com/gymie/gs/core/ThingTest.kt", "class ThingTest\n")
			res := r.run("135", sha)
			ok(t, "case 135: a preamble shorthand legend expands against Files", res.rc == 0, res)
		}},
		{"case 136", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["136"])
			sha := r.commit("feat: add thing", "src/main/kotlin/com/gymie/gs/core/Thing.kt", "class Thing\n")
			res := r.run("136", sha)
			ok(t, "case 136: an abbreviation with no legend stays literal and fails", res.rc == 1, res)
			ok(t, "case 136: names the untouched shorthand declaration", has(res.out, "gs/core/Thing.kt", "does not touch"), res)
		}},
		{"case 137", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["137"])
			sha := r.commit("add wrapped sentence", "alpha.txt", "This doc comment states that the guard folds wrapped\nlines before matching them, in prose.\n")
			res := r.run("137", sha)
			ok(t, "case 137: wrapped declared sentence passes", res.rc == 0, res)
		}},
		{"case 138", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["138"])
			sha := r.commit("add alpha only", "alpha.txt", "nothing to see here\n")
			res := r.run("138", sha)
			ok(t, "case 138: absent declared sentence fails", res.rc == 1, res)
			ok(t, "case 138: names the missing sentence", has(res.out, "a sentence nowhere in this commit", "not found in the diff"), res)
		}},
		{"case 139", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["139"])
			sha := r.commit("add export-ignored sentence", ".gitattributes", "packed.txt export-ignore\n",
				"packed.txt", "notes: the tree search sees every searchable\nblob it lists, packed or not.\n")
			res := r.run("139", sha)
			ok(t, "case 139: export-ignored blob still searched", res.rc == 0, res)
		}},
		{"case 140", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["140"])
			sha := r.commit("add alpha", "alpha.txt", "# test_alpha covers alpha\n")
			res := r.run("140", sha)
			ok(t, "case 140: evidence-free verified tag fails", res.rc == 1, res)
			ok(t, "case 140: names the task and the evidence rule", has(res.out, "task 140", "carries no evidence"), res)
		}},
		{"case 141", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["141"])
			sha := r.commit("add alpha", "alpha.txt", "# test_alpha covers alpha\n")
			res := r.run("141", sha)
			ok(t, "case 141: evidence-free measured comment fails", res.rc == 1, res)
			ok(t, "case 141: names the line and the comment", has(res.out, "measured: comment at tasks.md line", "carries no evidence"), res)
		}},
		{"case 142", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["142"])
			sha := r.commit("add alpha", "alpha.txt", "# test_alpha covers alpha\n")
			res := r.run("142", sha)
			ok(t, "case 142: tags with evidence pass", res.rc == 0, res)
		}},
		{"case 143", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["143"])
			sha := r.commit("add alpha", "alpha.txt", "# test_alpha covers alpha\n")
			res := r.run("143", sha)
			ok(t, "case 143: fence content comment not flagged", res.rc == 0, res)
		}},
		{"case 144", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["144"])
			sha := r.commit("add alpha", "alpha.txt", "# test_alpha covers alpha\n")
			res := r.run("144", sha)
			ok(t, "case 144: language-less fence bare tag fails", res.rc == 1, res)
		}},
		{"case 145", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["145"])
			sha := r.commit("add alpha", "alpha.txt", "# test_alpha covers alpha\n")
			res := r.run("145", sha)
			ok(t, "case 145: empty unverified fence fails", res.rc == 1, res)
			ok(t, "case 145: names the unverified tag", has(res.out, "unverified:", "carries no evidence"), res)
		}},
		{"case 146", func(t *testing.T) {
			r := fx.repo(t, "change-a")
			r.plan(tcfPlans["146"])
			sha := r.commit("add alpha", "alpha.txt", "# test_alpha covers alpha\n")
			res := r.run("146", sha)
			ok(t, "case 146: empty predicted comment fails", res.rc == 1, res)
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			c.fn(t)
		})
	}
}

// TestTaskFieldParseMatchesPython runs scripts/check-task-commit-fields.py's
// own parse_task_fields -- loaded the way check-task-records.py loads it --
// once over every plan fixture above, and fails where the Go parse of any
// task differs. check-task-records.py and check-plan-shape.py still read the
// Python parse, so a drift here is two guards disagreeing about one plan.
func TestTaskFieldParseMatchesPython(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("python3"); err != nil {
		t.Fatal("python3 not on PATH: a hard dependency of the live check-task-records.py, so this drift check must run")
	}
	dir := t.TempDir()
	keys := make([]string, 0, len(tcfPlans))
	for k := range tcfPlans {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var paths []string
	for _, k := range keys {
		p := dir + "/" + k + "/tasks.md"
		writeFile(t, p, tcfPlans[k])
		paths = append(paths, p)
	}
	const script = `
import importlib.util, json, sys
from pathlib import Path

scripts = Path(sys.argv[1])
sys.path.insert(0, str(scripts / "lib"))
spec = importlib.util.spec_from_file_location("check_task_commit_fields", str(scripts / "check-task-commit-fields.py"))
cff = importlib.util.module_from_spec(spec)
sys.modules["check_task_commit_fields"] = cff
spec.loader.exec_module(cff)

out = {}
for path in sys.argv[2:]:
    with open(path, "r", encoding="utf-8") as handle:
        lines = handle.read().splitlines()
    tasks = []
    for task_id in cff.collect_task_ids(lines):
        f = cff.parse_task_fields(lines, task_id)
        tasks.append({
            "id": f.id, "files": f.files, "tests": [s.label for s in f.tests],
            "tests_value": f.tests_value, "tree_names": cff._extract_tree_names(f.tests_value),
            "allowed_collateral": f.allowed_collateral, "commit": f.commit,
            "baseline": list(f.baseline) if f.baseline else None,
            "baseline_measured": f.baseline_measured, "build": f.build,
            "squash_value": f.squash_value, "squash_partners": f.squash_partners,
            "unclosed_fence_line": f.unclosed_fence_line,
        })
    out[path] = tasks
json.dump(out, sys.stdout)
`
	cmd := exec.Command("python3", append([]string{"-c", script, tcfScriptsDir(t)}, paths...)...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	pyOut, err := cmd.Output()
	if err != nil {
		t.Fatalf("python3: %v\n%s", err, stderr.String())
	}
	var want map[string]any
	if err := json.Unmarshal(pyOut, &want); err != nil {
		t.Fatal(err)
	}
	for _, p := range paths {
		body, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		lines, err := tcfPlanLines(body)
		if err != nil {
			t.Fatal(err)
		}
		var tasks []map[string]any
		for _, id := range tcfTaskIDs(lines) {
			f, _ := tcfParseTask(lines, id)
			var baseline, build, fence any
			if f.baseline != nil {
				baseline = []any{f.baseline[0], f.baseline[1]}
			}
			if f.build != "" {
				build = f.build
			}
			if f.unclosedFenceLine != 0 {
				fence = f.unclosedFenceLine
			}
			var tests []string
			for _, s := range f.tests {
				tests = append(tests, s.label)
			}
			tasks = append(tasks, map[string]any{
				"id": f.id, "files": f.files, "tests": tests,
				"tests_value": f.testsValue, "tree_names": tcfTreeNames(f.testsValue),
				"allowed_collateral": f.allowedCollateral, "commit": f.commit,
				"baseline": baseline, "baseline_measured": f.baselineMeasured, "build": build,
				"squash_value": f.squashValue, "squash_partners": f.squashPartners,
				"unclosed_fence_line": fence,
			})
		}
		// A JSON round trip gives both sides the same shapes: nil and empty
		// lists, *string and string, big.Int and float64.
		raw, err := json.Marshal(tasks)
		if err != nil {
			t.Fatal(err)
		}
		var got any
		if err := json.Unmarshal(raw, &got); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(tcfNullEmpty(got), tcfNullEmpty(want[p])) {
			t.Errorf("%s: Go parse differs from Python\n go: %s\n py: %s", p, raw, tcfJSON(want[p]))
		}
	}
}

// tcfNullEmpty maps an empty list to nil, so Python's [] and a nil Go slice
// compare equal; every other value is unchanged.
func tcfNullEmpty(v any) any {
	switch x := v.(type) {
	case []any:
		if len(x) == 0 {
			return nil
		}
		out := make([]any, len(x))
		for i := range x {
			out[i] = tcfNullEmpty(x[i])
		}
		return out
	case map[string]any:
		out := map[string]any{}
		for k, e := range x {
			out[k] = tcfNullEmpty(e)
		}
		return out
	}
	return v
}

func tcfJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// TestTcfFnmatchMatchesPython runs Python's own fnmatch over glob shapes
// whose bracket expressions Go's regexp and Python's re read differently.
func TestTcfFnmatchMatchesPython(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("python3"); err != nil {
		t.Fatal("python3 not on PATH: a hard dependency of the live check-task-records.py, so this drift check must run")
	}
	patterns := []string{"docs/[x[:alpha:]*[y]", "[[:alpha:]]", "a[[]b", "[![:digit:]]x", "*[[]*", "d[a-c[]"}
	names := []string{"docs/b", "docs/x", "docs/[y", "a", ":", "[", "a[b", "5x", "bx", "[x", "da", "d[", "dz"}
	var script strings.Builder
	script.WriteString("import fnmatch\n")
	for _, p := range patterns {
		for _, n := range names {
			fmt.Fprintf(&script, "print(int(fnmatch.fnmatch(%q, %q)))\n", n, p)
		}
	}
	out, err := exec.Command("python3", "-c", script.String()).Output()
	if err != nil {
		t.Fatal(err)
	}
	want := strings.Fields(string(out))
	i := 0
	for _, p := range patterns {
		for _, n := range names {
			got := "0"
			if tcfFnmatch(n, p) {
				got = "1"
			}
			if got != want[i] {
				t.Errorf("fnmatch(%q, %q): go %s, python %s", n, p, got, want[i])
			}
			i++
		}
	}
}
