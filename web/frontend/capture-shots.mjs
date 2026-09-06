// capture-shots.mjs — captures 1920×1080 du dashboard pour l'annexe E.
// Usage: node scripts/capture-shots.mjs [baseURL]  (défaut https://meteolink.dev:9090)
// Recapture après modifs UI, jamais de valeurs saisies : le serveur réel.
import { chromium } from 'playwright-core';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';

const base = process.argv[2] || 'https://meteolink.dev:9090';
const shots = [
  ['campagne', '../../book/figures/shot-campagne.png'],
  ['live', '../../book/figures/shot-live.png'],
  ['resultats', '../../book/figures/shot-resultats.png'],
  ['integrite', '../../book/figures/shot-integrite.png'],
];

const browser = await chromium.launch({ args: ['--no-sandbox'] });
const ctx = await browser.newContext({ viewport: { width: 1920, height: 1080 }, ignoreHTTPSErrors: true });
const page = await ctx.newPage();
await page.goto(base, { waitUntil: 'domcontentloaded', timeout: 30000 });
await page.waitForTimeout(4000);
for (const [panel, file] of shots) {
  await page.click(`button[data-panel="${panel}"]`, { timeout: 10000 });
  await page.waitForTimeout(2000);
  // capture en temp d'abord : l'indexeur/sync verrouille parfois les PNG
  // frais de book/figures (UNKNOWN open) — la copie réessaie, la capture non
  const tmp = path.join(fs.mkdtempSync(path.join(os.tmpdir(), 'shot-')), 'shot.png');
  await page.screenshot({ path: tmp });
  for (let i = 0; i < 10; i++) {
    try { fs.copyFileSync(tmp, file); break; }
    catch { if (i === 9) throw new Error('copie impossible: ' + file); await new Promise(r => setTimeout(r, 1500)); }
  }
  console.log('OK', file);
}
await browser.close();
