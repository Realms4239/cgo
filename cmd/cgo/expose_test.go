package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"
)

// scrubEnv — l'environnement d'un sous-processus de test ne transporte jamais
// de secret (pattern SAFE_INVOCATIONS de tools/verify.py).
var credRe = regexp.MustCompile(`(?i)key|token|secret|password|credential`)

func scrubEnv() []string {
	var env []string
	for _, kv := range os.Environ() {
		k := kv[:strings.IndexByte(kv, '=')]
		if credRe.MatchString(k) {
			continue
		}
		env = append(env, kv)
	}
	return env
}

func buildCgoBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "cgo-exposure-test")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Env = scrubEnv()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build : %v\n%s", err, out)
	}
	return bin
}

// execServe — lance `cgo --serve …` avec un environnement expurgé de tout
// secret ; tue le processus au bout de 15 s (un refus quitte en < 1 s, un
// serveur qui démarre est tué et l'échec est noté).
func execServe(t *testing.T, bin string, args ...string) (int, string, string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Env = scrubEnv()
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &outBuf, &errBuf
	err := cmd.Run()
	code := 0
	if err == nil {
		code = 0
	} else if exitErr, ok := err.(*exec.ExitError); ok {
		code = exitErr.ExitCode()
	} else {
		t.Fatalf("exec %v : %v (stdout=%q stderr=%q)", args, err, outBuf.String(), errBuf.String())
	}
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("le processus n'a pas quitté en 15 s (était-il en train de servir ?) : %v", args)
	}
	return code, outBuf.String(), errBuf.String()
}

func tokenFileAbsent(t *testing.T) {
	t.Helper()
	if _, err := os.Stat("/etc/cgo/auth.token"); err == nil {
		t.Skip("hôte avec /etc/cgo/auth.token — le refus exige l'absence de token fichier")
	}
}

// TestServeRefusesNonLoopbackPlaintextWithoutExpose — acceptance 3 :
// --addr 0.0.0.0:9090 --tls=false sans --expose ni token → exit 2, stderr,
// aucune écoute.
func TestServeRefusesNonLoopbackPlaintextWithoutExpose(t *testing.T) {
	tokenFileAbsent(t)
	bin := buildCgoBinary(t)
	code, stdout, stderr := execServe(t, bin, "--serve", "--addr", "0.0.0.0:9090", "--tls=false")
	if code != 2 {
		t.Fatalf("exit=%d, want 2 (stdout=%q)", code, stdout)
	}
	if strings.TrimSpace(stderr) == "" {
		t.Fatal("aucun message sur stderr au refus")
	}
	if !strings.Contains(stderr, "--expose") {
		t.Fatalf("stderr ne nomme pas --expose : %q", stderr)
	}
}

// TestServeRefusesNonLoopbackExposeWithoutToken — --expose posé mais aucun
// token (ni env expurgé, ni fichier) : exit 2 même en mode explicite.
func TestServeRefusesNonLoopbackExposeWithoutToken(t *testing.T) {
	tokenFileAbsent(t)
	bin := buildCgoBinary(t)
	code, _, stderr := execServe(t, bin, "--serve", "--addr", "0.0.0.0:9090", "--tls=false", "--expose")
	if code != 2 {
		t.Fatalf("exit=%d, want 2 (stderr=%q)", code, stderr)
	}
	if strings.TrimSpace(stderr) == "" {
		t.Fatal("aucun message sur stderr au refus")
	}
}

// TestServeRefusesNonLoopbackTLSWithoutExpose — TLS n'excuse pas l'absence
// de token : un bind non-loopback sans --expose est refusé même en HTTPS.
func TestServeRefusesNonLoopbackTLSWithoutExpose(t *testing.T) {
	bin := buildCgoBinary(t)
	code, _, stderr := execServe(t, bin, "--serve", "--addr", "192.0.2.10:9090", "--tls=true")
	if code != 2 {
		t.Fatalf("exit=%d, want 2 (stderr=%q)", code, stderr)
	}
}

// TestExposeGuardAllowsLoopback — loopback toujours permis, TLS ou non,
// --expose ou non.
func TestExposeGuardAllowsLoopback(t *testing.T) {
	for _, addr := range []string{"127.0.0.1:9090", "127.1.2.3:9090", "[::1]:9090", "localhost:9090"} {
		if err := checkExposure(addr, false, func(string) string { return "" }, filepath.Join(t.TempDir(), "absent")); err != nil {
			t.Fatalf("loopback %s refusé : %v", addr, err)
		}
	}
}

// TestExposeGuardAllowsNonLoopbackWithExposeAndToken — le passe-droit
// opérateur : --expose + CGO_API_TOKEN → autorisé (TLS optionnel).
func TestExposeGuardAllowsNonLoopbackWithExposeAndToken(t *testing.T) {
	token := func(k string) string {
		if k == "CGO_API_TOKEN" {
			return "expose-token"
		}
		return ""
	}
	if err := checkExposure("0.0.0.0:9090", true, token, filepath.Join(t.TempDir(), "absent")); err != nil {
		t.Fatalf("--expose + token refusé : %v", err)
	}
}

// TestExposeGuardAllowsNonLoopbackWithTokenFile — le jeton fichier
// /etc/cgo/auth.token suffit à lever le refus (chemin nominal du unit).
func TestExposeGuardAllowsNonLoopbackWithTokenFile(t *testing.T) {
	tokenFile := filepath.Join(t.TempDir(), "auth.token")
	if err := os.WriteFile(tokenFile, []byte("CGO_API_TOKEN=deadbeef\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	empty := func(string) string { return "" }
	if err := checkExposure("0.0.0.0:9090", true, empty, tokenFile); err != nil {
		t.Fatalf("--expose + token fichier refusé : %v", err)
	}
}

// TestExposeGuardRefusesNonLoopbackWithExposeAndNoTokenAnywhere — refus
// sec : --expose sans env ni fichier exploitable.
func TestExposeGuardRefusesNonLoopbackWithExposeAndNoTokenAnywhere(t *testing.T) {
	empty := func(string) string { return "" }
	err := checkExposure("0.0.0.0:9090", true, empty, filepath.Join(t.TempDir(), "absent"))
	if err == nil {
		t.Fatal("--expose sans token accepté")
	}
}

// TestExposeGuardRefusesUnparseableAddr — une adresse illisible échoue
// fermé, jamais ouverte.
func TestExposeGuardRefusesUnparseableAddr(t *testing.T) {
	empty := func(string) string { return "" }
	if err := checkExposure("pas-une-adresse", true, empty, ""); err == nil {
		t.Fatal("adresse illisible acceptée")
	}
}
