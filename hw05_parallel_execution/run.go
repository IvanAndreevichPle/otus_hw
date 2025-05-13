package hw05parallelexecution

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

func Run(tasks []Task, workersCount, maxErrors int) error {
	if workersCount <= 0 {
		return nil
	}

	if maxErrors <= 0 {
		maxErrors = 0
	}

	var (
		errCounter  int32
		ctx, cancel = context.WithCancel(context.Background())
		taskCh      = make(chan Task)
		wg          sync.WaitGroup
	)

	defer cancel()

	wg.Add(workersCount)
	for i := 0; i < workersCount; i++ {
		go worker(ctx, taskCh, &errCounter, maxErrors, cancel, &wg)
	}

	go sendTasks(ctx, tasks, taskCh)

	wg.Wait()

	if maxErrors == 0 && atomic.LoadInt32(&errCounter) > 0 {
		return ErrErrorsLimitExceeded
	}
	if maxErrors > 0 && atomic.LoadInt32(&errCounter) >= int32(maxErrors) {
		return ErrErrorsLimitExceeded
	}

	return nil
}

func worker(ctx context.Context, taskCh <-chan Task, errCounter *int32, maxErrors int, cancel context.CancelFunc, wg *sync.WaitGroup) {
	defer wg.Done()

	for task := range taskCh {
		if err := task(); err != nil {
			newCount := atomic.AddInt32(errCounter, 1)

			if (maxErrors > 0 && newCount >= int32(maxErrors)) || (maxErrors == 0 && newCount > 0) {
				cancel()
			}
		}
	}
}

func sendTasks(ctx context.Context, tasks []Task, taskCh chan<- Task) {
	defer close(taskCh)

	for _, task := range tasks {
		select {
		case taskCh <- task:
		case <-ctx.Done():
			return
		}
	}
}
