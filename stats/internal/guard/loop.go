package guard

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

// The pipeline-fix and self-review fix loops (skills/flow-contracts/pipeline.md
// **Pipeline defects found mid-run**, skills/flow-self-review/SKILL.md step 3)
// as mechanics rather than prose: check-loop-dispatch.sh admits each loop
// dispatch and records it in flowd, check-loop-landing.sh refuses a landing
// those records do not cover, and pipeline-fix.sh runs the loop's fixed
// steps. A loop is its own change in <agents repo>'s flowd project, named
// after its branch -- pipeline-fix-<name>-<k> or self-review-<name> --
// except a self-review on the change's own branch, which has no branch of its
// own and files its rows under <name> itself -- in the flowd project of the
// caller's working directory, which for a change in another project is that
// project, not <agents repo>. Either way the rows carry the loop's branch name
// as their session token, so one loop's rows never mix with the run's own.
func init() {
	Registry["check-loop-dispatch"] = checkLoopDispatch
}

// lpRow is one row of the pairs table in skills/flow-self-review/SKILL.md:
// the key suffixes it covers, the subagent types and model it runs on, and
// the subagent types a recorded no-progress escalation also admits.
type lpRow struct {
	suffix *regexp.Regexp
	types  []string
	model  string
	escal  []string
}

// lpSkill is the one canonical place the pairs are read from.
const lpSkill = "skills/flow-self-review/SKILL.md"

var (
	lpTick        = regexp.MustCompile("`([^`]+)`")
	lpPlaceholder = regexp.MustCompile(`<[a-z]+>`)
	lpPipelineKey = regexp.MustCompile(`^pipeline-fix-([0-9]+)-(.+)$`)
	lpDigits      = regexp.MustCompile(`^[0-9]+$`)
)

// lpTable parses the table headed `| Key suffix |` out of <repo>/lpSkill.
// Each cell's backticked tokens are its values; `<x>` in a suffix stands for
// a number. A file with no such table, or a row missing a cell, is an error:
// a loop dispatch is never admitted off a table that cannot be read.
func lpTable(repo string) ([]lpRow, error) {
	b, err := os.ReadFile(filepath.Join(repo, lpSkill))
	if err != nil {
		return nil, err
	}
	var rows []lpRow
	in := false
	for _, line := range strings.Split(string(b), "\n") {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "| Key suffix |") {
			in = true
			continue
		}
		if !in {
			continue
		}
		if !strings.HasPrefix(t, "|") {
			break
		}
		if strings.HasPrefix(t, "|---") {
			continue
		}
		cells := strings.Split(strings.Trim(t, "|"), "|")
		if len(cells) != 4 {
			return nil, fmt.Errorf("%s: pairs row %q has %d cells, want 4", lpSkill, t, len(cells))
		}
		ticks := func(c string) []string {
			var out []string
			for _, m := range lpTick.FindAllStringSubmatch(c, -1) {
				out = append(out, m[1])
			}
			return out
		}
		var alts []string
		for _, s := range ticks(cells[0]) {
			alts = append(alts, lpPlaceholder.ReplaceAllString(regexp.QuoteMeta(s), "[0-9]+"))
		}
		model := ticks(cells[2])
		if len(alts) == 0 || len(ticks(cells[1])) == 0 || len(model) != 1 {
			return nil, fmt.Errorf("%s: pairs row %q names no suffix, type or single model", lpSkill, t)
		}
		rows = append(rows, lpRow{
			suffix: regexp.MustCompile("^(?:" + strings.Join(alts, "|") + ")$"),
			types:  ticks(cells[1]),
			model:  model[0],
			escal:  ticks(cells[3]),
		})
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("%s carries no `| Key suffix |` pairs table", lpSkill)
	}
	return rows, nil
}

// lpLoop names one loop's records: the change they are filed under, the
// session token (the loop's branch name), the key prefix of its rows, and
// main, the main checkout every flow call and git config read for the loop
// runs with -C.
type lpLoop struct {
	change, token, prefix, main string
}

