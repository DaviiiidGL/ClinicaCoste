package hospital

type Orderly struct {
	Person

	patientsMoved []*Patient
}

func NewOrderly(name string, age int) *Orderly {
	return &Orderly{Person: NewPerson(name, age)}
}

func (o *Orderly) CanDiagnose() bool { return false }

func (o *Orderly) PatientsMoved() []*Patient {
	out := make([]*Patient, len(o.patientsMoved))
	copy(out, o.patientsMoved)
	return out
}

func (o *Orderly) Attend(p *Patient, location string) (EpisodeRecord, error) {
	if p == nil {
		return EpisodeRecord{}, ErrNilPatient
	}
	if p.State() != PatientStateAsleep {
		return EpisodeRecord{}, ErrPatientNotAsleep
	}
	p.setAssignedDoctor(o)
	o.patientsMoved = append(o.patientsMoved, p)
	return NewEpisodeRecord(p, o, location, p.AssignedRoom()), nil
}
