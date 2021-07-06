// GPIOTEST.go
// read switch input from raspberry pi 3+ GPIO and light LED
// uses command-line GPIOD.
// debouncing handled

// to playback audio must run as sudo:  go build GPIOTEST_dog.go && sudo ./GPIOTEST_dog

package main

import (
	"fmt"
	"math/rand"
	"sync"

	// "log"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/johnusher/ardpifi/pkg/wavs"
	log "github.com/sirupsen/logrus"

	"github.com/warthog618/gpiod"
)

// const (
// 	bounceTime = 150 * time.Millisecond
// )

type buttonPress struct {
	buttonDownTime  time.Time
	buttonFlag      int16
	button          gpiod.Line
	led             gpiod.Line
	buttonWavs      wavs.Wavs
	cancelButtonWav []chan struct{}
	sync.Mutex      // protects cancelButtonWav
}

// enqueue creates a new channel for each button pressed down (buttonStatus == 0)
// The new channel is added to the cancelButtonWav slice, and returned to the
// caller. Channel is closed via `flush`.
func (b *buttonPress) enqueue() chan struct{} {
	b.Lock()
	c := make(chan struct{})
	b.cancelButtonWav = append(b.cancelButtonWav, c)
	b.Unlock()
	return c
}

// flush closes all channels created via `enqueue`, and then clears the slice of
// channels, `cancelButtonWav`.
func (b *buttonPress) flush() {
	b.Lock()
	for _, c := range b.cancelButtonWav {
		close(c)
	}
	b.cancelButtonWav = make([]chan struct{}, 0)
	b.Unlock()
}

// var led gpiod.Line

// var button gpiod.Line

// var newtimer time.Timer
// var newtimer2 time.Timer
// var wavss wavs.Wavs
// var cancelButtonWav = make(chan struct{})

func delayedButtonHandle(pushButton *buttonPress) {
	buttonStatus, _ := pushButton.button.Value() // Read state from line (active / inactive)

	if buttonStatus == 0 { // low= button pressed down
		pushButton.led.SetValue(1) // light LED

		// first down: restart timer
		t := time.Now()
		pushButton.buttonDownTime = t

		// defer newtimer.Stop() // stop countdown timer

		// play long howl
		catMeowN := rand.Int31n(2) + 1
		catcat := fmt.Sprintf("howl%d.wav", catMeowN)

		// fmt.Println(catcat)
		// wavss.Play(catcat)

		// newtimerp2 := time.AfterFunc(150*time.Millisecond, func() { wavss.Play(catcat) }) // play wav after 150 ms

		// newtimer2 = *newtimerp2
		// defer newtimer2.Stop()

		// pushButton.cancelButtonWav = make(chan struct{})
		cancelButtonWav := pushButton.enqueue()
		go func() {
			// either play after 150ms, or bail if close(cancelButtonWav) is called
			select {
			case <-time.After(150 * time.Millisecond):
				log.Info("howl!", catcat)
				pushButton.buttonWavs.Play(catcat)
			case <-cancelButtonWav:
			}
		}()

		// wavss.Play("meow_1.wav")
	} else {
		// button has been lifted
		pushButton.led.SetValue(0) // turn off LED
		now := time.Now()
		elapsedTime := now.Sub(pushButton.buttonDownTime)

		pushButton.flush()
		// newtimer2.Stop()
		pushButton.buttonWavs.StopAll()

		if elapsedTime < 400*time.Millisecond {
			// play short bark:
			catMeowN2 := rand.Int31n(9) + 1
			catcat2 := fmt.Sprintf("bark%d.wav", catMeowN2)
			pushButton.buttonWavs.Play(catcat2)
			fmt.Println(elapsedTime)
		}

	}
	pushButton.buttonFlag = 0

}

