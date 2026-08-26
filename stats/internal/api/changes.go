package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/tweety53/agents/stats/internal/store"
)

// changeHandler serves the three change endpoints. It holds only what it
// needs -- the ChangeStore interface, never store.Store itself -- so a test
// can supply a fake with no database.
type changeHandler struct {
	store  ChangeStore
	logger *slog.Logger
}

// changeDTO is the wire representation of a store.Change: the JSON shape
// GET and PUT exchange. Field names match the state file contract's own
// vocabulary (design.md: "changes carries the current state file field for
// field, which is what makes the contract rewrite mechanical").
type changeDTO struct {
	ProjectKey     string          `json:"projectKey"`
	Name           string          `json:"name"`
	State          string          `json:"state"`
	Branch         *string         `json:"branch,omitempty"`
	Worktrees      json.RawMessage `json:"worktrees,omitempty"`
	ArtifactURL    *string         `json:"artifactUrl,omitempty"`
	JiraIssue      *string         `json:"jiraIssue,omitempty"`
	PlanningEffort *string         `json:"planningEffort,omitempty"`
	Models         json.RawMessage `json:"models,omitempty"`
	PRURL          *string         `json:"prUrl,omitempty"`
	Repos          []repoDTO       `json:"repos,omitempty"`
	// MainCheckoutPath bootstraps the project row when it does not yet
	// exist (store.Change.MainCheckoutPath); it is ignored by the store
	// once the project row is present.
	MainCheckoutPath string `json:"mainCheckoutPath,omitempty"`
	UpdatedAt        string `json:"updatedAt"`
	UpdatedBy        string `json:"updatedBy"`
}

type repoDTO struct {
	RepoRoot  string  `json:"repoRoot"`
	MergeBase *string `json:"mergeBase,omitempty"`
}

type listChangesResponse struct {
	Total   int         `json:"total"`
	Changes []changeDTO `json:"changes"`
}

// toDTO renders a store.Change as its wire shape.
func toDTO(c store.Change) changeDTO {
	dto := changeDTO{
		ProjectKey:     c.ProjectKey,
		Name:           c.Name,
		State:          string(c.State),
		Branch:         c.Branch,
		Worktrees:      c.Worktrees,
		ArtifactURL:    c.ArtifactURL,
		JiraIssue:      c.JiraIssue,
		PlanningEffort: c.PlanningEffort,
		Models:         c.Models,
		PRURL:          c.PRURL,
		UpdatedAt:      c.UpdatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedBy:      c.UpdatedBy,
	}
	for _, r := range c.Repos {
		dto.Repos = append(dto.Repos, repoDTO{RepoRoot: r.RepoRoot, MergeBase: r.MergeBase})
	}
	return dto
}

// toChange builds a store.Change from dto, taking the record's identity --
// project and name -- from the URL path rather than the request body. The
// path is canonical: a body naming a different projectKey or name would
// otherwise let a client silently write into a record its own URL did not
// name.
//
// c.Repos is deliberately left unset here: put derives it from c.Worktrees
// via reposFromWorktrees immediately before writing, rather than trusting
// dto.Repos. worktrees is already the state file contract's own record of
// every affected repository (per skills/myflow-contracts/state-file.md,
// "Multi-repo shape") -- a second, independently-supplied repos field on
// the wire would be a second copy of the same fact, free to disagree with
// the first. Deriving it here, in the one place PutChange is called, means
// every future caller of this endpoint (task 5's thin CLI included, which
// never populates repos at all) gets a correct repository set for free.
func (dto changeDTO) toChange(projectKey, name string) (store.Change, error) {
	c := store.Change{
		ProjectKey:       projectKey,
		MainCheckoutPath: dto.MainCheckoutPath,
		Name:             name,
		State:            store.State(dto.State),
		Branch:           dto.Branch,
		Worktrees:        dto.Worktrees,
		ArtifactURL:      dto.ArtifactURL,
		JiraIssue:        dto.JiraIssue,
		PlanningEffort:   dto.PlanningEffort,
		Models:           dto.Models,
		PRURL:            dto.PRURL,
		UpdatedBy:        dto.UpdatedBy,
	}
	if dto.UpdatedAt != "" {
		t, err := time.Parse(time.RFC3339, dto.UpdatedAt)
		if err != nil {
			return store.Change{}, fmt.Errorf("updatedAt: %q is not RFC 3339: %w", dto.UpdatedAt, err)
		}
		c.UpdatedAt = t
	}
	return c, nil
}

