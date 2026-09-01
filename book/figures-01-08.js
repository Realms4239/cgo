// Figures 1-8 : schematics (context, topology, architecture).
const fs = require('fs');
const path = require('path');
const { PAL, Svg, chartFrame, legend } = require('./fig-lib');

const OUT = path.join(__dirname, 'figures');
fs.mkdirSync(OUT, { recursive: true });
const save = (name, svg) => fs.writeFileSync(path.join(OUT, name + '.svg'), svg.toString());

// ---------------------------------------------------------------- fig01 : câbles sous-marins
(function fig01() {
  const s = new Svg(900, 560);
  s.arrowDef('aBlue', PAL.blue);
  // ocean
  s.rect(0, 0, 900, 560, { fill: '#f0f6fc' });
  s.text(70, 44, 'Océan Indien occidental', { size: 14, fill: PAL.sub, italic: true });
  // Africa east coast (stylized)
  s.path('M60 80 C130 110 150 200 135 290 C125 350 100 430 60 520 L0 560 L0 60 Z', { fill: '#e8e4d8', stroke: '#c9c2ae', sw: 1.5 });
  s.text(38, 310, 'Afrique du Sud', { size: 13, fill: PAL.sub, rotate: -90, anchor: 'middle' });
  // Madagascar (stylized)
  s.path('M560 150 C600 130 640 150 650 200 C662 260 655 330 635 390 C620 435 585 470 555 450 C525 430 520 370 528 300 C535 240 535 170 560 150 Z',
    { fill: '#dcefdc', stroke: '#79ad79', sw: 2 });
  s.text(590, 300, 'MADAGASCAR', { size: 15, weight: '700', fill: '#2f6b2f', anchor: 'middle', rotate: -78 });
  // Réunion / Maurice
  s.circle(790, 330, 12, { fill: '#e8e4d8', stroke: '#c9c2ae' });
  s.text(790, 360, 'La Réunion', { size: 11.5, anchor: 'middle', fill: PAL.sub });
  s.circle(845, 280, 12, { fill: '#e8e4d8', stroke: '#c9c2ae' });
  s.text(845, 258, 'Maurice', { size: 11.5, anchor: 'middle', fill: PAL.sub });
  // landing points
  const toama = [652, 262];   // Toamasina (east)
  const mahaj = [540, 205];   // Mahajanga (west-north)
  s.circle(...toama, 7, { fill: PAL.teal, stroke: '#fff', sw: 2 });
  s.text(toama[0] + 14, toama[1] + 4, 'Toamasina', { size: 12, fill: PAL.ink, weight: '600' });
  s.circle(...mahaj, 7, { fill: PAL.blue, stroke: '#fff', sw: 2 });
  s.text(mahaj[0] - 12, mahaj[1] - 10, 'Mahajanga', { size: 12, fill: PAL.ink, weight: '600', anchor: 'end' });
  // METISS cable: SA -> Madagascar (Toamasina) -> Réunion -> Maurice
  s.path(`M120 430 C300 500 500 460 ${toama[0]} ${toama[1]}`, { stroke: PAL.teal, sw: 3.5 });
  s.path(`M${toama[0]} ${toama[1]} C710 300 760 335 790 330 C815 315 835 295 845 280`, { stroke: PAL.teal, sw: 3.5 });
  s.circle(120, 430, 7, { fill: PAL.teal, stroke: '#fff', sw: 2 });
  s.text(230, 500, 'METISS (2021, 12 Tbit/s, 3 200 km)', { size: 13, fill: PAL.teal, weight: '700' });
  // 2Africa cable: big loop west -> Mahajanga
  s.path(`M100 100 C260 40 430 60 ${mahaj[0] - 6} ${mahaj[1] - 6}`, { stroke: PAL.blue, sw: 3.5 });
  s.path(`M100 100 C60 160 60 260 90 380`, { stroke: PAL.blue, sw: 3.5, dash: '8 5' });
  s.text(255, 34, '2Africa (180 Tbit/s à l\u2019échelle du système)', { size: 13, fill: PAL.blue, weight: '700' });
  // bottleneck callout
  s.box(600, 480, 280, 58, ['La contrainte n\u2019est plus internationale :', 'elle est au dernier kilomètre radio'], { fill: PAL.amberBg, stroke: PAL.amber, size: 12.5 });
  s.line(640, 480, 610, 400, { stroke: PAL.amber, sw: 1.6, dash: '4 3' });
  legend(s, 70, 540, [['Câble METISS', PAL.teal, 'line'], ['Câble 2Africa', PAL.blue, 'line'], ['Point d\u2019atterrissement', PAL.teal, 'sq']]);
  save('fig01-cables', s);
})();

