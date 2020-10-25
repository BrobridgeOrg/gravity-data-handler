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
			id:         i,
			bufferSize: opts.BufferSize,
			handler:    opts.Handler,
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

func (pm *Manager) ComputeWorkerID(pipelineID int32) int32 {
	if pipelineID == -1 {
		return -1
	}

	return jump.Hash(uint64(pipelineID), pm.options.WorkerCount)
}

func (pm *Manager) Push(pipelineID int32, data interface{}) {

	if pipelineID == -1 {
		pm.Dispatch(data)
		return
	}

	workerID := pm.ComputeWorkerID(pipelineID)

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
