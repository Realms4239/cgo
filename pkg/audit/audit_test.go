package audit

import (
	"context"
	"testing"
)

func TestAuditRunHappyPath(t *testing.T) {
	p := Params{AuditID: "a1", Site: "Dept X", LinkType: "5g", Provider: "Yas", Duration: 1, Target: "127.0.0.1"}
	deps := Deps{
		Ping: func(ctx context.Context, _ string, n int) []float64 { return []float64{20, 21, 22} },
		Small: func(ctx context.Context) (float64, error) { return 25, nil },
	}
	res, err := Run(context.Background(), p, deps)
	if err != nil {
		t.Fatal(err)
	}
	if res.RTTIdleP50 == 0 || res.HTTPSmallP95 == 0 {
		t.Fatalf("empty result %+v", res)
	}
}

func TestAppendLinkAudit(t *testing.T) {
	dir := t.TempDir()
	r := &Result{AuditID: "a1", Site: "Dept X", LinkType: "5g", Provider: "Yas", RTTIdleP50: 20, RTTIdleP95: 25, HTTPSmallP95: 30}
	if err := AppendLinkAudit(dir, r); err != nil {
		t.Fatal(err)
	}
	if err := AppendLinkAudit(dir, r); err != nil {
		t.Fatal(err)
	}
	// file should have header +2 rows
}
