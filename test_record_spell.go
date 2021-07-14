// test_record_spell.go

// requires a raspi 3 or zero
// connectd with a push button on GPIO and IMU (Bosch BNo055)
// determine what letter the user draws in the air

// NB binary must be run as sudo
// go build test_record_spell.go && sudo ./test_record_spell

// read switch input from raspberry pi 3+ GPIO and light LED
// when button is down for a "long" time (>500 ms): record IMU data.
// on button-up, we convert the quaternion data from IMU (ie accelerometer and gyroscope) into a 28x28 image
// the image is then piped to a tensorflowlite classify model in python
// the python app then returns the best guess letter and %prob

package main

import (
	"flag"

	"github.com/johnusher/ardpifi/pkg/acc"
	"github.com/johnusher/ardpifi/pkg/gpio"
	log "github.com/sirupsen/logrus"
)

func main() {

	noACC := flag.Bool("no-acc", false, "run without Bosch accelerometer")

	// init accelerometer module (Bosch)
	accChan := make(chan acc.ACCMessage)
	a, err := acc.Init(accChan, *noACC)
	if err != nil {
		log.Errorf("failed to initialize acc: %s", err)
		return
	}

	// init gpio module:
	gpioChan := make(chan gpio.GPIOMessage)
	// gp, err := gpio.Init(gpioChan, *noGPIO)  // TBD
	gp, err := gpio.Init(gpioChan)
	if err != nil {
		log.Errorf("failed to initialize GPIO: %s", err)
		return
	}
	defer gp.Close()

	// main loop here:
	// go forth

	go a.Run()

	errs := make(chan error)

	go func() {
		errs <- GPIOLoop(gpioChan, accChan)
	}()

	// block until ctrl-c or one of the loops returns an error
	select {
	case <-errs:
	}

}

func GPIOLoop(gpioCh <-chan gpio.GPIOMessage, accCh <-chan acc.ACCMessage) error {
	log.Info("Starting GPIO loop")

	gpioMessage := gpio.GPIOMessage{}

	more := false
	for {

		select {

		case gpioMessage, more = <-gpioCh:

			if !more {
				log.Infof("gpio channel closed\n")
				log.Infof("exiting")
				return nil
			}

			// log.Infof("gpio message %v", gpioMessage)
			// receive a button change from gpio

			buttonStatus := gpioMessage.ButtonFlag
			// buttonStatus := gpio.GPIOMessage.buttonFlag
			if buttonStatus == 0 {
				// button down
				log.Infof("button down %v", buttonStatus)

				// start recording quaternions from IMU
			}

			if buttonStatus == 1 {
				// button up
				log.Infof("button up %v", buttonStatus)

				// stop recording quaternions from IMU,
				// convert quaternions to 28x28 image
				// pipe to TF, Python

			}

		}

	}

}
