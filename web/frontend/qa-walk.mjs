// qa-walk.mjs — parcours réel des quatre destinations : erreurs console,
// longues tâches, clavier 1–4, tiroirs scientifiques, rapports d'audit.
// Usage: node qa-walk.mjs [baseURL]
import { chromium } from 'playwright-core'

const base = process.argv[2] || 'http://127.0.0.1:9091'
const browser = await chromium.launch({ args: ['--no-sandbox'], channel: process.env.PW_CHANNEL || 'msedge' })
  .catch(() => chromium.launch({ args: ['--no-sandbox'], channel: 'chrome' }))
const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 } })
await ctx.addInitScript(() => {
  window.__LT = []
  window.__MARK = []
  try {
    new PerformanceObserver(l => {
      for (const e of l.getEntries()) window.__LT.push({ t: Math.round(e.startTime), dur: Math.round(e.duration) })
    }).observe({ type: 'longtask', buffered: true })
  } catch {}
})
const page = await ctx.newPage()
const errors = []
const resourceErrors = []
page.on('console', m => {
  if (m.type() !== 'error') return
  const t = m.text()
  // un 404 réseau est une réponse attendue quand une ligne n'a pas de dossier de rapport
  if (/Failed to load resource/.test(t)) resourceErrors.push(t); else errors.push(t)
})
page.on('pageerror', e => errors.push(String(e)))

const steps = []
const ok = (name, pass, detail = '') => steps.push({ name, pass: !!pass, detail: String(detail).slice(0, 160) })
const attempt = async (name, fn) => {
  try { await fn() } catch (e) { ok(name, false, String(e).split(/[\r\n]/)[0]) }
}
// marque temporelle côté page : chaque longue tâche se rattache au geste qui
// l'a produite, sinon le chiffre reste un symptôme sans diagnostic.
const mark = async (label) => {
  await page.evaluate(l => { (window.__MARK ||= []).push({ label: l, t: performance.now() }) }, label).catch(() => {})
}

await page.goto(base, { waitUntil: 'domcontentloaded', timeout: 30000 })
// la marque doit venir après la navigation : __MARK est réinitialisé par le
// script d'init du nouveau document, une marque antérieure serait perdue.
await mark('ouverture')
await page.waitForTimeout(6000)

// 1 — tableau de bord : les deux mondes sont séparés et nommés
const realTag = await page.locator('.dash-tag.real').first().textContent().catch(() => null)
const simTag = await page.locator('.dash-tag.sim').first().textContent().catch(() => null)
ok('dashboard sépare réel et simulé', !!realTag && !!simTag, `${realTag} | ${simTag}`)
const drawers = page.locator('.sci-details-toggle')
const drawerCount = await drawers.count()
if (drawerCount > 0) {
  await attempt('tiroir (clic)', async () => { await drawers.first().click({ timeout: 5000 }) })
  await page.waitForTimeout(300)
  const open = await page.locator('.sci-details-body').first().isVisible().catch(() => false)
  ok('tiroir scientifique s’ouvre', open, `${drawerCount} tiroirs`)
  await attempt('tiroir (repli)', async () => { await drawers.first().click({ timeout: 5000 }) })
} else {
  ok('tiroir scientifique s’ouvre', false, 'aucun tiroir trouvé')
}

// 2 — clavier 1–4
await page.locator('body').click({ position: { x: 5, y: 300 } })
for (const [key, mode] of [['2', 'audit'], ['3', 'bench'], ['4', 'evidence'], ['1', 'dashboard']]) {
  await mark(`touche ${key} → ${mode}`)
  await page.keyboard.press(key)
  await page.waitForTimeout(500)
  const hidden = await page.locator(`#v-${mode}`).getAttribute('hidden')
  ok(`touche ${key} → ${mode}`, hidden === null)
}

