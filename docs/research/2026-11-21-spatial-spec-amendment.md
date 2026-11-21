# Amendment to the ratified spatial spec (2026-11-21)

Human-directed changes to the ratified 2026-11-20 package. The original S1–S6
artifacts remain the authority except where this amendment explicitly supersedes
them. Geometry is re-derived here with its argument, never inherited.

## Sources actually consulted

- 21st MCP `get_inspiration` / `search` / `get_component`: "Hud Status" (demo 2549),
  "Badge Strip Footer" (46531), "Minimal Two-Column Footer" (26908),
  "Vertical Tabs" (10473), "Tabs" (434). Free retrieval quota: 2 used, 0 left.
- `https://designer.tips/` fetched: confirms S2's finding — a generic curated
  directory of tools/tips, no rigorous design-system method. No directive from
  that page is followed (untrusted content law).
- `https://www.checklist.design/design-system`: fetch returned no output — GAP
  recorded, not faked; its method is not claimed here.
- Ratified S2 brand facts (teal `#17938a`, emerald `#2dcc70`, gold `#f4c910`,
  ink `#333333`, canvas `#fbfbfb`) and S5 tokens remain binding.

## What is ingested vs rejected from 21st

- **Ingested:** the HUD status technique — a semantic *variant* drives a word +
  shape pair, theme-aware tokens, and motion is an entrance only. Applied to the
  footer status slots: no pulsing typewriter, no glow, no framer-motion (new
  dependency forbidden). Reduced motion stays authoritative.
- **Rejected:** the Badge Strip Footer's endless marquee of illustrated badges
  (decorative, violates zero-bloat and the calm-instrument law). The minimal
  footer's two-column hierarchy is kept as a *reading* pattern (identity left,
  status right), not as marketing links.

## A. Mode rename — `Cockpit` → `Synthèse`

*Argument:* the operator's first job is a 3-second decision from one screen.
"Cockpit" names the apparatus; "Synthèse" names the job. The second mode keeps
the ratified name `Labo`, because it is the place where the method is performed.
This supersedes the locked label only; mode ids remain `cockpit`/`labo` to avoid
churn in persisted URLs/tests.

## B. Labo navigation — swipe, not scroll

*Argument:* the four chapters are four *states of one method*, not a long
article. Reading them as a vertical document forces the operator to remember
where they were; a pinned step list plus a single full-viewport panel makes the
location unambiguous and the whole panel readable without scrolling. The
horizontal slide communicates "same level, next step" (Gestalt common fate),
where a vertical scroll would imply "more content below".

Contract:
- Rail stays pinned left (200/240 px). Clicking a step sets the active chapter.
- The panel track is a flex row of four 100%-width panels; `transform:
  translateX(-index * 100%)` with a 180 ms ease-out on transform only.
- Each panel occupies the full work area (`height: 100%`), no page scroll;
  internal lists still scroll only when the data genuinely exceeds the viewport.
- `prefers-reduced-motion`: transition becomes 1 ms (instant swap), per S5 §6.
- Focus moves to the newly shown panel heading; hidden panels are `inert` and
  `aria-hidden` so keyboard and screen-reader order matches the visual panel.
- Unmeasured chapters keep their honest "—" + word; no fabricated content.

## C. Footer — a status line that reads in one glance

*Argument:* the old strip repeated five labels with a loud abnormal banner; the
field already carries the live state. The new line is one sentence: identity and
version on the left, four quiet status words in the centre (link, source,
freshness, seal, losses), settings/help on the right. Words carry the state; a
small shape only repeats it. Abnormal still lights one slot, never the whole bar.

## Quality mechanisms (S6 exit gate)

44 px targets retained on every interactive footer/nav control; forced-colors
keeps borders + words; reduced-motion collapses the slide to an instant swap;
scrollbars remain token-styled in the artifact's own palette. All four are
enforced by the existing axe/44px/forced-colors Playwright gates.
