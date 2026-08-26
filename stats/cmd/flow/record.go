package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/tweety53/agents/stats/internal/client"
	"github.com/tweety53/agents/stats/internal/fallback"
	"github.com/tweety53/agents/stats/internal/records"
)

// recordRoles is the closed set of roles a dispatch row may record --
// design.md's own `implementer · reviewer · panel-fix · red-partner`. It is
// checked here, before the store is ever contacted, for exactly the reason
// `stage begin` checks its stage key against internal/stages first: an
// unrecognised role is a defect in the caller, and a caller mistake taking
// the never-block fallback path would journal a write that a replay could
// only ever be refused for a second time.
//
// Unlike a stage key, a role has no documented table elsewhere to
// transcribe -- design.md's schema comment is the whole of it -- so the
// list lives here rather than behind a package of its own.
var recordRoles = []string{"implementer", "reviewer", "panel-fix", "red-partner"}

// validateFindingStatus judges a finding's status the way validateRole
// judges -role, before the store is ever contacted: "open" and "fixed" are
// the two terminal words, and "withdrawn" is legal only carrying a reason
// -- a bare "withdrawn" (or one trailing only whitespace) names nothing a
// reader could act on, so it is refused here rather than stored.
func validateFindingStatus(status string) error {
	if status == "open" || status == "fixed" {
		return nil
	}
	if rest, ok := strings.CutPrefix(status, "withdrawn"); ok {
		first, _ := utf8.DecodeRuneInString(rest)
		if unicode.IsSpace(first) && strings.TrimSpace(rest) != "" {
			return nil
		}
	}
	return fmt.Errorf("-status %q is not one of: open, fixed, withdrawn <reason>", status)
}

// validateFindingReproducer judges a finding's -reproducer before the store
// is contacted. Empty is always an error -- a finding with no reproducer at
// all is a finding nobody can act on. Where the first word is exactly
// "none" the only legal form is "none — <reason>": a bare "none" claims
// irreproducibility without saying why, which is the one shape this change's
// `reproducer-safety-in-shell` decision still refuses here rather than
// leaving to the guard. Any other non-empty value is a command, and this
// validator has nothing further to say about its shape -- metacharacters,
// absolute paths and ".." segments stay a guard-side check.
func validateFindingReproducer(reproducer string) error {
	if strings.TrimSpace(reproducer) == "" {
		return fmt.Errorf("-reproducer is required")
	}
	if fields := strings.Fields(reproducer); len(fields) > 0 && fields[0] == "none" {
		if rest, ok := strings.CutPrefix(reproducer, "none — "); !ok || strings.TrimSpace(rest) == "" {
			return fmt.Errorf("-reproducer %q must be a command, or \"none — <reason>\"", reproducer)
		}
	}
	return nil
}

const recordUsage = `usage: flow record dispatch begin [-addr url] [-timeout dur] [-C dir]
                             -change name [-task id] -role role [-slot name]
                             -model model [-agent-id id] [-diff-base sha]
                             -key key -session-token token -started-at rfc3339
       flow record dispatch end   [-addr url] [-timeout dur] [-C dir]
                             -change name -key key -session-token token
                             [-commit sha] [-outcome outcome] [-agent-id id]
                             -ended-at rfc3339
       flow record finding  [-addr url] [-timeout dur] [-C dir]
                             -change name -ref F<n> [-round n] -slot name
                             -severity sev [-location loc] -status status
                             [-reproducer cmd] [-dispatch-seq n] -note text
       flow record status   [-addr url] [-timeout dur] [-C dir]
                             -change name -ref F<n> -status status
       flow record findings [-addr url] [-timeout dur] [-C dir]
                             -change name
       flow record render   [-addr url] [-timeout dur] [-C dir]
                             -change name -kind ledger|panel|all -repo dir
       flow record journal-count [-C dir] -change name
       flow record cost-status [-addr url] [-timeout dur] [-C dir] -change name

A record write never blocks: on any store failure the intent is journalled,
one warning line is printed, and the command exits 0. A caller must never
branch on this command's exit code as a signal about the record.

journal-count prints how many of those journalled writes are still pending
for a change -- one decimal count on stdout, and "unknown" where no count
could be produced. It never blocks either, and for a sharper reason: it
exists so a handoff can state the number honestly, so it must never itself
become the reason the handoff does not print. It takes no -addr, because it
reads the local journal and never contacts the store.

cost-status prints, on one stdout line, how many of a change's dispatches
carry no cost figure and why: "N unattributed" when N is zero, and
"N unattributed — <reason>: <count>[, <reason>: <count>...]" (reasons sorted
alphabetically, for a deterministic line) when N is greater than zero.
Unlike journal-count it DOES contact the store -- only the store can answer
what it holds -- but it carries the identical never-block guarantee for the
identical reason: on any failure to reach or read the store it prints
"unknown" and exits 0, rather than putting a handoff's own output behind a
store that has no other reason to be up.

findings prints one change's findings as a JSON array on stdout -- ref,
status, and reproducer among the fields -- for a guard to query instead of
parsing the Markdown render writes. A change the store has never heard of
prints exactly "[]" and exits 0: no rows is a fact, not a failure. Unlike
every write above, a failed read never journals and never exits 0 -- there
is nothing to replay, so a store findings could not reach is reported to
stderr and exits non-zero rather than printing a JSON array a caller could
mistake for "no findings."

The only non-zero exits are caller mistakes -- a missing required flag, an
unrecognised -role, or a -session-token carrying a shell substitution --
a write the store was reached for and refused, and a read (findings) the
store could not answer.

A dispatch is recorded in TWO calls. "begin" writes the row as the
dispatch starts; "end" closes it as the dispatch finishes. Both are
required: the harvester commits its transcript offset every few seconds
and never re-reads past it, so a row that appears only at close is
invisible to every harvest tick the dispatch ran through -- that usage is
dropped or credited to an unrelated earlier dispatch -- and a row never
closed leaves an open window that goes on claiming its successors' tokens.

-key names the dispatch within its run. It is what "end" closes, and what
makes "begin" idempotent: a write whose response was lost is journalled
and replayed carrying the identical key, so it updates the row the first
attempt inserted instead of inserting a second row for one dispatch. Write
it as a literal, unique among the dispatches of one -session-token.

-role is one of: implementer, reviewer, panel-fix, red-partner.

-agent-id is the harness's own identifier for the subagent that was
dispatched, where the harness exposes one. It is optional because two of
the three supported harnesses expose none at all: a dispatch recorded
without it is ordinary, not degraded, and its cost is attributed by the
dispatch's own time window instead. Giving it is what lets two slots
dispatched at once be costed separately, since their windows overlap. It
may be given on "begin", on "end", or both: on Claude Code the harness
reports it only once the dispatch has launched, so "begin" cannot always
carry it and "end" may carry it instead. Given on "end", it overwrites
whatever "begin" recorded; omitted on "end", it leaves that value
untouched -- an "end" that omits it never clears an identifier already
recorded.

-session-token must be a literal, unique token this command writes -- never
a shell substitution ("$(...)", a backtick, or "$VAR"): the transcript
records the command text before the shell expands it, so a substitution
would be recorded identically by every caller and identify nothing. See
validateSessionToken (stage.go) for the whole of that reasoning, including
the shape no check at this layer can catch.
`

