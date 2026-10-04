import { atom, read, update } from 'claude-code'
import type { Register } from 'claude-code'

import type { Flow, Main, Phase, Row, RowState } from '../types'

const rows = atom({ plugin: 'subagent-board', key: 'rows' } as const, [] as Row[])
const runs = atom({ plugin: 'subagent-board', key: 'runs' } as const, {} as Record<string, string>)
const NO_FLOW: Flow = { phase: null, change: null, ticket: null, stage: null }
const flow = atom({ plugin: 'subagent-board', key: 'flow' } as const, NO_FLOW)
const main = atom({ plugin: 'subagent-board', key: 'main' } as const, null as Main | null)

const MAX_ROWS = 5

const EMOJI: Record<RowState, string> = { 'in progress': '🔄', done: '✅', blocked: '⛔', pending: '⏳' }
// The task-list look: its marker and colour per state; `claude` is the theme's accent (orange). A pending
// marker keeps the text's own colour.
const MARK: Record<RowState, string> = { 'in progress': '◼', done: '✔', blocked: '✘', pending: '◻' }
const MARK_COLOR: Record<RowState, string | undefined> = { 'in progress': 'claude', done: 'success', blocked: 'error', pending: undefined }

// A description that already carries its plan numbering keeps it: one task, "Task 3/22 (spec text)",
// or a /flow task group, "Tasks 3+4+7/22 (port guards)" (skills/flow/implement.md, **Dispatch sites**).
const NUMBERED = /^(Tasks? \d+(?:\+\d+)*\/\d+)\s*/

// What a running row is doing, read from its description, and the state word it shows; first match
// wins, so a panel fix is a fix. Each word is matched from its start, so "prefix", "preview" or
// "latest" names no kind.
const KINDS: [RegExp, string, string][] = [
  [/\bfix/i, '🔨', 'fix'],
  [/\bvisual[- ]verif/i, '👀', 'visual verify'],
  [/\b(?:review|panel)/i, '🔍', 'in review'],
  [/\b(?:verif|tests?\b|lint)/i, '🧪', 'verify'],
]

// How a row shows: a running row by its kind's word (in review, fix, visual verify, verify), a finished one by its state.
// `emoji` is what the hint line's tally counts it under; `mark` and `color` are its marker on the band.
export const look = (row: Row): { emoji: string; word: string; mark: string; color: string | undefined } => {
  const kind = row.state === 'in progress' ? KINDS.find(([re]) => re.test(row.desc)) : undefined
  return {
    emoji: kind?.[1] ?? EMOJI[row.state],
    word: kind?.[2] ?? row.state,
    mark: MARK[row.state],
    color: MARK_COLOR[row.state],
  }
}

// One board row's pieces; `lead` is "⎿ " on the first row and its width in spaces after, and `run`
// is padded to `width` so the marker column lines up. The main agent's row, n 0, is "main", never a task.
export const parts = (row: Row, total: number, run = '', width = 0, first = true) => {
  const m = NUMBERED.exec(row.desc)
  const desc = m ? row.desc.slice(m[0].length).replace(/^\((.*)\)$/, '$1') : row.desc
  const l = look(row)
  return {
    lead: first ? '⎿ ' : '  ',
    run: run.padEnd(width),
    mark: l.mark,
    unit: m?.[1] ?? (row.n === 0 ? 'main' : `Task ${row.n}/${total}`),
    desc: desc ? ` (${desc})` : '',
    state: l.word,
    color: l.color,
  }
}

export const line = (row: Row, total: number, run = '', width = 0, first = true): string => {
  const p = parts(row, total, run, width, first)
  return `${p.lead}${p.run ? `${p.run} ` : ''}${p.mark} ${p.unit}${p.desc} — ${p.state}`
}

// A plan's column-0 task line, "- [ ] 23. Title" or "- [x] 23. Title"; its step checkboxes are indented.
const TASK = /^- \[([ x])\] (\d+)\. (.*)$/gm

// The plan's unticked tasks no board row names in its "Task(s) …/n" numbering, as pending rows: what the
// running /flow change has not dispatched yet, derived from its tasks.md on every draw.
export const pendingRows = (tasksMd: string, rs: Row[]): Row[] => {
  const tasks = [...tasksMd.matchAll(TASK)]
  const taken = new Set(rs.flatMap(r => NUMBERED.exec(r.desc)?.[1]?.match(/\d+(?=[+/])/g) ?? []))
  return tasks
    .filter(([, box, n = '']) => box === ' ' && !taken.has(n))
    .map(([, , n = '', title = '']) => ({ id: `pending-${n}`, n: Number(n), desc: `Task ${n}/${tasks.length} (${title})`, state: 'pending' }))
}

