package guard

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Every case of scripts/test-check-model-keys.sh at d71a2327, one subtest
// per ok: label. The bash harness ran the guard, or a sandbox copy of it
// whose settings.go it wrote, under a PATH it chose; here the guard runs
// in-process with FLOW_GUARD_REPO_ROOT naming this checkout or a sandbox
// tree, and a stub `flow` on the injected PATH is exec'd for real, as the
// guard execs the CLI in production. Each harness case runs the guard once,
// shared by the subtests carrying its labels.

// mkRoot is new_root plus, when body is non-nil, its .flow/project.md.
func mkRoot(t *testing.T, body *string) string {
	t.Helper()
	root := t.TempDir()
	if body != nil {
		writeFile(t, root+"/.flow/project.md", *body)
	}
	return root
}

// mkStub is stub_flow: a directory holding an executable `flow` running
// script under bash.
func mkStub(t *testing.T, script string) string {
	t.Helper()
	dir := t.TempDir()
	writeExec(t, dir+"/flow", "#!/usr/bin/env bash\n"+script+"\n")
	return dir
}

// mkEnv is the guard's environment: the shim's FLOW_GUARD_REPO_ROOT, and
// PATH when path is non-empty (run_guard_path); every other name the
// process's, as run_guard inherited it.
func mkEnv(root, path string) Env {
	vars := map[string]string{"FLOW_GUARD_REPO_ROOT": root}
	if path != "" {
		vars["PATH"] = path
	}
	return crEnv(root, vars)
}

