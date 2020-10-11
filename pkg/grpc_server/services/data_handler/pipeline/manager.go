package pipeline

import (
	"sync/atomic"

	jump "github.com/lithammer/go-jump-consistent-hash"
)

type Manager struct {
	options   *Options
	pipelines []*Pipeline
	counter   int32
}

func NewManager(opts *Options) *Manager {

	// Initialize piplines
	pipelines := make([]*Pipeline, 0, opts.Caps)
	for i := int32(0); i < opts.Caps; i++ {

		pipeline := &Pipeline{
			id:         i,
			bufferSize: opts.BufferSize,
			handler:    opts.Handler,
		}

		pipeline.initialize()

		pipelines = append(pipelines, pipeline)
	}

	return &Manager{
		options:   opts,
		pipelines: pipelines,
		counter:   0,
	}
}

func (pm *Manager) Push(key string, data interface{}) {

	// Figure out pipeline we will use
	id := jump.HashString(key, pm.options.Caps, jump.NewCRC64())

	// Push data to pipeline
	pm.pipelines[id].input <- data
}

func (pm *Manager) Dispatch(data interface{}) {

	// Push data to pipeline
	pm.pipelines[pm.counter].input <- data
	// Update counter
	counter := atomic.AddInt32((*int32)(&pm.counter), 1)
	if counter == pm.options.Caps {
		atomic.StoreInt32((*int32)(&pm.counter), 0)
	}
}
