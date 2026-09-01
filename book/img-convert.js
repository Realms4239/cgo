// Convert the DGM organigramme JPEG to PNG (fig-org-dgm.png) and copy the
// DEM organigramme PNG into figures/, via Playwright Chromium screenshot.
const fs = require('fs');
const path = require('path');
const { chromium } = require('c:/thesis-cgo/thesis-cgo/web/frontend/node_modules/playwright');

const DIR = path.join(__dirname, 'figures');

(async () => {
  fs.mkdirSync(DIR, { recursive: true });
  const browser = await chromium.launch({ channel: 'msedge' });
  const page = await browser.newPage({ deviceScaleFactor: 2 });
  const jpg = path.join(__dirname, 'Organigramme_DGM-810x740.jpeg');
  await page.goto('file:///' + jpg.replace(/\\/g, '/'));
  const dim = await page.evaluate(() => ({ w: document.body.scrollWidth, h: document.body.scrollHeight }));
  await page.setViewportSize({ width: dim.w, height: dim.h });
  const out = path.join(DIR, 'fig-org-dgm.png');
  await page.screenshot({ path: out });
  fs.copyFileSync(path.join(__dirname, 'ORGANNIGRAME-DEM-Mai-2025-810x740.png'), path.join(DIR, 'fig-org-dem.png'));
  console.log('fig-org-dgm.png', dim.w + 'x' + dim.h, Math.round(fs.statSync(out).size / 1024) + 'ko');
  console.log('fig-org-dem.png copied');
  await browser.close();
})();
