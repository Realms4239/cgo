// perf-probe.mjs — mesure honnête du chargement initial et de l'ouverture du banc.
// Usage: node perf-probe.mjs [baseURL]
import { chromium } from 'playwright-core'

const base = process.argv[2] || 'http://127.0.0.1:9091'
const browser = await chromium.launch({ args: ['--no-sandbox'], channel: process.env.PW_CHANNEL || 'msedge' })
  .catch(() => chromium.launch({ args: ['--no-sandbox'], channel: 'chrome' }))
const ctx = await browser.newContext({ viewport: { width: 1366, height: 900 } })
const page = await ctx.newPage()
const errors = []
page.on('console', m => { if (m.type() === 'error') errors.push(m.text()) })
page.on('pageerror', e => errors.push(String(e)))
// CLS : la seule vitale que la sonde ne mesurait pas — une hiérarchie qui
// bouge après peinture doit être vue, pas supposée.
await page.addInitScript(() => {
  window.__CLS = 0
  try {
    new PerformanceObserver(l => {
      for (const e of l.getEntries()) if (!e.hadRecentInput) window.__CLS += e.value
    }).observe({ type: 'layout-shift', buffered: true })
  } catch {}
})
const t0 = Date.now()
await page.goto(base, { waitUntil: 'domcontentloaded', timeout: 30000 })
await page.waitForSelector('#v-dashboard:not([hidden]) h1', { timeout: 15000 })
const shellMs = Date.now() - t0
const nav = await page.evaluate(() => {
  const n = performance.getEntriesByType('navigation')[0]
  const res = performance.getEntriesByType('resource')
    .filter(r => /\.(js|css)(\?|$)/.test(r.name))
    .map(r => ({ name: r.name.split('/').pop(), transfer: r.transferSize, duration: Math.round(r.duration) }))
  return { dcl: Math.round(n.domContentLoadedEventEnd), load: Math.round(n.loadEventEnd), transfer: n.transferSize, resources: res }
})
const lcp = await page.evaluate(() => new Promise(resolve => {
  let value = null
  try {
    const obs = new PerformanceObserver(list => { const e = list.getEntries().at(-1); if (e) value = Math.round(e.startTime) })
    obs.observe({ type: 'largest-contentful-paint', buffered: true })
  } catch {}
  setTimeout(() => resolve(value), 1200)
}))
const cls = await page.evaluate(() => Math.round((window.__CLS || 0) * 1000) / 1000)
const t1 = Date.now()
await page.click('button[data-mode="bench"]')
await page.waitForSelector('#bench-config', { timeout: 15000 })
const benchMs = Date.now() - t1
const mem = await page.evaluate(() => performance.memory?.usedJSHeapSize ?? null)
console.log(JSON.stringify({ shellMs, benchMs, lcpMs: lcp, dclMs: nav.dcl, loadMs: nav.load, cls, resources: nav.resources, heapBytes: mem, errors }, null, 2))
await browser.close()
