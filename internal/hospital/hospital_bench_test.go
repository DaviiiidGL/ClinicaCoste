package hospital

import (
	"fmt"
	"testing"

	"ClinicaCoste/internal/hospital/clock"
)

// benchHospital arma un hospital con n salas de capacidad 1, admite n
// pacientes (repartidos entre los tres niveles) y los duerme TODOS para que
// AssignRoom/AssignNextWaiting trabajen sobre un pasillo realmente lleno, no
// sobre un hospital vacío. Usa un Attender sin turno fijo (NewShift(0, 24))
// para que dispatch() nunca tenga que descartar al personal, y evita el
// time.Sleep de 3-6s que Doctor.Attend hace en este código — un benchmark no
// debe pagar una espera artificial pensada para la demo en vivo.
type stubAttender struct{ id, name string }

func (s stubAttender) ID() string   { return s.id }
func (s stubAttender) Name() string { return s.name }
func (s stubAttender) Attend(p *Patient, location string) (EpisodeRecord, error) {
	return NewEpisodeRecord(p, s, location, p.AssignedRoom()), nil
}

func benchHospital(b *testing.B, n int) (*Hospital, []*Patient) {
	b.Helper()
	h, err := NewHospital("Bench", clock.RealClock{})
	if err != nil {
		b.Fatal(err)
	}
	if err := h.AddAttender(stubAttender{"S-1", "Staff"}); err != nil {
		b.Fatal(err)
	}
	for i := 0; i < n; i++ {
		if _, err := h.AddRoom(i+1, 1); err != nil {
			b.Fatal(err)
		}
	}
	levels := []NarcolepsyLevel{NarcolepsyLevelMild, NarcolepsyLevelModerate, NarcolepsyLevelSevere}
	patients := make([]*Patient, n)
	for i := 0; i < n; i++ {
		p := NewPatient(fmt.Sprintf("P-%d", i), 20+i%50, levels[i%3])
		if err := h.AdmitPatient(p); err != nil {
			b.Fatal(err)
		}
		patients[i] = p
	}
	return h, patients
}

var sizes = []int{10, 100, 1000}

