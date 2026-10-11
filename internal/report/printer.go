package report

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"ClinicaCoste/internal/hospital"
	"ClinicaCoste/internal/scenario"
)

func Print(w io.Writer, r *scenario.Result) error {
	p := &printer{w: w}

	p.f("════════════════════════════════════════════════════════════\n")
	p.f(" %s\n", r.HospitalName)
	p.f("════════════════════════════════════════════════════════════\n")

	p.printScript(r)
	p.print51(r)
	p.print52(r)
	p.print53(r)
	p.print54(r)
	p.printMetrics(r)
	return p.err
}

type printer struct {
	w   io.Writer
	err error
}

func (p *printer) f(format string, args ...any) {
	if p.err != nil {
		return
	}
	_, p.err = fmt.Fprintf(p.w, format, args...)
}

func (p *printer) printScript(r *scenario.Result) {
	p.f("\n== Guion: siestas inesperadas ==\n")
	for i, a := range r.Attacks {
		p.f("  %d. Siesta de %s en «%s»: %s\n", i+1, a.Patient.Name(), a.Location, roomOutcome(a.Room, a.Err))
	}
	if r.Woken != nil {
		p.f("  → %s despierta y libera la sala %d\n", r.Woken.Name(), r.ReleasedRoom.Number())
	}
	switch {
	case r.ReassignErr != nil:
		p.f("  → Reasignación: %s\n", readable(r.ReassignErr))
	case r.Reassigned != nil:
		p.f("  → %s deja el pasillo y pasa a la sala %d\n", r.Reassigned.Name(), r.ReassignedRoom.Number())
	}
}

func (p *printer) print51(r *scenario.Result) {
	p.f("\n== 5.1 Pacientes dormidos en el pasillo ==\n")
	p.f("  Tras las cuatro siestas:\n")
	p.patientList(r.HallwayBefore)
	p.f("  Al final del guion:\n")
	p.patientList(r.HallwayAfter)
}

func (p *printer) patientList(ps []*hospital.Patient) {
	if len(ps) == 0 {
		p.f("    (ninguno)\n")
		return
	}
	for _, pt := range ps {
		p.f("    P-%s  %s (%s)  — %s\n", short(pt.ID()), pt.Name(), pt.Level(), pt.Location())
	}
}

func (p *printer) print52(r *scenario.Result) {
	p.f("\n== 5.2 Episodios atendidos por cada médico ==\n")
	for _, d := range r.ByDoctor {
		p.f("  %s (%s)\n", d.Doctor.Name(), d.Doctor.Specialty())
		if len(d.Episodes) == 0 {
			p.f("    (sin episodios)\n")
		}
		for _, e := range d.Episodes {
			p.f("    %s  P-%s  %s → %s\n",
				e.Timestamp().Format("2006-01-02 15:04"), short(e.Patient().ID()), e.Location(), roomText(e.AssignedRoom()))
		}
	}
}

func (p *printer) print53(r *scenario.Result) {
	p.f("\n== 5.3 Asignación de sala ==\n")
	for _, a := range r.Attacks {
		p.f("  %s: %s\n", a.Patient.Name(), roomOutcome(a.Room, a.Err))
	}
	if r.ReassignErr == nil && r.Reassigned != nil {
		p.f("  %s (reintento al liberarse la sala): sala %d\n", r.Reassigned.Name(), r.ReassignedRoom.Number())
	}
}

func (p *printer) print54(r *scenario.Result) {
	p.f("\n== 5.4 Reporte de severidad (pacientes Severe) ==\n")
	if len(r.Severe) == 0 {
		p.f("  (no hay pacientes Severe)\n")
	}
	for _, e := range r.Severe {
		p.f("  P-%s  %-20s %d %s hoy\n", short(e.Patient.ID()), e.Patient.Name(), e.Episodes, plural(e.Episodes, "episodio", "episodios"))
	}
}

func (p *printer) printMetrics(r *scenario.Result) {
	m := r.Metrics
	p.f("\n== Métricas del día (extra) ==\n")
	p.f("  Ocupación de salas: %d/%d (%.0f%%)\n", m.Occupancy.Used, m.Occupancy.Capacity, m.Occupancy.Percent)
	p.f("  Episodios por paciente (promedio): %.2f\n", m.AvgEpisodesPerPatient)
	if m.BusiestDoctor != nil {
		p.f("  Médico más ocupado: %s (%d %s)\n", m.BusiestDoctor.Name(), m.BusiestDoctorEpisodes,
			plural(m.BusiestDoctorEpisodes, "episodio", "episodios"))
	} else {
		p.f("  Médico más ocupado: nadie ha atendido todavía\n")
	}
	p.f("  Mapa de calor de ubicaciones:\n")
	for _, l := range m.Heatmap {
		p.f("    %-40s %s %d\n", l.Location, strings.Repeat("█", l.Count), l.Count)
	}
}

func roomOutcome(room *hospital.Room, err error) string {
	if err != nil {
		return readable(err)
	}
	return fmt.Sprintf("sala %d", room.Number())
}

func roomText(room *hospital.Room) string {
	if room == nil {
		return "pasillo (sin sala)"
	}
	return fmt.Sprintf("sala %d", room.Number())
}

func readable(err error) string {
	switch {
	case errors.Is(err, hospital.ErrNoRoomAvailable):
		return "no hay sala disponible: el paciente se queda dormido en el pasillo"
	case errors.Is(err, hospital.ErrNoPatientWaiting):
		return "nadie esperaba en el pasillo"
	default:
		return err.Error()
	}
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

// acortar UUID
func short(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}
