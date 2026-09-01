// Figures 9-16 : data-driven results (E1 detection, E3 mitigation) + reproducibility chain.
const fs = require('fs');
const path = require('path');
const { PAL, Svg, chartFrame, legend } = require('./fig-lib');

const OUT = path.join(__dirname, 'figures');
fs.mkdirSync(OUT, { recursive: true });
const save = (name, svg) => fs.writeFileSync(path.join(OUT, name + '.svg'), svg.toString());
const S = JSON.parse(fs.readFileSync(path.join(__dirname, 'stats.json'), 'utf8'));

const DET_COLORS = { ks: PAL.blue, gradient: PAL.teal, baddrd: PAL.purple, ewma: PAL.amber, threshold: PAL.red };
const DET_LABEL = { ks: 'K-S', gradient: 'Gradient', baddrd: 'BaDDrD', ewma: 'EWMA', threshold: 'Seuil' };
const ARM_COLORS = { A0: PAL.red, A1: PAL.amber, A2: PAL.teal, A3: PAL.blue };
const ARM_LABEL = { A0: 'A0 pfifo_fast', A1: 'A1 fq_codel défaut', A2: 'A2 fq_codel réglé', A3: 'A3 adaptatif' };
const fr1 = v => String(v).replace('.', ',');

// ---------------------------------------------------------------- fig09 : ECDF du délai de détection (E1)
(function fig09() {
  const s = new Svg(900, 500);
  const m = { left: 85, top: 40, right: 860, bottom: 410 };
  const { sx, sy } = chartFrame(s, m, [800, 4100], [0, 1], {
    xLabel: 'Délai de détection confirmé (ms)', yLabel: 'Fraction cumulée des onsets', ny: 5, xGrid: true,
    yFmt: t => fr1(t),
  });
  // median guide
  s.line(sx(800), sy(0.5), sx(4100), sy(0.5), { stroke: PAL.axis, sw: 1.2, dash: '3 4' });
  const order = ['ks', 'gradient', 'baddrd', 'ewma', 'threshold'];
  for (const d of order) {
    const pts = S.e1.ecdf[d];
    let dstr = `M${sx(pts[0][0]).toFixed(1)} ${sy(pts[0][1]).toFixed(1)}`;
    for (let i = 1; i < pts.length; i++) {
      dstr += ` L${sx(pts[i][0]).toFixed(1)} ${sy(pts[i - 1][1]).toFixed(1)} L${sx(pts[i][0]).toFixed(1)} ${sy(pts[i][1]).toFixed(1)}`;
    }
    s.path(dstr, { stroke: DET_COLORS[d], sw: 2.4 });
  }
  // median annotation for ks and threshold
  const ksMed = S.e1.byDetector.ks.lagMedian, thMed = S.e1.byDetector.threshold.lagMedian;
  s.line(sx(ksMed), sy(0.5), sx(ksMed), m.bottom, { stroke: PAL.blue, sw: 1.3, dash: '4 3' });
  s.text(sx(ksMed), m.bottom - 8, '1 444 ms', { size: 11.5, fill: PAL.blue, weight: '700', anchor: 'middle' });
  s.line(sx(thMed), sy(0.5), sx(thMed), m.bottom, { stroke: PAL.red, sw: 1.3, dash: '4 3' });
  s.text(sx(thMed), m.bottom - 8, '2 329 ms', { size: 11.5, fill: PAL.red, weight: '700', anchor: 'middle' });
  legend(s, 90, 486, order.map(d => [DET_LABEL[d], DET_COLORS[d], 'line']));
  save('fig09-ecdf-lag', s);
})();

