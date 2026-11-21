import { test, expect, type APIRequestContext, type Page } from '@playwright/test';

// Parcours réel Labo → Comparer → Preuves contre le binaire Go servi par
// playwright.config.ts : aucun page.route, aucun stub, aucun hôte distant,
// aucune écriture dans le corpus gelé, aucun banc live. Le harnais transmet
// le Bearer : l'authentification reste intégralement active.

const SHA256 = /^[0-9a-f]{64}$/i;

async function apiJson(request: APIRequestContext, path: string): Promise<any> {
  const res = await request.get(path);
  expect(res.status(), path).toBe(200);
  expect(res.headers()['content-type'], path).toContain('application/json');
  return res.json();
}

// Le mode observation est la vérité de cet hôte : c'est lui qui rend le refus
// ARMER et l'absence de dossier scellé honnêtes.
async function expectObserve(request: APIRequestContext): Promise<void> {
  expect((await apiJson(request, '/api/health')).mode).toBe('observe');
}

async function goToLabo(page: Page, step: number): Promise<void> {
  await page.goto('/');
  await page.locator('[data-mode="labo"]').click();
  await expect(page.locator('.labo-view')).toBeVisible();
  await expect(page.locator('.pipeline-rail')).toBeVisible();
  await page.locator('.pr-step').nth(step).click();
  await expect(page.locator('.pr-step').nth(step)).toHaveAttribute('aria-current', 'step');
  await expect(page.locator('.pr-progress')).toHaveText(`Étape ${step + 1} sur 4`);
}

// Même règle que Comparer.tsx : strictement positif, sinon ce n'est pas mesuré.
const measured = (v: unknown): v is number =>
  typeof v === 'number' && Number.isFinite(v) && v > 0;

interface DomRow {
  rank: string;
  label: string;
  slots: Array<{ state: string | null; dash: boolean; word: string }>;
}

async function readRankRows(page: Page): Promise<DomRow[]> {
  return page.locator('#chapter-comparer .rank-table tbody .rank-row').evaluateAll(rows =>
    rows.map(r => ({
      rank: (r.querySelector('.rank-pin')?.textContent ?? '').trim(),
      label: (r.querySelector('th[scope="row"]')?.textContent ?? '').trim(),
      slots: Array.from(r.querySelectorAll('.state-slot')).map(s => ({
        state: s.getAttribute('data-state'),
        dash: !!s.querySelector('.state-slot-dash'),
        word: (s.querySelector('.state-slot-word')?.textContent ?? '').trim(),
      })),
    })),
  );
}