// ---------------------------------------------------------------- fig02 : décomposition du RTT (3 phases)
(function fig02() {
  const s = new Svg(900, 480);
  const m = { left: 80, top: 50, right: 860, bottom: 390 };
  const { sx, sy } = chartFrame(s, m, [0, 120], [0, 700], {
    xLabel: 'Temps (s)', yLabel: 'RTT (ms)', ny: 7,
  });
  // phases
  const phase = (x0, x1, fill, label) => {
    s.rect(sx(x0), m.top, sx(x1) - sx(x0), m.bottom - m.top, { fill, op: 0.5 });
    s.text((sx(x0) + sx(x1)) / 2, m.top + 22, label, { anchor: 'middle', size: 13, weight: '700', fill: PAL.sub });
  };
  phase(0, 40, '#f0f8f0', 'Phase normale');
  phase(40, 85, '#fdf0ec', 'Bufferbloat');
  phase(85, 120, '#eef4fb', 'Récupération');
  // propagation delay baseline
  s.line(sx(0), sy(80), sx(120), sy(80), { stroke: PAL.slate, sw: 2, dash: '7 5' });
  s.text(sx(3), sy(80) - 10, 'Délai de propagation (fixe) : 80 ms', { size: 12, fill: PAL.slate, weight: '600' });
  // RTT curve : flat, then climbs to ~620, then drops back
  const pts = [];
  for (let t = 0; t <= 120; t += 1) {
    let v;
    if (t < 40) v = 80 + 12 * Math.sin(t * 0.9) * Math.exp(-((t % 13)) / 9) + 6 * ((t * 7919) % 10) / 10;
    else if (t < 60) v = 80 + (t - 40) * 26 + 10 * Math.sin(t);
    else if (t < 85) v = 600 + 25 * Math.sin(t * 0.7);
    else v = Math.max(82, 610 - (t - 85) * 38) + 8 * Math.sin(t);
    pts.push([sx(t), sy(Math.min(v, 660))]);
  }
  s.path('M' + pts.map(p => p.map(v => v.toFixed(1)).join(' ')).join(' L '), { stroke: PAL.red, sw: 2.6 });
  // queue delay area annotation
  s.arrowDef('a2', PAL.ink);
  s.line(sx(72), sy(80), sx(72), sy(600), { stroke: PAL.ink, sw: 1.4, marker: 'a2' });
  s.line(sx(72), sy(600), sx(72), sy(80), { stroke: PAL.ink, sw: 1.4, marker: 'a2' });
  s.text(sx(73.5), sy(340), 'Délai de file (variable)', { size: 12.5, weight: '600' });
  s.text(sx(73.5), sy(300), 'jusqu\u2019à + 520 ms', { size: 12, fill: PAL.red });
  // onset marker
  s.line(sx(40), m.top + 34, sx(40), m.bottom, { stroke: PAL.amber, sw: 2, dash: '5 4' });
  s.text(sx(40), m.top + 46, 'Début de congestion (onset)', { size: 11.5, fill: PAL.amber, weight: '700', anchor: 'middle' });
  legend(s, 90, 470, [['RTT mesuré', PAL.red, 'line'], ['Délai de propagation', PAL.slate, 'dash']]);
  save('fig02-rtt-phases', s);
})();

