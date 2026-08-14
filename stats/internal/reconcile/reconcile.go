// Package reconcile replays the CLI's write-ahead journal
// (internal/fallback) into the store (internal/store): at daemon startup,
// whenever the daemon regains a database connection, and on demand via
// `myflow journal flush`. It consumes task 5's journal format and task 2's
// change repository directly -- it does not go through internal/api's HTTP
// surface -- but decodes a journal entry's Body with the exact same
// api.DecodeChangeBody the PUT handler uses, so a body the live API would
// accept is exactly a body replay accepts. Task 8's stage-mark journal
// (a sibling "*.journal.stage" file per change, deliberately not mixed
// into the same "*.journal" stream -- see cmd/myflow/stage.go's
// stageJournalPath doc comment) is replayed the same way, through
// api.ApplyBeginStageMark/ApplyEndStageMark rather than a second
// implementation of what a mark does.
//
// Nothing here reconstructs the monotonic-write rule: internal/store's
// PutChange already refuses a write that would move a record backwards, in
// either the pipeline-state dimension or, at the same state, the
// UpdatedAt dimension (store.PutChange's own doc comment). Replay's job is
// narrower -- feed every pending entry to that guard, in file order, and
// retire each one once the store has reached a *definitive* outcome for it:
// accepted, or refused for a reason that will not change on retry
// (api.IsDefinitiveChangeOutcome), which folds together a genuinely
// superseded write and a benign duplicate retry (see
// internal/api/changes.go's put doc comment for why reconciliation must not
// need to tell those two apart) alongside a refusal whose cause lives in
// the entry's own content -- an invalid state, an unresolvable project
// bootstrap -- and a body that fails to decode at all. A stage-mark entry's
// equivalent guard is api.IsDefinitiveMarkOutcome -- see its own doc
// comment for what counts as definitive for a mark.
package reconcile

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/tweety53/agents/stats/internal/api"
	"github.com/tweety53/agents/stats/internal/client"
	"github.com/tweety53/agents/stats/internal/fallback"
	"github.com/tweety53/agents/stats/internal/store"
)

// ChangeStore is the store dependency Reconciler needs for the state
// journal: exactly the one method it calls, defined here at the consumer
// per go-interface-design's "define interfaces at the consumer" -- a fake
// satisfying it needs no database.
type ChangeStore interface {
	PutChange(ctx context.Context, c store.Change) error
}

// var _ ChangeStore = (*store.Store)(nil) verifies at compile time that the
// real store satisfies the interface this package actually depends on.
var _ ChangeStore = (*store.Store)(nil)

// stageJournalSuffix is appended to a change's ordinary journal path by
// cmd/myflow/stage.go's stageJournalPath -- this package's own copy of
// that literal, for the same reason internal/client keeps its own literal
// copy of the daemon header rather than importing across the CLI/daemon
// boundary: cmd/myflow is a main package and cannot be imported, so the
// two sides are allowed to agree on a string constant without sharing a
// package. stageJournalSuffix_test.go (or the reconcile package's own
// tests) pins that the two literals agree.
const stageJournalSuffix = ".stage"

// Result totals one Run's outcome across every journal file it visited --
// both the state journal ("*.journal") and the stage-mark journal
// ("*.journal.stage").
type Result struct {
	// Journals is how many journal files (of either kind) Run found under
	// its root, whether or not any of them had pending entries.
	Journals int
	// Applied is how many entries the store accepted -- a state write, or
	// a stage mark ApplyBeginStageMark/ApplyEndStageMark recorded
	// successfully.
	Applied int
	// Refused is how many entries reached a definitive-but-not-accepted
	// outcome (api.IsDefinitiveChangeOutcome for a state entry,
	// api.IsDefinitiveMarkOutcome for a stage mark) -- a stale or
	// duplicate state entry, an invalid state or unresolvable project
	// bootstrap, a body that fails to decode at all, an undocumented stage
	// name, or an end mark with no open run left to close -- retired
	// exactly as an accepted one is (see the package doc comment).
	Refused int
}

// Reconciler replays every journal file under Root into Store.
type Reconciler struct {
	store      ChangeStore
	stageStore api.StageStore
	root       string
	logger     *slog.Logger

	// mu serializes Run: a startup replay and a reconnect-triggered replay
	// can fire at (or near) the same instant, and retirePrefix's
	// read-current/write-temp/rename sequence is not itself safe against a
	// second, concurrent instance of that same sequence racing it on the
	// same journal file. Run's own doc comment explains what a second,
	// blocked caller observes.
	mu sync.Mutex
}

