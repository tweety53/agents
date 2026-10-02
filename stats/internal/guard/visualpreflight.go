package guard

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// checkVisualPreflight is scripts/check-visual-preflight.sh: the four
// workspace checks `flow.visual-verify` step 3 runs before any verifier is
// dispatched (skills/flow/visual-verify.md). The shim's header carries the
// contract; design-verify.md § VV-18 records why checks 2 and 3 are the
// mechanical rules below.
//
// Nothing read from the tree is executed: `lsof` and `node` are this guard's
// own commands, resolved on PATH, and the project's `start` command is only
// searched as text.
func init() { Registry["check-visual-preflight"] = checkVisualPreflight }

const vpUsage = "Usage: check-visual-preflight.sh <worktree> <app-root>=<resolved-url> [...] < KEY=value lines\n"

// vpProbeTimeout bounds each liveness probe of a held port.
const vpProbeTimeout = 5 * time.Second

// vpMaxFile is the size above which a file is not scanned: a generated
// bundle or a data dump, never hand-written configuration.
const vpMaxFile = 1 << 20

// vpSkipDirs are never descended into: VCS metadata, installed or built
// output, reports, fixtures and this pipeline's own scratch.
var vpSkipDirs = map[string]bool{
	".git": true, "node_modules": true, "dist": true, "build": true, "coverage": true,
	"test-results": true, "playwright-report": true, "testdata": true, "vendor": true, ".superpowers": true,
}

var (
	vpOriginsKey = regexp.MustCompile(`(?i)allowed[_-]?origins`)
	vpURL        = regexp.MustCompile(`https?://[^\s"'<>,;\]\}\)]+`)
	vpConfigExt  = map[string]bool{".json": true, ".yml": true, ".yaml": true, ".toml": true, ".ini": true, ".properties": true, ".conf": true}
)

// vpTarget is one app argument. base is the worktree it is checked
// against: <worktree>, or — for an app root inside a different git
// repository — that repository's worktree root, read as its own worktree.
type vpTarget struct{ root, url, origin, port, base, dir string }

// vpRow is one `## workspace isolation` resource row with the value
// prepare-workspace.sh exported for it.
type vpRow struct{ res, variable, def, value string }

