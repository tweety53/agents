// Package config resolves flowd's daemon configuration -- the address it
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
	"strings"
)

// Defaults, applied when the corresponding environment variable is unset.
// DefaultDSN points at the dedicated flow-postgres compose stack's
// declaration (stats/docker-compose.yml), on host port 5433. It no longer
// matches the "flow" DSN stats/health_test.go and internal/store's test
// helpers still hardcode: the running container has not been renamed yet
// (an operator step, KAN-289's task 19), so those tests keep targeting the
// role and database that container actually has, while this default states
// the name it will answer to once the operator renames it.
const (
	DefaultHost = "127.0.0.1"
	DefaultPort = 4173
	DefaultDSN  = "postgres://flow:flow@localhost:5433/flow?sslmode=disable"
)

// Config is flowd's resolved configuration.
type Config struct {
	Host string
	Port int
	DSN  string

	// Jira is the outbound access the pipeline-transition endpoint needs
	// (internal/jira). Zero value means the daemon has none, which is a
	// normal state the endpoint reports per request -- never a startup
	// failure.
	Jira JiraConfig
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
	jira, err := resolveJira()
	if err != nil {
		return Config{}, err
	}
	cfg.Jira = jira
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

// JiraConfig is the outbound Jira access the pipeline-transition endpoint
// needs (internal/jira): the site's base URL and the basic-auth
// credentials. The three variables are one unit -- a site without its
// credentials, or credentials without a site, can never work -- so the
// configuration is either complete or refused, never half-applied.
type JiraConfig struct {
	Site     string
	Email    string
	APIToken string
}

// Configured reports whether the Jira block is present. resolveJira only
// ever produces an all-or-nothing block, so any non-empty field implies
// all three.
func (c JiraConfig) Configured() bool {
	return c.Site != ""
}

// Errors for the Jira environment block. Like ErrInvalidPort, both are
// startup refusals: a partially-set block or a non-https site would
// otherwise degrade silently into a daemon that looks configured and
// answers every transition with a credentials or URL failure. The site is
// held to https because the Atlassian Cloud API is always https, so an
// http:// site is a typo worth catching before any request is built.
var (
	ErrPartialJiraConfig = errors.New("config: FLOWD_JIRA_SITE, FLOWD_JIRA_EMAIL and FLOWD_JIRA_TOKEN must be set together")
	ErrInvalidJiraSite   = errors.New("config: FLOWD_JIRA_SITE must be an https:// URL")
)

// resolveJira reads FLOWD_JIRA_SITE, FLOWD_JIRA_EMAIL and FLOWD_JIRA_TOKEN.
// Unset everywhere means the daemon has no Jira access -- a normal state,
// not an error. Any other incomplete combination is refused with
// ErrPartialJiraConfig; a site that is not https://, with
// ErrInvalidJiraSite.
func resolveJira() (JiraConfig, error) {
	site := os.Getenv("FLOWD_JIRA_SITE")
	email := os.Getenv("FLOWD_JIRA_EMAIL")
	token := os.Getenv("FLOWD_JIRA_TOKEN")
	if site == "" && email == "" && token == "" {
		return JiraConfig{}, nil
	}
	if site == "" || email == "" || token == "" {
		return JiraConfig{}, fmt.Errorf("%w (site set: %t, email set: %t, token set: %t)",
			ErrPartialJiraConfig, site != "", email != "", token != "")
	}
	if !strings.HasPrefix(site, "https://") {
		return JiraConfig{}, fmt.Errorf("%w: %q", ErrInvalidJiraSite, site)
	}
	return JiraConfig{Site: strings.TrimRight(site, "/"), Email: email, APIToken: token}, nil
}
