package campagne

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// plan mort : small ET bulk fermés → refus explicite, pas de matrice à vide.
func TestPlaneReadyRefused(t *testing.T) {
	d := Deps{SmallURL: "http://127.0.0.1:18081/small", BulkAddr: "127.0.0.1:15201"}
	if err := PlaneReady(context.Background(), d); err == nil {
		t.Fatal("plan mort accepté — une matrice tournerait à vide")
	}
}

// plan vivant : httptest small + listener bulk → accepté.
func TestPlaneReadyOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(make([]byte, 16384))
	}))
	defer srv.Close()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			c.Close()
		}
	}()
	d := Deps{SmallURL: srv.URL + "/small", BulkAddr: ln.Addr().String()}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := PlaneReady(ctx, d); err != nil {
		t.Fatalf("plan vivant refusé : %v", err)
	}
	fmt.Println("plane ok")
}
