// cmd/server/main.go
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"ClinicaCoste/internal/hospital"
	"ClinicaCoste/internal/hospital/clock"
	"ClinicaCoste/internal/hospital/simulation"
	"ClinicaCoste/pkg/api/connect"
)

func main() {
	// 1. Reloj simulado (empieza a las 8:00 AM de hoy)
	startTime := time.Now()
	simClock := clock.NewSimulatedClock(startTime, 1.0)

	// 2. Inicializamos el hospital
	h, err := hospital.NewHospital("Clínica Coste", simClock)
	if err != nil {
		log.Fatalf("NewHospital: %v", err)
	}

	// 3. Hacemos avanzar el reloj simulado
	go func() {
		for {
			time.Sleep(1 * time.Second)
			simClock.Advance(1 * time.Minute)
		}
	}()

	// 4. Sembramos las habitaciones y el personal primero (esto es instantáneo)
	seedInfrastructure(h)

	// 5. Arrancamos el motor de simulación
	stopSimulation := simulation.Start(h)
	defer stopSimulation()

	// 6. Arrancamos el servidor web en una goroutine para que no bloquee la admisión
	mux := connectapi.NewServer(connectapi.NewHospitalServer(h))
	handler := withCORS(mux)
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}

	go func() {
		log.Printf("🏥 Clínica Coste escuchando en http://localhost%s", addr)
		if err := http.ListenAndServe(addr, handler); err != nil {
			log.Fatalf("ListenAndServe: %v", err)
		}
	}()

	log.Printf("⏰ Reloj iniciado a las %s.", startTime.Format("15:04"))

	// 7. ¡LA MAGIA! Admitimos a los pacientes escalonados (uno cada 800ms)
	seedPatientsGradually(h)

	// Mantenemos el programa vivo
	select {}
}

// seedInfrastructure crea las camas y el personal
func seedInfrastructure(h *hospital.Hospital) {
	for piso := 1; piso <= 3; piso++ {
		for i := 1; i <= 6; i++ {
			if _, err := h.AddRoom((piso*100)+i, 2); err != nil {
				log.Fatalf("AddRoom %d: %v", (piso*100)+i, err)
			}
		}
	}

	staff := []struct {
		role string
		add  func() error
	}{
		{"Dra. Manuela Montoya", func() error { return h.HireDoctor(hospital.NewDoctor("Manuela Montoya", 42, "Neurología")) }},
		{"Dr. Wilfrido Padilla", func() error { return h.HireDoctor(hospital.NewDoctor("Wilfrido Padilla", 45, "Neumología")) }},
		{"Dra. Sofía Vergara", func() error { return h.HireDoctor(hospital.NewDoctor("Sofía Vergara", 35, "Medicina General")) }},
		{"Enf. Andrea Gómez", func() error { return h.AddAttender(hospital.NewNurse("Andrea Gómez", 30)) }},
		{"Cam. Camilo Pérez", func() error { return h.AddAttender(hospital.NewOrderly("Camilo Pérez", 25)) }},
	}
	for _, s := range staff {
		if err := s.add(); err != nil {
			log.Fatalf("seed %s: %v", s.role, err)
		}
	}
	fmt.Printf("✅ Infraestructura lista: %d camas, %d staff\n", len(h.Rooms())*2, len(h.Staff()))
}

// seedPatientsGradually admite a los pacientes de a uno por vez
func seedPatientsGradually(h *hospital.Hospital) {
	niveles := []hospital.NarcolepsyLevel{
		hospital.NarcolepsyLevelSevere,
		hospital.NarcolepsyLevelModerate,
		hospital.NarcolepsyLevelMild,
	}
	nombres := []string{
		"Yeimy Padilla", "Andrés Mendoza", "Valentina Ríos", "Diego Torres",
		"Camila Restrepo", "Jorge Arango", "Mariana López", "Felipe Castro",
		"Lucía Martínez", "Sebastián Gómez", "Daniela Pineda", "Mateo Vargas",
		"Carolina Mejía", "Alejandro Rojas", "Sofía Hernández", "David Ospina",
		"Isabella Muñoz", "Santiago Cardona", "Valeria Zapata", "Nicolás Salazar",
	}

	log.Println("👥 Iniciando admisión escalonada de pacientes...")
	for i, nombre := range nombres {
		level := niveles[i%3]
		p := hospital.NewPatient(nombre, 20+(i%40), level)

		if err := h.AdmitPatient(p); err != nil {
			log.Printf("No se pudo admitir a %s: %v", nombre, err)
		} else {
			// Hacemos que lleguen uno por uno cada 800 milisegundos
			time.Sleep(800 * time.Millisecond)
		}
	}
	fmt.Printf("✅ Admisión completa: %d pacientes en el hospital.\n", len(h.Patients()))
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Connect-Protocol-Version, Connect-Timeout-Ms")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
