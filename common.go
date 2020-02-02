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
	ErrClosed        = errors.New("Object has been closed by a call to Wait")
	ErrWouldDeadlock = errors.New("Called Wait from Integrate; would deadlock")
)

// Result describes a result from calling a Run or Do function.  These
// functions are called in such a way as to capture panics, and the
// Result structure will contain both the return value and the
// captured panic.
type Result struct {
	Result interface{} // The function result
	Panic  interface{} // The captured panic
}

// panicer wraps a Run method and captures any panics caused within
// it.
func panicer(fn func(interface{}) interface{}, data interface{}) (result *Result) {
	// Ensure we capture panics
	defer func() {
		if panicData := recover(); panicData != nil {
			result = &Result{Panic: panicData}
		}
	}()

	return &Result{Result: fn(data)}
}

// pState describes the state of the worker or serializer.
type pState int

// Worker state values.
const (
	pNew     pState = iota // New state, not started
	pRunning               // Running state
	pClosed                // Closed state, result hasn't been received yet
	pResult                // Result state, result has been received
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
