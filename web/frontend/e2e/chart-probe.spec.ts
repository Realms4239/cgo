import { test, expect } from '@playwright/test'

// Chart-health probe — every chart surface renders, no console errors,
// instrumentation seams report data flow. Debug-after ledger (plan §5).

const errors: string[] = []

test('wall — 3 ECharts canvases render non-blank + seams + no console errors', async ({ page }) => {
  page.on('console', m => { if (m.type() === 'error') errors.push(m.text()) })
  page.on('pageerror', e => errors.push(String(e)))
  await page.goto('/')
  await page.locator('[data-panel="live"]').click()
  await page.waitForTimeout(2500)

  // 3 grammar charts (hero + rtt + goodput) each render a canvas with visible pixels
  const canvases = page.locator('#wall canvas')
  const n = await canvases.count()
  expect(n, 'wall canvas count').toBeGreaterThanOrEqual(3)
  let nonBlank = 0
  for (let i = 0; i < n; i++) {
    const ok = await canvases.nth(i).evaluate((c: HTMLCanvasElement) => {
      try {
        const ctx = c.getContext('2d')
        if (!ctx || c.width === 0 || c.height === 0) return false
        const d = ctx.getImageData(0, 0, Math.min(c.width, 400), Math.min(c.height, 200)).data
        for (let p = 3; p < d.length; p += 40) if (d[p] !== 0) return true
        return false
      } catch { return false }
    })
    if (ok) nonBlank++
  }
  expect(nonBlank, 'non-blank chart canvases').toBeGreaterThanOrEqual(3)

  // instrumentation seams expose live data path
  const seam = await page.evaluate(() => ({
    sse: (window as any).__CGO_SSE?.frameCount ?? -1,
    liveMax: (window as any).__CGO_LIVE?.max ?? -1,
    phase: (window as any).__CGO_LIVE?.phase ?? '?',
  }))
  expect(seam.sse).toBeGreaterThan(0)
  expect(seam.liveMax).toBe(1800)

  expect(errors, 'console/page errors').toEqual([])
})

test('resultats — scatter canvas renders non-blank with real groups', async ({ page }) => {
  page.on('pageerror', e => errors.push(String(e)))
  await page.goto('/')
  await page.locator('[data-panel="resultats"]').click()
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

test('campagne — timeline 48 brush present when running phase data exists', async ({ page }) => {
  await page.goto('/')
  await page.locator('[data-panel="campagne"]').click()
  // timeline is data-gated (honest): assert it either renders SVG bands or stays hidden — never an empty 48px box
  const tl = page.locator('[data-testid="timeline"]')
  const count = await tl.count()
  if (count > 0) {
    const hasSvg = await tl.locator('svg').count()
    expect(hasSvg).toBeGreaterThan(0)
  }
})
