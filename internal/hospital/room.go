package hospital

import "fmt"

const alertThresholdPercent = 75

type Room struct {
	number    int
	capacity  int
	occupants []*Patient
}

func NewRoom(number, capacity int) (*Room, error) {
	if capacity < 1 {
		return nil, fmt.Errorf("room %d: %w", number, ErrInvalidCapacity)
	}
	return &Room{
		number:    number,
		capacity:  capacity,
		occupants: make([]*Patient, 0, capacity),
	}, nil
}

func (r *Room) Number() int           { return r.number }
func (r *Room) Capacity() int         { return r.capacity }
func (r *Room) CurrentOccupancy() int { return len(r.occupants) }
func (r *Room) State() RoomState {
	n := len(r.occupants)
	switch {
	case n >= r.capacity:
		return RoomStateOccupied
	case n*100 >= r.capacity*alertThresholdPercent:
		return RoomStateAlerting
	default:
		return RoomStateAvailable
	}
}

func (r *Room) IsAvailable() bool {
	return len(r.occupants) < r.capacity
}

// Get a patient into the room
func (r *Room) Occupy(p *Patient) error {
	if p == nil {
		return ErrNilPatient
	}
	if !r.IsAvailable() {
		return fmt.Errorf("room %d: %w", r.number, ErrRoomFull)
	}
	if p.assignedRoom != nil {
		return fmt.Errorf("patient %s: %w", p.ID(), ErrPatientAlreadyInRoom)
	}
	r.occupants = append(r.occupants, p)
	p.setAssignedRoom(r)
	return nil
}

// Get a patient out of the room
func (r *Room) Release(p *Patient) error {
	if p == nil {
		return ErrNilPatient
	}
	for i, occupant := range r.occupants {
		if occupant.ID() == p.ID() {
			r.occupants = append(r.occupants[:i], r.occupants[i+1:]...)
			p.clearAssignedRoom()
			return nil
		}
	}
	return fmt.Errorf("room %d, patient %s: %w", r.number, p.ID(), ErrPatientNotInRoom)
}

func (r *Room) Occupants() []*Patient {
	out := make([]*Patient, len(r.occupants))
	copy(out, r.occupants)
	return out
}

func (r *Room) String() string {
	return fmt.Sprintf("Room %d | %d/%d | %s", r.number, len(r.occupants), r.capacity, r.State())
}