// ---------------------------------------------------------------- fig10 : comparaison des détecteurs (médiane + IQR, FP, ratés)
(function fig10() {
  const s = new Svg(900, 500);
  const m = { left: 85, top: 46, right: 620, bottom: 420 };
  const order = ['ks', 'gradient', 'baddrd', 'ewma', 'threshold'];
  const { sy } = chartFrame(s, m, [0, 1], [0, 3600], {
    yLabel: 'Délai de détection (ms)', ny: 6, xTicks: [],
  });
  const bw = 62, gap = (m.right - m.left - order.length * bw) / (order.length + 1);
  order.forEach((d, i) => {
    const st = S.e1.byDetector[d];
    const x = m.left + gap + i * (bw + gap);
    // IQR box + p95 whisker
    s.line(x + bw / 2, sy(st.lagP95), x + bw / 2, sy(st.lagP75), { stroke: DET_COLORS[d], sw: 1.6 });
    s.line(x + bw / 2 - 12, sy(st.lagP95), x + bw / 2 + 12, sy(st.lagP95), { stroke: DET_COLORS[d], sw: 1.6 });
    s.rect(x, sy(st.lagP75), bw, sy(st.lagP25) - sy(st.lagP75), { fill: DET_COLORS[d], op: 0.25, rx: 3, stroke: DET_COLORS[d], sw: 1.4 });
    s.line(x, sy(st.lagMedian), x + bw, sy(st.lagMedian), { stroke: DET_COLORS[d], sw: 3 });
    s.text(x + bw / 2, sy(st.lagP95) - 8, fr1(st.lagP95) + '', { anchor: 'middle', size: 10.5, fill: PAL.sub });
    s.text(x + bw / 2, sy(st.lagMedian) - 6, String(st.lagMedian).replace('.', ','), { anchor: 'middle', size: 11.5, weight: '700', fill: DET_COLORS[d] });
    s.text(x + bw / 2, m.bottom + 20, DET_LABEL[d], { anchor: 'middle', size: 12, weight: '600' });
    s.text(x + bw / 2, m.bottom + 38, 'FP ' + fr1(st.fpRate) + ' %', { anchor: 'middle', size: 10.5, fill: PAL.sub });
    s.text(x + bw / 2, m.bottom + 54, 'ratés ' + fr1(st.missRate) + ' %', { anchor: 'middle', size: 10.5, fill: PAL.sub });
  });
  // right panel : compromis précocité / fausses alarmes
  s.rect(650, 60, 230, 330, { rx: 10, fill: PAL.grayBg, stroke: PAL.axis, sw: 1.2 });
  s.text(765, 88, 'Lecture du compromis', { anchor: 'middle', size: 13, weight: '700' });
  const rows = [
    ['K-S', 'le plus précoce,', 'FP 5,39 %'],
    ['Gradient', 'précoce et sobre,', 'FP 2,45 %'],
    ['BaDDrD', 'régulier,', 'FP 1,96 %'],
    ['EWMA', 'zéro fausse alarme,', 'FP 0 %'],
    ['Seuil', 'le plus tardif,', 'variance élevée'],
  ];
  const ro = ['ks', 'gradient', 'baddrd', 'ewma', 'threshold'];
  rows.forEach((r, i) => {
    const y = 118 + i * 54;
    s.rect(664, y - 11, 12, 12, { fill: DET_COLORS[ro[i]], rx: 3 });
    s.text(684, y, r[0], { size: 12, weight: '700' });
    s.text(684, y + 16, r[1], { size: 11, fill: PAL.sub });
    s.text(684, y + 31, r[2], { size: 11, fill: PAL.sub });
  });
  s.text(90, 480, 'Boîte : intervalle interquartile (P25-P75) ; trait épais : médiane ; moustache : P95 ; n = 204 onsets par détecteur.',
    { size: 11.5, fill: PAL.sub });
  save('fig10-detecteurs', s);
})();

