// Agrège les CSV gelés book/data-real/runs/*/aqm_eval.csv → book/stats.json
// Médianes, QDI, VoIP R, coût par cellule. Filtre les outliers matériels (overflow 2^64).
const fs = require('fs');
const path = require('path');

function parseCsv(file) {
  const raw = fs.readFileSync(file, 'utf8').trim();
  if (!raw) return [];
  const lines = raw.split(/\r?\n/);
  const head = lines[0].split(',').map(s => s.trim());
  const out = [];
  for (let i = 1; i < lines.length; i++) {
    const l = lines[i].trim();
    if (!l) continue;
    const cells = l.split(',');
    if (cells.length < head.length) continue; // malformed
    const o = {};
    head.forEach((h, idx) => (o[h] = (cells[idx] || '').trim()));
    // skip header duplicates that slipped via cat concatenation
    if (o.profile === 'profile' || o.run_id === 'run_id') continue;
    out.push(o);
  }
  return out;
}

const num = v => parseFloat(v);
const sorted = a => [...a].sort((x, y) => x - y);
function quantile(a, q) {
  if (!a.length) return null;
  const s = sorted(a);
  const pos = (s.length - 1) * q;
  const lo = Math.floor(pos), hi = Math.ceil(pos);
  if (lo === hi) return s[lo];
  return s[lo] + (s[hi] - s[lo]) * (pos - lo);
}
const mean = a => a.length ? a.reduce((s, v) => s + v, 0) / a.length : null;
const median = a => quantile(a, 0.5);
function stddev(a) {
  if (a.length < 2) return 0;
  const m = mean(a);
  return Math.sqrt(a.reduce((s, v) => s + (v - m) * (v - m), 0) / (a.length - 1));
}
const r1 = v => v == null || isNaN(v) ? null : Math.round(v * 10) / 10;
const r2 = v => v == null || isNaN(v) ? null : Math.round(v * 100) / 100;
const r3 = v => v == null || isNaN(v) ? null : Math.round(v * 1000) / 1000;

function summary(vals) {
  const v = vals.filter(x => typeof x === 'number' && !isNaN(x) && isFinite(x));
  if (!v.length) return null;
  return {
    n: v.length,
    median: r2(median(v)),
    mean: r2(mean(v)),
    p25: r2(quantile(v, 0.25)),
    p75: r2(quantile(v, 0.75)),
    p95: r2(quantile(v, 0.95)),
    min: r2(Math.min(...v)),
    max: r2(Math.max(...v)),
    std: r2(stddev(v)),
  };
}

// seuils anti-overflow (2^64 et 1.2 Tbps fantaisistes)
const validators = {
  rtt_p50_ms: v => v >= 0 && v < 10000,
  rtt_p95_ms: v => v >= 0 && v < 10000,
  small_p95_ms: v => v >= 0 && v < 10000,
  qdi_ms: v => v >= 0 && v < 10000,
  voip_r: v => v >= 0 && v <= 100,
  deadline_ok_pct: v => v >= 0 && v <= 100,
  bulk_goodput_mbps: v => v >= 0 && v < 50000, // 50 Gbps max, 1.2 Tbps exclu
  drops: v => v >= 0 && v < 1e7,
  retransmissions: v => v >= 0 && v < 1e7,
  wasted_bytes: v => v >= 0 && v < 1e10, // 10 Go
  cost_ar_per_h: v => v >= 0 && v < 1e6,
  cpu_pct: v => v >= 0 && v <= 100,
};

const runsRoot = path.join(__dirname, 'data-real', 'runs');
if (!fs.existsSync(runsRoot)) {
  console.error('runsRoot introuvable:', runsRoot);
  process.exit(1);
}
const runDirs = fs.readdirSync(runsRoot).filter(d => {
  const p = path.join(runsRoot, d);
  return fs.statSync(p).isDirectory();
});
// run-smoke est un artefact de test, hors campagne gelée (133 runs). On l'exclut des stats mais on le garde sur disque.
const runDirsForStats = runDirs.filter(d => d !== 'run-smoke');

