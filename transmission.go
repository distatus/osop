package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/pyk/byten"
)

func bytonize(data []byte, speed bool, short bool) (string, error) {
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
	b, err := bytonize(data, false, false)
	*s = stringByten(b)
	return err
}

type stringBytenShort string

func (s *stringBytenShort) UnmarshalJSON(data []byte) error {
	b, err := bytonize(data, false, true)
	*s = stringBytenShort(b)
	return err
}

type stringBytenSpeed string

func (s *stringBytenSpeed) UnmarshalJSON(data []byte) error {
	b, err := bytonize(data, true, false)
	*s = stringBytenSpeed(b)
	return err
}

type stringBytenSpeedShort string

func (s *stringBytenSpeedShort) UnmarshalJSON(data []byte) error {
	b, err := bytonize(data, true, true)
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

func (t *Transmission) Get() interface{} {
	req, err := http.NewRequest(
		"POST", t.url, bytes.NewBufferString(`{"method":"session-stats"}`),
	)
	if err != nil {
		log.Printf("Error creating Transmission request: `%s`", err)
		return nil
	}
	req.Header.Add("X-Transmission-Session-Id", t.sessionId)
	resp, err := t.client.Do(req)
	if err != nil {
		log.Printf("Error sending Transmission request: `%s`", err)
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode == 409 {
		t.sessionId = resp.Header.Get("X-Transmission-Session-Id")
		return t.Get()
	}
	if resp.StatusCode != 200 {
		return nil
	}

	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()
	if t.shorts {
		var data struct {
			Arguments transmissionResponseShort
		}
		decoder.Decode(&data)
		return data.Arguments
	} else {
		var data struct {
			Arguments transmissionResponse
		}
		decoder.Decode(&data)
		return data.Arguments
	}
}

func NewTransmission(config config) (interface{}, error) {
	if config["address"] == nil {
		return nil, fmt.Errorf("Address parameter is required for Owm receiver")
	}
	_url, err := url.Parse(config["address"].(string))
	if err != nil {
		return nil, fmt.Errorf("Cannot parse Transmission address: `%s`", err)
	}
	if config["path"] != nil {
		_url.Path = config["path"].(string)
	} else {
		_url.Path = "transmission/rpc"
	}

	t := &Transmission{
		url:    _url.String(),
		client: &http.Client{},
	}

	if config["shorts"] != nil {
		t.shorts = config["shorts"].(bool)
	}

	return t, nil
}

func init() {
	registry.AddReceiver("Transmission", NewTransmission)
}