package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	if in == nil {
		out := make(chan interface{})
		close(out)
		return out
	}
	if len(stages) == 0 {
		return in
	}

	current := wrapWithDone(in, done)

	for _, stage := range stages {
		current = wrapWithDone(stage(current), done)
	}

	return current
}

func wrapWithDone(in In, done In) Out {
	out := make(chan interface{})

	go func() {
		defer func() {
			close(out)
			// Дренируем входной канал, чтобы разблокировать предыдущие стейджи
			for range in {
			}
		}()

		for {
			select {
			case <-done:
				return
			case v, ok := <-in:
				if !ok {
					return
				}

				select {
				case out <- v:
				case <-done:
					return
				}
			}
		}
	}()

	return out
}