let allRows = [];
let runsWithData = 0;
let runsEmpty = [];
let filesMissing = [];
for (const rd of runDirsForStats) {
  const csv = path.join(runsRoot, rd, 'aqm_eval.csv');
  if (!fs.existsSync(csv)) { filesMissing.push(rd); continue; }
  const rows = parseCsv(csv);
  if (rows.length) runsWithData++;
  else runsEmpty.push(rd);
  // tag run_id for debugging but keep original fields
  for (const r of rows) r._run_dir = rd;
  allRows = allRows.concat(rows);
}

// --- helpers to collect filtered values ---
function colFiltered(rows, col) {
  const v = validators[col];
  return rows
    .map(r => num(r[col]))
    .filter(nv => !isNaN(nv) && (v ? v(nv) : true));
}

const metrics = ['rtt_p50_ms','rtt_p95_ms','small_p95_ms','qdi_ms','voip_r','deadline_ok_pct','bulk_goodput_mbps','drops','retransmissions','wasted_bytes','cost_ar_per_h','cpu_pct'];
const profiles = [...new Set(allRows.map(r => r.profile))].filter(Boolean).sort();
const qdiscs = [...new Set(allRows.map(r => r.qdisc))].filter(Boolean).sort();
const ccs = [...new Set(allRows.map(r => r.cc))].filter(Boolean).sort();

function metricsSummary(rows) {
  const out = {};
  for (const m of metrics) {
    const vals = colFiltered(rows, m);
    // only emit if column existed in at least one row
    const existed = rows.some(r => r[m] !== undefined);
    if (!existed) continue;
    const s = summary(vals);
    if (s) out[m] = s;
    // keep also raw count of column presence
    out[m + '_raw_n'] = rows.filter(r => r[m] !== undefined && r[m] !== '').length;
  }
  // gate_status distribution
  const gates = {};
  for (const r of rows) { const g = r.gate_status || 'unknown'; gates[g] = (gates[g] || 0) + 1; }
  if (Object.keys(gates).length) out.gate_status = gates;
  out.n = rows.length;
  return out;
}

const overall = metricsSummary(allRows);

const byQdisc = {};
for (const q of qdiscs) byQdisc[q] = metricsSummary(allRows.filter(r => r.qdisc === q));

const byProfile = {};
for (const p of profiles) byProfile[p] = metricsSummary(allRows.filter(r => r.profile === p));

const byCc = {};
for (const c of ccs) byCc[c] = metricsSummary(allRows.filter(r => r.cc === c));

const byQdiscProfile = {};
for (const q of qdiscs) for (const p of profiles) {
  const key = `${q}@${p}`;
  const subset = allRows.filter(r => r.qdisc === q && r.profile === p);
  if (subset.length) byQdiscProfile[key] = metricsSummary(subset);
}

const byProfileQdisc = {};
for (const p of profiles) for (const q of qdiscs) {
  const key = `${p}@${q}`;
  const subset = allRows.filter(r => r.profile === p && r.qdisc === q);
  if (subset.length) byProfileQdisc[key] = metricsSummary(subset);
}

// QDI spécifique : médiane par qdisc (réutilisée par figures)
const qdiByQdisc = {};
for (const q of qdiscs) {
  const vals = colFiltered(allRows.filter(r => r.qdisc === q), 'qdi_ms');
  if (vals.length) qdiByQdisc[q] = r1(median(vals));
}

// small_p95 spécifique : médiane par qdisc
const smallByQdisc = {};
for (const q of qdiscs) {
  const vals = colFiltered(allRows.filter(r => r.qdisc === q), 'small_p95_ms');
  if (vals.length) smallByQdisc[q] = r1(median(vals));
}

// VoIP R médian par qdisc / profile
const voipByQdisc = {};
for (const q of qdiscs) {
  const vals = colFiltered(allRows.filter(r => r.qdisc === q), 'voip_r');
  if (vals.length) voipByQdisc[q] = r1(median(vals));
}

// coût médian par qdisc (Ar/h)
const costByQdisc = {};
for (const q of qdiscs) {
  const vals = colFiltered(allRows.filter(r => r.qdisc === q), 'cost_ar_per_h');
  if (vals.length) costByQdisc[q] = r2(median(vals));
}

