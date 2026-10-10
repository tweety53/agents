import { expect, mock, test } from 'claude-code/testing'

import { landed, landedIn } from './landed'
import { band, flowAfter, hintTail, line, mainLabel, pendingRows, planTasks, runLabel, statusLines, tally, taskCount, trim, unitKey } from './register'

const BAND = {
  plugin: 'flow-task-list',
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

const P22 = Array.from({ length: 22 }, (_, i) => i + 1)
const P5 = [1, 2, 3, 4, 5]

test('formats numbered and unnumbered rows; a group shows "Task 1+2+3/n" running and its highest task done', () => {
  expect(line({ id: 'a', n: 1, desc: 'Explore auth code', state: 'in progress' }, P22)).toBe('  ⎿ ◼ Explore auth code — in progress')
  expect(line({ id: 'b', n: 2, desc: 'Task 21/22 finished-workout look', state: 'done' }, P22)).toBe('  ⎿ ✔ Task 21/22 (finished-workout look) — done')
  expect(line({ id: 'c', n: 3, desc: 'Task 2/5 (spec text)', state: 'blocked' }, P5)).toBe('  ⎿ ✘ Task 2/5 (spec text) — blocked')
  expect(line({ id: 'd', n: 4, desc: 'Tasks 3+4+7/22 (port guards)', state: 'in progress' }, P22)).toBe('  ⎿ ◼ Task 3+4+7/22 (port guards) — in progress')
  expect(line({ id: 'e', n: 5, desc: 'Tasks 3+4+7/22 (port guards)', state: 'done' }, P22)).toBe('  ⎿ ✔ Task 7/22 (port guards) — done')
  expect(line({ id: 'e', n: 5, desc: 'Tasks 20–22/22 (e2e)', state: 'done' }, P22)).toBe('  ⎿ ✔ Task 22/22 (e2e) — done')
  expect(line({ id: 'f', n: 6, desc: 'panel-1 primary review', state: 'done' }, P22)).toBe('  ⎿ ✔ panel-1 primary review — done')
})

test('only a row naming tasks of the running plan is numbered; any other row shows its text as written', () => {
  const at = (desc: string, plan: number[]) => line({ id: 'x', n: 4, desc, state: 'in progress' }, plan)
  // An in-run pipeline fix, a panel, a fix round: no plan task, so no number, whatever the spawn order.
  expect(at('pipeline-fix-1 flow-task-list landed count', P22)).toBe('  ⎿ ◼ pipeline-fix-1 flow-task-list landed count — fix')
  expect(at('panel-1-primary correctness review', P22)).toBe('  ⎿ ◼ panel-1-primary correctness review — review')
  // A prefix naming a number the plan lacks, or another plan's size, or with no plan running: the text as written.
  expect(at('Task 23/22 (other)', P22)).toBe('  ⎿ ◼ Task 23/22 (other) — in progress')
  expect(at('Task 3/9 (other plan)', P22)).toBe('  ⎿ ◼ Task 3/9 (other plan) — in progress')
  expect(at('Tasks 3+30/22 (pair)', P22)).toBe('  ⎿ ◼ Tasks 3+30/22 (pair) — in progress')
  expect(at('Task 3/22 (port)', [])).toBe('  ⎿ ◼ Task 3/22 (port) — in progress')
  expect(at('Task 3/22 (port)', P22)).toBe('  ⎿ ◼ Task 3/22 (port) — in progress')
  expect(at('Task 26/26 review', Array.from({ length: 28 }, (_, i) => i + 1))).toBe('  ⎿ ◼ Task 26/26 review — review')
  // The main agent's row is first, never renumbered, and waits between its turns.
  expect(line({ id: 'main', n: 0, desc: 'Task 3/22 (port)', state: 'in progress' }, P22)).toBe('⎿ ◼ main: Task 3/22 (port) — in progress')
  expect(line({ id: 'main', n: 0, desc: 'main', state: 'pending' }, P22, 'opus-high')).toBe('⎿ opus-high ◻ main — waiting')
})

test('trim never drops a running row', () => {
  const six = [1, 2, 3, 4, 5, 6].map(running)
  expect(trim(six)).toEqual(six)
})

test('finished rows stay across prompts, the last two shown with no plan pending, a failed one among the running; past five the oldest finished goes', async ($, on) => {
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
      '⎿ ◻ main — waiting',
      '  ⎿ ✔ Explore auth — done · 0s',
      '  ⎿ ◼ Read diff — in progress · 0s',
      '  ⎿ ✘ Run tests — blocked · 0s',
    ])
  }

  await $.agent.spawn(spawn('Four'))
  await $.agent.spawn(spawn('Five'))
  await $.agent.spawn(spawn('Six'))
  expect(await shown('terminal')).toEqual([
    '⎿ ◻ main — waiting',
    '  ⎿ ◼ Read diff — in progress · 0s',
    '  ⎿ ✘ Run tests — blocked · 0s',
    '  ⎿ ◼ Four — in progress · 0s',
    '  ⎿ ◼ Five — in progress · 0s',
    '  ⎿ ◼ Six — in progress · 0s',
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
  const HINT = { plugin: 'flow-task-list', component: 'PromptHint', props: { isDraft: false, isWorking: true, hint: 'esc to interrupt' } } as const

  for (const surface of ['terminal', 'desktop'] as const) {
    await (await $.ui.mount({ ...HINT, surface })).unmount()
  }
  await $.agent.spawn(spawn('One'))
  await (await $.ui.mount({ ...HINT, surface: 'terminal' })).unmount()

  expect(tails).toEqual([undefined, undefined, '🔄 1'])
})

test('runLabel names the model family and effort', () => {
  expect(runLabel('claude-opus-5-5', 'high')).toBe('opus-high')
  expect(runLabel('claude-sonnet-5-5')).toBe('sonnet')
  expect(runLabel('custom-model', 3)).toBe('custom-model-3')
})

test('a row nested under the main row leads with an indented corner on every row', () => {
  const row = { id: 'r', n: 3, desc: 'Task 3/16 (port guards)', state: 'in progress' } as const
  const P16 = Array.from({ length: 16 }, (_, i) => i + 1)
  expect(line(row, P16, 'sonnet-low', 11)).toBe('  ⎿ sonnet-low  ◼ Task 3/16 (port guards) — in progress')
  expect(line(row, P16, 'opus-medium', 11)).toBe('  ⎿ opus-medium ◼ Task 3/16 (port guards) — in progress')
  expect(line({ id: 'p', n: 4, desc: 'Task 4/16 (later)', state: 'pending' }, P16)).toBe('  ⎿ ◻ Task 4/16 (later) — pending')
})

