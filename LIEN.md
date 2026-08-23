
Ce document remplace intégralement l’ancien sujet **“Dynamique de congestion sur liens d’accès contraints : un observatoire à vérité terrain appliqué au contexte malgache”**.

Il abandonne les claims trop fragiles :

- “observatoire à vérité terrain” ;
- “indépendance totale des canaux” ;
- “validation du réel par l’émulation” ;
- classement de détecteurs comme contribution principale ;
- expérimentation adaptative complexe ;
- matrice excessive de 356 événements.

Il conserve les acquis utiles :

- instrumentation Go ;
- banc Linux `tc/netem` ;
- comparaison AQM / BBR / CUBIC ;
- archivage reproductible ;
- métriques QoS ;
- contexte malgache ;
- lien avec Météo Madagascar.

Les zones entre crochets `[...]` doivent être remplacées par vos valeurs réelles.

---

# PROTECTION DU TRAFIC OPÉRATIONNEL CRITIQUE SUR LIENS D’ACCÈS CONTRAINTS : AUDIT QoS NON INTRUSIF ET ÉVALUATION REPRODUCTIBLE DES POLITIQUES AQM ET BBR DANS LE CONTEXTE DE MÉTÉO MADAGASCAR

## Titre anglais

**Protecting critical operational traffic on constrained access links: non-intrusive QoS audit and reproducible evaluation of AQM and BBR policies in the Meteo Madagascar context**

---

# RÉSUMÉ

Ce mémoire étudie la capacité des liens d’accès utilisés dans le contexte de Météo Madagascar à transporter le trafic opérationnel critique lorsque ces liens sont congestionnés. Les liaisons concernées, notamment 4G/5G et fibre, peuvent présenter une capacité variable, une latence élevée par moments, des coûts de données non négligeables et des phénomènes de bufferbloat. Dans un environnement opérationnel, de petits messages critiques — télémétrie, requêtes de tableau de bord, alertes ou échanges interactifs — peuvent être dégradés par des transferts de masse tels que téléchargements de modèles, images satellite ou sauvegardes.

Le travail propose une approche en deux volets. Premièrement, un audit QoS non intrusif de liens accessibles depuis un poste client, sans modification de l’infrastructure de production. Deuxièmement, un banc d’essai Linux reproductible permettant de rejouer des profils de liens contraints et de comparer plusieurs politiques de gestion active de file d’attente et de contrôle de congestion : `pfifo_fast`, `fq_codel`, `CAKE`, CUBIC et BBR. Les métriques principales sont la latence p95 du trafic critique, la latence p95 de petits objets, le taux de respect d’une échéance temporelle, le débit utile de masse, les pertes, les octets gaspillés et une estimation du coût en ariary.

Les résultats permettent de produire un rapport d’audit, une matrice de comparaison des configurations et des recommandations opérationnelles pour les équipements locaux sous contrôle de l’institution. Le mémoire fournit également un outil réutilisable, sous forme de binaire Go autonome, capable de mesurer, archiver et régénérer les résultats de manière vérifiable.

**Mots-clés :** congestion réseau ; bufferbloat ; gestion active de file d’attente ; AQM ; BBR ; CUBIC ; QoS ; liens d’accès contraints ; Météo Madagascar ; mesure reproductible.

---

# ABSTRACT

This thesis studies the ability of access links used in the Meteo Madagascar context to carry critical operational traffic when those links are congested. The considered links, including 4G/5G and fiber, may exhibit variable capacity, elevated latency at times, non-negligible data costs, and bufferbloat phenomena. In an operational environment, small critical messages — telemetry, dashboard queries, alerts, or interactive exchanges — may be degraded by bulk transfers such as model downloads, satellite imagery, or backups.

The work proposes a two-part approach. First, a non-intrusive QoS audit of links accessible from a client workstation, without modifying production infrastructure. Second, a reproducible Linux testbed that replays constrained-link profiles and compares several active queue management and congestion-control policies: `pfifo_fast`, `fq_codel`, `CAKE`, CUBIC, and BBR. The main metrics are p95 latency of critical traffic, p95 latency of small objects, deadline compliance rate, bulk goodput, losses, wasted bytes, and an estimated cost in ariary.

The results produce an audit report, a configuration comparison matrix, and operational recommendations for local equipment under institutional control. The thesis also provides a reusable tool, packaged as a self-contained Go binary, capable of measuring, archiving, and regenerating results in a verifiable manner.

**Keywords:** network congestion; bufferbloat; active queue management; AQM; BBR; CUBIC; QoS; constrained access links; Meteo Madagascar; reproducible measurement.

---

# CURRICULUM VITAE

[Insérer CV : état civil, parcours universitaire, compétences systèmes et réseaux, programmation Go/Linux, mesures réseau, stages, projets, langues.]

---

# DÉDICACE

[Facultatif.]

---

# REMERCIEMENTS

Je remercie mon encadreur pour son accompagnement, les membres du jury pour l’évaluation de ce travail, ainsi que l’équipe de Météo Madagascar pour m’avoir accueilli dans le cadre de mon stage. Je remercie également les personnes ayant facilité l’accès aux environnements de test et aux liens accessibles utilisés pour les mesures.

---

# SOMMAIRE

- Introduction générale
- Partie I : Contexte, état de l’art et problématique
  - I. Contexte opérationnel et contraintes des liens d’accès
  - II. Fondements théoriques
  - III. Travaux connexes
  - IV. Problématique et questions de recherche
- Partie II : Méthodologie
  - I. Vue d’ensemble du dispositif
  - II. Audit QoS non intrusif sur liens accessibles
  - III. Banc d’essai contrôlé pour l’évaluation AQM/BBR
  - IV. Métriques
  - V. Contrôle qualité et reproductibilité
  - VI. Menaces sur la validité
- Partie III : Mise en œuvre
  - I. Environnement technique
  - II. Architecture de l’instrument
  - III. Déploiement des campagnes
  - IV. Interface de supervision
  - V. Sécurité, éthique et limites d’accès
- Partie IV : Résultats, recommandations et perspectives
  - I. Résultats de l’audit des liens accessibles
  - II. Résultats de l’évaluation AQM/BBR
  - III. Recommandations opérationnelles
  - IV. Limites
  - V. Perspectives
- Conclusion générale
- Références bibliographiques
- Glossaire
- Annexes

---

# LISTE DES ABRÉVIATIONS

| Abréviation | Signification |
|---|---|
| AQM | Active Queue Management |
| AWS | Automatic Weather Station |
| BBR | Bottleneck Bandwidth and Round-trip propagation time |
| CAKE | Common Applications Kept Enhanced |
| CC | Congestion Control |
| CSV | Comma-Separated Values |
| DSI | Direction des Systèmes d’Information |
| FP | Faux positif |
| HTTP | Hypertext Transfer Protocol |
| ICMP | Internet Control Message Protocol |
| LFI | Lightweight Fairness Index |
| MGA | Ariary malgache |
| QoS | Quality of Service |
| RTT | Round-Trip Time |
| SSE | Server-Sent Events |
| TCP | Transmission Control Protocol |
| TTB | Time To Bufferbloat |
| UDP | User Datagram Protocol |
| VM | Machine virtuelle |
| VSAT | Very Small Aperture Terminal |

