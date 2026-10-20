package clock

import (
	"testing"
	"time"
)

func TestSimulatedClock(t *testing.T) {
	start := time.Date(2026, 10, 19, 8, 0, 0, 0, time.UTC)
	// Multiplicador 2.0: 1 hora real avanza 2 horas virtuales
	clk := NewSimulatedClock(start, 2.0)

	if !clk.Now().Equal(start) {
		t.Errorf("expected start time %v, got %v", start, clk.Now())
	}

	clk.Advance(1 * time.Hour)
	expected := start.Add(2 * time.Hour)
	if !clk.Now().Equal(expected) {
		t.Errorf("expected time after advance %v, got %v", expected, clk.Now())
	}
}

func TestRealClock(t *testing.T) {
	rc := RealClock{}
	now := rc.Now()
	if time.Since(now) > 1*time.Second {
		t.Errorf("RealClock.Now() returned an outdated time: %v", now)
	}
}