func checkVisualPreflight(args []string, env Env, stdout, stderr io.Writer) int {
	refuse := func(format string, a ...any) int {
		fmt.Fprintf(stderr, "check-visual-preflight: "+format+"\n", a...)
		return 2
	}
	if len(args) < 2 {
		fmt.Fprint(stderr, vpUsage)
		return 2
	}
	wt := args[0]
	wtAbs := wiAbs(env, wt)
	if !isDir(wtAbs) {
		return refuse("%s is not a directory — cannot check the workspace", wt)
	}

	var apps []vpTarget
	for _, a := range args[1:] {
		root, raw, ok := strings.Cut(a, "=")
		u, err := url.Parse(raw)
		if !ok || root == "" || err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
			return refuse("argument %q is not <app-root>=<http(s) URL>", a)
		}
		dir := filepath.Join(wtAbs, root)
		if !isDir(dir) {
			return refuse("app root %s is not a directory under %s", root, wt)
		}
		base := wtAbs
		if r := vpRepoRoot(dir); r != "" && r != vpRepoRoot(wtAbs) {
			base = r
		}
		port := u.Port()
		if port == "" {
			port = map[string]string{"http": "80", "https": "443"}[u.Scheme]
		}
		apps = append(apps, vpTarget{root: root, url: raw, origin: u.Scheme + "://" + u.Host, port: port, base: base, dir: dir})
	}

	exported := map[string]string{}
	if env.Stdin != nil {
		sc := bufio.NewScanner(env.Stdin)
		for sc.Scan() {
			line := strings.TrimRight(sc.Text(), "\r")
			if line == "" {
				continue
			}
			k, v, ok := strings.Cut(line, "=")
			if !ok || !wiVariable.MatchString(k) {
				return refuse("stdin line %q is not KEY=value — pipe prepare-workspace.sh's output in", line)
			}
			exported[k] = v
		}
		if sc.Err() != nil {
			return refuse("cannot read stdin: %v", sc.Err())
		}
	}

	// Each worktree checks 2–4 run against: <worktree> first, then every
	// foreign repository an app root resolves into, with its own rows,
	// start command and apps.
	type vpBase struct {
		dir   string
		rows  []vpRow
		start string
		apps  []vpTarget
	}
	var bases []*vpBase
	byDir := map[string]*vpBase{}
	dirs := []string{wtAbs}
	for _, a := range apps {
		dirs = append(dirs, a.base)
	}
	for _, d := range dirs {
		if byDir[d] != nil {
			continue
		}
		cfg := filepath.Join(d, ".flow/project.md")
		body, err := os.ReadFile(cfg)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return refuse("cannot read %s: %v", cfg, err)
		}
		b := &vpBase{dir: d}
		var msg string
		if b.rows, msg = vpRows(cfg, body, exported); msg != "" {
			return refuse("%s", msg)
		}
		if b.start, msg = vpStart(body); msg != "" {
			return refuse("%s", msg)
		}
		byDir[d] = b
		bases = append(bases, b)
	}
	for _, a := range apps {
		byDir[a.base].apps = append(byDir[a.base].apps, a)
	}
	rows := bases[0].rows

	var fails, infos []string

	// Check 1 — ports held only by their own app.
	lsof, ok := lookPath(env, "lsof")
	if !ok {
		return refuse("lsof is not on PATH — cannot tell which ports are held")
	}
	type probe struct{ port, url string }
	var ports []probe
	seen := map[string]int{}
	add := func(p, u string) {
		if i, ok := seen[p]; ok {
			if ports[i].url == "" {
				ports[i].url = u
			}
			return
		}
		seen[p] = len(ports)
		ports = append(ports, probe{p, u})
	}
	for _, r := range rows {
		if r.res == "port" {
			add(r.value, "")
		}
	}
	for _, a := range apps {
		add(a.port, a.url)
	}
	for _, p := range ports {
		cmd := exec.Command(lsof, "-nP", "-iTCP:"+p.port, "-sTCP:LISTEN")
		out, err := cmd.Output()
		var ee *exec.ExitError
		switch {
		case err == nil && len(bytes.TrimSpace(out)) > 0:
		case err == nil, errors.As(err, &ee) && ee.ExitCode() == 1:
			continue // free
		default:
			return refuse("lsof -nP -iTCP:%s -sTCP:LISTEN failed (%v) — cannot tell whether the port is held", p.port, err)
		}
		if vpAnswers(p.port, p.url) {
			continue // the already-running stack steps 5 and 6 probe
		}
		var holder []string
		for _, l := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			if !strings.HasPrefix(l, "COMMAND") {
				holder = append(holder, strings.TrimSpace(l))
			}
		}
		pr := "GET " + p.url
		if p.url == "" {
			pr = "tcp 127.0.0.1:" + p.port + " and [::1]:" + p.port
		}
		fails = append(fails, fmt.Sprintf("FAIL: port %s — held by %s; probe %s went unanswered", p.port, strings.Join(holder, " | "), pr))
	}

	node, haveNode := lookPath(env, "node")
	for _, b := range bases {
		// Check 2 — the app's base-URL configuration resolves to the workspace's
		// value: a literal naming a moved row's default, on a line that does not
		// name the row's Variable, under an app root, for a row the `start`
		// command does not override.
		for _, r := range b.rows {
			if r.value == r.def || vpNames(b.start, r.variable) || (r.value != "" && strings.Contains(b.start, r.value)) {
				continue
			}
			// A default is matched only as a whole literal: no identifier
			// character right before or after it, so `shop` never hits
			// `shopping-cart` and `…:8080` never hits `…:80801`.
			lit := regexp.MustCompile(`(^|[^A-Za-z0-9_])` + regexp.QuoteMeta(r.def) + `([^A-Za-z0-9_]|$)`)
			if r.res == "port" {
				lit = regexp.MustCompile(":" + regexp.QuoteMeta(r.def) + `([^0-9]|$)`)
			}
			hit := lit.MatchString
			for _, a := range b.apps {
				vpWalk(wtAbs, a.dir, func(rel string) bool {
					return !vpMarkdown(rel)
				}, func(rel string, n int, l string, _ []string) {
					if hit(l) && !vpNames(l, r.variable) {
						fails = append(fails, fmt.Sprintf("FAIL: base-url — %s:%d: names %s's default %s, which neither $%s nor the start command overrides (resolved: %s)", rel, n, r.variable, r.def, r.variable, r.value))
					}
				})
			}
		}

		// Check 3 — origin allowed: every allowed-origins list naming at least
		// one URL, in a configuration-shaped file, carries every app's origin.
		vpWalk(wtAbs, b.dir, vpConfigFile, func(rel string, n int, l string, rest []string) {
			if !vpOriginsKey.MatchString(l) {
				return
			}
			text := l
			if i := strings.LastIndex(l, "["); i >= 0 && !strings.Contains(l[i:], "]") {
				for _, next := range rest {
					text += "\n" + next
					if strings.Contains(next, "]") {
						break
					}
				}
			}
			found := map[string]bool{}
			for _, m := range vpURL.FindAllString(text, -1) {
				if u, err := url.Parse(m); err == nil {
					found[u.Scheme+"://"+u.Host] = true
				}
			}
			if len(found) == 0 {
				return
			}
			for _, a := range b.apps {
				if !found[a.origin] {
					fails = append(fails, fmt.Sprintf("FAIL: origins — %s:%d: the allowed-origins list omits %s", rel, n, a.origin))
				}
			}
		})

		// Check 4 — one Playwright checkout per workspace.
		realWT, err := filepath.EvalSymlinks(b.dir)
		if err != nil {
			realWT = b.dir
		}
		for _, a := range b.apps {
			if !haveNode {
				infos = append(infos, fmt.Sprintf("INFO: playwright — %s: node is not on PATH, so the module was not resolved", a.root))
				continue
			}
			root := a.dir
			out, err := exec.Command(node, "-e", "console.log(require.resolve('@playwright/test/package.json', {paths: [process.argv[1]]}))", root).Output()
			got := strings.TrimSpace(string(out))
			if err != nil || got == "" {
				infos = append(infos, fmt.Sprintf("INFO: playwright — %s: @playwright/test does not resolve yet (setup installs it)", a.root))
				continue
			}
			real, err := filepath.EvalSymlinks(got)
			if err != nil {
				real = filepath.Clean(got)
			}
			if real != realWT && !strings.HasPrefix(real, realWT+string(filepath.Separator)) {
				fails = append(fails, fmt.Sprintf("FAIL: playwright — %s resolves @playwright/test to %s, outside %s — install this worktree's own before re-running", a.root, got, b.dir))
			}
		}

	}

	for _, l := range append(infos, fails...) {
		fmt.Fprintln(stdout, l)
	}
	if len(fails) > 0 {
		fmt.Fprintf(stdout, "PREFLIGHT-FAILED: %s — %d failing check line(s)\n", wt, len(fails))
		return 1
	}
	fmt.Fprintf(stdout, "PREFLIGHT-OK: %s — ports, base URL, origins and Playwright all pass\n", wt)
	return 0
}

