package hospital

import (
	"testing"
	"time"

	"ClinicaCoste/internal/hospital/clock"
)

func registerAt(t *testing.T, h *Hospital, p *Patient, at time.Time) {
	t.Helper()
	e := EpisodeRecord{id: "e-" + at.Format(time.RFC3339Nano), timestamp: at, patient: p, location: "x"}
	if err := h.RegisterEpisode(e); err != nil {
		t.Fatalf("RegisterEpisode: %v", err)
	}
}

func TestSevereReport(t *testing.T) {
	noon := time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)
	clk := clock.NewSimulatedClock(noon, 1.0)
	h, _ := NewHospital("Hospital Test", clk)

	severeBusy := admit(t, h, "Yeimy", NarcolepsyLevelSevere)
	severeQuiet := admit(t, h, "Wilfrido", NarcolepsyLevelSevere)
	mild := admit(t, h, "Mild", NarcolepsyLevelMild)
	moderate := admit(t, h, "Moderate", NarcolepsyLevelModerate)

	registerAt(t, h, severeBusy, noon.Add(-1*time.Hour))
	registerAt(t, h, severeBusy, noon.Add(-3*time.Hour))
	registerAt(t, h, severeBusy, noon.Add(-30*time.Hour))
	registerAt(t, h, mild, noon.Add(-1*time.Hour))
	registerAt(t, h, moderate, noon.Add(-2*time.Hour))

	report := h.SevereReport()

	if len(report) != 2 {
		t.Fatalf("only the 2 Severe patients must appear, got %d: %v", len(report), report)
	}
	if report[severeBusy] != 2 {
		t.Errorf("Yeimy: want 2 episodes today, got %d", report[severeBusy])
	}
	if n, ok := report[severeQuiet]; !ok || n != 0 {
		t.Errorf("a Severe patient without episodes must appear with 0, got (%d, %v)", n, ok)
	}
	if _, ok := report[mild]; ok {
		t.Error("Mild patients must not be in the report")
	}
}

func TestSevereReport_NoSeverePatientsReturnsEmptyNonNilMap(t *testing.T) {
	h, _ := NewHospital("Hospital Test", clock.RealClock{})
	admit(t, h, "Mild", NarcolepsyLevelMild)

	report := h.SevereReport()

	if report == nil {
		t.Fatal("must return an empty map, not nil")
	}
	if len(report) != 0 {
		t.Errorf("want empty, got %v", report)
	}
}

func TestSevereReport_OnlyCountsTheDayOfTheHospitalClock(t *testing.T) {
	day1 := time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)
	day2 := day1.Add(24 * time.Hour)

	// Hospital para el día 1
	hDay1, _ := NewHospital("H1", clock.NewSimulatedClock(day1, 1.0))
	p1 := admit(t, hDay1, "Yeimy", NarcolepsyLevelSevere)
	registerAt(t, hDay1, p1, day1)
	registerAt(t, hDay1, p1, day2) // este es de day2, no debe contar en day1

	if got := hDay1.SevereReport()[p1]; got != 1 {
		t.Errorf("day 1: want 1, got %d", got)
	}

	// Hospital para el día 2
	hDay2, _ := NewHospital("H2", clock.NewSimulatedClock(day2, 1.0))
	p2 := admit(t, hDay2, "Yeimy", NarcolepsyLevelSevere)
	registerAt(t, hDay2, p2, day1)
	registerAt(t, hDay2, p2, day2)
	registerAt(t, hDay2, p2, day2.Add(time.Hour))

	if got := hDay2.SevereReport()[p2]; got != 2 {
		t.Errorf("day 2: want 2, got %d", got)
	}
}

func TestSevereReport_DayBoundary(t *testing.T) {
	day10 := time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)
	clk := clock.NewSimulatedClock(day10, 1.0)
	h, _ := NewHospital("Hospital Test", clk)
	p := admit(t, h, "Yeimy", NarcolepsyLevelSevere)

	registerAt(t, h, p, time.Date(2026, 10, 10, 0, 0, 0, 0, time.UTC))
	registerAt(t, h, p, time.Date(2026, 10, 10, 23, 59, 59, 0, time.UTC))
	registerAt(t, h, p, time.Date(2026, 10, 9, 23, 59, 59, 0, time.UTC))
	registerAt(t, h, p, time.Date(2026, 10, 11, 0, 0, 0, 0, time.UTC))

	if got := h.SevereReport()[p]; got != 2 {
		t.Errorf("want 2 (both edges of today), got %d", got)
	}
}

func TestSevereReportOrdered(t *testing.T) {
	noon := time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)
	clk := clock.NewSimulatedClock(noon, 1.0)
	h, _ := NewHospital("Hospital Test", clk)

	zoe := admit(t, h, "Zoe", NarcolepsyLevelSevere)
	ana := admit(t, h, "Ana", NarcolepsyLevelSevere)
	top := admit(t, h, "Top", NarcolepsyLevelSevere)

	registerAt(t, h, top, noon.Add(-1*time.Hour))
	registerAt(t, h, top, noon.Add(-2*time.Hour))
	registerAt(t, h, zoe, noon.Add(-1*time.Hour))
	registerAt(t, h, ana, noon.Add(-1*time.Hour))

	got := h.SevereReportOrdered()

	want := []*Patient{top, ana, zoe}
	if len(got) != 3 {
		t.Fatalf("want 3 entries, got %d", len(got))
	}
	for i, p := range want {
		if got[i].Patient != p {
			t.Errorf("position %d: got %s, want %s", i, got[i].Patient.Name(), p.Name())
		}
	}
	if got[0].Episodes != 2 || got[1].Episodes != 1 || got[2].Episodes != 1 {
		t.Errorf("unexpected counts: %+v", got)
	}
}

func TestSameDay_UsesTheZoneOfTheReference(t *testing.T) {
	bogota := time.FixedZone("COT", -5*3600)
	ref := time.Date(2026, 10, 10, 20, 0, 0, 0, bogota)
	sameLocalDay := time.Date(2026, 10, 10, 8, 0, 0, 0, bogota)
	if !sameDay(sameLocalDay, ref) {
		t.Error("same calendar day in the reference zone must match")
	}
	if !sameDay(ref.UTC(), ref) {
		t.Error("the same instant must always be the same day")
	}
}
