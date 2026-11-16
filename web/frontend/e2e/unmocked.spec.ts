// UNMOCKED — aucun page.route, aucun fetch réécrit. Le binaire cgo réel sert
// la SPA embarquée ET /api/* depuis book/data/runs (corpus gelé). Rendu réel,
// au trois paliers de largeur : 390 / 890 / 1920.
import { test, expect } from '@playwright/test'

test.describe.configure({ timeout: 120_000 })
import { spawn, type ChildProcess } from 'node:child_process'
import { existsSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const REPO = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..', '..', '..')
const PORT = 5174
const BASE = `http://127.0.0.1:${PORT}`

let proc: ChildProcess | null = null
let serverUp = false

async function waitForHealth(timeoutMs = 90_000) {
  const start = Date.now()
  while (Date.now() - start < timeoutMs) {
    try {
      const r = await fetch(`${BASE}/api/health`)
      if (r.ok) return true
    } catch { /* pas encore prêt */ }
    await new Promise(r => setTimeout(r, 500))
  }
  return false
}

test.beforeAll(async () => {
  test.setTimeout(120_000)
  const bin = process.env.CGO_E2E_BIN
  const cmd = bin ? bin : 'go'
  const args = bin ? ['--serve', '--mode', 'observe', '--addr', `127.0.0.1:${PORT}`, '--tls=false']
    : ['run', './cmd/cgo', '--serve', '--mode', 'observe', '--addr', `127.0.0.1:${PORT}`, '--tls=false']
  const dataRoot = existsSync(path.join(REPO, 'book', 'data', 'runs')) ? path.join(REPO, 'book', 'data') : path.join(REPO, 'data')
  proc = spawn(cmd, args, { cwd: REPO, env: { ...process.env, CGO_DATA_DIR: dataRoot }, stdio: ['ignore', 'pipe', 'pipe'] })
  proc.stdout?.on('data', d => process.stderr.write(`[cgo] ${d}`))
  proc.stderr?.on('data', d => process.stderr.write(`[cgo!] ${d}`))
  proc.on('error', () => { /* géré par le timeout */ })
  serverUp = await waitForHealth()
  if (!serverUp) throw new Error(`cgo serve n'a pas pris ${BASE}/api/health en 90 s`)
})

test.afterAll(async () => {
  proc?.kill('SIGTERM')
  await new Promise(r => setTimeout(r, 500))
  proc?.kill('SIGKILL')
})

test.use({ baseURL: BASE })

const VIEWPORTS = [
  { name: '390×844', width: 390, height: 844 },
  { name: '890×768', width: 890, height: 768 },
  { name: '1920×1080', width: 1920, height: 1080 },
]

for (const vp of VIEWPORTS) {
  test(`unmocked — ${vp.name} rend l'app réelle, aucun défilement horizontal`, async ({ page }) => {
    await page.setViewportSize({ width: vp.width, height: vp.height })
    const health = await page.request.get(`${BASE}/api/health`)
    expect(health.ok()).toBeTruthy()
    const body = await health.text()
    expect(body).toContain('"ok"')
    await page.goto('/', { waitUntil: 'domcontentloaded' })
    await expect(page.locator('main#main')).toBeVisible({ timeout: 15000 })
    const overflow = await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth + 1)
    expect(overflow).toBe(false)
  })
}

test('unmocked — corpus gelé vu par l’API réelle : 159 / 208 / 74 / 246', async ({ page }) => {
  await page.setViewportSize({ width: 1920, height: 1080 })
  await page.goto('/', { waitUntil: 'domcontentloaded' })
  // Le Cockpit porte le panneau Preuves & Rapports (intégré, pas une destination de plus)
  const panel = page.locator('[data-testid="cockpit-preuves"]')
  await expect(panel).toBeVisible({ timeout: 20000 })
  const summary = page.locator('.int-summary')
  await expect(summary).toContainText('campagnes archivées', { timeout: 15000 })
  await expect(summary).toContainText('mesures valides')
  await expect(summary).toContainText('159')
  await expect(summary).toContainText('208')
  await expect(summary).toContainText('74')
  await expect(summary).toContainText('246')
})

test('unmocked — aucune erreur console ou pageerror sur le chemin complet', async ({ page }) => {
  const errors: string[] = []
  page.on('pageerror', e => errors.push(`pageerror: ${e.message}`))
  page.on('console', m => { if (m.type() === 'error') errors.push(`console.error: ${m.text()}`) })
  await page.setViewportSize({ width: 1920, height: 1080 })
  await page.goto('/', { waitUntil: 'domcontentloaded' })
  await expect(page.locator('main#main')).toBeVisible({ timeout: 20000 })
  await page.getByRole('button', { name: /Labo/i }).first().click()
  await expect(page.getByText(/Configurer|Instruments|Campagne/i).first()).toBeVisible({ timeout: 20000 })
  await page.getByRole('button', { name: /Cockpit/i }).first().click()
  await expect(page.locator('[data-testid="cockpit-preuves"]')).toBeVisible({ timeout: 20000 })
  expect(errors).toEqual([])
})
