package bpf

import (
	"strings"
	"testing"
)

func TestStatusLadder(t *testing.T) {
	cases := []struct {
		name, goos string
		btf, caps  bool
		objectErr  error
		wantSource string
		wantReason string
	}{
		{"windows", "windows", true, true, nil, "tcpinfo", "hôte non-Linux"},
		{"no object", "linux", true, true, &BuildError{Reason: "objet non embarqué"}, "tcpinfo", "objet non embarqué"},
		{"no btf", "linux", false, true, nil, "tcpinfo", "BTF absent"},
		{"no caps", "linux", true, false, nil, "tcpinfo", "CAP_BPF absent"},
		{"ready", "linux", true, true, nil, "ebpf", ""},
	}
	for _, c := range cases {
		src, reason := Status(c.goos, c.btf, c.caps, c.objectErr)
		if src != c.wantSource || !strings.Contains(reason, c.wantReason) {
			t.Fatalf("%s: got (%q,%q), want (%q,%q)", c.name, src, reason, c.wantSource, c.wantReason)
		}
	}
}

func TestObjectAbsentByDefault(t *testing.T) {
	if _, err := ObjectBytes(); err == nil {
		t.Skip("built with -tags ebpf_object; absence cannot be tested here")
	} else if !strings.Contains(err.Error(), "objet") {
		t.Fatalf("absence error lost its reason: %v", err)
	}
	if _, err := ObjectSHA256(); err == nil {
		t.Fatal("SHA of an absent object returned nil error")
	}
}
