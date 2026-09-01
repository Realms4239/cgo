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
H1('Chapitre 8 : Le kit anti-questions');
P('Trente questions reviennent presque toujours. Voici les réponses, courtes et calmes. Les citations du jury sont reconstituées dans un langage plausible.', { noindent: true });
H2('Sur le contexte et le choix du sujet');
P('« Pourquoi ce sujet ? » La direction des systèmes d\'information devait arbitrer des configurations d\'équipement sans données locales ; le sujet comble ce manque.');
P('« Pourquoi pas une étude de satisfaction des utilisateurs ? » Le travail visait une mesure objective et reproductible, pas une enquête déclarative.');
P('« Quel est le lien avec la météorologie ? » Les applications météo produisent exactement le trafic mesuré : télémétrie légère et transferts massifs sur les mêmes liens.');
H2('Sur la méthode');
P('« Pourquoi quatre profils seulement ? » Ils couvrent les quatre classes de liens réellement utilisées par l\'institution ; la calibration par site reste une perspective.');
P('« Pourquoi un seul flux de charge ? » Le protocole compare des disciplines sur une charge identique et maîtrisée ; le multiplexage réaliste compliquerait l\'attribution des effets.');
P('« Les résultats dépendent-ils de votre machine ? » Une porte de qualité vérifie la charge du processeur à chaque événement, et les événements douteux sont mis en quarantaine.');
P('« Pourquoi ne pas avoir testé sur le réseau de production ? » La production d\'une institution de sécurité ne se sature pas à des fins d\'essai ; le banc garantit la comparaison reproductible.');
P('« Quelle est la durée totale d\'une campagne ? » Environ dix-huit minutes pour six événements, trois minutes par événement, cent huit minutes pour la matrice complète.');
H2('Sur les mesures');
P('« Pourquoi le p95 plutôt que la moyenne ? » La moyenne noie les cas défavorables ; le p95 décrit ce que subissent les requêtes lentes, qui décident de la sensation de blocage.');
P('« Le QDI vient d\'où ? » C\'est l\'écart entre le p95 et la médiane, une lecture classique des rapports d\'outils de mesure de latence, gardée sous un nom court dans le mémoire.');
P('« Le score de voix est-il une mesure réelle d\'appel ? » Non, c\'est l\'application du modèle E de l\'UIT aux conditions mesurées : délai, gigue, perte. La méthode est standard, le flux de voix simulé reste une perspective.');
P('« Pourquoi une échéance à 220 millisecondes ? » Elle se situe au niveau du p95 mesuré sur le profil 4G : c\'est le seuil où les disciplines se séparent franchement.');
P('« Le coût en ariary est-il une facturation ? » Non, c\'est une estimation lisible : le volume gaspillé multiplié par le prix au gigaoctet des forfaits réels de 2026, publiés par les opérateurs.');
H2('Sur les résultats');
P('« CAKE est-il toujours le meilleur ? » Sur les profils testés, il protège le mieux la latence critique ; sur le profil fibre, fq_codel peut suffire, et c\'est ce que recommande le mémoire.');
P('« Où est l\'amélioration en débit ? » Il n\'y en a pas à chercher : les disciplines actives préservent le débit tout en réduisant l\'attente, c\'est leur intérêt même.');
P('« Vos résultats contredisent-ils la littérature ? » Non, ils la confirment sur un contexte peu documenté : liens prépayés à forte latence de base.');
H2('Sur les limites, à assumer sans se dérober');
P('« Quelles sont vos trois limites principales ? » Des mesures ciblées sur la latence et le débit, pas exhaustives ; un flux de charge unique, pas un trafic multiplexé réaliste ; un banc qui reconstitue le lien mais pas la couche radio.');
P('« Pourquoi si peu de répétitions ? » Les campagnes finales ont conduit une répétition par cellule, les vagues antérieures en comptent davantage ; trois répétitions systématiques sont annoncées comme approfondissement.');
P('« Votre échantillon est-il représentatif ? » Il ne prétend pas l\'être : quatre classes de liens, choisies pour correspondre aux équipements de l\'institution.');
P('« Peut-on généraliser à tout Madagascar ? » Les tendances se transportent aux liens de même classe ; les valeurs absolues restent propres à chaque profil.');
H2('Sur la réalisation');
P('« Combien de temps pour développer l\'outil ? » L\'essentiel du stage, la conception de mesure ayant précédé le développement.');
P('« Pourquoi un outil maison plutôt qu\'existant ? » Les outils publics mesurent sur des liens réels sans vérité contrôlée et sans gel vérifiable ; le besoin était la comparaison reproductible et l\'ancrage dans les tarifs locaux.');
P('« Qui maintiendra l\'outil ? » Le code est déposé avec sa documentation, la direction des systèmes peut le reprendre.');
H2('Phrases de recadrage, en dernier recours');
P('« C\'est un choix de périmètre documenté dans la méthodologie. » « La réponse complète figure au chapitre correspondant ; l\'essentiel est le suivant. » « Je n\'ai pas cette valeur sous la main ; la démarche pour l\'obtenir serait celle-ci. » « Ce point dépasse mon périmètre, mais voici ce que j\'en comprends. »');

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
