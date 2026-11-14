package frontend

import "testing"

func TestEmbeddedAssetsPresent(t *testing.T) {
	if err := ValidateEmbeddedAssets(); err != nil {
		t.Fatalf("embedded SPA missing: %v", err)
	}
}
