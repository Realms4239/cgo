//go:build linux

package campagne

import (
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

// readCPUUsage échantillonne l'occupation CPU sur 200 ms via /proc/stat.
// Retourne NaN si la mesure est impossible : G7 reste alors non évaluée,
// jamais un faux passage à 0 %.
func readCPUUsage() float64 {
	idle0, total0, ok := readProcCPU()
	if !ok {
		return math.NaN()
	}
	time.Sleep(200 * time.Millisecond)
	idle1, total1, ok := readProcCPU()
	if !ok || total1 <= total0 {
		return math.NaN()
	}
	dt := total1 - total0
	di := idle1 - idle0
	if di > dt {
		di = dt
	}
	usage := 100 * float64(dt-di) / float64(dt)
	if usage < 0 {
		usage = 0
	}
	if usage > 100 {
		usage = 100
	}
	return usage
}

func readProcCPU() (idle, total uint64, ok bool) {
	b, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0, 0, false
	}
	line, _, _ := strings.Cut(string(b), "\n")
	return parseProcStatLine(line)
}

// parseProcStatLine lit la ligne agrégée "cpu  user nice system idle iowait ...".
// idle inclut iowait ; le total s'arrête à steal (guest est déjà compris dans
// user/nice et ne doit pas être compté deux fois).
func parseProcStatLine(line string) (idle, total uint64, ok bool) {
	f := strings.Fields(line)
	if len(f) < 9 || f[0] != "cpu" {
		return 0, 0, false
	}
	var v [8]uint64
	for i := 0; i < 8; i++ {
		n, err := strconv.ParseUint(f[i+1], 10, 64)
		if err != nil {
			return 0, 0, false
		}
		v[i] = n
		total += n
	}
	return v[3] + v[4], total, true
}
