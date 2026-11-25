import { test, expect, type Locator } from '@playwright/test'

// Sonde santé graphes — chaque surface vivante se rend, zéro erreur console,
// les coutures d'instrumentation rapportent le flux. Journal après-debug (plan §5).
// Les trois canvases du mur supprimé sont devenus un champ temporel partagé
// par destination : Synthèse et Labo/Observer chacun leur canvas ECharts.

const errors: string[] = []

async function nonBlank(canvas: Locator): Promise<boolean> {
  return canvas.evaluate((c: HTMLCanvasElement) => {
    try {
      const ctx = c.getContext('2d')
      if (!ctx || c.width === 0 || c.height === 0) return false
      const d = ctx.getImageData(0, 0, c.width, c.height).data
      for (let p = 3; p < d.length; p += 40) if (d[p] !== 0) return true
      return false
    } catch { return false }
  })
}

test('surfaces — each live ECharts canvas renders non-blank + seams + no console errors', async ({ page }) => {
  page.on('console', m => { if (m.type() === 'error') errors.push(m.text()) })
  page.on('pageerror', e => errors.push(String(e)))
  await page.goto('/')

  // Synthèse : son champ temporel est monté au premier rendu.
  const synth = page.locator('#v-cockpit .time-field canvas')
  await expect(synth).toHaveCount(1)
  await page.waitForTimeout(2500)
  expect(await nonBlank(synth), 'synthèse canvas non-blank').toBe(true)

  // Labo/Observer : le même instrument est monté à la visite du chapitre.
  await page.locator('[data-mode="labo"]').click()
  await page.locator('.pr-step').nth(1).click()
  const observe = page.locator('#chapter-observer .time-field canvas')
  await expect(observe).toHaveCount(1)
  // le chapitre vient d'être monté : ECharts amorce la grille puis se redimensionne
  await expect.poll(() => nonBlank(observe), { timeout: 8000 }).toBe(true)

  // les coutures d'instrumentation exposent le chemin live
  const seam = await page.evaluate(() => ({
    sse: (window as any).__CGO_SSE?.frameCount ?? -1,
    liveMax: (window as any).__CGO_LIVE?.max ?? -1,
    phase: (window as any).__CGO_LIVE?.phase ?? '?',
  }))
  expect(seam.sse).toBeGreaterThan(0)
  expect(seam.liveMax).toBe(1800)

  expect(errors, 'console/page errors').toEqual([])
})

// BLOQUÉ — cibles supprimées par la refonte deux destinations :
// `[data-testid="rank-verdict"]` : 0 occurrence dans le DOM réel ;
// `.panel-stack` : 0 occurrence ; aucun canvas de dispersion dans
// `#chapter-comparer` (qui rend `.rank-table`, sans ECharts).
// Assertion conservée mot pour mot, non exécutée. Voir task-1-report.md.
test.skip('resultats — scatter canvas renders non-blank with real groups', async ({ page }) => {
  page.on('pageerror', e => errors.push(String(e)))
  await page.goto('/')
  await page.locator('[data-mode="labo"]').click()
  await page.locator('.pr-step').nth(2).click()
  await expect(page.locator('[data-testid="rank-verdict"]')).toBeVisible()
  const canvas = page.locator('.panel-stack canvas').last()
  await expect(canvas).toHaveCount(1)
  const ok = await canvas.evaluate((c: HTMLCanvasElement) => {
    try {
      const ctx = c.getContext('2d')
      if (!ctx || c.width === 0) return false
      const d = ctx.getImageData(0, 0, Math.min(c.width, 400), Math.min(c.height, 160)).data
      for (let p = 3; p < d.length; p += 40) if (d[p] !== 0) return true
      return false
    } catch { return false }
  })
  expect(ok, 'scatter non-blank').toBe(true)
})

test('campagne — timeline bands present when running phase data exists', async ({ page }) => {
  await page.goto('/')
  await page.locator('[data-mode="labo"]').click()
  await page.locator('.pr-step').nth(1).click()
  // Timeline conditionnée aux données (honnête) : bandes réelles ou état « absent »,
  // jamais un bloc vide. L'ancien `[data-testid="timeline"]` est remplacé par
  // `.labo-timeline` (CampaignTimeline), seul chronomètre monté dans le Labo.
  const tl = page.locator('#chapter-observer .labo-timeline')
  await expect(tl).toBeVisible()
  if (await tl.getAttribute('data-state') !== 'absent') {
    expect(await tl.locator('.labo-tl-band').count()).toBeGreaterThan(0)
  }
})
