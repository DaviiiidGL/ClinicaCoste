package app

import (
	"ClinicaCoste/internal/hospital/clock"
	"ClinicaCoste/internal/hospital/simulation"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ClinicaCoste/internal/config"
	"ClinicaCoste/internal/scenario"
	"ClinicaCoste/internal/storage"
)

func Run(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("hospital", flag.ContinueOnError)
	fs.SetOutput(stderr)
	configPath := fs.String("config", "config.json", "ruta del archivo de configuración")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		return err
	}

	// Inicializar reloj simulado
	startTime := time.Date(2026, 10, 13, 8, 0, 0, 0, time.UTC)
	clk := clock.NewSimulatedClock(startTime, 5.0)

	sc, err := scenario.BuildWithClock(cfg, clk)
	if err != nil {
		return fmt.Errorf("building scenario: %w", err)
	}

	// Contexto con cancelación limpia por señales del sistema (Ctrl+C)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Suscribirse al canal de eventos en tiempo real
	eventChan, cancelSub := sc.Hospital().Subscribe(50)
	defer cancelSub()

	_, _ = fmt.Fprintln(stdout, "════════════════════════════════════════════════════════════")
	_, _ = fmt.Fprintln(stdout, " 🏥 La Clinica del Coste - Modo Simulación en Vivo (Concurrente)")
	_, _ = fmt.Fprintln(stdout, "════════════════════════════════════════════════════════════")

	// Goroutine para capturar e imprimir cada evento conforme va ocurriendo en el tiempo
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case episode, ok := <-eventChan:
				if !ok {
					return
				}
				pName := "Desconocido"
				if episode.Patient() != nil {
					pName = episode.Patient().Name()
				}

				attName := "Sin personal disponible (Pasillo)"
				if episode.AttendingDoctor() != nil {
					attName = episode.AttendingDoctor().Name()
				}

				roomInfo := "en el pasillo (sin sala)"
				if episode.AssignedRoom() != nil {
					roomInfo = fmt.Sprintf("asignado a la Sala %d", episode.AssignedRoom().Number())
				}

				// Usamos la hora actual del sistema para ver el flujo en tiempo real
				currentTime := time.Now().Format("15:04:05")
				_, _ = fmt.Fprintf(stdout, "[EN VIVO] [%s] 🚨 Ataque de sueño: %s en [%s] ➔ Atendido por: %s [%s]\n",
					currentTime, pName, episode.Location(), attName, roomInfo)
			}
		}
	}()

	// Iniciar el motor de simulación concurrente con los 100 pacientes y 30 médicos
	stopSim := simulation.Start(sc.Hospital()) // <-- usamos el hospital del escenario
	defer stopSim()

	_, _ = fmt.Fprintln(stdout, "--- Hospital operando dinámicamente. Presiona Ctrl+C para salir ---")

	// Mantener la aplicación abierta procesando eventos hasta que el usuario presione Ctrl+C
	<-ctx.Done()
	_, _ = fmt.Fprintln(stdout, "\nApagando simulación y liberando recursos...")

	// Opcional: Guardar historial acumulado al salir
	if path := cfg.Storage.HistoryPath; path != "" {
		repo := storage.NewJSONRepository(path)
		// Guardar episodios registrados durante la ejecución en vivo
		historyRecords := sc.Hospital().History()
		if len(historyRecords) > 0 {
			if err := repo.SaveHistory(storage.EntriesFromRecords(historyRecords)); err != nil {
				return fmt.Errorf("saving history: %w", err)
			}
			_, _ = fmt.Fprintf(stdout, "Historial guardado exitosamente en %s (%d episodios totales)\n", path, len(historyRecords))
		}
	}

	return nil
}
