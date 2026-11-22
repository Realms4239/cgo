# Labo step clarity — design (2026-11-22)

Authority: operator-approved Approach A 2026-11-22; `make-interfaces-feel-better` skill; ratified spatial spec §2. Scope is the Labo chapter headers and the Configurer declaration block only. No functional wiring, no new control.

## 1. Problem

Each Labo chapter opens with a good question ("Quel lien voulez-vous étudier ?", "Que se passe-t-il sur le lien ?"), but never says what the step produces or how to read its output. In Configurer, three `StateSlot`s read `site non déclaré`, `lien non déclaré`, `durée non fixée` with no explanation that these are honest states of the real report, not a broken form. A reader cannot tell what the step is for or what will happen next.

## 2. Goal

Every chapter states in one sentence what it gives the reader, and the declaration block says where its values come from. No value, state, seal or verdict changes.

## 3. Design

- `Chapter` gains an optional `outcome?: string` prop rendered as `<p className="chapter-outcome">` directly after the lede. `LaboView` supplies one real sentence per chapter, describing the method's output, never a result.
- `Configurer` gains one short line above the declaration `MetricRow`, stating that the values come from the real report and that an absent declaration keeps its word.
- `.chapter-outcome` and the Configurer context line use existing tokens only: `--fs-fine`, `--text-muted`, `text-wrap: pretty`. No new hex, font size, accent, motion, dependency or component file. The rail is unchanged.
- No new interactive control, so 44px is N/A. No new `aria-hidden` on meaningful text.

## 4. Honesty contract

- Outcome sentences describe what the reader will see; they never assert a measurement, PASS, seal, ranking or verdict. Absence words in Configurer stay exactly as they are.
- All existing data sources are untouched. No default value or fabricated number is introduced.

## 5. Accessibility

- Outcome and context are real text in DOM reading order; axe 2.2 AA, forced-colors and reduced-motion stay green at 890/1280/1920; no overflow or page scroll.

## 6. Verification

- DOM oracle: four chapters each expose an outcome sentence with no digit or verdict word; Configurer context line present; absence words intact.
- Existing `Chapter`/`PipelineRail`/`Observer`/`Comparer`/`Preuves` DOM tests and the whole frontend suite stay green; unmocked Edge Playwright at 890/1280/1920; screenshots read and validated; full gates once with the pre-existing `links` FAIL and `book/` SKIPs recorded, never edited.

## 7. Out of scope

Declaration inputs, store/API wiring, rail state words, chart/instrument work, backend, ledgers, frozen corpus, thesis, remote hosts, merge.
