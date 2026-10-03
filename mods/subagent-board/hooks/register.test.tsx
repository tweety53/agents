import { expect, test } from 'claude-code/testing'

import { flowAfter, hintTail, line, runLabel, tally, trim } from './register'

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
  expect(line({ id: 'a', n: 1, desc: 'Explore auth code', state: 'in progress' }, 3)).toBe('🔄 Task 1/3 (Explore auth code) — in progress')
  expect(line({ id: 'b', n: 2, desc: 'Task 21/22 finished-workout look', state: 'done' }, 3)).toBe('✅ Task 21/22 (finished-workout look) — done')
  expect(line({ id: 'c', n: 3, desc: 'Task 2/5 (spec text)', state: 'blocked' }, 3)).toBe('⛔ Task 2/5 (spec text) — blocked')
  expect(line({ id: 'd', n: 4, desc: 'Tasks 3+4+7/22 (port guards)', state: 'in progress' }, 4)).toBe('🔄 Tasks 3+4+7/22 (port guards) — in progress')
  expect(line({ id: 'e', n: 5, desc: 'Tasks 3+4+7/22 (review)', state: 'done' }, 5)).toBe('✅ Tasks 3+4+7/22 (review) — done')
  expect(line({ id: 'f', n: 6, desc: 'panel-1 primary review', state: 'done' }, 6)).toBe('✅ Task 6/6 (panel-1 primary review) — done')
})

test('trim never drops a running row', () => {
  const six = [1, 2, 3, 4, 5, 6].map(running)
  expect(trim(six)).toEqual(six)
})

test('finished rows stay across prompts; past five the oldest finished goes', async ($, on) => {
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
    const texts = (await ui.findAll({ type: 'Text' })).filter(t => /^\S.* — \S/.test(t.text) && t.text.includes('Task')).map(t => t.text)
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
      '✅ Task 1/3 (Explore auth) — done',
      '🔄 Task 2/3 (Read diff) — in progress',
      '⛔ Task 3/3 (Run tests) — blocked',
    ])
  }

  await $.agent.spawn(spawn('Four'))
  await $.agent.spawn(spawn('Five'))
  await $.agent.spawn(spawn('Six'))
  expect(await shown('terminal')).toEqual([
    '🔄 Task 2/6 (Read diff) — in progress',
    '⛔ Task 3/6 (Run tests) — blocked',
    '🔄 Task 4/6 (Four) — in progress',
    '🔄 Task 5/6 (Five) — in progress',
    '🔄 Task 6/6 (Six) — in progress',
  ])
})

test('tally counts rows by state, leaving zeros out', () => {
  expect(tally([])).toBe('')
  expect(tally([running(1), { ...running(2), state: 'done' }, { ...running(3), state: 'done' }])).toBe('🔄 1  ✅ 2')
})