// reposFromWorktrees derives the repository set PutChange writes
// transactionally alongside the change row (task 2.1) from worktrees --
// the state file contract's own `{"<absolute worktree path>": "<merge
// base or null>"}` map. Each entry becomes one store.Repo, keyed by the
// worktree path exactly as change_repos already keys a multi-repo change's
// rows (design.md's data model; store.Change.Repos' own doc comment).
//
// The result is sorted by RepoRoot so two calls with the same worktrees
// content always produce the same Repos order, regardless of Go's
// randomised map iteration -- callers (including this package's own
// tests) can assert on order without that assertion being incidentally
// flaky.
//
// An empty or absent worktrees map is not an error: it derives an empty
// repo set, exactly as a change confined to its own project's repository
// should.
func reposFromWorktrees(worktrees json.RawMessage) ([]store.Repo, error) {
	if len(worktrees) == 0 {
		return nil, nil
	}
	var m map[string]*string
	if err := json.Unmarshal(worktrees, &m); err != nil {
		return nil, fmt.Errorf("worktrees: %w", err)
	}
	if len(m) == 0 {
		return nil, nil
	}
	repos := make([]store.Repo, 0, len(m))
	for root, mergeBase := range m {
		repos = append(repos, store.Repo{RepoRoot: root, MergeBase: mergeBase})
	}
	sort.Slice(repos, func(i, j int) bool { return repos[i].RepoRoot < repos[j].RepoRoot })
	return repos, nil
}

// DecodeChangeBody decodes body -- the wire JSON shape a PUT request and a
// task 6 journal entry both carry (a journal entry's Body is, by
// construction, exactly the bytes `state set` sent as a PUT body -- see
// fallback.Entry's own doc comment) -- into a store.Change for
// project/name, deriving its repository set from the decoded worktrees
// field via reposFromWorktrees.
//
// This is the single decode path both put and internal/reconcile's journal
// replay use, so the two can never silently drift on what counts as a
// well-formed change body -- a body the live PUT handler would accept is
// exactly a body replay accepts, and vice versa.
func DecodeChangeBody(project, name string, body []byte) (store.Change, error) {
	var dto changeDTO
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&dto); err != nil {
		return store.Change{}, fmt.Errorf("malformed change body: %w", err)
	}

	c, err := dto.toChange(project, name)
	if err != nil {
		return store.Change{}, err
	}

	c.Repos, err = reposFromWorktrees(c.Worktrees)
	if err != nil {
		return store.Change{}, err
	}
	return c, nil
}

