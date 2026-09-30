package guard

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"testing/iotest"
	"time"
)

// Every case of scripts/test-check-visual-trigger.sh at 3915fbc0, one subtest
// per ok: label, each run against a real .flow/project.md in its own
// t.TempDir() with the changed paths on Env.Stdin (printf '%s\n' "$@", so no
// paths is empty stdin). The harness asserted an exit code or a substring of
// the merged output; the "exit" subtests here pin the whole stdout and stderr
// the bash printed, so the substring subtests only restate the harness's own
// assertion. The rows after the harness's cases pin bash behaviours the
// harness never reached, each measured against the script at 3915fbc0 under
// LC_ALL=C.

const (
	vtPre    = "## visual verification\n\n| Setting | Value |\n|---------|-------|\n"
	vtPost   = "| `screenshots` | `stats/web/tests/visual` |\n\n| Command | Runs |\n|---------|------|\n| `verify` | `npm run test:visual` |\n| `capture` | `npx playwright test <spec>` |"
	vtNoRoot = "check-visual-trigger: "
	vtMatch  = "VISUAL-TRIGGER-MATCH: $ROOT — at least one changed path matched `ui paths`\n"
	vtNone   = "VISUAL-TRIGGER-NO-MATCH: $ROOT — no changed path matched `ui paths`\n"
	vtAbsent = vtNoRoot + "`ui paths` is absent or empty in $ROOT/.flow/project.md — cannot resolve what to match against\n" + vtCannot + "`ui paths` is absent or empty\n"
	vtNotCfg = "VISUAL-TRIGGER-NOT-CONFIGURED: $ROOT — no visual verification section\n"
	vtCannot = "VISUAL-TRIGGER-CANNOT-ANSWER: $ROOT — "
	vtUsage  = vtNoRoot + "usage: check-visual-trigger.sh <project root> (changed paths on stdin)\nVISUAL-TRIGGER-CANNOT-ANSWER: (no root) — usage\n"
)

// vtSection is the harness's section with cell as the `ui paths` value cell.
func vtSection(cell string) string { return vtPre + "| `ui paths` | " + cell + " |\n" + vtPost }

var vtMinimal = vtSection("`stats/web/src/**`")

// vtMatched is one MATCH line as the bash printed it through sanitize_display.
func vtMatched(path, glob string) string { return "MATCH: " + path + " — matched `" + glob + "`\n" }

type vtCase struct {
	cfg      *string                         // written with printf '%s\n'; nil leaves no .flow/project.md
	bom      bool                            // write_cfg_bom
	setup    func(t *testing.T, root string) // after cfg, for shapes cfg cannot write
	sub      string                          // appended to the temp root: the root argument and where cfg goes
	root     string                          // non-empty: the root argument instead of the temp root
	stdin    string                          // raw stdin; paths below are appended as lines
	paths    []string
	code     int
	out, err string // "$ROOT" is the root argument; "$SROOT" the same through sanitize_display
	contains []string
	omits    []string
	within   time.Duration // non-zero: an extra subtest, label, bounds the run
	label    string
}

func vtStr(s string) *string { return &s }

func vtRun(t *testing.T, c vtCase) (code int, stdout, stderr, root string) {
	t.Helper()
	fn := Registry["check-visual-trigger"]
	if fn == nil {
		t.Fatal("check-visual-trigger is not registered")
	}
	root = t.TempDir() + c.sub
	mkdir(t, root+"/.flow")
	if c.cfg != nil {
		body := *c.cfg + "\n"
		if c.bom {
			body = "\xef\xbb\xbf" + body
		}
		writeFile(t, root+"/.flow/project.md", body)
	}
	if c.setup != nil {
		c.setup(t, root)
	}
	if c.root != "" {
		root = c.root
	}
	in := c.stdin
	for _, p := range c.paths {
		in += p + "\n"
	}
	var out, errb bytes.Buffer
	code = fn([]string{root}, Env{Getenv: os.Getenv, Dir: t.TempDir(), Stdin: strings.NewReader(in)}, &out, &errb)
	return code, out.String(), errb.String(), root
}

