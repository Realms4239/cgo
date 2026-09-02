// build-l1.js : « Meteolink expliqué à un étudiant » (~25 p.).
// Style : Times New Roman 12 pt, interligne 1,5, justifié, marges 25 mm.
// Ton oral simple ; chaque terme technique défini à sa première apparition.
'use strict';

const fs = require('fs');
const path = require('path');
const {
  Document, Packer, Paragraph, TextRun, AlignmentType,
} = require('docx');

const MM = (mm) => Math.round(mm * 56.6929);
const OUT = path.join(__dirname, 'livrable1-explication.docx');

// ---------------------------------------------------------------- helpers
const para = (text, opts = {}) => new Paragraph({
  alignment: AlignmentType.JUSTIFIED,
  spacing: { line: 360, after: 120 },
  indent: opts.noindent ? undefined : { firstLine: MM(12.5) },
  children: [new TextRun({ text, size: 24, font: 'Times New Roman', bold: !!opts.bold, italics: !!opts.ital })],
});

const h1 = (text) => new Paragraph({
  spacing: { before: 360, after: 200, line: 360 },
  children: [new TextRun({ text, size: 32, bold: true, font: 'Times New Roman' })],
});

const h2 = (text) => new Paragraph({
  spacing: { before: 240, after: 140, line: 360 },
  children: [new TextRun({ text, size: 26, bold: true, font: 'Times New Roman' })],
});

// Encadré « Si le jury demande » : paragraphe italique avec filet.
const jury = (text) => [
  new Paragraph({
    spacing: { before: 120, after: 40, line: 360 },
    border: { top: { style: 'single', size: 4, color: '808080' } },
    children: [new TextRun({ text: 'Si le jury demande :', size: 24, bold: true, italics: true, font: 'Times New Roman' })],
  }),
  new Paragraph({
    spacing: { after: 120, line: 360 },
    border: { bottom: { style: 'single', size: 4, color: '808080' } },
    children: [new TextRun({ text, size: 24, italics: true, font: 'Times New Roman' })],
  }),
];

const children = [];
const P = (t, o) => children.push(para(t, o));
const H1 = (t) => children.push(h1(t));
const H2 = (t) => children.push(h2(t));
const J = (t) => children.push(...jury(t));

// ================================================================ contenu

H1('Meteolink expliqué à un étudiant');
P('Ce document présente le mémoire de bout en bout, sans prérequis en réseaux. Il se lit en une soirée pour les six premiers chapitres ; les deux derniers servent de préparation directe à la soutenance. Chaque chapitre suit le même principe : partir d\'une image familière, en tirer l\'idée technique, puis retenir ce qu\'il faut pouvoir dire au jury.', { noindent: true });

// ---------------------------------------------------------------- chap 1
H1('Chapitre 1 : Météo Madagascar et ses liens');
P('Météo Madagascar, officiellement la Direction générale de la météorologie, a pour mission d\'observer le temps, de prévoir et de diffuser des informations qui protègent la population. Cette mission repose sur des systèmes d\'information : des stations automatiques envoient leurs relevés, des tableaux de bord consultés par les départements, des modèles numériques et des images satellite à télécharger.');
P('Toutes ces activités passent par des liens d\'accès à Internet très inégaux. Le siège à Antananarivo dispose d\'une connexion en fibre optique, rapide et stable. Les directions régionales travaillent souvent en 4G, chez l\'opérateur Yas, l\'ancien Telma. Les stations les plus isolées dépendent encore du satellite géostationnaire, appelé VSAT, lent et coûteux. Une alternative récente, Starlink, arrive sur le marché malgache avec un satellite en orbite basse : plus rapide que le VSAT, avec une facture mensuelle d\'environ 226 000 ariary d\'après la presse de 2025.');
P('La contrainte la plus forte n\'est pas technique mais économique. La connexion y est vendue par paquets prépayés : 25 000 ariary achètent 4,5 gigaoctets pour un mois chez Yas. Chaque octet gaspillé, par exemple lors de retransmissions dues à la congestion, représente de l\'argent perdu. En octobre 2025, l\'autorité de régulation ARTEC a d\'ailleurs convoqué les trois opérateurs à la suite de plaintes sur le prix de la donnée. Ce contexte donne au mémoire son enjeu réel : mesurer ce qui se passe sur ces liens, comparer les remèdes possibles, et chiffrer le gaspillage en ariary.');
J('Pourquoi ce contexte compte autant ? Parce que le jury attend un mémoire ancré dans une situation d\'entreprise réelle, pas un exercice de laboratoire. La réponse : la DSI de Météo Madagascar doit décider comment configurer ses équipements avec un budget de données limité ; ce mémoire lui donne des mesures et une recommandation.');

