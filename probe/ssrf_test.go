package probe

import (
	"context"
	"errors"
	"net"
	"testing"
)

// Garde SSRF — table des plages refusées par défaut (loopback, RFC1918,
// link-local, ULA, non spécifié, IPv4-mapped), plus le basculement banc.
func TestBlockedIP(t *testing.T) {
	blocked := map[string]string{
		"127.0.0.1":        "loopback",
		"127.1.2.3":        "loopback",
		"::1":              "loopback",
		"0.0.0.0":          "non spécifié",
		"::":               "non spécifié",
		"10.200.0.1":       "RFC1918",
		"10.30.0.2":        "RFC1918",
		"172.16.0.1":       "RFC1918",
		"172.31.255.254":   "RFC1918",
		"172.32.0.1":       "hors RFC1918",
		"192.168.174.1":    "RFC1918",
		"169.254.169.254":  "link-local",
		"fe80::1":          "fe80",
		"fe80::ffff:10.0.0.1": "fe80",
		"fc00::1":          "ULA",
		"fd12:3456::1":     "ULA",
		"::ffff:10.0.0.1":  "IPv4-mapped RFC1918",
		"::ffff:169.254.1.1": "IPv4-mapped link-local",
	}
	for s, want := range blocked {
		ip := net.ParseIP(s)
		if ip == nil {
			t.Fatalf("ParseIP(%q) échoué", s)
		}
		r, ok := blockedIP(ip)
		if want == "hors RFC1918" {
			if ok {
				t.Fatalf("%s : plage %q reconnue à tort", s, r)
			}
			continue
		}
		if !ok {
			t.Fatalf("%s : plage non reconnue (classe %q)", s, want)
		}
	}
	for _, s := range []string{"1.1.1.1", "8.8.8.8", "2606:4700:4700::1111", "172.15.0.1", "192.169.0.1"} {
		if ip := net.ParseIP(s); ip == nil {
			t.Fatalf("ParseIP(%q) échoué", s)
		} else if r, ok := blockedIP(ip); ok {
			t.Fatalf("%s : refusée (%q) alors qu'elle est publique", s, r)
		}
	}
}

func TestGuardHostLiteralPrivate(t *testing.T) {
	ctx := context.Background()
	for _, h := range []string{"10.30.0.2", "127.0.0.1", "169.254.169.254", "fd00::1", "192.168.0.1"} {
		err := GuardHost(ctx, h, false)
		var pte *PrivateTargetError
		if !errors.As(err, &pte) {
			t.Fatalf("%s : err = %v, want PrivateTargetError", h, err)
		}
	}
	if err := GuardHost(ctx, "1.1.1.1", false); err != nil {
		t.Fatalf("1.1.1.1 public refusé : %v", err)
	}
}

func TestGuardHostOptInFlips(t *testing.T) {
	t.Setenv("CGO_ALLOW_PRIVATE_TARGETS", "1")
	if !AllowPrivateTargets() {
		t.Fatal("opt-in non lu")
	}
	if err := GuardHost(context.Background(), "10.30.0.2", AllowPrivateTargets()); err != nil {
		t.Fatalf("banc refusé avec opt-in : %v", err)
	}
}

func TestGuardURLScheme(t *testing.T) {
	if err := GuardURL(context.Background(), "file:///etc/passwd", true); err == nil {
		t.Fatal("file:// accepté")
	}
	if err := GuardURL(context.Background(), "http://10.30.0.2:8081/small", false); err == nil {
		t.Fatal("privé sans opt-in accepté")
	}
	if err := GuardURL(context.Background(), "http://10.30.0.2:8081/small", true); err != nil {
		t.Fatalf("banc refusé avec opt-in : %v", err)
	}
	if err := GuardURL(context.Background(), "http://1.1.1.1/", false); err != nil {
		t.Fatalf("public refusé : %v", err)
	}
}

func TestGuardedDialPrivateRefused(t *testing.T) {
	// Aucune connexion réelle : la garde tranche AVANT le dial.
	for _, addr := range []string{"10.30.0.2:5201", "127.0.0.1:8081", "[fd00::1]:8081", "192.168.1.1:22"} {
		c, err := guardedDialContext(context.Background(), "tcp", addr, false, nil)
		if err == nil {
			_ = c.Close()
			t.Fatalf("%s : connexion établie malgré la garde", addr)
		}
		var pte *PrivateTargetError
		if !errors.As(err, &pte) {
			t.Fatalf("%s : err = %v, want PrivateTargetError", addr, err)
		}
	}
	// opt-in : 127.0.0.1 est accepté — on vérifie le refus seulement sur le
	// cas hors-dialogue (IP non joignable ≠ erreur de garde).
	if err := GuardHost(context.Background(), "10.30.0.2", true); err != nil {
		t.Fatalf("opt-in refusé : %v", err)
	}
}

func TestNewSafeTransportBlocksPrivateDial(t *testing.T) {
	tr := NewSafeTransport(false)
	if tr == nil || tr.DialContext == nil {
		t.Fatal("transport sans DialContext gardé")
	}
	c, err := tr.DialContext(context.Background(), "tcp", "10.30.0.2:5201")
	if err == nil {
		_ = c.Close()
		t.Fatal("dial privé accepté par le transport")
	}
	var pte *PrivateTargetError
	if !errors.As(err, &pte) {
		t.Fatalf("err = %v, want PrivateTargetError", err)
	}
}

func TestGuardedDialOptInLoopback(t *testing.T) {
	// Opt-in : le dial vers un loopback ouvert passe (127.0.0.1:port réel).
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		if c, err := ln.Accept(); err == nil {
			_ = c.Close()
		}
	}()
	c, err := guardedDialContext(context.Background(), "tcp", ln.Addr().String(), true, nil)
	if err != nil {
		t.Fatalf("dial opt-in refusé : %v", err)
	}
	_ = c.Close()
}