// runRecord implements `flow record`. Every subcommand but `render` and
// `journal-count` writes a record and shares the never-block fallback;
// `render` alone reads one back and turns it into Markdown.
func runRecord(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, recordUsage)
		return 2
	}

	switch args[0] {
	case "dispatch":
		return runRecordDispatchVerb(ctx, args[1:], stdout, stderr)
	case "finding":
		return runRecordFinding(ctx, args[1:], stdout, stderr)
	case "findings":
		return runRecordFindings(ctx, args[1:], stdout, stderr)
	case "status":
		return runRecordStatus(ctx, args[1:], stdout, stderr)
	case "render":
		return runRecordRender(ctx, args[1:], stdout, stderr)
	case "journal-count":
		return runRecordJournalCount(args[1:], stdout, stderr)
	case "cost-status":
		return runRecordCostStatus(ctx, args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "flow: unknown record command %q\n", args[0])
		fmt.Fprint(stderr, recordUsage)
		return 2
	}
}

// recordIdentityFlags is common to every record subcommand: the store
// connection, the working directory the project key resolves from, and the
// change the record belongs to.
//
// The change is a flag here, not the positional argument `stage begin`
// takes, because design.md's own flag set writes it that way -- a record
// subcommand carries no positional argument at all, so a stray one is a
// mistyped flag rather than a change name and is reported as such.
type recordIdentityFlags struct {
	addr    string
	timeout time.Duration
	dir     string
	change  string
}

func registerRecordIdentityFlags(fset *flag.FlagSet, f *recordIdentityFlags) {
	fset.StringVar(&f.addr, "addr", resolveDefaultAddr(), "flowd base URL")
	fset.DurationVar(&f.timeout, "timeout", defaultTimeout, "store request timeout before falling back")
	fset.StringVar(&f.dir, "C", "", "resolve the project key as if run from this directory (default: cwd)")
	fset.StringVar(&f.change, "change", "", "the change this record belongs to (required)")
}

func finishRecordIdentityFlags(fset *flag.FlagSet, f *recordIdentityFlags) error {
	if fset.NArg() != 0 {
		return fmt.Errorf("expected no positional arguments; the change is named by -change")
	}
	if f.change == "" {
		return fmt.Errorf("-change is required")
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

// parseRecordFlags parses args into fset and finishes the identity flags,
// reporting either failure the one way every record subcommand reports a
// caller mistake: the message, the usage block, exit 2. It returns whether
// parsing succeeded alongside the exit code, since flag.ErrHelp is a
// successful non-continuation rather than a mistake.
func parseRecordFlags(fset *flag.FlagSet, f *recordIdentityFlags, args []string, stderr io.Writer) (ok bool, code int) {
	if err := fset.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return false, 0
		}
		fmt.Fprintf(stderr, "flow: %v\n", err)
		fmt.Fprint(stderr, recordUsage)
		return false, 2
	}
	noteAddrEnvUsage(fset, stderr)
	if err := finishRecordIdentityFlags(fset, f); err != nil {
		fmt.Fprintf(stderr, "flow: %v\n", err)
		fmt.Fprint(stderr, recordUsage)
		return false, 2
	}
	return true, 0
}

// requireRecordFlags reports the first of flags whose value is empty as the
// caller mistake it is. Each pair is the flag's name and the value the
// caller gave it, in the order the usage block lists them, so a caller
// missing several is told about the first one they would fix.
func requireRecordFlags(stderr io.Writer, pairs ...[2]string) bool {
	for _, p := range pairs {
		if p[1] == "" {
			fmt.Fprintf(stderr, "flow: %s is required\n", p[0])
			fmt.Fprint(stderr, recordUsage)
			return false
		}
	}
	return true
}

// recordJournalPath is where a record write's fallback intent is
// journalled: beside the change's state file and its state journal
// (fallback.JournalFilePath), under a third distinct filename, exactly as
// stageJournalPath takes a second.
//
// The three files are deliberately not one. internal/reconcile decodes
// each one's entries as a different shape -- the state journal's as a
// whole change PUT body with DisallowUnknownFields, the stage journal's as
// a stage mark -- so a record body appended to either would make every
// entry after it fail to decode on replay. Using a third sibling file
// instead reuses fallback's own AppendJournalEntry unchanged, without
// touching either existing replay contract.
func recordJournalPath(projectKey, name string) string {
	return fallback.JournalFilePath(projectKey, name) + ".record"
}

// recordJournalBody is what gets journalled for a record write that could
// not reach the store: the write's own kind ("dispatch", "finding" or
// "status") alongside the exact wire request that would have been sent, so
// the reconciler has everything it needs to replay it without this file
// needing a second encoding. It is the same shape stageMarkJournalBody
// carries, for the same reason.
type recordJournalBody struct {
	Kind    string `json:"kind"`
	Request any    `json:"request"`
}

