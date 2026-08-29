import { test, expect } from '@playwright/test'
test('m8 flow', async ({ page }) => {
  await page.goto('/')
  await expect(page.locator('text=METEOLINK').first()).toBeVisible()
  await page.getByRole('button',{name:'Densité'}).click()
  await expect(page.locator('.rail')).toBeVisible()
  await page.keyboard.press('Meta+k')
  if (await page.getByRole('dialog', {name:'Palette de commandes'}).count() === 0) await page.keyboard.press('Control+k')
  await expect(page.getByRole('dialog', {name:'Palette de commandes'})).toBeVisible()
  await page.keyboard.press('Escape')
  await expect(page.getByRole('dialog', { name: 'Palette de commandes' })).toHaveCount(0)
  // Task 7 A/B — Live wall overlay baseline vs CAKE
  await page.getByLabel('Navigation').getByRole('button', {name: 'Temps réel'}).click()
  await expect(page.locator('.live-wall-overlay')).toBeVisible({timeout:5000})
  await expect(page.locator('.live-wall-overlay')).toContainText('baseline')
  await expect(page.locator('.live-wall-overlay')).toContainText('CAKE')
  // Résultats A/B diff badge + hardware provenance
  await page.getByLabel('Navigation').getByRole('button', {name: 'Résultats'}).click()
  await expect(page.locator('.ab-bento')).toBeVisible({timeout:5000})
  await expect(page.locator('.provenance, [data-testid="provenance"]')).toContainText(/hardware_recommendation|MikroTik|P1/)
})
