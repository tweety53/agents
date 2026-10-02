package guard

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// check-visual-preflight is visual-verify.md step 3's four checks as one
// guard (design-verify.md § VV-18). Every case runs the registered guard
// in-process against a real tree in t.TempDir(); `lsof` and `node` are fake
// scripts on a PATH that holds nothing else, so the machine's own listeners
// and Node install never decide a verdict — except in the cases that run the
// real `lsof` and `node` on purpose, which skip when the tool is absent.

// vpIsolation is a `## workspace isolation` section with one `port` row and
// one `url` row derived from it; vpStdin is what prepare-workspace.sh
// exports for it at offset 13.
const (
	vpIsolation = "## workspace isolation\n\n| Resource | Variable | Default | In a workspace |\n|---|---|---|---|\n" +
		"| `port` | `API_PORT` | `8080` | `+<offset>` |\n" +
		"| `url` | `API_BASE_URL` | `http://localhost:8080` | `http://localhost:<value:API_PORT>` |\n"
	vpStdin = "API_PORT=8093\nAPI_BASE_URL=http://localhost:8093\n"
	vpApp   = "web=http://localhost:8093"
)

// vpBin is a directory holding a fake lsof that reports each port in held
// as listened on (exit 0, a header and one row) and every other port free
// (exit 1), plus a fake node running nodeBody when it is non-empty.
func vpBin(t *testing.T, held []int, nodeBody string) string {
	t.Helper()
	bin := t.TempDir()
	var b strings.Builder
	b.WriteString("#!/bin/sh\ncase \" $* \" in\n")
	for _, p := range held {
		fmt.Fprintf(&b, "*\" -iTCP:%d \"*) printf 'COMMAND PID USER FD TYPE DEVICE SIZE/OFF NODE NAME\\nholder 4242 me 3u IPv4 0x0 0t0 TCP 127.0.0.1:%d (LISTEN)\\n'; exit 0;;\n", p, p)
	}
	b.WriteString("esac\nexit 1\n")
	writeExec(t, filepath.Join(bin, "lsof"), b.String())
	if nodeBody != "" {
		writeExec(t, filepath.Join(bin, "node"), "#!/bin/sh\n"+nodeBody)
	}
	return bin
}

// onlyPath is an Env.Getenv whose PATH is path alone.
func onlyPath(path string) func(string) string {
	return func(k string) string {
		if k == "PATH" {
			return path
		}
		return os.Getenv(k)
	}
}

func vpRun(t *testing.T, wt string, apps []string, stdin, path string) (int, string, string) {
	t.Helper()
	g := Registry["check-visual-preflight"]
	if g == nil {
		t.Fatal("no check-visual-preflight guard in Registry")
	}
	var out, errb bytes.Buffer
	code := g(append([]string{wt}, apps...), Env{Getenv: onlyPath(path), Stdin: strings.NewReader(stdin)}, &out, &errb)
	return code, out.String(), errb.String()
}

// vpTree writes files (relative path → body) under a fresh worktree.
func vpTree(t *testing.T, files map[string]string) string {
	t.Helper()
	wt := t.TempDir()
	mkdir(t, filepath.Join(wt, "web"))
	for rel, body := range files {
		writeFile(t, filepath.Join(wt, rel), body)
	}
	return wt
}

// freePort is a port nothing listens on at the moment it is returned.
func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	p := l.Addr().(*net.TCPAddr).Port
	l.Close()
	return p
}

func vpPort(t *testing.T, rawURL string) int {
	t.Helper()
	_, p, err := net.SplitHostPort(strings.TrimPrefix(rawURL, "http://"))
	if err != nil {
		t.Fatal(err)
	}
	var n int
	fmt.Sscan(p, &n)
	return n
}

type vpWant struct {
	code     int
	contains []string
	omits    []string
}

func (w vpWant) check(t *testing.T, code int, out, errOut string) {
	t.Helper()
	if code != w.code {
		t.Errorf("exit %d, want %d\nstdout:\n%s\nstderr:\n%s", code, w.code, out, errOut)
	}
	for _, s := range w.contains {
		if !strings.Contains(out+errOut, s) {
			t.Errorf("output lacks %q\nstdout:\n%s\nstderr:\n%s", s, out, errOut)
		}
	}
	for _, s := range w.omits {
		if strings.Contains(out+errOut, s) {
			t.Errorf("output carries %q\nstdout:\n%s\nstderr:\n%s", s, out, errOut)
		}
	}
}

