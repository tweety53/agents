import { expect, mock, test } from 'claude-code/testing'

import { flowAfter, hintTail, line, lineRows, pendingRows, runLabel, statusLines, tally, trim } from './register'

const BAND = {
  plugin: 'subagent-board',
  component: 'AbovePrompt',
  props: { hasSurvey: false, isWorking: true, maxRows: 20, bodyColumns: 100, scroll: { offset: 0, bodyRows: 20 }, view: {} },
} as const

const spawn = (description: string) => ({
  tool_use_id: `tu-${description}`,
  prompt: 'p',
  description,
  subagentType: 'general-purpose',
  provider: { plugin: 'engine', tier: 'core' },
  parentModel: 'opus',
  background: true,
  fork: false,
}) as const

const complete = (agentId: string, reason: 'answer' | 'error') => ({
  answer: '',
  durationMs: 1,
  isAborted: false,
  turnId: `t-${agentId}`,
  agentId,
  reason,
})

const running = (n: number) => ({ id: `r${n}`, n, desc: 'x', state: 'in progress' }) as const

test('formats numbered and unnumbered rows', () => {
  expect(line({ id: 'a', n: 1, desc: 'Explore auth code', state: 'in progress' }, 3)).toBe('⎿ ◼ Task 1/3 (Explore auth code) — in progress')
  expect(line({ id: 'b', n: 2, desc: 'Task 21/22 finished-workout look', state: 'done' }, 3)).toBe('⎿ ✔ Task 21/22 (finished-workout look) — done')
  expect(line({ id: 'c', n: 3, desc: 'Task 2/5 (spec text)', state: 'blocked' }, 3)).toBe('⎿ ✘ Task 2/5 (spec text) — blocked')
  expect(line({ id: 'd', n: 4, desc: 'Tasks 3+4+7/22 (port guards)', state: 'in progress' }, 4)).toBe('⎿ ◼ Tasks 3+4+7/22 (port guards) — in progress')
  expect(line({ id: 'e', n: 5, desc: 'Tasks 3+4+7/22 (review)', state: 'done' }, 5)).toBe('⎿ ✔ Tasks 3+4+7/22 (review) — done')
  expect(line({ id: 'f', n: 6, desc: 'panel-1 primary review', state: 'done' }, 6, '', 0, false)).toBe('  ✔ Task 6/6 (panel-1 primary review) — done')
})

test('trim never drops a running row', () => {
  const six = [1, 2, 3, 4, 5, 6].map(running)
  expect(trim(six)).toEqual(six)
})

test('finished rows stay across prompts; past five the oldest finished goes', async ($, on) => {
  mock.clock(on)
  let n = 0
  on('agent.spawn', () => ({ model: 'opus', agentId: `ag${++n}` }))
  on('turn.complete', () => ({ text: '' }))
  on('prompt.submit', (_$, e) => ({ text: e.text }))
  on('ui.render', (r, e) => {
    const { Box } = r.ui.resolve(e)
    return <Box />
  })
  const shown = async (surface: 'terminal' | 'desktop') => {
    const ui = await $.ui.mount({ ...BAND, surface })
    // A row is the outer Text holding the whole line; its styled pieces are Texts inside it.
    const texts = (await ui.findAll({ type: 'Text' })).filter(t => /^(?:⎿ | {2})\S.* — \S/.test(t.text)).map(t => t.text)
    await ui.unmount()
    return texts
  }

  for (const surface of ['terminal', 'desktop'] as const) {
    expect(await shown(surface)).toEqual([])
  }

  await $.agent.spawn(spawn('Explore auth'))
  await $.agent.spawn(spawn('Read diff'))
  await $.agent.spawn(spawn('Run tests'))
  await $.turn.complete(complete('ag1', 'answer'))
  await $.turn.complete(complete('ag3', 'error'))
  await $.prompt.submit({ text: 'next', wait: false, origin: { kind: 'composer' } })

  for (const surface of ['terminal', 'desktop'] as const) {
    expect(await shown(surface)).toEqual([
      '⎿ ✔ Task 1/3 (Explore auth) — done · 0s',
      '  ◼ Task 2/3 (Read diff) — in progress · 0s',
      '  ✘ Task 3/3 (Run tests) — blocked · 0s',
    ])
  }

  await $.agent.spawn(spawn('Four'))
  await $.agent.spawn(spawn('Five'))
  await $.agent.spawn(spawn('Six'))
  expect(await shown('terminal')).toEqual([
    '⎿ ◼ Task 2/6 (Read diff) — in progress · 0s',
    '  ✘ Task 3/6 (Run tests) — blocked · 0s',
    '  ◼ Task 4/6 (Four) — in progress · 0s',
    '  ◼ Task 5/6 (Five) — in progress · 0s',
    '  ◼ Task 6/6 (Six) — in progress · 0s',
  ])
})

