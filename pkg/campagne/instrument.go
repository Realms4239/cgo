//go:build instrumentation

package campagne

import "log"

// logTC trace les arguments tc vers kit/logs/tc.log (vérifiable en -race)
func logTC(args ...string) { log.Printf("instrument tc %v", args) }
