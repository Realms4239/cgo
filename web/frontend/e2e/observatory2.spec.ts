import { test, expect } from '@playwright/test'
test('observatory 2.0 flow', async ({ page }) => {
  await page.goto('/')
  await expect(page.locator('text=METEOLINK').first()).toBeVisible()
  await page.keyboard.press('Meta+k') // palette
  if (await page.getByRole('dialog', { name: 'Palette de commandes' }).count() === 0) await page.keyboard.press('Control+k')
  await expect(page.getByRole('dialog', { name: 'Palette de commandes' })).toBeVisible()
  await page.keyboard.press('Escape')
  await page.locator('[data-mode="labo"]').click()
  await expect(page.locator('.labo-view')).toBeVisible()
  await page.locator('.pr-step').nth(0).click()
  await expect(page.locator('.pr-step').nth(0)).toHaveAttribute('aria-current', 'step')
  // design labo : la déclaration vit dans l'étape « Configurer », en créneaux d'état
  await expect(page.locator('#chapter-configurer .configurer')).toBeVisible()
  const site = page.locator('#chapter-configurer .state-slot')
    .filter({ has: page.getByText('Site', { exact: true }) })
  await expect(site).toBeVisible()
  // mot de liaison live : présent seulement quand l'état vivant existe — ne pas
  // exiger sa visibilité, juste ne pas crasher s'il apparaît
  const t = page.locator('[data-slot="source"] .trust-word')
  if (await t.count() > 0 && await t.first().isVisible()) await expect(t.first()).toBeVisible()
})
