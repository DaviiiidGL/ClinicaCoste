package simulation

import (
	"math/rand/v2"
	"sync"
	"time"

	"ClinicaCoste/internal/hospital"
)

func Start(h *hospital.Hospital) func() {
	stopChan := make(chan struct{})
	var wg sync.WaitGroup

	// Tasa de refresco del reloj del hospital
	tickInterval := 200 * time.Millisecond

	wg.Add(1)
	go func() {
		defer wg.Done()

		nextAttack := make(map[string]time.Time)
		nextWake := make(map[string]time.Time)

		for {
			select {
			case <-stopChan:
				return
			case <-time.After(tickInterval):
				now := h.Clock().Now()
				patients := h.Patients()

				for _, p := range patients {
					// Usamos rand.IntN para garantizar que los tiempos NO sean iguales
					// No dependemos del reloj virtual para esto, así rompemos la sincronía.

					if p.State() == hospital.PatientStateAwake {
						delete(nextWake, p.ID())

						if _, exists := nextAttack[p.ID()]; !exists {
							// Se duermen en un rango amplio de 1 a 30 segundos virtuales
							delay := time.Duration(rand.IntN(30)+1) * time.Second
							nextAttack[p.ID()] = now.Add(delay)
						}

						if now.After(nextAttack[p.ID()]) {
							delete(nextAttack, p.ID())

							// Lanzamos en goroutine y dormimos AQUÍ ADENTRO (no en el motor principal)
							// para que el motor siga revisando a los demás.
							go func(patient *hospital.Patient) {
								time.Sleep(time.Duration(rand.IntN(200)) * time.Millisecond)
								_, _ = h.ReportSleepAttack(patient, "Pasillo (Automático)")
							}(p)
						}

					} else if p.State() == hospital.PatientStateAsleep {
						delete(nextAttack, p.ID())

						if _, exists := nextWake[p.ID()]; !exists {
							// Duermen entre 20 y 50 segundos virtuales
							delay := time.Duration(rand.IntN(30)+20) * time.Second
							nextWake[p.ID()] = now.Add(delay)
						}

						if now.After(nextWake[p.ID()]) {
							delete(nextWake, p.ID())

							go func(patient *hospital.Patient) {
								time.Sleep(time.Duration(rand.IntN(200)) * time.Millisecond)
								_ = h.WakePatient(patient)
							}(p)
						}
					}
				}
			}
		}
	}()

	return func() {
		close(stopChan)
		wg.Wait()
	}
}