// ---------------------------------------------------------------- fig11 : délai médian par famille de scénario
(function fig11() {
  const s = new Svg(900, 480);
  const m = { left: 85, top: 46, right: 860, bottom: 390 };
  const fams = ['step', 'ramp', 'diurnal'];
  const FAM_LABEL = { step: 'Échelon (step)', ramp: 'Rampe (ramp)', diurnal: 'Diurne (diurnal)' };
  const order = ['ks', 'gradient', 'baddrd', 'ewma', 'threshold'];
  const { sy } = chartFrame(s, m, [0, 1], [0, 3000], { yLabel: 'Délai médian de détection (ms)', ny: 6, xTicks: [] });
  const groupW = (m.right - m.left) / fams.length;
  const bw = 26, inGap = 8;
  fams.forEach((f, gi) => {
    const gx = m.left + gi * groupW + (groupW - (order.length * bw + (order.length - 1) * inGap)) / 2;
    order.forEach((d, i) => {
      const v = S.e1.byDetector[d].byFamily[f].median;
      const x = gx + i * (bw + inGap);
      s.rect(x, sy(v), bw, m.bottom - sy(v), { fill: DET_COLORS[d], rx: 3, op: 0.9 });
      s.text(x + bw / 2, sy(v) - 6, String(Math.round(v)), { anchor: 'middle', size: 10, fill: PAL.sub });
    });
    s.text(m.left + gi * groupW + groupW / 2, m.bottom + 24, FAM_LABEL[f], { anchor: 'middle', size: 12.5, weight: '600' });
  });
  legend(s, 90, 460, order.map(d => [DET_LABEL[d], DET_COLORS[d], 'sq']));
  save('fig11-familles', s);
})();

// ---------------------------------------------------------------- fig12 : QDI médian par bras et par profil (E3)
(function fig12() {
  const s = new Svg(900, 500);
  const m = { left: 80, top: 46, right: 860, bottom: 400 };
  const profs = ['clean', '50ms', '160ms', '4g-loss'];
  const PROF_LABEL = { clean: 'Profil propre', '50ms': 'RTT 50 ms', '160ms': 'RTT 160 ms', '4g-loss': '4G avec pertes' };
  const arms = ['A0', 'A1', 'A2', 'A3'];
  const { sy } = chartFrame(s, m, [0, 1], [0, 125], { yLabel: 'QDI médian (ms)', ny: 5, xTicks: [] });
  const groupW = (m.right - m.left) / profs.length;
  const bw = 34, inGap = 9;
  profs.forEach((p, gi) => {
    const gx = m.left + gi * groupW + (groupW - (arms.length * bw + (arms.length - 1) * inGap)) / 2;
    arms.forEach((a, i) => {
      const v = S.e3.qdiByArmProfile[a][p];
      const x = gx + i * (bw + inGap);
      s.rect(x, sy(v), bw, m.bottom - sy(v), { fill: ARM_COLORS[a], rx: 3, op: 0.9 });
      s.text(x + bw / 2, sy(v) - 6, fr1(v), { anchor: 'middle', size: 10.5, fill: PAL.sub });
    });
    s.text(m.left + gi * groupW + groupW / 2, m.bottom + 24, PROF_LABEL[p], { anchor: 'middle', size: 12.5, weight: '600' });
  });
  legend(s, 90, 478, arms.map(a => [ARM_LABEL[a], ARM_COLORS[a], 'sq']));
  save('fig12-qdi', s);
})();

// ---------------------------------------------------------------- fig13 : quatre métriques E3 côte à côte
(function fig13() {
  const s = new Svg(940, 520);
  const arms = ['A0', 'A1', 'A2', 'A3'];
  const panels = [
    { key: 'qdi_ms', label: 'QDI (ms)', max: 100 },
    { key: 'p95_ms', label: 'P95 du délai (ms)', max: 25 },
    { key: 'voip_jitter_ms', label: 'Gigue VoIP (ms)', max: 30 },
    { key: 'ttb_backlog_ms', label: 'TTB (ms)', max: 600 },
  ];
  const pw = 205, ph = 330, px0 = 70, py0 = 60, gapx = 22;
  panels.forEach((p, pi) => {
    const left = px0 + pi * (pw + gapx);
    const m = { left, top: py0, right: left + pw, bottom: py0 + ph };
    const { sy } = chartFrame(s, m, [0, 1], [0, p.max], { ny: 4, xTicks: [] });
    s.text(left + pw / 2, py0 - 16, p.label, { anchor: 'middle', size: 12.5, weight: '700' });
    const bw = 34, gap = (pw - arms.length * bw) / (arms.length + 1);
    arms.forEach((a, i) => {
      const v = S.e3.byArm[a][p.key].median;
      const x = left + gap + i * (bw + gap);
      s.rect(x, sy(v), bw, m.bottom - sy(v), { fill: ARM_COLORS[a], rx: 3, op: 0.9 });
      s.text(x + bw / 2, sy(v) - 6, fr1(v), { anchor: 'middle', size: 10.5, fill: PAL.sub });
      s.text(x + bw / 2, m.bottom + 18, a, { anchor: 'middle', size: 11.5, weight: '600' });
    });
  });
  legend(s, 80, 500, arms.map(a => [ARM_LABEL[a], ARM_COLORS[a], 'sq']));
  save('fig13-metriques-e3', s);
})();

