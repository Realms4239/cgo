package main

import (
	"context"
	"fmt"
	"io"
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

// runTestbedSrv sert le petit objet en HTTP et le puits de masse,
// tous deux liés dans le netns cgo-srv (kit/testbed.sh).
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
	hs := &http.Server{Addr: httpAddr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
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
				// protocole du banc : "D" ou "D:<cc>" = download (le serveur
				// INONDE avec la CC de la cellule) ; sinon puits historique
				// (le client inonde, le serveur discard). Un seul port, deux
				// sens, CC étiquetée. Lecture bornée : un client historique
				// n'envoie qu'un octet, on ne l'attend pas.
				c.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
				hdr := make([]byte, 24)
				nr, _ := io.ReadAtLeast(c, hdr, 1)
				c.SetReadDeadline(time.Time{})
				if dl, cc := probe.ParseDownloadHello(hdr[:nr]); dl {
					if err := probe.SetConnCC(c, cc); err != nil {
						log.Printf("[src] CC %q refusée: %v (défaut hôte)", cc, err)
					}
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
					defer cancel()
					// observabilité de fortune : ce que la source ENVOIE
					n, _ := probe.BulkSource(ctx, c) // mode source : download
					log.Printf("[src] download cc=%s: %d octets envoyés", cc, n)
					return
				}
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
				defer cancel()
				_, _ = probe.BulkReceive(ctx, c, nil) // puits : upload
			}(conn)
		}
	}()
	log.Printf("testbedsrv http=%s bulk=%s obj=%s", httpAddr, bulkAddr, obj)
	go func() {
		if err := hs.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("http: %v", err)
		}
	}()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	ln.Close()
	return hs.Close()
}

var testbedSrvFn = runTestbedSrv
