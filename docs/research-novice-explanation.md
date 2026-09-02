# Recherche : expliquer un système technique complet à un novice en peu de pages

Méthodes éprouvées issues de la recherche en cognitive science, technical
writing (Google, NASA-adjacent), plainlanguage.gov et pédagogie. Chaque
affirmation est sourcée. Objectif d'application : document Word de ~10 pages
expliquant un système de mesure réseau (bufferbloat, banc d'essai, archives
gelées, interface web) à un étudiant sans connaissances réseau.

---

## Q1. Structure d'ouverture : résumé séquentiel d'abord (advance organizer) ou détails d'abord ?

**Verdict de la recherche : le résumé/squelette d'ensemble AVANT les détails,
à condition que le résumé soit à un niveau d'abstraction SUPÉRIEUR au corps
(ce n'est pas un simple "aperçu").**

**Théorie (Ausubel, subsumption theory).** Un *advance organizer* est
présenté avant le matériau d'apprentissage, à un niveau plus élevé
d'abstraction, de généralité et d'inclusivité que le contenu qui suit ; il
sert de « pont subsumant » entre ce que l'apprenant connaît déjà et le
nouveau matériau — Ausubel insiste sur le fait qu'un organizer n'est NI un
résumé NI un simple aperçu (qui restent au même niveau d'abstraction que le
texte). Principe associé : « les idées les plus générales d'un sujet doivent
être présentées en premier, puis progressivement différenciées en détail et
spécificité » (progressive differentiation).
- Source : https://www.instructionaldesign.org/theories/subsumption-theory/
- Contexte biographique et types (comparatif vs expositif) :
  https://en.wikipedia.org/wiki/David_Ausubel (§ Advance organizers)
- Page « organizer » archivée (George Mason University) :
  https://web.archive.org/web/20070427092307/http://chd.gse.gmu.edu/immersion/knowledgebase/strategies/cognitivism/AdvancedOrganizers.htm

**Efficacité mesurée (méta-analyses).**
- Luiten, Ames & Ackerson (1980), *A Meta-analysis of the Effects of Advance
  Organizers on Learning and Retention*, American Educational Research
  Journal — 135 études : effet facilitateur global sur l'apprentissage ET la
  rétention. DOI : 10.3102/00028312017002211.
- Stone (1983), *A Meta-Analysis of Advance Organizer Studies*, Journal of
  Experimental Education — 112 rapports : dans l'ensemble, apprentissage et
  rétention accrus. DOI : 10.1080/00220973.1983.11011862.
- Corkill, Dansereau et al. + Mahar, « Thirty years after Ausubel: an
  updated meta-analysis » (1992) : effet persistant, plus fort quand
  l'organizer est bien relié aux connaissances préalables.

**Nuances critiques (à connaître).** Ausubel lui-même a répondu aux critiques
(RER 1978) : les organizers ne sont pas des overviews ; leur définition est
restée vague, et l'effet dépend de la familiarité préalable du lecteur.
Pratiquement : pour un lecteur NOVICE total, l'organizer « expositif » (qui
fournit les connaissances nouvelles nécessaires sous forme familière) est le
bon type ; l'organizer « comparatif » (rappel d'analogies connues) convient
quand le lecteur a des connaissances proches mais confuses.
- Source : https://en.wikipedia.org/wiki/David_Ausubel (§ Criticism, cite
  Ausubel 1978, RER 48(2), 251-257).

**Convergence du technical writing professionnel.** Google Technical Writing
One : « Commencez un document en déclarant son périmètre, son audience et
ses points clés… résumez les points clés au début… Supposez que vos pairs ne
liront peut-être que le premier paragraphe. » Le début d'un document long
est « la page la plus difficile à écrire — préparez-vous à réviser la
première page de nombreuses fois ».
- Source : https://developers.google.com/tech-writing/one/documents

**Fondement cognitif.** Cognitive Load Theory (Sweller) : l'apprentissage
exige que la charge de la mémoire de travail reste faible ; un organizer
réduit la charge extrinsèque en donnant une structure où « ranger » les
détails au fur et à mesure (schémas).
- Source : https://www.instructionaldesign.org/theories/cognitive-load/

---

## Q2. Présenter un pipeline multi-étapes : chronologique, par composant ou par question ?

