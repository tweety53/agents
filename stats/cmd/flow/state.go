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
	"path/filepath"
	"regexp"
	"slices"
	"time"

	"github.com/tweety53/agents/stats/internal/client"
	"github.com/tweety53/agents/stats/internal/fallback"
	"github.com/tweety53/agents/stats/internal/stages"
)

// defaultAddr is flowd's default bind address (internal/config.DefaultHost
// and DefaultPort), named again here rather than imported: cmd/flow does
// not otherwise depend on internal/config, and importing it for one string
// would pull the daemon's own configuration surface into the CLI for no
// reason -- the two are allowed to agree on a string constant without
// sharing a package.
const defaultAddr = "http://127.0.0.1:4173"

// resolveDefaultAddr returns the value the three -addr flag registrations
// (two in state.go, one in stage.go) take as their default: FLOW_ADDR
// when it is set to a non-empty value, defaultAddr otherwise -- mirroring
// how internal/config.FromEnv already treats an empty environment
// variable as unset.
//
// The -addr flag itself already existed and was never the problem: it
// simply was not passed. What was missing was a way to set the address
// once per session instead of remembering it on every single command --
// FLOW_ADDR is that mechanism. An explicit -addr still wins, since the
// flag's default is all this changes.
func resolveDefaultAddr() string {
	if v := os.Getenv("FLOW_ADDR"); v != "" {
		return v
	}
	return defaultAddr
}

// noteAddrEnvUsage prints one line to stderr naming the address a command
// is about to use, but only when that address came from FLOW_ADDR rather
// than from an explicit -addr flag on this invocation: fset.Visit only
// calls back for a flag actually set on the command line, so "addr" not
// appearing there means the flag's value is exactly what
// resolveDefaultAddr() returned as its default.
//
// This exists because FLOW_ADDR is meant to be exported once per shell
// session (see README.md's "The UI-test stack"), and an export outlives
// the command that motivated it -- every `flow state`/`flow stage`
// run afterwards in that shell silently inherits it, with no other signal,
// since a successful write exits 0 the same way whether it reached the
// live daemon or a test one. Call this after fset.Parse succeeds, once
// per command, so a stray or stale FLOW_ADDR is visible before its
// effect is.
func noteAddrEnvUsage(fset *flag.FlagSet, stderr io.Writer) {
	explicit := false
	fset.Visit(func(fl *flag.Flag) {
		if fl.Name == "addr" {
			explicit = true
		}
	})
	if explicit {
		return
	}
	if v := os.Getenv("FLOW_ADDR"); v != "" {
		fmt.Fprintf(stderr, "flow: using FLOW_ADDR=%s\n", v)
	}
}

// defaultTimeout bounds how long `state get`/`state set` waits for the
// store before taking the fallback path. It is short on purpose: every
// second spent waiting on a store that turns out to be down is a second
// added to every `/flow` command's runtime, and the whole point of the
// fallback is that the pipeline must not feel an outage.
const defaultTimeout = 2 * time.Second

// maxStdinBytes caps how much `state set` reads from stdin -- the same cap
// the daemon itself enforces on a PUT body (internal/api.maxRequestBodyBytes),
// so a caller that would be refused by the daemon anyway fails the same way
// locally instead of buffering an unbounded stream first.
const maxStdinBytes = 1 << 20

const stateUsage = `usage: flow state get     [-addr url] [-timeout dur] [-C dir] <name>
       flow state set     [-addr url] [-timeout dur] [-C dir] <name>
       flow state list    [-addr url] [-timeout dur] [-C dir]
       flow state resolve [-addr url] [-timeout dur] [-C dir]

state set reads the change's whole state as JSON from stdin.
state list enumerates every change the store holds for the resolved
project. On any store failure it falls back to what the local on-disk
fallback directory holds -- necessarily partial, since that directory
only ever holds records a failed write left behind -- prints one warning
line, and still exits 0. It prints one JSON object to stdout, carrying
"source" ("store" or "fallback") and "complete" (false whenever the
fallback was used), so a partial list is never mistaken for a full one.
state resolve prints the change-name candidate set a caller resolves a
bare change name against. Source "store": every board row whose state is
not FINISHED. Source "fallback": the union of every readable fallback
record's name and every directory directly under
<main-checkout>/spectre/changes/ other than "archive", minus any name
archived under spectre/changes/archive/<name>/. It carries the same
"source"/"complete" fields as state list, plus "candidates" and
"unreadable" (both always arrays, never null).
`

