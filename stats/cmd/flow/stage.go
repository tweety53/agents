package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/tweety53/agents/stats/internal/client"
	"github.com/tweety53/agents/stats/internal/fallback"
	"github.com/tweety53/agents/stats/internal/stages"
)

// defaultHarness is used when `stage begin` is not told which harness is
// running it (-harness, or the FLOW_HARNESS environment variable).
// stage_runs.harness is NOT NULL (design.md's data model), so this must
// never be empty -- "unknown" is an honest value the harvester and every
// statistics view can filter on, distinct from a harness that genuinely
// wrote no transcript (design.md's own tokens_available flag, task 10).
// This is also the one and only place a stage run's harness is ever
// recorded: `stage end` carries no harness of its own, deliberately --
// see internal/api's ApplyEndStageMark doc comment (task 10's post-commit
// review, finding F1) for why deriving tokens_available from anything an
// end mark supplies, rather than from this recorded value, was unsound.
const defaultHarness = "unknown"

const stageUsage = `usage: flow stage begin [-addr url] [-timeout dur] [-C dir] [-harness name] [-session id]
                          -command cmd -stage key -session-token token (<change> | -jira-key KEY)
       flow stage end [-addr url] [-timeout dur] [-C dir]
                        -command cmd -stage key -outcome outcome
                        [-fix-rounds n] [-panel-rounds n] [-findings json] (<change> | -jira-key KEY)
       flow stage wrap [-addr url] [-timeout dur] [-C dir] [-harness name] [-session id]
                        -command cmd -stage key -session-token token (<change> | -jira-key KEY)
                        -- <work command and args...>
       flow stage keys

-stage takes a stage KEY, not its prose name -- one of README.md's Level 1
-- the stages of each command table's Key column; an undocumented key is
rejected before it ever reaches the store.

stage keys prints every documented stage key, one per line, from this
checkout's own internal/stages table -- the vocabulary's one served source;
the check-stage-mark-calls guard consumes this output rather than keeping a
transcription of its own.

-jira-key records a /flow-plan session against its Jira key before the change
exists; it replaces the <change> argument and is accepted for the plan.session
stage only.

-session-token must be a literal, unique token this command writes -- never a
shell substitution ("$(...)", a backtick, or "$VAR"): the transcript
records the command text before the shell expands it, so a substitution
would be recorded identically by every caller and identify nothing.

This CLI can only reject a substitution shape it can still see -- write
-session-token $T (unquoted or double-quoted) and the calling shell expands
$T to its value before this program ever runs, so the value this program
receives is an ordinary literal and passes every check here, while the
transcript still records the unexpanded "$T" and the mark binds nothing.
There is no fix at this layer: type the literal token itself on the command
line, not a variable holding it.
`

func runStage(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, stageUsage)
		return 2
	}

	switch args[0] {
	case "begin":
		return runStageBegin(ctx, args[1:], stderr)
	case "end":
		return runStageEnd(ctx, args[1:], stderr)
	case "wrap":
		return runStageWrap(ctx, args[1:], stdin, stdout, stderr)
	case "keys":
		return runStageKeys(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "flow: unknown stage command %q\n", args[0])
		fmt.Fprint(stderr, stageUsage)
		return 2
	}
}

// stageIdentityFlags is common to `stage begin` and `stage end`: the store
// connection, the working directory the project key resolves from, and
// the command/stage identity every mark carries.
type stageIdentityFlags struct {
	addr    string
	timeout time.Duration
	dir     string
	command string
	stage   string
	name    string
	jiraKey string
}

func registerStageIdentityFlags(fset *flag.FlagSet, f *stageIdentityFlags) {
	fset.StringVar(&f.addr, "addr", resolveDefaultAddr(), "flowd base URL")
	fset.DurationVar(&f.timeout, "timeout", defaultTimeout, "store request timeout before falling back")
	fset.StringVar(&f.dir, "C", "", "resolve the project key as if run from this directory (default: cwd)")
	fset.StringVar(&f.command, "command", "", "the flow command this stage belongs to, e.g. /flow")
	fset.StringVar(&f.stage, "stage", "", "the stage key, exactly as README.md's Level 1 table's Key column documents it -- never its prose name")
	fset.StringVar(&f.jiraKey, "jira-key", "", "record against this Jira key instead of a change name -- /flow-plan's plan.session only")
}

