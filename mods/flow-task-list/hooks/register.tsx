import { atom, read, update } from 'claude-code'
import type { Hook, Register, Timer } from 'claude-code'

import type { Flow, Main, Phase, Row, RowState, StatusLine } from '../types'

import { landedIn } from './landed'

const rows = atom({ plugin: 'flow-task-list', key: 'rows' } as const, [] as Row[])
const runs = atom({ plugin: 'flow-task-list', key: 'runs' } as const, {} as Record<string, string>)
const NO_FLOW: Flow = { phase: null, change: null, ticket: null }
const flow = atom({ plugin: 'flow-task-list', key: 'flow' } as const, NO_FLOW)
const main = atom({ plugin: 'flow-task-list', key: 'main' } as const, null as Main | null)
const lines = atom({ plugin: 'flow-task-list', key: 'lines' } as const, {} as Record<string, StatusLine>)

const MAX_ROWS = 5

const EMOJI: Record<RowState, string> = { 'in progress': '🔄', 'in review': '🔍', done: '✅', blocked: '⛔', pending: '⏳' }
// The task-list look: its marker and colour per state; `claude` is the theme's accent (orange). A pending
// marker keeps the text's own colour.
const MARK: Record<RowState, string> = { 'in progress': '◼', 'in review': '◼', done: '✔', blocked: '✘', pending: '◻' }
const MARK_COLOR: Record<RowState, string | undefined> = {
  'in progress': 'claude',
  'in review': 'claude',
  done: 'success',
  blocked: 'error',
  pending: undefined,
}

// A description that already carries its plan numbering keeps it: one task, "Task 3/22 (spec text)",
// or a /flow task group, "Tasks 3+4+7/22 (port guards)" (skills/flow/implement.md, **Dispatch sites**), or a
// status line's range, "Tasks 23–25/25 (e2e)".
const NUMBERED = /^(Tasks? \d+(?:[+–-]\d+)*\/\d+)\s*/

// The plan task numbers a "Task(s) …/n" prefix names: "Tasks 3+4/22" → 3, 4; "Tasks 23–25/25" → 23, 24, 25.
export const taskNums = (text: string): number[] =>
  (NUMBERED.exec(text)?.[1]?.match(/\d+(?:[–-]\d+)?(?=[+/])/g) ?? []).flatMap(p => {
    const [a = 0, b = a] = p.split(/[–-]/).map(Number)
    return Array.from({ length: Math.max(0, b - a + 1) }, (_, i) => a + i)
  })

// "4m12s": a row's elapsed time.
const clockText = (ms: number): string => {
  const s = Math.floor(ms / 1000)
  const m = Math.floor(s / 60)
  const pad = (k: number) => String(k).padStart(2, '0')
  return s < 60 ? `${s}s` : m < 60 ? `${m}m${pad(s % 60)}s` : `${Math.floor(m / 60)}h${pad(m % 60)}m`
}

// " · 4m12s" for a subagent row: elapsed until `now` while it runs, frozen at its end; empty for the rows no
// spawn started (main, pending, status lines).
const stats = (row: Row, now: number): string => (row.start === undefined ? '' : ` · ${clockText((row.end ?? now) - row.start)}`)

// What a running row is doing, read from its description, and the state word it shows; the kind whose word
// comes first in the description wins, so "Review … fix" is a review and "Fix review findings" a fix. Each
// word is matched from its start, so "prefix", "preview" or "latest" names no kind.
const KINDS: [RegExp, string, string][] = [
  [/\bfix/i, '🔨', 'fix'],
  [/\bvisual[- ]verif/i, '👀', 'visual verify'],
  [/\b(?:review|panel(?!-fix))/i, '🔍', 'in review'],
  [/\b(?:verif|tests?\b|lint)/i, '🧪', 'verify'],
]