// ---------------------------------------------------------------- fig03 : topologie du banc
(function fig03() {
  const s = new Svg(940, 560);
  s.arrowDef('aB', PAL.blue); s.arrowDef('aT', PAL.teal); s.arrowDef('aP', PAL.purple); s.arrowDef('aA', PAL.amber);
  s.text(470, 34, 'Machine virtuelle Ubuntu 24.04 LTS (noyau 6.8, 2 vCPU, 4 Go RAM)', { anchor: 'middle', size: 14, weight: '700', fill: PAL.sub });
  s.rect(20, 50, 900, 420, { rx: 14, stroke: PAL.axis, sw: 1.6, dash: '8 6' });
  // ns-A
  s.rect(50, 90, 260, 340, { rx: 12, fill: '#f7fafd', stroke: PAL.blue, sw: 2 });
  s.text(180, 118, 'ns-A (émetteur)', { anchor: 'middle', size: 14.5, weight: '700', fill: PAL.blue });
  s.box(75, 140, 210, 56, ['Générateurs de trafic', 'BBR / CUBIC (rythme noyau)'], { size: 12 });
  s.box(75, 210, 210, 56, ['cgo-linux (binaire Go, 13 Mo)', 'oracle + détecteurs + contrôle'], { fill: PAL.tealBg, stroke: PAL.teal, size: 12 });
  s.box(75, 280, 210, 44, ['client iperf3 (réconciliation)'], { fill: PAL.grayBg, stroke: PAL.slate, size: 12 });
  s.box(75, 338, 210, 44, ['ping -D (5 Hz, horodatage noyau)'], { fill: PAL.purpleBg, stroke: PAL.purple, size: 12 });
  // ns-B
  s.rect(630, 90, 260, 340, { rx: 12, fill: '#f7fafd', stroke: PAL.blue, sw: 2 });
  s.text(760, 118, 'ns-B (récepteur)', { anchor: 'middle', size: 14.5, weight: '700', fill: PAL.blue });
  s.text(760, 140, '10.200.1.2 (portée lien)', { anchor: 'middle', size: 12, fill: PAL.sub, mono: true });
  s.box(655, 160, 210, 50, ['serveur iperf3'], { fill: PAL.grayBg, stroke: PAL.slate, size: 12 });
  s.box(655, 226, 210, 50, ['tcpdump (captures pcap)'], { fill: PAL.grayBg, stroke: PAL.slate, size: 12 });
  s.box(655, 292, 210, 50, ['répondeur ICMP'], { fill: PAL.purpleBg, stroke: PAL.purple, size: 12 });
  // bottleneck pipe
  const bx = 330, bw = 280, by = 180;
  s.rect(bx, by, bw, 120, { rx: 10, fill: PAL.amberBg, stroke: PAL.amber, sw: 2.2 });
  s.text(bx + bw / 2, by - 14, 'Goulot : thesis-veth0 (paire veth, offloads désactivés)', { anchor: 'middle', size: 12.5, weight: '700', fill: PAL.amber });
  s.box(bx + 16, by + 18, 74, 84, ['netem', 'délai', 'gigue', 'perte'], { fill: '#fff', stroke: PAL.amber, size: 11, weight: '600', lh: 15 });
  s.box(bx + 103, by + 18, 74, 84, ['TBF', '50 Mbit/s', 'seau de', 'jetons'], { fill: '#fff', stroke: PAL.amber, size: 11, weight: '600', lh: 15 });
  s.box(bx + 190, by + 18, 74, 84, ['fq_codel', 'cible 5 ms', 'AQM', 'par flux'], { fill: '#fff', stroke: PAL.amber, size: 11, weight: '600', lh: 15 });
  // data flow arrows
  s.line(310, 240, bx, 240, { stroke: PAL.blue, sw: 3, marker: 'aB' });
  s.line(bx + bw, 240, 630, 240, { stroke: PAL.blue, sw: 3, marker: 'aB' });
  s.text(470, 330, 'trafic de charge', { anchor: 'middle', size: 12, fill: PAL.blue, weight: '600' });
  // observation channels below
  s.box(50, 486, 400, 56, ['Oracle noyau : tc -s class show à 10 Hz', 'compteurs exacts de la file (vérité terrain)'], { fill: PAL.tealBg, stroke: PAL.teal, size: 12 });
  s.box(490, 486, 400, 56, ['Canaux indépendants : ping -D (5 Hz),', 'TCP_INFO (1 Hz), tcpdump (preuve brute)'], { fill: PAL.purpleBg, stroke: PAL.purple, size: 12 });
  s.line(250, 486, 440, 302, { stroke: PAL.teal, sw: 1.8, dash: '5 4', marker: 'aT' });
  s.line(690, 486, 500, 302, { stroke: PAL.purple, sw: 1.8, dash: '5 4', marker: 'aP' });
  save('fig03-banc', s);
})();

