package kit

import (
	"archive/tar"
	"compress/gzip"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Realms4239/cgo/internal/vm"
)

// ---- snapshot / revert — le garde-fou : point de restauration avant
// toute opération risquée (align, déploiements répétés). L'incident
// "VM réseau cassé après hard stops" n'aurait jamais coûté une heure :
// un revert et retour à l'état sain.

// Snapshot — crée un point de restauration de la VM (VMware) ;
// VirtualBox : snapshots natifs VBoxManage snapshot take.
func (r *Runner) Snapshot(c *Config, name string, deep bool) int {
	hyp, vmx, err := r.pickVM(c, deep)
	if err != nil {
		r.errf("[snapshot] %v", err)
		return 3
	}
	// une VM allumée doit être arrêtée proprement pour un snapshot stable
	live := false
	for _, run := range hyp.Running() {
		if samePath(run, vmx) {
			live = true
		}
	}
	if live {
		r.out("[snapshot] arrêt propre de la VM (le snapshot se prend à froid)…")
		if err := hyp.Stop(vmx); err != nil {
			r.errf("[snapshot] arrêt échoué : %v", err)
			return 4
		}
		time.Sleep(2 * time.Second)
	}
	switch h := hyp.(type) {
	case interface{ Exe() string }:
		_ = h
	}
	if strings.Contains(filepath.Base(hyp.Exe()), "vmrun") {
		out, err := runCmd(hyp.Exe(), "snapshot", vmx, name)
		if err != nil {
			r.errf("[snapshot] vmrun échoué : %s", out)
			return 4
		}
		r.out("[snapshot] %s posé (%s)", name, filepath.Base(vmx))
		return 0
	}
	// VirtualBox
	vboxName := strings.TrimSuffix(filepath.Base(vmx), ".vbox")
	out, err := runCmd(hyp.Exe(), "snapshot", vboxName, "take", name)
	if err != nil {
		r.errf("[snapshot] VBoxManage échoué : %s", out)
		return 4
	}
	r.out("[snapshot] %s posé (%s)", name, vboxName)
	return 0
}

// Revert — revient au snapshot nommé (défaut : dernier trouvé).
func (r *Runner) Revert(c *Config, name string, deep bool) int {
	hyp, vmx, err := r.pickVM(c, deep)
	if err != nil {
		r.errf("[revert] %v", err)
		return 3
	}
	if name == "" {
		// dernier snapshot listé
		if strings.Contains(filepath.Base(hyp.Exe()), "vmrun") {
			out, _ := runCmd(hyp.Exe(), "listsnapshots", vmx)
			lines := strings.Split(strings.TrimSpace(out), "\n")
			for i := len(lines) - 1; i >= 0; i-- {
				l := strings.TrimSpace(lines[i])
				if l != "" && !strings.HasPrefix(l, "Total") {
					name = l
					break
				}
			}
		} else {
			vboxName := strings.TrimSuffix(filepath.Base(vmx), ".vbox")
			out, _ := runCmd(hyp.Exe(), "snapshot", vboxName, "list", "--machinereadable")
			for _, ln := range strings.Split(out, "\n") {
				if strings.HasPrefix(ln, "SnapshotName") {
					f := strings.SplitN(ln, "=", 2)
					if len(f) == 2 {
						name = strings.Trim(f[1], "\"")
					}
				}
			}
		}
	}
	if name == "" {
		r.errf("[revert] aucun snapshot trouvé — faites d'abord : cgo kit snapshot")
		return 3
	}
	if strings.Contains(filepath.Base(hyp.Exe()), "vmrun") {
		out, err := runCmd(hyp.Exe(), "revertToSnapshot", vmx, name)
		if err != nil {
			r.errf("[revert] échoué : %s", out)
			return 4
		}
	} else {
		vboxName := strings.TrimSuffix(filepath.Base(vmx), ".vbox")
		out, err := runCmd(hyp.Exe(), "snapshot", vboxName, "restore", name)
		if err != nil {
			r.errf("[revert] échoué : %s", out)
			return 4
		}
	}
	r.out("[revert] VM revenue au snapshot %s — relancez : cgo kit ensure", name)
	return 0
}

