// Package config resolves myflowd's daemon configuration -- the address it
// binds and the database it connects to -- from the environment, and
// enforces the one non-negotiable rule about that address: it must be
// loopback-only. Binding any other interface is a configuration error the
// daemon refuses at startup, not a runtime option, per design.md's "Bind
// loopback only" requirement.
package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
)

// Defaults, applied when the corresponding environment variable is unset.
// DefaultDSN points at the dedicated flow-postgres compose stack's
// declaration (stats/docker-compose.yml), on host port 5433. It no longer
// matches the "myflow" DSN stats/health_test.go and internal/store's test
// helpers still hardcode: the running container has not been renamed yet
// (an operator step, KAN-289's task 19), so those tests keep targeting the
// role and database that container actually has, while this default states
// the name it will answer to once the operator renames it.
const (
	DefaultHost = "127.0.0.1"
	DefaultPort = 4173
	DefaultDSN  = "postgres://flow:flow@localhost:5433/flow?sslmode=disable"
)

// Config is myflowd's resolved configuration.
type Config struct {
	Host string
	Port int
	DSN  string
}

// ErrNonLoopbackHost is returned by Validate when Host does not resolve to
// a loopback address. Binding any other interface is a configuration error
// the daemon refuses at startup, never a runtime option -- see design.md,
// "Bind loopback only".
var ErrNonLoopbackHost = errors.New("config: host must be a loopback address (127.0.0.1 or ::1)")

// ErrInvalidPort is returned by FromEnv when FLOWD_PORT is set to
// something that does not parse as an integer. This is deliberately a
// startup error rather than a discarded parse failure: a silently-ignored
// bad value would start the daemon on DefaultPort with no diagnostic at
// all, which is a working-but-wrong-port daemon -- exactly the kind of
// failure that costs an hour to notice, the same class of mistake
// ErrNonLoopbackHost exists to catch for the host.
var ErrInvalidPort = errors.New("config: FLOWD_PORT must be a valid integer port")

// FromEnv resolves Config from FLOWD_HOST, FLOWD_PORT and FLOWD_DSN,
// falling back to the defaults above for anything unset. A FLOWD_PORT
// that fails to parse is refused with ErrInvalidPort rather than silently
// falling back to DefaultPort. FromEnv does not check Host for loopback --
// callers building a server from the result must call Validate (api.New
// does this itself).
func FromEnv() (Config, error) {
	cfg := Config{Host: DefaultHost, Port: DefaultPort, DSN: DefaultDSN}
	if v := os.Getenv("FLOWD_HOST"); v != "" {
		cfg.Host = v
	}
	if v := os.Getenv("FLOWD_PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, fmt.Errorf("%w: %q", ErrInvalidPort, v)
		}
		cfg.Port = p
	}
	if v := os.Getenv("FLOWD_DSN"); v != "" {
		cfg.DSN = v
	}
	return cfg, nil
}

// Validate refuses a Config whose Host does not parse as a loopback IP
// address. It is the one gate every caller that opens a listener --
// api.New in particular -- must pass through first, so a misconfigured
// bind address is caught before net.Listen is ever called, not after.
//
// Host must be a literal IP address ("127.0.0.1" or "::1"), not a name
// like "localhost": resolving a name can depend on the host's own
// /etc/hosts or DNS configuration, and this check must be able to answer
// the loopback question on the string alone.
func (c Config) Validate() error {
	ip := net.ParseIP(c.Host)
	if ip == nil || !ip.IsLoopback() {
		return fmt.Errorf("%w: %q", ErrNonLoopbackHost, c.Host)
	}
	return nil
}

// Addr returns the host:port string to bind, suitable for net.Listen.
func (c Config) Addr() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}