// New builds a Reconciler that replays journal files found under root
// (typically fallback.StateRoot()) into cs (state) and ss (stage marks).
// logger may be nil, in which case log output specific to reconciliation
// (a partial trailing journal line, an unresolvable stage mark, for
// instance) is simply not emitted -- Run and Watch still work.
func New(cs ChangeStore, ss api.StageStore, root string, logger *slog.Logger) *Reconciler {
	return &Reconciler{store: cs, stageStore: ss, root: root, logger: logger}
}

// Run walks r.root for every "*.journal" file, replaying each one's
// pending entries into the store in file order and retiring the ones it
// resolves. It returns once every journal file has been visited -- Run
// does not loop or retry on its own; a caller that wants "whenever the
// daemon regains a database connection" wires that trigger itself (see
// Watch) and calls Run again.
//
// Concurrent Run calls are serialized by r.mu: only one physically walks
// and rewrites journal files at a time. The second caller's Run blocks
// until the first returns, then itself walks and replays whatever the
// first left behind -- nothing, if the first cleared every journal, or the
// remainder, if the first stopped partway (a transport failure, ctx
// cancellation) or a concurrent CLI append landed after the first's read.
// Nothing is applied twice by this: PutChange's own monotonic guard
// refuses a second application of an entry already accepted, and that
// refusal retires the entry exactly as the first, accepting call did.
//
// A journal directory that does not exist yet -- no CLI invocation has
// ever taken the fallback path -- is not an error: Run reports a zero
// Result.
func (r *Reconciler) Run(ctx context.Context) (Result, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var result Result

	walkErr := filepath.WalkDir(r.root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if path == r.root && errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			return err
		}
		if d.IsDir() {
			return nil
		}

		// A stage-mark journal ("*.journal.stage") is recognised first --
		// it does *not* match the plain "*.journal" suffix below (it ends
		// in ".stage"), which is exactly why cmd/myflow/stage.go's
		// stageJournalPath chose that shape: the two entry kinds are
		// walked by this one loop, but never fed to the same decoder. See
		// the package doc comment.
		switch {
		case strings.HasSuffix(path, ".journal"+stageJournalSuffix):
			result.Journals++
			applied, refused, rerr := r.replayStageFile(ctx, path)
			result.Applied += applied
			result.Refused += refused
			return rerr
		case strings.HasSuffix(path, ".journal"):
			result.Journals++
			applied, refused, rerr := r.replayFile(ctx, path)
			result.Applied += applied
			result.Refused += refused
			return rerr
		default:
			return nil
		}
	})
	if walkErr != nil {
		if errors.Is(walkErr, fs.ErrNotExist) {
			return result, nil
		}
		return result, fmt.Errorf("reconcile: %w", walkErr)
	}
	return result, nil
}

// Watch runs until ctx is done, pinging the store every interval via ping
// and calling Run -- reporting its outcome through onResult, which may be
// nil -- whenever ping transitions from failing to succeeding. That
// transition is design.md's "regains a database connection"
// (`Reconciler.Run` alone only covers the "at startup" half of that
// requirement; a caller wires Watch for the "on reconnect" half).
//
// Watch assumes a connection is already up when it starts: the daemon's
// own startup sequence already replays once via a direct Run call, after
// store.Open's own ping has already succeeded, so Watch's first job is
// detecting the *next* failure-to-success transition, not replaying
// redundantly for the connection it started with.
func (r *Reconciler) Watch(ctx context.Context, ping func(context.Context) error, interval time.Duration, onResult func(Result, error)) {
	healthy := true

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, interval)
			err := ping(pingCtx)
			cancel()

			if err != nil {
				healthy = false
				continue
			}
			if !healthy {
				healthy = true
				result, runErr := r.Run(ctx)
				if onResult != nil {
					onResult(result, runErr)
				}
			}
		}
	}
}

// parsedEntry is one journal line already split out of a *.journal file's
// bytes, together with the byte offset (exclusive) where its line --
// including its trailing newline -- ends. That offset is what replayFile
// accumulates into retirePrefix's consumed argument: retiring is always
// "everything up to and including the last line replayFile resolved",
// never a count of entries, so a blank line (lineEnd set, IsBlank true)
// advances it exactly like a real entry.
type parsedEntry struct {
	fallback.Entry
	lineEnd int
	isBlank bool
}

