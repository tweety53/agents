package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestUnknownGuardExits2(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	if code := run([]string{"no-such-guard"}, nil, &stdout, &stderr); code != 2 {
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
	if code := run(nil, nil, &stdout, &stderr); code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "usage: flow-guard <name> [arguments]") {
		t.Fatalf("stderr = %q, want the usage", stderr.String())
	}
}

// The block-unsafe-commands hook puts -caller-agent right after the guard
// name (scripts/flow-guard reads the name first): main drops it and exports
// the agent, so the guard sees its own arguments and every flow it runs
// carries the agent's token.
func TestCallerAgentAfterTheNameIsExported(t *testing.T) {
	t.Setenv("FLOW_CALLER_AGENT", "")
	got := callerArgs([]string{"integrate-change", "-caller-agent", "a7", "prepare", "x"})
	if strings.Join(got, " ") != "integrate-change prepare x" || os.Getenv("FLOW_CALLER_AGENT") != "a7" {
		t.Fatalf("callerArgs = %q, FLOW_CALLER_AGENT %q; want integrate-change prepare x, a7", got, os.Getenv("FLOW_CALLER_AGENT"))
	}
	if got := callerArgs([]string{"hook"}); strings.Join(got, " ") != "hook" {
		t.Fatalf("callerArgs(hook) = %q", got)
	}
}
