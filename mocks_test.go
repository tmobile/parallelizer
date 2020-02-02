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
