package hospital

import "testing"

func TestStateStringsAreReadable(t *testing.T) {
	tests := []struct {
		got  string
		want string
	}{
		{PatientStateAwake.String(), "Awake"},
		{PatientStateDrowsy.String(), "Drowsy"},
		{PatientStateAsleep.String(), "Asleep"},
		{PatientState(99).String(), "Unknown"},
		{NarcolepsyLevelMild.String(), "Mild"},
		{NarcolepsyLevelModerate.String(), "Moderate"},
		{NarcolepsyLevelSevere.String(), "Severe"},
		{NarcolepsyLevel(99).String(), "Unknown"},
		{RoomStateAvailable.String(), "Available"},
		{RoomStateAlerting.String(), "Alerting"},
		{RoomStateOccupied.String(), "Occupied"},
		{RoomState(99).String(), "Unknown"},
	}
	for _, tc := range tests {
		if tc.got != tc.want {
			t.Errorf("got %q, want %q", tc.got, tc.want)
		}
	}
}

// La prioridad depende de que Severe > Moderate > Mild numéricamente.
func TestNarcolepsyLevelOrdering(t *testing.T) {
	if !(NarcolepsyLevelMild < NarcolepsyLevelModerate && NarcolepsyLevelModerate < NarcolepsyLevelSevere) {
		t.Error("levels must be ordered Mild < Moderate < Severe")
	}
}