// ---------------------------------------------------------------- fig04 : topologie 4G réelle
(function fig04() {
  const s = new Svg(940, 400);
  s.arrowDef('aB4', PAL.blue);
  const y = 150;
  s.box(40, y, 190, 110, ['Banc de mesure', 'VM cgo', '(portable sur batterie)'], { size: 12.5, lh: 17 });
  s.box(300, y, 190, 110, ['Modem 4G homologué', 'ou téléphone en partage', 'SIM prépayée (~30 000 Ar)'], { fill: PAL.tealBg, stroke: PAL.teal, size: 12.5, lh: 17 });
  // antenna
  s.line(600, y + 110, 600, y - 20, { stroke: PAL.slate, sw: 3 });
  s.path(`M580 ${y - 20} L620 ${y - 20} L600 ${y + 15} Z`, { fill: PAL.slate });
  s.path(`M570 ${y - 40} C590 ${y - 55} 610 ${y - 55} 630 ${y - 40}`, { stroke: PAL.amber, sw: 2 });
  s.path(`M560 ${y - 28} C588 ${y - 50} 612 ${y - 50} 640 ${y - 28}`, { stroke: PAL.amber, sw: 2, op: 0.55 });
  s.text(600, y + 140, 'Antenne opérateur', { anchor: 'middle', size: 12.5, weight: '600' });
  s.text(600, y + 158, '(Telma, Orange, Airtel, Yas)', { anchor: 'middle', size: 11.5, fill: PAL.sub });
  // internet cloud
  s.path('M720 190 C715 160 745 140 775 148 C783 122 823 118 840 140 C868 130 898 150 890 178 C908 196 894 226 868 226 L745 226 C720 226 710 208 720 190 Z',
    { fill: '#eef4fb', stroke: PAL.blue, sw: 2 });
  s.text(805, 192, 'Internet', { anchor: 'middle', size: 14, weight: '700', fill: PAL.blue });
  // links
  s.line(230, y + 55, 300, y + 55, { stroke: PAL.blue, sw: 2.6, marker: 'aB4' });
  s.text(265, y + 40, 'USB / Wi-Fi', { anchor: 'middle', size: 11, fill: PAL.sub });
  s.path(`M490 ${y + 45} C530 ${y + 20} 560 ${y + 5} 592 ${y - 5}`, { stroke: PAL.amber, sw: 2.4, dash: '6 4' });
  s.text(535, y + 5, 'liaison radio 4G', { anchor: 'middle', size: 11.5, fill: PAL.amber, weight: '600' });
  s.line(620, y + 30, 725, 195, { stroke: PAL.slate, sw: 2.2, dash: '2 4' });
  s.text(680, y + 62, 'réseau cœur + NAT', { anchor: 'middle', size: 11, fill: PAL.sub, rotate: 15 });
  // annotations
  s.box(40, 300, 850, 62, ['Aucun équipement actif interposé : la seule source de gigue et de perte est le lien d\u2019accès lui-même.',
    'Redondance terrain : deux opérateurs, deux chemins matériels distincts. Trafic synthétique uniquement (loi n° 2014-038).'],
    { fill: PAL.grayBg, stroke: PAL.slate, size: 12, weight: 'normal', lh: 20 });
  save('fig04-4g', s);
})();

