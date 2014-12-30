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

// osop - other side of the pipe - outputs formatted metrics to your Stdout.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"reflect"
	"strings"
	"text/template"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/adrg/xdg"
)

// fatal is a helper function to call when something terribly wrong
// may have happened. Logs given error and terminates application.
func fatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// config defines configuration structure.
type config map[string]interface{}

// PollingReceiver defines a basic type of receiver, which
// will run every config:`pollInterval` and try to get new data ASAP.
type PollingReceiver interface {
	Init(config config) error
	Get() (interface{}, error)
}

// EventedReceiver defines an advanced receiver, which
// is able to wait for a change to happen and only then report
// a new value.
//
// Note that it does not need to implement a fully functional Get() method.
// It is only used at the beginning, so user do not have to wait
// for an event to happen to get the initial value.
type EventedReceiver interface {
	PollingReceiver
	GetEvented() (interface{}, error)
}

// IRegistry defines interface for receivers registry.
//
// Default registry is provided as a globally accessible `registry`
// variable. All receivers shall add themselves there before they
// could be used (tip: init() function is a good way to do so).
type IRegistry interface {
	AddReceiver(string, PollingReceiver, interface{})
	GetReceiver(string) (PollingReceiver, error)
	GetZero(string) (interface{}, error)
}

// Registry is a default IRegistry implementation.
type Registry struct {
	receivers map[string]reflect.Type
	zeros     map[string]interface{}
}

// AddReceiver adds new receiver to registry.
//
// `zero` should be an initial (probably empty), expected response.
// It's to workaround text/template panicking on non-existing structure elements.
func (r *Registry) AddReceiver(name string, rec PollingReceiver, zero interface{}) {
	name = strings.ToLower(name)
	r.receivers[name] = reflect.TypeOf(rec).Elem()
	r.zeros[name] = zero
}

// GetReceiver gets existing receiver from registry.
// New instance is created on every call to allow multiple
// instances of the same receiver to co-exist.
//
// Note that receiver names are case insensitive.
func (r *Registry) GetReceiver(name string) (PollingReceiver, error) {
	v := r.receivers[strings.ToLower(name)]
	if v == nil {
		return nil, fmt.Errorf("Receiver `%s` not found", name)
	}
	return reflect.New(v).Interface().(PollingReceiver), nil
}

// GetZero gets zero response for an existing receiver.
//
// Note that receiver names are case insensitive.
func (r *Registry) GetZero(name string) (interface{}, error) {
	v, ok := r.zeros[strings.ToLower(name)]
	if !ok {
		return nil, fmt.Errorf("Receiver `%s` zero value not found", name)
	}
	return v, nil
}

// registry is a default, globally available Registry instance.
var registry IRegistry = &Registry{
	receivers: make(map[string]reflect.Type),
	zeros:     make(map[string]interface{}),
}

// Change is emitted for every receiver value change.
type Change struct {
	Name  string
	Value interface{}
}

// Worker processes receiver value changes.
//
// Responsible for getting the value from receiver and propagating it
// further to the template compilation method.
type Worker struct {
	pollInterval time.Duration
	receiver     PollingReceiver
	name         string
	once         bool
}

// doChange handles a single value change.
func (w *Worker) doChange(get func() (interface{}, error), ch chan Change) {
	value, err := get()
	if err != nil {
		log.Printf("%s: %s\n", w.name, err)
		return
	}
	if value != nil {
		ch <- Change{
			Name:  w.name,
			Value: value,
		}
	}
}

// Do acts as a Worker event loop.
//
// For PollingReceivers, spawns every config:`pollInterval`.
// For EventedReceivers, blocks until an event is generated.
func (w *Worker) Do(ch chan Change) {
	switch r := w.receiver.(type) {
	case EventedReceiver:
		// Get first value in "normal" manner,
		// so user won't have to wait for an event to occur.
		w.doChange(r.Get, ch)
		for {
			w.doChange(r.GetEvented, ch)
			if w.once {
				break
			}
		}
	case PollingReceiver:
		w.doChange(r.Get, ch)
		for _ = range time.Tick(w.pollInterval) {
			w.doChange(r.Get, ch)
			if w.once {
				break
			}
		}
	}
}

// NewWorker constructs new Worker instance with given name and config.
func NewWorker(name string, config config) *Worker {
	interval := time.Second
	if config["pollInterval"] != nil {
		_interval, err := time.ParseDuration(config["pollInterval"].(string))
		if err == nil {
			interval = _interval
		}
	}
	receiver, _ := registry.GetReceiver(config["receiver"].(string))

	err := receiver.Init(config)
	for err != nil {
		log.Println(err)
		time.Sleep(time.Second)
		err = receiver.Init(config)
	}

	return &Worker{
		pollInterval: interval,
		receiver:     receiver,
		name:         name,
	}
}

func main() {
	configFilename := flag.String("c", "", "Path to the configuration file")
	flag.Parse()

	var configs map[string]map[string]interface{}
	if _, err := toml.DecodeFile(*configFilename, &configs); err != nil {
		if _, ok := err.(*os.PathError); !ok {
			fatal(err)
		}
		if *configFilename == "" {
			*configFilename = "config.toml"
		}
		xdgFile, err := xdg.ConfigFile(path.Join("osop", *configFilename))
		fatal(err)
		if _, err := os.Stat(xdgFile); os.IsNotExist(err) {
			f, err := os.Create(xdgFile)
			fatal(err)
			f.WriteString(strings.TrimSpace(`
[Now]
receiver="date"
pollInterval="1s"
format="02/01/2006 15:04:05"

[Osop]
template="<.Now>"
			`))
			f.Close()
		}
		_, err = toml.DecodeFile(xdgFile, &configs)
		fatal(err)
	}

	delims, ok := configs["Osop"]["delims"].([]interface{})
	if !ok {
		delims = []interface{}{"<", ">"}
	}
	t, err := template.New("t").Delims(
		delims[0].(string), delims[1].(string),
	).Funcs(template.FuncMap{"stringify": func(arg interface{}) string {
		s, ok := arg.(string)
		if !ok {
			return ""
		}
		return s
	}}).Parse(
		configs["Osop"]["template"].(string) + "\n",
	)
	fatal(err)

	workers := make(chan *Worker)

	data := make(map[string]interface{})

	for name, conf := range configs {
		if name == "Osop" {
			continue
		}
		zero, err := registry.GetZero(conf["receiver"].(string))
		if err != nil {
			log.Printf("Error getting receiver (`%s`), not spawning worker\n", err)
			continue
		}
		data[name] = zero
		go func(ch chan *Worker, name string, conf config) {
			ch <- NewWorker(name, conf)
		}(workers, name, conf)
	}

	changes := make(chan Change)
	var cache string
	for {
		select {
		case worker := <-workers:
			if worker != nil {
				go worker.Do(changes)
			}
		case change := <-changes:
			data[change.Name] = change.Value
			var buf bytes.Buffer
			err := t.Execute(&buf, data)
			if err != nil {
				buf.WriteByte('\n')
			}

			str := buf.String()
			if str == cache {
				continue
			}
			cache = str

			fmt.Print(cache)
		}
	}
}
