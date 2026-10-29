//go:build linux

package campagne

import (
	"os"
	"strings"

	"github.com/Realms4239/cgo/pkg/model"
)

func readHostCCList() []string {
	raw, err := os.ReadFile("/proc/sys/net/ipv4/tcp_available_congestion_control")
	if err != nil {
		return nil
	}
	return model.ParseCCList(string(raw))
}

func readHostKernelRelease() string {
	raw, err := os.ReadFile("/proc/sys/kernel/osrelease")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(raw))
}
