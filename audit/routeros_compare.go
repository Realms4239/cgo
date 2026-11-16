package audit

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// RouterOSCheck is one field-level comparison shown in the proof artifact and
// the report. Expected/Found are always strings so a UI never has to guess.
type RouterOSCheck struct {
	Name     string `json:"name"`
	Expected string `json:"expected"`
	Found    string `json:"found"`
	Pass     bool   `json:"pass"`
}

// CompareRouterOS compares every object the report intended against what the
// device exported. Extra objects the device carries are ignored; duplicate
// copies of an intended name are a failure because the script is meant to be
// idempotent.
func CompareRouterOS(expected, found RouterOSConfig) []RouterOSCheck {
	var checks []RouterOSCheck
	for _, want := range expected.Objects {
		got := found.Get(want.Section, want.Name)
		if len(got) > 1 {
			checks = append(checks, RouterOSCheck{
				Name:     fmt.Sprintf("%s %s duplicate", want.Section, want.Name),
				Expected: "1 object",
				Found:    fmt.Sprintf("%d objects", len(got)),
				Pass:     false,
			})
		}
		if len(got) == 0 {
			checks = append(checks, RouterOSCheck{
				Name:     fmt.Sprintf("%s %s", want.Section, want.Name),
				Expected: fmt.Sprintf("%d properties", len(want.Props)),
				Found:    "absent",
				Pass:     false,
			})
			continue
		}
		keys := make([]string, 0, len(want.Props))
		for k := range want.Props {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			exp := want.Props[k]
			f, ok := got[0].Props[k]
			if !ok {
				f = "absent"
				// /export terse n'imprime pas les valeurs par défaut : une
				// propriété absente est correcte si elle vaut le défaut
				// usine RouterOS pour ce champ.
				if def, isDef := rosDefaults[k]; isDef && normalizeROS(def) == normalizeROS(exp) {
					f = "absent (défaut)"
					checks = append(checks, RouterOSCheck{
						Name:     fmt.Sprintf("%s %s.%s", want.Section, want.Name, k),
						Expected: exp,
						Found:    f,
						Pass:     true,
					})
					continue
				}
			}
			checks = append(checks, RouterOSCheck{
				Name:     fmt.Sprintf("%s %s.%s", want.Section, want.Name, k),
				Expected: exp,
				Found:    f,
				Pass:     ok && normalizeROS(exp) == normalizeROS(f),
			})
		}
	}
	sort.Slice(checks, func(i, j int) bool { return checks[i].Name < checks[j].Name })
	return checks
}

// normalizeROS absorbs RouterOS formatting noise only: quotes, boolean case
// and a trailing ".0" on integers. It never equates two different values.
// rosDefaults — défauts usine RouterOS non imprimés par /export terse.
var rosDefaults = map[string]string{
	"cake-flowmode": "triple-isolate",
}

var rateRe = regexp.MustCompile(`^([0-9]+(?:\.[0-9]+)?)\s*(k|m|g|t|p)?(bps|kbps|mbps|gbps|kbit|mbit|gbit)?$`)

// normalizeROS comparaison sémantique : les débits s'égalent, quel que soit
// le format d'affichage (« 19000000 » ≡ « 19.0Mbps » ≡ « 19M ») ; les autres
// chaînes se comparent textuellement.
func normalizeROS(v string) string {
	v = strings.Trim(strings.TrimSpace(v), `"`)
	v = strings.ToLower(v)
	v = strings.TrimSuffix(v, ".0")
	if m := rateRe.FindStringSubmatch(v); m != nil {
		x, _ := strconv.ParseFloat(m[1], 64)
		switch m[2] {
		case "k":
			x *= 1e3
		case "m":
			x *= 1e6
		case "g":
			x *= 1e9
		case "t":
			x *= 1e12
		case "p":
			x *= 1e15
		}
		return strconv.FormatFloat(x, 'f', -1, 64)
	}
	return v
}