// journalRecordWrite appends kind/req to the record journal for
// projectKey/name and prints the one warning line every fallback path
// prints. Errors from the journal write itself are deliberately swallowed,
// exactly as journalStageMark swallows them: there is nothing further to
// fall back to, and the guarantee this exists to uphold is "never block",
// not "never lose telemetry".
func journalRecordWrite(projectKey, name, kind string, req any, stderr io.Writer) {
	body, err := json.Marshal(recordJournalBody{Kind: kind, Request: req})
	if err == nil {
		_ = fallback.AppendJournalEntry(recordJournalPath(projectKey, name), projectKey, name, body, time.Now())
	}
	fmt.Fprintln(stderr, "⚠ flow: store unreachable — wrote local journal")
}

// callRecord runs one client call under addr/timeout, recovering from any
// panic in the client path and reporting it as client.ErrUnavailable --
// the never-block guarantee has to survive a panic exactly as beginStage's
// own recovery makes it survive one.
//
// It is generic over the call's result so the three write subcommands
// share one copy of the timeout, the client construction and the recovery,
// rather than one copy each. beginStage and endStage predate it and are
// left alone: rewriting a shipped, tested path to share this would be a
// refactor this task was not asked for.
func callRecord[T any](ctx context.Context, addr string, timeout time.Duration, fn func(context.Context, *client.Client) (T, error)) (out T, err error) {
	defer func() {
		if r := recover(); r != nil {
			var zero T
			out, err = zero, fmt.Errorf("%w: recovered panic: %v", client.ErrUnavailable, r)
		}
	}()

	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cl := client.New(addr, &http.Client{Timeout: timeout})
	return fn(reqCtx, cl)
}

// classifyRecordWrite turns one record call's error into this command's
// exit code, journalling the intent when -- and only when -- the store
// could not be reached.
//
// The split is the delta spec's: client.ErrRecordRejected (the daemon
// answered 400) and client.ErrNotFound (it answered 404) are the store
// answering correctly, so they are reported and exit non-zero. A replay of
// either would be refused identically, which is precisely why journalling
// them would be worse than not recording them at all: it would queue a
// write that can never succeed, forever. Everything else -- a dead port, a
// timeout, a look-alike server on the configured port, a 500 -- is a store
// that could not be reached, which journals, warns once and exits 0.
func classifyRecordWrite(err error, projectKey, change, kind string, req any, stderr io.Writer) int {
	switch {
	case err == nil:
		return 0
	case errors.Is(err, client.ErrRecordRejected), errors.Is(err, client.ErrNotFound):
		fmt.Fprintf(stderr, "flow: record %s refused: %v\n", kind, err)
		return 1
	default:
		journalRecordWrite(projectKey, change, kind, req, stderr)
		return 0
	}
}

// runRecordDispatchVerb routes `flow record dispatch` to its two halves.
// Neither is optional and neither is a variant of the other: `begin` opens
// the row and `end` closes it, and a run that performs one without the
// other produces a dispatch whose cost is wrong in a specific, silent way
// (runRecordDispatchBegin and runRecordDispatchEnd each say how).
func runRecordDispatchVerb(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, recordUsage)
		return 2
	}
	switch args[0] {
	case "begin":
		return runRecordDispatchBegin(ctx, args[1:], stdout, stderr)
	case "end":
		return runRecordDispatchEnd(ctx, args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "flow: unknown record dispatch command %q\n", args[0])
		fmt.Fprint(stderr, recordUsage)
		return 2
	}
}

