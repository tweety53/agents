// Package api implements myflowd's HTTP surface: today, task 4's change
// endpoints over internal/store; later tasks add the stage-mark and
// statistics endpoints beside them, through the same Server and the same
// error mapping. internal/store remains the only package that builds SQL --
// this package calls typed repository methods and never assembles a query.
package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/tweety53/agents/stats/internal/config"
	"github.com/tweety53/agents/stats/internal/store"
)

// HTTP server timeouts and the request body cap. This daemon serves a
// local CLI and a local SPA over loopback only -- never a public API and
// never a link a stranger's browser could open -- so these are not tuned
// for internet-facing traffic; they exist as default hygiene
// (go-api-design, go-security-audit both name an http.Server with no
// timeouts as a defect regardless of exposure) and to bound a stuck local
// client, not to defend against a hostile one:
//   - readHeaderTimeout is the one that matters most even on loopback: a
//     client that opens a connection and sends its headers one byte at a
//     time (reproduced with `nc`) would otherwise hold a goroutine and a
//     connection open indefinitely, with no relation to how fast the
//     handler itself runs. 5s is generous for a loopback client that is
//     genuinely trying.
//   - readTimeout covers the body too, so a PUT with a slow-trickling body
//     (the case maxRequestBodyBytes alone does not bound) cannot hang
//     forever either. 10s is enough for even a very large change record
//     over loopback, with headroom.
//   - writeTimeout bounds how long a response, including a statistics
//     aggregation, is allowed to take to write. 30s leans generous on
//     purpose: task 11's views compute over an arbitrary period and this
//     is not the place to make that call tight. Go's WriteTimeout covers
//     handler execution plus the write, not the write alone -- it is
//     really one shared per-request budget across every route on this
//     mux, state and stats alike, and it does not return a clean error
//     status when it fires: the connection is cut mid-write, so the
//     client sees a truncated body.
//   - idleTimeout bounds a keep-alive connection sitting idle between
//     requests, e.g. the SPA's tab left open. 120s matches the
//     go-api-design example default; there is nothing about this daemon
//     that argues for shorter.
//
// Task 11's decision on the above, made deliberately rather than left to
// this shared budget by default: GET /api/v1/stats/{view} is wrapped in its
// own http.TimeoutHandler (statsWriteTimeout, below) rather than trusting
// writeTimeout alone. writeTimeout's truncated-body failure mode is
// tolerable for the small, fast state and stage-mark routes it was sized
// for, but it is the wrong failure mode for an aggregation query: a client
// cannot tell a truncated JSON body apart from a mid-response network
// blip, and myflow's CLI-facing routes need that distinction. statsWriteTimeout
// sits comfortably inside writeTimeout so the TimeoutHandler itself always
// gets to write its own clean response before the shared budget could cut
// it off too.
const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 10 * time.Second
	writeTimeout      = 30 * time.Second
	idleTimeout       = 120 * time.Second

	// statsWriteTimeout bounds a single GET /api/v1/stats/{view} request.
	// See this const block's own doc comment above for why this route gets
	// its own timeout instead of sharing writeTimeout's truncate-on-expiry
	// behaviour. 20s leaves several seconds of headroom under writeTimeout
	// even accounting for header/body read time already spent before the
	// handler starts.
	statsWriteTimeout = 20 * time.Second

	// statsTimeoutMessage is the plain-text body http.TimeoutHandler writes,
	// with a 503, when a statistics query does not finish inside
	// statsWriteTimeout.
	statsTimeoutMessage = "myflowd: statistics query timed out"

	// maxRequestBodyBytes caps a request body -- PUT's whole-object change
	// record today. A few KB is realistic for the largest record this
	// change's own field vocabulary can produce (a handful of repos, a
	// metrics-shaped models blob); 1 MiB leaves generous headroom above
	// that without leaving the cap effectively unbounded.
	maxRequestBodyBytes = 1 << 20 // 1 MiB
)

