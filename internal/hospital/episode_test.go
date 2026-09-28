package hospital

import (
	"strings"
	"testing"
)

// fakeAttender es un doble de prueba: demuestra que cualquier tipo con estos
// tres métodos satisface Attender (lo usará la Fase 3 con Doctor/Orderly/Nurse).
type fakeAttender struct{ id, name string }

func (f fakeAttender) ID() string   { return f.id }
func (f fakeAttender) Name() string { return f.name }
func (f fakeAttender) Attend(p *Patient, location string) (EpisodeRecord, error) {
	return NewEpisodeRecord(p, f, location, nil), nil
}

var _ Attender = fakeAttender{}

func TestNewEpisodeRecord(t *testing.T) {
	p := NewPatient("Yeimy", 29, NarcolepsyLevelSevere)
	room, _ := NewRoom(101, 1)
	by := fakeAttender{"D-1", "Dr. Karen Ospina"}

	e := NewEpisodeRecord(p, by, "cafeteria", room)

	if e.ID() == "" || e.Timestamp().IsZero() {
		t.Error("id and timestamp must be set")
	}
	if e.Patient() != p || e.Location() != "cafeteria" || e.AssignedRoom() != room {
		t.Errorf("unexpected record: %v", e)
	}
	if e.AttendingDoctor().Name() != "Dr. Karen Ospina" {
		t.Error("attender not stored")
	}
}

func TestEpisodeSummary(t *testing.T) {
	p := NewPatient("Yeimy", 29, NarcolepsyLevelSevere)
	room, _ := NewRoom(101, 1)
	by := fakeAttender{"D-1", "Dr. Karen Ospina"}

	withRoom := NewEpisodeRecord(p, by, "cafeteria", room).Summary()
	for _, want := range []string{"cafeteria", "room 101", "Dr. Karen Ospina", shortID(p.ID())} {
		if !strings.Contains(withRoom, want) {
			t.Errorf("summary %q should contain %q", withRoom, want)
		}
	}

	hallway := NewEpisodeRecord(p, nil, "radiology queue", nil).Summary()
	if !strings.Contains(hallway, "hallway (no room)") || !strings.Contains(hallway, "unassigned") {
		t.Errorf("summary without room/attender: %q", hallway)
	}
}

func TestEpisodeRecordIsAValueType(t *testing.T) {
	e := NewEpisodeRecord(nil, nil, "x", nil)
	copyOfE := e
	if copyOfE.ID() != e.ID() {
		t.Error("value copy must keep the same data")
	}
	_ = e.Summary() // nil patient must not panic
}
