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
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSerializerImplementsSerializer(t *testing.T) {
	assert.Implements(t, (*Serializer)(nil), &serializer{})
}

func TestNewSerializer(t *testing.T) {
	doer := &MockDoer{}

	result := NewSerializer(doer)

	s, ok := result.(*serializer)
	require.True(t, ok)
	assert.Same(t, doer, s.doer)
	assert.NotNil(t, s.request)
	assert.NotNil(t, s.done)
	assert.Equal(t, &sync.Once{}, s.gonner)
}

func TestSerializerManagerBase(t *testing.T) {
	doer := &MockDoer{}
	doer.On("Do", "data").Return("result")
	obj := &serializer{
		doer:    doer,
		request: make(chan doRequest, 1),
		done:    make(chan bool, 1),
	}
	result := make(chan *Result, 1)
	obj.request <- doRequest{
		data:   "data",
		result: result,
	}
	close(obj.request)

	obj.manager()

	response := <-result
	assert.Equal(t, &Result{Result: "result"}, response)
	done := <-obj.done
	assert.True(t, done)
	doer.AssertExpectations(t)
}

func TestSerializerManagerNilResult(t *testing.T) {
	doer := &MockDoer{}
	doer.On("Do", "data").Return("result")
	obj := &serializer{
		doer:    doer,
		request: make(chan doRequest, 1),
		done:    make(chan bool, 1),
	}
	obj.request <- doRequest{data: "data"}
	close(obj.request)

	obj.manager()

	done := <-obj.done
	assert.True(t, done)
	doer.AssertExpectations(t)
}

func TestSerializerGetResult(t *testing.T) {
	doer := &MockDoer{}
	doer.On("Finish").Return("result")
	obj := &serializer{
		doer: doer,
	}

	obj.getResult()

	assert.Equal(t, "result", obj.result)
	assert.Equal(t, pResult, obj.state)
	doer.AssertExpectations(t)
}

func TestSerializerCallNew(t *testing.T) {
	doer := &MockDoer{}
	doer.On("Do", "data").Return("result")
	obj := &serializer{
		doer:    doer,
		request: make(chan doRequest, 1),
		done:    make(chan bool, 1),
	}

	result, err := obj.Call("data")

	assert.NoError(t, err)
	assert.Equal(t, &Result{Result: "result"}, result)
	assert.Equal(t, pRunning, obj.state)
	close(obj.request) // kill manager
	done := <-obj.done
	assert.True(t, done)
	doer.AssertExpectations(t)
}

func TestSerializerCallRunning(t *testing.T) {
	doer := &MockDoer{}
	doer.On("Do", "data").Return("result")
	obj := &serializer{
		state:   pRunning,
		doer:    doer,
		request: make(chan doRequest, 1),
		done:    make(chan bool, 1),
	}
	go obj.manager() // need the manager running for this

	result, err := obj.Call("data")

	assert.NoError(t, err)
	assert.Equal(t, &Result{Result: "result"}, result)
	assert.Equal(t, pRunning, obj.state)
	close(obj.request) // kill manager
	done := <-obj.done
	assert.True(t, done)
	doer.AssertExpectations(t)
}

func TestSerializerCallClosed(t *testing.T) {
	obj := &serializer{
		state: pClosed,
	}

	result, err := obj.Call("data")

	assert.Same(t, ErrClosed, err)
	assert.Nil(t, result)
	assert.Equal(t, pClosed, obj.state)
}

func TestSerializerCallResult(t *testing.T) {
	obj := &serializer{
		state: pResult,
	}

	result, err := obj.Call("data")

	assert.Same(t, ErrClosed, err)
	assert.Nil(t, result)
	assert.Equal(t, pResult, obj.state)
}

func TestSerializerCallAsyncNew(t *testing.T) {
	doer := &MockDoer{}
	doer.On("Do", "data").Return("result")
	obj := &serializer{
		doer:    doer,
		request: make(chan doRequest, 1),
		done:    make(chan bool, 1),
	}

	result, err := obj.CallAsync("data")

	assert.NoError(t, err)
	require.NotNil(t, result)
	cr := result.(*callResult)
	response := <-cr.response
	assert.Equal(t, &Result{Result: "result"}, response)
	assert.Equal(t, pRunning, obj.state)
	close(obj.request) // kill manager
	done := <-obj.done
	assert.True(t, done)
	doer.AssertExpectations(t)
}

func TestSerializerCallAsyncRunning(t *testing.T) {
	doer := &MockDoer{}
	doer.On("Do", "data").Return("result")
	obj := &serializer{
		state:   pRunning,
		doer:    doer,
		request: make(chan doRequest, 1),
		done:    make(chan bool, 1),
	}
	go obj.manager() // need the manager running for this

	result, err := obj.CallAsync("data")

	assert.NoError(t, err)
	require.NotNil(t, result)
	cr := result.(*callResult)
	response := <-cr.response
	assert.Equal(t, &Result{Result: "result"}, response)
	assert.Equal(t, pRunning, obj.state)
	close(obj.request) // kill manager
	done := <-obj.done
	assert.True(t, done)
	doer.AssertExpectations(t)
}

func TestSerializerCallAsyncClosed(t *testing.T) {
	obj := &serializer{
		state: pClosed,
	}

	result, err := obj.CallAsync("data")

	assert.Same(t, ErrClosed, err)
	assert.Nil(t, result)
	assert.Equal(t, pClosed, obj.state)
}

func TestSerializerCallAsyncResult(t *testing.T) {
	obj := &serializer{
		state: pResult,
	}

	result, err := obj.CallAsync("data")

	assert.Same(t, ErrClosed, err)
	assert.Nil(t, result)
	assert.Equal(t, pResult, obj.state)
}

