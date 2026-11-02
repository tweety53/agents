package guard

import (
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"
)

// The contract below moved here from the header of the deleted
// scripts/check-verify-green.sh shim (KAN-1058).
//
// flow-guard check-verify-green — whether this run's inline verify is green, run by
// `flow.visual-verify` step 3 before any verifier is dispatched
// (skills/flow/visual-verify.md): visual verification starts only after
// **Verify** (skills/flow/verify-and-handoff.md) has closed green, never
// beside a verify still running or blocked.
//
// Usage: flow-guard check-verify-green <worktree> <change-name> <session-token> [<worktree>...]
//
// <worktree> is the change's canonical worktree, the one its rows are
// recorded from; each further <worktree> is another worktree of the set.
//
// Reads `flow record dispatches -change <change-name> -C <worktree>` and
// judges this session token's inline verify rows — role `verifier`, key
// `verify` or `verify-<repository name>`; the visual verifier's
// `visual-verify…` rows never count. Only the latest such row per worktree is
// judged: rows group by key with a `-fix-<k>` suffix removed
// (skills/flow/verify-fix-loop.md step 4), and the last row of each group in
// record order stands.
//
// The HEAD rule (decisions verify-row-records-head and
// green-verify-spans-planning-commits): flow-guard run-verify-list's begin records
// the verified worktree's HEAD as the row's `diffBase`. A judged row is green
// only when its outcome is `completed` or `skipped`, its `diffBase` is
// non-empty, that sha resolves as a commit in one of the <worktree>
// arguments, and every commit since it there touches only
// `<spec root>/changes/` (`git log <sha>..HEAD -- . ':(exclude)<leaf>/changes/'`
// prints nothing) — a planning commit never voids a green verify, a code
// commit always does. Prints one verdict line:
//   VERIFY-GREEN:     <change> — <n> inline verify row(s) green: <key>@<short sha>[, …]
//   VERIFY-NOT-GREEN: <change> — this run recorded no inline verify
//   VERIFY-NOT-GREEN: <change> — <key> (<why>)[, …]
// where <why> is `running`, the row's outcome, `no HEAD recorded`,
// `HEAD resolves in no worktree`, or `<commit>[, …] since <short sha>`.
// Exit 0 every judged row is green and at least one exists — flow-guard run-verify-list records `skipped` when every test
// command's `# inputs:` (and `.flow/project.md`) are unchanged since the
// list's last green run on a clean tree;
// 1 VERIFY-NOT-GREEN; 2 cannot answer (usage, a missing worktree, a change
// name that is not plain, a failed or malformed store read, a read that
// returns no row of the change at all — <worktree> is not the canonical
// worktree its rows are recorded from) — the cause on stderr.

// checkVerifyGreen is scripts/check-verify-green.sh: that script's header
// comment is the contract. The dispatch read is Env.Dispatches when set,
// else `flow record dispatches`.
func init() { Registry["check-verify-green"] = checkVerifyGreen }

const vgPrefix = "check-verify-green: "

// A fix round's `verify-fix-<k>` and a re-run's `verify-<k>` (a verify re-run
// after a new code commit, skills/flow/verify-and-handoff.md's **Recording.**)
// both supersede the `verify[-<basename>]` row before them. The number ends at
// a dash or the key's end, so a repository name starting with a digit
// (`verify-3d-viewer`) is never read as a re-run number (kan-1048 self-review, 1-A).
var vgFixSuffix = regexp.MustCompile(`^verify-(fix-)?[0-9]+(-|$)`)

