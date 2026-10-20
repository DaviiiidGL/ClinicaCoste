package hospital

import (
	"ClinicaCoste/internal/hospital/clock"
	"container/heap"
	"slices"
	"strings"
	"sync"
)

type Hospital struct {
	name     string
	doctors  []*Doctor
	staff    []Attender
	nextDisp int
	patients []*Patient
	byID     map[string]*Patient
	history  []EpisodeRecord
	floors   []*Floor

	clock clock.Clock

	mu        sync.RWMutex
	subMu     sync.Mutex
	subs      map[int]chan EpisodeRecord
	nextSubID int
}

type PatientHeap []*Patient

func NewHospital(name string, clk clock.Clock) (*Hospital, error) {
	if strings.TrimSpace(name) == "" {
		return nil, ErrHospitalNameRequired
	}

	if clk == nil {
		clk = clock.RealClock{}
	}

	return &Hospital{
		name:   name,
		byID:   make(map[string]*Patient),
		floors: []*Floor{NewFloor(1, "Piso Principal", nil)},
		clock:  clk,
		subs:   make(map[int]chan EpisodeRecord),
	}, nil
}

func (h *Hospital) Name() string { return h.name }

func (h *Hospital) AdmitPatient(p *Patient) error {
	if p == nil {
		return ErrNilPatient
	}
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.byID[p.ID()]; exists {
		return ErrPatientAlreadyAdmitted
	}
	h.patients = append(h.patients, p)
	h.byID[p.ID()] = p
	return nil
}

func (h *Hospital) HireDoctor(d *Doctor) error {
	if d == nil {
		return ErrNilDoctor
	}
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, existing := range h.doctors {
		if existing.ID() == d.ID() {
			return ErrDoctorAlreadyHired
		}
	}
	h.doctors = append(h.doctors, d)
	h.staff = append(h.staff, d)
	return nil
}

func (h *Hospital) AddAttender(a Attender) error {
	if a == nil {
		return ErrNilAttender
	}
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, existing := range h.staff {
		if existing.ID() == a.ID() {
			return ErrAttenderAlreadyRegistered
		}
	}
	h.staff = append(h.staff, a)
	return nil
}

func (h *Hospital) Staff() []Attender {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return slices.Clone(h.staff)
}

func (h *Hospital) dispatch() (Attender, bool) {
	var available []Attender
	for _, a := range h.staff {
		if doc, ok := a.(*Doctor); ok {
			if doc.IsOnShift(h.clock) {
				available = append(available, doc)
			}
		} else {
			available = append(available, a)
		}
	}
	if len(available) == 0 {
		return nil, false
	}
	a := available[h.nextDisp%len(available)]
	h.nextDisp++
	return a, true
}

func (h *Hospital) AddRoom(number, capacity int) (*Room, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, floor := range h.floors {
		for _, r := range floor.Rooms() {
			if r.Number() == number {
				return nil, ErrRoomAlreadyRegistered
			}
		}
	}

	room, err := NewRoom(number, capacity)
	if err != nil {
		return nil, err
	}

	if len(h.floors) == 0 {
		h.floors = append(h.floors, NewFloor(1, "Piso Principal", nil))
	}

	if err := h.floors[0].AddRoom(room); err != nil {
		return nil, err
	}

	return room, nil
}

func (h *Hospital) ReportSleepAttack(p *Patient, location string) (*Room, error) {
	if p == nil {
		return nil, ErrNilPatient
	}

	h.mu.Lock()
	if _, exists := h.byID[p.ID()]; !exists {
		h.mu.Unlock()
		return nil, ErrPatientNotFound
	}
	h.mu.Unlock()

	if err := p.SufferSleepAttack(location); err != nil {
		return nil, err
	}

	h.mu.Lock()
	room, roomErr := h.assignRoomUnsafe(p)

	var e EpisodeRecord
	var attendingDoctor Attender

	if attender, ok := h.dispatch(); ok {
		if ep, err := attender.Attend(p, location); err == nil {
			attendingDoctor = attender
			e = ep
		}
	}

	if attendingDoctor == nil {
		e = NewEpisodeRecord(p, nil, location, room)
		e.timestamp = h.clock.Now()
	}

	h.history = append(h.history, e)
	h.publish(e)
	h.mu.Unlock()

	return room, roomErr
}

func (h *Hospital) ProcessQueue(attender Attender) error {
	if attender == nil {
		return ErrNilAttender
	}

	h.mu.Lock()
	waiting := h.patientsInHallwayUnsafe()
	if len(waiting) == 0 {
		h.mu.Unlock()
		return ErrNoPatientWaiting
	}
	hp := PatientHeap(waiting)
	heap.Init(&hp)

	nextPatient := heap.Pop(&hp).(*Patient)
	if nextPatient.AssignedRoom() == nil {
		_, _ = h.assignRoomUnsafe(nextPatient)
	}
	h.mu.Unlock()

	e, err := attender.Attend(nextPatient, nextPatient.Location())
	if err != nil {
		return err
	}
	return h.RegisterEpisode(e)
}

func (h *Hospital) attendAndRegister(p *Patient, location string) error {
	attender, ok := h.dispatch()
	if !ok {
		return h.RegisterEpisode(NewEpisodeRecord(p, nil, location, p.AssignedRoom()))
	}
	e, err := attender.Attend(p, location)
	if err != nil {
		return err
	}
	return h.RegisterEpisode(e)
}

func (h *Hospital) AssignRoom(p *Patient) (*Room, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.assignRoomUnsafe(p)
}

