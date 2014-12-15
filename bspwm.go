package main

import (
	"bufio"
	"fmt"
	"net"
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
		panic(err)
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
	socket, err := filepath.Glob("/tmp/bspwm*")
	if err != nil || len(socket) < 1 {
		return fmt.Errorf("Cannot find socket file")
	}
	conn, err := net.Dial("unix", socket[0])
	if err != nil {
		return fmt.Errorf("Cannot connect to socket: `%s`", err)
	}

	conn.Write([]byte("control\x00--subscribe\x00"))

	b.connection = conn
	b.reader = bufio.NewReader(conn)
	return nil
}

func init() {
	registry.AddReceiver("bspwm", &Bspwm{}, bspwmResponse{})
}
