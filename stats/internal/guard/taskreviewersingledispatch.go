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

// The contract below moved here from the header of the deleted
// scripts/check-task-reviewer-single-dispatch.sh shim (KAN-1058).
//
// flow-guard check-task-reviewer-single-dispatch <worktree> <change-name> <session-token>
//
// THE CHANGE NAME IS A SECOND ARGUMENT and the session token a third, for
// the same reason the name is one on flow-guard check-panel-fix-single-dispatch: a
// change's dispatches are rows in the store, keyed by change name, not a
// file at a fixed path, and the token scopes the count to THIS run --
// `flow record dispatches -change <name> -C <worktree>` answers with the
// decoded JSON array of that change's dispatch rows, and this guard reads
// only that array plus <worktree>/spectre/changes/<name>/tasks.md (the
// canonical worktree's copy -- the only one that exists on a cross-repo
// change, per skills/flow-contracts/worktree-resolution.md).
//
// This is the gate KAN-527 asked for, the gated-per-task-reviewer sibling
// of flow-guard check-panel-fix-single-dispatch: on that run the conductor launched
// nine one-task reviewer dispatches against nine gate-fired tasks that
// belonged to three implementer groups, before the operator stopped the
// run and implement.md's "The gated per-task reviewer" paragraph was
// rewritten to require ONE bundled dispatch per implementer group whose
// gate fired (on `big`; KAN-934 later dropped the gated reviewer on
// `micro`/`small`/`regular`). Prose
// alone already failed once for the panel's own fix step (KAN-482); this
// guard exists so the same failure mode is caught here too, at the stage's
// own close, immediately before `flow stage end -command '/flow' -stage
// flow.sdd-tdd`, the same position flow-guard check-panel-fix-single-dispatch holds
// at `flow.review-panel`'s close.
//
// THE COUNT IS SCOPED to the session token named on the command line --
// a change's rows span every run that ever touched it, and only this run's
// dispatches are this run's business. Within that token, every row whose
// role is `reviewer` and whose key matches the gated-per-task-reviewer
// family (`^task-[0-9]+(\+[0-9]+)*-reviewer(-fix-[0-9]+)?(-retry)?$`) must
// satisfy:
//
//   - its key is in that canonical shape -- a key naming one task
//     (`task-5-reviewer`) or several `+`-joined in plan order
//     (`task-13+14+15-reviewer`), optionally a fix round
//     (`task-13+14-reviewer-fix-1`) and optionally the handshake's one
//     retry (`task-13+14-reviewer-fix-1-retry`); a conductor that invents
//     one key per task (`task-5-reviewer`, `task-6-reviewer`, ... for tasks
//     that share an implementer group) is exactly the KAN-527 shape;
//   - its base (the key with a trailing `-retry` stripped) carries at most
//     one non-retry dispatch, at most two rows total, and two only when
//     exactly one of them is the `-retry` variant -- the handshake's one
//     allowed second dispatch, per implement.md's **The handshake**;
//   - a `-retry` row never stands alone: a retry with no original has no
//     bundle it could be a retry of.
//
// BEYOND the per-key shape (which is everything check-panel-fix-single-
// dispatch.sh's own key family checks), this guard also checks the
// BUNDLING ITSELF, since a KAN-527-shaped run can produce well-formed keys
// that are still one-per-task: no two ORIGINAL (non-fix, non-retry) keys
// may name tasks belonging to the same implementer group. Group membership
// is read from `<worktree>/spectre/changes/<name>/tasks.md` via
// `flow-guard plan-dispatch-bundles` (task -> bundle) composed with
// `flow-guard plan-dispatch-groups` (bundle -> group) -- the same source of truth
// `skills/flow/brainstorm-planner.md`'s Decide step and implement.md's own
// wave-grouping already use, never re-derived by this guard from the raw
// markdown. On the decision's `class` `micro`, `small` or `regular` (read via `flow
// record decisions -change <name> -C <worktree>`, the newest entry's -- the verb lists newest first --
// `.decision.class`; a failed read, output that is not JSON, or no class
// is treated as `big`, and JSON that is not an array of decision rows is a
// cannot-answer, exit 2 -- `big` tolerates gated reviewer rows, so reading
// it as `big` would loosen the check),
// implement.md dispatches no gated reviewer at all: every gate-fired task is
// the whole-branch panel's review focus instead (KAN-934) -- so on those
// classes, any gated-reviewer row of this token is itself a violation.
//
// Rows of other roles and rows of other session tokens never count. A task
// id that never appears in any `reviewer`-role key of this token was never
// gated, or its gate hasn't fired yet, or its bundle is still in flight --
// this guard says nothing about coverage, only about the shape of what was
// already dispatched.
//
// THE GUARD NEVER READS THE JOURNAL and never consults anything but the
// store's answer for dispatch rows, matching check-panel-fix-single-
// dispatch.sh: a read that failed is a question this guard cannot answer,
// not a clean verdict. The tasks.md read is a plain file read, not a store
// read; a missing or unparsable tasks.md is likewise "cannot answer" (exit
// 2), never a clean verdict, since without it group membership cannot be
// established at all.
//
// THE CHANGE-NAME CONTAINMENT CASE IS DUPLICATED, on purpose, from
// flow-guard check-panel-fix-single-dispatch's own copy at d71a2327 (now the
// CONTAINMENT check in stats/internal/guard/panelfixsingledispatch.go; itself duplicated from
// flow-guard check-panel-findings-closed) -- the change name arrives from a
// pull-request-editable state file and is passed to `flow record
// dispatches -change`, so `../../../planted` and a glob metacharacter are
// hazards here exactly as they are there. The worktree is canonicalised
// before it is ever passed to `flow` as `-C`, and refused rather than
// proceeded with if it vanished between the `-d` check and here, for the
// same reasons that guard's header names.
//
// Exit codes:
//   0  every gated-per-task reviewer row of this session token is
//      shape-clean, retry-clean, and bundled per implementer group (and,
//      on `micro`/`small`/`regular`, there is no such row)
//   1  at least one violation -- each offending key or group named on
//      stderr
//   2  cannot answer at all -- missing arguments, a non-directory
//      worktree, a change name outside the allowlist, the store
//      unreachable, tasks.md missing or unreadable, flow-guard plan-dispatch-bundles
//      or flow-guard plan-dispatch-groups failing, dispatch rows that are not
//      readable JSON, or decision output that is JSON but not an array of
//      decision rows

