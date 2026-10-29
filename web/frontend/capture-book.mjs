// capture-book.mjs — captures 1920×1080 pour l'annexe E du mémoire.
// Usage: node capture-book.mjs http://127.0.0.1:9092 ../book/figures/ui
import { chromium } from 'playwright-core';
import fs from 'fs';

const base = process.argv[2] || 'http://127.0.0.1:9092';
const out = process.argv[3] || '../../book/figures/ui';
fs.mkdirSync(out, { recursive: true });

const browser = await chromium.launch({ args: ['--no-sandbox'], channel: process.env.PW_CHANNEL || 'msedge' })
  .catch(() => chromium.launch({ args: ['--no-sandbox'], channel: 'chrome' }));
const ctx = await browser.newContext({ viewport: { width: 1920, height: 1080 } });
const page = await ctx.newPage();
const top = async () => {
  await page.evaluate(() => {
    window.scrollTo(0, 0);
    document.querySelectorAll('*').forEach(el => { if (el.scrollTop) el.scrollTop = 0; });
  });
  await page.waitForTimeout(300);
};

await page.goto(base, { waitUntil: 'domcontentloaded', timeout: 30000 });
await page.waitForTimeout(5000);

// Campagne — cockpit au repos
await page.click('button[data-panel="campagne"]', { timeout: 10000 });
await page.waitForTimeout(1800);
await top();
await page.screenshot({ path: `${out}/book-campagne.png` });
console.log('OK campagne');

// Résultats — classement sur le corpus gelé
await page.click('button[data-panel="resultats"]', { timeout: 10000 });
await page.waitForSelector('table.data-table tbody tr', { timeout: 20000 }).catch(() => {});
await page.waitForTimeout(3000);
await top();
await page.screenshot({ path: `${out}/book-resultats.png` });
console.log('OK resultats');

// Comparaison — deux cellules épinglées (le duel)
const pins = page.locator('button[aria-label^="comparer"]');
const n = await pins.count();
if (n >= 2) {
  await pins.nth(0).click();
  await page.waitForTimeout(400);
  await page.locator('button[aria-label^="comparer"]').nth(1).click();
  await page.waitForSelector('[data-testid="compare-view"]', { timeout: 15000 }).catch(() => {});
  await page.waitForTimeout(2500);
  const cv = page.locator('[data-testid="compare-view"]').first();
  if (await cv.count()) {
    await cv.scrollIntoViewIfNeeded().catch(() => {});
    await cv.screenshot({ path: `${out}/book-comparaison.png` }).catch(async () => {
      await page.screenshot({ path: `${out}/book-comparaison.png` });
    });
    console.log('OK comparaison');
  }
  await page.keyboard.press('Escape').catch(() => {});
  await page.waitForTimeout(500);
}

// Provenance — preuve gelée (208 valides / 74 dégradées / 246 écartées)
await page.click('button[data-panel="integrite"]', { timeout: 10000 });
await page.waitForSelector('[data-testid="proof-banner"]', { timeout: 20000 }).catch(() => {});
await page.waitForTimeout(2500);
await top();
await page.screenshot({ path: `${out}/book-integrite.png` });
console.log('OK integrite');

// Live — replay d'un run gelé, pas une mesure live
const replay = page.locator('[data-testid="run-rejouer"]').first();
if (await replay.count()) {
  await replay.click();
  await page.waitForTimeout(2000);
  const pass = page.locator('button:has-text("PASSER")').first();
  if (await pass.count()) { await pass.click().catch(() => {}); }
  await page.waitForTimeout(4000);
  await page.keyboard.press('Escape').catch(() => {});
  await page.waitForTimeout(700);
  await top();
  await page.screenshot({ path: `${out}/book-live.png` });
  console.log('OK live replay');
} else {
  await page.click('button[data-panel="live"]', { timeout: 10000 });
  await page.waitForTimeout(1500);
  await page.screenshot({ path: `${out}/book-live.png` });
  console.log('OK live idle');
}

await browser.close();
