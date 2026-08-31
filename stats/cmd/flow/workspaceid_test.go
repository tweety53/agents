package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

// TestDeriveWorkspaceID covers the worked examples
// skills/flow-contracts/workspace-isolation.md's verified bash block names
// verbatim, plus one case whose first normalised segment alone exceeds 12
// characters (the second branch of the prefix rule).
func TestDeriveWorkspaceID(t *testing.T) {
	cases := []struct {
		name string
		want string
	}{
		{"kan-15-parallel-myflow-do-task-lanes", "kan-15-55a6"},
		{"Demo_X", "demo-x-5197"},
		{"KAN-99-Fix.Thing", "kan-99-fix-feef"},
		{"İstanbul-test", "--stanbul-6b8a"},
		{"averylongfirstsegmentthatexceedstwelvechars", "averylongfir-0496"},
		// Uppercase 'Z' at the top of the A-Z lowering range (deriveWorkspaceID's
		// b >= 'A' && b <= 'Z' check): catches a mutation that drops 'Z' from the
		// range (b < 'Z').
		{"FOZ", "foz-1b9e"},
		// First normalised segment exactly 12 characters, alone (prefixOf's
		// len(segments[0]) > 12 check): "abcdefghijkl-d682" is byte-identical to
		// what mutating > 12 to >= 12 would also produce for this input, so this
		// boundary is an equivalent mutant for output-based testing -- see the
		// fix-run report for the mutation-proof result.
		{"abcdefghijkl", "abcdefghijkl-d682"},
		// Two segments whose joined length lands exactly on 12 (prefixOf's
		// accumulation loop's next > 12 check): catches a mutation to next >= 12,
		// which would exclude the second segment and yield "ab-<digest>" instead.
		{"ab-cdefghijk", "ab-cdefghijk-1c3e"},
	}
	for _, c := range cases {
		got := deriveWorkspaceID(c.name)
		if got != c.want {
			t.Errorf("deriveWorkspaceID(%q) = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestWorkspaceIDCommandPrintsID(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run(context.Background(), []string{"workspace-id", "kan-15-parallel-myflow-do-task-lanes"}, nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("code = %d, stderr = %s", code, stderr.String())
	}
	if stdout.String() != "kan-15-55a6\n" {
		t.Errorf("stdout = %q, want %q", stdout.String(), "kan-15-55a6\n")
	}
}

func TestWorkspaceIDCommandNoArgExitsNonZero(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run(context.Background(), []string{"workspace-id"}, nil, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("code = 0, want non-zero; stdout = %s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "usage:") {
		t.Errorf("stderr = %q, want a usage line", stderr.String())
	}
}

// TestWorkspaceIDCommandAcceptsLeadingDashName covers a change name beginning
// with '-' -- including a name normalising to nothing but separators, which
// workspace-isolation.md's "The workspace id" explicitly allows -- being
// taken as the positional change name rather than rejected as an
// unrecognized flag.
func TestWorkspaceIDCommandAcceptsLeadingDashName(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run(context.Background(), []string{"workspace-id", "---"}, nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("code = %d, stderr = %s", code, stderr.String())
	}
	want := deriveWorkspaceID("---") + "\n"
	if stdout.String() != want {
		t.Errorf("stdout = %q, want %q", stdout.String(), want)
	}
}

func TestWorkspaceIDCommandTooManyArgsExitsNonZero(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run(context.Background(), []string{"workspace-id", "a", "b"}, nil, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("code = 0, want non-zero; stdout = %s", stdout.String())
	}
}
