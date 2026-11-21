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
  const selector = '[data-mode], .chrome-trigger, .cmd-arm, .cmd-stop, .btn-primary, .btn-danger, .modeswitch-btn, .pr-step';

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
        // Les panneaux inactifs sont hors écran par conception : le contrôle de
        // coupe ne porte que sur ce que l'opérateur peut réellement voir.
        .filter(e => e.getClientRects().length && !e.closest('[inert], [aria-hidden="true"]'))
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
  const slideMs = await page.evaluate(() => {
    const track = document.querySelector('.panel-track')!;
    const v = getComputedStyle(track).transitionDuration;
    return parseFloat(v) * (v.includes('ms') ? 1 : 1000);
  });
  expect(slideMs, 'glissement instantané en mouvement réduit').toBeLessThanOrEqual(1);
  const overflow = await page.evaluate(() =>
    document.documentElement.scrollWidth - document.documentElement.clientWidth);
  expect(overflow).toBeLessThanOrEqual(1);
});

test('rail — clavier W3C et mot de progression, libellés non tronqués à 890', async ({ page }) => {
  test.setTimeout(90000);
  await page.setViewportSize({ width: 890, height: 512 });
  await page.goto('/');
  await page.locator('[data-mode="labo"]').click();
  await expect(page.locator('.labo-view')).toBeVisible();

  const steps = page.locator('.pr-step');
  const progress = page.locator('.pr-progress');
  await expect(progress, 'mot de progression visible').toHaveText('Étape 1 sur 4');

  await steps.first().focus();
  await expect(steps.first(), 'le focus entre sur l’étape active').toBeFocused();
  await expect(steps.first()).toHaveAttribute('aria-current', 'step');

  await page.keyboard.press('ArrowRight');
  await expect(steps.nth(1), 'Right déplace le focus').toBeFocused();
  await expect(steps.first(), 'activation manuelle : aria-current ne bouge qu’à Entrée')
    .toHaveAttribute('aria-current', 'step');
  await expect(progress).toHaveText('Étape 1 sur 4');

  await page.keyboard.press('End');
  await expect(steps.nth(3), 'Fin saute à la dernière étape').toBeFocused();
  await page.keyboard.press('Home');
  await expect(steps.first(), 'Début saute à la première étape').toBeFocused();
  await page.keyboard.press('ArrowLeft');
  await expect(steps.nth(3), 'Left boucle de la première à la dernière').toBeFocused();

  await page.keyboard.press('Enter');
  await expect(steps.nth(3), 'Entrée active l’étape').toHaveAttribute('aria-current', 'step');
  await expect(progress).toHaveText('Étape 4 sur 4');
  await expect(page.locator('#chapter-preuves-title'), 'le focus rejoint le panneau').toBeFocused();
  await expect(page.locator('#chapter-preuves')).not.toHaveAttribute('inert');

  const wrapped = await page.evaluate(() => [...document.querySelectorAll('.pr-label')]
    .every(el => getComputedStyle(el).overflowWrap !== 'normal' && el.scrollWidth <= el.clientWidth + 1));
  expect(wrapped, 'libellés repliés sans coupure').toBe(true);
  const overflow = await page.evaluate(() =>
    document.documentElement.scrollWidth - document.documentElement.clientWidth);
  expect(overflow, 'aucun débordement horizontal').toBeLessThanOrEqual(1);
});