// "opus-high" from the model id and effort a request went out with.
export const runLabel = (model: string, effort?: string | number): string => {
  const family = /opus|sonnet|haiku|fable/.exec(model)?.[0] ?? model
  return effort === undefined ? family : `${family}-${effort}`
}

// "🔄 1  🔍 2  ✅ 4": the board's rows counted by their kind's emoji (the stage emoji's vocabulary), running kinds first. The space
// after each emoji keeps the count clear of it where the terminal draws the emoji wider than it measures.
export const tally = (rs: Row[]): string => {
  const order = [EMOJI['in progress'], ...KINDS.map(([, e]) => e), EMOJI.done, EMOJI.blocked]
  return order
    .map(e => [e, rs.filter(r => look(r).emoji === e).length] as const)
    .filter(([, n]) => n > 0)
    .map(([e, n]) => `${e} ${n}`)
    .join('  ')
}

// Past MAX_ROWS, the oldest finished rows go first; running rows are never dropped.
export const trim = (rs: Row[]): Row[] => {
  const out = [...rs]
  while (out.length > MAX_ROWS) {
    const i = out.findIndex(r => r.state !== 'in progress')
    if (i < 0) break
    out.splice(i, 1)
  }
  return out
}

// Each flow.* stage key's phase, per skills/flow/stage-keys.md. `flow.decide` is left out: it marks
// both the brainstorm and a fix run, so it keeps whichever phase is current.
const PHASE_KEYS: Record<Phase, string[]> = {
  'flow-plan': ['kickoff', 'brainstorm', 'design-approval', 'create-artifacts', 'writing-plans'],
  'flow-implement': [
    'load-context', 'isolate-workspace', 'sdd-tdd', 'document-fix', 'review-panel', 'verify',
    'visual-verify', 'stage-diff', 'run-instructions', 'write-in-progress',
  ],
  'flow-integrate': [
    'preflight', 'unfinished-work-gate', 'landing-question', 'preserve-sessions', 'commit-two',
    'landing-routes', 'verify-merge', 'sync-archive', 'commit-archive', 'cleanup', 'verify-cleanup',
    'write-finished', 'self-review', 'refresh-main-checkout',
  ],
}
// The stage each run of /flow stops after: its end leaves no phase running.
const LAST_KEYS = ['writing-plans', 'write-in-progress', 'landing-routes', 'refresh-main-checkout']

