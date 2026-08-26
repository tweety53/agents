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
	"regexp"
	"strings"
	"time"

	"github.com/tweety53/agents/stats/internal/client"
	"github.com/tweety53/agents/stats/internal/fallback"
	"github.com/tweety53/agents/stats/internal/stages"
)

// defaultHarness is used when `stage begin` is not told which harness is
// running it (-harness, or the MYFLOW_HARNESS environment variable).
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

const stageUsage = `usage: myflow stage begin [-addr url] [-timeout dur] [-C dir] [-harness name] [-session id]
                          -command cmd -stage key -session-token token <change>
       myflow stage end [-addr url] [-timeout dur] [-C dir]
                        -command cmd -stage key -outcome outcome
                        [-fix-rounds n] [-panel-rounds n] [-findings json] <change>

-stage takes a stage KEY, not its prose name -- one of README.md's Level 1
-- the stages of each command table's Key column; an undocumented key is
rejected before it ever reaches the store.

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

func runStage(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, stageUsage)
		return 2
	}

	switch args[0] {
	case "begin":
		return runStageBegin(ctx, args[1:], stderr)
	case "end":
		return runStageEnd(ctx, args[1:], stderr)
	default:
		fmt.Fprintf(stderr, "myflow: unknown stage command %q\n", args[0])
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
}

func registerStageIdentityFlags(fset *flag.FlagSet, f *stageIdentityFlags) {
	fset.StringVar(&f.addr, "addr", resolveDefaultAddr(), "myflowd base URL")
	fset.DurationVar(&f.timeout, "timeout", defaultTimeout, "store request timeout before falling back")
	fset.StringVar(&f.dir, "C", "", "resolve the project key as if run from this directory (default: cwd)")
	fset.StringVar(&f.command, "command", "", "the myflow command this stage belongs to, e.g. /myflow-do")
	fset.StringVar(&f.stage, "stage", "", "the stage key, exactly as README.md's Level 1 table's Key column documents it -- never its prose name")
}

func finishStageIdentityFlags(fset *flag.FlagSet, f *stageIdentityFlags) error {
	if fset.NArg() != 1 {
		return fmt.Errorf("expected exactly one argument, the change name")
	}
	f.name = fset.Arg(0)
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

// resolveHarness returns harnessFlag if set, else MYFLOW_HARNESS, else
// defaultHarness -- never empty, since stage_runs.harness is NOT NULL.
func resolveHarness(harnessFlag string) string {
	if harnessFlag != "" {
		return harnessFlag
	}
	if v := os.Getenv("MYFLOW_HARNESS"); v != "" {
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
	fmt.Fprintln(stderr, "⚠ myflow: store unreachable — wrote local journal")
}

// runStageBegin implements `myflow stage begin`. It validates the stage
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
	fset := flag.NewFlagSet("myflow stage begin", flag.ContinueOnError)
	fset.SetOutput(stderr)
	var f stageIdentityFlags
	registerStageIdentityFlags(fset, &f)
	harnessFlag := fset.String("harness", "", "the harness running this mark (default: $MYFLOW_HARNESS, or \"unknown\")")
	sessionFlag := fset.String("session", "", "the harness session id, if known")
	sessionTokenFlag := fset.String("session-token", "", "a literal, unique token this run generates once and passes unchanged on every mark it makes -- never a shell substitution -- so the daemon can later bind every stage run carrying it to the session that made it (required)")
	if err := fset.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintf(stderr, "myflow: %v\n", err)
		fmt.Fprint(stderr, stageUsage)
		return 2
	}
	noteAddrEnvUsage(fset, stderr)
	if err := finishStageIdentityFlags(fset, &f); err != nil {
		fmt.Fprintf(stderr, "myflow: %v\n", err)
		fmt.Fprint(stderr, stageUsage)
		return 2
	}
	if *sessionTokenFlag == "" {
		// A missing -session-token is a caller mistake, not a stage outcome:
		// exit non-zero naming the flag, exactly as -command/-stage above
		// already do -- the one class of nonzero exit the never-block
		// rule permits (tasks.md, "Step 1: The column and the flags").
		fmt.Fprintln(stderr, "myflow: -session-token is required")
		fmt.Fprint(stderr, stageUsage)
		return 2
	}
	if err := validateSessionToken(*sessionTokenFlag); err != nil {
		fmt.Fprintf(stderr, "myflow: %v\n", err)
		return 2
	}

	if err := stages.Validate(stages.Command(f.command), f.stage); err != nil {
		fmt.Fprintf(stderr, "myflow: %v\n", err)
		return 2
	}

	projectKey, mainCheckout, err := fallback.ProjectKey(f.dir)
	if err != nil {
		fmt.Fprintf(stderr, "myflow: resolve project key: %v\n", err)
		return 1
	}

	var sessionID *string
	if *sessionFlag != "" {
		sessionID = sessionFlag
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
	}

	_, beginErr := beginStage(ctx, f.addr, f.timeout, req)
	switch {
	case beginErr == nil:
		return 0
	case errors.Is(beginErr, client.ErrUndocumentedStage), errors.Is(beginErr, client.ErrStageMarkRejected):
		// The store was reached and answered "no" -- either the same
		// mistake this CLI's own stages.Validate call above should have
		// already caught (ErrUndocumentedStage), or some other rejection
		// of the request (ErrStageMarkRejected, e.g. a missing required
		// field). Report it and exit non-zero either way: this is the
		// store answering correctly, not a reason to fall back.
		fmt.Fprintf(stderr, "myflow: stage begin refused: %v\n", beginErr)
		return 1
	default:
		journalStageMark(projectKey, f.name, "begin", req, stderr)
		return 0
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

// runStageEnd implements `myflow stage end`. It resolves no stage run id
// itself -- the daemon finds the currently open run for
// (project, change, command, stage), exactly as design.md's own example
// (`myflow stage end --change ... --outcome completed`) never names one
// either. Its never-block and fallback behaviour mirror runStageBegin's
// exactly.
func runStageEnd(ctx context.Context, args []string, stderr io.Writer) int {
	fset := flag.NewFlagSet("myflow stage end", flag.ContinueOnError)
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
		fmt.Fprintf(stderr, "myflow: %v\n", err)
		fmt.Fprint(stderr, stageUsage)
		return 2
	}
	noteAddrEnvUsage(fset, stderr)
	if err := finishStageIdentityFlags(fset, &f); err != nil {
		fmt.Fprintf(stderr, "myflow: %v\n", err)
		fmt.Fprint(stderr, stageUsage)
		return 2
	}
	if *outcome == "" {
		fmt.Fprintln(stderr, "myflow: -outcome is required")
		fmt.Fprint(stderr, stageUsage)
		return 2
	}

	if err := stages.Validate(stages.Command(f.command), f.stage); err != nil {
		fmt.Fprintf(stderr, "myflow: %v\n", err)
		return 2
	}

	metrics, err := buildEndMetrics(*fixRounds, *panelRounds, *findings)
	if err != nil {
		fmt.Fprintf(stderr, "myflow: %v\n", err)
		fmt.Fprint(stderr, stageUsage)
		return 2
	}

	projectKey, _, err := fallback.ProjectKey(f.dir)
	if err != nil {
		fmt.Fprintf(stderr, "myflow: resolve project key: %v\n", err)
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
	}

	_, endErr := endStage(ctx, f.addr, f.timeout, req)
	switch {
	case endErr == nil:
		return 0
	case errors.Is(endErr, client.ErrUndocumentedStage), errors.Is(endErr, client.ErrStageMarkRejected):
		fmt.Fprintf(stderr, "myflow: stage end refused: %v\n", endErr)
		return 1
	default:
		// ErrUnavailable, ErrNotFound (no open run -- the store was
		// reached, but the mark can no longer be attributed to anything;
		// there is nothing left to do but the same fallback as an
		// outage), or ErrRefused: none of these may block the pipeline.
		journalStageMark(projectKey, f.name, "end", req, stderr)
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
// `myflow stage end --change ... --outcome completed` example never
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
