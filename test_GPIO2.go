// test_GPIO2.go

// uses GPIO package

package main

import (
	// "bufio"
	// "encoding/base64"
	"flag"
	// "fmt"
	// "image"
	// "io"
	// "os/exec"
	// "runtime"
	// "strconv"
	// "strings"

	// "math"
	// "os"
	// "time"

	"github.com/johnusher/ardpifi/pkg/gpio"

	log "github.com/sirupsen/logrus"
)

func main() {

	// parse inut flags for no hardware
	// NB no-sound just means do not output sound- still need I2S connections (probably)
	noSound := flag.Bool("no-sound", false, "run without sound")

	logLevel := flag.String("log-level", "info", "log level, must be one of: panic, fatal, error, warn, info, debug, trace")

	flag.Parse()

	level, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Errorf("failed to parse log level [%s]: %s", *logLevel, err)
		return
	}
	log.SetLevel(level)

	// init gpio module:
	gpioChan := make(chan gpio.GPIOMessage)
	// gp, err := gpio.Init(gpioChan, *noGPIO)  // TBD
	gp, err := gpio.Init(gpioChan, *noSound)
	if err != nil {
		log.Errorf("failed to initialize GPIO: %s", err)
		return
	}
	defer gp.Close()

	// main loop here:
	// go forth

	errs := make(chan error)

	go func() {
		errs <- GPIOLoop(gpioChan)
	}()

	// block until ctrl-c or one of the loops returns an error
	select {
	case <-errs:
	}

}

func GPIOLoop(gpioCh <-chan gpio.GPIOMessage) error {
	// log.Info("Starting GPIO loop")

	gpioMessage := gpio.GPIOMessage{}
	// buttonDown := false
	n := 0

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
				// log.Infof("button down %v", buttonStatus)
				// buttonDown = true
				n = 0
				// start recording quaternions from IMU
			}

			if buttonStatus == 1 {
				// button up
				// log.Infof("button up %v", buttonStatus)
				// buttonDown = false

				// stop recording quaternions from IMU,
				// convert quaternions to 28x28 image
				// pipe to TF, Python

				if n > 20 {

				} else {
					log.Printf("shorty")
				}

			}

		}

	}

}
