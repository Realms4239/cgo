import { test, expect } from '@playwright/test'

// The control-and-shape loop driven through the UI (ARG.md pivot):
// lock a baseline from live rings → apply CAKE to the gateway → live diff.
test('edge control demo', async ({ page }) => {
  test.setTimeout(180000)
  await page.goto('http://192.168.174.128:9090/')
  await page.locator('[data-panel="live"]').click()
  const overlay = page.locator('[data-testid="live-wall-overlay"]')
  await overlay.scrollIntoViewIfNeeded()
  await expect(overlay).toBeVisible()
  await page.evaluate(async () => { await fetch('/api/run/start', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ profiles: ['P1'], reps: 1 }) }) })
  await page.waitForTimeout(15000)
  // Q11 vocabulary — the button is «FIGER L'AVANT», the locked state a Pill
  // whose accessible name is `${label} ${value}`.
  const lockBtn = page.getByRole('button', { name: /FIGER L'AVANT/ })
  await expect(lockBtn).toBeEnabled({ timeout: 30000 })
  await lockBtn.click()
  await page.waitForTimeout(1200)
  await expect(page.getByRole('button', { name: /avant figé/ })).toBeVisible({ timeout: 10000 })
  // the Pill's text content carries the locked value; its accessible name is the tooltip
  await expect(page.getByRole('button', { name: /avant figé/ })).toHaveText(/avant \d+(\.\d+)? ms @/)
  await page.getByRole('button', { name: 'cake', exact: true }).click()
  await expect(page.getByText(/appliqué au bord/)).toBeVisible({ timeout: 15000 })
  await page.waitForTimeout(12000)
  await overlay.scrollIntoViewIfNeeded()
  await page.screenshot({ path: 'shots/control-live-diff.png' })
  const txt = await page.locator('[data-testid="live-wall-overlay"]').textContent()
  console.log('OVERLAY-TEXT:', (txt ?? '').slice(0, 300).replace(/\s+/g, ' '))
})
