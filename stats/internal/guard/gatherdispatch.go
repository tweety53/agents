package guard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// The port of scripts/gather-dispatch-context.sh, whose header is canonical
// for the contract: arguments, VALIDATION, LEAF VALIDATION, the content
// directory, the incidents and hazards sections, and SKIP-WHEN-UNCHANGED.
// The helpers it sourced -- lib/resolve-file.sh, lib/project-section.sh and
// the since-deleted lib/within-root.sh and lib/lexical-normalize.sh -- are
// ported below as far as this guard uses them.
//
// Two bash dependencies are gone: the hash comes from crypto/sha256, so the
// "no hash tool available" rebuild cannot happen, and flow's JSON is parsed
// here, so jq is not required for the incidents and hazards sections. The
// reasoning for each branch, moved here from the bash body it replaced
// (0747740), sits beside the code it explains. The bash sourced its helpers
// from scripts/lib/ because it shipped through the skills/*/scripts/ symlink
// farm — the criterion each library's own header states for when a guard may
// source a sibling rather than carry its own copy; the port carries Go
// copies of those helpers instead, and no longer sources any of them.
func init() { Registry["gather-dispatch-context"] = gatherDispatchContext }

const gdcUsage = "usage: gather-dispatch-context.sh <worktree> <change-root> <name> <principles-path> <output-path> [<task-ids> [<canonical-worktree>]]\n"

var gdcTaskIDs = regexp.MustCompile(`^[0-9]+(,[0-9]+)*$`)