// vpRows is the `## workspace isolation` resource rows, each paired with its
// exported value, parsed by check-workspace-isolation's own validator. A
// non-empty message is a reason this guard cannot answer.
func vpRows(cfg string, body []byte, exported map[string]string) ([]vpRow, string) {
	recs := lines(body)
	count := 0
	for i := range recs {
		recs[i], _, _ = strings.Cut(recs[i], "\x00")
		if ccHeading.MatchString(ccASCIILower(recs[i])) {
			count++
		}
	}
	if count == 0 {
		return nil, ""
	}
	if count > 1 {
		return nil, cfg + " declares more than one `## workspace isolation` section — run check-workspace-isolation.sh"
	}
	p := wiValidate(cfg, recs, wiSpace(false), false, true)
	if len(p.viol) > 0 {
		return nil, cfg + "'s `## workspace isolation` section is invalid — run check-workspace-isolation.sh"
	}
	var rows []vpRow
	for _, l := range p.rows {
		f := strings.Split(l, "\t") // #ROW, resource, variable, default, in a workspace
		r := vpRow{res: strings.ToLower(f[1]), variable: f[2], def: f[3]}
		if r.res == "cache index" {
			continue // prepare-workspace.sh exports nothing for it
		}
		v, ok := exported[r.variable]
		if !ok {
			return nil, "no exported value for " + r.variable + " on stdin — pipe prepare-workspace.sh <worktree> into this guard"
		}
		r.value = v
		rows = append(rows, r)
	}
	return rows, ""
}

