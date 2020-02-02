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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockRunnerImplementsRunner(t *testing.T) {
	assert.Implements(t, (*Runner)(nil), &MockRunner{})
}

func TestMockRunnerRun(t *testing.T) {
	obj := &MockRunner{}
	obj.On("Run", "data").Return("result")

	result := obj.Run("data")

	assert.Equal(t, "result", result)
	obj.AssertExpectations(t)
}

func TestMockRunnerIntegrate(t *testing.T) {
	worker := &MockWorker{}
	obj := &MockRunner{}
	obj.On("Integrate", worker, &Result{Result: "result"})

	obj.Integrate(worker, &Result{Result: "result"})

	obj.AssertExpectations(t)
}

func TestMockRunnerResult(t *testing.T) {
	obj := &MockRunner{}
	obj.On("Result").Return("result")

	result := obj.Result()

	assert.Equal(t, "result", result)
	obj.AssertExpectations(t)
}

func TestMockWorkerImplementsWorker(t *testing.T) {
	assert.Implements(t, (*Worker)(nil), &MockWorker{})
}

func TestMockWorkerCall(t *testing.T) {
	obj := &MockWorker{}
	obj.On("Call", "data").Return(assert.AnError)

	err := obj.Call("data")

	assert.Same(t, assert.AnError, err)
	obj.AssertExpectations(t)
}

func TestMockWorkerWait(t *testing.T) {
	obj := &MockWorker{}
	obj.On("Wait").Return("result", assert.AnError)

	result, err := obj.Wait()

	assert.Same(t, assert.AnError, err)
	assert.Equal(t, "result", result)
	obj.AssertExpectations(t)
}

func TestMockDoerImplementsDoer(t *testing.T) {
	assert.Implements(t, (*Doer)(nil), &MockDoer{})
}

func TestMockDoerDo(t *testing.T) {
	obj := &MockDoer{}
	obj.On("Do", "data").Return("result")

	result := obj.Do("data")

	assert.Equal(t, "result", result)
	obj.AssertExpectations(t)
}

func TestMockDoerFinish(t *testing.T) {
	obj := &MockDoer{}
	obj.On("Finish").Return("result")

	result := obj.Finish()

	assert.Equal(t, "result", result)
	obj.AssertExpectations(t)
}

func TestMockCallResultImplementsCallResult(t *testing.T) {
	assert.Implements(t, (*CallResult)(nil), &MockCallResult{})
}

func TestMockCallResultWaitNil(t *testing.T) {
	obj := &MockCallResult{}
	obj.On("Wait").Return(nil)

	result := obj.Wait()

	assert.Nil(t, result)
	obj.AssertExpectations(t)
}

func TestMockCallResultWaitNonNil(t *testing.T) {
	expected := &Result{}
	obj := &MockCallResult{}
	obj.On("Wait").Return(expected)

	result := obj.Wait()

	assert.Same(t, expected, result)
	obj.AssertExpectations(t)
}

func TestMockCallResultTryWaitNil(t *testing.T) {
	obj := &MockCallResult{}
	obj.On("TryWait").Return(nil, true)

	result, ok := obj.TryWait()

	assert.Nil(t, result)
	assert.True(t, ok)
	obj.AssertExpectations(t)
}

func TestMockCallResultTryWaitNonNil(t *testing.T) {
	expected := &Result{}
	obj := &MockCallResult{}
	obj.On("TryWait").Return(expected, true)

	result, ok := obj.TryWait()

	assert.Same(t, expected, result)
	assert.True(t, ok)
	obj.AssertExpectations(t)
}

func TestMockCallResultChannelNil(t *testing.T) {
	obj := &MockCallResult{}
	obj.On("Channel").Return(nil)

	result := obj.Channel()

	assert.Nil(t, result)
	obj.AssertExpectations(t)
}

func TestMockCallResultChannelNonNil(t *testing.T) {
	expected := make(<-chan *Result)
	obj := &MockCallResult{}
	obj.On("Channel").Return(expected)

	result := obj.Channel()

	assert.Equal(t, expected, result)
	obj.AssertExpectations(t)
}

func TestMockSerializerImplementsSerializer(t *testing.T) {
	assert.Implements(t, (*Serializer)(nil), &MockSerializer{})
}

func TestMockSerializerCallNil(t *testing.T) {
	obj := &MockSerializer{}
	obj.On("Call", "data").Return(nil, assert.AnError)

	result, err := obj.Call("data")

	assert.Same(t, assert.AnError, err)
	assert.Nil(t, result)
	obj.AssertExpectations(t)
}

func TestMockSerializerCallNonNil(t *testing.T) {
	expected := &Result{}
	obj := &MockSerializer{}
	obj.On("Call", "data").Return(expected, assert.AnError)

	result, err := obj.Call("data")

	assert.Same(t, assert.AnError, err)
	assert.Same(t, expected, result)
	obj.AssertExpectations(t)
}

func TestMockSerializerCallAsyncNil(t *testing.T) {
	obj := &MockSerializer{}
	obj.On("CallAsync", "data").Return(nil, assert.AnError)

	result, err := obj.CallAsync("data")

	assert.Same(t, assert.AnError, err)
	assert.Nil(t, result)
	obj.AssertExpectations(t)
}

func TestMockSerializerCallAsyncNonNil(t *testing.T) {
	expected := &MockCallResult{}
	obj := &MockSerializer{}
	obj.On("CallAsync", "data").Return(expected, assert.AnError)

	result, err := obj.CallAsync("data")

	assert.Same(t, assert.AnError, err)
	assert.Same(t, expected, result)
	obj.AssertExpectations(t)
}

func TestMockSerializerCallOnly(t *testing.T) {
	obj := &MockSerializer{}
	obj.On("CallOnly", "data").Return(assert.AnError)

	err := obj.CallOnly("data")

	assert.Same(t, assert.AnError, err)
	obj.AssertExpectations(t)
}

func TestMockSerializerWait(t *testing.T) {
	obj := &MockSerializer{}
	obj.On("Wait").Return("result")

	result := obj.Wait()

	assert.Equal(t, "result", result)
	obj.AssertExpectations(t)
}