// ChangeStore is the store dependency the change endpoints need, defined
// here at the consumer rather than alongside store.Store, per
// go-interface-design's "define interfaces at the consumer": exactly the
// methods changeHandler calls, nothing more, so a fake implementing it for
// a test needs no database at all.
type ChangeStore interface {
	GetChange(ctx context.Context, projectKey, name string) (store.Change, error)
	PutChange(ctx context.Context, c store.Change) error
	QueryChanges(ctx context.Context, q store.Query) ([]store.Change, int, error)
	// projectResolver backs resolveProjectParam (stats.go), which
	// parseChangeQuery (changes.go) uses to resolve the changes list's own
	// "project" filter -- the same rule every stats view applies to its
	// own "project" parameter. Embedded rather than spelling
	// ProjectKeysByDisplayName's signature out a second time (panel round
	// 1, F2) -- see StatsStore's own embedding (stats.go) for the same
	// reasoning.
	projectResolver
}

// var _ ChangeStore = (*store.Store)(nil) verifies at compile time that the
// real store satisfies the interface this package actually depends on.
var _ ChangeStore = (*store.Store)(nil)

// DaemonHeader is set, to DaemonHeaderValue, on every response this daemon
// writes -- success and error alike -- so a caller can tell "myflowd
// answered" apart from "some other process is listening on this port and
// happened to also return a 409 or a 404".
//
// This exists for task 5's CLI: its whole never-block guarantee rests on
// telling a genuine store refusal (409) or a genuine not-found (404) apart
// from the store being unreachable, and both of those status codes are
// exactly the ones a foreign process on the port is likely to produce by
// accident -- a generic API gateway's default error page, a stale service
// left listening after a restart, a test fixture like Python's
// http.server. Trusting status codes alone made a bare 409 or 404 from
// anything squatting the port stop the pipeline with no store, and no
// database, involved at all. The CLI (internal/client) refuses to treat a
// response as a store answer unless this header is present with this
// value; see its own doc comment for the client-side half of the contract.
//
// The value is a fixed protocol marker, not a full semantic version: this
// header exists to answer "is this myflowd" for the fallback decision, not
// to support version negotiation between daemon and CLI generations, so it
// changes only on a deliberate wire-format break, not on every release.
const (
	DaemonHeader      = "Myflow-Daemon"
	DaemonHeaderValue = "myflowd/1"
)

// withDaemonHeader wraps next so DaemonHeader is set on every response the
// server writes, before the wrapped handler runs -- a handler that panics
// or returns early still leaves the header set, since http.ResponseWriter
// only sends headers once the first byte of the body (or an explicit
// WriteHeader) is written.
func withDaemonHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(DaemonHeader, DaemonHeaderValue)
		next.ServeHTTP(w, r)
	})
}

// Server is myflowd's HTTP server.
type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
}

// Option configures an optional part of New's behaviour. Every existing
// caller that passes no Option at all (every test in this package,
// predating this type) keeps building exactly the API-only mux it always
// has, with no SPA route mounted and nothing else changed.
type Option func(*serverOptions)

// serverOptions holds what the variadic Option parameters on New configure.
// It exists as its own type, rather than fields tacked onto Server, so a
// caller reads what New optionally accepts in one place.
type serverOptions struct {
	spa http.Handler
}

// WithSPA mounts h at "/" as the daemon's user interface, task 12's
// embedded SPA (internal/web.Handler) in production. internal/api itself
// never imports internal/web -- that dependency, and the requirement that
// building it needs `vite build` to have already run, belongs to whichever
// caller wants the SPA served, not to every test in this package that
// exercises only the API surface. Passing no WithSPA option (as every
// existing test here does) leaves "/" unmounted: a request to it 404s
// exactly as it always has.
func WithSPA(h http.Handler) Option {
	return func(o *serverOptions) {
		o.spa = h
	}
}

