package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/selfreview"
	"github.com/tweety53/agents/stats/internal/store"
	"testing"
)

// selfReviewPath is the bundle route's path for one project/change; repo
// rides as the query parameter the caller resolves from its own location
// inside the repository.
func selfReviewPath(project, change, repo string) string {
	p := "/api/v1/self-review/" + project + "/" + change + "/bundle"
	if repo != "" {
		p += "?repo=" + repo
	}
	return p
}

// TestSelfReviewBundleHandlerServesAssembledBundle drives the route end to
// end over a real repository: the ledger renders from the recorded
// dispatches, and the archived design.md is read out of the change branch
// spectre/<name> of the repository the request named — no path arrives
// from the store.
func TestSelfReviewBundleHandlerServesAssembledBundle(t *testing.T) {
	repo := bundleGitRepo(t)
	bundleCommit(t, repo, "spectre/changes/kan-1/tasks.md", "- [ ] 1. do it\n",
		"chore(spectre): plan and session records")
	bundleArchiveBranch(t, repo, "kan-1", map[string]string{
		"tasks.md":  "- [ ] 1. do it\n",
		"design.md": "# kan-1 design\n",
	})

	ts, _ := recordTestServer(t, "proj", "kan-1")
	if resp, body := postJSON(t, ts.URL+recordsPath("proj", "kan-1")+"/dispatches", dispatchBody("implementer", "opus")); resp.StatusCode != http.StatusCreated {
		t.Fatalf("seed POST dispatches = %d (%s), want 201", resp.StatusCode, body)
	}

	resp, err := http.Get(ts.URL + selfReviewPath("proj", "kan-1", repo))
	if err != nil {
		t.Fatalf("GET bundle: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET bundle = %d, want 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/markdown") {
		t.Errorf("Content-Type = %q, want text/markdown", ct)
	}
	raw, _ := io.ReadAll(resp.Body)
	body := string(raw)

	for _, want := range []string{
		"# Self-review context bundle for kan-1",
		"## .superpowers/sdd/ledgers/kan-1.md",
		"## spectre/changes/archive/kan-1/design.md",
		"# kan-1 design",
		"## git log --stat",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("bundle missing %q:\n%s", want, body)
		}
	}
}

// TestSelfReviewBundleHandlerUnknownChange carries the gather's "a missing
// source is never fatal" rule through the route: a change the store has
// never heard of still serves a bundle whose store-side sections report
// skipped — ledger and panel both, never an empty panel record nobody
// wrote — not a 404 a caller could read as the endpoint refusing.
func TestSelfReviewBundleHandlerUnknownChange(t *testing.T) {
	ts, _ := recordTestServer(t, "proj", "kan-1")

	code, body := doGet(t, ts, selfReviewPath("proj", "never-heard", t.TempDir()))
	if code != http.StatusOK {
		t.Fatalf("GET bundle = %d (%s), want 200", code, body)
	}
	if !strings.Contains(body, "skipped: .superpowers/sdd/ledgers/never-heard.md (absent)") {
		t.Errorf("unknown change's ledger not reported skipped:\n%s", body)
	}
	if !strings.Contains(body, "skipped: .superpowers/sdd/reviews/never-heard-panel.md (absent)") {
		t.Errorf("unknown change's panel not reported skipped:\n%s", body)
	}
}

// TestSelfReviewBundleHandlerRefusesRelativeRepo pins the one caller
// mistake the route itself can see: a repo parameter that is not an
// absolute path is a 400, before anything reads or renders.
func TestSelfReviewBundleHandlerRefusesRelativeRepo(t *testing.T) {
	ts, _ := recordTestServer(t, "proj", "kan-1")

	code, body := doGet(t, ts, selfReviewPath("proj", "kan-1", "relative/path"))
	if code != http.StatusBadRequest {
		t.Fatalf("GET bundle = %d (%s), want 400", code, body)
	}
}

// TestSelfReviewBundleHandlerStoreFailure keeps the one failure the bundle
// cannot absorb: a store read that fails for a real reason is a 5xx, never
// a bundle a caller could mistake for the change's own.
func TestSelfReviewBundleHandlerStoreFailure(t *testing.T) {
	ts, fs := recordTestServer(t, "proj", "kan-1")
	fs.runRecordErr = errors.New("store exploded")

	code, body := doGet(t, ts, selfReviewPath("proj", "kan-1", t.TempDir()))
	if code != http.StatusInternalServerError {
		t.Fatalf("GET bundle = %d (%s), want 500", code, body)
	}
}

// --- repository fixtures, the selfreview package's own in miniature ---

func bundleGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	bundleRunGit(t, dir, "init", "-b", "main")
	bundleRunGit(t, dir, "config", "user.email", "test@example.com")
	bundleRunGit(t, dir, "config", "user.name", "Test")
	bundleCommit(t, dir, "README.md", "repo\n", "initial")
	return dir
}

