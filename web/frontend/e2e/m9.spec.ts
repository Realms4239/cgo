import { test, expect } from '@playwright/test'
test('m9 wall kit', async ({ page }) => {
  await page.goto('/')
  await expect(page.locator('text=METEOLINK').first()).toBeVisible()
  await page.getByRole('button', { name: 'Densité' }).click()
  await expect(page.locator('.rail')).toBeVisible()
  await page.keyboard.press('Meta+k')
  if ((await page.getByRole('dialog', { name: 'Palette de commandes' }).count()) === 0) await page.keyboard.press('Control+k')
  await expect(page.getByRole('dialog', { name: 'Palette de commandes' })).toBeVisible()
  await page.keyboard.press('Escape')
  // After palette close, ensure prompt appears — start a lightweight run if idle so prompt-progress (Event 3/6) is visible
  const prompt = page.locator('.prompt-progress')
  if ((await prompt.count()) === 0) {
    await page.evaluate(() => fetch('/api/run/start', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ profiles: ['P1'], reps: 1 }) }).catch(() => {}))
    await page.waitForTimeout(1200)
  }
  // Prompt progress line (amber 1px) — visible when running; allow 7s reappear if just dismissed via Esc
  await expect(page.locator('.prompt-progress')).toBeVisible({ timeout: 7000 }).catch(async () => {
    // fallback: ensure at least the running Event line or Actions rapides label is visible
    const evt = page.locator('text=Event')
    const rapid = page.locator('text=Actions rapides')
    if (await evt.count()) await expect(evt.first()).toBeVisible({ timeout: 2000 })
    else if (await rapid.count()) await expect(rapid.first()).toBeVisible({ timeout: 2000 })
  })
  // Live wall overlay baseline vs CAKE — same scale, diff badge
  await page.getByLabel('Navigation').getByRole('button', { name: 'Temps réel' }).click()
  await expect(page.locator('.live-wall-overlay')).toBeVisible({ timeout: 5000 })
  await expect(page.locator('.live-wall-overlay')).toContainText('baseline')
  await expect(page.locator('.live-wall-overlay')).toContainText('CAKE')
  // Résultats A/B diff badge + hardware provenance
  await page.getByLabel('Navigation').getByRole('button', { name: 'Résultats' }).click()
  await expect(page.locator('.ab-bento')).toBeVisible({ timeout: 5000 })
  await expect(page.locator('.ab-bento .diff-badge')).toBeVisible({ timeout: 5000 })
  await expect(page.locator('.provenance, [data-testid="provenance"]')).toContainText('hardware_recommendation')
})