func runState(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, stateUsage)
		return 2
	}

	switch args[0] {
	case "get":
		return runStateGet(ctx, args[1:], stdout, stderr)
	case "set":
		return runStateSet(ctx, args[1:], stdin, stdout, stderr)
	case "list":
		return runStateList(ctx, args[1:], stdout, stderr)
	case "resolve":
		return runStateResolve(ctx, args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "flow: unknown state command %q\n", args[0])
		fmt.Fprint(stderr, stateUsage)
		return 2
	}
}

// stateFlags is the flag.FlagSet common to `state get` and `state set`, and
// the working directory and identity it resolves. dir is where git is
// asked to resolve the project key from -- exposed as -C, mirroring git's
// own flag, so a test (or an operator working from a script) can point
// resolution somewhere other than the process's actual cwd.
type stateFlags struct {
	addr    string
	timeout time.Duration
	dir     string
	name    string
}

func parseStateFlags(fset *flag.FlagSet, args []string, stderr io.Writer) (stateFlags, error) {
	fset.SetOutput(stderr)
	f := stateFlags{}
	fset.StringVar(&f.addr, "addr", resolveDefaultAddr(), "flowd base URL")
	fset.DurationVar(&f.timeout, "timeout", defaultTimeout, "store request timeout before falling back")
	fset.StringVar(&f.dir, "C", "", "resolve the project key as if run from this directory (default: cwd)")
	if err := fset.Parse(args); err != nil {
		return stateFlags{}, err
	}
	noteAddrEnvUsage(fset, stderr)
	if fset.NArg() != 1 {
		return stateFlags{}, fmt.Errorf("expected exactly one argument, the change name")
	}
	f.name = fset.Arg(0)
	if f.dir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return stateFlags{}, fmt.Errorf("resolve working directory: %w", err)
		}
		f.dir = wd
	}
	return f, nil
}

// markSyntheticIfNeeded decodes body as a JSON object and, when its
// "updatedBy" field is exactly stages.SyntheticChangeUpdatedBy, adds
// `"synthetic": true` to the object before re-encoding. This is the
// machine-checkable form a caller (skills/myflow-fast/SKILL.md's state
// gate) tests instead of comparing "updatedBy" strings itself -- a rule
// enforced by a field a caller can test beats one enforced only by a
// skill's prose (design.md, kan-174, "read state, then mark").
//
// Any decode failure, or a body with no "updatedBy" field at all, passes
// body through completely unchanged -- byte for byte when it is not a
// synthetic record, which is the common case and is what keeps this
// function from disturbing state get's existing pass-through behaviour.
func markSyntheticIfNeeded(body []byte) []byte {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return body
	}
	updatedByField, ok := raw["updatedBy"]
	if !ok {
		return body
	}
	var updatedBy string
	if err := json.Unmarshal(updatedByField, &updatedBy); err != nil {
		return body
	}
	if updatedBy != stages.SyntheticChangeUpdatedBy {
		return body
	}
	raw["synthetic"] = json.RawMessage("true")
	out, err := json.Marshal(raw)
	if err != nil {
		return body
	}
	return out
}

// runStateGet implements `flow state get <name>`. It prints the store's
// record to stdout on success. On any store failure -- unreachable daemon,
// timeout, malformed response, any non-2xx status other than a legitimate
// 404 -- it reads the on-disk fallback record instead, says so on stderr in
// exactly one line, and still exits 0: a `state get` must never block the
// pipeline any more than a `state set` may (design.md, "The pipeline never
// blocks on this subsystem", is not scoped to writes).
func runStateGet(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fset := flag.NewFlagSet("flow state get", flag.ContinueOnError)
	f, err := parseStateFlags(fset, args, stderr)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintf(stderr, "flow: %v\n", err)
		fmt.Fprint(stderr, stateUsage)
		return 2
	}

	projectKey, _, err := fallback.ProjectKey(f.dir)
	if err != nil {
		fmt.Fprintf(stderr, "flow: resolve project key: %v\n", err)
		return 1
	}

	body, getErr := getChange(ctx, f.addr, f.timeout, projectKey, f.name)
	switch {
	case getErr == nil:
		_, _ = stdout.Write(markSyntheticIfNeeded(body))
		return 0
	case errors.Is(getErr, client.ErrNotFound):
		fmt.Fprintf(stderr, "flow: no state recorded for %s/%s\n", projectKey, f.name)
		return 1
	default:
		// Every other outcome -- ErrUnavailable, or a panic recovered
		// inside getChange -- is a reason to fall back, never a reason to
		// block. Exactly one warning line, then the on-disk record if one
		// exists (silently nothing if it does not: there is nothing more
		// honest to print, and a missing fallback file is not a second
		// error on top of the one just reported).
		fmt.Fprintln(stderr, "⚠ flow: store unreachable — read local fallback")
		statePath := fallback.StateFilePath(projectKey, f.name)
		if diskBody, readErr := fallback.ReadStateFile(statePath); readErr == nil {
			_, _ = stdout.Write(markSyntheticIfNeeded(diskBody))
		}
		return 0
	}
}

