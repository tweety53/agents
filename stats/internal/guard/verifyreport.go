package guard

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// checkVerifyReport is scripts/check-verify-report.sh: that script's header
// comment is the contract. It reads a visual-verify verifier's report file
// and answers whether every required step carries an admissible status on the
// report's `- steps:` line, so the parent never accepts a partial report by
// reading its prose.
func init() { Registry["check-verify-report"] = checkVerifyReport }

var (
	vrSteps     = []string{"4", "8", "9", "10", "11", "motion"}
	vrStepsLine = regexp.MustCompile(`^\s*- steps:\s*(.*)$`)
	vrMotion    = regexp.MustCompile(`^\s*- motion:\s*(\d+)/(\d+)\b`)
	vrCitedExit = regexp.MustCompile(`\bexit [1-9][0-9]*\b`)
	// Step 10's own blocking exit (skills/flow/visual-verify-verifier.md step 10).
	vrComposeExit = regexp.MustCompile(`\bcompose-mockup-frames(\.sh)? exit [1-9][0-9]*\b`)
	// A bare `- fingerprint: not declared`: step 6 reports exit 2 as `not
	// declared — <the guard's stderr reason>`, and the verifier has that
	// reason only from the prompt (kan-1010's self-review: a verifier wrote
	// it bare for a project declaring the row, its parent's FRESH=0 unseen).
	vrBareFingerprint = regexp.MustCompile(`^\s*- fingerprint:\s*not declared\s*$`)
	// A `## Report` line attributing a departure to the merge base must name
	// its evidence (kan-1015: the kan-466 weight-tab button was attributed
	// pre-existing from the code path alone and re-baselined over the defect).
	vrPreExisting    = regexp.MustCompile(`(?i)(?:^|[^-\w])pre-existing|present at (?:the )?merge base`)
	vrNotPreExisting = regexp.MustCompile(`(?i)\bnot pre-existing`)
	// A `- wire contract:` line naming an endpoint: the dispatch named no
	// view, so steps 8-10 have nothing to capture (kan-1022's self-review: a
	// verifier wrote them `done` over a capture it never ran).
	vrWireContract = regexp.MustCompile(`^\s*- wire contract:\s*(.*)$`)
	// A backticked HTML or CSS path design.md cites (kan-1017).
	vrMockupSource = regexp.MustCompile("`([^`\\s]+\\.(?:html|htm|css))`")
	// A composed frame's `- <frame id>: <composite> diff=` line, and its own
	// `- <frame id> matrix:` line (kan-1016: one matrix line covered nine frames).
	vrPlatformSuffix = regexp.MustCompile(`-(darwin|linux|win32)$`)
	vrFrameLine      = regexp.MustCompile(`^\s*- (\S+): (\S+) diff=`)
	vrMatrixLine     = regexp.MustCompile(`^\s*- (\S+) matrix:`)
	// The capture spec a `## Report` names: every PNG the resolver prints for
	// it needs its own view line (group-allowed-pairs-roles self-review).
	vrSpecLine = regexp.MustCompile(`^\s*- spec:\s*(\S+)`)
	// A composed frame's `bands:`, `seams:` and `over-tolerance:` lines, and
	// any `sweeps:` line (decisions over-tolerance-line, by-hand-check-automated).
	vrBandsLine = regexp.MustCompile(`^\s*- (\S+) bands:`)
	vrSeamsLine = regexp.MustCompile(`^\s*- (\S+) seams:`)
	vrOverLine  = regexp.MustCompile(`^\s*- (\S+) over-tolerance:\s*(.*)$`)
	vrSweeps    = regexp.MustCompile(`\bsweeps:\s*(.*)$`)
	vrOverItem  = regexp.MustCompile(`^#(\d+):\s*(.*)$`)
	// A `-rest` or `-rest-<n>` segment of a report's key: its manifest is the key's without it.
	vrRestSegment = regexp.MustCompile(`-rest(-\d+)?\b`)
	// The n/a columns per row class (decision matrix-kind-opens-with-class).
	vrNAColumns = map[string][]string{
		"text run": {"border", "fill", "radius", "shadow"},
		"control":  {}, "row": {}, "card": {},
		"rule":      {"content", "alignment", "border", "radius", "shadow"},
		"divider":   {"content", "alignment", "border", "radius", "shadow"},
		"fill band": {"content", "alignment", "border", "radius", "shadow"},
	}
)

