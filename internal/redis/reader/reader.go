package reader

import (
	"context"
	"dead-letter-clerk/internal/config"
	"dead-letter-clerk/internal/redis"
)

type Reader interface {
	Read(ctx context.Context, taskCfg config.TaskConfig, client *redis.Client) ([]map[string]any, error)
}

func ExtractMaxCheckpoint(records []map[string]any, field string) string {
	var max string
	for _, rec := range records {
		if val, ok := rec[field]; ok {
			if s, ok := val.(string); ok && s > max {
				max = s
			}
		}
	}
	return max
}
