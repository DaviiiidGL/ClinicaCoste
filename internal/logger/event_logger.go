package logger

import (
	"log/slog"
	"os"
)

type EventLogger struct {
	logger *slog.Logger
}

func NewEventLogger() *EventLogger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	return &EventLogger{
		logger: slog.New(handler),
	}
}

func (el *EventLogger) LogSleepAttack(patientName, location string, roomNum int) {
	el.logger.Info("Sleep attack reported",
		slog.String("patient", patientName),
		slog.String("location", location),
		slog.Int("room", roomNum),
	)
}
