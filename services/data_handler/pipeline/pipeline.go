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

	pipeline.input = make(chan interface{}, pipeline.bufferSize)

	go func() {

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

		log.WithFields(log.Fields{
			"pipeline": pipeline.id,
		}).Error(err)

		// Retry in second
		time.Sleep(time.Second)
	}

	return nil
}