func (h *Hospital) WakePatient(p *Patient) error {
	if p == nil {
		return ErrNilPatient
	}

	h.mu.Lock()
	if _, exists := h.byID[p.ID()]; !exists {
		h.mu.Unlock()
		return ErrPatientNotFound
	}

	if err := p.WakeUp(); err != nil {
		h.mu.Unlock()
		return err
	}

	if room := p.AssignedRoom(); room != nil {
		if err := room.Release(p); err != nil {
			h.mu.Unlock()
			return err
		}
	}

	_, _, _ = h.AssignNextWaitingUnsafe()

	h.mu.Unlock()

	return nil
}

func (h *Hospital) Clock() clock.Clock {
	return h.clock
}

func (h *Hospital) assignRoomUnsafe(p *Patient) (*Room, error) {
	if p == nil {
		return nil, ErrNilPatient
	}
	if _, exists := h.byID[p.ID()]; !exists {
		return nil, ErrPatientNotFound
	}
	if p.State() != PatientStateAsleep {
		return nil, ErrPatientNotAsleep
	}
	if p.AssignedRoom() != nil {
		return nil, ErrPatientAlreadyInRoom
	}

	for _, floor := range h.floors {
		for _, room := range floor.Rooms() {
			if room.IsAvailable() {
				if err := room.Occupy(p); err != nil {
					return nil, err
				}
				return room, nil
			}
		}
	}
	return nil, ErrNoRoomAvailable
}

func (h *Hospital) AssignNextWaiting() (*Patient, *Room, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.assignNextWaitingUnsafe()
}

func (h *Hospital) assignNextWaitingUnsafe() (*Patient, *Room, error) {
	waiting := h.patientsInHallwayUnsafe()
	if len(waiting) == 0 {
		return nil, nil, ErrNoPatientWaiting
	}

	hp := PatientHeap(waiting)
	heap.Init(&hp)

	next := heap.Pop(&hp).(*Patient)

	room, err := h.assignRoomUnsafe(next)
	if err != nil {
		return nil, nil, err
	}
	return next, room, nil
}

func (h *Hospital) AssignNextWaitingUnsafe() (*Patient, *Room, error) {
	waiting := h.patientsInHallwayUnsafe()
	if len(waiting) == 0 {
		return nil, nil, ErrNoPatientWaiting
	}

	hp := PatientHeap(waiting)
	heap.Init(&hp)

	next := heap.Pop(&hp).(*Patient)

	room, err := h.assignRoomUnsafe(next)
	if err != nil {
		return nil, nil, err
	}
	return next, room, nil
}

func (h *Hospital) RegisterEpisode(e EpisodeRecord) error {
	if e.Patient() == nil {
		return ErrNilPatient
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.byID[e.Patient().ID()]; !exists {
		return ErrPatientNotFound
	}
	h.history = append(h.history, e)
	h.publish(e)
	return nil
}

func (h *Hospital) PatientsInHallway() []*Patient {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.patientsInHallwayUnsafe()
}

func (h *Hospital) patientsInHallwayUnsafe() []*Patient {
	out := make([]*Patient, 0, len(h.patients))

	for _, p := range h.patients {
		if p.IsAsleepInHallway() {
			out = append(out, p)
		}
	}
	return out
}

func (h *Hospital) Rooms() []*Room {
	h.mu.RLock()
	defer h.mu.RUnlock()
	var allRooms []*Room
	for _, f := range h.floors {
		allRooms = append(allRooms, f.Rooms()...)
	}
	return allRooms
}

func (h *Hospital) FloorSnapshot(floorNumber int) []PatientView {
	h.mu.RLock()
	defer h.mu.RUnlock()
	var targetFloor *Floor
	for _, f := range h.floors {
		if f.Number() == floorNumber {
			targetFloor = f
			break
		}
	}
	if targetFloor == nil {
		return nil
	}

	roomNumbers := make(map[int]bool)
	for _, r := range targetFloor.Rooms() {
		roomNumbers[r.Number()] = true
	}

	var snapshot []PatientView
	for _, p := range h.patients {
		roomNum := 0
		assignedToThisFloor := false

		if p.AssignedRoom() != nil {
			roomNum = p.AssignedRoom().Number()
			if roomNumbers[roomNum] {
				assignedToThisFloor = true
			}
		}
		if assignedToThisFloor {
			snapshot = append(snapshot, PatientView{
				ID:           p.ID(),
				ShortID:      shortID(p.ID()),
				Name:         p.Name(),
				State:        p.State().String(),
				Location:     p.Location(),
				AssignedRoom: roomNum,
				IsCrisis:     p.IsAsleepInHallway(),
			})
		}
	}
	return snapshot
}

func (h *Hospital) Patients() []*Patient {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return slices.Clone(h.patients)
}

func (h *Hospital) Doctors() []*Doctor {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return slices.Clone(h.doctors)
}

func (h *Hospital) History() []EpisodeRecord {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return slices.Clone(h.history)
}

func (h *Hospital) findPatient(id string) *Patient {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.byID[id]
}

func (h *PatientHeap) Len() int { return len(*h) }

func (h *PatientHeap) Less(i, j int) bool {
	return (*h)[i].HasHigherPriority((*h)[j])
}

func (h *PatientHeap) Swap(i, j int) {
	(*h)[i], (*h)[j] = (*h)[j], (*h)[i]
}

func (h *PatientHeap) Push(x any) {
	*h = append(*h, x.(*Patient))
}

func (h *PatientHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[0 : n-1]
	return item
}
