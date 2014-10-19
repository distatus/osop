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
	"text/template"
	"os"
	"time"
	"log"

	"github.com/BurntSushi/toml"
)

func fatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}


type config map[string]interface{}

type PollingReceiver interface {
	Get() interface{}
}

type EventedReceiver interface {
	GetEvented() interface{}
}

type receiverCtor func(config) interface{}

type Registry struct {
	receivers map[string]interface{}
}

func (r *Registry) AddReceiver(name string, fun receiverCtor) {
	r.receivers[name] = fun
}

func (r *Registry) GetReceiver(name string) (receiverCtor, error) {
	v := r.receivers[name]
	if v == nil {
		return nil, fmt.Errorf("Receiver `%s` not found", name)
	}
	return v.(receiverCtor), nil
}

var registry = Registry{
	receivers: make(map[string]interface{}),
}


type Change struct {
	Name string
	Value interface{}
}

type Worker struct {
	pollInterval time.Duration
	updateOnChange bool
	receiver interface{}
	name string
}

func (w *Worker) Do(ch chan Change) {
	switch r := w.receiver.(type) {
	case EventedReceiver:
		for {
			ch <- Change{
				Name: w.name,
				Value: r.GetEvented(),
			}
		}
	case PollingReceiver:
		for _ = range time.Tick(w.pollInterval) {
			ch <- Change{
				Name: w.name,
				Value: r.Get(),
			}
		}
	}
}

func NewWorker(name string, config config) *Worker {
	// TODO: Should be optional for EventedReceivers
	interval, err := time.ParseDuration(config["pollInterval"].(string))
	if err != nil {
		log.Printf("Error parsing pollInterval (`%s`), default to 1s", err)
		interval = time.Second
	}
	receiver, err := registry.GetReceiver(config["receiver"].(string))
	if err != nil {
		fmt.Printf("Error getting receiver (`%s`), not spawning worker", err)
		return nil
	}

	return &Worker{
		pollInterval: interval,
		receiver: receiver(config),
		name: name,
	}
}

func main() {
	var configs map[string]map[string]interface{}
	_, err := toml.DecodeFile("/home/kenji/osop/config.toml", &configs)
	fatal(err)

	delims := configs["Osop"]["delims"].([]interface{})
	t, err := template.New("t").Delims(delims[0].(string), delims[1].(string)).Parse(
		configs["Osop"]["template"].(string) + "\n",
	)
	fatal(err)

	ch := make(chan Change)

	data := make(map[string]interface{})
	for receiver, config := range configs {
		if receiver == "Osop" {
			continue
		}
		worker := NewWorker(receiver, config)
		if worker != nil {
			go worker.Do(ch)
		}
	}

	for change := range ch {
		data[change.Name] = change.Value
		t.Execute(os.Stdout, data)
	}
}
