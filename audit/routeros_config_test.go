package audit

import (
	"os"
	"testing"
)

const scriptCake = `# CGO remediation - RouterOS v7+
/queue tree remove [find where comment="cgo-managed"]
/queue type remove [find where name~"^cgo-"]

/queue type add name=cgo-cake kind=cake \
    cake-bandwidth=19000000 cake-rtt=25ms \
    cake-flowmode=triple-isolate cake-nat=yes cake-diffserv=diffserv4 \
    cake-ack-filter=filter cake-wash=yes

/queue tree add name=cgo-wan-out parent=ether1-wan max-limit=19000000 \
    queue=cgo-cake comment="cgo-managed"
`

func TestParseScriptDialect(t *testing.T) {
	c := ParseRouterOSConfig(scriptCake)
	types := c.Get("queue type", "cgo-cake")
	if len(types) != 1 {
		t.Fatalf("queue type cgo-cake: %d objects, want 1", len(types))
	}
	if got := types[0].Props["kind"]; got != "cake" {
		t.Fatalf("kind = %q, want cake", got)
	}
	if got := types[0].Props["cake-diffserv"]; got != "diffserv4" {
		t.Fatalf("cake-diffserv = %q, want diffserv4", got)
	}
	trees := c.Get("queue tree", "cgo-wan-out")
	if len(trees) != 1 {
		t.Fatalf("queue tree cgo-wan-out: %d objects, want 1", len(trees))
	}
	if got := trees[0].Props["parent"]; got != "ether1-wan" {
		t.Fatalf("parent = %q, want ether1-wan", got)
	}
	if got := trees[0].Props["comment"]; got != "cgo-managed" {
		t.Fatalf("comment = %q, want cgo-managed (quotes stripped)", got)
	}
}

func TestParseExportDialect(t *testing.T) {
	b, err := os.ReadFile("../kit/chr/export-fixture.txt")
	if err != nil {
		t.Fatal(err)
	}
	c := ParseRouterOSConfig(string(b))
	if len(c.Get("queue type", "cgo-cake")) != 1 {
		t.Fatal("export: queue type cgo-cake not found")
	}
	trees := c.Get("queue tree", "cgo-wan-out")
	if len(trees) != 1 {
		t.Fatalf("export: queue tree cgo-wan-out: %d, want 1", len(trees))
	}
	if got := trees[0].Props["max-limit"]; got != "19000000" {
		t.Fatalf("max-limit = %q, want 19000000", got)
	}
}

func TestParseIgnoresRemovalsAndComments(t *testing.T) {
	c := ParseRouterOSConfig("/queue tree remove [find where comment=\"cgo-managed\"]\n# note\n")
	if len(c.Objects) != 0 {
		t.Fatalf("%d objects, want 0 (removal and comment lines are not objects)", len(c.Objects))
	}
}