// runRecordDispatchBegin implements `flow record dispatch begin`: the
// opening call, sent as the dispatch starts, carrying everything already
// known then -- the task, the role, the slot, the model, the harness's
// agent id, the run's session token and the start instant.
//
// IT IS SENT AT THE START AND NOT AT THE CLOSE, and that is the whole
// reason this command is a pair rather than the single call it once was.
// The design this change shipped with recorded a dispatch once, at close,
// on the reasoning that "a begin/end pair would double the call count and
// buy nothing". That reasoning was wrong, and the thing it missed is the
// harvester: internal/harvest commits the transcript offset it has consumed
// every few seconds and never re-reads behind it (watcher.go). Attribution
// therefore sees each transcript record exactly once, at the tick that
// consumed it -- and a dispatch row that does not exist yet has no window
// for those records to fall in. Every tick that elapsed while a dispatch
// ran found no matching row, so its usage was dropped or, worse, credited
// to whichever earlier dispatch still had an open window. Since a subagent
// dispatch is routinely minutes long and a tick is seconds, that was most
// of the cost this table exists to record.
//
// -model is recorded intent, never derived: a slot whose model the
// dispatcher cannot read records the literal `unknown (agent-defined)`,
// which this command accepts verbatim, and never a plausible-looking slug.
// The token figures are not recorded here at all -- the harvester
// attributes them to this row afterwards, from the harness transcript.
//
// -agent-id is optional and unvalidated beyond being a string: it is the
// harness's identifier, not this tool's, and the only thing this command
// can say about it is whether it was reported. Left unset it is sent as
// nothing at all rather than as an empty value, because the attributor
// treats an absent id as "not reported" and must never pair two dispatches
// off by their shared absence.
//
// -diff-base is optional in the same shape as -agent-id: it names the sha
// the diff this dispatch was given was computed from, which a panel slot
// re-run against its own delta has and an implementer -- or a slot reading
// the whole diff -- does not. Left unset it is sent as nothing at all
// rather than as an empty value, because an absent base means the dispatch
// recorded none, and a rendered record must never show a base it invented.
//
// -key is required, and is the one flag whose value the caller invents. It
// names this dispatch within its run, so that `end` can close a row whose
// seq the caller may never have seen -- a begin that fell back to the
// journal returned no seq at all -- and so that a replayed begin collides
// with the row it already wrote rather than inserting a second one.
func runRecordDispatchBegin(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fset := flag.NewFlagSet("flow record dispatch begin", flag.ContinueOnError)
	fset.SetOutput(stderr)
	var f recordIdentityFlags
	registerRecordIdentityFlags(fset, &f)
	task := fset.String("task", "", "the task id from the tasks.md heading, where the role runs against a task")
	role := fset.String("role", "", "one of: "+strings.Join(recordRoles, ", ")+" (required)")
	slot := fset.String("slot", "", "the review-panel slot, where the role is a panel slot")
	model := fset.String("model", "", "the model this dispatch ran on, as recorded intent -- the literal \"unknown (agent-defined)\" where it cannot be read, never a guess (required)")
	agentID := fset.String("agent-id", "", "the harness's own identifier for the dispatched subagent, where it exposes one -- optional, since two of the three supported harnesses expose none")
	diffBase := fset.String("diff-base", "", "the sha the diff this dispatch was given was computed from, where it was given a delta -- optional, since an implementer and a slot reading the whole diff record none")
	key := fset.String("key", "", "this dispatch's own literal label, unique within the run -- what the end call closes, and what makes a replayed write land on one row (required)")
	sessionToken := fset.String("session-token", "", "the run's own literal session token, unchanged from the mark that opened the run -- never a shell substitution (required)")
	startedAt := fset.String("started-at", "", "when the dispatch started, RFC 3339 -- the instant the harvester attributes its cost from (required)")

	if ok, code := parseRecordFlags(fset, &f, args, stderr); !ok {
		return code
	}
	if !requireRecordFlags(stderr,
		[2]string{"-role", *role},
		[2]string{"-model", *model},
		[2]string{"-key", *key},
		[2]string{"-session-token", *sessionToken},
		[2]string{"-started-at", *startedAt},
	) {
		return 2
	}
	if !slices.Contains(recordRoles, *role) {
		fmt.Fprintf(stderr, "flow: -role %q is not one of: %s\n", *role, strings.Join(recordRoles, ", "))
		return 2
	}
	if err := validateSessionToken(*sessionToken); err != nil {
		fmt.Fprintf(stderr, "flow: %v\n", err)
		return 2
	}
	started, err := time.Parse(time.RFC3339, *startedAt)
	if err != nil {
		fmt.Fprintf(stderr, "flow: -started-at %q is not an RFC 3339 instant: %v\n", *startedAt, err)
		return 2
	}

	projectKey, _, err := fallback.ProjectKey(f.dir)
	if err != nil {
		fmt.Fprintf(stderr, "flow: resolve project key: %v\n", err)
		return 1
	}

	in := records.Dispatch{
		AgentID:      *agentID,
		Key:          *key,
		TaskID:       *task,
		Role:         *role,
		Slot:         *slot,
		Model:        *model,
		DiffBase:     *diffBase,
		SessionToken: *sessionToken,
		StartedAt:    started,
	}
	out, callErr := callRecord(ctx, f.addr, f.timeout, func(ctx context.Context, cl *client.Client) (records.Dispatch, error) {
		return cl.RecordDispatch(ctx, projectKey, f.change, in)
	})
	if callErr == nil {
		// The seq is the store's own allocation, and the identifier a
		// rendered record and a finding's -dispatch-seq name this
		// dispatch by, so it is what the caller is told. A write that
		// fell back to the journal prints no seq, because none was
		// allocated -- which is exactly why `end` closes the row by -key
		// and not by seq.
		fmt.Fprintf(stdout, "recorded: dispatch %d\n", out.Seq)
	}
	return classifyRecordWrite(callErr, projectKey, f.change, "dispatch", in, stderr)
}

// runRecordDispatchEnd implements `flow record dispatch end`: the closing
// call, sent as the dispatch finishes, carrying the three facts knowable
// only then -- the commit it produced, how it ended, and when.
//
// WITHOUT IT THE ATTRIBUTION WINDOW NEVER CLOSES. A dispatch row with no
// end instant is an open window, and an open window contains every later
// timestamp forever: the dispatch goes on being a candidate for usage that
// belongs to the dispatches after it, and the latest-start rule that
// arbitrates overlapping windows keeps a stale sibling in the running
// indefinitely.
//
// It names its row by -key and -session-token rather than by seq for the
// reason `begin` gives: a begin that fell back to the journal allocated no
// seq, so a caller may hold none, and an end that could not name its row
// would leave the window open -- the exact failure this call exists to
// prevent.
func runRecordDispatchEnd(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fset := flag.NewFlagSet("flow record dispatch end", flag.ContinueOnError)
	fset.SetOutput(stderr)
	var f recordIdentityFlags
	registerRecordIdentityFlags(fset, &f)
	key := fset.String("key", "", "the literal label this dispatch's begin call carried (required)")
	sessionToken := fset.String("session-token", "", "the run's own literal session token, unchanged from the mark that opened the run -- never a shell substitution (required)")
	commit := fset.String("commit", "", "the commit sha this dispatch produced")
	outcome := fset.String("outcome", "", "how the dispatch ended, e.g. completed")
	agentID := fset.String("agent-id", "", "the harness's own identifier for the dispatched subagent, where begin could not carry it -- optional, and never clears an identifier begin already recorded")
	endedAt := fset.String("ended-at", "", "when the dispatch ended, RFC 3339 -- the instant its attribution window closes (required)")

	if ok, code := parseRecordFlags(fset, &f, args, stderr); !ok {
		return code
	}
	if !requireRecordFlags(stderr,
		[2]string{"-key", *key},
		[2]string{"-session-token", *sessionToken},
		[2]string{"-ended-at", *endedAt},
	) {
		return 2
	}
	if err := validateSessionToken(*sessionToken); err != nil {
		fmt.Fprintf(stderr, "flow: %v\n", err)
		return 2
	}
	ended, err := time.Parse(time.RFC3339, *endedAt)
	if err != nil {
		fmt.Fprintf(stderr, "flow: -ended-at %q is not an RFC 3339 instant: %v\n", *endedAt, err)
		return 2
	}

	projectKey, _, err := fallback.ProjectKey(f.dir)
	if err != nil {
		fmt.Fprintf(stderr, "flow: resolve project key: %v\n", err)
		return 1
	}

	in := records.DispatchEnd{
		SessionToken: *sessionToken,
		Key:          *key,
		CommitSHA:    *commit,
		Outcome:      *outcome,
		EndedAt:      ended,
		AgentID:      *agentID,
	}
	out, callErr := callRecord(ctx, f.addr, f.timeout, func(ctx context.Context, cl *client.Client) (records.Dispatch, error) {
		return cl.EndDispatch(ctx, projectKey, f.change, in)
	})
	if callErr == nil {
		fmt.Fprintf(stdout, "updated: dispatch %d\n", out.Seq)
	}
	return classifyDispatchEnd(callErr, projectKey, f.change, in, stderr)
}