---

# LISTE DES TABLEAUX

- Tableau 1. Questions de recherche et hypothèses
- Tableau 2. Profils de liens utilisés dans le banc d’essai
- Tableau 3. Matrice expérimentale AQM/BBR
- Tableau 4. Métriques principales
- Tableau 5. Métriques secondaires
- Tableau 6. Schéma de la table `link_audit.csv`
- Tableau 7. Schéma de la table `aqm_eval.csv`
- Tableau 8. Portes de qualité
- Tableau 9. Résultats d’audit des liens accessibles
- Tableau 10. Résultats de comparaison AQM/BBR
- Tableau 11. Recommandations opérationnelles

---

# LISTE DES FIGURES

- Figure 1. Architecture générale : audit réel non intrusif et banc d’essai contrôlé
- Figure 2. Trafic critique et trafic de masse sur lien contraint
- Figure 3. Chaîne de mesure : sondes, qdisc, archivage, figures
- Figure 4. Exemple de latence au repos et sous charge
- Figure 5. Exemple de comparaison p95 par configuration
- Figure 6. Exemple de compromis latence/débit
- Figure 7. Interface de supervision : console, résultats, intégrité

Les figures sont régénérées automatiquement à partir des fichiers CSV gelés par la commande `make figures`.

---

# INTRODUCTION GÉNÉRALE

## Contexte général

Madagascar demeure l’un des pays les moins connectés au monde. Une part importante de la population n’a jamais utilisé Internet, et l’accès passe majoritairement par les réseaux mobiles. Dans les zones urbaines, la 4G/5G peut fournir des débits utilisables, mais la qualité réelle dépend fortement de la charge de la cellule, de l’heure, de la couverture radio et de l’encombrement du dernier kilomètre. Dans les zones rurales ou isolées, les liaisons VSAT ou les connexions cellulaires restent souvent les seules options disponibles, avec des latences élevées et des coûts de données significatifs.

Dans ce contexte, les institutions qui dépendent de données distantes pour leur mission opérationnelle sont particulièrement exposées aux dégradations de qualité de service. Météo Madagascar, qui collecte, traite et diffuse des informations météorologiques critiques, utilise des liens d’accès hétérogènes, notamment des liaisons fibre et des connexions cellulaires Yas 5G/4G, selon les sites et les départements. Ces liens peuvent transporter plusieurs catégories de trafic : consultations de tableaux de bord, échanges interactifs, remontées de télémétrie, téléchargements de modèles numériques, images satellite, sauvegardes ou mises à jour logicielles.

## Problème opérationnel

Sur un lien d’accès contraint, la dégradation la plus pénalisante n’est pas toujours la perte de paquets. Elle peut provenir de l’accumulation de paquets dans les files d’attente, phénomène connu sous le nom de bufferbloat. Lorsqu’un transfert volumineux sature un lien, les paquets attendent dans une file trop longue, ce qui augmente fortement le temps de réponse. Pour une institution opérationnelle, cela peut signifier qu’un petit message critique, une requête de supervision, un échange vocal ou une alerte met plusieurs secondes à passer alors que le lien est occupé par un téléchargement de masse.

Le problème n’est donc pas seulement de mesurer le débit. Il est de comprendre si le lien reste utilisable pour le trafic critique pendant les périodes de charge, et quelles configurations locales peuvent améliorer la situation lorsque l’institution dispose d’un contrôle limité sur l’infrastructure cœur de réseau.

## Positionnement du projet

Ce mémoire s’inscrit dans le cadre d’un stage effectué dans l’environnement de Météo Madagascar. Il ne suppose pas un accès administrateur aux serveurs de production ni aux équipements centraux. Il adopte une approche non intrusive :

1. mesurer, depuis un poste client autorisé, le comportement de liens accessibles ;
2. rejouer des profils de liens contraints dans un banc d’essai Linux contrôlé ;
3. comparer des politiques AQM et des algorithmes de contrôle de congestion ;
4. produire des recommandations opérationnelles pour les équipements locaux ou les politiques d’usage des liens.

Le projet ne prétend pas fournir une vérité absolue sur le réseau interne de Météo Madagascar. Il fournit une méthode, un instrument, des mesures situées et des recommandations fondées sur des expérimentations reproductibles.

## Contributions

Les contributions de ce travail sont les suivantes.

1. **Une méthode d’audit QoS non intrusif** pour liens d’accès accessibles, fondée sur des mesures synthétiques et archivées.
2. **Un banc d’essai reproductible** sous Linux utilisant `tc`, `netem`, des qdiscs AQM et les algorithmes CUBIC/BBR.
3. **Une évaluation comparée** de configurations de file et de contrôle de congestion sur des profils de liens inspirés du contexte malgache.
4. **Un ensemble de recommandations pratiques** pour protéger le trafic critique lorsque le lien est soumis à des transferts de masse.
5. **Un instrument logiciel autonome**, écrit en Go, capable d’exécuter les campagnes, d’archiver les données et de régénérer les figures.

## Plan du mémoire

La première partie présente le contexte, les fondements théoriques et la problématique. La deuxième partie décrit la méthodologie : audit non intrusif, banc d’essai, métriques, contrôle qualité et menaces sur la validité. La troisième partie détaille la mise en œuvre technique. La quatrième partie présente les résultats, les recommandations opérationnelles, les limites et les perspectives.

---

# PARTIE I : CONTEXTE, ÉTAT DE L’ART ET PROBLÉMATIQUE

## I. Contexte opérationnel et contraintes des liens d’accès

### 1. Environnement de connectivité à Madagascar

Le paysage malgache des télécommunications repose principalement sur les réseaux mobiles. Les opérateurs fournissent des accès cellulaires dont la capacité varie selon la densité d’utilisateurs, l’heure de la journée et la couverture radio. Les offres prépayées structurent les usages : un volume de données limité est acheté pour une durée déterminée, à un coût qui peut représenter une part significative du revenu local. Dans ce contexte, chaque octet retransmis inutilement ou chaque dégradation de service peut avoir un impact économique.

Les liaisons fixes existent dans certaines zones, notamment via la fibre dans des environnements institutionnels ou urbains. Toutefois, les sites distants peuvent dépendre de liaisons cellulaires ou satellitaires. Cette hétérogénéité est au cœur du problème étudié : une institution peut disposer simultanément de liens fibre relativement stables et de liens cellulaires plus variables.

### 2. Contexte de Météo Madagascar

Météo Madagascar a pour mission d’observer, prévoir et diffuser des informations météorologiques. Cette mission repose sur des systèmes d’information qui peuvent impliquer :

- des postes de travail dans différents départements ;
- des liens d’accès fibre ou cellulaires ;
- des services distants ou internes ;
- des téléchargements de données volumineuses ;
- des échanges interactifs ;
- des besoins de supervision et de réactivité.

Dans le cadre de ce mémoire, nous nous intéressons aux liens d’accès utilisés par les utilisateurs ou les services, et non à l’infrastructure cœur de l’opérateur. Plus précisément, l’étude porte sur des liens accessibles depuis un poste d’extrémité autorisé.