// IsDefinitiveChangeOutcome reports whether err represents a *definitive*
// answer to a change write -- success, or a refusal that will not change on
// retry -- as opposed to a transient failure (a dropped connection, a
// context cancellation, an unmapped store error) that a later replay might
// resolve differently. It is the change-record counterpart of
// IsDefinitiveMarkOutcome (stages.go), which internal/reconcile's stage-mark
// replay already uses for exactly this purpose; state replay is what this
// function now gives the same treatment.
//
// store.ErrMonotonicViolation is definitive because it covers a genuinely
// superseded write and a benign duplicate retry alike -- both leave the
// record correct, with nothing left for that entry to do (see
// internal/api/changes.go's put doc comment). store.ErrInvalidState,
// store.ErrInvalidMainCheckoutPath and store.ErrInvalidMergeBase are
// definitive for the same reason a journalled entry naming an undocumented
// stage is: the entry's own content is what is wrong, and replaying it
// again produces the identical refusal every time. store.ErrInvalidMergeBase
// reaches this path only from a hand-edited or out-of-band-modified fallback
// file -- `flow state set` refuses a malformed merge base before it can be
// journalled at all -- and that is exactly the entry that would otherwise
// block every entry queued behind it forever. store.ErrDuplicateRepoRoot is
// included for the same symmetry -- a payload naming one repo root twice
// would be equally unfixable by retrying -- though nothing on this path can
// actually produce it today: DecodeChangeBody derives Repos from the
// worktrees JSON object's own keys
// (reposFromWorktrees), and Go map keys are already unique once unmarshalled,
// so two identical repo roots can never reach PutChange this way.
//
// A body that fails to decode at all -- DecodeChangeBody itself erroring,
// including the closed-schema DisallowUnknownFields case -- is definitive
// for the identical reason: the bytes already in the journal will never
// decode differently on a later replay, so a caller checking a decode error
// alongside this function's result must treat that failure the same way.
func IsDefinitiveChangeOutcome(err error) bool {
	if err == nil {
		return true
	}
	switch {
	case errors.Is(err, store.ErrMonotonicViolation):
		return true
	case errors.Is(err, store.ErrInvalidState):
		return true
	case errors.Is(err, store.ErrInvalidMainCheckoutPath):
		return true
	case errors.Is(err, store.ErrInvalidMergeBase):
		return true
	case errors.Is(err, store.ErrDuplicateRepoRoot):
		return true
	default:
		return false
	}
}

// get serves GET /api/v1/changes/{project}/{name}.
func (h *changeHandler) get(w http.ResponseWriter, r *http.Request) {
	project := r.PathValue("project")
	name := r.PathValue("name")

	c, err := h.store.GetChange(r.Context(), project, name)
	if err != nil {
		status, msg := mapStoreError(h.logger, fmt.Sprintf("get change %s/%s", project, name), err)
		writeError(w, status, msg)
		return
	}
	writeJSON(w, http.StatusOK, toDTO(c))
}

// put serves PUT /api/v1/changes/{project}/{name}. It sends the whole
// record to the store -- store.Change.PutChange already renders every
// field, so a field this handler leaves at its zero value overwrites
// whatever was stored, exactly as the state file contract's whole-object
// write rule requires.
//
// For task 6: a client-side timeout on this call is not knowable to the
// client -- the write may have landed in the store before the response
// was lost, so retrying, or replaying the same entry from the journal, has
// to be safe. It is, but not because the retry silently re-applies:
// store.PutChange's guard (see its own doc comment) orders writes by
// state and, at equal state, strictly by UpdatedAt, so a *second*
// application of the exact same Change -- same state, same UpdatedAt -- is
// itself refused with ErrMonotonicViolation, exactly as a write superseded
// by someone else's later change would be. That refusal is the safe
// outcome, not a failure: the record is already correct, nothing was lost,
// and there is nothing left to apply.
//
// ErrMonotonicViolation therefore covers two situations task 6's
// implementer cannot tell apart from the error alone, and must not need
// to:
//   - a genuinely superseded write, where some other, later write already
//     moved the record forward;
//   - a benign duplicate, where this exact entry already landed and is
//     being retried or replayed.
//
// Both are safe for reconciliation to retire the journal entry over: an
// entry is removed once the store has accepted *or explicitly refused*
// it (design.md, "Availability and reconciliation"), and this is why one
// sentinel for both cases is the right design rather than a gap -- neither
// case is an error reconciliation needs to act on beyond retiring the
// entry.

func (h *changeHandler) put(w http.ResponseWriter, r *http.Request) {
	project := r.PathValue("project")
	name := r.PathValue("name")

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "read request body: "+err.Error())
		return
	}

	c, err := DecodeChangeBody(project, name, body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "malformed request body: "+err.Error())
		return
	}

	if err := h.store.PutChange(r.Context(), c); err != nil {
		status, msg := mapStoreError(h.logger, fmt.Sprintf("put change %s/%s", project, name), err)
		writeError(w, status, msg)
		return
	}
	writeJSON(w, http.StatusOK, toDTO(c))
}