func checkVerifyReport(args []string, env Env, stdout, stderr io.Writer) int {
	const prog = "check-verify-report"
	const usage = "check-verify-report.sh <report file> <motions named> [--design <design.md>]"
	design := ""
	if i := slices.Index(args, "--design"); i >= 0 {
		if i+1 == len(args) {
			fmt.Fprintf(stderr, "%s: usage: %s\n", prog, usage)
			return 2
		}
		design = args[i+1]
		args = slices.Delete(slices.Clone(args), i, i+2)
	}
	if len(args) != 2 {
		fmt.Fprintf(stderr, "%s: usage: %s\n", prog, usage)
		return 2
	}
	motions, err := strconv.Atoi(args[1])
	if err != nil || motions < 0 {
		fmt.Fprintf(stderr, "%s: usage: %s — %q is not a count\n", prog, usage, args[1])
		return 2
	}
	data, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Fprintf(stderr, "%s: cannot read %s: %v\n", prog, args[0], err)
		return 2
	}
	status := map[string]string{}
	motionLine, bareFingerprint := false, false
	inReport, unevidenced := false, ""
	noView := false
	var frames, framed []string
	matrix := map[string]bool{}
	spec := ""
	var reportLines []string
	bands, seams, over := map[string]bool{}, map[string]bool{}, map[string]string{}
	var badSweeps []string
	for _, line := range strings.Split(string(data), "\n") {
		// Every line is read from `## Report` to the next `## ` heading alone:
		// the `## Verdict` block above it is the handback, not the record.
		if strings.HasPrefix(line, "## ") {
			inReport = strings.TrimSpace(line) == "## Report"
		}
		if !inReport {
			continue
		}
		reportLines = append(reportLines, line)
		if m := vrSpecLine.FindStringSubmatch(line); m != nil && spec == "" {
			spec = m[1]
		}
		if unevidenced == "" && vrPreExisting.MatchString(line) && !vrNotPreExisting.MatchString(line) &&
			!strings.Contains(line, "merge-base baseline: ") && !strings.Contains(line, "merge-base capture: ") {
			unevidenced = strings.TrimSpace(line)
			if r := []rune(unevidenced); len(r) > 120 {
				unevidenced = string(r[:120])
			}
		}
		if m := vrStepsLine.FindStringSubmatch(line); m != nil {
			for _, item := range strings.Split(m[1], " | ") {
				id, rest, _ := strings.Cut(strings.TrimSpace(item), " ")
				status[id] = strings.TrimSpace(rest)
			}
		}
		if m := vrMotion.FindStringSubmatch(line); m != nil && m[1] == m[2] && m[2] == strconv.Itoa(motions) {
			motionLine = true
		}
		if m := vrFrameLine.FindStringSubmatch(line); m != nil {
			frames = append(frames, m[1])
			framed = append(framed, strings.TrimSuffix(filepath.Base(m[2]), ".png"))
		}
		if m := vrMatrixLine.FindStringSubmatch(line); m != nil {
			matrix[m[1]] = true
		}
		if m := vrBandsLine.FindStringSubmatch(line); m != nil {
			bands[m[1]] = true
		}
		if m := vrSeamsLine.FindStringSubmatch(line); m != nil {
			seams[m[1]] = true
		}
		if m := vrOverLine.FindStringSubmatch(line); m != nil {
			over[m[1]] = m[2]
		}
		if m := vrSweeps.FindStringSubmatch(line); m != nil {
			for _, item := range strings.Split(m[1], " | ") {
				_, st, _ := strings.Cut(strings.TrimSpace(item), " ")
				st = strings.TrimSpace(st)
				why, isNA := strings.CutPrefix(st, "n/a — ")
				if !(st == "done" || strings.HasPrefix(st, "done ") || isNA && strings.TrimSpace(why) != "") {
					badSweeps = append(badSweeps, "sweep item `"+strings.TrimSpace(item)+"` reads neither `done` nor `n/a — <why>`")
				}
			}
		}
		bareFingerprint = bareFingerprint || vrBareFingerprint.MatchString(line)
		if m := vrWireContract.FindStringSubmatch(line); m != nil && !strings.HasPrefix(strings.TrimSpace(m[1]), "n/a") {
			noView = true
		}
	}
	var undone []string
	if bareFingerprint {
		undone = append(undone, "fingerprint: `not declared` with no reason — state step 6's result as the prompt names it")
	}
	// sr12 (kan-1017): `no mockup source` while design.md cites an HTML or
	// CSS source that exists in the worktree, three directories above the
	// report, leaves step 10 undone.
	if design != "" {
		body, err := os.ReadFile(design)
		if err != nil {
			fmt.Fprintf(stderr, "%s: cannot read %s: %v\n", prog, design, err)
			return 2
		}
		wt := filepath.Dir(filepath.Dir(filepath.Dir(args[0])))
		vv, _ := os.ReadFile(filepath.Join(filepath.Dir(design), "visual-verification.md"))
		if strings.Contains(strings.Join(reportLines, "\n"), "no mockup source") || strings.Contains(string(vv), "no mockup source") {
			for _, m := range vrMockupSource.FindAllStringSubmatch(string(body), -1) {
				if isFile(filepath.Join(wt, m[1])) {
					status["10"] = "`no mockup source` while design.md cites " + m[1]
					break
				}
			}
		}
		// An approved prototype is the frame where no mockup is (kan-1038): step 10
		// n/a beside a cited HTML file that exists is undone, unless a wire
		// contract names no view to compose.
		if strings.HasPrefix(status["10"], "n/a") && !noView {
			for _, m := range vrMockupSource.FindAllStringSubmatch(string(body), -1) {
				if !strings.HasSuffix(m[1], ".css") && isFile(filepath.Join(wt, m[1])) {
					status["10"] = "n/a while design.md cites the prototype " + m[1] + " — render it as each capture's frame (visual-verify-verifier.md step 10)"
					break
				}
			}
		}
	}
	var designBody []byte
	if design != "" {
		designBody, _ = os.ReadFile(design)
	}
	// Every step-10 defect is named in one run, never only the last.
	step10 := badSweeps
	if design != "" {
		step10 = append(step10, vrMatrixCells(filepath.Join(filepath.Dir(design), "visual-verification.md"))...)
	}
	if unevidenced != "" {
		step10 = append(step10, "pre-existing attribution with no merge-base evidence: "+unevidenced)
	}
	if len(step10) > 0 {
		status["10"] = strings.Join(step10, "; ")
	}
	for _, id := range vrSteps {
		s, ok := status[id]
		naAllowed := id == "4" || id == "10" || (id == "motion" && motions == 0) || (noView && (id == "8" || id == "9"))
		reason, isNA := strings.CutPrefix(s, "n/a — ")
		cause, isBlocked := strings.CutPrefix(s, "blocked — ")
		switch {
		case !ok:
			undone = append(undone, id+" absent")
		case s == "done" && id == "motion" && motions > 0 && !motionLine:
			undone = append(undone, fmt.Sprintf("motion: no `- motion: %d/%d` line", motions, motions))
		case s == "done":
		// Step 10 is n/a only when no mockups are declared, so its reason must say so,
		// or when a wire contract names no view to compose.
		case isNA && naAllowed && strings.TrimSpace(reason) != "" && (id != "10" || noView || strings.Contains(reason, "mockups")):
		// Step 10 measures the captures step 8 wrote, so it blocks only when step 8 did or
		// on compose-mockup-frames' own exit (KAN-1000: two verifiers skipped step 10 with
		// every capture on disk, and departures surfaced one round at a time).
		case isBlocked && id == "10" && !strings.HasPrefix(status["8"], "blocked — ") && !vrComposeExit.MatchString(cause):
			undone = append(undone, strings.TrimSpace(id+" "+s)+" (step 8 captured)")
		case isBlocked && vrCitedExit.MatchString(cause):
		default:
			undone = append(undone, strings.TrimSpace(id+" "+s))
		}
	}
	if status["10"] == "done" {
		for _, f := range frames {
			if !matrix[f] {
				undone = append(undone, f+" has no matrix: line — step 10 is done only with every frame's matrix")
			}
			if !bands[f] {
				undone = append(undone, f+" has no bands: line")
			}
			if !seams[f] {
				undone = append(undone, f+" has no seams: line")
			}
		}
		if len(frames) > 0 {
			undone = append(undone, vrOverTolerance(args[0], frames, over, string(designBody))...)
		}
	}
	if spec != "" && !noView {
		undone = append(undone, vrUnaccounted(filepath.Dir(filepath.Dir(filepath.Dir(args[0]))), spec, env, framed, reportLines)...)
	}
	if len(undone) > 0 {
		fmt.Fprintf(stdout, "VERIFY-REPORT-INCOMPLETE: %s — %s\n", args[0], strings.Join(undone, "; "))
		return 1
	}
	fmt.Fprintf(stdout, "VERIFY-REPORT-COMPLETE: %s\n", args[0])
	return 0
}

