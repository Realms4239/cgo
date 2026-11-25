import { expect, test, type Page } from '@playwright/test'

// Deux destinations vivantes seulement : Synthèse (#v-cockpit) et Labo
// (.labo-view). Le binaire Go réel sert l'application : aucun page.route.
const MODES = ['cockpit', 'labo'] as const

async function horizontalOverflow(page: Page): Promise<number> {
  return page.evaluate(() =>
    document.documentElement.scrollWidth - document.documentElement.clientWidth,
  )
}

test('typing 2 in a text field does not switch modes at 890px', async ({ page }) => {
  await page.setViewportSize({ width: 890, height: 768 })
  await page.goto('/')
  await expect(page.locator('#v-cockpit')).toBeVisible()
  await expect(page.locator('[data-mode="cockpit"]')).toHaveAttribute('aria-current', 'page')

  // Le champ de saisie vivant est celui de la palette (Cmd-K). keyboard.type()
  // émet de vrais keydown, contrairement à fill() : c'est le garde-fou
  // isTextEntryTarget qui est exercé, pas seulement la valeur du champ.
  await page.keyboard.press('Control+k')
  const input = page.locator('.palette-input')
  await expect(input).toBeVisible()
  await input.focus()
  await page.keyboard.type('site 2')
  await expect(input).toHaveValue('site 2')
  await expect(page.locator('[data-mode="cockpit"]'), 'la saisie ne bascule pas en Labo')
    .toHaveAttribute('aria-current', 'page')
  await expect(page.locator('#v-labo'), 'aucune vue Labo montée').toHaveCount(0)
  await page.keyboard.press('Escape')
  await expect(page.locator('#palette')).toHaveCount(0)
  await expect(page.locator('[data-mode="cockpit"]')).toHaveAttribute('aria-current', 'page')
})

for (const viewport of [{ width: 890, height: 768 }, { width: 1920, height: 1080 }]) {
  test(`no horizontal overflow at ${viewport.width}×${viewport.height}`, async ({ page }) => {
    await page.setViewportSize(viewport)
    await page.goto('/')
    for (const mode of MODES) {
      await page.locator(`[data-mode="${mode}"]`).click()
      const view = mode === 'cockpit' ? page.locator('#v-cockpit') : page.locator('.labo-view')
      await expect(view).toBeVisible()
      await expect.poll(() => horizontalOverflow(page)).toBeLessThanOrEqual(1)
    }
  })
}
