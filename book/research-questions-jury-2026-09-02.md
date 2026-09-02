# Recherche : les vraies questions critiques que les jurys posent (soutenance M2 réseaux / bufferbloat-AQM)

Date : 2026-09-02. Méthode : revue web de guides de soutenance FR/EN (jury de master/doctorat), retours d'examinateurs, documentation communautaire bufferbloat/Flent. Application directe au mémoire M2 (banc netem, fq_codel/CAKE, CUBIC/BBR, profils fibre/4G/VSAT/Starlink, p95, échéance, QDI, ariary, CSV SHA-256).

Sources principales :
- ThesisLaunch, « 50 Most Common Defense Questions » (questions réelles de comités canadiens, niveau master : 15-25 questions, 2 rounds) : https://thesislaunch.com/guide/50-common-defense-questions-how-to-answer/
- MockDefense, « Methodology Defense Questions » (vue examinateur) : https://mockdefense.org/guides/methodology-defense-questions
- MockDefense, « How to Discuss the Limitations of Your Study » : https://mockdefense.org/guides/how-to-answer-limitations-questions
- Analisis de Datos Psicologia, « Thesis Defense Questions: the 12 Your Committee Will Ask » (les 12 questions qui reviennent dans TOUTES les soutenances) : https://analisisdedatospsicologia.com/en/blog/thesis-defense-questions
- JuryAI, « Soutenance de master : 6 questions du jury + réponses types » : https://www.juryai.app/blog/reussir-soutenance-master
- Komelate, « Réussir sa soutenance de mémoire » (contexte universités africaines francophones) : https://komelate.com/blog/soutenance-de-memoire
- eloqole, « Soutenance de mémoire : ouverture et questions du jury » (exemples verbatim Q/R) : https://eloqole.com/fr/exemples/soutenance-de-memoire-exemples/
- LabLeaz, « Difficult Thesis Defense Questions: 10 Strategies » : https://lableaz.com/blog/difficult-thesis-defense-question
- Bufferbloat.net : RRUL Chart Explanation (spéc RRUL 70 s = 5+60+5) : https://www.bufferbloat.net/projects/bloat/wiki/RRUL_Chart_Explanation/ ; Tests for Bufferbloat (Flent outil de référence, netperf/iperf3) : https://www.bufferbloat.net/projects/bloat/wiki/Tests_for_Bufferbloat/ ; FAQs (objection « charge artificielle », cake-autorate pour liens 4G/5G à débit variable) : https://www.bufferbloat.net/projects/bloat/wiki/Bufferbloat_FAQs/
- Flent.org (répétabilité, métadonnées, runs batch, box plots multi-runs) : https://flent.org
- Scribbr FR, « La soutenance d'un mémoire » (questions récurrentes) : https://www.scribbr.fr/memoire/la-soutenance-de-votre-memoire/
- Compilatio, « Soutenance de mémoire : guide complet » (anticiper questions sur limites méthodo ; contexte intégrité/IA) : https://www.compilatio.net/blog/soutenance-memoire

---

## PARTIE 1 — 25 questions réelles, catégorisées (formulation + piège caché)

Constat transversal des sources : les jurys ne posent presque jamais de questions exotiques. Toutes les sources convergent — les mêmes 10-12 fronts reviennent « defense after defense » (design, échantillon/répétitions, instruments, hypothèses stats, généralisation, contribution, limites). Les questions ci-dessous reformulent ces fronts dans le vocabulaire d'un jury réseau.

### A. Provenance des données

**A1. « D'où viennent vos données ? Prenez un chiffre de ce tableau : décrivez-nous exactement comment il a été produit, de la commande exécutée jusqu'au CSV. »**
Piège : le jury suit UN chiffre de bout en bout ; la moindre hésitation sur le pipeline révèle que vous n'êtes pas l'auteur du process (source : ThesisLaunch Q13 « walk us through your data collection process » — « examiners want evidence that you actually did the work systematically » ; Komelate : « le jury a lu ton mémoire, maîtrise tes chiffres »).