test('parcours réel Labo → Comparer → Preuves — données vives, absence honnête, sceau vérifié par acte', async ({ page, request }) => {
  test.setTimeout(90000);
  const crashes: string[] = [];
  page.on('pageerror', e => crashes.push('PAGEERROR ' + e.message));

  await expectObserve(request);
  const results = await apiJson(request, '/api/results');
  const integrity = await apiJson(request, '/api/integrity');
  const list = await apiJson(request, '/api/audit/list');
  expect(results.available, 'résultats réels du corpus gelé').toBe(true);
  const groups: any[] = results.groups;
  expect(groups.length).toBeGreaterThan(0);
  expect(Array.isArray(list.audits)).toBe(true);
  expect(list.audits, 'hôte observe : aucune preuve scellée').toEqual([]);

  // ---- Labo → Comparer ----
  await goToLabo(page, 2);
  await expect(page.locator('#chapter-comparer .comparer')).toBeVisible();
  await expect(page.locator('#chapter-comparer .comparer-lede'))
    .toContainText("n'a pas de valeur et se classe en dernier");

  // Une ligne par groupe réel de /api/results : zéro ligne inventée.
  const rows = page.locator('#chapter-comparer .rank-table tbody .rank-row');
  await expect(rows).toHaveCount(groups.length, { timeout: 15000 });
  let dom = await readRankRows(page);
  const absentDelay = groups.filter(g =>
    !measured(g.small_p95_valid_median ?? g.small_p95_median)).length;
  expect(dom.filter(r => r.rank === '—'), 'autant de « — » que de délais non mesurés')
    .toHaveLength(absentDelay);
  expect(absentDelay, 'le corpus fournit au moins une mesure absente').toBeGreaterThan(0);

  const firstAbsent = dom.findIndex(r => r.rank === '—');
  for (const r of dom.slice(firstAbsent)) {
    expect(r.rank, 'non mesuré classé en dernier').toBe('—');
    expect(r.slots[0], 'créneau délai honnête').toEqual({ state: 'absent', dash: true, word: 'non mesuré' });
  }
  expect(dom.slice(0, firstAbsent).every(r => r.rank !== '—'), 'mesurés avant les absents').toBe(true);

  // Inverser le tri ne doit jamais promouvoir un absent en tête.
  const sortDelay = page.locator('#chapter-comparer .rank-table thead .rank-sort')
    .filter({ hasText: 'Délai sous charge' });
  await sortDelay.click();
  await expect(page.locator('#chapter-comparer .rank-table thead th[aria-sort="descending"]')).toHaveCount(1);
  dom = await readRankRows(page);
  expect(dom.filter(r => r.rank === '—')).toHaveLength(absentDelay);
  expect(dom[dom.length - 1].rank, 'absent toujours dernier en sens inverse').toBe('—');
  await sortDelay.click(); // retour à l'ordre ascendant réel

  // Duel vide : des mots, jamais une égalité ni un gagnant fabriqué.
  const duel = page.locator('#chapter-comparer .duel');
  await expect(duel).toHaveAttribute('data-state', 'refused');
  await expect(duel).toContainText('Sélectionnez deux mesures');
  await expect(duel).toContainText('Une mesure absente ne vaut pas une mesure égale.');
  await expect(duel).not.toContainText('égalité');
  await expect(duel).not.toContainText('résultats proches');
  await expect(duel).not.toContainText('plus rapide');

  const pins = page.locator('#chapter-comparer .rank-pin');
  await pins.first().click(); // une seule épingle : toujours refusé
  await expect(duel).toHaveAttribute('data-state', 'refused');

  const absentPin = pins.filter({ hasText: /^—$/ }).first();
  await expect(absentPin).toBeVisible();
  await absentPin.click(); // deuxième épingle non mesurée : refus franc
  await expect(duel).toHaveAttribute('data-state', 'refused');
  await expect(duel).toContainText('Une mesure absente ne vaut pas une mesure égale.');
  await expect(duel).not.toContainText('plus rapide');
  await expect(duel.locator('.duel-delta')).toHaveCount(0);

  // Deux mesures réelles : le duel nomme un gagnant issu du corpus, sans invention.
  const measuredPins = pins.filter({ hasNotText: /^—$/ });
  expect(await measuredPins.count()).toBeGreaterThanOrEqual(2);
  await measuredPins.nth(0).click(); // troisième geste : on repose la référence
  await measuredPins.nth(1).click();
  await expect(duel).toHaveAttribute('data-state', 'compared');
  await expect(duel.locator('.duel-word')).toContainText('plus rapide');
  await expect(duel.locator('.duel-scale .duel-bar')).toHaveCount(2);
  await expect(duel.locator('[data-testid="duel-delta"]')).toContainText('ms');

  // ---- Comparer → Preuves, sans rechargement ----
  await page.locator('.pr-step').nth(3).click();
  await expect(page.locator('.pr-step').nth(3)).toHaveAttribute('aria-current', 'step');
  await expect(page.locator('#chapter-preuves .preuves')).toBeVisible();
  await expect(page.locator('.labo-view')).toBeVisible();
  await expect(page.locator('#chapter-preuves .preuves-lede'))
    .toContainText("rien n'est affirmé avant sa réponse");

  // Ligne de corpus : sans rapport scellé lu, elle dit son absence.
  const corpus = page.locator('#chapter-preuves .corpus-line');
  await expect(corpus).toContainText('—');
  await expect(corpus).toContainText('Preuve manquante');

  // Sceau : « À VÉRIFIER » au repos, l'acte seul peut le faire passer.
  const seal = page.locator('#chapter-preuves .seal-inspector');
  const badge = seal.locator('.seal-badge');
  await expect(badge).toHaveAttribute('data-state', 'unverified');
  await expect(badge).toHaveText('À VÉRIFIER');
  await expect(badge).not.toHaveText('VÉRIFIÉ');
  await expect(seal.locator('.seal-msg')).toHaveText('Sceau non vérifié.');

  const expectedVerified = integrity.available === true
    && typeof integrity.sha256 === 'string' && SHA256.test(integrity.sha256);
  const [verifyResp] = await Promise.all([
    page.waitForResponse(r => r.url().includes('/api/integrity') && r.request().method() === 'GET'),
    seal.locator('.seal-verify').click(),
  ]);
  expect(verifyResp.status()).toBe(200);
  if (expectedVerified) {
    await expect(badge).toHaveAttribute('data-state', 'verified');
    await expect(badge).toHaveText('VÉRIFIÉ');
    await expect(seal.locator('.seal-msg')).toHaveText('Intégrité vérifiée sur une empreinte complète.');
    await expect(seal.locator('.seal-hash')).toHaveText(`${integrity.sha256.slice(0, 8)}…`);
  } else {
    await expect(badge).toHaveAttribute('data-state', 'unverified');
    await expect(seal.locator('.seal-msg')).toContainText(/Empreinte incomplète|Sceau injoignable/);
    await expect(seal.locator('.seal-hash')).toHaveCount(0);
  }

  // Comptes et empreinte courte : exactement la réponse réelle /api/integrity.
  const archive = page.locator('#chapter-preuves .archive');
  if (typeof integrity.hash8 === 'string' && integrity.hash8) {
    await expect(archive).toContainText(`sceau ${integrity.hash8}`);
  }
  const counts = [
    ['Campagnes archivées', integrity.runs],
    ['Mesures valides', integrity.valid],
    ['Mesures écartées', integrity.quarantined],
  ] as const;
  for (const [label, value] of counts) {
    const card = archive.locator(`[data-metric="${label}"]`);
    await expect(card).toBeVisible();
    await expect(card.locator('.mc-label')).toHaveText(label);
    if (typeof value === 'number') await expect(card.locator('.mc-value')).toHaveText(String(value));
  }

  // Archive : /api/audit/list est vide → état vide honnête, aucune ligne forgée.
  await expect(archive.locator('.archive-count')).toHaveText('0 entrée');
  const empty = archive.locator('.c-empty.empty');
  await expect(empty).toBeVisible();
  await expect(empty).toContainText('aucune donnée');
  await expect(empty).toContainText("Aucun rapport d'audit archivé.");
  await expect(archive.locator('.archive-row')).toHaveCount(0);
  await expect(archive.getByRole('button', { name: 'Ouvrir' })).toHaveCount(0);
  await expect(archive.getByRole('link', { name: 'RAPPORT MD' })).toHaveCount(0);
  await expect(archive.getByRole('button', { name: 'CITER' })).toHaveCount(0);
  await expect(page.locator('.proof-drawer')).toHaveCount(0); // aucun faux dossier ouvert

  // PARKED : OUVRIR / RAPPORT MD / CITER sur une ligne scellée réelle demanderait
  // /api/audit/list non vide ; on n'invente aucun audit pour forcer le geste.
  // Sonde négative réelle : un id nommé absent échoue, sans rapport fabriqué.
  const missingId = `e2e-sans-dossier-${Date.now()}`;
  const missing = await request.get(`/api/audit/report?audit_id=${encodeURIComponent(missingId)}`);
  expect(missing.status(), 'dossier absent = 404').toBe(404);
  expect(await missing.text()).toContain('rapport introuvable');

  expect(crashes, 'zéro exception JavaScript').toEqual([]);
});

test('toast honnête — refus réel 501 du serveur en observation, aucune campagne', async ({ page, request }) => {
  test.setTimeout(90000);
  await expectObserve(request);
  await goToLabo(page, 0);

  const arm = page.locator('.cmd-arm');
  await expect(arm).toBeEnabled();
  const [startResp] = await Promise.all([
    page.waitForResponse(r => r.url().includes('/api/audit/start') && r.request().method() === 'POST'),
    arm.click(),
  ]);
  expect(startResp.status(), 'le service refuse en observation').toBe(501);
  expect((await startResp.json()).error).toContain('mode observation');

  // Le mot du toast vient de la réponse réelle, jamais d'une copie de succès.
  await expect(page.locator('.toast')).toContainText('Échec audit');
  await expect(page.locator('.toast')).toContainText('501');
  await expect(page.locator('.toast')).not.toContainText('Audit lancé');

  const state = await apiJson(request, '/api/state');
  expect(state.running, 'aucune campagne démarrée').toBe(false);
  expect(state.total_events).toBe(0);
  expect(await page.locator('.pr-progress').textContent()).toBe('Étape 1 sur 4');
});