// vrUnaccounted resolves the spec's captures and names each one no
// `- <view>: <path>` line states, and each stated view no frame is composed
// for that carries no per-view sweeps line. A capture is framed when its
// basename, less `.png` and any Playwright `-<platform>` suffix, is the stem
// of a frame line's composite path — a frame id names the mockup, not the
// view. A resolver exit other than 0 is not checked: step 9 blocks on it.
func vrUnaccounted(wt, spec string, env Env, framed, lines []string) []string {
	var buf strings.Builder
	if resolveVisualScreenshots([]string{wt, filepath.Base(spec)}, env, &buf, io.Discard) != 0 {
		return nil
	}
	report := strings.Join(lines, "\n")
	var gaps []string
	for _, png := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		m := regexp.MustCompile(`(?m)^\s*- (\S+): ` + regexp.QuoteMeta(png) + `(\s|$)`).FindStringSubmatch(report)
		switch {
		case m == nil:
			gaps = append(gaps, filepath.Base(png)+` is on no "- <view>: <path>" line — step 9 states every capture`)
		case !slices.Contains(framed, vrPlatformSuffix.ReplaceAllString(strings.TrimSuffix(filepath.Base(png), ".png"), "")) && !strings.Contains(report, "per view, no frame composed: "+m[1]+" sweeps:"):
			gaps = append(gaps, m[1]+` has no frame and no "per view, no frame composed: `+m[1]+` sweeps:" line`)
		}
	}
	return gaps
}

