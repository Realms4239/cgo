// build-novice.js — « Comprendre Meteolink en une soirée » (~8-10 p.).
// Première lecture pour étudiant novice : Mode B restreint (ordre par dépendance),
// ancre péage/autoroute verrouillée, typographie cognitive SPECTRA-EDU (gras des
// termes à première occurrence, italique des ponts, exemples indentés, gates crayon).
'use strict';

const fs = require('fs');
const path = require('path');
const {
  Document, Packer, Paragraph, TextRun, ImageRun, AlignmentType,
} = require('docx');

const MM = (mm) => Math.round(mm * 56.6929);
const OUT = path.join(__dirname, 'livrable1-novice.docx');
const FIG = (f) => path.join(__dirname, 'figures', f + '.png');

function pngSize(file) {
  const buf = fs.readFileSync(file);
  return { w: buf.readUInt32BE(16), h: buf.readUInt32BE(20) };
}

// ---------------------------------------------------------------- helpers
const children = [];

// Paragraphe courant : justifié, interligne 1,5, une idée par paragraphe.
// runs = tableau de fragments { t, b, i } (texte, gras, italique).
const P = (runs, opts = {}) => children.push(new Paragraph({
  alignment: AlignmentType.JUSTIFIED,
  spacing: { line: 360, after: opts.after ?? 120, before: opts.before ?? 0 },
  indent: opts.noindent ? undefined : { firstLine: MM(12.5) },
  children: (Array.isArray(runs) ? runs : [{ t: runs }]).map((r) =>
    new TextRun({ text: r.t, size: 24, font: 'Times New Roman', bold: !!r.b, italics: !!r.i })),
}));

const H1 = (t) => children.push(new Paragraph({
  spacing: { before: 360, after: 200, line: 360 },
  children: [new TextRun({ text: t, size: 32, bold: true, font: 'Times New Roman' })],
}));
const H2 = (t) => children.push(new Paragraph({
  spacing: { before: 260, after: 140, line: 360 },
  children: [new TextRun({ text: t, size: 26, bold: true, font: 'Times New Roman' })],
}));

// Pont logique : italique, léger retrait, respiration avant/après.
const BRIDGE = (t) => children.push(new Paragraph({
  alignment: AlignmentType.JUSTIFIED,
  spacing: { line: 360, before: 160, after: 200 },
  indent: { left: MM(8) },
  children: [new TextRun({ text: t, size: 24, italics: true, font: 'Times New Roman' })],
}));

// Exemple concret : bloc indenté des deux côtés, filet discret.
const EX = (t) => children.push(new Paragraph({
  alignment: AlignmentType.JUSTIFIED,
  spacing: { line: 340, before: 80, after: 160 },
  indent: { left: MM(15), right: MM(8) },
  border: { left: { style: 'single', size: 6, color: '9DB4CE' } },
  children: [new TextRun({ text: t, size: 23, font: 'Times New Roman' })],
}));

// Gate crayon : encadré synthèse < 60 s, jamais un quiz.
const GATE = (t) => children.push(new Paragraph({
  spacing: { line: 340, before: 200, after: 240 },
  indent: { left: MM(8), right: MM(8) },
  border: {
    top: { style: 'single', size: 4, color: '808080' },
    bottom: { style: 'single', size: 4, color: '808080' },
  },
  children: [
    new TextRun({ text: 'À toi, 30 secondes, un crayon : ', size: 23, bold: true, italics: true, font: 'Times New Roman' }),
    new TextRun({ text: t, size: 23, italics: true, font: 'Times New Roman' }),
  ],
}));

// Figure centrée + légende italique.
const FIGURE = (file, cap, hPx) => {
  const { w, h } = pngSize(file);
  const target = hPx ?? 300;
  const scale = Math.min(560 / w, target / h);
  children.push(new Paragraph({
    alignment: AlignmentType.CENTER,
    spacing: { before: 160, after: 60 },
    children: [new ImageRun({
      type: 'png', data: fs.readFileSync(file),
      transformation: { width: Math.round(w * scale), height: Math.round(h * scale) },
    })],
  }));
  children.push(new Paragraph({
    alignment: AlignmentType.CENTER,
    spacing: { after: 200, line: 276 },
    children: [new TextRun({ text: cap, size: 22, italics: true, font: 'Times New Roman' })],
  }));
};

