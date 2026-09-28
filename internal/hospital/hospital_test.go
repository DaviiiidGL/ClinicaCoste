package hospital

import (
	"errors"
	"testing"
)

// --- helpers ---

// newTestHospital crea un hospital con una sala por cada capacidad indicada
// (numeradas 101, 102, ...).
func newTestHospital(t *testing.T, capacities ...int) *Hospital {
	t.Helper()
	h, err := NewHospital("Hospital de los Costeños con Narcolepsia")
	if err != nil {
		t.Fatalf("NewHospital: %v", err)
	}
	for i, c := range capacities {
		if _, err := h.AddRoom(101+i, c); err != nil {
			t.Fatalf("AddRoom: %v", err)
		}
	}
	return h
}

func admit(t *testing.T, h *Hospital, name string, level NarcolepsyLevel) *Patient {
	t.Helper()
	p := NewPatient(name, 30, level)
	if err := h.AdmitPatient(p); err != nil {
		t.Fatalf("AdmitPatient: %v", err)
	}
	return p
}

// --- construcción y altas ---

func TestNewHospital_NameRequired(t *testing.T) {
	for _, name := range []string{"", "   "} {
		if _, err := NewHospital(name); !errors.Is(err, ErrHospitalNameRequired) {
			t.Errorf("name %q: got %v", name, err)
		}
	}
}

func TestAdmitPatient(t *testing.T) {
	h := newTestHospital(t)
	p := NewPatient("Yeimy", 29, NarcolepsyLevelSevere)

	if err := h.AdmitPatient(p); err != nil {
		t.Fatalf("first admit: %v", err)
	}
	if err := h.AdmitPatient(p); !errors.Is(err, ErrPatientAlreadyAdmitted) {
		t.Errorf("duplicate: got %v", err)
	}
	if err := h.AdmitPatient(nil); !errors.Is(err, ErrNilPatient) {
		t.Errorf("nil: got %v", err)
	}
	if len(h.Patients()) != 1 {
		t.Errorf("want 1 patient, got %d", len(h.Patients()))
	}
}

func TestHireDoctor(t *testing.T) {
	h := newTestHospital(t)
	d := NewDoctor("Karen", 38, "Neurology")

	if err := h.HireDoctor(d); err != nil {
		t.Fatalf("hire: %v", err)
	}
	if err := h.HireDoctor(d); !errors.Is(err, ErrDoctorAlreadyHired) {
		t.Errorf("duplicate: got %v", err)
	}
	if err := h.HireDoctor(nil); !errors.Is(err, ErrNilDoctor) {
		t.Errorf("nil: got %v", err)
	}
}

func TestAddRoom(t *testing.T) {
	h := newTestHospital(t)

	if _, err := h.AddRoom(101, 1); err != nil {
		t.Fatalf("add: %v", err)
	}
	if _, err := h.AddRoom(101, 2); !errors.Is(err, ErrRoomAlreadyRegistered) {
		t.Errorf("duplicate number: got %v", err)
	}
	if _, err := h.AddRoom(102, 0); !errors.Is(err, ErrInvalidCapacity) {
		t.Errorf("invalid capacity: got %v", err)
	}
	if len(h.Rooms()) != 1 {
		t.Errorf("want 1 room, got %d", len(h.Rooms()))
	}
}

// --- AssignRoom (sección 5.3 del enunciado) ---

func TestAssignRoom_WithFreeRoom(t *testing.T) {
	h := newTestHospital(t, 1, 1)
	p := admit(t, h, "Wilfrido", NarcolepsyLevelSevere)
	_ = p.SufferSleepAttack("cafeteria")

	room, err := h.AssignRoom(p)
	if err != nil {
		t.Fatalf("AssignRoom: %v", err)
	}
	if room.Number() != 101 {
		t.Errorf("must assign the FIRST available room, got %d", room.Number())
	}
	if room.State() != RoomStateOccupied {
		t.Errorf("room state = %v, want Occupied", room.State())
	}
	if !p.IsAsleepInBed() || p.AssignedRoom() != room {
		t.Errorf("patient must be asleep in bed: %v", p)
	}
}

func TestAssignRoom_NoRoomAvailable(t *testing.T) {
	h := newTestHospital(t, 1)
	first := admit(t, h, "A", NarcolepsyLevelMild)
	second := admit(t, h, "B", NarcolepsyLevelMild)
	_ = first.SufferSleepAttack("x")
	_ = second.SufferSleepAttack("hallway 2")
	if _, err := h.AssignRoom(first); err != nil {
		t.Fatalf("first: %v", err)
	}

	room, err := h.AssignRoom(second)

	if !errors.Is(err, ErrNoRoomAvailable) {
		t.Fatalf("want ErrNoRoomAvailable, got %v", err)
	}
	if room != nil {
		t.Error("must not invent a room")
	}
	if !second.IsAsleepInHallway() {
		t.Errorf("patient must stay in the hallway: %v", second)
	}
	if len(h.Rooms()) != 1 {
		t.Errorf("rooms changed: %d", len(h.Rooms()))
	}
}

