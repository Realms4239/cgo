package main

import (
	"bytes"
	"io"
	"strings"
	"sync"
	"testing"
	"time"
)

// replayCHR rejoue la console Manager MikroTik 7.24.6 telle qu'observée sur
// l'hôte de banc : chaque ligne tapée produit le prompt d'abord (marqueur
// curseur), puis l'écho de la commande, puis un prompt de réponse. C'est la
// séquence qui a perdu les écritures 11-16 en production.
type replayCHR struct {
	mu     sync.Mutex
	buf    bytes.Buffer
	closed bool
}

func newReplayCHR() *replayCHR {
	f := &replayCHR{}
	f.buf.WriteString("MikroTik RouterOS 7.24.6 (long-term)\r\n\r\nCHR Login: ")
	return f
}

func (f *replayCHR) Read(p []byte) (int, error) {
	for {
		f.mu.Lock()
		if f.buf.Len() > 0 {
			n, _ := f.buf.Read(p)
			f.mu.Unlock()
			return n, nil
		}
		if f.closed {
			f.mu.Unlock()
			return 0, io.EOF
		}
		f.mu.Unlock()
		time.Sleep(time.Millisecond)
	}
}

func (f *replayCHR) emit(s string) {
	f.mu.Lock()
	f.buf.WriteString(s)
	f.mu.Unlock()
}

func (f *replayCHR) Write(p []byte) (int, error) {
	line := strings.TrimSpace(string(p))
	switch {
	case line == "admin" || line == "admin+t500w":
		f.emit("\r\nPassword: ")
	case line == "":
		f.emit("\r\nChange your password (Ctrl-C to skip)\r\n\r\nnew password> ")
	case line == "chr-test-fixture":
		f.emit("**************************\r\nnew password> **************************\r\n\r\nrepeat new password> ")
	case strings.HasPrefix(line, "repeat-marker"):
		// unused — the repeat confirmation arrives as a distinct Write below
		f.emit("Password changed\r\n\r\n[admin@CHR] > ")
	case strings.HasPrefix(line, "/export"):
		f.emit("\r\n[admin@CHR] > /export terse\r\n/queue type\r\nadd cake-rtt=25ms cake-bandwidth=19000000 cake-flowmode=triple-isolate cake-nat=yes cake-diffserv=diffserv4 cake-ack-filter=filter cake-wash=yes kind=cake name=cgo-cake\r\n/queue tree\r\nadd name=cgo-wan-out parent=ether1 max-limit=19000000 queue=cgo-cake comment=cgo-managed\r\n[admin@CHR] > ")
	default:
		f.emit("\r\n[admin@CHR] > " + line + "\r\n[admin@CHR] > ")
	}
	return len(p), nil
}

func (f *replayCHR) Close() error {
	f.mu.Lock()
	f.closed = true
	f.mu.Unlock()
	return nil
}

// colorCHR — le CHR avec largeur +tNNNw allume les couleurs du prompt :
// « [admin@CHR] > » devient des séquences CSI. Le test doit prouver que le
// moteur passe quand même. C'est exactement le cas qui a bloqué la console
// en production (2026-11-16, 300 s de délai).
type colorCHR struct{ replayCHR }

const esc = "\x1b"

func colored(s string) string {
	return strings.ReplaceAll(s, "[admin@CHR] >", esc+"[9999B"+esc+"[K"+esc+"[m"+esc+"[36madmin"+esc+"[m@"+esc+"[m"+esc+"[32mCHR"+esc+"[m] >")
}

func (c *colorCHR) emit(s string) {
	c.replayCHR.emit(colored(s))
}

func newColorCHR() *colorCHR {
	f := &colorCHR{}
	f.buf.WriteString("MikroTik RouterOS 7.24.6 (long-term)\r\n\r\nCHR Login: ")
	return f
}

func TestColoredPromptConsolePath(t *testing.T) {
	tr := newColorCHR()
	script := "/queue type add name=cgo-cake kind=cake\n/queue tree add name=cgo-wan-out parent=ether1 queue=cgo-cake max-limit=19000000\n"
	transcript, _, importErr, err := drive(tr, script, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if importErr != "" {
		t.Fatalf("import refusé inattendu: %s", importErr)
	}
	if !strings.Contains(transcript, "cgo-cake") {
		t.Fatalf("transcript coloré sans la ligne attendue\n---\n%s", transcript)
	}
}

func TestReplayConsoleContinuationBlock(t *testing.T) {
	tr := newReplayCHR()
	script := "/queue type add name=cgo-cake kind=cake \\\n    cake-bandwidth=19000000 cake-rtt=25ms \\\n    cake-flowmode=triple-isolate cake-wash=yes\n"
	transcript, _, importErr, err := drive(tr, script, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if importErr != "" {
		t.Fatalf("import refusé inattendu: %s", importErr)
	}
	for _, want := range []string{"cake-bandwidth", "cake-flowmode", "cake-wash"} {
		if !strings.Contains(transcript, want) {
			t.Fatalf("transcript sans %q\n---\n%s", want, transcript)
		}
	}
}

func TestReplayConsoleCadenceFullLineDelivery(t *testing.T) {
	tr := newReplayCHR()
	script := "/queue tree remove [find where comment=\"cgo-managed\"]\n/queue type add name=cgo-cake kind=cake\n"
	transcript, export, importErr, err := drive(tr, script, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if importErr != "" {
		t.Fatalf("import refusé inattendu: %s", importErr)
	}
	for _, want := range []string{"/export terse", "cgo-cake"} {
		if !strings.Contains(transcript, want) {
			t.Fatalf("transcript sans %q\n---\n%s", want, transcript)
		}
	}
	if !strings.Contains(export, "cgo-wan-out") {
		t.Fatalf("export sans cgo-wan-out: %q", export)
	}
}
