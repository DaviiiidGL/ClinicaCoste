package hospital

import (
	"testing"
	"time"
)

func TestShiftIsActive(t *testing.T) {
	// Turno diurno: de 8:00 a 16:00
	shift := NewShift(8, 16)

	activeTime := time.Date(2026, 10, 19, 10, 30, 0, 0, time.UTC)
	if !shift.IsActive(activeTime) {
		t.Errorf("expected shift to be active at 10:30")
	}

	inactiveTime := time.Date(2026, 10, 19, 18, 0, 0, 0, time.UTC)
	if shift.IsActive(inactiveTime) {
		t.Errorf("expected shift NOT to be active at 18:00")
	}
}