func gatherDispatchContext(args []string, env Env, _, stderr io.Writer) int {
	arg := func(i int) string {
		if i < len(args) {
			return args[i]
		}
		return ""
	}
	worktree, changeRoot, name, principles, output := arg(0), arg(1), arg(2), arg(3), arg(4)
	taskIDs, canonical, shape := arg(5), arg(6), arg(7)

	if worktree == "" || changeRoot == "" || name == "" || principles == "" || output == "" {
		fmt.Fprint(stderr, gdcUsage)
		return 2
	}
	if taskIDs != "" && !gdcTaskIDs.MatchString(taskIDs) {
		fmt.Fprintf(stderr, "gather-dispatch-context: task ids '%s' must be a comma-separated list of integers\n", taskIDs)
		return 2
	}
	switch shape {
	case "", "all", "cross-repo", "single-repo":
	default:
		fmt.Fprintf(stderr, "gather-dispatch-context: shape '%s' must be one of all, cross-repo, single-repo\n", shape)
		return 2
	}
	if !plainChangeName(name) {
		fmt.Fprintf(stderr, "gather-dispatch-context: change name '%s' is not a plain change name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'\n", name)
		return 2
	}

	// --- worktree: must resolve to an existing, non-symlinked directory ---
	worktreeReal, mismatch := validatePath(env, worktree)
	if worktreeReal == "" || !isDir(worktreeReal) {
		fmt.Fprintf(stderr, "gather-dispatch-context: worktree '%s' does not resolve to an existing directory\n", worktree)
		return 2
	}
	if mismatch {
		fmt.Fprintf(stderr, "gather-dispatch-context: worktree '%s' resolves through a symlink somewhere between the repository root and the leaf\n", worktree)
		return 2
	}
	// --- change-root: must resolve to an existing, non-symlinked directory
	// inside the (already-validated) worktree ---
	changeRootReal, mismatch := validatePath(env, changeRoot)
	if changeRootReal == "" || !isDir(changeRootReal) {
		fmt.Fprintf(stderr, "gather-dispatch-context: change-root '%s' does not resolve to an existing directory\n", changeRoot)
		return 2
	}
	if mismatch {
		fmt.Fprintf(stderr, "gather-dispatch-context: change-root '%s' resolves through a symlink somewhere between the repository root and the leaf\n", changeRoot)
		return 2
	}
	if !withinRoot(changeRootReal, worktreeReal) {
		fmt.Fprintf(stderr, "gather-dispatch-context: change-root '%s' resolves outside the worktree '%s'\n", changeRoot, worktree)
		return 2
	}
	// --- principles-path: normalized and resolved the same way, but never
	// required to exist, never checked for containment under the worktree (it
	// may be a global install path outside it entirely), and NEVER refused for
	// a symlink divergence (F1). setup.sh global installs every skill as a
	// symlink (~/.claude/skills/flow -> the repository's skills/flow/), so
	// [PRINCIPLES_PATH] — resolved by the CALLER from its own installed skill
	// directory (skills/flow/review-panel.md's own rule), never derived here —
	// ALWAYS diverges lexically from its real path in that install shape. The
	// lexical/real divergence refusal exists to stop a change directory's own
	// (repo-tracked, pull-request-editable, attacker-influenced) content
	// escaping <change-root>; principles-path is not that — it is a trusted
	// argument the dispatching skill resolves for itself, so it is simply
	// resolved (following every symlink, leaf and ancestor alike) and read.
	// Containment and divergence refusal are UNCHANGED for <worktree> and
	// <change-root> above, and for the leaf content sources below. ---
	principlesReal, _ := validatePath(env, principles)

	// The bundle itself.
	b := &gdcBundle{}

	// The content directory (KAN-393): where this change's plan leaves
	// actually live. The <change-root> itself whenever it carries a tasks.md —
	// the single-repo case, byte-for-byte unchanged — and otherwise, when it
	// carries a satellite's link.md, the canonical change directory
	// changePlanDir resolves through the seventh argument or, only when that
	// argument is empty, through <worktree>/<spec-root>/peers. When even that
	// fails, the change IS a satellite (its link.md carries ## Part of — the
	// same definition check-unfinished-work.sh applies) and the three plan
	// leaves below are reported as skips with the distinct unresolved label;
	// a change with neither file is not a satellite, and its leaves keep the
	// ordinary absent-leaf handling. changePlanRef's answer is the
	// <peer>:<change-id> ref of the link the successful resolution already
	// validated, feeding the section label. Both are Go ports of
	// lib/change-plan.sh's change_plan_dir and change_plan_ref (changeplan.go);
	// the bash sourced that library as the owner of link.md's grammar.
	contentDir, suffix, satelliteLink := changeRootReal, "", ""
	if !isFile(changeRootReal + "/tasks.md") {
		// A relative canonical stays relative, as in bash: changePlanDir
		// probes it against the process cwd, and every leaf found under it
		// is then refused, its resolved path never within the relative root.
		if dir, rc := changePlanDir(env, io.Discard, worktreeReal, name, canonical); rc == 0 && dir != "" {
			contentDir = dir
			if ref, ok := changePlanRef(io.Discard, worktreeReal, name); ok {
				suffix = " (canonical " + ref + ")"
			}
		} else if link := changeRootReal + "/link.md"; isFile(link) && gdcHasPartOf(link) {
			contentDir, satelliteLink = "", link
		}
	}

	if contentDir != "" {
		for _, leaf := range []string{"proposal.md", "design.md", "tasks.md"} {
			b.addFixedSource(contentDir+"/"+leaf, leaf+suffix, contentDir)
		}
	} else {
		// A satellite whose canonical plan could not be reached: the bundle is
		// advisory and never gates a run, so this is a skip, not a refusal —
		// with its own label, so "the plan is elsewhere and unreachable" never
		// reads like a plain change's legitimately-absent design.md (KAN-393).
		for _, leaf := range []string{"proposal.md", "design.md", "tasks.md"} {
			b.skipped = append(b.skipped, leaf+" (satellite plan unresolved — link.md at "+satelliteLink+")")
		}
	}

	if taskIDs != "" {
		last := len(b.found) - 1
		if last < 0 || b.found[last].label != "tasks.md"+suffix {
			fmt.Fprintln(stderr, "gather-dispatch-context: task ids given but tasks.md is absent or refused")
			return 2
		}
		body, ok := scopeTasks(b.found[last].path, taskIDs, stderr)
		if !ok {
			return 2
		}
		b.found[last] = gdcSource{label: "tasks.md" + suffix + " (scoped to task(s) " + taskIDs + ")", text: body + "\n"}
	}

	if principlesReal != "" && isFile(principlesReal) {
		b.found = append(b.found, gdcSource{label: principles, path: principlesReal})
	} else {
		b.skipped = append(b.skipped, principles+" (absent)")
	}

	b.addProjectCommands(worktreeReal)
	b.addIncidents(env, worktreeReal)
	// The change's shape, computed by the caller from its resolved worktree
	// set. Empty means the caller does not know it; the hazards section then
	// filters to the always-on rows (the literal "all"), never to a guess.
	if shape == "" {
		shape = "all"
	}
	b.addHazards(env, worktreeReal, shape)

	body := strings.TrimRight(b.render(), "\n") // `BODY="$(render_body)"`
	// The hashing primitive SKIP-WHEN-UNCHANGED compares the body against:
	// sha256Hex (sha256.go), the repository's one helper, shared with
	// check-cleanup-complete (F1, this change's own review panel) — once a
	// second caller existed, rules/build-the-simplest-thing.mdc's "no
	// abstraction until a second caller exists" called for extracting the
	// shared helper rather than leaving two copies free to drift apart,
	// exactly the failure within_root suffered as two bash copies before
	// its extraction.
	newHash := sha256Hex(body)
	outputAbs := output
	if !strings.HasPrefix(outputAbs, "/") {
		outputAbs = env.Dir + "/" + outputAbs
	}
	hashPath := outputAbs + ".hash"

	reason := ""
	if !isFile(hashPath) || !isFile(outputAbs) {
		reason = "no cached bundle"
	} else if old, _ := os.ReadFile(hashPath); strings.TrimRight(string(old), "\n") != newHash {
		reason = "inputs changed"
	}
	if reason == "" {
		fmt.Fprintf(stderr, "gather-dispatch-context: bundle unchanged — reusing %s\n", output)
		return 0
	}

	head, ok := gitOut(env, "-C", worktreeReal, "rev-parse", "--short", "HEAD")
	if !ok {
		head = "unknown"
	}
	// The body's own trailing newline(s) are stripped above, as the bash's
	// `$(render_body)` stripped them (F2, this change's own review panel) —
	// the hash comparison is unaffected, since newHash is derived from that
	// same stripped body on every call, but the file this writes must still
	// end with a newline the way the pre-kan-288 stdout-redirected script
	// always did. The final "\n" restores exactly that: one trailing newline,
	// appended after the body's own content rather than baked into what gets
	// hashed.
	file := "# Dispatch context bundle for " + name + "\n" +
		"generated: " + time.Now().UTC().Format("2006-01-02T15:04:05Z") + "\n" +
		"head: " + head + "\n\n" + body + "\n"
	if err := os.MkdirAll(gdcDirname(outputAbs), 0o777); err != nil {
		fmt.Fprintf(stderr, "gather-dispatch-context: %v\n", err)
		return 1
	}
	if err := os.WriteFile(outputAbs, []byte(file), 0o666); err != nil {
		fmt.Fprintf(stderr, "gather-dispatch-context: %v\n", err)
		return 1
	}
	if err := os.WriteFile(hashPath, []byte(newHash), 0o666); err != nil {
		fmt.Fprintf(stderr, "gather-dispatch-context: %v\n", err)
		return 1
	}
	fmt.Fprintf(stderr, "gather-dispatch-context: bundle rebuilt — %s\n", reason)
	return 0
}