func bundleArchiveBranch(t *testing.T, repo, name string, files map[string]string) {
	t.Helper()
	for file, body := range files {
		path := filepath.Join(repo, "spectre/changes/archive", name, file)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	bundleRunGit(t, repo, "checkout", "-b", "spectre/"+name)
	bundleRunGit(t, repo, "add", "-A")
	bundleRunGit(t, repo, "commit", "-m", "chore(spectre): archive "+name)
	bundleRunGit(t, repo, "checkout", "main")
}

func bundleCommit(t *testing.T, repo, file, body, subject string) {
	t.Helper()
	path := filepath.Join(repo, file)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	bundleRunGit(t, repo, "add", "-A")
	bundleRunGit(t, repo, "commit", "-m", subject)
}

func bundleRunGit(t *testing.T, repo string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// TestSelfReviewBundleHandlerRequiresRepo pins repo as mandatory: without
// it the route could only misreport an archived change's sources as
// absent, so the request is refused before anything is read.
func TestSelfReviewBundleHandlerRequiresRepo(t *testing.T) {
	ts, _ := recordTestServer(t, "proj", "kan-1")

	code, body := doGet(t, ts, selfReviewPath("proj", "kan-1", ""))
	if code != http.StatusBadRequest {
		t.Fatalf("GET bundle without repo = %d (%s), want 400", code, body)
	}
	// The refusal names itself: a caller reading only the body learns repo
	// was the problem.
	if !strings.Contains(body, "repo is required") {
		t.Errorf("400 body = %s, want the missing repo named", body)
	}
}

// TestGitBoundInsideWriteBudget pins the bound's relation to the server's
// own write budget: a git call that hangs costs at most DefaultGitBound,
// comfortably inside writeTimeout, so the degraded note-bearing bundle is
// still delivered rather than cut off mid-write.
func TestGitBoundInsideWriteBudget(t *testing.T) {
	if selfreview.DefaultGitBound*2 >= 30*time.Second {
		t.Errorf("git bound %s must sit well inside the write budget 30s (server.go writeTimeout)", selfreview.DefaultGitBound)
	}
}

// TestSelfReviewBundleServesSummary pins the recorded change summary's
// place in the served bundle: a change whose store holds a summary gets it
// as the bundle's first section, verbatim, and an unknown change's bundle
// reports it skipped like every other missing source -- the store read
// failing with ErrChangeNotFound is absence, never a 5xx.
func TestSelfReviewBundleServesSummary(t *testing.T) {
	repo := t.TempDir()
	ts, fs := recordTestServer(t, "proj", "kan-1")
	fs.recordedSummary = "what changed and why, grouped by area"
	fs.summaryFound = true

	resp, err := http.Get(ts.URL + selfReviewPath("proj", "kan-1", repo))
	if err != nil {
		t.Fatalf("GET bundle: %v", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	body := string(raw)

	summaryAt := strings.Index(body, "## change summary")
	ledgerAt := strings.Index(body, "## .superpowers/sdd/ledgers/kan-1.md")
	if summaryAt == -1 {
		t.Fatalf("served bundle carries no change summary section:\n%s", body)
	}
	if ledgerAt != -1 && ledgerAt < summaryAt {
		t.Errorf("ledger section precedes the change summary:\n%s", body)
	}
	if !strings.Contains(body, "what changed and why, grouped by area") {
		t.Errorf("summary content not served verbatim:\n%s", body)
	}

	// An unknown change's summary reads as absence, not as a store
	// failure: the bundle still serves, with the source skipped.
	code, body2 := doGet(t, ts, selfReviewPath("proj", "never-heard", repo))
	if code != http.StatusOK {
		t.Fatalf("GET bundle for an unknown change = %d (%s), want 200", code, body2)
	}
	if !strings.Contains(body2, "skipped: change summary (absent)") {
		t.Errorf("unknown change's summary not reported skipped:\n%s", body2)
	}
}

// TestSelfReviewBundleHandlerSummaryStoreFailure pins the summary read's
// own failure branch: a store read of the change summary that fails for a
// real reason — not ErrChangeNotFound — is a 5xx, never a bundle a caller
// could mistake for the change's own, the same contract the run record
// read's failure branch carries.
func TestSelfReviewBundleHandlerSummaryStoreFailure(t *testing.T) {
	ts, fs := recordTestServer(t, "proj", "kan-1")
	fs.changeSummaryErr = errors.New("store exploded")

	code, body := doGet(t, ts, selfReviewPath("proj", "kan-1", t.TempDir()))
	if code != http.StatusInternalServerError {
		t.Fatalf("GET bundle with a failing summary read = %d (%s), want 500", code, body)
	}
}

// TestSelfReviewBundleHandlerPassesPanelRan pins the read the records-loss
// note stands on: a change whose store holds a completed flow.review-panel
// stage run but no dispatch rows gets the loss named loudly in the served
// bundle, not a plain "skipped (absent)" that reads as no panel ever ran.
func TestSelfReviewBundleHandlerPassesPanelRan(t *testing.T) {
	repo := t.TempDir()
	ts, fs := recordTestServer(t, "proj", "kan-1")
	fs.stageCompleted = true

	code, body := doGet(t, ts, selfReviewPath("proj", "kan-1", repo))
	if code != http.StatusOK {
		t.Fatalf("GET bundle = %d (%s), want 200", code, body)
	}
	if !strings.Contains(body, "note: RECORDS LOSS") {
		t.Errorf("served bundle carries no records-loss note:\n%s", body)
	}
}

// TestSelfReviewBundleHandlerStageReadFailureIs5xx pins the stage-completed
// read's own failure branch: a store read that fails for a real reason is a
// 5xx, never a bundle a caller could mistake for the change's own — the
// same contract the run record and summary reads carry.
func TestSelfReviewBundleHandlerStageReadFailureIs5xx(t *testing.T) {
	ts, fs := recordTestServer(t, "proj", "kan-1")
	fs.stageCompletedErr = errors.New("store exploded")

	code, body := doGet(t, ts, selfReviewPath("proj", "kan-1", t.TempDir()))
	if code != http.StatusInternalServerError {
		t.Fatalf("GET bundle with a failing stage read = %d (%s), want 500", code, body)
	}
}

// RecordSelfReviewFinding mirrors store.Store.RecordSelfReviewFinding's
// disposition refusal -- the one rule the endpoint test drives -- and
// allocates ids in insertion order.
func (f *fakeStore) RecordSelfReviewFinding(_ context.Context, projectKey string, in records.SelfReviewFinding) (records.SelfReviewFinding, error) {
	switch in.Disposition {
	case "fixed", "filed", "declined":
	default:
		return records.SelfReviewFinding{}, fmt.Errorf("%w: disposition %q", store.ErrSelfReviewFindingInvalid, in.Disposition)
	}
	out := in
	out.ID = int64(len(f.selfReviewFindings) + 1)
	f.selfReviewFindings = append(f.selfReviewFindings, selfReviewFindingRecord{finding: out, projectKey: projectKey})
	return out, nil
}

// ListSelfReviewFindings mirrors store.Store.ListSelfReviewFindings: one
// change's rows in insertion order, never nil.
func (f *fakeStore) ListSelfReviewFindings(_ context.Context, projectKey, change string) ([]records.SelfReviewFinding, error) {
	out := []records.SelfReviewFinding{}
	for _, r := range f.selfReviewFindings {
		if r.projectKey == projectKey && r.finding.Change == change {
			out = append(out, r.finding)
		}
	}
	return out, nil
}

// selfReviewFindingRecord is fakeStore's in-memory self_review_findings row.
type selfReviewFindingRecord struct {
	finding    records.SelfReviewFinding
	projectKey string
}

// TestSelfReviewFindingsEndpoint drives POST and GET
// /api/v1/self-review/{project}/{change}/findings: a valid row is 201 and
// echoed back, a store validation refusal is 400, and the read returns the
// posted rows in order.
func TestSelfReviewFindingsEndpoint(t *testing.T) {
	ts, _ := recordTestServer(t, "proj", "kan-1")
	url := ts.URL + "/api/v1/self-review/proj/kan-1/findings"

	for _, body := range []map[string]any{
		{"angle": "myflow-fix", "note": "guard gap", "disposition": "fixed", "ref": "abc1234", "blastRadius": 2},
		{"angle": "myflow-cost", "note": "redesign", "disposition": "filed", "ref": "KAN-9"},
	} {
		resp, raw := postJSON(t, url, body)
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("POST %v = %d (%s), want 201", body, resp.StatusCode, raw)
		}
		var got records.SelfReviewFinding
		if err := json.Unmarshal(raw, &got); err != nil {
			t.Fatalf("decode POST body %s: %v", raw, err)
		}
		if got.ID == 0 || got.Change != "kan-1" || got.Ref != body["ref"] {
			t.Errorf("POST echoed %+v, want an id, change kan-1 and ref %v", got, body["ref"])
		}
	}

	if resp, raw := postJSON(t, url, map[string]any{"angle": "myflow-fix", "note": "x", "disposition": "bogus"}); resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("POST bogus = %d (%s), want 400", resp.StatusCode, raw)
	}

	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET findings: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET findings = %d, want 200", resp.StatusCode)
	}
	var got []records.SelfReviewFinding
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode GET body: %v", err)
	}
	if len(got) != 2 || got[0].Disposition != "fixed" || got[1].Disposition != "filed" ||
		got[0].BlastRadius == nil || *got[0].BlastRadius != 2 {
		t.Errorf("GET = %+v, want the fixed row (blast radius 2) then the filed row", got)
	}
}