// ---------------------------------------------------------------- fig14 : distribution des notes (grades) par bras
(function fig14() {
  const s = new Svg(900, 460);
  const m = { left: 90, top: 50, right: 640, bottom: 380 };
  const arms = ['A0', 'A1', 'A2', 'A3'];
  const gradeOrder = ['A', 'B', 'C', 'D'];
  const GRADE_COLORS = { A: PAL.teal, B: PAL.blue, C: PAL.amber, D: PAL.red };
  const { sy } = chartFrame(s, m, [0, 1], [0, 100], {
    yLabel: 'Part des exécutions (%)', ny: 5, xTicks: [],
  });
  const bw = 84, gap = (m.right - m.left - arms.length * bw) / (arms.length + 1);
  arms.forEach((a, i) => {
    const g = S.e3.byArm[a].grades || {}, n = S.e3.byArm[a].n;
    const x = m.left + gap + i * (bw + gap);
    let acc = 0;
    for (const gr of gradeOrder) {
      const c = g[gr] || 0;
      if (!c) continue;
      const pct = (100 * c) / n;
      const y1 = sy(acc + pct), y0 = sy(acc);
      s.rect(x, y1, bw, y0 - y1, { fill: GRADE_COLORS[gr], op: 0.9 });
      if (pct >= 8) s.text(x + bw / 2, (y0 + y1) / 2 + 4, gr + ' : ' + Math.round(pct) + ' %', { anchor: 'middle', size: 11, fill: '#ffffff', weight: '700' });
      acc += pct;
    }
    s.text(x + bw / 2, m.bottom + 22, a, { anchor: 'middle', size: 12.5, weight: '700' });
    s.text(x + bw / 2, m.bottom + 40, 'n = ' + n, { anchor: 'middle', size: 11, fill: PAL.sub });
  });
  // right panel : grading scale
  s.rect(670, 70, 210, 240, { rx: 10, fill: PAL.grayBg, stroke: PAL.axis, sw: 1.2 });
  s.text(775, 98, 'Barème de qualité', { anchor: 'middle', size: 13, weight: '700' });
  const scale = [['A', 'expérience excellente'], ['B', 'dégradation légère'], ['C', 'dégradation sensible'], ['D', 'lien inutilisable']];
  scale.forEach((r, i) => {
    const y = 128 + i * 42;
    s.rect(686, y - 12, 14, 14, { fill: GRADE_COLORS[r[0]], rx: 3 });
    s.text(708, y, r[0] + ' : ' + r[1], { size: 11.5 });
  });
  save('fig14-grades', s);
})();

