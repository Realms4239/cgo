package api

import (
	"regexp"
	"testing"
)

func TestNewAuditIDFormat(t *testing.T) {
	rx := regexp.MustCompile(`^audit-[0-9]{10}-[0-9a-f]{8}$`)
	seen := map[string]bool{}
	for i := 0; i < 100; i++ {
		id := NewAuditID()
		if !rx.MatchString(id) {
			t.Fatalf("NewAuditID() = %q, want audit-<unix>-<8hex>", id)
		}
		if seen[id] {
			t.Fatalf("duplicate ID %q", id)
		}
		seen[id] = true
	}
}
