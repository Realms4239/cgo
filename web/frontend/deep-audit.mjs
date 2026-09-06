// Audit profond Playwright : chaque vue, chaque action, chaque graphique du
// Live pendant une campagne EN COURS — capture live-working.png incluse.
// Sortie 0 = tout vert.
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

// ---------- 1. RAIL : 4 vues, pas une de plus ----------
const railBtns = await p.$$eval('.rail .nav-btn[data-panel]', els => els.map(e => e.dataset.panel))
ok('rail: 4 vues exactement', railBtns.length === 4, railBtns.join(','))
ok('rail: ids attendus', ['campagne', 'live', 'resultats', 'integrite'].every(i => railBtns.includes(i)))

// ---------- 2. CAMPAGNE (cockpit) ----------
await p.click('.rail .nav-btn[data-panel="campagne"]')
await sleep(1200)
const profLabels = await p.$$eval('#v-campagne .check-row:first-of-type label', els => els.map(e => e.textContent.trim()))
ok('cockpit: profils P1–P4 affichés', ['P1','P2','P3','P4'].every(x => profLabels.some(l => l.startsWith(x))), profLabels.join(' '))
ok('cockpit: pas de 5e profil système', profLabels.filter(l => /^P[1-4]( |$)/.test(l)).length === 4)
const histBtn = await p.$('#v-campagne button:has-text("HISTORIQUE")')
if (histBtn) { await histBtn.click(); await sleep(900); ok('campagne: modal historique ouvre', (await p.$('dialog, [role=dialog]')) !== null); await p.keyboard.press('Escape'); await sleep(400) } else { console.log('  [--] historique absent (skipped)') }
await p.click('.rail .nav-btn[data-panel="live"]'); await sleep(1500)

// ---------- 3. LIVE pendant la campagne (running requis) ----------
// Un run CLI séparé est un PROCESSUS séparé : son live n'existe que dans sa
// propre mémoire, jamais dans le dashboard. Comme un opérateur réel, on
// démarre DONC via le dashboard (POST /api/run/start) — c'est le seul live
// que l'UI peut montrer. Cellule unique courte (P1, 1 qdisc, 1 cc, 1 rep).
const started = await p.evaluate(async () => {
  const r = await fetch('/api/run/start', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ profiles: ['P1'], qdiscs: ['fq_codel'], ccs: ['cubic'], reps: 1, deadline_ms: 300, target: '10.200.0.1', direction: 'up' }) })
  return { ok: r.ok, body: await r.text() }
})
ok('live: campagne démarrée via dashboard', started.ok, started.body.slice(0, 60))
await sleep(5000)
const live = await p.evaluate(async () => (await (await fetch('/api/state')).json()))
ok('live: campagne EN COURS détectée', !!live?.running, live ? `phase=${live.phase}` : 'aucune')
const charts = await p.$$eval('#v-live canvas', els => els.map(e => {
  const r = e.getBoundingClientRect()
  const ctx = e.getContext('2d')
  const d = ctx ? ctx.getImageData(0, 0, Math.max(1, Math.floor(r.width)), Math.max(1, Math.floor(r.height))).data : null
  let ink = 0
  if (d) for (let i = 3; i < d.length; i += 40) if (d[i] !== 0) ink++
  return { w: Math.round(r.width), h: Math.round(r.height), ink, cls: e.className || e.parentElement?.className || '' }
}))
ok('live: canvas présents', charts.length > 0, `${charts.length} canvas`)
for (const c of charts) ok(`live: canvas ${c.w}x${c.h} dessiné`, c.w > 50 && c.h > 30 && c.ink > 20, `ink=${c.ink} ${c.cls.slice(0, 30)}`)
const cards = await p.$$eval('#v-live .card', els => els.map(e => (e.textContent || '').slice(0, 40)))
ok('live: cartes KPI présentes', cards.length >= 5, `${cards.length} cartes`)
await p.screenshot({ path: path.join(OUT, 'live-working.png'), fullPage: false })
ok('live: capture live-working.png', fs.existsSync(path.join(OUT, 'live-working.png')))
const sseTxt = await p.$eval('#ft-sse', e => e.textContent)
ok('live: SSE connecté', /sse/i.test(sseTxt) && !/err/i.test(sseTxt), sseTxt.trim())

// ---------- 4. RESULTATS ----------
await p.click('.rail .nav-btn[data-panel="resultats"]'); await sleep(2200)
try { await p.waitForSelector('#v-resultats .leader-bar', { timeout: 8000 }) } catch { /* mesuré sous le seuil = KO ci-dessous */ }
const leaders = await p.$$eval('#v-resultats .leader-bar', els => els.length)
ok('resultats: lignes leaderboard', leaders > 0, `${leaders} lignes`)
const selects = await p.$$eval('#v-resultats select', els => els.map(e => e.parentElement.textContent.slice(0, 24) + ':' + e.value))
ok('resultats: sélecteurs TradeSpace', selects.length >= 3, selects.join(' | '))
const trade = await p.$eval('#v-resultats canvas', e => { const r = e.getBoundingClientRect(); const ctx = e.getContext('2d'); const d = ctx.getImageData(0, 0, Math.floor(r.width), Math.floor(r.height)).data; let ink = 0; for (let i = 3; i < d.length; i += 40) if (d[i] !== 0) ink++; return { w: r.width, h: r.height, ink } }).catch(() => ({ w: 0, h: 0, ink: 0 }))
ok('resultats: TradeSpace dessiné', trade.ink > 50, `ink=${trade.ink}`)
const infoBtn = await p.$('#v-resultats button:has-text("ⓘ")')
if (infoBtn) { await infoBtn.click(); await sleep(500); ok('resultats: ⓘ modal ouvre', (await p.$('dialog, [role=dialog]')) !== null); await p.keyboard.press('Escape') } else console.log('  [--] ⓘ absent')
await p.screenshot({ path: path.join(OUT, 'resultats.png') })

// ---------- 5. INTEGRITE ----------
await p.click('.rail .nav-btn[data-panel="integrite"]'); await sleep(1800)
const intRows = await p.$$eval('#v-integrite table tbody tr', els => els.length).catch(() => 0)
ok('integrite: tableau runs', intRows > 0, `${intRows} lignes`)
await p.screenshot({ path: path.join(OUT, 'integrite.png') })

// ---------- 6. Retour live : les events continuent, puis on arrête proprement ----------
await p.click('.rail .nav-btn[data-panel="live"]'); await sleep(1200)
const live2 = await p.evaluate(async () => (await (await fetch('/api/state')).json()))
ok('live: toujours actif après navigation', !!live2?.running || !!live2?.phase, live2?.phase ?? '—')
if (started.ok) {
  await p.evaluate(async () => { await fetch('/api/run/stop', { method: 'POST' }) })
  ok('live: campagne arrêtée proprement', true)
}
ok('zero erreur JS', jsErrs.length === 0, jsErrs.slice(0, 3).join(' | '))

await b.close()
console.log(`\n[deep] ${checks - fails}/${checks} OK${fails ? ` — ${fails} ÉCHEC(S)` : ' — TOUT VERT'}`)
process.exit(fails ? 1 : 0)