// gdcSource is one "## <label>" section: a file's bytes (path) or text
// already rendered, printed verbatim.
type gdcSource struct{ label, path, text string }

// gdcRefusal carries a reason per refused label — "outside_change" (or
// "outside_worktree" for project.md) for a genuine withinRoot containment
// breach, "unresolvable" for a resolveFile failure (a symlink loop past its
// own 40-hop cap, or a component that could not be walked, e.g. a race or a
// permissions failure) (F36, pass 7 of this change's own review panel).
// Before this, both cases were folded into one bucket and every one was
// reported with the same "(resolves outside the change directory)" text —
// true for the first case, and simply wrong for the second: a source
// resolveFile could not even walk was never compared against the change root
// at all, so nothing about it says "outside". Both dispositions still omit
// the source from the bundle and still let the run exit 0 — only the printed
// diagnostic differs.
type gdcRefusal struct{ label, reason string }

// gdcBundle's every skipped entry must already be the complete "skipped: "
// line's payload — either self-describing (e.g. "incidents (none)") or a bare
// identifier with " (absent)" already appended. render only echoes each entry
// verbatim; it does no further formatting.
type gdcBundle struct {
	found   []gdcSource
	skipped []string // each the complete payload of its "skipped: " line
	refused []gdcRefusal
}