func TestSerializerCallOnlyNew(t *testing.T) {
	doer := &MockDoer{}
	doer.On("Do", "data").Return("result")
	obj := &serializer{
		doer:    doer,
		request: make(chan doRequest, 1),
		done:    make(chan bool, 1),
	}

	err := obj.CallOnly("data")

	assert.NoError(t, err)
	assert.Equal(t, pRunning, obj.state)
	close(obj.request) // kill manager
	done := <-obj.done
	assert.True(t, done)
	doer.AssertExpectations(t)
}

func TestSerializerCallOnlyRunning(t *testing.T) {
	doer := &MockDoer{}
	doer.On("Do", "data").Return("result")
	obj := &serializer{
		state:   pRunning,
		doer:    doer,
		request: make(chan doRequest, 1),
		done:    make(chan bool, 1),
	}
	go obj.manager() // need the manager running for this

	err := obj.CallOnly("data")

	assert.NoError(t, err)
	assert.Equal(t, pRunning, obj.state)
	close(obj.request) // kill manager
	done := <-obj.done
	assert.True(t, done)
	doer.AssertExpectations(t)
}

func TestSerializerCallOnlyClosed(t *testing.T) {
	obj := &serializer{
		state: pClosed,
	}

	err := obj.CallOnly("data")

	assert.Same(t, ErrClosed, err)
	assert.Equal(t, pClosed, obj.state)
}

func TestSerializerCallOnlyResult(t *testing.T) {
	obj := &serializer{
		state: pResult,
	}

	err := obj.CallOnly("data")

	assert.Same(t, ErrClosed, err)
	assert.Equal(t, pResult, obj.state)
}

func TestSerializerWaitNew(t *testing.T) {
	doer := &MockDoer{}
	obj := &serializer{
		doer:   doer,
		gonner: &sync.Once{},
	}
	doer.On("Finish").Return("result").Run(func(args mock.Arguments) {
		assert.Equal(t, pClosed, obj.state)
	})

	result := obj.Wait()

	assert.Equal(t, "result", result)
	assert.Equal(t, pResult, obj.state)
	assert.Equal(t, "result", obj.result)
	doer.AssertExpectations(t)
}

func TestSerializerWaitRunning(t *testing.T) {
	doer := &MockDoer{}
	obj := &serializer{
		state:   pRunning,
		doer:    doer,
		request: make(chan doRequest, 1),
		done:    make(chan bool, 1),
		gonner:  &sync.Once{},
	}
	obj.done <- true
	doer.On("Finish").Return("result").Run(func(args mock.Arguments) {
		assert.Equal(t, pClosed, obj.state)
	})

	result := obj.Wait()

	assert.Equal(t, "result", result)
	assert.Equal(t, pResult, obj.state)
	assert.Equal(t, "result", obj.result)
	doer.AssertExpectations(t)
}

func TestSerializerWaitClosed(t *testing.T) {
	doer := &MockDoer{}
	obj := &serializer{
		state:  pClosed,
		doer:   doer,
		gonner: &sync.Once{},
	}
	doer.On("Finish").Return("result").Run(func(args mock.Arguments) {
		assert.Equal(t, pClosed, obj.state)
	})

	result := obj.Wait()

	assert.Equal(t, "result", result)
	assert.Equal(t, pResult, obj.state)
	assert.Equal(t, "result", obj.result)
	doer.AssertExpectations(t)
}

func TestSerializerWaitResult(t *testing.T) {
	doer := &MockDoer{}
	obj := &serializer{
		state:  pResult,
		doer:   doer,
		result: "result",
	}

	result := obj.Wait()

	assert.Equal(t, "result", result)
	assert.Equal(t, pResult, obj.state)
	assert.Equal(t, "result", obj.result)
	doer.AssertExpectations(t)
}

func TestCallResultImplementsCallResult(t *testing.T) {
	assert.Implements(t, (*CallResult)(nil), &callResult{})
}

func TestCallResultWaitBase(t *testing.T) {
	response := make(chan *Result, 1)
	response <- &Result{Result: "result"}
	obj := &callResult{response: response}

	result := obj.Wait()

	assert.Equal(t, &Result{Result: "result"}, result)
	assert.True(t, obj.closed)
}

func TestCallResultWaitClosed(t *testing.T) {
	obj := &callResult{closed: true}

	result := obj.Wait()

	assert.Nil(t, result)
	assert.True(t, obj.closed)
}

func TestCallResultTryWaitReady(t *testing.T) {
	response := make(chan *Result, 1)
	response <- &Result{Result: "result"}
	obj := &callResult{response: response}

	result, ok := obj.TryWait()

	assert.Equal(t, &Result{Result: "result"}, result)
	assert.True(t, ok)
	assert.True(t, obj.closed)
}

func TestCallResultTryWaitNotReady(t *testing.T) {
	response := make(chan *Result, 1)
	obj := &callResult{response: response}

	result, ok := obj.TryWait()

	assert.Nil(t, result)
	assert.True(t, ok)
	assert.False(t, obj.closed)
}

func TestCallResultTryWaitClosed(t *testing.T) {
	obj := &callResult{closed: true}

	result, ok := obj.TryWait()

	assert.Nil(t, result)
	assert.False(t, ok)
	assert.True(t, obj.closed)
}

func TestCallResultChannelBase(t *testing.T) {
	response := make(chan *Result, 1)
	obj := &callResult{response: response}

	channel := obj.Channel()

	assert.Equal(t, (<-chan *Result)(response), channel)
	assert.True(t, obj.closed)
}

func TestCallResultChannelClosed(t *testing.T) {
	obj := &callResult{closed: true}

	channel := obj.Channel()

	assert.Nil(t, channel)
	assert.True(t, obj.closed)
}
