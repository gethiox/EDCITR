package task

import (
	"context"
	"sync"
)

type Task interface {
	Run(ctx context.Context)
}

type TaskManager struct {
	mu            sync.Mutex
	running       bool
	cancel        context.CancelFunc
	doneChan      chan struct{}
	restartActive bool
	task          Task
}

func NewTaskManager(task Task) *TaskManager {
	ch := make(chan struct{})
	close(ch) // Initialize in a "finished" state
	return &TaskManager{
		doneChan: ch,
		task:     task,
	}
}

func (j *TaskManager) Start() bool {
	j.mu.Lock()
	defer j.mu.Unlock()

	if j.running {
		return false
	}

	j.running = true
	ctx, cancel := context.WithCancel(context.Background())
	j.cancel = cancel
	j.doneChan = make(chan struct{})

	go func() {
		defer func() {
			j.mu.Lock()
			j.running = false
			close(j.doneChan)
			j.mu.Unlock()
		}()

		j.task.Run(ctx) // Call the stored task
	}()

	return true
}

func (j *TaskManager) Stop() {
	j.mu.Lock()
	if !j.running {
		j.mu.Unlock()
		return
	}
	cancel := j.cancel
	done := j.doneChan
	j.mu.Unlock()

	cancel()
	<-done
}

func (j *TaskManager) StopAsync() {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.running {
		j.cancel()
	}
}

func (j *TaskManager) Wait() {
	j.mu.Lock()
	done := j.doneChan
	j.mu.Unlock()
	<-done
}

func (j *TaskManager) IsRunning() bool {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.running
}

func (j *TaskManager) RestartAsync() {
	j.mu.Lock()
	if j.restartActive {
		j.mu.Unlock()
		return
	}
	j.restartActive = true
	j.mu.Unlock()

	go func() {
		defer func() {
			j.mu.Lock()
			j.restartActive = false
			j.mu.Unlock()
		}()

		j.Stop()
		j.Start()
	}()
}
