package results

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadAQMDetailedCountsExtraFieldRows(t *testing.T) {
	p := filepath.Join(t.TempDir(), "aqm_eval.csv")
	body := "a,b\n1,2\n3,4,5\n6\n"
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	header, rows, malformed, err := ReadAQMDetailed(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(header) != 2 || len(rows) != 3 {
		t.Fatalf("header=%v rows=%v, want 2 columns and 3 rows", header, rows)
	}
	if malformed != 1 {
		t.Fatalf("malformed = %d, want 1 (the row with an extra field)", malformed)
	}
	h2, r2, err := ReadAQM(p)
	if err != nil || len(h2) != 2 || len(r2) != 3 {
		t.Fatalf("ReadAQM compatibility broken: %v %v %v", h2, r2, err)
	}
}