// splitCompleteLines splits raw into the leading span that ends in a
// complete, newline-terminated line ("complete") and whatever trails the
// last newline ("tail"). Only complete is ever parsed as entries.
//
// This matters because fallback.AppendJournalEntry writes one entry's JSON
// and its trailing newline in a single f.Write call (see its own doc
// comment on O_APPEND's atomicity guarantee) -- so on a healthy write,
// every line in the file is either wholly present, newline included, or
// not there at all. The one way a byte range after the last newline can be
// non-empty is a process that died mid-write, after the kernel had
// accepted some but not all of that single Write call's bytes: a real, if
// narrow, failure mode (a crash or a kill -9 mid-syscall), not merely a
// theoretical one, and worth naming rather than assuming away. Treating
// that trailing span as data to parse would either fail to decode (most
// likely, since it is a JSON fragment) or -- worse -- succeed by
// coincidence and be replayed as if it were a real entry. Leaving it alone
// is the safe response either way: it is never counted as consumed, so
// retirePrefix never removes it, and it stays exactly where it is for a
// human to inspect. It cannot ever complete itself into a valid line on a
// later Run, because AppendJournalEntry never resumes a partial write --
// every call starts a fresh line -- so a genuine partial tail is a
// standing (logged) anomaly, not a transient one this package attempts to
// repair.
func splitCompleteLines(raw []byte) (complete, tail []byte) {
	idx := bytes.LastIndexByte(raw, '\n')
	if idx < 0 {
		return nil, raw
	}
	return raw[:idx+1], raw[idx+1:]
}

// parseCompleteEntries decodes every complete line in complete (as
// splitCompleteLines defines "complete") into a parsedEntry, in file
// order. A blank line (whitespace only) decodes to a zero-value Entry with
// isBlank set, rather than being skipped outright, so its byte span is
// still accounted for in the running offset replayFile uses to retire.
//
// If a complete line fails to decode as JSON, parseCompleteEntries stops
// and returns the entries parsed so far alongside an error: a malformed
// *complete* line (newline-terminated, so not the partial-write case
// splitCompleteLines already isolated) means the journal itself is
// corrupt, not that this specific entry is skippable. Stopping rather than
// skipping past it keeps whatever comes after untouched in the journal
// instead of silently discarding potentially-recoverable content.
func parseCompleteEntries(complete []byte) ([]parsedEntry, error) {
	var out []parsedEntry
	start := 0
	for start < len(complete) {
		idx := bytes.IndexByte(complete[start:], '\n')
		if idx < 0 {
			// complete is defined to end in '\n' by splitCompleteLines, so
			// this is unreachable on any input this function is actually
			// called with. Guarded anyway rather than looping forever.
			break
		}
		end := start + idx + 1
		line := bytes.TrimSpace(complete[start : start+idx])
		if len(line) == 0 {
			out = append(out, parsedEntry{lineEnd: end, isBlank: true})
			start = end
			continue
		}

		var e fallback.Entry
		if err := json.Unmarshal(line, &e); err != nil {
			return out, fmt.Errorf("decode journal line at offset %d: %w", start, err)
		}
		out = append(out, parsedEntry{Entry: e, lineEnd: end})
		start = end
	}
	return out, nil
}

// errChangeEntryDecodeFailed wraps a DecodeChangeBody failure so that
// replayJournalFile's generic classifier can recognise it and treat it as
// definitive -- unconditionally, never delegated to
// IsDefinitiveChangeOutcome -- exactly as replayFile always has: a body
// that fails to decode (an undocumented field under the closed-schema
// DisallowUnknownFields check included) will decode identically on every
// future replay, so retiring it here is what stops one bad entry from
// permanently blocking every valid entry queued behind it.
var errChangeEntryDecodeFailed = errors.New("reconcile: journal entry body does not decode as a change")

