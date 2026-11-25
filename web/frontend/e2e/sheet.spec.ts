import { test, expect } from '@playwright/test'

// Couche 5 — feuille fixe droite 380 px et étapes numérotées dans l'espace de
// travail. L'ancienne `aside.panel-stack` du panneau Campagne n'existe plus :
// le panneau vivant est le tiroir Réglages (`.modal-drawer-panel`, 380 px),
// et les étapes vivent dans le rail des chapitres du Labo.

test('sheet 380 — fixed right sheet with blur, steps in workspace', async ({ page }) => {
  await page.goto('/')
  await page.locator('.chrome-trigger[aria-label="Réglages"]').click()
  const sheet = page.locator('.modal-drawer-panel')
  await expect(sheet).toBeVisible()
  const box = await sheet.boundingBox()
  expect(box?.width).toBe(380)
  await page.keyboard.press('Escape')
  await expect(sheet).toHaveCount(0)

  await page.locator('[data-mode="labo"]').click()
  for (const n of ['01', '02', '03']) {
    await expect(page.locator('.pr-step .pr-n', { hasText: n }).first()).toBeVisible()
  }
})
