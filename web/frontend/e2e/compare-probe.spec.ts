import { test, expect } from '@playwright/test'
const BASE = process.env.AUDIT_BASE || 'https://localhost:9090'
test('compare view probe', async ({ page }) => {
  test.setTimeout(90000)
  const errs: string[] = []
  page.on('pageerror', e => errs.push(String(e).slice(0, 200)))
  await page.goto(BASE + '/')
  await page.locator('[data-panel="resultats"]').click()
  // le classement de la vue Résultats — .data-table seul est ambigu depuis
  // que matrice (campagne) et runs/quarantaine (intégrité) en portent aussi
  const board = page.locator('#v-resultats .data-table')
  await expect(board).toBeVisible({ timeout: 8000 })
  // épingler les deux lignes les plus parlantes (première + dernière)
  const rows = board.locator('tbody tr')
  await rows.first().locator('button', { hasText: 'A' }).click()
  await rows.last().locator('button', { hasText: 'B' }).click()
  await expect(page.locator('[data-testid="compare-view"]')).toBeVisible({ timeout: 10000 })
  await page.screenshot({ path: 'shots/compare-view.png', fullPage: true })
  const txt = (await page.locator('[data-testid="compare-view"]').textContent() ?? '').replace(/\s+/g, ' ')
  console.log('COMPARE:', txt.slice(0, 260))
  if (errs.length) console.log('ERRORS:', JSON.stringify(errs.slice(0, 3)))
})