test('tally counts rows by state, leaving zeros out', () => {
  expect(tally([])).toBe('')
  expect(tally([running(1), { ...running(2), state: 'done' }, { ...running(3), state: 'done' }])).toBe('🔄 1  ✅ 2')
})

test('hint line gets the tally as its tail once a subagent exists', async ($, on) => {
  mock.clock(on)
  const tails: (string | undefined)[] = []
  on('agent.spawn', () => ({ model: 'opus', agentId: 'ag1' }))
  on('ui.render', (r, e) => {
    if (e.component === 'PromptHint') tails.push(e.props.tail)
    const { Text } = r.ui.resolve(e)
    return <Text>hint</Text>
  })
  const HINT = { plugin: 'subagent-board', component: 'PromptHint', props: { isDraft: false, isWorking: true, hint: 'esc to interrupt' } } as const

  for (const surface of ['terminal', 'desktop'] as const) {
    await (await $.ui.mount({ ...HINT, surface })).unmount()
  }
  await $.agent.spawn(spawn('One'))
  await (await $.ui.mount({ ...HINT, surface: 'terminal' })).unmount()

  expect(tails).toEqual([undefined, undefined, '· 🔄 1'])
})

test('runLabel names the model family and effort', () => {
  expect(runLabel('claude-opus-5-5', 'high')).toBe('opus-high')
  expect(runLabel('claude-sonnet-5-5')).toBe('sonnet')
  expect(runLabel('custom-model', 3)).toBe('custom-model-3')
})

test('a run label follows the lead, padded so the marker column lines up', () => {
  const row = { id: 'a', n: 2, desc: 'Count rules files', state: 'done' } as const
  expect(line(row, 6, 'opus-high', 9)).toBe('⎿ opus-high ✔ Task 2/6 (Count rules files) — done')
  expect(line({ ...row, state: 'in progress' }, 6, 'sonnet', 9, false)).toBe('  sonnet    ◼ Task 2/6 (Count rules files) — in progress')
  expect(line(row, 6, '', 9)).toBe('⎿           ✔ Task 2/6 (Count rules files) — done')
})

const NO_FLOW = { phase: null, change: null, ticket: null, stage: null }
const mark = (verb: string, key: string, name = 'kan-873-port-guards') =>
  verb === 'begin'
    ? `flow stage begin -command '/flow' -stage flow.${key} -harness claude -session-token mf-x ${name}`
    : `flow stage end -command '/flow' -stage flow.${key} -outcome completed ${name}`

test("flowAfter follows flow stage marks and the change name's ticket", () => {
  const impl = { phase: 'flow-implement', change: 'kan-873-port-guards', ticket: 'KAN-873', stage: 'sdd-tdd' } as const
  expect(flowAfter('ls -la', NO_FLOW)).toBe(NO_FLOW)
  expect(flowAfter(mark('begin', 'brainstorm'), NO_FLOW)).toEqual({ phase: 'flow-plan', change: 'kan-873-port-guards', ticket: 'KAN-873', stage: 'brainstorm' })
  expect(flowAfter(mark('begin', 'brainstorm', 'add-dark-mode'), NO_FLOW)).toEqual({ phase: 'flow-plan', change: 'add-dark-mode', ticket: null, stage: 'brainstorm' })
  expect(flowAfter(mark('begin', 'decide'), NO_FLOW)).toEqual(NO_FLOW)
  expect(flowAfter(mark('begin', 'decide'), impl)).toEqual({ ...impl, stage: 'decide' })
  expect(flowAfter(mark('begin', 'sdd-tdd'), NO_FLOW)).toEqual(impl)
  expect(flowAfter(`${mark('end', 'sdd-tdd')} && ${mark('begin', 'review-panel')}`, impl)).toEqual({ ...impl, stage: 'review-panel' })
  expect(flowAfter(mark('end', 'sdd-tdd'), impl)).toEqual({ ...impl, stage: null })
  expect(flowAfter(mark('end', 'review-panel'), impl)).toBe(impl)
  expect(flowAfter(mark('begin', 'preflight'), impl)).toEqual({ ...impl, phase: 'flow-integrate', stage: 'preflight' })
  expect(flowAfter(mark('end', 'write-in-progress'), impl)).toEqual(NO_FLOW)
  expect(flowAfter(mark('end', 'refresh-main-checkout'), { ...impl, phase: 'flow-integrate' })).toEqual(NO_FLOW)
})

