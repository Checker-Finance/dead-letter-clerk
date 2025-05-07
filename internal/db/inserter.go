package db

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"dead-letter-clerk/internal/config"
)

func (db *Database) InsertRows(ctx context.Context, taskCfg config.TaskConfig, records []map[string]any) error {
	if len(records) == 0 {
		return nil
	}

	columnMap := taskCfg.FieldMap
	columns := make([]string, 0, len(columnMap))
	for _, col := range columnMap {
		columns = append(columns, col)
	}
	
	placeholderCount := len(columns)
	valuePlaceholders := make([]string, 0, len(records))
	values := make([]any, 0, len(records)*placeholderCount)

	for i, record := range records {
		placeholders := make([]string, 0, placeholderCount)
		for j, redisField := range getSortedKeys(columnMap) {
			//dbCol := columnMap[redisField]
			placeholders = append(placeholders, fmt.Sprintf("$%d", i*placeholderCount+j+1))
			values = append(values, record[redisField])
		}
		valuePlaceholders = append(valuePlaceholders, fmt.Sprintf("(%s)", strings.Join(placeholders, ", ")))
	}

	sql := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES %s`,
		taskCfg.DBTable,
		strings.Join(columns, ", "),
		strings.Join(valuePlaceholders, ", "),
	)

	_, err := db.Pool.Exec(ctx, sql, values...)
	if err != nil {
		return fmt.Errorf("failed to insert into %s: %w", taskCfg.DBTable, err)
	}

	return nil
}

func getSortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	return keys
}