// BenchmarkAssignRoom mide el costo de asignar sala a UN paciente recién
// dormido, con 0 salas libres (todas ocupadas menos una al final de la
// lista): es el peor caso real, porque recorre todas las salas antes de
// encontrar la libre.
func BenchmarkAssignRoom(b *testing.B) {
	for _, n := range sizes {
		b.Run(fmt.Sprintf("rooms=%d", n), func(b *testing.B) {
			h, patients := benchHospital(b, n)
			// Ocupar todas las salas menos la última.
			for i := 0; i < n-1; i++ {
				if _, err := h.ReportSleepAttack(patients[i], "x"); err != nil {
					b.Fatal(err)
				}
			}
			last := patients[n-1]

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = last.WakeUp()
				if room := last.AssignedRoom(); room != nil {
					_ = room.Release(last)
				}
				if err := last.SufferSleepAttack("x"); err != nil {
					b.Fatal(err)
				}
				if _, err := h.AssignRoom(last); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkAssignRoom_NoRoomAvailable mide el caso sin sala libre: recorre
// TODAS las salas y siempre falla. Es el camino que la sección 5.3 exige que
// no haga panic ni invente una sala, y el que más cuesta.
func BenchmarkAssignRoom_NoRoomAvailable(b *testing.B) {
	for _, n := range sizes {
		b.Run(fmt.Sprintf("rooms=%d", n), func(b *testing.B) {
			h, patients := benchHospital(b, n)
			for i := 0; i < n; i++ {
				if _, err := h.ReportSleepAttack(patients[i], "x"); err != nil {
					b.Fatal(err)
				}
			}
			extra := NewPatient("extra", 30, NarcolepsyLevelSevere)
			if err := h.AdmitPatient(extra); err != nil {
				b.Fatal(err)
			}

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = extra.WakeUp()
				if err := extra.SufferSleepAttack("x"); err != nil {
					b.Fatal(err)
				}
				if _, err := h.AssignRoom(extra); err == nil {
					b.Fatal("expected ErrNoRoomAvailable")
				}
			}
		})
	}
}

// BenchmarkPatientsInHallway mide la consulta 5.1 con el pasillo lleno (todos
// los pacientes admitidos, sin sala): recorre todos los pacientes del
// hospital, así que es O(pacientes) por llamada — el peor caso para esta query.
func BenchmarkPatientsInHallway(b *testing.B) {
	for _, n := range sizes {
		b.Run(fmt.Sprintf("patients=%d", n), func(b *testing.B) {
			h, patients := benchHospital(b, 0) // 0 salas: todos quedan en el pasillo
			for i := 0; i < n; i++ {
				p := NewPatient(fmt.Sprintf("P-%d", i), 30, NarcolepsyLevelMild)
				if err := h.AdmitPatient(p); err != nil {
					b.Fatal(err)
				}
				if err := p.SufferSleepAttack("x"); err != nil {
					b.Fatal(err)
				}
			}
			_ = patients

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = h.PatientsInHallway()
			}
		})
	}
}

// BenchmarkAssignNextWaiting mide el costo de ordenar por prioridad: copia el
// pasillo completo y lo ordena. Es O(pacientes en pasillo · log) por llamada.
// BenchmarkAssignNextWaiting mide el costo de ordenar por prioridad cuando hay
// n pacientes esperando en el pasillo (1 sola sala en todo el hospital, para
// que la cola de espera sea de tamaño n de verdad): copia el pasillo y lo
// ordena, O(n log n) por llamada.
func BenchmarkAssignNextWaiting(b *testing.B) {
	for _, n := range sizes {
		b.Run(fmt.Sprintf("waiting=%d", n), func(b *testing.B) {
			h, err := NewHospital("Bench", clock.RealClock{})
			if err != nil {
				b.Fatal(err)
			}
			if _, err := h.AddRoom(1, 1); err != nil {
				b.Fatal(err)
			}
			levels := []NarcolepsyLevel{NarcolepsyLevelMild, NarcolepsyLevelModerate, NarcolepsyLevelSevere}
			patients := make([]*Patient, n+1)
			for i := range patients {
				p := NewPatient(fmt.Sprintf("P-%d", i), 30, levels[i%3])
				if err := h.AdmitPatient(p); err != nil {
					b.Fatal(err)
				}
				if _, err := h.ReportSleepAttack(p, "x"); err != nil {
					// el primero consigue la única sala; el resto (n) queda esperando
				}
				patients[i] = p
			}
			if got := len(h.PatientsInHallway()); got != n {
				b.Fatalf("setup: %d patients waiting, want %d", got, n)
			}
			// P-0 tiene la única sala: liberarla antes de medir, para que la
			// primera llamada de la medición ya encuentre una sala libre
			// (igual que en cada vuelta siguiente).
			if err := h.WakePatient(patients[0]); err != nil {
				b.Fatal(err)
			}

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				next, room, err := h.AssignNextWaiting()
				if err != nil {
					b.Fatal(err)
				}
				// Libera de inmediato para que siempre sigan esperando exactamente n.
				if err := room.Release(next); err != nil {
					b.Fatal(err)
				}
				_ = next.WakeUp()
				if err := next.SufferSleepAttack("x"); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkReportSleepAttack mide el camino COMPLETO (el que usa la
// simulación en vivo): dormir, buscar sala, despachar personal, registrar el
// episodio — con sala siempre disponible, para medir el caso feliz.
func BenchmarkReportSleepAttack(b *testing.B) {
	for _, n := range sizes {
		b.Run(fmt.Sprintf("rooms=%d", n), func(b *testing.B) {
			h, _ := benchHospital(b, 1) // 1 sala que se libera y reocupa en cada vuelta
			p := NewPatient("P", 30, NarcolepsyLevelSevere)
			if err := h.AdmitPatient(p); err != nil {
				b.Fatal(err)
			}
			_ = n

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if _, err := h.ReportSleepAttack(p, "cafetería"); err != nil {
					b.Fatal(err)
				}
				if err := h.WakePatient(p); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
