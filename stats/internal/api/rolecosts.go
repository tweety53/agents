package api

import (
	"context"
	"log/slog"
	"net/http"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/tweety53/agents/stats/internal/modelpolicy"
	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/store"
)

// RoleCostsStore backs GET /api/v1/stats/role-costs.
type RoleCostsStore interface {
	projectResolver
	RoleDispatches(ctx context.Context, period store.Period, project *string) ([]store.RoleDispatch, error)
}

// WithRoleCosts registers GET /api/v1/stats/role-costs against s; the role
// groups come from the table WithPolicy passes. The view ignores dispatches
// started before since (the zero time keeps every one): the data under it is
// shared with other views, so this resets only this one.
func WithRoleCosts(s RoleCostsStore, since time.Time) Option {
	return func(o *serverOptions) { o.roleCosts, o.roleCostsSince = s, since }
}

type roleCostsHandler struct {
	store RoleCostsStore
	since time.Time
	table []byte
	log   *slog.Logger
}

// rolePair is one model · effort pair a role ran on.
type rolePair struct {
	Pair       string  `json:"pair"`
	CostUSD    float64 `json:"costUSD"`
	Dispatches int     `json:"dispatches"`
}

// roleCost is one allowed-pairs role's dispatches in the period. Priced
// counts the dispatches carrying a cost; AgentSeconds sums every ended
// dispatch's span, parallel agents adding up; Failed counts the failed,
// aborted and blocked ones.
type roleCost struct {
	Role          string     `json:"role"`
	Group         string     `json:"group"`
	Dispatches    int        `json:"dispatches"`
	Priced        int        `json:"priced"`
	Failed        int        `json:"failed"`
	CostUSD       float64    `json:"costUSD"`
	AgentSeconds  float64    `json:"agentSeconds"`
	MedianSeconds float64    `json:"medianSeconds"`
	Pairs         []rolePair `json:"pairs"`
}

// roleWeek is one week's spend per role, the week starting on the
// Wednesday the weekly-costs view starts it on.
type roleWeek struct {
	Week  string             `json:"week"`
	Roles map[string]float64 `json:"roles"`
}

type roleCostsResponse struct {
	Roles []roleCost `json:"roles"`
	Weeks []roleWeek `json:"weeks"`
}

var (
	reviewerTail = regexp.MustCompile(`-review-(\d+)$`)
	panelRound   = regexp.MustCompile(`^panel-(\d+)-`)
)

// RoleOf names the allowed-pairs role (skills/flow-contracts/model-policy.md)
// a dispatch ran as, from its stored role and dispatch key; a dispatch no
// rule places is "other". A guard run's headless call (stored role
// "headless") is the role its key names.
func RoleOf(stored, key string) string {
	if stored == "headless" { // a guard's headless model call: the key is its role
		return key
	}
	if stored != "" && !records.IsLegacyRole(stored) { // already a roles-table role
		return stored
	}
	if strings.HasPrefix(key, "pipeline-fix-") || strings.HasPrefix(key, "self-review-") {
		switch {
		case strings.HasSuffix(key, "-plan") || stored == "planner":
			return "loop-plan"
		case strings.HasSuffix(key, "-pass"):
			return "self-review-pass"
		case reviewerTail.MatchString(key):
			if reviewerTail.FindStringSubmatch(key)[1] == "1" {
				return "loop-review-1"
			}
			return "loop-review-r"
		default:
			return "loop-fix"
		}
	}
	switch {
	case strings.HasPrefix(key, "operator-"):
		return "operator-asked"
	case strings.HasPrefix(key, "minors"):
		return "minor-fix"
	case strings.HasPrefix(key, "full-suite-fix"):
		return "panel-fix"
	case strings.HasPrefix(key, "runner-"):
		return "runner"
	case strings.HasPrefix(key, "plan-review"):
		return "decider"
	}
	switch stored {
	case "implementer":
		return "implementer"
	case "explorer":
		return "explorer"
	case "merger":
		return "panel-merger"
	case "panel-fix":
		return "panel-fix"
	case "verifier":
		if key == "verify" || strings.HasPrefix(key, "verify-") { // run-verify-list's inline verify rows
			return "runner"
		}
		return "visual-verifier"
	case "planner":
		if strings.HasPrefix(key, "design-") {
			return "plan-designer"
		}
		if strings.HasPrefix(key, "designer-") {
			return "designer"
		}
		return "planner"
	case "reviewer":
		if strings.HasPrefix(key, "task-") { // the retired per-task review's, folded into the panel's roles
			if strings.Contains(key, "-reviewer-fix") {
				return "panel-rerun"
			}
			return "panel-reviewer"
		}
		if m := panelRound.FindStringSubmatch(key); m != nil {
			if m[1] == "0" { // round 0 is the initial panel, 1..n fix rounds
				return "panel-reviewer"
			}
			return "panel-rerun"
		}
	}
	return "other"
}

