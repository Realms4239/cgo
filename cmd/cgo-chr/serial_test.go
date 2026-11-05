package main

import (
	"bytes"
	"io"
	"strings"
	"sync"
	"testing"
	"time"
)

// fakeCHR replays a RouterOS console: banner, login, password, prompt, then
// answers each line. It is the seam that keeps CI VM-free.
// Read blocks until a Write has produced output: returning (0, nil) would trip
// bufio's empty-read limit and fail the test for the wrong reason.
type fakeCHR struct {
	mu     sync.Mutex
	buf    bytes.Buffer
	closed bool
}

func newFakeCHR() *fakeCHR {
	f := &fakeCHR{}
	f.buf.WriteString("MikroTik RouterOS 7.15.3 (c) 1999-2025\r\n\r\nlogin: ")
	return f
}

func (f *fakeCHR) Read(p []byte) (int, error) {
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

func (f *fakeCHR) emit(s string) {
	f.mu.Lock()
	f.buf.WriteString(s)
	f.mu.Unlock()
}

func (f *fakeCHR) Write(p []byte) (int, error) {
	line := strings.TrimSpace(string(p))
	switch {
	case line == "admin":
		f.emit("\r\npassword: ")
	case line == "":
		f.emit("\r\n[admin@MikroTik] > ")
	case strings.HasPrefix(line, "/export"):
		f.emit("\r\n/queue type\r\nadd cake-ack-filter=filter cake-bandwidth=19000000 cake-diffserv=diffserv4 cake-flowmode=triple-isolate cake-nat=yes cake-rtt=25ms cake-wash=yes kind=cake name=cgo-cake\r\n/queue tree\r\nadd comment=cgo-managed max-limit=19000000 name=cgo-wan-out parent=ether1-wan queue=cgo-cake\r\n[admin@MikroTik] > ")
	case line == "REJECT":
		f.emit("\r\nfailure: bad command name\r\n[admin@MikroTik] > ")
	default:
		f.emit("\r\n[admin@MikroTik] > ")
	}
	return len(p), nil
}

func (f *fakeCHR) Close() error {
	f.mu.Lock()
	f.closed = true
	f.mu.Unlock()
	return nil
}

func TestDriveHappyPath(t *testing.T) {
	tr := newFakeCHR()
	transcript, export, importErr, err := drive(tr, "/queue type add name=cgo-cake kind=cake\n", 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if importErr != "" {
		t.Fatalf("import rejected: %s", importErr)
	}
	if !strings.Contains(transcript, "7.15.3") {
		t.Fatal("transcript lacks the version banner")
	}
	if !strings.Contains(export, "name=cgo-cake") {
		t.Fatalf("export not captured: %q", export)
	}
}

func TestDriveRejection(t *testing.T) {
	tr := newFakeCHR()
	_, _, importErr, err := drive(tr, "REJECT\n", 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(importErr, "failure: bad command name") {
		t.Fatalf("RouterOS error not stored verbatim: %q", importErr)
	}
}
