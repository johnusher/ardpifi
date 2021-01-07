package localkeyboard

// READS INPUT FROM USB PLUGGED IN KEYBOARD

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"unicode/utf8"

	log "github.com/sirupsen/logrus"
)

type Keyboard struct {
	keys chan<- rune
}

func Init(keys chan<- rune) (*Keyboard, error) {
	// err := termbox.Init()
	// if err != nil {
	// 	return nil, err
	// }

	// return &Keyboard{keys}, nil

	_, localKB_det := os.Open(findUSBKeyboard())
	if localKB_det != nil {
		return nil, localKB_det
		log.Print("No local USB keyboard found")
	}

	return &Keyboard{keys}, nil
}

func (k *Keyboard) Run() error {

	// check if we have USB keyboard attached:
	localKB, localKB_det := os.Open(findUSBKeyboard())
	if localKB_det != nil {
		log.Print("No local USB keyboard found")
	} else {
		defer localKB.Close() // do we need to do this?
	}

	b := make([]byte, 4)

	for {

		// // b := make(chan rune)
		// // b := make([]rune, 1)

		localKB.Read(b)
		// fmt.Printf("XX %b\n", b)

		r, _ := utf8.DecodeRune(b)

		if string(r) == "q" {
			return nil
		}

		if string(r) == "a" {
			k.keys <- r
			fmt.Printf("%b\n", b)
		}

		// k.keys <- r
		// fmt.Printf("%b\n", b)

		// switch ev := termbox.PollEvent(); ev.Type {

		// case termbox.EventKey:
		// 	// close termbox when "q" is pressed
		// 	if string(ev.Ch) == "q" {
		// 		return nil
		// 	}

		// 	k.keys <- ev.Ch
		// 	termbox.Flush()

		// case termbox.EventError:
		// 	log.Errorf("termbox error: %s", ev.Err)
		// 	return ev.Err
		// }
	}
}

func findUSBKeyboard() string {
	contents, _ := ioutil.ReadDir("/dev/input")

	// Look for what is mostly likely the Arduino device
	// NB this is kinda janky- we should have a system to robustly detect a duino, eg if we dont find one, then re-insert the duino USb cable and note which ports are new

	// JU: on my RASPI it shows in ttyAMA0
	for _, f := range contents {
		if strings.Contains(f.Name(), "event") {
			fmt.Println("USB KB found at /dev/input/", f.Name())
			return "/dev/input/" + f.Name()
		}
	}

	// Have not been able to find a USB device that 'looks'
	// like an Arduino.
	return ""
}

// package keyboard

// import (
// 	"github.com/nsf/termbox-go"
// 	log "github.com/sirupsen/logrus"
// )

// // Based on: https://github.com/nsf/termbox-go/blob/master/_demos/keyboard.go

// type Keyboard struct {
// 	keys chan<- rune
// }

// func Init(keys chan<- rune) (*Keyboard, error) {
// 	err := termbox.Init()
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &Keyboard{keys}, nil
// }

// func (k *Keyboard) Run() error {
// 	defer func() {
// 		close(k.keys)
// 		termbox.Close()
// 	}()

// 	termbox.Flush()

// 	for {
// 		switch ev := termbox.PollEvent(); ev.Type {
// 		case termbox.EventKey:
// 			// close termbox when "q" is pressed
// 			if string(ev.Ch) == "q" {
// 				return nil
// 			}

// 			k.keys <- ev.Ch
// 			termbox.Flush()

// 		case termbox.EventError:
// 			log.Errorf("termbox error: %s", ev.Err)
// 			return ev.Err
// 		}
// 	}
// }