// 2b — chrome : réglages et palette s'ouvrent et se referment
await attempt('réglages (ouverture)', async () => { await page.getByRole('button', { name: 'Réglages' }).click({ timeout: 5000 }) })
await page.waitForTimeout(350)
ok('les réglages s’ouvrent', await page.locator('[role="dialog"][aria-label="Réglages"]').count() > 0)
await page.keyboard.press('Escape')
await page.waitForTimeout(300)
ok('les réglages se ferment', await page.locator('[role="dialog"][aria-label="Réglages"]').count() === 0)
await attempt('palette (ouverture)', async () => { await page.keyboard.press('Control+k') })
await page.waitForTimeout(350)
ok('la palette s’ouvre', await page.locator('#palette').count() > 0)
await page.keyboard.press('Escape')
await page.waitForTimeout(300)
ok('la palette se ferme', await page.locator('#palette').count() === 0)

// 2c — aucun débordement horizontal aux largeurs de terrain
for (const [w, h] of [[1366, 768], [1024, 768]]) {
  await page.setViewportSize({ width: w, height: h })
  await page.waitForTimeout(350)
  const over = await page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth)
  ok(`pas de débordement à ${w} px`, over <= 1, `${over} px`)
}
await page.setViewportSize({ width: 1440, height: 900 })
await page.waitForTimeout(350)

// 3 — auditer : la déclaration est complète et l’avancé se replie
await page.keyboard.press('2')
await page.waitForTimeout(800)
const decisions = await page.locator('.audit-declare label').count()
ok('auditer expose les quatre décisions', decisions >= 4, `${decisions} étiquettes`)
const planText = await page.locator('.audit-declare').getByText(/au total/).first().textContent().catch(() => '')
ok('la durée annonce ses phases', /repos/.test(planText || ''), planText)
await attempt('avancé (clic)', async () => { await page.locator('button:has-text("CHAMPS AVANCÉS")').click({ timeout: 5000 }) })
await page.waitForTimeout(200)
ok('avancé se déplie', await page.locator('input[name="audit-provider"]').isVisible())
await attempt('avancé (repli)', async () => { await page.locator('button:has-text("MASQUER LES CHAMPS AVANCÉS")').click({ timeout: 5000 }) })

// 4 — banc : les trois sections, l’aperçu replié, l’affichage
await mark('banc (chunk différé)')
await page.keyboard.press('3')
await page.waitForTimeout(1200)
for (const id of ['bench-config', 'bench-observe', 'bench-results']) {
  ok(`section ${id}`, await page.locator(`#${id}`).count() > 0)
}
await mark('ancre résultats')
await attempt('ancre résultats', async () => { await page.locator('.bench-anchors button:has-text("RÉSULTATS")').click({ timeout: 5000 }) })
await page.waitForSelector('#bench-results .lb-card', { timeout: 20000 }).catch(() => {})
ok('résultats montés', await page.locator('#bench-results .lb-card').count() > 0)
// le nuage arrive après la première peinture : sans cette preuve, un montage
// différé silencieusement cassé passerait comme une victoire de performance.
await page.waitForSelector('#bench-results .ts-wrap canvas', { timeout: 8000 }).catch(() => {})
const canvasBox = await page.locator('#bench-results .ts-wrap canvas').first().boundingBox().catch(() => null)
ok('le TradeSpace est réellement peint', !!canvasBox && canvasBox.width > 100 && canvasBox.height > 100,
  canvasBox ? `${Math.round(canvasBox.width)}×${Math.round(canvasBox.height)} px` : 'aucun canvas')
const reset = page.locator('button:has-text("RÉINITIALISER")')
ok('réinitialiser apparaît après filtre', await reset.count() === 0 || await reset.isVisible())

// 5 — preuves : le premier clic ouvre un rapport, aucune impasse à jouer
await mark('preuves')
await page.keyboard.press('4')
await page.waitForTimeout(5000)
await attempt('ancre rapports', async () => { await page.locator('.evidence-anchors button:has-text("RAPPORTS")').click({ timeout: 5000 }) })
await page.waitForTimeout(600)
const rows = await page.locator('#evidence-reports .audit-reports-table tbody tr').count()
const buttons = page.locator('#evidence-reports button:has-text("OUVRIR")')
const n = await buttons.count()
const marked = await page.locator('#evidence-reports .data-missing').count()
ok('les lignes sans archive sont dites d’avance', rows === 0 || n + marked === rows, `${n} ouvrables, ${marked} non archivées / ${rows} lignes`)

