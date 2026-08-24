import { test, expect } from '@playwright/test'
test('madalink perfect flow', async ({ page }) => {
  await page.goto('/')
  await expect(page.locator('text=MadaLink')).toBeVisible()
  await page.getByRole('button', {name: 'Campagne'}).click()
  await expect(page.locator('text=Audit lien accessible')).toBeVisible()
  await page.getByRole('button', {name: 'Temps réel'}).click()
  await expect(page.locator('text=RTT')).toBeVisible()
  await page.getByRole('button', {name: 'Résultats'}).click()
  await expect(page.locator('text=small p95')).toBeVisible()
  await page.getByRole('button', {name: 'Intégrité'}).click()
  await expect(page.locator('text=archives gelées')).toBeVisible()
})
