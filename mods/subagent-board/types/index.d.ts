export type RowState = 'in progress' | 'done' | 'blocked'
// n: the subagent's spawn number in this session.
export type Row = { id: string; n: number; desc: string; state: RowState }
// The pipeline phase a /flow run is in, from the `flow stage` marks it runs.
export type Phase = 'flow-plan' | 'flow-implement' | 'flow-integrate'
// The running /flow phase, the tracker key ("KAN-873") its change name starts with, if any, and the
// flow.* stage key currently begun and not yet ended.
export type Flow = { phase: Phase | null; ticket: string | null; stage: string | null }

declare module 'claude-code' {
  interface PluginState {
    'subagent-board': {
      rows: Row[]
      // agent id → "opus-high": the model and effort its first request went out with.
      runs: Record<string, string>
      flow: Flow
    }
  }
}