func checkVerifyGreen(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) < 3 || args[1] == "" || args[2] == "" {
		fmt.Fprintf(stderr, "%susage: flow-guard check-verify-green <worktree> <change-name> <session-token> [<worktree>...]\n", vgPrefix)
		return 2
	}
	wt, name, token := args[0], args[1], args[2]
	worktrees := append([]string{wt}, args[3:]...)
	for _, w := range worktrees {
		if !isDir(pcAbs(env, w)) {
			fmt.Fprintf(stderr, "%snot a directory: %s\n", vgPrefix, w)
			return 2
		}
	}
	if !plainChangeName(name) {
		fmt.Fprintf(stderr, "%schange name '%s' is not a plain change name\n", vgPrefix, name)
		return 2
	}
	raw, err := pfdRead(env, env.Dispatches, "dispatches", name, pcAbs(env, wt))
	var rows []struct{ Key, Role, SessionToken, Outcome, DiffBase string }
	if err == nil {
		err = json.Unmarshal(raw, &rows)
	}
	if err != nil {
		fmt.Fprintf(stderr, "%sflow record dispatches failed for '%s' -- cannot answer\n", vgPrefix, name)
		return 2
	}
	// No row at all is a store read that resolved no change, never a run
	// that verified nothing: a sibling worktree resolves a project the
	// change's rows are not recorded under (kan-1008's self-review).
	if len(rows) == 0 {
		fmt.Fprintf(stderr, "%sflow record dispatches returned no row for '%s' from %s -- pass the change's canonical worktree, the one its rows are recorded from\n", vgPrefix, name, wt)
		return 2
	}
	// Rows come back in seq order; only the latest row per worktree is
	// judged, a fix round's `verify-fix-<k>[-<basename>]` or a re-run's
	// `verify-<k>[-<basename>]` superseding the `verify[-<basename>]` row
	// it re-ran.
	latest := map[string]int{}
	var groups []string
	for i, r := range rows {
		if r.SessionToken != token || r.Role != "verifier" || (r.Key != "verify" && !strings.HasPrefix(r.Key, "verify-")) {
			continue
		}
		g := vgFixSuffix.ReplaceAllString(r.Key, "verify$2")
		if _, ok := latest[g]; !ok {
			groups = append(groups, g)
		}
		latest[g] = i
	}
	seen := len(groups)
	git := envGit(env)
	var open, green []string
	for _, g := range groups {
		r := rows[latest[g]]
		switch r.Outcome {
		case "completed", "skipped":
		case "":
			open = append(open, r.Key+" (running)")
			continue
		default:
			open = append(open, r.Key+" ("+r.Outcome+")")
			continue
		}
		if why := vgHeadHolds(env, git, worktrees, r.DiffBase); why != "" {
			open = append(open, r.Key+" ("+why+")")
			continue
		}
		green = append(green, r.Key+"@"+vgShort(r.DiffBase))
	}
	switch {
	case seen == 0:
		fmt.Fprintf(stdout, "VERIFY-NOT-GREEN: %s — this run recorded no inline verify\n", name)
		return 1
	case len(open) > 0:
		fmt.Fprintf(stdout, "VERIFY-NOT-GREEN: %s — %s\n", name, strings.Join(open, ", "))
		return 1
	}
	fmt.Fprintf(stdout, "VERIFY-GREEN: %s — %d inline verify row(s) green: %s\n", name, seen, strings.Join(green, ", "))
	return 0
}

// vgHeadHolds is "" when sha, the HEAD a green row recorded, resolves in one
// of worktrees and every commit since it there touches only <leaf>/changes/
// (decision green-verify-spans-planning-commits), else why the row is not green.
func vgHeadHolds(env Env, git func(...string) *exec.Cmd, worktrees []string, sha string) string {
	if sha == "" {
		return "no HEAD recorded"
	}
	for _, w := range worktrees {
		w = pcAbs(env, w)
		if git("-C", w, "cat-file", "-e", sha+"^{commit}").Run() != nil {
			continue
		}
		out, ok := capture(git("-C", w, "log", "--format=%h", sha+"..HEAD", "--", ".", ":(exclude)"+specRootLeaf(w, io.Discard)+"/changes/"))
		if !ok {
			return "git log failed in " + w
		}
		if out != "" {
			return strings.Join(strings.Fields(out), ", ") + " since " + vgShort(sha)
		}
		return ""
	}
	return "HEAD resolves in no worktree"
}

func vgShort(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}
