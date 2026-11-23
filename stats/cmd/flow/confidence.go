package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"time"

	"github.com/tweety53/agents/stats/internal/client"
	"github.com/tweety53/agents/stats/internal/records"
)

const confidenceUsage = `usage: flow confidence tags [-addr url] [-timeout dur]

Prints every tag in use on the recorded confidence estimates, one
"<count> <tag>" line each, alphabetically; nothing when none is recorded.
Run it before "flow confidence record": reuse a printed tag only when it names
the same kind of gap, and write a new lowercase kebab-case tag otherwise. Exit 1
when the store cannot answer.

usage: flow confidence after [-addr url] [-timeout dur] [-C dir]
                             -change <name> -task <ids> -score <1-100>

Sets the estimate held once the task's tests are green on the latest estimate
that task (the dispatch's ids, "2+3") filed. Exit 2 when the task filed none.

usage: flow confidence record [-addr url] [-timeout dur] [-C dir]
                              -change <name> -stage <plan|implement> -score <1-100>
                              -gap <text> [-gap <text>…] -tag <tag> [-tag <tag>…]
                              [-task <ids>] [-disposition <raised|later>]

Files one confidence estimate (KAN-1130): how confident the caller is, 1-100,
that the plan is executed as written, each -gap one thing that would move it
to 100, in full. -stage plan takes -disposition: raised when the operator had
the gaps closed, later when they were deferred. -stage implement takes -task,
the dispatch's task ids ("2+3"). At 100 pass neither -gap nor -tag: the estimate is kept as a plain 100%, outside the review list. A filed estimate is always "to review": its
review status is set on the Confidence dashboard, never here. An unreachable
store is a warning and exit 0; a refused estimate is exit 2.
`

// changeProject is the project key the store holds change under: an estimate from a worktree of another
// repository of a multi-repository change resolves its own repository's key from -C, a key the run page never
// reads (KAN-1156: a frontend implementer's row sat under the frontend key). Exactly one record named change
// is authoritative; none or several, or a store that cannot answer, keeps the key -C gave.
func changeProject(ctx context.Context, addr string, timeout time.Duration, change, dirKey string) string {
	found, err := callRecord(ctx, addr, timeout, func(ctx context.Context, cl *client.Client) ([]client.StateChange, error) {
		return cl.FindState(ctx, change)
	})
	if err == nil && len(found) == 1 {
		return found[0].ProjectKey
	}
	return dirKey
}

// runConfidence implements `flow confidence`: `tags`, the vocabulary read,
// and `record`, the write.
func runConfidence(ctx context.Context, args []string, getenv func(string) string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, confidenceUsage)
		return 2
	}
	fset := flag.NewFlagSet("flow confidence "+args[0], flag.ContinueOnError)
	fset.SetOutput(stderr)
	var f recordIdentityFlags
	registerRecordConnFlags(fset, &f, getenv)
	switch args[0] {
	case "tags":
		if err := fset.Parse(args[1:]); err != nil || fset.NArg() != 0 {
			if errors.Is(err, flag.ErrHelp) {
				return 0
			}
			fmt.Fprint(stderr, confidenceUsage)
			return 2
		}
		tags, err := callRecord(ctx, f.addr, f.timeout, func(ctx context.Context, cl *client.Client) ([]records.ConfidenceTag, error) {
			return cl.ConfidenceTags(ctx)
		})
		if err != nil {
			fmt.Fprintf(stderr, "flow: confidence tags: %v\n", err)
			return 1
		}
		for _, t := range tags {
			fmt.Fprintf(stdout, "%d %s\n", t.Count, t.Tag)
		}
		return 0
	case "after":
		var task string
		var score int
		fset.StringVar(&f.change, "change", "", "the change the estimate belongs to (required)")
		fset.StringVar(&task, "task", "", "the dispatch's task ids (required)")
		fset.IntVar(&score, "score", 0, "the estimate once the tests are green, 1-100 (required)")
		projectKey, _, code := parseSelfReviewFlags(fset, &f, args[1:], getenv, stderr)
		if code >= 0 {
			return code
		}
		projectKey = changeProject(ctx, f.addr, f.timeout, f.change, projectKey)
		if task == "" {
			fmt.Fprintln(stderr, "flow: confidence after needs -task")
			return 2
		}
		_, err := callRecord(ctx, f.addr, f.timeout, func(ctx context.Context, cl *client.Client) (struct{}, error) {
			return struct{}{}, cl.ConfidenceAfter(ctx, projectKey, f.change, task, score)
		})
		switch {
		case err == nil:
			return 0
		case errors.Is(err, client.ErrRecordRejected), errors.Is(err, client.ErrNotFound):
			fmt.Fprintf(stderr, "flow: confidence after refused: %v\n", err)
			return 2
		default:
			fmt.Fprintf(stderr, "⚠ flow: confidence after not recorded — the store could not take it: %v\n", err)
			return 0
		}
	case "record":
		var in records.ConfidenceEstimate
		fset.StringVar(&f.change, "change", "", "the change the estimate belongs to (required)")
		fset.StringVar(&in.Stage, "stage", "", "plan or implement (required)")
		fset.IntVar(&in.Score, "score", 0, "the estimate, 1-100 (required)")
		fset.StringVar(&in.Task, "task", "", "the dispatch's task ids (required for -stage implement)")
		fset.StringVar(&in.Disposition, "disposition", "", "raised or later (required for -stage plan)")
		fset.Func("gap", "one thing that would move the estimate to 100, in full (repeatable)", func(v string) error {
			in.Gaps = append(in.Gaps, v)
			return nil
		})
		fset.Func("tag", "a lowercase kebab-case tag (repeatable)", func(v string) error {
			in.Tags = append(in.Tags, v)
			return nil
		})
		projectKey, _, code := parseSelfReviewFlags(fset, &f, args[1:], getenv, stderr)
		if code >= 0 {
			return code
		}
		projectKey = changeProject(ctx, f.addr, f.timeout, f.change, projectKey)
		_, err := callRecord(ctx, f.addr, f.timeout, func(ctx context.Context, cl *client.Client) (records.ConfidenceEstimate, error) {
			return cl.RecordConfidenceEstimate(ctx, projectKey, f.change, in)
		})
		switch {
		case err == nil:
			return 0
		case errors.Is(err, client.ErrRecordRejected):
			fmt.Fprintf(stderr, "flow: confidence estimate refused: %v\n", err)
			return 2
		default:
			fmt.Fprintf(stderr, "⚠ flow: confidence estimate not recorded — the store could not take it: %v\n", err)
			return 0
		}
	default:
		fmt.Fprintf(stderr, "flow: unknown confidence command %q\n", args[0])
		fmt.Fprint(stderr, confidenceUsage)
		return 2
	}
}
