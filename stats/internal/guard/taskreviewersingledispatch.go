package guard

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
)

// checkTaskReviewerSingleDispatch is
// scripts/check-task-reviewer-single-dispatch.sh: that script's header
// comment is the contract -- exit 0 every gated-per-task reviewer row of the
// session token is shape-clean, retry-clean and bundled per implementer
// group, 1 violations (each on stderr), 2 cannot answer. The dispatch read is
// Env.Dispatches when set, else `flow record dispatches`; the decisions read
// is always `flow record decisions` on env's PATH. The reasoning for each
// branch, moved here from the bash body it replaced (c5379c0a), sits beside
// the code it explains.
//
// The bash needed jq and refused at exit 2 without it; the rows are decoded
// with encoding/json here, so there is no tool left to find missing.
func init() {
	Registry["check-task-reviewer-single-dispatch"] = checkTaskReviewerSingleDispatch
}

const trsdPrefix = "check-task-reviewer-single-dispatch: "

var (
	// trsdFamilyRE is the jq filter's `test("^task-[0-9]+(\\+[0-9]+)*-reviewer")`:
	// the gated-per-task-reviewer family, matched against the key as stored.
	// jq's `^` anchors the string start only, as RE2's does.
	trsdFamilyRE = regexp.MustCompile(`^task-[0-9]+(\+[0-9]+)*-reviewer`)
	// trsdKeyRE is the bash's CANONICAL_KEY_RE; RE2 and bash's ERE agree on
	// it for the same reasons pfdKeyRE's comment gives.
	trsdKeyRE    = regexp.MustCompile(`^task-[0-9]+(\+[0-9]+)*-reviewer(-fix-[0-9]+)?(-retry)?$`)
	trsdBundleRE = regexp.MustCompile(`^bundle ([0-9]+): (.*)$`)
	trsdGroupRE  = regexp.MustCompile(`^group ([0-9]+): (.*)$`)
)

// trsdBase is one reviewer base key (the key with a trailing -retry
// stripped), its `+`-joined task ids unsplit, and its counts, kept in
// first-seen order as the bash's parallel arrays were.
type trsdBase struct {
	key, ids      string
	orig, retries int
}

