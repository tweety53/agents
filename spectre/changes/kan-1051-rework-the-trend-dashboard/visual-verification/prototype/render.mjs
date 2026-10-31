// Renders the approved prototype as each trend capture's frame (visual-verify step 10).
import { createRequire } from "node:module";
const require = createRequire(process.argv[2] + "/package.json");
const { chromium } = require("@playwright/test");
const proto = "file://" + process.argv[3], out = process.argv[4];
const shots = [
  { stem: "trend-wide", vp: { width: 1280, height: 1024 }, scheme: "light" },
  { stem: "trend-phone", vp: { width: 390, height: 844 }, scheme: "light" },
  { stem: "trend-chart-hover", vp: { width: 1280, height: 1024 }, scheme: "light", hover: 8 },
  { stem: "full-trend", vp: { width: 1280, height: 1024 }, scheme: "dark" },
];
const browser = await chromium.launch();
for (const s of shots) {
  const ctx = await browser.newContext({ viewport: s.vp, deviceScaleFactor: 1, colorScheme: s.scheme });
  const page = await ctx.newPage();
  await page.goto(proto);
  if (s.hover != null) { await page.locator(`.hit[data-i="${s.hover}"]`).hover(); await page.locator("#tip").waitFor({ state: "visible" }); }
  await page.screenshot({ path: `${out}/${s.stem}.png`, fullPage: true });
  await ctx.close();
}
await browser.close();
