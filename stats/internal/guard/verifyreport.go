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
// The folded check-verify-report guard (no longer a flow-guard subcommand; `close` and
// `close --check-only` run it in-process, visualverifypass_close.go) — answer whether a flow.visual-verify verifier's
// report is complete, so the parent never accepts a partial one by reading
// its prose (KAN-870: two verifiers ran the tests and the capture, listed
// steps 9-11 under "Not done", and both reports were accepted).
//
// Arguments: <report file> <motions named> [--design <design.md>] [--frames <frame>[,<frame>…]]
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
// an undone step.
//
// The report carries judgements alone: the spec, each frame's composite and
// diff, and the frame count are read from this dispatch's manifest,
// `visual-manifest-<key>.json` beside <report file>, <key> the report's own
// less any `-rest` or `-rest-<n>` segment — never the committed matrix JSON.
//
// With step 10 `done`, every frame id a `- <frame id>: match | differs`
// line names must be in a pair of that manifest; a missing or unreadable
// manifest, or a frame in none of its pairs, is undone. Every frame of the
// manifest's pairs (of --frames' list alone, when given) needs its own
// per-frame line or a mention on the `- frames:` line; each frame with neither
// is undone. The report lists only what differs: the frame line's verdict
// reads `match` for a frame with no difference, and `differs` needs one
// `- <frame id> differs: <element> — mockup <…>, capture <…>, <numbers>` line
// per difference. A `match` frame with a `differs` line, or another verdict
// with none, is undone. A `match` citing `(decision <id>)` on two frames
// whose manifest diff is above vrBlanketCeiling is undone.
//
// For each spec of the manifest's `specs`, the PNGs
// `flow-guard resolve-visual-screenshots <worktree> <spec basename>` prints, the
// worktree three directories above <report file>, each need a
// `- <view>: <absolute PNG path>` line; each gap is undone. A view whose own
// line reads `- <view>: <absolute PNG path> — capture unchanged since <round
// base>` accounts for its capture only in a scoped row — its key has a round
// file — and only for a capture one of the manifest's pairs holds: a scoped
// fix round's verifier does not read a framed view whose capture is unchanged
// (skills/flow/visual-verify-verifier.md step 9), and reads every other. No
// readable manifest, a wire contract naming an endpoint, or a resolver exit other
// than 0 skips the check (group-allowed-pairs-roles self-review: two round-1
// reports named only the prototype-framed captures). With --frames — the
// frames a `-rest` remainder's own report names, which close passes — a
// capture the spec's `<spec>.mockups` maps to no frame of that list is owed no
// line.
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

// checkVerifyReport is the folded check-verify-report guard: the comment above is the contract. It reads a visual-verify verifier's report file
// and answers whether every required step carries an admissible status on the
// report's `- steps:` line, so the parent never accepts a partial report by
// reading its prose.

var (
	vrSteps     = []string{"4", "8", "9", "10", "11", "motion"}
	vrStepsLine = regexp.MustCompile(`^\s*- steps:\s*(.*)$`)
	vrMotion    = regexp.MustCompile(`^\s*- motion:\s*(\d+)/(\d+)\b`)
	vrCitedExit = regexp.MustCompile(`\bexit [1-9][0-9]*\b`)
	// Step 10's own blocking exit (skills/flow/visual-verify-verifier.md step 10).
	vrComposeExit = regexp.MustCompile(`\bcompose-mockup-frames(\.sh)? exit [1-9][0-9]*\b`)
	// A `- wire contract:` line naming an endpoint: the dispatch named no
	// view, so steps 8-10 have nothing to capture (kan-1022's self-review: a
	// verifier wrote them `done` over a capture it never ran).
	vrWireContract = regexp.MustCompile(`^\s*- wire contract:\s*(.*)$`)
	// A backticked HTML or CSS path design.md cites (kan-1017).
	vrMockupSource = regexp.MustCompile("`([^`\\s]+\\.(?:html|htm|css))`")
	// A composed frame's `- <frame id>: match | differs` line, and the decision its match cites.
	vrFrameLine   = regexp.MustCompile(`^\s*- (\S+): ((?:match|differs)\b.*)$`)
	vrDecision    = regexp.MustCompile(`\(decision ([^\s)]+)\)`)
	vrFramesLine  = regexp.MustCompile(`^\s*- frames:\s*(.*)$`)
	vrDiffersLine = regexp.MustCompile(`^\s*- (\S+) differs:\s*\S`)
	// A `-rest` or `-rest-<n>` segment of a report's key: its manifest is the key's without it.
	vrRestSegment = regexp.MustCompile(`-rest(-\d+)?\b`)
)