test('hintTail joins the ticket, the phase, the running stage and the tally', () => {
  const impl = { phase: 'flow-implement', change: 'kan-873-port-guards', ticket: 'KAN-873', stage: null } as const
  expect(hintTail(NO_FLOW, '')).toBeUndefined()
  expect(hintTail(NO_FLOW, '✅ 5')).toBe('· ✅ 5')
  expect(hintTail({ ...NO_FLOW, phase: 'flow-plan' }, '')).toBe('· flow-plan')
  expect(hintTail({ ...impl, ticket: null }, '✅ 5')).toBe('· flow-implement -> ✅ 5')
  expect(hintTail(impl, '✅ 5')).toBe('· KAN-873 flow-implement -> ✅ 5')
  expect(hintTail({ ...impl, stage: 'review-panel' }, '🔍 2')).toBe('· KAN-873 flow-implement 🔍 -> 🔍 2')
  expect(hintTail({ ...impl, stage: 'document-fix' }, '')).toBe('· KAN-873 flow-implement 🔨')
  expect(hintTail({ ...impl, stage: 'verify' }, '')).toBe('· KAN-873 flow-implement 🧪')
  expect(hintTail({ ...impl, stage: 'visual-verify' }, '')).toBe('· KAN-873 flow-implement 👀')
  expect(hintTail({ ...impl, stage: 'sdd-tdd' }, '')).toBe('· KAN-873 flow-implement')
})

test('a running row shows its kind as its state word; a finished one its state', () => {
  const at = (desc: string, state: 'in progress' | 'done' = 'in progress') => line({ id: 'x', n: 1, desc, state }, 1)
  expect(at('Tasks 3+4/22 (review)')).toBe('⎿ ◼ Tasks 3+4/22 (review) — in review')
  expect(at('panel-1-primary correctness review')).toBe('⎿ ◼ Task 1/1 (panel-1-primary correctness review) — in review')
  expect(at('panel-fix-1 findings 1-6')).toBe('⎿ ◼ Task 1/1 (panel-fix-1 findings 1-6) — fix')
  expect(at('visual-verify login page')).toBe('⎿ ◼ Task 1/1 (visual-verify login page) — visual verify')
  expect(at('Run tests')).toBe('⎿ ◼ Task 1/1 (Run tests) — verify')
  expect(at('strip prefix')).toBe('⎿ ◼ Task 1/1 (strip prefix) — in progress')
  expect(at('read latest log')).toBe('⎿ ◼ Task 1/1 (read latest log) — in progress')
  expect(at('Run unittests')).toBe('⎿ ◼ Task 1/1 (Run unittests) — in progress')
  expect(at('Task 3/22 (port guard)')).toBe('⎿ ◼ Task 3/22 (port guard) — in progress')
  expect(at('Tasks 3+4/22 (review)', 'done')).toBe('⎿ ✔ Tasks 3+4/22 (review) — done')
  const rows = [
    { id: 'a', n: 1, desc: 'Task 1/3 (a)', state: 'in progress' },
    { id: 'b', n: 2, desc: 'panel-1 review', state: 'in progress' },
    { id: 'c', n: 3, desc: 'panel-2 review', state: 'in progress' },
    { id: 'd', n: 4, desc: 'x', state: 'done' },
  ] as const
  expect(tally([...rows])).toBe('🔄 1  🔍 2  ✅ 1')
})

