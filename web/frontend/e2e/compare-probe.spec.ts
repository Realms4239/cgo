import { test, expect } from '@playwright/test'

test('compare view probe', async ({ page }) => {
  test.setTimeout(90000)
  const errs: string[] = []
  page.on('pageerror', e => errs.push(String(e).slice(0, 200)))
  await page.goto('/')
  await page.locator('[data-mode="labo"]').click()
  await page.locator('.pr-step').nth(2).click()
  // le classement de l'étape Comparer — `.data-table` n'existe plus : la table
  // vivante est `.rank-table`, unique dans `#chapter-comparer`.
  const board = page.locator('#chapter-comparer .rank-table')
  await expect(board).toBeVisible({ timeout: 8000 })
  // épingler les deux lignes les plus parlantes (première + dernière)
  // les non mesurés sont classés en dernier et refusent le duel : on épingle
  // la première et la dernière ligne *mesurée*, sans inventer de valeur.
  await expect.poll(() => board.locator('tbody tr.rank-row').count()).toBeGreaterThan(1)
  const pins = board.locator('tbody tr.rank-row .rank-pin').filter({ hasText: /^\d+$/ })
  expect(await pins.count()).toBeGreaterThanOrEqual(2)
  await pins.first().click()
  await pins.last().click()
  await expect(page.locator('.duel[data-state="compared"]')).toBeVisible({ timeout: 10000 })
  await page.screenshot({ path: 'shots/compare-view.png', fullPage: true })
  const txt = (await page.locator('.duel').textContent() ?? '').replace(/\s+/g, ' ')
  console.log('COMPARE:', txt.slice(0, 260))
  if (errs.length) console.log('ERRORS:', JSON.stringify(errs.slice(0, 3)))
})
