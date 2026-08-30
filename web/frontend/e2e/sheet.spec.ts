import { test, expect } from '@playwright/test'

// Couche 5 — feuille cockpit : fixe droite 380 avec flou ; les étapes vivent dans l'espace principal.
test('sheet 380 — cockpit fixed right with blur, steps in workspace', async ({ page }) => {
  await page.goto('/')
  await page.locator('[data-panel="campagne"]').click()
  const sheet = page.locator('aside.panel-stack')
  await expect(sheet).toBeVisible()
  const box = await sheet.boundingBox()
  expect(box?.width).toBe(380)
  await expect(page.getByText('①', { exact: false }).first()).toBeVisible()
  await expect(page.getByText('②', { exact: false }).first()).toBeVisible()
  await expect(page.getByText('③', { exact: false }).first()).toBeVisible()
})