// One `flow stage begin|end ... <name>` mark, up to the next shell separator.
const STAGE_MARK = /\bflow stage (begin|end)\b([^;&|\n]*)/g
// A change name's leading tracker key: "kan-873-port-guards" → "KAN-873".
const TICKET = /^['"]?([a-z][a-z0-9]*-\d+)(?=-|['"]?$)/i

// The stages whose running gets an emoji on the hint line, after the phase.
const STAGE_EMOJI: Record<string, string> = {
  'document-fix': '🔨',
  'review-panel': '🔍',
  'visual-verify': '👀',
  verify: '🧪',
}

// The flow state after a Bash command's `flow stage` marks; `current` when it carries none that move it.
export const flowAfter = (command: string, current: Flow): Flow => {
  let f = current
  for (const [, verb, rest = ''] of command.matchAll(STAGE_MARK)) {
    const key = /-stage\s+flow\.([\w-]+)/.exec(rest)?.[1]
    if (!key) continue
    if (verb === 'end') {
      if (LAST_KEYS.includes(key)) f = NO_FLOW
      else if (f.stage === key) f = { ...f, stage: null }
      continue
    }
    const phase = (Object.keys(PHASE_KEYS) as Phase[]).find(ph => PHASE_KEYS[ph].includes(key)) ?? f.phase
    const name = rest.trim().split(/\s+/).at(-1) ?? ''
    const ticket = TICKET.exec(name)?.[1]?.toUpperCase() ?? f.ticket
    f = phase ? { phase, change: name.replace(/^['"]|['"]$/g, '') || f.change, ticket, stage: key } : NO_FLOW
  }
  return f
}

// "· KAN-873 flow-implement 🔍 -> ✅ 5", each part only where it exists.
export const hintTail = (f: Flow, t: string): string | undefined => {
  const stage = f.stage ? STAGE_EMOJI[f.stage] : undefined
  const head = f.phase ? [f.ticket, f.phase, stage].filter(Boolean).join(' ') : ''
  return head && t ? `· ${head} -> ${t}` : head || t ? `· ${head || t}` : undefined
}

export const register: Register = on => {
  on('agent.spawn', async ($, e, next) => {
    const result = await next(e)
    const id = result.deny === undefined ? result.agentId : undefined
    if (id) {
      const kept = await update($, rows, (rs): Row[] => {
        const n = (rs.at(-1)?.n ?? 0) + 1
        return trim([...rs, { id, n, desc: e.description, state: 'in progress' }])
      })
      // Drop the labels of rows trimmed away; keep the new agent's, which its first step may have set already.
      await update($, runs, rs => Object.fromEntries(Object.entries(rs).filter(([k]) => kept.some(r => r.id === k))))
    }
    return result
  })

  // A subagent's first model request names the model and effort it actually runs on. A main-loop step
  // means the main turn runs: its first starts the main row afresh, and any later step starts it when it is
  // missing, as when the plugin loads mid-turn or a main-loop step arrives after the main turn.complete.
  on('turn.step', async function* ($, e, next) {
    const id = e.agentId
    const label = runLabel(e.model, e.effort)
    if (id && e.index === 0) {
      await update($, runs, rs => ({ ...rs, [id]: label }))
    } else if (!id && (e.index === 0 || !(await read($, main)))) {
      await update($, main, () => ({ desc: '', run: label }))
    }
    return yield* next(e)
  })

  on('turn.complete', async ($, e, next) => {
    const id = e.agentId
    if (id) {
      const state: RowState = e.reason === 'answer' ? 'done' : 'blocked'
      await update($, rows, rs => rs.map(r => (r.id === id ? { ...r, state } : r)))
    } else {
      await update($, main, () => null)
    }
    return next(e)
  })

  // A main-loop command's description is what the main row says it is doing, set before the command
  // runs so the row names it while it runs. A denied command never ran, so its marks move nothing.
  on('tool.call', { tool: 'Bash' }, async ($, e, next) => {
    const desc = e.description
    if (!e.agentId && desc) {
      await update($, main, m => (m ? { ...m, desc } : m))
    }
    const result = await next(e)
    if (result.deny === undefined) {
      const current = await read($, flow)
      const after = flowAfter(e.command, current)
      if (after !== current) {
        await update($, flow, () => after)
      }
    }
    return result
  })

  on('ui.render', { component: 'PromptHint' }, async ($, e, next) => {
    const tail = hintTail(await read($, flow), tally(await read($, rows)))
    return next(tail ? { ...e, props: { ...e.props, tail } } : e)
  })

  on('ui.render', { component: 'AbovePrompt' }, async ($, e, next) => {
    const subs = await read($, rows)
    const m = await read($, main)
    // The main agent's row, first and only while its turn runs; it stays out of rows, so trim and the
    // task numbering never see it.
    const rs: Row[] = m ? [{ id: 'main', n: 0, desc: m.desc, state: 'in progress' }, ...subs] : subs
    // Shown while the main turn or a subagent runs; once all have finished the band hides and the hint line's tally stays.
    if (e.props.hasSurvey || !rs.some(r => r.state === 'in progress')) {
      return next(e)
    }
    // The running change's plan, read from its worktree (`<project>/.worktrees/<name>`, where /flow ticks it);
    // absent, no pending rows.
    const { change } = await read($, flow)
    const plan = change
      ? await $.fs.read(`.worktrees/${change}/spectre/changes/${change}/tasks.md`).catch(() => '')
      : ''
    const all = [...rs, ...pendingRows(plan, subs)]
    const labels = m ? { ...(await read($, runs)), main: m.run } : await read($, runs)
    const width = Math.max(0, ...all.map(r => labels[r.id]?.length ?? 0))
    const total = subs.at(-1)?.n ?? 0
    const { Box, Text } = $.ui.resolve(e)
    return (
      <Box flexDirection="column">
        {all.map((r, i) => {
          const p = parts(r, total, labels[r.id], width, i === 0)
          const isDone = r.state === 'done'
          return (
            <Text key={r.id}>
              <Text dimColor>{`${p.lead}${p.run ? `${p.run} ` : ''}`}</Text>
              <Text color={p.color}>{`${p.mark} `}</Text>
              <Text bold={r.state === 'in progress'} dimColor={isDone} strikethrough={isDone}>{`${p.unit}${p.desc}`}</Text>
              <Text dimColor>{' — '}</Text>
              <Text color={r.state === 'blocked' ? p.color : undefined} dimColor={isDone}>{p.state}</Text>
            </Text>
          )
        })}
      </Box>
    )
  })
}