// apiPathPrefix is the prefix every route this daemon's API owns falls
// under. It is registered on the mux in its own right (see New, below) so
// that a request under it which matches none of the specific routes above
// -- a typo'd path, a retired one, a method the route doesn't support --
// is answered as an API request (a JSON 404 naming the path) rather than
// falling through to the SPA route mounted at "/" and being "answered"
// with index.html. Go's http.ServeMux resolves the most specific
// registered pattern for a given request, so this prefix loses to every
// exact route above it and wins over the catch-all "/" WithSPA mounts --
// which is exactly the ordering task 12's "API paths must never be
// swallowed by that fallback" requirement needs, achieved by mux
// specificity rather than by handler order or by extra logic in the SPA
// handler itself, which has no notion of "/api/" at all (see
// internal/web.Handler's own doc comment).
//
// That specificity argument holds only for an ordinarily-escaped path. It
// does NOT hold for a path containing a percent-encoded slash: Go's
// enhanced ServeMux (1.22+) matches patterns against r.URL.EscapedPath(),
// which keeps "%2F"/"%2f" as three literal characters rather than the "/"
// it decodes to in r.URL.Path. "/api%2Fv1/changes" therefore has no
// literal "/" where "/api/" needs one to match -- its only real path
// separators are the leading "/" and the one before "changes" -- so it
// matches neither the exact route nor this prefix, and falls through to
// "/", the SPA. rejectEncodedSlash (below), wrapped around the whole mux
// rather than folded into this one prefix, is what actually closes that
// gap: it refuses the request before routing sees it at all, so no
// pattern here needs to account for what a percent-encoded segment
// separator might mean.
const apiPathPrefix = "/api/"

// rejectEncodedSlash guards every route this daemon serves -- API and SPA
// alike -- against a request whose path contains a percent-encoded slash
// or a percent-encoded percent sign. See apiPathPrefix's own doc comment
// for the mechanism this closes: such a request can fail to match either
// an exact API route or the "/api/" catch-all (because ServeMux's pattern
// matching operates on r.URL.EscapedPath(), which does not decode "%2F"
// to "/"), and fall through to the SPA route instead -- answered with
// index.html, HTML with no DaemonHeader, which internal/client's fallback
// logic reads as "store unreachable" and silently routes around on a
// perfectly healthy daemon. internal/client.go builds change URLs with
// url.PathEscape(project) / url.PathEscape(name), and PathEscape encodes a
// literal "/" as exactly "%2F" -- so a project key or change name
// containing a slash produces exactly the request this guards against, in
// ordinary use, not only as a crafted attack.
//
// This is enforced by outright refusal, not by re-routing the request to
// what its decoded path "really means": no path segment this API ever
// reads -- project key, change name, view name, command, stage -- has any
// legitimate use for an encoded slash, so refusing it here, once, before
// any handler sees the request, is preferred over decoding and
// re-dispatching, which would just relocate the same "whose decoding is
// authoritative" question instead of removing it. Checking is done
// against EscapedPath() (what ServeMux itself keys off) for the direct
// "%2f"/"%2F" case, and against "%25" (a percent-encoded percent sign) to
// catch a double-encoded slash such as "%252F" -- which, per Go's single
// decode pass, never becomes a literal "/" in EscapedPath() and so is
// merely garbled rather than routing-ambiguous, but is refused anyway
// because there is likewise no legitimate use for an encoded "%" in any
// path segment this API reads, and it is the only way a client could ever
// construct a multiply-encoded slash in the first place.
func rejectEncodedSlash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		escaped := strings.ToLower(r.URL.EscapedPath())
		if strings.Contains(escaped, "%2f") || strings.Contains(escaped, "%25") {
			writeError(w, http.StatusBadRequest, "path must not contain a percent-encoded slash or percent sign")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// New builds a Server bound to cfg.Addr(), or refuses with a wrapped
// config.ErrNonLoopbackHost when cfg.Host is not loopback. This validation
// happens here, before any listener is opened -- New itself never calls
// net.Listen -- so a non-loopback configuration is a startup error the
// caller (cmd/myflowd's main) can exit on immediately, never a runtime
// option a request could trip over later.
//
// logger may be nil, in which case slog.Default() is used.
//
// cs, ss, sts and rs are accepted as four separate interfaces, per
// go-interface-design (each handler depends on exactly the methods it
// calls), even though every real caller passes the same *store.Store for
// all four -- it satisfies ChangeStore, StageStore, StatsStore and
// RecordStore alike.
//
// opts is variadic and, absent WithSPA, changes nothing about the mux this
// function has always built -- see WithSPA's own doc comment for why the
// SPA route is opt-in here rather than a fixed part of every Server.
func New(cfg config.Config, cs ChangeStore, ss StageStore, sts StatsStore, rs RecordStore, logger *slog.Logger, opts ...Option) (*Server, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("api: %w", err)
	}
	if logger == nil {
		logger = slog.Default()
	}

	var so serverOptions
	for _, opt := range opts {
		opt(&so)
	}

	h := &changeHandler{store: cs, logger: logger}
	sh := &stageHandler{store: ss, logger: logger}
	sth := &statsHandler{store: sts, logger: logger}
	rh := &recordHandler{store: rs, logger: logger}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/changes", h.list)
	mux.HandleFunc("GET /api/v1/changes/{project}/{name}", h.get)
	mux.HandleFunc("PUT /api/v1/changes/{project}/{name}", h.put)
	mux.HandleFunc("POST /api/v1/stages/begin", sh.begin)
	mux.HandleFunc("POST /api/v1/stages/end", sh.end)
	mux.HandleFunc("GET /api/v1/stage-runs", sth.listStageRuns)
	mux.Handle("GET /api/v1/stats/{view}", http.TimeoutHandler(http.HandlerFunc(sth.view), statsWriteTimeout, statsTimeoutMessage))
	mux.HandleFunc("GET /api/v1/models", sth.listModels)
	mux.HandleFunc("GET /api/v1/records/{project}/{change}", rh.runRecord)
	mux.HandleFunc("POST /api/v1/records/{project}/{change}/dispatches", rh.recordDispatch)
	mux.HandleFunc("POST /api/v1/records/{project}/{change}/dispatches/end", rh.endDispatch)
	mux.HandleFunc("POST /api/v1/records/{project}/{change}/findings", rh.recordFinding)
	mux.HandleFunc("PATCH /api/v1/records/{project}/{change}/findings/{ref}", rh.setFindingStatus)
	mux.HandleFunc(apiPathPrefix, func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotFound, fmt.Sprintf("no such API route: %s %s", r.Method, r.URL.Path))
	})
	if so.spa != nil {
		mux.Handle("/", so.spa)
	}

	return &Server{
		httpServer: &http.Server{
			Addr: cfg.Addr(),
			// rejectEncodedSlash sits inside withDaemonHeader (never
			// outside it): a request it refuses is still a request this
			// daemon reached and answered, per DaemonHeader's own
			// contract, and internal/client's fallback decision depends
			// on that header being present on exactly that class of
			// response -- a deliberate 400, not an unreachable store.
			Handler:           withDaemonHeader(rejectEncodedSlash(mux)),
			ReadHeaderTimeout: readHeaderTimeout,
			ReadTimeout:       readTimeout,
			WriteTimeout:      writeTimeout,
			IdleTimeout:       idleTimeout,
		},
		logger: logger,
	}, nil
}

