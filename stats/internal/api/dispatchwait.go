package api

import (
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/tweety53/agents/stats/internal/records"
)

const (
	// dispatchWaitMax caps one long-poll, inside writeTimeout so the answer
	// is always written whole; a client re-issues the wait to go longer.
	dispatchWaitMax = 20 * time.Second
	// dispatchWaitRecheck re-reads the store while a wait is open: a row
	// closed by a process other than this daemon -- the CLI replaying its
	// own journal straight into the store -- sends no wakeup.
	dispatchWaitRecheck = 5 * time.Second
)

// dispatchEnds wakes every open long-poll when a dispatch row closes in
// this process: ApplyDispatchEnd closes ch and replaces it.
// ponytail: one channel for every change; waiters of other changes re-read
// and go back to sleep -- per-change channels if waiters ever number many.
var dispatchEnds = struct {
	mu sync.Mutex
	ch chan struct{}
}{ch: make(chan struct{})}

func dispatchEndWake() <-chan struct{} {
	dispatchEnds.mu.Lock()
	defer dispatchEnds.mu.Unlock()
	return dispatchEnds.ch
}

func notifyDispatchEnd() {
	dispatchEnds.mu.Lock()
	defer dispatchEnds.mu.Unlock()
	close(dispatchEnds.ch)
	dispatchEnds.ch = make(chan struct{})
}

// waitDispatches serves GET
// /api/v1/records/{project}/{change}/dispatches/wait?keys=<k,k…>&mode=any|all&timeout=<s>:
// `flow-guard await-rows`'s long-poll. It answers 200 with the change's
// dispatch rows -- the array `flow record dispatches` prints -- as soon as
// every named key's row is ended (mode=all, the default), any one is
// (mode=any), or a named key has no row; otherwise when timeout seconds
// (default and cap dispatchWaitMax) pass, with the rows as they stand. A
// closing write wakes it at once; it re-reads every dispatchWaitRecheck
// besides. No keys, a mode other than any/all, or a non-integer timeout is
// 400; an unknown change is 404.
func (h *recordHandler) waitDispatches(w http.ResponseWriter, r *http.Request) {
	project, change := r.PathValue("project"), r.PathValue("change")
	q := r.URL.Query()
	var keys []string
	for k := range strings.SplitSeq(q.Get("keys"), ",") {
		if k != "" {
			keys = append(keys, k)
		}
	}
	mode := q.Get("mode")
	if mode == "" {
		mode = "all"
	}
	timeout := dispatchWaitMax
	if t := q.Get("timeout"); t != "" {
		n, err := strconv.Atoi(t)
		if err != nil || n < 0 {
			writeError(w, http.StatusBadRequest, "timeout must be a non-negative integer of seconds")
			return
		}
		timeout = min(time.Duration(n)*time.Second, dispatchWaitMax)
	}
	if len(keys) == 0 || (mode != "any" && mode != "all") {
		writeError(w, http.StatusBadRequest, "keys must name at least one dispatch key and mode must be any or all")
		return
	}
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	recheck := time.NewTicker(dispatchWaitRecheck)
	defer recheck.Stop()
	for {
		wake := dispatchEndWake()
		rec, err := h.store.RunRecord(r.Context(), project, change)
		if err != nil {
			status, msg := mapStoreError(h.logger, fmt.Sprintf("wait on the dispatches of %s/%s", project, change), err)
			writeError(w, status, msg)
			return
		}
		rows := rec.Dispatches
		if rows == nil {
			rows = []records.Dispatch{}
		}
		if dispatchWaitDone(rows, keys, mode == "any") {
			writeJSON(w, http.StatusOK, rows)
			return
		}
		select {
		case <-wake:
		case <-recheck.C:
		case <-deadline.C:
			writeJSON(w, http.StatusOK, rows)
			return
		case <-r.Context().Done():
			return
		}
	}
}

// dispatchWaitDone reports whether a wait on keys is answered: a key with
// no row answers it at once (the caller reports the unknown key), and
// otherwise any or every key's row being ended does, by mode.
func dispatchWaitDone(rows []records.Dispatch, keys []string, anyMode bool) bool {
	ended := 0
	for _, k := range keys {
		i := slices.IndexFunc(rows, func(d records.Dispatch) bool { return d.Key == k })
		if i < 0 {
			return true
		}
		if rows[i].EndedAt != nil {
			ended++
		}
	}
	return ended == len(keys) || (anyMode && ended > 0)
}