### 3. Classes de trafic

Pour analyser la qualité de service, il est utile de distinguer deux grandes classes de trafic.

#### Trafic critique ou interactif

Ce trafic comprend :

- petites requêtes HTTP/API ;
- consultations de tableaux de bord ;
- messages de télémétrie ;
- échanges interactifs ;
- commandes ou validations nécessitant une réponse rapide.

Ce trafic est généralement peu volumineux mais sensible à la latence.

#### Trafic de masse

Ce trafic comprend :

- téléchargement de modèles météorologiques ;
- récupération d’images satellite ;
- sauvegardes ;
- mises à jour logicielles ;
- transferts de fichiers volumineux.

Ce trafic est volumineux et peut saturer le lien. Il n’a pas toujours besoin d’une latence faible, mais il peut provoquer du bufferbloat s’il n’est pas correctement régulé.

### 4. Contraintes locales

Plus contraintes influencent le dispositif :

- coût des données ;
- variabilité de la capacité radio ;
- possibles coupures d’électricité ;
- absence d’accès administrateur aux équipements centraux ;
- nécessité de ne pas perturber les utilisateurs ;
- besoin de mesures reproductibles et archivées.

Ces contraintes justifient une approche non intrusive et un banc d’essai contrôlé.

---

## II. Fondements théoriques

### 1. Congestion et bufferbloat

Lorsqu’un flux dépasse la capacité d’un goulot d’étranglement, les paquets s’accumulent dans une file d’attente. Tant que la file n’est pas pleine, il n’y a pas nécessairement de perte, mais le délai augmente. Si la file est trop grande, le délai peut devenir persistant. C’est le phénomène de bufferbloat.

La RFC 8290 rappelle l’objectif des mécanismes modernes de gestion de file : maintenir un délai de file faible sans sacrifier excessivement l’utilisation du lien.

### 2. Gestion active de file d’attente

La gestion active de file d’attente, ou AQM, regroupe des mécanismes qui régulent la file pour éviter l’accumulation excessive de délai. Trois configurations sont particulièrement étudiées dans ce mémoire.

#### `pfifo_fast`

`pfifo_fast` est une file simple, historiquement utilisée par défaut dans Linux. Elle peut laisser la file croître fortement sous charge, ce qui en fait une référence utile pour observer le bufferbloat.

#### `fq_codel`

`fq_codel` combine une file par flux et l’algorithme CoDel. Son objectif est de maintenir un faible délai de file tout en préservant un bon usage du lien.

#### `CAKE`

`CAKE` est une discipline de file plus complète, conçue pour simplifier la configuration tout en cherchant un bon compromis entre équité, faible latence et efficacité.

### 3. Contrôle de congestion TCP

Deux algorithmes de contrôle de congestion TCP sont comparés :

#### CUBIC

CUBIC est l’algorithme par défaut de Linux dans de nombreuses distributions. Il est fondé principalement sur la perte. Sous charge, il peut remplir les files, ce qui permet d’observer les effets de la congestion et de tester les AQM.

#### BBR

BBR cherche à estimer la bande passante disponible et le temps de propagation minimal. Il tente de maintenir une file plus faible. Son comportement peut différer sensiblement de celui de CUBIC selon les liens, les pertes et les tampons.

### 4. Qualité de service et trafic critique

La qualité de service ne se limite pas au débit maximal. Pour un trafic opérationnel critique, les dimensions importantes sont :

- la latence ;
- la latence de queue ;
- la stabilité du temps de réponse ;
- la capacité à respecter une échéance ;
- la perte de paquets ;
- le coût des retransmissions ;
- la prévisibilité.

Dans ce mémoire, nous introduisons la notion de **trafic critique mesuré par petits objets**. Il s’agit de requêtes ou de messages de petite taille dont le temps de complétion représente la réactivité perçue par une application opérationnelle.

---

## III. Travaux connexes

### 1. Mesure du bufferbloat

Le bufferbloat a été documenté comme une dégradation majeure dans les réseaux domestiques et mobiles. Des outils publics tels que bufferbloat.net, Flent ou bloat-o-meter permettent de mesurer des symptômes de congestion sur des liens réels. Ces outils sont précieux, mais ils fournissent généralement des mesures de bout en bout sans connaissance directe de la configuration interne des équipements.

### 2. AQM et contrôle de congestion

Les mécanismes `fq_codel` et `CAKE` ont été largement étudiés et déployés dans des routeurs domestiques ou communautaires. BBR a également fait l’objet de nombreuses études. Il est connu que ces mécanismes peuvent améliorer la latence sous charge, mais leur comportement dépend du contexte : pertes, tampons, capacité réelle, charge applicative.

### 3. Positionnement de ce travail

Ce mémoire ne cherche pas à prouver que l’AQM est utile en général. Ce résultat est largement établi. Il cherche plutôt à répondre à une question appliquée :

> Dans le contexte de liens d’accès accessibles à Météo Madagascar, comment évaluer de manière reproductible le comportement de configurations AQM/BBR et produire des recommandations opérationnelles pour protéger le trafic critique ?

La contribution est donc locale, méthodologique et opérationnelle.

---

## IV. Problématique et questions de recherche

## Problématique

Les liens d’accès utilisés dans un environnement opérationnel peuvent être soumis à des transferts volumineux qui dégradent la latence du trafic critique. L’institution ne dispose pas nécessairement d’un contrôle complet sur l’infrastructure de l’opérateur ou sur les serveurs centraux. Il est néanmoins possible :

- d’observer les liens accessibles depuis un poste client ;
- de rejouer des conditions semblables dans un environnement contrôlé ;
- de comparer des politiques AQM et des algorithmes de contrôle de congestion ;
- de formuler des recommandations pour les équipements locaux ou les politiques d’usage.

La problématique est donc la suivante :

> Comment évaluer, sans accès administrateur à l’infrastructure cœur, la capacité de liens d’accès contraints à protéger le trafic opérationnel critique, et quelles politiques AQM/BBR recommander à partir de mesures reproductibles ?

## Questions de recherche

### RQ1 — Audit des liens accessibles

> Les liens accessibles dans le contexte du stage présentent-ils des symptômes de dégradation de latence sous charge ?

Cette question est traitée par des mesures non intrusives sur les liens réellement accessibles.

### RQ2 — Évaluation contrôlée

> Dans des profils de liens contraints inspirés du contexte local, les politiques `fq_codel` et `CAKE` protègent-elles mieux le trafic critique que `pfifo_fast`, tout en conservant un débit de masse acceptable ?

Cette question est traitée dans le banc d’essai contrôlé.

### RQ3 — Recommandation opérationnelle

> Quelle configuration ou quelle politique d’usage peut être recommandée pour les équipements locaux ou les liens accessibles, afin de préserver le trafic critique pendant les transferts de masse ?

Cette question est traitée à partir des résultats d’audit et d’expérimentation.

## Hypothèses

### Tableau 1. Questions de recherche et hypothèses

