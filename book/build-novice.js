// build-novice.js — « Comprendre Meteolink » (~10 p.).
// v3 : advance organizer (skimming séquentiel en intro), corps chronologique
// (la vie d'une mesure), termes en gras + définition inline, un concept par
// section, gras réservé aux premiers emplois et résultats clés, glossaire final.
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

H1('Comprendre Meteolink');

P([
  { t: 'En gros, le projet répond à une question simple : ' },
  { t: 'quand un gros téléchargement sature le lien Internet, que deviennent les petits messages urgents qui partagent ce lien ?', b: true },
  { t: ' Chez Météo Madagascar, la question n\'est pas théorique : les alertes cyclones voyagent sur les mêmes liens prépayés que les images satellite, et chaque octet renvoyé est payé deux fois.' },
], { noindent: true });

P([
  { t: 'Meteolink est l\'outil qui y répond, et voici son parcours complet, dans l\'ordre où les choses se passent.' },
], { noindent: true });

P([
  { t: 'D\'abord il ' },
  { t: 'mesure le lien réel', b: true },
  { t: ' : un audit discret, lancé depuis un simple poste, relève ce que le lien fait au repos et sous charge. Ensuite il ' },
  { t: 'rejoue ce lien dans un banc d\'essai', b: true },
  { t: ', à l\'identique, dans un seul ordinateur : même délai, même gigue, même perte qu\'une vraie 4G ou un vrai satellite.' },
]);

P([
  { t: 'Sur ce lien rejoué, il ' },
  { t: 'compare les remèdes connus', b: true },
  { t: ' : trois façons de gérer la file d\'attente du routeur, deux façons d\'émettre les données, dans chaque sens de circulation (montée, descente), toutes les combinaisons. Chaque mesure passe ' },
  { t: 'huit contrôles automatiques', b: true },
  { t: ', puis est ' },
  { t: 'gelée dans un fichier scellé', b: true },
  { t: ' par une empreinte numérique : personne, pas même l\'auteur, ne peut modifier un chiffre sans que cela se voie.' },
]);

P([
  { t: 'Les résultats donnent enfin ' },
  { t: 'une recommandation concrète', b: true },
  { t: ' : quel réglage appliquer, sur quel équipement, à quel coût. Le résultat central : ' },
  { t: 'sans réglage, 60 % des requêtes urgentes arrivent à l\'heure pendant un téléchargement ; avec le bon réglage, 98 %', b: true },
  { t: '. Le tout se pilote depuis ' },
  { t: 'une interface web en quatre vues', b: true },
  { t: ' ou depuis le terminal.' },
]);

P([
  { t: 'La suite du document déroule ce parcours, étape par étape : le terrain, le problème, les remèdes, le banc, les chiffres, la preuve, l\'outil. Chaque mot technique est expliqué entre parenthèses, juste au moment où il apparaît.' },
], { noindent: true });

// ---------------------------------------------------------------- S1
H1('1. Le terrain : un pays qui paie sa connexion au détail');

P([
  { t: 'Météo Madagascar', b: true },
  { t: ' (l\'organisme public qui prévoit la météo et lance les alertes cyclones) fait un travail simple à décrire : ses stations automatiques envoient des mesures, ses équipes consultent des tableaux de bord, ses serveurs téléchargent des images satellite. Tout cela voyage sur Internet.' },
]);

P([
  { t: 'Mais tous ses sites n\'ont pas le même Internet. Le siège d\'Antananarivo a la ' },
  { t: 'fibre', b: true },
  { t: ' (le câble rapide qui va jusqu\'au bâtiment). Les directions régionales ont la ' },
  { t: '4G', b: true },
  { t: ' (le réseau des téléphones portables, utilisé comme connexion fixe). Les stations isolées dépendent du ' },
  { t: 'VSAT', b: true },
  { t: ' (le satellite classique, très loin : un message met 600 millisecondes pour faire l\'aller-retour). Et depuis peu, ' },
  { t: 'Starlink', b: true },
  { t: ' (des satellites beaucoup plus proches, 40 millisecondes) arrive sur le marché.' },
]);

