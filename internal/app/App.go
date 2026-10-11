package app

import (
	"flag"
	"fmt"
	"io"

	"ClinicaCoste/internal/config"
	"ClinicaCoste/internal/report"
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
	sc, err := scenario.Build(cfg)
	if err != nil {
		return fmt.Errorf("building scenario: %w", err)
	}
	res, err := sc.Run()
	if err != nil {
		return fmt.Errorf("running scenario: %w", err)
	}
	if err := report.Print(stdout, res); err != nil {
		return fmt.Errorf("printing report: %w", err)
	}

	if path := cfg.Storage.HistoryPath; path != "" {
		repo := storage.NewJSONRepository(path)
		if err := repo.SaveHistory(storage.EntriesFromRecords(res.History)); err != nil {
			return fmt.Errorf("saving history: %w", err)
		}
		if _, err := fmt.Fprintf(stdout, "\nHistorial guardado en %s (%d episodios)\n", path, len(res.History)); err != nil {
			return err
		}
	}
	return nil
}
