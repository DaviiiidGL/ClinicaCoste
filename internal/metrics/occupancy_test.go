package metrics_test

import (
	"testing"

	"ClinicaCoste/internal/metrics"
)

func TestRoomOccupancy(t *testing.T) {
	h := newHospital(t, 2, 2)
	for _, name := range []string{"A", "B", "C"} {
		attack(t, h, admit(t, h, name), "cafeteria")
	}

	got := metrics.RoomOccupancy(h)

	if got.Used != 3 || got.Capacity != 4 || got.Percent != 75 {
		t.Errorf("want 3/4 = 75%%, got %+v", got)
	}
}

func TestRoomOccupancy_NoRooms(t *testing.T) {
	h := newHospital(t)

	got := metrics.RoomOccupancy(h)

	if got.Capacity != 0 || got.Percent != 0 {
		t.Errorf("no rooms must give zeros (and no division by zero), got %+v", got)
	}
}

func TestAverageEpisodesPerPatient(t *testing.T) {
	h := newHospital(t, 1, 1, 1)
	a, b, c := admit(t, h, "A"), admit(t, h, "B"), admit(t, h, "C")
	admit(t, h, "D") // admitido, sin episodios
	attack(t, h, a, "x")
	attack(t, h, b, "y")
	attack(t, h, c, "z")

	if got := metrics.AverageEpisodesPerPatient(h); got != 0.75 {
		t.Errorf("3 episodes / 4 patients = 0.75, got %v", got)
	}
}

func TestAverageEpisodesPerPatient_NoPatients(t *testing.T) {
	if got := metrics.AverageEpisodesPerPatient(newHospital(t)); got != 0 {
		t.Errorf("no patients must give 0, got %v", got)
	}
}