// addFixedSource is add_fixed_source: absent when not a file, refused when
// its path cannot be walked or resolves outside root, found otherwise.
//
// Unlike <worktree>/<change-root>/<principles-path>, these content sources are
// never passed as arguments, so validatePath's exit-2-on-mismatch contract
// does not apply to them: a change legitimately carries no design.md, and
// "this leaf is a symlink" must not abort a whole dispatching stage. Instead
// this mirrors the retired self-review gather's (kan-526) check_boundary() +
// is_refused() pair: a regular-file test first (a missing target, including a
// dangling symlink, is a plain absence — skipped, not refused); then
// resolveFile (leaf and every ancestor symlink) and a withinRoot boundary
// check against root — the CONTENT DIRECTORY (KAN-393): the change-root itself
// for a local plan, the canonical change directory for a satellite, so a leaf
// can never resolve outside the change whose content it is. A leaf that
// resolves inside that root — including a symlink to another file inside it —
// is read normally. A leaf that resolveFile cannot even walk ("unresolvable"),
// or that resolves OUTSIDE it ("outside_change"), is refused per-source:
// omitted from the bundle, counted separately from "skipped", and the run
// still exits 0 — the spec's "a missing bundle never stops a run"
// requirement, and the same disposition the retired self-review gather's
// (kan-526) header documented for exactly this shape of problem (a per-source
// trust-boundary violation is not a malformed invocation).
func (b *gdcBundle) addFixedSource(path, label, root string) {
	if !isFile(path) {
		b.skipped = append(b.skipped, label+" (absent)")
		return
	}
	resolved, ok := resolveFile(path)
	switch {
	case !ok:
		b.refused = append(b.refused, gdcRefusal{label, "unresolvable"})
	case withinRoot(resolved, root):
		b.found = append(b.found, gdcSource{label: label, path: resolved})
	default:
		b.refused = append(b.refused, gdcRefusal{label, "outside_change"})
	}
}

// addProjectCommands carries the ## lint, ## test and ## run sections of
// <worktree>/.flow/project.md, so a dispatched subagent already carries this
// project's lint/test/run commands and never needs to open project.md itself
// (CLAUDE.md's lint-fix-priority rule sends every subagent there). Extracted,
// not the whole 20+ KB file, per design.md's
// scoped-dispatch-bundle-not-full-project-md decision. Appended after the
// principles file section, as the last "found" entry. projectSection is the
// port of lib/project-section.sh's project_section, the one shared definition
// check-model-keys.sh still sources.
func (b *gdcBundle) addProjectCommands(worktree string) {
	file := worktree + "/.flow/project.md"
	if !isFile(file) {
		b.skipped = append(b.skipped, "project commands (absent)")
		return
	}
	resolved, ok := resolveFile(file)
	if !ok {
		b.refused = append(b.refused, gdcRefusal{"project commands", "unresolvable"})
		return
	}
	if !withinRoot(resolved, worktree) {
		b.refused = append(b.refused, gdcRefusal{"project commands", "outside_worktree"})
		return
	}
	var text strings.Builder
	for _, key := range []string{"lint", "test", "run"} {
		section := projectSection(resolved, key)
		if strings.Trim(section, gdcSpace+"\n") != "" {
			text.WriteString("### " + key + "\n\n" + section + "\n\n")
		}
	}
	if text.Len() == 0 {
		b.skipped = append(b.skipped, "project commands (absent)")
		return
	}
	b.found = append(b.found, gdcSource{label: "project commands", text: text.String()})
}

// gdcIncident is one row of `flow record incidents`. minutesLost keeps its
// JSON literal, as jq's tostring prints it; change is null when unset.
type gdcIncident struct {
	OccurredAt  string      `json:"occurredAt"`
	Guard       string      `json:"guard"`
	Symptom     string      `json:"symptom"`
	Recovery    string      `json:"recovery"`
	MinutesLost json.Number `json:"minutesLost"`
	Change      *string     `json:"change"`
}