// vtCheck runs c and pins its exit code, stdout and stderr whole.
func vtCheck(t *testing.T, c vtCase) {
	t.Helper()
	code, out, errOut, root := vtRun(t, c)
	subst := strings.NewReplacer("$SROOT", strings.TrimSuffix(sanitizeDisplay(root), "\n"), "$ROOT", root)
	wantOut, wantErr := subst.Replace(c.out), subst.Replace(c.err)
	if code != c.code || out != wantOut || errOut != wantErr {
		t.Fatalf("got exit %d\nstdout %q\nstderr %q\nwant exit %d\nstdout %q\nstderr %q", code, out, errOut, c.code, wantOut, wantErr)
	}
}

func TestCheckVisualTrigger(t *testing.T) {
	t.Parallel()
	longGlob := "src/very/long/path/that/exceeds/any/plausible/caller/variable/length/here/**"
	pad := strings.Repeat("`", 3000)
	bigPad := strings.Repeat("`", 60000)
	noSection := "# Project\n\n## run\n\necho hi\n"
	cases := map[string]vtCase{
		"case 1": {paths: []string{"stats/web/src/App.tsx"}, code: 2,
			err: vtNoRoot + "$ROOT has no .flow/project.md — not configured, so nothing was matched\n" + vtNotCfg},
		"case 2": {cfg: &noSection, paths: []string{"stats/web/src/App.tsx"}, code: 2,
			err: vtNoRoot + "$ROOT/.flow/project.md declares no '## visual verification' section — not configured, so nothing was matched\n" + vtNotCfg},
		"case 3": {cfg: vtStr(vtPre + vtPost), paths: []string{"stats/web/src/App.tsx"}, code: 2, err: vtAbsent},
		"case 4": {cfg: vtStr(vtPre + "| `ui paths` | |\n" + vtPost), paths: []string{"stats/web/src/App.tsx"}, code: 2, err: vtAbsent},
		"case 5": {cfg: &vtMinimal, paths: []string{"stats/web/src/App.tsx"}, code: 0,
			out: vtMatched("stats/web/src/App.tsx", "stats/web/src/**") + vtMatch, contains: []string{"MATCH"}},
		"case 6":  {cfg: &vtMinimal, paths: []string{"README.md", "stats/internal/api/handler.go"}, code: 1, out: vtNone},
		"case 7":  {cfg: &vtMinimal, code: 1, out: vtNone},
		"case 8":  {cfg: &vtMinimal, paths: []string{"stats/web/src/components/dashboard/Panel.tsx"}, code: 0, out: vtMatched("stats/web/src/components/dashboard/Panel.tsx", "stats/web/src/**") + vtMatch},
		"case 9":  {cfg: vtStr(vtSection("`./stats/web/src/**`")), paths: []string{"stats/web/src/App.tsx"}, code: 0, out: vtMatched("stats/web/src/App.tsx", "./stats/web/src/**") + vtMatch},
		"case 10": {cfg: &vtMinimal, paths: []string{"./stats/web/src/App.tsx"}, code: 0, out: vtMatched("./stats/web/src/App.tsx", "stats/web/src/**") + vtMatch},
		"case 11": {cfg: vtStr(vtSection("`/stats/web/src/**`")), paths: []string{"stats/web/src/App.tsx"}, code: 0, out: vtMatched("stats/web/src/App.tsx", "/stats/web/src/**") + vtMatch},
		"case 12": {cfg: &vtMinimal, paths: []string{"/stats/web/src/App.tsx"}, code: 1, out: vtNone},
		"case 13": {cfg: &vtMinimal, paths: []string{"README.md"}, code: 1, out: vtNone},
		"case 14": {cfg: vtStr(vtSection("`stats/web/my folder/**`")), paths: []string{"stats/web/my folder/App.tsx"}, code: 0, out: vtMatched("stats/web/my folder/App.tsx", "stats/web/my folder/**") + vtMatch},
		"case 15": {cfg: vtStr(vtSection("`gymie-frontend/**, gymie-admin-frontend/**`")), paths: []string{"gymie-admin-frontend/src/App.tsx"}, code: 0,
			out: vtMatched("gymie-admin-frontend/src/App.tsx", "gymie-admin-frontend/**") + vtMatch},
		"case 16": {root: "/nonexistent/check-visual-trigger-test", paths: []string{"stats/web/src/App.tsx"}, code: 2,
			err: vtNoRoot + "$ROOT is not a directory — cannot tell whether it declares a visual verification section\n" + vtCannot + "not a directory\n"},
		"case 17": {setup: func(t *testing.T, root string) { mkdir(t, root+"/.flow/project.md") }, paths: []string{"stats/web/src/App.tsx"}, code: 2,
			err: vtNoRoot + "$ROOT/.flow/project.md is not a regular file — cannot resolve what it declares\n" + vtCannot + ".flow/project.md is not a regular file\n"},
		"case 18": {cfg: vtStr(vtMinimal + "\n\n" + vtMinimal), paths: []string{"stats/web/src/App.tsx"}, code: 2,
			err: vtNoRoot + "$ROOT/.flow/project.md declares 2 '## visual verification' sections — a second declaration is ambiguous, so neither was read\n" + vtCannot + "the section is declared more than once\n"},
		"case 19": {cfg: &vtMinimal, bom: true, paths: []string{"stats/web/src/App.tsx"}, code: 0,
			out: vtMatched("stats/web/src/App.tsx", "stats/web/src/**") + vtMatch, contains: []string{"MATCH"}, omits: []string{"not configured"}},
		"case 19b": {cfg: &vtMinimal, bom: true, paths: []string{"README.md"}, code: 1, out: vtNone, omits: []string{"not configured"}},
		"case 20": {cfg: vtStr(vtSection("`a/**`, `src/gateway/src/main/**`")), paths: []string{"src/gateway/src/main/resources/application.yml"}, code: 0,
			out: vtMatched("src/gateway/src/main/resources/application.yml", "src/gateway/src/main/**") + vtMatch, contains: []string{"matched `src/gateway/src/main/**`"}},
		"case 21": {cfg: vtStr(vtSection("`a/**`, `b/**`, `stats/web/src/**`")), paths: []string{"stats/web/src/App.tsx"}, code: 0,
			out: vtMatched("stats/web/src/App.tsx", "stats/web/src/**") + vtMatch, contains: []string{"matched `stats/web/src/**`"}},
		"case 22": {cfg: vtStr(vtSection("a/**, `stats/web/src/**`")), paths: []string{"a/foo.txt"}, code: 0,
			out: vtMatched("a/foo.txt", "a/**") + vtMatch, contains: []string{"matched `a/**`"}},
		"case 22b": {cfg: vtStr(vtSection("a/**, `stats/web/src/**`")), paths: []string{"stats/web/src/App.tsx"}, code: 0,
			out: vtMatched("stats/web/src/App.tsx", "stats/web/src/**") + vtMatch, contains: []string{"matched `stats/web/src/**`"}},
		"case 23": {cfg: vtStr(vtSection("`a/**`, `stats/web/my folder/**`")), paths: []string{"stats/web/my folder/App.tsx"}, code: 0,
			out: vtMatched("stats/web/my folder/App.tsx", "stats/web/my folder/**") + vtMatch, contains: []string{"matched `stats/web/my folder/**`"}},
		"case 24": {cfg: vtStr(vtSection("`" + longGlob + "`, `x/**`")), paths: []string{"src/very/long/path/that/exceeds/any/plausible/caller/variable/length/here/App.tsx"}, code: 0,
			out: vtMatched("src/very/long/path/that/exceeds/any/plausible/caller/variable/length/here/App.tsx", longGlob) + vtMatch, contains: []string{"matched `" + longGlob + "`"}},
		"case 25": {cfg: vtStr(vtSection("`a/**`, " + pad + "gymie-admin-frontend/**" + pad + ", `stats/web/src/**`")), paths: []string{"gymie-admin-frontend/src/App.tsx"}, code: 0,
			out: vtMatched("gymie-admin-frontend/src/App.tsx", "gymie-admin-frontend/**") + vtMatch, contains: []string{"matched `gymie-admin-frontend/**`"}},
		// The harness's third label carried the measured seconds; the bound
		// it asserted is the label here.
		"case 26": {cfg: vtStr(vtSection("`a/**`, " + bigPad + "gymie-admin-frontend/**" + bigPad + ", `stats/web/src/**`")), paths: []string{"gymie-admin-frontend/src/App.tsx"}, code: 0,
			out: vtMatched("gymie-admin-frontend/src/App.tsx", "gymie-admin-frontend/**") + vtMatch, contains: []string{"matched `gymie-admin-frontend/**`"},
			within: 5 * time.Second, label: "case 26: 60,000-backtick interior padding resolved in <= 5s"},
		"case 27": {cfg: vtStr(vtSection("`a\\|b/**`")), paths: []string{"a|b/App.tsx"}, code: 0,
			out: vtMatched("a|b/App.tsx", "a|b/**") + vtMatch, contains: []string{"matched `a|b/**`"}},
		// sanitize_display renders each literal backslash as two.
		"case 28": {cfg: vtStr(vtSection("`a\\\\b/**`")), paths: []string{`a\b/App.tsx`}, code: 0,
			out: vtMatched(`a\\b/App.tsx`, `a\\b/**`) + vtMatch, contains: []string{"matched `a\\\\b/**`"}},
		"case 29": {cfg: vtStr(vtSection("`a/**\\`")), paths: []string{`a/App.tsx\`}, code: 0,
			out: vtMatched(`a/App.tsx\\`, `a/**\\`) + vtMatch, contains: []string{"matched `a/**\\\\`"}},
		"case 30": {cfg: vtStr(vtSection("`a\\\\|b/**`")), paths: []string{"a/App.tsx"}, code: 2, err: vtAbsent,
			contains: []string{"`ui paths` is absent or empty"}},
		"case 31": {cfg: vtStr(vtSection("`a\\bc/**`")), paths: []string{`a\bc/App.tsx`}, code: 0,
			out: vtMatched(`a\\bc/App.tsx`, `a\\bc/**`) + vtMatch, contains: []string{"matched `a\\\\bc/**`"}},
	}
	for name, c := range cases {
		t.Run(fmt.Sprintf("%s: exit %d", name, c.code), func(t *testing.T) {
			t.Parallel()
			vtCheck(t, c)
		})
		for _, n := range c.contains {
			t.Run(name+": output names '"+n+"'", func(t *testing.T) {
				t.Parallel()
				if _, out, errOut, _ := vtRun(t, c); !strings.Contains(out+errOut, n) {
					t.Fatalf("output %q does not name %q", out+errOut, n)
				}
			})
		}
		for _, n := range c.omits {
			t.Run(name+": output correctly omits '"+n+"'", func(t *testing.T) {
				t.Parallel()
				if _, out, errOut, _ := vtRun(t, c); strings.Contains(out+errOut, n) {
					t.Fatalf("output %q names %q", out+errOut, n)
				}
			})
		}
		if c.within != 0 {
			t.Run(c.label, func(t *testing.T) {
				t.Parallel()
				start := time.Now()
				code, _, _, _ := vtRun(t, c)
				if took := time.Since(start); code != 0 || took > c.within {
					t.Fatalf("exit %d in %v, want exit 0 within %v — the quadratic cell splitter is back", code, took, c.within)
				}
			})
		}
	}

	// ---- beyond the harness: bash behaviours pinned from 3915fbc0 ---------
	t.Run("a relative root through a symlinked working directory prints the root as given", func(t *testing.T) {
		t.Parallel()
		base := t.TempDir()
		writeFile(t, base+"/real/proj/.flow/project.md", vtMinimal+"\n")
		if err := os.Symlink(base+"/real", base+"/link"); err != nil {
			t.Fatal(err)
		}
		fn := Registry["check-visual-trigger"]
		if fn == nil {
			t.Fatal("check-visual-trigger is not registered")
		}
		var out, errb bytes.Buffer
		code := fn([]string{"proj"}, Env{Getenv: os.Getenv, Dir: base + "/link",
			Stdin: strings.NewReader("stats/web/src/App.tsx\n")}, &out, &errb)
		want := vtMatched("stats/web/src/App.tsx", "stats/web/src/**") + strings.ReplaceAll(vtMatch, "$ROOT", "proj")
		if code != 0 || out.String() != want || errb.String() != "" {
			t.Fatalf("got exit %d stdout %q stderr %q, want 0 %q", code, out.String(), errb.String(), want)
		}
	})
	extra := []struct {
		label string
		args  []string // nil: the temp root
		c     vtCase
	}{
		{"no root argument is a usage error", []string{}, vtCase{code: 2, err: vtUsage}},
		{"two arguments is a usage error", []string{"a", "b"}, vtCase{code: 2, err: vtUsage}},
		{"an empty root is not a directory", []string{""}, vtCase{code: 2,
			err: vtNoRoot + " is not a directory — cannot tell whether it declares a visual verification section\nVISUAL-TRIGGER-CANNOT-ANSWER:  — not a directory\n"}},
		{"a final stdin line without a newline is read", nil, vtCase{cfg: &vtMinimal, stdin: "README.md\nstats/web/src/App.tsx", code: 0,
			out: vtMatched("stats/web/src/App.tsx", "stats/web/src/**") + vtMatch}},
		{"a NUL byte in a stdin line is dropped, as bash's read drops it", nil, vtCase{cfg: &vtMinimal, stdin: "stats/web/src/A\x00pp.tsx\n", code: 0,
			out: vtMatched("stats/web/src/App.tsx", "stats/web/src/**") + vtMatch}},
		{"blank stdin lines are skipped", nil, vtCase{cfg: &vtMinimal, stdin: "\n\n", code: 1, out: vtNone}},
		{"a path matching two globs prints a MATCH line per glob", nil, vtCase{cfg: vtStr(vtSection("`stats/**`, `stats/web/*/App.tsx`")), paths: []string{"stats/web/src/App.tsx"}, code: 0,
			out: vtMatched("stats/web/src/App.tsx", "stats/**") + vtMatched("stats/web/src/App.tsx", "stats/web/*/App.tsx") + vtMatch}},
		{"a non-ASCII glob matches its own bytes", nil, vtCase{cfg: vtStr(vtSection("`é/**`, `日本/**`")), paths: []string{"é/x", "日本/a.tsx", "e/x"}, code: 0,
			out: vtMatched("é/x", "é/**") + vtMatched("日本/a.tsx", "日本/**") + vtMatch}},
		{"a bare * does not span a slash", nil, vtCase{cfg: vtStr(vtSection("`stats/*`")), paths: []string{"stats/web/App.tsx"}, code: 1, out: vtNone}},
		{"? matches one byte except a slash", nil, vtCase{cfg: vtStr(vtSection("`a?b`, `c?d`")), paths: []string{"a/b", "cxd", "c\xc3\xa9d"}, code: 0,
			out: vtMatched("cxd", "c?d") + vtMatch}},
		{"a glob's regex metacharacters are literal", nil, vtCase{cfg: vtStr(vtSection("`a.(b)+{c}^$[d]`")), paths: []string{"axbc", "a.(b)+{c}^$[d]"}, code: 0,
			out: vtMatched("a.(b)+{c}^$[d]", "a.(b)+{c}^$[d]") + vtMatch}},
		{"a control byte in a matched path is escaped", nil, vtCase{cfg: vtStr(vtSection("`a/**`")), paths: []string{"a/\x1b[31mx"}, code: 0,
			out: vtMatched(`a/\x1b[31mx`, "a/**") + vtMatch}},
		{"a ui paths cell of separators and backticks alone resolves to no usable glob", nil, vtCase{cfg: vtStr(vtSection("` , `")), code: 2,
			err: vtNoRoot + "`ui paths` in $ROOT/.flow/project.md resolved to no usable glob\n" + vtCannot + "`ui paths` resolved to no usable glob\n"}},
		{"a NUL byte in the cell ends the row, as awk ends a record at it", nil, vtCase{cfg: vtStr(vtSection("`a/**\x00zz`, `b/**`")), paths: []string{"b/x"}, code: 1, out: vtNone}},
		{"a ui paths row outside the section is not read", nil, vtCase{cfg: vtStr("| `ui paths` | `a/**` |\n\n" + vtPre + vtPost), paths: []string{"a/x"}, code: 2, err: vtAbsent}},
		{"a deeper heading keeps the section open", nil, vtCase{cfg: vtStr(vtPre + "### sub\n| `ui paths` | `a/**` |\n"), paths: []string{"a/x"}, code: 0, out: vtMatched("a/x", "a/**") + vtMatch}},
		{"a dangling project.md symlink is not a regular file", nil, vtCase{setup: func(t *testing.T, root string) {
			if err := os.Symlink(root+"/nowhere", root+"/.flow/project.md"); err != nil {
				t.Fatal(err)
			}
		}, code: 2, err: vtNoRoot + "$ROOT/.flow/project.md is not a regular file — cannot resolve what it declares\n" + vtCannot + ".flow/project.md is not a regular file\n"}},
		{"an unreadable project.md is cannot-answer", nil, vtCase{cfg: &vtMinimal, setup: func(t *testing.T, root string) {
			if err := os.Chmod(root+"/.flow/project.md", 0); err != nil {
				t.Fatal(err)
			}
		}, code: 2, err: vtNoRoot + "$ROOT/.flow/project.md exists but is not readable — cannot resolve what it declares\n" + vtCannot + ".flow/project.md is not readable\n"}},
		// $ROOT reaches this line through sanitize_display; the other
		// refusals and the verdict line print it raw.
		{"a control byte in the root reaches the sanitized refusal escaped", nil, vtCase{sub: "/x\x1b", cfg: vtStr(vtPre + vtPost), code: 2,
			err: vtNoRoot + "`ui paths` is absent or empty in $SROOT/.flow/project.md — cannot resolve what to match against\n" + vtCannot + "`ui paths` is absent or empty\n"}},
		{"a control byte in the root reaches the verdict line raw", nil, vtCase{sub: "/x\x1b", cfg: &vtMinimal, code: 1, out: vtNone}},
	}
	for _, e := range extra {
		t.Run(e.label, func(t *testing.T) {
			t.Parallel()
			if e.args == nil {
				vtCheck(t, e.c)
				return
			}
			fn := Registry["check-visual-trigger"]
			if fn == nil {
				t.Fatal("check-visual-trigger is not registered")
			}
			var out, errb bytes.Buffer
			code := fn(e.args, Env{Getenv: os.Getenv, Dir: t.TempDir()}, &out, &errb)
			if code != e.c.code || out.String() != e.c.out || errb.String() != e.c.err {
				t.Fatalf("got exit %d stdout %q stderr %q, want %d %q %q", code, out.String(), errb.String(), e.c.code, e.c.out, e.c.err)
			}
		})
	}
	t.Run("an unreadable stdin is a refusal, not a verdict", func(t *testing.T) {
		t.Parallel()
		// compose-mockup-frames refuses the same failure with exit 2; "no
		// changed paths" would be a confident NO-MATCH on paths never read.
		root := t.TempDir()
		mkdir(t, root+"/.flow")
		writeFile(t, root+"/.flow/project.md", vtMinimal+"\n")
		var out, errb bytes.Buffer
		code := Registry["check-visual-trigger"]([]string{root}, Env{Getenv: os.Getenv, Dir: t.TempDir(), Stdin: iotest.ErrReader(errors.New("boom"))}, &out, &errb)
		if code != 2 || out.Len() != 0 || errb.String() != vtPrefix+"cannot read stdin: boom\nVISUAL-TRIGGER-CANNOT-ANSWER: "+root+" — stdin could not be read\n" {
			t.Fatalf("got exit %d stdout %q stderr %q", code, out.String(), errb.String())
		}
	})
	t.Run("a refusal answers before stdin is read", func(t *testing.T) {
		t.Parallel()
		// The bash refused before its read loop, so a slow producer on the
		// pipe never held a refusal back.
		var out, errb bytes.Buffer
		missing := t.TempDir() + "/missing"
		code := Registry["check-visual-trigger"]([]string{missing}, Env{Getenv: os.Getenv, Dir: t.TempDir(), Stdin: vtUnreadable{t}}, &out, &errb)
		if code != 2 || out.Len() != 0 || errb.String() != vtPrefix+missing+" is not a directory — cannot tell whether it declares a visual verification section\nVISUAL-TRIGGER-CANNOT-ANSWER: "+missing+" — not a directory\n" {
			t.Fatalf("got exit %d stdout %q stderr %q", code, out.String(), errb.String())
		}
	})
}

// vtUnreadable fails the test on any read: a refusal must not wait on stdin.
type vtUnreadable struct{ t *testing.T }

func (u vtUnreadable) Read([]byte) (int, error) {
	u.t.Error("stdin was read before the refusal")
	return 0, io.EOF
}

// TestVisualTriggerExit2Tokens pins the cause token every exit 2 ends its
// stderr with (design-verify.md, VH-25): NOT-CONFIGURED for the two
// "no section" answers a caller turns into `Visual: not configured`,
// CANNOT-ANSWER for every other refusal, and neither token on exit 0 or 1.
func TestVisualTriggerExit2Tokens(t *testing.T) {
	t.Parallel()
	const notCfg, cannot = "VISUAL-TRIGGER-NOT-CONFIGURED: ", "VISUAL-TRIGGER-CANNOT-ANSWER: "
	cases := []struct {
		label string
		c     vtCase
		token string // "" on exit 0/1: no token line at all
	}{
		{"no .flow/project.md", vtCase{paths: []string{"a"}}, notCfg},
		{"no visual verification section", vtCase{cfg: vtStr("# Project\n")}, notCfg},
		{"not a directory", vtCase{root: "/nonexistent/vt-token"}, cannot},
		{"project.md not a regular file", vtCase{setup: func(t *testing.T, root string) { mkdir(t, root+"/.flow/project.md") }}, cannot},
		{"project.md unreadable", vtCase{cfg: &vtMinimal, setup: func(t *testing.T, root string) {
			if err := os.Chmod(root+"/.flow/project.md", 0); err != nil {
				t.Fatal(err)
			}
		}}, cannot},
		{"duplicate section", vtCase{cfg: vtStr(vtMinimal + "\n\n" + vtMinimal)}, cannot},
		{"ui paths absent", vtCase{cfg: vtStr(vtPre + vtPost)}, cannot},
		{"ui paths no usable glob", vtCase{cfg: vtStr(vtSection("` , `"))}, cannot},
		{"matched", vtCase{cfg: &vtMinimal, paths: []string{"stats/web/src/App.tsx"}}, ""},
		{"unmatched", vtCase{cfg: &vtMinimal, paths: []string{"README.md"}}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			t.Parallel()
			code, out, errOut, root := vtRun(t, tc.c)
			vtAssertToken(t, code, out+errOut, errOut, tc.token, root)
		})
	}
	for _, args := range [][]string{{}, {"a", "b"}} {
		var out, errb bytes.Buffer
		code := Registry["check-visual-trigger"](args, Env{Getenv: os.Getenv, Dir: t.TempDir()}, &out, &errb)
		vtAssertToken(t, code, out.String()+errb.String(), errb.String(), cannot, "(no root)")
	}
	root := t.TempDir()
	writeFile(t, root+"/.flow/project.md", vtMinimal+"\n")
	var out, errb bytes.Buffer
	code := Registry["check-visual-trigger"]([]string{root}, Env{Getenv: os.Getenv, Dir: t.TempDir(), Stdin: iotest.ErrReader(errors.New("boom"))}, &out, &errb)
	vtAssertToken(t, code, out.String()+errb.String(), errb.String(), cannot, root)
}

// vtAssertToken: token "" wants exit 0/1 and no token anywhere; otherwise
// exit 2 with exactly one token line, stderr's last, naming root.
func vtAssertToken(t *testing.T, code int, all, errOut, token, root string) {
	t.Helper()
	if token == "" {
		if code == 2 || strings.Contains(all, "VISUAL-TRIGGER-NOT-CONFIGURED") || strings.Contains(all, "VISUAL-TRIGGER-CANNOT-ANSWER") {
			t.Fatalf("exit %d output %q: want a verdict and no exit-2 token", code, all)
		}
		return
	}
	lines := strings.Split(strings.TrimSuffix(errOut, "\n"), "\n")
	last := lines[len(lines)-1]
	if code != 2 || !strings.HasPrefix(last, token+root+" — ") ||
		strings.Count(all, "VISUAL-TRIGGER-NOT-CONFIGURED")+strings.Count(all, "VISUAL-TRIGGER-CANNOT-ANSWER") != 1 {
		t.Fatalf("exit %d stderr %q: want exit 2 ending in one %q line for %q", code, errOut, token, root)
	}
}
