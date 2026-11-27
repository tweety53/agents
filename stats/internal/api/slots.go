package api

import (
	"errors"
	"net/http"
	"slices"
	"strings"
	"sync"
	"syscall"
	"time"
)

// slotHandler is the heavy-run semaphore: at most max holders at once, across
// every worktree on the machine -- flow-guard's run-verify-list `test` and
// `visual` lists and recapture (internal/guard/heavyslot.go) each hold one
// for the life of the run. It lives in memory: a slot outlives no run, so a
// flowd restart losing the table loses nothing a holder's next renewal does
// not put back.
//
//	POST /api/v1/slots/acquire {"id","owner","pid","ttl_seconds"} -> 200 {"granted": bool, "max", "holders"}
//	POST /api/v1/slots/release {"id"}                             -> 204
//	GET  /api/v1/slots                                            -> 200 {"max", "holders"}
//
// An acquire by an id already holding renews it: its expiry moves to now plus
// ttl_seconds. A holder is dropped before every call once its ttl passed
// unrenewed or its pid is no longer alive -- flowd runs on the same machine as
// its callers -- so a crashed holder never starves the next one.
type slotHandler struct {
	mu      sync.Mutex
	max     int
	holders []slotHolder
	now     func() time.Time
	alive   func(pid int) bool
}

type slotHolder struct {
	ID      string    `json:"id"`
	Owner   string    `json:"owner"`
	PID     int       `json:"pid"`
	Since   time.Time `json:"since"`
	Expires time.Time `json:"expires"`
}

func newSlotHandler(max int) *slotHandler {
	return &slotHandler{max: max, now: time.Now, alive: pidAlive}
}

// pidAlive reports whether pid names a live process: signal 0 is delivered
// to nothing, and EPERM still means the process exists.
func pidAlive(pid int) bool {
	err := syscall.Kill(pid, 0)
	return err == nil || errors.Is(err, syscall.EPERM)
}

func (h *slotHandler) routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/slots/acquire", h.acquire)
	mux.HandleFunc("POST /api/v1/slots/release", h.release)
	mux.HandleFunc("GET /api/v1/slots", h.list)
}

// prune drops every expired or dead holder; h.mu is held.
func (h *slotHandler) prune() {
	now := h.now()
	h.holders = slices.DeleteFunc(h.holders, func(s slotHolder) bool {
		return now.After(s.Expires) || (s.PID > 0 && !h.alive(s.PID))
	})
}

func (h *slotHandler) state() map[string]any {
	return map[string]any{"max": h.max, "holders": slices.Clone(h.holders)}
}

func (h *slotHandler) acquire(w http.ResponseWriter, r *http.Request) {
	var in struct {
		ID         string `json:"id"`
		Owner      string `json:"owner"`
		PID        int    `json:"pid"`
		TTLSeconds int    `json:"ttl_seconds"`
	}
	if err := decodeJSONBody(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(in.ID) == "" || in.TTLSeconds <= 0 {
		writeError(w, http.StatusBadRequest, "id and a positive ttl_seconds are required")
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.prune()
	now := h.now()
	expires := now.Add(time.Duration(in.TTLSeconds) * time.Second)
	granted := true
	if i := slices.IndexFunc(h.holders, func(s slotHolder) bool { return s.ID == in.ID }); i >= 0 {
		h.holders[i].Expires = expires
	} else if len(h.holders) < h.max {
		h.holders = append(h.holders, slotHolder{ID: in.ID, Owner: in.Owner, PID: in.PID, Since: now, Expires: expires})
	} else {
		granted = false
	}
	out := h.state()
	out["granted"] = granted
	writeJSON(w, http.StatusOK, out)
}

func (h *slotHandler) release(w http.ResponseWriter, r *http.Request) {
	var in struct {
		ID string `json:"id"`
	}
	if err := decodeJSONBody(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.mu.Lock()
	h.holders = slices.DeleteFunc(h.holders, func(s slotHolder) bool { return s.ID == in.ID })
	h.mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

func (h *slotHandler) list(w http.ResponseWriter, _ *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.prune()
	writeJSON(w, http.StatusOK, h.state())
}