// ---------------------------------------------------------------- fig15 : octets gaspillés et coût données par bras
(function fig15() {
  const s = new Svg(900, 470);
  const arms = ['A0', 'A1', 'A2', 'A3'];
  // left panel : wasted bytes (ko)
  const mL = { left: 85, top: 56, right: 430, bottom: 380 };
  const { sy: syL } = chartFrame(s, mL, [0, 1], [0, 80], { yLabel: 'Octets gaspillés (ko, médiane)', ny: 4, xTicks: [] });
  s.text((mL.left + mL.right) / 2, 34, 'Gaspillage par événement', { anchor: 'middle', size: 13, weight: '700' });
  const bwL = 56, gapL = (mL.right - mL.left - arms.length * bwL) / (arms.length + 1);
  arms.forEach((a, i) => {
    const v = S.e3.byArm[a].wasted_bytes.median / 1000;
    const x = mL.left + gapL + i * (bwL + gapL);
    s.rect(x, syL(v), bwL, mL.bottom - syL(v), { fill: ARM_COLORS[a], rx: 3, op: 0.9 });
    s.text(x + bwL / 2, syL(v) - 6, fr1(Math.round(v * 10) / 10), { anchor: 'middle', size: 11, fill: PAL.sub });
    s.text(x + bwL / 2, mL.bottom + 20, a, { anchor: 'middle', size: 12, weight: '600' });
  });
  // right panel : cost Ar/h
  const mR = { left: 545, top: 56, right: 860, bottom: 380 };
  const { sy: syR } = chartFrame(s, mR, [0, 1], [0, 10], { yLabel: 'Coût données (Ar/h, médiane)', ny: 5, xTicks: [] });
  s.text((mR.left + mR.right) / 2, 34, 'Coût monétaire du gaspillage', { anchor: 'middle', size: 13, weight: '700' });
  const bwR = 50, gapR = (mR.right - mR.left - arms.length * bwR) / (arms.length + 1);
  arms.forEach((a, i) => {
    const v = S.e3.byArm[a].cost_ar_per_h.median;
    const x = mR.left + gapR + i * (bwR + gapR);
    s.rect(x, syR(v), bwR, mR.bottom - syR(v), { fill: ARM_COLORS[a], rx: 3, op: 0.9 });
    s.text(x + bwR / 2, syR(v) - 6, fr1(v), { anchor: 'middle', size: 11, fill: PAL.sub });
    s.text(x + bwR / 2, mR.bottom + 20, a, { anchor: 'middle', size: 12, weight: '600' });
  });
  legend(s, 90, 448, arms.map(a => [ARM_LABEL[a], ARM_COLORS[a], 'sq']));
  save('fig15-gaspillage', s);
})();

// ---------------------------------------------------------------- fig16 : chaîne de reproductibilité
(function fig16() {
  const s = new Svg(940, 430);
  s.arrowDef('aG', PAL.slate);
  const y = 90, h = 96, w = 158, gap = 30;
  const steps = [
    { t: ['Configuration', 'YAML + graine', '(seed) fixée'], c: PAL.blue, bg: PAL.blueBg },
    { t: ['Exécution', 'campagne E1/E2/E3', 'cgo-linux'], c: PAL.teal, bg: PAL.tealBg },
    { t: ['Journaux CSV', 'onsets, événements,', 'décisions'], c: PAL.amber, bg: PAL.amberBg },
    { t: ['Manifeste', 'SHA-256 chaîné', 'de chaque artefact'], c: PAL.purple, bg: PAL.purpleBg },
    { t: ['Agrégation', 'stats + figures', 'déterministes'], c: PAL.red, bg: PAL.redBg },
  ];
  steps.forEach((st, i) => {
    const x = 40 + i * (w + gap);
    s.box(x, y, w, h, st.t, { fill: st.bg, stroke: st.c, size: 12, lh: 17 });
    if (i < steps.length - 1) s.line(x + w, y + h / 2, x + w + gap, y + h / 2, { stroke: PAL.slate, sw: 2.4, marker: 'aG' });
  });
  // verification loop
  s.path(`M${40 + 4 * (w + gap) + w / 2} ${y + h} C ${40 + 4 * (w + gap) + w / 2} ${y + h + 90} ${40 + w / 2} ${y + h + 90} ${40 + w / 2} ${y + h + 4}`,
    { stroke: PAL.slate, sw: 2, dash: '7 5', marker: 'aG' });
  s.text(470, y + h + 78, 'Vérification : toute tierce partie rejoue la campagne et retrouve les mêmes tables et les mêmes figures',
    { anchor: 'middle', size: 12.5, fill: PAL.sub, italic: true });
  s.box(40, 320, 858, 66, ['Intégrité : chaque ligne du manifeste inclut le condensat de la ligne précédente (chaînage), toute altération',
    'a posteriori d\u2019un fichier de résultats invalide la chaîne complète et devient détectable.'],
    { fill: PAL.grayBg, stroke: PAL.slate, size: 12.5, weight: 'normal', lh: 20 });
  save('fig16-reproductibilite', s);
})();

console.log('figures 09-16 written to', OUT);
