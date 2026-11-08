# Module frontend citation + proof panel as standalone tool — primary-source research

*Date: 2026-11-07 · Scope: `web/frontend/src/lib/citation.ts`, `web/frontend/src/components/AuditReportCard.tsx`, `web/frontend/src/components/ArchiveTable.tsx`, `web/frontend/src/components/ProofDrawer.tsx`, plus direct deps `web/frontend/src/lib/types.ts`, `lib/auditFlow.ts`, `lib/archive.ts`, `lib/dashboard.ts`, `lib/format.ts`, `lib/modalFocus.ts`, `store/ui.ts`, `store/audit.ts`, `components/RemediationCard.tsx`, `components/AuditReportsPanel.tsx`, `web/frontend/package.json`, `web/frontend/vite.config.ts`, `web/frontend/embed.go`, `pkg/results/seed.go` (go:embed precedent). Research only — no code, no commit.*

Primary sources fetched this pass: (R1) W3C ARIA in HTML, W3C Recommendation 11 Aug 2026 (<https://www.w3.org/TR/html-aria/>); (R2) W3C APG Dialog (Modal) Pattern (<https://www.w3.org/WAI/ARIA/apg/patterns/dialog-modal/>); (R3) W3C APG Tabs Pattern (<https://www.w3.org/WAI/ARIA/apg/patterns/tabs/>); (R4) W3C HTML 4.01 Tables §11 (<https://www.w3.org/TR/html40/struct/tables.html>); (R5) OWASP Cross Site Scripting (XSS) (<https://owasp.org/www-community/attacks/xss/>); (R6) OWASP XSS Prevention Cheat Sheet (<https://cheatsheetseries.owasp.org/cheatsheets/Cross_Site_Scripting_Prevention_Cheat_Sheet.html>); (R7) OWASP Top 10 (<https://owasp.org/www-project-top-ten/>); (R8) Go `embed` package (<https://pkg.go.dev/embed>); (R9) Semantic Versioning 2.0.0 (<https://semver.org/>). Local format reference: sibling research `docs/research/2026-11-07-module-cli-frontend.md` (R1–R7 there = Go modules / semver / OCI / Cobra / urfave / Vite / React) and `docs/research/2026-11-07-module-evidence-attestor.md`.

Method: read four target files in full (27 + 345 + 130 + 277 lines), `citation.test.ts:1-39`, `AuditReportCard.dom.test.tsx:1-139`, `ArchiveTable.dom.test.tsx:1-78`, `ProofDrawer.dom.test.tsx:1-134`, `package.json:1-34`, `vite.config.ts:1-8`, `embed.go:1-21`. Every external claim cites `(R#)`; every repo claim cites `file:line`; anything else is `unverified — primary source not found`.

## 1. Responsibility

The four files form one inseparable "cite what was sealed, show what proves it, list what is archived" surface, today embedded in the CGO console, not a standalone tool:

1. **Citation string (pure, no DOM)** — `auditCitation(report)` in `web/frontend/src/lib/citation.ts:6-27` builds a one-line French citation from *only* sealed-report fields: `audit_id` (trimmed, required), `provenance.source_sha256` (trimmed, required), `site`, `link_type`/`provider`, `timestamp`, `measured.direction` via `directionLabel`, and load-measured flags. Absent fields are omitted, never zero-filled (`citation.ts:4-5`). If id or seal is missing it returns `''`, never a partial citation (`citation.ts:7-10`). Output shape: `Audit CGO <id> — <site, link/provider, timestamp>, direction <label> · charge [non] mesurée. Rapport scellé : source_sha256 <seal>.` (`citation.ts:18-26`).
2. **Full report card (single sealed report, read step)** — `AuditReportCard` in `web/frontend/src/components/AuditReportCard.tsx:1-345` renders the complete report without hiding absence: Fenêtre / Constat / Charge / SLA CAP / Économie / Preuves / Remédiation / Non mesuré sections (`AuditReportCard.tsx:132-342`). Its proof badge derives from `proofState` (missing = no id or no `source_sha256`; observation = no load; degraded = SLA breach or grade D/F; else valid) (`AuditReportCard.tsx:29-38`). The "CITER CE RAPPORT" button calls `auditCitation(report)` via the `onCopy` prop and is `disabled` when proof is `missing` (`AuditReportCard.tsx:276-277`).
3. **Archive list (many rows, windowed, no per-row fetch)** — `ArchiveTable` in `web/frontend/src/components/ArchiveTable.tsx:1-130` lists archived audit rows with only the visible window in the DOM; the full dossier is fetched only on operator action ("Ouvrir") (`ArchiveTable.tsx:1-3`, `ArchiveTable.tsx:127`). Sealed rows expose three actions (Ouvrir / RAPPORT MD / CITER); unsealed rows expose nothing but "Preuve manquante" (`ArchiveTable.tsx:104-118`). Status words come from `archiveStatus`/`archiveVerdictWord` in `web/frontend/src/lib/archive.ts:37-56`, never recomputed from raw metrics.
4. **Proof drawer (two readings of one sealed report)** — `ProofSheet`/`ProofDrawer` in `web/frontend/src/components/ProofDrawer.tsx:1-277` shows "what we deploy" (exact command blocks, sizing basis, RouterOS replay verdict) and "what proves it" (figures, sensor, hashes), with the explicit rule "Aucun chiffre n'est recalculé ici" (`ProofDrawer.tsx:1-3`). It is a modal dialog with two tabs (`commands` / `proof`), arrow-key tab navigation, portal to `document.body`, and focus trapped by `useModalFocus` (`ProofDrawer.tsx:209-277`, `web/frontend/src/lib/modalFocus.ts:15-65`).

What the module deliberately does not do: re-measure, recompute verdicts, invent missing values, or fetch except via injected callbacks (`AuditReportCard.tsx:104-107` `onCopy`, `ArchiveTable.tsx:17-25` `onOpen`/`onCopyCitation`, `ProofSheet` props `ProofDrawer.tsx:202-207`). Fetching lives in the container `AuditReportsPanel` (`web/frontend/src/components/AuditReportsPanel.tsx:89-131`: `GET /api/audit/list`, `GET /api/audit/report?audit_id=`).

## 2. Exported API (file:line)

### 2.1 `web/frontend/src/lib/citation.ts` (27 lines)

| Export | Kind | Location | Contract |
|---|---|---|---|
| `auditCitation(report: AuditReport): string` | function | `citation.ts:6-27` | Pure string builder. `''` iff `!audit_id.trim()` or `!provenance.source_sha256.trim()` (`citation.ts:7-10`). Depends only on `directionLabel` (`citation.ts:1`, defined `web/frontend/src/lib/auditFlow.ts:54-65`) and `AuditReport` type (`web/frontend/src/lib/types.ts:236-256`). |

No other export. No default export. No test file other than `citation.test.ts:11-39` (4 cases: full fields, omitted fields, missing seal → `''`, empty id → `''`).

### 2.2 `web/frontend/src/components/AuditReportCard.tsx` (345 lines)

| Export | Kind | Location | Notes |
|---|---|---|---|
| `capTone(cap?: AuditCapsla): BadgeTone` | function | `AuditReportCard.tsx:22-25` | `!measured` → `steel`; `breach` → `danger`; else `success`. Never green by default (`AuditReportCard.tsx:21`). |
| `gradeTone(grade?: string): BadgeTone` | function | `AuditReportCard.tsx:51-57` | A→`success`, B→`live`, C→`warning`, else `danger`; re-used by `AuditView` (`web/frontend/src/views/AuditView.tsx:6`). |
| `primaryARPerHour(report)` | function | `AuditReportCard.tsx:59-61` | `economics.primary.ar_per_hour ?? null`. |
| `b2bScenario(econ?)` | function | `AuditReportCard.tsx:63-66` | Finds scenario `model==='B' && name==='b2b-m2m-apn'`, falls back to first `model==='B'`. |
| `b2bARPerHour(report)` | function | `AuditReportCard.tsx:68-70` | `b2bScenario(...).ar_per_hour ?? null`. |
| `AuditReportCard({ report, onCopy })` | component | `AuditReportCard.tsx:104-345` | Props typed at `AuditReportCard.tsx:104-107`. Internal helpers (not exported): `proofState` (`AuditReportCard.tsx:29-38`), `proofTone` (`AuditReportCard.tsx:47-49`), `Section` (`AuditReportCard.tsx:87-94`), `Kpi` (`AuditReportCard.tsx:96-98`), `Muted` (`AuditReportCard.tsx:100-102`), `ms`/`bytes`/`ratioBasis` (`AuditReportCard.tsx:72-85`). |

### 2.3 `web/frontend/src/components/ArchiveTable.tsx` (130 lines)

| Export | Kind | Location | Notes |
|---|---|---|---|
| `ARCHIVE_ROW_PX = 44` | const | `ArchiveTable.tsx:14` | Row height for windowing. |
| `ARCHIVE_WINDOW_ROWS = 8` | const | `ArchiveTable.tsx:15` | Visible rows; viewport = `44 × 8 = 352px` (`ArchiveTable.tsx:32`). |
| `ArchiveTable({ rows, onOpen, onCopyCitation })` | component | `ArchiveTable.tsx:17-130` | Props at `ArchiveTable.tsx:17-25`. Uses `computeWindow` with `overscan: 3` (`ArchiveTable.tsx:33-39`, impl `web/frontend/src/lib/archive.ts:16-33`), `archiveStatus`/`archiveVerdictWord`/`formatArchiveDate` (`ArchiveTable.tsx:5-11`, impl `archive.ts:37-66`), `EmptyState` (`ArchiveTable.tsx:12`, `ArchiveTable.tsx:28-30`). |

### 2.4 `web/frontend/src/components/ProofDrawer.tsx` (277 lines)

| Export | Kind | Location | Notes |
|---|---|---|---|
| `replayState(v?: RouterOSValidation)` | function | `ProofDrawer.tsx:43-48` | `neutral` (no proof) / `amber` (zero checks) / `red` (any fail) / `green` (all pass). Zero checks is never validated — same rule as `AuditReportCard.tsx:315-317`. |
| `replayLine(v?)` | function | `ProofDrawer.tsx:50-52` | String wrapper over `replayState`. |
| `ProofSheetProps` | interface | `ProofDrawer.tsx:202-207` | `tab?`, `report?`, `onClose?`, `onTab?` — all optional; store-backed defaults at `ProofDrawer.tsx:209-213`. |
| `ProofSheet(props?)` | component | `ProofDrawer.tsx:209-264` | Renders dialog + tablist + panel without portal; test target (`ProofDrawer.dom.test.tsx:54`). |
| `ProofDrawer` (default) | component | `ProofDrawer.tsx:266-277` | Store-gated portal: `useUIStore(s => s.proof)` + `createPortal(..., document.body)` (`ProofDrawer.tsx:267-275`). Mounted once in `App.tsx:172` (grep `ProofDrawer` `App.tsx:14,172`). |

Internal (not exported): `sizingBasis` (`ProofDrawer.tsx:23-38`), `sensorLine` (`ProofDrawer.tsx:54-59`), `copyText` (`ProofDrawer.tsx:61-66`), `CommandsTab` (`ProofDrawer.tsx:68-131`), `ProofTab` (`ProofDrawer.tsx:133-198`), `IS_ABSENT_VALUE` (`ProofDrawer.tsx:19`), `onTabKey` (`ProofDrawer.tsx:219-229`).

## 3. Dependencies

### 3.1 Runtime (from `web/frontend/package.json:15-21`)

`react ^19.0.0`, `react-dom ^19.0.0`, `zustand ^5.0.0`, `animejs ^4.5.0`, `echarts ^5.5.0`. Of these, the four target files use only `react` (`AuditReportCard` implicit JSX, `ArchiveTable.tsx:4` `useState`, `ProofDrawer.tsx:4-5` `useRef/useState/createPortal`) and `zustand` transitively via stores (`ProofDrawer.tsx:12-13` `useAuditStore`/`useUIStore`). `animejs`/`echarts` are not imported by any of the four files (verified by reading `citation.ts:1-27`, `AuditReportCard.tsx:1-17`, `ArchiveTable.tsx:1-12`, `ProofDrawer.tsx:1-14` import blocks). Per R9, the package versions `1.3.5` (`package.json:4`) parse as MAJOR 1 / MINOR 3 / PATCH 5 (R9 §"normal version MUST take the form X.Y.Z").

### 3.2 Dev / build (from `package.json:22-33`, `vite.config.ts:1-8`)

`@playwright/test ^1.62.1`, `@types/react`, `@types/react-dom`, `@vitejs/plugin-react ^4.0.0`, `jsdom ^30.0.1`, `typescript ^5.7.0`, `vite ^6.0.0`, `vitest ^3.2.7`. Scripts: `dev: vite`, `build: tsc --noEmit && vite build && node scripts/check-bundle.mjs`, `test: vitest run` (`package.json:6-14`). Vite config is 8 lines: `@vitejs/plugin-react`, dev proxy `'/api' → https://localhost:9090, secure:false` (self-signed backend), `build.outDir='dist', sourcemap:false` (`vite.config.ts:4-8`). The package is `"private": true` (`package.json:3`), so it cannot be published as-is. The built `dist/` is embedded in the Go binary via `//go:embed all:dist` → `embed.FS` → `fs.Sub(embedded, "dist")` (`web/frontend/embed.go:11-21`). Per R8, `//go:embed` initialises a variable of type `string`/`[]byte`/`FS` from files in the package directory at compile time, patterns are `path.Match` relative to the package directory, and "each pattern in a //go:embed line must match at least one file" (R8 §Directives). Precedent in the same repo: `//go:embed runs-frozen.tar.gz` + `var runsFrozen []byte` (`pkg/results/seed.go:7,15-16`) and `//go:embed bpf_probe.o` (`pkg/probe/bpf/assets_embedded.go:5-7`, grep result).

### 3.3 Internal imports (the real coupling)

| File | Imports (file:line) | Why it matters for extraction |
|---|---|---|
| `citation.ts` | `directionLabel` from `./auditFlow` (`citation.ts:1`), `AuditReport` type from `./types` (`citation.ts:2`) | Tiny. `directionLabel` is a 11-line switch (`auditFlow.ts:54-65`); `AuditReport` is a 20-field interface (`types.ts:236-256`). Both extractable. |
| `AuditReportCard.tsx` | `Badge`/`CardHead`/`Stat`/`Provenance`/`KernelRing` (`AuditReportCard.tsx:4-8`), `directionLabel` (`:9`), `auditCitation` (`:10`), `formatBytesFR/formatCostAR/groupDigits` (`:11`), `AuditReport` etc. (`:12-17`) | Heaviest coupling: 5 UI primitives + 3 format helpers. `Provenance` is 30 lines (`web/frontend/src/components/ui/Provenance.tsx:1-30`); `format.ts` builders at `format.ts:19-47`. A standalone card must either bundle or peer these. |
| `ArchiveTable.tsx` | `archiveStatus/archiveVerdictWord/computeWindow/formatArchiveDate/ArchiveRow` (`ArchiveTable.tsx:5-11`), `EmptyState` (`ArchiveTable.tsx:12`) | Light. `archive.ts` is 66 lines, pure (`archive.ts:1-66`). Only `EmptyState` is UI. |
| `ProofDrawer.tsx` | `operatorFacts` from `../lib/dashboard` (`ProofDrawer.tsx:6`, impl `dashboard.ts:83-124`), `directionLabel` (`:7`), `numFr` (`:8`, impl `format.ts:32-33`), `loadSettings` (`:9`), `useModalFocus` (`:10`, impl `modalFocus.ts:15-65`), `AuditReport` etc. (`:11`), `useAuditStore` (`:12`), `useUIStore/ProofTab` (`:13`, `store/ui.ts:24,35,55`), `PLATFORM_LABEL/blocksOf/carriedPlatforms/readPlatform/savePlatform` from `./RemediationCard` (`:14`, impl `RemediationCard.tsx:12-56`) | Heaviest store coupling. `ProofTab` also hard-reads `loadSettings()` on every render (`ProofDrawer.tsx:134`). `CommandsTab` reads `localStorage` via `readPlatform()` (`ProofDrawer.tsx:71`, `RemediationCard.tsx:46-51`). |

Store surface: `useUIStore` holds `proof: { tab: ProofTab } | null`, `openProof`, `closeProof` (`store/ui.ts:24,35,55`); `useAuditStore` holds `report: AuditReport | null` (`store/audit.ts:53,141`). `ProofSheet` falls back to `getState()` when props are absent (`ProofDrawer.tsx:210-212`) — non-reactive, see §5.

## 4. Primary-source requirements

### 4.1 W3C — ARIA in HTML (R1)

R1 defines authoring rules for ARIA use on HTML elements, including: authors MUST NOT conflict with native semantics (R1 §1); implicit roles (e.g. `button` → `role=button`, `table` → implicit `table`, `th scope` handling) MUST NOT be redundantly overwritten (R1 §4 table); `aria-checked` is not for elements supporting `checked` (R1 §3.4 Example 10). Requirements for this module:

- `ArchiveTable` uses native `<table>/<caption>/<colgroup>/<th scope="col">/<tbody>` (`ArchiveTable.tsx:61-72`) — conforms to R1's "do not override native semantics" rule; no `role` attributes on the table itself (R1 §4: `<table>` implicit `table`). `aria-rowcount`/`aria-rowindex` are global ARIA attributes — `unverified — primary source not found` in R1 for their specific allowance on `tr` (not in the fetched R1 excerpt; needs the ARIA spec proper, not fetched).
- `ProofDrawer`'s radiogroup buttons (`ProofDrawer.tsx:88-101`, `role="radio"` + `aria-checked`) and `RemediationCard`'s identical pattern (`RemediationCard.tsx:125-138`): R1 §4 allows `radio` on `button` (R1 §4 row `button`, Roles include `radio`). Conformant in role choice; keyboard behaviour is governed by R3, not R1.
- `AuditReportCard`'s `<details>/<summary>` scenarios table (`AuditReportCard.tsx:220-241`): R1 §4 lists `details` implicit `group`, `summary` conditional allowances — the file uses no `role` on `summary`, which R1 §3.3 warns against adding (`role="button"` on `summary` is "unnecessary and can result in cross-platform issues", R1 §3.3 Example 7). Conformant by omission.

### 4.2 W3C APG — Dialog (Modal) (R2)

R2 requires: container `role="dialog"` + `aria-modal="true"`; accessible name via `aria-labelledby` (visible title) **or** `aria-label`; Tab/Shift+Tab trapped inside; Escape closes; focus moves inside on open and returns to invoker on close; a visible close button is "strongly recommended" (R2 §Keyboard Interaction, §WAI-ARIA Roles/States/Properties). `ProofSheet` satisfies: `role="dialog" aria-modal="true" aria-label="Preuves de l'audit"` (`ProofDrawer.tsx:231`), visible close button `✕` with `aria-label="Fermer"` (`ProofDrawer.tsx:238`), Escape + Tab trap + focus return in `useModalFocus` (`modalFocus.ts:36-63`), portal outside `#shell` so `inert` does not self-inert (`modalFocus.ts:10-14,29`). Gap: the visible `<h2>Preuves de l'audit</h2>` (`ProofDrawer.tsx:235`) is not referenced by `aria-labelledby` — R2 allows `aria-label` as the alternative, so this is conformant but duplicates the name in two places (AT will announce label, sighted users read h2).

### 4.3 W3C APG — Tabs (R3)

R3 requires: container `role="tablist"`, each tab `role="tab"` + `aria-controls`, active tab `aria-selected="true"`, panel `role="tabpanel"` + `aria-labelledby`, tablist labelled, Left/Right arrows cycle (wrapping), optional Home/End, automatic activation on focus "recommended" when panels preload (R3 §Keyboard Interaction, §WAI-ARIA Roles/States/Properties). `ProofSheet` uses `role="tablist" aria-label="Lectures du rapport"` (`ProofDrawer.tsx:240`), `role="tab" id="proof-tab-commands/proof" aria-selected aria-controls="proof-panel"` (`ProofDrawer.tsx:241-254`), `role="tabpanel" id="proof-panel" aria-labelledby` (`ProofDrawer.tsx:256`). Arrow handling cycles both ways and re-focuses the newly activated tab (`ProofDrawer.tsx:219-229`). Gaps vs R3: (a) no Home/End (optional per R3, so not a violation); (b) both tabs are `type="button"` in the tab order — R3's Tab key description implies only the active tab is in the tab sequence ("When focus moves into the tab list, places focus on the active tab element"), so a roving `tabindex` is missing; (c) a single shared `id="proof-panel"` for both tabs is allowed by R3 (`aria-controls` refers to "its associated tabpanel element") but means AT cannot distinguish two panels — acceptable for a 2-tab swap, questionable for a library.

### 4.4 W3C HTML Tables (R4)

R4 (HTML 4.01 §11) defines `<caption>` (one, immediately after `<table>`), `<colgroup>/<col>` for column widths enabling incremental rendering, `<th scope>` / `headers` / `abbr` for associating header info with data cells, and the rule "Tables should not be used purely as a means to layout document content" (R4 §11.1). `ArchiveTable` uses `<caption>` (`ArchiveTable.tsx:62`), `<colgroup>` with 7 `<col style width>` (`ArchiveTable.tsx:63-71`), `<th scope="col">` on all 7 headers (`ArchiveTable.tsx:74-80`). Conformant. Gap: the table is split by a spacer `<div>` with `paddingTop/paddingBottom` (`ArchiveTable.tsx:57-60`) — R4's table model assumes rows are direct `TR` descendants; the padding-spacer windowing is a rendering hack with no R4 basis (`unverified — primary source not found` for virtualised tables in R4). Screen-reader row count relies on `aria-rowcount`/`aria-rowindex` (§4.1 unverified).

### 4.5 OWASP — XSS and output encoding (R5, R6, R7)

R5 states XSS occurs when "an attacker uses a web application to send malicious code… to a different end user" via unvalidated input included in output, and "The primary defenses against XSS are described in the OWASP XSS Prevention Cheat Sheet" (R5 §How to Protect Yourself). R6's cheat sheet TOC (fetched excerpt) mandates context-specific output encoding: HTML contexts, HTML attribute contexts, JavaScript contexts, CSS contexts, **URL contexts** (R6 §Output Encoding subsections). R7 lists the Top 10 as "the most critical web application security risks" and "the first step towards more secure coding" (R7).

Applied to this module:

- All four files render user/sealed data only via JSX interpolation (`{report.site}`, `{row.site}`, `{p.source_sha256}`, etc. — e.g. `AuditReportCard.tsx:124-130`, `ArchiveTable.tsx:97-102`, `ProofDrawer.tsx:158-192`). No `dangerouslySetInnerHTML` appears in the four files (verified by reading all four in full; `grep` for `dangerouslySetInnerHTML` across `web/frontend/src` was not run — `unverified — primary source not found` for the rest of the tree). React's default escaping is the defence; R6 would additionally require URL-context encoding for `href` attributes.
- `href`/`download` construction uses `encodeURIComponent(id)` everywhere: `AuditReportCard.tsx:117-119` (`/api/audit/report?audit_id=…&format=md`), `ArchiveTable.tsx:109`, `ProofDrawer.tsx:193`, `AuditReportsPanel.tsx:102,112`. This matches R6's URL-context encoding requirement. The `download={\`${id}.md\`}` attribute (`ArchiveTable.tsx:110`) is inside the `canOpen` guard (`sealed && !!id`, `ArchiveTable.tsx:88`) — an unsealed row cannot inject it.
- `title` attributes carry raw values (`ArchiveTable.tsx:97-102` `title={row.timestamp?.trim()}`, `title={site}`, etc.; `AuditReportCard.tsx:257,263` `title={report.provenance.source_sha256}`). R6's HTML-attribute-context encoding applies; React escapes attribute values, but a standalone extractor must not switch to `innerHTML` templating.
- Clipboard writes (`ProofDrawer.tsx:61-66`, `AuditReportsPanel.tsx:124,134`) use `navigator.clipboard?.writeText` with a promise rejection toast — no XSS vector, but no fallback copy mechanism (R5/R6 do not cover clipboard; `unverified — primary source not found` for Clipboard API security requirements).

### 4.6 Go `go:embed` (R8) — for the Go half of the target tool

R8: "Go source files that import "embed" can use the //go:embed directive to initialize a variable of type string, []byte, or FS with the contents of files read from the package directory or subdirectories at compile time" (R8 §Overview); "Patterns must not match files outside the package's module… each pattern must match at least one file" (R8 §Directives); `FS` implements `io/fs.FS` for `http.FileServer(http.FS(content))` (R8 §File Systems). The current frontend embed (`web/frontend/embed.go:11-21`) and the frozen-corpus precedent (`pkg/results/seed.go:15-16`) both follow this. A standalone Go module serving the proof panel must follow R8's pattern restrictions (no `../`, no outside-module matches).

### 4.7 Semantic Versioning (R9) — for the npm/Go module half

R9: "Software using Semantic Versioning MUST declare a public API" (R9 §SemVer item 1); "A normal version number MUST take the form X.Y.Z" (R9 item 2); "Once a versioned package has been released, the contents of that version MUST NOT be modified" (R9 item 3); MINOR = backward-compatible feature, PATCH = backward-compatible bug fix, MAJOR = incompatible API change (R9 Summary + items 7–9). The current monolith `package.json` is `1.3.5` + `private:true` (`package.json:3-4`) — no declared public API, no publishable module. Any `cgo-cite` extraction must declare its API (the tables in §2) before claiming 1.0.0 (R9: "Version 1.0.0 defines the public API").

## 5. Gap analysis

Ordered by extraction risk. Each gap cites the code; severity is judgement, not a primary-source claim.

### G1 — Citation is a display string, not a verifiable reference (high)

`auditCitation` emits id + site/link/timestamp/direction + full 64-char seal in prose French (`citation.ts:24-26`), but: (a) no URL to the sealed report (`/api/audit/report?audit_id=` exists in the card at `AuditReportCard.tsx:117-119` but is not included in the citation string); (b) no machine-readable form (JSON) — `AuditReportsPanel.copyCitation` fetches the full report only to re-derive the string (`AuditReportsPanel.tsx:108-131`); (c) seal format is not validated — any non-empty `source_sha256` passes (`citation.ts:9-10`), tests use `'a'.repeat(64)` (`citation.test.ts:5`) but no hex/length check exists; (d) the card truncates the same seal to 12 chars for display (`AuditReportCard.tsx:253,259,265`) while the citation carries all 64 — a copier comparing the two sees different strings with no stated relationship. Extraction must decide: keep prose-only (documented as non-verifiable) or add a canonical JSON/`cite:` form. Today the string is citable-looking but not verifiable — the textbook toy risk.

### G2 — Two status vocabularies for the same seal (high)

`proofState` (`AuditReportCard.tsx:29-38`) speaks `valid/degraded/observation/missing`; `archiveStatus` (`archive.ts:37-50`) speaks `Valide/Dégradé/Observation seule/—` but with different triggers: archive `red` and `amber` both → `Dégradé` (`archive.ts:42-44`), unsealed → `—` (`archive.ts:38`), while the card maps breach-or-D/F → `degraded` and no-load → `observation` (`AuditReportCard.tsx:33-36`). `gradeTone` (`AuditReportCard.tsx:51-57`) adds a third scale (A/B/C/D). A standalone library with three unsynchronised status functions will confuse consumers; extraction must unify behind one `ProofState` type with a documented mapping table (card ↔ archive ↔ grade).

### G3 — `ProofSheet` store fallback is non-reactive (medium)

`ProofSheet` reads `useUIStore.getState()` / `useAuditStore.getState()` once per render when props are absent (`ProofDrawer.tsx:210-212`), while `ProofDrawer` (default) subscribes properly (`ProofDrawer.tsx:267-270`). A consumer mounting `ProofSheet` without props gets a snapshot, not live updates — and the DOM test relies on this via `setState` before render (`ProofDrawer.dom.test.tsx:51-55`). Extraction must make props required or subscribe, not `getState()`.

### G4 — Tabs: roving tabindex and Home/End missing (medium, R3)

See §4.3(b). Both tabs are tabbable; R3's keyboard model expects one tabbable tab + arrow navigation. Also `CommandsTab`'s radiogroup (`ProofDrawer.tsx:88-101`) has no arrow-key handling at all — R3 does not cover radiogroups (that is the APG Radio Group pattern, not fetched — `unverified — primary source not found` for its keyboard rules), but by analogy the same arrow expectation applies.

### G5 — Archive virtualisation breaks native table row semantics (medium, R4)

`computeWindow` (`archive.ts:16-33`) + spacer padding (`ArchiveTable.tsx:57-60`) means the DOM contains `start..end` rows only, with `aria-rowcount`/`aria-rowindex` patching the gap. R4 assumes all `TR` rows are present; the DOM test asserts `≤30` of 179 rows in initial HTML (`ArchiveTable.dom.test.tsx:25-30`). AT support for `aria-rowindex` on real `<tr>` elements inside a padded spacer is `unverified — primary source not found` (needs ARIA-in-HTML rowindex allowance, not validated in R1 fetch). A standalone table must either keep virtualisation and document the AT caveat, or paginate.

### G6 — Hardcoded locale, thresholds, and factors (medium)

French strings throughout (`citation.ts:21-25`, `AuditReportCard.tsx:130,150,164,188,195,242`, `ProofDrawer.tsx:37,44-47,54-59`); `sizingBasis` hardcodes `× 0,95` (`ProofDrawer.tsx:30-32`, test expects `5,1 Mbit/s × 0,95 → 4,8 Mbit/s` at `ProofDrawer.dom.test.tsx:78`); `operatorFacts` thresholds come from settings (`ProofDrawer.tsx:134-135`, `dashboard.ts:83-124`); `formatArchiveDate` uses local timezone despite the comment "l'archive parle locale, pas UTC ambiguë" (`archive.ts:58-66` uses `getFullYear/getMonth/...` local getters). None of these is parameterised. Extraction needs a `locale`/`labels` injection point or an explicit "French-only, 0.95, local-time" contract.

### G7 — Clipboard has no fallback and no secure-context check (low-medium)

All copy paths (`ProofDrawer.tsx:61-66`, `AuditReportsPanel.tsx:124,134-137`) call `navigator.clipboard?.writeText` and toast "Copie impossible — sélectionnez le texte" on rejection. No `document.execCommand` fallback, no permission check. The citation copy in the archive flow additionally re-fetches the report to build the string (`AuditReportsPanel.tsx:112-115`) — wasteful if the row already had the seal (rows do not carry `source_sha256` today; `ArchiveRow = Record<string,string>` at `archive.ts:4`).

### G8 — Versioning / packaging: private monolith, hardcoded backend (low for extraction, blocking for publish)

`package.json:3` `private:true`; Vite proxy hardcodes `https://localhost:9090` (`vite.config.ts:6`); `dist` is embedded by Go (`embed.go:11`), so the frontend cannot run standalone today. Per R9, no public API is declared for `1.3.5`. Per R8, a Go-side standalone panel must be its own package directory with its own `//go:embed` patterns (cannot embed `../dist` from another module — R8: patterns "must not match files outside the package's module").

### G9 — What is already good (keep verbatim)

- Fail-closed citation: `''` on missing id/seal, button disabled on `missing` (`citation.ts:7-10`, `AuditReportCard.tsx:276-277`, test `AuditReportCard.dom.test.tsx:63-69`), panel refuses to cite ("citation impossible — empreinte source absente", `AuditReportsPanel.tsx:116-123`).
- `encodeURIComponent` on every `audit_id` in URLs (§4.5).
- Zero-checks ≠ validation, twice (`AuditReportCard.tsx:315-317`, `ProofDrawer.tsx:45`) with tests (`AuditReportCard.dom.test.tsx:131-138`, `ProofDrawer.dom.test.tsx:111-119`).
- Unmeasured ≠ zero: `—` / "Pas encore mesuré" / "Non chiffré" discipline (`AuditReportCard.tsx:72-85`, `archive.ts:35-36,58-66`, `ProofDrawer.tsx:16-19`, `format.ts:28-33`).
- SEAL truncation is display-only; citation carries full hash (`AuditReportCard.tsx:253` vs `citation.ts:25`).

## 6. Target tool: `cgo-cite` library + proof panel as independent npm/Go module

Two artefacts, one repo, versioned together. SemVer per R9 (declare API first, §2 tables become the declared API; 0.x until G1–G5 close).

### 6.1 `cgo-cite` — zero-dependency TS/Go core (pure functions only)

- **TS (`cgo-cite`, npm):** `auditCitation(report, options?) → string` (rename internal, keep `auditCitation` alias), `proofState(report) → ProofState`, `archiveStatus(row)`, `archiveVerdictWord(row)`, `computeWindow(input)`, `formatArchiveDate(value, timeZone?)`, `directionLabel(direction)`, plus the unified types `AuditReport`/`ArchiveRow`/`ProofState`. Zero runtime deps (React not needed). `directionLabel` and `archive.ts` are already pure (`auditFlow.ts:54-65`, `archive.ts:16-66`); `citation.ts` is pure given them.
- **Go (`github.com/<org>/cgo-cite`, module):** `func Citation(r Report) (string, error)` + `func ProofState(r Report) ProofState` over a Go mirror of `AuditReport` (`types.ts:236-256` subset: `audit_id, site, link_type, provider, timestamp, measured{direction, load_measured, load_measured_down}, provenance{source_sha256}`), returning an error (not `''`) on missing id/seal so Go callers cannot silently copy a partial citation. Tests: golden vectors shared with the TS side (same fixtures as `citation.test.ts:7-39`).
- **Citation v2 (closes G1):** `auditCitation` gains an optional `url` base (defaults to `''`, preserving today's string) and a `format: 'prose' | 'json'` option; JSON form carries `{ id, site, link_type, provider, timestamp, direction, load_measured, source_sha256, report_url }`. Seal validation: require `/^[0-9a-f]{64}$/i` after trim, else `''`/error with reason — documented as a breaking change (MAJOR per R9) if enabled by default.

### 6.2 `@cgo/proof-panel` — React components (peer `react`, `react-dom`)

- `AuditReportCard` (slimmed: Fenêtre/Constat/Charge/Preuves/Non mesuré; Économie + Remédiation stay in CGO — they depend on `economics.*` and `RemediationCard`), `ArchiveTable`, `ProofSheet` (props required, no `getState()` fallback — closes G3), plus `ProofDrawer` as thin store adapter kept in CGO, not in the package.
- Peer deps: `react ^19`, `react-dom ^19` (matching `package.json:18-19`); optional `zustand` only for the adapter. Bundled or peered UI primitives: `Badge/CardHead/Stat/Provenance/EmptyState` must be inlined (they are 1-file each; `Provenance.tsx:1-30` is the proof) or declared as `@cgo/ui` peers — decide in Slice 2, do not leave as relative imports.
- A11y contract (from R2/R3): `aria-labelledby` to the visible `<h2>` (prefer over duplicate `aria-label`, R2), roving `tabindex` on tabs + Home/End (R3), arrow-key radiogroup in `CommandsTab` (analogy to R3; APG Radio Group pattern to be fetched before claiming conformance — `unverified — primary source not found` until then). Archive table keeps `<caption>` + `<colgroup>` + `<th scope="col">` (R4) and documents the virtualisation AT caveat (G5) or switches to pagination behind a `virtualized?: boolean` prop (default `false` for library consumers).
- i18n contract (closes G6): `labels?: Partial<Record<CitationLabelKey, string>>` + `factor?: number` (default `0.95`) + `timeZone?: string` (default local, documented) parameters; French remains the default dictionary so CGO output is byte-identical.

### 6.3 Go static host (closes G8)

A `cmd/cgo-cite` (or exported `Handler()` in the Go module) serving the built panel via `//go:embed` per R8: its own package directory containing `dist/` (or `assets/`), `//go:embed all:dist`, `http.FileServer(http.FS(sub))` following the `embed.go:11-21` precedent and R8 §File Systems. R8 constraint: embed patterns cannot reach outside the module, so the Vite `outDir` for the panel build must be inside the Go package directory (today `web/frontend/dist` + `web/frontend/embed.go` satisfies this only because both live under `web/frontend/`). Backend URL becomes a runtime config (today hardcoded `vite.config.ts:6` proxy) — the panel fetches from a configurable `apiBase` prop, CGO passes `/api`.

## 7. Slices (smallest complete steps, in order)

| # | Slice | Exit (done when) |
|---|---|---|
| S1 | Extract pure core: move `citation.ts` + `archive.ts` + `directionLabel` + `AuditReport` subset into `packages/cgo-cite/src/` (TS) + `cite/citation.go` (Go); re-export from old paths so CGO is untouched. Shared golden tests from `citation.test.ts` + `archive.test.ts`. | `vitest run` + `go test ./cite/...` green; CGO `make verify-fast` unchanged; citation output byte-identical on all `citation.test.ts` cases. |
| S2 | Unify status: one `ProofState` (`valid/degraded/observation/missing`) + mapping table card↔archive↔grade; `archiveStatus` and `proofState` delegate to it. Seal-format validator (`/^[0-9a-f]{64}$/i`) behind `strictSeal?: boolean` (default `false`). | New `proofState.test.ts` covers all four states × both vocabularies; existing `AuditReportCard.dom.test.tsx:63-93` + `ArchiveTable.dom.test.tsx:38-64` still pass. |
| S3 | Citation v2: `format:'prose'|'json'` + optional `baseUrl`; Go `Citation()` returns `(string, error)`. Document prose-only as non-verifiable in README. | JSON golden test; prose output unchanged when options omitted; README states what the string does **not** prove. |
| S4 | Extract components: `ArchiveTable` first (lightest deps, §3.3), then `AuditReportCard` slim, then `ProofSheet` with required props. Inline/peer the 5 UI primitives. Close G3 (no `getState()`), G4 (roving tabindex + Home/End + radiogroup arrows), G6 (labels/factor/timeZone). | `ProofDrawer.dom.test.tsx` + `ArchiveTable.dom.test.tsx` + `AuditReportCard.dom.test.tsx` pass against the package; CGO imports from the package; APG checklist (R2/R3) signed in the package README. |
| S5 | Go static host + publish: `cmd/cgo-cite` (or `cite.Handler`) with `//go:embed all:dist` (R8), configurable `apiBase`, npm `cgo-cite` + `@cgo/proof-panel` at `0.1.0` with declared API (R9). | `go build ./...` embeds panel; `npm pack --dry-run` shows only `dist` + types; semver README states 0.x instability per R9. |

Out of scope for all slices: re-measuring, re-sealing, signing (`unverified — primary source not found` for any signing requirement — no signing key exists in the citation path today), changing `AuditReport` fields produced by Go, touching `book/` frozen corpus.

## 8. Anti-toy criteria (a reviewer can fail the tool on any one)

1. **No fabricated citation.** Given a report without `audit_id` or without `source_sha256`, every entry point (`auditCitation`, Go `Citation`, CITER button, archive CITER) returns `''`/error and shows "citation impossible" — never a partial string (`citation.ts:7-10`, `AuditReportsPanel.tsx:116-123`). Test: `citation.test.ts:31-38` cases must pass in both TS and Go.
2. **No fabricated measurement.** Absent values render as `—` / "Pas encore mesuré" / "Non chiffré", never `0` (`archive.ts:58-66`, `ProofDrawer.tsx:16-19`, `format.ts:28-33`). Test: `ArchiveTable.dom.test.tsx:66-69`, `ProofDrawer.dom.test.tsx:129-133`.
3. **No silent success on zero checks.** Empty `checks[]` renders "non validée — aucune vérification", never "validée" (`AuditReportCard.tsx:315-320`, `ProofDrawer.tsx:45`). Test: `AuditReportCard.dom.test.tsx:131-138`.
4. **Round-trip on a real sealed report.** The tool must cite a report produced by the actual Go audit engine, fetch it back via `/api/audit/report?audit_id=`, and show the same 64-char seal in citation, card (full), and drawer (`ProofDrawer.tsx:181-183`) — truncated display (12 chars, `AuditReportCard.tsx:259`) must be labelled as truncation. A fixture-only demo fails this criterion.
5. **Standalone proof.** `npm install cgo-cite && node -e "require('cgo-cite').auditCitation(fixture)"` and `go run ./cmd/cgo-cite` both work with no CGO checkout, no `localhost:9090`, no `book/` corpus. The Vite proxy hardcode (`vite.config.ts:6`) must be gone from the panel build.
6. **A11y evidence, not assertion.** Dialog + tabs keyboard walkthrough (Tab trap, Escape, arrows, Home/End, focus return) run with a screen reader or the APG checklist per R2/R3; axe/equivalent scan clean for the three components. "It uses `role=dialog`" alone fails.
7. **XSS evidence.** Sealed fields containing `<script>`, `" onmouseover=`, `javascript:` render escaped as text; `audit_id` with `&`, `?`, `/` is `encodeURIComponent`-encoded in every URL (`AuditReportCard.tsx:118`, `ArchiveTable.tsx:109`, `ProofDrawer.tsx:193`). Test with malicious fixtures; rely on R5/R6, not on "React escapes it" folklore.
8. **Version honesty.** Package README declares the public API (§2 tables) and states 0.x instability per R9 until S1–S4 land; `package.json` is no longer `private:true` for the published artefact; version bumps follow R9 MAJOR/MINOR/PATCH rules (breaking citation-format change = MAJOR).
9. **No toy metrics.** The panel shows only fields present on the sealed report (`AuditReportCard.tsx:1-3` rule, `ProofDrawer.tsx:1-3` rule); any number the library computes (e.g. `sizingBasis` `×0.95`) is labelled as derived, with its factor visible (`ProofDrawer.tsx:30-32`), never presented as measured.

## 9. Unverified — primary source not found

- `aria-rowcount` / `aria-rowindex` allowance on native `<tr>` inside a virtualised spacer table (R1 fetched text does not cover `rowindex` on `tr`; ARIA spec proper not fetched).
- APG Radio Group keyboard rules for `CommandsTab`'s radiogroup (R3 covers tabs, not radiogroups; APG Radio Group pattern not fetched).
- Clipboard API secure-context / permission requirements (R5/R6 do not cover; MDN Clipboard API not fetched).
- Whether `navigator.clipboard` is available in the CGO deployment context (self-signed HTTPS, `vite.config.ts:6`) — deployment fact, not a standard; needs runtime check.
- `dangerouslySetInnerHTML` absence outside the four files read (only the four files + containers were read in full; full-tree grep not run this pass).
- Any requirement to sign citations (no signing in current code; SLSA/in-toto are covered in sibling research `2026-11-07-module-evidence-attestor.md`, not re-fetched here).
- React 19 / Vite 6 / Vitest 3 behavioural contracts (covered by sibling research `2026-11-07-module-cli-frontend.md` R6/R7, not re-fetched; this file relies only on `package.json:15-33` version pins + `vite.config.ts:1-8`).