// ---------------------------------------------------------------- chap 2
H1('Chapitre 2 : Le problème : le péage embouteillé');
P('Imaginons une autoroute avec un seul péage. Tant que le trafic est léger, chaque voiture passe vite. Quand un convoi de camions arrive, la file d\'attente s\'allonge devant le péage ; une petite voiture qui arrive derrière attend autant que les camions, même si elle seule mettrait une seconde à passer. C\'est exactement ce qui se produit sur un lien Internet chargé : le transfert d\'un gros fichier joue les camions, et les petits messages urgents attendent derrière eux.');
P('Ce phénomène s\'appelle le bufferbloat, littéralement la « boursouflure de la mémoire tampon ». Les équipements réseau disposent de tampons, des zones d\'attente pour les paquets de données. Tant que ces zones sont dimensionnées trop grand, elles se remplissent sous charge et chaque échange, même minuscule, subit l\'attente de toute la file. Une requête de télémétrie qui prend normalement 100 millisecondes peut en mettre 300 : le service semble planté alors que le débit, lui, reste correct.');
P('C\'est là l\'idée centrale du mémoire : le débit seul ne dit rien de la qualité d\'un lien pour le travail opérationnel. Ce qui compte, c\'est la latence, c\'à-dire le temps de réponse, pendant les périodes de charge. Météo Madagascar a besoin que ses alertes et ses tableaux de bord restent vifs pendant qu\'une image satellite se télécharge sur le même lien.');
J('Le jury peut demander pourquoi on parle de bufferbloat plutôt que de simple lenteur. Réponse : la lenteur constante vient du débit ; le bufferbloat est une dégradation de la latence qui apparaît sous charge, avec un débit intact. Les distinguer est précisément l\'objet de la mesure.');

// ---------------------------------------------------------------- chap 3
H1('Chapitre 3 : Les remèdes connus : la discipline de file et BBR');
H2('Trois façons de gérer la file d\'attente');
P('La discipline de file, en anglais queueing discipline, est la règle que l\'équipement applique à sa file d\'attente. Linux en propose trois principales, testées dans le mémoire. pfifo_fast, la plus simple, laisse entrer les paquets dans l\'ordre d\'arrivée : c\'est le péage sans agent, témoin de référence du bufferbloat. fq_codel et CAKE sont des disciplines actives, de la famille AQM pour gestion active de file : elles surveillent l\'attente et jettent ou marquent quelques paquets à l\'avance pour forcer les émetteurs à ralentir, ce qui maintient la file courte. CAKE, décrit dans la demande de commentaires numéro 8290, est la plus complète des trois : elle combine cette régulation avec un plafond de débit et un traitement équitable par flux.');
H2('Deux façons d\'émettre');
P('Le second levier est le contrôle de congestion, la règle que suit l\'émetteur pour adapter sa vitesse. CUBIC, présent par défaut dans Linux, détecte les pertes de paquets et en déduit la capacité : il remplit donc les files et apprend en cassant. BBR, mis au point par Google, mesure le débit disponible et l\'aller-retour minimal, puis ajuste son rythme pour tenir la file du réseau courte sans la remplir. Ces deux comportements interagissent avec la discipline de file : la question du mémoire est de savoir quelles combinaisons protègent le mieux le petit trafic de Météo Madagascar.');
J('Pourquoi tester les deux leviers ensemble ? Parce qu\'un opérateur ne choisit quasiment jamais la discipline de file des serveurs distants, mais il contrôle son routeur de site et parfois les émetteurs. Il faut donc connaître l\'effet de chaque combinaison pour recommander ce qui dépend de lui.');

