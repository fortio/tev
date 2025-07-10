package main

import (
	"os"

	"fortio.org/cli"
	"fortio.org/log"
	"fortio.org/terminal/ansipixels"
)

func main() {
	os.Exit(Main())
}

func Main() int {
	cli.Main()
	ap := ansipixels.NewAnsiPixels(20.)
	err := ap.Open()
	if err != nil {
		return log.FErrf("Failed to open terminal: %v", err)
	}
	defer ap.Restore()
	ap.OnResize = func() error {
		log.Infof("Terminal resized to %dx%d\r", ap.W, ap.H)
		return nil
	}
	cont := 3
	// TODO plug the writer that auto connverts \n to \r\n
	log.Infof("Fortio terminal event dump started. ^C 3 times to exit (or pkill tev)\r")
	for cont > 0 {
		err = ap.ReadOrResizeOrSignal()
		if err != nil {
			return log.FErrf("Error reading terminal: %v", err)
		}
		if len(ap.Data) == 0 {
			log.Infof("No input...\r")
			continue
		}
		log.Infof("Read %d bytes: %q\r", len(ap.Data), ap.Data)
		if ap.Data[0] == 3 { // Ctrl-C
			cont--
			log.Infof("Ctrl-C received, %d more to exit..\r", cont)
		} else {
			cont = 3 // reset count on any other input
		}
	}
	ap.MoveCursor(0, ap.H-1)
	return 0
}
