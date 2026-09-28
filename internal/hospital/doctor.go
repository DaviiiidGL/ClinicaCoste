package hospital

type Doctor struct {
	Person

	specialty string
	patients  []*Patient
	episodes  []EpisodeRecord
}

func NewDoctor(name string, age int, speciality string) *Doctor {
	return &Doctor{
		Person: NewPerson(name, age), specialty: speciality,
	}
}

func (d *Doctor) Specialty() string { return d.specialty }

func (d *Doctor) Patients() []*Patient {
	out := make([]*Patient, len(d.patients))
	copy(out, d.patients)
	return out
}

func (d *Doctor) MyEpisodes() []EpisodeRecord {
	out := make([]EpisodeRecord, len(d.episodes))
	copy(out, d.episodes)
	return out
}

func (d *Doctor) addPatient(p *Patient)         { d.patients = append(d.patients, p) }
func (d *Doctor) recordEpisode(e EpisodeRecord) { d.episodes = append(d.episodes, e) }
