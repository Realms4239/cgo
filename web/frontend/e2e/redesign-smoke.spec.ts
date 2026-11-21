import { expect, test } from '@playwright/test';

// Fumée NON simulée : le service Go répond réellement, le SPA embedé est
// servi par le binaire, aucune route n'est interceptée.
test('service réel, bascule des deux modes, action gardée, rechargement', async ({ page, request }) => {
  test.setTimeout(90000);

  const health = await request.get('/api/health');
  expect(health.ok()).toBeTruthy();
  expect((await health.json()).ok).toBe(true);

  await page.goto('/');
  await expect(page).toHaveTitle(/Cockpit/);
  await expect(page.locator('[data-mode="cockpit"]')).toHaveAttribute('aria-current', 'page');
  await expect(page.locator('#v-cockpit')).toBeVisible();

  await page.locator('[data-mode="labo"]').click();
  await expect(page).toHaveTitle(/Labo/);
  await expect(page.locator('[data-mode="labo"]')).toHaveAttribute('aria-current', 'page');
  await expect(page.locator('.labo-view')).toBeVisible();

  // La commande est gardée : ARRÊTER reste désactivé avant toute mesure.
  // ARMER produit TOUJOURS un résultat réel et visible : soit la campagne
  // démarre, soit le service refuse (mode observation sur hôte non-Linux)
  // et l'interface le dit. Le test n'invente jamais une campagne lancée.
  await expect(page.locator('.cmd-stop')).toBeDisabled();
  await page.locator('.cmd-arm').click();
  const outcome = await page.waitForFunction(() => {
    const state = document.querySelector('.command-header')?.getAttribute('data-state');
    const refused = [...document.querySelectorAll('.toast')]
      .some(el => (el.textContent ?? '').includes('Échec audit'));
    if (refused) return 'refused';
    if (state === 'observing') return 'observing';
    return null;
  });
  const result = await outcome.jsonValue();
  expect(['observing', 'refused']).toContain(result);
  if (result === 'observing') {
    await expect(page.locator('.ch-state')).toHaveText('Mesure en cours');
    // Nettoyage réel : on annule la campagne lancée, sans intercepter l'API.
    expect((await request.post('/api/audit/stop')).ok()).toBeTruthy();
  } else {
    // Refus honnête du service : aucun état « en cours » n'est simulé.
    await expect(page.locator('.command-header')).toHaveAttribute('data-state', 'idle');
    await expect(page.locator('.toast')).toContainText('Échec audit');
  }

  // Le rechargement rouvre la destination choisie, sans tiers caché.
  await page.locator('[data-mode="cockpit"]').click();
  await expect(page.locator('#v-cockpit')).toBeVisible();
  await page.reload();
  await expect(page).toHaveTitle(/Cockpit/);
  await expect(page.locator('[data-mode="cockpit"]')).toHaveAttribute('aria-current', 'page');

  await page.locator('[data-mode="labo"]').click();
  await page.reload();
  await expect(page).toHaveTitle(/Labo/);
  await expect(page.locator('[data-mode="labo"]')).toHaveAttribute('aria-current', 'page');
  await expect(page.locator('.labo-view')).toBeVisible();
});