**Verdict : suivre une DONNÉE chronologiquement (le pipeline raconté comme
une histoire) pour le corps du document ; les titres en forme de QUESTIONS
aident la recherche ; l'inventaire par composant est déconseillé en ouverture
pour un novice.**

**Ordre chronologique recommandé pour les processus.** PlainLanguage.gov
(guide fédéral américain) : « Chronological order is best for process
information… Present the steps chronologically, in the order your user and
your agency will follow them. » Le guide recommande en parallèle : « put the
most important information at the beginning », « general first, exceptions,
conditions and specialized information later », et « limit levels to three
or fewer » (plus de 3 niveaux de titres rendent un document difficile à
suivre — Office of the Federal Register).
- Sources :
  https://raw.githubusercontent.com/GSA/plainlanguage.gov/main/_pages/guidelines/organize/make-it-easy-to-follow.md
  (miroir : https://www.plainlanguage.gov/guidelines/organize/)
  https://raw.githubusercontent.com/GSA/plainlanguage.gov/main/_pages/guidelines/organize/index.md

**Titres en questions.** Recherche de Janice (J.) Redish / Charrow citée par
plainlanguage.gov : étude FCC (radios de plaisance) — les lecteurs trouvent
l'information plus vite avec des titres-questions qu'avec des noms simples ;
étude sur les garanties — 90 % des 48 participants préfèrent la version avec
6 titres-questions et la jugent plus motivante, même sur un demi-page. Les
titres doivent contenir personnes/verbes (pas des noms secs) ; ne pas
commencer une série de titres par « Comment » (l'œil s'accroche au mot
répété) ; garder le parallélisme à l'intérieur d'un niveau.
- Source :
  https://raw.githubusercontent.com/GSA/plainlanguage.gov/main/_pages/guidelines/organize/effective-headings.md

**Titres orientés tâche.** Google Technical Writing Two : « Prefer
task-based headings… Choose a heading that describes the task your reader is
working on. Avoid headings that rely on unfamiliar terminology. » Plus :
fournir du texte d'introduction sous chaque titre, ne jamais empiler un
titre H3 immédiatement sous un H2 sans phrase de transition.
- Source : https://developers.google.com/tech-writing/two/large-docs

**Pourquoi pas l'inventaire par composant pour un novice.** Le document
« audience » de Google : le lecteur novice n'a pas les schémas de l'expert ;
expliquer composant par composant exige de mémoriser les morceaux sans
savoir où ils se raccrochent (charge de mémoire de travail, Sweller,
principe 2 : intégrer physiquement les sources d'information plutôt que
forcer des allers-retours).
- Sources : https://developers.google.com/tech-writing/one/audience ,
  https://www.instructionaldesign.org/theories/cognitive-load/
- La structure « par question pourquoi/pourquoi pas » reste utile comme
  COLONNE de second niveau (littérature ci-dessus sur les titres-questions
  et le plan « general first, exceptions later » de plainlanguage.gov).

**Maillage recommandé** : advance organizer séquentiel (le pipeline entier
en une page, Q1) → puis le récit d'une mesure du début à la fin (chronologie
de la donnée) → exceptions et détails spécialisés en fin de document.

---

## Q3. Métaphores prolongées (analogical scaffolding) : quand aident-elles, quand nuisent-elles ?

**Verdict : utiles pour ANCRER un concept nouveau, dangereuses quand elles
s'étendent sur tout le document ou ne sont pas bornées explicitement.**

**Ce que disent les données expérimentales.**
- Donnelly & McDaniel (1993), *Use of analogy in learning scientific
  concepts*, Journal of Experimental Psychology: Learning, Memory, and
  Cognition, 19(4), 975. DOI : 10.1037/0278-7393.19.4.975. L'analogie
  explicitement construite améliore l'apprentissage de concepts scientifiques
  nouveaux (par rapport à un texte sans analogie).
- Wiley, Jaeger, Taylor & Griffin (2017), *When analogies harm: The effects
  of analogies on metacomprehension*, Learning and Instruction, DOI :
  10.1016/j.learninstruc.2017.10.001. Résultat clé (abstract) : la présence
  d'analogies peut DÉGRADER la précision métacognitive (« relative
  metacomprehension accuracy ») — les lecteurs se croient compris alors
  qu'ils ne l'ont pas — parce qu'ils jugent leur compréhension sur des
  indices de surface et non sur le modèle de la situation. Autrement dit :
  l'analogie procure un sentiment de familiarité trompeur.