// ---------------------------------------------------------------- chap 4
H1('Chapitre 4 : Le banc d\'essai installé sur le poste client');
H2('Rejouer le lien sans toucher à l\'infrastructure');
P('Le mémoire n\'a pas besoin de s\'approprier le réseau de production de l\'institution pour conduire ses essais. Le banc s\'installe entièrement sur le poste client de l\'agent : un programme unique, nommé cgo, crée une paire de liens virtuels, appelés veth, à l\'intérieur de la machine. Sur l\'un d\'eux, l\'outil ajoute un dispositif nommé netem, pour émulation réseau, qui impose au trafic le délai, la gigue, c\'est-à-dire la variation du délai, et le taux de perte du profil choisi. Par-dessus s\'applique la discipline de file testée. L\'autre extrémité héberge un petit serveur qui distribue un objet de 16 kilo-octets et reçoit le flux de charge.');
P('Le poste se fait ainsi croire qu\'il parle à un lien 4G ou à un lien satellite, alors que tout reste local. Aucun équipement de Météo Madagascar n\'est modifié, aucune donnée de production n\'est consommée : c\'est la condition pour expérimenter librement.');
H2('Quatre profils de lien');
P('Chaque essai rejoue l\'un des quatre profils retenus. P1 représente la fibre du siège : 80 mégabits par seconde, 20 millisecondes d\'aller-retour, sans perte. P2 représente la 4G de Yas, le cas d\'étude central : 20 mégabits, 100 millisecondes, 0,5 pour cent de perte. P3 représente le VSAT des stations isolées : 5 mégabits et surtout 600 millisecondes, l\'aller-retour imposé par un satellite géostationnaire placé à 36 000 kilomètres. P4, ajouté en fin de travail, représente Starlink : 100 mégabits, 40 millisecondes, avec une gigue marquée propre aux relais entre satellites en orbite basse.');
H2('Un protocole fixe et des portes de contrôle');
P('Chaque événement d\'essai suit le même déroulement : trente secondes de calme pour mesurer l\'état de repos, cent vingt secondes de charge pendant lesquelles un flux massif sature le lien, puis trente secondes de récupération. Huit portes de qualité, numérotées de zéro à sept, vérifient chaque événement : la cible répond-elle, la charge a-t-elle vraiment démarré, les sondes produisent-elles, la latence est-elle physiquement plausible, le débit cohérent, aucune ligne dupliquée, la mesure de repos stable, la machine pas saturée. Un événement qui échoue une porte majeure est mis en quarantaine : sa ligne est conservée mais exclue des résultats. Rien n\'est caché, tout est compté.');
J('Le jury demande souvent pourquoi la charge ne dure pas des heures. Réponse honnête : la fenêtre de 120 secondes révèle le régime établi de chaque combinaison ; les effets de saturation prolongée et les profils sur une journée entière restent une perspective, annoncée comme telle.');

