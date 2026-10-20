package metrics_test

import (
	"ClinicaCoste/internal/hospital/clock"
	"errors"
	"testing"

	"ClinicaCoste/internal/hospital"
)

// Estos tests viven en package metrics_test (caja negra): solo pueden usar la
// API pública de hospital, igual que cualquier consumidor real.

func newHospital(t *testing.T, capacities ...int) *hospital.Hospital {
	t.Helper()
	h, err := hospital.NewHospital("Hospital de los Costeños con Narcolepsia", clock.RealClock{})
	if err != nil {
		t.Fatalf("NewHospital: %v", err)
	}
	for i, c := range capacities {
		if _, err := h.AddRoom(101+i, c); err != nil {
			t.Fatalf("AddRoom: %v", err)
		}
	}
	return h
}

func admit(t *testing.T, h *hospital.Hospital, name string) *hospital.Patient {
	t.Helper()
	p := hospital.NewPatient(name, 30, hospital.NarcolepsyLevelMild)
	if err := h.AdmitPatient(p); err != nil {
		t.Fatalf("AdmitPatient: %v", err)
	}
	return p
}

func attack(t *testing.T, h *hospital.Hospital, p *hospital.Patient, location string) {
	t.Helper()
	// Quedarse sin sala es un resultado normal en estos escenarios.
	if _, err := h.ReportSleepAttack(p, location); err != nil && !errors.Is(err, hospital.ErrNoRoomAvailable) {
		t.Fatalf("ReportSleepAttack: %v", err)
	}
}
