package main

import (
	"flag"
	"os"

	"fortio.org/cli"
	"fortio.org/log"
	"fortio.org/terminal/ansipixels"
)

func main() {
	os.Exit(Main())
}

func Main() int {
	noMouseFlag := flag.Bool("no-mouse", false, "Disable mouse tracking events (default is to enable it)")
	cli.Main()
	ap := ansipixels.NewAnsiPixels(20.)
	err := ap.Open()
	if err != nil {
		return log.FErrf("Failed to open terminal: %v", err)
	}
	defer func() {
		ap.MoveCursor(0, ap.H-1)
		ap.MouseTrackingOff()
		ap.MouseClickOff()
		ap.Restore()
	}()
	ap.OnResize = func() error {
		log.Infof("Terminal resized to %dx%d\r", ap.W, ap.H)
		return nil
	}
	ap.NoDecode = true // we handle mouse events ourselves
	if !*noMouseFlag {
		ap.MouseTrackingOn()
		ap.Out.Flush()
		log.Infof("Mouse tracking enabled\r")
	} else {
		log.Infof("Mouse tracking disabled\r")
	}
	cont := 3
	// TODO plug the writer that auto connverts \n to \r\n
	log.Infof("Fortio terminal event dump started. ^C 3 times to exit (or pkill tev)\r")
	for {
		err = ap.ReadOrResizeOrSignal()
		if err != nil {
			return log.FErrf("Error reading terminal: %v", err)
		}
		if len(ap.Data) == 0 {
			log.Infof("No input...\r")
			continue
		}
		log.Infof("Read %d bytes: %q\r", len(ap.Data), ap.Data)
		ap.MouseDecode()
		if ap.Mouse {
			log.Infof("Mouse event detected: buttons %06b, x %d, y %d\r", ap.Mbuttons, ap.Mx, ap.My)
			continue
		}
		if ap.Data[0] == 3 { // Ctrl-C
			cont--
			if cont == 0 {
				log.Infof("3rd Ctrl-C received, exiting now.\r")
				return 0
			}
			log.Infof("Ctrl-C received, %d more to exit..\r", cont)
		} else {
			cont = 3 // reset count on any other input
		}
	}
	return 0
}
