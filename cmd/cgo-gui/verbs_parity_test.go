package main

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/internal/kit"
)

// Les deux exécutants (Win32 actions.go, Fyne fygui_linux.go) sont de JAMAIS
// compilés ensemble — c'est tout l'enjeu. Ce test lit leurs SOURCES et exige
// que la même liste de verbes y apparaisse, et que cette liste soit EXACTEMENT
// internal/kit.SuiteVerbs. Un verbe ajouté dans un seul exécutant fait
// échouer ce test : la Suite et les boutons ne peuvent plus diverger.
//
// Les fichiers restent dans le dépôt — la lecture de source est voulue.

var verbCase = regexp.MustCompile(`case kind == "([^"]+)"`)

func readVerbs(path string) map[string]int {
	b, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	counts := map[string]int{}
	for _, m := range verbCase.FindAllStringSubmatch(string(b), -1) {
		counts[m[1]]++
	}
	return counts
}

func TestExecutorsMatchSuiteVerbs(t *testing.T) {
	win := readVerbs("actions.go")
	fyn := readVerbs("fygui_linux.go")

	known := map[string]bool{}
	for _, v := range kit.SuiteVerbs {
		known[v] = true
	}

	for name, table := range map[string]map[string]int{"Win32": win, "Fyne": fyn} {
		for v := range table {
			if !known[v] {
				t.Errorf("%s exécute %q mais SuiteVerbs ne le connaît pas — bouton seulement", name, v)
			}
		}
		for _, v := range kit.SuiteVerbs {
			if table[v] == 0 {
				t.Errorf("SuiteVerbs connaît %q mais %s ne l'exécute pas — Suite envoie dans le vide", v, name)
			}
		}
	}

	// Les deux exécutants portent exactement le même jeu de verbes
	for v := range win {
		if fyn[v] != win[v] {
			t.Errorf("verbe %q : %d cas Win32 vs %d cas Fyne", v, win[v], fyn[v])
		}
	}
}

func TestExecutorsIdenticalVerbSets(t *testing.T) {
	win, fyn := readVerbs("actions.go"), readVerbs("fygui_linux.go")
	wk, fk := make([]string, 0, len(win)), make([]string, 0, len(fyn))
	for v := range win {
		wk = append(wk, v)
	}
	for v := range fyn {
		fk = append(fk, v)
	}
	sort.Strings(wk)
	sort.Strings(fk)
	if strings.Join(wk, ",") != strings.Join(fk, ",") {
		t.Fatalf("les exécutants divergent : Win32 %v / Fyne %v", wk, fk)
	}
}