| Identifiant | Hypothèse | Test | Interprétation |
|---|---|---|---|
| H1 | Sous charge de masse, `pfifo_fast` dégrade fortement la latence p95 des petits objets critiques | Comparaison appariée ou descriptive avec `fq_codel` et `CAKE` | Si confirmée, cela montre l’intérêt de l’AQM dans les profils testés |
| H2 | `fq_codel` ou `CAKE` réduit la latence p95 critique d’au moins [20 %] par rapport à `pfifo_fast`, avec une perte de débit utile limitée | Comparaison des métriques latence/débit | Si confirmée, la configuration peut être recommandée |
| H3 | BBR améliore certains indicateurs de latence mais son comportement dépend du profil | Analyse descriptive par profil | BBR n’est pas considéré comme universellement supérieur |

Les seuils exacts, par exemple 20 % de réduction ou 15 % de perte de débit acceptable, sont documentés dans la méthodologie et peuvent être ajustés selon les contraintes opérationnelles.

---

# PARTIE II : MÉTHODOLOGIE

## I. Vue d’ensemble du dispositif

La méthodologie repose sur deux volets complémentaires.

### 1. Audit non intrusif sur liens accessibles

L’audit consiste à mesurer le comportement de liens réellement accessibles depuis un poste client autorisé. Il ne modifie pas les équipements de production. Il utilise uniquement du trafic synthétique généré par l’expérimentateur.

### 2. Banc d’essai contrôlé

Le banc d’essai permet de comparer des configurations AQM et des algorithmes de contrôle de congestion dans des conditions reproductibles. Les profils de liens sont inspirés des observations issues de l’audit ou des connaissances du contexte local.

### Figure 1. Architecture générale

```
+---------------------+          +----------------------+
| Liens accessibles   |          | Banc d'essai Linux   |
| fibre / 5G / 4G     |          | tc + netem + qdisc   |
| audit client-side   |  profils | comparaison AQM/BBR  |
+----------+----------+          +-----------+----------+
           |                                  |
           v                                  v
+------------------------------------------------------+
| Instrument Go : mesures, campagnes, archives, figures|
+------------------------------------------------------+
```

---

## II. Audit QoS non intrusif sur liens accessibles

### 1. Objectif

L’audit a pour objectif de caractériser le comportement de liens accessibles dans des conditions normales ou légèrement chargées. Il ne s’agit pas de saturer longuement le réseau de l’institution, mais de produire des mesures comparables et utiles.

### 2. Conditions

Les mesures sont réalisées si les conditions suivantes sont réunies :

- autorisation explicite du responsable de stage ou de l’équipe informatique ;
- utilisation d’un poste client connecté au lien étudié ;
- trafic synthétique uniquement ;
- absence de capture du trafic des autres utilisateurs ;
- durée limitée des tests de charge ;
- documentation du site, du type de lien et de l’horaire.

### 3. Mesures réalisées

Pour chaque lien accessible, les mesures suivantes peuvent être effectuées :

#### Mesure au repos

- ping ICMP pendant 60 secondes ;
- calcul des percentiles p50, p95 et p99 ;
- perte de paquets.

#### Mesure de petits objets

- requêtes HTTP ou TCP portant sur de petits objets, par exemple 4 Ko à 32 Ko ;
- mesure du temps de complétion ;
- calcul du p95.

#### Mesure de débit

- test `iperf3` descendant ou montant de courte durée ;
- mesure du débit utile ;
- consignation de la durée et du volume approximatif.

#### Mesure sous charge contrôlée

- ping continu pendant une charge courte ;
- petits objets pendant la charge ;
- comparaison avec l’état au repos.

### 4. Sortie de l’audit

Les données sont consignées dans un fichier `link_audit.csv`.

### Tableau 6. Schéma de la table `link_audit.csv`

| Colonne | Description |
|---|---|
| audit_id | identifiant de la mesure |
| timestamp | horodatage |
| site | département ou localisation |
| link_type | fiber, 5g, 4g, vsat, other |
| provider | opérateur ou fournisseur |
| rtt_idle_p50_ms | ping p50 au repos |
| rtt_idle_p95_ms | ping p95 au repos |
| rtt_loaded_p50_ms | ping p50 sous charge |
| rtt_loaded_p95_ms | ping p95 sous charge |
| throughput_mbps | débit mesuré |
| loss_pct | perte de paquets |
| http_small_p95_ms | p95 des petits objets |
| data_used_mb | volume approximatif utilisé |
| notes | observations |

---

## III. Banc d’essai contrôlé pour l’évaluation AQM/BBR

### 1. Environnement

Le banc d’essai fonctionne sur une machine Linux, par exemple une VM Ubuntu 24.04 LTS. Il utilise :

- `netem` pour la latence, la gigue ou les pertes ;
- `TBF` ou un façonneur équivalent pour limiter la capacité ;
- les qdiscs `pfifo_fast`, `fq_codel`, `CAKE` ;
- les contrôles de congestion CUBIC et BBR ;
- du trafic TCP généré par le noyau ou par `iperf3` selon la configuration retenue ;
- des sondes ping et petits objets HTTP/TCP.

### 2. Profils de liens

Deux profils principaux sont utilisés.

### Tableau 2. Profils de liens utilisés dans le banc d’essai

| Profil | Capacité approximative | RTT approximatif | Perte | Gigue | Justification |
|---|---:|---:|---:|---:|---|
| P1 — fibre stable | [ex. 80 Mbps] | [ex. 20 ms] | 0 % | faible | lien stable de référence |
| P2 — cellulaire contraint | [ex. 20 Mbps variable] | [ex. 80–150 ms] | [0–0.5 %] | modérée | profil inspiré des liens 4G/5G |

Si des traces réelles ont été collectées pendant l’audit, elles peuvent être utilisées pour calibrer ces profils. Sinon, les profils sont présentés comme représentatifs de la classe de liens étudiée.

### 3. Configurations testées

### Tableau 3. Matrice expérimentale AQM/BBR

| Axe | Valeurs |
|---|---|
| Profils de lien | P1 fibre stable, P2 cellulaire contraint |
| Discipline de file | `pfifo_fast`, `fq_codel`, `CAKE` |
| Contrôle de congestion | CUBIC, BBR |
| Répétitions | 3 par cellule |

Nombre total d’événements :

> 2 profils × 3 qdiscs × 2 CC × 3 répétitions = 36 événements.

Si le temps ou la stabilité de la machine est insuffisant, une version réduite peut être utilisée :

> 1 profil contraint × 3 qdiscs × 2 CC × 3 répétitions = 18 événements.

Cette réduction doit être documentée.

### 4. Charge applicative

Chaque événement combine deux types de trafic.

#### Trafic critique

- petits objets de [4 Ko, 16 Ko ou 32 Ko] ;
- envoi périodique, par exemple toutes les 2 secondes ;
- mesure du temps de complétion ;
- calcul du p95 et du taux de respect d’une échéance.

#### Trafic de masse

- un ou plusieurs flux TCP descendants ou montants ;
- CUBIC ou BBR selon la cellule ;
- durée suffisante pour observer la congestion ;
- mesure du débit utile.

### 5. Déroulement d’un événement

Chaque événement suit trois phases :

1. **Baseline** : mesure du lien sans charge de masse.
2. **Charge** : activation du trafic de masse et poursuite des sondes critiques.
3. **Récupération** : arrêt de la charge de masse et observation du retour à l’état nominal.

