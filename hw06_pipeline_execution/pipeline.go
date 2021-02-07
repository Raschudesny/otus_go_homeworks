package hw06_pipeline_execution //nolint:golint,stylecheck

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func makeCloseOnErrorChannel(inputCh In, done In) Out {
	closeOnErrorCh := make(Bi)
	go func() {
		defer close(closeOnErrorCh)
		for {
			// try to add more priority to done channel... maybe it's unnecessary here
			select {
			case <-done:
				return
			default:
			}
			select {
			case <-done:
				return
			case input, ok := <-inputCh:
				if !ok {
					return
				}
				select {
				case <-done:
					return
				case closeOnErrorCh <- input:
				}
			}
		}
	}()
	return closeOnErrorCh
}

// Executes all stages(excepts nil stages) as a following pipeline
// in -> our -> stage -> our -> stage -> our -> ... -> stage -> our -> stage -> out.
func ExecutePipeline(in In, done In, stages ...Stage) Out {
	var prevStageOut In = in
	for _, stage := range stages {
		if stage != nil {
			prevStageOut = stage(makeCloseOnErrorChannel(prevStageOut, done))
		}
	}
	return prevStageOut
}
