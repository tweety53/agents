// Command flow's prompttuning.go implements `flow prompt-tuning`: render
// writes a prompt-tuning case folder's index.html from its case.json and
// repro files (internal/prompttuning), so the page is never written by hand
// and check-prompt-tuning-cases can compare it byte for byte; record and
// import feed the stats store's prompt_changes, the Review dynamics view's
// markers (prompt-changes-table).
package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/tweety53/agents/stats/internal/client"
	"github.com/tweety53/agents/stats/internal/prompttuning"
)

const promptTuningUsage = `usage: flow prompt-tuning render <case-dir>
       flow prompt-tuning record <case-dir> -sha <sha> [-at <RFC3339>]
       flow prompt-tuning import [-C <repo>] [-since YYYY-MM-DD]

render writes <case-dir>/index.html from the folder's case.json, finding.md,
bad.diff, good.diff and task.md. Exit 0 silent, 1 on a load or render error,
2 on usage.

record posts one tuning row for an applied case: sha <sha>:<target>, area and
slot from the prompt's file (else the case's type), summary its defect class.
-at defaults to the commit time of <sha> in the case dir's repository.

import posts one edit row per (commit, mapped prompt file) of the repository's
git history, keyed <sha>:<target>, and prints "imported: <n>". Re-running it
upserts. Both exit 0 on success, 1 on a git or store failure, 2 on usage.
`

func runPromptTuning(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	switch {
	case len(args) == 2 && args[0] == "render":
		return runPromptTuningRender(args[1], stderr)
	case len(args) >= 2 && args[0] == "record":
		return runPromptTuningRecord(ctx, args[1], args[2:], stderr)
	case len(args) >= 1 && args[0] == "import":
		return runPromptTuningImport(ctx, args[1:], stdout, stderr)
	}
	fmt.Fprint(stderr, promptTuningUsage)
	return 2
}

func runPromptTuningRender(dir string, stderr io.Writer) int {
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

// parsePromptTuningFlags parses args into fset, printing the usage on a
// parse error; it returns 0 to continue, else the exit code.
func parsePromptTuningFlags(fset *flag.FlagSet, args []string, stderr io.Writer) (int, bool) {
	if err := fset.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0, false
		}
		fmt.Fprintf(stderr, "flow: %v\n", err)
		fmt.Fprint(stderr, promptTuningUsage)
		return 2, false
	}
	if fset.NArg() != 0 {
		fmt.Fprintln(stderr, "flow: prompt-tuning: unexpected arguments")
		fmt.Fprint(stderr, promptTuningUsage)
		return 2, false
	}
	return 0, true
}

func runPromptTuningRecord(ctx context.Context, dir string, args []string, stderr io.Writer) int {
	fset := flag.NewFlagSet("flow prompt-tuning record", flag.ContinueOnError)
	fset.SetOutput(stderr)
	var f storeConnFlags
	registerStoreConnFlags(fset, &f)
	sha := fset.String("sha", "", "the commit that applied the case (required)")
	atFlag := fset.String("at", "", "when it was applied, RFC 3339 (default: the commit's time)")
	if code, ok := parsePromptTuningFlags(fset, args, stderr); !ok {
		return code
	}
	if *sha == "" {
		fmt.Fprintln(stderr, "flow: prompt-tuning record needs -sha")
		fmt.Fprint(stderr, promptTuningUsage)
		return 2
	}
	c, _, err := prompttuning.Load(dir)
	if err != nil {
		fmt.Fprintf(stderr, "flow: prompt-tuning record: %v\n", err)
		return 1
	}
	at := *atFlag
	if at == "" {
		if at, err = gitOutput(dir, "show", "-s", "--format=%cI", *sha); err != nil {
			fmt.Fprintf(stderr, "flow: prompt-tuning record: %v\n", err)
			return 1
		}
	}
	when, err := time.Parse(time.RFC3339, at)
	if err != nil {
		fmt.Fprintf(stderr, "flow: prompt-tuning record: -at: %v\n", err)
		return 2
	}

	area, slot, ok := prompttuning.AreaFor(c.Target)
	if !ok {
		area, slot = "review", c.Source.Slot
		if c.Type == "implementer" {
			area = "impl"
		}
	}
	row := client.PromptChange{SHA: *sha + ":" + c.Target, At: when, Kind: "tuning", Area: area,
		Target: c.Target, Summary: c.DefectClass, CaseID: &c.ID}
	if slot != "" {
		row.Slot = &slot
	}
	return postPromptChanges(ctx, f, []client.PromptChange{row}, stderr, "record")
}

