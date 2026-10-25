// Package prompttuning loads and renders one prompt-tuning case folder,
// docs/prompt-tuning/<case-id>/ in the agents repository. The case shape --
// case.json's fields and the folder's files -- is §2 of
// spectre/changes/kan-1032-prompt-tuning-angle/design.md, written by the
// self-review's flow-prompt-tuning angle (skills/flow-self-review/SKILL.md).
//
// Render is deterministic so check-prompt-tuning-cases can compare a
// folder's index.html with it byte for byte.
package prompttuning

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
)

// RequiredFiles is every file a case folder carries.
var RequiredFiles = []string{"case.json", "finding.md", "bad.diff", "good.diff", "task.md", "index.html"}

// Source names the run, finding and fix a case was taken from.
type Source struct {
	Project    string `json:"project"`
	Change     string `json:"change"`
	Jira       string `json:"jira"`
	Repo       string `json:"repo"`
	FixSHA     string `json:"fix_sha"`
	FindingRef string `json:"finding_ref"`
	Round      *int   `json:"round"`
	Slot       string `json:"slot"`
	Severity   string `json:"severity"`
}

// Case is case.json.
type Case struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	ReviewerKind string `json:"reviewer_kind,omitempty"`
	Scope        string `json:"scope"`
	Source       Source `json:"source"`
	Target       string `json:"target"`
	TargetSHA    string `json:"target_sha"`
	DefectClass  string `json:"defect_class"`
	WhyPicked    string `json:"why_picked"`
	Before       string `json:"before"`
	After        string `json:"after"`
}

// Files holds the folder's text files besides case.json and index.html.
type Files struct {
	Finding  string
	BadDiff  string
	GoodDiff string
	Task     string
}

// Load reads a case folder's case.json and repro files; an absent file is
// an error naming it.
func Load(dir string) (Case, Files, error) {
	read := func(name string) (string, error) {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return "", fmt.Errorf("%s: %w", name, err)
		}
		return string(b), nil
	}
	var c Case
	var f Files
	raw, err := read("case.json")
	if err != nil {
		return c, f, err
	}
	if err := json.Unmarshal([]byte(raw), &c); err != nil {
		return c, f, fmt.Errorf("case.json: %w", err)
	}
	for _, p := range []struct {
		name string
		dst  *string
	}{{"finding.md", &f.Finding}, {"bad.diff", &f.BadDiff}, {"good.diff", &f.GoodDiff}, {"task.md", &f.Task}} {
		if *p.dst, err = read(p.name); err != nil {
			return c, f, err
		}
	}
	return c, f, nil
}

// sectionHeadings is the page's section order, design.md §2's.
var sectionHeadings = []string{"Type", "Defect class", "Finding", "Fix", "Why picked", "Prompt update"}

var page = template.Must(template.New("case").Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.C.ID}}</title>
<style>
body{font-family:system-ui,sans-serif;margin:0 auto;max-width:1100px;padding:16px;line-height:1.45}
pre{background:#f4f4f4;padding:8px;overflow-x:auto;white-space:pre-wrap;word-break:break-word}
.ab{display:grid;grid-template-columns:1fr 1fr;gap:12px}
@media (max-width:700px){.ab{grid-template-columns:1fr}}
</style>
</head>
<body>
<h1>{{.C.ID}}</h1>
<h2>{{index .H 0}}</h2>
<p>{{.C.Type}}{{if .C.ReviewerKind}} — {{.C.ReviewerKind}}{{end}} · scope {{.C.Scope}}</p>
<h2>{{index .H 1}}</h2>
<p>{{.C.DefectClass}}</p>
<h2>{{index .H 2}}</h2>
<pre>{{.F.Finding}}</pre>
<pre><code>{{.F.BadDiff}}</code></pre>
<h2>{{index .H 3}}</h2>
<pre><code>{{.F.GoodDiff}}</code></pre>
<h2>{{index .H 4}}</h2>
<pre>{{.C.WhyPicked}}</pre>
<h2>{{index .H 5}}</h2>
<p>Target: <code>{{.C.Target}}</code> at <code>{{.C.TargetSHA}}</code></p>
<div class="ab">
<div><h3>Before</h3><pre>{{.C.Before}}</pre></div>
<div><h3>After</h3><pre>{{.C.After}}</pre></div>
</div>
</body>
</html>
`))

// Render returns the case's index.html.
func Render(c Case, f Files) ([]byte, error) {
	var buf bytes.Buffer
	err := page.Execute(&buf, struct {
		C Case
		F Files
		H []string
	}{c, f, sectionHeadings})
	return buf.Bytes(), err
}
