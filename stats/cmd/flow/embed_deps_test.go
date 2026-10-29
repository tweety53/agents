package main

import (
	"os/exec"
	"strings"
	"testing"
)

// TestFlowCLIDoesNotImportEmbeddedPolicy: modelpolicy/embedded embeds a file
// only make copies, and guards build and run cmd/flow without make
// (check-stage-mark-calls' `go run ./cmd/flow stage keys`), so no package
// cmd/flow depends on may import it (KAN-1067).
func TestFlowCLIDoesNotImportEmbeddedPolicy(t *testing.T) {
	out, err := exec.Command("go", "list", "-deps", ".").Output()
	if err != nil {
		t.Fatalf("go list -deps: %v", err)
	}
	for _, dep := range strings.Fields(string(out)) {
		if strings.HasSuffix(dep, "/internal/modelpolicy/embedded") {
			t.Fatalf("cmd/flow depends on %s; pass the table in from cmd/flowd instead", dep)
		}
	}
}
