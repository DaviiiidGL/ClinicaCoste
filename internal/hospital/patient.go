package hospital

import "fmt"

type Patient struct {
	Person
	level           NarcolepsyLevel
	state           PatientState
	currentLocation string
	assignedRoom    *Room
	assignedDoctor  interface{}
}

func newPatient(name string, age int, level NarcolepsyLevel) *Patient {
	return &Patient{
		Person:          *NewPerson(name, age),
		level:           level,
		state:           PatientStateAwake,
		currentLocation: "",
		assignedRoom:    nil,
		assignedDoctor:  nil,
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

func (p *Patient) AssignedDoctor() interface{} {
	return p.assignedDoctor
}

// Priority: 2>1>0 where 2 is maximum urgency
func (p *Patient) Priority() (int, int) {
	var state int
	if p.state == PatientStateAsleep && p.assignedRoom == nil {
		state = 2
	} else if p.state == PatientStateDrowsy {
		state = 1
	} else {
		state = 0
	}
	return state, int(p.level)
}

// Comparing
func (p *Patient) HasHigherPriority(other *Patient) bool {
	s1, l1 := p.Priority()
	s2, l2 := other.Priority()

	if s1 != s2 {
		return s1 > s2
	}
	return l1 > l2
}

// Patient asleep at location
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

// Wake up Patient
func (p *Patient) WakeUp() error {
	if p.state == PatientStateAwake {
		return fmt.Errorf("patient %s %w", p.ID(), ErrPatientAlreadyAwake)
	}

	p.state = PatientStateAwake
	p.currentLocation = ""

	return nil
}

func (p *Patient) SetAssignedRoom(room *Room) {
	p.assignedRoom = room
	if room != nil && p.state == PatientStateAsleep {
		//TODO: Update state of room
	}
}

func (p *Patient) ClearAssignedRoom() {
	p.assignedRoom = nil
}

func (p *Patient) SetAssignedDoctor(d interface{}) {
	p.assignedDoctor = d
}

func (p *Patient) String() string {
	room := "none"
	if p.assignedRoom != nil {
		room = fmt.Sprintf("%v", p.assignedRoom)
	}

	return fmt.Sprintf("P-%s | %s (age %d, %s, %s) | State: %s | Location: %s | Room: %s",
		p.ID()[:8], p.Name(), p.Age(), p.Level(), p.Level(),
		p.State(), p.Location(), room)
}