// runStateSet implements `flow state set <name>`, reading the whole
// record as JSON from stdin. A successful write to the store exits 0
// silently. A monotonic refusal (the store was reached and correctly said
// no) is reported and exits non-zero. Every other failure mode takes the
// fallback: write the on-disk state file, append the journal entry, print
// exactly one warning line, exit 0.
func runStateSet(ctx context.Context, args []string, stdin io.Reader, _, stderr io.Writer) int {
	fset := flag.NewFlagSet("flow state set", flag.ContinueOnError)
	f, err := parseStateFlags(fset, args, stderr)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintf(stderr, "flow: %v\n", err)
		fmt.Fprint(stderr, stateUsage)
		return 2
	}

	body, err := io.ReadAll(io.LimitReader(stdin, maxStdinBytes+1))
	if err != nil {
		fmt.Fprintf(stderr, "flow: read state from stdin: %v\n", err)
		return 1
	}
	if len(body) > maxStdinBytes {
		fmt.Fprintf(stderr, "flow: state on stdin exceeds %d bytes\n", maxStdinBytes)
		return 2
	}
	if !isJSONObject(body) {
		fmt.Fprintln(stderr, "flow: state on stdin must be a JSON object")
		return 2
	}

	// Stamp before anything reads body: the store request, the on-disk
	// fallback file and the journal entry are all derived from this
	// variable further down, so one clock read reaches all three.
	body, err = stampUpdatedAt(body)
	if err != nil {
		fmt.Fprintf(stderr, "flow: stamp updatedAt: %v\n", err)
		return 1
	}

	// Before the project key is resolved, and therefore before either
	// fallback destination exists to be written to -- see
	// validateWorktreeMergeBases for why the refusal must write nothing.
	if err := validateWorktreeMergeBases(body); err != nil {
		fmt.Fprintf(stderr, "flow: %v\n", err)
		return 2
	}

	projectKey, mainCheckout, err := fallback.ProjectKey(f.dir)
	if err != nil {
		fmt.Fprintf(stderr, "flow: resolve project key: %v\n", err)
		return 1
	}

	putErr := putChange(ctx, f.addr, f.timeout, projectKey, f.name, mainCheckout, body)
	switch {
	case putErr == nil:
		return 0
	case errors.Is(putErr, client.ErrRefused):
		fmt.Fprintf(stderr, "flow: state set refused: %v\n", putErr)
		return 1
	default:
		// ErrUnavailable, or a panic recovered inside putChange: the store
		// could not be reached to answer at all, so the write is not lost
		// -- it is recorded locally and journaled for task 6's reconciler,
		// and the pipeline is never made to wait on it.
		statePath := fallback.StateFilePath(projectKey, f.name)
		journalPath := fallback.JournalFilePath(projectKey, f.name)
		_ = fallback.WriteStateFile(statePath, body)
		_ = fallback.AppendJournalEntry(journalPath, projectKey, f.name, body, time.Now())
		fmt.Fprintln(stderr, "⚠ flow: store unreachable — wrote local journal")
		return 0
	}
}

// stateListFlags is parseStateFlags' sibling for `state list`, which takes
// no positional argument (there is no single change to name) -- kept as
// its own small parser rather than stretching parseStateFlags' "exactly
// one argument" rule to accept zero, which would silently loosen `state
// get`/`state set`'s own argument checking along with it.
type stateListFlags struct {
	addr    string
	timeout time.Duration
	dir     string
}

