// test_record_spell.go

// requires a raspi 3 or zero
// connectd with a push button and IMU (Bosch BNo055)
// determine what letter the user draws in the air


// NB binary must be run as sudo
// read switch input from raspberry pi 3+ GPIO and light LED
// when button is down for a "long" time (>500 ms): record IMU data.
// on button-up, we convert the quaternion data from IMU (ie accelerometer and gyroscope) into a 28x28 image
// the image is then piped to a tensorflowlite classify model in python
// the python app then returns the best guess letter and %prob

package main

import (

	"github.com/johnusher/ardpifi/pkg/gpio"
	log "github.com/sirupsen/logrus"


)




func main() {


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
	go gp.Run()

}
