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

import "errors"

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