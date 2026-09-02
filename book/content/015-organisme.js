// Partie I : Présentation de l'organisme d'accueil (Météo Madagascar / DGM) et déroulement du stage.
// Blocks : h1 (Partie), h2 (section 1.x), h3 (1.x.y), p, fig {id, file, caption}, table {id, caption, header, rows, widths?}.
// Cross-refs inline : «{figRef:ID}» et «{tabRef:ID}» sont remplacés par des champs REF dans le builder.
module.exports = [
  { h1: `PARTIE I : JUSTIFICATION ET CONTEXTE` },

  { p: `Cette première partie fonde le choix du sujet. Elle présente d'abord l'organisme d'accueil, la Direction Générale de la Météorologie (Météo Madagascar), à travers sa fiche d'identification, son histoire, ses missions, ses activités, son organisation et son patrimoine. Elle décrit ensuite le déroulement du stage : l'arrivée et l'accueil au sein de la structure, l'intégration dans l'équipe, l'environnement de travail et les missions confiées. Les sections suivantes élargissent le cadre : le contexte technique et économique des liens d'accès malgaches, la problématique qui en découle et les objectifs du mémoire.` },

  // ================================================================ 1. Présentation de Météo Madagascar
  { h2: `Présentation de Météo Madagascar` },
  { p: `Météo Madagascar est le nom d'usage de la Direction Générale de la Météorologie (DGM), l'institution nationale faisant autorité sur toutes les questions liées au temps, au climat et à l'eau à Madagascar. Organisme public placé sous la tutelle du Ministère des Transports et de la Météorologie, elle a pour mission première la protection des personnes, des biens et des moyens de subsistance face aux phénomènes météorologiques et hydrologiques extrêmes. Fort d'environ cent vingt années d'expérience et de près de deux cents agents, l'organisme constitue la première référence du pays en matière d'informations hydrométéorologiques et climatologiques au service de la prévention des catastrophes.` },

  { h3: `Fiche d'identification` },
  { p: `Le {tabRef:fiche} rassemble les éléments d'identification de l'organisme d'accueil.` },
  { table: {
      id: 'fiche',
      caption: `Fiche d'identification de la Direction Générale de la Météorologie (Météo Madagascar).`,
      header: ['Rubrique', 'Renseignement'],
      rows: [
        ['Dénomination', 'Direction Générale de la Météorologie (DGM), dite « Météo Madagascar »'],
        ['Statut juridique', 'Organisme public'],
        ['Ministère de tutelle', 'Ministère des Transports et de la Météorologie'],
        ['Domaine de compétence', 'Temps, climat et eau (météorologie, climatologie, hydrologie)'],
        ['Acte de naissance', 'Arrêté du général Gallieni du 16 février 1901 (création du Service Météorologique et Hydrologique National)'],
        ['Adhésion à l\u2019OMM', '15 décembre 1960'],
        ['Directions techniques', 'DEM (Exploitations Météorologiques) et DRDH (Recherches et Développement Hydrométéorologique)'],
        ['Effectif', 'Environ 200 agents'],
        ['Expérience', 'Environ 120 ans'],
        ['Siège social', 'BP 1254, Rue Farafaty, Ampandrianomby, 101 Antananarivo, Madagascar'],
        ['Téléphone', '+261 20 26 396 32'],
        ['Site web', 'www.meteomadagascar.mg'],
      ],
      widths: [26, 74],
  } },

  { h3: `Bref historique` },
  { p: `L'histoire de la météorologie à Madagascar remonte à 1884 environ, lorsque des missionnaires catholiques établis à Andohalo, à Antananarivo, ont commencé à relever des variables météorologiques pour des besoins agricoles. Entre la fin du XIXᵉ siècle et le début du XXᵉ, les efforts de documentation et d'obtention de renseignements météorologiques se multiplient, jusqu'à ce que le général Gallieni signe, le 16 février 1901, un arrêté considéré comme « l'acte de naissance de la météo malgache ». Cet arrêté institue le Service Météorologique et Hydrologique National (SMHN). C'est à travers ce service que Madagascar adhère, le 15 décembre 1960, à l'Organisation Météorologique Mondiale (OMM), organisation intergouvernementale fondée par la convention météorologique mondiale du 11 octobre 1947 entrée en vigueur le 23 mars 1950.` },
  { p: `Au fil des décennies, la structure a changé plusieurs fois de dénomination et de ministère de tutelle, au gré des réorganisations administratives successives : « Service de la Météorologie Agricole et Service de Prévision du Temps » (1901-1927), « Direction des Services Météorologiques de Madagascar et Dépendance » (1927-1949), « Service Météorologique de Madagascar et Dépendance » (1949-1962), puis, après l'indépendance, « Direction des Services de la Météorologie Nationale Malgache » (1962-1975), « Service de la Météorologie Nationale » (1975-1986), « Direction de la Météorologie Nationale » (1986-1989) et « Direction de la Météorologie et de l'Hydrologie » (1989-2001). Le {tabRef:histoire} retrace ces principales étapes.` },
  { table: {
      id: 'histoire',
      caption: `Principales étapes de l'évolution institutionnelle de la météorologie à Madagascar.`,
      header: ['Période', 'Dénomination / événement'],
      rows: [
        ['~1884', 'Premiers relevés par les missionnaires catholiques d\u2019Andohalo'],
        ['16 février 1901', 'Arrêté Gallieni : création du Service Météorologique et Hydrologique National (SMHN)'],
        ['1901-1927', 'Service de la Météorologie Agricole et Service de Prévision du Temps'],
        ['1927-1949', 'Direction des Services Météorologiques de Madagascar et Dépendance'],
        ['1949-1962', 'Service Météorologique de Madagascar et Dépendance'],
        ['15 décembre 1960', 'Adhésion de Madagascar à l\u2019Organisation Météorologique Mondiale (OMM)'],
        ['1962-1975', 'Direction des Services de la Météorologie Nationale Malgache'],
        ['1975-1986', 'Service de la Météorologie Nationale'],
        ['1986-1989', 'Direction de la Météorologie Nationale'],
        ['1989-2001', 'Direction de la Météorologie et de l\u2019Hydrologie'],
        ['Depuis 2002', 'Érection en Direction Générale de la Météorologie (DGM) et stabilité organisationnelle'],
      ],
      widths: [22, 78],
  } },
  { p: `L'année 2002 marque un tournant : le décret n° 2002-803 du 7 août 2002 relatif à l'organisation du Ministère des Transports et de la Météorologie inaugure la stabilité organisationnelle encore en vigueur aujourd'hui. La structure devient Direction Générale de la Météorologie et n'a plus connu, depuis, de réorganisation profonde, même si son ministère de tutelle a plusieurs fois changé d'intitulé et de secteurs rattachés (travaux publics, tourisme, transports). Les attributions actuelles découlent notamment des décrets successifs de 2005, 2006, 2007, 2011, 2016, 2019 et 2021, ce dernier (décret n° 2021-863) fixant les attributions du Ministre des Transports et de la Météorologie ainsi que l'organisation générale de son ministère.` },

  { h3: `Missions et objectifs` },
  { p: `En tant qu'institution nationale faisant autorité sur les questions liées au temps, au climat et à l'eau, la Direction Générale de la Météorologie a pour mission de couvrir les besoins de la société malagasy en matière de services météorologiques, climatologiques, hydrologiques et de gestion des ressources en eau, de services environnementaux et de résultats de recherche. Ces services concourent à quatre objectifs principaux : contribuer à la réduction des risques de catastrophes liées au temps, au climat et à l'eau et protéger les biens et les personnes ; assurer la sécurité et le fonctionnement économique des transports aériens, maritimes et terrestres ; renforcer la résilience face à la variabilité et à l'évolution du climat ; et favoriser l'exploitation durable des ressources naturelles afin de contribuer à la croissance économique du pays.` },
  { p: `Dans le cadre de ces missions, l'organisme est notamment chargé de mettre en place et d'entretenir le réseau national d'observation météorologique, climatologique et hydrologique ; de collecter, d'analyser et de traiter les données d'observation ; d'élaborer et de diffuser les prévisions du temps, les avis d'alerte pour la protection des biens et des personnes, ainsi que les informations climatiques aux échelles intra-saisonnière, saisonnière et décennale ; et de fournir des avis techniques scientifiquement fondés sur toutes les questions liées au temps, au climat et à l'eau. Cette dernière attribution fait de la DGM l'institution nationale de référence dans ces domaines et l'inscrit dans le Cadre National pour les Services Climatologiques, qui vise à fournir des informations utiles et utilisables répondant aux besoins spécifiques des usagers pour la prise de décision.` },

  { h3: `Activités et services` },
  { p: `Les activités de Météo Madagascar se déclinent en plusieurs domaines complémentaires. La météorologie opérationnelle assure l'analyse et la prévision générale du temps à courtes échéances, la prévision marine et la production d'avis d'alerte en cas d'événements extrêmes : cyclones, vents forts, fortes pluies, fortes houles. La climatologie documente le climat, suit sa variabilité et son évolution, et produit des projections à long terme. L'hydrologie établit les prévisions de crues en temps réel et gère les informations nécessaires à l'évaluation et à l'utilisation rationnelle des ressources en eau de surface du pays. À ces activités s'ajoutent l'agro-météorologie, l'assistance aux secteurs du transport, du tourisme et de l'environnement, ainsi que la recherche appliquée.` },
  { p: `Ces services s'appuient sur un réseau national d'observation hydrométéorologique, sur des moyens de télécommunication et d'informatique dédiés à la collecte et au traitement des données, et sur une chaîne de diffusion à destination des autorités, des secteurs clés et du grand public. L'organisme entretient par ailleurs des coopérations techniques et scientifiques aux niveaux national et international, notamment dans le cadre de l'OMM, et veille à la mise en œuvre des questions de normalisation dans le domaine des activités météorologiques et hydrologiques.` },

  { h3: `Organisation et organigramme` },
  { p: `La Direction Générale de la Météorologie est dirigée par un Directeur Général et structurée autour de deux directions techniques et d'un ensemble de services rattachés directement à la direction générale. La première direction technique est la Direction des Exploitations Météorologiques (DEM), en charge de l'observation, de la prévision opérationnelle et de l'hydrologie. La seconde est la Direction des Recherches et Développement Hydrométéorologique (DRDH), en charge de la recherche, de la banque de données et des applications. Les services rattachés assurent les fonctions support : secrétariat, division finance, administration et logistique (DAFL), division de la communication et des relations publiques (DIVCOMRP). La {figRef:orgDgm} présente l'organigramme général de la DGM.` },
  { fig: { id: 'orgDgm', file: 'fig-org-dgm', caption: `Organigramme général de la Direction Générale de la Météorologie (DGM).` } },
  { p: `Le stage s'est déroulé au sein de la Direction des Exploitations Météorologiques (DEM), qui constitue le cœur opérationnel de l'organisme. La DEM regroupe trois services : le Service de la Météorologie Opérationnelle (SMO), qui comprend une division de prévision opérationnelle et une division climatologique ; le Service de la Maintenance et des Installations Techniques (SMIT), qui assure le fonctionnement et l'entretien des équipements d'observation et de télécommunication ; et le Service de l'Hydrologie (SH), qui comprend une division de prévision hydrologique et une division de prévision hydrométrique. Un comptable, un secrétariat de direction et une division d'appui sont rattachés directement à la DEM. La {figRef:orgDem} détaille cette organisation.` },
  { fig: { id: 'orgDem', file: 'fig-org-dem', caption: `Organigramme de la Direction des Exploitations Météorologiques (DEM), structure d'accueil du stage.` } },
  { p: `Selon la définition de l'organisme, la DEM a pour mission de couvrir les besoins de la société malagasy en matière de sauvegarde des vies humaines et des biens face aux catastrophes naturelles, de mise en alerte de la population en cas d'événements météorologiques ou hydrologiques extrêmes, d'assistance météorologique et hydrologique aux activités et projets de développement, et d'installation, de fonctionnement et de maintenance des stations du réseau d'observation hydrométéorologique national. Ce positionnement, à la charnière de l'observation instrumentale et de la production d'alertes, fait de la DEM un cadre d'accueil pertinent pour un travail portant sur l'instrumentation et la mesure fine de phénomènes dynamiques.` },

  { h3: `Patrimoine et moyens` },
  { p: `Le patrimoine de Météo Madagascar est à la fois humain, matériel, informationnel et immatériel. Le patrimoine humain réunit près de deux cents agents, prévisionnistes, climatologues, hydrologues, techniciens de maintenance, informaticiens et personnel administratif, dont l'expérience cumulée constitue le principal actif de l'institution. Le patrimoine immobilier comprend le siège d'Ampandrianomby à Antananarivo et les implantations provinciales, chaque province étant dotée d'un service de la météorologie.` },
  { p: `Le patrimoine matériel est constitué du réseau national d'observation hydrométéorologique : stations synoptiques, climatologiques, agro-météorologiques et hydrologiques, instruments de mesure, systèmes de télécommunication et moyens de réception de données satellitaires, ainsi que des moyens informatiques de traitement et d'archivage. Le patrimoine informationnel, enfin, réside dans les archives climatologiques et hydrologiques accumulées depuis plus d'un siècle : ces longues séries de données, gérées par les services dédiés à la banque de données, constituent une ressource scientifique irremplaçable pour l'étude du climat et de sa variabilité, et sont protégées au titre des questions de normalisation et de conservation dont l'organisme a la charge.` },

  // ================================================================ 2. Descriptif du stage
  { h2: `Descriptif du stage` },

  { p: `Cette section décrit les conditions concrètes dans lesquelles le stage s'est déroulé au sein de la Direction des Exploitations Météorologiques : l'arrivée et l'accueil, l'intégration dans l'équipe, l'environnement de travail et les missions confiées, puis les apports retirés de cette expérience et leur articulation avec la problématique traitée dans la suite du mémoire.` },

  { h3: `Arrivée et accueil au sein de l'organisme` },
  { p: `L'arrivée au sein de Météo Madagascar s'est faite à la suite d'une convention de stage établie entre l'Université d'Antananarivo et la Direction Générale de la Météorologie. Dès le premier jour, l'accueil a été assuré par le responsable de la Direction des Exploitations Météorologiques, qui a présenté la structure, ses missions et son organisation, avant de désigner un encadreur professionnel chargé du suivi quotidien. Une visite des différents services de la direction, couvrant la météorologie opérationnelle, la maintenance et les installations techniques et l'hydrologie, a permis de comprendre l'enchaînement des activités, depuis l'observation instrumentale jusqu'à la production et la diffusion des prévisions et des alertes.` },
  { p: `Cette phase d'accueil s'est accompagnée d'une présentation des règles internes, des horaires, des consignes de sécurité et des outils de travail. La mise à disposition d'un poste de travail et l'accès aux ressources documentaires de la direction ont marqué le point de départ effectif du stage.` },

  { h3: `Facilité de l'intégration` },
  { p: `L'intégration a été facilitée par la disponibilité de l'encadreur professionnel et par la culture de collaboration qui caractérise les services techniques de la direction. Les échanges réguliers avec les prévisionnistes et les techniciens de maintenance ont permis de se familiariser rapidement avec le vocabulaire métier, avec les contraintes opérationnelles du travail en continu et avec les exigences de fiabilité propres à une institution productrice d'alertes. La taille humaine de l'équipe et la proximité entre les services ont favorisé une insertion rapide et un accès direct aux personnes-ressources.` },
  { p: `Les principales difficultés rencontrées au début du stage tenaient moins à l'accueil qu'à l'appropriation d'un domaine d'application nouveau et à la conciliation entre les impératifs opérationnels du service et le calendrier du travail de recherche. Ces difficultés ont été progressivement levées grâce à l'accompagnement de l'encadreur et à l'observation directe des pratiques du service.` },

  { h3: `Environnement de travail et missions confiées` },
  { p: `L'environnement de travail combine des activités opérationnelles menées en continu, observation, prévision, veille et alerte, et des activités d'appui technique et informatique. C'est dans ce dernier registre que se sont inscrites les missions confiées : observer la chaîne d'acquisition, de transmission et de traitement des données, comprendre les contraintes de fiabilité et de continuité de service imposées par la mission d'alerte, et identifier les points où la qualité des liaisons de communication conditionne la ponctualité et l'intégrité de l'information produite.` },
  { p: `Ces observations ont nourri directement la réflexion menée dans ce mémoire. La dépendance des activités de la direction à des liaisons de communication fiables, dans un contexte national marqué par une connectivité contrainte et des coupures d'alimentation fréquentes, a mis en évidence l'intérêt d'un instrument capable de caractériser rigoureusement la congestion des liens d'accès et de quantifier son coût. Le stage a ainsi fourni le cadre applicatif et la motivation concrète du travail de conception et d'expérimentation présenté dans les parties suivantes.` },

  { h3: `Apports du stage et lien avec le mémoire` },
  { p: `Sur le plan professionnel, le stage a permis de découvrir le fonctionnement d'une institution scientifique et technique nationale, d'observer l'articulation entre production opérationnelle et exigences de fiabilité, et de mesurer l'importance des infrastructures de communication dans la chaîne de valeur de l'information hydrométéorologique. Sur le plan technique, il a offert un terrain d'observation des contraintes réelles : variabilité des liaisons, coupures d'alimentation, coût de la donnée, qui structurent la problématique du mémoire.` },
  { p: `Le lien avec le travail de recherche est direct : les exigences de robustesse et de fiabilité observées au sein de la direction se retrouvent dans les choix de conception de l'observatoire CGO, qui doit survivre aux interruptions et produire des mesures vérifiables. La partie suivante développe le contexte technique de ce travail, les liens d'accès malgaches, leurs contraintes et le cadre qui les régit, avant l'exposé de la méthodologie et de l'instrument.` },
];