// ---------------------------------------------------------------- fig05 : profil diurne 4G
(function fig05() {
  const s = new Svg(900, 460);
  const m = { left: 80, top: 46, right: 860, bottom: 370 };
  const { sx, sy } = chartFrame(s, m, [0, 24], [0, 24], {
    xLabel: 'Heure de la journée', yLabel: 'Débit descendant mesuré (Mbit/s)', nx: 12, ny: 6,
    xFmt: t => `${t}h`,
  });
  // peak windows
  s.rect(sx(12), m.top, sx(14) - sx(12), m.bottom - m.top, { fill: PAL.redBg, op: 0.65 });
  s.rect(sx(18), m.top, sx(21) - sx(18), m.bottom - m.top, { fill: PAL.redBg, op: 0.65 });
  s.text((sx(12) + sx(14)) / 2, m.top + 18, 'pointe midi', { anchor: 'middle', size: 11.5, fill: PAL.red, weight: '600' });
  s.text((sx(18) + sx(21)) / 2, m.top + 18, 'pointe du soir', { anchor: 'middle', size: 11.5, fill: PAL.red, weight: '600' });
  // synthetic but realistic diurnal profile: high at night, collapse at peaks
  const base = t => {
    const dipNoon = 14 * Math.exp(-Math.pow(t - 13, 2) / 1.8);
    const dipEve = 16.5 * Math.exp(-Math.pow(t - 19.5, 2) / 2.6);
    return Math.max(2.2, 19.5 - dipNoon - dipEve - 1.5 * Math.exp(-Math.pow(t - 8.5, 2) / 3));
  };
  // envelope p10-p90
  let up = [], dn = [];
  for (let t = 0; t <= 24; t += 0.25) {
    const b = base(t);
    const spread = 2.2 + 1.6 * Math.sin(t / 2.7 + 1);
    up.push([sx(t), sy(Math.min(23, b + spread))]);
    dn.push([sx(t), sy(Math.max(1.2, b - spread * 0.8))]);
  }
  s.path('M' + up.map(p => p.map(v => v.toFixed(1)).join(' ')).join(' L ') + ' L ' + dn.reverse().map(p => p.map(v => v.toFixed(1)).join(' ')).join(' L ') + ' Z',
    { fill: PAL.blue, op: 0.14 });
  // median line
  const med = [];
  for (let t = 0; t <= 24; t += 0.25) med.push([sx(t), sy(base(t))]);
  s.path('M' + med.map(p => p.map(v => v.toFixed(1)).join(' ')).join(' L '), { stroke: PAL.blue, sw: 2.8 });
  // nominal line
  s.line(sx(0), sy(20), sx(24), sy(20), { stroke: PAL.slate, sw: 1.8, dash: '7 5' });
  s.text(sx(0.4), sy(20) - 9, 'débit nominal hors charge (~20 Mbit/s)', { size: 11.5, fill: PAL.slate });
  // collapse annotation
  s.arrowDef('a5', PAL.red);
  s.line(sx(19.5), sy(19.6), sx(19.5), sy(3.6), { stroke: PAL.red, sw: 1.8, marker: 'a5' });
  s.text(sx(19.9), sy(11), 'effondrement de 50 à 80 %', { size: 12, fill: PAL.red, weight: '700' });
  legend(s, 90, 442, [['Médiane des relevés', PAL.blue, 'line'], ['Enveloppe p10\u2013p90', '#bfd3f7', 'sq'], ['Heures de pointe', '#f6c9c9', 'sq']]);
  save('fig05-diurne', s);
})();

