import { expect, test } from '@playwright/test'

const VIEWS = [
  { label: 'Tableau de bord', selector: '#v-dashboard' },
  { label: 'Auditer', selector: '#v-audit' },
  { label: "Banc d'essai", selector: '#v-bench' },
  { label: 'Preuves & Rapports', selector: '#v-evidence' },
] as const

test.beforeEach(async ({ page }) => {
  await page.route('**/api/audit/status', route => route.fulfill({ json: { running: false, report: null } }))
  await page.route('**/api/audit/report', route => route.fulfill({ status: 404, body: 'not found' }))
  await page.route('**/api/audit/list', route => route.fulfill({ json: { audits: [] } }))
  await page.route('**/api/audit/presets', route => route.fulfill({ json: [] }))
  await page.route('**/api/stream', route => route.fulfill({ status: 200, contentType: 'text/event-stream', body: '' }))
  // Le backend local n’est pas joignable : les autres API sont des absences, pas du vide.
  await page.route('**/api/**', route => route.fulfill({ json: {} }))
})

test('typing 2 in Site does not switch modes at 890px', async ({ page }) => {
  await page.setViewportSize({ width: 890, height: 768 })
  await page.goto('/')
  await page.getByRole('button', { name: 'Auditer' }).click()
  const site = page.getByLabel('Site')
  await expect(site).toBeVisible()
  await site.fill('site 2')
  await expect(site).toHaveValue('site 2')
  await expect(page.getByRole('button', { name: 'Auditer' })).toHaveAttribute('aria-current', 'page')
})

for (const viewport of [{ width: 890, height: 768 }, { width: 1920, height: 1080 }]) {
  test(`no horizontal overflow at ${viewport.width}×${viewport.height}`, async ({ page }) => {
    await page.setViewportSize(viewport)
    await page.goto('/')
    for (const view of VIEWS) {
      await page.getByRole('button', { name: view.label }).click()
      await expect(page.locator(view.selector)).toBeVisible()
      await expect.poll(() => page.evaluate(() =>
        document.documentElement.scrollWidth - document.documentElement.clientWidth,
      )).toBeLessThanOrEqual(1)
    }
  })
}