test('a run label follows the lead, padded so the marker column lines up', () => {
  const row = { id: 'a', n: 2, desc: 'Count rules files', state: 'done' } as const
  expect(line(row, [], 'opus-high', 9)).toBe('  ⎿ opus-high ✔ Count rules files — done')
  expect(line({ ...row, state: 'in progress' }, [], 'sonnet', 9)).toBe('  ⎿ sonnet    ◼ Count rules files — in progress')
  expect(line(row, [], '', 9)).toBe('  ⎿           ✔ Count rules files — done')
})

const NO_FLOW = { phase: null, change: null, ticket: null, stage: null }
const mark = (verb: string, key: string, name = 'kan-873-port-guards') =>
  verb === 'begin'
    ? `flow stage begin -command '/flow' -stage flow.${key} -harness claude -session-token mf-x ${name}`
    : `flow stage end -command '/flow' -stage flow.${key} -outcome completed ${name}`

test("flowAfter follows flow stage marks, the stage begun last and the change name's ticket", () => {
  const impl = { phase: 'flow-implement', change: 'kan-873-port-guards', ticket: 'KAN-873', stage: 'sdd-tdd' } as const
  expect(flowAfter('ls -la', NO_FLOW)).toBe(NO_FLOW)
  expect(flowAfter(mark('begin', 'brainstorm'), NO_FLOW)).toEqual({ phase: 'flow-plan', change: 'kan-873-port-guards', ticket: 'KAN-873', stage: 'brainstorm' })
  expect(flowAfter(mark('begin', 'brainstorm', 'add-dark-mode'), NO_FLOW)).toEqual({ phase: 'flow-plan', change: 'add-dark-mode', ticket: null, stage: 'brainstorm' })
  expect(flowAfter(mark('begin', 'decide'), NO_FLOW)).toEqual(NO_FLOW)
  expect(flowAfter(mark('begin', 'decide'), impl)).toEqual({ ...impl, stage: 'decide' })
  expect(flowAfter(mark('begin', 'sdd-tdd'), NO_FLOW)).toEqual(impl)
  expect(flowAfter(`${mark('end', 'sdd-tdd')} && ${mark('begin', 'review-panel')}`, impl)).toEqual({ ...impl, stage: 'review-panel' })
  expect(flowAfter(mark('end', 'review-panel'), impl)).toBe(impl)
  expect(flowAfter(mark('begin', 'preflight'), impl)).toEqual({ ...impl, phase: 'flow-integrate', stage: 'preflight' })
  expect(flowAfter(mark('end', 'write-in-progress'), impl)).toEqual(NO_FLOW)
  expect(flowAfter(mark('end', 'refresh-main-checkout'), { ...impl, phase: 'flow-integrate' })).toEqual(NO_FLOW)
})

test('flowAfter resolves the change name from shell variables set earlier in the command', () => {
  const name = 'kan-924-audit-the-app-for-duplicate-api-storage-postgres'
  const verify = { phase: 'flow-implement', change: name, ticket: 'KAN-924', stage: 'verify' } as const
  // The shape the parent session writes.
  expect(flowAfter(`N=${name}; cd /x; flow stage begin -command '/flow' -stage flow.verify -harness claude-code -session-token mf-x $N`, NO_FLOW)).toEqual(verify)
  expect(flowAfter(`export N="${name}" && flow stage begin -command '/flow' -stage flow.verify \${N}`, NO_FLOW)).toEqual(verify)
  expect(flowAfter(`N='${name}'; flow stage begin -command '/flow' -stage flow.verify "$N"`, NO_FLOW)).toEqual(verify)
})

test('flowAfter never lets a name that is no plain change name replace the known change', () => {
  const impl = { phase: 'flow-implement', change: 'kan-873-port-guards', ticket: 'KAN-873', stage: 'verify' } as const
  expect(flowAfter(`flow stage begin -command '/flow' -stage flow.verify -session-token mf-x $N`, impl)).toEqual(impl)
  expect(flowAfter(`flow stage begin -command '/flow' -stage flow.verify \${N}`, impl)).toEqual(impl)
  expect(flowAfter(`M=x; flow stage begin -command '/flow' -stage flow.verify $N`, impl)).toEqual(impl)
  expect(flowAfter(`flow stage begin -command '/flow' -stage flow.verify a/b`, impl)).toEqual(impl)
  expect(flowAfter(`flow stage begin -command '/flow' -stage flow.preflight $N`, impl)).toEqual({ ...impl, phase: 'flow-integrate', stage: 'preflight' })
})

// No leading "· " (the engine puts " · " before a tail), no "->", " │ " between the main agent's part and
// the subagent tally, and the plan's task count after the phase.
test('hintTail: the ticket, the phase and the task count, " │ ", then the tally', () => {
  const impl = { phase: 'flow-implement', change: 'kan-873-port-guards', ticket: 'KAN-873', stage: 'sdd-tdd' } as const
  expect(hintTail(NO_FLOW, '', '')).toBeUndefined()
  expect(hintTail(NO_FLOW, '', '✅ 5')).toBe('✅ 5')
  expect(hintTail({ ...NO_FLOW, phase: 'flow-plan' }, '', '')).toBe('flow-plan')
  expect(hintTail({ ...impl, ticket: null }, '', '✅ 5')).toBe('flow-implement │ ✅ 5')
  expect(hintTail(impl, '', '✅ 5')).toBe('KAN-873 flow-implement │ ✅ 5')
  expect(hintTail(impl, '19/25', '👀 1  ✅ 4')).toBe('KAN-873 flow-implement 19/25 │ 👀 1  ✅ 4')
  expect(hintTail(impl, '19/25', '')).toBe('KAN-873 flow-implement 19/25')
})

test('taskCount counts the ticked column-0 tasks over all of them, never a step checkbox', () => {
  expect(taskCount(PLAN)).toBe('19/25')
  expect(taskCount('- [x] 1. One\n- [ ] 2. Two\n  - [x] **Step 1**\n')).toBe('1/2')
  expect(taskCount('# no tasks yet\n')).toBe('')
  expect(taskCount('')).toBe('')
})

