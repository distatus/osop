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
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

const URL = "http://api.openweathermap.org/data/2.5/weather"

type Owm struct {
	url string
}

type owmResponse struct {
	City     string
	Country  string
	Sunrise  uint64
	Sunset   uint64
	Temp     float64
	TempMin  float64
	TempMax  float64
	Pressure int
	Humidity int

	Wind struct {
		Speed float64
		Deg   int
	}

	Coord struct {
		Lon float64
		Lat float64
	}
}

func (o *Owm) Get() (interface{}, error) {
	resp, err := http.Get(o.url)
	if err != nil {
		return nil, fmt.Errorf("Cannot get response: `%s`", err)
	}
	defer resp.Body.Close()

	var decoded struct {
		Coord struct {
			Lon float64
			Lat float64
		}
		Sys struct {
			Country string
			Sunrise uint64
			Sunset  uint64
		}
		Main struct {
			Temp     float64
			Pressure int
			Humidity int
			Temp_min float64
			Temp_max float64
		}
		Wind struct {
			Speed float64
			Deg   int
		}
		Name string
	}
	json.NewDecoder(resp.Body).Decode(&decoded)

	return owmResponse{
		City:     decoded.Name,
		Country:  decoded.Sys.Country,
		Sunrise:  decoded.Sys.Sunrise,
		Sunset:   decoded.Sys.Sunset,
		Temp:     decoded.Main.Temp,
		TempMin:  decoded.Main.Temp_min,
		TempMax:  decoded.Main.Temp_max,
		Pressure: decoded.Main.Pressure,
		Humidity: decoded.Main.Humidity,
		Wind:     decoded.Wind,
		Coord:    decoded.Coord,
	}, nil
}

func NewOwm(config config) (interface{}, error) {
	if config["location"] == nil {
		return nil, fmt.Errorf("Location parameter is required for Owm receiver")
	}
	_url, err := url.Parse(URL)
	if err != nil {
		return nil, fmt.Errorf("Cannot parse Owm URL: `%s`", err)
	}
	location := config["location"].(string)
	_, err = strconv.Atoi(location)
	urlQuery := url.Values{}
	if err != nil {
		urlQuery.Add("q", location)
	} else {
		urlQuery.Add("id", location)
	}

	if config["apiKey"] != nil {
		urlQuery.Add("APPID", config["apiKey"].(string))
	}

	units := "metric"
	if config["units"] != nil {
		_units := config["units"].(string)
		if _units != "metric" && _units != "imperial" {
			log.Printf("Unknown units (%s), using `metric`\n", _units)
		} else {
			units = _units
		}
	}
	urlQuery.Add("units", units)

	_url.RawQuery = urlQuery.Encode()

	return &Owm{
		url: _url.String(),
	}, nil
}

func init() {
	registry.AddReceiver("Owm", NewOwm)
}
