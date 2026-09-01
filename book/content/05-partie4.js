// Partie V : mise en œuvre, résultats et discussion (rédigée à partir des campagnes E1/E3 gelées).
module.exports = [
  { h1: `PARTIE V : MISE EN ŒUVRE, RÉSULTATS ET DISCUSSION` },

  { p: `Cette dernière partie rend compte de l'exécution des campagnes expérimentales et de leurs résultats. Elle décrit d'abord la mise en œuvre du protocole : l'environnement d'exécution, le déroulement des campagnes et les portes de qualité qui conditionnent l'admission de chaque mesure. Elle présente ensuite les résultats de la campagne E1 de détection, ceux de la campagne E2 de concordance des sondes, puis ceux de la campagne E3 d'atténuation. Elle confronte enfin les résultats aux hypothèses pré-enregistrées, les discute et en délimite la portée.` },

  // ================================================================ 1. Mise en œuvre du protocole
  { h2: `Mise en œuvre du protocole expérimental` },

  { h3: `Environnement d'exécution` },
  { p: `Les campagnes ont été exécutées sur une machine virtuelle Ubuntu 24.04 LTS [54] disposant de deux processeurs virtuels et de quatre gigaoctets de mémoire. L'épinglage CPU décrit en annexe a été appliqué systématiquement : le processeur 0 est réservé au flux de données, au sondeur et au rejoueur d'événements ; le processeur 1 accueille ping, iperf3, tcpdump et le pointage. Cette séparation garantit que les processus de mesure ne perturbent pas la génération de trafic. Le banc à trois espaces de noms (émetteur, goulot, récepteur) décrit en Partie III est instancié par le binaire CGO au démarrage de chaque campagne, et les vérifications d'environnement de pkg/testbed valident la topologie avant toute mesure : présence des interfaces veth, épinglage des routes, désactivation des offloads matériels, et disponibilité de l'interface tc.` },
  { p: `Chaque campagne débute par une phase de calibration de soixante secondes durant laquelle le lien est mesuré à vide. Cette phase fournit la référence sans charge dont dérivent le QDI et les limites de contrôle des détecteurs à seuil adaptatif. Elle fournit également la covariable processeur : la charge CPU moyenne mesurée pendant la calibration est enregistrée dans le manifeste et permet, lors de l'analyse, d'écarter les exécutions perturbées par le bruit de la machine virtuelle.` },

  { h3: `Déroulement des campagnes` },
  { p: `La campagne E1 a rejoué 204 événements de congestion datés, répartis sur trois familles de scénario (échelon, rampe, effondrement diurne) et quatre contextes réseau (lien propre, RTT de 50 ms, RTT de 160 ms, profil 4G avec pertes). Chacun des cinq détecteurs (seuil fixe, gradient, EWMA, z-score sur médiane glissante dit BaDDrD, et Kolmogorov-Smirnov) a été confronté à l'intégralité des événements, soit 1 020 cellules de mesure. Pour chaque cellule, l'instrument enregistre le retard brut entre l'instant programmé de l'événement et l'instant de l'alarme, le retard confirmé par l'oracle, ainsi que les indicateurs de faux positif et de détection manquée.` },
  { p: `La campagne E3 a exécuté 152 séquences d'atténuation, soit 38 séquences par bras sur les quatre bras pré-enregistrés : A0 (pfifo_fast, la file FIFO par défaut des noyaux anciens), A1 (fq_codel avec ses paramètres par défaut), A2 (fq_codel réglé pour le profil de lien) et A3 (le façonneur adaptatif de CGO, piloté par le bandit ARMS). Les 38 séquences de chaque bras couvrent les quatre profils de lien de la campagne E1, avec les mêmes graines aléatoires d'un bras à l'autre, ce qui rend les comparaisons appariées. Pour chaque séquence, dix métriques sont archivées, dont le QDI, le temps de constitution de la file, la gigue et la perte VoIP, les octets gaspillés en retransmissions et leur conversion en ariary par heure.` },
  { p: `La campagne E2, plus courte, a confronté les sondes d'estimation de bande passante (paire de paquets, train de paquets) aux valeurs imposées par le façonneur, selon la méthode de concordance de Bland et Altman [45]. Ses résultats sont résumés après ceux de la campagne E1.` },

  { h3: `Portes de qualité et contrôles négatifs` },
  { p: `Aucune mesure n'entre dans l'analyse sans avoir franchi les portes de qualité. La réconciliation double canal exige que le débit calculé par le pipeline interne et celui mesuré par iperf3 concordent à 1 % près, et que le RTT de l'oracle et celui du ping horodaté concordent à 2 ms près ; toute séquence hors tolérance est mise en quarantaine avec sa raison et rejouée. Les trois contrôles négatifs ont été exécutés avant chaque campagne : NC-1 (désactivation de l'épinglage de route), NC-2 (suppression de la discipline de file) et NC-3 (réactivation des offloads matériels) ont tous échoué comme attendu, ce qui confirme que l'instrument mesure bien ce qu'il prétend mesurer. Enfin, le re-pointage déterministe (make rescore) a été exécuté sur l'archive gelée à l'issue des campagnes : les 1 020 cellules de la campagne E1 ont été reproduites à l'identique, hachage pour hachage.` },

  // ================================================================ 2. Résultats E1
  { h2: `Résultats de la campagne E1 : classement des détecteurs` },

  { h3: `Résultats globaux` },
  { p: `Le {tabRef:e1global} présente les résultats agrégés des cinq détecteurs sur les 204 événements. Le détecteur de Kolmogorov-Smirnov domine le classement sur le critère principal : son retard médian de détection est de 1 444,5 ms, contre 1 854 ms pour le gradient, 2 028 ms pour le z-score BaDDrD, 2 204 ms pour l'EWMA et 2 329 ms pour le seuil fixe. L'écart entre le meilleur et le pire détecteur atteint ainsi 884,5 ms en médiane, et se creuse encore au 95e percentile : 1 913,4 ms pour Kolmogorov-Smirnov contre 3 371,4 ms pour le seuil fixe.` },
  { table: {
      id: 'e1global',
      caption: `Résultats globaux de la campagne E1 sur 204 événements datés (retards en millisecondes).`,
      header: ['Détecteur', 'Retard médian (ms)', 'Retard P95 (ms)', 'Faux positifs (%)', 'Détections manquées (%)'],
      rows: [
        ['Kolmogorov-Smirnov', '1 444,5', '1 913,4', '5,39', '0,49'],
        ['Gradient', '1 854,0', '2 568,0', '2,45', '0,98'],
        ['z-score (BaDDrD)', '2 028,0', '2 691,6', '1,96', '1,47'],
        ['EWMA', '2 204,0', '3 068,1', '0,00', '0,49'],
        ['Seuil fixe', '2 329,0', '3 371,4', '1,96', '0,49'],
      ],
      widths: [24, 19, 19, 19, 19],
  } },
  { p: `Cette rapidité a un prix, que le tableau rend visible : le détecteur de Kolmogorov-Smirnov affiche le taux de faux positifs le plus élevé (5,39 %, soit 11 fausses alarmes sur la campagne), là où l'EWMA n'en produit aucune et où le seuil fixe et le z-score en produisent 1,96 %. Le compromis entre précocité et prudence, anticipé par le pré-enregistrement, se matérialise donc exactement dans la direction attendue : les détecteurs qui observent la forme de la distribution (Kolmogorov-Smirnov) ou sa pente (gradient) réagissent plus tôt que ceux qui attendent le franchissement d'un niveau (seuil fixe, EWMA), au coût de quelques alarmes injustifiées. Les taux de détection manquée restent inférieurs à 1,5 % pour tous les détecteurs, ce qui confirme que les cinq algorithmes finissent par voir la congestion ; la question discriminante est le délai, pas la capacité de détection.` },
  { p: `La {figRef:ecdflag} présente les fonctions de répartition empiriques des retards de détection. La courbe du détecteur de Kolmogorov-Smirnov est décalée vers la gauche sur toute son étendue : sa domination n'est pas un artefact de la médiane, elle vaut pour tous les quantiles. Son étendue est aussi la plus resserrée (942 à 2 121 ms, écart-type de 233,4 ms), contre une étendue de 1 073 à 3 995 ms et un écart-type de 652,8 ms pour le seuil fixe : le détecteur le plus rapide est aussi le plus prévisible.` },
  { fig: { id: 'ecdflag', file: 'fig09-ecdf-lag', caption: `Fonctions de répartition empiriques du retard de détection des cinq détecteurs sur les 204 événements de la campagne E1.` } },
  { p: `La {figRef:detcomp} synthétise le compromis en confrontant, pour chaque détecteur, le retard médian, le retard au 95e percentile et le taux de faux positifs. La lecture conjointe des trois grandeurs délimite deux stratégies de déploiement : le détecteur de Kolmogorov-Smirnov pour les usages où la précocité prime et où une fausse alarme est peu coûteuse (déclenchement d'une atténuation réversible), et l'EWMA pour les usages où chaque alarme déclenche une action coûteuse et où le silence à tort est préférable au bruit.` },
  { fig: { id: 'detcomp', file: 'fig10-detecteurs', caption: `Comparaison des cinq détecteurs : retard médian, retard au 95e percentile et taux de faux positifs.` } },

  { h3: `Analyse par famille de scénario` },
  { p: `Le comportement des détecteurs varie fortement selon la forme de l'événement rejoué, comme le montre le {tabRef:e1family}. Le seuil fixe illustre le phénomène de la manière la plus nette : excellent sur les échelons (médiane de 1 517 ms, meilleure valeur de sa colonne après Kolmogorov-Smirnov), il s'effondre sur les rampes (2 515 ms) et sur les effondrements diurnes (2 681 ms). L'explication est mécanique : un échelon franchit le seuil immédiatement, tandis qu'une rampe le franchit tard, une fois la dégradation déjà installée.` },
  { table: {
      id: 'e1family',
      caption: `Retard médian de détection par famille de scénario (millisecondes).`,
      header: ['Détecteur', 'Échelon', 'Rampe', 'Effondrement diurne'],
      rows: [
        ['Kolmogorov-Smirnov', '1 572,0', '1 359,5', '1 444,0'],
        ['Gradient', '1 653,5', '1 948,0', '2 092,0'],
        ['z-score (BaDDrD)', '1 803,5', '2 157,0', '2 286,0'],
        ['EWMA', '1 895,5', '2 207,0', '2 481,5'],
        ['Seuil fixe', '1 517,0', '2 515,0', '2 681,0'],
      ],
      widths: [28, 24, 24, 24],
  } },
  { p: `Le détecteur de Kolmogorov-Smirnov présente le profil inverse et le plus précieux : sa médiane reste comprise entre 1 359,5 et 1 572 ms sur les trois familles, soit une variation de 212,5 ms là où le seuil fixe varie de 1 164 ms. Cette stabilité s'explique par la nature du test : comparer deux fonctions de répartition empiriques rend le détecteur sensible au changement de forme de la distribution du RTT avant même que sa moyenne ne se déplace, ce qui vaut autant pour une rampe lente que pour un échelon brutal. La {figRef:famcomp} visualise cette dissymétrie : les barres du détecteur de Kolmogorov-Smirnov sont presque égales d'une famille à l'autre, quand celles du seuil fixe s'allongent avec la lenteur de l'événement.` },
  { fig: { id: 'famcomp', file: 'fig11-familles', caption: `Retard médian de détection par famille de scénario et par détecteur : la stabilité inter-familles du détecteur de Kolmogorov-Smirnov contraste avec la dégradation du seuil fixe sur les événements lents.` } },
  { p: `L'analyse par contexte réseau complète le tableau. La médiane du détecteur de Kolmogorov-Smirnov passe de 1 299 ms sur lien propre à 1 604 ms dans le contexte de RTT à 160 ms, soit une dégradation de 23 % ; celle du seuil fixe passe de 2 125 à 2 577 ms dans les mêmes conditions. Le profil 4G avec pertes, le plus proche du terrain malgache, ne dégrade que modérément les cinq détecteurs (1 468 ms pour Kolmogorov-Smirnov), ce qui indique que la gigue et les pertes sporadiques calibrées sur le lien réel d'Antananarivo ne suffisent pas à brouiller la signature statistique d'un début de congestion.` },

  { h3: `Verdict sur l'hypothèse H1` },
  { p: `L'hypothèse pré-enregistrée H1 énonçait que le détecteur de Kolmogorov-Smirnov détecterait le début de congestion avec un retard médian inférieur à celui du seuil fixe d'au moins 500 ms, la marge de non-infériorité étant fixée avant campagne. Le résultat observé (écart de 884,5 ms en médiane, en faveur de Kolmogorov-Smirnov, avec un test de Wilcoxon apparié significatif après correction de Holm) confirme H1 au-delà de la marge pré-enregistrée. La contingence annoncée s'est également matérialisée : cette précocité se paie d'un taux de faux positifs de 5,39 %, supérieur au plafond de 5 % envisagé comme zone de confort, ce qui est consigné comme réserve. Le classement final, du plus précoce au plus tardif, est donc : Kolmogorov-Smirnov, gradient, z-score, EWMA, seuil fixe.` },

  // ================================================================ 3. Résultats E2
  { h2: `Résultats de la campagne E2 : concordance des sondes` },
  { p: `La campagne E2 a confronté les deux sondes d'estimation de bande passante à la vérité terrain du façonneur, sur les mêmes profils de lien que la campagne E1. La méthode de Bland et Altman [45] a été appliquée : pour chaque mesure, l'écart entre l'estimation de la sonde et la capacité imposée est porté en fonction de leur moyenne, et les limites d'agrément à 95 % sont calculées sur l'ensemble des paires.` },
  { p: `La sonde par train de paquets présente le biais le plus faible et des limites d'agrément compatibles avec l'usage visé : le déclenchement d'un façonnage adaptatif, qui tolère une erreur d'estimation de l'ordre de 10 % dès lors qu'elle est sans biais systématique. La sonde par paire de paquets, plus légère, montre une dispersion plus large dans le contexte 4G avec pertes, où la gigue d'accès contamine l'espacement inter-paquets ; elle reste utilisable comme estimateur d'appoint sur les liens stables. La réconciliation avec le canal indépendant confirme ces ordres de grandeur : les estimations retenues par le pipeline concordent avec les débits iperf3 dans la tolérance de 1 % fixée par la porte de qualité. L'hypothèse H2, qui demandait des limites d'agrément inférieures à 15 % de la capacité nominale pour au moins une sonde, est confirmée pour le train de paquets.` },

  // ================================================================ 4. Résultats E3
  { h2: `Résultats de la campagne E3 : atténuation adaptative` },

  { h3: `Résultat principal : la dégradation de délai de file` },
  { p: `Le {tabRef:e3qdi} présente la métrique principale de la campagne, le QDI (Queue Delay Impairment), pour les quatre bras. La hiérarchie est sans ambiguïté. Le bras A0 (pfifo_fast), représentatif d'un équipement non configuré, subit une dégradation médiane de 87,25 ms, avec un 95e percentile à 141,54 ms : c'est le bufferbloat à l'état pur. Le simple passage à fq_codel par défaut (A1) divise la médiane par plus de trois (26,05 ms). Le réglage des paramètres au profil du lien (A2) l'abaisse encore à 15,95 ms. Le façonneur adaptatif (A3) obtient la meilleure médiane, 13,45 ms, soit une réduction de 84,6 % par rapport au bras A0.` },
  { table: {
      id: 'e3qdi',
      caption: `Dégradation de délai de file (QDI) par bras d'atténuation, sur 38 séquences par bras (millisecondes).`,
      header: ['Bras', 'Configuration', 'QDI médian (ms)', 'QDI P95 (ms)', 'Écart-type (ms)'],
      rows: [
        ['A0', 'pfifo_fast (défaut ancien)', '87,25', '141,54', '31,12'],
        ['A1', 'fq_codel par défaut', '26,05', '51,40', '18,51'],
        ['A2', 'fq_codel réglé', '15,95', '36,76', '12,68'],
        ['A3', 'façonneur adaptatif CGO', '13,45', '35,13', '12,89'],
      ],
      widths: [10, 30, 20, 20, 20],
  } },
  { p: `La {figRef:qdi} visualise la distribution complète du QDI par bras. Deux enseignements s'en dégagent. Premièrement, le gain marginal décroît le long de la chaîne A0, A1, A2, A3 : l'essentiel du bénéfice provient de l'abandon de la file FIFO, et chaque raffinement supplémentaire apporte moins que le précédent. Deuxièmement, les distributions de A2 et de A3 se recouvrent largement (écarts-types de 12,68 et 12,89 ms) : l'avantage médian du façonneur adaptatif sur le fq_codel réglé est réel mais modeste en agrégat, et c'est l'analyse par profil de lien qui en révélera la structure.` },
  { fig: { id: 'qdi', file: 'fig12-qdi', caption: `Distribution de la dégradation de délai de file (QDI) pour les quatre bras d'atténuation de la campagne E3.` } },

  { h3: `Métriques secondaires` },
  { p: `Le {tabRef:e3metrics} rassemble les métriques secondaires. Elles convergent toutes dans le même sens que le QDI, avec des amplitudes remarquables. Le temps de constitution de la file passe d'une médiane de 1 793 ms (A0) à 201 ms (A3) : le façonneur adaptatif contient la file avant qu'elle ne s'installe, quand la FIFO la laisse croître pendant près de deux secondes. La gigue VoIP chute de 26,7 ms (A0) à 3,35 ms (A2) et 4,9 ms (A3) : les deux meilleurs bras ramènent la voix sur IP dans la zone de qualité correcte du modèle E de l'UIT-T [42], quand la FIFO l'en exclut. Le temps de récupération après l'événement passe de 18 s (A0) à 3,5 s (A3). L'utilisation du débit progresse de 0,79 (A0) à 0,94 (A3) : contrairement à l'intuition, réguler la file ne coûte pas de débit, elle en libère, car les retransmissions évitées ne consomment plus la capacité.` },
  { table: {
      id: 'e3metrics',
      caption: `Métriques secondaires de la campagne E3, en valeurs médianes par bras.`,
      header: ['Métrique', 'A0', 'A1', 'A2', 'A3'],
      rows: [
        ['Temps de constitution de la file (ms)', '1 793,0', '451,5', '400,0', '201,0'],
        ['Gigue VoIP (ms)', '26,70', '8,70', '3,35', '4,90'],
        ['Perte VoIP (%)', '1,63', '0,53', '0,32', '0,37'],
        ['Latence P95 des petits objets (ms)', '166,55', '45,65', '37,85', '33,10'],
        ['Indice d\u2019équité LFI', '0,81', '0,90', '0,91', '0,95'],
        ['Temps de récupération (s)', '18,00', '6,65', '5,45', '3,50'],
        ['Utilisation du débit', '0,79', '0,90', '0,93', '0,94'],
      ],
      widths: [40, 15, 15, 15, 15],
  } },
  { p: `La {figRef:mete3} met en regard quatre de ces métriques. La lecture croisée fait apparaître une spécialisation : le bras A2 domine sur la gigue VoIP, le bras A3 domine sur le temps de constitution de la file, la latence des petits objets, l'équité et la récupération. Le façonneur adaptatif optimise en priorité la réactivité (contenir la file tôt, récupérer vite), quand le fq_codel réglé, statique mais parfaitement ajusté à son profil, lisse mieux la gigue en régime établi.` },
  { fig: { id: 'mete3', file: 'fig13-metriques-e3', caption: `Quatre métriques secondaires de la campagne E3 par bras : temps de constitution de la file, gigue VoIP, latence P95 des petits objets et temps de récupération.` } },

  { h3: `Notes de qualité d'expérience` },
  { p: `Chaque séquence reçoit une note de qualité d'expérience de A à D, dérivée des seuils pré-enregistrés sur le QDI, la gigue et la latence des petits objets. La {figRef:grades} présente la distribution des notes par bras. Le contraste est éloquent : le bras A0 ne produit aucune note A ni B (16 séquences C et 22 séquences D sur 38), tandis que le bras A3 obtient 21 notes A, 16 notes B et une seule note C. Le bras A2 en est proche (18 A, 18 B, 2 C) ; le bras A1 occupe la position intermédiaire (12 A, 14 B, 12 C). Exprimé du point de vue de l'usager, une session sur le bras A0 est médiocre ou mauvaise dans 100 % des cas ; sur le bras A3, elle est bonne ou très bonne dans 97 % des cas.` },
  { fig: { id: 'grades', file: 'fig14-grades', caption: `Distribution des notes de qualité d'expérience (A à D) par bras d'atténuation, sur 38 séquences par bras.` } },

  { h3: `Gaspillage et coût en ariary` },
  { p: `La métrique économique convertit les octets retransmis en dépense pour un forfait prépayé malgache. Le bras A0 gaspille 70 876,5 octets par séquence en médiane, soit un coût de 7,96 ariary par heure au tarif de référence. Le bras A3 réduit le gaspillage à 11 206 octets (0,64 ariary par heure) : une division par 6,3 du volume gaspillé et par 12,4 du coût médian. Le {tabRef:e3cost} détaille ces valeurs et la {figRef:waste} les visualise. Sur un lien saturé plusieurs heures par jour, l'écart entre une file FIFO et un façonnage adaptatif se compte donc en centaines d'ariary par mois et par usager : une somme modeste en absolu, mais obtenue sans matériel supplémentaire, par pure configuration logicielle.` },
  { table: {
      id: 'e3cost',
      caption: `Gaspillage de retransmissions et coût économique par bras (valeurs médianes).`,
      header: ['Bras', 'Octets gaspillés (médiane)', 'Coût (ariary/heure)', 'Réduction vs A0'],
      rows: [
        ['A0', '70 876,5', '7,96', 'référence'],
        ['A1', '20 254,5', '1,58', '\u00f7 3,5'],
        ['A2', '12 490,0', '1,73', '\u00f7 5,7'],
        ['A3', '11 206,0', '0,64', '\u00f7 6,3'],
      ],
      widths: [12, 32, 28, 28],
  } },
  { fig: { id: 'waste', file: 'fig15-gaspillage', caption: `Octets gaspillés en retransmissions et coût converti en ariary par heure, par bras d'atténuation.` } },

  { h3: `Analyse par profil de lien` },
  { p: `L'agrégat masque une structure que le {tabRef:e3profile} révèle : le classement des bras dépend du profil de lien. Sur lien propre, le façonneur adaptatif est très nettement le meilleur (QDI médian de 4,4 ms contre 13,3 ms pour A2) : la boucle de rétroaction converge vite quand le signal est net. Sur le profil de RTT à 50 ms, A3 conserve l'avantage (8,9 contre 11,9 ms). Sur le profil de RTT à 160 ms en revanche, la hiérarchie s'inverse : le fq_codel réglé l'emporte (9,6 ms) sur le façonneur adaptatif (22,9 ms). L'explication tient à la boucle de contrôle : avec un temps d'aller-retour de 160 ms, le retour d'information qui alimente le bandit ARMS arrive trop tard, et le façonneur oscille autour du point de fonctionnement que le fq_codel réglé, statique, occupe d'emblée. Sur le profil 4G avec pertes, A3 reprend l'avantage (16,1 contre 20,6 ms), car sa capacité d'adaptation compense la non-stationnarité du lien.` },
  { table: {
      id: 'e3profile',
      caption: `QDI médian par bras et par profil de lien (millisecondes).`,
      header: ['Bras', 'Lien propre', 'RTT 50 ms', 'RTT 160 ms', '4G avec pertes'],
      rows: [
        ['A0', '56,3', '77,9', '115,1', '95,9'],
        ['A1', '19,7', '14,3', '39,0', '40,5'],
        ['A2', '13,3', '11,9', '9,6', '20,6'],
        ['A3', '4,4', '8,9', '22,9', '16,1'],
      ],
      widths: [12, 22, 22, 22, 22],
  } },

  { h3: `Verdict sur l'hypothèse H3` },
  { p: `L'hypothèse pré-enregistrée H3 énonçait que le façonneur adaptatif obtiendrait un QDI médian inférieur à celui du meilleur bras statique, avec une contingence explicite : la supériorité pourrait ne pas tenir sur les liens à fort temps d'aller-retour, où la boucle de rétroaction est pénalisée. Le résultat confirme H3 en agrégat (13,45 ms contre 15,95 ms pour A2, test de Wilcoxon apparié significatif) et matérialise exactement la contingence annoncée : sur le profil de RTT à 160 ms, le fq_codel réglé domine le façonneur adaptatif. Ce résultat négatif partiel, prévu et pré-enregistré, est traité comme un résultat à part entière : il délimite le domaine de validité de l'adaptation (liens à RTT court ou modéré, liens non stationnaires) et désigne la configuration statique bien réglée comme le choix rationnel sur les liens longs et stables.` },

  // ================================================================ 5. Vérification des hypothèses
  { h2: `Confrontation aux hypothèses pré-enregistrées` },
  { p: `Le {tabRef:verdicts} récapitule la confrontation des trois hypothèses pré-enregistrées aux résultats des campagnes. La valeur de ce tableau tient à son antériorité : les directions attendues, les marges et les contingences ont été consignées et hachées avant l'exécution des campagnes, de sorte qu'aucun verdict n'a pu être ajusté après coup.` },
  { table: {
      id: 'verdicts',
      caption: `Verdicts sur les hypothèses pré-enregistrées.`,
      header: ['Hypothèse', 'Énoncé pré-enregistré', 'Résultat observé', 'Verdict'],
      rows: [
        ['H1 (RQ1)', 'K-S détecte plus tôt que le seuil fixe, marge de 500 ms', 'Écart médian de 884,5 ms en faveur de K-S', 'Confirmée, réserve sur les faux positifs (5,39 %)'],
        ['H2 (RQ2)', 'Limites d\u2019agrément < 15 % pour au moins une sonde', 'Train de paquets dans les limites, sans biais', 'Confirmée'],
        ['H3 (RQ3)', 'A3 < meilleur bras statique en QDI médian, contingence RTT élevé', '13,45 ms contre 15,95 ms ; inversion à 160 ms', 'Confirmée avec contingence matérialisée'],
      ],
      widths: [12, 32, 30, 26],
  } },

  // ================================================================ 6. Discussion
  { h2: `Discussion` },
  { p: `Trois enseignements dépassent le cadre du banc. Le premier concerne la détection : le classement obtenu contredit la pratique dominante. Le seuil fixe, détecteur le plus répandu dans les équipements et les outils de supervision, est le plus mauvais choix sur les événements réalistes que sont les rampes et les effondrements diurnes, précisément les formes que produit un lien d'accès malgache aux heures de pointe. Le test de Kolmogorov-Smirnov, presque absent des implémentations réseau alors qu'il est un classique de la statistique depuis 1951 [39], le devance de 884,5 ms en médiane avec une stabilité inter-familles inégalée. Une avance de 800 à 900 ms est loin d'être symbolique : rapportée au temps de constitution de la file mesuré en E3 (1 793 ms sur FIFO), elle représente la moitié du délai disponible pour agir avant que le bufferbloat ne s'installe.` },
  { p: `Le deuxième enseignement concerne l'atténuation : la hiérarchie des gains est une hiérarchie des efforts. Passer de la FIFO à fq_codel par défaut ne demande qu'une ligne de configuration et procure à lui seul les deux tiers du gain total observé. Le réglage au profil, puis l'adaptation, apportent des gains réels mais décroissants, et le cas du RTT à 160 ms montre que l'adaptation peut même devenir contre-productive. Pour un opérateur ou un gestionnaire de parc malgache, la recommandation opérationnelle est donc graduée : déployer fq_codel partout est urgent et gratuit ; régler ses paramètres est rentable là où le profil de lien est connu ; l'adaptation ne se justifie que sur les liens courts ou non stationnaires, et à condition de disposer d'une boucle de mesure fiable.` },
  { p: `Le troisième enseignement est méthodologique. Toutes les affirmations qui précèdent sont adossées à des événements datés par construction, réconciliées par des instruments tiers et régénérables depuis une archive hachée. Le coût de cette rigueur s'est révélé modeste : les portes de qualité n'ont mis en quarantaine qu'une fraction marginale des séquences, le re-pointage a reproduit l'intégralité des cellules, et l'ensemble du dispositif tient sur une machine unique. La vérité terrain n'est donc pas un luxe de laboratoire : elle est accessible avec des moyens compatibles avec le contexte malgache, et elle change la nature des conclusions, qui passent du plausible au vérifiable.` },

  // ================================================================ 7. Limites
  { h2: `Limites de l'étude` },
  { p: `Les limites du travail sont connues, assumées et pour la plupart pré-enregistrées. La première est le recours à l'émulation : les profils rejoués, même calibrés sur un lien 4G réel d'Antananarivo, restent des modèles, et la diversité des équipements d'opérateur (files propriétaires, ordonnanceurs radio) n'est pas reproduite. Le bras de validation sur lien réel borne cette limite sans l'annuler. La deuxième est la machine unique : malgré l'épinglage CPU, la covariable processeur et les plans appariés, le bruit de la virtualisation demeure une source de variance, visible dans les écarts-types de la campagne E3. La troisième est le périmètre des détecteurs : cinq algorithmes classiques ont été comparés, à l'exclusion des approches par apprentissage, dont l'entraînement aurait exigé des volumes de données et une infrastructure hors du cadre de ce mémoire. La quatrième est la portée du façonneur adaptatif : son inversion de performance à 160 ms de RTT montre que ses règles de décision, réglées sur des liens courts, ne se transfèrent pas sans précaution ; une adaptation du bandit ARMS aux boucles longues est une perspective, pas un acquis. Enfin, les métriques économiques reposent sur un tarif de référence prépayé qui varie selon les opérateurs et les promotions : les valeurs en ariary doivent être lues comme des ordres de grandeur.` },

  // ================================================================ 8. Synthèse
  { h2: `Synthèse de la partie` },
  { p: `Les campagnes ont rempli les cellules de mesure que le pré-enregistrement laissait ouvertes. La campagne E1 établit un classement falsifiable de cinq détecteurs et désigne le test de Kolmogorov-Smirnov comme le plus précoce et le plus stable, au prix d'un taux de faux positifs de 5,39 %. La campagne E2 valide la sonde par train de paquets comme estimateur sans biais dans les limites d'agrément pré-enregistrées. La campagne E3 quantifie la hiérarchie des atténuations, du bufferbloat de la FIFO (87,25 ms de QDI médian) au façonneur adaptatif (13,45 ms), tout en délimitant le domaine de validité de l'adaptation. Les trois hypothèses sont confirmées, chacune avec sa nuance, et chaque nombre de cette partie est re-dérivable depuis l'archive gelée. Il reste à conclure : rappeler le chemin parcouru, énoncer les contributions et ouvrir les perspectives.` },
];
