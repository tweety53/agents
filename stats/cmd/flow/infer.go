package main

// infer.go — the change a `flow` call names none of, inferred from the
// call's directory, so
// a skill's command line carries only what is not already known. An
// explicit flag always wins; an inference that cannot be made leaves the
// value empty, and the caller's own "required" check reports it.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// inferChange is the change the checkout holding dir works on: its branch,
// less a `spectre/` prefix, when that checkout carries the change's
// directory (.superpowers/sdd/<name> or spectre/changes/<name>) — the same
// evidence requireChangeCheckout accepts. "" anywhere else, the main
// checkout on its base branch included.
func inferChange(dir string) string {
	git := func(a ...string) string {
		out, err := exec.Command("git", append([]string{"-C", dir}, a...)...).Output()
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(out))
	}
	top, branch := git("rev-parse", "--show-toplevel"), git("symbolic-ref", "--short", "-q", "HEAD")
	name := strings.TrimPrefix(branch, "spectre/")
	if top == "" || name == "" || strings.Contains(name, "/") {
		return ""
	}
	for _, root := range []string{".superpowers/sdd", "spectre/changes"} {
		if fi, err := os.Stat(filepath.Join(top, root, name)); err == nil && fi.IsDir() {
			return name
		}
	}
	return ""
}
