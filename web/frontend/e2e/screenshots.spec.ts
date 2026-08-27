import { test } from '@playwright/test';

const shots = process.env.SHOT_DIR || '../../shots';

for (const [name, vp] of [
  ['desktop-1920', { width: 1920, height: 1080 }],
  ['desktop-1366', { width: 1366, height: 768 }],
  ['mobile-390', { width: 390, height: 844 }],
] as const) {
  test(`capture ${name}`, async ({ page }) => {
    await page.setViewportSize({ width: vp.width, height: vp.height });
    await page.goto('/');
    await page.waitForTimeout(2500); // SSE connect + first frames
    await page.waitForSelector('[data-panel="campagne"]', { timeout: 10000 });
    for (const panel of ['campagne','live','resultats','integrite'] as const) {
      const btn = page.locator(`[data-panel="${panel}"]`);
      if (await btn.count()) {
        await btn.click({ timeout: 5000 }).catch(()=> page.keyboard.press(panel==='campagne'?'1':panel==='live'?'2':panel==='resultats'?'3':'4'));
      } else {
        await page.keyboard.press(panel==='campagne'?'1':panel==='live'?'2':panel==='resultats'?'3':'4');
      }
      await page.waitForTimeout(900);
      await page.screenshot({ path: `${shots}/${name}-${panel}.png`, animations: 'disabled' });
    }
  });
}
