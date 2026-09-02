import { test, expect } from '@playwright/test'
test('madalink perfect flow', async ({ page }) => {
  await page.goto('/')
  await expect(page.locator('text=METEOLINK').first()).toBeVisible()
  await page.getByLabel('Navigation').getByRole('button', {name: 'Campagne de mesure'}).click()
  // design cockpit verrouillé : 3 cartes ① Auditer ② Comparer ③ Exporter
  await expect(page.locator('text=Auditer').first()).toBeVisible()
  await page.getByLabel('Navigation').getByRole('button', {name: 'Tableau live'}).click()
  await expect(page.locator('text=RTT').first()).toBeVisible()
  await page.getByLabel('Navigation').getByRole('button', {name: 'Résultats'}).click()
  await expect(page.locator('text=small p95').first()).toBeVisible()
  await page.getByLabel('Navigation').getByRole('button', {name: 'Provenance & archives'}).click()
  await expect(page.locator('text=archives gelées').first()).toBeVisible()
})
