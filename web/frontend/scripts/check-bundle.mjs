// Fail the build when the gzip-summed JS bundle exceeds the budget, or when
// the tree-shaken echarts chunk balloons (full-library re-import regression).
//
// Budget calibration — note honnête, 2026-11-14 : l'ancien cap « wall-kit »
// est retiré — audit DR-20 / Task 0.7 Step 4. Mesuré
// sur cet hôte à `3ef0fb5` : 358.4 KB gz TOTAL. echarts n'est pas encore un
// chunk séparé (in-chunk à ce jour) ; le watchdog echarts reste en attente du
// split manualChunks (Step 5). Cap TOTAL : 400 KB gz — headroom réelle,
// aucune dérogation.
import { gzipSync } from 'node:zlib';
import { readFileSync, readdirSync } from 'node:fs';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';

// Cap honnête 400 KB gz — ancien cap wall-kit retiré 2026-11-14 (audit DR-20 / Task 0.7).
const TOTAL_GZ_MAX = 400 * 1024; // honest total budget — 358.4 KB gz mesurés à `3ef0fb5`
const ECHARTS_GZ_MAX = 350 * 1024; // tree-shaken echarts watchdog — actif une fois echarts chunk séparé
const dir = fileURLToPath(new URL('../dist/assets', import.meta.url));
let files = [];
try { files = readdirSync(dir).filter((f) => f.endsWith('.js')); }
catch (e) { console.error(`FAIL : cannot read bundle dir ${dir} â€” ${e.message}`); process.exit(1); }
let total = 0;
let echartsGz = 0;
const rows = files.map((f) => {
  const raw = readFileSync(join(dir, f));
  const gz = gzipSync(raw).length;
  total += gz;
  if (f.startsWith('echarts-')) echartsGz = gz;
  return { f, raw: raw.length, gz };
}).sort((a, b) => b.gz - a.gz);
for (const r of rows) console.log(`${(r.gz / 1024).toFixed(1)} KB gz  ${(r.raw / 1024).toFixed(1)} KB raw  ${r.f}`);
console.log(`TOTAL gz : ${(total / 1024).toFixed(1)} KB`);
// echarts watchdog: only meaningful when echarts is a separate chunk. Since the
// wall-kit reunite build (single self-contained index chunk), echarts lives
// inside the main chunk — report that honestly instead of a fake 0.0 KB.
if (echartsGz > 0) {
  console.log(`echarts  : ${(echartsGz / 1024).toFixed(1)} KB gz / ${(ECHARTS_GZ_MAX / 1024).toFixed(0)} KB max`);
} else {
  console.log(`echarts  : in-chunk (no separate echarts-* chunk) — guarded by the TOTAL budget`);
}
let bad = false;
if (echartsGz > ECHARTS_GZ_MAX) {
  console.error(`FAIL : echarts chunk ${(echartsGz / 1024).toFixed(1)} KB gz > ${(ECHARTS_GZ_MAX / 1024).toFixed(0)} KB â€” full 'echarts' import re-added?`);
  bad = true;
}
if (total > TOTAL_GZ_MAX) {
  console.error(`FAIL : total ${(total / 1024).toFixed(1)} KB gz > ${(TOTAL_GZ_MAX / 1024).toFixed(0)} KB â€” budget 400 KB gz`);
  bad = true;
}
if (bad) process.exit(1);
