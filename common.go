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
	"errors"
	"reflect"
)

// Various errors that may be returned by Worker.Call.
var (
	ErrWorkerClosed  = errors.New("Worker has been closed by a call to Wait")
	ErrWouldDeadlock = errors.New("Called Wait from Integrate; would deadlock")
)

// workerState describes the state of the worker.
type workerState int

// Worker state values.
const (
	workerNew     workerState = iota // Worker is in the new state
	workerRunning                    // Worker has been started and is running
	workerClosed                     // Worker has been closed, but no results
	workerResult                     // Worker has been closed, results available
)

// selector groups together a reflect.SelectCase and a function that
// will be called if that case matches.
type selector struct {
	selectCase reflect.SelectCase                 // The select case
	fn         func(value reflect.Value, ok bool) // The function to call
}

// selectSend generates a selector value that may be used with
// doSelect for sending a value on a channel.
func selectSend(channel interface{}, value interface{}, fn func()) selector {
	return selector{
		selectCase: reflect.SelectCase{
			Dir:  reflect.SelectSend,
			Chan: reflect.ValueOf(channel),
			Send: reflect.ValueOf(value),
		},
		fn: func(value reflect.Value, ok bool) {
			fn()
		},
	}
}

// selectRecv generates a selector value that may be used with
// doSelect for receiving a value from a channel.
func selectRecv(channel interface{}, fn func(value interface{}, ok bool)) selector {
	return selector{
		selectCase: reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(channel),
		},
		fn: func(value reflect.Value, ok bool) {
			fn(value.Interface(), ok)
		},
	}
}

// doSelect takes a list of selector instances, performs the select,
// and arranges to have the appropriate function called when the
// select returns.
func doSelect(selectors []selector) {
	// Construct a list of select cases
	cases := make([]reflect.SelectCase, len(selectors))
	for i, sel := range selectors {
		cases[i] = sel.selectCase
	}

	// Run the select
	chosen, value, ok := reflect.Select(cases)

	// Call the appropriate function
	selectors[chosen].fn(value, ok)
}
