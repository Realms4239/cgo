package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestReportExportSh — le script tc déployable : pour chaque profil, la
// meilleure cellule gelée devient une recette tc complète et rejouable
// (reset, netem aux conditions du profil, shaper), pas une ligne isolée.
func TestReportExportSh(t *testing.T) {
	dir := chdirTemp(t)
	writeCSVRun(t, dir, "run-sh", header18,
		"run-sh,1,P2,pfifo_fast,bbr,1,111,168,57,328.2,60.2,18.2,443,5,641464,66.40,0.0,valid",
		"run-sh,2,P2,cake,bbr,1,100,114,14,209.2,98.1,18.1,81,0,117288,12.10,0.0,valid",
	)
	h := New(Deps{})
	srv := httptest.NewServer(h)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/report/export?format=sh")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if ct := resp.Header.Get("Content-Type"); ct != "text/x-shellscript" {
		t.Fatalf("content-type = %q", ct)
	}
	if cd := resp.Header.Get("Content-Disposition"); !strings.Contains(cd, "attachment") || !strings.Contains(cd, "aqm-recipe.sh") {
		t.Fatalf("disposition = %q", cd)
	}
	b, _ := io.ReadAll(resp.Body)
	s := string(b)
	// shebang + reset propre + conditions netem du profil + shaper best + hash de provenance
	for _, want := range []string{
		"#!/bin/sh", "qdisc del", "netem", "delay 100ms 15ms", "loss 0.5%",
		"qdisc replace", "cake bandwidth 16mbit", "provenance",
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("script manque %q dans:\n%s", want, s)
		}
	}
	if strings.Contains(s, "sudo") {
		t.Fatal("le script ne doit pas préfixer sudo — c'est à l'opérateur de le lancer en root")
	}
}