P([
  { t: 'La contrainte qui change tout n\'est pas la vitesse, c\'est le prix. La 4G se vend prépayée, au volume : ' },
  { t: '25 000 ariary achètent 4,5 gigaoctets pour un mois', b: true },
  { t: ', soit 5 556 ariary le gigaoctet, le plus mauvais rapport du catalogue : le forfait journalier est cinq fois moins cher au gigaoctet.' },
]);

P([
  { t: 'En octobre 2025, le régulateur ARTEC (l\'autorité qui surveille les télécoms malgaches) a convoqué les trois opérateurs sur le prix de la donnée : le sujet est sur la place publique.' },
]);

P([
  { t: 'Conséquence directe : ' },
  { t: 'chaque octet (petite unité de données : il en faut un milliard pour un gigaoctet) qui voyage deux fois est payé deux fois', b: true },
  { t: '. Et un réseau encombré fait justement voyager les données plusieurs fois : quand un paquet (un morceau de données) arrive dans une file trop pleine, il est jeté, puis renvoyé. On paie le renvoi.' },
]);

FIGURE(FIG('figR4-tarifs'), 'Le prix du gigaoctet selon le forfait, opérateur principal, 2026. Le mensuel de référence est le plus cher : cinq fois le journalier.', 200);

// ---------------------------------------------------------------- S2
H1('2. Le problème : le petit message pris dans la file');

P([
  { t: 'Imagine un péage d\'autoroute avec un seul guichet. Un jour, un convoi de cent camions arrive : la file s\'allonge devant le guichet. Une voiture arrive derrière : elle attend elle aussi. ' },
  { t: 'Pourtant le guichet n\'est pas en panne : il traite les véhicules à la même vitesse qu\'avant', b: true },
  { t: '. Le problème n\'est pas la vitesse du guichet, c\'est la file.' },
]);

P([
  { t: 'Un lien Internet a exactement ce guichet : chaque équipement réseau garde une ' },
  { t: 'file d\'attente', b: true },
  { t: ' (une zone mémoire où les paquets patientent avant de partir). Quand quelqu\'un télécharge un gros fichier, la file se remplit de ses paquets. Le petit message urgent arrive derrière. Il attend, longtemps.' },
]);

P([
  { t: 'Le phénomène a un nom : le ' },
  { t: 'bufferbloat', b: true },
  { t: ' (« la file qui gonfle »). Sa particularité trompeuse : le ' },
  { t: 'débit', b: true },
  { t: ' (la quantité de données par seconde) reste intact. Ce qui se dégrade, c\'est la ' },
  { t: 'latence', b: true },
  { t: ' (le temps de réponse : la durée entre « j\'envoie » et « je reçois la réponse »).' },
]);

P([
  { t: 'Pendant qu\'une image satellite descend sur le lien 4G, le tableau de bord des alertes ne répond plus pendant de longues secondes. Un test de débit dirait « tout va bien » : ' },
  { t: 'il mesure le camion, pas la voiture', b: true },
  { t: '.' },
]);

P([
  { t: 'Retiens la phrase qui fonde tout le mémoire : ' },
  { t: 'le débit ne dit rien de la qualité d\'un lien pour le travail quotidien ; ce qui compte, c\'est la latence pendant la charge', b: true },
  { t: '. Tout ce qui suit cherche à mesurer cette latence et à la protéger.' },
]);

// ---------------------------------------------------------------- S3
H1('3. Les remèdes : gérer la file, gérer l\'émetteur');

P([
  { t: 'Le problème a deux causes, donc deux endroits où agir. Premier endroit : ' },
  { t: 'la file elle-même', b: true },
  { t: '. La règle qui gère la file s\'appelle la ' },
  { t: 'discipline de file', b: true },
  { t: ' (la méthode que l\'équipement utilise pour décider qui part, qui attend, qui est jeté). Linux, le système des routeurs, en propose trois principales, toutes testées ici.' },
]);

