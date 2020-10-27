package pipeline

import (
	"time"

	log "github.com/sirupsen/logrus"
)

type Worker struct {
	id             int32
	bufferSize     int
	input          chan interface{}
	readyMsgs      chan interface{}
	prepareHandler func(int32, interface{}) (interface{}, error)
	handler        func(int32, interface{}) error
}

func (worker *Worker) initialize() {

	worker.input = make(chan interface{}, worker.bufferSize)
	worker.readyMsgs = make(chan interface{}, worker.bufferSize)

	go func() {

		for {
			select {
			case data := <-worker.input:
				worker.prepare(data)
			}
		}
	}()

	go func() {

		for {
			select {
			case data := <-worker.readyMsgs:
				worker.handle(data)
			}
		}
	}()
}

func (worker *Worker) prepare(data interface{}) error {

	for {

		msg, err := worker.prepareHandler(worker.id, data)
		if err == nil {
			worker.readyMsgs <- msg
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