// ---------------------------------------------------------------- chap 5
H1('Chapitre 5 : Les mesures et ce qu\'elles disent');
H2('Cinq indicateurs, chacun répond à une question');
P('La médiane, notée p50, donne le temps de réponse typique ; le quatre-vingt-quinzième centile, noté p95, donne le temps que subissent les cas les plus défavorables, ceux qui décident de la sensation de blocage. L\'indicateur deadline mesure la part des petites requêtes arrivées sous l\'échéance fixée, par exemple 220 millisecondes : c\'est la traduction directe d\'un engagement de service. L\'écart entre p95 et la médiane, appelé QDI dans le mémoire, renseigne la régularité du lien : un écart faible signifie un lien prévisible. Enfin, deux indicateurs économiques transforment la mesure en langage de direction : le score R de voix, tiré du modèle E de la recommandation G.107 de l\'UIT, estime la qualité d\'un appel sur le lien observé, et le coût en ariary par heure applique au volume gaspillé les tarifs réels des opérateurs.');
H2('Ce que les essais ont montré');
P('Sur le profil 4G, avec une échéance fixée à 220 millisecondes, la file simple laisse 60 pour cent des petites requêtes arriver à l\'heure, tandis que les disciplines actives en sauvent environ 96 à 98 pour cent. L\'écart de régularité raconte la même histoire : 57 millisecondes d\'écart sous pfifo avec BBR, contre environ 13 sous CAKE. Autrement dit, la discipline active tient la file courte et le trafic critique passe pendant le téléchargement massif, au prix de quelques paquets écartés. Le débit utile reste proche de la capacité du profil, ce que traduit l\'expression « le débit est préservé ».');
P('Sur la voix, le score R confirme la hiérarchie : plus de 88 sur le profil fibre, environ 85 sur Starlink, autour de 71 sur le VSAT chargé. Le coût, calculé au palier mensuel de 4,5 gigaoctets, rend le gaspillage sensible : quelques centaines d\'ariary par heure de charge sur un lien 4G, la fibre étant environ dix fois moins chère au gigaoctet.');
J('D\'où sortent ces chiffres ? De plusieurs centaines d\'événements gelés dans des fichiers CSV, chaque ligne vérifiable par une empreinte SHA-256. Le jury peut refaire le calcul : c\'est l\'objet du chapitre suivant.');

// ---------------------------------------------------------------- chap 6
H1('Chapitre 6 : La preuve vérifiable');
P('Un résultat de mesure ne vaut que si d\'autres peuvent le contrôler. Chaque campagne d\'essais écrit ses lignes dans un fichier CSV, un tableau de valeurs à colonnes fixes, puis fige le tout : le fichier ne bouge plus, et un manifeste enregistre l\'empreinte cryptographique SHA-256 de chaque fichier, c\'est-à-dire une somme unique qui change au moindre octet modifié. Une commande, cgo verify, recalcule ces sommes et signale toute divergence. Les figures du mémoire sont régénérées uniquement depuis ces fichiers gelés : aucune valeur n\'est saisie à la main.');
P('Cette chaîne répond à la question la plus redoutable d\'une soutenance : qui peut vérifier ? Tout lecteur disposant du dépôt peut relancer la vérification et refaire les graphiques. Les événements mis en quarantaine ne disparaissent pas : ils restent dans le journal, comptés, avec la raison de leur exclusion.');
J('Le jury peut demander si le banc fausse les résultats par rapport au réseau réel. Réponse : le banc ne remplace pas la mesure de production ; il garantit la reproductibilité des comparaisons entre disciplines, toutes soumises aux mêmes conditions. La mesure du lien réel existe par ailleurs comme méthode, l\'audit, applicable depuis n\'importe quel poste.');

// ---------------------------------------------------------------- chap 7
H1('Chapitre 7 : Raconter cela au jury');
H2('Les phrases qui structurent l\'exposé');
P('L\'ouverture, à connaître par c\u0153ur : « Météo Madagascar conduit sa mission sur des liens d\'accès inégaux et prépayés. Ce mémoire mesure ce que la charge fait au trafic critique sur ces liens, compare les remèdes disponibles, et chiffre le gaspillage en ariary. » La fermeture : « Les disciplines actives de file protègent le trafic critique à coût maîtrisé ; la preuve est gelée et vérifiable. » Entre les deux, chaque diapositive porte un titre qui énonce sa conclusion, jamais son sujet : « 60 pour cent contre 98 pour cent » vaut mieux que « résultats de l\'essai numéro un ».');
H2('Les règles de prise de parole');
P('Trente à cinquante mots par diapositive, pas davantage : le détail se dit à l\'oral. Regarder chaque membre du jury, notes autorisées trois secondes au plus. Remplir à peine 80 pour cent du temps imparti, car le stress accélère. Remercier le jury à la fin, brièvement, jamais à l\'ouverture. Si une question reste sans réponse : l\'admettre et proposer une piste raisonnée ; avouer une limite vaut infiniment mieux qu\'inventer.');
H2('Trois diapositives à sacrifier si le temps manque');
P('Les diapositives sur la régularité, sur la voix et sur Starlink peuvent être sautées sans casser le fil : le résultat central et la recommandation suffisent. Elles restent en réserve, dans la série de secours, si une question les réclame.');

