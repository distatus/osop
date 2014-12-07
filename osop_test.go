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
	"testing"

	"github.com/stretchr/testify/assert"
)

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