func TestCheckVisualPreflightPorts(t *testing.T) {
	t.Parallel()

	t.Run("free", func(t *testing.T) {
		wt := vpTree(t, nil)
		p := freePort(t)
		code, out, errOut := vpRun(t, wt, []string{fmt.Sprintf("web=http://127.0.0.1:%d", p)}, "", vpBin(t, nil, ""))
		vpWant{code: 0, contains: []string{"PREFLIGHT-OK: " + wt}, omits: []string{"FAIL:"}}.check(t, code, out, errOut)
	})

	t.Run("held and answering", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNotFound) }))
		defer srv.Close()
		wt := vpTree(t, nil)
		code, out, errOut := vpRun(t, wt, []string{"web=" + srv.URL}, "", vpBin(t, []int{vpPort(t, srv.URL)}, ""))
		vpWant{code: 0, contains: []string{"PREFLIGHT-OK: "}, omits: []string{"FAIL:"}}.check(t, code, out, errOut)
	})

	t.Run("held and silent", func(t *testing.T) {
		wt := vpTree(t, nil)
		p := freePort(t)
		url := fmt.Sprintf("http://127.0.0.1:%d", p)
		code, out, errOut := vpRun(t, wt, []string{"web=" + url}, "", vpBin(t, []int{p}, ""))
		vpWant{code: 1, contains: []string{
			fmt.Sprintf("FAIL: port %d", p),
			fmt.Sprintf("holder 4242 me 3u IPv4 0x0 0t0 TCP 127.0.0.1:%d (LISTEN)", p),
			"GET " + url,
			"PREFLIGHT-FAILED: " + wt,
		}}.check(t, code, out, errOut)
	})

	t.Run("port row held and answering", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		defer l.Close()
		p := l.Addr().(*net.TCPAddr).Port
		wt := vpTree(t, map[string]string{".flow/project.md": vpIsolation})
		stdin := fmt.Sprintf("API_PORT=%d\nAPI_BASE_URL=http://localhost:%d\n", p, p)
		app := fmt.Sprintf("web=http://127.0.0.1:%d", freePort(t))
		code, out, errOut := vpRun(t, wt, []string{app}, stdin, vpBin(t, []int{p}, ""))
		vpWant{code: 0, omits: []string{"FAIL:"}}.check(t, code, out, errOut)
	})

	t.Run("port row held and silent", func(t *testing.T) {
		p := freePort(t)
		wt := vpTree(t, map[string]string{".flow/project.md": vpIsolation})
		stdin := fmt.Sprintf("API_PORT=%d\nAPI_BASE_URL=http://localhost:%d\n", p, p)
		app := fmt.Sprintf("web=http://127.0.0.1:%d", freePort(t))
		code, out, errOut := vpRun(t, wt, []string{app}, stdin, vpBin(t, []int{p}, ""))
		vpWant{code: 1, contains: []string{fmt.Sprintf("FAIL: port %d", p), fmt.Sprintf("tcp 127.0.0.1:%d", p)}}.check(t, code, out, errOut)
	})

	t.Run("lsof missing", func(t *testing.T) {
		wt := vpTree(t, nil)
		code, out, errOut := vpRun(t, wt, []string{vpApp}, "", t.TempDir())
		vpWant{code: 2, contains: []string{"lsof"}, omits: []string{"PREFLIGHT-OK"}}.check(t, code, out, errOut)
	})

	t.Run("lsof cannot answer", func(t *testing.T) {
		wt := vpTree(t, nil)
		bin := t.TempDir()
		writeExec(t, filepath.Join(bin, "lsof"), "#!/bin/sh\nexit 7\n")
		code, out, errOut := vpRun(t, wt, []string{vpApp}, "", bin)
		vpWant{code: 2, contains: []string{"lsof"}, omits: []string{"PREFLIGHT-OK"}}.check(t, code, out, errOut)
	})

	// The real lsof, against a real listener: the boundary the fakes stand in for.
	t.Run("real lsof sees a real answering listener", func(t *testing.T) {
		lsof, err := exec.LookPath("lsof")
		if err != nil {
			if _, statErr := os.Stat("/usr/sbin/lsof"); statErr != nil {
				t.Skip("no lsof on this machine")
			}
			lsof = "/usr/sbin/lsof"
		}
		srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
		defer srv.Close()
		wt := vpTree(t, nil)
		code, out, errOut := vpRun(t, wt, []string{"web=" + srv.URL}, "", filepath.Dir(lsof))
		vpWant{code: 0, contains: []string{"PREFLIGHT-OK: "}}.check(t, code, out, errOut)
	})

	t.Run("usage", func(t *testing.T) {
		code, out, errOut := vpRun(t, vpTree(t, nil), nil, "", vpBin(t, nil, ""))
		vpWant{code: 2, contains: []string{"Usage: check-visual-preflight.sh"}}.check(t, code, out, errOut)
	})

	t.Run("stdin missing a declared row's value", func(t *testing.T) {
		wt := vpTree(t, map[string]string{".flow/project.md": vpIsolation})
		code, out, errOut := vpRun(t, wt, []string{vpApp}, "", vpBin(t, nil, ""))
		vpWant{code: 2, contains: []string{"API_PORT"}}.check(t, code, out, errOut)
	})
}

