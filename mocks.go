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

import "github.com/stretchr/testify/mock"

// MockRunner is a mock for the Runner interface.  It is provided to
// facilitate internal testing of the Runner implementations, but may
// be used by external users to test other code that utilizes a
// Runner.
type MockRunner struct {
	mock.Mock
}

// Run is the method that will be called to actually process the data.
// It will be passed the data that was passed to Worker.Call, and may
// return data that will be subsequently passed to the Integrate
// method.  The Run method may be called from any number of goroutines
// (workers), so any resources it interacts with, including those
// embedded in the object, must be accessed in a thread-safe fashion.
//
// It is not safe for Run to make any calls to Worker.Call; this may
// potentially lead to a deadlock scenario.  Instead, return those
// items and handle the calls to Worker.Call from the Integrate
// method.
func (m *MockRunner) Run(data interface{}) interface{} {
	args := m.MethodCalled("Run", data)

	return args.Get(0)
}

// Integrate is used to combine all the data returned by Run method
// invocations.  It is passed a Worker object, which it may use to
// make additional calls to Worker.Call, even if Worker.Wait has been
// called.  All instances of Integrate operate synchronously in a
// single goroutine, and must not block; a side-effect is that the
// elements they interact with may be safely accessed without concern
// for parallel calls to Integrate.  The idea of Integrate is to allow
// the results from the various Run method calls to be combined
// together into a single result, which may then be obtained through a
// call to Result.
//
// Note that Integrate is not running in the same goroutine as that
// which is making Worker.Call calls; in fact, those calls may be from
// multiple goroutines.
func (m *MockRunner) Integrate(worker Worker, result interface{}) {
	m.MethodCalled("Integrate", worker, result)
}

// Result is called by the Worker.Wait method a single time, once all
// the worker goroutines have been terminated.  It is intended to work
// in conjunction with Integrate to enable the final result of the
// work to be reported to the caller of Worker.Wait.  It runs in the
// same goroutine as Worker.Wait, and need not worry about any other
// goroutine calling any other method from the Runner.
func (m *MockRunner) Result() interface{} {
	args := m.MethodCalled("Result")

	return args.Get(0)
}

// MockWorker is a mock for the Worker interface.  It is provided to
// facilitate testing code that utilizes Worker implementations.
type MockWorker struct {
	mock.Mock
}

// Call is the method used to submit data to be worked in a call to
// the Runner.Run method.  It may return an error if the worker has
// been shut down through a call to Wait.
func (m *MockWorker) Call(data interface{}) error {
	args := m.MethodCalled("Call", data)

	return args.Error(0)
}

// Wait is called to shut down the worker and return the final result;
// it will block the caller until all data has been processed and all
// worker goroutines have stopped.  Note that the final result,
// generated by Runner.Result, is saved by Worker to satisfy later
// calls to Wait.  If Wait is called before any calls to Call, the
// worker will go straight to a stopped state, and no further Call
// calls may be made; no error will be returned in that case.
func (m *MockWorker) Wait() (interface{}, error) {
	args := m.MethodCalled("Wait")

	return args.Get(0), args.Error(1)
}