// addIncidents is this project's guard-incident log, from
// `flow record incidents -C <worktree>`, so every dispatch already carries the
// hazards prior runs hit instead of relying on someone re-reading memory
// before dispatching (KAN-451). Three outcomes: `flow` absent or the call
// failing skips as "incidents (flow unavailable)"; a `[]` result skips as
// "incidents (none)"; anything else renders the "## incidents" section, a
// six-column table. The section is part of the hashed body, so a newly
// recorded incident forces a rebuild.
func (b *gdcBundle) addIncidents(env Env, worktree string) {
	var rows []gdcIncident
	if !gdcFlowJSON(env, &rows, "record", "incidents", "-C", worktree) {
		b.skipped = append(b.skipped, "incidents (flow unavailable)")
		return
	}
	if len(rows) == 0 {
		b.skipped = append(b.skipped, "incidents (none)")
		return
	}
	pipe := strings.NewReplacer("|", `\|`)
	var text strings.Builder
	text.WriteString("| when | guard | symptom | recovery | minutes lost | change |\n|------|-------|---------|----------|--------------|--------|\n")
	for _, r := range rows {
		when := r.OccurredAt
		if len(when) > 10 {
			when = when[:10]
		}
		change := "—"
		if r.Change != nil {
			change = *r.Change
		}
		text.WriteString("| " + strings.Join([]string{when, r.Guard, pipe.Replace(r.Symptom), pipe.Replace(r.Recovery), r.MinutesLost.String(), change}, " | ") + " |\n")
	}
	b.found = append(b.found, gdcSource{label: "incidents", text: strings.TrimRight(text.String(), "\n") + "\n"})
}

// addHazards is this project's proactive warnings, from
// `flow hazards -C <worktree> -shape <shape>` (KAN-452) — the incidents
// section's three-outcome shape: `flow` absent or the call failing skips as
// "hazards (flow unavailable)"; a `[]` result skips as "hazards (none)";
// anything else renders the "## hazards" section, one bullet per row. The
// section is part of the hashed body, so a newly recorded hazard forces a
// rebuild. An empty shape arrives here as the literal "all" — fail-open to
// the always-on rows, per the header's HAZARDS paragraph.
func (b *gdcBundle) addHazards(env Env, worktree, shape string) {
	var rows []struct{ Name, Applies, Body string }
	if !gdcFlowJSON(env, &rows, "hazards", "-C", worktree, "-shape", shape) {
		b.skipped = append(b.skipped, "hazards (flow unavailable)")
		return
	}
	if len(rows) == 0 {
		b.skipped = append(b.skipped, "hazards (none)")
		return
	}
	var text strings.Builder
	for _, r := range rows {
		text.WriteString("- **" + r.Name + " (" + r.Applies + "):** " + r.Body + "\n")
	}
	b.found = append(b.found, gdcSource{label: "hazards", text: strings.TrimRight(text.String(), "\n") + "\n"})
}

// gdcFlowJSON runs `flow <args>` from env's PATH and decodes its stdout into
// v: false when flow is absent, fails, or prints no JSON array. JSON null is
// jq's length 0, an empty result.
func gdcFlowJSON(env Env, v any, args ...string) bool {
	flow, ok := lookPath(env, "flow")
	if !ok {
		return false
	}
	cmd := exec.Command(flow, args...)
	cmd.Dir = env.Dir
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return json.Unmarshal(out, v) == nil
}

// render is render_body: the census, the refused and skipped lines, then
// every found section — everything printed after the header's
// `generated:`/`head:` lines. Rendered once and hashed, so the header's own
// per-call metadata (the timestamp, the HEAD sha) never enters the comparison
// — see SKIP-WHEN-UNCHANGED in the header and design.md's
// header-excluded-from-hash decision.
func (b *gdcBundle) render() string {
	var s strings.Builder
	fmt.Fprintf(&s, "found: %d source(s); skipped: %d source(s); refused: %d source(s)\n", len(b.found), len(b.skipped), len(b.refused))
	for _, r := range b.refused {
		switch r.reason {
		case "outside_change":
			s.WriteString("refused: " + r.label + " (resolves outside the change directory)\n")
		case "outside_worktree":
			s.WriteString("refused: " + r.label + " (resolves outside the worktree)\n")
		default:
			s.WriteString("refused: " + r.label + " (could not be resolved: a symlink loop or other failure walking its path)\n")
		}
	}
	for _, l := range b.skipped {
		s.WriteString("skipped: " + l + "\n")
	}
	s.WriteString("\n")
	for _, src := range b.found {
		s.WriteString("## " + src.label + "\n\n")
		if src.path != "" {
			content, _ := os.ReadFile(src.path)
			s.Write(content)
		} else {
			s.WriteString(src.text)
		}
		s.WriteString("\n")
	}
	return s.String()
}