// ================================================================= CONTENU

H1('Comprendre Meteolink en une soirée');

P([
  { t: 'Ce document se lit en une soirée, avec un café. Il t\'apprend ce que dit le mémoire, pourquoi ça compte, et comment utiliser l\'outil. Lis-le ' },
  { t: 'avant tout le reste', b: true },
  { t: ' : ensuite, le document d\'explication complet prépare ton oral, et la thèse creuse chaque point.' },
], { noindent: true });

P([
  { t: 'Une règle de lecture : chaque terme technique est ' },
  { t: 'en gras à sa première apparition', b: true },
  { t: ', et défini juste avant d\'être utilisé. Les phrases en ' },
  { t: 'italique', i: true },
  { t: ' font le lien entre deux idées : elles relient ce que tu viens d\'apprendre à ce qui arrive.' },
], { noindent: true });

// ---------------------------------------------------------------- S1
H1('1. Pourquoi ce mémoire existe');

P([
  { t: 'Météo Madagascar', b: true },
  { t: ' protège la population : prévoir les cyclones, alerter à temps. Pour ça, ses stations envoient des relevés, ses équipes consultent des tableaux de bord, ses serveurs téléchargent des images satellite. Tout ce trafic passe par Internet.' },
]);

P([
  { t: 'Les quatre classes de liens', b: true },
  { t: ' de l\'institution ne se valent pas. Le siège d\'Antananarivo a la ' },
  { t: 'fibre optique', b: true },
  { t: ', rapide. Les directions régionales ont la ' },
  { t: '4G', b: true },
  { t: ' de l\'opérateur Yas. Les stations isolées dépendent du ' },
  { t: 'satellite géostationnaire', b: true },
  { t: ', appelé VSAT : lent, six cents millisecondes d\'aller-retour. Et récemment, ' },
  { t: 'Starlink', b: true },
  { t: ' arrive avec ses satellites en orbite basse : quarante millisecondes, mais une facture d\'environ 226 000 ariary par mois d\'après la presse de 2025.' },
]);

P([
  { t: 'La contrainte la plus forte n\'est pas technique : elle est ' },
  { t: 'économique', b: true },
  { t: '. La connexion se vend prépayée, au volume. Vingt-cinq mille ariary achètent 4,5 gigaoctets pour un mois. Chaque octet gaspillé est un octet payé deux fois.' },
]);

EX('Le même site régional fait les deux sur le même lien 4G : envoyer une alerte cyclone de quelques octets, et télécharger une image satellite de plusieurs gigaoctets. L\'alerte est urgente, l\'image est patiente. Sur un lien encombré, l\'urgence attend derrière la patience.');

BRIDGE('Ce partage d\'un lien étroit entre trafic urgent et trafic massif pose le problème central : que se passe-t-il exactement quand les deux se rencontrent ? La section suivante le montre.');

FIGURE(FIG('figR4-tarifs'), 'Le prix du gigaoctet selon le forfait, opérateur principal, 2026 : le mensuel de référence est le plus cher, cinq fois le journalier.', 200);

// ---------------------------------------------------------------- S2
H1('2. Le problème : le péage embouteillé');

P([
  { t: 'Imagine une autoroute avec un seul ' },
  { t: 'péage', b: true },
  { t: '. Tant que le trafic est léger, tout passe vite. Quand un convoi de camions arrive, la file s\'allonge devant le péage. La petite voiture arrivée derrière attend autant que les camions, alors qu\'elle seule mettrait une seconde à passer.' },
]);

P([
  { t: 'Un lien Internet fonctionne pareil. Chaque équipement réseau garde une ' },
  { t: 'file d\'attente', b: true },
  { t: ', une zone tampon pour les paquets en surnombre. Le transfert massif joue les camions : il remplit la file. Les petits messages urgents attendent derrière. Le phénomène s\'appelle le ' },
  { t: 'bufferbloat', b: true },
  { t: ' : l\'attente gonfle, le service semble planté.' },
]);

