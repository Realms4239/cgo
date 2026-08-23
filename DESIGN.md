# CGO — Design System (Experimental Observatory)

> Authority for the LIEN instrument UI (`web/frontend`). Written from the built
> product, per `docs/SPEC.md` §3.3. The previous BMW M marketing document is
> replaced; its useful identity decisions survive here.

## 1. Product truth

CGO is a self-contained Go instrument that audits constrained access links and
evaluates AQM/BBR policies for Météo Madagascar. Its interface is an
**experimental observatory**: a dark instrument panel where a jury or operator
reads live evidence at a glance. Mode: **Operate** — scanability and truthful
state outrank expression; brand lives in precise details.

The one memorable element is the **telemetry sidebar**: navigation above a live
campaign status block above the G0–G7 gate strip. Everything else stays quiet.

## 2. Color

Near-black canvas, hairline separation, no gradients, no rounded cards.

### Surfaces
| Token | Value | Use |
|---|---|---|
| `--canvas` | `#070707` | page floor |
| `--surface-soft` | `#0b0b0c` | header, sidebar, footer |
| `--surface-card` | `#101012` | panels |
| `--surface-elevated` | `#161618` | nested elements, banners |

### Hairlines & text
| Token | Value |
|---|---|
| `--hairline` `--hairline-strong` `--hairline-faint` | `#26262a` `#3a3a40` `#141416` |
| `--text-primary` → `--text-faint` | `#f2f2f4` `#d6d8dd` `#a9aeb6` `#8b9099` `#767b84` |

### Identity vs telemetry (the hard rule)
The **M tricolor** (`#0066b1` `#1c69d4` `#e22718`) marks identity only:
wordmark stripe, favicon. It is never a button, never a data color.

Telemetry colors are semantic and separate:

| Token | Value | Meaning |
|---|---|---|
| `--t-live` | `#5ad3e3` cyan | primary/live series, active nav |
| `--t-threshold` | `#f4b400` amber | thresholds, caution, degraded |
| `--t-danger` | `#e22718` red | drops, errors, failed gates |
| `--t-ok` | `#1fa348` green | passed, safe, best config ★ |
| `--t-bbr` | `#b48ae0` violet | bulk goodput series |
| `--t-neutral` | `#9aa3ad` steel | neutral series |

State is **never color-only**: every gate pip pairs with a text label
(`PASS/FAIL/—`) in Campagne; banners carry words (`CHARGE`, `OFFLINE`).

## 3. Typography

Self-hosted woff2 only (`public/fonts`).

| Role | Face | Use |
|---|---|---|
| Display | Cormorant Garamond 600, uppercase, tracked +0.04–0.06em | wordmark, view titles, chart titles |
| Body/nav | Inter var 400–500 | labels, copy |
| Data | JetBrains Mono 400–700, tabular | measurements, provenance, tables, keys, buttons (11px, +0.12em uppercase) |

Display sizes stay ≤20px in-app; hierarchy comes from face contrast
(serif display vs mono data), not scale.

## 4. Layout

CSS grid shell: `48px header / [sidebar 232px · main] / 28px footer`.
Sidebar collapses conceptually to icon width on narrow rails; below 900px the
sidebar becomes a horizontal strip and panels stack into readable bands —
no horizontal overflow anywhere. Spacing on a 4px base (`--sp-1…--sp-5`);
panels max-width 960px inside main; more space above headings than below.
Geometry is sharp: radius 0 everywhere except nothing — circles are forbidden
except the connection dot glyph.

## 5. Components

- **Sidebar** — nav buttons (`data-panel`), keys `1–4`; status block
  (`phase/event/SSE`); 8-cell gate strip with `data-gate`.
- **Card** — `--surface-card`, 1px hairline border, `card-head` eyebrow in
  mono caps. One level only; nested cards are wrong.
- **ArmButton** — arm-then-fire double confirm (`DÉMARRER` → `CONFIRMER ?`,
  5 s timeout). Primary = white fill/dark text; armed = danger fill. All
  mutating actions use it; read-only actions stay plain links/buttons.
- **Charts** — ECharts canvas via `chartGrammar.baseOption`: transparent bg,
  JetBrains Mono 10px axes in `#8b9099`, splitlines `#141416`, crosshair
  tooltip on `--surface-elevated`. Series: p50 cyan, p95 green, small-p95
  amber, goodput violet with 25%-opacity area. `animation: false` (reduced
  motion honored by construction).
- **Tables** — Résultats uses mono 12px rows, hairline row separators, best
  cell tinted `rgba(31,163,72,.08)` + ★.
- **Banner** — full-width state line (LIVE/BASELINE/CHARGE/OFFLINE) in mono
  caps on `--surface-elevated`.
- **Figures** — backend SVGs share the same palette (`#070707` bg, Cormorant
  title, mono labels); served from `/api/figures/*`.

## 6. Motion

None beyond hover/focus transitions at `80–150ms`. Charts do not animate.
`prefers-reduced-motion` collapses durations to 1ms via token override.
No entrance choreography — an instrument is always already on.

## 7. Accessibility floor

Skip link; `:focus-visible` cyan outline offset 2; landmarks (`header/aside/
main/footer`); all charts paired with DOM captions/values; keyboard panel nav
(`1–4`) guarded against input fields; touch targets ≥32px on mobile strip.

## 8. Do / Don't

**Do** keep provenance visible (`ft-prov`, captions); label states with words;
use mono for anything measured; spend emphasis sparingly (one accent moment
per viewport).

**Don't** introduce a new hue; round corners; add shadows/gradients/glass;
put the tricolor on actions or data; animate charts; nest cards; fake data —
an empty panel says so (`disponible après gel`).
