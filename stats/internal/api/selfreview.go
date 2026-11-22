package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/selfreview"
	"github.com/tweety53/agents/stats/internal/stages"
	"github.com/tweety53/agents/stats/internal/store"
)

// selfreviewStore is the store dependency the self-review bundle endpoint
// needs, defined here at the consumer per go-interface-design — exactly the
// reads the handler calls, so a test needs no database. The worktree the
// change-directory sources come from is not a store answer: change_repos
// carries no roots for the pipeline's real changes, and the caller —
// running inside the change's worktree — resolves it and passes it as the
// repo query parameter.
type selfreviewStore interface {
	RunRecord(ctx context.Context, projectKey, change string) (records.Run, error)

	// ChangeSummary is the change's recorded summary read — the row the
	// bundle serves as its first source. ErrChangeNotFound from it is the
	// missing-source case the bundle reports skipped, never a failure.
	ChangeSummary(ctx context.Context, projectKey, change string) (records.ChangeSummary, error)

	// StageCompleted reports whether a flow.review stage run for
	// the change completed in this store — the fact that separates a
	// records-source loss from a panel that never ran, which the bundle
	// names loudly when the run holds no dispatch rows (KAN-621).
	StageCompleted(ctx context.Context, projectKey, change, stage string) (bool, error)

	// ListRequests is a /flow-interactive session's requests (KAN-1061),
	// the bundle's git-log source for a change that has any.
	ListRequests(ctx context.Context, projectKey, change string) ([]records.Request, error)

	// RecordSelfReviewFinding and ListSelfReviewFindings back the findings
	// route: one row per self-review finding and its outcome (KAN-875).
	RecordSelfReviewFinding(ctx context.Context, projectKey string, in records.SelfReviewFinding) (records.SelfReviewFinding, error)
	ListSelfReviewFindings(ctx context.Context, projectKey, change string) ([]records.SelfReviewFinding, error)
	// GetSelfReviewFinding backs the by-id read the flow CLI's finding-get makes.
	GetSelfReviewFinding(ctx context.Context, projectKey string, id int64) (records.SelfReviewFinding, error)

	// ReviewSelfReviewFinding backs the review route: the operator's approve or decline of a suggestion.
	ReviewSelfReviewFinding(ctx context.Context, projectKey string, id int64, status, reason string) error
}

// selfreviewHandler serves GET
// /api/v1/self-review/{project}/{change}/bundle?repo=<abs-path>: the whole
// self-review context bundle, assembled server-side — the ledger and panel
// record rendered from the store exactly as the render route renders them,
// the change directory's tasks.md, design.md, narrative.md,
// visual-verification.md and live-verification.md read from disk in the
// worktree the caller named — gitignored, never committed (KAN-1073) — and
// the git log of its first-parent merge-base..HEAD range, derived here
// rather than in any Bash the caller would have to run. The CLI transports the result and
// constructs none of it, the record-render rule.
type selfreviewHandler struct {
	store  selfreviewStore
	git    selfreview.Runner
	build  string
	logger *slog.Logger
}

// bundle answers with text/markdown. repo is required: without it the
// route could only report the change-directory sources skipped, a false
// fact about the change rather than the caller mistake it is — a 400 says
// so plainly. A change the store has
// never heard of is not an error — its ledger and panel sections report
// skipped and the rest of the bundle still serves, the gather's own "a
// missing source is never fatal" rule. Only a store read that fails for a
// real reason is a 5xx.
func (h *selfreviewHandler) bundle(w http.ResponseWriter, r *http.Request) {
	project, change := r.PathValue("project"), r.PathValue("change")
	repo := r.URL.Query().Get("repo")
	if repo == "" {
		writeError(w, http.StatusBadRequest, "repo is required: the absolute path of the worktree holding the change")
		return
	}
	if !filepath.IsAbs(repo) {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("repo %q is not an absolute path", repo))
		return
	}

	rec, err := h.store.RunRecord(r.Context(), project, change)
	if err != nil {
		if !errors.Is(err, store.ErrChangeNotFound) {
			status, msg := mapStoreError(h.logger, fmt.Sprintf("read the run record for %s/%s", project, change), err)
			writeError(w, status, msg)
			return
		}
		rec = records.Run{Change: change}
	}

	// The recorded summary is the bundle's first source, and its read
	// follows the same rule the run record's does: ErrChangeNotFound is
	// the change having no summary — reported skipped inside the bundle —
	// while any other read failure is a 5xx, never a bundle a caller
	// could mistake for the change's own.
	summary, summaryFound := "", false
	cs, err := h.store.ChangeSummary(r.Context(), project, change)
	switch {
	case err == nil:
		summary, summaryFound = cs.Summary, true
	case !errors.Is(err, store.ErrChangeNotFound):
		status, msg := mapStoreError(h.logger, fmt.Sprintf("read the change summary for %s/%s", project, change), err)
		writeError(w, status, msg)
		return
	}

	// The stage-completed read follows the same failure rule the two
	// reads above carry: any non-nil error is a 5xx, never a degraded
	// bundle. A completed flow.review, or any retired key stages.Retired maps
	// to it (flow.review-panel included), proves a panel ran.
	panelKeys := []string{"flow.review"}
	for old, now := range stages.Retired {
		if now == "flow.review" {
			panelKeys = append(panelKeys, old)
		}
	}
	panelRan := false
	for _, key := range panelKeys {
		ran, err := h.store.StageCompleted(r.Context(), project, change, key)
		if err != nil {
			status, msg := mapStoreError(h.logger, fmt.Sprintf("read the review stage runs for %s/%s", project, change), err)
			writeError(w, status, msg)
			return
		}
		if ran {
			panelRan = true
			break
		}
	}

	// The real runner gets the bundle's overall git budget, so a slow
	// repository degrades the bundle rather than outrunning writeTimeout.
	git := h.git
	if er, ok := git.(selfreview.ExecRunner); ok {
		er.Deadline = time.Now().Add(selfreview.BundleBudget)
		git = er
	}
	// The requests read degrades rather than fails: a session's requests are
	// one git-log source among many, and without them the bundle still
	// serves the change branch's own log.
	requests, err := h.store.ListRequests(r.Context(), project, change)
	if err != nil {
		if h.logger != nil {
			h.logger.Warn("api: self-review bundle: list requests failed", "project", project, "change", change, "error", err)
		}
		requests = nil
	}
	bundle, err := selfreview.Bundle(change, rec, summary, summaryFound, panelRan, h.build, []string{repo}, requests, git)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(bundle))
}

