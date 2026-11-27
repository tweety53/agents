package setuptest

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// TestLoopHookTimeoutsFailClosed pins that every snippet setup.sh prints for
// the loop dispatch hook registers an explicit timeout, and that the hook's
// own subprocess wait stays well under it: a harness that times a hook out
// kills it, so the hook must give up first and log what went unrecorded.
func TestLoopHookTimeoutsFailClosed(t *testing.T) {
	setup, err := os.ReadFile("../../../setup.sh")
	if err != nil {
		t.Fatal(err)
	}
	reg := regexp.MustCompile(`"timeout": ([0-9]+)`)
	wait := regexp.MustCompile(`(?m)^const [A-Za-z]*Timeout = ([0-9]+) \* time\.Second`)
	// Each hook's Go port, the file whose wait the shim hooks/<hook>.py runs.
	for hook, src := range map[string]string{"enforce-loop-dispatch": "hookloopdispatch.go"} {
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
		src, err := os.ReadFile("../guard/" + src)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(src), "300 * time.Second") {
			t.Errorf("%s waits a literal 300s on a subprocess", hook)
		}
		m := wait.FindStringSubmatch(string(src))
		if m == nil {
			t.Fatalf("%s names no *Timeout constant", hook)
		}
		w, _ := strconv.Atoi(m[1])
		for _, r := range regs {
			if w > r*4/5 {
				t.Errorf("%s waits %ds against a registered %ds", hook, w, r)
			}
		}
	}
}
