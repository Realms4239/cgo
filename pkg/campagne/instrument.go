//go:build instrumentation

package campagne

import "log"

// logTC is the surgical Go tag seam — tc args trace for kit/logs/tc.log and -race verification
func logTC(args ...string) { log.Printf("instrument tc %v", args) }
