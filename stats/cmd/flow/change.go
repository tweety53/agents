package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/tweety53/agents/stats/internal/client"
	"github.com/tweety53/agents/stats/internal/fallback"
)

const changeUsage = `usage: flow change save [-addr url] [-timeout dur] [-C dir] <name>
       flow change file [-addr url] [-timeout dur] [-C dir] <name> <path>

save keeps the files of a change that outlive its worktree in flowd
(KAN-1073): its plan, <wt>/spectre/changes/<name>/tasks.md or else
/flow-fast's <wt>/.superpowers/sdd/<name>/tasks.md, saved as tasks.md, and
<wt>/.superpowers/sdd/<name>/self-review-context.md, <wt> the git worktree
-C names (default: cwd). A change with no plan at all (/flow-interactive)
saves the bundle alone. It also keeps every *-report-*.md and
visual-manifest-*.json under .superpowers/sdd/ of that worktree and of every
other worktree the change's record names, each as
reports.<repository>.<file>; one that cannot be read or saved is skipped. Every /flow* cleanup runs it once, before the
worktrees are removed. A missing bundle is refused, exit 1, naming it,
before anything is written; a second save replaces the bodies.

file prints one saved file's body on stdout; exit 1 when flowd holds none.

Both exit 2 when flowd cannot be reached, and on a usage mistake.
`

// changeSavedFiles are the files a change may hold that outlive its
// worktree, each by its saved path and its location under the worktree:
// the plan in either of its two places, then the self-review bundle.
// remove-change-worktrees' save gate mirrors this list.
func changeSavedFiles(name string) [][2]string {
	return [][2]string{
		{"tasks.md", filepath.Join("spectre", "changes", name, "tasks.md")},
		{"tasks.md", filepath.Join(".superpowers", "sdd", name, "tasks.md")},
		{"self-review-context.md", filepath.Join(".superpowers", "sdd", name, "self-review-context.md")},
	}
}

// runChange implements `flow change save|file`.
func runChange(ctx context.Context, args []string, getenv func(string) string, stdout, stderr io.Writer) int {
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
	registerRecordConnFlags(fset, &f, getenv)
	if err := fset.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprint(stderr, changeUsage)
		return 2
	}
	noteAddrUsage(fset, stderr, f.addr, getenv)
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
	// The first plan found is saved, and none is no refusal; the bundle is
	// required.
	var files [][2]string
	var bodies []string
	for _, file := range changeSavedFiles(name) {
		if file[0] == "tasks.md" && len(files) > 0 {
			continue
		}
		b, err := os.ReadFile(filepath.Join(wt, file[1]))
		if err != nil && file[0] == "tasks.md" {
			continue
		}
		if err != nil {
			fmt.Fprintf(stderr, "flow: change save: cannot read %s: %v\n", filepath.Join(wt, file[1]), err)
			return 1
		}
		files = append(files, file)
		bodies = append(bodies, string(b))
	}
	reports, reportBodies := changeReports(ctx, f, projectKey, name, wt)
	files, bodies = append(files, reports...), append(bodies, reportBodies...)
	for i, file := range files {
		_, err := callRecord(ctx, f.addr, f.timeout, func(ctx context.Context, cl *client.Client) (struct{}, error) {
			return struct{}{}, cl.PutChangeFile(ctx, projectKey, name, file[0], bodies[i])
		})
		if code := changeCallExit(err, "save", stderr); code != 0 && !strings.HasPrefix(file[0], "reports.") {
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

// changeReports reads every dispatch report and measure manifest the run
// left under .superpowers/sdd/ of wt and of every other worktree the
// change's record names, each by its saved path reports.<repository>.<file>,
// <repository> the worktree's main checkout's basename. Reports are evidence,
// never a gate: a record that cannot be read leaves wt alone, a worktree or
// file that cannot be read is skipped, and a report flowd refuses is named
// on stderr and leaves the save's exit alone.
func changeReports(ctx context.Context, f recordIdentityFlags, projectKey, name, wt string) (files [][2]string, bodies []string) {
	dirs := []string{wt}
	var record struct {
		Worktrees map[string]json.RawMessage `json:"worktrees"`
	}
	if body, err := getChange(ctx, f.addr, f.timeout, projectKey, name); err == nil && json.Unmarshal(body, &record) == nil {
		others := slices.Sorted(maps.Keys(record.Worktrees))
		dirs = append(dirs, slices.DeleteFunc(others, func(d string) bool { return d == wt })...)
	}
	for _, dir := range dirs {
		_, main, err := fallback.ProjectKey(dir)
		if err != nil {
			continue
		}
		for _, pattern := range []string{"*-report-*.md", "visual-manifest-*.json"} {
			matches, _ := filepath.Glob(filepath.Join(dir, ".superpowers", "sdd", pattern))
			for _, m := range matches {
				b, err := os.ReadFile(m)
				if err != nil {
					continue
				}
				saved := "reports." + filepath.Base(main) + "." + filepath.Base(m)
				files, bodies = append(files, [2]string{saved, m}), append(bodies, string(b))
			}
		}
	}
	return files, bodies
}
