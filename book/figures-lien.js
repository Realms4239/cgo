// figures-lien.js — figures du mémoire LIEN depuis les données réelles (data/stats.json).
const fs = require('fs');
const path = require('path');
const { PAL, Svg, chartFrame, legend } = require('./fig-lib');

const OUT = path.join(__dirname, 'figures');
fs.mkdirSync(OUT, { recursive: true });
const S = JSON.parse(fs.readFileSync(path.join(__dirname, 'data', 'stats.json'), 'utf8'));
const save = (name, svg) => fs.writeFileSync(path.join(OUT, name + '.svg'), svg.toString());

// ---------------------------------------------------------------- figR1 : topologie poste-client
(function figR1() {
  const s = new Svg(900, 380);
  // poste opérateur
  s.rect(40, 120, 230, 170, { fill: '#eef2f7', stroke: PAL.blue, sw: 2 });
  s.text(155, 150, 'POSTE CLIENT MÉTÉO', { size: 14, weight: '700', anchor: 'middle', fill: PAL.ink });
  s.text(155, 175, 'cgo : sonde small + ping + bulk', { size: 11, anchor: 'middle', fill: PAL.sub });
  s.text(155, 195, 'audit + campagne + dashboard', { size: 11, anchor: 'middle', fill: PAL.sub });
  s.text(155, 240, 'aucune modification de', { size: 10, anchor: 'middle', fill: PAL.sub, italic: true });
  s.text(155, 256, 'l\'infrastructure de l\'institution', { size: 10, anchor: 'middle', fill: PAL.sub, italic: true });
  // lien d'accès reconstitué
  s.rect(360, 100, 190, 200, { fill: '#fdf6e9', stroke: PAL.amber, sw: 2, dash: '6 3' });
  s.text(455, 128, 'LIEN D\'ACCÈS REJOUÉ', { size: 13, weight: '700', anchor: 'middle', fill: '#8a6d1a' });
  s.text(455, 152, 'veth : netem (délai, gigue, perte)', { size: 11, anchor: 'middle', fill: PAL.sub });
  s.text(455, 170, '+ discipline de file testée', { size: 11, anchor: 'middle', fill: PAL.sub });
  s.text(455, 195, 'pfifo_fast · fq_codel · CAKE', { size: 11, weight: '700', anchor: 'middle', fill: PAL.ink });
  s.text(455, 230, 'P1 80/20ms', { size: 10, anchor: 'middle', fill: PAL.sub });
  s.text(455, 246, 'P2 20/100ms', { size: 10, anchor: 'middle', fill: PAL.sub });
  s.text(455, 262, 'P3 5/600ms', { size: 10, anchor: 'middle', fill: PAL.sub });
  s.text(455, 278, 'P4 100/40ms', { size: 10, anchor: 'middle', fill: PAL.sub });
  // serveur de test
  s.rect(640, 120, 220, 170, { fill: '#e9f5ef', stroke: '#2f8f4f', sw: 2 });
  s.text(750, 150, 'SERVEUR DE TEST', { size: 14, weight: '700', anchor: 'middle', fill: '#1d5c36' });
  s.text(750, 175, 'objet 16 Ko (HTTP :8081)', { size: 11, anchor: 'middle', fill: PAL.sub });
  s.text(750, 193, 'puits bulk (TCP :5201)', { size: 11, anchor: 'middle', fill: PAL.sub });
  s.text(750, 218, 'les deux sauts vivent dans', { size: 10, anchor: 'middle', fill: PAL.sub, italic: true });
  s.text(750, 234, 'la même machine (paire veth)', { size: 10, anchor: 'middle', fill: PAL.sub, italic: true });
  // flèches
  s.line(270, 190, 360, 190, { stroke: PAL.blue, sw: 2.5, marker: 'aBlue' });
  s.line(550, 190, 640, 190, { stroke: PAL.amber, sw: 2.5, marker: 'aAmber' });
  s.text(315, 180, 'émission', { size: 11, anchor: 'middle', fill: PAL.blue });
  s.text(595, 180, 'réception', { size: 11, anchor: 'middle', fill: '#8a6d1a' });
  // légende protocole
  s.rect(40, 320, 820, 40, { fill: '#f8f9fb', stroke: PAL.grid, sw: 1 });
  s.text(60, 345, 'Protocole par événement : 30 s repos → 120 s charge (bulk) → 30 s récupération · portes G0–G7 · gel CSV + SHA-256', { size: 11.5, fill: PAL.ink });
  save('figR1-topologie', s);
})();