// ---------------------------------------------------------------- chap 8
H1('Chapitre 8 : Le kit anti-questions, armé');

P('Comment utiliser ce kit : chaque réponse suit le même squelette en quatre temps. Nommer la limite sans l\'adoucir. Borner : ce qu\'elle affecte, ce qu\'elle n\'affecte pas. Mitiger : ce qui est fait malgré elle. Refermer : la portée exacte de l\'affirmation. Chaque réponse tient en moins d\'une minute et porte au moins un chiffre exact des archives.', { noindent: true });

H2('Les quatre façons de perdre la soutenance');
P('Se braquer : le jury pousse une limite pour tester le calme, pas pour humilier ; deux secondes de silence, puis la réponse structurée. S\'excuser : « j\'aurais dû, je n\'ai pas pu » transforme une limite bornée en confession ; une limite est une information de portée, pas un regret. Inventer : un bluff se détecte en une relance et détruit la crédibilité de toutes les réponses précédentes. Tout concéder : si le jury pousse jusqu\'à « donc vous ne pouvez rien affirmer ? », la bonne réponse tient le cap : la limite réduit la portée, pas la conclusion.');
P('Une règle transversale : ne jamais présenter une contrainte comme une vertu méthodologique qu\'elle n\'était pas. Et citer la page exacte du mémoire quand on annonce une limite : cela change tout.');

H2('D\'où viennent vos données ?');
P('« Suivez un chiffre de ce tableau : décrivez exactement comment il a été produit, de la commande au fichier. » Le pipeline en quatre étapes : la sonde mesure pendant l\'événement de trois minutes ; l\'événement franchit les huit portes ; la ligne est écrite dans le CSV du run ; le manifeste du run enregistre l\'empreinte SHA-256 du fichier. Exemple : le 60,2 % de la file simple sous BBR, run 1788191429, événement 2, ligne visible dans l\'archive, empreinte vérifiable par cgo verify. Le jury peut suivre ce chemin-là, chiffre par chiffre.');

H2('Vos archives sont « gelées » : que prouve le hash exactement ?');
P('L\'empreinte prouve l\'intégrité : le fichier n\'a pas changé depuis le gel, toute modification d\'un octet serait détectée. Elle ne prouve pas la validité de la mesure : c\'est le rôle des huit portes. Les deux garanties sont complémentaires et se répondent dans cet ordre quand on les confond.');

H2('Qui a exécuté les campagnes, sur quelle machine, quand ?');
P('Les 134 runs sont horodatés par leur identifiant même : run-1788191429 commence par un horodatage Unix. La machine du banc : machine virtuelle Ubuntu, deux processeurs virtuels, et chaque événement porte sa colonne d\'utilisation processeur, zéro pour cent sur les six événements du run final : la machine n\'a pas influé. Réinstaller le banc et relancer la même campagne redonnerait des conditions identiques, c\'est l\'objet du banc.');

H2('154 quarantaines pour 83 valides : pourquoi vous croire ?');
P('Le ratio assume un choix : des portes larges auraient rempli les archives et affaibli les conclusions. Ce qui est en quarantaine : surtout les vagues de mise au point, où chaque durcissement d\'une porte rejetait en bloc les événements antérieurs, et les cellules dont le régime de charge ne s\'établissait pas. Ce que ça n\'affecte pas : les 83 lignes valides, dont chacune a passé les huit contrôles. Le compte est publié, la quarantaine est conservée à côté des valides, et le lecteur peut reconstituer le filtre.');

H2('Une répétition par cellule : pourquoi vous croire ?');
P('D\'abord nommer : la campagne finale conduit une répétition par cellule. Borner : n égale un interdit les comparaisons fines entre disciplines voisines ; il n\'interdit pas les ordres de grandeur quand l\'effet est massif : soixante contre quatre-vingt-dix-huit pour cent d\'échéances, ce n\'est pas un écart de bruit. Mitiger : les vagues antérieures comptent des répétitions, onze pour la file simple sous BBR sur le profil 4G, sept pour fq_codel, six pour CAKE, cinquante-et-une pour la cellule CUBIC de la fibre, et la fenêtre de charge elle-même agrège environ cent vingt observations de sondes : le p95 ne repose pas sur un point. Refermer : conclusions exploratoires sur les cellules à événement unique, consolidation par trois répétitions systématiques en premier travail futur.');