func checkVerifyReport(args []string, env Env, stdout, stderr io.Writer) int {
	const prog = "check-verify-report"
	const usage = "verify-report check, run in-process by visual-verify-pass close: <report file> <motions named> [--design <design.md>] [--frames <frame>[,<frame>…]]"
	design := ""
	var slice []string // --frames: a remainder's own frames, nil for every frame
	if i := slices.Index(args, "--frames"); i >= 0 {
		if i+1 == len(args) || args[i+1] == "" {
			fmt.Fprintf(stderr, "%s: usage: %s\n", prog, usage)
			Reason(env, "usage")
			return 2
		}
		slice = strings.Split(args[i+1], ",")
		args = slices.Delete(slices.Clone(args), i, i+2)
	}
	if i := slices.Index(args, "--design"); i >= 0 {
		if i+1 == len(args) {
			fmt.Fprintf(stderr, "%s: usage: %s\n", prog, usage)
			Reason(env, "usage")
			return 2
		}
		design = args[i+1]
		args = slices.Delete(slices.Clone(args), i, i+2)
	}
	if len(args) != 2 {
		fmt.Fprintf(stderr, "%s: usage: %s\n", prog, usage)
		Reason(env, "usage")
		return 2
	}
	motions, err := strconv.Atoi(args[1])
	if err != nil || motions < 0 {
		fmt.Fprintf(stderr, "%s: usage: %s — %q is not a count\n", prog, usage, args[1])
		Reason(env, "usage")
		return 2
	}
	data, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Fprintf(stderr, "%s: cannot read %s: %v\n", prog, args[0], err)
		Reason(env, "cannot-read")
		return 2
	}
	status := map[string]string{}
	motionLine := false
	inReport := false
	noView := false
	var frames []string
	framesLine := ""
	var reportLines []string
	verdicts, differs := map[string]string{}, map[string]int{}
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
		if m := vrFramesLine.FindStringSubmatch(line); m != nil {
			framesLine = m[1]
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
			verdicts[m[1]] = strings.TrimSpace(m[2])
		}
		if m := vrDiffersLine.FindStringSubmatch(line); m != nil {
			differs[m[1]]++
		}
		if m := vrWireContract.FindStringSubmatch(line); m != nil && !strings.HasPrefix(strings.TrimSpace(m[1]), "n/a") {
			noView = true
		}
	}
	var undone []string
	// sr12 (kan-1017): `no mockup source` while design.md cites an HTML or
	// CSS source that exists in the worktree, three directories above the
	// report, leaves step 10 undone.
	if design != "" {
		body, err := os.ReadFile(design)
		if err != nil {
			fmt.Fprintf(stderr, "%s: cannot read %s: %v\n", prog, design, err)
			Reason(env, "cannot-read")
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
	// The manifest is read once for every consumer below (KAN-1122): the spec,
	// each frame's diff and the frame list are its, never the report's.
	mpath := vrManifestPath(args[0])
	mbody, merr := os.ReadFile(mpath)
	var manifest vvmManifestOut
	mok := merr == nil && json.Unmarshal(mbody, &manifest) == nil
	if status["10"] == "done" && len(frames) > 0 {
		undone = append(undone, vrComposed(mpath, mbody, merr, frames, verdicts, differs)...)
	}
	if status["10"] == "done" && mok {
		undone = append(undone, vrUnjudged(manifest.Pairs, reportLines, framesLine, slice)...)
	}
	blanket := map[string][]string{}
	for _, p := range manifest.Pairs {
		d := vrDecision.FindStringSubmatch(verdicts[p.Frame])
		if r, _ := strconv.ParseFloat(p.Diff, 64); d != nil && r > vrBlanketCeiling && !slices.Contains(blanket[d[1]], p.Frame) {
			blanket[d[1]] = append(blanket[d[1]], p.Frame)
		}
	}
	for _, id := range slices.Sorted(maps.Keys(blanket)) {
		if fs := blanket[id]; len(fs) > 1 {
			undone = append(undone, fmt.Sprintf("decision %s excuses %d frames above diff=%v (%s) — each needs its own reason", id, len(fs), vrBlanketCeiling, strings.Join(fs, ", ")))
		}
	}
	if !noView {
		key := vrRestSegment.ReplaceAllString(strings.TrimSuffix(strings.TrimPrefix(filepath.Base(args[0]), "verify-report-"), ".md"), "")
		for _, spec := range manifest.Specs {
			undone = append(undone, vrUnaccounted(filepath.Dir(filepath.Dir(filepath.Dir(args[0]))), spec, env, reportLines, vrScopedCaptures(filepath.Dir(args[0]), key, mbody, merr), slice)...)
		}
	}
	if len(undone) > 0 {
		fmt.Fprintf(stdout, "VERIFY-REPORT-INCOMPLETE: %s — %s\n", args[0], strings.Join(undone, "; "))
		Reason(env, "verify-report-incomplete")
		return 1
	}
	fmt.Fprintf(stdout, "VERIFY-REPORT-COMPLETE: %s\n", args[0])
	return 0
}

// vrManifestPath is the manifest of report's dispatch: visual-manifest-<key>.json
// beside it, <key> the report's own less any `-rest` or `-rest-<n>` segment.
func vrManifestPath(report string) string {
	key := strings.TrimSuffix(strings.TrimPrefix(filepath.Base(report), "verify-report-"), ".md")
	return filepath.Join(filepath.Dir(report), "visual-manifest-"+vrRestSegment.ReplaceAllString(key, "")+".json")
}

// vrUnaccounted resolves the spec's captures and names each one no
// `- <view>: <path>` line states. A resolver exit other than 0 is not
// checked: step 9 blocks on it.
// A `capture unchanged since` line accounts only for a capture in scoped, the
// captures a scoped row may leave unread (vrScopedCaptures).
func vrUnaccounted(wt, spec string, env Env, lines, scoped, slice []string) []string {
	var buf strings.Builder
	if resolveVisualScreenshots([]string{wt, filepath.Base(spec)}, env, &buf, io.Discard) != 0 {
		return nil
	}
	report := strings.Join(lines, "\n")
	var gaps []string
	sidecar, _ := os.ReadFile(spec + ".mockups")
	for _, png := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		if slice != nil && !vrOwed(png, sidecar, slice) {
			continue
		}
		m := regexp.MustCompile(`(?m)^\s*- (\S+): ` + regexp.QuoteMeta(png) + `(\s.*)?$`).FindStringSubmatch(report)
		switch {
		case m == nil:
			gaps = append(gaps, filepath.Base(png)+` is on no "- <view>: <path>" line — step 9 states every capture`)
		case strings.HasPrefix(strings.TrimSpace(m[2]), "— capture unchanged since ") && !slices.Contains(scoped, png):
			// Only a scoped row may leave a framed view unread (KAN-1122).
			gaps = append(gaps, m[1]+` reads "capture unchanged" but is not a framed view of a scoped row — read it`)
		}
	}
	return gaps
}