func TestCheckModelKeys(t *testing.T) {
	t.Parallel()
	const (
		ask         = `[ "$1" = settings ] && [ "$2" = models ]`
		strict      = ask + ` || { echo "stub: unexpected args: $*" >&2; exit 3; }` + "\n"
		storeDown   = `echo "stub: store down" >&2; exit 1`
		memberless  = "package store\n\nvar Other = map[string]bool{\n\t\"x\": true,\n}\n"
		sysPath     = ":/usr/bin:/bin"
		settingsRel = "/stats/internal/store/settings.go"
	)
	key := func(v string) *string { s := "## self review model\n\n`" + v + "`\n"; return &s }
	// sandbox is new_sandbox: a checkout root whose settings.go is settings,
	// or absent when settings is nil.
	sandbox := func(t *testing.T, settings *string) string {
		sbx := t.TempDir()
		if settings != nil {
			writeFile(t, sbx+settingsRel, *settings)
		}
		return sbx
	}
	once := func(label string, n int) smcCheck {
		return smcCheck{label, func(_ int, out, _ string) bool { return strings.Count(out, "stub: store down") == n }}
	}
	cases := []struct {
		run    func(t *testing.T) (Env, []string)
		checks []smcCheck
	}{
		{func(t *testing.T) (Env, []string) { return mkEnv(smcRepoRoot, ""), []string{mkRoot(t, key("fable"))} },
			[]smcCheck{smcRC("valid self review model key passes", 0)}},
		{func(t *testing.T) (Env, []string) {
			return mkEnv(smcRepoRoot, ""), []string{mkRoot(t, key("bogus-model"))}
		},
			[]smcCheck{smcRC("invalid model value fails", 1),
				smcHas("invalid value is named", "bogus-model", "is not a ValidModels member")}},
		{func(t *testing.T) (Env, []string) { return mkEnv(smcRepoRoot, ""), []string{mkRoot(t, nil)} },
			[]smcCheck{smcRC("missing .flow/project.md passes", 0)}},
		{func(t *testing.T) (Env, []string) {
			return mkEnv(smcRepoRoot, ""), []string{mkRoot(t, nil), mkRoot(t, key("fable"))}
		}, []smcCheck{smcRC("mixed roots exit 0", 0),
			smcHas("a root with no .flow/project.md still counts toward the summary", "2 project(s) checked")}},
		{func(t *testing.T) (Env, []string) {
			return mkEnv(smcRepoRoot, ""), []string{t.TempDir() + "/check-model-keys-does-not-exist"}
		}, []smcCheck{smcRC("a nonexistent project root exits 2", 2)}},
		{func(t *testing.T) (Env, []string) {
			stub := mkStub(t, strict+`printf 'alpha\nbravo\n'`)
			return mkEnv(sandbox(t, nil), stub+sysPath), []string{mkRoot(t, key("bravo"))}
		}, []smcCheck{smcRC("with no source, the CLI's set governs", 0)}},
		{func(t *testing.T) (Env, []string) {
			stub := mkStub(t, strict+`printf 'sonnet\ngamma\n'`)
			return mkEnv(smcRepoRoot, stub+sysPath), []string{mkRoot(t, key("gamma"))}
		}, []smcCheck{smcRC("on drift a CLI-only name fails despite the CLI", 1),
			smcHas("the CLI-only value is named", "gamma", "is not a ValidModels member")}},
		{func(t *testing.T) (Env, []string) {
			return mkEnv(smcRepoRoot, mkStub(t, storeDown)+sysPath), []string{mkRoot(t, key("fable"))}
		}, []smcCheck{smcRC("a failing CLI falls back to the source parse", 0),
			smcHas("the fallback announcement carries the CLI's own diagnostic",
				"no answer from flow settings models", "stub: store down")}},
		{func(t *testing.T) (Env, []string) {
			return mkEnv(smcRepoRoot, "/usr/bin:/bin"), []string{mkRoot(t, key("fable"))}
		}, []smcCheck{smcRC("no flow on PATH falls back to the source parse", 0)}},
		{func(t *testing.T) (Env, []string) {
			return mkEnv(smcRepoRoot, mkStub(t, "exit 0")+sysPath), []string{mkRoot(t, key("fable"))}
		}, []smcCheck{smcRC("an empty CLI answer falls back to the source parse", 0)}},
		{func(t *testing.T) (Env, []string) {
			sbx := sandbox(t, nil)
			return mkEnv(sbx, "/usr/bin:/bin"), []string{sbx}
		}, []smcCheck{smcRC("no CLI and no settings.go exits 2", 2),
			smcHas("the missing-source refusal is named", "cannot read", "settings.go")}},
		{func(t *testing.T) (Env, []string) {
			s := memberless
			sbx := sandbox(t, &s)
			return mkEnv(sbx, "/usr/bin:/bin"), []string{sbx}
		}, []smcCheck{smcRC("a memberless settings.go exits 2", 2),
			smcHas("the empty-parse refusal is named", "no ValidModels members found")}},
		{func(t *testing.T) (Env, []string) {
			stub := mkStub(t, ask+` && printf 'sonnet\n'`+"\nexit 0")
			return mkEnv(smcRepoRoot, stub+sysPath), []string{mkRoot(t, key("sonnet"))}
		}, []smcCheck{smcRC("a subset CLI answer still validates its own members", 0),
			smcHas("the stale-install drift is announced", "disagrees with", "settings.go")}},
		{func(t *testing.T) (Env, []string) {
			s := memberless
			sbx := sandbox(t, &s)
			stub := mkStub(t, ask+` && printf 'sonnet\n'`+"\nexit 0")
			return mkEnv(sbx, stub+sysPath), []string{sbx}
		}, []smcCheck{smcRC("a shape-changed source keeps the CLI verdict", 0),
			smcHas("the disabled cross-check is announced", "cross-check is off")}},
		{func(t *testing.T) (Env, []string) {
			stub := mkStub(t, ask+` && printf 'fable\nhaiku\nopus\n'`+"\nexit 0")
			return mkEnv(smcRepoRoot, stub+sysPath), []string{mkRoot(t, key("sonnet"))}
		}, []smcCheck{smcRC("a source member the stale CLI lacks still validates", 0),
			smcHas("the stale-install drift is announced beside the flip", "disagrees with", "settings.go")}},
		{func(t *testing.T) (Env, []string) {
			stub := mkStub(t, ask+` && printf 'sonnet\nsonnet\nfable\nhaiku\nopus\n'`+"\nexit 0")
			return mkEnv(smcRepoRoot, stub+sysPath), []string{mkRoot(t, key("sonnet"))}
		}, []smcCheck{smcRC("a duplicated CLI member is not drift", 0),
			smcLacks("no drift announced for a duplicated member", "disagrees with")}},
		{func(t *testing.T) (Env, []string) {
			stub := mkStub(t, "echo \"stub: deprecation warning\" >&2\n"+ask+` && printf 'fable\n'`+"\nexit 0")
			return mkEnv(smcRepoRoot, stub+sysPath), []string{mkRoot(t, key("fable"))}
		}, []smcCheck{smcRC("an answering CLI with stderr still validates", 0),
			smcHas("the answering CLI's diagnostic is relayed", "flow settings models reported:", "stub: deprecation warning")}},
		{func(t *testing.T) (Env, []string) {
			return mkEnv(smcRepoRoot, mkStub(t, storeDown)+sysPath), []string{mkRoot(t, key("fable"))}
		}, []smcCheck{smcRC("a failed ask still falls back cleanly", 0),
			once("the failed ask's diagnostic is emitted exactly once", 1)}},
	}
	for i, c := range cases {
		// The fixtures belong to the parent test, so they outlive the
		// parallel subtest reading this case's one run.
		env, args := c.run(t)
		// One parallel subtest runs the case, its checks nested beneath it:
		// a parallel subtest per check held a -parallel slot apiece while
		// the one running the case worked and the rest waited on it.
		t.Run(fmt.Sprint("run ", i), func(t *testing.T) {
			t.Parallel()
			var out bytes.Buffer
			rc := checkModelKeys(args, env, &out, &out)
			for _, chk := range c.checks {
				t.Run(chk.label, func(t *testing.T) {
					if !chk.ok(rc, out.String(), "") {
						t.Fatalf("rc=%d out=%s", rc, out.String())
					}
				})
			}
		})
	}

	// Beyond the harness: the mechanisms the bash got implicitly from awk,
	// read, tr, sed, sort, comm and printf %q are explicit code here, so each
	// is pinned by running the bash at d71a2327 and the port over the same
	// tree, stub CLI and project roots, and comparing stdout, stderr and exit
	// byte for byte.
	tree := t.TempDir()
	for _, rel := range []string{"scripts/check-model-keys.sh", "scripts/lib/project-section.sh", "scripts/lib/strip-bom.sh"} {
		src, err := exec.Command(fixtureGit, "-C", smcRepoRoot, "show", "d71a2327:"+rel).Output()
		if err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(tree, rel), string(src))
	}
	// Two quoted strings on a line (the first wins), an empty "" before a
	// real one, a CRLF line, a NUL cutting a line short, a nested opener
	// line skipped, a line with no quotes, and `}` ending the block before a
	// member after it, and a second block after that.
	settings := "package store\n\nvar ValidModels = map[string]bool{\r\n" +
		"\t\"alpha\": true, \"ignored\": true,\n" +
		"\t\"\" \"c d\": true,\n" +
		"\t\"Gamma\": true,\r\n" +
		"\t\"delta\"\x00\"cut\": true,\n" +
		"\t\x00\"hidden\": true,\n" +
		"var ValidModels = map[string]bool{\"skipped\": true,\n" +
		"\t// no quotes here\n" +
		"\t\"_x\": true,\n" +
		"}\r\n\t\"after\": true,\n" +
		"var ValidModels = map[string]bool{\n\t\"second\": true,\n}\n"
	writeFile(t, tree+settingsRel, settings)
	bodies := []string{
		"`alpha`",          // valid
		"  `alpha`\t\r",    // edge whitespace trimmed, CR included
		"`alpha`\u00a0",    // NBSP: a space only under a UTF-8 locale
		"\u00a0\n`alpha`",  // an NBSP-only line: blank only under UTF-8
		"`alpha`\x00junk",  // NUL ends the line
		"alpha",            // no backticks
		"`c d`",            // an inner space
		"``",               // empty code span: no body at all
		"`",                // a lone backtick stays
		"`alpha`\n``",      // a trailing span that empties is cut
		"``\n`alpha`",      // a leading one is not
		"`a`\n\n`b`",       // multi-line
		"`bogus \"$x\" é`", // printf %q
		"`tab\there`",      // printf %q ANSI-C
		"`Gamma`",          // case-sensitive match
		"`gamma`",          // ...
		"`_x`",             // last member
		"`after`",          // past the block's end
		"`second`",         // in a second block awk never reached
		"`hidden`",         // behind a NUL
		"`delta`",          // cut at the NUL
	}
	var roots []string
	for _, b := range bodies {
		body := "# P\n\n## self review model\n\n" + b + "\n\n## next\nx\n"
		roots = append(roots, mkRoot(t, &body))
	}
	crlf := "## self review model\r\n`bogus`\r\n"
	roots = append(roots, mkRoot(t, nil), mkRoot(t, &crlf))
	dangling := t.TempDir()
	mkdir(t, dangling+"/.flow")
	if err := os.Symlink(dangling+"/missing", dangling+"/.flow/project.md"); err != nil {
		t.Fatal(err)
	}
	roots = append(roots, dangling)

	// The CLI's answer: mixed-case names for sort/comm collation, a
	// duplicate, a CR-ended name, a NUL inside a name, an empty line, and an
	// unterminated last line read never sees; its stderr carries interior
	// newlines and trailing whitespace, NBSP included, for tr and sed.
	drift := mkStub(t, `printf 'Zeta\nalpha\nBeta\nbeta\nBeta\nx\r\nal\0pha2\n\n_x\nGamma\ntail'
printf 'warn one\n\tsecond  \xc2\xa0 \n' >&2`)
	agree := mkStub(t, `printf 'alpha\nc d\nGamma\ndelta\n_x\n'`)
	failing := mkStub(t, `printf 'down\t\n\n' >&2; exit 1`)
	unterminated := mkStub(t, `printf 'alpha'`)
	emptyErr := mkStub(t, `printf 'alpha\n'; printf '\n' >&2`)
	dirSettings := t.TempDir()
	mkdir(t, dirSettings+settingsRel)
	for _, rel := range []string{"scripts/check-model-keys.sh", "scripts/lib/project-section.sh", "scripts/lib/strip-bom.sh"} {
		b, err := os.ReadFile(filepath.Join(tree, rel))
		if err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(dirSettings, rel), string(b))
	}
	notDir := t.TempDir() + "/file"
	writeFile(t, notDir, "x")
	pfDir := t.TempDir()
	mkdir(t, pfDir+"/.flow/project.md")
	unreadable := mkRoot(t, key("alpha"))
	if err := os.Chmod(unreadable+"/.flow/project.md", 0); err != nil {
		t.Fatal(err)
	}
	noRead := t.TempDir()
	if err := os.Chmod(noRead, 0o333); err != nil {
		t.Fatal(err)
	}
	for _, sc := range []struct {
		name, tree, path, locale string
		args                     []string
	}{
		{"drift under en_US.UTF-8", tree, drift, "en_US.UTF-8", roots},
		{"drift under C", tree, drift, "C", roots},
		{"an agreeing CLI", tree, agree, "en_US.UTF-8", roots},
		{"a failing CLI's stderr in the fallback line", tree, failing, "en_US.UTF-8", roots[:3]},
		{"an unterminated answer is empty", tree, unterminated, "en_US.UTF-8", roots[:3]},
		{"a blank-line stderr on an answering CLI", tree, emptyErr, "en_US.UTF-8", roots[:3]},
		{"no flow on PATH", tree, "", "en_US.UTF-8", roots[:3]},
		{"settings.go is a directory", dirSettings, "", "en_US.UTF-8", roots[:1]},
		{"no arguments checks the checkout", tree, agree, "en_US.UTF-8", nil},
		{"a relative root", tree, agree, "en_US.UTF-8", []string{"scripts", "."}},
		{"an empty root", tree, agree, "en_US.UTF-8", []string{""}},
		{"a violation, then a root that is a file", tree, agree, "en_US.UTF-8", []string{roots[1], roots[11], notDir, roots[0]}},
		{"project.md is a directory", tree, agree, "en_US.UTF-8", []string{roots[0], pfDir}},
		{"a root that is not readable", tree, agree, "en_US.UTF-8", []string{noRead}},
		{"project.md is unreadable", tree, agree, "en_US.UTF-8", []string{unreadable}},
	} {
		t.Run("port: "+sc.name+" matches the bash at d71a2327", func(t *testing.T) {
			t.Parallel()
			path := "/usr/bin:/bin"
			if sc.path != "" {
				path = sc.path + sysPath
			}
			env := crEnv(sc.tree, map[string]string{"FLOW_GUARD_REPO_ROOT": sc.tree, "PATH": path, "LC_ALL": sc.locale})
			var gout, gerr bytes.Buffer
			grc := checkModelKeys(sc.args, env, &gout, &gerr)

			cmd := exec.Command("bash", append([]string{filepath.Join(sc.tree, "scripts/check-model-keys.sh")}, sc.args...)...)
			cmd.Dir = sc.tree
			cmd.Env = append(os.Environ(), "PATH="+path, "LC_ALL="+sc.locale)
			var bout, berr bytes.Buffer
			cmd.Stdout, cmd.Stderr = &bout, &berr
			brc := 0
			if err := cmd.Run(); err != nil {
				ee, ok := err.(*exec.ExitError)
				if !ok {
					t.Fatal(err)
				}
				brc = ee.ExitCode()
			}
			if grc != brc || gout.String() != bout.String() || gerr.String() != berr.String() {
				t.Fatalf("port rc=%d stdout=\n%q\nstderr=\n%q\nbash rc=%d stdout=\n%q\nstderr=\n%q",
					grc, gout.String(), gerr.String(), brc, bout.String(), berr.String())
			}
		})
	}
}
