import { test, expect } from '@playwright/test'

// Layer 5 Task 5.2 — Archives = integrite panel via PanelChooser tri (SPA has no /archives route).
test('archives — PanelChooser nav + provenance hash pill', async ({ page }) => {
  await page.goto('/')
  await page.getByRole('button', { name: 'Archives', exact: true }).click()
  await expect(page.locator('[data-testid="provenance-hash"]')).toBeVisible()
})

test('archives — PanelChooser tri-toggle present', async ({ page }) => {
  await page.goto('/')
  const chooser = page.locator('.panel-chooser')
  await expect(chooser).toBeVisible()
  await expect(chooser).toHaveAttribute('data-tri', /m/)
  await expect(chooser).toHaveAttribute('data-tri', /c/)
})
