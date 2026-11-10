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

	"github.com/tweety53/agents/stats/internal/modelpolicy"
)

// The contract below moved here from the header of the deleted
// scripts/check-loop-dispatch.sh shim (KAN-1058).
//
// flow-guard check-loop-dispatch <agents-repo> <key> <subagent_type> <model> — admit
// one dispatch of a pipeline-fix or self-review fix loop, its prompt on stdin,
// or refuse it, and record it (hooks/enforce-loop-dispatch.py runs it on every
// Agent/Task call whose prompt carries a `Key: pipeline-fix-<k>-…` or
// `Key: self-review-…` line). The key's suffix names its role id of the
// **Allowed pairs** table in skills/flow-contracts/model-policy.md — the one
// place the pairs are read from.
//
// Refused (exit 1, one LOOP-DISPATCH-REFUSED line):
//   KEY          the key's suffix is no loop suffix;
//   NOT-STARTED  a pipeline-fix key with no pipeline-fix-<name>-<k> branch
//                (flow-guard pipeline-fix start makes it);
//   PAIR         subagent_type or model off the role's rows, or a marked
//                row's marker (`No-progress escalation`) not a prompt line;
//   EFFORT       a review-1 whose loop plan records a first-review-effort
//                decision naming no effort on its **Chosen:** line;
//   ESCALATION   the role's marked escalation pair with no recorded escalation
//                (flow-guard pipeline-fix escalate);
//   KEY-REUSED   the key already has a row;
//   OPEN         an earlier row of the loop is not closed
//                (flow-guard pipeline-fix end);
//   PLAN         a fix or review with no completed plan row;
//   PASS-DONE    a pass whose key already holds a completed row while its
//                loop branch self-review-<name> exists (kan-1102 self-review, F20);
//   NARRATIVE    a pass dispatched from the change branch (spectre/<name> or
//                <name>) whose prompt's first absolute …/bundle.md path is
//                missing, unreadable, or has no line `## Session narrative`
//                (kan-1102 self-review, F21).
// Admitted (exit 0, LOOP-DISPATCH-OK): the row is opened with `flow record
// dispatch begin` under the loop's change, its session token the loop's
// branch name — except a `pass`, which is opened under the change <name>
// itself in the caller's project, its token self-review-<name>, the parent
// closing it when the pass hands back. A
// self-review on the change's own branch (no self-review-<name> branch in
// <agents-repo>) reads and records in the flowd project of the working
// directory it runs in — the hook runs it in the payload's cwd.
// LOOP-DISPATCH-NA, exit 0: the key is no loop key.
// Exit 2: cannot answer — the table unreadable, flowd unreadable, the begin
// refused (a loop change no checkout carries: flow-guard pipeline-fix start), usage.
//
// Go, in flow-guard (stats/internal/guard/loop.go).

// The pipeline-fix and self-review fix loops (skills/flow-contracts/pipeline.md
// **Pipeline defects found mid-run**, skills/flow-self-review/SKILL.md step 3)
// as mechanics rather than prose: check-loop-dispatch.sh admits each loop
// dispatch and records it in flowd, check-loop-landing.sh refuses a landing
// those records do not cover, and flow-guard pipeline-fix runs the loop's fixed
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

var (
	lpPipelineKey = regexp.MustCompile(`^pipeline-fix-([0-9]+)-(.+)$`)
	lpDigits      = regexp.MustCompile(`^[0-9]+$`)
	lpBundlePath  = regexp.MustCompile("(?:^|[\\s`(])(/[^\\s`]*/bundle\\.md)")
	lpSuffix      = regexp.MustCompile(`^(?:pass|plan(?:-[0-9]+)?|fix(?:-[0-9]+|-g[0-9]+|-v[0-9]+)?|review-[0-9]+)$`)
)

// lpRole is the allowed-pairs role id (skills/flow-contracts/model-policy.md)
// of a loop key's suffix under prefix (self-review or pipeline-fix), "" when
// the suffix is no loop suffix.
func lpRole(prefix, suffix string) string {
	if !lpSuffix.MatchString(suffix) {
		return ""
	}
	fam := lpFamily(suffix)
	switch {
	case strings.HasPrefix(fam, "fix"):
		fam = "fix"
	case suffix == "review-1":
		fam = "review-1"
	case fam == "review":
		fam = "review-r"
	}
	return prefix + "-" + fam
}