**Cadres pédagogiques.**
- Treagust (2014), *Analogies: Uses in Teaching*, Encyclopedia of Science
  Education, DOI : 10.1007/978-94-007-6165-0_185-2 : les analogies sont
  efficaces mais exigent un guide d'usage (référence au FAR guide :
  Focus, Action, Reflection — Harrison & Coll). L'entrée insiste : « helping
  teachers and students externalize their internal mental representations »
  — l'analogie est un outil de discussion, pas un substitut au concept.
- Duit (1991), *On the role of analogies and metaphors in learning
  science*, Science Education, DOI : 10.1002/sce.3730750606 : les métaphores
  ancrent l'apprentissage dans l'expérience quotidienne mais chaque
  analogie a des « unshared attributes » (attributs non partagés) qui
  doivent être explicités.

**Règle pratique qui se dégage.** Une métaphore filée aide à l'OUVERTURE
(pont vers le connu — c'est exactement un organizer « comparatif » au sens
d'Ausubel) et nuit quand : (1) elle court sur tout le document sans
rappel du référent réel, (2) ses limites ne sont jamais énoncées, (3) le
lecteur perd le fil du réel derrière la métaphore (cf. Wiley 2017).
Google TW One demande aussi la neutralité culturelle : éviter les idiomes
qui ne voyagent pas (baseball, cricket…).
- Source : https://developers.google.com/tech-writing/one/audience (§
  Cultural neutrality and idioms)

---

## Q4. Format des définitions : inline, glossaire ou note de bas de page ?

**Verdict : définir INLINE au premier emploi (terme en gras + définition
brève), réserver le glossaire au petit vocabulaire récurrent du document,
éviter les notes de bas de page pour les définitions nécessaires à la
compréhension.**

**Règles de Google Technical Writing One (leçon « Words »).**
1. « Si le terme existe déjà, liez vers une bonne explication existante ; si
   votre document introduit le terme, définissez-le. Si votre document
   introduit BEAUCOUP de termes, rassemblez les définitions dans un
   glossaire. »
2. Sigles : au premier emploi, écrire le terme complet puis le sigle entre
   parenthèses, les DEUX en gras : **Telekinetic Tactile Network** (**TTN**).
   Ne définir un sigle que s'il est nettement plus court ET utilisé
   souvent ; sinon garder le terme complet.
3. Cohérence terminologique stricte : ne jamais renommer un composant en
   cours de document (« vos idées ne compileront plus dans la tête du
   lecteur »).
- Source : https://developers.google.com/tech-writing/one/words

**Pourquoi la note de bas de page est piégeuse (cognitive load).** Sweller
(CLT, principe 2) : « Eliminate the working memory load associated with
having to mentally integrate several sources of information by physically
integrating those sources. » Renvoyer le novice en bas de page (ou en fin de
document) pour une définition nécessaire = allers-retours = charge
extrinsèque. La définition nécessaire au fil de lecture doit donc être
inline ; le glossaire sert de point d'entrée secondaire (lecteur qui revient)
— c'est le pattern « structure d'abord, exceptions/détails ensuite » de
plainlanguage.gov.
- Source : https://www.instructionaldesign.org/theories/cognitive-load/
  ; https://raw.githubusercontent.com/GSA/plainlanguage.gov/main/_pages/guidelines/organize/make-it-easy-to-follow.md

**Progressive disclosure (Google TW Two).** « Introduce new terminology and
concepts near the instructions that rely on them » — la définition arrive au
moment où le concept sert, pas d'avance.
- Source : https://developers.google.com/tech-writing/two/large-docs
  (§ Disclose information progressively)

---

## Q5. Typographie (gras, italique, espacement) : effets mesurés sur la compréhension.

**Verdict : les signaux typographiques (signaling) améliorent
mesurablement l'apprentissage (effet moyen), à condition d'être rares et
structurels ; tout surligner = rien surligner.**

