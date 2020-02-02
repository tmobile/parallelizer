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

import "container/list"

// synchronousWorker is an implementation of the Worker interface that
// operates in a synchronous fashion; that is, there are no goroutines
// involved.  It might be used when adapting an existing synchronous
// algorithm to a parallel algorithm, or it might be used when the
// parallelization is intended to be optional, such as when ordering
// may be important.
type synchronousWorker struct {
	state   workerState // State of the worker
	runner  Runner      // The runner to be invoked by the workers
	queue   *list.List  // A queue of submitted work items
	running bool        // A flag indicating that Call is running
	result  interface{} // The result that came from calling Runner.Result
}

// NewSynchronousWorker constructs a synchronous worker.  Synchronous
// workers do not utilize parallelism at all; they are provided to
// allow for transition from a single-threaded algorithm to a
// multithreaded one, or to enable optional parallelization in cases
// where ordering may be important for certain invocations.
func NewSynchronousWorker(runner Runner) Worker {
	return &synchronousWorker{
		runner: runner,
		queue:  &list.List{},
	}
}

// run is a helper that runs the items on the queue.
func (w *synchronousWorker) run() {
	for w.queue.Len() > 0 {
		// Get an element off the queue
		elem := w.queue.Front()
		w.queue.Remove(elem)

		// Run the runner with that data
		result := panicer(w.runner, elem.Value)

		// Integrate the results
		w.runner.Integrate(w, result.result, result.panicData)
	}
}

// Call is the method used to submit data to be worked in a call to
// the Runner.Run method.  It may return an error if the worker has
// been shut down through a call to Wait.
func (w *synchronousWorker) Call(data interface{}) error {
	// Check the worker state
	switch w.state {
	case workerNew:
		w.state = workerRunning

	case workerClosed, workerResult:
		// Accept recursive Calls even when closed, so we work
		// all items
		if !w.running {
			return ErrWorkerClosed
		}
	}

	// Enqueue the data
	w.queue.PushBack(data)

	// If we're running, avoid recursion and allow the outside
	// Call to do it all
	if w.running {
		return nil
	}
	w.running = true

	// Run the queue
	w.run()

	// Done running
	w.running = false

	return nil
}

// Wait is called to shut down the worker and return the final result;
// it will block the caller until all data has been processed and all
// worker goroutines have stopped.  Note that the final result,
// generated by Runner.Result, is saved by Worker to satisfy later
// calls to Wait.  If Wait is called before any calls to Call, the
// worker will go straight to a stopped state, and no further Call
// calls may be made; no error will be returned in that case.
func (w *synchronousWorker) Wait() (interface{}, error) {
	// Detect deadlocks
	if w.running {
		return nil, ErrWouldDeadlock
	}

	// Check the worker state
	switch w.state {
	case workerNew, workerRunning, workerClosed: // Get the result
		w.result = w.runner.Result()
		w.state = workerResult
	}

	return w.result, nil
}
