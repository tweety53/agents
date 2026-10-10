import { atom, read, update } from 'claude-code'
import type { Hook, Register, Timer } from 'claude-code'

import type { Flow, Main, Phase, Row, RowState, StatusLine } from '../types'

import { landedIn } from './landed'

const rows = atom({ plugin: 'flow-task-list', key: 'rows' } as const, [] as Row[])
const runs = atom({ plugin: 'flow-task-list', key: 'runs' } as const, {} as Record<string, string>)
const NO_FLOW: Flow = { phase: null, change: null, ticket: null, stage: null }
const flow = atom({ plugin: 'flow-task-list', key: 'flow' } as const, NO_FLOW)
const main = atom({ plugin: 'flow-task-list', key: 'main' } as const, null as Main | null)
const lines = atom({ plugin: 'flow-task-list', key: 'lines' } as const, {} as Record<string, StatusLine>)

// The rows the band draws under main, at most, and the subagent rows `rows` keeps past its running ones.
const MAX_ROWS = 5
// The finished rows the band draws when no pending task is left.
const DONE_ROWS = 2

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

// Whether a "Task(s) …/n" prefix numbers tasks of the running plan, `plan` (its task numbers): every number it
// names is one of them and n is the plan's size.
const ofPlan = (text: string, plan: number[]): boolean => {
  const nums = taskNums(text)
  return nums.length > 0 && nums.every(k => plan.includes(k)) && Number(NUMBERED.exec(text)?.[1]?.split('/')[1]) === plan.length
}

// "4m12s": a row's elapsed time.
const clockText = (ms: number): string => {
  const s = Math.floor(ms / 1000)
  const m = Math.floor(s / 60)
  const pad = (k: number) => String(k).padStart(2, '0')
  return s < 60 ? `${s}s` : m < 60 ? `${m}m${pad(s % 60)}s` : `${Math.floor(m / 60)}h${pad(m % 60)}m`
}

// " · 4m12s" for a subagent row: elapsed until `now` while it runs, frozen at its end; empty for the rows no
// spawn started (main, pending).
const stats = (row: Row, now: number): string => (row.start === undefined ? '' : ` · ${clockText((row.end ?? now) - row.start)}`)

// A row's words after its "Task(s) …/n" prefix, without the parentheses around them.
const rest = (desc: string, prefix: string): string => desc.slice(prefix.length).trim().replace(/^\((.*)\)$/, '$1')

// A review key's round: "review-2", "panel-2-primary", "visual-verify-2"; "panel-fix-1" names none.
const ROUND = /\b(?:review|panel|visual[- ]verify)-(\d+)/i
// A review or verify key re-run after a fix: "visual-verify-fix-1", "visual-verify-fix-1-full",
// "task-3+4-reviewer-fix-1".
const RE_RUN = /\b(?:reviewer|review|verify)-fix-\d+/i
const FIX = /\bfix/i
const REVIEW = /\b(?:review|panel(?![- ]fix)|visual[- ]verif)/i
// The tally's emoji per kind.
const KIND_EMOJI = { fix: '🔨', review: '🔍', 're-review': '🔍' } as const
type Kind = keyof typeof KIND_EMOJI

// What a running subagent does, read from its description. A row with a "Task(s) …/n" prefix is a review or
// re-review only when its words are exactly that — the gated reviewer's "Tasks 3+4+7/22 (review)" — and an
// implementer otherwise, whatever its title says. Any other row: "re-review", a review key past round 1 or one
// re-run after a fix is a re-review, round 1 a review; otherwise the kind whose word comes first, so "Review … fix" is a review and
// "Fix review findings" a fix; neither, an implementer, shown as in progress. Each word is matched from its
// start, so "prefix" or "preview" names no kind.
export const kind = (desc: string): Kind | undefined => {
  const m = NUMBERED.exec(desc)
  if (m) {
    const words = rest(desc, m[0]).toLowerCase()
    return words === 'review' || words === 're-review' ? words : undefined
  }
  const round = Number(ROUND.exec(desc)?.[1] ?? 0)
  if (round > 1 || RE_RUN.test(desc) || /\bre-?review/i.test(desc)) return 're-review'
  if (round === 1) return 'review'
  const f = desc.search(FIX)
  const r = desc.search(REVIEW)
  return f < 0 && r < 0 ? undefined : r < 0 || (f >= 0 && f < r) ? 'fix' : 'review'
}

