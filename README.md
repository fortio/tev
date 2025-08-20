<!-- [![GoDoc](https://godoc.org/fortio.org/tev?status.svg)](https://pkg.go.dev/fortio.org/tev) -->
[![Go Report Card](https://goreportcard.com/badge/fortio.org/tev)](https://goreportcard.com/report/fortio.org/tev)
[![CI Checks](https://github.com/fortio/tev/actions/workflows/include.yml/badge.svg)](https://github.com/fortio/tev/actions/workflows/include.yml)
# tev

`tev` is a simple terminal/TUI event dump/decoder - similar to good old `xev` for X11. It helps debug raw terminal input.

## Install
You can get the binary from [releases](https://github.com/fortio/tev/releases)

Or just run
```
CGO_ENABLED=0 go install fortio.org/tev@latest  # to install (in ~/go/bin typically) or just
CGO_ENABLED=0 go run fortio.org/tev@latest  # to run without install
```

or even
```
docker run -ti fortio/tev # but that's obviously slower
```

or
```
brew install fortio/tap/tev
```

## Run

```sh
tev help
```

```sh
  -code string
         Additional code to send (will be unquoted, eg "\033[..." will send CSI code)
  -echo
         Echo input to stdout instead of logging escaped bytes, also turns off mouse tracking
  -mouse-clicks
         Enable mouse click events (instead of movement)
  -mouse-pixels
         Enable mouse pixel events (vs grid)
  -mouse-x10
         Enable mouse X10 events mode
  -no-bg-color-query
         Don't query terminal for background color
  -no-mouse
         Disable mouse tracking events (enabled by default)
  -no-paste-mode
         Disable bracketed paste mode
  -no-raw
         Stay in cooked mode, instead of defaulting to raw mode
  -quiet
         Quiet mode, sets loglevel to Error (quietly) to reduces the output
```

By default it will put the terminal in raw mode, turn on mouse tracking and show exactly what the terminal emulator is sending and in how many batches (of up to 1024 which is the internal ansipixels buffer size). Various flag allow to change what the terminal does (raw, mouse, bracketed paste, etc..)


Example output
```sh
$ tev -no-mouse
20:40:23.063 [INF] Mouse tracking disabled
20:40:23.064 [INF] Bracketed paste mode enabled
20:40:23.064 [INF] Fortio terminal event dump started. ^C 3 times to exit (or pkill tev). Ctrl-L clears the screen.
20:40:23.064 [INF] Tabs: [8 16 24 32 40 48 56 64 72 80 88 96]
20:40:25.239 [INF] Read 1 bytes: "a"
20:40:26.228 [INF] Read 1 bytes: "b"
20:40:26.502 [INF] Read 1 bytes: "c"
20:41:06.309 [INF] Read 6 bytes: "\x1b[200~"
20:41:06.309 [INF] Read 26 bytes: "line 1\nline 2\nline 3\x1b[201~"
20:41:09.359 [INF] Read 1 bytes: "\x03"
20:41:09.359 [INF] Ctrl-C received, 2 more to exit..
20:41:09.899 [INF] Read 1 bytes: "\x03"
20:41:09.899 [INF] Ctrl-C received, 1 more to exit..
20:41:10.428 [INF] Read 1 bytes: "\x03"
20:41:10.428 [INF] 3rd Ctrl-C received, exiting now.```
```

You can also see the effect of some codes:
```sh
$ tev -code '\033]11;?\007'
17:15:06.997 [INF] Sending code flag "\x1b]11;?\a"
17:15:06.998 [INF] Read 24 bytes: "\x1b]11;rgb:1e1e/1e1e/1e1e\a"
```

Which is now built-in (unless passing `-no-bg-color-query`):
```
20:06:38.480 [INF] Querying terminal's background color...
20:06:38.481 [INF] Read 24 bytes: "\x1b]11;rgb:1d1d/1e1e/1d1d\a"
20:06:38.481 [INF] OSC background color decoded: #1D1E1D
```