// ---------------------------------------------------------------- figR2 : deadline 220 (LE résultat)
(function figR2() {
  const final = S.final_run;
  const order = ['pfifo_fast', 'fq_codel', 'cake'];
  const cubic = final.filter(r => r.cc === 'cubic').sort((a, b) => order.indexOf(a.qdisc) - order.indexOf(b.qdisc));
  const bbr = final.filter(r => r.cc === 'bbr').sort((a, b) => order.indexOf(a.qdisc) - order.indexOf(b.qdisc));
  const svg = new Svg(900, 460);
  const m = { l: 60, r: 30, t: 60, b: 60 };
  const x0 = m.l, x1 = 900 - m.r, y0 = m.t, y1 = 400;
  // axes
  svg.line(x0, y0, x0, y1, { stroke: PAL.axis });
  svg.line(x0, y1, x1, y1, { stroke: PAL.axis });
  // échelle Y 0–100
  const y = v => y1 - (v / 100) * (y1 - y0);
  for (let g = 0; g <= 100; g += 25) {
    svg.line(x0, y(g), x1, y(g), { stroke: PAL.grid });
    svg.text(x0 - 10, y(g) + 4, g + ' %', { anchor: 'end', size: 11, fill: PAL.sub });
  }
  // groupes
  const groupW = (x1 - x0) / 3;
  const barW = 42, gap = 12;
  for (let gi = 0; gi < 3; gi++) {
    const gx = x0 + gi * groupW + groupW / 2;
    const cv = cubic[gi], bv = bbr[gi];
    svg.rect(gx - barW - gap / 2, y(+cv.deadline_ok_pct), barW, y(0) - y(+cv.deadline_ok_pct), { fill: PAL.slate });
    svg.rect(gx + gap / 2, y(+bv.deadline_ok_pct), barW, y(0) - y(+bv.deadline_ok_pct), { fill: PAL.amber });
    svg.text(gx - gap / 2, y(+cv.deadline_ok_pct) - 8, cv.deadline_ok_pct + '%', { anchor: 'middle', size: 12, weight: '700', fill: PAL.slate });
    svg.text(gx + barW + gap / 2, y(+bv.deadline_ok_pct) - 8, bv.deadline_ok_pct + '%', { anchor: 'middle', size: 12, weight: '700', fill: '#8a6d1a' });
    svg.text(gx, y1 + 24, order[gi], { anchor: 'middle', size: 13, fill: PAL.ink });
  }
  svg.text(x0 - 42, (y0 + y1) / 2, 'requêtes à l\'heure', { size: 12, fill: PAL.sub, rotate: -90, anchor: 'middle' });
  svg.text(450, 30, 'Respect de l\'échéance (220 ms) sous charge — profil 4G P2', { anchor: 'middle', size: 15, weight: '700', fill: PAL.ink });
  legend(svg, 620, 60, [['CUBIC', PAL.slate], ['BBR', PAL.amber]]);
  save('figR2-deadline', svg);
})();

// ---------------------------------------------------------------- figR3 : QDI par qdisc (données agrégées)
(function figR3() {
  const svg = new Svg(900, 460);
  const m = { l: 60, r: 30, t: 60, b: 60 };
  const x0 = m.l, x1 = 900 - m.r, y0 = m.t, y1 = 400;
  const cells = S.cells.filter(c => c.cell.startsWith('P2|'));
  const qdiscs = ['pfifo_fast', 'fq_codel', 'cake'];
  const maxY = 60;
  const y = v => y1 - (v / maxY) * (y1 - y0);
  svg.line(x0, y0, x0, y1, { stroke: PAL.axis });
  svg.line(x0, y1, x1, y1, { stroke: PAL.axis });
  for (let g = 0; g <= maxY; g += 20) {
    svg.line(x0, y(g), x1, y(g), { stroke: PAL.grid });
    svg.text(x0 - 10, y(g) + 4, g + ' ms', { anchor: 'end', size: 11, fill: PAL.sub });
  }
  const groupW = (x1 - x0) / qdiscs.length;
  qdiscs.forEach((q, gi) => {
    const cell = cells.find(c => c.cell === 'P2|' + q + '|bbr') || cells.find(c => c.cell === 'P2|' + q + '|cubic');
    if (!cell) return;
    const gx = x0 + gi * groupW + groupW / 2;
    svg.rect(gx - 30, y(cell.qdi_med), 60, y(0) - y(cell.qdi_med), { fill: q === 'cake' ? '#2f8f4f' : q === 'fq_codel' ? PAL.blue : PAL.red, op: 0.85 });
    svg.text(gx, y(cell.qdi_med) - 8, cell.qdi_med.toFixed(1) + ' ms', { anchor: 'middle', size: 13, weight: '700', fill: PAL.ink });
    svg.text(gx, y1 + 24, q, { anchor: 'middle', size: 13, fill: PAL.ink });
    svg.text(gx, y1 + 42, 'n=' + cell.n, { anchor: 'middle', size: 10, fill: PAL.sub });
  });
  svg.text(450, 30, 'QDI (p95 − médiane, ms) sous BBR — régularité du lien 4G P2', { anchor: 'middle', size: 15, weight: '700', fill: PAL.ink });
  svg.text(x0 - 42, (y0 + y1) / 2, 'écart p95 − p50', { size: 12, fill: PAL.sub, rotate: -90, anchor: 'middle' });
  save('figR3-qdi', svg);
})();