P([
  { t: 'Le piège : le ' },
  { t: 'débit', b: true },
  { t: ', la quantité de données par seconde, reste intact. Ce qui explose, c\'est la ' },
  { t: 'latence', b: true },
  { t: ', le temps de réponse. Une consultation qui prend normalement cent millisecondes peut en prendre trois cents pendant un téléchargement. Un test de débit dirait que tout va bien : il ment par omission.' },
]);

EX('Concrètement à Météo Madagascar : pendant qu\'une image satellite descend sur le lien 4G, le tableau de bord des alertes met trois fois plus de temps à répondre. Rien n\'est cassé, rien n\'est lent en apparence. Le service est juste devenu inutilisable au moment où il compte.');

P([
  { t: 'Retiens l\'essentiel : ' },
  { t: 'le débit ne dit rien de la qualité d\'un lien pour le travail quotidien. Ce qui compte, c\'est la latence pendant la charge', b: true },
  { t: '. Toute la suite du mémoire découle de cette phrase.' },
]);

BRIDGE('Maintenant que tu vois le péage embouteillé, la question devient : qui peut tenir la file courte ? Deux familles de remèdes existent, et c\'est l\'objet de la section suivante.');

GATE('Dessine le péage : un camion « transfert massif » devant, une voiture « alerte » derrière. Écris sous ton dessin, en une phrase, pourquoi la voiture attend alors que le péage n\'est pas en panne.');

// ---------------------------------------------------------------- S3
H1('3. Les remèdes : des agents au péage');

P([
  { t: 'Premier levier : la ' },
  { t: 'discipline de file', b: true },
  { t: ', la règle que l\'équipement applique à sa file. Linux en propose trois, toutes testées dans le mémoire.' },
]);

P([
  { t: 'pfifo_fast', b: true },
  { t: ' est le péage sans agent : premier arrivé, premier servi, la file grandit tant qu\'elle veut. ' },
  { t: 'fq_codel', b: true },
  { t: ' est l\'agent vigilant : il surveille l\'attente et écarte quelques paquets à l\'avance pour forcer les émetteurs à ralentir. ' },
  { t: 'CAKE', b: true },
  { t: ' est l\'agent complet : la même vigilance, plus un plafond de débit et un traitement équitable par flux. fq_codel et CAKE sont dites ' },
  { t: 'actives', b: true },
  { t: ' : elles agissent sur la file au lieu de la subir.' },
]);

P([
  { t: 'Deuxième levier : le ' },
  { t: 'contrôle de congestion', b: true },
  { t: ', la règle que suit l\'émetteur pour ajuster sa vitesse. ' },
  { t: 'CUBIC', b: true },
  { t: ', le défaut de Linux, apprend en cassant : il accélère jusqu\'à perdre un paquet, puis ralentit, puis recommence. ' },
  { t: 'BBR', b: true },
  { t: ', de Google, mesure d\'abord le débit disponible et l\'aller-retour minimal, puis émet pour remplir le lien sans déborder.' },
]);

EX('BBR face au péage sans agent : il connaît sa place exacte et la tient. La file grandit quand même autour de lui. Face à un agent vigilant, la même logique devient un atout : les paquets en surnombre sont écartés tôt, la file reste courte, et le trafic urgent passe.');

P([
  { t: 'Un opérateur ne choisit pas la discipline de file des serveurs distants. Mais il contrôle ' },
  { t: 'son routeur de site', b: true },
  { t: '. La question du mémoire : quelles combinaisons de discipline et de contrôle protègent le mieux le petit trafic de Météo Madagascar ?' },
]);

BRIDGE('Pour répondre, il faut comparer ces combinaisons dans des conditions identiques et reproductibles. C\'est exactement ce que fait le banc d\'essai, présenté maintenant.');

GATE('Relie par deux flèches : « CAKE » vers « file courte », puis « file courte » vers « alerte à l\'heure ». Sur chaque flèche, écris en trois mots le mécanisme.');

// ---------------------------------------------------------------- S4
H1('4. Le banc : rejouer le lien sans rien casser');

