package worker

import (
	"log"
	"time"

	"github.com/hibiken/asynq"
)

// Scheduler wraps the asynq.Scheduler
type Scheduler struct {
	scheduler *asynq.Scheduler
}

// NewScheduler creates a new scheduler instance
func NewScheduler(redisAddr string) *Scheduler {
	loc, err := time.LoadLocation("UTC")
	if err != nil {
		log.Fatal(err)
	}

	scheduler := asynq.NewScheduler(
		ParseRedisOpt(redisAddr),
		&asynq.SchedulerOpts{
			Location: loc,
			PostEnqueueFunc: func(info *asynq.TaskInfo, err error) {
				if err != nil {
					log.Printf("Scheduler error: %v", err)
					return
				}
				log.Printf("Synthesized Task: %s (queue=%s)", info.Type, info.Queue)
			},
		},
	)

	return &Scheduler{
		scheduler: scheduler,
	}
}

// RegisterEntry adds a new cron job
func (s *Scheduler) RegisterEntry(cronSpec string, task *asynq.Task, opts ...asynq.Option) (string, error) {
	return s.scheduler.Register(cronSpec, task, opts...)
}

// Run starts the scheduler
func (s *Scheduler) Run() error {
	log.Println("Starting Asynq Scheduler...")
	return s.scheduler.Run()
}

// Shutdown stops the scheduler
func (s *Scheduler) Shutdown() {
	s.scheduler.Shutdown()
}
