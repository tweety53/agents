package guard

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// checkSpecReach is scripts/check-spec-reach.sh's guard. The shim's header
// states the contract — which checkout is read, how Playwright decides
// reachability, what is enumerated, and the three exit codes; this file is
// the mechanism.
func init() {
	Registry["check-spec-reach"] = checkSpecReach
}

// srScripts prints every package.json script whose command contains
// `playwright test`, in the file's own key order.
const srScripts = `
  const s = JSON.parse(require("fs").readFileSync(process.argv[1], "utf8")).scripts || {};
  for (const [k, v] of Object.entries(s)) if (String(v).includes("playwright test")) console.log(k);
`

// srFiles prints the absolute file of every suite in a Playwright JSON
// listing read from stdin.
const srFiles = `
    let s = ""; process.stdin.on("data", d => s += d).on("end", () => {
      const j = JSON.parse(s.slice(s.indexOf("{")));
      const root = j.config.rootDir;
      const walk = (x) => { if (x.file) console.log(require("path").resolve(root, x.file)); (x.suites || []).forEach(walk); };
      (j.suites || []).forEach(walk);
    });
  `

func checkSpecReach(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: check-spec-reach.sh <project-root>")
		return 2
	}
	root := args[0]
	if !isDir(smcAbs(env, root)) {
		fmt.Fprintf(stderr, "check-spec-reach: %s is not a directory\n", root)
		return 2
	}
	rootAbs := filepath.Clean(pcAbs(env, root))
	cfg := rootAbs + "/.flow/project.md"
	text, err := os.ReadFile(cfg)
	if !isFile(cfg) || err != nil {
		fmt.Fprintln(stdout, "Spec reach: not configured")
		return 0
	}
	switch n := vvHeadingCount(string(text)); {
	case n > 1:
		fmt.Fprintf(stderr, "check-spec-reach: %s declares %d '## visual verification' sections — a second declaration is ambiguous, so neither was read\n", cfg, n)
		return 2
	case n == 0:
		fmt.Fprintln(stdout, "Spec reach: not configured")
		return 0
	}
	checkout := ""
	for _, l := range vvSectionLines(string(text)) {
		if cells := vtSplitCells(l.Text); len(cells) == 2 && vtFoldCell(cells[0]) == "regression checkout" {
			checkout = vtTrimCell(cells[1])
			break
		}
	}
	if checkout == "" {
		fmt.Fprintln(stdout, "Spec reach: not configured")
		return 0
	}
	if !isDir(smcAbs(env, checkout)) {
		fmt.Fprint(stderr, sanitizeDisplay(fmt.Sprintf("check-spec-reach: `regression checkout` names `%s`, which is not an existing directory\n", checkout)))
		return 2
	}
	checkout = filepath.Clean(pcAbs(env, checkout))
	// The declared checkout is a main checkout: while a worktree of it sits on
	// the project root's own branch, that worktree holds the change's specs
	// (KAN-924), as resolve-visual-screenshots.sh resolves it.
	git := envGit(env)
	if branch, _ := capture(git("-C", rootAbs, "branch", "--show-current")); branch != "" {
		list, _ := capture(git("-C", checkout, "worktree", "list", "--porcelain"))
		w := ""
		for _, l := range strings.Split(list, "\n") {
			if p, ok := strings.CutPrefix(l, "worktree "); ok {
				w = p
			} else if l == "branch refs/heads/"+branch {
				checkout = w
				break
			}
		}
	}
	if !isFile(checkout + "/package.json") {
		fmt.Fprint(stderr, sanitizeDisplay(fmt.Sprintf("check-spec-reach: `%s` holds no package.json — nothing declares what reaches its specs\n", checkout)))
		return 2
	}
	tools := map[string]string{}
	for _, tool := range []string{"node", "npm"} {
		p, ok := lookPath(env, tool)
		if !ok {
			fmt.Fprintf(stderr, "check-spec-reach: %s is not on PATH — cannot ask Playwright what the scripts reach\n", tool)
			return 2
		}
		tools[tool] = p
	}
	run := func(dir string, stdin string, name string, args ...string) (string, int) {
		cmd := exec.Command(tools[name], args...)
		cmd.Dir, cmd.Env = dir, srEnviron(env)
		cmd.Stdin = strings.NewReader(stdin)
		out, err := cmd.CombinedOutput()
		return strings.TrimRight(string(out), "\n"), exitCode(err)
	}

	scripts, rc := run(env.Dir, "", "node", "-e", srScripts, checkout+"/package.json")
	if rc != 0 {
		fmt.Fprint(stderr, sanitizeDisplay(fmt.Sprintf("check-spec-reach: could not read scripts from `%s/package.json`: %s\n", checkout, scripts)))
		return 2
	}
	if scripts == "" {
		fmt.Fprint(stderr, sanitizeDisplay(fmt.Sprintf("check-spec-reach: no script in `%s/package.json` contains `playwright test` — nothing can reach its specs\n", checkout)))
		return 2
	}
	reached := map[string]bool{}
	for _, script := range strings.Split(scripts, "\n") {
		if script == "" {
			continue
		}
		listing, rc := run(checkout, "", "npm", "run", "-s", script, "--", "--list", "--reporter=json")
		if rc != 0 {
			fmt.Fprint(stderr, sanitizeDisplay(fmt.Sprintf("check-spec-reach: `npm run -s %s -- --list --reporter=json` exited %d in `%s`:\n%s\n", script, rc, checkout, listing)))
			return 2
		}
		files, rc := run(env.Dir, listing, "node", "-e", srFiles)
		if rc != 0 {
			fmt.Fprint(stderr, sanitizeDisplay(fmt.Sprintf("check-spec-reach: `npm run -s %s -- --list --reporter=json` printed no parsable JSON in `%s`: %s\n", script, checkout, files)))
			return 2
		}
		for _, f := range strings.Split(files, "\n") {
			reached[f] = true
		}
	}

	var specs []string
	_ = filepath.WalkDir(checkout, func(p string, d fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return nil
		case d.Name() == "node_modules" || d.Name() == ".git" || d.Name() == ".worktrees":
			if d.IsDir() {
				return fs.SkipDir
			}
		case d.Type().IsRegular() && strings.HasSuffix(d.Name(), ".spec.ts"):
			specs = append(specs, p)
		}
		return nil
	})
	sort.Strings(specs)
	orphans := 0
	for _, spec := range specs {
		if !reached[spec] {
			fmt.Fprint(stdout, sanitizeDisplay(strings.TrimPrefix(spec, checkout+"/")+": reached by no package.json script\n"))
			orphans++
		}
	}
	if orphans != 0 {
		fmt.Fprintf(stderr, "Spec reach: %d of %d spec(s) reached by no package.json script\n", orphans, len(specs))
		return 1
	}
	fmt.Fprintf(stdout, "Spec reach: %d spec(s), all reached\n", len(specs))
	return 0
}

// srEnviron is the process environment with PATH taken from env, so a test's
// fake binaries are found, and the PLAYWRIGHT_JSON_OUTPUT_* variables the
// bash unset for every listing run removed.
func srEnviron(env Env) []string {
	out := []string{"PATH=" + env.Getenv("PATH")}
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, "PATH=") && !strings.HasPrefix(kv, "PLAYWRIGHT_JSON_OUTPUT_NAME=") &&
			!strings.HasPrefix(kv, "PLAYWRIGHT_JSON_OUTPUT_DIR=") && !strings.HasPrefix(kv, "PLAYWRIGHT_JSON_OUTPUT_FILE=") {
			out = append(out, kv)
		}
	}
	return out
}
