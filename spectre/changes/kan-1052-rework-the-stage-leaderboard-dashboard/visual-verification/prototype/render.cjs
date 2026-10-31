// Renders the approved prototype as each stage-leaderboard.spec.ts capture's frame.
const { chromium } = require("/Users/tweety53/Projects/agents-worktrees/kan-1052-rework-the-stage-leaderboard-dashboard/stats/web/node_modules/@playwright/test");
(async () => {
  const b = await chromium.launch();
  const shots = [["stage-leaderboard-wide", 1280, 1024, false], ["stage-leaderboard-phone", 390, 844, false], ["stage-leaderboard-chip-hover", 1280, 1024, true]];
  for (const [stem, width, height, hover] of shots) {
    const p = await b.newPage({ viewport: { width, height }, deviceScaleFactor: 1, colorScheme: "light" });
    await p.goto("file:///Users/tweety53/Projects/agents-worktrees/kan-1052-rework-the-stage-leaderboard-dashboard/.superpowers/sdd/stage-leaderboard-prototype.html");
    if (hover) await p.getByRole("group", { name: "Project" }).getByRole("button").nth(1).hover();
    await p.screenshot({ path: "/Users/tweety53/Projects/agents-worktrees/kan-1052-rework-the-stage-leaderboard-dashboard/spectre/changes/kan-1052-rework-the-stage-leaderboard-dashboard/visual-verification/prototype/" + stem + ".png", fullPage: true });
    await p.close();
  }
  await b.close();
})();
