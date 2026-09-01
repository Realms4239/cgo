// Front matter : page de titre, remerciements, résumé, abstract, sigles.
// Block format consumed by build-docx.js :
//   { titlePage: {...} }   page de titre structurée
//   { h0: 'TITRE' }        grand titre non numéroté (nouvelle page)
//   { p: 'texte' }         paragraphe justifié avec alinéa
//   { pNoIndent: 'texte' } paragraphe sans alinéa
//   { keywords: 'texte' }  ligne de mots-clés en italique
//   { sigles: [[sigle, définition], ...] } liste des abréviations
module.exports = [
  {
    titlePage: {
      university: `UNIVERSITÉ D'ANTANANARIVO`,
      faculty: `DOMAINE SCIENCES ET TECHNOLOGIES`,
      mention: `Mention : Informatique`,
      docType: `MÉMOIRE DE FIN D'ÉTUDES`,
      docSubtype: `en vue de l'obtention du Diplôme de Master en Informatique`,
      title: `Un instrument de mesure expérimentale pour la caractérisation de la congestion sur les liens d'accès malgaches`,
      subtitle: `CGO : Congestion Ground-truth Observatory`,
      author: `Présenté par : [NOM DE L'AUTEUR]`,
      supervisor: `Sous la direction de : [Nom du Directeur de Mémoire]`,
      year: `Année universitaire : 2025 – 2026`,
      jury: [
        `Soutenu le [date] devant le jury composé de :`,
        `Président : [Nom]`,
        `Examinateur : [Nom]`,
        `Encadreur : [Nom]`,
      ],
    },
  },

  // ================================================================ Remerciements
  { h0: `REMERCIEMENTS` },
  { p: `Ce mémoire est le fruit d'un travail mené dans des conditions qui, sans le soutien de plusieurs personnes et institutions, n'auraient pas permis d'aboutir. Je tiens à exprimer ma reconnaissance à celles et ceux qui ont contribué, de près ou de loin, à sa réalisation.` },
  { p: `Je remercie [Nom du Directeur], directeur de ce mémoire, pour la confiance accordée, la disponibilité et les conseils qui ont guidé ce travail du premier croquis jusqu'à la rédaction finale. Sa relecture exigeante a considérablement amélioré la clarté de l'argumentation.` },
  { p: `Je remercie les membres du jury qui ont accepté d'évaluer ce travail. Leurs remarques et suggestions contribueront à en améliorer la portée.` },
  { p: `Je remercie l'équipe pédagogique de la Mention Informatique de l'Université d'Antananarivo pour la formation dispensée, et en particulier [Nom] pour les discussions sur les méthodes de mesure réseau et la reproductibilité des expériences.` },
  { p: `Sur le plan matériel, ce travail a bénéficié du soutien de [partenaire / laboratoire]. Les mesures sur lien 4G réel ont été réalisées avec une carte prépayée acquise localement. Les coupures électriques ayant ponctué la phase de rédaction ont été l'occasion de tester en conditions réelles les mécanismes de reprise décrits dans ce mémoire.` },
  { p: `Enfin, je remercie ma famille pour son soutien inconditionnel, et les communautés du logiciel libre dont les outils (Linux, Go, Node.js) ont rendu ce travail possible sans dépendre d'aucun éditeur propriétaire.` },

  // ================================================================ Résumé
  { h0: `RÉSUMÉ` },
  { p: `Ce mémoire présente CGO (Congestion Ground-truth Observatory), un instrument de mesure expérimentale destiné à caractériser la dynamique de congestion sur les liens d'accès à haut temps de propagation, à capacité variable et à coût élevé, typiques des accès 4G et satellite de la classe malgache. Le dispositif repose sur une émulation contrôlée dotée d'une vérité terrain connue par construction : un ordonnanceur rejoue des baisses de capacité programmées (échelons, rampes, effondrements diurnes) sur un lien façonné par le noyau Linux, pendant qu'un oracle extrait à 10 Hz les compteurs exacts de la file d'attente.` },
  { p: `Trois questions sont traitées : quel détecteur d'événement de congestion repère le plus tôt le début réel de la congestion à coût de faux positif égal ; avec quelle précision des sondes actives légères retrouvent la bande passante disponible d'un lien limité en débit ; et si un façonneur adaptatif déclenché par la détection fait mieux que des configurations statiques de gestion active de file. Cinq détecteurs sont implémentés : seuil fixe, gradient, EWMA, z-score sur médiane glissante, et test de Kolmogorov-Smirnov à deux échantillons. Sur 204 débuts de congestion rejoués, le test de Kolmogorov-Smirnov détecte le plus tôt (médiane de 1 444 ms contre 2 329 ms pour le seuil fixe) au prix du taux de fausses alarmes le plus élevé (5,39 %). Sur 152 événements d'atténuation, le façonneur adaptatif réduit le délai médian de file de 87,3 ms (pfifo_fast) à 13,5 ms et divise par six le gaspillage d'octets. Chaque nombre publié est vérifié par un canal indépendant (tcpdump, iperf3, ping horodaté), et chaque figure est régénérable à partir d'une archive de données gelée et hachée par SHA-256.` },
  { keywords: `Mots-clés : congestion réseau, bufferbloat, gestion active de file d'attente, détection d'événements, test de Kolmogorov-Smirnov, EWMA, mesure reproductible, lien 4G, Madagascar, bandit manchot ARMS.` },

  // ================================================================ Abstract
  { h0: `ABSTRACT` },
  { p: `This thesis presents CGO (Congestion Ground-truth Observatory), an experimental measurement instrument designed to characterise congestion dynamics on high-latency, variable-capacity, expensive access links of the Malagasy class (4G and satellite). The setup relies on controlled emulation with ground truth known by construction: a scheduler replays programmed capacity drops (steps, ramps, diurnal collapses) on a kernel-shaped link while an oracle extracts exact queue counters at 10 Hz.` },
  { p: `Three questions are addressed: which congestion-onset detector finds ground-truth onsets earliest at equal false-positive cost; how accurately lightweight active probes recover the available bandwidth of a rate-limited link; and whether a detection-triggered adaptive shaper outperforms static active queue management configurations. Five detectors are implemented: fixed threshold, gradient, EWMA, sliding-window z-score, and two-sample Kolmogorov-Smirnov test. Over 204 replayed congestion onsets, the Kolmogorov-Smirnov test detects earliest (median 1,444 ms versus 2,329 ms for the fixed threshold) at the price of the highest false-alarm rate (5.39 %). Over 152 mitigation events, the adaptive shaper reduces the median queueing delay from 87.3 ms (pfifo_fast) to 13.5 ms and divides wasted bytes by six. Every published number is cross-checked by an independent channel (tcpdump, iperf3, timestamped ping), and every figure is regenerable from a frozen, hash-chained data archive.` },
  { keywords: `Keywords: network congestion, bufferbloat, active queue management, onset detection, Kolmogorov-Smirnov test, EWMA, reproducible measurement, 4G link, Madagascar, ARMS bandit.` },

  // ================================================================ Sigles et abréviations
  { h0: `LISTE DES SIGLES ET ABRÉVIATIONS` },
  { sigles: [
    ['AQM', `Active Queue Management (gestion active de file d'attente)`],
    ['ARMS', `Adaptive Reward Multi-armed Shaper (bandit manchot du façonneur adaptatif)`],
    ['ARTEC', `Autorité de Régulation des Technologies de Communication (Madagascar)`],
    ['BBR', `Bottleneck Bandwidth and Round-trip propagation time (contrôle de congestion)`],
    ['CGO', `Congestion Ground-truth Observatory (l'observatoire proposé)`],
    ['CoDel', `Controlled Delay (algorithme de gestion active de file)`],
    ['CSV', `Comma-Separated Values (format de données tabulaires)`],
    ['CUBIC', `Algorithme de contrôle de congestion par défaut de Linux`],
    ['DEM', `Direction des Exploitations Météorologiques (Météo Madagascar)`],
    ['DGM', `Direction Générale de la Météorologie (Météo Madagascar)`],
    ['DOI', `Digital Object Identifier (identifiant pérenne)`],
    ['DRDH', `Direction des Recherches et Développement Hydrométéorologique (Météo Madagascar)`],
    ['ECDF', `Empirical Cumulative Distribution Function (fonction de répartition empirique)`],
    ['ECN', `Explicit Congestion Notification (notification explicite de congestion)`],
    ['EWMA', `Exponentially Weighted Moving Average (moyenne mobile à pondération exponentielle)`],
    ['FP', `Faux positif (fausse alarme de détection)`],
    ['IETF', `Internet Engineering Task Force (organisme de normalisation)`],
    ['IQR', `Interquartile Range (intervalle interquartile, P25 à P75)`],
    ['JIRAMA', `Jiro sy Rano Malagasy (compagnie nationale d'eau et d'électricité)`],
    ['K-S', `Kolmogorov-Smirnov (test statistique à deux échantillons)`],
    ['LEO', `Low Earth Orbit (orbite terrestre basse)`],
    ['MGA', `Malagasy Ariary (monnaie malgache, aussi notée Ar)`],
    ['MTM', `Ministère des Transports et de la Météorologie (tutelle de la DGM)`],
    ['NAT', `Network Address Translation (translation d'adresses)`],
    ['netem', `Network Emulator (émulateur de lien du noyau Linux)`],
    ['ns', `Network namespace (espace de noms réseau Linux)`],
    ['OMM', `Organisation Météorologique Mondiale`],
    ['P95', `95e percentile d'une distribution`],
    ['QDI', `Queue Delay Increase (augmentation du délai de file, métrique principale de E3)`],
    ['qdisc', `Queueing discipline (discipline de file du noyau Linux)`],
    ['RQ', `Research Question (question de recherche)`],
    ['RTT', `Round-Trip Time (temps d'aller-retour)`],
    ['SHA-256', `Secure Hash Algorithm 256 bits (fonction de condensat)`],
    ['SMHN', `Service Météorologique et Hydrologique National (ancêtre de la DGM, 1901)`],
    ['SQM', `Smart Queue Management (gestion intelligente de file, OpenWrt)`],
    ['TBF', `Token Bucket Filter (limiteur de débit à seau de jetons)`],
    ['TCP', `Transmission Control Protocol`],
    ['TTB', `Time To Baseline (temps de retour à la référence)`],
    ['UDP', `User Datagram Protocol`],
    ['veth', `Virtual Ethernet (paire d'interfaces virtuelles Linux)`],
    ['VoIP', `Voice over IP (voix sur IP)`],
    ['VSAT', `Very Small Aperture Terminal (terminal satellite)`],
    ['YAML', `YAML Ain't Markup Language (format de configuration)`],
  ] },
];
