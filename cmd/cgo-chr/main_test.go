package main

import (
	"strings"
	"testing"
	"time"
)

func TestBuildQEMUArgsTCG(t *testing.T) {
	args := strings.Join(BuildQEMUArgs(RunConfig{QEMU: "qemu-system-x86_64", Image: "chr.img", Level: "windows-tcg", Timeout: time.Minute}), " ")
	for _, want := range []string{"-accel tcg", "-serial stdio", "-display none", "file=chr.img,format=raw,if=virtio"} {
		if !strings.Contains(args, want) {
			t.Fatalf("args lack %q: %s", want, args)
		}
	}
	if strings.Contains(args, "-accel kvm") {
		t.Fatalf("TCG level asked for KVM: %s", args)
	}
}

func TestBuildQEMUArgsKVM(t *testing.T) {
	args := strings.Join(BuildQEMUArgs(RunConfig{QEMU: "qemu-system-x86_64", Image: "chr.img", Level: "linux-namespace", Timeout: time.Minute}), " ")
	if !strings.Contains(args, "-accel kvm") {
		t.Fatalf("linux-namespace did not ask for KVM: %s", args)
	}
}

func TestLevelAllowed(t *testing.T) {
	if ok, _ := levelAllowed("windows-tcg", "linux"); !ok {
		t.Fatal("windows-tcg refused on Linux")
	}
	if ok, _ := levelAllowed("windows-tcg", "windows"); !ok {
		t.Fatal("windows-tcg refused on Windows")
	}
	if ok, reason := levelAllowed("linux-namespace", "windows"); ok || reason == "" {
		t.Fatalf("linux-namespace accepted on Windows (reason %q)", reason)
	}
	if ok, _ := levelAllowed("linux-namespace", "linux"); !ok {
		t.Fatal("linux-namespace refused on Linux")
	}
}
