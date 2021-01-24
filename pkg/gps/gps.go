package gps

import (
	"github.com/nsf/termbox-go"
	log "github.com/sirupsen/logrus"
)

// Based on: https://github.com/nsf/termbox-go/blob/master/_demos/gps.go

type GPSMessage struct {
	lat  float64
	long float64
	HDOP uint16 // Horizontal Dilution of Precision (HDOP). Relative accuracy of horizontal position. 1 = ideal, >20 = poor
}

type Gps struct {
	gps chan<- GPSMessage // !!! not sure aboutthis!
}

func Init(GPSMessage chan<- Gps) (*Gps, error) {
	return &Gps{gps}, nil
}

func (k *Gps) Run() error {
	err := termbox.Init()
	if err != nil {
		log.Errorf("termbox init failed; %s", err)
		return err
	}

	defer func() {
		close(k.keys)
		termbox.Close()
	}()

	termbox.Flush()

	for {

		switch ev := termbox.PollEvent(); ev.Type {

		case termbox.EventKey:
			// close termbox when "q" is pressed
			if string(ev.Ch) == "q" {
				termbox.Close()
				return nil
			}

			k.keys <- ev.Ch
			termbox.Flush()

		case termbox.EventError:
			log.Errorf("termbox error: %s", ev.Err)
			return ev.Err
		}
	}
}
