package reader

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"dead-letter-clerk/internal/config"
	"dead-letter-clerk/internal/redis"
)

type ListReader struct{}

func NewListReader() *ListReader {
	return &ListReader{}
}

func (r *ListReader) Read(ctx context.Context, taskCfg config.TaskConfig, client *redis.Client) ([]map[string]any, error) {
	key := taskCfg.RedisKey
	const batchSize = 100
	
	values, err := client.Client.LRange(ctx, key, 0, batchSize-1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to read list %s: %w", key, err)
	}

	var records []map[string]any

	for _, raw := range values {
		var rec map[string]any
		if err := json.Unmarshal([]byte(raw), &rec); err != nil {
			log.Printf("[list_reader] skipping malformed entry: %v", err)
			continue
		}
		records = append(records, rec)
	}

	if len(records) > 0 {
		if err := client.Client.LTrim(ctx, key, int64(len(records)), -1).Err(); err != nil {
			log.Printf("[list_reader] warning: failed to trim list: %v", err)
		}
	}

	return records, nil
}
