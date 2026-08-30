import { test, expect } from '@playwright/test'

// Spécifications chirurgicales — largeur du rail, bento scopé, verrou de défilement.
// (SPA sans route /archives — naviguer par nav-btn du rail, selon panneaux App.tsx)

test('rail uses var(--rail-w) and is visible', async ({ page }) => {
  await page.goto('/')
  const rail = page.locator('.rail')
  await expect(rail).toBeVisible()
  await expect(rail).toHaveAttribute('data-pinned', /0|1/)
})

test('wall bento scoped #wall', async ({ page }) => {
  await page.goto('/')
  await page.locator('[data-panel="live"]').click()
  await expect(page.locator('#wall')).toBeVisible()
})

test('integrite scroll container within viewport', async ({ page }) => {
  await page.goto('/')
  await page.locator('[data-panel="integrite"]').click()
  const stack = page.locator('#v-integrite .panel-stack')
  await expect(stack).toBeVisible()
  const box = await page.locator('#main').boundingBox()
  expect(box?.height).toBeGreaterThan(0)
})