package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Realms4239/cgo/pkg/probe"
)

// runTestbedSrv serves the small object over HTTP and the bulk sink,
// both bound inside the cgo-srv netns (deploy/testbed.sh).
func runTestbedSrv(httpAddr, bulkAddr string) error {
	obj := os.Getenv("CGO_TESTBED_OBJ")
	if obj == "" {
		obj = "testbed/obj16.bin"
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/small", func(w http.ResponseWriter, _ *http.Request) {
		b, err := os.ReadFile(obj)
		if err != nil {
			http.Error(w, "obj missing", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Length", strconv.Itoa(len(b)))
		w.Write(b)
	})
	hs := &http.Server{Addr: httpAddr, Handler: mux}
	ln, err := net.Listen("tcp", bulkAddr)
	if err != nil {
		return fmt.Errorf("bulk listen: %w", err)
	}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
				defer cancel()
				_, _ = probe.BulkReceive(ctx, c, nil) // discard
			}(conn)
		}
	}()
	log.Printf("testbedsrv http=%s bulk=%s obj=%s", httpAddr, bulkAddr, obj)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	return hs.Close()
}

var testbedSrvFn = runTestbedSrv
