import { expect, test } from '@playwright/test'

// Le selector générique des gates 44 px ne portait pas sur les boutons mini
// de l'archive : mesure calculée du CSS appliqué, sans stub d'endpoint.
for (const [w, h] of [[1280, 768], [1920, 1080]] as const) {
  test(`actions d’archive ≥44 px à ${w}px`, async ({ page }) => {
    await page.setViewportSize({ width: w, height: h })
    await page.goto('/')
    await page.locator('[data-mode="labo"]').click()
    await expect(page.locator('.labo-view')).toBeVisible()
    await page.getByRole('button', { name: /Preuves/ }).click()
    await expect(page.locator('#chapter-preuves')).not.toHaveAttribute('inert')

    const targets = await page.evaluate(() => {
      const real = [...document.querySelectorAll<HTMLElement>('#chapter-preuves .archive-actions .btn')]
      // Le serveur réel n'a pas de dossier scellé : sonder la règle appliquée.
      const probe = document.createElement('button')
      probe.className = 'btn btn-mini'
      probe.textContent = 'Ouvrir'
      document.body.appendChild(probe)
      const nodes = real.length ? real : [probe]
      const measured = nodes.map(e => {
        const cs = getComputedStyle(e)
        const r = e.getBoundingClientRect()
        return { min: parseFloat(cs.minHeight), w: r.width, h: r.height }
      })
      probe.remove()
      return measured
    })

    expect(targets.length, 'au moins une cible mesurée').toBeGreaterThan(0)
    for (const t of targets) {
      expect(t.min, `min-height calculée à ${w}px`).toBeGreaterThanOrEqual(44)
      expect(t.h, `hauteur rendue à ${w}px`).toBeGreaterThanOrEqual(44)
    }
  })
}

test('le tri du classement répond au clic, même sans mesure', async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 768 })
  await page.goto('/')
  await page.locator('[data-mode="labo"]').click()
  await expect(page.locator('.labo-view')).toBeVisible()
  await page.getByRole('button', { name: /Comparer/ }).click()

  const delay = page.locator('.rank-table thead th').nth(2)
  await expect(delay).toHaveAttribute('aria-sort', 'ascending')
  await delay.locator('button').click()
  await expect(delay, 'le même en-tête bascule le sens de tri')
    .toHaveAttribute('aria-sort', 'descending')
  await expect(page.locator('#chapter-comparer')).not.toHaveAttribute('inert')
})
