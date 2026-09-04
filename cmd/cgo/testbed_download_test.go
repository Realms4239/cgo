package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/Realms4239/cgo/pkg/probe"
)

// TestTestbedDownloadProtocol — le serveur du banc distingue les sens sur un
// seul port : "D" = il inonde (download), sinon puits (upload). Le client
// download reçoit des octets via BulkDownloadTo.
func TestTestbedDownloadProtocol(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	addr := ln.Addr().String()

	// comportement du serveur : lecture du 1er octet, dispatch
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				c.SetReadDeadline(time.Now().Add(2 * time.Second))
				hdr := make([]byte, 1)
				if _, err := io.ReadFull(c, hdr); err == nil && hdr[0] == 'D' {
					ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
					defer cancel()
					_, _ = probe.BulkSource(ctx, c)
					return
				}
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				_, _ = probe.BulkReceive(ctx, c, nil)
			}(c)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	n, err := probe.BulkDownloadTo(ctx, addr)
	if err != nil {
		t.Fatalf("download: %v", err)
	}
	if n < 64*1024 {
		t.Fatalf("download n'a reçu que %d octets", n)
	}
	fmt.Printf("download OK: %d octets reçus en 2s\n", n)
}
