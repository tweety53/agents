package core

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// heavySlotTTL is how long flowd keeps a slot no renewal reached: a holder
// renews every third of it, and one that died is dropped at the latest then
// -- at once, when flowd sees its pid gone (internal/api/slots.go).
const heavySlotTTL = 60 * time.Second

// HeavySlot holds a heavy run -- flow-guard run-verify-list's `test` and
// `visual` lists, flow-guard recapture -- until flowd grants it one of its
// heavy-run slots (FLOWD_HEAVY_SLOTS, 2 unset), so at most that many run at
// once across every worktree on the machine. owner names the run in flowd's
// GET /api/v1/slots listing. While no slot is free, name prints one `waiting`
// line on stderr and asks again every 2 seconds for up to FLOW_SLOT_WAIT
// seconds, 1800 unset; still refused then, HeavySlot returns false and the
// caller runs nothing, its reason `no-heavy-slot`. Granted, the slot is
// renewed in the background until the returned release, which the caller
// defers, hands it back the instant the run exits. A flowd that cannot be
// reached, or answers anything but a grant decision, holds nothing: one line
// on stderr and the run proceeds, never blocked on a missing daemon. An Env
// whose SlotGate is unset holds nothing. The time from the first ask to
// the answer is recorded as the run's wait (RecordWait).
func HeavySlot(env Env, name, owner string, stderr io.Writer) (release func(), ok bool) {
	none := func() {}
	if !env.SlotGate {
		return none, true
	}
	addr := FlowAddr(env)
	wait := 1800 * time.Second
	if s, err := strconv.Atoi(env.Getenv("FLOW_SLOT_WAIT")); err == nil && s >= 0 {
		wait = time.Duration(s) * time.Second
	}
	poll, renew := env.SlotPoll, env.SlotRenew
	if poll == 0 {
		poll = 2 * time.Second
	}
	if renew == 0 {
		renew = heavySlotTTL / 3
	}
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	id := hex.EncodeToString(b)
	hc := &http.Client{Timeout: 5 * time.Second}
	post := func(path string, body any) (*http.Response, error) {
		j, _ := json.Marshal(body)
		return hc.Post(addr+path, "application/json", strings.NewReader(string(j)))
	}
	acquire := func() (granted bool, err error) {
		resp, err := post("/api/v1/slots/acquire", map[string]any{"id": id, "owner": owner, "pid": os.Getpid(), "ttl_seconds": int(heavySlotTTL / time.Second)})
		if err != nil {
			return false, err
		}
		defer resp.Body.Close()
		var out struct {
			Granted *bool `json:"granted"`
		}
		if resp.StatusCode != http.StatusOK || json.NewDecoder(resp.Body).Decode(&out) != nil || out.Granted == nil {
			return false, fmt.Errorf("status %d", resp.StatusCode)
		}
		return *out.Granted, nil
	}
	start := time.Now()
	defer func() { RecordWait(env, time.Since(start)) }()
	for told := false; ; time.Sleep(poll) {
		granted, err := acquire()
		if err != nil {
			fmt.Fprintf(stderr, "%s: flowd unreachable for a heavy slot at %s (%v) — running without one\n", name, addr, err)
			return none, true
		}
		if granted {
			break
		}
		if time.Since(start) >= wait {
			fmt.Fprintf(stderr, "%s: no heavy slot freed after %s — no command ran; re-run once `curl -s %s/api/v1/slots` lists a free one\n", name, wait, addr)
			return none, false
		}
		if !told {
			fmt.Fprintf(stderr, "%s: every heavy slot is held (`curl -s %s/api/v1/slots`) — waiting up to %s for one\n", name, addr, wait)
			told = true
		}
	}
	stop, done := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(done)
		t := time.NewTicker(renew)
		defer t.Stop()
		for {
			select {
			case <-stop:
				return
			case <-t.C:
				_, _ = acquire()
			}
		}
	}()
	var once sync.Once
	return func() {
		once.Do(func() {
			close(stop)
			<-done
			if resp, err := post("/api/v1/slots/release", map[string]string{"id": id}); err == nil {
				_ = resp.Body.Close()
			}
		})
	}, true
}
