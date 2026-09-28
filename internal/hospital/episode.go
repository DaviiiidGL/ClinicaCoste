package hospital

import (
	"fmt"
	"time"
	"uuid"
)

type EpisodeRecord struct {
	id              string
	timestamp       time.Time
	patient         *Patient
	attendingDoctor Attender
	location        string
	assignedRoom    *Room
}

func NewEpisodeRecord(p *Patient, by Attender, location string, room *Room) EpisodeRecord {
	return EpisodeRecord{
		id:              uuid.New().String(),
		timestamp:       time.Now(),
		patient:         p,
		attendingDoctor: by,
		location:        location,
		assignedRoom:    room,
	}
}
func (e EpisodeRecord) ID() string                { return e.id }
func (e EpisodeRecord) Timestamp() time.Time      { return e.timestamp }
func (e EpisodeRecord) Patient() *Patient         { return e.patient }
func (e EpisodeRecord) AttendingDoctor() Attender { return e.attendingDoctor }
func (e EpisodeRecord) Location() string          { return e.location }
func (e EpisodeRecord) AssignedRoom() *Room       { return e.assignedRoom }

func (e EpisodeRecord) Summary() string {
	patient := "unknown"
	if e.patient != nil {
		patient = shortID(e.patient.ID())
	}
	room := "hallway (no room)"
	if e.assignedRoom != nil {
		room = fmt.Sprintf("room %d", e.assignedRoom.Number())
	}
	by := "unassigned"
	if e.attendingDoctor != nil {
		by = e.attendingDoctor.Name()
	}
	return fmt.Sprintf("%s  P-%s  %s -> %s  (%s)",
		e.timestamp.Format("2006-01-02 15:04"), patient, e.location, room, by)
}

func (e EpisodeRecord) String() string { return e.Summary() }