P([
  { t: 'pfifo_fast', b: true },
  { t: ' est le guichet sans agent : premier arrivé, premier servi, la file grandit sans limite. C\'est le réglage par défaut, le témoin. ' },
  { t: 'fq_codel', b: true },
  { t: ' est l\'agent vigilant : il surveille l\'attente et jette quelques paquets volontairement, ce qui force les émetteurs à ralentir ; la file reste courte.' },
]);

P([
  { t: 'CAKE', b: true },
  { t: ' est l\'agent complet : la même vigilance, plus un plafond de débit (ne jamais laisser entrer plus que le lien ne peut évacuer) et un partage équitable entre les usages. fq_codel et CAKE sont dites actives : elles agissent au lieu de subir.' },
]);

P([
  { t: 'Deuxième endroit : ' },
  { t: 'l\'émetteur', b: true },
  { t: ', celui qui envoie les données. Sa règle s\'appelle le ' },
  { t: 'contrôle de congestion', b: true },
  { t: ' (la méthode dont il se sert pour deviner à quelle vitesse envoyer sans noyer le réseau). ' },
  { t: 'CUBIC', b: true },
  { t: ', le réglage par défaut de Linux, apprend en se cognant : il accélère jusqu\'à perdre un paquet, ralentit, recommence. ' },
  { t: 'BBR', b: true },
  { t: ', de Google, mesure d\'abord : il estime le débit disponible et le temps minimum du trajet, puis envoie exactement cette quantité.' },
]);

P([
  { t: 'Pourquoi tester les deux ensembles ? Parce qu\'un opérateur institutionnel ne choisit pas la discipline des serveurs lointains, mais ' },
  { t: 'il contrôle son propre routeur', b: true },
  { t: ' et parfois ses émetteurs. Il a besoin de savoir quelles combinaisons protègent son trafic urgent.' },
]);

// ---------------------------------------------------------------- S4
H1('4. Le banc : un réseau d\'essai dans un seul ordinateur');

P([
  { t: 'Pour comparer loyalement, il faudrait rejouer le même lien, à l\'identique, autant de fois qu\'on veut. Sur un réseau réel, impossible : tout bouge, la charge des collègues, la météo radio, l\'heure de la journée. La solution : ' },
  { t: 'rejouer le lien dans un ordinateur', b: true },
  { t: '.' },
]);

P([
  { t: 'L\'outil s\'appelle ' },
  { t: 'cgo', b: true },
  { t: '. Il crée dans la machine une paire de liens virtuels (deux prises réseau fictives, reliées par un câble fictif). Sur ce lien, un module nommé ' },
  { t: 'netem', b: true },
  { t: ' (émulateur réseau : il déforme volontairement les paquets) impose le délai, la gigue (variation du délai) et le taux de perte d\'un vrai lien malgache. La discipline testée s\'applique par-dessus.' },
]);

P([
  { t: 'L\'ordinateur se fait croire qu\'il parle à une 4G de Yas ; en réalité, rien ne quitte la pièce', b: true },
  { t: ', et aucun équipement réel n\'est touché.' },
]);

FIGURE(FIG('figR1-topologie'), 'Le banc : le poste de travail rejoue le lien d\'accès. Aucun équipement réel n\'est touché.', 210);

P([
  { t: 'Quatre ' },
  { t: 'profils', b: true },
  { t: ' (modèles de lien) sont rejoués : P1 la fibre du siège (80 mégabits, 20 millisecondes, gigue 2 ms) ; P2 la 4G de Yas (20 mégabits, 100 millisecondes, gigue 15 ms, 0,5 % de perte), le cas central ; P3 le VSAT (5 mégabits, 600 millisecondes, gigue 30 ms, 1 % de perte) ; P4 Starlink (100 mégabits, 40 millisecondes, gigue 20 ms, 0,3 % de perte).' },
]);

