package hospital

import (
	"ClinicaCoste/internal/hospital/clock"
	"fmt"
)

type Doctor struct {
	Person

	specialty string
	shift     Shift
	patients  []*Patient
	episodes  []EpisodeRecord
}

func NewDoctor(name string, age int, speciality string) *Doctor {
	return &Doctor{
		Person:    NewPerson(name, age),
		specialty: speciality,
		shift:     NewShift(0, 24),
	}
}

func (d *Doctor) Specialty() string { return d.specialty }

func (d *Doctor) Patients() []*Patient {
	out := make([]*Patient, len(d.patients))
	copy(out, d.patients)
	return out
}

func (d *Doctor) MyEpisodes() []EpisodeRecord {
	out := make([]EpisodeRecord, len(d.episodes))
	copy(out, d.episodes)
	return out
}

func (d *Doctor) DiagnosePatient(p *Patient) error {
	if p == nil {
		return ErrNilPatient
	}
	p.setAssignedDoctor(d)
	for _, existing := range d.patients {
		if existing.ID() == p.ID() {
			return nil
		}
	}
	d.patients = append(d.patients, p)
	return nil
}

func (d *Doctor) AttendSleepEmergency(p *Patient) error {
	if p == nil {
		return ErrNilPatient
	}
	if p.State() != PatientStateAsleep {
		return fmt.Errorf("doctor %s, patient %s: %w", d.Name(), p.Name(), ErrPatientNotAsleep)
	}
	return nil
}

func (d *Doctor) Attend(p *Patient, location string) (EpisodeRecord, error) {
	if err := d.AttendSleepEmergency(p); err != nil {
		return EpisodeRecord{}, err
	}
	if err := d.DiagnosePatient(p); err != nil {
		return EpisodeRecord{}, err
	}

	e := NewEpisodeRecord(p, d, location, p.AssignedRoom())
	d.recordEpisode(e)
	return e, nil
}

func (d *Doctor) IsOnShift(clk clock.Clock) bool {
	if clk == nil {
		return true
	}
	return d.shift.IsActive(clk.Now())
}

func (d *Doctor) recordEpisode(e EpisodeRecord) { d.episodes = append(d.episodes, e) }
