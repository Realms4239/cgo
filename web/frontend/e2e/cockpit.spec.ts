import { expect, test } from '@playwright/test'

// Contrat DU COCKPIT (4 modes) — les specs legacy data-panel= sont antérieures
// à la refonte ; ce fichier couvre l'interface livrée.

// le runner de ce poste est lourd (vidéo + trace) : on ne veut pas un
// faux négatif de lenteur sur un rendu client.
test.describe.configure({ timeout: 120000 })

const NAV = { dashboard: 'Tableau de bord', audit: 'Auditer', bench: "Banc d'essai", evidence: 'Preuves & Rapports' }

async function mockApis(page: any) {
  await page.route('**/api/**', route => {
    const url = new URL(route.request().url())
    const path = url.pathname
    if (path.endsWith('/api/health')) {
      return route.fulfill({ json: { ok: true, version: '1.3.5', mode: 'observe', telemetry: 'eBPF' } })
    }
    if (path.endsWith('/api/doctor')) {
      return route.fulfill({
        json: {
          mode: 'observe',
          checks: [
            { name: 'os', status: 'ok', detail: 'linux/amd64 → mode observe' },
            { name: 'tc', status: 'warn', detail: 'iproute2 absent — façonnage impossible' },
          ],
        },
      })
    }
    if (path.endsWith('/api/integrity')) {
      return route.fulfill({
        json: { available: true, runs: 1, manifests: 1, valid: 12, degraded: 0, quarantined: 0, hash8: '9f3a2c1e', updated: '2026-11-10T09:00:00Z' },
      })
    }
    if (path.endsWith('/api/stream')) {
      return route.fulfill({ status: 200, contentType: 'text/event-stream', body: '' })
    }
    return route.fulfill({ json: {} })
  })
}

test('les quatre destinations existent et l\'une d\'elle est active', async ({ page }) => {
  await mockApis(page)
  await page.goto('/', { waitUntil: 'domcontentloaded' })
  const nav = page.getByLabel('Destinations')
  for (const label of Object.values(NAV)) {
    await expect(nav.getByRole('button', { name: label })).toBeVisible({ timeout: 15000 })
  }
  await expect(nav.locator('button.on')).toHaveCount(1)
})

test('chaque destination s\'ouvre et reste lisible', async ({ page }) => {
  await mockApis(page)
  await page.goto('/', { waitUntil: 'domcontentloaded' })
  const nav = page.getByLabel('Destinations')
  for (const [id, label] of Object.entries(NAV)) {
    await nav.getByRole('button', { name: label }).click()
    await expect(page.locator(`#v-${id}`)).toBeVisible({ timeout: 15000 })
  }
})

test('pas de débordement horizontal à 890, 1366 et 390', async ({ page }) => {
  await mockApis(page)
  for (const [w, h] of [[890, 768], [1366, 768], [390, 844]]) {
    await page.setViewportSize({ width: w, height: h })
    await page.goto('/', { waitUntil: 'domcontentloaded' })
    const overflow = await page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth)
    expect(overflow, `débordement horizontal à ${w}px`).toBeLessThanOrEqual(1)
  }
})

test('l\'estampille du poste est lisible au pied de page', async ({ page }) => {
  await mockApis(page)
  await page.goto('/', { waitUntil: 'domcontentloaded' })
  await expect(page.locator('.ft-post')).toContainText('cgo 1.3.5', { timeout: 15000 })
  await expect(page.locator('.ft-post')).toContainText('observe', { timeout: 15000 })
})

test('l\'état du poste est lisible dans les preuves', async ({ page }) => {
  await mockApis(page)
  await page.goto('/', { waitUntil: 'domcontentloaded' })
  await page.getByLabel('Destinations').getByRole('button', { name: 'Preuves & Rapports' }).click()
  const card = page.getByTestId('machine-card')
  await expect(card).toBeVisible({ timeout: 15000 })
  await expect(card).toContainText('cgo 1.3.5', { timeout: 15000 })
  await expect(card.locator('.int-post-check[data-status="ok"]')).toHaveCount(1)
  await expect(card.locator('.int-post-check[data-status="warn"]')).toHaveCount(1)
})