// replayFile replays the pending entries of the single journal file at
// path, applying each one to r.store in file order and retiring
// (removing) every entry whose outcome is *definitive*: an accepted write,
// a body that fails to decode (always definitive, see
// errChangeEntryDecodeFailed), or a PutChange refusal
// IsDefinitiveChangeOutcome reports as definitive. It returns the counts
// of each outcome for this file alone.
//
// Before this used IsDefinitiveChangeOutcome, a PutChange refusal for any
// reason other than errors.Is(putErr, store.ErrMonotonicViolation) (for
// example store.ErrInvalidState or store.ErrInvalidMainCheckoutPath) fell
// into the "unknown, stop here" branch and was left in the journal.
// Because that branch also halts the whole file, every entry queued
// behind the bad one was permanently blocked too, on every future
// replay -- the bad entry never becomes valid by being retried, so it
// stayed first in line forever. Both cases are now classified as
// definitive and retired, exactly as a stage mark's
// undocumented-stage-name refusal already is (api.IsDefinitiveMarkOutcome)
// -- a refusal whose cause lives in the entry itself, not in the store's
// availability, is precisely what "explicitly refused" already covers in
// the sentence above.
//
// The stop-vs-retire skeleton this shares with replayStageFile is factored
// out as replayJournalFile; only the decode-and-apply step and the
// definitive-outcome classifier differ between the two. See that
// function's doc comment.
func (r *Reconciler) replayFile(ctx context.Context, path string) (applied, refused int, err error) {
	apply := func(ctx context.Context, e fallback.Entry) error {
		c, decodeErr := api.DecodeChangeBody(e.Project, e.Name, e.Body)
		if decodeErr != nil {
			return fmt.Errorf("%w: %v", errChangeEntryDecodeFailed, decodeErr)
		}
		return r.store.PutChange(ctx, c)
	}
	isDefinitive := func(err error) bool {
		if errors.Is(err, errChangeEntryDecodeFailed) {
			return true
		}
		return api.IsDefinitiveChangeOutcome(err)
	}
	return r.replayJournalFile(ctx, path, "journal", apply, isDefinitive)
}

// stageMarkJournalBody mirrors cmd/myflow/stage.go's own
// stageMarkJournalBody exactly -- {"kind":"begin"|"end","request":<the
// client.BeginStageRequest or client.EndStageRequest that was
// journalled>}. This package cannot import cmd/myflow (a main package,
// which Go never allows importing), so this is a second declaration of the
// same wire shape rather than a shared type: this struct's two field names
// and cmd/myflow/stage.go's stageJournalPath's ".journal"+".stage" suffix
// are the two literals that must stay in step across that boundary.
// internal/reconcile's own tests exercise this decoder against entries
// built the same way stage.go builds them -- client.BeginStageRequest/
// EndStageRequest marshalled into this envelope -- which is the shared
// type both sides really agree on; only the envelope and the suffix are
// duplicated conventions, the same shape internal/client's own
// daemonHeaderName/daemonHeaderValue literals already duplicate against
// internal/api's for an identical reason (client.go's own doc comment).
type stageMarkJournalBody struct {
	Kind    string          `json:"kind"`
	Request json.RawMessage `json:"request"`
}

// replayStageFile replays the pending entries of the single stage-mark
// journal file at path -- a "*.journal.stage" file, cmd/myflow/stage.go's
// stageJournalPath -- against r.stageStore, in file order, retiring
// (removing) every entry whose outcome api.IsDefinitiveMarkOutcome reports
// as definitive.
//
// A body that fails to decode as a stageMarkJournalBody, or whose Request
// fails to decode as the begin/end shape its Kind names, or whose Kind
// names neither "begin" nor "end", does not match any of
// IsDefinitiveMarkOutcome's cases, so replay stops there (rather than
// retiring, and thereby discarding, a line it cannot make sense of) --
// unlike replayFile, where an equivalent decode failure is always
// definitive. See replayJournalFile's doc comment for the skeleton the two
// share and this difference between their classifiers.
func (r *Reconciler) replayStageFile(ctx context.Context, path string) (applied, refused int, err error) {
	return r.replayJournalFile(ctx, path, "stage mark", r.applyStageMarkEntry, api.IsDefinitiveMarkOutcome)
}

