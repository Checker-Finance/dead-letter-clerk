package db

import (
	"context"
	"fmt"

	"dead-letter-clerk/internal/config"
)

type Writer struct {
	DB *Database
}

func NewWriter(db *Database) *Writer {
	return &Writer{DB: db}
}

func (w *Writer) Write(ctx context.Context, taskCfg config.TaskConfig, records []map[string]any) error {
	if len(records) == 0 {
		return nil
	}

	err := w.DB.InsertRows(ctx, taskCfg, records)
	if err != nil {
		return fmt.Errorf("failed to write records for task %s: %w", taskCfg.Name, err)
	}

	return nil
}
