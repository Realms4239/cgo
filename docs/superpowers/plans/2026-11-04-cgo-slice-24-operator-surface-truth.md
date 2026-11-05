# Slice 24 — Vérité des surfaces (vocabulaire live + fixtures ingress)

*Plan écrit le 2026-11-04 par le builder. Spec : `docs/superpowers/specs/2026-11-04-cgo-slice-24-operator-surface-truth-design.md`. Exécution sur HEAD `7f3f70c`.*

## T0 — Spec + plan + index

```bash
git add docs/superpowers/specs/2026-11-04-cgo-slice-24-operator-surface-truth-design.md docs/superpowers/plans/2026-11-04-cgo-slice-24-operator-surface-truth.md docs/superpowers/plans/README.md
git commit -m 'docs: le spec slice 24 vérité des surfaces'
```

L'index passe la slice 23 en « builder pass complete, packet ready » et ouvre la slice 24.

## T1 — Vocabulaire live + fixtures ingress

**Files:**
- Modify: `web/frontend/src/views/LiveView.tsx` (2 étiquettes)
- Modify tests: `web/frontend/src/components/ProofDrawer.dom.test.tsx`, `web/frontend/src/components/RemediationCard.dom.test.tsx`

**Interfaces:**
- Consumes: `LiveView` labels, `remediate.RouterOSIngress()` output syntax.
- Produces: live edge labels `Maxi avant` / `sans réglage (file simple)`, ingress fixtures LAN queue tree.

- [ ] **Step 1 — Étiquettes du bord**

  ```tsx
  // LiveView.tsx:554
  Maxi avant{liveSmallMedian != null ? ` — ${liveSmallMedian.toFixed(1)} ms` : ''}
  // + title="mémoriser la médiane de référence, cliquer pour effacer"

  // LiveView.tsx:576
  'sans réglage (file simple)'
  ```

- [ ] **Step 2 — Fixtures ingress LAN queue tree**

  ```ts
  routeros_ingress: '/queue tree add name=cgo-ingress parent=bridge queue=cgoingress-cake',
  ```
  Les asserts de test portent la nouvelle syntaxe ; `/interface queue add` n'apparaît plus dans aucune fixture.

- [ ] **Step 3 — Tests**

  ```bash
  cd web/frontend && npm test src/components/ProofDrawer.dom.test.tsx src/components/RemediationCard.dom.test.tsx
  ```

  Verts, avec un assert explicite `not.toContain('/interface queue add')`.

- [ ] **Step 4 — Commit**

  ```bash
  git add web/frontend/src/views/LiveView.tsx web/frontend/src/components/ProofDrawer.dom.test.tsx web/frontend/src/components/RemediationCard.dom.test.tsx
  git commit -m 'fix(web): le vocabulaire bord en français courant et les fixtures ingress en queue tree (slice 24, T1)'
  ```

## T2 — Paquet de vérification (builder pass)

- `docs/verification/2026-11-04-slice-24.md` : commandes, hôte, date, commit, proposition 4 affirmations, bloc signature vide, non-prouvés explicites (pas de screenshot inventé, pas de live).
- Une ligne `PENDING` dans `docs/VERIFY.md`; checkpoint dans `docs/CHECKPOINTS.md`; paragraphe annexe dans le working tree (`book/` jamais committé).

```bash
go test ./... -count=1
python tools/verify.py
python book/tools/audit-archive.py
cd web/frontend && npm run build && npm test
git add docs/verification/2026-11-04-slice-24.md docs/VERIFY.md docs/CHECKPOINTS.md
git commit -m 'docs: le paquet de vérification tranche 24, sans auto-certification (slice 24, T2)'
```

## Acceptation

- Fixtures ingress LAN queue tree ; `ProofDrawer.dom.test.tsx` + `RemediationCard.dom.test.tsx` verts ; zéro `/interface queue add` dans les fixtures.
- Étiquettes LiveView en français courant.
- Aucun header, corpus, schéma CSV, live, CHR, bench, merge.
- Les réglages persistés des utilisateurs sont respectés.
