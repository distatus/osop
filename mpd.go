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
	"fmt"
	"log"

	"github.com/fhs/gompd/mpd"
)

type Mpd struct {
	address  string
	password string
}

type mpdResponse struct {
	Song   map[string]string
	Status map[string]string
}

func (m *Mpd) Get() (interface{}, error) {
	// FIXME: This should go to constructor.
	client, err := mpd.DialAuthenticated("tcp", m.address, m.password)
	if err != nil {
		return nil, fmt.Errorf("Connection error: `%s`", err)
	}

	return mpdResponse{
		Song:   m.getCurrentSong(client),
		Status: m.getStatus(client),
	}, nil
}

func (m *Mpd) getCurrentSong(client *mpd.Client) map[string]string {
	current, err := client.CurrentSong()
	if err != nil {
		log.Printf("Error getting Mpd current song: `%s`", err)
		return nil
	}
	return current
}

func (m *Mpd) getStatus(client *mpd.Client) map[string]string {
	status, err := client.Status()
	if err != nil {
		log.Printf("Error getting Mpd status: `%s`", err)
		return nil
	}
	return status
}

func (m *Mpd) Init(config config) error {
	address := config["address"].(string)
	password := config["password"].(string)

	m.address = address
	m.password = password
	return nil
}

func init() {
	registry.AddReceiver("Mpd", &Mpd{}, mpdResponse{})
}