func checkTaskReviewerSingleDispatch(args []string, env Env, stdout, stderr io.Writer) int {
	arg := func(i int) string {
		if i < len(args) {
			return args[i]
		}
		return ""
	}
	worktreeArg, name, token := arg(0), arg(1), arg(2)
	if worktreeArg == "" || !isDir(pcAbs(env, worktreeArg)) {
		shown := worktreeArg
		if shown == "" {
			shown = "<missing>"
		}
		fmt.Fprintf(stderr, "%snot a directory: %s\n", trsdPrefix, shown)
		return 2
	}
	if name == "" {
		fmt.Fprintf(stderr, "%susage: check-task-reviewer-single-dispatch.sh <worktree> <change-name> <session-token>\n", trsdPrefix)
		return 2
	}
	if token == "" {
		fmt.Fprintf(stderr, "%ssession token is required and must be this run's own literal token\n", trsdPrefix)
		return 2
	}

	// CONTAINMENT, identical to panelfixsingledispatch.go's: plainChangeName
	// (cleanupcomplete.go), whose comment is canonical for the reasoning.
	if !plainChangeName(name) {
		fmt.Fprintf(stderr, "%schange name '%s' is not a plain change name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'\n", trsdPrefix, name)
		return 2
	}

	// Canonicalise before the worktree is ever passed to `flow` as `-C`, and
	// refuse rather than proceed if it vanished -- panelfixsingledispatch.go
	// carries the same check and the reasoning for each refusal.
	worktree, err := filepath.EvalSymlinks(pcAbs(env, worktreeArg))
	if err != nil || strings.HasPrefix(worktreeArg, "-") || syscall.Access(pcAbs(env, worktreeArg), 1) != nil {
		fmt.Fprintf(stderr, "%sworktree vanished: \n", trsdPrefix)
		return 2
	}

	// The two siblings are exec'd from beside the shim, as the bash exec'd
	// them from $SCRIPT_DIR.
	scriptDir, ok := guardSelfDir(env, stderr, trsdPrefix, "check-task-reviewer-single-dispatch")
	if !ok {
		return 2
	}
	tasksMD := worktree + "/spectre/changes/" + name + "/tasks.md"
	if fi, err := os.Stat(tasksMD); err != nil || !fi.Mode().IsRegular() {
		fmt.Fprintf(stderr, "%sno tasks.md at %s -- cannot answer\n", trsdPrefix, tasksMD)
		return 2
	}

	// A store the call could not reach is "cannot answer", never "no rows":
	// an outage that read as zero rows would pass every run it was blind to.
	rows, err := pfdRead(env, env.Dispatches, "dispatches", name, worktree)
	if err != nil {
		fmt.Fprintf(stderr, "%sflow record dispatches failed for '%s' -- cannot answer\n", trsdPrefix, name)
		return 2
	}

	// Each reviewer row of THIS token whose key is in the gated-per-task-
	// reviewer family.
	keys, ok := trsdReviewerKeys(rows, token)
	if !ok {
		fmt.Fprintf(stderr, "%sdispatch rows were not readable JSON -- cannot answer\n", trsdPrefix)
		return 2
	}

	// Group membership: task id -> group number, from this run's own plan --
	// never re-derived from the raw markdown by this guard.
	bundleLines, ok := trsdSibling(env, scriptDir, "plan-dispatch-bundles.sh", tasksMD, stderr)
	if !ok {
		return 2
	}
	groupLines, ok := trsdSibling(env, scriptDir, "plan-dispatch-groups.sh", tasksMD, stderr)
	if !ok {
		return 2
	}
	groupOf := trsdGroups(bundleLines, groupLines)

	// The decision's class -- a failed read, output that is not JSON (the
	// bash's `jq empty` gate), or rows carrying no class read as `big`, as
	// the bash did. JSON that is not an array of decision rows is a
	// cannot-answer: the bash blocked on it (jq under set -e, exit 5, or a
	// top-level object's values read as rows), and `big` tolerates more
	// bundles than small/regular, so reading it as `big` would pass what
	// the bash blocked.
	class := "big"
	if out, _, rc := pcFlow(env, false, "record", "decisions", "-change", name, "-C", worktree); rc == 0 {
		c, ok := trsdNewestClass(out)
		if !ok {
			fmt.Fprintf(stderr, "%sdecision output was JSON but not an array of decision rows -- cannot answer\n", trsdPrefix)
			return 2
		}
		if c != "" {
			class = c
		}
	}

	var violations []string
	var bases []*trsdBase
	index := map[string]*trsdBase{}
	for _, key := range keys {
		if !trsdKeyRE.MatchString(key) {
			violations = append(violations, fmt.Sprintf("out-of-shape gated-reviewer key '%s' -- the canonical shape is task-<n[+n+n...]>-reviewer[-fix-<k>][-retry]", key))
			continue
		}
		base, retry := strings.CutSuffix(key, "-retry")
		b := index[base]
		if b == nil {
			// The bash's ${base#task-} then ${ids_part%-reviewer*}: the
			// shortest suffix from the last -reviewer.
			ids := strings.TrimPrefix(base, "task-")
			ids = ids[:strings.LastIndex(ids, "-reviewer")]
			b = &trsdBase{key: base, ids: ids}
			index[base] = b
			bases = append(bases, b)
		}
		if retry {
			b.retries++
		} else {
			b.orig++
		}
	}

	for _, b := range bases {
		total := b.orig + b.retries
		switch {
		case b.retries > 0 && b.orig == 0:
			violations = append(violations, fmt.Sprintf("bundle %s carries %d -retry dispatch(es) and no original -- a retry is never a bundle's only dispatch", b.key, b.retries))
		case b.orig > 1:
			violations = append(violations, fmt.Sprintf("bundle %s carries %d reviewer dispatches -- the bundle goes to ONE dispatch, never one per task, slot or finding", b.key, b.orig))
		case total > 2 || (total == 2 && b.retries != 1):
			violations = append(violations, fmt.Sprintf("bundle %s carries %d reviewer dispatches -- at most the original plus the handshake's one -retry", b.key, total))
		}
	}

	// ---- bundling-itself checks (never one dispatch per gate-fired task) ----
	// Only ORIGINAL bundles (a fix round's own re-review key, `-fix-<k>`, is
	// its own bundle by construction and is exempt: implement.md
	// re-dispatches one bundle per group's fixed tasks, and two groups fixing
	// in the same round are two legitimately separate re-review bundles,
	// never a KAN-527 shape).
	var origs []*trsdBase
	for _, b := range bases {
		if b.orig >= 1 && !strings.Contains(b.key, "-fix-") {
			origs = append(origs, b)
		}
	}

	// Same group, two separate original bundles: exactly the KAN-527 shape,
	// regardless of class.
	seen := map[string]string{} // group -> first original bundle naming it
	for _, b := range origs {
		for _, tid := range strings.Split(b.ids, "+") {
			g, ok := groupOf[tid]
			if !ok {
				continue
			}
			if first, ok := seen[g]; !ok {
				seen[g] = b.key
			} else if first != b.key {
				violations = append(violations, fmt.Sprintf("tasks in group %s are split across reviewer bundles '%s' and '%s' -- every gate-fired task of one implementer group goes out in ONE reviewer dispatch", g, first, b.key))
			}
		}
	}

	// On `micro`/`small`/`regular`, implement.md requires every gate-fired task of
	// the whole run to join ONE bundle at the last boundary -- more than one
	// original bundle is itself a violation on those classes.
	if (class == "micro" || class == "small" || class == "regular") && len(origs) > 1 {
		names := ""
		for _, b := range origs {
			names += " " + b.key
		}
		violations = append(violations, fmt.Sprintf("class '%s' carries %d original reviewer bundles (%s) -- every gate-fired task of the run joins ONE bundle at the last boundary on micro/small/regular", class, len(origs), names))
	}

	if len(violations) > 0 {
		for _, v := range violations {
			fmt.Fprintf(stderr, "%s%s\n", trsdPrefix, v)
		}
		return 1
	}
	fmt.Fprintln(stdout, "TASK-REVIEWER-SINGLE-DISPATCH-OK: every gated-per-task reviewer dispatch of this run is bundled by implementer group (bounds and retries included)")
	return 0
}

