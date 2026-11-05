// slice26-capture.mjs — preuves visuelles des tranches 22–25, en lecture seule.
// Usage: node slice26-capture.mjs [baseURL] [outDir]
// Aucun bouton qui démarre un audit, un banc ou une écriture n'est cliqué.
import { chromium } from 'playwright-core';
import { mkdirSync, writeFileSync } from 'node:fs';

const base = process.argv[2] || 'http://127.0.0.1:9091';
const out = process.argv[3] || 'docs/img/shots/2026-11-04';
mkdirSync(out, { recursive: true });

const TITLE = 'mémoriser la médiane de référence, cliquer pour effacer — la barre compare avant et maintenant, même fenêtre';
const evidence = {};

const browser = await chromium
  .launch({ args: ['--no-sandbox'], channel: process.env.PW_CHANNEL || 'msedge' })
  .catch(() => chromium.launch({ args: ['--no-sandbox'], channel: 'chrome' }));

for (const [width, height] of [[1920, 1080], [1024, 768]]) {
  const ctx = await browser.newContext({ viewport: { width, height } });
  const page = await ctx.newPage();
  const errors = [];
  page.on('pageerror', e => errors.push(String(e)));
  await page.goto(base, { waitUntil: 'domcontentloaded', timeout: 30000 });
  await page.waitForTimeout(9000);
  const e = (evidence[width] = { errors });

  // ---- 22 : rapport sans charge mesurée, remédiation supprimée
  await page.click('button[data-mode="audit"]', { timeout: 10000 });
  await page.waitForTimeout(1600);
  const last = page.locator('#main button:has-text("VOIR LE DERNIER RAPPORT")').first();
  await last.click({ timeout: 6000 }).catch(() => {});
  await page.waitForTimeout(1600);
  const read = page.locator('[data-testid="audit-read-card"]');
  await read.waitFor({ timeout: 8000 });
  const readText = await read.innerText();
  e.slice22 = {
    hasSuppressedLine: readText.includes('Aucune remédiation émise'),
    hasProtectCta: readText.includes('Protéger ce lien'),
    auditId: await page.locator('.audit-read .card-head').innerText().catch(() => ''),
  };
  await read.scrollIntoViewIfNeeded();
  await page.waitForTimeout(400);
  await page.screenshot({ path: `${out}/slice22-audit-remediation-suppressed-${width}.png` });

  // ---- 23a : tiroir Réglages, deadline 220 ms
  await page.click('header button[aria-label="Réglages"]', { timeout: 6000 });
  await page.waitForTimeout(1200);
  const drawer = page.locator('.modal-drawer-panel');
  await drawer.waitFor({ timeout: 6000 });
  const deadline = await page
    .locator('.settings-row', { hasText: 'Deadline small p95' })
    .locator('input')
    .inputValue();
  e.slice23 = { deadline, hasBbrv3Hint: false };
  await page.screenshot({ path: `${out}/slice23-settings-deadline-${width}.png` });
  await page.keyboard.press('Escape');
  await page.waitForTimeout(700);

  // ---- 23b : banc, pastille BBRv3 absente
  await page.click('button[data-mode="bench"]', { timeout: 10000 });
  await page.waitForTimeout(1600);
  // les capacités de l'hôte ne se montrent qu'en mode recherche (aucune mesure)
  await page.click('button[data-research]', { timeout: 6000 }).catch(() => {});
  await page.waitForTimeout(1500);
  const hint = page.locator('[data-testid="bench-bbr3-hint"]');
  e.slice23.hasBbrv3Hint = await hint.count().then(n => n > 0);
  const config = page.locator('#bench-config');
  await config.scrollIntoViewIfNeeded().catch(() => {});
  await page.waitForTimeout(500);
  await hint.scrollIntoViewIfNeeded();
  await page.waitForTimeout(400);
  await page.screenshot({ path: `${out}/slice23-bench-bbrv3-${width}.png` });

  // ---- 24 + 25 : bord vivant, libellés et infobulle
  await page.click('.bench-anchors button:has-text("OBSERVER")', { timeout: 6000 }).catch(() => {});
  await page.waitForTimeout(1200);
  const edge = page.locator('[data-testid="edge-control"]');
  await edge.waitFor({ timeout: 8000 });
  const edgeText = await edge.innerText();
  const before = edge.locator('button:has-text("Maxi avant")').first();
  const title = await before.getAttribute('title');
  e.slice24 = {
    hasMaxiAvant: edgeText.includes('Maxi avant'),
    hasSansReglage: edgeText.includes('sans réglage (file simple)'),
  };
  e.slice25 = { title, titleMatches: title === TITLE };
  await edge.scrollIntoViewIfNeeded();
  await page.waitForTimeout(400);
  await page.screenshot({ path: `${out}/slice24-bench-edge-labels-${width}.png` });
  // Pas de capture séparée pour 25 : un title natif n'est pas peint. Le texte
  // exact est extrait ci-dessous ; le bouton reste visible dans la capture 24.

  await ctx.close();
}

await browser.close();
writeFileSync(`${out}/evidence.json`, JSON.stringify(evidence, null, 2) + '\n');
writeFileSync(`${out}/slice25-tooltip.txt`, TITLE + '\n');
console.log(JSON.stringify(evidence, null, 2));