Durées recommandées :

| Phase | Durée indicative |
|---|---:|
| Baseline | 30 secondes |
| Charge | 120 secondes |
| Récupération | 30 secondes |

La durée totale d’un événement est d’environ 3 minutes. Pour 36 événements, la campagne complète représente environ 108 minutes, hors installation, reconfiguration et aléas techniques.

---

## IV. Métriques

### Tableau 4. Métriques principales

| Métrique | Définition | Utilité |
|---|---|---|
| rtt_p50_ms | médiane du ping | latence typique |
| rtt_p95_ms | percentile 95 du ping | latence de queue |
| small_p95_ms | percentile 95 du temps de complétion des petits objets | trafic critique |
| deadline_ok_pct | pourcentage de petits objets complétés sous une échéance donnée | respect d’un objectif opérationnel |
| bulk_goodput_mbps | débit utile du trafic de masse | efficacité du transfert |
| wasted_bytes | octets retransmis ou estimation | gaspillage |
| cost_ar_per_h | coût estimé du gaspillage en ariary par heure | impact économique |

### Tableau 5. Métriques secondaires

| Métrique | Définition | Utilité |
|---|---|---|
| drops | pertes au niveau qdisc ou TCP | diagnostic |
| retransmissions | retransmissions TCP | gaspillage |
| jitter_ms | variation de latence | interactivité |
| cpu_pct | usage CPU de la machine de test | covariable |
| fairness_index | équité entre flux, si mesurée | analyse secondaire |

### Définition du coût en ariary

Le coût du gaspillage peut être estimé ainsi :

\[
cost = \frac{wasted\_bytes}{4.5 \times 1024^3} \times 30000
\]

Cette formule utilise l’hypothèse documentaire d’un forfait de 30 000 ariary pour 4,5 Go. Elle doit être présentée comme une approximation destinée à rendre l’impact économique lisible, non comme une facturation exacte.

---

## V. Contrôle qualité et reproductibilité

### 1. Archivage

Chaque campagne produit :

- fichiers CSV gelés ;
- manifeste d’intégrité ;
- configuration hachée ;
- journaux d’exécution ;
- figures régénérables ;
- journal des événements invalides ou quarantaines.

### 2. Portes de qualité

Des portes simples sont appliquées à chaque événement.

### Tableau 8. Portes de qualité

| Porte | Condition | Action si échec |
|---|---|---|
| G0 | Le lien ou la cible est joignable | événement invalide |
| G1 | La charge de masse démarre effectivement | événement invalide |
| G2 | Les sondes critiques produisent des données | événement dégradé |
| G3 | La latence mesurée est physiquement plausible | événement invalide |
| G4 | Le débit mesuré est cohérent avec le profil | événement invalide |
| G5 | Absence de duplication de lignes | événement invalide |
| G6 | Baseline stable avant charge | événement dégradé |
| G7 | CPU non saturé au point de fausser la mesure | covariable |

### 3. Réconciliation interne

Les grandeurs importantes peuvent être comparées entre deux sources internes :

- compteurs noyau ;
- sortie `iperf3` ;
- mesures de sondes ;
- captures éventuelles.

Il est important de ne pas présenter ces canaux comme totalement indépendants s’ils fonctionnent sur la même machine. Le terme correct est :

> réconciliation interne ou contre-vérification logicielle.

### 4. Régénération des figures

Toutes les figures sont produites par :

```bash
make figures
```

Aucune figure n’est dessinée manuellement.

---

## VI. Menaces sur la validité

### 1. Absence de vérité opérateur

Les mesures sur liens réels sont effectuées depuis un poste client. Elles ne donnent pas accès à la configuration interne de l’opérateur, aux files du routeur distant ou à l’ordonnanceur radio. Les conclusions sur les liens réels sont donc observationnelles.

### 2. Effets de la machine virtuelle

Le banc d’essai fonctionne dans une machine virtuelle. Les performances peuvent être affectées par :

- l’hyperviseur ;
- le nombre de vCPU ;
- les interruptions ;
- l’usage disque ;
- la charge CPU.

Des précautions sont prises, mais le bruit de virtualisation reste une menace.

### 3. Petite taille d’échantillon

Le nombre d’événements est volontairement réduit pour rester soutenable. Les résultats sont donc présentés avec prudence : médianes, intervalles ou IQR, tendances, et non affirmations statistiques trop fortes.

### 4. Généralisation limitée

Les profils de liens sont inspirés du contexte local. Ils ne représentent pas tous les liens de Météo Madagascar ni tous les réseaux malgaches.

### 5. Trafic synthétique

Les mesures utilisent du trafic synthétique. Elles ne capturent pas toute la complexité du trafic réel des utilisateurs.

---

# PARTIE III : MISE EN ŒUVRE

## I. Environnement technique

### 1. Machine

| Élément | Valeur |
|---|---|
| OS | Ubuntu 24.04 LTS |
| Noyau | Linux 6.x |
| vCPU | 2 |
| RAM | 4 Go |
| Disque | 40 Go minimum |
| Binaire | Go statique |
| Interface | React/Vite/TypeScript, embarquée ou servie localement |

### 2. Outils principaux

| Fonction | Outil |
|---|---|
| Façonnage de lien | `tc`, `netem`, `TBF` |
| Files d’attente | `pfifo_fast`, `fq_codel`, `CAKE` |
| Contrôle de congestion | CUBIC, BBR |
| Mesure de latence | ping |
| Débit | `iperf3` ou flux TCP natifs |
| Archivage | CSV, manifestes SHA-256 |
| Figures | pipeline `make figures` |
| Interface | React, SSE |

---

## II. Architecture de l’instrument

L’instrument est organisé en modules.

### 1. Module campagne

Il orchestre :

- la préparation ;
- la configuration des qdiscs ;
- le lancement des flux ;
- la collecte des mesures ;
- l’écriture des CSV ;
- la clôture de l’événement.

### 2. Module qdisc

Il configure les disciplines de file via `tc` ou netlink.

### 3. Module sondes

Il mesure :

- ping ;
- petits objets ;
- débit ;
- pertes éventuelles.

### 4. Module métriques

Il calcule :

- percentiles ;
- goodput ;
- wasted bytes ;
- coût estimé ;
- indicateurs VoIP ou interactifs si présents.

### 5. Module archive

Il produit :

- `aqm_eval.csv` ;
- `link_audit.csv` ;
- manifeste ;
- journaux ;
- figures.

### Tableau 7. Schéma de la table `aqm_eval.csv`

| Colonne | Description |
|---|---|
| run_id | identifiant d’exécution |
| event_id | identifiant d’événement |
| profile | P1 ou P2 |
| qdisc | pfifo_fast, fq_codel, cake |
| cc | cubic ou bbr |
| repetition | numéro de répétition |
| rtt_p50_ms | ping p50 |
| rtt_p95_ms | ping p95 |
| small_p95_ms | p95 petits objets |
| deadline_ok_pct | pourcentage sous échéance |
| bulk_goodput_mbps | débit utile |
| drops | pertes |
| retransmissions | retransmissions |
| wasted_bytes | octets gaspillés |
| cost_ar_per_h | coût estimé |
| cpu_pct | usage CPU |
| gate_status | statut qualité |

