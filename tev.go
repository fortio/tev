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
	cli.Main()
	ap := ansipixels.NewAnsiPixels(0) // fps 0 means raw os.Stdin
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
		ap.Restore()
	}()
	crlfWriter := &terminal.CRLFWriter{Out: os.Stdout}
	terminal.LoggerSetup(crlfWriter)
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
	ap.Out.Flush()
	exitCount := 3
	log.Infof("Fortio terminal event dump started. ^C 3 times to exit (or pkill tev)")
	for {
		err := ap.ReadOrResizeOrSignal()
		if err != nil {
			return log.FErrf("Error reading terminal: %v", err)
		}
		if len(ap.Data) == 0 {
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
