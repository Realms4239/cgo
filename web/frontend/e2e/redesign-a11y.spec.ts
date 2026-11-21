import AxeBuilder from '@axe-core/playwright';
import { expect, test } from '@playwright/test';

const VIEWPORTS = [
  { width: 890, height: 512 },
  { width: 1280, height: 768 },
  { width: 1920, height: 1080 },
];
const TAGS = ['wcag2a', 'wcag2aa', 'wcag21aa', 'wcag22aa'];

async function openMode(page: import('@playwright/test').Page, mode: 'cockpit' | 'labo') {
  await page.goto('/');
  await page.locator(`[data-mode="${mode}"]`).click();
  if (mode === 'labo') await expect(page.locator('.labo-view')).toBeVisible();
  else await expect(page.locator('#v-cockpit')).toBeVisible();
}

for (const viewport of VIEWPORTS) {
  for (const mode of ['cockpit', 'labo'] as const) {
    test(`axe ${TAGS.join('+')} — ${viewport.width}×${viewport.height} ${mode}`, async ({ page }) => {
      test.setTimeout(90000);
      await page.setViewportSize(viewport);
      await openMode(page, mode);

      // Le seuil reste celui de la spec : on corrige la source, pas l'axe.
      const results = await new AxeBuilder({ page }).withTags(TAGS).analyze();
      const bad = results.violations
        .filter(v => v.impact === 'critical' || v.impact === 'serious')
        .map(v => ({ id: v.id, nodes: v.nodes.map(n => n.target) }));
      expect(bad, 'aucune violation critique ou sérieuse').toEqual([]);

      const overflow = await page.evaluate(() =>
        document.documentElement.scrollWidth - document.documentElement.clientWidth);
      expect(overflow, 'pas de débordement horizontal').toBeLessThanOrEqual(1);
    });
  }
}

test('cibles tactiles ≥44 px et tuiles non coupées de 890 à 1920 px', async ({ page }) => {
  test.setTimeout(90000);
  const selector = '[data-mode], .chrome-trigger, .cmd-arm, .cmd-stop, .btn-primary, .btn-danger, .modeswitch-btn';

  for (const viewport of VIEWPORTS) {
    await page.setViewportSize(viewport);
    await openMode(page, 'cockpit');
    const cockpitSmall = await page.evaluate((sel: string) => [...document.querySelectorAll(sel)]
      .filter(e => e.getClientRects().length)
      .map(e => ({ word: (e.textContent || '').trim().slice(0, 24), w: e.getBoundingClientRect().width, h: e.getBoundingClientRect().height }))
      .filter(x => x.w < 44 || x.h < 44), selector);
    expect(cockpitSmall, `cibles cockpit ${viewport.width}`).toEqual([]);

    await openMode(page, 'labo');
    const laboSmall = await page.evaluate((sel: string) => [...document.querySelectorAll(sel)]
      .filter(e => e.getClientRects().length)
      .map(e => ({ word: (e.textContent || '').trim().slice(0, 24), w: e.getBoundingClientRect().width, h: e.getBoundingClientRect().height }))
      .filter(x => x.w < 44 || x.h < 44), selector);
    expect(laboSmall, `cibles labo ${viewport.width}`).toEqual([]);

    // Les tuiles du référentiel sont entières : le défilement est vertical.
    await page.getByRole('button', { name: /Profils de campagne/ }).first().click();
    const clipped = await page.evaluate(() => {
      const width = window.innerWidth;
      return [...document.querySelectorAll('.profile-tile, .mc-card')]
        .filter(e => e.getClientRects().length)
        .map(e => {
          const r = e.getBoundingClientRect();
          return { w: r.width, h: r.height, left: r.left, right: r.right, over: e.scrollWidth - e.clientWidth };
        })
        .filter(x => x.left < -1 || x.right > width + 1 || x.w <= 0 || x.h <= 0 || x.over > 1);
    });
    expect(clipped, `tuiles coupées ${viewport.width}`).toEqual([]);
  }
});

test('mouvement réduit et contraste forcé restent lisibles', async ({ page }) => {
  test.setTimeout(90000);
  await page.emulateMedia({ reducedMotion: 'reduce', forcedColors: 'active' });
  await openMode(page, 'cockpit');
  // Le focus clavier est la seule porte d'entrée de :focus-visible.
  await page.keyboard.press('Tab');
  await page.keyboard.press('Tab');

  const motion = await page.evaluate(() => {
    const view = document.querySelector('.view');
    const mode = document.querySelector('[data-mode="cockpit"]');
    const ms = (v: string) => parseFloat(v) * (v.includes('ms') ? 1 : 1000);
    return {
      animation: ms(getComputedStyle(view!).animationDuration),
      transition: ms(getComputedStyle(mode!).transitionDuration),
      outline: getComputedStyle(mode!).outlineWidth,
      border: getComputedStyle(mode!).borderTopWidth,
    };
  });
  expect(motion.animation, 'animation collapse en mouvement réduit').toBeLessThanOrEqual(1);
  expect(motion.transition, 'transition collapse en mouvement réduit').toBeLessThanOrEqual(1);
  expect(parseFloat(motion.outline)).toBeGreaterThanOrEqual(2);
  expect(parseFloat(motion.border)).toBeGreaterThanOrEqual(1);

  await page.locator('[data-mode="labo"]').click();
  await expect(page.locator('.labo-view')).toBeVisible();
  const overflow = await page.evaluate(() =>
    document.documentElement.scrollWidth - document.documentElement.clientWidth);
  expect(overflow).toBeLessThanOrEqual(1);
});