test('hint line gets the tally as its tail once a subagent exists', async ($, on) => {
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

test('a run label leads the row, padded so the emoji column lines up', () => {
  const row = { id: 'a', n: 2, desc: 'Count rules files', state: 'done' } as const
  expect(line(row, 6, 'opus-high', 9)).toBe('opus-high ✅ Task 2/6 (Count rules files) — done')
  expect(line({ ...row, state: 'in progress' }, 6, 'sonnet', 9)).toBe('sonnet    🔄 Task 2/6 (Count rules files) — in progress')
  expect(line(row, 6, '', 9)).toBe('          ✅ Task 2/6 (Count rules files) — done')
})

const NO_FLOW = { phase: null, ticket: null, stage: null }
const mark = (verb: string, key: string, name = 'kan-873-port-guards') =>
  verb === 'begin'
    ? `flow stage begin -command '/flow' -stage flow.${key} -harness claude -session-token mf-x ${name}`
    : `flow stage end -command '/flow' -stage flow.${key} -outcome completed ${name}`

test("flowAfter follows flow stage marks and the change name's ticket", () => {
  const impl = { phase: 'flow-implement', ticket: 'KAN-873', stage: 'sdd-tdd' } as const
  expect(flowAfter('ls -la', NO_FLOW)).toBe(NO_FLOW)
  expect(flowAfter(mark('begin', 'brainstorm'), NO_FLOW)).toEqual({ phase: 'flow-plan', ticket: 'KAN-873', stage: 'brainstorm' })
  expect(flowAfter(mark('begin', 'brainstorm', 'add-dark-mode'), NO_FLOW)).toEqual({ phase: 'flow-plan', ticket: null, stage: 'brainstorm' })
  expect(flowAfter(mark('begin', 'decide'), NO_FLOW)).toEqual(NO_FLOW)
  expect(flowAfter(mark('begin', 'decide'), impl)).toEqual({ ...impl, stage: 'decide' })
  expect(flowAfter(mark('begin', 'sdd-tdd'), NO_FLOW)).toEqual(impl)
  expect(flowAfter(`${mark('end', 'sdd-tdd')} && ${mark('begin', 'review-panel')}`, impl)).toEqual({ ...impl, stage: 'review-panel' })
  expect(flowAfter(mark('end', 'sdd-tdd'), impl)).toEqual({ ...impl, stage: null })
  expect(flowAfter(mark('end', 'review-panel'), impl)).toBe(impl)
  expect(flowAfter(mark('begin', 'preflight'), impl)).toEqual({ ...impl, phase: 'flow-integrate', stage: 'preflight' })
  expect(flowAfter(mark('end', 'write-in-progress'), impl)).toEqual(NO_FLOW)
  expect(flowAfter(mark('end', 'push-archive'), { ...impl, phase: 'flow-integrate' })).toEqual(NO_FLOW)
})

test('hintTail joins the ticket, the phase, the running stage and the tally', () => {
  const impl = { phase: 'flow-implement', ticket: 'KAN-873', stage: null } as const
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

test('a running row shows its kind; a finished one its state', () => {
  const at = (desc: string, state: 'in progress' | 'done' = 'in progress') => line({ id: 'x', n: 1, desc, state }, 1)
  expect(at('Tasks 3+4/22 (review)')).toBe('🔍 Tasks 3+4/22 (review) — in review')
  expect(at('panel-1-primary correctness review')).toBe('🔍 Task 1/1 (panel-1-primary correctness review) — in review')
  expect(at('panel-fix-1 findings 1-6')).toBe('🔨 Task 1/1 (panel-fix-1 findings 1-6) — in progress')
  expect(at('visual-verify login page')).toBe('👀 Task 1/1 (visual-verify login page) — in progress')
  expect(at('Run tests')).toBe('🧪 Task 1/1 (Run tests) — in progress')
  expect(at('Task 3/22 (port guard)')).toBe('🔄 Task 3/22 (port guard) — in progress')
  expect(at('Tasks 3+4/22 (review)', 'done')).toBe('✅ Tasks 3+4/22 (review) — done')
  const rows = [
    { id: 'a', n: 1, desc: 'Task 1/3 (a)', state: 'in progress' },
    { id: 'b', n: 2, desc: 'panel-1 review', state: 'in progress' },
    { id: 'c', n: 3, desc: 'panel-2 review', state: 'in progress' },
    { id: 'd', n: 4, desc: 'x', state: 'done' },
  ] as const
  expect(tally([...rows])).toBe('🔄 1  🔍 2  ✅ 1')
})

test('rows carry their run label; the band hides once all finish; ticket and phase reach the hint', async ($, on) => {
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
    const texts = (await ui.findAll({ type: 'Text' })).filter(t => /^\S.* — \S/.test(t.text) && t.text.includes('Task')).map(t => t.text)
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
  expect(await shown()).toEqual(['opus-high 🔄 Task 2/7 (wire band) — in progress'])

  await $.turn.complete(complete('ag1', 'answer'))
  expect(await shown()).toEqual([])
  await (await $.ui.mount({ ...HINT, surface: 'terminal' })).unmount()
  expect(tails.at(-1)).toBe('· KAN-873 flow-implement -> ✅ 1')
})
