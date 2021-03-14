package hw06pipelineexecution

import (
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	sleepPerStage = time.Millisecond * 100
	fault         = sleepPerStage / 2
)

func TestPipeline(t *testing.T) {
	// Stage generator
	g := func(name string, f func(v interface{}) interface{}) Stage {
		return func(in In) Out {
			out := make(Bi)
			go func() {
				defer close(out)
				for v := range in {
					time.Sleep(sleepPerStage)
					out <- f(v)
				}
			}()
			return out
		}
	}

	stages := []Stage{
		g("Dummy", func(v interface{}) interface{} { return v }),
		g("Multiplier (* 2)", func(v interface{}) interface{} { return v.(int) * 2 }),
		g("Adder (+ 100)", func(v interface{}) interface{} { return v.(int) + 100 }),
		g("Stringifier", func(v interface{}) interface{} { return strconv.Itoa(v.(int)) }),
	}

	t.Run("simple case", func(t *testing.T) {
		in := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]string, 0, 10)
		start := time.Now()
		for s := range ExecutePipeline(in, nil, stages...) {
			result = append(result, s.(string))
		}
		elapsed := time.Since(start)

		require.Equal(t, []string{"102", "104", "106", "108", "110"}, result)
		require.Less(t,
			int64(elapsed),
			// ~0.8s for processing 5 values in 4 stages (100ms every) concurrently
			int64(sleepPerStage)*int64(len(stages)+len(data)-1)+int64(fault))
	})

	t.Run("done case", func(t *testing.T) {
		in := make(Bi)
		done := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		// Abort after 200ms
		abortDur := sleepPerStage * 2
		go func() {
			<-time.After(abortDur)
			close(done)
		}()

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]string, 0, 10)
		start := time.Now()
		for s := range ExecutePipeline(in, done, stages...) {
			result = append(result, s.(string))
		}
		elapsed := time.Since(start)

		require.Len(t, result, 0)
		require.Less(t, int64(elapsed), int64(abortDur)+int64(fault))
	})

	t.Run("no stages case", func(t *testing.T) {
		inputCh := make(In)
		resultCh := ExecutePipeline(inputCh, make(In))
		require.Equal(t, inputCh, resultCh)
	})

	t.Run("with nil stages", func(t *testing.T) {
		in := make(Bi)
		done := make(Bi)
		dataSize := rand.Intn(30)
		data := make([]int, 0, dataSize)

		for i := 0; i < dataSize; i++ {
			data = append(data, rand.Intn(100))
		}

		go func() {
			defer close(in)
			for _, v := range data {
				in <- v
			}
		}()

		result := make([]string, 0, dataSize)
		stagesWithNils := append(stages[:2], nil)
		stagesWithNils = append(stagesWithNils, stages[2:4]...)
		stagesWithNils = append(stagesWithNils, nil)
		stagesWithNils = append(stages[4:], nil)

		for s := range ExecutePipeline(in, done, stages...) {
			result = append(result, s.(string))
		}

		require.Len(t, result, dataSize)
	})

	t.Run("done channel already closed case ", func(t *testing.T) {
		in := make(Bi)
		done := make(Bi)
		dataSize := rand.Intn(30)
		data := make([]int, 0, dataSize)

		for i := 0; i < dataSize; i++ {
			data = append(data, rand.Intn(100))
		}

		go func() {
			defer close(in)
			for _, v := range data {
				in <- v
			}
		}()

		// whoops, i forget that i've already closed done =(
		close(done)

		result := make([]string, 0, dataSize)
		for s := range ExecutePipeline(in, done, stages...) {
			result = append(result, s.(string))
		}

		require.Empty(t, result)
	})

	t.Run("in channel already closed case ", func(t *testing.T) {
		in := make(Bi)
		// whoops, i forget that i've already closed in channel =(
		close(in)

		result := make([]string, 0)
		for s := range ExecutePipeline(in, nil, stages...) {
			result = append(result, s.(string))
		}

		require.Empty(t, result)
	})
}
