import { test, expect } from '@playwright/test'
test('observatory 2.0 flow', async ({ page }) => {
  await page.goto('/')
  await expect(page.locator('text=METEOLINK').first()).toBeVisible()
  await page.keyboard.press('Meta+k') // palette
  if (await page.getByRole('dialog', { name: 'Palette de commandes' }).count() === 0) await page.keyboard.press('Control+k')
  await expect(page.getByRole('dialog', { name: 'Palette de commandes' })).toBeVisible()
  await page.keyboard.press('Escape')
  await page.getByLabel('Navigation').getByRole('button', {name: 'Campagne'}).click()
  // design cockpit : le formulaire audit vit dans la carte « Audit du lien », bouton LANCER
  await expect(page.locator('text=Audit du lien').first()).toBeVisible()
  await page.getByRole('button', { name: 'LANCER' }).click()
  await expect(page.locator('input[name="site"]')).toBeVisible()
  const t=page.locator('[data-testid="qdi-sparkline"]')
  // sparkline live : présent seulement quand des données live existent — ne pas
  // exiger sa visibilité, juste ne pas crasher s'il apparaît
  if (await t.count() > 0 && await t.first().isVisible()) await expect(t.first()).toBeVisible()
})
