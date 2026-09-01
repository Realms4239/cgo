// Generate labeled placeholder SVGs for interface screenshots that do not
// exist yet (CGO console captures). svg-to-png.js renders them to PNG.
const fs = require('fs');
const path = require('path');
const { PAL, FONT, Svg } = require('./fig-lib');

const OUT = path.join(__dirname, 'figures');
const W = 1200, H = 640;

const items = [
  ['fig17-console', 'Capture de la console CGO', 'Parcours de l\u2019observatoire, oracle 10 Hz, indicateurs SHA-256'],
  ['fig18-tableau-de-bord', 'Tableau de bord CGO', '\u00c9tat du goulot, bande passante instantan\u00e9e, alertes'],
  ['fig19-replay', 'Relecture d\u2019une s\u00e9quence E1', 'S\u00e9rie de RTT, fen\u00eatre de d\u00e9tection, d\u00e9lai confirm\u00e9'],
  ['fig20-audit', 'V\u00e9rifieur d\u2019int\u00e9grit\u00e9', 'Sommes SHA-256 courantes vs manifeste'],
  ['fig21-preuves', 'Explorateur d\u2019archive', 'Paquet de preuves d\u2019un \u00e9v\u00e9nement (tcpdump, oracle, d\u00e9cisions)'],
  ['fig22-decisions', 'Journal des d\u00e9cisions', 'Condition, raison et r\u00e9sultat horodat\u00e9s'],
];

function placeholderSvg(label, sub) {
  const s = new Svg(W, H);
  s.rect(0, 0, W, H, { fill: '#f1f5f9' });
  s.rect(24, 24, W - 48, H - 48, { fill: '#f8fafc', stroke: '#94a3b8', sw: 2, dash: '10 8' });
  s.circle(W / 2, H / 2 - 70, 54, { fill: '#e2e8f0', stroke: PAL.slate, sw: 3 });
  s.text(W / 2, H / 2 + 30, 'CAPTURE D\u2019\u00c9CRAN \u00c0 VENIR', { size: 30, weight: 'bold', fill: PAL.ink, anchor: 'middle' });
  s.text(W / 2, H / 2 + 80, label, { size: 24, fill: PAL.ink, anchor: 'middle' });
  s.text(W / 2, H / 2 + 122, sub, { size: 17, fill: PAL.slate, anchor: 'middle' });
  s.text(W / 2, H / 2 + 180, 'Remplacer par la capture r\u00e9elle', { size: 14, fill: PAL.slate, anchor: 'middle' });
  return s.toString();
}

fs.mkdirSync(OUT, { recursive: true });
for (const [name, label, sub] of items) {
  const f = path.join(OUT, name + '.svg');
  fs.writeFileSync(f, placeholderSvg(label, sub));
  console.log('placeholder', name);
}