// deadline
const deadlineByQdisc = {};
for (const q of qdiscs) {
  const vals = colFiltered(allRows.filter(r => r.qdisc === q), 'deadline_ok_pct');
  if (vals.length) deadlineByQdisc[q] = r1(median(vals));
}

// ECDF small_p95 par qdisc (<=60 points, pour figures)
const ecdfSmall = {};
for (const q of qdiscs) {
  const vals = sorted(colFiltered(allRows.filter(r => r.qdisc === q), 'small_p95_ms'));
  if (!vals.length) continue;
  const pts = [];
  const step = Math.max(1, Math.floor(vals.length / 60));
  for (let i = 0; i < vals.length; i += step) pts.push([r1(vals[i]), r3((i + 1) / vals.length)]);
  if (pts[pts.length - 1][0] !== r1(vals[vals.length - 1])) pts.push([r1(vals[vals.length - 1]), 1]);
  else pts[pts.length - 1][1] = 1;
  ecdfSmall[q] = pts;
}

const out = {
  meta: {
    generatedAt: new Date().toISOString(),
    runsTotalDisk: runDirs.length,
    runsTotal: runDirsForStats.length, // gelés (hors run-smoke), attendu 133
    runsWithData,
    runsEmpty: runsEmpty.length,
    runsEmptyList: runsEmpty.slice(0, 20),
    filesMissing: filesMissing.length,
    totalRows: allRows.length,
    totalRowsFiltered: overall.n || 0,
    profiles,
    qdiscs,
    ccs,
    // pour traçabilité : liste des 5 runs d'exemple
    sampleRuns: runDirs.slice(0, 5),
  },
  overall,
  byQdisc,
  byProfile,
  byCc,
  byQdiscProfile,
  byProfileQdisc,
  // raccourcis exploités par les figures / chapitres
  qdiByQdisc,
  smallByQdisc,
  voipByQdisc,
  costByQdisc,
  deadlineByQdisc,
  ecdfSmall,
  // alias explicites pour la vérif Step 3 : grep small_p95 doit matcher
  small_p95: overall.small_p95_ms || null,
  qdi_ms: overall.qdi_ms || null,
  voip_r: overall.voip_r || null,
  cost_ar_per_h: overall.cost_ar_per_h || null,
};

const dst = path.join(__dirname, 'stats.json');
fs.writeFileSync(dst, JSON.stringify(out, null, 2), 'utf8');

console.log(`runs: total ${runDirsForStats.length} gelés (disque ${runDirs.length} avec run-smoke) dont ${runsWithData} avec données, ${runsEmpty.length} vides, ${out.meta.totalRows} lignes`);
console.log(`profiles: ${profiles.join(',')} | qdiscs: ${qdiscs.join(',')} | cc: ${ccs.join(',')}`);
console.log(`overall small_p95 median ${out.overall.small_p95_ms ? out.overall.small_p95_ms.median : 'na'} (n=${out.overall.small_p95_ms ? out.overall.small_p95_ms.n : 0}) p95 ${out.overall.small_p95_ms ? out.overall.small_p95_ms.p95 : 'na'}`);
if (out.overall.qdi_ms) console.log(`overall qdi_ms median ${out.overall.qdi_ms.median} (n=${out.overall.qdi_ms.n})`);
if (out.overall.voip_r) console.log(`overall voip_r median ${out.overall.voip_r.median} (n=${out.overall.voip_r.n})`);
for (const q of qdiscs) {
  const s = byQdisc[q];
  console.log(`${q} n=${s.n} small_p95 med ${s.small_p95_ms ? s.small_p95_ms.median : 'na'} deadline med ${s.deadline_ok_pct ? s.deadline_ok_pct.median : 'na'} cost med ${s.cost_ar_per_h ? s.cost_ar_per_h.median : 'na'} QDI med ${s.qdi_ms ? s.qdi_ms.median : 'na'}`);
}
for (const p of profiles) {
  const s = byProfile[p];
  console.log(`profile ${p} n=${s.n} small_p95 med ${s.small_p95_ms ? s.small_p95_ms.median : 'na'} deadline med ${s.deadline_ok_pct ? s.deadline_ok_pct.median : 'na'}`);
}
console.log('stats.json écrit →', dst);