**A2. « Vos archives sont "gelées" avec un SHA-256 : que prouve concrètement ce hash ? »**
Piège : l'intégrité cryptographique prouve que le fichier n'a pas changé, PAS que la mesure est valide — les confondre signale une incompréhension de la chaîne preuve (source : transposition de MockDefense « could your findings be an artefact of the method » ; le hash protège l'archivage, pas la validité).

**A3. « Qui a exécuté ces campagnes, sur quelle machine, quand ? Si je réinstalle votre banc demain, retrouve-je vos chiffres ? »**
Piège : question de reproductibilité ; une réponse vague (« c'est dans le mémoire ») est une non-réponse (source : Flent conçu précisément pour des mesures « repeatable » avec capture de métadonnées — flent.org ; ThesisLaunch Q14 fiabilité/validité).

**A4. « Comment avez-vous choisi vos profils fibre/4G/VSAT/Starlink ? Sur quelles mesures réelles reposent vos débits/RTT/pertes ? »**
Piège : des profils non sourcés = échantillonnage arbitraire ; tout le reste du mémoire hérite de ce défaut (source : MockDefense « how did you approach sampling, and why those criteria » — « don't pretend convenience sampling is something grander » ; 12-questions #3 « is your sample representative ? »).

### B. Méthodologie et validité

**B1. « Votre banc est émulé, pas le réseau malgache réel. Quelle est la valeur scientifique de vos conclusions ? »**
Piège : question de validité externe ; la mauvaise réponse est la négation (« l'émulation est fidèle »), la bonne borne la revendication (ce que netem peut/ne peut pas capturer) (source : MockDefense « artefact of the method » + limitation vs flaw ; bufferbloat FAQ objection 4.4 « the tests artificially add load » — réponse communautaire : « que se passe-t-il réellement quand vous téléchargez un fichier ? »).

**B2. « Pourquoi 120 secondes de charge ? D'où sort ce chiffre ? »**
Piège : la norme communautaire RRUL est 70 s (5 idle + 60 charge + 5 idle) ; s'écarter du standard sans justification = choix non délibéré exposé (source : bufferbloat.net RRUL Chart Explanation ; MockDefense : chaque décision méthodo doit avoir une justification en deux phrases).

**B3. « Pourquoi ne pas avoir utilisé Flent, l'outil de référence du domaine ? Vous avez réinventé la roue ? »**
Piège : « why this and not an alternative » ; défendre son outil maison sans citer/discuter Flent, RRUL, netperf, iperf3, betterspeedtest.sh = ignorance de l'état de l'art (source : 12-questions #7 « a member will often propose the [tool] they would have used » ; bufferbloat Tests page : Flent = « best » test, RRUL = standard de facto ; ThesisLaunch Q12).

**B4. « Avez-vous fait un pilote ? Comment avez-vous fixé vos paramètres (tailles de file, RTT cible, intensité de charge) ? »**
Piège : « pas eu le temps » est acceptable, « pas nécessaire » ne l'est pas — le jury distingue l'aveu honnête de la négligence assumée (source : ThesisLaunch Q16 « Did you conduct a pilot study? » ; lableaz « sample size explanation »).

### C. Statistiques et répétitions

**C1. « Vous avez UNE répétition par cellule. Pourquoi vous croire ? »**
Piège : LA question reine (« the flagship question ») ; la pire réponse est « c'est ce que j'ai pu faire » — il faut un raisonnement sur le minimum détectable et un bornage des conclusions (source : 12-questions #2 ; MockDefense « a smaller sample is not a fatal flaw if you are explicit about what it limits »).

**C2. « Quelle est la variance de vos mesures ? Un p95 sans dispersion, c'est un point sans barre d'erreur — comment savez-vous que 40 ms et 55 ms ne sont pas la même chose ? »**
Piège : rapporter un chiffre unique sans variabilité ; les box plots multi-runs sont la pratique standard du domaine (Flent box plots sur 4+ runs comparés) (source : bufferbloat RRUL Chart Explanation « Box Plot displays the result of multiple test runs » ; 12-questions #9 « do you report effect sizes, not just significance »).

**C3. « Pourquoi p95 et pas médiane, p99 ou RPM (round-trips per minute) ? »**
Piège : choix de métrique non justifié = critique facile et mortelle, car elle touche TOUTE la section résultats (source : transposition 12-questions #6/#7 « assumptions / why this analysis and not an alternative » ; bufferbloat Tests : l'écosystème a des métriques concurrentes — RPM d'Apple, CDF).

**C4. « Vous comparez N cellules × M configurations. Avez-vous corrigé pour les comparaisons multiples, ou cherchez-vous des différences qui n'existent que par hasard ? »**
Piège : question stats transposée ; « non » sans justification = « an easy criticism » (source : 12-questions #8 « no control without justification is an easy criticism »).

### D. Choix d'outils et IA

**D1. « Quelle part de l'IA générative a été utilisée dans ce travail — le code, le texte, les analyses ? »**
Piège : en 2026, le déni plat (« aucune ») est aussi suspect que l'aveu sans borne ; la question teste la transparence, pas la pureté (source : Compilatio (détecteurs IA/similitudes, contexte intégrité académique) ; LabLeaz : « discussions about academic transparency and integrity will likely become increasingly important » ; ThesisLaunch Q8 « how did your thesis evolve » — l'honnêteté sur le processus est valorisée).

**D2. « Pourquoi CUBIC et BBR, et pas [Westwood, L4S/BBR2, Hybla…] ? Que se passerait-il avec un autre contrôle ? »**
Piège : le membre du jury propose SON alternative préférée ; se braquer = perdre, ignorer l'alternative = perdre (source : 12-questions #7 : « do not get defensive: acknowledge the alternative, say what it adds and why your choice was reasonable » ; ThesisLaunch Q34 « could your data be interpreted differently »).

**D3. « Votre générateur de charge maison : comment savez-vous qu'il mesure ce qu'il prétend mesurer ? »**
Piège : question de validation d'instrument — un outil non comparé à un étalon (iperf3, netperf, RRUL) n'a aucune légitimité (source : 12-questions #4 « are your instruments validated » ; bufferbloat Tests : netperf/iperf2/iperf3 sont les étalons communautaires).

**D4. « Où se situe LE goulot d'étranglement sur votre banc ? Comment prouvez-vous que c'est bien lui qui gonfle, et pas un artefact de votre topologie ? »**
Piège : le bufferbloat est par définition un phénomène de goulot (fast link → slow link) ; si le goulot n'est pas démontré être là où vous placez l'AQM, vous ne mesurez rien (source : bufferbloat FAQ 2.3 « bufferbloat can occur anywhere there's a bottleneck » ; 3.4 la technique SQM crée un goulot interne — il faut pouvoir l'expliquer).

### E. Interprétation des résultats

**E1. « Vos chiffres contredisent-ils la littérature ? fq_codel/CAKE ont des décennies de résultats — où vous situez-vous ? »**
Piège : ne pas avoir de comparaison explicite avec la littérature (RFC 8290, papiers CoDel/CAKE) = « my findings align with X but diverge on Y » est attendu mot pour mot (source : ThesisLaunch Q29 « how do your findings compare to previous research » + Q22 « build on or challenge » ; bufferbloat.net cite RFC 8290 et les travaux fondateurs).

**E2. « Vous dites "CAKE réduit la latence de 60 %". Votre dispositif autorise-t-il ce langage causal ? »**
Piège : verbes causaux avec un design qui ne les supporte pas ; la réponse n'est pas de se rétracter mais de montrer qu'on distingue association/causalité (source : 12-questions #10 « you use causal language with a design that does not support it »).

**E3. « Quel résultat vous a surpris ? Comment l'expliquez-vous ? »**
Piège : « aucun, tout était attendu » = soit mensonge, soit absence de recul ; les surprises sont de l'or (source : ThesisLaunch Q28 « surprises are gold — they show you engaged deeply with data »).

**E4. « Que faudrait-il pour que vous admettiez que votre conclusion principale est fausse ? »**
Piège : question de falsifiabilité, plus fréquente qu'on ne croit ; aucune réponse = pas de maturité scientifique (source : ThesisLaunch Q47 « what would it take to convince you that your thesis is wrong »).

**E5. « Expliquez votre mémoire à un directeur d'opérateur non-technique en 60 secondes. »**
Piège : teste la sortie du jargon ; un doctorat/master se résume en question-méthode-résultat-enjeu (source : ThesisLaunch Q46 « explain your research to a non-expert in 60 seconds » ; LabLeaz PEEL).

### F. Impact opérationnel et généralisation

**F1. « Que fait la DSI de vos résultats lundi matin ? Donnez-nous UNE action concrète qu'elle peut exécuter. »**
Piège : dans les filières ingénierie, l'opérationnalité EST le critère ; « remplir un gap » ne veut rien dire — chaque recommandation doit être traçable à un résultat précis (source : JuryAI « école d'ingé : opérationnalité, qu'est-ce que ton travail apporte qu'on puisse utiliser » ; ThesisLaunch Q31 « who could use this? what decision might change? » + Q42 « recommendations traceable to a specific finding »).

**F2. « Le coût en ariary : c'est une vraie facturation ou une estimation ? Qui vous a donné ces tarifs ? »**
Piège : pseudo-précision — des chiffres présentés comme faits alors qu'ils sont modélisés ; l'aveu « estimation, bornée comme suit » est la seule sortie honorable (source : MockDefense measurement bounding « my estimates of workload should be read as upper bounds » ; ThesisLaunch Q33 calibrated confidence).

**F3. « Vos conclusions sur banc netem sont-elles généralisables aux liens d'accès malgaches ? Jusqu'où, exactement ? »**
Piège : le jury punit la sur-généralisation plus que l'aveu de limite (« a committee punishes overgeneralizing more than admitting a limit ») ; la réponse passe par la transférabilité (mécanisme plausible ailleurs, à retester) et non la généralisation statistique (source : 12-questions #3 ; MockDefense « transferability is not a weaker form of generalisation; it's a different kind »).

**F4. « Qui maintiendra ce code quand vous serez parti ? Votre outil est-il un livrable ou un prototype jetable ? »**
Piège : pérennité du livrable ; un prototype présenté comme solution pérenne sans plan de reprise = promesse non tenue (source : ThesisLaunch Q41 « next steps » / Q45 extension ; JuryAI « un corpus réutilisable » comme micro-contribution — le jury valorise le réutilisable, pas l'éphémère).

**F5. « Starlink : vous avez émulé un débit FIXE. Les liens LEO varient minute par minute — votre banc capture-t-il la seule difficulté qui compte ? »**
Piège : cake-autorate existe précisément parce que les liens 4G/5G/LEO varient en continu ; émuler à débit fixe et conclure sur Starlink réel = ignorance de la difficulté centrale, documentée par la communauté elle-même (source : bufferbloat FAQ 3.6 « cake-autorate handles links with varying rates » ; « my 5G cell connection is even worse — changing from minute to minute »).

---

## PARTIE 2 — Le top 10 des questions les plus récurrentes, simples et mortelles

Classement croisé : fréquence dans les sources × simplicité de formulation × dégâts si mal répondu. Les 10 de votre liste prioritaire sont toutes confirmées par au moins une source.

**1. « Vous avez UNE répétition par cellule. Pourquoi vous croire ? »**
Dangereuse parce que : question reine de toutes les soutenances (échantillon/répétitions = le front n°1 dans les 3 guides de comités) ; elle attaque la racine statistique de TOUTES les conclusions, et la réponse réflexe (« c'est ce que j'ai pu faire ») est exactement celle que les examinateurs identifient comme disqualifiante.
Bonne réponse (structure, pas contenu) : (a) reconnaître sans glisser — « une répétition par cellule, exactement » ; (b) borner : ce que n=1 interdit (comparaisons fines, inférence) vs ce qu'il permet (tendance d'ordre de grandeur si l'effet est massif) ; (c) mitiger : ce qui a été fait pour compenser (durée longue = nombreux échantillons de latence par run, gel des archives, p95 sur des milliers de RTT — un run de 120 s ≠ une observation) ; (d) refermer : « mes conclusions sont présentées comme exploratoires/bornées, pas comme définitives ; N répétitions est le premier travail futur ». Règle MockDefense : un petit échantillon n'est PAS un défaut fatal si on est explicite sur ce qu'il limite.

**2. « D'où viennent vos données ? »**
Dangereuse parce que : d'une simplicité désarmante ; le jury la pose en suivant du regard un chiffre précis du tableau, et elle vérifie trois choses d'un coup : l'autorité (avez-vous fait le travail), la traçabilité (la chaîne preuve existe-t-elle), et l'honnêteté.
Bonne réponse : raconter le pipeline en 4 étapes chronologiques (génération → capture → agrégation → gel), avec UN exemple chiffré déroulé devant eux ; avoir le hash, l'horodatage et la commande exacte sous la main. « Détails like "I transcribed each interview within 48 hours" demonstrate rigour » (ThesisLaunch) — l'équivalent réseau : « chaque CSV porte son SHA-256, le script de regeneration est X, voici le run ».

**3. « Votre banc est émulé, pas le réseau réel : quelle valeur ? »**
Dangereuse parce que : c'est la question de validité externe — la seule qui puisse, mal gérée, invalider le mémoire ENTIER d'une phrase (« donc tout votre travail est un artefact »).
Bonne réponse : ne JAMAIS nier la limite ; structure borner-mitiger-repositionner : (a) « l'émulation contrôle ce que le terrain ne permet pas de contrôler » (isoler la variable AQM) ; (b) ce que netem ne capture pas (variance de débit LEO, scheduling opérateur, shared-bottleneck) ; (c) la revendication exacte : « je démontre des mécanismes et des ordres de grandeur relatifs (fq_codel vs CAKE vs CUBIC), pas des chiffres absolus malgaches » ; (d) l'argument retourné de la communauté bufferbloat : « que se passe-t-il sur un vrai lien quand on sature réellement ? Exactement ce que mon banc reproduit — la charge artificielle est le point, pas le défaut » (bufferbloat FAQ 4.4).

**4. « Pourquoi ne pas avoir utilisé Flent / iperf3 ? »**
Dangereuse parce que : elle a une réponse factuelle dans le domaine — Flent est l'outil de référence conçu par les auteurs mêmes du domaine pour les tests de bufferbloat, avec RRUL comme protocole standard ; ne pas le connaître/discuter est un signal d'état de l'art non couvert, et « j'ai codé mon propre outil » sonne comme du bricolage si non défendu.
Bonne réponse : deux phrases de justification par décision (MockDefense) : ce que Flent apporte (répétabilité, RRUL, box plots multi-runs — le citer comme référence), et ce que votre outil ajoute que Flent n'a pas (les besoins spécifiques : échéance, QDI, tarification ariary, profils gelés, intégration bench) ; conclure « Flent valide le socle, mon outil l'étend au cas malgache ». Savoir aussi nommer netperf/iperf3/betterspeedtest.sh comme étalons (bufferbloat Tests page).

**5. « Quelle part de l'IA dans ce travail ? »**
Dangereuse parce que : en 2026 la question est presque systématique (integrité académique = tendance forte selon LabLeaz/Compilatio) ; elle ne poursuit pas la triche mais le mensonge — nier plat ou tergiverser détruit la crédibilité sur TOUTES les autres réponses.
Bonne réponse : inventaire factuel en trois tiroirs — code (généré/testé comment ?), rédaction (assistant pour style ?), analyse (les chiffres sont-ils calculés par vos scripts ou par une IA ?) ; puis la phrase clé : « chaque mesure vient des CSV gelés, régénérables par le script ; l'IA n'a touché ni les données ni les résultats ». La transparence détaillée transforme la question piège en preuve de rigueur.

**6. « Que fait la DSI de vos résultats lundi matin ? »**
Dangereuse parce que : dans une filière ingénierie/réseaux, l'opérationnalité est un critère de notation à part entière (JuryAI : école d'ingé = « qu'est-ce que ton travail apporte qu'on puisse utiliser ») ; une réponse abstraite (« améliorer la QoS à Madagascar ») révèle un travail sans ancrage.
Bonne réponse : UNE action, UNE cible, UNE trace : « activer fq_codel sur le CPE du lien VSAT du site X (testé ch. 6 : p95 de A à B ms), coût 0 ariary, réversible en 5 min — decision sheet p. Y ». Chaque recommandation traçable à un finding précis (ThesisLaunch Q42).

**7. « Qui maintiendra ce code après vous ? »**
Dangereuse parce que : elle teste si le livrable survit au départ de l'auteur — prototype jetable déguisé en solution = promesse non tenue, et le jury le sait.
Bonne réponse : plan de reprise en 3 lignes : documentation/README existants, dépendances (Go + netem standard, pas de stack exotique), proposition de repreneur (DSI/enseignant/étudiant suivant) ; recadrer honnêtement : « c'est un banc de recherche reproductible avant d'être un produit — le gel SHA-256 garantit que quiconque peut rejouer, pas dépanner ». La micro-contribution « corpus réutilisable » (JuryAI) est la bonne unité de valeur.

**8. « Le coût en ariary : c'est une vraie facturation ? »**
Dangereuse parce que : pseudo-précision — des montants monétaires présentés comme faits alors qu'ils sont modélisés sont une faute méthodologique (instrument de mesure non validé), et le jury y voit un test d'honnêteté.
Bonne réponse : dire la nature exacte de chaque chiffre (tarif public opérateur ? devis ? estimation maison ?), donner la direction d'erreur (« mes estimations sont des bornes supérieures/hautes », pattern MockDefense), et assumer : « c'est un modèle de coût décisionnel, pas une facture — son but est de comparer des options, pas de prédire la dépense au franc près ».

**9. « Pourquoi 120 secondes de charge ? »**
Dangereuse parce que : question d'une simplicité totale sur un paramètre fondateur ; la norme communautaire RRUL est 70 s (5+60+5) — tout écart non justifié expose TOUS les autres paramètres au même soupçon.
Bonne réponse : la justification doit pré-exister en deux phrases (règle MockDefense) : « le standard RRUL est 5+60+5 ; j'ai pris 120 s de charge pour [raison : laisser CUBIC sortir de slow-start et atteindre son régime / couvrir les cycles de variation 4G] ; j'ai vérifié sur les ECDF que la distribution p95 est stable sur la fenêtre ». Si la vraie réponse est « héritage du pilote », le dire et montrer que la stabilité a été vérifiée a posteriori.

**10. « Vos chiffres contredisent-ils la littérature ? »**
Dangereuse parce que : fq_codel/CAKE ont 10+ ans de littérature et des RFC (8290) ; une divergence inexpliquée = problème, une absence totale de comparaison = problème plus grave encore (le mémoire flotte hors du champ).
Bonne réponse : comparaison explicite en trois temps (ThesisLaunch Q29) : « mes résultats s'alignent avec [X] sur [Y], divergent de [Z] sur [W], la divergence s'explique par [différence de contexte : profil VSAT/charge maison vs RRUL netperf] » ; et nommer 3-5 références par cœur (RFC 8290, travaux CoDel/CAKE/OpenWrt) — les comités testent systématiquement l'engagement réel avec les sources.

---

## PARTIE 3 — Anti-patterns observés chez les étudiants qui s'écrasent (toutes sources)

**1. La défense agressive / se braquer.** Le sang monte, le ton se durcit face à « votre cadre est faible ». « Treating limitations questions as an attack telegraphs that you haven't thought carefully » (MockDefense) ; « ce qui plante les candidats, c'est de basculer dans la défensive agressive ou l'effondrement. Les deux sont mauvais signe. Ton directeur ne te défendra pas — ce n'est pas son rôle » (JuryAI). Contre-mesure : 2 secondes de silence + gorgée d'eau (invisible pour le jury), puis aïkido : reconnaître la part valide → nuancer → repositionner.

**2. La contrition préventive (sur-excuses).** S'excuser de tout, enchaîner 4 concessions sans bornage : « oui mon échantillon est petit, c'est un vrai problème, j'aurais dû… et je n'ai pas pu… » — à la fin le jury sait que vous regrettez le mémoire mais pas ce qu'il peut affirmer (MockDefense : « over-apologising », « itemised anxiety » ; l'exemple verbatim de la mauvaise réponse est exactement ça). Une limite est « a bounded claim about scope, not a confession of failure ».

**3. Le bluff / l'invention.** Répondre à côté plutôt que dire « je ne sais pas » : « Un jury respecte l'honnêteté bien plus que l'invention » (Komelate) ; « if you don't know, say so » (Scribbr). Le bluff se détecte en une relance, et détruit la crédibilité des réponses précédentes.

**4. Le directeur en bouclier / le choix par défaut.** « C'est la méthode qu'on m'a conseillée » = « posture désastreuse » (JuryAI) : vous transformez votre encadrant en auteur du mémoire. Idem « c'est ce qu'on m'a demandé » pour les paramètres. Contre-mesure : assumer chaque choix en son nom — « j'ai envisagé X, j'ai rejeté X parce que… ».

**5. La limite cosmétique.** « J'aurais aimé plus de répétitions / plus de temps » — « mou » (JuryAI) : tout le monde le sait, ça ne teste rien. La vraie limite est celle que vous nommez, qualifiez (impact précis sur quelle inférence) et bornez avec un « ce que je ne peux PAS affirmer » explicite.

**6. « Il n'y a pas de limite » / « j'ai tout traité ».** Pire réponse possible, disqualifiante : soit surconfidence soit confusion limite/défaut (MockDefense « claiming no limitations » ; JuryAI « tu perds tout »).

**7. L'effondrement / tout concéder.** Le jury pousse volontairement une limite (« donc vous ne pouvez rien affirmer ? ») pour voir si vous cassez : « examiner presses hard on a limitation to see whether the candidate will break » ; « conceding limits that collapse the core claim — you have done their job for them » (MockDefense). Si la limite réduit la portée mais pas la conclusion, il faut le dire et le tenir. « Answering every question with "I'd do it all differently" reads as lacking confidence » (MockDefense).

**8. Noyer le poisson / le discours flou long.** « Une réponse courte et juste vaut mieux qu'un long discours flou » (Komelate) ; réponses récitées par cœur qui s'effondrent dès qu'on reformule (JuryAI) ; jargon utilisé comme bouclier au lieu de répondre à la question posée — relue la question avant de répondre à celle qu'on a préparée (MockDefense : « the oral answer should be targeted, not a recitation »).

**9. La contrainte déguisée en vertu.** Présenter une limite d'accès/matériel comme un choix méthodologique noble (« un seul run par cellule pour garantir la pureté expérimentale ») : « never present a constraint as a methodological virtue it wasn't » (MockDefense). Nommer la contrainte comme contrainte est plus fort que l'habiller.

**10. Lister les limites en mitraille.** Six limites en 30 secondes sans bornage : signale qu'on ne sait pas trier ce qui compte, et donne au jury six fils à tirer (MockDefense « listing everything »). Deux-trois limites réelles, bornées, avec page de référence (JuryAI : « citer une page précise change tout »).

### Le squelette de bonne réponse commun à toutes les sources

Nommer (1 phrase, sans adoucissement) → Borner (ce que ça affecte / ce que ça n'affecte PAS) → Mitiger (ce que vous avez fait, ou pourquoi vous avez accepté) → Refermer la portée de la revendication (« j'affirme X dans le contexte Y, pas l'universel »).
= MockDefense (name/bound/mitigate/scope) = eloqole (admettre/situer/revenir) = JuryAI (reconnaître/nuancer/repositionner) = lableaz (PEEL). Quatre guides indépendants, une seule structure.
