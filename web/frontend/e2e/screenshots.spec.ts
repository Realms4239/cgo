import { test } from '@playwright/test';

const shots = process.env.SHOT_DIR || '../../shots';

// Captures de la vraie interface à deux destinations : Synthèse puis les
// quatre chapitres du Labo. Aucun onglet d'ancienne vue supprimée.
const surfaces = [
  { id: 'cockpit', chapter: -1 },
  { id: 'configurer', chapter: 0 },
  { id: 'observer', chapter: 1 },
  { id: 'comparer', chapter: 2 },
  { id: 'preuves', chapter: 3 },
] as const;

for (const [name, vp] of [
  ['desktop-1920', { width: 1920, height: 1080 }],
  ['desktop-1366', { width: 1366, height: 768 }],
  ['mobile-390', { width: 390, height: 844 }],
] as const) {
  test(`capture ${name}`, async ({ page }) => {
    await page.setViewportSize({ width: vp.width, height: vp.height });
    await page.goto('/');
    await page.waitForTimeout(2500); // SSE connect + premières trames
    await page.waitForSelector('[data-mode="cockpit"]', { timeout: 10000 });
    for (const surface of surfaces) {
      if (surface.chapter < 0) {
        await page.locator('[data-mode="cockpit"]').click();
      } else {
        await page.locator('[data-mode="labo"]').click();
        await page.locator('.pr-step').nth(surface.chapter).click();
      }
      await page.waitForTimeout(900);
      await page.screenshot({ path: `${shots}/${name}-${surface.id}.png`, animations: 'disabled' });
    }
  });
}
