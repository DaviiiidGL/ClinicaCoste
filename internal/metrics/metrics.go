package metrics

import (
	"ClinicaCoste/internal/hospital"
	"cmp"
	"slices"
	"strings"
)

type Occupancy struct {
	Used     int
	Capacity int
	Percent  float64
}

func RoomOccupancy(h *hospital.Hospital) Occupancy {
	var o Occupancy
	for _, r := range h.Rooms() {
		o.Used += r.CurrentOccupancy()
		o.Capacity += r.Capacity()
	}
	if o.Capacity > 0 {
		o.Percent = float64(o.Used) * 100 / float64(o.Capacity)
	}
	return o
}

func AverageEpisodesPerPatient(h *hospital.Hospital) float64 {
	patients := len(h.Patients())
	if patients == 0 {
		return 0
	}
	return float64(len(h.History())) / float64(patients)
}

func BusiestDoctor(h *hospital.Hospital) (doctor *hospital.Doctor, episodes int, ok bool) {
	if h == nil {
		return nil, 0, false
	}

	var bestDoc *hospital.Doctor
	maxEpisodes := 0

	for _, d := range h.Doctors() {
		n := len(d.MyEpisodes())
		if n == 0 {
			continue
		}
		if bestDoc == nil || n > maxEpisodes || (n == maxEpisodes && d.Name() < bestDoc.Name()) {
			bestDoc = d
			maxEpisodes = n
		}
	}
	if bestDoc == nil {
		return nil, 0, false
	}
	return bestDoc, maxEpisodes, true
}

type LocationCount struct {
	Location string
	Count    int
}

func LocationHeatmap(h *hospital.Hospital) []LocationCount {
	counts := make(map[string]int)
	for _, e := range h.History() {
		counts[strings.TrimSpace(e.Location())]++
	}

	out := make([]LocationCount, 0, len(counts))
	for loc, n := range counts {
		out = append(out, LocationCount{Location: loc, Count: n})
	}
	slices.SortFunc(out, func(a, b LocationCount) int {
		if c := cmp.Compare(b.Count, a.Count); c != 0 {
			return c
		}
		return strings.Compare(a.Location, b.Location)
	})
	return out
}
