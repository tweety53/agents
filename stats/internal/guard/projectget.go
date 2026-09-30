package guard

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// projectGet is scripts/project-get.sh: that script's header comment is the
// contract, and TestProjectGetParity pins this port against the bash it
// replaced (ae805186) on stdout, stderr and exit code.
func init() {
	Registry["project-get"] = projectGet
}

func projectGet(args []string, env Env, stdout, stderr io.Writer) int {
	// --enum <literal>... resolves a single-line-literal key: the body's
	// head is matched byte-for-byte against the literals.
	enum := len(args) > 3 && args[2] == "--enum"
	if len(args) != 2 && !enum {
		fmt.Fprintln(stderr, "usage: project-get.sh <project-root> <key>")
		return 2
	}
	root, key := args[0], args[1]
	if !isDir(smcAbs(env, root)) {
		fmt.Fprintf(stderr, "project-get: %s is not a directory\n", root)
		return 2
	}
	cfg := root + "/.flow/project.md"
	src := smcAbs(env, cfg)

	// HEAD's copy wins over the working tree's (KAN-520, the script's
	// header), read through a temporary file as the bash did. A caller's
	// ambient GIT_DIR/GIT_WORK_TREE/GIT_INDEX_FILE would hijack which
	// repository git reads, so none reaches it.
	git := pgGit(env)
	if rel, ok := capture(git("-C", root, "rev-parse", "--show-prefix")); ok {
		obj := "HEAD:" + rel + ".flow/project.md"
		if git("-C", root, "cat-file", "-e", obj).Run() == nil {
			tmp, err := pgHead(git("-C", root, "show", obj))
			if err != nil {
				fmt.Fprintf(stderr, "project-get: cannot read %s in %s: %v\n", obj, root, err)
				return 2
			}
			defer os.Remove(tmp)
			if isRegular(src) && projectSection(tmp, key) != projectSection(src, key) {
				fmt.Fprintf(stderr, "project-get: WARNING: %s diverges from HEAD on '## %s' — resolving from HEAD; commit or restore the working tree to silence this\n", cfg, key)
			}
			src = tmp
		}
	}

	content, err := os.ReadFile(src)
	found := err == nil
	if !found {
		fmt.Fprintf(stderr, "project-get: %s does not exist\n", cfg)
		return 1
	}
	count := 0
	for _, line := range lines(bytes.TrimPrefix(content, []byte("\xef\xbb\xbf"))) {
		if line == "## "+key {
			count++
		}
	}
	switch count {
	case 0:
		fmt.Fprintf(stderr, "project-get: %s declares no '## %s' section\n", cfg, key)
		return 1
	case 1:
		body := projectSection(src, key)
		if enum {
			return pgEnum(body, key, args[3:], stdout, stderr)
		}
		if body != "" {
			fmt.Fprintln(stdout, body)
		}
		return 0
	}
	fmt.Fprintf(stderr, "project-get: %s declares %d '## %s' sections — a second declaration is ambiguous, so neither was read\n", cfg, count, key)
	return 2
}

// pgEnum prints the literal body's head matches and exits 0; a head matching
// none exits 3 with one stderr line quoting it. The head is the first
// non-blank line, whitespace-trimmed, surrounding backticks removed, trimmed
// again (skills/flow-contracts/project-configuration.md).
func pgEnum(body, key string, literals []string, stdout, stderr io.Writer) int {
	head, _, _ := strings.Cut(body, "\n")
	head = strings.TrimSpace(strings.Trim(strings.TrimSpace(head), "`"))
	for _, l := range literals {
		if head == l {
			fmt.Fprintln(stdout, l)
			return 0
		}
	}
	fmt.Fprintf(stderr, "project-get: '## %s' head '%s' matches none of: '%s'\n", key, head, strings.Join(literals, "' '"))
	return 3
}

// pgGit is envGit with GIT_DIR, GIT_WORK_TREE and GIT_INDEX_FILE removed
// from the environment, as the bash `unset` them.
func pgGit(env Env) func(args ...string) *exec.Cmd {
	var keep []string
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, "GIT_DIR=") && !strings.HasPrefix(kv, "GIT_WORK_TREE=") && !strings.HasPrefix(kv, "GIT_INDEX_FILE=") {
			keep = append(keep, kv)
		}
	}
	base := envGit(env)
	return func(args ...string) *exec.Cmd {
		cmd := base(args...)
		cmd.Env = keep
		return cmd
	}
}

// pgHead writes cmd's stdout to a new temporary file and returns its path.
func pgHead(cmd *exec.Cmd) (string, error) {
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	f, err := os.CreateTemp("", "project-get.")
	if err != nil {
		return "", err
	}
	_, err = f.Write(out)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		os.Remove(f.Name())
		return "", err
	}
	return f.Name(), nil
}

// isRegular is `[ -f <p> ]`.
func isRegular(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.Mode().IsRegular()
}