func TestAssignRoom_Preconditions(t *testing.T) {
	h := newTestHospital(t, 1, 1)
	stranger := NewPatient("Not admitted", 30, NarcolepsyLevelMild)
	_ = stranger.SufferSleepAttack("x")
	awake := admit(t, h, "Awake", NarcolepsyLevelMild)
	seated := admit(t, h, "Seated", NarcolepsyLevelMild)
	_ = seated.SufferSleepAttack("x")
	_, _ = h.AssignRoom(seated)

	tests := []struct {
		name string
		p    *Patient
		want error
	}{
		{"nil", nil, ErrNilPatient},
		{"not admitted", stranger, ErrPatientNotFound},
		{"awake", awake, ErrPatientNotAsleep},
		{"already in a room", seated, ErrPatientAlreadyInRoom},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := h.AssignRoom(tc.p); !errors.Is(err, tc.want) {
				t.Errorf("got %v, want %v", err, tc.want)
			}
		})
	}
}

// --- ReportSleepAttack ---

func TestReportSleepAttack(t *testing.T) {
	h := newTestHospital(t, 1)
	a := admit(t, h, "A", NarcolepsyLevelMild)
	b := admit(t, h, "B", NarcolepsyLevelMild)

	room, err := h.ReportSleepAttack(a, "cafeteria")
	if err != nil || room == nil {
		t.Fatalf("A should get a room: %v", err)
	}

	room, err = h.ReportSleepAttack(b, "hallway 2")
	if !errors.Is(err, ErrNoRoomAvailable) || room != nil {
		t.Fatalf("B: want ErrNoRoomAvailable, got room=%v err=%v", room, err)
	}
	if !b.IsAsleepInHallway() || b.Location() != "hallway 2" {
		t.Errorf("B must be asleep in the hallway: %v", b)
	}

	if _, err := h.ReportSleepAttack(b, "again"); !errors.Is(err, ErrPatientAlreadyAsleep) {
		t.Errorf("second attack: got %v", err)
	}
	if _, err := h.ReportSleepAttack(nil, "x"); !errors.Is(err, ErrNilPatient) {
		t.Errorf("nil: got %v", err)
	}
	stranger := NewPatient("Stranger", 30, NarcolepsyLevelMild)
	if _, err := h.ReportSleepAttack(stranger, "x"); !errors.Is(err, ErrPatientNotFound) {
		t.Errorf("not admitted: got %v", err)
	}
}

// --- PatientsInHallway (consulta 5.1) ---

func TestPatientsInHallway(t *testing.T) {
	h := newTestHospital(t, 1)
	inBed := admit(t, h, "InBed", NarcolepsyLevelSevere)
	hall1 := admit(t, h, "Hall1", NarcolepsyLevelMild)
	awake := admit(t, h, "Awake", NarcolepsyLevelMild)
	hall2 := admit(t, h, "Hall2", NarcolepsyLevelSevere)

	_, _ = h.ReportSleepAttack(inBed, "a")
	_, _ = h.ReportSleepAttack(hall1, "b")
	_, _ = h.ReportSleepAttack(hall2, "c")

	got := h.PatientsInHallway()
	if len(got) != 2 || got[0] != hall1 || got[1] != hall2 {
		t.Fatalf("want [Hall1 Hall2] in admission order, got %v", got)
	}
	for _, p := range got {
		if p == inBed || p == awake {
			t.Errorf("%v should not be in the hallway", p)
		}
	}
}

// --- Despertar y prioridad ---

func TestWakePatient_ReleasesRoom(t *testing.T) {
	h := newTestHospital(t, 1)
	p := admit(t, h, "A", NarcolepsyLevelMild)
	room, _ := h.ReportSleepAttack(p, "x")

	if err := h.WakePatient(p); err != nil {
		t.Fatalf("WakePatient: %v", err)
	}
	if p.State() != PatientStateAwake || p.AssignedRoom() != nil || room.State() != RoomStateAvailable {
		t.Errorf("patient/room not cleaned up: %v / %v", p, room)
	}
	if err := h.WakePatient(p); !errors.Is(err, ErrPatientAlreadyAwake) {
		t.Errorf("wake twice: got %v", err)
	}
}

