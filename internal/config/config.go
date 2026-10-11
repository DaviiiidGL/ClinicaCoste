package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"ClinicaCoste/internal/hospital"
)

var ErrInvalidConfig = errors.New("invalid config")

type Config struct {
	HospitalName string          `json:"hospitalName"`
	Rooms        []RoomConfig    `json:"rooms"`
	Doctors      []DoctorConfig  `json:"doctors"`
	Orderlies    []StaffConfig   `json:"orderlies"`
	Nurses       []StaffConfig   `json:"nurses"`
	Patients     []PatientConfig `json:"patients"`
	Scenario     ScenarioConfig  `json:"scenario"`
	Storage      StorageConfig   `json:"storage"`
}

type RoomConfig struct {
	Number   int `json:"number"`
	Capacity int `json:"capacity"`
}

type DoctorConfig struct {
	Name      string `json:"name"`
	Age       int    `json:"age"`
	Specialty string `json:"specialty"`
}

type StaffConfig struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type PatientConfig struct {
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Level string `json:"level"`
}

type ScenarioConfig struct {
	Attacks     []AttackConfig `json:"attacks"`
	WakePatient string         `json:"wakePatient"`
}

type AttackConfig struct {
	Patient  string `json:"patient"`
	Location string `json:"location"`
}

type StorageConfig struct {
	HistoryPath string `json:"historyPath"`
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("reading config file %s: %w", path, err)
	}
	cfg, err := Parse(data)
	if err != nil {
		return Config{}, fmt.Errorf("config file %s: %w", path, err)
	}
	return cfg, nil
}

func Parse(data []byte) (Config, error) {
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	var cfg Config
	if err := dec.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("%w: parsing JSON: %w", ErrInvalidConfig, err)
	}
	if dec.More() {
		return Config{}, fmt.Errorf("%w: unexpected data after the JSON object", ErrInvalidConfig)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	var problems []error
	add := func(format string, args ...any) {
		problems = append(problems, fmt.Errorf("%w: "+format, append([]any{ErrInvalidConfig}, args...)...))
	}
	person := func(section string, i int, name string, age int) {
		if strings.TrimSpace(name) == "" {
			add("%s[%d]: name is required", section, i)
		}
		if age <= 0 {
			add("%s[%d]: age must be positive, got %d", section, i, age)
		}
	}

	if strings.TrimSpace(c.HospitalName) == "" {
		add("hospitalName is required")
	}

	if len(c.Rooms) == 0 {
		add("at least one room is required")
	}
	seenRooms := make(map[int]bool)
	for i, r := range c.Rooms {
		if r.Number <= 0 {
			add("rooms[%d]: number must be positive, got %d", i, r.Number)
		}
		if r.Capacity < 1 {
			add("rooms[%d]: capacity must be at least 1, got %d", i, r.Capacity)
		}
		if seenRooms[r.Number] {
			add("rooms[%d]: duplicate room number %d", i, r.Number)
		}
		seenRooms[r.Number] = true
	}

	for i, d := range c.Doctors {
		person("doctors", i, d.Name, d.Age)
		if strings.TrimSpace(d.Specialty) == "" {
			add("doctors[%d]: specialty is required", i)
		}
	}
	for i, o := range c.Orderlies {
		person("orderlies", i, o.Name, o.Age)
	}
	for i, n := range c.Nurses {
		person("nurses", i, n.Name, n.Age)
	}

	if len(c.Patients) == 0 {
		add("at least one patient is required")
	}
	patients := make(map[string]bool)
	for i, p := range c.Patients {
		person("patients", i, p.Name, p.Age)
		if _, err := hospital.ParseNarcolepsyLevel(p.Level); err != nil {
			add("patients[%d]: level must be Mild, Moderate or Severe, got %q", i, p.Level)
		}
		if patients[p.Name] {
			add("patients[%d]: duplicate name %q (the scenario refers to patients by name)", i, p.Name)
		}
		patients[p.Name] = true
	}

	if len(c.Scenario.Attacks) == 0 {
		add("scenario.attacks: at least one attack is required")
	}
	attackers := make(map[string]bool)
	for i, a := range c.Scenario.Attacks {
		if !patients[a.Patient] {
			add("scenario.attacks[%d]: unknown patient %q", i, a.Patient)
		}
		if strings.TrimSpace(a.Location) == "" {
			add("scenario.attacks[%d]: location is required", i)
		}
		if attackers[a.Patient] {
			add("scenario.attacks[%d]: patient %q appears twice (a sleeping patient can't fall asleep again)", i, a.Patient)
		}
		attackers[a.Patient] = true
	}
	if !attackers[c.Scenario.WakePatient] {
		add("scenario.wakePatient: %q must be one of the patients that fall asleep", c.Scenario.WakePatient)
	}

	return errors.Join(problems...)
}