// lpEscalation is the marker of the marked row of role that lists pair, ""
// when no marked row does -- the no-progress escalation's pair.
func lpEscalation(rows map[string][]modelpolicy.Row, role, pair string) string {
	for _, r := range rows[role] {
		if r.Marker != "" && slices.Contains(r.Pairs, pair) {
			return r.Marker
		}
	}
	return ""
}

// lpEscalatable is whether role has any marked row: a fix's, never a plan's or a review's.
func lpEscalatable(rows map[string][]modelpolicy.Row, role string) bool {
	return slices.ContainsFunc(rows[role], func(r modelpolicy.Row) bool { return r.Marker != "" })
}

// lpLoop names one loop's records: the change they are filed under, the
// session token (the loop's branch name), the key prefix of its rows, and
// main, the main checkout every flow call and git config read for the loop
// runs with -C.
type lpLoop struct {
	change, token, prefix, main string
}

// lpEffort is the effort a subagent_type names (flow-low, flow-planner-medium), "default" for any other.
func lpEffort(typ string) string {
	effort := strings.TrimPrefix(strings.TrimPrefix(typ, "flow-"), "planner-")
	if !slices.Contains([]string{"low", "medium", "high"}, effort) {
		return "default"
	}
	return effort
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

// lpFromKey splits key into its loop, its suffix and the suffix's role id.
// ok false with a message is a refusal (exit 1); ok false with msg "" means
// the key is no loop key at all.
func lpFromKey(env Env, git, repo, key string) (lpLoop, string, string, string, bool) {
	if m := lpPipelineKey.FindStringSubmatch(key); m != nil {
		k, suffix := m[1], m[2]
		role := lpRole("pipeline-fix", suffix)
		if role == "" {
			return lpLoop{}, "", "", fmt.Sprintf("KEY — %s: suffix %q is no loop suffix", key, suffix), false
		}
		_, out := kwRun(env, "", git, "-C", repo, "for-each-ref", "--format=%(refname:short)", "refs/heads/")
		var hits []string
		for _, b := range strings.Fields(out) {
			if strings.HasPrefix(b, "pipeline-fix-") && strings.HasSuffix(b, "-"+k) && plainChangeName(b) {
				hits = append(hits, b)
			}
		}
		if len(hits) != 1 {
			return lpLoop{}, "", "", fmt.Sprintf("NOT-STARTED — %s: %d branches pipeline-fix-<name>-%s; start the loop with flow-guard pipeline-fix start <name> %s", key, len(hits), k, k), false
		}
		return lpLoop{change: hits[0], token: hits[0], prefix: "pipeline-fix-" + k + "-", main: repo}, suffix, role, "", true
	}
	if !strings.HasPrefix(key, "self-review-") {
		return lpLoop{}, "", "", "", false
	}
	rest := strings.TrimPrefix(key, "self-review-")
	// The longest name wins: a role suffix is the short tail.
	for i := len(rest) - 1; i > 0; i-- {
		if rest[i] != '-' {
			continue
		}
		name, suffix := rest[:i], rest[i+1:]
		role := lpRole("self-review", suffix)
		if role == "" || !plainChangeName(name) {
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
		return loop, suffix, role, "", true
	}
	return lpLoop{}, "", "", fmt.Sprintf("KEY — %s: its suffix is no loop suffix", key), false
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

// lpEscalated is whether `flow-guard pipeline-fix escalate` recorded key for
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
	lpEntryID       = regexp.MustCompile(`\*\*ID:\*\*`)
	lpChosen        = regexp.MustCompile(`\*\*Chosen:\*\*[^\n]*`)
	lpStatus        = regexp.MustCompile(`\*\*Status:\*\*\s*([a-z]+)`)
	// The effort a **Chosen:** line names: a bare low, medium, flow-low or flow-medium opening the
	// value, else its first backticked low, medium, flow-low or flow-medium.
	lpChosenBare = regexp.MustCompile(`^\*\*Chosen:\*\*\s*(?:flow-)?(low|medium)\b`)
	lpChosenTick = regexp.MustCompile("`(?:flow-)?(low|medium)`")
)

// lpFirstReviewType is the subagent_type the planner recorded for review-1 in
// the loop's design.md (its `first-review-effort` decision's **Chosen:**
// line). The loop's plan is always spectre/changes/<loop branch name>/ --
// self-review-<name> on the change's own branch too -- read from disk in the
// loop's worktree, <main>-worktrees/<branch name without spectre/>: the loop's
// branch, or for a self-review on the change's own branch spectre/<name>
// (/flow) or <name> (/flow-fast). The decision is the first such entry whose **Status:**
// is active (or absent); flow-medium when it recorded none. ok is
// false when the decision is recorded but its **Chosen:** line names no
// effort: a refusal, never a silent flow-medium.
func lpFirstReviewType(env Env, git, repo string, lp lpLoop) (typ string, ok bool) {
	ref := lp.token
	for _, b := range []string{"spectre/" + lp.change, lp.change} {
		if lpBranchExists(env, git, repo, ref) {
			break
		}
		ref = b
	}
	b, err := os.ReadFile(repo + "-worktrees/" + strings.TrimPrefix(ref, "spectre/") + "/spectre/changes/" + lp.token + "/design.md")
	if err != nil {
		return "flow-medium", true
	}
	body := string(b)
	entry, found := "", false
	for _, loc := range lpFirstReviewID.FindAllStringIndex(body, -1) {
		entry = body[loc[1]:]
		if next := lpEntryID.FindStringIndex(entry); next != nil {
			entry = entry[:next[0]]
		}
		if s := lpStatus.FindStringSubmatch(entry); s == nil || s[1] == "active" {
			found = true
			break
		}
	}
	if !found {
		return "flow-medium", true
	}
	c := lpChosen.FindString(entry)
	for _, re := range []*regexp.Regexp{lpChosenBare, lpChosenTick} {
		if m := re.FindStringSubmatch(c); m != nil {
			return "flow-" + m[1], true
		}
	}
	return "", false
}

func checkLoopDispatch(args []string, env Env, stdout, stderr io.Writer) int {
	const self = "check-loop-dispatch"
	if len(args) != 4 || args[0] == "" || args[1] == "" {
		fmt.Fprint(stderr, "usage: flow-guard check-loop-dispatch <agents-repo> <key> <subagent_type> <model>\n")
		return 2
	}
	repo, key, typ, model := smcAbs(env, args[0]), args[1], args[2], args[3]
	git, ok := panelResolveGit(env, self, stderr)
	if !ok {
		return 2
	}
	rows, err := policyTable(env, repo)
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
	var prompt []byte
	if env.Stdin != nil {
		if prompt, err = io.ReadAll(env.Stdin); err != nil {
			fmt.Fprintf(stderr, "%s: reading the prompt: %v\n", self, err)
			return 2
		}
	}
	lp, suffix, role, msg, ok := lpFromKey(env, git, main, key)
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
	if err := policyAllows(rows, role, typ, model, "", ""); err != nil {
		// Off the unmarked rows: only the role's marked escalation row can admit it.
		if lpEscalation(rows, role, policyPair(typ, model, "")) == "" {
			return refuse("PAIR — " + err.Error())
		}
		if !lpEscalated(env, git, lp.main, lp, key) || !lpEarlierRound(lp, suffix, mine) {
			return refuse(fmt.Sprintf("ESCALATION — %s on %s needs a completed earlier round and its no-progress escalation recorded first: flow-guard pipeline-fix escalate %s", key, typ, key))
		}
		if err := policyAllows(rows, role, typ, model, "", string(prompt)); err != nil {
			return refuse("PAIR — " + err.Error())
		}
	} else if suffix == "review-1" {
		want, ok := lpFirstReviewType(env, git, lp.main, lp)
		if !ok {
			return refuse(fmt.Sprintf("EFFORT — %s: the loop's design.md first-review-effort decision names no effort on its **Chosen:** line; write low, medium, `flow-low` or `flow-medium` there", key))
		}
		if typ != want {
			return refuse(fmt.Sprintf("PAIR — %s runs on %s, as the loop's design.md first-review-effort decision records (flow-medium when none); this dispatch carries %q", key, want, typ))
		}
	}
	if family == "pass" {
		// The pass belongs to the change's own self-review stage, never to the
		// fix loop it may start: it runs in the change's session, before any
		// self-review-<name> change exists, so it is recorded under <name> in
		// the caller's project -- the row the run view reads it from.
		change, main := lpPassTarget(env, git, lp)
		// A pass already run is never run again while the loop its findings started is
		// open: a re-run would empty <sr-dir> and lose the findings the loop fixes
		// (kan-1102 self-review, F20).
		if lpBranchExists(env, git, lp.main, lp.token) {
			done, ok := lpRows(env, main, lpLoop{change: change, token: lp.token, prefix: lp.prefix})
			if !ok {
				fmt.Fprintf(stderr, "%s: flow record dispatches -change %s failed -- cannot answer (is flowd up?)\n", self, change)
				return 2
			}
			if slices.ContainsFunc(done, func(d lpDispatch) bool { return d.Key == key && d.Outcome == "completed" }) {
				return refuse(fmt.Sprintf("PASS-DONE — %s already completed and its loop branch %s exists: resume the loop at step 3 from <sr-dir>/findings.md (skills/flow-self-review/SKILL.md step 2)", key, lp.token))
			}
		}
		if msg := lpNarrative(env, git, key, change, string(prompt)); msg != "" {
			return refuse(msg)
		}
		if out, _, rc := pcFlow(env, true, "record", "dispatch", "begin", "-C", main, "-change", change,
			"-role", "reviewer", "-model", model, "-effort", lpEffort(typ), "-key", key, "-session-token", lp.token); rc != 0 {
			fmt.Fprintf(stderr, "%s: flow record dispatch begin failed (exit %d) -- cannot answer:\n%s", self, rc, out)
			return 2
		}
		fmt.Fprintf(stdout, "LOOP-DISPATCH-OK %s\n", key)
		return 0
	}
	planDone := false
	for _, d := range mine {
		if d.Key == key && d.Outcome != "aborted" {
			return refuse(fmt.Sprintf("KEY-REUSED — %s already ran; the next dispatch takes the next key", key))
		}
		if d.EndedAt == nil {
			return refuse(fmt.Sprintf("OPEN — %s has not been closed: flow-guard pipeline-fix end %s <outcome>", d.Key, d.Key))
		}
		if lpFamily(strings.TrimPrefix(d.Key, lp.prefix)) == "plan" {
			planDone = d.Outcome == "completed"
		}
	}
	if family != "plan" && !planDone {
		return refuse(fmt.Sprintf("PLAN — %s needs a completed %splan dispatch first", key, lp.prefix))
	}
	recordRole := map[string]string{"plan": "planner", "fix": "implementer", "fix-r": "implementer", "fix-v": "implementer", "review": "reviewer"}[family]
	out, _, rc := pcFlow(env, true, "record", "dispatch", "begin", "-C", lp.main, "-change", lp.change,
		"-role", recordRole, "-model", model, "-effort", lpEffort(typ), "-key", key, "-session-token", lp.token)
	if rc != 0 {
		fmt.Fprintf(stderr, "%s: flow record dispatch begin failed (exit %d) -- cannot answer:\n%s", self, rc, out)
		fmt.Fprintf(stderr, "%s: a loop's rows need its change: flow-guard pipeline-fix start creates it\n", self)
		return 2
	}
	fmt.Fprintf(stdout, "LOOP-DISPATCH-OK %s\n", key)
	return 0
}

// lpPassTarget is the change and checkout the self-review pass is recorded
// under: change <name>, the checkout the caller's directory belongs to when it
// resolves, lp.main otherwise. check-loop-dispatch begins the pass here and
// flow-guard pipeline-fix end closes it here.
func lpPassTarget(env Env, git string, lp lpLoop) (change, main string) {
	change, main = strings.TrimSuffix(strings.TrimPrefix(lp.prefix, "self-review-"), "-"), lp.main
	if m, ok := lpMain(env, git, env.Dir); env.Dir != "" && ok {
		main = m
	}
	return change, main
}

// lpNarrative is the NARRATIVE refusal of an in-run pass, "" when it is
// admitted: only a pass dispatched from the change branch -- spectre/<change>
// (/flow) or <change> (/flow-fast), skills/flow-self-review/SKILL.md step 1 --
// carries a run session to narrate, so only it is checked (decision
// f21-narrative-in-run-only, kan-1102 self-review, F21).
func lpNarrative(env Env, git, key, change, prompt string) string {
	if env.Dir == "" {
		return ""
	}
	rc, out := kwRun(env, "", git, "-C", env.Dir, "branch", "--show-current")
	if b := strings.TrimSpace(out); rc != 0 || (b != "spectre/"+change && b != change) {
		return ""
	}
	path := "no bundle path"
	if m := lpBundlePath.FindStringSubmatch(prompt); m != nil {
		path = m[1]
		if body, err := os.ReadFile(path); err == nil && slices.Contains(strings.Split(string(body), "\n"), "## Session narrative") {
			return ""
		}
	}
	return fmt.Sprintf("NARRATIVE — %s: an in-run pass's bundle must carry a ## Session narrative heading (skills/flow-contracts/finish-contract-run1.md, Run the self-review pass); %s carries none", key, path)
}
