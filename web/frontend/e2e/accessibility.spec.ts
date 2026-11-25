import { expect, test, type Page, type Route } from '@playwright/test'
import { contrastRatio, wcagAaNormalText } from '../src/lib/a11y'
import { ABSENT_REPORT, ARCHIVE_179_ROWS, ARCHIVE_ONE_ROW, CORRUPT_PAYLOAD, VALID_REPORT } from './fixtures'

// Cet hôte tourne en mode observation : /api/audit/list y est vide et aucun
// dossier scellé n'est disponible. Pas de page.route('**/api/**') global : les
// scénarios qui exigent une ligne scellée posent un stub LOCAL, uniquement sur
// la requête concernée, avec un jeu déterministe de fixtures. Aucune mesure
// n'est fabriquée dans le corpus gelé et aucun banc live n'est lancé.
async function stubArchiveList(page: Page, body: string): Promise<void> {
  await page.route(
    (url) => url.pathname === '/api/audit/list',
    (route: Route) => route.fulfill({ status: 200, contentType: 'application/json', body }),
  )
}

// Seuls les appels portant un audit_id sont stubbés : la restaure au montage
// (/api/audit/report sans paramètre) doit rester joignable au serveur réel.
async function stubSealedDossier(page: Page, body: string, status = 200): Promise<void> {
  await page.route(
    (url) => url.pathname === '/api/audit/report' && url.searchParams.has('audit_id'),
    (route: Route) => route.fulfill({
      status,
      contentType: status === 200 ? 'application/json' : 'text/plain',
      body,
    }),
  )
}

async function goToPreuves(page: Page): Promise<void> {
  await page.goto('/')
  await page.locator('[data-mode="labo"]').click()
  await expect(page.locator('.labo-view')).toBeVisible()
  await page.locator('.pr-step').nth(3).click()
  await expect(page.locator('.pr-step').nth(3)).toHaveAttribute('aria-current', 'step')
  await expect(page.locator('#chapter-preuves .preuves')).toBeVisible()
}

test('no report at all stays measured-absent, not green', async ({ page }) => {
  test.setTimeout(60000)
  await page.setViewportSize({ width: 890, height: 768 })
  await page.goto('/')
  await expect(page.locator('#v-cockpit')).toBeVisible()
  const band = page.locator('.answer-band')
  await expect(band).toBeVisible()
  await expect(page.locator('#verdict-word')).toHaveText('MESURE ABSENTE')
  await expect(page.locator('.band-number-absent')).toHaveText('— non mesuré')
  await expect(page.locator('.band-status')).toHaveText('Non mesurée')
  await expect(band).not.toContainText(/stable|sain|OK/i)
})

test('absent evidence cannot render green proof language', async ({ page }) => {
  await stubArchiveList(page, ARCHIVE_ONE_ROW)
  await stubSealedDossier(page, ABSENT_REPORT)
  await goToPreuves(page)

  await expect(page.locator('#chapter-preuves .corpus-line')).toContainText('Preuve manquante')
  await page.locator('#chapter-preuves .archive-table').getByRole('button', { name: 'Ouvrir' }).press('Enter')
  const drawer = page.locator('.proof-drawer[role="dialog"]')
  await expect(drawer).toBeVisible()
  await expect(drawer).toContainText('Preuve manquante')
  // Portée au dossier ouvert : ailleurs dans le Labo, « prouvé en lab » est une
  // recommandation matérielle, pas une affirmation de preuve scellée.
  await expect(drawer).not.toContainText(/PREUVE VÉRIFIABLE|prouvé|Vérifiable/i)

  await page.locator('.proof-close').click()
  await expect(drawer).toHaveCount(0)
  await page.locator('#chapter-preuves .archive-table').getByRole('button', { name: 'CITER' }).click()
  await expect(page.locator('#chapter-preuves .archive-citation'))
    .toHaveText('citation impossible — empreinte source absente')
})

