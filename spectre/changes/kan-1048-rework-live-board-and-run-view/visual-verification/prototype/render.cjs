// Renders the approved prototype (design.md ## Context) as each capture's frame,
// at playwright.config.ts's viewport 1280x1024, dsf 1, dark scheme.
const { chromium } = require(process.argv[2] + "/node_modules/playwright");
const proto = "file://" + process.argv[3];
const out = process.argv[4];
(async () => {
  const b = await chromium.launch({ args: ["--disable-gpu-rasterization", "--disable-partial-raster", "--disable-gpu"] });
  const p = await b.newPage({ viewport: { width: 1280, height: 1024 }, deviceScaleFactor: 1, colorScheme: "dark", reducedMotion: "reduce" });
  const shot = (n) => p.screenshot({ path: `${out}/${n}.png`, fullPage: true });
  await p.goto(proto); await p.waitForSelector(".tile");
  await shot("dashboard"); await shot("full-state-board");
  await p.locator('[data-project="agents"]').click();
  await p.locator('[data-state="STARTED"]').click();
  await p.mouse.move(0, 0); await shot("combination-chips-plus-tile");
  await p.goto(proto + "#run"); await p.waitForSelector("svg .bar");
  await shot("run-detail"); await shot("full-run-detail");
  await p.locator("svg .bar").first().hover(); await p.waitForSelector("#tip:not([hidden])");
  await shot("run-detail-hover");
  await b.close();
})();
