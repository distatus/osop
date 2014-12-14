[![Build Status](https://travis-ci.org/KenjiTakahashi/osop.png?branch=master)](https://travis-ci.org/KenjiTakahashi/osop)

**osop**

## screenshots

Plain output to terminal:

**SHOT1**

Paired with [gobar](https://github.com/KenjiTakahashi/gobar):

**SHOT2**

## installation

First, you have to [get Go](http://golang.org/doc/install).

Then, just

```bash
$ go get github.com/KenjiTakahashi/osop
```

should get you GOing.

## usage

The only command line switch is `-c` which specifies a configuration file location. Location can either be absolute or relative to `$XDG_CONFIG_DIR/osop`. Defaults to `$XDG_CONFIG_DIR/osop/config.toml`.

Configuration file uses [toml](https://github.com/toml-lang/toml) language and looks like this.

```toml
[<Local name assigned to receiver>]
receiver = "<receiver name>" # Required.
pollInterval = "1s" # Optional for Evented Receivers.
<receiver specific settings>

[Osop] # Required
template = "" # A text/template style template.
```

Each receiver outputs some information whose are then available for use inside the template, under the &lt;Local name assigned to receiver&gt;.

For available receivers, their settings and output format(s), see [receivers](#receivers) section.

For a more real world examples, see [my dotfiles](https://github.com/KenjiTakahashi/dotfiles/dotconfig/osop)

### receivers

**date** - current date and time.

##### Configuration

```toml
format = "02/01/2006 | 15:04:05" # Required, output format the Golang way.
```

##### Output

String with current date, formatted as configured.

**sys** - different system metrics.

```toml
metrics = [] # Required, list of metrics to gather.
shorts = false # Optional (defaults to false), use full (KB) or short (K) units.
```

##### Configuration

Available metrics are:

* `CPU percent [false|true]` - optional argument indicates whether to gather per cpu.
* `Network <eth0>` - required argument specifies which interface to monitor.
* `Memory`
* `Swap`
* `Uptime`

##### Output

Structure with a field for each specified metric (order like above):

* map of `cpu<num>` [for `true`, `cpu0` gathers all info].
* map of `<interface>:{Download,Upload,Sent,Recv}`.
* struct with fields `Total,UsedF,UserA`.
* struct with fields `Total,Used`.
* a single number.

**owm** - weather information based on OpenWeatherMap service.

**transmission** - transmission daemon information.

**mpd** - mpd daemon information.

**bspwm** - [bspwm](https://github.com/baskerville/bspwm) status pieces.

**wingo** - [wingo](https://github.com/BurntSushi/wingo) status pieces.
