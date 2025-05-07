package reader

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"dead-letter-clerk/internal/config"
	"dead-letter-clerk/internal/redis"
	goredis "github.com/redis/go-redis/v9"
)

type SortedSetReader struct{}

// NewSortedSetReader creates a new SortedSetReader.
func NewSortedSetReader() *SortedSetReader {
	return &SortedSetReader{}
}

// Read pulls entries from a Redis sorted set using ZRANGEBYSCORE and optional checkpointing.
func (r *SortedSetReader) Read(ctx context.Context, taskCfg config.TaskConfig, client *redis.Client) ([]map[string]any, error) {
	key := taskCfg.RedisKey
	checkpoint := taskCfg.Checkpoint

	// Determine min score
	min := "-inf"
	if checkpoint != nil && checkpoint.Enabled {
		lastVal, err := client.GetString(ctx, checkpoint.LastValueKey)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch last checkpoint: %w", err)
		}
		if lastVal != "" {
			min = fmt.Sprintf("(%s", lastVal) // exclusive
		}
	}

	// Define ZRangeBy (value, not pointer)
	zrange := goredis.ZRangeBy{
		Min:    min,
		Max:    "+inf",
		Offset: 0,
		Count:  100,
	}

	items, err := client.Client.ZRangeByScore(ctx, key, &zrange).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to read sorted set: %w", err)
	}

	var records []map[string]any
	var highestScore string

	for _, raw := range items {
		var rec map[string]any
		if err := json.Unmarshal([]byte(raw), &rec); err != nil {
			log.Printf("[sorted_set_reader] skipping invalid entry: %v", err)
			continue
		}

		// Check for checkpoint value
		if checkpoint != nil && checkpoint.Enabled {
			if val, ok := rec[checkpoint.Field]; ok {
				score := stringifyScore(val)
				if score > highestScore {
					highestScore = score
				}
			}
		}

		records = append(records, rec)
	}

	// Save checkpoint
	if checkpoint != nil && checkpoint.Enabled && highestScore != "" {
		if err := client.SetString(ctx, checkpoint.LastValueKey, highestScore); err != nil {
			log.Printf("[sorted_set_reader] warning: failed to store checkpoint: %v", err)
		} else {
			log.Printf("[sorted_set_reader] updated checkpoint: %s", highestScore)
		}
	}

	return records, nil
}

func stringifyScore(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	default:
		return ""
	}
}