// ---- ssh — shell direct : l'opérateur reste dans son terminal.

// SSHInteractive — ouvre un shell SSH complet (stdin/stdout/stderr branchés).
func (r *Runner) SSHInteractive(c *Config) int {
	key := expandKey(c.SSHKey)
	// VM propre : pré-diagnostiquer avant d'ouvrir le shell, pour que
	// l'opérateur voie la cause et la remédiation plutôt qu'un refus ssh sec.
	if !c.SSHUp() {
		probe, _ := c.SSH("true")
		if cls := classifySSHError(probe); cls != "" {
			r.sshDiag("[ssh]", probe)
			return 5
		}
	}
	cmd := exec.Command("ssh", "-o", "ConnectTimeout=6", "-o", "StrictHostKeyChecking=accept-new",
		"-p", c.SSHPort, "-i", key, c.SSHUser+"@"+c.SSHHost)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	r.out("[ssh] %s@%s — Ctrl-D pour sortir", c.SSHUser, c.SSHHost)
	return exitCode(cmd.Run())
}

// ---- ps — contrôle en direct du dashboard (2 Hz).

// Ps — rafraîchit le processus serveur + santé toutes les 2 s ; q pour sortir.
func (r *Runner) Ps(c *Config) int {
	r.out("[ps] Ctrl-C pour sortir")
	for {
		out, err := c.SSH("ps -o pid,etime,pcpu,pmem,cmd -C cgo-linux --no-headers 2>/dev/null || pgrep -af cgo-linux")
		if err != nil {
			r.out("\r\033[K[ps] SSH injoignable")
		} else {
			health := "KO"
			if c.Health() {
				health = "OK"
			}
			fmt.Fprintf(r.Stdout, "\r\033[K[ps] health=%s | %s", health, strings.ReplaceAll(strings.TrimSpace(out), "\n", " · "))
		}
		time.Sleep(2 * time.Second)
	}
}

// ---- svc — pilotage du dashboard distant (start/stop/restart/status).
// Le lanceur posé par deploy (start.sh + nohup userspace, voir vm-install.sh)
// ne demande aucun sudo : tout passe par la clé SSH (keysetup). Zéro GUI.

