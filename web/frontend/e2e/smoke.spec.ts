import { test, expect } from '@playwright/test';

// Coque vivante : deux destinations (Synthèse / Labo) à la place des quatre
// panneaux LIEN supprimés, et une bande de confiance qui dit l'état de liaison.

test('shell renders two live modes and SSE connects', async ({ page }) => {
  await page.goto('/');
  for (const mode of ['cockpit', 'labo']) {
    await expect(page.locator(`[data-mode="${mode}"]`)).toBeVisible();
  }
  // le suivi de run peut basculer la destination active — la section affichée doit se rendre
  await expect(page.locator('#main section.view:not([hidden])')).toBeVisible();
  // après connexion SSE (~1 s) la bande de confiance dit la liaison par un mot
  await expect(page.locator('[data-slot="connection"] .trust-word')).toHaveText('EN LIGNE', { timeout: 5000 });
});

test('labo instrument receives 10 Hz frames', async ({ page }) => {
  await page.goto('/');
  await page.locator('[data-mode="labo"]').click();
  await page.locator('.pr-step').nth(1).click();
  await expect(page.locator('#chapter-observer')).toBeVisible();
  // la bande de portes expose toujours 8 cellules (G0–G7)
  await expect(page.locator('.labo-gate-row')).toHaveCount(8);
  // le champ temporel de l'instrument est monté et peint
  await expect(page.locator('#chapter-observer .time-field')).toBeVisible();
  // l'état de liaison n'est plus un bandeau OFFLINE mais un mot dans la bande de confiance
  await expect(page.locator('[data-slot="connection"] .trust-word')).toHaveText('EN LIGNE', { timeout: 5000 });
});
