package kit

import (
	"context"
	"os/exec"
)

// BgCmd/BgCmdCtx — version exportée de bgCmd (vues GUI : cgo-gui lance
// les mêmes enfants batch sans flash console sous Windows).
func BgCmd(name string, args ...string) *exec.Cmd {
	return bgCmd(name, args...)
}

// BgCmdCtx — idem avec annulation.
func BgCmdCtx(ctx context.Context, name string, args ...string) *exec.Cmd {
	return bgCmdCtx(ctx, name, args...)
}
