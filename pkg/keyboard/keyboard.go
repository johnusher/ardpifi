package keyboard

import (
	"github.com/nsf/termbox-go"
	log "github.com/sirupsen/logrus"
)

// Based on: https://github.com/nsf/termbox-go/blob/master/_demos/keyboard.go

type Keyboard struct {
	keys chan<- rune
}

func Init(keys chan<- rune) (*Keyboard, error) {
	err := termbox.Init()
	if err != nil {
		return nil, err
	}

	return &Keyboard{keys}, nil
}

func (k *Keyboard) Run() error {
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
