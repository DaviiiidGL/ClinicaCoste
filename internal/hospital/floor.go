package hospital

type Floor struct {
	number int
	name   string
	rooms  []*Room
}

func NewFloor(number int, name string, rooms []*Room) *Floor {
	return &Floor{
		number: number,
		name:   name,
		rooms:  rooms,
	}
}

func (f *Floor) AddRoom(room *Room) error {
	if room == nil {
		return ErrNilRoom
	}
	for _, r := range f.rooms {
		if r.Number() == room.Number() {
			return ErrRoomAlreadyRegistered
		}
	}
	f.rooms = append(f.rooms, room)
	return nil
}

func (f *Floor) Number() int    { return f.number }
func (f *Floor) Name() string   { return f.name }
func (f *Floor) Rooms() []*Room { return f.rooms }

type PatientView struct {
	ID           string
	ShortID      string
	Name         string
	State        string
	Location     string
	AssignedRoom int // si no tiene 0
	IsCrisis     bool
}