// How a row shows: a running subagent row by its kind's word (fix, review, re-review), any other by its state.
// `emoji` is what the hint line's tally counts it under; `mark` and `color` are its marker on the band.
export const look = (row: Row): { emoji: string; word: string; mark: string; color: string | undefined } => {
  const k = row.state === 'in progress' ? kind(row.desc) : undefined
  return { emoji: k ? KIND_EMOJI[k] : EMOJI[row.state], word: k ?? row.state, mark: MARK[row.state], color: MARK_COLOR[row.state] }
}

// One band row's pieces. The main row, id `main`, leads with "⎿ ", reads "main: <its label>", and is in progress
// while its turn runs, waiting between turns; every other row is nested under it as "  ⎿ ". `run` is padded to
// `width` so the marker column lines up. A row naming tasks of the running plan (`ofPlan`) shows its prefix as
// "Task …/n" — a group "Task 1+2+3/10"; a done row with a "Task(s) …/n" prefix shows only its highest task over
// its own n, "Task 3/10", plan or none; any other row keeps its text as written, so "Task 26/26 review" never
// shrinks to "review".
export const parts = (row: Row, plan: number[], run = '', width = 0, now = 0) => {
  const isMain = row.id === 'main'
  const m = NUMBERED.exec(row.desc)
  const numbered = (prefix: string, head: string) => {
    const words = rest(row.desc, prefix)
    return words ? `${head} (${words})` : head
  }
  const text = isMain
    ? row.desc === 'main'
      ? 'main'
      : `main: ${row.desc}`
    : m && row.state === 'done'
      ? numbered(m[0], `Task ${Math.max(...taskNums(row.desc))}/${m[1]?.split('/')[1]}`)
      : m && ofPlan(row.desc, plan)
        ? numbered(m[0], (m[1] ?? '').replace(/^Tasks/, 'Task'))
        : row.desc
  const l = look(isMain ? { ...row, desc: '' } : row)
  return {
    lead: isMain ? '⎿ ' : '  ⎿ ',
    run: run.padEnd(width),
    mark: l.mark,
    text,
    state: isMain && row.state === 'pending' ? 'waiting' : l.word,
    color: l.color,
    stats: stats(row, now),
  }
}

export const line = (row: Row, plan: number[], run = '', width = 0, now = 0): string => {
  const p = parts(row, plan, run, width, now)
  return `${p.lead}${p.run ? `${p.run} ` : ''}${p.mark} ${p.text} — ${p.state}${p.stats}`
}

// A be-brief status line, "<emoji> <unit> — <state>" (rules/be-brief.mdc), at the start of its line; text after
// the state word (". Two new failing specs.", " again") is read past, a word running on from it ("fixing") is not.
const STATUS = /^[✅🔄🔍⏳⛔]\uFE0F? (.+) — (done|in progress|in review|pending|blocked)(?![\p{L}\p{N}])/gmu

