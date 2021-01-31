package hw05_parallel_execution //nolint:golint,stylecheck

import (
	"errors"
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestRun(t *testing.T) {
	defer goleak.VerifyNone(t)

	t.Run("if were errors in first M tasks, than finished not more N+M tasks", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32

		for i := 0; i < tasksCount; i++ {
			err := fmt.Errorf("error from task %d", i)
			tasks = append(tasks, func() error {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
				atomic.AddInt32(&runTasksCount, 1)
				return err
			})
		}

		workersCount := 10
		maxErrorsCount := 23
		err := Run(tasks, workersCount, maxErrorsCount)

		require.Truef(t, errors.Is(err, ErrErrorsLimitExceeded), "actual err - %v", err)
		require.LessOrEqual(t, runTasksCount, int32(workersCount+maxErrorsCount), "extra tasks were started")
	})

	t.Run("tasks without errors", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32
		var sumTime time.Duration

		for i := 0; i < tasksCount; i++ {
			taskSleep := time.Millisecond * time.Duration(rand.Intn(100))
			sumTime += taskSleep

			tasks = append(tasks, func() error {
				time.Sleep(taskSleep)
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})
		}

		workersCount := 5
		maxErrorsCount := 1

		start := time.Now()
		err := Run(tasks, workersCount, maxErrorsCount)
		elapsedTime := time.Since(start)
		require.NoError(t, err)

		require.Equal(t, int32(tasksCount), runTasksCount, "not all tasks were completed")
		require.LessOrEqual(t, int64(elapsedTime), int64(sumTime/2), "tasks were run sequentially?")
	})

	t.Run("lower than M errors in N tasks", func(t *testing.T) {
		tasksCount := 100
		realErrorsCount := 10
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32

		for i, errorsCounter := 0, 0; i < tasksCount; i, errorsCounter = i+1, errorsCounter+1 {
			// for example let's add error on first realErrorsCount iteration which is even number
			if i%2 == 0 && errorsCounter < realErrorsCount {
				err := fmt.Errorf("error from task %d", i)
				tasks = append(tasks, func() error {
					time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
					atomic.AddInt32(&runTasksCount, 1)
					return err
				})
			} else {
				tasks = append(tasks, func() error {
					time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
					atomic.AddInt32(&runTasksCount, 1)
					return nil
				})
			}
		}

		workersCount := 10
		maxErrorsCount := 23
		err := Run(tasks, workersCount, maxErrorsCount)

		require.Truef(t, err == nil, "actual err - %v", err)
		require.LessOrEqual(t, runTasksCount, int32(tasksCount), "extra tasks were started")
	})

	t.Run("asterisk task test", func(t *testing.T) {
		// here we need to check that tasks execution works concurrently without using time.sleep
		var runTasksCount int32
		workersCount := 2
		tasksCount := workersCount

		tasks := make([]Task, 0, tasksCount)
		blockignCh := make(chan struct{})
		tasks = append(tasks, func() error {
			defer atomic.AddInt32(&runTasksCount, 1)
			blockignCh <- struct{}{}
			return nil
		})
		tasks = append(tasks, func() error {
			defer atomic.AddInt32(&runTasksCount, 1)
			<-blockignCh
			return nil
		})
		err := Run(tasks, workersCount, tasksCount)
		require.Truef(t, err == nil, "actual err - %v", err)
		require.Eventually(t, func() bool { return int32(tasksCount) == runTasksCount }, time.Second*5, time.Second*1, "goroutines don't finished in 5 seconds, looks like tasks were run sequentially? ")
	})

	t.Run("zero errors allowed tests", func(t *testing.T) {
		tasks := []Task{func() error {
			fmt.Println("some random task")
			return nil
		}}
		workersCount := 10
		maxErrorsCount := 0

		err := Run(tasks, workersCount, maxErrorsCount)
		require.Truef(t, errors.Is(err, ErrErrorsLimitExceeded), "actual err - %v", err)
	})

	t.Run("zero workers provided", func(t *testing.T) {
		err := Run(nil, 0, 20)
		require.Truef(t, errors.Is(err, ErrNoWorkersProvided), "actual err - %v", err)
	})
}
