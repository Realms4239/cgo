package results

import (
	"path/filepath"
	"testing"
)

// BenchmarkScanFrozenCorpus — le coût réel d'un full-scan uncaché du corpus
// gelé (~135–190 ms sur 4 cœurs) : /api/results met le résultat en cache
// (Scan memoïsée), benchmark mesure la voie uncached.
func BenchmarkScanFrozenCorpus(b *testing.B) {
	dir := b.TempDir()
	if _, _, err := SeedRunsFor(dir); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		files, _ := filepath.Glob(filepath.Join(dir, "*", "aqm_eval.csv"))
		if _, err := scanUncached(dir, "", files); err != nil {
			b.Fatal(err)
		}
	}
}
