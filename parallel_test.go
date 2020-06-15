// Copyright (c) 2020 T-Mobile
//
// Licensed under the Apache License, Version 2.0 (the "License"); you
// may not use this file except in compliance with the License.  You
// may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied.  See the License for the specific language governing
// permissions and limitations under the License.

package parallelizer

import (
	"container/list"
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestParallelWorkerImplementsWorker(t *testing.T) {
	assert.Implements(t, (*Worker)(nil), &parallelWorker{})
}

func TestNewParallelWorkerBase(t *testing.T) {
	runner := &MockRunner{}

	result := NewParallelWorker(runner, 5)

	assert.Equal(t, &parallelWorker{
		runner:  runner,
		workers: 5,
		gonner:  &sync.Once{},
	}, result)
}

func TestNewParallelWorkerNumCPU(t *testing.T) {
	runner := &MockRunner{}

	result := NewParallelWorker(runner, 0)

	assert.Equal(t, &parallelWorker{
		runner:  runner,
		workers: runtime.NumCPU(),
		gonner:  &sync.Once{},
	}, result)
}

func TestParallelWorkerStartManager(t *testing.T) {
	obj := &parallelWorker{
		workers: 3,
	}

	obj.startManager()

	assert.NotNil(t, obj.manager)
	assert.Same(t, obj, obj.manager.worker)
	assert.Equal(t, 0, obj.manager.queue.Len())
	assert.NotNil(t, obj.manager.submit)
	assert.NotNil(t, obj.manager.work)
	assert.NotNil(t, obj.manager.results)
	assert.NotNil(t, obj.manager.done)
	obj.manager.submit <- &managerItem{done: true}
	value := <-obj.manager.done
	assert.True(t, value)
}

func TestParallelWorkerGetResult(t *testing.T) {
	runner := &MockRunner{}
	runner.On("Result").Return("result")
	obj := &parallelWorker{
		runner: runner,
	}

	obj.getResult()

	assert.Equal(t, "result", obj.result)
	assert.Equal(t, pResult, obj.state)
	runner.AssertExpectations(t)
}

func TestParallelWorkerCallNew(t *testing.T) {
	runner := &MockRunner{}
	runner.On("Run", "data").Return("result")
	runner.On("Integrate", mock.Anything, &Result{Result: "result"})
	obj := &parallelWorker{
		runner:  runner,
		workers: 3,
	}

	err := obj.Call("data")

	assert.NoError(t, err)
	assert.Equal(t, pRunning, obj.state)
	assert.NotNil(t, obj.manager)
	obj.manager.submit <- &managerItem{done: true}
	value := <-obj.manager.done
	assert.True(t, value)
	runner.AssertExpectations(t)
}

func TestParallelWorkerCallRunning(t *testing.T) {
	manager := &parallelManager{
		submit: make(chan *managerItem, 1),
	}
	obj := &parallelWorker{
		state:   pRunning,
		manager: manager,
	}

	err := obj.Call("data")

	assert.NoError(t, err)
	assert.Equal(t, pRunning, obj.state)
	assert.Same(t, manager, obj.manager)
	value := <-manager.submit
	assert.Equal(t, &managerItem{data: "data"}, value)
}

func TestParallelWorkerCallClosed(t *testing.T) {
	obj := &parallelWorker{
		state: pClosed,
	}

	err := obj.Call("data")

	assert.Same(t, ErrClosed, err)
	assert.Equal(t, pClosed, obj.state)
	assert.Nil(t, obj.manager)
}

func TestParallelWorkerCallResult(t *testing.T) {
	obj := &parallelWorker{
		state: pResult,
	}

	err := obj.Call("data")

	assert.Same(t, ErrClosed, err)
	assert.Equal(t, pResult, obj.state)
	assert.Nil(t, obj.manager)
}

func TestParallelWorkerWaitNew(t *testing.T) {
	runner := &MockRunner{}
	obj := &parallelWorker{
		runner: runner,
		gonner: &sync.Once{},
	}
	runner.On("Result").Return("result").Run(func(args mock.Arguments) {
		assert.Equal(t, pClosed, obj.state)
	})

	result, err := obj.Wait()

	assert.NoError(t, err)
	assert.Equal(t, "result", result)
	assert.Equal(t, pResult, obj.state)
	assert.Nil(t, obj.manager)
	assert.Equal(t, "result", obj.result)
	runner.AssertExpectations(t)
}

func TestParallelWorkerWaitRunning(t *testing.T) {
	runner := &MockRunner{}
	manager := &parallelManager{
		submit: make(chan *managerItem, 1),
		done:   make(chan bool, 1),
	}
	obj := &parallelWorker{
		state:   pRunning,
		runner:  runner,
		manager: manager,
		gonner:  &sync.Once{},
	}
	obj.manager.done <- true
	runner.On("Result").Return("result").Run(func(args mock.Arguments) {
		assert.Equal(t, pClosed, obj.state)
	})

	result, err := obj.Wait()

	assert.NoError(t, err)
	assert.Equal(t, "result", result)
	assert.Equal(t, pResult, obj.state)
	assert.Nil(t, obj.manager)
	assert.Equal(t, "result", obj.result)
	runner.AssertExpectations(t)
	submission := <-manager.submit
	assert.Equal(t, &managerItem{done: true}, submission)
}

func TestParallelWorkerWaitClosed(t *testing.T) {
	runner := &MockRunner{}
	obj := &parallelWorker{
		state:  pClosed,
		runner: runner,
		gonner: &sync.Once{},
	}
	runner.On("Result").Return("result").Run(func(args mock.Arguments) {
		assert.Equal(t, pClosed, obj.state)
	})

	result, err := obj.Wait()

	assert.NoError(t, err)
	assert.Equal(t, "result", result)
	assert.Equal(t, pResult, obj.state)
	assert.Nil(t, obj.manager)
	assert.Equal(t, "result", obj.result)
	runner.AssertExpectations(t)
}

func TestParallelWorkerWaitResult(t *testing.T) {
	runner := &MockRunner{}
	obj := &parallelWorker{
		state:  pResult,
		runner: runner,
		result: "result",
	}

	result, err := obj.Wait()

	assert.NoError(t, err)
	assert.Equal(t, "result", result)
	assert.Equal(t, pResult, obj.state)
	assert.Nil(t, obj.manager)
	assert.Equal(t, "result", obj.result)
	runner.AssertExpectations(t)
}

func TestParallelManagerImplementsWorker(t *testing.T) {
	assert.Implements(t, (*Worker)(nil), &parallelManager{})
}

func TestParallelManagerWorkRunner(t *testing.T) {
	work := make(chan interface{}, 2)
	work <- "data1"
	work <- "data2"
	close(work)
	runner := &MockRunner{}
	runner.On("Run", "data1").Return("result1")
	runner.On("Run", "data2").Return("result2")
	obj := &parallelManager{
		worker: &parallelWorker{
			runner: runner,
		},
		results: make(chan *managerItem, 3),
	}

	obj.workRunner(work)

	close(obj.results)
	results := []*managerItem{}
	for item := range obj.results {
		results = append(results, item)
	}
	assert.Equal(t, []*managerItem{
		{data: &Result{Result: "result1"}},
		{data: &Result{Result: "result2"}},
		{done: true},
	}, results)
	runner.AssertExpectations(t)
}

func TestParallelManagerStartWorkers(t *testing.T) {
	obj := &parallelManager{
		worker: &parallelWorker{
			workers: 3,
		},
		work:    make(chan interface{}),
		results: make(chan *managerItem),
	}

	obj.startWorkers()

	assert.Equal(t, 3, obj.count)
	close(obj.work)
	for i := 0; i < 3; i++ {
		result := <-obj.results
		assert.Equal(t, &managerItem{done: true}, result)
	}
}

func TestParallelManagerReceiveWorkItem(t *testing.T) {
	obj := &parallelManager{
		queue: &list.List{},
	}
	item := &managerItem{data: "data"}

	obj.receiveWork(item)

	assert.False(t, obj.exiting)
	assert.Equal(t, 1, obj.queue.Len())
	assert.Equal(t, "data", obj.queue.Front().Value)
}

func TestParallelManagerReceiveWorkExit(t *testing.T) {
	obj := &parallelManager{
		queue: &list.List{},
	}
	item := &managerItem{done: true}

	obj.receiveWork(item)

	assert.True(t, obj.exiting)
	assert.Equal(t, 0, obj.queue.Len())
}

func TestParallelManagerReceiveResultData(t *testing.T) {
	runner := &MockRunner{}
	obj := &parallelManager{
		worker: &parallelWorker{
			runner: runner,
		},
		count:   5,
		waiting: 3,
	}
	runner.On("Integrate", obj, &Result{Result: "data"})
	item := &managerItem{data: &Result{Result: "data"}}

	obj.receiveResult(item)

	assert.Equal(t, 5, obj.count)
	assert.Equal(t, 2, obj.waiting)
	runner.AssertExpectations(t)
}

func TestParallelManagerReceiveResultDone(t *testing.T) {
	runner := &MockRunner{}
	obj := &parallelManager{
		worker: &parallelWorker{
			runner: runner,
		},
		count:   5,
		waiting: 3,
	}
	item := &managerItem{done: true}

	obj.receiveResult(item)

	assert.Equal(t, 4, obj.count)
	assert.Equal(t, 3, obj.waiting)
	runner.AssertExpectations(t)
}

func TestParallelManagerManagerSelectSendWork(t *testing.T) {
	obj := &parallelManager{
		exiting: true,
		queue:   &list.List{},
		work:    make(chan interface{}, 1),
	}
	obj.queue.PushBack("data")

	result := obj.managerSelect()

	assert.True(t, result)
	work, ok := <-obj.work
	require.True(t, ok)
	assert.Equal(t, "data", work)
	assert.Equal(t, 1, obj.waiting)
}

func TestParallelManagerManagerSelectRecvWork(t *testing.T) {
	obj := &parallelManager{
		exiting: false,
		queue:   &list.List{},
		submit:  make(chan *managerItem, 1),
	}
	obj.submit <- &managerItem{data: "data"}

	result := obj.managerSelect()

	assert.True(t, result)
	assert.Equal(t, 1, obj.queue.Len())
	assert.Equal(t, "data", obj.queue.Front().Value)
}

func TestParallelManagerManagerSelectRecvResults(t *testing.T) {
	obj := &parallelManager{
		exiting: true,
		count:   1,
		waiting: 1,
		queue:   &list.List{},
		results: make(chan *managerItem, 1),
	}
	obj.results <- &managerItem{done: true}

	result := obj.managerSelect()

	assert.True(t, result)
	assert.Equal(t, 0, obj.count)
	assert.Equal(t, 1, obj.waiting)
}

func TestParallelManagerManagerSelectDone(t *testing.T) {
	obj := &parallelManager{
		exiting: true,
		queue:   &list.List{},
	}

	result := obj.managerSelect()

	assert.False(t, result)
}

func TestParallelManagerManagerBase(t *testing.T) {
	obj := &parallelManager{
		queue:  &list.List{},
		submit: make(chan *managerItem, 1),
		done:   make(chan bool, 1),
	}
	obj.submit <- &managerItem{done: true}

	obj.manager()

	assert.True(t, obj.exiting)
	done := <-obj.done
	assert.True(t, done)
}

func TestParallelManagerManagerClosesWork(t *testing.T) {
	work := make(chan interface{}, 1)
	obj := &parallelManager{
		queue:  &list.List{},
		submit: make(chan *managerItem, 1),
		work:   work,
		done:   make(chan bool, 1),
	}
	obj.submit <- &managerItem{done: true}

	obj.manager()

	assert.True(t, obj.exiting)
	assert.Nil(t, obj.work)
	_, ok := <-work
	assert.False(t, ok)
	done := <-obj.done
	assert.True(t, done)
}

func TestParallelManagerManagerStartsWorkers(t *testing.T) {
	work := make(chan interface{}, 1)
	obj := &parallelManager{
		worker:  &parallelWorker{},
		queue:   &list.List{},
		waiting: -1, // make sure manager exits
		submit:  make(chan *managerItem, 1),
		work:    work,
		done:    make(chan bool, 1),
	}
	obj.queue.PushBack("work")
	obj.submit <- &managerItem{done: true}

	obj.manager()

	assert.True(t, obj.exiting)
	assert.Nil(t, obj.work)
	data, ok := <-work
	require.True(t, ok)
	assert.Equal(t, "work", data)
	_, ok = <-work
	assert.False(t, ok)
	done := <-obj.done
	assert.True(t, done)
}

func TestParallelManagerCall(t *testing.T) {
	obj := &parallelManager{
		queue: &list.List{},
	}

	err := obj.Call("data")

	assert.NoError(t, err)
	assert.Equal(t, 1, obj.queue.Len())
	assert.Equal(t, "data", obj.queue.Front().Value)
}

func TestParallelManagerWait(t *testing.T) {
	obj := &parallelManager{}

	result, err := obj.Wait()

	assert.Same(t, ErrWouldDeadlock, err)
	assert.Nil(t, result)
}
