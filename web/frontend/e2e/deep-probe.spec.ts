import { test } from '@playwright/test'

// Phase 1 deep probe — chaque surface vivante × chaque viewport + balayage console.
const BASE = process.env.AUDIT_BASE || '' // vide : baseURL du harnais, jamais d'hôte codé en dur
const VIEWPORTS: [number, number, string][] = [[1920, 1080, '1920'], [1366, 768, '1366'], [390, 844, '390']]

// Les quatre « panneaux » supprimés sont devenus deux destinations : Synthèse
// puis les quatre chapitres du Labo (Configurer/Observer/Comparer/Preuves).
const surfaces = [
  { id: 'cockpit', chapter: -1 },
  { id: 'configurer', chapter: 0 },
  { id: 'observer', chapter: 1 },
  { id: 'comparer', chapter: 2 },
  { id: 'preuves', chapter: 3 },
] as const

for (const [w, h, tag] of VIEWPORTS) {
  test(`probe all surfaces @${tag}`, async ({ page }) => {
    test.setTimeout(120000)
    const errs: string[] = []
    page.on('pageerror', e => errs.push(String(e).slice(0, 150)))
    await page.setViewportSize({ width: w, height: h })
    for (const surface of surfaces) {
      await page.goto(BASE + '/')
      if (surface.chapter < 0) {
        await page.locator('[data-mode="cockpit"]').click()
      } else {
        await page.locator('[data-mode="labo"]').click()
        await page.locator('.pr-step').nth(surface.chapter).click()
      }
      await page.waitForTimeout(1500)
      await page.screenshot({ path: `shots/probe-${surface.id}-${tag}.png`, fullPage: true })
    }
    if (errs.length) console.log(`ERRORS-${tag}:`, JSON.stringify([...new Set(errs)].slice(0, 5)))
  })
}
