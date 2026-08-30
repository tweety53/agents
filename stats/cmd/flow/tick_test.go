package main

import (
	"errors"
	"strings"
	"testing"
)

func TestTickTaskFlipsOwnCheckboxAndSteps(t *testing.T) {
	plan := "# c\n\n- [ ] 1. First\n\n  - [ ] **Step 1: do a thing**\n\n  - [ ] **Step 2: do another**\n\n**Files:** `x`\n"
	out, steps, err := tickTask([]byte(plan), "1")
	if err != nil {
		t.Fatalf("tickTask: %v", err)
	}
	if steps != 2 {
		t.Fatalf("steps = %d, want 2", steps)
	}
	got := string(out)
	if !strings.Contains(got, "- [x] 1. First") {
		t.Errorf("task line not ticked:\n%s", got)
	}
	if strings.Contains(got, "- [ ] **Step") {
		t.Errorf("a step was left unticked:\n%s", got)
	}
	if !strings.Contains(got, "**Files:** `x`") {
		t.Errorf("field line was disturbed:\n%s", got)
	}
}

func TestTickTaskZeroSteps(t *testing.T) {
	plan := "- [ ] 1. First\n\nno steps here.\n\n**Files:** `x`\n"
	out, steps, err := tickTask([]byte(plan), "1")
	if err != nil {
		t.Fatalf("tickTask: %v", err)
	}
	if steps != 0 {
		t.Fatalf("steps = %d, want 0", steps)
	}
	if !strings.Contains(string(out), "- [x] 1. First") {
		t.Errorf("task line not ticked:\n%s", string(out))
	}
}

func TestTickTaskStopsAtNextTaskLine(t *testing.T) {
	plan := "- [ ] 1. First\n\n  - [ ] **Step 1: a**\n\n- [ ] 2. Second\n\n  - [ ] **Step 1: b**\n"
	out, steps, err := tickTask([]byte(plan), "1")
	if err != nil {
		t.Fatalf("tickTask: %v", err)
	}
	if steps != 1 {
		t.Fatalf("steps = %d, want 1 (task 2's step must not be touched)", steps)
	}
	got := string(out)
	if !strings.Contains(got, "- [x] 1. First") {
		t.Errorf("task 1 not ticked:\n%s", got)
	}
	if strings.Contains(got, "- [x] 2. Second") {
		t.Errorf("task 2 was ticked, must not be:\n%s", got)
	}
	if !strings.Contains(got, "- [ ] **Step 1: b**") {
		t.Errorf("task 2's step was flipped, must not be:\n%s", got)
	}
}

func TestTickTaskStopsAtNextHeading(t *testing.T) {
	plan := "- [ ] 1. First\n\n  - [ ] **Step 1: a**\n\n## Part 2\n\n- [ ] 2. Second\n\n  - [ ] **Step 1: b**\n"
	_, steps, err := tickTask([]byte(plan), "1")
	if err != nil {
		t.Fatalf("tickTask: %v", err)
	}
	if steps != 1 {
		t.Fatalf("steps = %d, want 1 (must stop at the ## heading)", steps)
	}
}

func TestTickTaskAlreadyTickedRefuses(t *testing.T) {
	plan := "- [x] 1. First\n\n  - [ ] **Step 1: a**\n"
	_, _, err := tickTask([]byte(plan), "1")
	if !errors.Is(err, ErrTaskAlreadyTicked) {
		t.Fatalf("err = %v, want ErrTaskAlreadyTicked", err)
	}
}

func TestTickTaskNotFound(t *testing.T) {
	plan := "- [ ] 1. First\n"
	_, _, err := tickTask([]byte(plan), "2")
	if !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("err = %v, want ErrTaskNotFound", err)
	}
}

func TestTickTaskInvalidID(t *testing.T) {
	plan := "- [ ] 1. First\n"
	for _, id := range []string{"1.1", "one", "-1", "", "1 "} {
		if _, _, err := tickTask([]byte(plan), id); !errors.Is(err, ErrTaskIDInvalid) {
			t.Errorf("id %q: err = %v, want ErrTaskIDInvalid", id, err)
		}
	}
}

func TestFenceTruncatesBody(t *testing.T) {
	plan := "- [ ] 1. First\n\n  - [ ] **Step 1: real a**\n\n```\n- [ ] 99. fake, inside a fence\n```\n\n  - [ ] **Step 2: real b, should be ticked**\n"
	out, steps, _ := tickTask([]byte(plan), "1")
	if steps != 2 {
		t.Fatalf("steps = %d, want 2 (fenced fake task line truncated the body): %s", steps, out)
	}
}

func TestTickTaskStopsAtBareHeadingWithNoFollowingTask(t *testing.T) {
	plan := "- [ ] 1. First\n\n  - [ ] **Step 1: a**\n\n## Part 2\n\nsome prose, no more tasks\n"
	_, steps, err := tickTask([]byte(plan), "1")
	if err != nil {
		t.Fatalf("tickTask: %v", err)
	}
	if steps != 1 {
		t.Fatalf("steps = %d, want 1 (must stop at the heading even with no following task line)", steps)
	}
}