test("the hint line shows the plan's X/N and a visual-verify stage's verifier once, as a review", async ($, on) => {
  mock.clock(on)
  on('session.root', () => ({ value: '/u/Projects/agents' }))
  on('agent.spawn', () => ({ model: 'opus', agentId: 'ag1' }))
  on('tool.call', () => ({ result: '' }))
  on('fs.read', (_$, e) => {
    if (e.path !== '/u/Projects/agents-worktrees/kan-873-port-guards/spectre/changes/kan-873-port-guards/tasks.md') throw new Error('ENOENT')
    return { value: PLAN }
  })
  const tails: (string | undefined)[] = []
  on('ui.render', (r, e) => {
    if (e.component === 'PromptHint') tails.push(e.props.tail)
    const { Text } = r.ui.resolve(e)
    return <Text>hint</Text>
  })
  await $.tool.call({ tool: 'Bash', command: mark('begin', 'visual-verify') })
  await $.agent.spawn(spawn('visual-verify login page'))
  await (await $.ui.mount({ plugin: 'flow-task-list', component: 'PromptHint', props: { isDraft: false, isWorking: true, hint: '' }, surface: 'terminal' })).unmount()
  expect(tails).toEqual(['KAN-873 flow-implement 19/25 │ 🔍 1'])
})

test('taskCount counts a task done when it is ticked or its Task-Id commit has landed', () => {
  expect(taskCount(PLAN, [3, 20, 21])).toBe('21/25')
  expect(taskCount(PLAN, [99])).toBe('19/25')
  expect(landed('20\n\n21\n\n3\n')).toEqual([20, 21, 3])
  expect(landed('')).toEqual([])
})

test('the hint line counts a gated task whose commit landed while its tick waits on its review', async ($, on) => {
  mock.clock(on)
  on('session.root', () => ({ value: '/u/Projects/agents' }))
  on('tool.call', () => ({ result: '' }))
  on('fs.read', () => ({ value: PLAN }))
  const runs: (readonly string[])[] = []
  const out = (exitCode: number, stdout: string) => ({ value: { exitCode, stdout, stderr: '', isStdoutTruncated: false, isStderrTruncated: false } })
  on('process.run', (_$, e) => {
    runs.push(e.argv)
    const verb = e.argv[3]
    return verb === 'branch' ? out(0, 'kan-873-port-guards\n') : verb === 'config' ? out(1, '') : out(0, '21\n\n20\n\n')
  })
  const tails: (string | undefined)[] = []
  on('ui.render', (r, e) => {
    if (e.component === 'PromptHint') tails.push(e.props.tail)
    const { Text } = r.ui.resolve(e)
    return <Text>hint</Text>
  })
  await $.tool.call({ tool: 'Bash', command: mark('begin', 'sdd-tdd') })
  await (await $.ui.mount({ plugin: 'flow-task-list', component: 'PromptHint', props: { isDraft: false, isWorking: true, hint: '' }, surface: 'terminal' })).unmount()
  expect(tails).toEqual(['KAN-873 flow-implement 21/25'])
  const dir = '/u/Projects/agents-worktrees/kan-873-port-guards'
  expect(runs).toEqual([
    ['git', '-C', dir, 'branch', '--show-current'],
    ['git', '-C', dir, 'config', '--get', 'branch.kan-873-port-guards.flowBase'],
    ['git', '-C', dir, 'log', '--format=%(trailers:key=Task-Id,valueonly)', 'refs/remotes/origin/HEAD..HEAD'],
  ])
})

test('the base is the recorded flowBase, its origin ref first, else the bare name', async () => {
  const at = async (originRef: boolean) => {
    const runs: string[][] = []
    const got = await landedIn(async argv => {
      runs.push(argv)
      const verb = argv[3]
      const ok = verb === 'branch' || verb === 'config' || verb === 'log' || originRef
      return { exitCode: ok ? 0 : 1, stdout: verb === 'branch' ? 'c\n' : verb === 'config' ? 'feat\n' : verb === 'log' ? '4\n' : '' }
    }, '/w')
    return [got, runs.at(-1)?.at(-1)]
  }
  expect(await at(true)).toEqual([[4], 'refs/remotes/origin/feat..HEAD'])
  expect(await at(false)).toEqual([[4], 'feat..HEAD'])
  // Git cannot say: nothing has landed, so the count falls back to the ticks.
  expect(await landedIn(async () => ({ exitCode: 128, stdout: '' }), '/w')).toEqual([])
})

test('a landed task waiting on its tick is drawn by no pending row', async ($, on) => {
  mock.clock(on)
  on('session.root', () => ({ value: '/u/Projects/agents' }))
  on('agent.spawn', () => ({ model: 'opus', agentId: 'ag1' }))
  on('tool.call', () => ({ result: '' }))
  on('fs.read', () => ({ value: PLAN }))
  on('process.run', (_$, e) => ({
    value: { exitCode: e.argv[3] === 'config' ? 1 : 0, stdout: e.argv[3] === 'log' ? '20\n\n21\n' : 'c\n', stderr: '', isStdoutTruncated: false, isStderrTruncated: false },
  }))
  on('ui.render', (r, e) => {
    const { Box } = r.ui.resolve(e)
    return <Box />
  })
  await $.tool.call({ tool: 'Bash', command: mark('begin', 'sdd-tdd') })
  await $.agent.spawn(spawn('Task 22/25 (Step 22)'))
  const ui = await $.ui.mount({ ...BAND, surface: 'terminal' })
  const texts = (await ui.findAll({ type: 'Text' })).filter(t => /^(?:⎿ | {2})\S.* — \S/.test(t.text)).map(t => t.text)
  await ui.unmount()
  expect(texts).toEqual([
    '⎿ ◻ main: sdd-tdd — waiting',
    '  ⎿ ◼ Task 22/25 (Step 22) — in progress · 0s',
    '  ⎿ ◻ Task 23/25 (Step 23) — pending',
    '  ⎿ ◻ Task 24/25 (Step 24) — pending',
    '  ⎿ ◻ Task 25/25 (Step 25) — pending',
  ])
})

