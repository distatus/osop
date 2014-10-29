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
	"strings"
	"time"

	"github.com/pyk/byten"
	"github.com/shirou/gopsutil"
)

func bytonizeUint(i uint64, speed, short bool) string {
	b := byten.Size(int64(i))
	if short {
		p := b[len(b)-2]
		if p < '0' || p > '9' {
			b = b[:len(b)-1]
		}
	}
	if speed {
		b += "/s"
	}
	return b
}

type Sys struct {
	metrics []string
	shorts  bool

	downloaded map[string]uint64
	uploaded   map[string]uint64
	interval   float64
}

type sysResponseNetwork struct {
	Sent     string
	Recv     string
	Download string
	Upload   string
}

type sysResponse struct {
	CPU struct {
		Percent map[string]string
	}
	Uptime uint64
	Memory struct {
		Total string
		UsedF string
		UsedA string
	}
	Swap struct {
		Total string
		Used  string
	}
	Network map[string]sysResponseNetwork
}

func (s *Sys) Get() (interface{}, error) {
	resp := sysResponse{}
	var err error
	for _, metric := range s.metrics {
		split := strings.Split(strings.ToLower(metric), " ")
		switch split[0] {
		case "cpu":
			if len(split) < 2 {
				err = fmt.Errorf("Sys: `cpu` requires argument")
			}
			switch split[1] {
			case "percent":
				var cpupercents []float32
				if len(split) < 3 || split[2] == "false" {
					cpupercents, err = gopsutil.CPUPercent(0, false)
				} else if split[2] == "true" {
					cpupercents, err = gopsutil.CPUPercent(0, true)
				} else {
					err = fmt.Errorf("Sys: `cpu percent` got wrong argument")
					break
				}
				resp.CPU.Percent = make(map[string]string)
				for i, cpupercent := range cpupercents {
					resp.CPU.Percent[fmt.Sprintf("cpu%d", i)] = fmt.Sprintf("%.2f%%", cpupercent)
				}
			}
		case "uptime":
			resp.Uptime, err = gopsutil.BootTime()
		case "memory":
			var mem *gopsutil.VirtualMemoryStat
			mem, err = gopsutil.VirtualMemory()
			resp.Memory.Total = bytonizeUint(mem.Total, false, s.shorts)
			resp.Memory.UsedF = bytonizeUint(mem.Used, false, s.shorts)
			resp.Memory.UsedA = bytonizeUint(mem.Total-mem.Available, false, s.shorts)
		case "swap":
			var mem *gopsutil.SwapMemoryStat
			mem, err = gopsutil.SwapMemory()
			resp.Swap.Total = bytonizeUint(mem.Total, false, s.shorts)
			resp.Swap.Used = bytonizeUint(mem.Used, false, s.shorts)
		case "network":
			var nic []gopsutil.NetIOCountersStat
			if len(split) < 2 || strings.ToLower(split[1]) == "all" {
				// FIXME: Returns eth0 only, seems gopsutil bug
				//nic, err = gopsutil.NetIOCounters(false)
				//if err != nil || len(nic) == 0 {
				//break
				//}
				//resp.Network = map[string]gopsutil.NetIOCountersStat{"All": nic[0]}
			} else {
				nic, err = gopsutil.NetIOCounters(true)
				if err != nil || len(nic) == 0 {
					break
				}
				resp.Network = make(map[string]sysResponseNetwork)
				for _, iface := range split[1:] {
					resp.Network[iface] = s.getNetworkByName(nic, iface)
				}
			}
		}
		if err != nil {
			log.Printf("Sys: Cannot get `%s`: `%s`\n", metric, err)
		}
	}

	return resp, nil
}

func (s *Sys) getNetworkByName(
	nices []gopsutil.NetIOCountersStat,
	name string,
) sysResponseNetwork {
	net := sysResponseNetwork{}
	for _, nic := range nices {
		if nic.Name == name {
			net.Sent = bytonizeUint(nic.BytesSent, false, s.shorts)
			net.Recv = bytonizeUint(nic.BytesRecv, false, s.shorts)
			net.Download = bytonizeUint(
				uint64((float64(nic.BytesRecv)-float64(s.downloaded[name]))/s.interval),
				true, s.shorts,
			)
			s.downloaded[name] = nic.BytesRecv
			net.Upload = bytonizeUint(
				uint64((float64(nic.BytesSent)-float64(s.uploaded[name]))/s.interval),
				true, s.shorts,
			)
			s.uploaded[name] = nic.BytesSent
		}
	}
	return net
}

func NewSys(config config) (interface{}, error) {
	if config["metrics"] == nil {
		return nil, fmt.Errorf("")
	}
	metrics := config["metrics"].([]interface{})
	s := &Sys{
		metrics:    make([]string, len(metrics)),
		downloaded: make(map[string]uint64),
		uploaded:   make(map[string]uint64),
	}
	for i, metric := range metrics {
		s.metrics[i] = metric.(string)
	}

	interval, _ := time.ParseDuration(config["pollInterval"].(string))
	s.interval = interval.Seconds()

	return s, nil
}

func init() {
	registry.AddReceiver("Sys", NewSys)
}
