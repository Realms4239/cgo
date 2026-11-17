package main

import "testing"

func TestResolvedTarget(t *testing.T) {
	cases := []struct {
		name, flag, detected, want string
	}{
		{"dualns absent flag keeps bench target", "", "10.30.0.2", "10.30.0.2"},
		{"legacy absent flag keeps bench target", "", "10.200.0.1", "10.200.0.1"},
		{"explicit flag wins", "1.1.1.1", "10.30.0.2", "1.1.1.1"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := resolvedTarget(c.flag, c.detected); got != c.want {
				t.Fatalf("resolvedTarget(%q, %q) = %q, want %q", c.flag, c.detected, got, c.want)
			}
		})
	}
}
