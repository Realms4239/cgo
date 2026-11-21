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

test('labo 1280px — rail épinglé à gauche, chapitres en document séquentiel', async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 768 });
  await page.goto('/');
  await page.getByRole('button', { name: 'Labo' }).click();
  await page.getByRole('heading', { name: 'Configurer' }).waitFor();
  const g = await page.evaluate(() => {
    const rail = document.querySelector('.pipeline-rail')!.getBoundingClientRect();
    const chapters = [...document.querySelectorAll('.labo-chapters > section')].map(e => e.getBoundingClientRect());
    return { railRight: rail.right, firstLeft: chapters[0]?.left,
      stacked: chapters.length === 4 && chapters.every(c => Math.abs(c.left - chapters[0].left) < 2) };
  });
  expect(g.railRight, 'le rail est à gauche du document').toBeLessThan(g.firstLeft!);
  expect(g.stacked, 'les quatre chapitres forment une séquence').toBe(true);
});
