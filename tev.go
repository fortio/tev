package main

import (
	"flag"
	"os"
	"strconv"

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

func Restore(ap *ansipixels.AnsiPixels) {
	ap.MoveCursor(0, ap.H-1)
	ap.MousePixelsOff()
	ap.MouseX10Off()
	ap.MouseTrackingOff()
	ap.MouseClickOff()
	ap.SetBracketedPasteMode(false)
	ap.Restore()
}

func Main() int {
	noMouseFlag := flag.Bool("no-mouse", false, "Disable mouse tracking events (enabled by default)")
	mousePixelsFlag := flag.Bool("mouse-pixels", false, "Enable mouse pixel events (vs grid)")
	mouseX10Flag := flag.Bool("mouse-x10", false, "Enable mouse X10 events mode")
	mouseClickFlag := flag.Bool("mouse-clicks", false, "Enable mouse click events (instead of movement)")
	noPasteModeFlag := flag.Bool("no-paste-mode", false, "Disable bracketed paste mode")
	fpsFlag := flag.Float64("fps", 0,
		"Ansi pixels debug/complex mode - fps arg (default is 0, meaning simplest code in ansipixels: blocking mode reads)")
	fpsticksFlag := flag.Bool("ticks", false, "Ansi pixels debug to use the FPSTicks loop instead of ReadOrResizeOrSignal")
	noRawFlag := flag.Bool("no-raw", false, "Stay in cooked mode, instead of defaulting to raw mode")
	echoFlag := flag.Bool("echo", false, "Echo input to stdout instead of logging escaped bytes, also turns off mouse tracking")
	codeFlag := flag.String("code", "", "Additional code to send (will be unquoted, eg \"\\033[...\" will send CSI code)")
	noBackgroundFlag := flag.Bool("no-bg-color-query", false, "Don't query terminal for background color")
	cli.Main()
	ap := ansipixels.NewAnsiPixels(*fpsFlag) // use the specified fps - if 0, it will be blocking mode.
	// We do logger setup ourselves below after opening the terminal to not get buffered/needing flushing.
	ap.AutoLoggerSetup = false
	extra := " 3 times"
	if !*noRawFlag {
		err := ap.Open()
		if err != nil {
			return log.FErrf("Failed to open terminal: %v", err)
		}
		crlfWriter := &terminal.CRLFWriter{Out: os.Stdout}
		terminal.LoggerSetup(crlfWriter)
	} else {
		log.LogVf("Not enabling raw mode, staying in cooked mode")
		extra = ""
		_ = ap.GetSize() // to set ap.H for restore.
	}
	// do it even in cooked mode to turn off mouse spam etc...
	defer Restore(ap)
	ap.OnResize = func() error {
		log.Infof("Terminal resized to %dx%d", ap.W, ap.H)
		return nil
	}
	ap.NoDecode = true // we handle mouse events ourselves
	echoMode := *echoFlag
	if *mouseClickFlag {
		ap.MouseClickOn()
		log.Infof("Mouse click events enabled")
		*noMouseFlag = true // mouse click implies no mouse tracking
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
	if *fpsticksFlag {
		log.Infof("Fortio terminal simplified FPSTicks event dump. ^C times to exit (or pkill tev).")
		return DebugLoopFPSTicks(ap)
	}
	switch {
	case *codeFlag != "":
		inp := "\"" + *codeFlag + "\""
		dec, err := strconv.Unquote(inp)
		if err != nil {
			return log.FErrf("Invalid quoted string %s: %v", inp, err)
		}
		log.Infof("Sending code flag %q", dec)
		ap.WriteString(dec)
	case !*noRawFlag:
		log.Infof("Tabs: %v", GetTabStops(ap))
	default:
		log.Infof("Sample tabs:\n\t0\t1\t2\t3\t4\t5\t6\t7\t8")
	}
	if !*noBackgroundFlag && !*noRawFlag {
		log.Infof("Querying terminal's background color...")
		ap.RequestBackgroundColor()
	} else {
		ap.GotBackground = true // pretend we already go it so we don't keep trying.
	}
	ap.Out.Flush()

	log.Infof("Fortio terminal event dump started. ^C%s to exit (or pkill tev). Ctrl-L clears the screen.", extra)
	return DebugLoop(ap, echoMode)
}

func ProcessMouse(ap *ansipixels.AnsiPixels, leftOver []byte) ([]byte, ansipixels.MouseStatus) {
	ap.Data = append(leftOver, ap.Data...) // prepend any left over from previous read
	leftOver = leftOver[:0]                // reset left over
	for {
		dec := ap.MouseDecode()
		switch dec {
		case ansipixels.MousePrefix:
			// doesn't seem to ever happen again with 1006h, did with short form on windows terminal.
			log.Infof("Partial/split mouse event, reading more...")
			leftOver = append(leftOver, ap.Data...)
			continue // wait for more data
		case ansipixels.MouseComplete:
			log.Infof("\tMouse event detected: buttons/modifiers %06b %s x %d, y %d", ap.Mbuttons, ap.MouseDebugString(), ap.Mx, ap.My)
		default:
			// no mouse or error already logged.
			return leftOver, dec
		}
	}
}

func DebugLoop(ap *ansipixels.AnsiPixels, echoMode bool) int {
	logLevel := log.Info
	if echoMode {
		logLevel = log.Verbose
	}
	exitCount := 3
	var leftOver []byte
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
		if !ap.GotBackground && ap.OSCDecode() {
			log.Infof("OSC background color decoded: %s", ap.Background)
			if len(ap.Data) == 0 {
				continue
			}
		}
		if echoMode {
			os.Stdout.Write(ap.Data)
		} else {
			var st ansipixels.MouseStatus
			if leftOver, st = ProcessMouse(ap, leftOver); st == ansipixels.MousePrefix {
				continue
			}
			if len(ap.Data) == 0 {
				// had nothing but mouse.
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
		case '\r':
			if echoMode {
				os.Stdout.Write([]byte{'\n'}) // echo \r as \n
			}
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

// DebugLoopFPSTicks is a simplified loop to check FPSTicks mode.
func DebugLoopFPSTicks(ap *ansipixels.AnsiPixels) int {
	exitCount := 3
	err := ap.FPSTicks(func() bool {
		l := len(ap.Data)
		log.LogVf("FPSTicks tick, data len=%d", l)
		if l == 0 { // not really possible.
			return true
		}
		log.Logf(log.Info, "Read %d bytes: %q", len(ap.Data), ap.Data)
		switch ap.Data[0] {
		case 3: // Ctrl-C
			exitCount--
			if exitCount == 0 {
				log.Infof("3rd Ctrl-C received, exiting now.")
				return false
			}
			log.Infof("Ctrl-C received, %d more to exit..", exitCount)
		default:
			exitCount = 3 // reset count on any other input
		}
		return true
	})
	if err != nil {
		return log.FErrf("Error reading terminal: %v", err)
	}
	return 0
}