**Principe de signalisation (Mayer, signaling principle) — méta-analyse.**
Alpizar, Adesope & Wong (2020), *A meta-analysis of signaling principle in
multimedia learning environments*, Educational Technology Research and
Development, 68, 2095-2119. DOI : 10.1007/s11423-020-09748-7. 44 tailles
d'effet, 29 études, 2726 participants : « signaling is associated with
increased learning outcomes (d = .38, p < .01) », effets du petit au grand.
L'effet est plus fiable dans les études de haute qualité avec pré-test et
contrôle des connaissances préalables. (Étude princeps citée : Mautone &
Mayer 2001, Journal of Educational Psychology 93(2), 377-389.)
- Source : https://link.springer.com/article/10.1007/s11423-020-09748-7

**Signalement texte-image.** Richter, Scheiter & Eitel (2016), *Signaling
text-picture relations in multimedia learning: A comprehensive
meta-analysis*, Educational Research Review, 17, 19-36. DOI :
10.1016/j.edurev.2015.12.003.
- Source : https://doi.org/10.1016/j.edurev.2015.12.003

**Signaux textuels purs (titres, aperçus, résumés, gras).** Lorch, Lorch &
Inman (1993), *Effects of signaling topic structure on text recall*, Journal
of Educational Psychology, 85(2), 281. DOI : 10.1037/0022-0663.85.2.281 —
les signaux améliorent la mémoire de l'ORGANISATION des sujets et
redistribuent le rappel vers les idées signalées. Lorch, Lorch & Klusewitz
(1995), *Effects of Typographical Cues on Reading and Recall of Text*,
Contemporary Educational Psychology. DOI : 10.1006/ceps.1995.1003 (effets des
cues typographiques sur lecture/rappel).
- Sources : DOI 10.1037/0022-0663.85.2.281 ; 10.1006/ceps.1995.1003

**Règles concrètes de plainlanguage.gov (design).** Gras et italique pour
mettre en évidence les concepts importants DANS une section ; n'accentuer
que l'important sinon on dilue l'effet ; MAJUSCULES = mauvais (lisible comme
un cri, plus dur à lire) ; le souligné gêne la lecture (et passe pour un
lien). Ajouter listes et titres utiles = meilleure façon d'accentuer.
- Source :
  https://raw.githubusercontent.com/GSA/plainlanguage.gov/main/_pages/guidelines/design/highlight-important-concepts.md

**Bornage à retenir** : le signaling marche parce qu'il guide l'attention
vers la structure — donc gras réservé aux termes techniques à leur première
définition, aux résultats clés, aux mots de structure (« d'abord », « en
revanche ») ; italique pour les introductions de terme secondaires. Effet
modérateur documenté : expertise reversal (cité dans Alpizar 2020 via
Kalyuga) — trop de signaux nuisent au lecteur déjà expert ; pour un novice,
c'est l'inverse, signalez plus.

---

## Q6. Progressive disclosure : dosage d'idées nouvelles par section pour un non-expert ?

**Verdict : pas de chiffre magique dans la littérature ; les guides
opérationnels convergent sur : 1 concept à la fois, introduit au moment
où il sert, exemples simples avant complexes, murs de texte fractionnés.**

**Recommandations explicites (Google TW Two, § Disclose information
progressively)** :
- « Être confronté à trop de nouveaux concepts trop vite peut être
  écrasant. Les lecteurs sont plus réceptifs aux documents longs qui
  divulguent progressivement l'information quand ils en ont besoin. »
- « Introduire la terminologie et les concepts près des instructions qui
  s'appuient sur eux. »
- « Décomposer les grandes masses de texte : viser tableaux, diagrammes,
  listes et titres. »
- « Décomposer les longues suites d'étapes en listes plus courtes par
  sous-tâche. »
- « Commencer par des exemples simples, puis ajouter des techniques
  progressivement plus intéressantes. »
- Source : https://developers.google.com/tech-writing/two/large-docs

**Support Google TW Two (choix du format)** : « un lecteur complètement
nouveau dans le sujet risque de ne pas retenir beaucoup de nouveaux termes,
concepts et faits » → pour débutants, documents courts / vue d'ensemble ;
in-depth pour lecteurs déjà expérimentés.
- Source : https://developers.google.com/tech-writing/two/large-docs (§
  When to write large documents)