// vpStart is the `## visual verification` section's `start` command, or "".
func vpStart(body []byte) (string, string) {
	if vvHeadingCount(string(body)) > 1 {
		return "", ".flow/project.md declares more than one `## visual verification` section"
	}
	for _, l := range vvSectionLines(string(body)) {
		if c := vtSplitCells(l.Text); len(c) >= 2 && vtFoldCell(c[0]) == "start" {
			return c[1], ""
		}
	}
	return "", ""
}

// vpAnswers is whether a held port answers its probe: an HTTP GET of the
// app URL the port came through (any response), else a TCP connect.
func vpAnswers(port, rawURL string) bool {
	if rawURL != "" {
		c := &http.Client{
			Timeout:       vpProbeTimeout,
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
			// A liveness probe of a local dev server: a self-signed
			// certificate still answers.
			Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
		}
		resp, err := c.Get(rawURL)
		if err != nil {
			return false
		}
		resp.Body.Close()
		return true
	}
	for _, host := range []string{"127.0.0.1", "::1"} {
		if conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), vpProbeTimeout); err == nil {
			conn.Close()
			return true
		}
	}
	return false
}

// vpNames is whether text names the variable as a whole word.
func vpNames(text, variable string) bool {
	return regexp.MustCompile(`(^|[^A-Za-z0-9_])` + regexp.QuoteMeta(variable) + `([^A-Za-z0-9_]|$)`).MatchString(text)
}

func vpMarkdown(rel string) bool {
	switch strings.ToLower(filepath.Ext(rel)) {
	case ".md", ".markdown", ".mdx":
		return true
	}
	return false
}

// vpConfigFile is whether a file is configuration-shaped: `.env*`, a
// config-format extension, or `*.config.*`. Source files are not, so a
// guard's own tests and fixtures never trigger check 3.
func vpConfigFile(rel string) bool {
	base := filepath.Base(rel)
	return strings.HasPrefix(base, ".env") || vpConfigExt[strings.ToLower(filepath.Ext(base))] || strings.Contains(base, ".config.")
}

// vpWalk calls visit for every line of every scannable file under dir whose
// worktree-relative path passes want, with the lines after it. Skipped:
// vpSkipDirs, a nested checkout (a directory carrying its own .git — a
// sibling worktree is not this one), non-regular files, files over
// vpMaxFile, and binaries (a NUL byte).
func vpWalk(wt, dir string, want func(string) bool, visit func(rel string, n int, line string, rest []string)) {
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if path != dir && path != wt {
				if vpSkipDirs[d.Name()] {
					return filepath.SkipDir
				}
				if _, err := os.Lstat(filepath.Join(path, ".git")); err == nil {
					return filepath.SkipDir
				}
			}
			return nil
		}
		rel, _ := filepath.Rel(wt, path)
		if !d.Type().IsRegular() || !want(rel) {
			return nil
		}
		if fi, err := d.Info(); err != nil || fi.Size() > vpMaxFile {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil || bytes.IndexByte(b, 0) >= 0 {
			return nil
		}
		ls := lines(b)
		for i, l := range ls {
			visit(rel, i+1, l, ls[i+1:])
		}
		return nil
	})
}

// vpRepoRoot is the nearest directory at or above dir carrying a `.git`
// entry (a checkout's directory or a linked worktree's file), or "".
func vpRepoRoot(dir string) string {
	for d := filepath.Clean(dir); ; d = filepath.Dir(d) {
		if _, err := os.Lstat(filepath.Join(d, ".git")); err == nil {
			return d
		}
		if filepath.Dir(d) == d {
			return ""
		}
	}
}
