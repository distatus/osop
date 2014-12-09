// osop
// Copyright (C) 2014 Karol 'Kenji Takahashi' Woźniak
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
// OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
// DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
// TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
// OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package main

import (
	"bytes"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var logR bytes.Buffer

var RegistryTests = []struct {
	input    string
	expected string
	zero     string
}{
	{"test", "test", "z1"},
	{"Test", "test", "z2"},
	{"TEST", "test", "z3"},
}

func TestRegistry(t *testing.T) {
	for _, tt := range RegistryTests {
		inputFun := func(c config) (interface{}, error) { return tt.input, nil }

		registry := Registry{
			receivers: make(map[string]interface{}),
			zeros:     make(map[string]interface{}),
		}
		registry.AddReceiver(tt.input, inputFun, tt.zero)

		assert.Equal(t, 1, len(registry.receivers))
		assert.Equal(t, 1, len(registry.zeros))
		result, rerr := registry.receivers[tt.expected].(receiverCtor)(map[string]interface{}{})
		assert.Equal(t, tt.input, result)
		assert.Nil(t, rerr)
		assert.Equal(t, tt.zero, registry.zeros[tt.expected])

		for _, ttt := range RegistryTests {
			receiver, err := registry.GetReceiver(ttt.expected)
			assert.Nil(t, err)
			result, rerr := receiver(map[string]interface{}{})
			assert.Equal(t, tt.input, result)
			assert.Nil(t, rerr)
			zero, err := registry.GetZero(ttt.expected)
			assert.Equal(t, tt.zero, zero)
			assert.Nil(t, err)
		}

		receiver, err := registry.GetReceiver("tset")
		assert.Nil(t, receiver)
		assert.Equal(t, "Receiver `tset` not found", err.Error())
		zero, err := registry.GetZero("tset")
		assert.Nil(t, zero)
		assert.Equal(t, "Receiver `tset` zero value not found", err.Error())
	}
}

type testReceiver interface {
	Swap()
}

type testReceiverPolling struct {
	Good  bool
	Count uint
}

func (r *testReceiverPolling) Get() (interface{}, error) {
	r.Count += 1
	if r.Good {
		return fmt.Sprintf("pollingTest%d", r.Count), nil
	}
	return nil, fmt.Errorf("pollingError%d", r.Count)
}

func (r *testReceiverPolling) Swap() {
	r.Good = !r.Good
}

type testReceiverEvented struct {
	testReceiverPolling
}

func (r *testReceiverEvented) GetEvented() (interface{}, error) {
	r.Count += 1
	if r.Good {
		return fmt.Sprintf("eventedTest%d", r.Count), nil
	}
	return nil, fmt.Errorf("eventedError%d", r.Count)
}

var WorkerTests = []struct {
	receiver testReceiver
	expected []string
}{
	{&testReceiverPolling{Good: true}, []string{"polling", "polling"}},
	{&testReceiverEvented{testReceiverPolling{Good: true}}, []string{"polling", "evented"}},
}

func TestWorker(t *testing.T) {
	for _, tt := range WorkerTests {
		receiver := tt.receiver
		worker := Worker{
			pollInterval: time.Millisecond,
			receiver:     receiver,
			name:         "testGood",
			once:         true,
		}
		ch := make(chan Change)

		go worker.Do(ch)
		// Yes, we're doing it twice
		assert.Equal(t, Change{Name: "testGood", Value: fmt.Sprintf("%sTest1", tt.expected[0])}, <-ch)
		assert.Equal(t, Change{Name: "testGood", Value: fmt.Sprintf("%sTest2", tt.expected[1])}, <-ch)

		receiver.Swap()
		worker.name = "testBad"

		// There should be no channel usage,
		// goroutine not used on purpose.
		worker.Do(ch)
		for i, e := range tt.expected {
			stderr, err := logR.ReadString('\n')
			assert.Nil(t, err)
			assert.Equal(t, fmt.Sprintf("testBad: %sError%d\n", e, i+3), stderr[20:len(stderr)])
			assert.Equal(t, 0, len(ch))
		}
	}
}

type testRegistry struct {
	Good bool
}

func (t *testRegistry) AddReceiver(name string, receiver receiverCtor, zero interface{}) {}

func (t *testRegistry) GetReceiver(name string) (receiverCtor, error) {
	return func(config config) (interface{}, error) {
		if !t.Good {
			t.Good = true
			return nil, fmt.Errorf("CtorError")
		}
		return true, nil
	}, nil
}

func (t *testRegistry) GetZero(name string) (interface{}, error) {
	return nil, nil
}

var NewWorkerTests = []struct {
	Good   bool
	Config config
	Assert func(t *testing.T, worker *Worker)
}{
	{true, map[string]interface{}{"receiver": "test"}, func(t *testing.T, worker *Worker) {
		assert.Equal(t, time.Second, worker.pollInterval)
		assert.Equal(t, true, worker.receiver)
		assert.Equal(t, "test", worker.name)

		_, err := logR.ReadString('\n')
		assert.NotNil(t, err)
		assert.Equal(t, "EOF", err.Error())
	}},
	{false, map[string]interface{}{"receiver": "test"}, func(t *testing.T, worker *Worker) {
		assert.Equal(t, time.Second, worker.pollInterval)
		assert.Equal(t, true, worker.receiver)
		assert.Equal(t, "test", worker.name)

		stderr, err := logR.ReadString('\n')
		assert.Nil(t, err)
		assert.Equal(t, "CtorError\n", stderr[20:len(stderr)])
	}},
	{true, map[string]interface{}{"receiver": "test", "pollInterval": "1m"}, func(t *testing.T, worker *Worker) {
		assert.Equal(t, time.Minute, worker.pollInterval)
	}},
}

func TestNewWorker(t *testing.T) {
	for _, tt := range NewWorkerTests {
		registry = &testRegistry{Good: tt.Good}

		worker := NewWorker("test", tt.Config)
		tt.Assert(t, worker)
	}
}

func init() {
	log.SetOutput(&logR)
}
