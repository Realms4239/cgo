// 02-partie1 — Partie I : Justification et contexte (15-18 p.)
module.exports = [
  { pagebreak: true },
  { h0: 'PARTIE I : JUSTIFICATION ET CONTEXTE' },
  { p: 'Cette première partie fonde le choix du sujet. Elle présente l\'institution d\'accueil et son organisation, décrit le contexte des liens d\'accès à Madagascar du point de vue technique et économique, formule la problématique à partir de ce contexte, et fixe les objectifs du travail.' },

  // ---------------- 1.1 présentation institution
  { h1: 'I.1. Présentation de l\'institution' },
  { h2: 'I.1.1 Fiche d\'identification' },
  { p: 'La Direction générale de la météorologie (Météo Madagascar) est l\'établissement public chargé de l\'observation, de la prévision et de la diffusion de l\'information météorologique à Madagascar. Elle relève du ministère chargé des transports et de la météorologie, exploite le réseau national d\'observation et représente le pays auprès de l\'Organisation météorologique mondiale. Ses installations couvrent le territoire national : siège et services centraux à Antananarivo, directions régionales dans les chefs-lieux de région, stations d\'observation réparties sur l\'ensemble du territoire.' },
  { p: 'La direction générale s\'organise en directions techniques et transversales. La direction de la météorologie et de l\'hydrologie conduit les activités d\'observation et de prévision. La direction des systèmes d\'information, dont relèvent les réseaux et les applications opérationnelles, est le demandeur direct de ce travail : elle exploite les liens d\'accès qui transportent les données présentées dans ce mémoire.' },
  { fig: { file: 'fig-org-dem', caption: 'Organigramme de la direction des exploitations météorologiques', screenshot: false } },
  { p: 'Le stage s\'est déroulé au sein de cette organisation, dans l\'équipe chargée des systèmes d\'information. Le sujet est né d\'un besoin exprimé par cette direction : disposer de mesures locales pour décider de la configuration des équipements d\'accès, alors que les décisions se prenaient jusqu\'ici sans données mesurées.' },

  { h2: 'I.1.2 Missions et activités' },
  { p: 'La mission de l\'institution couvre l\'observation météorologique, la prévision et l\'émission d\'avis d\'alerte. Ces activités produisent deux familles de trafic qui se partagent les mêmes liens d\'accès. Le premier correspond à des échanges courts et fréquents : envoi des relevés des stations automatiques, consultation des tableaux de bord par les prévisionnistes, alertes qui doivent arriver dans des délais courts. Le second correspond aux transferts volumineux : images satellitaires reçues chaque jour, sorties des modèles numériques de prévision, sauvegardes. Les premiers exigent une réponse rapide ; les seconds exigent du volume, et aucun des deux ne peut attendre la fin de l\'autre pour être utile.' },
  { p: 'Cette coexistence sur un même lien définit le problème opérationnel. Un lien reste sous-utilisé la plupart du temps, puis un téléchargement de modèle ou d\'imagerie le sature pendant des minutes ; pendant ces minutes, la consultation d\'un tableau de bord ou la remontée d\'une donnée de station peuvent passer de 100 à 300 millisecondes de temps de réponse. Le service n\'est pas coupé, il est dégradé exactement quand il sert. Le phénomène porte un nom, le bufferbloat, et dispose de remèdes documentés dans la littérature. Ce que l\'institution n\'avait pas, c\'est la mesure locale qui permet d\'arbitrer entre ces remèdes sur ses propres liens.' },

  { h2: 'I.1.2 Organisation et déroulement du stage' },
  { p: 'Le stage s\'est déroulé au sein de la direction générale de la météorologie, dans l\'équipe en charge des systèmes d\'information. Le travail s\'est déroulé en trois temps : l\'état des lieux des liens et des usages, la construction d\'un instrument de mesure adapté aux contraintes de l\'institution, puis l\'exécution des campagnes d\'essais et l\'analyse des résultats. Les responsables de la direction des systèmes d\'information ont validé le périmètre : la mesure porte sur des liens accessibles depuis un poste client, et aucun équipement de production n\'a été modifié.' },
  { fig: { file: 'fig-org-dem', caption: 'Organisation de la direction des systèmes d\'information', screenshot: false } },

  { h2: 'I.2 Les liens d\'accès malgaches' },
  { h3: 'I.2.1 Les quatre classes de liens' },
  { p: 'Quatre classes de liens d\'accès structurent la connectivité institutionnelle malgache. La fibre optique, présente dans les environnements institutionnels et urbains, offre des capacités de plusieurs dizaines de mégabits par seconde avec un aller-retour de l\'ordre de 20 ms. Les réseaux mobiles 4G et 5G, opérés principalement par Yas (l\'ancien Telma), Airtel et Orange, desservent la majorité des sites avec des capacités variables selon la charge de la cellule et des délais de l\'ordre de 100 ms. Le satellite géostationnaire, ou VSAT, dessert les sites hors de portée des autres technologies : il ajoute un délai de propagation d\'environ 600 ms, imposé par la distance jusqu\'à l\'orbite géostationnaire, pour une capacité faible. Enfin Starlink, arrivé sur le marché malgache en 2024, propose une liaison par satellite en orbite basse avec un délai d\'environ 40 ms et un débit élevé, pour un abonnement mensuel rapporté à 226 000 ariary par la presse de 2025.' },
  { p: 'Ces quatre classes se traduisent en quatre profils de banc, notés P1 à P4. P1 calque la fibre institutionnelle : 80 Mbit/s, 20 ms, sans perte. P2 calque la 4G : 20 Mbit/s, 100 ms, 0,5 % de perte, avec une gigue marquée. P3 calque le VSAT : 5 Mbit/s, 600 ms, 1 % de perte. P4 calque une liaison de satellite en orbite basse : 100 Mbit/s, 40 ms, gigue forte de 20 ms. Ces valeurs sont représentatives des classes de liens plutôt que des mesures d\'un site précis ; la calibration depuis un audit de terrain reste un prolongement.' },
  { table: { id: 'profils', caption: 'Profils de liens utilisés dans le banc d\'essai', header: ['Profil', 'Classe de lien', 'Capacité', 'Délai (RTT)', 'Gigue', 'Perte'],
    widths: [12, 30, 14, 14, 14, 28],
    rows: [
      ['P1', 'Fibre', '80 Mbit/s', '20 ms', '2 ms', '0 %'],
      ['P2', '4G/5G cellulaire', '20 Mbit/s', '100 ms', '15 ms', '0,5 %'],
      ['P3', 'VSAT géostationnaire', '5 Mbit/s', '600 ms', '30 ms', '1 %'],
      ['P4', 'Starlink (orbite basse)', '100 Mbit/s', '40 ms', '20 ms', '0,3 %'],
    ] } },
  { p: 'Le profil P2 concentre l\'étude : il correspond au cas d\'usage le plus fréquent hors du siège, un lien cellulaire partagé entre trafic de consultation et transferts. Les trois autres profils servent de bornes : un lien stable en référence, un lien très contraint, et l\'alternative satellite récente.' },

  { h2: 'I.2.2 La contrainte économique : la donnée prépayée' },
  { p: 'La spécificité malgache la plus déterminante pour ce travail est le modèle de tarification. La connexion y est le plus souvent vendue par paquets prépayés : un volume, une durée, un prix. Chez l\'opérateur principal, 25 000 ariary achètent 4,5 gigaoctets pour trente jours, soit un prix du gigaoctet cinq fois supérieur à celui du forfait de 100 gigaoctets du même opérateur. Le quotidien le plus petit, un gigaoctet pour 24 heures à 1 000 ariary, offre paradoxalement le meilleur rapport au gigaoctet du catalogue. La fibre fixe commence à 49 000 ariary pour 100 gigaoctets mensuels, soit environ dix fois moins cher au gigaoctet que le forfait cellulaire de référence.' },
  { fig: { file: 'figR4-tarifs', caption: 'Prix du gigaoctet selon le forfait, opérateur principal, 2026' } },
  { p: 'Cette structure de prix donne un sens économique direct à chaque paquet retransmis inutilement. En octobre 2025, l\'autorité de régulation a convoqué les opérateurs à la suite d\'plaintes publiques sur le coût de la donnée ; le débat public porte précisément sur la valeur d\'un mégaoctet. Un mémoire qui mesure le gaspillage en octets et le convertit en ariary à l\'heure produit donc une information lisible par la direction financière autant que par les ingénieurs réseau.' },

  { h2: 'I.2.3 Cadre réglementaire et limites d\'accès' },
  { p: 'Le cadre réglementaire malgache laisse les prix libres : la loi 2005-023 ne donne pas à l\'autorité de régulation le pouvoir de fixer les tarifs, mais l\'autorité dispose d\'un pouvoir d\'interpellation qu\'elle a exercé en octobre 2025 en convoquant les trois opérateurs après une vague de plaintes. Les taxes sectorielles représentent un enjeu public débattu dans la presse de novembre 2025, les opérateurs subordonnant toute baisse tarifaire à une révision de leur fiscalité. Ce cadre éclaire la décision méthodologique centrale du mémoire : un banc d\'essai local, qui mesure et compare sans saturer la production, s\'impose dans un contexte où la consommation de données coûte et où les équipements centraux échappent au contrôle de l\'institution.' },

  { h2: 'I.3 Problématique et questions de recherche' },
  { p: 'Les applications opérationnelles de Météo Madagascar, télémétrie, tableaux de bord, alertes, partagent leurs liens d\'accès avec des transferts de masse. Quand un tel transfert sature un lien, les files des équipements grossissent et le temps de réponse du petit trafic se dégrade, sans que le débit mesuré ne baisse : c\'est le bufferbloat. Les remèdes connus, les disciplines de file actives et les contrôles de congestion modernes, sont documentés dans la littérature mais aucun chiffre local n\'existait pour décider de leur déploiement sur les liens de l\'institution.' },
  { p: 'La problématique s\'énonce alors ainsi : comment, depuis un poste client et sans modifier l\'infrastructure de production, évaluer de manière reproductible la capacité d\'un lien d\'accès contraint à protéger le trafic critique pendant les transferts de masse, et quelles politiques de configuration recommander à partir de mesures vérifiables ?' },
  { table: {
    id: 'rq',
    caption: 'Questions de recherche et hypothèses',
    widths: [12, 48, 40],
    header: ['N°', 'Question', 'Test'],
    rows: [
      ['RQ1', 'Les liens d\'accès accessibles présentent-ils des symptômes de dégradation de latence sous charge ?', 'Audit non intrusif depuis un poste client, comparaison repos contre charge'],
      ['RQ2', 'Les disciplines actives protègent-elles le trafic critique mieux que la file simple, à débit de masse acceptable ?', 'Banc reproductible, matrice profil × discipline × contrôle de congestion, portes de qualité'],
      ['RQ3', 'Quelle configuration recommander aux équipements locaux, à quel coût ?', 'Interprétation des mesures au regard des tarifs réels et des équipements maîtrisés'],
    ] } },
  { h2: 'I.4 Objectifs' },
  { p: 'Trois objectifs découlent de ces questions. Le premier est méthodologique : disposer d\'une méthode d\'audit non intrusive applicable à tout lien accessible depuis un poste autorisé. Le deuxième est expérimental : disposer d\'un banc reproductible qui compare les politiques de gestion de file et de contrôle de congestion sur des profils de liens représentatifs du contexte, avec un gel vérifiable des données. Le troisième est opérationnel : produire des recommandations pour les équipements locaux, chiffrées en gain de latence protégée et en coût de gaspillage évité, accompagnées de la configuration réseau correspondante.' },

];
