// debug-390.spec.ts — diagnostic reproductible du recouvrement header/nav à 390px.
import { test } from '@playwright/test';

test('diagnostic 390px : géométrie header vs rail nav', async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 });
  await page.goto('/');
  await page.waitForTimeout(1500);
  const geo = await page.evaluate(() => {
    const header = document.querySelector('header') as HTMLElement;
    const rail = document.querySelector('.rail') as HTMLElement;
    const btn = document.querySelector('[data-panel="integrite"]') as HTMLElement;
    const chooser = document.querySelector('.panel-chooser') as HTMLElement;
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
});