// lpMain is <repo>'s main checkout, the -C every flow call runs with.
func lpMain(env Env, git, repo string) (string, bool) {
	rc, out := kwRun(env, "", git, "-C", repo, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if rc != 0 {
		return "", false
	}
	return filepath.Dir(strings.TrimSpace(out)), true
}

// lpBranchExists is whether refs/heads/<b> exists in repo.
func lpBranchExists(env Env, git, repo, b string) bool {
	rc, _ := kwRun(env, "", git, "-C", repo, "rev-parse", "-q", "--verify", "refs/heads/"+b)
	return rc == 0
}

// lpFromKey splits key into its loop and its suffix, and finds the suffix's
// pairs row. ok false with a message is a refusal (exit 1); ok false with
// msg "" means the key is no loop key at all.
func lpFromKey(env Env, git, repo, key string, rows []lpRow) (lpLoop, string, lpRow, string, bool) {
	rowOf := func(s string) (lpRow, bool) {
		for _, r := range rows {
			if r.suffix.MatchString(s) {
				return r, true
			}
		}
		return lpRow{}, false
	}
	if m := lpPipelineKey.FindStringSubmatch(key); m != nil {
		k, suffix := m[1], m[2]
		row, ok := rowOf(suffix)
		if !ok {
			return lpLoop{}, "", lpRow{}, fmt.Sprintf("KEY — %s: no row of the pairs table (%s) covers suffix %q", key, lpSkill, suffix), false
		}
		_, out := kwRun(env, "", git, "-C", repo, "for-each-ref", "--format=%(refname:short)", "refs/heads/")
		var hits []string
		for _, b := range strings.Fields(out) {
			if strings.HasPrefix(b, "pipeline-fix-") && strings.HasSuffix(b, "-"+k) && plainChangeName(b) {
				hits = append(hits, b)
			}
		}
		if len(hits) != 1 {
			return lpLoop{}, "", lpRow{}, fmt.Sprintf("NOT-STARTED — %s: %d branches pipeline-fix-<name>-%s; start the loop with scripts/pipeline-fix.sh start <name> %s", key, len(hits), k, k), false
		}
		return lpLoop{change: hits[0], token: hits[0], prefix: "pipeline-fix-" + k + "-", main: repo}, suffix, row, "", true
	}
	if !strings.HasPrefix(key, "self-review-") {
		return lpLoop{}, "", lpRow{}, "", false
	}
	rest := strings.TrimPrefix(key, "self-review-")
	// The longest name wins: a role suffix is the short tail.
	for i := len(rest) - 1; i > 0; i-- {
		if rest[i] != '-' {
			continue
		}
		name, suffix := rest[:i], rest[i+1:]
		row, ok := rowOf(suffix)
		if !ok || !plainChangeName(name) {
			continue
		}
		loop := lpLoop{token: "self-review-" + name, prefix: "self-review-" + name + "-", change: name, main: repo}
		if lpBranchExists(env, git, repo, loop.token) {
			loop.change = loop.token
		} else if m, ok := lpMain(env, git, env.Dir); env.Dir != "" && ok {
			// On the change's own branch the change lives in the caller's
			// project -- gymie's for a gymie run -- never necessarily <agents repo>.
			loop.main = m
		}
		return loop, suffix, row, "", true
	}
	return lpLoop{}, "", lpRow{}, fmt.Sprintf("KEY — %s: no row of the pairs table (%s) covers its suffix", key, lpSkill), false
}

// lpDispatch is one `flow record dispatches` row, the fields the loop reads.
type lpDispatch struct {
	Seq          int     `json:"seq"`
	Key          string  `json:"key"`
	Outcome      string  `json:"outcome"`
	SessionToken string  `json:"sessionToken"`
	EndedAt      *string `json:"endedAt"`
}

// lpRows is the loop's own rows, in seq order. ok false: flowd could not be
// read, which is never "no rows".
func lpRows(env Env, main string, lp lpLoop) ([]lpDispatch, bool) {
	out, _, rc := pcFlow(env, false, "record", "dispatches", "-change", lp.change, "-C", main)
	if rc != 0 {
		return nil, false
	}
	var all []lpDispatch
	if json.Unmarshal(out, &all) != nil {
		return nil, false
	}
	var mine []lpDispatch
	for _, d := range all {
		if d.SessionToken == lp.token && strings.HasPrefix(d.Key, lp.prefix) {
			mine = append(mine, d)
		}
	}
	slices.SortFunc(mine, func(a, b lpDispatch) int { return a.Seq - b.Seq })
	return mine, true
}

// lpFamily is a suffix's place in the loop: plan, fix (the first fixers,
// `fix` and `fix-g<g>`), fix-r (a review's fix), fix-v (a red line's fix),
// review or pass.
func lpFamily(suffix string) string {
	switch {
	case suffix == "pass":
		return "pass"
	case suffix == "plan" || strings.HasPrefix(suffix, "plan-"):
		return "plan"
	case suffix == "fix" || strings.HasPrefix(suffix, "fix-g"):
		return "fix"
	case strings.HasPrefix(suffix, "fix-v"):
		return "fix-v"
	case strings.HasPrefix(suffix, "fix-"):
		return "fix-r"
	case strings.HasPrefix(suffix, "review-"):
		return "review"
	}
	return ""
}

// lpEscalated is whether `scripts/pipeline-fix.sh escalate` recorded key for
// this loop: the marker is `<loop branch>:<key>`, so another loop reusing the
// key never inherits it, and land and drop clear the loop's markers.
func lpEscalated(env Env, git, repo string, lp lpLoop, key string) bool {
	_, out := kwRun(env, "", git, "-C", repo, "config", "--get-all", "flow.escalation")
	return slices.Contains(strings.Fields(out), lp.token+":"+key)
}

// lpEarlierRound is whether rows hold a completed earlier round of suffix's
// kind -- any fix family for a fix -- the no-progress escalation's premise,
// re-read from the rows on every admission so a hand-set marker alone admits
// nothing.
func lpEarlierRound(lp lpLoop, suffix string, rows []lpDispatch) bool {
	fam := lpFamily(suffix)
	isFix := strings.HasPrefix(fam, "fix")
	for _, d := range rows {
		f := lpFamily(strings.TrimPrefix(d.Key, lp.prefix))
		if d.Outcome == "completed" && (f == fam || (isFix && strings.HasPrefix(f, "fix"))) {
			return true
		}
	}
	return false
}

var (
	lpFirstReviewID = regexp.MustCompile(`\*\*ID:\*\*\s*first-review-effort\b`)
	lpChosen        = regexp.MustCompile(`\*\*Chosen:\*\*[^\n]*`)
	lpReviewType    = regexp.MustCompile("`(flow-(?:low|medium))`")
)

// lpFirstReviewType is the subagent_type the planner recorded for review-1 in
// the loop's design.md (its `first-review-effort` decision's **Chosen:**
// line), read from the loop's branch -- the change's own spectre/<name> for a
// self-review on it; flow-medium when it recorded none.
func lpFirstReviewType(env Env, git, repo string, lp lpLoop) string {
	ref := lp.token
	if !lpBranchExists(env, git, repo, ref) {
		ref = "spectre/" + lp.change
	}
	rc, body := kwRun(env, "", git, "-C", repo, "show", "refs/heads/"+ref+":spectre/changes/"+lp.change+"/design.md")
	if rc != 0 {
		return "flow-medium"
	}
	if loc := lpFirstReviewID.FindStringIndex(body); loc != nil {
		if c := lpChosen.FindString(body[loc[1]:]); c != "" {
			if m := lpReviewType.FindStringSubmatch(c); m != nil {
				return m[1]
			}
		}
	}
	return "flow-medium"
}

func checkLoopDispatch(args []string, env Env, stdout, stderr io.Writer) int {
	const self = "check-loop-dispatch"
	if len(args) != 4 || args[0] == "" || args[1] == "" {
		fmt.Fprint(stderr, "usage: check-loop-dispatch.sh <agents-repo> <key> <subagent_type> <model>\n")
		return 2
	}
	repo, key, typ, model := smcAbs(env, args[0]), args[1], args[2], args[3]
	git, ok := panelResolveGit(env, self, stderr)
	if !ok {
		return 2
	}
	rows, err := lpTable(repo)
	if err != nil {
		fmt.Fprintf(stderr, "%s: %v\n", self, err)
		return 2
	}
	main, ok := lpMain(env, git, repo)
	if !ok {
		fmt.Fprintf(stderr, "%s: %s is not a git checkout\n", self, repo)
		return 2
	}
	refuse := func(msg string) int {
		fmt.Fprintf(stdout, "LOOP-DISPATCH-REFUSED: %s\n", msg)
		return 1
	}
	lp, suffix, row, msg, ok := lpFromKey(env, git, main, key, rows)
	if !ok {
		if msg == "" {
			fmt.Fprintf(stdout, "LOOP-DISPATCH-NA %s\n", key)
			return 0
		}
		return refuse(msg)
	}
	family := lpFamily(suffix)
	var mine []lpDispatch
	if family != "pass" {
		if mine, ok = lpRows(env, lp.main, lp); !ok {
			fmt.Fprintf(stderr, "%s: flow record dispatches -change %s failed -- cannot answer (is flowd up?)\n", self, lp.change)
			return 2
		}
	}
	if model != row.model || !slices.Contains(row.types, typ) {
		if model != row.model || !slices.Contains(row.escal, typ) {
			return refuse(fmt.Sprintf("PAIR — %s runs on subagent_type %s, model %s (the pairs table, %s); this dispatch carries subagent_type %q, model %q",
				key, strings.Join(row.types, " or "), row.model, lpSkill, typ, model))
		}
		if !lpEscalated(env, git, lp.main, lp, key) || !lpEarlierRound(lp, suffix, mine) {
			return refuse(fmt.Sprintf("ESCALATION — %s on %s needs a completed earlier round and its no-progress escalation recorded first: scripts/pipeline-fix.sh escalate %s", key, typ, key))
		}
	} else if suffix == "review-1" {
		if want := lpFirstReviewType(env, git, lp.main, lp); typ != want {
			return refuse(fmt.Sprintf("PAIR — %s runs on %s, as the loop's design.md first-review-effort decision records (flow-medium when none); this dispatch carries %q", key, want, typ))
		}
	}
	if family == "pass" {
		fmt.Fprintf(stdout, "LOOP-DISPATCH-OK %s\n", key)
		return 0
	}
	planDone := false
	for _, d := range mine {
		if d.Key == key && d.Outcome != "aborted" {
			return refuse(fmt.Sprintf("KEY-REUSED — %s already ran; the next dispatch takes the next key", key))
		}
		if d.EndedAt == nil {
			return refuse(fmt.Sprintf("OPEN — %s has not been closed: scripts/pipeline-fix.sh end %s <outcome>", d.Key, d.Key))
		}
		if lpFamily(strings.TrimPrefix(d.Key, lp.prefix)) == "plan" {
			planDone = d.Outcome == "completed"
		}
	}
	if family != "plan" && !planDone {
		return refuse(fmt.Sprintf("PLAN — %s needs a completed %splan dispatch first", key, lp.prefix))
	}
	role := map[string]string{"plan": "planner", "fix": "implementer", "fix-r": "implementer", "fix-v": "implementer", "review": "reviewer"}[family]
	effort := strings.TrimPrefix(typ, "flow-")
	if !slices.Contains([]string{"low", "medium", "high"}, effort) {
		effort = "default"
	}
	out, _, rc := pcFlow(env, true, "record", "dispatch", "begin", "-C", lp.main, "-change", lp.change,
		"-role", role, "-model", model, "-effort", effort, "-key", key, "-session-token", lp.token)
	if rc != 0 {
		fmt.Fprintf(stderr, "%s: flow record dispatch begin failed (exit %d) -- cannot answer:\n%s", self, rc, out)
		fmt.Fprintf(stderr, "%s: a loop's rows need its change: scripts/pipeline-fix.sh start creates it\n", self)
		return 2
	}
	fmt.Fprintf(stdout, "LOOP-DISPATCH-OK %s\n", key)
	return 0
}
