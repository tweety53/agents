// Command flow's prompttuning.go implements `flow prompt-tuning render
// <case-dir>`: it renders a prompt-tuning case folder's index.html from its
// case.json and repro files (internal/prompttuning), so the page is never
// written by hand and check-prompt-tuning-cases can compare it byte for byte.
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/tweety53/agents/stats/internal/prompttuning"
)

const promptTuningUsage = `usage: flow prompt-tuning render <case-dir>

render writes <case-dir>/index.html from the folder's case.json, finding.md,
bad.diff, good.diff and task.md. Exit 0 silent, 1 on a load or render error,
2 on usage.
`

func runPromptTuning(_ context.Context, args []string, _, stderr io.Writer) int {
	if len(args) != 2 || args[0] != "render" {
		fmt.Fprint(stderr, promptTuningUsage)
		return 2
	}
	dir := args[1]
	c, f, err := prompttuning.Load(dir)
	if err == nil {
		var page []byte
		if page, err = prompttuning.Render(c, f); err == nil {
			err = os.WriteFile(filepath.Join(dir, "index.html"), page, 0o644)
		}
	}
	if err != nil {
		fmt.Fprintf(stderr, "flow: prompt-tuning render: %v\n", err)
		return 1
	}
	return 0
}