func finishStageIdentityFlags(fset *flag.FlagSet, f *stageIdentityFlags) error {
	switch {
	case f.jiraKey != "" && fset.NArg() == 0:
		f.name = ""
	case f.jiraKey == "" && fset.NArg() == 1:
		f.name = fset.Arg(0)
	default:
		return fmt.Errorf("expected exactly one argument, the change name, or -jira-key with no argument")
	}
	if f.command == "" || f.stage == "" {
		return fmt.Errorf("-command and -stage are both required")
	}
	if f.dir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("resolve working directory: %w", err)
		}
		f.dir = wd
	}
	return nil
}

// resolveHarness returns harnessFlag if set, else FLOW_HARNESS, else
// defaultHarness -- never empty, since stage_runs.harness is NOT NULL.
func resolveHarness(harnessFlag string) string {
	if harnessFlag != "" {
		return harnessFlag
	}
	if v := os.Getenv("FLOW_HARNESS"); v != "" {
		return v
	}
	return defaultHarness
}

// sessionTokenShellVarPattern matches a "$" immediately followed by a shell
// variable-name character ($VAR, $HOME, $_x) -- the third of the three
// substitution shapes validateSessionToken rejects. It deliberately does not
// match a bare "$" or "$$" (a PID expansion) on their own: design.md and
// tasks.md name exactly three shapes ("$(", a backtick, and "$" followed
// by a name), and this is the pattern for the third.
//
// internal/api/stages.go's validateSessionTokenShape carries the identical
// pattern and the identical reasoning, server side -- see its own doc
// comment for why the two are not shared code.
var sessionTokenShellVarPattern = regexp.MustCompile(`\$[A-Za-z_]`)

// validateSessionToken rejects a sessionToken that cannot identify anything: one
// carrying a shell-substitution shape a caller's own shell would expand
// before the transcript is ever written.
//
// Why each shape is rejected: `tool_use.input.command` -- what a later
// harvest cycle (KAN-172, task 2) reads a transcript for -- records the
// command text exactly as it was handed to the tool, before the shell
// ever expands it. A sessionToken built from `$(...)`, a backtick, or `$VAR`
// therefore lands in every calling session's transcript as the identical,
// unexpanded literal, and discriminates nothing between them -- design.md's
// "the sessionToken is a literal, never a shell substitution".
//
// This is the CLI-side half of the check; internal/api's
// validateSessionTokenShape is the server-side half, run again on every mark
// (including one replayed from the journal, which never passes back
// through this function) as defence in depth, exactly as stages.Validate
// is checked in both places for an undocumented stage key.
//
// A KNOWN, UNCLOSEABLE GAP (tasks.md task 1b): this check, and its
// server-side twin, both operate on the string this process's own argv
// carries -- and argv is populated by the calling shell *after* it has
// already expanded any `$VAR`. A caller who writes `-session-token $T`
// hands this function the literal value of $T, indistinguishable here
// from a token the caller typed by hand; the transcript records the
// command exactly as typed, i.e. still carrying `$T`, so the token this
// function saw and validated never appears in any transcript at all. Nothing
// observable at this layer -- not argv, not the environment, not the
// process tree -- carries the pre-expansion command text, so no check
// here can catch this shape: rejecting it would require rejecting every
// ordinary literal too, which is exactly the false-negative-over-false-
// positive trade this file's own package doc (tasks.md's "Global
// Constraints") forbids inverting.
//
// The actual defence against this shape lives downstream, not here:
// internal/harvest's Watcher already treats a token that never matches
// any transcript as a run to give up on after a bounded number of
// resolution cycles, logging a warning rather than binding it silently
// (Watcher.resolveSessionTokens's `case 0` branch, gated by
// maxSessionTokenResolutionCycles) -- the same bounded-give-up path a
// token from `$(...)`, a backtick, or `$VAR` written *without* the shell
// getting to expand it (e.g. inside single quotes) already falls into if
// this function's three rejections above were ever bypassed. That
// warning is the only place in this system able to observe "this
// specific token was recorded here but is absent everywhere it should
// appear" -- the CLI, by construction, only ever sees one command's
// worth of already-expanded argv and can never compare it against a
// transcript.
func validateSessionToken(sessionToken string) error {
	switch {
	case strings.Contains(sessionToken, "$("):
		return fmt.Errorf("-session-token %q contains a command substitution \"$(...)\" -- "+
			"the transcript records the command text before the shell expands it, "+
			"so this value would be recorded identically by every caller and identify nothing; "+
			"write a literal token instead", sessionToken)
	case strings.Contains(sessionToken, "`"):
		return fmt.Errorf("-session-token %q contains a backtick -- backtick command substitution is expanded by "+
			"the shell after the transcript already recorded the unexpanded text, "+
			"so this value would be recorded identically by every caller and identify nothing; "+
			"write a literal token instead", sessionToken)
	case sessionTokenShellVarPattern.MatchString(sessionToken):
		return fmt.Errorf("-session-token %q contains a shell variable reference ($VAR) -- "+
			"the transcript records the command text before the shell expands it, "+
			"so this value would be recorded identically by every caller and identify nothing; "+
			"write a literal token instead", sessionToken)
	}
	return nil
}

