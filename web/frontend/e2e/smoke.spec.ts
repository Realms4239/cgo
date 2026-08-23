import { test, expect } from '@playwright/test';

test('shell renders four LIEN panels and SSE connects', async ({ page }) => {
  await page.goto('/');
  for (const lbl of ['Campagne','Temps réel','Résultats','Intégrité']) {
    await expect(page.getByRole('button', { name: lbl })).toBeVisible();
  }
  await expect(page.locator('#v-campagne')).toBeVisible();
  // after SSE connects (~1s) the header shows ● connecté
  await expect(page.locator('header')).toContainText('connecté', { timeout: 5000 });
});

test('live panel receives 10 Hz frames', async ({ page }) => {
  await page.goto('/');
  await page.getByRole('button', { name: 'Temps réel' }).click();
  await expect(page.locator('#v-live')).toBeVisible();
  // gates strip always 8 cells
  await expect(page.locator('.gates .gate')).toHaveCount(8);
  // live chart containers exist
  await expect(page.locator('#v-live .card').first()).toBeVisible();
  // after a moment the banner switches from OFFLINE to IDLE/BASELINE
  await expect(page.locator('.banner')).not.toContainText('OFFLINE', { timeout: 5000 });
});
