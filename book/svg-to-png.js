// Render every SVG in figures/ to a high-resolution PNG (3x) via Playwright Chromium.
const fs = require('fs');
const path = require('path');
const { chromium } = require('C:/cgo/web/frontend/node_modules/playwright');

const DIR = path.join(__dirname, 'figures');
const SCALE = 3;

(async () => {
  const svgs = fs.readdirSync(DIR).filter(f => f.endsWith('.svg'));
  const browser = await chromium.launch({ channel: 'msedge' });
  const page = await browser.newPage({ deviceScaleFactor: SCALE });
  for (const f of svgs) {
    const svg = fs.readFileSync(path.join(DIR, f), 'utf8');
    const wm = svg.match(/width="(\d+)"/), hm = svg.match(/height="(\d+)"/);
    const w = parseInt(wm[1], 10), h = parseInt(hm[1], 10);
    await page.setViewportSize({ width: w, height: h });
    await page.setContent(`<!DOCTYPE html><html><head><style>*{margin:0;padding:0}</style></head><body>${svg}</body></html>`);
    await page.waitForTimeout(60);
    const out = path.join(DIR, f.replace(/\.svg$/, '.png'));
    await page.screenshot({ path: out, clip: { x: 0, y: 0, width: w, height: h } });
    const kb = Math.round(fs.statSync(out).size / 1024);
    console.log(f, '->', path.basename(out), `${w * SCALE}x${h * SCALE}px`, kb + 'ko');
  }
  await browser.close();
})();
