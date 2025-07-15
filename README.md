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
  -echo
        Echo input to stdout instead of logging escaped bytes, also turns off mouse tracking
  -mouse-pixels
        Enable mouse pixel events (vs grid)
  -mouse-x10
        Enable mouse X10 events mode
  -no-mouse
        Disable mouse tracking events (enabled by default)
  -no-paste-mode
        Disable bracketed paste mode
  -no-raw
        Stay in cooked mode, instead of defaulting to raw mode
```

By default it will put the terminal in raw mode, turn on mouse tracking and show exactly what the terminal emulator is sending and in how many batches (of up to 1024 which is the internal ansipixels buffer size). Various flag allow to change what the terminal does (raw, mouse, bracketed paste, etc..)


Example output
```sh
$ tev -no-mouse
13:28:27.673 [INF] Mouse tracking disabled
13:28:27.673 [INF] Bracketed paste mode enabled
13:28:27.673 [INF] Fortio terminal event dump started. ^C 3 times to exit (or pkill tev). Ctrl-L clears the screen.
13:28:35.068 [INF] Read 1 bytes: "a"
13:28:35.201 [INF] Read 1 bytes: "b"
13:28:35.372 [INF] Read 1 bytes: "c"
13:28:38.776 [INF] Read 6 bytes: "\x1b[200~"
13:28:38.776 [INF] Read 26 bytes: "line 1\nline 2\nline 3\x1b[201~"
13:28:42.056 [INF] Read 1 bytes: "\x03"
13:28:42.056 [INF] Ctrl-C received, 2 more to exit..
13:28:43.249 [INF] Read 1 bytes: "\x03"
13:28:43.249 [INF] Ctrl-C received, 1 more to exit..
13:28:43.850 [INF] Read 1 bytes: "\x03"
13:28:43.850 [INF] 3rd Ctrl-C received, exiting now.
```
