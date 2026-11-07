package api

import "testing"

func TestAuditStartDirectionParam(t *testing.T) {
	// La requête doit accepter direction/flows et les passer à l'audit.
	// (Test au niveau Params : le handler ne peut pas lancer d'audit réel ici.)
	p := auditParamsFromRequest(map[string]any{
		"site": "S", "link_type": "4g", "duration": 30,
		"direction": "down", "flows": 8,
	})
	if p.Direction != "down" || p.Flows != 8 {
		t.Fatalf("params = %+v", p)
	}
	q := auditParamsFromRequest(map[string]any{"site": "S"})
	if q.Direction != "up" || q.Flows != 4 {
		t.Fatalf("défauts up/4 : %+v", q)
	}
	// Les champs de terrain passent tels quels — jamais réécrits en silence.
	r := auditParamsFromRequest(map[string]any{
		"site": "S", "target": "10.30.0.1", "small_url": "http://x/y", "bulk_addr": "10.30.0.1:5201",
	})
	if r.Target != "10.30.0.1" || r.SmallURL != "http://x/y" || r.BulkAddr != "10.30.0.1:5201" {
		t.Fatalf("overrides terrain : %+v", r)
	}
}
