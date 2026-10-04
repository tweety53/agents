// pending: a plan task of the running /flow change not yet dispatched, read from its tasks.md, never stored.
export type RowState = 'in progress' | 'done' | 'blocked' | 'pending'
// n: the subagent's spawn number in this session.
export type Row = { id: string; n: number; desc: string; state: RowState }
// The main agent's running turn: its latest Bash description and its "opus-high" run label.
export type Main = { desc: string; run: string }
// The pipeline phase a /flow run is in, from the `flow stage` marks it runs.
export type Phase = 'flow-plan' | 'flow-implement' | 'flow-integrate'
// The running /flow phase, its change name, the tracker key ("KAN-873") that name starts with, if any,
// and the flow.* stage key currently begun and not yet ended.
export type Flow = { phase: Phase | null; change: string | null; ticket: string | null; stage: string | null }

declare module 'claude-code' {
  interface PluginState {
    'subagent-board': {
      rows: Row[]
      // agent id → "opus-high": the model and effort its first request went out with.
      runs: Record<string, string>
      flow: Flow
      // The main agent's turn while it runs; null between turns.
      main: Main | null
    }
  }
}
