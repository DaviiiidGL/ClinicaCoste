package hospital

import "fmt"

type Room struct {
	number    int
	capacity  int
	state     RoomState
	occupants []*Patient
}

func NewRoom(number int, capacity int) (*Room, error) {
	if capacity < 1 {
		return nil, fmt.Errorf("room: %d: %w", number, ErrInvalidCapacity)
	}

	return &Room{
		number:    number,
		capacity:  capacity,
		occupants: make([]*Patient, 0, capacity),
		state:     RoomStateAvailable,
	}, nil
}

func (r *Room) Number() int {
	return r.number
}

func (r *Room) Capacity() int {
	return r.capacity
}

func (r *Room) CurrentCapacity() int {
	return len(r.occupants)
}

func (r *Room) State() RoomState {
	return r.state
}

func (r *Room) IsAvailable() bool {
	return len(r.occupants) < r.capacity
}

// Get a patient into the room
func (r *Room) Occupy(patient *Patient) error {
	if !r.IsAvailable() {
		return fmt.Errorf("room %d: %w", r.number, ErrRoomFull)
	}

	r.occupants = append(r.occupants, patient)
	patient.SetAssignedRoom(r)

	if len(r.occupants) == 1 {
		r.state = RoomStateOccupied
	}

	return nil
}
