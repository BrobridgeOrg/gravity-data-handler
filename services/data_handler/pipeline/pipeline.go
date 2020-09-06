package pipeline

type Pipeline struct {
	bufferSize int
	input      chan interface{}
	handler    func(interface{}) error
}

func (pipeline *Pipeline) initialize() {

	go func() {
		pipeline.input = make(chan interface{}, pipeline.bufferSize)

		for {
			select {
			case data := <-pipeline.input:
				pipeline.handle(data)
			}
		}
	}()
}

func (pipeline *Pipeline) handle(data interface{}) error {
	return pipeline.handler(data)
}