func parseStateListFlags(fset *flag.FlagSet, args []string, stderr io.Writer) (stateListFlags, error) {
	fset.SetOutput(stderr)
	f := stateListFlags{}
	fset.StringVar(&f.addr, "addr", resolveDefaultAddr(), "flowd base URL")
	fset.DurationVar(&f.timeout, "timeout", defaultTimeout, "store request timeout before falling back")
	fset.StringVar(&f.dir, "C", "", "resolve the project key as if run from this directory (default: cwd)")
	if err := fset.Parse(args); err != nil {
		return stateListFlags{}, err
	}
	noteAddrEnvUsage(fset, stderr)
	if fset.NArg() != 0 {
		return stateListFlags{}, fmt.Errorf("state list takes no positional arguments")
	}
	if f.dir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return stateListFlags{}, fmt.Errorf("resolve working directory: %w", err)
		}
		f.dir = wd
	}
	return f, nil
}

// stateListRecord is one row of `state list`'s output -- the fields common
// to both of its sources: a store answer (client.StateBoardRow, unpacked
// field for field) and a parsed fallback file. Unreadable is set only for
// a fallback-directory entry whose file could not be read or did not
// parse as JSON -- the fallback path's own version of the CLI-wide rule
// that an unreadable record is named, not silently dropped (see
// runFallbackStateList below).
type stateListRecord struct {
	Name       string `json:"name"`
	State      string `json:"state,omitempty"`
	UpdatedAt  string `json:"updatedAt,omitempty"`
	UpdatedBy  string `json:"updatedBy,omitempty"`
	Unreadable bool   `json:"unreadable,omitempty"`
}

// stateListOutput is the one JSON object `state list` prints to stdout.
// Source and Complete are what let a caller (skills/flow-status/SKILL.md,
// skills/flow-contracts/pipeline.md's Change name resolution) tell a
// live enumeration apart from a degraded one without parsing stderr:
// Complete is true only for Source == "store" -- the fallback directory
// can never be presented as a full list, because it holds only the
// records a failed write left behind, never every change the project has
// ever had.
type stateListOutput struct {
	Source   string            `json:"source"`
	Complete bool              `json:"complete"`
	Records  []stateListRecord `json:"records"`
}

// runStateList implements `flow state list`. On success it prints every
// change the store holds for the resolved project, source "store",
// complete true. On any store failure -- unreachable daemon, timeout, a
// non-2xx response, a response missing the daemon header -- it prints one
// warning line, then whatever the local on-disk fallback directory holds,
// source "fallback", complete false, and still exits 0: an enumeration
// must never block the pipeline any more than a single record's read or
// write may.
func runStateList(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fset := flag.NewFlagSet("flow state list", flag.ContinueOnError)
	f, err := parseStateListFlags(fset, args, stderr)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintf(stderr, "flow: %v\n", err)
		fmt.Fprint(stderr, stateUsage)
		return 2
	}

	projectKey, _, err := fallback.ProjectKey(f.dir)
	if err != nil {
		fmt.Fprintf(stderr, "flow: resolve project key: %v\n", err)
		return 1
	}

	rows, listErr := listStateBoard(ctx, f.addr, f.timeout, projectKey)
	if listErr == nil {
		return writeStateListOutput(stdout, stderr, stateListOutput{
			Source:   "store",
			Complete: true,
			Records:  toStateListRecords(rows),
		})
	}

	// Every outcome other than a clean store answer -- ErrUnavailable, or a
	// panic recovered inside listStateBoard -- takes the fallback: the
	// local directory, reported as exactly what it is, never dressed up as
	// a complete list.
	fmt.Fprintln(stderr, "⚠ flow: store unreachable — listing local fallback files (partial: only failed writes are recorded there)")
	records, fbErr := fallbackStateListRecords(projectKey)
	if fbErr != nil {
		fmt.Fprintf(stderr, "flow: state list: read local fallback directory: %v\n", fbErr)
	}
	return writeStateListOutput(stdout, stderr, stateListOutput{
		Source:   "fallback",
		Complete: false,
		Records:  records,
	})
}

