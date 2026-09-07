//go:build windows

package kit

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/sys/windows"
)

// readPassword — saisie masquée via la console Win32 (pas d'astérisques,
// silence total). Le secret ne va ni dans les logs, ni dans la config,
// ni dans l'historique : lu, utilisé, oublié. Terminal requis.
func readPassword(prompt string) (string, error) {
	h := windows.Handle(os.Stdin.Fd())
	var mode uint32
	if err := windows.GetConsoleMode(h, &mode); err != nil {
		return "", fmt.Errorf("entrée non interactive — relancer dans un terminal")
	}
	if err := windows.SetConsoleMode(h, mode & ^uint32(windows.ENABLE_ECHO_INPUT)); err != nil {
		return "", err
	}
	defer windows.SetConsoleMode(h, mode)
	fmt.Fprint(os.Stderr, prompt)
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}
