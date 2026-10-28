const { chromium } = require(process.argv[4]);
(async () => { const b = await chromium.launch(); const p = await b.newPage({ viewport: { width: 1280, height: 1024 }, deviceScaleFactor: 1, colorScheme: process.argv[5]||'dark' });
await p.goto('file://' + process.argv[2]); await p.waitForTimeout(500); await p.screenshot({ path: process.argv[3], fullPage: true }); await b.close(); })();