// même contrat pour la table des runs : un run sans ligne archivée ne propose
// pas un replay qui ne peut produire que du vide
const runRows = await page.locator('[data-testid="runs-table"] tbody tr').count()
const replayBtns = await page.locator('[data-testid="runs-table"] [data-testid="run-rejouer"]').count()
const notReplayable = await page.locator('[data-testid="runs-table"] .data-missing').count()
ok('un run vide ne propose pas de replay', runRows === 0 || replayBtns + notReplayable === runRows,
  `${replayBtns} rejouables, ${notReplayable} non rejouables / ${runRows} lignes`)
let opened = false
if (n > 0) {
  resourceErrors.length = 0 // le débit ci-dessous ne doit plus rien aux 404
  await mark('ouverture rapport')
  await attempt('ouverture rapport', async () => { await buttons.first().click({ timeout: 5000 }) })
  await page.waitForTimeout(1500)
  opened = (await page.locator('#evidence-reports [data-testid="audit-report-card"]').count()) > 0
  ok('le premier OUVRIR ouvre un rapport', opened, `${n} boutons ouvrables`)
  ok('ouverture sans 404', resourceErrors.length === 0, resourceErrors[0] || '')
  if (opened) {
    ok('citation présente', await page.locator('#evidence-reports button:has-text("CITER CE RAPPORT")').count() > 0)
    await attempt('retour à la liste', async () => { await page.locator('button:has-text("RETOUR À LA LISTE")').click({ timeout: 5000 }) })
  }
}

// chaque longue tâche se rattache au geste qui l'a produite ; le budget, lui,
// est une étape comme les autres : au-delà de 400 ms le banc est jugé lent.
const longTasks = await page.evaluate(() => window.__LT || [])
const marks = await page.evaluate(() => window.__MARK || [])
const afterMark = (t) => { let label = '—'; for (const m of marks) if (m.t <= t) label = m.label; return label }
const byStep = {}
for (const e of longTasks) {
  const k = afterMark(e.t)
  byStep[k] = byStep[k] || { count: 0, max: 0 }
  byStep[k].count++
  byStep[k].max = Math.max(byStep[k].max, e.dur)
}
const worst = longTasks.reduce((m, e) => Math.max(m, e.dur), 0)
const worstStep = longTasks.length ? afterMark(longTasks.reduce((a, b) => (b.dur > a.dur ? b : a)).t) : '—'
ok('aucune longue tâche > 400 ms', worst <= 400, `${worst} ms après « ${worstStep} »`)

// 6 — téléphone : le bandeau se comprime, les cibles tactiles tiennent 44 px.
// Après la collecte : le rechargement remet __LT à zéro, les chiffres mesurés
// restent ceux du parcours desktop.
await page.setViewportSize({ width: 390, height: 844 })
await page.reload({ waitUntil: 'domcontentloaded' })
await page.waitForTimeout(2500)
const corpusShown = await page.locator('.ribbon-corpus').isVisible().catch(() => false)
const sealShown = await page.locator('.ribbon-seal').isVisible().catch(() => false)
ok('390 px : le corpus et le sceau restent dans Preuves', !corpusShown && !sealShown,
  `corpus ${corpusShown ? 'visible' : 'masqué'}, sceau ${sealShown ? 'visible' : 'masqué'}`)
const btnBox = await page.locator('.status-ribbon .btn').first().boundingBox().catch(() => null)
ok('390 px : cible tactile au moins 44 px', !!btnBox && btnBox.height >= 44, btnBox ? `${Math.round(btnBox.height)} px` : 'aucune cible')

console.log(JSON.stringify({
  steps,
  failures: steps.filter(s => !s.pass).length,
  consoleErrors: errors, resourceErrors,
  longTasks: { count: longTasks.length, over200: longTasks.filter(e => e.dur > 200).length, max: worst, byStep },
}, null, 2))
await browser.close()
process.exit(steps.some(s => !s.pass) || errors.length ? 1 : 0)