test('rows carry their run label; the band hides once all finish; ticket and phase reach the hint', async ($, on) => {
  mock.clock(on)
  on('session.root', () => ({ value: '/u/Projects/agents' }))
  on('agent.spawn', () => ({ model: 'claude-opus-5-5', agentId: 'ag1' }))
  on('turn.complete', () => ({ text: '' }))
  on('turn.step', async function* (_$, e) {
    return { turnId: e.turnId, index: e.index, answer: '', toolUses: [], stopReason: 'end_turn', usage: null }
  })
  on('tool.call', () => ({ result: '' }))
  const tails: (string | undefined)[] = []
  on('ui.render', (r, e) => {
    if (e.component === 'PromptHint') tails.push(e.props.tail)
    const { Box } = r.ui.resolve(e)
    return <Box />
  })
  const shown = async () => {
    const ui = await $.ui.mount({ ...BAND, surface: 'terminal' })
    // A row is the outer Text holding the whole line; its styled pieces are Texts inside it.
    const texts = (await ui.findAll({ type: 'Text' })).filter(t => /^(?:⎿ | {2})\S.* — \S/.test(t.text)).map(t => t.text)
    await ui.unmount()
    return texts
  }
  const HINT = { plugin: 'subagent-board', component: 'PromptHint', props: { isDraft: false, isWorking: true, hint: '' } } as const

  await $.tool.call({ tool: 'Bash', command: mark('begin', 'sdd-tdd') })
  await $.agent.spawn(spawn('Task 2/7 (wire band)'))
  const step = $.turn.step({ turnId: 't1', index: 0, model: 'claude-opus-5-5', effort: 'high', messageCount: 1, agentId: 'ag1' })
  for await (const _ of step) {
    // drain
  }
  expect(await shown()).toEqual(['⎿ opus-high ◼ Task 2/7 (wire band) — in progress · 0s'])

  await $.turn.complete(complete('ag1', 'answer'))
  expect(await shown()).toEqual([])
  await (await $.ui.mount({ ...HINT, surface: 'terminal' })).unmount()
  expect(tails.at(-1)).toBe('· KAN-873 flow-implement -> ✅ 1')
})

test('a denied Bash call leaves the flow state as it was', async ($, on) => {
  mock.clock(on)
  on('tool.call', () => ({ deny: 'no' }))
  const tails: (string | undefined)[] = []
  on('ui.render', (r, e) => {
    if (e.component === 'PromptHint') tails.push(e.props.tail)
    const { Text } = r.ui.resolve(e)
    return <Text>hint</Text>
  })
  await $.tool.call({ tool: 'Bash', command: mark('begin', 'sdd-tdd') })
  await (await $.ui.mount({ plugin: 'subagent-board', component: 'PromptHint', props: { isDraft: false, isWorking: true, hint: '' }, surface: 'terminal' })).unmount()
  expect(tails).toEqual([undefined])
})

test('rows take the task-list styles: a done row green-ticked, dim and struck; a running row bold', async ($, on) => {
  mock.clock(on)
  let n = 0
  on('agent.spawn', () => ({ model: 'opus', agentId: `ag${++n}` }))
  on('turn.complete', () => ({ text: '' }))
  on('ui.render', (r, e) => {
    const { Box } = r.ui.resolve(e)
    return <Box />
  })
  await $.agent.spawn(spawn('Explore auth'))
  await $.agent.spawn(spawn('Read diff'))
  await $.agent.spawn(spawn('Run tests'))
  await $.turn.complete(complete('ag1', 'answer'))
  await $.turn.complete(complete('ag3', 'error'))
  for (const surface of ['terminal', 'desktop'] as const) {
    const ui = await $.ui.mount({ ...BAND, surface })
    const at = async (text: string) => (await ui.findAll({ type: 'Text' })).find(t => t.text === text)?.props
    expect(await at('✔ ')).toMatchObject({ color: 'success' })
    expect(await at('Task 1/3 (Explore auth)')).toMatchObject({ dimColor: true, strikethrough: true, bold: false })
    expect(await at('◼ ')).toMatchObject({ color: 'claude' })
    expect(await at('Task 2/3 (Read diff)')).toMatchObject({ bold: true, dimColor: false, strikethrough: false })
    expect(await at('✘ ')).toMatchObject({ color: 'error' })
    expect(await at('Task 3/3 (Run tests)')).toMatchObject({ bold: false, strikethrough: false })
    await ui.unmount()
  }
})

