package hospital

import (
	"errors"
	"testing"
)

func TestPatientSleepCycle(t *testing.T) {
	p := NewPatient("Wilfrido", 40, NarcolepsyLevelSevere)

	if err := p.WakeUp(); !errors.Is(err, ErrPatientAlreadyAwake) {
		t.Errorf("WakeUp awake: got %v", err)
	}
	if err := p.SufferSleepAttack("cafeteria"); err != nil {
		t.Fatalf("SufferSleepAttack: %v", err)
	}
	if !p.IsAsleepInHallway() || p.Location() != "cafeteria" {
		t.Errorf("want asleep in hallway at cafeteria, got %v", p)
	}
	if err := p.SufferSleepAttack("x"); !errors.Is(err, ErrPatientAlreadyAsleep) {
		t.Errorf("double attack: got %v", err)
	}
}

func TestPriority_HallwaySevereBeatsBedSevere(t *testing.T) {
	inBed := NewPatient("bed", 30, NarcolepsyLevelSevere)
	waiting := NewPatient("wait", 30, NarcolepsyLevelMild)
	room, _ := NewRoom(1, 1)

	_ = inBed.SufferSleepAttack("a")
	_ = room.Occupy(inBed)
	_ = waiting.SufferSleepAttack("b")

	if !waiting.HasHigherPriority(inBed) {
		t.Error("a Mild in the hallway must outrank a Severe already in bed")
	}
}

func TestPriority_SeverityBreaksTies(t *testing.T) {
	mild := NewPatient("m", 30, NarcolepsyLevelMild)
	severe := NewPatient("s", 30, NarcolepsyLevelSevere)
	_ = mild.SufferSleepAttack("a")
	_ = severe.SufferSleepAttack("b")

	if !severe.HasHigherPriority(mild) || mild.HasHigherPriority(severe) {
		t.Error("among hallway patients, Severe goes first")
	}
}

func TestStateStrings(t *testing.T) {
	if PatientStateDrowsy.String() != "Drowsy" ||
		NarcolepsyLevelModerate.String() != "Moderate" ||
		RoomStateAlerting.String() != "Alerting" {
		t.Error("unexpected String() output")
	}
}

func TestNewPatientStartsAwakeAndUnassigned(t *testing.T) {
	p := NewPatient("Wilfrido", 40, NarcolepsyLevelModerate)
	if p.State() != PatientStateAwake || p.AssignedRoom() != nil || p.AssignedDoctor() != nil || p.Location() != "" {
		t.Errorf("unexpected initial state: %v", p)
	}
}

func TestWakeUpClearsLocationButKeepsRoomUntilReleased(t *testing.T) {
	p := NewPatient("a", 30, NarcolepsyLevelSevere)
	room, _ := NewRoom(1, 1)
	_ = p.SufferSleepAttack("hallway 2")
	_ = room.Occupy(p)

	if !p.IsAsleepInBed() {
		t.Fatal("should be asleep in bed")
	}
	if err := p.WakeUp(); err != nil {
		t.Fatalf("WakeUp: %v", err)
	}
	if p.Location() != "" || p.State() != PatientStateAwake {
		t.Errorf("WakeUp must clear location: %v", p)
	}
	if p.AssignedRoom() == nil {
		t.Error("the room is released by Room.Release, not by WakeUp")
	}
	_ = room.Release(p)
	if p.AssignedRoom() != nil {
		t.Error("Release must clear the patient's room")
	}
}

func TestPriorityValues(t *testing.T) {
	awake := NewPatient("a", 30, NarcolepsyLevelSevere)
	if u, _ := awake.Priority(); u != 0 {
		t.Errorf("awake urgency = %d, want 0", u)
	}
	hall := NewPatient("b", 30, NarcolepsyLevelModerate)
	_ = hall.SufferSleepAttack("x")
	if u, s := hall.Priority(); u != 2 || s != int(NarcolepsyLevelModerate) {
		t.Errorf("hallway priority = (%d,%d)", u, s)
	}
}
