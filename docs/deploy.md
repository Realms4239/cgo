# Déploiement — installation machine propre (Meteolink 1.2.3)

Ce document couvre l'installation sur une machine vierge : Windows en
**mode observation**, Linux en **mode complet**, la VM de banc et la
recette mini-PC edge.

Les deux modes :

- **Windows = mode observation** : audit du lien + consultation
  (résultats, tableau live, provenance, journal). Campagne, façonnage
  et surveillance renvoient HTTP 501 avec un message clair.
- **Linux = mode complet** : tout est disponible (campagne, façonnage
  du bord, surveillance continue).

Le mode se force via `--mode observe|full|auto` (défaut `auto`).
Vérification : `GET /api/health` → `{ok, version, mode}`.

---

## 1. Windows — mode observation

Deux voies, au choix :

```powershell
# Binaire précompilé depuis GitHub Releases
#    Télécharger l'archive de plateforme, décompresser, lancer cgo.exe
```

Aucune dépendance système : le mode observation n'a pas besoin de `tc`
ni de privilèges réseau. L'audit du lien (client-side) fonctionne tel
quel :

```powershell
cgo audit --link-type 5g --site "Site X" --duration 300
```

## 2. Linux — mode complet

### 2.1 Installation des prérequis

```bash
bash kit/install.sh
```

Le script installe `iproute2`, `curl`, `bc`, charge le module
`tcp_bbr`, vérifie la présence de `tc` et signale l'absence de
CAP_NET_ADMIN. Il **ne change jamais** l'algorithme de contrôle de
congestion par défaut du système.

### 2.2 Diagnostic

```bash
cgo doctor
```

Vérifie : mode OS, `tc` dans le PATH, CAP_NET_ADMIN/root, BBR
disponible, `ping`. Ne passez pas à la suite tant qu'un point reste rouge.

### 2.3 Privilèges de façonnage

Le façonnage (`tc`/netem) exige CAP_NET_ADMIN. Deux options :

```bash
# Option a : capability sur le binaire (recommandé, pas de root au runtime)
sudo setcap cap_net_admin+ep ./cgo

# Option b : lancer le serveur en root
sudo ./cgo --serve
```

### 2.4 Service systemd

```bash
cgo service install
```

Installe l'unité `/etc/systemd/system/meteolink.service` puis active le
démarrage automatique (`systemctl enable --now`). Le service démarre en
mode `auto` : Linux complet.

### 2.5 Nettoyage du façonnage

`cgo shape --restore` garantit le **nettoyage des qdiscs** appliqués
par le façonnage (netem, cake, fq_codel…). Il est aussi exécuté
automatiquement **au démarrage du serveur** : même après un arrêt
brutal, le bord ne reste jamais façonné à l'insu de l'opérateur.

## 3. VM de banc (développement / mesure contrôlée)

Le déploiement hôte → VM est piloté par `cgo kit` (moteur Go, mêmes codes de
sortie que l'ancien bash ; `kit/engine.sh --action X` reste un shim de compat) :

```bash
cgo kit scan      # trouver le .vmx/.vbox, sauvegarder l'unique
cgo kit ensure    # démarrer la VM si SSH est down, attendre SSH
cgo kit deploy    # push + install + health-check (recompile si sources dispo, sinon binaire précompilé)
cgo kit status    # SSH + process + santé du dashboard
cgo kit logs      # tail du log serveur sur la VM
cgo kit align     # NIC vmxnet3 + CPU/mémoire mini du banc (à froid)
```

Le déploiement est **binaire-first et idempotent** : la SPA est
embarquée dans le binaire, aucune toolchain Go n'est requise sur la VM.

Prérequis VM (une fois) :

```bash
sudo apt update && sudo apt install -y iproute2 curl bc
sudo modprobe tcp_bbr || true
```

## 4. Recette mini-PC edge (Debian)

Pour un banc fixe sur site :

1. **Debian** vierge, accès sudo.
2. Installer les prérequis : `bash kit/install.sh`.
3. Installer le binaire (archive de la Release GitHub : `cgo-linux-amd64.tar.gz`).
4. Diagnostic : `cgo doctor` — tout vert.
5. Privilèges : `sudo setcap cap_net_admin+ep ./cgo`.
6. Service : `cgo service install`.
7. **Réseau** : ouvrir le **port 9090** (dashboard + API) —
   `sudo ufw allow 9090/tcp` ou équivalent.
8. Vérifier : `curl -k https://<ip-edge>:9090/api/health` →
   `{"ok":true,"version":"1.2.3","mode":"full"}` (-k : certificat auto-signé).

Le mini-PC est alors prêt pour la boucle opérateur complète
(auditer → campagner → façonner → surveiller → comparer → constat),
voir `docs/usage.fr.md`.

## 5. Checklist de mise en service

```text
[ ] /api/health renvoie {ok:true, version, mode attendu}
[ ] cgo doctor passe (Linux : tc, CAP_NET_ADMIN, bbr, ping)
[ ] port 9090 ouvert (accès distant)
[ ] cgo shape --restore exécuté une fois (bord propre au départ)
[ ] service systemd actif : systemctl status meteolink
```