func TestAssignNextWaiting_ServesSevereBeforeEarlierMild(t *testing.T) {
	h := newTestHospital(t, 1)
	holder := admit(t, h, "Holder", NarcolepsyLevelMild)
	mild := admit(t, h, "Mild", NarcolepsyLevelMild)
	severe := admit(t, h, "Severe", NarcolepsyLevelSevere)

	_, _ = h.ReportSleepAttack(holder, "a") // toma la única sala
	_, _ = h.ReportSleepAttack(mild, "b")   // espera (admitido antes que severe)
	_, _ = h.ReportSleepAttack(severe, "c") // espera

	if _, _, err := h.AssignNextWaiting(); !errors.Is(err, ErrNoRoomAvailable) {
		t.Fatalf("no free room yet: got %v", err)
	}

	_ = h.WakePatient(holder)

	next, room, err := h.AssignNextWaiting()
	if err != nil {
		t.Fatalf("AssignNextWaiting: %v", err)
	}
	if next != severe || room.Number() != 101 {
		t.Errorf("Severe must go first, got %v in %v", next, room)
	}
	if !mild.IsAsleepInHallway() {
		t.Error("Mild must keep waiting")
	}
}

func TestAssignNextWaiting_FIFOAmongEquals(t *testing.T) {
	h := newTestHospital(t, 1)
	holder := admit(t, h, "Holder", NarcolepsyLevelMild)
	first := admit(t, h, "First", NarcolepsyLevelModerate)
	second := admit(t, h, "Second", NarcolepsyLevelModerate)

	_, _ = h.ReportSleepAttack(holder, "a")
	_, _ = h.ReportSleepAttack(second, "b")
	_, _ = h.ReportSleepAttack(first, "c")
	_ = h.WakePatient(holder)

	next, _, err := h.AssignNextWaiting()
	if err != nil || next != first {
		t.Errorf("equal priority: admission order wins, got %v (%v)", next, err)
	}
}

func TestAssignNextWaiting_NobodyWaiting(t *testing.T) {
	h := newTestHospital(t, 1)
	admit(t, h, "Awake", NarcolepsyLevelSevere)

	if _, _, err := h.AssignNextWaiting(); !errors.Is(err, ErrNoPatientWaiting) {
		t.Errorf("got %v", err)
	}
}

// --- Historial ---

func TestRegisterEpisode(t *testing.T) {
	h := newTestHospital(t, 1)
	p := admit(t, h, "A", NarcolepsyLevelMild)
	by := fakeAttender{"S-1", "Dr. Karen"}
	room, _ := h.ReportSleepAttack(p, "cafeteria")

	e := NewEpisodeRecord(p, by, "cafeteria", room)
	if err := h.RegisterEpisode(e); err != nil {
		t.Fatalf("RegisterEpisode: %v", err)
	}
	if hist := h.History(); len(hist) != 1 || hist[0].ID() != e.ID() {
		t.Fatalf("history = %v", hist)
	}

	hist := h.History()
	hist[0] = EpisodeRecord{}
	if h.History()[0].ID() != e.ID() {
		t.Error("History must return a copy")
	}

	stranger := NewPatient("Stranger", 30, NarcolepsyLevelMild)
	if err := h.RegisterEpisode(NewEpisodeRecord(stranger, by, "x", nil)); !errors.Is(err, ErrPatientNotFound) {
		t.Errorf("not admitted: got %v", err)
	}
	if err := h.RegisterEpisode(NewEpisodeRecord(nil, by, "x", nil)); !errors.Is(err, ErrNilPatient) {
		t.Errorf("nil patient: got %v", err)
	}
}

// --- El escenario de la sección 6, a nivel de dominio ---

func TestDemoScenarioFlow(t *testing.T) {
	h := newTestHospital(t, 1, 1, 1)
	ps := []*Patient{
		admit(t, h, "P1", NarcolepsyLevelMild),
		admit(t, h, "P2", NarcolepsyLevelSevere),
		admit(t, h, "P3", NarcolepsyLevelModerate),
		admit(t, h, "P4", NarcolepsyLevelSevere),
		admit(t, h, "P5", NarcolepsyLevelMild),
	}

	for i, loc := range []string{"cafeteria", "radiology queue", "pharmacy", "hallway 2"} {
		_, err := h.ReportSleepAttack(ps[i], loc)
		if i < 3 && err != nil {
			t.Fatalf("patient %d should get a room: %v", i, err)
		}
		if i == 3 && !errors.Is(err, ErrNoRoomAvailable) {
			t.Fatalf("4th patient must fail with ErrNoRoomAvailable, got %v", err)
		}
	}
	if got := h.PatientsInHallway(); len(got) != 1 || got[0] != ps[3] {
		t.Fatalf("only P4 should wait, got %v", got)
	}

	if err := h.WakePatient(ps[0]); err != nil {
		t.Fatalf("wake P1: %v", err)
	}
	next, room, err := h.AssignNextWaiting()
	if err != nil || next != ps[3] || room.Number() != 101 {
		t.Fatalf("P4 should take room 101, got %v %v %v", next, room, err)
	}
	if len(h.PatientsInHallway()) != 0 {
		t.Error("hallway should be empty")
	}
}
