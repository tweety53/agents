package guard

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// The contract below moved here from the header of the deleted
// scripts/project-get.sh shim (KAN-1058).
//
// flow-guard project-get — print the body of one `## <key>` heading from a project's
// `.flow/project.md`, so every /flow phase file resolves a project-configured
// key through one script instead of loading a contract file to read one value.
//
// Usage: flow-guard project-get <project-root> <key>
//        flow-guard project-get <project-root> <key> --enum <literal>...
//
// <project-root> is the directory holding `.flow/project.md`, never the file
// itself. <key> is the heading text after `## `, quoted when it carries
// spaces (`flow-guard project-get <root> "default landing route"`).
//
// THREE EXIT CODES:
//   0  the key is declared exactly once; its body is on stdout.
//   1  `<project-root>/.flow/project.md` is absent, or declares no
//      `## <key>` heading — one line on stderr says which. Every
//      "optional key absent -> skip" case in the phase files is this exit.
//   2  usage (not exactly two arguments), `<project-root>` is not a
//      directory, or `## <key>` is declared more than once — the same
//      ambiguity refusal flow-guard check-visual-trigger makes: a second
//      declaration is ambiguous, so neither is read.
//
// --enum RESOLVES A SINGLE-LINE-LITERAL KEY (skills/flow-contracts/
// project-configuration.md). The body's head — its first non-blank line,
// whitespace-trimmed, surrounding backticks removed, trimmed again — matches
// a <literal> on equality, or when the literal leads it at a word boundary
// (the byte after it cannot extend it into a longer word); the rest of the
// head line is prose for the reader, never read:
//   0  the head matches; the literal alone is on stdout.
//   1  unchanged: the file or the key is absent.
//   2  unchanged; also `--enum` given no literal.
//   3  the key is declared but its head matches no literal — one stderr
//      line quotes the head; the caller reports it and resolves as absent.
//
// Derives no repository root by walking a fixed number of directory levels
// above its own location (check-guard-symlinks.sh rule 4) — the project
// root is always an argument, never inferred from where this script lives.
//
// RESOLUTION ORDER (KAN-520): when <project-root> sits inside a git work tree
// and HEAD carries .flow/project.md, the key is counted and read from HEAD's
// copy (`git show HEAD:<path>`), never from the working tree — a staged-but-
// uncommitted edit there must not silently override the landed value every
// /flow* stage reads (the kan-512 run-2 incident). The working tree is read
// only when HEAD lacks the file or the root is not a git repository — today's
// behavior, unchanged there. When the working tree's file exists and diverges
// from HEAD on the key being resolved, one stderr warning names the file, the
// key, and that HEAD won.
//
// The logic lives in stats/internal/guard/projectget.go, a byte-for-byte port
// of this script's bash (ae805186); flow-guard is built from this checkout,
// never taken from PATH — scripts/lib/flow-guard.sh derives it, and exits 2
// with the cause when it cannot.

// projectGet is scripts/project-get.sh: that script's header comment is the
// contract, and TestProjectGetParity pins this port against the bash it
// replaced (ae805186) on stdout, stderr and exit code.
func init() {
	Registry["project-get"] = projectGet
}

func projectGet(args []string, env Env, stdout, stderr io.Writer) int {
	// --enum <literal>... resolves a single-line-literal key: the body's
	// head is matched against the literals — on equality or on a leading,
	// word-bounded literal (pgEnum), the longest match winning.
	enum := len(args) > 3 && args[2] == "--enum"
	if len(args) != 2 && !enum {
		fmt.Fprintln(stderr, "usage: flow-guard project-get <project-root> <key>")
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

// pgEnum prints the literal the body's head matches and exits 0; a head matching
// none exits 3 with one stderr line quoting it. The head is the first
// non-blank line, whitespace-trimmed, surrounding backticks removed, trimmed
// again (skills/flow-contracts/project-configuration.md). The head matches a
// literal when it equals it or begins with it and the byte after the literal
// cannot extend it into a longer word — prose the operator writes after the
// literal on the head line is for the reader, never read (kan-716). When
// several literals match, the longest wins, so a literal that word-boundedly
// opens a longer one never shadows it.
func pgEnum(body, key string, literals []string, stdout, stderr io.Writer) int {
	head, _, _ := strings.Cut(body, "\n")
	head = strings.TrimSpace(strings.Trim(strings.TrimSpace(head), "`"))
	best, found := "", false
	for _, l := range literals {
		if matched := head == l || (strings.HasPrefix(head, l) && !pgWordByte(head[len(l)])); matched && (!found || len(l) > len(best)) {
			best, found = l, true
		}
	}
	if !found {
		fmt.Fprintf(stderr, "project-get: '## %s' head '%s' matches none of: '%s'\n", key, head, strings.Join(literals, "' '"))
		return 3
	}
	fmt.Fprintln(stdout, best)
	return 0
}

// pgWordByte reports whether b can extend a matched literal into a longer
// word: a letter, digit or underscore, the same set Go's \w admits.
func pgWordByte(b byte) bool {
	return b == '_' || ('a' <= b && b <= 'z') || ('A' <= b && b <= 'Z') || ('0' <= b && b <= '9')
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
