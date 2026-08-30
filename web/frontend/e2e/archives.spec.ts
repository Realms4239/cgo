import { test, expect } from '@playwright/test'

// Couche 5 Tâche 5.2 — Archives = panneau intégrité par tri PanelChooser (SPA sans route /archives).
// Vocabulaire Q11 : le troisième panneau du sélecteur s'appelle désormais «Provenance».
test('archives — PanelChooser nav + RDF frozen provenance card', async ({ page }) => {
  await page.goto('/')
  await page.getByRole('button', { name: 'Provenance', exact: true }).click()
  await expect(page.getByText('RDF — provenance gelée')).toBeVisible()
  // honnête : pastille sha256 réelle quand /api/integrity l'expose, sinon EmptyState
  const hash = page.locator('[data-testid="provenance-hash"]')
  const empty = page.getByText('empreinte SHA non exposée')
  await expect(hash.or(empty).first()).toBeVisible()
})

test('archives — PanelChooser tri-toggle present', async ({ page }) => {
  await page.goto('/')
  const chooser = page.locator('.panel-chooser')
  await expect(chooser).toBeVisible()
  await expect(chooser).toHaveAttribute('data-tri', /m/)
  await expect(chooser).toHaveAttribute('data-tri', /c/)
})