test('a running row shows its kind as its state word: in progress, review, fix or re-review; a finished one its state', () => {
  const at = (desc: string, state: 'in progress' | 'done' = 'in progress') => line({ id: 'x', n: 1, desc, state }, P22)
  expect(at('Tasks 3+4/22 (review)')).toBe('  ⎿ ◼ Task 3+4/22 (review) — review')
  expect(at('panel-1-primary correctness review')).toBe('  ⎿ ◼ panel-1-primary correctness review — review')
  expect(at('panel-fix-1 findings 1-6')).toBe('  ⎿ ◼ panel-fix-1 findings 1-6 — fix')
  expect(at('panel fix round 2')).toBe('  ⎿ ◼ panel fix round 2 — fix')
  // The kind whose word comes first wins.
  expect(at('Review in-run fix 6 (Tasks 27-28)')).toBe('  ⎿ ◼ Review in-run fix 6 (Tasks 27-28) — review')
  expect(at('Fix review findings')).toBe('  ⎿ ◼ Fix review findings — fix')
  expect(at('Faster integrate fix-1 (review findings)')).toBe('  ⎿ ◼ Faster integrate fix-1 (review findings) — fix')
  expect(at('Review faster integrate (items 1,3,5)')).toBe('  ⎿ ◼ Review faster integrate (items 1,3,5) — review')
  // A review past its first round, or one that says so, is a re-review; round 1 is a review, wherever "fix" stands.
  expect(at('Self-review review-2 (re-review after fix-1)')).toBe('  ⎿ ◼ Self-review review-2 (re-review after fix-1) — re-review')
  expect(at('panel-2-primary correctness review')).toBe('  ⎿ ◼ panel-2-primary correctness review — re-review')
  expect(at('pipeline-fix-1-review-2 landed count')).toBe('  ⎿ ◼ pipeline-fix-1-review-2 landed count — re-review')
  expect(at('pipeline-fix-1-review-1 landed count')).toBe('  ⎿ ◼ pipeline-fix-1-review-1 landed count — review')
  expect(at('pipeline-fix-1-fix-1 landed count')).toBe('  ⎿ ◼ pipeline-fix-1-fix-1 landed count — fix')
  expect(at('visual-verify login page')).toBe('  ⎿ ◼ visual-verify login page — review')
  expect(at('visual-verify-2 login page')).toBe('  ⎿ ◼ visual-verify-2 login page — re-review')
  // An implementer, or any other agent: in progress.
  expect(at('Run tests')).toBe('  ⎿ ◼ Run tests — in progress')
  expect(at('strip prefix')).toBe('  ⎿ ◼ strip prefix — in progress')
  expect(at('read latest log')).toBe('  ⎿ ◼ read latest log — in progress')
  expect(at('Task 3/22 (port guard)')).toBe('  ⎿ ◼ Task 3/22 (port guard) — in progress')
  expect(at('Tasks 3+4/22 (review)', 'done')).toBe('  ⎿ ✔ Task 4/22 (review) — done')
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
  const HINT = { plugin: 'flow-task-list', component: 'PromptHint', props: { isDraft: false, isWorking: true, hint: '' } } as const

  await $.tool.call({ tool: 'Bash', command: mark('begin', 'sdd-tdd') })
  await $.agent.spawn(spawn('Task 2/7 (wire band)'))
  const step = $.turn.step({ turnId: 't1', index: 0, model: 'claude-opus-5-5', effort: 'high', messageCount: 1, agentId: 'ag1' })
  for await (const _ of step) {
    // drain
  }
  expect(await shown()).toEqual(['⎿ ◻ main: sdd-tdd — waiting', '  ⎿ opus-high ◼ Task 2/7 (wire band) — in progress · 0s'])

  await $.turn.complete(complete('ag1', 'answer'))
  expect(await shown()).toEqual([])
  await (await $.ui.mount({ ...HINT, surface: 'terminal' })).unmount()
  expect(tails.at(-1)).toBe('KAN-873 flow-implement │ ✅ 1')
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
  await (await $.ui.mount({ plugin: 'flow-task-list', component: 'PromptHint', props: { isDraft: false, isWorking: true, hint: '' }, surface: 'terminal' })).unmount()
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
    expect(await at('Explore auth')).toMatchObject({ dimColor: true, strikethrough: true, bold: false })
    expect(await at('◼ ')).toMatchObject({ color: 'claude' })
    expect(await at('Read diff')).toMatchObject({ bold: true, dimColor: false, strikethrough: false })
    expect(await at('✘ ')).toMatchObject({ color: 'error' })
    expect(await at('Run tests')).toMatchObject({ bold: false, strikethrough: false })
    await ui.unmount()
  }
})

test("the main agent's row is always first, never a Bash command's description, and waits between its turns", async ($, on) => {
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
  expect(await shown()).toEqual(['⎿ opus-high ◼ main — in progress'])

  await $.agent.spawn(spawn('Explore auth'))
  await drain('ag1', 0, 'claude-sonnet-5-5')
  expect(await shown()).toEqual(['⎿ opus-high ◼ main — in progress', '  ⎿ sonnet ◼ Explore auth — in progress · 0s'])

  // The main turn ends while its background agent runs: the main row stays, waiting.
  await $.turn.complete({ ...complete('ag1', 'answer'), agentId: undefined })
  expect(await shown()).toEqual(['⎿ opus-high ◻ main — waiting', '  ⎿ sonnet ◼ Explore auth — in progress · 0s'])
  await $.turn.complete(complete('ag1', 'answer'))
  expect(await shown()).toEqual([])
})

const PLAN = [
  '# Plan',
  '',
  ...Array.from({ length: 25 }, (_, i) => `- [${i < 19 ? 'x' : ' '}] ${i + 1}. Step ${i + 1}\n  - [ ] a step checkbox, not a task`),
].join('\n')

