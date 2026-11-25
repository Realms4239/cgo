import { test } from '@playwright/test'

// Audit visuel — captures pleine page des deux destinations vivantes,
// plus vérification du tunnel SSE. Le mur à quatre panneaux n'existe plus :
// on capture Synthèse puis les quatre chapitres du Labo.

test('idle — both live modes @1920', async ({ page }) => {
  test.setTimeout(90000)
  await page.setViewportSize({ width: 1920, height: 1080 })
  await page.goto('/')
  await page.waitForTimeout(1800)
  await page.screenshot({ path: 'shots/audit-idle-cockpit-1920.png', fullPage: true })

  await page.locator('[data-mode="labo"]').click()
  for (const [i, id] of [[0, 'configurer'], [1, 'observer'], [2, 'comparer'], [3, 'preuves']] as const) {
    await page.locator('.pr-step').nth(i).click()
    await page.waitForTimeout(1500)
    await page.screenshot({ path: `shots/audit-idle-${id}-1920.png`, fullPage: true })
  }
})

test('live run — instrument with streaming data', async ({ page }) => {
  test.setTimeout(150000)
  await page.setViewportSize({ width: 1920, height: 1080 })
  await page.goto('/')
  await page.locator('[data-mode="labo"]').click()
  await page.locator('.pr-step').nth(1).click()
  // vérité SSE côté navigateur : frameCount doit croître
  const grew = await page.waitForFunction(() => {
    const f = (window as any).__CGO_SSE?.frameCount ?? 0
    if (!f) (window as any).__first = f
    return f > ((window as any).__first ?? 0) + 20
  }, null, { timeout: 8000 }).then(() => true).catch(() => false)
  await page.screenshot({ path: 'shots/audit-idle-instrument-before.png', fullPage: true })
  // lance un run réel par l'API pour alimenter les anneaux du chapitre Observer
  await page.evaluate(async () => {
    await fetch('/api/run/start', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ profiles: ['P1'], reps: 1 }) })
  })
  await page.waitForFunction(() => (window as any).__CGO_LIVE?.phase === 'charge', null, { timeout: 45000 }).catch(() => {})
  await page.waitForTimeout(6000)
  await page.screenshot({ path: 'shots/audit-live-instrument-1920.png', fullPage: true })
  const seam = await page.evaluate(() => ({
    frames: (window as any).__CGO_SSE?.frameCount,
    phase: (window as any).__CGO_LIVE?.phase,
    ring: (window as any).__CGO_LIVE?.rtt95?.length,
    max: (window as any).__CGO_LIVE?.max,
  }))
  console.log('LIVE-SEAM:', JSON.stringify(seam), 'sseGrew:', grew)
})

// BLOQUÉ — le relais SSE externe piloté par AUDIT_TUNNEL n'a pas d'équivalent
// dans le binaire local (aucun hôte codé en dur, aucun tunnel monté ici).
// Vérifié : `__CGO_SSE` n'est alimenté que par /api/stream du serveur local.
// Assertion conservée, non exécutée. Voir task-1-report.md.
test.skip('tunnel — browser EventSource through external relay', async ({ page }) => {
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
  await page.locator('[data-mode="labo"]').click()
  await page.locator('.pr-step').nth(1).click()
  await page.waitForTimeout(1500)
  await page.screenshot({ path: 'shots/audit-tunnel-instrument-1920.png', fullPage: true })
})