func TestTickTaskStopsAtHeadingWithNothingAfterTheHashes(t *testing.T) {
	// "##" with nothing following it (no space, no text) is a heading
	// BODY_BOUNDARY_RE closes on too -- the old headingPattern required a
	// literal trailing space and would have missed this line entirely.
	// No task line follows either, so a missed boundary here can only be
	// caught by a step-checkbox-shaped line past the heading being
	// wrongly ticked -- with a following task line present, the task
	// line alone would still correctly bound the scan and mask the bug.
	plan := "- [ ] 1. First\n\n  - [ ] **Step 1: a**\n\n##\n\n  - [ ] should not count, past a heading\n"
	_, steps, err := tickTask([]byte(plan), "1")
	if err != nil {
		t.Fatalf("tickTask: %v", err)
	}
	if steps != 1 {
		t.Fatalf("steps = %d, want 1 (must stop at a bare \"##\" heading)", steps)
	}
}

func TestTickTaskRequiresSpaceAfterDot(t *testing.T) {
	plan := "- [ ] 1.NoSpace\n"
	_, _, err := tickTask([]byte(plan), "1")
	if !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("err = %v, want ErrTaskNotFound (missing space after the dot means no task line matches)", err)
	}
}

func TestFenceSuppressesStepTicking(t *testing.T) {
	plan := "- [ ] 1. First\n\n  - [ ] **Step 1: real**\n\n```\n  - [ ] fake step inside fence\n```\n"
	out, steps, err := tickTask([]byte(plan), "1")
	if err != nil {
		t.Fatalf("tickTask: %v", err)
	}
	if steps != 1 {
		t.Fatalf("steps = %d, want 1 (the fenced fake step must not be ticked)", steps)
	}
	if strings.Contains(string(out), "[x] fake step") {
		t.Errorf("fenced fake step was ticked, must not be:\n%s", out)
	}
}

func TestFourSpaceIndentIsNotAFence(t *testing.T) {
	// A "```" indented by four or more spaces is CommonMark's own
	// fence-indent boundary -- an indented code block, not a fence -- so
	// it must not toggle fence state. If it wrongly did, step 2 (which
	// sits between two such lines) would be wrongly suppressed as if
	// inside a fence.
	plan := "- [ ] 1. First\n\n  - [ ] **Step 1: a**\n\n    ```\n\n  - [ ] **Step 2: b**\n\n    ```\n"
	out, steps, err := tickTask([]byte(plan), "1")
	if err != nil {
		t.Fatalf("tickTask: %v", err)
	}
	if steps != 2 {
		t.Fatalf("steps = %d, want 2 (4-space indented \"```\" is not a fence, must not suppress step 2)", steps)
	}
	if strings.Contains(string(out), "- [ ] **Step 2") {
		t.Errorf("step 2 was not ticked:\n%s", out)
	}
}

func TestLevelOneHeadingDoesNotCloseBody(t *testing.T) {
	plan := "- [ ] 1. First\n\n  - [ ] **Step 1: a**\n\n# Not a boundary\n\n  - [ ] **Step 2: b**\n"
	_, steps, err := tickTask([]byte(plan), "1")
	if err != nil {
		t.Fatalf("tickTask: %v", err)
	}
	if steps != 2 {
		t.Fatalf("steps = %d, want 2 (a level-1 heading must not close the task's body)", steps)
	}
}

func TestTickTaskLeavesRestByteIdentical(t *testing.T) {
	plan := "# Title\n\nprose before.\n\n- [ ] 1. First\n\n  - [ ] **Step 1: a**\n\n- [ ] 2. Second\n\nprose after.\n"
	out, _, err := tickTask([]byte(plan), "1")
	if err != nil {
		t.Fatalf("tickTask: %v", err)
	}
	want := "# Title\n\nprose before.\n\n- [x] 1. First\n\n  - [x] **Step 1: a**\n\n- [ ] 2. Second\n\nprose after.\n"
	if string(out) != want {
		t.Fatalf("out =\n%s\nwant\n%s", out, want)
	}
}

func TestStepCheckboxRequiresExactlyTwoSpaceIndent(t *testing.T) {
	plan := "- [ ] 1. First\n\n - [ ] wrong indent, not a step\n\n  - [ ] **Step 1: real**\n"
	out, steps, err := tickTask([]byte(plan), "1")
	if err != nil {
		t.Fatalf("tickTask: %v", err)
	}
	if steps != 1 {
		t.Fatalf("steps = %d, want 1 (the 1-space-indented line must not count as a step)", steps)
	}
	if strings.Contains(string(out), "[x] wrong indent") {
		t.Errorf("wrongly-indented line was ticked, must not be:\n%s", out)
	}
}

func TestStepCheckboxIgnoresAlreadyTickedSteps(t *testing.T) {
	plan := "- [ ] 1. First\n\n  - [x] **Step 1: already done**\n\n  - [ ] **Step 2: new**\n"
	out, steps, err := tickTask([]byte(plan), "1")
	if err != nil {
		t.Fatalf("tickTask: %v", err)
	}
	if steps != 1 {
		t.Fatalf("steps = %d, want 1 (the already-ticked step must not be recounted)", steps)
	}
	if !strings.Contains(string(out), "- [x] **Step 1: already done**") {
		t.Errorf("already-ticked step line was disturbed:\n%s", out)
	}
}
