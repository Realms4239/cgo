//go:build !windows

package kit

import (
	"context"
	"os/exec"
)

// bgCmd — hors Windows, pas de console à masquer : simple passthrough
// (même signature, appelants inchangés).
func bgCmd(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}

// bgCmdCtx — idem avec annulation.
func bgCmdCtx(ctx context.Context, name string, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, args...)
}
