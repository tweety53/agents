package main

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"time"

	"github.com/tweety53/agents/stats/internal/client"
	"github.com/tweety53/agents/stats/internal/fallback"
	"github.com/tweety53/agents/stats/internal/harvest"
)

const statsTraceUsage = `flow stats trace [-addr url] [-timeout dur] [-project key] [-no-transcripts] [-o file] <change>

trace prints one change as Chrome Trace Event Format JSON -- drop the file
on https://ui.perfetto.dev or chrome://tracing for a zoomable waterfall.
One process is the change; its threads are lanes: request, stage, one per
dispatch (role and dispatch key), that dispatch's tool calls, guard and
verify. Each row is a complete event ("ph":"X", ts and dur in
microseconds from the change's first row); rows that overlap within a lane
spill into "<lane> (2)", ... A row still open ends at the trace's last
instant, marked "open" in its args. Tool calls come from each dispatch's
own subagent transcript under ~/.claude/projects (FLOW_TRANSCRIPTS_DIR
overrides it), one span per tool_use to its tool_result, kept inside the
dispatch; -no-transcripts leaves them out. Each stage run and dispatch
carries the transcript time split flowd stored for it into llmMs, toolMs
(toolMsBy per tool, Bash by command word; toolLabels names its harness
wait keys, peer and task-notification) and humanMs (harvest.TimeSplit
states the rule). The project defaults to the one
flowd knows the change under. -o writes the JSON to a file and prints its
event count. Exits 1 when flowd cannot be reached or has no such change, 2
on a usage mistake.
`

// traceSpan is one row of flowd's GET /api/v1/stats/trace.
type traceSpan struct {
	Kind    string     `json:"kind"`
	Lane    string     `json:"lane"`
	Name    string     `json:"name"`
	AgentID string     `json:"agentId"`
	Outcome string     `json:"outcome"`
	Start   time.Time  `json:"start"`
	End     *time.Time `json:"end"`
	// Split is the row's stored harvest.TimeSplit, null where none is.
	Split *harvest.TimeSplit `json:"split"`
}

// chromeEvent is one Trace Event Format event: a complete event (ph X)
// or a process/thread metadata event (ph M).
type chromeEvent struct {
	Name string         `json:"name"`
	Cat  string         `json:"cat,omitempty"`
	Ph   string         `json:"ph"`
	Ts   int64          `json:"ts"`
	Dur  *int64         `json:"dur,omitempty"`
	Pid  int            `json:"pid"`
	Tid  int            `json:"tid"`
	Args map[string]any `json:"args,omitempty"`
}

type chromeTrace struct {
	DisplayTimeUnit string            `json:"displayTimeUnit"`
	OtherData       map[string]string `json:"otherData"`
	TraceEvents     []chromeEvent     `json:"traceEvents"`
}

// traceKinds is the order lanes appear in, top to bottom.
var traceKinds = []string{"request", "stage", "dispatch", "guard", "verify"}

