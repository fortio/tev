package main

import (
	"flag"
	"os"

	"fortio.org/cli"
	"fortio.org/log"
	"fortio.org/terminal"
	"fortio.org/terminal/ansipixels"
)

func main() {
	os.Exit(Main())
}

// GetTabStops measures the current tab stop positions.
// Taken/copied from fortio/gvi.
func GetTabStops(ap *ansipixels.AnsiPixels) []int {
	ap.WriteString("\r\t")
	var tabs []int
	prevX := 0
	for {
		x, _, err := ap.ReadCursorPosXY()
		if err != nil {
			log.Errf("Error reading cursor position: %v", err)
			return nil
		}
		if x == prevX || x == ap.W-1 {
			break
		}
		tabs = append(tabs, x)
		ap.WriteString("\t")
		prevX = x
	}
	return tabs
}

func Main() int {
	noMouseFlag := flag.Bool("no-mouse", false, "Disable mouse tracking events (enabled by default)")
	mousePixelsFlag := flag.Bool("mouse-pixels", false, "Enable mouse pixel events (vs grid)")
	mouseX10Flag := flag.Bool("mouse-x10", false, "Enable mouse X10 events mode")
	noPasteModeFlag := flag.Bool("no-paste-mode", false, "Disable bracketed paste mode")
	fpsFlag := flag.Float64("fps", 0,
		"Ansi pixels debug/complex mode - fps arg (default is 0, meaning simplest code in ansipixels: blocking mode reads)")
	noRawFlag := flag.Bool("no-raw", false, "Stay in cooked mode, instead of defaulting to raw mode")
	echoFlag := flag.Bool("echo", false, "Echo input to stdout instead of logging escaped bytes, also turns off mouse tracking")
	cli.Main()
	ap := ansipixels.NewAnsiPixels(*fpsFlag) // use the specified fps - if 0, it will be blocking mode.
	if !*noRawFlag {
		err := ap.Open()
		if err != nil {
			return log.FErrf("Failed to open terminal: %v", err)
		}
		crlfWriter := &terminal.CRLFWriter{Out: os.Stdout}
		terminal.LoggerSetup(crlfWriter)
	} else {
		log.LogVf("Not enabling raw mode, staying in cooked mode")
		_ = ap.GetSize() // to set ah.H for restore.
	}
	defer func() {
		// do it even in cooked mode to turn off mouse spam etc...
		ap.MoveCursor(0, ap.H-1)
		ap.MousePixelsOff()
		ap.MouseX10Off()
		ap.MouseTrackingOff()
		ap.MouseClickOff()
		ap.SetBracketedPasteMode(false)
		ap.Restore()
	}()
	ap.OnResize = func() error {
		log.Infof("Terminal resized to %dx%d", ap.W, ap.H)
		return nil
	}
	ap.NoDecode = true // we handle mouse events ourselves
	echoMode := *echoFlag
	logLevel := log.Info
	if echoMode {
		logLevel = log.Verbose
	}
	if !echoMode && !*noMouseFlag {
		ap.MouseTrackingOn()
		log.Infof("Mouse tracking enabled")
	} else {
		log.Infof("Mouse tracking disabled")
	}
	if *mousePixelsFlag {
		ap.MousePixelsOn()
		log.Infof("Mouse pixel events enabled")
	}
	if *mouseX10Flag {
		ap.MouseX10On()
		log.Infof("Mouse X10 events enabled")
	}
	if !*noPasteModeFlag {
		ap.SetBracketedPasteMode(true)
		log.Infof("Bracketed paste mode enabled")
	} else {
		log.Infof("Bracketed paste mode disabled")
	}
	ap.Out.Flush()
	exitCount := 3
	log.Infof("Fortio terminal event dump started. ^C 3 times to exit (or pkill tev). Ctrl-L clears the screen.")
	if !*noRawFlag {
		log.Infof("Tabs: %v", GetTabStops(ap))
	} else {
		log.Infof("Sample tabs:\n\t0\t1\t2\t3\t4\t5\t6\t7\t8")
	}
	for {
		err := ap.ReadOrResizeOrSignal()
		if err != nil {
			return log.FErrf("Error reading terminal: %v", err)
		}
		if len(ap.Data) == 0 { // not really possible.
			log.Warnf("No input (unexpected)...")
			continue
		}
		log.Logf(logLevel, "Read %d bytes: %q", len(ap.Data), ap.Data)
		if echoMode {
			os.Stdout.Write(ap.Data)
		} else {
			ap.MouseDecode()
			if ap.Mouse {
				log.Infof("Mouse event detected: buttons %06b, x %d, y %d", ap.Mbuttons, ap.Mx, ap.My)
				continue
			}
		}
		switch ap.Data[0] {
		case 3: // Ctrl-C
			exitCount--
			if exitCount == 0 {
				log.Infof("3rd Ctrl-C received, exiting now.")
				return 0
			}
			log.Infof("Ctrl-C received, %d more to exit..", exitCount)
		case '\f': // Ctrl-L
			log.Infof("Ctrl-L received, clearing screen.")
			ap.ClearScreen()
			ap.Out.Flush()
			fallthrough // also reset ^C count.
		default:
			exitCount = 3 // reset count on any other input
		}
	}
}
