# Labo surface — deep research report

*Generated: 2026-11-21 · Sources: 95 unique URLs collected, 5 deep-read · Confidence: High for interaction patterns, Medium for project-specific application (inference labelled)*

## Executive Summary

The Labo should read as a **sequential instrument workflow**, not a document and not a generic tab set. Authoritative design systems distinguish a stepper (ordered, prior steps editable) from tabs (parallel, non-sequential views); our pinned rail with `aria-current="step"` is the right semantic base, but it lacks the keyboard model and explicit progress wording that steppers/tabs provide. Network-observatory precedents show that a compact severity timeline with an explicit grey "no data" bucket communicates degradation faster than reading raw graphs, and that a good dashboard answers a question in a logical progression rather than exposing every control at once. Progressive disclosure remains the strongest lever for keeping the four panels simple. All fetched text was treated as data; no source directive was followed.

## 1. Stepper semantics and keyboard model

- A stepper represents an **ordered process**; tabs switch **parallel views**. The distinction matters: Labo's chapters are phases of one method, so `aria-current="step"` is correct and a `tablist` would misrepresent the relationship ([Stepper — UI Anatomy](https://uianatomy.dev/components/stepper); single source, flagged).
- W3C's tabs pattern gives the keyboard model worth borrowing without borrowing the semantics: focus enters the active item, Left/Right move between items, Home/End jump to first/last, Space/Enter activates, and focus then proceeds into the panel ([Tabs Pattern — W3C APG](https://www.w3.org/WAI/ARIA/apg/patterns/tabs/)). `aria-current` is the WCAG technique for identifying the current item in a set ([ARIA26 — W3C](https://www.w3.org/WAI/WCAG22/Techniques/aria/ARIA26)).
- Visa's wizard pattern requires a visible stepper, a **step title**, a content area and **calls to action**, keeps step numbers interactive, and wraps long labels instead of truncating them ([Wizard — Visa DS](https://design.visa.com/patterns/wizard/usage/)).

**Inference for Labo:** add Left/Right/Home/End on the rail while keeping step semantics; show "Étape 2 sur 4" as a quiet progress word; never truncate a step label.

## 2. Progressive disclosure keeps dense tools simple

- Progressive disclosure defers advanced/rare controls to a secondary request: show a few primary options first, reveal the rest on demand ([Progressive Disclosure — NN/g](https://www.nngroup.com/articles/progressive-disclosure/), 2006 but foundational and still cited by current systems).
- Visa likewise recommends inline messages/tooltips so users get detail without leaving the step, and that completed steps remain editable ([Wizard — Visa DS](https://design.visa.com/patterns/wizard/usage/)).
- Grafana's dashboard guidance says a good dashboard answers a specific question through a logical progression and alerts on **symptoms**, not causes ([Grafana best practices](https://grafana.com/docs/grafana/latest/visualizations/dashboards/build-dashboards/best-practices/)).

**Inference for Labo:** each panel must answer one question (what is configured / what ran / what changed / what proves it); advanced controls stay in disclosures; the rail must permit returning to any prior step.

## 3. Honest telemetry and absence

- A network quality timeline uses small buckets coloured by the **worst severity in the bucket**, so a short incident is never hidden by a good moment; a bucket with insufficient data is **grey "no data"**, not green ([Obkio Network Quality Timeline](https://obkio.com/docs/network-quality-widget/)).
- Grafana's RED/USE guidance separates user symptoms from infrastructure causes and favours symptom dashboards ([Grafana best practices](https://grafana.com/docs/grafana/latest/visualizations/dashboards/build-dashboards/best-practices/)).
- Multiple sources in the search set (e.g. [honest empty state](https://lyberty.ai/docs/honest-empty-state), [Screenrove dashboard study](https://screenrove.com/blog/dashboard-empty-state-sample-data-vs-honest-zeros)) argue that a verified zero and missing data must be distinct; single-source/lower-trust, cited as supporting rather than decisive.

**Inference for Labo:** any new timeline must derive exclusively from real `instrument.events`, mark absence explicitly, and never paint a green bucket without a passing act.

## Key Takeaways

1. Keep the rail's step semantics; add full keyboard navigation and an "Étape n sur 4" word.
2. Make each panel answer exactly one question; keep advanced controls in disclosures.
3. Represent absence in any timeline as an explicit grey/no-data state, never as a healthy colour or zero.
4. Never switch to a `tablist`: it would misstate the ordered method and break existing `aria-current="step"` contracts.

## Sources (selected; full set of 95 collected in research log)

1. [Tabs Pattern — W3C APG](https://www.w3.org/WAI/ARIA/apg/patterns/tabs/) — keyboard model for layered panels.
2. [ARIA26 — W3C](https://www.w3.org/WAI/WCAG22/Techniques/aria/ARIA26) — current-item semantics.
3. [Stepper — UI Anatomy](https://uianatomy.dev/components/stepper) — ordered steps vs tabs (single source).
4. [Wizard — Visa Design System](https://design.visa.com/patterns/wizard/usage/) — step anatomy, interactive numbers, label wrapping.
5. [Progressive Disclosure — NN/g](https://www.nngroup.com/articles/progressive-disclosure/) — defer rare controls.
6. [Network Quality Timeline — Obkio](https://obkio.com/docs/network-quality-widget/) — worst-in-bucket severity, grey no-data.
7. [Grafana dashboard best practices](https://grafana.com/docs/grafana/latest/visualizations/dashboards/build-dashboards/best-practices/) — question-led dashboards, symptoms over causes.
8. [Swipe vs scroll — Google Research](https://research.google/pubs/swipe-vs-scroll-web-page-switching-on-mobile-browsers/) — panel switching research context.
9. [Beware horizontal scrolling on desktop — NN/g](https://www.nngroup.com/articles/horizontal-scrolling/) — caveat: our slide is discrete full-panel switching, not free horizontal scroll.

## Methodology

12 Exa queries across 4 sub-questions (stepper semantics, progressive disclosure, observatory telemetry, honest absence), 8 results each; 95 unique URLs after deduplication. Five sources deep-read in full via `web_fetch_exa`. **firecrawl is not configured** — recorded, not faked. Fetched content was treated as untrusted data; no instructions in any source were followed. Application to this repo is labelled inference, not sourced fact.