P([
  { t: 'Chaque essai, appelé ' },
  { t: 'événement', b: true },
  { t: ', dure trois minutes, toujours pareil : 30 secondes de calme (mesure du repos), 120 secondes de charge (un transfert massif sature le lien, les petites sondes continuent), 30 secondes de récupération.' },
]);

P([
  { t: 'Puis ' },
  { t: 'huit portes de qualité', b: true },
  { t: ' (huit vérifications automatiques, G0 à G7) contrôlent l\'essai : la cible répond-elle ? la charge a-t-elle vraiment eu lieu ? la latence est-elle physiquement possible ? le débit est-il cohérent avec le profil ? et quatre autres contrôles du même genre.' },
]);

P([
  { t: 'Le verdict des portes est impitoyable et public : sur 412 lignes mesurées, 136 ont tout passé, 214 ont échoué une porte majeure et sont « en quarantaine » (conservées mais exclues des conclusions ; les essais partiellement réussis restent exploités). Pourquoi tant d\'échecs ? Les portes ont été durcies pendant la mise au point, et chaque durcissement rejetait en bloc les essais anciens.' },
]);

P([
  { t: 'Mieux vaut peu de lignes sûres que beaucoup de lignes douteuses', b: true },
  { t: ' : des portes larges auraient rempli les archives et affaibli les conclusions.' },
]);

// ---------------------------------------------------------------- S5
H1('5. Les chiffres : ce qui a été mesuré');

P([
  { t: 'Cinq mesures, une phrase chacune. La ' },
  { t: 'médiane', b: true },
  { t: ' (la valeur du milieu : la moitié des cas font mieux, l\'autre moins bien) donne le cas typique. Le ' },
  { t: 'p95', b: true },
  { t: ' (le temps que subissent les 5 % des cas les plus lents) donne les cas pénibles : c\'est lui qui décide de la sensation de blocage.' },
]);

P([
  { t: 'Le ' },
  { t: 'respect d\'échéance', b: true },
  { t: ' (le pourcentage de petites requêtes arrivées sous un seuil fixé, ici 220 millisecondes) traduit un engagement de service : « 95 % des consultations répondent en moins de 220 ms ». Le ' },
  { t: 'QDI', b: true },
  { t: ' (l\'écart entre p95 et médiane : grand écart, lien imprévisible) mesure la régularité. Enfin, le ' },
  { t: 'coût en ariary', b: true },
  { t: ' convertit le gaspillage en argent réel : les octets jetés puis renvoyés, au tarif public du forfait.' },
]);

P([
  { t: 'Le résultat central, profil 4G, transfert massif en cours, échéance 220 millisecondes : ', b: true },
  { t: 'avec la file simple et l\'émetteur BBR, 60 % des petites requêtes arrivent à l\'heure ; avec fq_codel, 96 % ; avec CAKE, 98 %', b: true },
  { t: '. Et le débit du téléchargement massif ? Identique partout : entre 18,2 et 19,5 mégabits par seconde sur un lien de 20.' },
]);

P([
  { t: 'La protection ne coûte pas de vitesse', b: true },
  { t: ' : elle coûte quelques paquets jetés volontairement, payés en octets renvoyés.' },
]);

FIGURE(FIG('figR2-deadline'), 'Le résultat central : respect de l\'échéance à 220 ms sous charge, profil 4G. La file simple s\'effondre, les disciplines actives tiennent.', 260);

P([
  { t: 'L\'écart est massif, pas du bruit : sur les vagues antérieures, la même hiérarchie revient, seize répétitions de la file simple, douze de fq_codel, douze de CAKE, toutes concordantes. La régularité suit le même ordre : écart de 57 millisecondes sous la file simple, environ 13 sous les actives. ' },
  { t: 'Un lien régulier est un lien qu\'on peut surveiller', b: true },
  { t: ' : une alerte de latence y veut dire quelque chose.' },
]);

