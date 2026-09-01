import { test } from '@playwright/test'

// Phase 1 deep probe — every view × every viewport + console sweep.
const BASE = process.env.AUDIT_BASE || 'http://localhost:9090'
const VIEWPORTS: [number, number, string][] = [[1920, 1080, '1920'], [1366, 768, '1366'], [390, 844, '390']]

for (const [w, h, tag] of VIEWPORTS) {
  test(`probe all panels @${tag}`, async ({ page }) => {
    test.setTimeout(120000)
    const errs: string[] = []
    page.on('pageerror', e => errs.push(String(e).slice(0, 150)))
    await page.setViewportSize({ width: w, height: h })
    for (const p of ['campagne', 'live', 'resultats', 'integrite']) {
      await page.goto(BASE + '/')
      await page.locator(`[data-panel="${p}"]`).click()
      await page.waitForTimeout(1500)
      await page.screenshot({ path: `shots/probe-${p}-${tag}.png`, fullPage: true })
    }
    if (errs.length) console.log(`ERRORS-${tag}:`, JSON.stringify([...new Set(errs)].slice(0, 5)))
  })
}
