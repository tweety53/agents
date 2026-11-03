package guard

import (
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// The contract below moved here from the header of the deleted
// scripts/check-verify-report.sh shim (KAN-1058).
//
// flow-guard check-verify-report — answer whether a flow.visual-verify verifier's
// report is complete, so the parent never accepts a partial one by reading
// its prose (KAN-870: two verifiers ran the tests and the capture, listed
// steps 9-11 under "Not done", and both reports were accepted).
//
// Usage: flow-guard check-verify-report <report file> <motions named> [--design <design.md>]
//
// Given --design, the report, or the `visual-verification.md` beside
// <design.md>, reading `no mockup source` while <design.md> cites a
// backticked `.html`, `.htm` or `.css` path that exists relative to the
// worktree (three directories above <report file>) leaves step 10 undone
// (kan-1017: every verifier left 7 of 11 matrix columns n/a). Step 10
// reported n/a while <design.md> cites a backticked `.html` or `.htm` path
// that exists is likewise undone, unless a wire contract names no view: an
// approved prototype is rendered as each capture's frame (kan-1038). An
// unreadable <design.md> cannot be answered (exit 2).
//
// <report file> is the relay contract's
// <abs-worktree>/.superpowers/sdd/verify-report-<key>.md
// (skills/flow/visual-verify.md); <motions named> is the count of motions the
// dispatch prompt named, 0 for `motions: none`.
//
// Every line below is read from the report's `## Report` heading to the next
// `## ` heading alone: the `## Verdict` block above it is the handback, never
// the record (decision verdict-only-handback). The verifier runs no `verify`
// (decision verifier-runs-no-verify), so step 7 is no required step and a
// `7` item on the `- steps:` line is ignored.
//
// The report's `- steps:` line carries one `<step> <status>` item per required
// step, ` | `-separated, for steps 4, 8, 9, 10, 11 and `motion`
// (skills/flow/visual-verify-verifier.md). A status is admissible when it is
//   done
//   n/a — <reason>       steps 4 and 10 only, and `motion` when <motions named> is 0;
//                        steps 8, 9 and 10 too when a `- wire contract:` line
//                        names an endpoint — the dispatch named no view, so
//                        nothing is captured (kan-1022's self-review)
//   blocked — <reason>   the reason citing a non-zero `exit <n>`; step 10 only
//                        when step 8 is blocked too or the reason cites
//                        `compose-mockup-frames exit <n>` (KAN-1000)
// and `motion done` with <motions named> above 0 also needs a
// `- motion: <n>/<n>` line, <n> that count — a strip recorded beyond the
// named motions goes on a sub-bullet, never into <n>. Every other status — absent,
// `not done`, an n/a on a required step, a block citing no failing exit — is
// an undone step. A `- fingerprint: not declared` line with no ` — <reason>`
// is undone too: step 6 reports its exit 2 with the guard's reason, which the
// prompt carries (kan-1010's self-review).
//
// A `## Report` line attributing a departure as `pre-existing` or `present at
// the merge base` that names neither `merge-base baseline: <sha>:<path>` nor
// `merge-base capture: <path>` leaves step 10 undone (kan-1015: a re-baseline
// over a defect attributed from the code path alone).
// Every merge-base baseline: <sha>:<path> must resolve with git cat-file -e in
// the report's worktree (<report file>'s third parent directory); one that
// does not leaves step 10 undone (kan-1057 self-review, F6).
//
// With step 10 `done`, every frame id a `- <frame id>: <composite> diff=`
// line names needs its own `- <frame id> matrix:` line; each that has none is
// undone as `<frame id> has no matrix: line — step 10 is done only with every
// frame's matrix` (kan-1016: one matrix line covered nine frames).
// Each such frame needs its `- <frame id> bands:` and `- <frame id> seams:`
// lines too, and its `- <frame id> over-tolerance: <n> — #<row>: <cause>; …`
// line (decisions over-tolerance-line, by-hand-check-automated): <n> must
// equal the `over` count the frame's pairs carry in this dispatch's manifest,
// `visual-manifest-<key>.json` beside <report file>, <key> the report's own
// less any `-rest` or `-rest-<n>` segment — never the committed matrix JSON —
// and every row on that list must be named with one cause from step 10's
// closed list: `departure`, `annotation`, `transcription`, `text width`,
// `decision <ID>` with `**ID:** <ID>` in <design.md>, `present at the merge
// base` beside `merge-base baseline: ` or `merge-base capture: ` evidence,
// or `mockup source` on a row whose `keys` name `unresolved`. A missing or
// unreadable manifest, a frame in none of its pairs, a missing line, a wrong
// count, an unnamed row or a cause off the list is undone.
// Each item whose manifest row carries figures quotes one as <key> a=<a> b=<b>;
// an item quoting none of its row's figures is undone (kan-1057 self-review, F20).
//
// Every `sweeps:` line's ` | `-separated items each read `<sweep> done…` or
// `<sweep> n/a — <why>`; any other leaves step 10 undone.
//
// Given --design, every matrix table in the `visual-verification.md` beside
// <design.md> — a table whose header's first cell opens `element` — is read
// cell by cell (decision matrix-kind-opens-with-class). The row's class is its
// `kind` cell's opening; a cell reading `n/a…` or `—` stands only in a column
// that does not apply to the class — `border`, `fill`, `radius`, `shadow` for
// a `text run`; none for a `control`, `row` or `card`; `content`,
// `alignment`, `border`, `radius`, `shadow` for a `rule`, `divider` or
// `fill band` — or when it reads `no mockup source`; an unknown class admits
// only `no mockup source`. An empty cell, or any other n/a cell (kan-938: a
// control's `border` left `n/a — measure failed`), leaves step 10 undone,
// naming `visual-verification.md:<line>`.
//
// Given a `## Report` line `- spec: <path>`, the PNGs
// `flow-guard resolve-visual-screenshots <worktree> <spec basename>` prints, the
// worktree three directories above <report file>, each need a
// `- <view>: <absolute PNG path>` line, and each such view that no
// `- <view>: <composite> diff=` line names needs a
// `per view, no frame composed: <view> sweeps:` line; each gap is undone. No
// `- spec:` line, a wire contract naming an endpoint, or a resolver exit other
// than 0 skips the check (group-allowed-pairs-roles self-review: two round-1
// reports named only the prototype-framed captures).
//
// Prints ONE verdict line to stdout:
//   VERIFY-REPORT-COMPLETE: <report file>
//   VERIFY-REPORT-INCOMPLETE: <report file> — <each undone step and its status, `; `-separated>
//
// Exit 0 complete, 1 incomplete, 2 cannot answer (a usage error, a count
// that is not a non-negative integer, a report file it
// cannot read).
//
// Go, in flow-guard (stats/internal/guard/verifyreport.go).

// checkVerifyReport is `flow-guard check-verify-report`: the comment above is the contract. It reads a visual-verify verifier's report file
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
	// A `merge-base baseline: <sha>:<path>` citation, which must resolve in the
	// report's worktree (kan-1057 self-review, F6: a mangled one stood as evidence).
	vrBaselineCite = regexp.MustCompile(`merge-base baseline: (\S+)`)
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
	const usage = "flow-guard check-verify-report <report file> <motions named> [--design <design.md>]"
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
	var badSweeps, cites []string
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
		for _, m := range vrBaselineCite.FindAllStringSubmatch(line, -1) {
			if c := strings.TrimRight(m[1], ";,).`"); !slices.Contains(cites, c) {
				cites = append(cites, c)
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
			framed = append(framed, vrPlatformSuffix.ReplaceAllString(strings.TrimSuffix(filepath.Base(m[2]), ".png"), ""))
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
	wt := filepath.Dir(filepath.Dir(filepath.Dir(args[0])))
	for _, c := range cites {
		sha, path, _ := strings.Cut(c, ":")
		if _, ok := capture(envGit(env)("-C", wt, "cat-file", "-e", c)); sha == "" || path == "" || !ok {
			step10 = append(step10, "merge-base baseline `"+c+"` does not resolve in "+wt)
		}
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
	Row     int                          `json:"row"`
	Status  string                       `json:"status"`
	Keys    string                       `json:"keys"`
	Figures map[string]map[string]string `json:"figures"`
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
			// The item quotes its own row's figures (kan-1057 self-review, F20: a
			// cause read off a neighbouring row stood).
			if r, ok := rows[n]; ok && len(r.Figures) > 0 {
				var quotes []string
				quoted := false
				for _, k := range slices.Sorted(maps.Keys(r.Figures)) {
					q := k + " a=" + r.Figures[k]["a"] + " b=" + r.Figures[k]["b"]
					quotes = append(quotes, q)
					quoted = quoted || strings.Contains(mm[2], q)
				}
				if !quoted {
					undone = append(undone, fmt.Sprintf("%s over-tolerance #%d: quotes none of its row's figures (%s)", f, n, strings.Join(quotes, "; ")))
				}
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
