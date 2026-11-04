# Slice 25 — Honneur de l'infobulle Maxi avant (tooltip + barre)

*Plan écrit le 2026-11-04 par le builder. Spec : `docs/superpowers/specs/2026-11-04-cgo-slice-25-tooltip-honesty-design.md`. Exécution sur HEAD `f62a5a8`.*

## T0 — Spec + plan + index

```bash
git add docs/superpowers/specs/2026-11-04-cgo-slice-25-tooltip-honesty-design.md docs/superpowers/plans/2026-11-04-cgo-slice-25-tooltip-honesty.md docs/superpowers/plans/README.md
git commit -m 'docs: le spec slice 25 honneur infobulle'
```

L'index ouvre la slice 25.

## T1 — Tooltip Maxi avant

**Files:**
- Modify: `web/frontend/src/views/LiveView.tsx` (une ligne)

**Interfaces:**
- Consumes: `LiveView` tooltip string.
- Produces: tooltip `mémoriser la médiane de référence, cliquer pour effacer — la barre compare avant et maintenant, même fenêtre`.

- [ ] **Step 1 — Infobulle**

  ```tsx
  title="mémoriser la médiane de référence, cliquer pour effacer — la barre compare avant et maintenant, même fenêtre"
  ```

- [ ] **Step 2 — Commit**

  ```bash
  git add web/frontend/src/views/LiveView.tsx
  git commit -m 'fix(web): nomme la barre dans l’infobulle Maxi avant (slice 25, T1)'
  ```

## T2 — Paquet de vérification (builder pass)

- `docs/verification/2026-11-04-slice-25.md` : commandes, hôte, date, commit, proposition 3 affirmations, bloc signature vide, non-prouvés explicites (pas de screenshot inventé, pas de live).
- Une ligne `PENDING` dans `docs/VERIFY.md`; checkpoint dans `docs/CHECKPOINTS.md`; paragraphe annexe dans le working tree (`book/` jamais committé).

```bash
go test ./... -count=1
python tools/verify.py
python book/tools/audit-archive.py
cd web/frontend && npm run build && npm test
git add docs/verification/2026-11-04-slice-25.md docs/VERIFY.md docs/CHECKPOINTS.md
git commit -m 'docs: le paquet de vérification tranche 25, sans auto-certification (slice 25, T2)'
```

## Acceptation

- Une ligne changée dans LiveView.tsx, aucune fixture ni aucun test ne pin le tooltip actuel.
- Aucun header, corpus, schéma CSV, live, CHR, bench, merge.
- Les réglages persistés des utilisateurs sont respectés.
