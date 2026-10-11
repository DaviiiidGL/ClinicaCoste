package scenario

import (
	"errors"
	"fmt"

	"ClinicaCoste/internal/config"
	"ClinicaCoste/internal/hospital"
	"ClinicaCoste/internal/metrics"
)

type Scenario struct {
	cfg      config.ScenarioConfig
	hospital *hospital.Hospital
	patients map[string]*hospital.Patient
}

func Build(cfg config.Config) (*Scenario, error) {
	h, err := hospital.NewHospital(cfg.HospitalName)
	if err != nil {
		return nil, err
	}
	for _, r := range cfg.Rooms {
		if _, err := h.AddRoom(r.Number, r.Capacity); err != nil {
			return nil, fmt.Errorf("building room %d: %w", r.Number, err)
		}
	}
	for _, d := range cfg.Doctors {
		if err := h.HireDoctor(hospital.NewDoctor(d.Name, d.Age, d.Specialty)); err != nil {
			return nil, fmt.Errorf("hiring doctor %s: %w", d.Name, err)
		}
	}
	for _, o := range cfg.Orderlies {
		if err := h.AddAttender(hospital.NewOrderly(o.Name, o.Age)); err != nil {
			return nil, fmt.Errorf("hiring orderly %s: %w", o.Name, err)
		}
	}
	for _, n := range cfg.Nurses {
		if err := h.AddAttender(hospital.NewNurse(n.Name, n.Age)); err != nil {
			return nil, fmt.Errorf("hiring nurse %s: %w", n.Name, err)
		}
	}

	patients := make(map[string]*hospital.Patient, len(cfg.Patients))
	for _, pc := range cfg.Patients {
		level, err := hospital.ParseNarcolepsyLevel(pc.Level)
		if err != nil {
			return nil, fmt.Errorf("patient %s: %w", pc.Name, err)
		}
		p := hospital.NewPatient(pc.Name, pc.Age, level)
		if err := h.AdmitPatient(p); err != nil {
			return nil, fmt.Errorf("admitting patient %s: %w", pc.Name, err)
		}
		patients[pc.Name] = p
	}
	return &Scenario{cfg: cfg.Scenario, hospital: h, patients: patients}, nil
}

func (s *Scenario) Hospital() *hospital.Hospital { return s.hospital }

type AttackOutcome struct {
	Patient  *hospital.Patient
	Location string
	Room     *hospital.Room
	Err      error
}

type DoctorEpisodes struct {
	Doctor   *hospital.Doctor
	Episodes []hospital.EpisodeRecord
}

type Metrics struct {
	Occupancy             metrics.Occupancy
	AvgEpisodesPerPatient float64
	BusiestDoctor         *hospital.Doctor
	BusiestDoctorEpisodes int
	Heatmap               []metrics.LocationCount
}

type Result struct {
	HospitalName string

	Attacks []AttackOutcome

	HallwayBefore []*hospital.Patient
	Woken         *hospital.Patient
	ReleasedRoom  *hospital.Room

	Reassigned     *hospital.Patient
	ReassignedRoom *hospital.Room
	ReassignErr    error

	HallwayAfter []*hospital.Patient
	ByDoctor     []DoctorEpisodes
	Severe       []hospital.SeverityEntry
	Metrics      Metrics
	History      []hospital.EpisodeRecord
}

func (s *Scenario) Run() (*Result, error) {
	res := &Result{HospitalName: s.hospital.Name()}

	for _, a := range s.cfg.Attacks {
		p, ok := s.patients[a.Patient]
		if !ok {
			return nil, fmt.Errorf("attack: unknown patient %q", a.Patient)
		}
		room, err := s.hospital.ReportSleepAttack(p, a.Location)
		if err != nil && !errors.Is(err, hospital.ErrNoRoomAvailable) {
			return nil, fmt.Errorf("attack of %s at %q: %w", p.Name(), a.Location, err)
		}
		res.Attacks = append(res.Attacks, AttackOutcome{Patient: p, Location: a.Location, Room: room, Err: err})
	}
	res.HallwayBefore = s.hospital.PatientsInHallway()

	woken, ok := s.patients[s.cfg.WakePatient]
	if !ok {
		return nil, fmt.Errorf("wake: unknown patient %q", s.cfg.WakePatient)
	}
	released := woken.AssignedRoom()
	if released == nil {
		return nil, fmt.Errorf("wake: %s has no room to release: %w", woken.Name(), hospital.ErrPatientNotInRoom)
	}
	if err := s.hospital.WakePatient(woken); err != nil {
		return nil, fmt.Errorf("waking %s: %w", woken.Name(), err)
	}
	res.Woken, res.ReleasedRoom = woken, released

	res.Reassigned, res.ReassignedRoom, res.ReassignErr = s.hospital.AssignNextWaiting()

	res.HallwayAfter = s.hospital.PatientsInHallway()
	for _, d := range s.hospital.Doctors() {
		res.ByDoctor = append(res.ByDoctor, DoctorEpisodes{Doctor: d, Episodes: d.MyEpisodes()})
	}
	res.Severe = s.hospital.SevereReportOrdered()
	res.History = s.hospital.History()

	res.Metrics.Occupancy = metrics.RoomOccupancy(s.hospital)
	res.Metrics.AvgEpisodesPerPatient = metrics.AverageEpisodesPerPatient(s.hospital)
	if d, n, ok := metrics.BusiestDoctor(s.hospital); ok {
		res.Metrics.BusiestDoctor, res.Metrics.BusiestDoctorEpisodes = d, n
	}
	res.Metrics.Heatmap = metrics.LocationHeatmap(s.hospital)
	return res, nil
}
