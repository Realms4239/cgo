// slice26-exhaustive.mjs — captures exhaustives des tranches 20A–26′, en lecture seule.
// Usage: node slice26-exhaustive.mjs [baseURL] [outDir]
// Aucun bouton qui démarre un audit, un banc ou une écriture n'est cliqué.
import { chromium } from 'playwright-core';
import { mkdirSync, writeFileSync } from 'node:fs';

const base = process.argv[2] || 'http://127.0.0.1:9091';
const out = process.argv[3] || 'docs/img/shots/2026-11-04';
mkdirSync(out, { recursive: true });

const TITRE = 'mémoriser la médiane de référence, cliquer pour effacer — la barre compare avant et maintenant, même fenêtre';
const evidence = {};
const browser = await chromium.launch({ args: ['--no-sandbox'], channel: process.env.PW_CHANNEL || 'msedge' })
  .catch(() => chromium.launch({ args: ['--no-sandbox'], channel: 'chrome' }));
const log = step => console.log('STEP', step);

async function shot(page, name) {
  await page.screenshot({ path: `${out}/${name}.png` });
  evidence[name] = { width: await page.evaluate(() => window.innerWidth), height: await page.evaluate(() => window.innerHeight) };
}

for (const [width, height] of [[1920, 1080], [1024, 768]]) {
  const ctx = await browser.newContext({ viewport: { width, height } });
  const page = await ctx.newPage();
  const errors = [];
  page.on('pageerror', e => errors.push(String(e)));
  page.setDefaultTimeout(5000);
  page.setDefaultNavigationTimeout(30000);
  log(`viewport ${width}x${height} start`);
  await page.goto(base, { waitUntil: 'domcontentloaded', timeout: 30000 });
  await page.waitForTimeout(9000);
  const e = (evidence[String(width)] = { errors });

  // ---- Tableau de bord
  log('dashboard');
  await page.click('button[data-mode="dashboard"]').catch(err => errors.push(`dashboard click: ${err}`));
  await page.waitForTimeout(1500);
  await shot(page, `dash-dashboard-${width}`);
  e.dashboard = { cost: await page.locator('[data-testid="dash-cost"]').count().then(n => n > 0) };

  // ---- Auditer : déclarer / observer / lecture
  log('audit');
  await page.click('button[data-mode="audit"]').catch(err => errors.push(`audit click: ${err}`));
  await page.waitForTimeout(1500);
  e.audit = {
    declare: await page.locator('[data-testid="audit-declare-card"]').count().then(n => n > 0),
    observe: await page.locator('[data-testid="audit-observe-card"]').count().then(n => n > 0),
  };
  await shot(page, `audit-declare-${width}`);
  await page.locator('[data-testid="audit-observe-card"]').scrollIntoViewIfNeeded().catch(() => {});
  await page.waitForTimeout(400);
  await shot(page, `audit-observe-${width}`);
  const last = page.locator('#main button:has-text("VOIR LE DERNIER RAPPORT")').first();
  await last.click({ timeout: 6000 }).catch(() => {});
  await page.waitForTimeout(1600);
  const read = page.locator('[data-testid="audit-read-card"]');
  const readOk = await read.waitFor().then(() => true).catch(err => { errors.push(`audit read: ${err}`); return false; });
  const readText = readOk ? await read.innerText().catch(() => '') : '';
  e.slice22 = { hasSuppressedLine: readText.includes('Aucune remédiation émise'), hasProtectCta: readText.includes('Protéger ce lien') };
  await read.scrollIntoViewIfNeeded();
  await page.waitForTimeout(400);
  await shot(page, `slice22-audit-remediation-suppressed-${width}`);

  // ---- Réglages : deadline 220
  log('settings');
  await page.click('header button[aria-label="Réglages"]').catch(err => errors.push(`settings click: ${err}`));
  await page.waitForTimeout(1200);
  await page.locator('.modal-drawer-panel').waitFor().catch(err => errors.push(`settings drawer: ${err}`));
  e.slice23 = { deadline: await page.locator('.settings-row', { hasText: 'Deadline small p95' }).locator('input').inputValue() };
  await shot(page, `slice23-settings-deadline-${width}`);
  await page.keyboard.press('Escape');
  await page.waitForTimeout(700);

  // ---- Banc : configurer / observer / résultats (recherche, aucune mesure)
  log('bench config');
  await page.click('button[data-mode="bench"]').catch(err => errors.push(`bench click: ${err}`));
  await page.waitForTimeout(1600);
  await page.evaluate(() => document.getElementById('main')?.scrollTo(0, 0));
  await shot(page, `bench-configurer-${width}`);
  log('bench research');
  await page.click('button[data-research]').catch(err => errors.push(`research click: ${err}`));
  await page.waitForTimeout(1500);
  e.slice23.hasBbrv3Hint = await page.locator('[data-testid="bench-bbr3-hint"]').count().then(n => n > 0);
  const hint = page.locator('[data-testid="bench-bbr3-hint"]');
  await hint.scrollIntoViewIfNeeded().catch(() => {});
  await page.waitForTimeout(400);
  await shot(page, `slice23-bench-bbrv3-${width}`);
  await page.click('.bench-anchors button:has-text("OBSERVER")', { timeout: 6000 }).catch(() => {});
  await page.waitForTimeout(1200);
  const edge = page.locator('[data-testid="edge-control"]');
  const edgeOk = await edge.waitFor().then(() => true).catch(err => { errors.push(`edge control: ${err}`); return false; });
  const edgeText = edgeOk ? await edge.innerText().catch(() => '') : '';
  const title = edgeOk ? await edge.locator('button:has-text("Maxi avant")').first().getAttribute('title').catch(() => '') : '';
  e.slice24 = { hasMaxiAvant: edgeText.includes('Maxi avant'), hasSansReglage: edgeText.includes('sans réglage (file simple)') };
  e.slice25 = { title, titleMatches: title === TITRE };
  await edge.scrollIntoViewIfNeeded().catch(() => {});
  await page.waitForTimeout(400);
  await shot(page, `slice24-bench-edge-labels-${width}`);
  await shot(page, `bench-observer-${width}`);
  // ---- Résultats
  log('bench results');
  await page.click('.bench-anchors button:has-text("RÉSULTATS")').catch(err => errors.push(`results anchor: ${err}`));
  await page.waitForSelector('#bench-results .lb-card', { timeout: 20000 }).catch(() => {});
  await page.evaluate(() => document.getElementById('bench-results')?.scrollIntoView({ block: 'start' }));
  await page.waitForTimeout(1500);
  await shot(page, `bench-resultats-${width}`);
  e.duel = await page.locator('[data-testid="rank-verdict"]').count().then(n => n > 0);

  // ---- Preuves & rapports
  log('evidence');
  await page.click('button[data-mode="evidence"]').catch(err => errors.push(`evidence click: ${err}`));
  const evidenceLoaded = await page.waitForSelector('text=campagnes archivées', { timeout: 30000 }).then(() => true).catch(err => { errors.push(`evidence load: ${err}`); return false; });
  if (evidenceLoaded) {
    await page.click('[data-testid="corpus-details"] > summary', { timeout: 5000 }).catch(err => errors.push(`corpus details: ${err}`));
  } else {
    await page.evaluate(() => { const d = document.querySelector('[data-testid="corpus-details"]'); if (d) d.open = true; });
  }
  const corpusReady = evidenceLoaded && await page.waitForSelector('[data-testid="runs-table"]', { state: 'visible', timeout: 10000 }).then(() => true).catch(err => { errors.push(`corpus scan: ${err}`); return false; });
  e.reports = await page.locator('[data-testid="audit-reports"]').count().then(n => n > 0);
  await page.click('.evidence-anchors button:has-text("RAPPORTS D\'AUDIT")').catch(err => errors.push(`reports anchor: ${err}`));
  await page.waitForTimeout(500);
  await shot(page, `evidence-preuves-${width}`);
  if (corpusReady) {
    await page.click('.evidence-anchors button:has-text("MESURES & PREUVES")').catch(err => errors.push(`corpus anchor: ${err}`));
    await page.locator('#evidence-corpus').scrollIntoViewIfNeeded().catch(() => {});
    await page.waitForTimeout(500);
    await shot(page, `evidence-corpus-${width}`);
  }
  e.integrity = {
    runs: await page.locator('[data-testid="runs-table"]').count().then(n => n > 0),
    quarantine: await page.locator('[data-testid="quarantine-table"]').count().then(n => n > 0),
  };
  // Ouvre un rapport archivé (lecture seule) ; aucun audit n'est démarré.
  await page.locator('#evidence-reports button:has-text("OUVRIR")').first().click().catch(err => errors.push(`open report: ${err}`));
  await page.waitForSelector('[data-testid="audit-report-card"]', { timeout: 30000 }).catch(err => errors.push(`report card: ${err}`));
  const card = page.locator('[data-testid="audit-report-card"]').first();
  if (await card.count()) {
    await card.scrollIntoViewIfNeeded().catch(() => {});
    await page.waitForTimeout(400);
    await shot(page, `evidence-rapport-${width}`);
  }

  await ctx.close();
}
await browser.close();
writeFileSync(`${out}/evidence.json`, JSON.stringify(evidence, null, 2) + '\n');
writeFileSync(`${out}/slice25-tooltip.txt`, TITRE + '\n');
console.log(JSON.stringify(evidence, null, 2));