// writeStateListOutput encodes out as one line of JSON to stdout. A
// marshal failure here would mean stateListOutput itself cannot be
// encoded -- not a store or filesystem failure -- so it is reported and
// this exits 1 rather than silently printing nothing and claiming 0.
func writeStateListOutput(stdout, stderr io.Writer, out stateListOutput) int {
	encoded, err := json.Marshal(out)
	if err != nil {
		fmt.Fprintf(stderr, "flow: state list: encode output: %v\n", err)
		return 1
	}
	_, _ = stdout.Write(encoded)
	fmt.Fprintln(stdout)
	return 0
}

// stateResolveOutput is the one JSON object `state resolve` prints to
// stdout. Source and Complete mean exactly what state list's own pair
// mean; Candidates is the change-name candidate set a caller resolves a
// bare name against, and Unreadable names every fallback record that
// could not be read or parsed -- both always present as arrays, never
// null, so a caller can range over them without a nil check.
type stateResolveOutput struct {
	Source     string            `json:"source"`
	Complete   bool              `json:"complete"`
	Candidates []stateListRecord `json:"candidates"`
	Unreadable []string          `json:"unreadable"`
}

// runStateResolve implements `flow state resolve`. It shares its
// connection flags and store/fallback boundary with `state list`
// (parseStateListFlags, listStateBoard, fallbackStateListRecords) and
// differs only in what it does with the rows once it has them:
// resolveCandidates drops FINISHED rows on the store path, and unions the
// fallback records with spectre/changes/ directory names (minus the
// archived ones) on the fallback path. Never blocks: a store failure
// falls back exactly as state list's does, and a changes-directory read
// error is reported on stderr while the fallback candidates still print,
// exit 0.
func runStateResolve(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fset := flag.NewFlagSet("flow state resolve", flag.ContinueOnError)
	f, err := parseStateListFlags(fset, args, stderr)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintf(stderr, "flow: %v\n", err)
		fmt.Fprint(stderr, stateUsage)
		return 2
	}

	projectKey, mainCheckout, err := fallback.ProjectKey(f.dir)
	if err != nil {
		fmt.Fprintf(stderr, "flow: resolve project key: %v\n", err)
		return 1
	}

	rows, listErr := listStateBoard(ctx, f.addr, f.timeout, projectKey)
	if listErr == nil {
		cands, unreadable, _ := resolveCandidates(toStateListRecords(rows), "", true)
		return writeStateResolveOutput(stdout, stderr, stateResolveOutput{
			Source:     "store",
			Complete:   true,
			Candidates: cands,
			Unreadable: unreadable,
		})
	}

	// Every outcome other than a clean store answer takes the fallback,
	// exactly as runStateList's own fallback branch does.
	fmt.Fprintln(stderr, "⚠ flow: store unreachable — listing local fallback files (partial: only failed writes are recorded there)")
	records, fbErr := fallbackStateListRecords(projectKey)
	if fbErr != nil {
		fmt.Fprintf(stderr, "flow: state resolve: read local fallback directory: %v\n", fbErr)
	}
	changesDir := filepath.Join(mainCheckout, "spectre", "changes")
	cands, unreadable, dirErr := resolveCandidates(records, changesDir, false)
	if dirErr != nil {
		fmt.Fprintf(stderr, "flow: state resolve: read %s: %v\n", changesDir, dirErr)
	}
	return writeStateResolveOutput(stdout, stderr, stateResolveOutput{
		Source:     "fallback",
		Complete:   false,
		Candidates: cands,
		Unreadable: unreadable,
	})
}

// resolveCandidates turns rows (either the store's board rows or the
// fallback directory's parsed records) into the change-name candidate
// set. fromStore selects which rule applies: on the store path a
// FINISHED row is dropped and changesDir is never consulted; on the
// fallback path every directory directly under changesDir is unioned in
// (skipping "archive" itself and any name already seen), then any name
// archived under changesDir/archive/<name>/ is removed regardless of
// which side it came from -- a fallback record for an archived change is
// exactly as stale as a leftover directory for one.
func resolveCandidates(rows []stateListRecord, changesDir string, fromStore bool) (cands []stateListRecord, unreadable []string, err error) {
	seen := map[string]bool{}
	for _, r := range rows {
		switch {
		case r.Unreadable:
			unreadable = append(unreadable, r.Name)
		case fromStore && r.State == "FINISHED":
		case !seen[r.Name]:
			seen[r.Name] = true
			cands = append(cands, r)
		}
	}
	if fromStore {
		return cands, unreadable, nil
	}

	entries, err := os.ReadDir(changesDir)
	for _, e := range entries {
		if e.IsDir() && e.Name() != "archive" && !seen[e.Name()] {
			seen[e.Name()] = true
			cands = append(cands, stateListRecord{Name: e.Name()})
		}
	}
	cands = slices.DeleteFunc(cands, func(r stateListRecord) bool {
		_, statErr := os.Stat(filepath.Join(changesDir, "archive", r.Name))
		return statErr == nil
	})
	return cands, unreadable, err
}

