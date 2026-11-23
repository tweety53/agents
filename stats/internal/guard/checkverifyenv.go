package guard

// flow-guard check-verify-env — the verify environment's preflight: before a
// verify or visual-verify round runs, prove every backend the suite talks to
// accepts a connection, so a backend that is down (KAN-1157: 15+ unrelated
// specs failing on `ECONNREFUSED ::1:8650`) reads as an environment failure,
// cured before the run, and never as a failed fix round.
//
// Usage: flow-guard check-verify-env <url>...
//
// No project.md key declares backend URLs, so the caller passes each one the
// project's `## run` section starts. Each URL's host:port is dialled over TCP.
//
// Exit codes:
//   0  every URL accepted a connection; prints VERIFY-ENV-OK.
//   1  at least one URL refused or timed out; each is named on stderr.
//   2  usage — no URL, or one that parses without a host.

import (
	"fmt"
	"io"
	"net"
	"net/url"
	"time"
)

func init() {
	Registry["check-verify-env"] = checkVerifyEnv
}

func checkVerifyEnv(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		Reason(env, "usage")
		fmt.Fprintln(stderr, "usage: flow-guard check-verify-env <url>...")
		return 2
	}
	down := 0
	for _, a := range args {
		u, err := url.Parse(a)
		if err != nil || u.Hostname() == "" {
			Reason(env, "usage")
			fmt.Fprintf(stderr, "check-verify-env: %s is not a URL with a host\n", a)
			return 2
		}
		port := u.Port()
		if port == "" {
			port = map[string]string{"https": "443"}[u.Scheme]
			if port == "" {
				port = "80"
			}
		}
		c, err := net.DialTimeout("tcp", net.JoinHostPort(u.Hostname(), port), 3*time.Second)
		if err != nil {
			down++
			fmt.Fprintf(stderr, "VERIFY-ENV-DOWN: %s — %v\n", a, err)
			continue
		}
		c.Close()
	}
	if down > 0 {
		Reason(env, "backend-unreachable")
		return 1
	}
	fmt.Fprintf(stdout, "VERIFY-ENV-OK: %d reachable\n", len(args))
	return 0
}
