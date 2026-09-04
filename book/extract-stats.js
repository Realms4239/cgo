// extract-stats.js — réelles données des runs gelés → stats pour figures + tableaux thèse.
const fs = require('fs');
const path = require('path');

const RUNS = path.join(__dirname, 'data', 'runs');
const OUT = path.join(__dirname, 'data', 'stats.json');

function parseCSV(text) {
  const lines = text.trim().split(/\r?\n/);
  const head = lines[0].split(',');
  return lines.slice(1).filter(l => l).map(l => {
    const v = l.split(',');
    const o = {};
    head.forEach((h, i) => { o[h] = v[i]; });
    return o;
  });
}

function median(xs) { const s = [...xs].sort((a, b) => a - b); const n = s.length; return n % 2 ? s[(n - 1) / 2] : (s[n / 2 - 1] + s[n / 2]) / 2; }
function p95(xs) { const s = [...xs].sort((a, b) => a - b); return s[Math.min(s.length - 1, Math.ceil(0.95 * s.length) - 1)]; }

const rows = [];
for (const run of fs.readdirSync(RUNS)) {
  // run-smoke* = fixture de test du pipeline, pas une mesure : exclu du
  // corpus publié (sa seule ligne valid polluait P2|fq_codel|cubic n=1)
  if (run.startsWith('run-smoke')) continue;
  const f = path.join(RUNS, run, 'aqm_eval.csv');
  if (!fs.existsSync(f)) continue;
  for (const r of parseCSV(fs.readFileSync(f, 'utf8'))) rows.push({ run, ...r });
}

// cellules P2 récentes (les runs finaux deadline 220 : run-1788191429)
// cellules : clé profil|qdisc|cc, +|down quand la ligne porte direction=down.
// Les lignes sans colonne direction (150 runs historiques) sont up par
// construction — même règle que Go (results.Scan) : up et down ne fusionnent
// JAMAIS, sinon une cellule download (ex. small 373,8) contaminerait la
// médiane upload publiée. Les consommateurs par nom exact ('P2|cake|bbr')
// continuent de voir l'up sans changer une ligne.
const byCell = {};
for (const r of rows) {
  if (r.gate_status === 'invalid') continue;
  const dir = (r.direction && r.direction !== 'up') ? '|down' : '';
  const k = `${r.profile}|${r.qdisc}|${r.cc}${dir}`;
  (byCell[k] ||= []).push(r);
}
const cells = Object.entries(byCell).map(([k, rs]) => {
  const small = rs.map(r => +r.small_p95_ms).filter(v => v > 0);
  const rtt = rs.map(r => +r.rtt_p95_ms).filter(v => v > 0);
  const qdi = rs.map(r => +r.qdi_ms).filter(v => v > 0);
  const voip = rs.map(r => +r.voip_r).filter(v => v > 0);
  const dl = rs.map(r => +r.deadline_ok_pct);
  const gp = rs.map(r => +r.bulk_goodput_mbps).filter(v => v > 0);
  const cost = rs.map(r => +r.cost_ar_per_h).filter(v => v >= 0);
  return {
    cell: k, n: rs.length,
    small_p95_med: median(small), small_p95_p95: p95(small),
    rtt_p95_med: median(rtt), qdi_med: median(qdi),
    voip_med: voip.length ? median(voip) : null,
    deadline_med: median(dl), goodput_med: median(gp),
    cost_med: cost.length ? median(cost) : null,
  };
});

// le run final P2 deadline 220 (la démonstration clé)
const finalRun = rows.filter(r => r.run === 'run-1788191429');

const stats = {
  total_runs: fs.readdirSync(RUNS).length,
  total_rows: rows.length,
  valid: rows.filter(r => r.gate_status === 'valid').length,
  quarantined: rows.filter(r => r.gate_status === 'invalid').length,
  cells: cells.sort((a, b) => a.cell.localeCompare(b.cell)),
  final_run: finalRun,
};
fs.writeFileSync(OUT, JSON.stringify(stats, null, 2));
console.log('runs:', stats.total_runs, '| lignes:', stats.total_rows,
  '| valid:', stats.valid, '| quarantaine:', stats.quarantined, '| cellules:', stats.cells.length);
console.log('final run (deadline 220):', finalRun.length, 'lignes');