P([
  { t: 'Le banc s\'installe sur ' },
  { t: 'le poste client lui-même', b: true },
  { t: '. Un programme unique, nommé ' },
  { t: 'cgo', b: true },
  { t: ', crée une paire de liens virtuels dans la machine. Sur l\'un, un module nommé ' },
  { t: 'netem', b: true },
  { t: ' impose au trafic le délai, la gigue et la perte du profil choisi : le poste se fait croire qu\'il parle à un lien 4G ou satellite, alors que tout reste local. Par-dessus s\'applique la discipline testée.' },
]);

P([
  { t: 'Aucun équipement de production n\'est modifié, aucune donnée réelle n\'est consommée. C\'est la condition pour expérimenter librement dans une institution de sécurité.' },
]);

FIGURE(FIG('figR1-topologie'), 'La topologie du banc : le poste client rejoue le lien d\'accès, l\'infrastructure reste intacte.', 250);

P([
  { t: 'Quatre ' },
  { t: 'profils', b: true },
  { t: ' de lien sont rejoués : ' },
  { t: 'P1', b: true },
  { t: ' la fibre du siège, 80 mégabits et 20 millisecondes ; ' },
  { t: 'P2', b: true },
  { t: ' la 4G de Yas, 20 mégabits, 100 millisecondes, un demi-point de perte : le cas d\'étude central ; ' },
  { t: 'P3', b: true },
  { t: ' le VSAT isolé, 5 mégabits et 600 millisecondes ; ' },
  { t: 'P4', b: true },
  { t: ' Starlink, 100 mégabits et 40 millisecondes avec une gigue marquée.' },
]);

P([
  { t: 'Chaque essai, appelé ' },
  { t: 'événement', b: true },
  { t: ', suit le même déroulé : trente secondes de calme pour mesurer le repos, cent vingt secondes de charge saturante, trente secondes de récupération. Puis ' },
  { t: 'huit portes de qualité', b: true },
  { t: ', de G0 à G7, vérifient l\'événement : cible joignable, charge réellement établie, latence physiquement plausible, débit cohérent avec le profil, aucune ligne dupliquée, repos stable, machine non saturée.' },
]);

EX('Un événement qui échoue une porte majeure part en quarantaine : sa ligne est conservée, comptée, mais exclue des résultats. Sur 286 lignes mesurées, 83 sont valides et 154 mises en quarantaine : le compte est publié tel quel, il fait partie des résultats. Rien n\'est caché.');

P([
  { t: 'Ce ratio assume un choix : des portes larges auraient rempli les archives et affaibli les conclusions. ' },
  { t: 'Mieux vaut peu de lignes sûres que beaucoup de lignes douteuses', b: true },
  { t: '.' },
]);

BRIDGE('Les portes garantissent la qualité de chaque mesure. Reste à savoir ce que l\'on mesure, et ce que ça a donné : c\'est la section suivante.');

GATE('Classe les quatre profils du plus rapide au plus lent en réponse. Sous chacun, écris son pire ennemi en deux mots : distance, débit, gigue ou perte.');

// ---------------------------------------------------------------- S5
H1('5. Les chiffres : soixante contre quatre-vingt-dix-huit');

P([
  { t: 'Cinq indicateurs, une phrase chacun.' },
], { noindent: true });

P([
  { t: 'La ' },
  { t: 'médiane', b: true },
  { t: ' (p50) est le temps de réponse typique. Le ' },
  { t: 'p95', b: true },
  { t: ' est le temps que subissent les cas défavorables : c\'est lui qui décide de la sensation de blocage. Le ' },
  { t: 'respect d\'échéance', b: true },
  { t: ' compte la part des petites requêtes arrivées sous un seuil fixé, ici 220 millisecondes sur le profil 4G : la traduction directe d\'un engagement de service. Le ' },
  { t: 'QDI', b: true },
  { t: ', l\'écart entre p95 et médiane, mesure la régularité : un écart faible, un lien prévisible. Le ' },
  { t: 'coût en ariary', b: true },
  { t: ' applique au volume gaspillé les tarifs réels des opérateurs.' },
]);

P([
  { t: 'Le résultat central du mémoire, sur le profil 4G, échéance à 220 millisecondes, transfert massif en cours :', b: true },
]);