// Svc — start : lance (idempotent) + attend la santé 15 s ; stop : coupe et
// vérifie ; restart : stop + start ; status (défaut) : processus + santé.
func (r *Runner) Svc(c *Config, sub string) int {
	procs, _ := c.SSH(`pgrep -u "$USER" -f '[c]go-linux --serve' | tr '\n' ' '`)
	procs = strings.TrimSpace(procs)
	switch sub {
	case "", "status":
		r.out("[svc] process=%s health=%s (%s@%s)", firstOr(procs, "—"),
			humanBool(c.Health()), c.SSHUser, c.SSHHost)
		return 0
	case "stop":
		if procs == "" {
			r.out("[svc] déjà arrêté")
			return 0
		}
		_, _ = c.SSH(`pkill -u "$USER" -f '[c]go-linux --serve'`)
		for i := 0; i < 10; i++ {
			time.Sleep(time.Second)
			if out, _ := c.SSH(`pgrep -u "$USER" -f '[c]go-linux --serve'`); strings.TrimSpace(out) == "" {
				r.out("[svc] arrêté (pid %s)", procs)
				return 0
			}
		}
		r.errf("[svc] arrêt incomplet, pids restants — voir : cgo kit ps")
		return 5
	case "start":
		if c.Health() {
			r.out("[svc] déjà en ligne (pid %s)", firstOr(procs, "?"))
			return 0
		}
		// le lanceur d'un vieux deploy démarre en --tls=false : le soigner
		// avant de lancer, sinon le health HTTPS échoue sur un dashboard
		// sain en HTTP brut (faux KO). Idempotent : sans le flag, no-op.
		if out, _ := c.SSH("cd " + c.ProjectDir + " && grep -q -- '--tls=false' start.sh 2>/dev/null && sed -i 's/--tls=false //' start.sh && echo healed || true"); strings.Contains(out, "healed") {
			r.out("[svc] lanceur guéri (TLS) — redéployez pour la version canonique (kit deploy)")
		}
		if _, err := c.SSH("cd " + c.ProjectDir + " && (setsid nohup ./start.sh >/dev/null 2>&1 &)"); err != nil {
			// un seul essai manqué ne doit pas laisser le banc sans dashboard
			// (vu en prod : SSH vide juste après un stop) — on réessaie.
			time.Sleep(2 * time.Second)
			if _, err2 := c.SSH("cd " + c.ProjectDir + " && (setsid nohup ./start.sh >/dev/null 2>&1 &)"); err2 != nil {
				r.sshDiag("[svc]", "")
				return 5
			}
		}
		for i := 0; i < 15; i++ {
			time.Sleep(time.Second)
			if c.Health() {
				r.out("[svc] en ligne → %s", c.dashURL())
				return 0
			}
		}
		r.errf("[svc] santé KO après 15 s — voir : cgo kit logs")
		return 8
	case "restart":
		if code := r.Svc(c, "stop"); code != 0 {
			return code
		}
		return r.Svc(c, "start")
	default:
		r.errf("[svc] sous-commande inconnue : %s (start|stop|restart|status)", sub)
		return 2
	}
}

func firstOr(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

func humanBool(b bool) string {
	if b {
		return "OK"
	}
	return "KO"
}

// Backup — tar.gz horodaté de data/runs de la VM vers dest/ (défaut ./backup).
func (r *Runner) Backup(c *Config, dest string) int {
	if err := os.MkdirAll(dest, 0o755); err != nil {
		r.errf("[backup] %v", err)
		return 7
	}
	stamp := time.Now().Format("20060102-150405")
	remote := "/tmp/cgo-backup-" + stamp + ".tar.gz"
	r.out("[backup] tar des runs sur la VM…")
	if _, err := c.SSH("cd " + c.ProjectDir + " && tar czf " + remote + " data/runs"); err != nil {
		r.errf("[backup] tar distant échoué")
		return 7
	}
	local := filepath.Join(dest, "runs-"+stamp+".tar.gz")
	// scp : c.SCP pousse local→distant ; ici il faut l'inverse : ssh cat.
	key := expandKey(c.SSHKey)
	cat := exec.Command("ssh", "-o", "ConnectTimeout=6", "-p", c.SSHPort, "-i", key,
		c.SSHUser+"@"+c.SSHHost, "cat "+remote)
	f, err := os.Create(local)
	if err != nil {
		r.errf("[backup] %v", err)
		return 7
	}
	pipe, err := cat.StdoutPipe()
	if err != nil {
		return 7
	}
	cat.Stderr = os.Stderr
	if err := cat.Start(); err != nil {
		return 7
	}
	n, cpErr := io.Copy(f, pipe)
	_ = cat.Wait()
	f.Close()
	if cpErr != nil || n == 0 {
		r.errf("[backup] transfert vide — SSH vers la VM ?")
		return 7
	}
	_, _ = c.SSH("rm -f " + remote)
	// vérification locale : le tar.gz est-il sain ?
	if err := verifyTarGz(local); err != nil {
		r.errf("[backup] archive corrompue : %v", err)
		return 7
	}
	r.out("[backup] %s (%d octets, archive vérifiée)", local, n)
	return 0
}

// verifyTarGz — l'archive s'ouvre et contient au moins un aqm_eval.csv.
func verifyTarGz(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	saw := false
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if strings.HasSuffix(hdr.Name, "aqm_eval.csv") {
			saw = true
		}
	}
	if !saw {
		return fmt.Errorf("aucun aqm_eval.csv dans l'archive")
	}
	return nil
}

