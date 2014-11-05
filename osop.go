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
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/BurntSushi/toml"
)

func fatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type config map[string]interface{}

type PollingReceiver interface {
	Get() (interface{}, error)
}

type EventedReceiver interface {
	PollingReceiver
	GetEvented() (interface{}, error)
}

type receiverCtor func(config) (interface{}, error)

type Registry struct {
	receivers map[string]interface{}
}

func (r *Registry) AddReceiver(name string, fun receiverCtor) {
	r.receivers[strings.ToLower(name)] = fun
}

func (r *Registry) GetReceiver(name string) (receiverCtor, error) {
	v := r.receivers[strings.ToLower(name)]
	if v == nil {
		return nil, fmt.Errorf("Receiver `%s` not found", name)
	}
	return v.(receiverCtor), nil
}

var registry = Registry{
	receivers: make(map[string]interface{}),
}

type Change struct {
	Name  string
	Value interface{}
}

type Worker struct {
	pollInterval   time.Duration
	updateOnChange bool
	receiver       interface{}
	name           string
}

func (w *Worker) Do(ch chan Change) {
	doChange := func(r PollingReceiver, ch chan Change) {
		value, err := r.Get()
		if err != nil {
			log.Printf("%s: %s\n", w.name, err)
		}
		if value != nil {
			ch <- Change {
				Name:  w.name,
				Value: value,
			}
		}
	}

	switch r := w.receiver.(type) {
	case EventedReceiver:
		// Get first value in "normal" manner,
		// so user won't have to wait for an event to occur.
		doChange(r, ch)
		for {
			value, err := r.GetEvented()
			if err != nil {
				log.Printf("%s: %s\n", w.name, err)
				continue
			}
			if value != nil {
				ch <- Change{
					Name:  w.name,
					Value: value,
				}
			}
		}
	case PollingReceiver:
		doChange(r, ch)
		for _ = range time.Tick(w.pollInterval) {
			doChange(r, ch)
		}
	}
}

func NewWorker(name string, config config) *Worker {
	interval := time.Second
	if config["pollInterval"] != nil {
		_interval, err := time.ParseDuration(config["pollInterval"].(string))
		if err == nil {
			interval = _interval
		}
	}
	receiver, err := registry.GetReceiver(config["receiver"].(string))
	if err != nil {
		log.Printf("Error getting receiver (`%s`), not spawning worker\n", err)
		return nil
	}

	receiverInstance, err := receiver(config)
	for err != nil {
		log.Println(err)
		time.Sleep(time.Second)
		receiverInstance, err = receiver(config)
	}

	return &Worker{
		pollInterval: interval,
		receiver:     receiverInstance,
		name:         name,
	}
}

func main() {
	configFilename := flag.String("c", "", "Path to the configuration file")
	flag.Parse()

	var configs map[string]map[string]interface{}
	_, err := toml.DecodeFile(*configFilename, &configs)
	fatal(err)

	delims := configs["Osop"]["delims"].([]interface{})
	t, err := template.New("t").Delims(delims[0].(string), delims[1].(string)).Parse(
		configs["Osop"]["template"].(string) + "\n",
	)
	fatal(err)

	workers := make(chan *Worker)
	var wg sync.WaitGroup

	for receiver, conf := range configs {
		if receiver == "Osop" {
			continue
		}
		wg.Add(1)
		go func(ch chan *Worker, receiver string, conf config) {
			defer wg.Done()
			ch <- NewWorker(receiver, conf)
		}(workers, receiver, conf)
	}

	changes := make(chan Change)
	data := make(map[string]interface{})
	for {
		select {
		case worker := <-workers:
			if worker != nil {
				go worker.Do(changes)
			}
		case change := <-changes:
			data[change.Name] = change.Value
			err := t.Execute(os.Stdout, data)
			if err != nil {
				fmt.Println()
			}
		}
	}
}