test('pendingRows lists the plan tasks neither ticked nor landed that no row has taken', () => {
  const taken = [{ id: 'a', n: 1, desc: 'Task 20/25 (Step 20)', state: 'in progress' }, { id: 'b', n: 2, desc: 'Tasks 21+22/25 (pair)', state: 'done' }] as const
  expect(pendingRows(PLAN, [...taken]).map(r => line(r, planTasks(PLAN)))).toEqual([
    '  ⎿ ◻ Task 23/25 (Step 23) — pending',
    '  ⎿ ◻ Task 24/25 (Step 24) — pending',
    '  ⎿ ◻ Task 25/25 (Step 25) — pending',
  ])
  // Task 24's commit has landed while its tick waits on review: counted done, so never drawn pending.
  expect(pendingRows(PLAN, [...taken], [24]).map(r => r.id)).toEqual(['pending-23', 'pending-25'])
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
  expect(await shown()).toEqual(['⎿ ◻ main — waiting', '  ⎿ ◼ Explore auth — in progress · 0s'])
  expect(reads).toEqual([])

  await $.tool.call({ tool: 'Bash', command: mark('begin', 'sdd-tdd') })
  await $.agent.spawn(spawn('Task 20/25 (Step 20)'))
  await $.agent.spawn(spawn('Tasks 21+22/25 (pair)'))
  expect(await shown()).toEqual([
    '⎿ ◻ main: sdd-tdd — waiting',
    '  ⎿ ◼ Explore auth — in progress · 0s',
    '  ⎿ ◼ Task 20/25 (Step 20) — in progress · 0s',
    '  ⎿ ◼ Task 21+22/25 (pair) — in progress · 0s',
    '  ⎿ ◻ Task 23/25 (Step 23) — pending',
    '  ⎿ ◻ Task 24/25 (Step 24) — pending',
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
    '⎿ opus ◼ main: sdd-tdd — in progress',
    '  ⎿ ◼ Task 1/25 (Step 1) — in progress · 0s',
    '  ⎿ ◼ Task 2/25 (Step 2) — in progress · 0s',
    '  ⎿ ◼ Task 3/25 (Step 3) — in progress · 0s',
    '  ⎿ ◻ Task 4/25 (Step 4) — pending',
    '  ⎿ ◻ Task 5/25 (Step 5) — pending',
  ])

  for (const k of [4, 5, 6, 7]) await $.agent.spawn(spawn(`Task ${k}/25 (Step ${k})`))
  expect(await shown()).toEqual([
    '⎿ opus ◼ main: sdd-tdd — in progress',
    ...[1, 2, 3, 4, 5].map(k => `  ⎿ ◼ Task ${k}/25 (Step ${k}) — in progress · 0s`),
  ])
})

test('done rows show only once no task is pending: the last two to finish, above the running rows', async ($, on) => {
  mock.clock(on)
  on('session.root', () => ({ value: '/u/Projects/agents' }))
  let n = 0
  on('agent.spawn', () => ({ model: 'opus', agentId: `ag${++n}` }))
  on('turn.step', async function* (_$, e) {
    return { turnId: e.turnId, index: e.index, answer: '', toolUses: [], stopReason: 'end_turn', usage: null }
  })
  on('turn.complete', () => ({ text: '' }))
  on('tool.call', () => ({ result: '' }))
  on('fs.read', () => ({ value: PLAN }))
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
  for (const k of [1, 2, 3, 4, 5]) await $.agent.spawn(spawn(`Review ${k}`))
  for (const k of [1, 2, 3, 4, 5]) await $.turn.complete(complete(`ag${k}`, 'answer'))

  // Five done rows and six pending tasks: no done row, the pending tasks take all five slots.
  expect(await shown()).toEqual([
    '⎿ opus ◼ main: sdd-tdd — in progress',
    ...[20, 21, 22, 23, 24].map(k => `  ⎿ ◻ Task ${k}/25 (Step ${k}) — pending`),
  ])

  const done = (k: number, end: number) => ({ id: `d${k}`, n: k, desc: `D${k}`, state: 'done', end }) as const
  const pend = (k: number) => ({ id: `pending-${k}`, n: k, desc: `Task ${k}/25`, state: 'pending' }) as const
  expect(band([done(1, 1), running(2), done(3, 3), done(4, 2)], [pend(5), pend(6)]).map(r => r.id)).toEqual(['r2', 'pending-5', 'pending-6'])
  // No pending row: the last two to finish, in the order they finished, above the running rows.
  expect(band([done(1, 1), running(2), done(3, 3), done(4, 2)], []).map(r => r.id)).toEqual(['d4', 'd3', 'r2'])
  expect(band([running(2)], []).map(r => r.id)).toEqual(['r2'])
  // Four running and no plan: one done row fits, and no running row is cut for a done one.
  expect(band([done(1, 1), done(2, 2), running(3), running(4), running(5), running(6)], []).map(r => r.id)).toEqual(['d2', 'r3', 'r4', 'r5', 'r6'])
  expect(band([done(1, 1), ...[2, 3, 4, 5, 6].map(running)], []).map(r => r.id)).toEqual(['r2', 'r3', 'r4', 'r5', 'r6'])
  // Done rows are the marked change's only; with none marked, those done since the main turn began.
  expect(band([{ ...done(1, 1), change: 'a' }, { ...done(2, 2), change: 'b' }], [], 'b').map(r => r.id)).toEqual(['d2'])
  expect(band([done(1, 1), done(2, 5)], [], null, 3).map(r => r.id)).toEqual(['d2'])
  // A failed row stays among the running ones, pending tasks or not, until the next main turn begins.
  const failed = { id: 'f3', n: 3, desc: 'F3', state: 'blocked', end: 4 } as const
  expect(band([running(2), failed], [pend(5)], null, 4).map(r => r.id)).toEqual(['r2', 'f3', 'pending-5'])
  expect(band([running(2), failed], [pend(5)], null, 5).map(r => r.id)).toEqual(['r2', 'pending-5'])
})

test('kind: a plan-task row is a review only when its words say exactly that; a re-run after a fix is a re-review', () => {
  const at = (desc: string, state: 'in progress' | 'done' = 'in progress') => line({ id: 'x', n: 1, desc, state }, P22)
  expect(at('Task 20/22 (In-run fix 1 — visual-verify: drawer)')).toBe('  ⎿ ◼ Task 20/22 (In-run fix 1 — visual-verify: drawer) — in progress')
  expect(at('Task 5/22 (Review panel copy)')).toBe('  ⎿ ◼ Task 5/22 (Review panel copy) — in progress')
  expect(at('Task 5/22 (Visual-verify login)')).toBe('  ⎿ ◼ Task 5/22 (Visual-verify login) — in progress')
  expect(at('Tasks 3+4/22 (re-review)')).toBe('  ⎿ ◼ Task 3+4/22 (re-review) — re-review')
  expect(at('visual-verify-fix-1 login page')).toBe('  ⎿ ◼ visual-verify-fix-1 login page — re-review')
  expect(at('visual-verify-fix-2-full login page')).toBe('  ⎿ ◼ visual-verify-fix-2-full login page — re-review')
  expect(at('task-3+4-reviewer-fix-1 drawer')).toBe('  ⎿ ◼ task-3+4-reviewer-fix-1 drawer — re-review')
  expect(at('panel-fix-1 findings 1-6')).toBe('  ⎿ ◼ panel-fix-1 findings 1-6 — fix')
  // Once done, a group shows its highest task over its own n, a plan loaded or not.
  expect(line({ id: 'g', n: 1, desc: 'Tasks 1+2+3/10 (last three)', state: 'done' }, [])).toBe('  ⎿ ✔ Task 3/10 (last three) — done')
})

