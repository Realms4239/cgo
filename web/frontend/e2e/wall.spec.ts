import { test, expect } from '@playwright/test'

// Couche 5 Tâche 5.2 — le mur quatre-vues et ses MetricCard de mesure
// n'existent plus. Vérifié contre le binaire réel : `[data-testid="live-wall-overlay"]`
// et les huit `[data-metric="…"]` (rtt_p50 … deadline_ok) ont 0 occurrence ;
// le Labo expose seulement trois cartes `[data-metric]` aux libellés
// « Campagnes archivées / Mesures valides / Mesures écartées ».
// Assertions conservées mot pour mot, non exécutées. Voir task-1-report.md.

// BLOQUÉ — `[data-testid="live-wall-overlay"]` : 0 occurrence ; aucun overlay
// équivalent dans `#v-cockpit` ni dans `#chapter-observer`.
test.skip('wall clip — live-wall-overlay + small_p95 metric card', async ({ page }) => {
  await page.goto('/')
  await page.locator('[data-mode="labo"]').click()
  await page.locator('.pr-step').nth(1).click()
  await expect(page.locator('[data-testid="live-wall-overlay"]')).toBeVisible()
  await expect(page.locator('[data-metric="small_p95"]')).toBeVisible()
  await page.screenshot({ path: 'shots/wall-hero-1920.png', clip: { x: 0, y: 0, width: 1920, height: 400 } })
})

// BLOQUÉ — les huit `[data-metric]` de mesure du mur n'ont pas d'équivalent
// vivant (3 cartes d'archives seulement, et non ces huit termes).
test.skip('wall MetricCards — 8 cards with data-metric render', async ({ page }) => {
  await page.goto('/')
  await page.locator('[data-mode="labo"]').click()
  await page.locator('.pr-step').nth(1).click()
  for (const m of ['rtt_p50', 'rtt_p95', 'small_p95', 'bulk_goodput', 'drops', 'wasted', 'cost_ar_per_h', 'deadline_ok']) {
    await expect(page.locator(`[data-metric="${m}"]`)).toBeVisible()
  }
})

// BLOQUÉ — `[data-metric="small_p95"]` : 0 occurrence. Aucune carte vivante
// n'affiche ce terme, donc pas de frontière de vérité « — » à vérifier ici.
test.skip('wall vide honnête — cartes repos montrent — pas des 0 fabriqués', async ({ page }) => {
  await page.goto('/')
  await page.locator('[data-mode="labo"]').click()
  await page.locator('.pr-step').nth(1).click()
  const small = page.locator('[data-metric="small_p95"]')
  await expect(small).toBeVisible()
  // serveur au repos (SSE running:false) — la carte doit afficher — selon la frontière de vérité
  await expect(small).toContainText('—')
})
