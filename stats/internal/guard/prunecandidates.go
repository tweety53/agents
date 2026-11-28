package guard

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/tweety53/agents/stats/internal/records"
)

// prune-candidates [-days N] — the telemetry the self-review pass's `flow-cut` and `flow-simplify`
// angles read (skills/flow-self-review/SKILL.md). A report, never a verdict: it names candidates
// with their evidence and the pass decides.
//
// Over the last N days (default 30) it reads flowd's guard-health view and every registered
// project's incidents, and prints, each list sorted:
//
//	window: <from> .. <to>
//	never-loaded window: <first load> .. <to>           the part of the window the never-loaded lines cover
//	never-ran <guard>                                   a registered guard with no run in the window
//	never-fired <guard> runs=<n>                        a guard that ran and never refused
//	friction <guard> incidents=<n> minutes-lost=<m> latest: <symptom>
//	                                                    a guard or hook with two or more incidents
//	never-loaded <file> [conditional: <condition>]      a .md under <repo>/skills/ no session loaded
//
// A load is what the flow-skill-load hook records (stats/internal/guard/hookskillload.go): a Read
// of the file or a Skill call of its SKILL.md, counted per file — the hook cannot name the section
// a Read took, so nothing is reported per section. A file a conditional load set names
// (conditionalLoadSets) carries that set's condition, so the pass can judge whether the
// condition could have held in the window. A window holding no load at all — the hook unregistered,
// or newly so — prints `never-loaded: no load recorded in the window` instead of every file.
// <repo> is FLOW_GUARD_REPO_ROOT, the checkout scripts/flow-guard exports; unset, no never-loaded
// line is printed.
//
// Helpers (records.GuardHelpers) refuse nothing and are never listed. The daemon is
// FLOW_RECORDS_ADDR, then FLOW_ADDR, then the dev flowd — where guard runs are recorded.
//
// Exit 0 with the report; 2 when flowd cannot be read or the flags are wrong.
func init() { Registry["prune-candidates"] = pruneCandidates }

