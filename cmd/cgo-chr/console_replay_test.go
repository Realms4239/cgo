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
	case line == "admin":
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