test("the main agent's running turn is a row of its own, outside the task numbering", async ($, on) => {
  mock.clock(on)
  on('agent.spawn', () => ({ model: 'claude-sonnet-5-5', agentId: 'ag1' }))
  on('turn.complete', () => ({ text: '' }))
  on('turn.step', async function* (_$, e) {
    return { turnId: e.turnId, index: e.index, answer: '', toolUses: [], stopReason: 'end_turn', usage: null }
  })
  on('tool.call', () => ({ result: '' }))
  on('ui.render', (r, e) => {
    const { Box } = r.ui.resolve(e)
    return <Box />
  })
  const shown = async () => {
    const ui = await $.ui.mount({ ...BAND, surface: 'terminal' })
    const texts = (await ui.findAll({ type: 'Text' })).filter(t => /^(?:⎿ | {2})\S.* — \S/.test(t.text)).map(t => t.text)
    await ui.unmount()
    return texts
  }
  const drain = async (agentId?: string, index = 0, model = 'claude-opus-5-5', effort?: 'high') => {
    for await (const _ of $.turn.step({ turnId: 't1', index, model, effort, messageCount: 1, agentId })) {
      // drain
    }
  }

  await drain(undefined, 0, 'claude-opus-5-5', 'high')
  expect(await shown()).toEqual(['⎿ opus-high ◼ main — in progress'])

  await $.tool.call({ tool: 'Bash', command: 'scripts/run-guard-tests.sh', description: 'Run guard tests' })
  await drain(undefined, 1, 'claude-opus-5-5', 'high')
  expect(await shown()).toEqual(['⎿ opus-high ◼ main (Run guard tests) — verify'])

  await $.agent.spawn(spawn('Explore auth'))
  await drain('ag1', 0, 'claude-sonnet-5-5')
  // The subagent's own command, its loop named as the engine names it; it leaves the main row's description alone.
  const sub = { tool: 'Bash', command: 'ls', description: 'List subagent files', agentId: 'ag1' } as const
  await $.tool.call(sub)
  expect(await shown()).toEqual([
    '⎿ opus-high ◼ main (Run guard tests) — verify',
    '  sonnet    ◼ Task 1/1 (Explore auth) — in progress · 0s',
  ])

  await $.turn.complete({ ...complete('ag1', 'answer'), agentId: undefined })
  expect(await shown()).toEqual(['⎿ sonnet ◼ Task 1/1 (Explore auth) — in progress · 0s'])
  await $.turn.complete(complete('ag1', 'answer'))
  expect(await shown()).toEqual([])
})

test("the main row names a Bash command's description while the command runs", async ($, on) => {
  mock.clock(on)
  let release = () => {}
  const gate = new Promise<void>(resolve => {
    release = resolve
  })
  let reached = () => {}
  const atCore = new Promise<void>(resolve => {
    reached = resolve
  })
  on('turn.step', async function* (_$, e) {
    return { turnId: e.turnId, index: e.index, answer: '', toolUses: [], stopReason: 'end_turn', usage: null }
  })
  on('tool.call', async () => {
    reached()
    await gate
    return { result: '' }
  })
  on('ui.render', (r, e) => {
    const { Box } = r.ui.resolve(e)
    return <Box />
  })
  for await (const _ of $.turn.step({ turnId: 't1', index: 0, model: 'claude-opus-5-5', effort: 'high', messageCount: 1 })) {
    // drain
  }

  const call = $.tool.call({ tool: 'Bash', command: 'scripts/run-guard-tests.sh', description: 'Run guard tests' })
  // Mid-call: the command has reached the core and is held open there, so the row must already name it.
  await atCore
  const ui = await $.ui.mount({ ...BAND, surface: 'terminal' })
  const texts = (await ui.findAll({ type: 'Text' })).filter(t => /^(?:⎿ | {2})\S.* — \S/.test(t.text)).map(t => t.text)
  await ui.unmount()
  release()
  await call
  expect(texts).toEqual(['⎿ opus-high ◼ main (Run guard tests) — verify'])
})

const PLAN = [
  '# Plan',
  '',
  ...Array.from({ length: 25 }, (_, i) => `- [${i < 19 ? 'x' : ' '}] ${i + 1}. Step ${i + 1}\n  - [ ] a step checkbox, not a task`),
].join('\n')

