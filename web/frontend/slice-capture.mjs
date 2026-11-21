// Captures réelles de la dernière tranche : le service Go embedé répond,
// aucune route n'est interceptée, aucune mesure n'est fabriquée.
import { chromium } from '@playwright/test'
import { mkdirSync, rmSync } from 'node:fs'
import { join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const baseURL = process.argv[2] || 'http://127.0.0.1:19091'
const out = resolve(process.argv[3] || fileURLToPath(new URL('./shots-redesign', import.meta.url)))
mkdirSync(out, { recursive: true })

const VIEWPORTS = [
  { width: 890, height: 512 },
  { width: 1280, height: 768 },
  { width: 1920, height: 1080 },
]
const STATES = [
  { slug: 'cockpit', mode: 'cockpit', action: null },
  { slug: 'cockpit-preuve', mode: 'cockpit', action: 'preuve' },
  { slug: 'labo', mode: 'labo', action: null },
  { slug: 'labo-profils', mode: 'labo', action: 'profils' },
]

const browser = await chromium.launch({
  headless: true,
  executablePath: process.env.PWCHROMIUM || undefined,
})
const failures = []
// Seules les douze cibles sont effacées ; aucun autre fichier n'est touché.
for (const state of STATES) {
  for (const viewport of VIEWPORTS) {
    rmSync(join(out, `${state.slug}-${viewport.width}.png`), { force: true })
  }
}

for (const viewport of VIEWPORTS) {
  const context = await browser.newContext({ viewport, colorScheme: 'dark' })
  const page = await context.newPage()
  page.on('pageerror', error => failures.push(`pageerror: ${error.message}`))
  for (const state of STATES) {
    await page.goto(baseURL, { waitUntil: 'load', timeout: 60000 })
    await page.locator(`[data-mode="${state.mode}"]`).click()
    if (state.action === 'preuve') {
      await page.locator('.inline-disclosure-trigger').first().click()
      await page.getByRole('button', { name: 'Ouvrir la preuve' }).click()
      await page.locator('.ui-drawer-panel').waitFor({ state: 'visible' })
    }
    if (state.action === 'profils') {
      await page.getByRole('button', { name: /Profils de campagne/ }).first().click()
      await page.locator('.profile-tile').first().waitFor({ state: 'visible' })
    }
    await page.waitForTimeout(200)
    const file = join(out, `${state.slug}-${viewport.width}.png`)
    await page.screenshot({ path: file })
    console.log(`captured ${file}`)
  }
  await context.close()
}
await browser.close()
if (failures.length) {
  console.error(failures.join('\n'))
  process.exit(1)
}
