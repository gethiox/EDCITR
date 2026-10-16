package parser

import (
	"sync"
)

type ProcessLineFunc[T any] func(line string) (T, error)

type Result[ResultT any] struct {
	Result ResultT
	Error  error
}

type CountResult struct {
	Success uint64
	Fail    uint64
	Bytes   uint64
}

// LineProcessor allows to parallelize ProcessLineFunc execution on every line up to the number of defined workers.
type LineProcessor[ResultT any] struct {
	workers int

	currentWorker int
	wg            sync.WaitGroup

	inputs  []chan string
	outputs []chan Result[ResultT]
	final   chan Result[ResultT]

	counterSize    []uint64
	counterSuccess []uint64
	counterFail    []uint64
}

func NewLineProcessor[T any](workers, workerChanSize, resultChanSize int, processFunc ProcessLineFunc[T]) *LineProcessor[T] {
	lp := &LineProcessor[T]{
		workers: workers,

		currentWorker: 0,
		wg:            sync.WaitGroup{},

		inputs:  make([]chan string, workers),
		outputs: make([]chan Result[T], workers),
		final:   make(chan Result[T], resultChanSize),

		counterSize:    make([]uint64, workers),
		counterSuccess: make([]uint64, workers),
		counterFail:    make([]uint64, workers),
	}

	for workerID := 0; workerID < workers; workerID++ {
		lp.inputs[workerID] = make(chan string, workerChanSize)
		lp.outputs[workerID] = make(chan Result[T], workerChanSize)

		lp.wg.Add(1)
		go func(workerID int) {
			defer lp.wg.Done()
			defer close(lp.outputs[workerID])

			for rawLine := range lp.inputs[workerID] {
				ev, err := processFunc(rawLine)
				lp.outputs[workerID] <- Result[T]{ev, err}
				if err != nil {
					lp.counterFail[workerID]++
				} else {
					lp.counterSuccess[workerID]++
				}
				lp.counterSize[workerID] += uint64(len(rawLine))
			}
		}(workerID)
	}

	lp.wg.Add(1)
	go func() {
		defer lp.wg.Done()
		defer close(lp.final)

		var currentWorkerID int
		var openWorkers = workers
		var closedWorkers = make([]bool, workers)

		for openWorkers > 0 {
			if closedWorkers[currentWorkerID] {
				currentWorkerID = (currentWorkerID + 1) % workers
				continue
			}

			res, ok := <-lp.outputs[currentWorkerID]
			if !ok {
				closedWorkers[currentWorkerID] = true
				openWorkers--
				currentWorkerID = (currentWorkerID + 1) % workers
				continue
			}

			lp.final <- res
			currentWorkerID = (currentWorkerID + 1) % workers
		}
	}()

	return lp
}

// Push adds a line to the processing queue.
// Warning: not thread-safe (must be called from a single producer synchronously).
func (lp *LineProcessor[ResultT]) Push(line string) {
	lp.inputs[lp.currentWorker] <- line
	lp.currentWorker = (lp.currentWorker + 1) % lp.workers
}

func (lp *LineProcessor[ResultT]) Results() <-chan Result[ResultT] {
	return lp.final
}

// Close will block until all provided data is processed and sent to the result channel,
// once all data has been sent, result channel is closed.
func (lp *LineProcessor[ResultT]) Close() CountResult {
	for _, ch := range lp.inputs {
		close(ch)
	}
	lp.wg.Wait()
	return lp.summarizeCounters()
}

func (lp *LineProcessor[T]) summarizeCounters() CountResult {
	result := CountResult{}

	for _, v := range lp.counterSize {
		result.Bytes += v
	}
	for _, v := range lp.counterSuccess {
		result.Success += v
	}
	for _, v := range lp.counterFail {
		result.Fail += v
	}

	return result
}
