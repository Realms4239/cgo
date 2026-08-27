import { test, expect } from '@playwright/test'
test('observatory 2.0 flow', async ({ page }) => {
  await page.goto('/')
  await expect(page.locator('text=METEOLINK').first()).toBeVisible()
  await page.keyboard.press('Meta+k') // palette
  if (await page.getByRole('dialog', { name: 'Palette de commandes' }).count() === 0) await page.keyboard.press('Control+k')
  await expect(page.getByRole('dialog', { name: 'Palette de commandes' })).toBeVisible()
  await page.keyboard.press('Escape')
  await expect(page.locator('text=Actions rapides')).toBeVisible({timeout: 7000})
  await page.getByLabel('Navigation').getByRole('button', {name: 'Campagne'}).click()
  await expect(page.locator('input[name="site"]')).toBeVisible()
  const t=page.locator('[data-testid="qdi-sparkline"]')
  if(await t.count()>0) await expect(t.first()).toBeVisible()
})
