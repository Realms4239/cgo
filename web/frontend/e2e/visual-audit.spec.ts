import { test } from '@playwright/test'

// Visual audit — full-page captures. Idle + live-run + tunnel SSE check.
const BASE = process.env.AUDIT_BASE || 'http://192.168.174.128:9090'

test('idle — all 4 panels @1920', async ({ page }) => {
  test.setTimeout(90000)
  await page.setViewportSize({ width: 1920, height: 1080 })
  for (const p of ['campagne', 'live', 'resultats', 'integrite']) {
    await page.goto(BASE + '/')
    await page.locator(`[data-panel="${p}"]`).click()
    await page.waitForTimeout(1800)
    await page.screenshot({ path: `shots/audit-idle-${p}-1920.png`, fullPage: true })
  }
})

test('live run — wall with streaming data', async ({ page }) => {
  test.setTimeout(150000)
  await page.setViewportSize({ width: 1920, height: 1080 })
  await page.goto(BASE + '/')
  await page.locator('[data-panel="live"]').click()
  // browser-side SSE truth: frameCount must grow
  const grew = await page.waitForFunction(() => {
    const f = (window as any).__CGO_SSE?.frameCount ?? 0
    if (!f) (window as any).__first = f
    return f > ((window as any).__first ?? 0) + 20
  }, { timeout: 8000 }).then(() => true).catch(() => false)
  await page.screenshot({ path: 'shots/audit-idle-wall-before.png', fullPage: true })
  // start a quick P1 run from the API so charts get real data
  await page.evaluate(async () => {
    await fetch('/api/run/start', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ profiles: ['P1'], reps: 1 }) })
  })
  // wait for charge phase or 30s
  await page.waitForFunction(() => (window as any).__CGO_LIVE?.phase === 'charge', { timeout: 45000 }).catch(() => {})
  await page.waitForTimeout(6000)
  await page.screenshot({ path: 'shots/audit-live-wall-1920.png', fullPage: true })
  const seam = await page.evaluate(() => ({
    frames: (window as any).__CGO_SSE?.frameCount,
    phase: (window as any).__CGO_LIVE?.phase,
    ring: (window as any).__CGO_LIVE?.rtt95?.length,
    max: (window as any).__CGO_LIVE?.max,
  }))
  console.log('LIVE-SEAM:', JSON.stringify(seam), 'sseGrew:', grew)
})

test('tunnel — browser EventSource through lhr.life', async ({ page }) => {
  const tunnel = process.env.AUDIT_TUNNEL
  if (!tunnel) { test.skip(true, 'no tunnel'); return }
  await page.setViewportSize({ width: 1920, height: 1080 })
  await page.goto(tunnel + '/')
  await page.waitForTimeout(6000)
  const seam = await page.evaluate(() => ({
    frames: (window as any).__CGO_SSE?.frameCount ?? 0,
    badge: document.querySelector('.hd-right .mono')?.textContent ?? '',
  }))
  console.log('TUNNEL-SSE:', JSON.stringify(seam))
  await page.locator('[data-panel="live"]').click()
  await page.waitForTimeout(1500)
  await page.screenshot({ path: 'shots/audit-tunnel-wall-1920.png', fullPage: true })
})
