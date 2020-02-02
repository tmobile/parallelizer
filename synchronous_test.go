// Copyright (c) 2020 Kevin L. Mitchell
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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSynchronousWorkerImplementsWorker(t *testing.T) {
	assert.Implements(t, (*Worker)(nil), &synchronousWorker{})
}

func TestNewSynchronousWorker(t *testing.T) {
	runner := &MockRunner{}

	result := NewSynchronousWorker(runner)

	assert.Equal(t, &synchronousWorker{
		runner: runner,
		queue:  &list.List{},
	}, result)
}

func TestSynchronousWorkerRun(t *testing.T) {
	runner := &MockRunner{}
	obj := &synchronousWorker{
		runner: runner,
		queue:  &list.List{},
	}
	obj.queue.PushBack("value")
	runner.On("Run", "value").Return("result")
	runner.On("Integrate", obj, "result")

	obj.run()

	assert.Equal(t, 0, obj.queue.Len())
	runner.AssertExpectations(t)
}

func TestSynchronousWorkerCallBase(t *testing.T) {
	runner := &MockRunner{}
	obj := &synchronousWorker{
		state:  workerRunning,
		runner: runner,
		queue:  &list.List{},
	}
	runner.On("Run", "data").Return("result").Run(func(args mock.Arguments) {
		assert.True(t, obj.running)
	})
	runner.On("Integrate", obj, "result")

	err := obj.Call("data")

	assert.NoError(t, err)
	assert.Equal(t, workerRunning, obj.state)
	assert.False(t, obj.running)
	assert.Equal(t, 0, obj.queue.Len())
	runner.AssertExpectations(t)
}

func TestSynchronousWorkerCallRunning(t *testing.T) {
	runner := &MockRunner{}
	obj := &synchronousWorker{
		state:   workerRunning,
		runner:  runner,
		queue:   &list.List{},
		running: true,
	}

	err := obj.Call("data")

	assert.NoError(t, err)
	assert.Equal(t, workerRunning, obj.state)
	assert.True(t, obj.running)
	assert.Equal(t, 1, obj.queue.Len())
	assert.Equal(t, "data", obj.queue.Front().Value)
	runner.AssertExpectations(t)
}

func TestSynchronousWorkerCallNew(t *testing.T) {
	runner := &MockRunner{}
	obj := &synchronousWorker{
		runner: runner,
		queue:  &list.List{},
	}
	runner.On("Run", "data").Return("result").Run(func(args mock.Arguments) {
		assert.True(t, obj.running)
	})
	runner.On("Integrate", obj, "result")

	err := obj.Call("data")

	assert.NoError(t, err)
	assert.Equal(t, workerRunning, obj.state)
	assert.False(t, obj.running)
	assert.Equal(t, 0, obj.queue.Len())
	runner.AssertExpectations(t)
}

func TestSynchronousWorkerCallClosed(t *testing.T) {
	runner := &MockRunner{}
	obj := &synchronousWorker{
		state:  workerClosed,
		runner: runner,
		queue:  &list.List{},
	}

	err := obj.Call("data")

	assert.Same(t, ErrWorkerClosed, err)
	assert.Equal(t, workerClosed, obj.state)
	assert.False(t, obj.running)
	assert.Equal(t, 0, obj.queue.Len())
	runner.AssertExpectations(t)
}

func TestSynchronousWorkerCallResult(t *testing.T) {
	runner := &MockRunner{}
	obj := &synchronousWorker{
		state:  workerResult,
		runner: runner,
		queue:  &list.List{},
	}

	err := obj.Call("data")

	assert.Same(t, ErrWorkerClosed, err)
	assert.Equal(t, workerResult, obj.state)
	assert.False(t, obj.running)
	assert.Equal(t, 0, obj.queue.Len())
	runner.AssertExpectations(t)
}

func TestSynchronousWorkerCallClosedRunning(t *testing.T) {
	runner := &MockRunner{}
	obj := &synchronousWorker{
		state:   workerClosed,
		runner:  runner,
		queue:   &list.List{},
		running: true,
	}

	err := obj.Call("data")

	assert.NoError(t, err)
	assert.Equal(t, workerClosed, obj.state)
	assert.True(t, obj.running)
	assert.Equal(t, 1, obj.queue.Len())
	assert.Equal(t, "data", obj.queue.Front().Value)
	runner.AssertExpectations(t)
}

func TestSynchronousWorkerCallResultRunning(t *testing.T) {
	runner := &MockRunner{}
	obj := &synchronousWorker{
		state:   workerResult,
		runner:  runner,
		queue:   &list.List{},
		running: true,
	}

	err := obj.Call("data")

	assert.NoError(t, err)
	assert.Equal(t, workerResult, obj.state)
	assert.True(t, obj.running)
	assert.Equal(t, 1, obj.queue.Len())
	assert.Equal(t, "data", obj.queue.Front().Value)
	runner.AssertExpectations(t)
}

func TestSynchronousWorkerWaitNew(t *testing.T) {
	runner := &MockRunner{}
	obj := &synchronousWorker{
		runner: runner,
	}
	runner.On("Result").Return("result")

	result, err := obj.Wait()

	assert.NoError(t, err)
	assert.Equal(t, "result", result)
	assert.Equal(t, workerResult, obj.state)
	assert.Equal(t, "result", obj.result)
}

func TestSynchronousWorkerWaitRunning(t *testing.T) {
	runner := &MockRunner{}
	obj := &synchronousWorker{
		state:  workerRunning,
		runner: runner,
	}
	runner.On("Result").Return("result")

	result, err := obj.Wait()

	assert.NoError(t, err)
	assert.Equal(t, "result", result)
	assert.Equal(t, workerResult, obj.state)
	assert.Equal(t, "result", obj.result)
}

func TestSynchronousWorkerWaitClosed(t *testing.T) {
	runner := &MockRunner{}
	obj := &synchronousWorker{
		state:  workerClosed,
		runner: runner,
	}
	runner.On("Result").Return("result")

	result, err := obj.Wait()

	assert.NoError(t, err)
	assert.Equal(t, "result", result)
	assert.Equal(t, workerResult, obj.state)
	assert.Equal(t, "result", obj.result)
}

func TestSynchronousWorkerWaitResult(t *testing.T) {
	runner := &MockRunner{}
	obj := &synchronousWorker{
		state:  workerResult,
		runner: runner,
		result: "result",
	}

	result, err := obj.Wait()

	assert.NoError(t, err)
	assert.Equal(t, "result", result)
	assert.Equal(t, workerResult, obj.state)
	assert.Equal(t, "result", obj.result)
}

func TestSynchronousWorkerWaitDeadlock(t *testing.T) {
	runner := &MockRunner{}
	obj := &synchronousWorker{
		state:   workerRunning,
		runner:  runner,
		running: true,
		result:  "result",
	}

	result, err := obj.Wait()

	assert.Same(t, ErrWouldDeadlock, err)
	assert.Nil(t, result)
	assert.Equal(t, workerRunning, obj.state)
}