P([
  { t: 'La file simple, pfifo_fast avec BBR, ne sauve que ' },
  { t: '60 %', b: true },
  { t: ' des requêtes à l\'heure. Les disciplines actives en sauvent ' },
  { t: '96 à 98 %', b: true },
  { t: ' : fq_codel 96,2 %, CAKE 98,1 %. La régularité suit : cinquante-sept millisecondes d\'écart sous la file simple, environ treize sous les disciplines actives.' },
]);

FIGURE(FIG('figR2-deadline'), 'Le résultat central : respect de l\'échéance à 220 ms sous charge, profil 4G. La file simple s\'effondre, les disciplines actives tiennent.', 300);

P([
  { t: 'Et le débit, dans tout ça ? Il est ' },
  { t: 'préservé', b: true },
  { t: ' : entre 18,2 et 19,5 mégabits par seconde sur un profil de 20. La protection de la latence ne se paie pas en débit ; elle se paie en quelques paquets écartés, chiffrés en ariary.' },
]);

EX('Sur la voix, le score R du modèle E de l\'UIT confirme la hiérarchie : environ 85 sur la 4G au repos, autour de 71 sur le VSAT chargé, une barre qu\'aucun réglage local ne peut relever. Sur le coût : quelques centaines d\'ariary par heure de charge sur le lien 4G au palier mensuel, la fibre étant dix fois moins chère au gigaoctet.');

P([
  { t: 'Trois des six cellules de la campagne finale, celles avec CUBIC, sont en quarantaine : pas pour leur latence, excellente, mais parce que CUBIC s\'effondre sur la perte de base du lien 4G et ne tient pas la charge prévue. ' },
  { t: 'Sur un lien cellulaire chargé, CUBIC ne tient pas la charge, quel que soit le soin apporté à la file', b: true },
  { t: ' : c\'est en soi un résultat.' },
]);

BRIDGE('Ces chiffres, d\'où sortent-ils exactement ? Sont-ils vérifiables, ou la parole de l\'auteur ? C\'est toute la section suivante.');

GATE('Trace deux barres côte à côte : pfifo à 60 %, CAKE à 98 % de requêtes à l\'heure. Sous les barres, écris en une phrase le mécanisme qui les sépare.');

// ---------------------------------------------------------------- S6
H1('6. La preuve : gelée et vérifiable');

P([
  { t: 'Chaque campagne écrit ses lignes dans un fichier ' },
  { t: 'CSV', b: true },
  { t: ', un tableau à colonnes fixes, une ligne par événement. Puis le fichier est ' },
  { t: 'gelé', b: true },
  { t: ' : il ne bouge plus. Un ' },
  { t: 'manifeste', b: true },
  { t: ' enregistre l\'empreinte ' },
  { t: 'SHA-256', b: true },
  { t: ' de chaque fichier : une somme unique qui change au moindre octet modifié.' },
]);

P([
  { t: 'Une commande, ' },
  { t: 'cgo verify', b: true },
  { t: ', recalcule ces sommes et signale toute divergence. Les figures du mémoire se régénèrent uniquement depuis ces fichiers gelés : ' },
  { t: 'aucune valeur n\'est saisie à la main', b: true },
  { t: '. Les lignes en quarantaine ne disparaissent pas : elles restent, comptées, avec leur raison d\'exclusion.' },
]);

EX('La chaîne complète d\'un chiffre : la sonde mesure, l\'événement franchit les huit portes, la ligne est écrite dans le CSV, le manifeste en prend l\'empreinte, la figure se régénère depuis le fichier. Un lecteur disposant du dépôt peut tout recalculer : c\'est la réponse à « qui peut vérifier ? ».');

BRIDGE('Comprendre le problème, le remède, le banc et la preuve : le mémoire est maintenant entre tes mains. Reste à savoir t\'en servir. La grande section suivante est un guide par gestes.');

// ---------------------------------------------------------------- S7
H1('7. L\'interface en cinq gestes');

P([
  { t: 'Le tableau de bord s\'ouvre dans un navigateur, quatre panneaux dans la barre du haut. Ce chapitre est un guide : ' },
  { t: 'fais ceci, vois cela, comprends ceci', b: true },
  { t: '.' },
], { noindent: true });

