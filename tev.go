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

func Main() int {
	noMouseFlag := flag.Bool("no-mouse", false, "Disable mouse tracking events (default is to enable it)")
	mousePixelsFlag := flag.Bool("mouse-pixels", false, "Enable mouse pixel events (default is to disable it)")
	mouseX10Flag := flag.Bool("mouse-x10", false, "Enable mouse X10 events (default is to disable it)")
	noPasteModeFlag := flag.Bool("no-paste-mode", false, "Disable bracketed paste mode (default is to enable it)")
	fpsFlag := flag.Float64("fps", 0, "Ansi pixels debug/complex mode - fps arg (default is 0, meaning simplest code in ansipixels: blocking mode reads)")
	noRawFlag := flag.Bool("no-raw", false, "Stay in cooked mode, don't do raw mode (default is to enable it)")
	cli.Main()
	ap := ansipixels.NewAnsiPixels(*fpsFlag) // use the specified fps - if 0, it will be blocking mode.
	if !*noRawFlag {
		err := ap.Open()
		if err != nil {
			return log.FErrf("Failed to open terminal: %v", err)
		}
		defer func() {
			ap.MoveCursor(0, ap.H-1)
			ap.MousePixelsOff()
			ap.MouseX10Off()
			ap.MouseTrackingOff()
			ap.MouseClickOff()
			ap.SetBracketedPasteMode(false)
			ap.Restore()
		}()
		crlfWriter := &terminal.CRLFWriter{Out: os.Stdout}
		terminal.LoggerSetup(crlfWriter)
	} else {
		log.LogVf("Not enabling raw mode, staying in cooked mode")
	}
	ap.OnResize = func() error {
		log.Infof("Terminal resized to %dx%d", ap.W, ap.H)
		return nil
	}
	ap.NoDecode = true // we handle mouse events ourselves
	if !*noMouseFlag {
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
	log.Infof("Fortio terminal event dump started. ^C 3 times to exit (or pkill tev)")
	for {
		err := ap.ReadOrResizeOrSignal()
		if err != nil {
			return log.FErrf("Error reading terminal: %v", err)
		}
		if len(ap.Data) == 0 { // not really possible.
			log.Infof("No input...")
			continue
		}
		log.Infof("Read %d bytes: %q", len(ap.Data), ap.Data)
		ap.MouseDecode()
		if ap.Mouse {
			log.Infof("Mouse event detected: buttons %06b, x %d, y %d", ap.Mbuttons, ap.Mx, ap.My)
			continue
		}
		if ap.Data[0] == 3 { // Ctrl-C
			exitCount--
			if exitCount == 0 {
				log.Infof("3rd Ctrl-C received, exiting now.")
				return 0
			}
			log.Infof("Ctrl-C received, %d more to exit..", exitCount)
		} else {
			exitCount = 3 // reset count on any other input
		}
	}
}