// classifyDispatchEnd is classifyRecordWrite with ONE exception, and the
// exception is the whole reason it exists: a 404 from this route is
// journalled rather than reported.
//
// Everywhere else in this verb a 404 is definitive -- a finding ref naming
// nothing is a typo, and a replay of it would be refused identically
// forever. Here it is not. "No dispatch under this key" has a second,
// entirely ordinary cause: the `begin` that would have created the row was
// itself journalled, and is still sitting in the same journal file ahead of
// this entry. internal/reconcile replays a record journal strictly in file
// order, so replaying the begin and then this end is exactly what resolves
// it. Reporting the 404 instead would exit non-zero, journal nothing, and
// leave the window that begin opened open forever -- reintroducing, by a
// narrower route, the very defect the end call was added to fix.
func classifyDispatchEnd(err error, projectKey, change string, in records.DispatchEnd, stderr io.Writer) int {
	if errors.Is(err, client.ErrNotFound) {
		journalRecordWrite(projectKey, change, "dispatch-end", in, stderr)
		return 0
	}
	return classifyRecordWrite(err, projectKey, change, "dispatch-end", in, stderr)
}

// runRecordFinding implements `flow record finding`: one call per
// finding, as the slot raises it. A ref is unique per change, not per
// round, so recording the same ref twice replaces the row rather than
// appending a second -- which is what makes a change's findings never
// cumulative.
//
// The two outcome words are the whole point of the daemon's 201/200 split:
// "recorded:" when the write inserted, "updated:" when it replaced. A
// panel run can then tell a finding it has just raised from one a fix
// round restated, which is a fact no caller can work out for itself.
func runRecordFinding(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fset := flag.NewFlagSet("flow record finding", flag.ContinueOnError)
	fset.SetOutput(stderr)
	var f recordIdentityFlags
	registerRecordIdentityFlags(fset, &f)
	ref := fset.String("ref", "", "the finding's reference, F<n>, unique per change (required)")
	// 0 is the initial panel and a meaningful value, so -round carries no
	// "not given" sentinel: a finding raised without one is a finding the
	// initial panel raised.
	round := fset.Int("round", 0, "the round that raised it: 0 for the initial panel, 1..n for a fix round")
	slot := fset.String("slot", "", "the panel slot that raised it (required)")
	severity := fset.String("severity", "", "the finding's severity (required)")
	location := fset.String("location", "", "where the finding is, e.g. path:line")
	status := fset.String("status", "", "one of: open, fixed, withdrawn <reason> (required)")
	reproducer := fset.String("reproducer", "", "the command that reproduces it, or \"none — <reason>\" (required)")
	note := fset.String("note", "", "the finding itself, in the slot's own words (required)")
	// A dispatch's seq starts at 1, so 0 is an unambiguous "not given":
	// a finding no single dispatch raised leaves dispatch_id NULL, which
	// design.md names as the legitimate case rather than a missing value.
	dispatchSeq := fset.Int("dispatch-seq", 0, "the seq of the dispatch that raised it, where one did")

	if ok, code := parseRecordFlags(fset, &f, args, stderr); !ok {
		return code
	}
	if !requireRecordFlags(stderr,
		[2]string{"-ref", *ref},
		[2]string{"-slot", *slot},
		[2]string{"-severity", *severity},
		[2]string{"-status", *status},
		[2]string{"-note", *note},
	) {
		return 2
	}
	if err := validateFindingStatus(*status); err != nil {
		fmt.Fprintf(stderr, "flow: %v\n", err)
		return 2
	}
	if err := validateFindingReproducer(*reproducer); err != nil {
		fmt.Fprintf(stderr, "flow: %v\n", err)
		return 2
	}

	projectKey, _, err := fallback.ProjectKey(f.dir)
	if err != nil {
		fmt.Fprintf(stderr, "flow: resolve project key: %v\n", err)
		return 1
	}

	in := records.Finding{
		Ref:        *ref,
		Round:      *round,
		Slot:       *slot,
		Severity:   *severity,
		Location:   *location,
		Note:       *note,
		Status:     *status,
		Reproducer: *reproducer,
	}
	if *dispatchSeq > 0 {
		in.DispatchSeq = dispatchSeq
	}

	created, callErr := callRecord(ctx, f.addr, f.timeout, func(ctx context.Context, cl *client.Client) (bool, error) {
		_, created, err := cl.RecordFinding(ctx, projectKey, f.change, in)
		return created, err
	})
	if callErr == nil {
		if created {
			fmt.Fprintf(stdout, "recorded: %s\n", *ref)
		} else {
			fmt.Fprintf(stdout, "updated: %s\n", *ref)
		}
	}
	return classifyRecordWrite(callErr, projectKey, f.change, "finding", in, stderr)
}

// runRecordStatus implements `flow record status`: the one column a fix
// round rewrites about a finding it has resolved, and nothing else.
func runRecordStatus(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fset := flag.NewFlagSet("flow record status", flag.ContinueOnError)
	fset.SetOutput(stderr)
	var f recordIdentityFlags
	registerRecordIdentityFlags(fset, &f)
	ref := fset.String("ref", "", "the finding's reference, F<n> (required)")
	status := fset.String("status", "", "one of: open, fixed, withdrawn <reason> (required)")

	if ok, code := parseRecordFlags(fset, &f, args, stderr); !ok {
		return code
	}
	if !requireRecordFlags(stderr, [2]string{"-ref", *ref}, [2]string{"-status", *status}) {
		return 2
	}
	if err := validateFindingStatus(*status); err != nil {
		fmt.Fprintf(stderr, "flow: %v\n", err)
		return 2
	}

	projectKey, _, err := fallback.ProjectKey(f.dir)
	if err != nil {
		fmt.Fprintf(stderr, "flow: resolve project key: %v\n", err)
		return 1
	}

	// The journalled request is the whole call, ref included: the URL
	// carries the ref on the wire, so a body carrying only the status
	// would be a replay with nothing to apply it to.
	in := recordStatusRequest{Ref: *ref, Status: *status}
	_, callErr := callRecord(ctx, f.addr, f.timeout, func(ctx context.Context, cl *client.Client) (struct{}, error) {
		return struct{}{}, cl.SetFindingStatus(ctx, projectKey, f.change, *ref, *status)
	})
	if callErr == nil {
		fmt.Fprintf(stdout, "updated: %s\n", *ref)
	}
	return classifyRecordWrite(callErr, projectKey, f.change, "status", in, stderr)
}

