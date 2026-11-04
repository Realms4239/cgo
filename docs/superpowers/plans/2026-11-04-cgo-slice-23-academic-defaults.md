# Slice 23 — Defaults académiques (220 ms + pill BBRv3)

*Plan écrit le 2026-11-04 par le builder. Spec : `docs/superpowers/specs/2026-11-04-cgo-slice-23-academic-defaults-design.md`. Exécution sur HEAD `8477d20`.*

## T0 — Spec + plan + index

```bash
git add docs/superpowers/specs/2026-11-04-cgo-slice-23-academic-defaults-design.md docs/superpowers/plans/2026-11-04-cgo-slice-23-academic-defaults.md docs/superpowers/plans/README.md
git commit -m 'docs: le spec slice 23 defaults académiques'
```

L'index passe la slice 22 en « builder pass complete, packet ready » et ouvre la slice 23.

## T1 — Defaults 220 + pill BBRv3 + tests

**Files:**
- Modify: `web/frontend/src/lib/settings.ts`, `web/frontend/src/lib/schema.ts`, `web/frontend/src/components/BenchConfig.tsx` (une ligne), `web/frontend/src/components/BenchCapabilities.tsx` (une ligne)
- Modify tests: `web/frontend/src/components/BenchCapabilities.dom.test.tsx`
- Create tests: `web/frontend/src/lib/settings.test.ts`, `web/frontend/src/lib/schema.test.ts`

**Interfaces:**
- Consumes: `DEFAULTS`, `LOCAL_FALLBACK`, `loadSettings()`, `BenchCapabilities` props.
- Produces: défaut produit `deadlineMs = 220`, `deadline_ms` param `default = 220`, étiquette `BBRv3 non détecté`.

- [ ] **Step 1 — Défauts 220**

  ```ts
  // settings.ts
  warnMs: 40, critMs: 100, deadlineMs: 220, ringMax: 1800, watchMs: 500,

  // schema.ts
  { key: 'deadline_ms', label: 'Deadline small p95', min: 200, max: 5000, default: 220, ... },

  // BenchConfig.tsx:265
  onChange={e => setDeadlineMs(parseInt(e.target.value) || 220)}
  ```

- [ ] **Step 2 — Pill BBRv3**

  ```tsx
  BBRv3 non détecté
  ```

  Le `title` garde `noyau ${kernel}` ; le chemin `kit/install-bbr3.sh` reste dans le code et le runbook serveur.

- [ ] **Step 3 — Tests**

  - `settings.test.ts` : `DEFAULTS.deadlineMs === 220` ; `loadSettings()` vide → 220 ; persistance `localStorage` garde un choix explicite (1000 reste 1000).
  - `schema.test.ts` : le param `deadline_ms` de `LOCAL_FALLBACK` vaut 220.
  - `BenchCapabilities.dom.test.tsx` : contient `BBRv3 non détecté`, ne contient pas `install-bbr3.sh`.

  ```bash
  cd web/frontend && npm test src/lib/settings.test.ts src/lib/schema.test.ts src/components/BenchCapabilities.dom.test.tsx
  ```

- [ ] **Step 4 — Commit**

  ```bash
  git add web/frontend/src/lib/settings.ts web/frontend/src/lib/schema.ts web/frontend/src/components/BenchConfig.tsx web/frontend/src/components/BenchCapabilities.tsx web/frontend/src/components/BenchCapabilities.dom.test.tsx web/frontend/src/lib/settings.test.ts web/frontend/src/lib/schema.test.ts
  git commit -m 'fix(web): aligne les défauts sur 220 ms et retire le chemin du label (slice 23, T1)'
  ```

## T2 — Paquet de vérification (builder pass)

- `docs/verification/2026-11-04-slice-23.md` : commandes, hôte, date, commit, proposition 4 affirmations, bloc signature vide, non-prouvés explicites (pas de screenshot inventé, pas de live).
- Une ligne `PENDING` dans `docs/VERIFY.md`; checkpoint dans `docs/CHECKPOINTS.md`; paragraphe annexe dans le working tree (`book/` jamais committé).

```bash
go test ./... -count=1
python tools/verify.py
python book/tools/audit-archive.py
cd web/frontend && npm run build && npm test
git add docs/verification/2026-11-04-slice-23.md docs/VERIFY.md docs/CHECKPOINTS.md
git commit -m 'docs: le paquet de vérification tranche 23, sans auto-certification (slice 23, T2)'
```

## Acceptation

- `settings.test.ts` + `schema.test.ts` verts (défaut produit = 220).
- `BenchCapabilities.dom.test.tsx` vert sans `kit/install-bbr3.sh` dans le DOM.
- Aucun header, corpus, schéma CSV, live, CHR, bench, merge.
- Les réglages persistés des utilisateurs sont respectés.