test('pendingRows lists the unticked plan tasks no row has taken', () => {
  const taken = [{ id: 'a', n: 1, desc: 'Task 20/25 (Step 20)', state: 'in progress' }, { id: 'b', n: 2, desc: 'Tasks 21+22/25 (pair)', state: 'done' }] as const
  expect(pendingRows(PLAN, [...taken]).map(r => line(r, 0, '', 0, false))).toEqual([
    '  ◻ Task 23/25 (Step 23) — pending',
    '  ◻ Task 24/25 (Step 24) — pending',
    '  ◻ Task 25/25 (Step 25) — pending',
  ])
  expect(pendingRows('', [])).toEqual([])
})

test("sibling layout: the running /flow change's pending tasks, read from <repo>-worktrees/<change>, follow the dispatched rows", async ($, on) => {
  mock.clock(on)
  let n = 0
  on('agent.spawn', () => ({ model: 'opus', agentId: `ag${++n}` }))
  on('tool.call', () => ({ result: '' }))
  on('session.root', () => ({ value: '/u/Projects/agents' }))
  const reads: string[] = []
  on('fs.read', (_$, e) => {
    reads.push(e.path)
    if (e.path !== '/u/Projects/agents-worktrees/kan-873-port-guards/spectre/changes/kan-873-port-guards/tasks.md') throw new Error('ENOENT')
    return { value: PLAN }
  })
  on('ui.render', (r, e) => {
    const { Box } = r.ui.resolve(e)
    return <Box />
  })
  const shown = async () => {
    const ui = await $.ui.mount({ ...BAND, surface: 'terminal' })
    const texts = (await ui.findAll({ type: 'Text' })).filter(t => /^(?:⎿ | {2})\S.* — \S/.test(t.text)).map(t => t.text)
    await ui.unmount()
    return texts
  }

  await $.agent.spawn(spawn('Explore auth'))
  expect(await shown()).toEqual(['⎿ ◼ Task 1/1 (Explore auth) — in progress · 0s'])
  expect(reads).toEqual([])

  await $.tool.call({ tool: 'Bash', command: mark('begin', 'sdd-tdd') })
  await $.agent.spawn(spawn('Task 20/25 (Step 20)'))
  await $.agent.spawn(spawn('Tasks 21+22/25 (pair)'))
  expect(await shown()).toEqual([
    '⎿ ◼ Task 1/3 (Explore auth) — in progress · 0s',
    '  ◼ Task 20/25 (Step 20) — in progress · 0s',
    '  ◼ Tasks 21+22/25 (pair) — in progress · 0s',
    '  ◻ Task 23/25 (Step 23) — pending',
    '  ◻ Task 24/25 (Step 24) — pending',
  ])
})

test('the band draws at most main plus five rows: dispatched rows first, then the earliest pending', async ($, on) => {
  mock.clock(on)
  on('session.root', () => ({ value: '/u/Projects/agents' }))
  let n = 0
  on('agent.spawn', () => ({ model: 'opus', agentId: `ag${++n}` }))
  on('turn.step', async function* (_$, e) {
    return { turnId: e.turnId, index: e.index, answer: '', toolUses: [], stopReason: 'end_turn', usage: null }
  })
  on('tool.call', () => ({ result: '' }))
  on('fs.read', () => ({ value: PLAN.replace(/- \[x\] (?:[4-9]|1\d)\. /g, m => m.replace('x', ' ')) }))
  on('ui.render', (r, e) => {
    const { Box, Text } = r.ui.resolve(e)
    return e.component === 'PromptHint' ? <Text>hint</Text> : <Box />
  })
  const shown = async () => {
    const ui = await $.ui.mount({ ...BAND, surface: 'terminal' })
    const texts = (await ui.findAll({ type: 'Text' })).filter(t => /^(?:⎿ | {2}) *\S.* — \S/.test(t.text)).map(t => t.text)
    await ui.unmount()
    return texts
  }
  for await (const _ of $.turn.step({ turnId: 't1', index: 0, model: 'claude-opus-5-5', messageCount: 1 })) {
    // drain
  }

  await $.tool.call({ tool: 'Bash', command: mark('begin', 'sdd-tdd') })
  await $.agent.spawn(spawn('Task 1/25 (Step 1)'))
  await $.agent.spawn(spawn('Task 2/25 (Step 2)'))
  await $.agent.spawn(spawn('Task 3/25 (Step 3)'))
  expect(await shown()).toEqual([
    '⎿ opus ◼ main — in progress',
    '       ◼ Task 1/25 (Step 1) — in progress · 0s',
    '       ◼ Task 2/25 (Step 2) — in progress · 0s',
    '       ◼ Task 3/25 (Step 3) — in progress · 0s',
    '       ◻ Task 4/25 (Step 4) — pending',
    '       ◻ Task 5/25 (Step 5) — pending',
  ])

  for (const k of [4, 5, 6, 7]) await $.agent.spawn(spawn(`Task ${k}/25 (Step ${k})`))
  expect(await shown()).toEqual([
    '⎿ opus ◼ main — in progress',
    ...[1, 2, 3, 4, 5].map(k => `       ◼ Task ${k}/25 (Step ${k}) — in progress · 0s`),
  ])
})

