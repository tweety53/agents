package api

import (
	"context"
	"time"

	"github.com/tweety53/agents/stats/internal/jira"
)

// WithStatsTimeout replaces statsWriteTimeout for one server, so a test
// proving the timeout's clean 503 does not wait the real 20s out.
func WithStatsTimeout(d time.Duration) Option {
	return func(o *serverOptions) {
		o.statsTimeout = d
	}
}

// WithJiraTransitioner replaces the Jira client New builds from cfg.Jira,
// so a test of the handler's error mapping need not run the real client's
// retry ladder in real time.
func WithJiraTransitioner(j JiraTransitioner) Option {
	return func(o *serverOptions) {
		o.jira = j
	}
}

// JiraTransitionFunc adapts a function to JiraTransitioner.
type JiraTransitionFunc func(ctx context.Context, issueKey string, target jira.Position) (jira.Result, error)

// Transition calls f.
func (f JiraTransitionFunc) Transition(ctx context.Context, issueKey string, target jira.Position) (jira.Result, error) {
	return f(ctx, issueKey, target)
}