func TestCheckVisualPreflightBaseURL(t *testing.T) {
	t.Parallel()
	start := func(cmd string) string {
		return "\n## visual verification\n\n| Setting | Value |\n|---|---|\n| `ui paths` | `web/src/**` |\n| `screenshots` | `web/tests` |\n\n" +
			"| Command | Runs |\n|---|---|\n| `verify` | `npm test` |\n| `capture` | `npx playwright test <spec>` |\n| `start` | `" + cmd + "` |\n"
	}
	cases := []struct {
		name  string
		files map[string]string
		want  vpWant
	}{
		{"the url default literal hardcoded", map[string]string{"web/src/api.ts": "export const base = \"http://localhost:8080/api\"\n"},
			vpWant{code: 1, contains: []string{"FAIL: base-url — web/src/api.ts:1:", "API_BASE_URL"}}},
		{":<default> matched", map[string]string{"web/src/api.ts": "// base\nfetch(\"//localhost:8080/x\")\n"},
			vpWant{code: 1, contains: []string{"FAIL: base-url — web/src/api.ts:2:", "API_PORT"}, omits: []string{"API_BASE_URL"}}},
		{"a bare number not matched", map[string]string{"web/playwright.config.ts": "// the daemon on 8080\nconst other = \"localhost:80801\"\n"},
			vpWant{code: 0, omits: []string{"FAIL:"}}},
		{"a default inside a longer literal not matched", map[string]string{"web/src/api.ts": "const other = \"http://localhost:80801/x\"\n"},
			vpWant{code: 0, omits: []string{"FAIL:"}}},
		{"a line naming the Variable overridden", map[string]string{"web/src/api.ts": "const port = process.env.API_PORT ?? \":8080\"\n"},
			vpWant{code: 0, omits: []string{"FAIL:"}}},
		{"a start command naming the Variable overrides the row", map[string]string{
			".flow/project.md": vpIsolation + start("API_PORT=$API_PORT npm run dev"),
			"web/src/api.ts":   "fetch(\"//localhost:8080/x\")\n"},
			vpWant{code: 0, omits: []string{"FAIL:"}}},
		{"a start command naming the resolved value overrides the row", map[string]string{
			".flow/project.md": vpIsolation + start("npm run dev -- --port 8093"),
			"web/src/api.ts":   "fetch(\"//localhost:8080/x\")\n"},
			vpWant{code: 0, omits: []string{"FAIL:"}}},
		{"a start command overrides only the row it names", map[string]string{
			".flow/project.md": vpIsolation + start("API_PORT=$API_PORT npm run dev"),
			"web/src/api.ts":   "export const base = \"http://localhost:8080/api\"\n"},
			vpWant{code: 1, contains: []string{"FAIL: base-url — web/src/api.ts:1:", "API_BASE_URL"}}},
		{"skipped dirs, Markdown, binaries and paths outside the app root", map[string]string{
			"web/node_modules/x/index.js": "fetch(\"//localhost:8080\")\n",
			"web/dist/app.js":             "fetch(\"//localhost:8080\")\n",
			"web/testdata/fixture.ts":     "fetch(\"//localhost:8080\")\n",
			"web/README.md":               "Runs on http://localhost:8080\n",
			"web/logo.bin":                "\x00\x01fetch(\"//localhost:8080\")\n",
			"server/main.go":              "addr := \"localhost:8080\"\n"},
			vpWant{code: 0, omits: []string{"FAIL:"}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			files := map[string]string{".flow/project.md": vpIsolation}
			for k, v := range c.files {
				files[k] = v
			}
			wt := vpTree(t, files)
			code, out, errOut := vpRun(t, wt, []string{vpApp}, vpStdin, vpBin(t, nil, ""))
			c.want.check(t, code, out, errOut)
		})
	}

	t.Run("a row whose value is its default is not scanned", func(t *testing.T) {
		t.Parallel()
		wt := vpTree(t, map[string]string{".flow/project.md": vpIsolation, "web/src/api.ts": "fetch(\"http://localhost:8080\")\n"})
		code, out, errOut := vpRun(t, wt, []string{"web=http://localhost:8080"}, "API_PORT=8080\nAPI_BASE_URL=http://localhost:8080\n", vpBin(t, nil, ""))
		vpWant{code: 0, omits: []string{"FAIL:"}}.check(t, code, out, errOut)
	})

	t.Run("an invalid isolation section cannot be answered", func(t *testing.T) {
		t.Parallel()
		bad := strings.Replace(vpIsolation, "`+<offset>`", "`9090`", 1)
		wt := vpTree(t, map[string]string{".flow/project.md": bad})
		code, out, errOut := vpRun(t, wt, []string{vpApp}, vpStdin, vpBin(t, nil, ""))
		vpWant{code: 2, contains: []string{"check-workspace-isolation.sh"}}.check(t, code, out, errOut)
	})
}

