package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writePlan(t *testing.T, dir, changesLeaf, change, body string) {
	t.Helper()
	planDir := filepath.Join(dir, changesLeaf, "changes", change)
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(planDir, "tasks.md"), []byte(body), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

func TestTasksTickWritesFile(t *testing.T) {
	dir := t.TempDir()
	writePlan(t, dir, "spectre", "c", "- [ ] 1. First\n\n  - [ ] **Step 1: a**\n")

	var stdout, stderr bytes.Buffer
	code := run(context.Background(), []string{"tasks", "tick", "-C", dir, "c", "1"}, nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "ticked task 1 (1 steps)") {
		t.Errorf("stdout = %q", stdout.String())
	}
	got, err := os.ReadFile(filepath.Join(dir, "spectre", "changes", "c", "tasks.md"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !strings.Contains(string(got), "- [x] 1. First") || !strings.Contains(string(got), "- [x] **Step 1: a**") {
		t.Errorf("file not ticked:\n%s", got)
	}
}

func TestTasksTickAlreadyTickedExitsNonZero(t *testing.T) {
	dir := t.TempDir()
	writePlan(t, dir, "spectre", "c", "- [x] 1. First\n")

	var stdout, stderr bytes.Buffer
	code := run(context.Background(), []string{"tasks", "tick", "-C", dir, "c", "1"}, nil, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("code = 0, want non-zero; stderr = %s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "already ticked") {
		t.Errorf("stderr = %q", stderr.String())
	}
}

func TestTasksTickTaskNotFoundExitsNonZero(t *testing.T) {
	dir := t.TempDir()
	writePlan(t, dir, "spectre", "c", "- [ ] 1. First\n")

	var stdout, stderr bytes.Buffer
	code := run(context.Background(), []string{"tasks", "tick", "-C", dir, "c", "9"}, nil, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("code = 0, want non-zero; stderr = %s", stderr.String())
	}
}

func TestTasksTickInvalidIDExitsTwo(t *testing.T) {
	dir := t.TempDir()
	writePlan(t, dir, "spectre", "c", "- [ ] 1. First\n")

	var stdout, stderr bytes.Buffer
	code := run(context.Background(), []string{"tasks", "tick", "-C", dir, "c", "1.1"}, nil, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("code = %d, want 2; stderr = %s", code, stderr.String())
	}
}

func TestTasksTickMissingPlanExitsNonZero(t *testing.T) {
	dir := t.TempDir()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(), []string{"tasks", "tick", "-C", dir, "c", "1"}, nil, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("code = 0, want non-zero; stderr = %s", stderr.String())
	}
}

func TestTasksTickBadArgCountExitsTwo(t *testing.T) {
	dir := t.TempDir()
	var stdout, stderr bytes.Buffer
	code := run(context.Background(), []string{"tasks", "tick", "-C", dir, "c"}, nil, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("code = %d, want 2; stderr = %s", code, stderr.String())
	}
}

func TestPlanPathRejectsContainmentEscapes(t *testing.T) {
	dir := t.TempDir()
	for _, change := range []string{"../escape", ".hidden", "-flag", "_underscore", ""} {
		if _, err := planPath(dir, change); err == nil {
			t.Errorf("planPath(%q) = nil error, want rejection", change)
		}
	}
}

func TestPlanPathAcceptsValidName(t *testing.T) {
	dir := t.TempDir()
	got, err := planPath(dir, "valid-name-123")
	if err != nil {
		t.Fatalf("planPath: %v", err)
	}
	want := filepath.Join(dir, "spectre", "changes", "valid-name-123", "tasks.md")
	if got != want {
		t.Errorf("planPath = %q, want %q", got, want)
	}
}

func TestTasksTickRejectsContainmentEscapesEndToEnd(t *testing.T) {
	dir := t.TempDir()
	for _, change := range []string{"../escape", ".hidden", "-flag", "_underscore"} {
		var stdout, stderr bytes.Buffer
		code := run(context.Background(), []string{"tasks", "tick", "-C", dir, change, "1"}, nil, &stdout, &stderr)
		if code != 2 {
			t.Errorf("change %q: code = %d, want 2; stderr = %s", change, code, stderr.String())
		}
	}
}

func TestSpecRootLeafPrefersSpectre(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "spectre", "changes"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "openspec", "changes"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := specRootLeaf(dir); got != "spectre" {
		t.Errorf("specRootLeaf = %q, want spectre", got)
	}
}

func TestSpecRootLeafFallsBackToOpenspec(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "openspec", "changes"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := specRootLeaf(dir); got != "openspec" {
		t.Errorf("specRootLeaf = %q, want openspec", got)
	}
}

func TestSpecRootLeafDefaultsToSpectreWhenNeitherExists(t *testing.T) {
	dir := t.TempDir()
	if got := specRootLeaf(dir); got != "spectre" {
		t.Errorf("specRootLeaf = %q, want spectre", got)
	}
}

func TestSpecRootLeafIgnoresFileAtSpectrePath(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "spectre"), 0o755); err != nil {
		t.Fatal(err)
	}
	// A plain file at <dir>/spectre/changes must not count as the
	// spectre tree existing -- only a directory does.
	if err := os.WriteFile(filepath.Join(dir, "spectre", "changes"), []byte("not a dir"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "openspec", "changes"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := specRootLeaf(dir); got != "openspec" {
		t.Errorf("specRootLeaf = %q, want openspec", got)
	}
}

func TestTasksTickHelpExitsZero(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run(context.Background(), []string{"tasks", "tick", "-h"}, nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("code = %d, want 0; stderr = %s", code, stderr.String())
	}
}

func TestTasksTickExtraPositionalArgExitsTwo(t *testing.T) {
	dir := t.TempDir()
	writePlan(t, dir, "spectre", "c", "- [ ] 1. First\n")

	var stdout, stderr bytes.Buffer
	code := run(context.Background(), []string{"tasks", "tick", "-C", dir, "c", "1", "bogus-extra-arg"}, nil, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("code = %d, want 2; stderr = %s", code, stderr.String())
	}
}

func TestPlanPathRejectsTrailingCharsAfterValidPrefix(t *testing.T) {
	dir := t.TempDir()
	for _, change := range []string{"valid/x", "valid/../../etc/passwd"} {
		if _, err := planPath(dir, change); err == nil {
			t.Errorf("planPath(%q) = nil error, want rejection", change)
		}
	}
}
