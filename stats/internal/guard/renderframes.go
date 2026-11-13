package guard

import (
	"archive/tar"
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image/png"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// flow-guard render-frames — render the design repository's tagged frames a
// plan or a sidecar cites into the project's `mockups` root, so every frame
// guard resolves them like any committed frame (kan-1096, Q13, Q15, Q21).
//
// Usage: flow-guard render-frames <project root> <file>...
//
// <project root> is the directory `.flow/project.md` lives in; each <file> a
// plan artifact (design.md, tasks.md) or a `<spec>.mockups` sidecar.
//
// A FRAME REFERENCE is `<tag>/<frame id>`, written two ways:
//   - in any file, a citation `<mockups root>/<tag>/<frame id>.png`, not
//     preceded by a letter, digit, `_`, `.`, `/` or `-` — the form a plan
//     cites a frame by (skills/flow/brainstorm-planner.md, section D);
//   - in a `.mockups` sidecar, a pair line's frame field `<tag>/<frame id>`.
// A <frame id> `dk-<id>` names the second (dark) phone of frame <id>.
//
// Each reference not yet rendered — its PNG absent or blank (below the top
// eighth, under 0.5% of pixels off the centre colour: a lazy pane caught
// before it painted) — is rendered into
// `<project root>/<mockups root>/<tag>/`, a directory this guard creates
// with a `.gitignore` of `*`, so nothing in it is ever committed. A <tag>
// equal to a committed sub-directory of the mockups root is refused (exit
// 2): its frames would be the committed ones, not the tag's.
//
// Per tag: the tag must resolve in the `design repo` clone; a missing tag is
// fetched from the declared remote once (`git fetch <remote>
// refs/tags/<tag>:refs/tags/<tag>`). The tree is exported with `git archive`
// into a temp directory, so the clone, its work tree and any server on it
// are never touched. The frame is found by its id alone: exactly one
// `id="<id>" data-frame` across the tree's top-level `*.html` pages
// (`_gen/` and other sub-directories hold generator sources that repeat
// ids). The renderer runs `node` with the program below on the `regression
// checkout`'s own `@playwright/test` (flow-guard check-visual-preflight's
// precedent): one page at DPR 2, requests to http://design.local/ fulfilled
// from the exported tree (no server process), the frame scrolled into view
// and its lazy panes filled — the data-mig-filled flag set and every
// [data-mig-pane] holding an element, since the page sets the flag when it
// queues the panes and fills them later (KAN-1087). Each of the frame's phones — outermost
// elements 390 css px wide with 28 px corners, in document order — is
// screenshotted at an integer-aligned 390×844 box with animations disabled:
// the first as `<id>.png`, the second as `dk-<id>.png`; a one-phone
// (light-only) frame gets no `dk-`. A render that comes out blank is
// deleted and reported as a finding, never left as the frame.
//
// Prints one `RENDER-FINDING: <tag>/<id> — <why>` line per reference that
// cannot be rendered, then ONE verdict line:
//   FRAMES-RENDERED: <n> rendered, <m> already present
//   FRAMES-MISSING:  <k> reference(s) not rendered
//
// Exit 0 every reference rendered or present; 1 a reference whose id is not
// at its tag (absent or not unique), whose phone is not found, or that
// renders blank; 2 cannot
// answer: a usage error, an unreadable <file>, no `design repo`, `mockups`
// or `regression checkout` row, the clone missing, a tag unfetchable or
// unexportable, a tag naming a committed sub-directory, `node` or
// `@playwright/test` unresolvable, or the renderer failing.
//
// Nothing read from the files is run through a shell: the tag and id pass
// a strict pattern, and git and node take them as arguments.

func init() { Registry["render-frames"] = renderFrames }

var (
	rfSidecarFrame = regexp.MustCompile(`^([A-Za-z0-9][A-Za-z0-9_.-]*)/([A-Za-z0-9_-]+)$`)
	rfDark         = "dk-"
)

// rfRef is one `<tag>/<frame id>` reference.
type rfRef struct{ tag, id string }

func renderFrames(args []string, env Env, stdout, stderr io.Writer) int {
	refuse := func(format string, a ...any) int {
		fmt.Fprintf(stderr, "render-frames: "+format+"\n", a...)
		return 2
	}
	if len(args) < 2 {
		return refuse("usage: flow-guard render-frames <project root> <file>...")
	}
	root := wiAbs(env, args[0])
	cfg, _ := os.ReadFile(filepath.Join(root, ".flow", "project.md"))
	cells := vvCells(string(cfg))
	design := strings.Fields(strings.ReplaceAll(cells["design repo"], "`", ""))
	if len(design) != 2 {
		return refuse("%s/.flow/project.md declares no `design repo` row of `<clone> <remote>` under `## visual verification`", args[0])
	}
	clone, remote := cellPath(root, design[0]), design[1]
	mockups := strings.TrimSuffix(strings.Trim(cells["mockups"], "`"), "/")
	if mockups == "" {
		return refuse("%s/.flow/project.md declares no `mockups` row — there is no root to render into", args[0])
	}
	checkout := cellPath(root, cells["regression checkout"])
	if checkout == "" {
		return refuse("%s/.flow/project.md declares no `regression checkout` row — there is no Playwright to render with", args[0])
	}
	if vvGitClean(env, "-C", clone, "rev-parse", "--git-dir").Run() != nil {
		return refuse("the `design repo` clone %s is missing or not a git checkout", clone)
	}

	cite := regexp.MustCompile(`(^|[^A-Za-z0-9_./-])` + regexp.QuoteMeta(mockups) + `/([A-Za-z0-9][A-Za-z0-9_.-]*)/([A-Za-z0-9_-]+)\.png`)
	byTag := map[string][]string{}
	seen := map[rfRef]bool{}
	add := func(tag, id string) {
		if r := (rfRef{tag, id}); !seen[r] {
			seen[r] = true
			byTag[tag] = append(byTag[tag], id)
		}
	}
	for _, file := range args[1:] {
		data, err := os.ReadFile(wiAbs(env, file))
		if err != nil {
			return refuse("cannot read %s: %v", file, err)
		}
		for _, line := range strings.Split(string(data), "\n") {
			for _, m := range cite.FindAllStringSubmatch(line, -1) {
				add(m[2], m[3])
			}
			if f := strings.Fields(line); strings.HasSuffix(file, ".mockups") && len(f) >= 2 && !strings.HasPrefix(f[0], "#") && f[0] != "frame" {
				if m := rfSidecarFrame.FindStringSubmatch(f[1]); m != nil {
					add(m[1], m[2])
				}
			}
		}
	}

	tags := make([]string, 0, len(byTag))
	for t := range byTag {
		tags = append(tags, t)
	}
	sort.Strings(tags)
	rendered, present := 0, 0
	var findings []string
	for _, tag := range tags {
		if tag == "." || tag == ".." || strings.Contains(tag, "..") {
			return refuse("`%s` is not a tag name", tag)
		}
		rel := mockups + "/" + tag
		tracked, ok := capture(vvGitClean(env, "-C", root, "ls-files", "--", rel))
		if !ok {
			return refuse("cannot tell whether %s is committed in %s", rel, args[0])
		}
		if tracked != "" {
			// A tracked directory whose tag the design repo does not hold is the
			// frames' only source (an epic that committed them before tags
			// existed): its committed frames stand, and a cited frame missing
			// from it is the refusal.
			if vvGitClean(env, "-C", clone, "rev-parse", "-q", "--verify", "refs/tags/"+tag).Run() == nil {
				return refuse("tag `%s` names a committed sub-directory %s — its frames would be the committed ones, not the tag's", tag, rel)
			}
			for _, id := range byTag[tag] {
				if !isFile(filepath.Join(root, rel, id+".png")) {
					return refuse("tag `%s` is committed in %s and not in the design repo, and frame %s is missing from it", tag, rel, id)
				}
				present++
			}
			continue
		}
		out := filepath.Join(root, rel)
		var pending []string
		for _, id := range byTag[tag] {
			if p := filepath.Join(out, id+".png"); isFile(p) && !rfBlank(p) {
				present++
			} else {
				pending = append(pending, id)
			}
		}
		if len(pending) == 0 {
			continue
		}
		n, f, rc := rfRenderTag(env, stderr, clone, remote, checkout, tag, out, pending)
		if rc != 0 {
			return rc
		}
		rendered += n
		findings = append(findings, f...)
	}
	for _, f := range findings {
		fmt.Fprint(stdout, sanitizeDisplay("RENDER-FINDING: "+f+"\n"))
	}
	if len(findings) > 0 {
		fmt.Fprintf(stdout, "FRAMES-MISSING: %d reference(s) not rendered\n", len(findings))
		return 1
	}
	fmt.Fprintf(stdout, "FRAMES-RENDERED: %d rendered, %d already present\n", rendered, present)
	return 0
}

// rfRenderTag renders tag's pending ids into out: the tag fetched once when
// absent, its tree exported, each id's page found, then one renderer run.
// It returns the references rendered, the findings, and a non-zero exit when
// it cannot answer (the reason already on stderr).
func rfRenderTag(env Env, stderr io.Writer, clone, remote, checkout, tag, out string, pending []string) (int, []string, int) {
	refuse := func(format string, a ...any) (int, []string, int) {
		fmt.Fprintf(stderr, "render-frames: "+format+"\n", a...)
		return 0, nil, 2
	}
	ref := "refs/tags/" + tag
	if vvGitClean(env, "-C", clone, "rev-parse", "-q", "--verify", ref).Run() != nil {
		if b, err := vvGitClean(env, "-C", clone, "fetch", "-q", remote, ref+":"+ref).CombinedOutput(); err != nil {
			return refuse("tag `%s` is not in %s and cannot be fetched from %s: %s", tag, clone, remote, strings.TrimSpace(string(b)))
		}
	}
	tree, err := os.MkdirTemp("", "render-frames-")
	if err != nil {
		return refuse("cannot make a temp directory: %v", err)
	}
	defer os.RemoveAll(tree)
	if err := rfExport(env, clone, ref, tree); err != nil {
		return refuse("cannot export tag `%s` from %s: %v", tag, clone, err)
	}
	pages, err := filepath.Glob(filepath.Join(tree, "*.html"))
	if err != nil {
		return refuse("cannot list the pages of tag `%s`: %v", tag, err)
	}
	texts := map[string][]byte{}
	for _, p := range pages {
		if texts[p], err = os.ReadFile(p); err != nil {
			return refuse("cannot read %s at tag `%s`: %v", filepath.Base(p), tag, err)
		}
	}

	type job struct{ File, ID string }
	var jobs []job
	var findings []string
	queued := map[string]bool{}
	for _, id := range pending {
		frame := strings.TrimPrefix(id, rfDark)
		if queued[frame] {
			continue
		}
		var in []string
		count := 0
		for _, p := range pages {
			if c := bytes.Count(texts[p], []byte(`id="`+frame+`" data-frame`)); c > 0 {
				count += c
				in = append(in, filepath.Base(p))
			}
		}
		switch {
		case count == 0:
			findings = append(findings, tag+"/"+id+" — no frame `"+frame+"` at the tag")
			continue
		case count > 1:
			findings = append(findings, tag+"/"+id+" — frame `"+frame+"` is not unique at the tag ("+strconv.Itoa(count)+" in "+strings.Join(in, ", ")+")")
			continue
		}
		queued[frame] = true
		jobs = append(jobs, job{in[0], frame})
	}
	if len(jobs) == 0 {
		return 0, findings, 0
	}

	node, ok := lookPath(env, "node")
	if !ok {
		return refuse("node is not on PATH — cannot render")
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return refuse("cannot create %s: %v", out, err)
	}
	if err := os.WriteFile(filepath.Join(out, ".gitignore"), []byte("*\n"), 0o644); err != nil {
		return refuse("cannot write %s/.gitignore: %v", out, err)
	}
	spec, _ := json.Marshal(jobs)
	cmd := exec.Command(node, "-e", rfProgram, checkout, tree, out, string(spec))
	var errb bytes.Buffer
	cmd.Stderr = &errb
	stdout, err := cmd.Output()
	switch rc := exitCode(err); rc {
	case 0:
	case 3:
		return refuse("@playwright/test does not resolve from the regression checkout %s: %s", checkout, strings.TrimSpace(errb.String()))
	default:
		return refuse("the renderer exited %d: %s", rc, strings.TrimSpace(errb.String()))
	}
	phones := map[string]int{}
	sc := bufio.NewScanner(bytes.NewReader(stdout))
	for sc.Scan() {
		if f := strings.Fields(sc.Text()); len(f) == 3 && f[0] == "PHONES" {
			phones[f[1]], _ = strconv.Atoi(f[2])
		}
	}
	rendered := 0
	for _, id := range pending {
		frame := strings.TrimPrefix(id, rfDark)
		if !queued[frame] {
			continue
		}
		need := 1
		if frame != id {
			need = 2
		}
		if phones[frame] < need {
			findings = append(findings, fmt.Sprintf("%s/%s — frame `%s` draws %d phone(s), none at position %d", tag, id, frame, phones[frame], need))
			continue
		}
		if p := filepath.Join(out, id+".png"); rfBlank(p) {
			os.Remove(p)
			findings = append(findings, tag+"/"+id+" — renders blank: its panes never painted")
			continue
		}
		rendered++
	}
	return rendered, findings, 0
}

// rfExport writes `git archive <ref>`'s regular files and directories under
// dir; any other entry, and any name leaving dir, is skipped.
func rfExport(env Env, clone, ref, dir string) error {
	cmd := vvGitClean(env, "-C", clone, "archive", "--format=tar", ref)
	var errb bytes.Buffer
	cmd.Stderr = &errb
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	tr := tar.NewReader(pipe)
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			cmd.Wait()
			return err
		}
		if !filepath.IsLocal(h.Name) {
			continue
		}
		p := filepath.Join(dir, h.Name)
		switch h.Typeflag {
		case tar.TypeDir:
			err = os.MkdirAll(p, 0o755)
		case tar.TypeReg:
			if err = os.MkdirAll(filepath.Dir(p), 0o755); err == nil {
				var f *os.File
				if f, err = os.Create(p); err == nil {
					_, err = io.Copy(f, tr)
					if cerr := f.Close(); err == nil {
						err = cerr
					}
				}
			}
		}
		if err != nil {
			cmd.Wait()
			return err
		}
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("%v: %s", err, strings.TrimSpace(errb.String()))
	}
	return nil
}

