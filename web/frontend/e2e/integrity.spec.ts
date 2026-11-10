import { test, expect } from '@playwright/test'

// Provenance gelée — le panneau Intégrité vit DANS « Preuves & Rapports »
// (vue embarquée, plus une destination). Vocabulaire Q11 : « RDF — provenance scellée ».
// Honête : la pastille sha256 réelle quand /api/integrity l'expose, sinon EmptyState.

async function mockApis(page: any, integrity: Record<string, unknown>) {
  await page.route('**/api/**', route => {
    const path = new URL(route.request().url()).pathname
    if (path.endsWith('/api/integrity')) return route.fulfill({ json: integrity })
    if (path.endsWith('/api/audit/status') || path.endsWith('/api/audit/live')) {
      return route.fulfill({ json: { running: false, report: null } })
    }
    if (path.endsWith('/api/audit/list')) return route.fulfill({ json: { audits: [] } })
    if (path.endsWith('/api/stream')) return route.fulfill({ status: 200, contentType: 'text/event-stream', body: '' })
    if (path.endsWith('/api/replay/list')) return route.fulfill({ json: { runs: [] } })
    return route.fulfill({ json: {} })
  })
}

test('intégrité — la provenance gelée s’affiche avec le hash exposé', async ({ page }) => {
  await mockApis(page, { available: true, runs: 159, manifests: 159, valid: 208, degraded: 74, quarantined: 246, sha256: 'deadbeefcafe1234', hash8: 'deadbeef', updated: '2026-10-31T12:00:00Z' })
  await page.goto('/')
  await page.getByRole('button', { name: 'Preuves & Rapports' }).click()
  await page.locator('[data-testid="corpus-details"] > summary').click()
  await expect(page.getByText('RDF — provenance scellée')).toBeVisible()
  await expect(page.locator('[data-testid="provenance-hash"]')).toContainText('deadbeefcafe1234')
})

test('intégrité — sans sha exposé, EmptyState honnête plutôt que hint fabriqué', async ({ page }) => {
  await mockApis(page, { available: true, runs: 159, manifests: 159, valid: 208, sha256: '', hash8: '' })
  await page.goto('/')
  await page.getByRole('button', { name: 'Preuves & Rapports' }).click()
  await page.locator('[data-testid="corpus-details"] > summary').click()
  await expect(page.getByText('RDF — provenance scellée')).toBeVisible()
  await expect(page.getByText('empreinte SHA non exposée')).toBeVisible()
})

test('banc — tri d’affichage conservé (groupes m / source c)', async ({ page }) => {
  await mockApis(page, { available: true })
  await page.goto('/')
  await page.getByRole('button', { name: "Banc d'essai" }).click()
  const controls = page.locator('.display-controls')
  await expect(controls).toBeVisible()
  await expect(controls).toHaveAttribute('data-tri', /m/)
  await expect(controls).toHaveAttribute('data-tri', /c/)
})
