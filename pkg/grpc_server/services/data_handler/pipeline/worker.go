package pipeline

import (
	"time"

	log "github.com/sirupsen/logrus"
)

type Worker struct {
	id         int32
	bufferSize int
	input      chan interface{}
	handler    func(int32, interface{}) error
}

func (worker *Worker) initialize() {

	worker.input = make(chan interface{}, worker.bufferSize)

	go func() {

		for {
			select {
			case data := <-worker.input:
				worker.handle(data)
			}
		}
	}()
}

func (worker *Worker) handle(data interface{}) error {

	for {
		err := worker.handler(worker.id, data)
		if err == nil {
			break
		}

		log.WithFields(log.Fields{
			"worker": worker.id,
		}).Error(err)

		// Retry in second
		time.Sleep(time.Second)
	}

	return nil
}