H2('Votre banc est émulé, pas le réseau réel : quelle valeur ?');
P('L\'émulation contrôle ce que le terrain ne permet pas : isoler la discipline de file, comparer des configurations à conditions strictement identiques, rejouer. Ce que netem ne capture pas : la couche radio et le multiplexage réel des applications, deux limites nommées dans le mémoire. La revendication exacte : des mécanismes et des ordres de grandeur relatifs, pas des valeurs absolues de site malgache. Et la mesure du réel existe par ailleurs : l\'audit, disponible depuis n\'importe quel poste autorisé, le déploiement multi-sites étant la première perspective.');

H2('Le VSAT P3 : aucune cellule valide. Vous cachez un échec ?');
P('L\'assumer avant qu\'il soit trouvé : aucune cellule valide sur le profil P3. Un flux soutenu à cinq mégabits devant six cents millisecondes d\'aller-retour ne tient pas la porte du débit cohérent ; aucune combinaison testée n\'y est parvenue. Le résultat est double : la stricte application des portes exclut ce profil, et la physique du satellite géostationnaire place le plancher de latence hors de portée de toute configuration. Le mémoire le publie tel quel, c\'est une borne du dispositif, pas un angle mort.');

H2('Pourquoi pas Flent, l\'outil de référence du domaine ?');
P('Flent mesure la latence sous charge et est l\'outil de référence des auteurs du domaine ; le principe de la mesure, ici, vient de la même école, RFC 8290. Le besoin du contexte : Flent suppose des exécutables aux deux extrémités du lien, ce qu\'un client institutionnel n\'obtient pas de son opérateur ; iperf3 quantifie la capacité mais perd la victime, la latence du trafic concurrent ; les tests grand public perdent tout sauf le chiffre. L\'outil du mémoire se place entièrement côté client, ajoute l\'échéance, la régularité, le coût en ariary et le gel vérifiable. Flent valide le socle ; cet outil l\'étend au cas malgache.');

H2('Pourquoi 120 secondes de charge, et pas 70 comme le standard RRUL ?');
P('Deux raisons, assumées : laisser à CUBIC le temps d\'atteindre son régime établi, qui met plusieurs dizaines de secondes à monter sur un lien à cent millisecondes d\'aller-retour ; et borner le volume consommé par campagne, ce que le contexte prépayé impose. Le standard de la communauté fait cinq plus soixante plus cinq secondes : la fenêtre retenue couvre le régime établi des deux contrôles testés, et la répétition des événements compense ce qu\'elle ne couvre pas des effets prolongés. La saturation d\'une journée entière reste une limite nommée.');

H2('Votre p95, sans barre d\'erreur : 40 et 55 millisecondes, est-ce différent ?');
P('Sur les cellules à répétitions multiples, l\'écart médiane contre p95 des p95 existe et est publié : la cellule CUBIC de la fibre tient une médiane de p95 à cinquante-cinq millisecondes et un p95 des p95 à sept cent quarante-six : la dispersion est visible, pas cachée. Sur les cellules à événement unique, n égale un interdit le test de dispersion : c\'est la limite nommée, et la première perspective la lève. Ce que l\'écart massif soixante contre quatre-vingt-dix-huit pour cent ne risque pas : un chevauchement de bruit.');

H2('Pourquoi p95 plutôt que moyenne, p99 ou maximum ?');
P('La moyenne noie les cas défavorables qui décident de la sensation de blocage. Le maximum capture la requête la plus malheureuse, dominée par les événements ponctuels de la machine hôte. Le p99, sur cent vingt observations, ne reposerait que sur une douzaine de points : trop maigre. Le p95 agrège environ cent vingt observations par fenêtre et capture la queue de distribution qui caractérise la discipline de file.');