// Geste 1
H2('Geste 1 : t\'orienter dans les quatre panneaux');

P([
  { t: 'Campagne de mesure', b: true },
  { t: ' pilote : c\'est la vue d\'accueil. ' },
  { t: 'Tableau live', b: true },
  { t: ' montre la mesure en direct, dix fois par seconde. ' },
  { t: 'Résultats', b: true },
  { t: ' classe les campagnes finies. ' },
  { t: 'Provenance', b: true },
  { t: ', aussi appelée Intégrité, prouve que les archives n\'ont pas bougé.' },
]);

P([
  { t: 'Retiens le parcours : piloter, suivre, lire, prouver. C\'est aussi l\'ordre de ce guide.' },
]);

FIGURE(FIG('shot-campagne'), 'La vue Campagne : le cockpit trois cartes, l\'audit du lien, le formulaire de campagne.', 225);

// Geste 2
H2('Geste 2 : auditer ton lien réel');

P([
  { t: 'Dans la vue Campagne, ouvre la carte ' },
  { t: 'Audit du lien', b: true },
  { t: ' et clique LANCER. Renseigne le site, le type de lien (fibre, 5g, 4g, vsat), la durée et la cible. L\'audit mesure ton lien réel : latence au repos, latence sous charge courte, débit, perte.' },
]);

P([
  { t: 'Tu obtiens une ligne dans les archives : le profil réel de ton lien. ' },
  { t: 'On ne façonne jamais un bord sans avoir mesuré ce que le lien fait à vide', b: true },
  { t: ' : l\'audit est le point d\'entrée de toute la démarche, et il fonctionne même depuis un poste Windows.' },
]);

EX('Dans le terminal, l\'équivalent : cgo audit --link-type 4g --site "Site X" --duration 30. Le résultat rejoint data/link_audit.csv.');

// Geste 3
H2('Geste 3 : lancer ta première campagne');

P([
  { t: 'Toujours dans la vue Campagne : coche un profil (P2 pour retrouver les chiffres de ce guide), choisis les répétitions et l\'échéance. Le démarrage se fait ' },
  { t: 'en deux confirmations', b: true },
  { t: ' : la campagne engage le poste plusieurs dizaines de minutes, rien ne démarre par un clic accidentel.' },
]);

P([
  { t: 'Passe au ' },
  { t: 'Tableau live', b: true },
  { t: ' pour suivre : cartes de latence et de débit, mur de comparaison entre la référence et la configuration testée, journal des événements. La divergence entre les deux côtés du mur, c\'est ton résultat en train de naître.' },
]);

FIGURE(FIG('shot-live'), 'Le tableau live pendant une campagne : cartes, mur de comparaison, progression de l\'événement courant.', 225);

EX('Dans le terminal : cgo run --profiles P2 --reps 1. En cas de coupure de courant, relance la même commande : la campagne reprend où elle s\'était arrêtée, sans perdre une ligne déjà gelée.');

// Geste 4
H2('Geste 4 : lire et comparer');

P([
  { t: 'La vue ' },
  { t: 'Résultats', b: true },
  { t: ' classe les cellules par critère : échéance, latence, régularité, débit ou coût. Épingle une cellule A, une cellule B : la comparaison s\'affiche, avec un ' },
  { t: 'verdict en mots', b: true },
  { t: ' calculé depuis les valeurs, et une prescription de configuration prête à copier.' },
]);

P([
  { t: 'Le verdict ne dit pas « la cellule 4 est meilleure ». Il dit : ' },
  { t: '« échéance tenue à 98 %, régularité maintenue, débit préservé, coût de 356 ariary par heure ; recommandation : CAKE calé sous la capacité mesurée »', i: true },
  { t: '. Tu lis une décision, pas un tableau.' },
]);

FIGURE(FIG('shot-resultats'), 'La vue Résultats : classement, comparaison épinglée, verdict calculé, prescription copiable.', 225);

// Geste 5
H2('Geste 5 : prouver et exporter');

