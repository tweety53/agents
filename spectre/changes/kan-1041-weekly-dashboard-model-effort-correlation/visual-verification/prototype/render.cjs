// Renders the approved prototype as each weekly-costs capture's frame (visual-verify step 10).
// Usage: node render.cjs <playwright module dir> <prototype html> <out dir>
const { chromium } = require(process.argv[2]);
const [proto, out] = process.argv.slice(3);
const states = [
  ["weekly-costs", "dark", []],
  ["weekly-costs-light", "light", []],
  ["weekly-costs-toggled", "dark", [0]],
  ["weekly-costs-model-effort", "dark", [1, 2, 3, 4, 5, 6]],
  ["weekly-costs-tooltip", "dark", [1, 2, 3, 4, 5, 6]],
];
(async () => {
  const browser = await chromium.launch();
  for (const [stem, scheme, off] of states) {
    const page = await browser.newPage({ viewport: { width: 1280, height: 1024 }, colorScheme: scheme });
    await page.goto("file://" + proto);
    for (const i of off) await page.locator("#t0 button").nth(i).click();
    await page.mouse.move(0, 0);
    await page.screenshot({ path: `${out}/${stem}.png`, fullPage: true });
    await page.close();
  }
  await browser.close();
})();
