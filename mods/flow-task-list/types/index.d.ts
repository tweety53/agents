// pending: a plan task of the running /flow change not yet dispatched, read from its tasks.md, never stored, or the
// main agent between its turns; in review: only a main-loop status line says it.
export type RowState = 'in progress' | 'in review' | 'done' | 'blocked' | 'pending'
// n: the subagent's spawn number in this session. start and end: `$.clock.now()` as its spawn resolved and as
// it finished. change: the /flow change marked when it was spawned, if any.
export type Row = { id: string; n: number; desc: string; state: RowState; start?: number; end?: number; change?: string | null }
// A main-loop status line's unit as last written and its state.
export type StatusLine = { unit: string; state: RowState }
// The main agent: its "opus-high" run label, whether its turn is running, and `$.clock.now()` as its latest turn began.
export type Main = { run: string; busy: boolean; since?: number }
// The pipeline phase a /flow run is in, from the `flow stage` marks it runs.
export type Phase = 'flow-plan' | 'flow-implement' | 'flow-integrate'
// The running /flow phase, its change name, the tracker key ("KAN-873") that name starts with, if any, and the
// stage key its latest `flow stage begin` named ("sdd-tdd").
export type Flow = { phase: Phase | null; change: string | null; ticket: string | null; stage: string | null }

declare module 'claude-code' {
  interface PluginState {
    'flow-task-list': {
      rows: Row[]
      // agent id → "opus-high": the model and effort its first request went out with.
      runs: Record<string, string>
      flow: Flow
      // The main agent, from its first model request on; null before.
      main: Main | null
      // The main loop's latest status line per unit, keyed by its leading name (`unitKey`), the latest written last; a done line deletes its key, and a blocked one lasts until the next main turn starts.
      lines: Record<string, StatusLine>
    }
  }
}
