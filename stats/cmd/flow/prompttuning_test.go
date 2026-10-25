package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/tweety53/agents/stats/internal/prompttuning"
)

func TestPromptTuningRenderWritesIndex(t *testing.T) {
	dir := t.TempDir()
	c := prompttuning.Case{ID: "x-1", Type: "implementer", Scope: "generic", Target: "skills/flow/implement.md"}
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{
		"case.json": string(data), "finding.md": "f", "bad.diff": "b", "good.diff": "g", "task.md": "t",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	var stdout, stderr bytes.Buffer
	if code := run(context.Background(), []string{"prompt-tuning", "render", dir}, nil, &stdout, &stderr); code != 0 {
		t.Fatalf("code = %d, stderr = %s", code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout = %q, want silence", stdout.String())
	}
	got, err := os.ReadFile(filepath.Join(dir, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	lc, lf, err := prompttuning.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	want, err := prompttuning.Render(lc, lf)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Error("index.html differs from Render's output")
	}

	stderr.Reset()
	if code := run(context.Background(), []string{"prompt-tuning", "render"}, nil, &stdout, &stderr); code != 2 {
		t.Errorf("missing argument: code = %d, want 2", code)
	}
	if code := run(context.Background(), []string{"prompt-tuning", "rendr", dir}, nil, &stdout, &stderr); code != 2 {
		t.Errorf("unknown subcommand: code = %d, want 2", code)
	}
	stderr.Reset()
	if code := run(context.Background(), []string{"prompt-tuning", "render", filepath.Join(dir, "absent")}, nil, &stdout, &stderr); code != 1 {
		t.Errorf("absent folder: code = %d, want 1", code)
	}
}
