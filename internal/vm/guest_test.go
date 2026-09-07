package vm

import (
	"reflect"
	"testing"
)

func TestVmwareGuestArgs(t *testing.T) {
	got := vmwareGuestArgs("u", "p", "D:/x.vmx", "runProgramInGuest", "/bin/echo", []string{"hi"})
	want := []string{"-T", "ws", "-gu", "u", "-gp", "p", "runProgramInGuest", "D:/x.vmx", "/bin/echo", "hi"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestVboxGuestArgs(t *testing.T) {
	got := vboxGuestArgs("u", "p", "ubu", "/bin/true", nil)
	want := []string{"guestcontrol", "ubu", "run", "--username", "u", "--password", "p", "--wait-stdout", "--wait-stderr", "--exe", "/bin/true", "--"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q want %q", got, want)
	}
}
