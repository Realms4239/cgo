//go:build windows

package kit

import (
	"context"
	"os/exec"

	"golang.org/x/sys/windows"
)

// bgCmd — subprocess SANS fenêtre console (CREATE_NO_WINDOW) : depuis
// cgo-gui.exe (sous-système GUI), chaque ssh/scp/vmrun spawné ouvrait un
// flash de console — le « cmd openssh qui s'ouvre au scan ». Réservé aux
// commandes batch à sortie capturée ; JAMAIS pour l'interactif (mot de
// passe keysetup, installs) dont le prompt doit rester visible.
func bgCmd(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &windows.SysProcAttr{CreationFlags: windows.CREATE_NO_WINDOW}
	return cmd
}

// bgCmdCtx — idem avec annulation (sondes bornées).
func bgCmdCtx(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.SysProcAttr = &windows.SysProcAttr{CreationFlags: windows.CREATE_NO_WINDOW}
	return cmd
}