// runRecordJournalCount implements `flow record journal-count`: how many
// of a change's record writes are still sitting in the journal because the
// store could not be reached.
//
// IT EXISTS SO THAT A CONTRACT DOES NOT HAVE TO SPELL A PATH. The
// IN_PROGRESS handoff has to state that number whether or not anything was
// journalled -- a line printed only on a bad run is indistinguishable from
// a line nobody printed -- and without this subcommand the only way to
// produce it was to hand-derive
// <state-dir>/<project-key>/<name>.journal.record, which is exactly the
// recipe the state-file contract moved into this CLI so that skills stop
// running it by hand.
//
// IT TAKES NO -addr AND NO -timeout. It reads the local journal and never
// contacts the store, so offering a store address would suggest the count
// means something about what the store holds, which it does not: it counts
// what has NOT reached the store.
//
// IT NEVER BLOCKS, and the reasoning is one step sharper than the write
// path's. A write journals so the work it describes can proceed; this
// command's whole purpose is to make one handoff line honest, so a count
// it cannot produce says `unknown` and gets out of the way rather than
// putting the gate's own output behind a filesystem the run has no other
// reason to care about. Exit 0 for every such case; the only non-zero exit
// is a caller mistake, which the shared identity flags already report.
func runRecordJournalCount(args []string, stdout, stderr io.Writer) int {
	fset := flag.NewFlagSet("flow record journal-count", flag.ContinueOnError)
	fset.SetOutput(stderr)
	var f recordIdentityFlags
	fset.StringVar(&f.dir, "C", "", "resolve the project key as if run from this directory (default: cwd)")
	fset.StringVar(&f.change, "change", "", "the change whose pending record writes are counted (required)")

	if ok, code := parseRecordFlags(fset, &f, args, stderr); !ok {
		return code
	}

	projectKey, _, err := fallback.ProjectKey(f.dir)
	if err != nil {
		fmt.Fprintln(stdout, recordJournalCountUnknown)
		fmt.Fprintf(stderr, "flow: resolve project key: %v\n", err)
		return 0
	}

	path := recordJournalPath(projectKey, f.change)
	n, err := countRecordJournalEntries(path)
	if err != nil {
		fmt.Fprintln(stdout, recordJournalCountUnknown)
		fmt.Fprintf(stderr, "flow: count %s: %v\n", path, err)
		return 0
	}
	fmt.Fprintln(stdout, n)
	return 0
}

// recordJournalCountUnknown is the one word this subcommand prints instead
// of a number. It is deliberately not "0": zero pending writes is the
// clean run, and reporting a failed count as one would turn a filesystem
// problem into a claim that everything reached the store.
const recordJournalCountUnknown = "unknown"

// countRecordJournalEntries counts the complete entries in the record
// journal at path. An ABSENT file is 0 rather than an error -- the
// ordinary case is that every write reached the store and nothing was ever
// journalled.
//
// A PARTIAL TRAILING LINE IS NOT AN ENTRY, which is why this is not a line
// count. internal/reconcile's splitCompleteLines already draws exactly
// this line for replay: fallback.AppendJournalEntry writes an entry and
// its newline in one Write, so the only way bytes can trail the last
// newline is a process that died mid-syscall, and that span is never
// parsed and never retired. Counting it would report one more pending
// write than a replay will ever apply -- and the count is read at a gate
// whose whole job is to be believable.
//
// Blank lines are skipped for the same reason parseCompleteEntries treats
// them as blank rather than as entries: they carry no write to replay.
func countRecordJournalEntries(path string) (int, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return 0, nil
		}
		return 0, err
	}

	n := 0
	for {
		idx := bytes.IndexByte(raw, '\n')
		if idx < 0 {
			// Whatever trails the last newline is the partial write
			// above, or nothing at all. Either way it is not an entry.
			break
		}
		if len(bytes.TrimSpace(raw[:idx])) > 0 {
			n++
		}
		raw = raw[idx+1:]
	}
	return n, nil
}

// costStatusUnknown is the one word runRecordCostStatus prints instead of
// a figure, on the identical reasoning recordJournalCountUnknown carries:
// zero unattributed dispatches is the clean, ordinary answer, and reporting
// a failure to reach or read the store as that same clean answer would
// turn a store problem into a claim that every dispatch is costed.
const costStatusUnknown = "unknown"

// runRecordCostStatus implements `flow record cost-status`: how many of
// a change's dispatches carry no cost figure, and why, read from the
// store.
//
// UNLIKE journal-count, THIS VERB CAN CONTACT THE STORE. journal-count
// answers a question about the local filesystem and never does; this one
// takes the same -addr/-timeout every writing subcommand takes and reaches
// for the store when it gets far enough to need one -- but it does not
// always get that far: fallback.ProjectKey is resolved first, and its
// failure prints costStatusUnknown and returns before the store is ever
// contacted, exactly as if the store itself had failed to answer.
//
// IT NEVER BLOCKS, for the identical reason journal-count does not: it
// exists so a handoff can state a change's cost honestly, so it must never
// itself become the reason the handoff does not print. Any failure --
// a project key that cannot be resolved, an unreachable daemon, a malformed
// response, a change the store has never heard of -- prints
// costStatusUnknown and exits 0, on the single exit path below. The only
// non-zero exit is a caller mistake, which the shared identity flags
// already report.
func runRecordCostStatus(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fset := flag.NewFlagSet("flow record cost-status", flag.ContinueOnError)
	fset.SetOutput(stderr)
	var f recordIdentityFlags
	registerRecordIdentityFlags(fset, &f)

	if ok, code := parseRecordFlags(fset, &f, args, stderr); !ok {
		return code
	}

	projectKey, _, err := fallback.ProjectKey(f.dir)
	if err != nil {
		fmt.Fprintln(stdout, costStatusUnknown)
		fmt.Fprintf(stderr, "flow: resolve project key: %v\n", err)
		return 0
	}

	cs, callErr := callRecord(ctx, f.addr, f.timeout, func(ctx context.Context, cl *client.Client) (records.CostStatus, error) {
		return cl.GetCostStatus(ctx, projectKey, f.change)
	})
	if callErr != nil {
		fmt.Fprintln(stdout, costStatusUnknown)
		fmt.Fprintf(stderr, "flow: cost-status: %v\n", callErr)
		return 0
	}
	fmt.Fprintln(stdout, formatCostStatusLine(cs))
	return 0
}

