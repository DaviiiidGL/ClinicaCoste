package hospital

import "fmt"

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
		return EpisodeRecord{}, fmt.Errorf("orderly %s, patient %s: %w", o.Name(), p.Name(), ErrPatientNotAsleep)
	}
	p.setAssignedDoctor(o) // "quién lo está atendiendo ahora", no implica que sea médico
	o.patientsMoved = append(o.patientsMoved, p)
	return NewEpisodeRecord(p, o, location, p.AssignedRoom()), nil
}
