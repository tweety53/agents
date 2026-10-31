package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tweety53/agents/stats/internal/client"
	"github.com/tweety53/agents/stats/internal/fallback"
)

const changeUsage = `usage: flow change save [-addr url] [-timeout dur] [-C dir] <name>
       flow change file [-addr url] [-timeout dur] [-C dir] <name> <path>

save keeps the two files of a change that outlive its worktree in flowd
(KAN-1073): <wt>/spectre/changes/<name>/tasks.md and
<wt>/.superpowers/sdd/<name>/self-review-context.md, <wt> the git worktree
-C names (default: cwd). Run 2's cleanup runs it once, before the
worktrees are removed. A missing file is refused, exit 1, naming each one,
before anything is written; a second save replaces both bodies.

file prints one saved file's body on stdout; exit 1 when flowd holds none.

Both exit 2 when flowd cannot be reached, and on a usage mistake.
`

// changeSavedFiles are the files save keeps, each by its saved path and its
// location under the worktree.
func changeSavedFiles(name string) [][2]string {
	return [][2]string{
		{"tasks.md", filepath.Join("spectre", "changes", name, "tasks.md")},
		{"self-review-context.md", filepath.Join(".superpowers", "sdd", name, "self-review-context.md")},
	}
}

// runChange implements `flow change save|file`.
func runChange(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, changeUsage)
		return 2
	}
	var want int
	switch args[0] {
	case "save":
		want = 1
	case "file":
		want = 2
	default:
		fmt.Fprintf(stderr, "flow: unknown change command %q\n", args[0])
		fmt.Fprint(stderr, changeUsage)
		return 2
	}

	fset := flag.NewFlagSet("flow change "+args[0], flag.ContinueOnError)
	fset.SetOutput(stderr)
	var f recordIdentityFlags
	registerRecordConnFlags(fset, &f)
	if err := fset.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprint(stderr, changeUsage)
		return 2
	}
	noteAddrUsage(fset, stderr, f.addr)
	if fset.NArg() != want {
		fmt.Fprintf(stderr, "flow: change %s takes %d argument(s), got %d\n", args[0], want, fset.NArg())
		fmt.Fprint(stderr, changeUsage)
		return 2
	}
	name := fset.Arg(0)
	if f.dir == "" {
		wd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(stderr, "flow: resolve working directory: %v\n", err)
			return 1
		}
		f.dir = wd
	}
	projectKey, _, err := fallback.ProjectKey(f.dir)
	if err != nil {
		fmt.Fprintf(stderr, "flow: resolve project key: %v\n", err)
		return 1
	}

	if args[0] == "file" {
		body, err := callRecord(ctx, f.addr, f.timeout, func(ctx context.Context, cl *client.Client) (string, error) {
			return cl.GetChangeFile(ctx, projectKey, name, fset.Arg(1))
		})
		if code := changeCallExit(err, "file", stderr); code != 0 {
			return code
		}
		_, _ = io.WriteString(stdout, body)
		return 0
	}

	top, err := exec.Command("git", "-C", f.dir, "rev-parse", "--show-toplevel").Output()
	if err != nil {
		fmt.Fprintf(stderr, "flow: change save: %s is not inside a git worktree: %v\n", f.dir, err)
		return 1
	}
	wt := strings.TrimSpace(string(top))
	files := changeSavedFiles(name)
	bodies := make([]string, len(files))
	missing := false
	for i, file := range files {
		b, err := os.ReadFile(filepath.Join(wt, file[1]))
		if err != nil {
			fmt.Fprintf(stderr, "flow: change save: cannot read %s: %v\n", filepath.Join(wt, file[1]), err)
			missing = true
			continue
		}
		bodies[i] = string(b)
	}
	if missing {
		return 1
	}
	for i, file := range files {
		_, err := callRecord(ctx, f.addr, f.timeout, func(ctx context.Context, cl *client.Client) (struct{}, error) {
			return struct{}{}, cl.PutChangeFile(ctx, projectKey, name, file[0], bodies[i])
		})
		if code := changeCallExit(err, "save", stderr); code != 0 {
			return code
		}
	}
	return 0
}

// changeCallExit maps one flowd call's error onto the command's exit code:
// a store answer refusing or not holding the record is 1, an unreachable
// store 2.
func changeCallExit(err error, verb string, stderr io.Writer) int {
	switch {
	case err == nil:
		return 0
	case errors.Is(err, client.ErrNotFound), errors.Is(err, client.ErrRecordRejected):
		fmt.Fprintf(stderr, "flow: change %s: %v\n", verb, err)
		return 1
	default:
		fmt.Fprintf(stderr, "flow: change %s: %v\n", verb, err)
		return 2
	}
}
