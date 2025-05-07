package reader

import (
	"context"
	"fmt"
	"log"
	"strings"

	"dead-letter-clerk/internal/config"
	"dead-letter-clerk/internal/redis"
	goredis "github.com/redis/go-redis/v9"
)

type StreamReader struct{}

// NewStreamReader returns a new StreamReader instance.
func NewStreamReader() *StreamReader {
	return &StreamReader{}
}

// Read reads entries from a Redis stream and decodes them to map[string]any.
func (r *StreamReader) Read(ctx context.Context, taskCfg config.TaskConfig, client *redis.Client) ([]map[string]any, error) {
	key := taskCfg.RedisKey
	checkpoint := taskCfg.Checkpoint

	startID := "0-0" // default: beginning of stream
	if checkpoint != nil && checkpoint.Enabled {
		val, err := client.GetString(ctx, checkpoint.LastValueKey)
		if err != nil {
			return nil, fmt.Errorf("failed to get last checkpoint: %w", err)
		}
		if val != "" {
			startID = val
		}
	}

	const count = 100
	args := &goredis.XReadArgs{
		Streams: []string{key, startID},
		Count:   count,
		Block:   0, // non-blocking
	}

	streams, err := client.Client.XRead(ctx, args).Result()
	if err != nil && err != goredis.Nil {
		return nil, fmt.Errorf("failed to read from stream: %w", err)
	}

	var records []map[string]any
	var lastID string

	for _, stream := range streams {
		for _, msg := range stream.Messages {
			record := make(map[string]any)
			for k, v := range msg.Values {
				record[k] = v
			}
			record["__stream_id"] = msg.ID // inject for checkpointing
			lastID = msg.ID
			records = append(records, record)
		}
	}

	// Store checkpoint
	if checkpoint != nil && checkpoint.Enabled && lastID != "" {
		// Don't reuse the same ID to avoid duplicate delivery
		parts := strings.Split(lastID, "-")
		if len(parts) == 2 {
			seq := parts[1]
			lastID = fmt.Sprintf("%s-%d", parts[0], toInt(seq)+1)
		}
		if err := client.SetString(ctx, checkpoint.LastValueKey, lastID); err != nil {
			log.Printf("[stream_reader] warning: failed to store checkpoint: %v", err)
		} else {
			log.Printf("[stream_reader] updated checkpoint: %s", lastID)
		}
	}

	return records, nil
}

func toInt(s string) int {
	n, _ := fmt.Sscanf(s, "%d", new(int))
	return n
}