// stageJournalPath is where a stage mark's fallback intent is journalled:
// beside the change's state file and its own state journal
// (fallback.JournalFilePath), but under a distinct filename. This is
// deliberate, not an oversight: a stage mark's journal entry is not a
// whole-object state write, and internal/reconcile's replay
// (already shipped in task 6) interprets every entry in
// fallback.JournalFilePath's file as exactly that -- a change PUT body,
// decoded with DisallowUnknownFields. Mixing a stage-mark body into that
// same file would make every entry after it fail to decode on replay.
// Using a sibling file instead means this reuses fallback's own
// AppendJournalEntry/ReadJournalEntries (the exact functions `state set`'s
// fallback already reuses -- see this file's own doc comments) without
// touching, or risking, the state journal's replay contract.
func stageJournalPath(projectKey, name string) string {
	return fallback.JournalFilePath(projectKey, name) + ".stage"
}

// stageMarkJournalBody is what gets journalled for a stage mark that could
// not reach the store: the mark's own kind ("begin" or "end") alongside
// the exact wire request that would have been sent, so a future
// reconciler has everything it needs to replay it without this file
// needing a second encoding.
type stageMarkJournalBody struct {
	Kind    string `json:"kind"`
	Request any    `json:"request"`
}

// journalStageMark appends kind/req to the stage mark journal for
// projectKey/name and prints the one warning line every fallback path
// prints. Errors from the journal write itself are deliberately
// swallowed, exactly as runStateSet's fallback swallows
// WriteStateFile/AppendJournalEntry errors: there is nothing further to
// fall back to, and the guarantee this exists to uphold is "never block",
// not "never lose telemetry".
func journalStageMark(projectKey, name, kind string, req any, stderr io.Writer) {
	body, err := json.Marshal(stageMarkJournalBody{Kind: kind, Request: req})
	if err == nil {
		_ = fallback.AppendJournalEntry(stageJournalPath(projectKey, name), projectKey, name, body, time.Now())
	}
	fmt.Fprintln(stderr, "⚠ flow: store unreachable — wrote local journal (flow journal flush's stderr carries the real cause)")
}

// warnNoOpenStageRun prints the KAN-700 diagnosis on a stage end the store
// answered with no open stage run: the store was reached and gave a
// definitive answer, so the failure is a bookkeeping slip -- the begin was
// never recorded, or another session closed the run -- and the message
// names it instead of the store-unreachable fallback an outage gets.
func warnNoOpenStageRun(stderr io.Writer, name, stage string) {
	fmt.Fprintf(stderr, "⚠ flow: no open stage run for %s/%s — was `stage begin` recorded?\n", name, stage)
}