// list serves GET /api/v1/changes?project=&state=&... -- every filter,
// search, sort and page parameter is carried through a store.Query built
// by parseChangeQuery and handed to QueryChanges, never assembled into SQL
// here. A field parseChangeQuery cannot resolve against store's own
// allowlist surfaces as ErrUnknownField, mapped below to 400, naming the
// field and the accepted alternatives -- never a 500, and never a
// silently ignored parameter.
func (h *changeHandler) list(w http.ResponseWriter, r *http.Request) {
	q, err := parseChangeQuery(r.Context(), h.store, r)
	if err != nil {
		writeParseOrStoreError(w, h.logger, "resolve project for changes list", err)
		return
	}

	changes, total, err := h.store.QueryChanges(r.Context(), q)
	if err != nil {
		status, msg := mapStoreError(h.logger, "list changes", err)
		writeError(w, status, msg)
		return
	}

	dtos := make([]changeDTO, len(changes))
	for i, c := range changes {
		dtos[i] = toDTO(c)
	}
	writeJSON(w, http.StatusOK, listChangesResponse{Total: total, Changes: dtos})
}

// reservedQueryParams are the query parameters list interprets itself.
// Every other parameter present is treated as an equality filter on that
// field name -- resolved, or rejected, by store.QueryChanges' own
// allowlist. This handler never decides which field names are valid; it
// only shapes the request into a store.Query and lets the store answer
// that question, per query.go's "the entire dynamic-SQL surface" design.
var reservedQueryParams = map[string]bool{
	"q":      true,
	"sort":   true,
	"limit":  true,
	"offset": true,
}

// parseChangeQuery turns r's query string into a store.Query: "q" becomes
// the free-text search term, "sort" a comma-separated list of fields (a
// "-" prefix means descending), "limit"/"offset" the page, and every other
// parameter an equality filter. Filters are sorted by field name so the
// resulting Query is deterministic regardless of net/http's unordered
// url.Values map -- callers that inspect q.Filters (including this
// package's own tests) see a stable order.
//
// The "project" filter, if present, is resolved through
// resolveProjectParam (stats.go) before it becomes a Filter -- the same
// display-name rule every stats view applies to its own "project"
// parameter, applied here at this endpoint's one parse site rather than
// reimplemented.
func parseChangeQuery(ctx context.Context, resolver projectResolver, r *http.Request) (store.Query, error) {
	values := r.URL.Query()
	q := store.Query{Search: values.Get("q")}

	if sortParam := values.Get("sort"); sortParam != "" {
		for _, field := range strings.Split(sortParam, ",") {
			field = strings.TrimSpace(field)
			if field == "" {
				continue
			}
			desc := false
			if strings.HasPrefix(field, "-") {
				desc = true
				field = field[1:]
			}
			q.Sort = append(q.Sort, store.SortKey{Field: field, Desc: desc})
		}
	}

	if v := values.Get("limit"); v != "" {
		// parseLimit (stats.go) rejects a negative value -- including
		// store.NoLimit's own -1 -- which a bare strconv.Atoi would not:
		// see its own doc comment (post-commit review finding F3, task 11)
		// for why this endpoint had the same gap as the stage-runs listing.
		n, err := parseLimit(v)
		if err != nil {
			return store.Query{}, err
		}
		q.Limit = n
	}
	if v := values.Get("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return store.Query{}, fmt.Errorf("invalid offset %q: must be an integer", v)
		}
		q.Offset = n
	}

	for field, vals := range values {
		if reservedQueryParams[field] || len(vals) == 0 || vals[0] == "" {
			continue
		}
		value := vals[0]
		if field == "project" {
			resolved, err := resolveProjectParam(ctx, resolver, value)
			if err != nil {
				return store.Query{}, err
			}
			value = resolved
		}
		q.Filters = append(q.Filters, store.Filter{Field: field, Op: store.OpEq, Value: value})
	}
	sort.Slice(q.Filters, func(i, j int) bool { return q.Filters[i].Field < q.Filters[j].Field })

	return q, nil
}