// ---------------------------------------------------------------- fig06 : ECDF RTT 4G
(function fig06() {
  const s = new Svg(900, 460);
  const m = { left: 80, top: 46, right: 860, bottom: 370 };
  const { sx, sy } = chartFrame(s, m, [0, 400], [0, 1], {
    xLabel: 'RTT (ms)', yLabel: 'Fraction cumulée', ny: 5, nx: 8,
    yFmt: t => t.toFixed(2),
  });
  // ecdf curves: off-peak (tight), pre-onset (widened tail), peak (shifted)
  const logi = (x, mu, s0, tail = 0) => {
    let v = 1 / (1 + Math.exp(-(x - mu) / s0));
    if (tail > 0) v = Math.min(1, v * (1 - tail) + tail * (1 / (1 + Math.exp(-(x - mu * 2.6) / (s0 * 3)))));
    return v;
  };
  const draw = (mu, s0, tail, color, dash) => {
    const pts = [];
    for (let x = 20; x <= 400; x += 4) pts.push([sx(x), sy(logi(x, mu, s0, tail))]);
    s.path('M' + pts.map(p => p.map(v => v.toFixed(1)).join(' ')).join(' L '), { stroke: color, sw: 2.6, dash });
  };
  draw(72, 9, 0, PAL.teal);
  draw(80, 14, 0.30, PAL.amber);
  draw(180, 34, 0.12, PAL.red, '7 5');
  // annotation of the tail widening
  s.arrowDef('a6', PAL.amber);
  s.path(`M ${sx(255)} ${sy(0.86)} C ${sx(285)} ${sy(0.80)} ${sx(300)} ${sy(0.90)} ${sx(292)} ${sy(0.935)}`, { stroke: PAL.amber, sw: 1.6, marker: 'a6' });
  s.text(sx(250), sy(0.82), 'élargissement de la queue de distribution', { size: 12, fill: PAL.amber, weight: '600', anchor: 'end' });
  s.text(sx(250), sy(0.765), 'avant tout déplacement de la médiane', { size: 12, fill: PAL.amber, anchor: 'end' });
  // median markers
  s.line(sx(72), sy(0.5), sx(180), sy(0.5), { stroke: PAL.grid, sw: 1.2 });
  s.circle(sx(72), sy(0.5), 4, { fill: PAL.teal });
  s.circle(sx(80), sy(0.5), 4, { fill: PAL.amber });
  s.circle(sx(180), sy(0.5), 4, { fill: PAL.red });
  legend(s, 90, 442, [['Creux (nuit)', PAL.teal, 'line'], ['Précurseur (avant pointe)', PAL.amber, 'line'], ['Heure de pointe', PAL.red, 'dash']]);
  save('fig06-ecdf-4g', s);
})();

// ---------------------------------------------------------------- fig07 : architecture CGO (couches)
(function fig07() {
  const s = new Svg(940, 640);
  s.arrowDef('aA7', PAL.slate);
  s.text(470, 36, 'Observatoire CGO : binaire Go statique unique (~13 Mo)', { anchor: 'middle', size: 15, weight: '700' });
  s.rect(30, 54, 880, 520, { rx: 14, stroke: PAL.axis, sw: 1.8 });
  const rows = [
    { y: 80, fill: PAL.blueBg, stroke: PAL.blue, title: 'Présentation et API',
      cells: [['Console web', 'xterm.js + go:embed'], ['Explorateur', 'd\u2019archive'], ['Vérifieur', 'SHA-256'], ['API HTTP', 'config / preuves']] },
    { y: 180, fill: PAL.tealBg, stroke: PAL.teal, title: 'Contrôle',
      cells: [['Contrôleur adaptatif', 'pkg/adaptive'], ['Bandit ARMS', 'sélection de bras'], ['Règles de politique', 'PolicyRule'], ['Journal des décisions', 'decision_log']] },
    { y: 280, fill: PAL.purpleBg, stroke: PAL.purple, title: 'Observabilité et analyse',
      cells: [['Oracle noyau 10 Hz', 'pkg/oracle'], ['Collecte TCP_INFO', 'pkg/metrics'], ['5 détecteurs', 'pkg/detect'], ['Statistiques', 'pkg/stats']] },
    { y: 380, fill: PAL.amberBg, stroke: PAL.amber, title: 'Plan de données (noyau Linux)',
      cells: [['netlink qdisc', 'pkg/qdisc'], ['Façonneur', 'pkg/shaper'], ['netem + TBF', 'goulot'], ['fq_codel', 'AQM']] },
    { y: 480, fill: PAL.grayBg, stroke: PAL.slate, title: 'Expérimentation et intégrité',
      cells: [['Moteur de campagne', 'pkg/campaign'], ['Injection de fautes', 'pkg/fault'], ['Bac à sable', 'pkg/testbed'], ['Manifeste chaîné', 'SHA-256 + Zenodo']] },
  ];
  for (const r of rows) {
    s.text(58, r.y + 14, r.title, { size: 13, weight: '700', fill: r.stroke });
    r.cells.forEach((c, i) => {
      const x = 58 + i * 208;
      s.box(x, r.y + 24, 190, 56, c, { fill: r.fill, stroke: r.stroke, size: 11.8, lh: 16 });
    });
    if (r.y < 480) {
      s.line(470, r.y + 80, 470, r.y + 100, { stroke: PAL.slate, sw: 2, marker: 'aA7' });
      s.line(470, r.y + 100, 470, r.y + 80, { stroke: PAL.slate, sw: 2, marker: 'aA7' });
    }
  }
  s.text(470, 606, 'Chaque couche communique avec la suivante par des interfaces internes ; aucune dépendance externe à l\u2019exécution.',
    { anchor: 'middle', size: 12, fill: PAL.sub, italic: true });
  save('fig07-archi', s);
})();

