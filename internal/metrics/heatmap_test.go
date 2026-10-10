package metrics_test

import (
	"testing"

	"ClinicaCoste/internal/metrics"
)

func TestLocationHeatmap_SortedByFrequencyThenName(t *testing.T) {
	h := newHospital(t, 1, 1, 1, 1, 1)
	for i, loc := range []string{"cafeteria", "radiology queue", "cafeteria", "  cafeteria ", "pharmacy"} {
		attack(t, h, admit(t, h, string(rune('A'+i))), loc)
	}

	got := metrics.LocationHeatmap(h)

	want := []metrics.LocationCount{
		{"cafeteria", 3}, // incluye la variante con espacios sobrantes
		{"pharmacy", 1},
		{"radiology queue", 1},
	}
	if len(got) != len(want) {
		t.Fatalf("want %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("position %d: want %v, got %v", i, want[i], got[i])
		}
	}
}

func TestLocationHeatmap_EmptyHistory(t *testing.T) {
	got := metrics.LocationHeatmap(newHospital(t))
	if got == nil || len(got) != 0 {
		t.Errorf("want empty non-nil slice, got %v", got)
	}
}
