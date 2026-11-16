package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	loginRe  = regexp.MustCompile(`(?i)login:\s*$`)
	passRe   = regexp.MustCompile(`(?i)password:\s*$`)
	promptRe = regexp.MustCompile(`\[admin@[^\]]*\]\s*>\s*$`)
	// premier démarrage CHR : la console demande d'abord d'agréer le contrat
	// de licence (réponse n), puis exige un mot de passe admin (choix jetable,
	// local à l'essai, jamais un secret). Le serveur renvoie EXACTEMENT les
	// octets reçus — le transcript documente la séquence réelle.
	licenseRe = regexp.MustCompile(`(?i)license`)
	// marqueurs exacts de la console CHR neuve — jamais un bare "password"
	// (RouterOS répète le mot dans des lignes intermédiaires) ; l'écho masqué
	// est une ligne « new password> ***... » (étoiles SUR LA MÊME LIGNE que le
	// prompt), jamais la barre de progression précédente.
	newPassRe    = regexp.MustCompile(`(?i)new password`)
	repeatRe     = regexp.MustCompile(`(?i)repeat new password`)
	changedRe    = regexp.MustCompile(`(?i)(changed|successfully)`)
	errorRe      = regexp.MustCompile(`(?i)(failure:|syntax error|bad command name|expected end of command|input does not match)`)
	loginAfterRe = regexp.MustCompile(`(?i)((\[admin@[^\]]*\]\s*>)|license|new password|repeat new password|changed|successfully)`)
)

var errTimeout = errors.New("chr: délai dépassé")

// console — un seul goroutine pompe octets par octets depuis la console du
// CHR ; tous les expect() puisent dans ce même tuyau. Les octets sont rendus
// dans l'ordre, jamais dérobés par un lecteur zombie (c'est l'implémentation
// qui correcte pour une console bavarde partagée).
type console struct {
	bytes chan byte
	errs  chan error
	done  chan struct{}
	once  sync.Once
}

func newConsole(r io.ReadWriteCloser) *console {
	c := &console{
		bytes: make(chan byte, 64),
		errs:  make(chan error, 1),
		done:  make(chan struct{}),
	}
	go func() {
		defer close(c.done)
		br := bufio.NewReader(r)
		for {
			b, err := br.ReadByte()
			if err != nil {
				c.errs <- err
				return
			}
			c.bytes <- b
		}
	}()
	return c
}

// expect accumule jusqu'à ce que le tampon se termine par re ET que le tuyau
// soit muet depuis 40 ms (réponse complète), ou qu'une erreur arrive, ou que
// le délai global soit dépassé.
func (c *console) expect(re *regexp.Regexp, d time.Duration) (string, error) {
	start := time.Now()
	var out strings.Builder
	for {
		select {
		case err := <-c.errs:
			return out.String(), err
		case b := <-c.bytes:
			out.WriteByte(b)
		case <-time.After(40 * time.Millisecond):
			if re.MatchString(out.String()) {
				return out.String(), nil
			}
			if time.Since(start) > d {
				return out.String(), errTimeout
			}
		}
	}
}

// drive logs in, sends the script line by line, then captures the export.
// The transcript keeps every byte the device sent, including refusals.
func drive(tr io.ReadWriteCloser, script string, timeout time.Duration) (transcript, export, importErr string, err error) {
	cons := newConsole(tr)
	var all strings.Builder
	read := func(re *regexp.Regexp) (string, error) {
		chunk, err := cons.expect(re, timeout)
		all.WriteString(chunk)
		return chunk, err
	}
	if _, err := read(loginRe); err != nil {
		return all.String(), "", "", err
	}
	if _, err := tr.Write([]byte("admin\n")); err != nil {
		return all.String(), "", "", err
	}
	if _, err := read(passRe); err != nil {
		return all.String(), "", "", err
	}
	if _, err := tr.Write([]byte("\n")); err != nil {
		return all.String(), "", "", err
	}
	// handshake premier démarrage : la console CHR neuve enchaîne sur la
	// licence, puis sur l'exigence d'un mot de passe admin (new password),
	// puis une confirmation (repeat new password). Réponses : n, puis un
	// jetable, le même pour la confirmation.
	throwaway := fmt.Sprintf("chr-test-%s", time.Now().UTC().Format("2006-01-02-150405"))
	wroteNew, wroteRepeat := false, false
	for attempt := 0; attempt < 8; attempt += 1 {
		chunk, err := read(loginAfterRe)
		if err != nil {
			return all.String(), "", "", err
		}
		switch {
		case promptRe.MatchString(chunk):
			goto loggedIn
		case licenseRe.MatchString(chunk):
			if _, err := tr.Write([]byte("n\n")); err != nil {
				return all.String(), "", "", err
			}
		case repeatRe.MatchString(chunk):
			// l'écho masqué de notre saisie re-matche « repeat » : on ne
			// répond qu'une fois par étape.
			if !wroteRepeat {
				if _, err := tr.Write([]byte(throwaway + "\n")); err != nil {
					return all.String(), "", "", err
				}
				wroteRepeat = true
			}
		case newPassRe.MatchString(chunk):
			if !wroteNew {
				if _, err := tr.Write([]byte(throwaway + "\n")); err != nil {
					return all.String(), "", "", err
				}
				wroteNew = true
			}
		case changedRe.MatchString(chunk):
			// confirmation — on consomme la ligne, le prompt suit.
		case errorRe.MatchString(chunk):
			return all.String(), "", "", fmt.Errorf("échec console: %s", lastErrorLine(chunk))
		default:
			tail := chunk
			if len(tail) > 120 {
				tail = tail[len(tail)-120:]
			}
			return all.String(), "", "", fmt.Errorf("console inattendue après mot de passe: %q", tail)
		}
	}
	return all.String(), "", "", fmt.Errorf("plage premier démarrage épuisée (8 cycles)")

loggedIn:
	seq := 0
	for _, line := range strings.Split(script, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		seq += 1
		if _, err := tr.Write([]byte(line + "\n")); err != nil {
			return all.String(), "", "", err
		}
		if strings.HasSuffix(line, "\\") {
			// bloc à continuation (« ligne \ ») : la console imprime une
			// invite « \... » — aucun prompt ne vient ici. La réponse
			// arrive à la fin du bloc (ligne sans antislash final).
			continue
		}
		chunk, err := cons.expect(promptRe, timeout)
		all.WriteString(chunk)
		if err != nil {
			return all.String(), "", "", err
		}
		if errorRe.MatchString(chunk) {
			return all.String(), "", lastErrorLine(chunk), nil
		}
	}
	if _, err := tr.Write([]byte("/export terse\n")); err != nil {
		return all.String(), "", "", err
	}
	chunk, err := cons.expect(promptRe, timeout)
	all.WriteString(chunk)
	if err != nil {
		return all.String(), "", "", err
	}
	return all.String(), cleanupExport(chunk), "", nil
}

// cleanupExport drops the echoed command and the trailing prompt, keeping
// only what RouterOS printed for /export.
func cleanupExport(chunk string) string {
	if i := strings.Index(chunk, "\n"); i >= 0 {
		chunk = chunk[i+1:]
	}
	if i := strings.LastIndex(chunk, "[admin@"); i >= 0 {
		chunk = chunk[:i]
	}
	return strings.TrimSpace(strings.ReplaceAll(chunk, "\r", ""))
}

func lastErrorLine(chunk string) string {
	lines := strings.Split(strings.ReplaceAll(chunk, "\r", ""), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if errorRe.MatchString(lines[i]) {
			return lines[i]
		}
	}
	return ""
}
