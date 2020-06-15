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

// Package parallelizer is a library for enabling the addition of
// controlled parallelization utilizing a pool of worker goroutines in
// a simple manner.  This is not intended as an external job queue,
// where outside programs may submit jobs, although it could easily be
// used to implement such a tool.
//
// The parallelizer package provides a Runner interface, which is for
// client applications to implement.  Instances of the Runner
// interface may then be passed to the constructor functions
// NewSynchronousWorker or NewParallelWorker, which construct objects
// conforming to the Worker interface.  Data items may then be passed
// to the Worker instances via the Worker.Call method, and the
// processing completed and the final result obtained by calling
// Worker.Wait.
//
// A Runner implementation must provide a Runner.Run method, which
// will actually process the data in a goroutine and return a result;
// the result is then passed to the Runner.Integrate method, which is
// run synchronously with other Runner.Integrate calls, and which can
// submit additional data items for processing.  Once all data is
// processed, and the client code has called Worker.Wait, the Worker
// will call the Runner.Result method to obtain the result.  The
// Runner.Result method will be called exactly once; the returned
// value is cached in the Worker to be returned by future calls to
// Worker.Wait.  The Worker.Call method may not be called again after
// Worker.Wait has been called.
//
// The parallelizer package also provides a Doer interface, which is
// for client applications to implement.  Instances of the Doer
// interface may then be passed to the constructor function
// NewSerializer, which constructs objects conforming to the
// Serializer interface.  Data items may then be passed to the
// Serializer objects to be executed via Doer.Do in a synchronous
// manner without necessarily blocking the calling goroutine.
//
// A Doer implementation must provide a Doer.Do method, which will
// actually process the data in a separate goroutine; each call will
// be executed synchronously, so thread-safety in Doer.Do is not a
// concern.  When the code using the wrapping Serializer is done, it
// will call Serializer.Wait, which will call the Doer.Finish method
// and return its result to the caller.  Note that none of the
// Serializer.Call methods may be called again after calling
// Serializer.Wait.
package parallelizer

// Runner is an interface describing the work to be done.  A Worker is
// typically instantiated by passing it a Runner, which it will then
// use to process the submitted data.
type Runner interface {
	// Run is the method that will be called to actually process
	// the data.  It will be passed the data that was passed to
	// Worker.Call, and may return data that will be subsequently
	// passed to the Integrate method.  The Run method may be
	// called from any number of goroutines (workers), so any
	// resources it interacts with, including those embedded in
	// the object, must be accessed in a thread-safe fashion.
	//
	// It is not safe for Run to make any calls to Worker.Call;
	// this may potentially lead to a deadlock scenario.  Instead,
	// return those items and handle the calls to Worker.Call from
	// the Integrate method.
	Run(data interface{}) interface{}

	// Integrate is used to combine all the data returned by Run
	// method invocations.  It is passed a Worker object, which it
	// may use to make additional calls to Worker.Call, even if
	// Worker.Wait has been called.  All instances of Integrate
	// operate synchronously in a single goroutine, and must not
	// block; a side-effect is that the elements they interact
	// with may be safely accessed without concern for parallel
	// calls to Integrate.  The idea of Integrate is to allow the
	// results from the various Run method calls to be combined
	// together into a single result, which may then be obtained
	// through a call to Result.  Note that if Run panics, the
	// data will be passed to Integrate as the "panicData"
	// parameter, and the "result" parameter will be nil.
	//
	// Note that Integrate is not running in the same goroutine as
	// that which is making Worker.Call calls; in fact, those
	// calls may be from multiple goroutines.
	Integrate(worker Worker, result *Result)

	// Result is called by the Worker.Wait method a single time,
	// once all the worker goroutines have been terminated.  It is
	// intended to work in conjunction with Integrate to enable
	// the final result of the work to be reported to the caller
	// of Worker.Wait.  It runs in the same goroutine as
	// Worker.Wait, and need not worry about any other goroutine
	// calling any other method from the Runner.
	Result() interface{}
}

// Worker is an interface describing implementations of the
// parallelizer.  A Worker is typically initialized by passing a
// Runner instance to a constructor; data submitted with Worker.Call
// is then passed to the Runner.Run methods, the results of which are
// integrated using Runner.Integrate; finally, the caller calls
// Worker.Wait to shut down the parallelizer, which in turn will call
// Worker.Result to obtain the final result of the processing.
type Worker interface {
	// Call is the method used to submit data to be worked in a
	// call to the Runner.Run method.  It may return an error if
	// the worker has been shut down through a call to Wait.
	Call(data interface{}) error

	// Wait is called to shut down the worker and return the final
	// result; it will block the caller until all data has been
	// processed and all worker goroutines have stopped.  Note
	// that the final result, generated by Runner.Result, is saved
	// by Worker to satisfy later calls to Wait.  If Wait is
	// called before any calls to Call, the worker will go
	// straight to a stopped state, and no further Call calls may
	// be made; no error will be returned in that case.
	Wait() (interface{}, error)
}

// Doer is an interface describing an operation to be done in a
// synchronized fashion, such as building a data structure.
type Doer interface {
	// Do does some operation.  It receives some data and returns
	// some result.  A serializer wraps a Doer to ensure the
	// operation is done in a single goroutine, synchronously.
	Do(data interface{}) interface{}

	// Finish is called when the manager goroutine of a Serializer
	// implementation has been signaled to exit.  It may return a
	// value, which becomes the return value from Serializer.Wait.
	Finish() interface{}
}

// CallResult is an interface describing a "future" returned by
// Serializer.CallAsync.  It allows the call to be made without
// blocking the goroutine calling Serializer.CallAsync, but the result
// of the call may still be waited upon at a later date.
type CallResult interface {
	// Wait is used to retrieve the result of the call.  The
	// result is not cached in the CallResult object, so
	// subsequent calls to Wait will return nil.
	Wait() *Result

	// TryWait is a non-blocking variant of Wait.  It attempts to
	// retrieve the result, and returns the value and a boolean
	// value that indicates whether the result has already been
	// retrieved.
	TryWait() (*Result, bool)

	// Channel returns the channel that the CallResult object uses
	// to receive the results.  This allows the caller to directly
	// select on the channel.  Note that if the result has already
	// been received, the channel returned by this method will be
	// nil.  Using this method effectively closes the CallResult;
	// subsequent calls to Wait and TryWait will return nil
	// results.
	Channel() <-chan *Result
}

// Serializer is an interface for doing the opposite of parallelizing
// an operation.  The use case for Serializer is when something must
// be done in a single goroutine, for synchronization, but the calls
// need to be able to come from a multitude of goroutines.  In this
// case, the client code would implement the Doer interface, then wrap
// it using Serializer.
type Serializer interface {
	// Call is used to invoke the Doer.Do method of the wrapped
	// Doer.  It may return an error if the Serializer is closed.
	// Call is synchronous, and will not return until the Doer.Do
	// method has completed.
	Call(data interface{}) (*Result, error)

	// CallAsync is used to invoke the Doer.Do method, like Call,
	// but it does not block; instead, it returns a CallResult
	// object, which may be queried later for the result of the
	// call.
	CallAsync(data interface{}) (CallResult, error)

	// CallOnly is used to invoke the Doer.Do method, but it does
	// not block; instead, the result of the call is discarded.
	CallOnly(data interface{}) error

	// Wait signals the manager goroutine to exit, then waits for
	// it to do so.  The manager will call the Doer.Finish method
	// and return its result to Wait, which will in turn return it
	// to the caller.  The result will be cached to satisfy future
	// calls to Wait.
	Wait() interface{}
}
