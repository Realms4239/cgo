// capture-shots.mjs — captures 1920×1080 du dashboard pour l'annexe E.
// Usage: node scripts/capture-shots.mjs [baseURL]  (défaut http://192.168.174.131:9090)
// Recapture après modifs UI, jamais de valeurs saisies : le serveur réel.
import { chromium } from 'playwright-core';

const base = process.argv[2] || 'http://192.168.174.131:9090';
const shots = [
  ['campagne', '../../book/figures/shot-campagne.png'],
  ['live', '../../book/figures/shot-live.png'],
  ['resultats', '../../book/figures/shot-resultats.png'],
  ['integrite', '../../book/figures/shot-integrite.png'],
];

const browser = await chromium.launch();
const page = await browser.newPage({ viewport: { width: 1920, height: 1080 } });
await page.goto(base, { waitUntil: 'networkidle', timeout: 30000 });
await page.waitForTimeout(2500);
for (const [panel, file] of shots) {
  await page.click(`button[data-panel="${panel}"]`, { timeout: 10000 });
  await page.waitForTimeout(2000);
  await page.screenshot({ path: file });
  console.log('OK', file);
}
await browser.close();
