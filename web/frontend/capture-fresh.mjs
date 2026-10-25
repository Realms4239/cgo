// capture-fresh.mjs — captures fraîches 1920×1080 (+1366) du dashboard courant.
// Usage: node capture-fresh.mjs [baseURL]  (défaut http://127.0.0.1:9091)
import { chromium } from 'playwright-core';

const base = process.argv[2] || 'http://127.0.0.1:9091';
const out = process.argv[3] || 'shots-fresh';
const shots = [
  ['campagne', 'campagne-idle', false],
  ['campagne', 'campagne-audit-open', 'audit'],
  ['live', 'live-idle', false],
  ['resultats', 'resultats-idle', false],
  ['integrite', 'integrite-idle', false],
];
const browser = await chromium.launch({ args: ['--no-sandbox'], channel: process.env.PW_CHANNEL || 'msedge' }).catch(() => chromium.launch({ args: ['--no-sandbox'], channel: 'chrome' }));
for (const width of [1920, 1366]) {
  const ctx = await browser.newContext({ viewport: { width, height: 1080 } });
  const page = await ctx.newPage();
  await page.goto(base, { waitUntil: 'domcontentloaded', timeout: 30000 });
  await page.waitForTimeout(3500);
  for (const [panel, name, openAudit] of shots) {
    if (width === 1366 && name !== 'campagne-idle' && name !== 'resultats-idle') continue;
    await page.click(`button[data-panel="${panel}"]`, { timeout: 10000 });
    await page.waitForTimeout(1200);
    if (openAudit) {
      await page.click('button.btn:has-text("LANCER")', { timeout: 5000 }).catch(() => {});
      await page.waitForTimeout(800);
    }
    await page.screenshot({ path: `${out}/${name}-${width}.png` });
    console.log('OK', name, width);
    if (openAudit) {
      await page.click('button.btn:has-text("FERMER")', { timeout: 5000 }).catch(() => {});
    }
  }
  await ctx.close();
}
await browser.close();
