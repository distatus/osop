// osop
// Copyright (C) 2015 Karol 'Kenji Takahashi' Woźniak
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

import "hkjn.me/power"

type batteryResponse struct {
	Charge  float32
	Percent float32
	State   string
}

type Battery struct {
	number int
}

func (b *Battery) Get() (interface{}, error) {
	res, err := power.GetNumber(b.number)
	if err != nil {
		return nil, err
	}

	percent := float32(res.Charge * 100)
	// If battery controller does not work as expected.
	if percent > 100 {
		percent = 100
	}
	return batteryResponse{
		Charge:  float32(res.Charge),
		Percent: percent,
		State:   res.State.String(),
	}, nil
}

func (b *Battery) Init(config config) error {
	if config["number"] != nil {
		b.number = int(config["number"].(int64))
	}
	return nil
}

func init() {
	registry.AddReceiver("Battery", &Battery{}, batteryResponse{})
}
