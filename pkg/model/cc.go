package model

import "strings"

// ParseCCList tokenise la liste du noyau
// (/proc/sys/net/ipv4/tcp_available_congestion_control).
// L'ordre du noyau est conservé ; vides et doublons sont écartés.
func ParseCCList(raw string) []string {
	seen := map[string]bool{}
	var out []string
	for _, f := range strings.Fields(raw) {
		if seen[f] {
			continue
		}
		seen[f] = true
		out = append(out, f)
	}
	return out
}
