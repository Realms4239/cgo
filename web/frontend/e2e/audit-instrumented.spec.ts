// audit-instrumented.spec.ts — audit passif : console, réseau, 4xx/5xx, sur les deux destinations.
// Ne modifie rien, n'échoue que sur de vraies erreurs frontend (pas les 501 attendues en mode observe).
import { test, expect } from '@playwright/test';

const DESTINATIONS = [
  { mode: 'cockpit', view: '#v-cockpit' },
  { mode: 'labo', view: '.labo-view' },
] as const;

test('audit: les 4 vues sans erreur console ni requête échouée', async ({ page }) => {
  const consoleErrors: string[] = [];
  const failedRequests: string[] = [];
  const badResponses: string[] = [];

  page.on('console', (msg) => {
    if (msg.type() === 'error') consoleErrors.push(msg.text());
  });
  page.on('pageerror', (err) => {
    consoleErrors.push('PAGEERROR: ' + err.message);
  });
  page.on('requestfailed', (req) => {
    // ignorer les annulations de navigation SSE reconnect
    if (req.failure()?.errorText !== 'net::ERR_ABORTED') {
      failedRequests.push(req.method() + ' ' + req.url() + ' :: ' + req.failure()?.errorText);
    }
  });
  page.on('response', (resp) => {
    const s = resp.status();
    // 501 = mode observe attendu sur Windows ; 404 favicon toléré ; le reste est un bug
    if (s >= 400 && s !== 501 && !resp.url().includes('favicon')) {
      badResponses.push(s + ' ' + resp.request().method() + ' ' + resp.url());
    }
  });

  await page.goto('/');
  // Le chrome dit l'état de liaison par un mot : EN LIGNE ou HORS LIGNE, jamais « connecté ».
  await expect(page.locator('[data-slot="connection"] .trust-word'))
    .toHaveText(/EN LIGNE|HORS LIGNE/, { timeout: 8000 });

  for (const d of DESTINATIONS) {
    await page.locator(`[data-mode="${d.mode}"]`).click();
    await expect(page.locator(d.view)).toBeVisible();
    await page.waitForTimeout(1200); // laisser SSE + requêtes se dérouler
  }

  // Retour aux destinations réelles : filtrer le bruit de test environment
  const realErrors = consoleErrors.filter((e) =>
    !e.includes('favicon') && !e.includes('ERR_CONNECTION_REFUSED'));
  if (realErrors.length) console.log('CONSOLE ERRORS:\n' + realErrors.join('\n---\n'));
  if (failedRequests.length) console.log('FAILED REQUESTS:\n' + failedRequests.join('\n'));
  if (badResponses.length) console.log('BAD RESPONSES:\n' + badResponses.join('\n'));

  expect(realErrors, 'erreurs console').toEqual([]);
  expect(failedRequests, 'requêtes échouées').toEqual([]);
  expect(badResponses, 'réponses 4xx/5xx inattendues').toEqual([]);
});