// ---------------------------------------------------------------- figR4 : paliers tarifaires réels
(function figR4() {
  const svg = new Svg(900, 420);
  const tiers = [
    { n: 'Ye\'low One\n1 Go / 24 h', ar: 1000 },
    { n: 'Net Month\n4,5 Go / 30 j', ar: 5556 },
    { n: 'Net Month 15 Go', ar: 5000 },
    { n: 'Net Month\n100 Go', ar: 2000 },
    { n: 'Fibre IN\n100 Go (FTTH)', ar: 490 },
  ];
  const m = { l: 160, r: 40, t: 70, b: 90 };
  const x0 = m.l, x1 = 900 - m.r, y0 = m.t, y1 = 330;
  const maxY = 6000;
  const y = v => y1 - (v / maxY) * (y1 - y0);
  svg.line(x0, y0, x0, y1, { stroke: PAL.axis });
  svg.line(x0, y1, x1, y1, { stroke: PAL.axis });
  for (let g = 0; g <= 6000; g += 1500) {
    svg.line(x0, y(g), x1, y(g), { stroke: PAL.grid });
    svg.text(x0 - 10, y(g) + 4, g.toLocaleString('fr') + ' Ar', { anchor: 'end', size: 11, fill: PAL.sub });
  }
  const bw = (x1 - x0) / tiers.length;
  tiers.forEach((t, i) => {
    const bx = x0 + i * bw + bw / 2 - 32;
    const col = t.ar > 4000 ? PAL.red : t.ar > 1500 ? PAL.amber : '#2f8f4f';
    svg.rect(bx, y(t.ar), 64, y(0) - y(t.ar), { fill: col, op: 0.85 });
    svg.text(bx + 32, y(t.ar) - 8, t.ar.toLocaleString('fr') + ' Ar', { anchor: 'middle', size: 11.5, weight: '700', fill: PAL.ink });
    const lines = t.n.split('\n');
    lines.forEach((ln, j) => svg.text(bx + 32, y1 + 22 + j * 14, ln, { anchor: 'middle', size: 10.5, fill: PAL.sub }));
  });
  svg.text(450, 36, 'Prix du gigaoctet selon le forfait — Yas 2026 (yas.mg)', { anchor: 'middle', size: 15, weight: '700', fill: PAL.ink });
  svg.text(x0 - 40, (y0 + y1) / 2, 'Ar / Go', { size: 12, fill: PAL.sub, rotate: -90, anchor: 'middle' });
  save('figR4-tarifs', svg);
})();

// ---------------------------------------------------------------- figR5 : pipeline gel + SHA-256
(function figR5() {
  const svg = new Svg(900, 300);
  const steps = [
    { x: 60, t: 'Événement', d: '30/120/30 + portes' },
    { x: 230, t: 'CSV gelé', d: 'aqm_eval.csv (append)' },
    { x: 420, t: 'Manifeste', d: 'SHA-256 par fichier' },
    { x: 620, t: 'Vérification', d: 'cgo verify' },
    { x: 780, t: 'Figures', d: 'SVG régénérés du gel' },
  ];
  const boxW = 130;
  steps.forEach((st, i) => {
    svg.rect(st.x, 100, boxW, 80, { fill: i === 3 ? '#e9f5ef' : '#eef2f7', stroke: i === 3 ? '#2f8f4f' : PAL.blue, sw: 1.8 });
    svg.text(st.x + boxW / 2, 128, st.t, { anchor: 'middle', size: 12.5, weight: '700', fill: PAL.ink });
    svg.text(st.x + boxW / 2, 150, st.d, { anchor: 'middle', size: 10, fill: PAL.sub });
    if (i < steps.length - 1) {
      svg.line(st.x + boxW + 4, 140, steps[i + 1].x - 6, 140, { stroke: PAL.blue, sw: 2, marker: 'aBlue' });
    }
  });
  svg.text(450, 60, 'Chaîne de preuve : chaque nombre publié est vérifiable', { anchor: 'middle', size: 15, weight: '700', fill: PAL.ink });
  svg.text(450, 230, 'Les événements en quarantaine restent comptés dans le journal — rien n\'est masqué.', { anchor: 'middle', size: 11.5, fill: PAL.sub, italic: true });
  save('figR5-gel', svg);
})();

console.log('5 figures réelles écrites dans', OUT);
