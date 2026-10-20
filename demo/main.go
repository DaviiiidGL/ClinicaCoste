// cmd/demo/main.go
package main

import (
	"fmt"

	"ClinicaCoste/internal/hospital"
	"ClinicaCoste/internal/hospital/clock"
)

func main() {
	fmt.Println("=== INICIANDO ESCENARIO DE DEMOSTRACIÓN (Punto 6) ===")

	// 1. Crear el hospital (Usamos el reloj real para que la fecha sea la de hoy)
	h, err := hospital.NewHospital("Hospital de los Costeños", clock.RealClock{})
	if err != nil {
		panic(err)
	}

	// 2. Contratar 2 doctores y 1 camillero
	d1 := hospital.NewDoctor("Karen Ospina", 38, "Neurología")
	d2 := hospital.NewDoctor("Wilfrido Berrio", 45, "Neumología")
	o1 := hospital.NewOrderly("Camilo Pérez", 25)
	_ = h.HireDoctor(d1)
	_ = h.HireDoctor(d2)
	_ = h.AddAttender(o1)

	// 3. Registrar 3 habitaciones con capacidad 1
	room101, _ := h.AddRoom(101, 1)
	_, _ = h.AddRoom(102, 1)
	_, _ = h.AddRoom(103, 1)

	// 4. Admitir 5 pacientes (al menos 2 con nivel Severe)
	p1 := hospital.NewPatient("Yeimy Padilla", 29, hospital.NarcolepsyLevelSevere)
	p2 := hospital.NewPatient("Wilfrido Berrio", 40, hospital.NarcolepsyLevelSevere)
	p3 := hospital.NewPatient("Carlos Ruiz", 35, hospital.NarcolepsyLevelModerate)
	p4 := hospital.NewPatient("Ana Gómez", 22, hospital.NarcolepsyLevelMild)
	p5 := hospital.NewPatient("Luis Torres", 50, hospital.NarcolepsyLevelMild)

	for _, p := range []*hospital.Patient{p1, p2, p3, p4, p5} {
		_ = h.AdmitPatient(p)
	}

	// 5. Cuatro pacientes sufren un ataque en diferentes locaciones.
	// Los primeros 3 obtienen habitación.
	_, _ = h.ReportSleepAttack(p1, "Cafetería")
	_, _ = h.ReportSleepAttack(p2, "Pasillo 2")
	_, _ = h.ReportSleepAttack(p3, "Fila de radiología")

	// El cuarto falla y se queda en el pasillo (Imprimimos el error legible)
	_, errP4 := h.ReportSleepAttack(p4, "Farmacia")
	fmt.Println("\n== Asignando habitación al 4to paciente ==")
	if errP4 != nil {
		fmt.Printf("Error: %v\n", errP4) // Mostrará "no room available: patient stays in the hallway"
	} else {
		fmt.Println("Se le asignó habitación al 4to paciente.")
	}

	// 6. Uno de los pacientes se despierta, libera la habitación.
	// El paciente en el pasillo (p4) es asignado a esa habitación automáticamente por AssignNextWaiting.
	fmt.Printf("\n== Despertando a %s (libera la habitación %d) ==\n", p1.Name(), room101.Number())
	_ = h.WakePatient(p1)

	// 7. Imprimir los resultados de las 4 consultas de la Sección 5.

	// Consulta 5.1: Pacientes dormidos en el pasillo
	fmt.Println("\n== Pacientes dormidos en el pasillo ==")
	inHallway := h.PatientsInHallway()
	if len(inHallway) == 0 {
		fmt.Println("(Ninguno)")
	} else {
		for _, p := range inHallway {
			fmt.Printf("- %s (Ubicación: %s)\n", p.Name(), p.Location())
		}
	}

	// Consulta 5.2: Historial por doctor
	fmt.Printf("\n== Episodios atendidos por %s ==\n", d1.Name())
	for _, e := range d1.MyEpisodes() {
		fmt.Println("- " + e.Summary())
	}

	// Consulta 5.3: Disponibilidad de camas (Mostramos las habitaciones y su estado)
	fmt.Println("\n== Estado de las habitaciones ==")
	for _, r := range h.Rooms() {
		fmt.Printf("- Habitación %d | Estado: %s | Ocupantes: %d\n", r.Number(), r.State(), r.CurrentOccupancy())
	}

	// Consulta 5.4: Reporte de severidad
	fmt.Println("\n== Reporte de severidad (Pacientes graves) ==")
	report := h.SevereReportOrdered()
	for _, entry := range report {
		fmt.Printf("- %s : %d episodios hoy\n", entry.Patient.Name(), entry.Episodes)
	}

	fmt.Println("\n=== FIN DEL ESCENARIO ===")
}
