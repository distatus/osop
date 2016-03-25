// osop
// Copyright (C) 2014-2016 Karol 'Kenji Takahashi' Woźniak
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
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

type Bspwm struct {
	connection net.Conn
	reader     *bufio.Reader
}

type bspwmDesktop struct {
	Name  string
	State string
}

type bspwmMonitor struct {
	Name     string
	State    byte
	Index    uint
	Desktops []bspwmDesktop
}

type bspwmResponse struct {
	Monitors []*bspwmMonitor
}

func (b *Bspwm) GetEvented() (interface{}, error) {
	return b.Get()
}

func (b *Bspwm) Get() (interface{}, error) {
	status, err := b.reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("Cannot read from socket: `%s`", err)
	}

	res := bspwmResponse{}
	var monitor *bspwmMonitor
	var index uint
	for _, piece := range strings.Split(status[1:len(status)], ":") {
		state, value := piece[0], piece[1:len(piece)]
		switch state {
		case 'M', 'm':
			monitor = &bspwmMonitor{Name: value, State: state, Index: index}
			res.Monitors = append(res.Monitors, monitor)
			index += 1
		case 'L':
			// TODO
		case 'O', 'o', 'F', 'f', 'U', 'u':
			monitor.Desktops = append(monitor.Desktops, bspwmDesktop{
				Name:  value,
				State: string(state),
			})
		}
	}
	return res, nil
}

func (b *Bspwm) Init(config config) error {
	socket := os.Getenv("BSPWM_SOCKET")
	if socket == "" {
		sockets, err := filepath.Glob("/tmp/bspwm*")
		if err != nil || len(sockets) < 1 {
			return fmt.Errorf("Cannot find socket file")
		}
		socket = sockets[0]
	}

	conn, err := net.Dial("unix", socket)
	if err != nil {
		return fmt.Errorf("Cannot connect to socket: `%s`", err)
	}

	if _, err = conn.Write([]byte("subscribe\x00report\x00")); err != nil {
		conn.Close()
		return fmt.Errorf("Cannot write to socket: `%s`", err)
	}

	b.connection = conn
	b.reader = bufio.NewReader(conn)
	return nil
}

func init() {
	registry.AddReceiver("bspwm", &Bspwm{}, bspwmResponse{})
}
