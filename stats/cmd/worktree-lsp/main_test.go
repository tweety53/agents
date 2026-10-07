package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestUsage(t *testing.T) {
	t.Parallel()
	for _, args := range [][]string{nil, {"--"}, {"gopls"}, {"gopls", "--"}} {
		var stdout, stderr bytes.Buffer
		if code := run(context.Background(), args, strings.NewReader(""), &stdout, &stderr); code != 2 {
			t.Errorf("worktree-lsp %q exited %d, want 2", args, code)
		}
		if stderr.String() != usage {
			t.Errorf("worktree-lsp %q stderr %q, want the usage line %q", args, stderr.String(), usage)
		}
		if stdout.Len() != 0 {
			t.Errorf("worktree-lsp %q wrote %q to stdout, which carries only LSP traffic", args, stdout.String())
		}
	}
}
