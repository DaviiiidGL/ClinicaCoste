package storage

import (
	"time"

	"ClinicaCoste/internal/hospital"
)

type EpisodeEntry struct {
	ID           string    `json:"id"`
	Timestamp    time.Time `json:"timestamp"`
	PatientID    string    `json:"patientId"`
	PatientName  string    `json:"patientName"`
	AttenderID   string    `json:"attenderId,omitempty"`
	AttenderName string    `json:"attenderName,omitempty"`
	Location     string    `json:"location"`
	RoomNumber   *int      `json:"roomNumber,omitempty"`
}

type Repository interface {
	SaveHistory(entries []EpisodeEntry) error
	LoadHistory() ([]EpisodeEntry, error)
}

func EntriesFromRecords(records []hospital.EpisodeRecord) []EpisodeEntry {
	out := make([]EpisodeEntry, 0, len(records))
	for _, r := range records {
		e := EpisodeEntry{
			ID:        r.ID(),
			Timestamp: r.Timestamp(),
			Location:  r.Location(),
		}
		if p := r.Patient(); p != nil {
			e.PatientID, e.PatientName = p.ID(), p.Name()
		}
		if a := r.AttendingDoctor(); a != nil {
			e.AttenderID, e.AttenderName = a.ID(), a.Name()
		}
		if room := r.AssignedRoom(); room != nil {
			n := room.Number()
			e.RoomNumber = &n
		}
		out = append(out, e)
	}
	return out
}
