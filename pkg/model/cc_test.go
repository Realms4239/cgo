package model

import (
	"reflect"
	"testing"
)

func TestParseCCList(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"reno cubic bbr\n", []string{"reno", "cubic", "bbr"}},
		{"  bbr3   bbr bbr3 ", []string{"bbr3", "bbr"}},
		{"", nil},
	}
	for _, c := range cases {
		if got := ParseCCList(c.in); !reflect.DeepEqual(got, c.want) {
			t.Fatalf("ParseCCList(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestBBRv3IsKnownButNotDefault(t *testing.T) {
	known := false
	for _, c := range KnownCC {
		if c == BBRv3 {
			known = true
		}
	}
	if !known {
		t.Fatal("KnownCC sans bbr3")
	}
	for _, c := range AllCC {
		if c == BBRv3 {
			t.Fatal("AllCC ne doit pas inclure bbr3 par défaut")
		}
	}
}
