package vm

import (
	"os"
	"testing"
)

// NetMode se lit dans le fichier, sans hyperviseur : ponté/nat/hôte-only.
// NOTE — runProgramInGuest/copyFile PROUVÉS INOPÉRANTS sur open-vm-tools 13
// + Workstation 1x (VIX_E_FILE_NOT_FOUND, exit 0 vide) : aucun transport
// invité par ce canal, l'apt+ssh reste la voie. Ne pas réintroduire sans
// banc apparié qui le prouve.
func TestNetModeVmx(t *testing.T) {
	dir := t.TempDir()
	cases := map[string]string{
		`ethernet0.connectionType = "bridged"`: "ponté",
		`ethernet0.connectionType = "nat"`:     "nat",
		`ethernet0.connectionType = "hostonly"`: "hôte-only",
		`ethernet0.present = "TRUE"`:            "inconnu",
	}
	v := &vmware{}
	i := 0
	for content, want := range cases {
		i++
		p := dir + "/t.vmx"
		os.WriteFile(p, []byte(content+"\n"), 0644)
		if got := v.NetMode(p); got != want {
			t.Errorf("cas %d: got %q want %q", i, got, want)
		}
	}
	if got := v.NetMode(dir + "/absent.vmx"); got != "inconnu" {
		t.Errorf("fichier absent: got %q", got)
	}
}