// ---------------------------------------------------------------- fig08 : décomposition RTT avant/après fq_codel
(function fig08() {
  const s = new Svg(900, 470);
  const m = { left: 80, top: 60, right: 860, bottom: 380 };
  const { sx, sy } = chartFrame(s, m, [0, 60], [0, 600], {
    xLabel: 'Temps depuis le début de la charge (s)', yLabel: 'RTT (ms)', ny: 6,
  });
  // propagation
  s.rect(sx(0), sy(80), sx(60) - sx(0), sy(0) - sy(80), { fill: '#e9edf3', op: 0.9 });
  s.text(sx(29), sy(38), 'délai de propagation : 80 ms (fixe)', { anchor: 'middle', size: 12, fill: PAL.slate, weight: '600' });
  // pfifo curve
  const pf = [], fq = [];
  for (let t = 0; t <= 60; t += 0.5) {
    const ramp = Math.min(1, t / 12);
    pf.push([sx(t), sy(80 + 440 * ramp + 18 * Math.sin(t * 1.3))]);
    fq.push([sx(t), sy(80 + Math.min(38, 60 * ramp) + 7 * Math.sin(t * 1.7))]);
  }
  s.path('M' + pf.map(p => p.map(v => v.toFixed(1)).join(' ')).join(' L '), { stroke: PAL.red, sw: 2.6 });
  s.path('M' + fq.map(p => p.map(v => v.toFixed(1)).join(' ')).join(' L '), { stroke: PAL.teal, sw: 2.6 });
  s.text(sx(43), sy(555), 'pfifo_fast : file profonde, bufferbloat', { size: 12.5, fill: PAL.red, weight: '700' });
  s.text(sx(43), sy(175), 'fq_codel : délai de file contenu sous ~40 ms', { size: 12.5, fill: PAL.teal, weight: '700' });
  // queue delay arrows
  s.arrowDef('a8r', PAL.red); s.arrowDef('a8t', PAL.teal);
  s.line(sx(56), sy(85), sx(56), sy(510), { stroke: PAL.red, sw: 1.6, marker: 'a8r' });
  s.text(sx(56.7), sy(300), '+440 ms', { size: 12, fill: PAL.red, weight: '600' });
  s.line(sx(30), sy(85), sx(30), sy(132), { stroke: PAL.teal, sw: 1.6, marker: 'a8t' });
  s.text(sx(30.8), sy(122), '+38 ms', { size: 12, fill: PAL.teal, weight: '600' });
  s.text(470, 34, 'Décomposition : RTT = propagation (fixe) + délai de file (variable, contrôlé par l\u2019AQM)', { anchor: 'middle', size: 13.5, weight: '700', fill: PAL.sub });
  legend(s, 90, 452, [['pfifo_fast (sans AQM)', PAL.red, 'line'], ['fq_codel (AQM par flux)', PAL.teal, 'line'], ['Délai de propagation', '#e9edf3', 'sq']]);
  save('fig08-decomposition', s);
})();

console.log('figures 01-08 written to', OUT);