test('the band scopes done rows to the marked change, else the main turn, and keeps a failed row until the next main turn', async ($, on) => {
  const clock = mock.clock(on)
  on('session.root', () => ({ value: '/u/Projects/agents' }))
  let n = 0
  on('agent.spawn', () => ({ model: 'opus', agentId: `ag${++n}` }))
  on('turn.complete', () => ({ text: '' }))
  on('tool.call', () => ({ result: '' }))
  let plan = P10_PLAN([1, 2, 3, 4, 5, 6, 7, 8, 9, 10])
  on('fs.read', () => ({ value: plan }))
  on('process.run', (_$, e) => ({
    value: { exitCode: e.argv[3] === 'config' ? 1 : 0, stdout: e.argv[3] === 'log' ? '' : 'c\n', stderr: '', isStdoutTruncated: false, isStderrTruncated: false },
  }))
  on('turn.step', async function* (_$, e) {
    return { turnId: e.turnId, index: e.index, answer: '', toolUses: [], stopReason: 'end_turn', usage: null }
  })
  on('ui.render', (r, e) => {
    const { Box } = r.ui.resolve(e)
    return <Box />
  })
  const mainTurn = async () => {
    await clock.advance(1000)
    for await (const _ of $.turn.step({ turnId: 't1', index: 0, model: 'claude-opus-5-5', messageCount: 1 })) {
      // drain
    }
  }

  // No change marked: a row done in an earlier main turn never reappears.
  await mainTurn()
  await $.agent.spawn(spawn('Old run'))
  await $.turn.complete(complete('ag1', 'answer'))
  await mainTurn()
  await $.agent.spawn(spawn('This turn'))
  await $.turn.complete(complete('ag2', 'answer'))
  await $.agent.spawn(spawn('Running'))
  expect(await bandOf($)).toEqual(['⎿ opus ◼ main — in progress', '  ⎿ ✔ This turn — done · 0s', '  ⎿ ◼ Running — in progress · 0s'])

  // A change marked: only its own done rows, whichever main turn they finished in.
  await $.tool.call({ tool: 'Bash', command: mark('begin', 'sdd-tdd', 'kan-1-x') })
  await $.agent.spawn(spawn('Task 9/10 (Step 9)'))
  await $.turn.complete(complete('ag4', 'answer'))
  await mainTurn()
  expect(await bandOf($)).toEqual(['⎿ opus ◼ main: sdd-tdd — in progress', '  ⎿ ✔ Task 9/10 (Step 9) — done · 0s', '  ⎿ ◼ Running — in progress · 1s'])

  // A failed task stays visible, blocked, among the running rows while tasks are pending — then pending again.
  plan = P10_PLAN([1, 2, 3, 4, 5, 6, 7])
  await $.agent.spawn(spawn('Task 8/10 (Step 8)'))
  await $.turn.complete(complete('ag5', 'error'))
  expect(await bandOf($)).toEqual([
    '⎿ opus ◼ main: sdd-tdd — in progress',
    '  ⎿ ◼ Running — in progress · 1s',
    '  ⎿ ✘ Task 8/10 (Step 8) — blocked · 0s',
    '  ⎿ ◻ Task 9/10 (Step 9) — pending',
    '  ⎿ ◻ Task 10/10 (Step 10) — pending',
  ])
  await mainTurn()
  expect(await bandOf($)).toEqual([
    '⎿ opus ◼ main: sdd-tdd — in progress',
    '  ⎿ ◼ Running — in progress · 2s',
    '  ⎿ ◻ Task 8/10 (Step 8) — pending',
    '  ⎿ ◻ Task 9/10 (Step 9) — pending',
    '  ⎿ ◻ Task 10/10 (Step 10) — pending',
  ])
})

test("a subagent row's elapsed time follows its state", () => {
  const P8 = [1, 2, 3, 4, 5, 6, 7, 8]
  const row = { id: 'a', n: 3, desc: 'Task 3/8 (desc)', state: 'in progress', start: 1000 } as const
  expect(line(row, P8, 'opus-medium', 11, 1000 + 252_000)).toBe('  ⎿ opus-medium ◼ Task 3/8 (desc) — in progress · 4m12s')
  expect(line({ ...row, state: 'done', end: 1000 + 3_600_000 + 65_000 }, P8, '', 0, 9e9)).toBe('  ⎿ ✔ Task 3/8 (desc) — done · 1h01m')
  expect(line(row, P8, '', 0, 1000)).toBe('  ⎿ ◼ Task 3/8 (desc) — in progress · 0s')
  expect(line({ id: 'p', n: 4, desc: 'Task 4/8 (later)', state: 'pending' }, P8, '', 0, 5000)).toBe('  ⎿ ◻ Task 4/8 (later) — pending')
})

