package pipeline

import (
	"time"

	log "github.com/sirupsen/logrus"
)

type Pipeline struct {
	id         int32
	bufferSize int
	input      chan interface{}
	handler    func(int32, interface{}) error
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

	for {
		err := pipeline.handler(pipeline.id, data)
		if err == nil {
			break
		}

		log.Error(err)
		time.Sleep(time.Second)
	}

	return nil
}
