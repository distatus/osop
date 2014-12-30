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
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/pyk/byten"
)

func bytonizeByte(data []byte, speed bool, short bool) (string, error) {
	i, err := strconv.ParseInt(string(data), 0, 0)
	if err != nil {
		return "", err
	}
	b := byten.Size(i)
	if short {
		p := b[len(b)-2]
		if p < '0' || p > '9' {
			b = b[:len(b)-1]
		}
	}
	if speed {
		b += "/s"
	}
	return b, nil
}

type stringByten string

func (s *stringByten) UnmarshalJSON(data []byte) error {
	b, err := bytonizeByte(data, false, false)
	*s = stringByten(b)
	return err
}

type stringBytenShort string

func (s *stringBytenShort) UnmarshalJSON(data []byte) error {
	b, err := bytonizeByte(data, false, true)
	*s = stringBytenShort(b)
	return err
}

type stringBytenSpeed string

func (s *stringBytenSpeed) UnmarshalJSON(data []byte) error {
	b, err := bytonizeByte(data, true, false)
	*s = stringBytenSpeed(b)
	return err
}

type stringBytenSpeedShort string

func (s *stringBytenSpeedShort) UnmarshalJSON(data []byte) error {
	b, err := bytonizeByte(data, true, true)
	*s = stringBytenSpeedShort(b)
	return err
}

type Transmission struct {
	url       string
	sessionId string
	shorts    bool
	client    *http.Client
}

type transmissionResponseStats struct {
	Uploaded      stringByten `json:"uploadedBytes"`
	Downloaded    stringByten `json:"downloadedBytes"`
	FilesAdded    uint64
	SessionCount  uint64
	SecondsActive uint64
}

type transmissionResponseStatsShort struct {
	Uploaded      stringBytenShort `json:"uploadedBytes"`
	Downloaded    stringBytenShort `json:"downloadedBytes"`
	FilesAdded    uint64
	SessionCount  uint64
	SecondsActive uint64
}

type transmissionResponse struct {
	TorrentCount       uint64
	ActiveTorrentCount uint64
	PausedTorrentCount uint64
	DownloadSpeed      stringBytenSpeed
	UploadSpeed        stringBytenSpeed
	Cumulative         transmissionResponseStats `json:"cumulative-stats"`
	Current            transmissionResponseStats `json:"current-stats"`
}

type transmissionResponseShort struct {
	TorrentCount       uint64
	ActiveTorrentCount uint64
	PausedTorrentCount uint64
	DownloadSpeed      stringBytenSpeedShort
	UploadSpeed        stringBytenSpeedShort
	Cumulative         transmissionResponseStatsShort `json:"cumulative-stats"`
	Current            transmissionResponseStatsShort `json:"current-stats"`
}

func (t *Transmission) Get() (interface{}, error) {
	req, err := http.NewRequest(
		"POST", t.url, bytes.NewBufferString(`{"method":"session-stats"}`),
	)
	if err != nil {
		return nil, fmt.Errorf("Cannot create request: `%s`", err)
	}
	req.Header.Add("X-Transmission-Session-Id", t.sessionId)
	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Cannot send request: `%s`", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 409 {
		t.sessionId = resp.Header.Get("X-Transmission-Session-Id")
		return t.Get()
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Wrong status code: `%d`", resp.StatusCode)
	}

	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()
	if t.shorts {
		var data struct {
			Arguments transmissionResponseShort
		}
		decoder.Decode(&data)
		return data.Arguments, nil
	} else {
		var data struct {
			Arguments transmissionResponse
		}
		decoder.Decode(&data)
		return data.Arguments, nil
	}
}

func (t *Transmission) Init(config config) error {
	if config["address"] == nil {
		return fmt.Errorf("Address parameter is required for Transmission receiver")
	}
	_url, err := url.Parse(config["address"].(string))
	if err != nil {
		return fmt.Errorf("Cannot parse Transmission address: `%s`", err)
	}
	if config["path"] != nil {
		_url.Path = config["path"].(string)
	} else {
		_url.Path = "transmission/rpc"
	}

	t.url = _url.String()
	t.client = &http.Client{}

	if config["shorts"] != nil {
		t.shorts = config["shorts"].(bool)
	}

	return nil
}

func init() {
	registry.AddReceiver("Transmission", &Transmission{}, transmissionResponse{})
}