// rfProgram is the renderer: argv is the regression checkout, the exported
// tree, the output directory and the jobs as JSON ([{File, ID}]). It prints
// `PHONES <id> <n>` per job, exits 3 when @playwright/test does not resolve.
const rfProgram = `
const [checkout, tree, out, spec] = process.argv.slice(1);
const fs = require('fs'), path = require('path');
let pw;
try { pw = require(require.resolve('@playwright/test', { paths: [checkout] })); }
catch (e) { console.error(String(e)); process.exit(3); }
const jobs = JSON.parse(spec);
(async () => {
  const browser = await pw.chromium.launch();
  try {
    const page = await browser.newPage({ viewport: { width: 1000, height: 1000 }, deviceScaleFactor: 2 });
    const base = path.resolve(tree) + path.sep;
    await page.route('http://design.local/**', r => {
      const f = path.resolve(tree, '.' + decodeURIComponent(new URL(r.request().url()).pathname));
      if (f.startsWith(base) && fs.existsSync(f) && fs.statSync(f).isFile()) return r.fulfill({ path: f });
      return r.fulfill({ status: 404 });
    });
    let loaded = '';
    for (const { File, ID } of jobs) {
      if (loaded !== File) {
        await page.goto('http://design.local/' + encodeURIComponent(File), { waitUntil: 'load', timeout: 60000 });
        loaded = File;
      }
      const frame = page.locator('[id="' + ID + '"][data-frame]');
      await frame.scrollIntoViewIfNeeded();
      await frame.evaluate(f => new Promise(res => {
        const t0 = Date.now();
        const filled = () => f.dataset.migFilled && [...f.querySelectorAll('[data-mig-pane]')].every(p => p.firstElementChild);
        const chk = () => (f.dataset.mig === undefined || filled() || Date.now() - t0 > 8000) ? res() : setTimeout(chk, 100);
        chk();
      }));
      await page.evaluate(() => document.fonts.ready);
      const n = await frame.evaluate(f => {
        const all = [...f.querySelectorAll('div')].filter(d =>
          Math.round(d.getBoundingClientRect().width) === 390 && getComputedStyle(d).borderTopLeftRadius === '28px');
        const outer = all.filter(d => !all.some(o => o !== d && o.contains(d)));
        f.querySelectorAll('[data-render-phone]').forEach(d => d.removeAttribute('data-render-phone'));
        outer.forEach((d, i) => d.setAttribute('data-render-phone', i));
        return outer.length;
      });
      for (let k = 0; k < Math.min(n, 2); k++) {
        const phone = frame.locator('[data-render-phone="' + k + '"]');
        await phone.evaluate(d => d.scrollIntoView({ block: 'center', inline: 'center' }));
        await phone.evaluate(d => Promise.all([...d.querySelectorAll('img')].map(i => i.complete ? 0 : new Promise(r => { i.onload = i.onerror = r; }))));
        const b = await phone.boundingBox();
        await page.screenshot({
          path: path.join(out, (k ? 'dk-' : '') + ID + '.png'),
          clip: { x: Math.round(b.x), y: Math.round(b.y), width: 390, height: 844 },
          animations: 'disabled',
        });
      }
      console.log('PHONES ' + ID + ' ' + n);
    }
  } finally {
    await browser.close();
  }
})().catch(e => { console.error(String(e && e.stack || e)); process.exit(1); });
`

// rfBlank reports whether the PNG at path is a blank phone: below its top
// eighth (the status line), fewer than 0.5% of pixels differ from the
// colour at its centre — a render caught before the frame's panes painted.
// An unreadable or undecodable file counts as blank, so it is re-rendered.
func rfBlank(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return true
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		return true
	}
	b := img.Bounds()
	cr, cg, cb, _ := img.At((b.Min.X+b.Max.X)/2, (b.Min.Y+b.Max.Y)/2).RGBA()
	far := func(a, b uint32) bool { return a > b+0x1000 || b > a+0x1000 }
	diff, total := 0, 0
	for y := b.Min.Y + b.Dy()/8; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, _ := img.At(x, y).RGBA()
			total++
			if far(r, cr) || far(g, cg) || far(bl, cb) {
				diff++
			}
		}
	}
	return diff*200 < total
}