func pruneCandidates(args []string, env Env, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("prune-candidates", flag.ContinueOnError)
	fs.SetOutput(stderr)
	days := fs.Int("days", 30, "window length in days")
	if fs.Parse(args) != nil || *days < 1 {
		fmt.Fprintln(stderr, "usage: flow-guard prune-candidates [-days N]  (N >= 1)")
		Reason(env, "usage")
		return 2
	}
	addr := "http://127.0.0.1:4173"
	for _, k := range []string{"FLOW_RECORDS_ADDR", "FLOW_ADDR"} {
		if v := env.Getenv(k); v != "" {
			addr = strings.TrimRight(v, "/")
			break
		}
	}
	to := time.Now().UTC()
	from := to.AddDate(0, 0, -*days)
	get := func(path string, into any) error {
		resp, err := (&http.Client{Timeout: 10 * time.Second}).Get(addr + path)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("GET %s: %s", path, resp.Status)
		}
		return json.NewDecoder(resp.Body).Decode(into)
	}

	var health struct {
		Rows []struct {
			Guard  string `json:"guard"`
			Runs   int    `json:"runs"`
			Fired  int    `json:"fired"`
			Helper bool   `json:"helper"`
		} `json:"rows"`
	}
	q := url.Values{"from": {from.Format(time.RFC3339)}, "to": {to.Format(time.RFC3339)}}
	var projects struct {
		Projects []string `json:"projects"`
	}
	err := get("/api/v1/stats/guard-health?"+q.Encode(), &health)
	if err == nil {
		err = get("/api/v1/projects", &projects)
	}
	type incident struct {
		Guard       string    `json:"guard"`
		Symptom     string    `json:"symptom"`
		MinutesLost int       `json:"minutesLost"`
		OccurredAt  time.Time `json:"occurredAt"`
	}
	var incidents []incident
	var loads struct {
		Rows []skillLoadRow `json:"rows"`
	}
	if err == nil {
		err = get("/api/v1/skill-loads?"+q.Encode(), &loads)
	}
	for _, p := range projects.Projects {
		if err != nil {
			break
		}
		var got []incident
		err = get("/api/v1/incidents/"+url.PathEscape(p), &got)
		incidents = append(incidents, got...)
	}
	if err != nil {
		fmt.Fprintf(stderr, "prune-candidates: cannot read flowd at %s: %v\n", addr, err)
		Reason(env, "flowd-unreadable")
		return 2
	}

	fmt.Fprintf(stdout, "window: %s .. %s\n", from.Format(time.RFC3339), to.Format(time.RFC3339))
	helper := func(g string) bool { return slices.Contains(records.GuardHelpers, g) }
	ran := map[string]bool{}
	var fired []string
	for _, r := range health.Rows {
		ran[r.Guard] = true
		if r.Fired == 0 && !r.Helper && !helper(r.Guard) {
			fired = append(fired, fmt.Sprintf("never-fired %s runs=%d", r.Guard, r.Runs))
		}
	}
	var never []string
	for g := range Registry {
		if !ran[g] && !helper(g) {
			never = append(never, "never-ran "+g)
		}
	}
	type tally struct {
		n, minutes int
		latest     incident
	}
	byGuard := map[string]*tally{}
	for _, in := range incidents {
		if in.OccurredAt.Before(from) {
			continue
		}
		t := byGuard[in.Guard]
		if t == nil {
			t = &tally{}
			byGuard[in.Guard] = t
		}
		t.n++
		t.minutes += in.MinutesLost
		if !in.OccurredAt.Before(t.latest.OccurredAt) {
			t.latest = in
		}
	}
	var friction []string
	for g, t := range byGuard {
		if t.n >= 2 {
			friction = append(friction, fmt.Sprintf("friction %s incidents=%d minutes-lost=%d latest: %s", g, t.n, t.minutes, strings.TrimSpace(strings.SplitN(t.latest.Symptom, "\n", 2)[0])))
		}
	}
	var unloaded []string
	root := env.Getenv("FLOW_GUARD_REPO_ROOT")
	if root != "" {
		unloaded = neverLoaded(root, loads.Rows)
	}
	// Load recording may be younger than the window: say which part the never-loaded lines cover.
	var first time.Time
	for _, r := range loads.Rows {
		if !r.First.IsZero() && (first.IsZero() || r.First.Before(first)) {
			first = r.First
		}
	}
	if root != "" && !first.IsZero() {
		fmt.Fprintf(stdout, "never-loaded window: %s .. %s\n", first.UTC().Format(time.RFC3339), to.Format(time.RFC3339))
	}
	for _, lines := range [][]string{never, fired, friction, unloaded} {
		sort.Strings(lines)
		for _, l := range lines {
			fmt.Fprintln(stdout, l)
		}
	}
	return 0
}

// skillLoadRow is one row of GET /api/v1/skill-loads: a repo-relative path and its loads in the window.
type skillLoadRow struct {
	Path  string    `json:"path"`
	Loads int       `json:"loads"`
	First time.Time `json:"first"`
}

// neverLoaded is the never-loaded lines for root's skills/, given the window's per-path loads.
func neverLoaded(root string, rows []skillLoadRow) []string {
	if len(rows) == 0 {
		return []string{"never-loaded: no load recorded in the window"}
	}
	loaded := map[string]bool{}
	for _, r := range rows {
		loaded[r.Path] = r.Loads > 0
	}
	cond := map[string][]string{}
	for _, set := range conditionalLoadSets(root) {
		for _, f := range set.files {
			cond[f] = append(cond[f], set.condition)
		}
	}
	var out []string
	_ = filepath.WalkDir(filepath.Join(root, "skills"), func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(p) != ".md" {
			return nil
		}
		rel, _ := filepath.Rel(root, p)
		if loaded[filepath.ToSlash(rel)] {
			return nil
		}
		line := "never-loaded " + filepath.ToSlash(rel)
		if c := cond[p]; len(c) > 0 {
			line += " conditional: " + strings.Join(c, "; ")
		}
		out = append(out, line)
		return nil
	})
	return out
}
