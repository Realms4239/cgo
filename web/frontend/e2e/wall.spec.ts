import { test, expect } from '@playwright/test'

// Couche 5 Tâche 5.2 — clip mur : overlay hero + MetricCard HD sélective, capture 1920x400.
test('wall clip — live-wall-overlay + small_p95 metric card', async ({ page }) => {
  await page.goto('/')
  await page.locator('[data-panel="live"]').click()
  await expect(page.locator('[data-testid="live-wall-overlay"]')).toBeVisible()
  await expect(page.locator('[data-metric="small_p95"]')).toBeVisible()
  await page.screenshot({ path: 'shots/wall-hero-1920.png', clip: { x: 0, y: 0, width: 1920, height: 400 } })
})

test('wall MetricCards — 8 cards with data-metric render', async ({ page }) => {
  await page.goto('/')
  await page.locator('[data-panel="live"]').click()
  for (const m of ['rtt_p50', 'rtt_p95', 'small_p95', 'bulk_goodput', 'drops', 'wasted', 'cost_ar_per_h', 'deadline_ok']) {
    await expect(page.locator(`[data-metric="${m}"]`)).toBeVisible()
  }
})

test('wall vide honnête — cartes repos montrent — pas des 0 fabriqués', async ({ page }) => {
  await page.goto('/')
  await page.locator('[data-panel="live"]').click()
  const small = page.locator('[data-metric="small_p95"]')
  await expect(small).toBeVisible()
  // serveur au repos (SSE running:false) — la carte doit afficher — selon la frontière de vérité
  await expect(small).toContainText('—')
})
