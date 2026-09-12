package harvest

import "testing"

// Fixture line recorded from a real ZCode subagent rollout file
// (~/.zcode/cli/rollout/model-io-sess_subagent_agent_541a9582-890b-417e-a3c3-419047bb4ee6.jsonl,
// captured 2026-09-12 while the dispatch ran), trimmed of request-body bulk
// only. A subagent dispatch writes its own rollout file named
// model-io-sess_subagent_agent_<uuid>.jsonl and stamps every line with the
// per-dispatch session id sess_subagent_agent_<uuid> plus
// querySource "subagent"; the <uuid> is byte-identical to the agentId the
// dispatcher's Agent tool result returns, which `flow record dispatch begin
// -agent-id` already records on the dispatch row.
const rolloutSubagentLine = `{"completedAt":"2026-09-12T20:00:47.950Z","durationMs":28045,"requestId":"7fe37765","attempt":1,` +
	`"querySource":"subagent","traceId":"94fc46f4","type":"model_io",` +
	`"sessionId":"sess_subagent_agent_541a9582-890b-417e-a3c3-419047bb4ee6",` +
	`"model":{"modelId":"GLM-5.3-Flash","providerId":"builtin:zai-coding-plan","role":"main","source":"session","variant":"high"},` +
	`"response":{"usage":{"inputTokens":48023,"outputTokens":1419,"totalTokens":49442,"cacheReadTokens":44800,"cacheWriteTokens":0}}}`

func TestParseRolloutRecordsSubagentSessionMarksSidechainAndAgentID(t *testing.T) {
	records := ParseRolloutRecords([]byte(rolloutSubagentLine + "\n"))
	if len(records) != 1 {
		t.Fatalf("got %d records, want 1", len(records))
	}
	r := records[0]
	if r.SessionID != "sess_subagent_agent_541a9582-890b-417e-a3c3-419047bb4ee6" {
		t.Errorf("SessionID = %q", r.SessionID)
	}
	if !r.IsSidechain {
		t.Errorf("IsSidechain = false, want true: a subagent dispatch's usage is its dispatcher's sidechain spend, and the dispatch-grain pass only credits sidechain records")
	}
	if r.AgentID != "agent_541a9582-890b-417e-a3c3-419047bb4ee6" {
		t.Errorf("AgentID = %q, want the dispatch's own agent id as the dispatcher recorded it", r.AgentID)
	}
	if r.Model != "GLM-5.3-Flash" || r.Usage.OutputTokens != 1419 {
		t.Errorf("usage decimation changed: Model = %q, Usage = %+v", r.Model, r.Usage)
	}

	t.Run("main-session lines still carry neither field", func(t *testing.T) {
		records := ParseRolloutRecords([]byte(rolloutFullLine + "\n"))
		if len(records) != 1 {
			t.Fatalf("got %d records, want 1", len(records))
		}
		if r := records[0]; r.IsSidechain || r.AgentID != "" {
			t.Errorf("main-session record grew subagent fields: IsSidechain = %v, AgentID = %q", r.IsSidechain, r.AgentID)
		}
	})

	t.Run("a main session id that merely contains subagent does not match", func(t *testing.T) {
		line := `{"type":"model_io","sessionId":"sess_subagent_agentX-1","completedAt":"2026-09-12T20:00:47.950Z","model":{"modelId":"m"},"response":{"usage":{"inputTokens":1,"outputTokens":1,"cacheReadTokens":0,"cacheWriteTokens":0}}}`
		records := ParseRolloutRecords([]byte(line + "\n"))
		if len(records) != 1 {
			t.Fatalf("got %d records, want 1", len(records))
		}
		if r := records[0]; r.IsSidechain || r.AgentID != "" {
			t.Errorf("prefix matched a non-subagent session id: IsSidechain = %v, AgentID = %q", r.IsSidechain, r.AgentID)
		}
	})
}