// ---- utilitaires

// Snapshots — liste les instantanés de la VM (nom de l'hyperviseur en tête).
// Le numéro sert directement à : cgo kit revert <nom>.
func (r *Runner) Snapshots(c *Config, deep bool) int {
	hyp, vmx, err := r.pickVM(c, deep)
	if err != nil {
		r.errf("[snapshots] %v", err)
		return 3
	}
	var names []string
	if strings.Contains(filepath.Base(hyp.Exe()), "vmrun") {
		out, err := runCmd(hyp.Exe(), "listsnapshots", vmx)
		if err != nil {
			r.errf("[snapshots] vmrun échoué : %s", out)
			return 4
		}
		for _, ln := range strings.Split(out, "\n") {
			l := strings.TrimSpace(ln)
			if l != "" && !strings.HasPrefix(l, "Total") {
				names = append(names, l)
			}
		}
	} else {
		vboxName := strings.TrimSuffix(filepath.Base(vmx), ".vbox")
		out, err := runCmd(hyp.Exe(), "snapshot", vboxName, "list", "--machinereadable")
		if err != nil {
			r.errf("[snapshots] VBoxManage échoué : %s", out)
			return 4
		}
		for _, ln := range strings.Split(out, "\n") {
			if strings.HasPrefix(ln, "SnapshotName") {
				f := strings.SplitN(ln, "=", 2)
				if len(f) == 2 {
					names = append(names, strings.Trim(f[1], "\""))
				}
			}
		}
	}
	if len(names) == 0 {
		r.out("[snapshots] aucun instantané — posez-en un : cgo kit snapshot")
		return 0
	}
	r.out("[snapshots] %s (%d) :", filepath.Base(vmx), len(names))
	for _, n := range names {
		r.out("  %s", n)
	}
	r.out("retour : cgo kit revert <nom>")
	return 0
}

// Verify — recalcul des empreintes SHA-256 des manifests SUR LA VM.
// Le binaire distant fait le travail ; on ne rapatrie rien.
func (r *Runner) Verify(c *Config) int {
	r.out("[verify] empreintes des archives sur %s…", c.SSHHost)
	out, err := c.SSH("cd " + c.ProjectDir + " && ./cgo-linux verify 2>&1 || cgo verify 2>&1")
	if err != nil {
		r.sshDiag("[verify]", out)
		return 5
	}
	r.out("%s", strings.TrimRight(out, "\n"))
	if strings.Contains(out, "échec") || strings.Contains(out, "incohérent") {
		return 8
	}
	return 0
}

// Health — état du dashboard distant : JSON complet + verdict une ligne.
func (r *Runner) Health(c *Config) int {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	cl := &http.Client{Timeout: 4 * time.Second, Transport: tr}
	resp, err := cl.Get(c.healthURL())
	if err != nil {
		r.errf("[health] injoignable : %v", err)
		return 5
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		r.errf("[health] lecture échouée : %v", err)
		return 5
	}
	r.out("[health] %s", string(body))
	if !strings.Contains(string(body), `"ok":true`) {
		return 8
	}
	return 0
}

func expandKey(k string) string {
	if strings.HasPrefix(k, "~/") {
		if h := userHome(); h != "" {
			return filepath.Join(h, k[2:])
		}
	}
	return k
}

func runCmd(exe string, args ...string) (string, error) {
	out, err := exec.Command(exe, args...).CombinedOutput()
	return string(out), err
}

func samePath(a, b string) bool {
	return strings.EqualFold(filepath.Clean(a), filepath.Clean(b))
}

// unused guard — vm importé pour NatForwarder dans Ensure.
var _ vm.Hypervisor = (vm.Hypervisor)(nil)