// recordFinding serves POST /api/v1/self-review/{project}/{change}/findings:
// one self-review finding and its outcome. The change comes from the path,
// never the body. A store validation refusal is 400 through mapStoreError.
func (h *selfreviewHandler) recordFinding(w http.ResponseWriter, r *http.Request) {
	project, change := r.PathValue("project"), r.PathValue("change")

	var in records.SelfReviewFinding
	if err := decodeJSONBody(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	in.Change = change

	out, err := h.store.RecordSelfReviewFinding(r.Context(), project, in)
	if err != nil {
		status, msg := mapStoreError(h.logger, fmt.Sprintf("record self-review finding for %s/%s", project, change), err)
		writeError(w, status, msg)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

// listFindings serves GET /api/v1/self-review/{project}/{change}/findings:
// the change's self-review findings in the order they were recorded.
func (h *selfreviewHandler) listFindings(w http.ResponseWriter, r *http.Request) {
	project, change := r.PathValue("project"), r.PathValue("change")

	out, err := h.store.ListSelfReviewFindings(r.Context(), project, change)
	if err != nil {
		status, msg := mapStoreError(h.logger, fmt.Sprintf("list self-review findings for %s/%s", project, change), err)
		writeError(w, status, msg)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// getFinding serves GET /api/v1/self-review-finding/{project}/{id} (a root of its own: under
// /self-review/{project}/ it would clash with the {change}/bundle route): one finding by its id.
func (h *selfreviewHandler) getFinding(w http.ResponseWriter, r *http.Request) {
	project := r.PathValue("project")
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "finding id must be a number")
		return
	}
	out, err := h.store.GetSelfReviewFinding(r.Context(), project, id)
	if err != nil {
		status, msg := mapStoreError(h.logger, fmt.Sprintf("read self-review finding %d", id), err)
		writeError(w, status, msg)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// reviewFinding serves PATCH /api/v1/self-review/{project}/findings/{id}/review: the operator's
// to-review, approved or declined on a suggested finding. A row that is no suggestion is 404.
func (h *selfreviewHandler) reviewFinding(w http.ResponseWriter, r *http.Request) {
	project := r.PathValue("project")
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "finding id must be a number")
		return
	}
	var in struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	}
	if err := decodeJSONBody(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.store.ReviewSelfReviewFinding(r.Context(), project, id, in.Status, in.Reason); err != nil {
		status, msg := mapStoreError(h.logger, fmt.Sprintf("review self-review finding %d", id), err)
		writeError(w, status, msg)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// selfReviewPageMax caps one page; the dashboard asks for far fewer.
const selfReviewPageMax = 100

type selfReviewPageDTO struct {
	Total    int                `json:"total"`
	ToReview int                `json:"toReview"`
	Approved int                `json:"approved"`
	Declined int                `json:"declined"`
	Limit    int                `json:"limit"`
	Offset   int                `json:"offset"`
	Rows     []selfReviewRowDTO `json:"rows"`
}

// selfReviewPage serves GET /api/v1/stats/self-review-findings?from&to[&project]&show=fixed|to-review
// [&status=to-review|approved|declined][&limit][&offset]: one page of the dashboard's list, with
// the list's total and the suggestion counts by status. The dashboard never loads every row.
func (h *statsHandler) selfReviewPage(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()
	for param := range values {
		switch param {
		case "from", "to", "project", "show", "status", "limit", "offset":
		default:
			writeError(w, http.StatusBadRequest, fmt.Sprintf("unrecognised query parameter %q", param))
			return
		}
	}
	period, project, err := parsePeriodAndProject(r.Context(), h.store, values)
	if err != nil {
		writeParseOrStoreError(w, h.logger, "resolve project for self-review findings", err)
		return
	}
	limit, offset := 25, 0
	for name, dst := range map[string]*int{"limit": &limit, "offset": &offset} {
		if raw := firstValue(values, name); raw != "" {
			n, err := strconv.Atoi(raw)
			if err != nil || n < 0 || (name == "limit" && (n == 0 || n > selfReviewPageMax)) {
				writeError(w, http.StatusBadRequest, fmt.Sprintf("%s must be a whole number (limit 1-%d, offset 0 or more), got %q", name, selfReviewPageMax, raw))
				return
			}
			*dst = n
		}
	}
	show, status := firstValue(values, "show"), firstValue(values, "status")
	if show == "to-review" && status == "" {
		status = "to-review"
	}
	page, err := h.store.SelfReviewFindingsPage(r.Context(), period, project, show, status, limit, offset)
	if err != nil {
		s, m := mapStoreError(h.logger, "self-review findings page", err)
		writeError(w, s, m)
		return
	}
	writeJSON(w, http.StatusOK, selfReviewPageDTO{
		Total: page.Total, ToReview: page.ToReview, Approved: page.Approved, Declined: page.Declined,
		Limit: limit, Offset: offset, Rows: toSelfReviewDTOs(page.Rows, h.commitBase),
	})
}
