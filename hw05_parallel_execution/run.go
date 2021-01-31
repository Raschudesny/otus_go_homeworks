package hw05_parallel_execution //nolint:golint,stylecheck

import (
	"errors"
	"sync"
	"sync/atomic"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")
var ErrNoWorkersProvided = errors.New("wrong number of workers provided for Run func")

type Task func() error

// Run starts tasks in N goroutines and stops its work when receiving M errors from tasks.
func Run(tasks []Task, n int, m int) error {
	if n == 0 {
		return ErrNoWorkersProvided
	}
	if len(tasks) == 0 {
		return nil
	}

	wg := sync.WaitGroup{}
	var errorsCounter uint32
	tasksCh := make(chan Task, len(tasks))

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range tasksCh {
				if int(atomic.LoadUint32(&errorsCounter)) >= m {
					return
				}
				if taskError := task(); taskError != nil {
					atomic.AddUint32(&errorsCounter, 1)
				}
			}
		}()
	}
	for _, task := range tasks {
		tasksCh <- task
	}
	close(tasksCh)

	wg.Wait()
	if int(errorsCounter) >= m {
		return ErrErrorsLimitExceeded
	}

	return nil
}