test('failure state has a recovery action and adequate contrast', async ({ page }) => {
  // Stub local : seul un dossier absent permet d'exercer l'état d'échec du
  // tiroir preuve sur un hôte qui ne sert aucun dossier scellé.
  await stubArchiveList(page, ARCHIVE_ONE_ROW)
  await stubSealedDossier(page, 'dossier introuvable', 404)
  await goToPreuves(page)

  const open = page.locator('#chapter-preuves .archive-table').getByRole('button', { name: 'Ouvrir' })
  await open.click()
  const drawer = page.locator('.proof-drawer[role="dialog"]')
  await expect(drawer.getByRole('status')).toHaveText('Preuve manquante — dossier scellé illisible.')

  const retry = drawer.locator('.proof-close')
  await expect(retry).toBeVisible()
  await expect(retry).toBeEnabled()
  const panel = drawer.locator('.proof-drawer-panel')
  const { foreground, background } = await panel.evaluate(el => ({
    foreground: getComputedStyle(el).color,
    background: getComputedStyle(el).backgroundColor,
  }))
  expect(wcagAaNormalText(contrastRatio(foreground, background)), 'surface d’échec en AA').toBe(true)

  await retry.click()
  await expect(drawer).toHaveCount(0)
  await open.click()
  await expect(drawer).toBeVisible()
})

test('keyboard reaches the archive and opens the dossier', async ({ page }) => {
  await stubArchiveList(page, ARCHIVE_ONE_ROW)
  await stubSealedDossier(page, VALID_REPORT)
  await page.setViewportSize({ width: 890, height: 768 })
  await page.goto('/')

  await page.keyboard.press('Tab')
  await expect(page.locator('.skip-link')).toBeFocused()
  for (const mode of ['cockpit', 'labo']) {
    await page.keyboard.press('Tab')
    await expect(page.locator(`[data-mode="${mode}"]`)).toBeFocused()
  }
  await page.keyboard.press('Enter')
  await expect(page.locator('.labo-view')).toBeVisible()

  await page.locator('.pr-step').nth(3).focus()
  await page.keyboard.press('Enter')
  await expect(page.locator('.pr-step').nth(3)).toHaveAttribute('aria-current', 'step')

  const open = page.locator('#chapter-preuves .archive-table').getByRole('button', { name: 'Ouvrir' })
  await expect(open).toBeVisible()
  await open.focus()
  await page.keyboard.press('Enter')
  await expect(page.locator('.proof-drawer[role="dialog"]')).toBeVisible()
})

test('archive window keeps aria row count and limits mounted rows', async ({ page }) => {
  test.setTimeout(60000)
  // Stub local : 179 lignes déterministes pour prouver la fenêtre virtualisée ;
  // l'hôte observe ne sert aucun dossier scellé (voir fullstack-workflow.spec.ts).
  await stubArchiveList(page, ARCHIVE_179_ROWS)
  await goToPreuves(page)

  const table = page.locator('#chapter-preuves .archive-table')
  await expect(table).toHaveAttribute('aria-rowcount', '179')
  const mounted = await table.locator('tbody tr').count()
  expect(mounted).toBeGreaterThan(0)
  expect(mounted).toBeLessThanOrEqual(30)
})

test('corrupt evidence fails to open', async ({ page }) => {
  await stubArchiveList(page, ARCHIVE_ONE_ROW)
  await stubSealedDossier(page, CORRUPT_PAYLOAD)
  await goToPreuves(page)

  await page.locator('#chapter-preuves .archive-table').getByRole('button', { name: 'Ouvrir' }).click()
  const drawer = page.locator('.proof-drawer[role="dialog"]')
  await expect(drawer.getByRole('status')).toHaveText('Preuve manquante — dossier scellé illisible.')
  await expect(drawer).not.toContainText(/PREUVE VÉRIFIABLE/i)
})

for (const viewport of [{ width: 890, height: 768 }, { width: 1920, height: 1080 }]) {
  test(`no horizontal overflow at ${viewport.width}×${viewport.height}`, async ({ page }) => {
    test.setTimeout(60000)
    await page.setViewportSize(viewport)
    await page.goto('/')
    for (const mode of ['cockpit', 'labo']) {
      await page.locator(`[data-mode="${mode}"]`).click()
      const view = mode === 'cockpit' ? page.locator('#v-cockpit') : page.locator('.labo-view')
      await expect(view).toBeVisible()
      await expect.poll(() => page.evaluate(() =>
        document.documentElement.scrollWidth - document.documentElement.clientWidth,
      )).toBeLessThanOrEqual(1)
    }
  })
}
