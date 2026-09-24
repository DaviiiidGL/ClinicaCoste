package hospital_test

import (
	"fmt"
	"testing"

	"ClinicaCoste/internal/hospital"
)

func TestStates_String(t *testing.T) {
	tests := []struct {
		name     string
		input    fmt.Stringer
		expected string
	}{
		// PatientState
		{name: "PatientState Awake", input: hospital.PatientStateAwake, expected: "Awake"},
		{name: "PatientState Drowsy", input: hospital.PatientStateDrowsy, expected: "Drowsy"},
		{name: "PatientState Asleep", input: hospital.PatientStateAsleep, expected: "Asleep"},
		{name: "PatientState Unknown", input: hospital.PatientState(99), expected: "Unknown PatientState(99)"},

		// NarcolepsyLevel
		{name: "NarcolepsyLevel Mild", input: hospital.NarcolepsyLevelMild, expected: "Mild"},
		{name: "NarcolepsyLevel Moderate", input: hospital.NarcolepsyLevelModerate, expected: "Moderate"},
		{name: "NarcolepsyLevel Severe", input: hospital.NarcolepsyLevelSevere, expected: "Severe"},
		{name: "NarcolepsyLevel Unknown", input: hospital.NarcolepsyLevel(99), expected: "Unknown NarcolepsyLevel(99)"},

		// RoomState
		{name: "RoomState Empty", input: hospital.RoomStateAvailable, expected: "Empty"},
		{name: "RoomState Occupied", input: hospital.RoomStateOccupied, expected: "Occupied"},
		{name: "RoomState Alerting", input: hospital.RoomStateAlerting, expected: "Alerting"},
		{name: "RoomState Unknown", input: hospital.RoomState(99), expected: "Unknown RoomState(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.String()
			if got != tt.expected {
				t.Errorf("String() = %q; se esperaba %q", got, tt.expected)
			}
		})
	}
}