// replayJournalFile is the skeleton replayFile and replayStageFile share:
// read the file, split off any partial trailing line (warning but never
// parsing it), decode the complete lines into entries, and walk them in
// order applying each one via apply -- the one step that differs between a
// state journal ("*.journal") and a stage-mark journal
// ("*.journal.stage"). isDefinitive classifies apply's result: nil is
// always applied; a non-nil error isDefinitive reports true for is
// refused and retired; anything else stops the walk, leaving that entry
// and everything behind it in the journal for the next Run -- which is
// what makes an interrupted replay repeat rather than lose work
// (design.md, "an entry is removed from the journal only once the store
// has accepted or explicitly refused it"). kind names the journal kind in
// log messages ("journal" or "stage mark").
//
// The retire policy here used to live twice, hand-kept in sync between
// replayFile and replayStageFile -- and this change has already been
// bitten by exactly that drift once (IsDefinitiveChangeOutcome not
// existing until after IsDefinitiveMarkOutcome did, see replayFile's own
// history in its doc comment). Only the decode-and-apply closure and the
// definitive-error predicate are still allowed to differ; everything else
// -- line splitting, cancellation, retiring, error wrapping -- lives here
// exactly once.
func (r *Reconciler) replayJournalFile(
	ctx context.Context,
	path string,
	kind string,
	apply func(context.Context, fallback.Entry) error,
	isDefinitive func(error) bool,
) (applied, refused int, err error) {
	raw, readErr := os.ReadFile(path)
	if readErr != nil {
		if errors.Is(readErr, fs.ErrNotExist) {
			return 0, 0, nil
		}
		return 0, 0, fmt.Errorf("read %s: %w", path, readErr)
	}

	complete, tail := splitCompleteLines(raw)
	if len(bytes.TrimSpace(tail)) > 0 && r.logger != nil {
		r.logger.Warn("reconcile: ignoring partial trailing "+kind+" journal line",
			"path", path, "bytes", len(bytes.TrimSpace(tail)))
	}

	entries, parseErr := parseCompleteEntries(complete)

	consumed := 0
entryLoop:
	for _, pe := range entries {
		if ctx.Err() != nil {
			break entryLoop
		}

		if pe.isBlank {
			consumed = pe.lineEnd
			continue
		}

		switch applyErr := apply(ctx, pe.Entry); {
		case applyErr == nil:
			applied++
			consumed = pe.lineEnd
		case isDefinitive(applyErr):
			// A refusal that will not change on retry -- see apply's and
			// isDefinitive's own callers (replayFile, replayStageFile) for
			// what that means for each journal kind. Safe to retire
			// exactly as an accepted entry is.
			refused++
			consumed = pe.lineEnd
			if r.logger != nil {
				r.logger.Warn("reconcile: "+kind+" journal entry refused definitively, retiring",
					"path", path, "project", pe.Project, "name", pe.Name, "error", applyErr)
			}
		default:
			// Outcome unknown -- stop here so this entry, and everything
			// after it, stays in the journal for the next Run.
			if r.logger != nil {
				r.logger.Warn("reconcile: "+kind+" journal replay stopped on an unresolved entry",
					"path", path, "project", pe.Project, "name", pe.Name, "error", applyErr)
			}
			break entryLoop
		}
	}

	if consumed > 0 {
		if retireErr := retirePrefix(path, raw, consumed); retireErr != nil {
			return applied, refused, fmt.Errorf("retire %s: %w", path, retireErr)
		}
	}

	if parseErr != nil {
		return applied, refused, fmt.Errorf("%s: %w", path, parseErr)
	}
	if ctxErr := ctx.Err(); ctxErr != nil {
		return applied, refused, ctxErr
	}
	return applied, refused, nil
}

// applyStageMarkEntry decodes e.Body as a stageMarkJournalBody and applies
// it via api.ApplyBeginStageMark or api.ApplyEndStageMark -- the same two
// functions internal/api's own HTTP handlers call for a live mark, so
// replay can never implement "what a begin/end mark does" differently
// from the handler that serves it live.
func (r *Reconciler) applyStageMarkEntry(ctx context.Context, e fallback.Entry) error {
	var body stageMarkJournalBody
	if err := json.Unmarshal(e.Body, &body); err != nil {
		return fmt.Errorf("decode stage journal entry body: %w", err)
	}

	switch body.Kind {
	case "begin":
		var req client.BeginStageRequest
		if err := json.Unmarshal(body.Request, &req); err != nil {
			return fmt.Errorf("decode begin stage mark: %w", err)
		}
		_, err := api.ApplyBeginStageMark(ctx, r.stageStore, r.logger, api.BeginStageMark{
			ProjectKey:       req.ProjectKey,
			MainCheckoutPath: req.MainCheckoutPath,
			ChangeName:       req.ChangeName,
			RepoRoot:         req.RepoRoot,
			Harness:          req.Harness,
			SessionID:        req.SessionID,
			Command:          req.Command,
			Stage:            req.Stage,
			StartedAt:        req.StartedAt,
		})
		return err
	case "end":
		var req client.EndStageRequest
		if err := json.Unmarshal(body.Request, &req); err != nil {
			return fmt.Errorf("decode end stage mark: %w", err)
		}
		_, err := api.ApplyEndStageMark(ctx, r.stageStore, api.EndStageMark{
			ProjectKey: req.ProjectKey,
			ChangeName: req.ChangeName,
			Command:    req.Command,
			Stage:      req.Stage,
			EndedAt:    req.EndedAt,
			Outcome:    req.Outcome,
			Metrics:    req.Metrics,
		})
		return err
	default:
		return fmt.Errorf("unknown stage mark kind %q", body.Kind)
	}
}