// scopeTasks is scope_tasks: the plan's header plus the named tasks' blocks
// in document order, trailing newlines stripped as `$(...)` strips them.
// Each named id with no task line is reported on stderr, and ok is false.
// Task-line, body-boundary and fence grammar are plan-dispatch-bundles.py's
// (see its docstring), so the ids that script printed are the ids this finds.
func scopeTasks(path, ids string, stderr io.Writer) (string, bool) {
	content, _ := os.ReadFile(path)
	want := strings.Split(ids, ",")
	wanted := map[string]bool{}
	for _, id := range want {
		wanted[id] = true
	}
	seen := map[string]bool{}
	fence, header, cur := false, true, ""
	var out []string
	for _, line := range lines(content) {
		if gdcFence.MatchString(line) {
			fence = !fence
			if header || cur != "" {
				out = append(out, line)
			}
			continue
		}
		if m := gdcTaskLine.FindStringSubmatch(line); !fence && m != nil {
			header = false
			cur = ""
			if wanted[m[1]] {
				cur = m[1]
				seen[cur] = true
			}
		} else if !fence && !header && gdcSectionHeading.MatchString(line) {
			cur = ""
		}
		if header || cur != "" {
			out = append(out, line)
		}
	}
	ok := true
	for _, id := range want {
		if !seen[id] {
			fmt.Fprintf(stderr, "gather-dispatch-context: task %s not found in tasks.md\n", id)
			ok = false
		}
	}
	return strings.TrimRight(strings.Join(out, "\n"), "\n"), ok
}

var (
	gdcFence          = regexp.MustCompile("^ {0,3}(```|~~~)")
	gdcTaskLine       = regexp.MustCompile(`^- \[[ x]\] ([0-9]+)\. `)
	gdcSectionHeading = regexp.MustCompile(`^##(#)?([ \t]|$)`)
)

// gdcSpace is awk's and tr's [:space:] less the newline lines never carry.
const gdcSpace = " \t\v\f\r"

// projectSection is lib/project-section.sh's project_section: the body of
// "## <key>" up to the next "## " heading, a leading UTF-8 BOM dropped,
// blank lines trimmed at both ends, trailing newlines stripped.
func projectSection(file, key string) string {
	content, _ := os.ReadFile(file)
	content = bytes.TrimPrefix(content, []byte("\xef\xbb\xbf"))
	var body []string
	grab := false
	for _, line := range lines(content) {
		if line == "## "+key {
			grab = true
			continue
		}
		if strings.HasPrefix(line, "## ") && grab {
			break
		}
		if grab {
			body = append(body, line)
		}
	}
	for len(body) > 0 && strings.Trim(body[0], gdcSpace) == "" {
		body = body[1:]
	}
	for len(body) > 0 && strings.Trim(body[len(body)-1], gdcSpace) == "" {
		body = body[:len(body)-1]
	}
	return strings.TrimRight(strings.Join(body, "\n"), "\n")
}

// gdcHasPartOf is `grep -qxF '## Part of'`.
func gdcHasPartOf(file string) bool {
	content, err := os.ReadFile(file)
	if err != nil {
		return false
	}
	for _, line := range lines(content) {
		if line == "## Part of" {
			return true
		}
	}
	return false
}