---

## III. Déploiement des campagnes

### 1. Construction

```bash
make build
make test
```

### 2. Vérification des modules noyau

```bash
sudo modprobe tcp_bbr
sysctl net.ipv4.tcp_congestion_control
tc qdisc --help
```

### 3. Campagne de fumée

Avant la campagne principale :

- vérifier qu’un événement simple s’exécute ;
- vérifier que les CSV sont écrits ;
- vérifier que `make figures` fonctionne ;
- vérifier la stabilité sur un événement de 5 minutes.

### 4. Audit réel

Exemple de commande ou d’action dans l’interface :

```bash
./cgo audit --link-type 5g --site "Department X" --duration 300
```

Si l’outil n’a pas encore cette commande, l’audit peut être réalisé par scripts simples, puis importé dans `link_audit.csv`.

### 5. Campagne contrôlée

```bash
./cgo run --matrix aqm-bbr --profiles P1,P2 --reps 3
```

ou, selon l’implémentation réelle :

```bash
./cgo --e3 --matrix reduced
```

### 6. Vérification

```bash
make verify
make figures
```

---

## IV. Interface de supervision

L’interface n’a pas besoin d’être complexe. Elle doit être stable et lisible.

### 1. Panneau campagne

Il affiche :

- profil sélectionné ;
- qdisc ;
- contrôle de congestion ;
- répétition ;
- état de l’événement ;
- statut des portes.

### 2. Panneau mesures temps réel

Via SSE, il affiche :

- RTT p50 ;
- RTT p95 ;
- petits objets p95 ;
- débit ;
- pertes ;
- statut de charge.

### 3. Panneau résultats

Il affiche :

- tableau comparatif ;
- médianes ;
- IQR ou intervalles ;
- quarantaines ;
- meilleure configuration selon les métriques choisies.

### 4. Panneau intégrité

Il affiche :

- identifiant d’archive ;
- manifeste SHA-256 ;
- nombre d’événements valides ;
- nombre d’événements quarantenés ;
- bouton de régénération des figures.

### 5. Actions nécessaires

| Action | Description |
|---|---|
| Start audit | lancer une mesure sur lien accessible |
| Import profile | importer un profil réel |
| Start run | démarrer une cellule expérimentale |
| Stop | arrêter proprement |
| Resume | reprendre après interruption |
| Verify | vérifier l’intégrité |
| Replay | rejouer une archive |
| Export report | exporter un rapport court |

---

## V. Sécurité, éthique et limites d’accès

### 1. Pas de modification de production

L’instrument ne doit pas modifier :

- serveurs de Météo Madagascar ;
- routeurs centraux ;
- équipements opérateur ;
- configuration réseau non maîtrisée.

### 2. Trafic synthétique uniquement

Les captures et mesures ne concernent que le trafic généré par l’outil. Le trafic des autres utilisateurs n’est pas capturé.

### 3. Autorisation

Les mesures sur lien réel doivent être autorisées. Les tests de charge doivent être courts et, si possible, réalisés hors des périodes critiques.

### 4. Coût des données

Les tests de débit consomment du volume. Sur des liens prépayés ou limités, la consommation doit être estimée et documentée.

---

# PARTIE IV : RÉSULTATS, RECOMMANDATIONS ET PERSPECTIVES

## I. Résultats de l’audit des liens accessibles

Cette section présente les mesures réalisées sur les liens accessibles.

### Tableau 9. Résultats d’audit des liens accessibles

| Site | Type de lien | Fournisseur | RTT repos p50 | RTT repos p95 | RTT charge p95 | Débit | Small p95 | Tendance bufferbloat |
|---|---|---|---:|---:|---:|---:|---:|---|
| [Département A] | [fiber] | [Yas/fibre] | [.. ms] | [.. ms] | [.. ms] | [.. Mbps] | [.. ms] | [faible/moyenne/forte] |
| [Département B] | [5G] | [Yas] | [.. ms] | [.. ms] | [.. ms] | [.. Mbps] | [.. ms] | [faible/moyenne/forte] |
| [Département C] | [4G/5G] | [Yas] | [.. ms] | [.. ms] | [.. ms] | [.. Mbps] | [.. ms] | [faible/moyenne/forte] |

### Interprétation type

Si les mesures montrent une forte augmentation du p95 sous charge :

> Les mesures indiquent que la latence de queue augmente significativement lorsque le lien est chargé. Cette dégradation peut affecter les requêtes critiques et les échanges interactifs, même si le débit reste acceptable.

Si les mesures ne montrent pas de dégradation significative :

> Les mesures effectuées sur cet échantillon limité ne montrent pas de dégradation majeure de la latence sous charge. Cela peut indiquer que le lien testé dispose d’une capacité suffisante, que la charge générée était trop faible, ou que le phénomène est variable selon l’heure et le site.

---

## II. Résultats de l’évaluation AQM/BBR

Cette section présente les résultats du banc d’essai contrôlé.

### Tableau 10. Résultats de comparaison AQM/BBR

| Profil | Qdisc | CC | RTT p95 | Small p95 | Deadline OK | Goodput | Wasted bytes | Cost Ar/h |
|---|---|---|---:|---:|---:|---:|---:|---:|
| P2 | pfifo_fast | CUBIC | [..] | [..] | [..] | [..] | [..] | [..] |
| P2 | fq_codel | CUBIC | [..] | [..] | [..] | [..] | [..] | [..] |
| P2 | CAKE | CUBIC | [..] | [..] | [..] | [..] | [..] | [..] |
| P2 | pfifo_fast | BBR | [..] | [..] | [..] | [..] | [..] | [..] |
| P2 | fq_codel | BBR | [..] | [..] | [..] | [..] | [..] | [..] |
| P2 | CAKE | BBR | [..] | [..] | [..] | [..] | [..] | [..] |
| P1 | pfifo_fast | CUBIC | [..] | [..] | [..] | [..] | [..] | [..] |
| P1 | fq_codel | CUBIC | [..] | [..] | [..] | [..] | [..] | [..] |
| P1 | CAKE | CUBIC | [..] | [..] | [..] | [..] | [..] | [..] |
| P1 | pfifo_fast | BBR | [..] | [..] | [..] | [..] | [..] | [..] |
| P1 | fq_codel | BBR | [..] | [..] | [..] | [..] | [..] | [..] |
| P1 | CAKE | BBR | [..] | [..] | [..] | [..] | [..] | [..] |

### Figure 5. Comparaison p95 par configuration

[Insérer figure régénérée : petit objet p95 selon qdisc et CC.]

### Figure 6. Compromis latence/débit

[Insérer figure régénérée : small p95 vs bulk goodput.]

### Interprétation type

Si `fq_codel` ou `CAKE` réduisent fortement le p95 :

> Les configurations AQM réduisent la latence de queue du trafic critique par rapport à la file de référence. Le débit de masse reste acceptable dans les limites définies.

Si BBR réduit la latence mais crée des différences selon le profil :

