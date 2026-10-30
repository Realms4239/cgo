package audit

import (
	"fmt"
	"sort"
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
func normalizeROS(v string) string {
	v = strings.Trim(strings.TrimSpace(v), `"`)
	v = strings.ToLower(v)
	v = strings.TrimSuffix(v, ".0")
	return v
}
