package kit

import (
	"strings"
	"testing"
)

// TestScheduleUnit — le timer systemd nightly qui rejoue la campagne
// filtrée : l'unité générée doit être un timer + service valides, la
// commande porte les filtres et la direction, et l'heure est configurable.
func TestScheduleUnit(t *testing.T) {
	unit, timer := ScheduleUnits("02:30", "P2", "cake,fq_codel", "bbr", "both", 1)
	if !strings.Contains(timer, "OnCalendar=*-*-* 02:30:00") {
		t.Fatalf("timer sans OnCalendar 02:30: %s", timer)
	}
	if !strings.Contains(timer, "Persistent=true") {
		t.Fatal("timer sans Persistent : une VM éteinte la nuit doit rattraper au boot")
	}
	if !strings.Contains(unit, "cgo-linux run --profiles P2") {
		t.Fatalf("service sans la commande campagne: %s", unit)
	}
	if !strings.Contains(unit, "--qdiscs cake,fq_codel") || !strings.Contains(unit, "--cc bbr") {
		t.Fatalf("service sans les filtres sous-matrice: %s", unit)
	}
	if !strings.Contains(unit, "--direction both") {
		t.Fatal("service sans la direction")
	}
	if !strings.Contains(unit, "[Unit]") || !strings.Contains(timer, "[Timer]") {
		t.Fatal("unités mal formées")
	}
}

// TestScheduleUnitDefaults — sans filtres : la matrice pleine, direction up.
func TestScheduleUnitDefaults(t *testing.T) {
	unit, _ := ScheduleUnits("03:00", "P2", "", "", "", 0)
	if strings.Contains(unit, "--qdiscs") || strings.Contains(unit, "--direction") {
		t.Fatalf("les défauts ne doivent pas passer de flags vides: %s", unit)
	}
	if !strings.Contains(unit, "--reps 3") {
		t.Fatalf("reps=0 doit devenir --reps 3 (l'API refuse 0): %s", unit)
	}
}

// TestScheduleCronLine — repli cron sans privilège : heure, commande
// complète, répertoire de gels, log.
func TestScheduleCronLine(t *testing.T) {
	line := ScheduleCronLine("02:30", "P2", "cake", "bbr", "both", 1)
	if !strings.HasPrefix(line, "30 02 * * * ") {
		t.Fatalf("champ horaire: %s", line)
	}
	for _, want := range []string{"--profiles P2", "--qdiscs cake", "--cc bbr", "--direction both", "--reps 1", "cd /home/altfloat/cgo", "campaign-cron.log"} {
		if !strings.Contains(line, want) {
			t.Fatalf("ligne cron sans %q: %s", want, line)
		}
	}
}
