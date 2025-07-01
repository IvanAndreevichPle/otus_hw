package hw06pipelineexecution

import (
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
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
	g := func(_ string, f func(v interface{}) interface{}) Stage {
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
			int64(sleepPerStage)*int64(len(stages)+len(data)-1)+int64(fault))
	})

	t.Run("done case", func(t *testing.T) {
		in := make(Bi)
		done := make(Bi)
		data := []int{1, 2, 3, 4, 5}

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
}

func TestAllStageStop(t *testing.T) {
	wg := sync.WaitGroup{}
	g := func(_ string, f func(v interface{}) interface{}) Stage {
		return func(in In) Out {
			out := make(Bi)
			wg.Add(1)
			go func() {
				defer wg.Done()
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

	t.Run("done case", func(t *testing.T) {
		in := make(Bi)
		done := make(Bi)
		data := []int{1, 2, 3, 4, 5}

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
		for s := range ExecutePipeline(in, done, stages...) {
			result = append(result, s.(string))
		}
		wg.Wait()

		require.Len(t, result, 0)
	})
}

func TestEdgeCases(t *testing.T) {
	t.Run("empty stages", func(t *testing.T) {
		in := make(Bi)
		data := []int{1, 2, 3}

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]int, 0)
		for v := range ExecutePipeline(in, nil) {
			result = append(result, v.(int))
		}

		require.Equal(t, []int{1, 2, 3}, result)
	})

	t.Run("single stage", func(t *testing.T) {
		stage := func(in In) Out {
			out := make(Bi)
			go func() {
				defer close(out)
				for v := range in {
					out <- v.(int) * 10
				}
			}()
			return out
		}

		in := make(Bi)
		data := []int{1, 2, 3}

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]int, 0)
		for v := range ExecutePipeline(in, nil, stage) {
			result = append(result, v.(int))
		}

		require.Equal(t, []int{10, 20, 30}, result)
	})

	t.Run("empty input", func(t *testing.T) {
		stage := func(in In) Out {
			out := make(Bi)
			go func() {
				defer close(out)
				for v := range in {
					out <- v.(int) * 2
				}
			}()
			return out
		}

		in := make(Bi)
		close(in)

		result := make([]int, 0)
		for v := range ExecutePipeline(in, nil, stage) {
			result = append(result, v.(int))
		}

		require.Empty(t, result)
	})
}

func TestConcurrency(t *testing.T) {
	t.Run("race condition test", func(t *testing.T) {
		stage := func(in In) Out {
			out := make(Bi)
			go func() {
				defer close(out)
				for v := range in {
					time.Sleep(time.Microsecond)
					out <- v.(int) + 1
				}
			}()
			return out
		}

		stages := []Stage{stage, stage, stage}

		for i := 0; i < 100; i++ {
			in := make(Bi)
			data := make([]int, 1000)
			for j := range data {
				data[j] = j
			}

			go func() {
				for _, v := range data {
					in <- v
				}
				close(in)
			}()

			result := make([]int, 0)
			for v := range ExecutePipeline(in, nil, stages...) {
				result = append(result, v.(int))
			}

			require.Len(t, result, 1000)
			for j, v := range result {
				require.Equal(t, j+3, v)
			}
		}
	})
}

func TestDoneSignalTiming(t *testing.T) {
	t.Run("immediate done", func(t *testing.T) {
		stage := func(in In) Out {
			out := make(Bi)
			go func() {
				defer close(out)
				for v := range in {
					time.Sleep(sleepPerStage)
					out <- v.(int) * 2
				}
			}()
			return out
		}

		in := make(Bi)
		done := make(Bi)
		close(done)

		go func() {
			for i := 0; i < 10; i++ {
				in <- i
			}
			close(in)
		}()

		result := make([]int, 0)
		start := time.Now()
		for v := range ExecutePipeline(in, done, stage) {
			result = append(result, v.(int))
		}
		elapsed := time.Since(start)
		require.Empty(t, result)
		require.Less(t, elapsed, sleepPerStage/2)
	})

	t.Run("done during processing", func(t *testing.T) {
		var processedCount int32 // исправлено: атомарный счетчик
		stage := func(in In) Out {
			out := make(Bi)
			go func() {
				defer close(out)
				for v := range in {
					atomic.AddInt32(&processedCount, 1) // атомарный инкремент
					time.Sleep(sleepPerStage)
					out <- v.(int) * 2
				}
			}()
			return out
		}

		in := make(Bi)
		done := make(Bi)

		go func() {
			time.Sleep(sleepPerStage * 3)
			close(done)
		}()

		go func() {
			for i := 0; i < 10; i++ {
				in <- i
				time.Sleep(sleepPerStage / 10)
			}
			close(in)
		}()

		result := make([]int, 0)
		for v := range ExecutePipeline(in, done, stage) {
			result = append(result, v.(int))
		}

		require.Less(t, len(result), 10)
		require.Greater(t, atomic.LoadInt32(&processedCount), int32(0)) // атомарное чтение
	})
}

func TestLargeDataset(t *testing.T) {
	t.Run("performance with large dataset", func(t *testing.T) {
		stage := func(in In) Out {
			out := make(Bi)
			go func() {
				defer close(out)
				for v := range in {
					out <- v.(int) + 1
				}
			}()
			return out
		}

		stages := []Stage{stage, stage, stage, stage}
		dataSize := 10000

		in := make(Bi)
		go func() {
			for i := 0; i < dataSize; i++ {
				in <- i
			}
			close(in)
		}()

		result := make([]int, 0, dataSize)
		start := time.Now()
		for v := range ExecutePipeline(in, nil, stages...) {
			result = append(result, v.(int))
		}
		elapsed := time.Since(start)

		require.Len(t, result, dataSize)
		for i, v := range result {
			require.Equal(t, i+4, v)
		}
		require.Less(t, elapsed, time.Second)
	})
}

func TestErrorHandling(t *testing.T) {
	t.Run("nil input channel", func(t *testing.T) {
		stage := func(in In) Out {
			out := make(Bi)
			go func() {
				defer close(out)
				for v := range in {
					out <- v.(int) * 2
				}
			}()
			return out
		}

		result := make([]int, 0)
		for v := range ExecutePipeline(nil, nil, stage) {
			result = append(result, v.(int))
		}

		require.Empty(t, result)
	})
}

func TestMemoryLeaks(t *testing.T) {
	t.Run("goroutine cleanup", func(t *testing.T) {
		initialGoroutines := runtime.NumGoroutine()

		for i := 0; i < 10; i++ {
			stage := func(in In) Out {
				out := make(Bi)
				go func() {
					defer close(out)
					for v := range in {
						out <- v.(int) * 2
					}
				}()
				return out
			}

			in := make(Bi)
			done := make(Bi)

			go func() {
				for j := 0; j < 100; j++ {
					in <- j
				}
				close(in)
			}()

			go func() {
				time.Sleep(time.Millisecond)
				close(done)
			}()

			for range ExecutePipeline(in, done, stage, stage, stage) {
			}
		}

		time.Sleep(time.Millisecond * 100)
		runtime.GC()
		time.Sleep(time.Millisecond * 100)

		finalGoroutines := runtime.NumGoroutine()
		require.Less(t, finalGoroutines-initialGoroutines, 5)
	})
}