// formatCostStatusLine renders cs as the one line cost-status prints.
//
// N == 0 renders `0 unattributed`, with no trailing clause: there is
// nothing to name a reason for, and a dash into an empty list would read
// as a line that lost its second half.
//
// N > 0 renders `N unattributed — <reason>: <count>[, <reason>: <count>...]`,
// one clause per reason cs.Reasons carries, SORTED ALPHABETICALLY BY REASON
// STRING. cs.Reasons is a Go map, whose iteration order is randomised by
// design, and this line can carry two reasons at once (a session that
// never bound and a dispatch ambiguity both stamp dispatches under the same
// change); printing it in whatever order the map happened to yield would
// make the same underlying figures read as a different line on every run,
// which is exactly the non-determinism a handoff line must not have.
func formatCostStatusLine(cs records.CostStatus) string {
	if cs.Unattributed == 0 {
		return fmt.Sprintf("%d unattributed", cs.Unattributed)
	}
	reasons := make([]string, 0, len(cs.Reasons))
	for reason := range cs.Reasons {
		reasons = append(reasons, reason)
	}
	sort.Strings(reasons)
	clauses := make([]string, len(reasons))
	for i, reason := range reasons {
		clauses[i] = fmt.Sprintf("%s: %d", reason, cs.Reasons[reason])
	}
	return fmt.Sprintf("%d unattributed — %s", cs.Unattributed, strings.Join(clauses, ", "))
}

// recordStatusRequest is the journalled form of a status write. The wire
// PATCH carries the ref in its URL and only the status in its body, so
// this type exists to keep both in the journal entry -- a replay resolves
// the route from what it reads, never from a path this file would have to
// encode a second time.
type recordStatusRequest struct {
	Ref    string `json:"ref"`
	Status string `json:"status"`
}

// runRecordRender implements `flow record render`: the read half of this
// verb, and what makes a change's archive readable without a daemon.
//
// It reports ONE OUTCOME WORD PER KIND, and the three that exit 0 are
// three different facts rather than three shades of success:
//
//	rendered: <dest>   there was a record to write, and it was written
//	MISSING: ledger    the store holds no dispatch rows for the change
//	journalled: <kind> the store could not be reached; nothing was written
//
// MISSING IS THE LEDGER'S ALONE. A panel always renders, because a panel
// that raised no finding has to SAY so -- see renderRecordKind, which
// carries the reasoning and the requirement it comes from.
//
// `journalled:` is the odd one, and deliberately so: there is nothing to
// journal for a READ. The word records that the render did not happen and
// why, in the vocabulary the write subcommands already use, rather than
// writing an empty record that a reader could not tell from a real one.
// Non-zero keeps its single existing meaning -- a write attempted and
// refused or failed -- so a caller branching on the exit status never
// reads an empty record as a failure.
//
// THE DESTINATIONS ARE RESOLVED BEFORE THE STORE IS CONTACTED. A refused
// change name or a destination outside the repository is a caller
// mistake, and judging it after the fetch would let an unreachable store
// turn it into the exit-0 `journalled:` path -- the failure direction
// records.Destination's own protections exist to remove.
func runRecordRender(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fset := flag.NewFlagSet("flow record render", flag.ContinueOnError)
	fset.SetOutput(stderr)
	var f recordIdentityFlags
	registerRecordIdentityFlags(fset, &f)
	kind := fset.String("kind", "", "which record to render: ledger, panel, or all (required)")
	repo := fset.String("repo", "", "the repository root the record renders into (required)")

	if ok, code := parseRecordFlags(fset, &f, args, stderr); !ok {
		return code
	}
	if !requireRecordFlags(stderr, [2]string{"-kind", *kind}, [2]string{"-repo", *repo}) {
		return 2
	}

	kinds, err := resolveRenderKinds(*kind)
	if err != nil {
		fmt.Fprintf(stderr, "flow: %v\n", err)
		fmt.Fprint(stderr, recordUsage)
		return 2
	}

	// Fixed once for the whole run, so that two kinds rendered together
	// can never straddle midnight and take different dates.
	today := time.Now().UTC()
	dests := make(map[string]string, len(kinds))
	for _, k := range kinds {
		dest, err := records.Destination(*repo, k, f.change, today)
		if err != nil {
			fmt.Fprintf(stderr, "flow: %v\n", err)
			return 1
		}
		dests[k] = dest
	}

	projectKey, _, err := fallback.ProjectKey(f.dir)
	if err != nil {
		fmt.Fprintf(stderr, "flow: resolve project key: %v\n", err)
		return 1
	}

	run, callErr := callRecord(ctx, f.addr, f.timeout, func(ctx context.Context, cl *client.Client) (records.Run, error) {
		return cl.GetRunRecord(ctx, projectKey, f.change)
	})
	switch {
	case callErr == nil:
	case errors.Is(callErr, client.ErrNotFound):
		// The store was reached and has never heard of this change, which
		// is "no rows of any kind" and reports as MISSING below -- not a
		// failure. GetRunRecord already distinguishes it from a change
		// that exists and holds nothing, and both answer the render's
		// question the same way.
		run = records.Run{Change: f.change}
	default:
		for _, k := range kinds {
			fmt.Fprintf(stdout, "journalled: %s\n", k)
		}
		fmt.Fprintln(stderr, "⚠ flow: store unreachable — nothing rendered")
		return 0
	}

	for _, k := range kinds {
		body, ok := renderRecordKind(k, run)
		if !ok {
			fmt.Fprintf(stdout, "MISSING: %s — no rows for %s\n", k, f.change)
			continue
		}
		// The directory is created only now, so a kind that reports
		// MISSING leaves no empty directory behind. Destination has
		// already resolved every symlink component and required the
		// result to sit inside the repository, so this creates and
		// writes through a path nothing can re-follow.
		if err := os.MkdirAll(filepath.Dir(dests[k]), 0o755); err != nil {
			fmt.Fprintf(stderr, "flow: create %s: %v\n", filepath.Dir(dests[k]), err)
			return 1
		}
		if err := os.WriteFile(dests[k], []byte(body), 0o644); err != nil {
			fmt.Fprintf(stderr, "flow: write %s: %v\n", dests[k], err)
			return 1
		}
		fmt.Fprintf(stdout, "rendered: %s\n", dests[k])
	}
	return 0
}

