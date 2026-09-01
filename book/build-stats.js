// Aggregate E1/E3 campaign CSVs into stats JSON used by figures + results chapter.
const fs = require('fs');
const path = require('path');

function parseCsv(file) {
  const lines = fs.readFileSync(file, 'utf8').trim().split(/\r?\n/);
  const head = lines[0].split(',');
  return lines.slice(1).map(l => {
    const cells = l.split(',');
    const o = {};
    head.forEach((h, i) => (o[h] = cells[i]));
    return o;
  });
}

const num = v => parseFloat(v);
const sorted = a => [...a].sort((x, y) => x - y);
function quantile(a, q) {
  const s = sorted(a);
  const pos = (s.length - 1) * q;
  const lo = Math.floor(pos), hi = Math.ceil(pos);
  return s[lo] + (s[hi] - s[lo]) * (pos - lo);
}
const mean = a => a.reduce((s, v) => s + v, 0) / a.length;
const median = a => quantile(a, 0.5);
function stddev(a) {
  const m = mean(a);
  return Math.sqrt(a.reduce((s, v) => s + (v - m) * (v - m), 0) / (a.length - 1));
}
const r1 = v => Math.round(v * 10) / 10;
const r2 = v => Math.round(v * 100) / 100;

const root = 'c:/thesis-cgo/thesis-cgo';
const e1 = parseCsv(path.join(root, 'data/e1_onsets.csv'));
const e3 = parseCsv(path.join(root, 'data/e3_events.csv'));

// ---------- E1 : detection lag per detector ----------
const detectors = [...new Set(e1.map(r => r.detector))];
const families = [...new Set(e1.map(r => r.family))];
const contexts = [...new Set(e1.map(r => r.context))];

const e1ByDetector = {};
for (const d of detectors) {
  const rows = e1.filter(r => r.detector === d);
  const valid = rows.filter(r => r.false_positive === 'false' && r.missed === 'false');
  const lags = valid.map(r => num(r.confirmed_lag_ms));
  const fp = rows.filter(r => r.false_positive === 'true').length;
  const missed = rows.filter(r => r.missed === 'true').length;
  e1ByDetector[d] = {
    n: rows.length,
    detected: valid.length,
    fp, missed,
    fpRate: r2((100 * fp) / rows.length),
    missRate: r2((100 * missed) / rows.length),
    lagMedian: r1(median(lags)),
    lagMean: r1(mean(lags)),
    lagP25: r1(quantile(lags, 0.25)),
    lagP75: r1(quantile(lags, 0.75)),
    lagP95: r1(quantile(lags, 0.95)),
    lagMin: r1(Math.min(...lags)),
    lagMax: r1(Math.max(...lags)),
    lagStd: r1(stddev(lags)),
  };
  // per family
  e1ByDetector[d].byFamily = {};
  for (const f of families) {
    const fr = rows.filter(r => r.family === f && r.false_positive === 'false' && r.missed === 'false')
      .map(r => num(r.confirmed_lag_ms));
    if (fr.length) e1ByDetector[d].byFamily[f] = { n: fr.length, median: r1(median(fr)), p95: r1(quantile(fr, 0.95)) };
  }
  // per context
  e1ByDetector[d].byContext = {};
  for (const c of contexts) {
    const cr = rows.filter(r => r.context === c && r.false_positive === 'false' && r.missed === 'false')
      .map(r => num(r.confirmed_lag_ms));
    if (cr.length) e1ByDetector[d].byContext[c] = { n: cr.length, median: r1(median(cr)) };
  }
}

// ECDF points of confirmed lag per detector (for figure)
const e1Ecdf = {};
for (const d of detectors) {
  const lags = sorted(e1.filter(r => r.detector === d && r.false_positive === 'false' && r.missed === 'false')
    .map(r => num(r.confirmed_lag_ms)));
  // subsample to <=60 points
  const pts = [];
  const step = Math.max(1, Math.floor(lags.length / 60));
  for (let i = 0; i < lags.length; i += step) pts.push([lags[i], (i + 1) / lags.length]);
  pts.push([lags[lags.length - 1], 1]);
  e1Ecdf[d] = pts;
}

// ---------- E3 : per arm metrics ----------
const arms = [...new Set(e3.map(r => r.arm))].sort();
const metricsE3 = ['qdi_ms', 'ttb_backlog_ms', 'voip_jitter_ms', 'voip_loss_pct', 'wasted_bytes', 'p95_ms', 'p95_small_ms', 'lfi', 'recovery_time_s', 'throughput_util', 'cost_ar_per_h'].filter(m => m in e3[0]);
const e3ByArm = {};
for (const a of arms) {
  const rows = e3.filter(r => r.arm === a);
  e3ByArm[a] = { n: rows.length };
  for (const m of metricsE3) {
    const vals = rows.map(r => num(r[m])).filter(v => !isNaN(v));
    e3ByArm[a][m] = { median: r2(median(vals)), mean: r2(mean(vals)), p95: r2(quantile(vals, 0.95)), std: r2(stddev(vals)), min: r2(Math.min(...vals)), max: r2(Math.max(...vals)) };
  }
  // grade distribution
  if ('grade' in e3[0]) {
    e3ByArm[a].grades = {};
    for (const r of rows) e3ByArm[a].grades[r.grade] = (e3ByArm[a].grades[r.grade] || 0) + 1;
  }
}

// per arm x link_profile QDI medians (for grouped bar figure)
const profiles = [...new Set(e3.map(r => r.link_profile))];
const e3QdiByArmProfile = {};
for (const a of arms) {
  e3QdiByArmProfile[a] = {};
  for (const p of profiles) {
    const vals = e3.filter(r => r.arm === a && r.link_profile === p).map(r => num(r.qdi_ms));
    if (vals.length) e3QdiByArmProfile[a][p] = r1(median(vals));
  }
}

const out = {
  e1: { total: e1.length, detectors, families, contexts, byDetector: e1ByDetector, ecdf: e1Ecdf,
        onsets: [...new Set(e1.map(r => r.onset_id))].length },
  e3: { total: e3.length, arms, profiles, byArm: e3ByArm, qdiByArmProfile: e3QdiByArmProfile },
};
fs.writeFileSync(path.join(__dirname, 'stats.json'), JSON.stringify(out, null, 2));
console.log('detectors:', detectors.join(','));
console.log('families:', families.join(','), '| contexts:', contexts.join(','));
console.log('arms:', arms.join(','), '| profiles:', profiles.join(','));
for (const d of detectors) console.log(d, 'median lag', e1ByDetector[d].lagMedian, 'fp%', e1ByDetector[d].fpRate, 'miss%', e1ByDetector[d].missRate);
for (const a of arms) console.log(a, 'QDI median', (e3ByArm[a].qdi_ms || {}).median, 'wasted', (e3ByArm[a].wasted_bytes || {}).median);
