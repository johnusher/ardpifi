// GPIOTEST.gop

package main

import (
	"fmt"
	// "log"
	"os"
	"os/exec"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/warthog618/gpiod"
)

const (
	bounceTime = 150 * time.Millisecond
)

type buttonPress struct {
	lastMessageTypeReceived string
	lastButtonEventType     string
	lastMessageReceived     time.Time
}

var buttonTimes buttonPress // have to make this a global!!

func buttonEventHandler(evt gpiod.LineEvent) {
	oldTime := buttonTimes.lastMessageReceived
	t := time.Now()

	timeDiff := t.Sub(oldTime)

	if timeDiff < bounceTime {
		// log.Info("TOO QQUICK! timeDiff = ", timeDiff)
		return
	}
	log.Info("timeDiff = ", timeDiff)

	buttonTimes.lastMessageReceived = t

	edge := "rising"
	if evt.Type == gpiod.LineEventFallingEdge {
		edge = "falling"
	}

	fmt.Printf("%s \n", edge)
	// fmt.Printf("event:%3d %-7s %s (%s)\n",
	// 	evt.Offset,
	// 	edge,
	// 	t.Format(time.RFC3339Nano),
	// 	evt.Timestamp)
}

func main() {

	buttonTimes.lastButtonEventType = "falling"
	buttonTimes.lastMessageReceived = time.Now()

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

	// Set up button with interrupt watch using gpiod
	// offset := rpi.J8p13
	offset := 27

	c, err := gpiod.NewChip("gpiochip0")
	if err != nil {
		panic(err)
	}

	defer c.Close()

	// t, err := c.RequestLine(redButton,
	// 	gpiod.WithRisingEdge)
	// if err != nil {
	// 	panic(err)
	// }
	// defer t.Close()

	// period := 10 * time.Millisecond
	// l, _ = c.RequestLine(4, gpiod.WithDebounce(period))// during request The WithDebounce option requires Linux v5.10 or later.
	// l.Reconfigure(gpiod.WithDebounce(period))         // once requested

	l, err := c.RequestLine(offset,
		gpiod.WithBothEdges,
		gpiod.WithEventHandler(buttonEventHandler))

	// NB remove pullup from the gpiod function call: requires kernel 5.5 for pullup/pulldown support.
	if err != nil {
		fmt.Printf("RequestLine returned error: %s\n", err)
		if err == syscall.Errno(22) {
			fmt.Println("Note that the WithPullUp option requires kernel V5.5 or later - check your kernel version.")
		}
		os.Exit(1)
	}
	defer l.Close()

	// l.Reconfigure(gpiod.WithDebounce(period)) // once requested

	// In a real application the main thread would do something useful.
	// But we'll just run for a minute then exit.
	fmt.Printf("Watching Pin %d...\n", offset)
	time.Sleep(time.Minute)
	fmt.Println("exiting...")

}
