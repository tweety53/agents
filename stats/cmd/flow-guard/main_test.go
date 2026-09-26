package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestUnknownGuardExits2(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	if code := run([]string{"no-such-guard"}, &stdout, &stderr); code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "unknown guard no-such-guard") {
		t.Fatalf("stderr = %q, want it to name the unknown guard", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestNoGuardNameExits2(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	if code := run(nil, &stdout, &stderr); code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "usage: flow-guard <name> [arguments]") {
		t.Fatalf("stderr = %q, want the usage", stderr.String())
	}
}

func TestCannotAnswerIsTheGuardsOwnCode(t *testing.T) {
	t.Parallel()
	for name, want := range map[string]int{
		"run-reproducer":           4,
		"check-task-commit-fields": 2,
		"check-cleanup-complete":   2,
	} {
		if got := cannotAnswer(name); got != want {
			t.Errorf("cannotAnswer(%q) = %d, want %d", name, got, want)
		}
	}
}
