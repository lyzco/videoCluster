package job_test

import (
	"testing"
	"time"

	"videoCluster/internal/node/job"
)

func TestStartRunsScanImmediately(t *testing.T) {
	run := make(chan struct{}, 1)
	manager := job.NewJobManager(func() { run <- struct{}{} }, time.Hour, false)
	defer manager.Stop()

	manager.Start()

	select {
	case <-run:
	case <-time.After(time.Second):
		t.Fatal("scan was not run when the job manager started")
	}
}

func TestSkipOverlappingRunWhenConcurrentRunsAreDisabled(t *testing.T) {
	started := make(chan struct{}, 1)
	release := make(chan struct{})
	done := make(chan struct{})
	manager := job.NewJobManager(func() {
		started <- struct{}{}
		<-release
		close(done)
	}, time.Hour, false)

	go manager.Run()
	<-started
	go manager.Run()

	select {
	case <-started:
		t.Fatal("overlapping task was not skipped")
	case <-time.After(50 * time.Millisecond):
	}

	release <- struct{}{}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("first task did not finish")
	}
}

func TestAllowOverlappingRunWhenConcurrentRunsAreEnabled(t *testing.T) {
	started := make(chan struct{}, 2)
	release := make(chan struct{}, 2)
	done := make(chan struct{}, 2)
	manager := job.NewJobManager(func() {
		started <- struct{}{}
		<-release
		done <- struct{}{}
	}, time.Hour, true)

	go manager.Run()
	go manager.Run()

	for range 2 {
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("concurrent task did not start")
		}
	}
	release <- struct{}{}
	release <- struct{}{}
	for range 2 {
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("concurrent task did not finish")
		}
	}
}
