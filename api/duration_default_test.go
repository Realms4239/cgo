package api

import "testing"

func TestAuditDurationDefaultIs60(t *testing.T) {
	for _, s := range schemaParams {
		if s.Key == "audit_duration_s" {
			if s.Default != 60 {
				t.Fatalf("audit_duration_s default = %v, want 60", s.Default)
			}
			return
		}
	}
	t.Fatal("audit_duration_s setting missing")
}
