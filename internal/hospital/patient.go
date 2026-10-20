package hospital

import (
	"fmt"
	"math/rand/v2"
	"time"
)

type Patient struct {
	Person
	level           NarcolepsyLevel
	state           PatientState
	currentLocation string
	assignedRoom    *Room
	assignedDoctor  Attender
	rng             *rand.Rand
}

func NewPatient(name string, age int, level NarcolepsyLevel) *Patient {
	return &Patient{
		Person: NewPerson(name, age),
		level:  level,
		state:  PatientStateAwake,
		rng:    rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())),
	}
}

func (p *Patient) Level() NarcolepsyLevel {
	return p.level
}

func (p *Patient) State() PatientState {
	return p.state
}

func (p *Patient) Location() string {
	return p.currentLocation
}

func (p *Patient) AssignedRoom() *Room {
	return p.assignedRoom
}

func (p *Patient) AssignedDoctor() Attender {
	return p.assignedDoctor
}

func (p *Patient) IsAsleepInHallway() bool {
	return p.state == PatientStateAsleep && p.assignedRoom == nil
}

func (p *Patient) IsAsleepInBed() bool {
	return p.state == PatientStateAsleep && p.assignedRoom != nil
}

func (p *Patient) Priority() (urgency, severity int) {
	switch {
	case p.IsAsleepInHallway():
		urgency = 2
	case p.state == PatientStateDrowsy:
		urgency = 1
	default:
		urgency = 0
	}
	return urgency, int(p.level)
}

func (p *Patient) HasHigherPriority(other *Patient) bool {
	u1, s1 := p.Priority()
	u2, s2 := other.Priority()
	if u1 != u2 {
		return u1 > u2
	}
	return s1 > s2
}

func (p *Patient) SufferSleepAttack(location string) error {
	if p.state == PatientStateAsleep {
		return fmt.Errorf("patient %s %w", p.ID(), ErrPatientAlreadyAsleep)
	}

	p.currentLocation = location
	p.state = PatientStateDrowsy

	//TODO Concurrence
	p.state = PatientStateAsleep
	return nil
}

func (p *Patient) WakeUp() error {
	if p.state == PatientStateAwake {
		return fmt.Errorf("patient %s: %w", p.ID(), ErrPatientAlreadyAwake)
	}
	p.state = PatientStateAwake
	p.currentLocation = ""
	return nil
}
func (p *Patient) nextAttackDelay() time.Duration {
	base := 5 * time.Second
	switch p.level {
	case NarcolepsyLevelSevere:
		base = 2 * time.Second
	case NarcolepsyLevelMild:
		base = 10 * time.Second
	default:
		base = 5 * time.Second
	}
	n := p.rng.IntN(3)
	return base + time.Duration(n)*time.Second
}

func (p *Patient) setAssignedRoom(r *Room)      { p.assignedRoom = r }
func (p *Patient) clearAssignedRoom()           { p.assignedRoom = nil }
func (p *Patient) setAssignedDoctor(a Attender) { p.assignedDoctor = a }

func (p *Patient) String() string {
	room := "none"
	if p.assignedRoom != nil {
		room = fmt.Sprintf("%d", p.assignedRoom.Number())
	}
	return fmt.Sprintf("P-%s | %s (age %d, %s) | %s | location: %q | room: %s",
		shortID(p.ID()), p.Name(), p.Age(), p.level, p.state, p.currentLocation, room)
}
