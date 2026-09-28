package hospital

type PatientState int

const (
	PatientStateAwake PatientState = iota
	PatientStateDrowsy
	PatientStateAsleep
)

func (s PatientState) String() string {
	switch s {
	case PatientStateAwake:
		return "Awake"
	case PatientStateDrowsy:
		return "Drowsy"
	case PatientStateAsleep:
		return "Asleep"
	default:
		return "Unknown"
	}
}

type NarcolepsyLevel int

const (
	NarcolepsyLevelMild NarcolepsyLevel = iota
	NarcolepsyLevelModerate
	NarcolepsyLevelSevere
)

func (n NarcolepsyLevel) String() string {
	switch n {
	case NarcolepsyLevelMild:
		return "Mild"
	case NarcolepsyLevelModerate:
		return "Moderate"
	case NarcolepsyLevelSevere:
		return "Severe"
	default:
		return "Unknown"
	}
}

type RoomState int

const (
	RoomStateAvailable RoomState = iota
	RoomStateOccupied
	RoomStateAlerting
)

func (r RoomState) String() string {
	switch r {
	case RoomStateAvailable:
		return "Available"
	case RoomStateOccupied:
		return "Occupied"
	case RoomStateAlerting:
		return "Alerting"
	default:
		return "Unknown"
	}
}
