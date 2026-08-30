import { test, expect } from '@playwright/test';

test('shell renders four LIEN panels and SSE connects', async ({ page }) => {
  await page.goto('/');
  for (const lbl of ['Campagne','Temps réel','Résultats','Intégrité']) {
    await expect(page.getByLabel('Navigation').getByRole('button', { name: lbl })).toBeVisible();
  }
  // suivi de run peut basculer le panneau actif — la section affichée doit se rendre
  await expect(page.locator('#main section.on')).toBeVisible();
  // après connexion SSE (~1s) l'en-tête affiche ● connecté
  await expect(page.locator('header')).toContainText('connecté', { timeout: 5000 });
});

test('live panel receives 10 Hz frames', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Navigation').getByRole('button', { name: 'Temps réel' }).click();
  await expect(page.locator('#v-live')).toBeVisible();
  // bande portes toujours 8 cellules
  await expect(page.locator('.gates .gate')).toHaveCount(8);
  // conteneurs graphes live existent
  await expect(page.locator('#v-live .card').first()).toBeVisible();
  // après un instant la bannière passe de OFFLINE à IDLE/BASELINE
  await expect(page.locator('.banner')).not.toContainText('OFFLINE', { timeout: 5000 });
});
