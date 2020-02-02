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

import "sync"

// Size of the request channel.
const (
	requestBuffer = 100
)

// doRequest contains a request to call the Doer.Do method.  It
// contains an optional result channel, through which the results will
// be returned to the caller.  The channel must have a buffer size of
// at least one to avoid blocking the serializer manager goroutine.
type doRequest struct {
	data   interface{}    // The data to send to Doer.Do
	result chan<- *Result // Optional channel to send the result to
}

// serializer is an implementation of the Serializer interface.
type serializer struct {
	sync.Mutex
	state   pState         // State of the serializer
	doer    Doer           // The Doer wrapped
	request chan doRequest // Channel for requests
	done    chan bool      // Channel for signaling done
	gonner  *sync.Once     // A once incarnation for getting the result
	result  interface{}    // The result from finishing the operation
}

// NewSerializer constructs a serializer wrapping the specified Doer.
// All calls to Doer.Do will occur in a single manager goroutine, but
// the calls can be made from almost any other goroutine.  Note that
// Doer.Do cannot call any of the Call* methods of Serializer due to
// the potential for deadlocks.
func NewSerializer(doer Doer) Serializer {
	return &serializer{
		doer:    doer,
		request: make(chan doRequest, requestBuffer),
		done:    make(chan bool, 1),
		gonner:  &sync.Once{},
	}
}

// manager is the manager goroutine.
func (s *serializer) manager() {
	defer func() { s.done <- true }()

	for req := range s.request {
		// Run the request and send back the result
		if req.result == nil {
			panicer(s.doer.Do, req.data)
		} else {
			req.result <- panicer(s.doer.Do, req.data)
		}
	}
}

// getResult is a helper for Wait to retrieve the result of calling
// Doer.Finish.  It's called with serializer.gonner to ensure that it
// only gets called once.
func (s *serializer) getResult() {
	s.Lock()
	defer s.Unlock()

	s.result = s.doer.Finish()
	s.state = pResult
}

// Call is used to invoke the Doer.Do method of the wrapped Doer.  It
// may return an error if the Serializer is closed.  Call is
// synchronous, and will not return until the Doer.Do method has
// completed.
func (s *serializer) Call(data interface{}) (*Result, error) {
	s.Lock()

	switch s.state {
	case pNew: // Need to start the manager
		go s.manager()
		s.state = pRunning

	case pClosed, pResult: // Serializer is closed
		s.Unlock()
		return nil, ErrClosed
	}

	// OK, construct a result channel and send the request
	result := make(chan *Result, 1)
	s.request <- doRequest{
		data:   data,
		result: result,
	}
	s.Unlock()

	// Get the response and return it
	return <-result, nil
}

// CallAsync is used to invoke the Doer.Do method, like Call, but it
// does not block; instead, it returns a CallResult object, which may
// be queried later for the result of the call.
func (s *serializer) CallAsync(data interface{}) (CallResult, error) {
	s.Lock()

	switch s.state {
	case pNew: // Need to start the manager
		go s.manager()
		s.state = pRunning

	case pClosed, pResult: // Serializer is closed
		s.Unlock()
		return nil, ErrClosed
	}

	// OK, construct a result channel and send the request
	result := make(chan *Result, 1)
	s.request <- doRequest{
		data:   data,
		result: result,
	}
	s.Unlock()

	// Return a callResult
	return &callResult{response: result}, nil
}

// CallOnly is used to invoke the Doer.Do method, but it does not
// block; instead, the result of the call is discarded.
func (s *serializer) CallOnly(data interface{}) error {
	s.Lock()
	defer s.Unlock()

	switch s.state {
	case pNew: // Need to start the manager
		go s.manager()
		s.state = pRunning

	case pClosed, pResult: // Serializer is closed
		return ErrClosed
	}

	// OK, send the request
	s.request <- doRequest{data: data}

	return nil
}

// Wait signals the manager goroutine to exit, then waits for it to do
// so.  The manager will call the Doer.Finish method and return its
// result to Wait, which will in turn return it to the caller.  The
// result will be cached to satisfy future calls to Wait.
func (s *serializer) Wait() interface{} {
	s.Lock()

	switch s.state {
	case pNew: // Haven't even started yet
		s.state = pClosed
		s.Unlock()
		s.gonner.Do(s.getResult)

	case pRunning: // Signal to die, wait for done
		s.state = pClosed
		s.Unlock()
		close(s.request)
		<-s.done
		s.gonner.Do(s.getResult)

	case pClosed: // Closed, waiting for result
		s.Unlock()
		s.gonner.Do(s.getResult)

	case pResult: // Have result, just need to unlock
		s.Unlock()
	}

	return s.result
}

// callResult is an implementation of the CallResult interface.
type callResult struct {
	response <-chan *Result // The channel we'll get the response on
	closed   bool           // A boolean flag indicating if the CallResult is closed
}

// Wait is used to retrieve the result of the call.  The result is not
// cached in the CallResult object, so subsequent calls to Wait will
// return nil.
func (c *callResult) Wait() *Result {
	// Are we closed?
	if c.closed {
		return nil
	}

	// OK, wait for the response and close the call result
	c.closed = true
	return <-c.response
}

// TryWait is a non-blocking variant of Wait.  It attempts to retrieve
// the result, and returns the value and a boolean value that
// indicates whether the result has already been retrieved.
func (c *callResult) TryWait() (*Result, bool) {
	// Are we closed?
	if c.closed {
		return nil, false
	}

	// OK, use select with a default to try to get the result
	select {
	case response := <-c.response: // Got a response
		c.closed = true
		return response, true

	default: // Would block
		return nil, true
	}
}

// Channel returns the channel that the CallResult object uses to
// receive the results.  This allows the caller to directly select on
// the channel.  Note that if the result has already been received,
// the channel returned by this method will be nil.  Using this method
// effectively closes the CallResult; subsequent calls to Wait and
// TryWait will return nil results.
func (c *callResult) Channel() <-chan *Result {
	// Are we closed?
	if c.closed {
		return nil
	}

	// OK, close ourself and return the channel
	c.closed = true
	return c.response
}
