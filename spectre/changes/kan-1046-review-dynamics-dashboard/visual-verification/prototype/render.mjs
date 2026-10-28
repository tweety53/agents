// Renders the approved prototype as each review-dynamics.spec.ts capture's frame,
// full page at the capture project's 1280x1024 viewport, scale 1.
import { chromium } from "@playwright/test";
const [proto, out] = process.argv.slice(2);
const b = await chromium.launch();
async function shot(scheme, name, drive) {
  const p = await b.newPage({ viewport: { width: 1280, height: 1024 }, deviceScaleFactor: 1, colorScheme: scheme });
  await p.goto("file://" + proto);
  if (drive) await drive(p);
  await p.screenshot({ path: `${out}/${name}.png`, fullPage: true });
  await p.close();
}
await shot("dark", "review-dynamics");
await shot("dark", "review-dynamics-version", async (p) => {
  await p.locator("#models button[data-m=version]").click();
  await p.mouse.move(0, 0);
});
// The prototype opens on Last 8 weeks pressed: the default-preset frame.
await shot("dark", "review-dynamics-default-preset");
await shot("dark", "review-dynamics-hover", async (p) => {
  const box = await p.locator("#c1").boundingBox();
  // bucket index 2 of 8, x = 34 + 2*(436/7) in viewBox 480
  await p.mouse.move(box.x + (34 + 2 * 436 / 7) * box.width / 480, box.y + box.height / 2);
});
await shot("dark", "review-dynamics-heat-hover", (p) => p.locator(".heat td").first().hover());
await shot("light", "review-dynamics-light");
await b.close();