// trsdReviewerKeys is `.[] | select((.sessionToken // "") == $tok and
// (.role // "") == "reviewer" and ((.key // "") | test(<family>)))` then
// each row's `.key // ""` as jq -r printed it. A null row selects nothing;
// any other non-object row, a selected row whose key is neither a string nor
// null/false (jq's `test` failing), or a top-level value that is not an array
// is jq failing -- pfdPanelFixKeys's comment covers the top-level object.
func trsdReviewerKeys(raw []byte, token string) ([]string, bool) {
	vals, ok := pfdStream(raw)
	if !ok {
		return nil, false
	}
	var keys []string
	for _, v := range vals {
		arr, ok := v.([]any)
		if !ok {
			return nil, false
		}
		for _, e := range arr {
			if e == nil {
				continue
			}
			row, ok := e.(map[string]any)
			if !ok {
				return nil, false
			}
			if row["sessionToken"] != token || row["role"] != "reviewer" {
				continue
			}
			var key string
			switch k := row["key"].(type) {
			case nil:
			case bool:
				if k {
					return nil, false
				}
			case string:
				key = k
			default:
				return nil, false
			}
			if trsdFamilyRE.MatchString(key) {
				keys = append(keys, pfdRawText(key))
			}
		}
	}
	return keys, true
}

// trsdSibling execs one plan-dispatch sibling on tasks.md with stdout and
// stderr combined, as the bash's `2>&1` captured them, and returns that
// output as `$(...)` did, run under LC_ALL=C as the bash exported it. On failure it prints the bash's cannot-answer line
// and then the output (an empty line when there was none, as `echo ""`
// printed); a sibling that could not be started prints its error instead.
func trsdSibling(env Env, dir, script, tasksMD string, stderr io.Writer) (string, bool) {
	cmd := exec.Command(dir+"/"+script, tasksMD)
	cmd.Dir = env.Dir
	cmd.Env = append(os.Environ(), "LC_ALL=C") // the bash exported LC_ALL=C
	var b bytes.Buffer
	cmd.Stdout, cmd.Stderr = &b, &b
	err := cmd.Run()
	out := strings.TrimRight(strings.ReplaceAll(b.String(), "\x00", ""), "\n")
	if err == nil {
		return out, true
	}
	var ee *exec.ExitError
	if !errors.As(err, &ee) {
		out = err.Error()
	}
	fmt.Fprintf(stderr, "%s%s failed on %s -- cannot answer\n%s\n", trsdPrefix, script, tasksMD, out)
	return "", false
}