// writeStateResolveOutput encodes out as one line of JSON to stdout,
// mirroring writeStateListOutput -- with the one addition that
// Candidates/Unreadable are normalised to empty slices first, so the
// printed JSON always carries "candidates":[] and "unreadable":[] rather
// than "null" when either is empty.
func writeStateResolveOutput(stdout, stderr io.Writer, out stateResolveOutput) int {
	if out.Candidates == nil {
		out.Candidates = []stateListRecord{}
	}
	if out.Unreadable == nil {
		out.Unreadable = []string{}
	}
	encoded, err := json.Marshal(out)
	if err != nil {
		fmt.Fprintf(stderr, "flow: state resolve: encode output: %v\n", err)
		return 1
	}
	_, _ = stdout.Write(encoded)
	fmt.Fprintln(stdout)
	return 0
}

// toStateListRecords converts the store's own row shape into this
// command's output shape -- a one-to-one field copy, kept as its own
// small function so runStateList's success branch reads as one line.
func toStateListRecords(rows []client.StateBoardRow) []stateListRecord {
	records := make([]stateListRecord, len(rows))
	for i, r := range rows {
		records[i] = stateListRecord{Name: r.Name, State: r.State, UpdatedAt: r.UpdatedAt, UpdatedBy: r.UpdatedBy}
	}
	return records
}

// fallbackStateListRecords scans the local on-disk fallback directory for
// projectKey (fallback.ListStateFileNames) and returns whatever it can
// parse from each file. A file that cannot be read or does not parse as a
// JSON object with the expected fields is still reported -- name and
// Unreadable: true -- never silently dropped from the list, per this
// command's own never-rebuild-by-inference rule (the same one `state get`
// and skills/flow-status/SKILL.md already follow for a single record).
func fallbackStateListRecords(projectKey string) ([]stateListRecord, error) {
	names, err := fallback.ListStateFileNames(projectKey)
	if err != nil {
		return nil, err
	}
	records := make([]stateListRecord, 0, len(names))
	for _, name := range names {
		body, readErr := fallback.ReadStateFile(fallback.StateFilePath(projectKey, name))
		if readErr != nil {
			records = append(records, stateListRecord{Name: name, Unreadable: true})
			continue
		}
		var parsed struct {
			State     string `json:"state"`
			UpdatedAt string `json:"updatedAt"`
			UpdatedBy string `json:"updatedBy"`
		}
		if jsonErr := json.Unmarshal(body, &parsed); jsonErr != nil {
			records = append(records, stateListRecord{Name: name, Unreadable: true})
			continue
		}
		records = append(records, stateListRecord{
			Name: name, State: parsed.State, UpdatedAt: parsed.UpdatedAt, UpdatedBy: parsed.UpdatedBy,
		})
	}
	return records, nil
}

// listStateBoard calls the store's GET /api/v1/stats/state-board endpoint
// under addr/timeout, recovering from any panic in the client path and
// reporting it as client.ErrUnavailable -- the same guarantee
// getChange/putChange's own recover provides.
func listStateBoard(ctx context.Context, addr string, timeout time.Duration, projectKey string) (rows []client.StateBoardRow, err error) {
	defer func() {
		if r := recover(); r != nil {
			rows, err = nil, fmt.Errorf("%w: recovered panic: %v", client.ErrUnavailable, r)
		}
	}()

	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cl := client.New(addr, &http.Client{Timeout: timeout})
	return cl.ListStateBoard(reqCtx, projectKey)
}

