package job

import (
	"context"
	"github.com/robfig/cron/v3"
	"videoCluster/internal/biz"
)

type JobManager struct {
	c *cron.Cron
}

func NewJobManager(scanner *biz.VideoScanner) *JobManager {
	c := cron.New()

	// 每5分钟扫描
	c.AddFunc("*/30 * * * *", func() {
		err := scanner.ScanAndCache(context.Background())
		if err != nil {
			return
		}
	})

	return &JobManager{c: c}
}

func (j *JobManager) Start() {
	j.c.Start()
}

func (j *JobManager) Stop() context.Context {
	return j.c.Stop()
}