func TestCheckVisualPreflightOrigins(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		files map[string]string
		want  vpWant
	}{
		{"a list missing the app origin", map[string]string{"config/app.yaml": "cors:\n  allowed_origins: [\"http://localhost:8080\"]\n"},
			vpWant{code: 1, contains: []string{"FAIL: origins — config/app.yaml:2:", "http://localhost:8093"}}},
		{"a list carrying the app origin", map[string]string{".env": "ALLOWED_ORIGINS=http://localhost:8093,http://example.test\n"},
			vpWant{code: 0, omits: []string{"FAIL:"}}},
		{"a bracketed list across lines, carrying it", map[string]string{"server/cors.json": "{\n  \"allowedOrigins\": [\n    \"http://localhost:8093/\"\n  ]\n}\n"},
			vpWant{code: 0, omits: []string{"FAIL:"}}},
		{"a bracketed list across lines, missing it", map[string]string{"server/cors.json": "{\n  \"allowedOrigins\": [\n    \"http://localhost:8080\"\n  ]\n}\n"},
			vpWant{code: 1, contains: []string{"FAIL: origins — server/cors.json:2:"}}},
		{"a key naming no URL", map[string]string{".env.local": "ALLOWED_ORIGINS=${ORIGINS}\n"},
			vpWant{code: 0, omits: []string{"FAIL:"}}},
		{"source files excluded", map[string]string{"server/cors.go": "var allowed_origins = []string{\"http://localhost:8080\"}\n"},
			vpWant{code: 0, omits: []string{"FAIL:"}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			files := map[string]string{".flow/project.md": vpIsolation}
			for k, v := range c.files {
				files[k] = v
			}
			wt := vpTree(t, files)
			code, out, errOut := vpRun(t, wt, []string{vpApp}, vpStdin, vpBin(t, nil, ""))
			c.want.check(t, code, out, errOut)
		})
	}
}

