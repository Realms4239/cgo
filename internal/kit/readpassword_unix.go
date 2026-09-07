//go:build !windows

package kit

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// readPassword — saisie masquée via stty sur le stdin hérité (le TUI
// suspend avec le vrai terminal). Même contrat que la version Windows.
func readPassword(prompt string) (string, error) {
	stty := exec.Command("stty", "-echo")
	stty.Stdin = os.Stdin
	if err := stty.Run(); err != nil {
		return "", fmt.Errorf("entrée non interactive — relancer dans un terminal")
	}
	defer exec.Command("stty", "echo").Run()
	fmt.Fprint(os.Stderr, prompt)
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}
