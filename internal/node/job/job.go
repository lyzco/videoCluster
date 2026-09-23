package job

import (
	"context"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

type JobManager struct {
	c               *cron.Cron
	allowConcurrent bool
	runMu           sync.Mutex
	task            func()
}

func NewJobManager(task func(), interval time.Duration, allowConcurrent bool) *JobManager {
	c := cron.New()
	manager := &JobManager{c: c, allowConcurrent: allowConcurrent, task: task}
	c.Schedule(cron.Every(interval), cron.FuncJob(manager.Run))

	return manager
}

// Run executes the task once.
func (j *JobManager) Run() {
	if !j.allowConcurrent {
		if !j.runMu.TryLock() {
			return
		}
		defer j.runMu.Unlock()
	}
	j.task()
}

func (j *JobManager) Start() {
	// Populate the local catalog before scheduling subsequent scans.
	j.Run()
	j.c.Start()
}

func (j *JobManager) Stop() context.Context {
	return j.c.Stop()
}
