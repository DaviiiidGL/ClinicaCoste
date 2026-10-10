package hospital

import (
	"cmp"
	"slices"
	"strings"
	"time"
)

type SeverityEntry struct {
	Patient  *Patient
	Episodes int
}

func (h *Hospital) SevereReport() map[*Patient]int {
	return h.severeReportOn(h.now())
}

func (h *Hospital) SevereReportOrdered() []SeverityEntry {
	report := h.SevereReport()
	out := make([]SeverityEntry, 0, len(report))
	for p, n := range report {
		out = append(out, SeverityEntry{Patient: p, Episodes: n})
	}
	slices.SortFunc(out, func(a, b SeverityEntry) int {
		if c := cmp.Compare(b.Episodes, a.Episodes); c != 0 {
			return c
		}
		if c := strings.Compare(a.Patient.Name(), b.Patient.Name()); c != 0 {
			return c
		}
		return strings.Compare(a.Patient.ID(), b.Patient.ID())
	})
	return out
}

func (h *Hospital) severeReportOn(day time.Time) map[*Patient]int {
	perPatient := make(map[string]int)
	for _, e := range h.history {
		if p := e.Patient(); p != nil && sameDay(e.Timestamp(), day) {
			perPatient[p.ID()]++
		}
	}

	report := make(map[*Patient]int)
	for _, p := range h.patients {
		if p.Level() == NarcolepsyLevelSevere {
			report[p] = perPatient[p.ID()]
		}
	}
	return report
}

func sameDay(a, b time.Time) bool {
	ay, am, ad := a.In(b.Location()).Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}