// checkTaskReviewerSingleDispatch is `flow-guard check-task-reviewer-single-dispatch`: the comment above is the contract -- exit 0 every gated-per-task reviewer row of the
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
		fmt.Fprintf(stderr, "%susage: flow-guard check-task-reviewer-single-dispatch <worktree> <change-name> <session-token>\n", trsdPrefix)
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

	tasksMD := ChangeDir(worktree+"/spectre/changes", name) + "/tasks.md"
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
	bundleLines, ok := trsdSibling(env, "FLOW_GUARD_PLAN_DISPATCH_BUNDLES", "plan-dispatch-bundles", tasksMD, stderr)
	if !ok {
		return 2
	}
	groupLines, ok := trsdSibling(env, "FLOW_GUARD_PLAN_DISPATCH_GROUPS", "plan-dispatch-groups", tasksMD, stderr)
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

	// On `micro`/`small`/`regular`, implement.md dispatches no gated reviewer
	// at all -- every gate-fired task is the whole-branch panel's review
	// focus -- so any gated-reviewer row of this token is itself a violation
	// on those classes.
	if (class == "micro" || class == "small" || class == "regular") && len(keys) > 0 {
		violations = append(violations, fmt.Sprintf("class '%s' carries %d gated reviewer dispatch(es) (%s) -- on micro/small/regular no gated reviewer runs; the whole-branch panel reviews every gate-fired task", class, len(keys), strings.Join(keys, " ")))
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
func trsdSibling(env Env, override, script, tasksMD string, stderr io.Writer) (string, bool) {
	cmd := siblingCmd(env, override, script, tasksMD)
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
