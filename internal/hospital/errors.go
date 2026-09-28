package hospital

import "errors"

// Patient Errors
var (
	ErrPatientAlreadyAsleep   = errors.New("patient is already asleep")
	ErrPatientAlreadyAwake    = errors.New("patient is already awake")
	ErrNilPatient             = errors.New("patient cannot be nil")
	ErrPatientNotFound        = errors.New("patient not found")
	ErrPatientAlreadyAdmitted = errors.New("patient already admitted")
)

// Doctor Errors
var (
	ErrNilDoctor          = errors.New("doctor cannot be nil")
	ErrDoctorNotFound     = errors.New("doctor not found")
	ErrDoctorAlreadyHired = errors.New("doctor already hired")
)

// Room Errors
var (
	ErrRoomFull             = errors.New("room is full")
	ErrRoomNotFound         = errors.New("room not found")
	ErrPatientNotInRoom     = errors.New("patient is not in room")
	ErrPatientAlreadyInRoom = errors.New("patient is already in a room")
	ErrInvalidCapacity      = errors.New("room capacity must be at least 1")
	ErrNoRoomAvailable      = errors.New("no room available")
)

// Attender Errors
var (
	ErrNilAttender            = errors.New("attender cannot be nil")
	ErrNoAttenderAvailable    = errors.New("no attender available")
	ErrAttenderCannotDiagnose = errors.New("this attender cannot diagnose patients")
)

// Episode Record Errors
var (
	ErrEpisodeNotFound = errors.New("episode record not found")
)

// General Errors
var (
	ErrHospitalNameRequired = errors.New("hospital name is required")
	ErrInvalidInput         = errors.New("invalid input")
)
