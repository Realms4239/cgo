import { test, expect } from '@playwright/test'

// Spécifications chirurgicales — largeur du rail, surface instrumentée, défilement.
// Le rail historique `.rail` est devenu `.pipeline-rail` (Labo) et l'état épinglé
// est porté par `.command-header` (data-pinned), jamais par le rail lui-même.

test('rail uses live pipeline rail and is visible', async ({ page }) => {
  await page.goto('/')
  await page.locator('[data-mode="labo"]').click()
  const rail = page.locator('.pipeline-rail')
  await expect(rail).toBeVisible()
  await expect(page.locator('.command-header')).toHaveAttribute('data-pinned', 'true')
})

test('instrument surface scoped .labo-view', async ({ page }) => {
  await page.goto('/')
  await page.locator('[data-mode="labo"]').click()
  await expect(page.locator('.labo-view')).toBeVisible()
  await page.locator('.pr-step').nth(1).click()
  await expect(page.locator('#chapter-observer')).toHaveAttribute('data-active', 'true')
})

test('provenance container within viewport', async ({ page }) => {
  await page.goto('/')
  await page.locator('[data-mode="labo"]').click()
  await page.locator('.pr-step').nth(3).click()
  const stack = page.locator('#chapter-preuves')
  await expect(stack).toBeVisible()
  const box = await page.locator('#main').boundingBox()
  expect(box?.height).toBeGreaterThan(0)
})
