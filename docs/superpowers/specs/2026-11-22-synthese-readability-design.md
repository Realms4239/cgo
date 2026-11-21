# Synthèse footer and ruler readability — design (2026-11-22)

Authority: operator-approved design 2026-11-22; `make-interfaces-feel-better` skill; ratified spatial spec `docs/research/2026-11-20-spatial-spec.md` §1.2. Scope is the field instrument only — `TimeField.tsx` footer and ruler markup plus their CSS. Nothing else in Synthèse or Labo changes in this slice.

## 1. Problem

The field footer currently renders roughly ten honesty fields as one run-on flex line of `--fs-micro` monospace text: `nom — non renseigné  RX — non mesuré  TX — non mesuré  fenêtre — non renseigné  pertes RX — non mesuré  pertes TX — non mesuré  source — non renseigné  écart — non renseigné`, followed by a separate equity line and provenance line. Every value is honest, but the reading order is flat and the eye cannot land on a group. The ruler legend has the same small, evenly-weighted typography.

## 2. Goal

Make the instrument skimmable at a glance without changing any measurement, state or honesty rule: a reader can see in one pass which source is configured, which measurements exist, and which are absent.

## 3. Design

- Three labelled groups — **Interface**, **Équité**, **Provenance** — each a compact label→value grid. Groups are separated by space and a quiet hairline, not by cards or background fills.
- Each group carries a short uppercase label using the existing `--tracking-caps` token. Values keep their current words; absent values keep `—` plus the existing French absence word.
- Live numeric values use `font-variant-numeric: tabular-nums`. Absence sentences use `text-wrap: pretty`. Group labels use `text-wrap: balance`. No new font size, hex, accent or dependency.
- The ruler legend keeps its glyph + word pairs but adopts the same label/value typography so eye movement from ruler to footer is continuous.
- No new interactive control, so the 44px rule is N/A. No motion is added; existing `animation: false` and reduced-motion behaviour are untouched.

## 4. Honesty contract

- All values still come from `live.iface`, the JFI source and trace provenance. No default value, no zero, no green, no fabricated source.
- Missing or invalid input still renders `—` + a French absence word. The regrouping is presentation only.
- The deterministic ruler alignment contract (`containLabel: false`, grid `left:64 / right:16`, CSS track inset derived from those values) must remain intact and its DOM oracle green.

## 5. Accessibility

- Labels and values remain real text in reading order; no CSS-generated content carries meaning.
- WCAG 2.2 AA, forced-colors and reduced-motion rules stay green at 890/1280/1920. No page overflow or horizontal scroll.

## 6. Verification

- DOM oracle: three labelled groups exist; each absence value renders `—` + word; no digit or fabricated-green class appears for absent values; no existing measurement field is lost.
- Existing `TimeField.spatial.dom.test.tsx`, all redesign/a11y/layout/feel and fullstack Playwright specs stay green.
- Unmocked screenshots at 890×512, 1280×768 and 1920×1080, read and validated.
- Full gates once; the pre-existing `links` false positive and `book/` SKIPs are recorded, never edited.

## 7. Out of scope

Labo step clarity, verdict/action hierarchy, chart geometry, ring buffers, SSE, backend, gates, ledgers, frozen corpus, thesis, remote hosts. A separate slice is required for each.