type vrOverRow struct {
	Row    int    `json:"row"`
	Status string `json:"status"`
	Keys   string `json:"keys"`
}

// vrOverTolerance reconciles each composed frame's `over-tolerance:` line with
// the `over` rows this dispatch's manifest carries for it — never the
// committed matrix JSON, so a pass that skipped `measure` cannot stand on last
// round's numbers. The manifest sits beside the report as
// visual-manifest-<key>.json, <key> the report's own less any `-rest[-<n>]`.
func vrOverTolerance(report string, frames []string, lines map[string]string, design string) []string {
	key := strings.TrimSuffix(strings.TrimPrefix(filepath.Base(report), "verify-report-"), ".md")
	path := filepath.Join(filepath.Dir(report), "visual-manifest-"+vrRestSegment.ReplaceAllString(key, "")+".json")
	var m struct {
		Pairs []struct {
			Frame string      `json:"frame"`
			Over  []vrOverRow `json:"over"`
		} `json:"pairs"`
	}
	body, err := os.ReadFile(path)
	if err == nil {
		err = json.Unmarshal(body, &m)
	}
	if err != nil {
		return []string{"composed frames and no readable manifest " + path + " — run step 9's measure"}
	}
	var undone []string
	for _, f := range frames {
		rows, seen := map[int]vrOverRow{}, false
		for _, p := range m.Pairs {
			if p.Frame == f {
				seen = true
				for _, r := range p.Over {
					rows[r.Row] = r
				}
			}
		}
		line, ok := lines[f]
		switch {
		case !seen:
			undone = append(undone, f+" is in no pair of the manifest")
			continue
		case !ok:
			undone = append(undone, f+" has no over-tolerance: line")
			continue
		}
		count, items, _ := strings.Cut(line, " — ")
		if strings.TrimSpace(count) != strconv.Itoa(len(rows)) {
			undone = append(undone, fmt.Sprintf("%s over-tolerance: %s, the manifest's over count is %d", f, strings.TrimSpace(count), len(rows)))
		}
		named := map[int]bool{}
		for _, item := range strings.Split(items, "; ") {
			mm := vrOverItem.FindStringSubmatch(strings.TrimSpace(item))
			if mm == nil {
				continue
			}
			n, _ := strconv.Atoi(mm[1])
			named[n] = true
			if r, ok := rows[n]; ok && !vrCauseAdmitted(mm[2], r, design) {
				undone = append(undone, fmt.Sprintf("%s over-tolerance #%d: cause %q is off step 10's closed list", f, n, strings.TrimSpace(mm[2])))
			}
		}
		for n := range rows {
			if !named[n] {
				undone = append(undone, fmt.Sprintf("%s over-tolerance: row #%d unnamed", f, n))
			}
		}
	}
	slices.Sort(undone)
	return undone
}

