package pipeline

import (
	"sync/atomic"

	jump "github.com/lithammer/go-jump-consistent-hash"
)

type Manager struct {
	options *Options
	workers []*Worker
	counter int32
}

func NewManager(opts *Options) *Manager {

	// Initialize piplines
	workers := make([]*Worker, 0, opts.WorkerCount)
	for i := int32(0); i < opts.WorkerCount; i++ {

		worker := &Worker{
			id:             i,
			bufferSize:     opts.BufferSize,
			prepareHandler: opts.PrepareHandler,
			handler:        opts.Handler,
		}

		worker.initialize()

		workers = append(workers, worker)
	}

	return &Manager{
		options: opts,
		workers: workers,
		counter: 0,
	}
}

func (pm *Manager) ComputePipelineID(key string) int32 {
	if len(key) == 0 {
		return -1
	}

	return jump.HashString(key, pm.options.Caps, jump.NewCRC64())
}

func (pm *Manager) ComputeWorkerID(key string) int32 {
	if len(key) == 0 {
		return -1
	}

	return jump.HashString(key, pm.options.WorkerCount, jump.NewCRC64())
}

func (pm *Manager) Push(key string, data interface{}) {

	workerID := pm.ComputeWorkerID(key)

	if workerID == -1 {
		pm.Dispatch(data)
		return
	}

	// Push data to worker
	pm.workers[workerID].input <- data
}

func (pm *Manager) Dispatch(data interface{}) {

	// Push data to pipeline
	pm.workers[pm.counter].input <- data
	// Update counter
	counter := atomic.AddInt32((*int32)(&pm.counter), 1)
	if counter == pm.options.WorkerCount {
		atomic.StoreInt32((*int32)(&pm.counter), 0)
	}
}
