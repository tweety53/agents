package guard

import (
	"net"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

func TestCheckVerifyEnv(t *testing.T) {
	t.Parallel()
	up, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { up.Close() })
	down, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	downAddr := down.Addr().String()
	down.Close()
	upURL, downURL := "http://"+up.Addr().String()+"/", "http://"+downAddr

	cases := []struct {
		name string
		args []string
		rc   int
		want string
	}{
		{"reachable", []string{upURL}, 0, "VERIFY-ENV-OK"},
		{"refused", []string{upURL, downURL}, 1, downURL},
		{"no args", nil, 2, "usage"},
		{"bad url", []string{"::nope"}, 2, "::nope"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			r := runGuard("check-verify-env", c.args, core.Env{})
			if r.Rc != c.rc || !strings.Contains(r.Out, c.want) {
				t.Fatalf("rc=%d out=%q, want rc=%d containing %q", r.Rc, r.Out, c.rc, c.want)
			}
		})
	}
}
