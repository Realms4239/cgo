// Guide d'interface Meteolink — build le DOCX (français, zéro-connaissance).
// Usage : node build-guide-docx.js   (nécessite ../shots/guide/*.png)
const { Document, Packer, Paragraph, TextRun, HeadingLevel, ImageRun,
  Table, TableRow, TableCell, WidthType, ShadingType, PageBreak } = require('docx')
const fs = require('fs')
const path = require('path')

const SHOTS = path.resolve('../shots/guide')
const OUT = path.resolve('../../guide-interface-meteolink.docx')

function pngSize(file) {
  const b = fs.readFileSync(file)
  return { w: b.readUInt32BE(16), h: b.readUInt32BE(20) }
}
const MAXW = 600
function fig(name, caption) {
  const file = path.join(SHOTS, name + '.png')
  const { w, h } = pngSize(file)
  const W = Math.min(MAXW, w)
  const H = Math.round((h * W) / w)
  return [
    new Paragraph({
      children: [new ImageRun({ data: fs.readFileSync(file), transformation: { width: W, height: H }, type: 'png' })],
      spacing: { before: 160, after: 80 },
    }),
    new Paragraph({
      children: [new TextRun({ text: caption, italics: true, size: 18, color: '555555' })],
      spacing: { after: 200 },
    }),
  ]
}
const T = (text, o = {}) => new TextRun(Object.assign({ text, size: 22 }, o))
const B = (text, o = {}) => new TextRun(Object.assign({ text, bold: true, size: 22 }, o))
const I = (text, o = {}) => new TextRun(Object.assign({ text, italics: true, size: 22 }, o))
const M = (text, o = {}) => new TextRun(Object.assign({ text, font: 'Consolas', size: 20 }, o))
const P = (...runs) => new Paragraph({ children: runs, spacing: { after: 140 } })
const H1 = (t) => new Paragraph({ heading: HeadingLevel.HEADING_1, children: [new TextRun({ text: t, size: 32, bold: true, color: '111111' })], spacing: { before: 0, after: 200 } })
const H2 = (t) => new Paragraph({ heading: HeadingLevel.HEADING_2, children: [new TextRun({ text: t, size: 26, bold: true, color: '0E6E8C' })], spacing: { before: 240, after: 140 } })
const BR = () => new Paragraph({ children: [new PageBreak()] })
function legende(items) {
  // items: ["texte du repère", ...] — numérotés 1..n comme les pastilles
  return items.map((t, i) => new Paragraph({
    children: [B(`Repère ${i + 1} — `), T(t)],
    spacing: { after: 60 },
  }))
}
function table(head, rows, widths) {
  const tot = widths.reduce((a, b) => a + b, 0)
  const cell = (t, bold, shade) => new TableCell({
    width: { size: 0, type: WidthType.AUTO },
    shading: shade ? { type: ShadingType.CLEAR, fill: shade } : undefined,
    children: [new Paragraph({ children: [new TextRun({ text: t, size: 20, bold: !!bold, color: shade ? 'F2F2F4' : undefined })] })],
  })
  return new Table({
    columnWidths: widths,
    width: { size: tot, type: WidthType.DXA },
    rows: [
      new TableRow({ children: head.map(h => cell(h, true, '26262A')) }),
      ...rows.map(r => new TableRow({ children: r.map(c => cell(c, false)) })),
    ],
  })
}
const children = []
const push = (...xs) => children.push(...xs)

