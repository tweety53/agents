// Command worktree-lsp is the language server the worktree-lsp plugin
// (mods/worktree-lsp) starts for Claude Code:
//
//	worktree-lsp -- <server> [args...]
//
// It speaks LSP on stdin/stdout and runs one <server> per git worktree
// behind that single connection (internal/lspmux). Its own log lines and
// every server's stderr go to stderr. Exit 2 is a usage error, 1 a broken
// client stream, 0 a session that ended (exit, stdin closed, or a signal).
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/tweety53/agents/stats/internal/lspmux"
)

const usage = "usage: worktree-lsp -- <server> [args...]\n"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	code := run(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr)
	stop()
	os.Exit(code)
}

func run(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) < 2 || args[0] != "--" {
		fmt.Fprint(stderr, usage)
		return 2
	}
	m := &lspmux.Mux{Server: args[1:], Log: stderr}
	if err := m.Run(ctx, stdin, stdout); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintf(stderr, "worktree-lsp: %v\n", err)
		return 1
	}
	return 0
}
