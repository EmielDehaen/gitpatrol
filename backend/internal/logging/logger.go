package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"gitpatrol/internal/database"
)

type DBHandler struct {
	slog.Handler
	db          *database.DB
	wsBroadcast func(map[string]interface{})
}

func NewDBHandler(db *database.DB, wsBroadcast func(map[string]interface{})) *DBHandler {
	// Base handler is a JSON handler writing to stdout
	base := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	return &DBHandler{
		Handler:     base,
		db:          db,
		wsBroadcast: wsBroadcast,
	}
}

func (h *DBHandler) Handle(ctx context.Context, r slog.Record) error {
	// Process standard handling (stdout)
	err := h.Handler.Handle(ctx, r)

	// Process attributes to JSON string
	attrs := make(map[string]interface{})
	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.Any()
		return true
	})

	attrJSON, _ := json.Marshal(attrs)

	// Insert into DB
	go func() {
		_, dbErr := h.db.Exec(`
			INSERT INTO system_logs (level, message, attributes, created_at)
			VALUES (?, ?, ?, ?)
		`, r.Level.String(), r.Message, string(attrJSON), r.Time)
		
		if dbErr != nil {
			fmt.Fprintf(os.Stderr, "failed to insert log to db: %v\n", dbErr)
		}
	}()

	// Broadcast via WebSocket
	if h.wsBroadcast != nil {
		logEntry := map[string]interface{}{
			"level":      r.Level.String(),
			"message":    r.Message,
			"attributes": attrs,
			"time":       r.Time,
		}
		h.wsBroadcast(logEntry)
	}

	return err
}

// Setup initializes the global slog logger with the custom DBHandler
func Setup(db *database.DB, wsBroadcast func(map[string]interface{})) {
	handler := NewDBHandler(db, wsBroadcast)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