// How a row shows: a running row by its kind's word (in review, fix, visual verify, verify), a finished one by its state.
// `emoji` is what the hint line's tally counts it under; `mark` and `color` are its marker on the band.
export const look = (row: Row): { emoji: string; word: string; mark: string; color: string | undefined } => {
  const at = (re: RegExp) => row.desc.search(re)
  const kind =
    row.state === 'in progress'
      ? KINDS.filter(([re]) => at(re) >= 0).sort(([a], [b]) => at(a) - at(b))[0]
      : undefined
  return {
    emoji: kind?.[1] ?? EMOJI[row.state],
    word: kind?.[2] ?? row.state,
    mark: MARK[row.state],
    color: MARK_COLOR[row.state],
  }
}

// One board row's pieces; `lead` is "⎿ " on the first row and its width in spaces after, or, for a row
// `nested` under the main agent's row, "  ⎿ " on each; `run` is padded to `width` so the marker column
// lines up. A row keeps its "Task(s) …/n" prefix only when every number it names is a task of the running
// plan, `plan` (its task numbers), and n is that plan's size; a row numbered for another plan keeps its text
// as written, so "Task 26/26 review" never shrinks to "review"; the main agent's row is "main".
export const parts = (row: Row, plan: number[], run = '', width = 0, first = true, now = 0, nested = false) => {
  const text = row.unit ?? row.desc
  const m = NUMBERED.exec(text)
  const nums = taskNums(text)
  const ours = !!m && nums.length > 0 && nums.every(k => plan.includes(k)) && Number(m[1]?.split('/')[1]) === plan.length
  const rest = (m ? text.slice(m[0].length) : text).replace(/^\((.*)\)$/, '$1')
  const l = look(row)
  const [unit, desc] = row.id === 'main' ? ['main', row.desc] : ours ? [m[1] ?? '', rest] : [text, '']
  return {
    lead: nested ? '  ⎿ ' : first ? '⎿ ' : '  ',
    run: run.padEnd(width),
    mark: l.mark,
    unit,
    desc: desc ? ` (${desc})` : '',
    state: l.word,
    color: l.color,
    stats: stats(row, now),
  }
}

export const line = (row: Row, plan: number[], run = '', width = 0, first = true, now = 0, nested = false): string => {
  const p = parts(row, plan, run, width, first, now, nested)
  return `${p.lead}${p.run ? `${p.run} ` : ''}${p.mark} ${p.unit}${p.desc} — ${p.state}${p.stats}`
}

// A be-brief status line, "<emoji> <unit> — <state>" (rules/be-brief.mdc), at the start of its line; text after
// the state word (". Two new failing specs.", " again") is read past, a word running on from it ("fixing") is not.
const STATUS = /^[✅🔄🔍⏳⛔]\uFE0F? (.+?) — (done|in progress|in review|pending|blocked)(?=$|[., ])/gmu