// get serves GET /api/v1/stats/role-costs: every role that ran in the
// period, grouped by pipeline stage in the roles table's order, costliest
// first within a stage, and the weekly spend per role.
func (h *roleCostsHandler) get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	period, project, err := parsePeriodAndProject(ctx, h.store, r.URL.Query())
	if err != nil {
		writeParseOrStoreError(w, h.log, "resolve project for role-costs", err)
		return
	}
	if period.From.Before(h.since) {
		period.From = h.since
		if period.To.Before(period.From) {
			period.To = period.From
		}
	}
	rows, err := h.store.RoleDispatches(ctx, period, project)
	if err != nil {
		status, msg := mapStoreError(h.log, "list dispatches for role-costs", err)
		writeError(w, status, msg)
		return
	}
	groups := map[string]string{}
	var order []string
	if entries, err := modelpolicy.Rows(h.table); err == nil {
		for _, e := range entries {
			groups[e.Role] = e.Group
			if !slices.Contains(order, e.Group) {
				order = append(order, e.Group)
			}
		}
	}
	writeJSON(w, http.StatusOK, foldRoleCosts(rows, groups, append(order, "Other")))
}

// foldRoleCosts folds rows per role, the roles in order's stage order and
// costliest first within a stage.
func foldRoleCosts(rows []store.RoleDispatch, groups map[string]string, order []string) roleCostsResponse {
	byRole := map[string]*roleCost{}
	spans := map[string][]float64{}
	weeks := map[string]map[string]float64{}
	for _, d := range rows {
		role := d.PolicyRole
		if role == "" {
			role = RoleOf(d.Role, d.Key)
		}
		c := byRole[role]
		if c == nil {
			g := groups[role]
			if g == "" {
				g = "Other"
			}
			c = &roleCost{Role: role, Group: g, Pairs: []rolePair{}}
			byRole[role] = c
		}
		c.Dispatches++
		if d.Outcome == "failed" || d.Outcome == "aborted" || d.Outcome == "blocked" {
			c.Failed++
		}
		if d.EndedAt != nil {
			s := d.EndedAt.Sub(d.StartedAt).Seconds()
			c.AgentSeconds += s
			spans[role] = append(spans[role], s)
		}
		pair := d.Model + " · " + d.Effort
		i := slices.IndexFunc(c.Pairs, func(p rolePair) bool { return p.Pair == pair })
		if i < 0 {
			c.Pairs = append(c.Pairs, rolePair{Pair: pair})
			i = len(c.Pairs) - 1
		}
		c.Pairs[i].Dispatches++
		if d.CostUSD != nil {
			c.Priced++
			c.CostUSD += *d.CostUSD
			c.Pairs[i].CostUSD += *d.CostUSD
			week := d.StartedAt.UTC().Add(-48 * time.Hour)
			week = week.AddDate(0, 0, -int(week.Weekday()+6)%7).Add(48 * time.Hour).Truncate(24 * time.Hour)
			k := week.Format(time.DateOnly)
			if weeks[k] == nil {
				weeks[k] = map[string]float64{}
			}
			weeks[k][role] += *d.CostUSD
		}
	}
	out := roleCostsResponse{Roles: []roleCost{}, Weeks: []roleWeek{}}
	for role, c := range byRole {
		if s := spans[role]; len(s) > 0 {
			sort.Float64s(s)
			if n := len(s); n%2 == 1 {
				c.MedianSeconds = s[n/2]
			} else {
				c.MedianSeconds = (s[n/2-1] + s[n/2]) / 2
			}
		}
		sort.SliceStable(c.Pairs, func(i, j int) bool { return c.Pairs[i].CostUSD > c.Pairs[j].CostUSD })
		out.Roles = append(out.Roles, *c)
	}
	sort.Slice(out.Roles, func(i, j int) bool {
		if gi, gj := slices.Index(order, out.Roles[i].Group), slices.Index(order, out.Roles[j].Group); gi != gj {
			return gi < gj
		}
		if out.Roles[i].CostUSD != out.Roles[j].CostUSD {
			return out.Roles[i].CostUSD > out.Roles[j].CostUSD
		}
		return out.Roles[i].Role < out.Roles[j].Role
	})
	for k, m := range weeks {
		out.Weeks = append(out.Weeks, roleWeek{Week: k, Roles: m})
	}
	sort.Slice(out.Weeks, func(i, j int) bool { return out.Weeks[i].Week < out.Weeks[j].Week })
	return out
}