// Handler returns the server's http.Handler, for tests that want to drive
// it through httptest.NewServer or httptest.NewRequest without going
// through a real listener.
func (s *Server) Handler() http.Handler {
	return s.httpServer.Handler
}

// Serve accepts connections on ln until the server is shut down. It
// returns nil on a clean shutdown (http.ErrServerClosed) and the
// underlying error otherwise -- so a caller selecting on this channel and
// on a shutdown signal can tell "shut down as asked" apart from "died on
// its own".
func (s *Server) Serve(ln net.Listener) error {
	err := s.httpServer.Serve(ln)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

// ListenAndServe opens a listener on the configured address and serves on
// it. It blocks until the server is shut down.
func (s *Server) ListenAndServe() error {
	ln, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return fmt.Errorf("api: listen on %s: %w", s.httpServer.Addr, err)
	}
	return s.Serve(ln)
}

// Shutdown gracefully shuts the server down: requests already in flight
// are allowed to complete, no new connection is accepted, and Shutdown
// returns once every in-flight request has finished (or ctx expires
// first, whichever comes first). The caller closes the store's connection
// pool only after Shutdown returns, so a request that reaches the store
// mid-shutdown is never cut off underneath it.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// errorResponse is the JSON body every non-2xx response carries. Code is
// omitted by default -- most callers of writeError have nothing more
// specific to say than the status and the message -- and set only where a
// caller needs the *reason* a 400 was raised to be machine-distinguishable
// from every other 400, not merely human-readable. CodeUndocumentedStage
// is the one code this daemon defines today; see its own doc comment for
// why status alone is not enough here, the same lesson DaemonHeader
// already teaches for a different ambiguity.
type errorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code,omitempty"`
}

