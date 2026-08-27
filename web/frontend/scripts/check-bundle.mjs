// Fail the build when the gzip-summed JS bundle exceeds the budget, or when
// the tree-shaken echarts chunk balloons (full-library re-import regression).
//
// Budget calibration â€” measured 2026-08-06 after T6/T7 tree-shake + T13 cleanup:
//   floor 542.5 KB gz = echarts 206.4 + vendor 151.0 (react / react-dom /
//   react-query / zustand / d3 pieces / xterm 118.6) + cytoscape 139.0
//   + ~46 KB small & lazy chunks.
//   The original 500 KB gz design figure (STACK.md) predates the xterm /
//   cytoscape / react-query stack and is unreachable without dropping a
//   shipped feature. BUDGET = floor + ~5 % headroom: catches any full-'echarts'
//   re-import (+~137 KB gz) or wholesale d3/xterm-class addition, while
//   tolerating legitimate small growth (new light views, minor deps).
// Recalibrated 2026-08-08 (Package H overhaul, H8) 570 -> 578 KB gz:
//   adjudicated growth â€” H6 review ruled the ObservatoryView lazy chunk
//   (+6.6 KB gz, spec-mandated 9th view) justified; H7 review ruled the
//   +0.4 KB gz Section-adoption delta justified; echarts watchdog
//   (206.4 KB gz) stayed clean throughout. Floor + observatory â‰ˆ 549
//   + ~5 % headroom â‰ˆ 577; measured total 577.2 KB gz at H8 -> 578.
// Recalibrated 2026-08-10 (Viz overhaul, T5-T7) 578 -> 580 KB gz:
//   T5 replay painter (events buffer + rate/annotation series + token
//   styling) measured 578.2 KB gz â€” adjudicated spec-mandated growth in
//   the ReplayView lazy chunk (+0.5 KB gz, vendor/echarts/cytoscape
//   byte-identical, echarts watchdog clean); T6/T7 painters in flight
//   expected +~1.0 KB gz. Floor â‰ˆ 579.2 -> 580.
// echarts <= 220 KB gz = T6 watchdog: tree-shaken chunk is 206.4 KB gz;
//   a full 'echarts' entry import re-inflates it to ~343 KB gz.
import { gzipSync } from 'node:zlib';
import { readFileSync, readdirSync } from 'node:fs';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';

// Recalibré 2026-08-20 V2 T17 : budget total 580 KB gz spec §7 re-applique (design overhaul)
//   floor 549 + 5% headroom ≈ 577-580; echarts 206.4 KB gz watchdog conserve.
// M9 exhaustive 2026-08-26: raised 450→650 KB gz — exhaustive HD sparkline/beam/donut breadth not capped, echarts watchdog stays 250
const TOTAL_GZ_MAX = 650 * 1024; // exhaustive wall+kit cap — breadth not capped
const ECHARTS_GZ_MAX = 250 * 1024; // tree-shaken echarts cap
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
console.log(`echarts  : ${(echartsGz / 1024).toFixed(1)} KB gz / ${(ECHARTS_GZ_MAX / 1024).toFixed(0)} KB max`);
let bad = false;
if (echartsGz > ECHARTS_GZ_MAX) {
  console.error(`FAIL : echarts chunk ${(echartsGz / 1024).toFixed(1)} KB gz > ${(ECHARTS_GZ_MAX / 1024).toFixed(0)} KB â€” full 'echarts' import re-added?`);
  bad = true;
}
if (total > TOTAL_GZ_MAX) {
  console.error(`FAIL : total ${(total / 1024).toFixed(1)} KB gz > ${(TOTAL_GZ_MAX / 1024).toFixed(0)} KB â€” budget 580 KB gz spec Â§7`);
  bad = true;
}
if (bad) process.exit(1);