> BBR réduit la latence dans certains profils, mais son comportement varie selon la perte, la gigue et la capacité. Il ne peut donc pas être recommandé sans test préalable sur le lien concerné.

Si les résultats sont non concluants :

> Les résultats ne permettent pas de dégager une supériorité nette d’une configuration sur l’ensemble des cellules. Les variations peuvent provenir du bruit de la machine virtuelle, de la petite taille de l’échantillon ou des paramètres des profils. Cette incertitude est rapportée honnêtement.

---

## III. Recommandations opérationnelles

Les recommandations suivantes sont proposées sous réserve que les équipements locaux soient sous contrôle de l’institution.

### Tableau 11. Recommandations opérationnelles

| Situation | Recommandation | Justification |
|---|---|---|
| Lien fibre stable | `fq_codel` par défaut | suffisant dans de nombreux cas |
| Lien cellulaire contraint avec trafic critique | tester `fq_codel` puis `CAKE` | meilleure protection contre bufferbloat |
| Téléchargements volumineux simultanés | limiter légèrement la capacité utilisée | éviter de saturer totalement le lien |
| BBR | tester avant déploiement large | comportement dépendant du lien |
| Trafic critique | prioriser les petits flux si l’équipement le permet | préserver la réactivité |
| Mesure | conserver l’outil d’audit | suivi dans le temps |

### Exemple de recommandation prudente

> Pour un lien cellulaire contraint transportant à la fois des téléchargements volumineux et du trafic interactif, il est recommandé d’évaluer une configuration `fq_codel` ou `CAKE` sur un équipement de test, avec une capacité façonnée légèrement inférieure à la capacité mesurée. Le déploiement doit être effectué pendant une fenêtre de maintenance et suivi par des mesures de latence p95.

### Exemple de commande indicative

```bash
# Exemple indicatif à adapter au matériel réel
sudo tc qdisc replace dev eth0 root fq_codel
```

ou, si un façonneur est utilisé :

```bash
# Exemple indicatif avec CAKE
sudo tc qdisc replace dev eth0 root cake bandwidth 18mbit
```

Ces commandes sont données à titre illustratif. Elles doivent être adaptées à l’interface réseau, à la capacité réelle et aux contraintes de l’équipement.

---

## IV. Limites

Les limites suivantes sont assumées.

### 1. Mesures client-side seulement

L’audit ne permet pas de connaître la configuration interne des équipements opérateur. Il mesure l’effet observable depuis un poste client.

### 2. Échantillon limité

Le nombre de mesures réelles et d’événements expérimentaux est volontairement limité. Les résultats doivent être interprétés comme une étude de cas.

### 3. Environnement virtualisé

Le banc d’essai peut introduire du bruit. Les tendances sont plus importantes que les valeurs absolues.

### 4. Absence de déploiement réel

Les recommandations ne sont pas déployées sur l’infrastructure de production. Elles constituent une aide à la décision.

### 5. Trafic synthétique

Le trafic généré ne reproduit pas parfaitement les applications réelles de Météo Madagascar.

---

## V. Perspectives

Plus prolongements sont possibles.

### 1. Extension de l’audit

Il serait utile de multiplier les mesures :

- plusieurs heures de la journée ;
- plusieurs jours ;
- plusieurs sites ;
- différents types de liens.

### 2. Intégration à une supervision continue

L’instrument pourrait être utilisé comme sonde périodique pour produire des rapports hebdomadaires ou mensuels.

### 3. Déploiement contrôlé

Si la DSI dispose d’équipements locaux maîtrisés, une expérimentation réelle avec `fq_codel` ou `CAKE` pourrait être réalisée dans une fenêtre de maintenance.

### 4. Prise en compte du coût

La métrique de coût en ariary pourrait être raffinée avec les tarifs réels des forfaits utilisés par l’institution.

### 5. Étude BBR plus approfondie

BBR mériterait une étude spécifique sur les liens cellulaires locaux, notamment en présence de pertes et de variabilité radio.

---

# CONCLUSION GÉNÉRALE

Ce mémoire a étudié la protection du trafic opérationnel critique sur liens d’accès contraints dans le contexte de Météo Madagascar. Plutôt que de proposer une solution théorique globale, le travail a adopté une approche pragmatique et reproductible : audit non intrusif de liens accessibles, expérimentation contrôlée sous Linux, comparaison de configurations AQM et de contrôles de congestion, production de recommandations opérationnelles.

Les principaux apports sont les suivants. Premièrement, une méthode d’audit client-side adaptée aux contraintes d’un stage où l’accès administrateur aux équipements de production n’est pas disponible. Deuxièmement, un banc d’essai permettant de comparer `pfifo_fast`, `fq_codel` et `CAKE` avec CUBIC et BBR sur des profils de liens inspirés du contexte local. Troisièmement, un instrument logiciel autonome capable d’archiver les mesures, de vérifier leur intégrité et de régénérer les figures. Quatrièmement, une série de recommandations prudentes pour les équipements locaux ou les politiques d’usage des liens.

Les résultats montrent que la question pertinente n’est pas seulement le débit maximal disponible, mais la capacité du lien à préserver la réactivité du trafic critique pendant les périodes de charge. Dans un environnement où les données sont coûteuses et où les liens peuvent être variables, cette dimension est opérationnellement importante.

Ce travail ne prétend pas fournir une vérité définitive sur l’ensemble des réseaux utilisés par Météo Madagascar. Il fournit une méthode, des données situées, un outil réutilisable et des recommandations fondées sur des expérimentations contrôlées. À ce titre, il constitue une base utile pour des audits ultérieurs, des expérimentations encadrées et une amélioration progressive de la qualité de service.

---

# RÉFÉRENCES BIBLIOGRAPHIQUES

Nichols K., Bashir A., McGregor A., IETF, 2018. **RFC 8290: Controlling Queue Delay**.

Gettys J., Nichols K., ACM Queue, 2011. **Bufferbloat: Dark Buffers in the Internet**.

Jiang H., Papadopoulos C., Proceedings of ACM IMC, 2010. **An experimental study of home network performance**.

Cardwell N., Cheng Y., Gunn C. S., Yeganeh S. H., Jacobson V., ACM Queue, 2016. **BBR: Congestion-Based Congestion Control**.

Ha S., Rhee I., Xu L., ACM SIGOPS Operating Systems Review, 2008. **CUBIC: A new TCP-friendly high-speed TCP variant**.

Høiland-Jørgensen T., McKenney P., Taht D., Gettys J., Dumazet E., IEEE Access, 2018. **The CAKE Adaptive Queue Management Algorithm**.

IETF, 2018. **RFC 8312: CUBIC for Fast and Long-Distance Networks**.

Allman M., Paxson V., Stevens W., IETF, 2009. **RFC 5681: TCP Congestion Control**.

ITU-T, 2019. **Recommendation G.107: The E-model, a computational model for use in transmission planning**.

World Meteorological Organization. **WMO Information System and Global Basic Observing Network documentation**. [à compléter si utilisé]

ARTEC, Antananarivo. **Rapports sur la qualité de service des réseaux de communications électroniques**. [à compléter si utilisé]

République de Madagascar, 2014. **Loi n° 2014-038 sur la protection des données personnelles**.