// trsdGroups composes `bundle <i>: <ids>` with `group <g>: <bundle ids>`
// into task id -> group number. A task id listed twice keeps its first
// group, as the bash's linear search found the first; a bundle line repeated
// keeps its last, as the bash's array assignment did. Bundle numbers compare
// as decimal values; the producers never print a leading zero, which bash's
// arithmetic would have read as octal.
func trsdGroups(bundleLines, groupLines string) map[string]string {
	ifs := func(r rune) bool { return r == ' ' || r == '\t' || r == '\n' }
	bundles := map[int64]string{}
	for _, l := range strings.Split(bundleLines, "\n") {
		if m := trsdBundleRE.FindStringSubmatch(l); m != nil {
			bundles[bashDecimal(m[1])] = m[2]
		}
	}
	groupOf := map[string]string{}
	for _, l := range strings.Split(groupLines, "\n") {
		m := trsdGroupRE.FindStringSubmatch(l)
		if m == nil {
			continue
		}
		for _, bid := range strings.FieldsFunc(m[2], ifs) {
			for _, tid := range strings.FieldsFunc(bundles[bashDecimal(bid)], ifs) {
				if _, ok := groupOf[tid]; !ok {
					groupOf[tid] = m[1]
				}
			}
		}
	}
	return groupOf
}

// trsdNewestClass is `[.[] | .decision.class // empty] | first // empty` under
// jq -r, captured by `$(...)`: one output per input value, "" when there is
// none. Output that is not JSON reads as "" (class `big`), as the bash's
// `jq empty` gate did. JSON jq failed on -- a non-object row, a `.decision`
// that is neither an object nor null, a top-level value that is not an
// array -- reports !ok, the caller's cannot-answer. A top-level object, which jq
// iterated, is refused the same way, as pfdPanelFixKeys's comment explains:
// the verb prints an array.
// `flow record decisions` lists newest first, so `first` is the newest
// decision -- a class raised on a later run is the one in force.
func trsdNewestClass(raw []byte) (string, bool) {
	vals, ok := pfdStream(raw)
	if !ok {
		return "", true
	}
	var out strings.Builder
	for _, v := range vals {
		arr, ok := v.([]any)
		if !ok {
			return "", false
		}
		var last any
		for _, e := range arr {
			if e == nil {
				continue
			}
			row, ok := e.(map[string]any)
			if !ok {
				return "", false
			}
			var class any
			switch d := row["decision"].(type) {
			case nil:
			case map[string]any:
				class = d["class"]
			default:
				return "", false
			}
			if class != nil && class != false && last == nil {
				last = class
			}
		}
		switch l := last.(type) {
		case nil:
		case string:
			out.WriteString(l + "\n")
		default:
			j, _ := json.Marshal(l)
			out.WriteString(string(j) + "\n")
		}
	}
	return strings.TrimRight(strings.ReplaceAll(out.String(), "\x00", ""), "\n"), true
}

// guardSelfDir is the directory of the shim a guard was invoked through: the
// shim exports the path it was invoked by as FLOW_GUARD_SELF, relative to the
// caller's directory when not absolute, and the guard execs its siblings from
// beside it. Unset is a cannot-answer, printed under the guard's prefix and
// naming the guard's own script.
func guardSelfDir(env Env, stderr io.Writer, prefix, name string) (string, bool) {
	self := env.Getenv("FLOW_GUARD_SELF")
	if self == "" {
		fmt.Fprintf(stderr, "%sFLOW_GUARD_SELF is unset — run scripts/%s.sh, which sets it\n", prefix, name)
		return "", false
	}
	if !filepath.IsAbs(self) {
		self = filepath.Join(env.Dir, self)
	}
	return filepath.Dir(self), true
}
