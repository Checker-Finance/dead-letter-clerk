package task

import (
	"context"
	"fmt"
	"log"

	"dead-letter-clerk/internal/config"
	"dead-letter-clerk/internal/db"
	"dead-letter-clerk/internal/redis"

	readerpkg "dead-letter-clerk/internal/redis/reader"
)

type Task struct {
	Config      config.TaskConfig
	RedisReader readerpkg.Reader // e.g., ListReader, StreamReader
	Writer      *db.Writer
	RedisClient *redis.Client // For checkpoint persistence
}

// Run executes the task: reads from Redis and writes to Postgres.
func (t *Task) Run(ctx context.Context) error {
	log.Printf("[task:%s] Starting", t.Config.Name)

	// Fetch from Redis (with checkpoint if enabled)
	records, err := t.RedisReader.Read(ctx, t.Config, t.RedisClient)
	if err != nil {
		return fmt.Errorf("failed to read from Redis: %w", err)
	}

	if len(records) == 0 {
		log.Printf("[task:%s] No new data found", t.Config.Name)
		return nil
	}

	// Insert into Postgres
	if err := t.Writer.Write(ctx, t.Config, records); err != nil {
		return fmt.Errorf("failed to write to Postgres: %w", err)
	}

	// Optional: update checkpoint
	if t.Config.Checkpoint != nil && t.Config.Checkpoint.Enabled {
		latest := readerpkg.ExtractMaxCheckpoint(records, t.Config.Checkpoint.Field)
		if latest != "" {
			if err := t.RedisClient.SetString(ctx, t.Config.Checkpoint.LastValueKey, latest); err != nil {
				log.Printf("[task:%s] Failed to store checkpoint: %v", t.Config.Name, err)
			} else {
				log.Printf("[task:%s] Updated checkpoint: %s", t.Config.Name, latest)
			}
		}
	}

	log.Printf("[task:%s] Completed (%d records)", t.Config.Name, len(records))
	return nil
}