func mkButtonEventHandler(pushButton *buttonPress) func(gpiod.LineEvent) {
	return func(evt gpiod.LineEvent) {
		if pushButton.buttonFlag == 0 {

			pushButton.buttonFlag = 1 // flag =1 , ie button active

			go func() {
				time.Sleep(3 * time.Millisecond)
				delayedButtonHandle(pushButton)
			}()
			// newtimerp := time.AfterFunc(3*time.Millisecond, delayedButtonHandle) // debounce: after 3 ms, check status again
			// newtimer = *newtimerp

			// defer newtimer.Stop()

			// newtimer := time.NewTimer(100 * time.Millisecond) // start timer for 100 ms, when expired, check GPIO level
		} else {
			// timer already running
			log.Info("bounce")
			return
		}
	}
}

func main() {

	wavsp := wavs.InitWavs()

	pushButton := &buttonPress{
		buttonFlag:      0,
		buttonDownTime:  time.Now(),
		cancelButtonWav: make([]chan struct{}, 0),
	}
	pushButton.buttonWavs = *wavsp
	buttonEventHandler := mkButtonEventHandler(pushButton)

	// hack from https://www.raspberrypi.org/forums/viewtopic.php?t=270376:

	//  Physical pin 13 = BCM pin 27, GPIO27 = J8p13
	// gpio readall

	// The library uses the raw BCM2835 pin numbers, not the ports as they are mapped
	// on the J8 output pins for the Raspberry Pi.
	// A mapping from J8 to BCM is provided for those wanting to use the J8 numbering.
	// eg physica; pin

	app := "gpio"

	arg0 := "-g"
	arg1 := "mode"
	arg2 := "27"
	arg3 := "in"

	cmd := exec.Command(app, arg0, arg1, arg2, arg3)
	log.Printf("gpio set-up part 1")
	err := cmd.Run()
	if err != nil {
		log.Printf("Command finished with error: %v", err)
	}

	arg2 = "22"
	arg3 = "out"

	cmd = exec.Command(app, arg0, arg1, arg2, arg3)
	log.Printf("gpio set-up part 1.1")
	err = cmd.Run()
	if err != nil {
		log.Printf("Command finished with error: %v", err)
	}

	// Set up GPIO 27 with pullup resistor
	app = "gpio"

	arg0 = "-g"
	arg1 = "mode"
	arg2 = "27"
	arg3 = "up"

	cmd = exec.Command(app, arg0, arg1, arg2, arg3)
	log.Printf("gpio set-up part 2...")
	err = cmd.Run()
	if err != nil {
		log.Printf("Command finished with error: %v", err)
	}

	log.Printf("gpio set-up part 3...")
	c, err := gpiod.NewChip("gpiochip0")
	if err != nil {
		panic(err)
	}

	defer c.Close()

	// Set up button with interrupt watch using gpiod
	// offset := rpi.J8p13
	offset := 27
	buttonp, err := c.RequestLine(offset,
		gpiod.WithBothEdges,
		gpiod.WithEventHandler(buttonEventHandler))

	if err != nil {
		fmt.Printf("RequestLine returned error: %s\n", err)
		if err == syscall.Errno(22) {
			fmt.Println("Note that the WithPullUp option requires kernel V5.5 or later - check your kernel version.")
		}
		os.Exit(1)
	}

	pushButton.button = *buttonp

	defer pushButton.button.Close()

	// Set up button with interrupt watch using gpiod
	// offset := rpi.J8p13
	offset = 22
	ledp, err := c.RequestLine(offset, gpiod.AsOutput(0)) // during request

	pushButton.led = *ledp

	// NB remove pullup from the gpiod function call: requires kernel 5.5 for pullup/pulldown support.
	if err != nil {
		fmt.Printf("RequestLine2 returned error: %s\n", err)
		os.Exit(1)
	}
	defer pushButton.led.Close()

	fmt.Printf("Watching Pin %d...\n", offset)
	time.Sleep(time.Hour)
	fmt.Println("exiting...")

}
