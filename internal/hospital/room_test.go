package hospital

import (
	"errors"
	"testing"
)

func mustRoom(t *testing.T, number, capacity int) *Room {
	t.Helper()
	r, err := NewRoom(number, capacity)
	if err != nil {
		t.Fatalf("NewRoom: %v", err)
	}
	return r
}

func TestNewRoom_InvalidCapacity(t *testing.T) {
	if _, err := NewRoom(1, 0); !errors.Is(err, ErrInvalidCapacity) {
		t.Fatalf("want ErrInvalidCapacity, got %v", err)
	}
}

func TestRoomState(t *testing.T) {
	tests := []struct {
		name     string
		capacity int
		filled   int
		want     RoomState
		avail    bool
	}{
		{"capacity 1 empty", 1, 0, RoomStateAvailable, true},
		{"capacity 1 full", 1, 1, RoomStateOccupied, false},
		{"capacity 4 with 2", 4, 2, RoomStateAvailable, true},
		{"capacity 4 with 3 (75%)", 4, 3, RoomStateAlerting, true},
		{"capacity 4 full", 4, 4, RoomStateOccupied, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := mustRoom(t, 1, tc.capacity)
			for i := 0; i < tc.filled; i++ {
				if err := r.Occupy(NewPatient("p", 30, NarcolepsyLevelMild)); err != nil {
					t.Fatalf("Occupy: %v", err)
				}
			}
			if got := r.State(); got != tc.want {
				t.Errorf("State() = %v, want %v", got, tc.want)
			}
			if got := r.IsAvailable(); got != tc.avail {
				t.Errorf("IsAvailable() = %v, want %v", got, tc.avail)
			}
		})
	}
}

func TestRoomOccupyErrors(t *testing.T) {
	r := mustRoom(t, 1, 1)
	p1 := NewPatient("a", 30, NarcolepsyLevelSevere)
	p2 := NewPatient("b", 30, NarcolepsyLevelMild)

	if err := r.Occupy(nil); !errors.Is(err, ErrNilPatient) {
		t.Errorf("nil: got %v", err)
	}
	if err := r.Occupy(p1); err != nil {
		t.Fatalf("Occupy p1: %v", err)
	}
	if err := r.Occupy(p2); !errors.Is(err, ErrRoomFull) {
		t.Errorf("full: got %v", err)
	}
	other := mustRoom(t, 2, 1)
	if err := other.Occupy(p1); !errors.Is(err, ErrPatientAlreadyInRoom) {
		t.Errorf("already in room: got %v", err)
	}
}

func TestRoomRelease(t *testing.T) {
	r := mustRoom(t, 1, 1)
	p := NewPatient("a", 30, NarcolepsyLevelSevere)
	stranger := NewPatient("b", 30, NarcolepsyLevelMild)
	_ = r.Occupy(p)

	if err := r.Release(stranger); !errors.Is(err, ErrPatientNotInRoom) {
		t.Errorf("stranger: got %v", err)
	}
	if err := r.Release(p); err != nil {
		t.Fatalf("Release: %v", err)
	}
	if p.AssignedRoom() != nil || r.State() != RoomStateAvailable {
		t.Errorf("room/patient not cleaned up: %v / %v", p.AssignedRoom(), r.State())
	}
}

func TestRoomOccupants_ReturnsACopy(t *testing.T) {
	r := mustRoom(t, 1, 2)
	_ = r.Occupy(NewPatient("a", 30, NarcolepsyLevelMild))

	list := r.Occupants()
	list[0] = nil
	if r.Occupants()[0] == nil {
		t.Error("Occupants must return a copy")
	}
}

func TestRoomCapacityOneNeverAlerts(t *testing.T) {
	r := mustRoom(t, 1, 1)
	if r.State() != RoomStateAvailable {
		t.Fatalf("empty: %v", r.State())
	}
	_ = r.Occupy(NewPatient("a", 30, NarcolepsyLevelMild))
	if r.State() != RoomStateOccupied {
		t.Errorf("full: %v", r.State())
	}
}