H2('Le coût en ariary : facturation ou estimation ?');
P('Estimation, et le mémoire le dit : le volume gaspillé multiplié par le prix public du gigaoctet des forfaits de 2026, la fenêtre de trois minutes extrapolée à l\'heure. Les tarifs viennent des grilles publiques des opérateurs, datées et citées : vingt-cinq mille ariary pour quatre virgule cinq gigaoctets, soit cinq mille cinq cent cinquante-six ariary le gigaoctet au palier mensuel, le journalier étant à mille. Le chiffre sert à comparer des options, pas à prédire une facture. Un test automatisé verrouille le calcul : quatre virgule cinq gigaoctets gaspillés à ce palier font cinq cent mille ariary par heure.');

H2('Une ligne de vos données affiche 18 ms de RTT sur le profil 4G : physiquement impossible ?');
P('La connaître avant le jury : la cellule fq_codel CUBIC du profil 4G, une ligne, affiche un aller-retour p95 à dix-huit millisecondes alors que le profil impose cent. C\'est une ligne dégradée, passée dans un agrégat ancien, dont la médiane de cellule ne repose que sur elle : la ligne existe dans l\'archive, elle est comptée, et c\'est exactement pour cela que les agrégats des vagues anciennes se lisent avec précaution, et que la campagne finale, une répétition par cellule, portes armées, est la référence du mémoire. L\'honnêteté de l\'archive est de garder la ligne ; celle de l\'analyse est de dire laquelle se lit.');

H2('Que fait la direction des systèmes d\'information de vos résultats, lundi matin ?');
P('Une action, une cible, une trace : activer une discipline active sur le routeur du site cellulaire le plus chargé, la commande est dans le mémoire, une ligne de configuration, calée sous la capacité mesurée du lien. Coût : zéro ariary d\'équipement. Durée : une fenêtre de maintenance. Retour : cinq minutes, la commande est réversible. Et l\'instrument devient sentinelle : toute requête au-delà de l\'échéance alerte au journal. La prescription ne porte que sur les équipements maîtrisés par la direction ; le dernier kilomètre opérateur relève de la discussion avec l\'opérateur, appuyée par les mêmes mesures.');

H2('CAKE est-il toujours le meilleur choix ?');
P('Sur le profil cellulaire testé, il protège le mieux l\'échéance : quatre-vingt-dix-huit pour cent. Sur la fibre, dont la latence de base est faible, toutes les cellules tiennent l\'échéance : la file simple s\'y comporte convenablement, et fq_codel par défaut y suffit. La recommandation du mémoire est différenciée par classe de lien, pas absolue.');

H2('Vos résultats contredisent-ils la littérature ?');
P('Ils la confirment sur un contexte peu documenté : des liens prépayés à forte latence de base. La hiérarchie des disciplines actives sur la file simple est celle de la RFC 8290 et des outils du domaine ; l\'écart soixante contre quatre-vingt-dix-huit sur l\'échéance à deux cent vingt millisecondes est l\'apport local : la mesure de ce que la charge fait au trafic critique dans le contexte malgache, convertie en ariary, n\'existait nulle part.');

H2('Combien de temps pour développer l\'outil ?');
P('La conception de la mesure a précédé le développement : le protocole, les portes et les indicateurs d\'abord, l\'instrument ensuite. Le résultat : un binaire unique de treize mégaoctets, sans dépendance d\'exécution, avec ses tests par paquet, un assistant d\'installation en sept étapes et dix-huit commandes de déploiement et de contrôle.');

H2('Un outil maison plutôt qu\'existant : pourquoi ?');
P('Deux raisons techniques : les outils publics de mesure s\'exécutent sur des liens réels sans vérité contrôlée, ce qui interdit la comparaison reproductible de configurations ; et aucun n\'archive ses résultats en fichiers gelés vérifiables, prolongeables par les tarifs locaux. L\'apport : l\'échéance comme engagement de service, la régularité comme indicateur de surveillance, le coût en ariary, et la chaîne de preuve du gel.');