// runStageBegin implements `flow stage begin`. It validates the stage
// key against internal/stages' documented table -- README.md's Level 1
// table, transcribed there -- before ever contacting the store: an
// undocumented stage key is a defect in the caller, not a store failure,
// so it is reported and exits 2 (a usage error), never taking the
// never-block fallback path a store failure would.
//
// A successful begin, and every store failure, exit 0: a mark must never
// block the pipeline any more than `state set` may (design.md, "The
// pipeline never blocks on this subsystem"). On any failure other than a
// successful store answer or a caller mistake, the mark is journalled
// (journalStageMark, built on fallback.AppendJournalEntry -- the exact
// function `state set`'s fallback already uses, not a second
// implementation) and one warning line is printed.
func runStageBegin(ctx context.Context, args []string, stderr io.Writer) int {
	fset := flag.NewFlagSet("flow stage begin", flag.ContinueOnError)
	fset.SetOutput(stderr)
	var f stageIdentityFlags
	registerStageIdentityFlags(fset, &f)
	harnessFlag := fset.String("harness", "", "the harness running this mark (default: $FLOW_HARNESS, or \"unknown\")")
	sessionFlag := fset.String("session", "", "the harness session id, if known; defaults to CLAUDE_CODE_SESSION_ID when set")
	sessionTokenFlag := fset.String("session-token", "", "a literal, unique token this run generates once and passes unchanged on every mark it makes -- never a shell substitution -- so the daemon can later bind every stage run carrying it to the session that made it (required)")
	if err := fset.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintf(stderr, "flow: %v\n", err)
		fmt.Fprint(stderr, stageUsage)
		return 2
	}
	noteAddrUsage(fset, stderr, f.addr)
	if err := finishStageIdentityFlags(fset, &f); err != nil {
		fmt.Fprintf(stderr, "flow: %v\n", err)
		fmt.Fprint(stderr, stageUsage)
		return 2
	}
	if *sessionTokenFlag == "" {
		// A missing -session-token is a caller mistake, not a stage outcome:
		// exit non-zero naming the flag, exactly as -command/-stage above
		// already do -- the one class of nonzero exit the never-block
		// rule permits (tasks.md, "Step 1: The column and the flags").
		fmt.Fprintln(stderr, "flow: -session-token is required")
		fmt.Fprint(stderr, stageUsage)
		return 2
	}
	if err := validateSessionToken(*sessionTokenFlag); err != nil {
		fmt.Fprintf(stderr, "flow: %v\n", err)
		return 2
	}

	if err := stages.Validate(stages.Command(f.command), f.stage); err != nil {
		fmt.Fprintf(stderr, "flow: %v\n", err)
		return 2
	}

	projectKey, mainCheckout, err := fallback.ProjectKey(f.dir)
	if err != nil {
		fmt.Fprintf(stderr, "flow: resolve project key: %v\n", err)
		return 1
	}

	// The transcript records a mark's command text before the shell expands
	// it, so a token written as -session-token $TOKEN is never found by the
	// harvester's search (internal/harvest's isSessionMarkCommand). A row
	// born with its session_id already set skips that search entirely
	// (store.UnresolvedSessionTokens selects session_id IS NULL): binding
	// from CLAUDE_CODE_SESSION_ID here, when the caller did not pass
	// -session, makes the mark self-attributing on Claude Code. The token
	// search remains the fallback for a mark made where the variable is
	// unset.
	var sessionID *string
	switch {
	case *sessionFlag != "":
		sessionID = sessionFlag
	default:
		if v := os.Getenv("CLAUDE_CODE_SESSION_ID"); v != "" {
			sessionID = &v
		}
	}
	req := client.BeginStageRequest{
		ProjectKey:       projectKey,
		MainCheckoutPath: mainCheckout,
		ChangeName:       f.name,
		Harness:          resolveHarness(*harnessFlag),
		SessionID:        sessionID,
		SessionToken:     *sessionTokenFlag,
		Command:          f.command,
		Stage:            f.stage,
		StartedAt:        time.Now(),
		JiraKey:          f.jiraKey,
	}

	result, beginErr := beginStage(ctx, f.addr, f.timeout, req)
	switch {
	case beginErr == nil:
		warnSupersededRuns(stderr, result)
		return 0
	case errors.Is(beginErr, client.ErrUndocumentedStage), errors.Is(beginErr, client.ErrStageMarkRejected):
		// The store was reached and answered "no" -- either the same
		// mistake this CLI's own stages.Validate call above should have
		// already caught (ErrUndocumentedStage), or some other rejection
		// of the request (ErrStageMarkRejected, e.g. a missing required
		// field). Report it and exit non-zero either way: this is the
		// store answering correctly, not a reason to fall back.
		fmt.Fprintf(stderr, "flow: stage begin refused: %v\n", beginErr)
		return 1
	default:
		journalStageMark(projectKey, journalName(f), "begin", req, stderr)
		return 0
	}
}

// journalName returns the name a stage mark's journal path is keyed by:
// f.name when set, else a synthetic "plan-<jira key, lowercased>" name for
// a -jira-key mark, which carries no change name of its own --
// journalStageMark's own doc comment requires a non-empty name for the
// journal path.
func journalName(f stageIdentityFlags) string {
	if f.name != "" {
		return f.name
	}
	return "plan-" + strings.ToLower(f.jiraKey)
}

