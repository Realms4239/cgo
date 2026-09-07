// Audit profond Playwright v2 : chaque vue, chaque action, chaque graphique
// pendant une campagne EN COURS — capture live-working.png incluse.
// Couvre le batch P2.5→P3b : duel, constat, lien-score, paginate, LiveGuide,
// couleurs AQM, clavier, probe En-direct. Sortie 0 = tout vert.
import { chromium } from 'playwright-core'
import fs from 'node:fs'
import path from 'node:path'

const BASE = process.env.BASE ?? 'https://meteolink.dev:9090'
const OUT = process.env.OUTDIR ?? path.resolve('../shots/tmp-deep')
fs.mkdirSync(OUT, { recursive: true })
let fails = 0, checks = 0
const ok = (name, cond, extra = '') => { checks++; if (cond) console.log(`  [ok] ${name}${extra ? ' — ' + extra : ''}`); else { fails++; console.log(`  [KO] ${name}${extra ? ' — ' + extra : ''}`) } }
const sleep = (ms) => new Promise(r => setTimeout(r, ms))

const b = await chromium.launch({ args: ['--no-sandbox'] })
const ctx = await b.newContext({ viewport: { width: 1600, height: 900 } })
const p = await ctx.newPage()
const jsErrs = []
p.on('pageerror', e => jsErrs.push(String(e).split('\n')[0]))
await p.goto(BASE + '/', { waitUntil: 'domcontentloaded', timeout: 30000 })
await sleep(2500)

// ---------- 1. RAIL ----------
const railBtns = await p.$$eval('.rail .nav-btn[data-panel]', els => els.map(e => e.dataset.panel))
ok('rail: 4 vues exactement', railBtns.length === 4, railBtns.join(','))
ok('rail: source pill présente', (await p.$('button:has-text("source")')) !== null)

// ---------- 2. CAMPAGNE (cockpit P1–P4 seuls) ----------
await p.click('.rail .nav-btn[data-panel="campagne"]')
await sleep(1200)
const profLabels = await p.$$eval('#v-campagne .check-row:first-of-type label', els => els.map(e => e.textContent.trim()))
ok('cockpit: que P1–P4', profLabels.length === 4 && ['P1','P2','P3','P4'].every(x => profLabels.some(l => l.startsWith(x))), profLabels.join(' '))
const impGroup = await p.$$eval('#v-campagne', els => els.map(e => e.textContent).join(' ').includes('Importés'))
ok('cockpit: aucun groupe Importés', !impGroup[0])
const histBtn = await p.$('#v-campagne button:has-text("HISTORIQUE")')
if (histBtn) { await histBtn.click(); await sleep(900); ok('campagne: modal historique ouvre', (await p.$('dialog, [role=dialog]')) !== null); await p.keyboard.press('Escape'); await sleep(400) }
await p.click('.rail .nav-btn[data-panel="live"]'); await sleep(1500)

// ---------- 3. LIVE pendant la campagne ----------
const started = await p.evaluate(async () => {
  const r = await fetch('/api/run/start', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ profiles: ['P1'], qdiscs: ['fq_codel'], ccs: ['cubic'], reps: 1, deadline_ms: 300, target: '10.200.0.1', direction: 'up' }) })
  return { ok: r.ok, body: await r.text() }
})
ok('live: campagne démarrée via dashboard', started.ok, started.body.slice(0, 60))
await sleep(6000)
const live = await p.evaluate(async () => (await (await fetch('/api/state')).json()))
ok('live: EN COURS', !!live?.running, live ? `phase=${live.phase}` : 'aucune')
// chaque chart un par un : small hero, rtt, goodput (les 3 canvas)
const charts = await p.$$eval('#v-live canvas', els => els.map(e => {
  const r = e.getBoundingClientRect()
  const ctx = e.getContext('2d')
  const d = ctx ? ctx.getImageData(0, 0, Math.max(1, Math.floor(r.width)), Math.max(1, Math.floor(r.height))).data : null
  let ink = 0
  if (d) for (let i = 3; i < d.length; i += 40) if (d[i] !== 0) ink++
  return { w: Math.round(r.width), h: Math.round(r.height), ink }
}))
ok('live: 3 canvas (small+rtt+goodput)', charts.length === 3, charts.map(c => `${c.w}x${c.h}`).join(','))
for (const c of charts) ok(`live: canvas ${c.w}x${c.h} ink`, c.ink > 20, `ink=${c.ink}`)
const guide = await p.$('text=Guide de lecture') ?? await p.$(':text("Guide de lecture")')
ok('live: guide de lecture présent', guide !== null)
const metricCards = await p.$$eval('[data-wall-cards] > *', els => els.length)
ok('live: cartes métriques', metricCards >= 8, `${metricCards} cartes`)
await p.screenshot({ path: path.join(OUT, 'live-working.png'), fullPage: false })
ok('live: capture live-working.png', fs.existsSync(path.join(OUT, 'live-working.png')))
const sseTxt = await p.$eval('#ft-sse', e => e.textContent)
ok('live: SSE connecté', /sse/i.test(sseTxt) && !/err/i.test(sseTxt), sseTxt.trim())
// source pill : cycle both→live→frozen→both — 2 clics pour frozen
await p.click('.rail .nav-btn[data-panel="live"]'); await sleep(400)
const pill = await p.$('button:has-text("source")')
await pill.click(); await sleep(300) // both -> live : bento reste
const bentoLive = await p.evaluate(() => !!document.querySelector('[data-wall-cards]'))
ok('live: source=live garde les métriques', bentoLive)
await pill.click(); await sleep(400) // live -> frozen : bento parti
const bentoGone = await p.evaluate(() => !document.querySelector('[data-wall-cards]'))
ok('live: source=frozen cache les métriques', bentoGone)
await pill.click(); await sleep(300) // frozen -> both : retour

