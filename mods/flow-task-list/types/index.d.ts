// pending: a plan task of the running /flow change not yet dispatched, read from its tasks.md, never stored;
// in review: only a main-loop status line says it.
export type RowState = 'in progress' | 'in review' | 'done' | 'blocked' | 'pending'
// n: the subagent's spawn number in this session. start and end: `$.clock.now()` as its spawn resolved and as
// it finished. unit: a status-line row's unit, shown as written.
export type Row = { id: string; n: number; desc: string; state: RowState; start?: number; end?: number; unit?: string }
// A main-loop status line's unit as last written and its state.
export type StatusLine = { unit: string; state: RowState }
// The main agent's running turn: its latest Bash description and its "opus-high" run label.
export type Main = { desc: string; run: string }
// The pipeline phase a /flow run is in, from the `flow stage` marks it runs.
export type Phase = 'flow-plan' | 'flow-implement' | 'flow-integrate'
// The running /flow phase, its change name and the tracker key ("KAN-873") that name starts with, if any.
export type Flow = { phase: Phase | null; change: string | null; ticket: string | null }

declare module 'claude-code' {
  interface PluginState {
    'flow-task-list': {
      rows: Row[]
      // agent id → "opus-high": the model and effort its first request went out with.
      runs: Record<string, string>
      flow: Flow
      // The main agent's turn while it runs; null between turns.
      main: Main | null
      // The main loop's latest status line per unit, keyed by the unit without its "(few words)"; a done line deletes its key.
      lines: Record<string, StatusLine>
    }
  }
}