// warnSupersededRuns prints the daemon's supersession report (KAN-618):
// a begin that landed while this session's earlier stage run was still
// open is a marking slip -- that run's end mark never landed -- and the
// superseded rows it leaves read as gaps in the work unless the write
// itself names them. A warning, never a refusal: the run is recorded
// exactly as asked.
func warnSupersededRuns(stderr io.Writer, result client.BeginStageResult) {
	if len(result.Superseded) == 0 {
		return
	}
	units := "s"
	if len(result.Superseded) == 1 {
		units = ""
	}
	fmt.Fprintf(stderr, "flow: warning: this stage begin superseded %d still-open run%s of the same session (the end mark never landed):\n", len(result.Superseded), units)
	for _, r := range result.Superseded {
		fmt.Fprintf(stderr, "  %s %s attempt %d (stage run %d)\n", r.Command, r.Stage, r.Attempt, r.ID)
	}
}

// beginStage calls the store's stage-begin endpoint under addr/timeout,
// recovering from any panic in the client path and reporting it as
// client.ErrUnavailable -- the never-block guarantee has to survive a
// panic exactly as runStateSet's putChange already does.
func beginStage(ctx context.Context, addr string, timeout time.Duration, req client.BeginStageRequest) (result client.BeginStageResult, err error) {
	defer func() {
		if r := recover(); r != nil {
			result, err = client.BeginStageResult{}, fmt.Errorf("%w: recovered panic: %v", client.ErrUnavailable, r)
		}
	}()

	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cl := client.New(addr, &http.Client{Timeout: timeout})
	return cl.BeginStage(reqCtx, req)
}

// runStageEnd implements `flow stage end`. It resolves no stage run id
// itself -- the daemon finds the currently open run for
// (project, change, command, stage), exactly as design.md's own example
// (`flow stage end --change ... --outcome completed`) never names one
// either. Its never-block and fallback behaviour mirror runStageBegin's
// exactly.
func runStageEnd(ctx context.Context, args []string, stderr io.Writer) int {
	fset := flag.NewFlagSet("flow stage end", flag.ContinueOnError)
	fset.SetOutput(stderr)
	var f stageIdentityFlags
	registerStageIdentityFlags(fset, &f)
	outcome := fset.String("outcome", "", "the stage's outcome, e.g. completed")
	fixRounds := fset.Int("fix-rounds", -1, "metrics.fix_rounds, if this stage tracked fix rounds")
	panelRounds := fset.Int("panel-rounds", -1, "metrics.panel_rounds, if this stage ran a review panel")
	findings := fset.String("findings", "", "metrics.findings_by_severity, as a JSON object")
	if err := fset.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintf(stderr, "flow: %v\n", err)
		fmt.Fprint(stderr, stageUsage)
		return 2
	}
	noteAddrUsage(fset, stderr, f.addr)
	if err := finishStageIdentityFlags(fset, &f); err != nil {
		fmt.Fprintf(stderr, "flow: %v\n", err)
		fmt.Fprint(stderr, stageUsage)
		return 2
	}
	if *outcome == "" {
		fmt.Fprintln(stderr, "flow: -outcome is required")
		fmt.Fprint(stderr, stageUsage)
		return 2
	}

	if err := stages.Validate(stages.Command(f.command), f.stage); err != nil {
		fmt.Fprintf(stderr, "flow: %v\n", err)
		return 2
	}

	metrics, err := buildEndMetrics(*fixRounds, *panelRounds, *findings)
	if err != nil {
		fmt.Fprintf(stderr, "flow: %v\n", err)
		fmt.Fprint(stderr, stageUsage)
		return 2
	}

	projectKey, _, err := fallback.ProjectKey(f.dir)
	if err != nil {
		fmt.Fprintf(stderr, "flow: resolve project key: %v\n", err)
		return 1
	}

	req := client.EndStageRequest{
		ProjectKey: projectKey,
		ChangeName: f.name,
		Command:    f.command,
		Stage:      f.stage,
		EndedAt:    time.Now(),
		Outcome:    *outcome,
		Metrics:    metrics,
		JiraKey:    f.jiraKey,
	}

	_, endErr := endStage(ctx, f.addr, f.timeout, req)
	switch {
	case endErr == nil:
		return 0
	case errors.Is(endErr, client.ErrUndocumentedStage), errors.Is(endErr, client.ErrStageMarkRejected):
		fmt.Fprintf(stderr, "flow: stage end refused: %v\n", endErr)
		return 1
	case errors.Is(endErr, client.ErrNotFound):
		// The store was reached and answered definitively: no open
		// stage run matches. That is a bookkeeping slip, not an
		// outage -- name the missing begin, and journal nothing:
		// replaying an end for a run that never opened can only fail
		// the same way again (KAN-700).
		warnNoOpenStageRun(stderr, journalName(f), f.stage)
		return 0
	default:
		// ErrUnavailable or ErrRefused: neither may block the
		// pipeline. (ErrNotFound no longer lands here -- a definitive
		// no-open-run answer names the missing begin above.)
		journalStageMark(projectKey, journalName(f), "end", req, stderr)
		return 0
	}
}

