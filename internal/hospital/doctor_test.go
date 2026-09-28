package hospital

import "testing"

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
	d.addPatient(NewPatient("a", 30, NarcolepsyLevelMild))

	list := d.Patients()
	list[0] = nil
	if d.Patients()[0] == nil {
		t.Error("Patients must return a copy")
	}
}
