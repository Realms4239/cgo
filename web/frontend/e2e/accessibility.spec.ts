import { expect, test } from '@playwright/test'
import { contrastRatio, wcagAaNormalText } from '../src/lib/a11y'
import { ABSENT_REPORT, ARCHIVE_179_ROWS, ARCHIVE_ONE_ROW, CORRUPT_PAYLOAD, VALID_REPORT } from './fixtures'

async function mockApis(page: any, options: { listStatus?: number; listBody?: string; reportStatus?: number; reportBody?: string } = {}) {
  await page.route('**/api/**', route => {
    const url = new URL(route.request().url())
    const path = url.pathname
    if (path.endsWith('/api/audit/status') || path.endsWith('/api/audit/live')) {
      return route.fulfill({ json: { running: false, report: null } })
    }
    if (path.endsWith('/api/audit/list')) {
      return route.fulfill({
        status: options.listStatus ?? 200,
        contentType: 'application/json',
        body: options.listBody ?? ARCHIVE_ONE_ROW,
      })
    }
    if (path.endsWith('/api/audit/report')) {
      if (options.reportStatus === 404) {
        return route.fulfill({ status: 404, body: 'not found' })
      }
      if (options.reportBody === CORRUPT_PAYLOAD) {
        return route.fulfill({ status: 200, contentType: 'application/json', body: CORRUPT_PAYLOAD })
      }
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: options.reportBody ?? VALID_REPORT,
      })
    }
    if (path.endsWith('/api/audit/presets')) return route.fulfill({ json: [] })
    if (path.endsWith('/api/stream')) {
      return route.fulfill({ status: 200, contentType: 'text/event-stream', body: '' })
    }
    return route.fulfill({ json: {} })
  })
}

test('no report at all stays measured-absent, not green', async ({ page }) => {
  test.setTimeout(60000)
  await mockApis(page, { reportStatus: 404 })
  await page.setViewportSize({ width: 890, height: 768 })
  await page.goto('/', { waitUntil: 'domcontentloaded' })
  const absent = page.locator('[data-testid="dash-absent"]')
  await expect(absent).toBeVisible()
  await expect(absent).not.toContainText(/stable|sain|OK/i)
})

test('absent evidence cannot render green proof language', async ({ page }) => {
  await mockApis(page, { reportBody: ABSENT_REPORT })
  await page.setViewportSize({ width: 890, height: 768 })
  await page.goto('/', { waitUntil: 'domcontentloaded' })
  await page.getByRole('button', { name: 'Preuves & Rapports' }).click()
  await page.getByRole('button', { name: 'Ouvrir' }).press('Enter')
  await expect(page.getByRole('region', { name: 'Rapport ouvert' })).toBeVisible()
  await expect(page.locator('body')).not.toContainText(/PREUVE VÉRIFIABLE|prouvé|Vérifiable/i)
  await expect(page.getByText('Preuve manquante').first()).toBeVisible()
  await expect(page.getByRole('button', { name: 'CITER CE RAPPORT' })).toBeDisabled()
})

test('failure state has a recovery action and adequate contrast', async ({ page }) => {
  await mockApis(page, { listStatus: 500 })
  await page.setViewportSize({ width: 890, height: 768 })
  await page.goto('/', { waitUntil: 'domcontentloaded' })
  await page.getByRole('button', { name: 'Preuves & Rapports' }).click()
  const retry = page.getByTestId('audit-reports').getByRole('button', { name: 'Réessayer' })
  await expect(retry).toBeVisible()
  const error = page.getByTestId('audit-reports').locator('.state-error')
  const { foreground, background } = await error.evaluate(el => ({
    foreground: getComputedStyle(el).color,
    background: getComputedStyle(el).backgroundColor,
  }))
  expect(wcagAaNormalText(contrastRatio(foreground, background)), 'state-error must meet AA normal text').toBe(true)
})

test('keyboard reaches the archive and opens the valid report', async ({ page }) => {
  test.setTimeout(60000)
  await mockApis(page)
  await page.setViewportSize({ width: 890, height: 768 })
  await page.goto('/', { waitUntil: 'domcontentloaded' })

  await page.keyboard.press('Tab')
  await expect(page.locator('.skip-link')).toBeFocused()
  for (const mode of ['dashboard', 'audit', 'bench', 'evidence']) {
    await page.keyboard.press('Tab')
    await expect(page.locator(`[data-mode="${mode}"]`)).toBeFocused()
  }

  await page.keyboard.press('Enter')
  await expect(page.locator('#v-evidence')).toBeVisible()
  const open = page.getByRole('button', { name: 'Ouvrir' })
  await expect(open).toBeVisible()
  await open.focus()
  await page.keyboard.press('Enter')
  await expect(page.getByRole('region', { name: 'Rapport ouvert' })).toBeVisible()
})

test('archive window keeps aria row count and limits mounted rows', async ({ page }) => {
  test.setTimeout(60000)
  await mockApis(page, { listBody: ARCHIVE_179_ROWS })
  await page.setViewportSize({ width: 890, height: 768 })
  await page.goto('/', { waitUntil: 'domcontentloaded' })
  await page.getByRole('button', { name: 'Preuves & Rapports' }).click()
  const table = page.locator('.archive-table')
  await expect(table).toHaveAttribute('aria-rowcount', '179')
  const mounted = await table.locator('tbody tr').count()
  expect(mounted).toBeGreaterThan(0)
  expect(mounted).toBeLessThanOrEqual(30)
})

test('corrupt evidence fails to open', async ({ page }) => {
  await mockApis(page, { reportBody: CORRUPT_PAYLOAD })
  await page.setViewportSize({ width: 890, height: 768 })
  await page.goto('/', { waitUntil: 'domcontentloaded' })
  await page.getByRole('button', { name: 'Preuves & Rapports' }).click()
  await page.getByRole('button', { name: 'Ouvrir' }).click()
  await expect(page.getByText('rapport indisponible — réessayez')).toBeVisible()
  await expect(page.locator('body')).not.toContainText(/PREUVE VÉRIFIABLE/i)
})

for (const viewport of [{ width: 890, height: 768 }, { width: 1920, height: 1080 }]) {
  test(`no horizontal overflow at ${viewport.width}×${viewport.height}`, async ({ page }) => {
    await mockApis(page)
    test.setTimeout(60000)
    await page.setViewportSize(viewport)
    await page.goto('/', { waitUntil: 'domcontentloaded' })
    for (const label of ['Tableau de bord', 'Auditer', "Banc d'essai", 'Preuves & Rapports']) {
      await page.getByRole('button', { name: label }).click()
      await expect.poll(() => page.evaluate(() =>
        document.documentElement.scrollWidth - document.documentElement.clientWidth,
      )).toBeLessThanOrEqual(1)
    }
  })
}