// buildEndMetrics assembles the metrics patch `stage end`'s own flags
// describe -- fix_rounds and panel_rounds only when their flag was
// explicitly given (a negative sentinel default distinguishes "not
// given" from "given as 0", since 0 fix rounds is a real, meaningful
// value design.md's own metrics table documents), and
// findings_by_severity only when -findings was given, validated as JSON
// before it is ever sent. Returns nil (no metrics call at all) when none
// of the three were given -- MergeMetrics requires a non-nil patch
// (store.ErrNilMetricsPatch), and an empty PATCH would be a wasted round
// trip for a `stage end` that only carries an outcome.
//
// This deliberately carries no harness parameter, and never has:
// tokens_available -- whether this stage run's token metrics can ever be
// harvested -- is derived server-side, from the harness `stage begin`
// already recorded on the row, not from anything an end mark supplies.
// See internal/api's ApplyEndStageMark and its withTokensUnavailable
// helper for where that now lives, and their doc comments for why (task
// 10's post-commit review, finding F1): an end mark carrying its own,
// separately-resolved harness could disagree with the one actually
// recorded at begin, and design.md's own canonical
// `flow stage end --change ... --outcome completed` example never
// passes -harness at all, so a claude-code run ended without it would
// have been marked unavailable despite being genuinely measured.
func buildEndMetrics(fixRounds, panelRounds int, findings string) (json.RawMessage, error) {
	patch := map[string]json.RawMessage{}
	if fixRounds >= 0 {
		patch["fix_rounds"] = json.RawMessage(fmt.Sprintf("%d", fixRounds))
	}
	if panelRounds >= 0 {
		patch["panel_rounds"] = json.RawMessage(fmt.Sprintf("%d", panelRounds))
	}
	if findings != "" {
		if !json.Valid([]byte(findings)) {
			return nil, fmt.Errorf("-findings must be valid JSON, got %q", findings)
		}
		patch["findings_by_severity"] = json.RawMessage(findings)
	}
	if len(patch) == 0 {
		return nil, nil
	}
	out, err := json.Marshal(patch)
	if err != nil {
		return nil, fmt.Errorf("encode metrics: %w", err)
	}
	return out, nil
}

// endStage calls the store's stage-end endpoint under addr/timeout,
// recovering from any panic in the client path exactly as beginStage does.
func endStage(ctx context.Context, addr string, timeout time.Duration, req client.EndStageRequest) (result client.BeginStageResult, err error) {
	defer func() {
		if r := recover(); r != nil {
			result, err = client.BeginStageResult{}, fmt.Errorf("%w: recovered panic: %v", client.ErrUnavailable, r)
		}
	}()

	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cl := client.New(addr, &http.Client{Timeout: timeout})
	return cl.EndStage(reqCtx, req)
}

// runStageKeys implements `flow stage keys`: the served stage-key
// vocabulary, one key per line, exit 0. It is deliberately pure local
// output -- no store contact, no project resolution, no flags -- because
// its consumer, the check-stage-mark-calls guard, runs it from any
// checkout at lint time and must get the checked-out tree's own
// internal/stages table, never a daemon's and never an installed binary's.
// An unexpected positional argument is the one usage error it reports:
// the same exit-2 contract `flow state list` holds, so a mistyped
// invocation fails loudly instead of printing a vocabulary nobody asked
// for.
func runStageKeys(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 {
		fmt.Fprintf(stderr, "flow: stage keys takes no positional arguments\n")
		fmt.Fprint(stderr, stageUsage)
		return 2
	}
	for _, key := range stages.Keys() {
		fmt.Fprintln(stdout, key)
	}
	return 0
}