test("a subagent row's elapsed time and tokens follow its state", () => {
  const row = { id: 'a', n: 3, desc: 'Task 3/8 (desc)', state: 'in progress', start: 1000 } as const
  expect(line(row, 8, 'opus-medium', 11, true, 1000 + 252_000)).toBe('⎿ opus-medium ◼ Task 3/8 (desc) — in progress · 4m12s')
  expect(line({ ...row, tokens: 1_234_567 }, 8, '', 0, true, 1000 + 7_000)).toBe('⎿ ◼ Task 3/8 (desc) — in progress · 7s · 1.2M tok')
  expect(line({ ...row, state: 'done', end: 1000 + 3_600_000 + 65_000, tokens: 45_200 }, 8, '', 0, true, 9e9)).toBe('⎿ ✔ Task 3/8 (desc) — done · 1h01m · 45.2k tok')
  expect(line({ ...row, tokens: 850 }, 8, '', 0, true, 1000)).toBe('⎿ ◼ Task 3/8 (desc) — in progress · 0s · 850 tok')
  expect(line({ id: 'p', n: 4, desc: 'Task 4/8 (later)', state: 'pending' }, 8, '', 0, true, 5000)).toBe('⎿ ◻ Task 4/8 (later) — pending')
})

test('a running row ticks and takes its latest response tokens; a finished row freezes', async ($, on) => {
  const clock = mock.clock(on)
  let n = 0
  on('agent.spawn', () => ({ model: 'opus', agentId: `ag${++n}` }))
  on('turn.complete', () => ({ text: '' }))
  const usage = (input: number, output: number) => ({ model: 'claude-opus-5-5', input_tokens: input, output_tokens: output, cache_read_input_tokens: 1_000_000, cache_creation_input_tokens: 100_000 })
  let next = usage(0, 0)
  on('turn.step', async function* (_$, e) {
    return { turnId: e.turnId, index: e.index, answer: '', toolUses: [], stopReason: 'end_turn', usage: next }
  })
  on('ui.render', (r, e) => {
    const { Box } = r.ui.resolve(e)
    return <Box />
  })
  const step = async (index: number) => {
    for await (const _ of $.turn.step({ turnId: 't1', index, model: 'claude-opus-5-5', effort: 'medium', messageCount: 1, agentId: 'ag1' })) {
      // drain
    }
  }
  const rows = async (ui: { findAll: (q: { type: 'Text' }) => Promise<{ text: string }[]> }) =>
    (await ui.findAll({ type: 'Text' })).filter(t => /^(?:⎿ | {2})\S.* — \S/.test(t.text)).map(t => t.text)

  await $.agent.spawn(spawn('Task 3/8 (desc)'))
  next = usage(50_000, 2_000)
  await step(0)
  next = usage(130_000, 3_000)
  await step(1)
  const ui = await $.ui.mount({ ...BAND, surface: 'terminal' })
  expect(await rows(ui)).toEqual(['⎿ opus-medium ◼ Task 3/8 (desc) — in progress · 0s · 1.2M tok'])
  await clock.advance(252_000)
  expect(await rows(ui)).toEqual(['⎿ opus-medium ◼ Task 3/8 (desc) — in progress · 4m12s · 1.2M tok'])
  await ui.unmount()

  await $.agent.spawn(spawn('Task 4/8 (other)'))
  await $.turn.complete(complete('ag1', 'answer'))
  await clock.advance(60_000)
  const after = await $.ui.mount({ ...BAND, surface: 'terminal' })
  expect((await rows(after))[0]).toBe('⎿ opus-medium ✔ Task 3/8 (desc) — done · 4m12s · 1.2M tok')
  await after.unmount()
})