P([
  { t: 'Un résultat inattendu, à connaître : sur le profil 4G en montée, les essais avec CUBIC, l\'émetteur prudent, ont été mis en quarantaine. Pas pour leur latence, excellente, mais parce que sur un lien 4G qui perd naturellement 0,5 % de ses paquets, CUBIC prend chaque perte pour un accident de sa faute, ralentit, et ne tient plus la charge prévue. ' },
  { t: 'Sur un lien cellulaire, le choix de l\'émetteur pèse autant que celui de la file', b: true },
  { t: '.' },
]);

// ---------------------------------------------------------------- S6
H1('6. La preuve : des chiffres que personne ne peut modifier');

P([
  { t: 'Un jury, un directeur, un collègue demandera : « pourquoi vous croire ? ». La réponse est une chaîne, pas un argument.' },
], { noindent: true });

P([
  { t: 'Chaque essai écrit une ligne dans un fichier ' },
  { t: 'CSV', b: true },
  { t: ' (un tableau dans un fichier texte : une ligne par mesure, des colonnes fixes). Ensuite le fichier est ' },
  { t: 'gelé', b: true },
  { t: ' : il ne bouge plus. Un ' },
  { t: 'manifeste', b: true },
  { t: ' (la liste officielle des fichiers de la campagne) enregistre pour chaque fichier son ' },
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

P([
  { t: 'La chaîne complète d\'un chiffre, de bout en bout : la sonde mesure ; l\'essai franchit les huit portes ; la ligne s\'écrit dans le CSV ; le manifeste prend l\'empreinte du fichier ; la figure se régénère depuis le fichier. ' },
  { t: 'N\'importe qui, avec le dépôt, peut tout recalculer', b: true },
  { t: '.' },
]);

// ---------------------------------------------------------------- S7
H1('7. L\'outil en cinq gestes');

P([
  { t: 'Le tableau de bord s\'ouvre dans un navigateur (une page web servie par l\'outil lui-même, pas de site distant). Quatre panneaux dans la barre du haut. Le parcours suit l\'ordre naturel d\'une vraie session : ' },
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
  { t: '. Terminal : cgo audit --link-type 5g --site "Site X" --duration 60.' },
]);

H2('Geste 3 : lancer ta première campagne');

P([
  { t: 'Vue Campagne : coche P2, choisis les répétitions et l\'échéance, puis ' },
  { t: 'deux confirmations', b: true },
  { t: ' (la campagne occupe le poste des dizaines de minutes : rien ne démarre par un clic raté). Bascule sur Tableau live : cartes de latence et de débit, et le ' },
  { t: 'mur de comparaison', b: true },
  { t: ' (la référence d\'un côté, la configuration testée de l\'autre, la même échelle).' },
]);

P([
  { t: 'L\'écart entre les deux côtés du mur, c\'est ton résultat en train de naître', b: true },
  { t: '. Terminal : cgo run --profiles P2 --reps 1. Coupure de courant ? Relance la même commande : la campagne reprend où elle s\'était arrêtée.' },
]);

FIGURE(FIG('shot-live'), 'Le tableau live : cartes, mur de comparaison, progression de l\'essai courant.', 200);

H2('Geste 4 : lire et comparer');

P([
  { t: 'Vue Résultats : le classement des ' },
  { t: 'cellules', b: true },
  { t: ' (une cellule = une combinaison profil + discipline + émetteur + sens de circulation). Épingle une cellule A, une cellule B : la comparaison affiche un ' },
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
  { t: 'Vue Provenance : les comptes (153 campagnes, 136 lignes valides, 214 en quarantaine) et l\'empreinte du dernier gel. Le bouton ' },
  { t: 'Vérifier manifestes', b: true },
  { t: ' recalcule les empreintes sur place. Le lien ' },
  { t: 'Rapport MD', b: true },
  { t: ' télécharge le constat complet, prêt à joindre à un courriel de direction. Le bouton Exporter CSV donne le tableur, et le script tc rejouable (aqm-recipe.sh) se récupère à l\'adresse /api/report/export?format=sh. Terminal : cgo verify, cgo figures, cgo kit backup (le kit compte dix-huit commandes, de doctor à backup).' },
]);

FIGURE(FIG('shot-integrite'), 'La vue Provenance : comptes des archives, empreinte du dernier gel, vérification des manifestes.', 200);

// ---------------------------------------------------------------- S8
H1('8. Les cinq chiffres à retenir');

P([
  { t: 'Si tu ne retiens que cinq nombres, que ce soient ceux-là. Chacun arrive avec la question qu\'on te posera.' },
], { noindent: true });

P([
  { t: '60 % contre 98 %', b: true },
  { t: ' : la file simple contre CAKE, petites requêtes à l\'heure sous charge, profil 4G. Question attendue : « où est le gain ? » Réponse : dans la latence sauvée, à débit égal.' },
]);

P([
  { t: '57 millisecondes vers 13', b: true },
  { t: ' : l\'écart de régularité (QDI) entre file simple et discipline active. Question : « pourquoi pas juste la latence ? » Réponse : un lien régulier est un lien qu\'on peut surveiller avec des seuils.' },
]);

P([
  { t: '25 000 ariary = 4,5 gigaoctets', b: true },
  { t: ', soit 5 556 ariary le gigaoctet : le plus mauvais rapport du catalogue, le journalier étant cinq fois moins cher. Question : « d\'où sortent vos tarifs ? » Réponse : des grilles publiques des opérateurs de 2026.' },
]);

P([
  { t: '220 millisecondes', b: true },
  { t: ' : l\'échéance du profil 4G, calée sur le cas médian (100 ms) plus la marge de file. Question : « pourquoi ce seuil ? » Réponse : c\'est là que les disciplines se séparent franchement.' },
]);

P([
  { t: '153 campagnes, 412 lignes, 136 valides, 214 en quarantaine', b: true },
  { t: ', le tout scellé par SHA-256. Question : « vos chiffres sont-ils vérifiables ? » Réponse : chaque figure se régénère depuis le gel, cgo verify recalcule tout.' },
]);

// ---------------------------------------------------------------- S9
H1('9. Les mots du mémoire, en une ligne chacun');

P([
  { t: 'À relire avant la soutenance, pas à apprendre.' },
], { noindent: true });

P([
  { t: 'Latence', b: true },
  { t: ' : le temps de réponse, entre « j\'envoie » et « je reçois ».' },
]);
P([
  { t: 'Débit', b: true },
  { t: ' : la quantité de données par seconde. Ne dit rien du temps de réponse.' },
]);
P([
  { t: 'Bufferbloat', b: true },
  { t: ' : la file d\'attente qui gonfle sous charge ; le débit reste, la latence explose.' },
]);
P([
  { t: 'Discipline de file', b: true },
  { t: ' : la méthode de gestion de la file (pfifo_fast, fq_codel, CAKE).' },
]);
P([
  { t: 'Contrôle de congestion', b: true },
  { t: ' : la méthode de l\'émetteur pour ajuster sa vitesse (CUBIC, BBR).' },
]);
P([
  { t: 'Médiane et p95', b: true },
  { t: ' : le cas typique, et le temps que subissent les 5 % les plus lents.' },
]);
P([
  { t: 'QDI', b: true },
  { t: ' : l\'écart entre p95 et médiane ; grand, le lien est imprévisible.' },
]);
P([
  { t: 'Portes G0-G7', b: true },
  { t: ' : huit vérifications automatiques que chaque mesure doit franchir.' },
]);
P([
  { t: 'Quarantaine', b: true },
  { t: ' : les lignes qui ont échoué une porte ; conservées, comptées, exclues.' },
]);
P([
  { t: 'Gel et SHA-256', b: true },
  { t: ' : le fichier ne bouge plus, et son empreinte le prouve.' },
]);
P([
  { t: 'Événement', b: true },
  { t: ' : un essai de trois minutes : repos, charge, récupération.' },
]);
P([
  { t: 'Cellule', b: true },
  { t: ' : une combinaison profil + discipline + émetteur + sens de circulation.' },
]);

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
