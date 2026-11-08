// Captures réelles du cockpit : l’application construite est parcourue dans un
// état absent déterministe ; aucun moteur ne remplace une mesure manquante.
import { chromium } from '@playwright/test'
import { mkdirSync, rmSync } from 'node:fs'
import { join, resolve } from 'node:path'

const [baseURL, outputDir] = process.argv.slice(2)
if (!baseURL || !outputDir) {
  console.error('usage: node web/frontend/slice-cockpit-capture.mjs <baseURL> <outputDir>')
  process.exit(2)
}
const out = resolve(outputDir)
mkdirSync(out, { recursive: true })

const VIEWS = [
  { slug: 'dashboard', label: 'Tableau de bord', selector: '#v-dashboard', expected: 'Ce site n’a pas encore été mesuré' },
  { slug: 'audit', label: 'Auditer', selector: '#v-audit', expected: 'Audit unique : 10 s au repos' },
  { slug: 'bench', label: "Banc d'essai", selector: '#v-bench', expected: 'Une page : configurer, observer, lire.' },
  { slug: 'evidence', label: 'Preuves & Rapports', selector: '#v-evidence', expected: 'Preuves & Rapports' },
]
const VIEWPORTS = [
  { width: 890, height: 768 },
  { width: 1280, height: 800 },
  { width: 1920, height: 1080 },
]

async function mockApis(page) {
  await page.route('**/api/**', route => {
    const path = new URL(route.request().url()).pathname
    if (path.endsWith('/api/audit/status') || path.endsWith('/api/audit/live')) {
      return route.fulfill({ json: { running: false, report: null } })
    }
    if (path.endsWith('/api/audit/report')) {
      return route.fulfill({ status: 404, body: 'not found' })
    }
    if (path.endsWith('/api/audit/list')) return route.fulfill({ json: { audits: [] } })
    if (path.endsWith('/api/stream')) {
      return route.fulfill({ status: 200, contentType: 'text/event-stream', body: '' })
    }
    return route.fulfill({ json: {} })
  })
}

const browser = await chromium.launch({
  headless: true,
  executablePath: process.env.PWCHROMIUM || undefined,
})
const failures = []
// Seules les douze cibles sont effacées ; aucun autre fichier du dossier n’est touché.
for (const view of VIEWS) {
  for (const viewport of VIEWPORTS) {
    rmSync(join(out, `cockpit-${view.slug}-${viewport.width}.png`), { force: true })
  }
}
for (const viewport of VIEWPORTS) {
  const page = await browser.newPage({ viewport, colorScheme: 'dark' })
  page.on('console', message => {
    if (message.type() === 'error') console.error(`console: ${message.text()}`)
  })
  page.on('pageerror', error => console.error(`pageerror: ${error.message}`))
  await mockApis(page)
  await page.goto(baseURL, { waitUntil: 'load', timeout: 60000 })
  for (const view of VIEWS) {
    await page.getByRole('button', { name: view.label }).click()
    const panel = page.locator(view.selector)
    await panel.waitFor({ state: 'visible', timeout: 10000 })
    // La vue paresseuse n’est pas prête tant que son texte attendu n’existe pas.
    const ready = await panel.waitForFunction(
      (node, expected) => node.textContent?.includes(expected) ?? false,
      view.expected,
      { timeout: 30000 },
    ).then(() => true, () => false)
    if (!ready) {
      failures.push(`${view.slug}-${viewport.width}: texte attendu absent (${view.expected})`)
      continue
    }
    const text = await panel.textContent()
    if (!text?.includes(view.expected)) {
      failures.push(`${view.slug}-${viewport.width}: texte attendu absent (${view.expected})`)
      continue
    }
    const file = join(out, `cockpit-${view.slug}-${viewport.width}.png`)
    await page.screenshot({ path: file })
    console.log(`captured ${file}`)
  }
  await page.close()
}
await browser.close()
if (failures.length) {
  console.error(failures.join('\n'))
  process.exit(1)
}
