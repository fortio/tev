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

```
tev help
```

By defaults it will put the terminal in raw mode, turn on mouse tracking and show exactly what the terminal emulator is sending and in how many batches (of up to 1024 which is the internal ansipixels buffer size). Various flag allow to change what the terminal does (raw, mouse, bracketed paste, etc..)