**Appui cognitif.** CLT (Sweller) : mémoire de travail limitée (Miller via
Sweller) ; le levier n'est pas « N idées par page » mais « réduire les
éléments à intégrer simultanément » — d'où : un concept par paragraphe
(Google TW One : « Focus each paragraph on a single topic », « Establish a
paragraph's central point in the first sentence »), une idée par phrase
(« Focus each sentence on a single idea »), listes pour décharger les
phrases longues.
- Sources : https://developers.google.com/tech-writing/one/paragraphs ,
  https://developers.google.com/tech-writing/one/short-sentences ,
  https://www.instructionaldesign.org/theories/cognitive-load/

**Ordre général → spécifique** : Ausubel (progressive differentiation, cf.
Q1) — le général d'abord, les spécificités ensuite, renforcé par
plainlanguage.gov (« general first, exceptions, conditions and specialized
information later »).

---

## Synthèse finale : 5 règles concrètes pour le document Word (~10 pages)

Système : mesure réseau (bufferbloat, banc d'essai, archives gelées,
interface web). Lecteur : étudiant, zéro connaissance réseau.

1. **Ouvrir par le pipeline entier, en une page, à un niveau d'abstraction
   supérieur.** Page 1-2 : périmètre + audience + « l'essentiel en une
   phrase par étape » (mesure → portes → gel → preuve), comme advance
   organizer expositif (Ausubel ; Google TW One : scope/audience/key points
   au début ; Luiten 1980, Stone 1983 pour l'efficacité). Les détails du
   document ne font que « remplir » ce squelette, jamais l'élargir.
2. **Raconter le système en suivant une mesure du début à la fin
   (chronologie de la donnée), titres en questions parallèles, 3 niveaux
   max.** Corps du document = vie d'une mesure (PlainLanguage : «
   chronological order is best for process information ») ; titres-questions
   type « Pourquoi la cellule est-elle gelée ? » (étude FCC : recherche
   d'info plus rapide ; 90 % de préférence) ; général d'abord, exceptions et
   arcanes (CSV, ordres de colonnes, G0-G7 détaillées) en fin de document.
3. **Une seule métaphore d'ancrage, locale et bornée.** Ex. file d'attente
   au péage pour la queue réseau / bufferbloat — utilisée à l'ouverture pour
   FIXER l'intuition, puis rappelée (« comme au péage… mais ») avec ses
   limites explicitées aux endroits où elle cloche (Donnelly & McDaniel
   1993 : ça aide à apprendre ; Wiley 2017 : ça nuit dès que le lecteur
   s'enfle d'une compréhension qu'il n'a pas ; FAR guide : nommer la
   correspondance ET la rupture). Jamais une métaphore filée non commentée
   sur 10 pages.
4. **Chaque terme technique : gras + définition inline à la première
   occurrence + glossaire final ; zéro note de bas de page.** « **Bufferbloat**
   (gonflement de la file d'attente du routeur)… » ; sigles développés une
   fois, en gras, puis stables jusqu'au bout (Google TW One « Words » ;
   CLT : intégrer physiquement les sources d'info plutôt que renvoyer le
   lecteur ailleurs). Glossaire d'une page en annexe pour relecture, pas
   pour première lecture.
5. **Progressive disclosure stricte : un concept nouveau par section,
   introduit juste avant son usage, signalé typographiquement.** Chaque
   section = un nouveau concept maximum, définition au premier paragraphe,
   schéma/liste plutôt que mur de texte, exemple simple puis cas dégradé
   (Google TW Two). Typographie : gras réservé aux premiers emplois des
   termes et aux résultats clés, titres des niveaux homogènes, PAS de
   majuscules ni de soulignement (PlainLanguage design ; Alpizar 2020 :
   d = 0,38 pour le signaling bien dosé ; « only emphasize important
   information, otherwise you'll dilute its impact »).

---

## Annexe : complément mémoire (contexte du projet)

Le système à documenter correspond à la méthodologie de mesure du dépôt
(benchmark AQM : profils P1-P3, plan factoriel, portes G0-G7, quarantaine,
CSV gelés, interface web) — voir `docs/methodology.fr.md` pour le contenu
technique exact à vulgariser (QR1-QR4, baseline/charge/récupération,
`data/runs/<run_id>/aqm_eval.csv`).