---

# RÉFÉRENCES WEBOGRAPHIQUES

https://www.bufferbloat.net — Bufferbloat.net.

https://flent.org — Flent: The FLExible Network Tester.

https://www.artec.mg — Autorité de Régulation des Technologies de Communication.

https://www.itu.int — International Telecommunication Union.

https://zenodo.org — Zenodo, archivage de données de recherche.

https://wmo.int — World Meteorological Organization.

---

# GLOSSAIRE

**Bufferbloat** : dégradation de la performance causée par l’accumulation persistante de paquets dans des files d’attente surdimensionnées, ajoutant du délai sans perte visible immédiate.

**AQM** : Active Queue Management, ensemble de mécanismes régulant le délai de file plutôt que la seule perte de paquets.

**BBR** : algorithme de contrôle de congestion fondé sur l’estimation de la bande passante et du temps de propagation.

**CAKE** : discipline de file cherchant à fournir faible latence, équité et simplicité de configuration.

**CUBIC** : algorithme de contrôle de congestion TCP fondé principalement sur la perte.

**Fichier CSV gelé** : fichier de résultats archivé, non modifié après validation, utilisé pour régénérer les figures.

**Manifeste SHA-256** : liste d’empreintes cryptographiques permettant de vérifier l’intégrité des fichiers archivés.

**Petit objet critique** : requête ou message de faible taille dont le temps de complétion représente la réactivité d’une application opérationnelle.

**Profil de lien** : ensemble de paramètres représentant le comportement d’un lien d’accès : capacité, RTT, perte, gigue, variabilité.

**Quarantaine** : mise à l’écart d’un événement invalide ou dégradé, consignée dans un journal.

**Trafic de masse** : trafic volumineux, tel que téléchargement de fichiers ou de modèles, pouvant saturer un lien.

**Trafic critique** : trafic sensible à la latence, tel que petites requêtes, télémétrie ou échanges interactifs.

---

# ANNEXES

## Annexe A. Commandes de base

```bash
make build
make test
make figures
make verify
make offline-kit
```

Si l’instrument conserve les commandes de l’ancien projet, elles peuvent être documentées, mais les expériences non centrales doivent être présentées comme exploratoires :

```bash
./cgo --dry-run
./cgo --serve
./cgo --e3
./cgo --sweep
```

## Annexe B. Commandes réseau indicatives

### Vérification des qdiscs

```bash
tc qdisc show dev eth0
```

### Chargement du module BBR

```bash
sudo modprobe tcp_bbr
sysctl net.ipv4.tcp_available_congestion_control
```

### Exemple de configuration `fq_codel`

```bash
sudo tc qdisc replace dev eth0 root fq_codel
```

### Exemple de configuration `CAKE`

```bash
sudo tc qdisc replace dev eth0 root cake bandwidth 20mbit
```

Ces commandes sont indicatives. Elles doivent être adaptées à l’interface réelle, à la capacité mesurée et à l’équipement utilisé.

---

## Annexe C. Protocole d’audit simplifié

Pour un lien accessible :

1. Vérifier l’autorisation.
2. Identifier le site, le type de lien et le fournisseur.
3. Mesurer le ping au repos pendant 60 secondes.
4. Mesurer le temps de complétion de petits objets.
5. Mesurer un débit court avec `iperf3` si autorisé.
6. Mesurer le ping pendant une charge courte.
7. Enregistrer les résultats dans `link_audit.csv`.
8. Archiver la mesure et calculer son empreinte.

---

## Annexe D. Protocole de campagne contrôlée

1. Définir le profil de lien.
2. Configurer la qdisc.
3. Configurer le contrôle de congestion.
4. Lancer la baseline.
5. Lancer le trafic de masse.
6. Maintenir les sondes critiques.
7. Arrêter la charge.
8. Observer la récupération.
9. Écrire les métriques dans `aqm_eval.csv`.
10. Vérifier les portes de qualité.
11. Répéter trois fois par cellule.
12. Régénérer les figures.

---

## Annexe E. Exemple de rapport court pour la DSI

```text
Rapport d’audit QoS — [site]
Date : [date]
Lien : [fibre / 5G / 4G]
Fournisseur : [fournisseur]

Observations :
- RTT au repos : [.. ms]
- RTT sous charge : [.. ms]
- Débit mesuré : [.. Mbps]
- Small-object p95 : [.. ms]
- Tendance bufferbloat : [faible / moyenne / forte]

Recommandation :
- [exemple : tester fq_codel sur l’équipement local]
- [exemple : éviter les téléchargements volumineux simultanés pendant les périodes critiques]
- [exemple : mesurer à nouveau après configuration]

Limites :
- Mesure client-side uniquement.
- Pas de modification de l’infrastructure opérateur.
- Échantillon limité.
```

---

## Annexe F. Ce qui est retiré de l’ancien projet

Les éléments suivants ne constituent plus le cœur du mémoire :

1. Le classement de cinq détecteurs de congestion.
2. L’hypothèse K-S comme contribution principale.
3. L’expérience E2 d’estimation de bande passante par packet-pair/packet-train.
4. Le façonneur adaptatif complexe.
5. La matrice de 356 événements.
6. Les vagues défectueuses utilisées comme résultats principaux.
7. L’affirmation de vérité terrain appliquée aux liens réels.
8. L’affirmation de canaux totalement indépendants lorsqu’ils sont exécutés sur la même machine.

Ils peuvent éventuellement être mentionnés en annexe comme travaux exploratoires, à condition que cela soit clairement indiqué.

---

# VERSION COURTE DU MÉMOIRE

Si vous devez présenter le projet en une page, utilisez cette version.

## Titre

**Protection du trafic opérationnel critique sur liens d’accès contraints : audit QoS non intrusif et évaluation des politiques AQM/BBR dans le contexte de Météo Madagascar**

## Problème

Les liens d’accès utilisés par Météo Madagascar peuvent être soumis à des transferts volumineux qui dégradent la latence du trafic critique. Sans accès administrateur à l’infrastructure cœur, il est difficile de diagnostiquer le phénomène et de recommander des configurations adaptées.

## Solution

Le projet propose :

1. un audit non intrusif de liens accessibles ;
2. un banc d’essai Linux reproductible ;
3. une comparaison de `pfifo_fast`, `fq_codel`, `CAKE`, CUBIC et BBR ;
4. des recommandations pour les équipements locaux.

## Méthode

- Mesures client-side sur liens accessibles.
- Émulation de profils contraints avec `tc/netem`.
- Trafic critique simulé par petits objets.
- Trafic de masse simulé par flux TCP.
- Archivage CSV, manifestes, figures régénérables.

## Résultats attendus ou obtenus

- Identification des symptômes de bufferbloat sur les liens audités.
- Comparaison des configurations AQM/BBR.
- Recommandation de configurations pour protéger le trafic critique.

## Contribution pour l’entreprise

- Rapport d’audit.
- Outil réutilisable.
- Recommandations opérationnelles.
- Méthode de mesure reproductible.

## Limites

- Pas d’accès aux équipements opérateur.
- Mesures client-side.
- Trafic synthétique.
- Échantillon limité.

---