H2('Qui maintiendra ce code quand vous serez parti ?');
P('Les faits : le dépôt est public sous licence libre, le binaire est statique sans dépendance, chaque paquet a ses tests, la documentation couvre l\'usage, le déploiement et l\'architecture. Un technicien de la direction peut reprendre l\'installation et l\'exploitation sans l\'auteur. Le recadrage honnête : c\'est d\'abord un banc de recherche reproductible ; le gel garantit qu\'on peut rejouer et vérifier, pas qu\'un produit est prêt à dépanner.');

H2('Généralisable à tout Madagascar ?');
P('Les tendances se transportent aux liens de même classe : ce qui protège un lien cellulaire chargé protège un autre lien cellulaire chargé. Les valeurs absolues restent propres à chaque profil et se recalibrent par site : l\'audit de terrain, en première perspective, produit exactement ces profils calés.');

H2('Trois limites principales ?');
P('Première : le trafic est synthétique, un flux de charge unique ne reproduit pas le multiplexage réel des applications. Deuxième : le banc reconstitue le lien d\'accès, délai, gigue, perte, mais non la couche radio. Troisième : la fenêtre de charge de cent vingt secondes n\'atteint pas la saturation prolongée d\'une journée. Le mémoire en documente six en tout, avec leurs bornes : les trois citées ici sont celles qui touchent l\'interprétation des chiffres, les trois autres, l\'empan du dispositif.');

H2('Starlink à débit fixe : les liens réels varient sans cesse ?');
P('Juste : le profil P4 tient une capacité fixe et une gigue propre aux relais, pas la variation continue de débit minute par minute des liens cellulaires et orbitaux réels. Ce que le banc en dit : la latence de base passe de six cents à quarante millisecondes, la difficulté se déplace du délai vers la gigue, et la discipline active garde prise sur la gigue de file, pas sur celle de la constellation. La variation continue de capacité est une limite nommée, et l\'adaptation de débit en continu relève des travaux qui existent dans la communauté : c\'est une perspective, pas un acquis.');

H2('Quel résultat vous a surpris ?');
P('L\'effondrement de CUBIC sur le profil cellulaire : trois cellules mises en quarantaine, non pour leur latence, excellente, mais parce que le contrôle fondé sur la perte interprète la perte de base du lien comme une congestion, réduit sa fenêtre et ne tient plus la charge. Le résultat détourné devient un enseignement : sur un lien cellulaire à perte de base, le choix du contrôle de congestion décide autant que la discipline de file.');

H2('Que faudrait-il pour que votre conclusion principale soit fausse ?');
P('Une campagne répétée, trois répétitions systématiques par cellule, qui montrerait l\'écart d\'échéance soixante contre quatre-vingt-dix-huit pour cent se refermer ou s\'inverser sur le profil cellulaire. L\'écart actuel est massif et cohérent avec la mécanique des files : mais la réponse exacte est celle-ci, la falsification possible est nommée, et la première perspective la met à l\'épreuve.');

H2('Expliquez votre mémoire à un directeur non technique, en une minute.');
P('Nos liens téléphoniques prépayés transportent deux choses : les alertes qui sauvent des vies et les téléchargements qui remplissent les disques. Sur nos liens, les alertes attendaient derrière les téléchargements : soixante pour cent seulement arrivaient à temps. Un réglage du routeur, sans nouvel équipement, fait passer quatre-vingt-dix-huit pour cent des alertes à temps, au même prix d\'abonnement. Le réglage tient en une ligne, la preuve est archivée et vérifiable, et le gaspillage évitable se chiffre en ariary.');

H2('Phrases de dernier recours, à n\'utiliser qu\'en déplacement.');
P('« C\'est un choix de périmètre, documenté au chapitre méthodologie. » « La réponse complète figure dans le mémoire ; l\'essentiel est le suivant. » « Je n\'ai pas cette valeur sous la main ; la démarche pour l\'obtenir serait celle-ci. » « Ce point dépasse mon périmètre ; voici ce que j\'en comprends. » Chacune recadre sans esquiver ; aucune ne remplace une réponse chiffrée.');


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