test('a running row ticks and shows no token count; a finished row freezes', async ($, on) => {
  const clock = mock.clock(on)
  let n = 0
  on('agent.spawn', () => ({ model: 'opus', agentId: `ag${++n}` }))
  on('turn.complete', () => ({ text: '' }))
  const usage = { model: 'claude-opus-5-5', input_tokens: 130_000, output_tokens: 3_000, cache_read_input_tokens: 1_000_000, cache_creation_input_tokens: 100_000 }
  on('turn.step', async function* (_$, e) {
    return { turnId: e.turnId, index: e.index, answer: '', toolUses: [], stopReason: 'end_turn', usage }
  })
  on('ui.render', (r, e) => {
    const { Box } = r.ui.resolve(e)
    return <Box />
  })
  const rows = async (ui: { findAll: (q: { type: 'Text' }) => Promise<{ text: string }[]> }) =>
    (await ui.findAll({ type: 'Text' })).filter(t => /^(?:⎿ | {2})\S.* — \S/.test(t.text)).map(t => t.text)

  await $.agent.spawn(spawn('Task 3/8 (desc)'))
  for await (const _ of $.turn.step({ turnId: 't1', index: 0, model: 'claude-opus-5-5', effort: 'medium', messageCount: 1, agentId: 'ag1' })) {
    // drain
  }
  const ui = await $.ui.mount({ ...BAND, surface: 'terminal' })
  expect(await rows(ui)).toEqual(['⎿ ◻ main — waiting', '  ⎿ opus-medium ◼ Task 3/8 (desc) — in progress · 0s'])
  await clock.advance(252_000)
  expect(await rows(ui)).toEqual(['⎿ ◻ main — waiting', '  ⎿ opus-medium ◼ Task 3/8 (desc) — in progress · 4m12s'])
  await ui.unmount()

  await $.agent.spawn(spawn('Task 4/8 (other)'))
  await $.turn.complete(complete('ag1', 'answer'))
  await clock.advance(60_000)
  const after = await $.ui.mount({ ...BAND, surface: 'terminal' })
  expect((await rows(after))[1]).toBe('  ⎿ opus-medium ✔ Task 3/8 (desc) — done · 4m12s')
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
    '⛔ Visual verify (final full run) — blocked. Two new failing specs.',
    '✅ Visual verify (final): Quick mode — done. Clean.',
    '🔍 Flow pipeline change — in review again, after the fixes',
    '✅ Task 6/8 — done, landed',
    '⛔ Task 7/8 (gate) — blocked: needs a decision',
    '🔄 Task 5/8 (port — pending bits) — in progress',
  ].join('\n')
  expect(statusLines(text)).toEqual([
    ['Full backend test suite', 'pending'],
    ['Tasks 23–25/25 (e2e, fidelity, live)', 'in progress'],
    ['Review panel (primary + principles)', 'in review'],
    ['Task 4/8 (port)', 'blocked'],
    ['Task 22/22 (spec text)', 'done'],
    ['Visual verify (final full run)', 'blocked'],
    ['Visual verify (final): Quick mode', 'done'],
    ['Flow pipeline change', 'in review'],
    ['Task 6/8', 'done'],
    ['Task 7/8 (gate)', 'blocked'],
    ['Task 5/8 (port — pending bits)', 'in progress'],
  ])
})

test('unitKey: one key for a unit written two ways', () => {
  expect(unitKey('Visual verify (final full run)')).toBe('Visual verify')
  expect(unitKey('Visual verify (final): Quick mode')).toBe('Visual verify')
  expect(unitKey('Flow pipeline change review')).toBe('Flow pipeline change')
  expect(unitKey('Flow pipeline change (narrow re-verify in the fix loop)')).toBe('Flow pipeline change')
  expect(unitKey('Tasks 23–25/25 (e2e)')).toBe('Tasks 23–25/25')
  expect(unitKey('Full backend test suite')).toBe('Full backend test suite')
})


test('mainLabel: the inline plan task, else the /flow stage, else the latest unit of its own, else main', () => {
  const P10 = Array.from({ length: 10 }, (_, i) => i + 1)
  const l = (unit: string, state: 'in progress' | 'in review' = 'in progress') => [unitKey(unit), { unit, state }] as const
  const subs = [
    { id: 'a', n: 1, desc: 'Review faster integrate (items 1,3,5)', state: 'done' },
    { id: 'b', n: 2, desc: 'Faster integrate fix-1 (review findings)', state: 'in progress' },
    { id: 'c', n: 3, desc: 'Tasks 2+3/10 (pair)', state: 'in progress' },
  ] as const
  const screenshot = Object.fromEntries([
    l('Self-review review-2 (re-review after fix-1)', 'in review'),
    l('Items 1, 3, 5 (one integrate script, …)'),
    l('Faster integrate fix-1 (4 Important, 5 Minor review findings)'),
  ])
  // The line about the running fix is that subagent's; the latest of the main agent's own is the label.
  expect(mainLabel(screenshot, [...subs], null, [])).toBe('Items 1, 3, 5 (one integrate script, …)')
  expect(mainLabel(screenshot, [...subs], 'review-panel', [])).toBe('review-panel')
  // A plan task a subagent carries is not the main agent's; one no subagent names is, ahead of the stage.
  const inline = Object.fromEntries([l('Task 4/10 (Step 4)'), l('Task 3/10 (pair)'), l('Full suite')])
  expect(mainLabel(inline, [...subs], 'sdd-tdd', P10)).toBe('Task 4/10 (Step 4)')
  expect(mainLabel(Object.fromEntries([l('Task 3/10 (pair)')]), [...subs], 'sdd-tdd', P10)).toBe('sdd-tdd')
  expect(mainLabel({}, [], null, P10)).toBe('main')
})

// The band's rows as the terminal draws them, for the fixtures below.
const bandOf = async ($: Parameters<Parameters<typeof test>[1]>[0]) => {
  const ui = await $.ui.mount({ ...BAND, surface: 'terminal' })
  const texts = (await ui.findAll({ type: 'Text' })).filter(t => /^(?:⎿ | {2}) *\S.* — \S/.test(t.text)).map(t => t.text)
  await ui.unmount()
  return texts
}

test("the operator's screenshot: one main row, the running fix, the last done review; status lines are no rows", async ($, on) => {
  const clock = mock.clock(on)
  let n = 0
  on('agent.spawn', () => ({ model: 'opus', agentId: `ag${++n}` }))
  on('turn.complete', () => ({ text: '' }))
  let answer = ''
  on('turn.step', async function* (_$, e) {
    return { turnId: e.turnId, index: e.index, answer, toolUses: [], stopReason: 'end_turn', usage: null }
  })
  on('ui.render', (r, e) => {
    const { Box } = r.ui.resolve(e)
    return <Box />
  })
  const step = async (index: number, agentId?: string, text = '') => {
    answer = text
    for await (const _ of $.turn.step({ turnId: 't1', index, model: 'claude-opus-5-5', effort: 'high', messageCount: 1, agentId })) {
      // drain
    }
  }

  await step(0)
  await step(1, undefined, '🔍 Self-review review-2 (re-review after fix-1) — in review')
  await $.agent.spawn(spawn('Review faster integrate (items 1,3,5)'))
  await step(0, 'ag1')
  await step(2, undefined, '🔄 Items 1, 3, 5 (one integrate script, …) — in progress')
  await clock.advance(465_000)
  await $.turn.complete(complete('ag1', 'answer'))
  await $.agent.spawn(spawn('Faster integrate fix-1 (review findings)'))
  await step(0, 'ag2')
  await step(3, undefined, '🔄 Faster integrate fix-1 (4 Important, 5 Minor review findings) — in progress')
  await clock.advance(105_000)
  expect(await bandOf($)).toEqual([
    '⎿ opus-high ◼ main: Items 1, 3, 5 (one integrate script, …) — in progress',
    '  ⎿ opus-high ✔ Review faster integrate (items 1,3,5) — done · 7m45s',
    '  ⎿ opus-high ◼ Faster integrate fix-1 (review findings) — fix · 1m45s',
  ])
  // Between turns, the background fix still running: the main row stays, waiting.
  await $.turn.complete({ ...complete('x', 'answer'), agentId: undefined })
  expect((await bandOf($))[0]).toBe('⎿ opus-high ◻ main: Items 1, 3, 5 (one integrate script, …) — waiting')
  // A unit written again is the latest; a done line hands the label back to the one before.
  await step(0, undefined, '🔍 Self-review review-2 (re-review after fix-1) — in review')
  expect((await bandOf($))[0]).toBe('⎿ opus-high ◼ main: Self-review review-2 (re-review after fix-1) — in progress')
  await step(1, undefined, '✅ Self-review review-2 — done')
  expect((await bandOf($))[0]).toBe('⎿ opus-high ◼ main: Items 1, 3, 5 (one integrate script, …) — in progress')
})