// The unit's leading name, its key in `lines`: the text before any "(", ":" or " —", without a trailing
// " review", so "Visual verify (final): Quick mode" and "Visual verify (final full run)" are one unit, and
// "Flow pipeline change review" and "Flow pipeline change (narrow re-verify)" another.
export const unitKey = (unit: string): string =>
  (unit.split(/\(|:| —/)[0] ?? '').trim().replace(/ review$/i, '') || unit

// Each status line of a response, in order, as its unit and state.
export const statusLines = (text: string): [string, RowState][] =>
  [...text.matchAll(STATUS)].map(([, unit = '', state]) => [unit, state as RowState])

// What the main agent is on, never a subtask: the latest open status-line unit of its own naming a task of the
// running plan (`plan`), the task it runs inline; else the running /flow `stage`; else its latest open
// status-line unit of its own; else "main". A unit is its own when no subagent row in `subs` shares its key or
// names one of its task numbers — a line about a subagent's work is that row's, never the main agent's.
export const mainLabel = (ls: Record<string, StatusLine>, subs: Row[], stage: string | null, plan: number[]): string => {
  const keys = new Set(subs.map(r => unitKey(r.desc)))
  const taken = new Set(subs.flatMap(r => taskNums(r.desc)))
  const own = Object.entries(ls)
    .filter(([key, l]) => !keys.has(key) && !taskNums(l.unit).some(k => taken.has(k)))
    .map(([, l]) => l.unit)
    .reverse()
  return own.find(u => ofPlan(u, plan)) ?? stage ?? own[0] ?? 'main'
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

// The plan's tasks neither ticked nor `done` (landed) that no row in `rs` (the running rows and the main row)
// names in its "Task(s) …/n" numbering, as pending rows: what the running /flow change has not started, derived
// from its tasks.md on every draw.
export const pendingRows = (tasksMd: string, rs: Row[], done: number[] = []): Row[] => {
  const tasks = [...tasksMd.matchAll(TASK)]
  const taken = new Set(rs.flatMap(r => taskNums(r.desc)))
  return tasks
    .filter(([, box, n = '']) => box === ' ' && !done.includes(Number(n)) && !taken.has(Number(n)))
    .map(([, , n = '', title = '']) => ({ id: `pending-${n}`, n: Number(n), desc: `Task ${n}/${tasks.length} (${title})`, state: 'pending' }))
}

// "opus-high" from the model id and effort a request went out with.
export const runLabel = (model: string, effort?: string | number): string => {
  const family = /opus|sonnet|haiku|fable/.exec(model)?.[0] ?? model
  return effort === undefined ? family : `${family}-${effort}`
}

// "🔄 1  🔍 2  ✅ 4": the subagent rows counted by their kind's emoji, running kinds first. The space
// after each emoji keeps the count clear of it where the terminal draws the emoji wider than it measures.
export const tally = (rs: Row[]): string => {
  const order = [EMOJI['in progress'], KIND_EMOJI.fix, KIND_EMOJI.review, EMOJI.done, EMOJI.blocked]
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

// The running subagent rows and those that failed since the main turn began at `since`, in spawn order.
export const live = (subs: Row[], since = 0): Row[] =>
  subs.filter(r => r.state === 'in progress' || (r.state === 'blocked' && (r.end ?? 0) >= since))

// The band's rows under main, MAX_ROWS at most: the `live` rows, then the `pending` rows below them; with no
// pending row left, the last DONE_ROWS done rows of `change` — with none marked, those done since `since` — in
// the order they finished, above them, in room the live rows leave.
export const band = (subs: Row[], pending: Row[], change: string | null = null, since = 0): Row[] => {
  const shown = live(subs, since)
  const room = Math.max(0, Math.min(DONE_ROWS, MAX_ROWS - shown.length))
  const done =
    pending.length || !room
      ? []
      : subs
          .filter(r => r.state === 'done' && (change ? r.change === change : (r.end ?? 0) >= since))
          .sort((a, b) => (a.end ?? 0) - (b.end ?? 0))
          .slice(-room)
  return [...done, ...shown, ...pending].slice(0, MAX_ROWS)
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
    f = phase ? { phase, change: name ?? f.change, ticket, stage: key } : NO_FLOW
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
  // Redraws the band each second while a subagent row runs, so its elapsed time ticks, and while the main turn runs a
  // /flow change, so a task it ticks or lands inline leaves the pending rows.
  let tick: Timer | undefined

  on('agent.spawn', async ($, e, next) => {
    const result = await next(e)
    const id = result.deny === undefined ? result.agentId : undefined
    if (id) {
      const start = await $.clock.now()
      const { change } = await read($, flow)
      const kept = await update($, rows, (rs): Row[] => {
        const n = (rs.at(-1)?.n ?? 0) + 1
        return trim([...rs, { id, n, desc: e.description, state: 'in progress', start, change }])
      })
      // Drop the labels of rows trimmed away; keep the new agent's, which its first step may have set already.
      await update($, runs, rs => Object.fromEntries(Object.entries(rs).filter(([k]) => kept.some(r => r.id === k))))
    }
    return result
  })

  // A subagent's first model request names the model and effort it actually runs on. A main-loop step
  // means the main turn runs: its first marks the main agent busy, and any later step does when it is not,
  // as when the plugin loads mid-turn or a main-loop step arrives after the main turn.complete.
  // A main-loop response's status lines update `lines`, the latest written last; a blocked line lasts until
  // the next main turn starts.
  on('turn.step', async function* ($, e, next) {
    const id = e.agentId
    const label = runLabel(e.model, e.effort)
    if (id && e.index === 0) {
      await update($, runs, rs => ({ ...rs, [id]: label }))
    } else if (!id && (e.index === 0 || !(await read($, main))?.busy)) {
      const now = await $.clock.now()
      await update($, main, m => ({ run: label, busy: true, since: e.index === 0 ? now : (m?.since ?? now) }))
      if (e.index === 0) {
        await update($, lines, ls => Object.fromEntries(Object.entries(ls).filter(([, l]) => l.state !== 'blocked')))
      }
    }
    const result = yield* next(e)
    const said = id ? [] : statusLines(result.answer)
    if (said.length) {
      // ponytail: a unit left in progress, in review or pending whose closing line never comes stays until the
      // session ends; it labels the main row only while no later open line of its own does.
      await update($, lines, ls => {
        const out = { ...ls }
        for (const [unit, state] of said) {
          const key = unitKey(unit)
          delete out[key]
          if (state !== 'done') out[key] = { unit, state }
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
      await update($, main, m => (m ? { ...m, busy: false } : m))
    }
    return next(e)
  })

  // A Bash command's `flow stage` marks move the flow state; a denied command never ran, so its marks move nothing.
  on('tool.call', { tool: 'Bash' }, async ($, e, next) => {
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
    const f = await read($, flow)
    const ticking = subs.some(r => r.state === 'in progress')
    // A drawing is cached until a state it read is written, and the plan's ticks and landed commits are no state:
    // while a /flow change runs inline, nothing else redraws the band, so it redraws each second then too.
    const refresh = ticking || (!!m?.busy && f.change !== null)
    if (refresh && !tick) {
      tick = $.clock.every(1000, () => $.ui.invalidate('ui.render'))
    } else if (!refresh && tick) {
      tick.cancel()
      tick = undefined
    }
    // Shown while the main turn or a subagent runs; once all have finished the band hides and the hint line's tally stays.
    if (e.props.hasSurvey || !(m?.busy || ticking)) {
      return next(e)
    }
    const plan = await planOf($, f.change)
    const tasks = planTasks(plan)
    // The main agent's row, always first; it stays out of rows, so trim and the tally never see it.
    const top: Row = { id: 'main', n: 0, desc: mainLabel(await read($, lines), subs, f.stage, tasks), state: m?.busy ? 'in progress' : 'pending' }
    const since = m?.since ?? 0
    const all = [top, ...band(subs, pendingRows(plan, [top, ...live(subs, since)], await landedOf($, f.change)), f.change, since)]
    const labels: Record<string, string> = { ...(await read($, runs)), main: m?.run ?? '' }
    // The marker column lines up among the rows nested under main; the main row's label is not padded.
    const width = Math.max(0, ...all.filter(r => r.id !== 'main').map(r => labels[r.id]?.length ?? 0))
    const now = await $.clock.now()
    const { Box, Text } = $.ui.resolve(e)
    return (
      <Box flexDirection="column">
        {all.map(r => {
          const p = parts(r, tasks, labels[r.id], r.id === 'main' ? 0 : width, now)
          const isDone = r.state === 'done'
          return (
            <Text key={r.id}>
              <Text dimColor>{`${p.lead}${p.run ? `${p.run} ` : ''}`}</Text>
              <Text color={p.color}>{`${p.mark} `}</Text>
              <Text bold={r.state === 'in progress'} dimColor={isDone} strikethrough={isDone}>{p.text}</Text>
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