// isJSONObject reports whether body is valid JSON whose top-level value is
// a JSON object -- true for `{}` and `{"state":"IN_PROGRESS"}`, false for
// `null`, `[]`, `"x"`, `42`, and for anything that is not valid JSON at
// all. `state set` requires this, not merely json.Valid: a bare JSON
// `null` satisfies json.Valid, decodes successfully into a nil
// map[string]json.RawMessage (encoding/json's documented behaviour for
// null into a map -- not a decode error), and previously reached
// withMainCheckoutPath, which then panicked assigning into that nil map.
// The panic was recovered, so the never-block guarantee still held and the
// process still exited 0 -- but the literal bytes `null` were then written
// to both the state file and the journal as if they were a real record,
// which a later `state get` or task 6's replay would hand back to the
// daemon as an all-zero write. Rejecting a non-object body here, before
// either the store call or the fallback write, is what stops that record
// from ever being treated as storable in the first place.
func isJSONObject(body []byte) bool {
	var v any
	if err := json.Unmarshal(body, &v); err != nil {
		return false
	}
	_, ok := v.(map[string]any)
	return ok
}

// getChange calls the store's GET endpoint under addr/timeout, recovering
// from any panic in the client path and reporting it as
// client.ErrUnavailable -- a panic is exactly the kind of failure the
// never-block guarantee has to survive too, not just a clean error return.
func getChange(ctx context.Context, addr string, timeout time.Duration, projectKey, name string) (body []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			body, err = nil, fmt.Errorf("%w: recovered panic: %v", client.ErrUnavailable, r)
		}
	}()

	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cl := client.New(addr, &http.Client{Timeout: timeout})
	return cl.GetChange(reqCtx, projectKey, name)
}

// putChange calls the store's PUT endpoint under addr/timeout, injecting
// mainCheckoutPath so the store can bootstrap the project row on its first
// write. Its recover covers two distinct things, both correctly folded
// into the same ErrUnavailable fallback signal: a panic in the client
// path proper (the same guarantee getChange's recover provides), and a
// panic inside withMainCheckoutPath's own body encoding, e.g. a JSON value
// that decodes without error but is not a map -- runStateSet's
// isJSONObject check now rejects that case before putChange is ever
// called, so this recover is a second, cheaper line of defence for that
// specific bug rather than the only one.
func putChange(ctx context.Context, addr string, timeout time.Duration, projectKey, name, mainCheckout string, body []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%w: recovered panic: %v", client.ErrUnavailable, r)
		}
	}()

	reqBody, err := withMainCheckoutPath(body, mainCheckout)
	if err != nil {
		return fmt.Errorf("%w: %v", client.ErrUnavailable, err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cl := client.New(addr, &http.Client{Timeout: timeout})
	return cl.PutChange(reqCtx, projectKey, name, reqBody)
}

// stampUpdatedAt returns body with its "updatedAt" field set to this
// process's own clock, at full precision, overwriting whatever value the
// body carried. Every other field is left untouched.
//
// The field is CLI-owned (skills/flow-contracts/state-file.md): a
// caller supplies it or not, and either way this value is the one that
// reaches the store, the on-disk fallback file and the journal entry.
// That single ownership is the whole fix for KAN-284. The store orders a
// same-state write by this instant, and it was being fed two clocks at
// two precisions -- a skill's `date -u +%Y-%m-%dT%H:%M:%SZ`, truncated to
// the second, against a stage mark's synthetic bootstrap at nanosecond
// precision. A truncated instant compares as earlier than any sub-second
// instant inside the same second, so a write that followed a bootstrap in
// the same second was refused as moving state backwards -- which blocks a
// fix run, whose writes never move the state by design. With one writer
// at one precision the store's strict `>` is satisfied by every live
// write, so the monotonic rule needs no change.
//
// body is expected to already be a JSON object -- runStateSet's
// isJSONObject check enforces that before this is ever called. The nil
// check below is the same second, cheap guard withMainCheckoutPath
// carries, and matters slightly more here: this call site sits outside
// putChange's recover, so a nil-map assignment would take the process
// down rather than degrade to the fallback.
func stampUpdatedAt(body []byte) ([]byte, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decode state as a JSON object: %w", err)
	}
	if raw == nil {
		raw = map[string]json.RawMessage{}
	}
	stamped, err := json.Marshal(formatUpdatedAt(time.Now()))
	if err != nil {
		return nil, fmt.Errorf("encode updatedAt: %w", err)
	}
	raw["updatedAt"] = stamped
	out, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("encode stamped state: %w", err)
	}
	return out, nil
}

