package profile

import (
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/Realms4239/cgo/pkg/model"
)

// TestImportConcurrentSafe — imports, chargements et lectures simultanés :
// sans verrou, écriture et itération concurrentes de model.Profiles sont
// fatales au runtime Go. Chdir temp : Import persiste data/profiles.json.
func TestImportConcurrentSafe(t *testing.T) {
	dir := t.TempDir()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(old) }()

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			id := fmt.Sprintf("PX-%d", i)
			for j := 0; j < 25; j++ {
				if err := Import(model.Profile{ID: id, CapacityMbps: float64(j + 1)}); err != nil {
					t.Error(err)
					return
				}
				model.ProfilesMu.RLock()
				_ = model.Profiles[id]
				_ = len(model.Profiles)
				model.ProfilesMu.RUnlock()
				Load()
			}
		}(i)
	}
	wg.Wait()
	for i := 0; i < 8; i++ {
		if _, ok := model.Profiles[fmt.Sprintf("PX-%d", i)]; !ok {
			t.Fatalf("PX-%d lost after concurrent imports", i)
		}
	}
}
