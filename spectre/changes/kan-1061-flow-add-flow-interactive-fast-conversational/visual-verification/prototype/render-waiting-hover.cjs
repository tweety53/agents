const { chromium } = require('/Users/tweety53/Projects/agents-worktrees/kan-1061-flow-add-flow-interactive-fast-conversational/stats/web/node_modules/playwright');
(async () => {
  const b = await chromium.launch();
  const p = await b.newPage({ viewport: { width: 1280, height: 1024 }, deviceScaleFactor: 1, colorScheme: 'dark' });
  await p.goto('file:///Users/tweety53/Projects/agents-worktrees/kan-1061-flow-add-flow-interactive-fast-conversational/.superpowers/sdd/planner/prototype-board.html');
  const c = p.locator('[aria-label="kan-1073-interactive-session-2026-10-30"]');
  await c.hover();
  console.log(await c.evaluate(e => getComputedStyle(e).borderTopColor));
  await p.screenshot({ path: '/Users/tweety53/Projects/agents-worktrees/kan-1061-flow-add-flow-interactive-fast-conversational/.superpowers/sdd/interactive-frames/prototype-waiting-card.png' });
  await b.close();
})();
