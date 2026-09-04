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
      university: `[[NOM DE L'UNIVERSITÉ]]`,
      faculty: `[[DOMAINE / FACULTÉ]]`,
      mention: `[[Mention : FILIÈRE]]`,
      docType: `MÉMOIRE DE FIN D'ÉTUDES`,
      docSubtype: `en vue de l'obtention du Diplôme de Master en Informatique`,
      title: `Protection du trafic critique sur les liens d'accès contraints : banc d'essai reproductible, méthode d'audit et recommandations opérationnelles`,
      subtitle: `Cas des liens d'accès malgaches prépayés — Météo Madagascar : disciplines de file, contrôles de congestion et coût du gaspillage`,
      author: `Présenté par : [[NOM DE L'AUTEUR]]`,
      supervisor: `Sous la direction de : [[NOM DU DIRECTEUR DE MÉMOIRE]]`,
      year: `Année universitaire : 2025 – 2026`,
      jury: [
        `Soutenu le [[DATE DE SOUTENANCE]] devant le jury composé de :`,
        `Président : [[NOM DU PRÉSIDENT]]`,
        `Examinateur : [[NOM DE L'EXAMINATEUR]]`,
        `Encadreur : [[NOM DE L'ENCADREUR]]`,
      ],
    },
  },

  // ================================================================ Remerciements
  { h0: `REMERCIEMENTS` },
  { p: `Ce mémoire est le fruit d'un travail mené dans des conditions qui, sans le soutien de plusieurs personnes et institutions, n'auraient pas permis d'aboutir. Je tiens à exprimer ma reconnaissance à celles et ceux qui ont contribué, de près ou de loin, à sa réalisation.` },
  { p: `Je remercie [[NOM DU DIRECTEUR]], directeur de ce mémoire, pour la confiance accordée, la disponibilité et les conseils qui ont guidé ce travail du premier croquis jusqu'à la rédaction finale. Sa relecture exigeante a considérablement amélioré la clarté de l'argumentation.` },
  { p: `Je remercie les membres du jury qui ont accepté d'évaluer ce travail : [[NOM DU PRÉSIDENT]], président, [[NOM DE L'EXAMINATEUR]], examinateur, et [[NOM DE L'ENCADREUR]], encadreur. Leurs remarques et suggestions contribueront à en améliorer la portée.` },
  { p: `Je remercie l'équipe pédagogique de [[NOM DE L'UNIVERSITÉ]] pour la formation dispensée, et en particulier [[NOM]] pour les discussions sur les méthodes de mesure réseau et la reproductibilité des expériences.` },
  { p: `Je remercie [[NOM DE L'ENCADREUR PROFESSIONNEL]], encadreur professionnel au sein de la Direction des Exploitations Météorologiques de Météo Madagascar, pour l'accueil, la disponibilité quotidienne et la transmission du métier. Je remercie [[NOM DU RESPONSABLE DEM]] ainsi que les équipes des services [[NOMS DES SERVICES]] pour leur collaboration.` },
  { p: `Sur le plan matériel, ce travail a bénéficié du soutien de [[PARTENAIRE / LABORATOIRE]]. Les campagnes présentées ont été conduites sur un banc reconstituant les classes de liens de l'institution, à partir des conditions publiées par les opérateurs ; l'audit des liens réels de production reste une perspective décrite dans ce mémoire. Les coupures électriques ayant ponctué la phase de mesure ont été l'occasion d'éprouver, en conditions réelles, les mécanismes de reprise décrits dans ces pages.` },
  { p: `Enfin, je remercie [[NOMS DES PROCHES — FAMILLE]] pour leur soutien inconditionnel, et les communautés du logiciel libre dont les outils (Linux, Go, Node.js) ont rendu ce travail possible sans dépendre d'aucun éditeur propriétaire.` },

  // ================================================================ Résumé
  { h0: `RÉSUMÉ` },
  { p: `Les liens d'accès de Météo Madagascar partagent un même canal entre trafic opérationnel critique et transferts massifs, sur des technologies allant de la fibre au satellite. Lorsqu'un transfert volumineux sature le lien, le temps de réponse des petites requêtes se dégrade sans que le débit ne baisse : c'est le bufferbloat. Les remèdes connus, disciplines de file actives et contrôles de congestion modernes, n'avaient fait l'objet d'aucune mesure locale permettant de choisir une configuration pour les équipements de l'institution. La connexion y est de surcroît vendue au volume prépayé : chaque octet retransmis à cause de la congestion est un octet payé deux fois.` },
  { p: `La démarche présentée comble ce manque en deux volets complémentaires. Une méthode d'audit non intrusif mesure tout lien accessible depuis un poste client autorisé, sans toucher aux équipements de production. Un banc d'essai reproductible, installé sur le même poste, reconstitue quatre classes de liens par une paire d'interfaces virtuelles, applique les disciplines et contrôles à comparer, et soumet chaque événement à huit portes de qualité avant le gel des données, vérifiable par empreintes SHA-256. Sur le profil cellulaire retenu, la file simple ne sauve qu'à peine trois requêtes critiques sur cinq, quand les disciplines actives en sauvent la quasi-totalité ; l'irrégularité du lien est divisée par quatre et le débit de masse reste préservé. Le gaspillage, converti aux tarifs prépayés réels, se chiffre en quelques centaines d'ariary par heure de charge selon le palier du forfait souscrit. Ces résultats fondent une recommandation opérationnelle différenciée par classe de lien, applicable sur les équipements sous contrôle de la direction des systèmes d'information.` },
  { keywords: `Mots-clés : bufferbloat, gestion active de file d'attente, fq_codel, CAKE, BBR, CUBIC, mesure reproductible, lien d'accès cellulaire et satellite, Madagascar, tarification prépayée.` },

  // ================================================================ Abstract
  { h0: `ABSTRACT` },
  { p: `The access links of Météo Madagascar carry both critical operational traffic and bulk transfers, over technologies ranging from fibre to satellite. When a large transfer saturates the link, the response time of small requests degrades without any drop in throughput: this is bufferbloat. Known remedies, active queue management disciplines and modern congestion controls, had never been measured locally to guide equipment configuration. Moreover, connectivity is sold as prepaid volume: every byte retransmitted because of congestion is paid for twice.` },
  { p: `This thesis presents the approach that closes this gap. A non-intrusive audit method measures any link reachable from an authorised client workstation, without touching production equipment. A reproducible test bench, installed on the same workstation, recreates four link classes through a pair of virtual interfaces, applies the disciplines and controls under comparison, and submits every event to eight quality gates before freezing the data, verifiable through SHA-256 hashes. On the selected cellular profile, the simple queue saves barely three critical requests out of five, while the active disciplines save nearly all of them; link irregularity is divided by four and bulk throughput stays preserved. Waste, converted at real prepaid tariffs, is priced at a few hundred ariary per charge hour according to the subscribed bundle tier. These results ground a differentiated operational recommendation per link class, applicable on the equipment under the control of the information systems directorate.` },
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
