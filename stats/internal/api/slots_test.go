package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func slotCall(t *testing.T, h *slotHandler, method, path, body string) (int, map[string]any) {
	t.Helper()
	mux := http.NewServeMux()
	h.routes(mux)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(method, path, strings.NewReader(body)))
	var out map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	return rec.Code, out
}

func granted(t *testing.T, h *slotHandler, id string, pid int) bool {
	t.Helper()
	code, out := slotCall(t, h, "POST", "/api/v1/slots/acquire", `{"id":"`+id+`","owner":"o-`+id+`","pid":`+itoa(pid)+`,"ttl_seconds":60}`)
	if code != http.StatusOK {
		t.Fatalf("acquire %s: status %d %v", id, code, out)
	}
	return out["granted"] == true
}

func itoa(n int) string { b, _ := json.Marshal(n); return string(b) }

// TestSlotsExcludeBeyondTheMax: the max holders get slots, the next waits;
// a renewal by a holder is granted; a release frees the slot at once.
func TestSlotsExcludeBeyondTheMax(t *testing.T) {
	t.Parallel()
	h := newSlotHandler(2)
	h.alive = func(int) bool { return true }
	if _, list := slotCall(t, h, "GET", "/api/v1/slots", ""); list["holders"] == nil {
		t.Fatalf("empty listing = %v, want holders [] never null", list)
	}
	if !granted(t, h, "a", 1) || !granted(t, h, "b", 2) {
		t.Fatal("first two acquires refused")
	}
	if granted(t, h, "c", 3) {
		t.Fatal("third acquire granted over a max of 2")
	}
	if !granted(t, h, "a", 1) {
		t.Fatal("holder's renewal refused")
	}
	if code, _ := slotCall(t, h, "POST", "/api/v1/slots/release", `{"id":"a"}`); code != http.StatusNoContent {
		t.Fatalf("release status %d", code)
	}
	if !granted(t, h, "c", 3) {
		t.Fatal("acquire after a release refused")
	}
	_, list := slotCall(t, h, "GET", "/api/v1/slots", "")
	if list["max"] != 2.0 || len(list["holders"].([]any)) != 2 {
		t.Errorf("listing = %v, want max 2 and 2 holders", list)
	}
}

// TestSlotsExpireACrashedHolder: a holder whose pid is gone, or whose ttl
// passed unrenewed, never keeps a slot from the next caller.
func TestSlotsExpireACrashedHolder(t *testing.T) {
	t.Parallel()
	h := newSlotHandler(1)
	dead := map[int]bool{}
	h.alive = func(pid int) bool { return !dead[pid] }
	if !granted(t, h, "a", 1) {
		t.Fatal("acquire refused")
	}
	dead[1] = true
	if !granted(t, h, "b", 2) {
		t.Fatal("dead holder's slot not freed")
	}
	now := time.Now()
	h.now = func() time.Time { return now.Add(61 * time.Second) }
	if !granted(t, h, "c", 3) {
		t.Fatal("expired holder's slot not freed")
	}
}

func TestSlotsRefuseABadAcquire(t *testing.T) {
	t.Parallel()
	h := newSlotHandler(1)
	for _, body := range []string{`{}`, `{"id":"a","ttl_seconds":0}`, `{"id":"a","ttl_seconds":-1}`} {
		if code, _ := slotCall(t, h, "POST", "/api/v1/slots/acquire", body); code != http.StatusBadRequest {
			t.Errorf("%s: status %d, want 400", body, code)
		}
	}
}
