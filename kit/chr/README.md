# Validation RouterOS 7 par CHR (QEMU)

Prouve qu'un RouterOS 7.x accepte le script de remédiation du rapport scellé et
que son `export` correspond à la structure voulue. Deux niveaux :
`windows-tcg` (import + export + diff) et `linux-namespace` (idem, plus les
interfaces CHR attachées aux namespaces du banc).

## Prérequis humains (non automatisables)

1. Installer QEMU (`qemu-system-x86_64`). Absent sur DESKTOP-B6K9UCM au
   2026-10-30 ; l'installation est une décision de l'humain.
2. Télécharger l'image CHR sur mikrotik.com et accepter la licence (CHR
   gratuit à débit limité ou licence d'évaluation). Ne jamais commiter l'image.
3. Placer l'image dans `kit/chr/cache/`, renseigner `chr-image.sha256`.

## Usage

```bash
bash kit/chr/setup-mikrotik-chr.sh --dry-run \
  --image kit/chr/cache/chr-7.15.3.img --level windows-tcg

bash kit/chr/setup-mikrotik-chr.sh \
  --report data/audits/<audit_id>/audit_report.json \
  --image kit/chr/cache/chr-7.15.3.img --level windows-tcg
```

Le sceau est écrit dans `data/chr/<audit_id>/` (`proof.json`, `export.txt`,
`transcript.txt`) ; jamais commité. Un refus de RouterOS est conservé mot pour
mot et la preuve dit `accepted: false` — jamais « validée ».

Rejeu sans VM (utile en revue) :

```bash
bash kit/chr/setup-mikrotik-chr.sh --report <rapport.json> \
  --export-file kit/chr/export-fixture.txt --level windows-tcg
```

## Ce que cette preuve ne dit pas

Elle ne prouve ni débit, ni latence, ni qu'une production a changé. Le niveau
`linux-namespace` ne s'obtient que si l'attachement aux namespaces réussit.