func TestCheckVisualPreflightPlaywright(t *testing.T) {
	t.Parallel()

	t.Run("resolved outside the worktree", func(t *testing.T) {
		wt := vpTree(t, nil)
		code, out, errOut := vpRun(t, wt, []string{vpApp}, "", vpBin(t, nil, "echo /elsewhere/node_modules/@playwright/test/package.json\n"))
		vpWant{code: 1, contains: []string{"FAIL: playwright — web", "/elsewhere/node_modules/@playwright/test/package.json"}}.check(t, code, out, errOut)
	})

	t.Run("resolved inside the worktree", func(t *testing.T) {
		wt := vpTree(t, map[string]string{"web/node_modules/@playwright/test/package.json": "{}\n"})
		body := "echo " + filepath.Join(wt, "web/node_modules/@playwright/test/package.json") + "\n"
		code, out, errOut := vpRun(t, wt, []string{vpApp}, "", vpBin(t, nil, body))
		vpWant{code: 0, omits: []string{"FAIL:", "INFO:"}}.check(t, code, out, errOut)
	})

	t.Run("unresolvable", func(t *testing.T) {
		wt := vpTree(t, nil)
		code, out, errOut := vpRun(t, wt, []string{vpApp}, "", vpBin(t, nil, "echo 'Cannot find module' >&2\nexit 1\n"))
		vpWant{code: 0, contains: []string{"INFO: playwright — web"}, omits: []string{"FAIL:"}}.check(t, code, out, errOut)
	})

	t.Run("node absent", func(t *testing.T) {
		wt := vpTree(t, nil)
		code, out, errOut := vpRun(t, wt, []string{vpApp}, "", vpBin(t, nil, ""))
		vpWant{code: 0, contains: []string{"INFO: playwright — web"}, omits: []string{"FAIL:"}}.check(t, code, out, errOut)
	})

	// The real node's module resolution: inside the worktree passes; a
	// checkout in a directory above the worktree is the cross-worktree
	// conflict the check exists for.
	realNode := func(t *testing.T) string {
		node, err := exec.LookPath("node")
		if err != nil {
			t.Skip("no node on this machine")
		}
		return filepath.Dir(node)
	}
	t.Run("real node resolving inside", func(t *testing.T) {
		nodeDir := realNode(t)
		wt := vpTree(t, map[string]string{"web/node_modules/@playwright/test/package.json": "{\"name\":\"@playwright/test\"}\n"})
		code, out, errOut := vpRun(t, wt, []string{vpApp}, "", vpBin(t, nil, "")+":"+nodeDir)
		vpWant{code: 0, omits: []string{"FAIL:", "INFO:"}}.check(t, code, out, errOut)
	})
	t.Run("real node resolving in a checkout above the worktree", func(t *testing.T) {
		nodeDir := realNode(t)
		outer := t.TempDir()
		writeFile(t, filepath.Join(outer, "node_modules/@playwright/test/package.json"), "{\"name\":\"@playwright/test\"}\n")
		wt := filepath.Join(outer, "wt")
		mkdir(t, filepath.Join(wt, "web"))
		code, out, errOut := vpRun(t, wt, []string{vpApp}, "", vpBin(t, nil, "")+":"+nodeDir)
		vpWant{code: 1, contains: []string{"FAIL: playwright — web", "node_modules/@playwright/test/package.json"}}.check(t, code, out, errOut)
	})

	// This repository, as it stands, passes: its one app (stats/web on the
	// UI-test stack) hardcodes no isolated default and pins no origin list.
	// lsof reports every port free and node is absent, so only checks 2–3
	// decide — against the real .flow/project.md and the real tree.
	t.Run("this repository is clean", func(t *testing.T) {
		repo, err := filepath.Abs("../../..")
		if err != nil {
			t.Fatal(err)
		}
		stdin := "FLOWD_DSN=postgres://flow:flow@localhost:5433/flow_kan_x?sslmode=disable\nFLOWD_PORT=4180\nFLOW_ADDR=http://127.0.0.1:4180\nFLOW_RECORDS_ADDR=http://127.0.0.1:4173\n"
		code, out, errOut := vpRun(t, repo, []string{"stats/web=http://127.0.0.1:4174"}, stdin, vpBin(t, nil, ""))
		vpWant{code: 0, contains: []string{"PREFLIGHT-OK: "}, omits: []string{"FAIL:"}}.check(t, code, out, errOut)
	})
}

// An app root inside a different git repository than <worktree> is checked
// as its own worktree — its own `.flow/project.md` rows and start command —
// so <worktree>'s moved rows are never judged against the other
// repository's files, while per-port probing still covers both.
func TestCheckVisualPreflightCrossRepo(t *testing.T) {
	t.Parallel()
	top := t.TempDir()
	be, fe := filepath.Join(top, "backend"), filepath.Join(top, "frontend")
	writeFile(t, filepath.Join(be, ".git"), "gitdir: /elsewhere\n")
	writeFile(t, filepath.Join(be, ".flow/project.md"), vpIsolation)
	mkdir(t, filepath.Join(fe, ".git"))
	writeFile(t, filepath.Join(fe, "src/api.ts"), "fetch(\"//localhost:8080/x\")\n")

	code, out, errOut := vpRun(t, be, []string{"../frontend=http://localhost:8093"}, vpStdin, vpBin(t, nil, ""))
	vpWant{code: 0, contains: []string{"PREFLIGHT-OK: " + be}, omits: []string{"FAIL:"}}.check(t, code, out, errOut)

	// The same tree with the app root in <worktree>'s own repository keeps
	// today's verdict: the moved row's default fails.
	mkdir(t, filepath.Join(be, "web"))
	writeFile(t, filepath.Join(be, "web/api.ts"), "fetch(\"//localhost:8080/x\")\n")
	code, out, errOut = vpRun(t, be, []string{vpApp}, vpStdin, vpBin(t, nil, ""))
	vpWant{code: 1, contains: []string{"FAIL: base-url — web/api.ts:1:"}}.check(t, code, out, errOut)
}