const P10_PLAN = (ticked: number[]) =>
  ['# Plan', '', ...Array.from({ length: 10 }, (_, i) => `- [${ticked.includes(i + 1) ? 'x' : ' '}] ${i + 1}. Step ${i + 1}`)].join('\n')

test('a plan with pending tasks: the inline task on main, the running agents, the pending tasks below, no done row', async ($, on) => {
  mock.clock(on)
  on('session.root', () => ({ value: '/u/Projects/agents' }))
  let n = 0
  on('agent.spawn', () => ({ model: 'opus', agentId: `ag${++n}` }))
  on('turn.complete', () => ({ text: '' }))
  on('tool.call', () => ({ result: '' }))
  on('fs.read', () => ({ value: P10_PLAN([]) }))
  let answer = ''
  on('turn.step', async function* (_$, e) {
    return { turnId: e.turnId, index: e.index, answer, toolUses: [], stopReason: 'end_turn', usage: null }
  })
  on('ui.render', (r, e) => {
    const { Box } = r.ui.resolve(e)
    return <Box />
  })
  const step = async (index: number, agentId?: string, text = '') => {
    answer = text
    for await (const _ of $.turn.step({ turnId: 't1', index, model: 'claude-opus-5-5', effort: 'high', messageCount: 1, agentId })) {
      // drain
    }
  }

  await step(0)
  await $.tool.call({ tool: 'Bash', command: mark('begin', 'sdd-tdd', 'kan-1-x') })
  await $.agent.spawn(spawn('Task 1/10 (Step 1)'))
  await step(0, 'ag1')
  await $.turn.complete(complete('ag1', 'answer'))
  await $.agent.spawn(spawn('Tasks 2+3/10 (pair)'))
  await step(0, 'ag2')
  await $.agent.spawn(spawn('Task 1/10 (review)'))
  await step(0, 'ag3')
  await step(1, undefined, ['🔄 Task 1/10 (Step 1) — in progress', '🔄 Tasks 2+3/10 (pair) — in progress', '🔄 Task 4/10 (Step 4) — in progress'].join('\n'))
  expect(await bandOf($)).toEqual([
    '⎿ opus-high ◼ main: Task 4/10 (Step 4) — in progress',
    '  ⎿ opus-high ◼ Task 2+3/10 (pair) — in progress · 0s',
    '  ⎿ opus-high ◼ Task 1/10 (review) — review · 0s',
    '  ⎿           ◻ Task 5/10 (Step 5) — pending',
    '  ⎿           ◻ Task 6/10 (Step 6) — pending',
    '  ⎿           ◻ Task 7/10 (Step 7) — pending',
  ])
})

test('no task pending: the last two done above the running rows, and a finished group shows its highest task', async ($, on) => {
  mock.clock(on)
  on('session.root', () => ({ value: '/u/Projects/agents' }))
  let n = 0
  on('agent.spawn', () => ({ model: 'opus', agentId: `ag${++n}` }))
  on('turn.complete', () => ({ text: '' }))
  on('tool.call', () => ({ result: '' }))
  on('fs.read', () => ({ value: P10_PLAN([1, 2, 3, 4, 5, 6, 7]) }))
  let log = ''
  on('process.run', (_$, e) => ({
    value: { exitCode: e.argv[3] === 'config' ? 1 : 0, stdout: e.argv[3] === 'log' ? log : 'c\n', stderr: '', isStdoutTruncated: false, isStderrTruncated: false },
  }))
  on('turn.step', async function* (_$, e) {
    return { turnId: e.turnId, index: e.index, answer: '', toolUses: [], stopReason: 'end_turn', usage: null }
  })
  on('ui.render', (r, e) => {
    const { Box } = r.ui.resolve(e)
    return <Box />
  })
  for await (const _ of $.turn.step({ turnId: 't1', index: 0, model: 'claude-opus-5-5', effort: 'high', messageCount: 1 })) {
    // drain
  }
  await $.tool.call({ tool: 'Bash', command: mark('begin', 'sdd-tdd', 'kan-1-x') })
  for (const k of [5, 6, 7]) await $.agent.spawn(spawn(`Task ${k}/10 (Step ${k})`))
  for (const k of [1, 2, 3]) await $.turn.complete(complete(`ag${k}`, 'answer'))
  await $.agent.spawn(spawn('Tasks 8+9+10/10 (last three)'))
  expect(await bandOf($)).toEqual([
    '⎿ opus-high ◼ main: sdd-tdd — in progress',
    '  ⎿ ✔ Task 6/10 (Step 6) — done · 0s',
    '  ⎿ ✔ Task 7/10 (Step 7) — done · 0s',
    '  ⎿ ◼ Task 8+9+10/10 (last three) — in progress · 0s',
  ])
  log = '8+9+10\n'
  await $.turn.complete(complete('ag4', 'answer'))
  expect(await bandOf($)).toEqual(['⎿ opus-high ◼ main: sdd-tdd — in progress', '  ⎿ ✔ Task 7/10 (Step 7) — done · 0s', '  ⎿ ✔ Task 10/10 (last three) — done · 0s'])
})