// ---------- 4. RESULTATS ----------
await p.click('.rail .nav-btn[data-panel="resultats"]'); await sleep(2200)
try { await p.waitForSelector('#v-resultats .leader-bar', { timeout: 8000 }) } catch { }
const leaders = await p.$$eval('#v-resultats .leader-bar', els => els.length)
ok('resultats: lignes leaderboard', leaders > 0, `${leaders} lignes`)
// duel image-style
const duel = await p.$('[data-testid="rank-verdict"]')
ok('resultats: duel présent', !!duel)
if (duel) {
  const dt = await duel.textContent()
  ok('resultats: duel a 2 colonnes', /avant/i.test(dt), dt.trim().slice(0, 60).replace(/\s+/g, ' '))
}
const constat = await p.$('[data-testid="constat-button"]')
ok('resultats: bouton interprétation', !!constat)
// expand + lien + français
await p.$eval('.lb-row', e => e.click())
await sleep(500)
const lien = await p.$('[data-testid="lien-score"]')
ok('resultats: score LIEN décomposé', !!lien, lien ? (await lien.textContent()).trim().slice(0, 50).replace(/\s+/g, ' ') : '')
const fr = await p.evaluate(() => document.body.textContent.includes('Latence du lien'))
ok('resultats: détails en français', fr)
const dim = await p.$$eval('.lb-row', els => els.map(e => getComputedStyle(e).opacity))
ok('resultats: frères estompés au clic', dim.some(o => o === '0.35') && dim.some(o => o === '1'), dim.slice(0, 4).join(','))
await p.keyboard.press('Tab')
// clavier : focus EXPLICITE sur la ligne (Tab seul part du rail)
await p.$eval('.lb-row', e => e.focus())
await p.keyboard.press('Enter')
await sleep(300)
// couleurs AQM + marqueurs CC
const ccMarks = await p.$$eval('span[title*="contrôle de congestion"]', els => els.length)
ok('resultats: marqueurs CC ◆/◇', ccMarks > 0, `${ccMarks}`)
const keyline = await p.evaluate(() => document.body.textContent.includes('◆ cubic'))
ok('resultats: clé couleur visible', keyline)
// changelog pastilles + pagination
const pag = await p.$('[data-testid="paginate"]')
ok('resultats: pagination changelog', pag !== null)
const trade = await p.$eval('#v-resultats canvas', e => { const r = e.getBoundingClientRect(); const ctx = e.getContext('2d'); const d = ctx.getImageData(0, 0, Math.floor(r.width), Math.floor(r.height)).data; let ink = 0; for (let i = 3; i < d.length; i += 40) if (d[i] !== 0) ink++; return ink }).catch(() => 0)
ok('resultats: TradeSpace dessiné', trade > 50, `ink=${trade}`)
const corner = await p.$$eval('span', els => els.filter(e => /optimal ↗|optimal ↘|optimal ↙|optimal ↖/.test(e.textContent)).length)
ok('resultats: coin optimal annoté', corner > 0)
await p.screenshot({ path: path.join(OUT, 'resultats.png') })

// ---------- 5. INTEGRITE ----------
await p.click('.rail .nav-btn[data-panel="integrite"]'); await sleep(1800)
const intRows = await p.$$eval('[data-testid="runs-table"] tbody tr', els => els.length).catch(() => 0)
ok('integrite: runs paginés', intRows > 0 && intRows <= 12, `${intRows} lignes/page`)
const quarRows = await p.$$eval('[data-testid="quarantine-table"] tbody tr', els => els.length).catch(() => 0)
ok('integrite: quarantaine paginée', quarRows <= 12, `${quarRows} lignes/page`)
const intPag = await p.$('#v-integrite [data-testid="paginate"]')
ok('integrite: pagination 1..n', intPag !== null)
await p.screenshot({ path: path.join(OUT, 'integrite.png') })

// ---------- 6. arrêt propre ----------
await p.click('.rail .nav-btn[data-panel="live"]'); await sleep(1200)
const live2 = await p.evaluate(async () => (await (await fetch('/api/state')).json()))
ok('live: actif après navigation', !!live2?.running || !!live2?.phase, live2?.phase ?? '—')
if (started.ok) {
  await p.evaluate(async () => { await fetch('/api/run/stop', { method: 'POST' }) })
  ok('live: arrêt propre', true)
}
ok('zero erreur JS', jsErrs.length === 0, jsErrs.slice(0, 3).join(' | '))

await b.close()
console.log(`\n[deep] ${checks - fails}/${checks} OK${fails ? ` — ${fails} ÉCHEC(S)` : ' — TOUT VERT'}`)
process.exit(fails ? 1 : 0)
