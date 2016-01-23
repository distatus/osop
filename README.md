[![Build Status](https://travis-ci.org/KenjiTakahashi/osop.png?branch=master)](https://travis-ci.org/KenjiTakahashi/osop)

**osop** - other side of *the* pipe - outputs formatted metrics to your Stdout.

## screenshots

Plain output to terminal:

![screenshot1](https://copy.com/IjgI8uHhhxGK6HzE)

Paired with [gobar](https://github.com/KenjiTakahashi/gobar):

![screenshot2](https://copy.com/qJKvFzqROUoBqCri)

## installation

First, you have to [get Go](http://golang.org/doc/install). Note that version >= 1.2 is required.

Then, just

```bash
$ go get github.com/KenjiTakahashi/osop
```

should get you GOing.

## usage

The only command line switch is `-c` which specifies a configuration file location. Location can either be absolute or relative to `$XDG_CONFIG_DIR/osop`. Defaults to `$XDG_CONFIG_DIR/osop/config.toml`.

Configuration file uses [toml](https://github.com/toml-lang/toml) language and consists of two sections.

The mandatory **Osop** section.

```toml
[Osop]
delims = ["<", ">"]
template = ""
```

Where the required **template** is a text/template string and optional **delims** specify action delimiters (defaults to `<` and `>`)

In addition to standard actions, a `stringify` action is defined to take one argument and *always* return (possibly empty) string no matter what. This proved to be useful in some cases.

Zero or more **Receiver** sections.

```toml
[Name]
receiver = "date"
pollInterval = "1s"
otherSetting1 = "setsth"
otherSetting2 = "setsthelse"
```

**Receiver** represents a single data unit, such as date, system information, weather, etc. Each is configured with `[Name]` which will be used inside **Osop** template string and `receiver`, which tells what receiver to run. *Receivers are case insensitive.*

Different receivers might use different strategies to get the data. Some are evented (passively waiting for data to arrive), others are actively polling for data on time interval. Time interval is configured with `pollInterval`, which is required for polling receivers and ignored by evented ones.

Other settings might be exposed as needed by specific receivers.

For available receivers, their settings and output format(s), see [receivers](#receivers) section.

For a more real world examples, see [my dotfiles](https://github.com/KenjiTakahashi/dotfiles/tree/master/dotconfig/osop).

### receivers

#### date

Current date and/or time.

**Configuration:**

* format *(required)* - [Golang style](http://golang.org/pkg/time/#Time.Format) date format string.

**Output:** String.

#### battery

Current battery state.

**Configuration:**

* number *(optional)* - Battery number/index. *Defaults to 0*.

**Output:** Struct:

* Charge - Current charge number as returned by the system.
* Percent - Current charge in percentage form.
* State - Possible values: "Empty", "Full", "Charging", "Discharging", "Unknown".

#### sys

System metrics.

**Configuration:**

* metrics *(required)* - List of any number of values from:
    * cpu percent *[true|false]* - Current CPU usage in percents, *[per core|cumulative]*.
    * uptime
    * memory
    * swap
    * network *&lt;interface>* - Network statistics for given interface (e.g. *wlan0*).
* shorts *(optional)* - Use short (*K*) units, instead of full (*KB*). *Defaults to false.*

**Output:** Struct:

* CPU
    * Percent - Dictionary of *cpu0*, *cpu1*, etc. For cumulative, only *cpu0* is filled.
* Uptime
* Memory
    * Total
    * UsedF
    * UsedA
* Swap
    * Total
    * Used
* Network - Dictionary of network interface names to Struct:
    * Sent
    * Recv
    * Download
    * Upload

*Note that only parts relevant to values set in `metrics` will actually be filled.*

#### owm

Weather information based on OpenWeatherMap service.

**Configuration:**

* location *(required)* - Either "City,Country Code" *(e.g. "London,UK")* or a location code.
* apiKey *(required)* - OpenWeatherMap API key.
* units *(optional)* - Either "metric" or "imperial". *Defaults to "metric".*

**Output:** Struct:

* City
* Country
* Sunrise
* Sunset
* Temp
* TempMin
* TempMax
* Pressure
* Humidity
* Wind
    * Speed
    * Deg
* Coord
    * Lon
    * Lat

#### transmission

Transmission daemon information.

**Configuration:**

* address *(required)* - URL to transmission RPC server.
* path *(optional)* - Path to transmission RPC server. *Defaults to "transmission/rpc".*
* shorts *(optional)* - Use short (*K/s*) units, instead of full (*KB/s*). *Defaults to false.*

**Output:** Struct:

* TorrentCount
* ActiveTorrentCount
* PausedTorrentCount
* DownloadSpeed
* UploadSpeed
* Cumulative
    * Downloaded
    * Uploaded
    * FilesAdded
    * SessionCount
    * SecondsActive
* Current
    * Downloaded
    * Uploaded
    * FilesAdded
    * SessionCount
    * SecondsActive

#### mpd

Mpd information.

**Configuration:**

* address *(required)* - URL to MPD server.
* password *(optional)* - Password to MPD server.

**Output:** Struct:

* Song - Dictionary with current song's metadata.
* Status - Dictionary with status, as described [here](http://www.musicpd.org/doc/protocol/command_reference.html).

#### bspwm

[Bspwm](https://github.com/baskerville/bspwm) status pieces.

**Output:** Struct:

* Monitors - List of Struct:
    * Name
    * State
    * Index
    * Desktops - List of Struct:
        * Name
        * State

#### wingo

[Wingo](https://github.com/BurntSushi/wingo) status pieces.

**Output:** Struct:

* Workspaces - List of Struct:
    * Name
    * Active
    * ActiveOn
    * Alerted
    * Layout
    * Clients - number of clients.
    * HasClients