// vrScopedCaptures is every capture key's manifest pairs with a frame, when key
// has a round file — a scoped row's verifier may leave those unread; an
// unscoped row, or a view no frame is composed for, is always read (KAN-1122).
func vrScopedCaptures(sdd, key string, body []byte, rerr error) []string {
	if _, _, _, ok := readRound(sdd, key); !ok {
		return nil
	}
	var m vvmManifestOut
	if rerr != nil || json.Unmarshal(body, &m) != nil {
		return nil
	}
	var caps []string
	for _, p := range m.Pairs {
		caps = append(caps, p.Capture)
	}
	return caps
}

// vrComposed checks each composed frame against this dispatch's manifest —
// never the committed matrix JSON, so a pass that skipped `measure` cannot
// stand on last round's numbers — and its verdict against its `differs:`
// lines. The manifest sits beside the report as visual-manifest-<key>.json,
// <key> the report's own less any `-rest[-<n>]`.
func vrComposed(path string, body []byte, err error, frames []string, verdicts map[string]string, differs map[string]int) []string {
	var m struct {
		Pairs []struct {
			Frame string `json:"frame"`
		} `json:"pairs"`
	}
	if err == nil {
		err = json.Unmarshal(body, &m)
	}
	if err != nil {
		return []string{"composed frames and no readable manifest " + path + " — run step 9's measure"}
	}
	var undone []string
	for _, f := range frames {
		seen := false
		for _, p := range m.Pairs {
			seen = seen || p.Frame == f
		}
		if !seen {
			undone = append(undone, f+" is in no pair of the manifest")
			continue
		}
		match := strings.HasPrefix(verdicts[f], "match")
		switch {
		case match && differs[f] > 0:
			undone = append(undone, fmt.Sprintf("%s reads match and has %d `- %s differs:` lines", f, differs[f], f))
		case !match && differs[f] == 0:
			undone = append(undone, fmt.Sprintf("%s reads neither match nor lists a `- %s differs: <element> — mockup <…>, capture <…>, <numbers>` line", f, f))
		}
	}
	slices.Sort(undone)
	return undone
}

// vrUnjudged names each frame of pairs — of slice alone when it is not nil —
// with no per-frame verdict line that the frames line does not name either:
// the manifest is the frame list a verifier cannot shorten.
func vrUnjudged(pairs []vvmPair, lines []string, framesLine string, slice []string) []string {
	var undone []string
	for _, p := range pairs {
		f := p.Frame
		if slice != nil && !slices.Contains(slice, f) || flPerFrame(f, lines) || flToken(f).MatchString(framesLine) {
			continue
		}
		if l := fmt.Sprintf("%s has no `- %s: match | differs` line", f, f); !slices.Contains(undone, l) {
			undone = append(undone, l)
		}
	}
	return undone
}

// vrBlanketCeiling is the diff above which a frame's match needs a reason of
// its own: one decision cited by two such frames is a blanket.
const vrBlanketCeiling = 0.03

// vrOwed reports whether a --frames report owes png: a capture the sidecar
// maps to a frame of slice.
func vrOwed(png string, sidecar []byte, slice []string) bool {
	for _, l := range mockupsLines(sidecar) {
		if len(cmfMatchCapture(l[0], []string{png})) > 0 && slices.Contains(slice, l[1]) {
			return true
		}
	}
	return false
}
