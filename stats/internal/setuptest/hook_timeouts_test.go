package setuptest

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// TestLoopHookTimeoutsFailClosed pins that every snippet setup.sh prints for
// the two fail-closed loop hooks registers an explicit timeout, and that each
// hook's own subprocess wait stays well under it: a harness that times a hook
// out lets the call through, so the hook must give up first and deny.
func TestLoopHookTimeoutsFailClosed(t *testing.T) {
	setup, err := os.ReadFile("../../../setup.sh")
	if err != nil {
		t.Fatal(err)
	}
	reg := regexp.MustCompile(`"timeout": ([0-9]+)`)
	wait := regexp.MustCompile(`(?m)^[A-Z_]*TIMEOUT = ([0-9]+)`)
	for _, hook := range []string{"enforce-loop-dispatch", "protect-main-checkout"} {
		var regs []int
		for _, line := range strings.Split(string(setup), "\n") {
			if !strings.Contains(line, "hooks/"+hook+".py") {
				continue
			}
			m := reg.FindStringSubmatch(line)
			if m == nil {
				t.Errorf("setup.sh registers %s with no timeout: %s", hook, strings.TrimSpace(line))
				continue
			}
			n, _ := strconv.Atoi(m[1])
			regs = append(regs, n)
		}
		if len(regs) == 0 {
			t.Fatalf("setup.sh prints no snippet for %s", hook)
		}
		src, err := os.ReadFile("../../../hooks/" + hook + ".py")
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(src), "timeout=300") {
			t.Errorf("%s waits a literal 300s on a subprocess", hook)
		}
		m := wait.FindStringSubmatch(string(src))
		if m == nil {
			t.Fatalf("%s names no *TIMEOUT constant", hook)
		}
		w, _ := strconv.Atoi(m[1])
		for _, r := range regs {
			if w > r*4/5 {
				t.Errorf("%s waits %ds against a registered %ds", hook, w, r)
			}
		}
	}
}