// ---------- PAGE DE GARDE ----------
push(
  new Paragraph({ spacing: { before: 2400 }, children: [] }),
  new Paragraph({ children: [new TextRun({ text: 'Meteolink', size: 72, bold: true })], spacing: { after: 100 } }),
  new Paragraph({ children: [new TextRun({ text: "Guide de l'interface — lisez l'écran sans connaître le réseau", size: 30, italics: true, color: '0E6E8C' })], spacing: { after: 300 } }),
  P(T("Ce guide explique "), B("chaque vue, chaque chiffre, chaque bouton"), T(", avec une capture numérotée à chaque étape. Il est écrit pour quelqu'un qui n'a jamais vu l'outil et ne connaît aucun mot du réseau : chaque terme est traduit en français simple dès sa première apparition.")),
  P(B("Version 1.2.3"), T(" — captures du 7 septembre 2026 — public : technicien DSI, jury, directeur.")),
  BR(),
)
// ---------- SOMMAIRE ----------
push(
  H1('Sommaire'),
  ...[
    '0. Avant de commencer — la seule règle à retenir',
    '1. Le vocabulaire en 2 minutes — 8 mots, zéro jargon',
    '2. Campagne — lancer une mesure, auditer un lien réel',
    '3. Tableau live — regarder la mesure respirer',
    '4. Résultats — lire le classement comme un podium',
    '5. Intégrité — la preuve que les chiffres sont vrais',
    '6. Interprétation — la phrase qui dit quoi faire',
    '7. Kit TUI — piloter la machine sans taper de commande',
    '8. Si ça coince — 5 pannes, 5 remèdes',
    'Glossaire — 12 mots à garder sous la main',
  ].map(t => P(B(t.split(' — ')[0] + ' — '), T(t.split(' — ').slice(1).join(' — ')))),
  BR(),
)
// ---------- CH 0 ----------
push(
  H1('0. Avant de commencer'),
  P(B("Meteolink mesure la qualité des liens internet de Météo Madagascar"), T(" (fibre, 4G, VSAT satellite) et compare les réglages qui les rendent plus ou moins réactifs. Quatre vues, un rail à gauche, toujours la même logique.")),
  P(B("La seule règle : "), I("un chiffre affiché vient soit d'une mesure en cours (bandeau coloré en haut = ça mesure MAINTENANT), soit d'une archive gelée et signée (tout le reste)."), T(" Le gelé ne ment jamais : chaque ligne est signée (SHA-256, vue Intégrité).")),
  P(T("Astuce valable partout : "), B("cliquez une ligne du classement pour l'agrandir — le reste s'estompe ; recliquez pour revenir.")),
  BR(),
)
// ---------- CH 1 ----------
push(
  H1('1. Le vocabulaire en 2 minutes'),
  P(T("Huit mots. Avec une image de la vie courante à chaque fois — ensuite, tout l'écran se lit tout seul.")),
  table(['Mot affiché', 'Image', 'En une phrase'],
    [
      ['Campagne', 'Une série de pesées', 'On mesure chaque réglage à tour de rôle, automatiquement.'],
      ['Cellule', 'Une pesée', 'Un trio profil × file × congestion, mesuré 1 à 3 fois.'],
      ['Profil (P1–P4)', 'Le terrain', 'P1 = fibre, P2 = ADSL, P3 = satellite VSAT, P4 = 4G : le décor imposé.'],
      ['File / AQM', 'Le caissier', 'pfifo = aucun ordre ; fq_codel et cake = files intelligentes qui protègent le trafic urgent.'],
      ['Congestion (BBR/cubic)', 'La conduite', 'La façon dont l’ordinateur envoie : BBR sonde en douceur, cubic remplit les files.'],
      ['Échéance', 'Le minuteur', 'Le délai cible (ex. 220 ms) : au-delà, la mesure est ratée.'],
      ['Quarantaine', 'Le contrôle qualité', 'Les mesures incohérentes sont écartées AVANT tout calcul — et montrées, pas cachées.'],
      ['Score LIEN /100', 'La note', 'La moyenne de 5 critères ramenés à 100 : une seule note pour comparer.'],
    ], [2200, 2200, 5000]),
  BR(),
)
// ---------- CH 2 ----------
push(
  H1('2. Campagne — lancer une mesure'),
  P(B("C'est ici que tout commence :"), T(" on choisit ce qu'on mesure (profils), on appuie sur "), B("DÉMARRER"), T(", et on regarde la matrice se remplir. La vue Audit (même écran, à droite) mesure un lien réel en 30 secondes.")),
  ...fig('guide-campagne', 'Figure 1 — Le cockpit Campagne au repos : cocher, régler, démarrer.'),
  ...legende([
    'Profils P1–P4 — que les quatre profils du référentiel, jamais d’autres. Cochez ceux à mesurer.',
    'Échéance (deadline) — le minuteur en millisecondes : la cible à ne pas dépasser.',
    'DÉMARRER — lance la campagne (double confirmation : il faut cliquer deux fois, pas d’accident).',
    'LANCER — ouvre le panneau Audit du lien réel (figure 2).',
    'HISTORIQUE — rouvre tous les audits gelés (figure 3).',
  ]),
  H2('Le tableau Matrice — état par cellule'),
  P(T("Chaque ligne = une cellule prévue (ex. "), B("P2 · fq_codel × bbr"), T("). La colonne "), B("état"), T(" dit "), B("attente"), T(", "), B("en cours"), T(" ou "), B("terminée"), T(", avec le compteur "), I("0/3, 1/3…"), T(" (répétitions faites). Au repos, tout est en attente : c'est normal.")),
  H2('Les portes G0–G7 — le contrôle qualité en direct'),
  P(T("Huit gardiens valident chaque mesure pendant qu'elle se fait : "), B("PASS"), T(" vert = ça passe, "), B("FAIL"), T(" rouge = ça casse, "), B("—"), T(" = pas encore jugé. Exemples : "), I("G0 cible joignable, G3 latence plausible, G6 baseline stable."), T(" Une mesure qui échoue une porte part en quarantaine (vue Intégrité) au lieu de fausser les moyennes.")),
  H2('Auditer un lien réel (30 secondes, sans droits admin)'),
  ...fig('guide-audit', 'Figure 2 — Le panneau Audit : un preset ou un formulaire, puis un bouton.'),
  ...legende([
    'Presets — trois boutons prêts : fibre du siège, 4G heure creuse, 4G heure de pointe. Un clic remplit tout.',
    'Site — le nom du lieu mesuré (ex. « Département X ») : il voyagera avec la mesure.',
    'LANCER AUDIT — démarre les 30 secondes de sondes (double confirmation).',
  ]),
  P(B("Règle d'or : "), T("un audit ne touche à rien — il envoie de petites sondes et écoute. Aucun droit admin, aucun risque.")),
  ...fig('guide-audit-modal', 'Figure 3 — L’historique : chaque audit gelé, son rapprochement, son bouton Profil.'),
  ...legende([
    'Une ligne = un audit gelé : site, type de lien, latences à vide et en charge, note, et le profil de labo le plus proche (« P1, RTT 0 vs 20 ms »).',
    'ACTUALISER — recharge la liste après un nouvel audit.',
  ]),
  P(B("Le bouton "), B("→ PROFIL"), T(" transforme un audit réel en profil rejouable sur le banc : le terrain entre au labo en un clic.")),
  BR(),
)
// ---------- CH 3 ----------
push(
  H1('3. Tableau live — regarder la mesure respirer'),
  P(B("Pendant une campagne, cet écran montre les chiffres NAÎTRE :"), T(" la courbe monte en direct, les cartes clignotent au rythme des sondes. Quand rien ne tourne, il affiche des tirets honnêtes — jamais de fausses données.")),
  ...fig('guide-live-working', 'Figure 4 — Le live en pleine charge : tout ce qui bouge est mesuré à l’instant.'),
  ...legende([
    'Bandeau d’état — ce qui se passe MAINTENANT (ici : CHARGE, bulk actif, P95 partiel). Orange = ça mesure.',
    'Courbe héro — la réactivité des petits objets, seconde par seconde. La zone colorée = la phase de charge.',
    'Cartes métriques — dix compteurs. Chacun se lit : nom, valeur, mini-courbe. Détail au chapitre « Lire une carte ».',
    'Guide de lecture — le bouton AFFICHER déplie ce que veut dire chaque chiffre (haut / moyen / bas).',
    'Provenance — la source des chiffres, en bas (ici : campagne P1 en cours).',
    'Source live — live = mesures en cours uniquement ; gelé = archives uniquement ; both = les deux.',
  ]),
  H2('Lire une carte métrique — haut, moyen, bas'),
  P(T("Chaque carte répond à trois questions. Exemple avec la plus importante, "), B("small p95"), T(" (réactivité du trafic critique) :")),
  table(['Valeur', 'Ça veut dire', 'Effet sur le trafic'],
    [
      ['HAUT (> 1000 ms)', 'Les pages et sondes traînent pendant les transferts.', 'Télémétrie et alertes inutilisables en charge.'],
      ['MOYEN (100–1000 ms)', 'Ça rame mais ça passe.', 'Tableaux de bord lents, alertes en retard.'],
      ['BAS (< 100 ms)', 'Le lien reste vif même chargé.', 'Tout cohabite : critique + transferts lourds.'],
    ], [2400, 3400, 3600]),
  P(T("Même lecture pour les autres : "), B("RTT"), T(" = temps d'aller-retour (bas = lien nerveux) ; "), B("goodput"), T(" = débit utile réel (haut = ça pousse) ; "), B("échéance"), T(" = % de sondes dans les temps (100 % = parfait) ; "), B("QDI"), T(" = gonflement de file (bas = file saine) ; "), B("drops/gaspillé/coût"), T(" = paquets perdus, octets et ariarys jetés.")),
  H2('Le mode gelé — quand on veut le passé, rien que le passé'),
  ...fig('guide-live-frozen', 'Figure 5 — Source « gelé » : les courbes live disparaissent, restent les archives.'),
  ...legende([
    'Comparaison figée — la référence grise (pfifo) contre le meilleur cyan (cake), même échelle, écart −39 %.',
    'Source — ici sur « frozen » : l’écran jure de ne montrer QUE des archives.',
  ]),
  P(B("Promesse tenue par construction : "), T("en mode "), B("gelé"), T(", les courbes, cartes et chronologie du direct sont cachées — impossible de confondre une mesure d'il y a dix secondes avec une archive d'il y a dix jours.")),
  BR(),
)
// ---------- CH 4 ----------
push(
  H1('4. Résultats — lire le classement comme un podium'),
  P(B("Toutes les campagnes gelées, rangées du meilleur au moins bon"), T(" sur le critère choisi. La barre la plus longue gagne — comme aux Jeux, pas besoin de lire les chiffres pour voir qui mène.")),
  ...fig('guide-resultats', 'Figure 6 — Le leaderboard : source, constat, duel, lignes, échelle partagée.'),
  ...legende([
    'Source — En direct = dernier run gelé avec des données ; tous runs = tout l’historique ; ou un run précis.',
    'Interprétation complète — ouvre la lecture rédigée du profil en tête (chapitre 6).',
    'Duel — pfifo (avant) contre le 1er, même échelle, écart géant au centre (−39 % ↓).',
    'Une ligne — rang, nom, barre + moustaches d’incertitude, valeur ± marge. Cliquez : ça s’ouvre, le reste s’estompe.',
    'Échelle partagée — 0 % = pire visible, 100 % = meilleur visible : toutes les barres parlent la même langue.',
  ]),
  H2('Lire une ligne en 5 secondes'),
  P(B("De gauche à droite : "), T("le "), B("rang"), T(" (petit, gris — ce n'est qu'un numéro), le "), B("nom"), T(" (« P1 · cake / cubic » = terrain fibre, file cake, envoi cubic ; ◆ = cubic, ◇ = BBR), la "), B("barre"), T(" (longueur = bonté), la "), B("valeur"), T(" (« 42,2 ms ±3 » = médiane plus/moins l'incertitude). En dessous, en une ligne : "), I("6 mesures valides · débit 53,2 Mb/s · échéances 100 % · 311 Kio gaspillés · 33 Ar.")),
  H2('Le duel — avant contre après'),
  P(B("À gauche, pfifo (sans file intelligente) : 69,5 ms. À droite, le 1er (cake/cubic) : 42,2 ms. Au centre, en énorme : "), B("−39 % ↓"), T(" — lire : "), I("« avec cake, le trafic critique répond 39 % plus vite ». En régime de perte (satellite), le centre dit « égalité » : aucune file ne sépare, la retransmission gouverne — le constater, c'est déjà décider.")),
  H2('Cliquez une ligne : le détail en français'),
  ...fig('guide-expanded', 'Figure 7 — Ligne dépliée : mesures traduites + score LIEN décomposé.'),
  ...legende([
    'Score LIEN décomposé — la note /100 et ses 5 piliers à 20 % chacun : on voit D’OÙ vient la note.',
  ]),
  P(T("Le détail dit : "), B("Latence du lien"), T(" (aller-retour sous charge), "), B("Réactivité"), T(" (charger une page pendant un transfert), "), B("Marge d'incertitude"), T(" (« la vraie valeur est là 95 fois sur 100 »), "), B("mesures écartées"), T(" (combien, pourquoi), "), B("recommandation terrain"), T(" (quoi acheter/configurer, en clair).")),
  H2('Le TradeSpace — le nuage qui montre les compromis'),
  P(T("Chaque point = une combinaison file×congestion. Axe vertical : le critère du classement. Axe horizontal : au choix (pastilles : débit, coût, latence…). "), B("En haut à droite, « optimal ↗ »"), T(" : c'est par là que vivent les bons compromis. Survolez un point pour son nom exact ; seuls les 3 meilleurs sont étiquetés, pour rester lisible.")),
  P(B("Exporter CSV / MD"), T(" — télécharge le classement pour le rapport (annexe du mémoire).")),
  BR(),
)
// ---------- CH 5 ----------
push(
  H1('5. Intégrité — la preuve que les chiffres sont vrais'),
  P(B("Si Résultats est le podium, Intégrité est l'anti-dopage : "), T("chaque mesure gelée est signée (SHA-256), chaque mesure écartée est listée avec sa cause. Rien ne disparaît en silence.")),
  ...fig('guide-integrite', 'Figure 8 — Le bandeau de preuve et les archives : combien, combien de valides, combien d’écartées.'),
  ...legende([
    'Bandeau de preuve — 155 runs, 140 manifests, 228 valides, 233 quarantaine : le bilan en une ligne.',
    'Runs archivés — 12 par page (1 … 21) : lignes, valides, écartées, lien résultats, bouton REJOUER.',
  ]),
  ...fig('guide-integrite-bas', 'Figure 9 — Quarantaine paginée : chaque ligne écartée dit pourquoi (portes G..).'),
  ...legende([
    'Quarantaine — run, événement, cellule, statut, portes fautives : l’exclusion est documentée, pas cachée.',
  ]),
  ...fig('guide-changelog', 'Figure 10 — Le journal : qui a fait quoi, quand (campagnes, audits, tris).'),
  ...legende([
    'Changelog — horodaté, pastille par nature (cyan = campagne, ambre = audit) : l’histoire de la machine.',
  ]),
  P(B("VÉRIFIER MANIFESTES"), T(" recalcule toutes les signatures : si un CSV a été touché, ça se voit ici. "), B("REJOUER"), T(" rejoue un run gelé en direct sur le Tableau live.")),
  BR(),
)
// ---------- CH 6 ----------
push(
  H1("6. Interprétation — la phrase qui dit quoi faire"),
  P(B("Le bouton "), B("Interprétation complète"), T(" (constat, ou ligne dépliée) ouvre une lecture rédigée du profil : pas de jargon, des phrases.")),
  ...fig('guide-interpretation', 'Figure 11 — L’interprétation P1 : verdict, 4 mesures traduites, recommandation, commande copiable.'),
  ...legende([
    'Verdict — une phrase : qui gagne, de combien, sur quoi.',
    'COPIER — la commande exacte pour la passerelle, prête à coller (sudo tc … cake …).',
  ]),
  P(B("Lisez dans l'ordre : "), T("verdict → 4 mesures (chacune dit ce qu'elle mesure et ce que « bas » veut dire) → recommandation terrain → commande → prochaines étapes numérotées. En 2 minutes, un non-spécialiste sait quoi décider.")),
  BR(),
)
// ---------- CH 7 ----------
push(
  H1('7. Kit TUI — piloter la machine sans taper de commande'),
  P(M("cgo.exe kit tui"), T(" ouvre le centre de contrôle : flèches + entrée, cinq étapes, zéro commande à taper.")),
  table(['Étape', 'On y fait quoi', 'Touches'],
    [
      ['1 Dépendances', 'Vérifie OpenSSH, hyperviseur, config (installe OpenSSH si besoin).', 'entrée = installer / revérifier'],
      ['2 Machine', 'Scanne VMware + VirtualBox, affiche état + réseau (nat/ponté), verrouille au clavier.', 'j/k + entrée = verrouiller'],
      ['3 Accès SSH', 'Diagnostic clé/port/auth/IP, création de clé, édition user/hôte/port, pose de clé, boot.', 'entrée = agir, esc = retour'],
      ['4 Déployer', 'Pousse le binaire, installe, vérifie la santé, ouvre le dashboard.', 'entrée = déployer'],
      ['5 Contrôle', 'Dashboard start/stop, VM start/stop, réseau invité, logs, dns, tls, snapshot, backup.', '1-5 = sauter, q = quitter'],
    ], [2200, 4200, 3000]),
  P(B("Règle d'or : "), T("suivez les numéros 1→5 la première fois ; ensuite, la 5 suffit au quotidien.")),
  BR(),
)
// ---------- CH 8 ----------
push(
  H1('8. Si ça coince — 5 pannes, 5 remèdes'),
  table(['Ça dit / ça fait', 'Ça veut dire', 'Faites ceci'],
    [
      ['Port 22 fermé', 'Pas de serveur SSH dans la VM.', 'Dans la console Ubuntu : sudo apt install -y openssh-server && sudo systemctl enable --now ssh, puis Réessayer.'],
      ['Clé refusée', 'La VM ne connaît pas votre clé.', 'Poser la clé SSH (mot de passe de la VM, une fois).'],
      ['Run vide, 0 groupe', 'Campagne tuée avant le premier gel.', 'Relancez ; En direct saute les runs vides tout seul.'],
      ['Égalité partout (P3)', 'La perte gouverne, pas la file.', 'Normal sur satellite : lisez la recommandation, pas le podium.'],
      ['Avertissement HTTPS', 'Certificat auto-signé inconnu.', 'kit tls (admin) ou Avancé → Continuer une fois.'],
    ], [2400, 3200, 3800]),
  BR(),
  H1('Glossaire — 12 mots à garder sous la main'),
  ...[
    ['AQM', 'File intelligente (cake, fq_codel) : range les paquets pour protéger l’urgent.'],
    ['BBR / cubic', 'Deux conduites d’envoi : BBR sonde en douceur, cubic remplit les files.'],
    ['Cellule', 'Un trio profil × file × congestion mesuré.'],
    ['Échéance', 'Le délai cible en ms : au-delà, la mesure est ratée.'],
    ['G0–G7', 'Les 8 gardiens qualité de chaque mesure.'],
    ['LIEN /100', 'La note composite : 5 critères à 20 % chacun.'],
    ['p95', '« 95 % des mesures sont sous cette valeur » : le pire honnête, pas la moyenne qui ment.'],
    ['Profil', 'Le décor : P1 fibre, P2 ADSL, P3 satellite, P4 4G.'],
    ['Quarantaine', 'Mesures écartées AVANT calcul, listées avec leur cause.'],
    ['Run', 'Une campagne gelée : dossier + CSV + signature.'],
    ['SSE', 'Le tuyau temps réel qui pousse les mesures à l’écran.'],
    ['TradeSpace', 'Le nuage compromis débit/latence/coût.'],
  ].map(([t, d]) => P(B(t + ' — '), T(d))),
  P(B('Meteolink 1.2.3 — guide d’interface. Bonne mesure.')),
)

const doc = new Document({
  styles: {
    default: {
      document: { run: { font: 'Calibri', size: 22 } },
      heading1: { run: { font: 'Calibri', size: 32, bold: true } },
      heading2: { run: { font: 'Calibri', size: 26, bold: true } },
    },
  },
  sections: [{ children }],
})

Packer.toBuffer(doc).then((buf) => {
  fs.writeFileSync(OUT, buf)
  console.log('OK', OUT, buf.length, 'octets')
})
