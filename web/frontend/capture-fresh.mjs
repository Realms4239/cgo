// capture-fresh.mjs — captures des quatre modes et des trois sections du banc.
// Usage: node capture-fresh.mjs [baseURL] [outDir]  (défaut http://127.0.0.1:9091)
import { chromium } from 'playwright-core';

const base = process.argv[2] || 'http://127.0.0.1:9091';
const out = process.argv[3] || 'shots-fresh';
const shots = [
  ['dashboard', 'tableau-de-bord', null],
  ['audit', 'auditer', null],
  ['bench', 'banc-configurer', 'config'],
  ['bench', 'banc-observer', 'observe'],
  ['bench', 'banc-resultats', 'results'],
  ['evidence', 'preuves', null],
];
const browser = await chromium.launch({ args: ['--no-sandbox'], channel: process.env.PW_CHANNEL || 'msedge' }).catch(() => chromium.launch({ args: ['--no-sandbox'], channel: 'chrome' }));
for (const [width, height] of [[1920, 1080], [1366, 1080], [390, 844]]) {
  const ctx = await browser.newContext({ viewport: { width, height } });
  const page = await ctx.newPage();
  await page.goto(base, { waitUntil: 'domcontentloaded', timeout: 30000 });
  // l'intégrité scanne 159 manifestes : laisser la preuve se peindre avant la capture
  await page.waitForTimeout(9000);
  for (const [mode, name, anchor] of shots) {
    if (width !== 1920 && !['tableau-de-bord', 'banc-resultats'].includes(name)) continue;
    await page.click(`button[data-mode="${mode}"]`, { timeout: 10000 });
    await page.waitForTimeout(mode === 'evidence' ? 5000 : 1600);
    if (anchor === 'config') {
      // le titre du mode reste visible : pas d'ancre, on remonte en haut
      await page.evaluate(() => document.getElementById('main')?.scrollTo(0, 0));
      await page.waitForTimeout(500);
    }
    if (anchor === 'observe') {
      await page.click('.bench-anchors button:has-text("OBSERVER")', { timeout: 5000 }).catch(() => {});
      await page.waitForTimeout(800);
    }
    if (anchor === 'results') {
      await page.click('.bench-anchors button:has-text("RÉSULTATS")', { timeout: 5000 }).catch(() => {});
      await page.waitForSelector('#bench-results .lb-card', { timeout: 20000 }).catch(() => {});
      await page.evaluate(() => document.getElementById('bench-results')?.scrollIntoView({ block: 'start' }));
      await page.waitForTimeout(1500);
    }
    await page.screenshot({ path: `${out}/${name}-${width}.png`, fullPage: false });
    console.log('OK', name, width);
  }
  await ctx.close();
}
await browser.close();