func runPromptTuningImport(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fset := flag.NewFlagSet("flow prompt-tuning import", flag.ContinueOnError)
	fset.SetOutput(stderr)
	var f storeConnFlags
	registerStoreConnFlags(fset, &f)
	repo := fset.String("C", ".", "the repository whose history to import")
	since := fset.String("since", "", "only commits on or after YYYY-MM-DD")
	if code, ok := parsePromptTuningFlags(fset, args, stderr); !ok {
		return code
	}

	// Reviewer prompts present in the repository but off the fixed map --
	// a slot added later -- are imported too (prompttuning.AreaFor).
	listed, err := gitOutput(*repo, "ls-files", "--", "skills/flow/*-reviewer-prompt.md")
	if err != nil {
		fmt.Fprintf(stderr, "flow: prompt-tuning import: %v\n", err)
		return 1
	}
	targets := prompttuning.MappedTargets()
	for _, t := range strings.Fields(listed) {
		if !slices.Contains(targets, t) {
			targets = append(targets, t)
		}
	}

	var rows []client.PromptChange
	for _, target := range targets {
		gitArgs := []string{"log", "--format=%H%x09%cI%x09%s"}
		if *since != "" {
			gitArgs = append(gitArgs, "--since="+*since)
		}
		out, err := gitOutput(*repo, append(gitArgs, "--", target)...)
		if err != nil {
			fmt.Fprintf(stderr, "flow: prompt-tuning import: %v\n", err)
			return 1
		}
		area, slot, _ := prompttuning.AreaFor(target)
		sc := bufio.NewScanner(strings.NewReader(out))
		for sc.Scan() {
			parts := strings.SplitN(sc.Text(), "\t", 3)
			if len(parts) != 3 {
				continue
			}
			when, err := time.Parse(time.RFC3339, parts[1])
			if err != nil {
				fmt.Fprintf(stderr, "flow: prompt-tuning import: %s: %v\n", parts[0], err)
				return 1
			}
			row := client.PromptChange{SHA: parts[0] + ":" + target, At: when, Kind: "edit", Area: area,
				Target: target, Summary: parts[2]}
			if slot != "" {
				s := slot
				row.Slot = &s
			}
			rows = append(rows, row)
		}
	}
	if code := postPromptChanges(ctx, f, rows, stderr, "import"); code != 0 {
		return code
	}
	fmt.Fprintf(stdout, "imported: %d\n", len(rows))
	return 0
}

func postPromptChanges(ctx context.Context, f storeConnFlags, rows []client.PromptChange, stderr io.Writer, verb string) int {
	cl := client.New(f.addr, &http.Client{Timeout: f.timeout})
	for _, row := range rows {
		reqCtx, cancel := context.WithTimeout(ctx, f.timeout)
		err := cl.RecordPromptChange(reqCtx, row)
		cancel()
		if err != nil {
			fmt.Fprintf(stderr, "flow: prompt-tuning %s: %s: %v\n", verb, row.SHA, err)
			return 1
		}
	}
	return 0
}

// gitOutput runs git in dir and returns its trimmed stdout.
func gitOutput(dir string, args ...string) (string, error) {
	var out, errb bytes.Buffer
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Stdout, cmd.Stderr = &out, &errb
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %v: %s", strings.Join(args, " "), err, strings.TrimSpace(errb.String()))
	}
	return strings.TrimSpace(out.String()), nil
}
