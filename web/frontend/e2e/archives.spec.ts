import { test, expect } from '@playwright/test'

// Couche 5 Tâche 5.2 — Archives : la provenance vit désormais dans le Labo,
// étape « Preuves », via le strip de provenance. PanelChooser et carte RDF
// appartiennent à la vue supprimée ; les sélecteurs ci-dessous sont les vivants.
test('archives — PanelChooser nav + RDF frozen provenance card', async ({ page }) => {
  await page.goto('/')
  await page.locator('[data-mode="labo"]').click()
  await expect(page.locator('.labo-view')).toBeVisible()
  await page.locator('.pr-step').nth(3).click()
  await expect(page.locator('.pr-step').nth(3)).toHaveAttribute('aria-current', 'step')
  await expect(page.locator('#chapter-preuves .preuves')).toBeVisible()
  await page.getByRole('button', { name: 'Provenance des preuves' }).click()
  const strip = page.locator('[data-testid="provenance-strip"]')
  await expect(strip).toBeVisible()
  // honnête : empreinte réelle quand /api/integrity l'expose, sinon mot d'absence
  const hash = strip.locator('[data-line="integrity"] .provenance-value.mono')
  const absent = strip.locator('[data-line="integrity"] .provenance-absent')
  await expect(hash.or(absent).first()).toBeVisible()
})

test('archives — PanelChooser tri-toggle present', async ({ page }) => {
  await page.goto('/')
  await page.locator('[data-mode="labo"]').click()
  await expect(page.locator('.labo-view')).toBeVisible()
  // Le tri-toggle du PanelChooser devient le rail d'étapes : quatre états, un courant.
  const rail = page.locator('.pipeline-rail')
  await expect(rail).toBeVisible()
  await expect(rail.locator('.pr-step')).toHaveCount(4)
  await expect(rail.locator('.pr-step[aria-current="step"]')).toHaveCount(1)
})