// vrCauseAdmitted answers whether cause is on step 10's closed list for row.
func vrCauseAdmitted(cause string, row vrOverRow, design string) bool {
	for _, c := range []string{"departure", "annotation", "transcription", "text width"} {
		if strings.HasPrefix(cause, c) {
			return true
		}
	}
	if id, ok := strings.CutPrefix(cause, "decision "); ok {
		id = strings.Fields(id + " ")[0]
		return id != "" && regexp.MustCompile(`(?m)^\*\*ID:\*\*\s*`+regexp.QuoteMeta(id)+`\s*$`).MatchString(design)
	}
	// Its merge-base evidence is already required of every pre-existing attribution above.
	if strings.HasPrefix(cause, "present at the merge base") {
		return true
	}
	if strings.HasPrefix(cause, "mockup source") {
		return slices.Contains(strings.Split(row.Keys, ", "), "unresolved")
	}
	return false
}

// vrMatrixCells checks every matrix table in visual-verification.md — a table
// whose header's first cell opens `element` — cell by cell: an empty cell, or
// an n/a (or `—`) cell in a column that applies to its row's class, not
// reading `no mockup source`, is refused by its line. The class is the
// `kind` cell's opening; an unknown class admits n/a only as `no mockup source`.
func vrMatrixCells(path string) []string {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	cells := func(l string) []string {
		parts := strings.Split(strings.Trim(strings.TrimSpace(l), "|"), "|")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		return parts
	}
	var bad, header []string
	for i, l := range strings.Split(string(body), "\n") {
		if !strings.HasPrefix(strings.TrimSpace(l), "|") {
			header = nil
			continue
		}
		c := cells(l)
		switch {
		case header == nil:
			if strings.HasPrefix(c[0], "element") {
				header = c
			} else {
				header = []string{}
			}
			continue
		case len(header) == 0 || strings.Trim(strings.Join(c, ""), "-: ") == "":
			continue
		}
		kind := ""
		if k := slices.Index(header, "kind"); k >= 0 && k < len(c) {
			kind = c[k]
		}
		var na []string
		known := false
		for class, cols := range vrNAColumns {
			if strings.HasPrefix(kind, class) {
				na, known = cols, true
			}
		}
		for j, col := range header {
			v := ""
			if j < len(c) {
				v = c[j]
			}
			// `— / 14px` names one side missing: a measurement, not an n/a.
			isNA := strings.HasPrefix(v, "n/a") || (strings.HasPrefix(v, "—") && !strings.Contains(v, "/"))
			switch {
			case v == "":
				bad = append(bad, fmt.Sprintf("visual-verification.md:%d: %s.%s is empty", i+1, c[0], col))
			case isNA && !strings.Contains(v, "no mockup source") && !(known && slices.Contains(na, col)):
				bad = append(bad, fmt.Sprintf("visual-verification.md:%d: %s.%s reads %q — the column applies to `%s`", i+1, c[0], col, v, kind))
			}
		}
	}
	return bad
}
