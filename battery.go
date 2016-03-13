// osop
// Copyright (C) 2015-2016 Karol 'Kenji Takahashi' Woźniak
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

import "github.com/distatus/battery"

type batteryResponse struct {
	Charge  float32
	Percent float32
	State   string
}

type Battery struct {
	number int
}

func (b *Battery) Get() (interface{}, error) {
	res, err := battery.Get(b.number)
	if err != nil {
		return nil, err
	}

	charge := res.Current / res.Full
	percent := float32(charge * 100)
	// If battery controller does not work as expected.
	if percent > 100 {
		percent = 100
	}
	return batteryResponse{
		Charge:  float32(charge),
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
