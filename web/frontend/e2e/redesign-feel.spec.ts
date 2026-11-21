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

test('boutons : focus clavier ≥2px au Labo et libellé primaire lisible au survol', async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 768 });
  await page.goto('/');
  await page.locator('[data-mode="labo"]').click();
  await expect(page.locator('.labo-view')).toBeVisible();

  // :focus-visible n'est porté que par le clavier : on tabule jusqu'à un bouton du Labo.
  let focus: { width: string; color: string; offset: string; visible: boolean } | null = null;
  for (let i = 0; i < 40; i++) {
    await page.keyboard.press('Tab');
    focus = await page.evaluate(() => {
      const el = document.activeElement;
      if (!(el instanceof HTMLElement) || !el.matches('.labo-view .btn')) return null;
      const s = getComputedStyle(el);
      return { width: s.outlineWidth, color: s.outlineColor, offset: s.outlineOffset, visible: el.matches(':focus-visible') };
    });
    if (focus) break;
  }
  expect(focus, 'un bouton du Labo reçoit le focus clavier').not.toBeNull();
  expect(focus!.visible, ':focus-visible actif sur le bouton Labo').toBe(true);
  expect(parseFloat(focus!.width), 'focus Labo ≥2px').toBeGreaterThanOrEqual(2);
  expect(focus!.offset, 'décalage 2px').toBe('2px');
  expect(focus!.color, 'encre teal ratifiée').toBe('rgb(15, 107, 98)');

  const luminance = (rgb: string): number => {
    const [r, g, b] = (rgb.match(/\d+(?:\.\d+)?/g) ?? []).slice(0, 3).map(Number);
    const lin = (v: number) => {
      const c = v / 255;
      return c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4;
    };
    return 0.2126 * lin(r) + 0.7152 * lin(g) + 0.0722 * lin(b);
  };
  const contrast = (fg: string, bg: string): number => {
    const a = luminance(fg);
    const b = luminance(bg);
    return (Math.max(a, b) + 0.05) / (Math.min(a, b) + 0.05);
  };

  for (const mode of ['cockpit', 'labo'] as const) {
    await page.goto('/');
    await page.locator(`[data-mode="${mode}"]`).click();
    const primary = mode === 'labo'
      ? page.locator('.labo-view .btn-primary:visible').first()
      : page.locator('#v-cockpit .btn-primary:visible').first();
    await primary.hover();
    const colors = await primary.evaluate((el) => {
      const s = getComputedStyle(el);
      return { fg: s.color, bg: s.backgroundColor };
    });
    expect(contrast(colors.fg, colors.bg), `libellé primaire lisible au survol (${mode})`).toBeGreaterThanOrEqual(4.5);
  }
});

const JARGON = [
  'pas encore mesuré', 'anneau 0', 'Règle temporelle', 'DERNIERS FAITS',
  'Une seule règle de temps', 'en attente de la première sonde',
];

for (const mode of ['Synthèse', 'Labo'] as const) {
  test(`${mode} — aucun jargon de sonde dans le texte visible`, async ({ page }) => {
    await page.goto('/');
    if (mode === 'Labo') {
      await page.getByRole('button', { name: 'Labo' }).click();
      await page.getByRole('heading', { name: 'Configurer' }).waitFor();
    }
    const text = (await page.locator('main').innerText()).toLowerCase();
    for (const term of JARGON) {
      expect(text, `terme banni : ${term}`).not.toContain(term.toLowerCase());
    }
  });
}
