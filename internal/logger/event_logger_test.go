package logger

import (
	"testing"
	"time"

	"ClinicaCoste/internal/hospital"
	"ClinicaCoste/internal/hospital/clock"
)

func TestEventLogger(t *testing.T) {
	clk := clock.NewSimulatedClock(time.Now(), 1.0)
	h, err := hospital.NewHospital("Hospital Test", clk)
	if err != nil {
		t.Fatalf("NewHospital: %v", err)
	}

	// Usamos la nueva firma: pasamos el tamaño del buffer y recibimos el canal y la función cancel()
	ch, cancel := h.Subscribe(10)
	defer cancel() // Limpia la suscripción automáticamente al salir del test

	p := hospital.NewPatient("Test Patient", 25, hospital.NarcolepsyLevelMild)
	_ = h.AdmitPatient(p)

	// Reportamos un ataque para generar un episodio
	_, _ = h.ReportSleepAttack(p, "lobby")

	select {
	case episode := <-ch:
		if episode.Patient() == nil || episode.Patient().Name() != "Test Patient" {
			t.Errorf("expected episode for Test Patient, got %+v", episode)
		}
	case <-time.After(1 * time.Second):
		t.Error("timeout waiting for event log notification")
	}
}
