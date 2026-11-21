import { expect, test } from '@playwright/test';

// Le contrat spatial ratifié : à partir de 1024 px, la décision est une
// colonne épinglée à gauche, le champ à droite, et rien ne défile en page.
for (const [w, h, spineMin] of [[1280, 768, 300], [1920, 1080, 340]] as const) {
  test(`cockpit ${w}px — spine à gauche, champ à droite, aucun défilement de page`, async ({ page }) => {
    await page.setViewportSize({ width: w, height: h });
    await page.goto('/');
    await expect(page.locator('#v-cockpit')).toBeVisible();

    const layout = await page.evaluate(() => {
      const main = document.getElementById('main')!;
      const spine = document.querySelector('.decision-spine')!.getBoundingClientRect();
      const field = document.querySelector('.cockpit-field')!.getBoundingClientRect();
      return {
        pageOverflow: main.scrollHeight - main.clientHeight,
        spineBottom: spine.bottom,
        spineRight: spine.right,
        fieldLeft: field.left,
        fieldBottom: field.bottom,
        viewportBottom: document.documentElement.clientHeight,
      };
    });

    expect(layout.pageOverflow, 'aucun scroll de page dans le cockpit ≥1024').toBeLessThanOrEqual(1);
    expect(layout.spineRight, 'la décision est à gauche du champ').toBeLessThan(layout.fieldLeft);
    expect(layout.spineRight, 'colonne de décision épinglée').toBeGreaterThanOrEqual(spineMin);
    expect(layout.fieldBottom, 'le champ tient dans la zone de travail').toBeLessThanOrEqual(layout.viewportBottom - 40);
  });
}

async function openLabo(page: import('@playwright/test').Page, w = 1280, h = 768) {
  await page.setViewportSize({ width: w, height: h });
  await page.goto('/');
  await page.locator('[data-mode="labo"]').click();
  await expect(page.locator('.labo-view')).toBeVisible();
}

for (const [w, h] of [[890, 512], [1280, 768], [1920, 1080]] as const) {
  test(`labo ${w}px — rail épinglé à gauche, un seul panneau, aucun scroll de page`, async ({ page }) => {
    await openLabo(page, w, h);
    const g = await page.evaluate(() => {
      const rail = document.querySelector('.pipeline-rail')!.getBoundingClientRect();
      const vp = document.querySelector('.panel-viewport')!.getBoundingClientRect();
      const panels = [...document.querySelectorAll('.panel-track > section')] as HTMLElement[];
      const main = document.getElementById('main')!;
      return {
        railRight: rail.right,
        viewportLeft: vp.left,
        count: panels.length,
        active: panels.filter(p => p.dataset.active === 'true').length,
        inside: panels.filter(p => {
          const r = p.getBoundingClientRect();
          return r.left >= vp.left - 1 && r.right <= vp.right + 1;
        }).length,
        overflow: main.scrollHeight - main.clientHeight,
      };
    });
    expect(g.railRight, 'le rail est à gauche du panneau').toBeLessThan(g.viewportLeft);
    expect(g.count, 'quatre panneaux montés').toBe(4);
    expect(g.active, 'un seul panneau actif').toBe(1);
    expect(g.inside, 'un seul panneau dans la fenêtre').toBe(1);
    expect(g.overflow, 'aucun scroll de page dans le labo').toBeLessThanOrEqual(1);
  });
}

test('labo — clic du rail : glissement, panneaux inertes, focus au titre, sans scroll', async ({ page }) => {
  await openLabo(page);

  await page.getByRole('button', { name: /Comparer/ }).click();

  await expect(page.locator('.pr-step').nth(2), 'l’étape courante est annoncée').toHaveAttribute('aria-current', 'step');
  await expect(page.locator('.panel-track'), 'translation d’un panneau entier')
    .toHaveAttribute('style', /translateX\(-200%\)/);
  for (const id of ['configurer', 'observer', 'preuves']) {
    await expect(page.locator(`#chapter-${id}`), `${id} inerte`).toHaveAttribute('inert', '');
    await expect(page.locator(`#chapter-${id}`), `${id} masqué aux lecteurs d’écran`)
      .toHaveAttribute('aria-hidden', 'true');
  }
  const active = page.locator('#chapter-comparer');
  await expect(active, 'panneau actif non inerte').not.toHaveAttribute('inert');
  await expect(active, 'panneau actif exposé').not.toHaveAttribute('aria-hidden');
  await expect(page.locator('#chapter-comparer-title'), 'focus sur le titre du panneau').toBeFocused();

  const overflow = await page.evaluate(() => {
    const main = document.getElementById('main')!;
    return main.scrollHeight - main.clientHeight;
  });
  expect(overflow, 'aucun scroll de page après le glissement').toBeLessThanOrEqual(1);
});