// validatePath is the header's VALIDATION mechanism: the lexical form of
// path (made absolute against env.Dir), its fully resolved form ("" when
// resolution fails — a plain absence), and whether the two differ, i.e. a
// symlink sits somewhere between an existing ancestor and the leaf. The one
// mechanism the header documents: normalize, resolve, compare.
func validatePath(env Env, path string) (real string, mismatch bool) {
	if !strings.HasPrefix(path, "/") {
		path = env.Dir + "/" + path
	}
	lexical := lexicallyCollapse(path)
	// resolveFile is given the already lexically-normalized form, not the raw
	// argument (F2): the bash resolve_file's dirname/basename split mishandled
	// a bare "." (the result gained a spurious trailing "/.") and a
	// trailing-slash path (a spurious trailing "/") — neither of which
	// lexicallyCollapse ever produces, so either form was wrongly compared as
	// a "divergence" against a clean lexical path even with no symlink
	// anywhere in the chain. The lexical form is always absolute with no "."
	// segment and no trailing slash, so resolveFile never sees either shape
	// here. resolveFile itself also normalizes the trailing slash on its own
	// input (F9, lib/resolve-file.sh's own "NORMALIZES ITS OWN INPUT"
	// comment), so this pre-normalization is belt-and-braces, not the only
	// thing standing between this function and the bug.
	real, ok := resolveFile(lexical)
	if !ok {
		return "", false
	}
	return real, real != lexical
}

// lexicallyCollapse is the port of the deleted lib/lexical-normalize.sh's
// lexically_collapse: drop
// empty and "." segments, pop on "..", never past the root — pure string
// manipulation with no filesystem access. validatePath makes a relative path
// absolute against the process's own directory first; only that "make it
// absolute" half stays local to each caller, the collapse being shared (F34,
// this change's own review panel), so it cannot be fooled by a symlink.
func lexicallyCollapse(abs string) string {
	var stack []string
	for _, part := range strings.Split(abs, "/") {
		switch part {
		case "", ".":
		case "..":
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		default:
			stack = append(stack, part)
		}
	}
	return "/" + strings.Join(stack, "/")
}

// withinRoot is the port of the deleted lib/within-root.sh's within_root: a
// path BOUNDARY test, not a string prefix test (a bare prefix would wrongly
// accept "/foo/bar-evil" against "/foo/bar").
func withinRoot(p, root string) bool {
	return p == root || strings.HasPrefix(p, root+"/")
}

// resolveFile is lib/resolve-file.sh's resolve_file:
// follow the leaf's symlinks (at most 40 hops), then resolve its directory
// physically, as `cd -P`. false where the bash function returns non-zero.
func resolveFile(p string) (string, bool) {
	for p != "/" && strings.HasSuffix(p, "/") {
		p = strings.TrimSuffix(p, "/")
	}
	for hops := 1; ; hops++ {
		fi, err := os.Lstat(p)
		if err != nil || fi.Mode()&os.ModeSymlink == 0 {
			break
		}
		if hops > 40 {
			return "", false
		}
		target, err := os.Readlink(p)
		if err != nil {
			return "", false
		}
		switch dir := gdcDirname(p); {
		case strings.HasPrefix(target, "/"):
			p = target
		case dir == "/":
			p = "/" + target
		default:
			p = dir + "/" + target
		}
	}
	base := p[strings.LastIndex(p, "/")+1:]
	if base == "." || base == ".." {
		return gdcPhysicalDir(p)
	}
	dir, ok := gdcPhysicalDir(gdcDirname(p))
	if !ok {
		return "", false
	}
	if dir == "/" {
		return "/" + base, true
	}
	return dir + "/" + base, true
}

// gdcDirname is dirname(1).
func gdcDirname(p string) string {
	if p != "" && strings.Trim(p, "/") == "" {
		return "/"
	}
	p = strings.TrimRight(p, "/")
	i := strings.LastIndex(p, "/")
	if i < 0 {
		return "."
	}
	d := strings.TrimRight(p[:i+1], "/")
	if d == "" {
		return "/"
	}
	return d
}

// gdcPhysicalDir is `cd -P -- dir && pwd -P`: always absolute, a relative
// dir taken from the process's physical cwd, as the shell's.
func gdcPhysicalDir(dir string) (string, bool) {
	if !strings.HasPrefix(dir, "/") {
		cwd, err := os.Getwd()
		if err != nil {
			return "", false
		}
		dir = cwd + "/" + dir
	}
	real, err := filepath.EvalSymlinks(dir)
	if err != nil || !isDir(real) {
		return "", false
	}
	return real, true
}
