// UNMOCKED — aucun page.route : le vrai binaire cgo sert la SPA et /api/*.
import AxeBuilder from '@axe-core/playwright'
import { expect, test } from '@playwright/test'

for (const width of [390, 1920]) {
  test(`axe wcag 2.2 aa @ ${width}px`, async ({ page }) => {
    await page.setViewportSize({ width, height: 844 })
    await page.goto('/')
    await page.waitForSelector('main#main', { state: 'visible' })
    await expect(page.locator('main#main')).toBeVisible()
    const results = await new AxeBuilder({ page })
      .withTags(['wcag2a', 'wcag2aa', 'wcag21aa', 'wcag22aa'])
      .analyze()
    const serious = results.violations.filter((v) => v.impact === 'critical' || v.impact === 'serious')
    expect(serious.map((v) => `${v.id}: ${v.help}`)).toEqual([])
  })
}