// runStageWrap implements `flow stage wrap`: one call that marks begin,
// runs the work named after `--`, and marks end -- so a missing `end`
// becomes impossible rather than merely detectable (KAN-323).
//
// The never-block guarantee shapes every store interaction here: a store
// failure on either half journals and warns and the work still runs, and
// the wrapper's exit code is the child's own in every post-child path --
// the wrapper records the work, it never alters it. Caller mistakes
// (usage errors, an undocumented stage key) are the one exception, exactly
// as runStageBegin already draws the line: exit non-zero before anything
// runs, so the defect is surfaced rather than swallowed by a fallback.
func runStageWrap(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	// The terminating signal is observed as it arrives, from the first
	// line of the wrapper (F8, F11): signal.Notify multicasts to every
	// registered channel, so this one receives the signal main's
	// NotifyContext consumed in parallel. The residual blind window is
	// the process's own startup up to this registration -- microseconds
	// of runtime initialisation, not the parse-and-begin work below --
	// and a signal inside it exits 143 rather than the signal-specific
	// 130/143 convention.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, handledSignals...)
	defer signal.Stop(sigCh)

	sep := -1
	for i, a := range args {
		if a == "--" {
			sep = i
			break
		}
	}
	if sep < 0 {
		fmt.Fprintln(stderr, `flow: stage wrap requires "--" followed by the work command to run`)
		fmt.Fprint(stderr, stageUsage)
		return 2
	}
	childArgs := args[sep+1:]
	if len(childArgs) == 0 {
		fmt.Fprintln(stderr, `flow: stage wrap requires a work command after "--"`)
		fmt.Fprint(stderr, stageUsage)
		return 2
	}

	fset := flag.NewFlagSet("flow stage wrap", flag.ContinueOnError)
	fset.SetOutput(stderr)
	var f stageIdentityFlags
	registerStageIdentityFlags(fset, &f)
	harnessFlag := fset.String("harness", "", "the harness running this mark (default: $FLOW_HARNESS, or \"unknown\")")
	sessionFlag := fset.String("session", "", "the harness session id, if known; defaults to CLAUDE_CODE_SESSION_ID when set")
	sessionTokenFlag := fset.String("session-token", "", "a literal, unique token this run generates once and passes unchanged on every mark it makes (required)")
	if err := fset.Parse(args[:sep]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintf(stderr, "flow: %v\n", err)
		fmt.Fprint(stderr, stageUsage)
		return 2
	}
	noteAddrUsage(fset, stderr, f.addr)
	if err := finishStageIdentityFlags(fset, &f); err != nil {
		fmt.Fprintf(stderr, "flow: %v\n", err)
		fmt.Fprint(stderr, stageUsage)
		return 2
	}
	if *sessionTokenFlag == "" {
		fmt.Fprintln(stderr, "flow: -session-token is required")
		fmt.Fprint(stderr, stageUsage)
		return 2
	}
	if err := validateSessionToken(*sessionTokenFlag); err != nil {
		fmt.Fprintf(stderr, "flow: %v\n", err)
		return 2
	}

	if err := stages.Validate(stages.Command(f.command), f.stage); err != nil {
		fmt.Fprintf(stderr, "flow: %v\n", err)
		return 2
	}

	projectKey, mainCheckout, err := fallback.ProjectKey(f.dir)
	if err != nil {
		fmt.Fprintf(stderr, "flow: resolve project key: %v\n", err)
		return 1
	}

	var sessionID *string
	switch {
	case *sessionFlag != "":
		sessionID = sessionFlag
	default:
		if v := os.Getenv("CLAUDE_CODE_SESSION_ID"); v != "" {
			sessionID = &v
		}
	}
	beginReq := client.BeginStageRequest{
		ProjectKey:       projectKey,
		MainCheckoutPath: mainCheckout,
		ChangeName:       f.name,
		Harness:          resolveHarness(*harnessFlag),
		SessionID:        sessionID,
		SessionToken:     *sessionTokenFlag,
		Command:          f.command,
		Stage:            f.stage,
		StartedAt:        time.Now(),
		JiraKey:          f.jiraKey,
	}

	// A store refusal on begin is a caller defect -- the same one
	// runStageBegin exits 1 for -- and the child is not run behind it: the
	// marks name a stage run that would never close cleanly. Any other
	// failure journals and the work still runs.
	_, beginErr := beginStage(ctx, f.addr, f.timeout, beginReq)
	switch {
	case beginErr == nil:
	case errors.Is(beginErr, client.ErrUndocumentedStage), errors.Is(beginErr, client.ErrStageMarkRejected):
		fmt.Fprintf(stderr, "flow: stage wrap refused: %v\n", beginErr)
		return 1
	default:
		journalStageMark(projectKey, journalName(f), "begin", beginReq, stderr)
	}

	// The argv after -- is the caller's own work command, the whole point
	// of the subcommand: it runs directly, never shell-interpolated.
	// CommandContext ties the child to the ctx main derives from
	// signal.NotifyContext (F1): a SIGINT/SIGTERM to the wrapper cancels
	// ctx, whose Cancel forwards SIGTERM to the child, so a harness
	// timeout-kill terminates the whole invocation instead of hanging
	// until the child finishes on its own. WaitDelay bounds that
	// teardown (F9): a child ignoring SIGTERM is force-killed after
	// defaultTimeout, the same bound every store call here already
	// carries, so Run cannot block on a stubborn child indefinitely.
	cmd := exec.CommandContext(ctx, childArgs[0], childArgs[1:]...)
	cmd.Cancel = func() error { return cmd.Process.Signal(syscall.SIGTERM) }
	cmd.WaitDelay = defaultTimeout
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	childErr := cmd.Run()
	exitCode := 0
	switch {
	case childErr == nil:
	case errors.As(childErr, new(*exec.ExitError)):
		if ws, ok := childErr.(*exec.ExitError).Sys().(syscall.WaitStatus); ok && ws.Signaled() {
			// A signal-killed child has no exit code; the shell
			// convention stands in for one: 128+signal (F2), never
			// the raw -1 an unsigned byte would surface as 255.
			exitCode = 128 + int(ws.Signal())
		} else {
			exitCode = childErr.(*exec.ExitError).ExitCode()
		}
	default:
		// The child never started (no such command, no permission):
		// 127 is the convention a shell uses for the same failure. A
		// cancellation is not a spawn failure -- the wrapper was
		// signalled, the ctx.Err() branch below reports that with the
		// 128+signal convention, and printing a run-the-work error
		// would misread it (F11).
		if ctx.Err() == nil {
			fmt.Fprintf(stderr, "flow: stage wrap: run the work: %v\n", childErr)
			exitCode = 127
		}
	}
	if ctx.Err() != nil {
		// The wrapper itself was signalled. The child has been torn down
		// above; the wrapper must not outlive the signal that ended it,
		// and its own exit code follows the same 128+signal convention:
		// 130 for SIGINT, 143 for SIGTERM, 143 when no signal of the two
		// was observed (only signals cancel main's ctx in practice).
		exitCode = 143
		select {
		case s := <-sigCh:
			if s == os.Interrupt {
				exitCode = 130
			}
		default:
		}
	}

	endReq := client.EndStageRequest{
		ProjectKey: projectKey,
		ChangeName: f.name,
		Command:    f.command,
		Stage:      f.stage,
		EndedAt:    time.Now(),
		Outcome:    stages.OutcomeCompleted,
		JiraKey:    f.jiraKey,
	}
	if exitCode != 0 {
		endReq.Outcome = "failed"
	}

	// Past this line the work has run, so no store failure may alter its
	// reported result: the exit code below is the child's, whatever the
	// end half does. A refusal is printed and swallowed here (unlike
	// runStageEnd's own exit 1) because the work's result outranks the
	// mark's rejection once the work is already done. The end call runs
	// on a fresh context -- a ctx cancelled by the signal that killed
	// the child must not force the one mark that records the work's
	// actual outcome into the journal; -timeout still bounds it.
	_, endErr := endStage(context.Background(), f.addr, f.timeout, endReq)
	switch {
	case endErr == nil:
	case errors.Is(endErr, client.ErrUndocumentedStage), errors.Is(endErr, client.ErrStageMarkRejected):
		fmt.Fprintf(stderr, "flow: stage end refused: %v\n", endErr)
	case errors.Is(endErr, client.ErrNotFound):
		// A definitive no-open-run answer names the missing begin
		// (KAN-700), exactly as runStageEnd's own case does, and
		// journals nothing.
		warnNoOpenStageRun(stderr, journalName(f), f.stage)
	default:
		journalStageMark(projectKey, journalName(f), "end", endReq, stderr)
	}
	return exitCode
}