// retirePrefix removes the first consumed bytes of raw from the journal at
// path -- the entries replayFile has just resolved -- while preserving
// everything after them, including any entry appended to the file after
// raw was read.
//
// Truncation in place is deliberately not used: os.Truncate can only chop
// bytes off the end of a file at a length the caller names, and any length
// this function could name is derived from raw, a snapshot already stale
// by the time retirePrefix runs. A concurrent append that landed after raw
// was read has already extended the file past what raw shows; truncating
// to a length computed from raw would cut that entry off rather than
// preserve it -- silent data loss, and exactly the bug the plan calls out
// truncation for.
//
// Instead this re-reads the file's current content immediately before
// writing anything, checks that it still begins with the exact bytes just
// consumed (raw[:consumed]) -- a safety check that only ever fires if
// something outside this package's own single-writer-per-Run discipline
// rewrote the journal's already-processed prefix, which Reconciler.Run's
// mutex is supposed to make impossible -- and writes the current content's
// remainder into a temp file in the same directory (guaranteeing the
// rename below is same-filesystem, the precondition for it being atomic),
// then swaps it into place with a single os.Rename. POSIX guarantees
// rename is atomic: a reader (or another replay) observes either the whole
// old file or the whole new one, never a partial write.
//
// This section runs under fallback.LockJournal's sidecar advisory lock
// (acquired below, held for the whole read-current/write-temp/rename
// sequence), which is what actually closes the race a previous version of
// this comment described and left open: fallback.AppendJournalEntry opens
// the journal, then writes, as two separate steps -- if an append's Open
// call resolved the pre-rename file an instant before this function's
// rename replaced it, and that append's Write executed after the rename,
// the appended bytes would land on the now-unlinked old file and be lost
// once every open handle to it closed. That was measured, not
// hypothetical: an unlocked 20,000-entry stress reproducer lost roughly
// 71% of entries this way (F1). AppendJournalEntry now takes the same
// sidecar lock (with a short, bounded wait -- see its own doc comment for
// why it must never block unboundedly) before its own open+write, which
// removes the window entirely rather than merely narrowing it. See
// fallback.LockJournal's doc comment for why the lock is a *sidecar* file
// (path + ".lock") rather than an flock taken on the journal file itself
// -- that choice, not merely "add a lock", is what makes this safe across
// this function's own rename.
//
// This lock also protects something Reconciler.Run's in-process mutex
// cannot: two separate *processes* retiring the same journal at once (two
// myflowd instances, or a daemon racing a concurrent `myflow journal
// flush`). The in-process mutex stays in place too -- it is cheaper for
// the common, single-process case, and does not conflict with also
// holding this lock.
func retirePrefix(path string, raw []byte, consumed int) error {
	unlock, err := fallback.LockJournal(path)
	if err != nil {
		return fmt.Errorf("acquire retire lock: %w", err)
	}
	defer unlock()

	current, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("re-read before retiring: %w", err)
	}
	if !bytes.HasPrefix(current, raw[:consumed]) {
		return fmt.Errorf("journal prefix changed unexpectedly since read")
	}
	retained := current[consumed:]

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp journal: %w", err)
	}
	tmpPath := tmp.Name()
	// Every early return below removes tmpPath: a failed retire must not
	// leave a stray .tmp-* file beside the journal for a later Run (or a
	// human) to trip over.
	if _, err := tmp.Write(retained); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("write temp journal: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("sync temp journal: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("close temp journal: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename temp journal into place: %w", err)
	}
	return nil
}