P([
  { t: 'La vue ' },
  { t: 'Provenance', b: true },
  { t: ' affiche les comptes : 134 campagnes archivées, 83 lignes valides, 154 en quarantaine, et l\'empreinte SHA-256 du dernier gel. Le bouton ' },
  { t: 'Vérifier manifestes', b: true },
  { t: ' recalcule les empreintes : tout écart s\'afficherait immédiatement.' },
]);

P([
  { t: 'Le lien ' },
  { t: 'Rapport MD', b: true },
  { t: ' télécharge le constat complet : la synthèse de la campagne, prête à joindre à un courriel de direction.' },
]);

FIGURE(FIG('shot-integrite'), 'La vue Provenance : comptes des archives, empreinte du dernier gel, vérification des manifestes.', 225);

EX('Dans le terminal : cgo verify recalcule toutes les empreintes, cgo figures régénère les graphiques depuis le gel, cgo kit backup rapatrie les archives. Le kit complet compte dix-huit actions, de doctor à backup.');

GATE('Écris la chaîne de confiance en quatre mots reliés par des flèches : mesure, gel, empreinte, vérification. Elle tient en une ligne de crayon et c\'est la colonne vertébrale de tout le mémoire.');

// ---------------------------------------------------------------- S8
H1('8. Les cinq chiffres à retenir');

P([
  { t: 'Si tu ne retiens que cinq nombres, que ce soient ceux-là. Chacun porte sa question probable : c\'est ainsi qu\'ils se révèlent utiles.' },
], { noindent: true });

EX('60 % contre 98 % : la file simple contre CAKE, requêtes critiques à l\'heure sous charge, profil 4G. Question : « Où est le gain ? » Réponse : dans la latence sauvée, à débit égal.');

EX('57 millisecondes vers 13 : l\'écart de régularité, QDI, entre file simple et discipline active. Question : « Pourquoi pas juste la latence ? » Réponse : parce qu\'un lien régulier, c\'est un lien que l\'on peut surveiller avec des seuils.');

EX('25 000 ariary = 4,5 gigaoctets, soit 5 556 ariary le gigaoctet, le plus mauvais rapport du catalogue. Question : « D\'où sortent vos tarifs ? » Réponse : des grilles publiques des opérateurs de 2026, le journalier étant cinq fois moins cher au gigaoctet.');

EX('220 millisecondes : l\'échéance du profil 4G, calée sur la médiane du profil, cent millisecondes, plus une marge de file. Question : « Pourquoi ce seuil ? » Réponse : parce qu\'il se situe là où les disciplines se séparent franchement.');

EX('134 campagnes, 286 lignes, 83 valides, 154 en quarantaine, chaque fichier scellé par SHA-256. Question : « Vos chiffres sont-ils vérifiables ? » Réponse : oui, chaque figure se régénère depuis le gel, et cgo verify recalcule tout.');

// ---------------------------------------------------------------- Clôture
H1('Et maintenant ?');

P([
  { t: 'Tu as le contexte, le problème, les remèdes, le banc, les chiffres et la preuve : ' },
  { t: 'le mémoire tient dans ta tête', b: true },
  { t: '. Trois suites possibles, dans l\'ordre de l\'effort.' },
]);

P([
  { t: 'Pour préparer ton oral : le document d\'explication complet, avec la rhétorique de soutenance et le kit de questions-réponses armées. Pour creuser un point : la thèse, chaque affirmation y est développée avec ses chiffres. Pour t\'en servir : le tableau de bord, les cinq gestes de ce guide, et le terminal avec ses dix-huit commandes kit.' },
]);

P([
  { t: 'Une seule chose à ne jamais oublier : ' },
  { t: 'le lien d\'accès contraint n\'est pas une fatalité', b: true },
  { t: '. Correctement configuré, le même abonnement prépayé transporte les alertes à l\'heure et les modèles à volume.' },
], { after: 240 });

// ---------------------------------------------------------------- doc
const doc = new Document({
  sections: [{
    properties: {
      page: { margin: { top: MM(25), bottom: MM(25), left: MM(25), right: MM(25) } },
    },
    children,
  }],
});

Packer.toBuffer(doc).then((buf) => {
  fs.writeFileSync(OUT, buf);
  console.log('OK', OUT, buf.length, 'octets');
});
