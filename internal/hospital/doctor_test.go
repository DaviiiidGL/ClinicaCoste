package hospital

import (
	"errors"
	"testing"
)

func TestNewDoctor(t *testing.T) {
	d := NewDoctor("Karen Ospina", 38, "Neurology")

	if d.Specialty() != "Neurology" || d.Name() != "Karen Ospina" {
		t.Errorf("unexpected doctor: %+v", d)
	}
	if len(d.Patients()) != 0 || len(d.MyEpisodes()) != 0 {
		t.Error("a new doctor has no patients and no episodes")
	}
}

func TestDoctorMyEpisodes_ReturnsOnlyOwnAndACopy(t *testing.T) {
	d := NewDoctor("Karen", 38, "Neurology")
	p := NewPatient("Wilfrido", 40, NarcolepsyLevelSevere)

	d.recordEpisode(NewEpisodeRecord(p, nil, "cafeteria", nil))
	d.recordEpisode(NewEpisodeRecord(p, nil, "radiology queue", nil))

	eps := d.MyEpisodes()
	if len(eps) != 2 || eps[0].Location() != "cafeteria" {
		t.Fatalf("unexpected episodes: %v", eps)
	}

	eps[0] = EpisodeRecord{} // mutar la copia no debe afectar al doctor
	if d.MyEpisodes()[0].Location() != "cafeteria" {
		t.Error("MyEpisodes must return a copy")
	}
}

func TestDoctorPatients_ReturnsACopy(t *testing.T) {
	d := NewDoctor("Karen", 38, "Neurology")
	_ = d.DiagnosePatient(NewPatient("a", 30, NarcolepsyLevelMild))

	list := d.Patients()
	list[0] = nil
	if d.Patients()[0] == nil {
		t.Error("Patients must return a copy")
	}
}

func TestDiagnosePatient(t *testing.T) {
	d := NewDoctor("Karen", 38, "Neurology")

	if err := d.DiagnosePatient(nil); !errors.Is(err, ErrNilPatient) {
		t.Errorf("nil: got %v", err)
	}

	p := NewPatient("Wilfrido", 40, NarcolepsyLevelSevere)
	if err := d.DiagnosePatient(p); err != nil {
		t.Fatalf("DiagnosePatient: %v", err)
	}
	if p.AssignedDoctor() == nil || p.AssignedDoctor().ID() != d.ID() {
		t.Error("patient should be assigned to this doctor")
	}
	if err := d.DiagnosePatient(p); err != nil { // idempotente
		t.Fatalf("DiagnosePatient again: %v", err)
	}
	if len(d.Patients()) != 1 {
		t.Errorf("patient must not be duplicated, got %d", len(d.Patients()))
	}
}

func TestAttendSleepEmergency(t *testing.T) {
	d := NewDoctor("Karen", 38, "Neurology")

	if err := d.AttendSleepEmergency(nil); !errors.Is(err, ErrNilPatient) {
		t.Errorf("nil: got %v", err)
	}
	awake := NewPatient("Awake", 30, NarcolepsyLevelMild)
	if err := d.AttendSleepEmergency(awake); !errors.Is(err, ErrPatientNotAsleep) {
		t.Errorf("awake patient: got %v", err)
	}
}

func TestDoctorAttend_DiagnosesAndRecordsItsOwnEpisode(t *testing.T) {
	d := NewDoctor("Karen", 38, "Neurology")
	p := NewPatient("Wilfrido", 40, NarcolepsyLevelSevere)

	if _, err := d.Attend(p, "cafeteria"); !errors.Is(err, ErrPatientNotAsleep) {
		t.Fatalf("attending an awake patient: got %v", err)
	}

	_ = p.SufferSleepAttack("cafeteria")
	e, err := d.Attend(p, "cafeteria")
	if err != nil {
		t.Fatalf("Attend: %v", err)
	}
	if e.AttendingDoctor().ID() != d.ID() || e.Location() != "cafeteria" {
		t.Errorf("unexpected episode: %v", e)
	}
	if p.AssignedDoctor() == nil || p.AssignedDoctor().ID() != d.ID() {
		t.Error("Attend must diagnose (assign) the patient too")
	}
	if got := d.MyEpisodes(); len(got) != 1 || got[0].ID() != e.ID() {
		t.Errorf("MyEpisodes should include this episode: %v", got)
	}
}

// Attend es la interfaz Attender: comprobación en tiempo de compilación de
// que Doctor la satisface.
var _ Attender = (*Doctor)(nil)
