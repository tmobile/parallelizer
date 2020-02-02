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
		runner: runner,
		queue:  &list.List{},
	}
	runner.On("Run", "data").Return("result").Run(func(args mock.Arguments) {
		assert.True(t, obj.running)
	})
	runner.On("Integrate", obj, "result")

	err := obj.Call("data")

	assert.NoError(t, err)
	assert.False(t, obj.running)
	assert.Equal(t, 0, obj.queue.Len())
	runner.AssertExpectations(t)
}

func TestSynchronousWorkerCallRunning(t *testing.T) {
	runner := &MockRunner{}
	obj := &synchronousWorker{
		runner:  runner,
		queue:   &list.List{},
		running: true,
	}

	err := obj.Call("data")

	assert.NoError(t, err)
	assert.True(t, obj.running)
	assert.Equal(t, 1, obj.queue.Len())
	assert.Equal(t, "data", obj.queue.Front().Value)
	runner.AssertExpectations(t)
}

func TestSynchronousWorkerCallClosed(t *testing.T) {
	runner := &MockRunner{}
	obj := &synchronousWorker{
		runner: runner,
		queue:  &list.List{},
		closed: true,
	}

	err := obj.Call("data")

	assert.Same(t, ErrWorkerClosed, err)
	assert.False(t, obj.running)
	assert.Equal(t, 0, obj.queue.Len())
	runner.AssertExpectations(t)
}

func TestSynchronousWorkerWaitBase(t *testing.T) {
	runner := &MockRunner{}
	obj := &synchronousWorker{
		runner: runner,
		queue:  &list.List{},
	}
	runner.On("Result").Return("result")

	result := obj.Wait()

	assert.Equal(t, "result", result)
	assert.True(t, obj.closed)
	assert.True(t, obj.haveResult)
	assert.Equal(t, "result", obj.result)
	assert.Equal(t, 0, obj.queue.Len())
}

func TestSynchronousWorkerWaitHaveResult(t *testing.T) {
	runner := &MockRunner{}
	obj := &synchronousWorker{
		runner:     runner,
		queue:      &list.List{},
		closed:     true,
		haveResult: true,
		result:     "result",
	}

	result := obj.Wait()

	assert.Equal(t, "result", result)
	assert.True(t, obj.closed)
	assert.True(t, obj.haveResult)
	assert.Equal(t, "result", obj.result)
	assert.Equal(t, 0, obj.queue.Len())
}

func TestSynchronousWorkerWaitEnqueued(t *testing.T) {
	runner := &MockRunner{}
	obj := &synchronousWorker{
		runner: runner,
		queue:  &list.List{},
	}
	runner.On("Run", "data").Return("integrate")
	runner.On("Integrate", obj, "integrate")
	runner.On("Result").Return("result")
	obj.queue.PushBack("data")

	result := obj.Wait()

	assert.Equal(t, "result", result)
	assert.True(t, obj.closed)
	assert.True(t, obj.haveResult)
	assert.Equal(t, "result", obj.result)
	assert.Equal(t, 0, obj.queue.Len())
}

func TestSynchronousWorkerWaitEnqueuedGivesResult(t *testing.T) {
	runner := &MockRunner{}
	obj := &synchronousWorker{
		runner: runner,
		queue:  &list.List{},
	}
	runner.On("Run", "data").Return("integrate")
	runner.On("Integrate", obj, "integrate").Run(func(args mock.Arguments) {
		obj.haveResult = true
		obj.result = "result"
	})
	obj.queue.PushBack("data")

	result := obj.Wait()

	assert.Equal(t, "result", result)
	assert.True(t, obj.closed)
	assert.True(t, obj.haveResult)
	assert.Equal(t, "result", obj.result)
	assert.Equal(t, 0, obj.queue.Len())
}