func runStatsTrace(ctx context.Context, args []string, getenv func(string) string, stdout, stderr io.Writer) int {
	var f storeConnFlags
	fset := flag.NewFlagSet("flow stats trace", flag.ContinueOnError)
	fset.SetOutput(stderr)
	registerStoreConnFlags(fset, &f, getenv)
	project := fset.String("project", "", "the change's project (default: the one flowd knows it under)")
	noTranscripts := fset.Bool("no-transcripts", false, "leave out the tool-call lanes")
	outFile := fset.String("o", "", "write the trace to this file")
	if err := fset.Parse(args); err != nil || fset.NArg() != 1 {
		fmt.Fprint(stderr, statsTraceUsage)
		return 2
	}
	change := fset.Arg(0)
	if *project == "" {
		dirKey, _, _ := fallback.ProjectKey(".")
		*project = changeProject(ctx, f.addr, f.timeout, change, dirKey)
	}
	raw, err := withFlowdClient(ctx, f, func(ctx context.Context, cl *client.Client) (json.RawMessage, error) {
		return cl.TraceSpans(ctx, *project, change)
	})
	var spans []traceSpan
	if err == nil {
		err = json.Unmarshal(raw, &spans)
	}
	if err != nil {
		fmt.Fprintf(stderr, "flow: stats trace: %v\n", err)
		return 1
	}
	root := ""
	if !*noTranscripts {
		if root = getenv(harvest.DefaultTranscriptsRootEnv); root == "" {
			root, _ = harvest.DefaultTranscriptsRoot()
		}
	}
	trace := buildTrace(*project, change, spans, root, stderr)
	body, err := encodeTrace(trace)
	if err != nil {
		fmt.Fprintf(stderr, "flow: stats trace: %v\n", err)
		return 1
	}
	if *outFile == "" {
		_, _ = stdout.Write(body)
		return 0
	}
	if err := os.WriteFile(*outFile, body, 0o644); err != nil {
		fmt.Fprintf(stderr, "flow: stats trace: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "%s: %d events\n", *outFile, len(trace.TraceEvents))
	return 0
}

// buildTrace lays spans out as one process whose threads are lanes, reading
// each dispatch's tool calls from its transcript under root ("" for none).
// Lanes get tids in first-use order, kinds in traceKinds order, a
// dispatch's tool lane right after its own; a span overlapping the lane's
// last one spills into the next free "<lane> (n)".
func buildTrace(project, change string, spans []traceSpan, root string, stderr io.Writer) chromeTrace {
	slices.SortStableFunc(spans, func(a, b traceSpan) int {
		return cmp.Or(cmp.Compare(slices.Index(traceKinds, a.Kind), slices.Index(traceKinds, b.Kind)),
			a.Start.Compare(b.Start), cmp.Compare(a.Name, b.Name))
	})
	type toolSpan struct {
		of   int // index into spans
		call harvest.ToolCall
	}
	var tools []toolSpan
	var t0, last time.Time
	stretch := func(start, end time.Time) {
		if t0.IsZero() || start.Before(t0) {
			t0 = start
		}
		last = latest(last, start, end)
	}
	calls := map[string][]harvest.ToolCall{}
	for i, s := range spans {
		end := time.Time{}
		if s.End != nil {
			end = *s.End
		}
		stretch(s.Start, end)
		if root == "" || s.Kind != "dispatch" || s.AgentID == "" {
			continue
		}
		cs, seen := calls[s.AgentID]
		if !seen {
			var err error
			if cs, _, err = harvest.AgentToolCalls(root, s.AgentID); err != nil {
				fmt.Fprintf(stderr, "flow: stats trace: agent %s transcript: %v\n", s.AgentID, err)
			}
			calls[s.AgentID] = cs
		}
		for _, c := range cs {
			if c.Start.Before(s.Start) || (s.End != nil && !c.Start.Before(*s.End)) {
				continue
			}
			if s.End != nil && c.End.After(*s.End) {
				c.End = *s.End
			}
			stretch(c.Start, c.End)
			tools = append(tools, toolSpan{of: i, call: c})
		}
	}

	type sublane struct {
		tid int
		end time.Time
	}
	lanes := map[string][]sublane{}
	var names []string // tid-1 -> lane name
	place := func(base string, start, end time.Time) int {
		for i, sl := range lanes[base] {
			if !start.Before(sl.end) {
				lanes[base][i].end = end
				return sl.tid
			}
		}
		name := base
		if n := len(lanes[base]); n > 0 {
			name = fmt.Sprintf("%s (%d)", base, n+1)
		}
		names = append(names, name)
		lanes[base] = append(lanes[base], sublane{tid: len(names), end: end})
		return len(names)
	}
	us := func(t time.Time) int64 { return t.Sub(t0).Microseconds() }
	var events []chromeEvent
	emit := func(cat, lane, name string, start, end time.Time, args map[string]any) {
		dur := max(0, us(end)-us(start))
		events = append(events, chromeEvent{Name: name, Cat: cat, Ph: "X", Ts: us(start), Dur: &dur, Pid: 1, Tid: place(lane, start, end), Args: args})
	}
	for i, s := range spans {
		args := map[string]any{}
		end := last
		if s.End != nil {
			end = *s.End
		} else {
			args["open"] = true
		}
		if s.Outcome != "" {
			args["outcome"] = s.Outcome
		}
		if s.AgentID != "" {
			args["agentId"] = s.AgentID
		}
		if split := s.Split; split != nil {
			args["llmMs"], args["toolMs"], args["humanMs"] = split.LLMMs, split.ToolMs, split.HumanMs
			if split.ByTool != nil {
				args["toolMsBy"] = split.ByTool
				labels := map[string]string{}
				for k := range split.ByTool {
					if l := harvest.ToolLabels[k]; l != "" {
						labels[k] = l
					}
				}
				if len(labels) > 0 {
					args["toolLabels"] = labels
				}
			}
		}
		emit(s.Kind, s.Lane, s.Name, s.Start, end, args)
		for _, t := range tools {
			if t.of != i {
				continue
			}
			var args map[string]any
			if t.call.IsError {
				args = map[string]any{"error": true}
			}
			emit("tool", s.Lane+" tools", t.call.Name, t.call.Start, t.call.End, args)
		}
	}

	meta := []chromeEvent{{Name: "process_name", Ph: "M", Pid: 1, Args: map[string]any{"name": project + "/" + change}}}
	for i, n := range names {
		meta = append(meta,
			chromeEvent{Name: "thread_name", Ph: "M", Pid: 1, Tid: i + 1, Args: map[string]any{"name": n}},
			chromeEvent{Name: "thread_sort_index", Ph: "M", Pid: 1, Tid: i + 1, Args: map[string]any{"sort_index": i + 1}})
	}
	other := map[string]string{"project": project, "change": change}
	if !t0.IsZero() {
		other["t0"] = t0.UTC().Format(time.RFC3339Nano)
	}
	return chromeTrace{DisplayTimeUnit: "ms", OtherData: other, TraceEvents: append(meta, events...)}
}

// encodeTrace is trace's JSON with one event per line, so a diff of two
// traces reads event by event.
func encodeTrace(trace chromeTrace) ([]byte, error) {
	events := trace.TraceEvents
	trace.TraceEvents = nil
	head, err := json.Marshal(trace)
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	b.Write(head[:len(head)-len(`"traceEvents":null}`)])
	b.WriteString(`"traceEvents":[`)
	for i, e := range events {
		line, err := json.Marshal(e)
		if err != nil {
			return nil, err
		}
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString("\n")
		b.Write(line)
	}
	b.WriteString("\n]}\n")
	return b.Bytes(), nil
}

// latest is the latest of ts, zero times ignored.
func latest(ts ...time.Time) time.Time {
	var out time.Time
	for _, t := range ts {
		if t.After(out) {
			out = t
		}
	}
	return out
}
