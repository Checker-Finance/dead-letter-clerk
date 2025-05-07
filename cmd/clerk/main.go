package main

import (
	"context"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"os/signal"
	"syscall"

	"dead-letter-clerk/internal/config"
	"dead-letter-clerk/internal/db"
	"dead-letter-clerk/internal/redis"
	readerpkg "dead-letter-clerk/internal/redis/reader"
	"dead-letter-clerk/internal/task"
)

func loadConfig(path string) (*config.AppConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg config.AppConfig
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle CTRL+C
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig
		log.Println("Received interrupt. Shutting down...")
		cancel()
	}()

	cfg, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Initialize clients
	redisClient := redis.NewClient(cfg.Redis)
	defer redisClient.Close()

	pgDB, err := db.NewDatabase(ctx, cfg.Postgres)
	if err != nil {
		log.Fatalf("failed to connect to Postgres: %v", err)
	}
	defer pgDB.Close()

	writer := db.NewWriter(pgDB)

	// Build tasks
	var tasks []*task.Task

	for _, tCfg := range cfg.Tasks {
		var redisReader readerpkg.Reader

		switch tCfg.RedisType {
		case "list":
			redisReader = readerpkg.NewListReader()
		case "stream":
			redisReader = readerpkg.NewStreamReader()
		case "sorted_set":
			redisReader = readerpkg.NewSortedSetReader()
		default:
			log.Fatalf("unsupported redis_type: %s", tCfg.RedisType)
		}

		t := &task.Task{
			Config:      tCfg,
			RedisReader: redisReader,
			Writer:      writer,
			RedisClient: redisClient,
		}
		tasks = append(tasks, t)
	}

	// Start scheduler
	scheduler := task.NewScheduler(tasks)
	if err := scheduler.Start(ctx); err != nil {
		log.Fatalf("failed to start scheduler: %v", err)
	}

	// Wait for termination
	<-ctx.Done()
	scheduler.Stop()
}
