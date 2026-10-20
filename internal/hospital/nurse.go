package hospital

type Nurse struct {
	Person
	stabilized []*Patient
}

func NewNurse(name string, age int) *Nurse {
	return &Nurse{Person: NewPerson(name, age)}
}

func (n *Nurse) CanDiagnose() bool { return false }

func (n *Nurse) Stabilized() []*Patient {
	out := make([]*Patient, len(n.stabilized))
	copy(out, n.stabilized)
	return out
}

func (n *Nurse) Attend(p *Patient, location string) (EpisodeRecord, error) {
	if p == nil {
		return EpisodeRecord{}, ErrNilPatient
	}
	if p.State() != PatientStateAsleep {
		return EpisodeRecord{}, ErrPatientNotAsleep
	}
	p.setAssignedDoctor(n)
	n.stabilized = append(n.stabilized, p)
	return NewEpisodeRecord(p, n, location, p.AssignedRoom()), nil
}
