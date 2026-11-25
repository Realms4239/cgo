// debug-390.spec.ts — diagnostic reproductible du recouvrement header/nav à 390px.
// Le shell a remplacé `.rail` par `.pipeline-rail` (Labo), le bouton
// `[data-panel="integrite"]` par l'étape Preuves du rail et `.panel-chooser`
// par `.command-header`. Ces cibles existent à 390 px une fois le Labo ouvert :
// les anciens sélecteurs rendaient getBoundingClientRect nul.
import { test, expect } from '@playwright/test';

test('diagnostic 390px : géométrie header vs rail nav', async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 });
  await page.goto('/');
  await page.locator('[data-mode="labo"]').click();
  await page.waitForTimeout(1500);
  await page.locator('.pr-step').nth(3).click();
  await page.waitForTimeout(500);
  const geo = await page.evaluate(() => {
    const header = document.querySelector('header') as HTMLElement;
    const rail = document.querySelector('.pipeline-rail') as HTMLElement;
    const btn = document.querySelector('.pr-step[aria-current="step"]') as HTMLElement;
    const chooser = document.querySelector('.command-header') as HTMLElement;
    const hr = header.getBoundingClientRect();
    const rr = rail.getBoundingClientRect();
    const br = btn.getBoundingClientRect();
    const cr = chooser.getBoundingClientRect();
    const style = getComputedStyle(document.querySelector('#shell') as HTMLElement);
    return {
      headerRect: { top: hr.top, bottom: hr.bottom, h: hr.height },
      railRect: { top: rr.top, bottom: rr.bottom },
      btnRect: { top: br.top, bottom: br.bottom },
      chooserRect: { top: cr.top, bottom: cr.bottom },
      chooserBottomMinusBtnTop: cr.bottom - br.top,
      gridRows: style.gridTemplateRows,
      elementAtBtnCenter: document.elementFromPoint(br.left + br.width / 2, br.top + br.height / 2)?.className,
    };
  });
  console.log('GEO 390:', JSON.stringify(geo, null, 2));
  // garde-fou géométrique : à 390 px le rail ne recouvre pas l'en-tête.
  expect(geo.railRect.top).toBeGreaterThanOrEqual(geo.headerRect.bottom);
});
