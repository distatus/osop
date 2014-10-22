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

var RegistryAddReceiverTests = []struct {
	input    string
	expected string
}{
	{"test", "test"},
	{"Test", "test"},
	{"TEST", "test"},
}

func TestRegistry(t *testing.T) {
	inputFun := func(c config) interface{} { return "" }
	for _, tt := range RegistryAddReceiverTests {
		expected := map[string]interface{}{tt.expected: inputFun}

		registry := Registry{receivers: make(map[string]interface{})}
		registry.AddReceiver(tt.input, inputFun)

		for k, v := range expected {
			assert.Equal(t, registry.receivers[k], v)
		}
	}
}