// formatUpdatedAt renders at as the "updatedAt" field's wire value: UTC,
// at full precision. It is stampUpdatedAt's format step, named separately
// so the format can be asserted against a fixed instant instead of
// against whatever the clock happens to read
// (TestStateSetStampIsFinerThanSecondPrecision).
//
// time.RFC3339Nano is the format toDTO already emits the field in, so the
// value a `state get` prints and the value a `state set` sends are the
// same shape. Narrowing it to time.RFC3339 would truncate to the second
// and reintroduce the collision described on stampUpdatedAt.
func formatUpdatedAt(at time.Time) string {
	return at.UTC().Format(time.RFC3339Nano)
}

// mergeBasePattern is the only shape a recorded merge base may take: a
// 40-character lowercase hexadecimal sha, exactly as git prints one.
// Uppercase is excluded deliberately rather than folded -- every producer
// in the pipeline is `git merge-base`, which emits lowercase, so an
// uppercase value is a hand-edit and worth reporting as one.
var mergeBasePattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

// validateWorktreeMergeBases reports the first `worktrees` value that is
// neither JSON null nor a sha, naming the worktree path it was recorded
// against and the value itself. `null` is legal and means *no merge base
// recorded* (skills/flow-contracts/state-file.md), which is what a
// worktree registered before its base is known carries.
//
// Its caller in runStateSet exits 2 rather than taking the never-block
// fallback path, and calls it before fallback.ProjectKey: both fallback
// destinations -- the on-disk state file and the journal -- are derived
// from the project key, so returning before it is resolved is what makes
// "writes nothing" a property of the control flow rather than a promise.
// The fallback exists for a store outage; a malformed merge base is a
// caller mistake, and journalling it would hide the bad value until
// check-finish-preflight.sh refuses at the finish gate -- KAN-265's
// failure, which this check exists to stop at the point the value is
// written.
//
// The paths are sorted so that a body carrying several bad values names
// the same one on every run. Decoding into *string is what separates a
// JSON null (nil) from the empty string (non-nil, empty), which are
// different answers: the first is legal, the second is not.
func validateWorktreeMergeBases(body []byte) error {
	var raw struct {
		Worktrees map[string]*string `json:"worktrees"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return fmt.Errorf("decode worktrees: %w", err)
	}
	paths := make([]string, 0, len(raw.Worktrees))
	for path := range raw.Worktrees {
		paths = append(paths, path)
	}
	slices.Sort(paths)
	for _, path := range paths {
		mergeBase := raw.Worktrees[path]
		if mergeBase == nil {
			continue
		}
		if !mergeBasePattern.MatchString(*mergeBase) {
			return fmt.Errorf("worktree %s: merge base %q is neither null nor a 40-character lowercase hex sha", path, *mergeBase)
		}
	}
	return nil
}

// withMainCheckoutPath returns body with a "mainCheckoutPath" field added,
// leaving every field body already carries untouched. This field is
// transport-only: it is never part of the on-disk state file's shape (see
// skills/flow-contracts/state-file.md), which is exactly why it is
// injected into the wire request here rather than being something the
// caller of `state set` has to know to include.
//
// body is expected to already be a JSON object -- runStateSet's
// isJSONObject check enforces that before this is ever called. The nil
// check below is a second, cheap guard rather than the enforcement point:
// encoding/json decodes a JSON `null` into a nil map with no error (this
// is documented encoding/json behaviour, not a bug in the decode), and
// without the guard, assigning into that nil map on the next line would
// panic -- which putChange's recover would still catch, but silently
// writing the literal bytes `null` to disk as if it were a real record is
// the actual defect isJSONObject exists to prevent at the boundary.
func withMainCheckoutPath(body []byte, mainCheckout string) ([]byte, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decode state as a JSON object: %w", err)
	}
	if raw == nil {
		raw = map[string]json.RawMessage{}
	}
	encodedPath, err := json.Marshal(mainCheckout)
	if err != nil {
		return nil, fmt.Errorf("encode main checkout path: %w", err)
	}
	raw["mainCheckoutPath"] = encodedPath
	out, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("encode request body: %w", err)
	}
	return out, nil
}
