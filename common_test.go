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
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPanicerBase(t *testing.T) {
	fnCalled := false

	result := panicer(func(data interface{}) interface{} {
		assert.Equal(t, "data", data)
		fnCalled = true
		return "result"
	}, "data")

	assert.Equal(t, &Result{
		Result: "result",
	}, result)
	assert.True(t, fnCalled)
}

func TestPanicerPanic(t *testing.T) {
	fnCalled := false

	result := panicer(func(data interface{}) interface{} {
		assert.Equal(t, "data", data)
		fnCalled = true
		panic("this is a test")
	}, "data")

	assert.Equal(t, &Result{
		Panic: "this is a test",
	}, result)
	assert.True(t, fnCalled)
}

func TestSelectSend(t *testing.T) {
	funcCalled := false
	channel := make(chan int)

	result := selectSend(channel, 1234, func() {
		funcCalled = true
	})

	assert.Equal(t, reflect.SelectSend, result.selectCase.Dir)
	assert.Equal(t, channel, result.selectCase.Chan.Interface())
	assert.Equal(t, 1234, result.selectCase.Send.Interface())
	assert.NotNil(t, result.fn)
	result.fn(reflect.ValueOf(1234), true)
	assert.True(t, funcCalled)
}

func TestSelectRecv(t *testing.T) {
	funcCalled := false
	channel := make(chan int)

	result := selectRecv(channel, func(value interface{}, ok bool) {
		assert.Equal(t, 1234, value)
		assert.True(t, ok)
		funcCalled = true
	})

	assert.Equal(t, reflect.SelectRecv, result.selectCase.Dir)
	assert.Equal(t, channel, result.selectCase.Chan.Interface())
	assert.NotNil(t, result.fn)
	result.fn(reflect.ValueOf(1234), true)
	assert.True(t, funcCalled)
}

func TestDoSelect(t *testing.T) {
	funcCalled := false
	channel := make(chan int, 1)
	channel <- 1234
	sel := selector{
		selectCase: reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(channel),
		},
		fn: func(value reflect.Value, ok bool) {
			assert.Equal(t, 1234, value.Interface())
			assert.True(t, ok)
			funcCalled = true
		},
	}

	doSelect([]selector{sel})

	assert.True(t, funcCalled)
}
