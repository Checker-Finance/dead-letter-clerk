package task

import (
	"context"
	"log"

	_ "dead-letter-clerk/internal/config"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron  *cron.Cron
	tasks []*Task
}

func NewScheduler(tasks []*Task) *Scheduler {
	return &Scheduler{
		cron:  cron.New(),
		tasks: tasks,
	}
}

func (s *Scheduler) Start(ctx context.Context) error {
	for _, t := range s.tasks {
		task := t // capture for closure

		_, err := s.cron.AddFunc(task.Config.Schedule, func() {
			if err := task.Run(ctx); err != nil {
				log.Printf("[scheduler] Task %s failed: %v", task.Config.Name, err)
			}
		})
		if err != nil {
			return err
		}

		log.Printf("[scheduler] Scheduled task: %s (%s)", task.Config.Name, task.Config.Schedule)
	}

	s.cron.Start()
	log.Println("[scheduler] All tasks scheduled")
	return nil
}

// Stop stops the cron scheduler gracefully.
func (s *Scheduler) Stop() {
	s.cron.Stop()
	log.Println("[scheduler] Scheduler stopped")
}
