package hospital

import (
	"fmt"
	"slices"
	"strings"
	"sync"
)

type Hospital struct {
	name     string
	doctors  []*Doctor
	patients []*Patient
	byID     map[string]*Patient
	rooms    []*Room
	history  []EpisodeRecord

	subMu     sync.Mutex
	subs      map[int]chan EpisodeRecord
	nextSubID int
}

func NewHospital(name string) (*Hospital, error) {
	if strings.TrimSpace(name) == "" {
		return nil, ErrHospitalNameRequired
	}
	return &Hospital{
		name: name,
		byID: make(map[string]*Patient),
		subs: make(map[int]chan EpisodeRecord),
	}, nil
}

func (h *Hospital) Name() string { return h.name }

func (h *Hospital) AdmitPatient(p *Patient) error {
	if p == nil {
		return ErrNilPatient
	}
	if _, exists := h.byID[p.ID()]; exists {
		return fmt.Errorf("patient %s: %w", p.Name(), ErrPatientAlreadyAdmitted)
	}
	h.patients = append(h.patients, p)
	h.byID[p.ID()] = p
	return nil
}

func (h *Hospital) HireDoctor(d *Doctor) error {
	if d == nil {
		return ErrNilDoctor
	}
	for _, existing := range h.doctors {
		if existing.ID() == d.ID() {
			return fmt.Errorf("doctor %s: %w", d.Name(), ErrDoctorAlreadyHired)
		}
	}
	h.doctors = append(h.doctors, d)
	return nil
}

func (h *Hospital) AddRoom(number, capacity int) (*Room, error) {
	for _, r := range h.rooms {
		if r.Number() == number {
			return nil, fmt.Errorf("room %d: %w", number, ErrRoomAlreadyRegistered)
		}
	}
	room, err := NewRoom(number, capacity)
	if err != nil {
		return nil, err
	}
	h.rooms = append(h.rooms, room)
	return room, nil
}

func (h *Hospital) ReportSleepAttack(p *Patient, location string) (*Room, error) {
	if p == nil {
		return nil, ErrNilPatient
	}
	if h.findPatient(p.ID()) == nil {
		return nil, fmt.Errorf("patient %s: %w", p.Name(), ErrPatientNotFound)
	}
	if err := p.SufferSleepAttack(location); err != nil {
		return nil, err
	}
	return h.AssignRoom(p)
}

func (h *Hospital) AssignRoom(p *Patient) (*Room, error) {
	if p == nil {
		return nil, ErrNilPatient
	}
	if h.findPatient(p.ID()) == nil {
		return nil, fmt.Errorf("patient %s: %w", p.Name(), ErrPatientNotFound)
	}
	if p.State() != PatientStateAsleep {
		return nil, fmt.Errorf("patient %s: %w", p.Name(), ErrPatientNotAsleep)
	}
	if p.AssignedRoom() != nil {
		return nil, fmt.Errorf("patient %s: %w", p.Name(), ErrPatientAlreadyInRoom)
	}
	for _, room := range h.rooms {
		if room.IsAvailable() {
			if err := room.Occupy(p); err != nil {
				return nil, err
			}
			return room, nil
		}
	}
	return nil, fmt.Errorf("patient %s: %w", p.Name(), ErrNoRoomAvailable)
}

func (h *Hospital) WakePatient(p *Patient) error {
	if p == nil {
		return ErrNilPatient
	}
	if h.findPatient(p.ID()) == nil {
		return fmt.Errorf("patient %s: %w", p.Name(), ErrPatientNotFound)
	}
	if err := p.WakeUp(); err != nil {
		return err
	}
	if room := p.AssignedRoom(); room != nil {
		return room.Release(p)
	}
	return nil
}

func (h *Hospital) AssignNextWaiting() (*Patient, *Room, error) {
	waiting := h.PatientsInHallway() // copia: se puede reordenar sin tocar h.patients
	if len(waiting) == 0 {
		return nil, nil, ErrNoPatientWaiting
	}
	slices.SortStableFunc(waiting, func(a, b *Patient) int {
		switch {
		case a.HasHigherPriority(b):
			return -1
		case b.HasHigherPriority(a):
			return 1
		default:
			return 0
		}
	})
	next := waiting[0]
	room, err := h.AssignRoom(next)
	if err != nil {
		return nil, nil, err
	}
	return next, room, nil
}

func (h *Hospital) RegisterEpisode(e EpisodeRecord) error {
	if e.Patient() == nil {
		return ErrNilPatient
	}
	if h.findPatient(e.Patient().ID()) == nil {
		return fmt.Errorf("patient %s: %w", e.Patient().Name(), ErrPatientNotFound)
	}
	h.history = append(h.history, e)
	h.publish(e)
	return nil
}

func (h *Hospital) PatientsInHallway() []*Patient {
	var out []*Patient
	for _, p := range h.patients {
		if p.IsAsleepInHallway() {
			out = append(out, p)
		}
	}
	return out
}

func (h *Hospital) Patients() []*Patient     { return slices.Clone(h.patients) }
func (h *Hospital) Doctors() []*Doctor       { return slices.Clone(h.doctors) }
func (h *Hospital) Rooms() []*Room           { return slices.Clone(h.rooms) }
func (h *Hospital) History() []EpisodeRecord { return slices.Clone(h.history) }

func (h *Hospital) findPatient(id string) *Patient {
	return h.byID[id]
}