// The unit's leading name, its key in `lines`: the text before any "(", ":" or " —", without a trailing
// " review", so "Visual verify (final): Quick mode" and "Visual verify (final full run)" are one unit, and
// "Flow pipeline change review" and "Flow pipeline change (narrow re-verify)" another.
export const unitKey = (unit: string): string =>
  (unit.split(/\(|:| —/)[0] ?? '').trim().replace(/ review$/i, '') || unit

// Each status line of a response, in order, as its unit and state.
export const statusLines = (text: string): [string, RowState][] =>
  [...text.matchAll(STATUS)].map(([, unit = '', state]) => [unit, state as RowState])

// The main loop's open status lines as rows, after the subagent rows; a unit whose task numbers a subagent row
// already names is left to that row.
export const lineRows = (ls: Record<string, StatusLine>, subs: Row[]): Row[] => {
  const taken = new Set(subs.flatMap(r => taskNums(r.desc)))
  return Object.entries(ls)
    .filter(([, l]) => !taskNums(l.unit).some(k => taken.has(k)))
    .map(([key, l]) => ({ id: `line:${key}`, n: 0, desc: '', unit: l.unit, state: l.state }))
}

// A plan's column-0 task line, "- [ ] 23. Title" or "- [x] 23. Title"; its step checkboxes are indented.
const TASK = /^- \[([ x])\] (\d+)\. (.*)$/gm

// The plan's task numbers, in order.
export const planTasks = (tasksMd: string): number[] => [...tasksMd.matchAll(TASK)].map(([, , k]) => Number(k))

// "19/25": the plan's done tasks over all its tasks, derived from its tasks.md on every draw; empty without tasks.
// A task is done when it is ticked or its implementation commit has `landed` on the change branch: a gated
// task's tick waits on its reviewer (skills/flow/implement.md), so ticks alone lag the work by a review round.
export const taskCount = (tasksMd: string, done: number[] = []): string => {
  const tasks = [...tasksMd.matchAll(TASK)]
  const n = tasks.filter(([, box, k]) => box === 'x' || done.includes(Number(k))).length
  return tasks.length ? `${n}/${tasks.length}` : ''
}

// The plan's tasks neither ticked nor `done` (landed) that no board row names in its "Task(s) …/n" numbering,
// as pending rows: what the running /flow change has not dispatched yet, derived from its tasks.md on every draw.
export const pendingRows = (tasksMd: string, rs: Row[], done: number[] = []): Row[] => {
  const tasks = [...tasksMd.matchAll(TASK)]
  const taken = new Set(rs.flatMap(r => taskNums(r.unit ?? r.desc)))
  return tasks
    .filter(([, box, n = '']) => box === ' ' && !done.includes(Number(n)) && !taken.has(Number(n)))
    .map(([, , n = '', title = '']) => ({ id: `pending-${n}`, n: Number(n), desc: `Task ${n}/${tasks.length} (${title})`, state: 'pending' }))
}

// "opus-high" from the model id and effort a request went out with.
export const runLabel = (model: string, effort?: string | number): string => {
  const family = /opus|sonnet|haiku|fable/.exec(model)?.[0] ?? model
  return effort === undefined ? family : `${family}-${effort}`
}

// "🔄 1  🔍 2  ✅ 4": the board's rows counted by their kind's emoji, running kinds first. The space
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

// The band's rows besides main, MAX_ROWS at most: the `shown` rows (subagent rows, then status-line rows) in order,
// then the `pending` rows; while a pending row would be crowded out, the earliest done row yields its slot to it.
export const band = (shown: Row[], pending: Row[]): Row[] => {
  const out = [...shown]
  while (out.length + pending.length > MAX_ROWS) {
    const i = out.findIndex(r => r.state === 'done')
    if (i < 0) break
    out.splice(i, 1)
  }
  return [...out, ...pending].slice(0, MAX_ROWS)
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
const TICKET = /^([a-z][a-z0-9]*-\d+)(?=-|$)/i
// A plain `VAR=value` assignment, its value bare or quoted.
const ASSIGN = /(?:^|[\s;&|])([A-Za-z_]\w*)=('[^']*'|"[^"]*"|[^\s;&|'"]*)/g
const unquote = (s: string): string => s.replace(/^(['"])(.*)\1$/, '$2')

// The change name a mark ends with, its `$VAR` and `${VAR}` resolved from the plain assignments before it
// in the same command; null when what is left is not a plain change name.
const changeName = (word: string, before: string): string | null => {
  const vars = Object.fromEntries([...before.matchAll(ASSIGN)].map(([, k = '', v = '']) => [k, unquote(v)]))
  const name = unquote(word).replace(/\$(?:\{(\w+)\}|(\w+))/g, (ref, a, b) => vars[a ?? b] ?? ref)
  return /^[\w.-]+$/.test(name) ? name : null
}

// The flow state after a Bash command's `flow stage` marks; `current` when it carries none that move it.
export const flowAfter = (command: string, current: Flow): Flow => {
  let f = current
  for (const { 1: verb, 2: rest = '', index } of command.matchAll(STAGE_MARK)) {
    const key = /-stage\s+flow\.([\w-]+)/.exec(rest)?.[1]
    if (!key) continue
    if (verb === 'end') {
      if (LAST_KEYS.includes(key)) f = NO_FLOW
      continue
    }
    const phase = (Object.keys(PHASE_KEYS) as Phase[]).find(ph => PHASE_KEYS[ph].includes(key)) ?? f.phase
    // A name that is no plain change name (an unset `$N`, a stray word) never replaces the known change.
    const name = changeName(rest.trim().split(/\s+/).at(-1) ?? '', command.slice(0, index))
    const ticket = name ? (TICKET.exec(name)?.[1]?.toUpperCase() ?? f.ticket) : f.ticket
    f = phase ? { phase, change: name ?? f.change, ticket } : NO_FLOW
  }
  return f
}

// "KAN-873 flow-implement 19/25 │ 🔍 2  ✅ 5": the main agent's part — tracker key, phase, the plan's task count
// — " │ ", then the subagent rows' tally, each part only where it exists. No leading separator: the engine
// puts " · " between its own line and the tail. The running stage gets no emoji of its own: the tally
// already counts the rows running it under that emoji, and a second copy read as two of them.
export const hintTail = (f: Flow, count: string, t: string): string | undefined => {
  const head = f.phase ? [f.ticket, f.phase, count].filter(Boolean).join(' ') : ''
  return [head, t].filter(Boolean).join(' │ ') || undefined
}

// The running change's plan, read from its worktree (`<dirname project>/<basename project>-worktrees/<name>`,
// beside the main checkout, where /flow ticks it); absent, ''. `session.root()` is the launch directory —
// `session.cwd()` follows the shell, so a `cd stats && …` would aim it at `stats-worktrees/`.
const planOf = async ($: Parameters<Hook<'ui.render'>>[0], change: string | null): Promise<string> =>
  change ? $.fs.read(`${await $.session.root()}-worktrees/${change}/spectre/changes/${change}/tasks.md`).catch(() => '') : ''

const landedOf = async ($: Parameters<Hook<'ui.render'>>[0], change: string | null): Promise<number[]> =>
  change ? landedIn(argv => $.process.run(argv), `${await $.session.root()}-worktrees/${change}`) : []

export const register: Register = on => {
  // Redraws the band each second while a subagent row runs, so its elapsed time ticks.
  let tick: Timer | undefined

  on('agent.spawn', async ($, e, next) => {
    const result = await next(e)
    const id = result.deny === undefined ? result.agentId : undefined
    if (id) {
      const start = await $.clock.now()
      const kept = await update($, rows, (rs): Row[] => {
        const n = (rs.at(-1)?.n ?? 0) + 1
        return trim([...rs, { id, n, desc: e.description, state: 'in progress', start }])
      })
      // Drop the labels of rows trimmed away; keep the new agent's, which its first step may have set already.
      await update($, runs, rs => Object.fromEntries(Object.entries(rs).filter(([k]) => kept.some(r => r.id === k))))
    }
    return result
  })

  // A subagent's first model request names the model and effort it actually runs on. A main-loop step
  // means the main turn runs: its first starts the main row afresh, and any later step starts it when it is
  // missing, as when the plugin loads mid-turn or a main-loop step arrives after the main turn.complete.
  // A main-loop response's status lines update `lines`; a blocked line's row lasts until the next main turn starts.
  on('turn.step', async function* ($, e, next) {
    const id = e.agentId
    const label = runLabel(e.model, e.effort)
    if (id && e.index === 0) {
      await update($, runs, rs => ({ ...rs, [id]: label }))
    } else if (!id && (e.index === 0 || !(await read($, main)))) {
      await update($, main, () => ({ desc: '', run: label }))
      if (e.index === 0) {
        await update($, lines, ls => Object.fromEntries(Object.entries(ls).filter(([, l]) => l.state !== 'blocked')))
      }
    }
    const result = yield* next(e)
    const said = id ? [] : statusLines(result.answer)
    if (said.length) {
      // ponytail: a unit left in progress, in review or pending whose closing line never comes stays until the
      // session ends; it shows only while the band does.
      await update($, lines, ls => {
        const out = { ...ls }
        for (const [unit, state] of said) {
          const key = unitKey(unit)
          if (state === 'done') delete out[key]
          else out[key] = { unit, state }
        }
        return out
      })
    }
    return result
  })

  on('turn.complete', async ($, e, next) => {
    const id = e.agentId
    if (id) {
      const state: RowState = e.reason === 'answer' ? 'done' : 'blocked'
      const end = await $.clock.now()
      await update($, rows, rs => rs.map(r => (r.id === id ? { ...r, state, end } : r)))
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
    const f = await read($, flow)
    const tail = hintTail(f, taskCount(await planOf($, f.change), await landedOf($, f.change)), tally(await read($, rows)))
    return next(tail ? { ...e, props: { ...e.props, tail } } : e)
  })

  on('ui.render', { component: 'AbovePrompt' }, async ($, e, next) => {
    const subs = await read($, rows)
    const m = await read($, main)
    // The main agent's row, first and only while its turn runs; it stays out of rows, so trim and the
    // task numbering never see it.
    const rs: Row[] = m ? [{ id: 'main', n: 0, desc: m.desc, state: 'in progress' }, ...subs] : subs
    const ticking = subs.some(r => r.state === 'in progress')
    if (ticking && !tick) {
      tick = $.clock.every(1000, () => $.ui.invalidate('ui.render'))
    } else if (!ticking && tick) {
      tick.cancel()
      tick = undefined
    }
    // Shown while the main turn or a subagent runs; once all have finished the band hides and the hint line's tally stays.
    if (e.props.hasSurvey || !rs.some(r => r.state === 'in progress')) {
      return next(e)
    }
    const change = (await read($, flow)).change
    const plan = await planOf($, change)
    // The main row plus at most MAX_ROWS others: the board's rows first, then the main loop's open status lines,
    // then the earliest pending tasks no row above names; past MAX_ROWS running rows, only the earliest show,
    // while the hint line's tally still counts them all.
    const said = lineRows(await read($, lines), subs)
    const all = [
      ...rs.filter(r => r.id === 'main'),
      ...band([...subs, ...said], pendingRows(plan, [...subs, ...said], await landedOf($, change))),
    ]
    const labels = m ? { ...(await read($, runs)), main: m.run } : await read($, runs)
    // The marker column lines up among the rows sharing a lead: with the main row shown, the rows nested under
    // it; the main row's label is not padded.
    const width = Math.max(0, ...all.filter(r => r.id !== 'main').map(r => labels[r.id]?.length ?? 0))
    const tasks = planTasks(plan)
    const now = await $.clock.now()
    const { Box, Text } = $.ui.resolve(e)
    return (
      <Box flexDirection="column">
        {all.map((r, i) => {
          // While the main row shows, every row after it is nested one level under it.
          const p = parts(r, tasks, labels[r.id], r.id === 'main' ? 0 : width, i === 0, now, m !== null && i > 0)
          const isDone = r.state === 'done'
          return (
            <Text key={r.id}>
              <Text dimColor>{`${p.lead}${p.run ? `${p.run} ` : ''}`}</Text>
              <Text color={p.color}>{`${p.mark} `}</Text>
              <Text bold={r.state === 'in progress'} dimColor={isDone} strikethrough={isDone}>{`${p.unit}${p.desc}`}</Text>
              <Text dimColor>{' — '}</Text>
              <Text color={r.state === 'blocked' ? p.color : undefined} dimColor={isDone}>{p.state}</Text>
              <Text dimColor>{p.stats}</Text>
            </Text>
          )
        })}
      </Box>
    )
  })
}
