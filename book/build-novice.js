// build-novice.js — « Comprendre Meteolink en une soirée » (~8-10 p.).
// OVERHAUL v2 : point de départ vécu, mots simples, chaque terme technique
// expliqué entre parenthèses à SA première apparition, progression où chaque
// section répond à la question laissée ouverte par la précédente.
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

// P(runs) : paragraphe courant. runs = texte ou fragments { t, b, i }.
const P = (runs, opts = {}) => children.push(new Paragraph({
  alignment: AlignmentType.JUSTIFIED,
  spacing: { line: 360, after: opts.after ?? 140, before: opts.before ?? 0 },
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

// EX : exemple concret, bloc indenté avec filet.
const EX = (t) => children.push(new Paragraph({
  alignment: AlignmentType.JUSTIFIED,
  spacing: { line: 340, before: 80, after: 180 },
  indent: { left: MM(15), right: MM(8) },
  border: { left: { style: 'single', size: 6, color: '9DB4CE' } },
  children: [new TextRun({ text: t, size: 23, font: 'Times New Roman' })],
}));

// GATE : synthèse papier-crayon < 60 s.
const GATE = (t) => children.push(new Paragraph({
  spacing: { line: 340, before: 220, after: 260 },
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

// FIGURE : image centrée + légende.
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

// "Pourquoi cette section ?" : une ligne italique sous chaque titre de section.
const WHY = (t) => children.push(new Paragraph({
  spacing: { after: 160, line: 340 },
  indent: { left: MM(8) },
  children: [new TextRun({ text: t, size: 23, italics: true, font: 'Times New Roman' })],
}));

// ================================================================= CONTENU

H1('Comprendre Meteolink en une soirée');

P([
  { t: 'Tu n\'as pas lu la thèse. Normal : elle fait quatre-vingts pages. Ce document la remplace pour ce soir. Tu y trouveras ' },
  { t: 'ce qui se passe, pourquoi ça se passe, comment ça se passe', b: true },
  { t: ', et comment te servir de l\'outil. Une seule règle : chaque mot technique est expliqué entre parenthèses, juste au moment où il apparaît. Tu n\'as rien à savoir d\'avance.' },
], { noindent: true });

P([
  { t: 'Lis dans l\'ordre : chaque section répond à la question laissée ouverte par la précédente. Si tu comprends une section, la suivante s\'ouvre. À la fin, les cinq chiffres à retenir referment le tout.' },
], { noindent: true });

// ---------------------------------------------------------------- S1
H1('1. Le point de départ : un pays qui paie sa connexion au détail');

WHY('Pourquoi commencer ici ? Parce que tout le mémoire découle de cette situation. Sans elle, le sujet n\'existe pas.');

P([
  { t: 'Météo Madagascar', b: true },
  { t: ' (l\'organisme public qui prévoit la météo et lance les alertes cyclones) fait un travail simple à décrire : ses stations automatiques envoient des mesures, ses équipes regardent des tableaux de bord, ses serveurs téléchargent des images satellite. Tout cela voyage sur Internet.' },
]);

P([
  { t: 'Mais tous ses sites n\'ont pas le même Internet. Quatre situations existent, du plus rapide au plus lent :' },
]);

EX('Le siège, à Antananarivo : la fibre (le câble qui va jusqu\'au bâtiment, très rapide). Les directions régionales : la 4G (le réseau des téléphones portables, utilisé comme connexion fixe). Les stations isolées : le VSAT (le satellite classique, très loin : un message met 600 millisecondes pour faire l\'aller-retour). Et depuis peu : Starlink (des satellites beaucoup plus proches, 40 millisecondes).');

P([
  { t: 'La contrainte qui change tout n\'est pas la vitesse. C\'est le ' },
  { t: 'prix', b: true },
  { t: '. La 4G se vend prépayée, au volume : ' },
  { t: '25 000 ariary achètent 4,5 gigaoctets pour un mois', b: true },
  { t: '. Le gigaoctet coûte donc 5 556 ariary à ce palier, le plus mauvais rapport du catalogue : le forfait journalier, 1 000 ariary pour 1 gigaoctet, est cinq fois moins cher au gigaoctet.' },
]);

P([
  { t: 'En octobre 2025, le régulateur ARTEC (l\'autorité qui surveille les télécoms malgaches) a convoqué les trois opérateurs sur le prix de la donnée : le sujet est sur la place publique.' },
]);

P([
  { t: 'Conséquence directe : ' },
  { t: 'chaque octet (petite unité de données, il en faut un milliard pour un gigaoctet) qui voyage deux fois est payé deux fois', b: true },
  { t: '. Et un réseau encombré fait justement voyager les données plusieurs fois : quand un paquet (un morceau de données) arrive dans une file trop pleine, il est jeté, puis renvoyé. Tu paies le renvoi.' },
]);

FIGURE(FIG('figR4-tarifs'), 'Le prix du gigaoctet selon le forfait, opérateur principal, 2026. Le mensuel de référence est le plus cher : cinq fois le journalier.', 200);

WHY('Voilà le terrain. La question suivante s\'impose d\'elle-même : que fait cette congestion aux messages urgents ? C\'est la section 2.');

// ---------------------------------------------------------------- S2
H1('2. Le problème : le petit message pris dans la file');

WHY('Pourquoi cette section ? Parce qu\'il faut voir le problème de ses propres yeux avant d\'entendre son nom savant.');

P([
  { t: 'Imagine un péage d\'autoroute. Un seul guichet. Un jour, un convoi de cent camions arrive : la file d\'attente devant le guichet s\'allonge. Une voiture arrive derrière : elle aussi attend. ' },
  { t: 'Pourtant le guichet n\'est pas en panne : il traite les véhicules à la même vitesse qu\'avant.', b: true },
  { t: ' Le problème n\'est pas la vitesse du guichet, c\'est la file.' },
]);

P([
  { t: 'Un lien Internet a exactement le même guichet : chaque équipement réseau garde une ' },
  { t: 'file d\'attente', b: true },
  { t: ' (une zone mémoire où les paquets patientent avant de partir). Quand quelqu\'un télécharge un gros fichier, la file se remplit de ses paquets. Le petit message urgent arrive derrière. Il attend. Longtemps.' },
]);

P([
  { t: 'Ce phénomène a un nom : le ' },
  { t: 'bufferbloat', b: true },
  { t: ' (littéralement « la file qui gonfle » : buffer = file d\'attente, bloat = gonflement). Sa particularité trompeuse : ' },
  { t: 'le débit (la quantité de données par seconde) reste intact', b: true },
  { t: '. Ce qui se dégrade, c\'est la ' },
  { t: 'latence', b: true },
  { t: ' (le temps de réponse : la durée entre « j\'envoie » et « je reçois la réponse »).' },
]);

EX('Sur le lien 4G de Météo Madagascar : pendant qu\'une image satellite de plusieurs gigaoctets descend, le tableau de bord des alertes ne répond plus pendant de longues secondes. Un test de débit dirait « tout va bien ». Il ment : il mesure le camion, pas la voiture.');

P([
  { t: 'Retiens la phrase qui fonde tout le mémoire : ' },
  { t: 'le débit ne dit rien de la qualité d\'un lien pour le travail quotidien. Ce qui compte, c\'est la latence pendant la charge', b: true },
  { t: '. Tout ce qui suit cherche à mesurer cette latence et à la protéger.' },
]);

GATE('Dessine le péage : cent camions, une voiture derrière. Sous ton dessin, réponds en une phrase : pourquoi la voiture attend-elle alors que le guichet n\'est pas en panne ?');

WHY('Le problème est posé. Question suivante : qui peut tenir la file courte ? Deux réponses existent, et les connaître change la façon de configurer le réseau.');

// ---------------------------------------------------------------- S3
H1('3. Deux endroits où agir : la file et l\'émetteur');

WHY('Pourquoi cette section ? Parce qu\'un remède seul ne suffit pas : il faut savoir que le problème a deux causes, donc deux leviers.');

P([
  { t: 'Premier endroit : ' },
  { t: 'la file elle-même', b: true },
  { t: '. La règle qui gère la file s\'appelle la ' },
  { t: 'discipline de file', b: true },
  { t: ' (la méthode que l\'équipement utilise pour décider qui part, qui attend, qui est jeté). Linux, le système des routeurs, en propose trois principales, toutes testées ici :' },
]);

EX('pfifo_fast : le guichet sans agent. Premier arrivé, premier servi, la file grandit sans limite. C\'est le réglage par défaut, le témoin.\n\nfq_codel : l\'agent vigilant. Il surveille l\'attente et jette quelques paquets volontairement, ce qui force les émetteurs à ralentir : la file reste courte.\n\nCAKE : l\'agent complet. La même vigilance, plus deux raffinements : un plafond de débit (ne jamais laisser entrer plus que le lien ne peut évacuer) et un partage équitable entre les usages.\n\nfq_codel et CAKE sont dites « actives » : elles agissent au lieu de subir.');

P([
  { t: 'Deuxième endroit : ' },
  { t: 'l\'émetteur', b: true },
  { t: ', celui qui envoie les données. Sa règle s\'appelle le ' },
  { t: 'contrôle de congestion', b: true },
  { t: ' (la méthode dont il se sert pour deviner à quelle vitesse il peut envoyer sans noyer le réseau). Deux écoles :' },
]);

EX('CUBIC : l\'écolier qui apprend en se cognant. Il accélère jusqu\'à ce qu\'un paquet soit perdu, déduit qu\'il est allé trop vite, ralentit, puis recommence. C\'est le réglage par défaut de Linux.\n\nBBR : l\'écolier qui mesure d\'abord. Il estime le débit disponible et le temps minimum du trajet, puis envoie exactement cette quantité, sans attendre de casser quelque chose.');

P([
  { t: 'Pourquoi tester les deux ensembles ? Parce qu\'un opérateur institutionnel ne choisit pas la discipline des serveurs lointains. Mais ' },
  { t: 'il contrôle son propre routeur', b: true },
  { t: ' et parfois ses émetteurs. Il a besoin de savoir quelles combinaisons protègent son trafic urgent. C\'est une question de terrain, pas de curiosité.' },
]);

GATE('Deux flèches à tracer : « CAKE » vers « file courte », puis « file courte » vers « alerte à l\'heure ». Sur chaque flèche, le mécanisme en trois mots.');

WHY('Les remèdes sont nommés. Question suivante : comment les comparer loyalement ? Il faudrait un réseau d\'essai identique à chaque fois. C\'est le banc.');

// ---------------------------------------------------------------- S4
H1('4. Le banc : un réseau d\'essai dans un seul ordinateur');

WHY('Pourquoi cette section ? Parce que sans banc, aucune comparaison n\'est loyale : le réseau réel change tout le temps.');

P([
  { t: 'Le problème pour comparer : sur un réseau réel, tout bouge. La charge des collègues, la météo radio, le moment de la journée. Tu ne peux jamais rejouer deux fois la même chose. La solution : ' },
  { t: 'rejouer le lien dans un ordinateur, à l\'identique, autant de fois qu\'on veut', b: true },
  { t: '.' },
]);

P([
  { t: 'L\'outil s\'appelle ' },
  { t: 'cgo', b: true },
  { t: '. Il crée dans la machine une paire de liens virtuels (deux prises réseau fictives, reliées par un câble fictif). Sur ce lien fictif, un module nommé ' },
  { t: 'netem', b: true },
  { t: ' (émulateur réseau : il déforme volontairement les paquets) impose le délai, la gigue (variation du délai) et le taux de perte d\'un vrai lien malgache. La discipline testée s\'applique par-dessus.' },
]);

P([
  { t: 'L\'ordinateur se fait croire qu\'il parle à une 4G de Yas ; en réalité, rien ne quitte la pièce', b: true },
  { t: '. Aucun équipement réel n\'est touché, aucune donnée réelle n\'est consommée.' },
]);

FIGURE(FIG('figR1-topologie'), 'Le banc : le poste de travail rejoue le lien d\'accès. Aucun équipement réel n\'est touché.', 210);

P([
  { t: 'Quatre ' },
  { t: 'profils', b: true },
  { t: ' (modèles de lien) sont rejoués : P1 la fibre du siège (80 mégabits, 20 millisecondes), P2 la 4G de Yas (20 mégabits, 100 millisecondes, 0,5 % de perte : le cas central), P3 le VSAT (5 mégabits, 600 millisecondes), P4 Starlink (100 mégabits, 40 millisecondes, gigue marquée).' },
]);

P([
  { t: 'Chaque essai, appelé ' },
  { t: 'événement', b: true },
  { t: ', dure trois minutes et suit toujours le même déroulé : 30 secondes de calme (mesure du repos), 120 secondes de charge (un téléchargement massif sature le lien, les petites sondes continuent), 30 secondes de récupération.' },
]);

P([
  { t: 'Puis ' },
  { t: 'huit portes de qualité', b: true },
  { t: ' (huit vérifications automatiques, G0 à G7) contrôlent l\'essai : la cible répond-elle ? la charge a-t-elle vraiment eu lieu ? la latence est-elle physiquement possible ? le débit est-il cohérent avec le profil ? et quatre autres contrôles du même genre.' },
]);

EX('Le résultat de ces portes est impitoyable et public : sur 286 lignes mesurées, 83 ont passé toutes les portes, 154 ont échoué au moins une porte majeure et sont « en quarantaine » (conservées mais exclues des conclusions). Le mémoire publie ce ratio tel quel.\n\nPourquoi tant d\'échecs ? Parce que les portes ont été durcies pendant la mise au point : chaque durcissement rejetait en bloc les essais anciens. Des portes larges auraient rempli les archives et affaibli les conclusions : mieux vaut peu de lignes sûres.');

GATE('Classe les quatre profils du plus rapide au plus lent à répondre. Sous chacun, son pire ennemi en deux mots : distance, débit, gigue ou perte.');

WHY('Le banc produit des essais propres. Question suivante : que mesure-t-on exactement dans ces essais, et qu\'ont-ils donné ?');

// ---------------------------------------------------------------- S5
H1('5. Les chiffres : ce qui a été mesuré, et ce que ça dit');

WHY('Pourquoi cette section ? Parce que c\'est le résultat du mémoire. Tout ce qui précède sert à rendre ces chiffres crédibles.');

P([
  { t: 'Cinq mesures, une phrase chacune.' },
], { noindent: true });

P([
  { t: 'La ' },
  { t: 'médiane', b: true },
  { t: ' (la valeur du milieu : la moitié des cas font mieux, l\'autre moitié moins bien) donne le cas typique. Le ' },
  { t: 'p95', b: true },
  { t: ' (le temps que subissent les 5 % des cas les plus lents) donne les cas pénibles : c\'est lui qui décide de la sensation de blocage.' },
]);

P([
  { t: 'Le ' },
  { t: 'respect d\'échéance', b: true },
  { t: ' (le pourcentage de petites requêtes arrivées sous un seuil fixé, ici 220 millisecondes) traduit un engagement de service : « 95 % des consultations répondent en moins de 220 ms ». Le ' },
  { t: 'QDI', b: true },
  { t: ' (l\'écart entre p95 et médiane : grand écart = lien imprévisible) mesure la régularité.' },
]);

P([
  { t: 'Enfin, le ' },
  { t: 'coût en ariary', b: true },
  { t: ' convertit le gaspillage en argent réel : les octets jetés puis renvoyés, au tarif public du forfait.' },
]);

P([
  { t: 'Le résultat central, profil 4G, transfert massif en cours, échéance 220 millisecondes :', b: true },
]);

EX('Avec la file simple (pfifo_fast) et l\'émetteur BBR : 60 % des petites requêtes arrivent à l\'heure. Avec fq_codel : 96 %. Avec CAKE : 98 %. Et le débit du téléchargement massif ? Identique partout : entre 18,2 et 19,5 mégabits par seconde sur un lien de 20. La protection ne coûte pas de vitesse ; elle coûte quelques paquets jetés volontairement, payés en octets renvoyés.');

FIGURE(FIG('figR2-deadline'), 'Le résultat central : respect de l\'échéance à 220 ms sous charge, profil 4G. La file simple s\'effondre, les actives tiennent.', 260);

P([
  { t: 'Pourquoi cet écart est-il massif et pas du bruit ? Sur les vagues antérieures, la même hiérarchie revient : onze répétitions de la file simple, sept de fq_codel, six de CAKE, toutes concordantes.' },
]);

P([
  { t: 'La régularité suit le même ordre : écart de 57 millisecondes sous la file simple, environ 13 sous les actives. ' },
  { t: 'Un lien régulier, c\'est un lien qu\'on peut surveiller', b: true },
  { t: ' : une alerte de latence y veut dire quelque chose.' },
]);

P([
  { t: 'Un résultat inattendu, à connaître : les essais avec CUBIC, l\'émetteur prudent, ont été mis en quarantaine. Pas pour leur latence : excellente. Parce que sur un lien 4G qui perd naturellement 0,5 % de ses paquets, CUBIC prend chaque perte pour un accident de sa faute, ralentit, et ne tient plus la charge prévue. ' },
  { t: 'Sur un lien cellulaire, le choix de l\'émetteur pèse autant que celui de la file', b: true },
  { t: '.' },
]);

GATE('Trace deux barres : pfifo à 60 %, CAKE à 98 % de requêtes à l\'heure. Dessous, le mécanisme en une phrase.');

WHY('Ces chiffres seraient de la parole d\'auteur s\'ils n\'étaient pas vérifiables. Question finale avant l\'outil : qui peut les contrôler ?');

// ---------------------------------------------------------------- S6
H1('6. La preuve : des chiffres que personne ne peut modifier');

WHY('Pourquoi cette section ? Parce qu\'un jury, un directeur, un collègue demandera : « pourquoi vous croire ? ». La réponse est une chaîne, pas un argument.');

P([
  { t: 'Chaque essai écrit une ligne dans un fichier ' },
  { t: 'CSV', b: true },
  { t: ' (un tableau dans un fichier texte : une ligne par mesure, des colonnes fixes). Ensuite le fichier est ' },
  { t: 'gelé', b: true },
  { t: ' : il ne bouge plus.' },
]);

P([
  { t: 'Et un ' },
  { t: 'manifeste', b: true },
  { t: ' (la liste officielle des fichiers du run) enregistre pour chaque fichier son ' },
  { t: 'SHA-256', b: true },
  { t: ' (une empreinte digitale numérique : une suite de caractères unique, calculée à partir du fichier entier, qui change si un seul octet change).' },
]);

P([
  { t: 'Une commande, ' },
  { t: 'cgo verify', b: true },
  { t: ', recalcule toutes les empreintes et signale tout écart. ' },
  { t: 'Les figures du mémoire sont régénérées uniquement depuis ces fichiers gelés : aucune valeur n\'est jamais tapée à la main', b: true },
  { t: '. Même les 154 lignes en quarantaine restent visibles, avec leur raison d\'exclusion.' },
]);

EX('La chaîne complète d\'un chiffre, de bout en bout : la sonde mesure ; l\'événement franchit les huit portes ; la ligne s\'écrit dans le CSV ; le manifeste prend l\'empreinte du fichier ; la figure se régénère depuis le fichier. N\'importe qui, avec le dépôt, peut tout recalculer. C\'est la réponse à « qui peut vérifier ? ».');

GATE('Écris la chaîne de confiance en quatre mots reliés par des flèches : mesure, gel, empreinte, vérification. C\'est la colonne vertébrale du mémoire.');

WHY('Tu as maintenant le pourquoi, le problème, les remèdes, le banc, les chiffres et la preuve. Reste l\'outil lui-même : comment s\'en servir, concrètement, gestes par gestes.');

// ---------------------------------------------------------------- S7
H1('7. L\'outil en cinq gestes');

WHY('Pourquoi cette section ? Parce que comprendre ne suffit pas : la soutenance peut demander une démonstration, et la DSI veut voir l\'outil tourner.');

P([
  { t: 'Le tableau de bord s\'ouvre dans un navigateur (une page web servie par l\'outil lui-même, pas de site distant). Quatre panneaux dans la barre du haut. Le parcours de ce guide suit l\'ordre naturel d\'une vraie session : ' },
  { t: 'piloter, suivre, lire, prouver', b: true },
  { t: '.' },
], { noindent: true });

H2('Geste 1 : t\'orienter');

P([
  { t: 'Campagne', b: true },
  { t: ' prépare et pilote les essais : c\'est la vue d\'accueil. ' },
  { t: 'Tableau live', b: true },
  { t: ' montre la mesure en direct, dix fois par seconde. ' },
  { t: 'Résultats', b: true },
  { t: ' classe les essais finis. ' },
  { t: 'Provenance', b: true },
  { t: ' (parfois appelée Intégrité) prouve que rien n\'a bougé. Quand tu ne sais pas où regarder : demande-toi si tu pilotes, suis, lis ou prouves.' },
]);

FIGURE(FIG('shot-campagne'), 'La vue Campagne : le cockpit en trois cartes, la carte d\'audit, le formulaire de campagne.', 200);

H2('Geste 2 : auditer ton lien réel');

P([
  { t: 'Vue Campagne, carte ' },
  { t: 'Audit du lien', b: true },
  { t: ' : clique LANCER, renseigne le site, le type de lien (fibre, 4g, vsat...), la durée, la cible. L\'outil mesure ton lien réel, en douceur : latence au repos, latence pendant une courte charge, débit, perte. ' },
  { t: 'Règle d\'or : on ne configure jamais un routeur sans avoir mesuré le lien à vide d\'abord', b: true },
  { t: '. Terminal : cgo audit --link-type 4g --site "Site X" --duration 30.' },
]);

H2('Geste 3 : lancer ta première campagne');

P([
  { t: 'Vue Campagne : coche P2, choisis les répétitions et l\'échéance, puis ' },
  { t: 'deux confirmations', b: true },
  { t: ' (la campagne occupe le poste des dizaines de minutes : rien ne démarre par un clic raté).' },
]);

P([
  { t: 'Bascule sur ' },
  { t: 'Tableau live', b: true },
  { t: ' : cartes de latence et de débit, et le mur de comparaison (la référence d\'un côté, la configuration testée de l\'autre, la même échelle). ' },
  { t: 'L\'écart entre les deux côtés du mur, c\'est ton résultat en train de naître', b: true },
  { t: '. Terminal : cgo run --profiles P2 --reps 1. Coupure de courant ? Relance la même commande : la campagne reprend où elle s\'était arrêtée.' },
]);

FIGURE(FIG('shot-live'), 'Le tableau live : cartes, mur de comparaison, progression de l\'essai courant.', 200);

H2('Geste 4 : lire et comparer');

P([
  { t: 'Vue ' },
  { t: 'Résultats', b: true },
  { t: ' : le classement des cellules (une cellule = une combinaison profil + discipline + émetteur). Épingle une cellule A, une cellule B : la comparaison affiche un ' },
  { t: 'verdict en mots', b: true },
  { t: ' calculé depuis les valeurs, avec une prescription de configuration (le réglage recommandé, prêt à copier).' },
]);

P([
  { t: 'Le verdict ne dit pas « la cellule 4 gagne ». Il dit : ' },
  { t: '« échéance tenue à 98 %, régularité maintenue, débit préservé, coût 356 ariary par heure ; recommandation : CAKE calé sous la capacité mesurée »', i: true },
  { t: '. Tu lis une décision, pas un tableau.' },
]);

FIGURE(FIG('shot-resultats'), 'La vue Résultats : classement, comparaison épinglée, verdict calculé, prescription copiable.', 200);

H2('Geste 5 : prouver et exporter');

P([
  { t: 'Vue ' },
  { t: 'Provenance', b: true },
  { t: ' : les comptes (134 campagnes, 83 lignes valides, 154 en quarantaine) et l\'empreinte du dernier gel. Le bouton ' },
  { t: 'Vérifier manifestes', b: true },
  { t: ' recalcule les empreintes sur place. Le lien ' },
  { t: 'Rapport MD', b: true },
  { t: ' télécharge le constat complet, prêt à joindre à un courriel de direction. Terminal : cgo verify, cgo figures, cgo kit backup (le kit compte dix-huit commandes, de doctor à backup).' },
]);

FIGURE(FIG('shot-integrite'), 'La vue Provenance : comptes des archives, empreinte du dernier gel, vérification des manifestes.', 200);

GATE('Écris les quatre panneaux et leur verbe : Campagne pilote, Live suit, Résultats lit, Provenance prouve. Quatre mots, un chacun.');

// ---------------------------------------------------------------- S8
H1('8. Les cinq chiffres à retenir');

P([
  { t: 'Si tu ne retiens que cinq nombres, que ce soient ceux-là. Chacun arrive avec la question qu\'on te posera.' },
], { noindent: true });

EX('60 % contre 98 % : la file simple contre CAKE, petites requêtes à l\'heure sous charge, profil 4G. Question attendue : « où est le gain ? » Réponse : dans la latence sauvée, à débit égal.');

EX('57 millisecondes vers 13 : l\'écart de régularité (QDI) entre file simple et discipline active. Question : « pourquoi pas juste la latence ? » Réponse : un lien régulier est un lien qu\'on peut surveiller avec des seuils.');

EX('25 000 ariary = 4,5 gigaoctets, soit 5 556 ariary le gigaoctet : le plus mauvais rapport du catalogue, le journalier étant cinq fois moins cher. Question : « d\'où sortent vos tarifs ? » Réponse : des grilles publiques des opérateurs de 2026.');

EX('220 millisecondes : l\'échéance du profil 4G, calée sur le cas médian (100 ms) plus la marge de file. Question : « pourquoi ce seuil ? » Réponse : c\'est là que les disciplines se séparent franchement.');

EX('134 campagnes, 286 lignes, 83 valides, 154 en quarantaine, le tout scellé par SHA-256. Question : « vos chiffres sont-ils vérifiables ? » Réponse : chaque figure se régénère depuis le gel, cgo verify recalcule tout.');

// ---------------------------------------------------------------- Clôture
H1('Et maintenant ?');

P([
  { t: 'Tu as le terrain, le problème, les remèdes, le banc, les chiffres et la preuve : ' },
  { t: 'le mémoire tient maintenant dans ta tête', b: true },
  { t: '. Trois suites, par ordre d\'effort : le document d\'explication complet (il prépare ton oral, avec le kit de questions-réponses armées) ; la thèse (chaque affirmation, développée avec ses chiffres) ; et l\'outil lui-même, avec ses cinq gestes et ses dix-huit commandes.' },
]);

P([
  { t: 'Une seule phrase à ne jamais perdre : ' },
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