test('statusLines reads the be-brief status lines and nothing else', () => {
  const text = [
    'Some prose — not a line.',
    '⏳ Full backend test suite — pending',
    '🔄 Tasks 23–25/25 (e2e, fidelity, live) — in progress',
    '🔍 Review panel (primary + principles) — in review',
    '⛔ Task 4/8 (port) — blocked',
    '✅ Task 22/22 (spec text) — done',
    '🔄 Task 5/8 — fixing',
  ].join('\n')
  expect(statusLines(text)).toEqual([
    ['Full backend test suite', 'pending'],
    ['Tasks 23–25/25 (e2e, fidelity, live)', 'in progress'],
    ['Review panel (primary + principles)', 'in review'],
    ['Task 4/8 (port)', 'blocked'],
    ['Task 22/22 (spec text)', 'done'],
  ])
})

test('lineRows leaves out a unit a subagent row already numbers', () => {
  const lines = {
    'Full backend test suite': { unit: 'Full backend test suite', state: 'pending' },
    'Tasks 23–25/25': { unit: 'Tasks 23–25/25 (e2e, fidelity, live)', state: 'in progress' },
    'Task 3/25': { unit: 'Task 3/25 (port)', state: 'in progress' },
  } as const
  const subs = [{ id: 'a', n: 1, desc: 'Tasks 2+3/25 (pair)', state: 'in progress' }] as const
  expect(lineRows(lines, [...subs]).map(r => line(r, 1, '', 0, false))).toEqual([
    '  ◻ Full backend test suite — pending',
    '  ◼ Tasks 23–25/25 (e2e, fidelity, live) — in progress',
  ])
})

test("the main loop's status lines are rows: latest line per unit, done drops off, no unit twice", async ($, on) => {
  mock.clock(on)
  on('session.root', () => ({ value: '/u/Projects/agents' }))
  let n = 0
  on('agent.spawn', () => ({ model: 'opus', agentId: `ag${++n}` }))
  on('tool.call', () => ({ result: '' }))
  on('fs.read', () => ({ value: PLAN }))
  let answer = ''
  on('turn.step', async function* (_$, e) {
    return { turnId: e.turnId, index: e.index, answer, toolUses: [], stopReason: 'end_turn', usage: null }
  })
  on('ui.render', (r, e) => {
    const { Box } = r.ui.resolve(e)
    return <Box />
  })
  const say = async (text: string, agentId?: string) => {
    answer = text
    for await (const _ of $.turn.step({ turnId: 't1', index: 1, model: 'claude-opus-5-5', messageCount: 1, agentId })) {
      // drain
    }
  }
  const shown = async () => {
    const ui = await $.ui.mount({ ...BAND, surface: 'terminal' })
    const texts = (await ui.findAll({ type: 'Text' })).filter(t => /^(?:⎿ | {2}) *\S.* — \S/.test(t.text)).map(t => t.text)
    await ui.unmount()
    return texts
  }

  await $.tool.call({ tool: 'Bash', command: mark('begin', 'sdd-tdd') })
  await $.agent.spawn(spawn('Task 20/25 (Step 20)'))
  await say('⏳ Full backend test suite — pending\n🔄 Task 20/25 (Step 20) — in progress\n🔄 Tasks 21–22/25 (pair) — in progress')
  // A subagent's own status lines are its own business.
  await say('⏳ Subagent unit — pending', 'ag1')
  expect(await shown()).toEqual([
    '⎿ opus ◼ main — in progress',
    '       ◼ Task 20/25 (Step 20) — in progress · 0s',
    '       ◻ Full backend test suite — pending',
    '       ◼ Tasks 21–22/25 (pair) — in progress',
    '       ◻ Task 23/25 (Step 23) — pending',
    '       ◻ Task 24/25 (Step 24) — pending',
  ])

  await say('🔄 Full backend test suite (gradle) — in progress\n✅ Tasks 21–22/25 (pair) — done')
  expect(await shown()).toEqual([
    '⎿ opus ◼ main — in progress',
    '       ◼ Task 20/25 (Step 20) — in progress · 0s',
    '       ◼ Full backend test suite (gradle) — in progress',
    '       ◻ Task 21/25 (Step 21) — pending',
    '       ◻ Task 22/25 (Step 22) — pending',
    '       ◻ Task 23/25 (Step 23) — pending',
  ])
})
