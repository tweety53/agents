package lspmux

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// StopUnder stops every language server a worktree-lsp wrapper runs at or
// under dir: each process whose working directory is dir or below it and
// whose parent's command is worktree-lsp gets SIGTERM, then SIGKILL if it
// is still running after stopGrace. /flow cleanup calls it before its
// process check (design.md, stop-at-cleanup-and-removal), so a session's
// servers never hold a worktree that is about to be removed. Paths are
// compared in their physical form, as check-worktree-processes.sh compares
// them. Only the invoking user's processes are visible to lsof.
func StopUnder(dir string) error {
	phys, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return err
	}
	out, err := exec.Command("lsof", "-d", "cwd", "-FpRn").Output()
	if err != nil {
		return fmt.Errorf("lsof -d cwd -FpRn: %w", err)
	}
	var pids []int
	pid, ppid := 0, 0
	for _, line := range bytes.Split(out, []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		v := string(line[1:])
		switch line[0] {
		case 'p':
			pid, _ = strconv.Atoi(v)
		case 'R':
			ppid, _ = strconv.Atoi(v)
		case 'n':
			if (v == phys || strings.HasPrefix(v, phys+"/")) && isWrapper(ppid) {
				pids = append(pids, pid)
			}
		}
	}
	for _, p := range pids {
		_ = syscall.Kill(p, syscall.SIGTERM)
	}
	deadline := time.Now().Add(stopGrace)
	for _, p := range pids {
		// ponytail: polled — p is not our child, so there is nothing to wait on.
		for processExists(p) && time.Now().Before(deadline) {
			time.Sleep(20 * time.Millisecond)
		}
		if processExists(p) {
			_ = syscall.Kill(p, syscall.SIGKILL)
		}
	}
	return nil
}

// isWrapper reports whether pid's command is worktree-lsp — ps prints
// argv[0], which for the built wrapper the plugin execs is its full path.
func isWrapper(pid int) bool {
	out, err := exec.Command("ps", "-o", "comm=", "-p", strconv.Itoa(pid)).Output()
	return err == nil && filepath.Base(strings.TrimSpace(string(out))) == "worktree-lsp"
}

func processExists(pid int) bool { return syscall.Kill(pid, 0) == nil }