// CodeUndocumentedStage is errorResponse.Code's value on the one 400 that
// internal/client's BeginStage/EndStage must react to differently from
// every other 400: a stage name internal/stages.Validate rejected. Before
// this code existed, both methods mapped *any* 400 -- a missing required
// field, a malformed body, an undocumented stage, anything else this
// handler might one day reject with 400 -- to the same
// client.ErrUndocumentedStage, so an operator whose request was rejected
// for an unrelated reason was told "undocumented stage", a confidently
// wrong diagnosis. internal/client keeps its own literal copy of this
// string (client.go's undocumentedStageErrorCode), for the same reason it
// keeps its own copy of DaemonHeader/DaemonHeaderValue: the CLI's
// boundary rule is that it knows only HTTP, never this package.
const CodeUndocumentedStage = "undocumented_stage"

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}

// writeErrorWithCode is writeError plus a machine-readable Code -- see
// errorResponse.Code's own doc comment for why this exists as a distinct
// function rather than an optional parameter on writeError: every other
// call site has nothing more specific to say, and a "" default would be
// easy to pass by accident where a real code was meant.
func writeErrorWithCode(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, errorResponse{Error: msg, Code: code})
}

// mapStoreError turns a typed store error into the HTTP status and message
// this daemon reports for it. Every typed error this task's requirements
// name is covered explicitly; anything else -- an unwrapped driver error,
// a context cancellation, a bug -- is reported as a generic 500 with no
// internal detail leaked to the caller, and logged here since nothing else
// will see it.
//
// ErrMonotonicViolation's status (409) is the one this mapping exists to
// keep distinguishable above all others: task 5's CLI decides "the store
// answered correctly, report it" versus "the store is unreachable, take
// the journal fallback" on exactly whether the response carried this
// status. A 5xx (ErrTooManyAttemptCollisions, or the unmapped-error
// default) is deliberately grouped with "unreachable" in that decision --
// it signals the store could not complete the request, which is the same
// fact a dead connection signals -- while every 4xx here, monotonic
// refusal included, signals the store was reached and answered. That is
// why 409 rather than, say, 503 was chosen for the refusal: it must never
// share a status class with the failures the CLI's fallback exists for.
//
// Every other typed error below gets its own deliberate status rather than
// folding into a blanket 500, chosen by what kind of failure it actually
// is:
//   - 404 Not Found: the named change does not exist (ErrChangeNotFound).
//   - 400 Bad Request: the caller's request was malformed -- an invalid
//     state, an unresolvable project bootstrap, a duplicate repository, or
//     an unrecognised query field or operator. Every one of these is
//     something a different request would have avoided.
//   - 422 Unprocessable Entity: the request was well-formed and the named
//     resource exists, but the data needed to fulfil it is not there
//     (ErrTokensUnavailable, ErrPricingNotFound) -- distinct from 400
//     because the caller did nothing wrong.
//   - 503 Service Unavailable: transient contention inside the store
//     itself (ErrTooManyAttemptCollisions) that a retry, not a different
//     request, might resolve.
func mapStoreError(logger *slog.Logger, action string, err error) (status int, msg string) {
	switch {
	case errors.Is(err, store.ErrChangeNotFound):
		return http.StatusNotFound, err.Error()
	case errors.Is(err, store.ErrMonotonicViolation):
		return http.StatusConflict, err.Error()
	case errors.Is(err, store.ErrInvalidState):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, store.ErrInvalidMainCheckoutPath):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, store.ErrDuplicateRepoRoot):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, store.ErrUnknownField):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, store.ErrUnknownFilterOp):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, store.ErrInvalidMetricsKeyPath):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, store.ErrOffsetWithNoLimit):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, store.ErrTokensUnavailable):
		return http.StatusUnprocessableEntity, err.Error()
	case errors.Is(err, store.ErrPricingNotFound):
		return http.StatusUnprocessableEntity, err.Error()
	case errors.Is(err, store.ErrFindingNotFound):
		return http.StatusNotFound, err.Error()
	case errors.Is(err, store.ErrDispatchNotFound):
		return http.StatusNotFound, err.Error()
	case errors.Is(err, store.ErrTooManyAttemptCollisions):
		return http.StatusServiceUnavailable, err.Error()
	case errors.Is(err, store.ErrTooManyDispatchSeqCollisions):
		return http.StatusServiceUnavailable, err.Error()
	default:
		if logger != nil {
			logger.Error(action, "error", err)
		}
		return http.StatusInternalServerError, "internal error"
	}
}
