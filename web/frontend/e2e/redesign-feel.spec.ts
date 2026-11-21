import { expect, test } from '@playwright/test';

test('détails finaux : lissage, focus, transitions et titres', async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 768 });
  await page.goto('/');

  const html = await page.evaluate(() => {
    const s = getComputedStyle(document.documentElement);
    return { smoothing: s.getPropertyValue('-webkit-font-smoothing'), moz: s.getPropertyValue('-moz-osx-font-smoothing') };
  });
  expect(html.smoothing, 'lissage antialiased à la racine').toBe('antialiased');
  // Propriété Firefox/macOS : Chromium ne l'expose pas dans getComputedStyle,
  // on vérifie donc sa présence dans la feuille, pas une valeur calculée.
  const css = await page.evaluate(async () => {
    const link = [...document.querySelectorAll('link[rel=stylesheet]')][0];
    if (!link) return '';
    const r = await fetch((link as HTMLLinkElement).href);
    return r.text();
  });
  expect(css, 'lissage macOS dans la feuille').toContain('-moz-osx-font-smoothing:grayscale');

  await page.keyboard.press('Tab');
  await page.keyboard.press('Tab');
  const focus = await page.evaluate(() => {
    const el = document.activeElement!;
    const s = getComputedStyle(el);
    return { tag: el.tagName, width: s.outlineWidth, color: s.outlineColor, offset: s.outlineOffset };
  });
  expect(focus.width, 'focus 2px').toBe('2px');
  expect(focus.offset, 'focus décalé de 2px').toBe('2px');
  expect(focus.color, 'focus encre teal ratifiée').toBe('rgb(15, 107, 98)');

  const btn = await page.evaluate(() => {
    const s = getComputedStyle(document.querySelector('.btn')!);
    return { prop: s.transitionProperty, dur: s.transitionDuration };
  });
  expect(btn.prop).toContain('transform');
  expect(btn.prop).toContain('background-color');
  expect(btn.prop, 'jamais transition: all').not.toBe('all');

  await page.getByRole('button', { name: 'Labo' }).click();
  await page.getByRole('heading', { name: 'Configurer' }).waitFor();
  const h2 = await page.evaluate(() => {
    const s = getComputedStyle(document.querySelector('.chapter-title')!);
    return { size: s.fontSize, wrap: s.textWrap };
  });
  expect(h2.size, 'titre de chapitre = échelle --fs-title').toBe('20px');
  expect(h2.wrap, 'titres équilibrés').toBe('balance');
});