// runRecordFindings implements `flow record findings`: a read-only verb
// that prints one change's findings as a JSON array, so a guard can query
// the store directly instead of re-deriving the same facts by parsing the
// Markdown `record render` writes for human readers.
//
// It marshals Run.Findings ALONE, never the whole Run -- a guard consuming
// this verb needs ref/status/reproducer, not dispatch rows it has no use
// for and no reason to be coupled to.
//
// Unlike every write subcommand, a failed call here never journals: the
// never-block guarantee exists so a write is not lost, but a read has
// nothing to replay -- journalling a GET would queue a call that produces
// no new information the second time either. So a store the call could not
// reach is reported to stderr and this exits non-zero, exactly the
// opposite of `record render`'s fallback, which prints "journalled:" and
// exits 0. A caller must never be told "no findings" for a question this
// command could not actually answer.
func runRecordFindings(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fset := flag.NewFlagSet("flow record findings", flag.ContinueOnError)
	fset.SetOutput(stderr)
	var f recordIdentityFlags
	registerRecordIdentityFlags(fset, &f)

	if ok, code := parseRecordFlags(fset, &f, args, stderr); !ok {
		return code
	}

	projectKey, _, err := fallback.ProjectKey(f.dir)
	if err != nil {
		fmt.Fprintf(stderr, "flow: resolve project key: %v\n", err)
		return 1
	}

	run, callErr := callRecord(ctx, f.addr, f.timeout, func(ctx context.Context, cl *client.Client) (records.Run, error) {
		return cl.GetRunRecord(ctx, projectKey, f.change)
	})
	switch {
	case callErr == nil:
	case errors.Is(callErr, client.ErrNotFound):
		// The store was reached and has never heard of this change --
		// "no rows" is a fact, not a failure, and prints as an empty
		// array below exactly as runRecordRender treats it as MISSING
		// rather than an error.
		run = records.Run{Change: f.change, Findings: []records.Finding{}}
	default:
		fmt.Fprintf(stderr, "flow: findings: %v\n", callErr)
		return 1
	}

	if run.Findings == nil {
		run.Findings = []records.Finding{}
	}

	body, err := json.Marshal(run.Findings)
	if err != nil {
		fmt.Fprintf(stderr, "flow: encode findings: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, string(body))
	return 0
}

// resolveRenderKinds turns -kind into the kinds to render. "all" renders
// every kind in one call.
//
// Three pipeline steps ask for a kind and they do not all ask for the same
// one: /myflow-do renders -kind panel at panel close, finish run 1 renders
// -kind ledger before it stages, and /myflow-do's `prUrl` push path (its
// SKILL.md section 7) renders -kind all, because a fix pushed onto an open
// PR carries both records forward at once and has no reason to name them
// separately. An operator re-rendering a change's records by hand asks for
// it for the same reason.
func resolveRenderKinds(kind string) ([]string, error) {
	if kind == "all" {
		return records.Kinds(), nil
	}
	if slices.Contains(records.Kinds(), kind) {
		return []string{kind}, nil
	}
	return nil, fmt.Errorf("-kind %q is not one of: %s, all", kind, strings.Join(records.Kinds(), ", "))
}

// renderRecordKind renders one kind, reporting whether there is a record
// to write. THE TWO KINDS ANSWER THAT QUESTION DIFFERENTLY, and the
// asymmetry is the point rather than an oversight.
//
// A LEDGER WITH NO DISPATCH ROWS IS MISSING. A change nothing was
// dispatched for genuinely has no ledger, and an empty one on disk would
// be indistinguishable from a real ledger that happened to be empty --
// the distinction the run-record requirement insists stays reportable.
//
// A PANEL WITH NO FINDINGS IS STILL A PANEL, and always renders.
// openspec/specs/myflow-review-panel-economics/spec.md requires a record
// with no total line to count as outstanding whatever else it contains,
// and says a panel that raised no finding says so with `findings-total: 0`
// -- "which is a declaration and clears, where silence is not". Reporting
// MISSING for a clean panel writes no record, so check-unfinished-work.sh
// finds none and reports OUTSTANDING for a change that is genuinely clean:
// the render would manufacture the very state the panel proved absent.
// RenderPanel already emits the zero form correctly -- the total line, no
// markers, and the matching empty reproducer block -- so the only thing
// that ever suppressed it was this rule.
//
// The command is invoked at panel close, so THAT INVOCATION is the
// evidence a panel ran. Nothing in the store has to stand in for it, and
// no sentinel row is written to make one.
func renderRecordKind(kind string, run records.Run) (string, bool) {
	switch kind {
	case "ledger":
		if len(run.Dispatches) == 0 {
			return "", false
		}
		return records.RenderLedger(run), true
	case "panel":
		return records.RenderPanel(run), true
	default:
		// Unreachable: resolveRenderKinds has already refused any other
		// value, and it is the only producer of this argument.
		return "", false
	}
}
