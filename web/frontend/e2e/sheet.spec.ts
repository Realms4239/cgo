import { test, expect } from '@playwright/test'

// Layer 5 Task 5.2 — Sheet 380 cockpit: fixed right sheet with backdrop blur, 3-step kit helpers.
test('sheet 380 — campagne panel-stack fixed right with blur', async ({ page }) => {
  await page.goto('/')
  await page.locator('[data-panel="campagne"]').click()
  const sheet = page.locator('.panel-stack')
  await expect(sheet).toBeVisible()
  const box = await sheet.boundingBox()
  expect(box?.width).toBe(380)
  await expect(page.getByText('① Auditer', { exact: false })).toBeVisible()
  await expect(page.getByText('② Comparer', { exact: false })).toBeVisible()
  await expect(page.getByText('③ Exporter', { exact: false })).toBeVisible()
})
