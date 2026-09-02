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
      title: `Un banc d'essai et une méthode d'audit pour la protection du trafic critique sur les liens d'accès malgaches`,
      subtitle: `Météo Madagascar : disciplines de file, contrôles de congestion et coût du gaspillage en contexte prépayé`,
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
  { p: `Les liens d'accès de Météo Madagascar partagent un même canal entre trafic opérationnel critique et transferts massifs, sur des technologies allant de la fibre au satellite. Lorsqu'un transfert volumineux sature le lien, le temps de réponse des petites requêtes se dégrade sans que le débit ne baisse : c'est le bufferbloat. Les remèdes connus, disciplines de file actives et contrôles de congestion modernes, n'avaient fait l'objet d'aucune mesure locale permettant de choisir une configuration pour les équipements de l'institution. La connexion y est de surcroît vendue au volume prépayé : chaque octet retransmis à cause de la congestion est un octet payé deux fois.` },
  { p: `Ce mémoire présente la démarche qui comble ce manque. Une méthode d'audit non intrusif mesure tout lien accessible depuis un poste client autorisé, sans toucher aux équipements de production. Un banc d'essai reproductible, installé sur le même poste, reconstitue quatre classes de liens par une paire d'interfaces virtuelles, applique les disciplines et contrôles à comparer, et soumet chaque événement à huit portes de qualité avant le gel des données, vérifiable par empreintes SHA-256. Sur le profil cellulaire retenu, à échéance de 220 ms, la file simple sous BBR ne sauve que 60,2 % des requêtes critiques, contre 96,2 % sous fq_codel et 98,1 % sous CAKE ; l'écart p95 moins médiane de latence tombe de 57,4 ms à environ 13 ms, et le débit de masse reste entre 18,2 et 19,5 Mb/s. Le gaspillage, converti aux tarifs prépayés réels de 2026, se chiffre en ariary par heure selon le palier du forfait souscrit. Ces résultats fondent une recommandation opérationnelle pour les équipements sous contrôle de la direction des systèmes d'information.` },
  { keywords: `Mots-clés : bufferbloat, gestion active de file d'attente, fq_codel, CAKE, BBR, CUBIC, mesure reproductible, lien d'accès cellulaire et satellite, Madagascar, tarification prépayée.` },

  // ================================================================ Abstract
  { h0: `ABSTRACT` },
  { p: `The access links of Météo Madagascar carry both critical operational traffic and bulk transfers, over technologies ranging from fibre to satellite. When a large transfer saturates the link, the response time of small requests degrades without any drop in throughput: this is bufferbloat. Known remedies, active queue management disciplines and modern congestion controls, had never been measured locally to guide equipment configuration. Moreover, connectivity is sold as prepaid volume: every byte retransmitted because of congestion is paid for twice.` },
  { p: `This thesis presents the approach that closes this gap. A non-intrusive audit method measures any link reachable from an authorised client workstation, without touching production equipment. A reproducible test bench, installed on the same workstation, recreates four link classes through a pair of virtual interfaces, applies the disciplines and controls under comparison, and submits every event to eight quality gates before freezing the data, verifiable through SHA-256 hashes. On the selected cellular profile, at a 220 ms deadline, the simple queue under BBR saves only 60.2 % of critical requests, against 96.2 % under fq_codel and 98.1 % under CAKE; the p95 minus median latency spread drops from 57.4 ms to about 13 ms, and bulk throughput stays between 17.8 and 19.6 Mbps. Waste, converted at real 2026 prepaid tariffs, is priced in ariary per hour according to the subscribed bundle tier. These results ground an operational recommendation for the equipment under the control of the information systems directorate.` },
  { keywords: `Keywords: bufferbloat, active queue management, fq_codel, CAKE, BBR, CUBIC, reproducible measurement, cellular and satellite access links, Madagascar, prepaid tariffs.` },

  // ================================================================ Sigles et abréviations
  { h0: `LISTE DES SIGLES ET ABRÉVIATIONS` },
  { sigles: [
    [`AQM`, `Active Queue Management (gestion active de file d'attente)`],
    [`ARTEC`, `Autorité de Régulation des Technologies de Communication (Madagascar)`],
    [`BBR`, `Bottleneck Bandwidth and Round-trip propagation time (contrôle de congestion)`],
    [`CAKE`, `Common Applications Kept Enhanced (discipline de file active)`],
    [`CGO`, `Congestion Ground-truth Observatory (l'instrument du mémoire)`],
    [`CSV`, `Comma-Separated Values (format de données tabulaires)`],
    [`CUBIC`, `Algorithme de contrôle de congestion par défaut de Linux`],
    [`DEM`, `Direction des Exploitations Météorologiques (Météo Madagascar)`],
    [`DGM`, `Direction Générale de la Météorologie (Météo Madagascar)`],
    [`DRDH`, `Direction des Recherches et Développement Hydrométéorologique (Météo Madagascar)`],
    [`DSI`, `Direction des Systèmes d'Information`],
    [`FTTH`, `Fibre To The Home (fibre optique jusqu'à l'abonné)`],
    [`HTTP`, `Hypertext Transfer Protocol`],
    [`ICMP`, `Internet Control Message Protocol (protocole des pings)`],
    [`IETF`, `Internet Engineering Task Force (organisme de normalisation)`],
    [`ITU-T`, `International Telecommunication Union, Telecommunication Standardization Sector (UIT-T)`],
    [`LEO`, `Low Earth Orbit (orbite terrestre basse)`],
    [`OMM`, `Organisation Météorologique Mondiale`],
    [`P95`, `95e percentile d'une distribution`],
    [`QDI`, `écart p95 moins médiane de la latence (régularité du lien)`],
    [`RTT`, `Round-Trip Time (temps d'aller-retour)`],
    [`SHA-256`, `Secure Hash Algorithm 256 bits (fonction de condensat)`],
    [`SMHN`, `Service Météorologique et Hydrologique National (ancêtre de la DGM, 1901)`],
    [`TCP`, `Transmission Control Protocol`],
    [`TUI`, `Terminal User Interface (interface en mode terminal)`],
    [`UIT`, `Union Internationale des Télécommunications`],
    [`VSAT`, `Very Small Aperture Terminal (terminal satellite géostationnaire)`],
  ] },
];
